# Design Decisions — The Hard Choices

> **Status:** Locked for v1.0.0. **Date:** 2026-07-06.
> **Context:** Each decision shapes the entire API. Made before coding, not during.

---

## Decision 1: Service owns the mux

**Decision:** `Service` creates and owns `*http.ServeMux`, exposed via `svc.Mux`.

**Rationale:** If the consumer creates the mux, they must also wire health registration (`httputil.RegisterHealth`) and middleware wrapping themselves — defeating the purpose. The mux is the composition point: the consumer registers routes on it, appkit wraps it with middleware.

**Alternative considered:** Consumer passes mux to `NewService(cfg, mux)`. **Rejected:** forces consumer to know about health registration ordering and middleware wrapping. Breaks the "15-line service" promise.

**Implication:** Consumers who want a non-ServeMux router (chi, gin) can either: (a) use `svc.Mux` as the base and mount their router via `svc.Mux.Handle("/", chiRouter)`, or (b) skip `NewService` and use httputil directly. appkit is a stdlib-mux framework.

---

## Decision 2: Run(ctx) blocks AND Start() is available

**Decision:** `Run(ctx)` blocks (start → signal → shutdown → return). `Start() <-chan error` is available for non-blocking use.

**Rationale:** The value proposition is "one method and you're running." A non-blocking Run would force the consumer to write the signal-wait loop themselves. But production services with multiple servers or custom drain logic need non-blocking start.

**Both APIs:**

```go
// Simple (90% of services):
if err := svc.Run(ctx); err != nil { log.Fatal(err) }

// Advanced (multi-server, custom drain):
errCh := svc.Start()
// ... start gRPC server, workers, etc.
select {
case <-ctx.Done():
case err := <-errCh:
case sig := <-signalCh:
}
svc.Shutdown(shutdownCtx)
```

**Implication:** `Run()` is implemented as: `Start()` → wait → `Shutdown()`.

---

## Decision 3: Middleware is replaceable, not just extendable

**Decision:**

- `cfg.Middlewares []httputil.Middleware` — if non-nil, REPLACES the default stack.
- `cfg.ExtraMiddlewares []httputil.Middleware` — appends to the default stack.
- If both are nil, uses the opinionated default: Recovery → RequestID → Logging → Timeout → SecurityHeaders.

**Rationale:** A framework should trust its users. If I'm building an internal tool and don't want per-request logging, I should be able to disable it. The default stack is opinionated but overridable.

**Changed from v3:** v3 said "extendable only." This was paternalistic. Frameworks should have strong defaults and escape hatches, not locked defaults.

---

## Decision 4: SQLite via cqrs-lite/stack is a separate module

**Decision:** Core `go-appkit` has NO database dependency. SQLite/CQRS lives in `go-appkit/cqrs`.

**Rationale:** 90% of Go services don't need CQRS/ES. Forcing the entire go-cqrs-lite/stack/sqlite dependency tree on every consumer makes the framework hostile to its largest audience.

**Implication:** `ServiceConfig` in core has no `SQLitePath` field. The cqrs module's `EventConfig` has it. Core services are stateless by default.

---

## Decision 5: Logger via charmbracelet/log, not custom slog handler

**Decision:** `InitLogger` creates a `charmbracelet/log` logger, wraps it as a `slog.Handler` via `log.NewHandler()`, returns `slog.New(handler)`.

**Rationale:** The current `logger.go` is 108 lines of custom slog handler with manual `isTerminal()` detection, format switching, and level mapping. charmbracelet/log does all of this in a battle-tested library with pretty colored TTY output, JSON mode, and level control. ~5 lines replaces 108.

**Compatibility:** `LogLevel`/`LogFormat` types remain. `InitLogger(cfg) (*slog.Logger, error)` signature unchanged. Consumers see no difference — the logger just gets prettier.

**Namespace note:** `github.com/charmbracelet/log` v1.0.0 is the canonical module. `charm.land/*` v2 packages (fang, huh, glamour, lipgloss) are rebranded forks of OTHER charm tools. `charmbracelet/log` has not been rebranded and remains at `github.com/charmbracelet/log`. No conflict.

---

## Decision 6: go-error-family — proper 3-layer adoption

**Decision:** Not sentinel registration. appkit's errors ARE classified at construction.

**Layer 1 (library code — appkit itself):**

- Replace `errors.New("sqlite path is required")` with `errorfamily.NewRejection("sqlite.path_required", "sqlite path is required")`.
- Replace `fmt.Errorf("unsupported log level: %q", l)` with `errorfamily.NewRejection("log.level_invalid", "unsupported log level: %s", l)`.
- Use `errorfamily.WrapInfrastructure(err, "server.shutdown_failed", "shutdown failed")` for wrapping.
- Classifiers ONLY for third-party errors: `RegisterClassifier` for `*sqlite.Error` → Transient.

**Layer 2 (boundary terminators — in the framework):**

- `errorfamily.HTTPHandler(handler)` wraps error-returning HTTP handlers → maps family to status code → writes safe JSON (never leaks `err.Error()`).
- `errorfamily.LogError(err, logger)` in shutdown/error paths → auto severity (Transient=Warn, others=Error) → structured fields.
- `errorfamily.HandleError(err)` as the example `main()` CLI terminator → exit code + Wix-quality stderr message.

**Layer 3 (application enrichment — NOT in appkit):**

- `go-error-family/bridge` + `samber/oops` for stack traces, trace IDs, rich typed context.
- This is application-only. appkit as a library does NOT import the bridge.
- Documented as a pattern for consumers who want oops enrichment.

**Anti-pattern (what we do NOT do):**

- `fmt.Errorf` for appkit's own errors.
- Sentinel registration for appkit's own errors (only for third-party).
- Importing `go-error-family/bridge` (that's application territory).

---

## Decision 7: Huma is documented, not depended on

**Decision:** appkit documents the Huma integration pattern but does NOT import Huma.

**Rationale:** Huma wraps `svc.Mux` INSIDE via `humago.New(mux, config)`. httputil middleware wraps OUTSIDE via `httputil.Chain`. The layers never collide — httputil already [documented and proved this](https://github.com/LarsArtmann/httputil/blob/main/docs/integrations/huma.md). Huma is a consumer choice, not a framework dependency.

**What appkit provides:**

- `svc.Mux` — the mux, exposed for wrapping.
- README section documenting the Huma pattern.
- Example showing both plain mux and Huma-wrapped mux paths.

**What the consumer does (if they want Huma):**

```go
api := humago.New(svc.Mux, huma.DefaultConfig("My API", "1.0.0"))
huma.Get(api, "/users/{id}", typedHandler)
```

---

## Decision 8: catalog/docserver is an opt-in module

**Decision:** Auto-documentation (catalog, AsyncAPI, D2, EventCatalog) lives in `go-appkit/docs`, not core.

**Rationale:** Auto-documentation is heavy (reflection, JSON Schema, embedded JS/CSS for Scalar/AsyncAPI React). Most services don't need it at runtime. It's a power feature for teams that want self-documenting services.

**What the docs module provides:**

- `docs.NewCatalogBuilder(title, version)` → register commands/events/queries.
- `docs.RegisterDocs(svc.Mux, builder)` → mounts `/docs/openapi`, `/docs/asyncapi`, `/docs/diagram`.
- D2 architecture diagrams from Go types.
- EventCatalog MDX file tree export.

---

## Decision 9: Version target — core v1.0.0, sub-modules v0.1.0

**Decision:**

- `go-appkit` (core): tag `v1.0.0`. The `Service` API is the v1 contract. Additive-only after.
- `go-appkit/cqrs`: tag `v0.1.0`. Experimental. API may change.
- `go-appkit/docs`: tag `v0.1.0`. Experimental. API may change.

**Rationale:** Core is the 80% — it must be stable for adoption. CQRS and docs are opt-in power features — they earn v1.0.0 when their APIs prove stable in real services.

**Pre-v1 dependency risk:** httputil (v0.4.0) and go-error-family (v0.6.1) are pre-v1. Documented risk: v1.0.0 means "the Service API is stable," not "all deps are frozen at v1+." go-error-family root has zero deps and a proven stable interface. httputil's Server/middleware API has been stable since v0.3.0.

---

## Decision 10: Graceful drain is explicit, not implicit

**Decision:** Document the drain sequence. Make it configurable.

**The sequence:**

```
1. Ready probe starts returning 503    → LB stops sending new traffic
2. Wait drainDelay (default: 5s)       → LB notices, propagates
3. Server.Shutdown(ctx)                → stop accepting, finish in-flight
4. Bundle.GracefulClose(ctx)           → drain bus, flush projections, close DB
5. Logger flush                        → ensure all buffered logs written
```

**Config:**

- `ServiceConfig.DrainDelay time.Duration` — delay before Shutdown (step 1-2). Default: 5s. Set to 0 to disable.
- `ServiceConfig.ShutdownTimeout time.Duration` — timeout for steps 3-4. Default: 15s.

**Implementation:** If `DrainDelay > 0`, `Run()` flips `ReadyHandlerWithProbe` to returning `false` before sleeping `DrainDelay`, then calls `Shutdown()`.
