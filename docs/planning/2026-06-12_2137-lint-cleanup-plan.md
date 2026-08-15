# Plan: Resolve All `golangci-lint run` Issues

**Date:** 2026-06-12 21:37
**Scope:** `github.com/larsartmann/go-appkit` (library, single package, Go 1.26.3)
**Goal:** Get `golangci-lint run ./...` to exit 0 (currently 42 issues) without regressing any tests.

## Issue Inventory (42 total, sorted by linter and impact)

| #  | Linter             | File:Line                                              | Category         | Impact | Effort | Priority |
| -- | ------------------ | ------------------------------------------------------ | ---------------- | ------ | ------ | -------- |
| 1  | `depguard`         | sqlite.go:9                                            | Security/Cfg     | High   | Low    | 1        |
| 2  | `errcheck`         | shutdown_test.go:82                                    | Test correctness | Med    | Low    | 2        |
| 3  | `errchkjson`       | health.go:44                                           | Correctness      | High   | Low    | 3        |
| 4  | `wrapcheck`        | server.go:119                                          | API contract     | Med    | Low    | 4        |
| 5  | `contextcheck`     | shutdown.go:44,48                                      | Correctness      | High   | Low    | 5        |
| 6  | `forcetypeassert`  | server_test.go:194                                     | Test safety      | Med    | Low    | 6        |
| 7  | `gosec`            | server_test.go:189                                     | Test safety      | Med    | Low    | 7        |
| 8  | `gochecknoglobals` | sqlite.go:12                                           | Style            | Low    | Med    | 8        |
| 9  | `goconst`          | sqlite.go:13                                           | Style            | Low    | Low    | 9        |
| 10 | `exhaustruct`      | server.go:23,70,72; logger.go:75                       | Style            | Low    | Low    | 10       |
| 11 | `mnd`              | server.go:24-27, shutdown.go:21                        | Readability      | Med    | Low    | 11       |
| 12 | `err113`           | logger.go:38,51; sqlite.go:45,74; shutdown_test.go:42  | Quality          | Med    | Low    | 12       |
| 13 | `varnamelen`       | logger.go:42,66; sqlite.go:48; sqlite_test.go:15,56,79 | Readability      | Low    | Low    | 13       |
| 14 | `noctx`            | server.go:84; sqlite.go:77; tests                      | Correctness      | High   | Med    | 14       |

## Strategy

All issues are mechanical lint cleanups. No behavioral changes. Tests must stay green.

The most efficient order is **by file** (one pass per file fixes multiple linters):

1. `.golangci.yml` (if needed) — allow `modernc.org/sqlite` for depguard
2. `server.go` — mnd, exhaustruct, noctx, wrapcheck
3. `logger.go` — err113, varnamelen, exhaustruct
4. `sqlite.go` — depguard (via cfg), err113, goconst, gochecknoglobals, noctx, varnamelen
5. `shutdown.go` — contextcheck, mnd
6. `health.go` — errchkjson
7. `*_test.go` — noctx, errcheck, err113, varnamelen, forcetypeassert, gosec

## Tasks (≤ 12 min each)

| #   | Task                                                                                                       | Time | Why                                     |
| --- | ---------------------------------------------------------------------------------------------------------- | ---- | --------------------------------------- |
| T1  | Add `modernc.org/sqlite` to `depguard.rules.main.allow`                                                    | 2m   | Required dep, depguard should permit it |
| T2  | Extract `server.go` magic numbers to `defaultPort`/`defaultReadTimeout`/`...` consts                       | 4m   | mnd                                     |
| T3  | Add `//nolint:exhaustruct` or fill missing fields in `ServerConfig{}`/`Server{}`/`http.Server{}`           | 5m   | exhaustruct                             |
| T4  | Replace `net.Listen("tcp", addr)` with `&net.ListenConfig{}.Listen(ctx, "tcp", addr)`                      | 5m   | noctx                                   |
| T5  | Wrap `s.server.Shutdown(ctx)` error with `fmt.Errorf("shutdown: %w", err)`                                 | 2m   | wrapcheck                               |
| T6  | Replace `fmt.Errorf("unsupported log level %q", l)` with `fmt.Errorf("%w: %q", errUnsupportedLogLevel, l)` | 4m   | err113                                  |
| T7  | Same for log format in `logger.go`                                                                         | 3m   | err113                                  |
| T8  | Rename `w` to `writer` in `logger.go` (param + var)                                                        | 3m   | varnamelen                              |
| T9  | Add `AddSource: false` to `slog.HandlerOptions`                                                            | 2m   | exhaustruct                             |
| T10 | Replace `errors.New("sqlite path is required")` with `fmt.Errorf("%w", errSQLitePathRequired)`             | 3m   | err113                                  |
| T11 | Same for `unsupported PRAGMA` error in `sqlite.go`                                                         | 3m   | err113                                  |
| T12 | Define `pragmaJournalMode`/etc consts and use them in `DefaultSQLitePRAGMAs` + map                         | 6m   | goconst + gochecknoglobals              |
| T13 | Change `db.Exec` to `db.ExecContext` in `sqlite.go`, threading ctx                                         | 5m   | noctx                                   |
| T14 | Rename `db` to `database` in `sqlite.go` and `sqlite_test.go`                                              | 3m   | varnamelen                              |
| T15 | Refactor `shutdown.go` to pass ctx directly (inline the helper)                                            | 5m   | contextcheck                            |
| T16 | Extract `defaultShutdownTimeout = 15 * time.Second` const                                                  | 2m   | mnd                                     |
| T17 | Check `json.NewEncoder(w).Encode(...)` error in `health.go`                                                | 3m   | errchkjson                              |
| T18 | Replace all `httptest.NewRequest(...)` with `httptest.NewRequestWithContext(...)` in test files            | 5m   | noctx                                   |
| T19 | Replace `http.Get(...)` with `http.NewRequestWithContext` + `client.Do` in `server_test.go`                | 5m   | noctx                                   |
| T20 | Replace `db.Ping()` with `db.PingContext(ctx)` in `sqlite_test.go`                                         | 3m   | noctx                                   |
| T21 | Handle `syscall.Kill` return value in `shutdown_test.go`                                                   | 2m   | errcheck                                |
| T22 | Use static sentinel error in `shutdown_test.go` instead of `errors.New("shutdown failed")`                 | 3m   | err113                                  |
| T23 | Use `_, ok := ln.Addr().(*net.TCPAddr)` pattern in `freePort`                                              | 2m   | forcetypeassert                         |
| T24 | Use `127.0.0.1:0` (or `localhost:0`) in `freePort` to avoid G102                                           | 2m   | gosec                                   |
| T25 | Run `golangci-lint run ./...` + `go test ./... -race` + `go vet ./...` — must all pass                     | 5m   | Verification                            |

## Out of Scope

- Code duplication (`dupl`) — currently 0 issues.
- Architectural rewrites — keep this PR scoped to lint cleanup.
- Adding new dependencies.

## Verification

```bash
go test ./... -race      # must pass
go vet ./...             # must pass
golangci-lint run ./...  # must exit 0 with 0 issues
```
