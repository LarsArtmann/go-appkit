package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

const testKey = "test-secret-key-0123456789"

func serveWithAPIKey(t *testing.T, middleware func(http.Handler) http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	var reached bool
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !reached && rec.Code == http.StatusOK {
		t.Fatal("handler not reached but response was 200 — a middleware returned empty-200 without calling Next()")
	}

	return rec
}

// TestAPIKeyAuth_AcceptsCorrectKey also guards the missing-Next() empty-200
// trap: a middleware that returns without calling Next AND without writing
// a status leaves the recorder at 200, indistinguishable from success.
func TestAPIKeyAuth_AcceptsCorrectKey(t *testing.T) {
	t.Parallel()

	mw := security.APIKeyAuth(testKey)

	rec := serveWithAPIKey(t, mw, http.MethodGet, "/resource")
	if rec.Code != http.StatusOK {
		t.Errorf("header GET: status = %d, want 200", rec.Code)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/resource", nil)
	req.Header.Set(security.APIKeyHeader, testKey)
	rec2 := httptest.NewRecorder()
	mw(http.NotFoundHandler()).ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Errorf("header POST: status = %d, want 200", rec2.Code)
	}
}

// TestAPIKeyAuth_QueryKeyRejectedOnPost pins the access-log leak rule: the
// ?key= query fallback is safe-methods-only, so a key on a mutating request
// URL is rejected, never honored.
func TestAPIKeyAuth_QueryKeyRejectedOnPost(t *testing.T) {
	t.Parallel()

	mw := security.APIKeyAuth(testKey)

	rec := serveWithAPIKey(t, mw, http.MethodPost, "/resource?key="+testKey)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("query-key POST: status = %d, want 401 (keys must not ride URLs on writes)", rec.Code)
	}
}

func TestAPIKeyAuth_QueryKeyAcceptedOnGet(t *testing.T) {
	t.Parallel()

	mw := security.APIKeyAuth(testKey)

	rec := serveWithAPIKey(t, mw, http.MethodGet, "/resource?key="+testKey)
	if rec.Code != http.StatusOK {
		t.Errorf("query-key GET: status = %d, want 200", rec.Code)
	}
}

// TestAPIKeyAuth_HeaderWinsOverQuery pins the credential precedence: when
// both are present and the header is WRONG, the request fails even if the
// query key is correct — a presented credential is authoritative.
func TestAPIKeyAuth_HeaderWinsOverQuery(t *testing.T) {
	t.Parallel()

	mw := security.APIKeyAuth(testKey)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/resource?key="+testKey, nil)
	req.Header.Set(security.APIKeyHeader, "wrong-key")

	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong header + right query: status = %d, want 401", rec.Code)
	}
}

func TestAPIKeyAuth_EmptyKeyIsInactive(t *testing.T) {
	t.Parallel()

	mw := security.APIKeyAuth("")

	rec := serveWithAPIKey(t, mw, http.MethodPost, "/resource")
	if rec.Code != http.StatusOK {
		t.Errorf("empty key (dev mode): status = %d, want 200", rec.Code)
	}
}

func TestAPIKeyAuth_WrongKeyFailsClosed(t *testing.T) {
	t.Parallel()

	mw := security.APIKeyAuth(testKey)

	rec := serveWithAPIKey(t, mw, http.MethodGet, "/resource?key=wrong")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong query key: status = %d, want 401", rec.Code)
	}
}

func TestAPIKeyMatches(t *testing.T) {
	t.Parallel()

	if !security.APIKeyMatches(testKey, testKey) {
		t.Error("matching keys should compare equal")
	}

	if security.APIKeyMatches("wrong", testKey) {
		t.Error("wrong key must not match")
	}

	if security.APIKeyMatches(testKey, "") {
		t.Error("empty configured key must never match (inactive guard, not allow-all)")
	}
}
