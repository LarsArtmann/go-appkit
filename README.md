# go-appkit

Production-ready HTTP service framework for Go. Composes [httputil](https://github.com/LarsArtmann/httputil),
[charmbracelet/log](https://github.com/charmbracelet/log), and [go-error-family](https://github.com/LarsArtmann/go-error-family)
into one coherent service lifecycle.

## Quick start

```go
package main

import (
	"context"
	"net/http"
	"os"

	appkit "github.com/larsartmann/go-appkit"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:     ":8080",
		LogLevel: appkit.LogLevelInfo,
	})
	if err != nil {
		panic(err)
	}
	defer svc.Close()

	svc.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	if err := svc.Run(context.Background()); err != nil {
		os.Exit(errorfamily.HandleError(err))
	}
}
```

That gives you:

- HTTP server with graceful drain + shutdown on SIGINT/SIGTERM
- Middleware: Recovery → RequestID → Logging → Timeout → SecurityHeaders
- Health endpoints: `GET /health`, `GET /health/live`, `GET /health/ready`
- Pretty structured logging via charmbracelet/log
- Error classification via go-error-family

## Install

```bash
go get github.com/larsartmann/go-appkit
```

Requires Go 1.26.7 or later (this is the toolchain the framework is
verified on; `encoding/json/v2` must be available, see below).

## Modules in this repository

Each module is independently versioned and usable on its own:

| Module                                        | Import path                                             | What it adds                                                                                                           |
| --------------------------------------------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| [core](README.md) (this module)               | `github.com/larsartmann/go-appkit`                      | Service lifecycle, middleware, health, logging                                                                         |
| [cqrs](cqrs/README.md)                        | `github.com/larsartmann/go-appkit/cqrs`                 | CQRS/event-sourcing over go-cqrs-lite's `system` engine — event store, projections, DLQ, command/query facade, metrics |
| [docs](docs/)                                 | `github.com/larsartmann/go-appkit/docs`                 | Auto-generated OpenAPI/AsyncAPI/D2 docs from Go types                                                                  |
| [errorpages](errorpages/README.md)            | `github.com/larsartmann/go-appkit/errorpages`           | Pretty classified error pages (HTML) and contracts (JSON)                                                              |
| [realtime](realtime/)                         | `github.com/larsartmann/go-appkit/realtime`             | SSE hub + handler: broadcast, replay, heartbeat, Last-Event-ID resume                                                  |
| [otel](otel/)                                 | `github.com/larsartmann/go-appkit/otel`                 | Opt-in OpenTelemetry: provider setup, otelhttp tracing/metrics middleware, trace-correlated logs                       |
| [flightrecorder](flightrecorder/)             | `github.com/larsartmann/go-appkit/flightrecorder`       | On-demand runtime/trace capture middleware + snapshot endpoint                                                         |
| [flightrecorderhealth](flightrecorderhealth/) | `github.com/larsartmann/go-appkit/flightrecorderhealth` | Bridges flight recorder with go-health: dashboard visibility + auto-capture on health failures                         |
| [health](health/)                             | `github.com/larsartmann/go-appkit/health`               | go-health probes (critical/non-critical, startup latch) + real-time health dashboard, one wiring call                  |

> **JSON v2 requirement:** the dependency stack (go-cqrs-lite,
> templ-components, go-health, go-sse) uses `encoding/json/v2`. Go 1.26.7
> enables it **by default** — no configuration needed. On older 1.26.x
> toolchains where it is still gated, prefix builds and tests with
> `GOEXPERIMENT=jsonv2`. Each module's README documents its own history
> with the experiment; the per-module notes in `AGENTS.md` track the
> gory details.
>
> Editor note: gopls running under a toolchain/environment without the
> experiment reports FALSE `string literal not terminated` diagnostics on
> files importing `encoding/json/v2` — set `GOEXPERIMENT=jsonv2` in your
> editor's Go environment before suspecting the source.

## Configuration

All config is via `ServiceConfig`. Zero-value fields get production defaults:

| Field              | Type                    | Default   | Description                                                                        |
| ------------------ | ----------------------- | --------- | ---------------------------------------------------------------------------------- |
| `Addr`             | `string`                | `":8080"` | Listen address                                                                     |
| `LogLevel`         | `LogLevel`              | `"info"`  | Log level: debug, info, warn, error                                                |
| `LogFormat`        | `LogFormat`             | `"auto"`  | Log format: text, json, auto                                                       |
| `ReadTimeout`      | `time.Duration`         | `10s`     | HTTP read timeout                                                                  |
| `WriteTimeout`     | `time.Duration`         | `30s`     | HTTP write timeout                                                                 |
| `IdleTimeout`      | `time.Duration`         | `60s`     | HTTP idle timeout                                                                  |
| `ShutdownTimeout`  | `time.Duration`         | `15s`     | Max time to wait for shutdown                                                      |
| `DrainDelay`       | `time.Duration`         | `5s`      | Delay after flipping ready probe before shutdown; `NoDrainDelay` sentinel skips it |
| `Middlewares`      | `[]httputil.Middleware` | `nil`     | Replace the default middleware stack                                               |
| `ExtraMiddlewares` | `[]httputil.Middleware` | `nil`     | Append to the default middleware stack                                             |
| `OuterMiddlewares` | `[]httputil.Middleware` | `nil`     | Wrap the entire chain (default stack included), outermost — where tracing sits     |
| `DrainHooks`       | `[]func(ctx) error`     | `nil`     | Run once at drain start, while traffic is still served (errors joined)             |
| `ShutdownHooks`    | `[]func(ctx) error`     | `nil`     | Run once after connections are released (e.g. telemetry flush; errors joined)      |
| `RegisterHealth`   | `*bool`                 | `&true`   | Set to `&false` to opt out of health endpoints                                     |
| `ReadyCheck`       | `func() bool`           | `nil`     | Extra readiness gate for `/health/ready` (e.g. `cqrs.EventService.ReadyCheck`)     |
| `Metrics`          | `*MetricsConfig`        | `nil`     | Opt-in Prometheus surface: `GET /metrics` text exposition + request histogram, response totals, in-flight and build-info gauges (see below) |
| `Version`          | `string`                | `""`      | Build version: serves `GET /version` (JSON) and labels the `appkit_build_info` metric |

### Metrics (opt-in Prometheus surface)

Set `cfg.Metrics = &appkit.MetricsConfig{...}` to expose `GET /metrics` in
pure Prometheus text format with ZERO dependencies (no prometheus client,
no otel). Metric names are a stable contract:

```
appkit_http_request_duration_seconds  histogram {method, route, status}
appkit_http_responses_total           counter   {method, route, status}
appkit_http_requests_in_flight        gauge
appkit_build_info                     gauge=1   {version}
```

Route labels carry the ServeMux pattern (`GET /users/{id}`), never raw
paths — cardinality stays bounded; unmatched paths collapse to
`unmatched`. Authentication is mandatory by default: set
`BasicAuthUser`/`BasicAuthPass`, or `AllowUnauthenticated: true`
explicitly (proxy-fronted or loopback-only deployments) — an unauthenticated
config is rejected at construction.

The OTEL `_ratio` exporter trap: OTEL's Prometheus exporter appends `_ratio`
to unit-1 metrics — names from the otel module's exporter do NOT match this
surface's names, and vice versa. Diff metric names when migrating between
the two.

### Log volume

The default level is INFO, and the request-completion line is the only
per-request emission. Measured per-request cost (2026-09-04 benchmark,
output discarded): bare handler ~17µs; log suppressed at WARN ~18µs
(+0.8µs); emitting at INFO ~47µs (+30µs — the charmbracelet line
formatting dominates, 162 allocs/emission). High-throughput services
should set `LogLevel: LogLevelWarn`; the middleware overhead itself is
negligible either way.

## Middleware

The default stack is opinionated but replaceable:

```go
// Use defaults (Recovery → RequestID → Logging → Timeout → SecurityHeaders):
svc, _ := appkit.NewService(appkit.ServiceConfig{Addr: ":8080"})

// Replace the entire stack:
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr:        ":8080",
    Middlewares: []httputil.Middleware{
        httputil.Recovery(logger),
        httputil.RequestID(httputil.DefaultRequestIDConfig()),
    },
})

// Extend the default stack:
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr:             ":8080",
    ExtraMiddlewares: []httputil.Middleware{myMiddleware},
})
```

## Health endpoints

Three endpoints registered by default (via [httputil](https://github.com/LarsArtmann/httputil)):

- `GET /health` — liveness (always 200 `{"status":"up"}`)
- `GET /health/live` — Kubernetes liveness probe
- `GET /health/ready` — Kubernetes readiness probe (flips to 503 during graceful drain;
  also 503 while a configured `ReadyCheck` reports not ready)

The readiness probe is connected to the graceful drain sequence: when `Shutdown` is called,
the probe immediately starts returning 503 so load balancers stop sending traffic before the
server stops accepting connections.

## Lifecycle

```go
// Simple: Run blocks until signal/error, handles shutdown internally.
err := svc.Run(ctx)

// Advanced: Start returns a channel, you manage the lifecycle.
errCh, err := svc.Start()
// ... start other servers, workers, etc.
err := svc.Shutdown(ctx)
```

### Graceful drain sequence

When `Shutdown` is called:

1. Ready probe flips to 503 (load balancer stops sending new traffic)
2. `DrainHooks` run while traffic is still served (e.g. the health module flips its readiness surfaces in lockstep)
3. Wait `DrainDelay` (default 5s) for LB propagation
4. `server.Shutdown(ctx)` stops accepting + finishes in-flight requests
5. `ShutdownHooks` run after connections are released (e.g. telemetry flush)

Each phase emits one INFO log line with its name and duration — `ready_flip`,
`drain_hooks`, `drain_wait` (or a `shutdown phase skipped` line when
`NoDrainDelay` is set), `listener_close`, `shutdown_hooks` — followed by a
`graceful shutdown complete` line carrying the total, so deploys are
diagnosable from logs alone.

## Error handling

Errors use [go-error-family](https://github.com/LarsArtmann/go-error-family) constructors:

```go
import errorfamily "github.com/larsartmann/go-error-family"

// In your handlers — wrap errors at construction:
err := errorfamily.NewRejection("user.not_found", "user %s not found", id)

// appkit re-exports for convenience:
status := appkit.HTTPStatus(err)  // 400 for Rejection
appkit.LogError(err, logger)      // auto-severity (Transient→Warn, others→Error)
```

For HTTP handlers that return errors, use `errorfamily.HTTPHandler`:

```go
svc.Mux.Handle("POST /users", errorfamily.HTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
    // Return a classified error — HTTPHandler maps family→status, writes safe JSON.
    return errorfamily.NewRejection("user.invalid", "name is required")
}))
```

### SSE and other long-lived responses

A fixed `WriteTimeout` (and the default stack's per-request Timeout middleware)
kills streams that outlive it. For SSE (e.g. the `realtime` module) disable both
with the sentinel:

```go
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr:         ":8080",
    WriteTimeout: appkit.NoTimeout, // drops server deadline AND Timeout middleware
    ReadTimeout:  appkit.NoTimeout, // optional; headers/keep-alive reaping stays on
})
```

`ReadHeaderTimeout` and `IdleTimeout` (the slowloris/keep-alive reaping pair)
remain enabled either way.

## When NOT to use appkit

- **You want a specific router** (chi, gin, echo): Use [httputil](https://github.com/LarsArtmann/httputil) directly — it gives you middleware without opinionated server lifecycle.
- **You want type-safe API generation**: Use [Huma](https://github.com/danielgtaylor/huma) and wrap `svc.Mux` with `humago.New`.
- **You want a full-stack framework**: Use [Buffalo](https://gobuffalo.io) or [GoFr](https://gofr.dev).

appkit is for services that use Go stdlib `http.ServeMux` (Go 1.22+ method routing) and want
middleware, health, logging, and shutdown wired in one import.

## Development

Standard Go toolchain:

```bash
go test ./...
go vet ./...
go build ./...
```

## License

PROPRIETARY. All rights reserved — see [LICENSE](LICENSE) (licensing
inquiries: `git@lars.software`). Note: because the license text is not a
classifiable open-source license, pkg.go.dev hides godoc for this module; the
code and this README are the documentation of record.
