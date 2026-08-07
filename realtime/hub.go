package realtime

import (
	"context"

	"github.com/larsartmann/go-sse"
)

// Hub pairs an SSE [sse.Broadcaster] with an optional [sse.EventStore] for
// reconnection replay. Create one per realtime endpoint, share it across
// handlers, and call [Hub.Shutdown] during service drain.
//
// The Hub is a thin lifecycle wrapper — all fan-out, backpressure, filtering,
// and replay logic lives in go-sse. The Hub adds:
//
//   - Store pairing (Broadcaster + EventStore in one type)
//   - [Hub.BroadcastPatch] convenience for DataStar-style patches
//   - Lifecycle methods ([Shutdown], [Close], [Health]) for service integration
type Hub struct {
	Broadcaster *sse.Broadcaster[sse.Event]
	Store       sse.EventStore // nil = no replay
}

// Option configures a [Hub] at construction time.
type Option func(*hubConfig)

type hubConfig struct {
	store         sse.EventStore
	bufferSize    int
	onSubscribe   func()
	onUnsubscribe func()
}

// NewHub creates a Hub with a fresh [sse.Broadcaster] and optional configuration.
//
//	hub := realtime.NewHub()
//	hub := realtime.NewHub(
//	    realtime.WithStore(myStore),
//	    realtime.WithBufferSize(256),
//	)
func NewHub(opts ...Option) *Hub {
	cfg := hubConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}

	var bOpts []sse.Option[sse.Event]

	if cfg.bufferSize > 0 {
		bOpts = append(bOpts, sse.WithBufferSize[sse.Event](cfg.bufferSize))
	}

	b := sse.NewBroadcaster[sse.Event](bOpts...)

	if cfg.onSubscribe != nil {
		b.OnSubscribe(cfg.onSubscribe)
	}

	if cfg.onUnsubscribe != nil {
		b.OnUnsubscribe(cfg.onUnsubscribe)
	}

	return &Hub{
		Broadcaster: b,
		Store:       cfg.store,
	}
}

// WithStore sets the [sse.EventStore] for reconnection replay. When a browser
// reconnects with a Last-Event-ID header, [Handler] replays missed events from
// this store before delivering live events.
func WithStore(store sse.EventStore) Option {
	return func(c *hubConfig) { c.store = store }
}

// WithBufferSize sets the per-subscriber channel buffer capacity. The default
// is 64 (go-sse's [sse.defaultSubscriberBuffer]). Larger buffers absorb longer
// consumer slowdowns before events are dropped; smaller buffers reduce memory
// per subscriber at the cost of earlier drops.
func WithBufferSize(size int) Option {
	return func(c *hubConfig) { c.bufferSize = size }
}

// WithOnSubscribe registers a callback fired after each successful Subscribe.
// Use for connection metrics or triggering initial state sends.
func WithOnSubscribe(fn func()) Option {
	return func(c *hubConfig) { c.onSubscribe = fn }
}

// WithOnUnsubscribe registers a callback fired after each Unsubscribe.
// Use for connection metrics.
func WithOnUnsubscribe(fn func()) Option {
	return func(c *hubConfig) { c.onUnsubscribe = fn }
}

// Broadcast sends an event to all active subscribers. Slow subscribers with
// full buffers have the event silently dropped (go-sse's non-blocking send).
func (h *Hub) Broadcast(evt sse.Event) {
	h.Broadcaster.Broadcast(evt)
}

// BroadcastMany sends a batch of events to all subscribers in a single locked
// fan-out pass. Cheaper than looping [Hub.Broadcast].
func (h *Hub) BroadcastMany(evts ...sse.Event) {
	h.Broadcaster.BroadcastMany(evts...)
}

// PatchLike matches the datastar.Patch interface without importing go-datastar.
// Any type with an Event() sse.Event method satisfies it, including
// datastar.ElementsPatch, datastar.SignalsPatch, datastar.ScriptPatch, etc.
type PatchLike interface {
	Event() sse.Event
}

// BroadcastPatch converts a patch to an [sse.Event] via [PatchLike.Event] and
// broadcasts it. This is the convenience method for DataStar consumers —
// realtime does not import go-datastar.
func (h *Hub) BroadcastPatch(p PatchLike) {
	h.Broadcaster.Broadcast(p.Event())
}

// Shutdown gracefully drains the broadcaster: stops accepting new subscribers,
// waits for subscriber buffers to empty, then closes all channels. Call this
// BEFORE shutting down the HTTP server so browsers auto-reconnect to another
// instance. Returns ctx.Err() if the deadline fires before drain completes.
func (h *Hub) Shutdown(ctx context.Context) error {
	return h.Broadcaster.Shutdown(ctx)
}

// Close instantly closes all subscriber channels. Use during hard shutdown
// when in-flight events do not need to be delivered. For graceful drain,
// use [Hub.Shutdown].
func (h *Hub) Close() {
	h.Broadcaster.Close()
}

// Health returns a lifecycle snapshot suitable for health check endpoints
// (k8s liveness/readiness, load balancer probes).
func (h *Hub) Health() sse.BroadcasterHealth {
	return h.Broadcaster.Health()
}

// SubscriberCount returns the current number of active subscribers.
func (h *Hub) SubscriberCount() int {
	return h.Broadcaster.SubscriberCount()
}
