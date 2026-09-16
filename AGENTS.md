# Agent Notes: go-appkit

Production-ready HTTP service framework composing httputil, charmbracelet/log, and go-error-family.

## Project Type

- Go multi-module repository (`github.com/larsartmann/go-appkit`), Go 1.26.7.
- Ten Go modules in one repo, independently versioned (nine released + the unreleased `integration` test module):
  - **core** (`/`) — package `appkit`, HTTP service framework. v0.4.0 (2026-09-04, pushed: lifecycle hooks OuterMiddlewares/ShutdownHooks/DrainHooks + NoDrainDelay); v1.0.0 target.
  - **cqrs** (`/cqrs`) — package `cqrs`, CQRS/ES integration over go-cqrs-lite's `system` engine. v0.5.0 (2026-09-07, pushed: BREAKING — engine room moved off the deprecated `stack/sqlite` preset onto `system.New`; `Bundle()` → `System()`; `SQLitePath` → `DSN`/`Driver`/`Pragmas`; adds the typed command/query facade and operator config via `ConfigPath`/`Deployment`). v0.4.0 had aligned `FlightRecorder` to `*go-flightrecorder.Recorder`.
  - **realtime** (`/realtime`) — package `realtime`, SSE transport layer built on go-sse. v0.1.0 (pushed, proxy-verified).
  - **otel** (`/otel`) — package `otel` (alias `appkitotel`), OpenTelemetry provider setup + otelhttp middleware bridge + trace-correlated logging. v0.1.1 (2026-09-16, pushed: httputil v1.2.0 bump ships pattern-named spans + `http.route` through `OuterMiddlewares`; no API changes). Has `benchmark_test.go` — no-op ~21µs vs full tracing+metrics ~27µs per request (see otel README Performance).
  - **flightrecorder** (`/flightrecorder`) — package `flightrecorder`, HTTP middleware for Go runtime trace capture. v0.1.0 (pushed, proxy-verified).
  - **flightrecorderhealth** (`/flightrecorderhealth`) — package `flightrecorderhealth`, bridges go-flightrecorder with go-health: dashboard visibility + auto-capture on health failures. v0.1.1 (2026-09-04, pushed; go-health v0.0.2 → v0.1.1).
  - **health** (`/health`) — package `health` (alias `appkithealth`), bridges go-health probes + go-health-dashboard real-time UI into appkit services. v0.1.0 (2026-09-04, pushed; example consumes published core v0.4.0). See the Health Module section below.
  - **docs** (`/docs`) — opt-in auto-documentation via catalog/v4. v0.3.0 (2026-09-16, pushed: UN-GHOSTED — path-A repath `docs-mod/` → `docs/` so the module path `.../docs` matches the directory; the v0.2.0 tag was UNFETCHABLE from the proxy). See the docs-ghost entry in TODO_LIST P1 for the history.
  - **errorpages** (`/errorpages`) — pretty classified error pages (HTML/JSON) via templ-components/errorpage. v0.1.0 current (no re-tag needed: since-tag delta is test-only, `83c91bc`).
  - **integration** (`/integration`) — cross-module + cross-repo E2E composition tests, never released. Pins PUBLISHED tags (core + realtime + cqrs-htmx/transport + otel) so it always tests what consumers resolve; does NOT require `GOEXPERIMENT=jsonv2` (transport is lean; the otel pin does not change this). Added 2026-09-04. See the Integration Module section below.
- Library consumed by Go applications. Reference consumer: cqrs-htmx `setup` (ADR-001 adoption decided; consumes go-appkit v0.3.0 from the module proxy since their v4.9.0 train; their `RunWithAppkit` fold-in is pending on the cqrs-htmx side).
- Source in repository root. Example in `example/main.go`.
- CI: `.github/workflows/ci.yml` (per-module build/vet/test matrix under `GOEXPERIMENT=jsonv2`, fresh-consumer proxy smoke, cqrs-lint job; md/docs paths ignored) + `.github/dependabot.yml` (weekly grouped gomod updates per module). No Makefile, justfile, or flake.nix — use standard Go tooling locally.
- **`encoding/json/v2` is DEFAULT-ON in Go 1.26.7** (verified 2026-09-04: every module builds plain `go build` with no GOEXPERIMENT; `go env GOEXPERIMENT` reports `jsonv2`). The per-module `GOEXPERIMENT=jsonv2` prefixes below are only needed on OLDER 1.26.x toolchains where jsonv2 was still gated — kept for completeness, drop them when the toolchain floor moves past 1.26.7.

## Release State

- Every module tag is ON ORIGIN through **cqrs v0.5.0** (2026-09-07, system-based engine): core v0.4.0, cqrs v0.5.0, realtime v0.1.0, otel v0.1.1 (2026-09-16 pattern fix), health v0.1.0, flightrecorder v0.1.0, flightrecorderhealth v0.1.1, errorpages v0.1.0, docs v0.2.0 (ghost — see module list; un-ghost fix in flight). Wave history and per-version deltas live in the module CHANGELOGs; the pending-work queue lives in TODO_LIST.
- **pkg.go.dev:** all submodule pages 404'd and core rendered `License: UNKNOWN` with godoc hidden (pkg.go.dev does not index modules without a module-root LICENSE; an unclassifiable proprietary LICENSE hides core godoc). LICENSE files landed in every module root 2026-09-04 and ship in the wave-2+ tags. **License DECIDED 2026-09-16: stays proprietary/unlicensed** (Lars's ruling; no MIT swap) — godoc stays hidden by choice, permanently; do not re-open or re-ask. Remaining: only the pkg.go.dev re-crawl/render check (pages exist, 404s gone).
- **Tag hygiene:** no module requires an UNRELEASED sibling (errorpages → core is a published-tag require), so every tag is independently consumer-valid. NEVER tag a module whose go.mod carries a filesystem `replace` — working-tree replaces used for cross-repo debugging (e.g. the otel → local httputil replace used while developing the pattern-propagation fix) must be removed before tagging.
- **Cross-repo context:** the setup-vs-appkit comparison (10 findings, all routed) lives at `/home/lars/projects/docs/review/2026-08-16_setup-vs-go-appkit-comparison.md`; execution plan: `docs/planning/2026-08-16_12-04-SUPERB-release-wave-and-harvest.html`.
- **Sibling-project integration research (2026-09-04):** cordis (reactive DI meta-framework, zero-dep, untagged), go-plugin-mvp (Kernovia marketplace, proprietary, pre-1.0), and PapDashboard (event-sourced notification hub, published app `papdashboard` v0.2.0) assessed as potential integrations — verdict: none as a go-appkit dependency; reverse adoption (consumer repo hosts on appkit) recommended for go-plugin-mvp AND PapDashboard (the latter with version-identical family deps; their cqrshtmx root `Chain` stays appkit-free even at cqrs-htmx v4.9.0; surfaced gap: core has no TLS support). Full analyses + re-entry triggers: `docs/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md`, `docs/planning/2026-09-04_papdashboard-integration.md`; tracked in TODO_LIST P3.
- **Battery program:** `docs/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md` (CV consumer perspective, 40 proposed batteries in 7 clusters) is the CANONICAL spec — waves routed to TODO_LIST P2/P3; anti-recommendations hold: no security in the default stack, no DI/templ/cron/WebSocket/storage-engine in core.

## Sub-module Build Commands

```bash
# Core (requires GOEXPERIMENT=jsonv2 — httputil/httpspec test dep imports encoding/json/v2)
GOEXPERIMENT=jsonv2 go test ./... -race -count=1
GOEXPERIMENT=jsonv2 go vet ./... && GOEXPERIMENT=jsonv2 go build ./...

# cqrs module (requires GOEXPERIMENT=jsonv2 — codec/v4 uses encoding/json/jsontext)
cd cqrs && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd cqrs && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# docs module (requires GOEXPERIMENT=jsonv2 — catalog/v4 uses encoding/json/v2)
cd docs && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd docs && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# realtime module (requires GOEXPERIMENT=jsonv2)
cd realtime && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd realtime && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...

# otel module (no GOEXPERIMENT required — plain encoding/json)
cd otel && GOWORK=off go test ./... -race -count=1
cd otel && GOWORK=off go vet ./... && GOWORK=off go build ./...

# errorpages module (requires GOEXPERIMENT=jsonv2 — errorpage uses encoding/json/v2)
cd errorpages && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd errorpages && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# flightrecorder module (requires GOEXPERIMENT=jsonv2 — imports encoding/json/v2 directly)
cd flightrecorder && GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd flightrecorder && GOEXPERIMENT=jsonv2 go vet ./... && GOEXPERIMENT=jsonv2 go build ./...

# flightrecorderhealth module (GOEXPERIMENT=jsonv2 required since the go-health v0.1.1 bump, 2026-09-04)
cd flightrecorderhealth && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd flightrecorderhealth && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# health module (requires GOEXPERIMENT=jsonv2 — go-health json/v2 + go-sse)
cd health && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1
cd health && GOWORK=off GOEXPERIMENT=jsonv2 go vet ./... && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...

# integration module (no GOEXPERIMENT required — cqrs-htmx/transport is lean; pins PUBLISHED tags)
cd integration && GOWORK=off go test ./... -race -count=1
cd integration && GOWORK=off go vet ./... && GOWORK=off golangci-lint run ./...
```

BuildFlow runs as pre-commit hook (auto-fixes formatting/lint on commit).

**Linting (2026-08-18 standard, exhaustruct_v5 2026-09-04):** every module — root included — carries its own `.golangci.yml`. Lint each module **from its own directory** (`cd <module> && golangci-lint run ./...`); never lint satellites from the workspace root (the root config's depguard allowlist covers only core + family deps, and per-module configs are the source of truth). Run **one module's lint at a time** — concurrent golangci-lint processes race on the shared build cache (`/mnt/buildcache`) and produce flickering or incomplete findings (verified 2026-08-18: a parallel batch under-reported 3 of 7 real findings); if results look wrong, re-run sequentially. Root depguard allow now includes `go-appkit` itself (example/ self-import), `go-flightrecorder`, `go-health`, and `go-sse`. All modules sit at **0 issues** with `go test -race` green across the board (re-verified 2026-09-04 after the exhaustruct → exhaustruct_v5 migration: v5 does not honor old `//nolint:exhaustruct` directives — they were all renamed; the root `http.Server` literal carries an explicit v5 nolint for its deliberate zero fields). otel's config extends the ireturn allow-list with `go\.opentelemetry\.io/otel/.*` — the OTel API is interface-based by design. Shared test-exclusion union for `_test.go`: mnd, exhaustruct_v5, err113, paralleltest, gochecknoglobals, goconst, varnamelen, wsl_v5, funlen, cyclop, testpackage.

**Structure linting:** `go-structure-linter . --exclude root-package-files --exclude internal-directory --exclude examples-directory` from the repo root → 0 findings. The root package IS the public `appkit` package, so those three rules are wrong for this repo. The linter also caps AGENTS.md length (≤377 counted lines) — keep this file lean, extract detail to module READMEs; the yaml config file path (`exclude_patterns`) is inert in the installed binary despite the source comment saying "removes rules by name" — use the CLI flags until the binary updates.

## Core Module — Code Organization

| File              | Concern                                                                                                                                                                                                                                                            |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `service.go`      | `Service` type: NewService, Start, Run, Shutdown (runs ShutdownHooks once, after connections are released; errors joined; emits per-phase INFO lines — see Gotchas), Close, Addr, Running. Owns `http.Server` + `net.Listener` + `readyProbe atomic.Bool`.         |
| `config.go`       | `ServiceConfig` struct (`OuterMiddlewares`, `DrainHooks`, `ShutdownHooks`), `DefaultServiceConfig()`, `applyDefaults()`, `Validate()`. Sentinels: `NoTimeout` (-1), `NoDrainDelay` (-2).                                                                           |
| `middleware.go`   | `defaultMiddlewareStack()` + `buildMiddleware()` + `concatMiddlewares()`. Order: OuterMiddlewares → (Middlewares replacement OR default stack + ExtraMiddlewares); `concatMiddlewares` allocates a fresh slice so config backing arrays are never written through. |
| `logger.go`       | `LogLevel`/`LogFormat` types, `InitLogger()` using charmbracelet/log (Logger IS slog.Handler).                                                                                                                                                                     |
| `health.go`       | `RegisterHealth(mux)` delegates to httputil. `ReadyHandlerWithProbe(ready)`.                                                                                                                                                                                       |
| `errors.go`       | Re-exports `HTTPStatus()` and `LogError()` from go-error-family.                                                                                                                                                                                                   |
| `shutdown.go`     | `WaitForSignal()` for SIGINT/SIGTERM. Preserved for backward compat.                                                                                                                                                                                               |
| `doc.go`          | Package doc statement.                                                                                                                                                                                                                                             |
| `example/main.go` | 12-line demo service.                                                                                                                                                                                                                                              |

## Realtime Module — Code Organization

| File         | Concern                                                                                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`     | Package doc. SSE-only constraint stated explicitly.                                                                                                                             |
| `hub.go`     | `Hub` type: pairs `sse.Broadcaster[sse.Event]` + optional `sse.EventStore`. `NewHub`, `Broadcast`, `BroadcastPatch`, `Shutdown`, `Close`, `Health`.                             |
| `handler.go` | `Handler` (canonical SSE endpoint: CORS→subscribe→replay-with-dedup→heartbeat→forward) + `Mount` convenience for stdlib mux. Functional options for heartbeat, CORS, filtering. |
| `README.md`  | Module overview, quick start, options table, shutdown ordering, known limitations.                                                                                              |

- **SSE only.** No WebSocket support, provided, or planned.
- Depends on `go-sse v0.6.0` only (no core, no go-datastar, no go-cqrs-lite dependency).
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
- Dependencies: `go-flightrecorder v0.2.0`, `go-health v0.1.3` (2026-09-16; contract assertions unchanged), `samber/do v2.1.0`, `go-error-family v0.10.1`.

## Health Module — Code Organization

| File       | Concern                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`   | Package doc: quick start, route map, lifecycle ordering (Start → Run; DrainHooks → ShutdownHooks), aliasing guidance, GOEXPERIMENT note.                                                                                                                                                                                                                                                                                                                                                    |
| `probe.go` | `NewProbe(checks, opts...)` — injector-free go-health probe from a `map[string]CheckFunc`; concurrent per-check batches (`wg.Go`), per-check panic isolation (`errorfamily.Infrastructure` `health.check_panicked`), SDK options pass through (`WithCriticalServices`, `WithTimeout`, …).                                                                                                                                                                                                   |
| `mount.go` | `New(probe, opts...)` + `Mounted.RegisterRoutes(mux)` + `Mount(mux, ...)` sugar; `Mounted` lifecycle: `Start(ctx)` (initial sync batch + refresh + pusher), `Drain()` (probe MarkShuttingDown: readiness → 503, refresh loop KEEPS RUNNING so the cache stays fresh), `Shutdown(ctx)` (idempotent, re-Start legal; explicitly stops the refresh loop via probe.Shutdown), `Ready()`, `Probe()`, `Dashboard()`; options `WithDashboard(opts...)` (opt-in), `WithProbeRoutes(health.Routes)`. |
| `example/` | Demo service: critical + flapping non-critical check, dashboard w/ trend + metrics, full appkit wiring; verified live E2E (lockstep drain 503s). Consumes PUBLISHED core (no replace directives).                                                                                                                                                                                                                                                                                           |

- NO core dependency (mount works on any `*http.ServeMux`); the appkit composition is config-level (`DrainHooks`/`ShutdownHooks`/`ReadyCheck`).
- Consumers alias this module as `appkithealth` when they also import go-health (both packages are named `health`).
- Dashboard is opt-in (`WithDashboard`); it then registers the probe endpoints from ITS route config (WithBasePath applies uniformly) and serves `/health` — consumers must set `RegisterHealth: &false` (mux panics on the duplicate otherwise).
- Without dashboard: probe routes only (`/healthz`, `/readyz`, `/startupz`), coexists with appkit's default health endpoints.
- `Mounted.Drain` in `DrainHooks` = go-health readiness 503 for the WHOLE drain window, in lockstep with appkit's own ready probe (the reason core gained `DrainHooks`).
- Dependencies: `go-health v0.1.3`, `go-health-dashboard v0.8.1` (2026-09-16 bump: operator-trust release — truth strip, WCAG AA, nonce bootstrap), `go-error-family v0.10.1`.

## otel Module — Code Organization

| File              | Concern                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`          | Package doc: wiring recipe, emitted signals, shutdown ordering, log correlation, one-Setup-per-process rule.                                                                                                                                                                                                                                                                                                                                                                                                            |
| `setup.go`        | `Setup(opts...)` + `Provider` (`AsTracerProvider`/`AsMeterProvider`/`Shutdown`). Options: `WithService`, `WithSpanExporter`, `WithSampler`, `WithMetricReader`, `WithPropagator`, `WithStdoutExporter`, `WithoutGlobalRegistration`. Registers globals + W3C propagator; Shutdown ForceFlushes BOTH providers first.                                                                                                                                                                                                    |
| `middleware.go`   | `Middleware(opts...)` bridging otelhttp v0.71. Span named by matched ServeMux pattern (`r.Pattern`, e.g. `GET /users/{id}`) falling back to method — on PUBLISHED httputil this works only adjacent to the mux; fixed upstream in httputil master (pattern propagation, 2026-09-16, ships with v1.2 — see otel Gotchas). Options: `WithTracerProvider`, `WithMeterProvider`, `WithServerName`, `WithPublicEndpoint` (remote parents → links), `WithFilter`, `WithFilteredPaths`. Health paths unconditionally filtered. |
| `logging.go`      | `TraceHandler` slog decorator stamping `trace_id`/`span_id` when ctx carries a span; `TraceIDFromContext`/`SpanIDFromContext`/`ContextLogger` (return `"none"` without a span).                                                                                                                                                                                                                                                                                                                                         |
| `views.go`        | `NewHTTPViews()` pinning `http.server.request.duration` to `HTTPDurationBoundaries` (semconv 0..10s, 15 values); exact-name match only.                                                                                                                                                                                                                                                                                                                                                                                 |
| `attributes.go`   | `ServiceResourceAttributes` (semconv v1.26.0), `NewTextMapPropagator` (TraceContext+Baggage).                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `example/main.go` | Demo service honoring `PORT` env (8080 occupied on dev machines); E2E-verified live.                                                                                                                                                                                                                                                                                                                                                                                                                                    |

- Strictly opt-in: without a provider the middleware is a no-op and globals stay untouched.
- Library code has NO core dependency (`Middleware` fits any `http.Handler`); the module and example consume PUBLISHED core (v0.4.0) — no replace directives in the tagged module. Never tag a go.mod carrying a filesystem `replace` (see Release State, tag hygiene).
- 23 tests: `middleware_test.go` (9), `setup_test.go` (6), `metrics_test.go` (3, incl. cardinality proof: 3 user IDs → one `/users/{id}` series), `logging_test.go` (5). Race-clean, 5x stable.

## Architecture and Control Flow

- `Service` owns the mux (`svc.Mux`), logger (`svc.Logger`), and HTTP server.
- Consumer registers routes on `svc.Mux`, then calls `svc.Run(ctx)` which blocks.
- appkit does NOT delegate to `httputil.Server`, but the original justification ("no listener access") is STALE since httputil v1.1.x: `Server` now exposes `ListenerAddr() (net.Addr, bool)`, `Start`/`StartTLS` (TLS support appkit lacks!), a double-start guard, and graceful shutdown. Composing `httputil.Server` into `Service` is the recommended refactor (2026-09-16 deep-dive audit; unlocks the deferred Core TLS option).
- httputil is used for: middleware (`Chain`, `Recovery`, `Logging`, etc.), health (`RegisterHealth`), and types (`Middleware`).
- `ServiceConfig.RegisterHealth` is `*bool`: nil or `&true` = register health, `&false` = opt out.
- Graceful drain: `Shutdown()` flips `readyProbe` to false → runs `DrainHooks` (external readiness signals flip in lockstep; errors joined, classified Infrastructure) → waits `DrainDelay` → `server.Shutdown(ctx)` → runs `ShutdownHooks` (once, in order; failures don't stop the rest; errors joined + classified Infrastructure).
- `Run()` uses `signal.NotifyContext` for SIGINT/SIGTERM internally.
- Errors use go-error-family constructors (`NewRejection`, `WrapInfrastructuref`) instead of `fmt.Errorf`.

## Core Dependencies

| Module                                   | Version | Role                                                 |
| ---------------------------------------- | ------- | ---------------------------------------------------- |
| `github.com/larsartmann/httputil`        | v1.1.1  | Middleware, health endpoints, Middleware type        |
| `github.com/charmbracelet/log`           | v1.0.0  | Pretty slog handler (Logger implements slog.Handler) |
| `github.com/larsartmann/go-error-family` | v0.10.1 | Error classification, HTTPStatus, LogError           |

## Realtime Module Dependencies

| Module                                   | Version | Role                                                   |
| ---------------------------------------- | ------- | ------------------------------------------------------ |
| `github.com/larsartmann/go-sse`          | v0.6.0  | SSE transport: Stream, Broadcaster, EventStore, Replay |
| `github.com/larsartmann/go-error-family` | v0.10.1 | Error classification (shared with core)                |
| `github.com/larsartmann/go-branded-id`   | v0.5.1  | Phantom-typed EventID (transitive via go-sse)          |

## Flightrecorder Module Dependencies

| Module                                     | Version | Role                                                   |
| ------------------------------------------ | ------- | ------------------------------------------------------ |
| `github.com/larsartmann/go-flightrecorder` | v0.2.0  | Flight recorder core: Recorder, triggers, typed errors |
| `github.com/larsartmann/httputil`          | v1.1.1  | Middleware type, ResponseRecorder for status capture   |

## otel Module Dependencies

| Module                                           | Version | Role                                           |
| ------------------------------------------------ | ------- | ---------------------------------------------- |
| `go.opentelemetry.io/contrib/.../otelhttp`       | v0.71.0 | Server spans, semconv metrics, W3C propagation |
| `go.opentelemetry.io/otel` (+sdk, metric, trace) | v1.46.0 | Tracer/meter providers, SDK, stdout exporter   |
| `github.com/larsartmann/httputil`                | v1.1.1  | `Middleware` type (bridge target)              |

## cqrs Module Dependencies

| Module                                                           | Version | Role                                                                                     |
| ---------------------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------- |
| `github.com/larsartmann/go-cqrs-lite/system/v4`                  | v4.7.0  | System builder (domain config, decider/command/query registration, host)                 |
| `github.com/larsartmann/go-cqrs-lite/command/v4`                 | v4.10.0 | Command types, dispatcher                                                                |
| `github.com/larsartmann/go-cqrs-lite/query/v4`                   | v4.8.0  | Query types, dispatcher                                                                  |
| `github.com/larsartmann/go-cqrs-lite/decider/v4`                 | v4.6.0  | Decider type                                                                             |
| `github.com/larsartmann/go-cqrs-lite/middleware/v4`              | v4.6.0  | Command middleware (Recovery/Tracing/Logging)                                            |
| `github.com/larsartmann/go-cqrs-lite/otel/v4`                    | v4.4.0  | `cqrsotel.Tracer` for command tracing middleware                                         |
| `github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4` | v4.3.0  | SQLite meta engine (blank import for driver registration)                                |
| `github.com/larsartmann/go-cqrs-lite/projectionhost/v4`          | v4.4.0  | Projection host (DLQ, logger, FR, metrics, lag, readiness)                               |
| `github.com/larsartmann/go-cqrs-lite/event/v4`                   | v4.11.0 | Event types, stream refs, event construction                                             |
| `github.com/larsartmann/go-cqrs-lite/id/v4`                      | v4.6.0  | Branded IDs (stream, event)                                                              |
| `github.com/larsartmann/go-cqrs-lite/projection/v4`              | v4.3.0  | Projection type and `NewProjection`                                                      |
| `github.com/larsartmann/go-cqrs-lite/storage/v4`                 | v4.9.0  | SQLite checkpoint store + schema                                                         |
| `github.com/larsartmann/go-flightrecorder`                       | v0.2.0  | Flight recorder (projectionhost v4.4.0 unified on it; shared with appkit/flightrecorder) |
| `github.com/larsartmann/go-error-family`                         | v0.10.1 | Error classification (shared with core)                                                  |

All pinned cqrs-lite subpackage versions match the latest tags (verified 2026-09-16 against the local checkout).

- Migrated to go-cqrs-lite v4 on 2026-08-15, then onto the `system` engine in v0.5.0 (2026-09-07). Migration guide: go-cqrs-lite `docs/migration/MIGRATION-GUIDE.md`.
- **GOEXPERIMENT=jsonv2 required** (codec/v4 → encoding/json/jsontext).
- v4 codec default flipped JSON→CBOR for new writes; old JSON data still reads (self-describing events). SSE consumers of raw event payloads need CBOR→JSON transcoding.
- Storage posture (v0.5.0): `EventConfig.DSN`/`Driver`/`Pragmas` replace the old stack/sqlite preset (`SQLitePath` still works as a deprecated alias, removed at v0.6.0; `StackOptions` are GONE — use `Pragmas`). SQLite defaults: WAL + busy_timeout. Operator config: `ConfigPath` (koanf YAML + `CQRS_` env overrides) and `Deployment` (precedence: Deployment > ConfigPath > DSN/Driver/Pragmas).
- Projection readiness: `EventService.ReadyCheck()` + `EventService.LagPerProjection()`; core `ServiceConfig.ReadyCheck func() bool` composes external checks with the drain probe for `/health/ready`. v0.5.0 semantics: `ReadyCheck` reports NOT-ready before `StartProjections` (safer default).
- Read-your-writes: projections are async — `EventService.CheckStaleness(budget)` / `CheckProjectionStaleness(name, budget)` are read-time guards (Transient error on lag > budget). projectionhost v4.3.0 has NO public Sync/Drain; cqrs-lint E014's suggested APIs don't exist in the pinned version.
- cqrs-lint: run from inside `cqrs/` — from a workspace root it attributes sub-module imports to the root go.mod (A018 false positive). A/P-series findings on this wrapper (no Save/Publish calls, no WithBatchSize) are by design: consumers get `Bundle()`, tuning flows via `HostOptions`. `.cqrs-lint.json` uses `library-framework` preset (disables ALL F-series adoption-coaching rules) with pinned feature profile and 3 config-level disables (A018, V003, V006). `docs/.cqrs-lint.json` disables A018/A009 (docs module, not an event-sourced app). cqrs-lint source lives at `go-cqrs-lite/cmd/cqrs-lint` — file linter bugs there.
- Command/query facade (v0.5.0): `RegisterDecider`/`RegisterCommand`/`RegisterQuery`, `Dispatch`, `DispatchQuery`, `DispatchQueryChecked` (staleness-gated answering), `CommandDispatcher()`/`QueryDispatcher()` accessors; `System()` exposes the full go-cqrs-lite surface (`Bundle()` is GONE). `DefaultCommandMiddleware(logger, tracer)` = recovery + optional OTel tracing + logging; in-flight command drain runs on `Shutdown`.
- `EventConfig` full option set (v0.5.0): `DSN`/`Driver`/`Pragmas`/`ConfigPath`/`Deployment`, `CheckpointStore` (default: persistent SQLite store), `Logger`, `DLQ *DLQConfig` (SQLite store by default, `ReplayDeadLetters`/`ResetProjection` accessors; `Store` REQUIRED for non-sqlite drivers), `FlightRecorder *fr.Recorder` (same type as appkit/flightrecorder middleware — one shared instance; still one ACTIVE recorder per process), `FlightRecorderTrigger fr.TriggerFunc` (nil = OnAlways; Kind "projection" contexts), `Metrics projectionhost.MetricsRecorder` (backend-agnostic; wire `cqrs.NewOTelProjectionMetrics(meter)` from `otelmetrics.go` for OTel), `CommandMiddleware`/`QueryMiddleware`, `HostOptions` passthrough (`WithCheckpointEvery(n)`, `WithOnFailed(fn)`, `WithMaxRestarts`, `WithBatchSize`; derived wiring wins conflicts).
- cqrs README ends with a cookbook: scenario DSL decider/projection tests (scenario/v4), testutil helpers (CapturingSlogHandler, DelayedJournal), and cqrs-lint usage — all verified against scenario/v4 v4.2.0, testutil/v4 v4.2.0, cqrs-lint 4.6.0.

## Errorpages Module — Code Organization

| File              | Concern                                                                                                                                   |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`          | Package doc.                                                                                                                              |
| `errorpages.go`   | `Config`, `Mount`, `Wrap`, `Handler`, `Write` bridging templ-components/errorpage to go-error-family classification + Accept negotiation. |
| `example/main.go` | Demo service with pretty 404s and a classified error route.                                                                               |

- Depends on `templ-components/errorpage v1.17.0`, `go-error-family v0.10.1`, `go-appkit v0.4.0` (example only; replace `../` for local dev).
- Family → status identical to `appkit.HTTPStatus` (Rejection 400, Conflict 409, Transient 503, Corruption 500, Infrastructure 503).
- `Wrap` mirrors net/http's `cleanPath` to preserve the mux's path-cleaning redirects (doubled slashes, dot segments); only the canonical request that follows gets the pretty 404.
- Render failures fall back to a plain-text response with the correct status (inherited from errorpage's buffer-before-write rendering).

## Testing

- Standard `testing` package, no external frameworks; tests run with `t.Parallel()`.
- Server tests use `freePort()` and `waitForRunning()` helpers (no `time.Sleep`).
- All tests pass with `-race` flag and `-count=1`.
- Default `DrainDelay` (5s) makes tests slow; use `DrainDelay: NoDrainDelay` in non-drain tests — `DrainDelay: 0` applies the 5s default (zero-value production safety), it does NOT disable the wait. Converting the suite dropped wall time ~30s → ~6s.

## Gotchas

- `NewService` registers `/health`, `/health/live`, `/health/ready` by default.
- `Service.Addr()` returns `nil` before `Start()` is called.
- `RegisterHealth` is `*bool` — use `&false` to opt out, not `false`.
- `InitLogger` returns errors (not panics) for invalid config.
- charmbracelet/log `Logger` directly implements `slog.Handler` — no adapter needed.
- `Service.Shutdown` is idempotent — safe to call multiple times.
- **Shutdown phase logging** (2026-09-04): each phase emits `shutdown phase complete` with `phase` ∈ {`ready_flip`, `drain_hooks`, `drain_wait`, `listener_close`, `shutdown_hooks`} + `duration`; `NoDrainDelay` logs `shutdown phase skipped` instead of the wait phase; a final `graceful shutdown complete` line carries `total` + `result` (`ok`/`error`). The names are a grep-able contract — `shutdownlog_test.go` pins the sequence; don't rename without a migration note. A never-started service logs nothing.
- `NoTimeout` sentinel (`-1`): `ReadTimeout`/`WriteTimeout` = `NoTimeout` disables the deadline (server field 0 AND drops the Timeout middleware from the default stack). Required for SSE services; `ReadHeaderTimeout`/`IdleTimeout` reaping stays on. Only `-1` exactly — other negatives are rejected by Validate.
- `NoDrainDelay` sentinel (`-2`): skips the drain wait in Shutdown (ready probe still flips immediately). Only `-2` exactly — other negatives rejected by Validate. Sentinel registry so far: `NoTimeout` = -1, `NoDrainDelay` = -2; the next sentinel takes -3.
- BuildFlow auto-fixes lint on commit (gofumpt, golines, gci).
- Doc snippets are code: compile-check README/doc.go examples in a scratch module before shipping (the health README check caught two real bugs).

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
- Journal-backed replay (CQRS event stores): wire `transport.NewJournalSSEStore(journal, mapper)` from `github.com/larsartmann/cqrs-htmx/v4/transport` (lean 4-dep sub-package) into `realtime.NewHub(realtime.WithStore(store))`. realtime itself stays cqrs-free. **End-to-end verified 2026-09-04** by `integration/integration_test.go` (`TestJournalBackedReplayThroughAppkitService`): cold-start connections get NO history replay (replay is a reconnect mechanism — handler.go returns early on zero Last-Event-ID), a Last-Event-ID reconnect replays exactly the missed journal suffix, live broadcasts interleave; all through an appkit `Service` default stack.
- The `PatchLike` interface intentionally matches `datastar.Patch` — duck typing avoids the import.
- No go-appkit core dependency — `realtime.Mount` works on any `*http.ServeMux`.

## Integration Module — Code Organization

| File                  | Concern                                                                                                                   |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `doc.go`              | Package doc: cross-module + cross-repo E2E compositions; never released.                                                  |
| `integration_test.go` | 2 E2E tests: SSE header flush through the appkit default stack; journal replay via cqrs-htmx `transport.JournalSSEStore`. |
| `.golangci.yml`       | Cloned from realtime's module config (go 1.26.7).                                                                         |

- **Pins PUBLISHED tags only** (`go-appkit v0.4.0`, `realtime v0.1.0`, `cqrs-htmx v4.9.0`, `go-sse v0.6.0` + `go-sse/ssetest`, `go-appkit/otel v0.1.1` → `httputil v1.2.0` + OTel SDK v1.46.0) — it always tests exactly what a fresh consumer resolves from the proxy. Post-v0.4.0 APIs are intentionally unavailable until re-pinned; tests use a 1ms explicit `DrainDelay` (or `NoDrainDelay`, available at the v0.4.0 pin).
- `TestSSEHeadersFlushThroughAppkitDefaultStack` ports the cqrs-htmx ADR-001 spike's M18.3 flush test: headers must arrive well before the first event through appkit's full default middleware stack (Recovery → RequestID → Logging → SecurityHeaders).
- `TestJournalBackedReplayThroughAppkitService` pins the cross-repo contract: no cold-start replay (zero Last-Event-ID), exact missed-suffix replay on reconnect, live broadcast interleave.
- Does NOT require `GOEXPERIMENT=jsonv2` — `cqrs-htmx/transport` only pulls event/id + go-sse + go-error-family (verified 2026-09-04, plain and jsonv2 runs both green 5×).
- Added to `go.work`; add `./integration` to any future workspace listings.

## otel Module Gotchas

- **One `Setup` per process** across go-appkit/otel AND go-cqrs-lite's otel module — both register globals; pick one owner for the tracer/meter providers.
- **Span names + `http.route` metrics through `OuterMiddlewares` — FIXED AND SHIPPED 2026-09-16 (otel v0.1.1 + httputil v1.2.0).** The 2026-09-15 regression (spans flushed as bare `GET`, no `http.route` through the documented `OuterMiddlewares` wiring; root cause: pure-httputil forking middlewares shadowed ServeMux's `r.Pattern` — all FIVE sites: `RequestID`, `Timeout`, `ClientIPMiddleware`, `Nonce`, `ServerTimingMiddlewareWhen`) ships FIXED in httputil v1.2.0 (upstream `TestPatternPropagation*`, zero added allocations); otel v0.1.1 carries the bump, and the otel README known-issue block is deleted. Regression-pinned for consumers by `integration/otel_pattern_test.go` (`TestSpanNameAndRouteThroughAppkitOuterMiddlewares`) against PUBLISHED tags — it fails on any pin older than the train tags, by design. Standing contract: any httputil middleware that forks with `r.WithContext` must propagate `r2.Pattern` back. Known uncovered: CSRF (forks behind justinas/nosurf, which forks internally) — not in the default stack. Note: httputil v1.2 also changes Compression's default for requests without `Accept-Encoding` (appkit's default stack doesn't enable Compression, so no appkit impact). Full bisection: `docs/status/2026-09-15_19-48_otel-telemetry-status.html`.
- **Benchmark drift (2026-09-15/16):** the otel README table now carries the 2026-09-16 n=10 re-baseline (no-op 20.0µs ± 0.4 / traced 21.8µs ± 0.9 / traced+metered 23.3µs ± 2.0). The 2026-09-15 "~25% higher" readings were machine load — the same box measured ~25% LOWER one day later; run-to-run drift here is ±25%, so treat single-run deltas under that as noise and don't chase phantom regressions.
- **go-cqrs-lite's otel `Provider.Shutdown` ForceFlush bug is FIXED upstream** (verified 2026-09-15: their setup.go flushes tracer AND meter before Shutdown, pinned by `TestProvider_Shutdown_FlushesBeforeShutdown`) — older session notes calling it "latent upstream" are obsolete.

## Health Module Gotchas

- **`GOEXPERIMENT=jsonv2` required** to build and test (go-health handlers + go-sse). Always prefix commands with it; `GOWORK=off` for hermetic runs.
- **WithDashboard requires `RegisterHealth: &false`** — the dashboard owns `/health`; a forgotten opt-out panics in `RegisterRoutes` (duplicate pattern), not silently.
- **No `GET /` method-qualified catch-all alongside the dashboard** — the dashboard registers method-agnostic `/health`; Go's ServeMux precedence panics on the pair. Register the root handler without a method (the example documents this inline).
- **`NewProbe` bypasses the `HealthRecorder` path** — go-health ignores `WithHealthRecorder` for `NewWithHealthCheck` probes (documented in the SDK), so `flightrecorderhealth.Trigger` needs an injector-built `health.New` probe; do not promise trigger capture via `NewProbe`.
- **Dashboard `Start` is not idempotent** (spawns a new pusher each call); `Mounted.Start` guards it — call `Mounted` methods, not the dashboard's, for lifecycle.
- The pusher reads `CachedResponse` per tick; a probe that was never `Start`ed serves a zero-value response — `Mounted.Start` runs the initial batch synchronously, so use it before serving.
- `Provider.Shutdown` ForceFlushes both providers BEFORE shutting down: plain Shutdown does not drain the batch processor's async queue and can silently drop final spans (go-cqrs-lite's otel module had this same bug — fixed upstream, verified 2026-09-15).
- Span name carries the method prefix (`GET /users/{id}`); the `http.route` metric attribute does NOT (`/users/{id}`). Assert accordingly.
- `tracetest.InMemoryExporter.Shutdown` RESETS its buffer — read spans after an explicit `ForceFlush`, BEFORE calling Shutdown.
- Health endpoints (`/health`, `/health/live`, `/health/ready`) are unconditionally filtered from tracing and metrics.
- `TraceHandler` correlates only handler-level logs; httputil's `Logging` middleware emits the request-completion line without request context, so that line stays uncorrelated (known limitation, documented in `doc.go`).
- Build with `GOWORK=off` when hermetically testing: the workspace otherwise resolves core via the example's `replace ../`.

## Release Ritual (added 2026-09-04)

1. **API-break check before every tag:** `git archive <old-tag> | tar -x -C /tmp/old && GOWORK=off GOEXPERIMENT=jsonv2 go doc -all . > /tmp/new.txt` (old from the extracted dir, new from the working tree), then diff — additions only → minor bump; ANY removal or signature change → breaking (0.x: minor bump + migration notes in CHANGELOG). Proven during wave 2 (core v0.3.0 → v0.4.0). `goapidiff`/`apidiff` not installed; `go install @latest` is network-blocked in this environment.
2. Date the module's CHANGELOG `[Unreleased]` → `[version] - <date>`.
3. Hermetic verify the module (`GOWORK=off`, jsonv2 only where the toolchain still needs it), fresh-consumer proxy test after push (clean /tmp module → `go get module@tag` → blank import → build).
4. Annotated tags only, message states the semantic delta.

## Adoption & Drift Rituals (added 2026-09-04)

- After adding a cqrs wrapper feature: re-run `cqrs-lint scorecard` (from `cqrs/`, build must be green or the scorecard silently reports 0%) and record the delta in the module CHANGELOG entry.
- After each go-cqrs-lite release: re-verify the cqrs README cookbook snippets (scenario/v4, testutil) still match the shipped APIs.

## Performance Baselines (2026-09-04, go 1.26.7 linux/amd64)

- **Per-request logging cost** (`logging_bench_test.go`, output to io.Discard): bare handler ~17.2µs; Logging-at-WARN (suppressed) ~18.0µs (+0.8µs); Logging-at-INFO (emits) ~47.3µs (+30µs, +174%, 162 allocs). This fully explains the 2.8x cqrs-htmx comparison delta — the cost is charmbracelet formatting of the emitted line, not the middleware. The production default-level decision stays USER-GATED (options: default WARN, sampling, consumer-provided logger).
- **otel middleware** (`otel/benchmark_test.go`): no-op ~21µs / traced ~26µs / traced+metered ~27µs per request; export I/O excluded. Numbers live in otel README.

## Deferred Register (deliberately NOT built — trigger conditions only)

| Item                                                                                  | Trigger to revisit                                                                                                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| cqrs EventConfig opt-ins: encryption/v4, signing/v4, idempotency/sqlstore, scheduling | A real consumer asks for it                                                                                                                                                                                                                                                                                                                                                                            |
| Core TLS option                                                                       | PapDashboard `PAP_TLS_CERT/KEY` becomes an appkit-hosting requirement                                                                                                                                                                                                                                                                                                                                  |
| Cordis bridge module                                                                  | cordis tags go/v0.1.0 AND a consumer states the reactive-composition requirement AND core v1.0.0 exit criteria shipped (`docs/planning/core-v1-exit-criteria.md`)                                                                                                                                                                                                                                      |
| PapDashboard appkit-side code                                                         | They ship an appkit-hosted release (reverse adoption — their repo, not ours)                                                                                                                                                                                                                                                                                                                           |
| BuildFlow dprint exit-14 on CHANGELOG-only commits                                    | Fix belongs upstream in buildflow: skip the dprint step (or pass `--allow-no-files`) when the staged file set ∩ dprint's non-excluded set is empty. dprint.json excludes `**/CHANGELOG.md`; removing the exclude was rejected because the resulting formatting churn cannot be dry-run verified here (no standalone dprint binary). Escape hatch until then: `git commit --no-verify` + justification. |
| httputil `docs/integrations/huma.md` 404 on GitHub                                    | File exists unpushed in the httputil repo — push it, then optionally re-link from `docs/planning/design-decisions.md` Decision 7 (currently inlined instead)                                                                                                                                                                                                                                           |
| cqrs-lint yaml `exclude_patterns`                                                     | Inert in installed binary f7e33e03 despite source comment "removes rules by name" — file a buildflow-style fix in go-structure-linter or wait for a newer binary                                                                                                                                                                                                                                       |
