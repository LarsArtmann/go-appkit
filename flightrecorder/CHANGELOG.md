# Changelog

## [Unreleased]

### Added

- Nothing yet.

### Fixed

- Nothing yet.

## [0.1.0] - 2026-08-16

First tagged release of the flightrecorder module. Requires
`GOEXPERIMENT=jsonv2` (imports `encoding/json/v2` directly). Depends on
`go-flightrecorder v0.2.0` and `httputil v0.11.0`.

### Added

- `Middleware(rec, trigger, opts...)` — captures a Go runtime flight trace (via
  go-flightrecorder) when the `fr.TriggerFunc` fires — e.g. on error-status or
  latency-threshold breach. By default the recorder auto-resets after a capture
  so later incidents are captured too.
- `WithErrorThreshold(code)` — trigger option: capture when the response status
  reaches the threshold.
- `WithLogger` — route capture events to a custom `slog` logger.
- `WithAutoReset(enabled)` — keep the first capture instead of auto-resetting.
- `SnapshotHandler(rec)` — serves the most recent trace snapshot as JSON;
  `Mount(mux, pattern, rec)` registers it on a stdlib mux.
- Snapshot writing uses the `encoding/json/v2` API (`json.MarshalWrite`).
