# Features

Honest inventory by module. Statuses are verified against code and tests, not
aspirations.

| Status               | Meaning                                             |
| -------------------- | --------------------------------------------------- |
| FULLY_FUNCTIONAL     | Code present and working (test suite passes).       |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps or documented limitations. |
| PLANNED              | Designed/documented, no code yet.                   |

## core (`github.com/larsartmann/go-appkit`)

| Feature                                   | Status           | Evidence                                       |
| ----------------------------------------- | ---------------- | ---------------------------------------------- |
| `Service` lifecycle (Run/Start/Shutdown)  | FULLY_FUNCTIONAL | `service.go`, `service_test.go`                |
| Graceful drain (ready→503, delay, stop)   | FULLY_FUNCTIONAL | `service.go:134`                               |
| Default middleware stack (replace/extend) | FULLY_FUNCTIONAL | `middleware.go`                                |
| Health endpoints `/health`, live, ready   | FULLY_FUNCTIONAL | `health.go`, `health_test.go`                  |
| `ReadyCheck` external readiness gate      | FULLY_FUNCTIONAL | `config.go:50`, `service.go:192`               |
| charmbracelet/logging (`InitLogger`)      | FULLY_FUNCTIONAL | `logger.go`                                    |
| error-family re-exports (`HTTPStatus`…)   | FULLY_FUNCTIONAL | `errors.go`                                    |
| SSE-safe `WriteTimeout` configuration     | FULLY_FUNCTIONAL | `config.go` (`NoTimeout`), `notimeout_test.go` |

## cqrs (`github.com/larsartmann/go-appkit/cqrs`)

| Feature                              | Status           | Evidence                 |
| ------------------------------------ | ---------------- | ------------------------ |
| `EventService` over go-cqrs-lite v4  | FULLY_FUNCTIONAL | `eventservice.go`        |
| `EventConfig.Logger` worker logging  | FULLY_FUNCTIONAL | `logger_test.go`         |
| DLQ (SQLite) + replay/purge          | FULLY_FUNCTIONAL | `dlq_test.go`            |
| `EventConfig.FlightRecorder`         | FULLY_FUNCTIONAL | `flightrecorder_test.go` |
| `EventConfig.Metrics` recorder hook  | FULLY_FUNCTIONAL | `metrics_test.go`        |
| Projection readiness + lag accessors | FULLY_FUNCTIONAL | `readiness_test.go`      |
| Read-your-writes staleness guards    | FULLY_FUNCTIONAL | `staleness_test.go`      |

## docs (`github.com/larsartmann/go-appkit/docs`)

| Feature                               | Status           | Evidence                  |
| ------------------------------------- | ---------------- | ------------------------- |
| Catalog builder (OpenAPI/AsyncAPI/D2) | FULLY_FUNCTIONAL | `docs.go`, `docs_test.go` |
| `RegisterDocs` docserver mounting     | FULLY_FUNCTIONAL | `docs.go:34`              |

## errorpages (`github.com/larsartmann/go-appkit/errorpages`)

| Feature                            | Status           | Evidence                             |
| ---------------------------------- | ---------------- | ------------------------------------ |
| Pretty 404/405 (`Mount`/`Wrap`)    | FULLY_FUNCTIONAL | `errorpages_test.go`                 |
| Family-classified error pages      | FULLY_FUNCTIONAL | `errorpages_test.go:25`              |
| JSON contract negotiation          | FULLY_FUNCTIONAL | `errorpages_test.go:57`              |
| Render-failure plain-text fallback | FULLY_FUNCTIONAL | `errorpages_test.go` (failingWriter) |
| stdlib redirect parity in `Wrap`   | FULLY_FUNCTIONAL | `errorpages_test.go:246`             |

## realtime (`github.com/larsartmann/go-appkit/realtime`)

| Feature                               | Status           | Evidence                         |
| ------------------------------------- | ---------------- | -------------------------------- |
| SSE hub (broadcast, patch, shutdown)  | FULLY_FUNCTIONAL | `hub.go`                         |
| SSE handler (replay, heartbeat, CORS) | FULLY_FUNCTIONAL | `handler.go`                     |
| Subscribe-before-replay + live dedup  | FULLY_FUNCTIONAL | `handler.go`, `realtime_test.go` |
| Last-Event-ID resume                  | FULLY_FUNCTIONAL | `handler.go`                     |

Known limitation: on event bursts larger than the broadcaster buffer (64),
excess events are dropped and healed by client Last-Event-ID reconnect.

## flightrecorder (`github.com/larsartmann/go-appkit/flightrecorder`)

| Feature                                | Status           | Evidence               |
| -------------------------------------- | ---------------- | ---------------------- |
| Trigger-based trace capture middleware | FULLY_FUNCTIONAL | `middleware.go`, tests |
| Snapshot endpoint (`SnapshotHandler`)  | FULLY_FUNCTIONAL | `handler.go`, tests    |
| Auto-reset for repeated captures       | FULLY_FUNCTIONAL | `middleware.go`        |

## flightrecorderhealth (`github.com/larsartmann/go-appkit/flightrecorderhealth`)

| Feature                                                | Status           | Evidence                       |
| ------------------------------------------------------ | ---------------- | ------------------------------ |
| `Checkable` (health-dashboard visibility)              | FULLY_FUNCTIONAL | `adapter.go`, 5 tests          |
| `Trigger` (auto-capture on health failure)             | FULLY_FUNCTIONAL | `adapter.go`, 10 tests         |
| `Register` convenience                                 | FULLY_FUNCTIONAL | `adapter.go`, 2 tests          |
| Compile-time contract assertions (`health.HealthRecorder`, `do.HealthcheckerWithContext`) | FULLY_FUNCTIONAL | `contract_test.go` |
| Real `health.New` Probe end-to-end wiring                | FULLY_FUNCTIONAL | `TestIntegration_RealProbeEndToEnd` |
| Runnable godoc examples (`Register`, `NewCheckable`, `NewTrigger`) | FULLY_FUNCTIONAL | `example_test.go` |
| No-capture hot-path benchmark                          | FULLY_FUNCTIONAL | `benchmark_test.go` (~4.7µs/batch) |
| Custom trigger funcs (`OnError`, `OnAlways`, etc.)     | FULLY_FUNCTIONAL | `TestTrigger_CustomTriggerFunc`|
| Cooldown (trace-flood prevention)                      | FULLY_FUNCTIONAL | `TestTrigger_WithCooldown`     |
| Concurrency-safe `lastCapture` (`sync.Mutex`)          | FULLY_FUNCTIONAL | `TestTrigger_ConcurrentCooldownIsRaceFree` |
| Logger integration (`WithTriggerLogger`)               | FULLY_FUNCTIONAL | `TestTrigger_WithLogger`       |
| Multi-trigger observability (`WithServiceName`)        | FULLY_FUNCTIONAL | `TestTrigger_WithServiceName_NotLoggedWhenEmpty` |
| go-error-family classification (`Rejection` / `Infrastructure`) | FULLY_FUNCTIONAL | `adapter.go:HealthCheck`       |
