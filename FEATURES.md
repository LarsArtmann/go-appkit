# Features

Honest inventory by module. Statuses are verified against code and tests, not
aspirations.

| Status               | Meaning                                             |
| -------------------- | --------------------------------------------------- |
| FULLY_FUNCTIONAL     | Code present and working (test suite passes).       |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps or documented limitations. |
| PLANNED              | Designed/documented, no code yet.                   |

## core (`github.com/larsartmann/go-appkit`)

| Feature                                           | Status                                                            | Evidence                                          |
| ------------------------------------------------- | ----------------------------------------------------------------- | ------------------------------------------------- |
| `Service` lifecycle (Run/Start/Shutdown)          | FULLY_FUNCTIONAL                                                  | `service.go`, `service_test.go`                   |
| Graceful drain (ready→503, delay, stop)           | FULLY_FUNCTIONAL                                                  | `service.go:134`                                  |
| Default middleware stack (replace/extend)         | FULLY_FUNCTIONAL                                                  | `middleware.go`                                   |
| Health endpoints `/health`, live, ready           | FULLY_FUNCTIONAL                                                  | `health.go`, `health_test.go`                     |
| `ReadyCheck` external readiness gate              | FULLY_FUNCTIONAL                                                  | `config.go:50`, `service.go:192`                  |
| `OuterMiddlewares` outermost hook                 | FULLY_FUNCTIONAL                                                  | `config.go:65`, `middleware_test.go`              |
| `ShutdownHooks` post-shutdown flush hooks         | FULLY_FUNCTIONAL                                                  | `service.go`, `shutdownhooks_test.go`             |
| `DrainHooks` drain-start readiness hooks          | FULLY_FUNCTIONAL                                                  | `config.go`, `service.go`, `drainhooks_test.go`   |
| `StartHooks` post-listen startup seam             | FULLY_FUNCTIONAL                                                  | `config.go`, `service.go`, `starthooks_test.go`   |
| Shutdown phase logging (per-phase INFO)           | FULLY_FUNCTIONAL                                                  | `service.go` (`logPhase`), `shutdownlog_test.go`  |
| `NoDrainDelay` fast-test shutdown sentinel        | FULLY_FUNCTIONAL                                                  | `config.go:37`, `config_test.go`                  |
| charmbracelet/logging (`InitLogger`)              | FULLY_FUNCTIONAL                                                  | `logger.go`                                       |
| error-family re-exports (`HTTPStatus`…)           | FULLY_FUNCTIONAL                                                  | `errors.go`                                       |
| SSE-safe `WriteTimeout` configuration             | FULLY_FUNCTIONAL                                                  | `config.go` (`NoTimeout`), `notimeout_test.go`    |
| Opt-in Prometheus surface (`Metrics`, `/metrics`) | FULLY_FUNCTIONAL                                                  | `metrics.go`, `metrics_test.go`                   |
| `Version` + `GET /version` build info             | FULLY_FUNCTIONAL                                                  | `version.go`, `metrics_test.go`                   |
| `testkit.Serve` full-chain test harness           | FULLY_FUNCTIONAL                                                  | `testkit/testkit.go`, `testkit/testkit_test.go`   |
| `testkit.NewService` test-safe defaults           | FULLY_FUNCTIONAL                                                  | `testkit/testkit.go`, `testkit_test.go` (v0.9.0)  |
| `testkit.HTTPGetURL`/`HTTPGetContext` GET helpers | FULLY_FUNCTIONAL (v0.10.0)                                        | `testkit/httptest.go`, `testkit/httptest_test.go` |
| `testkit.WaitForRunning`/`WaitForRunningWithin`   | FULLY_FUNCTIONAL (v0.10.0)                                        | `testkit/httptest.go`, `testkit/httptest_test.go` |
| `testkit.FreeAddr` rebind-verified port reserve   | FULLY_FUNCTIONAL (v0.10.0)                                        | `testkit/httptest.go`, `testkit/httptest_test.go` |
| Server-Timing outermost on the default stack      | FULLY_FUNCTIONAL (v0.10.0; `ServerTiming *bool` knob, default on) | `middleware.go`, `servertiming_bench_test.go`     |
| `RegisterClassifications` fleet error registry    | FULLY_FUNCTIONAL (v0.10.0; idempotent, not auto-called)           | `errors.go`, `errors_test.go`                     |

Shipped through core **v0.12.0** (2026-11-07; the composition swap — `Service`
now delegates the server loop to a fresh `httputil.Server` per `Start` via
`NewServerListener`: bind + sync `listen_failed` + `Addr()`/`Running()` views

- six phase lines stay appkit-owned, restart-after-full-shutdown now serves
  again, configs the delegated server rejects surface as `server.config_invalid`
  at `Start`, deadline-less `Close()` inherits `ShutdownTimeout` as its bound;
  ZERO public API delta, `go doc -all` diff empty at root and testkit).
  Prior cut: **v0.11.0** (2026-11-07; behavior/docs-only — hook
  failures classify through `errorfamily.WrapOnce` (own family/code
  preserved, `hook_index` in context), floor 1.27.3, README
  RegisterClassifications/Server-Timing-phase recipes; zero public API
  delta, `go doc -all` diff empty at root and testkit).
  **v0.10.0** (2026-11-06; lifecycle hooks in v0.4.0,
  metrics/version/testkit in v0.5.0, `appkit.Hook` alias +
  `testkit.DrainWindowProbe` in v0.7.0, `ServiceConfig.StartHooks` in v0.8.0,
  `testkit.NewService(tb, cfg)` — the config-taking constructor with test-safe
  defaults (ephemeral loopback addr + `NoDrainDelay`), closing issue #26's
  5s-per-test teardown trap — in v0.9.0; Server-Timing outermost +
  `RegisterClassifications()` + the testkit GET/wait/FreeAddr helpers in
  v0.10.0, released for the preset module's spine; additions-only API diffs,
  proxy-tested).

## cqrs (`github.com/larsartmann/go-appkit/cqrs`)

Shipped through cqrs **v0.9.0** (2026-11-08: folded notifications — the
broadcast-strictly-after-fold contract, W5-C2; additions-only API diff vs
v0.8.0).

| Feature                                                                                                                                                                                                                           | Status           | Evidence                                                                                                                                                                |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `EventService` over go-cqrs-lite `system` engine (v0.5.0)                                                                                                                                                                         | FULLY_FUNCTIONAL | `eventservice.go`                                                                                                                                                       |
| Command/query facade (`RegisterDecider`/`RegisterCommand`/`RegisterQuery`, `Dispatch`, `DispatchQueryChecked`)                                                                                                                    | FULLY_FUNCTIONAL | `commands.go`, `commands_test.go`                                                                                                                                       |
| Operator config (`DSN`/`Driver`/`Pragmas`, `ConfigPath` YAML+env, `Deployment`)                                                                                                                                                   | FULLY_FUNCTIONAL | `eventservice.go`, `example_test.go`                                                                                                                                    |
| In-flight command drain on `Shutdown`                                                                                                                                                                                             | FULLY_FUNCTIONAL | `eventservice_test.go`                                                                                                                                                  |
| `EventConfig.Logger` worker logging                                                                                                                                                                                               | FULLY_FUNCTIONAL | `logger_test.go`                                                                                                                                                        |
| DLQ (SQLite) + replay/purge                                                                                                                                                                                                       | FULLY_FUNCTIONAL | `dlq_test.go`                                                                                                                                                           |
| `EventConfig.FlightRecorder` (shared `*fr.Recorder`)                                                                                                                                                                              | FULLY_FUNCTIONAL | `flightrecorder_test.go`                                                                                                                                                |
| `EventConfig.Metrics` recorder hook                                                                                                                                                                                               | FULLY_FUNCTIONAL | `metrics_test.go`                                                                                                                                                       |
| Projection readiness + lag accessors (`ReadyCheck` reports NOT-ready before `StartProjections`)                                                                                                                                   | FULLY_FUNCTIONAL | `readiness_test.go`                                                                                                                                                     |
| Read-your-writes staleness guards                                                                                                                                                                                                 | FULLY_FUNCTIONAL | `staleness_test.go` (incl. `TestEventService_CheckStaleness_BudgetMonotonicity`)                                                                                        |
| OTel projection metrics adapter                                                                                                                                                                                                   | FULLY_FUNCTIONAL | `otelmetrics.go`, real-cycle E2E `otelmetrics_e2e_test.go`                                                                                                              |
| `EventConfig.Domain` passthrough (metaengine `Lookup`/`QuerySet`/`Count`/`Evolve`, coeffect gate, Timers; pinned merge contract — drain tracker outermost, derived host-options/`CheckpointStore` win, nil Domain byte-identical) | FULLY_FUNCTIONAL | `domain_test.go` (16 tests incl. sqlite-restart persistence + timers tripwire — flipped green 2026-10-09, upstream `Close` gap fixed; v0.7.0 shipped, v0.8.0 re-pinned) |
| Engine health accessors (`HealthCheck`/`EngineHealth`/`ScreamReport`) + boot-time SCREAM logging                                                                                                                                  | FULLY_FUNCTIONAL | `healthaccessors.go`, `eventservice_test.go` (shipped v0.7.0)                                                                                                           |
| Engine-backed default projection checkpoints (aux DB now DLQ-only; ADR-0142)                                                                                                                                                      | FULLY_FUNCTIONAL | `domain_test.go` checkpoint tests (shipped v0.7.0)                                                                                                                      |
| `Shutdown` drains via the system drainer seam and ALWAYS closes (drain + close errors joined)                                                                                                                                     | FULLY_FUNCTIONAL | `eventservice_test.go` (shipped v0.7.0)                                                                                                                                 |
| README recipes pinned to module code + Examples (invented/dangling API refs fail the suite)                                                                                                                                       | FULLY_FUNCTIONAL | `readme_sync_test.go` (shipped v0.7.0 train)                                                                                                                            |
| Folded notifications: `NotifyFolded` wrapper (non-blocking send, drop counter, nil-channel off-switch, failed fold never notifies) + `FoldedEvent` consumed structurally by realtime's `FoldedSignal`                             | FULLY_FUNCTIONAL | `folded.go`, `folded_test.go` (6 tests); E2E discriminator pinned by `integration/folded_e2e_test.go`                                                                   |

## docs (`github.com/larsartmann/go-appkit/docs`)

| Feature                               | Status           | Evidence                  |
| ------------------------------------- | ---------------- | ------------------------- |
| Catalog builder (OpenAPI/AsyncAPI/D2) | FULLY_FUNCTIONAL | `docs.go`, `docs_test.go` |
| `RegisterDocs` docserver mounting     | FULLY_FUNCTIONAL | `docs.go:34`              |

Release gap CLOSED 2026-09-16: the historical `docs/v0.2.0` tag was UNFETCHABLE
from the module proxy (module path said `.../docs` while the directory was
`docs-mod/`, so the proxy found no `docs/go.mod` — a ghost release). Fixed by
the path-A repath (`docs-mod/` → `docs/`, project docs moved to `docs/`) and
re-tagged as `docs/v0.3.0`, which is the first fetchable docs release.

## security (`github.com/larsartmann/go-appkit/security`)

| Feature                                                     | Status           | Evidence                                                |
| ----------------------------------------------------------- | ---------------- | ------------------------------------------------------- |
| API-key auth (header any method; `?key=` GET/HEAD only)     | FULLY_FUNCTIONAL | `apikey.go`, `apikey_test.go` (4 CV pins ported)        |
| CSRF + fail-closed API-key bypass                           | FULLY_FUNCTIONAL | `csrf.go`, `csrf_bypass_test.go`                        |
| Keyed rate limits (mandatory `MaxKeys`; 429 aborts chain)   | FULLY_FUNCTIONAL | `ratelimit.go`, `ratelimit_test.go` (429-overwrite pin) |
| Origin check (Origin→Referer, same-origin always allowed)   | FULLY_FUNCTIONAL | `origin_check.go`, `origin_check_test.go`               |
| Typed body limit (`*http.MaxBytesError`)                    | FULLY_FUNCTIONAL | `bodylimit.go`, `bodylimit_test.go`                     |
| Text/URL sanitization (bluemonday; `&not=` trap handled)    | FULLY_FUNCTIONAL | `sanitize.go`, `sanitize_test.go`                       |
| CSP nonce infra + deterministic policy builder (eval NEVER) | FULLY_FUNCTIONAL | `csp_nonce.go`, `csp.go`, `csp_test.go`                 |
| Env-tuned security headers (HSTS production-only)           | FULLY_FUNCTIONAL | `headers.go`, `headers_test.go`                         |

ALL opt-in — nothing joins the default stack (anti-recommendation held).
Ported from the CV production stack per the canonical battery spec (W2).
Shipped through security **v0.2.0** (2026-09-29: `example/` hardened-chain
demo, `THREAT_MODEL.md`, example-only core dependency). The example dep
tracks the latest core it needs (pinned v0.12.0 — swept forward from the
2026-11-06 v0.10.0 pin by the daemon's dependency sweeps; every core delta
since v0.10.0 is additions-only, so the example
edge stays demo-only and the package itself stays appkit-free).

## errorpages (`github.com/larsartmann/go-appkit/errorpages`)

| Feature                            | Status           | Evidence                             |
| ---------------------------------- | ---------------- | ------------------------------------ |
| Pretty 404/405 (`Mount`/`Wrap`)    | FULLY_FUNCTIONAL | `errorpages_test.go`                 |
| Family-classified error pages      | FULLY_FUNCTIONAL | `errorpages_test.go:25`              |
| JSON contract negotiation          | FULLY_FUNCTIONAL | `errorpages_test.go:57`              |
| Render-failure plain-text fallback | FULLY_FUNCTIONAL | `errorpages_test.go` (failingWriter) |
| stdlib redirect parity in `Wrap`   | FULLY_FUNCTIONAL | `errorpages_test.go:246`             |

Shipped through errorpages **v0.1.3** (2026-11-07: `Mount` 405-swallow
godoc fix — a registered catch-all turns stdlib 405s into pretty 404s,
pinned by integration — plus a dependabot sweep and floor 1.27.3;
doc-only, zero API delta).

## realtime (`github.com/larsartmann/go-appkit/realtime`)

Shipped through realtime **v0.2.0** (2026-11-08: `FoldedSource`/`FoldedSignal`
— folded signals to full-state hub broadcasts, the realtime half of the
W5-C2 folded contract; additions-only API diff vs v0.1.3).

| Feature                                                                                                     | Status           | Evidence                                |
| ----------------------------------------------------------------------------------------------------------- | ---------------- | --------------------------------------- |
| SSE hub (broadcast, patch, shutdown)                                                                        | FULLY_FUNCTIONAL | `hub.go`                                |
| SSE handler (replay, heartbeat, CORS)                                                                       | FULLY_FUNCTIONAL | `handler.go`                            |
| Subscribe-before-replay + live dedup                                                                        | FULLY_FUNCTIONAL | `handler.go`, `realtime_test.go`        |
| Last-Event-ID resume                                                                                        | FULLY_FUNCTIONAL | `handler.go`                            |
| `FoldedSource[S FoldedSignal]` pump (snapshot-after-signal, full-state drop safety, no second drop counter) | FULLY_FUNCTIONAL | `folded.go`, `folded_test.go` (7 tests) |

Known limitation: on event bursts larger than the broadcaster buffer (64),
excess events are dropped and healed by client Last-Event-ID reconnect.

## flightrecorder (`github.com/larsartmann/go-appkit/flightrecorder`)

| Feature                                                                                                                                             | Status           | Evidence                                                                        |
| --------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------------------------------------- |
| Trigger-based trace capture middleware                                                                                                              | FULLY_FUNCTIONAL | `middleware.go`, tests                                                          |
| Snapshot endpoint (`SnapshotHandler`)                                                                                                               | FULLY_FUNCTIONAL | `handler.go`, tests                                                             |
| Auto-reset for repeated captures                                                                                                                    | FULLY_FUNCTIONAL | `middleware.go`                                                                 |
| `SnapshotHandler` 503 on disabled recorder (no silent 200)                                                                                          | FULLY_FUNCTIONAL | `handler.go` (pinned; shipped v0.2.0)                                           |
| `OpsRecorderPreset` production options                                                                                                              | FULLY_FUNCTIONAL | `middleware.go` (shipped v0.2.0)                                                |
| Async capture (`SnapshotIfAsync`, detached ctx — requests never pay trace-write latency; completion telemetry via `fr.WithMetrics` at construction) | FULLY_FUNCTIONAL | `middleware.go`, `TestMiddleware_CaptureIsNonBlocking` (v0.2.0)                 |
| Snapshot download mode (`?download=1`, `WithSnapshotFilename`)                                                                                      | FULLY_FUNCTIONAL | `handler.go`, `TestSnapshotHandler_DownloadMode*` (v0.2.0)                      |
| `OpsRecorderLoggerPreset` (retention failures → service log)                                                                                        | FULLY_FUNCTIONAL | `preset.go`, `TestOpsRecorderLoggerPreset_LoggerHookReceivesLifecycle` (v0.2.0) |
| Runnable example (dir-sink ops preset end-to-end)                                                                                                   | FULLY_FUNCTIONAL | `example/main.go` (v0.2.0)                                                      |

## flightrecorderhealth (`github.com/larsartmann/go-appkit/flightrecorderhealth`)

| Feature                                                                                   | Status           | Evidence                                         |
| ----------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------ |
| `Checkable` (health-dashboard visibility)                                                 | FULLY_FUNCTIONAL | `adapter.go`, 5 tests                            |
| `Trigger` (auto-capture on health failure)                                                | FULLY_FUNCTIONAL | `adapter.go`, 10 tests                           |
| `Register` convenience                                                                    | FULLY_FUNCTIONAL | `adapter.go`, 3 tests                            |
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

| Feature                                                                                  | Status           | Evidence                                                                                                                              |
| ---------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `NewProbe` injector-free probe (concurrent, panic-isolated checks)                       | FULLY_FUNCTIONAL | `probe.go`, `probe_test.go`                                                                                                           |
| Critical/non-critical readiness classification                                           | FULLY_FUNCTIONAL | `TestNewProbe_ClassificationFollowsCriticality`                                                                                       |
| `New` + `RegisterRoutes` / `Mount` mux wiring                                            | FULLY_FUNCTIONAL | `mount.go`, `mount_test.go`                                                                                                           |
| Kubelet probe routes (custom paths opt-in)                                               | FULLY_FUNCTIONAL | `TestMount_ProbeOnlyRegistersKubeletRoutes`                                                                                           |
| Real-time dashboard (HTML/JSON/SSE/metrics/trend) opt-in                                 | FULLY_FUNCTIONAL | `TestMount_WithDashboardServesHTMLJSONAndProbes`                                                                                      |
| Drain lockstep (`Drain` → readiness 503 for the drain window)                            | FULLY_FUNCTIONAL | `TestMount_DrainFlipsDashboardReadiness`, live E2E                                                                                    |
| Lifecycle guards (double-Start rejection, idempotent Shutdown)                           | FULLY_FUNCTIONAL | `TestMount_LifecycleGuardsAndIdempotence`                                                                                             |
| SDK validation errors surface via `Start` (errors.Is preserved)                          | FULLY_FUNCTIONAL | `TestMount_StartPropagatesProbeValidationErrors`                                                                                      |
| Runnable example (verified live: dashboard, probes, drain 503s)                          | FULLY_FUNCTIONAL | `example/main.go`                                                                                                                     |
| Compile-time contract assertion (`dashboard.Prober`)                                     | FULLY_FUNCTIONAL | `contract_test.go` (shipped in v0.1.2)                                                                                                |
| `WithProbeRoutes`+`WithDashboard` conflict semantics tested-as-documented                | FULLY_FUNCTIONAL | `mount_test.go` (shipped in v0.1.2)                                                                                                   |
| `DashboardHardenedPreset` (BasePath + nonce extractor bundled)                           | FULLY_FUNCTIONAL | `mount.go`, `TestDashboardHardenedPreset_*` (shipped in v0.1.2)                                                                       |
| `NewProbe` batch benchmark (N=1/5/20) + panic-isolation fuzz                             | FULLY_FUNCTIONAL | `benchmark_test.go`, `FuzzNewProbe_*` (shipped in v0.1.2)                                                                             |
| Runnable godoc examples (criticality grading, panic isolation, F113 two-probe aggregate) | FULLY_FUNCTIONAL | `example_test.go` (shipped in v0.1.2)                                                                                                 |
| Per-check durations + deadline-bound batches (`NewProbe` → go-health `NewChecks`, #25)   | FULLY_FUNCTIONAL | `probe.go`, `probe_test.go` (v0.2.0, 2026-10-12; serves `duration_ns`, wedged check fails closed at the batch deadline)               |
| Recovered panic fails closed (probe `fail`/readiness 503, not `warn`/200)                | FULLY_FUNCTIONAL | `TestNewProbe_PanicIsolatedPerCheck`, `ExampleNewProbe_panickingCheck` (v0.3.0, 2026-10-27 — go-health v0.8.1 semantic change pinned) |
| `WithHealthRecorder` on the function path fails at boot (`ErrRecorderNotApplicable`)     | FULLY_FUNCTIONAL | `TestNewProbe_RecorderOptionFailsAtBoot` (v0.3.0, 2026-10-27 — the silent zero-trace cliff is gone upstream)                          |
| Dashboard rides go-health-dashboard v0.12.2 (markup a11y, byte-identical CSV export)     | FULLY_FUNCTIONAL | `mount_test.go`, `TestMount_WithDashboardServesHTMLJSONAndProbes` (v0.3.1, 2026-11-06 — dependency alignment, no library API change)  |

Shipped through health **v0.3.2** (2026-11-07: the LICENSE file finally
ships on the module page — every released module root now carries it —
plus floor 1.27.3; file/toolchain only, zero API delta). Prior cut:
**v0.3.1** (2026-11-06 dependency alignment so
consumers can pin a real tag instead of a pseudo-version — go-appkit#54:
go-health-dashboard v0.12.2, go-error-family v0.12.0, templ-components
v1.26.0, httputil v1.8.0, samber/lo direct require; zero API delta).
Prior cut: **v0.3.0** (2026-10-27; go-health v0.8.1 adoption — the
upstream recorder-sentinel ask go-health#7 SHIPPED — plus the
panic-fails-closed flip; additions-only API diff).

## otel (`github.com/larsartmann/go-appkit/otel`)

| Feature                                                                                                                                                     | Status                                                   | Evidence                                                                                                                                                                                                                                                 |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Provider setup (`Setup`, options incl. `WithSampler`)                                                                                                       | FULLY_FUNCTIONAL                                         | `setup.go`, `setup_test.go`                                                                                                                                                                                                                              |
| Flush-safe shutdown (ForceFlush before Shutdown)                                                                                                            | FULLY_FUNCTIONAL                                         | `setup.go:135`, `setup_test.go`                                                                                                                                                                                                                          |
| otelhttp middleware bridge (server spans)                                                                                                                   | FULLY_FUNCTIONAL                                         | `middleware.go`, `middleware_test.go`; pattern-named spans survive `OuterMiddlewares` since httputil v1.2.0 (fixed 2026-09-16, regression-pinned by the integration module's `TestSpanNameAndRouteThroughAppkitOuterMiddlewares` against published tags) |
| W3C trace-context + baggage propagation                                                                                                                     | FULLY_FUNCTIONAL                                         | `attributes.go`, `middleware_test.go`                                                                                                                                                                                                                    |
| Health-endpoint tracing/metrics filter (unconditional)                                                                                                      | FULLY_FUNCTIONAL                                         | `middleware.go:154,180` (unconditional layering comment + `isHealthPass`)                                                                                                                                                                                |
| Public-endpoint mode (remote parents → links)                                                                                                               | FULLY_FUNCTIONAL                                         | `middleware_test.go`                                                                                                                                                                                                                                     |
| Custom path/predicate filters                                                                                                                               | FULLY_FUNCTIONAL                                         | `middleware_test.go`                                                                                                                                                                                                                                     |
| Route-attributed, cardinality-safe HTTP metrics                                                                                                             | FULLY_FUNCTIONAL                                         | `metrics_test.go`; `http.route` survives the documented `OuterMiddlewares` wiring since httputil v1.2.0 (integration-module pin test)                                                                                                                    |
| Semconv histogram views (`http.server.request.duration`)                                                                                                    | FULLY_FUNCTIONAL                                         | `views.go`, `metrics_test.go`                                                                                                                                                                                                                            |
| Trace-correlated logging (`TraceHandler`, ID helpers)                                                                                                       | FULLY_FUNCTIONAL                                         | `logging.go`, `logging_test.go`                                                                                                                                                                                                                          |
| `NewFlightRecorderMetricsHook` (fr capture events → OTel meter, per-op `type` attr)                                                                         | FULLY_FUNCTIONAL                                         | `frmetrics.go`, real-capture E2E `frmetrics_e2e_test.go` (shipped v0.2.0)                                                                                                                                                                                |
| Strictly opt-in no-op mode (no provider → pass-through)                                                                                                     | FULLY_FUNCTIONAL                                         | `middleware_test.go`                                                                                                                                                                                                                                     |
| Runnable example (PORT-aware, E2E-verified)                                                                                                                 | FULLY_FUNCTIONAL                                         | `example/main.go`                                                                                                                                                                                                                                        |
| OTLP/HTTP export, both signals, one option (`WithOTLP` + endpoint/headers/timeout; base-URL path-prefix semantics)                                          | FULLY_FUNCTIONAL                                         | `otlp.go`, real OTLP/HTTP round-trips vs in-process collector stub in `otlp_test.go` (shipped v0.2.0)                                                                                                                                                    |
| Zero-code env setup (`OTEL_EXPORTER_OTLP_ENDPOINT` auto-enable per signal, explicit code wins; env-native service name/resource/sampler; `WithEnvironment`) | FULLY_FUNCTIONAL                                         | `setup.go`, `setup_test.go` (shipped v0.2.0)                                                                                                                                                                                                             |
| Exception capture (`Recovery(logger)` panic middleware + `RecordError` → semconv exception events, the SigNoz Exceptions-view feed)                         | FULLY_FUNCTIONAL                                         | `exceptions.go`, 7 tests in `exceptions_test.go` (shipped v0.2.0)                                                                                                                                                                                        |
| Importable SigNoz dashboard (`otel/dashboards/appkit-http-dashboard.json`)                                                                                  | FULLY_FUNCTIONAL (unvalidated against live SigNoz)       | `dashboards/appkit-http-dashboard.json` + `dashboards/README.md`; template-derived, JSON-parse-verified only (shipped v0.2.0)                                                                                                                            |
| SDK self-diagnostics through slog (`Setup` wires `otel.SetErrorHandler` + `otel.SetLogger`; `WithDiagnosticsLogger`, `otel export failed` log contract)     | FULLY_FUNCTIONAL                                         | `diagnostics.go`, `diagnostics_test.go` (shipped v0.3.0); catalogued in `docs/TELEMETRY.md` §1                                                                                                                                                           |
| Outbound client transport (`Transport(opts...)` — CLIENT spans, W3C injection, client metrics; `WithBaseTransport`)                                         | FULLY_FUNCTIONAL                                         | `client.go`, `client_test.go` (shipped v0.3.0)                                                                                                                                                                                                           |
| Process identity in the resource (`process.pid`, `process.runtime.*`, `process.executable.name`; command_args deliberately excluded)                        | FULLY_FUNCTIONAL                                         | `setup.go` `buildResource`, `TestSetup_ResourceCarriesProcessIdentity` (shipped v0.3.0)                                                                                                                                                                  |
| Exporter off-switch (`OTEL_TRACES_EXPORTER=none` / `OTEL_METRICS_EXPORTER=none` veto the env path; explicit code wins; other names documented no-ops)       | FULLY_FUNCTIONAL                                         | `otlp.go` `envExporterOff`, `applyOTLPWiring` tests in `otlp_test.go` (shipped v0.3.0)                                                                                                                                                                   |
| Baggage correlation helpers (`CorrelationID`/`WithCorrelationID`/`CorrelationIDKey` — read-or-empty, set-never-invent)                                      | FULLY_FUNCTIONAL                                         | `baggage.go`, `baggage_test.go` incl. W3C round-trip (shipped v0.3.0)                                                                                                                                                                                    |
| Deploy guide + SigNoz alert example (`otel/deploy.md` k8s/Compose snippets; `dashboards/appkit-5xx-alert.example.json` v2alpha1, OpenAPI-verified grammar)  | FULLY_FUNCTIONAL (alert unvalidated against live SigNoz) | `deploy.md`, `dashboards/` (v0.3.0)                                                                                                                                                                                                                      |

Shipped through otel **v0.4.0** (2026-11-07; semconv re-pin v1.26.0 →
v1.43.0 WITH the `deployment.environment` → `deployment.environment.name`
rename shipped together — the watch row's rule; SigNoz main dual-reads the
old spelling so saved filters keep resolving; doc-only godoc diff,
wire-behavior change → minor).

Known limitation RESOLVED in v0.4.0: httputil v1.6.0 shipped
`LoggingFromContext`, so the request-completion line now carries
trace context (doc.go's note points at the fix; README carries the
Logging → LoggingFromContext recipe).

## systemd (`github.com/larsartmann/go-appkit/systemd`)

Shipped through systemd **v0.2.0** (2026-11-08: `Install` returns the wired
`Hooks` — the failed-start watchdog-stop escape hatch; signature change,
source-compatible in practice, migration note in the CHANGELOG). Prior cut:
**systemd v0.1.1** (2026-11-07: runnable godoc Examples
`ExampleInstall`/`ExampleNew` pinning the append-only Install contract,
README release-state/Install/deps truth pass, floor 1.27.3; zero API
delta). Prior cut: **systemd v0.1.0** (2026-10-07, the same train as core v0.8.0
which carries the `StartHooks` seam it depends on):

| Feature                                                                                                       | Status           | Evidence                                                   |
| ------------------------------------------------------------------------------------------------------------- | ---------------- | ---------------------------------------------------------- |
| `Install(&cfg)` one-call sd_notify lifecycle (append-only)                                                    | FULLY_FUNCTIONAL | `hooks.go`, `hooks_test.go`                                |
| `New(opts...)` manual wiring (`Start`/`Drain`/`Shutdown` hooks)                                               | FULLY_FUNCTIONAL | `hooks.go`                                                 |
| `READY=1` post-listen via `StartHooks` (send failure fails the start)                                         | FULLY_FUNCTIONAL | `hooks_test.go` — first datagram on a real unixgram socket |
| `STOPPING=1` at drain start via `DrainHooks`                                                                  | FULLY_FUNCTIONAL | `hooks_test.go`                                            |
| Watchdog keepalive at `WatchdogSec/2`, pinging to the final phase                                             | FULLY_FUNCTIONAL | `notify.go`, `TestRunWatchdog_PingsUntilStopped`           |
| No-op outside systemd (`NOTIFY_SOCKET` unset)                                                                 | FULLY_FUNCTIONAL | `TestInstall_NoSocketOutsideSystemd`                       |
| `Counters()` notify-activity snapshot (ReadySent/StoppingSent/WatchdogPings/NotifyFailures)                   | FULLY_FUNCTIONAL | `notify.go:48`, `TestNotify_CountersTrackSuccessfulSends`  |
| Full lifecycle through a live `appkit.Service`                                                                | FULLY_FUNCTIONAL | `TestInstall_LifecycleThroughAppkitService`                |
| `Install` returns the wired `Hooks` (failed-start watchdog stop; idempotent, identical to the wired Shutdown) | FULLY_FUNCTIONAL | `hooks.go`, `TestInstall_ReturnsTheWiredHooks`             |

## httpx (`github.com/larsartmann/go-appkit/httpx`)

Shipped through httpx **v0.2.0** (2026-11-08: the W3 remainder batteries
B3/B4/B5/B7 + the module LICENSE riding the tag + go-etag/server and
go-codec as direct requires; additions-only API diff). Prior cut: **v0.1.0**
(2026-11-07, SUPERB v14 C2-C4; the handler-DX battery — B1 ResultHandler,
B2 bind + validation, B9 no-leak errors, and the test-pinned
classification-parity oracle vs errorpages).

| Feature                                                                                                                                           | Status                                | Evidence                                                                                                              |
| ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `ResultHandler` return-(T, error) seam                                                                                                            | FULLY_FUNCTIONAL                      | `result.go`, `result_test.go`                                                                                         |
| Direct-write ownership + ResponseController reachability                                                                                          | FULLY_FUNCTIONAL                      | `resultWriter` (`Unwrap`), `TestResultHandler_DirectWriteTakesOwnership`                                              |
| `ResultHandlerWithValidation` bind battery (`bind.*` codes, 1 MiB cap, `Validate` hook)                                                           | FULLY_FUNCTIONAL                      | `bind.go`, `bind_test.go`                                                                                             |
| No-leak error contract (family gate + leak-pattern scrub + `Public()` opt-out)                                                                    | FULLY_FUNCTIONAL                      | `sanitize.go`, `sanitize_test.go` (leaked-detail matrix)                                                              |
| Classification parity vs errorpages (same status + JSON contract shape per family)                                                                | FULLY_FUNCTIONAL (test-pinned oracle) | `parity_test.go` (test-only dep on errorpages)                                                                        |
| B3 `WrapWriter` ResponseWriter contract (Flush/Hijack forward, `Unwrap` always) + `httpxtest` forwarding probe                                    | FULLY_FUNCTIONAL                      | `writer.go`, `httpxtest/assert.go` (3 buggy archetypes pinned failing)                                                |
| B4 `Conditional` (go-etag/server strong sha256 + IMS-only-when-INM-absent, RFC 9110 precedence; `ETagFunc`, `SkipConditional`, `SetLastModified`) | FULLY_FUNCTIONAL                      | `conditional.go`, `conditional_test.go` (CV matrix); E2E through the default stack in `integration/httpx_e2e_test.go` |
| B5 content negotiation on go-codec (`Negotiate`, `RequestCodec` 415, `ResponseCodec`)                                                             | FULLY_FUNCTIONAL                      | `negotiate.go`, `negotiate_test.go`                                                                                   |
| B7 `WriteDeadline(extension, prefixes...)` per-route deadline extension (loud failure on opaque chains)                                           | FULLY_FUNCTIONAL                      | `writedeadline.go`, `writedeadline_test.go` (real-server E2E)                                                         |

## config (`github.com/larsartmann/go-appkit/config`)

Shipped through config **v0.1.1** (2026-11-07; x/sys v0.1.0 → v0.49.0
clears GO-2026-5024 from the module-level scan record entirely + the
collectFields complexity refactor that v0.1.0 shipped lint-red; zero API
delta).
Prior cut: **config v0.1.0** (2026-11-06, SUPERB v12 C-train, battery F1; tag
on origin, pushed 06:59):

| Feature                                                                                                  | Status           | Evidence                                          |
| -------------------------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------- |
| `Load[T]` / `Loader[T]` env-first struct-tag binding, precedence flag > env > file > default             | FULLY_FUNCTIONAL | `config.go`, precedence-matrix tests              |
| Tag grammar: `env` / `,required` / `default` / nested `env:"PREFIX"` prefixes / embedded promotion       | FULLY_FUNCTIONAL | `fields.go`, embedded + nested tests              |
| Untagged fields never touched (embedding `appkit.ServiceConfig` is safe)                                 | FULLY_FUNCTIONAL | `fields.go`, safety test                          |
| Scalar parsers: string/bool/int+uint widths/float32-64/Duration/[]string; others fail `unsupported_type` | FULLY_FUNCTIONAL | `parse.go`, scalar round-trip + unsupported tests |
| File layer: YAML + JSON, strict unknown-key rejection, dotted nested key space                           | FULLY_FUNCTIONAL | `file.go`, file-layer tests                       |
| One Rejection wrapping typed `FieldError`s; `FieldErrorsOf` collects the full tree                       | FULLY_FUNCTIONAL | `config.go`, error-taxonomy tests                 |
| `Keys[T]` sorted declared-key discoverability                                                            | FULLY_FUNCTIONAL | `config.go`, Keys tests                           |
| `WithValidate` cross-field hook (runs last)                                                              | FULLY_FUNCTIONAL | validation-hook test                              |

## preset (`github.com/larsartmann/go-appkit/preset`)

Shipped as **preset v0.1.0** (2026-11-06, SUPERB v12/v13 P-train — the one-call
composition module):

| Feature                                                                                                                                                                                                                          | Status           | Evidence                                    |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------- |
| `New(Config)` always-on spine: core stack, systemd sd_notify, health dashboard (owns `/health` + `/healthz` `/readyz` `/startupz`; core's family only under `NoDashboard`), errorpages, `RegisterClassifications`, version stamp | FULLY_FUNCTIONAL | `preset.go`, `spine_test.go`                |
| Env-switch legs `APPKIT_OTEL`/`APPKIT_REALTIME`/`APPKIT_ERRORPAGES`/`APPKIT_SECURITY` with explicit-code-wins precedence, fail-fast Rejection on invalid values                                                                  | FULLY_FUNCTIONAL | `env.go`, `env_test.go`, `conflict_test.go` |
| Hooks passthrough consumer-first (consumer hooks run before spine closures in every phase)                                                                                                                                       | FULLY_FUNCTIONAL | `preset.go`, hook-order test                |
| Realtime leg: hub at `/events` (configurable pattern), `Hub` out-field, graceful SSE shutdown                                                                                                                                    | FULLY_FUNCTIONAL | `realtime_test.go`                          |
| Security profiles: standard (HSTS/nosniff/body-limit) and strict (origin allowlist, boot-rejects without origins)                                                                                                                | FULLY_FUNCTIONAL | `security_test.go`                          |
| otel leg: provider Setup + middleware + shutdown flush; code wins over env; off by default                                                                                                                                       | FULLY_FUNCTIONAL | `otel_test.go`                              |
| Metrics passthrough verbatim (never invents credentials)                                                                                                                                                                         | FULLY_FUNCTIONAL | `metrics_test.go`                           |
| 35-LOC example app (acceptance: full production posture)                                                                                                                                                                         | FULLY_FUNCTIONAL | `example/main.go`, example-smoke green      |

## integration (`github.com/larsartmann/go-appkit/integration` — never released)

Cross-module + cross-repo E2E contracts against PUBLISHED tags only:

| Contract                                                                                                                               | Status           | Evidence                       |
| -------------------------------------------------------------------------------------------------------------------------------------- | ---------------- | ------------------------------ |
| SSE header flush through the appkit default stack                                                                                      | FULLY_FUNCTIONAL | `integration_test.go`          |
| Journal replay via cqrs-htmx `transport.JournalSSEStore`                                                                               | FULLY_FUNCTIONAL | `integration_test.go`          |
| Span name + `http.route` through `OuterMiddlewares` (pins the 2026-09-16 fix)                                                          | FULLY_FUNCTIONAL | `otel_pattern_test.go`         |
| Metrics/version/testkit compose + drain-window ordering (`Addr()` nil inside DrainHooks)                                               | FULLY_FUNCTIONAL | `composition_contract_test.go` |
| Rate-limit in front of SSE (429 + Retry-After aborts the chain)                                                                        | FULLY_FUNCTIONAL | `security_realtime_test.go`    |
| One-`Setup`-per-process + errorpages family→status parity                                                                              | FULLY_FUNCTIONAL | `contract_parity_test.go`      |
| SSE×health drain choreography (hub close frees Shutdown, /readyz 503, clean EOF)                                                       | FULLY_FUNCTIONAL | `sse_health_drain_test.go`     |
| cqrs domain SSE through a live Service (Domain passthrough + journal replay)                                                           | FULLY_FUNCTIONAL | `domain_sse_test.go`           |
| systemd lifecycle through a live Service (unixgram socket transport)                                                                   | FULLY_FUNCTIONAL | `systemd_lifecycle_test.go`    |
| docs docserver composition through the full default chain (openapi/asyncapi/catalog.json)                                              | FULLY_FUNCTIONAL | `docs_composition_test.go`     |
| Folded-contract discriminator: post-fold broadcast carries the mid-fold marker; naive pre-fold wiring fails the same assertion (W5-C2) | FULLY_FUNCTIONAL | `folded_e2e_test.go`           |
| httpx B4 Conditional through the default stack (ETag 200 → If-None-Match 304, Server-Timing still present)                             | FULLY_FUNCTIONAL | `httpx_e2e_test.go`            |

## Consumers

Reference consumer: **[cqrs-htmx](https://github.com/LarsArtmann/cqrs-htmx)
`setup/v4`** — its one-call composition root for event-sourced full-stack apps.

- ADR-001 ("appkit-as-foundation") decided and spike-validated: `RunWithAppkit`
  drives an `appkit.Service` with `NoTimeout` (SSE-safe), `ReadyCheck` wired to
  projection readiness, and the appkit middleware stack wrapped outside the
  bundle's domain chain. Verified equivalences live in their
  `setup/run_appkit_test.go` (SSE header flush through the full stack, drain
  readiness transitions, response parity, hardened adoption benchmark).
- Consumed version: **`go-appkit v0.10.0`** — verified in the PUBLISHED `setup/v4.17.0` tag (2026-11-07; they adopted core v0.10.0 same-day; prior pin was v0.9.0 through setup/v4.16.0)
  (`setup/go.mod`; guard-checked by `scripts/check-consumer-claims.sh`).
  **M4 is SHIPPED in that tag**:
  `Config.Metrics`/`Config.Version` thread into `appkit.ServiceConfig` on the
  `RunWithAppkit` path (`run_appkit.go:78-79`). Still theirs: the
  `Bundle.Run`/`Bundle.RunHandler` internals still serve via httputil
  (Metrics/Version are RunWithAppkit-only there), and the v5-window revisit
  (ADR-0052). Their `go.work` go-etag stub-replace is GONE — setup/v4.15.0+
  carries the split go-etag/entitytag+server v0.6.1.
- Consumer #2: **[PapDashboard](https://github.com/LarsArtmann/PapDashboard)
  `v0.3.0`** — published 2026-10-13; `go.mod` requires **`go-appkit v0.4.0`**
  (core only, plain Service usage, TLS envs removed with the adoption; their
  repo owns version refresh; guard-checked by `scripts/check-consumer-claims.sh`).

All released modules verified as fresh proxy consumers (blank-import smoke
modules in clean dirs, `go build` green): the 2026-09-04 waves, cqrs v0.5.0
(2026-09-07), the 2026-09-16 train (core v0.4.0, realtime v0.1.1, security
v0.1.0, health v0.1.1, frh v0.1.2, docs v0.3.0, otel v0.1.1), and core
v0.5.0 (2026-09-17, full behavior check: `/version`, `/metrics` auth 401/200,
shutdown phase logs). The historical docs-ghost (`docs/v0.2.0` unfetchable)
was FIXED 2026-09-16 by the path-A repath — `docs/v0.3.0` is the first
fetchable docs release and its pkg.go.dev page RENDERS (verified 2026-09-17,
all module pages render; godoc hidden by the proprietary-license choice).
