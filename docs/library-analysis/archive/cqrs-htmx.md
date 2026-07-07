# cqrs-htmx — Integration Analysis

> **Verdict: 🟩 DO NOT DEPEND ON IT (wrong layer + inverted dependency).** cqrs-htmx is a
> high-level **application framework** that itself depends on httputil and cqrs-lite. go-appkit is
> low-level infrastructure. appkit importing cqrs-htmx would invert the stack and create a conceptual
> cycle. They are peers in an application that composes both — appkit for infra, cqrs-htmx for the
> app layer — never one beneath the other.

---

## 1. Library identity

| Attribute      | Value                                                                                                                           |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Module path    | `github.com/larsartmann/cqrs-htmx/v4`                                                                                           |
| Go version     | 1.26.4                                                                                                                          |
| Latest release | **v4.1.1**                                                                                                                      |
| License        | MIT                                                                                                                             |
| Kind           | **Library/SDK — NOT an application** (explicit AGENTS.md banner); `adminui/`, `usermgmt/`, `examples/` are separate sub-modules |
| Multi-module   | 8 Go modules (root, usermgmt, usermgmt/{totp,webauthn,oauth2}, adminui, integration_test, 4 examples)                           |

It wires together **go-cqrs-lite + HTMX + templ + Casbin** into a single, framework-agnostic
(`net/http`/Chi/Gin compatible) SDK: HTTP→CQRS decode, identity propagation, Casbin policy checks,
CQRS-error→HTTP mapping (HTMX-aware), partial vs full-page rendering, response headers, CSRF
(nosurf double-submit), rate limiting, security headers, SSE/WS streaming, and notifications.

---

## 2. What it provides (overview only — appkit will not consume it)

- **`App`** — central builder holding `*command.Dispatcher`, `*query.Dispatcher`, `Enforcer`,
  `UserIDExtractor`, `ErrorHandler`, hooks, timeout, maxBodySize.
- **`New(cfg Config)`** / **`MustNew`**; `App.Command(...)`, `App.Query(...)` → `http.HandlerFunc`.
- **`HandlerOption`** compositional API: `Authorize`, `RequireAuth`, `RequestGuard`, `CSRFProtect`,
  `DecodeJSON[T]`/`DecodeForm[T]`, `Validate*`, `Render*`/`RenderTempl*`, `Redirect`, `Trigger*`,
  `Notify*`, `OnError`, `Timeout`.
- Middlewares: `ContextEnrichmentMiddleware`, `HTMXMiddleware`, `SecurityHeadersMiddleware`,
  `CSRFMiddleware`, `RecoveryMiddleware`, `RateLimiterMiddleware`, `ServerTimingMiddleware`,
  `Chain`.
- ACK system (honest-UI command confirmation), idempotency, embedded HTMX v2 + SSE/WS/idiomorph.
- 5-family error model (via go-error-family); `ProblemDetailsErrorHandler` emits RFC 7807.

---

## 3. cqrs-htmx's dependencies — the smoking gun

From its `go.mod`:

| Dependency                                                                             | Version    |
| -------------------------------------------------------------------------------------- | ---------- |
| **go-cqrs-lite/{command,query,event,id,idempotency,storage/memory,transport/http}/v3** | v3.5.0     |
| **go-branded-id**                                                                      | v0.3.1     |
| **go-error-family** (indirect via event/v3)                                            | v0.6.0     |
| **httputil**                                                                           | **v0.4.0** |

**cqrs-htmx already depends on httputil.** That single fact settles the layering question: httputil
(and by extension appkit, which overlaps httputil) sits _below_ cqrs-htmx. For appkit to depend on
cqrs-htmx would be appkit depending on something that depends on appkit's own peer layer — a layered
inversion even if not a literal Go import cycle.

---

## 4. Current usage in go-appkit

**Zero.** Not in `go.mod`; no imports; name appears only in this analysis folder.

---

## 5. Applicability assessment — why this is a hard "no"

appkit is a **server skeleton**: bind a port, register `/health`, init a logger, open SQLite, handle
signals. cqrs-htmx is an **application framework**: decode commands, enforce RBAC, render templ,
stream SSE, manage CSRF. These are different elevations of the stack.

| cqrs-htmx concern                        | Belongs to…                                | appkit's role                                                            |
| ---------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------ |
| CQRS command/query dispatch              | the application                            | none                                                                     |
| Casbin authorization                     | the application                            | none                                                                     |
| templ rendering, HTMX response headers   | the application/presentation               | none                                                                     |
| CSRF, security headers                   | the application (httputil offers them too) | none                                                                     |
| `App.HealthHandler()`                    | overlaps appkit's health                   | appkit's health is the infra-level one; cqrs-htmx's is the app-level one |
| `App.Middleware()` (identity enrichment) | the application                            | none                                                                     |

Even the apparent overlap (health, middleware, security headers) resolves in httputil's favor: those
are httputil's job at the infra layer, and cqrs-htmx's job at the app layer. appkit should compose
httputil (see [httputil.md](./httputil.md)), not reach up to cqrs-htmx.

---

## 6. Integration analysis — if someone mistakenly tried

- **Cost:** enormous. cqrs-htmx pulls Casbin, nosurf, oklog/ulid, form decoder, rate limiter, the
  full cqrs-lite module set, templ (in adminui), and more. appkit would balloon from "tiny skeleton"
  to "depends on a full app framework."
- **Correctness:** a layering violation. Infrastructure cannot sit beneath a framework that consumes
  infrastructure.
- **Conceptual cycle:** cqrs-htmx → httputil → (peer of) appkit. Adding appkit → cqrs-htmx closes a
  loop the architecture explicitly forbids.

---

## 7. What "fully" would mean here

A category error. "appkit uses cqrs-htmx fully" would mean appkit stops being a skeleton and becomes
an opinionated HTMX/CQRS application — duplicating cqrs-htmx's purpose. Two products should not
collapse into one (see also [go-cqrs-lite.md](./go-cqrs-lite.md)).

---

## 8. Recommendation

- **Do not add cqrs-htmx to appkit.** It is a consumer-layer framework, not an infra dependency.
- The correct relationship is **composition in the application**: a service's `main` builds an
  `appkit.Server` (infra skeleton) and a `cqrs-htmx.App` (application framework) and mounts the latter's
  handlers on the former's mux. Side by side, no cross-import.
- Optional: a one-line pointer in appkit's README toward cqrs-htmx as the recommended application
  framework for HTMX/CQRS services built on appkit.

---

## 9. Summary

- **Using today?** No.
- **Fully?** No — and never should be. Wrong layer; cqrs-htmx already depends on httputil (appkit's
  peer), so the dependency direction is inverted.
- **Action:** None on the code. Optional README pointer toward cqrs-htmx as the app-layer companion.
