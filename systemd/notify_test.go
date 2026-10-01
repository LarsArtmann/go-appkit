package systemd

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// envMu serializes tests that touch NOTIFY_SOCKET / WATCHDOG_USEC: the
// environment is process-global, exactly like flightrecorderhealth's
// recorderMu serializes its process-global singleton.
var envMu sync.Mutex

// discardLogger silences hook diagnostics in tests.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil)) //nolint:exhaustruct_v5 // zero HandlerOptions is the quiet default
}

// notifySocket is a test double for systemd's $NOTIFY_SOCKET: a bound
// unixgram socket that receives (and lets tests read) real sd_notify
// datagrams, exercising the same transport the service manager uses.
type notifySocket struct {
	conn *net.UnixConn
}

// newNotifySocket binds the socket and points NOTIFY_SOCKET at it. The
// caller must hold envMu (t.Setenv also forbids t.Parallel).
func newNotifySocket(t *testing.T) *notifySocket {
	t.Helper()

	path := filepath.Join(t.TempDir(), "notify.sock")

	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: path, Net: "unixgram"})
	if err != nil {
		t.Fatalf("listen unixgram: %v", err)
	}

	t.Setenv("NOTIFY_SOCKET", path)
	t.Cleanup(func() { _ = conn.Close() })

	return &notifySocket{conn: conn}
}

// awaitMessage reads one sd_notify datagram or fails the test.
func (s *notifySocket) awaitMessage(t *testing.T, what string) string {
	t.Helper()

	if err := s.conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	buf := make([]byte, 4096)

	n, err := s.conn.Read(buf)
	if err != nil {
		t.Fatalf("await %s: %v", what, err)
	}

	return string(buf[:n])
}

// collectMessages reads datagrams until the window elapses, then returns
// them. Used around shutdown boundaries where the exact count is racy but
// membership is not.
func (s *notifySocket) collectMessages(t *testing.T, window time.Duration) []string {
	t.Helper()

	var messages []string

	deadline := time.Now().Add(window)

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return messages
		}

		if err := s.conn.SetReadDeadline(time.Now().Add(remaining)); err != nil {
			t.Fatalf("set read deadline: %v", err)
		}

		buf := make([]byte, 4096)

		n, err := s.conn.Read(buf)
		if err != nil {
			return messages // window elapsed
		}

		messages = append(messages, string(buf[:n]))
	}
}

// assertQuiet fails the test if any datagram arrives within the window.
func (s *notifySocket) assertQuiet(t *testing.T, window time.Duration) {
	t.Helper()

	if err := s.conn.SetReadDeadline(time.Now().Add(window)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	buf := make([]byte, 4096)

	if n, err := s.conn.Read(buf); err == nil {
		t.Fatalf("expected silence for %v, got %q", window, string(buf[:n]))
	}
}

func TestNotify_DeliversStateToSocket(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	sock := newNotifySocket(t)

	sent, err := notify("READY=1")
	if err != nil {
		t.Fatalf("notify: %v", err)
	}

	if !sent {
		t.Fatal("notify must report sent=true with NOTIFY_SOCKET set")
	}

	if got := sock.awaitMessage(t, "READY=1"); got != "READY=1" {
		t.Fatalf("datagram = %q, want READY=1", got)
	}
}

func TestNotify_NoSocketIsANoOp(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	t.Setenv("NOTIFY_SOCKET", "")

	sent, err := notify("READY=1")
	if err != nil {
		t.Fatalf("notify outside systemd must not error: %v", err)
	}

	if sent {
		t.Error("notify must report sent=false without NOTIFY_SOCKET")
	}
}

func TestNotify_SendFailureIsClassified(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	// A socket path nothing listens on: connecting the unixgram peer is
	// refused, which is the transport failure the classification exists
	// for.
	t.Setenv("NOTIFY_SOCKET", filepath.Join(t.TempDir(), "missing.sock"))

	sent, err := notify("READY=1")
	if err == nil {
		t.Fatal("notify to a dead socket must fail")
	}

	if sent {
		t.Error("notify must report sent=false on failure")
	}

	if code := errorfamily.Code(err); code != "systemd.notify_failed" {
		t.Errorf("error code = %q, want systemd.notify_failed", code)
	}

	if !errors.Is(err, syscall.ENOENT) {
		t.Errorf("the transport errno must stay reachable through the wrapper: %v", err)
	}
}

func TestWatchdogInterval_ReadsEnv(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	t.Setenv("WATCHDOG_USEC", "2000000")
	t.Setenv("WATCHDOG_PID", "")

	interval, err := watchdogInterval()
	if err != nil {
		t.Fatalf("watchdogInterval: %v", err)
	}

	if interval != 2*time.Second {
		t.Errorf("interval = %v, want 2s", interval)
	}
}

func TestWatchdogInterval_UnsetMeansDisabled(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	t.Setenv("WATCHDOG_USEC", "")

	interval, err := watchdogInterval()
	if err != nil {
		t.Fatalf("watchdogInterval: %v", err)
	}

	if interval != 0 {
		t.Errorf("interval = %v, want 0 when WATCHDOG_USEC is unset", interval)
	}
}

func TestWatchdogInterval_InvalidValueIsClassified(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	t.Setenv("WATCHDOG_USEC", "not-a-number")

	interval, err := watchdogInterval()
	if err == nil {
		t.Fatal("invalid WATCHDOG_USEC must error")
	}

	if interval != 0 {
		t.Errorf("interval = %v, want 0 on error", interval)
	}

	if code := errorfamily.Code(err); code != "systemd.watchdog_check_failed" {
		t.Errorf("error code = %q, want systemd.watchdog_check_failed", code)
	}
}

func TestRunWatchdog_PingsUntilStopped(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	sock := newNotifySocket(t)

	stop := make(chan struct{})

	const pingEvery = 20 * time.Millisecond

	go runWatchdog(discardLogger(), stop, pingEvery)

	if got := sock.awaitMessage(t, "first WATCHDOG=1"); got != "WATCHDOG=1" {
		t.Fatalf("first datagram = %q, want WATCHDOG=1", got)
	}

	// More pings arrive while running.
	_ = sock.awaitMessage(t, "second WATCHDOG=1")

	close(stop)

	// Allow the goroutine to observe the close and any in-flight datagram
	// to land, then demand silence for several ping periods.
	_ = sock.collectMessages(t, 100*time.Millisecond)

	sock.assertQuiet(t, 5*pingEvery)
}
