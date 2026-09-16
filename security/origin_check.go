package security

import (
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// OriginCheck returns middleware that validates the request's Origin header
// (falling back to the Referer's origin) against the allowlist.
//
// Defense-in-depth CSRF protection for browser-submitted forms: it
// complements a token-based CSRF middleware with an explicit, route-level
// origin check that does not depend on cookie state.
//
// Semantics (CV production contract):
//   - same-origin requests are ALWAYS allowed — browsers attach Origin to
//     every non-GET fetch, including same-origin ones, so without this rule
//     action buttons 403 on any host not explicitly allowlisted
//     (e.g. http://localhost:8080);
//   - cross-origin requests must appear in the allowlist (case-insensitive);
//   - requests with NO Origin and NO Referer pass through — non-browser
//     clients (curl, machine clients) send neither, and the token-based
//     CSRF layer owns their validation;
//   - an empty allowlist or a "*" entry means allow-all: this is a
//     deliberate operator decision, logged once at construction so a
//     placeholder "*" never silently reaches production.
func OriginCheck(allowedOrigins []string, logger *slog.Logger) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 0 || slices.Contains(allowedOrigins, "*")
	if allowAll && logger != nil {
		logger.Warn("security.OriginCheck: allow-all origins configured (empty list or '*') — defense-in-depth disabled")
	}

	normalized := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		normalized[strings.ToLower(origin)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowAll {
				next.ServeHTTP(w, r)

				return
			}

			origin := extractRequestOrigin(r)
			if origin == "" {
				next.ServeHTTP(w, r)

				return
			}

			if _, ok := normalized[strings.ToLower(origin)]; ok {
				next.ServeHTTP(w, r)

				return
			}

			if sameOrigin(r, origin) {
				next.ServeHTTP(w, r)

				return
			}

			if logger != nil {
				logger.Warn("security.OriginCheck: origin rejected",
					"origin", origin, "path", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":"ORIGIN_REJECTED","message":"The origin of this request is not allowed"}}`))
		})
	}
}

// sameOrigin reports whether origin matches the host the request was served
// on. The Origin header is browser-controlled (it reflects the real page
// origin and cannot be forged from JavaScript), so a host match proves the
// request came from a page this server itself served.
func sameOrigin(r *http.Request, origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}

	return strings.EqualFold(u.Host, r.Host)
}

// extractRequestOrigin returns the request's Origin header, falling back to
// the scheme://host extracted from the Referer header. Returns "" if
// neither is present.
func extractRequestOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin != "" {
		return origin
	}

	referer := r.Header.Get("Referer")
	if referer == "" {
		return ""
	}

	u, err := url.Parse(referer)
	if err != nil {
		return referer
	}

	return u.Scheme + "://" + u.Host
}
