# Status reports index

Status reports are POINT-IN-TIME snapshots — the living truth lives in
`AGENTS.md` (Release State section) and `TODO_LIST.md`. This index is one
line per report; verdicts marked DONE are annotated inline in each file.

## Current (unarchived)

| Report | One-line summary |
| ------ | ---------------- |
| `2026-09-15_19-48_otel-telemetry-status.html` | OTEL pattern-regression bisection (historical — fix shipped in otel v0.1.1 / httputil v1.2.0) |
| `2026-09-16_07-03_otel-session-self-review.md` | OTEL session self-review (annotated) |
| `2026-09-16_08-40_otel-pattern-propagation-fix-and-self-review.md` | Pattern-fix session report (fix shipped same day) |
| `2026-09-16_08-53_library-utilization-audit-and-fixes.md` | Library utilization audit — findings routed to TODO_LIST, most closed 2026-09-16 |
| `2026-09-16_09-38_docs-health-full-repo-audit.md` | Full-repo docs-health audit (annotated) |
| `2026-09-16_14-41_docs-health-second-pass-full-repo-audit.md` | Second-pass audit + health report (annotated) |

Entries are one line each; fully-resolved reports move to `archived/` with
their verdicts annotated inline (git mv, citations updated).

## Annotation standard

- `[x]` items in TODO_LIST and inline `~~strikethrough~~ + verdict` in
  reports record WHAT happened and WHEN, with evidence (tag, test, commit).
- A closed item without evidence is not closed — it is unverified.
- Archived = fully-resolved report, moved with `git mv`, citations updated in
  living docs; the archive gate greps for unresolved tildes in `archived/`.

## Release-state ownership (decided 2026-09-16)

`AGENTS.md` "Release State" is the SINGLE OWNER of release facts (which tags
exist, what shipped). TODO_LIST's header links here-and-there for history
only; status reports record point-in-time claims with their as-of date and
never update them afterward.
