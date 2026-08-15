# Changelog

## [Unreleased]

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
