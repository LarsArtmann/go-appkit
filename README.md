# go-appkit

Production-ready HTTP service framework for Go. Composes [httputil](https://github.com/LarsArtmann/httputil),
[charmbracelet/log](https://github.com/charmbracelet/log), and [go-error-family](https://github.com/LarsArtmann/go-error-family)
into one coherent service lifecycle.

> **The fleet default.** A NEW fleet service starts on go-appkit — owner
> ruling, 2026-11-21. The `preset` module is the one-call base; the
> [go-service-template](https://github.com/LarsArtmann/go-service-template)
> repo scaffolds everything around it. The "Start here" section below is the
> front door.

### Verify these claims

Run, don't trust — these blocks execute green before every release:

```bash
GOWORK=off GOTOOLCHAIN=go1.27.4 go test ./... -race -count=1   # the suite (root + testkit)
PORT=18099 go run ./example &                                   # the live tour (bash)
curl -s localhost:18099/health; curl -si localhost:18099/ | grep -i server-timing
kill -TERM %1                                                   # watch the shutdown phase lines
./scripts/example-smoke.sh                                      # the mechanical version: all examples, boot→probe→graceful-stop
```

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
- Middleware: ServerTiming → Recovery → RequestID → Logging → Timeout → SecurityHeaders (W3C `Server-Timing` header via the family's [httputil/server_timing](https://github.com/LarsArtmann/httputil) — every appkit service exposes request timings to DevTools/APMs out of the box; replace `Middlewares` to drop it)
- Health endpoints: `GET /health`, `GET /health/live`, `GET /health/ready`
- A route registry: register through `svc.HandleFunc` / `svc.Handle` instead of
  `svc.Mux` and `GET /endpoints` can list every route (see "Endpoint listing")
- Pretty structured logging via charmbracelet/log
- Error classification via go-error-family

## Install

```bash
go get github.com/larsartmann/go-appkit
```

Requires Go 1.27 or later (the core module declares the `go 1.27` floor;
satellite modules declare their dependency-forced `go mod tidy` minimums
between `go 1.27.2` and `go 1.27.4` — `go get` demands the highest directive
in the chain until the family re-tags at the floor; the repo builds and tests
on the patched `go1.27.4` toolchain).

## Modules in this repository

Each module is independently versioned and usable on its own:

| Module                                        | Import path                                             | What it adds                                                                                                                            |
| --------------------------------------------- | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| [core](README.md) (this module)               | `github.com/larsartmann/go-appkit`                      | Service lifecycle, middleware, health, logging                                                                                          |
| [cqrs](cqrs/README.md)                        | `github.com/larsartmann/go-appkit/cqrs`                 | CQRS/event-sourcing over go-cqrs-lite's `system` engine — event store, projections, DLQ, command/query facade, metrics                  |
| [docs](docs/)                                 | `github.com/larsartmann/go-appkit/docs`                 | Auto-generated OpenAPI/AsyncAPI/D2 docs from Go types                                                                                   |
| [errorpages](errorpages/README.md)            | `github.com/larsartmann/go-appkit/errorpages`           | Pretty classified error pages (HTML) and contracts (JSON)                                                                               |
| [realtime](realtime/)                         | `github.com/larsartmann/go-appkit/realtime`             | SSE hub + handler: broadcast, replay, heartbeat, Last-Event-ID resume                                                                   |
| [otel](otel/)                                 | `github.com/larsartmann/go-appkit/otel`                 | Opt-in OpenTelemetry: provider setup, otelhttp tracing/metrics middleware, trace-correlated logs                                        |
| [flightrecorder](flightrecorder/)             | `github.com/larsartmann/go-appkit/flightrecorder`       | On-demand runtime/trace capture middleware + snapshot endpoint                                                                          |
| [flightrecorderhealth](flightrecorderhealth/) | `github.com/larsartmann/go-appkit/flightrecorderhealth` | Bridges flight recorder with go-health: dashboard visibility + auto-capture on health failures                                          |
| [health](health/)                             | `github.com/larsartmann/go-appkit/health`               | go-health probes (critical/non-critical, startup latch) + real-time health dashboard, one wiring call                                   |
| [security](security/)                         | `github.com/larsartmann/go-appkit/security`             | Hardening middleware: CSRF, rate limiting, body limits, sanitization, CSP nonce, security headers                                       |
| [systemd](systemd/)                           | `github.com/larsartmann/go-appkit/systemd`              | sd_notify for `Type=notify` units: READY=1 post-listen, STOPPING=1 at drain start, `WatchdogSec/2` keepalive                            |
| [httpx](httpx/README.md)                      | `github.com/larsartmann/go-appkit/httpx`                | Handler DX: `ResultHandler` return-(T, error), bind + validation, no-leak errors, conditional GET, content negotiation, write deadlines |
| [config](config/README.md)                    | `github.com/larsartmann/go-appkit/config`               | Env-first struct-tag configuration: `Load[T]`, precedence flag > env > file > default, strict YAML/JSON file layer                      |
| [preset](preset/README.md)                    | `github.com/larsartmann/go-appkit/preset`               | One-call composition: production spine (core + systemd + health dashboard + errorpages) plus env-switched otel/security/realtime legs   |

> The `security` module is fully opt-in: NOTHING from it sits in core's
> default middleware stack — batteries compose explicitly via
> `ServiceConfig.OuterMiddlewares`/`ExtraMiddlewares` when you want them.

> **JSON v2 requirement:** the dependency stack (go-cqrs-lite,
> templ-components, go-health, go-sse) uses `encoding/json/v2`, which the
> Go 1.27 floor enables **by default** — no configuration, no experiment
> prefix (the per-module `GOEXPERIMENT=jsonv2` notes were retired
> 2026-09-29).

## Start here (a new service)

Three ways in, fastest first. All three get the full lifecycle: graceful
drain, the default middleware stack, health endpoints, and the classified
error contract.

**1. The one-call spine — `preset` (recommended default):**

```go
svc, err := preset.New(preset.Config{
    Addr:    ":8080",
    Version: "1.2.3",
})
// spine: core Service + systemd sd_notify + health dashboard + pretty error
// pages + fleet error classifications; legs via APPKIT_* env switches.
_ = svc.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
    _, _ = w.Write([]byte("hello"))
}, "Say hello")
_ = svc.Run(context.Background())
```

**2. The bare core** — the Quick start above, when you want zero composition
opinion: `appkit.NewService(cfg)` + routes through `svc.HandleFunc` (or
`svc.Mux`) + `svc.Run(ctx)`.

**3. Scaffolding — the [go-service-template](https://github.com/LarsArtmann/go-service-template)
repo**: preset spine + `cmd/<name>` + internal layers + flake/CI/systemd
skeletons, so a new service's first hour is one clone. Refresh-stamped
against the appkit floor by the release ritual.

### The three recipes you'll actually reach for

**HTTP handlers with return-(T, error) — [httpx](httpx/README.md):**

```go
mux.Handle("GET /api/things/{id}", httpx.ResultHandler(
    func(c *httpx.Context) (*Thing, error) {
        return svc.Get(c.Request.Context(), c.PathValue("id"))
    },
))
// nil error → JSON 200 (204 for nil pointer); non-nil → classified error contract.
```

**Events + projections — [cqrs](cqrs/README.md):**

```go
es, err := cqrs.NewEventService(cqrs.EventConfig{DSN: "app.db", Logger: svc.Logger})
_ = es.Host().Register(myProjection)
_ = cqrs.RegisterQuery[TaskQuery, TaskView](es, "task.view", myQueryHandler)
_ = es.StartProjections(ctx)
```

**Config from env/files — [config](config/README.md):** `config.Load[T]()`
with precedence flag > env > file > default and typed field errors.

### One-line wins (replace your hand-rolled X)

| You hand-rolled               | appkit replaces it with                                                                                                                           |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| exit-code mapping             | `os.Exit(errorfamily.HandleError(err))` — classified families map to exit codes                                                                   |
| a version endpoint            | `Version: "1.2.3"` → `GET /version` (JSON) + `appkit_build_info` metric label                                                                     |
| health/diag handlers          | on by default: `GET /health`, `/health/live`, `/health/ready` (drain-aware 503s; `ReadyCheck` adds your gate)                                     |
| API description docs          | the [docs](docs/) module: OpenAPI/AsyncAPI/D2 generated from Go types                                                                             |
| security middleware battery   | the [security](security/) module: CSRF, keyed rate limits, body limits, sanitization, CSP — all opt-in                                            |
| `httptest.NewServer(svc.Mux)` | `testkit.NewService(t, cfg)` — the REAL chain, test-safe defaults, leak-checked teardown (raw-mux tests miss the bugs)                            |
| shutdown orchestration        | nothing to write: SIGINT/SIGTERM → ready-flip → DrainHooks → drain wait → listener close → ShutdownHooks, one grep-able log line each             |
| a hand-maintained route list  | `svc.HandleFunc(pattern, h, summary)` + `Endpoints` → `GET /endpoints`, the live route table (JSON or HTML) — it cannot drift from what is served |

## Configuration

All config is via `ServiceConfig`. Zero-value fields get production defaults:

| Field              | Type                    | Default   | Description                                                                                                                                                       |
| ------------------ | ----------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Addr`             | `string`                | `":8080"` | Listen address                                                                                                                                                    |
| `Listener`         | `net.Listener`          | `nil`     | Serve on a pre-bound listener (unix socket, socket activation) instead of binding `Addr`; ownership + restart rules in the section below                          |
| `LogLevel`         | `LogLevel`              | `"info"`  | Log level: debug, info, warn, error                                                                                                                               |
| `LogFormat`        | `LogFormat`             | `"auto"`  | Log format: text, json, auto                                                                                                                                      |
| `LogHandler`       | `slog.Handler`          | `nil`     | Replace the built-in logger entirely — every line flows through your slog.Handler; `LogLevel`/`LogFormat` ignored (see below)                                     |
| `ReadTimeout`      | `time.Duration`         | `10s`     | HTTP read timeout                                                                                                                                                 |
| `WriteTimeout`     | `time.Duration`         | `30s`     | HTTP write timeout                                                                                                                                                |
| `IdleTimeout`      | `time.Duration`         | `60s`     | HTTP idle timeout                                                                                                                                                 |
| `ShutdownTimeout`  | `time.Duration`         | `15s`     | Max time to wait for shutdown                                                                                                                                     |
| `DrainDelay`       | `time.Duration`         | `5s`      | Delay after flipping ready probe before shutdown; `NoDrainDelay` sentinel skips it                                                                                |
| `Middlewares`      | `[]httputil.Middleware` | `nil`     | Replace the default middleware stack                                                                                                                              |
| `ExtraMiddlewares` | `[]httputil.Middleware` | `nil`     | Append to the default middleware stack                                                                                                                            |
| `OuterMiddlewares` | `[]httputil.Middleware` | `nil`     | Wrap the entire chain (default stack included), outermost — where tracing sits                                                                                    |
| `ServerTiming`     | `*bool`                 | `&true`   | W3C `Server-Timing` header via the default stack's outermost middleware; set to `&false` to omit                                                                  |
| `StartHooks`       | `[]Hook`                | `nil`     | Run once inside `Start`, after the listener binds but before serving (`Addr`/`Running` already live); any error fails the start                                   |
| `DrainHooks`       | `[]func(ctx) error`     | `nil`     | Run once at drain start, while traffic is still served (errors joined)                                                                                            |
| `ShutdownHooks`    | `[]func(ctx) error`     | `nil`     | Run once after connections are released (e.g. telemetry flush; errors joined)                                                                                     |
| `RegisterHealth`   | `*bool`                 | `&true`   | Set to `&false` to opt out of health endpoints                                                                                                                    |
| `ReadyCheck`       | `func() bool`           | `nil`     | Extra readiness gate for `/health/ready` (e.g. `cqrs.EventService.ReadyCheck`)                                                                                    |
| `Metrics`          | `*MetricsConfig`        | `nil`     | Opt-in Prometheus surface: `GET /metrics` text exposition + request histogram, response totals, in-flight and build-info gauges (see below)                       |
| `Endpoints`        | `*EndpointsConfig`      | `nil`     | Opt-in route listing: `GET /endpoints` serves every route registered through `svc.Handle`/`svc.HandleFunc` as JSON or HTML; auth mandatory by default (see below) |
| `Version`          | `string`                | `""`      | Build version: serves `GET /version` (JSON) and labels the `appkit_build_info` metric                                                                             |

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

Enable it (with `Version` stamping `/version` and the build-info label):

```go
svc, err := appkit.NewService(appkit.ServiceConfig{
    Addr:    ":8080",
    Version: "1.2.3", // inject from your build (e.g. GoReleaser ldflags)
    Metrics: &appkit.MetricsConfig{
        Path:          "/metrics",
        BasicAuthUser: "metrics",
        BasicAuthPass: os.Getenv("METRICS_PASS"),
    },
})
```

For tests, the `testkit` sub-package boots the REAL service — full
middleware chain, framework routes and all — and tears it down with a
goroutine-leak assert. Never `httptest.NewServer(svc.Mux)`: a raw-mux
server bypasses the entire chain (three production bugs in the reference
consumer were only visible through the full chain). Build the service
with `testkit.NewService(t, cfg)` — test-safe defaults (ephemeral loopback
address, `DrainDelay: NoDrainDelay`) so teardown does not wait out the 5s
production drain — then register routes and `Serve`:

```go
func TestReady(t *testing.T) {
    svc := testkit.NewService(t, cfg) // test-safe defaults (no 5s drain)

    testkit.Serve(t, svc) // starts svc, registers leak-checked teardown

    resp, err := http.Get("http://" + svc.Addr().String() + "/health/ready")
    // ...
}
```

The kit also carries the shared HTTP/lifecycle helpers — `HTTPGetURL` /
`HTTPGetContext` (noctx-compliant GETs that fatal with the URL in the
message), `WaitForRunning`, and `FreeAddr` (an ephemeral-port reserve that
re-verifies bindability) — so per-module test files stop re-implementing
them. One in-package caveat: a test file INSIDE `package appkit` cannot
import `testkit` (import cycle); whitebox helpers stay local there.

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

### Endpoint listing (opt-in route registry)

`http.ServeMux` cannot enumerate the patterns registered on it, so a route
table is usually a second, hand-maintained list that drifts. The core
service carries a route registry instead
([httputil/endpoints](https://github.com/LarsArtmann/httputil/tree/master/endpoints),
stdlib-only): register routes through the service and the mount and the
listing entry are one declaration:

```go
svc, err := appkit.NewService(appkit.ServiceConfig{
    Addr: ":8080",
    Endpoints: &appkit.EndpointsConfig{
        BasicAuthUser: "ops",
        BasicAuthPass: os.Getenv("ENDPOINTS_PASS"),
    },
})
// ...
err = svc.HandleFunc("GET /users/{id}", getUser, "Fetch one user")
err = svc.Handle("POST /users", http.HandlerFunc(createUser), "Create a user")
```

`GET /endpoints` then serves the whole table, the framework's own routes
(`/health`, `/metrics`, `/version`) included, as JSON for API clients or an
HTML table for browsers (negotiated on `Accept`; `Vary: Accept`,
`Cache-Control: no-store`). `curl -u ops:$ENDPOINTS_PASS :8080/endpoints` is
the one-command answer to "what does this service serve?".

A pattern the mux rejects (malformed, or already registered) comes back as
an error from `Handle`/`HandleFunc` (code `appkit.route_rejected`) instead of
the panic `http.ServeMux` raises. Routes added after construction still
appear: every request reads a fresh snapshot.

Two things to know:

- **Authentication is mandatory by default.** The listing IS the route
  topology. Set `BasicAuthUser`/`BasicAuthPass`, hand the service's own auth
  middleware to `Guard`, or set `AllowUnauthenticated: true` explicitly
  (loopback-only, or a proxy that enforces auth in front). An unprotected
  config is rejected at construction (`appkit.endpoints_config_invalid`).
- **Only routes registered through the service are listed.** `svc.Mux` stays
  public for libraries that mount their own route table (Huma, a framework
  UI); such routes are served but absent from the listing. The registry is
  always populated, listing enabled or not: `svc.Routes.Routes()` returns
  the recorded rows.

### Log volume

The default level is INFO, and the request-completion line is the only
per-request emission. Measured per-request cost (2026-09-04 benchmark,
output discarded): bare handler ~17µs; log suppressed at WARN ~18µs
(+0.8µs); emitting at INFO ~47µs (+30µs — the charmbracelet line
formatting dominates, 162 allocs/emission). High-throughput services
should set `LogLevel: LogLevelWarn`; the middleware overhead itself is
negligible either way.

### Bring your own slog handler

The built-in logger is charmbracelet/log (v1 lane). If your service lives on
another lane — plain `slog`, a v2 logger, a bridged structured logger — supply
`LogHandler` and the ENTIRE lifecycle emits through it: request lines, startup
and shutdown phase lines, everything. `LogLevel`/`LogFormat` are ignored in
that mode (the handler owns level and formatting):

```go
svc, err := appkit.NewService(appkit.ServiceConfig{
    Addr:       ":8080",
    LogHandler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
})
```

The grep-able shutdown contract (`shutdown phase complete` with the phase
name, `graceful shutdown complete` with the total) holds for ANY handler —
it is a message contract, not a formatter contract. See
`docs/planning/2026-11-21_fleet-log-lane-owner-memo.md` for the fleet-lane
ruling this seam was built for.

### Custom listener (unix sockets, socket activation)

Set `Listener` to a pre-bound `net.Listener` and the service serves on it
instead of binding `Addr` — unix-socket IPC daemons, systemd socket
activation, or any listener from another source:

```go
ln, _ := net.Listen("unix", "/run/mydaemon/app.sock")

svc, err := appkit.NewService(appkit.ServiceConfig{
    Listener:   ln, // Addr is ignored; Addr() reports the socket path
    DrainHooks: []appkit.Hook{ /* flip external readiness */ },
})
```

Contract: ownership transfers at `Start` (every synchronous failure path and
graceful shutdown close it, exactly like a self-bound listener); one serve per
listener — a restart after full shutdown needs a FRESH listener; timeouts,
sentinels, and the middleware stack apply unchanged. Design note:
`docs/planning/2026-11-21_listener-injection-seam-design-note.md`.

## Middleware

The default stack is opinionated but replaceable:

```go
// Use defaults (ServerTiming → Recovery → RequestID → Logging → Timeout → SecurityHeaders):
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

### Server-Timing handler phases (sub-metrics beyond `total`)

The default stack's outermost middleware stamps `Server-Timing: total;dur=…` at
first write. Handlers can add their own named phases through the
[httputil/server_timing](https://github.com/LarsArtmann/httputil/tree/main/server_timing)
context helpers — nil-safe no-ops without the middleware, so handlers need no
branching. Measure BEFORE your first write: the header is committed at
response start, so a deferred measure that fires after the first write misses
the header (the phase is still recorded, it just cannot ship).

```go
import "github.com/larsartmann/httputil/server_timing"

func getUser(w http.ResponseWriter, r *http.Request) {
	defer servertiming.MeasureServerTiming(r.Context(), "db_fetch")() // "db_fetch;dur=…"
	fetchUser(r.Context())
	w.WriteHeader(http.StatusOK)
}
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

`Start` runs `StartHooks` synchronously after the listener binds and
before the first request is served (`Addr`/`Running` already observe the
listener there — the startup mirror of the drain-window contract). Any
StartHook error fails the start; the phase emits a `startup phase complete`
log line. Post-listen announcements (sd_notify `READY=1` via the `systemd`
module, dynamic port registration) are the canonical uses.

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

### Fleet-default error classification (opt-in, once per binary)

`appkit.RegisterClassifications()` seeds the process-wide registry with the
fleet defaults: go-error-family's documented stdlib classifications plus
`*http.MaxBytesError` → Rejection (a body over the limit is the client's
fault — 413, not a 500; the `security` module's BodyLimit relies on it).
Call it once at boot, before serving — idempotent, and deliberately not
auto-called (global-state hygiene: your binary owns the registry):

```go
func main() {
	appkit.RegisterClassifications() // once per binary; re-calls are no-ops
	// ... NewService, Run
}
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

## Related projects

The family blocks this framework composes, plus the siblings that pair with
it:

- [cmdguard](https://github.com/LarsArtmann/cmdguard) — the Cobra wrapper the
  fleet builds ops CLIs on: typed struct-tag flags, per-command config, and a
  machine-readable surface for agent-driven CLIs. Natural pairing: cmdguard
  for the control CLI, appkit for the daemon it drives.
- [go-cqrs-lite](https://github.com/LarsArtmann/go-cqrs-lite) — the CQRS/ES
  engine under the [cqrs](cqrs/README.md) module.
- [go-website-template](https://github.com/LarsArtmann/go-website-template) —
  the web-app sibling template; [go-service-template](https://github.com/LarsArtmann/go-service-template)
  is the service archetype of the same pattern.

## Development

Go >= 1.27 (the core module declares the `go 1.27` floor; satellite go.mods
carry dependency-forced minimums up to `go 1.27.4`; the commands below pin
`GOTOOLCHAIN=go1.27.4`, the patched toolchain this repo tests with):

```bash
go test ./...            # workspace: builds + tests the 14 member modules
GOTOOLCHAIN=go1.27.4 go test ./... -race -count=1
```

The `integration` module (cross-module E2E) is deliberately NOT a workspace
member — it pins PUBLISHED tags only, so its suite runs hermetic from its own
directory: `cd integration && GOWORK=off GOTOOLCHAIN=go1.27.4 go test ./...`.

Per-module verify (hermetic, resolves like a consumer via the proxy) and
per-module lint commands live in each module README under
**Build & verify**.

The whole loop as one command — hermetic build + vet + `go test -race`
plus sequential per-module lint across all fourteen module directories
(root + the 13 satellites, `scripts/verify-all.sh`), with per-module
log files and explicit PASS/FAIL EXIT verdict lines — is
`scripts/verify-all.sh` (toolchain derived from the root go.mod). On
shared machines, `--quiet[=LOADMAX]` re-execs the loop through
`scripts/quiet-run.sh`: it waits (up to 1h) for a quiet window before
starting, because deadline-bearing tests fail under load spikes.

## License

PROPRIETARY. All rights reserved — see [LICENSE](LICENSE) (licensing
inquiries: `git@lars.software`). Note: because the license text is not a
classifiable open-source license, pkg.go.dev hides godoc for this module; the
code and this README are the documentation of record.
