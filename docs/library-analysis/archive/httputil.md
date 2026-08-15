# httputil — Integration Analysis

> **Verdict: 🟥 ACT.** go-appkit's `server.go` + `health.go` duplicate a strict subset of httputil's
> `Server` + `Health`. This is a real split-brain, not a hypothetical integration. Resolve it by
> composing httputil rather than re-implementing it.

---

## 1. Library identity

| Attribute      | Value                                                                          |
| -------------- | ------------------------------------------------------------------------------ |
| Module path    | `github.com/larsartmann/httputil`                                              |
| Go version     | 1.26.4                                                                         |
| Latest release | **v0.4.0** (2026-07-02)                                                        |
| License        | present (per CI release workflow)                                              |
| Runtime deps   | **1** — `github.com/larsartmann/go-error-family v0.6.1` (zero transitive deps) |
| Subpackage     | `httputil/httpspec` — reusable BDD HTTP spec suite (pure stdlib)               |

A batteries-included HTTP middleware/utility kit for `net/http`. Stdlib-first, framework-agnostic —
every middleware is `func(http.Handler) http.Handler`.

---

## 2. What httputil provides (the parts that matter to appkit)

### Server lifecycle (`server.go`) — **directly competes with appkit `server.go`**

- `Server` wraps `http.Server`.
- `NewServer(cfg)` calls `cfg.Validate()`.
- `Start()` is **non-blocking**, returns `<-chan error` for listen failures.
- `Shutdown(ctx)` graceful shutdown.
- `Addr()` access to bound address.
- `DefaultServerConfig()` — `:8080`, read 10s / read-header 5s / write 30s / idle 60s.
- Populates the full `http.Server` struct literally (enforced by `exhaustruct` lint).

### Health (`health.go`) — **directly competes with appkit `health.go`**

- `HealthHandler`, `LiveHandler` (semantically distinct alias).
- `ReadyHandler`, `ReadyHandlerWithProbe(ready func() bool)` (200 up / 503 down).
- `RegisterHealth(mux)` registers `GET /health`, `/health/live`, `/health/ready`.
- Uses Go 1.22+ method-pattern routing.

### 13 production middlewares appkit does **not** have

`CORS`, `Compression` (RFC 7231 negotiation, brotli/zstd/gzip/deflate, q-values, per-encoding pools),
`ETag` (FNV-64a, conditional 304, 1 MB cap), `SecurityHeaders` (nosniff, X-Frame-Options,
Referrer-Policy, CSP, HSTS), `RequestID` (time-ordered 16-byte IDs), `Recovery`, `Timeout`,
`Logging`, `MaxBodySize`, `RateLimit` (pluggable `RateLimiter`, `TokenBucketLimiter`), `Metrics`
(pluggable `MetricsRecorder`), `ClientIPMiddleware`, plus `ClientIP(r)`.

### Composition & utilities

- `Middleware = func(http.Handler) http.Handler` (universal unit).
- `Chain(handler, middlewares...)` — first arg outermost (reverse-declaration order).
- `MiddlewareStack` — named collector, duplicate prevention, ordering validation (Recovery outermost).
- `ResponseRecorder`, `DetectCapabilities`, `DefaultIncompressibleTypes`, `DefaultWriterFactories`,
  `RegisterErrorClassifications`, `RequestIDFromContext`, `ClientIPFromContext`.

### `httpspec` subpackage

`httpspec.Run(t, handler, opts...)` validates standard HTTP conventions via 18 parallel specs
(routing, headers, methods, security). Pure stdlib. Useful for testing appkit's server, too.

---

## 3. Current usage in go-appkit

**Zero.** Evidence:

- `go.mod` requires only `modernc.org/sqlite v1.53.0` (and its indirect deps).
- No import of `github.com/larsartmann/httputil` anywhere in the source tree.
- The name "httputil" appears only in this analysis folder.

---

## 4. Applicability assessment — the duplication, side by side

| Concern                                  | appkit (`server.go` / `health.go`)               | httputil (`server.go` / `health.go`)        | Duplication?                                    |
| ---------------------------------------- | ------------------------------------------------ | ------------------------------------------- | ----------------------------------------------- |
| Wrap `http.Server`                       | `Server` struct                                  | `Server` struct                             | ✅ Duplicate                                    |
| Non-blocking start, listen-failure error | `Start(ctx)` returns `error` (blocks on `errCh`) | `Start()` returns `<-chan error`            | ✅ Near-duplicate (httputil's design is better) |
| Graceful `Shutdown(ctx)`                 | yes                                              | yes                                         | ✅ Duplicate                                    |
| Bound-address access (`Addr()`)          | yes, `sync.RWMutex` guarded                      | yes                                         | ✅ Duplicate                                    |
| Config with defaults + timeouts          | `ServerConfig` + `applyDefaults()`               | `ServerConfig` + `Validate()`               | ✅ Duplicate                                    |
| Register `/health`                       | `NewServer` registers `GET /health`              | `RegisterHealth(mux)` registers 3 endpoints | ✅ Overlapping                                  |
| Health status → HTTP status              | `HealthStatus.HTTPStatus()`                      | `ReadyHandlerWithProbe` (200/503)           | ✅ Overlapping                                  |
| Production middlewares                   | **none**                                         | 13                                          | ❌ httputil-only                                |
| Method-pattern routing                   | `mux.HandleFunc("GET /health", …)`               | same                                        | ✅ Same approach                                |
| Config validation                        | `applyDefaults()` (zero-fill)                    | `Validate() error` (startup checks)         | ⚠️ httputil is stricter                          |

**Conclusion:** appkit's server/health are a **strict subset** of httputil, re-implemented. The only
delta in appkit's favor is the `RegisterHealth bool` opt-out toggle and the explicit `HealthStatus`
enum (ok/ready/unhealthy/degraded) — both trivially expressible on top of httputil.

This is the textbook "split brain" the project's own AGENTS.md philosophy warns against
(_"Split brains? → Check for duplicate type definitions; consolidate"_).

---

## 5. Integration analysis — dependency direction & cost

- **Direction is correct.** appkit is low-level glue; httputil is same-layer HTTP kit. appkit
  importing httputil is a clean, non-inverted dependency. (Contrast with cqrs-htmx, which must
  _not_ be a dependency of appkit.)
- **Cost is low.** httputil's only dep is go-error-family (zero transitive deps). Adding it brings
  essentially one transitive-free library into appkit's tree — consistent with appkit's "small,
  opinionated skeleton" ethos.
- **Risk is migration, not correctness.** Existing appkit consumers call `appkit.NewServer`,
  `appkit.ServerConfig`, `appkit.DefaultHealthHandler`, etc. Replacing these is a breaking change.

### Two viable resolutions

**Option A — Delegate (recommended for v1): keep appkit's API, re-implement on httputil.**
appkit's `Server` becomes a thin wrapper that constructs and owns a `httputil.Server`, and
`health.go` calls `httputil.RegisterHealth`. Public surface stays stable; implementation stops
duplicating. Middlewares become available to consumers via appkit re-export or direct httputil import.

**Option B — Retire (cleaner, breaking): tell consumers to use httputil directly.**
Delete `server.go` + `health.go` from appkit; document httputil as the server/health layer;
keep appkit focused on `logger.go`, `sqlite.go`, `shutdown.go` (the parts httputil does _not_ cover).
Document the migration. This yields the smallest, clearest responsibility split.

**Recommendation:** **Option B long-term, Option A as an interim step** if breaking changes aren't
acceptable right now. Either way, the duplication must end.

---

## 6. Concrete adoption plan (Option A — interim, non-breaking)

1. Add `github.com/larsartmann/httputil v0.4.0` to `go.mod`.
2. Rewrite `server.go` internals: `Server` embeds/owns `*httputil.Server`; `Start` delegates to
   `httputil.Server.Start` (adapt the blocking-vs-channel difference in a small goroutine bridge);
   `Shutdown`/`Addr`/`Running` delegate directly.
3. Rewrite `health.go` to delegate registration to `httputil.RegisterHealth`, keeping the
   `HealthStatus` enum as appkit's public convenience but mapping it to httputil's handlers.
4. Optionally re-export the middlewares most relevant to a skeleton
   (`Recovery`, `RequestID`, `Logging`, `Timeout`) so consumers get sane defaults.
5. Run `httpspec.Run` against appkit's server in tests (new coverage for free).
6. Keep `go test ./... -race` and `go vet ./...` green.

### Code sketch (Option A, `server.go` core)

```go
type Server struct {
    inner *httputil.Server
    cfg   ServerConfig
}

func NewServer(cfg ServerConfig, mux *http.ServeMux) *Server {
    cfg.applyDefaults()
    if cfg.RegisterHealth {
        httputil.RegisterHealth(mux) // registers /health, /health/live, /health/ready
    }
    hcfg := httputil.DefaultServerConfig()
    hcfg.ReadTimeout, hcfg.WriteTimeout, hcfg.IdleTimeout = cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout
    // httputil.NewServer port mapping etc. (hcfg holds its own port)
    return &Server{inner: httputil.NewServer(hcfg, mux), cfg: cfg}
}

func (s *Server) Start(ctx context.Context) error {
    errCh := s.inner.Start() // non-blocking, returns <-chan error
    select {
    case <-ctx.Done():
        return nil
    case err := <-errCh:
        return err
    }
}
```

---

## 7. What "fully" would mean here

For appkit to use httputil "fully" would mean: **appkit owns no server or health code of its own.**
All HTTP server lifecycle, health endpoints, and the default middleware stack come from httputil.
appkit's remaining job is the skeleton concerns httputil does not cover: **logger init, SQLite open,
signal-based shutdown orchestration** — i.e. exactly the three files httputil doesn't touch.

---

## 8. Summary

- **Using today?** No.
- **Fully?** No — and appkit shouldn't "fully" adopt every middleware; it should stop duplicating
  the `Server`/`Health` core and let consumers pull middlewares directly from httputil.
- **Action:** Resolve the duplication (Option A interim / Option B long-term). This is the single
  highest-value change in this whole analysis folder.
