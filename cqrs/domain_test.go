package cqrs

// Tests for the EventConfig.Domain merge contract (mergeDomain) and the
// behavioral metaengine surface it unlocks through a live EventService:
// QuerySet/Find round-trips, the Events coeffect typo gate, bootstrap-skip,
// and Evolve/Lookup point reads.

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// ── Merge contract (pure unit) ──

func TestMergeDomain_NilDomain_KeepsWrapperBehavior(t *testing.T) {
	t.Parallel()

	merged := mergeDomain(EventConfig{}, newInFlightTracker(), nil)

	if len(merged.Projections) != 1 {
		t.Errorf("nil Domain: expected the single bootstrap projection, got %d", len(merged.Projections))
	}

	if len(merged.Middleware) != 1 {
		t.Errorf("nil Domain: expected only the in-flight tracker middleware, got %d", len(merged.Middleware))
	}

	if merged.CheckpointStore != nil {
		t.Error("nil Domain: expected nil CheckpointStore (system default)")
	}

	if len(merged.ProjectionHostOptions) != 0 {
		t.Errorf("nil Domain: expected no host options, got %d", len(merged.ProjectionHostOptions))
	}
}

func TestMergeDomain_MiddlewareOrder_TrackerThenDomainThenConfig(t *testing.T) {
	t.Parallel()

	tracker := newInFlightTracker()

	var order []string

	var pendingInsideDomainTag int

	domainMW := func(next command.Handler) command.Handler {
		return func(ctx context.Context, cmd command.Command) error {
			pendingInsideDomainTag = tracker.pendingCount()
			order = append(order, "domain")

			return next(ctx, cmd)
		}
	}

	cfgMW := func(next command.Handler) command.Handler {
		return func(ctx context.Context, cmd command.Command) error {
			order = append(order, "config")

			return next(ctx, cmd)
		}
	}

	merged := mergeDomain(EventConfig{
		CommandMiddleware: []command.Middleware{cfgMW},
		Domain:            &system.DomainConfig{Middleware: []command.Middleware{domainMW}},
	}, tracker, nil)

	if len(merged.Middleware) != 3 {
		t.Fatalf("expected 3 middleware (tracker, domain, config), got %d", len(merged.Middleware))
	}

	handler := command.Handler(func(_ context.Context, _ command.Command) error { return nil })
	for i := len(merged.Middleware) - 1; i >= 0; i-- {
		next, mw := handler, merged.Middleware[i]
		handler = mw(next)
	}

	cmd, err := command.New("merge.order", id.NewStreamID())
	if err != nil {
		t.Fatalf("create command: %v", err)
	}

	if err := handler(context.Background(), cmd); err != nil {
		t.Fatalf("invoke chain: %v", err)
	}

	if !slices.Equal(order, []string{"domain", "config"}) {
		t.Errorf("expected domain middleware before config middleware, got %v", order)
	}

	if pendingInsideDomainTag != 1 {
		t.Errorf("domain middleware must run INSIDE the tracker (pending=1), saw pending=%d",
			pendingInsideDomainTag)
	}

	if got := tracker.pendingCount(); got != 0 {
		t.Errorf("tracker must release the command after the chain, pending=%d", got)
	}
}

func TestMergeDomain_ConsumerProjections_ReplaceBootstrap(t *testing.T) {
	t.Parallel()

	consumer := []system.ProjectionDeclaration{
		system.Lookup[TaskView]("get-task").
			On("task.created", TaskCreated{}).
			Done(),
	}

	merged := mergeDomain(EventConfig{
		Domain: &system.DomainConfig{Projections: consumer},
	}, newInFlightTracker(), nil)

	if len(merged.Projections) != 1 {
		t.Fatalf("consumer declared 1 projection: expected it verbatim, got %d", len(merged.Projections))
	}

	if &merged.Projections[0] != &consumer[0] {
		t.Error("consumer projections must be passed through verbatim, not copied")
	}
}

func TestMergeDomain_EmptyDomainProjections_KeepBootstrap(t *testing.T) {
	t.Parallel()

	merged := mergeDomain(EventConfig{
		Domain: &system.DomainConfig{},
	}, newInFlightTracker(), nil)

	if len(merged.Projections) != 1 {
		t.Errorf("empty Domain projections: expected the bootstrap declaration, got %d", len(merged.Projections))
	}
}

func TestMergeDomain_HostOptions_DerivedAppendedLast(t *testing.T) {
	t.Parallel()

	marker := projectionhost.WithBatchSize(1)

	merged := mergeDomain(EventConfig{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Domain: &system.DomainConfig{ProjectionHostOptions: []projectionhost.HostOption{marker}},
	}, newInFlightTracker(), nil)

	// Domain's marker (1) + derived WithLogger (1); derived wiring sits
	// after consumer options so it wins conflicts (hostOptions contract).
	if len(merged.ProjectionHostOptions) != 2 {
		t.Errorf("expected domain option + derived WithLogger, got %d", len(merged.ProjectionHostOptions))
	}
}

func TestMergeDomain_CheckpointStore_ConfigWins_ThenDomainFallback(t *testing.T) {
	t.Parallel()

	fromDomain := &countingCheckpointStore{}
	fromConfig := &countingCheckpointStore{}

	fallback := mergeDomain(EventConfig{
		Domain: &system.DomainConfig{CheckpointStore: fromDomain},
	}, newInFlightTracker(), nil)
	if fallback.CheckpointStore != fromDomain {
		t.Error("nil EventConfig.CheckpointStore must fall back to Domain.CheckpointStore")
	}

	override := mergeDomain(EventConfig{
		CheckpointStore: fromConfig,
		Domain:          &system.DomainConfig{CheckpointStore: fromDomain},
	}, newInFlightTracker(), nil)
	if override.CheckpointStore != fromConfig {
		t.Error("EventConfig.CheckpointStore must win over Domain.CheckpointStore")
	}
}

func TestMergeDomain_Passthrough_Verbatim(t *testing.T) {
	t.Parallel()

	commands := func(*system.System) {}
	queries := func(*system.System) {}
	timers := func(*system.System) {}
	decoder := func(_ string, _ []byte) (any, error) { return nil, nil }

	consumer := &system.DomainConfig{
		Commands:                   commands,
		Queries:                    queries,
		Timers:                     timers,
		Events:                     []event.Type{"task.created", "task.completed"},
		DisableCoeffectValidation:  true,
		ProjectionDecoder:           decoder,
		ShutdownDependencies:        []system.ShutdownDependency{{Before: "a", After: "b"}},
		Evolutions: []system.EvolutionSpec{
			system.Evolve[TaskView]("task").On("task.created", TaskCreated{}).Done(),
		},
	}

	merged := mergeDomain(EventConfig{Domain: consumer}, newInFlightTracker(), nil)

	if reflect.ValueOf(merged.Commands).Pointer() != reflect.ValueOf(commands).Pointer() ||
		reflect.ValueOf(merged.Queries).Pointer() != reflect.ValueOf(queries).Pointer() ||
		reflect.ValueOf(merged.Timers).Pointer() != reflect.ValueOf(timers).Pointer() {
		t.Error("Commands/Queries/Timers funcs must pass through verbatim")
	}

	if reflect.ValueOf(merged.ProjectionDecoder).Pointer() != reflect.ValueOf(decoder).Pointer() {
		t.Error("ProjectionDecoder must pass through verbatim")
	}

	if !slices.Equal(merged.Events, consumer.Events) {
		t.Errorf("Events must pass through verbatim, got %v", merged.Events)
	}

	if !merged.DisableCoeffectValidation {
		t.Error("DisableCoeffectValidation must pass through verbatim")
	}

	if !slices.EqualFunc(merged.ShutdownDependencies, consumer.ShutdownDependencies,
		func(a, b system.ShutdownDependency) bool { return a == b }) {
		t.Errorf("ShutdownDependencies must pass through verbatim, got %v", merged.ShutdownDependencies)
	}

	if len(merged.Evolutions) != 1 || merged.Evolutions[0] != consumer.Evolutions[0] {
		t.Error("Evolutions must pass through verbatim")
	}
}

// ── Behavioral surface (live EventService) ──

// TaskView is the declared read model for the domain tests.
type TaskView struct {
	ID       string
	Title    string
	Status   string
	Priority int
}

// TaskCreated follows the Created/Updated/Deleted sample-name convention.
type TaskCreated struct {
	ID       string
	Title    string
	Status   string
	Priority int
}

// TaskCompleted has no convention suffix: it needs an explicit fold.
type TaskCompleted struct {
	ID string
}

// newDomainService builds a memory-backed service whose Domain declares a
// filterable, sortable TaskView QuerySet.
func newDomainService(t *testing.T) *EventService {
	t.Helper()

	return newDomainServiceCfg(t, func(domain *system.DomainConfig) {
		domain.Projections = []system.ProjectionDeclaration{
			system.QuerySet[TaskView]("tasks").
				On("task.created", TaskCreated{}).
				Filterable("status").
				Sortable("priority", false).
				Done(),
		}
	})
}

// newDomainServiceCfg lets a test customize the DomainConfig before
// construction; Driver and Domain are always set.
func newDomainServiceCfg(t *testing.T, customize func(*system.DomainConfig)) *EventService {
	t.Helper()

	domain := &system.DomainConfig{}
	customize(domain)

	eventSvc, err := NewEventService(EventConfig{
		Driver: memoryDriver,
		Domain: domain,
	})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	t.Cleanup(func() { _ = eventSvc.Shutdown(context.Background()) })

	return eventSvc
}

// appendDomainEvent saves payload as the given version on the entity's
// stream in the journal.
func appendDomainEvent(
	t *testing.T,
	eventSvc *EventService,
	eventType event.Type,
	entity string,
	version event.Version,
	payload any,
) {
	t.Helper()

	streamID, err := id.ParseStreamID(entity)
	if err != nil {
		t.Fatalf("parse stream id %q: %v", entity, err)
	}

	evt, err := event.New(eventType, streamID, "tasks", version, payload)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	ref := id.NewStreamRef("tasks", streamID)

	err = eventSvc.System().EventStore().Save(context.Background(), ref, []event.Event{evt}, version-1)
	if err != nil {
		t.Fatalf("save event: %v", err)
	}
}

// waitFindTasks polls the declared QuerySet until it holds want rows.
func waitFindTasks(t *testing.T, eventSvc *EventService, want int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		rows, err := system.Find[TaskView](context.Background(), eventSvc.System(), "tasks")
		if err == nil && len(rows) == want {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %d rows in the tasks QuerySet", want)
}

func TestEventService_DomainQuerySet_FindRoundTrip(t *testing.T) {
	t.Parallel()

	eventSvc := newDomainService(t)

	appendDomainEvent(t, eventSvc, "task.created", "task-1", 1,
		TaskCreated{ID: "task-1", Title: "Write domain tests", Status: "active", Priority: 2})
	appendDomainEvent(t, eventSvc, "task.created", "task-2", 1,
		TaskCreated{ID: "task-2", Title: "Ship v0.7.0", Status: "active", Priority: 5})
	appendDomainEvent(t, eventSvc, "task.created", "task-3", 1,
		TaskCreated{ID: "task-3", Title: "File the coeffect typo", Status: "done", Priority: 1})

	if err := eventSvc.StartProjections(context.Background()); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	waitFindTasks(t, eventSvc, 3)

	ctx := context.Background()
	sys := eventSvc.System()

	all, err := system.Find[TaskView](ctx, sys, "tasks")
	if err != nil {
		t.Fatalf("find all: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(all))
	}

	active, err := system.Find[TaskView](ctx, sys, "tasks", system.Where("status", "active"))
	if err != nil {
		t.Fatalf("find filtered: %v", err)
	}

	if len(active) != 2 {
		t.Fatalf("expected 2 active tasks, got %d", len(active))
	}

	for _, task := range active {
		if task.Status != "active" {
			t.Errorf("filter leaked a %q task", task.Status)
		}
	}

	byPriority, err := system.Find[TaskView](ctx, sys, "tasks",
		system.OrderBy("priority", system.Desc))
	if err != nil {
		t.Fatalf("find sorted: %v", err)
	}

	if len(byPriority) != 3 || byPriority[0].ID != "task-2" || byPriority[2].ID != "task-3" {
		t.Errorf("priority-desc sort wrong: %+v", byPriority)
	}

	top, err := system.Find[TaskView](ctx, sys, "tasks",
		system.Where("status", "active"),
		system.OrderBy("priority", system.Desc),
		system.Limit(1))
	if err != nil {
		t.Fatalf("find filtered+sorted+limited: %v", err)
	}

	if len(top) != 1 || top[0].ID != "task-2" {
		t.Errorf("combined options wrong: %+v", top)
	}
}

func TestNewEventService_DomainCoeffectGate_RejectsDanglingSubscription(t *testing.T) {
	t.Parallel()

	_, err := NewEventService(EventConfig{
		Driver: memoryDriver,
		Domain: &system.DomainConfig{
			Events: []event.Type{"task.created"},
			Projections: []system.ProjectionDeclaration{
				system.QuerySet[TaskView]("tasks").
					On("task.created", TaskCreated{}).
					On("task.creted", TaskCreated{}). // the typo the gate exists for
					Done(),
			},
		},
	})
	if err == nil {
		t.Fatal("expected construction to fail on the undeclared event type")
	}

	if !errors.Is(err, system.ErrDanglingEventSubscription) {
		t.Errorf("expected ErrDanglingEventSubscription in the chain, got: %v", err)
	}
}

func TestEventService_DomainBootstrapSkipped_WhenConsumerDeclaresProjections(t *testing.T) {
	t.Parallel()

	withDomain := newDomainService(t)

	names := planQueryNames(t, withDomain)
	if !slices.Contains(names, "tasks") {
		t.Errorf("consumer projection missing from plan: %v", names)
	}

	if slices.Contains(names, hostBootstrapDeclaration) {
		t.Errorf("bootstrap must be skipped when the consumer declares projections: %v", names)
	}

	withoutDomain, err := NewEventService(EventConfig{Driver: memoryDriver})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	t.Cleanup(func() { _ = withoutDomain.Shutdown(context.Background()) })

	if !slices.Contains(planQueryNames(t, withoutDomain), hostBootstrapDeclaration) {
		t.Error("nil Domain must keep the bootstrap declaration (host guarantee)")
	}
}

// planQueryNames collects the projection plan's query names.
func planQueryNames(t *testing.T, eventSvc *EventService) []string {
	t.Helper()

	plan := eventSvc.System().ProjectionPlan()
	if plan == nil {
		t.Fatal("nil projection plan")
	}

	names := make([]string, 0, len(plan.Queries))
	for _, q := range plan.Queries {
		names = append(names, q.Name)
	}

	return names
}

func TestEventService_DomainEvolveLookup_GetPointRead(t *testing.T) {
	t.Parallel()

	eventSvc := newDomainServiceCfg(t, func(domain *system.DomainConfig) {
		domain.Evolutions = []system.EvolutionSpec{
			system.OnEvolution(
				system.Evolve[TaskView]("task").On("task.created", TaskCreated{}),
				"task.completed", TaskCompleted{},
				func(_ TaskCompleted, v *TaskView) { v.Status = "done" },
			).Done(),
		}
		domain.Projections = []system.ProjectionDeclaration{
			system.Lookup[TaskView]("get-task").Done(), // inherits the evolution folds
		}
	})

	appendDomainEvent(t, eventSvc, "task.created", "task-1", 1,
		TaskCreated{ID: "task-1", Title: "Inherit folds", Status: "active", Priority: 1})

	if err := eventSvc.StartProjections(context.Background()); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	ctx := context.Background()
	sys := eventSvc.System()

	waitFor(t, "created task visible", func() bool {
		task, err := system.Get[TaskView](ctx, sys, "get-task", "task-1")
		return err == nil && task.Title == "Inherit folds"
	})

	appendDomainEvent(t, eventSvc, "task.completed", "task-1", 2, TaskCompleted{ID: "task-1"})

	waitFor(t, "explicit fold applied", func() bool {
		task, err := system.Get[TaskView](ctx, sys, "get-task", "task-1")
		return err == nil && task.Status == "done"
	})

	if _, err := system.Get[TaskView](ctx, sys, "get-task", "missing"); !errors.Is(err, system.ErrNotFound) {
		t.Errorf("expected ErrNotFound for a missing key, got: %v", err)
	}
}
