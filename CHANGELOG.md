# Changelog

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
