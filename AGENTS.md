# Agent Notes: go-appkit

This is a small, single-package Go library that provides a shared application skeleton for Go services.

## Project Type

- Go module (`github.com/larsartmann/go-appkit`), Go 1.26.3.
- No `main` package — it is a library intended to be imported by other Go applications.
- All source lives in the repository root under package name `appkit`.

## Essential Commands

There is no Makefile, justfile, CI config, or flake.nix. Use standard Go tooling:

```bash
go test ./...
go vet ./...
go build ./...
go mod tidy
```

## Code Organization

Each file owns one self-contained concern:

- `server.go` — `Server` wrapper around `http.Server`, optional `/health` registration, graceful shutdown, `Running()` status.
- `health.go` — `HealthStatus` typed enum, `DefaultHealthHandler`, `NewHealthHandler(status)`, shared `writeHealthResponse`.
- `shutdown.go` — `WaitForSignal` for SIGINT/SIGTERM handling, logs via `slog`.
- `logger.go` — `LogLevel`/`LogFormat` typed strings, `InitLogger` returns `(*slog.Logger, error)`.
- `sqlite.go` — `OpenSQLite` with WAL-mode pragmas, PRAGMA key allowlist for injection safety.

Tests mirror source files: `*_test.go` for each module.

## Architecture and Control Flow

- Configuration is passed as structs (`ServerConfig`, `ShutdownConfig`, `LoggerConfig`, `SQLiteConfig`).
- Each config has a `Default*Config()` constructor that returns opinionated defaults.
- `ServerConfig.applyDefaults()` fills zero-value fields from defaults (called once, not redundantly).
- `ServerConfig.RegisterHealth` controls whether `/health` is auto-registered (default: true; set false to opt out).
- `Server.ln` is protected by `sync.RWMutex`; use `Addr()` and `Running()` for thread-safe access.
- `InitLogger` returns an error for invalid level/format instead of panicking.
- `WaitForSignal` logs signal receipt via `slog.Info` instead of writing to stderr directly.
- Typical usage pattern is shown in `README.md`.

## Conventions and Patterns

- Package name: `appkit`. Import alias is conventional: `appkit "github.com/larsartmann/go-appkit"`.
- Config structs use typed fields (`LogLevel`, `LogFormat`, `HealthStatus`); zero values are replaced by defaults.
- Error wrapping style: `fmt.Errorf("...: %w", err)`.
- Logging uses `log/slog`; `InitLogger` returns errors for invalid config.
- SQLite uses `modernc.org/sqlite` (CGO-free), default pragmas include WAL mode, busy timeout, and foreign keys.
- PRAGMA keys validated against `allowedPRAGMAs` allowlist to prevent SQL injection.

## Testing

- Standard `testing` package; no external test frameworks.
- Tests run with `t.Parallel()`.
- HTTP handlers tested with `net/http/httptest`.
- SQLite tests use `t.TempDir()` for transient databases.
- Server tests use `freePort()` helper to get ephemeral ports and `waitForAddr()`/`waitForRunning()` polling helpers instead of `time.Sleep`.
- All tests pass with `-race` flag.

## Gotchas

- `NewServer` registers `GET /health` by default; set `ServerConfig.RegisterHealth = false` to opt out.
- `Server.Start` binds the port inside the function, not in `NewServer`; `Server.Addr()` returns `nil` until `Start` has been called.
- `Server.Port: 0` gets overwritten to 8080 by defaults; use an explicit free port in tests.
- `InitLogger` returns errors (not panics) for unknown log levels/formats.
- `OpenSQLite` validates PRAGMA keys against an allowlist; unsupported keys return an error.
- `WaitForSignal` calls `signal.Notify` and `signal.Stop` internally; do not register overlapping signal handlers for the same signals.
- The module has no CI/build scripts; any changes must keep `go test ./... -race` and `go vet ./...` passing.
