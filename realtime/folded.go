package realtime

import (
	"context"
	"encoding/json/v2"

	"github.com/larsartmann/go-sse"
)

// FoldedSignal is what a folded-notification channel carries: the identity
// of the event whose projection fold completed. The appkit cqrs module's
// FoldedEvent satisfies it structurally — neither module imports the
// other; the composition is the consumer's one line:
//
//	folded := make(chan cqrs.FoldedEvent, 64)
//	proj := cqrs.NotifyFolded(myProjection, folded)
//	realtime.FoldedSource(ctx, folded, func() any { return readModel.Snapshot() }, hub)
type FoldedSignal interface {
	// EventName is the trigger event's type; it becomes the SSE event name
	// of the re-broadcast, so clients subscribe by what changed.
	EventName() string

	// EventID is the trigger event's ID; it becomes the SSE event ID, so
	// reconnection replay (Last-Event-ID) deduplicates full-state payloads.
	EventID() string
}

// FoldedSource turns folded signals into hub broadcasts, owning the
// broadcast-strictly-after-fold ordering so consumers cannot get it wrong:
// for every signal it receives, it calls snapshot() — only after the
// signal, never concurrently with the fold (the signal IS the
// happens-after edge) — and broadcasts the result as a full-state event
// named after the trigger.
//
// The pump runs until ctx is cancelled or the folded channel closes. Each
// of these is one goroutine receiving from one channel, so per-projection
// broadcast order is preserved; with several projections feeding one hub,
// give each its own channel and FoldedSource — the hub serializes
// fan-out.
//
// Full-state payloads are the invariant that makes drops safe: a folded
// signal dropped by the notifier's bounded buffer (see cqrs.NotifyFolded)
// or a broadcast dropped by a slow subscriber only delays re-sync — the
// next delivered broadcast carries the whole state, not a delta, so no
// client is left showing stale data. That is also why this module keeps
// no drop counter of its own: the notifier's DroppedNotifications is the
// single observability seam for the folded side (the W5-C1 counter work
// dedupes there, not here).
//
// A snapshot that cannot be JSON-encoded skips its broadcast (the payload
// is the point of the re-announce); that is a programming error in the
// snapshot function, and the invariant makes it self-healing on the next
// signal. An event ID that go-sse rejects (control characters) is
// likewise omitted — the broadcast still fires.
func FoldedSource[S FoldedSignal](ctx context.Context, folded <-chan S, snapshot func() any, hub *Hub) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case sig, ok := <-folded:
				if !ok {
					return
				}

				broadcastFoldedSnapshot(hub, sig, snapshot())
			}
		}
	}()
}

// broadcastFoldedSnapshot is the single broadcast step, split out so the
// ordering contract has one name: snapshot was taken by the caller AFTER
// the fold signal arrived; this function only renders and sends.
func broadcastFoldedSnapshot[S FoldedSignal](hub *Hub, sig S, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	evt := sse.Event{ //nolint:exhaustruct_v5 // ID is attached conditionally below; Retry is deliberately unset
		Event: sig.EventName(),
		Data:  string(data),
	}

	if id, err := sse.ParseEventID(sig.EventID()); err == nil {
		evt.ID = id
	}

	hub.Broadcast(evt)
}
