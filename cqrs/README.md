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
    Logger:     svc.Logger, // projection worker logs flow into your service log
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

| Field          | Type              | Default          | Effect                                                                                                                                 |
| -------------- | ----------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `SQLitePath`   | `string`          | — (required)     | Path of the SQLite database file.                                                                                                      |
| `StackOptions` | `[]sqlite.Option` | none             | Passed through to `stack/sqlite.New` (v4 option set).                                                                                  |
| `Logger`       | `*slog.Logger`    | `slog.Default()` | Receives projection worker lifecycle events (crashes, restarts, dead-letter captures). Wire the same logger you gave `appkit.Service`. |

## Accessors

| Method                  | Returns                | Purpose                                                    |
| ----------------------- | ---------------------- | ---------------------------------------------------------- |
| `Bundle()`              | `*stack.Bundle`        | Event/command/query sinks and sources, journal, snapshots. |
| `Host()`                | `*projectionhost.Host` | Register projections before `StartProjections`.            |
| `DB()`                  | `(*sql.DB, error)`     | Raw SQLite handle for own queries.                         |
| `StartProjections(ctx)` | `error`                | Starts projection workers.                                 |
| `Shutdown(ctx)`         | `error`                | Stops workers and closes the store. Idempotent.            |

`Shutdown` joins and returns both the projection-host stop error and the
bundle-close error instead of swallowing them.
