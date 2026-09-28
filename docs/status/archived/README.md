# Archived status reports

Point-in-time snapshots whose every forward-looking item is resolved: shipped
(strikethrough + evidence), explicitly rejected (`Won't implement` /
`NOT-DO`), or now owned by a living doc (`TODO_LIST.md`, `ROADMAP.md`,
`AGENTS.md`). Inline `~~strikethrough~~ done at ...` markers carry the
verdicts — scan for `~~` to see why each item closed. Never delete these;
they are the archaeology layer. New reports land in `docs/status/` and
migrate here only after a docs-health pass resolves their items.

## Annotation standard (recorded 2026-09-16)

Verdicts are recorded INLINE at the item they resolve: `~~strikethrough~~`
plus a one-line verdict carrying WHAT happened, WHEN, and the EVIDENCE (tag,
test name, or commit). Depth rule: one sentence per resolved item — enough
for a future reader to re-verify without re-deriving; no essays. Rows that
were never resolved stay untouched; wrong-but-historical claims get a
correcting note, not a silent rewrite.
