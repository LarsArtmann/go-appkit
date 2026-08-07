# Realtime SSE Module Design — Revised

> **Date:** 2026-08-07. **Status:** Design complete, grounded in verified APIs.
> **Replaces:** The broken realtime proposal from the 2026-08-07 06:55 status report.
> **Constraint:** **SSE only. No WebSockets. Ever.**

---

## The Key Insight

**go-sse and go-datastar already exist and solve the hard problems.**
appkit/realtime is NOT a reimplementation — it is a thin lifecycle and integration layer.

| Problem                         | Already solved by                                                | appkit/realtime does              |
| ------------------------------- | ---------------------------------------------------------------- | --------------------------------- |
| SSE wire format                 | go-sse (`Event`, `WriteEvent`, `KeyedLines`)                     | Nothing — re-exports              |
| Single connection lifecycle     | go-sse (`Stream`: Send, Heartbeat, Context, LastEventID)         | Nothing — uses directly           |
| Fan-out / subscriber management | go-sse (`Broadcaster[T]`: Subscribe, SubscribeFilter, Broadcast) | Wraps in Hub for lifecycle        |
| Backpressure (drop-on-full)     | go-sse (non-blocking send, 64-buffer default, `WithBufferSize`)  | Nothing — passes through          |
| Graceful drain                  | go-sse (`Broadcaster.Shutdown(ctx)`)                             | Integrates into Service shutdown  |
| Health snapshot                 | go-sse (`Broadcaster.Health()`)                                  | Exposes via Hub                   |
| Reconnection replay             | go-sse (`EventStore`, `Replay`, `ReplayFiltered`)                | Pairs store with Hub              |
| DataStar protocol               | go-datastar (`Patch`, `ElementsPatch`, `SignalsPatch`, etc.)     | Documents integration pattern     |
| DataStar JS client              | go-datastar (`ScriptHandler()`)                                  | Documents mounting pattern        |
| Domain event → SSE bridge       | cqrs-htmx (`EventBridge` pattern)                                | Documents; consumer wires closure |

**What go-sse deliberately does NOT provide** (and appkit/realtime SHOULD):

> "No `Broadcaster.ServeSSE` convenience handler (would bake in heartbeat, replay, and event-loop opinions; the example/ package shows the canonical pattern)"
> — go-sse README

appkit/realtime provides exactly this: the opinionated canonical handler that go-sse refuses to bake in.

---

## Architecture

```
┌─ go-appkit/realtime ─────────────────────────────────────┐
│                                                           │
│  Hub                                                      │
│  ├── Broadcaster *sse.Broadcaster[sse.Event]  (fan-out)  │
│  └── Store       sse.EventStore               (replay)   │
│                                                           │
│  Mount(mux, pattern, hub, opts...)                        │
│  └── Canonical SSE handler:                               │
│      1. CORS headers (configurable)                       │
│      2. NewStream (sets SSE headers + 200 OK)             │
│      3. Replay from Store (if Last-Event-ID present)      │
│      4. Subscribe / SubscribeFilter                       │
│      5. Heartbeat goroutine                               │
│      6. Forward loop: channel → stream.Send               │
│                                                           │
└───────────────────────────────────────────────────────────│
                                                            ▼
                                              ┌─ go-sse ──────────────────┐
                                              │ Stream, Broadcaster[T],   │
                                              │ EventStore, Replay,        │
                                              │ Heartbeat, Shutdown        │
                                              └────────────────────────────┘
```

### Dependency chain

```
go-appkit/realtime
  └── github.com/larsartmann/go-sse v0.4.0
        ├── github.com/larsartmann/go-error-family v0.10.0
        └── github.com/larsartmann/go-branded-id v0.5.1
```

**No dependency on go-appkit core.** Realtime mounts on any `*http.ServeMux`.
This keeps it composable: works with appkit's Service, raw stdlib, or any framework.

**No dependency on go-datastar.** DataStar patches produce `sse.Event` via `patch.Event()`,
which the consumer broadcasts through the Hub. go-datastar is a consumer-side choice.

**No dependency on go-cqrs-lite.** Domain event wiring is a 3-line closure the consumer writes.
The cqrs-htmx EventBridge pattern is documented, not depended on.

---

## API Design

### Hub — lifecycle wrapper

```go
// Hub pairs an SSE Broadcaster with an optional EventStore for reconnection
// replay. Create one per realtime endpoint, share it across handlers, and
// call Shutdown during service drain.
type Hub struct {
    Broadcaster *sse.Broadcaster[sse.Event]
    Store       sse.EventStore // nil = no replay
}

func NewHub(opts ...Option) *Hub

type Option func(*config)
func WithStore(store sse.EventStore) Option
func WithBufferSize(size int) Option
func WithOnSubscribe(fn func()) Option
func WithOnUnsubscribe(fn func()) Option

// Lifecycle
func (h *Hub) Shutdown(ctx context.Context) error  // graceful drain
func (h *Hub) Close()                                // instant close
func (h *Hub) Health() sse.BroadcasterHealth

// Broadcasting
func (h *Hub) Broadcast(evt sse.Event)
func (h *Hub) BroadcastMany(evts ...sse.Event)

// DataStar convenience (no go-datastar import needed)
func (h *Hub) BroadcastPatch(p PatchLike)

// PatchLike is the duck-typed interface matching datastar.Patch.
// Any type with Event() sse.Event satisfies it.
type PatchLike interface { Event() sse.Event }

// Introspection
func (h *Hub) SubscriberCount() int
```

### Mount — canonical SSE handler

```go
// Mount registers an SSE endpoint on the mux. The handler implements the
// canonical pattern: CORS, replay, subscribe, heartbeat, forward.
func Mount(mux *http.ServeMux, pattern string, hub *Hub, opts ...MountOption)

type MountOption func(*mountConfig)
type mountConfig struct {
    heartbeat time.Duration
    cors      string
    filter    func(sse.Event) bool
}

func WithHeartbeat(d time.Duration) MountOption   // default: 15s
func WithCORSOrigin(origin string) MountOption     // default: "*" (permissive)
func WithFilter(pred func(sse.Event) bool) MountOption // default: nil (all events)
```

### Consumer wiring pattern

```go
// Create the hub (optionally with a replay store)
hub := realtime.NewHub()

// Mount the SSE endpoint on the service mux
realtime.Mount(svc.Mux, "GET /events", hub,
    realtime.WithHeartbeat(15*time.Second),
)

// Push events from anywhere:
hub.Broadcast(sse.Event{Event: "update", Data: `{"count": 42}`})

// Or with DataStar (consumer imports go-datastar):
hub.BroadcastPatch(datastar.NewElementsPatch("<div>Hi</div>",
    datastar.WithSelector("#feed")))

// Wire domain events (consumer's event bus → SSE):
eventBus.Subscribe(func(e DomainEvent) {
    hub.Broadcast(toSSEEvent(e))
})

// Graceful shutdown (drain SSE before HTTP):
errCh := svc.Start()
<-errCh
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
_ = hub.Shutdown(ctx)  // drain SSE subscribers first
_ = svc.Shutdown(ctx)   // then drain HTTP
```

---

## Reconnection Replay

When a browser reconnects after a network drop, it sends `Last-Event-ID`.
The canonical handler checks this and replays missed events from the Store:

```go
// Inside Mount's handler:
if lastID := stream.LastEventID(); !lastID.IsZero() && hub.Store != nil {
    sse.Replay(stream, hub.Store, lastID)
}
```

For filtered endpoints, `sse.ReplayFiltered` pushes the predicate into the store
query (if the store implements `FilteredEventStore`), or falls back to
in-memory post-filter.

### JournalStore adapter (future, not in v0.1.0)

A `JournalStore` type adapting CQRS event journals to `sse.EventStore` is
extracted from cqrs-htmx's proven pattern. Deferred to v0.2.0 because it
requires choosing a CQRS version (v3 vs v4), which is unresolved.

---

## Shutdown Integration

Fits into Decision 10's locked drain sequence:

```
1. Ready probe → 503           (Service.Shutdown flips readyProbe)
2. Wait drainDelay (5s)         (Service.Shutdown sleeps)
3. Server.Shutdown(ctx)         (HTTP server stops accepting)
4. Hub.Shutdown(ctx)            ← NEW: drain SSE subscriber buffers
5. Bundle.GracefulClose(ctx)    (if cqrs module is used)
6. Logger flush
```

Steps 4–5 order: SSE drains first (browser auto-reconnects to another instance),
then the event store closes. The consumer calls these in sequence; appkit does
not automate cross-module shutdown ordering in v0.1.0.

---

## What This Module Is NOT

- **No WebSocket support.** SSE only. This is a hard constraint.
- **No custom fan-out.** go-sse's Broadcaster is the fan-out. Period.
- **No custom backpressure.** go-sse's drop-on-full is the backpressure. Period.
- **No EventBridge type.** The consumer writes a 3-line closure to wire their event bus.
- **No go-datastar dependency.** DataStar is a consumer choice. BroadcastPatch uses a duck-typed interface.
- **No go-cqrs-lite dependency.** Domain event types are the consumer's concern.
- **No go-appkit core dependency.** Mount works on any *http.ServeMux.
- **No auth middleware.** Auth is the consumer's middleware stack (Decision 3).
- **No SSE event bus.** This is a transport + fan-out layer, not a domain event bus.

---

## SSE Authentication

SSE uses `EventSource` which cannot set custom headers from the browser.
Auth options (consumer's choice, not module's concern):

1. **Cookies** — EventSource sends cookies automatically. Best for same-origin.
2. **Query param** — `GET /events?token=abc`. Simple but tokens appear in logs.
3. **Service Worker** — Intercept and add auth headers. Most flexible.

The consumer applies auth via middleware on the SSE route before Mount's handler:

```go
svc.Mux.HandleFunc("GET /events", authMiddleware(sseHandler))
```

Or via `cfg.ExtraMiddlewares` in ServiceConfig.

---

## CORS for SSE

The canonical handler sets `Access-Control-Allow-Origin` before `NewStream`.
Default: `"*"` (permissive, like the go-sse example). Configurable via
`WithCORSOrigin("https://app.example.com")` for production tightening.

---

## Versioning

Per Decision 9: `go-appkit/realtime` tags `v0.1.0`. Experimental.
The API may change until proven in real services.

---

## Comparison with Previous (Broken) Design

| Aspect           | Previous Design (broken)         | Revised Design                                 |
| ---------------- | -------------------------------- | ---------------------------------------------- |
| SSE transport    | Custom Hub with custom fan-out   | go-sse Broadcaster (already built)             |
| Backpressure     | Custom DropOldest/CloseClient    | go-sse's drop-on-full (already built)          |
| Heartbeat        | Custom implementation            | go-sse's Stream.Heartbeat (already built)      |
| Replay           | Vague "SeekableJournal" mention  | go-sse's EventStore + Replay (already built)   |
| DataStar         | Not mentioned                    | go-datastar Patch interface via BroadcastPatch |
| Dependencies     | Implicit, unverified             | go-sse v0.4.0 only (explicit, verified)        |
| Core dependency  | Proposed changing Service API    | Zero core dependency — mounts on any mux       |
| EventBridge      | Custom type with unverified APIs | 3-line consumer closure, documented pattern    |
| WebSocket        | Mentioned as alternative         | **Eliminated. SSE only.**                      |
| App orchestrator | Proposed replacing Service       | Not needed — Hub + Mount compose with Service  |
