## TL;DR

Per-module `.golangci.yml` fleets fail the findings gate out of the box: ≥50-line duplicated config blocks land at error severity, and the only consumer escape — a `.jscpd.json` — silently REPLACES the curated defaults (format restriction + vendor safety net) instead of merging. Ask: merge user config over the defaults, and ignore known config files (or demote config-file clones below error).

## Problem 1: config-file duplication gate-blocks by default

Every multi-module Go repo in this fleet carries one `.golangci.yml` per module (deliberately standalone — golangci-lint has no config inheritance), so the configs share large blocks by design. jscpd flags them like code:

- `tools/providers/jscpd_provider.go:49` (@ 8a349283a): `jscpdIgnorePatterns` covers vendor/node_modules/result/.git/.cache/.direnv and generated-code globs — no config files.
- Line 53: `largeCloneThreshold = 50` escalates to `SeverityError` → the findings gate (`--fail-on=error`) blocks the whole run.

Live (go-appkit, 2026-10-12): 14 error findings, every single one `.golangci.yml`×`.golangci.yml` (57/54-line blocks) or generated dashboard JSON (51-line panel blocks). Zero were code. The fix each repo lands is the same hand-copied `.jscpd.json` ignore list — go-branded-id already ships `**/.golangci.yml` in its ignore for exactly this.

## Problem 2: a user `.jscpd.json` replaces the curated defaults wholesale

`jscpd_provider.go:102-116`: `if FileExists(.jscpd.json) → args += --config` else the full curated flag set (min-lines/min-tokens/format/ignore). jscpd treats `--config` as the entire configuration, so adding one ignore key silently drops the source-format restriction AND the vendor/node_modules safety net. Customizing the tool weakens it — the opposite of what a partial override implies.

## Also odd

With a user config present, the report stats showed `"filesScanned": 0` while clones were found, and findings print as `Duplicated N lines (0 tokens)` for yaml/json — token counting looks broken for non-code formats, which may also be why line-count alone drives severity.

## Ask

1. Merge `.jscpd.json` over the curated defaults (or warn loudly that it replaces).
2. Ship default ignores for standard config files (`**/.golangci.yml` at minimum), or cap config-file clones at warning — duplicated config is not copy-paste debt.
3. Look at the 0-token stats while you're in there.

💘 Generated with Crush
