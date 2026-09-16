package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/health"
	gohealth "github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
)

// gohealthRoutes builds the custom probe-route set (go-health's Routes type
// aliased locally to avoid the package-name collision in this file).
func gohealthRoutes() gohealth.Routes {
	return gohealth.Routes{
		Liveness:  "/custom/live",
		Readiness: "/custom/ready",
		Startup:   "/custom/start",
	}
}

// TestWithDashboard_ReplacesProbeRoutes pins the documented conflict
// semantics: with the dashboard on, `WithProbeRoutes` is IGNORED (the
// dashboard owns ALL routes from its own config) — and that is a tested,
// documented ignore, not a panic. If this behavior ever changes (e.g. a
// loud construction rejection), this test is the tripwire.
func TestWithDashboard_ReplacesProbeRoutes(t *testing.T) {
	t.Parallel()

	mounted, err := health.New(
		health.NewProbe(nil),
		health.WithDashboard(),
		health.WithProbeRoutes(gohealthRoutes()),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	mux := http.NewServeMux()
	mounted.RegisterRoutes(mux)

	for _, path := range []string{"/custom/live", "/custom/ready", "/custom/start"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("%s registered despite WithProbeRoutes being ignored under the dashboard", path)
		}
	}

	// The dashboard's own default surface is live instead.
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("dashboard /health: status = %d, want 200 (dashboard owns the surface)", rec.Code)
	}
}

// TestDashboardWithBasePath_UniformRouting pins the documented claim: the
// dashboard's WithBasePath applies uniformly — the dashboard HTML AND its
// probe endpoints all move under the prefix together.
func TestDashboardWithBasePath_UniformRouting(t *testing.T) {
	t.Parallel()

	mounted, err := health.New(
		health.NewProbe(nil),
		health.WithDashboard(dashboard.WithBasePath("/ops")),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	mux := http.NewServeMux()
	mounted.RegisterRoutes(mux)

	// Dashboard moved under the prefix.
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ops/health", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /ops/health: status = %d, want 200 (WithBasePath applies to the dashboard)", rec.Code)
	}

	// Probe endpoints moved WITH it (uniform routing).
	probePaths := map[string]bool{
		"/ops/healthz": false, "/ops/readyz": false, "/ops/startupz": false,
	}

	for path := range probePaths {
		r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		rc := httptest.NewRecorder()
		mux.ServeHTTP(rc, r)

		if rc.Code == http.StatusNotFound {
			t.Errorf("%s: 404 — WithBasePath did not apply uniformly", path)
		}

		probePaths[path] = true
	}

	// And the un-prefixed probes are gone (no duplicate surface).
	for _, path := range []string{"/healthz", "/readyz", "/startupz"} {
		r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		rc := httptest.NewRecorder()
		mux.ServeHTTP(rc, r)

		if rc.Code != http.StatusNotFound {
			t.Errorf("%s still registered without the base path — routing is not uniform", path)
		}
	}
}
