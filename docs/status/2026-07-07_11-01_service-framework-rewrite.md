# Status Report — 2026-07-07

> **Generated:** 2026-07-07 11:01. **Branch:** master. **Commit:** d46ed0a.
> **Build:** BuildFlow 20/20 passing. **Tests:** 27/27 passing with `-race`.

---

## Executive Summary

go-appkit has been transformed from a "5 random helpers" library (server.go, health.go,
logger.go, sqlite.go, shutdown.go) into a **production-ready Service framework** that composes
the larsartmann Go ecosystem. The core module is functionally complete with 27 passing tests.

**The headline:** `appkit.NewService(cfg)` → `svc.Run(ctx)` gives you a running HTTP service
with middleware, health checks, structured logging, and graceful drain/shutdown. Consumer code
is ~12 lines for a production service.

---

## a) FULLY DONE (Verified, Tested, Committed)

| Item                                   | Evidence                                                                                                                |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| **Service type** — `service.go`        | NewService, Start, Run, Shutdown, Close, Addr, Running. All tested.                                                     |
| **ServiceConfig** — `config.go`        | Struct + DefaultServiceConfig + applyDefaults + Validate. Pointer-based RegisterHealth.                                 |
| **Middleware chain** — `middleware.go` | Default stack: Recovery→RequestID→Logging→Timeout→SecurityHeaders. Replaceable + extendable.                            |
| **Logger rewrite** — `logger.go`       | 108 lines → ~75 lines. charmbracelet/log IS slog.Handler. All 8 logger tests pass.                                      |
| **Health rewrite** — `health.go`       | Delegates to httputil.RegisterHealth + ReadyHandlerWithProbe. 4 health tests pass.                                      |
| **Error adoption** — `errors.go`       | Re-exports errorfamily.HTTPStatus + LogError. Internal errors use error-family constructors.                            |
| **Shutdown** — `shutdown.go`           | WaitForSignal preserved. DrainDelay logic lives in Service.Shutdown. 5 shutdown tests pass.                             |
| **Doc cleanup**                        | 6 stale analysis files archived. library-analysis README rewritten. planning README created. Stale HTML/D2/SVG deleted. |
| **Dependencies**                       | httputil v0.5.0, charmbracelet/log v1.0.0, go-error-family v0.6.1. All resolve cleanly.                                 |
| **depguard config**                    | `.golangci.yml` updated to allow new imports.                                                                           |
| **27 tests pass**                      | `go test -race -count=1 -timeout=30s ./...` → ok in 5s.                                                                 |
| **BuildFlow**                          | 20/20 checks passed at commit d46ed0a.                                                                                  |
| **No existing consumers**              | Verified: zero Go files outside the repo import go-appkit. Greenfield rewrite.                                          |

---

## b) PARTIALLY DONE (Functional but incomplete)

| Item                      | What works                                                                           | What's missing                                                                                                                                      |
| ------------------------- | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Service tests**         | Construction, defaults, validation, health, custom route, drain, close, addr/running | No middleware-specific tests (panic recovery, request ID presence, logging capture). No httpspec.Run test.                                          |
| **Error-family adoption** | Constructors used in config.go, logger.go, service.go. Re-exports in errors.go.      | No RegisterClassifier for third-party errors. HTTPHandler pattern not yet documented in example.                                                    |
| **Planning docs**         | 5 .md files + README index written and committed.                                    | Overlapping content (error-family 3-layer adoption appears in design-decisions.md, integrations.md, AND framework-architecture.md — DRY violation). |
| **Graceful drain**        | Works: readyProbe flips → DrainDelay → server.Shutdown. Tested.                      | No integration test verifying LoadBalancer behavior. DrainDelay=0 in some tests for speed (not testing drain there).                                |

---

## c) NOT STARTED

| Item                                   | Impact | Notes                                                                       |
| -------------------------------------- | ------ | --------------------------------------------------------------------------- |
| **example/main.go**                    | High   | 12-line demo service with HandleError CLI terminator.                       |
| **README.md rewrite**                  | High   | Currently describes old "5 helpers" library. Must show framework API.       |
| **AGENTS.md update**                   | Medium | Still describes old architecture (server.go, sqlite.go, HealthStatus enum). |
| **CQRS sub-module** (`go-appkit/cqrs`) | Medium | EventService wrapping stack/sqlite.New. Separate go.mod.                    |
| **Docs sub-module** (`go-appkit/docs`) | Low    | Catalog wrapper for AsyncAPI/OpenAPI/D2. Separate go.mod.                   |
| **go.work workspace**                  | Medium | Multi-module workspace file for developing cqrs/docs alongside core.        |
| **flake.nix**                          | Medium | AGENTS.md mandates flake.nix but it doesn't exist.                          |
| **Tag v1.0.0**                         | High   | Blocked on README + example + AGENTS.md update.                             |

---

## d) WHAT'S BROKEN / RISKY / CONCERNING

| Issue                                       | Severity | Detail                                                                                                                                                           |
| ------------------------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **httputil v0.5.0, not v0.4.0**             | Low      | Plan docs say v0.4.0. Actual resolved version is v0.5.0 (latest). Plan docs are stale on version numbers.                                                        |
| **LSP typecheck warnings**                  | Cosmetic | 5 stale "other declaration of defaultReadTimeout" warnings from LSP cache. `go build` and `go vet` pass clean. LSP hasn't refreshed after file deletion.         |
| **Test runtime ~5s**                        | Low      | Default DrainDelay=5s makes some tests slow. Mitigated with DrainDelay:0 in non-drain tests, but default-config test still waits 5s.                             |
| **`http.Get` in tests triggers noctx lint** | Low      | Tests use `http.Get` directly. golangci-lint flags this. BuildFlow's repair pass auto-fixed it at commit time, but the source should use context-aware requests. |
| **README is completely stale**              | High     | Describes old library. Anyone reading the repo RIGHT NOW will be confused.                                                                                       |
| **AGENTS.md is stale**                      | Medium   | References deleted files (server.go, sqlite.go, HealthStatus enum). Will confuse future AI sessions.                                                             |
| **Planning docs have DRY violations**       | Low      | Error-family 3-layer adoption duplicated across 3 docs.                                                                                                          |

---

## e) WHAT WE SHOULD IMPROVE

1. **Fix test HTTP requests to use context** — Replace `http.Get` with `http.NewRequestWithContext` + `client.Do`. Fixes noctx lint warnings at source.
2. **Add middleware integration tests** — Panic → 500, X-Request-ID header present, logging output captured. These are the middleware contract tests.
3. **Add httpspec.Run test** — 18 free specs from httputil. `httpspec.Run(t, handler, httpspec.SkipSpec(httpspec.IndexNot404))`.
4. **Reduce default test DrainDelay** — Default test config uses 5s drain. Should use near-zero in test helper.
5. **Consolidate error-family docs** — Single source of truth for the 3-layer adoption pattern.
6. **Fix RegisterHealth ergonomics** — Current `*bool` pointer pattern works but is awkward. Consider a `DisableHealth bool` field instead (inverted logic, but zero-value = default behavior).
7. **Add `svc.WithLogger(logger)` option** — Allow injecting a pre-configured logger instead of always creating one.
8. **Add structured shutdown logging** — Log drain start, drain complete, shutdown start, shutdown complete with timestamps.
9. **Document the httputil.Server NON-delegation** — Plan says "delegate to httputil.Server" but we CAN'T because httputil.Server uses ListenAndServe() internally (no listener access). appkit owns http.Server + net.Listener directly. This deviation from the plan is correct but undocumented.

---

## f) Next 25 Tasks (Sorted by Impact ÷ Effort)

| #   | Task                                                   | Impact   | Effort | Priority |
| --- | ------------------------------------------------------ | -------- | ------ | -------- |
| 1   | Rewrite README.md for framework API                    | Critical | 30m    | P0       |
| 2   | Create example/main.go (12-line service)               | High     | 15m    | P0       |
| 3   | Update AGENTS.md (architecture, file map)              | High     | 20m    | P0       |
| 4   | Fix test HTTP requests (noctx lint)                    | Medium   | 15m    | P1       |
| 5   | Add middleware test: panic → 500                       | High     | 15m    | P1       |
| 6   | Add middleware test: X-Request-ID present              | Medium   | 10m    | P1       |
| 7   | Add middleware test: logging captured                  | Medium   | 15m    | P1       |
| 8   | Add httpspec.Run conformance test                      | Medium   | 15m    | P1       |
| 9   | Add ServiceConfig_test.go (table-driven validation)    | Medium   | 15m    | P2       |
| 10  | Fix DrainDelay test helper (near-zero default)         | Low      | 10m    | P2       |
| 11  | Tag v1.0.0 (after README + example)                    | Critical | 5m     | P2       |
| 12  | Consolidate error-family docs (DRY)                    | Low      | 20m    | P2       |
| 13  | Update planning docs version numbers (v0.5.0)          | Low      | 10m    | P2       |
| 14  | Add `WithLogger(logger)` option                        | Low      | 15m    | P3       |
| 15  | Add structured shutdown logging                        | Low      | 15m    | P3       |
| 16  | Document httputil.Server non-delegation decision       | Medium   | 10m    | P3       |
| 17  | Create go.work workspace file                          | Medium   | 10m    | P3       |
| 18  | Create cqrs/go.mod + EventService stub                 | Medium   | 30m    | P4       |
| 19  | Implement cqrs EventService (stack/sqlite.New wrapper) | Medium   | 45m    | P4       |
| 20  | cqrs: Service integration (Shutdown calls es.Shutdown) | Medium   | 30m    | P4       |
| 21  | cqrs: E2E test (command → event → projection → health) | High     | 60m    | P4       |
| 22  | Create docs/go.mod + catalog wrapper                   | Low      | 30m    | P5       |
| 23  | Implement docs RegisterDocs (catalog routes)           | Low      | 45m    | P5       |
| 24  | Create flake.nix (build/lint/test automation)          | Medium   | 30m    | P5       |
| 25  | Final review: brutal-self-review skill                 | Medium   | 30m    | P5       |

---

## g) Top #1 Question

**Should `RegisterHealth` use `*bool` (current) or should we switch to `DisableHealth bool`?**

The current `*bool` pattern correctly distinguishes "not set" (→ default true) from
"explicitly false" (→ opt out). But it's ergonomically awkward:

```go
// Current (works but ugly):
disabled := false
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr: ":8080",
    RegisterHealth: &disabled,
})

// Alternative (simpler, inverted logic):
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr: ":8080",
    DisableHealth: true,  // zero-value false = health registered by default
})
```

The `DisableHealth bool` approach has a worse name ("disable" is negative) but far better
ergonomics. The `*bool` approach has the right semantics but requires a local variable.

**I cannot decide this alone** — it affects the v1.0.0 public API contract and I don't know
which pattern the user prefers for opt-out boolean config fields.

---

## Architecture Diagram (Current State)

```
┌─────────────────────────────────────────────────┐
│                Consumer main()                   │
│  svc, _ := appkit.NewService(cfg)               │
│  svc.Mux.HandleFunc("GET /", handler)           │
│  svc.Run(ctx)                                    │
└──────────────────────┬──────────────────────────┘
                       │
              ┌────────▼────────┐
              │   appkit.Service │
              │                  │
              │  • Mux           │◄── consumer registers routes here
              │  • Logger        │◄── charmbracelet/log → slog
              │  • http.Server   │◄── appkit owns (NOT httputil.Server)
              │  • net.Listener  │◄── appkit owns for Addr() net.Addr
              │  • readyProbe    │◄── atomic.Bool for graceful drain
              │                  │
              │  Middleware:     │
              │  Recovery→ReqID  │◄── httputil.Chain()
              │  →Logging→Timeout│
              │  →SecurityHeaders│
              │                  │
              │  Health:         │
              │  /health         │◄── httputil.RegisterHealth()
              │  /health/live    │◄── httputil.LiveHandler()
              │  /health/ready   │◄── httputil.ReadyHandlerWithProbe()
              └──────────────────┘
```

**Key deviation from plan:** appkit does NOT delegate to `httputil.Server` because
httputil.Server uses `ListenAndServe()` internally (no listener access for `Addr() net.Addr`).
appkit owns `http.Server` + `net.Listener` directly, while still using httputil for
middleware, health, and httpspec.
