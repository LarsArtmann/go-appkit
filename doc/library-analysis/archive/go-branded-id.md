# go-branded-id — Integration Analysis

> **Verdict: 🟩 NOT APPLICABLE (no surface).** go-branded-id prevents mixing entity IDs via phantom
> types. go-appkit has **no entities, no ID fields, no domain types** — there is literally nothing
> to brand. Adopting it would be a solution without a problem.

---

## 1. Library identity

| Attribute      | Value                                                                             |
| -------------- | --------------------------------------------------------------------------------- |
| Module path    | `github.com/larsartmann/go-branded-id`                                            |
| Package        | single flat package `id` at repo root (no `internal/`, no `pkg/`)                 |
| Go version     | 1.26.4                                                                            |
| Latest release | **v0.3.1** (2026-06-17)                                                           |
| License        | MIT                                                                               |
| Deps           | **zero** — stdlib only (`cmp`, `encoding/*`, `database/sql`, `fmt`, `strconv`, …) |

A compile-time guard against mixing IDs of different entities. Plain `string`/`int64` IDs accept any
value — `GetOrder(userID)` compiles and silently bugs. Phantom types make `ID[UserBrand, string]`
and `ID[OrderBrand, string]` distinct, incompatible types; a cross-brand assignment is a compile
error.

---

## 2. What it provides (for completeness)

- Core type: `type ID[B any, V comparable] struct{ value V }` — `B` is a phantom brand, `V` the
  underlying value.
- Full serialization for `string` + 10 numeric types: JSON, SQL (`Scanner`/`Valuer`), Text
  (XML/TOML), Binary (little-endian), Gob. Core ops only for other comparable types.
- Constructors/helpers: `NewID`, `FromPtr`, `Ptr`, `BrandName[B]()`, `ValidateID`/`MustValidateID`,
  `Equal`, `Compare` (ordered types; `ErrNotOrdered` otherwise), `Or`, `Reset`.
- Methods: `Get() V`, `IsZero() bool`, `String()` (brand-aware display, `"Brand:value"`),
  `GoString()`, `Format` (`fmt.Formatter`).
- Opt-in `BrandNamer` interface on brand types: `func (UserBrand) Name() string { return "User" }`.
- `cmd/namer` codemod: finds brand types missing `Name()` (AST scan, dry-run by default).
- Downstream: 14 ecosystem repos depend on it; cqrs-lite/cqrs-htmx use it for domain IDs.

---

## 3. Current usage in go-appkit

**Zero.** Not in `go.mod`; no imports; no ID types in appkit's API.

appkit's public surface has **no identifiers at all**:

- `server.go` — `Server`, `ServerConfig` (Port, timeouts, handler, bool).
- `health.go` — `HealthStatus`, handlers.
- `shutdown.go` — `WaitForSignal`, `ShutdownConfig`.
- `logger.go` — `LogLevel`, `LogFormat`, `LoggerConfig`.
- `sqlite.go` — `SQLiteConfig` (Path string, conns, pragmas), `OpenSQLite`.

No struct field in appkit holds an ID. No function takes or returns an ID. There is nothing to brand.

---

## 4. Applicability assessment

Branded IDs solve a **domain-modeling** problem: preventing `userID` and `orderID` from being
swapped at call sites. That problem exists **inside domain types** (entities, value objects,
command/query payloads). appkit defines none of those.

| Hypothetical use in appkit           | Reality                                                                                                                                                                                                    |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Brand a "Server ID"?                 | Servers don't have IDs; they have addresses. No mixing risk.                                                                                                                                               |
| Brand a "DB connection ID"?          | `*sql.DB` is a pointer, not an ID.                                                                                                                                                                         |
| Brand config keys?                   | Keys are strings validated by allowlists, not identities.                                                                                                                                                  |
| Brand request IDs / correlation IDs? | That's a request-tracing concern — and it lives in httputil's `RequestID` middleware or cqrs-lite's `id` module, both of which already use go-branded-id internally. appkit itself has no tracing surface. |

There is no field in appkit where a phantom-branded type would prevent a bug. Adding it would be
pure ceremony.

---

## 5. Integration analysis — dependency direction & cost

- **Cost would be zero** (the lib is stdlib-only) — but zero cost does not justify zero value.
- **Direction is technically fine** (low-level lib can depend on a lower-level stdlib-only lib), but
  "can" ≠ "should."
- **Consistency argument is the only pull:** if appkit ever grew an ID-bearing concept, using the
  ecosystem's branded-id lib would keep conventions uniform. But that's hypothetical; YAGNI applies.

---

## 6. What "fully" would mean here

Nothing. There is no surface to adopt. "appkit uses go-branded-id fully" has no referent.

---

## 7. Recommendation

- **Do not add go-branded-id to appkit.** No IDs exist; none are planned in scope.
- The library is correctly consumed by cqrs-lite (`id.Of[T]`) and cqrs-htmx (`ActorID`,
  `ImpersonatorID`) — the domain layers where IDs actually live.
- Revisit **only if** appkit gains a first-class identity concept (e.g. a built-in request/correlation
  ID propagated through context). At that point, prefer re-using httputil's `RequestID` over
  re-implementing branding in appkit.

---

## 8. Summary

- **Using today?** No.
- **Fully?** No — not applicable. appkit has no IDs to brand.
- **Action:** None. The library is correctly placed in the domain layer (cqrs-lite, cqrs-htmx), not
  the infra layer (appkit).
