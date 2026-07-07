# Agent Notes: go-appkit

Production-ready HTTP service framework composing httputil, charmbracelet/log, and go-error-family.

## Project Type

- Go module (`github.com/larsartmann/go-appkit`), Go 1.26.4.
- Library (package `appkit`) imported by Go applications.
- Source in repository root. Example in `example/main.go`.
- No Makefile, justfile, CI config, or flake.nix. Use standard Go tooling.

## Essential Commands

```bash
go test ./...
go vet ./...
go build ./...
go mod tidy
```

BuildFlow runs as pre-commit hook (20 checks).

## Code Organization

| File              | Concern                                                                                                                                 |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `service.go`      | `Service` type: NewService, Start, Run, Shutdown, Close, Addr, Running. Owns `http.Server` + `net.Listener` + `readyProbe atomic.Bool`. |
| `config.go`       | `ServiceConfig` struct, `DefaultServiceConfig()`, `applyDefaults()`, `Validate()`.                                                      |
| `middleware.go`   | `defaultMiddlewareStack()` + `buildMiddleware()`. Default: Recovery→RequestID→Logging→Timeout→SecurityHeaders.                          |
| `logger.go`       | `LogLevel`/`LogFormat` types, `InitLogger()` using charmbracelet/log (Logger IS slog.Handler).                                          |
| `health.go`       | `RegisterHealth(mux)` delegates to httputil. `ReadyHandlerWithProbe(ready)`.                                                            |
| `errors.go`       | Re-exports `HTTPStatus()` and `LogError()` from go-error-family.                                                                        |
| `shutdown.go`     | `WaitForSignal()` for SIGINT/SIGTERM. Preserved for backward compat.                                                                    |
| `doc.go`          | Package doc statement.                                                                                                                  |
| `example/main.go` | 12-line demo service.                                                                                                                   |

## Architecture and Control Flow

- `Service` owns the mux (`svc.Mux`), logger (`svc.Logger`), and HTTP server.
- Consumer registers routes on `svc.Mux`, then calls `svc.Run(ctx)` which blocks.
- appkit does NOT delegate to `httputil.Server` (it uses `ListenAndServe()` internally, no listener access). appkit owns `http.Server` + `net.Listener` directly for `Addr() net.Addr`.
- httputil is used for: middleware (`Chain`, `Recovery`, `Logging`, etc.), health (`RegisterHealth`), and types (`Middleware`).
- `ServiceConfig.RegisterHealth` is `*bool`: nil or `&true` = register health, `&false` = opt out.
- Graceful drain: `Shutdown()` flips `readyProbe` to false → waits `DrainDelay` → `server.Shutdown(ctx)`.
- `Run()` uses `signal.NotifyContext` for SIGINT/SIGTERM internally.
- Errors use go-error-family constructors (`NewRejection`, `WrapInfrastructuref`) instead of `fmt.Errorf`.

## Dependencies

| Module                                   | Version | Role                                                 |
| ---------------------------------------- | ------- | ---------------------------------------------------- |
| `github.com/larsartmann/httputil`        | v0.5.0  | Middleware, health endpoints, Middleware type        |
| `github.com/charmbracelet/log`           | v1.0.0  | Pretty slog handler (Logger implements slog.Handler) |
| `github.com/larsartmann/go-error-family` | v0.6.1  | Error classification, HTTPStatus, LogError           |

## Testing

- Standard `testing` package; no external frameworks.
- Tests run with `t.Parallel()`.
- Server tests use `freePort()` and `waitForRunning()` helpers (no `time.Sleep`).
- All tests pass with `-race` flag and `-count=1`.
- Default `DrainDelay` (5s) makes some tests slow; use `DrainDelay: 0` in non-drain tests.

## Gotchas

- `NewService` registers `/health`, `/health/live`, `/health/ready` by default.
- `Service.Addr()` returns `nil` before `Start()` is called.
- `RegisterHealth` is `*bool` — use `&false` to opt out, not `false`.
- `InitLogger` returns errors (not panics) for invalid config.
- charmbracelet/log `Logger` directly implements `slog.Handler` — no adapter needed.
- `Service.Shutdown` is idempotent — safe to call multiple times.
- BuildFlow auto-fixes lint on commit (gofumpt, golines, gci).
