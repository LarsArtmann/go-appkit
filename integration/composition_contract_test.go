package integration_test

// The core composition-contract suite: the v0.5.0 surfaces (metrics,
// /version, testkit) must compose with the DEFAULT health contract and the
// drain-phase ordering, tested against the PUBLISHED core tag so the pins
// fail if a consumer-resolvable version breaks the contract.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/testkit"
)

// getWithAuth performs a GET against a base URL, optionally with Basic
// Auth credentials, and returns the status, content type, and body.
// Callers capture the base URL BEFORE shutting down: Service.Shutdown nils
// the listener reference first, so svc.Addr() is already nil inside
// DrainHooks — a contract this suite pins intentionally.
func getWithAuth(
	t *testing.T,
	base, path string,
	user, pass string,
) (int, string, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, base+path, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if user != "" {
		req.SetBasicAuth(user, pass)
	}

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, resp.Header.Get("Content-Type"), string(body)
}

// TestVersionAndMetricsComposeWithDefaultHealth pins that the opt-in
// surfaces coexist with the default health endpoints in ONE service: no
// route conflicts, correct response shapes, mandatory Basic Auth, and the
// stable metric-name contract — all through the full middleware chain via
// testkit.Serve (which also pins the testkit sub-package itself).
func TestVersionAndMetricsComposeWithDefaultHealth(t *testing.T) {
	t.Parallel()

	const (
		buildVersion   = "integration-v0.5.0"
		metricsUser    = "scraper"
		metricsPass    = "hunter2"
	)

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:       freeAddr(t),
		DrainDelay: appkit.NoDrainDelay,
		Version:    buildVersion,
		Metrics: &appkit.MetricsConfig{
			BasicAuthUser: metricsUser,
			BasicAuthPass: metricsPass,
		},
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	testkit.Serve(t, svc)

	base := "http://" + svc.Addr().String()

	status, contentType, body := getWithAuth(t, base, "/health/live", "", "")
	if status != http.StatusOK {
		t.Errorf("/health/live status = %d, want 200", status)
	}

	status, _, _ = getWithAuth(t, base, "/health/ready", "", "")
	if status != http.StatusOK {
		t.Errorf("/health/ready status = %d, want 200 while running", status)
	}

	status, contentType, body = getWithAuth(t, base, "/version", "", "")
	if status != http.StatusOK {
		t.Errorf("/version status = %d, want 200", status)
	}

	if contentType != "application/json" {
		t.Errorf("/version content type = %q, want application/json", contentType)
	}

	if want := fmt.Sprintf(`{"version":%q}`, buildVersion); body != want {
		t.Errorf("/version body = %q, want %q", body, want)
	}

	status, _, _ = getWithAuth(t, base, "/metrics", "", "")
	if status != http.StatusUnauthorized {
		t.Errorf("/metrics without credentials status = %d, want 401", status)
	}

	status, contentType, body = getWithAuth(t, base, "/metrics", metricsUser, metricsPass)
	if status != http.StatusOK {
		t.Errorf("/metrics status = %d, want 200", status)
	}

	if !strings.HasPrefix(contentType, "text/plain; version=0.0.4") {
		t.Errorf("/metrics content type = %q, want the Prometheus text exposition type", contentType)
	}

	for _, metric := range []string{
		"appkit_build_info",
		"appkit_http_requests_in_flight",
		"appkit_http_responses_total",
	} {
		if !strings.Contains(body, metric) {
			t.Errorf("/metrics exposition missing contract metric %q", metric)
		}
	}

	if !strings.Contains(body, fmt.Sprintf("appkit_build_info{version=%q}", buildVersion)) {
		t.Errorf("/metrics build info does not carry the configured version:\n%s", body)
	}

	// A served request must land in the response counter with its status.
	status, _, _ = getWithAuth(t, base, "/version", "", "")
	if status != http.StatusOK {
		t.Fatalf("/version re-request status = %d, want 200", status)
	}

	_, _, body = getWithAuth(t, base, "/metrics", metricsUser, metricsPass)
	if !strings.Contains(body, `status="200"`) {
		t.Errorf("/metrics responses_total missing a status=\"200\" series:\n%s", body)
	}
}

// TestDrainWindowContract pins the drain-phase ordering end to end: the
// readiness probe flips BEFORE DrainHooks run, in-flight traffic is still
// served DURING the drain window, and ShutdownHooks run only AFTER the
// connections are released. This is the lockstep contract the health module
// composes with (go-health readiness 503 for the whole drain window).
func TestDrainWindowContract(t *testing.T) {
	t.Parallel()

	type hookObservation struct {
		readyStatus int
		pingStatus  int
	}

	hookRan := make(chan hookObservation, 1)
	shutdownHookRan := false

	var svc *appkit.Service

	var baseURL string

	var err error

	svc, err = appkit.NewService(appkit.ServiceConfig{
		Addr:       freeAddr(t),
		DrainDelay: 250 * time.Millisecond,
		DrainHooks: []func(context.Context) error{
			func(_ context.Context) error {
				readyStatus, _, _ := getWithAuth(t, baseURL, "/health/ready", "", "")
				pingStatus, _, _ := getWithAuth(t, baseURL, "/ping", "", "")
				hookRan <- hookObservation{readyStatus: readyStatus, pingStatus: pingStatus}

				return nil
			},
		},
		ShutdownHooks: []func(context.Context) error{
			func(_ context.Context) error {
				shutdownHookRan = true

				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testkit.Serve(t, svc)

	baseURL = "http://" + svc.Addr().String()

	if status, _, _ := getWithAuth(t, baseURL, "/ping", "", ""); status != http.StatusOK {
		t.Fatalf("/ping before shutdown status = %d, want 200", status)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	select {
	case obs := <-hookRan:
		if obs.readyStatus != http.StatusServiceUnavailable {
			t.Errorf("readiness during drain hook = %d, want 503 (flip precedes DrainHooks)", obs.readyStatus)
		}

		if obs.pingStatus != http.StatusOK {
			t.Errorf("/ping during drain hook = %d, want 200 (traffic served during drain window)", obs.pingStatus)
		}
	default:
		t.Fatal("DrainHook never ran during Shutdown")
	}

	if !shutdownHookRan {
		t.Error("ShutdownHook did not run")
	}

	if svc.Running() {
		t.Error("service still Running after Shutdown returned")
	}

	pingReq, pingErr := http.NewRequestWithContext(
		context.Background(), http.MethodGet, baseURL+"/ping", nil)
	if pingErr != nil {
		t.Fatalf("rebuild ping request: %v", pingErr)
	}

	_, pingErr = (&http.Client{Timeout: 2 * time.Second}).Do(pingReq)
	if pingErr == nil {
		t.Error("request after Shutdown succeeded, want a connection error (listener released)")
	}
}
