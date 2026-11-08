package integration_test

// W5-C2 discriminator E2E (design note §4, `docs/planning/
// 2026-11-07_w5-c2-folded-contract-design-note.md`) — the folded contract
// through PUBLISHED tags (cqrs + realtime + core), full chain:
//
//	POST /things → command → journal → projection fold (writes a MID-FOLD
//	  marker into the snapshot store) → cqrs.NotifyFolded signal →
//	  realtime.FoldedSource snapshot → hub broadcast → SSE subscriber.
//
// The discrimination: the broadcast's snapshot payload must CONTAIN the
// marker written by the fold of exactly the trigger event — a pre-fold
// (notify-on-append) broadcast cannot contain it, because the marker is
// written inside Handle and the event's ID only reaches the snapshot when
// Handle runs. The negative control pins that the assertion discriminates:
// a naive signal-before-fold wiring broadcasts a payload that LACKS the
// marker for the same event id, deterministically (the control's Handle
// blocks until the stale broadcast has been observed).
//
// Also proves the structural decoupling: this file is the only place cqrs
// and realtime meet — neither module imports the other.

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-sse/ssetest"
)

// ── shared domain fixtures (mirror domain_sse_test.go's board) ──

type foldedThingState struct{}

var foldedDecider = decider.Decider[foldedThingState]{
	Initial: foldedThingState{},
	Apply: func(state foldedThingState, evt event.Event) (foldedThingState, error) {
		return state, nil
	},
}

// markerStore is the read model FoldedSource snapshots. The projection
// writes one marker per folded event id; the payload IS the assertion
// surface.
type markerStore struct {
	mu      sync.Mutex
	markers map[string]string
}

func newMarkerStore() *markerStore {
	return &markerStore{markers: map[string]string{}}
}

func (s *markerStore) set(eventID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.markers[eventID] = "folded:" + eventID
}

func (s *markerStore) snapshot() any {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]string, len(s.markers))
	for id, marker := range s.markers {
		out[id] = marker
	}

	return out
}

// markerProjection folds one event type into the marker store. gate, when
// non-nil, blocks Handle until closed — the negative control's determinism
// device (the fold is provably still running when the stale broadcast
// lands).
type markerProjection struct {
	store *markerStore
	gate  <-chan struct{}
}

func (p *markerProjection) Name() string { return "folded-markers" }

func (p *markerProjection) EventTypes() []event.Type {
	return []event.Type{"folded.thing_added"}
}

func (p *markerProjection) Handle(_ context.Context, evt event.Event) error {
	if p.gate != nil {
		<-p.gate
	}

	p.store.set(evt.ID().String())

	return nil
}

func dispatchAddThing(es *cqrs.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := command.New("folded.add_thing", id.NewStreamID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if dispErr := es.Dispatch(r.Context(), cmd); dispErr != nil {
			http.Error(w, dispErr.Error(), http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// foldedHarness is one live discriminator scenario: engine with the given
// (already wrapped) projection registered raw on the host, realtime hub on
// a full-chain service, and one subscribed SSE stream.
type foldedHarness struct {
	stream *http.Response
	base   string
	hub    *realtime.Hub
}

func newFoldedHarness(t *testing.T, proj projection.Projection) *foldedHarness {
	t.Helper()

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

	if err = es.Host().Register(proj); err != nil {
		t.Fatalf("host.Register(marker projection): %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()

	if err = es.StartProjections(startCtx); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	hub := realtime.NewHub()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = testkit.FreeAddr(t)
	cfg.DrainDelay = 1 * time.Millisecond
	cfg.ReadTimeout = appkit.NoTimeout
	cfg.WriteTimeout = appkit.NoTimeout
	cfg.DrainHooks = []appkit.Hook{hub.Shutdown}
	cfg.ShutdownHooks = []appkit.Hook{es.Shutdown}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	realtime.Mount(svc.Mux, "/sse", hub)
	svc.Mux.HandleFunc("POST /things", dispatchAddThing(es))

	ts := testkit.Serve(t, svc)
	base := ts.FullChainURL

	// The stream's lifetime is the TEST's, not this helper's: no deferred
	// cancel here (it would fire when the harness returns and kill the
	// stream mid-read). The client timeout below bounds a stuck stream.
	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, base+"/sse", nil)
	if reqErr != nil {
		t.Fatalf("build SSE request: %v", reqErr)
	}

	stream, doErr := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if doErr != nil {
		t.Fatalf("open SSE stream through the full chain: %v", doErr)
	}

	if ct := stream.Header.Get("Content-Type"); ct != "text/event-stream" {
		_ = stream.Body.Close()
		t.Fatalf("stream content-type = %q, want text/event-stream", ct)
	}

	// Broadcasts (from FoldedSource, not this test) only land after the
	// handler's subscription is registered; SubscriberCount is the sync
	// point (integration house rule).
	subscribed := false

	for range 50 {
		if hub.SubscriberCount() == 1 {
			subscribed = true

			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !subscribed {
		_ = stream.Body.Close()
		t.Fatal("SSE handler never subscribed to the hub")
	}

	return &foldedHarness{stream: stream, base: base, hub: hub}
}

// markerFor builds the exact string a correct post-fold snapshot must carry
// for the given SSE event id — and a pre-fold snapshot cannot.
func markerFor(eventID string) string { return `"folded:` + eventID + `"` }

// TestFoldedContract_BroadcastReflectsTheFold is the positive discriminator:
// cqrs.NotifyFolded + realtime.FoldedSource through published tags — the
// broadcast payload contains the marker of exactly the trigger event.
func TestFoldedContract_BroadcastReflectsTheFold(t *testing.T) {
	t.Parallel()

	store := newMarkerStore()
	folded := make(chan cqrs.FoldedEvent, 64)

	h := newFoldedHarness(t, cqrs.NotifyFolded(&markerProjection{store: store}, folded))
	defer func() { _ = h.stream.Body.Close() }() //respdrain:ignore SSE stream — Close IS the disconnect

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	realtime.FoldedSource(ctx, folded, store.snapshot, h.hub)

	if status := postStatus(t, h.base+"/things"); status != http.StatusNoContent {
		t.Fatalf("POST /things status = %d, want 204", status)
	}

	events := ssetest.MustReadNEvents(t, h.stream.Body, 1)
	evt := events[0]

	if evt.Type != "folded.thing_added" {
		t.Fatalf("SSE event type = %q, want folded.thing_added", evt.Type)
	}

	if evt.ID == "" {
		t.Fatal("SSE event has no id — the folded signal's EventID must ride the broadcast")
	}

	if !strings.Contains(evt.Data(), markerFor(evt.ID)) {
		t.Fatalf(
			"broadcast payload %q lacks the marker %q — the snapshot was taken BEFORE the fold completed",
			evt.Data(), markerFor(evt.ID),
		)
	}
}

// naiveSignal is the wrong wiring's announcement: a locally-built
// realtime.FoldedSignal sent BEFORE the fold runs (the notify-on-append
// shape the contract forbids).
type naiveSignal struct {
	name string
	id   string
}

func (s naiveSignal) EventName() string { return s.name }
func (s naiveSignal) EventID() string   { return s.id }

// naivePreFoldProjection emits the signal before folding — and its inner
// fold blocks on gate until the test has observed the broadcast, making the
// staleness deterministic rather than a race.
type naivePreFoldProjection struct {
	inner *markerProjection
	out   chan<- naiveSignal
}

func (p *naivePreFoldProjection) Name() string { return p.inner.Name() }

func (p *naivePreFoldProjection) EventTypes() []event.Type { return p.inner.EventTypes() }

func (p *naivePreFoldProjection) Handle(ctx context.Context, evt event.Event) error {
	p.out <- naiveSignal{name: string(evt.Type()), id: evt.ID().String()}

	return p.inner.Handle(ctx, evt)
}

// TestFoldedContract_PreFoldWiringFailsTheSameAssertion is the negative
// control: the SAME assertion shape must FAIL for a naive pre-fold wiring,
// or the discriminator discriminates nothing.
func TestFoldedContract_PreFoldWiringFailsTheSameAssertion(t *testing.T) {
	t.Parallel()

	store := newMarkerStore()

	gate := make(chan struct{})
	release := sync.Once{}
	t.Cleanup(func() { release.Do(func() { close(gate) }) }) // never strand the host worker

	naive := make(chan naiveSignal, 1)

	h := newFoldedHarness(t,
		&naivePreFoldProjection{inner: &markerProjection{store: store, gate: gate}, out: naive})
	defer func() { _ = h.stream.Body.Close() }() //respdrain:ignore SSE stream — Close IS the disconnect

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	realtime.FoldedSource(ctx, naive, store.snapshot, h.hub)

	if status := postStatus(t, h.base+"/things"); status != http.StatusNoContent {
		t.Fatalf("POST /things status = %d, want 204", status)
	}

	// The fold is still blocked on the gate: any broadcast arriving now is
	// provably pre-fold.
	events := ssetest.MustReadNEvents(t, h.stream.Body, 1)
	evt := events[0]

	if evt.Type != "folded.thing_added" || evt.ID == "" {
		t.Fatalf("SSE event = %q/%q, want folded.thing_added with an id", evt.Type, evt.ID)
	}

	if strings.Contains(evt.Data(), markerFor(evt.ID)) {
		t.Fatalf(
			"pre-fold broadcast payload %q already contains the marker %q — the control is not discriminating",
			evt.Data(), markerFor(evt.ID),
		)
	}

	// Release the fold; the marker lands (provably) after the stale
	// broadcast, which is exactly the client-visible staleness the real
	// contract exists to prevent.
	release.Do(func() { close(gate) })
}
