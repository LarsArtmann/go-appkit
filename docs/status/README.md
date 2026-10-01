# Status reports index

Status reports are POINT-IN-TIME snapshots — the living truth lives in
`AGENTS.md` (Release State section) and `TODO_LIST.md`. This index is one
line per report; verdicts are annotated inline in each file.

**This is the CANONICAL status tree** (decided 2026-09-28): the historical
`doc/status/` tree (reports + `archived/` + this index) migrated here with
`git mv`; the `doc/` singular directory keeps planning, feedback, recipes,
research, and library analysis. New reports land HERE.

## Current (unarchived)

| Report                                                            | One-line summary                                                                                                                                                                       |
| ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `2026-09-24_16-52_dedup-session-and-toolchain-drift.md`           | Dedup episode 1: 2 of 3 clone groups extracted, repo-unbuildable root cause (toolchain drift), partial 1.27.1 unification; annotated by the 2026-09-28 pass                            |
| `2026-09-28_23-03_dedup-sweep-t1-and-unification-continuation.md` | Dedup episode 2 (`-t 1`): 7 accepted clone groups with rationale, sorted-keys stdlib swaps, testkit.Serve adoption in integration; its go-appkit items executed by the 09-28/29 passes |
| `2026-09-29_00-05_whouses-member-key-via-fix-status.md`           | pdg session: severed go.work member-key traversal fixed in who-uses; crm consumer path revealed; cross-repo follow-ups routed                                                          |
| `2026-10-01_07-16_systemd-integration-execution-status.md`        | go-daemon/go-aichat research EXECUTED: core `StartHooks` seam + the unreleased `/systemd` module landed race-green + wired; release train (core v0.8.0 first) remains |

Execution plans are indexed beside the reports they execute — the ACTIVE
plan is `docs/planning/2026-09-29_00-45_SUPERB-v5-unification-trains-and-gated-unlocks.md`
(location per explicit instruction 2026-09-29); the standing verdict docs in
`doc/planning/` (composition-spike, cordis, papdashboard); executed/superseded
plans live in `doc/planning/archived/` with verdict banners (the health-stack
plan and SUPERB v3/v4 joined them on 2026-09-28).

Everything else is resolved and lives in `archived/` (42 files as of
2026-09-28). New reports land here and migrate to `archived/` after a
docs-health pass resolves their items.

## Archived (42 files as of 2026-09-28)

Recent additions (verdicts inline; the full list lives in `archived/`):

| Report                                                                                      | One-line summary                                                                                                                          |
| ------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `archived/2026-09-23_16-20_integration-health.html`                                         | Integration audit (CI-secret P1, go 1.27 re-drift, composition watchlist) — banner routes every finding; suite re-proven green 2026-09-28 |
| `archived/2026-09-23_15-58_todo-list-execution-and-security-trio-status.md`                 | TODO execution train: stale-item purge, config-parity guard, testkit.Shutdown, security trio                                              |
| `archived/2026-09-20_14-05_health-stack-execution-train-status.md`                          | Health-stack train T01–T27: health v0.1.2 + frh v0.1.3, directive-parity guard, integration E2Es                                          |
| `archived/2026-09-20_11-37_samber-do-health-review-session.md`                              | samber/do × health review: DO-1..6 clean, F1–F8 all dispositioned (executed by the train)                                                 |
| `archived/2026-09-17_19-13_docs-health-full-sweep-annotate-archive-and-living-doc-truth.md` | Prior full docs-health sweep: 12 archived, ~200 inline verdicts                                                                           |
| `archived/2026-09-28_22-47_whouses-decode-and-consumer-pin-drift-status.md`                 | who-uses output decode + consumer-pin fix; its go-appkit/verification items executed by the 2026-09-28 pass                               |

The companion plans (15-01, v3, v4, the 08-16 wave, the 11-47 health plan)
live in `doc/planning/archived/` with verdict banners.

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
