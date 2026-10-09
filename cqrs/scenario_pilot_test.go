package cqrs

// systemscenario pilot (2026-10-09): the go-cqrs-lite system-level BDD
// harness (ADR-0153) driving THIS wrapper's booted system. The facade owns
// the system.New call (NewEventService), so the pilot uses
// systemscenario.Adopt over EventService.System() — Given/When/Then against
// the exact surface consumers use:
// NewEventService → RegisterDecider/RegisterCommand/RegisterQuery →
// Dispatch → DispatchQuery → Shutdown.

import (
	"context"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/systemscenario/v4"
)

func TestScenarioPilot_FacadeRoundtrip(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	sc := systemscenario.Adopt(t, context.Background(), eventSvc.System())

	streamID := id.NewStreamID()

	sc.Given().Command(newFacadeCommand(t, streamID)).
		When(newFacadeCommand(t, streamID)).
		Then("facade.bumped").
		ThenCommands("facade.bump")
}

func TestScenarioPilot_FacadeQueryAct(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	sc := systemscenario.Adopt(t, context.Background(), eventSvc.System())

	sc.Given().Command(newFacadeCommand(t, id.NewStreamID())).
		WhenQuery(facadeQuery{}).
		ThenSuccess().
		ThenResult(42)
}
