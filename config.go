package appkit

import (
	"context"
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

// NoDrainDelay skips the drain wait in Shutdown. Zero is NOT "no delay" — it
// applies the default — because the zero-value config must stay
// production-safe for load-balancer propagation. Assign NoDrainDelay when
// shutdown speed matters more than in-flight load-balancer convergence:
// tests are the canonical case. The ready probe still flips to false
// immediately either way; only the wait between probe flip and listener
// close is skipped.
const NoDrainDelay time.Duration = -2

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

	// OuterMiddlewares wraps the entire chain — including the default stack
	// or a configured Middlewares replacement — and therefore runs outermost,
	// before Recovery. Use it for instrumentation that must observe the full
	// request lifetime and seed request-scoped context for everything
	// downstream; OpenTelemetry tracing is the canonical case (see the otel
	// module). Optional.
	OuterMiddlewares []httputil.Middleware

	// DrainHooks run once, in order, at the start of the shutdown drain —
	// after the ready probe flips to false but BEFORE the DrainDelay wait
	// and the listener close — each receiving the shutdown context. Use them
	// to flip external readiness signals in lockstep with the service's own
	// probe, so every readiness endpoint (the framework's and any mounted
	// probe's, e.g. the health module's /readyz) reports down for the whole
	// drain window and load balancers stop routing immediately. A failing
	// hook does not stop the drain or the remaining hooks; errors are joined
	// with the shutdown result and classified as infrastructure failures. A
	// service that never started does not run its hooks. Optional.
	DrainHooks []func(context.Context) error

	// ShutdownHooks run once, in order, after the server has shut down and
	// released its connections, each receiving the shutdown context. Use them
	// to flush telemetry providers — their spans then cover the final
	// in-flight requests — and to close downstream resources. A failing hook
	// does not stop the rest; errors are joined and classified as
	// infrastructure failures. A service that never started does not run its
	// hooks — defer provider.Shutdown yourself on startup-error paths.
	// Optional.
	ShutdownHooks []func(context.Context) error

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

	return ServiceConfig{ //nolint:exhaustruct // optional fields stay zero
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

	if cfg.DrainDelay < 0 && cfg.DrainDelay != NoDrainDelay {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.drain_delay_negative",
			"DrainDelay must not be negative (except NoDrainDelay): %v",
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
