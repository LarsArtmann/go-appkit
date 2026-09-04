# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `NewProbe(checks, opts...)` — an injector-free go-health probe from a map
  of named `CheckFunc`s. Checks run concurrently per batch, are
  panic-isolated per check (a panicking check fails as that check's error,
  classified by criticality), and receive the batch's timeout-bounded
  context. All go-health `Option`s pass through, including
  `WithCriticalServices` and `WithTimeout`.
- `New(probe, opts...)` — creates the health surface lifecycle handle
  (`*Mounted`) without routes; the primary appkit flow lets `ServiceConfig`
  reference the handle before the service's mux exists.
- `Mounted.RegisterRoutes(mux)` — registers the kubelet probe routes
  (`/healthz`, `/readyz`, `/startupz` by default, overridable via
  `WithProbeRoutes`), plus all dashboard routes when the dashboard is
  enabled.
- `Mount(mux, probe, opts...)` — `New` + `RegisterRoutes` convenience for
  muxes that already exist.
- `WithDashboard(opts...)` — opt-in real-time HTML dashboard
  (go-health-dashboard v0.5.0): SSE updates, trend history, Prometheus
  exposition, webhooks, content negotiation on `/health`. The dashboard
  then also registers the probe endpoints, so `WithBasePath` applies
  uniformly.
- `Mounted.Start(ctx)` — initial synchronous health batch (fresh data for
  the dashboard's first patch), background cache refresh, dashboard SSE
  pusher. Rejected when already started; legal again after `Shutdown`.
- `Mounted.Drain()` — marks the probe shutting down: every readiness
  surface reports 503 immediately, liveness stays 200. Wire into
  `ServiceConfig.DrainHooks` for lockstep drain with the framework's own
  ready probe.
- `Mounted.Shutdown(ctx)` — drains the probe and stops the pusher and
  refresh loop. Wire into `ServiceConfig.ShutdownHooks`. Idempotent.
- `Mounted.Ready()`, `Mounted.Probe()`, `Mounted.Dashboard()` — cached
  readiness verdict for `ServiceConfig.ReadyCheck`, plus escape hatches for
  advanced wiring.
- `example/` — runnable demo service (critical + flapping non-critical
  check, dashboard with trend + metrics, full appkit lifecycle wiring),
  verified live end to end.
- 14 tests: probe classification, panic isolation, concurrency, bounded
  contexts; mount routes (probe-only, custom, dashboard), drain behavior,
  lifecycle guards, validation propagation. Race-clean.

### Dependencies

- `github.com/larsartmann/go-health v0.1.1`
- `github.com/larsartmann/go-health-dashboard v0.5.0`
- `github.com/larsartmann/go-error-family v0.10.0`
- Requires `GOEXPERIMENT=jsonv2` (go-health's `encoding/json/v2`
  serialization; dashboard's go-sse).
- The example carries a local `replace github.com/larsartmann/go-appkit =>
  ../` until the core version with `DrainHooks` is published — REMOVE AT
  RELEASE TIME (same procedure as the otel module).
