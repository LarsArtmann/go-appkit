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
| `appkit-5xx-alert.example.json` | SigNoz alert (v2alpha1 schema): fires when the 5xx share of server spans per service+route exceeds 5% — same trace grammar as the dashboard's 5xx panel |

The `service.name` variable filters everything (multi-select + "all"); the
filters expect HTTP **server** spans (`spanKind = 'Server'`), which is what
`appkitotel.Middleware` produces.

Alert import (the example file): there is no UI "Import JSON" for alerts
([SigNoz#9266](https://github.com/SigNoz/signoz/issues/9266) open) — create
it via the API: `curl -X POST $SIGNOZ/api/v2/rules -d @appkit-5xx-alert.example.json`
(set your channel ID in `condition.thresholds.spec[0].channels` first; empty
means notification settings only). Grammar is field-name-verified against
SigNoz's OpenAPI spec (v2alpha1 `PostableRule`: `alert` (not `name`),
`condition.compositeQuery.queries[]` builder_query/builder_formula envelopes,
`evaluation`, `notificationSettings`) — but, like the dashboard, it has NOT
been applied to a live SigNoz; treat it as spec-proven, not
instance-proven.

Version note: the dashboard uses schemaVersion **v6** (template-derived from
SigNoz's own `apm/http-api-monitoring.json` of the ~v0.135 era) — older
SigNoz instances that predate v6 may fail the import. The JSON is
parse-verified but has NOT been imported into a live SigNoz (no instance in
this environment); treat panel behavior as template-proven, not
instance-proven.

**Exceptions panel — builder blocker (recorded 2026-10-18, source-verified
at SigNoz main):** the QueryBuilder v5 signal enum carries exactly
`traces`/`logs`/`metrics` (`pkg/types/telemetrytypes/signal.go`) — there is
NO `exceptions` signal and no version ever added one, so a builder panel
cannot chart the exception events this module feeds (SigNoz's Exceptions
view reads `signoz_error_index_v2` through its own dedicated API, not the
builder). Workaround when wanted: a ClickHouse-SQL panel over
`signoz_traces.distributed_signoz_error_index_v2` (the shape SigNoz's own
exceptions-based ALERT defaults use) — not shipped here because the
dashboard-panel JSON shape for ClickHouse queries is unverified in this
environment; revisit alongside the live-instance validation (accepted-risk
row in TODO_LIST).

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
