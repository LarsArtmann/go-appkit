# SigNoz Dashboards for go-appkit services

Import-ready [SigNoz](https://signoz.io) dashboards for services instrumented
with `go-appkit/otel`. All panels run on the **traces** signal — the server
spans `appkitotel.Middleware` emits — so the dashboards light up the moment
spans reach SigNoz, no metric configuration required.

## Import

SigNoz UI: **Dashboards → New Dashboard → Import JSON** (or the folder "+" →
Import), then pick the file.

| File                         | What it shows                                                                                                 |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `appkit-http-dashboard.json` | Request rate, 5xx error rate, p50/p90/p99 latency and status-code distribution per route, top-endpoints table |

The `service.name` variable filters everything (multi-select + "all"); the
filters expect HTTP **server** spans (`spanKind = 'Server'`), which is what
`appkitotel.Middleware` produces.

## Wiring that feeds these dashboards

The zero-code path — configure only the environment, no code changes:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"   # SigNoz OTLP/HTTP
export OTEL_SERVICE_NAME="orders-api"
```

```go
provider, err := appkitotel.Setup() // env turns export on
```

Or fully in code:

```go
provider, err := appkitotel.Setup(
    appkitotel.WithService("orders-api", "1.0.0", pod),
    appkitotel.WithOTLP(appkitotel.WithOTLPEndpoint("http://localhost:4318")),
)
```

## Exceptions (the SigNoz exceptions view)

SigNoz's **Exceptions** view (left navigation) reads `exception` span events.
Two pieces of this module produce them — add both for full coverage:

1. `appkitotel.Recovery(logger)` in `ExtraMiddlewares` turns panics into
   `exception.type` / `exception.message` / `exception.stacktrace` events.
2. `appkitotel.RecordError(r.Context(), err)` inside handlers records handled
   errors the same way.

Then SigNoz shows each exception with its stack trace, service, and a link
into the trace waterfall.

## SigNoz quick start (local)

```bash
git clone -b main https://github.com/SigNoz/signoz.git
cd signoz/deploy
./install.sh   # or: docker compose up -d (newer versions)
```

SigNoz listens on OTLP `:4318` (HTTP) and `:4317` (gRPC); the UI is on
`:3301`. Point `OTEL_EXPORTER_OTLP_ENDPOINT` at `http://localhost:4318`, run
your app, import the dashboard, and hit some endpoints.
