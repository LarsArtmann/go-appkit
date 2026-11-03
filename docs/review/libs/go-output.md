# go-output — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). First dedicated
> report.

## Verdict

**NOT USED — and there is no adoption path to record: a category mismatch. Zero imports,
zero graph presence across all 12 modules. go-output solves "write your data once, render it
anywhere" for CLI tools and report generators (16 output formats, tree/graph/D2/PlantUML
diagrams, NOM-style terminal progress). go-appkit is an HTTP service framework: its one
data-emission surface (the docs module) serves machine-readable catalogs over HTTP, not
terminal tables. No battery, TODO row, or roadmap item names output rendering. The library's
consumers live in the family's CLI-shaped projects (cmdguard-built tools), not here.**

| Fact           | Value                                                                       | Evidence                                                       |
| -------------- | --------------------------------------------------------------------------- | -------------------------------------------------------------- |
| Direct imports | 0                                                                           | `rg 'go-output' --glob '*.go'` → empty                         |
| In graph?      | No                                                                          | absent from all go.mod/go.sum                                  |
| Latest tag     | v0.38.7                                                                     | `git -C ~/projects/go-output tag --sort=-v:refname \| head -1` |
| Shape          | Root module (zero heavy deps) + 18 submodules, each independently versioned | `~/projects/go-output/go.mod`, 19 go.mods total                |
| Version traps  | v0.32.1 / v0.33.0 / v0.34.0 RETRACTED (bogus/stale tags)                    | `~/projects/go-output/go.mod:21-28`                            |
| Upstream state | HEAD 6 commits past v0.38.7 ([Unreleased] empty)                            | `~/projects/go-output/CHANGELOG.md`                            |

## 1. Capability inventory

- **Core model**: `Table`/`TreeNode`/`Graph` builders → immutable snapshots → pure
  renderers; `Format` enum of 16 (`Table JSON CSV TSV Markdown XML D2 YAML HTML Tree
  Mermaid DOT JSONL AsciiDoc TOML PlantUML`) with `Format.Supports(Shape)` capability
  matrix; `RenderTable(data, format, opts)` dispatch with runtime format choice.
- **Format submodules**: delimited (CSV/TSV streaming), serialization (JSON/YAML/TOML/JSONL),
  markdown, markup (XML/HTML/AsciiDoc + streaming HTML), tree, graph (DOT/Mermaid), d2,
  plantuml, daghtml (interactive HTML DAG with pan/zoom).
- **Progress/TUI**: `nom` (NOM-style inline terminal renderer with phase graphs, critical
  path, timing estimates) and `tui` (bubbletea full-screen).
- **Infra**: per-format escaping module, `testhelpers` (own module), golden-file test
  conventions, CQRS-shaped API (v0.30.0 breaking rename).

## 2. Capability vs applicability

| Capability              | Status                     | Reason                                                                                                                      |
| ----------------------- | -------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| Table/diagram rendering | ⚪ N/A — category mismatch | appkit emits HTTP JSON (docs catalog, health endpoints), not terminal/report output; no CLI ships in this repo              |
| NOM progress            | ⚪ N/A                     | Long-running-work visualization targets build/install tools, not an HTTP framework                                          |
| testhelpers             | ⚪ N/A                     | Repo testing is stdlib `testing` + family `ssetest`/`httpspec`; adopting a second assert library would fragment conventions |
| Version currency        | 🟢                         | v0.38.7 is the newest non-retracted tag (moot until adoption)                                                               |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **No rendering consumer exists.** The repo's "users" of output are Go consumers of HTTP
   APIs and library imports. An unused require here would be pure dead weight.
2. **The docs module's contract is machine-readable** (OpenAPI/AsyncAPI/catalog JSON) —
   human-facing rendering of that data belongs to the consumer's tooling, and if the family
   ever builds a fleet CLI for it, that CLI is a cmdguard project consuming go-output —
   still not go-appkit.
3. **When it flips:** only with the birth of an appkit-facing CLI or operator tool (e.g. a
   `appkit` diagnostics binary rendering service catalogs/probes as tables). Then: import
   the root module + only the needed format submodules, and honor the retracted-tag list
   when pinning (v0.32.1/v0.33.0/v0.34.0 must never appear in a go.mod).

## 4. Recommendations

1. Nothing to track — this report is the record of the category-mismatch decision.
2. If a fleet CLI materializes, this is the rendering library to reach for (16 formats,
   zero-dep root), NOT hand-rolled tabwriters — cite this file as prior art at that point.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'larsartmann/go-output' --glob '*.go' --glob 'go*.sum'   # empty → absent
ls docs/docs.go docs/README.md                                  # docs module = HTTP catalog, not tables
head -1 /home/lars/projects/go-output/go.mod
sed -n '21,28p' /home/lars/projects/go-output/go.mod            # retracted tags
git -C /home/lars/projects/go-output tag --sort=-v:refname | head -1  # v0.38.7
```
