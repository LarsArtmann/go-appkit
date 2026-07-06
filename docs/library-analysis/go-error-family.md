# go-error-family — Integration Analysis

> **Verdict: 🟧 CONSIDER ADOPTING (partial).** appkit produces real errors (SQLite open/PRAGMA,
> listen, encode) that currently sit as plain `fmt.Errorf` + untyped sentinels. Adopting
> go-error-family classification for these would add retry/HTTP/exit semantics at near-zero cost.
> "Fully" is not the right target — appkit has a small error surface — but a meaningful, scoped
> adoption is warranted.

---

## 1. Library identity

| Attribute                | Value                                                                           |
| ------------------------ | ------------------------------------------------------------------------------- |
| Module path              | `github.com/larsartmann/go-error-family`                                        |
| Go version               | 1.26.4 (uses Go 1.26 `errors.AsType[T]`)                                        |
| Latest release           | **v0.6.1** (2026-07-05)                                                         |
| Root deps                | **zero** (stdlib only) — genuinely zero-dependency                              |
| Experimental sub-modules | `agent/`, `bridge/` (samber/oops), `diagnose/` (+ `git`, `postgres`) — all v0.x |

A library that classifies errors into **5 families** and derives retry policy, HTTP status, exit
code, audience, and messaging from that single classification. Tagline: _"Share the protocol, not
the implementation."_

---

## 2. What it provides (the parts that matter to appkit)

### The 5-family taxonomy (`family.go`)

| Family           | Retryable | HTTP | Exit (BSD `sysexits.h`) | Audience |
| ---------------- | --------- | ---- | ----------------------- | -------- |
| `Rejection`      | no        | 400  | 1                       | User     |
| `Conflict`       | no        | 409  | 1                       | User     |
| `Transient`      | **yes**   | 503  | 75 (EX_TEMPFAIL)        | All      |
| `Corruption`     | no        | 500  | 65 (EX_DATAERR)         | Ops      |
| `Infrastructure` | no        | 503  | 69 (EX_UNAVAILABLE)     | Ops      |

Severity is a total order; multi-error (`errors.Join`) classification picks the **worst** by
severity, deterministically. Unknown → Transient (fail-open for retry). `Classify(nil)` → Rejection.

### The four interfaces (the real contract, `interfaces.go`)

```go
type Coded      interface { error; ErrorCode() string }
type Classified interface { error; ErrorFamily() Family }
type Contextual interface { error; ErrorContext() map[string]string }
type Retryable  interface { error; IsRetryable() bool }
```

`Error` is just a reference implementation; domain types implement only the interfaces they need.

### Functions appkit would use

- `Classify(err) Family`, `IsRetryable(err) bool`, `ExitCode(err) int`, `HTTPStatus(err) int`,
  `Code(err) string`, `ParseFamily(s)`.
- Constructors: `New`/`Newf`/`Wrap`/`Wrapf` + family-specific `New{Rejection,Conflict,Transient,Corruption,Infrastructure}`
  and `Wrap{...}`/`Wrap{...}f`.
- Registration: `RegisterClassification(sentinel, family)`, `RegisterClassifier(fn)` — for errors
  appkit does **not** own (e.g. `*sqlite.Error`, `net.OpError`).
- HTTP: `HTTPStatus(err)`, `HTTPHandler(func(http.ResponseWriter,*http.Request) error) http.Handler`
  (writes safe JSON, **never leaks `err.Error()`** — message only from registered templates).
- Logging: `LogError(err, logger)` / `LogErrorContext(ctx, err, logger)`.

### Registry (injectable, thread-safe)

`NewRegistry()` / `DefaultRegistry` / `Registry.Clone()` — lock-free hot path via `atomic.Pointer`
to immutable snapshots. `Clone()` enables test isolation without `t.Cleanup`.

---

## 3. Current usage in go-appkit

**Zero.** Evidence:

- Not in `go.mod`.
- appkit's errors are plain: `fmt.Errorf("...: %w", err)` and untyped sentinels
  (`errSQLitePathRequired`, `errPRAGMAAllowlist`, `errUnsupportedLogLevel`, `errUnsupportedLogFormat`).
- `health.go` hand-rolls a `HealthStatus.HTTPStatus()` switch (duplicating what error-family's
  `HTTPStatus(err)` generalizes).
- `health.go`'s `writeHealthResponse` writes a status string with `http.Error` on encode failure —
  no classification, no retry hints, leaks the raw error string on the 500 path.

---

## 4. Applicability assessment

appkit is infrastructure glue, and infrastructure glue is **exactly where error classification pays
off**: every error appkit raises maps cleanly onto a family:

| appkit error site                                | Current                           | Natural family                                                          | HTTP | Retry? |
| ------------------------------------------------ | --------------------------------- | ----------------------------------------------------------------------- | ---- | ------ |
| `OpenSQLite`: bad path                           | `errSQLitePathRequired` sentinel  | **Rejection**                                                           | 400  | no     |
| `OpenSQLite`: unsupported PRAGMA                 | `errPRAGMAAllowlist` sentinel     | **Rejection**                                                           | 400  | no     |
| `OpenSQLite`: `sql.Open` / `ExecContext` failure | `fmt.Errorf` wrapping driver err  | **Infrastructure** or **Transient** (via classifier on `*sqlite.Error`) | 503  | maybe  |
| `Server.Start`: listen failure                   | `fmt.Errorf("listen …: %w", err)` | **Infrastructure** (port in use) / **Transient**                        | 503  | maybe  |
| `InitLogger`: bad level/format                   | sentinel                          | **Rejection**                                                           | 400  | no     |
| health encode failure                            | `http.Error`                      | **Infrastructure**                                                      | 500  | no     |
| `Shutdown` failure                               | `fmt.Errorf`                      | **Infrastructure**                                                      | 503  | no     |

So the fit is real but **narrow**: appkit's public error surface is ~6 error sites. "Fully" adopting
error-family (agents, diagnostics, oops bridge, templates) would be vast overkill. The right scope
is: **classify appkit's own errors + optionally offer an error-family-aware health handler.**

---

## 5. Integration analysis — dependency direction & cost

- **Direction is correct & cheap.** error-family root has zero deps. Adding it costs appkit nothing
  beyond one import. (httputil already depends on it, so if appkit adopts httputil per
  [httputil.md](./httputil.md), error-family comes along automatically.)
- **Non-breaking adoption is possible.** Wrap existing sentinels with families via
  `RegisterClassification`; keep returning them as today. Consumers that don't care see no change;
  consumers that call `errorfamily.Classify(err)` get richer behavior.
- **Do NOT import experimental sub-modules** (`agent/`, `bridge/`, `diagnose/`). They are v0.x and
  pull in `samber/oops`. appkit should import only the zero-dep root.

---

## 6. Concrete adoption plan (scoped, non-breaking)

1. Add `github.com/larsartmann/go-error-family v0.6.1` (root only) to `go.mod`.
2. In a new `errors.go`, classify appkit's existing sentinels at package init:
   ```go
   func init() {
       errorfamily.RegisterClassifications(
           errorfamily.Sentinel{Err: errSQLitePathRequired, Family: errorfamily.Rejection},
           errorfamily.Sentinel{Err: errPRAGMAAllowlist, Family: errorfamily.Rejection},
           errorfamily.Sentinel{Err: errUnsupportedLogLevel, Family: errorfamily.Rejection},
           errorfamily.Sentinel{Err: errUnsupportedLogFormat, Family: errorfamily.Rejection},
       )
   }
   ```
   _(Check the exact `RegisterClassification`/`Sentinel` API signature against the v0.6.1 docs before
   committing — the research summary lists `RegisterClassification(sentinel, family)` and a batch
   `RegisterClassifications`.)_
3. Add a classifier for `*sqlite.Error` (and optionally `net.OpError`) so driver/runtime errors
   classify as Transient vs Infrastructure without appkit owning those types:
   ```go
   errorfamily.RegisterClassifier(func(err error) (errorfamily.Family, bool) {
       var sqliteErr *sqlite.Error
       if errors.As(err, &sqliteErr) {
           return errorfamily.Transient, true // busy/locked → retry
       }
       return 0, false
   })
   ```
4. Optionally replace `health.go`'s hand-rolled encode-failure `http.Error` with
   `errorfamily.LogError` + a classified response, and stop leaking the raw error string.
5. Optionally expose `appkit.HTTPStatus(err)` as a thin alias over `errorfamily.HTTPStatus` so
   consumers get server-error → status mapping for free.
6. Tests: assert families (`errorfamilytest.AssertFamily`) for each sentinel.

---

## 7. What "fully" would mean here

appkit using error-family "fully" would mean: **every error appkit returns is classified**, every
health-handler failure path is logged via `LogError`, and an `HTTPHandler`-style option is offered.
That is achievable for appkit's small surface — but importing `agent/`, `diagnose/`, or `bridge/`
would be overreach and wrong for an infra library. Stay on the zero-dep root.

---

## 8. Summary

- **Using today?** No.
- **Fully?** No — and shouldn't be (experimental sub-modules are out of scope). But classifying
  appkit's ~6 own error sites is a small, high-value, non-breaking win.
- **Action:** Optional. If adopted, scope strictly to the zero-dep root; classify sentinels + the
  SQLite driver error; stop leaking raw error strings from the health handler. If appkit also adopts
  httputil, error-family arrives transitively for free.
