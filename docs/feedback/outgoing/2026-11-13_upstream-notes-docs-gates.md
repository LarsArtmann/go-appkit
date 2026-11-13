# Upstream notes: docs-gates — three observations from the go-appkit corpus sweeps

**Status:** DRAFT — filing is USER-gated (target repo: github.com/Artmann-Technologies/docs-gates, own-tools repo; premise verified against the binary's behavior across the 2026-11-02..13 sweeps). These are behavior notes, not bugs — each with a suggested direction.

## Note 1 — routed≠resolved: `→ routed` markers satisfy routing but not freshness

An item carrying `**→ routed — <destination>**` counts as disposed for the
ROUTING gate (no unrouted open rows) but stays OPEN for FRESHNESS (closes
only on `~~` or `<!-- fresh:skip`). Consequence: a routed item whose
destination row is later purged loses its living-line anchor and resurfaces
as UNHARVESTED in `--audit-all` — the strict parser then flags the
latest-per-module list as malformed ("unparseable entries").

Suggested direction: document the dual semantics prominently (the go-appkit
AGENTS corpus now carries a hand-strike recipe: `~~sentence~~ done —
<evidence> **→ routed …**`), or let an anchored routed marker with a
destination that still exists count as resolved-for-freshness.

## Note 2 — strict latest-per-module list format is undocumented

`check-pin-drift.sh`'s claims parser rejects annotated entries in the
AGENTS.md Latest-per-module line: `preset v0.1.1 — <prose>` fails
("unparseable entries (strict list)") while bare `preset v0.1.1` parses.
The accepted grammar (comma-separated bare `name vX.Y.Z` entries, one
line) deserves a line in the failure output — the current message names the
offense, not the accepted shape.

## Note 3 — DF genericity drifts as living docs grow

The DF<8 distinctiveness threshold counts LINES across
TODO_LIST/ROADMAP/CHANGELOG(/AGENTS). Every new living-doc line pushes
historical item tokens toward genericity — a harvest that was valid at
mint time can silently DECAY into accidental-overlap (audit-all then
resurfaces the item with "only generic tokens matched"). `--pre-flight`
catches WOULD-DRIFT for lines about to land; there is no equivalent
post-hoc sweep for items whose anchors DECAYED (we hit one: an item routed
to a TODO row purged in a later wave). Suggested direction: an
`--audit-all --decay-report` mode listing items whose only anchors are
generic (DF>=8) so sweeps can re-anchor them before they resurface as
regressions.

## Provenance

All three observed 2026-11-13 in go-appkit's docs-health full pass (128
reports replayed, 693 living lines; the battery otherwise green 5/5).
