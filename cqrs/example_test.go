package cqrs_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// The canonical wiring: SQLite event store and projection lifecycle
// logging. The service is NOT ready before StartProjections — the
// system auto-projection worker is idle until started, so /health/ready
// serves 503 during startup and flips once workers are caught up.
func ExampleNewEventService() {
	dir, err := os.MkdirTemp("", "cqrs-example")
	if err != nil {
		fmt.Println("temp dir:", err)

		return
	}

	defer func() { _ = os.RemoveAll(dir) }()

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		DSN:    filepath.Join(dir, "events.db"),
		Logger: slog.Default(),
	})
	if err != nil {
		fmt.Println("construct:", err)

		return
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	fmt.Println("ready before start:", es.ReadyCheck())

	err = es.Host().Register(projection.NewProjection(
		"example-projection",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"example.event"},
	))
	if err != nil {
		fmt.Println("register:", err)

		return
	}

	fmt.Println("registered workers:", len(es.Host().Status()))

	err = es.StartProjections(context.Background())
	if err != nil {
		fmt.Println("start:", err)

		return
	}

	deadline := time.Now().Add(5 * time.Second)

	for !es.ReadyCheck() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("ready after start:", es.ReadyCheck())

	// Output:
	// ready before start: false
	// registered workers: 2
	// ready after start: true
}

// The DLQ keeps a poison event from stalling a projection: after Threshold
// failures the event is quarantined and the checkpoint advances. Replay is
// a pure retry against the current store contents — with nothing
// quarantined it succeeds and replays nothing; entries reported in
// ReplayResult.Replayed must be removed from the store by the caller.
func ExampleEventService_ReplayDeadLetters() {
	dir, err := os.MkdirTemp("", "cqrs-dlq")
	if err != nil {
		fmt.Println("temp dir:", err)

		return
	}

	defer func() { _ = os.RemoveAll(dir) }()

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		DSN: filepath.Join(dir, "events.db"),
		DLQ: &cqrs.DLQConfig{}, // SQLite store in the event database, threshold 3
	})
	if err != nil {
		fmt.Println("construct:", err)

		return
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	result, err := es.ReplayDeadLetters(context.Background(), "user-projection")
	if err != nil {
		fmt.Println("replay:", err)

		return
	}

	fmt.Println("replayed:", len(result.Replayed))

	// Output:
	// replayed: 0
}

// Domain declarations replace hand-rolled host projections with typed
// read models: declare a QuerySet once, and the planner builds the
// collection, its indexes, and the fold wiring. Reads go through
// system.Find/Get against es.System().
func ExampleNewEventService_domain() {
	type taskView struct {
		ID       string
		Title    string
		Status   string
		Priority int
	}

	type taskCreated struct {
		ID       string
		Title    string
		Status   string
		Priority int
	}

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		Driver: "memory",
		Domain: &system.DomainConfig{
			Projections: []system.ProjectionDeclaration{
				system.QuerySet[taskView]("tasks").
					On("task.created", taskCreated{}).
					Filterable("status").
					Sortable("priority", false).
					Done(),
			},
		},
	})
	if err != nil {
		fmt.Println("construct:", err)

		return
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	// Events normally arrive from command handlers; appending straight to
	// the journal keeps the example focused on the read side.
	for _, task := range []taskCreated{
		{ID: "task-1", Title: "Write domain tests", Status: "active", Priority: 2},
		{ID: "task-2", Title: "Ship v0.7.0", Status: "active", Priority: 5},
	} {
		streamID, streamErr := id.ParseStreamID(task.ID)
		if streamErr != nil {
			fmt.Println("stream id:", streamErr)

			return
		}

		evt, evtErr := event.New("task.created", streamID, "tasks", 1, task)
		if evtErr != nil {
			fmt.Println("event:", evtErr)

			return
		}

		saveErr := es.System().EventStore().Save(
			context.Background(), id.NewStreamRef("tasks", streamID), []event.Event{evt}, 0)
		if saveErr != nil {
			fmt.Println("save:", saveErr)

			return
		}
	}

	startErr := es.StartProjections(context.Background())
	if startErr != nil {
		fmt.Println("start:", startErr)

		return
	}

	ctx := context.Background()

	deadline := time.Now().Add(5 * time.Second)

	var active []taskView

	for time.Now().Before(deadline) {
		active, err = system.Find[taskView](ctx, es.System(), "tasks",
			system.Where("status", "active"),
			system.OrderBy("priority", system.Desc))
		if err == nil && len(active) == 2 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	for _, task := range active {
		fmt.Println(task.Title)
	}

	// Output:
	// Ship v0.7.0
	// Write domain tests
}
