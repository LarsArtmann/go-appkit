package preset_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/preset"
)

// TestMetrics_NilStaysNil: nil Metrics passes through verbatim — the preset
// never invents credentials, so no /metrics route exists (the errorpages
// catch-all answers instead).
func TestMetrics_NilStaysNil(t *testing.T) {

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp, err := http.Get(ts + "/metrics")
	if err != nil {
		t.Fatalf("get /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("/metrics = %d, want 404 (nil must stay nil)", resp.StatusCode)
	}
}

// TestMetrics_Passthrough: a set MetricsConfig flows to the core untouched —
// auth-enforced scrape endpoint, 401 without credentials.
func TestMetrics_Passthrough(t *testing.T) {

	cfg := testConfig()
	cfg.Metrics = &appkit.MetricsConfig{
		Path:           "/metrics",
		BasicAuthUser:  "scrape",
		BasicAuthPass:  "secret",
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	unauthenticated, err := http.Get(ts + "/metrics")
	if err != nil {
		t.Fatalf("get /metrics (no auth): %v", err)
	}
	defer unauthenticated.Body.Close()

	if unauthenticated.StatusCode != http.StatusUnauthorized {
		t.Errorf("/metrics without auth = %d, want 401", unauthenticated.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, ts+"/metrics", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.SetBasicAuth("scrape", "secret")

	authenticated, err := (&http.Client{Timeout: httpTimeout}).Do(req)
	if err != nil {
		t.Fatalf("get /metrics (auth): %v", err)
	}
	defer authenticated.Body.Close()

	body, err := io.ReadAll(authenticated.Body)
	if err != nil {
		t.Fatalf("read /metrics: %v", err)
	}

	if authenticated.StatusCode != http.StatusOK {
		t.Errorf("/metrics with auth = %d, want 200", authenticated.StatusCode)
	}

	if len(body) == 0 {
		t.Error("/metrics body is empty — no exposition served")
	}
}
