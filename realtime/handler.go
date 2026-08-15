package realtime

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/larsartmann/go-sse"
)

// defaultHeartbeat is the interval between SSE comment-frame pings sent to
// keep connections alive through reverse proxies (Nginx, Cloudflare, AWS ALB).
// 15 seconds is short enough to reset typical 60s proxy idle timers while
// keeping bandwidth negligible.
const defaultHeartbeat = 15 * time.Second

// defaultCORSOrigin is the permissive CORS origin matching the go-sse example.
// Consumers should tighten this for production via [WithCORSOrigin].
const defaultCORSOrigin = "*"

// MountOption configures the SSE endpoint handler registered by [Mount].
type MountOption func(*mountConfig)

type mountConfig struct {
	heartbeat time.Duration
	cors      string
	filter    func(sse.Event) bool
}

// WithHeartbeat sets the interval between SSE comment-frame keepalive pings.
// Pass 0 to disable heartbeat entirely. Default: 15s.
func WithHeartbeat(d time.Duration) MountOption {
	return func(c *mountConfig) { c.heartbeat = d }
}

// WithCORSOrigin sets the Access-Control-Allow-Origin header value. Pass "" to
// suppress the header entirely (when SSE is same-origin only). Default: "*".
func WithCORSOrigin(origin string) MountOption {
	return func(c *mountConfig) { c.cors = origin }
}

// WithFilter sets a predicate for filtered subscriptions — only events for
// which the predicate returns true are delivered to the subscriber's buffer.
// The predicate is also pushed into replay via [sse.ReplayFiltered] when a
// store is configured. It must be pure, fast, and non-blocking.
func WithFilter(pred func(sse.Event) bool) MountOption {
	return func(c *mountConfig) { c.filter = pred }
}

// Handler returns an http.Handler that implements the canonical SSE endpoint:
//
//  1. Sets CORS headers (configurable via [WithCORSOrigin])
//  2. Creates an [sse.Stream] (sets SSE headers + 200 OK)
//  3. Subscribes to the broadcaster FIRST (optionally filtered) — live
//     events buffer in the subscriber channel while replay runs, so no
//     event can slip between the store snapshot and the subscription
//  4. Replays missed events from the Hub's store (if Last-Event-ID present),
//     recording replayed IDs
//  5. Starts a heartbeat goroutine (configurable via [WithHeartbeat])
//  6. Forwards buffered and live events to the stream, skipping IDs that
//     were already replayed (subscribe-before-replay makes overlap possible;
//     dedup keeps delivery exactly-once), until the client disconnects
//
// Residual risk: a burst larger than the subscriber buffer (default 64)
// during a slow store read can still drop live events; clients heal via the
// standard Last-Event-ID reconnect, which replays exactly the gap.
//
// This is the opinionated handler that go-sse deliberately does not provide.
// Use [Mount] for stdlib mux convenience, or [Handler] directly with any
// router that accepts [http.Handler].
func Handler(hub *Hub, opts ...MountOption) http.Handler {
	cfg := mountConfig{
		heartbeat: defaultHeartbeat,
		cors:      defaultCORSOrigin,
		filter:    nil,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.cors != "" {
			w.Header().Set("Access-Control-Allow-Origin", cfg.cors)
		}

		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		// Flush headers immediately so the client receives the 200 OK
		// without waiting for the first event or heartbeat.
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		ctx := stream.Context()

		// Subscribe before replay: everything broadcast from here on buffers
		// in the channel and is forwarded after the replayed events (minus
		// dedup), closing the snapshot-to-subscription gap the old
		// replay-then-subscribe order had.
		var ch <-chan sse.Event

		if cfg.filter != nil {
			ch = hub.Broadcaster.SubscribeFilter(cfg.filter)
		} else {
			ch = hub.Broadcaster.Subscribe()
		}

		defer hub.Broadcaster.Unsubscribe(ch)

		replayed, ok := replayMissedEvents(stream, hub, cfg.filter)
		if !ok {
			return
		}

		if cfg.heartbeat > 0 {
			go stream.Heartbeat(ctx, cfg.heartbeat)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-ch:
				if !ok {
					return
				}

				// Skip events that raced the store snapshot: appended after
				// subscribe (so they are in this channel) and before the
				// snapshot read (so they were also replayed).
				if id := evt.ID.Get(); id != "" {
					if _, dup := replayed[id]; dup {
						continue
					}
				}

				if stream.Send(evt) != nil {
					return
				}
			}
		}
	})
}

// Mount registers an SSE endpoint on the mux using [Handler]. This is the
// convenience entry point for stdlib mux consumers.
//
//	realtime.Mount(svc.Mux, "GET /events", hub,
//	    realtime.WithHeartbeat(15*time.Second),
//	    realtime.WithCORSOrigin("https://app.example.com"),
//	)
func Mount(mux *http.ServeMux, pattern string, hub *Hub, opts ...MountOption) {
	mux.Handle(pattern, Handler(hub, opts...))
}

// replayMissedEvents replays events from the store based on the Last-Event-ID
// header and returns the IDs it sent for live-loop deduplication. ok is false
// when the handler should abort: a store read failure or a write failure
// mid-replay means the stream state cannot be trusted. A nil map with ok true
// means nothing was replayed (no store, no Last-Event-ID, or empty result).
//
// Callers must have subscribed to the live broadcaster BEFORE calling this,
// so events racing the store snapshot are captured rather than lost.
func replayMissedEvents(
	stream *sse.Stream,
	hub *Hub,
	filter func(sse.Event) bool,
) (replayed map[string]struct{}, ok bool) {
	if hub.Store == nil {
		return nil, true
	}

	lastID := stream.LastEventID()
	if lastID.IsZero() {
		return nil, true
	}

	events, err := eventsAfter(hub.Store, lastID, filter)
	if err != nil {
		slog.ErrorContext(
			stream.Context(),
			"realtime: replay store read failed, aborting connection",
			"last_event_id", lastID.Get(),
			"err", err,
		)

		return nil, false
	}

	replayed = make(map[string]struct{}, len(events))

	for _, evt := range events {
		err := stream.Send(evt)
		if err != nil {
			slog.ErrorContext(
				stream.Context(),
				"realtime: replay send failed, aborting connection",
				"last_event_id", lastID.Get(),
				"err", err,
			)

			return replayed, false
		}

		if id := evt.ID.Get(); id != "" {
			replayed[id] = struct{}{}
		}
	}

	return replayed, true
}

// eventsAfter reads the store for replay. A nil filter is pushed to the plain
// EventsAfter call; a non-nil filter uses FilteredEventStore when the store
// implements it and otherwise falls back to EventsAfter plus in-memory
// post-filtering with panic recovery, mirroring sse.ReplayFiltered's
// fallback contract.
func eventsAfter(
	store sse.EventStore,
	lastID sse.EventID,
	filter func(sse.Event) bool,
) ([]sse.Event, error) {
	if filter == nil {
		return store.EventsAfter(lastID)
	}

	if filtered, ok := store.(sse.FilteredEventStore); ok {
		return filtered.EventsAfterFiltered(lastID, filter)
	}

	all, err := store.EventsAfter(lastID)
	if err != nil {
		return nil, err
	}

	events := make([]sse.Event, 0, len(all))
	for _, evt := range all {
		if safeFilter(filter, evt) {
			events = append(events, evt)
		}
	}

	return events, nil
}

// safeFilter applies pred with panic recovery, treating a panic as a
// non-match.
func safeFilter(pred func(sse.Event) bool, evt sse.Event) (match bool) {
	defer func() {
		if recover() != nil {
			match = false
		}
	}()

	return pred(evt)
}
