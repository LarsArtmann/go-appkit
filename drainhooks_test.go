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
	stillRunningDuringDrain := false

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 150 * time.Millisecond,
		DrainHooks: []func(context.Context) error{
			func(context.Context) error {
				resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/health/ready")
				defer resp.Body.Close() //nolint:errcheck // test response body

				_, _ = io.Copy(io.Discard, resp.Body)

				readyDownDuringDrain = resp.StatusCode == http.StatusServiceUnavailable
				stillRunningDuringDrain = svc.Running()

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

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	err = svc.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if !readyDownDuringDrain {
		t.Error("readiness must already report 503 while drain hooks execute")
	}

	if !stillRunningDuringDrain {
		t.Error("drain hooks must run before the listener closes, while traffic is still served")
	}

	if len(events) != 2 || events[0] != "drain" || events[1] != "shutdown" {
		t.Errorf("hook order = %v, want [drain shutdown]", events)
	}

	assertServerStopped(t, errCh)
}

func TestDrainHooks_AllRunDespiteFailure(t *testing.T) {
	t.Parallel()

	secondRan := false

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		DrainHooks: []func(context.Context) error{
			func(context.Context) error { return errDrainHookFailed },
			func(context.Context) error {
				secondRan = true

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

	if !secondRan {
		t.Error("a failing drain hook must not stop the remaining hooks")
	}

	assertServerStopped(t, errCh)
}

func TestDrainHooks_SecondShutdownDoesNotRerun(t *testing.T) {
	t.Parallel()

	called := make(chan struct{}, 2)

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		DrainHooks: []func(context.Context) error{
			func(context.Context) error {
				called <- struct{}{}

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

	ctx := t.Context()

	err = svc.Shutdown(ctx)
	if err != nil {
		t.Fatalf("first shutdown: %v", err)
	}

	err = svc.Shutdown(ctx)
	if err != nil {
		t.Fatalf("second shutdown: %v", err)
	}

	if len(called) != 1 {
		t.Fatalf("drain hook ran %d times, want exactly 1", len(called))
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
