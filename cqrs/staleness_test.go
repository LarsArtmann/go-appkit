package cqrs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	errorfamily "github.com/larsartmann/go-error-family"
)

func TestEventService_CheckStaleness_FreshWithoutProcessedEvents(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	// No event processed yet: lag is 0, so even a nanosecond budget passes.
	err = eventSvc.CheckStaleness(time.Nanosecond)
	if err != nil {
		t.Errorf("expected fresh (nil) with no processed events, got: %v", err)
	}
}

func TestEventService_CheckStaleness_DisabledByNonPositiveBudget(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	for _, budget := range []time.Duration{0, -time.Second} {
		err := eventSvc.CheckStaleness(budget)
		if err != nil {
			t.Errorf("expected disabled check (nil) for budget %v, got: %v", budget, err)
		}
	}
}

func TestEventService_CheckStaleness_FreshProjectionWithinBudget(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"fresh-projection",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.fresh"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.fresh")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "projection caught up", eventSvc.ReadyCheck)

	// The projection has processed an event and is well within a 1-hour budget.
	err = eventSvc.CheckStaleness(time.Hour)
	if err != nil {
		t.Errorf("expected nil within generous budget after processing, got: %v", err)
	}
}

func TestEventService_CheckStaleness_StaleProjectionIsTransient(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"stale-projection",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.stale"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.stale")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "projection caught up", eventSvc.ReadyCheck)

	// After the worker processes the event, lag = time.Since(lastProcessedAt).
	// A nanosecond budget is already exceeded — the projection is stale.
	err = eventSvc.CheckStaleness(time.Nanosecond)
	if err == nil {
		t.Fatal("expected staleness error with nanosecond budget after processing an event")
	}

	if !errors.Is(err, projectionhost.ErrProjectionStale) {
		t.Errorf("expected errors.Is(err, ErrProjectionStale), got: %v", err)
	}

	familyErr, ok := errors.AsType[*errorfamily.Error](err)
	if !ok {
		t.Fatalf("expected *errorfamily.Error, got %T", err)
	}

	if familyErr.Family() != errorfamily.Transient {
		t.Errorf("expected family %q, got %q", errorfamily.Transient, familyErr.Family())
	}

	if got := errorfamily.HTTPStatus(err); got != 503 {
		t.Errorf("expected HTTP status 503 for Transient, got %d", got)
	}
}

func TestEventService_CheckProjectionStaleness_FreshProjectionWithinBudget(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"named-fresh",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.named-fresh"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.named-fresh")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "projection caught up", eventSvc.ReadyCheck)

	err = eventSvc.CheckProjectionStaleness("named-fresh", time.Hour)
	if err != nil {
		t.Errorf("expected nil within generous budget, got: %v", err)
	}
}

func TestEventService_CheckProjectionStaleness_StaleProjectionIsTransient(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"named-stale",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.named-stale"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.named-stale")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "projection caught up", eventSvc.ReadyCheck)

	err = eventSvc.CheckProjectionStaleness("named-stale", time.Nanosecond)
	if err == nil {
		t.Fatal("expected staleness error with nanosecond budget after processing an event")
	}

	if !errors.Is(err, projectionhost.ErrProjectionStale) {
		t.Errorf("expected errors.Is(err, ErrProjectionStale), got: %v", err)
	}

	familyErr, ok := errors.AsType[*errorfamily.Error](err)
	if !ok {
		t.Fatalf("expected *errorfamily.Error, got %T", err)
	}

	if familyErr.Family() != errorfamily.Transient {
		t.Errorf("expected family %q, got %q", errorfamily.Transient, familyErr.Family())
	}

	if got := errorfamily.HTTPStatus(err); got != 503 {
		t.Errorf("expected HTTP status 503 for Transient, got %d", got)
	}
}

func TestEventService_CheckProjectionStaleness_UnknownProjectionRejected(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	err = eventSvc.CheckProjectionStaleness("no-such-projection", time.Hour)
	if err == nil {
		t.Fatal("expected error for unregistered projection name")
	}

	if !errors.Is(err, projectionhost.ErrProjectionStale) {
		// Unknown-name rejection, not staleness — assert family and code.
		familyErr, ok := errors.AsType[*errorfamily.Error](err)
		if !ok {
			t.Fatalf("expected *errorfamily.Error, got %T", err)
		}

		if familyErr.Family() != errorfamily.Rejection {
			t.Errorf("expected family %q, got %q", errorfamily.Rejection, familyErr.Family())
		}

		if familyErr.Code() != "projectionhost.unknown_projection" {
			t.Errorf("expected code projectionhost.unknown_projection, got %q", familyErr.Code())
		}

		if got := errorfamily.HTTPStatus(err); got != 400 {
			t.Errorf("expected HTTP status 400 for Rejection, got %d", got)
		}
	}
}

func TestEventService_CheckProjectionStaleness_DisabledBeforeRegistrationCheck(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	// A non-positive budget returns nil before the name is looked up.
	err = eventSvc.CheckProjectionStaleness("no-such-projection", 0)
	if err != nil {
		t.Errorf("expected disabled check (nil), got: %v", err)
	}
}
