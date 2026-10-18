# Status Report — OTel Deep-Dive Session (library-deep-dive on go.opentelemetry.io/otel)

**Date:** 2026-10-18 01:23 CEST
**Scope of this report:** THIS SESSION ONLY — the single task "How well do we do on OTEL?",
executed as a `library-deep-dive` (OpenTelemetry Go SDK) and closed with the report
`docs/research/2026-10-18_otel-deep-dive.html` (commit `7f176f9`, not pushed). Point-in-time
snapshot — it goes stale. No other research was done; everything below is from this session's
own run and what it noticed in passing.

**Verification baseline at write time:** no code was written or modified this session (audit +
report only), so there is no build/test verdict to report. The one shipped artifact (the HTML
report) passed its self-consistency gates: 44/44 file:line citations resolve in-range, the score
denominator recomputes to 78 from the appendix rows, both sorted tables verified descending,
HTML tag-balance clean.

---

## a) FULLY DONE

1. **The OTel deep dive, all phases, shipped and committed.** Prior-report overlap pass (3
   research reports + `doc/TELEMETRY.md` + the 2026-09-29 otel status report + the archived
   2026-08-18 otel backlog, read for overlap before auditing); full usage catalog (71 OTel
   import sites; every file of the `otel` module read line-level, plus `cqrs/otelmetrics.go`
   and the integration pin tests); live upstream research (releases, contrib, semconv history,
   SigNoz attribute handling, exact API signatures for every code example in the report);
   45-row capability inventory scored 78/100 (high band: 25 of 37 applicable fully leveraged,
   8 partial, 4 missed, 0 anti-patterns); HTML report written from the kit template by
   copy-and-edit (no line-splice assembly) at
   `docs/research/2026-10-18_otel-deep-dive.html`; committed `7f176f9` with a detailed message.
   Not pushed — push stays user-gated.
2. **Version currency established as PERFECT, not assumed:** otel v1.47.0 + otelhttp v0.72.0
   are the latest stable releases (upstream 2026-10-02, sixteen days before the audit); zero
   gap on both modules. The only newer work (v1.48.0-in-progress) touches nothing this repo
   uses — verified against the deprecated/removed surface list.
3. **The headline finding, diagnosed and evidenced:** the otel README's OTLP-logs non-goal
   (`otel/README.md:252-254`) claims `go.opentelemetry.io/otel/log` "is still experimental" —
   false since 2026-10-02, when v1.47.0 (the version this repo already pins) shipped the first
   stable Logs API+SDK. The documented revisit trigger has half-fired; the module's own doc no
   longer describes the world it pins.
4. **Supersede bookkeeping done properly:** the 2026-10-06 flightrecorder deep dive's "OTel
   bridge: built, tested — and wired to nothing" finding is marked superseded in the new report
   (otel v0.2.0 shipped the wiring the same day: `otel/example/main.go:107`); the 2026-10-11
   sibling audit confirmed otel-free (no contradiction); four items of the 2026-09-29 otel
   status report §f verified DONE during the audit (metric-interval pin test, gzip/retry docs,
   tail-sampling recipe, status-description integration pin) and the still-open ones re-ranked.
5. **Report self-consistency gates ran and caught a real defect before ship:** one wrong
   citation (`dashboards/README.md:22-26` → actually 21-26) was found by the content
   spot-check and fixed in the shipped version. 44/44 citations resolve; score denominator
   (45 rows, 8 N/A, 37 applicable) re-derived from the appendix and matches every number in
   the prose.
6. **Owner gates respected:** no product docs edited, no TODO_LIST harvest executed, nothing
   pushed — the session was an audit, and the audit stayed an audit. House cleanup rules held
   (temp body file removed via trash/mv, never `rm`).

## b) PARTIALLY DONE

1. **External-claim verification — single-sourced.** Every upstream claim (v1.47.0-is-latest,
   logs-stable, semconv rename history, SigNoz's default-off feature flag, logr/otelslog/sdk-log
   signatures) was established via one agentic-fetch synthesis per topic, with source URLs
   listed in the report, but I never fetched a primary source directly myself. The
   `verify-external-claims` discipline ("confirm the tool/API/release before encoding") is met
   in letter (sources named, APIs cross-checked against tagged sources by the fetch agent) but
   not in spirit (no second leg). The load-bearing claims — the logs-stable headline and the
   currency verdict — deserve a direct fetch of the releases page as confirmation. Listed as
   next-task #15.
2. **Self-verification gates — ran, caught one bug, but needed two re-runs themselves** (see
   d1/d2). The final gate state is trustworthy; the path there was sloppy.
3. **Session-close ritual (AGENTS gotcha, 2026-10-13):** CI state of the final commit was NOT
   verified. Mitigating fact noticed in passing: `ci.yml` ignores md/docs paths, so
   `7f176f9` (docs/ HTML only) should not even trigger the matrix — but "should" is not
   "verified." Next-task #17.
4. **Cross-skill phase of the deep-dive skill (Phase 6):** done implicitly (docs-health harvest
   debt named, deduplicate/data-model skills considered not applicable), but never stated
   explicitly in-session. Cosmetic.

## c) NOT STARTED (correctly queued, nothing built this session)

1. **All implementation follow-ups from the audit** — they were findings, not work: the
   `otel.SetErrorHandler`/`SetLogger` wiring, `appkitotel.Transport()`, `WithOTLPLogs()`,
   `resource.WithProcess()`, the exporter-selector off-switch, the semconv/exemplars docs.
   Nothing was coded — correct for an audit session, and the top items are owner-gated anyway.
2. **TODO_LIST/ROADMAP harvest of the report's priority table** — deliberately not executed
   ("wait for instructions"), but the debt is real: the 12-item table currently lives only in
   the HTML report. Next-task #13.
3. **The otel README non-goal truth-fix** — diagnosed with certainty, queued instead of fixed
   on sight. See d6 and question Q1 for why this is genuinely two-sided.
4. **Consumer-side adoption audit** — the audit covered THIS repo's OTel usage; how
   consumers (cqrs-htmx `setup`, PapDashboard) adopt `appkitotel` was out of scope and remains
   unaudited. Noticed while reading the consumer-claims guard docs; not researched further
   per this session's constraints. Next-task #31.
5. **SigNoz live validation (dashboard import + Exceptions-view E2E)** — user-gated, carried
   open since 2026-09-29 (#4/#5), re-confirmed open. Question Q2.

## d) TOTALLY FUCKED UP

1. **My verification tooling was itself buggy — twice.** First gate run: executed from
   `docs/research/`, so all 44 repo-root-relative citations "failed" (wrong cwd, wasted run).
   Second: the inventory-status regex used `\w+`, which cannot match `N/A` (contains `/`), so
   the denominator gate reported N/A=0 and "passed" for a partially wrong reason — the 37
   matched rows happened to be exactly the applicable set, so the recomputed score was right
   by luck as much as by construction. Gates exist to catch MY errors; these needed two
   re-runs to become trustworthy.
2. **One wrong citation shipped into the first write** (`dashboards/README.md:22-26`, true
   range 21-26). Caught by my own content spot-check and fixed — but it only proves the
   spot-check earned its keep, not that the writing was careful.
3. **Single-sourced the two load-bearing upstream claims** (see b1). If the fetch agent's
   synthesis of "v1.47.0 = first stable logs release" is wrong, the report's headline finding
   is wrong. Probability low (tagged-source citations were returned), but the honesty bar here
   is "verify before encoding," and I encoded on one leg.
4. **Daemon interleaving polluted history around my work, and I did not fully watch it.**
   The auto-commit daemon's `a584b6f` (01:19) swept in: my 924-line TEMP body file
   (`.otel-deep-dive-body.html` — now committed-then-deleted, permanently in history), plus
   changes I did NOT author and did NOT read: AGENTS.md (+2 lines), go.sum (−13 lines),
   integration/go.sum (−76 lines). I correctly did not revert them (never touch what you
   didn't author) but I ALSO never read them — "READ it, judge it" applies to noticing, not
   just reverting. Knock-on effect: my commit `7f176f9` shows as "2 insertions, 926 deletions"
   while its message says "add report" — the message misdescribes the diff for future readers
   because the daemon pre-committed the bulk at 01:19. Audit of `a584b6f`'s foreign hunks is
   next-task #16.
5. **Edit-tool collision after the python splice** (`file modified since read`) — one wasted
   round trip re-viewing before the title edit. Self-inflicted: any out-of-band file mutation
   must be followed by a re-read before `edit`. Cheap lesson, real cost.
6. **The stale README non-goal got queued instead of fixed.** The 2026-09-06 standing
   permission says trivial, already-understood doc staleness is fixed ON SIGHT, even during
   report-then-wait commands — explicitly to avoid "two-line fixes aging in a backlog." I had
   a certainty-grade diagnosis and queued it as priority #3 in a report instead. My
   justification (it touches a RELEASED module's README and interacts with an owner-decided
   posture — the wait-for-consumer gate — so it is a rewrite, not a deletion) is defensible,
   but it is exactly the rationalization the standing permission was written against. Q1 asks
   you to settle the policy.

## e) WHAT WE SHOULD IMPROVE

1. **Write verification scripts once, correctly, against the real data.** Check cwd before
   path-sensitive checks; design regexes against the actual tokens in the data (`N/A` has a
   slash); sanity-check that a "pass" is passing for the RIGHT reason (assert the N/A count
   explicitly, don't let it be silently zero).
2. **Double-source every claim that becomes a headline.** Agentic fetch is research, not
   verification. When a finding will drive a doc rewrite or a priority table, fetch the
   primary source (releases page, tagged source file) directly before encoding it.
3. **After ANY out-of-band file mutation (python/bash), immediately re-view before `edit`.**
   The tool contract is read-then-edit; splicing violates it silently.
4. **Reconcile daemon commits before closing a session:** read every hunk that landed during
   the session — including the ones you didn't author — so history pollution and foreign
   changes get classified the day they happen, not archaeologically later.
5. **End every audit by naming its harvest debt explicitly** (which sections belong in
   TODO_LIST/ROADMAP, which are snapshot-only), so "wait for instructions" doesn't quietly
   become "entomb the findings."
6. **Fix-on-sight policy needs one sharpened edge (Q1):** truth-restoring two-line fixes in
   RELEASED module docs — in-session under the standing permission, or report-first when the
   stale sentence also encodes an owner posture?

## f) TOP THINGS TO GET DONE NEXT (ranked; up to 50 per instruction — 34 real ones, no filler)

_Impact × Ease, 1–5 each. Items 1–12 are the audit's own queue; 13–20 are session-hygiene
debts from THIS run; 21–34 are pre-existing items re-confirmed open in passing (kept terse to
avoid double-tracking — TODO_LIST owns the wording). This section is the HARVEST feed; not
harvested (awaiting instructions)._

| #  | Task                                                                                                                                                                    | Impact | Ease | Priority |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ---- | -------- |
| 1  | Wire `otel.SetErrorHandler` + `otel.SetLogger` into `Setup` (exporter failures become grep-able ERROR lines; add TELEMETRY.md §1 line + pin test; logr already in tree) | 4      | 5    | 20       |
| 2  | Export `appkitotel.Transport()` client-span helper + "Outbound calls" README section (ROADMAP-tracked)                                                                  | 4      | 4    | 16       |
| 3  | Truth-fix the otel README OTLP-logs non-goal (stable-since-v1.47.0; demand-gate unchanged) — pending Q1 posture                                                         | 3      | 5    | 15       |
| 4  | Add `resource.WithProcess()` (or targeted `WithProcessRuntime*`) to `buildResource` + test assertion (mind `process.command_args` sensitivity)                          | 3      | 5    | 15       |
| 5  | Build `WithOTLPLogs()` + otelslog recipe when a consumer asks (pending Q3); prep the design sketch so the answer is hours, not days                                     | 5      | 3    | 15       |
| 6  | Honor `OTEL_TRACES_EXPORTER`/`OTEL_METRICS_EXPORTER=none` as env off-switch in `applyOTLPWiring`; document HTTP-only protocol divergence                                | 3      | 3    | 9        |
| 7  | Record the semconv v1.26→v1.43 migration trigger (SigNoz semconv-families flag watch; `deployment.environment` → `.name`)                                               | 2      | 4    | 8        |
| 8  | Decide + document exemplars and cardinality posture in otel README (two sentences; bridge/cqrs `context.Background()` opt-out by design)                                | 2      | 4    | 8        |
| 9  | Baggage correlation-ID helpers (ROADMAP; README propagation-row note in the meantime)                                                                                   | 3      | 2    | 6        |
| 10 | Composition-gap E2E #1: full-stack panic→exception through a real Service (integration module)                                                                          | 3      | 2    | 6        |
| 11 | Validate dashboard JSON against a live SigNoz + Exceptions-view E2E (pending Q2; open since 2026-09-29 #4/#5)                                                           | 2      | 2    | 4        |
| 12 | Re-baseline middleware ns/op bench rows in a CPU-pinned quiet window (twice deferred; B/op+allocs already ±0%)                                                          | 1      | 2    | 2        |
| 13 | HARVEST this report §f + the deep-dive priority table into TODO_LIST/ROADMAP (docs-health) — closes the entombment created today                                        | 3      | 4    | 12       |
| 14 | Create `docs/research/README.md` index (4 rows: 2026-10-06 ×2, 2026-10-11, 2026-10-18) — the series has no index; discoverability is word-of-mouth                      | 2      | 5    | 10       |
| 15 | Double-source the deep-dive's two headline claims (fetch otel releases page directly; logs-stable + currency) and annotate the report if anything shifts                | 3      | 5    | 15       |
| 16 | Audit daemon commit `a584b6f`'s foreign hunks: AGENTS.md +2, root go.sum −13, integration go.sum −76 — read, classify, confirm intentional                              | 2      | 5    | 10       |
| 17 | Verify CI state of `7f176f9` (expected: not triggered, docs-path-ignored — confirm, don't assume)                                                                       | 1      | 5    | 5        |
| 18 | Render-check the deep-dive HTML in a headless browser (parse-check ≠ render-check)                                                                                      | 1      | 4    | 4        |
| 19 | Fold the two unscored ROADMAP telemetry items (errorpages `trace_id`, fr-snapshot→span link) into the inventory at next revision                                        | 1      | 4    | 4        |
| 20 | Propose a scripts/ research-report checker (citations + denominator) only if the deep-dive series keeps growing — otherwise skip, hand-ran gates suffice                | 1      | 3    | 3        |
| 21 | Composition-gap E2E #2: otel×realtime (SSE spans/metrics on the NoTimeout path)                                                                                         | 2      | 2    | 4        |
| 22 | Composition-gap E2E #3: otel×health-dashboard (traced vs filtered routes)                                                                                               | 2      | 2    | 4        |
| 23 | Composition-gap E2E #4: errorpages×otel (pretty 500s + exception events together)                                                                                       | 1      | 2    | 2        |
| 24 | SigNoz alert example (5xx error-rate alert JSON) next to the dashboard                                                                                                  | 1      | 3    | 3        |
| 25 | Multi-env deploy doc (k8s/Compose snippet with the OTEL_* variables)                                                                                                    | 1      | 3    | 3        |
| 26 | Exceptions-oriented dashboard panel once SigNoz builder supports error-index queries cleanly                                                                            | 1      | 3    | 3        |
| 27 | Example: metrics-visible-in-SigNoz demo note (counter via Provider meter)                                                                                               | 1      | 4    | 4        |
| 28 | `check-pin-drift.sh`: add the dashboard JSON to checked artifacts (name/path guard)                                                                                     | 1      | 4    | 4        |
| 29 | Document `WithMessageEvents` non-adoption decision in otel README (evaluated+declined 2026-08-18; currently only in an archived report)                                 | 1      | 5    | 5        |
| 30 | When upstream v1.48.0 ships: ride the train; note the retry-bound-by-WithTimeout improvement for `WithOTLPTimeout` semantics in the CHANGELOG                           | 2      | 3    | 6        |
| 31 | Consumer-side OTel adoption audit (cqrs-htmx `setup`, PapDashboard) — does the fleet's reference consumer actually use `appkitotel`, and how                            | 2      | 3    | 6        |
| 32 | Watch SigNoz semconv-families flag default; when it flips, execute #7's rename                                                                                          | 1      | 2    | 2        |
| 33 | `OTEL_METRIC_EXPORT_INTERVAL`-style env-native reader timeout note (PeriodicReader.WithTimeout) if ever needed — YAGNI watch only                                       | 1      | 2    | 2        |
| 34 | Retire the temp-file-in-git class: teach the daemon's path filter (or pre-commit check) to ignore `docs/**/.*-body.html` working files                                  | 1      | 3    | 3        |

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Fix-on-sight vs report-first for truth-restoring fixes in RELEASED module docs.** The
   2026-09-06 standing permission covers trivial two-line doc staleness ("do NOT queue them
   into the next backlog"). The otel README non-goal I found is exactly that class — but it
   sits in a released module's README and its rewrite touches an owner-decided posture (the
   wait-for-consumer gate on OTLP logs). In-session fix + CHANGELOG `[Unreleased]` note, or
   keep report-first? (Today I chose report-first; #3 and #15 above are queued either way.)
2. **Is there a live SigNoz instance (and which version) I can point at?** The dashboard
   import validation and the Exceptions-view E2E have been user-gated since 2026-09-29 and are
   the only ways to close the report's two "unvalidated-against-live" flags (#11). If none
   exists and none is planned, I will reclassify them from "blocked" to "accepted-risk,
   documented."
3. **Should the consumer-demand half of the OTLP-logs trigger be considered MET now?** The
   logs-signal build (#5) is gated on "a consumer asks." I cannot see cqrs-htmx setup's,
   PapDashboard's, KeyHolderAI's, or InboxClean's roadmaps. If any of them would take
   trace-correlated log ingestion today, the gate is met and #5 jumps the queue; if not, it
   stays parked with the design sketch as prep.

---

_Point-in-time snapshot. Section (f) is the HARVEST feed for TODO_LIST/ROADMAP — NOT harvested;
awaiting your go. Nothing pushed; `7f176f9` is local on master (the daemon's `11b925e` landed
after it and is not this session's work)._

---

## Appendix: T09/T02 verdicts (SUPERB v8 execution, 2026-10-18)

- **`a584b6f` AGENTS.md +2:** legitimate — the two Gotchas entries (shared-cache contention
  masquerading as dependency-bump compile errors; module-inventory `stale-dependabot-module`
  deliberate non-fix) are dated, sourced memory updates authored by the concurrent BuildFlow
  triage session. Kept; no action.
- **`a584b6f` go.sum −13 / integration go.sum −76:** deleted lines are `/go.mod`-suffix
  metadata hashes only, for modules outside the require graphs — the hermetic-`go mod tidy`
  CLEANUP leg that the 2026-10-18 AGENTS Gotcha prescribes after fossilized junk additions.
  Legitimate; no action.
- **CI state of `7f176f9`:** no workflow run exists (expected: CI ignores md/docs paths —
  confirmed via `gh run list --commit 7f176f9` empty).
- **Latest master runs red by design:** all build/test/proxy/lint jobs green; the only X is
  `ci-preflight` (SSH-secret watchdog, designed red) and `structure-linter` skip (rides
  preflight until the secret exists). Standing documented state, not a regression.
- **go.work.sum residue:** stayed as committed (additive tooling-hash lines, harmless — plan
  §0 tree note stands; no action).
- **Foreign untracked status file** `docs/status/2026-10-18_01-45_buildflow-full-run-triage-and-remediation.md`:
  belongs to the concurrent triage session; left untouched (its own session owns it).
- **T02 headline double-source (executed same day):** fetched the otel-go and contrib release
  pages directly — **v1.47.0 is the latest stable core release (2026-10-02, "first stable
  release of the Logs API and SDK")** and contrib `v1.47.0/v2.6.0/v0.72.0/…` pairs with it.
  Both report headlines CONFIRMED against primary sources; no report drift found.
