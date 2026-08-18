# otel — OpenTelemetry for go-appkit services

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-appkit/otel.svg)](https://pkg.go.dev/github.com/larsartmann/go-appkit/otel)

Opt-in OpenTelemetry instrumentation for HTTP services: provider setup, an
`otelhttp` middleware bridge (spans + semantic-convention metrics + W3C
propagation), HTTP histogram views, and slog trace correlation.

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

```go
provider, err := appkitotel.Setup(
    appkitotel.WithService("myapp", "1.0.0", os.Getenv("POD_NAME")),
    appkitotel.WithStdoutExporter(os.Stdout), // development; OTLP in production
)
if err != nil {
    return err
}

cfg := appkit.DefaultServiceConfig()
cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}

logger := slog.New(appkitotel.TraceHandler(slog.NewJSONHandler(os.Stdout, nil)))

svc, err := appkit.NewService(cfg)
// ... register handlers; log via logger.InfoContext(r.Context(), ...) ...
err = svc.Run(ctx)
```

That is the whole wiring: one span per request, semantic-convention metrics,
W3C propagation in and out, trace IDs on handler logs, and a provider flush
that runs after the server released its connections during graceful
shutdown. Run the example: `go run ./example`.

## What you get

| Signal    | Instrument                                           | Notes                                                   |
| --------- | ---------------------------------------------------- | ------------------------------------------------------- |
| Traces    | one SERVER span per request                          | named after the ServeMux pattern (`GET /users/{id}`)    |
| Traces    | W3C `traceparent`/`baggage` in and out               | continues caller traces; feeds downstream calls         |
| Metrics   | `http.server.request.duration` (+ size, active)      | method/route/status attributes; route-based, no blowups |
| Logs      | `trace_id` + `span_id` on records logged with ctx    | `TraceHandler` decorates any `slog.Handler`             |
| Lifecycle | provider `Shutdown` in `ServiceConfig.ShutdownHooks` | flush after drain — spans cover the final requests      |

## Options that matter

- `Middleware(WithPublicEndpoint())` — internet-facing APIs: incoming
  `traceparent` headers become links, not parents (untrusted callers cannot
  forge trace continuity or skew sampling).
- `Middleware(WithFilteredPaths("/metrics", "/static/"))` — health endpoints
  are filtered by default; extend for other chatty paths.
- `Setup(WithSampler(sdktrace.TraceIDRatioBased(0.1)))` — head sampling for
  high-volume services (default: parent-based always-sample).
- `Setup(WithoutGlobalRegistration())` — isolated providers for tests and
  multi-service processes.

## Production exporters

`Setup` takes any `sdktrace.SpanExporter` / `sdkmetric.Reader`; OTLP is the
usual choice and stays in your dependency tree, not this module's:

```go
exp, _ := otlptracehttp.New(ctx) // go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
provider, _ := appkitotel.Setup(
    appkitotel.WithService("myapp", version, instance),
    appkitotel.WithSpanExporter(exp),
)
```

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
- **Cardinality safety**: metrics attribute on the matched route pattern
  (`/users/{id}`), never the raw path.
- **SSE-safe**: with `WriteTimeout: appkit.NoTimeout`, the request span ends
  when the stream ends — no artificial cutoff.
- **Panic-correct**: a recovered 500 marks the span status error; the outer
  placement (`OuterMiddlewares`) means Recovery sits inside the span.

## Go build notes

Builds with plain `go build` — unlike six of the seven sibling modules, this
one does not need `GOEXPERIMENT=jsonv2`.
