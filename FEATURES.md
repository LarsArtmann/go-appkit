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
| Opt-in Prometheus surface (`Metrics`, `/metrics`) | FULLY_FUNCTIONAL | `metrics.go`, `metrics_test.go`          |
| `Version` + `GET /version` build info       | FULLY_FUNCTIONAL | `version.go`, `metrics_test.go`                  |
| `testkit.Serve` full-chain test harness     | FULLY_FUNCTIONAL | `testkit/testkit.go`, `testkit/testkit_test.go`  |

Shipped in core **v0.5.0** (2026-09-17; additions-only API diff vs v0.4.0,
proxy-tested; v0.5.1 the same day is doc-only — zero API delta).

## cqrs (`github.com/larsartmann/go-appkit/cqrs`)

| Feature                                                                                                        | Status           | Evidence                             |
| -------------------------------------------------------------------------------------------------------------- | ---------------- | ------------------------------------ |
| `EventService` over go-cqrs-lite `system` engine (v0.5.0)                                                      | FULLY_FUNCTIONAL | `eventservice.go`                    |
| Command/query facade (`RegisterDecider`/`RegisterCommand`/`RegisterQuery`, `Dispatch`, `DispatchQueryChecked`) | FULLY_FUNCTIONAL | `commands.go`, `commands_test.go`    |
| Operator config (`DSN`/`Driver`/`Pragmas`, `ConfigPath` YAML+env, `Deployment`)                                | FULLY_FUNCTIONAL | `eventservice.go`, `example_test.go` |
| In-flight command drain on `Shutdown`                                                                          | FULLY_FUNCTIONAL | `eventservice_test.go`               |
| `EventConfig.Logger` worker logging                                                                            | FULLY_FUNCTIONAL | `logger_test.go`                     |
| DLQ (SQLite) + replay/purge                                                                                    | FULLY_FUNCTIONAL | `dlq_test.go`                        |
| `EventConfig.FlightRecorder` (shared `*fr.Recorder`)                                                           | FULLY_FUNCTIONAL | `flightrecorder_test.go`             |
| `EventConfig.Metrics` recorder hook                                                                            | FULLY_FUNCTIONAL | `metrics_test.go`                    |
| Projection readiness + lag accessors (`ReadyCheck` reports NOT-ready before `StartProjections`)                | FULLY_FUNCTIONAL | `readiness_test.go`                  |
| Read-your-writes staleness guards                                                                              | FULLY_FUNCTIONAL | `staleness_test.go` (incl. `TestEventService_CheckStaleness_BudgetMonotonicity`) |
| OTel projection metrics adapter                                                                                | FULLY_FUNCTIONAL | `otelmetrics.go`, real-cycle E2E `otelmetrics_e2e_test.go` |

## docs (`github.com/larsartmann/go-appkit/docs`)

| Feature                               | Status           | Evidence                  |
| ------------------------------------- | ---------------- | ------------------------- |
| Catalog builder (OpenAPI/AsyncAPI/D2) | FULLY_FUNCTIONAL | `docs.go`, `docs_test.go` |
| `RegisterDocs` docserver mounting     | FULLY_FUNCTIONAL | `docs.go:34`              |

Release gap CLOSED 2026-09-16: the historical `docs/v0.2.0` tag was UNFETCHABLE
from the module proxy (module path said `.../docs` while the directory was
`docs-mod/`, so the proxy found no `docs/go.mod` — a ghost release). Fixed by
the path-A repath (`docs-mod/` → `docs/`, project docs moved to `doc/`) and
re-tagged as `docs/v0.3.0`, which is the first fetchable docs release.

## security (`github.com/larsartmann/go-appkit/security`)

| Feature                                                      | Status           | Evidence                                              |
| ------------------------------------------------------------ | ---------------- | ----------------------------------------------------- |
| API-key auth (header any method; `?key=` GET/HEAD only)      | FULLY_FUNCTIONAL | `apikey.go`, `apikey_test.go` (4 CV pins ported)      |
| CSRF + fail-closed API-key bypass                            | FULLY_FUNCTIONAL | `csrf.go`, `csrf_bypass_test.go`                      |
| Keyed rate limits (mandatory `MaxKeys`; 429 aborts chain)    | FULLY_FUNCTIONAL | `ratelimit.go`, `ratelimit_test.go` (429-overwrite pin) |
| Origin check (Origin→Referer, same-origin always allowed)    | FULLY_FUNCTIONAL | `origin_check.go`, `origin_check_test.go`             |
| Typed body limit (`*http.MaxBytesError`)                     | FULLY_FUNCTIONAL | `bodylimit.go`, `bodylimit_test.go`                   |
| Text/URL sanitization (bluemonday; `&not=` trap handled)     | FULLY_FUNCTIONAL | `sanitize.go`, `sanitize_test.go`                     |
| CSP nonce infra + deterministic policy builder (eval NEVER)  | FULLY_FUNCTIONAL | `csp_nonce.go`, `csp.go`, `csp_test.go`               |
| Env-tuned security headers (HSTS production-only)            | FULLY_FUNCTIONAL | `headers.go`, `headers_test.go`                       |

ALL opt-in — nothing joins the default stack (anti-recommendation held).
Ported from the CV production stack per the canonical battery spec (W2).

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
| Snapshot endpoint (`SnapshotHandler`)  | FULLY_FUNCTIONAL | `handler.go`, tests |
| Auto-reset for repeated captures       | FULLY_FUNCTIONAL | `middleware.go`        |
| `SnapshotHandler` 503 on disabled recorder (no silent 200) | FULLY_FUNCTIONAL | `handler.go` (pinned; UNRELEASED at v0.1.0) |
| `OpsRecorderPreset` production options | FULLY_FUNCTIONAL | `middleware.go` (UNRELEASED at v0.1.0) |

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
| Compile-time contract assertion (`dashboard.Prober`)               | FULLY_FUNCTIONAL | `contract_test.go` (UNRELEASED at v0.1.1)          |
| `WithProbeRoutes`+`WithDashboard` conflict semantics tested-as-documented | FULLY_FUNCTIONAL | `mount_test.go` (UNRELEASED at v0.1.1)      |
| `DashboardHardenedPreset` (BasePath + nonce extractor bundled)     | FULLY_FUNCTIONAL | `mount.go`, `TestDashboardHardenedPreset_*` (UNRELEASED at v0.1.1) |
| `NewProbe` batch benchmark (N=1/5/20) + panic-isolation fuzz       | FULLY_FUNCTIONAL | `probe_benchmark_test.go`, `FuzzNewProbe_*` (UNRELEASED at v0.1.1) |
| Runnable godoc examples (criticality grading, panic isolation, F113 two-probe aggregate) | FULLY_FUNCTIONAL | `example_test.go` (UNRELEASED at v0.1.1)   |

## otel (`github.com/larsartmann/go-appkit/otel`)

| Feature                                                  | Status               | Evidence                                                                                                                                                                 |
| -------------------------------------------------------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Provider setup (`Setup`, options incl. `WithSampler`)    | FULLY_FUNCTIONAL     | `setup.go`, `setup_test.go`                                                                                                                                              |
| Flush-safe shutdown (ForceFlush before Shutdown)         | FULLY_FUNCTIONAL     | `setup.go:135`, `setup_test.go`                                                                                                                                          |
| otelhttp middleware bridge (server spans)                | FULLY_FUNCTIONAL     | `middleware.go`, `middleware_test.go`; pattern-named spans survive `OuterMiddlewares` since httputil v1.2.0 (fixed 2026-09-16, regression-pinned by the integration module's `TestSpanNameAndRouteThroughAppkitOuterMiddlewares` against published tags) |
| W3C trace-context + baggage propagation                  | FULLY_FUNCTIONAL     | `attributes.go`, `middleware_test.go`                                                                                                                                    |
| Health-endpoint tracing/metrics filter (unconditional)   | FULLY_FUNCTIONAL     | `middleware.go:170`                                                                                                                                                      |
| Public-endpoint mode (remote parents → links)            | FULLY_FUNCTIONAL     | `middleware_test.go`                                                                                                                                                     |
| Custom path/predicate filters                            | FULLY_FUNCTIONAL     | `middleware_test.go`                                                                                                                                                     |
| Route-attributed, cardinality-safe HTTP metrics          | FULLY_FUNCTIONAL     | `metrics_test.go`; `http.route` survives the documented `OuterMiddlewares` wiring since httputil v1.2.0 (integration-module pin test)                                    |
| Semconv histogram views (`http.server.request.duration`) | FULLY_FUNCTIONAL     | `views.go`, `metrics_test.go`                                                                                                                                            |
| Trace-correlated logging (`TraceHandler`, ID helpers)    | FULLY_FUNCTIONAL     | `logging.go`, `logging_test.go`                                                                                                                                          |
| `NewFlightRecorderMetricsHook` (fr capture events → OTel meter) | FULLY_FUNCTIONAL | `frmetrics.go`, real-capture E2E `frmetrics_e2e_test.go` (UNRELEASED at v0.1.1)                                                                                  |
| Strictly opt-in no-op mode (no provider → pass-through)  | FULLY_FUNCTIONAL     | `middleware_test.go`                                                                                                                                                     |
| Runnable example (PORT-aware, E2E-verified)              | FULLY_FUNCTIONAL     | `example/main.go`                                                                                                                                                        |

Known limitation: httputil's `Logging` middleware emits the request-completion
line without request context, so only handler-level logs correlate with spans
(documented in `doc.go`).

## integration (`github.com/larsartmann/go-appkit/integration` — never released)

Cross-module + cross-repo E2E contracts against PUBLISHED tags only:

| Contract                                                       | Status           | Evidence                                    |
| -------------------------------------------------------------- | ---------------- | ------------------------------------------- |
| SSE header flush through the appkit default stack              | FULLY_FUNCTIONAL | `integration_test.go`                       |
| Journal replay via cqrs-htmx `transport.JournalSSEStore`       | FULLY_FUNCTIONAL | `integration_test.go`                       |
| Span name + `http.route` through `OuterMiddlewares` (pins the 2026-09-16 fix) | FULLY_FUNCTIONAL | `otel_pattern_test.go`                |
| Metrics/version/testkit compose + drain-window ordering (`Addr()` nil inside DrainHooks) | FULLY_FUNCTIONAL | `composition_contract_test.go` |
| One-`Setup`-per-process + errorpages family→status parity      | FULLY_FUNCTIONAL | `contract_parity_test.go`                   |

## Consumers

Reference consumer: **[cqrs-htmx](https://github.com/LarsArtmann/cqrs-htmx)
`setup/v4`** — its one-call composition root for event-sourced full-stack apps.

- ADR-001 ("appkit-as-foundation") decided and spike-validated: `RunWithAppkit`
  drives an `appkit.Service` with `NoTimeout` (SSE-safe), `ReadyCheck` wired to
  projection readiness, and the appkit middleware stack wrapped outside the
  bundle's domain chain. Verified equivalences live in their
  `setup/run_appkit_test.go` (SSE header flush through the full stack, drain
  readiness transitions, response parity, hardened adoption benchmark).
- Consumed version: `go-appkit v0.5.0` from the module proxy
  (`setup/go.mod:95`; their M3 bump landed 2026-09-17, after this repo's
  14:22 audit recorded v0.4.0). Their `Config.Metrics`/`Config.Version`
  threading (M4) is still open on their side; so is the default-flip (b)-(f).
- Their fold-in (flipping `RunHandler` internals to appkit) is unblocked and
  pending on the cqrs-htmx side (their
  `docs/planning/archived/2026-08-30_appkit-foldin-revalidation.md`).

All released modules verified as fresh proxy consumers (blank-import smoke
modules in clean dirs, `go build` green): the 2026-09-04 waves, cqrs v0.5.0
(2026-09-07), the 2026-09-16 train (core v0.4.0, realtime v0.1.1, security
v0.1.0, health v0.1.1, frh v0.1.2, docs v0.3.0, otel v0.1.1), and core
v0.5.0 (2026-09-17, full behavior check: `/version`, `/metrics` auth 401/200,
shutdown phase logs). The historical docs-ghost (`docs/v0.2.0` unfetchable)
was FIXED 2026-09-16 by the path-A repath — `docs/v0.3.0` is the first
fetchable docs release and its pkg.go.dev page RENDERS (verified 2026-09-17,
all module pages render; godoc hidden by the proprietary-license choice).
