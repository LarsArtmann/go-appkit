package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// APIKeyHeader is the HTTP header carrying the shared secret for machine
// clients (timers, scripts, calendar clients).
//
//nolint:gosec // G101 false positive: this is the header NAME, not a credential
const APIKeyHeader = "X-Api-Key"

// APIKeyQueryParam is the query parameter carrying the key for clients that
// cannot set headers on a navigation (a browser link click, a calendar
// client's ICS poll). Only honored on safe methods — see APIKeyAuth.
//

const APIKeyQueryParam = "key"

// APIKeyAuth guards routes with a shared secret.
//
// When apiKey is empty the guard is inactive (local development, tests).
// When set, requests must present the exact key via one of two credentials:
//
//  1. the X-Api-Key header — any method. The machine-client path.
//  2. the ?key= query parameter — safe methods (GET/HEAD) only. Browser
//     navigation cannot set headers on a link click. Mutating methods stay
//     header-only so the key never rides a URL on a write (query strings
//     land in access logs).
//
// A PRESENTED but WRONG credential fails closed: it is never rescued by
// anything else, and there is no fallback credential — a request without
// any key is rejected with 401.
//
// The comparison hashes both values and uses constant-time comparison so
// neither the key's content nor its length leaks through timing.
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	expectedHash := sha256.Sum256([]byte(apiKey))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				next.ServeHTTP(w, r)

				return
			}

			provided := r.Header.Get(APIKeyHeader)
			if provided == "" && methodIsSafe(r.Method) {
				provided = r.URL.Query().Get(APIKeyQueryParam)
			}

			providedHash := sha256.Sum256([]byte(provided))
			if provided != "" && subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1 {
				next.ServeHTTP(w, r)

				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(
				[]byte(`{"error":"missing or invalid ` + APIKeyHeader + ` header (GET links may use ?key=" + "only)"}`),
			)
		})
	}
}

// APIKeyMatches compares a presented credential against the configured key
// in constant time (same hashing discipline as APIKeyAuth), so a custom
// login or rotation endpoint can never drift from the guard.
func APIKeyMatches(provided, apiKey string) bool {
	if apiKey == "" || provided == "" {
		return false
	}

	providedHash := sha256.Sum256([]byte(provided))
	expectedHash := sha256.Sum256([]byte(apiKey))

	return subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1
}

// methodIsSafe reports whether the method is read-only per RFC 9110 — the
// methods a plain browser navigation or calendar poll issues.
func methodIsSafe(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}
