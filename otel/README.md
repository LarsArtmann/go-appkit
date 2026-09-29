# otel — OpenTelemetry for go-appkit services

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-appkit/otel.svg)](https://pkg.go.dev/github.com/larsartmann/go-appkit/otel)

Opt-in OpenTelemetry instrumentation for HTTP services: provider setup (with
env-driven OTLP export), an `otelhttp` middleware bridge (spans +
semantic-convention metrics + W3C propagation), panic/error exception
recording for backends like SigNoz, HTTP histogram views, import-ready
SigNoz dashboards, and slog trace correlation.

```bash
go get github.com/larsartmann/go-appkit/otel
```

Import with the alias `appkitotel` (the package is named `otel`, like
go.opentelemetry.io/otel):

```go
import appkitotel "github.com/larsartmann/go-appkit/otel"
```

The library code has no go-appkit core dependency — every piece works on
plain `net/http`. The module's `example/` wires it into an appkit service.

## Quick Start

Zero code, environment-driven — the deployment decides where telemetry goes:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"  # SigNoz OTLP/HTTP
export OTEL_SERVICE_NAME="myapp"
```

```go
provider, err := appkitotel.Setup()
if err != nil {
    return err
}

cfg := appkit.DefaultServiceConfig()
cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
cfg.ExtraMiddlewares = []httputil.Middleware{appkitotel.Recovery(logger)}
cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}

logger := slog.New(appkitotel.TraceHandler(slog.NewJSONHandler(os.Stdout, nil)))

svc, err := appkit.NewService(cfg)
// ... register handlers; log via logger.InfoContext(r.Context(), ...) ...
err = svc.Run(ctx)
```

That is the whole wiring: one span per request, semantic-convention metrics,
W3C propagation in and out, trace IDs on handler logs, OTLP export to the
collector named in the environment, and a provider flush that runs after the
server released its connections during graceful shutdown. Run the example:
`go run ./example`.

Explicit code wins per signal when you want it —
`WithOTLP(WithOTLPEndpoint("http://localhost:4318"))` pins the endpoint,
`WithStdoutExporter(os.Stdout)` prints spans instead, `WithMetricReader`
swaps the metric pipeline.

## SigNoz in 3 steps

1. **Run SigNoz** (OTLP on `:4318`, UI on `:3301`):

   ```bash
   git clone -b main https://github.com/SigNoz/signoz.git && cd signoz/deploy && ./install.sh
   ```

2. **Point your app at it** — either of:

   ```bash
   export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"   # zero code
   ```

   ```go
   appkitotel.Setup(appkitotel.WithService("myapp", "1.0.0", pod),
       appkitotel.WithOTLP(appkitotel.WithOTLPEndpoint("http://localhost:4318")))
   ```

   SigNoz Cloud instead of local: keep the endpoint pattern
   (`https://ingest.<region>.signoz.cloud:443`) and add
   `WithOTLPHeaders(map[string]string{"signoz-ingestion-key": key})` — or
   `OTEL_EXPORTER_OTLP_HEADERS`.

3. **Import the dashboard** — `dashboards/appkit-http-dashboard.json` (UI:
   Dashboards → New Dashboard → Import JSON). See
   `dashboards/README.md`.

## Exceptions (SigNoz's Exceptions view)

SigNoz turns `exception` span events into a dedicated exceptions view with
stack traces and trace links. This module produces them two ways:

- **Panics** — `Recovery(logger)` in `ExtraMiddlewares` records
  `exception.type` / `exception.message` / `exception.stacktrace` plus error
  span status, logs the familiar `panic recovered` line, and answers 500. It
  is a drop-in sibling of httputil's Recovery; nil logger falls back to
  `slog.Default()`.
- **Handled errors** — one line in any handler:

  ```go
  if err != nil {
      appkitotel.RecordError(r.Context(), err)
      http.Error(w, "checkout failed", http.StatusInternalServerError)
      return
  }
  ```

Gotcha: when `Recovery` runs inside the span middleware, otelhttp overwrites
the span status *description* at span end — the exception event (type,
message, stacktrace) is what survives, and that is exactly what the
exceptions view indexes.

## What you get

| Signal    | Instrument                                           | Notes                                                   |
| --------- | ---------------------------------------------------- | ------------------------------------------------------- |
| Traces    | one SERVER span per request                          | named after the ServeMux pattern (`GET /users/{id}`)    |
| Traces    | W3C `traceparent`/`baggage` in and out               | continues caller traces; feeds downstream calls         |
| Traces    | `exception` events (panics + handled errors)         | `Recovery` + `RecordError`; feeds SigNoz's Exceptions view |
| Metrics   | `http.server.request.duration` (+ size, active)      | method/route/status attributes; route-based, no blowups |
| Logs      | `trace_id` + `span_id` on records logged with ctx    | `TraceHandler` decorates any `slog.Handler`             |
| Export    | OTLP/HTTP for traces + metrics, env-driven or in code | `WithOTLP`; `OTEL_EXPORTER_OTLP_*` natively honored    |
| Lifecycle | provider `Shutdown` in `ServiceConfig.ShutdownHooks` | flush after drain — spans cover the final requests      |

## Options that matter

- `Middleware(WithPublicEndpoint())` — internet-facing APIs: incoming
  `traceparent` headers become links, not parents (untrusted callers cannot
  forge trace continuity or skew sampling).
- `Middleware(WithFilteredPaths("/metrics", "/static/"))` — health endpoints
  are filtered by default; extend for other chatty paths.
- `Setup(WithSampler(sdktrace.TraceIDRatioBased(0.1)))` — head sampling in
  code; without it, `OTEL_TRACES_SAMPLER`/`_ARG` configure sampling from the
  deployment (default: parent-based always-sample).
- `Setup(WithEnvironment("production"))` — the `deployment.environment`
  resource attribute SigNoz filters by; `OTEL_RESOURCE_ATTRIBUTES` works too.
- `Setup(WithoutGlobalRegistration())` — isolated providers for tests and
  multi-service processes.

## Production exporters

`WithOTLP` is the built-in production path (OTLP/HTTP, traces via batch
processor, metrics via a periodic reader honoring
`OTEL_METRIC_EXPORT_INTERVAL`):

```go
provider, _ := appkitotel.Setup(
    appkitotel.WithService("myapp", version, instance),
    appkitotel.WithOTLP(
        appkitotel.WithOTLPEndpoint("http://otel-collector:4318"), // or env
        appkitotel.WithOTLPHeaders(map[string]string{"signoz-ingestion-key": key}),
    ),
)
```

Unset options defer to the standard `OTEL_EXPORTER_OTLP_*` environment
variables, which the exporters honor natively (endpoint, headers, timeout,
compression, retries). Any custom `sdktrace.SpanExporter` /
`sdkmetric.Reader` still fits via `WithSpanExporter` / `WithMetricReader` —
and takes precedence over OTLP for its signal.

## Relationship to go-cqrs-lite's otel module

Both expose a `Setup` with the same shape and both register the
process-global providers — use exactly ONE of them per process; either wires
both HTTP and CQRS instrumentation. This module exists so plain HTTP
services need no go-cqrs-lite dependency.

For CQRS-side projection metrics with an OTel meter, see the cqrs module
README's observability section.

## Design notes

- **Opt-in and no-op without Setup**: unconfigured, `Middleware` propagates
  nothing and records nothing — near-zero overhead.
- **Route-attributed metrics**: metrics attribute on the matched route pattern
  (`/users/{id}`), never the raw path — works through the documented
  `OuterMiddlewares` wiring (httputil v1.2.0 pattern propagation;
  regression-pinned by the integration module's
  `TestSpanNameAndRouteThroughAppkitOuterMiddlewares`).
- **SSE-safe**: with `WriteTimeout: appkit.NoTimeout`, the request span ends
  when the stream ends — no artificial cutoff.
- **Panic-correct**: `Recovery` records the exception event inside the span
  and answers 500; the outer placement of `Middleware` (`OuterMiddlewares`)
  means both sit inside the span. The `http.ErrAbortHandler` sentinel
  re-panics untouched, matching httputil.Recovery.

## Go build notes

Builds with plain `go build` — unlike six of the seven sibling modules, this
one does not need `GOEXPERIMENT=jsonv2`.

## Performance

`BenchmarkMiddleware_*` (see `benchmark_test.go`; go 1.26.7, linux/amd64,
mean of 10 runs, 2026-09-16; `httptest` round-trip against
`w.WriteHeader(200)`; run-to-run drift on this box is ±25%, see the
2026-09-15 re-run in `TODO_LIST.md`):

| Variant                 | ns/op  | B/op   | allocs/op |
| ----------------------- | ------ | ------ | --------- |
| NoOp (no providers)     | 20,000 | 7,918  | 90        |
| Traced (spans)          | 21,800 | 11,051 | 90        |
| Traced + metered (full) | 23,300 | 11,055 | 90        |

2026-09-29 re-run (n=5, go 1.27.1 floor, httputil v1.4.0): B/op and
allocs/op are UNCHANGED against the table (NoOp 7,973 B/90 allocs; traced
11,092 B/90; traced+metered 11,095 B/90) — no allocation regression from
the toolchain or dependency bump. The ns/op re-baseline is deferred: the
box was under an external nix-build storm (load average 150), which
poisons wall-clock numbers ~2.3x; per the ±25% drift rule below, do not
re-baseline ns/op under load.

Full instrumentation costs ~~3.3us/req (~~+17%) and ~3.1KB over the no-op path,
with zero additional allocations; export I/O is excluded by design (batching
processor, no exporter wired). The no-op baseline is the cost of the
middleware existing in the chain with the module imported but no providers —
the strictly-opt-in guarantee in numbers.
