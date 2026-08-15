# Integrations — How the Ecosystem Fits Together

> **Status:** Verified against source. **Date:** 2026-07-06.
> All APIs verified by reading source code, not summaries.

---

## httputil (v0.4.0) — HTTP Server, Middleware, Health

**Module:** `github.com/larsartmann/httputil`
**Dep dependency:** `go-error-family v0.6.1` (zero transitive deps)

### What appkit delegates to httputil

| Concern          | httputil API                                                         | appkit wrapper                                          |
| ---------------- | -------------------------------------------------------------------- | ------------------------------------------------------- |
| HTTP server      | `NewServer(cfg ServerConfig, handler http.Handler) (*Server, error)` | appkit owns net.Listener separately for richer `Addr()` |
| Server start     | `Server.Start() <-chan error` (non-blocking)                         | appkit bridges to blocking `Start(ctx) error`           |
| Server shutdown  | `Server.Shutdown(ctx) error`                                         | Delegated directly                                      |
| Health endpoints | `RegisterHealth(mux *http.ServeMux)` — registers 3 routes            | Called in `NewService`                                  |
| Ready probe      | `ReadyHandlerWithProbe(ready func() bool)`                           | Used for graceful drain                                 |
| Recovery         | `Recovery(logger *slog.Logger) Middleware`                           | In default middleware stack                             |
| Request ID       | `RequestID(cfg RequestIDConfig) Middleware`                          | In default middleware stack                             |
| Request logging  | `Logging(logger *slog.Logger) Middleware`                            | In default middleware stack                             |
| Timeout          | `Timeout(duration time.Duration) Middleware`                         | In default middleware stack                             |
| Security headers | `SecurityHeaders(cfg SecurityHeadersConfig) Middleware`              | In default middleware stack                             |
| Middleware chain | `Chain(handler http.Handler, mws ...Middleware) http.Handler`        | Used in `NewService`                                    |
| HTTP spec tests  | `httpspec.Run(t, handler, opts...)`                                  | Used in tests (18 free specs)                           |

### Key correction from v1 plan

httputil's `Server.Addr()` returns `string` (configured address), NOT `net.Addr` (live listener address). appkit owns the `net.Listener` separately to preserve `Addr() net.Addr` — critical for tests using `:0` (random port).

### httputil.Server consumers — corrected 2026-08-15

The 2026-07-06 assumption that "cqrs-htmx uses only `httputil.ClientIP`" is outdated: since v4.x cqrs-htmx's `setup.Run` builds its server on `httputil.NewServer` (security-headers + nonce + recovery chain, `/health` with projection-aware readiness). appkit does **not** use `httputil.Server` at all — it owns `http.Server` + `net.Listener` directly to preserve `Addr() net.Addr`.

**Decided relationship (ADR-001, see [design-decisions.md §11](./design-decisions.md)):** cqrs-htmx may adopt `appkit.Service` as its generic server layer (replacing its direct `httputil.Server` use), gated by the M18 prototype spike. appkit is the generic layer; cqrs-htmx keeps its domain middleware and projection-aware readiness.

---

## charmbracelet/log (v1.0.0) — Pretty slog Handler

**Module:** `github.com/charmbracelet/log`

### What it replaces

appkit's current `logger.go` (108 lines): custom slog handler with manual `isTerminal()` detection, format switching, level mapping.

### What appkit does instead (~5 lines)

```go
func InitLogger(cfg LoggerConfig) (*slog.Logger, error) {
    level := charmlog.Level(cfg.Level.slogLevel())
    cl := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
        Level:  level,
        Format: charmlog.TextFormatter, // or JSONFormatter
    })
    return slog.New(cl.NewHandler()), nil
}
```

### Why charmbracelet/log

- Pretty colored TTY output by default (aligned key=value pairs, colored levels).
- JSON mode for production (`logger.SetFormatter(charmlog.JSONFormatter)`).
- Level control via `logger.SetLevel(charmlog.DebugLevel)`.
- Native `slog.Handler` bridge via `(*charmlog.Logger).NewHandler()`.
- Used by 4 other ecosystem projects (ast-state-analyzer, auto-deduplicate, github-local-sync, go-plugin-mvp).
- cmdguard has `LogLevel.SlogLevel()` for the same slog conversion pattern.

### Namespace note

`github.com/charmbracelet/log` v1.0.0 is the canonical module. `charm.land/*` v2 packages (fang, huh, glamour, lipgloss, bubbles) are rebranded forks of OTHER charm tools. `charmbracelet/log` has not been rebranded. No conflict — different packages, same org (Charm).

---

## go-error-family (v0.6.1) — Error Classification

**Module:** `github.com/larsartmann/go-error-family` (root, zero deps)
**Bridge module:** `github.com/larsartmann/go-error-family/bridge` (adds samber/oops)

### Proper 3-layer adoption

See [design-decisions.md §6](./design-decisions.md#decision-6-go-error-family--proper-3-layer-adoption).

**Layer 1 — appkit's own errors (constructors, not sentinels):**

| Old (sentinel)                                | New (constructor)                                                                     |
| --------------------------------------------- | ------------------------------------------------------------------------------------- |
| `errSQLitePathRequired = errors.New("...")`   | `errorfamily.NewRejection("sqlite.path_required", "sqlite path is required")`         |
| `errPRAGMAAllowlist = errors.New("...")`      | `errorfamily.NewRejection("sqlite.pragma_disallowed", "unsupported PRAGMA: %s", key)` |
| `errUnsupportedLogLevel = errors.New("...")`  | `errorfamily.NewRejection("log.level_invalid", "unsupported log level: %s", l)`       |
| `errUnsupportedLogFormat = errors.New("...")` | `errorfamily.NewRejection("log.format_invalid", "unsupported log format: %s", f)`     |
| `fmt.Errorf("listen on %s: %w", addr, err)`   | `errorfamily.WrapInfrastructure(err, "server.listen_failed", "listen on %s", addr)`   |
| `fmt.Errorf("serve: %w", err)`                | `errorfamily.WrapInfrastructure(err, "server.serve_failed", "serve failed")`          |
| `fmt.Errorf("server shutdown: %w", err)`      | `errorfamily.WrapInfrastructure(err, "server.shutdown_failed", "shutdown failed")`    |

**Layer 2 — boundary terminators:**

| Boundary            | Terminator                                  | Usage in appkit                          |
| ------------------- | ------------------------------------------- | ---------------------------------------- |
| HTTP error response | `errorfamily.HTTPHandler(func(w, r) error)` | Documented pattern for consumer handlers |
| Error logging       | `errorfamily.LogError(err, logger)`         | In `Shutdown()` error paths              |
| CLI exit            | `errorfamily.HandleError(err) int`          | In `example/main.go`                     |
| HTTP status         | `errorfamily.HTTPStatus(err) int`           | Available as `appkit.HTTPStatus(err)`    |

**Layer 3 — bridge/oops (application-only, NOT in appkit):**

Documented as a consumer pattern:

```go
// In the application's main package:
import errorfamilybridge "github.com/larsartmann/go-error-family/bridge"
import "github.com/samber/oops"

rich := oops.In("database").Tags("timeout").With("host", dbHost).Wrap(dbErr)
classified := errorfamilybridge.AutoWrap(rich) // infers Transient from domain "database"
```

appkit does NOT import the bridge. It stays zero-dep on oops.

### The 5-family taxonomy

| Family         | Retryable | HTTP | Exit | When                                            |
| -------------- | --------- | ---- | ---- | ----------------------------------------------- |
| Rejection      | no        | 400  | 1    | User error (bad input, not found, unauthorized) |
| Conflict       | no        | 409  | 1    | Version mismatch, duplicate                     |
| Transient      | **yes**   | 503  | 75   | Timeout, busy, network blip                     |
| Corruption     | no        | 500  | 65   | Data corrupt, schema broken                     |
| Infrastructure | no        | 503  | 69   | DB down, port in use, startup failure           |

---

## go-cqrs-lite/stack/sqlite (v3.6.0) — CQRS/ES Engine

**Module:** `github.com/larsartmann/go-cqrs-lite/stack/sqlite/v3`
**Used by:** `go-appkit/cqrs` (opt-in sub-module)

### What stack/sqlite.New provides

```go
bundle, err := sqlite.New("app.db", opts...)
```

One call gives you:

| Capability            | Bundle field                       | Type                                   |
| --------------------- | ---------------------------------- | -------------------------------------- |
| Write events          | `Bundle.EventSink`                 | `event.EventSink`                      |
| Read events           | `Bundle.EventSource`               | `event.EventSource`                    |
| Full event journal    | `Bundle.Journal`                   | `event.Journal`                        |
| Write commands        | `Bundle.CommandSink`               | `command.CommandSink`                  |
| Read commands         | `Bundle.CommandSource`             | `command.CommandSource`                |
| Write queries (audit) | `Bundle.QuerySink`                 | `query.QuerySink`                      |
| Snapshot store        | `Bundle.SnapshotStore`             | `snapshot.SnapshotStore`               |
| Checkpoint store      | `Bundle.CheckpointStore`           | `event.CheckpointStore`                |
| Read models (KV)      | `Bundle.ReadModels`                | `kv.Store`                             |
| In-process bus        | `Bundle.Publisher` / `.Subscriber` | `event.Publisher` / `event.Subscriber` |
| Raw DB                | `Bundle.Database()`                | `any` → type-assert `*sql.DB`          |

**All auto-migrated** (tables created on startup). **WAL mode** by default. **busy_timeout=5000**. Options: `WithoutWAL()`, `WithOptimizations()`, `WithForeignKeys()`, `WithoutAutoMigrate()`, `WithEventDB(dsn)`, `WithQueryDB(dsn)`, `WithViewDB(dsn)`.

### Repository builder

```go
repo, err := stack.Repository[MyState](bundle, decider, opts...)
err = repo.Execute(ctx, aggID, cmdType, decideFunc)
state, err := repo.Load(ctx, aggID)
```

Event-sourcing aggregate loader. Uses singleflight for load coalescing. Supports snapshots.

### Projection host (separate import)

```go
import "github.com/larsartmann/go-cqrs-lite/projectionhost/v3"

host := projectionhost.New(bundle.Subscriber, bundle.CheckpointStore, opts...)
host.Register("user-projection", myProjection, projectionhost.WithDLQ(dlqStore))
host.Start() // managed lifecycle, crash recovery, backoff
```

### What appkit/cqrs wraps

appkit's `EventService` wraps `stack/sqlite.New` and adds:

- Integration with `Service.Run()` for graceful shutdown (`bundle.GracefulClose(ctx)` after server stops).
- Projection host lifecycle management.
- DB accessor (`svc.DB()` type-asserts `Bundle.Database()`).

---

## Huma (v2) — Type-Safe HTTP API Generation

**Module:** `github.com/danielgtaylor/huma/v2` + `github.com/danielgtaylor/huma/v2/adapters/humago`
**Status:** Documented pattern, NOT a dependency.

### How Huma and appkit coexist

Huma wraps `svc.Mux` INSIDE. httputil middleware (wired by appkit) wraps OUTSIDE. The layers never collide — [httputil already documented this](https://github.com/LarsArtmann/httputil/blob/main/docs/integrations/huma.md).

```
Client → httputil.Chain (Recovery → ReqID → Logging → Timeout → Sec)
       → svc.Mux (*http.ServeMux)
       → humago adapter → huma operation handler (typed input → validation → OpenAPI 3.1)
       → OR: plain http.HandleFunc (no Huma needed)
```

### Consumer code (opt-in)

```go
// Standard mux path (default — no Huma):
svc.Mux.HandleFunc("GET /users/{id}", handler)

// Huma path (consumer imports huma themselves):
api := humago.New(svc.Mux, huma.DefaultConfig("My API", "1.0.0"))
huma.Get(api, "/users/{id}", func(ctx context.Context, input *struct {
    ID string `path:"id" maxLength:"30"`
}) (*UserOutput, error) {
    return &UserOutput{Body: struct{ Name string `json:"name"` }{"Alice"}}, nil
})
```

Huma generates OpenAPI 3.1 + JSON Schema from Go types. Serves `/docs` with Scalar UI. No conflict with appkit's middleware or health endpoints.

---

## go-cqrs-lite/catalog (v3.6.0) — Auto-Documentation

**Module:** `github.com/larsartmann/go-cqrs-lite/catalog/v3`
**Used by:** `go-appkit/docs` (opt-in sub-module)

### What catalog provides

```go
builder := catalog.NewBuilder("My Service", "1.0.0")
builder.AddCommand[CreateUserCmd]("create-user", catalog.MsgOperation("POST", "/users"))
builder.AddEvent[UserCreatedEvent]("user-created", catalog.Sends)
builder.AddQuery[GetUserQry]("get-user", catalog.MsgOperation("GET", "/users/{id}"))

cat := builder.Build()
```

`SchemaFromType[T]()` reflects on Go types and generates JSON Schema from struct tags (`json`, `doc`, `format`, `enum`, `nullable`, `default`, `pattern`, `deprecated`).

### Exporters

| Exporter               | Output           | Format            |
| ---------------------- | ---------------- | ----------------- |
| `catalog/openapi`      | OpenAPI 3.0.3    | JSON/YAML         |
| `catalog/asyncapi`     | AsyncAPI 3.0.0   | JSON/YAML         |
| `catalog/d2`           | D2 diagram       | DSL string        |
| `catalog/eventcatalog` | EventCatalog MDX | File tree on disk |

### docserver — serves documentation UI

```go
import "github.com/larsartmann/go-cqrs-lite/catalog/docserver/v3"

ds := docserver.NewDocsServer(func() *catalog.Catalog { return builder.Build() }, docserver.DefaultConfig())
ds.RegisterRoutes(svc.Mux) // adds /docs/openapi, /docs/asyncapi, /docs/diagram
```

Routes:

- `GET /docs/openapi` — Scalar HTML UI
- `GET /docs/openapi.json` / `.yaml`
- `GET /docs/asyncapi` — AsyncAPI React HTML UI
- `GET /docs/asyncapi.json` / `.yaml`
- `GET /docs/catalog.json` — raw catalog JSON

### Huma + catalog together

They're complementary, not competing:

|                      | Huma                 | catalog                                       |
| -------------------- | -------------------- | --------------------------------------------- |
| **Domain**           | HTTP REST APIs       | Events, commands, queries                     |
| **Schema mechanism** | `huma.*` struct tags | `json`/`doc`/`format` struct tags             |
| **Primary output**   | OpenAPI 3.1          | AsyncAPI 3.0, OpenAPI 3.0.3, D2, EventCatalog |
| **UI**               | Scalar / Stoplight   | Scalar / AsyncAPI React                       |

Both reflect on the same Go types independently. No conflict. docserver serves both side-by-side on the same mux.

---

## Competitive Landscape

### Honest differentiation

| Library                    | What it owns                | What appkit adds                                                       |
| -------------------------- | --------------------------- | ---------------------------------------------------------------------- |
| **chi v5**                 | Routing                     | Server lifecycle, middleware defaults, logging, health, CQRS, shutdown |
| **huma v2**                | Type-safe API gen, OpenAPI  | Server lifecycle, logging, CQRS, health, shutdown                      |
| **httputil**               | HTTP server, middleware     | Logger init, CQRS, signal, lifecycle orchestration                     |
| **Gin / Echo**             | HTTP + routing + middleware | CQRS/ES, charmbracelet logging, error classification                   |
| **Buffalo / Fiber / GoFr** | Full-stack framework        | Different stack — appkit composes the larsartmann ecosystem            |

**Honest claim:** appkit is the only framework that composes httputil + go-cqrs-lite + go-error-family + charmbracelet/log into a unified service. It's not "nothing else exists" — Buffalo and GoFr exist. It's "nothing else composes THESE libraries with THIS architecture."

### Target user

> "I use Go stdlib `http.ServeMux` (Go 1.22+ method routing). I want middleware, health checks, pretty logging, error classification, and graceful shutdown without writing 50 lines of boilerplate. If I need CQRS/ES, I want the event store, projections, and bus wired in one import."

Nobody else serves this user. chi gives routing but not lifecycle. huma gives API types but not logging or DB. httputil gives HTTP utilities but not composition. appkit gives a **running service**.
