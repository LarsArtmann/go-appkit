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
   call rejected); `Shutdown` idempotent; `Addr()` nil before `Start` and
   from the moment `Shutdown` begins (drain-window contract, documented
   2026-09-17).
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

## Graduation status — 2026-09-29 (vs the 1.27.1 floor + core v0.7.0 surface)

Assessed against master at core v0.7.0 (2026-09-29 satellite-train day), toolchain unified at go 1.27.1:

| # | Criterion                                         | Status  | Evidence / gap                                                                                                                                                                       |
| - | ------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | API frozen one minor cycle                        | NOT MET | v0.6.0→v0.7.0 added surface (testkit.DrainWindowProbe, []Hook) — additions-only, removal-free streak intact, but the identifier SET changed. Candidate freeze window: v0.7.0→v0.8.0. |
| 2 | Zero data-loss shutdown paths                     | MET     | ShutdownHooks once + hook error-code contract pinned by tests (v0.6.0/v0.7.0 trains); drain-window ordering pinned in integration.                                                   |
| 3 | Lifecycle guarantees + named tests                | MET     | Addr() nil-during-drain pinned (2026-09-17, composition-contract suite); Shutdown idempotent; health opt-out tested.                                                                 |
| 4 | Error contract stable                             | MET     | go-error-family classification + HTTPStatus tests, shared with errorpages (identical mapping).                                                                                       |
| 5 | Two independent consumers, full cycle, no replace | NOT MET | One family consumer (cqrs-htmx setup, pinned v0.5.1 stable since 2026-09-07). PapDashboard not appkit-hosted; no second consumer. THE long pole.                                     |
| 6 | Telemetry seam v1-shaped                          | MET     | OuterMiddlewares + hooks stable; otel pattern/route propagation pinned against PUBLISHED tags (otel_pattern_test).                                                                   |
| 7 | Docs tell the truth                               | PARTIAL | FEATURES.md otel section still lists only the v0.1.x surface (split brain, flagged 2026-09-29 11:59 report); fix queued. Everything else current through the v0.7.0 train.           |

Soft: GOEXPERIMENT-free floor MET (retired 2026-09-29); benchmark baselines recorded (logging 2026-09-04, otel re-baselined 2026-09-16); health v0.1.4 and otel v0.1.1 both BELOW the ≥v0.2.0 soft bar (otel v0.2.0 queued, health's v0.2.0 unscoped).

**Bottom line:** 4/7 hard criteria met, none regressed; criterion 5 (consumer proof) is the structural blocker; criterion 1 starts its freeze window at the NEXT core tag.
