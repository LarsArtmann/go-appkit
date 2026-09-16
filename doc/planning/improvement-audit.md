# Improvement Audit — 15 Weaknesses in the v3 Plan

> **Status:** Critical analysis. **Date:** 2026-07-06.
> **Context:** Brutal self-review of the monolithic v3 plan before rewriting as modular.

---

## 🟥 Critical — Architecturally Broken

### 1. Scope Explosion: Two Products Bundled as One

**What's wrong:** The v3 plan bundles Product A (HTTP service framework: server + middleware + health + logger + signal) with Product B (CQRS/ES framework: event store + command store + bus + projections). 90% of Go services don't need CQRS/ES. By forcing every consumer to pull in the entire `go-cqrs-lite/stack/sqlite` dependency tree (event/v3, command/v3, query/v3, codec, storage/sql, watermill, CBOR, OTel SDK), we make the framework hostile to its largest audience.

**Fix:** Multi-module architecture. Core HTTP framework is dep-free of CQRS. CQRS is an opt-in sub-module. See [framework-architecture.md](./framework-architecture.md).

### 2. Projection Host Gap

**What's wrong:** A CQRS framework that gives you an event store but no way to run projections is half a framework. go-cqrs-lite has `projectionhost` (managed lifecycle, crash recovery, DLQ, backoff). The plan doesn't mention it. A CQRS service needs projections running — where do they live?

**Fix:** The `go-appkit/cqrs` module owns the projection host lifecycle. Projections start when the service starts, restart on crash with backoff, drain on shutdown. `EventService.RegisterProjection(name, projection, opts...)` manages workers.

### 3. No Story for the Bus

**What's wrong:** stack/sqlite gives you an in-process GoChannel bus. But production services need NATS/Kafka/Redis for distributed pub-sub, and SSE/WebSocket for client updates. The plan doesn't address how the bus is exposed or extended.

**Fix:** Document three tiers: (1) in-process (default, from stack/sqlite), (2) external (consumer swaps in `stack.WithBus(watermill.NewEventBus(backend))` for NATS/Kafka/Redis), (3) client-facing (SSE via `transport/http.SSEBroker`, WebSocket via cqrs-htmx). appkit doesn't own the bus — it exposes `svc.Bundle.Publisher/Subscriber` and lets the consumer choose.

### 4. Run() Blocking is Wrong for Production

**What's wrong:** Real production services need multiple servers (HTTP + gRPC + admin), background workers, and graceful drain (stop accepting → LB reroutes → wait in-flight → shutdown). A blocking `Run()` that owns signal handling makes all of these harder.

**Fix:** Provide BOTH:

- `Run(ctx)` — convenience, blocks, handles signals, calls Shutdown. For simple services and `main()`.
- `Start() <-chan error` + `Shutdown(ctx)` — for advanced use: multiple servers, workers, custom drain logic.

Document the drain sequence explicitly (see §11 below).

---

## 🟧 Important — Design Smells

### 5. "Non-Removable Middleware" is Paternalistic

**What's wrong:** Decision 3 says you can't remove Recovery or Logging. But a framework should trust its users. Internal tools may not want per-request logging. Test servers may not want Recovery.

**Fix:** `cfg.Middlewares []httputil.Middleware` — if non-nil, REPLACES the default stack. If nil, uses defaults. Plus `cfg.DisableDefaultMiddleware bool` as a shortcut for "I'll provide my own."

### 6. charmbracelet/log Namespace Inconsistency

**What's wrong:** cmdguard uses `charm.land/*` v2 packages (fang, huh, glamour, lipgloss, bubbles). The plan uses `github.com/charmbracelet/log` v1.0.0. Are these the same org? Will `charm.land/log` v2 appear? The plan should verify compatibility.

**Fix:** Research the Charm ecosystem migration. `github.com/charmbracelet/log` is the canonical module; `charm.land/*` v2 packages are the rebranded forks of OTHER charm tools (lipgloss, bubbletea, etc.). `charmbracelet/log` v1.0.0 is current and stable. No conflict — different packages, same org. Document this in the plan.

### 7. Pre-v1 Dependencies Threaten the v1.0.0 Contract

**What's wrong:**

- httputil: v0.4.0 (pre-v1)
- go-error-family: v0.6.1 (pre-v1)
- go-cqrs-lite/stack: v3.6.0 (stable, v3)
- charmbracelet/log: v1.0.0 (stable, v1)

If httputil or go-error-family bump major versions, appkit's v1.0.0 breaks.

**Fix:** Two options:

- **(a) Wait:** Don't tag appkit v1.0.0 until httputil and error-family reach v1. Ship appkit v0.1.0 in the meantime.
- **(b) Accept and document:** Tag v1.0.0 with a documented dependency-floor guarantee. v1.0.0 means "the Service API is stable," not "all deps are frozen at v1+."

Recommend **(b)** with explicit documentation. go-error-family root has zero deps and a proven stable interface. httputil's Server/middleware API has been stable since v0.3.0. The risk is real but bounded.

### 8. No End-to-End Integration Test

**What's wrong:** The "15-line service" is a demo, not a proven contract. There's no test that exercises: Service creation → CQRS command dispatch → event stored → projection read → health check → shutdown.

**Fix:** An e2e test in the cqrs module that:

1. Creates an EventService
2. Registers a command handler via Repository
3. Dispatches a command
4. Verifies the event was stored (read back from EventSource)
5. Runs a projection
6. Reads the projection from KV
7. Hits `/health/ready` → 200
8. Shuts down → verifies bundle closed

This test IS the CQRS integration contract.

---

## 🟨 Should Improve — Missing Depth

### 9. Huma + catalog Should Be Opt-In, Not Defaults

**What's wrong:** Huma + catalog are powerful but heavy (reflection, JSON Schema, embedded JS/CSS for Scalar/AsyncAPI React). Baking them into core bloats every consumer's binary.

**Fix:** `go-appkit/docs` sub-module. Consumer imports it only if they want auto-documentation. Core framework stays lean. See [framework-architecture.md](./framework-architecture.md).

### 10. go-error-family Adoption is Still Underspecified

**What's wrong:** The plan says "use constructors" but doesn't answer:

- Does `ServiceConfig.Validate()` return `*errorfamily.Error` or plain `error`?
- How does `HTTPHandler` integrate — wrap the mux? Individual handlers?
- Does appkit still export its own sentinels?
- Does `example/main.go` use `HandleError(err)` as CLI terminator?

**Fix:** Full specification in [error-family-adoption.md](./error-family-adoption.md). Short version:

- `Validate()` returns `error` (which happens to be `*errorfamily.Error` — satisfies the interface, callers don't need to know).
- `HTTPHandler` wraps individual handlers, not the mux. Documented as a pattern: `mux.Handle("POST /", errorfamily.HTTPHandler(myHandler))`.
- appkit does NOT export its own sentinels. All errors are `errorfamily.New*` constructors.
- `example/main.go` uses `os.Exit(errorfamily.HandleError(err))` as the main() terminator.

### 11. Graceful Drain Sequence is Undocumented

**What's wrong:** The plan says "signal → server.Shutdown → bundle.Close" but production needs a drain sequence:

**Fix:** Document the explicit drain sequence:

```
1. Ready probe starts returning 503    → LB stops sending new traffic
2. Wait drainDelay (e.g., 5s)          → LB notices, propagates
3. Server.Stop accepting connections   → in-flight requests continue
4. Server.Shutdown(ctx)                → wait for in-flight to finish (up to timeout)
5. Bundle.GracefulClose(ctx)           → drain bus, flush projections, close DB
6. Logger flush                        → ensure all buffered logs are written
```

`ServiceConfig.ShutdownTimeout` covers steps 3-5. `ServiceConfig.DrainDelay` (new field) covers step 1-2. If `DrainDelay > 0`, Run() flips the ready probe to failing before calling Shutdown.

### 12. No Story for Multiple Servers

**What's wrong:** A service often runs HTTP + gRPC, or HTTP + internal admin port. `Service` owns one mux and one server.

**Fix:** Document that consumers create multiple `Service` instances for multi-port. This is the simplest approach. Do NOT build `Service.AddServer()` — it adds complexity for a rare use case. The Service type is for the primary HTTP server. If you need gRPC, create a second Service or manage the gRPC server manually.

### 13. The Competitive Analysis Claim is Dishonest

**What's wrong:** "Nothing else in the ecosystem does this" — Buffalo, Fiber, and GoFr all bundle server + DB + middleware + logging.

**Fix:** Honest differentiator: "appkit is the only framework that composes the larsartmann ecosystem (httputil + cqrs-lite + error-family + charmbracelet/log) into a unified service." That's a valid niche. It's not "nothing else exists" — it's "nothing else composes THESE libraries."

### 14. No Configuration Validation for the Bundle

**What's wrong:** The plan validates `ServiceConfig` but not the CQRS stack. What if SQLitePath is read-only? What if schema migration fails?

**Fix:** Document: `cqrs.NewEventService` calls `stack/sqlite.New` which calls `SQLiteInitSchema`. Any failure (read-only dir, disk full, corrupt DB) returns an `Infrastructure`-classified error. `NewEventService` returns the error, the service never starts. This is correct behavior — document it explicitly.

### 15. The Plan is Becoming the Thing It Critiqued

**What's wrong:** v1 was "5 random helpers." v3 is "27 tasks wiring 4+ libraries into a monolithic framework with 18 sections." The simplest valuable thing is still: Service type + NewService + Run + httputil delegation.

**Fix:** Modular framework. Build core first (the 80%). CQRS and docs are separate modules that build ON the core. Each ships independently. See [framework-architecture.md](./framework-architecture.md).
