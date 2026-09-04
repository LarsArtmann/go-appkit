package cqrs_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
)

// The canonical wiring: SQLite event store and projection lifecycle
// logging. With no projections registered the service reports ready, so
// the appkit service can start serving immediately and /health/ready
// flips as soon as workers are registered and catching up.
func ExampleNewEventService() {
	dir, err := os.MkdirTemp("", "cqrs-example")
	if err != nil {
		fmt.Println("temp dir:", err)

		return
	}

	defer func() { _ = os.RemoveAll(dir) }()

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		SQLitePath: filepath.Join(dir, "events.db"),
		Logger:     slog.Default(),
	})
	if err != nil {
		fmt.Println("construct:", err)

		return
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	fmt.Println("ready before any projection:", es.ReadyCheck())

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

	// Output:
	// ready before any projection: true
	// registered workers: 1
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
		SQLitePath: filepath.Join(dir, "events.db"),
		DLQ:        &cqrs.DLQConfig{}, // SQLite store in the event database, threshold 3
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
