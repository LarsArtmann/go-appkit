# OTLP Logs Signal — Design Sketch (`WithOTLPLogs()`)

**2026-10-18 · Status: SKETCH ONLY — build is GATED on the consumer-demand half of the
OTLP-logs trigger (status-report Q3). No `sdk/log` dependency enters `otel/go.mod` until
that gate opens.** Referenced by the otel README non-goals bullet.

## Why a sketch now

The old non-goal ("`go.opentelemetry.io/otel/log` is still experimental") went stale on
2026-10-02: otel v1.47.0 — the exact version this module pins — shipped the first STABLE
Logs API+SDK. The technical blocker is gone; what remains is the owner's wait-for-consumer
gate. This sketch makes the future build a one-session change instead of a research
project, without spending the dependency budget early.

## Option shape (mirrors `WithOTLP`)

```go
// WithOTLPLogs wires OTLP/HTTP log export, mirroring WithOTLP's option set.
func WithOTLPLogs(opts ...OTLPOption) SetupOption   // reuses WithOTLPEndpoint/Headers/Timeout
```

- `setupConfig` gains `otlpLogs *otlpConfig` + `logsEnabled bool`; `WithOTLPLogs` sets
  both. `applyOTLPWiring` grows a third leg: env-auto-enable via
  `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` (fallback `OTEL_EXPORTER_OTLP_ENDPOINT`), vetoed by
  `OTEL_LOGS_EXPORTER=none` (same `envExporterOff` helper as traces/metrics), explicit
  code wins per signal.
- Exporter: `otlploghttp.New(ctx, opts...)` (same base-URL + `/v1/logs` path convention
  via `otlpSignalURL`) wrapped in `sdklog.NewBatchProcessor` — the traces-side batching
  semantics, not the metric-side periodic model (logs are push-based records).
- Provider: `sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(...))`.
- Bridge: `otelslog.NewLogger(name, otelslog.WithLoggerProvider(lp))` converts slog
  records into log records carrying trace/span IDs natively. Interaction with
  `TraceHandler` (the current `trace_id`/`span_id` decoration) must be decided: with the
  logs SDK active, `TraceHandler` becomes redundant for OTLP-bound output but still
  formats local stdout — proposal: `WithOTLPLogs` does NOT auto-replace the consumer's
  logger; document the recipe instead.

## Dependency cost (the reason the gate exists)

| Package                                                       | Version at sketch time                                     | Stability        |
| ------------------------------------------------------------- | ---------------------------------------------------------- | ---------------- |
| `go.opentelemetry.io/otel/log`                                | v1.47.0                                                    | STABLE (API)     |
| `go.opentelemetry.io/otel/sdk/log`                            | v1.47.0-adjacent (module `otel/sdk/log`, own version line) | STABLE (SDK)     |
| `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp` | v0.23.0                                                    | PRE-1.0 exporter |
| `go.opentelemetry.io/contrib/bridges/otelslog`                | v0.21.0                                                    | PRE-1.0 bridge   |

Net: three new require lines, all otel-family, two pre-1.0. Fleet policy takes new deps
only on consumer demand — hence the gate.

## Shutdown order

`Provider.Shutdown` runs four phases today (tracer flush, tracer shutdown, meter flush,
meter shutdown). The logs leg inserts as: **tracer flush → logs flush
(`loggerProvider.ForceFlush`) → logs shutdown → tracer shutdown → meter flush → meter
shutdown** — logs are flushed EARLY because the batch processor holds the same mid-flight
window as spans, and a lost log record that references a kept span breaks the correlation
the whole module exists for. The final `graceful shutdown complete` line's phase list
grows `logs_flush`/`logs_shutdown` (grep-able contract, pinned in the shutdown-order
test).

## Gate status (Q3)

- 2026-10-18: **CLOSED — NOT MET (empirically, T16 audit).** Both known
  consumers grepped at their published tags: cqrs-htmx `setup` v4.15.0
  carries ZERO `go-appkit/otel` imports; PapDashboard imports go-appkit core
  only, no otel module. Nobody consumes `appkitotel`, so nobody demands
  OTLP log ingestion. The gate re-opens only when a consumer adopts the
  otel module AND asks for log export (T16 re-runs on consumer-claims
  refreshes).
- If the answer stays "no": this sketch is the final artifact; the README
  non-goal keeps pointing here instead of a build date.
