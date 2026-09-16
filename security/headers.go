package security

import (
	"net/http"

	"github.com/larsartmann/httputil"
)

// productionHSTS is the HSTS value for production per the battery spec:
// two-year max-age with subdomains and preload. Dev and staging MUST leave
// HSTS off — a browser that once saw HSTS on a LAN-over-http host refuses
// to talk to it in plaintext afterwards (the lockout trap).
const productionHSTS = "max-age=63072000; includeSubDomains; preload"

// HeadersConfig tunes the environment-aware security-headers middleware.
type HeadersConfig struct {
	// Environment selects HSTS behavior: Production sets the two-year
	// includeSubDomains+preload HSTS header; Development and Staging force
	// it off regardless of what hsts says (an empty override wins —
	// explicit is better than an accidental lockout).
	Environment Environment

	// HSTS overrides the production HSTS value. Only honored when
	// Environment is Production.
	HSTS string

	// ContentSecurityPolicy sets the CSP. Empty means no CSP header — pair
	// with [BuildCSP] + the nonce middleware when you want one.
	ContentSecurityPolicy string
}

// SecurityHeaders returns environment-tuned security headers, delegating to
// httputil's [httputil.SecurityHeaders] for the standard header set
// (nosniff, frame options, referrer policy, permissions policy).
//
// The environment half of the contract: HSTS is a PRODUCTION-ONLY header.
// Shipping it from a laptop or a staging box over plain http locks that
// host out of the affected browsers for the max-age window.
func SecurityHeaders(cfg HeadersConfig) func(http.Handler) http.Handler {
	base := httputil.DefaultSecurityHeadersConfig()
	base.ContentSecurityPolicy = cfg.ContentSecurityPolicy

	switch cfg.Environment {
	case Production:
		base.StrictTransportSecurity = productionHSTS
		if cfg.HSTS != "" {
			base.StrictTransportSecurity = cfg.HSTS
		}
	case Development, Staging:
		base.StrictTransportSecurity = ""
	default:
		base.StrictTransportSecurity = ""
	}

	return httputil.SecurityHeaders(base)
}
