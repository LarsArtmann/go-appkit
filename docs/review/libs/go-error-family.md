# go-error-family — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). Supersedes the
> go-error-family rows of the archived `docs/library-analysis/` assessments (stale 2026-07-05
> verdicts, already marked superseded there) and narrows the 2026-09-16 ecosystem deep-dive.

## Verdict

**USED — a core dependency in 11 of 12 modules at the latest tag (v0.12.0), adopted as the
repo-wide error protocol since day one. But appkit touches only the constructor/wrap/
HTTP-status half of the library: the dynamic-classification half (`Classify`, registries,
sentinel registrations), the context-builder half, the CLI half (`ExitCode`,
`HandleErrorWithConfig`), and the `errorfamilytest` asserts are all at zero adoption.**
Score **46/100 (low)** — the number reads harsher than reality because 6 rows are honestly
N/A (errorpages deliberately supersedes the library's HTTP/CLI rendering), but the
classification-registry gap is real and cheap to close.

| Fact                    | Value                                                                                                                        | Evidence                                      |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| In-graph version        | v0.12.0 (= latest tag, 2026-10-14)                                                                                           | `go.mod:7` + 10 satellite go.mods             |
| Direct importers        | root, cqrs, docs, errorpages, flightrecorder, flightrecorderhealth, health, integration, otel, systemd; indirect in realtime | `go.mod` files                                |
| Runtime symbols touched | 21 exported; zero registry/classify calls                                                                                    | `rg 'errorfamily\.[A-Z]' --glob '!*_test.go'` |
| Note                    | AGENTS.md "Core Dependencies" table says v0.10.1 — stale vs go.mod v0.12.0                                                   | AGENTS.md:67 vs `go.mod:7`                    |

## 1. Usage evidence

**Core re-exports** — `HTTPStatus` and `LogError` are re-exported as `appkit.HTTPStatus` /
`appkit.LogError` for consumers (`errors.go:13,20`).

**Wrap-at-source discipline** — every module classifies errors at creation via family
constructors instead of classifying at boundaries: `WrapInfrastructuref`/
`WrapInfrastructure` through the lifecycle (`service.go:111,143,235,301`),
`systemd/notify.go:78,103`, `testkit/testkit.go:238-258`; `NewInfrastructure` in otel
(`otel/setup.go:164-169`, `otel/otlp.go:202`); `NewRejection` in metrics (`metrics.go:92`)
and errorpages (`errorpages/errorpages.go:159`); `Newf(Rejection, …)` config validation in
core (`config.go:201-211`) and health (`health/mount.go:62-63,93-94,165-166`).

**CLI boundary** — `HandleError` used in example mains only (`systemd/example/main.go:42`,
`otel/example/main.go:62`) — correct: the repo ships libraries, not CLIs.

**What is NOT called anywhere (code, not comments):** `Classify`, `RegisterClassification(s)`,
`RegisterClassifier(ies)`, `RegisterStdlibDefaults`, `Registry`, templates
(`RegisterTemplate`/`TemplateForCode`/`MessageTemplate`), `HTTPHandler`, `ExitCode`,
`HandleErrorWithConfig`/`HandleErrorDetailed`, `Compose`, `WrapOnce`, all `WithContext*`
builders, `RetryPolicy`, `LogErrorContext`, `errorfamilytest`, and the
bridge/diagnose/agent submodules.

## 2. Capability inventory vs adoption

| #  | go-error-family capability                                                | Status            | Evidence / gap meaning                                                                                                                                                                                                                                                                                   |
| -- | ------------------------------------------------------------------------- | ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Family constructors: Rejection / Infrastructure / Transient               | 🟢 Full           | wrap-at-source discipline, 20+ sites                                                                                                                                                                                                                                                                     |
| 2  | Constructors: Conflict / Corruption / Orchestration                       | 🟡 Partial        | test-only (`NewConflict`/`NewCorruption` in tests); runtime never produces them — arguably correct for an HTTP framework's error population                                                                                                                                                              |
| 3  | `New`/`Newf` generic constructors                                         | 🟢 Full           | `config.go:201`, `health/mount.go:62`, `errorpages:176`                                                                                                                                                                                                                                                  |
| 4  | Wrap family (`Wrap`, `WrapInfrastructure(f)`, `WrapRejectionf`)           | 🟢 Full           | `service.go:111-301`, `systemd/notify.go`, cqrs/eventservice                                                                                                                                                                                                                                             |
| 5  | `HTTPStatus` mapping                                                      | 🟢 Full           | re-exported `errors.go:13`                                                                                                                                                                                                                                                                               |
| 6  | `LogError` / `LogErrorContext`                                            | 🟡 Partial        | `LogError` re-exported (`errors.go:20`); `LogErrorContext` (trace correlation) unused — otel correlation is done differently                                                                                                                                                                             |
| 7  | `HandleError` CLI exit pipeline                                           | 🟡 Partial        | examples only; no shipped CLI (correct scope, partial adoption)                                                                                                                                                                                                                                          |
| 8  | `Classify` dynamic classification                                         | 🔴 Missed         | zero calls. Appkit's wrap-at-source style mostly doesn't need it, but boundaries that receive FOREIGN errors — errorpages rendering arbitrary handler errors (unclassified → 500 by contract), health probes wrapping third-party check errors — could upgrade unclassified errors instead of defaulting |
| 9  | Registry API (sentinel/classifier registration, `RegisterStdlibDefaults`) | 🔴 Missed         | zero registrations. e.g. `http.MaxBytesError` (security's BodyLimit contract) and `context.Canceled` are handled ad hoc per module instead of one fleet-default classification                                                                                                                           |
| 10 | WrapOnce                                                                  | 🔴 Missed (small) | idempotent-wrap sites exist by construction: `Shutdown` is idempotent (`shutdown.go`) and re-wraps hook errors on every call                                                                                                                                                                             |
| 11 | `WithContext*` builders                                                   | 🔴 Missed         | zero use. Context-rich errors would flow into `LogError` fields and otel exception capture for free                                                                                                                                                                                                      |
| 12 | `errorfamilytest` asserts                                                 | 🔴 Missed         | zero imports; tests hand-roll family assertions (`otel/setup_test.go:302-322`, `systemd/notify_test.go:173`)                                                                                                                                                                                             |
| 13 | Templates (`What/Why/Fix/WayOut`)                                         | ⚪ N/A            | errorpages is the rich rendering layer (HTML/JSON via templ-components); templates target plain-text CLI output the framework doesn't emit                                                                                                                                                               |
| 14 | `HTTPHandler` (JSON error handler)                                        | ⚪ N/A            | superseded BY errorpages — the deliberate architecture                                                                                                                                                                                                                                                   |
| 15 | `ExitCode` / sysexits mapping                                             | ⚪ N/A            | no CLI binaries in the repo; examples don't need the mapping                                                                                                                                                                                                                                             |
| 16 | `Compose`                                                                 | ⚪ N/A            | alias of `errors.Join`; the shutdown path already joins errors directly                                                                                                                                                                                                                                  |
| 17 | `RetryPolicy`                                                             | ⚪ N/A            | no retry loops in the framework; go-retry is its consumer, not appkit                                                                                                                                                                                                                                    |
| 18 | submodules: diagnose / agent / bridge                                     | ⚪ N/A            | debug-agent/diagnostic-rule class features, explicitly app-side (bridge README positioning)                                                                                                                                                                                                              |

**Score derivation**: A = 12 applicable (13-18 N/A with reasons). Green = 4 (rows 1, 3, 4, 5).
Partial = 3 × 0.5 = 1.5 (rows 2, 6, 7). Missed = 5 (rows 8, 9, 10, 11, 12).
**Score = (4 + 1.5) / 12 = 46/100 (low band).**

## 3. Findings (Pareto order)

### F1 — The registry is the missing half (one-time, fleet-wide payoff)

Wrap-at-source covers appkit's own errors; it does nothing for errors appkit RECEIVES:
third-party probe errors in health, handler errors rendered by errorpages, store errors in
cqrs. Today every such site either default-classifies (Transient fail-open) or hand-maps.
A single `appkit.RegisterFleetClassifications()` (calling `RegisterStdlibDefaults` plus
`RegisterClassification` for `http.MaxBytesError` → Rejection, `context.Canceled` →
Infrastructure/ignore-class, `sql.ErrNoRows` → Rejection) would make `Classify` meaningful
at those boundaries and centralize decisions currently duplicated per module. Additions-only,
v1-shaped, demand-gated per house rules — but write it down as the adoption path, because
"register at boot" is the pattern the library is designed around and appkit never performs it.

### F2 — `WrapOnce` at idempotent seams (small, free)

`Service.Shutdown` is documented idempotent (`service.go` contract) and re-runs
`WrapInfrastructuref` over hook errors. `WrapOnce` exists precisely for re-executed wrap
sites. No behavior bug today (double-wrap is cosmetic), but it is the textbook use.

### F3 — `WithContext` + `errorfamilytest`: DX upgrades, zero-risk adoption

Context-carrying errors would enrich `LogError` output and the otel exception capture
(otel already indexes exception events); `errorfamilytest`'s seven asserts would replace the
hand-rolled family assertions in otel/systemd tests. Both are mechanical, additions-only.

### F4 — AGENTS.md table drift (fix on sight, done in this audit's memory pass)

AGENTS.md:65/67 still say httputil v1.4.0 / go-error-family v0.10.1 while go.mod carries
v1.7.0 / v0.12.0. Updated in this audit.

## 4. Should appkit use it MORE? (the explicit answer)

Yes — the registry/classification half (F1) is the real gap, with WrapOnce/WithContext/
errorfamilytest as cheap add-ons (F2/F3). No — for `HTTPHandler`, templates, and `ExitCode`:
errorpages owns HTTP rendering by design, and the repo ships no CLI; forcing those in would
duplicate the framework's own layers.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -o 'errorfamily\.[A-Z][A-Za-z]*' --glob '*.go' --glob '!*_test.go' | sed 's/.*://' | sort -u
rg -n 'errorfamily\.(Classify|RegisterClassif|RegisterTemplate|HTTPHandler|ExitCode|WrapOnce|WithContext)' \
  --glob '*.go' --glob '!*_test.go' | grep -v '//'   # 1 hit → registry row landed, rest stand
grep 'go-error-family' go.mod   # v0.12.0
```

## Errata (2026-11-06, re-run of the verify block)

- **The symbol-list command had the same `-o` truncation bug as httputil.md**
  (`errorfamily\.[A-Z]` matched one character after the dot): fixed above to
  `[A-Z][A-Za-z]*`; 17 distinct runtime symbols in non-test files as of today.
- **F1's registry battery PARTIALLY landed 2026-11-03** (composition-mission
  train): `appkit.RegisterClassifications()` engages
  `errorfamily.RegisterClassificationType` (stdlib defaults +
  `*http.MaxBytesError` → Rejection; idempotent; deliberately not auto-called —
  global-state hygiene). The registry row moves from "missed" to "shipped";
  the remaining gap rows (`Classify`, `RegisterClassification(s)` beyond the
  battery, `RegisterTemplate`, `HTTPHandler`, `ExitCode`, `WrapOnce`,
  `WithContext`) still stand.
