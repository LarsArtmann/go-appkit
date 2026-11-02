// The security × errorpages composition: the pretty error pages served
// BEHIND the appkit default stack's SecurityHeaders middleware, with an
// APIKeyAuth-guarded route alongside the errorpages.Mount catch-all. Pins
// the contract as it ACTUALLY composes (source-verified against
// errorpages.Mount: it registers a catch-all that renders pretty 404s only —
// method mismatches keep the stdlib mux's plain 405 with Allow):
//
//   - auth-gated route, missing key → plain JSON 401 from APIKeyAuth — the
//     security guard's response is NOT prettified (errorpages owns 404s),
//     and the default-stack security headers are still on it.
//   - unknown path → pretty page from the catch-all, security headers
//     present (the "behind the security stack" pin), no auth required —
//     the catch-all sits outside the route-scoped APIKeyAuth wrapper.
//     The page is a styled <main> fragment (templ component), not a full
//     <html> document.
//   - unknown path with Accept: application/json → JSON contract, not HTML.
//   - method mismatch on a registered pattern → the catch-all SWALLOWS it
//     into a pretty 404: Mount registers "/" so the mux never reaches its
//     plain 405 path. Pretty 405s are Wrap's job (custom-server setups).
//   - classified handler error rendered via errorpages.Write → pretty page
//     with the error-family status and the security headers.

package integration_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	"github.com/larsartmann/go-appkit/security"
	"github.com/larsartmann/go-appkit/testkit"
	errorfamily "github.com/larsartmann/go-error-family"
)

func newSecuredErrorpagesService(t *testing.T) (*testkit.TestServer, string) {
	t.Helper()

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:       freeAddr(t),
		DrainDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	errorpages.Mount(svc.Mux, errorpages.Config{})

	apiKey := "itest-security-errorpages-key"

	greet := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	svc.Mux.Handle("GET /api", security.APIKeyAuth(apiKey)(greet))

	unclassified := errorfamily.NewInfrastructure(
		"itest.security_errorpages",
		"infrastructure failure behind the pretty pages",
	)

	svc.Mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		errorpages.Write(w, r, unclassified, errorpages.Config{})
	})

	ts := testkit.Serve(t, svc)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := ts.Shutdown(ctx); shutdownErr != nil {
			t.Errorf("shutdown: %v", shutdownErr)
		}
	})

	return ts, apiKey
}

func securityHeadersPresent(t *testing.T, header http.Header) {
	t.Helper()

	for name, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	} {
		if got := header.Get(name); got != want {
			t.Errorf("header %s = %q, want %q (default-stack SecurityHeaders must wrap this response)", name, got, want)
		}
	}
}

func TestErrorpagesBehindSecurityStack(t *testing.T) {
	t.Parallel()

	ts, apiKey := newSecuredErrorpagesService(t)
	base := ts.FullChainURL

	t.Run("authed route serves through the stack", func(t *testing.T) {
		t.Parallel()

		secErrAuthedRoute(t, base, apiKey)
	})

	t.Run("missing key stays a plain JSON 401", func(t *testing.T) {
		t.Parallel()

		secErrPlain401(t, base)
	})

	t.Run("unknown path renders the pretty 404 with security headers", func(t *testing.T) {
		t.Parallel()

		secErrPretty404(t, base)
	})

	t.Run("JSON contract for API clients", func(t *testing.T) {
		t.Parallel()

		secErrJSONContract(t, base)
	})

	t.Run("method mismatch is swallowed by the catch-all into a pretty 404", func(t *testing.T) {
		t.Parallel()

		secErrMethodMismatch(t, base)
	})

	t.Run("classified handler error renders pretty with error-family status", func(t *testing.T) {
		t.Parallel()

		secErrClassifiedError(t, base)
	})
}

func secErrAuthedRoute(t *testing.T, base, apiKey string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, base+"/api", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.Header.Set(security.APIKeyHeader, apiKey)

	authed, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("GET /api with key: %v", err)
	}
	defer func() { _ = authed.Body.Close() }()

	if authed.StatusCode != http.StatusOK {
		t.Fatalf("GET /api with key: status = %d, want 200", authed.StatusCode)
	}

	securityHeadersPresent(t, authed.Header)
}

func secErrPlain401(t *testing.T, base string) {
	t.Helper()

	resp := e2eGet(t, base+"/api")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("GET /api without key: status = %d, want 401", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(string(body), "<html") {
		t.Error("401 body is an HTML page — errorpages must NOT render auth rejections (it owns 404s only)")
	}

	if !strings.Contains(string(body), `"error"`) {
		t.Errorf("401 body = %q, want the APIKeyAuth JSON contract", string(body))
	}

	securityHeadersPresent(t, resp.Header)
}

func secErrPretty404(t *testing.T, base string) {
	t.Helper()

	resp := e2eGet(t, base+"/definitely-missing")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET unknown path: status = %d, want 404", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("404 Content-Type = %q, want text/html (pretty page)", ct)
	}

	body, _ := io.ReadAll(resp.Body)

	if !strings.Contains(string(body), "<main") {
		t.Errorf("404 body = %.80q, want the pretty <main> fragment", string(body))
	}

	securityHeadersPresent(t, resp.Header)
}

func secErrJSONContract(t *testing.T, base string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, base+"/definitely-missing", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("GET unknown path with Accept json: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("404 Content-Type = %q, want application/json for Accept: application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(string(body), "<html") {
		t.Error("404 body is HTML — the JSON contract must win for API clients")
	}
}

func secErrMethodMismatch(t *testing.T, base string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, base+"/api", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("POST /api: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"POST /api: status = %d, want 404 (Mount's \"/\" catch-all matches every method, so the mux's plain 405 never surfaces)",
			resp.StatusCode,
		)
	}

	body, _ := io.ReadAll(resp.Body)

	if !strings.Contains(string(body), "<main") {
		t.Errorf("POST /api body = %.80q, want the pretty <main> 404 fragment", string(body))
	}
}

func secErrClassifiedError(t *testing.T, base string) {
	t.Helper()

	resp := e2eGet(t, base+"/boom")
	defer func() { _ = resp.Body.Close() }()

	want := appkit.HTTPStatus(
		errorfamily.NewInfrastructure("itest.security_errorpages", "infrastructure failure behind the pretty pages"),
	)
	if resp.StatusCode != want {
		t.Fatalf("GET /boom: status = %d, want %d (error-family classification)", resp.StatusCode, want)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("error page Content-Type = %q, want text/html", ct)
	}

	securityHeadersPresent(t, resp.Header)
}
