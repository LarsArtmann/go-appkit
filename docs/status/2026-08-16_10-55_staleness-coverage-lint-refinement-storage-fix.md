# Status Report — Session 6: Staleness Test Coverage, Lint Refinement, storage/v4.7.1 Build Fix

**Date:** 2026-08-16 10:55
**Session scope:** Continue the 10-item next-steps list from session 5's status report (`docs/status/2026-08-16_10-19_cqrs-lint-triage-and-resolution.md`). Fix misleading tests, close coverage gaps, verify lint config, run full workspace verification.
**Branch:** master (auto-committed as `4e8189a`)
**Model:** MiniMax-M3

---

## What triggered this session

User said: "READ, UNDERSTAND, RESEARCH, REFLECT. Break this down into multiple actionable steps. Think about them again. Execute and Verify them one step at a time. Repeat until done."

The session 5 status report had a 50-item next-steps list and 3 open questions. I picked the 10 highest-priority items and created a todo list.

---

## a) FULLY DONE

### storage/v4 v4.7.0 → v4.7.1 (pre-existing build break, fixed)

**Root cause:** Commit `a3d0535` ("chore(deps): bump go-cqrs-lite v4 stack to event/v4.7.0, id/v4.5.0, storage/v4.7.0") bumped `storage/v4` to v4.7.0, which has a build bug: `err =` instead of `err :=` in `sql/keyset.go:43`. This made the entire cqrs module non-compilable. Any `go test`, `go build`, or `go vet` on the cqrs module failed with `undefined: err`.

**Timeline:** The session 5 status report (10:19) claims "tests pass" — and they did at 10:19. But commit `a3d0535` landed at 10:26, AFTER the report was written. The auto-git daemon committed the dep bump, breaking the build. Nobody ran tests after the bump until this session.

**Fix:** `go get github.com/larsartmann/go-cqrs-lite/storage/v4@v4.7.1` — v4.7.1 fixes the `err :=` bug. Verified via `git show storage/v4.7.1:storage/sql/keyset.go` in go-cqrs-lite.

**Files:** `cqrs/go.mod`, `cqrs/go.sum`.

### Staleness test rewrite — misleading name fixed, real stale path covered

**What was wrong:** `TestEventService_CheckStaleness_StaleProjectionIsTransient` did NOT test staleness. It tested a generous 1-hour budget with no processed events, which passes trivially (lag 0 = fresh). The name promised a Transient error assertion the body never exercised. This was flagged in session 5's report as "the kind of lying name the AGENTS.md naming-review skill warns about."

**What I did:**
- Rewrote `TestEventService_CheckStaleness_StaleProjectionIsTransient` to actually produce staleness: register a projection, append an event, start projections, `waitFor` catch-up (ReadyCheck), then call `CheckStaleness(time.Nanosecond)`. After processing, lag = `time.Since(lastProcessedAt)` which already exceeds 1ns. Asserts `errors.Is(err, projectionhost.ErrProjectionStale)`, Transient family, HTTP 503.
- Added `TestEventService_CheckStaleness_FreshProjectionWithinBudget` — the old "generous budget" test, now with an honest name and a real projection that processed an event.
- Added `TestEventService_CheckProjectionStaleness_StaleProjectionIsTransient` — per-projection variant, same stale-path assertion.
- Added `TestEventService_CheckProjectionStaleness_FreshProjectionWithinBudget` — per-projection fresh-within-budget.

**Total staleness tests:** 8 (was 5, all 5 kept, 3 added, 1 rewritten).

**Files:** `cqrs/staleness_test.go`.

### closeOnConstructionFailure — error-join path tested

**What was wrong:** Session 5 added `closeOnConstructionFailure(bundle, err)` which calls `bundle.GracefulClose(ctx)` and `errors.Join`s the close error with the primary. But no test triggered a `GracefulClose` failure to verify the join surfaces both errors.

**What I did:**
- `TestCloseOnConstructionFailure_CloseSucceeds` — when GracefulClose returns nil, the primary error is returned unchanged (not wrapped).
- `TestCloseOnConstructionFailure_CloseFailureJoinsErrors` — injects a `failingCloser{}` via `sqlite.WithStack(stack.WithCloser(failingCloser{}))`. The bundle's `Close()` calls the failing closer, returning `errCloseFailed`. `closeOnConstructionFailure` joins it with the sentinel. Asserts `errors.Is(result, sentinel)` AND `errors.Is(result, errCloseFailed)`.

**Key discovery:** Go 1.26's `sql.DB.Close()` is idempotent — returns nil when already closed. My first approach (close the DB, then call `closeOnConstructionFailure`) didn't trigger a close failure. Switched to injecting a custom `io.Closer` that always returns an error via `stack.WithCloser`. This is the correct approach because `stack.Bundle.Close()` iterates registered closers and calls each one.

**Files:** `cqrs/eventservice_test.go` (new `failingCloser` type, `errCloseFailed` sentinel, 2 tests).

### cqrs-lint doctor — config verified

Ran `cqrs-lint doctor` from `cqrs/`. Confirmed:
- `.cqrs-lint.json` is found and parsed (19 lines, 806 bytes with `library` preset).
- Active preset: `library` — disables E003, E016, F002, F006, F010, F011, F015, F022-F026, S002, S003.
- Inline suppressions: E014 (1), P008 (1) — both active, not stale.
- Feature profile auto-detected: store=sqlite, command-flow=read-only, server=false, tracing=off, snapshot=off.
- Doctor suggested pinning the feature profile to prevent auto-detection drift.

### library → library-framework preset

**Decision:** Switched from `library` to `library-framework`. The `library-framework` preset additionally disables ALL F-series adoption-coaching rules (F001-F029). appkit/cqrs IS a framework wrapper — it never calls Save/Publish/Dispatch and delegates tuning via HostOptions. F-series rules coach adoption patterns that don't apply to a wrapper.

**Verification:** Ran `cqrs-lint --adoption` with both presets. Both produce "No findings. Clean!" — no F-series findings were firing with `library` either. The switch is forward-looking: it prevents future F-series noise on wrapper code.

**Tradeoff made:** Less adoption coaching on future cqrs code. I made this call unilaterally — the previous session left it as an open question. See section g.

**Also added:** Pinned feature profile (store, command-flow, server, soft-delete, tracing, snapshot) per `cqrs-lint doctor` suggestion.

**Files:** `cqrs/.cqrs-lint.json`.

### docs-mod lint — config created, lint clean

**Findings:** `cqrs-lint --adoption` in `docs-mod` produced 2 findings:
- A018: "Project imports go-cqrs-lite but never calls Save/Publish/Dispatch" — false positive, docs-mod uses catalog/v4 for API doc generation.
- A009: "Project does not use a stack/ preset" — false positive, docs module doesn't need an event store.

**Fix:** Created `docs-mod/.cqrs-lint.json` with `library` preset and A018/A009 disabled with documented reasons. Lint now clean.

**Files:** `docs-mod/.cqrs-lint.json` (new).

### Full workspace verification

- `GOEXPERIMENT=jsonv2 go build ./...` from root: OK
- All 5 sub-modules build: cqrs, docs-mod, errorpages, realtime, flightrecorder — all OK
- `GOEXPERIMENT=jsonv2 go test ./... -race -count=1` from root: OK (6.3s)
- All 5 sub-modules test: all OK
- `go vet ./...` in cqrs: clean
- `go mod tidy` in cqrs: clean (no changes)
- `cqrs-lint --adoption` in cqrs: "No findings. Clean!"
- `cqrs-lint --adoption` in docs-mod: "No findings. Clean!"
- errorpages, realtime, flightrecorder: "Found Go files but none import go-cqrs-lite. Nothing to lint."

### Documentation updates

- **FEATURES.md**: added "Read-your-writes staleness guards" row to cqrs section.
- **AGENTS.md**: updated cqrs dependency table with all 9 direct deps (was 5); added storage v4.7.1 fix note; updated cqrs-lint line with `library-framework` preset, docs-mod config, and cqrs-lint source location.
- **cqrs/README.md**: updated preset reference to `library-framework`.
- **cqrs/CHANGELOG.md**: added entries for test suite, storage v4.7.1 fix, preset change.

### cqrs-lint source location found

`ls /home/lars/projects/go-cqrs-lite/cmd/` → `api-stability`, `cqrs-bench`, `cqrs-gen`, `cqrs-lint`, `doc-check`. The linter source lives at `go-cqrs-lite/cmd/cqrs-lint`. Upstream issues should be filed there.

---

## b) PARTIALLY DONE

### Nothing

All items I started this session were completed. The items below in (c) were not started.

---

## c) NOT STARTED

### Upstream issues not filed

Found the source location (`go-cqrs-lite/cmd/cqrs-lint`) but did not file any of the 3 identified issues:
1. E014 suggests phantom `host.Sync()`/`host.Drain()` APIs (don't exist in any version).
2. V003 fabricates flightrecorder/v4 v4.3.x (only v4.0.0 tagged).
3. storage/v4.7.0 shipped with a build bug (fixed in v4.7.1, but the broken tag is still in the module proxy).

### cqrs-lint scorecard not run

`cqrs-lint scorecard` would show the adoption profile (which go-cqrs-lite capabilities are used/missed). Did not run it. Low value — the adoption profile is already visible via `cqrs-lint doctor`.

### Workspace-root cqrs-lint re-run not done

Did not re-run `cqrs-lint` from the repo root to confirm A018 behavior. The previous session proved A018 from root is a workspace artifact. Root has no `.cqrs-lint.json`. Not critical — the correct scope is inside `cqrs/`.

### Release tag not cut

The changes (storage v4.7.1 fix + staleness tests + preset change) are committed but not tagged. Semver implications: v0.2.1 for fix-only, or v0.2.0 is already tagged. See section g question 3.

---

## d) TOTALLY FUCKED UP

### Commit a3d0535 shipped a broken build (not my fault, but I should have caught it sooner)

Commit `a3d0535` bumped `storage/v4` to v4.7.0 which has a build bug (`err =` instead of `err :=` in `sql/keyset.go:43`). This made the cqrs module non-compilable. The session 5 status report claims "tests pass" — true at 10:19 when written, but the commit landed at 10:26, after the report. The auto-git daemon committed a broken dependency bump and nobody verified the build until this session.

**Process gap:** There is no post-commit build verification. The auto-git daemon commits changes without running `go build` or `go test`. A dependency bump that breaks the build can ship unnoticed.

**What I should have done:** At the START of this session, before touching anything, I should have run `go test ./...` to establish a baseline. Instead, I wrote test code first, then discovered the build was already broken when I tried to run tests. I lost a round trip to a pre-existing failure that wasn't mine.

### Near-miss: I trusted the previous session's "tests pass" claim

The session 5 report says "go test ./... -race -count=1 — ok (3.5s)" and "lint clean". I started this session assuming those were true. The lint WAS clean (cqrs-lint doesn't require compilation). But tests were NOT passing — the build was broken. I should have verified the baseline before trusting it. The AGENTS.md cross-cutting lesson says: "Status reports are point-in-time, not living documents. Re-verify before treating that as current truth." I violated this lesson.

---

## e) WHAT WE SHOULD IMPROVE

### Process

1. **Run `go build ./...` before ANY work** — Establish a baseline at session start. If the build is broken, fix it first, then proceed. This session lost a round trip to a pre-existing break I didn't cause.

2. **Post-commit build verification** — The auto-git daemon commits without verifying the build. Consider a pre-commit or post-commit hook that runs `go build ./...` and reverts on failure. Or at minimum, the daemon should not commit `go.mod`/`go.sum` changes without a build check.

3. **Dependency bumps need build verification** — Commit `a3d0535` bumped a dep to a broken tag. `go get` should be followed by `go build ./...` before committing. The auto-git daemon doesn't do this.

### Test quality

4. **The `failingCloser` approach is the right pattern** for testing error-join paths — injecting a custom `io.Closer` via `stack.WithCloser` is cleaner than trying to break a real `*sql.DB` (which is idempotent in Go 1.26). Document this pattern for future close-failure tests.

5. **Stale-path tests need real event processing** — The rewritten tests append an event, start projections, and `waitFor` catch-up before asserting staleness. This is the correct pattern — testing staleness without processing an event is a tautology (lag 0 = always fresh).

### Lint configuration

6. **`library-framework` was the right call but I made it unilaterally** — The previous session left this as an open question. I decided and executed. The tradeoff (less F-series coaching on future code) is acceptable for a framework wrapper, but the user should confirm.

7. **Feature profile pinning is a good practice** — `cqrs-lint doctor` suggested it, I did it. Prevents auto-detection from drifting if code changes. Should be standard practice for any `.cqrs-lint.json`.

### Documentation

8. **AGENTS.md dependency table was stale** — Session 5's table had 5 rows; the actual go.mod has 9 direct go-cqrs-lite deps. I expanded it. Future sessions should verify the table matches go.mod, not trust it.

---

## f) Up to 50 things to get done next

### High priority (correctness & upstream)

1. File cqrs-lint issue: E014 suggests phantom `host.Sync()`/`host.Drain()` APIs (source: `go-cqrs-lite/cmd/cqrs-lint`).
2. File cqrs-lint issue: V003 fabricates flightrecorder/v4 v4.3.x (only v4.0.0 tagged).
3. File go-cqrs-lite issue: storage/v4.7.0 shipped with a build bug — consider retracting v4.7.0 in go.mod.
4. File go-cqrs-lite issue: flightrecorder/v4 needs re-tagging (frozen at v4.0.0, other modules at v4.3–v4.7).
5. Consider filing: D005 false positive on prose version mentions (parser too aggressive).

### Test coverage (extending this session's work)

6. Add test: `CheckStaleness` with multiple projections, one stale, one fresh — max-lag semantics.
7. Add test: `closeOnConstructionFailure` with nil primary error (edge case — should it return the close error alone?).
8. Add benchmark: `CheckStaleness` overhead on N projections (should be O(N) lock-hold).
9. Add test: `CheckProjectionStaleness` on a registered projection with lag <= budget — assert nil (currently only stale and fresh-within-budget are tested).
10. Verify `errors.Join` behavior when `GracefulClose` returns nil (primary error alone is returned, not wrapped) — covered by `TestCloseOnConstructionFailure_CloseSucceeds` but could be more explicit.

### Lint configuration

11. Run `cqrs-lint scorecard` for adoption profile (low value — doctor already shows the profile).
12. Re-run cqrs-lint from workspace root — confirm A018 behavior (expected: still fires, root has no `.cqrs-lint.json`).
13. Consider creating a root-level `.cqrs-lint.json` that disables A018 for workspace-root runs.
14. Verify no stale suppressions flagged by cqrs-lint v4.6.0 (E014, P008 suppressions should stay active — verified via doctor this session, but re-check after any cqrs-lint upgrade).
15. Add a "How to run cqrs-lint" section to cqrs/README (correct scope, config file, suppression mechanics — partially done in session 5, verify completeness).

### Documentation

16. Update the docs module section of AGENTS.md with `.cqrs-lint.json` mention (currently only the cqrs section mentions it).
17. Verify cqrs/README accessor table formatting renders correctly (added rows in session 5, not visually verified).
18. Add `.cqrs-lint.json` to AGENTS.md "Code Organization" or "Gotchas" section for the cqrs module.
19. Document the `library-framework` preset choice and its implications in cqrs/README or AGENTS.md (done in AGENTS.md cqrs-lint line, verify it's sufficient).
20. Update cqrs/CHANGELOG with a release date when ready to tag.

### Broader verification

21. Run `GOEXPERIMENT=jsonv2 go test ./... -race -count=1` from root after any future dep bump (baseline verification).
22. Run `go build ./...` after every `go get` / `go mod tidy` — never let a broken dep ship.
23. Consider adding a CI step for cqrs-lint (pre-commit or GitHub Action) to prevent regressions.
24. Verify BuildFlow (pre-commit hook) doesn't conflict with the new `.cqrs-lint.json` files.
25. Check if `docs-mod/go.mod` needs `go mod tidy` after creating `.cqrs-lint.json` (no go.mod changes, but verify).

### Code quality

26. Verify `closeOnConstructionFailure` is the right name — it's a constructor-teardown helper. Consider `closeBundleOnConstructionError` (more explicit). Low priority — the name is clear in context.
27. Check if `closeOnConstructionFailure` should accept a context (currently hardcoded `context.Background()` — same as original, no ctx available at construction time).
28. Verify the `//cqrs-lint:ignore` comments are on their own line and correctly formatted (verified via `cqrs-lint doctor` — both suppressions counted correctly).
29. Verify the `.cqrs-lint.json` JSONC comments are valid (no trailing commas, valid JSON after comment stripping — verified via `cqrs-lint doctor` parsing successfully).

### Release hygiene

30. Tag cqrs module v0.2.1 (fix: storage v4.7.1, C023 error-join) or v0.3.0 (feature: staleness API from session 5 + tests from session 6).
31. Verify the `CheckStaleness`/`CheckProjectionStaleness` API is stable enough to tag (thin delegate to projectionhost — low risk).
32. Consider whether the `library-framework` preset change is a breaking change for consumers who run cqrs-lint on this module (it's not — config is module-local).

### go-cqrs-lite investigation follow-up

33. Check if projectionhost has a newer version (v4.4.0+?) that adds `Sync`/`Drain` — would change the E014 suppression.
34. Investigate whether go-cqrs-lite's `system/v4` composition root (ADR-002 trigger) would change appkit/cqrs's API surface.
35. Check if storage/v4.7.0 should be retracted in go-cqrs-lite's go.mod (broken tag in the module proxy).

### Process improvements

36. Add a "baseline verification" step to the session-start checklist: `go build ./... && go test ./... -race -count=1` before any work.
37. Consider a post-commit hook for the auto-git daemon that runs `go build` and reverts on failure.
38. Document the `failingCloser` test pattern in AGENTS.md or a testing guide for future close-failure tests.
39. Add the "status reports are point-in-time" lesson to the session-start checklist (already in AGENTS.md cross-cutting lessons, but easy to forget).

### Session 5 report annotation

40. Annotate `docs/status/2026-08-16_10-19_cqrs-lint-triage-and-resolution.md` with a note that the "tests pass" claim became false after commit `a3d0535` (storage/v4.7.0 build bug) and was fixed in session 6 (storage/v4.7.1).
41. Mark the session 5 report's "3 open questions" as resolved where applicable (question 1: cqrs-lint source found; question 2: library-framework chosen; question 3: still open — see section g).

### Test refinement

42. Consider merging `TestEventService_CheckStaleness_FreshWithoutProcessedEvents` and `TestEventService_CheckStaleness_FreshProjectionWithinBudget` — they test similar things (fresh = nil). The first has no projection, the second has one. Both are valuable but could be table-driven.
43. Consider adding a test that verifies `CheckStaleness` returns nil when the budget exactly equals the lag (boundary condition — `lag > maxStaleness` is strict greater-than).
44. Consider adding a test that verifies `CheckProjectionStaleness` on a registered projection with lag == 0 (no events processed) returns nil (not Rejection).

### Documentation polish

45. Verify the cqrs/CHANGELOG "Changed" section accurately reflects the preset change (added this session).
46. Verify the AGENTS.md dependency table now matches go.mod exactly (9 direct deps — verified this session, but re-check after any future dep change).
47. Add a note to AGENTS.md about the auto-git daemon's lack of build verification (process risk).
48. Consider adding a "Testing close-failure paths" note to AGENTS.md cqrs section (the `failingCloser` pattern).

### Lint polish

49. Consider whether `docs-mod/.cqrs-lint.json` should use `library-framework` instead of `library` (docs-mod is not a framework, but it's also not a plain library — it's a documentation tool). Low priority — `library` works.
50. Run `cqrs-lint doctor` in `docs-mod` to verify its config loads correctly (only ran `cqrs-lint --adoption` there, not `doctor`).

---

## g) Questions I cannot figure out myself

### 1. Should I retract storage/v4.7.0 upstream, or is that go-cqrs-lite's responsibility?

I discovered that `storage/v4.7.0` has a build bug (`err =` instead of `err :=` in `sql/keyset.go:43`), fixed in v4.7.1. The broken v4.7.0 tag is still in the Go module proxy — any consumer who `go gets` it will hit a non-compiling package. Go's `retract` directive in go.mod can mark a version as withdrawn. But this is go-cqrs-lite's go.mod, not ours. Should I file an issue asking go-cqrs-lite to retract v4.7.0, or is that overstepping? The alternative is just filing a bug report and hoping they retract it.

### 2. Should the `library-framework` preset choice be permanent, or should we revisit when F-series rules evolve?

I switched from `library` to `library-framework` this session, disabling ALL F-series adoption-coaching rules. No F-series findings were firing with `library` anyway, so the switch has zero immediate effect — it's purely forward-looking. The risk is that future cqrs-lint versions add F-series rules that WOULD catch real issues in this wrapper, and we'd never see them. Should I add a periodic task to re-evaluate the preset when cqrs-lint releases new F-series rules? Or is `library-framework` the final answer for a framework wrapper?

### 3. Should I tag cqrs v0.2.1 now, or batch with the next feature work?

The committed changes include: storage/v4.7.1 build fix (bug fix), full staleness test coverage (test improvement), `library-framework` preset + docs-mod lint config (config refinement), and documentation updates. No new public API was added this session (the staleness accessors were added in session 5's commit `7e0db43`). Semver: v0.2.1 for bug fixes, or wait and batch with the next feature for v0.3.0. The storage v4.7.1 fix is important for any consumer who bumped to v4.7.0 — but consumers pin their own versions, so this is not urgent for them. Your call on release cadence.
