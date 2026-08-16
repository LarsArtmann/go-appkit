# Changelog

## [Unreleased]

### Added

- Nothing yet.

### Fixed

- Nothing yet.

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
- `WithNowFunc(now)` — overrides the clock for testing.
- `Trigger` — a `health.HealthRecorder` that intercepts every health-check
  batch and triggers a flight recorder snapshot when the configured trigger
  function fires. Uses `SnapshotIfAsync` for non-blocking capture.
- `NewTrigger(rec, opts...)` — constructs a `Trigger` with options.
- `WithTriggerFunc(trigger)` — sets the flight recorder trigger function
  (default: `fr.OnError()`).
- `WithTriggerLogger(logger)` — routes capture events to a slog logger.
- `WithCooldown(d)` — sets a minimum duration between captures to prevent
  trace flooding on flapping services.
- `Register(injector, rec, name, opts...)` — convenience function that
  creates a `Checkable` and registers it in the samber/do injector.
