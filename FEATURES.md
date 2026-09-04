# Features

Honest inventory by module. Statuses are verified against code and tests, not
aspirations.

| Status               | Meaning                                             |
| -------------------- | --------------------------------------------------- |
| FULLY_FUNCTIONAL     | Code present and working (test suite passes).       |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps or documented limitations. |
| PLANNED              | Designed/documented, no code yet.                   |

## core (`github.com/larsartmann/go-appkit`)

| Feature                                    | Status           | Evidence                                         |
| ------------------------------------------ | ---------------- | ------------------------------------------------ |
| `Service` lifecycle (Run/Start/Shutdown)   | FULLY_FUNCTIONAL | `service.go`, `service_test.go`                  |
| Graceful drain (ready→503, delay, stop)    | FULLY_FUNCTIONAL | `service.go:134`                                 |
| Default middleware stack (replace/extend)  | FULLY_FUNCTIONAL | `middleware.go`                                  |
| Health endpoints `/health`, live, ready    | FULLY_FUNCTIONAL | `health.go`, `health_test.go`                    |
| `ReadyCheck` external readiness gate       | FULLY_FUNCTIONAL | `config.go:50`, `service.go:192`                 |
| `OuterMiddlewares` outermost hook          | FULLY_FUNCTIONAL | `config.go:65`, `middleware_test.go`             |
| `ShutdownHooks` post-shutdown flush hooks  | FULLY_FUNCTIONAL | `service.go`, `shutdownhooks_test.go`            |
| `DrainHooks` drain-start readiness hooks   | FULLY_FUNCTIONAL | `config.go`, `service.go`, `drainhooks_test.go`  |
| Shutdown phase logging (per-phase INFO)    | FULLY_FUNCTIONAL | `service.go` (`logPhase`), `shutdownlog_test.go` |
| `NoDrainDelay` fast-test shutdown sentinel | FULLY_FUNCTIONAL | `config.go:37`, `config_test.go`                 |
| charmbracelet/logging (`InitLogger`)       | FULLY_FUNCTIONAL | `logger.go`                                      |
| error-family re-exports (`HTTPStatus`…)    | FULLY_FUNCTIONAL | `errors.go`                                      |
| SSE-safe `WriteTimeout` configuration      | FULLY_FUNCTIONAL | `config.go` (`NoTimeout`), `notimeout_test.go`   |

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
| OTel projection metrics adapter      | FULLY_FUNCTIONAL | `otelmetrics.go`         |

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

| Feature                                                                                   | Status           | Evidence                                         |
| ----------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------ |
| `Checkable` (health-dashboard visibility)                                                 | FULLY_FUNCTIONAL | `adapter.go`, 5 tests                            |
| `Trigger` (auto-capture on health failure)                                                | FULLY_FUNCTIONAL | `adapter.go`, 10 tests                           |
| `Register` convenience                                                                    | FULLY_FUNCTIONAL | `adapter.go`, 2 tests                            |
| Compile-time contract assertions (`health.HealthRecorder`, `do.HealthcheckerWithContext`) | FULLY_FUNCTIONAL | `contract_test.go`                               |
| Real `health.New` Probe end-to-end wiring                                                 | FULLY_FUNCTIONAL | `TestIntegration_RealProbeEndToEnd`              |
| Runnable godoc examples (`Register`, `NewCheckable`, `NewTrigger`)                        | FULLY_FUNCTIONAL | `example_test.go`                                |
| No-capture hot-path benchmark                                                             | FULLY_FUNCTIONAL | `benchmark_test.go` (~4.7µs/batch)               |
| Custom trigger funcs (`OnError`, `OnAlways`, etc.)                                        | FULLY_FUNCTIONAL | `TestTrigger_CustomTriggerFunc`                  |
| Cooldown (trace-flood prevention)                                                         | FULLY_FUNCTIONAL | `TestTrigger_WithCooldown`                       |
| Concurrency-safe `lastCapture` (`sync.Mutex`)                                             | FULLY_FUNCTIONAL | `TestTrigger_ConcurrentCooldownIsRaceFree`       |
| Logger integration (`WithTriggerLogger`)                                                  | FULLY_FUNCTIONAL | `TestTrigger_WithLogger`                         |
| Multi-trigger observability (`WithServiceName`)                                           | FULLY_FUNCTIONAL | `TestTrigger_WithServiceName_NotLoggedWhenEmpty` |
| go-error-family classification (`Rejection` / `Infrastructure`)                           | FULLY_FUNCTIONAL | `adapter.go:HealthCheck`                         |

## health (`github.com/larsartmann/go-appkit/health`)

| Feature                                                            | Status           | Evidence                                           |
| ------------------------------------------------------------------ | ---------------- | -------------------------------------------------- |
| `NewProbe` injector-free probe (concurrent, panic-isolated checks) | FULLY_FUNCTIONAL | `probe.go`, `probe_test.go`                        |
| Critical/non-critical readiness classification                     | FULLY_FUNCTIONAL | `TestNewProbe_ClassificationFollowsCriticality`    |
| `New` + `RegisterRoutes` / `Mount` mux wiring                      | FULLY_FUNCTIONAL | `mount.go`, `mount_test.go`                        |
| Kubelet probe routes (custom paths opt-in)                         | FULLY_FUNCTIONAL | `TestMount_ProbeOnlyRegistersKubeletRoutes`        |
| Real-time dashboard (HTML/JSON/SSE/metrics/trend) opt-in           | FULLY_FUNCTIONAL | `TestMount_WithDashboardServesHTMLJSONAndProbes`   |
| Drain lockstep (`Drain` → readiness 503 for the drain window)      | FULLY_FUNCTIONAL | `TestMount_DrainFlipsDashboardReadiness`, live E2E |
| Lifecycle guards (double-Start rejection, idempotent Shutdown)     | FULLY_FUNCTIONAL | `TestMount_LifecycleGuardsAndIdempotence`          |
| SDK validation errors surface via `Start` (errors.Is preserved)    | FULLY_FUNCTIONAL | `TestMount_StartPropagatesProbeValidationErrors`   |
| Runnable example (verified live: dashboard, probes, drain 503s)    | FULLY_FUNCTIONAL | `example/main.go`                                  |

## otel (`github.com/larsartmann/go-appkit/otel`)

| Feature                                                  | Status           | Evidence                              |
| -------------------------------------------------------- | ---------------- | ------------------------------------- |
| Provider setup (`Setup`, options incl. `WithSampler`)    | FULLY_FUNCTIONAL | `setup.go`, `setup_test.go`           |
| Flush-safe shutdown (ForceFlush before Shutdown)         | FULLY_FUNCTIONAL | `setup.go:135`, `setup_test.go`       |
| otelhttp middleware bridge (pattern-named server spans)  | FULLY_FUNCTIONAL | `middleware.go`, `middleware_test.go` |
| W3C trace-context + baggage propagation                  | FULLY_FUNCTIONAL | `attributes.go`, `middleware_test.go` |
| Health-endpoint tracing/metrics filter (unconditional)   | FULLY_FUNCTIONAL | `middleware.go:170`                   |
| Public-endpoint mode (remote parents → links)            | FULLY_FUNCTIONAL | `middleware_test.go`                  |
| Custom path/predicate filters                            | FULLY_FUNCTIONAL | `middleware_test.go`                  |
| Route-attributed, cardinality-safe HTTP metrics          | FULLY_FUNCTIONAL | `metrics_test.go`                     |
| Semconv histogram views (`http.server.request.duration`) | FULLY_FUNCTIONAL | `views.go`, `metrics_test.go`         |
| Trace-correlated logging (`TraceHandler`, ID helpers)    | FULLY_FUNCTIONAL | `logging.go`, `logging_test.go`       |
| Strictly opt-in no-op mode (no provider → pass-through)  | FULLY_FUNCTIONAL | `middleware_test.go`                  |
| Runnable example (PORT-aware, E2E-verified)              | FULLY_FUNCTIONAL | `example/main.go`                     |

Known limitation: httputil's `Logging` middleware emits the request-completion
line without request context, so only handler-level logs correlate with spans
(documented in `doc.go`).

## Consumers

Reference consumer: **[cqrs-htmx](https://github.com/LarsArtmann/cqrs-htmx)
`setup/v4`** — its one-call composition root for event-sourced full-stack apps.

- ADR-001 ("appkit-as-foundation") decided and spike-validated: `RunWithAppkit`
  drives an `appkit.Service` with `NoTimeout` (SSE-safe), `ReadyCheck` wired to
  projection readiness, and the appkit middleware stack wrapped outside the
  bundle's domain chain. Verified equivalences live in their
  `setup/run_appkit_test.go` (SSE header flush through the full stack, drain
  readiness transitions, response parity, hardened adoption benchmark).
- Consumed version: `go-appkit v0.3.0` from the module proxy (their dev
  `replace` stripped at their v4.9.0 train; re-verified 2026-09-04).
- Their fold-in (flipping `RunHandler` internals to appkit) is unblocked and
  pending on the cqrs-htmx side (`docs/planning/2026-08-30_appkit-foldin-revalidation.md`).

All five released modules verified as fresh proxy consumers on 2026-09-04
(blank-import smoke modules in clean dirs, `go build` green):
core v0.3.0, cqrs v0.3.0, realtime v0.1.0, flightrecorder v0.1.0,
flightrecorderhealth v0.1.0.
