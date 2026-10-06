# Changelog

## [0.2.0] - 2026-10-06

### Changed

- **BEHAVIOR — captures are now asynchronous (non-blocking).** The
  middleware initiates snapshots via `go-flightrecorder`'s
  `SnapshotIfAsync` (detached `context.WithoutCancel`), so a request never
  pays trace-write latency (a MiB-scale file write on the hot path).
  Migration notes:
  - `WithLogger` now logs capture INITIATION — message renamed
    `flightrecorder: trace snapshot captured` → `flightrecorder: trace
    capture initiated` (grep-able contract change), still with
    method/path/duration/status for request correlation.
  - Completion telemetry (bytes, artifact path, sink errors) is invisible
    to the middleware by construction — wire `fr.WithMetrics` (or a
    logger hook) on the RECORDER at construction; see
    `OpsRecorderLoggerPreset` and the new example.
  - `WithAutoReset` semantics are preserved (re-arm per initiated
    capture). With a writer/file sink the once-latch still deduplicates
    tight bursts (best-effort); prefer a snapshot-dir sink for a
    deterministic file per incident — every initiation writes a new
    retained file there.
- `go-flightrecorder` floor v0.2.0 → **v0.2.1** (REQUIRED for the async
  middleware: v0.2.0 raced `Reset` against in-flight async captures on its
  once-latch — fatal under `-race`; fixed upstream in v0.2.1 with an
  atomic latch swap and a regression test).
- The new `example/` requires `go-appkit` (published v0.7.0) — example-only
  dependency; the library surface stays core-free.

### Added

- `OpsRecorderLoggerPreset(dir, maxSnapshots, maxBytes, log)`: the ops
  preset plus `fr.WithLogger` bridged to slog — recorder lifecycle events
  and, critically, RETENTION FAILURES (full disk, permission errors) land
  in the service log instead of vanishing. Nil logger skips the hook.
- Snapshot download mode: append `?download=1` to the snapshot endpoint to
  stream the raw trace as an `application/octet-stream` attachment — works
  for ANY sink (bypasses the configured destination via
  `fr.SnapshotToWriter`), so operators fetch the artifact over HTTP
  without shell access. Buffer-before-write keeps the JSON error contract
  on failure; Content-Length is exact.
- `SnapshotHandler(rec, opts...)` and `Mount(..., opts...)` accept handler
  options (additions-only: existing `SnapshotHandler(rec)` call sites
  compile unchanged). `WithSnapshotFilename(name)` sets the attachment
  filename (default `trace.trace`; use `trace.trace.gz` when the recorder
  compresses — godoc states the caveat).
- `doc.go` cookbook: `OnAll(OnError, OnLatency)` slow-failures-only
  narrowing, the typed `errors.Is(fr.ErrAlreadyEnabled)` shared-recorder
  start pattern, and a min-interval cooldown recipe for flapping
  dependencies.
- `example/`: the full ops loop — retention preset with slog lifecycle
  logging, `OnAll` narrowing middleware, download mount, completion
  telemetry via the recorder's metrics hook, fail-closed classified
  start. Verified live (trace.gz lands, initiation + completion logged,
  download streams gzip).

## [0.1.1] - 2026-09-29

### Changed

- Bumped `httputil` v0.11.0 → v1.2.0 (BuildFlow dependency sweep — the
  module shipped with a pre-1.0 httputil pin; the `Middleware` type and
  `ResponseRecorder` surface it uses are unchanged). Go directive
  1.27 → 1.27.1 (repo-wide toolchain unification).

### Added

- `OpsRecorderPreset(dir, maxSnapshots, maxBytes)`: the documented production
  option preset (snapshot dir/prefix, retention caps, gzip level 3, 10s
  minimum trace age).

### Changed

- `SnapshotHandler` returns an explicit 503 "recorder not enabled" when the
  recorder is disabled — previously a disabled recorder answered 200
  "snapshot captured" (silent data loss on a debug endpoint).
- The middleware's synthetic status errors are now go-error-family
  Infrastructure (`flightrecorder.http_status_error`), so consumer routing
  classifies them like any other dependency failure.

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
