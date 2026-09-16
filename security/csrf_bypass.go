package security

import "net/http"

// APIKeyCSRFBypass wraps a CSRF middleware so requests bearing the
// [APIKeyHeader] skip CSRF validation.
//
// The threat model, in full: CSRF protects against cross-site requests that
// ride AMBIENT credentials (cookies the browser attaches automatically).
// The X-Api-Key header is a NON-ambient credential: a cross-site attacker
// page cannot set it — custom headers on cross-origin fetches trigger a CORS
// preflight that the CORS layer denies, and HTML form posts cannot carry
// custom headers at all. A request that DOES present the header is by
// construction a programmatic client, and its key is still validated
// (fail-closed) by [APIKeyAuth] in the chain — this bypass only skips the
// token dance, never the authentication. Compose guard-inside-bypass:
//
//	chain := security.APIKeyCSRFBypass(csrfMiddleware)(security.APIKeyAuth(key)(handler))
//
// Safe methods are never special-cased here: a CSRF middleware passes them
// through anyway (nosurf validates only unsafe methods).
func APIKeyCSRFBypass(csrf func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		guarded := csrf(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(APIKeyHeader) != "" {
				next.ServeHTTP(w, r)

				return
			}

			guarded.ServeHTTP(w, r)
		})
	}
}
