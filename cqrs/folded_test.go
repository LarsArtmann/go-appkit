package cqrs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
)

func foldedTestEvent(t *testing.T, eventType string, version event.Version) event.Event {
	t.Helper()

	evt, err := event.NewEvent(event.Type(eventType), id.NewStreamID(), id.StreamType("thing"), version, []byte(`{}`))
	if err != nil {
		t.Fatalf("build event: %v", err)
	}

	return evt
}

func TestNotifyFolded_ForwardsProjectionIdentity(t *testing.T) {
	t.Parallel()

	inner := projection.NewProjection(
		"counts",
		func(context.Context, event.Event) error { return nil },
		[]event.Type{"ThingRenamed"},
	)
	folded := make(chan FoldedEvent, 1)
	wrapped := NotifyFolded(inner, folded)

	if wrapped.Name() != "counts" {
		t.Fatalf("Name() = %q, want %q", wrapped.Name(), "counts")
	}

	types := wrapped.EventTypes()
	if len(types) != 1 || string(types[0]) != "ThingRenamed" {
		t.Fatalf("EventTypes() = %v, want [ThingRenamed]", types)
	}
}

func TestNotifyFolded_RecordsFoldedEventAfterSuccessfulHandle(t *testing.T) {
	t.Parallel()

	inner := projection.NewProjection("counts", func(context.Context, event.Event) error { return nil }, nil)
	folded := make(chan FoldedEvent, 4)
	wrapped := NotifyFolded(inner, folded)

	evt1 := foldedTestEvent(t, "ThingRenamed", 1)
	evt2 := foldedTestEvent(t, "ThingRemoved", 2)

	if err := wrapped.Handle(context.Background(), evt1); err != nil {
		t.Fatalf("Handle evt1: %v", err)
	}

	if err := wrapped.Handle(context.Background(), evt2); err != nil {
		t.Fatalf("Handle evt2: %v", err)
	}

	first := <-folded
	second := <-folded

	if first.ProjectionName() != "counts" || second.ProjectionName() != "counts" {
		t.Fatalf("ProjectionName = %q / %q, want counts twice", first.ProjectionName(), second.ProjectionName())
	}

	if first.EventName() != "ThingRenamed" || second.EventName() != "ThingRemoved" {
		t.Fatalf("EventName = %q / %q, want the trigger types", first.EventName(), second.EventName())
	}

	if first.EventID() != evt1.ID().String() || second.EventID() != evt2.ID().String() {
		t.Fatal("EventID does not match the trigger event's ID")
	}

	if first.OccurredAt().IsZero() {
		t.Fatal("OccurredAt is zero — the trigger's occurrence time was not carried")
	}
}

func TestNotifyFolded_PreservesFoldOrder(t *testing.T) {
	t.Parallel()

	var handled []string

	inner := projection.NewProjection("ordered", func(_ context.Context, evt event.Event) error {
		handled = append(handled, string(evt.Type()))

		return nil
	}, nil)

	folded := make(chan FoldedEvent, 8)
	wrapped := NotifyFolded(inner, folded)

	for i, eventType := range []string{"A", "B", "C"} {
		if err := wrapped.Handle(context.Background(), foldedTestEvent(t, eventType, event.Version(i+1))); err != nil {
			t.Fatalf("Handle %s: %v", eventType, err)
		}
	}

	for _, want := range []string{"A", "B", "C"} {
		if got := (<-folded).EventName(); got != want {
			t.Fatalf("folded order broken: got %q, want %q", got, want)
		}
	}
}

func TestNotifyFolded_FailedFoldDoesNotNotify(t *testing.T) {
	t.Parallel()

	foldErr := errors.New("projection exploded")
	inner := projection.NewProjection("broken", func(context.Context, event.Event) error { return foldErr }, nil)
	folded := make(chan FoldedEvent, 1)
	wrapped := NotifyFolded(inner, folded)

	err := wrapped.Handle(context.Background(), foldedTestEvent(t, "ThingRenamed", 1))
	if !errors.Is(err, foldErr) {
		t.Fatalf("Handle error = %v, want the projection's own error", err)
	}

	select {
	case got := <-folded:
		t.Fatalf("folded channel received %v after a FAILED fold — the fold never completed", got.EventName())
	default:
	}
}

func TestNotifyFolded_NonBlockingSendDropsAndCounts(t *testing.T) {
	t.Parallel()

	inner := projection.NewProjection("counts", func(context.Context, event.Event) error { return nil }, nil)
	folded := make(chan FoldedEvent, 1) // deliberately tiny: one slot, no receiver
	wrapped := NotifyFolded(inner, folded)

	start := time.Now()

	for range 4 {
		if err := wrapped.Handle(context.Background(), foldedTestEvent(t, "ThingRenamed", 1)); err != nil {
			t.Fatalf("Handle: %v", err)
		}
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Handle blocked for %s — the fold stalled behind a full channel", elapsed)
	}

	if got := wrapped.DroppedNotifications(); got != 3 {
		t.Fatalf("DroppedNotifications() = %d, want 3 (first send filled the buffer)", got)
	}
}

func TestNotifyFolded_NilChannelIsNoOp(t *testing.T) {
	t.Parallel()

	inner := projection.NewProjection("counts", func(context.Context, event.Event) error { return nil }, nil)
	wrapped := NotifyFolded(inner, nil)

	done := make(chan struct{})

	go func() {
		_ = wrapped.Handle(context.Background(), foldedTestEvent(t, "ThingRenamed", 1))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Handle with a nil folded channel never returned — a nil-channel send must fall through to default")
	}

	if got := wrapped.DroppedNotifications(); got != 0 {
		t.Fatalf("DroppedNotifications() = %d, want 0 with a nil channel (recording is off, not failing)", got)
	}
}
