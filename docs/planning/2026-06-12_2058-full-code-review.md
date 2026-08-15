# Full Code Review: go-appkit

Date: 2026-06-12

## Pareto Breakdown

### 1% → 51% Impact (Do First)

| # | Task                                                                           | Impact              | Effort | File                          |
| - | ------------------------------------------------------------------------------ | ------------------- | ------ | ----------------------------- |
| 1 | Fix PRAGMA SQL injection — validate keys against allowlist                     | Security critical   | 15min  | `sqlite.go`                   |
| 2 | Make `InitLogger` return `(*slog.Logger, error)` instead of panic              | Library correctness | 15min  | `logger.go`, `logger_test.go` |
| 3 | Fix race condition in `server_test.go` — poll `Addr()` instead of `time.Sleep` | Test reliability    | 10min  | `server_test.go`              |

### 4% → 64% Impact

| # | Task                                                                           | Impact          | Effort | File                          |
| - | ------------------------------------------------------------------------------ | --------------- | ------ | ----------------------------- |
| 4 | Define `LogLevel` and `LogFormat` types (replace raw strings)                  | Type safety     | 20min  | `logger.go`, `logger_test.go` |
| 5 | Define `HealthStatus` type, couple with HTTP status code                       | Type safety, DX | 15min  | `health.go`, `health_test.go` |
| 6 | Fix flaky `Server` default config — call `DefaultServerConfig()` once          | Code quality    | 10min  | `server.go`                   |
| 7 | Add test coverage for `logger.go` — json format, auto format                   | Test coverage   | 15min  | `logger_test.go`              |
| 8 | Add test coverage for `sqlite.go` — custom pragmas, pool settings, error paths | Test coverage   | 20min  | `sqlite_test.go`              |

### 20% → 80% Impact

| #  | Task                                                                   | Impact        | Effort | File               |
| -- | ---------------------------------------------------------------------- | ------------- | ------ | ------------------ |
| 9  | Allow opting out of `/health` route registration                       | Flexibility   | 15min  | `server.go`        |
| 10 | Add `Server.Running() bool` and protect `ln` with sync                 | Safety        | 15min  | `server.go`        |
| 11 | Replace `fmt.Fprintf(os.Stderr)` with slog in `shutdown.go`            | Consistency   | 10min  | `shutdown.go`      |
| 12 | Add tests for `Server.Shutdown()`, `Addr()` nil, port conflicts        | Test coverage | 20min  | `server_test.go`   |
| 13 | Add tests for signal delivery and default config in `shutdown_test.go` | Test coverage | 15min  | `shutdown_test.go` |
| 14 | DRY `healthHandler` — extract shared JSON encode helper                | Code quality  | 10min  | `health.go`        |
| 15 | Commit uncommitted go.mod/go.sum changes                               | Housekeeping  | 5min   | `go.mod`, `go.sum` |

## Execution Graph (D2)

```d2
direction: down

p1_group: {
  label: "Phase 1 — 1% → 51% (Security & Correctness)"

  fix_pragma_injection: Fix PRAGMA SQL injection {
    file: sqlite.go
    effort: 15min
  }

  fix_logger_panic: InitLogger returns error {
    file: logger.go
    effort: 15min
  }

  fix_test_race: Fix time.Sleep race {
    file: server_test.go
    effort: 10min
  }

  fix_pragma_injection -> fix_logger_panic -> fix_test_race
}

p2_group: {
  label: "Phase 2 — 4% → 64% (Type Safety)"

  log_types: Define LogLevel + LogFormat types {
    file: logger.go
    effort: 20min
  }

  health_types: Define HealthStatus type {
    file: health.go
    effort: 15min
  }

  server_defaults: DRY server defaults {
    file: server.go
    effort: 10min
  }

  logger_tests: Add logger format tests {
    file: logger_test.go
    effort: 15min
  }

  sqlite_tests: Add sqlite config tests {
    file: sqlite_test.go
    effort: 20min
  }

  log_types -> health_types -> server_defaults -> logger_tests -> sqlite_tests
}

p3_group: {
  label: "Phase 3 — 20% → 80% (Polish)"

  health_opt_out: Allow /health opt-out {
    file: server.go
    effort: 15min
  }

  server_running: Add Server.Running() {
    file: server.go
    effort: 15min
  }

  shutdown_slog: Use slog in shutdown {
    file: shutdown.go
    effort: 10min
  }

  server_tests: Add server edge case tests {
    file: server_test.go
    effort: 20min
  }

  shutdown_tests: Add shutdown signal tests {
    file: shutdown_test.go
    effort: 15min
  }

  health_dry: DRY health handler {
    file: health.go
    effort: 10min
  }

  commit_mods: Commit go.mod/go.sum {
    effort: 5min
  }

  health_opt_out -> server_running -> shutdown_slog -> server_tests -> shutdown_tests -> health_dry -> commit_mods
}

p1_group -> p2_group -> p3_group
```
