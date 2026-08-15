# Framework Architecture — Modular Design

> **Status:** Proposed. **Date:** 2026-07-06.
> **Replaces:** The monolithic plan in `2026-07-06_02-53-path2-superb-service-framework.html`.

---

## The Problem with Monolithic

v3 of the plan bundled everything into one module: HTTP server, middleware, logging, CQRS/ES event store, command/query dispatch, projections, auto-documentation, and catalog generation. 27 tasks. 4+ heavy dependencies.

**This is wrong for two reasons:**

1. **90% of Go services don't need CQRS/ES.** Forcing every consumer to pull in the entire `go-cqrs-lite/stack/sqlite` dependency tree (event/v3, command/v3, query/v3, codec, storage/sql, watermill, CBOR, OTel SDK) makes the framework hostile to its largest audience — people who just want "HTTP server with middleware, logging, and graceful shutdown."

2. **It couples unrelated release cycles.** If httputil ships a breaking change, the CQRS integration shouldn't need a new tag. If go-cqrs-lite adds a new storage backend, the HTTP framework shouldn't need a release.

---

## The Modular Framework

Three Go modules in one repo, each independently versioned, each building on the previous:

```
go-appkit                        Core HTTP Framework (MUST ship first)
│   go.mod: httputil + charmbracelet/log + go-error-family
│
├── go-appkit/cqrs               CQRS Integration (opt-in)
│   go.mod: go-cqrs-lite/stack/sqlite + go-appkit (core)
│
└── go-appkit/docs               Auto-Documentation (opt-in)
    go.mod: go-cqrs-lite/catalog + go-appkit (core)
```

### Module 1: `go-appkit` — Core HTTP Framework

**The 80%. Ships first. Most services only need this.**

**Dependencies:** httputil@v0.4.0, charmbracelet/log@v1.0.0, go-error-family@v0.6.1

**What it provides:**

| Concern               | Source                       | API                                                        |
| --------------------- | ---------------------------- | ---------------------------------------------------------- |
| HTTP server lifecycle | httputil.Server (delegated)  | `NewService(cfg) (*Service, error)`                        |
| Middleware defaults   | httputil.Chain               | Recovery → RequestID → Logging → Timeout → SecurityHeaders |
| Health endpoints      | httputil.RegisterHealth      | `/health`, `/health/live`, `/health/ready`                 |
| Structured logging    | charmbracelet/log → slog     | `svc.Logger *slog.Logger`                                  |
| Error classification  | go-error-family constructors | `NewRejection`, `WrapTransient`, `HTTPHandler`, `LogError` |
| Signal-based shutdown | appkit (kept)                | `Run(ctx)` blocks, `Shutdown(ctx)` coordinates             |
| Config validation     | appkit                       | `ServiceConfig.Validate() error`                           |
| Bound address         | appkit owns net.Listener     | `Addr() net.Addr` (richer than httputil's string)          |

**Consumer code (~12 lines):**

```go
func main() {
    svc, err := appkit.NewService(appkit.ServiceConfig{
        Addr:       ":8080",
        LogLevel:   appkit.LogLevelInfo,
    })
    if err != nil { log.Fatal(err) }
    defer svc.Close()

    svc.Mux.HandleFunc("GET /", handler)

    if err := svc.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

**No CQRS. No event store. No projections. No schema generation.** Just a production-ready HTTP service with pretty logging, security middleware, health checks, and graceful shutdown.

### Module 2: `go-appkit/cqrs` — CQRS Integration

**Opt-in for event-sourced services. Builds on core.**

**Dependencies:** go-cqrs-lite/stack/sqlite@v3.6.0, go-appkit (core)

**What it adds:**

| Concern                   | Source                            | API                                                                    |
| ------------------------- | --------------------------------- | ---------------------------------------------------------------------- |
| SQLite-backed event store | stack/sqlite.New                  | `cqrs.NewEventService(cfg) (*EventService, error)`                     |
| CQRS Bundle               | stack.Bundle                      | `svc.Bundle *stack.Bundle` (EventSink, EventSource, CommandSink, etc.) |
| Projection host lifecycle | go-cqrs-lite/projectionhost       | Managed workers, crash recovery, DLQ, backoff                          |
| Bus exposure              | stack.Bundle.Publisher/Subscriber | In-process (GoChannel) by default; extensible                          |
| Repository builder        | stack.Repository[State]           | Event-sourcing aggregate loader                                        |
| Read model store          | stack.ReadModel[T,K]              | Typed KV for projections                                               |
| Auto schema migration     | storage.SQLiteInitSchema          | Tables created on startup                                              |

**Consumer code (~20 lines):**

```go
func main() {
    svc, err := appkit.NewService(appkit.ServiceConfig{
        Addr:       ":8080",
        LogLevel:   appkit.LogLevelInfo,
    })
    if err != nil { log.Fatal(err) }
    defer svc.Close()

    // CQRS opt-in
    es, err := cqrs.NewEventService(cqrs.EventConfig{
        SQLitePath: "app.db",
    })
    if err != nil { log.Fatal(err) }

    repo, _ := stack.Repository[MyState](es.Bundle, myDecider)
    svc.Mux.HandleFunc("POST /commands", makeHandler(repo))

    if err := svc.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

**Key design:** `EventService` wraps `stack/sqlite.New` and integrates with `Service`:

- `Service.Run()` calls `EventService.Shutdown()` during graceful shutdown (after server stops).
- `EventService.Bundle` is the composition point for CQRS handlers.
- Projection host is managed by `EventService` — started/stopped with the service lifecycle.

### Module 3: `go-appkit/docs` — Auto-Documentation

**Opt-in for self-documenting services. Builds on core.**

**Dependencies:** go-cqrs-lite/catalog/v3@v3.6.0, go-appkit (core)

**What it adds:**

| Concern                     | Source                              | API                                                             |
| --------------------------- | ----------------------------------- | --------------------------------------------------------------- |
| Event/command/query catalog | catalog.Builder + SchemaFromType[T] | `docs.NewCatalogBuilder(title, version)`                        |
| AsyncAPI 3.0 export         | catalog/asyncapi                    | `GET /docs/asyncapi` + `/docs/asyncapi.json`                    |
| OpenAPI export              | catalog/openapi                     | `GET /docs/openapi` + `/docs/openapi.json`                      |
| D2 architecture diagram     | catalog/d2                          | `GET /docs/diagram`                                             |
| EventCatalog MDX export     | catalog/eventcatalog                | Write MDX file tree to disk                                     |
| Docserver UI                | catalog/docserver                   | Scalar (OpenAPI), AsyncAPI React (events)                       |
| Huma pattern                | documented, NOT depended on         | Consumer wraps `svc.Mux` with `humago.New` for HTTP OpenAPI 3.1 |

**Consumer code:**

```go
// Register CQRS messages for auto-documentation
builder := docs.NewCatalogBuilder("My Service", "1.0.0")
builder.AddCommand[CreateUserCmd]("create-user")
builder.AddEvent[UserCreatedEvent]("user-created", catalog.Sends)

// Mount docserver on svc.Mux
docs.RegisterDocs(svc.Mux, builder)

// Optionally wrap mux with Huma for HTTP API docs
// api := humago.New(svc.Mux, huma.DefaultConfig("My API", "1.0.0"))
```

**Key design:** Huma is documented as a pattern, NOT a dependency. The consumer imports `huma` themselves if they want type-safe HTTP API generation + OpenAPI 3.1. catalog provides AsyncAPI 3.0 + D2 + EventCatalog from the same Go types. Both layers coexist on the same mux without conflict.

---

## Dependency Direction

```
               ┌──────────────────┐
               │   Application     │
               │   (main package)  │
               └────────┬──────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
┌────────▼──────┐ ┌─────▼─────┐ ┌─────▼──────┐
│ go-appkit/cqrs│ │go-appkit/ │ │ go-appkit  │
│               │ │  docs     │ │   (core)   │
└───────┬───────┘ └─────┬─────┘ └─────┬──────┘
        │               │              │
        │               │              ├── httputil
        │               │              ├── charmbracelet/log
        │               │              └── go-error-family
        │               │
        ├── go-cqrs-lite/stack/sqlite  │
        ├── go-cqrs-lite/projectionhost│
        └── go-appkit (core) ──────────┘
                        │
               ┌────────┴──────────┐
               │  go-cqrs-lite/    │
               │  catalog/v3       │
               └───────────────────┘
```

- Core never imports cqrs or docs.
- cqrs imports core (for Service integration).
- docs imports core (for mux access).
- cqrs and docs are independent of each other (but can be used together).
- Application imports whichever modules it needs.

---

## Versioning Strategy

| Module             | Initial Tag | Stability                                             |
| ------------------ | ----------- | ----------------------------------------------------- |
| `go-appkit` (core) | v1.0.0      | Stable. Service type is the v1 contract.              |
| `go-appkit/cqrs`   | v0.1.0      | Experimental. API may change as CQRS patterns mature. |
| `go-appkit/docs`   | v0.1.0      | Experimental. API may change as catalog evolves.      |

**Rationale:** Core is the 80% — it must be stable. CQRS and docs are opt-in power features — they earn v1.0.0 when their APIs prove stable in real services.

**Pre-v1 dependency risk:** httputil (v0.4.0) and go-error-family (v0.6.1) are pre-v1. appkit core v1.0.0 tracking pre-v1 deps is a documented risk. Mitigation: go-error-family root has zero deps and a stable interface contract; httputil's Server/middleware API has been stable since v0.3.0.
