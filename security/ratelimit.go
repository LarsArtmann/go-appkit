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

// Named profile limits (per-minute request limits from the CV production
// stack). They are VALUES, not singletons: build one RateLimit per endpoint
// group that needs isolation — two endpoints sharing one built middleware
// share one bucket set, so a busy group can starve another.
const (
	generalLimit  uint = 60
	generalBurst  uint = 100
	generalCap    uint = 10_000
	analysisLimit uint = 10
	analysisBurst uint = 15
	analysisCap   uint = 5_000
	exportLimit   uint = 5
	exportBurst   uint = 8
	exportCap     uint = 5_000
	contactLimit  uint = 8
	contactBurst  uint = 10
	contactCap    uint = 5_000
)

// profile builds a config with every field explicit (Window and
// KeyExtractor zero means "RateLimit applies its defaults").
func profile(limit, burst, maxKeys uint) RateLimitConfig {
	return RateLimitConfig{
		Limit:        limit,
		Burst:        burst,
		Window:       0,
		MaxKeys:      maxKeys,
		KeyExtractor: nil,
	}
}

// GeneralProfile fits broad endpoints (health, root, metrics):
// 60 req/min, burst 100.
var GeneralProfile = profile(generalLimit, generalBurst, generalCap) //nolint:gochecknoglobals // public preset value

// AnalysisProfile fits interactive analysis endpoints: 10 req/min, burst 15.
//
//nolint:gochecknoglobals // public preset value
var AnalysisProfile = profile(analysisLimit, analysisBurst, analysisCap)

// ExportProfile fits heavy export endpoints (PDF/report generation):
// 5 req/min, burst 8.
var ExportProfile = profile(exportLimit, exportBurst, exportCap) //nolint:gochecknoglobals // public preset value

// ContactProfile fits form submissions: 8 req/min, burst 10.
var ContactProfile = profile(contactLimit, contactBurst, contactCap) //nolint:gochecknoglobals // public preset value

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
	host, _, splitErr := net.SplitHostPort(r.RemoteAddr)
	if splitErr == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}
