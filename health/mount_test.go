package health

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-health"
)

const testTimeout = 2 * time.Second

func healthyChecks() map[string]CheckFunc {
	return map[string]CheckFunc{
		"database": func(context.Context) error { return nil },
	}
}

func getBody(t *testing.T, url string, header http.Header) (int, http.Header, string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s: %v", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, resp.Header, string(body)
}

func TestMount_RejectsNilArguments(t *testing.T) {
	t.Parallel()

	_, err := Mount(nil, nil)
	if err == nil {
		t.Error("nil mux must be rejected")
	}

	_, err = Mount(http.NewServeMux(), nil)
	if err == nil {
		t.Error("nil probe must be rejected")
	}

	_, err = New(nil)
	if err == nil {
		t.Error("nil probe must be rejected by New")
	}
}

func TestNew_RegisterRoutesIsThePrimaryAppkitFlow(t *testing.T) {
	t.Parallel()

	mounted, err := New(NewProbe(healthyChecks()))
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	mux := http.NewServeMux()
	mounted.RegisterRoutes(mux)

	server := httptest.NewServer(mux)
	defer server.Close()

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	code, _, _ := getBody(t, server.URL+"/readyz", nil)
	if code != http.StatusOK {
		t.Errorf("GET /readyz = %d, want 200", code)
	}
}

func TestMount_ProbeOnlyRegistersKubeletRoutes(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	mounted, err := Mount(mux, NewProbe(healthyChecks()))
	if err != nil {
		t.Fatalf("mount: %v", err)
	}

	if mounted.Dashboard() != nil {
		t.Error("dashboard must be nil when mounted without WithDashboard")
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	for _, route := range []string{"/healthz", "/readyz", "/startupz"} {
		code, _, _ := getBody(t, server.URL+route, nil)
		if code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", route, code)
		}
	}

	code, _, _ := getBody(t, server.URL+"/health", nil)
	if code != http.StatusNotFound {
		t.Errorf("GET /health without dashboard = %d, want 404 (must not collide with appkit's route)", code)
	}

	mounted.Drain()

	code, _, body := getBody(t, server.URL+"/readyz", nil)
	if code != http.StatusServiceUnavailable {
		t.Errorf("GET /readyz after drain = %d, want 503", code)
	}

	if !strings.Contains(body, `"shutting_down":true`) {
		t.Errorf("readyz body = %s, want shutting_down flag", body)
	}

	err = mounted.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestMount_CustomProbeRoutes(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	mounted, err := Mount(mux, NewProbe(healthyChecks()), WithProbeRoutes(health.Routes{
		Liveness:  "/live",
		Readiness: "/ready",
		Startup:   "/boot",
	}))
	if err != nil {
		t.Fatalf("mount: %v", err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	for _, route := range []string{"/live", "/ready", "/boot"} {
		code, _, _ := getBody(t, server.URL+route, nil)
		if code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", route, code)
		}
	}
}

func TestMount_WithDashboardServesHTMLJSONAndProbes(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	mounted, err := Mount(mux, NewProbe(healthyChecks()), WithDashboard())
	if err != nil {
		t.Fatalf("mount: %v", err)
	}

	if mounted.Dashboard() == nil {
		t.Fatal("dashboard accessor must be non-nil when mounted with WithDashboard")
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	code, headers, body := getBody(t, server.URL+"/health", nil)
	if code != http.StatusOK {
		t.Fatalf("GET /health = %d, want 200", code)
	}

	if contentType := headers.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Errorf("content type = %q, want HTML", contentType)
	}

	if !strings.Contains(body, "Health") {
		t.Error("dashboard HTML must render the health view")
	}

	code, _, body = getBody(t, server.URL+"/health", http.Header{"Accept": []string{"application/json"}})
	if code != http.StatusOK {
		t.Fatalf("GET /health as JSON = %d, want 200", code)
	}

	var payload struct {
		Status string `json:"status"`
	}

	err = json.Unmarshal([]byte(body), &payload)
	if err != nil {
		t.Fatalf("decode json %q: %v", body, err)
	}

	if payload.Status != "pass" {
		t.Errorf("json status = %q, want pass", payload.Status)
	}

	code, _, _ = getBody(t, server.URL+"/healthz", nil)
	if code != http.StatusOK {
		t.Errorf("GET /healthz = %d, want 200 (dashboard registers probe routes)", code)
	}
}

func TestMount_DrainFlipsDashboardReadiness(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	mounted, err := Mount(mux, NewProbe(healthyChecks()), WithDashboard())
	if err != nil {
		t.Fatalf("mount: %v", err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if !mounted.Ready() {
		t.Fatal("mounted must report ready after a passing batch")
	}

	mounted.Drain()

	code, _, _ := getBody(t, server.URL+"/readyz", nil)
	if code != http.StatusServiceUnavailable {
		t.Errorf("GET /readyz after drain = %d, want 503", code)
	}

	if mounted.Ready() {
		t.Error("Ready must report false after drain")
	}
}

func TestMount_LifecycleGuardsAndIdempotence(t *testing.T) {
	t.Parallel()

	mounted, err := Mount(http.NewServeMux(), NewProbe(healthyChecks()), WithDashboard())
	if err != nil {
		t.Fatalf("mount: %v", err)
	}

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	err = mounted.Start(t.Context())
	if err == nil {
		t.Error("second Start must be rejected")
	}

	err = mounted.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	err = mounted.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("second shutdown: %v", err)
	}

	err = mounted.Start(t.Context())
	if err != nil {
		t.Fatalf("restart after shutdown: %v", err)
	}
}

func TestMount_StartPropagatesProbeValidationErrors(t *testing.T) {
	t.Parallel()

	mounted, err := Mount(http.NewServeMux(), NewProbe(healthyChecks(), health.WithTimeout(-1*time.Second)))
	if err != nil {
		t.Fatalf("mount: %v", err)
	}

	err = mounted.Start(t.Context())
	if err == nil {
		t.Fatal("invalid probe configuration must fail Start")
	}
}
