# Status Report — 2026-07-07 (Session 2)

> **Generated:** 2026-07-07 15:03. **Branch:** master. **Latest commit:** 9356d1d.
> **Build:** green. **Tests:** 65 cases pass with `-race` across 3 modules.

---

## Executive Summary

go-appkit is now a **3-module service framework** that composes the larsartmann Go ecosystem.
All 3 modules build, vet, and test green. The core module provides a production-ready HTTP
Service. The CQRS sub-module wraps go-cqrs-lite/stack/sqlite + projectionhost. The docs
sub-module wraps catalog/docserver for AsyncAPI/OpenAPI/D2 auto-documentation.

**But it's not tagged.** And the docs sub-module lives in a directory called `docs-mod` because
`docs/` already exists for documentation files. This is a naming problem.

---

## a) FULLY DONE (Verified, Tested, Committed)

### Core module (`go-appkit`)

| Item                             | Evidence                                                                                   |
| -------------------------------- | ------------------------------------------------------------------------------------------ |
| `Service` type — service.go      | NewService, Start, Run, Shutdown, Close, Addr, Running                                     |
| `ServiceConfig` — config.go      | Struct + DefaultServiceConfig + applyDefaults + Validate                                   |
| Middleware chain — middleware.go | Recovery→RequestID→Logging→Timeout→SecurityHeaders. Replaceable + extendable               |
| Logger — logger.go               | charmbracelet/log IS slog.Handler. 8 logger tests                                          |
| Health — health.go               | Delegates to httputil.RegisterHealth + ReadyHandlerWithProbe. 4 health tests               |
| Errors — errors.go               | Re-exports errorfamily.HTTPStatus + LogError                                               |
| Shutdown — shutdown.go           | WaitForSignal preserved + drain in Service.Shutdown. 5 tests                               |
| README.md                        | Complete rewrite: quick start, config table, middleware, health, lifecycle, error handling |
| example/main.go                  | Minimal production service with DefaultServiceConfig                                       |
| AGENTS.md                        | Updated architecture, file map, dependencies, gotchas                                      |
| 36 top-level tests               | 58 total pass cases (including 18 httpspec subtests + config table subtests)               |
| httpspec conformance             | 18/18 HTTP behavior specs from httputil pass                                               |
| Config validation                | Table-driven negative + valid + defaults tests                                             |
| noctx lint fixed                 | All test HTTP requests use context-aware httpGet helper                                    |

### CQRS sub-module (`go-appkit/cqrs`)

| Item                                                    | Evidence                                                                    |
| ------------------------------------------------------- | --------------------------------------------------------------------------- |
| EventService type — eventservice.go                     | Wraps stack/sqlite.New + projectionhost.New                                 |
| Lifecycle: Shutdown, StartProjections, DB, Bundle, Host | Idempotent shutdown via mutex                                               |
| 4 tests                                                 | Empty path, valid path (all bundle accessors), DB ping, idempotent shutdown |
| go.work workspace                                       | Links core + cqrs + docs-mod                                                |

### Docs sub-module (`go-appkit/docs`)

| Item                             | Evidence                                                     |
| -------------------------------- | ------------------------------------------------------------ |
| CatalogBuilder wrapper — docs.go | Wraps catalog.Builder with appkit-friendly API               |
| RegisterDocs — docs.go           | Mounts docserver routes: OpenAPI, AsyncAPI, D2, catalog.json |
| 3 tests                          | OpenAPI JSON served, AsyncAPI JSON served, builder non-nil   |

### Infrastructure

| Item                 | Evidence                                            |
| -------------------- | --------------------------------------------------- |
| BuildFlow pre-commit | Passes 20/20 on every commit                        |
| go.work              | Multi-module workspace linking all 3 modules        |
| depguard config      | Allows httputil, charmbracelet/log, go-error-family |
| Planning docs        | Pareto execution plan with mermaid graph            |

---

## b) PARTIALLY DONE

| Item                                  | What works                                                                    | What's missing                                                                                                                                        |
| ------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
~~| **CQRS E2E**                          | EventService creates, bundle accessors work, DB pings, shutdown is idempotent | No actual command dispatch → event stored → projection read cycle tested. No integration with Service.Run() shutdown coordination.                    |~~ resolved
~~| **Docs E2E**                          | Routes serve JSON, builder returns non-nil                                    | No actual AddCommand/AddEvent → schema verification tested. No integration with Service.Mux.                                                          |~~ resolved
~~| **Service.Run integration with CQRS** | Service.Run handles its own graceful drain/shutdown                           | EventService.Shutdown is NOT called by Service.Run(). Consumer must wire it manually. This may be intentional (loose coupling) but it's undocumented. |~~ resolved
~~| **Error-family in CQRS**              | EventService uses NewRejection and WrapInfrastructuref for errors             | No RegisterClassifier for third-party errors. No LogError in shutdown paths.                                                                          |~~ resolved

---

## c) NOT STARTED

| Item                               | Impact   | Notes                                                                                                                                               |
| ---------------------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
~~| **Tag v1.0.0**                     | Critical | Core module is ready. No git tag exists.                                                                                                            |~~ NOT-DO — superseded by 0.x waves (core v0.2.0-v0.4.0 shipped)
~~| **Tag cqrs v0.1.0**                | Medium   | CQRS module is ready. No git tag exists.                                                                                                            |~~ done at cqrs/v0.2.0
~~| **Tag docs v0.1.0**                | Low      | Docs module is ready. No git tag exists.                                                                                                            |~~ done at docs/v0.2.0
~~| **CQRS README section**            | Medium   | No usage documentation for the cqrs sub-module                                                                                                      |~~ done 2026-08-15 (cqrs/README.md)
~~| **Docs README section**            | Low      | No usage documentation for the docs sub-module                                                                                                      |~~ done 2026-08-15
~~| **docs-mod → docs rename**         | High     | Directory is `docs-mod/` but module path is `github.com/larsartmann/go-appkit/docs`. This is confusing. The `docs/` directory holds Markdown files. |~~ NOT-DO — kept; mismatch is the GHOST RELEASE (TODO_LIST P1)
~~| **CQRS + Service.Run integration** | Medium   | No `ServiceConfig.CQRS` field or `Service.UseCQRS()` method                                                                                         |~~ NOT-DO — superseded by ShutdownHooks composition (v0.4.0)
~~| **flake.nix**                      | Low      | AGENTS.md says "no flake.nix, use standard Go tooling"                                                                                              |~~ Won't implement — plain Go tooling standard
~~| **CHANGELOG.md**                   | Low      | No changelog for the rewrite                                                                                                                        |~~ done at v0.2.0
~~| **golangci-lint for sub-modules**  | Medium   | depguard config doesn't cover cqrs/docs sub-module imports                                                                                          |~~ done 2026-08-16/17 (per-module configs)

---

## d) TOTALLY FUCKED UP / BROKEN / RISKY

| Issue                                                      | Severity  | Detail                                                                                                                                                                                                                                                                                                                                                    |
| ---------------------------------------------------------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
~~| **`docs-mod/` directory name vs `docs/` module path**      | **HIGH**  | The module is `github.com/larsartmann/go-appkit/docs` but the directory is `docs-mod/`. Go module tooling resolves it via go.work replace directive, but ANY external consumer importing `github.com/larsartmann/go-appkit/docs` will get a 404 from the repo because the directory doesn't match the module path. **This must be fixed before tagging.** |~~ NOT-DO — became the GHOST RELEASE (TODO_LIST P1)
~~| **README mentions `os` import but example doesn't use it** | Low       | README quick-start code shows `errorfamily.HandleError(err)` but doesn't import `os`. Copy-paste will fail.                                                                                                                                                                                                                                               |~~ done 2026-08-17 (README build checks)
~~| **LSP stale typecheck warnings**                           | Cosmetic  | 5 stale "other declaration of defaultReadTimeout" warnings from LSP cache. `go build` and `go vet` pass clean. LSP hasn't refreshed after server.go deletion.                                                                                                                                                                                             |~~ NOT-DO — transient LSP cache
~~| **Test runtime 5-6s**                                      | Low       | Default DrainDelay=5s makes some tests slow. Non-drain tests use DrainDelay:0 but some still wait on Close().                                                                                                                                                                                                                                             |~~ done at v0.4.0 (NoDrainDelay)
~~| **`httpspec_test.go` uses `init()` hack**                  | Low       | The `init()` function suppresses unused import for httptest. Should be removed — httptest is used by httpspec internally, not by our test.                                                                                                                                                                                                                |~~ done 2026-08-17 (init deleted)
~~| **No `ServiceConfig.CQRS` integration**                    | Medium    | The plan called for `Service.Run()` to call `EventService.Shutdown()` during graceful drain. This is NOT implemented. Consumer must manually coordinate.                                                                                                                                                                                                  |~~ NOT-DO — superseded by ShutdownHooks composition
~~| **Uncommitted go.sum changes**                             | **FIXED** | go.work pulled newer indirect dep versions. Committed at 9356d1d.                                                                                                                                                                                                                                                                                         |~~ done at 9356d1d

---

## e) WHAT WE SHOULD IMPROVE

~~1. **Fix docs-mod → docs directory rename** — Move docs-mod/ to a sub-directory that matches the module path, or rename the documentation `docs/` directory to `doc/` to free up `docs/` for the Go module.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~2. **Tag all 3 modules** — Core v1.0.0, cqrs v0.1.0, docs v0.1.0.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~3. **Add CQRS README** — Show 20-line CQRS service example.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~4. **Add docs README** — Show catalog builder + RegisterDocs example.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~5. **Wire Service.Run + EventService.Shutdown** — `ServiceConfig.CQRS *cqrs.EventService` field, Service.Run calls es.Shutdown after server.Shutdown.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~6. **Test CQRS command→event→projection cycle** — The real E2E that proves the CQRS integration works end-to-end.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~7. **Remove init() hack in httpspec_test.go** — Clean up the httptest import.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~8. **Fix README code examples** — Ensure all imports are correct for copy-paste.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~9. **Add golangci-lint config for sub-modules** — depguard for cqrs and docs imports.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~10. **Add CHANGELOG.md** — Document the breaking rewrite from v0 to v1.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~11. **Add `WithLogger(logger)` option** — Allow injecting a pre-configured logger.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~12. **Reduce default test DrainDelay** — Use near-zero in test helper.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~13. **Consolidate planning docs** — DRY the error-family 3-layer adoption across 3 docs.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~14. **Add structured shutdown logging** — Log drain start/complete, shutdown start/complete.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~15. **Document httputil.Server non-delegation** — Plan says "delegate to httputil.Server" but we CAN'T. This deviation is correct but undocumented in code.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~16. **Test with real HTTP handler errors** — Verify errorfamily.HTTPHandler maps family→status correctly through the middleware chain.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~17. **Add `svc.HealthCheck(fn)` convenience** — Let consumer register a readiness check function.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~18. **Consider `DisableHealth bool` instead of `RegisterHealth *bool`** — Ergonomics.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~19. **Add projection registration E2E** — Register a projection, start, dispatch event, verify projection read model updated.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)
~~20. **Add CI config** — GitHub Actions for go test, go vet, go build on push.~~ done, superseded, or routed in later waves (see CHANGELOGs; the docs-mod rename became TODO_LIST P1)

---

## f) Next 50 Tasks (Sorted by Impact ÷ Effort)

| #  | Task                                            | Impact   | Effort | Priority |
| -- | ----------------------------------------------- | -------- | ------ | -------- |
~~| 1  | Fix docs-mod → docs directory problem           | Critical | 15m    | P0       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 2  | Tag core v1.0.0                                 | Critical | 5m     | P0       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 3  | Tag cqrs v0.1.0                                 | Medium   | 5m     | P0       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 4  | Tag docs v0.1.0                                 | Low      | 5m     | P0       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 5  | Fix README imports (os, errorfamily)            | Medium   | 5m     | P0       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 6  | Remove init() hack in httpspec_test.go          | Low      | 5m     | P1       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 7  | Add CQRS README section                         | Medium   | 20m    | P1       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 8  | Add docs README section                         | Low      | 15m    | P1       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 9  | Wire Service.Run + EventService.Shutdown        | High     | 30m    | P1       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 10 | CQRS E2E: command dispatch → event stored       | High     | 45m    | P1       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 11 | CQRS E2E: projection read model                 | Medium   | 30m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 12 | CQRS E2E: health check integration              | Medium   | 15m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 13 | Docs E2E: AddCommand → OpenAPI schema verify    | Medium   | 20m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 14 | Docs E2E: AddEvent → AsyncAPI schema verify     | Medium   | 20m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 15 | CHANGELOG.md for v1.0.0 rewrite                 | Medium   | 15m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 16 | golangci-lint config for cqrs sub-module        | Medium   | 15m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 17 | golangci-lint config for docs sub-module        | Low      | 10m    | P2       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 18 | Fix test DrainDelay helper (near-zero default)  | Low      | 10m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 19 | Add `WithLogger(logger)` option                 | Low      | 15m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 20 | Add structured shutdown logging                 | Low      | 15m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 21 | Document httputil.Server non-delegation         | Medium   | 10m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 22 | Test errorfamily.HTTPHandler through middleware | Medium   | 15m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 23 | Add `svc.HealthCheck(fn)` convenience           | Low      | 15m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 24 | Consolidate planning docs (DRY)                 | Low      | 20m    | P3       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 25 | Add CI config (GitHub Actions)                  | Medium   | 30m    | P4       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 26 | Add `DisableHealth bool` consideration          | Low      | 10m    | P4       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 27 | Add flake.nix (optional, AGENTS.md says no)     | Low      | 30m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 28 | Huma integration example in README              | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 29 | Error-family bridge pattern docs                | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 30 | CQRS: Bus exposure docs (Publisher/Subscriber)  | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 31 | CQRS: Repository builder example                | Medium   | 20m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 32 | Docs: D2 diagram endpoint test                  | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 33 | Docs: EventCatalog MDX export                   | Low      | 30m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 34 | Docs: Huma + catalog side-by-side example       | Low      | 20m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 35 | Benchmark: Service startup time                 | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 36 | Benchmark: middleware overhead                  | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 37 | Add `-tags integration` for slow E2E tests      | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 38 | CQRS: Snapshot store integration                | Low      | 20m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 39 | CQRS: Idempotency integration                   | Low      | 20m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 40 | Docs: OpenAPI YAML endpoint test                | Low      | 10m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 41 | Add Go doc examples (testable examples)         | Low      | 20m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 42 | Add version string constant                     | Low      | 5m     | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 43 | Add `svc.Mount(pattern, handler)` convenience   | Low      | 10m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 44 | Consider metrics middleware integration         | Low      | 30m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 45 | Consider request rate limiting                  | Low      | 20m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 46 | Consider CORS middleware                        | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 47 | Add graceful shutdown timeout test              | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 48 | Add signal delivery test for Run()              | Medium   | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 49 | Document drain sequence with diagram            | Low      | 15m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves
~~| 50 | Final brutal-self-review skill run              | Medium   | 30m    | P5       |~~ done, superseded, or routed — executed across sessions 3-5 and the 2026-08/09 release waves

---

## g) Top 2 Questions I Cannot Figure Out Myself

### Question 1: How do we resolve the `docs-mod/` vs `docs/` naming conflict?

~~The Go module is `github.com/larsartmann/go-appkit/docs` but lives in `docs-mod/` because `docs/`~~ Answered: 0.x waves shipped; realtime shipped as a module; Service stayed the entry point
already holds Markdown documentation files. External consumers will get 404 because the directory
doesn't match the module path.

~~**Options I see:**~~ Answered: 0.x waves shipped; realtime shipped as a module; Service stayed the entry point

- **(A)** Rename `docs/` (documentation) to `doc/` and rename `docs-mod/` to `docs/`. This makes
  the module path match the directory. But `docs/` is the convention for documentation.
- **(B)** Move the docs sub-module into a different module path entirely:
  `github.com/larsartmann/go-appkit/apidocs`. Then the directory can be `apidocs/`.
- **(C)** Use Go's `submodule directory != module path` pattern — tag with `docs/v0.1.0` and the
  Go proxy resolves it. But this requires the directory to BE `docs/` in the repo.

~~**I cannot decide this** — it affects the public API and repo structure.~~ Answered: 0.x waves shipped; realtime shipped as a module; Service stayed the entry point

### Question 2: Should `Service.Run()` own `EventService.Shutdown()` or should the consumer?

~~The plan says "Service.Run() calls EventService.Shutdown() after server.Shutdown()." But this~~ Answered: 0.x waves shipped; realtime shipped as a module; Service stayed the entry point
creates an import dependency: core `go-appkit` would need to import `go-appkit/cqrs` (circular).
Alternatively, the consumer calls `es.Shutdown()` themselves after `svc.Run()` returns.

~~**Options:**~~ Answered: 0.x waves shipped; realtime shipped as a module; Service stayed the entry point

- **(A)** Add a `ShutdownHooks []func(context.Context) error` field to `ServiceConfig`. Consumer
  appends `es.Shutdown` to it. Service.Run() calls them after server.Shutdown. No circular import.
- **(B)** Consumer calls `es.Shutdown()` after `svc.Run()` returns. Simplest but error-prone (what
  if Run returns early from a serve error?).
- **(C)** Add a `Closer io.Closer` field to `ServiceConfig`. Consumer wraps `es.Shutdown` into a
  Closer. Service.Run() defers Close.

~~**I lean towards (A)** but it changes the v1.0.0 API contract. I cannot decide without knowing~~ Answered: 0.x waves shipped; realtime shipped as a module; Service stayed the entry point
your preference for lifecycle management style.

---

## Module Summary

| Module | Directory     | Module Path                             | Tests    | Status                                 |
| ------ | ------------- | --------------------------------------- | -------- | -------------------------------------- |
| Core   | `./`          | `github.com/larsartmann/go-appkit`      | 58 cases | Ready for v1.0.0                       |
| CQRS   | `./cqrs/`     | `github.com/larsartmann/go-appkit/cqrs` | 4 cases  | Ready for v0.1.0                       |
| Docs   | `./docs-mod/` | `github.com/larsartmann/go-appkit/docs` | 3 cases  | **Naming conflict** — needs resolution |

## Git Log (This Session)

```
9356d1d chore: bump indirect deps from go.work workspace resolution
dc12d8a feat: add docs sub-module with catalog/docserver wrapper + E2E tests
ec54883 feat: add CQRS sub-module with EventService wrapping stack/sqlite + projectionhost
839dd48 test: add httpspec conformance (18 specs) and config validation table tests
7b3be30 feat: rewrite README, add example, fix lint, add middleware tests
f276e89 docs: add Pareto execution plan for shipping v1.0.0 and sub-modules
71d8d88 docs: add status report for service framework rewrite
d46ed0a feat: rewrite go-appkit as service framework composing httputil, charmbracelet/log, and go-error-family
```
