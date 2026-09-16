package security

import (
	"slices"
	"strings"
)

// scriptSrcDirective is the CSP directive controlling script execution.
const scriptSrcDirective = "script-src"

// selfSrc is the CSP 'self' source token.
const selfSrc = "'self'"

// Environment names the deployment tier a CSP policy is built for.
type Environment string

// Deployment tiers. Production is the strict tier; Development and Staging
// exist so policies can be TESTED locally without weakening the production
// shape (see BuildCSP — the eval carve-out does not exist in any tier).
const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// CSPConfig configures [BuildCSP].
type CSPConfig struct {
	// Environment selects the strictness tier. Production emits
	// `default-src 'self'; script-src 'self' 'nonce-…'; …`. Development and
	// Staging additionally allow ws:/wss: connects (live-reload dashboards).
	Environment Environment

	// Nonce is the per-request nonce minted by [GenerateNonce]. When empty,
	// the script-src carries no nonce source.
	Nonce string

	// StyleInline grants `style-src 'unsafe-inline'`. This stays grantable
	// BY DECISION in any tier: dynamic inline styles (JS-set style
	// attributes, component libraries) cannot be hash-pinned, and
	// style-based injection is materially less dangerous than
	// script-based injection. Script eval is NOT symmetrically grantable —
	// see the eval note on BuildCSP.
	StyleInline bool

	// ConnectSrc extends `connect-src` beyond 'self' (API hosts, SSE
	// endpoints on another origin). Empty means 'self'.
	ConnectSrc []string

	// ImgSrc extends `img-src` beyond 'self' (data: URIs, CDN hosts).
	// Empty means 'self' plus `data:`.
	ImgSrc []string

	// JSONLD allows `<script type="application/ld+json">` blocks by adding
	// the JSON-LD media type to script-src. JSON-LD is data, not code —
	// browsers do not execute it — so the exemption is safe and standard.
	JSONLD bool
}

// BuildCSP assembles a deterministic Content-Security-Policy string.
//
// THE eval RULE (the family's most expensive CSP lesson): DataStar/Alpine-
// style clients compile `data-*` expressions via `new Function()`, which is
// dead under `script-src 'self'` AND under `script-src *` — a wildcard does
// NOT imply eval, because `'unsafe-eval'` is its own source token. Whole
// dashboard classes died on this (CV 2026-08-16). There is deliberately NO
// option to emit 'unsafe-eval' here, in ANY environment: if a client
// genuinely needs eval, it must be a reviewed, hand-written policy — not a
// builder flag.
//
// Directives are emitted in a fixed, sorted order so the header is
// byte-identical across requests and parseable by policy tests.
func BuildCSP(cfg CSPConfig) string {
	const directiveCount = 9 // default, base-uri, form-action, frame-ancestors, object-src, script-src, style-src, connect-src, img-src

	directives := make([]string, 0, directiveCount)
	directives = append(directives,
		"default-src "+selfSrc,
		"base-uri "+selfSrc,
		"form-action "+selfSrc,
		"frame-ancestors 'none'",
		"object-src 'none'",
	)

	scriptSrc := []string{selfSrc}
	if cfg.Nonce != "" {
		scriptSrc = append(scriptSrc, fmtNonce(cfg.Nonce))
	}

	if cfg.JSONLD {
		scriptSrc = append(scriptSrc, "'application/ld+json'")
	}

	directives = append(directives, "script-src "+strings.Join(scriptSrc, " "))

	styleSrc := []string{selfSrc}
	if cfg.StyleInline {
		styleSrc = append(styleSrc, "'unsafe-inline'")
	}

	directives = append(directives, "style-src "+strings.Join(styleSrc, " "))

	connectSrc := []string{selfSrc}

	switch cfg.Environment {
	case Development, Staging:
		connectSrc = append(connectSrc, "ws:", "wss:")
	case Production:
	default:
	}

	connectSrc = append(connectSrc, cfg.ConnectSrc...)
	directives = append(directives, "connect-src "+strings.Join(connectSrc, " "))

	imgSrc := make([]string, 0, 2+len(cfg.ImgSrc))
	imgSrc = append(imgSrc, selfSrc, "data:")
	imgSrc = append(imgSrc, cfg.ImgSrc...)
	directives = append(directives, "img-src "+strings.Join(imgSrc, " "))

	slices.Sort(directives)

	return strings.Join(directives, "; ")
}

// fmtNonce renders the nonce token in CSP source-list syntax.
func fmtNonce(nonce string) string {
	return "'nonce-" + nonce + "'"
}
