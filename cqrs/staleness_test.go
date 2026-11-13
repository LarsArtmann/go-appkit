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
	"github.com/larsartmann/go-error-family/errorfamilytest"
)

func TestEventService_CheckStaleness_FreshWithoutProcessedEvents(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	// No event processed yet: lag is 0, so even a nanosecond budget passes.
	err := eventSvc.CheckStaleness(time.Nanosecond)
	if err != nil {
		t.Errorf("expected fresh (nil) with no processed events, got: %v", err)
	}
}

func TestEventService_CheckStaleness_DisabledByNonPositiveBudget(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	for _, budget := range []time.Duration{0, -time.Second} {
		err := eventSvc.CheckStaleness(budget)
		if err != nil {
			t.Errorf("expected disabled check (nil) for budget %v, got: %v", budget, err)
		}
	}
}

func TestEventService_CheckStaleness_FreshProjectionWithinBudget(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	startNoOpProjection(t, eventSvc, "fresh-projection", "test.fresh")

	// The projection has processed an event and is well within a 1-hour budget.
	err := eventSvc.CheckStaleness(time.Hour)
	if err != nil {
		t.Errorf("expected nil within generous budget after processing, got: %v", err)
	}
}

func TestEventService_CheckStaleness_StaleProjectionIsTransient(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	startNoOpProjection(t, eventSvc, "stale-projection", "test.stale")

	// After the worker processes the event, lag = time.Since(lastProcessedAt).
	// A nanosecond budget is already exceeded — the projection is stale.
	err := eventSvc.CheckStaleness(time.Nanosecond)
	if err == nil {
		t.Fatal("expected staleness error with nanosecond budget after processing an event")
	}

	if !errors.Is(err, projectionhost.ErrProjectionStale) {
		t.Errorf("expected errors.Is(err, ErrProjectionStale), got: %v", err)
	}

	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertRetryable(t, err, true)
	errorfamilytest.AssertHTTPStatus(t, err, 503)
}

func TestEventService_CheckProjectionStaleness_FreshProjectionWithinBudget(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	startNoOpProjection(t, eventSvc, "named-fresh", "test.named-fresh")

	err := eventSvc.CheckProjectionStaleness("named-fresh", time.Hour)
	if err != nil {
		t.Errorf("expected nil within generous budget, got: %v", err)
	}
}

func TestEventService_CheckProjectionStaleness_StaleProjectionIsTransient(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	startNoOpProjection(t, eventSvc, "named-stale", "test.named-stale")

	err := eventSvc.CheckProjectionStaleness("named-stale", time.Nanosecond)
	if err == nil {
		t.Fatal("expected staleness error with nanosecond budget after processing an event")
	}

	if !errors.Is(err, projectionhost.ErrProjectionStale) {
		t.Errorf("expected errors.Is(err, ErrProjectionStale), got: %v", err)
	}

	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertRetryable(t, err, true)
	errorfamilytest.AssertHTTPStatus(t, err, 503)
}

func TestEventService_CheckProjectionStaleness_UnknownProjectionRejected(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	err := eventSvc.CheckProjectionStaleness("no-such-projection", time.Hour)
	if err == nil {
		t.Fatal("expected error for unregistered projection name")
	}

	if !errors.Is(err, projectionhost.ErrProjectionStale) {
		// Unknown-name rejection, not staleness — assert family and code.
		errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
		errorfamilytest.AssertCode(t, err, "projectionhost.unknown_projection")
		errorfamilytest.AssertRetryable(t, err, false)
		errorfamilytest.AssertHTTPStatus(t, err, 400)
	}
}

func TestEventService_CheckProjectionStaleness_DisabledBeforeRegistrationCheck(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	// A non-positive budget returns nil before the name is looked up.
	err := eventSvc.CheckProjectionStaleness("no-such-projection", 0)
	if err != nil {
		t.Errorf("expected disabled check (nil), got: %v", err)
	}
}

// TestEventService_CheckStaleness_BudgetMonotonicity pins the boundary
// property: staleness is monotone non-increasing in the budget — if a
// smaller budget passes, every larger budget passes too. This guards the
// exactly-at-threshold comparison against off-by-one drift (lag > budget
// must stay STRICTLY greater; a >= regression would flip equal-to-budget
// deployments to stale) without pretending wall-clock equality is
// deterministic.
func TestEventService_CheckStaleness_BudgetMonotonicity(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	startNoOpProjection(t, eventSvc, "mono-projection", "test.mono")

	// Sample budgets from nanoseconds to hours; outcomes must be monotone
	// (once passing, always passing as the budget grows).
	budgets := []time.Duration{
		time.Nanosecond, time.Microsecond, time.Millisecond,
		time.Second, time.Minute, time.Hour,
	}

	passed := false
	for _, budget := range budgets {
		err := eventSvc.CheckStaleness(budget)
		if err == nil {
			passed = true

			continue
		}

		if passed {
			t.Errorf("budget %v failed after a smaller budget passed — staleness is not monotone in the budget", budget)
		}

		if !errors.Is(err, projectionhost.ErrProjectionStale) {
			t.Errorf("budget %v: expected ErrProjectionStale, got: %v", budget, err)
		}
	}
}

// startNoOpProjection drives an EventService to the state every staleness
// test needs: a named no-op projection over eventType is registered, one
// event of that type is appended, the projection workers are started, and
// the service has caught up (ReadyCheck green).
func startNoOpProjection(t *testing.T, eventSvc *EventService, name string, eventType event.Type) {
	t.Helper()

	proj := projection.NewProjection(
		name,
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{eventType},
	)

	err := eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, eventType)

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "projection caught up", eventSvc.ReadyCheck)
}
