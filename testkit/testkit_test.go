package testkit_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
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

	server := testkit.Serve(t, svc)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.FullChainURL+"/hello", nil)
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
// Built via NewService, so this test also proves the test-safe defaults
// compose with the explicit-stop flow (and does not pay the 5s drain).
func TestServe_ShutdownIsClean(t *testing.T) {
	t.Parallel()

	server := testkit.Serve(t, testkit.NewService(t, appkit.DefaultServiceConfig()))

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.FullChainURL+"/health", nil)
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

	if err := server.Shutdown(ctx); err != nil {
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

	server := testkit.Serve(t, svc)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("explicit shutdown: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.FullChainURL+"/hello", nil)
	if err != nil {
		t.Fatalf("build GET: %v", err)
	}

	doResp, doErr := http.DefaultClient.Do(req)
	if doErr == nil {
		_ = doResp.Body.Close()

		t.Error("request after Shutdown succeeded — server still reachable")
	}

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("second Shutdown = %v, want nil (idempotent)", err)
	}
}

// TestNewService_NeutralizesProductionDefaults pins the test-safe
// defaults: a config straight from DefaultServiceConfig — production
// ":8080" address and 5s DrainDelay — serves on a loopback ephemeral
// port and shuts down without waiting out the drain. Both neutralizations
// are behavioral: the URL shows the loopback address, and the explicit
// shutdown completes far inside the 5s a production drain would cost.
func TestNewService_NeutralizesProductionDefaults(t *testing.T) {
	t.Parallel()

	server := testkit.Serve(t, testkit.NewService(t, appkit.DefaultServiceConfig()))

	parsed, err := url.Parse(server.FullChainURL)
	if err != nil {
		t.Fatalf("parse FullChainURL %q: %v", server.FullChainURL, err)
	}

	if host := parsed.Hostname(); host != "127.0.0.1" {
		t.Errorf("served host = %q, want 127.0.0.1 (loopback, not the production addr)", host)
	}

	if port := parsed.Port(); port == "8080" {
		t.Error("served port = 8080, want an ephemeral port (production addr must not leak into tests)")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	start := time.Now()

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if elapsed := time.Since(start); elapsed >= 3*time.Second {
		t.Errorf("shutdown took %v, want <3s (the 5s production drain default must be neutralized)", elapsed)
	}
}

// TestNewService_KeepsExplicitDrainDelay pins the explicit-value-wins rule:
// a custom DrainDelay passes through untouched, so drain-window tests
// (DrainWindowProbe) keep working through the test-safe constructor.
func TestNewService_KeepsExplicitDrainDelay(t *testing.T) {
	t.Parallel()

	cfg := appkit.DefaultServiceConfig()
	cfg.DrainDelay = 120 * time.Millisecond

	server := testkit.Serve(t, testkit.NewService(t, cfg))

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	start := time.Now()

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if elapsed := time.Since(start); elapsed < 120*time.Millisecond {
		t.Errorf("shutdown took %v, want >=120ms (explicit DrainDelay must be kept)", elapsed)
	}
}

// TestDrainWindowProbe_ObservesDrainContract exercises the drain-window
// helper end to end: a drain hook built from DrainWindowProbe sees the ready
// probe already at 503 while /ping still answers 200 — the exact assertion
// consumers hand-roll today.
func TestDrainWindowProbe_ObservesDrainContract(t *testing.T) {
	t.Parallel()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "127.0.0.1:0"
	cfg.DrainDelay = 150 * time.Millisecond

	readyStatus, pingStatus := 0, 0

	// base is assigned after Serve; the probe's supplier resolves lazily at
	// hook-run time, which is exactly the late-capture the helper documents.
	var base string

	cfg.DrainHooks = append(cfg.DrainHooks,
		testkit.DrainWindowProbe(func() string { return base }, "/health/ready", &readyStatus),
		testkit.DrainWindowProbe(func() string { return base }, "/ping", &pingStatus),
	)

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := testkit.Serve(t, svc)

	base = server.FullChainURL

	shutdownCtx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if readyStatus != http.StatusServiceUnavailable {
		t.Errorf("ready status during drain = %d, want 503", readyStatus)
	}

	if pingStatus != http.StatusOK {
		t.Errorf("ping status during drain = %d, want 200 (socket still serves)", pingStatus)
	}
}
