package appkit_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
)

func newMetricsService(t *testing.T, mutate func(*appkit.ServiceConfig)) *appkit.Service {
	t.Helper()

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay
	cfg.Metrics = &appkit.MetricsConfig{
		BasicAuthUser: "metrics",
		BasicAuthPass: "secret",
	}
	cfg.Addr = "127.0.0.1:0"

	if mutate != nil {
		mutate(&cfg)
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	return svc
}

func startTestService(t *testing.T, svc *appkit.Service) string {
	t.Helper()

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := svc.Shutdown(ctx); err != nil {
			t.Errorf("shutdown: %v", err)
		}

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("server error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop after shutdown")
		}
	})

	deadline := time.Now().Add(2 * time.Second)
	for !svc.Running() {
		if time.Now().After(deadline) {
			t.Fatal("service did not start within timeout")
		}

		time.Sleep(time.Millisecond)
	}

	return "http://" + svc.Addr().String()
}

func scrapeMetrics(t *testing.T, baseURL string) string {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/metrics", nil)
	if err != nil {
		t.Fatalf("build scrape request: %v", err)
	}

	req.SetBasicAuth("metrics", "secret")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read exposition: %v", err)
	}

	return string(body)
}

func TestMetrics_UnauthenticatedConstructionRejected(t *testing.T) {
	t.Parallel()

	cfg := appkit.DefaultServiceConfig()
	cfg.Metrics = &appkit.MetricsConfig{}

	_, err := appkit.NewService(cfg)
	if err == nil {
		t.Fatal("unauthenticated metrics must be rejected at construction")
	}
}

func TestMetrics_BasicAuthEnforced(t *testing.T) {
	t.Parallel()

	svc := newMetricsService(t, nil)
	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	baseURL := startTestService(t, svc)

	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no credentials: status = %d, want 401", resp.StatusCode)
	}

	if ch := resp.Header.Get("WWW-Authenticate"); ch == "" {
		t.Error("401 must carry WWW-Authenticate")
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/metrics", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.SetBasicAuth("metrics", "wrong")

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET with wrong creds: %v", err)
	}

	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong credentials: status = %d, want 401", resp2.StatusCode)
	}

	if body := scrapeMetrics(t, baseURL); !strings.Contains(body, "appkit_build_info") {
		t.Errorf("correct credentials: expected exposition, got:\n%s", body)
	}
}

// TestMetrics_ExpositionContract pins the stable metric-name contract and
// the route-pattern label (cardinality-safe: raw paths stay out).
func TestMetrics_ExpositionContract(t *testing.T) {
	t.Parallel()

	svc := newMetricsService(t, nil)
	svc.Mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	baseURL := startTestService(t, svc)

	resp, err := http.Get(baseURL + "/users/42")
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	defer resp.Body.Close()

	body := scrapeMetrics(t, baseURL)

	for _, want := range []string{
		"# TYPE appkit_http_request_duration_seconds histogram",
		"# TYPE appkit_http_responses_total counter",
		"# TYPE appkit_http_requests_in_flight gauge",
		"# TYPE appkit_build_info gauge",
		`appkit_http_responses_total{method="GET",route="GET /users/{id}",status="200"}`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("exposition missing %q", want)
		}
	}

	if strings.Contains(body, "/users/42") {
		t.Error("raw path leaked into labels — cardinality blow-up; want the route pattern only")
	}
}

func TestMetrics_BuildInfoCarriesVersion(t *testing.T) {
	t.Parallel()

	svc := newMetricsService(t, func(cfg *appkit.ServiceConfig) {
		cfg.Version = "v9.9.9-test"
	})

	baseURL := startTestService(t, svc)

	if body := scrapeMetrics(t, baseURL); !strings.Contains(body, `appkit_build_info{version="v9.9.9-test"} 1`) {
		t.Errorf("build info missing version label")
	}
}

func TestMetrics_UnmatchedPathBounded(t *testing.T) {
	t.Parallel()

	svc := newMetricsService(t, nil)
	baseURL := startTestService(t, svc)

	resp, err := http.Get(baseURL + "/no/such/path")
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	defer resp.Body.Close()

	if body := scrapeMetrics(t, baseURL); !strings.Contains(body, `route="unmatched"`) {
		t.Error("unmatched requests must fall back to the bounded 'unmatched' label")
	}
}

// TestMetrics_SSEStillFlushes guards the statusRecorder passthrough: an SSE
// stream behind the metrics middleware must still flush (headers before the
// first event).
func TestMetrics_SSEStillFlushes(t *testing.T) {
	t.Parallel()

	svc := newMetricsService(t, nil)
	svc.Mux.HandleFunc("GET /stream", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
	})

	baseURL := startTestService(t, svc)

	resp, err := http.Get(baseURL + "/stream")
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("SSE through metrics middleware: status = %d, want 200", resp.StatusCode)
	}
}

// TestMetrics_VersionEndpoint pins the F5 battery: Version → /version.
func TestMetrics_VersionEndpoint(t *testing.T) {
	t.Parallel()

	svc := newMetricsService(t, func(cfg *appkit.ServiceConfig) {
		cfg.Version = "v1.2.3"
		cfg.Metrics.AllowUnauthenticated = true
		cfg.Metrics.BasicAuthUser = ""
		cfg.Metrics.BasicAuthPass = ""
	})

	baseURL := startTestService(t, svc)

	resp, err := http.Get(baseURL + "/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /version: %v", err)
	}

	if !strings.Contains(string(body), `"version":"v1.2.3"`) {
		t.Errorf("/version = %q, want the configured version", string(body))
	}
}
