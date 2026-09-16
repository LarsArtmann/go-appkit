# realtime — SSE for go-appkit services

Thin lifecycle and integration layer for [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events),
built on [go-sse](https://github.com/LarsArtmann/go-sse). Pairs go-sse's
`Broadcaster` (fan-out) and `EventStore` (reconnection replay) in a `Hub`, and
provides `Handler`/`Mount` — the canonical SSE endpoint go-sse deliberately
does not include.

**SSE only.** No WebSocket support is provided or planned.

## Install

```bash
go get github.com/larsartmann/go-appkit/realtime
```

> Requires `GOEXPERIMENT=jsonv2` only on older 1.26.x toolchains (go-sse's
> dependency chain uses `encoding/json/v2`; Go 1.26.7 enables it by default).

## Quick start

```go
hub := realtime.NewHub()
realtime.Mount(svc.Mux, "GET /events", hub)

// Push events from anywhere:
hub.Broadcast(sse.Event{Event: "update", Data: `{"count": 42}`})
```

`Mount` works on any `*http.ServeMux` — no appkit core dependency.

## Reconnection replay

When a browser reconnects it sends `Last-Event-ID`. Give the Hub an
`sse.EventStore` and the handler replays exactly the missed suffix before live
delivery, deduplicated against live events (the handler subscribes BEFORE
reading the store, so no event slips between replay and live):

```go
hub := realtime.NewHub(realtime.WithStore(myStore))
```

Journal-backed replay for CQRS event stores: wire
`transport.NewJournalSSEStore(journal, mapper)` from
`github.com/larsartmann/cqrs-htmx/v4/transport` — proven end-to-end by the
`integration` module in the go-appkit repo.

## DataStar integration

`Hub.BroadcastPatch` accepts any type with an `Event() sse.Event` method
(`PatchLike`), so DataStar patches flow through without importing go-datastar:

```go
hub.BroadcastPatch(datastar.NewElementsPatch("<div>Hi</div>",
    datastar.WithSelector("#feed")))
```

## Options

| Option                               | Default | Effect                              |
| ------------------------------------ | ------- | ----------------------------------- |
| `NewHub(WithStore(store))`           | nil     | Enable `Last-Event-ID` replay       |
| `NewHub(WithBufferSize(n))`          | 64      | Per-subscriber event buffer         |
| `Mount(..., WithHeartbeat(d))`       | 15s     | Comment-ping interval; `0` disables |
| `Mount(..., WithCORSOrigin(origin))` | `*`     | Tighten for production              |
| `Mount(..., WithFilter(pred))`       | nil     | Per-endpoint event predicate        |

The handler flushes headers immediately after `NewStream`, so clients and
reverse proxies get the 200 OK without waiting for the first event.

## Shutdown ordering

Drain SSE clients BEFORE the HTTP server shuts down, so browsers reconnect to
another instance while connections are still alive:

```go
_ = hub.Shutdown(ctx) // stop fan-out, close streams gracefully
_ = svc.Shutdown(ctx) // then the server
```

Wire `hub.Shutdown` into `ServiceConfig.DrainHooks` or call it ahead of
`svc.Shutdown` in your composition root.

## Known limitations

- Burst losses larger than the subscriber buffer (default 64) during a slow
  replay read are dropped; clients heal via `Last-Event-ID` reconnect.
- Single-process broadcaster only — no cross-instance fan-out (put a real bus
  in front if you scale horizontally).
- Store-failure replay aborts now send a named `event: error` with a
  `retry: 30000` backoff hint before dropping the connection (fixed
  2026-09-16 — previously the abort was silent, causing immediate reconnects
  against the same dead store). The hint is advisory: non-browser clients
  choose their own backoff.
- The handler sets `X-Accel-Buffering: no` so nginx does not buffer the
  stream (fixed 2026-09-16); other proxies may need their own
  buffering-off configuration.

## License

PROPRIETARY — see [LICENSE](LICENSE).
