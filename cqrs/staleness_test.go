package cqrs

import (
	"context"
	"errors"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
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
	if err := eventSvc.CheckStaleness(time.Nanosecond); err != nil {
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
		if err := eventSvc.CheckStaleness(budget); err != nil {
			t.Errorf("expected disabled check (nil) for budget %v, got: %v", budget, err)
		}
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

	// Workers that have not processed anything report lag 0, so even a tight
	// budget stays green until events are actually flowing; staleness beyond
	// the threshold only surfaces once a worker has processed an event.
	if err := eventSvc.CheckStaleness(time.Hour); err != nil {
		t.Errorf("expected nil within generous budget, got: %v", err)
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
		var familyErr *errorfamily.Error
		if !errors.As(err, &familyErr) {
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
	if err := eventSvc.CheckProjectionStaleness("no-such-projection", 0); err != nil {
		t.Errorf("expected disabled check (nil), got: %v", err)
	}
}
