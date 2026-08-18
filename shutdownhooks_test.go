package appkit

import (
	"context"
	"errors"
	"testing"
)

var errHookFlushFailed = errors.New("flush failed")

func TestShutdownHooks_RunOnceAfterConnectionsReleased(t *testing.T) {
	t.Parallel()

	var svc *Service

	runningDuringHook := true
	calls := 0

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		ShutdownHooks: []func(context.Context) error{
			func(context.Context) error { return nil },
			func(context.Context) error {
				calls++

				// The listener is cleared before hooks run: Shutdown has
				// already progressed past its point of no return. (A TCP dial
				// would NOT be a reliable proof here — a closed listener's
				// backlog can still complete handshakes at the kernel level.)
				runningDuringHook = svc.Running()

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

	if calls != 1 {
		t.Fatalf("hook calls = %d, want 1", calls)
	}

	if runningDuringHook {
		t.Error("service still reports running while hooks execute; listener must be released first")
	}

	assertServerStopped(t, errCh)
}

func TestShutdownHooks_SecondShutdownDoesNotRerun(t *testing.T) {
	t.Parallel()

	called := make(chan struct{}, 2)

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		ShutdownHooks: []func(context.Context) error{
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
		t.Fatalf("hook ran %d times, want exactly 1", len(called))
	}

	assertServerStopped(t, errCh)
}

func TestShutdownHooks_AllRunDespiteFailure(t *testing.T) {
	t.Parallel()

	secondRan := false

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		ShutdownHooks: []func(context.Context) error{
			func(context.Context) error { return errHookFlushFailed },
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
	if !errors.Is(shutdownErr, errHookFlushFailed) {
		t.Fatalf("shutdown error = %v, want wrapped %v", shutdownErr, errHookFlushFailed)
	}

	if !secondRan {
		t.Error("a failing hook must not stop the remaining hooks")
	}

	assertServerStopped(t, errCh)
}

func TestShutdownHooks_NeverStartedServiceSkipsHooks(t *testing.T) {
	t.Parallel()

	called := false

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		ShutdownHooks: []func(context.Context) error{
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
		t.Error("hooks must not run for a service that never started")
	}
}
