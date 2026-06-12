# Changelog

## [Unreleased]

### Changed

- `OpenSQLite` now takes a `context.Context` as its first argument.
- Magic numbers extracted to named constants (`defaultPort`, `defaultReadTimeout`, `defaultShutdownTimeout`, etc.).
- All tests use `httptest.NewRequestWithContext`, `http.NewRequestWithContext`, and `*sql.DB.PingContext`/`ExecContext` to honor `noctx` rules.
- `Server.Shutdown` wraps the underlying error (`fmt.Errorf("server shutdown: %w", err)`).
- `WriteHealthResponse` now reports JSON encoding failures via `http.Error` (no more swallowed error).
- `depguard` now allows `modernc.org/sqlite` in `main` ruleset.

## [0.1.0] - 2026-06-12

### Added

- `Server` with configurable timeouts and a `/health` endpoint.
- `DefaultHealthHandler` and `NewHealthHandler` helpers.
- `WaitForSignal` for graceful shutdown on SIGINT/SIGTERM.
- `InitLogger` for structured JSON/text logging.
- `OpenSQLite` for SQLite connections with WAL-mode pragmas.
