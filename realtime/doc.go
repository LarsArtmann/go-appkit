// Package realtime provides a thin lifecycle and integration layer for
// Server-Sent Events, built on [github.com/larsartmann/go-sse].
//
// It pairs go-sse's [sse.Broadcaster] (fan-out) and [sse.EventStore]
// (reconnection replay) in a [Hub] type, and provides [Handler] — the canonical
// SSE endpoint handler that go-sse deliberately does not include.
//
// # Quick start
//
//	hub := realtime.NewHub()
//	realtime.Mount(svc.Mux, "GET /events", hub)
//
//	// Push events from anywhere:
//	hub.Broadcast(sse.Event{Event: "update", Data: `{"count": 42}`})
//
// # DataStar integration
//
// DataStar patches (from [github.com/larsartmann/go-datastar]) produce
// [sse.Event] via patch.Event(). Use [Hub.BroadcastPatch] for the duck-typed
// convenience — it accepts any type with an Event() sse.Event method, so
// realtime does not need to import go-datastar:
//
//	hub.BroadcastPatch(datastar.NewElementsPatch("<div>Hi</div>",
//	    datastar.WithSelector("#feed")))
//
// # Reconnection replay
//
// When a browser reconnects after a network drop, it sends Last-Event-ID. If
// the Hub has an [sse.EventStore], [Handler] replays missed events before
// live delivery:
//
//	hub := realtime.NewHub(realtime.WithStore(myStore))
//
// # SSE only
//
// This module supports Server-Sent Events exclusively. No WebSocket support
// is provided or planned.
package realtime
