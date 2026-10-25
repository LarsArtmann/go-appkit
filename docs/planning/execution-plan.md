# Execution Plan — Modular Framework

> **Status:** Active. **Date:** 2026-07-06.
> **Architecture:** See [framework-architecture.md](./framework-architecture.md).
> **Decisions:** See [design-decisions.md](./design-decisions.md).

---

## Module 1: `go-appkit` (Core) — Ship First

**Goal:** Production-ready HTTP service in 12 lines.
**Dependencies:** httputil@v0.4.0, charmbracelet/log@v1.0.0, go-error-family@v0.6.1
**Estimated effort:** ~14h

### Phase 1: Foundation (1% → 51%) — ~3.5h

| # | Task                                                                                                                                                                                                                                                                                                                                      | Impact   | Effort |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ |
| 1 | **Define `ServiceConfig` + `Service` types** in `service.go`. Config: Addr, LogLevel, LogFormat, ShutdownTimeout, DrainDelay, Middlewares, ExtraMiddlewares, RegisterHealth, Server timeouts. Service: logger, inner *httputil.Server, mux *http.ServeMux, ln net.Listener, mu sync.RWMutex, closeOnce sync.Once, readyProbe atomic.Bool. | Critical | 45m    |
| 2 | **Implement `NewService(cfg) (*Service, error)`** — Validate → InitLogger (charmbracelet/log) → create mux → RegisterHealth → build middleware chain → httputil.NewServer. Returns first error.                                                                                                                                           | Critical | 75m    |
| 3 | **Implement `Service.Run(ctx) error`** — drain sequence (flip ready probe → wait DrainDelay → Shutdown). Implement `Start() <-chan error` (non-blocking).                                                                                                                                                                                 | Critical | 60m    |
| 4 | **Accessors** — Mux, Logger as public fields. `Addr() net.Addr` from listener. `Running() bool`. `DB() *sql.DB` (returns nil in core — no DB).                                                                                                                                                                                            | High     | 20m    |

### Phase 2: Composition (4% → 64%) — ~4h

| #  | Task                                                                                                                                                                                                                                | Impact   | Effort |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ |
| 5  | **Add deps:** httputil@v0.4.0, charmbracelet/log@v1.0.0, go-error-family@v0.6.1 + go mod tidy.                                                                                                                                      | Critical | 15m    |
| 6  | **Rewrite `logger.go`** — replace 108 lines with charmbracelet/log: create logger → NewHandler() → slog.New(). Keep LogLevel/LogFormat types + InitLogger signature.                                                                | Critical | 45m    |
| 7  | **Rewrite `server.go`** — delegate to httputil.Server. appkit owns net.Listener for Addr() net.Addr. Start bridges <-chan error to blocking. Delete ServerConfig (use ServiceConfig).                                               | High     | 75m    |
| 8  | **Rewrite `health.go`** — delegate to httputil.RegisterHealth. Map HealthStatus to httputil up/down. DefaultHealthHandler = alias for httputil.HealthHandler().                                                                     | High     | 30m    |
| 9  | **Implement `middleware.go`** — defaultMiddlewareStack(logger, cfg) returns []httputil.Middleware: Recovery → RequestID → Logging → Timeout → SecurityHeaders. Support cfg.Middlewares (replace) and cfg.ExtraMiddlewares (append). | Critical | 60m    |
| 10 | **Wire middleware** — httputil.Chain(mux, stack...) → wrapped handler → httputil.NewServer(cfg, handler).                                                                                                                           | High     | 30m    |

### Phase 3: Production-grade (20% → 80%) — ~2.5h

| #  | Task                                                                                                                                                                                                                                                                                                   | Impact | Effort |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ |
| 11 | **Implement `Service.Shutdown(ctx)`** — flip ready probe → server.Shutdown → logger flush. sync.Once. LogError on failure.                                                                                                                                                                             | High   | 45m    |
| 12 | **Create `errors.go`** — PROPER error-family adoption: replace sentinels with constructors (NewRejection, WrapInfrastructure). RegisterClassifier for \*sqlite.Error → Transient (even though core doesn't use SQLite, the classifier is harmless). Export appkit.HTTPStatus = errorfamily.HTTPStatus. | High   | 60m    |
| 13 | **Implement `Validate()` + `DefaultServiceConfig()`** — addr non-empty, log level valid, timeout > 0. Defaults: Addr=":8080", Info, Auto, 5s drain, 15s shutdown.                                                                                                                                      | Medium | 30m    |

### Phase 4: Ship (100%) — ~4h

| #  | Task                                                                                                                                                                                | Impact | Effort |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| 14 | **`Service.Close()`** — Shutdown(context.Background()) via sync.Once.                                                                                                               | Low    | 15m    |
| 15 | **Rewrite README** — "Production-ready Go service in 12 lines." Before/after. ServiceConfig table. "When NOT to use" → httputil, cqrs-htmx, chi, cmdguard. Huma pattern documented. | High   | 75m    |
| 16 | **`example/main.go`** — 12-line HTTP service with `errorfamily.HandleError(err)` as CLI terminator.                                                                                 | High   | 30m    |
| 17 | **Test: construction** — defaults, custom, invalid. With/without middleware replacement.                                                                                            | Medium | 45m    |
| 18 | **Test: Run + health** — /health, /health/live, /health/ready → 200. ReadyHandlerWithProbe(false) → 503.                                                                            | Medium | 45m    |
| 19 | **Test: shutdown + drain** — cancel ctx, Run returns nil. Verify drain delay works (ready probe flips).                                                                             | Medium | 45m    |
| 20 | **Test: middleware** — panic → 500, X-Request-ID present, logging captured.                                                                                                         | Medium | 45m    |
| 21 | **Test: validation** — table-driven invalid configs. Verify error-family classification.                                                                                            | Low    | 30m    |
| 22 | **Test: httpspec** — `httpspec.Run(t, handler, SkipSpec(IndexNot404))`.                                                                                                             | Medium | 30m    |
| 23 | **Update `doc.go`** — "production-ready service framework composing httputil."                                                                                                      | Low    | 15m    |
| 24 | **Update `AGENTS.md`** — architecture, decisions, file map, gotchas.                                                                                                                | Medium | 30m    |
| 25 | **`flake.nix`** — Go 1.26.4, golangci-lint, test/lint/build apps.                                                                                                                   | Medium | 30m    |
| 26 | **Final green + tag v1.0.0** — go test -race, vet, build, nix lint. Tag.                                                                                                            | High   | 15m    |

**Total core: 26 tasks, ~14h**

---

## Module 2: `go-appkit/cqrs` — CQRS Integration

**Goal:** Production-ready CQRS/ES service in 20 lines.
**Dependencies:** go-cqrs-lite/stack/sqlite@v3.6.0, go-cqrs-lite/projectionhost@v3.6.0, go-appkit (core)
**Estimated effort:** ~8h
**Tag:** v0.1.0 (experimental)

### Tasks

| #  | Task                                                                                                                                                                                                  | Impact   | Effort |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ |
| 1  | **Create `go-appkit/cqrs/go.mod`** — module `github.com/larsartmann/go-appkit/cqrs`. Deps: stack/sqlite, projectionhost, go-appkit.                                                                   | Critical | 15m    |
| 2  | **Define `EventConfig` + `EventService` types** — Config: SQLitePath, StackOptions []sqlite.Option, ProjectionHostConfig. EventService: bundle *stack.Bundle, host *projectionhost.Host, db \*sql.DB. | Critical | 30m    |
| 3  | **Implement `NewEventService(cfg) (*EventService, error)`** — stack/sqlite.New(dsn, opts...). Verify schema migration. Extract \*sql.DB via Database().                                               | Critical | 45m    |
| 4  | **Integrate with Service** — `Service.UseCQRS(es *EventService)` or `ServiceConfig.CQRS *cqrs.EventConfig`. Service.Run() calls es.Shutdown() after server.Shutdown().                                | Critical | 60m    |
| 5  | **Projection host lifecycle** — EventService.RegisterProjection(name, projection, opts...). Start on Service.Start(), stop+drain on Service.Shutdown().                                               | High     | 75m    |
| 6  | **Bundle accessor** — `svc.Bundle *stack.Bundle` (nil if no CQRS). Expose EventSink, EventSource, CommandSink, Publisher, Subscriber.                                                                 | High     | 30m    |
| 7  | **DB accessor** — `svc.DB() *sql.DB` (type-asserts Bundle.Database()).                                                                                                                                | Medium   | 15m    |
| 8  | **Shutdown coordination** — EventService.Shutdown(ctx): projectionhost.Stop() → bundle.GracefulClose(ctx). sync.Once. LogError on failure.                                                            | High     | 45m    |
| 9  | **E2E test** — create EventService → register command → dispatch → verify event stored → read projection → hit /health/ready → shutdown.                                                              | High     | 90m    |
| 10 | **README section** — "CQRS Integration" with 20-line example. Document Bundle access, projection registration, bus extension.                                                                         | Medium   | 45m    |
| 11 | **Tag v0.1.0** — go test -race, vet, build. Tag.                                                                                                                                                      | Medium   | 15m    |

**Total cqrs: 11 tasks, ~8h**

---

## Module 3: `go-appkit/docs` — Auto-Documentation

**Goal:** Self-documenting service with AsyncAPI + OpenAPI + D2 from Go types.
**Dependencies:** go-cqrs-lite/catalog/v3@v3.6.0, go-appkit (core)
**Estimated effort:** ~6h
**Tag:** v0.1.0 (experimental)

### Tasks

| # | Task                                                                                                                                                       | Impact   | Effort |
| - | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ |
| 1 | **Create `go-appkit/docs/go.mod`** — module `github.com/larsartmann/go-appkit/docs`. Deps: catalog, go-appkit.                                             | Critical | 15m    |
| 2 | **Wrap catalog.Builder** — `docs.NewCatalogBuilder(title, version)` returns a wrapped builder with sane defaults.                                          | Critical | 30m    |
| 3 | **Implement `RegisterDocs(mux, builder)`** — mounts catalog/docserver routes: /docs/openapi, /docs/asyncapi, /docs/diagram, /docs/catalog.json.            | Critical | 45m    |
| 4 | **Document the Huma pattern** — README section showing how to wrap svc.Mux with humago.New for HTTP OpenAPI 3.1 alongside catalog's AsyncAPI.              | High     | 30m    |
| 5 | **D2 diagram integration** — `GET /docs/diagram` serves auto-generated D2 architecture diagram from registered services/events.                            | Medium   | 30m    |
| 6 | **EventCatalog export** — `docs.ExportEventCatalog(builder, outputDir)` writes MDX file tree for EventCatalog.org.                                         | Medium   | 45m    |
| 7 | **E2E test** — register command + event → build catalog → hit /docs/openapi.json → verify schema present → hit /docs/asyncapi.json → verify event present. | High     | 60m    |
| 8 | **README section** — "Auto-Documentation" with example. Document Huma opt-in. Show both docs side-by-side.                                                 | Medium   | 45m    |
| 9 | **Tag v0.1.0** — go test -race, vet, build. Tag.                                                                                                           | Medium   | 15m    |

**Total docs: 9 tasks, ~6h**

---

## Summary

| Module               | Tasks  | Effort   | Tag    | Dependencies                              |
| -------------------- | ------ | -------- | ------ | ----------------------------------------- |
| **go-appkit** (core) | 26     | ~14h     | v1.0.0 | httputil, charmbracelet/log, error-family |
| **go-appkit/cqrs**   | 11     | ~8h      | v0.1.0 | + stack/sqlite, projectionhost            |
| **go-appkit/docs**   | 9      | ~6h      | v0.1.0 | + catalog                                 |
| **Total**            | **46** | **~28h** |        |                                           |

### Critical path

**Core Phase 1 (1→2→3→4) → Phase 2 (5→6→7→8→9→10) → Phase 3 (11→12→13) → Phase 4 (15→17→26)** = ~10h critical path for core.

cqrs and docs can start only after core Phase 2 completes (they depend on `svc.Mux` and `Service`). They can be built in parallel with core Phase 3-4 if separate developers.

### Ship order

1. **Core v1.0.0** — ship first. Most services only need this.
2. **cqrs v0.1.0** — ship second. For event-sourced services.
3. **docs v0.1.0** — ship third. For self-documenting services.

Each module is independently useful. Each can ship on its own timeline.
