package appkit

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

var errDrainHookFailed = errors.New("drain failed")

func TestDrainHooks_RunInTrafficWindowBeforeShutdown(t *testing.T) {
	t.Parallel()

	var svc *Service

	var events []string

	readyDownDuringDrain := false
	drainHookAddr := ""

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 150 * time.Millisecond,
		DrainHooks: []func(context.Context) error{
			func(ctx context.Context) error {
				resp := httpGet(t, ctx, "http://"+drainHookAddr+"/health/ready")
				defer resp.Body.Close()

				_, _ = io.Copy(io.Discard, resp.Body)

				// The request completing at all proves the socket still
				// serves traffic; the code proves the ready probe already
				// flipped for the whole drain window.
				readyDownDuringDrain = resp.StatusCode == http.StatusServiceUnavailable

				events = append(events, "drain")

				return nil
			},
		},
		ShutdownHooks: []func(context.Context) error{
			func(context.Context) error {
				events = append(events, "shutdown")

				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	// Shutdown detaches the listener before drain hooks run (the socket
	// itself keeps serving until server.Shutdown), so capture the address
	// beforehand.
	drainHookAddr = svc.Addr().String()

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	err = svc.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if !readyDownDuringDrain {
		t.Error("readiness must already report 503 while drain hooks execute")
	}

	if len(events) != 2 || events[0] != "drain" || events[1] != "shutdown" {
		t.Errorf("hook order = %v, want [drain shutdown]", events)
	}

	assertServerStopped(t, errCh)
}

func TestDrainHooks_AllRunDespiteFailure(t *testing.T) {
	t.Parallel()

	calls := 0

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		DrainHooks: []func(context.Context) error{
			func(context.Context) error { return errDrainHookFailed },
			func(context.Context) error {
				calls++

				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	shutdownErr := svc.Close()
	if !errors.Is(shutdownErr, errDrainHookFailed) {
		t.Fatalf("shutdown error = %v, want wrapped %v", shutdownErr, errDrainHookFailed)
	}

	err = svc.Close()
	if err != nil {
		t.Fatalf("second close: %v", err)
	}

	if calls != 1 {
		t.Errorf("hooks ran %d times across two shutdowns, want exactly 1", calls)
	}

	assertServerStopped(t, errCh)
}

func TestDrainHooks_NeverStartedServiceSkipsHooks(t *testing.T) {
	t.Parallel()

	called := false

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		DrainHooks: []func(context.Context) error{
			func(context.Context) error {
				called = true

				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.Close()
	if err != nil {
		t.Fatalf("close: %v", err)
	}

	if called {
		t.Error("drain hooks must not run for a service that never started")
	}
}
