# Status Report — cqrs composition-root adoption closure, execution E1–E8

**Date:** 2026-10-06 22:21 · **Session:** Full Execution Mode of
`docs/planning/2026-10-06_15-20_SUPERB-cqrs-composition-root-adoption-closure.md`
(15 epics, 63 micro-tasks) · **Scope:** this session's run only
(cqrs module + environment); a second concurrent session worked
flightrecorder/otel and tagged `cqrs/v0.6.1` — their work is out of scope here.

**TL;DR:** E1–E7 (every safe-value tier, 440 min of planned work) are DONE,
race-green and lint-clean, committed. E8's decision brief is recorded; the
owner verdict was being asked when this report interrupted it. E9–E15 not
started. One environment crisis (actively-corrupting shared module cache) was
contained hermetically but NOT repaired for other users.

---

## a) FULLY DONE (verified green: `-race -count=1` suite, `golangci-lint` 0 issues, integration suite green)

### E1 — Drainer seam integration (the 1% → 51% tier) — commit `b335413` + daemon-carried parts
- `inFlightTracker.Drain` exported (implements `system.Drainer`); registered via
  `sys.RegisterDrainer(inFile)` in `buildSystem`; `EventService.inFile` field deleted.
- `Shutdown` collapsed to: idempotence guard → `sys.Drain(ctx)` → **always** `sys.Close()` →
  `errors.Join`. **Plan deviation (deliberate, documented):** the plan said "delegate to
  `GracefulClose`", but system v4.10.2's `GracefulClose` **returns without closing when its
  drain phase fails — contradicting its own godoc** (verified in the pinned module AND on
  go-cqrs-lite master). Blind delegation would have failed the E1.4 exit criterion; Drain+Close
  guarantees it.
- New test `TestEventService_Shutdown_DrainTimeoutStillCloses`: parked in-flight command +
  expired drain ctx → timeout error surfaces AND the registered closer still ran (no engine
  leak). Passes under `-race`.
- Godocs updated on `Shutdown` + tracker (incl. why not bare GracefulClose).

### E2 — Drain + checkpoint regression tests — commit `41856f7`
- Restart-resume persistence already pinned by the existing
  `TestEventService_DefaultCheckpointStore_PersistsAcrossRestart` (kept as the regression proof).
- NEW `TestEventService_CheckpointStoreOverrideWins` + `countingCheckpointStore`: proves a
  consumer-supplied `CheckpointStore` actually saves/loads (guards the E3 default switch).
- Integration suite re-run green — no drain-semantics reconciliation was needed (E2.4).

### E3 — Checkpoint simplification — landed via daemon commit `1ed6156`
- Aux `*sql.DB` now opens **only** for the default DLQ; `wantDefaultCP` logic,
  `applyCheckpointSchema`, and the `storage/v4/eventstore` import deleted (~50 LOC net);
  `buildSystem` signature simplified; `storage/v4` became `// indirect` after `go mod tidy`.
- Restart-resume test passes through system's engine-backed default (ADR-0142,
  `system_checkpoints`) — persistence proven end-to-end.
- `TestNewEventService_ConfigPath` / `TestEventService_DB` updated to enable the DLQ
  (the aux DB's only remaining purpose); `DB()`, `auxDSN`, `CheckpointStore` godocs corrected.

### E4 — Upgrade notes + escape hatch — landed via daemon commits
- `cqrs/CHANGELOG.md`: `[Unreleased]` section with both Changed entries (drain path +
  checkpoint migration). **Also fixed a pre-existing defect:** the v0.6.1 release had
  inserted its block ABOVE the `# Changelog` title (another session's bug).
- README `## Upgrading from v0.6.x`: `checkpoints` table → `system_checkpoints` collection,
  one-time replay is safe (derived read models), `DB()`/`Shutdown` consequences, escape hatch.
- Escape-hatch snippet **compile-checked** in a scratch module against the real tree
  (`/tmp/escape-hatch`, local replace) — compiles.

### E5 — Health/SCREAM accessors — commit `964f1e6`
- New `cqrs/healthaccessors.go`: `HealthCheck(ctx)`, `EngineHealth(ctx) []system.EngineHealth`,
  `ScreamReport() *system.ScreamReport` (delegations, godoc'd).
- `TestEventService_HealthAccessors` green (sqlite svc: nil error, non-empty named engines,
  non-nil report).
- README Accessors table rows + **go-health wiring snippet** — compile-checked
  (`/tmp/health-bridge`, both replaces) — compiles (pre-clears E12.2).

### E6 — Construction-time SCREAM surfacing — commit `180be9a`
- `logScreamFindings` in `NewEventService`: WARN+OVERRIDE → WARN (with the
  `acknowledge_warnings` escape hatch named), ADVISORY → INFO, SCREAM → ERROR (defensive);
  clean report logs nothing; nil Logger falls back to `slog.Default()`.
- **Plan deviation (reality-driven):** the plan's E6.2 expected a scream rule for "deployment
  without RoleProjections" — no such rule exists in v4.10.2. Tests use the real rules:
  `TestNewEventService_LogsScreamWarningsForVolatileSourceOfTruth` (captured-log assertion)
  and `TestNewEventService_CleanDeploymentLogsNoScream` (empty buffer). Both green.

### E7 — README deployment-shapes section — landed via daemon commits `8d66e89`/`0a64aab`
- `cqrs/testdata/deployment-{buses,priority-views,manifest}.yaml` + `deployment_test.go`:
  every snippet round-trips through the pinned `system.LoadConfig` (the exact parser
  `ConfigPath` uses) with field-level spot assertions (publish fan-out count, perQuery
  priority, materialized view fields, mixed pools, manifest path, acks, cache capacity).
- README `### Deployment shapes` covers buses/publish, priorities/materialized
  views/durability/pools, manifest/acknowledge_warnings/cache — each snippet labeled with its
  testdata source of truth. Stale `CheckpointStore` row in the config table fixed.

### Environment crisis contained (not repo work, but the session's biggest fight)
- The shared `/mnt/buildcache` module cache corrupted mid-session at scale (empty `.go` files
  across otel v1.47.0, stdr, backoff, auto/sdk, shortuuid, modernc/libc; later even stdlib and
  the toolchain — `go tool compile: segmentation fault`). Files verified intact on disk one
  second, read as EOF the next; concurrent `go build`/`go mod tidy`/multiple gopls/buildflow
  instances race on that disk. Disk is NOT full.
- **Working hermetic recipe (session-scoped):** `/tmp/gorun.sh` → private
  `GOMODCACHE=/tmp/cqrs-gomod`, `GOPROXY=file:///mnt/buildcache/.../cache/download,file:///home/lars/go/pkg/mod/cache/download`
  (zips intact; only 2 zips were missing and existed in the default cache), `GOSUMDB=off`,
  `GOWORK=off`, direct `/tmp/go-toolchain-1.27.1/bin/go` binary (GOTOOLCHAIN switching fights
  GOSUMDB=off; golangci-lint needs the toolchain `bin/` prepended to PATH).
- All green verifies above ran under this recipe.

---

## b) PARTIALLY DONE

### E8 — Owner gate (brief done, verdict NOT recorded)
- E8.1 DONE: decision brief appended to the TODO_LIST cqrs item (line ~47): demand evidence
  (zero cqrs+realtime composition consumers; W5-C2 first projected consumer), additive-only
  proposal with pinned merge rules, cost ~2 focused days, accepted v5-rename exposure,
  GO/NO-GO/GO-on-W5-C2 checkboxes.
- E8.2 INTERRUPTED: the structured owner question was fired and interrupted by this status
  request. **Verdict still ⬜⬜⬜ — E9–E13 are blocked on it.**
- Related drift to absorb: the plan's E14 baseline said `cqrs/v0.6.0`; another session tagged
  **`cqrs/v0.6.1` today** (dependency sweep). The API-break check must diff against v0.6.1.

---

## c) NOT STARTED (blocked or queued)

- **E9** `EventConfig.Domain` seam + merge rules (GO only)
- **E10** Domain tests + compile-checked runnable example (GO only)
- **E11** `WithCommandLifecycle` recipe (GO only)
- **E12** go-health bridge recipe — the snippet is ALREADY compile-checked (E5); README
  section + cross-links remain (GO only)
- **E13** ServeSSE read-model example + stretch integration E2E (GO only)
- **E14** Release train `cqrs/v0.7.0` (7 micro-tasks; baseline v0.6.1; runs on either verdict —
  on NO-GO it ships E1–E7 alone)
- **E15** Post-train bookkeeping (AGENTS Release State + dep-table `storage/v4 // indirect`,
  TODO_LIST close-out, `go-structure-linter` AGENTS ≤ 377)

---

## d) TOTALLY FUCKED UP

1. **The shared `/mnt/buildcache` is still corrupt for everyone else.** My fix is private
   (`/tmp`); gopls in this editor remains broken (segfaults on the corrupted cache — every
   LSP diagnostic this session was environment noise, not code). Root cause unproven
   (concurrent-extraction race vs. multi-machine shared disk). NOT repaired beyond deleting
   my 5 corrupted module dirs.
2. **4 of 8 epic commits carry generic daemon messages** (`chore: auto-commit … heuristic`):
   the E1 core edits, E3, E4, and E7 all lost the race against the auto-commit daemon despite
   immediate staging. The semantic story survives only in the CHANGELOG — git history
   readability for those four is degraded. (E2/E5/E6 + the E1 final fix carry proper messages.)
3. **TODO_LIST inaccuracy I introduced:** the E8 brief says "upstream ask drafted" for the
   GracefulClose doc/impl mismatch — **no draft file exists yet** (only the intent). Either
   draft `doc/feedback/outgoing/2026-10-06_upstream-ask-gocqrslite-gracefulclose.md` or fix
   the wording.
4. **Repo ritual gap:** "after adding a cqrs wrapper feature: re-run `cqrs-lint` scorecard and
   record the delta in the CHANGELOG" — NOT done for the E1–E7 wrapper changes. The sweep
   must run before the v0.7.0 tag.
5. **CHANGELOG `[Unreleased]` has no `### Added` section yet** — health accessors + scream
   logging are user-facing additions, not just Changed entries. Must land before dating.

---

## e) WHAT WE SHOULD IMPROVE

- **Persist the hermetic recipe** (private GOMODCACHE + dual file:// GOPROXY + direct
  toolchain binary) into project memory — `/tmp/gorun.sh` dies with the machine and the next
  session will re-derive it the hard way.
- **Beat the daemon:** `git add <files> && git commit` in ONE bash call immediately after the
  verify step (worked for E2/E5/E6; splitting the calls lost E1/E3/E4/E7).
- **Point Crush's LSP at the private cache too** (or disable gopls during corrupted-cache
  incidents) — the diagnostics spam drowned real signal all session.
- **File the upstream GracefulClose bug** (go-cqrs-lite): godoc promises close-attempt on
  drain expiry; implementation returns early. Verified on v4.10.2 AND master. Follow
  verify-before-filing + github-voice when drafting.
- **Plan-vs-reality deviations worked** (GracefulClose → Drain+Close; E6.2's nonexistent
  scream rule → real rules) but were only documented in code/CHANGELOG — future plans should
  require a deviation note in the plan file itself when an exit criterion contradicts the
  specified mechanism.
- **Question-gate placement:** asking the E8 verdict mid-execution stalls the whole run —
  next time, front-load owner gates BEFORE Full Execution Mode starts.

---

## f) NEXT — up to 50 things (ordered; #1 is blocking)

| # | Task | Gate |
|---|------|------|
| 1 | **E8.2: record owner verdict** (GO / NO-GO / GO-on-W5-C2) in the TODO_LIST brief | BLOCKING |
| 2 | E9.1 add `Domain *system.DomainConfig` field + merge-contract godoc | GO only |
| 3 | E9.2 `mergeDomain`: bootstrap appended ONLY when consumer declares no projections | GO only |
| 4 | E9.3 middleware merge `[inFlight] + Domain.Middleware + CommandMiddleware` | GO only |
| 5 | E9.4 merge HostOptions/CheckpointStore (derived wins); passthrough Events/Timers/Evolutions/decoders/ShutdownDependencies | GO only |
| 6 | E9.5 prove nil-Domain path unchanged (zero test edits, suite green) | GO only |
| 7 | E9.6 lint pass (exhaustruct nolints, wrapcheck delegations) | GO only |
| 8 | E9.7 `cqrs-lint` scorecard + CHANGELOG delta | GO only |
| 9 | E10.1 QuerySet → `system.Find` filtered+sorted round-trip test | GO only |
| 10 | E10.2 coeffect gate test (`ErrDanglingEventSubscription` on undeclared event) | GO only |
| 11 | E10.3 bootstrap-skip test (consumer declares ≥1 projection) | GO only |
| 12 | E10.4 `Evolve`/`Lookup` + typed `Get` point-read test | GO only |
| 13 | E10.5 runnable godoc example with verified output | GO only |
| 14 | E10.6 hermetic race suite + README quickstart scratch-compile | GO only |
| 15 | E11.1 `WithCommandLifecycle` recipe README section | GO only |
| 16 | E11.2 compile-check the recipe snippet | GO only |
| 17 | E11.3 cross-link from TODO_LIST cqrs item | GO only |
| 18 | E12.1 go-health bridge README section (snippet already compiles) | GO only |
| 19 | E12.3 optional cross-link in health module README | GO only |
| 20 | E13.1 `metaengine.NewWatcher` + `ServeSSE` example | GO only |
| 21 | E13.2 example test: httptest client asserts streamed events after dispatch | GO only |
| 22 | E13.3 README note: raw-event vs materialized read-model streaming | GO only |
| 23 | E13.4 stretch: integration E2E through an appkit Service (`NoTimeout`) | GO only |
| 24 | E4-completion: add CHANGELOG `### Added` (accessors, scream logging, [Domain]) | either |
| 25 | Ritual-debt: run `cqrs-lint` scorecard for E1–E7, record delta in CHANGELOG | either |
| 26 | Draft upstream ask: system GracefulClose doc/impl mismatch (or fix TODO_LIST wording) | either |
| 27 | E14.1 API-break check: `git archive cqrs/v0.6.1` vs worktree `go doc -all` diff | either |
| 28 | E14.2 date CHANGELOG `[Unreleased]` → `[0.7.0] - 2026-10-07`; hermetic verify cqrs + integration | either |
| 29 | E14.3 integration re-pin `cqrs v0.7.0` + `documentedPins` fixture + `go mod tidy` (GOWORK=off) | either |
| 30 | E14.4 guards: `check-pin-drift.sh` + `check-go-directives.sh` green | either |
| 31 | E14.5 annotated tag + `pre-tag-checks.sh cqrs/v0.7.0` | either |
| 32 | E14.6 push master + tag; fresh-consumer proxy check (needs networked machine) | either |
| 33 | E14.7 record adoption-closure outcome + scorecard delta in CHANGELOG | either |
| 34 | E15.1 AGENTS: Release State + cqrs bullet (v0.7.0, new surface, storage/v4 → indirect) | either |
| 35 | E15.2 TODO_LIST: close the P2 cqrs item per verdict; carry NO-GO remainder to watchlist | either |
| 36 | E15.3 `go-structure-linter` — AGENTS ≤ 377 counted lines | either |
| 37 | Re-verify integration suite once more before the tag | either |
| 38 | Update AGENTS cqrs dependency table rows for the tidy (storage/v4 indirect) | either |
| 39 | Persist the hermetic build recipe to project memory | either |
| 40 | Decide: repair `/mnt/buildcache` (re-extract corrupted modules) or escalate to disk owner | either |
| 41 | Configure Crush LSP against the private cache / silence broken gopls | either |
| 42 | Annotate the SUPERB plan file with execution state (docs-health ANNOTATE, non-destructive) | either |
| 43 | Tick the verdict checkbox in the TODO_LIST E8 brief after the owner answers | either |
| 44 | Clean up or document the /tmp scratch modules (escape-hatch, health-bridge) for reuse in E10–E13 | GO only |
| 45 | Re-check `Shutdown` godoc wording after any E9 Timers decision (timers stop only in GracefulClose Phase 0 — Drain+Close leaves them if Domain.Timers lands) | GO only |
| 46 | Consider a wrapper test asserting `sys.ShutdownOrder()` sanity after Domain passthrough | GO only |
| 47 | After tagging: re-pin check that no consumer resolves a broken v0.7.0 (proxy smoke) | either |
| 48 | Post-release: docs-health HARVEST from this report's section (f) into TODO_LIST | either |

---

## g) QUESTIONS I CANNOT ANSWER MYSELF (max 3)

1. **E8 verdict (blocking E9–E13):** GO (build the `EventConfig.Domain` seam now, ship in
   v0.7.0) / NO-GO (v0.7.0 ships E1–E7 only; revisit on first real consumer) /
   GO-on-W5-C2 (defer until the folded-contract demand lands)?
2. **Push timing:** master is 25 commits ahead of origin (mine + other sessions' work
   interleaved). Push now, or hold everything for the E14.6 train push (master + tag together)?
3. **`/mnt/buildcache` ownership:** is that disk shared across your machines (corruption
   smells like multi-writer), and do you want me to attempt a full re-extraction repair for
   all corrupted modules, or is another owner/session already on it?
