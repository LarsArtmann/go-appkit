# Changelog

## [Unreleased]

### Added

- Nothing yet.

### Fixed

- Resolved all golangci-lint findings: Go 1.26 blank-assignment fix (`_, _ = fmt.Fprint`), noinlineerr conversions in tests, `init()` replaced by a constructor for the recorder mutex, package doc reordered so the cqrs-lint directive folds into the godoc block, justified nolint directives on interface-mirroring returns and pure delegation.


## [0.3.0] - 2026-08-16

> Minor bump: new opt-in read-your-writes APIs; no breaking changes.

### Added

- `EventService.CheckStaleness(budget)` and `CheckProjectionStaleness(name, budget)`
  — read-time staleness guards for read-your-writes: Transient error when
  projection lag exceeds the budget (cqrs-lint E014's supported v4 answer;
  projectionhost v4.3.0 has no post-command drain API).
- README: "Read-your-writes" section and cqrs-lint workspace/wrapper gotchas.
- Staleness test suite: fresh, disabled, stale (Transient), unknown projection
  (Rejection), and per-projection variants. `closeOnConstructionFailure`
  error-join path tested via injected failing closer.

### Fixed

- Constructor-abort paths no longer discard `bundle.GracefulClose` errors
  (cqrs-lint C023): a close failure during `NewEventService` teardown is
  appended to the primary error via `errors.Join`.
- Bumped `storage/v4` v4.7.0 → v4.7.1: v4.7.0 had a build bug (`err =` instead
  of `err :=` in `sql/keyset.go`).

### Changed

- `.cqrs-lint.json` preset: `library` → `library-framework` (disables ALL
  F-series adoption-coaching rules — this module is a framework wrapper, not
  a go-cqrs-lite app). Added pinned feature profile to prevent auto-detection drift.

## [0.2.0] - 2026-08-15

First tagged release of the cqrs module. It supersedes the untagged v3-based
snapshots that shipped inside root `v0.2.0`; the module now targets
go-cqrs-lite **v4** and requires `GOEXPERIMENT=jsonv2` (Go 1.25+).

### Changed — go-cqrs-lite v3 → v4

- Migrated to `stack/v4`, `stack/sqlite/v4`, `projectionhost/v4` (v4.3.0), with
  `storage/v4` v4.6.0, `event/v4` v4.6.0, `id/v4` v4.4.0. The v4 SQLite stack
  removes the SQLITE_BUSY risk of the v3 shared-pool design.
- `DB()` misconfiguration errors are classified via go-error-family
  (`cqrs.db_not_sql` rejection) instead of `fmt.Errorf`.
- `Shutdown` surfaces projection-host stop failures via `errors.Join` instead of
  swallowing them.

### Added — projection host wiring (all optional `EventConfig` fields)

- `EventConfig.Logger` — projection worker lifecycle logs (crashes, restarts,
  dead-letter captures) flow into the service logger.
- `EventConfig.DLQ` — poison-event dead-lettering: after `DLQ.Threshold` handler
  failures the event moves to a SQLite dead-letter store and the checkpoint
  advances. Accessors: `DeadLetterStore()`, `ReplayDeadLetters()`,
  `ResetProjection(ctx, name, WithPurgeDeadLetters())`.
- `EventConfig.FlightRecorder` — captures a runtime/trace snapshot when a
  projection worker terminally fails (pairs with the appkit `flightrecorder`
  module).
- `EventConfig.Metrics` — backend-agnostic `projectionhost.MetricsRecorder`
  hook: processed, errored, dead-lettered, worker restarts/failures, checkpoint
  advances. No Prometheus/OpenTelemetry dependency imposed.
- `ReadyCheck()` and `LagPerProjection()` — projection-aware readiness for
  appkit's `/health/ready` (503 until the host is live) and per-projection lag
  introspection.

## [0.1.0] - 2026-07-26

Untagged baseline that shipped inside root `v0.2.0`: `EventService` over
go-cqrs-lite v3 (`stack/sqlite` + `projectionhost`) with lifecycle-managed
start/stop and a SQLite event store.
