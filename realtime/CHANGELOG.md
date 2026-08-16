# Changelog

## [Unreleased]

### Added

- Nothing yet.

### Fixed

- Nothing yet.

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
