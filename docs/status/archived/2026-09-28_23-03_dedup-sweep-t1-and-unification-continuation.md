# Status Report: art-dupl Sweep at -t 1 + Toolchain Unification Continuation

- **Date:** 2026-09-28 23:03 CEST
- **Scope of this report:** this session only — the `art-dupl -t 1` deduplication sweep and everything noticed along the way. No other research was done.
- **Trigger:** `art-dupl --sort total-tokens -t 1 --suggest-generics --timing --rich-text` → 9 shown clone groups (440 detected; 306 non-actionable, 125 filtered suppressed).
- **Series context:** this is episode 2 of the dedup series. Episode 1: `docs/status/2026-09-24_16-52_dedup-session-and-toolchain-drift.md` (read and cross-referenced; its items #10, #18, #27, #30, #47 and its questions 1/3 are directly entangled with this session).

---

## Stat Cards

| Metric                                             | Value                                                                               |
| -------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Actionable clone groups shown                      | 9 → **7** (both eliminated groups: real fixes; all 7 remaining: accepted-by-design) |
| Detected clone groups                              | 440 → 438                                                                           |
| Code removed by dedup                              | **−52 lines / +12** across 3 files (`d1dd7ea`)                                      |
| Modules verified green (vet + test -race + lint 0) | **3** (core, cqrs, integration)                                                     |
| Toolchain unification                              | 3 of 11 go.mods at 1.27.1 (root, health, **integration NEW**) + go.work             |
| Pre-existing red test fixed                        | 1 (`TestGoModGoDirectiveMatchesDocumentedToolchain`, red since the 1.27 bump)       |
| AGENTS.md length                                   | 376/377 cap (1 line headroom — still effectively full)                              |
| Suppressed groups ever inspected (both episodes)   | **0 of ~431**                                                                       |

---

## a) FULLY DONE

1. **All 9 clone groups read in source and triaged** — no site was accepted or changed from the report alone. Judgment: 2 real fixes, 7 accepted (rationale below).
2. **Collect-sort-key modernization (episode-1 item #27 executed):** `metrics.go:280` `sortedRoutes` and `cqrs/eventservice.go:405` `sortedEngineNames` hand-rolled map-key collect+`sort.Strings` → one-line `slices.Sorted(maps.Keys(...))`. `sort` import dropped from both files. Repo-wide grep confirms the pattern is now **fully accounted**: the only remaining `sort.Strings` is `flightrecorderhealth/adapter.go:251`, the accepted filtered-collect.
3. **testkit.Serve adoption in integration:** `newSSEService` (`integration/integration_test.go`) hand-rolled Start + 2s poll + 18-line cleanup that `testkit.Serve` already owns and that `security_realtime_test.go` already used. −23 lines; the two SSE E2E tests now also get the goroutine-baseline leak assertion on teardown (SSE hubs are exactly where leaks hide).
4. **Pre-existing red test turned green the sanctioned way:** `go mod tidy` (provoked by the testkit import) bumped `integration/go.mod` to `go 1.27.1`, tripping the module's own drift guard — which was **already red before this session** (`go 1.27` ≠ fixture `1.26.7`). Completed the guard's mandated 4-way update: `documentedGoDirective` → `1.27.1` in `pin_drift_test.go`, AGENTS.md toolchain note rewritten to the true state (go.work/root/health/integration at 1.27.1; 5 satellites lag).
5. **Accepted-clone rationale documented** in AGENTS.md ("Clone posture" bullet): options-pattern types + loops (idiom; sharing couples independent modules), example mains' bootstrap (self-contained demos), otel `Shutdown`'s four consecutive err checks (idiom), `failingServiceNames` (filtered collect, not a keys collect). Plus two standing rules: `slices.Sorted(maps.Keys(...))` for sorted map keys; integration tests start via `testkit.Serve`.
6. **AGENTS.md duplicate removed:** the "BuildFlow runs as pre-commit hook" line duplicated the more detailed Gotchas bullet — deleted, freeing cap headroom (377 → 376 with the new posture bullet added).
7. **Verification loop closed:** vet + `go test -race -count=1` green on core, cqrs, integration; `golangci-lint` 0 issues on all three (run sequentially with `GOTOOLCHAIN=go1.27.1`, after one wasted cycle forgetting the prefix); `check-pin-drift.sh` green; `go-structure-linter` 0 findings; art-dupl re-run (flags corrected per the tool's own note: `--type-aware` dropped): 9 → 7 groups.
8. **Daemon capture verified:** all session changes landed in auto-commits `d1dd7ea` (code, −52/+12), `49fd8c0` (guards/manifests/AGENTS), `57e145a` (AGENTS). Tree clean except an unrelated untracked status report from another session.
9. **Episode-1 report cross-referenced before writing this one** — the series-continuity step that was missing last time (see d-1).

## b) PARTIALLY DONE

1. ~~**Toolchain unification:** 4 of 11 (root, health, integration, go.work). **5 satellites still lag** — flightrecorder (1.27), flightrecorderhealth (1.26.5), otel (1.27), realtime (1.26.7), security (1.27) — so `scripts/check-go-directives.sh` remains FAIL on those 5 and CI's directives job stays red. Integration moved by tidy, not by decision; the floor policy question (episode-1 Q1) is still unanswered.~~ done — all 11 unified 2026-09-29 (SUPERB v5 T01/T02); guard green 11/11
2. ~~**Episode-1 (f) list:** executed this session: #10 (integration suite under go1.27.1 — green), #27 (sort sweep — 2 of 3 sites; third accepted with rationale), #47 partially (rationale lives in AGENTS posture note, not per-site `// intentional:` markers), #30 partially (rationale in AGENTS, NOT in `doc/planning/design-decisions.md` as specified). Everything else from that list untouched.~~ done — the remainder executed by the 2026-09-29 passes; Decision 13 owns the accepted-clone rationale
3. **`failingServiceNames`:** accepted rather than modernized — it filters (non-nil errors) then sorts; a stdlib one-liner doesn't exist without iterator gymnastics, and iterator gymnastics would be less readable. Deliberate, documented. (Settled — Decision 13 / AGENTS Clone posture.)
4. **art-dupl process:** the dedup skill says re-run after EACH refactor and test after every change; this session batched — edits first, then one verification pass per module suite, art-dupl once at the end. Outcome identical, process deviation noted. (Incident note — no open action.)
5. ~~**`--type-aware` flag note:** the tool itself printed that `--suggest-generics` supersedes it. The corrected invocation (`art-dupl --sort total-tokens -t 1 --suggest-generics --timing --rich-text`) is what this session converged on, but it is not yet recorded anywhere durable (AGENTS has no art-dupl invocation note).~~ done — recorded in the Decision 13 amendment (2026-09-29)

## c) NOT STARTED

- ~~The 5-satellite go-directive unification decision + rollout (blocked on episode-1 Q1, still open — re-asked below).~~ done — 2026-09-29
- ~~HARVEST of either episode's section (f) into `TODO_LIST.md` (both episodes instructed to wait; the debt compounds).~~ done — 2026-09-29 docs-health pass
- ~~CI reality check: `go.work` is still untracked while `check-go-directives.sh` reads it — the CI job's true remote state is still unverified.~~ done — tracked + CI green (run 36542306033)
- ~~Audit of the ~431 suppressed/non-actionable art-dupl groups (never inspected across BOTH episodes).~~ done — 477 groups audited 2026-09-29 (12-45 session)
- ~~art-dupl standing policy: recurring ritual step vs ad-hoc; suppression/baseline config for the 7 accepted groups.~~ decided — event-driven, no suppression config (Decision 13 amendment)
- ~~`doc/planning/design-decisions.md` entry for the accepted-duplication policy (episode-1 item #30, still the right home).~~ done — Decision 13 (+amendment)
- ~~`[Unreleased]` CHANGELOG entries for BOTH episodes' internal refactors (see d-1 — this session repeated the gap).~~ done — added 2026-09-28
- ~~Post-unification otel benchmark re-baseline (episode-1 item #50; toolchain jump invalidates 2026-09-16 numbers on principle).~~ routed — TODO_LIST benchstat candidate (allocs identical 2026-09-29; ns/op pending quiet machine)

## d) TOTALLY FUCKED UP

Nothing new this session is truly fucked. The honest list, in order of embarrassment:

1. **Two written-down improvement items from episode 1 were REPEATED, not avoided.** Episode 1's e-list said: (a) "Changelog-with-refactor — internal refactors on released modules get an `[Unreleased]` CHANGELOG line in the same train"; (b) "Baseline-first discipline." This session did neither up front: no CHANGELOG entries for the core (`metrics.go`) and cqrs (`eventservice.go`) refactors on released modules (v0.5.1 / v0.5.0), and no pre-edit integration baseline — so when the drift guard failed, "pre-existing red" had to be _reasoned out_ instead of _known_. The root cause is structural: **I didn't read the prior episode's findings section before starting.** Both episodes are the same series; the e-list is a checklist that was ignored.
2. **"Zero harmful duplication" claims still rest on an unaudited suppression heap.** 9 shown / 431 suppressed at -t 1; episode 1 had 3 shown / 82 suppressed. Neither episode looked inside. The claim "every remaining clone is accepted" is only proven for the shown set.
3. **AGENTS.md remains structurally full** (376/377). Every future note costs a trade. The cap exists and keeps biting; the file needs a section extracted to module READMEs, not line-golf.
4. **Minor: wasted cycle on golangci-lint** without the `GOTOOLCHAIN=go1.27.1` prefix — the exact gotcha already documented in AGENTS and already used for `go test` minutes earlier. Sloppy consistency.
5. **Behavioral deltas shipped silently in the moment** (surfaced only here): SSE test start-timeout 2s → 5s (testkit's constant); empty-map `sortedRoutes`/`sortedEngineNames` now return `nil` instead of empty non-nil slices (verified safe for all callers — both are range-only — but it is a semantic change worth knowing about).

## e) WHAT WE SHOULD IMPROVE

1. **Series continuity discipline:** every status-report/self-review session in a series must read the prior episode's (d) and (e) sections FIRST and treat them as a checklist. This one change would have prevented d-1 entirely.
2. **Baseline-first, no exceptions** — even (especially) when the environment is known-broken. A 10-second `go test ./...` before editing converts "why is this red?" archaeology into a diff.
3. **Changelog-with-refactor on released modules** — same train, no exceptions. (Written twice now; make it a ritual step, not a memory.)
4. **Audit the suppression heap once:** one pass over art-dupl's suppressed/non-actionable groups to confirm nothing harmful hides there; then the "zero harmful duplication" claim is actually earned.
5. **Decide the unification instead of drifting into it:** integration joined 1.27.1 because tidy pushed it — the unification is happening by gravity, not decision. Either commit to the floor or pin against it; the in-between state keeps the guard red.
6. **AGENTS.md headroom:** extract one section (candidates: the per-module build-command table, or module gotchas) to a module README before the next note is needed.
7. **Record the corrected art-dupl invocation** in AGENTS tooling notes (drop `--type-aware` when `--suggest-generics` is set) so the next session doesn't re-derive it from the tool's warning.
8. **Behavioral deltas deserve a sentence at change time,** not an appendix in a status report.

## f) NEXT — up to 50 things to get done (impact-ordered; brainstorm, not commitment)

**P0 — decide + unblock**

1. ~~Answer the go-directive floor question (g-1) and unify the 5 lagging satellites (flightrecorder, flightrecorderhealth, otel, realtime, security) to the chosen floor — one mechanical train.~~ done — unified 2026-09-29 (SUPERB v5 T01/T02) — floor question settled by execution
2. ~~Re-run `scripts/check-go-directives.sh` to green; the CI `go-directives` job stops being red.~~ done — green 11/11 since 2026-09-29
3. ~~Decide `go.work` tracked status (CI reads it; untracked = CI parses nothing on fresh checkout) — commit it or add a CI generation step.~~ done — go.work tracked 2026-09-29 + charter-guarded 2026-10-07
4. ~~Re-run the full 11-module matrix (build/vet/test -race) once unification lands; drop the `GOTOOLCHAIN=go1.27.1 GOWORK=off` prefixes and stale GOEXPERIMENT notes from AGENTS build commands.~~ done — matrix green 2026-09-29 — delta: GOTOOLCHAIN prefixes KEPT (machine default still 1.26.7; AGENTS documents why)
5. ~~Verify gopls/LSP recovers post-unification (AGENTS says it currently fails under the drift).~~ done — LSP recovered (live diagnostics 2026-09-29)
6. ~~`[Unreleased]` CHANGELOG lines for: core (sortedRoutes stdlib swap), cqrs (sortedEngineNames stdlib swap) — and retroactively for episode 1's `setStarted`/`runHooks` refactors if still missing.~~ done — entries added 2026-09-28; shipped in the core v0.6.0/v0.7.0 trains
7. ~~Run `check-pin-drift.sh` after the unification train (release-ritual step 5).~~ done — green (release-ritual step 5)
8. ~~Verify the CI `go-directives` job's actual remote state (suspected red since episode 1).~~ done — CI go-directives green parsing the committed go.work (run 36542306033)

**P1 — close this session's gaps**

9. ~~HARVEST both episodes' (f) lists into `TODO_LIST.md` / `ROADMAP.md` (docs-health HARVEST mode; dedup against each other).~~ done — 2026-09-29 docs-health pass rebuilt TODO_LIST (64 lines open-only)
10. ~~One audit pass over art-dupl's suppressed/non-actionable groups (~431); confirm nothing harmful hides there.~~ done — audited 2026-09-29 (12-45 session) — 477 groups inspected, one harmful clone collapsed
11. ~~Decide art-dupl standing policy: recurring BuildFlow/ritual step vs ad-hoc (episode-1 item #31, still open).~~ done — DECIDED 2026-09-29 — event-driven (Decision 13 amendment)
12. ~~If recurring: add suppression/baseline config for the 7 accepted groups so future runs surface only NEW clones.~~ **NOT-DO — NOT-DO — suppression/baseline generator rejected (Decision 13 amendment).**
13. ~~Write the accepted-duplication entry in `doc/planning/design-decisions.md` (episode-1 item #30 — AGENTS posture note is the summary, not the decision record).~~ done — Decision 13 (2026-09-28; amended 2026-09-29)
14. Add the corrected art-dupl invocation (no `--type-aware` with `--suggest-generics`) to AGENTS tooling notes. **→ recorded in the Decision 13 amendment (not AGENTS — cap discipline); AGENTS Clone posture carries the standards**
15. ~~Decide per-site `// intentional:` markers vs repo-level posture note (episode-1 item #47; AGENTS note currently carries it alone).~~ done — decided repo-level — Decision 13 records the no-exclude posture
16. ~~Run `check-dependabot-parity.sh` once (episode-1 item #21; the tidy-driven directive bumps may be dependabot-sourced — item #22 of episode 1).~~ done — green 2026-09-28; in .githooks/pre-commit since 2026-10-07
17. ~~Audit what originally drifted the module directives (episode-1 item #15, auto-commit `ea17377`) — still unknown.~~ **Won't implement — moot — the directives guard blocks the class at commit time; the historical auto-commit is frozen history.**
18. ~~Confirm root/health/integration go.mods carry no other dead indirects like the `go-etag v0.6.0` line tidy dropped here.~~ done — 2026-10-06 fleet sweep tidied all modules; go.sum clean across the fleet (2026-10-12 scan)

**P2 — quality / hygiene**

19. ~~Testkit could grow a drain-window assertion helper (episode-1 item #35; the drain contract tests hand-roll it).~~ done — testkit.DrainWindowProbe shipped in core v0.7.0
20. ~~Typed `seriesKey` for `metrics.go` to kill the `|`-join/`SplitN` stringly pair (episode-1 item #39; re-noticed this session at the same file).~~ done — typed seriesKey shipped in core v0.7.0
21. ~~Add the hook-error-code pinning test (`server.drain_hook_failed` / `server.shutdown_hook_failed`) — episode-1 d-3, still unpinned.~~ done — hookcodes_test.go (2026-09-28; core v0.6.0)
22. ~~`errors.Join` multi-hook failure message shape test (episode-1 item #42).~~ done — shipped in core v0.7.0
23. ~~Live E2E of `example/` and `health/example` after both refactors (episode-1 item #29).~~ done — T19b 2026-09-29 — root + health examples live-probed; fr + otel examples live-verified on their v0.2.0 trains
24. Post-unification otel benchmark re-baseline (episode-1 item #50). **→ routed — TODO_LIST benchstat re-baseline candidate (allocs identical 2026-09-29; ns/op waits for a quiet machine)**
25. ~~`go work use` dry-run in a scratch work file before the real unification (episode-1 item #32).~~ **NOT-DO — NOT-DO — moot: go.work was already regenerated at 1.27.1 by the unification train.**
26. ~~Cross-project lesson to crush-config `references/lessons.md`: "check GOTOOLCHAIN/go.work parity before trusting AGENTS build commands" (episode-1 item #25/48 of its c-list).~~ done — the caller-wins-toolchain lesson class lives in crush-config references/lessons.md (2026-09-22 entry)
27. ~~Wire `check-go-directives.sh` into BuildFlow pre-commit so parity breaks before push (episode-1 item #24).~~ done — .githooks/pre-commit runs check-go-directives.sh FATAL (2026-10-07)
28. ~~Record the `/mnt/buildcache` toolchain-cache trick as a recipe (episode-1 item #43).~~ **Won't implement — not written — the AGENTS Toolchain paragraph already carries the /mnt/buildcache fact.**
29. Session-startup doctor: `go env GOTOOLCHAIN` + `check-go-directives.sh` before trusting any build command (episode-1 item #44). **→ AGENTS Build & Verify owns the GOTOOLCHAIN-prefix contract instead; no doctor step adopted**
30. ~~CI job ordering: directives-guard first, fail fast before the expensive matrix (episode-1 item #40).~~ **Won't implement — moot — guards are cheap independent jobs and ci-preflight fails loud since 2026-10-07.**
31. ~~Add "run check-go-directives.sh" to the AGENTS Release Ritual next to check-pin-drift (episode-1 item #41).~~ done — AGENTS Release Ritual step 3 names check-go-directives.sh alongside check-pin-drift.sh
32. ~~Consider a `toolchain` directive in go.mods so the floor survives `GOTOOLCHAIN=local` machines (episode-1 item #38).~~ **NOT-DO — NOT-DO — pre-tag-checks.sh rejects toolchain lines; the repo mechanism is the GOTOOLCHAIN prefix.**
33. ~~Named `type Hook func(context.Context) error` to shorten `runHooks` signature (episode-1 item #28).~~ done — appkit.Hook alias shipped in core v0.7.0
34. ~~Write `server.*_hook_failed` codes into `doc/DOMAIN_LANGUAGE.md` if present (episode-1 item #45).~~ done — error codes recorded in doc/DOMAIN_LANGUAGE.md Events table (2026-09-28)
35. ~~Explicit test for `runHooks` run-every-hook semantics (episode-1 item #46).~~ done — TestHooks_EveryHookRunsWhenAnEarlierOneFails (2026-09-28)
36. ~~Confirm episode-1's `runHooks`-shaped `errors.Join` sweep finding (item #48: "none found; confirm").~~ done — none found (re-confirmed 2026-09-28)
37. ~~Re-check health module's go-health v0.3.0 evaluation note accuracy (episode-1 item #34; AGENTS still flags it).~~ done — OVERTAKEN — health swept to go-health v0.4.1 (2026-09-29) then v0.5.0 (2026-10-06)
38. ~~Verify errorpages example's local-dev `replace ../` never leaks into a tag (tag-hygiene spot check).~~ done — pre-tag-checks.sh asserts no filesystem replace in the tagged artifact; errorpages v0.1.0/v0.1.1 shipped clean

**P3 — roadmap fuel**

39. File or drop the drafted go-health `NewProbe` recorder-sentinel upstream ask (`doc/feedback/outgoing/2026-09-20_...`, filing gated). **→ user-gated filing — TODO_LIST upstream-asks row (draft doc/feedback/outgoing/2026-09-20_upstream-ask-gohealth-recorder-sentinel.md)**
40. Core TLS option — still trigger-gated on PapDashboard (keep deferred unless the trigger fires). **→ deferred register — trigger-gated on PapDashboard TLS demand**
41. httputil `NewServerListener` ask — still the composition blocker for appkit-over-httputil; nudge or track. **→ user-gated — TODO_LIST upstream-asks row (the composition blocker, doc/planning/2026-09-16_composition-spike-verdict.md)**
42. ~~cqrs v0.6.0: remove the deprecated `SQLitePath` alias (scheduled breaking change; plan the train).~~ done — cqrs v0.6.0 shipped 2026-09-29 (BREAKING SQLitePath removal + migration note)
43. ~~Fresh-consumer proxy smoke for core before its next tag (go.mod/go.sum changed trains).~~ done — v0.6.0/v0.7.0/v0.8.0 all fresh-consumer proxy-PASS
44. Cordis bridge — trigger-gated, unchanged. **→ watchlist — TODO_LIST P3 cordis row (trigger 1/3 met, zero consumer demand)**
45. Decide whether dependabot's grouped bumps should be allowed to touch `go` directives (ignore-rule if so). **→ no dependabot ignore rules — the commit-time directives guard covers the class; parity guard green**

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. ~~**Go-directive floor policy (carried over from episode 1, now more urgent):** commit to **go 1.27.1 across all 11 modules** (accept what go-etag v0.6.0 forces, unify the 5 stragglers, guard goes green, GOTOOLCHAIN prefixes die) — or pin against 1.27.1 to keep a lower floor? Integration has now drifted in by tidy; the in-between state keeps the guard and CI red. This single answer unblocks items 1–8.~~ done — resolved by execution — unified 2026-09-29; GOTOOLCHAIN prefixes deliberately KEPT (machine default still go1.26.7)
2. ~~**art-dupl standing policy:** should the 7 accepted clone groups be suppressed/baselined so future runs surface only NEW clones (and should art-dupl become a recurring ritual step) — or stay visible every run as a recurrence tripwire? This decides items 11–12 and 15.~~ done — DECIDED 2026-09-29 — no suppression config, accepted clones stay visible, event-driven cadence (Decision 13 amendment)
3. ~~**`doc/` vs `docs/` (carried over from episode 1, still unresolved):** history keeps planning/feedback/recipes under `doc/` (singular; `docs/` is the catalog Go module), while your instructions and both episode reports use `docs/status/`. Which path is canonical going forward? (Both episodes obeyed your explicit instruction and live in `docs/status/` — the split is now permanent-looking.)~~ done — RESOLVED 2026-09-28 — docs/status/ canonical; doc/ keeps planning/feedback/recipes/research

---

_Generated by Crush session 2026-09-28 (art-dupl -t 1 sweep). Point-in-time snapshot — annotate, don't rewrite, when bringing current._
