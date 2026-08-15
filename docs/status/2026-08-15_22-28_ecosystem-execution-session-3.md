# Status Report — Ecosystem Execution, Session 3 (Tier 4 in flight)

**Date:** 2026-08-15 22:28
**Scope:** This session only (continuation of SUPERB-ECOSYSTEM-EXECUTION plan)
**Repo:** go-appkit (master, 13 unpushed commits) + cqrs-htmx (branch `feat/transport-package`, 1 unpushed commit)
**Plan:** `docs/planning/2026-08-15_19-27_SUPERB-ECOSYSTEM-EXECUTION.md`

## What I Did This Session (chronological, with commit IDs)

| # | Task                     | Commit                         | Result                                                                                                   |
| - | ------------------------ | ------------------------------ | -------------------------------------------------------------------------------------------------------- |
| 1 | Fix errorpages.Wrap bug  | `0eaea2f`                      | Wrap no longer swallows stdlib 307 path-cleaning redirects (`/a//b` → `/a/b`); parity tests vs bare mux  |
| 2 | M10 metrics              | `3cd9297`                      | `EventConfig.Metrics` → `projectionhost.WithMetrics`; 3 e2e tests incl. `/metrics` endpoint; README docs |
| 3 | M11 bridge (re-designed) | `ac743f30` (cqrs-htmx, branch) | `transport/` sub-package extracted: `JournalSSEStore` + deprecated root re-exports, 4-dep import surface |
| 4 | M11 realtime side        | `a19f080`                      | Handler subscribes BEFORE replay + ID dedup; 2 deterministic regression tests (gap + dedup)              |
| 5 | M12 context hygiene      | uncommitted                    | service.go `WithoutCancel`, docs_test/health_test `NewRequestWithContext`, realtime ctx threading        |
| 6 | cqrs lint debt cleanup   | uncommitted                    | 15/17 findings fixed (varnamelen `es`→`eventSvc`, `db`→`bundleDB`/`sqlDB`; noinlineerr; 1 noctx)         |

Daemon auto-commits this session: `83ebcbd`, `ce8b245` (dep drift).

## a) FULLY DONE ✅ (verified: test+vet+build+BuildFlow green)

1. **Wrap redirect bug (ledger item #1) cleared.** Root cause narrowed honestly:
   stdlib `cleanPath` PRESERVES trailing slashes, so `/health/` 404s even in
   the bare mux (pretty 404 = correct parity). The real bug: doubled slashes
   and dot segments (`/a//b`, `/x/../y`) return `(RedirectHandler, "")` from
   `mux.Handler` — empty pattern — which Wrap misread as 404. Fix: delegate
   when `cleanPath(escaped) != escaped`. Tests assert bare-mux parity
   (status + Location) and pretty-404 on the canonical follow-up.
2. **M10 metrics accessor.** Chose accessor-over-Setup (per plan's explicit
   choice point): `projectionhost.MetricsRecorder` interface in, zero new
   deps; no OTel-backed recorder exists upstream, so bundling prometheus/otel
   would tax all consumers. README documents the prometheus/v4 composition.
   Also backfilled missing README rows (FlightRecorder, HostOptions,
   ReadyCheck, LagPerProjection).
3. **M11 — user-redesigned mid-session.** Original plan (new go-appkit bridge
   module) was challenged by the user ("Why do we need to reinvent?").
   Verified facts, user chose: lean `transport/` sub-package INSIDE cqrs-htmx.
   - `transport/journalsse.go`: verbatim move, 2 fixes (WithMaxReplay doc now
     truthful `<=0 → default`; error codes `transport.sse.*`), 11 moved tests.
   - Root re-exports keep dashboardui + integration tests compiling.
   - dashboardui migration deferred (pins published root v4.8.0) with in-code
     note: migrate after next root release.
4. **realtime replay/live boundary (M11.4, done in go-appkit).** Old order
   (replay→subscribe) lost every event appended during the store read. New:
   subscribe-first + replayed-ID dedup in live loop. Strictly better than
   cqrs-htmx dashboardui (subscribe-first, no dedup). Honest limit documented:
   bursts >64 (subscriber buffer) during slow store reads can still drop;
   healed by Last-Event-ID reconnect. `Stream.ctx = r.Context()` verified in
   go-sse source before relying on it.
5. **M12 code changes complete but uncommitted** (see partial below): core
   `Run()` uses `context.WithoutCancel(ctx)` (same semantics, inherited);
   realtime handler threads ctx into replay logging; docs_test/health_test
   use `NewRequestWithContext`.

## b) PARTIALLY DONE ⏳

1. **M12 commit + sweep.** All code changes are in the working tree, verified
   green per-module (core, realtime, cqrs, docs-mod), but NOT committed; the
   plan's M12.3 all-module sweep (single pass incl. flightrecorder +
   errorpages) not yet run as one command.
2. **cqrs lint debt (session-2 ledger item #3).** 15 of 17 findings fixed
   (renames verified green under -race). Remaining: `metrics_test.go:185`
   `srv.Client().Get` → `Do(NewRequestWithContext)`.
3. **M11 loose ends.** cqrs-htmx branch not merged to master, not pushed;
   no root release containing `transport/` (blocks dashboardui migration);
   go-appkit side documents the composition path in AGENTS.md but ships no
   compile-tested consumer example of transport→realtime.Hub.

## c) NOT STARTED ❌

- **M13** docs refresh (README/FEATURES/CHANGELOGs; AGENTS.md still says five
  modules — errorpages missing; ledger item #4).
- **M14** templ-components FamilyOrchestration status-map fix (templ-components
  v1.8.3 now exists in module cache and cqrs-htmx already swept to it).
- **M15** releases (verify, CHANGELOGs, tags cqrs/v0.2.0 + docs/v0.2.0 +
  errorpages/v0.1.0, fresh-consumer smoke test).
- **M16** ADR-002 (EventService future: sqlite-first vs stack-generic vs
  `system`).
- **M17** cqrs README cookbook (scenario DSL, testutil, cqrs-lint).
- **M18** cqrs-htmx spike (appkit.Service behind `setup.Run`) — ADR-001
  approved; requires core SSE-safe WriteTimeout opt-out (P1) first.
- **Session-2 carry-over:** M09.4 render-failure-fallback test still not
  restored (ledger item #2).

## d) TOTALLY FUCKED UP 💥 (honest)

1. **Misdiagnosed my own Wrap bug scope.** Ledger claimed `/health/` vs
   `/health` triggered it; actually stdlib 404s `/health/` by design. Only
   `path.Clean`-rewritten paths redirect. Cost: one wrong test table + one
   doc-comment correction. Commit message states the corrected scope.
2. **Overstated cqrs-htmx's weight.** Claimed "single-package framework with
   casbin+ginkgo compiling in" — user called it out ("It is not??!").
   Truth: 21+ sub-packages exist; root package compiles ~22 external pkgs
   (no casbin — it's not in root deps; ginkgo is test-only). Root-cause of my
   error: reasoning from the AGENTS.md architecture section instead of
   running `go list -deps` first.
3. **Wrote a test against unverified store semantics.** My gap/dedup tests
   seeded cursor "1" not present in `memStore` → empty replay → false test
   failures and a 10-minute package timeout (hung `readSSEFrame`). Debug
   cycle wasted before checking `memStore.EventsAfter` (cursor must exist).
4. **Two malformed question-tool calls** (invalid `type`, then missing
   `description`) before the successful one.
5. **Mechanical renames committed-to-tree without immediate build.** Regex
   rename of `db`→`bundleDB` missed a reference in an error-format call
   (`eventservice.go:270`) — caught by the verification build minutes later,
   but it sat broken in the tree between edit and check.
6. **Question drift:** posed the delivery decision as 3 options (A/B/C);
   user invented a better 4th (D: transport/ sub-package). Should have
   offered the sub-package option myself — it was derivable from the facts.

## e) HOW TO IMPROVE (concrete)

- Run `go list -deps` / read the actual interface BEFORE making
  dependency-graph claims. AGENTS.md summaries are marketing, not source.
- Read the test-double's implementation (memStore cursor semantics) before
  writing tests against it. A hung SSE read is a 10-minute diagnosis tax.
- After ANY mechanical multi-file rename: build the package immediately,
  before touching anything else (this session proved the point twice).
- Offer the "move it inside the existing repo" option by default when an
  extraction is proposed — sub-package beats new module when the code's
  home already depends on everything the code needs.
- Verify question-tool payloads (type enum, description required) before
  sending.

## f) NEXT 50 (priority order)

1. Commit M12 + cqrs lint cleanup (tree is green; includes metrics_test noctx fix first)
2. M12.3 all-module sweep as one pass (6 modules, jsonv2 flags, -race)
3. Merge cqrs-htmx `feat/transport-package` → master (needs Q1 approval)
4. Push cqrs-htmx master (needs Q1)
5. Tag cqrs-htmx root release containing `transport/` (unblocks dashboardui)
6. Migrate dashboardui to `transport.NewJournalSSEStore` (code note marks it)
7. Bump errorpages to templ-components v1.8.3 (now in cache; was blocked)
8. M13.1 README: v4 migration note, new options, errorpages + transport story
9. M13.2 FEATURES.md + CHANGELOG.md for cqrs, docs-mod, errorpages, realtime
10. M13.3 AGENTS.md: six-module list, updated dep tables, transport gotchas
11. Restore M09.4 render-failure-fallback test (carry-over ledger #2)
12. M14.1 verify FamilyOrchestration exists in go-error-family v0.10.0 source
13. M14.2 reproduce missing mapping in templ-components v1.8.3
14. M14.3 fix + test in templ-components (own branch there)
15. M14.4 PR upstream (needs Q3 approval)
16. M15.1 pre-release verification: all modules test/vet/build green
17. M15.2 cut CHANGELOGs + version bumps
18. M15.3 tag `cqrs/v0.2.0`, `docs/v0.2.0`, `errorpages/v0.1.0`
19. M15.4 fresh-consumer `go get` smoke test (clean cache, GOWORK=off)
20. M16.1 evaluate go-cqrs-lite `system` package vs EventService fit
21. M16.2 write ADR-002 (sqlite-first vs stack-generic vs system)
22. M16.3 add chosen-path follow-ups to TODO_LIST.md
23. M17.1 cqrs README cookbook: scenario DSL usage
24. M17.1 cookbook: testutil leverage (fakes, harness)
25. M17.1 cookbook: cqrs-lint adoption (`library` preset reference)
26. M17.2 link cookbook from AGENTS.md
27. M18.0 core: SSE-safe WriteTimeout opt-out (ADR-001 P1) + tests
28. M18.1 spike: appkit.Service behind cqrs-htmx `setup.Run` (flag/branch)
29. M18.2 wire appkit drain probe ↔ `ProjectionReadinessCheck`
30. M18.3 verify SSE header flush survives appkit middleware chain
31. M18.4 smoke benchmark vs baseline server path
32. M18.5 adopt/reject report + follow-up tasks
33. Push 13 go-appkit commits (needs Q1)
34. errorpages README: document redirect-preserving Wrap semantics (new)
35. realtime README/AGENTS: document 64-buffer drop limit + heal semantics
36. go-appkit integration example: transport store → realtime.Hub (compile-tested)
37. docs/planning/integrations.md: add transport/ package entry
38. design-decisions.md ADR-001: note P1–P3 prerequisite status after M18
39. Fix go.work `use ./docs-mod` path-mismatch BuildFlow warning
40. Install go-licenses in devshell (BuildFlow preflight warning)
41. realtime: extract memStore/blockingStore into a testutil helper package
42. cqrs: coverage check after metrics/DLQ additions (project gate)
43. erraudit sweep on changed modules (skipped in pre-commit mode)
44. templ-components v1.8.3 CHANGELOG review for errorpage fixes
45. Add errorpages Wrap 405-Allow-header test (covered implicitly, make explicit)
46. Consider `WithMaxReplay(0)` true-unlimited upstream fix (transport doc vs code now truthful; decide semantics)
47. cqrs-htmx dashboardui: add dedup to its sseHandler (mirrors realtime fix)
48. Re-run full BuildFlow on final tree before tagging (warnings-only expected)
49. Sweep remaining `es`/`db` varnamelen (verify zero findings in cqrs)
50. Final status report + update this file's follow-ups

## Open Questions for the User

1. **Push/merge approval (blocking #3–5, #33):** May I merge
   `feat/transport-package` → master in cqrs-htmx and push? May I push the 13
   local commits on go-appkit master? Both repos are yours; nothing has been
   pushed.
2. **cqrs-htmx release timing (blocking #5–6):** tag a root release
   containing `transport/` now (e.g. v4.9.0) so dashboardui can migrate, or
   batch it with M15?
3. **Cross-repo PRs (blocking #15):** for M14, may I push a branch to
   templ-components and open a PR? (M18 spike stays local in cqrs-htmx unless
   you say otherwise.)
