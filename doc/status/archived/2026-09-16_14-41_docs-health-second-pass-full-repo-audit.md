# Status Report — Docs-Health Second Pass (Full-Repo AUDIT: HARVEST + VERIFY + ANNOTATE + ARCHIVE)

**Date:** 2026-09-16 14:41 CEST
**Session scope:** "View ALL `**/2026-0*` files, execute the docs-health skill properly, make the six living docs SUPERB, archive fully-done files with inline strikethroughs." One session, one repo (go-appkit). This report covers only this session's work and what it noticed in passing.
**Companion artifact:** the inline health report printed in-session (Accuracy 96% → verified-clean, Fitness 98% post-pass).
**Repo state at close:** all session work committed by the auto-daemon (heuristic commits, git-mv renames preserved); one 1-line edit (`docs/status/2026-09-16_09-38…md`) pending the next daemon sweep. A `git status --short` immediately before writing this report showed exactly that one file.

---

## Signals

| Signals              | Count |
| -------------------- | ----- |
| a) Fully done        | 10    |
| b) Partially done    | 6     |
| c) Not started       | 6     |
| d) Totally fucked up | 6     |
| f) Next tasks listed | 50    |
| g) Questions asked   | 3     |

---

## a) FULLY DONE

1. **All 41 `**/2026-0*` files inventoried and dispositioned.** 33 archived markdown reports/plans (27 from the morning pass + 6 archived this session), 4 fresh 2026-09-16 reports annotated-not-archived (genuinely open work), 4 HTML reports (1 corrected inline, 1 resolution-annotated, 2 left per the morning pass's classification), the canonical batteries spec and 2 trigger-gated research docs left alone.
2. **VERIFY with hard evidence:** full 10/10-module `-race` sweep green (core 6.0s, cqrs 4.2s, integration 1.7s — after my code edit); `go vet` + `go build` green; golangci-lint 0 issues on flightrecorder, realtime, health, flightrecorderhealth; `go-structure-linter` 0 findings; `gofmt` clean; `git tag -l` (16 tags) reconciled against AGENTS/TODO release state; CI matrix + proxy-smoke job + dependabot verified present.
3. **TODO_LIST claims spot-verified before routing (grep, not trust):** `X-Accel-Buffering` still missing from `realtime/handler.go`; `SnapshotHandler` still silent-200; health module still has zero `NoTimeout` mentions outside tests and lacks example/contract/benchmark files; errorpages still hand-rolls `statusRecorder`; the two root test helpers had **0 usages repo-wide** before deletion.
4. **HARVEST closed the remaining report→backlog gaps:** 15 items routed into TODO_LIST (P2: dashboard-CSP/WriteTimeout verification — open since 2026-09-04 and never routed until today; `httputil.Server` composition spike. P3: integration expansion, health quality parity, error-classification sweep, fr ops preset, dashboard hardening passthrough, go-sse `ReplayFiltered` ask, statusRecorder USER GATE, dependency-currency proof, health E2E re-run, shutdown-log-level decision, release-state single owner, golines root cause, nosurf verification, version cuts, cqrs/root doc-test polish) — every item deduped against existing entries (e.g. `WithOnDrop` folded into battery W5 C1 with an explicit don't-double-track note) and every citation carries a `file.md §section` source.
5. **ROADMAP gained 3 raw ideas** (multi-recorder coordination ADR, cqrs ops/recipes backlog: CBOR→JSON helper + shared-recorder demo + DLQ recipes + snapshot accessors, route-cardinality fuzz guard) — each ported only after an annotation verdict claimed ROADMAP ownership (claims made true within the same pass).
6. **ANNOTATE at scale with the skill's own scripts (dry-run first):** ~500 inline verdicts (`~~item~~ done — evidence` / `Won't implement`) across nine files — the five 2026-09-04 status reports (16-52, 17-17, 17-38, 18-57, 21-03), the SUPERB plan (C-task level with an F-inheritance note), and the four 2026-09-16 reports. The morning pass's malformed row-number-only strikes (`| ~~ | 2 |`) were stripped from 4 files and replaced with real per-cell strikethroughs carrying evidence.
7. **ARCHIVE per the skill:** 6 fully-resolved files `git mv`'d (5 status reports + the SUPERB plan) — 33 archived total. Completeness gate `grep -rLn '~~' docs/status/archived/ docs/planning/archived/` → **0 files**. Post-move citation sweep fixed 5 references (TODO_LIST ×4, 18-57's plan pointer), and the 09-38 audit's now-false "five reports stay in docs/status/" claim was inline-corrected in place.
8. **The /tmp ghost E2E test was recovered before reboot killed it.** `/tmp/appkit-otel-verify/verify_test.go` (the A/B proof of the pattern-propagation fix — published httputil v1.1.1 → span `"GET"`; fixed → `"GET /users/{id}"` + `http.route`) is now verbatim at `docs/planning/2026-09-16_otel-pattern-pin-test.md` with landing instructions for the integration module, and TODO_LIST P2 points at it instead of /tmp.
9. **Fix-on-sight sweep (7 code/doc defects, all verified after):** deleted dead `expectError`/`freePort` (0 usages; root suite re-run green); completed the `AGENTS.md` `middleware.go` table row (truncated mid-sentence since 09-04, flagged twice, never fixed); bumped 6 stale `go: 1.26.5` lint pins to 1.26.7 (open item from three separate reports); removed the Release-State/Deferred-Register duplication; harmonized the two-voice otel gotcha bullet; added the "doc snippets are code" Testing rule (line-budget neutral — structure linter re-run green); realtime + otel CHANGELOG `[Unreleased]` "Documented" entries.
10. **HTML annotations, non-destructive:** the 09-15 telemetry status doc got two inline HTML-comment corrections (the "Logging's ctx helper" mislabel → ClientIPMiddleware; 3 sites → 5 sites + `ff44c5f`), closing 08-40 §d-7/§f-13; the 09-16 ecosystem deep-dive got a resolution banner (4 fixed findings, module count corrected to ten, open items → TODO_LIST pointers) — closing 08-53 §c-12/§b-5.

## b) PARTIALLY DONE

1. ~~**"View ALL files" was full-fidelity for every `.md` but shallow for 3 HTMLs.** The 2026-08-15 research, 2026-08-16 planning, and 2026-09-04 research HTML files were classified via grep + the morning pass's verdicts, not opened end-to-end. I concurred with LEAVE-ALONE without an independent full read.~~ done (resolved 2026-09-17 — this pass opened all three HTMLs and banner-annotated them (research kept with banners, the executed 08-16 plan archived))
2. ~~**infertypeargs verification was 3 of 7 sites.** I personally verified explicit type args at `cqrs/commands.go` ×2 and `realtime/hub.go`; the 4 test-file sites (`cqrs/commands_test.go:73,90,187,245`) were trusted from the 08-53 report's claim that the debt was paid, plus clean builds. A gopls pass would close this definitively.~~ done (closed 2026-09-17 — verified OBSOLETE (zero markers, zero findings))
3. ~~**Lint coverage was 4 of 10 modules** (the ones with pending verification debt plus the two I touched). cqrs, docs-mod, errorpages, otel, integration were race-swept but not linted this session; the AGENTS "0 issues across the board" claim still rests on the 2026-09-04 sweep for those.~~ done (closed 2026-09-17 — every module linted sequentially, 0 issues)
4. ~~**AGENTS slim-down stayed line-level.** Two lines freed, two added (net 0, 376 wc / 377 linter-counted — exactly at the new budget), but the file is still ~58 KB; the deep "what graduates to module READMEs" decision remains open (09-38 §b-2).~~ done (still open — routed to TODO_LIST P3 (AGENTS deep slim-down decision), 2026-09-17)
5. ~~**08-40 §b-8 (core `-race` suite against the fixed local httputil via temporary replace)** was deliberately skipped and annotation-marked as such — it is cheap pre-train assurance nobody has run yet.~~ done (moot — the train shipped 2026-09-16; published-tag pins verify the fix)
6. ~~**dprint behavior on my ~50 edited table rows is daemon-trusted, not verified.** The morning pass flagged the same gap; no formatter dry-run exists in this environment.~~ **Won't implement — unverifiable in this env — tracked via the dprint exit-14 TODO item.**

## c) NOT STARTED (deliberate or missed)

1. **All routed P1/P2 work** — license posture, docs-ghost fix, OTEL release train, logging posture, realtime correctness, W2 security, W1 leftovers. Docs-health does not write features; every item is sourced and owned in TODO_LIST.
2. ~~**nosurf source verification** (now TODO_LIST P3) — a ~10-minute read of justinas/nosurf that I deferred to the backlog instead of doing in-session.~~ done (done 2026-09-16 — verified TRUE against justinas/nosurf v1.2.0 source; httputil note pushed (a03db5c))
3. ~~**`docs/status/README.md` index** — declined this pass (w-marked in 09-38 §f-47); the archived dirs have convention READMEs, the flat status dir stays small.~~ done (done 2026-09-17 — the index was created by the 05-52 session (C29))
4. **LSP hygiene** — the stale `testhelpers_test.go` typecheck diagnostic persisted all session (CLI proved it false twice); I never restarted the LSP to clear the cache.
5. **lychee link check** — not installed; I ran a manual internal-link check over the seven living docs (clean) instead.
6. **`docs/status/archived/` counts** — the archived-dir README describes the convention but carries no index/counts; left as-is (see 3).

## d) TOTALLY FUCKED UP

1. **An annotate-tool call returned "4 written" for 5 specs and I moved on without root-causing it.** The 17-17 §b batch (specs 1,2,3,3.1,4) reported `[1,2,3,4]`. I verified the final state was correct (item 3 carries the right verdict; "3.1" was a malformed spec on my part — the prose tool takes integers), but "outcome correct, cause unexplained" is exactly the instrument-discipline failure the 09-15 session owned. It cost nothing because the tool is atomic and shape-checking; the _process_ was still guess-then-check.
2. **I fed the tools file shapes they can't parse — twice — and fell back to hand edits.** Multi-item lines (09-38 §f items 34–45 share one line) and a code-span `~~` literal (item 32) broke the annotator; the 21-03 35-spec batch broke once on my own shell quoting (`eval` mangling `"## f)"`). All failures were atomic (zero partial writes) and the three manual fallbacks were single-line, prefix-asserted edits — but they are precisely the hand-rolling the skill bans, and a dry-run per file shape would have caught all three before the first real run.
3. **A stale-read guard bounce cost a round trip on 17-17 §c** because I read the file via `sed` in bash, edited via the tool, and the guard tracks view-tool reads. Mechanical, but it happened _after_ the session had already demonstrated (stale injected AGENTS) that reads must go through the tracking tool.
4. **I overrode a prior session's flagged concern with a judgment call.** 08-53 called the health `mount_test.go:379` `time.Sleep(5ms)` a probable rule violation; I annotated it as the house `waitForRunning` pattern (deadline + poll + lint 0 issues — which it is). The evidence is solid, but silently overruling another session's HIGH-flagged concern in an annotation, rather than noting the disagreement explicitly in the verdict text, hides a legitimate editorial disagreement from the next reader.
5. **The completeness gate is weaker than it looks and I report it as stronger.** `grep -rLn '~~' archived/` passes trivially because the README convention docs themselves contain literal `~~`; it would still pass if a future archived file's only tilde-bearing content was inside inline code spans. The gate proves _presence of tildes_, not _presence of verdicts_. I stated it as "0 files" (true) without stating its blindness.
6. **Claimed-then-made-true ordering on ROADMAP.** I wrote annotation verdicts pointing at ROADMAP ownership (fuzz guard, recipes) before adding the corresponding ROADMAP lines — the window was under a minute and the end state is consistent, but for one tool cycle the docs claimed ownership that didn't exist. Same class as "don't publish numbers you haven't produced."

**Did I lie?** No. Two statements were thinner than their presentation: the archive-gate blindness (d-5) and the infertypeargs "already explicit" claim resting on 3 personally-verified of 7 sites (b-2). Both are corrected here.

## e) WHAT WE SHOULD IMPROVE

1. **Dry-run per file SHAPE, not per file.** Table vs prose vs multi-item-line vs code-span-literals — one `--dry-run` probe of each shape class before any real batch would have prevented every tool failure this session (d-2) and yesterday's hand-rolling incident (09-38 §d-6).
2. **All reads through the tracking tool.** The stale-read bounce (d-3) and the stale injected AGENTS snapshot are the same lesson from opposite directions: bash `sed` reads and conversation-start snapshots are not state.
3. **Fix-on-sight sweep as session step zero.** The truncated AGENTS row and the `go: 1.26.5` pins sat through at least four sessions that each read the relevant docs. A standing 60-second sweep (grep the known rot signatures) catches them on contact.
4. **/tmp artifacts are recovered the moment they prove valuable, not at the next audit.** The pin test survived because I happened to run the recovery the same day it was flagged urgent. "Ephemeral" and "load-bearing" must never coexist for a tool cycle.
5. **When overriding a prior session's verdict, say so in the verdict.** d-4's annotation should have read "corrected: this is the house pattern (deadline-bounded poll), contrary to 08-53's flag" — disagreements belong in the record.
6. **Explain tool anomalies before proceeding** (d-1): atomic tools make guess-then-check _safe_, not _correct_. One diagnostic round trip on the 5→4 mystery would have bought a root cause.
7. **State the limits of gates.** The archive-completeness gate (d-5) is a tilde-presence check. A stricter gate: grep for `done at\|done —\|Won't implement` per archived file, excluding READMEs.
8. **Lint-matrix claims need the full matrix.** "Modules sit at 0 issues" is only true as of the last session that ran all ten; this session refreshed four. Either run all ten or scope the claim by date.

## f) UP TO 50 THINGS TO GET DONE NEXT

**This session's direct follow-ups:**

1. ~~Decide the annotation-depth house standard (grouped verdicts for declared-brainstorm blocks vs strict per-item) — g-1 below; applied by practice this pass, needs blessing.~~ done (answered by practice — the annotation standard is recorded in the archived READMEs (2026-09-16/17))
2. ~~Confirm the HTML classification (leave-alone for the two 2026-08 research/planning HTMLs) or commission resolution banners for them — g-3 below.~~ done (resolved 2026-09-17 — docs-health pass banner-annotated the research HTMLs and archived the 08-16 plan HTML)
3. ~~Decide the single owner for release-state facts (AGENTS Release State vs TODO_LIST header) — TODO_LIST P3, asked since three sessions.~~ done (DECIDED 2026-09-16 — AGENTS Release State is the single owner; rule text in doc/status/README.md)
4. ~~AGENTS slim-down decision: what graduates to module READMEs (Release Ritual? per-module Gotchas?) to get under the ~30 KB the morning pass flagged.~~ done (still open — routed to TODO_LIST P3 (AGENTS deep slim-down decision), 2026-09-17)
5. ~~Run the full 10-module lint sweep so the "0 issues across the board" claim carries a fresh date (6 modules not linted this session).~~ done (done 2026-09-17 — the 08-08 session linted every module sequentially, all 0 issues)
6. ~~Verify the 4 remaining infertypeargs sites (`cqrs/commands_test.go:73,90,187,245`) via gopls; restart the LSP to clear the stale `testhelpers_test.go` diagnostic.~~ done (done 2026-09-17 — verified OBSOLETE (zero markers in the file, zero linter findings))
7. ~~Verify dprint didn't reflow the ~50 edited table rows on the next commit hook run (daemon-trusted until then).~~ **Won't implement — unverifiable in this env (no dprint dry-run) — tracked via the dprint exit-14 TODO item.**
8. ~~Delete `/tmp/appkit-otel-verify` once its recovery commit is confirmed pushed (the durable copy is in-repo).~~ done (moot — /tmp is ephemeral; the durable in-repo copy was retired to doc/planning/archived/ (08-08))
9. ~~Root CHANGELOG: confirm no entry is owed for the test-helper deletion (test-only, core behavior untouched — believed correct, worth one glance at the next train).~~ done (confirmed correct — test-only deletion, core behavior untouched)
10. ~~When the docs-ghost fix lands: update the three archived files that narrate it (07-07 §d, FEATURES docs section is living — only the archived narration stays as-is).~~ done (correct as-is — archived narration stays historical per the annotation standard)

**Standing user-gated (repeated for one-place visibility; already in TODO_LIST — do not re-harvest):**

11. ~~License posture decision (P2 USER GATE; blocks pkg.go.dev godoc for every module; cqrs-htmx is MIT).~~ done (DECIDED 2026-09-16 — stays proprietary, permanent)
12. ~~Docs-ghost fix path A vs B, then re-tag `docs/v0.3.0` + fresh-consumer proxy test (P1).~~ done (done 2026-09-16 — path A executed; docs/v0.3.0 proxy-proven)
13. ~~OTEL release train authorization: tag httputil → bump core+otel → re-tag otel → land the recovered pin test in `integration/` → delete the README known-issue block (P2).~~ done (SHIPPED 2026-09-16 — full train: httputil v1.2.0, otel v0.1.1, integration pin test landed)
14. ~~Logging posture: default WARN vs sampling vs consumer logger (+benchstat) (P2).~~ done (DECIDED 2026-09-16 — status quo INFO + tuning docs (core README Log volume))
15. ~~pkg.go.dev re-crawl verification after 11–13 (P1).~~ done (VERIFIED 2026-09-17 — every module page renders)

**Standing feature/correctness backlog (owned, one-place visibility):**

16. ~~Realtime: `X-Accel-Buffering: no` + SSE `event: error` before abort + failure-path test (P2).~~ done (SHIPPED 2026-09-16 — realtime v0.1.1 (X-Accel-Buffering + error event, wire-pinned))
17. ~~Dashboard CSP + SSE longevity verification (chromedp or manual browser run) + `WriteTimeout: NoTimeout` posture (P2, restored today).~~ done (VERIFIED 2026-09-16 — both fears resolved with evidence; strict-CSP browser pass routed to TODO_LIST P2)
18. ~~`httputil.Server` composition spike → Service refactor → Core TLS unlock (P2, USER GATE on API posture).~~ done (done 2026-09-16 — spike verdict: BLOCKED on upstream API (composition-spike-verdict.md))
19. ~~W2 security module (A2/A3/A4/A8/A6 quick wins first; MaxKeys cap is a HARD requirement) (P2).~~ done (SHIPPED 2026-09-16 — security/v0.1.0, all 8 batteries)
20. ~~W1 leftovers: G2 Prometheus (auth-wired, stable metric names), F5 BuildInfo, E1 testkit seed (P2).~~ done (SHIPPED 2026-09-17 — core v0.5.0 (G2/F5/E1))
21. ~~Telemetry documentation bundle, 6 items (P2).~~ done (SHIPPED 2026-09-16 — doc/TELEMETRY.md, all 6 items)
22. ~~Integration-module expansion: 3 seam tests + composition-contract suite + add `./integration` to the CI matrix (P3).~~ done (DONE 2026-09-16/17 — 5/5 sub-items (F104 scoped into the cqrs module))
23. ~~Health-module quality parity: godoc examples, benchmark, fuzz, contract assertion, aggregate example, govulncheck, conflict-semantics + BasePath tests (P3).~~ done (done 2026-09-16/17 — all landed except govulncheck (env-blocked; TODO_LIST P2))
24. ~~Error-classification sweep: otel setup sentinels + cqrs drain; `errorfamilytest` adoption (P3).~~ done (DONE 2026-09-16 — otel sentinels + cqrs drain classified; errorfamilytest adopted)
25. ~~Flightrecorder polish: explicit not-enabled status, `statusError` → error-family, `WithRecorderOptions` preset, MetricsHook → otel meter (P3).~~ done (DONE 2026-09-16/17 — polish + OpsRecorderPreset + MetricsHook + real-capture E2E)
26. ~~Dashboard hardening passthrough (`WithShutdownDrain`, `WithNonce`+`RecommendedCSP`, `WithRateLimit`) (P3).~~ done (SHIPPED 2026-09-16 — DashboardHardenedPreset (rate limiting deliberately out))
27. ~~Upstream go-sse ask: dedup-aware `ReplayFiltered` (P3, verify-before-filing).~~ done (DRAFTED 2026-09-16 — filing USER-GATED (doc/feedback/outgoing/))
28. ~~errorpages `statusRecorder` → `httputil.ResponseRecorder` (P3 USER GATE).~~ done (STILL OPEN — TODO_LIST P2 USER GATE (open since 2026-09-16))
29. ~~Dependency-currency proof: `git ls-remote --tags` across the family vs pins (P3).~~ done (PROVEN 2026-09-16 — zero drift across 14 family repos vs pins)
30. ~~Health example live E2E re-run post-two-phase-drain (P3).~~ done (PASS 2026-09-16 — lockstep 503s observed live)
31. ~~`shutdown phase skipped` log-level decision (P3, restored today).~~ done (DECIDED 2026-09-16 — stays INFO (rationale indexed in doc/TELEMETRY.md))
32. ~~golines LSP-vs-CLI root cause (P3).~~ done (ROOT-CAUSED 2026-09-16 — stale LSP buffer; the CLI was clean)
33. ~~nosurf "forks internally" verification + httputil-side CSRF known-limitation note (P3).~~ done (VERIFIED TRUE 2026-09-16 — nosurf v1.2.0 source; httputil note pushed (a03db5c))
34. ~~Version cuts for pending behavior changes (health drain fix, frh bumps, missing error-family CHANGELOG lines) (P3).~~ done (DONE — health v0.1.1 + frh v0.1.2 tagged; v0.10.1 lines verified in all CHANGELOGs (2026-09-17))
35. ~~cqrs/root doc-test polish: `Status()`-as-slice README line, staleness boundary property test, root example predates v0.4.0 hooks (P3).~~ done (DONE 2026-09-16 — cqrs README slice + monotonicity test + root example hooks demo)
36. ~~Toolchain bump past 1.26.7 when nixpkgs carries it (P2).~~ done (CHECKED 2026-09-16 — nixpkgs still 1.26.7; carried by the TODO_LIST P3 watchlist)
37. ~~Benchstat install attempt (`nix run nixpkgs#benchstat`) — never actually probed (P3 candidate).~~ done (still a candidate — nix run never probed (TODO_LIST P2 benchstat item))
38. ~~Cut health/frh versions per the Release Ritual when the train runs (rides 13).~~ done (DONE — health v0.1.1 + frh v0.1.2 tagged 2026-09-16)

**Roadmap-grade (owned by ROADMAP; listed so nothing lives only in snapshots):**

~~39. cordis bridge triggers (P3). 40. PapDashboard reverse-adoption door (P3). 41. cqrs encryption/signing opt-ins on demand (P3). 42. W3–W5 battery waves per the canonical spec (P3). 43. Multi-recorder coordination ADR (added to ROADMAP today). 44. cqrs ops/recipes backlog (added today). 45. Route-cardinality fuzz guard (added today). 46. Core v1.0.0 exit-criteria graduation (fold in the documented-wiring-test lesson). 47. httputil `Logging` ctx-aware emit + F2 timing battery (P3). 48. BuildFlow dprint exit-14 upstream fix (P3/Deferred Register). 49. `docs/status/README.md` index (declined this pass; revisit if the flat dir grows). 50. The two-voice AGENTS bullet pattern: adopt a rule that verification sentences _replace_ enumerations instead of appending to them (this pass harmonized one; the rule prevents the next).~~ done — all owned as of the 2026-09-17 docs-health pass: 39/40/41/42/47/48 live in TODO_LIST P3 (39 updated: cordis `go/v0.1.0` tagged = trigger 1/3 met); 43/44/45 live in ROADMAP (landed there as claimed); 46 in TODO_LIST P2 (exit-criteria draft); 49 DONE 2026-09-17 (the index was created by the 05-52 session); 50 adopted as working practice.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. ~~**Annotation-depth standard for archived brainstorm lists** (carried over from the morning pass's unanswered g-3): I applied "per-item verdicts where individually decidable, grouped verdicts only for blocks the report itself declared as routed/ROADMAP-fuel" — ~500 markers today. Is that the house standard, or do you want full per-item git-archaeology on 50-item tables (~3–4× the effort, real noise risk)? My recommendation: bless the current practice and record it in the archived-dirs' README so the next pass doesn't re-decide it.~~ done (answered by practice — the standard is recorded in the archived READMEs)
2. ~~**Release-state single owner:** AGENTS.md Release State, TODO_LIST header, or a dedicated `docs/RELEASE_STATE.md`? Three sessions have flagged the triplication; today all copies agree because I manually reconciled them. One-line answer picks the owner and the other two become pointers.~~ done (DECIDED 2026-09-16 — AGENTS Release State single owner; rule in doc/status/README.md)
3. ~~**The four HTML reports:** keep the current LEAVE-ALONE posture for the two 2026-08 research/planning HTMLs (superseded by archived md reports covering the same sessions), or should they get the same resolution-banner treatment the 09-15/09-16 HTMLs received today? I cannot judge their residual readership value without you.~~ done (resolved 2026-09-17 — research HTMLs banner-annotated; the executed 08-16 plan HTML archived)

---

_Everything behavioral above was executed and verified inside this session (sweeps, lints, greps, gate runs, tag/CI reads). The verification debts I am consciously carrying forward: b-1 (3 HTMLs unopened), b-2 (4 unverified infertypeargs sites), b-3 (6 modules unlinted), b-6 (dprint unverified), and the gate-blindness caveat in d-5._
