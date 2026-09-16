# Changelog

## [Unreleased]

### Added

- README.md: module overview, quick start, options table, shutdown ordering,
  journal-backed replay via cqrs-htmx `transport.JournalSSEStore`, and known
  limitations (burst drops healed by Last-Event-ID reconnect, nginx
  `X-Accel-Buffering` caveat).

### Fixed

- Reverse-proxy buffering: the SSE handler now sets `X-Accel-Buffering: no`
  (before the first write) so nginx with default `proxy_buffering` forwards
  events immediately instead of queuing them — previously latency degraded
  from milliseconds to seconds behind a default nginx. Pinned by
  `TestHandler_AccelBufferingHeader`.
- Replay/store failure is no longer silent: before aborting the connection,
  the handler emits a named `event: error` carrying a `retry: 30000`
  reconnection-backoff hint. A silent abort is indistinguishable from a
  network blip, so browsers reconnected immediately into the same failing
  store — a reconnect storm. Pinned by
  `TestHandler_StoreFailure_SendsErrorEventBeforeAbort`.
- Resolved all golangci-lint findings: extracted `forwardLive` to bring `Handler` under the gocognit threshold, context-aware HTTP test helper (`httpGetURL`), array-based read buffers (makezero), justified nolint directives on store pass-throughs and the recover-dependent named return.


## [0.1.0] - 2026-08-16

First tagged release of the realtime module. Requires `GOEXPERIMENT=jsonv2`
(transitive: go-sse → go-branded-id). **SSE only** — no WebSocket support,
provided, or planned. Depends on `go-sse v0.5.0` only: no core, no go-cqrs-lite,
no go-datastar dependency.

### Added

- `Hub` — pairs a `sse.Broadcaster[sse.Event]` with an optional `sse.EventStore`
  (`WithStore`), plus `WithBufferSize`, `WithOnSubscribe`, `WithOnUnsubscribe`:
  `NewHub`, `Broadcast`, `BroadcastMany`, `BroadcastPatch` (duck-typed
  `PatchLike` — works with go-datastar patches without importing go-datastar),
  `Shutdown` (graceful drain), `Close` (instant), `Health` snapshot,
  `SubscriberCount`.
- `Handler` — the canonical SSE endpoint: CORS → subscribe → replay-with-dedup →
  heartbeat → forward, with `WithHeartbeat`, `WithCORSOrigin`, and `WithFilter`
  functional options. `Mount` registers it on a stdlib mux.
- SSE headers flush immediately after `NewStream`, so clients receive `200 OK`
  without waiting for the first event.
- Replay-to-live gap closed: subscribe-before-replay ordering with live-event
  dedup, so no event is lost or doubled across the reconnect boundary.
- Request context threaded through shutdown, replay, and stream lifecycles.
