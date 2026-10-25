# Changelog

## [Unreleased]

### Changed

- httputil v1.5.0 → v1.6.0; go-error-family v0.11.0 → v0.12.0 (MVS). The
  v1.6.0 tag carries `LoggingFromContext`, which CLOSES the documented
  completion-line limitation: doc.go's log-correlation note now points at
  the fix, and the README gains a "Correlating the request-completion line"
  recipe (mirrors the core default stack with the Logging →
  LoggingFromContext swap; compile-verified against published otel v0.3.0 +
  httputil v1.6.0). NOTE: the tagged httputil v1.6.0 go.mod carries a
  filesystem `replace` (server_timing → ./server_timing) — harmless to
  consumers (go build/tidy verified clean) but a violation of the family's
  own no-replace release rule; upstream ask drafted in
  `docs/feedback/outgoing/2026-10-25_upstream-ask-httputil-v160-replace.md`.
- README: `BenchmarkMiddleware_*` ns/op re-baselined 2026-10-18 (benchstat
  n=10, go 1.27.2, quiet-gate protocol — `scripts/quiet-run.sh`, 8 primary
  threads pinned, load1 11.7): NoOp 19.67µs ± 5%, traced 20.45µs ± 13%,
  traced+metered 25.33µs ± 16% (all within ±9% of the 2026-09-16 numbers —
  the 1.27.2 toolchain jump costs nothing measurable; B/op + allocs/op
  unchanged ±0%). Protocol finding recorded in the README: a busy-based
  "least-idle cores" pin that includes SMT sibling pairs makes these rows
  bimodal (±26-36% sd); the gate helper's picker is SMT-aware now. The
  `Recovery_*` rows stay at their 2026-10-12 values — happy/panic proved
  cross-run unstable (16.35µs↔27.29µs) under the same protocol.
- README: Recovery-bench instability quantified (SUPERB v9 evening probes):
  the 8-core pin was the noise source — GOMAXPROCS=4 + 4 primary-thread
  pins measures ±3-13% within-run and ±12-16% cross-run; GOMAXPROCS=1 is
  worse (±14-39%). The 2026-10-12 table values stay until the 4-core
  protocol repeats at cross-run <±10%.

## [0.3.0] - 2026-10-18

### Added

- `CorrelationID(ctx)` / `WithCorrelationID(ctx, id)` + `CorrelationIDKey`:
  baggage correlation helpers over the W3C baggage the middleware already
  propagates. Read-or-empty, set-never-invent: empty or W3C-invalid values
  leave the context untouched.
- **Exporter off-switch**: `OTEL_TRACES_EXPORTER=none` /
  `OTEL_METRICS_EXPORTER=none` now veto the environment-driven signal in
  `Setup` (case/whitespace tolerant; explicit code still wins). Other
  exporter names stay documented no-ops — the HTTP-only OTLP posture
  stands, this adds a VETO, not an exporter zoo.
- **Process identity in the resource**: `Setup`'s merged resource now
  includes the non-sensitive process detectors — `process.pid`,
  `process.runtime.name`/`version`, `process.executable.name`.
  `process.command_args` is deliberately excluded (secrets ride flags).
- `Transport(opts ...MiddlewareOption)`: instrumented outbound
  `http.RoundTripper` over otelhttp's client path — one CLIENT span per
  downstream call, W3C traceparent/baggage injection continuing the inbound
  trace, client-side semconv metrics. Defaults to the process globals from
  `Setup`; pass-through without one. `WithBaseTransport` takes a tuned base
  RoundTripper; server-side middleware options are documented as not
  applying. README gained an "Outbound calls" section (incl. the
  otelhttptrace `WithoutSubSpans()` high-fanout guidance); the example makes
  one downstream call through it.
- **SDK self-diagnostics through slog**: `Setup` now wires
  `otel.SetErrorHandler` + `otel.SetLogger` — exporter failures surface as
  grep-able `otel export failed` ERROR lines (catalogued in
  `docs/TELEMETRY.md` §1) and SDK-internal records join the normal log
  pipeline, instead of both dying on stderr. `WithDiagnosticsLogger(logger)`
  points the pipeline at the service's logger; without it, `slog.Default()`.
  `WithoutGlobalRegistration()` skips the wiring with the other process
  globals. Depends on `github.com/go-logr/logr` (already an indirect
  require; now direct).

### Documented

- Deploy guide: `otel/deploy.md` — k8s Deployment + Compose snippets for
  the zero-code setup (env names exact against `Setup`'s verified
  contract), the off-switch usage, the tuning-knob table, and the
  what-NOT-to-configure posture. README links it; the stale `go 1.27.1`
  build-note line corrected to 1.27.2.
- README truth pass (2026-10-18): the logs non-goal no longer claims
  `go.opentelemetry.io/otel/log` is experimental — the Logs API+SDK went
  stable in v1.47.0 (2026-10-02), the version this module pins; the
  wait-for-consumer gate is what remains. Added: semconv v1.26.0 pin
  rationale (SigNoz default = old `deployment.environment` key; rename
  trigger), exemplar/cardinality SDK defaults (`TraceBasedFilter`, 2000
  series cap, both env-tunable), the `WithMessageEvents` NOT-DO record
  (evaluated + declined 2026-08-18), and the retired-`GOEXPERIMENT` build
  note refresh.

### Changed

- Test-only: adopted `samber/lo` (v1.53.0, direct require) at the
  metric-name collect per fleet policy's required functional stack;
  runtime code stays hand-rolled (dependency-lean boundary, 2026-10-12
  standing decision).

## [0.2.0] - 2026-10-06


### Added

- `NewFlightRecorderMetricsHook(meter)`: bridges go-flightrecorder's
  `MetricsHook` into an OTel meter (`appkit_flightrecorder_snapshots_total`
  by source/kind/type + `appkit_flightrecorder_snapshot_duration_seconds`).
  Nil-meter returns a no-op hook.
- **SigNoz-grade OTLP export**: `WithOTLP(...)` wires OTLP/HTTP for BOTH
  signals in one option (traces via batch processor, metrics via a periodic
  reader honoring `OTEL_METRIC_EXPORT_INTERVAL`). Options:
  `WithOTLPEndpoint` (base URL; the standard `/v1/traces` + `/v1/metrics`
  paths are appended, so a path on the endpoint becomes a base prefix),
  `WithOTLPHeaders` (e.g. SigNoz Cloud's `signoz-ingestion-key`),
  `WithOTLPTimeout`. Unset options defer to the standard
  `OTEL_EXPORTER_OTLP_*` environment variables, which the exporters honor
  natively.
- **Zero-code setup**: `Setup` auto-enables OTLP per signal when
  `OTEL_EXPORTER_OTLP_ENDPOINT` (or the signal-specific endpoint variables)
  is set and no explicit exporter option owns that signal — explicit code
  always wins. Combined with `OTEL_SERVICE_NAME`,
  `OTEL_RESOURCE_ATTRIBUTES`, and `OTEL_TRACES_SAMPLER`(+`_ARG`) (all
  natively honored now), a bare `Setup()` is fully deployment-configurable.
  `WithEnvironment(name)` adds the `deployment.environment` resource
  attribute SigNoz filters by.
- **Exceptions**: `Recovery(logger)` — panic-recovery middleware that
  records semconv `exception` events (type/message/stacktrace) plus error
  span status on the active span, logs httputil.Recovery's exact
  `panic recovered` line, and answers 500; drop into
  `ServiceConfig.ExtraMiddlewares`. `RecordError(ctx, err)` records handled
  errors the same way from any handler. Both feed SigNoz's Exceptions view.
  `http.ErrAbortHandler` re-panics untouched; nil logger falls back to
  `slog.Default()`.
- `otel/dashboards/appkit-http-dashboard.json`: import-ready SigNoz
  dashboard (schemaVersion v6) on the traces signal — request rate, 5xx
  error rate, p50/p90/p99 latency per route, status-code distribution, top
  endpoints — with a `service.name` variable. See
  `dashboards/README.md`.
- `example/` now demonstrates the zero-code OTLP path (stdout fallback when
  no endpoint env is set), `Recovery` wiring, and both exception flavors
  (`/boom` handled error, `/panic`).
- 20 new tests incl. real OTLP/HTTP round-trips against an in-process
  collector stub (protobuf decode of span + metric exports), env
  auto-enable and precedence, resource env variables, env sampler, and the
  exception/panic contracts. Suite race-green.

- `example/` now wires the flight-recorder bridge end to end: a dir-sink
  recorder built with `fr.WithMetrics(NewFlightRecorderMetricsHook(meter))`,
  a `/slow` route firing captures via `fr.SnapshotIf`, and `rec.Close` in
  the service's ShutdownHooks (drains in-flight async captures before the
  telemetry flush).
- README gains a "Flight-recorder metric bridge" section: signal-table rows
  for both `appkit_flightrecorder_*` metrics, the construction-time wiring
  snippet, and the SigNoz charting hint (chart `snapshots_total` by `type`).

### Changed

- Design posture: OTLP/HTTP exporters are now part of this module's
  dependency tree (`otlptracehttp` + `otlpmetrichttp`, otel v1.46.0) — the
  price of one-line/zero-code SigNoz wiring; still no runtime effect until
  Setup asks for export. Hand-built exporters keep working via
  `WithSpanExporter`/`WithMetricReader` and take precedence per signal.
- The sampler is no longer forced to `ParentBased(AlwaysSample)` when
  `WithSampler` is unset — the SDK then honors `OTEL_TRACES_SAMPLER`
  itself, with the same default when the variable is absent.
- `go-flightrecorder` floor v0.2.0 → v0.2.1 (upstream data-race fix in
  `Reset` vs in-flight async captures; required by any consumer combining
  `SnapshotIfAsync` with `Reset` — the flightrecorder module's middleware
  does exactly that).
- The metrics bridge emits a third attribute, `type` (the
  `TriggerContext.Type` operation label, e.g. `GET /slow`), on both
  `appkit_flightrecorder_*` metrics — dashboards can attribute captures per
  operation; manual captures carry an empty type.

## [0.1.1] - 2026-09-16

### Fixed

- Pattern-named spans (`GET /users/{id}`) and `http.route` metrics now work
  through the documented `OuterMiddlewares` wiring: the fix ships via the
  `httputil v1.2.0` bump (all request-forking middlewares propagate the
  matched pattern back up; regression-pinned upstream by
  `TestPatternPropagation*` and here by the integration module's
  `TestSpanNameAndRouteThroughAppkitOuterMiddlewares` against published
  tags). The README known-issue block is removed in the same change.

### Documented

- Benchmark table re-baselined (n=10, mean±sd) with the dated methodology.

 - 2026-09-04

### Added

- First version of the otel module (package `otel`, import alias
  `appkitotel`): opt-in OpenTelemetry for HTTP services with no go-appkit
  core dependency in the library code (the `example/` wires an appkit
  service; the module therefore requires core for it, with a local
  `replace`).
- `Setup` + `Provider`: one-call TracerProvider/MeterProvider creation with
  W3C propagator registration, service resource attributes, HTTP histogram
  views, and a unified `Shutdown` that force-flushes before shutting down
  (batch-queued spans are not guaranteed to reach an exporter on Shutdown
  alone). Options: `WithService`, `WithSpanExporter`, `WithSampler`,
  `WithMetricReader`, `WithPropagator`, `WithStdoutExporter`,
  `WithoutGlobalRegistration`. API shape mirrors go-cqrs-lite's otel module
  for muscle-memory consistency; a process uses exactly one of the two.
- `Middleware`: `httputil.Middleware` bridging otelhttp v0.68 — one SERVER
  span per request named after the matched ServeMux pattern, W3C
  trace-context/baggage propagation, semantic-convention metrics
  (`http.server.request.duration` with method/route/status attributes) when
  a meter provider is registered. Health endpoints filtered by default.
  Options: `WithTracerProvider`, `WithMeterProvider`, `WithServerName`,
  `WithPublicEndpoint` (remote parents become links), `WithFilter`,
  `WithFilteredPaths`. No-op without a provider — OTel stays opt-in.
- `TraceHandler`: slog.Handler decorator stamping `trace_id`/`span_id` on
  records logged with a span context; records without a span pass through
  unchanged. Plus `TraceIDFromContext`, `SpanIDFromContext`, `ContextLogger`
  (mirroring the cqrs otel module's helpers).
- `NewHTTPViews`: semantic-convention histogram boundaries for
  `http.server.request.duration` (applied by `Setup`; exact-name match so
  size histograms keep SDK defaults).
- `example/main.go`: runnable service demoing spans, correlated logs,
  health filtering, 500-error spans, and flush-on-graceful-shutdown.
- 23 tests: span naming/kind/status, handler-visible span context, W3C
  parent continuity, public-endpoint link semantics, filters, no-op mode,
  route-attributed metrics with cardinality proof, view boundaries, setup
  resource/globals/sampler/shutdown-error-join, log correlation.
