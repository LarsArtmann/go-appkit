# Status reports index

Status reports are POINT-IN-TIME snapshots — the living truth lives in
`AGENTS.md` (Release State section) and `TODO_LIST.md`. This index is one
line per report; verdicts are annotated inline in each file.

**This is the CANONICAL status tree** (decided 2026-09-28): the historical
`doc/status/` tree (reports + `archived/` + this index) migrated here with
`git mv`; the `doc/` singular directory keeps planning, feedback, recipes,
research, and library analysis. New reports land HERE.

## Current (unarchived)

Index rebuilt 2026-10-12 (docs-health pass — it had rotted to 6 of 27 rows).

| Report | One-line summary |
| ------ | ---------------- |
| `2026-09-24_16-52_dedup-session-and-toolchain-drift.md` | Dedup episode 1: 2 of 3 clone groups extracted, repo-unbuildable root cause (toolchain drift), partial 1.27.1 unification; annotated by the 2026-09-28 pass |
| `2026-09-28_23-03_dedup-sweep-t1-and-unification-continuation.md` | Dedup episode 2 (`-t 1`): 7 accepted clone groups with rationale, sorted-keys stdlib swaps, testkit.Serve adoption in integration; its go-appkit items executed by the 09-28/29 passes |
| `2026-09-29_00-05_whouses-member-key-via-fix-status.md` | pdg session: severed go.work member-key traversal fixed in who-uses; crm consumer path revealed; cross-repo follow-ups routed (their repo) |
| `2026-09-29_00-38_docs-health-full-pass.md` | Prior full docs-health AUDIT: living-doc truthing, 380 inline verdicts, doc/-vs-docs/ resolution, guard extensions |
| `2026-09-29_01-36_directonly-hint-and-concurrent-session-status.md` | pdg session: hidden-indirect consumer hint shipped + concurrent-session collision survived; go-appkit items (F22-F27) resolved or routed |
| `2026-09-29_06-40_unification-release-trains-and-buildflow-hardening-status.md` | The 1.27.1 unification + core v0.6.0/security v0.2.0/health v0.1.3/frh v0.1.4(+v0.1.5) trains + pre-commit guards + go.work tracked |
| `2026-09-29_09-53_satellite-waves-polish-batch-and-core-v070-status.md` | cqrs v0.6.0 (SQLitePath removal) + pre-tag-checks.sh + T17/T19/T24-T26 polish + four satellite waves + core v0.7.0 mid-train |
| `2026-09-29_11-47_core-v070-train-completion-and-queue-sweep-status.md` | core v0.7.0 tagged/pushed/proxy-PASS + integration DrainWindowProbe migration + CI guards proven parsing committed go.work; BuildFlow full-mode mutation hazard found |
| `2026-09-29_11-59_otel-signoz-support-status.md` | otel v0.2.0 build: OTLP/HTTP both signals, zero-code env setup, exception capture, SigNoz dashboard (live-SigNoz validation still user-infra-gated) |
| `2026-09-29_12-45_queue-sweep-buildflow-verdict-and-extraction-safe-etag-status.md` | Queue sweep: BuildFlow verdict loop (otel mutations judged), go-etag replace dropped via integration excludes, T20 batch 1 (AGENTS 377→342), T21-T32 closed |
| `2026-10-01_07-16_systemd-integration-execution-status.md` | go-daemon/go-aichat research EXECUTED: core `StartHooks` seam + the then-unreleased `/systemd` module landed race-green + wired; release train (core v0.8.0 first) remains |
| `2026-10-01_14-27_cqrs-htmx-pin-bump-and-consumer-claim-reverify.md` | cqrs-htmx root v4.13.2 first seen by the pin-drift guard; integration pin bumped + suite re-verified; consumer claims re-verified |
| `2026-10-06_22-21_cqrs-adoption-closure-execution-e1-e8.md` | cqrs v0.7.0 adoption-closure plan E1–E8 executed (Domain passthrough, health accessors, engine checkpoints, always-close Shutdown) |
| `2026-10-06_22-21_superb-flightrecorder-execution-status.md` | Mid-execution snapshot of the flightrecorder-adoption plan: T1–T4 done incl. the upstream fr v0.2.1 race fix; supersedes its own open questions (answered same day) |
| `2026-10-06_23-23_superb-flightrecorder-execution-closeout.md` | Plan EXECUTED end-to-end: flightrecorder/v0.2.0 + otel/v0.2.0 + upstream fr v0.2.1 tagged/pushed/proxy-verified; integration re-pinned; cooldown ask drafted (filing gated); 50 next items |
| `2026-10-06_23-34_cqrs-v0.7.0-adoption-closure-self-review.md` | Self-review of the cqrs v0.7.0 train (findings F1-F4 closed, residual watch items) |
| `2026-10-07_00-40_cqrs-v0.7.0-hardening-followups-status.md` | cqrs hardening follow-ups: rubric re-run 87/100, timers tripwire, dual-checkpoint-knob decision |
| `2026-10-07_01-20_cqrs-followups-execution-and-charter-guard-first-fire.md` | cqrs follow-ups executed + the workspace-charter guard's first live catch |
| `2026-10-07_01-22_session-final-cqrs-followups-and-charter-guard-live-fire.md` | Final cqrs follow-ups + charter guard observed green on origin; harvested into TODO_LIST 2026-10-07 (ci-preflight, githooks landed) |
| `2026-10-07_10-07_docs-health-audit-live-leak-and-pin-bump.md` | Prior docs-health audit: caught the missed docs v0.3.2 claim (pin-drift guard's first live run) + workspace leak cleanup |
| `2026-10-07_11-00_pma-daemon-root-cause-and-dossier.md` | pma daemon root-cause dossier (watcher/stager/committer; hooks blind to it; owner follow-ups §6) |
| `2026-10-07_11-01_SUPERB-execution-mid-flight-status.md` | Mid-flight snapshot of the 2026-10-07 SUPERB pareto execution plan |
| `2026-10-10_01-43_i18n-research-session-status.md` | i18n module research (fleet sweep N=4; negotiation lives ONCE in httputil `Language()`) |
| `2026-10-10_10-21_i18n-resolution-and-httputil-docdebt-session-status.md` | i18n RESOLVED (appkit module does NOT fire; appkit content lane parked) + httputil doc-debt session |
| `2026-10-12_14-16_issue-25-26-testkit-drain-and-health-durations.md` | Issues #25/#26 verified + fixed in tree (testkit.NewService; NewProbe→NewChecks), unreleased; release trains user-gated — harvested 2026-10-12 |
| `2026-10-12_14-17_code-quality-scan-session-status.md` | Fleet 12-module quality scan: all green after fixes (charter leak #6, integration fixture + lint backlog, testkit flake cap 10s); HTML report in docs/reviews/ — harvested 2026-10-12 |

Execution plans are indexed beside the reports they execute. Active plans:
`docs/planning/2026-10-07_10-11_SUPERB-pareto-execution-plan.md` (partially
executed — see the 10-07/10-12 reports) and
`docs/planning/2026-10-06_16-06_superb-flightrecorder-adoption-plan.md`
(EXECUTED; its closeout report is indexed above). SUPERB v5 was completed and
archived 2026-10-12 (`docs/planning/archived/`). The standing verdict docs in
`doc/planning/` (composition-spike, cordis, papdashboard) stay live as
watchlist references; executed/superseded plans live in `doc/planning/archived/`
and `docs/planning/archived/` with verdict banners.

Everything else is resolved and lives in `archived/` (43 files as of
2026-10-12). New reports land here and migrate to `archived/` after a
docs-health pass resolves their items.

## Archived (43 files as of 2026-10-12)

Recent additions (verdicts inline; the full list lives in `archived/`):

| Report | One-line summary |
| ------ | ---------------- |
| `archived/2026-09-23_16-20_integration-health.html` | Integration audit (CI-secret P1, go 1.27 re-drift, composition watchlist) — banner routes every finding; suite re-proven green 2026-09-28 |
| `archived/2026-09-23_15-58_todo-list-execution-and-security-trio-status.md` | TODO execution train: stale-item purge, config-parity guard, testkit.Shutdown, security trio |
| `archived/2026-09-20_14-05_health-stack-execution-train-status.md` | Health-stack train T01–T27: health v0.1.2 + frh v0.1.3, directive-parity guard, integration E2Es |
| `archived/2026-09-20_11-37_samber-do-health-review-session.md` | samber/do × health review: DO-1..6 clean, F1–F8 all dispositioned (executed by the train) |
| `archived/2026-09-17_19-13_docs-health-full-sweep-annotate-archive-and-living-doc-truth.md` | Prior full docs-health sweep: 12 archived, ~200 inline verdicts |
| `archived/2026-09-28_22-47_whouses-decode-and-consumer-pin-drift-status.md` | who-uses output decode + consumer-pin fix; its go-appkit/verification items executed by the 2026-09-28 pass |
| `../planning/archived/2026-09-29_00-45_SUPERB-v5-unification-trains-and-gated-unlocks.md` | SUPERB v5: unification + release trains — executed same-day in trains; queue closed 09-29; archived 2026-10-12 with zero unrouted items |

Archive gate — no unresolved items may survive the move:

```bash
grep -rLn '~~' docs/status/archived/ --include='*.md'   # must print NOTHING
```

## Annotation standard

- Completed work leaves TODO_LIST (open items only) and lives in the module
  CHANGELOGs; inline `~~strikethrough~~ + verdict` in reports records WHAT
  happened and WHEN, with evidence (tag, test, commit).
- A closed item without evidence is not closed — it is unverified.
- Archived = fully-resolved report, moved with `git mv`, citations updated in
  living docs; the gate above greps for files without any resolution marker.
- Grouped verdicts are blessed for declared brainstorm/routed blocks
  (recorded per the 2026-09-16/17 practice; see the archived READMEs).

## Release-state ownership (decided 2026-09-16)

`AGENTS.md` "Release State" is the SINGLE OWNER of release facts (which tags
exist, what shipped). TODO_LIST's header links there for history only;
status reports record point-in-time claims with their as-of date and never
update them afterward.

## Rituals

Release and verification rituals live in the recipes dir — start at
`doc/recipes/fresh-consumer-proxy-check.md` (the post-push step of every
release ritual).
