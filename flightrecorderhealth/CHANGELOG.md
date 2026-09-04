# Changelog

## [Unreleased]

### Changed

- Bumped `go-health` from v0.0.2 to v0.1.1. The compile-time contract
  assertions still hold (`health.HealthRecorder`'s
  `RecordHealthCheckWithContext(ctx, do.Injector) map[string]error`
  signature is unchanged); the runtime code path never calls go-health, so
  this is a hygiene bump picking up jsonv2 serialization, the method-set
  handler guard, the `aggregate` package, and the 2026-09-04 probe
  lifecycle race fix.
- **Build requirement change:** the module now requires
  `GOEXPERIMENT=jsonv2` (go-health v0.1.1 imports `encoding/json/v2`).
  Previously advertised as jsonv2-free.
- Go directive bumped 1.26.5 → 1.26.7.

## [0.1.0] - 2026-08-16

First release of the flightrecorderhealth module. Bridges
`go-flightrecorder` with `go-health`, exposing flight-recorder state and
health-check failures through the health dashboard.

### Added

- `Checkable` — a health-checkable wrapper around a flight recorder that
  reports the recorder's operational state (enabled/disabled) in the health
  dashboard. Implements `do.HealthcheckerWithContext`.
- `NewCheckable(rec, opts...)` — constructs a `Checkable` with options.
- `WithCheckableName(name)` — sets the display name for the dashboard.
- `Trigger` — a `health.HealthRecorder` that intercepts every health-check
  batch and triggers a flight recorder snapshot when the configured trigger
  function fires. Uses `SnapshotIfAsync` for non-blocking capture.
- `NewTrigger(rec, opts...)` — constructs a `Trigger` with options.
- `WithTriggerFunc(trigger)` — sets the flight recorder trigger function
  (default: `fr.OnError()`).
- `WithTriggerLogger(logger)` — routes capture events to a slog logger.
- `WithCooldown(d)` — sets a minimum duration between captures to prevent
  trace flooding on flapping services.
- `WithServiceName(name)` — sets an identifier included in trigger log
  messages (useful when multiple triggers run in the same process).
- `Register(injector, rec, name, opts...)` — convenience function that
  creates a `Checkable` and registers it in the samber/do injector.
- Compile-time interface assertions: `*Trigger` satisfies
  `health.HealthRecorder` and `*Checkable` satisfies
  `do.HealthcheckerWithContext` (`contract_test.go`). Interface changes in
  either dependency become compile errors instead of silent breaks. The full
  chain (Probe batch → Trigger → snapshot on disk) is verified against a real
  `health.New` Probe in `TestIntegration_RealProbeEndToEnd`.
- Runnable godoc examples for `Register`, `NewCheckable`, and `NewTrigger`
  with verified output.
- `BenchmarkTrigger_RecordHealthCheckWithContext_AllPass` for the no-capture
  hot path (~4.7µs per batch with two passing services).
- Errors classified via [go-error-family](https://github.com/LarsArtmann/go-error-family):
  `flightrecorder.recorder_missing` (Rejection) for nil recorder, and
  `flightrecorder.recorder_disabled` (Infrastructure) for not-started recorder.

### Concurrency

- `Trigger.lastCapture` is guarded by `sync.Mutex`, allowing concurrent
  health-check batches to safely read and update the cooldown state.
- `Trigger.RecordHealthCheckWithContext` is safe for concurrent use; the
  underlying `SnapshotIfAsync` is also race-safe.

### Notes

- Requires `GOEXPERIMENT=jsonv2` since the go-health v0.1.1 bump (see
  Unreleased).
- Process-global singleton constraint applies: Go's `runtime/trace` allows
  only ONE active `flightrecorder.Recorder` per process. See the package
  doc and `go-flightrecorder` README for details.
