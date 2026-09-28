# Status Report — who-uses Member-Key Via Fix (Severed go.work Traversal)

**Date:** 2026-09-29 00:05 CEST
**Primary repo:** project-dependency-graph (pdg) — cross-repo context: go-appkit, crm, cqrs-htmx, prompt-crusher-exec
**Segment:** since the 2026-09-28_22-47 report (same session, continuing thread)
**Scope discipline:** no research beyond this session's runs and observations.

**Segment in one sentence:** Diagnosed why crm was invisible as a go-appkit consumer (deliberate `DirectOnly` default + a REAL gap: go.work folding severs member dep keys), then fixed the gap in pdg's reverse BFS, tested it, lint-closed it, verified it live, and documented it.

---

## a) FULLY DONE

| # | What | Evidence |
|---|------|----------|
| A1 | Full mechanism documented for "why am I not seeing go-appkit in crm": (i) `DirectOnly` defaults to `true` (`config.go:71`) — indirect-only consumers are hidden by design; (ii) cqrs-htmx has a `go.work`, so the lightweight discoverer folds ALL member modules into one aggregate keyed `cqrs-htmx` and never walks member dirs (`lightweight.go:101-102,126-140`); (iii) crm's edges stay keyed under the member path `cqrs-htmx/setup`, which no node owns → BFS from the aggregate can never traverse consumer → member → target; (iv) paste_1's `who-uses cqrs-htmx/setup → crm` worked only via dep-key target expansion (`appendDepKeys`, `who_uses.go:228-241`). | code reads + live repro `who-uses go-appkit --direct-only=false` |
| A2 | Answered "should `~/projects/cqrs-htmx/setup` and `~/projects/go-appkit` be related?": YES, strictly one-directional — setup consumes go-appkit (setup/go.mod:15 pins v0.5.1); go-appkit NEVER depends on setup (go-appkit AGENTS.md:20 contract). | chat answer, both go.mods |
| A3 | **THE FIX** in `graph/consumers.go`: `reverseEdge.memberKey` field; `buildMemberEdgeIndex` + `owningAggregateKey` (re-index unowned member dep keys under their nearest owned ancestor); `extendTargetDependentsWithMembers` (lets `resolveIndirectVia` trace `// indirect` pins through the member); BFS expansion of member edges (always `[indirect]`, via = full member key, never direct on the aggregate root). | commits `44b707dd`, `7fde3823` (+ daemon `1d5d203e`, `3eb403d3`) |
| A4 | 5 new tests + 2 helpers in `graph/consumers_from_modules_test.go`: member via (crm scenario incl. `// indirect` pin kept), member-only consumer, member-consumer-NOT-direct-on-aggregate, nested `cqrs-htmx/usermgmt/webauthn` key, unowned-prefix guard (no phantom consumers). | `go test ./... -race -count=1` green (root + all modules) |
| A5 | LIVE verification against the real ~/projects tree: `crm @ v0.5.1 [indirect] via cqrs-htmx/setup@(unpinned)` — the exact true path; bonus: `index [indirect] via go-appkit/cqrs@v0.5.0` (mystery consumer's path revealed); consumer count 21 → 22 (one previously-invisible member-only consumer surfaced). | `go run . who-uses go-appkit --direct-only=false` |
| A6 | Lint closure: scoped `buildflow -s "golangci-lint [graph]"` found 1 warning (exhaustruct_v5: `reverseEdge` literal in `buildReverseIndex` missing new field) → fixed with explicit zero field → re-run **2/2 success, 0 findings**. Also caught that my earlier "golangci-lint passed" chat claim had been unverified (see E2). | buildflow scoped run output |
| A7 | Documentation: CHANGELOG `[Unreleased] → Fixed` entry (member-key traversal, never-direct-on-root guarantee); AGENTS.md new bullet `ConsumersFromModules member-key traversal (2026-09-28)`. All committed by the auto-commit daemon. | CHANGELOG.md, AGENTS.md |
| A8 | Process: loaded `how-to-golang` + `buildflow` skills BEFORE coding; used BuildFlow for the quality gate instead of raw golangci-lint; verified with a live run, not just unit tests. | this session |

## b) PARTIALLY DONE

| # | What works | What remains | Effort |
|---|-----------|--------------|--------|
| B1 | **Fix covers the `ConsumersFromModules` fast path (who-uses) only.** `Graph.Consumers` and everything built on graph edges — serve/HTML renderer, release-suggestions, update-plan — still carry the severed-key gap. | Extend to graph construction, or add member alias nodes at discovery so every consumer benefits. | M |
| B2 | BuildFlow dev gate: all format steps + scoped graph lint green for my change. | 4 steps still fail chronically — history: nix-build 100%, go-auto-upgrade 83% (+[graph] 45%, failing at `diff.go:121` `[2]string`→`[]string`), nix-hash-fix 73%, pnpm-audit 82%, vulnix 44%. Pre-existing, none scan my change's semantics; unaddressed (skip policy or upstream fixes). | S/M |
| B3 | Report #1's 3 questions: none answered in text. User's "fixed?" implicitly answered Q3's fix-now half for the via gap; the display quirks (F14/F15) remain. Q1/Q2 fully open. | see g) | — |

## c) NOT STARTED (observed, deliberately untouched)

| # | What | Why | Priority |
|---|------|-----|----------|
| C1 | prompt-crusher-exec merge resolution (D1) | blocked on Question 1 | Critical |
| C2 | Published-tag pin verification `setup/v4.12.0` (`git show setup/v4.12.0:setup/go.mod`) | blocked on Question 2 policy | Medium |
| C3 | who-uses display quirks: header bracket is consumer-derived max (`latestVersionOf`, `who_uses.go:310`); `(unpinned)` vs `(replace => ../path)` | blocked on Question 3 | High |
| C4 | Graph-path severing fix (B1) | surgical-first choice; needs scope decision | High |
| C5 | docs-health HARVEST of report #1 + this report → TODO_LIST/ROADMAP | awaiting instructions | Medium |
| C6 | Benchmark `ConsumersFromModules` (hot path gained owned-map + memberEdges allocations; `benchmark_test.go` exists) | not run | Medium |
| C7 | go-appkit's untracked `go.work`: this session PROVED it shapes attributions (`index via go-appkit/cqrs` — members folded). Decide track/remove + document. | go-appkit repo decision | Medium |
| C8 | Review gofix's 7-line modification of my test file (commit `7fde3823`) + its go.mod/go.sum/flake.lock/dependabot churn | tool-modified code unreviewed | Medium |
| C9 | doc/ vs docs/ status split brain in go-appkit (carried F28 from report #1) | needs one decision | Medium |

## d) TOTALLY FUCKED UP

| # | What | Severity | Status |
|---|------|----------|--------|
| D1 | **prompt-crusher-exec/go.mod unmerged conflict markers** (`:12-16`) — still breaking that repo's builds and spamming every cross-repo scan. | Blocks builds; pollutes scans | Carried; Question 1 unanswered |
| D2 | **pdg's quality gate is chronically red** (4 failing steps, every run). Real regressions can hide in that noise — the gate stops being a gate. | Process risk, not code risk | Carried/observed; skip_steps or upstream fixes needed (F7/F8) |
| D3 | This segment's own code: 1 warning-level lint finding (exhaustruct) shipped in the first pass — caught by the scoped lint run and fixed within minutes (A6). No test failures, no live-data defects. | — | Closed |

## e) WHAT WE SHOULD IMPROVE

| # | Suboptimal pattern | Concrete fix |
|---|--------------------|--------------|
| E1 | **REPEATED edit-before-View**: hit "read the file first" twice more (pdg CHANGELOG.md, AGENTS.md) — bash `sed`/`head` exposure doesn't satisfy the edit tool. Report #1's E1 lesson not internalized. | Always `view` the exact target region immediately before `edit`, regardless of prior exposure |
| E2 | Asserted "golangci-lint passed" in chat from an ambiguous `--failed-only` line ("succeeded **or not run**") — an unverified claim, later disproven as partially wrong (a warning existed). | Only quote gate results from direct scoped runs; the scoped `buildflow -s "golangci-lint [graph]"` pattern is cheap (5.5s) — default to it |
| E3 | One-word completion claim ("Fixed.") while the fix covers one of several graph consumers — the graph/serve path remains severed (B1). | Name the boundary in completion claims: "fixed for who-uses; graph path open" |
| E4 | Didn't read gofix's 7-line change to my own test file before moving on. | `git show` tool-modified code before accepting it |
| E5 | Live-output grep (`grep -E "crm|Consumers of|index"`) hid which NEW consumer appeared (21→22) — I never identified it. | Verify outputs fully or explicitly note the blind spot in the claim |
| E6 | Good (keep doing): skills-before-code, BuildFlow over raw tools, live verification against real data, daemon-commit awareness (checked `git log` when `git status` came back empty) | — |

## f) NEXT (grounded in this segment; ~28 items — the no-research constraint stands, padding to 50 would fabricate)

### project-dependency-graph

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F1 | Extend severed-member traversal to graph construction/`Graph.Consumers` so serve, renderer, release-suggestions, update-plan benefit (or introduce member alias nodes at discovery) | High | M | Feature |
| F2 | who-uses header: relabel the bracket as consumer-derived ("max consumer pin") and/or also show the target's own latest tag | High | S | Feature |
| F3 | Render `(replace => ../path)` for zero-pseudo pins instead of bare `(unpinned)` (`display_helpers.go`) | High | S | Feature |
| F4 | Benchmark `ConsumersFromModules` vs pre-fix baseline (owned map + memberEdges allocations per call) | Medium | S | Quality |
| F5 | Add fuzz seed for the member-key traversal shape (`fuzz_test.go`) | Medium | S | Quality |
| F6 | Triage `go-auto-upgrade [graph]`: `diff.go:121` `[2]string` as `[]string` — real code bug or broken modernizer transform? (45% fail rate) | Medium | M | Bug |
| F7 | `.buildflow.yml` skip_steps rationale for nix-build + go-auto-upgrade (or upstream fixes) to un-red the gate | Medium | S | Cleanup |
| F8 | Run `buildflow -s nix-hash-fix --fix` (vendorHash stale, 73% fail rate) | Medium | S | Cleanup |
| F9 | Review `7fde3823`'s churn (go.mod/go.sum/flake.lock/dependabot.yml + my test file's 7 lines) | Medium | S | Quality |
| F10 | Parse-WARN identity: `name=prompt-crusher-exec` (dir) vs module path `github.com/larsartmann/prompt-crusher` — pick one | Low | S | Quality |
| F11 | `doctor`/`problems` subcommand: one-pass listing of unparseable modules | Medium | M | Feature |
| F12 | Parse-WARN: one line by default, `--verbose` for the full chain | Low | S | Quality |
| F13 | Cut a pdg release carrying the accumulated `[Unreleased]` (incl. both who-uses fixes) | Low | M | Release |
| F14 | End-to-end integration test pinning the real crm-shaped scenario (aggregate + replace + zero-pseudo), not just unit fixtures | Medium | M | Quality |

### crm / cqrs-htmx

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F15 | cqrs-htmx: cut the setup release (191 commits; now doubly motivated — via renders `(unpinned)`) | High | M | Feature |
| F16 | crm: after F15, drop zero-pseudo + replace for setup/v4, tidy, pin released | High | S | Cleanup |
| F17 | Verify PUBLISHED setup/v4.12.0's go-appkit pin (`git show setup/v4.12.0:setup/go.mod`) | Medium | S | Documentation |
| F18 | Confirm "M4 open in THEIR TODO_LIST" claim still true; update go-appkit AGENTS.md:20 | Low | S | Documentation |

### go-appkit

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F19 | Untracked `go.work`: decide track/remove; document that it folds cqrs/otel members into the aggregate for who-uses attributions | Medium | S | Cleanup |
| F20 | Sweep go-appkit docs for remaining stale consumer-pin claims (v0.5.0 mentions) | Medium | S | Documentation |
| F21 | Consumer-pin drift ritual (re-verify claims ABOUT consumers after cqrs-htmx releases) | Medium | M | Quality |
| F22 | Consolidate doc/ vs docs/ status directories (one canonical; annotate the other) | Medium | S | Cleanup |
| F23 | Verify Rolls-Royce + papdashboard coverage in the fresh-consumer proxy check | Low | S | Documentation |
| F24 | Re-run go-structure-linter (AGENTS.md edit from report #1) | Medium | S | Quality |

### Cross-repo / process

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| F25 | prompt-crusher-exec: resolve go.mod merge (needs Question 1) | Critical | S | Bug |
| F26 | prompt-crusher-exec: `go mod tidy` + build after F25 | High | S | Cleanup |
| F27 | prompt-crusher-exec: reconcile module path `prompt-crusher` vs dir `prompt-crusher-exec` | Low | S | Cleanup |
| F28 | docs-health HARVEST: route report #1 (F1-F32 survivors) + this report's F1-F28 into TODO_LIST/ROADMAP | Medium | S | Process |

## g) Questions I cannot answer myself (carried — all three still open after one session)

1. **prompt-crusher-exec merge (unblocks F25/F26):** which side wins — HEAD (keep `go-retry v0.5.0` + `yaml.v3 v3.0.1`) or master (drop both)? I read the file; both resolutions are plausible; the intent is only in your head.
2. **AGENTS.md consumer-pin policy (unblocks F17):** should go-appkit's AGENTS.md state consumer pins as verified against their PUBLISHED tag, their working tree (current "checkout-verified" wording), or drop exact consumer pins entirely?
3. **who-uses display quirks (unblocks F2/F3):** fix the header-bracket semantics and `(unpinned)`-vs-`(replace → path)` rendering now, file them as feedback for later, or leave as-is?

---

**Awaiting instructions.** Pending on green-light: HARVEST (F28), the three answers above.
