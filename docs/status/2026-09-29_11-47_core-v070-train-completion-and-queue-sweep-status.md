# Status — Core v0.7.0 Train Completion, Integration Migration & Queue Sweep

**Time:** 2026-09-29 11:47 CEST · **Session window:** ~09:55–11:47 (resumed from the 06-40 handoff with the v0.7.0 tag command interrupted) · **Author:** Crush session
**Predecessor:** `docs/status/2026-09-29_09-53_satellite-waves-polish-batch-and-core-v070-status.md` (its §"Exact Next Steps" 1–6 are this session's spine)
**HEAD at report time:** f063f82 (mine) + 4 daemon sweeps; **3 files carry UNCOMMITTED BuildFlow mutations** in otel (see b1/d1 — not mine, not yet judged).

---

## a) FULLY DONE

1. **Core v0.7.0 train — COMPLETE (the interrupted train).** Verified the interrupted state first: tag `v0.7.0` existed locally on e4c6281, NOT on origin, and the tagged CHANGELOG still said `[Unreleased]`. Clean recovery: deleted the unpushed tag, dated the root CHANGELOG `[0.7.0] - 2026-09-29`, updated AGENTS release line (through **core v0.7.0**) + core bullet, refreshed the TODO header; structure lint green (375 counted lines vs 377 cap); committed b09ac9d; re-created the annotated tag ON the dated commit; **pre-tag gate GREEN** (pin-drift 14 OK incl. cross-repo proxy checks, dependabot parity 11/11, tagged go.mod floor-exact 1.27.1, no toolchain line, no fs replace); pushed master + tag; **proxy PASS** (`go run doc/recipes/_proxycheck/main.go github.com/larsartmann/go-appkit@v0.7.0`). Tag message states the full semantic delta.
2. **Integration migration to published v0.7.0 — COMPLETE.** `go mod tidy` resolved v0.7.0 from the now-live proxy; **found and fixed a genuinely red fixture** (`documentedPins` still said core v0.6.0 against a go.mod saying v0.7.0 — the pin_drift Go test would have failed); migrated `TestDrainWindowContract` to `testkit.DrainWindowProbe` (readiness observation) with the ping-during-drain assertion kept as a second hook — observation semantics preserved; switched both new E2E hook slices to `[]appkit.Hook` (sse_health_drain_test.go); suite `-race -count=1` green (2.3s), vet green, golangci-lint 0 issues, pin-drift + directives guards green; committed f063f82, pushed.
3. **Report tail items executed:**
   - **#50** — SUPERB v5 doc got the EXECUTED banner: T01–T09/T12/T14/T17/T18/T24–T26 marked done, with deltas (frh v0.1.4→v0.1.5 directive-floor correction + the pre-tag gate born from it; T19 otel ns/op deferred under load; T04 executed as "GOEXPERIMENT retired, GOTOOLCHAIN prefixes KEPT"; the four satellite trains + core v0.7.0 as beyond-plan scope; open list intact).
   - **#38** — broken links fixed in `doc/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md`: cordis relative path corrected (`../../../forks/cordis` → `../../../../forks/cordis`; actual location `/home/lars/forks/cordis`); **go-plugin-mvp no longer exists anywhere on disk** → relinked to `https://github.com/LarsArtmann/go-plugin-mvp` (module path documented in the file itself). PapDashboard link verified fine; every other relative link in doc/ + docs/ verified resolving.
   - **#46** — verified clean: grep found ZERO stale floor claims ("core v0.5.1"/"v0.6.0" release-state) in README/FEATURES/ROADMAP. Nothing to sync.
   - **#42** — `doc/DOMAIN_LANGUAGE.md` gained the DLQ poison-contract vocabulary (projectionhost v4.5.1 semantics): **Poison event** + **Quarantine vs restart** rows (Rejection/Corruption → quarantined, replayable via `ReplayDeadLetters`; Transient/Infrastructure → retryable, worker restarts, never quarantined).
   - **#44** — AGENTS integration bullet gained the T14 clause (cross-repo contract legs are version-locked in the fixture: go-health tracks the health module's requirement; cqrs-htmx/go-sse/ssetest/httputil legs are proxy-checked against latest); structure lint still green after.
   - **#41** — CI verified on the pushed go.work commit (run 36542306033): **all four guard jobs GREEN** — cqrs-lint 21s, pin-drift 16s, config-parity 6s, **go-directives 40s parsing the committed go.work** (T11's CI half is now proven); proxy-smoke skipped (0s — see f31); ALL 11 matrix `test` jobs die ONLY at `webfactory/ssh-agent` (`ssh-private-key argument is empty`) = the known owner-gated T10.1, workflow itself healthy.
   - **#47** — verified already resolved: no go-health v0.3.0 evaluation entry remains in TODO_LIST; AGENTS marks it OVERTAKEN by the v0.4.1 sweep (health v0.1.4).
   - **#40** — verified: `security/go.mod` example-only appkit dep was ALREADY auto-bumped to v0.6.0. **But the security CHANGELOG v0.2.0 entry still says v0.5.1** — doc drift found, follow-up queued (b4/f4).
   - **#49** — pkg.go.dev render check: **core v0.7.0 ✓, realtime v0.1.2 ✓, health v0.1.4 ✓** (all "Published: Sep 29, 2026", license-hidden by the DECIDED proprietary posture); flightrecorder v0.1.1 + docs v0.3.1 still 404 — crawler lag, per the ritual explicitly not a failure.
4. **T29 watchlist research (the findings, docs write-back pending — see b2):**
   - **cordis: `go/v0.1.0` IS NOW TAGGED** (verified in `/home/lars/forks/cordis`). The Deferred Register's bridge-module trigger is **1 of 3 MET** (remaining: a consumer states the reactive-composition requirement AND core v1.0.0 exit criteria shipped).
   - PapDashboard: latest still **v0.3.0** (no v0.3.1+). cqrs-htmx: latest still **v4.12.0** — integration pin and all AGENTS claims remain accurate.
5. **Same-train doc hygiene held throughout:** every mutation landed with its guards green; the TODO unreleased-deltas line now truthfully says "none" (twice updated: before and after the integration migration).

## b) PARTIALLY DONE

1. **BuildFlow full-mode verification — RUNNING, and it is MUTATING THE TREE.** Launched 10:23 as `env -u GOTOOLCHAIN buildflow --build-mode full` (shell 021), still running at report time with no summary out. Critical observation: **otel/exceptions_test.go, otel/otlp_test.go, otel/setup.go carry uncommitted modernization diffs** (`t.Context()`, named-return removal, range-over-func modernization) — authored by BuildFlow's fix-capable steps, NOT by me, and NOT yet verified by a post-mutation test/lint pass. They must be judged (diff review → otel suite → lint → commit-or-revert) AFTER the run exits — never left for the daemon to sweep unverified. My earlier claim "full mode is detect-only" was wrong (see d1).
2. **T29 — research 3/5 done, write-back 0/5.** cordis/PapDashboard/cqrs-htmx checked (a4); nixpkgs go version + buildflow version probes hung in a background shell (038) with no result — need a rerun with a timeout and a log file. None of the as-of dates are written back to AGENTS/TODO yet.
3. **#40 follow-through — HELD deliberately:** plan is bump security's example dep v0.6.0 → v0.7.0 + add the missing CHANGELOG `[Unreleased]` entry (covering the v0.6.0 auto-bump); held because a go.mod edit mid-BuildFlow would race its module-graph steps.
4. **#39 etag-replace drop — scoped, HELD:** the `replace github.com/larsartmann/go-etag => … v0.6.0` is still in go.work; with ALL satellite trains now shipped the graphs should carry etag ≥ the split, so the replace is droppable — but same mid-run race, held until BuildFlow exits.
5. **#43 /tmp cleanup:** the three guessed paths (`/tmp/core-old`, cspcheck, probe binaries) don't exist; a broader `/tmp` sweep was not performed.

## c) NOT STARTED

- **T30** — external-consumer proxy ritual: decision + recipe paragraph (policy gate — mine to draft).
- **T32** — rituals re-arm: cqrs cookbook re-verify trigger verify, core v1 exit-criteria graduation check, battery demand re-check.
- **T20** — AGENTS deep slim-down batch 1 (build commands → module READMEs; restores cap headroom at 375/377) + batches 2–3 (gotcha groups).
- **T21** — art-dupl suppression audit (~431 groups) + standing policy + Decision 13 cross-link.
- **T31** — lessons.md entry in the crush-config repo (commit there): "re-run guards between dependency-automation runs and tagging; check GOTOOLCHAIN/go.work parity before trusting build commands."
- **#35/#36/#37** — BuildFlow-finding triage (exhaustruct_v5.exclude config warning; 5 go-auto-upgrade findings; jscpd 13-findings policy) — all waiting on the full run's output.
- **Upstream drafts** — go-licenses ≥1.27 runner; go-structure-linter per-tool excludes.
- **Gated (untouched, as ordered):** T13 (G3), T15 (upstream filings), T16 (errorpages swap), T22 (pdg), T23 (prompt-crusher), T27 (govulncheck), T28 (browser CSP), T10.1 (owner SSH secret), battery waves W3–W5.

## d) TOTALLY FUCKED UP (own mistakes, no varnish)

1. **The worst one: I assumed BuildFlow full mode was detect-only and edited the tree concurrently.** I even wrote the reasoning down ("--fix is a separate step, so plain run doesn't write — safe to edit"). Disproven: the run is mutating otel/ right now. Nothing collided only because my concurrent edits were doc-only — the RISK was unbounded and the assumption was unverified. Rule that now exists: fix-capable automation runs on a QUIET tree, period; its write behavior is learned from evidence, not from a skill's step table.
2. **The otel mutations were discovered by luck, not process** — I only saw them because I ran `git status` while preparing THIS report. Without that accident the daemon could have swept modernization diffs into a commit UNVERIFIED. Detection should have been a scheduled check, not a side effect of writing a status doc.
3. **One bare `golangci-lint` invocation** (dropped the `GOTOOLCHAIN=go1.27.1` prefix in a command chain) → go 1.26.7 load failure, one wasted round trip. The exact footgun AGENTS documents, committed by me, again.
4. **Two edit-tool rejections on mtime-skewed files** (AGENTS.md, SUPERB v5) — I retried one blind before re-reading; both ended in python replaces. No damage, but the failure mode is known and I paid for it twice anyway.
5. **BuildFlow launched with output piped to `tail -60`** — zero intermediate visibility. A hang or early crash would be indistinguishable from "still running" (and indeed I stared at "no output" twice). Should have been `| tee /tmp/buildflow-full.log` from the start.
6. **`nix eval nixpkgs#go.version` ran in foreground, hung, got auto-backgrounded — no result to this day.** Two of five T29 items hang on it. A timeout + log file was the obvious shape from the start.
7. **A self-inflicted false alarm in my own verification:** a grep pattern with unescaped regex brackets reported the link fix "missing" when it was present and committed. Cost one extra verification round; the lesson is verify against the FILE, never against memory of what I typed.
8. **Forgotten until now:** #43's broader /tmp sweep; and an explicit "which files did BuildFlow touch" diff across ALL modules (I only found otel because it happened to be dirty — a full `git diff` baseline pass against f063f82 is still owed, f35).

## e) WHAT WE SHOULD IMPROVE

1. **Quiet-tree discipline for mutating automation.** BuildFlow (any mode with fix-capable steps) gets an exclusive tree. Check `git status` clean + no background automation before launching; the pre-tag gate pattern (mechanical, scripted) is the model — extend it with a "tree clean, no automation running" precondition.
2. **Every long-running command tees to a log file.** No more end-of-pipe blindness on multi-minute runs.
3. **Post-automation verdict loop as a ritual step:** automation ran → enumerate its diff → test → lint → commit-or-revert, explicitly, BEFORE any other work. Never let the daemon make that call.
4. **Script the watchlist sweep** (T29): one command hitting cordis tags, PapDashboard releases, cqrs-htmx tags, nixpkgs go, buildflow version, with as-of dates in the output. Interactive probing is how 2/5 items got left hanging.
5. **GOTOOLCHAIN-prefix discipline needs mechanical enforcement** — a shell alias/function or a guard in the recipe docs' commands; humans and agents both miss it under chain-composition pressure.
6. **Security CHANGELOG drift happened silently** (auto-bump without entry) — the release ritual's "verify CHANGELOG then" line exists precisely for this; it should be checked whenever BuildFlow/dependabot touch example deps, not only at tag time.
7. **The session proved the gate again:** five tags shipped today through `pre-tag-checks.sh`, zero frh-class incidents. The remaining gap is upstream (the v0.1.4 downgrade entered between guards and tag) — which the gate now closes; keep it mandatory in the ritual text (done) and in practice (done today, twice).

## f) NEXT (impact-ordered; ~40)

1. **Wait for BuildFlow 021 → read the summary** → triage the findings gate: #35 (exhaustruct_v5.exclude config-verify warning — find the module, fix key), #36 (go-auto-upgrade findings: lo.FromPtr in frh adapter, lo.Map/lo.Filter in otel/realtime tests — fix or policy-skip with rationale), remaining-findings exit code.
2. **Judge the otel diffs** (3 files): review each hunk → `GOTOOLCHAIN=go1.27.1 go test ./... -race` + golangci-lint in otel → commit as a modernization note or revert. Explicit verdict, before anything else touches the tree.
3. **Full-tree diff audit vs f063f82** — confirm otel was BuildFlow's ONLY mutation target (f35/d8).
4. **Drop the go.work etag replace** (#39): remove → `go work sync` → workspace build+test → guards → commit.
5. **Security example dep → v0.7.0 + CHANGELOG [Unreleased] entry** covering the silent v0.6.0 auto-bump (#40 follow-through); example build + test.
6. **Write back T29:** AGENTS Deferred Register cordis row (go/v0.1.0 TAGGED, trigger 1/3 MET, parts 2–3 open) + TODO watchlist as-of dates (PapDashboard v0.3.0, cqrs-htmx v4.12.0).
7. **Finish T29's last two probes** (nixpkgs go, buildflow version) with timeout + log.
8. **T30:** draft the external-consumer proxy ritual decision (proposal: verify-on-touch + quarterly cadence, never per-release; record as a numbered Decision in design-decisions.md).
9. **T32:** verify the cqrs cookbook re-verify trigger wording; read core-v1-exit-criteria.md and record graduation status against the 1.27.1 floor + v0.7.0 surface; battery demand re-check line in TODO.
10. **T31:** lessons.md entry in crush-config (commit THERE): "re-run guards between dependency-automation runs and tagging."
11. **Upstream draft:** BuildFlow go-licenses ≥1.27 runner ask (from the license-check skip rationale).
12. **Upstream draft:** go-structure-linter per-tool excludes / inert yaml `exclude_patterns` ask.
13. **T20 batch 1:** extract AGENTS per-module build commands into module READMEs (AGENTS links them); restores ≥10 lines of cap headroom.
14. **T20 batches 2–3:** gotcha-group extractions (otel pattern-propagation block → otel README; health NewProbe recorder-cliff → health README; …).
15. **T21:** art-dupl suppression audit (~431 groups, both episodes), standing policy, Decision 13 cross-link.
16. **#37:** jscpd policy — the per-module `.golangci.yml` duplication (277-line clones) is by-design; suppress with rationale or accept noise; record either way.
17. **#43:** broader /tmp sweep (trash, never rm).
18. **#45:** eyeball dependabot grouped config vs the four swept modules; confirm parity guard covers what it claims.
19. **f31 from this session:** figure out why the CI **proxy-smoke job skipped** on both pushes (0s) — intended SSH-gating or a broken condition; fix or document.
20. **Consider an AGENTS BuildFlow note** once b1 is confirmed: "full mode mutates the tree — quiet-tree rule" (prevents the next agent from repeating d1).
21. **Modernization-policy decision** (from b1): are BuildFlow's test-file modernizations wanted per-module, or noise to revert? One sentence in AGENTS or a buildflow skip.
22. **pkg.go.dev re-check next session** for flightrecorder v0.1.1 + docs v0.3.1 (crawler lag items from #49).
23. **After the next cqrs-htmx release:** re-verify their go-etag stub-replace drop (their move; D2 ritual; core v0.6.0+ unblocked them).
24. **otel ns/op re-baseline when machine load < 2** (T19 delta; allocs already identical, table annotated).
25. **AGENTS cap re-baseline after T20** (structure-linter budget).
26. ** Satellite-train policy:** security + otel now carry small unreleased deltas — batch them into the next natural train; no dedicated tag for doc-level drift.
27. **T13** the moment G3 is answered (g1): script or strip.
28–36. **Gated, do NOT start without the word:** T15 six upstream filings (go-health pack could be direct fixes in our repo); T16 errorpages `httputil.ResponseRecorder` swap; T22 pdg backlog sweep; T23 prompt-crusher merge resolution; T27 govulncheck (health + security); T28 browser CSP pass + THREAT_MODEL link; T10.1 SSH secret → T10.3 watch one green master run; battery waves W3–W5 (demand-gated); cordis bridge (trigger 1/3 now lit — see g3).
37. **Next docs-health pass:** archive/annotate this report per convention; the status tree is growing.
38. **Repeat the guards trio** (pin-drift, directives, parity) as the closing sweep once f1–f16 land.
39. **cqrs-lint re-run** if any cqrs-wrapper-adjacent change sneaks in (scorecard delta goes in the module CHANGELOG).
40. **Answer-bottleneck note:** ~10 of the items above collapse the moment the three §g questions are answered.

## g) QUESTIONS ONLY YOU CAN ANSWER

1. **Machine toolchain (standing):** is upgrading the machine's default go to ≥1.27.1 on your agenda (nix-managed, I can't touch it)? YES → I retire the entire GOTOOLCHAIN-prefix discipline (AGENTS command blocks, .buildflow.yml env, hook notes) in one follow-up train. NO → the prefix stays and keeps being documented. (My nixpkgs probe hung — I cannot answer this myself.)
2. **G3 (standing):** consumer-pin claims in AGENTS/FEATURES — KEEP them and I build `scripts/check-consumer-claims.sh` (re-verifies claims after each cqrs-htmx release), or STRIP exact consumer pins from AGENTS entirely? T13 is blocked on this word.
3. **cordis bridge promotion (NEW today):** `go/v0.1.0` is now TAGGED — trigger 1 of 3 for the Deferred Register's cordis bridge is MET. Do you want the bridge PROMOTED from the register to a tracked TODO/planned item (it would still wait for a consumer stating the reactive-composition requirement AND the core v1.0.0 exit criteria), or does it stay cold in the register until all three light up?

---

**Tag state at report time:** ten tags on origin through **core v0.7.0** — all gate-green, all proxy-verified. Guards: pin-drift 14 OK, directives 11/11 @ 1.27.1, structure lint green, dependabot parity 11/11. Working tree: otel's three BuildFlow-mutated files uncommitted (b1) — everything else swept.
