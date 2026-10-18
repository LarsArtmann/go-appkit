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

Your own metrics ride the same pipe: any instrument created from
`provider.AsMeterProvider().Meter("myapp")` (counters, histograms) exports
through the identical OTLP endpoint — the example's flight-recorder bridge
(`appkit_flightrecorder_snapshots_total{type=...}`) is exactly that shape,
and shows up in SigNoz next to the HTTP metrics with no extra wiring.

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
the span status _description_ at span end — the exception event (type,
message, stacktrace) is what survives, and that is exactly what the
exceptions view indexes.

## What you get

| Signal    | Instrument                                                | Notes                                                      |
| --------- | --------------------------------------------------------- | ---------------------------------------------------------- |
| Traces    | one SERVER span per request                               | named after the ServeMux pattern (`GET /users/{id}`)       |
| Traces    | W3C `traceparent`/`baggage` in and out                    | continues caller traces; feeds downstream calls            |
| Traces    | `exception` events (panics + handled errors)              | `Recovery` + `RecordError`; feeds SigNoz's Exceptions view |
| Metrics   | `http.server.request.duration` (+ size, active)           | method/route/status attributes; route-based, no blowups    |
| Metrics   | `appkit_flightrecorder_snapshots_total{source,kind,type}` | flight-recorder bridge; per-operation capture counts       |
| Metrics   | `appkit_flightrecorder_snapshot_duration_seconds`         | flight-recorder bridge; snapshot write latency             |
| Logs      | `trace_id` + `span_id` on records logged with ctx         | `TraceHandler` decorates any `slog.Handler`                |
| Export    | OTLP/HTTP for traces + metrics, env-driven or in code     | `WithOTLP`; `OTEL_EXPORTER_OTLP_*` natively honored        |
| Lifecycle | provider `Shutdown` in `ServiceConfig.ShutdownHooks`      | flush after drain — spans cover the final requests         |

## Outbound calls

Wrap your client's transport and every downstream request continues the
inbound trace — one CLIENT span (`HTTP <method>`) per call, W3C
`traceparent`/`baggage` injected, client-side semconv metrics recorded:

```go
client := &http.Client{Transport: appkitotel.Transport()}
```

Providers default to the process globals from `Setup`; without one the
transport is a pass-through (same strictly-opt-in contract as the server
middleware). `Transport(WithBaseTransport(rt))` takes a tuned base
RoundTripper — connection timeouts and proxies belong there, request
timeouts stay on the `http.Client`.

For low-level sub-spans (DNS, dial, TLS handshakes as span events) opt into
`go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace`
on top of this transport — and on high-fanout call paths prefer
`otelhttptrace.WithoutSubSpans()`: one client span per call is usually the
right granularity, sub-spans multiply span volume per request.

## Flight-recorder metric bridge

`NewFlightRecorderMetricsHook` turns every go-flightrecorder capture (manual,
trigger, or async) into the two `appkit_flightrecorder_*` metrics above — no
counter plumbing, no vendor lock-in. Wire it when the RECORDER is built (the
hook fires on capture completion, which the HTTP middleware never sees):

```go
rec, err := fr.New(
    fr.WithSnapshotDir("/var/lib/myapp/traces"),
    fr.WithMaxSnapshots(5),
    fr.WithMetrics(appkitotel.NewFlightRecorderMetricsHook(provider.AsMeterProvider().Meter("myapp"))),
)
```

Chart `snapshots_total` by `type` in SigNoz to see which operations produced
traces; a silent counter plus a full disk is exactly the retention-failure
case the flightrecorder module's `OpsRecorderLoggerPreset` surfaces in logs.
See `example/main.go` for the full loop (the `/slow` route fires captures).

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
- `Setup(WithDiagnosticsLogger(logger))` — route the SDK's self-diagnostics
  through YOUR logger: exporter failures surface as grep-able
  `otel export failed` error lines (see `doc/TELEMETRY.md` §1) and
  SDK-internal logs become regular slog records, instead of both dying on
  stderr. Without it, the same contract routes through `slog.Default()`;
  `WithoutGlobalRegistration()` skips the wiring with the rest of the
  process globals.

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

**Off-switch and protocol divergence (deliberate):** `Setup` honors the
spec's `OTEL_TRACES_EXPORTER=none` / `OTEL_METRICS_EXPORTER=none` as a veto
on the ENVIRONMENT-driven signal (case/whitespace tolerant) — the
deployment can kill one signal without a code change, e.g. during an
incident. Explicit code options win over the veto, as over every other
environment variable. Other exporter names (`console`, `prometheus`,
`zipkin`, ...) are documented NO-OPS: this module exports OTLP/HTTP only —
the spec's exporter zoo is out of scope, so the variables change nothing.

Wire-level knobs worth knowing (source-verified against
`otlp{trace,metric}http` v1.47.0):

- **Compression:** `OTEL_EXPORTER_OTLP_COMPRESSION=gzip` (zero-code) or the
  per-signal `OTEL_EXPORTER_OTLP_TRACES_COMPRESSION` /
  `OTEL_EXPORTER_OTLP_METRICS_COMPRESSION`; the value must be exactly
  `gzip` (anything else parses as no-compression).
- **Retries:** built in — the OTLP/HTTP exporters back off and retry
  throttled responses (429 with `Retry-After`) and transient 5xx/2xx-error
  bodies themselves; nothing to configure on our side.
- **Export cadence:** traces flush on batch (span count / 5s default);
  metrics ride `OTEL_METRIC_EXPORT_INTERVAL` (milliseconds; SDK default
  60s — pinned by `TestWithOTLP_MetricExportIntervalEnv`).

### Tail sampling

Head sampling (`WithSampler`, `OTEL_TRACES_SAMPLER`) drops spans at capture
time and is the cheap default. True tail sampling (decide AFTER the trace
is complete, e.g. keep 100% of errors, 1% of the rest) needs a collector:

```go
// Send everything to a local collector; let IT sample.
provider, _ := appkitotel.Setup(
    appkitotel.WithService("myapp", version, instance),
    appkitotel.WithOTLP(appkitotel.WithOTLPEndpoint("http://localhost:4318")),
)
```

Run `otelcol` with a `tail_sampling` processor (policies: `status_code`
for errors, `probabilistic` for the rest) and forward from there. The
application cannot do tail sampling itself — a Go SDK drops the span at
End() unless a tail-sampling collector receives the WHOLE trace first
(the reason `WithSampler` × `OTEL_TRACES_SAMPLER` is head-only by design).

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
- **Semantic conventions pin**: semconv **v1.26.0**, imported explicitly
  (`setup.go`) — one version behind upstream (v1.43.0 shipped) ON PURPOSE:
  SigNoz's default filter key is still the old `deployment.environment` this
  module emits; the new `deployment.environment.name` key would break the
  dashboard filter until SigNoz flips its default. Re-pin + rename together
  when it does (watch tracked in TODO_LIST).
- **Exemplars + cardinality** (SDK defaults, called out here so they are
  discoverable): the metric provider ships `exemplar.TraceBasedFilter` —
  spans in the measurement context become exemplars on
  `http.server.request.duration` — plus a 2000-series cardinality limit per
  instrument. Both are env-tunable without code:
  `OTEL_METRICS_EXEMPLAR_FILTER` (always_on/always_off/trace_based) and
  `OTEL_GO_X_CARDINALITY_LIMIT`. The flight-recorder bridge (`frmetrics.go`)
  and the cqrs projection metrics report against `context.Background()`
  deliberately — no span context, so no exemplar coupling of async captures
  and projection cycles to request traces.
- **`otelhttp.WithMessageEvents` declined** (evaluated 2026-08-18): byte-count
  message events add span-event noise with no consumer; NOT-DO stands.
- **Baggage correlation helpers**: baggage propagates in and out (W3C
  carrier); `CorrelationID(ctx)` / `WithCorrelationID(ctx, id)` read and set
  the `correlation_id` member (read-or-empty, set-never-invent — invalid or
  empty values leave the context untouched).
- **Resource identity**: the merged resource carries the env/SDK/host
  detectors plus the NON-SENSITIVE process subset — `process.pid`,
  `process.runtime.name`/`version`, `process.executable.name`.
  `process.command_args` is deliberately EXCLUDED (secrets get passed
  through flags and would leak into telemetry); add targeted
  `resource.WithProcess*` detectors yourself if you need more.
- **Non-goals (deliberate)**: no OTLP **logs** signal — the Logs API+SDK went
  STABLE in otel v1.47.0 (2026-10-02, the exact version this module pins), so
  the old "still experimental" blocker is gone; what remains is the consumer
  gate — build when a consumer asks for trace-correlated log ingestion. The
  design sketch (`WithOTLPLogs()` shape, dependency cost, shutdown order)
  lives in `docs/planning/2026-10-18_otel-otlp-logs-design.md`. And a handler
  that writes 5xx WITHOUT panicking produces an error span but NO exception
  event (SigNoz's Exceptions view only indexes `Recovery`/`RecordError` output)
  — an opt-in auto-exception-on-5xx is plausible but unbuilt.

## Go build notes

Builds with plain `go build` (the repo's `GOEXPERIMENT=jsonv2` build-command
prefixes were retired 2026-09-29; `encoding/json/v2` is default-on).

## Performance

`BenchmarkMiddleware_*` (see `benchmark_test.go`; go 1.27.2, linux/amd64,
benchstat n=10, re-baselined 2026-10-18 in a quiet-run-gated window —
load1 11.7, 8 primary threads pinned via `scripts/quiet-run.sh`;
`httptest` round-trip against `w.WriteHeader(200)`; run-to-run drift on
this box is ±25%):

| Variant                 | ns/op (mean ± sd) | B/op    | allocs/op |
| ----------------------- | ----------------- | ------- | --------- |
| NoOp (no providers)     | 19.67µs ± 5%      | 7.66Ki  | 90        |
| Traced (spans)          | 20.45µs ± 13%     | 10.67Ki | 90        |
| Traced + metered (full) | 25.33µs ± 16%     | 10.67Ki | 90        |

2026-09-29 re-run (n=5, go 1.27.1 floor, httputil v1.4.0): B/op and
allocs/op are UNCHANGED against the table (NoOp 7,973 B/90 allocs; traced
11,092 B/90; traced+metered 11,095 B/90) — no allocation regression from
the toolchain or dependency bump. The ns/op re-baseline is deferred: the
box was under an external nix-build storm (load average 150), which
poisons wall-clock numbers ~2.3x; per the ±25% drift rule below, do not
re-baseline ns/op under load.

2026-10-13 re-run (n=10, go 1.27.2 floor, benchstat, SUPERB v7 T10):
B/op and allocs/op UNCHANGED again (NoOp 7,762 B/90 ±0%; traced 10.84Ki
B/90; traced+metered 10.85Ki/90; Recovery happy 4,631 B; panic 11.66Ki;
RecordError-recording 3,969 B — all ±0%, load-independent). The ns/op
re-baseline is deferred a second time: ambient load rose 9→26 mid-run
(external test binary at 430% CPU); the middleware rows came out ±40-45%,
far outside the ±25% honesty bar. ns/op rows stay at the 2026-09-16
baseline until a genuinely quiet window appears.

2026-10-18 RE-BASELINE (n=10, go 1.27.2, benchstat, quiet-gate protocol —
`scripts/quiet-run.sh`: load1 11.7 at start, taskset-pinned to 8 PRIMARY
threads): the middleware table above is the fresh baseline — all three
rows within ±9% of 2026-09-16 (NoOp 19.67µs ± 5%; traced 20.45µs ± 13%;
traced+metered 25.33µs ± 16%), so the 1.26.7→1.27.2 toolchain jump costs
nothing measurable. B/op and allocs/op UNCHANGED again (NoOp 7.66Ki/90;
traced 10.67Ki/90; traced+metered 10.67Ki/90 — ±0%). PROTOCOL FINDING:
the first same-day attempt pinned the 8 "least-busy" cores, which
included two SMT sibling pairs (12/28, 9/25) — sibling contention made
the same rows bimodal (±26-36% sd); the picker is SMT-aware now (primary
threads only) and the middleware noise collapsed to ±5-16%. The Recovery
happy/panic rows below remain cross-run UNSTABLE (16.35µs↔27.29µs happy,
33.45µs↔40.65µs panic across two same-protocol runs — suspected
frequency-scaling sensitivity of the deferred-recover path) and keep
their 2026-10-12 values until a stable panic-path protocol exists;
RecordError rows are confirmed (recording 4.71µs vs 5.96µs, -21%, within
the bar; no-span 5.6ns).

Full instrumentation costs ~5.7µs/req (~+29% over the no-op path at the
2026-10-18 baseline — the no-op row got faster on go 1.27.2 while the
metered row held, widening the relative gap from the +17% seen at
2026-09-16) and ~3.0KiB over the no-op path, with zero additional
allocations; export I/O is excluded by design (batching processor, no
exporter wired). The no-op baseline is the cost of the middleware
existing in the chain with the module imported but no providers — the
strictly-opt-in guarantee in numbers.

`BenchmarkRecovery_*` + `BenchmarkRecordError_*` (added 2026-10-12, same
harness, go 1.27.2):

| Variant                          | ns/op  | Notes                                        |
| -------------------------------- | ------ | -------------------------------------------- |
| Recovery happy path              | 17,056 | deferred-recover tax every request           |
| Recovery panic path              | 50,080 | event + stacktrace + status + 500 + log line |
| RecordError, recording span      | 5,962  | exception event with stacktrace + status     |
| RecordError, no span (fast path) | 6      | guard clause only                            |

A captured panic costs ~33µs over the happy path; a handled error costs
~6µs per `RecordError` call. The panic-path logger writes to `io.Discard`
in the benchmark, so terminal I/O is excluded there too.

## Deploying with telemetry

The zero-code setup is deployment-config-only — k8s Deployment and
Compose snippets, the off-switch, and the tuning-knob table live in
[deploy.md](deploy.md).

## Build & verify

```bash
cd otel
GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.2 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.2 go build ./...
GOWORK=off GOTOOLCHAIN=go1.27.2 go test -bench . -benchmem -run '^$' ./...
golangci-lint run ./...        # from this directory
```

`GOTOOLCHAIN=go1.27.2` is required for every command below on
machines whose default toolchain is older (this repo pins
`go 1.27.2` in every go.mod). `GOWORK=off` makes the run
hermetic: the module resolves exactly what its own go.mod pins,
via the module proxy — how consumers resolve it.

Benchmark drift on this suite is ±25% run-to-run (measured 2026-09-15/16
on identical code): treat single-run deltas under that as noise.
