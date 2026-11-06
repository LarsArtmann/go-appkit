package preset_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/testkit"
)

// TestNew_ZeroAppCodeBoots pins the v12 thesis: an EMPTY Config, no env —
// the whole spine comes up on its own. The dashboard owns /health (core's
// default /health family is opted out to avoid the mux panic), the K8s
// probe endpoints serve, and teardown is a graceful shutdown.
func TestNew_ZeroAppCodeBoots(t *testing.T) {

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	for _, path := range []string{"/healthz", "/readyz", "/startupz"} {
		resp, err := http.Get(ts.FullChainURL + path)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s = %d, want 200", path, resp.StatusCode)
		}
	}

	dash, err := http.Get(ts.FullChainURL + "/health")
	if err != nil {
		t.Fatalf("dashboard: get: %v", err)
	}
	defer dash.Body.Close()

	if dash.StatusCode != http.StatusOK {
		t.Errorf("dashboard /health = %d, want 200", dash.StatusCode)
	}

	body, err := io.ReadAll(dash.Body)
	if err != nil {
		t.Fatalf("dashboard: read: %v", err)
	}

	if !strings.Contains(dash.Header.Get("Content-Type"), "text/html") {
		t.Errorf("dashboard Content-Type = %q, want text/html", dash.Header.Get("Content-Type"))
	}

	if !strings.Contains(strings.ToLower(string(body)), "<!doctype html") {
		t.Error("dashboard body is not an HTML document")
	}
}

// TestNew_HookWiring proves the spine's lifecycle wiring: the health
// Mounted starts with the service (its checks run), and consumer hooks
// pass through and run in phase order (start, then drain, then shutdown).
func TestNew_HookWiring(t *testing.T) {

	var (
		checkRuns atomic.Int64
		events    []string
	)

	cfg := testConfig()
	cfg.HealthChecks = map[string]health.CheckFunc{
		"counting": func(context.Context) error {
			checkRuns.Add(1)
			return nil
		},
	}
	cfg.StartHooks = append(cfg.StartHooks, func(context.Context) error {
		events = append(events, "start")
		return nil
	})
	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
		events = append(events, "drain")
		return nil
	})
	cfg.ShutdownHooks = append(cfg.ShutdownHooks, func(context.Context) error {
		events = append(events, "shutdown")
		return nil
	})

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	// Mounted.Start ran: the probe's initial synchronous batch executes the
	// check at least once before the first dashboard patch.
	deadline := time.Now().Add(5 * time.Second)
	for checkRuns.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("health check never ran — Mounted.Start did not execute")
		}

		time.Sleep(5 * time.Millisecond)
	}

	if err := ts.Shutdown(t.Context()); err != nil {
		t.Fatalf("graceful shutdown: %v", err)
	}

	if len(events) != 3 || events[0] != "start" || events[1] != "drain" || events[2] != "shutdown" {
		t.Errorf("hook order = %v, want [start drain shutdown]", events)
	}
}

// TestNew_VersionPassthrough stamps ServiceConfig.Version.
func TestNew_VersionPassthrough(t *testing.T) {

	cfg := testConfig()
	cfg.Version = "preset-test-1.2.3"

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	resp, err := http.Get(ts.FullChainURL + "/version")
	if err != nil {
		t.Fatalf("get /version: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /version: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("/version = %d, want 200", resp.StatusCode)
	}

	if !strings.Contains(string(body), "preset-test-1.2.3") {
		t.Errorf("/version body %q does not contain the version", body)
	}
}

// TestNew_NoDashboard keeps core's own /health family: the probe endpoints
// still mount and coexist (no pattern collision).
func TestNew_NoDashboard(t *testing.T) {

	cfg := testConfig()
	cfg.NoDashboard = true

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	for _, path := range []string{"/healthz", "/readyz", "/health", "/health/ready"} {
		resp, err := http.Get(ts.FullChainURL + path)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s = %d, want 200", path, resp.StatusCode)
		}
	}
}
