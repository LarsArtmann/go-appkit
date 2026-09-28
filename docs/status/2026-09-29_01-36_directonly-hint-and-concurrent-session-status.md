# Status Report — DirectOnly Hidden-Indirect Hint + Concurrent-Session Collision

**Date:** 2026-09-29 01:36 CEST
**Primary repo:** project-dependency-graph (pdg) — cross-repo observations: index, go-appkit, crm, cqrs-htmx, prompt-crusher-exec
**Segment:** since the 2026-09-29_00-05 report
**Scope discipline:** no research beyond this session's runs and observations.

**Segment in one sentence:** Diagnosed a suspected caching issue as real input drift (index became a true direct go-appkit consumer), explained the `DirectOnly=true` rationale, then — on your "do" — shipped the hidden-indirect hint so the adoption-only default view stops silently hiding blast radius, while surviving a live collision with a second concurrent session in the same repo.

---

## a) FULLY DONE

| # | What | Evidence |
|---|------|----------|
| A1 | **"Caching?" question resolved — no cache exists in the who-uses path.** Each run re-discovers from disk (pipeline logs show a fresh `discovering modules` pass, ~110-135ms). The output delta was real input drift: `index`'s go.mod was committed 00:07:35 (daemon `7c080a6`) promoting go-appkit to a direct require, AND index's code genuinely imports it (`internal/webapp/service.go`, `cmd/indexer-web/main.go`). So `index @ v0.5.1 [direct]` is a true positive, not comment drift. Module count 193→194 = disk change too. | rg on index/go.mod + `git log -- go.mod` + import scan |
| A2 | **`DirectOnly=true` rationale documented and explained** from the tool's own sources: config.go:71 help text ("indirect-only requires are not usage"); `who_uses_test.go:908-944` regression test ("presenting go.mod bookkeeping as adoption" inflated a consumer count from 1 to ~20); `graph/update_plan.go:15` (indirect consumers aren't actionable — they resolve when the direct dep bumps). Plus the legitimate counterpoint: for impact questions transitive consumers ARE affected, and the hidden view was useless until tonight's via fix. | chat answer + code cites |
| A3 | **THE HINT SHIPPED** (`who_uses.go`): `targetResult.hiddenIndirect` counted pre-filter (zeroed when `--direct-only=false`, so the hint never double-reports); `buildTargetTreeNode` appends `… and N hidden indirect consumers (--direct-only=false to show)` after the last consumer; a target with zero direct consumers now renders `No direct consumers for X, but N hidden indirect consumers (--direct-only=false to show).` instead of implying zero consumers. | code + commit via daemon |
| A4 | **4 new tests** in `who_uses_test.go`: hint with plural (2 hidden, names NOT leaked), singular (1), no-direct-consumers variant, and no-hint-when-indirect-shown. All pre-existing who-uses tests pass unmodified (hint is additive). | `go test . -run TestRunWhoUses` ok; full suite green |
| A5 | **LIVE verification** on the real tree: `who-uses go-appkit` → 4 direct consumers + `… and 18 hidden indirect consumers (--direct-only=false to show)` — exactly matching the `--direct-only=false` count (22 total). | `go run . who-uses go-appkit --dir ~/projects` |
| A6 | **Lint closure:** scoped findings (golines on my test line 966, gci on the other session's update_plan_logic_test.go) auto-fixed via `buildflow -s golangci-lint --fix` → **14/14 success, gate 7/7 passed**. | buildflow output |
| A7 | **Docs:** CHANGELOG `[Unreleased] → Added` entry; AGENTS.md `DirectOnly` bullet extended with the hint + no-direct variant. Daemon-committed by `03d787ff`. | CHANGELOG.md, AGENTS.md |
| A8 | **Concurrent-session collision survived without damage** (see D3/E-notes): another session committed broken test files mid-flight (`849adec1` at 00:37:27 — `display_helpers_test.go` used `types.CommitAnalysis` without importing `types`; `update_plan_logic_test.go` referenced a nonexistent `PlannedRelease.Consumers`). I repaired what was unambiguous (the import — raced and lost to the other writer by 4s, which fixed it itself), refused to implement the other session's half-landed feature, polled twice, and the tree self-healed at `174128e4`. Full suite green immediately after. | git log 849adec1→174128e4, build results |
| A9 | Re-confirmed the member-key via fix still holds under drift: `crm @ v0.5.1 [indirect] via cqrs-htmx/setup@(unpinned)` unchanged in the flag view. | live run |

## b) PARTIALLY DONE

| # | What works | What remains | Effort |
|---|-----------|--------------|--------|
| B1 | Hint covers the tree renderer. Machine-readable parity not checked — if who-uses has (or gains) `--format json|yaml`, `hiddenIndirect` should serialize there too. | Verify/extend output formats. | S |
| B2 | Report #2's B-items stand: severed-member fix covers only the `ConsumersFromModules` fast path (graph/serve/suggestions/update-plan still severed); BuildFlow gate chronic red (nix-build, go-auto-upgrade 2×, nix-hash-fix, pnpm-audit, vulnix in history) untouched. | as before | M |
| B3 | Your 3 questions from both prior reports remain unanswered; Q3's scope shrank again (hint done; header-bracket semantics F2 and `(unpinned)`-vs-`(replace → path)` F3 still open). | see g) | — |

## c) NOT STARTED

| # | What | Why | Priority |
|---|------|-----|----------|
| C1 | prompt-crusher-exec merge (D1) — the WARN appeared in EVERY run tonight including yours at 00:37 | blocked on Question 1 | Critical |
| C2 | Review the concurrent session's landed work: `849adec1` + `174128e4` (+94 lines release_suggestions_test.go, +39 update_plan_logic_test.go with a new `mergeFirstByKey` generic, graph_test.go changes) — I deliberately did not audit another session's in-flight feature | out of my scope; you may want it reviewed | Medium |
| C3 | Benchmark `ConsumersFromModules` (carried twice; hot path now allocates owned map + memberEdges + per-run hiddenIndirect counts) | not run | Medium |
| C4 | F2/F3 display quirks (blocked on Question 3) | — | High |
| C5 | HARVEST: three status reports of unharvested fuel now sit in docs/status/ | awaiting instructions | Medium |
| C6 | pdg AGENTS.md length governance — bullets keep growing; markdown-lint is skipped in dev mode so nothing checks it | unverified | Low |
| C7 | Carried: published-tag pin verification (Q2), setup re-tag, crm re-pin, doc/-vs-docs/, index added to go-appkit's consumer inventory (it is now a DIRECT consumer — A1), consumer-pin drift ritual, go.work decision | carried | Medium |

## d) TOTALLY FUCKED UP

| # | What | Severity | Status |
|---|------|----------|--------|
| D1 | **prompt-crusher-exec/go.mod conflict** — STILL unresolved after ~3.5 hours and ≥4 scans; the corruption WARN is now a permanent fixture of every command output. | Blocks builds; normalizing broken output | Carried; Question 1 unanswered |
| D2 | **Two sessions + a heuristic daemon committing to one repo concurrently produced a BROKEN COMMIT on master** (`849adec1`: test files that don't compile, landed 00:37:27). Master's `go build ./...` was red for ~19 minutes (00:37→00:56). Anyone pulling in that window got a broken tree. Root cause: heuristic auto-commit + concurrent sessions + no pre-commit compile gate on test-only changes. | Master build broken (window) | Self-healed by the other session at `174128e4`; the PROCESS hole remains (see F29) |
| D3 | This segment's own work: no defects shipped. The golines finding in my test was auto-fixed pre-commit; the hint logic verified against live data twice (22 = 4+18 invariant). | — | Clean |

## e) WHAT WE SHOULD IMPROVE

| # | Suboptimal pattern | Concrete fix |
|---|--------------------|--------------|
| E1 | I nearly edited a file the concurrent session was actively rewriting (`display_helpers_test.go`, 4-second mod-time race). Blind edits under concurrency can clobber in-flight work. | Before editing in pdg while other sessions may be live: `git log -1 --format=%ci -- <file>` + `git status` immediately prior; treat a dirty-but-unexpected state as a stop signal |
| E2 | I waited out the breakage with blind 45s/60s sleeps instead of reading the other session's intent from its commit contents immediately (`git show 849adec1` gave me the answer in one command the FIRST time). | Diagnose from diffs first; poll only after the diagnosis says "in-flight, wait" |
| E3 | `buildflow -s golangci-lint --fix` reformatted the OTHER session's file (gci import order in update_plan_logic_test.go) while it was possibly still editing. Formatting-only and verified green, but still a cross-session write. | Prefer scoped `--fix` on OWN files when a concurrent session is active; batch the rest |
| E4 | Repeated from report #2: golines auto-modified my new test line and I did not re-read the result. Third occurrence of "tool-modified my code, unreviewed". | `git show` after any `--fix` that touches my files — 10 seconds |
| E5 | The `DirectOnly` debate itself: the adoption-vs-impact tension was resolved by a flag default + a hidden view. Now that the via paths exist, the default view could arguably group consumers (direct, then indirect grouped via their path) instead of hiding. Design conversation worth 10 minutes before more rendering code lands on top. | Decide the target UX once (Question 3), then implement |
| E6 | Good (keep): live-verification invariant (22 = 4 direct + 18 hidden checked both ways), refusing to implement another session's half-landed feature, docs written same-segment | — |

## f) NEXT (grounded; ~30 — same no-research constraint)

### project-dependency-graph

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F1 | Check who-uses machine-readable output; serialize `hiddenIndirect` if formats exist (B1) | Medium | S | Feature |
| F2 | who-uses header bracket: relabel as consumer-derived max / show target's own tag (carried F2) | High | S | Feature |
| F3 | `(unpinned)` → `(replace => ../path)` rendering (carried F3) | High | S | Feature |
| F4 | Extend severed-member traversal to Graph.Consumers/serve/suggestions/update-plan (carried) | High | M | Feature |
| F5 | Benchmark ConsumersFromModules (carried twice) | Medium | S | Quality |
| F6 | Fuzz seed for member-key traversal (carried) | Medium | S | Quality |
| F7 | Triage go-auto-upgrade [graph] `diff.go:121` (carried) | Medium | M | Bug |
| F8 | `.buildflow.yml` skip_steps rationale for chronically red steps (carried) | Medium | S | Cleanup |
| F9 | `buildflow -s nix-hash-fix --fix` (carried) | Medium | S | Cleanup |
| F10 | F22 parse-WARN identity dir-vs-module (carried) | Low | S | Quality |
| F11 | doctor/problems subcommand (carried) | Medium | M | Feature |
| F12 | Parse-WARN one-line default + --verbose (carried) | Low | S | Quality |
| F13 | Audit the concurrent session's landed work (`849adec1`, `174128e4`: mergeFirstByKey + release-suggestions tests) for correctness and intent | Medium | M | Quality |
| F14 | Document a multi-session protocol in pdg AGENTS.md: before committing broken test files, the daemon (or sessions) should gate on `go build ./...`; sessions should announce scope files | High | S | Process |
| F15 | Consider a pre-commit (or daemon) `go vet ./...` gate so master can't go red on test-only changes like `849adec1` | High | M | Process |
| F16 | Cut a pdg release with the accumulated [Unreleased] (who-uses via fix, hint, other session's work once reviewed) | Low | M | Release |
| F17 | pdg AGENTS.md length governance: run markdown-lint scoped (`buildflow -s markdown-lint`) on AGENTS/CHANGELOG edits | Low | S | Quality |

### index / consumer fleet

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F18 | Add index to go-appkit's consumer inventory (it is now a real DIRECT consumer as of 00:07) + note in AGENTS.md consumer claims | Medium | S | Documentation |
| F19 | cqrs-htmx: cut setup release (carried; via still renders `(unpinned)`) | High | M | Feature |
| F20 | crm: re-pin released setup after F19 (carried) | High | S | Cleanup |
| F21 | Verify published setup/v4.12.0's appkit pin (carried; Question 2) | Medium | S | Documentation |

### go-appkit

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F22 | go-appkit untracked go.work: decide + document attribution effect (carried) | Medium | S | Cleanup |
| F23 | Sweep docs for stale consumer-pin claims (carried) | Medium | S | Documentation |
| F24 | Consumer-pin drift ritual (carried) | Medium | M | Quality |
| F25 | doc/ vs docs/ consolidation (carried) | Medium | S | Cleanup |
| F26 | Rolls-Royce + papdashboard fresh-consumer coverage (carried) | Low | S | Documentation |
| F27 | go-structure-linter re-run (carried) | Medium | S | Quality |

### Cross-repo / process

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F28 | prompt-crusher-exec merge + tidy + module-path reconcile (carried; Question 1) | Critical | S | Bug |
| F29 | Root-cause the broken-master window: why did the daemon commit non-compiling test files (`849adec1`)? Check whether a build gate exists and why it didn't run on test-only diffs | High | M | Bug |
| F30 | HARVEST all three status reports into TODO_LIST/ROADMAP | Medium | S | Process |

## g) Questions I cannot answer myself (carried — third time, still the real blockers)

1. **prompt-crusher-exec merge (unblocks F28, kills the eternal WARN):** HEAD side (keep `go-retry v0.5.0` + `yaml.v3 v3.0.1`) or master side (drop both)?
2. **Consumer-pin policy (unblocks F21/F24):** AGENTS.md consumer claims verified against PUBLISHED tags, working trees ("checkout-verified"), or drop exact pins?
3. **who-uses target UX (unblocks F2/F3, and E5's design conversation):** fix header-bracket semantics + `(replace → path)` rendering now, file for later, or leave? And do you want the default view to eventually GROUP indirect consumers by via-path instead of hiding them behind the hint?

---

**Awaiting instructions.** Pending on green-light: HARVEST (F30), the three answers above.
