package security

import (
	"log/slog"
	"slices"

	"github.com/larsartmann/httputil"
)

// CSRF builds the double-submit-cookie CSRF middleware (a thin, opinionated
// configuration of httputil's nosurf-based [httputil.CSRFMiddleware]).
//
// The subtle trap this wrapper documents: a Trusted-Origin header does NOT
// bypass the token check — cross-site allowed origins still need a valid
// token; the origin list only permits the cross-SITE flow at all.
//
// The "*" misconfiguration: httputil's CSRF REJECTS "*" in TrustedOrigins
// (fail-closed). This wrapper keeps that semantics — "*" never means
// allow-all. If "*" is present, the wrapper strips it, logs the
// misconfiguration loudly, and runs with the remaining origins (or
// same-origin-only when none remain): degrade-with-visibility, never
// allow-all.
func CSRF(cfg httputil.CSRFConfig, logger *slog.Logger) func(http.Handler) http.Handler {
	if slices.Contains(cfg.TrustedOrigins, "*") {
		origins := make([]string, 0, len(cfg.TrustedOrigins))
		for _, o := range cfg.TrustedOrigins {
			if o != "*" {
				origins = append(origins, o)
			}
		}

		cfg.TrustedOrigins = origins

		if logger != nil {
			logger.Warn("security.CSRF: \"*\" in TrustedOrigins is not allow-all; " +
				"stripped — the middleware runs with the remaining origins " +
				"(same-origin only when none remain)")
		}
	}

	return httputil.CSRFMiddleware(cfg)
}
