# go-appkit/flightrecorderhealth

Bridge between [go-flightrecorder](https://github.com/LarsArtmann/go-flightrecorder)
and [go-health](https://github.com/LarsArtmann/go-health) — exposes flight
recorder state in the health dashboard and auto-captures a trace snapshot when
health checks fail.

> **Build note:** this module does NOT require `GOEXPERIMENT=jsonv2`. Its
> dependencies (`go-health`, `samber/do/v2`, `go-flightrecorder`) all use plain
> `encoding/json`.

## What you get

- **Dashboard visibility** — wrap the recorder as a `do.HealthcheckerWithContext`
  so its own operational state (enabled / disabled) shows up as a row in the
  health dashboard.
- **Auto-capture on failure** — wrap the recorder as a `health.HealthRecorder`
  that intercepts every health-check batch and fires `SnapshotIfAsync` when a
  configurable trigger function evaluates to true. Default trigger: `fr.OnError()`
  — captures a trace snapshot when any health check returns an error.
- **Non-blocking capture** — `SnapshotIfAsync` runs the snapshot on a
  background goroutine, so the health-check loop is never delayed by trace I/O.
- **Cooldown** — `WithCooldown(d)` prevents trace flooding when a service
  flaps repeatedly.
- **Nil-safe** — nil `Trigger` and nil `Checkable` pass through to the injector
  without panicking, so they can be constructed before the recorder is wired.

## Quick start

```go
package main

import (
	"log"
	"time"

	frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/larsartmann/go-health"
	"github.com/samber/do/v2"
)

func main() {
	// Build the recorder. A directory sink keeps timestamped snapshots on disk.
	rec, err := fr.New(
		fr.WithSnapshotDir("/var/traces"),
		fr.WithMinAge(50*time.Millisecond),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := rec.Start(); err != nil {
		log.Fatal(err)
	}
	defer rec.Close()

	injector := do.New()

	// 1. Register the recorder as a health-checkable service so its status
	//    appears in the dashboard.
	frhealth.Register(injector, rec, "flight-recorder")

	// 2. Wire the trigger to auto-capture a trace snapshot on health failures.
	probe := health.New(injector,
		health.WithCriticalServices("database", "cache"),
		health.WithHealthRecorder(frhealth.NewTrigger(rec,
			frhealth.WithTriggerFunc(fr.OnError()),
			frhealth.WithServiceName("flight-recorder"),
			frhealth.WithCooldown(30*time.Second),
		)),
	)
	_ = probe // serve probe.ReadinessHandler(), probe.LivenessHandler(), ...
}
```

## When to use

Use this module when you want:

- A row in the health dashboard confirming that trace capture is running.
- Automatic trace snapshots when your services degrade, so you can pull up
  `go tool trace` against the failure window without having to manually trigger
  a capture.

**When NOT to use:**

- If you only have one health check or you control the failure path directly,
  skip the `Trigger` and call `rec.Snapshot()` / `rec.SnapshotToFile()` from
  your error handler.
- If you don't want runtime traces at all, this module is overhead.

## Trigger functions

| Function                | Fires when                                             |
| ----------------------- | ------------------------------------------------------ |
| `fr.OnError()`          | Any health check returns a non-nil error (default).    |
| `fr.OnErrorOrLatency()` | A check fails OR takes longer than the latency budget. |
| `fr.OnAlways()`         | Every health-check batch (useful for sampling).        |
| `fr.OnAny(fns...)`      | Any of the composed functions fire.                    |
| `fr.OnAll(fns...)`      | All of the composed functions fire.                    |

`fr.TriggerContext.Err` is set to the first failing service's error, so
`OnError` and `OnErrorOrLatency` fire correctly on health-check failures.

## Process-global singleton

Go's `runtime/trace` allows only ONE active `flightrecorder.Recorder` per
process. Wire a single recorder at startup and share it across all
integrations. If another package starts a recorder first, `Start()` returns
`flightrecorder.ErrAlreadyEnabled` and `Checkable.HealthCheck` reports
unhealthy for the losing side.

## Configuration

### Checkable

| Option              | Default           | Effect                                |
| ------------------- | ----------------- | ------------------------------------- |
| `WithCheckableName` | "flight-recorder" | Display name in the health dashboard. |

### Trigger

| Option              | Default        | Effect                                                                       |
| ------------------- | -------------- | ---------------------------------------------------------------------------- |
| `WithTriggerFunc`   | `fr.OnError()` | Decides when to capture.                                                     |
| `WithCooldown`      | 0              | Minimum duration between captures (30s–60s recommended for directory sinks). |
| `WithTriggerLogger` | nil            | slog logger for capture events.                                              |
| `WithServiceName`   | ""             | Identifier logged with each capture (multi-trigger setups).                  |

## API surface

| Symbol                                              | Purpose                                                          |
| --------------------------------------------------- | ---------------------------------------------------------------- |
| `NewCheckable(rec, opts...) *Checkable`             | Health-checkable wrapper for the recorder.                       |
| `Checkable.HealthCheck(ctx) error`                  | Reports nil when enabled, error otherwise.                       |
| `Checkable.Name() string`                           | Service name (for dashboard display).                            |
| `NewTrigger(rec, opts...) *Trigger`                 | HealthRecorder that captures on trigger fire.                    |
| `Trigger.RecordHealthCheckWithContext(ctx, inj)`    | Runs the batch, fires `SnapshotIfAsync` if trigger returns true. |
| `Register(injector, rec, name, opts...) *Checkable` | Convenience: creates + registers a `Checkable` in one call.      |

## Errors

Errors are constructed via [go-error-family](https://github.com/LarsArtmann/go-error-family):

| Error                              | Family         | When                                               |
| ---------------------------------- | -------------- | -------------------------------------------------- |
| `flightrecorder.recorder_missing`  | Rejection      | `Checkable.HealthCheck` with nil recorder.         |
| `flightrecorder.recorder_disabled` | Infrastructure | `Checkable.HealthCheck` when recorder not started. |
