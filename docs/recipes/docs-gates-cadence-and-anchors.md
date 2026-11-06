# Recipe: docs-gates cadence — thresholds, overrides, and anchor hygiene

**What this covers:** the `cadence` gate's thresholds (`.docs-gates.json`,
NEVER edited casually), when and how to override them deliberately, and the
`TODO row N` anchor-hygiene verdict (2026-11-06 sweep).

## The thresholds (as committed)

```json
"cadence": { "max_unarchived": 15, "oldest_days": 29, "struck_rows": 20 }
```

| Threshold            | Meaning                                         | Why this value                                                                                                                                                                                                              |
| -------------------- | ----------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `max_unarchived: 15` | at most 15 unarchived reports in `docs/status/` | the corpus is the OPEN work ledger; more than ~15 concurrent reports means sessions are writing reports nobody adjudicates (the 2026-11-06 treadmill finding — concurrent sessions refilled a 12-report drain within hours) |
| `oldest_days: 29`    | no unarchived report older than 29 days         | a month is the honest ceiling for "open"; anything older is either archived or its items routed to TODO rows                                                                                                                |
| `struck_rows: 20`    | at most 20 struck DONE rows in TODO_LIST.md     | struck rows are purged to CHANGELOGs/reports on the docs-health pass; 20 un-purged strikes means the purge ritual is behind                                                                                                 |

## Override rationale (the deliberate red)

The gates are designed to red for REAL reasons and get fixed at the cause —
never rethresholded to green a state you don't like. The one legitimate
override is a DELIBERATE cadence change, recorded in the same commit:

1. **`max_unarchived` during a planned bulk-archive wave** — a wave that
   lands 16+ new session reports between gate runs (the 2026-11-06 wave-3
   pattern) can transiently exceed 15. Do NOT raise the threshold: run the
   wave and let the final state satisfy the gate. Raise it ONLY if the
   corpus is being intentionally re-scoped (e.g. a new report CLASS is
   added that lives unarchived by design) — and record the new steady-state
   rationale in the commit message plus here.
2. **`oldest_days` extension** — same shape: only for a report that is
   DELIBERATELY kept open past a month with an owner (e.g. a long-running
   program log that is the live work ledger). The honest alternative is
   archiving with a `Live remainder:` banner and TODO rows — which is why
   this override has never been needed.
3. **`struck_rows` purge pressure** — if 20 strikes accumulate, the fix is
   running the docs-health purge (dispositions → module CHANGELOGs +
   status reports), not raising the ceiling.

Mechanics if an override is ever approved: edit `.docs-gates.json` in the
same commit as the rationale, run `docs-gates run .` to see the new state,
and note the change in `docs/CHANGELOG.md`. A threshold edit without all
three is the failure mode this recipe exists to prevent.

## `TODO row N` anchor hygiene (F-80 sweep verdict, 2026-11-06)

`TODO row N` references are FRAGILE: TODO_LIST rows renumber on every purge
(the charter removes completed rows), so a row number is a pointer to a
moving target. The sweep rule:

- **Living docs** (TODO_LIST, ROADMAP, AGENTS, READMEs, CHANGELOGs): ZERO
  `TODO row N` anchors — swept 2026-11-06, none found. Keep it that way;
  reference rows by CONTENT ("the floor-train row"), never by number.
- **Point-in-time docs** (docs/planning, docs/review, docs/status): row
  numbers describe the document's own as-of snapshot and STAY (the same
  rule as tag-pinned integration test comments). Renumbering them would be
  rewriting history.

The existing planning/review/status hits (2026-11-02 plan F89, the
integration-burden audit's Errata 2, the archived reports) are all
point-in-time and correct as written.
