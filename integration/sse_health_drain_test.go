package integration_test

// T25c — SSE × health lifecycle E2E: an open realtime stream on a service
// that also carries the health module's probe surface. Pins the documented
// shutdown choreography (realtime README: hub drain BEFORE service shutdown;
// health module: Mounted.Drain flips readiness for the whole drain window):
//
//	drain begins → hub.Shutdown closes the SSE stream (so http.Server.Shutdown
//	  is not blocked waiting on the stream, which is NOT a hijacked conn) →
//	  mounted.Drain flips /readyz to 503 in lockstep → listener close.
//
// Without the hub drain IN the drain window, an SSE subscriber outlives the
// drain and Shutdown blocks (or times out) on the stream connection.

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
)

func TestSSEStreamClosesInLockstepWithHealthDrain(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	mounted, err := appkithealth.New(
		appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
			"always-ok": func(context.Context) error { return nil },
		}))
	if err != nil {
		t.Fatalf("health mount: %v", err)
	}

	healthDisabled := false
	drainHooksRan := make(chan struct{})

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:           freeAddr(t),
		ReadTimeout:    appkit.NoTimeout,
		WriteTimeout:   appkit.NoTimeout,
		DrainDelay:     150 * time.Millisecond,
		RegisterHealth: &healthDisabled,
		DrainHooks: []appkit.Hook{
			hub.Shutdown,
			func(context.Context) error {
				mounted.Drain()

				return nil
			},
			func(context.Context) error {
				close(drainHooksRan)

				return nil
			},
		},
		ShutdownHooks: []appkit.Hook{mounted.Shutdown},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	realtime.Mount(svc.Mux, "/sse", hub)
	mounted.RegisterRoutes(svc.Mux)

	ts := testkit.Serve(t, svc)
	base := ts.FullChainURL

	// Open one stream and prove it is live through the full chain.
	req, reqErr := http.NewRequestWithContext(
		context.Background(), http.MethodGet, base+"/sse", nil)
	if reqErr != nil {
		t.Fatalf("build stream request: %v", reqErr)
	}

	stream, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}

	defer func() { _ = stream.Body.Close() }()

	if ct := stream.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("stream content-type = %q, want text/event-stream", ct)
	}

	// Broadcast only AFTER the handler's subscription is registered: the
	// header flush precedes the subscribe step, so an earlier broadcast can
	// race it (intermittently lost). SubscriberCount is the sync point.
	subscribed := false
	for range 50 {
		if hub.SubscriberCount() == 1 {
			subscribed = true

			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !subscribed {
		t.Fatal("handler never subscribed to the hub")
	}

	hub.Broadcast(sse.Event{Event: "live", Data: "{}", ID: sse.NewEventID("sse-drain-1")})
	hub.Broadcast(sse.Event{Event: "live", Data: "{}", ID: sse.NewEventID("sse-drain-2")})

	events := ssetest.MustReadNEvents(t, stream.Body, 2)
	for i, id := range []string{"sse-drain-1", "sse-drain-2"} {
		if events[i].Type != "live" || events[i].ID != id {
			t.Fatalf("event %d = %q/%s, want live/%s", i, events[i].Type, events[i].ID, id)
		}
	}

	// Shutdown: the drain hooks close the hub and flip health readiness.
	shutdownErrCh := make(chan error, 1)
	go func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErrCh <- svc.Shutdown(shutdownCtx)
	}()

	select {
	case <-drainHooksRan:
	case <-time.After(5 * time.Second):
		t.Fatal("drain hooks never ran")
	}

	// Lockstep: the probe's readiness is 503 for the whole drain window.
	readyCtx, readyCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer readyCancel()
	status, _, _ := getWithAuth(t, readyCtx, base, "/readyz", "", "")
	if status != http.StatusServiceUnavailable {
		t.Errorf("/readyz during drain = %d, want 503 (health flips in lockstep with the SSE drain)", status)
	}

	// The stream ENDED during the drain: a clean EOF, not a hang — this is
	// what frees http.Server.Shutdown from waiting on the stream connection.
	eof := make(chan error, 1)
	go func() {
		_, readErr := stream.Body.Read(make([]byte, 1))
		eof <- readErr
	}()

	select {
	case readErr := <-eof:
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			t.Fatalf("stream read after drain = %v, want clean EOF", readErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("stream still open after the hub drained — Shutdown would block on it")
	}

	select {
	case err := <-shutdownErrCh:
		if err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown never returned")
	}

	if svc.Running() {
		t.Error("service still Running after Shutdown returned")
	}
}
