package security

import (
	"net"
	"net/http"
	"time"

	"github.com/larsartmann/httputil"
)

// limiterTTL is how long an idle bucket is kept before eviction.
const limiterTTL = 10 * time.Minute

// RateLimitConfig configures one keyed rate-limit profile.
type RateLimitConfig struct {
	// Limit is the maximum requests per Window per key.
	Limit uint

	// Burst is the maximum burst above Limit. Zero defaults to Limit.
	Burst uint

	// Window is the window length. Zero defaults to 1 minute.
	Window time.Duration

	// MaxKeys caps how many distinct keys the limiter tracks. This field is
	// a HARD requirement, not a tuning knob: a keyed limiter without a key
	// cap is a memory-exhaustion DoS vector (every distinct attacker IP
	// mints a bucket). CV caps key-space per profile at 5k–10k.
	MaxKeys uint

	// KeyExtractor produces the bucket key. Nil defaults to the RemoteAddr
	// host part, so the key is never "" (an empty key silently disables
	// per-client limits). Deployments behind a trusted proxy should extract
	// the client IP the same way their ClientIP middleware does — an
	// extractor that trusted X-Forwarded-For unconditionally would let any
	// client forge itself into fresh buckets (shared-bucket semantics live
	// at the extractor, by decision).
	KeyExtractor httputil.KeyExtractor
}

// Named profiles (per-minute request limits from the CV production stack).
// They are VALUES, not singletons: build one RateLimit per endpoint group
// that needs isolation — two endpoints sharing one built middleware share
// one bucket set, so a busy group can starve another.
//
//nolint:gochecknoglobals // public preset values, immutable by convention
var (
	// GeneralProfile fits broad endpoints (health, root, metrics):
	// 60 req/min, burst 100.
	GeneralProfile = RateLimitConfig{
		Limit:   60,
		Burst:   100,
		MaxKeys: 10_000,
	} //nolint:exhaustruct_v5,mnd // documented preset; Window/KeyExtractor defaulted by RateLimit

	// AnalysisProfile fits interactive analysis endpoints:
	// 10 req/min, burst 15.
	AnalysisProfile = RateLimitConfig{
		Limit:   10,
		Burst:   15,
		MaxKeys: 5_000,
	} //nolint:exhaustruct_v5,mnd // documented preset; Window/KeyExtractor defaulted by RateLimit

	// ExportProfile fits heavy export endpoints (PDF/report generation):
	// 5 req/min, burst 8.
	ExportProfile = RateLimitConfig{
		Limit:   5,
		Burst:   8,
		MaxKeys: 5_000,
	} //nolint:exhaustruct_v5,mnd // documented preset; Window/KeyExtractor defaulted by RateLimit

	// ContactProfile fits form submissions: 8 req/min, burst 10.
	ContactProfile = RateLimitConfig{
		Limit:   8,
		Burst:   10,
		MaxKeys: 5_000,
	} //nolint:exhaustruct_v5,mnd // documented preset; Window/KeyExtractor defaulted by RateLimit
)

// RateLimit builds a keyed rate-limit middleware from a profile.
//
// Contract: when the limit trips, the middleware writes 429 with a
// Retry-After header and the chain ABORTS — next is never invoked, so no
// later handler can overwrite the 429 with a 200 (CV's 2026-08-16 bug made
// every limit bypassable that way; the regression test pins it).
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	window := cfg.Window
	if window == 0 {
		window = time.Minute
	}

	keyExtractor := cfg.KeyExtractor
	if keyExtractor == nil {
		keyExtractor = remoteAddrHostKey
	}

	return httputil.KeyedRateLimiterMiddleware(
		httputil.KeyedRateLimiterConfig{ //nolint:exhaustruct_v5 // event hooks are optional
			Limit:        cfg.Limit,
			Window:       window,
			Burst:        cfg.Burst,
			KeyExtractor: keyExtractor,
			TTL:          limiterTTL,
			MaxKeys:      cfg.MaxKeys,
		},
	)
}

// remoteAddrHostKey extracts the host part of RemoteAddr, falling back to
// the raw address when it carries no port.
func remoteAddrHostKey(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}
