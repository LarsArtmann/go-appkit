# Changelog

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

## [Unreleased]

### Added

- `NewFlightRecorderMetricsHook(meter)`: bridges go-flightrecorder's
  `MetricsHook` into an OTel meter (`appkit_flightrecorder_snapshots_total`
  by source/kind + `appkit_flightrecorder_snapshot_duration_seconds`).
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

### Changed

- Design posture: OTLP/HTTP exporters are now part of this module's
  dependency tree (`otlptracehttp` + `otlpmetrichttp`, otel v1.46.0) — the
  price of one-line/zero-code SigNoz wiring; still no runtime effect until
  Setup asks for export. Hand-built exporters keep working via
  `WithSpanExporter`/`WithMetricReader` and take precedence per signal.
- The sampler is no longer forced to `ParentBased(AlwaysSample)` when
  `WithSampler` is unset — the SDK then honors `OTEL_TRACES_SAMPLER`
  itself, with the same default when the variable is absent.

## [0.1.0] - 2026-09-04

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
