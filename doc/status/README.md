# Status reports index

Status reports are POINT-IN-TIME snapshots — the living truth lives in
`AGENTS.md` (Release State section) and `TODO_LIST.md`. This index is one
line per report; verdicts are annotated inline in each file.

## Current (unarchived)

| Report                                                                             | One-line summary                                                                                                                                                                                                  |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `2026-09-23_16-20_integration-health.html`                                         | Integration audit by execution: E2E suite 14/14 green (-race, hermetic, published pins), pins 8/8 = latest tags, 7/7 module suites green; master CI DEAD (0/100 green — SSH secret empty) + go 1.27 re-drift live |
| `2026-09-23_15-58_todo-list-execution-and-security-trio-status.md`                 | TODO execution train: 10 stale items purged, pin contract → integration/doc.go, config-parity guard, testkit.Shutdown, security trio (threat model + example + security×realtime E2E); F6 re-drift live           |
| `2026-09-20_11-37_samber-do-health-review-session.md`                              | samber/do × health review: DO-1..6 clean, rubric 4.57/5, F1-F8 routed; +2 execution-train addenda (composed-stack proof, benchmark/live E2E re-verification)                                                      |
| `2026-09-17_19-13_docs-health-full-sweep-annotate-archive-and-living-doc-truth.md` | Docs-health full sweep: pkg.go.dev P1 closed, TODO_LIST rebuilt open-only, 12 files archived, ~200 inline verdicts, doc/-vs-docs split brain removed                                                              |

Execution plans are indexed beside the reports they execute:
`doc/planning/2026-09-20_11-47_health-stack-pareto-execution-plan.md`
(health-stack correctness train — T01-T21 shipped 2026-09-20: health v0.1.2,
frh v0.1.3, directive-parity CI guard, integration E2Es, Mounted.Start fix;
T22-T27 same-session remainder).

Everything else is resolved and lives in `archived/` (36 files). New reports
land here and migrate to `archived/` after a docs-health pass resolves their
items.

## Archived (36 files as of 2026-09-17)

Recent additions (verdicts inline; the full list lives in `archived/`):

| Report                                                                         | One-line summary                                                                                                |
| ------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------- |
| `archived/2026-09-15_19-48_otel-telemetry-status.html`                         | OTEL pattern-regression bisection (fix shipped in otel v0.1.1 / httputil v1.2.0; inline CORRECTIONs 2026-09-16) |
| `archived/2026-09-16_07-03_otel-session-self-review.md`                        | OTEL session self-review (fully annotated)                                                                      |
| `archived/2026-09-16_08-40_otel-pattern-propagation-fix-and-self-review.md`    | Pattern-fix session report (train shipped same day)                                                             |
| `archived/2026-09-16_08-53_library-utilization-audit-and-fixes.md`             | Library utilization audit — all findings routed/closed                                                          |
| `archived/2026-09-16_09-38_docs-health-full-repo-audit.md`                     | Full-repo docs-health audit #1                                                                                  |
| `archived/2026-09-16_14-41_docs-health-second-pass-full-repo-audit.md`         | Second-pass audit + health report                                                                               |
| `archived/2026-09-17_05-52_superb-plan-v2-execution-and-honest-gaps.md`        | SUPERB plan v2 execution — all 30 C-tasks, 7 tags                                                               |
| `archived/2026-09-17_08-08_core-v050-train-e2e-closure-and-honest-gaps.md`     | Core v0.5.0 train + F104/composition/MetricsHook E2Es + health examples                                         |
| `archived/2026-09-17_14-22_setup-usage-verification-and-agentsmd-drift-fix.md` | setup-usage verification (direction: setup → core) + AGENTS drift fix                                           |

The companion plan `2026-09-16_15-01_SUPERB-visibility-correctness...` and
the executed wave plan `2026-08-16_12-04-SUPERB-release-wave-and-harvest.html`
live in `doc/planning/archived/` with verdict banners.

Archive gate — no unresolved items may survive the move:

```bash
grep -rLn '~~' doc/status/archived/ --include='*.md'   # must print NOTHING
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

Release and verification rituals live in `../recipes/` — start at
`../recipes/fresh-consumer-proxy-check.md` (the post-push step of every
release ritual).
