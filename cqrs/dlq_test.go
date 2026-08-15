package cqrs

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
)

// newDLQService creates an EventService with the SQLite DLQ enabled at the
// given threshold.
func newDLQService(t *testing.T, threshold int) *EventService {
	t.Helper()

	es, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		DLQ:        &DLQConfig{Threshold: threshold},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Cleanup(func() { _ = es.Shutdown(context.Background()) })

	return es
}

// flakyProjection fails while broken is true and succeeds afterwards.
func flakyProjection(broken *atomic.Bool) projection.Projection {
	return projection.NewProjection(
		"dlq-projection",
		func(_ context.Context, _ event.Event) error {
			if broken.Load() {
				return errPoison
			}

			return nil
		},
		[]event.Type{"test.poison", "test.fine"},
	)
}

func TestEventService_DLQ_PoisonEventQuarantinedAndReplayed(t *testing.T) {
	t.Parallel()

	es := newDLQService(t, 2)

	broken := &atomic.Bool{}
	broken.Store(true)

	err := es.Host().Register(flakyProjection(broken))
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, es, "test.poison")
	appendTestEvent(t, es, "test.fine")

	err = es.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	dlq := es.DeadLetterStore()
	if dlq == nil {
		t.Fatal("expected DeadLetterStore to be non-nil when DLQ configured")
	}

	// The poison event must land in the dead-letter store.
	waitFor(t, "poison event in DLQ", func() bool {
		entries, err := dlq.List(context.Background(), "dlq-projection")
		if err != nil {
			return false
		}

		return len(entries) == 1
	})

	// The worker must NOT have failed: the checkpoint advanced past the
	// poison event and the projection kept running.
	waitFor(t, "worker to drain without failing", func() bool {
		for _, state := range es.Host().Status() {
			if state.Name == "dlq-projection" && state.Status == projectionhost.WorkerFailed {
				return false
			}
		}

		return true
	})

	// Fix the handler bug, then replay: the quarantined event succeeds.
	broken.Store(false)

	result, err := es.ReplayDeadLetters(context.Background(), "dlq-projection")
	if err != nil {
		t.Fatalf("replay dead letters: %v", err)
	}

	if len(result.Replayed) != 1 {
		t.Fatalf("expected 1 replayed entry, got %d (still failing: %d)",
			len(result.Replayed), len(result.StillFailing))
	}

	if result.Replayed[0].EventType != "test.poison" {
		t.Errorf("expected replayed event type test.poison, got %s",
			result.Replayed[0].EventType)
	}

	// Replay is pure: the entry stays until the caller purges it.
	err = dlq.Purge(context.Background(), "dlq-projection")
	if err != nil {
		t.Fatalf("purge: %v", err)
	}

	entries, err := dlq.List(context.Background(), "dlq-projection")
	if err != nil {
		t.Fatalf("list after purge: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries after purge, got %d", len(entries))
	}
}

func TestEventService_DLQ_DisabledByDefault(t *testing.T) {
	t.Parallel()

	es, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	if es.DeadLetterStore() != nil {
		t.Error("expected nil DeadLetterStore when DLQ not configured")
	}

	_, err = es.ReplayDeadLetters(context.Background(), "")
	if err == nil {
		t.Error("expected error from ReplayDeadLetters when DLQ disabled")
	}
}

func TestEventService_DLQ_MemoryStorePassthrough(t *testing.T) {
	t.Parallel()

	store := projectionhost.NewMemoryDeadLetterStore()

	es, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		DLQ:        &DLQConfig{Threshold: 1, Store: store},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	if es.DeadLetterStore() != store {
		t.Error("expected configured store to be returned verbatim")
	}
}

var errPoison = &poisonError{}

type poisonError struct{}

func (*poisonError) Error() string { return "poison" }
