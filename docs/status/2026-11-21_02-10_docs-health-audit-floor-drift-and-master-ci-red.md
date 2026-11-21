# Status: Docs-Health Audit — Floor Drift Found + Fixed, Master CI Red Since 11-15 Diagnosed

**Created:** 2026-11-21 02:10 CET · Crush agent (docs-health AUDIT dispatch: "View ALL \*\*/2026-0\* files; living docs must be superb") · backlog basis: THIS session's run only; the only follow-on research was the CI blast radius of the drift I found (in scope — it is the standing net that should have caught it).

**Session shape:** docs-gates battery → 6 non-archived `2026-0*` corpus files → 7 living docs read → VERIFY claims-vs-code → fix drift → HARVEST → battery re-green. One critical repo-state find plus one critical repo-HISTORY find (unpushed/red CI) along the way.

---

## Headline findings

1. **CRITICAL (fixed in-tree):** four go.mods — root, `config`, `realtime`, `flightrecorderhealth` — sat at `go 1.27`, BELOW the 1.27.4 floor, and `check-go-directives` FAILED locally, since daemon commit `b04fc20` (2026-11-13 19:40) carried a gvac-normalize downgrade one hour after the floor train itself landed (`f722d04`, 16:10). The pre-commit guard never saw it (daemon commits bypass hooks by design). Restored all four to `go 1.27.4`; guard green; hermetic `build+vet+test -race` green on all four modules.
2. **CRITICAL (diagnosed, NOT fixed — needs push):** the drift commit is ON ORIGIN, and **every master push-triggered CI run since 2026-11-15 has FAILED — four consecutive reds (11-15 06:36, 11-15 11:36, 11-18 18:46, 11-20 10:21 = tip `2292838`), unattended.** `go-directives` fails from the drift; the same runs show `agents-deps`, `pin-drift`, `art-dupl-gate`, `fleet-lint`, `cqrs-lint`, all proxy-smokes, and many `test (./…)` jobs red (cascade from the drifted base and/or genuine — triage only possible post-push). The 2026-11-13 19:24 close-out certified "tip CI green" against a tree that predated the 19:40 drift commit by 16 minutes — the session-close ritual ("verify the FINAL pushed commit") was skipped by every subsequent session that let the daemon push.
3. **Local is AHEAD of origin by 4 commits** (my three fix sets + the concurrent SUPERB v17 session's TODO_LIST/plan work). Local suites, guards, and battery are green; origin CI is red. Nothing is pushed (no-push-without-ask rule).

## a) FULLY DONE (this session)

1. All 6 non-archived `2026-0*` files viewed and verdict-checked: composition-spike (swap shipped, §6 G1 sketch live), samber-do lazy-healthcheck draft (FILE-WORTHY, user-gated, premise re-verified 11-02), samber-do health architecture review (all 8 §7 items struck 11-13), PapDashboard integration research (§7 addendum, guard leg live), cordis/go-plugin-mvp research (trigger 1/3 met, re-verified 11-13), batteries gap analysis (canonical spec, triage header). `archived/` left closed per skill charter.
2. All 7 living docs read in full; release-state claims verified against git: **all 14 newest-tag claims match** (core v0.12.0 … preset v0.1.1), all 15 go.mod directives + go.work verified, lint-pin configs verified (13 satellites at `run.go: "1.27.4"`, root major.minor by design).
3. Floor drift FIXED: 4 go.mods restored to `go 1.27.4`; `check-go-directives` green; per-module hermetic verify green under load ~69.
4. README module table: httpx, config, preset rows ADDED — the table now lists all 14 released modules (was 11 of 14; missing since their 11-06/11-07 launches).
5. ROADMAP Raw-ideas de-staled: the shipped `httpx` and `config`/F1 rows reworded to SHIPPED facts (koanf-declined recorded); only the B8 route-seam idea and F4 hot-reload remain as ideas; header stamp updated.
6. AGENTS BuildFlow gotcha rewritten: stale skip_steps mechanics corrected (go_mod_ignore_dirs replacement), the falsified "cannot recur" claim replaced with the `b04fc20` recurrence + the floor-above-toolchain precondition explanation; `.buildflow.yml` comment corrected to match. AGENTS stays 155/220 lines.
7. Root CHANGELOG glued heading fixed (`### Changed- Repo-level…`).
8. HARVEST executed: **go-health restart/rearm ask** (never filed, owner-gated — the recorder-sentinel sibling went upstream as go-health#7 and closed) minted as a new P3 row; **A5 PapDashboard security swap** made matchable ("battery A5" now in the battery-swap row) and its unharvestable archived item closed with `<!-- fresh:skip -->` + route recorded (all 5 of its tokens are DF≥8 generic; the tokenizer strips "A5" itself — the item can never match mechanically).
9. Battery discipline: first restart/rearm draft failed pre-flight (my "sentinel"/"file"/"gate" wording would have genericized an already-harvested item); reworded to SAFE before landing. All new living-doc lines pre-flighted.
10. `.docs-gates-preflight.json` added to `.gitignore` (the battery's own note).
11. Final state: battery **5/5 PASS** (freshness, audit-all, annotate, routing, cadence); **10/10 offline guards PASS** (features-claims, status-index, routed-destinations, dependabot-parity, workspace-charter, agents-deps, sse-protocol, module-licenses, go-directives, golangci-configs); dprint repo-wide green; AGENTS cap respected.

## b) PARTIALLY DONE

1. **Master CI repair:** root cause fixed locally, but the fix is among the 4 UNPUSHED commits — origin CI stays red until push + post-push verification. Deliberately halted at the push boundary (never push without explicit ask).
2. **Session-close ritual:** diagnosis done (push lag + per-run verdicts + job list captured), but the ritual itself (push → verify tip green → confirm no new-job-without-live-run) is incomplete by design pending your instruction.
3. **VERIFY external-claims leg:** existing dated external claims accepted as within cadence (watchlist re-stamped 11-07; consumer-claims guard green last run); NO external state re-verified this session (go-sse v0.7.0 hold, cqrs-htmx latest, nixpkgs go 1.26.10) — 8 days stale at next touch.
4. **TODO_LIST header provenance:** my harvest rows carry inline provenance; no header block added — the concurrent v17 session owns today's header block and I did not clobber it.

## c) NOT STARTED (discovered this session, untouched)

1. Post-push red-job triage: `agents-deps`, `pin-drift`, `art-dupl-gate`, `fleet-lint`, `cqrs-lint`, proxy-smokes, and the failing `test (./…)` jobs each need a root cause once the base heals (some will be cascade; dependabot's 11-18 go-etag v0.6.2 bump and the 11-20 PR bumps may carry genuine failures).
2. README module-table parity guard — no standing gate computes the module list vs the README table; that is the class that let 3 modules go unlisted for 8+ days.
3. docs-gates `--margins` proactive run (near-bar token report) — never run here.
4. Archive cadence: docs/status/ held 14 unarchived reports (threshold 15) with oldest at 14 days (threshold 29); this report makes 15 — the annotate+archive wave for the five 2026-11-07 reports is now due.
5. Deep verify of `docs/agents/module-details.md`, `docs/TELEMETRY.md`, module READMEs, integration README — outside the 7-doc scope of this dispatch.
6. The restart/rearm upstream DRAFT FILE — still not drafted even as a draft (row minted; filing and drafting are owner-gated).

## d) TOTALLY FUCKED UP (state as found — plain)

1. **Master CI red ~6 days, four failed push runs, zero human/agent attention.** The one guard the architecture trusts for daemon drift (CI, because hooks are bypassed) was screaming and nobody looked. Every session since 11-13 that let the daemon push skipped the close-out ritual.
2. **The floor train damaged itself:** `b04fc20` — the same day's own daemon ride-along — reverted 4 of the 15 go.mods the train had just raised. The 19:24 close-out report certified green ~16 minutes before the damaging commit existed.
3. **`respect_patch_floor` protection is structurally void during floor trains** (floor 1.27.4 > machine default 1.27.2): this class has now bitten twice (gvac v0.1.4 era; 2026-11-13). Current "fix" is documentation + manual restore; real prevention is an unmade policy decision (see questions).
4. **PR queue blocked on the rotten base:** the 11-20 dependabot PRs run CI off the drifted master — a wall of red PRs mixing cascade noise with possibly-real failures.
5. **The freshness model punishes doc improvement:** my truthful new living-doc lines (README/ROADMAP/AGENTS edits) flipped 2 historical items to DRIFT mid-pass. Recovered cleanly (harvest + fresh:skip), but the corpus's health depends on token budgets nobody watches between passes (`--margins` exists for exactly this and is never run).

## e) WHAT WE SHOULD IMPROVE (what I forgot / could do better / systemic)

1. **I forgot to check CI FIRST.** When `check-go-directives` failed locally, I fixed and moved on — the assumption "CI would have caught it" was wrong in the worst direction: CI HAD caught it, six days of red, unwatched. Next time a local guard fails on a daemon-carried drift, `gh run list` is step two, not an afterthought.
2. **Mechanize the session-close ritual.** Four skipped rituals in six days means memory is the only enforcer. A `scripts/check-tip-ci.sh` (push-state + last push-run verdict + red-job list) wired into the post-commit watch or a cron closes the gap owner-side.
3. **Count-derived docs beat eyeballs.** I caught the README module-table gap only on a second read; the fleet's own rule is "prefer a standing gate that re-derives doc numbers." A module-table parity guard is cheap and mechanical.
4. **Pre-flight ALL new living-doc lines, not just TODO rows.** I pre-flighted the 2 TODO candidates but not the README/ROADMAP/AGENTS edits — those caused the audit-all red. The battery's additions-only model makes pre-flight unconditional.
5. **Daemon blind spot is structural, not incidental:** hooks bypassed → guards live in CI → CI watching is voluntary. A machine-level cron watchdog for the cheap guard scripts is the only closure that does not depend on session discipline.
6. **Lint and network guards were NOT run this session** (no Go source changed; no release) — stated explicitly so the next session does not inherit "green" assumptions.
7. Minor: I hand-placed the `<!-- fresh:skip -->` marker (the writer has no subcommand for it) — precedent exists (2026-11-13 hand-strike), but a writer subcommand would keep the no-hand-rolling rule intact; upstream-notes candidate.

## f) Up to 50 things to get done next

**P0 — the red CI (blocked on your word):**
1. Push master (4 commits: floor fix ×1 ride, doc fixes, v17 rows) — heals `go-directives` at the root.
2. Verify tip CI green; expect `agents-deps` + `art-dupl-gate` to flip green with the base.
3. Triage every still-red job post-push (pin-drift, fleet-lint, cqrs-lint, proxy-smokes, test jobs) — separate cascade from genuine.
4. Re-base/refresh the four blocked dependabot PRs (11-20) once master is green; re-verify the 11-18 go-etag v0.6.2 bump run's failures separately.
5. Confirm the daemon-carried commits on origin byte-match the verified tree (daemon snapshot races dprint twice today).

**P1 — guards & rituals (this week):**
6. `scripts/check-tip-ci.sh` + wire into post-commit watch (mechanize the close-out ritual).
7. README module-table parity guard (computed from go.mods; all three venues).
8. Floor-train prevention policy (your decision — question 2) then implement.
9. Machine-cron watchdog for cheap guard scripts (directives, charter) — independent of sessions.
10. docs/status archive wave: annotate + `check-rows --strict` + `bare-rows` + archive the five 2026-11-07 reports; re-baseline cadence.
11. Run `docs-gates --margins`; reword near-bar tokens proactively.
12. External-claims re-verify sweep at next touch (go-sse hold, cqrs-htmx, nixpkgs go).
13. Writer subcommand for `fresh:skip` placement (or fold into the upstream notes draft).
14. Add "expected-red annotated" treatment for the designed-red jobs on the PR branch view (ci-preflight/structure-linter show as noise in PR reds).

**P2 — standing queue (unchanged owners; do not re-derive):**
15. CI `SSH_PRIVATE_KEY` secret restore (P1, owner-only).
16. Owner decision batch D1 (D014 wire format) + D2 (art-dupl cadence/stale entries).
17. Auth-bridge decision memo (a/b/c).
18. Scaffold: create `go-service-template` sibling repo + spec + one-real-app dogfood probe.
19. SUPERB v17 default-onramp package (concurrent session's plan — coordinate, don't duplicate).
20. Preset six-month adoption re-sweep (direct 7/41 baseline).
21. pma dossier §6 owner steps (cooldown, staged-respect, ignore-patterns).
22. Machine ops: /tmp `go-build*` cleanup policy; BuildFlow binary rebuild (2d6bbee); cache-DB VACUUM; `interrogate` install.
23. project-dependency-graph polish list (their repo).
24. Upstream watchlist: go-sse v0.7.0/cqrs-htmx v4.16/templ-components absorb trigger; go-etag v1.0.0 plan; go-idempotency CHANGELOG hygiene.
25. W4/W5 batteries (worker supervisor, sqlite ops kit, polite client, do bridge; W5 remainders).
26. B6/B8 behind the H8 route-seam owner ask (open since 11-07).
27. v1.0.0 exit criterion 5 (consumer version overlap) — nudge or wait.
28. restart/rearm upstream ask — draft + file on your go (my new P3 row).
29. A5 PapDashboard security swap recipe instance (consumer repo).
30. dnsblockd / ChastityAPI / pbx-lens battery-swap instances.
31. SigNoz live validation reclassify (when an instance exists).
32. OTLP logs gate re-run on next consumer-claims refresh.
33. Realtime subscribe-before-flush (demand-gated).
34. cqrs EventConfig opt-ins (encryption/signing; demand-gated).
35. samber/do: auditlog example (demand) + lazy-healthcheck filing (owner approval).
36. cordis bridge — remaining 2 of 3 triggers.
37. `security.APIKeyMatches` re-review on security's next auth-touching train.
38. Watchlist quarterly items (external consumers verify-on-touch).
39. config: informational windows-only x/sys finding rides config's next train.
40. frh drops hand-rolled cooldown when go-flightrecorder#2 ships upstream.
41. go-aichat/realtime convergence watch (ssehub `Drops()` as the W5-C1 reference).
42. TODO_LIST micro-backlog owner rulings (plan-revision convention; "rev N (expanded)").

## g) Questions I cannot answer myself

Asking below — push authorization, the floor-train prevention policy, and the concurrent v17 session's state.

---

*Nothing in this report re-derives the standing backlog (TODO_LIST/ROADMAP remain the canonical queues); everything above a)–e) is from this session's run.*
