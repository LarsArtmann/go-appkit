package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

func TestZZDebugProbe(t *testing.T) {
	folded := make(chan cqrs.FoldedEvent, 64)

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		Driver: "memory",
		Domain: &system.DomainConfig{},
	})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	t.Logf("host nil? %v", es.Host() == nil)

	if err = cqrs.RegisterDecider(es, "Folded", foldedDecider); err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	if err = cqrs.RegisterCommand[*command.BasicCommand, foldedThingState](es, "folded.add_thing",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[foldedThingState] {
			return system.Execute(ctx, cmd.StreamID(), "Folded",
				func(_ foldedThingState, ver event.Version) ([]event.Event, error) {
					evt, evtErr := event.New("folded.thing_added", cmd.StreamID(), "Folded", ver+1,
						struct{ ID string }{ID: cmd.StreamID().String()})
					if evtErr != nil {
						return nil, evtErr //nolint:wrapcheck // test boundary
					}

					return []event.Event{evt}, nil
				})
		}); err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}

	hits := make(chan string, 8)
	debugProj := projectionFuncs{
		name: "folded-markers",
		handle: func(ctx context.Context, evt event.Event) error {
			t.Logf("HANDLE CALLED: type=%s id=%s", evt.Type(), evt.ID().String())
			hits <- evt.ID().String()

			return nil
		},
	}

	if err = es.Host().Register(cqrs.NotifyFolded(&debugProj, folded)); err != nil {
		t.Fatalf("register: %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()

	if err = es.StartProjections(startCtx); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	cmd, _ := command.New("folded.add_thing", id.NewStreamID())
	if err = es.Dispatch(context.Background(), cmd); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	t.Log("dispatch returned")

	select {
	case id := <-hits:
		t.Logf("handle hit: %s", id)
	case <-time.After(3 * time.Second):
		t.Error("Handle never called within 3s")
	}

	select {
	case fe := <-folded:
		t.Logf("folded signal: %s/%s", fe.EventName(), fe.EventID())
	case <-time.After(1 * time.Second):
		t.Error("no folded signal within 1s")
	}
}

type projectionFuncs struct {
	name   string
	handle func(ctx context.Context, evt event.Event) error
}

func (p *projectionFuncs) Name() string { return p.name }

func (p *projectionFuncs) EventTypes() []event.Type {
	return []event.Type{"folded.thing_added"}
}

func (p *projectionFuncs) Handle(ctx context.Context, evt event.Event) error {
	return p.handle(ctx, evt)
}
