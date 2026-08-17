# Agent Notes: go-appkit

Production-ready HTTP service framework composing httputil, charmbracelet/log, and go-error-family.

## Project Type

- Go multi-module repository (`github.com/larsartmann/go-appkit`), Go 1.26.5.
- Six Go modules in one repo, independently versioned:
  - **core** (`/`) — package `appkit`, HTTP service framework. v0.3.0 prepared 2026-08-16 (push pending); v1.0.0 target.
  - **cqrs** (`/cqrs`) — package `cqrs`, CQRS/ES integration via go-cqrs-lite v4. v0.3.0 prepared 2026-08-16 (push pending).
  - **realtime** (`/realtime`) — package `realtime`, SSE transport layer built on go-sse. v0.1.0 prepared 2026-08-16 (push pending; first CHANGELOG added same day).
  - **flightrecorder** (`/flightrecorder`) — package `flightrecorder`, HTTP middleware for Go runtime trace capture. v0.1.0 prepared 2026-08-16 (push pending; first CHANGELOG added same day).
  - **flightrecorderhealth** (`/flightrecorderhealth`) — package `flightrecorderhealth`, bridges go-flightrecorder with go-health: dashboard visibility + auto-capture on health failures. **v0.1.0 tagged** at `d3e3e51` (2026-08-16, push pending).
  - **docs** (`/docs-mod`) — opt-in auto-documentation via catalog/v4. v0.2.0 current (no re-tag needed: since-tag delta is config-only).
  - **errorpages** (`/errorpages`) — pretty classified error pages (HTML/JSON) via templ-components/errorpage. v0.1.0 current (no re-tag needed: since-tag delta is test-only, `83c91bc`).
- Library consumed by Go applications. Reference consumer: cqrs-htmx `setup` (ADR-001 adoption, blocked only on the push).
- Source in repository root. Example in `example/main.go`.
- No Makefile, justfile, CI config, or flake.nix. Use standard Go tooling.
- **Six of seven modules require `GOEXPERIMENT=jsonv2`** (core via httputil/httpspec test dep; cqrs via codec/v4; docs via catalog/v4; errorpages via errorpage; realtime via go-sse → go-branded-id; flightrecorder imports encoding/json/v2 directly).
- **`flightrecorderhealth` does NOT require `GOEXPERIMENT=jsonv2`** — its deps (`go-health`, `samber/do/v2`, `go-flightrecorder`) all use plain `encoding/json`. Builds and tests run with plain `go build`/`go test`.

## Release State (2026-08-16)

- **Wave prepared at `f938d65`:** core v0.3.0 (`NoTimeout` + `ReadyCheck`), cqrs v0.3.0 (staleness guards, storage v4.7.1), realtime v0.1.0, flightrecorder v0.1.0 — all four CHANGELOGs cut, all six modules hermetically verified (build+vet+race, GOWORK=off), four annotated tags at that one commit.
- **PUSH PENDING (user gate):** `git push origin master && git push origin v0.3.0 cqrs/v0.3.0 realtime/v0.1.0 flightrecorder/v0.1.0`. Post-push verification checklist: `TODO_LIST.md` P1 (fresh-consumer proxy test per module + pkg.go.dev).
- **No sibling-require chicken-and-egg:** no module carries `replace` directives; the only sibling require is errorpages → core v0.2.0 (published), so all four tags are independently consumer-valid the moment they land.
- **BuildFlow gotcha:** the pre-commit `dprint` step exits 14 ("no files found to format") on CHANGELOG-only commits because `dprint.json` excludes `**/CHANGELOG.md` — commit with `--no-verify` + justification in that case (tracked in TODO_LIST P3).
- **Cross-repo context:** the setup-vs-appkit comparison (10 findings, all routed) lives at `/home/lars/projects/docs/review/2026-08-16_setup-vs-go-appkit-comparison.md`; execution plan: `docs/planning/2026-08-16_12-04-SUPERB-release-wave-and-harvest.html`.

## Sub-module Build Commands

```bash
# Core (requires GOEXPERIMENT=jsonv2 — httputil/httpspec test dep imports encoding/json/v2)
GOEXPERIMENT=jsonv2 go test ./... -race -count=1
GOEXPERIMENT=jsonv2 go vet ./... && GOEXPERIMENT=jsonv2 go build ./...

# cqrs module (requires GOEXPERIMENT=jsonv2 — codec/v4 uses encoding/json/jsontext)
cd cqrs && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd cqrs && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# docs module (requires GOEXPERIMENT=jsonv2 — catalog/v4 uses encoding/json/v2)
cd docs-mod && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd docs-mod && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# realtime module (requires GOEXPERIMENT=jsonv2)
cd realtime && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd realtime && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...

# errorpages module (requires GOEXPERIMENT=jsonv2 — errorpage uses encoding/json/v2)
cd errorpages && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd errorpages && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# flightrecorder module (requires GOEXPERIMENT=jsonv2 — imports encoding/json/v2 directly)
cd flightrecorder && GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd flightrecorder && GOEXPERIMENT=jsonv2 go vet ./... && GOEXPERIMENT=jsonv2 go build ./...

# flightrecorderhealth module (no GOEXPERIMENT required — plain encoding/json)
go test ./flightrecorderhealth/... -race -count=1
cd flightrecorderhealth && go vet ./... && go build ./...
```

BuildFlow runs as pre-commit hook (auto-fixes formatting/lint on commit).

**Linting (2026-08-17 standard):** every module — root included — carries its own `.golangci.yml`. Lint each module **from its own directory** (`cd <module> && golangci-lint run ./...`); never lint satellites from the workspace root (the root config's depguard allowlist covers only core + family deps, and per-module configs are the source of truth). Root depguard allow now includes `go-appkit` itself (example/ self-import), `go-flightrecorder`, `go-health`, and `go-sse`. All 7 modules sit at **0 issues** (verified 2026-08-17 with `go test -race` green across the board). Shared test-exclusion union for `_test.go`: mnd, exhaustruct, err113, paralleltest, gochecknoglobals, goconst, varnamelen, wsl_v5, funlen, cyclop, testpackage.

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

| File         | Concern                                                                                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`     | Package doc. SSE-only constraint stated explicitly.                                                                                                                             |
| `hub.go`     | `Hub` type: pairs `sse.Broadcaster[sse.Event]` + optional `sse.EventStore`. `NewHub`, `Broadcast`, `BroadcastPatch`, `Shutdown`, `Close`, `Health`.                             |
| `handler.go` | `Handler` (canonical SSE endpoint: CORS→subscribe→replay-with-dedup→heartbeat→forward) + `Mount` convenience for stdlib mux. Functional options for heartbeat, CORS, filtering. |

- **SSE only.** No WebSocket support, provided, or planned.
- Depends on `go-sse v0.5.0` only (no core, no go-datastar, no go-cqrs-lite dependency).
- `BroadcastPatch` uses duck-typed `PatchLike interface { Event() sse.Event }` — works with go-datastar patches without importing go-datastar.
- Handler flushes headers immediately after `NewStream` so clients receive 200 OK without waiting for first event.

## Flightrecorder Module — Code Organization

| File            | Concern                                                                                                                                                 |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`        | Package doc. Process-global singleton constraint, import aliasing guidance.                                                                             |
| `middleware.go` | `Middleware(recorder, trigger, opts...)` — HTTP middleware that captures traces on configurable triggers. Options: error threshold, logger, auto-reset. |
| `handler.go`    | `SnapshotHandler(recorder)` + `Mount(mux, pattern, recorder)` — manual snapshot endpoint with JSON response and reset-before-snapshot.                  |

- Wraps [github.com/larsartmann/go-flightrecorder] with HTTP integration.
- `Middleware` uses `httputil.ResponseRecorder` to capture status codes; converts status >= threshold to `fr.TriggerContext.Err`.
- `Middleware` auto-resets the recorder's once-latch after each capture (default), allowing multiple snapshots. Disable via `WithAutoReset(false)`.
- `SnapshotHandler` resets the latch before snapshotting so manual captures work even after an automatic middleware capture.
- Tests serialized via `recorderMu sync.Mutex` because Go's `runtime/trace` allows only ONE active flight recorder per process.

## Flightrecorderhealth Module — Code Organization

| File                | Concern                                                                                                                                                                                                    |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`            | Package doc. Quick start, import aliasing, process-global singleton note.                                                                                                                                  |
| `adapter.go`        | `Checkable` (health-checkable wrapper, implements `do.HealthcheckerWithContext`), `Trigger` (implements `health.HealthRecorder`), `Register` convenience, options.                                         |
| `adapter_test.go`   | 20 tests: Checkable health states, Trigger capture/no-capture/pass-through, cooldown, logger, custom trigger, Register, integration (incl. real `health.New` Probe end-to-end), concurrency-safe cooldown. |
| `contract_test.go`  | Compile-time assertions: `*Trigger` satisfies `health.HealthRecorder`, `*Checkable` satisfies `do.HealthcheckerWithContext`. Makes contract drift a compile error.                                         |
| `example_test.go`   | 3 runnable godoc examples (`Register`, `NewCheckable`, `NewTrigger`) with verified output — the compile-checked source of truth for doc snippets.                                                          |
| `benchmark_test.go` | `BenchmarkTrigger_RecordHealthCheckWithContext_AllPass` — no-capture hot path (~4.7µs/batch, 2 services).                                                                                                  |
| `.golangci.yml`     | Module lint config reconciled with go-flightrecorder's enable-list (deliberate divergences documented in file header); tests serialized via `recorderMu` due to singleton constraint.                      |
| `README.md`         | Module overview, quick start (compile-checked verbatim in a scratch module), trigger functions, error taxonomy.                                                                                            |
| `CHANGELOG.md`      | keep-a-changelog format with `[Unreleased]` scaffold.                                                                                                                                                      |

- Bridges [github.com/larsartmann/go-flightrecorder] with [github.com/larsartmann/go-health].
- `Checkable` reports the recorder's own operational state (enabled/disabled) as a health-checkable service in the dashboard.
- `Trigger` intercepts every health-check batch via `HealthRecorder.RecordHealthCheckWithContext`, constructs a `fr.TriggerContext` with `Kind="health.check"`, and calls `SnapshotIfAsync` when the trigger function fires.
- `TriggerContext.Err` is set to the first failing service's error so `fr.OnError` fires on failures. `fr.OnAlways` fires on every batch regardless.
- `WithCooldown` prevents trace flooding on flapping services; `lastCapture` is guarded by `sync.Mutex` for concurrent batch safety.
- `WithServiceName` includes an identifier in trigger log messages (multi-trigger setups).
- Nil-safe: nil `Trigger` or nil recorder pass-through to `injector.HealthCheckWithContext`.
- The `go-health` dependency exists solely for the compile-time interface assertion in `contract_test.go` — no runtime usage. If go-health's `HealthRecorder` interface changes, the build breaks instead of failing silently.
- Errors use [go-error-family](https://github.com/LarsArtmann/go-error-family) constructors: `flightrecorder.recorder_missing` is `Rejection`, `flightrecorder.recorder_disabled` is `Infrastructure`.
- Tests use `do.New()` with registered `healthSvc` mocks, `WithMinAge(50ms)` + `WithMaxBytes(1MiB)` + 100ms warmup sleep for trace data.
- Dependencies: `go-flightrecorder v0.2.0`, `go-health v0.0.2`, `samber/do v2.1.0`, `go-error-family v0.10.0`.

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
| `github.com/larsartmann/httputil`        | v0.11.0 | Middleware, health endpoints, Middleware type        |
| `github.com/charmbracelet/log`           | v1.0.0  | Pretty slog handler (Logger implements slog.Handler) |
| `github.com/larsartmann/go-error-family` | v0.10.0 | Error classification, HTTPStatus, LogError           |

## Realtime Module Dependencies

| Module                                   | Version | Role                                                   |
| ---------------------------------------- | ------- | ------------------------------------------------------ |
| `github.com/larsartmann/go-sse`          | v0.5.0  | SSE transport: Stream, Broadcaster, EventStore, Replay |
| `github.com/larsartmann/go-error-family` | v0.10.0 | Error classification (shared with core)                |
| `github.com/larsartmann/go-branded-id`   | v0.5.1  | Phantom-typed EventID (transitive via go-sse)          |

## Flightrecorder Module Dependencies

| Module                                     | Version | Role                                                   |
| ------------------------------------------ | ------- | ------------------------------------------------------ |
| `github.com/larsartmann/go-flightrecorder` | v0.1.1  | Flight recorder core: Recorder, triggers, typed errors |
| `github.com/larsartmann/httputil`          | v0.11.0 | Middleware type, ResponseRecorder for status capture   |

## cqrs Module Dependencies

| Module                                                  | Version | Role                                                                       |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------- |
| `github.com/larsartmann/go-cqrs-lite/stack/v4`          | v4.3.0  | Bundle (events, commands, queries, snapshots, checkpoints)                 |
| `github.com/larsartmann/go-cqrs-lite/stack/sqlite/v4`   | v4.3.0  | SQLite preset (WAL, `SetMaxOpenConns(1)` via `ConfigureSQLitePool`)        |
| `github.com/larsartmann/go-cqrs-lite/projectionhost/v4` | v4.3.0  | Projection host (DLQ, logger, FR, metrics, lag, readiness)                 |
| `github.com/larsartmann/go-cqrs-lite/event/v4`          | v4.7.0  | Event types, stream refs, event construction                               |
| `github.com/larsartmann/go-cqrs-lite/id/v4`             | v4.5.0  | Branded IDs (stream, event)                                                |
| `github.com/larsartmann/go-cqrs-lite/projection/v4`     | v4.3.0  | Projection type and `NewProjection`                                        |
| `github.com/larsartmann/go-cqrs-lite/flightrecorder/v4` | v4.0.0  | CQRS-specific flight recorder (process-global singleton)                   |
| `github.com/larsartmann/go-cqrs-lite/storage/v4`        | v4.7.1  | indirect — v4.7.0 had a build bug (`err :=` in keyset.go), fixed in v4.7.1 |
| `github.com/larsartmann/go-error-family`                | v0.10.0 | Error classification (shared with core)                                    |

- Migrated to v4 on 2026-08-15 (was v3.7.x). Migration guide: go-cqrs-lite `docs/migration/MIGRATION-GUIDE.md`.
- **GOEXPERIMENT=jsonv2 required** (codec/v4 → encoding/json/jsontext).
- v4 codec default flipped JSON→CBOR for new writes; old JSON data still reads (self-describing events). SSE consumers of raw event payloads need CBOR→JSON transcoding.
- v4 sqlite options changed: `WithPragmas`, `WithDSN`, `WithDurability`, `WithBusyTimeout`, `WithCacheSize`, `WithDriverName`, `WithStack` (v3's `WithoutWAL`/`WithOptimizations`/`WithForeignKeys`/`WithoutAutoMigrate`/`WithEventDB` are gone).
- Projection readiness: `EventService.ReadyCheck()` + `EventService.LagPerProjection()`; core `ServiceConfig.ReadyCheck func() bool` composes external checks with the drain probe for `/health/ready`.
- Read-your-writes: projections are async — `EventService.CheckStaleness(budget)` / `CheckProjectionStaleness(name, budget)` are read-time guards (Transient error on lag > budget). projectionhost v4.3.0 has NO public Sync/Drain; cqrs-lint E014's suggested APIs don't exist in the pinned version.
- cqrs-lint: run from inside `cqrs/` — from a workspace root it attributes sub-module imports to the root go.mod (A018 false positive). A/P-series findings on this wrapper (no Save/Publish calls, no WithBatchSize) are by design: consumers get `Bundle()`, tuning flows via `HostOptions`. `.cqrs-lint.json` uses `library-framework` preset (disables ALL F-series adoption-coaching rules) with pinned feature profile and 3 config-level disables (A018, V003, V006). `docs-mod/.cqrs-lint.json` disables A018/A009 (docs module, not an event-sourced app). cqrs-lint source lives at `go-cqrs-lite/cmd/cqrs-lint` — file linter bugs there.
- `EventConfig` full option set: `Logger`, `DLQ *DLQConfig` (SQLite store by default, `ReplayDeadLetters`/`ResetProjection` accessors), `FlightRecorder` (cqrs-lite flightrecorder/v4 type — NOT go-flightrecorder; only one active per process), `Metrics projectionhost.MetricsRecorder` (backend-agnostic lifecycle recorder; no prometheus/otel dep), `HostOptions` passthrough.
- cqrs README ends with a cookbook: scenario DSL decider/projection tests (scenario/v4), testutil helpers (CapturingSlogHandler, DelayedJournal), and cqrs-lint usage — all verified against scenario/v4 v4.2.0, testutil/v4 v4.2.0, cqrs-lint 4.6.0.

## Errorpages Module — Code Organization

| File              | Concern                                                                                                                                   |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`          | Package doc.                                                                                                                              |
| `errorpages.go`   | `Config`, `Mount`, `Wrap`, `Handler`, `Write` bridging templ-components/errorpage to go-error-family classification + Accept negotiation. |
| `example/main.go` | Demo service with pretty 404s and a classified error route.                                                                               |

- Depends on `templ-components/errorpage v1.8.2`, `go-error-family v0.10.0`, `go-appkit v0.2.0` (example only; replace `../` for local dev).
- Family → status identical to `appkit.HTTPStatus` (Rejection 400, Conflict 409, Transient 503, Corruption 500, Infrastructure 503).
- `Wrap` mirrors net/http's `cleanPath` to preserve the mux's path-cleaning redirects (doubled slashes, dot segments); only the canonical request that follows gets the pretty 404.
- Render failures fall back to a plain-text response with the correct status (inherited from errorpage's buffer-before-write rendering).

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
- `NoTimeout` sentinel (`-1`): `ReadTimeout`/`WriteTimeout` = `NoTimeout` disables the deadline (server field 0 AND drops the Timeout middleware from the default stack). Required for SSE services; `ReadHeaderTimeout`/`IdleTimeout` reaping stays on. Only `-1` exactly — other negatives are rejected by Validate.
- BuildFlow auto-fixes lint on commit (gofumpt, golines, gci).

## Flightrecorder Module Gotchas

- **Process-global singleton**: Go's `runtime/trace` allows only ONE active flight recorder per process. Create one at startup.
- **Package name collision**: This package is named `flightrecorder`, same as the underlying library. Alias the underlying library import as `fr`.
- **Once-latch semantics**: The recorder captures only the first snapshot via `sync.Once`. `Middleware` auto-resets after each capture by default; `SnapshotHandler` resets before each manual capture.
- **lazyFile handle caching**: `fr.WithFile` opens the file on first write and caches the handle. Deleting the file and expecting a new file on the next capture does NOT work (the handle stays open). Use `fr.WithWriter(&bytes.Buffer{})` for multi-capture tests.
- No go-appkit core dependency — works on any `*http.ServeMux` or `http.Handler`.

## Realtime Module Gotchas

- **`GOEXPERIMENT=jsonv2` required** to build (transitive via go-sse → go-branded-id). Always prefix commands with it.
- **`GOWORK=off` recommended** if a parent `go.work` includes sibling projects with stale checksums.
- Handler flushes headers immediately after `NewStream` — this is critical for Go HTTP clients and reverse proxies.
- Hub's `BroadcastPatch` accepts any type with `Event() sse.Event` — no go-datastar import needed.
- Default heartbeat is 15s; pass `WithHeartbeat(0)` to disable.
- Default CORS is `*`; tighten via `WithCORSOrigin` for production.
- Shutdown ordering: drain `hub.Shutdown(ctx)` BEFORE `svc.Shutdown(ctx)` so browsers reconnect to another instance.
- Replay/live boundary: the handler subscribes BEFORE reading the replay store and deduplicates replayed IDs in the live loop — no event can slip between store snapshot and subscription. Bursts larger than the subscriber buffer (default 64) during a slow store read can still drop; clients heal via Last-Event-ID reconnect.
- Journal-backed replay (CQRS event stores): wire `transport.NewJournalSSEStore(journal, mapper)` from `github.com/larsartmann/cqrs-htmx/v4/transport` (lean 4-dep sub-package) into `realtime.NewHub(realtime.WithStore(store))`. realtime itself stays cqrs-free.
- The `PatchLike` interface intentionally matches `datastar.Patch` — duck typing avoids the import.
- No go-appkit core dependency — `realtime.Mount` works on any `*http.ServeMux`.
