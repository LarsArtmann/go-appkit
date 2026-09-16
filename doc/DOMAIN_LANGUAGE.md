# Domain Language

A **Unified Language** for go-appkit — shared across Customer, Product Owner, Developer, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a customer, define it here.

## Glossary

| Term                        | Definition                                                                                                                   | Context                                          |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| **appkit**                  | This framework: a production-ready HTTP service lifecycle composing httputil, charmbracelet/log, and go-error-family         | The product                                      |
| **Service**                 | The core type owning the HTTP server, listener, mux, and logger; the consumer's single entry point                           | `appkit.NewService(cfg)`                         |
| **Drain**                   | The graceful-shutdown window after the readiness probe flips to 503 but before connections close                             | Load balancers stop sending traffic during drain |
| **Drain hook**              | Consumer callback run at drain start, while traffic is still served (e.g. health-surface 503 flips)                          | `ServiceConfig.DrainHooks`                       |
| **Shutdown hook**           | Consumer callback run after connections are released (e.g. telemetry flush)                                                  | `ServiceConfig.ShutdownHooks`                    |
| **Outer middleware**        | Middleware wrapping the ENTIRE chain (default stack included), outermost — where tracing belongs                             | `ServiceConfig.OuterMiddlewares`                 |
| **Sentinel**                | A magic duration value that opts out of a default instead of configuring it; registry: `NoTimeout` = -1, `NoDrainDelay` = -2 | Duration fields; the next sentinel takes -3      |
| **Ready probe**             | The internal readiness state behind `/health/ready`; flips 503 for the whole drain window                                    | `service.go`                                     |
| **Ready check**             | A consumer-supplied gate ANDed with the ready probe (e.g. projection catch-up)                                               | `ServiceConfig.ReadyCheck`                       |
| **Module**                  | An independently versioned Go module in this repo (core + satellites); never a subpackage of core                            | Repository layout                                |
| **Battery**                 | A consumer-demanded capability shipped as an opt-in module, ported from a proven in-family implementation                    | Battery program, `docs/feedback/processed/`      |
| **Module bay**              | The established slot pattern a new battery fills (go.mod, LICENSE, .golangci.yml, doc.go, README, CHANGELOG)                 | Adding a satellite module                        |
| **Default stack**           | Recovery → RequestID → Logging → Timeout → SecurityHeaders, replaceable via `Middlewares`                                    | Middleware                                       |
| **Error family**            | One of the go-error-family classifications mapped to HTTP status (Rejection 400, Conflict 409, Transient 503, …)             | Error handling, errorpages                       |
| **CQRS/ES**                 | Command Query Responsibility Segregation / Event Sourcing, provided by the cqrs module over go-cqrs-lite                     | cqrs module                                      |
| **EventService**            | The cqrs module's lifecycle-managed handle over go-cqrs-lite's `system` engine                                               | `cqrs.NewEventService(cfg)`                      |
| **Engine**                  | The storage/composition backend behind EventService (`system.New`; drivers: sqlite, memory, …)                               | cqrs v0.5.0+                                     |
| **Projection**              | An asynchronous read model built from events; readiness and staleness are first-class                                        | `Host()`, `ReadyCheck()`, staleness guards       |
| **DLQ (dead-letter queue)** | Quarantine for poison events so one failure cannot stall a projection; supports replay and purge                             | `EventConfig.DLQ`                                |
| **Staleness guard**         | Read-time check that returns a Transient error when projection lag exceeds a budget (read-your-writes)                       | `CheckStaleness(budget)`                         |
| **Hub**                     | The realtime module's broadcaster + optional event-store pair                                                                | `realtime.NewHub()`                              |
| **Replay**                  | Resending missed events to a reconnecting SSE client via `Last-Event-ID`, deduplicated against live delivery                 | `realtime.Handler`, `WithStore`                  |
| **Flight recorder**         | go-flightrecorder's ring buffer of recent runtime/trace data; ONE active recorder per process                                | flightrecorder + flightrecorderhealth modules    |
| **Snapshot**                | A captured trace window written for `go tool trace` analysis                                                                 | `SnapshotHandler`, trigger middleware            |
| **Trigger**                 | The predicate deciding when a snapshot fires (error, latency, always, health failure)                                        | `fr.TriggerFunc`                                 |
| **Probe**                   | A go-health check set with critical/non-critical classification and background caching                                       | health module, `NewProbe`                        |
| **Health dashboard**        | The opt-in go-health-dashboard UI (HTML/JSON/SSE/metrics) mounted by the health module                                       | `WithDashboard`                                  |
| **Integration module**      | The never-released `integration/` module pinning PUBLISHED tags to E2E-test exactly what consumers resolve                   | Cross-module/cross-repo contracts                |

## Entities

Objects with identity and lifecycle.

| Term           | Definition                                                               | Context                             |
| -------------- | ------------------------------------------------------------------------ | ----------------------------------- |
| `Service`      | Created via config, started, drained, shut down; idempotent shutdown     | Core lifecycle                      |
| `Hub`          | Created with options, broadcasts, drains before the server stops         | realtime lifecycle                  |
| `Recorder`     | Started once per process, captures snapshots, reset between captures     | flightrecorder once-latch semantics |
| `EventService` | Constructed from `EventConfig`, projections started, drained on shutdown | cqrs lifecycle                      |

## Value Objects

Immutable values with meaning.

| Term                   | Definition                                                                     | Context                      |
| ---------------------- | ------------------------------------------------------------------------------ | ---------------------------- |
| `EventID`              | Phantom-typed SSE event identifier used for replay deduplication               | go-sse / go-branded-id       |
| `TriggerContext`       | Immutable snapshot of why a capture fired (Kind, Type, Duration, Err)          | flightrecorder               |
| `LogLevel`/`LogFormat` | Typed strings; invalid values are construction errors, not panics              | logger                       |
| Error code             | Namespaced string (`health.check_panicked`, `flightrecorder.recorder_missing`) | go-error-family constructors |

## Events

Things that happen in the domain.

| Term                    | Definition                                                              | Context                       |
| ----------------------- | ----------------------------------------------------------------------- | ----------------------------- |
| Shutdown phase complete | The grep-able per-phase log line contract (`ready_flip`, …)             | Core shutdown logging         |
| WorkerFailed            | Terminal projection-worker failure; the DLQ/flight-recorder hook point  | cqrs / projectionhost         |
| Health-check batch      | One round of all probe checks; the flightrecorderhealth intercept point | health / flightrecorderhealth |

## Commands

Actions the system performs.

| Term                | Definition                                              | Context         |
| ------------------- | ------------------------------------------------------- | --------------- |
| `Dispatch`          | Run a command through the cqrs middleware chain         | cqrs C/Q facade |
| `ReplayDeadLetters` | Pure retry of quarantined events into their projections | DLQ admin       |
| `ResetProjection`   | Rewind a projection checkpoint (optionally purging DLQ) | DLQ admin       |

## Bounded Contexts

| Context               | Description                                                                    |
| --------------------- | ------------------------------------------------------------------------------ |
| Core (HTTP lifecycle) | Server, middleware, drain, health endpoints, logging                           |
| CQRS                  | Events, commands, queries, projections, engines, checkpoints                   |
| Realtime              | SSE transport: hubs, streams, replay, heartbeats                               |
| Telemetry             | otel providers/spans/metrics, flight recorder snapshots, trace-correlated logs |
| Health                | Probes, criticality, dashboards, drain lockstep                                |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
