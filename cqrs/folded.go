package cqrs

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
)

// FoldedEvent records that a projection finished folding one event: the
// projection's Handle returned successfully for it. It is the "fold done
// for E" signal of the broadcast-strictly-after-fold contract — a
// notification that announces a projection's state may only be emitted
// after the fold that incorporates its trigger event has completed, and
// the payload must reflect that fold (or a later one; folds are
// monotonic). See the W5-C2 design note (docs/planning) for the race this
// closes: a naive notify-on-append wiring broadcasts PRE-fold state and
// never re-announces, so clients show stale data until the next unrelated
// event.
//
// FoldedEvent satisfies realtime.FoldedSignal structurally — neither
// module imports the other; the composition is the consumer's one line.
type FoldedEvent struct {
	//cqrs-lint:ignore-start(C013,D014) in-memory notification struct: the fields are
	// unexported and getter-fed (realtime.FoldedSignal consumes EventName/EventID
	// getters), never CBOR/JSON-encoded — an event.Instant field or json tags here
	// would be inert ceremony on a value that crosses a Go channel, not a wire
	projectionName string
	eventType      string
	eventID        string
	occurredAt     time.Time
	//cqrs-lint:ignore-end
}

// ProjectionName is the name of the projection whose fold completed.
func (e FoldedEvent) ProjectionName() string { return e.projectionName }

// EventName is the type of the trigger event (the event whose fold just
// completed). As a realtime.FoldedSignal it becomes the SSE event name of
// the re-broadcast.
func (e FoldedEvent) EventName() string { return e.eventType }

// EventID is the ID of the trigger event. As a realtime.FoldedSignal it
// becomes the SSE event ID, so reconnection replay (Last-Event-ID)
// deduplicates full-state payloads.
func (e FoldedEvent) EventID() string { return e.eventID }

// OccurredAt is the trigger event's occurrence time.
func (e FoldedEvent) OccurredAt() time.Time { return e.occurredAt }

// FoldedNotifier wraps a projection so that every successfully folded
// event is recorded on the folded channel BEFORE Handle returns. The
// projection host sees an ordinary projection; the folded channel sees
// exactly the events whose folds completed, in fold order.
type FoldedNotifier struct {
	inner   projection.Projection
	folded  chan<- FoldedEvent
	dropped atomic.Uint64
}

// NotifyFolded wraps p so that each successful Handle(ctx, evt) sends a
// FoldedEvent to folded immediately after the inner Handle returns — the
// happens-after edge of the broadcast-strictly-after-fold contract. Wire
// it between the projection and the engine:
//
//	folded := make(chan cqrs.FoldedEvent, 64)
//	proj := cqrs.NotifyFolded(myProjection, folded)
//	// register proj where myProjection would have gone;
//	// feed folded to realtime.FoldedSource.
//
// The send is non-blocking: a full channel means the broadcast side is
// slower than the fold, and stalling the fold for the hub is exactly the
// wrong trade. A dropped notification only delays re-sync — the next
// broadcast carries a full-state payload (the W5 invariant). Count drops
// via DroppedNotifications. A nil channel records nothing (the select's
// send case is never ready); use it to disable the signal without
// unwiring the wrapper.
//
// A buffer of 64 is the recommended starting point (matching the realtime
// hub's per-subscriber default); projection rebuilds fold many events in
// bursts and drops are the documented behavior, not an error.
func NotifyFolded(p projection.Projection, folded chan<- FoldedEvent) *FoldedNotifier {
	return &FoldedNotifier{inner: p, folded: folded, dropped: atomic.Uint64{}}
}

// Name forwards the wrapped projection's name.
func (n *FoldedNotifier) Name() string { return n.inner.Name() }

// Handle folds evt through the wrapped projection and, only when the fold
// succeeded, records the completed event on the folded channel before
// returning. A failed fold does not notify: the fold did not complete, so
// announcing its state would violate the contract.
func (n *FoldedNotifier) Handle(ctx context.Context, evt event.Event) error {
	if err := n.inner.Handle(ctx, evt); err != nil {
		return err //nolint:wrapcheck // the projection's own error is the honest pass-through
	}

	if n.folded == nil {
		return nil // recording is off (the documented disable switch)
	}

	select {
	case n.folded <- FoldedEvent{
		projectionName: n.inner.Name(),
		eventType:      string(evt.Type()),
		eventID:        evt.ID().String(),
		occurredAt:     evt.OccurredAt(),
	}:
	default:
		n.dropped.Add(1)
	}

	return nil
}

// EventTypes forwards the wrapped projection's event type filter.
func (n *FoldedNotifier) EventTypes() []event.Type { return n.inner.EventTypes() }

// DroppedNotifications counts folded signals dropped because the channel
// was full. A growing counter means the broadcast side cannot keep up;
// raise the channel buffer or the hub's throughput. Backpressure is the
// hub's job, never the fold's — this counter is the observability seam
// for that decision (the W5-C1 drop-counter work dedupes here rather
// than growing a second counter realtime-side).
func (n *FoldedNotifier) DroppedNotifications() uint64 {
	return n.dropped.Load()
}

// compile-time contract: the wrapper IS a projection.
var _ projection.Projection = (*FoldedNotifier)(nil)
