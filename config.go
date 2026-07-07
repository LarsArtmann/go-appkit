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

// ServiceConfig holds all configuration for a Service.
// Zero-value fields are replaced with sensible defaults by NewService.
type ServiceConfig struct {
	Addr              string
	LogLevel          LogLevel
	LogFormat         LogFormat
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
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
	if cfg.ReadTimeout < 0 {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.read_timeout_negative",
			"ReadTimeout must not be negative: %v",
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

	if cfg.WriteTimeout < 0 {
		return errorfamily.Newf(
			errorfamily.Rejection,
			"config.write_timeout_negative",
			"WriteTimeout must not be negative: %v",
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
