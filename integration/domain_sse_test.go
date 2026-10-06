package integration_test

// E13-stretch — Domain-declared read models streaming through a live appkit
// Service, against PUBLISHED tags (cqrs v0.7.0 + core v0.7.0 + health v0.1.5):
// the composition a consumer builds for a live task board —
//
//	NewEventService(memory, Domain{QuerySet}) → RegisterDecider/Command
//	  → StartProjections → appkit ServiceConfig{ReadyCheck, ShutdownHooks}
//	  → GET /events/tasks (metaengine.ServeSSE over a NewWatcher)
//	  → POST /tasks dispatches the command that materializes the streamed row
//
// Also compiles and exercises the cqrs README's go-health bridge recipe:
// appkithealth.NewProbe over EventService.HealthCheck + CheckStaleness, and
// the LagPerProjection ops loop.
//
// Proven here:
//  1. A command dispatched over HTTP lands on the SSE stream as a
//     materialized read-model change — through the FULL appkit chain
//     (default middleware stack included), not a bare httptest server.
//  2. The health-bridge recipe from the cqrs README compiles against the
//     published tags and its checks pass on a live Domain service.

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/cqrs"
	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// ── Domain SSE E2E board (minimal, mirrors cqrs' own domain fixtures) ──

type boardState struct{ Tasks int }

var boardDecider = decider.Decider[boardState]{
	Initial: boardState{},
	Apply: func(state boardState, evt event.Event) (boardState, error) {
		if evt.Type() == "board.task_added" {
			state.Tasks++
		}

		return state, nil
	},
}

type boardTaskAdded struct {
	ID     string
	Title  string
	Status string
}

type boardTaskView struct {
	ID     string
	Title  string
	Status string
}

func dispatchAddTask(es *cqrs.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := command.New("board.add_task", id.NewStreamID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		dispErr := es.Dispatch(r.Context(), cmd)
		if dispErr != nil {
			http.Error(w, dispErr.Error(), http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func TestCQRSDomainSSEThroughAppkitService(t *testing.T) {
	t.Parallel()

	es, err := cqrs.NewEventService(cqrs.EventConfig{
		Driver: "memory",
		Domain: &system.DomainConfig{
			Projections: []system.ProjectionDeclaration{
				system.QuerySet[boardTaskView]("board-tasks").
					On("board.task_added", boardTaskAdded{}).
					Done(),
			},
		},
	})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	err = cqrs.RegisterDecider(es, "Board", boardDecider)
	if err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	err = cqrs.RegisterCommand[*command.BasicCommand, boardState](es, "board.add_task",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[boardState] {
			return system.Execute(ctx, cmd.StreamID(), "Board",
				func(_ boardState, ver event.Version) ([]event.Event, error) {
					evt, evtErr := event.New("board.task_added", cmd.StreamID(), "Board", ver+1,
						boardTaskAdded{
							ID:     cmd.StreamID().String(),
							Title:  "E2E Domain task",
							Status: "active",
						})
					if evtErr != nil {
						return nil, evtErr //nolint:wrapcheck // test boundary
					}

					return []event.Event{evt}, nil
				})
		})
	if err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()

	err = es.StartProjections(startCtx)
	if err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	// The cqrs README's go-health bridge recipe, compiled against the
	// published tags and called on a live Domain service.
	probe := appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
		"cqrs-engines": es.HealthCheck,
		"cqrs-lag": func(_ context.Context) error {
			return es.CheckStaleness(2 * time.Second)
		},
	})
	_ = probe // wired into a Mounted probe/dashboard in a real service

	err = es.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck on a live Domain service: %v", err)
	}

	err = es.CheckStaleness(2 * time.Second)
	if err != nil {
		t.Fatalf("CheckStaleness on a caught-up Domain service: %v", err)
	}

	if len(es.LagPerProjection()) == 0 {
		t.Fatal("LagPerProjection returned no projections for a declared QuerySet")
	}

	watcher := metaengine.NewWatcher[boardTaskView](es.System().MetaEngine(), "board-tasks")

	defer watcher.Close()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.DrainDelay = 1 * time.Millisecond
	cfg.ReadyCheck = es.ReadyCheck
	cfg.ShutdownHooks = []func(context.Context) error{es.Shutdown}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("GET /events/tasks", func(w http.ResponseWriter, r *http.Request) {
		_ = metaengine.ServeSSE(w, r, watcher) // stream ends with the request
	})
	svc.Mux.HandleFunc("POST /tasks", dispatchAddTask(es))

	ts := testkit.Serve(t, svc)
	baseURL := ts.FullChainURL

	waitReady(t, baseURL)

	// Subscribe through the FULL chain first, then dispatch: the streamed
	// row proves command -> journal -> projection -> watcher -> SSE through
	// the appkit default middleware stack.
	sseCtx, sseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer sseCancel()

	req, err := http.NewRequestWithContext(sseCtx, http.MethodGet, baseURL+"/events/tasks", nil)
	if err != nil {
		t.Fatalf("build SSE request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect to the SSE stream through the full chain: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("SSE stream status = %d, want 200", resp.StatusCode)
	}

	if status := postStatus(t, baseURL+"/tasks"); status != http.StatusNoContent {
		t.Fatalf("POST /tasks status = %d, want 204", status)
	}

	scanner := bufio.NewScanner(resp.Body)

	found := false

	for scanner.Scan() {
		if data, ok := strings.CutPrefix(scanner.Text(), "data:"); ok {
			if strings.Contains(data, "E2E Domain task") {
				found = true

				break
			}
		}
	}

	if !found {
		t.Fatal("the dispatched task never materialized on the SSE stream through the full chain")
	}
}
