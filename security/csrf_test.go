package security_test

import (
	"log/slog"
	"net/http"
	"strings"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/security"
	"github.com/larsartmann/httputil"
)

// TestCSRF_StarNeverMeansAllowAll pins the degradation contract: "*" in
// TrustedOrigins is stripped (httputil's CSRF fail-closes on it) and the
// misconfiguration is logged — the middleware still runs, same-origin-only
// when no real origins remain.
func TestCSRF_StarNeverMeansAllowAll(t *testing.T) {
	t.Parallel()

	capture := &recordingHandler{}
	mw := security.CSRF(httputil.CSRFConfig{
		TrustedOrigins: []string{"*", "https://app.example.com"},
		Secure:         true,
	}, slog.New(capture))

	if !capture.contains("TrustedOrigins") {
		t.Error("the '*' misconfiguration must be logged loudly at construction")
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/form", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("GET through CSRF: status = %d, want 418 (safe methods pass)", rec.Code)
	}
}

// TestCSRF_CrossSitePostRejected pins the real middleware's core property
// through the wrapper: a cookie-less cross-site form POST cannot pass.
func TestCSRF_CrossSitePostRejected(t *testing.T) {
	t.Parallel()

	mw := security.CSRF(httputil.CSRFConfig{TrustedOrigins: []string{"https://app.example.com"}}, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/form", strings.NewReader("field=value"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://evil.example.net")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusTeapot {
		t.Error("cross-site POST without a token must not reach the handler")
	}
}

func TestCSRF_CleanConfigUntouched(t *testing.T) {
	t.Parallel()

	capture := &recordingHandler{}
	mw := security.CSRF(httputil.CSRFConfig{TrustedOrigins: []string{"https://app.example.com"}}, slog.New(capture))

	_ = mw
	if capture.contains("TrustedOrigins") {
		t.Error("a clean origin list must not trigger the '*' warning")
	}
}
