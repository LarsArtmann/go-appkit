# Core v1.0.0 Exit Criteria (draft)

**Status:** DRAFT — graduates to actionable when the consumer count grows
beyond the current family (cqrs-htmx adoption in flight). AGENTS.md names
v1.0.0 as the core target; this document defines what "done" means so the
target stays honest.

## Hard criteria (all must hold)

1. **API surface frozen for one full minor cycle.** The set of exported
   identifiers in the root package is unchanged across two consecutive
   tagged releases, verified mechanically by the release-ritual API
   snapshot diff (`go doc -all` at old tag vs working tree — the method
   proven during the 2026-09-04 wave: additions only → minor; any
   removal/signature change → minor-with-migration-notes at 0.x, never
   silent).
2. **Zero data-loss shutdown paths.** `Shutdown` runs DrainHooks → drain
   wait → listener close → ShutdownHooks exactly once, joins errors, and
   every hook-family has a race-detected test. NoDrainDelay semantics
   unchanged (explicit sentinel, never a default).
3. **Lifecycle guarantees documented and tested:** `NewService` registers
   health endpoints unless opted out; `Start` is idempotent-safe (second
   call rejected); `Shutdown` idempotent; `Addr()` nil before `Start`.
   Each guarantee has a named test.
4. **Error contract stable.** Every error leaving the framework is
   classified by go-error-family; HTTP status mapping covered by
   tests (Rejection 400, Conflict 409, Transient 503, Corruption 500,
   Infrastructure 503) — shared with errorpages.
5. **Consumer proof.** At least two independent consumers run the same
   core version in production-like setups for a full release cycle
   without patch-level workarounds (`replace` directives count as
   workarounds).
6. **Telemetry seam v1-shaped.** `OuterMiddlewares` + `ShutdownHooks` +
   `DrainHooks` (and the otel module's one-Setup-per-process contract)
   are stable enough that a consumer can wire full tracing/metrics
   without touching core internals.
7. **Docs tell the truth.** README quick start compiles verbatim;
   FEATURES.md statuses verified against code; no documented default
   diverges from `applyDefaults()`.

## Soft criteria (strong signals, not blockers)

- `GOEXPERIMENT` requirements gone from every module dependency chain
  (json/v2 default-on already removed the practical burden on Go 1.26.7+).
- Benchmark baselines recorded for the default middleware stack and the
  logging cost (2026-09-04 baseline: Logging-at-INFO ≈ +30µs/req vs
  suppressed ≈ +0.8µs — see `logging_bench_test.go`).
- Health module and otel module at ≥ v0.2.0 each (post-freeze hardening).

## Explicitly NOT required for v1

- New features (TLS option, additional middleware) — v1 freezes, not adds.
- WebSocket support (realtime is SSE-only by design).
- Multi-listener or HTTP/3 support.
