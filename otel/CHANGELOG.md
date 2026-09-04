# Changelog

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
