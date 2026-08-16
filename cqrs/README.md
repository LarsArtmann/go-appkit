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

### Read-your-writes

Projections are asynchronous: a command handler returning does **not** imply the
read model has moved. go-cqrs-lite v4's answer is a read-time staleness guard,
not a post-command drain — `EventService` exposes it directly:

```go
mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
    if err := eventSvc.CheckProjectionStaleness("task-list", 2*time.Second); err != nil {
        // Transient classification: serve 503, or stale data with a warning.
        http.Error(w, "read model catching up", http.StatusServiceUnavailable)
        return
    }
    // serve the read model
})
```

`CheckStaleness(budget)` guards against the maximum lag across all workers,
`CheckProjectionStaleness(name, budget)` against one named read model. A
budget <= 0 disables the check; a worker that has not processed any event yet
counts as fresh. On large streams, tune catch-up throughput with
`HostOptions: []projectionhost.HostOption{projectionhost.WithBatchSize(n)}` —
the default batch size trades throughput for latency smoothness.

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
| `CheckStaleness(budget)`              | `error`                          | Read-time guard: Transient error when max lag exceeds budget.     |
| `CheckProjectionStaleness(name, budget)` | `error`                       | Per-projection read-time guard; Rejection for unknown names.      |
| `StartProjections(ctx)`               | `error`                          | Starts projection workers.                                        |
| `Shutdown(ctx)`                       | `error`                          | Stops workers and closes the store. Idempotent.                   |

`Shutdown` joins and returns both the projection-host stop error and the
bundle-close error instead of swallowing them.

## Cookbook: testing and linting your CQRS code

These recipes use go-cqrs-lite's companion packages directly — they are
test/dev-only tools, so appkit/cqrs does not depend on them. Add them to your
own `go.mod` when you use them.

### Decider tests with the scenario DSL

`github.com/larsartmann/go-cqrs-lite/scenario/v4` gives your decide/fold pairs
a fluent Given/When/Then suite — no database needed:

```go
import "github.com/larsartmann/go-cqrs-lite/scenario/v4"

func TestRenameTask(t *testing.T) {
    scenario.Given[renameCmd, taskState](t, foldTask, taskState{},
        mustEvent(evtTaskCreated{Title: "old"}),
    ).
        When(renameCmd{Title: "new"}, decideRename).
        Then(event.Type("task.renamed"))        // emitted event types
}

func TestRenameMissingTask(t *testing.T) {
    scenario.Given[renameCmd, taskState](t, foldTask, taskState{}).
        When(renameCmd{Title: "x"}, decideRename).
        ThenError(errTaskNotFound)              // or .ThenState(fold, initial, want)
}
```

### Projection tests

Same package, projection flavor — feed events, assert handler outcome:

```go
scenario.GivenProjection(t, taskListProjection, evt1, evt2, evt3).ThenNoError()
scenario.GivenProjection(t, taskListProjection, poisonEvent).ThenError()
```

### Test helpers that pair well with EventConfig

`github.com/larsartmann/go-cqrs-lite/testutil/v4`:

| Helper                  | Use with               | Recipe                                                                                      |
| ----------------------- | ---------------------- | ------------------------------------------------------------------------------------------- |
| `CapturingSlogHandler`  | `EventConfig.Logger`   | Point the config's logger at it and assert worker lifecycle lines (restarts, DLQ captures). |
| `DelayedJournal`        | slow-store edge cases  | Wraps a `SeekableJournal` with a delay — rehearsal for slow replay reads.                   |
| `NewCmd`, `NoopCommand` | handler plumbing tests | Quick command records / handler stubs.                                                      |
| `rapid` generators      | property tests         | `EventType()`, `StreamType()`, `Version()`, `MetadataMap()` feed rapid-based fuzzing.       |

### cqrs-lint: domain-aware linting

`cqrs-lint` (a standalone binary) statically analyzes go-cqrs-lite
consumers for CQRS anti-patterns and API misuse:

```bash
cqrs-lint init        # create .cqrs-lint.json (library preset for reusable modules)
cqrs-lint ./...       # lint
cqrs-lint scorecard   # which go-cqrs-lite capabilities you use / miss
cqrs-lint rules       # rule reference
```

Gotchas worth knowing:

- **Build first.** On a project that does not compile, old cqrs-lint versions
  silently reported "no Go files found"; current versions exit non-zero and
  name the load errors — either way, don't trust lint output on a broken build.
- **Run it inside the module.** From a workspace root it attributes
  sub-module imports to the root `go.mod` — go-appkit's root module has zero
  go-cqrs-lite dependencies yet earns an A018 "dead import" finding when run
  from the repo root. `cd cqrs && cqrs-lint ./...` is the honest scope.
- **Wrappers trip usage detectors.** This module never calls `Save`/`Publish`/
  `Dispatch` itself (A018) and passes `projectionhost.New` no `WithBatchSize`
  (P008) **by design** — it hands consumers the `Bundle` and forwards tuning
  via `HostOptions`. Take A/P-series findings on a library wrapper as
  questions, not orders.
- **Suppressions:** `//cqrs-lint:ignore(RULE) reason` on its own line above
  the finding (comma-separate multiple rules); `ignore-start`/`ignore-end`
  for ranges. v4.6.0 flags stale suppressions whose rule no longer fires —
  remove them when told they are safe to drop.
