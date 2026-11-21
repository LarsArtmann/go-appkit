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

// unixHTTPClient returns an http.Client dialing a unix socket. The only
// consumer-side machinery the listener seam leaves to the caller.
func unixHTTPClient(socketPath string) *http.Client {
	return &http.Client{ //nolint:exhaustruct_v5 // only the transport matters
		Transport: &http.Transport{ //nolint:exhaustruct_v5 // only DialContext matters
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
}

func drainAndClose(t *testing.T, resp *http.Response) {
	t.Helper()

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func TestListener_PreboundTCP_ServesAndReportsAddr(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("pre-bind: %v", err)
	}

	svc, err := NewService(ServiceConfig{
		Listener:   ln,
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

	if got, want := svc.Addr().String(), ln.Addr().String(); got != want {
		t.Errorf("Addr() = %q, want the supplied listener's %q", got, want)
	}

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	resp := httpGet(t, ctx, "http://"+ln.Addr().String()+"/health")
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
	if _, err := net.Dial("tcp", ln.Addr().String()); err == nil {
		t.Error("supplied listener still accepting after shutdown; want closed")
	}
}

func TestListener_UnixSocket_Serves(t *testing.T) {
	t.Parallel()

	socketPath := filepath.Join(t.TempDir(), "appkit-test.sock")

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("unix bind: %v", err)
	}

	svc, err := NewService(ServiceConfig{
		Listener:   ln,
		DrainDelay: NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// Poll the socket file because a unix listener's Addr().String() gives a
	// path, not a dialable host:port — waitForRunning's svc.Running() probe
	// is the sync point, the client call is the behavioral pin.
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

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("pre-bind: %v", err)
	}

	addr := ln.Addr().String()

	svc, err := NewService(ServiceConfig{
		Listener:   ln,
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
		if _, err := net.Dial("tcp", addr); err != nil {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Error("supplied listener still accepting after failed start; want closed")
}

func TestListener_ConfigInvalid_ClosesSuppliedListener(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("pre-bind: %v", err)
	}

	addr := ln.Addr().String()

	// The delegated server rejects ReadHeaderTimeout > ReadTimeout > 0 —
	// surfaces synchronously as server.config_invalid BEFORE any lifecycle
	// state, and the supplied listener must be closed again.
	svc, err := NewService(ServiceConfig{
		Listener:          ln,
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
		if _, err := net.Dial("tcp", addr); err != nil {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Error("supplied listener still accepting after config_invalid; want closed")
}
