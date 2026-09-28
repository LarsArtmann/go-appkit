# Status Report — Docs-Health Full Pass (AUDIT + HARVEST + ANNOTATE + ARCHIVE + Living-Doc Truth)

**Date:** 2026-09-29 00:38 CEST
**Repo:** go-appkit
**Scope discipline:** this session = the docs-health pass over all 66 `**/2026-0*` files + the six living docs. No other research. Everything below was directly observed, run, or fetched this session.
**Session in one sentence:** full docs-health AUDIT executed with real fixes — every living-doc claim re-derived from the tree, four consumer/contract drifts corrected in code and guards, 380 inline verdicts across 6 status reports, 10 files archived, the doc/-vs-docs/ split brain resolved — and one concurrent session's reports discovered and absorbed mid-pass.

---

## Headline numbers

| Metric | Value |
| --- | --- |
| `**/2026-0*` files inventoried | 66 |
| Living docs rebuilt/updated | 6 (README verified current — untouched by evidence) |
| Stale consumer/contract claims corrected | 4 (setup pin, M4-shipped, toolchain-laggard count, fold-in state) |
| Code/tests shipped this session | hook-code pin test (2 tests), frh `slices.Sort`, security-example PORT origin, integration pin bumps ×3, guard extension |
| Inline annotations | **380 `~~` markers** across 6 reports (scripts, dry-run first) |
| Files archived | 6 status reports + 1 HTML audit → `docs/status/archived/` (now 42); 3 plans → `doc/planning/archived/` (verdict banners) |
| Split brains resolved | doc/ vs docs/ status tree (`docs/status/` canonical; `doc/status/` migrated via `git mv` and removed) |
| Gates | pin-drift ✅ · dependabot-parity ✅ · structure-linter ✅ (AGENTS 376/377) · suites `-race` green on all 7 touched modules ✅ · go-directives ❌ expected (7 laggards, USER-gated train) |

---

## a) FULLY DONE (each verified this session)

| # | What | Evidence |
|---|------|----------|
| A1 | **Consumer-pin drift corrected with published-tag evidence:** `git -C cqrs-htmx show setup/v4.12.0:setup/go.mod` pins go-appkit **v0.5.1** (checkout matches, `setup/go.mod:15`). AGENTS/FEATURES/TODO rewritten off the published tag. | tag show + checkout grep, this session |
| A2 | **M4 state corrected:** the `Config.Metrics`/`Config.Version` threading into `appkit.ServiceConfig` is SHIPPED in setup/v4.12.0 (`run_appkit.go:78-79` in the TAG); `Bundle.Run`/`RunHandler` still serve via httputil (RunWithAppkit-only, their config.go docs). All three living docs now state exactly that. | tag + checkout source reads |
| A3 | **Toolchain note made true:** 8 laggards at pass start (AGENTS said 5), 7 after security joined 1.27.1 (side effect of this session's tidy — see d-5). AGENTS now names all 7 with exact directives; `check-go-directives.sh` output reproduced verbatim in-session. | `grep '^go ' */go.mod go.work` + guard run |
| A4 | **Hook error-code contract pinned:** new `hookcodes_test.go` asserts `errorfamily.Code(err)` == `server.drain_hook_failed` / `server.shutdown_hook_failed` (exact) + `TestHooks_EveryHookRunsWhenAnEarlierOneFails`. Root suite `-race` green (6.0s). Closes dedup-episode-1 d-3 / episode-2 f-21. | test run output |
| A5 | **Cross-repo pin guard added** to `scripts/check-pin-drift.sh` (check 2b: cqrs-htmx/v4, go-sse, ssetest, httputil vs the module proxy). It immediately caught **httputil v1.2.0 pinned vs v1.4.0 latest** — the exact unguarded gap; integration bumped + fixture updated + suite green. | guard FAIL→OK transition, live |
| A6 | **Integration at ALL-latest pins:** cqrs-htmx v4.12.0, go-sse v0.6.1, httputil v1.4.0; suite `-race -count=1` green (2.3s), lint 0. The "deliberately NOT setup's pin" charter note updated (pins are now ALIGNED with setup anyway). | test + guard output |
| A7 | **380 inline verdicts** across the 09-17, 09-20 ×2, 09-23, 09-24, and 09-28_22-47 reports (annotate-prose/annotate-rows, dry-run first); `check-rows.py` audited — every flagged row verified open-by-design, zero planted misses. | script outputs + row audit |
| A8 | **Archive migration:** 6 annotated reports + the 09-23 integration-audit HTML (verdict banner inserted) → `docs/status/archived/` (42 files); SUPERB v3 (superseded banner pre-existing), SUPERB v4 + the 11-47 health plan (new EXECUTED banners citing evidence) → `doc/planning/archived/`. Archive gate `grep -rLn '~~' docs/status/archived/ --include='*.md'` prints NOTHING. | git mv log + gate run |
| A9 | **doc/-vs-docs/ resolved:** `docs/status/` is canonical (per the standing user instruction); the entire `doc/status/` tree (reports + archive + index) migrated with `git mv`; `doc/status/` removed from disk; index (`docs/status/README.md`) rebuilt with the decision recorded; all living-doc path references updated (AGENTS :329, ROADMAP :35, TODO_LIST). | tree state + link check |
| A10 | **TODO_LIST rebuilt (harvest):** 64 lines, open-only; 9 new rows, ~25 done-in-code removals (each verified), cross-repo pdg item pointing at the archived report (no content duplication), cqrs v0.6.0 SQLitePath-removal train added (episode-2 f42). | TODO_LIST diff |
| A11 | **Living-doc truthing:** FEATURES health rows "(UNRELEASED at v0.1.1)" → "shipped in v0.1.2"; security section + example/THREAT_MODEL; core section + testkit.Shutdown + hook-code tests; integration table + security×realtime row; Consumers section rewritten (A1/A2). AGENTS: security_realtime + doadapter rows, integration bullet, ritual step 5 (+go-directives), published-API gotcha, 376/377 + linter 0. ROADMAP: consumer line, multi-recorder↔aggregate cross-ref, aggregate gate note, date. | per-file edits + linter run |
| A12 | **Small fixes shipped & verified:** frh `sort.Strings`→`slices.Sort` (suite green, lint 0 after `--fix` for a gci nit); security example PORT-aware demo origin (build/vet/test/lint green); THREAT_MODEL linked from security README + doc.go; realtime README composition section (cites the integration test); `doc/DOMAIN_LANGUAGE.md` + 4 consumer-matchable error codes (exact strings grepped from service.go); Decision 13 (accepted-duplication) in design-decisions.md; CHANGELOG `[Unreleased]` entries for core (runHooks + sortedRoutes + hook tests), health (setStarted), frh (slices.Sort), cqrs (sortedEngineNames). | per-module builds/tests |
| A13 | **pkg.go.dev re-verified live:** health v0.1.2 + frh v0.1.3 pages RENDER (Published Sep 20) — closes the 09-20 train §b1 "never returned to confirm" debt. | fetch of both pages |
| A14 | **README verified current, untouched:** config table matches `config.go` field-for-field (17 fields); "Requires Go 1.26.7" matches the released v0.5.1 tag directive (`git show v0.5.1:go.mod`); middleware order, metrics surface, drain phases all corroborated. The 09-17 audit's "nothing owed" verdict re-confirmed. | grep + tag read |
| A15 | **Concurrent-session absorption:** two new reports appeared in `docs/status/` mid-pass (09-28_23-03 dedup ep. 2, 09-29_00-05 pdg). Read both; their already-executable go-appkit items (hook test, Decision 13, ritual step, DOMAIN_LANGUAGE codes, doc/docs resolution, consumer sweep, structure-linter re-run) were executed BY this pass; their cross-repo items routed into TODO P2/P3. Both reports left CURRENT (their sessions own them). | report reads + TODO diff |

## b) PARTIALLY DONE

| # | What works | What remains | Blocker | Effort |
|---|-----------|--------------|---------|--------|
| B1 | 1.27.1 unification: root, go.work, health, integration, security (the last two joined via tidy this pass/episode-2). | 7 satellites lag (cqrs/docs/realtime 1.26.7; errorpages/flightrecorder/otel 1.27; frh 1.26.5) → guard red. Direction is set by gravity (go-etag v0.6.0 + httputil v1.4.0 both demand ≥1.27.1); the deliberate one-train completion is TODO P2. | USER go-ahead | S/M |
| B2 | Release trains: all four modules' `[Unreleased]` deltas are now accurately CHANGELOG'd (core, security, health, frh) and enumerated in TODO P2. | The trains themselves (API-break checks, dating, tagging, proxy checks, AGENTS/pin updates). cqrs-htmx is waiting on core > v0.5.1 (their go-etag stub-replace drop). | USER timing | M |
| B3 | Consumer-claim accuracy: this pass fixed all known stale claims and added the ritual item. | The standing ritual (re-verify claims ABOUT consumers after each cqrs-htmx release) is a TODO item, not yet a script — nothing mechanical guards claims about OTHER repos' state. | policy decision (see g-2) | M |
| B4 | LSP health: gopls produced live diagnostics throughout this session (hookcodes_test.go errors surfaced and resolved immediately) — the "workspace broken" state from 09-23/24 is empirically GONE under the go.work 1.27.1 + per-module GOWORK=off regime. | Not systematically verified across all modules/editors; the TODO item stays until the unification train lands. | none | S |
| B5 | The two current reports (23-03, 00-05): their §f lists are fully harvested into TODO (episode-2 items are near-duplicates of episode-1 numbering; the pdg list is routed as one pointer row). | Their own annotation+archive pass (not mine to do — their sessions own those reports). | next docs-health pass | S |
| B6 | Guard coverage: pin-drift now guards both sides of the cross-repo contracts. | go-directives remains string-equality brittle (1.26 vs 1.26.7), has no `toolchain`-directive coverage, and is not wired into pre-commit — one TODO P2 item owns all three. | none | S |

## c) NOT STARTED (observed this session, deliberately untouched)

- The 7-satellite unification train (USER-gated, see g-1) — every workspace command still needs `GOTOOLCHAIN=go1.27.1 GOWORK=off`.
- Release trains (USER-gated, see g-2).
- Upstream filings (go-sse ReplayFiltered, httputil Logging, NewServerListener, go-health pack ×3, samber/do note) — drafts ready in `doc/feedback/outgoing/`, filing USER-gated.
- errorpages `statusRecorder` → `httputil.ResponseRecorder` (verified still hand-rolled at errorpages/errorpages.go:83,205).
- Browser CSP pass over `DashboardHardenedPreset` (no browser in this env).
- govulncheck on health + security (network-blocked env).
- art-dupl suppression-heap audit (~431 groups never inspected across both dedup episodes — the "zero harmful duplication" claim is only proven for the shown set).
- pdg's own backlog (header relabel, `(replace …)` rendering, doctor subcommand, `Graph.Consumers`-side severed-key fix) — routed, their repo.
- CI remote state (dead SSH secret — P1, owner-only).

## d) TOTALLY FUCKED UP

| # | What | Severity | Root cause | Mitigation |
|---|------|----------|------------|------------|
| D1 | **Wrote a verdict ahead of the evidence — again.** Annotated 09-23 §f-12 as "recorded — AGENTS gotchas carry the published-API lesson" BEFORE the gotcha existed in AGENTS; caught it on self-review and made it true (added the gotcha line, verified 377/377). For one tool-call window the record claimed something that wasn't. | Process (caught in-session) | The 09-17 pass's d-1 failure class, repeated by me in the pass that enforces the rule | Marker discipline: edit-then-annotate, never annotate-then-edit; I did the correction in the right order only after the slip |
| D2 | **security-example multiedit shipped 7 compile errors in one blast** — my replacement introduced `origin :=` (string) shadow-colliding with the middleware chain that CALLS `origin(...)`, plus undefined `originStr`. gopls flagged all 7 instantly; fixed in one follow-up edit. | Caught by gates, 1 wasted round | Edited a function body I'd only seen via a 12-line grep window; the multiedit old_strings were individually unique but semantically colliding | View the WHOLE function before surgical example edits (E1 lesson from the 09-28 report, not internalized) |
| D3 | **Two wasted cycles on the 09-28 report's move:** `git mv` failed (file is UNTRACKED — daemon hadn't committed it), then plain `mv` failed because I looked for it in `doc/status/` when it lived in `docs/status/` all along. | Noise | Path assumptions during a 14-command migration chain; the `&&` chain also silently stopped at the failure, deferring 4 later moves to a second call | `ls` the exact source path immediately before each mv; split migrations into verified chunks |
| D4 | **The `&&` migration chain partially applied on failure** — archived/, 4 reports moved; HTML/index/plans deferred. No damage (moves are order-independent here), but a failed chained batch leaves a state that LOOKS half-migrated, which is exactly when a second agent (or the daemon) can interleave. | Process risk | One long chain instead of verified chunks | Verify state between migration steps, not after the batch |
| D5 | **`go mod tidy` flipped security's directive 1.27 → 1.27.1 as an unanticipated side effect** during a DOCS pass. Favorable (reduced laggards to 7) and the daemon captured it, but I rewrote the AGENTS toolchain note twice because my first rewrite used the pre-tidy facts. | Low, favorable drift | Didn't predict tidy raises the directive to the toolchain floor | `grep '^go '` after every tidy; write AGENTS facts only after the tree stops moving |
| D6 | **My first AGENTS toolchain rewrite itself carried a wrong count** (I wrote "8 laggards" context into the note structure before re-running the guard post-tidy; corrected to 7 with the full per-module list in the same edit round). | Caught same-round | Same root cause as D5 | Facts from a single frozen snapshot: run all probes, THEN write |
| D7 | **No pre-edit baseline for root/integration** before the pin bump and hook-test additions (tests passed first try — luck, not discipline). The episode-2 e-list flags this exact pattern as REPEATED; I repeated it in the same night. | Discipline | Baseline-first still not a reflex | 10-second `go test` before touching anything, even docs-adjacent code fixes |
| D8 | **README config-table check used a broken grep** (stray `\t` warning, empty output) and I moved on, closing it only near the end of the pass. The table WAS fine — but for ~40 minutes a Critical-class check (README field lies) ran on a no-op command. | Latent | Trusted a syntactically-broken probe instead of re-running it | Never accept empty output from a probe that warned; fix the probe first |

## e) WHAT WE SHOULD IMPROVE

1. **Guards must check BOTH sides of a contract.** `check-pin-drift.sh` guarded OUR pins for 11 days while the OTHER side (httputil v1.4.0 vs v1.2.0 pinned, cqrs-htmx v4.12.0 vs v4.9.0) rotted silently. The 2b extension is the template: for every "X must equal Y" invariant, ask what enforces Y itself.
2. **Verdict-after-evidence as a hard ordering rule** (D1): annotate only after the action's verification output exists in-session. The annotate scripts' atomicity doesn't help if the underlying claim is premature.
3. **Whole-function context before example edits** (D2): grep windows are for locating, not for editing. `view` the function, then edit.
4. **Migration hygiene** (D3/D4): `ls` source before mv; untracked files can't `git mv`; split chains; verify between chunks — especially in a daemon environment where every intermediate state may be committed.
5. **Tidy is not fact-neutral** (D5/D6): it can raise directives. Freeze the snapshot (all probes at once), then write docs once.
6. **Read sibling-session reports BEFORE the pass, not at migration time** (A15/D5-adjacent): the 23-03 report already contained half my TODO diff. A first-five-minutes `ls docs/status/` + skim would have deduped the harvest once instead of twice.
7. **Baseline-first, no exceptions** (D7) — the e-list of BOTH dedup episodes says it; this pass still skipped it because "tests passed first try." The rule exists for the day they don't.
8. **Broken probes invalidate downstream conclusions** (D8): a warning in a verification command means the check didn't run. Fix or replace the probe before citing its output.
9. **`docs/` (the Go module dir) hosting the canonical status tree is working but remains structurally odd** — the module's `go.mod`/`docs.go` live beside timestamped reports. It's the user's explicit choice; the index README records it. Flagged once here, not re-litigated.

## f) NEXT — up to 50 things to get done (grounded ONLY in this session; impact-ordered)

**P0 — unblock (USER-gated, one answer each)**

1. Green-light the 1.27.1 unification train: bump the 7 laggards (`go mod tidy` each), keep go.work/root/AGENTS/`documentedGoDirective` in sync, re-run the 11-module matrix + lint sweeps, drop the `GOTOOLCHAIN=go1.27.1 GOWORK=off` prefixes and stale GOEXPERIMENT notes from AGENTS.
2. After (1): verify gopls/LSP across modules (this session's live diagnostics already suggest it recovers).
3. Green-light the batched release trains: core v0.6.0 (testkit.Shutdown + hook-code tests + refactors), security v0.2.0 (example + THREAT_MODEL + example-only dep), health v0.1.3 (NonceFromContext fix + setStarted), frh v0.1.4 (Register + determinism + slices.Sort).
4. After core ships: bump integration pins + `documentedPins` + doc.go in one change; simplify `security_realtime_test.go` to the now-published `ts.Shutdown`.
5. CI: restore `SSH_PRIVATE_KEY` (owner-only) + gate the ssh-agent step off `pull_request`; then one fully green master run.
6. Decide go.work tracked status (commit it — it now matches the train — or CI-generate); note it shapes pdg attributions (folds member modules).
7. Wire `check-go-directives.sh` + `check-dependabot-parity.sh` into BuildFlow pre-commit; make the guard semver-aware; add `toolchain`-directive coverage.
8. Answer the consumer-claim policy (g-2): published-tag-verified claims + the after-each-release ritual, or drop exact consumer pins from AGENTS.
9. cqrs v0.6.0 train: remove the deprecated `SQLitePath` alias (announced breaking change; plan + API-break diff).
10. Extend `documentedPins` to assert go-health/samber-do/go-flightrecorder too (episode-2 f-7; the fixture pins family + 4 cross-repo only).

**P1 — close this pass's residual gaps**

11. errorpages: `statusRecorder` → `httputil.ResponseRecorder` (USER gate, ~10 lines; verified open this pass).
12. File the upstream asks (USER gate): go-sse `ReplayFiltered`, httputil Logging context, `NewServerListener` go/no-go, go-health pack (sentinel + Evaluate-cache + restart/rearm), samber/do lazy-healthy note.
13. Browser CSP pass over `DashboardHardenedPreset` (needs a browser-capable machine).
14. Run govulncheck on health + security (networked machine).
15. E2E gap: cqrs wrapper lifecycle through a live appkit Service (register/dispatch/query + in-flight drain).
16. Fresh-consumer proxy smoke for core before its next tag (go.mod changed this pass via refactors? — verify at train time).
17. Decide the consumer-claim ritual mechanics (script vs checklist) if (8) keeps pins.
18. Post-unification otel benchmark re-baseline (toolchain jump invalidates 09-16 numbers on principle).
19. Verify the example E2Es (`example/`, `health/example`) once on the unified toolchain.

**P2 — quality/polish (batch on touch)**

20. art-dupl: ONE audit pass over the ~431 suppressed/non-actionable groups (both episodes skipped it; the zero-harmful claim is unproven).
21. art-dupl standing policy: recurring ritual step vs ad-hoc + whether to baseline the 7 accepted groups; record the corrected invocation (drop `--type-aware` with `--suggest-generics`) somewhere durable (AGENTS is at cap — module README or recipes).
22. `metrics.go` typed `seriesKey` (kill the `|`-join/SplitN pair).
23. testkit drain-window assertion helper (drain tests hand-roll it).
24. Named `type Hook func(context.Context) error` to shorten `runHooks`/config signatures.
25. `errors.Join` multi-failure message-shape test.
26. Hardened-dashboard test: assert `frame-ancestors 'none'` + sorted directive order.
27. SSE+health combined lifecycle test (drain during an open stream).
28. health/example `-hardened` mode wired like the T21 composition.
29. AGENTS deep slim-down (extract a section to module READMEs; the cap has ZERO headroom at 376/377).
30. Cross-project lesson → crush-config `references/lessons.md`: "check GOTOOLCHAIN/go.work parity before trusting AGENTS build commands" (carried since 09-24).
31. pdg (their repo): header relabel, `(replace …)` rendering, `doctor` subcommand, `/v4`-matching regression test, extend the severed-key fix to `Graph.Consumers`/serve/renderer, benchmark `ConsumersFromModules`, triage `go-auto-upgrade [graph]` diff.go:121, review `7fde3823` churn, cut a pdg release. Full lists in the two current reports' §f.
32. prompt-crusher-exec: resolve the go.mod merge conflict (needs the user's intent — carried across three reports now), then tidy + build + module-path reconciliation.
33. errorpages/realtime/docs/otel/flightrecorder suites were NOT re-run this pass (untouched code) — fold into the post-train matrix.
34. Confirm `sortedRoutes`/`sortedEngineNames` nil-vs-empty callers once more post-train (behavioral delta noted in episode-2 d-5).
35. Verify errorpages example's local-dev `replace ../` never leaks into a tag (tag-hygiene spot check, episode-2 f-38).

**P3 — watchlist/demand-gated (unchanged, restated for completeness)**

36. cqrs-htmx v5-window revisit (their ADR-0052; master runs 191 commits past setup/v4.12.0); cut their setup release (doubly motivated — via-pins render `(unpinned)`).
37. External consumers in the release ritual: rolls-royce-mtuGoHelpCenter-golang + papdashboard pin v0.5.1 — decide whether the ritual verifies THEIR integrations (the proxy check covers only OUR modules).
38. Cordis bridge (2 of 3 triggers pending), PapDashboard reverse adoption, nixpkgs > 1.26.7 watch, dprint exit-14 upstream, samber-do-auditlog (demand), battery waves W3-W5 (demand), core v1.0.0 exit-criteria graduation, go-health v0.3.0 aggregate adoption (floor gate resolves with train (1)).
39. Composition-gap watchlist (otel×realtime, otel×dashboard, security×errorpages, flightrecorder E2E, docs-module E2E — now +1: SSE×health from (27)).
40. cqrs README cookbook re-verification on the next go-cqrs-lite release.

_(40 grounded items; the remaining 10 slots stay empty rather than padded — the two source reports' §f lists are already routed, and fabricating scope violates the no-research constraint.)_

## g) QUESTIONS I CANNOT ANSWER MYSELF (3)

1. **Unification train — go for 1.27.1 across all 11 now?** The bump side has effectively won (go-etag v0.6.0 + httputil v1.4.0 require it; 5 of 11 already moved, 2 of those by tidy gravity). Say "unify" and the 7 laggards + matrix + AGENTS prefix cleanup land in one train — or say "hold" and I pin the guard expectation to the mixed state so CI isn't permanently red.
2. **Release trains — now or batch?** Four modules carry `[Unreleased]` deltas (enumerated in TODO P2), and cqrs-htmx is BLOCKED on core > v0.5.1 (their go-etag stub-replace can't drop until we ship). "Train all four" / "core only (unblocks them)" / "batch" — your call; the ritual handles the mechanics either way.
3. **doc/ vs docs/ — confirm the scope of the resolution.** I made `docs/status/` canonical (your standing instruction) and migrated ONLY the status tree; planning/feedback/recipes/research stay under `doc/` (singular). Is that the intended end state, or do you want the whole `doc/` tree consolidated under `docs/` (which would touch ~100 referenced paths across all living docs and module READMEs)?

---

**Awaiting instructions.** Pending on green-light: (1) the unification train, (2) release trains, (3) doc-tree scope confirmation — plus the standing USER gates (upstream filings, errorpages swap, browser CSP environment).
