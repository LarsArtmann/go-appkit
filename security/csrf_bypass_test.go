package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

// noopCSRF stands in for a CSRF middleware in composition tests: it rejects
// requests that do not carry the bypass header, proving the bypass routing.
func noopCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// TestAPIKeyCSRFBypass_HeaderSkipsTokenDance pins the bypass: a request
// bearing the API key header skips the CSRF middleware entirely.
func TestAPIKeyCSRFBypass_HeaderSkipsTokenDance(t *testing.T) {
	t.Parallel()

	chain := security.APIKeyCSRFBypass(noopCSRF)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/write", nil)
	req.Header.Set(security.APIKeyHeader, testKey)

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("header-bearing POST: status = %d, want 200 (bypassed CSRF)", rec.Code)
	}
}

// TestAPIKeyCSRFBypass_StillValidates pins fail-closed: the bypass only
// SKIPS the token dance — the authentication still happens downstream, so
// the composed chain rejects a header-bearing request with a wrong key.
func TestAPIKeyCSRFBypass_StillValidates(t *testing.T) {
	t.Parallel()

	guard := security.APIKeyAuth(testKey)
	chain := security.APIKeyCSRFBypass(noopCSRF)(guard(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/write", nil)
	req.Header.Set(security.APIKeyHeader, "wrong-key")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong-key POST through bypass: status = %d, want 401 (bypass skips CSRF, never auth)", rec.Code)
	}
}

func TestAPIKeyCSRFBypass_WithoutHeaderKeepsCSRF(t *testing.T) {
	t.Parallel()

	chain := security.APIKeyCSRFBypass(noopCSRF)(http.NotFoundHandler())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/write", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("keyless POST: status = %d, want 403 (CSRF dance still runs)", rec.Code)
	}
}
