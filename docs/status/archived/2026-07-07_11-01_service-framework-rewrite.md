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

| Item | What works                | What's missing                                                                       |
| ---- | ------------------------- | ------------------------------------------------------------------------------------ |
| ~~   | **Service tests**         | Construction, defaults, validation, health, custom route, drain, close, addr/running |
| ~~   | **Error-family adoption** | Constructors used in config.go, logger.go, service.go. Re-exports in errors.go.      |
| ~~   | **Planning docs**         | 5 .md files + README index written and committed.                                    |
| ~~   | **Graceful drain**        | Works: readyProbe flips → DrainDelay → server.Shutdown. Tested.                      |

---

## c) NOT STARTED

| Item | Impact                                 | Notes  |
| ---- | -------------------------------------- | ------ |
| ~~   | **example/main.go**                    | High   |
| ~~   | **README.md rewrite**                  | High   |
| ~~   | **AGENTS.md update**                   | Medium |
| ~~   | **CQRS sub-module** (`go-appkit/cqrs`) | Medium |
| ~~   | **Docs sub-module** (`go-appkit/docs`) | Low    |
| ~~   | **go.work workspace**                  | Medium |
| ~~   | **flake.nix**                          | Medium |
| ~~   | **Tag v1.0.0**                         | High   |

---

## d) WHAT'S BROKEN / RISKY / CONCERNING

| Issue | Severity                                    | Detail   |
| ----- | ------------------------------------------- | -------- |
| ~~    | **httputil v0.5.0, not v0.4.0**             | Low      |
| ~~    | **LSP typecheck warnings**                  | Cosmetic |
| ~~    | **Test runtime ~5s**                        | Low      |
| ~~    | **`http.Get` in tests triggers noctx lint** | Low      |
| ~~    | **README is completely stale**              | High     |
| ~~    | **AGENTS.md is stale**                      | Medium   |
| ~~    | **Planning docs have DRY violations**       | Low      |

---

## e) WHAT WE SHOULD IMPROVE

~~1. **Fix test HTTP requests to use context** — Replace `http.Get` with `http.NewRequestWithContext` + `client.Do`. Fixes noctx lint warnings at source.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~2. **Add middleware integration tests** — Panic → 500, X-Request-ID header present, logging output captured. These are the middleware contract tests.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~3. **Add httpspec.Run test** — 18 free specs from httputil. `httpspec.Run(t, handler, httpspec.SkipSpec(httpspec.IndexNot404))`.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~4. **Reduce default test DrainDelay** — Default test config uses 5s drain. Should use near-zero in test helper.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~5. **Consolidate error-family docs** — Single source of truth for the 3-layer adoption pattern.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~6. **Fix RegisterHealth ergonomics** — Current `*bool` pointer pattern works but is awkward. Consider a `DisableHealth bool` field instead (inverted logic, but zero-value = default behavior).~~ Won't implement — `*bool` kept by decision (anti-Verslimmbessern, 2026-07-07 plan)
~~7. **Add `svc.WithLogger(logger)` option** — Allow injecting a pre-configured logger instead of always creating one.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~8. **Add structured shutdown logging** — Log drain start, drain complete, shutdown start, shutdown complete with timestamps.~~ done or superseded in later waves (tags at v0.2.0+, per-module lint standard 2026-08-17, CHANGELOGs exist)
~~9. **Document the httputil.Server NON-delegation** — Plan says "delegate to httputil.Server" but we CAN'T because httputil.Server uses ListenAndServe() internally (no listener access). appkit owns http.Server + net.Listener directly. This deviation from the plan is correct but undocumented.~~ superseded 2026-09-16 — httputil v1.1.x Server exposes ListenerAddr/StartTLS; composition is the recommended refactor (AGENTS.md)

---

## f) Next 25 Tasks (Sorted by Impact ÷ Effort)

| #  | Task | Impact                                                 | Effort   | Priority |
| -- | ---- | ------------------------------------------------------ | -------- | -------- |
| ~~ | 1    | Rewrite README.md for framework API                    | Critical | 30m      |
| ~~ | 2    | Create example/main.go (12-line service)               | High     | 15m      |
| ~~ | 3    | Update AGENTS.md (architecture, file map)              | High     | 20m      |
| ~~ | 4    | Fix test HTTP requests (noctx lint)                    | Medium   | 15m      |
| ~~ | 5    | Add middleware test: panic → 500                       | High     | 15m      |
| ~~ | 6    | Add middleware test: X-Request-ID present              | Medium   | 10m      |
| ~~ | 7    | Add middleware test: logging captured                  | Medium   | 15m      |
| ~~ | 8    | Add httpspec.Run conformance test                      | Medium   | 15m      |
| ~~ | 9    | Add ServiceConfig_test.go (table-driven validation)    | Medium   | 15m      |
| ~~ | 10   | Fix DrainDelay test helper (near-zero default)         | Low      | 10m      |
| ~~ | 11   | Tag v1.0.0 (after README + example)                    | Critical | 5m       |
| ~~ | 12   | Consolidate error-family docs (DRY)                    | Low      | 20m      |
| ~~ | 13   | Update planning docs version numbers (v0.5.0)          | Low      | 10m      |
| ~~ | 14   | Add `WithLogger(logger)` option                        | Low      | 15m      |
| ~~ | 15   | Add structured shutdown logging                        | Low      | 15m      |
| ~~ | 16   | Document httputil.Server non-delegation decision       | Medium   | 10m      |
| ~~ | 17   | Create go.work workspace file                          | Medium   | 10m      |
| ~~ | 18   | Create cqrs/go.mod + EventService stub                 | Medium   | 30m      |
| ~~ | 19   | Implement cqrs EventService (stack/sqlite.New wrapper) | Medium   | 45m      |
| ~~ | 20   | cqrs: Service integration (Shutdown calls es.Shutdown) | Medium   | 30m      |
| ~~ | 21   | cqrs: E2E test (command → event → projection → health) | High     | 60m      |
| ~~ | 22   | Create docs/go.mod + catalog wrapper                   | Low      | 30m      |
| ~~ | 23   | Implement docs RegisterDocs (catalog routes)           | Low      | 45m      |
| ~~ | 24   | Create flake.nix (build/lint/test automation)          | Medium   | 30m      |
| ~~ | 25   | Final review: brutal-self-review skill                 | Medium   | 30m      |

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
