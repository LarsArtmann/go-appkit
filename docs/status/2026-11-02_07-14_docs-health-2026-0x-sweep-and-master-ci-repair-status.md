# Status Report — Docs-Health 2026-0x Sweep, Master-Tip CI Repair, and TODO_LIST Rebuild

**Date:** 2026-11-02 07:14 CET
**Scope of this report:** the single session that (1) repaired master-tip's red CI (release-state + pin-contract doc drift behind the health v0.3.0 train and a dependabot integration bump), (2) executed the docs-health AUDIT over ALL open `**/2026-0*` files (annotate → check-rows → archive 9 reports → index manifest), (3) rebuilt TODO_LIST per its own charter, (4) truth-synced README/AGENTS/FEATURES/ROADMAP for the health v0.3.0 train and the httputil#41 ship. No other research.

**Scope discipline — what this report VERIFIED (command/artifact):**

- Master-tip CI job breakdown: `gh run view 41326012420` (push of 23fa8ca: `test (./integration)` X at Test(race), `pin-drift` X, `ci-preflight` X by design, 20+ jobs ✓).
- Root causes reproduced + fixed LOCALLY: `./scripts/check-pin-drift.sh` FAIL(exit 1, "health: AGENTS claims v0.2.0 but the newest tag is health/v0.3.0") → **28 OK, exit 0** after the AGENTS fix; `cd integration && GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1` → **ok 7.2s** after the `documentedPins` fixture fix.
- Tag/remote truth: `git ls-remote --tags origin` (health/v0.3.0^{}`= a28523d, matching local),`git rev-list --count origin/master..HEAD` = 0 at session start.
- health module after the godoc edit: build + vet + `go test -race` green, `golangci-lint` **0 issues**.
- Annotation gates: `annotate.go check-rows docs/status/2026-0*.md` → **9/9 COMPLETE**; completeness `grep -L '~~'` over the 9 MOVED files → empty; whole-archive gate `grep -rLn '~~' docs/status/archived/ --include='*.md'` → empty; `emit-keys` round-trips read IN FULL before every spec.
- Living-doc hygiene: TODO `^- \[x\]` count = **0**; AGENTS **133/220** lines; `check-go-directives.sh` + `check-workspace-charter.sh` exit 0; all 9 cited doc paths exist on disk.
- prompt-crusher (their repo): `/home/lars/projects/prompt-crusher/go.mod` read directly — 0 conflict markers, `module github.com/larsartmann/prompt-crusher`, last go.mod touch `a082213` (2026-10-25).

**What this report did NOT verify (and why):**

- **CI on the new tip** — nothing was pushed this session (owner-gated push per standing practice; P1 row carries the acceptance ritual). All greens are local.
- **FEATURES sections other than health/Consumers/core** — spot-checked only, not row-by-row re-verified (owned by earlier passes + guards).
- **ROADMAP raw-ideas rows** — not fact-checked one-by-one.
- **The 36 open 2026-10+ reports** — out of this pass's `2026-0*` charter; index-level treatment only.
- **3 research HTMLs + the samber-do arch-review doc** — classified LEAVE (point-in-time references); inspected at title/head level only.
- **pkg.go.dev** — not re-probed this session (TODO row owns it; last check 2026-10-18).

**Session in one sentence:** the health v0.3.0 train and a daemon-committed dependabot bump had outrun the repo's own documentation contract — this session closed that debt, then applied the docs-health pipeline to every open pre-2026-10 report and made the TODO list tell the truth again.

---

## Headline numbers

| Metric                       | Value                                                                                                                                                                                                                                                                                                                                                 |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Master-tip CI reds diagnosed | 2 real (`test (./integration)`, `pin-drift`) + 1 designed (`ci-preflight`) — both real reds fixed in-tree, verified green locally                                                                                                                                                                                                                     |
| Reports annotated            | 9 (every open `docs/status/2026-06..09` report) — 33 tool-applied verdicts + ~60 exact-match marker refreshes                                                                                                                                                                                                                                         |
| Reports archived             | 9 via `git mv` (`archived/` 44 → **53 files**) + bulk-archive manifest in the index                                                                                                                                                                                                                                                                   |
| TODO_LIST                    | 16 done `[x]` rows purged, 3 stale rows removed, 3 trimmed, 2 live USER questions consolidated — **0 `[x]` remain**                                                                                                                                                                                                                                   |
| Split brains closed          | README floor (1.27.1 vs 1.27.2), AGENTS release state (health v0.2.0 vs v0.3.0), documentedPins fixture, TODO watchlist (go-health#7 "still OPEN" vs CLOSED COMPLETED), ROADMAP #41 line, 2 feedback drafts ("DRAFT, NOT filed"), spike-verdict "BLOCKED" banner, cordis trigger line, TODO row 65 SSE+health leg (found in self-review post-rebuild) |
| Commits                      | `561a476` (22 files, +212/−183) — push REMAINS OWNER-GATED                                                                                                                                                                                                                                                                                            |

---

## a) FULLY DONE

1. **Master-tip CI reds root-caused and fixed in tree.** The 2026-11-01 daemon commit (`23fa8ca`) swept dependabot's integration bump (health v0.2.0→v0.3.0, go-health v0.5.1→v0.8.1) past the `documentedPins` fixture AND the AGENTS release state — the next CI run went red on exactly the two guards built for this class. Fixes: fixture → `health v0.3.0` / `go-health v0.8.1` with the composition-contract comment rewritten (MVS now resolves v0.8.1 even though frh still pins v0.5.1 — documented); AGENTS Release State gained the full health v0.3.0 train entry. Verified: integration `-race` green (7.2s), pin-drift 28 OK.
2. **The health v0.3.0 train's missed same-train release-state updates landed** (the exact class Release Ritual step 3 exists for): AGENTS train entry (go-health v0.8.1 — v0.8.0's own CI was red; recovered-panic now FAILS CLOSED; `WithHealthRecorder` fails at boot with `ErrRecorderNotApplicable` = the go-health#7 ask SHIPPED upstream; art-dupl gate's missing-executable-bit fix noted), FEATURES +2 rows + "shipped through v0.3.0" line (evidence test names verified at `probe_test.go:88,307`, `example_test.go:56`), health CHANGELOG `[Unreleased]` entry for the godoc change.
3. **README truth pass:** floor claim 1.27.1 → **1.27.2** (line 56 contradicted line 312 — the 10-25 floor wave missed this one); the stale GOEXPERIMENT/jsonv2 paragraph rewritten to the retired-prefix reality (AGENTS 2026-09-29).
4. **TODO_LIST rebuilt per its own charter:** 16 pure-`[x]` rows deleted (dispositions live in module CHANGELOGs/AGENTS/status reports — each audited for a durable home before removal); rows 44/46/51 removed (recorder cliff closed by health v0.3.0 upstream+local; upstream-asks batch filed; consumer-claim ritual superseded by `scripts/check-consumer-claims.sh`); rows 42/43/62 trimmed to live residuals; P1 tip-CI row rewritten with the current fix state + acceptance; watchlist row truthed (go-health#7 CLOSED COMPLETED upstream — adopted+pinned by v0.3.0; go-flightrecorder#2 still OPEN); Open questions consolidated to the **2 live ones** (machine go default — newly harvested from the reports' standing §g ask; cross-repo tracking posture); prompt-crusher merge marked resolved-in-their-repo.
5. **ANNOTATE sweep, all 9 open 2026-06..09 status reports — every numbered item carries a verdict.** 33 verdicts via `annotate.go repair` (TSV spec; `emit-keys` round-trips verified first, dry-run before every apply) + ~60 stale-marker refreshes via exact-match edits. Highlights: otel benchstat re-baseline → **done 2026-10-18**; 6 upstream filings → **4 filed 2026-10-13** + 2 resolved-without-filing; errorpages swap → **DECLINED at source** (v0.1.2 CHANGELOG); browser CSP + govulncheck → **done 2026-10-13**; CI SSH secret → **dissolved by the 2026-10-14 split** (matrix unauthenticated; secret now gates only structure-linter); AGENTS slim-down → **overtaken** (366→126 on 2026-10-12); the "otel polish batch" phantom destination → **every item verified done in code** (Recovery/RecordError benchmarks at `otel/benchmark_test.go:80,88,116`, gzip/retry notes `otel/README.md:263-268`, tail-sampling section :274+, `OTEL_METRIC_EXPORT_INTERVAL` pin test `otlp_test.go:255-261`, span-status assertion `exceptions_test.go:249`, metrics demo note `example/main.go:96`, deploy.md + 5xx-alert example + dashboard pin-drift check 4); prompt-crusher rows → **resolved in THEIR repo** (go.mod clean, renamed `prompt-crusher`, verified on disk today); machine-toolchain question → routed to the standing TODO Open question.
6. **ARCHIVE with full gates:** 9 reports `git mv`'d to `docs/status/archived/` (44 → 53 files) — check-rows BEFORE the move (gate-first), completeness grep over the moved files prints NOTHING, whole-archive gate prints NOTHING (the index's gate line stays true). `§d` status-column tables ledger-exempted (`<!-- check-rows:ledger -->` — exact token), one non-canonical arrow reworded to the `→ routed` form so the gate classifies it ROUTED.
7. **Status index rebuilt:** 9 Current rows removed, 9 archived-rows added, counts updated (53 @ 2026-11-02), re-sweep note, and the **mandatory bulk-archive manifest** (per-file classification + deciding reason — a bulk archive without a manifest is invisible history).
8. **Planning/feedback docs annotated (non-destructively, kept live):** composition-spike verdict + **2026-10-25 UNBLOCKED banner** (httputil v1.7.0 `NewServerListener` closes #41 — verdict below marked historical, refactor tracked in the TODO composition row); cordis doc trigger state (1 of 3 met); gosse/httputil drafts header + Draft 2 annotated (RESOLVED without filing: `ReplayFiltered` exists upstream; `LoggingFromContext` shipped in v1.6.0, wired appkit-side in otel v0.3.0); go-health sentinel draft (FILED as #7 → CLOSED COMPLETED upstream → adopted+pinned by health v0.3.0).
9. **ROADMAP truthed:** #41 line now says SHIPPED (composition + Core TLS unblocked), Updated line bumped; jsonv2-adjacent ask line closed.
10. **health `setStarted` godoc drain-latch cross-link** — the last open item of the 09-24 report (§f36) — shipped + CHANGELOG'd; module build/vet/race/lint green.
11. **Committed `561a476`** (22 files, +212/−183) with a detailed message; pre-commit guards green; working tree clean after.
12. **Corpus classification completed:** papdashboard planning doc SKIP (current — §7 addendum 2026-10-13 is fresh); batteries doc SKIP (canonical spec, triage banner); 3 research HTMLs + samber-do arch-review doc LEAVE (point-in-time references); doc/ vs docs/ consolidation verified done (10-25 wave).

## b) PARTIALLY DONE

1. **Wave acceptance:** everything committed locally, but the P1 acceptance (owner push + session-close ritual on the FINAL pushed commit — every job green except the designed ci-preflight red) awaits the push. Local verification is complete; remote proof is not.
2. **FEATURES freshness:** health/Consumers/core verified this session; the other 9 module sections NOT row-by-row re-verified (no drift found in spot-checks; guards own release-state lines).
3. **TODO_LIST claim audit:** removals were each verified against code/CHANGELOGs, and one KEPT row's stale claim was caught in self-review (row 65, see d-6) — but the remaining kept `[ ]`/`[~]` rows were not each re-verified against code this session.
4. **The docs-health pipeline applied to `2026-0*` only:** the 36 open 2026-10+ reports got index-level truth (their earlier §f lists were harvested by their own passes per the index) but no annotation pass.

## c) NOT STARTED (observed, deliberately untouched)

- The **2026-10+ reports annotate+archive sweep** (36 files) — the natural next docs-health pass.
- **Composition refactor** (`Service` on `httputil.Server`) — unblocked since 2026-10-25, spike re-run + swap not started.
- **Machine go default decision** and the GOTOOLCHAIN-prefix retirement train (USER question, standing since 2026-09-29).
- **Cross-repo tracking posture** decision (USER question).
- **Dependabot wave merge** (10 PRs open; integration leg now needs re-tidy-not-rebase due to the v0.8.1 MVS resolution).
- **Test-helper dedup into testkit** (blocked on the next core train), **guard-coverage symmetry review**, **`security.APIKeyMatches` ghost-API review**, **realtime subscribe-before-flush** (demand-gated), **fr#2 cooldown → frh deletion** (upstream-gated), **pkg.go.dev flightrecorder** (owner-gated last resort), **pma §6 owner decisions** + working-file ignore ask, **go.work leak #1-7 mutator hunt**, **SigNoz live validation** (accepted-risk), **battery waves W3-W5** (demand-gated), **cordis triggers 2-3**, **core v1 exit-criteria final blockers**, **samber/do lazy-healthy draft** (still USER-gated — the only draft with no disposition), **huma.md push** (owner, Deferred Register).

## d) TOTALLY FUCKED UP (own mistakes, no varnish)

1. **The first repair spec skipped F25-F27 — the exact planted-miss class the gate exists for.** I planned those lines, ran emit-keys on only the lines I remembered, and built the TSV from memory. `check-rows` caught 3 CLEAN rows in a struck table immediately; emit-keys + a second spec fixed it. The tool's own 2026-11-01 lesson says build the spec FROM the emitted keys — I half-applied it.
2. **Pipe-gotcha discipline violated twice:** `… | tail -N; echo "exit=$?"` after `check-rows` and after the archive grep captures the PIPE's exit code, not the tool's — the banned 2026-10-12 pattern, typed by me, in the pass that polices docs. Self-caught: the real verdicts came from the tools' own verdict lines ("9 file(s) complete"), never from the echoed numbers — but the pattern should never have been typed.
3. **Ledger-marker format guess cost a cycle:** wrote `<!-- check-rows:ledger — explanation -->`; the matcher requires the EXACT token `<!-- check-rows:ledger -->` (verified by the gate's LEDGER-EXEMPT note appearing only after the sed to the bare token). A decorated comment silently fails the exemption.
4. **Stale-read edit refusals after annotate.go writes:** the repair tool rewrites files, so my follow-up multiedits were refused twice ("modified since read") — 2 view+retry rounds. The annotate-then-edit sequence needs a fresh view between, every time, no exceptions.
5. **Non-unique old_string:** the multi-tag routed marker text appeared on TWO rows (f35 + f40 of the 09-53 report); multiedit applied 9/10 and I had to grep-diagnose the failure and re-apply with `replace_all`. Related: one index edit was reported "failed" while it had actually applied — I burned a verification round re-grepping before trusting the file over the message.
6. **I introduced a carried-over staleness and caught it only in self-review:** TODO row 65's REMAINING list still named "SSE+health combined lifecycle" as open — but `integration/sse_health_drain_test.go` shipped 2026-09-29 (T25c). I rebuilt TODO_LIST from as-found text, audited every REMOVAL against code, and did NOT re-verify every KEPT claim. Fixed on sight during this report's prep (row now says FIVE CLOSED with the evidence). The verify-checklist's job-fitness row ("factual-only VERIFY passes; kept-claims auditing is the gap") is precisely this.
7. **Verification-scope choices were optimistic:** FEATURES (9 sections) and ROADMAP raw-ideas got spot-checks, not full row-by-row passes — declared in the scope line, but d-6 shows what that choice costs.

## e) WHAT WE SHOULD IMPROVE

1. **Emit-keys-first, no memory specs:** every annotate spec must be generated from `emit-keys` output read in full — the map is wrong exactly when you are sure it is right.
2. **Exit codes never from a pipe:** `cmd > file; echo $?` or trust the tool's verdict line; never `| tail; echo $?`.
3. **View-after-annotate as a mechanical step** before any edit to a just-annotated file.
4. **Kept-claims audit during living-doc rebuilds:** removals AND retentions both get a code check; the second half is where d-6 lived.
5. **Uniqueness pre-check for table-marker edits:** `grep -c` the old_string first; identical marker texts recur across table rows — context-qualify or replace_all.
6. **New cross-file gate candidate (for the verify checklist):** every `→ routed — TODO_LIST/ROADMAP/AGENTS …` marker's destination must resolve to a LIVE row/artifact; dead destinations (the "otel polish batch" phantom, the removed consumer-claim row) are silent rot. This session fixed ~8 instances of it by hand; a grep-able check would catch the class.
7. **Same-train release-state updates keep being missed by trains** (v0.3.0 shipped without AGENTS/FEATURES/fixture updates; the CI guards caught it a run later). The pre-tag ritual covers TAG-time, but daemon/dependabot commits bypass hooks — the net is CI-only. A scheduled doc-drift probe (guard script run outside push events) would catch the class before a runner does.
8. **Exact-token guards deserve loud documentation:** the ledger exemption's exact-token requirement should be stated where the exemption is documented (skill-side note; this repo's index README quotes the bare token already — keep it that way).

## f) NEXT — up to 50 things to get done (grounded ONLY in this session; ~30 kept honest, the rest stay empty rather than padded — brainstorm, not commitment)

**P0 — unblock**

1. ~~Fix TODO row 65's stale SSE+health leg~~ done — during this report's prep (FIVE CLOSED + evidence; found via d-6 self-review).
2. **Owner push + tip-CI ritual** on the wave tip (every job green except designed ci-preflight red; `fleet-lint`/`art-dupl-gate`/`example-smoke` stay green; integration + pin-drift flip back green on the pushed fixture fix).
3. **Answer: machine go default ≥1.27.2?** (standing since 2026-09-29; yes → one train retires the per-command `GOTOOLCHAIN` prefix discipline across AGENTS/hooks/docs).
4. **Answer: cross-repo tracking posture** (actively nudge cqrs-htmx items vs report-only).
5. **Merge the open dependabot wave after master green** (10 PRs; NOTE: integration's go-health v0.8.1 is already MVS-resolved in-tree — that PR must re-tidy, not blind-rebase).
6. **Restore the `SSH_PRIVATE_KEY` secret** (owner-only; gates exactly the structure-linter job now).
7. **The 2026-10+ reports docs-health sweep** (36 files; same pipeline: annotate → check-rows → archive → index manifest).
8. **Stand up the routed-destination liveness check** (e6; grep-able guard or verify-checklist leg; born-red proof expected — ~8 dead destinations existed before this pass).

**P1 — close the session's residuals**

9. **Composition refactor:** re-run the spike against httputil v1.7.0's single-serve semantics (listener consumed after a serve ends; `server.listener_consumed`), re-pin `Addr()`/`Running()` + phase-log contract, then swap; Core TLS rides the same seam.
10. **Guard-coverage symmetry review** (pre-tag-only vs CI-only guards; decide the intended matrix; do NOT default-add network-heavy guards to CI).
11. **FEATURES row-by-row re-verification** for the 9 sections only spot-checked this session.
12. **TODO kept-claims audit completion** (the remaining `[ ]`/`[~]` rows each checked against code — d-6's lesson generalized).
13. **ROADMAP raw-ideas row-by-row re-verification.**
14. **DOMAIN_LANGUAGE.md verify pass** (terms vs code; untouched this session).
15. **frh go-health v0.5.1 → v0.8.1 sweep watch** (integration already MVS-resolves v0.8.1; frh's own pin moves on its next natural train — keep the fixture comment true when it does).

**P2 — quality / hygiene**

16. Test-helper dedup into testkit (blocked on next core train).
17. `security.APIKeyMatches` ghost-API review (consumer sweep incl. the PapDashboard security signal).
18. Realtime subscribe-before-flush (demand-gated; sketch ready in the 17-49 report).
19. fr#2 cooldown combinator → delete frh's hand-rolled cooldown when it ships.
20. pkg.go.dev flightrecorder re-check (owner-gated re-tag last resort).
21. go.work leak #1-7 mutator hunt (inotify) or accept-the-cadence ruling.
22. pma dossier §6 decisions walkthrough (owner; memo recommendations ready) + working-file `ignore_patterns` ask.
23. SigNoz live validation reclassification trigger (instance appears → import dashboard + Exceptions proof).
24. i18n `Language()` composition E2E leg (demand-gated; fits the composition-gap row).
25. Composition-gap remainders: security×errorpages, flightrecorder middleware E2E (the other legs closed — see TODO row 65, now FIVE CLOSED).
26. samber/do lazy-healthy draft: file-or-shelve decision (the only draft with no disposition).
27. huma.md push (owner; Deferred Register).
28. Core TLS demand watch (PapDashboard removed its TLS envs — the trigger's consumer signal is currently nil; the TLS train waits on a real ask).
29. Research HTML trio: keep-as-reference ruling or archive-with-banner (the 2026-09-04 one has a placeholder `<title>`).
30. Cross-project lesson → crush-config `references/lessons.md`: exact-token guards fail silently on decorated comments (the ledger-marker class).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF (3)

1. **Push now?** The wave (`561a476`) is committed and locally green; the P1 acceptance needs a push + the tip-CI ritual. Say the word and it goes; otherwise it waits owner-side as usual.
2. **Machine go default ≥1.27.2?** (Standing since 2026-09-29, asked in every report §g since.) YES → I retire the whole `GOTOOLCHAIN=go1.27.2` prefix discipline (AGENTS command blocks, hook notes, README hints) in one train; NO → the prefix stays documented and I stop asking.
3. **Scope of the next docs-health pass:** should the 36 open 2026-10+ reports get the same annotate+archive treatment (one pass, or split by month), and do you want the routed-destination liveness check (f-8) built as a standing script before that pass runs?

---

_Awaiting instructions. Point-in-time snapshot — annotate, don't rewrite, when bringing current._
