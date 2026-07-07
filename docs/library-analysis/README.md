# Library Analysis — Integration Status

> **Date:** 2026-07-05 (revised 2026-07-06). **Status:** Superseded by [planning docs](../planning/).
>
> The original per-library analyses (in [`archive/`](./archive/)) assessed go-appkit as a
> "low-level skeleton library." That assessment is **stale**. go-appkit is being rebuilt as a
> **modular service framework** that composes these libraries as first-class dependencies.

---

## Integration Status

| Library                                | Role in go-appkit                                          | Module | Status                |
| -------------------------------------- | ---------------------------------------------------------- | ------ | --------------------- |
| **httputil** v0.4.0                    | Middleware chain, health endpoints, httpspec tests         | Core   | **Core dependency**   |
| **charmbracelet/log** v1.0.0           | Pretty slog handler (replaces 108-line custom handler)     | Core   | **Core dependency**   |
| **go-error-family** v0.6.1             | Error classification, HTTP mapping, CLI exit codes         | Core   | **Core dependency**   |
| **go-cqrs-lite/stack/sqlite** v3.6.0   | CQRS/ES event store, projections, bus                      | `cqrs` | **Opt-in sub-module** |
| **go-cqrs-lite/catalog** v3.6.0        | AsyncAPI, OpenAPI, D2 auto-documentation                   | `docs` | **Opt-in sub-module** |
| **go-cqrs-lite/projectionhost** v3.6.0 | Projection lifecycle (crash recovery, DLQ, backoff)        | `cqrs` | **Opt-in sub-module** |
| cqrs-htmx v4.1.1                       | Application framework (consumer of appkit, not dependency) | —      | Not used              |
| go-branded-id v0.3.1                   | Phantom-typed IDs (transitive via cqrs-lite)               | —      | Transitive only       |
| cmdguard v2.10.2                       | CLI framework (application-level, not library)             | —      | Not used              |

## What changed

The original analysis concluded that httputil was "duplication to resolve," go-cqrs-lite was
"wrong layer," and go-error-family was "consider adopting." All three conclusions are now
**inverted** — they are dependencies that appkit composes into a unified service framework.

See:

- [framework-architecture.md](../planning/framework-architecture.md) — modular 3-module design
- [design-decisions.md](../planning/design-decisions.md) — 10 locked decisions
- [integrations.md](../planning/integrations.md) — verified ecosystem APIs
- [execution-plan.md](../planning/execution-plan.md) — 46 tasks across 3 modules

## Archive

The original per-library analyses are preserved in [`archive/`](./archive/) for historical
context. They reflect the pre-framework assessment and should not be used as current guidance.
