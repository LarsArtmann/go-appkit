# Library Analysis: Is go-appkit Using These Libraries Fully?

> **Scope:** go-appkit (`github.com/larsartmann/go-appkit`) vs. six sibling repositories.
> **Date:** 2026-07-05. **go-appkit version:** current `master`, Go 1.26.4.

---

## TL;DR — The short answer

**go-appkit currently uses NONE of the six libraries.** Its only dependency is `modernc.org/sqlite`
(see `go.mod`). A `grep` for `cqrs|branded|error-family|cmdguard|httputil|larsartmann` across the
source tree returns zero import hits — only the module's own path and docs references.

The more useful question is not "are we using them?" (clearly no) but **"should appkit use them,
and is 'fully' even the right target for a low-level skeleton library?"** The honest, layer-aware
answer differs sharply per library:

| Library             | Using? | Fully? | Architectural fit for **appkit**                                                                                                    | Verdict                              |
| ------------------- | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| **httputil**        | ❌ No  | —      | **Massive duplication.** appkit's `server.go` + `health.go` are a strict subset of httputil's `Server` + `Health`.                  | 🟥 **Act.** De-duplicate or compose. |
| **go-error-family** | ❌ No  | —      | **Good fit, optional.** appkit produces real errors (SQLite, listen, PRAGMA) that would benefit from classification + HTTP mapping. | 🟧 **Consider adopting.**            |
| **go-cqrs-lite**    | ❌ No  | —      | **Wrong layer.** CQRS/ES is a domain pattern; appkit is infrastructure. Inverted dependency.                                        | 🟩 **Do not depend on.**             |
| **cqrs-htmx**       | ❌ No  | —      | **Wrong layer + inverted.** A high-level app framework that itself depends on httputil + cqrs-lite. appkit must never depend on it. | 🟩 **Do not depend on.**             |
| **go-branded-id**   | ❌ No  | —      | **No surface.** appkit has no entity IDs, no domain types. Nothing to brand.                                                        | 🟩 **Not applicable.**               |
| **cmdguard**        | ❌ No  | —      | **Wrong layer.** CLI concerns belong in `main`, not a server library.                                                               | 🟩 **Not applicable.**               |

**Key finding:** The only genuinely actionable overlap is **httputil**. go-appkit is currently
re-implementing a subset of httputil's `Server` and `Health` abstractions. That is a real split-brain
that should be resolved. Everything else is either wrong-layer, inverted-dependency, or no-surface.

---

## What go-appkit actually is (so the analysis is grounded)

A tiny single-package library (`package appkit`) providing a **shared application skeleton** for Go
services. Five files, ~1190 lines including tests:

| File          | Concern               | What it does                                                                                                                                         |
| ------------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `server.go`   | HTTP server lifecycle | `Server` wraps `http.Server`; binds port in `Start()`; optional `GET /health` registration; graceful `Shutdown()`; thread-safe `Addr()`/`Running()`. |
| `health.go`   | Health endpoint       | `HealthStatus` typed string, `DefaultHealthHandler`, `NewHealthHandler(status)`, shared `writeHealthResponse` (JSON).                                |
| `shutdown.go` | Signal handling       | `WaitForSignal(ctx, cfg, onShutdown)` for SIGINT/SIGTERM, logs via `slog`.                                                                           |
| `logger.go`   | Logging init          | `LogLevel`/`LogFormat` typed strings, `InitLogger` returns `(*slog.Logger, error)`.                                                                  |
| `sqlite.go`   | DB bootstrap          | `OpenSQLite(ctx, cfg)` with WAL pragmas, PRAGMA-key allowlist for injection safety.                                                                  |

It is **infrastructure glue**, explicitly _not_ a domain/CQRS/UI/CLI library. That position in the
stack is the lens for every verdict below.

---

## The six questions, answered

### 1. go-cqrs-lite — using it fully? **No, and it would be wrong to.**

go-cqrs-lite (v3.6.0, 47 modules) is a CQRS + Event Sourcing domain-pattern library. appkit is
infrastructure. For appkit to `import` go-cqrs-lite would invert the dependency direction: cqrs-lite
sits _above_ infrastructure, consumed by applications. Overlap is limited to tangential shared
concerns (SSE broker, OTel helpers) that appkit has no business owning. See
[go-cqrs-lite.md](./go-cqrs-lite.md).

### 2. go-branded-id — using it fully? **No, and there is nothing to use it for.**

go-branded-id (v0.3.1, zero deps) prevents mixing entity IDs via phantom types. appkit has **no
entities, no ID fields, no domain types** — there is literally nothing to brand. Adopting it would be
a solution without a problem. See [go-branded-id.md](./go-branded-id.md).

### 3. go-error-family — using it fully? **No. Could partially adopt with real benefit.**

go-error-family (v0.6.1, zero-dep root) classifies errors into 5 families and maps them to HTTP
status / exit codes / retry policy. appkit's errors are currently plain `fmt.Errorf` + sentinels.
Adopting classification for SQLite/listen/PRAGMA errors and offering an error-family-aware health
handler is a legitimate, scoped improvement — not "fully," but meaningfully. See
[go-error-family.md](./go-error-family.md).

### 4. cqrs-htmx — using it fully? **No, and never should.**

cqrs-htmx (v4.1.1) is a CQRS+HTMX+Casbin+templ application framework. It **already depends on
httputil** and cqrs-lite. It is a consumer of the infrastructure layer, not a dependency of it.
appkit importing it would create a cycle in spirit and an inverted layer in fact. See
[cqrs-htmx.md](./cqrs-htmx.md).

### 5. cmdguard — using it fully? **No. Wrong layer (CLI, not server library).**

cmdguard (v2.10.2) is a type-safe Cobra wrapper for building CLIs. appkit exposes no commands; CLI
concerns live in `main` packages of consuming services. Not appkit's job. See [cmdguard.md](./cmdguard.md).

### 6. httputil — using it fully? **No. This is the real problem.**

httputil (v0.4.0) is an HTTP middleware/utility kit whose **`Server` and `Health` abstractions
overlap almost 1:1 with appkit's `server.go` and `health.go`**, plus 13 production middlewares appkit
lacks entirely. This is an active duplication / split-brain, not a hypothetical integration. See
[httputil.md](./httputil.md) — the most important file in this folder.

---

## How to read this folder

- Each library has its own file with: identity, what it provides, current usage (evidence),
  applicability assessment, integration analysis, a concrete verdict, and (where relevant) a
  step-by-step adoption plan with code sketches.
- Files are written from **appkit's perspective** as a low-level skeleton. "Fully" is evaluated
  against what would actually make sense for such a library — not as a blanket "import everything."

## Files

- [go-cqrs-lite.md](./go-cqrs-lite.md)
- [go-branded-id.md](./go-branded-id.md)
- [go-error-family.md](./go-error-family.md)
- [cqrs-htmx.md](./cqrs-htmx.md)
- [cmdguard.md](./cmdguard.md)
- [httputil.md](./httputil.md) ← **start here**
