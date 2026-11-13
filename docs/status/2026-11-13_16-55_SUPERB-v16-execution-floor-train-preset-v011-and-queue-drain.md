# Status — SUPERB v16 execution: floor train, preset v0.1.1, and the queue drain

**Date:** 2026-11-13 16:55 CET · **Scope:** executing the SUPERB v16 plan (`docs/planning/2026-11-13_15-54_SUPERB-v16-pareto-floor-train-and-owner-unblock-plan.md`) — the 1% tier complete, the 4% tier complete except owner-gated halves, selective 20%/remainder legs.

## Scope Line — VERIFIED vs NOT

**VERIFIED (command/artifact):**

- **M1 ✅** push + tip-CI: the daemon delivered the 13+ commits; the 14:52 UTC CI run on `738c44a` (the final code-bearing tree before the docs tail) shows **`test (./preset): success`** and preset proxy-smoke success — the deadlock fix is CI-verified. The run's reds: designed ci-preflight + a MID-FLIGHT status-index (a concurrent session's report landed between index updates; the current tree is index-green, re-proven by the tip run).
- **M2 ✅ floor train go 1.27.3 → 1.27.4**: 15/15 go.mods + go.work + 14 lint pins + integration fixture + charter-guard default + script pins + 18 living docs' current-floor prose (historical train facts kept verbatim). VERIFIED: directives guard green; parity/charter/touched-modules OK; `verify-all.sh` ALL GREEN (30 PASS, 0 FAIL); root + integration `-race` green; **govulncheck CLEAN on root/otel/preset/cqrs/integration — GO-2026-6859 GONE**; the 11-11/11-12 nightly failures confirmed = the SAME finding (all three GO-2026-6859; the vuln DB published it ~2026-11-10).
- **M3 ✅ preset v0.1.1 released**: CHANGELOG dated (incl. removing an empty duplicate heading the Unreleased rename exposed), API diff vs v0.1.0 = ZERO delta (doc prose only), annotated tag created, integration pin + `documentedPins` bumped via the insteadOf trick BEFORE push, **pre-tag gate GREEN** (after the guards caught two same-train release-state drifts — AGENTS latest-per-module + a stale FEATURES security-example pin claim — both fixed), tag pushed, **proxy `.zip` 200 on first probe**, the deferred **gosum-canonical leg re-run GREEN**, **fresh-consumer probe GREEN** ("preset v0.1.1 compiles + constructs" from a fresh GOMODCACHE).
- **M6 ✅ consumer-verification batch**: `verify-consumer.sh` green for config@v0.1.1 AND preset@v0.1.1; `example-smoke.sh` GREEN all 8 legs incl. "preset graceful shutdown".
- **M4 ✅ owner-decision batch memo**: `docs/planning/2026-11-13_16-20_owner-decision-batch.md` — eight rulings (D014, art-dupl policy, H8, auth-bridge, plan-revision, rev-N, watermill filing, daemon dossier) with options + recommendations; the D014 and art-dupl TODO rows link to it.
- **M5 ✅ cqrs test-debt batch**: helpers consolidated into `testhelpers_test.go` (suite `-race` green, lint 0); `startNoOpProjection` stays staleness-scoped BY DECISION (readiness/metrics assert phase transitions an all-in-one helper would hide); the root-CHANGELOG split brain RESOLVED (the `cqrs` deployment bullet moved to the cqrs module CHANGELOG per the module-CHANGELOGs-only convention).
- **M7 ✅ engine-lesson pair**: the `StartProjections(ctx)` run-ctx warning shipped in the cqrs README; the watermill ask SHARPENED AT SOURCE — `GoChannel.Publish` is not unconditionally synchronous; the blocking is `BlockPublishUntilSubscriberAck` (go-cqrs-lite defaults TRUE, E014 enforces) while `pubsub.go:86`'s comment claims the opposite — draft staged (`docs/feedback/outgoing/2026-11-13_upstream-ask-watermill-gochannel-publish-doc.md`, filing user-gated).
- **M9 ✅ conventions batch**: docs-tree archive conventions in AGENTS (154/220 lines); cordis doc trigger-status line; docs-gates upstream notes drafted.
- **M10 ✅ AsType sweep resolved as NO-OP**: zero `errors.As(` sites fleet-wide (production + tests), nine `errors.AsType[E]` already in use across 7 files; row closed with evidence.
- **M16 ✅ corpus remainder**: the three docs-gates upstream notes drafted; `annotate_max` already pinned 0 (moot); a GLUED TODO row repaired.
- Full battery **5/5 PASS** at close; status-index OK.

**NOT VERIFIED / NOT DONE (and why):**

- **The tip-CI run on the final tree (`25945b8`) was IN FLIGHT at close** (run `43456484511`, started 15:49 UTC). The interim run on `2eb50b5` failed `pin-drift` — the DOCUMENTED tag-push vs tip-CI visibility race (its checkout's AGENTS still claimed preset v0.1.0 while origin already served the v0.1.1 tag); per the gotcha the fix is a rerun, and the tip run supersedes it. Close-out requires confirming the tip run green.
- **M8 (D014 implementation)** — blocked on the owner ruling (memo D1).
- **M11–M13 (AsType batches)** — MOOT: the sweep resolved as no-op (zero `errors.As` sites).
- **M14/M15 (machine ops)** — not attempted this session (L-impact ops tasks; the session budget went to the P1/4% tiers). Rows remain.
- **M17/M18** — demand-gated (go-sse break unreleased; W4 polite gated on H8 context).
- **M19 (scaffold spec fill)** — not attempted (context budget; the plan's 90-minute estimate makes it a clean next-session task).
- **M20** — verified ALREADY TRUTHFUL (the recipes row already states consumer-repo ownership; no edit needed).

---

## a) FULLY DONE

1. **Floor train go 1.27.4** — the whole recipe executed + verified (see Scope Line); GO-2026-6859 cleared; nightly-watch obligation satisfied by the same-day confirmation of all three nightly failures as the same finding.
2. **preset v0.1.1 released end-to-end** — date → API-diff proof → tag → integration pin via insteadOf → pre-tag gate (two caught drifts fixed) → push → proxy 200 → deferred gosum leg green → fresh-consumer green → AGENTS/FEATURES truth-synced.
3. **Push + CI verification of the deadlock fix** — `test (./preset)` green in CI on the tree carrying the fix; the five-day red saga closed.
4. **Consumer-verification batch** — verify-consumer ×2 + example-smoke 8/8 legs.
5. **Owner-decision memo** — eight rulings in one document, TODO rows linked.
6. **cqrs test-debt batch** — consolidation + the phase-transition decline verdict + the split-brain move, all suite/lint-proven.
7. **Engine-lesson pair** — README warning shipped; the watermill draft sharpened at source (the naive "Publish is synchronous" claim corrected to the BlockPublishUntilSubscriberAck mechanism before staging).
8. **Conventions + corpus remainder + AsType no-op + the glued-row repair** — see Scope Line.
9. **All living docs truth-synced** (AGENTS ×5 paragraphs, TODO rows ×6, README floors, FEATURES preset + security-example rows).

## b) PARTIALLY DONE

1. **The session-close ritual** — tip-CI in flight at close; confirm-and-record is the one open step.
2. **SUPERB v16 plan execution** — 1% + 4% tiers done (M8 halves owner-gated), 20% tier: M10 done (as no-op), M16 done, M14/M15 open, M11–M13 moot; remainder: M19 open, M17/M18 gated, M21's report = this file.
3. **The watermill ask** — drafted + source-verified, filing user-gated.

## c) NOT STARTED

1. **M19 scaffold template spec fill** (90 min, clean next-session task).
2. **M14/M15 machine ops** (/tmp sweeper, BuildFlow rebuild, VACUUM, interrogate).
3. **M17 go-sse absorb rehearsal** (gated on go-sse shipping the breaking change) and **M18 W4 polite spike** (gated on H8).
4. **M8 D014 implementation** (gated on the owner ruling).
5. **The eight owner rulings themselves** (the memo awaits answers; the one-ask design means ONE owner session unblocks everything).

## d) TOTALLY FUCKED UP

1. **I broke `waitFor` during the helper move** — my line-range extraction dropped the `t.Fatalf("timed out waiting for %s", what)` tail, and the unparam linter (not the tests!) caught it. A test helper that silently stops failing on timeout is exactly the Verschlimmbesser class the plan guards against; the mechanical line-sed surgery was the wrong tool — `lsp_replace_symbol` or a whole-block move was right. Restored + verified, but it cost four fix cycles.
2. **I skipped pre-flight AGAIN** (the TODO row closures) and flipped a historical item's anchor (v14 report item 44) red — the same failure mode as this morning, now twice in one day. The fix (a considered Won't-implement with real reasoning) improved the corpus, but the sequence was wrong both times.
3. **The M3 plan assumed integration does NOT pin preset** — it does. The plan's step 3.3 was written from unverified assumption; the health-v0.3.1 ritual saved the execution, but the plan doc itself carried a wrong step.
4. **One mangled TODO row from a partial-match edit** (the art-dupl row) — caught by re-reading the edit output, repaired immediately; still, partial-match edits on mega-lines keep being a trap (third occurrence across sessions).

## e) WHAT WE SHOULD IMPROVE

1. **Never move code with line-range sed** — even "simple" helper moves. Symbol-level tools or full-file rewrite; the Fatalf loss was invisible to the test suite (it only matters on timeout) and only the linter saw it.
2. **Pre-flight is now proven necessary TWICE in one day** — make it a literal step in the TODO-edit flow, not a memory.
3. **Release-train plans must verify pin-state assumptions** (`grep module integration/go.mod`) before writing steps — the plan is a point-in-time artifact but its wrong step cost a mid-train pivot.
4. **The status-index mid-flight race** — concurrent sessions' reports land between index updates; the tree is only index-consistent at commit boundaries. The tip run re-proving it is fine, but the interim red is noise a single-session index commit would avoid.
5. **The `→ routed` + freshness dual semantics** now have THREE documented workarounds (hand-strike, re-anchor, considered-decline) — the upstream note (drafted today) is the real fix.

## f) Up to 50 things to get done next

1. Confirm the tip-CI run (`43456484511`) green; rerun the raced `2eb50b5` run if the ritual wants per-tree proof.
2. Collect the EIGHT owner rulings (memo: `docs/planning/2026-11-13_16-20_owner-decision-batch.md`) — one owner hour.
3. Implement D014 per the ruling (M8, 30 min).
4. Implement the art-dupl policy per the ruling (warn-first sizing, then fatal).
5. M19 scaffold template spec fill (90 min).
6. M14 machine ops 1 (/tmp sweeper + BuildFlow rebuild).
7. M15 machine ops 2 (VACUUM + state-DB + interrogate).
8. Watch tonight's govulncheck nightly — expect GREEN (first clean night since 11-10).
9. Watch the flake nightly alongside it.
10. H8 ruling execution (B6/B8 + `svc.Routes()` answer) once D3 lands.
11. Auth-bridge ruling execution once D4 lands.
12. File the watermill ask when approved (draft ready).
13. File the docs-gates notes when approved.
14. Monitor `check-consumer-claims.sh` for setup/v4.18+ and PapDashboard refreshes (their pins still ride core v0.10.0-era claims; the guard fires on their next tags).
15. cqrs next train: the [Unreleased] section now carries the deployment bullet (moved from root) + test notes — cut `cqrs/v0.9.1` or fold into the next feature train (owner call, rides the D014 answer).
16. Fold the "module CHANGELOGs only" adoption into AGENTS' release ritual prose (one sentence, next AGENTS touch).
17. preset v0.1.1 consumer notice to PapDashboard (their preset usage is example-level; only if they adopt preset legs — watch the claims guard).
18. Re-verify the go-sse watch: the breaking `ssetest` change is already absorbed at ssetest v0.5.0 (confirmed in the watchlist row today); the templ-breakage hold on go-sse v0.7.0 + cqrs-htmx v4.16.0 remains the live gate.
19. M17 go-sse absorb rehearsal when the templ unblock trigger fires.
20. M18 W4 polite spike when H8 lands.
21. Scaffold repo creation (owner) once M19's spec is reviewed.
22. Adoption-recipe instances in the consumer repos (dnsblockd, ChastityAPI, PapDashboard, pbx-lens — their repos, their sessions).
23. Cordis bridge: awaits consumer statement + core v1.0.0 exit (trigger 2/3 and 3/3).
24. Core TLS G1: awaits the PapDashboard `PAP_TLS_CERT/KEY` trigger.
25. W5 batteries demand-gated: W5-C1 is the go-sse `WithOnDrop` TODO in realtime's hub (the fleet audit's go-sse 81/100 row carries it) — opens when a consumer needs drop callbacks.
26. go-error-family engagement (46/100): demand-gated.
27. httputil server-seam refactor: upstream repo.
28. Six-month preset adoption re-sweep (due 2027-05; `count-adopters.sh` re-run).
29. Machine default bump: the per-user nix profile go1.27.2 → 1.27.4 at the next comfortable nix pull (the GOTOOLCHAIN=auto path makes it cosmetic).
30. Keep AGENTS ≤220 lines (154 — headroom fine).

Items 31-50: **declined as padding** — the actionable queue above plus the standing trigger register in TODO_LIST P3 covers everything observed; inventing more would be fabrication, not planning (the dedup session's §f discipline).

## g) Questions only you can answer

1. **D014 + art-dupl policy + the other six rulings** — the batch memo (`docs/planning/2026-11-13_16-20_owner-decision-batch.md`) needs your answers; each unblocks a bounded task.
2. **cqrs release shape**: the [Unreleased] section now carries the deployment bullet + test refactors — cut `cqrs/v0.9.1` as a standalone patch train, or accumulate into the next feature train?
3. **Machine-ops priority**: M14/M15 are L-impact but touch the shared box's hygiene (the /tmp sweeper especially) — schedule them as a dedicated ops session, or fold into the next natural pause?
