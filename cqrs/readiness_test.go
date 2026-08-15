package cqrs

import (
	"context"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
)

func TestEventService_ReadyCheck_NoProjectionsReady(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	if !eventSvc.ReadyCheck() {
		t.Error("expected ready with no projections registered")
	}
}

func TestEventService_ReadyCheck_503To200Transition(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"ready-projection",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.ready"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	// Idle before StartProjections: not ready.
	if eventSvc.ReadyCheck() {
		t.Fatal("expected not ready while worker is idle")
	}

	appendTestEvent(t, eventSvc, "test.ready")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	// After drain the worker stops: ready. (Without WithSubscriber the host
	// is a batch drainer; "stopped" means caught up, not broken.)
	waitFor(t, "ready after projections caught up", eventSvc.ReadyCheck)
}

func TestEventService_ReadyCheck_FailedProjectionNotReady(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		HostOptions: []projectionhost.HostOption{
			projectionhost.WithMaxRestarts(0),
			projectionhost.WithBackoff(1, 1),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"doomed-projection",
		func(_ context.Context, _ event.Event) error { return errPoison },
		[]event.Type{"test.doomed"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.doomed")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "WorkerFailed", func() bool {
		for _, state := range eventSvc.Host().Status() {
			if state.Name == "doomed-projection" && state.Status == projectionhost.WorkerFailed {
				return true
			}
		}

		return false
	})

	if eventSvc.ReadyCheck() {
		t.Error("expected not ready with a failed projection")
	}
}

func TestEventService_LagPerProjection(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	lag := eventSvc.LagPerProjection()
	if lag == nil {
		t.Fatal("expected non-nil lag map")
	}
}
