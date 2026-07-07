# Planning — Modular Service Framework

> Transforming go-appkit from a "5 random helpers" library into a modular service framework
> that composes the larsartmann Go ecosystem into a unified, production-ready service.

## Documents

Read in this order:

1. **[framework-architecture.md](./framework-architecture.md)** — The 3-module design
   (core / cqrs / docs), dependency direction, versioning strategy. Start here.
2. **[design-decisions.md](./design-decisions.md)** — 10 locked decisions with rationale
   and alternatives considered. Understanding WHY before HOW.
3. **[integrations.md](./integrations.md)** — All ecosystem APIs verified from source code.
   httputil, charmbracelet/log, error-family, cqrs-lite stack, Huma, catalog.
4. **[execution-plan.md](./execution-plan.md)** — 46 tasks across 3 modules, sequenced by
   critical path, ~28h total effort. The build checklist.
5. **[improvement-audit.md](./improvement-audit.md)** — 15 weaknesses found in the earlier
   monolithic plan with severity ratings and fixes. Why we went modular.

## TL;DR

| Module             | What                                                                | Deps                                      | Tag    |
| ------------------ | ------------------------------------------------------------------- | ----------------------------------------- | ------ |
| `go-appkit` (core) | HTTP service framework: server, middleware, health, logging, errors | httputil, charmbracelet/log, error-family | v1.0.0 |
| `go-appkit/cqrs`   | CQRS/ES integration: event store, projections, bus                  | stack/sqlite, projectionhost, core        | v0.1.0 |
| `go-appkit/docs`   | Auto-documentation: AsyncAPI, OpenAPI, D2                           | catalog, core                             | v0.1.0 |

Core ships first. CQRS and docs are opt-in sub-modules.
