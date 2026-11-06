package preset_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/preset"
)

// getNotFound fetches an unregistered path through the full chain and
// returns status, Content-Type, and body — no *http.Response crosses the
// boundary, so nothing needs closing at call sites.
func getNotFound(t *testing.T, url, accept string) (int, string, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url+"/definitely-not-registered", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := (&http.Client{Timeout: httpTimeout}).Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, resp.Header.Get("Content-Type"), string(body)
}

// TestErrorPages_Pretty404Default: the leg is ON by default (spine member) —
// unknown paths serve the pretty HTML page.
func TestErrorPages_Pretty404Default(t *testing.T) {
	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	status, contentType, body := getNotFound(t, ts, "")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}

	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}

	if !strings.Contains(body, "Page not found") {
		t.Errorf("404 body %q is not the pretty page", body[:min(len(body), 200)])
	}
}

// TestErrorPages_JSONByAccept: the default JSONWhen serves the JSON contract
// when the request accepts application/json.
func TestErrorPages_JSONByAccept(t *testing.T) {
	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	status, contentType, body := getNotFound(t, ts, "application/json")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}

	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	if !strings.Contains(body, "not_found") {
		t.Errorf("JSON body %q has no not_found code", body)
	}
}

// TestErrorPages_CodeDisabled: Enabled=false removes the catch-all — the
// stdlib mux's plain 404 serves instead.
func TestErrorPages_CodeDisabled(t *testing.T) {
	cfg := testConfig()
	cfg.ErrorPages = &preset.ErrorPagesConfig{Enabled: false}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	status, contentType, body := getNotFound(t, ts, "")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}

	if strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want the stdlib plain 404", contentType)
	}

	if !strings.Contains(body, "404 page not found") {
		t.Errorf("body %q is not the stdlib 404", body)
	}
}

// TestErrorPages_EnvOff: APPKIT_ERRORPAGES=off disables the spine member
// without code.
func TestErrorPages_EnvOff(t *testing.T) {
	t.Setenv("APPKIT_ERRORPAGES", "off")

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	status, _, body := getNotFound(t, ts, "")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}

	if !strings.Contains(body, "404 page not found") {
		t.Errorf("body %q is not the stdlib 404 (catch-all still mounted)", body)
	}
}

// TestErrorPages_ConsumerRouteWins: routes registered after New are more
// specific than the catch-all "/" and win per stdlib ServeMux precedence.
func TestErrorPages_ConsumerRouteWins(t *testing.T) {
	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("consumer route"))
	})

	ts := servePreset(t, svc)

	status, body := getURL(t, ts+"/hello")

	if status != http.StatusOK || body != "consumer route" {
		t.Errorf("/hello = %d %q, want 200 %q", status, body, "consumer route")
	}
}
