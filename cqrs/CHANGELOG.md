# Changelog

## [Unreleased]

### Changed

- `Shutdown` now drains in-flight commands through the system drainer seam
  (`RegisterDrainer`) and ALWAYS closes the system afterwards — engines and
  registered closers no longer leak when the drain context expires on a
  stuck command handler. Drain and close errors are joined instead of the
  drain short-circuiting the close; the `cqrs.drain_inflight_failed`
  Infrastructure error is gone (the drainer's own `cqrs.drain_cancelled`
  Transient error surfaces through `sys.Drain`). Idempotence is unchanged.
- Default projection checkpoints moved from the aux SQL table `checkpoints`
  to system's engine-backed checkpoint store (ADR-0142, the
  `system_checkpoints` collection on the projection engine). Consequences
  for sqlite-file deployments: the aux sqlite connection no longer opens
  unless the default DLQ wants it (`DB()` now rejects without a DLQ);
  upgrading replays every projection exactly once (the old table is
  orphaned — read models are derived, so the replay is safe). Escape
  hatch: keep the legacy SQL store via `EventConfig.CheckpointStore`
  (README, "Upgrading from v0.6.x").

## [0.6.1] - 2026-10-06

### Changed

- Bumped the go-cqrs-lite fleet to the 2026-10-05/06 waves (command/event v4.13.1, id v4.7.1, query v4.10.1, system v4.10.2, metaengine v4.16.1 + sqliteengine v4.5.1 + projectionadapter v4.5.2, decider/middleware v4.7.2, otel v4.5.2, projection v4.4.2, projectionhost v4.5.3, storage v4.10.4, claiming v4.0.2, commandlifecycle(+projections) v4.2.2, dedup v4.2.4, dispatcher v4.5.2, kv v4.3.3, metadata v4.7.3, record v4.6.2, snapshot v4.6.1, watermill v4.6.4). Dependency-only; test-suite polish since v0.6.0 rides along.

## [0.6.0] - 2026-09-29

### Breaking

- Removed the deprecated `EventConfig.SQLitePath` alias (announced at
  v0.5.0 for removal at v0.6.0). Migration: rename the field to `DSN` —
  same value, no other change.

### Changed

- Dependency sweep (2026-09-29): go-cqrs-lite system v4.7.0 → v4.10.0,
  command v4.10.0 → v4.12.0, query v4.8.0 → v4.9.0, decider v4.6.0 →
  v4.7.0, middleware v4.6.0 → v4.7.0, otel/v4 v4.4.0 → v4.5.0,
  sqliteengine v4.3.0 → v4.4.0, projectionhost v4.4.0 → v4.5.1, event
  v4.11.0 → v4.12.0, id v4.6.0 → v4.6.1, projection v4.3.0 → v4.4.0,
  storage v4.9.0 → v4.10.1, go-error-family v0.10.0 → v0.11.0 (via the
  unreleased v0.10.1 hygiene pin — a docs/CI-only upstream release with
  zero code changes). Behavior
  change absorbed: projectionhost v4.5+ quarantines ONLY
  Rejection/Corruption-family errors to the DLQ (retryable errors restart
  the worker instead) — the DLQ test's poison fixture is now a Rejection
  that fails only the poison event (the old fixture relied on a backoff
  timing window that v4.5 closed). Go directive 1.26.7 → 1.27.1.
- Internal refactor: `sortedEngineNames` collects sorted map keys via
  `slices.Sorted(maps.Keys(...))` instead of a hand-rolled collect + sort;
  returns nil (not an empty non-nil slice) for empty maps — safe for all
  callers (range-only). Behavior-neutral otherwise.

## [0.5.0] - 2026-09-07

### Changed

- **Breaking (v5-survival migration):** the engine room moved from the
  deprecated `stack/sqlite` preset (removed at go-cqrs-lite v5, ADR-0123) to
  `system.New` with the `DomainConfig`/`DeploymentConfig` split. `EventService`
  now holds a `*system.System`; `Bundle()` is replaced by `System()` with
  direct accessors for the common surface (`Host()`, `DB()`,
  `DeadLetterStore()`, staleness guards, `ReadyCheck`).
- `EventConfig.SQLitePath` → `DSN` + `Driver` + `Pragmas`. `SQLitePath` still
  works as a deprecated alias (removed at v0.6.0). An empty DSN is a
  construction Rejection unless `Driver: "memory"`. SQLite defaults to WAL +
  busy_timeout pragmas (same defaults the old preset shipped); `StackOptions`
  are gone — use `Pragmas`.
- `ReadyCheck` now reports NOT-ready before `StartProjections` (the system
  auto-projection worker is idle until started). Previously an idle,
  never-started service reported ready — the new semantics are safer for
  /health/ready.
- Dead letters + checkpoints for the SQLite default now live on an auxiliary
  `*sql.DB` opened on the same database file (registered as a System closer),
  not inside a bundle. `DLQConfig.Store` is REQUIRED for non-sqlite drivers.

### Added

- **Command/query facade:** `RegisterDecider`, `RegisterCommand`,
  `RegisterQuery` (typed passthroughs), `Dispatch`, `DispatchQuery`,
  `DispatchQueryChecked` (staleness-gated query answering), and
  `CommandDispatcher()`/`QueryDispatcher()` accessors.
- `DefaultCommandMiddleware(logger, tracer)` — recovery + optional OTel
  tracing + logging; retry/idempotency/circuit-breaking stay consumer
  opt-ins via `EventConfig.CommandMiddleware`/`QueryMiddleware`.
- Operator config surfaces: `EventConfig.ConfigPath` (koanf YAML +
  `CQRS_` env overrides via `system.LoadConfig`) and
  `EventConfig.Deployment` (fully pre-loaded `*system.DeploymentConfig`).
  Storage resolution precedence: Deployment > ConfigPath > Driver/DSN/Pragmas.
- `EventConfig.CheckpointStore` override; default is a persistent
  SQLite checkpoint store (v0.4.0 parity — checkpoints survive restarts).
- In-flight command drain: `Shutdown` waits for commands executing through
  the middleware chain before closing engines (bounded by the context).
- Default engine pool: one `primary` engine with RoleSourceOfTruth +
  RoleProjections instances (mirrors the reference consumer). Drivers
  beyond `sqlite`/`memory` must be blank-imported by the consumer
  (engine self-registration contract).

## [0.4.0] - 2026-09-04

### Changed

- **Breaking (dependency alignment):** `EventConfig.FlightRecorder` is now
  `*github.com/larsartmann/go-flightrecorder.Recorder` — go-cqrs-lite
  `projectionhost/v4.4.0` switched `WithFlightRecorder` from its internal
  `flightrecorder/v4` type to go-flightrecorder, so the module did not
  compile against the pinned version until this migration. The
  `flightrecorder/v4` dependency is dropped; `go-flightrecorder v0.2.0`
  becomes a direct dependency. Consumers gain the ability to share ONE
  recorder instance between the appkit `flightrecorder` HTTP middleware and
  the projection host (Go still allows only one active recorder per
  process).

### Added

- `EventConfig.FlightRecorderTrigger fr.TriggerFunc` — optional gate for
  terminal-failure captures; receives an `fr.TriggerContext` (Kind
  `"projection"`, Type = projection name, Err = terminal error). Nil keeps
  the upstream default (capture every terminal failure, OnAlways).
- `NewOTelProjectionMetrics(meter)`: `projectionhost.MetricsRecorder`
  implementation on OTel instruments (`cqrs.projection.event.count`,
  `.event.duration`, `.worker.count`, `.checkpoint.lag`) with go-cqrs-lite
  `cqrs.*` attribute conventions. Closes the metrics path for
  `EventConfig.Metrics` (planning M10): pair with the new appkit `otel`
  module's `Setup` for one-provider HTTP + projection telemetry. Adds only
  the interface-only OTel API (`go.opentelemetry.io/otel/{attribute,metric}`)
  to the direct dependency set — no SDK, no exporter.

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
