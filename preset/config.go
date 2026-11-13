package preset

import (
	"time"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	"github.com/larsartmann/go-appkit/health"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/go-appkit/realtime"
)

// TimeoutsConfig passes the core timeout knobs through to the ServiceConfig
// verbatim; zero values keep appkit's defaults (and the NoTimeout /
// NoDrainDelay sentinels flow through unchanged).
type TimeoutsConfig struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	DrainDelay        time.Duration
}

// Config assembles a service from the always-on spine plus opt-in legs.
//
// Every leg is a pointer: nil means the corresponding APPKIT_* env switch
// decides (see the package doc for the frozen grammar); non-nil means the leg
// decides itself via its Enabled field and the env var is ignored.
type Config struct {
	// Addr is the listen address (appkit.ServiceConfig.Addr).
	Addr string

	// LogLevel / LogFormat pass through to the core logger.
	LogLevel  appkit.LogLevel
	LogFormat appkit.LogFormat

	// Version stamps GET /version and the build-info metric label.
	Version string

	// Metrics passes through verbatim — the preset never invents
	// credentials. Nil disables the metrics surface.
	Metrics *appkit.MetricsConfig

	// ReadyCheck passes through to /health/ready.
	ReadyCheck func() bool

	// Timeouts passes through to the core ServiceConfig.
	Timeouts TimeoutsConfig

	// HealthChecks are the probe's named checks. The nil map boots the
	// probe with no checks (endpoints still serve; checks can be added
	// later by the health module's own API).
	HealthChecks map[string]health.CheckFunc

	// NoDashboard omits the health module's real-time dashboard
	// (healthz/readyz/startupz still mount).
	NoDashboard bool

	// Lifecycle hooks pass through to the core ServiceConfig. Consumer
	// hooks run FIRST in every phase; the preset's spine closures (health
	// Mounted, realtime hub, otel provider flush) run after, so otel spans
	// cover app shutdown work.
	StartHooks    []appkit.Hook
	DrainHooks    []appkit.Hook
	ShutdownHooks []appkit.Hook

	// Legs — nil means the env switch decides.
	OTEL       *OTELConfig
	Security   *SecurityConfig
	Realtime   *RealtimeConfig
	ErrorPages *ErrorPagesConfig
}

// OTELConfig controls the OpenTelemetry leg. Enabled=true boots otel.Setup
// with WithService (when ServiceName/ServiceVersion are set) and mounts
// otel.Middleware outermost; the provider's Shutdown joins ShutdownHooks so
// spans cover the final in-flight requests.
type OTELConfig struct {
	Enabled bool

	// ServiceName / ServiceVersion / InstanceID map to otel.WithService.
	// Version falls back to Config.Version when ServiceVersion is empty.
	ServiceName    string
	ServiceVersion string
	InstanceID     string

	// Options append to otel.Setup (after WithService).
	Options []appkitotel.SetupOption
}

// SecurityProfile selects the security leg's shape.
type SecurityProfile string

const (
	// SecurityNone adds nothing — the absent-env default. Security never
	// silently joins a stack.
	SecurityNone SecurityProfile = "none"

	// SecurityStandard adds SecurityHeaders (production posture) plus a
	// request body limit (1 MiB unless BodyLimitBytes overrides).
	SecurityStandard SecurityProfile = "standard"

	// SecurityStrict adds OriginCheck on top of standard. Origins is
	// REQUIRED — New rejects without it (a strict profile with no origins
	// would reject every browser request; fail fast instead).
	SecurityStrict SecurityProfile = "strict"
)

// SecurityConfig controls the security leg.
type SecurityConfig struct {
	Enabled bool

	// Profile selects the composed batteries. Empty defaults to Standard
	// when Enabled (explicit code, explicit shape).
	Profile SecurityProfile

	// Origins is the OriginCheck allowlist — required for Strict.
	Origins []string

	// BodyLimitBytes overrides the standard-profile body limit
	// (default 1 MiB; <=0 keeps the default).
	BodyLimitBytes int64
}

// RealtimeConfig controls the SSE leg: a hub mounted at Pattern (default
// "GET /events"); the hub's graceful Shutdown joins DrainHooks (streams must
// terminate BEFORE server.Shutdown waits on connections — Hub.Shutdown's own
// contract).
type RealtimeConfig struct {
	Enabled bool

	// Pattern is the mount pattern (stdlib ServeMux syntax, method
	// included). Default "GET /events".
	Pattern string

	// Options pass to realtime.Mount (e.g. WithHeartbeat).
	Options []realtime.MountOption

	// Hub receives the constructed hub when New returns — register
	// broadcasts on it after boot. Optional.
	Hub **realtime.Hub
}

// ErrorPagesConfig controls the pretty-error-pages leg. The zero Config
// mounts errorpages with its defaults.
type ErrorPagesConfig struct {
	Enabled bool

	Config errorpages.Config
}
