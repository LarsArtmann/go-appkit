# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Fixed

- The `DashboardHardenedPreset` usage example in the godoc passed
  `security.NonceFromContext` (context-based) directly as the request-based
  extractor; the signatures do not match. The example now shows the
  one-line request→context bridge lambda (caught by composing the preset in
  the integration module's hardened-dashboard test).

## [0.1.2] - 2026-09-20

### Added

- `DashboardHardenedPreset(basePath, nonceFn)` — bundles the
  hardened-posture dashboard options (everything under basePath, the
  dashboard's per-request CSP nonce extracted with `nonceFn`). Pair it
  with a nonce-carrying Content-Security-Policy built by the security
  module's `BuildCSP`; the CSP middleware lives OUTSIDE this module (rate
  limiting and security headers belong in the chain in front of the mux —
  the health module stays core-free by decision).
- Docs: the `NewProbe` godoc now leads with an explicit warning that
  `WithHealthRecorder` is silently dropped on the injector-free path
  (go-health's `NewWithHealthCheck` nils the recorder — the flagship
  trace-capture wiring vanishes with zero feedback) and points to the
  working injector path. The runnable
  `ExampleNewProbe_recorderViaInjector` output-pins the guarantee:
  through `health.New` over a samber/do injector, the recorder sees every
  batch. The package doc gains a "Trace capture on health batches"
  section with the `frhealth.Register` + `NewTrigger` + `health.New`
  composition (snippet compile-checked in a scratch module against
  published tags).

### Changed

- Bumped `go-health` v0.1.3 → v0.2.0 (per-check observability metadata
  the dashboard renders; `HealthRecorder`'s interface is unchanged —
  verified in both module sources) and `go-health-dashboard` v0.8.1 →
  v0.9.0 (display-only per-check observability release; JSON and webhook
  wire contracts untouched). Example dependency `go-appkit` core
  v0.4.0 → v0.5.1. `samber/do/v2` moves to a direct require —
  example/test-only import; the module API stays injector-free.
- API delta vs v0.1.1 is additions-only (one new function, docs, dep
  bumps; zero removals, zero signature changes — verified by a
  `go doc -all` diff against the archived tag).

## [0.1.1] - 2026-09-16

### Changed

- `Mounted.Drain` now uses go-health's two-phase API
  (`Probe.MarkShuttingDown`): readiness surfaces flip to 503 immediately,
  while the background refresh loop KEEPS RUNNING, so the cached response
  and dashboard stay fresh during a long drain window. Previously `Drain`
  called `Probe.Shutdown`, stopping the loop and freezing the last cached
  snapshot for the rest of the drain. `Mounted.Shutdown` now explicitly
  calls `Probe.Shutdown` to stop the loop (it previously relied on `Drain`
  doing so). Pinned by `TestMount_DrainKeepsRefreshLoopRunning`.
- Bumped `go-health-dashboard` v0.7.0 → v0.8.1 (operator-trust release:
  failure-evidence truth strip, WCAG AA status colors, consolidated
  nonce-carried bootstrap, templ-components v1.17.0). Every
  `Mounted`/`WithDashboard` symbol is signature-identical between the two
  versions (verified by tag diff), so the bump is drop-in.
- Bumped `go-error-family` v0.10.0 → v0.10.1 (docs/CI-only release, zero
  public API change).

## [0.1.0] - 2026-09-04

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
