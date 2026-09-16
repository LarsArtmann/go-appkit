// Package security provides opt-in HTTP security batteries for appkit
// services: API-key authentication, CSRF protection and its API-key bypass,
// keyed rate-limit profiles, origin checks, request-body limits, HTML/URL
// sanitization, CSP nonce infrastructure with a policy builder, and
// environment-tuned security headers.
//
// # Opt-in only
//
// Nothing in this package joins any default middleware stack. Every battery
// is an explicit middleware (func(http.Handler) http.Handler) or function
// the consumer wires per route or per service.
//
// # Quick start
//
//	key := os.Getenv("API_KEY") // empty key = guard inactive (dev/test)
//	guarded := security.APIKeyAuth(key)(handler)
//	limited := security.RateLimit(security.ExportProfile)(guarded)
//	http.Handle("POST /api/export", limited)
//
// # Provenance
//
// The batteries are ports of the CV family stack's production scar tissue
// (github.com/LarsArtmann/CV platform/middleware + internal/sanitization);
// the port-from inventory and acceptance tests are specified in
// go-appkit's doc/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md
// (section W2). Each file's doc comment carries the threat model that
// motivated it.
package security
