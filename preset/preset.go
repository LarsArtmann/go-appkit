package preset

import (
	"context"
	"log/slog"
	"slices"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	"github.com/larsartmann/go-appkit/health"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/systemd"
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
	var provider *appkitotel.Provider

	if legs.otel {
		opts := cfg.OTEL.setupOptions(cfg.Version)

		provider, err = appkitotel.Setup(opts...)
		if err != nil {
			return nil, err
		}
	}

	var (
		mounted *health.Mounted
		hub     *realtime.Hub
	)

	sc := appkit.ServiceConfig{
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
			if mounted == nil {
				return nil
			}

			return mounted.Start(ctx)
		}),
		DrainHooks: append(slices.Clone(cfg.DrainHooks), func(context.Context) error {
			if mounted == nil {
				return nil
			}

			mounted.Drain()

			return nil
		}),
		ShutdownHooks: append(slices.Clone(cfg.ShutdownHooks), func(ctx context.Context) error {
			if hub != nil {
				if err := hub.Shutdown(ctx); err != nil {
					return err
				}
			}

			if provider != nil {
				return provider.Shutdown(ctx)
			}

			return nil
		}),
	}

	// The dashboard owns /health (the health module's documented contract):
	// core's default /health family would collide and panic the mux. Without
	// the dashboard, core's family registers normally.
	registerHealth := cfg.NoDashboard
	sc.RegisterHealth = &registerHealth

	if provider != nil {
		sc.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
	}

	if legs.hasSecurity {
		middlewares, secErr := securityMiddlewares(cfg.Security)
		if secErr != nil {
			flushProvider(provider)

			return nil, secErr
		}
		// ExtraMiddlewares run inside the default stack; httputil's header
		// writes use Set semantics, so the profile's environment-tuned
		// values are authoritative over the default stack's basic set.
		sc.ExtraMiddlewares = middlewares
	}

	systemd.Install(&sc)
	appkit.RegisterClassifications()

	svc, err := appkit.NewService(sc)
	if err != nil {
		flushProvider(provider)

		return nil, err
	}

	probe := health.NewProbe(cfg.HealthChecks)

	mountOpts := make([]health.MountOption, 0, 1)
	if !cfg.NoDashboard {
		mountOpts = append(mountOpts, health.WithDashboard())
	}

	mounted, err = health.Mount(svc.Mux, probe, mountOpts...)
	if err != nil {
		flushProvider(provider)

		return nil, err
	}

	if legs.errorPages {
		var pages errorpages.Config
		if cfg.ErrorPages != nil {
			pages = cfg.ErrorPages.Config
		}

		errorpages.Mount(svc.Mux, pages)
	}

	if legs.realtime {
		hub = realtime.NewHub()

		pattern := defaultRealtimePattern
		if cfg.Realtime != nil && cfg.Realtime.Pattern != "" {
			pattern = cfg.Realtime.Pattern
		}

		var mountOpts []realtime.MountOption
		if cfg.Realtime != nil {
			mountOpts = cfg.Realtime.Options
		}

		realtime.Mount(svc.Mux, pattern, hub, mountOpts...)

		if cfg.Realtime != nil && cfg.Realtime.Hub != nil {
			*cfg.Realtime.Hub = hub
		}
	}

	return svc, nil
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
