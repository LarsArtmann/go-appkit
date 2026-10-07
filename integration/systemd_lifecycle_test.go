package integration_test

// T04 — systemd × core composition E2E, against PUBLISHED tags (systemd
// v0.1.0 + core v0.8.0): a real appkit Service with the sd_notify lifecycle
// installed announces READY=1 from the StartHook (after the listener binds,
// before the first request can be served) and STOPPING=1 from the DrainHook
// during shutdown — observed on a real unixgram NOTIFY_SOCKET, exactly the
// unit-file contract a Type=notify consumer depends on.
//
// Deliberately leaner than systemd's own module suite: no watchdog here
// (its WatchdogSec/2 keepalive timing is module-pinned by
// TestInstall_LifecycleThroughAppkitService); this test pins only the
// cross-module composition on published pins. Not parallel: it mutates the
// process environment (NOTIFY_SOCKET).

import (
	"net"
	"path/filepath"
	"slices"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	appkitsystemd "github.com/larsartmann/go-appkit/systemd"
	"github.com/larsartmann/go-appkit/testkit"
)

// notifySocket is a bound unixgram stand-in for the manager's NOTIFY_SOCKET.
type notifySocket struct {
	conn *net.UnixConn
}

func bindNotifySocket(t *testing.T) *notifySocket {
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

// readOne reads a single sd_notify datagram within the deadline.
func (s *notifySocket) readOne(t *testing.T, window time.Duration) string {
	t.Helper()

	if err := s.conn.SetReadDeadline(time.Now().Add(window)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	var buf [4096]byte

	n, err := s.conn.Read(buf[:])
	if err != nil {
		t.Fatalf("read datagram: %v", err)
	}

	return string(buf[:n])
}

// drain reads every queued datagram until the socket goes quiet for the
// window, returning them all.
func (s *notifySocket) drain(t *testing.T, quiet time.Duration) []string {
	t.Helper()

	var messages []string
	quietDeadline := time.Now().Add(quiet)

	for {
		_ = s.conn.SetReadDeadline(quietDeadline)

		var buf [4096]byte

		n, err := s.conn.Read(buf[:])
		if err != nil {
			return messages
		}

		messages = append(messages, string(buf[:n]))
	}
}

func TestSystemdLifecycleThroughAppkitService(t *testing.T) {
	sock := bindNotifySocket(t)

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:0"
	cfg.DrainDelay = appkit.NoDrainDelay

	appkitsystemd.Install(&cfg)

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	ts := testkit.Serve(t, svc)

	// READY=1 is the first datagram, sent from the StartHook after the
	// listener bound — by the time Serve returned, the service is serving.
	if got := sock.readOne(t, 2*time.Second); got != "READY=1" {
		t.Fatalf("first datagram = %q, want READY=1", got)
	}

	if err := ts.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	// STOPPING=1 was sent by the DrainHook before connections were released.
	messages := sock.drain(t, 300*time.Millisecond)
	if !slices.Contains(messages, "STOPPING=1") {
		t.Fatalf("no STOPPING=1 observed during shutdown (got %v)", messages)
	}
}
