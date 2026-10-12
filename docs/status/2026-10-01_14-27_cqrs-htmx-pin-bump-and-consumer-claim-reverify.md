# Status Report — go-appkit — cqrs-htmx pin bump + consumer-claim re-verify

> **Date:** 2026-10-01 14:27 Thursday CEST
> **Scope:** This session only (resumed go-daemon/go-aichat evaluation thread → verification → an accidental-but-real guard fix). Point-in-time snapshot; goes stale.
> **Format note:** The `status-report` skill's canonical output is a styled HTML dashboard. This report is **Markdown per explicit user instruction** ("write a full status report at `docs/status/<...>.md`"). The override is one-off; it is not propagated back into the skill.
> **Evidence discipline:** every "done" item below cites a commit hash, a guard exit code, or a passing test command from this session.

---

## 0. Session in one paragraph

The session resumed a task the prior session had marked complete: the **go-daemon / go-aichat value-add evaluation**. I re-verified the in-tree deliverables (the research doc and the already-implemented `systemd` module + core `StartHooks` seam) rather than trusting the summary, ran the repo's guard scripts, and **found the `pin-drift` guard RED**: the `integration` module pinned `cqrs-htmx/v4 v4.12.0` while `v4.13.0` was published. I fixed it (pin + fixture + `go mod tidy`), then — because the fix required re-reading consumer claims — **re-verified the claims ABOUT cqrs-htmx `setup` against the newly published `setup@v4.13.2`** and corrected genuinely stale documentation. All repo-owned gates are green; the only red things left are pre-existing and out of scope for this session.

---

## a) FULLY DONE (verifiable, this session)

| #   | Done                                                                                                                                                                                                                                                                                                                                                    | Evidence                                                                                       |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| a1  | **Original evaluation deliverable verified present & coherent** — `doc/planning/2026-10-01_go-daemon-and-go-aichat-integration.md` (148 lines, verdict table + §1-§7), plus the follow-on it spawned: the `/systemd` module (`systemd/{doc,notify,hooks,example}`, tests) and core `ServiceConfig.StartHooks` post-listen seam.                         | `ls systemd/`, `view doc/planning/...`, `service.go` `Start()` runs `runStartHooks` after bind |
| a2  | **Core module race-green** — `go build ./...` + `go test ./... -race -count=1` passed (`ok ... 6.039s`, `ok .../testkit 6.036s`).                                                                                                                                                                                                                       | command output, `GOWORK=off GOTOOLCHAIN=go1.27.1`                                              |
| a3  | **Systemd module race-green** with its dev-only `replace => ../`.                                                                                                                                                                                                                                                                                       | `ok github.com/larsartmann/go-appkit/systemd 2.760s`                                           |
| a4  | **Integration pin drift FIXED** — `cqrs-htmx/v4` v4.12.0 → **v4.13.0** in `integration/go.mod` + the `documentedPins` fixture (`pin_drift_test.go:37`). Proxy-verified latest (`go list -m -json ...@latest` = v4.13.0, 2026-10-01T05:48Z); delta is additive (`setup/` seams; `transport/` package integration imports untouched).                     | commit **374183e**; `check-pin-drift.sh` exit **0** (was `FAIL`)                               |
| a5  | **`go mod tidy` cleanup of integration** — pruned now-unused indirects (`onsi/ginkgo`, `onsi/gomega`, `go-playground/form/v4`, `httputil/server_timing`) and the stale cqrs-htmx v4.12.0 go.sum lines. Repo convention confirmed tidy (root `go mod tidy -diff` empty).                                                                                 | commit **048137e** (`integration/go.mod -4`, `go.sum -13`)                                     |
| a6  | **Integration module fully re-verified after the bump+tidy** — build, `go test -race`, `go vet`, and `golangci-lint run ./...` (**0 issues**).                                                                                                                                                                                                          | command outputs                                                                                |
| a7  | **Consumer-claim re-verification** (the documented "consumer-claim drift ritual") against the PUBLISHED `setup@v4.13.2`: core pin moved **v0.5.1 → v0.7.0**; the `go-etag` stub-replace is **GONE** (tag carries split `go-etag/entitytag+server v0.6.0`). Updated `AGENTS.md` (21, 305), `FEATURES.md` (221-230), `TODO_LIST.md` (header + watchlist). | commit **048137e**; `setup@v4.13.2.mod` read from module cache                                 |
| a8  | **Guard sweep green** — `check-pin-drift.sh` exit **0**, `check-go-directives.sh` exit **0** (12/12 modules at `go 1.27.1`).                                                                                                                                                                                                                            | command outputs                                                                                |
| a9  | **Structure linter green** — `go-structure-linter . --exclude root-package-files --exclude internal-directory --exclude examples-directory` → **0 findings**; AGENTS at **357/377** counted lines (under the cap).                                                                                                                                      | command output                                                                                 |
| a10 | **Workspace sanity green** — workspace `go build ./...` clean + workspace `go test ./integration/... -count=1` passed.                                                                                                                                                                                                                                  | command outputs                                                                                |

**Nothing moved to production / no tag shipped** — see (b) and (g).

---

## b) PARTIALLY DONE

| #  | Item                              | What works                                                                                                                     | What remains / blocker                                                                                                                                                                                                                                                                                                                                                                | Effort |
| -- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| b1 | **systemd release train**         | Core `StartHooks` + the `/systemd` module are implemented and race-green in-tree. Wired into `go.work`, CI matrix, dependabot. | Deliberately UNRELEASED. `systemd/go.mod` still carries a dev-only `replace github.com/larsartmann/go-appkit => ../`; the require is v0.7.0. Sequence: **tag core v0.8.0 → lift the replace + bump require → tag `systemd/v0.1.0` → integration pin + `documentedPins` + CI proxy-smoke slot**. Blocker: tagging/pushing is **harness-forbidden without explicit user approval** (g). | M      |
| b2 | **TODO_LIST header pass-log**     | Header now records the 2026-10-01 systemd execution AND this session's pin bump + claim re-verify.                             | The header edit is the one file left uncommitted at report time (daemon picks it up in seconds).                                                                                                                                                                                                                                                                                      | S      |
| b3 | **Evaluation doc re-entry paths** | §7 appendix lists 5 concrete re-entry triggers.                                                                                | Not yet exercised — they are triggers, not tasks (UDS, content-negotiation, ssehub convergence, reverse adoption). By design.                                                                                                                                                                                                                                                         | —      |

---

## c) NOT STARTED (planned / surfaced, no code)

| #  | Item                                                                                                                                                                              | Why not started                                                                                   | Still wanted?                                 |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| c1 | **Mechanical guard for claims ABOUT consumers** — nothing verifies AGENTS/FEATURES/ROADMAP consumer claims against the proxy; today's drift was caught only by manual re-reading. | Posture is USER-GATED (TODO §g-1 "keep exact pins vs drop them").                                 | Yes — recommend a `check-consumer-claims.sh`. |
| c2 | **SSE-hub fleet coherence decision** (appkit/realtime vs go-aichat/ssehub)                                                                                                        | Owner-level call spanning two repos; watch item recorded 2026-10-01.                              | Yes (watch; don't merge blindly).             |
| c3 | **go-aichat reverse-adoption proposal** (KeyHolderAI hosting on `appkit.Service`)                                                                                                 | Executes in THEIR repo; USER-GATED.                                                               | Recommended, user-gated.                      |
| c4 | **UDS serving in core** (`Network`/`SocketPath`) — go-daemon Option C                                                                                                             | Outside appkit's TCP-HTTP charter; no consumer demand.                                            | Demand-gated.                                 |
| c5 | **Content negotiation (JSON/CBOR)** — go-daemon Option D                                                                                                                          | Routed to battery W3-B5 (`httpx`) on `go-codec`; not built.                                       | Demand-gated.                                 |
| c6 | **Core TLS option**                                                                                                                                                               | Gated on the same upstream httputil listener-injection seam as the `httputil.Server` composition. | Gated.                                        |
| c7 | **Cross-repo CI verification of consumer repos** in the release ritual                                                                                                            | Ritual currently covers OUR modules only.                                                         | Open question.                                |

---

## d) TOTALLY FUCKED UP (radical honesty)

> None of the items below were **caused by this session**; they are pre-existing conditions I observed. Listed here because they are the most valuable part of the report.

| #  | What is broken                                                                                                                                                                                                                                                                                             | Severity                           | Root cause                                                                                                         | Mitigation                                                                                            |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------- |
| d1 | **Master CI is a dead signal.** The `SSH_PRIVATE_KEY` Actions secret resolves empty → every `test` matrix job dies at `webfactory/ssh-agent` since ≥2026-09-17 (0 green in the last 100 runs). My green local runs are currently the ONLY verification signal for this repo.                               | **Critical** (blocks all CI trust) | Repo secret empty; agents cannot read repo secrets.                                                                | Owner-only fix (restore the secret; gate the ssh-agent step off `pull_request`).                      |
| d2 | **BuildFlow findings gate is RED** — `buildflow --dry-run` completes its pipeline (172 success / 0 failed) but exits **69** on 36 pre-existing error findings (`jscpd` 14, `erraudit [otel]` 7 / `[root]` 5, and more). Means `buildflow` is not currently a pass/fail gate.                               | High                               | Pre-existing tool engagement (jscpd duplicate `.golangci.yml`; erraudit context-loss/panic/sentinel findings).     | Documented P2 program; do NOT blanket-`--fix` (would mutate unrelated files).                         |
| d3 | **`go.work` is UNTRACKED but a CI job reads it** (`go-directives`): on a fresh checkout the guard is vacuous or red for the wrong reason.                                                                                                                                                                  | High                               | Unresolved tracked-vs-CI-generated decision (TODO P2).                                                             | Commit it or CI-generate it; also assert the 11 module dirs.                                          |
| d4 | **Possible now-dead `exclude` block** in `integration/go.mod`: it excludes `go-etag` **root** v0.1.1/0.3.1/0.4.0/0.5.0 to defend against the old monolith's `server/` package. After the cqrs-htmx bump moved the graph onto the **split** `go-etag/entitytag+server v0.6.0`, those excludes may be inert. | Medium (latent confusion)          | I noticed this while reviewing the bump but **did not verify** whether the root module is still in the build list. | Verify with `go list -m go-etag`; remove if absent (AGENTS already did this for the go.work replace). |
| d5 | **`go.work.sum` churn** — modified at session start by someone else, then again by the workspace commands; the auto-commit daemon keeps folding single hash lines in.                                                                                                                                      | Low (noise)                        | Workspace resolution + daemon.                                                                                     | Leave; not ours to revert.                                                                            |

---

## e) WHAT WE SHOULD IMPROVE

1. ~~**Make the pin-drift guard impossible to ignore locally.** It was red until I ran it by hand. If d1 (dead CI) persists, move the guard into the pre-commit hook so a red pin never leaves a workstation. _Impact:_ prevents shipping a stale-pin release train. _Fix:_ append the guard block to the BuildFlow pre-commit hook (the hook already carries project-guards).~~ done — githooks carry the guards FATAL since 2026-10-07
2. ~~**Script the consumer-claim ritual** (`c1`). Manual re-reading caught drift twice (2026-09-28 and today). _Fix:_ a `scripts/check-consumer-claims.sh` that resolves `setup@latest`'s `go.mod` from the proxy and diffs the claimed pin in AGENTS/FEATURES/ROADMAP.~~ done — shipped as check-consumer-claims.sh — clone+ls-remote design (the proxy cannot resolve the nested setup module tag scheme; found 2026-10-12)
3. ~~**Clarify doc-format ownership.** I could not find a `dprint` step in `buildflow --dry-run`; the visible formatter was `prettier-format` and `markdown-lint` is _skipped in full mode_. AGENTS references "BuildFlow's dprint step". Either the step name differs or markdown formatting is unowned in this repo. _Fix:_ pin the step name in AGENTS or confirm prettier owns markdown.~~ done — dprint IS the md formatter; dprint.json excludes CHANGELOG; check green 2026-10-12
4. **Don't duplicate BuildFlow by hand.** I ran `golangci-lint run` directly on `integration` (harmless — 0 issues), but the skill says delegate. _Fix:_ habit — `buildflow -s golangci-lint [integration]` instead. **→ repo convention stays per-module CLI lint (AGENTS documents it; BuildFlow full mode deadlocks here)**
5. ~~**`--fix` discipline.** `buildflow --fix` / full mode would mutate unrelated files (modernize/lo rewrites) and is documented to deadlock on lychee/shellcheck here. Keep using `--dry-run` + targeted `-s` runs.~~ done — codified in the AGENTS BuildFlow gotcha block
6. ~~**Report-format divergence tracking.** The skill's canonical output is HTML; the user asked for `.md`. Keep the override visible (done) and don't let it leak into the skill default.~~ **NOT-DO — override stayed session-local; skill default unchanged.**

---

## f) Top next tasks (50) — sorted by impact

> Legend: **Impact** = Critical/High/Medium/Low · **Effort** = S/M/L · **Category** = Bug/Feature/Quality/Cleanup/Documentation. Items marked **[NEW]** were surfaced by this session; the rest are drawn from `TODO_LIST.md` / this session's observations. **This section is the HARVEST input** — most of it should land in `TODO_LIST.md`/`ROADMAP.md`, not stay entombed here.

| #  | Task                                                                                                                                            | Impact   | Effort   | Category                                         |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------- | ------------------------------------------------ |
| 1  | Restore the `SSH_PRIVATE_KEY` Actions secret and get one fully green master run                                                                 | Critical | S        | Bug                                              |
| 2  | Gate the CI ssh-agent step off `pull_request` (Dependabot can never read the secret)                                                            | High     | S        | Bug                                              |
| ~~3~~  | ~~[NEW] Verify & remove the possibly-dead `go-etag` root `exclude` block in `integration/go.mod`~~ done — exclude block is ALIVE — it spans the package-extraction boundary (AGENTS 2026-09-29 note) | ~~Medium~~ | ~~S~~ | ~~Cleanup~~ |
| ~~4~~  | ~~[NEW] Add `scripts/check-consumer-claims.sh` (proxy-verify consumer pin claims)~~ done — scripts/check-consumer-claims.sh shipped + wired into pre-tag-checks (ac30874, 2026-10-12) | ~~High~~ | ~~M~~ | ~~Quality~~ |
| ~~5~~  | ~~Wire `check-pin-drift.sh` into the pre-commit hook / daemon~~ done — .githooks guards FATAL since 2026-10-07; pre-tag-checks runs pin-drift | ~~High~~ | ~~S~~ | ~~Quality~~ |
| ~~6~~  | ~~Decide go.work tracked-vs-CI-generated; assert the 11 module dirs~~ done — check-workspace-charter.sh asserts exactly the module dirs | ~~High~~ | ~~S~~ | ~~Cleanup~~ |
| ~~7~~  | ~~Execute the systemd release train (core v0.8.0 → systemd v0.1.0) — user-gated~~ done — core v0.8.0 + systemd v0.1.0 tagged 2026-10-07 | ~~High~~ | ~~M~~ | ~~Feature~~ |
| ~~8~~  | ~~Release the otel v0.2.0 train (SigNoz support is code-complete)~~ done — otel v0.2.0 shipped 2026-10-06 | ~~High~~ | ~~M~~ | ~~Feature~~ |
| ~~9~~  | ~~Structural fix for the go-directive re-drift CLASS (BuildFlow pre-commit wiring + semver-aware guard + `toolchain` directive)~~ done — go-directives guard + FATAL pre-commit + CI job landed 2026-10-07/12 | ~~High~~ | ~~M~~ | ~~Quality~~ |
| ~~10~~ | ~~Add cqrs wrapper lifecycle E2E through a live appkit Service (the highest-value uncovered module)~~ done — integration cqrs_lifecycle_test.go through a live Service (pinned 2026-09-29) | ~~High~~ | ~~M~~ | ~~Quality~~ |
| 11 | Decide the BuildFlow findings-gate posture for appkit (drive jscpd/erraudit to 0 vs documented program)                                         | High     | L        | Quality                                          |
| ~~12~~ | ~~Run `buildflow -s jscpd --format finding` and triage the 14 duplicate `.golangci.yml` findings~~ done — .jscpd.json ignores **/.golangci.yml (deliberate standalone configs, 2026-10-12) | ~~Medium~~ | ~~M~~ | ~~Cleanup~~ |
| 13 | Run `buildflow -s erraudit [root                                                                                                                | otel     | realtime | ...]` and triage the pre-existing error findings |
| ~~14~~ | ~~Consumer-claim posture decision: keep exact consumer pins (guarded) or drop them from AGENTS~~ done — G3 posture A decided and guarded (2026-10-12) | ~~Medium~~ | ~~S~~ | ~~Documentation~~ |
| 15 | security: browser CSP pass over the health dashboard under strict-CSP + nonce                                                                   | Medium   | M        | Quality                                          |
| 16 | errorpages: replace hand-rolled `statusRecorder` with `httputil.ResponseRecorder` (user gate)                                                   | Medium   | S        | Cleanup                                          |
| 17 | health: file the upstream `WithHealthRecorder` sentinel-error ask (user-gated)                                                                  | Medium   | S        | Documentation                                    |
| 18 | Re-run `otel` benchstat after the 1.27.1 unification (toolchain jump invalidates the old baseline)                                              | Medium   | S        | Quality                                          |
| 19 | Set up `govulncheck` on a networked machine for health + security                                                                               | Medium   | S        | Quality                                          |
| 20 | Advance the `httputil.Server` composition spike when upstream ships listener injection                                                          | Medium   | L        | Feature                                          |
| 21 | Battery W3 `httpx`: B1 ResultHandler family sharing errorpages' taxonomy (classification-parity test)                                           | Medium   | L        | Feature                                          |
| 22 | Battery W3 `httpx`: B5 content negotiation on `go-codec` (go-daemon `negotiation.go` as reference)                                              | Medium   | M        | Feature                                          |
| 23 | Battery W5 realtime: C1 drop/backpressure counters (borrow ssehub's `Drops()` shape)                                                            | Medium   | M        | Feature                                          |
| 24 | Battery W5 realtime: C2 projection→broadcast folded contract (the must-have)                                                                    | Medium   | M        | Feature                                          |
| ~~25~~ | ~~Health/polish backlog batch: hardened-dashboard `frame-ancestors 'none'` assert; named `Hook` type; typed `seriesKey`; `errors.Join` shape test~~ done — rode core v0.7.0 (appkit.Hook, typed seriesKey) + hardened-dashboard CSP test | ~~Low~~ | ~~M~~ | ~~Cleanup~~ |
| ~~26~~ | ~~testkit: add a drain-window assertion helper (drain tests hand-roll it)~~ done — testkit.DrainWindowProbe shipped v0.7.0 | ~~Low~~ | ~~S~~ | ~~Quality~~ |
| ~~27~~ | ~~AGENTS deep slim-down (extract per-module Gotchas to module READMEs; AGENTS 357/377)~~ done — AGENTS 126/220 after the 2026-10-12 slim-down (module-details.md) | ~~Medium~~ | ~~L~~ | ~~Documentation~~ |
| ~~28~~ | ~~cqrs README cookbook re-verification after the next go-cqrs-lite release~~ done — re-verified vs scenario/v4 v4.4.3 + testutil v4.3.4 on the cqrs v0.8.0 train (2026-10-12) | ~~Low~~ | ~~S~~ | ~~Documentation~~ |
| 29 | Watchlist refresh: cordis consumers; PapDashboard version; nixpkgs > 1.26.7; dprint exit-14 upstream                                            | Low      | S        | Cleanup                                          |
| 30 | art-dupl: one audit pass over the ~431 suppressed groups + record the corrected invocation                                                      | Low      | M        | Quality                                          |
| 31 | Full cqrs-htmx setup suite hermetic run (extend 3/3 verified to suite-green)                                                                    | Low      | M        | Quality                                          |
| 32 | Decide BuildFlow dprint exit-14 on CHANGELOG-only commits (escape hatch exists)                                                                 | Low      | S        | Cleanup                                          |
| ~~33~~ | ~~feat: named `Hook` type is (a6/25) — already landed as `appkit.Hook` alias; verify docs reference it~~ done — appkit.Hook landed v0.7.0; referenced across AGENTS/README | ~~Low~~ | ~~S~~ | ~~Documentation~~ |
| ~~34~~ | ~~[NEW] Record the `systemd` module in the module CHANGELOG `[Unreleased]` before tagging~~ done — systemd CHANGELOG dated at the v0.1.0 tag (2026-10-07) | ~~Medium~~ | ~~S~~ | ~~Documentation~~ |
| ~~35~~ | ~~[NEW] Add a `systemd` release smoke test to the fresh-consumer proxy check recipe~~ done — proxy recipe is module-generic; ran green for cqrs v0.8.0 + frh v0.1.7 (2026-10-12) | ~~Medium~~ | ~~S~~ | ~~Quality~~ |
| 36 | [NEW] Confirm `setup@v4.13.2`'s `RunWithAppkit` still passes their equivalence suite after the core v0.7.0 pin                                  | Medium   | M        | Quality                                          |
| ~~37~~ | ~~[NEW] Re-check whether `integration` still needs `samber/do`, `go-humanize`, etc. after the bump (tidy already pruned)~~ done — tidy no-op verified repeatedly since (floor train 13 files one-line each) | ~~Low~~ | ~~S~~ | ~~Cleanup~~ |
| 38 | [NEW] Document the `go-aichat` reverse-adoption proposal in the watchlist with concrete trigger                                                 | Low      | S        | Documentation                                    |
| 39 | [NEW] Add a CI job that fails when `check-consumer-claims.sh` detects drift **→ scoped to the pre-tag gate by design — the content leg needs the sibling cqrs-htmx clone (CI cannot run it)** | Medium   | M        | Quality                                          |
| 40 | papdashboard: re-check v0.3.0 family-dep parity + TLS demand on the next look                                                                   | Low      | S        | Cleanup                                          |
| 41 | Add core TLS support (gated on the same upstream seam)                                                                                          | Medium   | L        | Feature                                          |
| 42 | Add UDS serving (`Network`/`SocketPath`) when a consumer demands it                                                                             | Low      | M        | Feature                                          |
| 43 | Battery W4 `polite` outbound resilience client (go-aichat `client.Resilient` as reference)                                                      | Low      | L        | Feature                                          |
| ~~44~~ | ~~codify "verify the delivering layer, not just the commit" for module bumps~~ done — codified — fresh-consumer proxy check is Release Ritual standard (PASS x4 today) | ~~Low~~ | ~~S~~ | ~~Quality~~ |
| ~~45~~ | ~~Add a `systemd` module README build & verify block verification~~ done — systemd README Build & verify block verified in the 2026-10-12 prefix sweep | ~~Low~~ | ~~S~~ | ~~Documentation~~ |
| ~~46~~ | ~~Reconcile `AGENTS.md` ↔ module READMEs for the `StartHooks` seam (core + systemd)~~ done — examples migrated to StartHooks + READMEs swept (2026-10-12) | ~~Low~~ | ~~S~~ | ~~Documentation~~ |
| ~~47~~ | ~~[NEW] Confirm `buildflow -s markdown-lint` surfaces nothing on my doc edits~~ done — dprint check green over the repo 2026-10-12 | ~~Low~~ | ~~S~~ | ~~Quality~~ |
| ~~48~~ | ~~[NEW] Consider a `docs/status` index update (`docs/status/README.md`) for this report~~ done — docs/status/README.md index exists; this T18 pass updates it | ~~Low~~ | ~~S~~ | ~~Documentation~~ |
| 49 | Triage the `buildflow` "9 tools unavailable (health check failed)" set                                                                          | Low      | S        | Cleanup                                          |
| ~~50~~ | ~~Decide whether the release ritual should verify external consumer integrations~~ done — Release Ritual step 5 now runs check-consumer-claims.sh | ~~Low~~ | ~~S~~ | ~~Quality~~ |

> **HARVEST reminder:** items 4, 5, 6, 11-13, 14, 39 are the durable, non-duplicate additions; 1-3 mirror existing P1/P2 entries. This table should be harvested into `TODO_LIST.md` after the report (not left here).

---

## g) Top questions I cannot answer myself (3)

1. **Release execution authority.** The systemd train is code-complete but blocked on tagging/pushing, which the harness forbids without your explicit say-so. Do you want me to execute it (tag **core v0.8.0**, lift the `replace`, tag **systemd v0.1.0**, update integration pins, push), or hold? _(I cannot infer this: it is an irreversible, remote-visible action.)_ **→ answered by events — the 2026-10-07 train shipped core v0.8.0 + systemd v0.1.0; SUPERB v6 full-execution authorizes gated trains**
2. **Consumer-claim posture (the standing §g-1 USER-GATE).** Keep citing exact consumer pins in AGENTS/FEATURES/ROADMAP (and let me add a guard `check-consumer-claims.sh` so they stay honest), or drop the exact pins and describe consumers qualitatively? _(I drafted both paths; the choice is policy, not fact.)_ **→ answered 2026-10-12 — G3 posture A: pins kept, guarded by scripts/check-consumer-claims.sh**
3. **BuildFlow findings-gate posture.** Should appkit drive the pre-existing 36 error findings (jscpd/erraudit) to zero as project work, or keep them as a documented P2/upstream program and treat `buildflow` exit-69 as expected? _(I cannot decide: it is a scope/ownership call, and a blanket `--fix` is unsafe.)_ **→ answered 2026-10-12 — the 15:53 session drove the gate 69 to 0 with skips documented**

---

## Verification commands run (for reproducibility)

```bash
# core + systemd
cd <module> && GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
# integration after bump + tidy
cd integration && GOWORK=off GOTOOLCHAIN=go1.27.1 go build ./... && \
  GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1 && \
  GOWORK=off GOTOOLCHAIN=go1.27.1 go vet ./... && \
  GOWORK=off GOTOOLCHAIN=go1.27.1 golangci-lint run ./...
# guards + structure
./scripts/check-pin-drift.sh && ./scripts/check-go-directives.sh
go-structure-linter . --exclude root-package-files --exclude internal-directory --exclude examples-directory
```

**End state:** repo-owned gates green; `TODO_LIST.md` header edit pending the auto-commit daemon.
