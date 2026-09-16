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

// TestServe_ShutdownIsClean pins the goroutine-baseline teardown: after the
// cleanup runs, no goroutine leak is reported (the cleanup itself would
// fail the test).
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

	if err := svc.Shutdown(ctx); err != nil {
		t.Errorf("explicit shutdown: %v", err)
	}
}
