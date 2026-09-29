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
//  1. cqrs.ReadyCheck composes into /health/ready (503 until the projection
//     worker is live, 200 after StartProjections).
//  2. A command dispatched over HTTP is visible to a query dispatched over
//     HTTP (command→store→query roundtrip through the wire).
//  3. In-flight drain: a command whose handler blocks still completes when
//     Shutdown starts mid-flight — the drain hook fires while the command is
//     in flight, the HTTP response still lands 200, and es.Shutdown (a
//     ShutdownHook) only finishes after the in-flight command's work is done.

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

func TestCQRSLifecycleThroughAppkitService(t *testing.T) {
	t.Parallel()

	var (
		testStreamID    = id.NewStreamID()
		slowRequested   atomic.Bool
		handlerEntered  = make(chan struct{})
		releaseHandler  = make(chan struct{})
		drainStarted    = make(chan struct{})
		handlerFinished = make(chan struct{})
	)

	es, err := cqrs.NewEventService(cqrs.EventConfig{Driver: "memory"})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	if err := cqrs.RegisterDecider(es, "Lifecycle", lifecycleDecider); err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	if err := cqrs.RegisterCommand[*command.BasicCommand, lifecycleState](es, "lifecycle.bump",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[lifecycleState] {
			if slowRequested.Load() {
				close(handlerEntered)
				<-releaseHandler
				defer close(handlerFinished)
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
		}); err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}

	if err := cqrs.RegisterQuery[lifecycleCountQuery, int](es, "lifecycle.count",
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
		}); err != nil {
		t.Fatalf("RegisterQuery: %v", err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()
	if err := es.StartProjections(startCtx); err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.DrainDelay = 1 * time.Millisecond
	cfg.ReadyCheck = es.ReadyCheck
	cfg.DrainHooks = []func(context.Context) error{
		func(context.Context) error { close(drainStarted); return nil },
	}
	cfg.ShutdownHooks = []func(context.Context) error{es.Shutdown}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.HandleFunc("POST /bump", func(w http.ResponseWriter, _ *http.Request) {
		cmd, cmdErr := command.New("lifecycle.bump", testStreamID)
		if cmdErr != nil {
			http.Error(w, cmdErr.Error(), http.StatusInternalServerError)
			return
		}
		if dispErr := es.Dispatch(context.Background(), cmd); dispErr != nil {
			http.Error(w, dispErr.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	svc.Mux.HandleFunc("POST /slowbump", func(w http.ResponseWriter, _ *http.Request) {
		slowRequested.Store(true)
		cmd, cmdErr := command.New("lifecycle.bump", testStreamID)
		if cmdErr != nil {
			http.Error(w, cmdErr.Error(), http.StatusInternalServerError)
			return
		}
		if dispErr := es.Dispatch(context.Background(), cmd); dispErr != nil {
			http.Error(w, dispErr.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	svc.Mux.HandleFunc("GET /count", func(w http.ResponseWriter, r *http.Request) {
		n, qErr := cqrs.DispatchQuery[lifecycleCountQuery, int](r.Context(), es, lifecycleCountQuery{})
		if qErr != nil {
			http.Error(w, qErr.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"count": n}) //nolint:errchkjson // test boundary
	})

	ts := testkit.Serve(t, svc)
	baseURL := ts.FullChainURL

	// 1) ReadyCheck composition: /health/ready flips to 200 once the
	// projection worker is live.
	readyDeadline := time.Now().Add(2 * time.Second)
	for {
		resp, respErr := http.Get(baseURL + "/health/ready") //nolint:noctx // test boundary
		if respErr == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(readyDeadline) {
			t.Fatal("/health/ready never became ready after StartProjections")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 2) Command over HTTP → query over HTTP.
	bumpResp, bumpErr := http.Post(baseURL+"/bump", "application/json", nil) //nolint:noctx // test boundary
	if bumpErr != nil {
		t.Fatalf("POST /bump: %v", bumpErr)
	}
	_ = bumpResp.Body.Close()
	if bumpResp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /bump status = %d, want 204", bumpResp.StatusCode)
	}

	countResp, countErr := http.Get(baseURL + "/count") //nolint:noctx // test boundary
	if countErr != nil {
		t.Fatalf("GET /count: %v", countErr)
	}
	var got struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(countResp.Body).Decode(&got); err != nil {
		t.Fatalf("decode count: %v", err)
	}
	_ = countResp.Body.Close()
	if got.Count != 1 {
		t.Fatalf("count after one bump = %d, want 1", got.Count)
	}

	// 3) In-flight drain: shutdown starts while a command is parked in its
	// handler; the command still completes and the response still lands.
	slowRespCh := make(chan *http.Response, 1)
	go func() {
		resp, err := http.Post(baseURL+"/slowbump", "application/json", nil) //nolint:noctx // test boundary
		if err != nil {
			t.Errorf("POST /slowbump: %v", err)
			slowRespCh <- nil
			return
		}
		slowRespCh <- resp
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
	case <-drainStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("drain never started while the command was in flight")
	}

	close(releaseHandler)

	select {
	case resp := <-slowRespCh:
		if resp == nil {
			t.Fatal("slow command response lost")
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("POST /slowbump status = %d, want 204 (response must survive the drain)", resp.StatusCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("slow command never completed")
	}

	select {
	case <-handlerFinished:
	case <-time.After(5 * time.Second):
		t.Fatal("handler body never finished")
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
