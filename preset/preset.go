package preset

import (
	"context"
	"log/slog"
	"net/http"
	"slices"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	"github.com/larsartmann/go-appkit/health"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/systemd"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"
)

// defaultBodyLimitBytes is the standard-profile body limit (1 MiB).
const defaultBodyLimitBytes int64 = 1 << 20

// defaultRealtimePattern is the SSE mount pattern when RealtimeConfig leaves
// it empty.
const defaultRealtimePattern = "GET /events"

// New assembles the service: the always-on spine (core Service, sd_notify,
// health probe + dashboard, error classification, version stamp) plus every
// leg the grammar resolves to on. The consumer registers routes on
// svc.Mux (specific patterns win over the error-pages catch-all per stdlib
// ServeMux precedence) and calls svc.Run.
//
// Fail-fast contract: invalid env values reject BEFORE any listener binds.
func New(cfg Config) (*appkit.Service, error) {
	legs, err := resolveLegs(cfg)
	if err != nil {
		return nil, err
	}

	// The otel provider exists before the service does; its Shutdown joins
	// ShutdownHooks via a closure, and startup-error paths below flush it
	// (the core never runs hooks for a service that never started).
	provider, setupErr := setupProvider(cfg, legs.otel)
	if setupErr != nil {
		return nil, setupErr
	}

	var (
		mounted *health.Mounted
		hub     *realtime.Hub
	)

	svcConfig := appkit.ServiceConfig{ //nolint:exhaustruct_v5 // maps only the preset-owned fields; core defaults carry the rest
		Addr:              cfg.Addr,
		LogLevel:          cfg.LogLevel,
		LogFormat:         cfg.LogFormat,
		Version:           cfg.Version,
		Metrics:           cfg.Metrics,
		ReadyCheck:        cfg.ReadyCheck,
		ReadTimeout:       cfg.Timeouts.ReadTimeout,
		ReadHeaderTimeout: cfg.Timeouts.ReadHeaderTimeout,
		WriteTimeout:      cfg.Timeouts.WriteTimeout,
		IdleTimeout:       cfg.Timeouts.IdleTimeout,
		ShutdownTimeout:   cfg.Timeouts.ShutdownTimeout,
		DrainDelay:        cfg.Timeouts.DrainDelay,

		// Consumer hooks run FIRST in every phase; the spine closures below
		// run after, so otel spans (flushed last) cover app shutdown work.
		//
		// Nil-guard closures: the pointed-to values are assigned after
		// NewService (below) — hooks only run at Start/Run or Shutdown,
		// long after assignment, so the closures are safe.
		StartHooks: append(slices.Clone(cfg.StartHooks), func(ctx context.Context) error {
			return startMounted(ctx, mounted)
		}),
		DrainHooks: append(slices.Clone(cfg.DrainHooks), func(context.Context) error {
			drainMounted(mounted)

			return nil
		}),
		ShutdownHooks: append(slices.Clone(cfg.ShutdownHooks), func(ctx context.Context) error {
			return shutdownLegs(ctx, hub, provider)
		}),
	}

	// The dashboard owns /health (the health module's documented contract):
	// core's default /health family would collide and panic the mux. Without
	// the dashboard, core's family registers normally.
	registerHealth := cfg.NoDashboard
	svcConfig.RegisterHealth = &registerHealth

	if provider != nil {
		svcConfig.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
	}

	if legs.hasSecurity {
		if err := applySecurity(&svcConfig, cfg.Security); err != nil {
			return nil, bootFail(provider, err)
		}
	}

	systemd.Install(&svcConfig)
	appkit.RegisterClassifications()

	svc, err := appkit.NewService(svcConfig)
	if err != nil {
		return nil, bootFail(provider, err)
	}

	probe := health.NewProbe(cfg.HealthChecks)

	mounted, err = health.Mount(svc.Mux, probe, dashboardOpts(cfg.NoDashboard)...)
	if err != nil {
		return nil, bootFail(provider, err)
	}

	if legs.errorPages {
		mountErrorPages(svc.Mux, cfg)
	}

	if legs.realtime {
		hub = mountRealtime(svc.Mux, &cfg)
	}

	return svc, nil
}

// startMounted starts the health dashboard when mounted (nil until New
// finishes wiring — the hook only runs at Start/Run, long after assignment).
func startMounted(ctx context.Context, mounted *health.Mounted) error {
	if mounted == nil {
		return nil
	}

	return mounted.Start(ctx) //nolint:wrapcheck // health module classifies its own start errors
}

// drainMounted flips the dashboard to draining when mounted.
func drainMounted(mounted *health.Mounted) {
	if mounted == nil {
		return
	}

	mounted.Drain()
}

// shutdownLegs tears the legs down in reverse dependency order: the realtime
// hub first (stops streams), then the otel provider (flushes spans last).
func shutdownLegs(ctx context.Context, hub *realtime.Hub, provider *appkitotel.Provider) error {
	if hub != nil {
		if err := hub.Shutdown(ctx); err != nil {
			return errorfamily.WrapInfrastructuref(err, "preset.realtime_shutdown_failed", "shutdown realtime hub")
		}
	}

	if provider != nil {
		return provider.Shutdown(ctx) //nolint:wrapcheck // flush errors propagate to core's classified join
	}

	return nil
}

// setupProvider builds and starts the otel provider when the leg resolved
// on; a nil provider (leg off) makes every downstream no-op.
func setupProvider(cfg Config, enabled bool) (*appkitotel.Provider, error) {
	if !enabled {
		return nil, nil //nolint:nilnil // nil provider IS the leg-off signal — every consumer nil-checks
	}

	provider, err := appkitotel.Setup(cfg.OTEL.setupOptions(cfg.Version)...)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "preset.otel_setup_failed", "setup otel provider")
	}

	return provider, nil
}

// mountRealtime mounts the SSE hub at the configured pattern (default
// /events) and hands the hub to the config's out-field when set.
func mountRealtime(mux *http.ServeMux, cfg *Config) *realtime.Hub {
	hub := realtime.NewHub()

	pattern := defaultRealtimePattern
	if cfg.Realtime != nil && cfg.Realtime.Pattern != "" {
		pattern = cfg.Realtime.Pattern
	}

	var mountOpts []realtime.MountOption
	if cfg.Realtime != nil {
		mountOpts = cfg.Realtime.Options
	}

	realtime.Mount(mux, pattern, hub, mountOpts...)

	if cfg.Realtime != nil && cfg.Realtime.Hub != nil {
		*cfg.Realtime.Hub = hub
	}

	return hub
}

// applySecurity resolves the profile's middlewares onto the service config.
// ExtraMiddlewares run inside the default stack; httputil's header writes
// use Set semantics, so the profile's environment-tuned values are
// authoritative over the default stack's basic set.
func applySecurity(svcConfig *appkit.ServiceConfig, secCfg *SecurityConfig) error {
	middlewares, err := securityMiddlewares(secCfg)
	if err != nil {
		return err
	}

	svcConfig.ExtraMiddlewares = middlewares

	return nil
}

// dashboardOpts mounts the real-time dashboard unless opted out.
func dashboardOpts(noDashboard bool) []health.MountOption {
	if noDashboard {
		return nil
	}

	return []health.MountOption{health.WithDashboard()}
}

// mountErrorPages mounts the pretty-error catch-all with the consumer's
// embedded config when set.
func mountErrorPages(mux *http.ServeMux, cfg Config) {
	var pages errorpages.Config
	if cfg.ErrorPages != nil {
		pages = cfg.ErrorPages.Config
	}

	errorpages.Mount(mux, pages)
}

// bootFail is the shared startup-error exit: flush the partially-initialized
// otel provider (the core never runs ShutdownHooks for a service that never
// started), then propagate the cause untouched.
func bootFail(provider *appkitotel.Provider, err error) error {
	flushProvider(provider)

	return err
}

// flushProvider shuts the otel provider down on startup-error paths — the
// core never runs ShutdownHooks for a service that never started.
func flushProvider(provider *appkitotel.Provider) {
	if provider == nil {
		return
	}

	if err := provider.Shutdown(context.Background()); err != nil {
		slog.Error("preset: otel provider flush failed", "error", err)
	}
}

// setupOptions builds the otel Setup options, prefixing WithService when the
// leg names a service (version falls back to the spine's Version).
func (c *OTELConfig) setupOptions(version string) []appkitotel.SetupOption {
	if c == nil {
		return nil
	}

	opts := make([]appkitotel.SetupOption, 0, 1+len(c.Options))

	svcVersion := c.ServiceVersion
	if svcVersion == "" {
		svcVersion = version
	}

	if c.ServiceName != "" || svcVersion != "" {
		opts = append(opts, appkitotel.WithService(c.ServiceName, svcVersion, c.InstanceID))
	}

	return append(opts, c.Options...)
}
