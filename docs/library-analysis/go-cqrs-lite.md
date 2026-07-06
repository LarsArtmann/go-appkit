# go-cqrs-lite — Integration Analysis

> **Verdict: 🟩 DO NOT DEPEND ON IT (wrong layer).** go-cqrs-lite is a CQRS + Event Sourcing
> **domain-pattern** library. go-appkit is **infrastructure glue**. For appkit to import it inverts
> the dependency direction. They coexist in an application that uses _both_ — appkit below,
> cqrs-lite above — not one importing the other.

---

## 1. Library identity

| Attribute      | Value                                                                                    |
| -------------- | ---------------------------------------------------------------------------------------- |
| Module path    | `github.com/larsartmann/go-cqrs-lite` (workspace root)                                   |
| Sub-modules    | **47** independent modules on `/v3` import paths (e.g. `.../event/v3`, `.../command/v3`) |
| Go version     | 1.26.4 (root), 1.26.3 (sub-modules)                                                      |
| Latest release | **v3.6.0** (2026-07-05); "v3.0.0 released — 84–100% coverage on core"                    |
| License        | MIT                                                                                      |
| Build          | Nix-based; uses Go experiment tags `goexperiment.arenas goexperiment.jsonv2`             |
| Sibling deps   | **go-branded-id v0.3.1**, **go-error-family v0.5.1**                                     |

---

## 2. What it provides (overview only — appkit will not consume it)

A lightweight **CQRS + Event Sourcing library (not framework)**: immutable events, branded ULID IDs,
a pure-function aggregate (Decider) pattern, optimistic-concurrency event stores, command/query
dispatch with middleware, projections and read models, plus production primitives (event signing,
payload encryption, schema evolution via upcasters, OTel tracing/metrics, Prometheus bridge, DLQ,
durable scheduling, idempotency).

47 independent modules so consumers import only what they need. `stack/` presets wire zero-config
defaults. Examples: `event`, `command`, `query`, `decider`, `deriver`, `id`, `idempotency`,
`dispatcher`, `codec` (CBOR-recommended), `schema`, `storage/{memory,sql,pebble,turso}`, `kv`,
`snapshot`, `graph`, `middleware`, `signing`, `encryption`, `listing`, `projection`,
`projectionhost`, `scenario`, `scheduling`, `otel`, `prometheus`, `watermill`,
`transport/{http,grpc}`, `catalog`, plus `stack/*` presets and `cmd/*` codegen tools.

---

## 3. Current usage in go-appkit

**Zero.** Not in `go.mod`; no imports; name appears only in this analysis folder.

---

## 4. Applicability assessment — why "wrong layer" is the right verdict

appkit is a **skeleton for Go services**: server lifecycle, health, logging, SQLite open,
signal-based shutdown. It owns **no domain model, no aggregates, no events, no commands**. cqrs-lite
is a **domain-pattern library** that an _application_ adopts to structure its business logic.

The dependency graph an application should form is:

```
   ┌───────────────────────────────┐
   │   Application (main package)   │
   └───────────────┬───────────────┘
            uses both, independently
        ┌──────────┴──────────┐
        ▼                     ▼
  go-cqrs-lite          go-appkit  (← and httputil)
  (domain layer)        (infra layer)
```

For appkit to `import` cqrs-lite would force **every consumer of appkit** (even non-CQRS services) to
pull the entire CQRS/ES module tree — violating appkit's "small, opinionated skeleton" promise. It
would also invert the layering: infrastructure must not depend on a domain pattern.

### The only conceivable overlaps — and why they still don't justify a dependency

| cqrs-lite module              | Tempting because…           | Why appkit still shouldn't                                                                              |
| ----------------------------- | --------------------------- | ------------------------------------------------------------------------------------------------------- |
| `transport/http` (SSE broker) | appkit has an HTTP server   | SSE is a domain-delivery concern, not infra. Consumers that want SSE import cqrs-lite themselves.       |
| `otel`, `prometheus`          | observability helpers       | OTel/Prometheus belong to the application or to httputil's `Metrics` middleware, not a server-skeleton. |
| `command`/`query` middleware  | "middleware" naming overlap | That's CQRS dispatch middleware; unrelated to HTTP middleware.                                          |
| `id` (branded ULIDs)          | IDs sound generic           | Domain IDs have no place in infra glue (see [go-branded-id.md](./go-branded-id.md)).                    |

---

## 5. Integration analysis — if someone mistakenly tried

- **Cost:** catastrophic for appkit's footprint. Even the leanest cqrs-lite module (`event/v3`)
  pulls `oklog/ulid`, `cbor`, branded-id, error-family. Importing the workspace root would drag in
  Pebble, Postgres drivers, Watermill, OTel SDK, gRPC… — the opposite of "skeleton."
- **Correctness:** a hard architecture violation. appkit would gain a transitive dependency on
  domain concepts it has no business knowing about.
- **Direction:** wrong. cqrs-lite is _above_ infra; appkit is infra.

---

## 6. What "fully" would mean here

Nothing sensible. "appkit uses cqrs-lite fully" is a category error — appkit would have to _become_ a
CQRS application framework, which is precisely what cqrs-htmx already is (see [cqrs-htmx.md](./cqrs-htmx.md)).
Two products should not collapse into one.

---

## 7. Recommendation

- **Do not add go-cqrs-lite to appkit.** It belongs in the consuming application, layered above
  appkit.
- If a consuming service wants both, the wiring is in its `main`: `appkit.NewServer(...)` for the
  HTTP skeleton, `stack/sqlite.New(...)` (or similar) for the CQRS/ES stack — side by side, no
  import between the libraries.
- The **only** legitimate touchpoint is documentation: appkit's README could _mention_ that
  go-cqrs-lite is the recommended domain layer for services built on appkit. That is a pointer, not
  a dependency.

---

## 8. Summary

- **Using today?** No.
- **Fully?** No — and never should be. Wrong layer; inverted dependency.
- **Action:** None on the code. Optional: a one-line pointer in appkit's README toward cqrs-lite as
  the domain layer for CQRS services.
