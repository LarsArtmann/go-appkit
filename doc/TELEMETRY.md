# go-appkit Telemetry Reference

**Status:** living reference (2026-09-16, telemetry bundle C25/C26). Every
line/metric name below is shipped and grep-able — this document is the
index, not a plan.

---

## 1. Log-line catalogue

| Line / message                        | Level | Source                          | Fields                                                             | Notes |
| ------------------------------------- | ----- | ------------------------------- | ------------------------------------------------------------------ | ----- |
| `shutdown phase complete`             | INFO  | core `Service.Shutdown`         | `phase`, `duration`                                                | Grep-able contract, pinned by `shutdownlog_test.go`. Phases: `ready_flip`, `drain_hooks`, `drain_wait`, `listener_close`, `shutdown_hooks` |
| `shutdown phase skipped`              | INFO  | core, with `NoDrainDelay`       | `phase=drain_wait`                                                 | Level decision 2026-09-16: stays INFO (rationale in root README "Log volume") |
| `graceful shutdown complete`          | INFO  | core, end of shutdown           | `total`, `result` (`ok`/`error`)                                   | The deploy-diagnosis line |
| `draining traffic`                    | INFO  | core, before the drain wait     | `delay`                                                            | |
| `request method=… duration=…`         | INFO  | httputil `Logging`              | `method`, `path`, `status`, `duration`, `client_ip`, `request_id`  | The only per-request line. NOT context-correlated upstream (see §4 recipe); suppressed for free at `LogLevel: Warn` |
| `context cancelled, shutting down`    | INFO  | core `Service.Run`              | —                                                                  | |
| `realtime: replay store read failed…` | ERROR | realtime                        | `last_event_id`, `err`                                             | Followed by the SSE `event: error` (see §3) |
| `security.OriginCheck: origin rejected` | WARN | security module                | `origin`, `path`                                                   | |
| `httputil: CSRF validation failed`    | WARN  | httputil (via security module)  | `method`, `path`, `reason`                                         | |
| `security.CSRF: "*" in TrustedOrigins…` | WARN | security module, construction  | —                                                                  | Misconfiguration warning, fires once |

Default level is INFO (decided 2026-09-16, README "Log volume": emitting
costs ~+30µs/req, suppression at WARN ~+0.8µs — flip `LogLevel` for
high-throughput services).

## 2. Metric catalogue (names are contracts)

Core `ServiceConfig.Metrics` surface (dependency-free Prometheus text):

| Metric                                    | Type      | Labels                |
| ----------------------------------------- | --------- | --------------------- |
| `appkit_http_request_duration_seconds`    | histogram | `method`, `route`, `status` |
| `appkit_http_responses_total`             | counter   | `method`, `route`, `status` |
| `appkit_http_requests_in_flight`          | gauge     | —                     |
| `appkit_build_info`                       | gauge=1   | `version`             |

Route labels are ServeMux patterns (`GET /users/{id}`), never raw paths;
unmatched requests collapse to `route="unmatched"`.

OTel module surface (when `Setup` + the middleware are wired):

- SERVER spans named by matched pattern (`GET /users/{id}`); W3C propagation;
  health paths unconditionally filtered; `http.server.request.duration` with
  `http.route` (no method prefix — assert accordingly).
- `appkit_flightrecorder_snapshots_total{source,kind}` +
  `appkit_flightrecorder_snapshot_duration_seconds` (fr bridge).

cqrs surface (via `NewOTelProjectionMetrics`):
`cqrs.projection.event.count{projection,event_type,status}`,
`cqrs.projection.event.duration`, `cqrs.projection.worker.count`,
`cqrs.projection.checkpoint.lag`.

**The `_ratio` trap:** OTEL's Prometheus exporter appends `_ratio` to unit-1
metrics — names on the OTel path do NOT match the core surface's names.
Diff on migration; never assume 1:1.

## 3. Backpressure semantics (lossy vs blocking, per sink)

| Sink                          | Behavior under pressure                                                                 |
| ----------------------------- | --------------------------------------------------------------------------------------- |
| OTel batch processor          | LOSSY: queue-full spans/metrics are dropped silently; `Provider.Shutdown` ForceFlushes first, so only the mid-flight window is exposed |
| charmbracelet log formatting  | BLOCKING per line (~+30µs): a slow stdout slows the request goroutine — the real reason to raise `LogLevel` |
| SSE subscriber buffer (64)    | LOSSY: bursts above the buffer drop live events; clients heal via Last-Event-ID replay (exactly-once, dedup-pinned) |
| SSE store-failure             | ABORT with signal: named `event: error` + `retry: 30000` before the drop (no reconnect storm) |
| Health probe cache            | LATEST-WINS: the dashboard reads `CachedResponse` per push tick — a failed batch surfaces, history compresses |
| Rate limiter                  | REJECT: 429 + `Retry-After`, chain aborts (never overwritten — regression-pinned)        |

## 4. Incident-debug recipe

1. **Start from the shutdown lines.** `graceful shutdown complete` with
   `result=error` → grep the phase lines above it: `drain_hooks` /
   `shutdown_hooks` failures are joined into the result; `listener_close`
   failures are server-level.
2. **Correlate a slow request.** The `request` line's `request_id` matches
   the handler-scoped logs. For traces: the otel middleware names spans by
   route — pull `GET /route` spans, not method-only spans (those mean the
   middleware order regressed; the integration-module pin test fails first).
3. **Second exporter without touching baseline.** Run a SECOND provider with
   `WithoutGlobalRegistration` + `WithStdoutExporter` next to the production
   one and feed the sidecar provider explicitly where you need verbosity —
   the global owner stays untouched (one-`Setup`-per-process rule still
   holds: only the GLOBAL owner registers).
4. **Projection lag.** `CheckStaleness`/`LagPerProjection` for the read-time
   truth; `cqrs.projection.*` metrics for the trend; DLQ for the poison
   events (keep the DLQ store SEPARATE from the event store — §5).
5. **Goroutine leak.** `testkit.Serve` asserts the baseline in tests; in
   prod, `runtime.NumGoroutine()` deltas after a deploy pair with the
   `appkit_http_requests_in_flight` gauge (leak = gauge drains, goroutines
   don't).

## 5. Store-separation doctrine (cqrs)

DLQ, telemetry, and checkpoint stores must NOT share the event store's
database. Reasons: (a) a poison event's DLQ retries become event-store
write-load exactly when the event store is already struggling; (b) telemetry
writes turn a metrics outage into an event-store outage; (c) backup/restore
granularity diverges (event store = durability, DLQ = triage queue). The
wrapper's `EventConfig.DSN`/`Driver`/`Pragmas` are one store — open separate
ones for DLQ/telemetry sinks.

## 6. Candidate (not scheduled): SSE filtered-telemetry battery

Per-subscriber delivery counters (`delivered`, `filtered`, `dropped`) as an
opt-in realtime option — candidate for W5 (battery spec C1 covers drop
counters; do not double-track). Recorded here so the idea survives TODO_LIST
rebuilds.
