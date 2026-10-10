package integration_test

// T18 — layer-2 harness pilot: systemscenario.Adopt over the facade's
// booted system, with the ACTS driven over real HTTP (testkit.Serve full
// chain) and the ASSERTIONS from the harness. The layer-1 pilot
// (cqrs/scenario_pilot_test.go) proves Adopt over the facade API; this one
// proves the same scenario layer composes in front of an HTTP-fronted
// appkit Service: the act path is the wire (POST /bump through every
// middleware testkit installs), the assert path is journal diffing,
// command capture, and read-model queries.
//
// Baseline trick worth remembering: a WhenQuery act with no prior acts
// snapshots the journal/command baselines without side effects — acts that
// happen AFTER it over HTTP are diffable by the harness exactly like
// harness-dispatched acts (Adopt's capture middleware sits on the system
// dispatcher, which the HTTP handlers dispatch through).

import (
	"context"
	"net/http"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-cqrs-lite/systemscenario/v4"
)

// registerPilotBump registers the plain lifecycle.bump command (no
// slow-park — that variant belongs to the drain test in
// cqrs_lifecycle_test.go).
func registerPilotBump(t *testing.T, es *cqrs.EventService) {
	t.Helper()

	err := cqrs.RegisterCommand[*command.BasicCommand, lifecycleState](es, "lifecycle.bump",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[lifecycleState] {
			return system.Execute(ctx, cmd.StreamID(), "Lifecycle",
				func(_ lifecycleState, ver event.Version) ([]event.Event, error) {
					evt, evtErr := event.New("lifecycle.bumped", cmd.StreamID(), "Lifecycle", ver+1,
						struct{ N int }{N: 1})
					if evtErr != nil {
						return nil, evtErr //nolint:wrapcheck // test boundary
					}

					return []event.Event{evt}, nil
				})
		})
	if err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}
}

func TestHarnessPilot_HTTPActsHarnessAssertions(t *testing.T) {
	t.Parallel()

	streamID := id.NewStreamID()

	es, err := cqrs.NewEventService(cqrs.EventConfig{Driver: "memory"})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	if err := cqrs.RegisterDecider(es, "Lifecycle", lifecycleDecider); err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	registerPilotBump(t, es)

	if err := cqrs.RegisterQuery[lifecycleCountQuery, int](es, "lifecycle.count",
		func(ctx context.Context, _ lifecycleCountQuery) (int, error) {
			events, loadErr := es.System().EventStore().Load(
				ctx, id.NewStreamRef("Lifecycle", streamID))
			if loadErr != nil {
				return 0, loadErr //nolint:wrapcheck // test boundary
			}

			n := 0
			for _, evt := range events {
				if evt.Type() == "lifecycle.bumped" {
					n++
				}
			}

			return n, nil
		}); err != nil {
		t.Fatalf("RegisterQuery: %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()

	if err := es.StartProjections(startCtx); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.ReadyCheck = es.ReadyCheck

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("POST /bump", dispatchBump(es, streamID, nil, nil, nil))

	ts := testkit.Serve(t, svc)
	baseURL := ts.FullChainURL

	waitReady(t, baseURL)

	ctx := context.Background()
	sc := systemscenario.Adopt(t, ctx, es.System())
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = es.Shutdown(shutdownCtx)
	})

	// Baseline: a query act snapshots journal + command baselines with no
	// side effects, so wire-driven acts below are diffable.
	sc.WhenQuery(lifecycleCountQuery{}).ThenSuccess()

	// Acts over the real wire.
	if status := postStatus(t, baseURL+"/bump"); status != http.StatusNoContent {
		t.Fatalf("POST /bump status = %d, want 204", status)
	}

	if status := postStatus(t, baseURL+"/bump"); status != http.StatusNoContent {
		t.Fatalf("second POST /bump status = %d, want 204", status)
	}

	// Assertions from the harness: journal diff, captured wire commands,
	// read model — the same Then* surface as in-process scenarios.
	sc.Phase().
		Then("lifecycle.bumped", "lifecycle.bumped").
		ThenCommands("lifecycle.bump", "lifecycle.bump").
		ThenQuery(func() (any, error) {
			return cqrs.DispatchQuery[lifecycleCountQuery, int](ctx, es, lifecycleCountQuery{})
		}, 2)
}
