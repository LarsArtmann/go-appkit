package health

import (
	"context"
	"net/http"
	"sync"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
)

// MountOption configures [Mount].
type MountOption func(*mountConfig)

type mountConfig struct {
	dashboardEnabled bool
	dashboardOptions []dashboard.Option
	probeRoutes      health.Routes
}

// WithDashboard opts into the real-time HTML dashboard (SSE updates, trend
// history, Prometheus exposition, webhooks — everything the dashboard's own
// Option values control). The dashboard registers the probe endpoints from
// its route configuration, so WithBasePath and WithRoutes apply uniformly to
// everything Mount serves; WithProbeRoutes is ignored in this mode.
//
// The dashboard serves /health by default, which collides with appkit's
// default health endpoint — set ServiceConfig.RegisterHealth to &false when
// enabling it.
func WithDashboard(opts ...dashboard.Option) MountOption {
	return func(c *mountConfig) {
		c.dashboardEnabled = true
		c.dashboardOptions = opts
	}
}

// WithProbeRoutes overrides the Kubernetes probe paths. Only used when the
// dashboard is disabled (the default): the dashboard registers probe
// endpoints from its own route configuration.
func WithProbeRoutes(routes health.Routes) MountOption {
	return func(c *mountConfig) { c.probeRoutes = routes }
}

// New creates the health surface lifecycle handle without registering any
// routes. Use it when the config must reference the handle before the
// service's mux exists — the primary appkit flow:
//
//	mounted, err := appkithealth.New(probe, appkithealth.WithDashboard())
//	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
//		mounted.Drain()
//		return nil
//	})
//	cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)
//	svc, err := appkit.NewService(cfg)
//	mounted.RegisterRoutes(svc.Mux)
//
// Register the routes with [Mounted.RegisterRoutes] before serving. [Mount]
// is the one-shot convenience for muxes that already exist.
func New(probe *health.Probe, opts ...MountOption) (*Mounted, error) {
	if probe == nil {
		return nil, errorfamily.Newf(
			errorfamily.Rejection,
			"health.probe_missing",
			"probe must not be nil",
		)
	}

	cfg := mountConfig{ //nolint:exhaustruct // dashboard fields default to disabled
		probeRoutes: health.DefaultRoutes(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	m := &Mounted{ //nolint:exhaustruct // dashboard, mu, started are deliberate zero values
		probe:       probe,
		probeRoutes: cfg.probeRoutes,
	}

	if cfg.dashboardEnabled {
		m.dashboard = dashboard.New(probe, cfg.dashboardOptions...)
	}

	return m, nil
}

// Mount registers the health surface on the mux and returns lifecycle
// handles for appkit's DrainHooks and ShutdownHooks. It is [New] followed by
// [Mounted.RegisterRoutes] — use it when the mux already exists.
func Mount(mux *http.ServeMux, probe *health.Probe, opts ...MountOption) (*Mounted, error) {
	if mux == nil {
		return nil, errorfamily.Newf(
			errorfamily.Rejection,
			"health.mount_mux_missing",
			"mux must not be nil",
		)
	}

	m, err := New(probe, opts...)
	if err != nil {
		return nil, err
	}

	m.RegisterRoutes(mux)

	return m, nil
}

// RegisterRoutes registers the health surface's routes on the mux: the probe
// endpoints (/healthz, /readyz, /startupz by default), plus the dashboard's
// routes when mounted with [WithDashboard] — the dashboard then owns the
// probe endpoints too, so its WithBasePath / WithRoutes options apply
// uniformly. Register exactly once; the mux panics on duplicate patterns.
func (m *Mounted) RegisterRoutes(mux *http.ServeMux) {
	if m.dashboard != nil {
		m.dashboard.RegisterRoutes(mux)

		return
	}

	m.probe.RegisterRoutes(mux, m.probeRoutes)
}

// Mounted is the lifecycle handle for a health surface created by [New] and
// registered with [Mounted.RegisterRoutes]. Wire [Mounted.Drain] into
// ServiceConfig.DrainHooks and [Mounted.Shutdown] into
// ServiceConfig.ShutdownHooks; call [Mounted.Start] before serving.
type Mounted struct {
	probe     *health.Probe
	dashboard *dashboard.Dashboard

	probeRoutes health.Routes

	mu      sync.Mutex
	started bool
}

// Start begins the probe's background evaluation — an initial synchronous
// batch runs first, so handlers and the dashboard see fresh data
// immediately — and, when mounted with [WithDashboard], the SSE pusher
// goroutine. Call it once before serving traffic; call it again only after
// [Mounted.Shutdown].
//
// An invalid probe configuration surfaces here as the SDK's validation
// error (match with errors.Is against health.ErrInvalidTimeout /
// health.ErrInvalidRefreshInterval).
func (m *Mounted) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()

		return errorfamily.Newf(
			errorfamily.Rejection,
			"health.already_started",
			"health surface already started",
		)
	}

	m.started = true
	m.mu.Unlock()

	err := m.probe.Start(ctx)
	if err != nil {
		m.mu.Lock()
		m.started = false
		m.mu.Unlock()

		return err
	}

	if m.dashboard != nil {
		err := m.dashboard.Start(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Drain marks the probe as shutting down: every readiness surface it serves
// reports 503 immediately — even from a stale cache — while liveness stays
// 200. Wire it into ServiceConfig.DrainHooks so external readiness flips at
// the start of appkit's drain window, in lockstep with the framework's own
// ready probe:
//
//	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
//		mounted.Drain()
//		return nil
//	})
//
// Safe to call multiple times; [Mounted.Shutdown] drains too.
func (m *Mounted) Drain() {
	m.probe.Shutdown()
}

// Shutdown drains the probe, stops the SSE pusher (draining connected
// dashboard clients when the dashboard's WithShutdownDrain option is set),
// and stops the background refresh loop. Wire it into
// ServiceConfig.ShutdownHooks:
//
//	cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)
//
// Safe to call multiple times. [Mounted.Start] may be called again
// afterwards.
func (m *Mounted) Shutdown(_ context.Context) error {
	m.Drain()

	m.mu.Lock()
	m.started = false
	m.mu.Unlock()

	if m.dashboard != nil {
		m.dashboard.Shutdown()
	}

	return nil
}

// Ready reports whether the probe's cached view allows traffic: not
// shutting down and roll-up status not fail. Pass it as
// ServiceConfig.ReadyCheck to gate appkit's /health/ready on probe health in
// addition to the drain probe.
func (m *Mounted) Ready() bool {
	return m.probe.Ready()
}

// Probe returns the underlying go-health probe for advanced wiring —
// programmatic accessors (Status, Alive, AwaitReady), the injector-free
// HealthCheck for samber/do registration, or WithEvaluationHook metrics.
func (m *Mounted) Probe() *health.Probe {
	return m.probe
}

// Dashboard returns the dashboard, or nil when mounted without
// [WithDashboard]. Use it for post-construction wiring that has no Mount
// passthrough — webhooks, middleware protection, CSP nonce extraction.
func (m *Mounted) Dashboard() *dashboard.Dashboard {
	return m.dashboard
}
