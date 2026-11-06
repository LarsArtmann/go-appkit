package preset_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/preset"
)

// getNotFound fetches an unregistered path through the full chain.
func getNotFound(t *testing.T, url, accept string) (*http.Response, string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url+"/definitely-not-registered", nil)
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp, string(body)
}

// TestErrorPages_Pretty404Default: the leg is ON by default (spine member) —
// unknown paths serve the pretty HTML page.
func TestErrorPages_Pretty404Default(t *testing.T) {
	t.Parallel()

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp, body := getNotFound(t, ts, "")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %q, want text/html", resp.Header.Get("Content-Type"))
	}

	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("404 body is not the pretty HTML page")
	}
}

// TestErrorPages_JSONByAccept: the default JSONWhen serves the JSON contract
// when the request accepts application/json.
func TestErrorPages_JSONByAccept(t *testing.T) {
	t.Parallel()

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp, body := getNotFound(t, ts, "application/json")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		t.Errorf("Content-Type = %q, want application/json", resp.Header.Get("Content-Type"))
	}

	if !strings.Contains(body, "404") {
		t.Errorf("JSON body %q does not mention the status", body)
	}
}

// TestErrorPages_CodeDisabled: Enabled=false removes the catch-all — the
// stdlib mux's plain 404 serves instead.
func TestErrorPages_CodeDisabled(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.ErrorPages = &preset.ErrorPagesConfig{Enabled: false}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp, body := getNotFound(t, ts, "")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %q, want the stdlib plain 404", resp.Header.Get("Content-Type"))
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

	resp, body := getNotFound(t, ts, "")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}

	if !strings.Contains(body, "404 page not found") {
		t.Errorf("body %q is not the stdlib 404 (catch-all still mounted)", body)
	}
}

// TestErrorPages_ConsumerRouteWins: routes registered after New are more
// specific than the catch-all "/" and win per stdlib ServeMux precedence.
func TestErrorPages_ConsumerRouteWins(t *testing.T) {
	t.Parallel()

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("consumer route"))
	})

	ts := servePreset(t, svc)

	resp, err := http.Get(ts + "/hello")
	if err != nil {
		t.Fatalf("get /hello: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if resp.StatusCode != http.StatusOK || string(body) != "consumer route" {
		t.Errorf("/hello = %d %q, want 200 %q", resp.StatusCode, body, "consumer route")
	}
}
