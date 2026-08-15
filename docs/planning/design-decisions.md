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

---

## Decision 11 / ADR-001: cqrs-htmx ↔ appkit relationship — appkit-as-foundation, validated by spike

> **Status:** Decided 2026-08-15. **Supersedes:** the undocumented "cqrs-htmx becomes appkit's first real consumer" plan (docs/planning/integrations.md, 2026-07-06), which was silently abandoned — cqrs-htmx v4.8.0 has zero appkit references.
> **Evidence:** source audit of cqrs-htmx `setup/` vs appkit core, 2026-08-15 ([Ecosystem Utilization Audit](./../research/2026-08-15_ecosystem-deep-dive.html)).

**Decision:** appkit becomes the **generic server layer** for cqrs-htmx (`setup.Run` constructs `appkit.Service` behind its existing config), cqrs-htmx remains the **domain layer** (session/CSRF/CSP-nonce middleware, projection-aware readiness). Adoption is gated by a bounded prototype spike (plan task M18); if the spike shows regressions that cannot be fixed with additive appkit changes, we fall back to Option B automatically — no rewrite of either side.

### Options considered

| Option                                       | Verdict       | Why                                                                                                                                                                                                                                     |
| -------------------------------------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **A. appkit-as-foundation** (chosen)         | **Adopt**     | appkit has 7 capabilities cqrs-htmx lacks (see below); both regressions are fixable additively; ends the split brain where two frameworks duplicate generic service concerns on the same upstreams (httputil, go-error-family, go-sse). |
| B. Parallel assemblies (independent futures) | Fallback only | Ratifies the split brain permanently: drain, readiness, flight recorder, live/ready split would each be built twice. Only wins if the spike proves the swap regresses cqrs-htmx irreparably.                                            |
| C. Retire the relationship plan entirely     | Rejected      | Orphans the consumer story for appkit's cqrs/errorpages/realtime modules and discards verified unique value. The plan being dead as originally conceived does not mean the relationship should be.                                      |

### Verified appkit-unique value (cqrs-htmx `setup/` lacks all of these)

1. Drain phase on shutdown — readiness flip → `DrainDelay` → `server.Shutdown` (`service.go:134-164`); cqrs-htmx's `setup.RunHandler` calls `Shutdown` directly (`setup/run.go:93-103`) with zero LB-drain handling.
2. `Addr() net.Addr` — live listener address (`service.go:172-181`); cqrs-htmx never captures the listener (httputil `Start()` uses `ListenAndServe`).
3. Flight-recorder middleware module (trace capture on 5xx/latency); cqrs-htmx has none.
4. RequestID + request-logging + Timeout in the default middleware stack (`middleware.go:12-18`); cqrs-htmx's chain is security-headers + nonce + recovery only.
5. Live/ready split (`/health`, `/health/live`, `/health/ready` with atomic probe).
6. Built-in SIGINT/SIGTERM handling in `Run()`.
7. Charmbracelet-backed logger option.

### Verified cqrs-htmx advantages (must not regress)

- **SSE-safe timeout policy:** `setup/run.go` deliberately omits Read/WriteTimeout. appkit force-defaults `WriteTimeout` to 30s (`config.go:86-88`) — this would kill SSE streams >30s, including appkit's own `realtime` module mounted behind `appkit.Service` **today**. → Prerequisite P1 (additive opt-out) below.
- **Projection-aware readiness:** `ProjectionReadinessCheck` with named checks and JSON detail is richer than appkit's boolean probe. → Prerequisite P2 (M08 `ReadyCheck()` composition) below.
- CSP nonce, session, CSRF middleware: preserved — appkit's middleware stack is replaceable (Decision 3), cqrs-htmx keeps its own chain.

### Prerequisites (all additive, no public API breaks)

- **P1 — SSE-safe timeouts:** add an opt-out so `WriteTimeout` can be disabled for SSE services (zero value currently means "default", never "off"). Fixes a live appkit+realtime limitation regardless of the spike outcome.
- **P2 — readiness composition:** M08's `EventService.ReadyCheck()` plus appkit's probe pattern let cqrs-htmx keep projection-aware 503 semantics under appkit.
- **P3 — spike M18:** swap `setup.Run`'s server layer for `appkit.Service` on a cqrs-htmx branch; verify SSE header flush through appkit's middleware chain, drain-probe ↔ readiness wiring, and no behavioral regressions vs baseline. Adopt/reject on evidence.

### Consequences

- M18 (prototype spike) runs; its report decides final adoption (A confirmed vs fallback B).
- The 2026-07-06 claim "appkit will be httputil.Server's first real consumer" is corrected: appkit owns `http.Server` directly; cqrs-htmx uses httputil's `Server` today and would replace it with appkit only via the spike.
- cqrs-htmx keeps its richer domain middleware; appkit never grows session/CSRF concerns — those are application territory.
