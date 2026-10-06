# go-appkit/cqrs

CQRS/event-sourcing integration for [go-appkit](../README.md) services, wrapping
[go-cqrs-lite](https://github.com/LarsArtmann/go-cqrs-lite) v4's `system` composition
layer (metaengine + `projectionhost`) behind a lifecycle-managed `EventService`.

> **v0.5.0 breaking change:** the engine room moved off the deprecated `stack/sqlite`
> preset (removed at go-cqrs-lite v5) onto `system.New`. Operators can swap engines at
> deployment time; see the Configuration table and the C/Q facade section below.

> **Build note:** Go 1.27.1 (the module's `go` floor). `encoding/json/v2` is
> default-on at that version — no `GOEXPERIMENT` needed (retired 2026-09-29).

## Usage

```go
es, err := cqrs.NewEventService(cqrs.EventConfig{
    DSN:    "app.db",          // or ConfigPath: "cqrs.yaml", or Driver: "memory"
    Logger: svc.Logger,        // projection worker logs flow into your service log
    DLQ:    &cqrs.DLQConfig{}, // poison events quarantined, not fatal
})
if err != nil {
    return err
}
defer func() { _ = es.Shutdown(context.Background()) }()

// Register projections on the host, then start them.
err = es.Host().Register(myProjection)

// Register commands and queries (the typed C/Q facade):
_ = cqrs.RegisterDecider(es, "Task", myDecider)
_ = cqrs.RegisterCommand[*command.BasicCommand, TaskState](es, "task.create", myHandler)
_ = cqrs.RegisterQuery[TaskQuery, TaskView](es, "task.view", myQueryHandler)

_ = es.StartProjections(ctx)

// Dispatch:
_ = es.Dispatch(ctx, cmd)
view, _ := cqrs.DispatchQueryChecked[TaskQuery, TaskView](ctx, es, 2*time.Second, q)

// Graceful stop: in-flight commands drain, projections stop, engines close.
err = es.Shutdown(ctx)
```

## Configuration

| Field                   | Type                             | Default              | Effect                                                                                                                                                                                                                                                                       |
| ----------------------- | -------------------------------- | -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DSN`                   | `string`                         | — (required)         | Database path/connection string. (The deprecated `SQLitePath` alias was removed in v0.6.0 — rename to `DSN`.) Required unless `ConfigPath`/`Deployment` is set or `Driver: "memory"`.                                                                                        |
| `Driver`                | `string`                         | `sqlite`             | metaengine driver name (`sqlite`, `memory`, `pebble`, `postgres`, ...). Drivers beyond `sqlite`/`memory` must be blank-imported by your module so they self-register.                                                                                                        |
| `Pragmas`               | `[]string`                       | WAL + busy_timeout   | SQLite pragmas. The defaults match the old `stack/sqlite` preset (`journal_mode=WAL`, `busy_timeout=5000`).                                                                                                                                                                  |
| `ConfigPath`            | `string`                         | —                    | Load the deployment from YAML via `system.LoadConfig` (koanf tags + `CQRS_` env overrides, e.g. `CQRS_ENGINES__PRIMARY__DRIVER`). Wins over `DSN`/`Driver`/`Pragmas`.                                                                                                        |
| `Deployment`            | `*system.DeploymentConfig`       | —                    | Fully pre-loaded operator config; wins over everything. Must declare a `RoleProjections` instance.                                                                                                                                                                           |
| `CheckpointStore`       | `event.CheckpointStore`          | system engine-backed | Projection checkpoint store override. Default: system's engine-backed store (ADR-0142, the `system_checkpoints` collection) — persists for sqlite/postgres, volatile for memory. Escape hatch for the legacy SQL table: see "Upgrading from v0.6.x".                         |
| `CommandMiddleware`     | `[]command.Middleware`           | none                 | Wraps every dispatched command. Compose via `DefaultCommandMiddleware(logger, tracer)` + your own. An in-flight drain tracker is installed outermost automatically.                                                                                                          |
| `QueryMiddleware`       | `[]query.Middleware`             | none                 | Wraps every dispatched query.                                                                                                                                                                                                                                                |
| `Logger`                | `*slog.Logger`                   | `slog.Default()`     | Receives projection worker lifecycle events (crashes, restarts, dead-letter captures). Wire the same logger you gave `appkit.Service`.                                                                                                                                       |
| `DLQ`                   | `*DLQConfig`                     | nil (disabled)       | Enables poison-event capture. Default store: SQLite table in the event database; default threshold: 3.                                                                                                                                                                       |
| `FlightRecorder`        | `*fr.Recorder`                   | nil (disabled)       | Captures a runtime/trace snapshot when a worker terminally fails (WorkerFailed). Type is `github.com/larsartmann/go-flightrecorder` — the same recorder the appkit `flightrecorder` middleware uses, so ONE shared instance can serve both. One active recorder per process. |
| `FlightRecorderTrigger` | `fr.TriggerFunc`                 | nil (= OnAlways)     | Gate for the capture: receives an `fr.TriggerContext` (Kind `"projection"`, Type = projection name, Err = terminal error). Default captures every terminal failure.                                                                                                          |
| `Metrics`               | `projectionhost.MetricsRecorder` | nil (disabled)       | Observes projection lifecycle events (processed, errored, dead-lettered, restarts, checkpoint lag). Backend-agnostic.                                                                                                                                                        |
| `HostOptions`           | `[]projectionhost.HostOption`    | none                 | Advanced host tuning — e.g. `WithCheckpointEvery(n)` (batch live-phase checkpoint saves), `WithOnFailed(fn)` (failure callback), `WithMaxRestarts`, `WithBatchSize`. Derived wiring (Logger, Metrics, FlightRecorder, DLQ) wins conflicts.                                   |

### Deployment shapes

Everything below is operator config, loaded through `ConfigPath`/`Deployment`
with the same YAML `system.LoadConfig` parses — every snippet here lives in
[`testdata/`](testdata/) and round-trips through the real parser in a test,
so it cannot drift from the pinned system version.

**Buses and publish fan-out.** Events from an instance publish to every bus
in its `publish` list (multi-bus fan-out); the projection host consumes from
the deployment's bus topology. Non-gochannel bus drivers (`nats`, `redis`)
must be blank-imported by the consumer to self-register:

```yaml
# testdata/deployment-buses.yaml
engines:
  primary:
    driver: sqlite
    dsn: file:events.db
buses:
  local:
    driver: gochannel
  edge:
    driver: nats
    url: nats://localhost:4222
instances:
  - role: source-of-truth
    engine: primary
    publish: [local, edge]
  - role: projections
    engine: primary
```

**Priorities, materialized views, durability, engine pools.** Layout
priorities (ADR-0124) resolve query-level → engine-level → global →
Balanced. Materialized views accelerate engine-side aggregates — a
deployment-time concern, never declared in code. A projections instance
with `engines:` (plural) is a mixed pool the planner routes freely within:

```yaml
# testdata/deployment-priority-views.yaml
engines:
  hot:
    driver: sqlite
    dsn: file:events.db
    priority: ReadSpeed
    materialized_views:
      - collection: orders
        fn: SUM
        column: amount
        group_by: customer
  archive:
    driver: sqlite
    dsn: file:archive.db
    priority: StorageSpace
priority:
  global: Balanced
  perEngine:
    hot: ReadSpeed
  perQuery:
    orders_by_customer: ReadSpeed
instances:
  - role: source-of-truth
    engine: hot
    durability: normal
  - role: projections
    engines: [hot, archive]
```

Durability tiers: `strict` (fsync every commit), `normal` (crash-safe, WAL
checkpoint window may be lost on power loss), `relaxed` (volatile — triggers
a WARN+OVERRIDE safety finding unless acknowledged).

**Manifest pinning, acknowledged warnings, cache.** `manifest_path` pins the
projection plan across restarts: removing a read model or changing its query
shape SCREAMs instead of silently orphaning data. `acknowledge_warnings`
silences a specific WARN+OVERRIDE finding by `rule:role`. `cache` adds a
read-through W-TinyLFU tier in front of an instance's event store:

```yaml
# testdata/deployment-manifest.yaml
engines:
  primary:
    driver: sqlite
    dsn: file:events.db
instances:
  - role: source-of-truth
    engine: primary
    cache:
      capacity: 10000
  - role: projections
    engine: primary
manifest_path: /var/lib/app/projection-manifest.json
acknowledge_warnings:
  - "durability-downgrade:events"
```

`NewEventService` logs any unacknowledged WARN+OVERRIDE findings at WARN on
boot (rule + detail + the acknowledgment escape hatch); `ScreamReport()`
returns the structured findings for dashboards.

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

### Sharing one flight recorder

`FlightRecorder` takes `*go-flightrecorder.Recorder` — the same type the
appkit `flightrecorder` middleware uses. Start ONE recorder at startup and
hand it to both layers; Go allows only a single active recorder per process,
and one instance serves both HTTP-level and projection-level captures:

```go
rec, _ := fr.New(fr.WithFile("trace.out"))
_ = rec.Start()
defer rec.Stop()

appkitCfg.OuterMiddlewares = []httputil.Middleware{
    flightrecorder.Middleware(rec, fr.OnError()),
}
cfg.FlightRecorder = rec        // projection captures on WorkerFailed
cfg.FlightRecorderTrigger = nil // nil = capture every terminal failure
```

`FlightRecorderTrigger` gates projection captures per failure; it receives
an `fr.TriggerContext` (Kind `"projection"`, Type = projection name,
Err = terminal error) and must stay fast — it runs on the worker goroutine.

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
counts as fresh. For dashboards, `es.Host().Status()` returns the
projection states as a SLICE (one entry per registered projection), and
`LagPerProjection()` maps each projection to its lag. On large streams, tune
catch-up throughput with
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
    DSN:     "events.db",
    Metrics: myRecorder, // implements projectionhost.MetricsRecorder
})

mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
    // serve myRecorder's counters, e.g. promhttp.Handler()
})
```

For OpenTelemetry (any reader: OTLP, Prometheus, stdout), use this module's
`NewOTelProjectionMetrics` — it implements `MetricsRecorder` on
`cqrs.projection.*` instruments, adding only the interface-only OTel API to
this module's dependency tree. Wire it together with the appkit `otel`
module for the HTTP side:

```go
provider, _ := appkitotel.Setup(          // github.com/larsartmann/go-appkit/otel
    appkitotel.WithService("myapp", version, instance),
    appkitotel.WithSpanExporter(otlpExporter),
    appkitotel.WithMetricReader(otlpMetricReader),
)
defer provider.Shutdown(ctx)

projectionMetrics, _ := cqrs.NewOTelProjectionMetrics(otel.Meter("myapp"))

es, _ := cqrs.NewEventService(cqrs.EventConfig{
    DSN:     "events.db",
    Metrics: projectionMetrics,
})
```

Instruments: `cqrs.projection.event.count` (projection, event type, status),
`cqrs.projection.event.duration` (ms), `cqrs.projection.worker.count`
(restarted/failed), `cqrs.projection.checkpoint.lag` (ms). Attribute keys
follow go-cqrs-lite's `cqrs.*` conventions, so one dashboard schema covers
HTTP spans and projection metrics.

## Upgrading from v0.6.x

**Checkpoint storage moved.** Default projection checkpoints no longer live
in the aux SQL table `checkpoints`; they ride system's engine-backed
checkpoint store (ADR-0142, the `system_checkpoints` collection on the
projection engine) — same file, one fewer sqlite connection. On the first
start after upgrading, every projection replays its stream exactly once
because the new collection starts empty; read models are derived, so the
replay is safe (it is the same work `ResetProjection` does). The old
`checkpoints` table is orphaned and can be dropped manually.

Two smaller consequences: `DB()` (the aux handle) now exists only when the
default DLQ wants it, and `Shutdown` always closes the system even when the
in-flight drain hits its deadline (previously the engines leaked on a stuck
command handler — the drain error is now joined with the close error).

**Escape hatch** — keep the pre-v0.7 SQL checkpoint table instead of the
engine-backed default:

```go
import (
	"database/sql"

	"github.com/larsartmann/go-cqrs-lite/storage/v4/eventstore"
	// the "sqlite" driver is registered by importing this module (blank import).
)

db, err := sql.Open("sqlite", "file:events.db?_pragma=busy_timeout(5000)")
if err != nil {
	return err
}

if _, err := db.ExecContext(ctx, eventstore.SQLiteCheckpointSchema()); err != nil {
	return err
}

store, err := eventstore.NewSQLiteCheckpointStore(db)
if err != nil {
	return err
}

eventSvc, err := cqrs.NewEventService(cqrs.EventConfig{
	DSN:             "file:events.db",
	CheckpointStore: store, // legacy SQL checkpoints; close db yourself after Shutdown
})
```

## Accessors

| Method                                   | Returns                          | Purpose                                                                                     |
| ---------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------- |
| `System()`                               | `*system.System`                 | Full go-cqrs-lite surface: MetaEngine, Publisher, EventStore, SnapshotStore, introspection. |
| `Host()`                                 | `*projectionhost.Host`           | Register projections before `StartProjections`.                                             |
| `DB()`                                   | `(*sql.DB, error)`               | Raw SQLite handle for own queries.                                                          |
| `DeadLetterStore()`                      | `projectionhost.DeadLetterStore` | The configured DLQ store, or nil when disabled.                                             |
| `ReplayDeadLetters(ctx, name)`           | `(ReplayResult, error)`          | Pure retry of quarantined events into their projections.                                    |
| `ResetProjection(ctx, name, opts...)`    | `error`                          | Rewind a projection checkpoint (optionally purging dead letters).                           |
| `ReadyCheck()`                           | `bool`                           | All workers live or drained; wire to appkit's `/health/ready`.                              |
| `HealthCheck(ctx)`                       | `error`                          | K8s-grade liveness/readiness: engines ping + no failed workers.                             |
| `EngineHealth(ctx)`                      | `[]system.EngineHealth`          | Per-engine health (name + error) for dashboards — all engines, not just the first failure.  |
| `ScreamReport()`                         | `*system.ScreamReport`           | Config-safety findings; `NewEventService` logs warnings at boot.                            |
| `LagPerProjection()`                     | `map[string]time.Duration`       | Event-age lag per projection.                                                               |
| `CheckStaleness(budget)`                 | `error`                          | Read-time guard: Transient error when max lag exceeds budget.                               |
| `CheckProjectionStaleness(name, budget)` | `error`                          | Per-projection read-time guard; Rejection for unknown names.                                |
| `StartProjections(ctx)`                  | `error`                          | Starts projection workers.                                                                  |
| `Shutdown(ctx)`                          | `error`                          | Stops workers and closes the store. Idempotent.                                             |

`Shutdown` drains in-flight commands first, then stops workers and closes
engines, joining any errors instead of swallowing them. Even when the drain
context expires on a stuck command, the close still runs — engines never
leak.

### Wiring health into go-health / appkithealth

`HealthCheck` composes cleanly with the `appkithealth` module — one probe,
one dashboard, for HTTP and CQRS engines alike:

```go
import appkithealth "github.com/larsartmann/go-appkit/health"

probe := appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
	"cqrs-engines": eventSvc.HealthCheck,
	"cqrs-lag": func(_ context.Context) error {
		return eventSvc.CheckStaleness(2 * time.Second)
	},
})

// Per-projection lag for ops dashboards (a map, not an error):
for name, lag := range eventSvc.LagPerProjection() {
	slog.Info("projection lag", "projection", name, "lag", lag)
}
```

## Domain declarations

`EventConfig.Domain` passes a `system.DomainConfig` through to the composition
root, replacing hand-rolled host projections with typed read models: declare
the shape once and the planner builds the collection, its indexes, and the
fold wiring. Merge contract: the wrapper's in-flight drain tracker stays
outermost, its host-bootstrap projection is appended only when `Domain`
declares no projections of its own, derived host-option wiring wins conflicts,
and `EventConfig.CheckpointStore` wins over `Domain.CheckpointStore`.

```go
type TaskView struct {
	ID       string
	Title    string
	Status   string
	Priority int
}

type TaskCreated struct { // Created/Updated/Deleted suffix drives the fold
	ID       string
	Title    string
	Status   string
	Priority int
}

es, err := cqrs.NewEventService(cqrs.EventConfig{
	DSN: "events.db",
	Domain: &system.DomainConfig{
		Projections: []system.ProjectionDeclaration{
			system.QuerySet[TaskView]("tasks").
				On("task.created", TaskCreated{}).
				Filterable("status"). // generates an index
				Sortable("priority", false).
				Done(),
		},
	},
})

// Reads, after StartProjections and catch-up:
active, err := system.Find[TaskView](ctx, es.System(), "tasks",
	system.Where("status", "active"),
	system.OrderBy("priority", system.Desc))
```

`Lookup` point reads inherit their folds from the matching `Evolve`, and
non-convention events get explicit folds via `OnEvolution`:

```go
Domain: &system.DomainConfig{
	Evolutions: []system.EvolutionSpec{
		system.OnEvolution(
			system.Evolve[TaskView]("task").On("task.created", TaskCreated{}),
			"task.completed", TaskCompleted{},
			func(_ TaskCompleted, v *TaskView) { v.Status = "done" },
		).Done(),
	},
	Projections: []system.ProjectionDeclaration{
		system.Lookup[TaskView]("get-task").Done(), // no samples: inherits
	},
}

task, err := system.Get[TaskView](ctx, es.System(), "get-task", taskID)
```

Declare `Events` (the complete journal vocabulary: command emissions plus
external imports) and construction enforces the coeffect graph: consuming an
UNDECLARED type fails `NewEventService` with
`system.ErrDanglingEventSubscription`, which is nearly always a typo.

```go
Events: []event.Type{"task.created", "task.completed"},
```

### Command lifecycle audit trail

ADR-0117's `WithCommandLifecycle` is one call returning the recorder, a
middleware pair, and prebuilt projection declarations (dead-letter queue,
retry counts, failure/rejection logs, processing time, commands by actor).
The recorder appends lifecycle events to the event store, so it needs a store
handle BEFORE construction, while through this wrapper the store only exists
after `NewEventService`. Bind it lazily: the `Commands` hook runs inside
construction, before any dispatch can happen, so the delegate is always set
by the time the middleware first emits.

```go
// lazyStore delegates to the real event store once the service exists.
type lazyStore struct {
	mu     sync.Mutex
	target event.Store
}

func (s *lazyStore) bind(store event.Store) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.target = store
}

func (s *lazyStore) delegate() event.Store {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.target
}

func (s *lazyStore) Save(
	ctx context.Context, ref id.StreamRef, events []event.Event, expected event.Version,
) error {
	return s.delegate().Save(ctx, ref, events, expected)
}

func (s *lazyStore) AppendBatch(
	ctx context.Context, ref id.StreamRef, events []event.Event,
) error {
	return s.delegate().AppendBatch(ctx, ref, events)
}

func (s *lazyStore) Load(ctx context.Context, ref id.StreamRef) ([]event.Event, error) {
	return s.delegate().Load(ctx, ref)
}

func (s *lazyStore) LoadFromVersion(
	ctx context.Context, ref id.StreamRef, version event.Version,
) ([]event.Event, error) {
	return s.delegate().LoadFromVersion(ctx, ref, version)
}

func (s *lazyStore) LoadToVersion(
	ctx context.Context, ref id.StreamRef, maxVersion event.Version,
) ([]event.Event, error) {
	return s.delegate().LoadToVersion(ctx, ref, maxVersion)
}

func (s *lazyStore) LoadToTimestamp(
	ctx context.Context, ref id.StreamRef, maxTime time.Time,
) ([]event.Event, error) {
	return s.delegate().LoadToTimestamp(ctx, ref, maxTime)
}

store := &lazyStore{}
cl := system.WithCommandLifecycle(store)

es, err := cqrs.NewEventService(cqrs.EventConfig{
	DSN: "events.db",
	Domain: &system.DomainConfig{
		// Outer emits received/completed/dead-lettered; Attempt emits
		// failed/retried per attempt; retry middleware belongs between them.
		Middleware:  []command.Middleware{cl.OuterMiddleware, cl.AttemptMiddleware},
		Projections: cl.Projections, // prebuilt lifecycle read models
		Commands: func(sys *system.System) {
			store.bind(sys.EventStore())
		},
	},
})

// Read the lifecycle models like any declared query:
counts, err := metaengine.ExecuteTypedByName[
	clprojections.RetryCountQuery, map[string]int64](
	ctx, es.System().MetaEngine(), clprojections.RetryCount().Name,
	clprojections.RetryCountQuery{})
```

Lifecycle events are ordinary journal events: they replay, checkpoint, and
drain with everything else.

Every recipe in this section is exercised by a godoc Example in
`example_test.go` — keep them in sync when editing either side.

### Streaming read models

Declared collections are watchable: `metaengine.NewWatcher` over the
collection plus `metaengine.ServeSSE` streams every materialized change to
browser EventSource clients. Attach a replay journal with
`watcher.WithReplay(n)` and a reconnecting client that sends
`Last-Event-ID` catches up on missed changes (capped by
`WithSSEReplayLimit`).

```go
watcher := metaengine.NewWatcher[TaskView](es.System().MetaEngine(), "tasks")
replay := watcher.WithReplay(1000) // enable Last-Event-ID reconnection
defer watcher.Close()

mux.HandleFunc("GET /events/tasks", func(w http.ResponseWriter, r *http.Request) {
	_ = metaengine.ServeSSE(w, r, watcher, // stream ends with the request
		metaengine.WithSSEHeartbeat(30*time.Second))
})
```

This streams MATERIALIZED read models (post-fold state). For raw domain
events (pre-fold journal records), pair the
[appkit/realtime](../realtime) module's journal-backed replay instead — the
two answer different questions ("what does the list look like now" vs "what
happened").

## Command/query facade

The service exposes go-cqrs-lite's typed C/Q surface directly — no need to
reach for `System()`:

```go
// Registration (before dispatching):
_ = cqrs.RegisterDecider(es, "Task", TaskDecider)
_ = cqrs.RegisterCommand[*command.BasicCommand, TaskState](es, "task.create", createHandler)
_ = cqrs.RegisterQuery[TaskQuery, TaskView](es, "task.view", viewHandler)

// Dispatch:
_ = es.Dispatch(ctx, cmd)                                            // command
view, _ := cqrs.DispatchQuery[TaskQuery, TaskView](ctx, es, q)       // query

// Staleness-gated query: returns the Transient staleness error INSTEAD of
// answering when the read model lags beyond the budget:
view, _ = cqrs.DispatchQueryChecked[TaskQuery, TaskView](ctx, es, 2*time.Second, q)
```

`DefaultCommandMiddleware(logger, tracer)` composes recovery + optional OTel
tracing + logging as a starting chain; retry, idempotency, and circuit
breaking change command semantics, so they stay opt-in — append them via
`EventConfig.CommandMiddleware`.

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
cqrs-lint init        # create .cqrs-lint.json (library-framework preset for framework wrappers)
cqrs-lint .           # lint (4.8.1+ rejects the ./... form — run from inside the module)
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

## Build & verify

```bash
cd cqrs
GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.1 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.1 go build ./...
golangci-lint run ./...        # from this directory; the module .golangci.yml is the source of truth
```

`GOTOOLCHAIN=go1.27.1` is required for every command below on
machines whose default toolchain is older (this repo pins
`go 1.27.1` in every go.mod). `GOWORK=off` makes the run
hermetic: the module resolves exactly what its own go.mod pins,
via the module proxy — how consumers resolve it.

`codec/v4` uses `encoding/json/jsontext`, so the toolchain floor is hard.
`cqrs-lint` runs from inside this directory too (workspace-root runs
misattribute sub-module imports). After adding a wrapper feature, re-run
`cqrs-lint scorecard` (build must be green or the scorecard reports 0%).
