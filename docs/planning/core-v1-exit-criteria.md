# Core v1.0.0 Exit Criteria

**Status:** ACTIONABLE (2026-10-13) — the draft's own graduation trigger
fired: consumer #2 (PapDashboard v0.3.0, published 2026-10-13) exists
beyond the family. This document is now the active v1.0.0 checklist;
each future core release train re-assesses the latest table below.
AGENTS.md names v1.0.0 as the core target; this document defines what
"done" means so the target stays honest.

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

## Graduation status — 2026-10-13 (vs core v0.9.0, consumer #2 arrived)

Assessed against master at core v0.9.0 (2026-10-12), go floor 1.27.2,
SUPERB v7 T11 (refreshing the 2026-09-29 table):

| # | Criterion                                         | Status           | Evidence / gap                                                                                                                                                                                                                                                                                                                                                      |
| - | ------------------------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | API frozen one minor cycle                        | NOT MET          | Removal-free streak intact since v0.4.0, but additions kept moving the identifier SET (v0.8.0 `StartHooks`, v0.9.0 `testkit.NewService`). Freeze window now: v0.9.0 → next tag must be additions-EMPTY to start the cycle; v1.0.0 would be its second point. The check is mechanical (Release Ritual step 1, proven 2026-09-04).                                    |
| 2 | Zero data-loss shutdown paths                     | MET              | Unchanged since 2026-09-29; v0.8.0 added the startup mirror (hook failure closes the listener, pinned by starthooks_test).                                                                                                                                                                                                                                          |
| 3 | Lifecycle guarantees + named tests                | MET              | Unchanged; Addr()/Running() startup-mirror contract pinned (v0.8.0).                                                                                                                                                                                                                                                                                                |
| 4 | Error contract stable                             | MET              | Unchanged; errorpages parity re-verified at errorpages v0.1.2 (2026-10-13).                                                                                                                                                                                                                                                                                         |
| 5 | Two independent consumers, full cycle, no replace | NOT MET (closer) | Consumer #2 EXISTS: PapDashboard v0.3.0 requires core v0.4.0, zero `replace` directives (verified 2026-10-13), plain Service usage. Gaps: versions don't overlap (setup v0.8.0 / PapDashboard v0.4.0) and neither pair has run a full release cycle together. Their refresh is THEIR repo's call (do not nudge); the consumer-claims guard fires on their next tag. |
| 6 | Telemetry seam v1-shaped                          | MET              | Unchanged; otel v0.2.0 (SigNoz OTLP + exception capture) composes without core changes.                                                                                                                                                                                                                                                                             |
| 7 | Docs tell the truth                               | MET              | The 2026-09-29 FEATURES otel split brain closed (v0.2.0 surface listed); FEATURES truth-swept 2026-10-13 (SUPERB v6 TODO cleanup).                                                                                                                                                                                                                                  |

Soft: GOEXPERIMENT-free floor MET; benchmarks recorded (logging 2026-09-04; otel re-baseline at the 1.27.2 floor in flight, SUPERB v7 T10); health v0.2.0 AND otel v0.2.0 both shipped — the soft bar is fully MET.

**Bottom line (2026-10-13):** 5/7 hard criteria met. The doc graduated
from DRAFT on its own trigger (second consumer exists). Remaining
blockers, both external-clock: criterion 1 needs one additions-EMPTY
core tag pair; criterion 5 needs a version overlap + one observed
release cycle (PapDashboard's refresh cadence is theirs). Nothing in
this repo blocks v1.0.0 except restraint.

## Graduation status — 2026-11-02 (SUPERB v10 M23 re-verify of the 2 blockers)

Assessed against master at 4eb7cf0 (2026-11-02), both blockers re-checked
mechanically:

| # | Criterion                                         | Status             | Evidence / gap                                                                                                                                                                                                                                                                                                                |
| - | ------------------------------------------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | API frozen one minor cycle                        | NOT MET, ARMED     | Release-ritual API diff v0.9.0 → working tree = ZERO lines (exported set identical, 23 days additions-free). The NEXT core tag cut from this tree IS the additions-empty anchor; v1.0.0 remains its second point. Clock starts at the next tag, whenever that train fires.                                                    |
| 5 | Two independent consumers, full cycle, no replace | NOT MET, UNCHANGED | Consumer-claims guard green 2026-11-02: setup/v4.15.0 still the newest setup tag (pins core v0.9.0 = latest — overlap holds on that leg); PapDashboard v0.3.0 still their newest (pins core v0.4.0 — overlap still absent, refresh is their repo's call, guard fires on their next tag). No observed joint release cycle yet. |

All other criteria unchanged from the 2026-10-13 table (5/7 met; soft bar
fully met).

**Bottom line (2026-11-02):** blocker state unchanged but sharpened —
criterion 1 is now ARMED (tree verified additions-empty vs v0.9.0; one
quiet doc-only-or-no-op core tag starts the cycle); criterion 5 waits on
PapDashboard's version refresh (external clock, guarded). This repo still
needs nothing except restraint.

## Graduation status — 2026-11-08 (post-v0.12.0: criterion 1 MET)

Assessed against master (2026-11-08), both blockers re-checked against the
released tag chain:

| # | Criterion                                         | Status             | Evidence / gap                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| - | ------------------------------------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | API frozen one minor cycle                        | MET                | The ARMED window reset once — v0.9.0→v0.10.0 was additions-only (RegisterClassifications + five testkit funcs), restarting the clock. Since then TWO consecutive additions-EMPTY diffs: v0.10.0→v0.11.0 (WrapOnce train, `go doc -all` proof, zero delta) and v0.11.0→v0.12.0 (the C1 httputil.Server composition swap, zero delta). That is one full frozen minor cycle — the anchor semantics hold, and any future core tag keeps the criterion only by staying additions-empty (or resets it). |
| 5 | Two independent consumers, full cycle, no replace | NOT MET, UNCHANGED | setup/v4.17.0 (newest setup tag, published-tag verified) pins core v0.10.0 — which is API-IDENTICAL to latest (both v0.11.0 and v0.12.0 were zero-diff), so the surface overlap holds; PapDashboard v0.3.0 still pins core v0.4.0 (their refresh is their call; consumer-claims guard green). No observed joint release cycle yet — the sole remaining blocker.                                                                                                                                   |

All other criteria unchanged from the 2026-10-13 table (now 6/7 met; soft
bar fully met).

**Bottom line (2026-11-08):** criterion 1 is MET — the repo's half of
v1.0.0 is done; criterion 5 (PapDashboard overlap + one observed joint
release cycle) is the only remaining blocker, and it is external-clock.
Restraint note: core tags between now and v1.0.0 must stay additions-empty
to keep criterion 1, by the table's own semantics.
