package cqrs_test

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	clprojections "github.com/larsartmann/go-cqrs-lite/commandlifecycle/projections/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
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

// exampleLazyStore delegates to the real event store once the service
// exists, so system.WithCommandLifecycle can be constructed before
// cqrs.NewEventService while still appending lifecycle events to the real
// journal. The DomainConfig.Commands hook binds the delegate during
// construction, before any dispatch can emit.
type exampleLazyStore struct {
	mu     sync.Mutex
	target event.Store
}

func (s *exampleLazyStore) bind(store event.Store) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.target = store
}

func (s *exampleLazyStore) delegate() event.Store {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.target
}

func (s *exampleLazyStore) Save(
	ctx context.Context, ref id.StreamRef, events []event.Event, expected event.Version,
) error {
	return s.delegate().Save(ctx, ref, events, expected)
}

func (s *exampleLazyStore) AppendBatch(
	ctx context.Context, ref id.StreamRef, events []event.Event,
) error {
	return s.delegate().AppendBatch(ctx, ref, events)
}

func (s *exampleLazyStore) Load(ctx context.Context, ref id.StreamRef) ([]event.Event, error) {
	return s.delegate().Load(ctx, ref)
}

func (s *exampleLazyStore) LoadFromVersion(
	ctx context.Context, ref id.StreamRef, version event.Version,
) ([]event.Event, error) {
	return s.delegate().LoadFromVersion(ctx, ref, version)
}

func (s *exampleLazyStore) LoadToVersion(
	ctx context.Context, ref id.StreamRef, maxVersion event.Version,
) ([]event.Event, error) {
	return s.delegate().LoadToVersion(ctx, ref, maxVersion)
}

func (s *exampleLazyStore) LoadToTimestamp(
	ctx context.Context, ref id.StreamRef, maxTime time.Time,
) ([]event.Event, error) {
	return s.delegate().LoadToTimestamp(ctx, ref, maxTime)
}

// Mirrors the README "Command lifecycle audit trail" recipe: ADR-0117's
// one-call lifecycle wiring plus the lazy store binding that solves the
// recorder-needs-a-store-before-construction ordering. Compile-checked
// corpus for the recipe; not executed (no Output).
func ExampleNewEventService_domainCommandLifecycle() {
	store := &exampleLazyStore{}
	cl := system.WithCommandLifecycle(store)

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		Driver: "memory",
		Domain: &system.DomainConfig{
			// Outer emits received/completed/dead-lettered; Attempt emits
			// failed/retried per attempt; retry middleware belongs between them.
			Middleware:  []command.Middleware{cl.OuterMiddleware, cl.AttemptMiddleware},
			Projections: cl.Projections, // prebuilt lifecycle read models
			Commands: func(sys *system.System) {
				store.bind(sys.EventStore())
			},
		},
	})
	if err != nil {
		fmt.Println("construct:", err)

		return
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	// Read the lifecycle models like any declared query (after commands
	// have flowed and the lifecycle projections caught up):
	counts, err := metaengine.ExecuteTypedByName[
		clprojections.RetryCountQuery, map[string]int64](
		context.Background(), es.System().MetaEngine(), clprojections.RetryCount().Name,
		clprojections.RetryCountQuery{})
	if err != nil {
		fmt.Println("retry counts:", err)

		return
	}

	fmt.Println("commands tracked:", len(counts))
}

// Mirrors the README "Streaming read models" recipe: watch a declared
// collection and stream every materialized change to EventSource clients.
// Reconnection replay is automatic when the client sends Last-Event-ID,
// capped by WithSSEReplayLimit. Compile-checked corpus for the recipe;
// not executed (no Output).
func ExampleEventService_System() {
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
					Done(),
			},
		},
	})
	if err != nil {
		fmt.Println("construct:", err)

		return
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	watcher := metaengine.NewWatcher[taskView](es.System().MetaEngine(), "tasks")

	defer watcher.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /events/tasks", func(w http.ResponseWriter, r *http.Request) {
		_ = metaengine.ServeSSE(w, r, watcher, //nolint:errcheck // stream ends with the request
			metaengine.WithSSEHeartbeat(30*time.Second))
	})

	_ = mux // mount on the appkit Service's mux in a real service
}
