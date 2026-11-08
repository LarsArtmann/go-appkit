package integration_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-sse/ssetest"
)

func TestZZDebugProbe3(t *testing.T) {
	store := newMarkerStore()
	folded := make(chan cqrs.FoldedEvent, 64)

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		Driver: "memory",
		Domain: &system.DomainConfig{},
	})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

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

	if err = es.Host().Register(cqrs.NotifyFolded(&markerProjection{store: store}, folded)); err != nil {
		t.Fatalf("register: %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()

	if err = es.StartProjections(startCtx); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	hub := realtime.NewHub()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = testkit.FreeAddr(t)
	cfg.DrainDelay = appkit.NoDrainDelay
	cfg.ReadTimeout = appkit.NoTimeout
	cfg.WriteTimeout = appkit.NoTimeout
	cfg.DrainHooks = []appkit.Hook{hub.Shutdown}
	cfg.ShutdownHooks = []appkit.Hook{es.Shutdown}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	realtime.Mount(svc.Mux, "/sse", hub)

	ts := testkit.Serve(t, svc)
	base := ts.FullChainURL

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, base+"/sse", nil)
	stream, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}

	defer func() { _ = stream.Body.Close() }() //respdrain:ignore SSE stream — Close IS the disconnect

	for range 50 {
		if hub.SubscriberCount() == 1 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Logf("subscribers: %d", hub.SubscriberCount())

	var snapshotCalls atomic.Int64

	srcCtx, srcCancel := context.WithCancel(context.Background())
	defer srcCancel()

	realtime.FoldedSource(srcCtx, folded, func() any {
		snapshotCalls.Add(1)

		return store.snapshot()
	}, hub)

	cmd, _ := command.New("folded.add_thing", id.NewStreamID())
	if err = es.Dispatch(context.Background(), cmd); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	t.Log("direct dispatch done")

	events := ssetest.MustReadNEvents(t, stream.Body, 1)
	t.Logf("SSE event: type=%s id=%s data=%s (snapshotCalls=%d)",
		events[0].Type, events[0].ID, events[0].Data(), snapshotCalls.Load())
}
