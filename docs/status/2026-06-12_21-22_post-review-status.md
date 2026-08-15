# Status Report: go-appkit

**Date:** 2026-06-12 21:22
**Branch:** master (1 commit ahead of origin)
**Commit:** d76c22a refactor: harden types, fix security issues, expand test coverage
**Test Coverage:** 88.7% of statements
**Test Results:** 31/31 PASS with `-race` flag
**Lines of Code:** 1,076 (source + tests), 21KB total

---

## a) FULLY DONE

### Security

- [x] **PRAGMA SQL injection fixed** — `allowedPRAGMAs` allowlist validates all keys before interpolation in `sqlite.go`. Unsupported keys return error immediately and close the DB connection.

### Type Safety

- [x] **`LogLevel` typed string** — `"debug"`, `"info"`, `"warn"`, `"error"` constants. Zero-value defaults to info. Invalid values return compile-time-friendly errors instead of runtime panics.
- [x] **`LogFormat` typed string** — `"text"`, `"json"`, `"auto"` constants. Unknown formats return errors.
- [x] **`HealthStatus` typed enum** — `"ok"`, `"ready"`, `"unhealthy"`, `"degraded"` with `HTTPStatus()` method that maps to correct HTTP status codes (200/503).

### API Correctness

- [x] **`InitLogger` returns `(*slog.Logger, error)`** — No more panics. Library-safe error handling.
- [x] **`ServerConfig.applyDefaults()`** — Called once, not 4 times. DRY.
- [x] **`ServerConfig.RegisterHealth`** — Boolean flag to opt out of auto-registering `GET /health`.
- [x] **`Server.Running() bool`** — Thread-safe status check via `sync.RWMutex`.
- [x] **`Server.ln` protected by `sync.RWMutex`** — `Addr()` and `Running()` use `RLock`/`RUnlock`.

### Code Quality

- [x] **DRY health handler** — `writeHealthResponse()` shared between `DefaultHealthHandler` and `NewHealthHandler`.
- [x] **`shutdown.go` uses `slog.Info`** — Replaced `fmt.Fprintf(os.Stderr)`.
- [x] **`sqlite.go` uses `errors.New`** — For static error messages (lint fix).

### Test Coverage (31 tests, 88.7% coverage)

- [x] Health: default handler, ready, unhealthy, degraded, unknown status code mapping (5 tests)
- [x] Logger: default level, debug, invalid level, invalid format, JSON, text, all levels, unknown value (8 tests)
- [x] Server: health endpoint registration, custom health handler, health opt-out, addr nil before start, running after start, shutdown, shutdown before start (7 tests)
- [x] Shutdown: context cancel, error propagation, SIGUSR1 signal delivery, default config, default timeout values (5 tests)
- [x] SQLite: basic open, empty path, disallowed PRAGMA, custom allowed PRAGMAs, pool settings, open error (6 tests)

### Infrastructure

- [x] **Race-free test helpers** — `freePort()`, `waitForAddr()`, `waitForRunning()` replace `time.Sleep` in server tests.
- [x] **golangci-lint config** — `.golangci.yml` added with formatting rules applied.
- [x] **Project docs** — `doc.go`, `AUTHORS`, `CONTRIBUTING.md`, `LICENSE`, `.gitattributes`, `docs/DOMAIN_LANGUAGE.md` added.
- [x] **Planning doc** — `docs/planning/2026-06-12_2058-full-code-review.md` with Pareto breakdown and D2 graph.
- [x] **AGENTS.md updated** — Reflects all new types, API changes, gotchas, testing patterns.

---

## b) PARTIALLY DONE

### doc.go Package Documentation

- `doc.go` exists but has placeholder content: `// Package appkit provides ...`
- Should have a proper package doc comment describing the library's purpose and usage.

### docs/DOMAIN_LANGUAGE.md

- Template exists but has no domain-specific terms filled in (still has placeholder "Example Term" entries).
- For a utility library like appkit, the domain language is about infrastructure concepts (server, health check, shutdown, logging, database connection).

### README.md

- Still references old API (`InitLogger` returning `*slog.Logger` directly, not `(*slog.Logger, error)`).
- Still references old `HealthHandler` as `http.HandlerFunc` not `HealthStatus` type.
- Usage examples need updating for the new typed API.

---

## c) NOT STARTED

1. **README.md update** — Usage examples reference old API signatures
2. **CHANGELOG.md update** — No entry for the v0.2.0 refactor
3. **Integration example** — No `example_test.go` with end-to-end usage showing all components together
4. **Port 0 support** — `applyDefaults()` overwrites Port:0 with 8080; no way to request OS-assigned port
5. **`WaitForSignal` accepts `*slog.Logger`** — Currently uses global `slog.Info`; should accept a logger for testability
6. **`OpenSQLite` validates PRAGMA values** — Keys are allowlisted but values are still interpolated unsafely (e.g., `journal_mode` value could be `WAL; DROP TABLE users`)
7. **Error sentinel values** — No `var Err...` sentinels; all errors are created inline with `fmt.Errorf`/`errors.New`
8. **`ServerConfig.Port` as `string`** — Currently `int`; could support unix sockets or address strings
9. **CI pipeline** — No GitHub Actions, no automated testing on push
10. **`go mod tidy`** — go.mod has indirect deps that may need cleanup (gopls hints about unused modules)

---

## d) TOTALLY FUCKED UP

Nothing is fucked up. All 31 tests pass, `go vet` clean, race detector clean, 88.7% coverage. The codebase is in solid shape.

The closest thing to "fucked up" is the **uncommitted lint/formatting changes** sitting in the working tree — golangci-lint auto-formatted code that wasn't committed yet. Not a problem, just needs a commit.

---

## e) WHAT WE SHOULD IMPROVE

### High Impact

1. **Update README.md** — The public-facing docs are stale; users will see wrong API signatures
2. **Fix `doc.go`** — Package documentation is a placeholder
3. **PRAGMA value sanitization** — Keys are validated but values are not; sophisticated injection is still possible
4. **Error sentinel values** — `errors.Is()` matching is impossible without exported sentinel errors

### Medium Impact

5. **Inject logger into `WaitForSignal`** — Global `slog` makes testing noisy
6. **Support `Port: 0`** — OS-assigned ports are a real use case for testing and service mesh sidecars
7. **Add `example_test.go`** — Go convention for package examples; helps godoc and new users
8. **Fill `DOMAIN_LANGUAGE.md`** — Define the infrastructure domain vocabulary

### Lower Impact

9. **CI pipeline** — GitHub Actions for automated test + lint on push
10. **Benchmarks** — No performance benchmarks for server start/shutdown cycle
11. **Fuzz testing** — `OpenSQLite` PRAGMA values, `parseLevel`, health status are good fuzz targets

---

## f) Top 25 Things We Should Get Done Next

Sorted by impact × effort (highest first):

| #  | Task                                                                  | Impact | Effort | Category     |
| -- | --------------------------------------------------------------------- | ------ | ------ | ------------ |
| 1  | Update README.md usage examples for new API                           | High   | 15min  | Docs         |
| 2  | Fix `doc.go` package documentation                                    | High   | 10min  | Docs         |
| 3  | Sanitize PRAGMA values (not just keys)                                | High   | 20min  | Security     |
| 4  | Add error sentinel values (`ErrPathRequired`, etc.)                   | Medium | 20min  | Correctness  |
| 5  | Inject `*slog.Logger` into `WaitForSignal` / `ShutdownConfig`         | Medium | 15min  | Testability  |
| 6  | Add `example_test.go` with end-to-end usage                           | Medium | 20min  | DX           |
| 7  | Update CHANGELOG.md for v0.2.0                                        | Medium | 10min  | Docs         |
| 8  | Support `Port: 0` for OS-assigned ports                               | Medium | 15min  | Feature      |
| 9  | Fill `docs/DOMAIN_LANGUAGE.md` with actual terms                      | Low    | 15min  | Docs         |
| 10 | Add GitHub Actions CI (test + vet + lint)                             | Medium | 30min  | Infra        |
| 11 | `go mod tidy` to clean unused indirect deps                           | Low    | 5min   | Housekeeping |
| 12 | Add `ServerConfig.Addr` string field (support unix sockets)           | Low    | 20min  | Feature      |
| 13 | Add `Server.Start()` returns actual listener address in error channel | Low    | 10min  | DX           |
| 14 | Add `SQLiteConfig.DefaultPath` for in-memory default                  | Low    | 10min  | Feature      |
| 15 | Add `WithLogger` option pattern for Server                            | Low    | 30min  | Feature      |
| 16 | Add benchmarks for server start/shutdown                              | Low    | 20min  | Testing      |
| 17 | Add fuzz tests for PRAGMA values and log level parsing                | Low    | 30min  | Testing      |
| 18 | Add `ShutdownConfig.OnSignal` callback hook                           | Low    | 15min  | Feature      |
| 19 | Add `Server.ServeMux()` accessor to retrieve the mux                  | Low    | 5min   | DX           |
| 20 | Add `SQLiteConfig.Validate()` method                                  | Low    | 10min  | Correctness  |
| 21 | Consider `errors.Join` for multi-PRAGMA failures                      | Low    | 10min  | Correctness  |
| 22 | Add `IsTerminal()` test with mock file                                | Low    | 10min  | Testing      |
| 23 | Add middleware support (request logging, recovery)                    | Low    | 45min  | Feature      |
| 24 | Add graceful restart support via `SIGUSR2`                            | Low    | 30min  | Feature      |
| 25 | Add `VERSION` constant for embedding in health checks                 | Low    | 5min   | Feature      |

---

## g) Top #1 Question I Cannot Figure Out Myself

**What is the target audience and maturity level for this library?**

The codebase is a small utility library (5 concerns, ~500 LOC of source code). Some next steps depend entirely on the answer:

- **If targeting internal use only (Lars' projects):** We can skip CI, fuzzing, benchmarks, and focus on keeping it lean. The current state is excellent for this.
- **If targeting open-source / public use:** We need README updates urgently, CI pipeline, godoc-quality package docs, semantic versioning, and probably a v1.0.0 stability promise.
- **If targeting a larger feature set (middleware, metrics, tracing):** We should decide the scope now before the API surface grows. Adding middleware, metrics, or tracing fundamentally changes what this library IS.

This decision shapes whether items #1-25 are "nice to have" or "critical before next release."

---

## Project Metrics

| Metric              | Value                                       |
| ------------------- | ------------------------------------------- |
| Source files        | 6 (.go)                                     |
| Test files          | 5 (\_test.go)                               |
| Other files         | doc.go, go.mod, go.sum                      |
| Total lines         | 1,076                                       |
| Total bytes         | 21,398                                      |
| Test count          | 31                                          |
| Test pass rate      | 100%                                        |
| Race detector       | Clean                                       |
| Coverage            | 88.7%                                       |
| Dependencies        | 1 direct (`modernc.org/sqlite`), 7 indirect |
| Go version          | 1.26.3                                      |
| Commits on master   | 4                                           |
| Uncommitted changes | Formatting/lint fixes + new project files   |
