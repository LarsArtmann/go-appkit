# go-appkit/cqrs

CQRS/event-sourcing integration for [go-appkit](../README.md) services, wrapping
[go-cqrs-lite](https://github.com/LarsArtmann/go-cqrs-lite) v4 (`stack/sqlite` + `projectionhost`)
behind a lifecycle-managed `EventService`.

> **Build note:** requires `GOEXPERIMENT=jsonv2` (go-cqrs-lite's codec/v4 uses
> `encoding/json/jsontext`), available from Go 1.25.

## Usage

```go
es, err := cqrs.NewEventService(cqrs.EventConfig{
    SQLitePath: "app.db",
    Logger:     svc.Logger,     // projection worker logs flow into your service log
    DLQ:        &cqrs.DLQConfig{}, // poison events quarantined, not fatal
})
if err != nil {
    return err
}
defer func() { _ = es.Shutdown(context.Background()) }()

// Register projections on the host, then start them.
err = es.Host().Register(myProjection)

err = es.StartProjections(ctx)

// Graceful stop: projections drain, then the event store closes.
err = es.Shutdown(ctx)
```

## Configuration

| Field            | Type                             | Default          | Effect                                                                                                                                 |
| ---------------- | -------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `SQLitePath`     | `string`                         | — (required)     | Path of the SQLite database file.                                                                                                      |
| `StackOptions`   | `[]sqlite.Option`                | none             | Passed through to `stack/sqlite.New` (v4 option set).                                                                                  |
| `Logger`         | `*slog.Logger`                   | `slog.Default()` | Receives projection worker lifecycle events (crashes, restarts, dead-letter captures). Wire the same logger you gave `appkit.Service`. |
| `DLQ`            | `*DLQConfig`                     | nil (disabled)   | Enables poison-event capture. Default store: SQLite table in the event database; default threshold: 3.                                 |
| `FlightRecorder` | `*flightrecorder.Recorder`       | nil (disabled)   | Captures a runtime/trace snapshot when a worker terminally fails (WorkerFailed). One active recorder per process.                      |
| `Metrics`        | `projectionhost.MetricsRecorder` | nil (disabled)   | Observes projection lifecycle events (processed, errored, dead-lettered, restarts, checkpoint lag). Backend-agnostic.                  |
| `HostOptions`    | `[]projectionhost.HostOption`    | none             | Advanced host tuning; derived wiring (Logger, Metrics, FlightRecorder, DLQ) wins conflicts.                                            |

### Dead-letter queue

With `DLQ` set, an event that fails more than `DLQ.Threshold` times is moved to
the dead-letter store and the projection checkpoint advances — one poison
event cannot stall a projection. Inspect, replay, and purge:

```go
entries, _ := es.DeadLetterStore().List(ctx, "user-projection")

// after fixing the handler bug:
result, _ := es.ReplayDeadLetters(ctx, "user-projection") // pure retry
_ = es.DeadLetterStore().Purge(ctx, "user-projection")     // then clear

// rebuild a projection from scratch, clearing its dead letters too:
_ = es.ResetProjection(ctx, "user-projection", projectionhost.WithPurgeDeadLetters())
```

The default SQLite store also implements `projectionhost.DeadLetterStoreAdmin`
(Count, ListPaged, PurgeBefore) — type-assert to use it for admin dashboards.

### Readiness

`ReadyCheck()` reports whether every projection worker is live or fully
drained. Wire it into appkit's composable readiness so `/health/ready` serves
503 until projections catch up (and flips back if a worker dies):

```go
appkitCfg := appkit.DefaultServiceConfig()
appkitCfg.ReadyCheck = eventSvc.ReadyCheck // composes with the drain probe

lag := eventSvc.LagPerProjection() // map[projectionName]time.Duration
```

### Metrics

`Metrics` takes any `projectionhost.MetricsRecorder` — a six-method,
backend-agnostic interface (`EventProcessed`, `EventErrored`,
`EventDeadLettered`, `WorkerRestarted`, `WorkerFailed`,
`CheckpointAdvanced`). Implementations must be concurrency-safe and must
not block; the host records fire-and-forget from every worker goroutine.
This module deliberately adds no metrics dependency — forward the calls to
whatever backend you run:

```go
es, _ := cqrs.NewEventService(cqrs.EventConfig{
    SQLitePath: "events.db",
    Metrics:    myRecorder, // implements projectionhost.MetricsRecorder
})

mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
    // serve myRecorder's counters, e.g. promhttp.Handler()
})
```

For Prometheus via OpenTelemetry, compose
`github.com/larsartmann/go-cqrs-lite/prometheus/v4` (`prometheus.Setup()` →
`otel.SetMeterProvider` → `mux.Handle("/metrics", provider.Handler())`) with
an adapter that implements `MetricsRecorder` on top of your meter.

## Accessors

| Method                                | Returns                          | Purpose                                                           |
| ------------------------------------- | -------------------------------- | ----------------------------------------------------------------- |
| `Bundle()`                            | `*stack.Bundle`                  | Event/command/query sinks and sources, journal, snapshots.        |
| `Host()`                              | `*projectionhost.Host`           | Register projections before `StartProjections`.                   |
| `DB()`                                | `(*sql.DB, error)`               | Raw SQLite handle for own queries.                                |
| `DeadLetterStore()`                   | `projectionhost.DeadLetterStore` | The configured DLQ store, or nil when disabled.                   |
| `ReplayDeadLetters(ctx, name)`        | `(ReplayResult, error)`          | Pure retry of quarantined events into their projections.          |
| `ResetProjection(ctx, name, opts...)` | `error`                          | Rewind a projection checkpoint (optionally purging dead letters). |
| `ReadyCheck()`                        | `bool`                           | All workers live or drained; wire to appkit's `/health/ready`.    |
| `LagPerProjection()`                  | `map[string]time.Duration`       | Event-age lag per projection.                                     |
| `StartProjections(ctx)`               | `error`                          | Starts projection workers.                                        |
| `Shutdown(ctx)`                       | `error`                          | Stops workers and closes the store. Idempotent.                   |

`Shutdown` joins and returns both the projection-host stop error and the
bundle-close error instead of swallowing them.
