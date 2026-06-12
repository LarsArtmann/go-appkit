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

- `server.go` — `Server` wrapper around `http.Server`, registers `/health` automatically, supports graceful shutdown.
- `health.go` — `DefaultHealthHandler` and `NewHealthHandler(status)`.
- `shutdown.go` — `WaitForSignal` for SIGINT/SIGTERM handling.
- `logger.go` — `InitLogger` using `log/slog`.
- `sqlite.go` — `OpenSQLite` with default WAL-mode pragmas.

Tests mirror source files: `*_test.go` for each module.

## Architecture and Control Flow

- Configuration is passed as structs (`ServerConfig`, `ShutdownConfig`, `LoggerConfig`, `SQLiteConfig`).
- Each config has a `Default*Config()` constructor that returns opinionated defaults.
- `NewServer` mutates zero-value fields in the passed config to defaults, then registers `GET /health` on the supplied `*http.ServeMux`.
- `Server.Start` blocks until the context is cancelled or the server returns a non-shutdown error.
- `WaitForSignal` blocks on a signal or context cancellation, then invokes the provided shutdown callback with a timeout-bound context.
- Typical usage pattern is shown in `README.md`.

## Conventions and Patterns

- Package name: `appkit`. Import alias is conventional: `appkit "github.com/larsartmann/go-appkit"`.
- Config structs use plain fields; zero values are replaced by defaults inside constructors/functions.
- Error wrapping style: `fmt.Errorf("...: %w", err)`.
- Logging uses `log/slog`; `InitLogger` panics on unsupported level strings.
- SQLite uses `modernc.org/sqlite` (CGO-free), default pragmas include WAL mode, busy timeout, and foreign keys.

## Testing

- Standard `testing` package; no external test frameworks.
- Tests run with `t.Parallel()`.
- HTTP handlers tested with `net/http/httptest`.
- SQLite tests use `t.TempDir()` for transient databases.
- Server tests start a real listener on an ephemeral port via `Server.Addr()` after a short sleep; cancellation is used to stop the server.

## Gotchas

- `NewServer` always registers `GET /health` on the mux passed in, even if you provide a custom `HealthHandler`.
- `Server.Start` binds the port inside the function, not in `NewServer`; `Server.Addr()` returns `nil` until `Start` has been called.
- `InitLogger` panics for unknown log levels; validate input at application startup.
- `OpenSQLite` sets PRAGMAs via string interpolation (`fmt.Sprintf`); values are not parameterized.
- `WaitForSignal` calls `signal.Notify` and `signal.Stop` internally; do not register overlapping signal handlers for the same signals.
- The module has no CI/build scripts; any changes must keep `go test ./...` and `go vet ./...` passing.
