package appkit

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func listenerClosed(t *testing.T, ctx context.Context, network, addr string) bool {
	t.Helper()

	var d net.Dialer

	_, err := d.DialContext(ctx, network, addr)

	return err != nil
}

// unixHTTPClient returns an http.Client dialing a unix socket. The only
// consumer-side machinery the listener seam leaves to the caller.
func unixHTTPClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer

				return d.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

func drainAndClose(t *testing.T, resp *http.Response) {
	t.Helper()

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func prebindTCP(t *testing.T) net.Listener {
	t.Helper()

	var lc net.ListenConfig

	ln, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("pre-bind: %v", err)
	}

	return ln
}

func TestListener_PreboundTCP_ServesAndReportsAddr(t *testing.T) {
	t.Parallel()

	listener := prebindTCP(t)

	svc, err := NewService(ServiceConfig{
		Listener:   listener,
		DrainDelay: NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	if got, want := svc.Addr().String(), listener.Addr().String(); got != want {
		t.Errorf("Addr() = %q, want the supplied listener's %q", got, want)
	}

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	resp := httpGet(t, ctx, "http://"+listener.Addr().String()+"/health")
	drainAndClose(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want 200", resp.StatusCode)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(t.Context(), testTimeout)
	defer cancelShutdown()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	assertServerStopped(t, errCh)

	// Ownership pin: graceful shutdown closed the supplied listener.
	dialCtx, cancelDial := context.WithTimeout(t.Context(), testTimeout)
	defer cancelDial()

	if !listenerClosed(t, dialCtx, "tcp", listener.Addr().String()) {
		t.Error("supplied listener still accepting after shutdown; want closed")
	}
}

func TestListener_UnixSocket_Serves(t *testing.T) {
	t.Parallel()

	socketPath := filepath.Join(t.TempDir(), "appkit-test.sock")

	var lc net.ListenConfig

	listener, err := lc.Listen(t.Context(), "unix", socketPath)
	if err != nil {
		t.Fatalf("unix bind: %v", err)
	}

	svc, err := NewService(ServiceConfig{
		Listener:   listener,
		DrainDelay: NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	client := unixHTTPClient(socketPath)

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/health", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unix-dialed request: %v", err)
	}

	drainAndClose(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("health over unix socket = %d, want 200", resp.StatusCode)
	}

	if svc.Addr().String() != socketPath {
		t.Errorf("Addr() = %q, want socket path %q", svc.Addr().String(), socketPath)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(t.Context(), testTimeout)
	defer cancelShutdown()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	assertServerStopped(t, errCh)

	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Errorf("socket file still present after shutdown (stat err = %v); want removed by close", err)
	}
}

func TestListener_StartHookFailure_ClosesSuppliedListener(t *testing.T) {
	t.Parallel()

	listener := prebindTCP(t)

	addr := listener.Addr().String()

	svc, err := NewService(ServiceConfig{
		Listener:   listener,
		DrainDelay: NoDrainDelay,
		StartHooks: []Hook{
			func(context.Context) error { return errStartHookFailed },
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := svc.Start(); err == nil {
		t.Fatal("start with failing hook must fail")
	}

	// Ownership pin: the failure path closed the SUPPLIED listener too.
	deadline := time.Now().Add(testTimeout)

	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)

		closed := listenerClosed(t, ctx, "tcp", addr)

		cancel()

		if closed {
			return
		}
	}

	t.Error("supplied listener still accepting after failed start; want closed")
}

func TestListener_ConfigInvalid_ClosesSuppliedListener(t *testing.T) {
	t.Parallel()

	listener := prebindTCP(t)

	addr := listener.Addr().String()

	// The delegated server rejects ReadHeaderTimeout > ReadTimeout > 0 —
	// surfaces synchronously as server.config_invalid BEFORE any lifecycle
	// state, and the supplied listener must be closed again.
	svc, err := NewService(ServiceConfig{
		Listener:          listener,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, startErr := svc.Start()
	if startErr == nil {
		t.Fatal("start with config_invalid must fail")
	}

	deadline := time.Now().Add(testTimeout)

	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)

		closed := listenerClosed(t, ctx, "tcp", addr)

		cancel()

		if closed {
			return
		}
	}

	t.Error("supplied listener still accepting after config_invalid; want closed")
}
