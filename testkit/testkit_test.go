package testkit_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/testkit"
)

// TestServe_FullChainServesDefaultStack pins the harness's core claim: a
// request through TestServer.FullChainURL crosses the REAL default stack —
// evidenced by the SecurityHeaders middleware's X-Frame-Options header,
// which httptest.NewServer(svc.Mux) would never add.
func TestServe_FullChainServesDefaultStack(t *testing.T) {
	t.Parallel()

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "127.0.0.1:0"
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := testkit.Serve(t, svc)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/hello", nil)
	if err != nil {
		t.Fatalf("build GET: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	if resp.Header.Get("X-Frame-Options") == "" {
		t.Error("full chain not applied: SecurityHeaders header missing (raw-mux server would omit it)")
	}
}

// TestServe_ShutdownIsClean pins the goroutine-baseline teardown: an
// explicit TestServer.Shutdown runs the leak assertion and leaves cleanup
// nothing to redo (either would fail the test if a goroutine leaked).
func TestServe_ShutdownIsClean(t *testing.T) {
	t.Parallel()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "127.0.0.1:0"

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	ts := testkit.Serve(t, svc)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/health", nil)
	if err != nil {
		t.Fatalf("build GET: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ts.Shutdown(ctx); err != nil {
		t.Errorf("explicit shutdown: %v", err)
	}
}

// TestServe_ExplicitShutdownIsIdempotentAndUnreachable pins the
// TestServer.Shutdown contract: after the explicit stop the endpoint is
// unreachable, a second Shutdown is a nil no-op, and cleanup's repeated
// stop sequence is skipped (the leak assertion already ran).
func TestServe_ExplicitShutdownIsIdempotentAndUnreachable(t *testing.T) {
	t.Parallel()

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "127.0.0.1:0"
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := testkit.Serve(t, svc)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ts.Shutdown(ctx); err != nil {
		t.Fatalf("explicit shutdown: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/hello", nil)
	if err != nil {
		t.Fatalf("build GET: %v", err)
	}

	if _, err := http.DefaultClient.Do(req); err == nil {
		t.Error("request after Shutdown succeeded — server still reachable")
	}

	if err := ts.Shutdown(ctx); err != nil {
		t.Errorf("second Shutdown = %v, want nil (idempotent)", err)
	}
}
