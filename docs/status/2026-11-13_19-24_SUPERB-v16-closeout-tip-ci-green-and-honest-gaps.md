# Status — SUPERB v16 close-out: tip-CI green at 43/44, the honest gaps, and the next-50

**Date:** 2026-11-13 19:24 CET · **Scope:** the execution half of this session — from the SUPERB v16 plan delivery (`docs/planning/2026-11-13_15-54_SUPERB-v16-pareto-floor-train-and-owner-unblock-plan.md`) through the full-plan run: floor train, preset v0.1.1, consumer verification, the owner memo, the cqrs/test/docs batches, and this close-out. The earlier half (docs-health corpus pass + the preset-deadlock discovery) is the 15-39 report; the plan itself is the 15-54 planning doc.

## Scope Line — VERIFIED vs NOT

**VERIFIED (command/artifact):**

- **Tip-CI: 43/44 jobs SUCCESS on the final tree (`25945b8`, run `43456484511`)** — the ONLY failure is the designed `ci-preflight` secret red. This proves on origin: the go 1.27.4 floor fleet (build+vet+test per module), the preset v0.1.1 integration pin + `documentedPins`, the release-state sync, and **status-index GREEN** (the 14:52 interim failure was a mid-flight index, resolved in-tree).
- **`test (./preset)`: success** in the 14:52 run on `738c44a` — the SSE shutdown-deadlock fix CI-verified (yesterday's five-day red saga closed).
- **M2 floor train executed + verified**: 15/15 go.mods + go.work + 14 `.golangci.yml` pins + integration fixture + charter-guard default + script pins + 18 living docs; directives/parity/charter guards OK; `verify-all.sh` 30 PASS / 0 FAIL; root + integration `-race` green; **govulncheck CLEAN on root/otel/preset/cqrs/integration** — GO-2026-6859 gone; the 11-11/11-12/11-13 nightly failures confirmed as the SAME finding (GO-2026-6859; DB published ~2026-11-10, the 11-10 run was the last success).
- **M3 preset v0.1.1 released end-to-end**: API diff vs v0.1.0 = ZERO delta; annotated tag; integration pin + `documentedPins` bumped via the insteadOf trick; pre-tag gate GREEN after it caught two same-train release-state drifts (AGENTS latest-per-module + a stale FEATURES security-example pin claim — both fixed, the guards worked exactly as designed); tag pushed; **proxy `.zip` 200 on first probe**; the DEFERRED gosum-canonical leg re-run GREEN post-push; **fresh-consumer probe GREEN** ("preset v0.1.1 compiles + constructs", fresh GOMODCACHE).
- **M6 consumer-verification batch**: `verify-consumer.sh` GREEN for config@v0.1.1 and preset@v0.1.1; `example-smoke.sh` GREEN, all 8 legs incl. "preset graceful shutdown".
- **M4 owner-decision memo**: eight rulings in one document (`docs/planning/2026-11-13_16-20_owner-decision-batch.md`), D014 + art-dupl TODO rows linked.
- **M5 cqrs test-debt batch**: `appendTestEvent`/`waitFor`/`testEventSeq` consolidated into `testhelpers_test.go`; `startNoOpProjection` stays staleness-scoped BY DECISION (the readiness/metrics sites assert phase transitions an all-in-one helper would hide); the root-CHANGELOG split brain RESOLVED (the `cqrs` deployment bullet moved to the cqrs module CHANGELOG per the module-CHANGELOGs-only convention; root [Unreleased] now states the rule). Suite `-race` green, lint 0.
- **M7 engine-lesson pair**: the `StartProjections(ctx)` run-ctx warning shipped in the cqrs README API table; the watermill ask SHARPENED AT SOURCE before drafting — `GoChannel.Publish` is not unconditionally synchronous: the blocking is `BlockPublishUntilSubscriberAck` (go-cqrs-lite defaults it TRUE, cqrs-lint E014 enforces it), while watermill v1.5.3 `pubsub.go:86`'s comment claims the opposite — the real defect is a docs gap; draft staged (`docs/feedback/outgoing/2026-11-13_upstream-ask-watermill-gochannel-publish-doc.md`, filing user-gated).
- **M9 conventions**: docs-tree archive conventions recorded in AGENTS (154/220 lines); the cordis planning doc gained the inline trigger-status line (go/v0.1.0 tagged — its verdict-table "untagged" cell is stale point-in-time wording); the docs-gates upstream notes drafted (`docs/feedback/outgoing/2026-11-13_upstream-notes-docs-gates.md`: routed≠resolved + the undocumented strict-list grammar + the DF-decay report ask).
- **M10 AsType sweep: RESOLVED AS NO-OP** — repo-wide grep finds ZERO `errors.As(` call sites (production AND tests) and NINE `errors.AsType[E]` already in use across 7 files; erraudit confirmed present on PATH. Row closed with evidence.
- **M16 corpus remainder**: `annotate_max` already pinned 0 in `.docs-gates.json` (moot); a GLUED TODO row (missing newline between two rows, an old formatting defect) repaired.
- **M20**: the adoption-recipes row verified ALREADY TRUTHFUL (consumer-repo ownership stated) — no edit needed.
- Full battery **5/5 PASS** and status-index OK at close; all work committed and pushed (`0b76bfe` tip).

**NOT VERIFIED / NOT DONE (and why):**

- **The raced interim run (`43452173968`, tree `2eb50b5`) was NOT re-run** — its pin-drift failure is the documented tag-push vs tip-CI visibility race (checkout's AGENTS claimed v0.1.0 while origin already served v0.1.1); the tip run on the newer tree supersedes it, and `gh run rerun` would only re-prove a superseded tree.
- **M8 (D014 implementation)** — blocked on the owner ruling (batch memo D1).
- **M11–M13 (AsType module batches)** — MOOT (the sweep resolved as no-op; nothing to migrate).
- **M14/M15 (machine ops: /tmp sweeper, BuildFlow rebuild, cache-DB VACUUM, `interrogate`)** — NOT attempted (L-impact ops; the session budget went to the P1 + 4% tiers).
- **M19 (scaffold template spec fill)** — NOT attempted (90 min, clean next-session task).
- **M17 (go-sse absorb rehearsal)** — demand-gated (the breaking ssetest change is already absorbed at ssetest v0.5.0; the templ-breakage hold on go-sse v0.7.0 remains the live gate).
- **M18 (W4 polite spike)** — demand-gated on the H8 ruling.
- **Tonight's govulncheck + flake nightlies** — expected GREEN for the first time since 11-10; confirmation is tomorrow-morning's check.

---

## a) FULLY DONE

1. **M1 push + tip-CI verification** — the daemon delivered the commits; `test (./preset)` green in CI on the tree carrying the fix; the tip run 43/44 with the designed red only.
2. **M2 floor train go 1.27.4** — the full 2026-11-07 recipe executed same-day: directives, lint pins, fixtures, script pins, 18 living docs' current-floor prose with HISTORICAL train facts deliberately kept verbatim; guards + fleet sweep + race runs + govulncheck all green.
3. **M3 preset v0.1.1 release train** — date → API-diff proof → tag → integration pin (insteadOf) → pre-tag gate (two caught drifts fixed) → push → proxy 200 → deferred gosum leg → fresh-consumer green → AGENTS/FEATURES truth-synced.
4. **M6 consumer-verification batch** — both verify-consumer probes + the 8-leg smoke.
5. **M4 owner-decision batch memo** — eight rulings, one ask, TODO rows linked.
6. **M5 cqrs test-debt batch** — consolidation + the phase-transition decline verdict + the split-brain move, all proven.
7. **M7 engine-lesson pair** — README warning shipped; the watermill draft sharpened at source (the naive claim corrected to the documented mechanism before staging — the verify-before-filing gate DID its job: the source read changed the ask's content).
8. **M9 conventions batch** — archive conventions in AGENTS; cordis trigger line; upstream notes drafted.
9. **M10 AsType survey → no-op closure** — with the fleet-wide grep evidence.
10. **M16 corpus remainder** — notes drafted; annotate_max moot; glued-row repair.
11. **M21 close-out** — this report + index + battery green; M20 verified truthful.

## b) PARTIALLY DONE

1. **The nightly watch** — the floor train SHOULD flip tonight's govulncheck to green; confirmation is tomorrow's check (the run fires 03:24).
2. **The owner-decision process** — the memo is delivered; the rulings themselves await the owner (they gate M8, the art-dupl policy implementation, H8/B6/B8, auth-bridge legs, the convention rows, the watermill filing).
3. **SUPERB v16 plan completion** — 1% + 4% tiers DONE; 20% tier: M9/M10/M16 done, M14/M15 open; remainder: M19 open, M17/M18 gated, M21 done.
4. **The docs-gates upstream notes** — drafted, filing user-gated (same gate as every external ask).

## c) NOT STARTED

1. **M19** scaffold template spec fill (the plan's 90-minute task).
2. **M14/M15** machine ops (the /tmp sweeper policy is the one with real hygiene value).
3. **M8** D014 implementation (owner-gated).
4. **M17/M18** (trigger-gated).
5. **The eight owner rulings** (D1–D8).
6. **cqrs release shape** — the module CHANGELOG [Unreleased] now carries the deployment bullet + test-refactor notes; cut `cqrs/v0.9.1` or accumulate: owner call, rides the D014 answer.
7. **Tonight's nightly confirmation** (tomorrow morning).

## d) TOTALLY FUCKED UP

1. **I broke `waitFor` during the helper move and the TEST SUITE DIDN'T CATCH IT.** My sed line-range extraction dropped `t.Fatalf("timed out waiting for %s", what)` from the moved body; the suite stayed green (the line only fires on timeout) and only the `unparam` linter surfaced it. Root cause: mechanical line-surgery on code — the wrong tool; `lsp_replace_symbol` or a whole-function move was right. Four fix cycles to converge (missing brace, import, Fatalf), each visible only in the next tool's error. The worst moment of the session: a test helper that silently stops failing on timeout is a time bomb for every future cqrs test.
2. **I skipped the freshness pre-flight TWICE in one day.** Morning: the TODO rewrite flipped a routed item's anchor. Evening: the TODO row closures flipped ANOTHER (v14 report item 44) and my new report's §f item 25 was born too generic to harvest. The gate caught both post-hoc; both fixes were genuine improvements (a considered-decline verdict + a re-aligned anchor), but the discipline exists precisely to prevent the surprise.
3. **The plan doc carried a WRONG step** (M3.3: "integration does NOT pin preset" — it does, at v0.1.0). Written from assumption, caught mid-execution, handled via the health-v0.1.1 ritual — but a plan step that asserts an unverified negative is a trap for the next reader too.
4. **Known gotchas I re-learned the hard way**: `chmod -R u+w` before rm on module-cache entries (AGENTS documents it — I let `rm -rf` spew hundreds of Permission-denied lines first); the GOMODCACHE inside the module dir breaks tidy; the tag-push vs tip-CI visibility race (AGENTS documents it — I re-derived it from the run instead of recognizing it instantly).
5. **A partial-match edit mangled the art-dupl TODO row** (truncated old_string matched a prefix, leaving a dangling `§f#6/#16):**` fragment) — caught on the immediate re-read, repaired; but partial-match edits on mega-lines are now a three-time-offender class.
6. **The preset CHANGELOG carried a hidden second `[Unreleased]` heading** — my rename produced two `[0.1.1]` sections; caught by the post-edit grep, fixed by deleting the empty trailing one. A 10-second structure survey before the rename would have avoided it.

## e) WHAT WE SHOULD IMPROVE

1. **Code moves: symbol-level tools ONLY.** Line-range sed on Go source is banned for me going forward — the Fatalf loss proves tests don't guard test-helpers.
2. **Pre-flight before EVERY living-doc rewrite** — twice-bitten; the 2-second check replaces a 15-minute surprise.
3. **Plan steps must carry their verification command** — a step that asserts a negative ("integration does not pin X") without a grep is a draft, not a step.
4. **Read the AGENTS gotcha paragraph BEFORE the operation it covers** (chmod+rm, tag-race, cache placement were all documented). The gotchas exist because each cost a session once; re-paying them is pure waste.
5. **CHANGELOG renames need a heading census first** (`grep -c '^## \[Unreleased\]'`).
6. **The touched-modules guard classifies test-only Go changes as "doc-only"** — it skipped verification on exactly the commit where I broke a helper. A `--tests` habit (the flag exists) would have caught the break at commit time.
7. **The owner-decision memo pattern works** — eight gates, one ask. Repeat it for every future ruling cluster instead of per-row pings.

## f) Up to 50 things we should get done next

Impact-ordered. `[S]` = surfaced this session; `[R]` = standing queue.

**P1 — close this session's last loop**

1. `[S]` Confirm tonight's govulncheck nightly GREEN (03:24 — first candidate clean night since 11-10; the floor train should clear GO-2026-6859 from the runner).
2. `[S]` Confirm the flake nightly alongside it.
3. `[R]` Collect the EIGHT owner rulings from `docs/planning/2026-11-13_16-20_owner-decision-batch.md` — one owner hour unblocks items 4-11 below.

**P2 — unblocked work (post-ruling or independent)**

4. `[S]` **M8 D014 implementation** (30 min once D1 rules: tag or suppress, test, CHANGELOG).
5. `[S]` **M19 scaffold template spec fill** (90 min; repo creation itself stays owner-gated).
6. `[S]` **M14 machine ops 1**: /tmp `go-build*` sweeper policy + script (TMPDIR-guarded, stale-only) + BuildFlow binary rebuild (74a9212 → 2d6bbee).
7. `[S]` **M15 machine ops 2**: cache-DB VACUUM + legacy state-DB deletion + storage-layout fallback verify + `interrogate` install.
8. `[S]` cqrs release shape (owner): cut `cqrs/v0.9.1` (the [Unreleased] carries the deployment bullet + test refactors) or accumulate into the next feature train.
9. `[S]` The art-dupl policy implementation once D2 rules (warn-first sizing run before fatal).
10. `[R]` H8 route-seam ruling execution (B6/B8 httpx batteries + the `svc.Routes()` answer) — memo exists.
11. `[R]` Auth-bridge ruling execution — memo exists (staged recommendation).
12. `[R]` File the watermill ask when approved (draft ready + source-verified).
13. `[R]` File the docs-gates notes when approved.
14. `[R]` Machine default bump: per-user nix profile go1.27.2 → 1.27.4 at the next nix pull (cosmetic under GOTOOLCHAIN=auto).
15. `[R]` The plan-revision + rev-N convention rulings (D5/D6) — fold into one naming answer.
16. `[R]` The pma daemon dossier §6 rulings (D8) — retire the TODO row.

**P2/P3 — standing queue (carried, verified current today)**

17. `[R]` Monitor `check-consumer-claims.sh` for setup (cqrs-htmx) and PapDashboard refreshes — their pins ride core v0.10.0-era claims; their next tags fire the guard.
18. `[R]` go-sse watch: the templ-breakage hold (v0.7.0 + cqrs-htmx v4.16.0) is the live gate; the breaking ssetest change is ALREADY absorbed at v0.5.0 (confirmed today).
19. `[R]` M17 go-sse absorb rehearsal when the templ unblock trigger fires.
20. `[R]` W4 `polite` battery spike when H8 lands (go-etag client+metrics v0.6.0 relevance).
21. `[R]` Core TLS G1 — deferred register (PapDashboard `PAP_TLS_CERT/KEY` trigger; the spike §6 sketch is the starting point).
22. `[R]` Cordis bridge — trigger 1/3 met (tagged); awaits the consumer statement + core v1.0.0 exit-criteria graduation.
23. `[R]` go-error-family utilization engagement (46/100) — demand-gated.
24. `[R]` httputil server-seam refactor (72/100) — upstream repo.
25. `[R]` PapDashboard reverse-adoption watch (criterion 5 of core v1.0.0 exit criteria rides it).
26. `[R]` Adoption-recipe instances in the consumer repos (dnsblockd, ChastityAPI, PapDashboard security swap, pbx-lens cmdguard) — their repos, their sessions.
27. `[R]` Scaffold repo creation (owner) once M19's spec is filled and reviewed.
28. `[R]` Six-month preset adoption re-sweep (due 2027-05; `scripts/count-adopters.sh`).
29. `[R]` integration `documentedPins` bumps ride the next release trains (standard ritual).
30. `[R]` The `interrogate` unavailable-tool inspection (rides M15).

**Process self-commitments (from §e, tracked here so they're checkable)**

31. `[S]` Symbol-level tools only for code moves (no line-range sed on Go source).
32. `[S]` `docs-gates freshness . --pre-flight` before every living-doc rewrite — no exceptions, twice-bitten.
33. `[S]` Plan steps carry their verification grep/command inline.
34. `[S]` CHANGELOG heading census before any section rename.
35. `[S]` `--tests` habit on test-touching commits (the touched-modules guard skips test-only Go changes as doc-only).
36. `[S]` Read the relevant AGENTS gotcha paragraph BEFORE operations it covers.

**Items 37-50: declined as padding** — items 1-36 plus the standing trigger register in TODO_LIST P3 and the AGENTS Deferred Register cover every observed open thread; the dedup session's discipline (don't invent count-filler) applies.

## g) Questions only you can answer

1. **The eight owner rulings** (D1–D8 in `docs/planning/2026-11-13_16-20_owner-decision-batch.md`): D014 json-tag vs suppress; art-dupl exact-count policy; H8 seam; auth-bridge; plan-revision + rev-N conventions; watermill filing; pma dossier §6. One line of answers unblocks items 4, 9, 10, 11, 15, 16 and the two filings.
2. **cqrs release shape**: cut `cqrs/v0.9.1` now (the [Unreleased] carries the system.Standard deployment bullet + the test refactors — a real consumer-visible change sitting unpublished), or accumulate into the next feature train?
3. **Machine-ops scheduling**: M14/M15 touch the shared box's hygiene (/tmp sweeper, cache VACUUM) — a dedicated ops session, or fold into the next natural pause between trains?
