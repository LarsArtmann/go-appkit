## TL;DR

`go-work-sync` treats any on-disk module missing from go.work as drift and re-adds it, and gomod-checker's work checks flag the same gap forever — both ignore the `go_mod_ignore_dirs` config that #18 wired, because that wiring stopped at the go.mod checker. Ask: honor ignore dirs in `DetectGoWorkDrift`, `SyncGoWorkFile`, and `work-use-missing-module`.

## Problem

go-appkit keeps `./integration` out of go.work deliberately: the module pins published tags only (charter in `integration/doc.go`), workspace membership would resolve its pins to local code, and it merges its requires into the repo aggregate node — which created a false `cqrs-htmx → go-appkit` cycle in dependency graphs. A dynamic guard (`scripts/check-workspace-charter.sh`) enforces the exact membership in both directions.

BuildFlow keeps undoing that decision:

- 2026-10-12, live run: `WARN go.work is missing module "integration" (present on disk)` → `go-work-sync 1 fixed, rewrote: go.work` → the auto-commit daemon landed the mutation 62 seconds later (pma runs with hooks disabled, so no guard saw it). This is leak #8 in that repo; leaks #1–7 came from other workspace-sync tooling.
- gomod-check has flagged `work-use-missing-module` (info) for the same directory on every run — permanent noise with no consumer-side fix.

## Evidence (BuildFlow @ 8a349283a)

- `tools/providers/go_work_sync.go:73-79` — `detect` reports every `MissingFromWork` entry; `repair` (line 114) calls `SyncGoWorkFile`, which re-adds them. Neither the `dtool.Tool` spec nor the config carries any exclusion input.
- `tools/gomod/module.go:106-170` — `DetectGoWorkDrift` is a set difference between `FindAllGoModDirs()` and the `use` directives; `FindAllGoModDirs` (line 20) is a raw filesystem walk that consults nothing.
- `tools/providers/sdk_checkers.go:168` — #18's fix wires `cfg.GoModIgnoreDirs` into `gomodchecker.NewCheckerWithIgnoreDirs`. But line 172 builds the work checker with `gomodchecker.NewWorkChecker()`, which takes no ignore dirs (`modules/gomod-checker/work_checker.go:17`), and `work_checker_extended.go:101` emits `work-use-missing-module` for any `**/go.mod` not in `use`.
- The `ignore`-directive machinery exists (`Checker.IgnoreDirs`, `NewCheckerWithIgnoreDirs`, the ignore-directive checks) — it just never reaches the two surfaces that produce the use-drift findings.

Precedent: #29 added a snapshot/restore guard because sync silently dropped load-bearing **replaces**. Deliberate **use** exclusions have no equivalent protection — sync overwrites them just as silently. #18's fix wired ignore dirs; it stopped one checker short.

## Impact

Any multi-module repo with a deliberate exclusion (pin charters, generated code, example workspaces) gets silently mutated and re-committed go.work files plus a permanent finding. The only consumer-side workaround today is `skip_steps: [go-work-sync]` — which also forfeits legit drift sync in the other direction (stale use entries).

## Ask

Plumb `cfg.GoModIgnoreDirs` into `NewWorkChecker` and into `DetectGoWorkDrift`/`SyncGoWorkFile` (skip ignored dirs during discovery, and never re-add them), or honor go.work's own `ignore` directive in those paths. Either mechanism turns the recurring re-add into a config declaration instead of a skip-the-tool tradeoff.

💘 Generated with Crush
