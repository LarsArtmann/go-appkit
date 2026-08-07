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
//  3. Replays missed events from the Hub's store (if Last-Event-ID present)
//  4. Subscribes to the broadcaster (optionally filtered)
//  5. Starts a heartbeat goroutine (configurable via [WithHeartbeat])
//  6. Forwards events from the subscription channel to the stream until
//     the client disconnects
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

		ctx := stream.Context()

		if !replayMissedEvents(stream, hub, cfg.filter) {
			return
		}

		var ch <-chan sse.Event

		if cfg.filter != nil {
			ch = hub.Broadcaster.SubscribeFilter(cfg.filter)
		} else {
			ch = hub.Broadcaster.Subscribe()
		}

		defer hub.Broadcaster.Unsubscribe(ch)

		if cfg.heartbeat > 0 {
			go stream.Heartbeat(ctx, cfg.heartbeat)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-ch:
				if !ok || stream.Send(evt) != nil {
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
// header. Returns false if the handler should abort (write failure during
// replay means the connection is dead).
func replayMissedEvents(stream *sse.Stream, hub *Hub, filter func(sse.Event) bool) bool {
	if hub.Store == nil {
		return true
	}

	lastID := stream.LastEventID()
	if lastID.IsZero() {
		return true
	}

	var err error

	if filter != nil {
		_, err = sse.ReplayFiltered(stream, hub.Store, lastID, filter)
	} else {
		_, err = sse.Replay(stream, hub.Store, lastID)
	}

	if err != nil {
		slog.ErrorContext(
			stream.Context(),
			"realtime: replay failed, aborting connection",
			"last_event_id", lastID.Get(),
			"err", err,
		)

		return false
	}

	return true
}
