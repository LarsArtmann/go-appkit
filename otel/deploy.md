# Deploying go-appkit services with telemetry

Deployment-config-only telemetry: the `otel` module's zero-code setup means
the environment alone turns export on, scopes it, and tunes it — no code
changes, no config files. This page collects the exact variable sets for
the common deployment targets. Every variable is honored by `appkitotel.Setup`
(verified in `setup_test.go` / `otlp_test.go`); exporter-level variables
(compression, retries) are honored natively by the OTLP/HTTP exporters.

## Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders-api
spec:
  template:
    spec:
      containers:
        - name: orders-api
          image: registry.example.com/orders-api:1.4.2
          env:
            - name: OTEL_EXPORTER_OTLP_ENDPOINT
              value: http://signoz-otel-collector.observability:4318
            - name: OTEL_SERVICE_NAME
              value: orders-api
            # deployment.environment feeds SigNoz's environment filter
            # (the same attribute appkitotel.WithEnvironment sets in code).
            - name: OTEL_RESOURCE_ATTRIBUTES
              value: deployment.environment=production
            # Head sampling at 10% (optional — default samples everything):
            - name: OTEL_TRACES_SAMPLER
              value: parentbased_traceidratio
            - name: OTEL_TRACES_SAMPLER_ARG
              value: "0.1"
            # SigNoz Cloud ingestion (self-hosted omits this):
            # - name: OTEL_EXPORTER_OTLP_HEADERS
            #   value: signoz-ingestion-key=<key>
```

Kill one signal without touching the image (e.g. during a collector
incident):

```yaml
- name: OTEL_TRACES_EXPORTER
  value: none # traces off; metrics keep flowing
```

Tuning knobs (all optional):

| Variable                         | Effect                                               |
| -------------------------------- | ---------------------------------------------------- |
| `OTEL_METRIC_EXPORT_INTERVAL`    | Metric reader cadence in ms (SDK default 60000)      |
| `OTEL_METRICS_EXEMPLAR_FILTER`   | `always_on` / `always_off` / `trace_based` (default) |
| `OTEL_GO_X_CARDINALITY_LIMIT`    | Per-instrument series cap (default 2000)             |
| `OTEL_EXPORTER_OTLP_TIMEOUT`     | Per-export request deadline (default 10s)            |
| `OTEL_EXPORTER_OTLP_COMPRESSION` | `gzip` (value must be exactly `gzip`)                |

## docker-compose

```yaml
services:
  orders-api:
    image: registry.example.com/orders-api:1.4.2
    environment:
      OTEL_EXPORTER_OTLP_ENDPOINT: http://signoz-otel-collector:4318
      OTEL_SERVICE_NAME: orders-api
      OTEL_RESOURCE_ATTRIBUTES: deployment.environment=staging
    ports:
      - "8080:8080"

  signoz-otel-collector: # standalone collector for local testing
    image: signoz/signoz-otel-collector:latest
    ports:
      - "4318:4318" # OTLP/HTTP — the protocol appkit speaks
```

## What NOT to configure

- **No SDK-side metric scraping.** The module exports metrics via OTLP only
  (no Prometheus endpoint) — the collector is the scraper/prometheus-bridge.
- **No `OTEL_*_EXPORTER` values other than `none` change anything.**
  `console`, `prometheus`, `zipkin`, ... are documented no-ops: the posture
  is OTLP/HTTP-only.
- **No `applicationinsights.logging` wiring.** There is no OTLP logs signal
  yet (design sketch: `docs/planning/2026-10-18_otel-otlp-logs-design.md`);
  logs go to stdout and the collector's file/stdout pipelines as usual.
- **The exporter options in code stay authoritative.** Explicit
  `appkitotel.WithOTLP*` code options win over every variable here —
  environment config is the no-redeploy path, not an override path.
