# Status Report — Release-Train Research, Dependency Absorption, and Verify (Interrupted by Parallel Session)

**Date:** 2026-11-02 12:25 CET
**Scope of this report:** the single session that answered "Time for some new releases?" — loaded the go-release skill, researched the full pending-release surface, decided the 8-module wave shape, absorbed the pending dependabot bumps in-tree, and ran hermetic verification until the box's environment (load 245, full /tmp, shared-cache races) and a PARALLEL SESSION working the same tree interrupted the run. No tags were cut, nothing was pushed, no CHANGELOG was edited.

**Scope discipline — what this session VERIFIED (command/artifact):**

- Per-module pending-change inventory: `git diff <tag>..HEAD -- <module>/` for all 11 tagged modules (command log in session; results in §a-1).
- Latest family versions: `GOWORK=off go list -m -versions` against proxy.golang.org for error-family (v0.12.0), httputil (v1.7.0), go-health-dashboard (v0.12.2), ssetest (v0.5.0), go-cqrs-lite system v4.16.0 / event v4.15.2 / otel-v4 v4.5.3 — all confirmed LATEST.
- Dependabot PR contents: `gh pr diff` read for #44, #46, #47, #48, #50, #28; titles for #37/#38/#42/#51. 11 PRs open at session start; list captured.
- Dep absorption: `go get <pin> && go mod tidy` hermetic per module — exit 0 on all 7 modules (go.mod/go.sum updated, verified by grep of the bumped lines).
- Builds/tests as listed in §a/§b with exact commands (`GOWORK=off GOTOOLCHAIN=go1.27.2`, `-race -count=1`).
- `gh release list` → EMPTY (tags-only release convention confirmed; no GitHub Release objects to create).
- Root/security go.mod read directly (root has NO go-cqrs-lite requires — PR50's cqrs lines are go.sum-only; security DOES carry a direct `go-appkit v0.9.0` require despite the CHANGELOG's "zero appkit dependencies" package claim — require is example-motivated, package-import status NOT verified).
- CSP truth-check: `grep -n "slices.Sort" security/csp.go` → line 122 present; the Unreleased comment change describes existing behavior accurately.

**What this session did NOT verify (and why):**

- **API-break diffs (Release Ritual step 1) for ANY module** — the `go doc -all` old-tag-vs-tree comparison never ran; the wave was still in the verify phase. This is the single most important un-run ritual step.
- cqrs/otel/realtime `-race` suites — runs launched but results LOST (see §d-2).
- golangci-lint on ANY module this session — not reached.
- The parallel session's commits (a467328 M3-M14, b9bf812 golangci-config converge) — NOT reviewed by this session; their diffs are unexamined here.
- security package's non-import of appkit (the "example-only dependency" CHANGELOG claim) — read the go.mod, did not grep the package imports.

**Session in one sentence:** the release wave was fully scoped and its dependency floor was assembled and mostly verified, but a saturated machine plus a parallel session (which merged dependabot PRs, executed SUPERB v10 M3-M14, and committed this session's go.mod bumps inside its own commit) stopped the train before the first CHANGELOG was cut.

---

## Headline numbers

| Metric                                      | Value                                                                                                                                          |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Modules scoped for the wave                 | 8 (core v0.9.1, cqrs v0.9.0, otel v0.3.1, flightrecorder v0.2.1, health v0.3.1, realtime v0.1.4, security v0.2.1, systemd v0.1.1)              |
| Modules skipped (no unreleased content)     | 3 (flightrecorderhealth, errorpages, docs)                                                                                                     |
| Dependency bumps absorbed in-tree           | 12 bumps across 7 modules, all tidied                                                                                                          |
| Suites green this session                   | root/core (6.2s), health (1.3s), flightrecorder package (1.4s), cqrs build+vet, otel build+vet                                                 |
| Suites UNRESOLVED (lost background jobs)    | cqrs -race, otel -race, realtime build/vet/-race, flightrecorder example link                                                                  |
| Environment incidents                       | 4 (templ-embed cache race; go-mod permission storm; go-build link race; /tmp 100% FULL)                                                        |
| Parallel-session commits landed mid-session | 8+ (PR merges #37/#38/#51, docs catalog v4.11.0, M3-M14 a467328, templ-components/errorpage f323363, golangci converge b9bf812, merge cbcfeba) |
| Tags cut / pushes                           | 0 / 0                                                                                                                                          |

---

## a) FULLY DONE

1. **Release-skill loading and ritual reconciliation.** go-release SKILL.md + its multi-module reference read in full; reconciled against the repo's own AGENTS.md Release Ritual (which overrides the generic skill on multi-tag waves, pre-tag-checks, and the insteadOf tidy trick).
2. **Full pending-release inventory, all 11 tagged modules.** Per-module `git diff <latest-tag>..HEAD` file lists + diffstats collected. Verdicts: real runtime source changes only in cqrs (lo.Flatten concat, Shutdown → `sys.GracefulClose` delegation, new scenario-store pin test) — everything else is deps, godoc/comments, READMEs, examples, or CI/scripts.
3. **Version decisions for the 8-module wave** (recorded here as the session's contract): core **v0.9.1** patch (httputil v1.7.0 align + error-family v0.12.0 + the CI/scripts infra already in-tree since v0.9.0); cqrs **v0.9.0** MINOR (go-cqrs-lite dep train + samber/lo runtime adoption + GracefulClose behavior delegation — precedent: cqrs v0.8.0 was a dep-sweep minor); otel v0.3.1, flightrecorder v0.2.1, realtime v0.1.4, security v0.2.1, systemd v0.1.1 (patch); health v0.3.1 (godoc-only, `docs:` tag prefix per the v0.5.1 convention). frh/errorpages/docs SKIPPED (no `[Unreleased]` content, no package source change).
4. **Latest-version calibration against the proxy.** All absorb targets confirmed LATEST at decision time; notably go-health-dashboard was already at v0.12.2 upstream — TWO minors ahead of PR #44's proposed v0.11.0.
5. **Dependabot PR surface mapped.** 11 open PRs; per-PR content extracted. Absorption plan: #39 (systemd error-family), #47 (flightrecorder error-family + httputil v1.7.0), #44 (health; superseded by taking v0.12.2), #46 (security httputil v1.7.0), #28 (realtime ssetest v0.5.0), #48 (cqrs event v4.15.2 / otel-v4 v4.5.3 / system v4.16.0 — all NEWER than in-tree), #50 (root indirects — root's cqrs lines proved go.sum-only). PRs #37/#38/#42/#51 were MERGED by the parallel session mid-run.
6. **Dependency absorption executed and tidied — 7 modules, 12 bumps, all exit 0:**
   - root: httputil v1.6.0 → **v1.7.0** (aligns core with otel; v1.7.0 = v1.6.0 content minus the replace leak, plus `NewServerListener`).
   - cqrs: event/v4 v4.15.1 → **v4.15.2**, otel/v4 v4.5.2 → **v4.5.3**, system/v4 v4.15.4 → **v4.16.0**.
   - flightrecorder: error-family v0.11.0 → **v0.12.0**, httputil v1.5.0 → **v1.7.0**.
   - health: error-family → **v0.12.0**, go-health-dashboard v0.10.2 → **v0.12.2** (which lifted templ-components v1.21.0 → **v1.24.0** + its four submodules — a bigger jump than PR #44, accepted deliberately with suite-green as the gate).
   - realtime: ssetest v0.4.0 → **v0.5.0** (the 10-16 PR, oldest open).
   - security: httputil v1.5.0 → **v1.7.0** (error-family indirect lifted to v0.12.0 by tidy).
   - systemd: error-family v0.11.0 → **v0.12.0**.
7. **Hermetic suites that completed, GREEN:** root/core build+vet+`-race` (ok 6.2s + testkit 6.5s); health build+vet+`-race` (ok 1.3s, dashboard v0.12.2 + templ v1.24.0 — the risky jump, proven); flightrecorder package `-race` ok 1.4s; cqrs build+vet exit 0; otel build+vet exit 0.
8. **Environment diagnosis, all four incidents root-caused, none repo-caused:** (i) templ-components v1.24.0 embed `lstat … no such file` = shared-cache extraction race (retry green, file present in the extracted dir); (ii) `mkdir /mnt/buildcache/go-mod: permission denied` storm mid-cqrs = concurrent process swapping the cache dir (timestamps 08:33/08:39; writable again on re-check); (iii) otel link `…/go-build/03/…-d: no such file` = the documented shared go-build race; (iv) **/tmp tmpfs 48G at 100%** — worked around with `TMPDIR=/mnt/buildcache/tmp-lars` (76G free) for all subsequent go invocations; /tmp NOT cleaned (mixed ownership incl. live BuildFlow dirs).
9. **Two discoverable-later traps closed for the record:** root go.mod carries NO go-cqrs-lite requires (PR #50's cqrs lines are go.sum-only — nothing to absorb there beyond indirects); `gh release list` is EMPTY — this repo's convention is annotated tags only, no GitHub Release objects, so skill Phase 7 is a no-op here.
10. **Scope discipline held:** the SUPERB v10 plan's M26 (testkit helper dedup, "rides the next natural core train") was consciously NOT pulled into core v0.9.1 — the plan itself ranks it rest-tier, and this train ships what exists.

## b) PARTIALLY DONE

1. **Module verification matrix — 4 of 8 wave modules fully green, 3 unresolved, 1 not started.** Green: core, health, flightrecorder (package), plus cqrs/otel build+vet. Unresolved (runs launched, results lost with the background shells after the tool interruption): cqrs `-race`, otel `-race`, realtime full suite, flightrecorder example link. Not started: security, systemd.
2. **The wave itself — dependency floor assembled, zero release artifacts produced.** No CHANGELOG section dated, no AGENTS/FEATURES/TODO release-state sync, no integration re-pin, no commit by this session, no tags, no pre-tag-checks, no push, no post-push proxy/fresh-consumer checks.
3. **Dependabot closure — bumps landed but PR bookkeeping unverified.** The absorbed bumps are >= every open PR's proposed version, so #28/#39/#44/#46/#47/#48/#50 should auto-close as superseded on dependabot's next run — NOT verified, and no manual closes were issued.
4. **This session's go.mod/go.sum work is in master's tree but NOT by this session's hand:** the parallel session's `a467328` ("Execute SUPERB v10 M3-M14") swept the 7 modules' bumped go.mod/go.sum into its commit (verified: `git log -- go.mod cqrs/go.mod health/go.mod` → a467328). Attribution inside that commit message does not mention the dependency absorption — the release CHANGELOGs must carry the true content list regardless.

## c) NOT STARTED

1. **Release Ritual step 1 — API-break `go doc -all` diff for all 8 modules.** Nothing tagged can happen before this runs.
2. **CHANGELOG cuts** (date `[Unreleased]` → `[version] - 2026-11-02`, reset placeholders) for all 8 modules — including WRITING the missing cqrs go-cqrs-lite dep-train entry (its `[Unreleased]` currently records only the lo adoption; the v4.15.x→v4.16.0 train and the GracefulClose delegation are undocumented).
3. **Same-train release-state updates:** AGENTS `_Release State_` + per-module bullets, FEATURES.md version-truth rows, TODO_LIST header, then `check-pin-drift.sh` + `check-go-directives.sh` green.
4. **Integration re-pin** to the 8 new tags via the documented `url.insteadOf` trick + hermetic integration suite + `documentedPins` fixture update.
5. **Lint pass** (`golangci-lint run ./...` per changed module, sequential, from module dirs) — zero modules linted this session.
6. **Commit, local annotated tags ×8, `./scripts/pre-tag-checks.sh <all 8 tags>`, push master + tags together.**
7. **Post-push ritual:** proxy propagation checks, fresh-consumer proxy check per `docs/recipes/fresh-consumer-proxy-check.md`, pkg.go.dev fetch triggers, session-close ritual (CI green on the FINAL pushed commit — this also closes the P1 push-row acceptance and SUPERB v10 M1/M2).
8. **The parallel session's dirty docs tree** (FEATURES.md, DOMAIN_LANGUAGE.md, SUPERB v10 plan, 18 status files) — uncommitted work from the OTHER session; untouched here, must land or be reconciled before any release-prep commit to avoid sweeping чужой work into the train.

## d) TOTALLY FUCKED UP

1. **Lost background-job results through sloppy orchestration.** Three verification jobs (cqrs `-race`, otel+flightrecorder-example, realtime suite) were launched and never collected before the session was interrupted; the shells are gone and their results are unrecoverable. Root cause: launching new jobs while old ones were pending, instead of collecting in order. Cost: three suites must re-run (~minutes), and the session's verify claim is weaker than the work deserves.
2. **Two early runs were launched without the TMPDIR guard and died on the /tmp-full class** (cqrs run 2, otel run 1) — the first /tmp-full failure (cqrs run 2) should have immediately gated EVERY subsequent go invocation behind `TMPDIR=/mnt/buildcache/tmp-lars`; instead one more run (otel retry) went out unprotected before the pattern was recognized.
3. **Parallel-session collision was not anticipated.** Two sessions working one tree: the other session's `a467328` committed this session's in-flight go.mod edits inside its own M3-M14 commit. No harm to content (the bumps are exactly as this session built them — verified by grep), but attribution is tangled and the release train now starts from a tree whose last commits this session has NOT reviewed (M3-M14 changed guards, CI venues, and 20+ docs). A release train should never ride commits it hasn't read; that review is now mandatory pre-work.
4. **Nothing else.** No data loss, no bad tags (none created), no pushed state, no tree corruption — the failed runs were environment classes, and every completed artifact (go.mod/go.sum bumps, green suites) is reproducible from the recorded commands.

## e) WHAT WE SHOULD IMPROVE

1. **Gate every go invocation on a preflight trio** when the box is loaded: `uptime` + `df -h /tmp` + a cache-writability probe — one helper would have prevented incidents (ii)–(iv) wholesale. `scripts/quiet-run.sh` already exists for load; it needs a /tmp-space and cache-writability gate added.
2. **Collect background jobs FIFO.** Never have more than one uncollected job; the interruption-proof pattern is one background suite at a time, result landed in a file under `/mnt/buildcache/tmp-lars/` before the next launch (also survives lost shells: the FILE survives).
3. **Session handshake for shared trees:** before mutating anything, record `git log -1 --format=%H` and re-check it before every write phase; if HEAD moved, STOP and reconcile (this session's dep edits got committed by someone else's commit — legal but fragile).
4. **Write the CHANGELOG entry at absorb time.** The 12 bumps have no CHANGELOG text yet; bump and document in the same motion so the train's truth lives where it ships.
5. **Absorb PRs by reading the PR first, proxy second** — worked here, but PR #44's v0.11.0 vs upstream v0.12.2 shows dependabot's proposed version can lag upstream by minors; always calibrate against `go list -m -versions` (this session did — keep it as the standing order).
6. **Run the API-break diff EARLY** (right after inventory, before any bump) — it is the ritual's step 1 and the only step that can change a version decision (patch↔minor); this session decided versions from diffstat reading alone.

## f) NEXT — ranked list (30 items)

1. Reconcile with the parallel session's uncommitted docs tree (land or park it) — nothing else proceeds on a mixed dirty tree.
2. Review `a467328` (M3-M14) + `b9bf812` (golangci converge) diffs — the release train must ride read commits; b9bf812 touched all 12 `.golangci.yml` files which CHANGES the lint contract mid-train.
3. Re-run the three lost suites with the TMPDIR guard, output to files: cqrs `-race`, otel `-race`, realtime build/vet/`-race`; flightrecorder example build.
4. Verify security (build/vet/`-race` + lint) and systemd (same) — the never-started two.
5. Lint all 8 changed modules sequentially from their own dirs (NEW configs from b9bf812 are now the contract).
6. Release Ritual step 1: `go doc -all` old-tag vs tree for all 8 modules; confirm the patch/minor decisions (esp. cqrs v0.9.0 minor).
7. Health decision: keep dashboard v0.12.2 (suite green) vs fall back to PR44's v0.11.0 — owner question (g-2).
8. Cut 8 CHANGELOG sections; write cqrs's missing go-cqrs-lite train + GracefulClose + scenario-pin entries; health keeps `docs:`-prefix tag message.
9. Sync AGENTS `_Release State_` + module bullets (new "Latest per module" line).
10. Sync FEATURES.md rows for all 8 modules.
11. Rewrite TODO_LIST header Updated line (release train) + close the P2 dependabot-wave row once PRs resolve.
12. Run `check-pin-drift.sh` + `check-go-directives.sh` (must be green BEFORE the release commit is final).
13. Update integration `documentedPins` fixture for the 8 new pins.
14. Integration re-pin: `git config --global url."…/go-appkit".insteadOf …` trick → tidy integration against LOCAL tags → hermetic integration suite (`GOWORK=off`, verbatim README commands) → unset the rewrite.
15. Decide and record whether sibling go.mods bump their `go-appkit` require to v0.9.1 this train (recommendation: NO — v0.9.1 is dep-only, MVS covers consumers; only integration's pin moves).
16. Commit the release prep (single train commit, message states the 8 tags + semantic deltas).
17. Create 8 local annotated tags with semantic-delta messages (`docs:` prefix for health).
18. `./scripts/pre-tag-checks.sh <all 8 tags>` — one gate run, all green.
19. Push `origin master` + all 8 tags in one push.
20. Post-push: proxy `.info` checks per module; fresh-consumer proxy check recipe; pkg.go.dev fetch triggers (remember flightrecorder's page has never indexed — re-trigger, cosmetic).
21. Session-close ritual: `gh run view` EVERY job on the FINAL pushed commit; green except designed ci-preflight red; close P1 push row + SUPERB v10 M1/M2.
22. Confirm dependabot auto-closed #28/#39/#44/#46/#47/#48/#50; manually close any survivor with a superseded-by comment; confirm #37/#38/#42/#51 merges' CI green.
23. Run `check-consumer-claims.sh` post-tag (setup/PapDashboard claims unaffected by v0.9.1, but the guard is the proof).
24. Verify the root README "Development" section + AGENTS dependency table (still says httputil v1.4.0 — STALE, caught this session, unfixed) — update to v1.7.0 truth in the train's doc sync.
25. Update AGENTS core-dependencies table row for go-error-family (v0.10.1 in the table — also stale vs v0.12.0).
26. /tmp hygiene: owner decision on clearing stale `go-build*` dirs (mixed ownership incl. live BuildFlow scratch) — or add a nightly sweeper; until then every session needs the TMPDIR guard (g-3).
27. Add /tmp-space + cache-writability gates to `scripts/quiet-run.sh` (improvement e-1 made mechanical).
28. Check go-health v0.8.1 is still health's newest (PR44 didn't move it; a newer may exist since 10-27) before health v0.3.1 is final.
29. Post-train: re-run `verify-all.sh` once on the tagged tree as the whole-fleet sweep proof.
30. Consider a tiny follow-up train note in the docs module CHANGELOG: the docs MODULE was skipped this wave but its directory's status-report corpus grew — decide if docs/v0.3.4 ever ships for corpus-only growth (probably never; record the ruling).

## g) Questions ONLY the owner can answer

1. **Parallel-session precedence:** another session executed SUPERB v10 M3–M14 and committed this session's dependency bumps inside `a467328`, leaving 20+ docs files dirty. Should I (a) continue the release train from this exact tree after reviewing their commits, or (b) wait until that session closes its docs work? (I recommend (a) — review, then train.)
2. **health v0.3.1 content:** keep the absorbed go-health-dashboard **v0.12.2** jump (two minors past PR #44's proposal, suite `-race` green, templ-components lifted to v1.24.0), or fall back to the PR-tested v0.11.0? (I recommend keeping v0.12.2 — latest, proven locally.)
3. **Machine hygiene:** /tmp (tmpfs, 48G) is 100% full with mixed-ownership `go-build*` scratch, and `/mnt/buildcache/go-mod` was transiently swapped by another process mid-run. May I clean stale `/tmp/go-build*` dirs owned by `lars` (risk: interrupting a live BuildFlow run), or will you sweep? Until then I run everything under `TMPDIR=/mnt/buildcache/tmp-lars`.

---

**Push/tag state at report time:** local master contains the parallel session's commits + this session's committed dep bumps; NOTHING pushed by this session; ZERO tags created; branch delta vs origin not re-checked after the parallel merges (origin was merged INTO local at cbcfeba).
