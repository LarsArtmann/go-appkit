# go-appkit/health

Production-grade health checks and a real-time status dashboard for
[go-appkit] services, in one wiring call. Bridges [go-health] — the
three-probe Kubernetes health-check SDK (liveness, readiness, startup with
critical/non-critical classification, background caching, shutdown
awareness) — and [go-health-dashboard] — the real-time HTML dashboard with
SSE updates, trend history, Prometheus exposition, and webhooks — into
appkit's lifecycle.

> Requires `GOEXPERIMENT=jsonv2` (go-health serializes with
> `encoding/json/v2`; the dashboard depends on go-sse).

## Why

appkit's default `/health`, `/health/live`, `/health/ready` answer *is the
process up* with a two-state JSON. This module upgrades the health surface
to answer *is the app actually usable*:

| Concern                              | Default endpoints | With this module                          |
| ------------------------------------ | ----------------- | ----------------------------------------- |
| Critical vs non-critical services    | —                 | fail (503) vs warn (200, degraded)        |
| Per-check results                    | —                 | named checks with errors in the response  |
| Background caching (kubelet storms)  | —                 | 1s refresh, lock-free reads               |
| Startup latch (slow boots)           | —                 | `/startupz`, one-way latch                |
| Real-time dashboard                  | —                 | `/health` HTML + SSE, trend, dark mode    |
| Prometheus exposition                | —                 | `/health/metrics` (opt-in)                |
| Webhooks on transitions              | —                 | change-only JSON pushes (opt-in)          |
| Drain coherence                      | appkit probe only | probe `/readyz` flips 503 in lockstep     |

## Quick start

```go
no := false

checks := map[string]appkithealth.CheckFunc{
	"database": func(ctx context.Context) error { return db.PingContext(ctx) },
	"cache":    func(ctx context.Context) error { return cache.PingContext(ctx) },
}

probe := appkithealth.NewProbe(checks, health.WithCriticalServices("database"))

mounted, err := appkithealth.New(probe, appkithealth.WithDashboard(
	dashboard.WithTrend(300),
	dashboard.WithMetrics(true),
))

cfg := appkit.DefaultServiceConfig()
cfg.RegisterHealth = &no // this module serves the richer health surface
cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
	mounted.Drain() // /readyz → 503 at the start of the drain window
	return nil
})
cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)

svc, err := appkit.NewService(cfg)
mounted.RegisterRoutes(svc.Mux)

_ = mounted.Start(ctx) // probe cache + dashboard pusher, initial batch runs synchronously
_ = svc.Run(ctx)       // blocks until SIGINT/SIGTERM
```

Routes served: `/healthz`, `/readyz`, `/startupz` (JSON probes), `/health`
(dashboard HTML; JSON with `Accept: application/json`), `/health/sse`,
`/health/metrics`, `/favicon.svg`. Without `WithDashboard`, only the probe
routes are registered and appkit's default endpoints can stay enabled.

A runnable demo (critical check + flapping non-critical check) lives in
[`example/`](example/): `GOWORK=off GOEXPERIMENT=jsonv2 go run ./example`.

## Drain ordering, in one picture

```
SIGTERM ─▶ readyProbe=false ─▶ DrainHooks ─▶ DrainDelay ─▶ connections close ─▶ ShutdownHooks
                                    │
                                    └─ mounted.Drain(): probe shutting-down → /readyz 503
                                       (every readiness surface down for the WHOLE drain window)
```

## API

| Symbol                     | Purpose                                                                          |
| -------------------------- | -------------------------------------------------------------------------------- |
| `NewProbe(checks, opts...)` | Injector-free go-health probe: per-check `CheckFunc` map, concurrent batches, per-check panic isolation. SDK options pass through (`WithCriticalServices`, `WithTimeout`, `WithRefreshInterval`, …). |
| `New(probe, opts...)`      | Lifecycle handle without routes — the primary appkit flow (config references the handle before the mux exists). |
| `Mounted.RegisterRoutes(mux)` | Registers probe routes (and dashboard routes with `WithDashboard`).           |
| `Mount(mux, probe, opts...)` | `New` + `RegisterRoutes` for muxes that already exist.                         |
| `Mounted.Start(ctx)`       | Initial synchronous batch + background refresh + dashboard pusher.               |
| `Mounted.Drain()`          | Probe → shutting-down: every readiness surface 503, liveness stays 200.          |
| `Mounted.Shutdown(ctx)`    | Drain + stop pusher/refresh loop. Idempotent.                                    |
| `Mounted.Ready()`          | Cached readiness verdict — pass as `ServiceConfig.ReadyCheck` if you keep appkit's endpoints. |
| `Mounted.Probe()` / `.Dashboard()` | Escape hatches for advanced wiring (webhooks, middleware, accessors).    |

## Gotchas

- **Set `RegisterHealth: &false` with the dashboard** — the dashboard serves
  `/health` by default; the mux panics on the duplicate registration if you
  forget. Without the dashboard, appkit's defaults and this module's probe
  routes coexist peacefully.
- **No `GET /` catch-all with the dashboard** — the dashboard registers
  method-agnostic routes (`/health`), and Go's ServeMux precedence rules
  panic on `GET /` vs `/health`. Register your root handler without a method.
- **Checks must honor their context** — they receive the batch's
  timeout-bounded context (5s default); a check that ignores cancellation
  stalls its whole batch until the deadline.
- **Eagerly exercise lazily-initialized dependencies** — a check that only
  lazily connects on first use reports pass on a cold process. Ping for real.
- **One dashboard per process is plenty** — one probe can carry all checks;
  go-health's `aggregate` merges multi-probe setups if you outgrow that.
- **Import aliasing** — this package and the SDK are both named `health`.
  Alias this module (`appkithealth "github.com/larsartmann/go-appkit/health"`)
  when you also import `github.com/larsartmann/go-health`.

## Integration with flightrecorderhealth

`flightrecorderhealth.Trigger` auto-captures flight-recorder traces on
health-check failures. It plugs into go-health's *recorder path*, which
requires a samber/do injector — build the probe with the SDK's
`health.New(injector, health.WithHealthRecorder(trigger))` instead of
`NewProbe` (whose explicit batch function bypasses the recorder, by design).
`flightrecorderhealth.NewCheckable(rec)` also registers the recorder's own
state as a check. See that module's README for the full wiring.

[go-appkit]: https://github.com/larsartmann/go-appkit
[go-health]: https://github.com/larsartmann/go-health
[go-health-dashboard]: https://github.com/larsartmann/go-health-dashboard
