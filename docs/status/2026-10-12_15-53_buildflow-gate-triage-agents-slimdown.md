# Status Report — BuildFlow Gate Triage, AGENTS.md Slim-Down, and the Leak #8 Attribution

**When:** 2026-10-12 15:53 CEST
**Scope:** this session only (15:11–15:53) — the red BuildFlow full-mode gate (exit 69, 15 error-severity findings) driven to green (exit 0), plus everything the triage surfaced.
**Session type:** BuildFlow failure triage (per the buildflow skill's close-the-loop rule: every finding ends fixed, filed, or documented).

---

## Headline numbers

| Metric                  | Before (15:11 run)                   | After (15:46 re-run)                                       |
| ----------------------- | ------------------------------------ | ---------------------------------------------------------- |
| Gate verdict            | `ERROR` exit **69**                  | **exit 0**, "passed with warnings"                         |
| Error-severity findings | 15 (jscpd 14, go-structure-linter 1) | **0**                                                      |
| Steps green             | 223/258                              | 222/258 (1 step fewer: go-work-sync now skipped by config) |
| Skipped via config      | 3                                    | 4 (+ go-work-sync)                                         |
| AGENTS.md lines         | 366 (cap 220)                        | **126**                                                    |
| New permanent config    | —                                    | `.jscpd.json`, `.buildflow.yml` skip_steps                 |

---

## a) FULLY DONE

1. **go.work leak #8 reverted and attributed.** The 15:11 run's `go-work-sync` step re-added `./integration` (present on disk, excluded by charter); the auto-commit daemon committed it 62 seconds later. Reverted by hand; `scripts/check-workspace-charter.sh` green ("root + 10 modules, ./integration excluded").
2. **Recurrence prevented:** `go-work-sync` added to `.buildflow.yml` `skip_steps` with a full rationale comment. Verified live in the green re-run: `⊘ go-work-sync (skipped via skip_steps config)`. Membership changes are now hand-made when the dynamic charter guard demands them.
3. **AGENTS.md slimmed 366 → 126 lines** (cap 220, enforced by go-structure-linter AND the BuildFlow preflight `docs/agents-md-size`). All per-module code-organization tables, dependency tables, and gotchas moved **verbatim** to the new `docs/agents/module-details.md`; AGENTS.md keeps one-line pointers. `go-structure-linter .` → 0 findings.
4. **Memory-loss check on the move caught one casualty and fixed it:** the integration-module "T14 nuance" (cross-repo contract legs version-locking) was missing from the moved section — re-added to `docs/agents/module-details.md`.
5. **jscpd gate errors 14 → 0.** New root `.jscpd.json` (fleet convention, matches go-branded-id) ignoring `**/.golangci.yml` (per-module lint configs are deliberately standalone — golangci-lint has no config inheritance) and `**/dashboards/**` (generated SigNoz JSON panel blocks). jscpd now reports warnings only (test-code clones — accepted-by-design per the clone posture).
6. **cqrs `dlq_test.go:42` wrapcheck finding fixed** with a reasoned `//nolint:wrapcheck` (the classified Rejection IS the fixture — projectionhost must see the family to quarantine). cqrs lint: 0 issues; `go test -race` green.
7. **shellcheck SC2034 fixed** in `scripts/post-commit-head-watch.sh`: the dead `reason` variable (only ever set on success branches, never read — empty exactly when it would matter) converted to per-branch comments. shellcheck clean (via `nix shell nixpkgs#shellcheck`).
8. **lychee dead links ×2 fixed:** archived SUPERB plan's audit link had the wrong relative depth (`../research/` → `../../research/`); `improvement-audit.md` linked `error-family-adoption.md`, which **never existed in git history** (a dangling promise to an unwritten spec — the adoption itself shipped repo-wide) — link removed, substance kept.
9. **golangci-lint config hygiene (real bug found):** all 12 module `.golangci.yml` files carried `exhaustruct_v5.exclude` settings blocks that were **dead** — golangci-lint v2.14.0's schema accepts _no_ exhaustruct_v5 settings, and a stripped-nolint probe on `service.go:91` proved the excludes never suppressed anything (the `//nolint:exhaustruct_v5` directives were the real mechanism all along). All 12 dead blocks removed; `golangci-lint config verify` green in all 12 modules; lint results unchanged (spot-checked root/cqrs/otel/integration = 0 issues).
10. **otel module verified after BuildFlow's autofix** (`setup.go` var-block realignment): `go test -race` + lint green.
11. **Final full gate re-run: EXIT=0** (was 69), 32s, `222/258` steps, warnings-only remainder.
12. **Branching-flow spot-checks closed:** `metrics.go:176/262` bucket-index findings are false positives (buckets provably `len(durationBuckets)+1`, allocated under the same mutex as every access); `errorpages.go:198` is guarded by the `p == ""` early return; cqrs `eventservice.go:341` `_ = aux.Close()` is the idiomatic error-cleanup-path ignore (documented by-design class).
13. **samber/lo question resolved by policy:** how-to-golang's required-libraries reference names `lo`+`mo` as the REQUIRED functional stack — the ~29 go-auto-upgrade suggestions are legitimate but each module's go.mod gains a require, and **member go.mod changes ride release trains only**. Recorded as a deliberate non-fix + queued train work in TODO_LIST with every flagged site enumerated.
14. **Memory updated:** AGENTS.md (leak #8 attribution + skip, 220-line cap, jscpd decision, exhaustruct dead-settings discovery, gomod-check red-by-design note); TODO_LIST (header, leak row updated to #7+#8 with the BuildFlow attribution, new samber/lo train row, closed the now-done "AGENTS deep slim-down decision" row).

## b) PARTIALLY DONE

1. **BuildFlow upstream issue list** — identified this session (see the BuildFlow section below), **not filed**. Filing requires verify-before-filing source checks in `/home/lars/projects/BuildFlow`, which has a **concurrent session** (HEAD moved 893de6b → 34e4526 → binary rebuilt to 8a34928 mid-session) — I deliberately did not touch that repo.
2. **samber/lo adoption** — policy decided, sites enumerated (cqrs/domain.go:96,101 + domain_test.go:86,477; realtime/handler.go:273 + realtime_test.go:693; security/csrf.go:27 + origin_check.go:39; health/probe.go:60; otel/metrics_test.go:48; frh/adapter.go:75 withheld-by-tool), code untouched until a train.
3. **Status-index hygiene** — this report lands unindexed in `docs/status/README.md` (docs-health owns the index; a concurrent session also added reports at 15-12/15-17 today).
4. **Section (f) harvest** — the list below is brainstorm-grade; TODO_LIST already carries the durable ones; a docs-health HARVEST pass should route the rest to ROADMAP or reject them.

## c) NOT STARTED (noticed this session, untouched)

1. **Filing the BuildFlow/golangci-lint upstream issues** (list below).
2. **Root `.golangci.yml` staleness I saw but did not fix:** `run.go: 1.26.7` while the repo floor is 1.27.1, and `build-tags` still listing the RETIRED `goexperiment.arenas/jsonv2/runtimesecret/simd` prefixes (retired fleet-wide 2026-09-29). Lints are green so it is inert-ish — but it contradicts the unification train and nobody caught it for two weeks. Should ride the next config-touching change or the 1.27.2 floor train.
3. **The "9 tools unavailable (health check failed)" line** — appears every run; I never ran `--verbose` once to see which nine.
4. **The `+2 more` cqrs-lint findings** — I enumerated 5 of 7; the truncated two were never shown (the never-pipe-output rule says: run `buildflow -s cqrs-lint --format finding` and read the file).
5. **Pre-existing queue items** (all tracked in TODO_LIST, none started here): 1.27.2 floor train, release trains #25/#26, go-cqrs-lite v4.13.0 cookbook ritual, pkg.go.dev flightrecorder 404, browser CSP pass, errorpages statusRecorder swap, upstream asks (go-health sentinel, NewServerListener, fr cooldown, GracefulClose/timers), otel polish batch, composition-gap E2Es, battery waves W3–W5, example StartHooks migration, i18n residue, pma dossier follow-ups.

## d) TOTALLY FUCKED UP

**Nothing from this session.** No reverts needed, no data loss, no broken states left behind; the daemon committed each work batch cleanly and the final gate is green.

The closest to "fucked up" is what the session **found** (all now fixed or documented):

- **Leak #8 landed silently**: BuildFlow mutated go.work, and pma committed it 62s later — and pma disables all git hooks by design (`core.hooksPath=/var/empty`), so the charter pre-commit guard NEVER sees daemon commits. Only CI (or luck) catches that class. The system worked exactly as documented and still let the violation land.
- **A month of dead lint config**: the exhaustruct_v5 exclude blocks (authored 2026-09-04) never worked; every module's "0 issues" was nolint-driven, not config-driven, and the configs failed `config verify` the whole time. Two weeks of "green" hid an invalid fleet-wide config.

## e) WHAT WE SHOULD IMPROVE

1. **Daemon-bypass gap**: guards wired into `.githooks/` are structurally blind to pma commits. A post-buildflow (or post-daemon-write) charter+directives guard run, or a watcher, would close the 62-second window CI currently covers.
2. **Cache/config coupling**: `.jscpd.json` changes didn't invalidate jscpd's cached verdict — needed `BUILDFLOW_NO_RESULT_CACHE=1`. Tool config files belong in result-cache keys (BuildFlow-side).
3. **Verify golangci config validity fleet-wide on a cadence**: `golangci-lint config verify` per module is cheap and would have caught the exhaustruct block the day it was written. BuildFlow's config-verify step exists — it just never gated (warning-level).
4. **Full-matrix trust, but verify the tail**: I leaned on BuildFlow's full-mode test-race for module coverage after config changes (correct) but only hand-verified lint on 4 of 12 modules; the final green run covers it — the lesson is to name which layer (BuildFlow vs hand) carries each verification claim.
5. **Truncated findings lists**: two `+N more` ellipses hid findings; the never-pipe rule exists and I still let the summary do the reading. Drill into every `+N more` before declaring a class by-design.
6. **Document tool-config conventions where the fleet can find them**: `.jscpd.json`'s existence/keys live only in sibling repos — one line in the buildflow skill references would have saved the discovery loop.
7. **Status-report HTML-vs-md**: the status-report skill is HTML-canonical; the operator demanded `.md` — honored, flagged here per the skill's override rule.

## BuildFlow issue candidates (fleet-wide, from this session's evidence)

All of these apply to many/all fleet projects, not just go-appkit. Per verify-before-filing, each needs source-level verification in the BuildFlow repo before filing — **not done** (concurrent session in that repo). Evidence noted is run-log level.

| # | Candidate                                                                                                                                                           | Evidence (this session)                                                                                                                                                                                                 | Proposed ask                                                                                                                                                                                                                       |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Workspace-exclusion concept**: go-work-sync treats any on-disk module missing from go.work as drift and re-adds it, fighting deliberate exclusions (pin charters) | Leak #8: `WARN go.work is missing module "integration"` → `go-work-sync 1 fixed` → pma committed it; charter comment ignored                                                                                            | One BuildFlow-level "workspace exclusions" convention (go.work comment directive, `.buildflow.yml` option, or charter-guard autodetect) honored by go-work-sync AND gomod-check; skip_steps is the blunt all-or-nothing workaround |
| 2 | **gomod-check has no acknowledgement path** for deliberate states                                                                                                   | Every run reports info-level `work-use-missing-module` for the same charter exclusion — permanent noise                                                                                                                 | tool_options (exclude rules) or a baseline/ack mechanism for known-deliberate findings                                                                                                                                             |
| 3 | **Result cache ignores tool-config changes**                                                                                                                        | jscpd replayed stale error findings after `.jscpd.json` gained ignores; only `BUILDFLOW_NO_RESULT_CACHE=1` cleared them; fresh scan diverged wildly (16 vs 107 findings)                                                | Include each tool's config file(s) in the result-cache key                                                                                                                                                                         |
| 4 | **jscpd defaults gate-block on config-file clones**                                                                                                                 | 14 error-severity findings, all `.golangci.yml`×`.golangci.yml` — a fleet-standard, BuildFlow-adjacent pattern (per-module configs) failed the gate out of the box; every repo hand-rolls the same `.jscpd.json` ignore | Ship default ignores for known config files (`.golangci.yml`, `dprint.json`, lockfiles) or demote config-file clones below error; also investigate the `Duplicated N lines (0 tokens)` + `filesScanned: 0` stats quirk             |
| 5 | **go-auto-upgrade lo suggestions are unacknowledgeable**                                                                                                            | ~29 permanent warnings per run: tool withholds import-adding rewrites when `samber/lo` isn't a go.mod require; library repos deferring adoption by policy see the noise forever                                         | tool_options switch (`samber_lo: off/suggest/rewrite`) or a pending-adoption ack                                                                                                                                                   |
| 6 | **lychee private-repo policy is undecided at BuildFlow level**                                                                                                      | Preflight warn in every run: "fleet authenticate-vs-exclude policy is undecided" — the tool itself print-admits the gap                                                                                                 | Decide and implement one: GITHUB_TOKEN injection or default namespace exclude                                                                                                                                                      |
| 7 | **golangci-lint-auto-configure vs config-verify self-consistency** (IF auto-configure authored the exhaustruct blocks — unverified who wrote them 2026-09-04)       | Config-verify rejects what the configs carry; 12 modules invalid for a month at warning level                                                                                                                           | Verify the author; if auto-configure emits schema-invalid settings, fix it; consider promoting config-verify failures for schema-invalid configs from warn to error                                                                |
| 8 | (adjacent, NOT BuildFlow) **exhaustruct_v5 empty settings schema in golangci-lint v2.14.0**                                                                         | Schema accepts zero settings keys; v5 dropped include/exclude with no migration note                                                                                                                                    | File at golangci-lint upstream, not BuildFlow                                                                                                                                                                                      |

## f) NEXT — up to 50, ranked by impact (brainstorm-grade; USER-gated marked)

1. ~~**File BuildFlow issues #1–#6 above** (after verify-before-filing source checks; mind the concurrent BuildFlow session).~~ done — BuildFlow #46-49 + go-auto-upgrade #3 filed (17-27 session, source-verified)
2. ~~**Root `.golangci.yml` hygiene**: `run.go` 1.26.7 → align with floor; drop retired `goexperiment.*` build-tags (verify analysis behavior first).~~ done — swept on the 1.27.2 floor train (run.go + dead blocks + retired tags gone)
3. ~~**1.27.2 floor train** (go.work + 12 go.mods + AGENTS + fixture in ONE train): unblocks cqrs-htmx v4.14.0 pin AND clears govulncheck GO-2026-6617 (stdlib HTTP/2 HPACK race) across every module. USER-gated.~~ done — wave 1: 12 go.mods + go.work + fixture at 1.27.2; GO-2026-6617 cleared
4. ~~**Release trains #25/#26** (testkit.NewService + health NewProbe — code-complete): tag/push USER-gated per ritual.~~ done — core v0.9.0 + health v0.2.0 tagged/pushed; issues closed
5. **P1: restore `SSH_PRIVATE_KEY` Actions secret** (owner-only) + one fully green master run.
6. **samber/lo adoption train** (~12 sites, 8 modules; ride the next dependency train). **→ decided test-only 2026-10-12 (runtime sites stay hand-rolled); train pending go-auto-upgrade #3**
7. ~~**go-cqrs-lite system/v4 v4.13.0**: cookbook re-verification ritual + pin-sweep decision (DUE since 2026-10-12).~~ done — cqrs v0.8.0 (cookbook re-verified + full consistent sweep)
8. ~~**Harvest this report's section f** via docs-health into TODO_LIST/ROADMAP.~~ done — this pass (2026-10-12)
9. ~~**Index this report** (+ today's concurrent-session reports) in `docs/status/README.md`.~~ done — this pass (five rows 15-12 through 20-36)
10. **One `buildflow --verbose` run** to name the 9 unavailable tools + audit the 67 not-applicable list once.
11. ~~**Enumerate the truncated `+2` cqrs-lint findings** (`-s cqrs-lint --format finding` → file, read whole file).~~ done — cqrs-lint exits 0 at cqrs v0.8.0 (E014 trailing form + C023 closed the tail)
12. **Charter/directives guard for daemon commits** — post-buildflow step or watcher to close the pma bypass window.
13. **Decide `.jscpd.json` ignore strategy**: path-scoped (current) vs format-scoped (drop json entirely, go-branded-id style).
14. **Consider taming jscpd warning volume** (107 warnings, mostly test clones): minTokens/minLines tuning vs documenting the accepted class.
15. **File exhaustruct_v5 schema issue at golangci-lint** (not BuildFlow).
16. ~~**Verify who authored the dead exhaustruct blocks** (auto-configure vs hand) — determines BuildFlow issue #7.~~ done — resolved — dead blocks removed in-session; no BuildFlow issue needed
17. **Quick fleet sweep for go-work-sync re-adds in sibling repos** (leak #8's mutator was fleet-wide tooling).
18. **pma dossier follow-up**: `journalctl -t pma` around leak windows; per-repo cooldown for train repos (pre-existing).
19. **Browser CSP pass** over the hardened health dashboard (pre-existing P2).
20. **errorpages: swap hand-rolled `statusRecorder`** for httputil.ResponseRecorder (USER-gated, pre-existing).
21. **go-health recorder-sentinel upstream ask** — filing USER-gated (pre-existing).
22. **httputil `NewServerListener` ask** — unblocks Service composition + Core TLS (pre-existing, USER-gated).
23. **fr cooldown-combinator ask** (pre-existing, USER-gated filing).
24. **GracefulClose + timer-stop upstream flip watch** (pre-existing; four-file tripwire enumerated). **→ TODO row — timers half closed at cqrs v0.8.0; GracefulClose watch open**
25. ~~**fr v0.2.1 sweep** on the frh/cqrs trains (pre-existing).~~ done — fr v0.2.1 in cqrs/frh/otel go.mods (2026-10-12 trains)
26. ~~**Example StartHooks migration** (trigger met 2026-10-07; pre-existing).~~ done — decc462
27. ~~**health follow-through after the #25 train** (bench re-check, fuzz session, godoc note; pre-existing).~~ done — health v0.2.0 + fuzz/benchstat evidence in its CHANGELOG
28. **pkg.go.dev flightrecorder 404** — >48h escalation rule met (pre-existing).
29. ~~**otel polish batch** (benchmarks, env-pin test, tail-sampling recipe; pre-existing P3).~~ done — 104f28c (TODO row struck)
30. **Composition-gap E2E watchlist** (otel×realtime, security×errorpages, …; pre-existing).
31. **Battery waves W3–W5** (canonical spec referenced, demand-gated; pre-existing).
32. **Full cqrs-htmx setup suite hermetic run** (pre-existing).
33. ~~**Consumer-claim drift ritual posture** — stop citing exact consumer pins vs guard them (USER-gated, pre-existing).~~ done — resolved — pins kept + guarded (check-consumer-claims.sh, T05)
34. **i18n residue post-httputil v1.6.0** (pin-drift sweep + optional E2E leg; pre-existing).
35. **BuildFlow dprint exit-14 on CHANGELOG-only commits** — upstream fix (pre-existing deferred register).
36. **Lychee GITHUB_TOKEN-vs-exclude decision** (ties to BuildFlow issue #6).
37. **BuildFlow skill references**: document `.jscpd.json` convention + gomod-check/go-work-sync charter caveat (skill-repo PR).
38. **Link `docs/agents/module-details.md`** from the modules' READMEs (discoverability) or accept AGENTS-pointer-only.
39. **Confirm the 220-line cap's source of truth** (BuildFlow preflight vs linter default) and note the fleet standard where repos can see it — the repo doc said 377 until today.
40. **Daemon attribution test**: confirm whether gopls-class tools still mutate go.work now that go-work-sync is skipped (leaks #1–7's author is still unnamed).
41. **Consider `golangci-lint config verify` as a CI job** (cheap; would have caught the exhaustruct class at authoring time).
42. ~~**Re-check `run.build-tags` necessity** in root .golangci.yml (arenas/runtimesecret/simd — analysis-only tags with no build counterpart since the GOEXPERIMENT retirement).~~ done — retired tags gone from root .golangci.yml (1.27.2 sweep)
43. **module-details.md decay watch**: the moved content is point-in-time verbatim — first module train should re-verify its section (pointer exists in AGENTS.md). **→ cqrs section re-verified at the v0.8.0 train; watch continues**
44. ~~**TODO_LIST "go.work re-add class" row**: close or rewrite after BuildFlow issue #1 is filed (the skip makes the inotify hunt moot for THIS mutator).~~ done — TODO row 30 rewritten (17-27 session)
45. **Roadmap fuel** (from section f extras): workspace-exclusion convention as a fleet pattern; per-tool config schema docs; charter-guard-as-library (charter scripts are per-repo copies). **→ harvested into ROADMAP (this pass)**
46. **junk-drawer check on `docs/agents/`**: one file today; decide the directory's charter (agent-only, never user docs) before it grows. **→ harvested into TODO (this pass)**
47. ** benchstat re-baseline** when an optimization candidate exists (pre-existing P3 candidate).
48. ~~**Re-run `check-pin-drift.sh` post-trains** — the standing ritual; held cqrs-htmx leg flips with the 1.27.2 train.~~ done — 27 pin-drift legs green post-trains
49. **Update the buildflow skill's Verification-status table** with this session's verified behaviors (go-work-sync semantics, cache-key gap, jscpd config respect). **→ harvested into ROADMAP (this pass)**
50. ~~**Close the loop on the concurrent BuildFlow session** — before filing anything upstream, check whether 893de6b..HEAD already addresses any candidate (the binary jumped mid-session).~~ done — 17-27 session checked the concurrent range before filing

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. ~~**BuildFlow filings**: given the concurrent session in `/home/lars/projects/BuildFlow` (binary jumped 893de6b → 8a34928 mid-session), do you want me to verify + file the fleet-wide issues now, or wait until that session lands so I can check `893de6b..HEAD` for existing fixes first?~~ done — resolved — verified + filed (BuildFlow #46-49, 17-27 session)
2. ~~**Root `.golangci.yml` staleness**: is `run.go: 1.26.7` + the retired `goexperiment.*` build-tags a deliberate analysis pin, or an oversight to fix now (vs riding the 1.27.2 floor train)? Lint is green either way; I can't tell policy from accident.~~ done — resolved — oversight; fixed on the 1.27.2 floor train
3. ~~**samber/lo as a library dependency**: for the adoption train — is `samber/lo` acceptable as a RUNTIME require your appkit consumers inherit (policy says the stack is required, but library-surface leakage is a call only you can make), or test/dev-only with runtime sites kept hand-rolled?~~ done — resolved 2026-10-12 — test-only; runtime sites stay hand-rolled

---

**Files touched this session:** `go.work`, `.buildflow.yml`, `.jscpd.json` (new), `AGENTS.md`, `docs/agents/module-details.md` (new), `TODO_LIST.md`, `cqrs/dlq_test.go`, `scripts/post-commit-head-watch.sh`, `doc/planning/archived/2026-08-15_19-27_SUPERB-ECOSYSTEM-EXECUTION.md`, `doc/planning/improvement-audit.md`, `otel/setup.go` (BuildFlow autofix, verified), 12× `.golangci.yml` (dead-block removal). All committed by the daemon except this report.

**Verification state:** BuildFlow full gate exit 0 (15:46); workspace-charter + go-directives guards green; go-structure-linter 0; all 12 `golangci-lint config verify` green; per-module lint spot-checks (root/cqrs/otel/integration) 0 issues; otel + cqrs `-race` suites green; full-matrix test-race covered by the green BuildFlow run.
