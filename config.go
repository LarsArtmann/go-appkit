package appkit

import (
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"
)

const (
	defaultAddr              = ":8080"
	defaultReadTimeout       = 10 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 15 * time.Second
	defaultDrainDelay        = 5 * time.Second
)

// NoTimeout disables a server deadline instead of defaulting it. Assign it
// to ReadTimeout or WriteTimeout when responses outlive any fixed deadline —
// long-lived SSE streams are the canonical case: a WriteTimeout would cut
// them mid-stream. With WriteTimeout set to NoTimeout the default stack also
// drops its per-request Timeout middleware, and http.Server runs without the
// corresponding deadline. ReadHeaderTimeout and IdleTimeout (the slowloris /
// keep-alive reaping pair) stay enabled either way.
const NoTimeout time.Duration = -1

// ServiceConfig holds all configuration for a Service.
// Zero-value fields are replaced with sensible defaults by NewService.
type ServiceConfig struct {
	Addr              string
	LogLevel          LogLevel
	LogFormat         LogFormat
	ReadTimeout       time.Duration // 0 = default; NoTimeout = disabled
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration // 0 = default; NoTimeout = disabled (SSE)
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	DrainDelay        time.Duration

	// Middlewares replaces the default middleware stack if non-nil.
	// Default: Recovery → RequestID → Logging → Timeout → SecurityHeaders.
	Middlewares []httputil.Middleware

	// ExtraMiddlewares appends to the default stack (ignored if Middlewares is non-nil).
	ExtraMiddlewares []httputil.Middleware

	// RegisterHealth controls whether /health, /health/live, /health/ready are auto-registered.
	// Uses a pointer so the zero-value (nil) defaults to true. Set to false to opt out.
	RegisterHealth *bool

	// ReadyCheck, when set, is consulted by /health/ready in addition to the
	// internal drain probe: the endpoint reports 200 only while BOTH the
	// drain probe is up AND ReadyCheck returns true. Use it to gate traffic
	// on external state — e.g. cqrs.EventService.ReadyCheck (projections
	// caught up) or a dependency ping. It must be safe for concurrent use.
	// Optional — nil keeps probe-only readiness.
	ReadyCheck func() bool
}

// DefaultServiceConfig returns a config with production-ready defaults.
func DefaultServiceConfig() ServiceConfig {
	t := true

	return ServiceConfig{
		Addr:              defaultAddr,
		LogLevel:          LogLevelInfo,
		LogFormat:         LogFormatAuto,
		ReadTimeout:       defaultReadTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ShutdownTimeout:   defaultShutdownTimeout,
		DrainDelay:        defaultDrainDelay,
		RegisterHealth:    &t,
	}
}

func (cfg *ServiceConfig) applyDefaults() {
	def := DefaultServiceConfig()

	if cfg.Addr == "" {
		cfg.Addr = def.Addr
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = def.LogLevel
	}

	if cfg.LogFormat == "" {
		cfg.LogFormat = def.LogFormat
	}

	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = def.ReadTimeout
	}

	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = def.ReadHeaderTimeout
	}

	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = def.WriteTimeout
	}

	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = def.IdleTimeout
	}

	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = def.ShutdownTimeout
	}

	if cfg.DrainDelay == 0 {
		cfg.DrainDelay = def.DrainDelay
	}
}

// Validate checks the config for invalid values after defaults are applied.
func (cfg *ServiceConfig) Validate() error {
	if invalidTimeout(cfg.ReadTimeout) {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.read_timeout_negative",
			"ReadTimeout must not be negative (except NoTimeout): %v",
			cfg.ReadTimeout,
		)
	}

	if cfg.ReadHeaderTimeout < 0 {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.read_header_timeout_negative",
			"ReadHeaderTimeout must not be negative: %v",
			cfg.ReadHeaderTimeout,
		)
	}

	if invalidTimeout(cfg.WriteTimeout) {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.write_timeout_negative",
			"WriteTimeout must not be negative (except NoTimeout): %v",
			cfg.WriteTimeout,
		)
	}

	if cfg.IdleTimeout < 0 {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.idle_timeout_negative",
			"IdleTimeout must not be negative: %v",
			cfg.IdleTimeout,
		)
	}

	if cfg.ShutdownTimeout <= 0 {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.shutdown_timeout_invalid",
			"ShutdownTimeout must be positive: %v",
			cfg.ShutdownTimeout,
		)
	}

	if cfg.DrainDelay < 0 {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.drain_delay_negative",
			"DrainDelay must not be negative: %v",
			cfg.DrainDelay,
		)
	}

	return nil
}

// invalidTimeout reports whether d is a negative duration other than the
// NoTimeout sentinel. Zero means "apply the default" and is always valid.
func invalidTimeout(d time.Duration) bool {
	return d < 0 && d != NoTimeout
}

// serverTimeout translates a config timeout for http.Server: the NoTimeout
// sentinel becomes 0, which the stdlib treats as "no deadline".
func serverTimeout(d time.Duration) time.Duration {
	if d == NoTimeout {
		return 0
	}

	return d
}
