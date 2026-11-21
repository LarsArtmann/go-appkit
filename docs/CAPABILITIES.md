# Capability map — what appkit replaces

Per-capability inventory: each module, the hand-rolled code it replaces, its
dependency surface, and the version it arrived in. Pair with the machine
inventory (`scripts/module-inventory.sh` → `docs/MODULE_INVENTORY.md`) for the
current newest tag per module — this page holds capability truth, the
inventory holds version truth.

**As of:** 2026-11-21 (SUPERB v17 T21; newest tags at that date: core v0.12.0,
cqrs v0.9.0, httpx v0.2.0, otel v0.4.0, realtime v0.2.0, health v0.3.2,
security v0.2.0, systemd v0.2.0, config v0.1.1, preset v0.1.1).

| Module | Replaces your hand-rolled… | Dep surface (beyond stdlib) | Since |
| ------ | -------------------------- | --------------------------- | ----- |
| core (`appkit`) | service main(): bind, signals, graceful drain, health handlers, request logging, recovery, request IDs, security headers, exit-code mapping | httputil, server_timing, charmbracelet/log, go-error-family | v0.1.0 (lifecycle); metrics/version/testkit v0.5.0; StartHooks v0.8.0; Server-Timing default stack + `RegisterClassifications` v0.10.0; listener seam + log-handler seam in [Unreleased] |
| [preset](preset/README.md) | the ~50-line `main.go` composition boilerplate (spine wiring, env switches, health mount, error pages) | core, systemd, health, errorpages, otel, realtime, security | v0.1.0 |
| [config](config/README.md) | the koanf/viper config package (env parsing, precedence, file layers, typed errors) | go-error-family, go-faster/yaml | v0.1.0 |
| [cqrs](cqrs/README.md) | the event-store/projection/DLQ wiring over go-cqrs-lite (lifecycle, readiness, staleness guards, typed C/Q facade) | go-cqrs-lite `system` + engines | v0.1.0; Domain passthrough v0.7.0; folded contract (NotifyFolded) v0.9.0 |
| [realtime](realtime/README.md) | the SSE layer: hub, canonical endpoint, Last-Event-ID replay, heartbeats | go-sse, go-error-family | v0.1.0; FoldedSource v0.2.0 |
| [httpx](httpx/README.md) | the per-handler bind/validate/error-render boilerplate; no-leak error rendering; conditional GET; content negotiation; write deadlines | go-error-family, go-etag/server, go-codec | v0.1.0 (B1/B2/B9); v0.2.0 (B3/B4/B5/B7) |
| [security](security/README.md) | the hand-rolled middleware battery: API-key auth, CSRF, rate limits, origin check, body limits, sanitization, CSP nonces | bluemonday, golang.org/x/time | v0.1.0 (all batteries); example + threat model v0.2.0 |
| [errorpages](errorpages/README.md) | the pretty-404/405/500 HTML error pages + their JSON twin | templ-components/errorpage, go-error-family | v0.1.0 |
| [health](health/README.md) | the healthz plumbing + status dashboard (HTML/JSON/SSE) | go-health, go-health-dashboard | v0.1.0; per-check durations (NewChecks) v0.2.0 |
| [otel](otel/README.md) | the telemetry provider setup, otelhttp bridge, trace-correlated logs, client spans | otel, otelhttp, otlp exporters | v0.1.0; self-diagnostics + `Transport()` + detectors v0.3.0; semconv v1.43.0 rename v0.4.0 |
| [flightrecorder](flightrecorder/README.md) | runtime trace capture middleware (trigger predicates, snapshot endpoint) | go-flightrecorder | v0.1.0; async initiation v0.2.0 |
| [flightrecorderhealth](flightrecorderhealth/README.md) | the health-failure → trace-capture glue | go-flightrecorder, go-health | v0.1.0 |
| [systemd](systemd/README.md) | the sd_notify hand-roll (READY=1, STOPPING=1, watchdog keepalive) | go-systemd/v22 | v0.1.0; `Install` returns the wired Hooks v0.2.0 |
| [docs](docs/README.md) | hand-written API docs that drift from code (OpenAPI/AsyncAPI/D2 from Go types) | catalog/v4 | v0.1.0 (v0.3.0 `docs/` repath) |
| testkit (in core) | the per-repo test-boot helper (and its raw-mux footgun) | core only | v0.5.0; test-safe `NewService` v0.9.0; `FreeAddr`/`WaitForRunning`/GET helpers v0.10.0 |

Deliberately NOT in the bay (recorded non-goals): DI containers, cron
schedulers, WebSocket transport (SSE only), storage engines (cqrs engines are
go-cqrs-lite's), application-shaped code of any kind. The demand ledger for
new batteries: `docs/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md`.
