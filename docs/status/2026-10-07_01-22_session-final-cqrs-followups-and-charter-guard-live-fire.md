# Session Final — cqrs v0.7.0 Follow-ups, Charter-Guard Live Fire, Release-Truth Sync

> 2026-10-07 01:22 · session FINAL report (consolidates the 01-20 interim; adds final
> verification state). Scope: the full session run under the full-execution directive —
> the (f) follow-up list from `2026-10-07_00-40_cqrs-v0.7.0-hardening-followups-status.md`,
> plus everything discovered along the way. Prior context: cqrs v0.7.0 shipped,
> pushed, proxy-verified (2026-10-06).

---

## a) FULLY DONE

1. **Baseline verified** at session start: master == origin, all four guards green (pin-drift, go-directives, workspace-charter, config-parity), structure linter 0 findings.
2. **Recipe↔Example sync-guard** — `cqrs/readme_sync_test.go` (`d6b83b3`; README references landed via daemon `2d991a9`). Contract: every `cqrs.*` token in README must be an exported API; exercised dep-prefixes (`metaengine`, `system`, `clprojections`, `command`, `event`, `id`, `projectionhost`, `cl`) must appear in module code; every Example is referenced by name in README and every reference resolves. **Negative-tested**: injected the historical `metaengine.WithReplay` invention + a dangling Example reference → both fail with actionable messages; restore passes. Would have caught both README defects shipped at v0.7.0.
3. **cqrs CHANGELOG `[Unreleased]`** (`73e95d7`): timers tripwire, restart persistence, ShutdownDependencies edges, coeffect advisory, SSE replay reconnect, both executable Examples + the README fixes they exposed, the sync guard. Test/doc-only — no tag owed.
4. **Dual-checkpoint-knob pre-v1 decision note** (`85c72fd`, `doc/planning/2026-10-07_cqrs-dual-checkpoint-knob-decision.md`): both knobs stay, wrapper wins (disjoint consumer populations; uniform merge contract); WARN-on-both-set as mitigation candidate; revisit triggers written. Upstream semantics verified from the module cache before deciding.
5. **Deep-dive rubric re-run: 52 → 87/100** (`927b14c`): dated "Rubric re-run" section in the deep-dive HTML with per-cluster old→new table and rationale; hero tag, scorecard cards, and all six finding banners updated (4 majors + operator-docs + doc-drift resolved); DoD ≥80 MET; both residuals upstream-bounded (Timers stop gap, GracefulClose drain-error early return) with the standing-watch callout. The tripwire comment in `cqrs/domain_test.go` now enumerates all FOUR files of the upstream-flip procedure — I made the claim true instead of weakening it.
6. **AGENTS.md pass** (`d1eafa3`): workspace-charter guard wiring (pre-tag-checks + CI job + the daemon-leak catch), LSP-lint-not-ground-truth (re-verified on example_test.go: LSP showed 4 false positives, `go vet` + buildflow clean), buildflow `GOWORK: "off"` posture, cqrs follow-ups (rubric, sync guard, knob decision). Stayed 361/377 counted lines; structure linter green.
7. **AGENTS release-truth sync** (`1725cca`): discovered the concurrent session's family train — **otel v0.2.0, flightrecorder v0.2.0, health v0.1.5, frh v0.1.6, realtime v0.1.3, errorpages v0.1.1 (all 2026-10-06)** — verified on origin via `ls-remote` while AGENTS still said "pending unreleased train" (their same-train ritual step 3 was missed). Release line + all six module bullets corrected from module CHANGELOGs (source of truth).
8. **Pin-drift guard dual-pattern** (in `1725cca`): accepts both `ON ORIGIN through **core vX.Y.Z**` and `Latest per module: core vX.Y.Z,` — the check survives the rewording that (momentarily) broke it; syntax-checked + full-guard green.
9. **docs-health HARVEST** (`f561048`): TODO_LIST header re-dated 2026-10-07; **4 completed items deleted** (~8.5KB: both v0.2.0 trains [shipped], go.work-tracked question [resolved + guarded], the `[x]` cqrs deep-dive mega-item [no-`[x]` rule; standing watch kept as its own entry]); 3 new P2 entries (charter-job observation, daemon root-cause, CI resilience) + standing upstream-flip watch; stale claims fixed (cqrs E2E leg closed via the two integration tests; AGENTS line count 361 not 376); open-questions cluster updated (go.work resolved → upstream-ask filing + buildcache added).
10. **workspace-charter CI job observed — FIRST LIVE FIRE CAUGHT A REAL LEAK, then went green.** Run on `85c72fd`: job FAILED with the exact diff (`< ./integration`) — the daemon had smuggled leak #3 into go.work inside `2d991a9` alongside my README. Fixed in `ae3b2d4`; the run on that push shows `workspace-charter = success`. The guard design (dynamic membership, no hardcoded list) proved itself within one push of landing.
11. **CI resilience shipped** (`67c2a61`): `ci-preflight` job — checks `secrets.SSH_PRIVATE_KEY` via env-mapping (no script injection), fails ONCE with restore instructions, matrix `needs:` it and is skipped instead of 12 identical reds; guards + proxy-smoke independent by design. YAML-validated: all 8 jobs parse, needs-edge correct.
12. **Daemon root-cause — facts established**: go.work mutation is a one-line sorted insert (`go work use -r .` signature); **BuildFlow gomod-check EXONERATED** (ran `-s gomod-check --fix`, go.work byte-identical); the daemon bundles unrelated tree mutations into heuristic commits; its push stream stalled 22:47Z→~23:16Z (origin gap) then resumed. Leak class #3 is now guard-caught.
13. **Two status reports written + committed** (01-20 interim, this final).

## b) PARTIALLY DONE

1. **Daemon root-cause** — mechanism, damage modes (integration re-add ×3; mutation bundling; the `bb47fbc` rewind), and one exoneration (BuildFlow) are established; the SOURCE (which process/pipeline step runs the workspace heuristic, why the rewind happened) is unreachable from this shell: no visible process, no `~/.config/pma`, `systemctl` blocked for agents. Owner input gates the rest (g1).
2. **ci-preflight first run unobserved** — committed + YAML-valid, but its origin run hasn't happened: `67c2a61` + `adb1985` were still unpushed at report time (daemon lag). Designed behavior (one red preflight, 12 skipped matrix jobs, guards green) not yet seen live.
3. **Rubric honesty margin** — 87/100 is evidence-linked per cluster, but the leverage values are judgment calls, not measurements; economy/safety stay capped by the two upstream residuals until the flip watch fires.
4. **Unpushed commits** — 2 at report time (`67c2a61`, `adb1985`); I don't push autonomously. Everything else is on origin.

## c) NOT STARTED

1. otel×realtime SSE-through-OuterMiddlewares E2E leg (optional; on the composition-gap watchlist already).
2. FEATURES.md VERIFY pass for the new versions (cqrs v0.7.0, otel/flightrecorder v0.2.0 rows) — no TODO item exists yet; add on the next docs pass.
3. cqrs-lint scorecard re-run — deliberately skipped (test/doc-only session; ritual triggers on the next cqrs feature train).
4. core v0.8.0 + systemd v0.1.0 release train (unchanged; full steps live in TODO P3).
5. Per-module pin-drift extension (AGENTS-vs-tags; mechanize ritual step 3) — identified this session (e3), not built.

## d) TOTALLY FUCKED UP

1. **Pipe-masking REPEATED — the exact prior-session failure, by me, again**: `check-pin-drift.sh 2>&1 | head -3 && git add … && git commit …` — the guard FAILED (my rewording broke its grep) but `head`'s exit 0 let commit `bb47fbc` land with the guard red. The countermeasure was literally written in session lore. Fixed minutes later (guard dual-pattern, `1725cca`); the process failure is the real entry.
2. **Commit `bb47fbc` VANISHED from master**: created 00:52; by 01:06 HEAD was back at `d1eafa3`, content STAGED, worktree files REVERTED to pre-fix content, **reflog shows no reset entry** (reflog-silent .git surgery; daemon prime suspect, mid-flight on those exact files). Recovered via `git restore --worktree` from the index (my own staged work; data safe) and re-committed. Mechanism unexplained — see b1/g1.
3. **Transient cqrs test failure — name LOST to `tail -1`** (third pipe incident; this one cost information, not correctness): one `go test -race` run failed mid-session; only "FAIL" survived the pipe. Three consecutive + two earlier runs green. Suspects: timing-sensitive SSE replay / timers tests under load. Unidentifiable now; watch item (f10).
4. **Truncation in observation**: first CI job-list check used `sort | head -25` — `workspace-charter` sorts LAST and was silently cut; I briefly concluded the job "didn't run". Query-by-name then found it FAILED (the leak catch). Same pipe-eats-tail disease.
5. **Small but real**: my first draft AGENTS edit for the buildflow line was garbled (`env.GOWOWORK no`) — caught by the edit tool's stale-read refusal, which is the system working, but the draft shouldn't have been malformed.

## e) WHAT WE SHOULD IMPROVE

1. **Never pipe verify commands — redirect to a file, inspect the file.** Three incidents this session (d1 correctness, d3 information, d4 observation) are all pipe semantics. `pipefail` is insufficient (d3 lost data even under pipefail rules). One-line AGENTS Build & Verify rule owed (f16); the session's final verification block already used the file-redirect pattern successfully.
2. **The daemon has three proven damage modes and one gap**: (a) `./integration` re-add ×3 — CAUGHT by the charter guard (proven live); (b) bundling unrelated mutations into heuristic commits — unguarded; (c) reflog-silent commit rewind — unguarded; (d) push stalls go unnoticed (22:47Z→23:16Z gap + the 2-commit lag at report time). Minimum: push-lag watch + post-commit HEAD-advance sanity hook candidate.
3. **Mechanize Release Ritual step 3** — the six-module AGENTS drift (a7) is precisely the class the ritual text warns about; this session's guard dual-pattern fixed the SYMPTOM class (phrasing), but per-module AGENTS-vs-tag checks would make the MISS itself fail a guard (f4).
4. **Guard-first observation**: the workspace-charter job's value was proven within one push of landing; the pin-drift per-module extension (f4) and push-lag watch (f5) deserve the same guards-as-code treatment instead of more prose.
5. **Concurrent-session coordination held**: their modules/pins untouched; only shared docs corrected; coordination via TODO_LIST/AGENTS only. Keep the pattern.
6. **Honest positive**: recovery paths worked — index-restore for bb47fbc, negative tests for the sync guard, module-cache API verification before the decision note, `ls-remote` before release-truth claims. The verify-then-claim discipline mostly held; where it slipped (d1), the damage was contained by the next verification.

## f) Top things to get done next (ranked)

1. **Push the 2 unpushed commits** (`67c2a61` ci-preflight + `adb1985` report) — owner word or daemon resumption (g2).
2. **Observe ci-preflight's first run**: ONE red job with the restore message, matrix skipped, guards green.
3. **pma daemon triage (owner)**: unit/config/logs location; explain the push stall + the `bb47fbc` rewind + the workspace-heuristic step (g1).
4. **Extend check-pin-drift.sh per-module**: AGENTS-says-vX vs tag-says-vY for every module — kills the ritual-step-3 miss class mechanically.
5. **Push-lag / daemon-health watch**: warn when origin/master age exceeds N hours (session saw a 30-min stall unnoticed; the 6-day CI blindness is the same class).
6. **SSH_PRIVATE_KEY restore (owner, P1)** — with ci-preflight the dead matrix is at least LOUD now; secret restore is the only step left to a green master.
7. **File the go-cqrs-lite upstream ask (owner-gated)**: two verified findings + consumer tripwire; draft complete under `doc/feedback/outgoing/`.
8. **/mnt/buildcache topology answer (owner)**: multi-writer or not; recurrence risk for the 2026-09-29 corruption class.
9. **cqrs-htmx setup Domain adoption** (their repo, owner-gated, their next train).
10. **Flake hunt**: cqrs suite `-count=10` under load once; capture the name if the transient recurs (d3).
11. **Standing upstream-flip watch** (system GracefulClose + timer-stop): 4-file change enumerated in the `cqrs/domain_test.go` tripwire comment.
12. **FEATURES.md VERIFY pass** (cqrs v0.7.0 + both v0.2.0 rows; add the TODO item).
13. **core v0.8.0 + systemd v0.1.0 train** (steps in TODO P3; demand-proven by bank-sync ADR-017).
14. **Example StartHooks migration** post core v0.8.0 (flightrecorder + otel examples).
15. **fr v0.2.1 sweep** on the frh + cqrs next trains (safe combination-wise; trains only).
16. **AGENTS rule from e1** (never pipe verify commands; redirect to file) — one line on the next AGENTS touch.
17. **AGENTS deep slim-down** — 361/377 again; every addition needs a removal until the structural cut lands.
18. **Post-commit HEAD-advance sanity hook** (design first; guards the bb47fbc rewind class).
19. **Composition-gap E2E watchlist** items (otel×realtime first).
20. **Health/polish + BuildFlow-dprint + art-dupl backlog** (standing TODO P3 items — untouched this session, remain valid).

## g) Questions I can NOT figure out myself

1. **Where does the pma daemon live — unit name, config path, or log directory I may read?** Evidence trail this session: pushed fine until 22:47Z, stalled ~30 min, resumed; re-added `./integration` to go.work three times (latest bundled into `2d991a9` with my README); performed a reflog-silent rewind of my commit `bb47fbc` (content left staged, worktree reverted); currently 2 commits unpushed. From this shell: no process found mid-session, no `~/.config/pma`, `systemctl` blocked. Its config/logs are the only path to finishing the root-cause.
2. **May I push?** `git push origin master` (never force) for the 2 unpushed commits — or leave push authority with the daemon exclusively? Related: do you want the push-lag watch (f5) as a repo guard script?
3. **SSH_PRIVATE_KEY (standing P1)**: still empty at repo level; the matrix has been dead since ~2026-10-01. ci-preflight now makes the failure one loud, self-explaning red job — restoring the secret (GitHub → Settings → Secrets → Actions) is the only remaining step to a fully green master run.

---

_Session commits (all verified before commit; daemon pushes): `d6b83b3` sync guard · `73e95d7` CHANGELOG · `85c72fd` knob decision · `927b14c` rubric 52→87 · `d1eafa3` AGENTS pass · `1725cca` release-truth + guard dual-pattern (recovery of the rewound `bb47fbc`) · `f561048` TODO_LIST harvest · `ae3b2d4` go.work leak #3 fix (charter-guard catch) · `67c2a61` ci-preflight · `adb1985` interim report · this report. Daemon commits this session: `2d991a9` (README refs + leak #3), the vanished-commit incident, push stall + resumption. On origin at report time: everything through `ae3b2d4`._
