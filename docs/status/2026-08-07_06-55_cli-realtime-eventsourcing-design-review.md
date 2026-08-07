# Status Report: CLI + Real-Time + Event Sourcing Design Review

> **Date:** 2026-08-07 06:55
> **Session scope:** Design proposal for go-appkit targeting CLI apps, real-time web frontend, superb error handling, and event sourcing via go-cqrs-lite/system.
> **Verdict:** Design proposed, NOT implemented. Several gaps and fabrications found on self-review.

---

## a) FULLY DONE

### Design proposal delivered (but not validated)

A complete architectural design was produced covering:

1. **`App` orchestrator type** — single composition root owning `*system.System`, logger, CLI, HTTP service, and realtime hub. Three entry points: `RunCLI`, `Serve`, `Run` (hybrid).
2. **CLI runtime** — dispatches commands through the same `system.System`, with mode-aware lifecycle (no listener, no signal handling, synchronous projection drain).
3. **Real-time layer (`appkit/realtime`)** — SSE hub bridging `system.Bus()` to browsers with bounded per-client buffers, configurable backpressure (DropOldest / CloseClient), and query-based filtering.
4. **Error handling design** — audience-aware rendering matrix (CLI exit codes, HTTP status codes, SSE error events) from a single go-error-family classification.
5. **Module evolution plan** — rewrite `cqrs` to target `system`, add `realtime` as 4th module, keep core unchanged.
6. **Consumer code example** — ~30 line wiring showing domain config, deployment config, routes, CLI commands, realtime mount, and `app.Run(ctx)`.

### Codebase study completed

Read and analyzed during this session:

- `go-appkit/service.go` — Service type, lifecycle (Start/Run/Shutdown/Close)
- `go-appkit/config.go` — ServiceConfig, defaults, validation
- `go-appkit/middleware.go` — default middleware stack and override logic
- `go-appkit/errors.go` — re-exports of HTTPStatus and LogError (THIS WAS READ TOO LATE)
- `go-appkit/cqrs/eventservice.go` — current EventService wrapping stack/sqlite.Bundle
- `go-appkit/cqrs/go.mod` — confirms v3 dependencies (NOT v4)
- `go-appkit/example/main.go` — current consumer pattern with `errorfamily.HandleError`
- `go-appkit/docs/planning/framework-architecture.md` — modular 3-module plan
- `go-appkit/docs/planning/design-decisions.md` — 10 LOCKED decisions (READ TOO LATE)
- `go-cqrs-lite/system/system.go` — System type, ownership model
- `go-cqrs-lite/system/constructor.go` — New(), engine wiring, projection host setup
- `go-cqrs-lite/system/register.go` — RegisterDecider, RegisterCommand, RegisterQuery generics
- `go-cqrs-lite/system/config_types.go` — DomainConfig + DeploymentConfig separation
- `go-cqrs-lite/system/bus.go` — simpleBus in-process event bus implementation
- `go-cqrs-lite/system/README.md` — API surface and quick-start

---

## b) PARTIALLY DONE

### Nothing is partially implemented.

This was a design-only session. Zero lines of code were written, zero tests, zero files modified. The design is a proposal that needs validation before any implementation begins.

---

## c) NOT STARTED

Everything below was discussed in the design but has zero implementation:

1. **`App` type** — no `app.go` file exists
2. **`AppConfig` / `AppOption`** — referenced in design, never defined
3. **CLI runtime** — no `cli.go`, no subcommand parser, no framework choice (cobra? urfave/cli? custom?)
4. **`appkit/realtime` module** — no directory, no `go.mod`, no Hub type, no SSE handler
5. **`cqrs` rewrite** — current `eventservice.go` still wraps `stack/sqlite.Bundle` (v3)
6. **Error renderer** — no `ErrorRenderer` type, no audience-aware rendering
7. **System health integration** — no wiring between `system.Health()` and appkit's `/health/ready`
8. **Graceful shutdown ordering** — no code for projection-stop → bus-drain → HTTP-drain → store-close sequence
9. **Migration path** — no plan for existing `EventService` consumers
10. **Tests** — none written, no test strategy defined

---

## d) TOTALLY FUCKED UP

### 1. Read design-decisions.md AFTER proposing the design

`design-decisions.md` contains **10 LOCKED decisions** for v1.0.0. I proposed a design that:

- **Violates Decision 4** ("SQLite via cqrs-lite/stack is a separate module") by proposing to rewrite cqrs to target `system` instead of `stack/sqlite` — without acknowledging the locked decision or proposing how to amend it.
- **Violates Decision 9** ("core v1.0.0, sub-modules v0.1.0") by casually proposing a new `realtime` module and a `cqrs` rewrite without addressing versioning implications.
- **Ignores Decision 10** (graceful drain sequence: ready-probe → drain-delay → server-shutdown → bundle-close → logger-flush) — my design proposed a different shutdown ordering without referencing the locked one.
- **Duplicates Decision 6** (go-error-family 3-layer adoption) — I proposed an `ErrorRenderer` as if it were new, when Decision 6 already defines `HTTPHandler`, `LogError`, and `HandleError` as boundary terminators.

This is the biggest failure of the session. I should have read the locked decisions BEFORE designing, not after.

### 2. Proposed `system.DefaultSQLiteDeployment("app.db")` — may not exist

The system README shows `DeploymentConfig` constructed from structs. I invented a convenience constructor that was never verified to exist in the `system` package. This is fabrication.

### 3. Invented error classification categories

My error → audience matrix included `NotFound` (404) and `Conflict` (409) as distinct classifications. go-error-family has three families: **Rejection**, **Transient**, **Infrastructure**. `NotFound` and `Conflict` are not top-level families — they may be sub-categories or may not exist at all. I did not verify the actual go-error-family API before presenting the table as fact.

### 4. Proposed `errorfamily.Classify(err)` — unverified API

I used `errorfamily.Classify(err)` in the `RenderError` code sample. The actual API may be different (e.g., `errorfamily.Family(err)` or type-assertion on classified error types). I never checked.

### 5. Ignored the v3 → v4 version gap

The current `cqrs/go.mod` depends on `go-cqrs-lite/*/v3`. The `system` package is `go-cqrs-lite/system/v4`. My design proposed switching to `system` without acknowledging this is a **major version migration** across 20+ transitive dependencies. The v3 → v4 jump affects event, command, query, decider, id, metaengine, projectionhost — all of which would need version bumps in appkit's dependency tree.

### 6. Proposed dropping `cqrs/eventservice.go` without impact analysis

I wrote "The current `cqrs/eventservice.go` dies" — destructive language with zero analysis of:

- What consumers depend on `EventService.Bundle()`, `EventService.Host()`, `EventService.DB()`
- Whether `eventservice_test.go` would need full rewrite
- Whether any existing downstream code uses the v3 `stack.Bundle` API directly

### 7. Didn't verify `system.Bus()` returns a usable pub/sub for SSE fanout

I assumed `system.Bus()` (which returns `event.Bus`) is suitable for realtime fanout. But `simpleBus` dispatches **synchronously on the publishing goroutine** — meaning a slow SSE client write would block the command dispatch path. My design mentioned this ("without letting a slow browser block the bus") but the actual solution (async fan-out goroutine + bounded channels) was hand-waved, not designed.

### 8. No CLI framework decision

I proposed `app.CLI().Command(...)` without choosing or even discussing what CLI framework to use. The options (cobra, urfave/cli, custom stdlib parser) have radically different dependency profiles and API ergonomics. This is a fundamental architectural decision that was skipped entirely.

---

## e) WHAT WE SHOULD IMPROVE

### Design process improvements

1. **Read ALL planning docs before designing** — design-decisions.md, framework-architecture.md, execution-plan.md, and all prior status reports. Designing without reading locked decisions is malpractice.
2. **Verify external APIs before writing code samples** — every function name, type, and category I referenced from go-error-family and system should have been verified against actual source, not assumed from READMEs.
3. **Address version compatibility explicitly** — the v3 → v4 gap is a project-defining constraint. It should have been the FIRST thing discussed, buried in a footnote.
4. **Impact analysis before proposing deletion** — "X dies" requires knowing what depends on X.
5. **Choose frameworks, don't hand-wave them** — CLI framework, SSE library (or raw stdlib), WebSocket library — these are real decisions with real dependency costs.

### Design gaps to close

6. **Shutdown ordering** — must reconcile my proposed ordering with Decision 10's locked sequence. The System's `Close()` stops projections then closes engines. appkit's `Shutdown()` drains HTTP then closes. These must compose correctly.
7. **Health integration** — `system.Health()` and `system.Explain()` exist. appkit's `/health/ready` should integrate with system health, not just the HTTP ready probe.
8. **SSE authentication** — no discussion of how auth tokens reach the SSE endpoint (query param? cookie? header?). SSE cannot set custom headers from the browser side.
9. **Event replay on reconnect** — when an SSE client reconnects, it needs to catch up. The event store supports `SeekableJournal` — the design should specify how the Hub uses `Last-Event-ID` header to replay missed events from the store.
10. **CORS for SSE** — browser SSE requires CORS headers. No discussion.
11. **Projection host lifecycle in CLI mode** — CLI commands that produce events need projections to process them. But projections are async by default. Does CLI mode wait for projection catch-up? How?
12. **DeploymentConfig loading** — `system/config_loader.go` exists for YAML loading. The design should specify how appkit loads deployment config (env vars? YAML file? code?).
13. **Multi-instance realtime** — if the app scales to multiple processes, the in-process `simpleBus` doesn't fan out across instances. The design should acknowledge this limitation and point to external bus drivers (NATS, Redis) for horizontal scaling.

---

## f) Up to 50 Things We Should Get Done Next

### Validation (do these FIRST — before any code)

1. **Read `docs/planning/execution-plan.md`** — understand the current execution plan
2. **Read all status reports in `docs/status/`** — understand what's been tried and decided
3. **Read `docs/planning/improvement-audit.md`** — understand known issues
4. **Read `docs/planning/integrations.md`** — understand integration plans
5. **Verify go-error-family actual API** — `Classify` vs `Family`, actual error categories, `HandleError` signature
6. **Verify `system` package actual public API** — does `DefaultSQLiteDeployment` exist? What constructors are available?
7. **Check if `system/v4` is stable or experimental** — go-cqrs-lite ROADMAP.md and FEATURES.md
8. **Verify `event.Bus` interface** — exact method signatures for Publish, Subscribe, SubscribeAll
9. **Check `system/config_loader.go`** — how DeploymentConfig loads from YAML/env
10. **Read `system/introspection.go`** — Health(), Explain(), Snapshot() signatures

### Design reconciliation

11. **Reconcile proposed design with Decision 4** (stack/sqlite as separate module) — either amend the decision or design within its constraint
12. **Reconcile proposed design with Decision 9** (versioning) — define version strategy for new realtime module and cqrs rewrite
13. **Reconcile shutdown ordering with Decision 10** — compose System.Close() with Service.Shutdown() correctly
14. **Reconcile ErrorRenderer with Decision 6** — Decision 6 already defines boundary terminators; extend, don't replace
15. **Decide: evolve `EventService` or replace it** — impact analysis of each path
16. **Decide CLI framework** — cobra vs urfave/cli vs custom, with dependency cost analysis
17. **Decide SSE implementation** — raw `net/http` (stdio) vs library (depends on complexity)
18. **Decide WebSocket strategy** — opt-in submodule? part of realtime? deferred?
19. **Design the v3 → v4 migration path** — if cqrs moves to system/v4, all transitive deps bump
20. **Design SSE auth model** — cookie, query param, or token-in-header via EventSource polyfill
21. **Design SSE reconnect/replay** — Last-Event-ID → SeekableJournal.Seek → replay missed events
22. **Design projection catch-up in CLI mode** — synchronous wait or fire-and-forget
23. **Design `AppConfig` type** — what fields, what defaults, what validation
24. **Design `AppOption` functional options** — logger, middleware, realtime, CLI config
25. **Design health integration** — `/health/ready` calls `system.Health()` in addition to ready probe

### Implementation (after design is validated)

26. **Create `app.go`** — App type with New(), RunCLI(), Serve(), Run(), Close()
27. **Create `app_test.go`** — test all three runtime modes
28. **Create `cli.go`** — CLI subcommand registry and dispatch
29. **Create `cli_test.go`** — test CLI dispatch and exit codes
30. **Create `errorrender.go`** — audience-aware error rendering (CLI, HTTP, SSE)
31. **Create `errorrender_test.go`** — test all classification → audience mappings
32. **Create `appkit/realtime/` module** — go.mod, hub.go, sse.go
33. **Create `realtime/hub.go`** — Hub type subscribing to event.Bus, fan-out to clients
34. **Create `realtime/sse.go`** — SSE HTTP handler with Last-Event-ID support
35. **Create `realtime/hub_test.go`** — test fan-out, backpressure, client lifecycle
36. **Create `realtime/sse_test.go`** — test SSE encoding, reconnect, filtering
37. **Rewrite `cqrs/systemservice.go`** — wrap system.System instead of stack/sqlite.Bundle
38. **Create `cqrs/systemservice_test.go`** — test lifecycle integration
39. **Update `cqrs/go.mod`** — v3 → v4 dependencies (if migration is approved)
40. **Update `example/main.go`** — show CLI + HTTP + realtime in one example
41. **Create `example/main_test.go`** — test the example end-to-end
42. **Add SSE CORS middleware** — configurable CORS for realtime endpoints
43. **Add SSE filtering** — query param filtering by event type and stream ID
44. **Add backpressure strategy config** — DropOldest vs CloseClient, buffer size
45. **Integrate system introspection** — `/health/system` endpoint calling system.Health()

### Documentation

46. **Update `AGENTS.md`** — add App, CLI, realtime to module table
47. **Update `docs/planning/design-decisions.md`** — add decisions for CLI, realtime, error rendering
48. **Update `docs/planning/framework-architecture.md`** — add realtime module to the diagram
49. **Write realtime module README** — SSE usage, auth, backpressure, scaling limitations
50. **Write migration guide** — EventService → SystemService for existing consumers

---

## g) Questions I Cannot Answer Myself

### 1. Should cqrs migrate from go-cqrs-lite v3 to v4 (system)?

The `system` package is v4. The current `cqrs` module depends on v3 (`stack/sqlite/v3`, `projectionhost/v3`, etc.). Migrating to `system/v4` means bumping 20+ transitive dependencies across major versions. This is a project-defining decision that depends on:

- Is `system/v4` considered stable by go-cqrs-lite's own roadmap?
- Are there breaking changes in event/command/query/decider between v3 and v4?
- Is `stack/sqlite/v3` going to be maintained alongside `system/v4`?

I cannot determine this from code alone. **What is the intended relationship between `stack/v3` and `system/v4` in go-cqrs-lite? Is system meant to replace stack, or coexist?**

### 2. Is the realtime/SSE feature in scope for go-appkit, or should it be a separate repo?

The modular design principle (Decision 4) keeps heavy deps out of core. But `realtime` would add either:

- Zero new deps (raw stdlib SSE — more code to maintain)
- A lightweight SSE library dependency

Given that appkit is currently a thin HTTP lifecycle shell, real-time event streaming feels like it might belong in a separate consumer-facing framework, not in the infrastructure library. **Is realtime a first-class appkit concern, or should it be documented as a pattern (like Huma in Decision 7) rather than shipped as a module?**

### 3. Should the `App` type replace `Service`, or compose alongside it?

My design proposed `App` as a new orchestrator that owns `Service` internally. But Decision 1 locks `Service` as the v1.0.0 API contract. If `App` replaces `Service`, that's a breaking change to the locked v1 API. If `App` composes `Service`, the consumer has two entry points (`NewService` vs `NewApp`) which may confuse.

**Is the intent to keep `Service` as the stable v1 API forever, or is `App` meant to eventually supersede it as the primary entry point?**
