# cqrs v0.7.0 Follow-ups Execution + Charter-Guard First Fire — Status

> 2026-10-07 01:20 · session scope: the (f) follow-up list from
> `2026-10-07_00-40_cqrs-v0.7.0-hardening-followups-status.md`, full-execution directive.
> Prior session context: cqrs v0.7.0 shipped+proxy-verified; this session closed its follow-ups.

---

## a) FULLY DONE

1. **Baseline verified** — master == origin at start; all four guards green (pin-drift, go-directives, workspace-charter, config-parity/dependabot); structure linter 0 findings.
2. **Recipe↔Example sync-guard** (`cqrs/readme_sync_test.go`, `d6b83b3` + README refs via daemon `2d991a9`): fails when a README recipe references an API the module does not carry (`cqrs.*` must be exported; exercised dep-prefixes must appear in module code — would have caught both shipped README defects, including the historical `metaengine.WithReplay` invention) or when an Example and its README reference drift apart in either direction. **Negative-tested**: injected both defect classes, both fail, restore passes. README now names all 5 Examples at their recipes.
3. **cqrs CHANGELOG `[Unreleased]` bookkeeping** (`73e95d7`): timers tripwire, restart persistence, coeffect advisory, SSE replay reconnect, the 2 executable Examples + README fixes, the sync guard. Test/doc-only — no tag needed.
4. **Dual-checkpoint-knob pre-v1 decision note** (`doc/planning/2026-10-07_cqrs-dual-checkpoint-knob-decision.md`, `85c72fd`): both knobs stay, wrapper wins (disjoint consumer populations; uniform merge contract); WARN-when-both-set is the mitigation candidate; revisit triggers written.
5. **Deep-dive rubric re-run: 52 → 87/100** (`927b14c`): dated addendum section in the deep-dive HTML with per-cluster old→new rationale; hero tag + scorecard + finding banners updated (all 4 majors + doc nits marked resolved at v0.7.0); DoD ≥80 MET; both residuals upstream-bounded (Timers stop gap, GracefulClose early return). The tripwire test comment now enumerates all FOUR files of the upstream-flip procedure (was 3 — made the report's claim true rather than editing the claim).
6. **AGENTS.md pass 1** (`d1eafa3`): workspace-charter guard wiring, LSP-lint-is-not-ground-truth (re-verified on example_test.go this session), buildflow `GOWORK: "off"`, cqrs follow-ups (rubric, sync guard, knob decision). File stayed 361/377 counted lines; structure linter green.
7. **AGENTS release-truth sync** (`1725cca`): DISCOVERED the concurrent session's family train — **otel v0.2.0, flightrecorder v0.2.0, health v0.1.5, frh v0.1.6, realtime v0.1.3, errorpages v0.1.1, all 2026-10-06** — tagged on origin (ls-remote-verified) while AGENTS still said "pending unreleased train" (their same-train ritual step 3 missed). Fixed the release line + all six module bullets from their CHANGELOGs (source of truth).
8. **Pin-drift guard dual-pattern** (in `1725cca`): accepts the old `ON ORIGIN through **core vX.Y.Z**` and the new `Latest per module: core vX.Y.Z,` phrasings — the release-line check survives the rewording it originally failed on.
9. **docs-health HARVEST** (`f561048`): TODO_LIST header re-dated 2026-10-07 with this pass; **removed 4 completed items** (~8.5KB: both v0.2.0 release trains [shipped], the go.work-tracked question [resolved + charter-guarded], the `[x]` cqrs deep-dive mega-item [rule: no `[x]` items; standing watch kept as its own entry]); added P2 entries (charter-job observation, daemon root-cause, CI resilience) + the standing upstream-flip watch; stale claims fixed (cqrs E2E leg CLOSED via the two integration tests; AGENTS 361/377 not 376/377); open-questions cluster updated (go.work resolved; upstream-ask filing + buildcache topology added).
10. **workspace-charter CI job observed — first live fire CAUGHT A REAL LEAK, then went green.** The run on `85c72fd` failed the job: the auto-commit daemon had smuggled `./integration` back into go.work (leak #3, inside `2d991a9` together with my README). Fixed in `ae3b2d4`; the run on that push shows `workspace-charter = success`. Guard-as-code beat guard-as-ritual exactly as designed. (Overall run conclusion stays "failure" while the ssh-gated matrix is dead — see d/g.)
11. **CI resilience shipped** (`67c2a61`): `ci-preflight` job fails ONCE with restore instructions when `SSH_PRIVATE_KEY` is empty; the 12-job matrix `needs:` it and is skipped instead of failing 12 identical times; guards + proxy-smoke stay independent. YAML-validated (all 8 jobs parse; needs-edge correct).
12. **Daemon root-cause — facts established**: the go.work mutation is a one-line sorted insert (`go work use -r .` signature); **BuildFlow gomod-check EXONERATED** (ran `-s gomod-check --fix` — go.work unchanged); the daemon bundles unrelated tree mutations into heuristic commits (`2d991a9`: my README + its go.work edit); daemon pushes stalled 22:47Z→01:1x (no process visible mid-session) then resumed.

## b) PARTIALLY DONE

1. **Daemon root-cause** — mechanism and damage modes are established (see a12), but the SOURCE (its config/pipeline step that runs the workspace heuristic) is not reachable from this shell: no process mid-session, no `~/.config/pma`, `systemctl` blocked for agents. Owner input required (g1).
2. **Rubric re-run honesty margin** — 87/100 is re-derived per-cluster with evidence links, but cluster leverage values (95/90/85/75/85/100) are judgment calls, not measurements; the two residuals cap economy/safety until upstream moves.
3. **CI preflight verification** — committed and YAML-valid, but its first origin run hasn't happened yet (unpushed at report time); the matrix-skips-one-red behavior is designed, not yet observed.

## c) NOT STARTED

1. otel×realtime SSE-through-OuterMiddlewares E2E leg (optional; already on the composition-gap watchlist).
2. FEATURES.md VERIFY pass (cqrs v0.7.0 rows, the two v0.2.0 module rows) — out of this session's scope; the TODO carries no explicit item, worth adding on the next docs pass.
3. cqrs-lint scorecard re-run — deliberately skipped: no wrapper API changed this session (test/doc-only); the ritual triggers on the next cqrs feature train.
4. core v0.8.0 + systemd v0.1.0 release train (unchanged; steps live in TODO P3).

## d) TOTALLY FUCKED UP

1. **Pipe-masking REPEATED — the exact prior-session d1 failure**: `./scripts/check-pin-drift.sh 2>&1 | head -3 && git add … && git commit …` — the guard FAILED (my release-line rewording broke its grep) but `head`'s exit 0 let the commit (`bb47fbc`) land while the guard was red. The countermeasure (`set -o pipefail`) was already written in AGENTS lore and I still didn't apply it in that chain. Caught minutes later; fixed via the guard's dual-pattern (a8). Root lesson now moved from "know it" to "redirect-to-file, never pipe verify commands" (e1).
2. **My commit `bb47fbc` VANISHED from master mid-session**: created 00:52, by 01:06 HEAD was back at `d1eafa3`, the commit's content sat STAGED, the worktree files were REVERTED to pre-fix content, and the reflog shows NO reset entry (raw .git surgery — the daemon is the prime suspect; it was mid-flight on those files). Recovered cleanly via `git restore --worktree` from the index (my own staged work, data safe) and re-committed as `1725cca`. Mechanism unexplained (b1, g1).
3. **Lost the failing-test name of a transient cqrs failure**: one `go test -race` run failed mid-session; its output went through `tail -1` so only "FAIL" survived — three pipe-incidents in one session, this one cost information. Three consecutive re-runs (plus two earlier) green; the flake is unidentifiable now (suspects: timing-sensitive SSE replay / timers tests under load). Watch item (f11).
4. **Truncation disease in observation too**: first CI job-list check used `sort | head -25` — `workspace-charter` sorts last and was silently cut, briefly convincing me the job "didn't run". The follow-up query-by-name found it FAILED (the leak catch). Same class as d1/d3: pipes eating tails.

## e) WHAT WE SHOULD IMPROVE

1. **Never pipe verify commands** — redirect to a file, then grep/inspect the file (the final verification block did exactly this and it worked). `pipefail` is not enough; d3 failed WITH information loss even under pipefail semantics. Proposal: make it an AGENTS Build & Verify rule.
2. **The daemon now has three proven damage modes**: (a) `./integration` re-add (3x — guard catches it, proven live); (b) bundling unrelated mutations into heuristic commits (`2d991a9`: my README + its go.work line); (c) rewinding a commit with staged-content preservation (`bb47fbc`). Nothing guards (b)/(c). Minimum: a push-lag/daemon-health watch (origin age > N hours → warn) + a post-commit HEAD-advance sanity hook candidate.
3. **Release Ritual step 3 has no mechanical enforcement** — this session's six-module AGENTS drift (a7) is the same class the ritual step 3 text warns about. Extend `check-pin-drift.sh` with per-module "AGENTS says vX vs tag says vY" checks so the miss fails a guard, not a session.
4. **Concurrent-session coordination worked** (their modules/pins untouched; only shared docs corrected) — keep the TODO_LIST/AGENTS-only coordination pattern.
5. **Good bones worth keeping**: the charter guard's first live fire caught a real leak within one push of its landing; the index-recovery of bb47fbc was clean; the sync guard's negative test proved both target defect classes fail loudly.

## f) Top things to get done next (ranked)

1. **Push the last commit** (`67c2a61`, ci-preflight) — daemon was at it; verify it lands (origin was 1 behind at report time).
2. **Observe the ci-preflight job's first run** on that push: ONE red job with the restore message, matrix skipped — the designed behavior.
3. **pma daemon triage (owner)**: where it lives, why pushes stalled 22:47Z→01:1x, what rewound bb47fbc, which step runs the workspace heuristic (g1).
4. **Extend check-pin-drift.sh per-module** AGENTS-vs-tag checks (mechanize ritual step 3; this session's drift class dies permanently).
5. **Push-lag / daemon-health watch**: local guard or hook warning when origin/master age exceeds N hours (the stall went unnoticed ~30min mid-session; the 6-day CI blindness is the same class at larger scale).
6. **SSH_PRIVATE_KEY restore (owner, P1)** — with ci-preflight shipped, the blind spot is at least loud now.
7. **File the go-cqrs-lite upstream ask (owner-gated)** — two verified findings + consumer tripwire; draft complete.
8. **/mnt/buildcache topology answer (owner)** — recurrence risk for the corruption class.
9. **cqrs-htmx setup Domain adoption** — their next train, owner-gated cross-repo.
10. **Flake hunt**: rerun the cqrs suite `-count=10` under load once; if the transient recurs, capture the name (this session lost it to d3).
11. **Standing upstream-flip watch** (system GracefulClose + timer-stop): the 4-file change is enumerated in the `cqrs/domain_test.go` tripwire comment.
12. **FEATURES.md VERIFY pass** (cqrs v0.7.0 + otel/flightrecorder v0.2.0 rows).
13. **core v0.8.0 + systemd v0.1.0 train** (steps in TODO P3; demand-proven by bank-sync ADR-017).
14. **Example StartHooks migration** post core v0.8.0 (flightrecorder + otel examples; TODO P3).
15. **fr v0.2.1 sweep** on the frh + cqrs next trains (safe-but-don't-bump-casually; TODO P3).
16. **AGENTS deep slim-down** — 361/377 again after this session's additions; every future addition needs a removal until the structural cut lands.
17. **Composition-gap E2E watchlist** items (otel×realtime first; existing TODO entry).
18. **BuildFlow dprint exit-14 upstream fix** (standing; escape hatch documented).
19. **Health/polish backlog batch** (existing TODO P3 smalls — untouched, standing).
20. **AGENTS rule addition from e1** (never pipe verify commands; redirect to file) — one line, next AGENTS touch.

## g) Questions I can NOT figure out myself

1. **Where does the pma daemon live, and may I read its config/logs?** It pushed for the prior session, stalled 22:47Z→01:1x, re-added `./integration` to go.work three times, bundled its mutations into my staged commits (`2d991a9`), and performed the unexplained `bb47fbc` rewind (reflog-silent .git surgery). From this shell: no process mid-session, no `~/.config/pma`, `systemctl` blocked. I need its unit name/path or a log directory to finish the root-cause.
2. **May I push?** One commit (`67c2a61`, ci-preflight) was unpushed at report time and the daemon's push timing is unpredictable — say the word and I'll `git push origin master` (no force, ever). Related: do you want the push-lag watch (f5) as a guard script?
3. **SSH_PRIVATE_KEY (standing P1)**: still empty at repo level; the entire matrix has been dead since ~2026-10-01. With this session's ci-preflight, the failure is now ONE loud red job with instructions instead of 12 masked ones — restoring the secret is the only remaining step to a fully green master.

---

_Session commits (daemon pushes; at report time all but `67c2a61` were on origin): `d6b83b3` sync guard · `73e95d7` CHANGELOG · `85c72fd` knob decision · `927b14c` rubric re-run · `d1eafa3` AGENTS pass 1 · `1725cca` release-truth + guard dual-pattern (recovery of the rewound `bb47fbc`) · `f561048` TODO_LIST harvest · `ae3b2d4` go.work leak #3 fix (charter-guard catch) · `67c2a61` ci-preflight. Daemon: `2d991a9` (README refs + leak #3), push stall, then resumption._
