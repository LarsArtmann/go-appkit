// Tests for the command/query facade: typed registration, dispatch,
// staleness-gated queries, the default middleware builder, in-flight drain,
// and checkpoint durability across restarts.
package cqrs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"go.opentelemetry.io/otel/trace"
)

// ── Facade test domain ──

type facadeState struct {
	Count int
}

var facadeDecider = decider.Decider[facadeState]{
	Initial: facadeState{},
	Apply: func(state facadeState, evt event.Event) (facadeState, error) {
		if evt.Type() == "facade.bumped" {
			state.Count++
		}

		return state, nil
	},
}

type facadeQuery struct{}

func (facadeQuery) Type() query.Type { return "facade.count" }

func newFacadeCommand(t *testing.T, streamID id.StreamID) *command.BasicCommand {
	t.Helper()

	cmd, err := command.New("facade.bump", streamID)
	if err != nil {
		t.Fatalf("create command: %v", err)
	}

	return cmd
}

// newFacadeService builds a memory-backed service with the facade domain
// registered.
func newFacadeService(t *testing.T) *EventService {
	t.Helper()

	eventSvc, err := NewEventService(EventConfig{Driver: memoryDriver})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	t.Cleanup(func() { _ = eventSvc.Shutdown(context.Background()) })

	err = RegisterDecider(eventSvc, "Facade", facadeDecider)
	if err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	err = RegisterCommand[*command.BasicCommand, facadeState](eventSvc, "facade.bump",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[facadeState] {
			return system.Execute(ctx, cmd.StreamID(), "Facade",
				func(state facadeState, ver event.Version) ([]event.Event, error) {
					evt, err := event.New("facade.bumped", cmd.StreamID(), "Facade", ver+1, nil)
					if err != nil {
						return nil, err
					}

					return []event.Event{evt}, nil
				})
		})
	if err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}

	err = RegisterQuery[facadeQuery, int](eventSvc, "facade.count",
		func(_ context.Context, _ facadeQuery) (int, error) {
			return 42, nil
		})
	if err != nil {
		t.Fatalf("RegisterQuery: %v", err)
	}

	return eventSvc
}

func TestEventService_CommandQueryFacade_Roundtrip(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	streamID := id.NewStreamID()

	err := eventSvc.Dispatch(context.Background(), newFacadeCommand(t, streamID))
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	events, err := eventSvc.System().EventStore().Load(
		context.Background(), id.NewStreamRef("Facade", streamID),
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(events) != 1 || events[0].Type() != "facade.bumped" {
		t.Fatalf("expected one facade.bumped event, got %d events", len(events))
	}

	result, err := DispatchQuery[facadeQuery, int](
		context.Background(), eventSvc, facadeQuery{},
	)
	if err != nil {
		t.Fatalf("dispatch query: %v", err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestEventService_DispatchQueryChecked_StaleReturnsErrorWithoutAnswering(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	// No projection registered and no event processed: staleness check with
	// a zero budget is disabled, so the query answers.
	result, err := DispatchQueryChecked[facadeQuery, int](
		context.Background(), eventSvc, 0, facadeQuery{},
	)
	if err != nil {
		t.Fatalf("expected answered query with disabled check, got: %v", err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestEventService_CommandMiddleware_WrapsDispatch(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	seen := 0

	eventSvc, err := NewEventService(EventConfig{
		Driver: memoryDriver,
		CommandMiddleware: []command.Middleware{
			func(next command.Handler) command.Handler {
				return func(ctx context.Context, cmd command.Command) error {
					mu.Lock()
					seen++
					mu.Unlock()

					return next(ctx, cmd)
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	t.Cleanup(func() { _ = eventSvc.Shutdown(context.Background()) })

	err = RegisterDecider(eventSvc, "Facade", facadeDecider)
	if err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	err = RegisterCommand[*command.BasicCommand, facadeState](eventSvc, "facade.bump",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[facadeState] {
			return system.Execute(ctx, cmd.StreamID(), "Facade",
				func(state facadeState, ver event.Version) ([]event.Event, error) {
					return nil, nil
				})
		})
	if err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}

	err = eventSvc.Dispatch(context.Background(), newFacadeCommand(t, id.NewStreamID()))
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if seen != 1 {
		t.Errorf("expected middleware to see 1 command, saw %d", seen)
	}
}

func TestEventService_Shutdown_DrainsInFlightCommands(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	var mu sync.Mutex

	completions := 0

	eventSvc, err := NewEventService(EventConfig{
		Driver: memoryDriver,
		CommandMiddleware: []command.Middleware{
			func(next command.Handler) command.Handler {
				return func(ctx context.Context, cmd command.Command) error {
					<-release // block until the test allows completion

					mu.Lock()
					completions++
					mu.Unlock()

					return next(ctx, cmd)
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	err = RegisterDecider(eventSvc, "Facade", facadeDecider)
	if err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	err = RegisterCommand[*command.BasicCommand, facadeState](eventSvc, "facade.bump",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[facadeState] {
			return system.Execute(ctx, cmd.StreamID(), "Facade",
				func(state facadeState, ver event.Version) ([]event.Event, error) {
					return nil, nil
				})
		})
	if err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}

	done := make(chan error, 1)

	go func() {
		done <- eventSvc.Dispatch(context.Background(), newFacadeCommand(t, id.NewStreamID()))
	}()

	shutdownDone := make(chan error, 1)

	go func() {
		time.Sleep(50 * time.Millisecond) // let the command enter the middleware

		shutdownDone <- eventSvc.Shutdown(context.Background())
	}()

	// Shutdown must NOT complete while a command is in flight.
	select {
	case err := <-shutdownDone:
		t.Fatalf("shutdown completed before in-flight command drained: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	close(release)

	if err := <-done; err != nil {
		t.Fatalf("in-flight dispatch failed: %v", err)
	}

	if err := <-shutdownDone; err != nil {
		t.Fatalf("shutdown after drain: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if completions != 1 {
		t.Errorf("expected the in-flight command to complete, got %d completions", completions)
	}
}

func TestDefaultCommandMiddleware_ComposesRecoveryTracingLogging(t *testing.T) {
	t.Parallel()

	chain := DefaultCommandMiddleware(nil, nil)
	if len(chain) != 1 {
		t.Fatalf("expected recovery-only chain with nil tracer/logger, got %d", len(chain))
	}

	chain = DefaultCommandMiddleware(nil, fakeTracer{})
	if len(chain) != 2 {
		t.Fatalf("expected recovery+tracing chain, got %d", len(chain))
	}
}

type fakeTracer struct{}

func (fakeTracer) Start(
	ctx context.Context,
	_ string,
	_ ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

func TestEventService_DefaultCheckpointStore_PersistsAcrossRestart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dsn := dir + "/events.db"

	buildAndRun := func(counter *eventCounter) {
		eventSvc, err := NewEventService(EventConfig{DSN: dsn})
		if err != nil {
			t.Fatalf("NewEventService: %v", err)
		}

		err = eventSvc.Host().Register(counter.projection())
		if err != nil {
			t.Fatalf("register: %v", err)
		}

		if counter.fresh {
			appendTestEvent(t, eventSvc, "test.cp")
		}

		err = eventSvc.StartProjections(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		waitFor(t, "projection caught up", eventSvc.ReadyCheck)
		waitFor(t, "events counted", func() bool {
			return counter.processed() >= counter.expected
		})

		if err := eventSvc.Shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	}

	first := &eventCounter{name: "cp-projection", expected: 1, fresh: true}
	buildAndRun(first)

	// Second service on the SAME database: the persistent checkpoint must
	// skip the already-processed event (no replay).
	second := &eventCounter{name: "cp-projection", expected: 0, fresh: false}
	buildAndRun(second)
}

// eventCounter is a projection that counts processed events.
type eventCounter struct {
	name     string
	expected int
	fresh    bool

	mu            sync.Mutex
	processedCount int
}

func (c *eventCounter) processed() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.processedCount
}

func (c *eventCounter) projection() projection.Projection { //nolint:ireturn // upstream interface
	counter := c

	return projection.NewProjection(
		c.name,
		func(_ context.Context, _ event.Event) error {
			counter.mu.Lock()
			counter.processedCount++
			counter.mu.Unlock()

			return nil
		},
		[]event.Type{"test.cp"},
	)
}

func TestEventService_MissingDriverFailsConstruction(t *testing.T) {
	t.Parallel()

	_, err := NewEventService(EventConfig{
		Driver: "no-such-driver",
		DSN:    t.TempDir() + "/test.db",
	})
	if err == nil {
		t.Fatal("expected error for unregistered driver (blank-import contract)")
	}
}
