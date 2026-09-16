package security_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/security"
)

// TestRateLimit_429NotOverwritten is THE regression: CV's 2026-08-16 bug
// made every limit bypassable because a later handler overwrote the 429
// with a 200. The chain must abort — a tripped limit stays a 429 no matter
// what the wrapped handler would write.
func TestRateLimit_429NotOverwritten(t *testing.T) {
	t.Parallel()

	limiter := security.RateLimit(security.RateLimitConfig{
		Limit:   1,
		Burst:   0,
		MaxKeys: 100,
	})

	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK) // the "buggy later handler" — must never win
	}))

	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	first := send()
	if first.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want 200", first.Code)
	}

	second := send()
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: status = %d, want 429", second.Code)
	}
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	t.Parallel()

	limiter := security.RateLimit(security.RateLimitConfig{Limit: 1, MaxKeys: 100})

	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for range 5 {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			retryAfter := rec.Header().Get("Retry-After")
			if retryAfter == "" {
				t.Fatal("429 response missing Retry-After header")
			}

			_, convErr := strconv.Atoi(retryAfter)
			if convErr != nil {
				t.Errorf("Retry-After = %q, want integer seconds", retryAfter)
			}

			return
		}
	}

	t.Fatal("never hit the limit")
}

func TestRateLimit_ProfilesCarryMandatoryMaxKeys(t *testing.T) {
	t.Parallel()

	// MaxKeys is a HARD requirement: a keyed limiter without a key cap is a
	// memory-exhaustion DoS vector. The named profiles must never ship with
	// a zero cap.
	for name, profile := range map[string]security.RateLimitConfig{
		"GeneralProfile":  security.GeneralProfile,
		"AnalysisProfile": security.AnalysisProfile,
		"ExportProfile":   security.ExportProfile,
		"ContactProfile":  security.ContactProfile,
	} {
		if profile.MaxKeys == 0 {
			t.Errorf("%s.MaxKeys = 0: a keyed limiter without a key cap is a DoS vector", name)
		}

		if profile.Limit == 0 {
			t.Errorf("%s.Limit = 0: a zero limit denies everything", name)
		}
	}
}

func TestRateLimit_KeysAreIsolated(t *testing.T) {
	t.Parallel()

	limiter := security.RateLimit(security.RateLimitConfig{Limit: 1, MaxKeys: 100})
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	exhaust := func(ip string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		req.RemoteAddr = ip + ":1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	if code := exhaust("10.1.0.1").Code; code != http.StatusOK {
		t.Fatalf("first client first request: %d, want 200", code)
	}

	if code := exhaust("10.1.0.2").Code; code != http.StatusOK {
		t.Errorf("second client first request: %d, want 200 (buckets must be per-key)", code)
	}

	if code := exhaust("10.1.0.1").Code; code != http.StatusTooManyRequests {
		t.Errorf("first client second request: %d, want 429", code)
	}
}

func TestRateLimit_WindowRespected(t *testing.T) {
	t.Parallel()

	// A tiny window proves the limit RESets (a limiter that never resets is
	// a permanent ban, not a rate limit).
	limiter := security.RateLimit(security.RateLimitConfig{Limit: 1, Window: 30 * time.Millisecond, MaxKeys: 100})
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	send := func() int {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		req.RemoteAddr = "10.2.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code
	}

	if code := send(); code != http.StatusOK {
		t.Fatalf("first: %d, want 200", code)
	}

	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("second (same window): %d, want 429", code)
	}
}
