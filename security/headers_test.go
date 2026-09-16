package security_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

func headerOf(middleware func(http.Handler) http.Handler, name string) string {
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec.Header().Get(name)
}

// TestSecurityHeaders_HSTSOffOutsideProduction pins the LAN-over-http
// lockout trap: a browser that once saw HSTS on a plain-http LAN host
// refuses it for the max-age window. Only production may emit HSTS.
func TestSecurityHeaders_HSTSOffOutsideProduction(t *testing.T) {
	t.Parallel()

	for _, env := range []security.Environment{security.Development, security.Staging} {
		if got := headerOf(
			security.SecurityHeaders(security.HeadersConfig{Environment: env}),
			"Strict-Transport-Security",
		); got != "" {
			t.Errorf("%s: Strict-Transport-Security = %q, want empty (HSTS off outside production)", env, got)
		}
	}

	// An explicit HSTS override must NOT leak into non-production either.
	mw := security.SecurityHeaders(security.HeadersConfig{
		Environment: security.Development,
		HSTS:        "max-age=1",
	})

	if got := headerOf(mw, "Strict-Transport-Security"); got != "" {
		t.Errorf("dev with explicit HSTS override: got %q, want empty", got)
	}
}

func TestSecurityHeaders_HSTSOnInProduction(t *testing.T) {
	t.Parallel()

	got := headerOf(
		security.SecurityHeaders(security.HeadersConfig{Environment: security.Production}),
		"Strict-Transport-Security",
	)
	if got != "max-age=63072000; includeSubDomains; preload" {
		t.Errorf("production HSTS = %q", got)
	}

	custom := headerOf(security.SecurityHeaders(security.HeadersConfig{
		Environment: security.Production,
		HSTS:        "max-age=300",
	}), "Strict-Transport-Security")
	if custom != "max-age=300" {
		t.Errorf("production custom HSTS = %q", custom)
	}
}

func TestSecurityHeaders_StandardHeadersPresent(t *testing.T) {
	t.Parallel()

	if got := headerOf(
		security.SecurityHeaders(security.HeadersConfig{Environment: security.Development}),
		"X-Content-Type-Options",
	); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}

	if got := headerOf(
		security.SecurityHeaders(security.HeadersConfig{Environment: security.Development}),
		"X-Frame-Options",
	); got == "" {
		t.Error("X-Frame-Options missing")
	}
}

func TestSecurityHeaders_CSPPassedThrough(t *testing.T) {
	t.Parallel()

	got := headerOf(security.SecurityHeaders(security.HeadersConfig{
		Environment:           security.Production,
		ContentSecurityPolicy: "default-src 'self'",
	}), "Content-Security-Policy")

	if got != "default-src 'self'" {
		t.Errorf("Content-Security-Policy = %q", got)
	}
}
