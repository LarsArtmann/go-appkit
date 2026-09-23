# Changelog

## [Unreleased]

### Added

- `testkit.TestServer.Shutdown`: stop the harness service at call time —
  the stop sequence (server-error drain + goroutine-baseline assertion)
  runs once through a `sync.Once` shared with the `t.Cleanup`, so tests
  that assert post-shutdown behavior no longer pay a second shutdown wait
  in cleanup. Idempotent; later calls return nil.

## [0.5.1] - 2026-09-17

### Fixed

- `Service.Addr` and `Service.Running` godoc now state the full lifecycle
  contract: the listener is reaped at the START of `Shutdown` (before the
  drain hooks run), so both accessors flip to nil/false for the whole
  drain window — capture the base URL before calling `Shutdown`. The
  behavior itself is unchanged (it was always this; the composition-
  contract suite pins it) — only the documentation lied. Doc-only
  release: `go doc -all` diff vs v0.5.0 is the two comment blocks, zero
  API delta.

## [0.5.0] - 2026-09-17

### Added

- Opt-in Prometheus metrics surface (`ServiceConfig.Metrics`): a
  dependency-free text exposition at `GET /metrics` (default) with the
  stable-contract metric names `appkit_http_request_duration_seconds`
  (histogram by method/route/status), `appkit_http_responses_total`,
  `appkit_http_requests_in_flight`, and `appkit_build_info` (version label).
  Route labels use the ServeMux pattern (cardinality-bounded; unmatched
  paths collapse to `unmatched`), and Basic Auth is mandatory by default —
  an unauthenticated configuration is a construction Rejection unless
  `AllowUnauthenticated` is set explicitly. See README "Metrics" for the
  OTEL `_ratio` exporter trap.
- `ServiceConfig.Version` + `GET /version` (JSON): the F5 build-info
  battery. Also labels `appkit_build_info`.
- `testkit` sub-package (`github.com/larsartmann/go-appkit/testkit`):
  `testkit.Serve(t, svc)` starts the REAL service (full middleware chain —
  the raw-mux `httptest.NewServer(svc.Mux)` trap is encoded as API) and
  registers teardown with a goroutine-baseline leak assert.

- Shutdown phase logging: every phase of `Service.Shutdown` emits one INFO
  line ("shutdown phase complete" carrying the phase name and its duration) —
  `ready_flip`, `drain_hooks`, `drain_wait` (or a `shutdown phase skipped`
  line when `NoDrainDelay` is set), `listener_close`, `shutdown_hooks` —
  followed by a "graceful shutdown complete" line with the total duration and
  the joined `result` (`ok`/`error`), so deploys are diagnosable from logs
  alone. A service that never started still logs nothing and runs no hooks.
- README: the JSON v2 build note now warns that gopls running without
  `GOEXPERIMENT=jsonv2` (older-toolchain setups) reports FALSE `string
  literal not terminated` diagnostics on files importing `encoding/json/v2`,
  and the graceful drain sequence documents the hook phases plus the
  per-phase log lines.

### Changed

- Bumped `go-error-family` v0.10.0 → v0.10.1 (docs/CI-only release
  upstream: zero code changes; readme/licensing/CI work). Hygiene pin — no
  consumer-visible behavior change.

## [0.4.0] - 2026-09-04

### Added

- `ServiceConfig.OuterMiddlewares`: middlewares that wrap the entire chain —
  including the default stack or a configured `Middlewares` replacement — and
  run outermost, before Recovery. The hook point instrumentation (OpenTelemetry
  tracing, see the new `otel` module) needs to observe the full request
  lifetime and seed context for everything downstream.
- `ServiceConfig.ShutdownHooks`: run once, in order, after the server shut
  down and released its connections; errors are joined and classified as
  infrastructure failures. The canonical use is flushing telemetry providers
  so their spans cover the final in-flight requests. A service that never
  started does not run its hooks.
- `ServiceConfig.DrainHooks`: run once, in order, at the start of the
  shutdown drain — after the ready probe flips to false but before the drain
  wait and listener close. The hook point the `health` module needs to flip
  mounted go-health readiness surfaces in lockstep with the framework's own
  probe, so every readiness endpoint reports 503 for the whole drain window;
  errors are joined with the shutdown result. A service that never started
  does not run its hooks.
- `NoDrainDelay` sentinel: skips the drain wait in `Shutdown`. Zero is not
  "no delay" — it applies the 5s default — so tests previously paid 5s per
  shutdown; `NoDrainDelay` is the explicit opt-out (the ready probe still
  flips immediately). The core test suite now uses it throughout.
- `otel` module (package `otel`, import alias `appkitotel`): opt-in
  OpenTelemetry support — provider `Setup`, an `otelhttp` middleware bridge
  (spans + semantic-convention metrics + W3C propagation, health endpoints
  filtered by default), HTTP histogram views, and slog trace correlation
  (`TraceHandler`). No core dependency: works on any `http.Handler`.
- `health` module (package `health`, import alias `appkithealth`): bridges
  go-health (three-probe checks with critical/non-critical classification,
  background caching, startup latch, shutdown awareness) and the
  go-health-dashboard real-time UI (SSE, trend, Prometheus, webhooks) into
  appkit services. `NewProbe` for injector-free checks, `New`/`Mount` for
  mux wiring, `Drain`/`Shutdown` wired via `DrainHooks`/`ShutdownHooks`. No
  core dependency; requires `GOEXPERIMENT=jsonv2`.

### Fixed

- Resolved all golangci-lint findings (exhaustruct, gochecknoinits, noctx, noinlineerr, wrapcheck, varnamelen): context-aware `net.ListenConfig`/`NewRequestWithContext` in code and tests, removed the httpspec `init()` workaround, justified nolint directives on deliberate zero-value structs, depguard allowlist now covers the module family.


## [0.3.0] - 2026-08-16

> Minor bump: two new opt-in APIs (`NoTimeout`, `ServiceConfig.ReadyCheck`); no
> breaking changes. This is the release the cqrs-htmx `setup` adoption (ADR-001)
> waits on.

### Added

- `NoTimeout` sentinel and opt-out for `ReadTimeout`/`WriteTimeout`: assign
  `appkit.NoTimeout` to disable the deadline (server field and the default
  stack's Timeout middleware) for long-lived responses such as SSE streams.
  `ReadHeaderTimeout`/`IdleTimeout` reaping stays enabled.
- `ServiceConfig.ReadyCheck`: optional gate consulted by `/health/ready` in addition to
  the drain probe — e.g. `cqrs.EventService.ReadyCheck` keeps readiness 503 until
  projections are live.

### Changed

- `Shutdown` derives its timeout context from the caller's context via
  `context.WithoutCancel`, so context values (tracing, logger) survive into the
  shutdown sequence. Behavior is otherwise unchanged.
- Tests build requests with `NewRequestWithContext` throughout (noctx hygiene).

## [0.2.0] - 2026-07-26

> Reconstructed after the fact: the tag was cut without a changelog section.
> The former "Unreleased" items below 0.1.0 (ctx on `OpenSQLite`, `Server.Shutdown`
> wrapping, etc.) were superseded by this rewrite — every `Server`-era API was removed.

### Changed — complete rewrite as a service framework

- `Server` → `Service`: owns `http.Server` + `net.Listener` (`Addr() net.Addr`), the
  service mux (`svc.Mux`), and the logger (`svc.Logger`).
- Middleware via [httputil](https://github.com/LarsArtmann/httputil): Recovery →
  RequestID → Logging → Timeout → SecurityHeaders, replaceable (`Middlewares`) or
  extendable (`ExtraMiddlewares`).
- Logging via [charmbracelet/log](https://github.com/charmbracelet/log) (`InitLogger`,
  `LogLevel`, `LogFormat`); the logger doubles as an `slog.Handler`.
- Health endpoints `/health`, `/health/live`, `/health/ready` with a drain-aware
  readiness probe (`RegisterHealth: *bool` to opt out).
- Graceful drain sequence: ready probe flips 503 → `DrainDelay` → `server.Shutdown`;
  `Shutdown` is idempotent.
- Errors classified via [go-error-family](https://github.com/LarsArtmann/go-error-family);
  `HTTPStatus` and `LogError` re-exported.
- Repository split into independently versioned modules: `cqrs` (go-cqrs-lite
  integration) and `docs` (catalog/docserver auto-documentation).

## [0.1.0] - 2026-06-12

### Added

- `Server` with configurable timeouts and a `/health` endpoint.
- `DefaultHealthHandler` and `NewHealthHandler` helpers.
- `WaitForSignal` for graceful shutdown on SIGINT/SIGTERM.
- `InitLogger` for structured JSON/text logging.
- `OpenSQLite` for SQLite connections with WAL-mode pragmas.
