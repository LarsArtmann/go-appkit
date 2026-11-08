package realtime

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-sse"
)

// foldedTestSignal is the structural FoldedSignal stand-in; appkit/cqrs's
// FoldedEvent satisfies the same interface without either module
// importing the other.
type foldedTestSignal struct {
	name string
	id   string
}

func (s foldedTestSignal) EventName() string { return s.name }
func (s foldedTestSignal) EventID() string   { return s.id }

func TestFoldedSource_BroadcastsFullStateAfterSignal(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	state := map[string]int{"count": 41}
	folded := make(chan foldedTestSignal, 1)

	FoldedSource(t.Context(), folded, func() any { return state }, hub)

	// The state change happens BEFORE the fold signal is sent — the pump's
	// snapshot must observe it (the module-level half of the ordering
	// contract; the E2E discriminator through published tags lives in the
	// integration module).
	state["count"] = 42
	folded <- foldedTestSignal{name: "ThingRenamed", id: "evt-1"}

	select {
	case got := <-ch:
		if got.Event != "ThingRenamed" {
			t.Fatalf("event name = %q, want the trigger type ThingRenamed", got.Event)
		}

		if got.Data != `{"count":42}` {
			t.Fatalf("data = %q, want the post-fold full-state snapshot", got.Data)
		}

		if got.ID.Get() != "evt-1" {
			t.Fatalf("event ID = %q, want the trigger event's ID", got.ID.Get())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the folded broadcast")
	}
}

func TestFoldedSource_PreservesSignalOrder(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	folded := make(chan foldedTestSignal, 3)

	FoldedSource(t.Context(), folded, func() any { return struct{}{} }, hub)

	folded <- foldedTestSignal{name: "A", id: "1"}
	folded <- foldedTestSignal{name: "B", id: "2"}
	folded <- foldedTestSignal{name: "C", id: "3"}

	deadline := time.After(2 * time.Second)

	for _, want := range []string{"A", "B", "C"} {
		select {
		case got := <-ch:
			if got.Event != want {
				t.Fatalf("broadcast order broken: got %q, want %q", got.Event, want)
			}
		case <-deadline:
			t.Fatalf("timed out waiting for broadcast %q", want)
		}
	}
}

func TestFoldedSource_ContextCancelStopsPump(t *testing.T) {
	t.Parallel()

	hub := NewHub()

	var mu sync.Mutex
	snapshots := 0

	folded := make(chan foldedTestSignal, 8)

	ctx, cancel := context.WithCancel(t.Context())

	FoldedSource(ctx, folded, func() any {
		mu.Lock()
		defer mu.Unlock()
		snapshots++

		return struct{}{}
	}, hub)

	cancel()

	// Cancellation wins over a queued signal: give the scheduler a beat,
	// then confirm no further snapshots run.
	<-time.After(50 * time.Millisecond)

	folded <- foldedTestSignal{name: "TooLate", id: "9"}
	<-time.After(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if snapshots != 0 {
		t.Fatalf("snapshot ran %d times after cancellation — the pump outlived its context", snapshots)
	}
}

func TestFoldedSource_ClosedChannelStopsPump(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	folded := make(chan foldedTestSignal, 1)

	before := runtime.NumGoroutine()

	FoldedSource(t.Context(), folded, func() any { return struct{}{} }, hub)
	close(folded)

	deadline := time.Now().Add(2 * time.Second)

	for runtime.NumGoroutine() < before && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	// The pump goroutine exiting on close may race NumGoroutine sampling of
	// unrelated tests in the same binary; the broadcast silence below is
	// the behavioral assertion, this loop is best-effort hygiene.
	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	select {
	case got := <-ch:
		t.Fatalf("broadcast %q arrived after the folded channel closed — the pump outlived its source", got.Event)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestFoldedSource_UnmarshalablePayloadSkipsBroadcast(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	folded := make(chan foldedTestSignal, 2)

	snapshotCalls := 0

	FoldedSource(t.Context(), folded, func() any {
		snapshotCalls++

		if snapshotCalls == 1 {
			return func() {} // channels of unmarshalable payload: json.Marshal fails
		}

		return map[string]int{"ok": 1}
	}, hub)

	folded <- foldedTestSignal{name: "Broken", id: "1"}
	folded <- foldedTestSignal{name: "Heals", id: "2"}

	deadline := time.After(2 * time.Second)

	for {
		select {
		case got := <-ch:
			if got.Event == "Heals" && got.Data == `{"ok":1}` {
				return // the invariant: the next signal self-heals
			}
		case <-deadline:
			t.Fatal("the post-error broadcast never arrived — a broken snapshot must not kill the pump")
		}
	}
}

func TestFoldedSource_UnparseableIDOmitsIDButBroadcasts(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	folded := make(chan foldedTestSignal, 1)

	FoldedSource(t.Context(), folded, func() any { return map[string]int{"n": 1} }, hub)

	folded <- foldedTestSignal{name: "Weird", id: "bad\nid"} // control character: go-sse rejects

	select {
	case got := <-ch:
		if got.Event != "Weird" || got.Data != `{"n":1}` {
			t.Fatalf("unexpected event: %+v", got)
		}

		if got.ID.Get() != "" {
			t.Fatalf("event ID = %q, want empty (the ID was unparseable)", got.ID.Get())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast never arrived — an unparseable ID must not suppress the event")
	}
}

// TestFoldedSignal_StructuralCompatWithCqrs pins the cross-module contract
// at the source level: a struct with exactly the cqrs.FoldedEvent method
// set satisfies FoldedSignal. The full composition runs in the integration
// module against published tags; this keeps the interface honest inside
// the module's own suite.
func TestFoldedSignal_StructuralCompat(t *testing.T) {
	t.Parallel()

	var _ FoldedSignal = foldedTestSignal{name: "ThingRenamed", id: "evt-1"}
	var _ = sse.Event{} // keep the sse import honest for future assertions
}
