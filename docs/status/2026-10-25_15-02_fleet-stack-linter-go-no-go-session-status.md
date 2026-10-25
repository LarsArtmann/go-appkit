# Session Status: Fleet-Stack Linter Go/No-Go Consultation

**Date:** 2026-10-25, window ending 15:02 CET
**Repo:** go-appkit (session cwd; work is fleet-wide consultation, zero code changes here)
**Scope:** Analysis of whether to build a "fleet-stack linter" (detecting hand-rolled duplicates of LarsArtmann fleet library APIs) for cases like the company-master-data httputil fleet-stack deep-dive audit, mirroring the samber-linter / go-humanize-linter precedent.
**Session type:** Pure consultation. **No files written, no commits made, no code changed.** The only durable input was reading; the only durable output was a chat verdict.

---

## Verdict Delivered (one-line recap)

NOT YET on the linter: the audit's findings decompose into 1 linter-eligible class (hand-rolled duplicate, corpus = 1 known instance), 8-10 non-lintable adoption-judgment rows (4 externally gated), and 1 version gap already covered by dependabot. Trigger to revisit: 3+ rules with real corpus instances, built by copying go-humanize-linter.

---

## a) FULLY DONE

| # | What                                                                                                                                                              | Evidence                                                                                                                                                                                            |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Loaded and applied the `linter-building` skill's decision process (3 questions, mechanism matrix) to the question                                                 | `~/.config/crush/skills/linter-building/SKILL.md`, `references/mechanism-choice.md` rows 3/4 (walker vs go/analysis) read in full                                                                   |
| 2 | Ingested the subject audit: extracted text from the 75 KB / 1,762-line HTML report and analyzed its summary, findings 1-11, version-currency, and Pareto sections | `/home/lars/projects/company-master-data/docs/research/2026-10-25_httputil-fleet-stack-deep-dive.html`; extraction at `/tmp/httputil-deepdive.txt` (ephemeral, 1,138 lines, read through line ~820) |
| 3 | Read both precedent linters' intent, rule shapes, and confidence/exit-code architecture                                                                           | `samber-linter/README.md` (HW-0..HW-8 rules, 0.75 default confidence, triage exit 2), `go-humanize-linter/README.md` (H001-H012 syntactic rules, go-linter-sdk + go-finding stack)                  |
| 4 | Consumer-corpus probe: grepped 4 consumer repos for hand-rolled patterns (charmbracelet `log.NewWithOptions`, `DrainDelay: 0` footgun)                            | cqrs-htmx, PapDashboard, InboxClean: zero hits, zero charmbracelet imports; company-master-data: exactly one hit at `internal/app/injector.go:179` (`log.NewWithOptions(os.Stdout, ...)`)           |
| 5 | Verified the version-gap class is already automated: company-master-data has `.github/dependabot.yml`                                                             | `ls` confirmed 2026-10-25                                                                                                                                                                           |
| 6 | Sized the precedent build cost                                                                                                                                    | go-humanize-linter: 12,297 Go LOC total (incl. tests + generated pattern files)                                                                                                                     |
| 7 | Delivered the go/no-go verdict with decomposition table, FP-budget reasoning, and a named re-entry trigger                                                        | Chat answer, 2026-10-25 (see recap above)                                                                                                                                                           |

---

## b) PARTIALLY DONE

| # | What                                                                  | Works                                                                                                | Missing                                                                                                                                                                                                                                                                  | Blocker                            | Effort |
| - | --------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------- | ------ |
| 1 | Reading the source audit end-to-end                                   | ~72% read (lines 1-820 of 1,138 extracted): summary, all 11 findings, version currency, Pareto table | Appendix: the full 44-row capability inventory (incl. 13 N/A verdicts with reasons) and the "How to Verify" section (lines ~820-1,138) unread                                                                                                                            | None — pure oversight              | S      |
| 2 | Fleet-wide corpus claim ("corpus = 1")                                | Verified across 4 repos × 3 grep patterns                                                            | The fleet has more consumers (go-plugin-mvp, cordis, KeyHolderAI, browser-history, and others per AGENTS cross-project notes) — none probed                                                                                                                              | Scope was self-imposed; no blocker | M      |
| 3 | Effort estimate for a would-be linter                                 | 12,297 LOC total cited                                                                               | Non-test/generated LOC never decomposed, so "how big is a copy" is loose                                                                                                                                                                                                 | None                               | S      |
| 4 | Cross-checking my class decomposition against the audit's own numbers | Finding-level count (1 dup / 1 version / 9 adoption rows) derived from findings 1-11                 | Chat table said "8" adoption rows while the audit headline says **10 missed opportunities** — reconcilable only via finding-vs-capability counting (finding 9 bundles rate-limit + CORS + SSE-cap + systemd; finding 4 bundles 3 dashboard levers), which I never stated | None                               | S      |

---

## c) NOT STARTED

| # | What                                                                                                                                               | Why not started                                                          | Still wanted?                  |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | ------------------------------ |
| 1 | **Durable decision record** of the go/no-go verdict + re-entry trigger                                                                             | Consultation ended at the chat answer; nothing persisted                 | YES — this is the top gap      |
| 2 | Building the fleet-stack linter itself                                                                                                             | Deliberately deferred per verdict (corpus = 1 < trigger of 3+ rules)     | Conditionally — see trigger    |
| 3 | Golden-corpus collection (the skill's "write the golden corpus first" gate): census of duplicate instances per candidate rule across ALL consumers | Not started; prerequisite to any build                                   | YES                            |
| 4 | Wider consumer sweep (go-plugin-mvp, cordis, KeyHolderAI, browser-history, remaining fleet repos)                                                  | Out of the 4-repo probe scope                                            | YES                            |
| 5 | Executing company-master-data audit items 1-6 (ETag, dashboard drain/trend, InitLogger swap, Server-Timing segments, checks.Disk, httputil v1.6.0) | Different repo, different session's lane; this session was advisory only | YES (audit's own Pareto order) |
| 6 | Verifying the audit's 10 items are tracked in company-master-data's TODO_LIST (docs-health HARVEST surface)                                        | Did not check                                                            | Unknown — needs check          |
| 7 | Reading `linter-building/references/ecosystem.md` (which repo to copy per architecture)                                                            | Verdict was "not yet"; needed only if build is triggered                 | On trigger                     |

---

## d) TOTALLY FUCKED UP

Radical honesty: **nothing in any codebase is broken by this session** (it changed no code). What is fucked up is process-level:

| # | What's fucked up                                                                                                                                                                                                                                                                                 | Severity                                                                                                                                   | Root cause                                                                                                                                                                                            | Mitigation                                                    |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| 1 | **The verdict left zero durable artifacts.** A consultation that changes no files and records no decision is ephemeral: a future session will either re-derive the whole analysis (wasted hour) or, worse, skip the FP-budget reasoning and just build the linter off the audit's single finding | High (future wasted/wrong work, not current breakage)                                                                                      | Treated "should we X?" as a pure Q&A turn instead of engineering work with a write-down obligation; AGENTS.md memory rules ("No threshold — if it's new information, write it down") were not applied | Write the decision record (repo = open question, see g1)      |
| 2 | **Precision slip in the delivered verdict:** said "8" adoption rows; audit headline says 10 missed opportunities; my probe framing said "fleet-wide" but covered 4 of ~8+ consumer repos with 3 patterns                                                                                         | Medium — the verdict direction (not yet) survives both corrections, but the stated numbers don't reconcile without unstated counting rules | Counted findings (numbered 1-11) instead of capabilities; generalized from a scoped sample                                                                                                            | Recount from the appendix; restate claims with explicit scope |
| 3 | **Grep hygiene violation of this repo's own gotcha:** piped `rg` output through `grep -v _test` in the probe. The AGENTS.md "Never pipe verify/guard output" rule targets guards specifically, but pipe-swallowing is the flagged failure class                                                  | Low (informal probe, results visually confirmed)                                                                                           | Habit                                                                                                                                                                                                 | Use `rg -g '!*_test.go'` glob filtering instead of pipes      |

---

## e) WHAT WE SHOULD IMPROVE

1. **Consultations must still write a decision record.** Any "should we X?" turn that ends in a verdict with a trigger condition needs a one-paragraph ADR-style note in the repo that owns the question. Impact: prevents re-derivation and contradictory future verdicts. Fix: make it a standing rule in project AGENTS.md or crush-config lessons.
2. **Read the whole source document before verdicts that lean on its appendix.** My class decomposition used the curated summary; the 44-row appendix (with per-row N/A reasons) is where counting disputes get settled. Fix: for any "classify all rows" verdict, read to EOF first.
3. **Scope-qualify corpus claims.** "Fleet-wide corpus = 1" should have been "4 repos × 3 patterns: corpus = 1." Small phrasing discipline, big honesty difference.
4. **Cross-link verdicts to the producing skill's cadence.** The audit was produced by the `library-deep-dive` skill and self-schedules a re-run post-deploy; the linter verdict should have been recorded as an annotation on that cadence (the linter trigger and the audit re-run are the same tripwire).
5. **Effort estimates need decomposition.** "12.3k LOC precedent" mixes tests, generated patterns, and glue; a copy-vs-from-scratch decision needs the non-test number.

---

## f) Top #35 Things We Should Get Done Next

Ranked by impact; each tagged with owning repo. (User asked for up to 50; stopping at 35 — every item here is traceable to this session's evidence, and padding to 50 would violate the "brainstorm, not commitment" rule. Items 19-35 are roadmap-fuel grade.)

### Tier 1 — unblock + record (this week)

| #  | Task                                                                                                                                                                 | Impact   | Effort | Category      | Repo                                 |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- | ------------------------------------ |
| 1  | Write the fleet-stack-linter go/no-go decision record (verdict, FP-budget reasoning, re-entry trigger)                                                               | Critical | S      | Documentation | TBD (see g1)                         |
| 2  | Answer the exposure/deploy question (LAN vs internet + host) — it has been asked 4x per the audit and gates 4 of 10 findings                                         | Critical | ?      | Decision      | user                                 |
| 3  | Verify company-master-data's TODO_LIST tracks audit items 1-11; HARVEST the gaps                                                                                     | High     | S      | Documentation | company-master-data                  |
| 4  | Read the audit appendix (extract lines ~820-1,138) and reconcile the 8-vs-10 class count                                                                             | High     | S      | Quality       | go-appkit (this session's follow-up) |
| 5  | Execute audit item 1: ETag/If-None-Match via go-etag/server (v0.6.1 already transitive) + `httpspec.ExpectNotModifiedWithETag` assertion                             | High     | S      | Feature       | company-master-data                  |
| 6  | Execute audit item 6: bump httputil v1.5.0→v1.6.0 + server_timing v1.0.2→v1.0.3 (Range-aware compression fix; re-paste flake vendorHash)                             | High     | S      | Feature       | company-master-data                  |
| 7  | Execute audit item 2: dashboard `WithShutdownDrain` + `WithTrend`/`WithMetrics`                                                                                      | High     | S      | Feature       | company-master-data                  |
| 8  | Execute audit item 3: replace `newLogger` (injector.go:168-185) with `appkit.InitLogger`; write down stdout-vs-stderr/timestamp deltas as deliberate non-fix if kept | High     | S      | Cleanup       | company-master-data                  |
| 9  | Execute audit item 5: `checks.Disk(cfg.DataDir, 1<<30)` as non-critical probe check                                                                                  | High     | S      | Feature       | company-master-data                  |
| 10 | Execute audit item 4: Server-Timing `lookup`/`list` segments beside `render`                                                                                         | Medium   | S      | Feature       | company-master-data                  |

### Tier 2 — corpus building (only if the linter stays wanted)

| #  | Task                                                                                                                                                                                             | Impact | Effort | Category | Repo     |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | -------- | -------- |
| 11 | Wider consumer census: sweep go-plugin-mvp, cordis, KeyHolderAI, browser-history + remaining fleet repos for hand-rolled duplicate patterns                                                      | High   | M      | Quality  | fleet    |
| 12 | Golden-corpus fixtures for candidate rule `manual-logger-init` (charm import + `log.NewWithOptions` + level parsing → suggest `appkit.InitLogger`)                                               | High   | S      | Quality  | TBD (g1) |
| 13 | Golden-corpus fixtures for candidate rule `manual-error-classification` (`fmt.Errorf` + status mapping → go-error-family constructors)                                                           | High   | S      | Quality  | TBD (g1) |
| 14 | Golden-corpus fixtures for candidate rule `drain-delay-zero` (`DrainDelay: 0` intent → `NoDrainDelay`; near-zero FP, exact sentinel semantics)                                                   | Medium | S      | Quality  | TBD (g1) |
| 15 | State the FP budget per rule BEFORE any detection code (linter-building skill hard gate)                                                                                                         | High   | S      | Quality  | TBD (g1) |
| 16 | If corpus reaches 3+ rules with real instances: scaffold the linter by copying go-humanize-linter (go-linter-sdk `RuleFunc` + go-finding `Finding`, syntactic walker per mechanism-choice row 3) | High   | L      | Feature  | TBD (g1) |
| 17 | Decide distribution fronts up front if built: CLI + golangci-lint module plugin from ONE detector core (go-humanize-linter precedent)                                                            | Medium | S      | Decision | TBD (g1) |
| 18 | Rule-spec sketch: hand-rolled `/health` JSON handlers vs `appkithealth.NewProbe`; hand-rolled SSE sync vs `hub.SubscriberCount()` house rule                                                     | Medium | S      | Quality  | TBD (g1) |

### Tier 3 — roadmap fuel (not commitment)

| #  | Task                                                                                                                                                                                                                                | Impact | Effort | Category      | Repo                |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- | ------------------- |
| 19 | Audit item 9a: rate limiting + CORS + SSE cap as `CMD_` config variants (gated on #2)                                                                                                                                               | High   | M      | Feature       | company-master-data |
| 20 | Audit item 8: errorpages for the HTML datasheet route (decide alongside group-datasheet redesign)                                                                                                                                   | Medium | M      | Feature       | company-master-data |
| 21 | Audit item 10: drain-window e2e via `testkit.DrainWindowProbe` (pairs with #7)                                                                                                                                                      | Medium | M      | Quality       | company-master-data |
| 22 | Audit items 8+11 deploy train: tracing (`OuterMiddlewares` + appkit/otel + `OnPanic`) and systemd module + `StartHooks` sd_notify, in one session when the deploy target exists                                                     | High   | L      | Feature       | company-master-data |
| 23 | Re-run the deep-dive audit post-deploy (the report's own instruction; tripwire shared with the linter trigger)                                                                                                                      | High   | M      | Quality       | company-master-data |
| 24 | Fold go-health-dashboard's unreleased CSV formula-injection fix in at the next dashboard bump (do NOT pin HEAD early)                                                                                                               | Medium | S      | Security      | company-master-data |
| 25 | Pre-write LAN vs internet config variants so the exposure answer flips config, not code (fleet session §e.3, still unimplemented)                                                                                                   | Medium | M      | Feature       | company-master-data |
| 26 | Decompose go-humanize-linter LOC (test vs non-test vs generated) so future copy-cost estimates are honest                                                                                                                           | Low    | S      | Documentation | go-humanize-linter  |
| 27 | Consumer census doc: one page listing every fleet consumer + adoption surface (cqrs-htmx setup, PapDashboard, company-master-data confirmed this session; others unknown) — makes corpus probes one lookup instead of a grep safari | Medium | M      | Documentation | go-appkit           |
| 28 | Consider a `verify-external-claims` pass on this session's numbers (12,297 LOC; adoption 61/100; corpus counts) before they get reused in any doc                                                                                   | Medium | S      | Quality       | fleet               |
| 29 | If linter is built: add `linter-building/references/ecosystem.md` row for it (which-repo-to-copy map is that file's job)                                                                                                            | Low    | S      | Documentation | skills repo         |
| 30 | If linter is built: baseline/ratchet support from day one so fleet consumers adopt incrementally (go-finding baseline machinery)                                                                                                    | Medium | M      | Feature       | TBD (g1)            |
| 31 | If linter is built: wire into BuildFlow via toolsdk self-registration (10 shipped consumers precedent)                                                                                                                              | Medium | M      | Feature       | TBD (g1)            |
| 32 | Annotate the audit HTML with the linter go/no-go outcome (cross-repo write — only with owner approval)                                                                                                                              | Low    | S      | Documentation | company-master-data |
| 33 | Promote "consultations end in a decision record" into crush-config `references/lessons.md` (cross-project lesson, by commit)                                                                                                        | Medium | S      | Documentation | crush-config        |
| 34 | Check whether `library-deep-dive` skill should auto-emit "duplicate-class findings" as a structured section (it found the one lintable class; a stable output shape would feed the future corpus directly)                          | Medium | M      | Feature       | skills repo         |
| 35 | Grep hygiene: replace pipe-filtered probes with `rg -g` globs as standing habit (this session's d3)                                                                                                                                 | Low    | S      | Cleanup       | habit               |

---

## g) Three Questions I Cannot Figure Out Myself

1. **Where does the fleet-stack-linter decision record belong?** It is neither company-master-data's (they are one consumer), nor purely go-appkit's (the linter would cover all fleet libs), nor skills-repo content. I looked for an existing fleet-wide planning/decisions doc this session and found none. Candidates: a new fleet-planning doc, go-appkit's Deferred Register, or the linter-building skill's ecosystem.md. Your call decides item f1's home.
2. **Will the exposure/deploy answer (LAN vs internet, host) ever land, and what would unblock it?** It has been asked four times per the audit, gates 4 of 10 findings plus the deploy train, and I have no access to the deployment intent. Until it lands, a quarter of the audit stays frozen.
3. **What bar do you actually want for building the linter?** I proposed "3+ rules with real corpus instances," but an equally defensible bar is audit-keyed ("build after the next 2 deep-dive audits each surface a new duplicate class") or never (the audit skill may simply be the permanent automation). This is a preference/ROI judgment I cannot derive from code.

---

## Handoff Notes

- **Format override:** skill default is a styled HTML dashboard; user explicitly demanded `.md` at this path — honored per the skill's own override rule, flagged here as required.
- **HARVEST:** section (f) is the input for `docs-health` HARVEST, but most items are owned by OTHER repos (company-master-data, skills repo, user decisions). User instructed WAIT FOR INSTRUCTIONS — so no harvest was run in this session; if this session continues, harvest items 1-18 into the repos named per row first.
- **No commits made** (harness forbids without explicit request); the auto-commit daemon will pick this file up.
- Related prior reports in this series: `docs/status/2026-10-19_07-34_post-fix-wave-status-and-self-review.md` (different scope — go-appkit v9 hardening, not fleet consultation).
