package integration_test

// T17 — cqrs lifecycle E2E through a live appkit Service, against PUBLISHED
// tags (cqrs v0.6.0 + core v0.6.0): the full composition a consumer builds —
//
//	NewEventService(memory) → RegisterDecider/Command/Query → StartProjections
//	  → appkit ServiceConfig{ReadyCheck: es.ReadyCheck, ShutdownHooks: es.Shutdown}
//	  → HTTP endpoints dispatching commands and queries through the REAL
//	    full-chain test harness (testkit.Serve).
//
// Proven here:
//  1. cqrs.ReadyCheck composes into /health/ready (200 after StartProjections).
//  2. A command dispatched over HTTP is visible to a query dispatched over
//     HTTP (command→store→query roundtrip through the wire).
//  3. In-flight drain: a command whose handler blocks still completes when
//     Shutdown starts mid-flight — the ONLY thing that unblocks the handler
//     is the DrainHook (i.e. the drain has begun), the HTTP response still
//     lands 204, and es.Shutdown (a ShutdownHook) only finishes after the
//     in-flight command's work is done.

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// ── Lifecycle E2E domain (minimal, mirrors cqrs' own facade fixtures) ──

type lifecycleState struct {
	Bumps int
}

var lifecycleDecider = decider.Decider[lifecycleState]{
	Initial: lifecycleState{},
	Apply: func(state lifecycleState, evt event.Event) (lifecycleState, error) {
		if evt.Type() == "lifecycle.bumped" {
			state.Bumps++
		}

		return state, nil
	},
}

type lifecycleCountQuery struct{}

func (lifecycleCountQuery) Type() query.Type { return "lifecycle.count" }

// dispatchBump builds the POST endpoint that dispatches one lifecycle.bump
// command against the shared stream. When slowRequested is non-nil the
// handler parks until release closes — and release is owned by the drain
// hook, so a parked command proves Shutdown waited for in-flight work.
func dispatchBump(
	es *cqrs.EventService,
	streamID id.StreamID,
	slowRequested *atomic.Bool,
	handlerEntered, release chan struct{},
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if slowRequested != nil {
			slowRequested.Store(true)
			close(handlerEntered)
			<-release
		}

		cmd, err := command.New("lifecycle.bump", streamID)
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

func countHandler(es *cqrs.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		n, err := cqrs.DispatchQuery[lifecycleCountQuery, int](r.Context(), es, lifecycleCountQuery{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"count": n})
	}
}

// waitReady polls /health/ready until it flips to 200 (projection worker live).
func waitReady(t *testing.T, baseURL string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := http.Get(baseURL + "/health/ready") //nolint:noctx // test boundary
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		if time.Now().After(deadline) {
			t.Fatal("/health/ready never became ready after StartProjections")
		}

		time.Sleep(20 * time.Millisecond)
	}
}

func postStatus(t *testing.T, url string) int {
	t.Helper()

	resp, err := http.Post(url, "application/json", nil) //nolint:noctx,gosec // test boundary
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func TestCQRSLifecycleThroughAppkitService(t *testing.T) {
	t.Parallel()

	var (
		testStreamID   = id.NewStreamID()
		slowRequested  atomic.Bool
		handlerEntered = make(chan struct{})
		release        = make(chan struct{})
	)

	es, err := cqrs.NewEventService(cqrs.EventConfig{Driver: "memory"})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	err = cqrs.RegisterDecider(es, "Lifecycle", lifecycleDecider)
	if err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	err = cqrs.RegisterCommand[*command.BasicCommand, lifecycleState](es, "lifecycle.bump",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[lifecycleState] {
			if slowRequested.Load() {
				<-release
			}

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

	err = cqrs.RegisterQuery[lifecycleCountQuery, int](es, "lifecycle.count",
		func(ctx context.Context, _ lifecycleCountQuery) (int, error) {
			events, loadErr := es.System().EventStore().Load(
				ctx, id.NewStreamRef("Lifecycle", testStreamID))
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
		})
	if err != nil {
		t.Fatalf("RegisterQuery: %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()
	err = es.StartProjections(startCtx)
	if err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.DrainDelay = 1 * time.Millisecond
	cfg.ReadyCheck = es.ReadyCheck
	// The drain hook is the ONLY releaser of the parked command: once it has
	// run, the ready probe is down and the drain window is open, so the
	// command below was provably in flight across the shutdown boundary.
	cfg.DrainHooks = []func(context.Context) error{
		func(context.Context) error {
			close(release)

			return nil
		},
	}
	cfg.ShutdownHooks = []func(context.Context) error{es.Shutdown}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("POST /bump", dispatchBump(es, testStreamID, nil, nil, nil))
	svc.Mux.HandleFunc("POST /slowbump", dispatchBump(es, testStreamID, &slowRequested, handlerEntered, release))
	svc.Mux.HandleFunc("GET /count", countHandler(es))

	ts := testkit.Serve(t, svc)
	baseURL := ts.FullChainURL

	// 1) ReadyCheck composition: /health/ready flips to 200 once the
	// projection worker is live.
	waitReady(t, baseURL)

	// 2) Command over HTTP → query over HTTP.
	if status := postStatus(t, baseURL+"/bump"); status != http.StatusNoContent {
		t.Fatalf("POST /bump status = %d, want 204", status)
	}

	countCtx, countCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer countCancel()
	_, _, countBody := getWithAuth(t, countCtx, baseURL, "/count", "", "")

	var got struct {
		Count int `json:"count"`
	}

	err = json.Unmarshal([]byte(countBody), &got)
	if err != nil {
		t.Fatalf("decode count %q: %v", countBody, err)
	}

	if got.Count != 1 {
		t.Fatalf("count after one bump = %d, want 1", got.Count)
	}

	// 3) In-flight drain: shutdown starts while a command is parked in its
	// handler; the command still completes and the response still lands.
	slowDone := make(chan int, 1)
	go func() {
		slowDone <- postStatus(t, baseURL+"/slowbump")
	}()

	select {
	case <-handlerEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("slow handler never entered")
	}

	shutdownErrCh := make(chan error, 1)
	go func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErrCh <- svc.Shutdown(shutdownCtx)
	}()

	select {
	case status := <-slowDone:
		if status != http.StatusNoContent {
			t.Errorf("POST /slowbump status = %d, want 204 (response must survive the drain)", status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("slow command never completed")
	}

	select {
	case err := <-shutdownErrCh:
		if err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown never returned after the in-flight command completed")
	}

	if svc.Running() {
		t.Error("service still Running after Shutdown returned")
	}
}
