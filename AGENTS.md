# Agent Notes: go-appkit

Production-ready HTTP service framework composing httputil, charmbracelet/log, and go-error-family.

## Project Type

- Go multi-module repository (`github.com/larsartmann/go-appkit`), Go 1.26.5.
- Four Go modules in one repo, independently versioned:
  - **core** (`/`) — package `appkit`, HTTP service framework. v1.0.0 target.
  - **cqrs** (`/cqrs`) — package `cqrs`, CQRS/ES integration via go-cqrs-lite. v0.1.0.
  - **realtime** (`/realtime`) — package `realtime`, SSE transport layer built on go-sse. v0.1.0.
  - **docs** (`/docs`) — opt-in auto-documentation. v0.1.0.
- Library consumed by Go applications.
- Source in repository root. Example in `example/main.go`.
- No Makefile, justfile, CI config, or flake.nix. Use standard Go tooling.
- **realtime module requires `GOEXPERIMENT=jsonv2`** (transitive dep via go-sse → go-branded-id).

## Sub-module Build Commands

```bash
# Core (no special flags)
go test ./... && go vet ./... && go build ./...

# cqrs module
cd cqrs && go test ./... && go vet ./... && go build ./...

# realtime module (requires GOEXPERIMENT=jsonv2)
cd realtime && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd realtime && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...
```

BuildFlow runs as pre-commit hook (20 checks).

## Core Module — Code Organization

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

## Realtime Module — Code Organization

| File         | Concern                                                                                                                                                              |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`     | Package doc. SSE-only constraint stated explicitly.                                                                                                                  |
| `hub.go`     | `Hub` type: pairs `sse.Broadcaster[sse.Event]` + optional `sse.EventStore`. `NewHub`, `Broadcast`, `BroadcastPatch`, `Shutdown`, `Close`, `Health`.                  |
| `handler.go` | `Handler` (canonical SSE endpoint: CORS→replay→subscribe→heartbeat→forward) + `Mount` convenience for stdlib mux. Functional options for heartbeat, CORS, filtering. |

- **SSE only.** No WebSocket support, provided, or planned.
- Depends on `go-sse v0.4.0` only (no core, no go-datastar, no go-cqrs-lite dependency).
- `BroadcastPatch` uses duck-typed `PatchLike interface { Event() sse.Event }` — works with go-datastar patches without importing go-datastar.
- Handler flushes headers immediately after `NewStream` so clients receive 200 OK without waiting for first event.

## Architecture and Control Flow

- `Service` owns the mux (`svc.Mux`), logger (`svc.Logger`), and HTTP server.
- Consumer registers routes on `svc.Mux`, then calls `svc.Run(ctx)` which blocks.
- appkit does NOT delegate to `httputil.Server` (it uses `ListenAndServe()` internally, no listener access). appkit owns `http.Server` + `net.Listener` directly for `Addr() net.Addr`.
- httputil is used for: middleware (`Chain`, `Recovery`, `Logging`, etc.), health (`RegisterHealth`), and types (`Middleware`).
- `ServiceConfig.RegisterHealth` is `*bool`: nil or `&true` = register health, `&false` = opt out.
- Graceful drain: `Shutdown()` flips `readyProbe` to false → waits `DrainDelay` → `server.Shutdown(ctx)`.
- `Run()` uses `signal.NotifyContext` for SIGINT/SIGTERM internally.
- Errors use go-error-family constructors (`NewRejection`, `WrapInfrastructuref`) instead of `fmt.Errorf`.

## Core Dependencies

| Module                                   | Version | Role                                                 |
| ---------------------------------------- | ------- | ---------------------------------------------------- |
| `github.com/larsartmann/httputil`        | v0.9.1  | Middleware, health endpoints, Middleware type        |
| `github.com/charmbracelet/log`           | v1.0.0  | Pretty slog handler (Logger implements slog.Handler) |
| `github.com/larsartmann/go-error-family` | v0.10.0 | Error classification, HTTPStatus, LogError           |

## Realtime Module Dependencies

| Module                                   | Version | Role                                                   |
| ---------------------------------------- | ------- | ------------------------------------------------------ |
| `github.com/larsartmann/go-sse`          | v0.4.0  | SSE transport: Stream, Broadcaster, EventStore, Replay |
| `github.com/larsartmann/go-error-family` | v0.10.0 | Error classification (shared with core)                |
| `github.com/larsartmann/go-branded-id`   | v0.5.1  | Phantom-typed EventID (transitive via go-sse)          |

## cqrs Module Dependencies

Depends on `go-cqrs-lite` v3 packages (stack/sqlite, projectionhost, stack). Not yet migrated to v4 (system). See `docs/planning/realtime-sse-design.md` for the migration question.

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

## Realtime Module Gotchas

- **`GOEXPERIMENT=jsonv2` required** to build (transitive via go-sse → go-branded-id). Always prefix commands with it.
- **`GOWORK=off` recommended** if a parent `go.work` includes sibling projects with stale checksums.
- Handler flushes headers immediately after `NewStream` — this is critical for Go HTTP clients and reverse proxies.
- Hub's `BroadcastPatch` accepts any type with `Event() sse.Event` — no go-datastar import needed.
- Default heartbeat is 15s; pass `WithHeartbeat(0)` to disable.
- Default CORS is `*`; tighten via `WithCORSOrigin` for production.
- Shutdown ordering: drain `hub.Shutdown(ctx)` BEFORE `svc.Shutdown(ctx)` so browsers reconnect to another instance.
- The `PatchLike` interface intentionally matches `datastar.Patch` — duck typing avoids the import.
- No go-appkit core dependency — `realtime.Mount` works on any `*http.ServeMux`.
