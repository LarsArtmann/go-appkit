package appkit

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

var errStartHookFailed = errors.New("start hook failed")

func TestStartHooks_RunAfterListenInsideStart(t *testing.T) {
	t.Parallel()

	var svc *Service

	var (
		hookAddr          string
		runningInHook     bool
		runningAfterStart bool
	)

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		StartHooks: []func(context.Context) error{
			func(context.Context) error {
				// The post-listen contract, mirror of the drain-window
				// contract (Addr nil inside DrainHooks): inside StartHooks
				// the listener is already bound and observable.
				hookAddr = svc.Addr().String()
				runningInHook = svc.Running()

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

	// The hook ran synchronously inside Start: its observations are
	// already final the moment Start returns, before any polling.
	runningAfterStart = svc.Running()

	if hookAddr == "" {
		t.Error("Addr must be non-nil inside StartHooks (post-listen contract)")
	}

	if !runningInHook {
		t.Error("Running must report true inside StartHooks (post-listen contract)")
	}

	if !runningAfterStart {
		t.Error("Running must stay true after a successful Start")
	}

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	assertServerStopped(t, errCh)
}

func TestStartHooks_RunInOrder(t *testing.T) {
	t.Parallel()

	var calls []string

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		StartHooks: []func(context.Context) error{
			func(context.Context) error {
				calls = append(calls, "first")

				return nil
			},
			func(context.Context) error {
				calls = append(calls, "second")

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

	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Errorf("hook order = %v, want [first second]", calls)
	}

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	assertServerStopped(t, errCh)
}

func TestStartHooks_FailureFailsStartClosesListener(t *testing.T) {
	t.Parallel()

	secondRan := false

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		StartHooks: []func(context.Context) error{
			func(context.Context) error { return errStartHookFailed },
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
	if err == nil {
		t.Fatal("expected start to fail when a StartHook fails")
	}

	if errCh != nil {
		t.Error("no serve channel may exist for a failed start")
	}

	if !errors.Is(err, errStartHookFailed) {
		t.Errorf("errors.Is(start err, sentinel) = false: %v", err)
	}

	if code := errorfamily.Code(err); code != "server.start_hook_failed" {
		t.Errorf("error code = %q, want server.start_hook_failed", code)
	}

	// Every hook runs even when an earlier one fails (runHooks semantics,
	// identical to DrainHooks/ShutdownHooks).
	if !secondRan {
		t.Error("the second StartHook must still run after the first failed")
	}

	// The listener was closed again: the service never started.
	if svc.Running() {
		t.Error("Running must be false after a failed start")
	}

	if svc.Addr() != nil {
		t.Error("Addr must be nil after a failed start")
	}

	// Never-started semantics hold: Shutdown is a no-op, not an error.
	if err := svc.Close(); err != nil {
		t.Errorf("Close after failed start = %v, want nil", err)
	}
}

func TestStartLogsPhaseLine(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	capture := &capturedLog{}
	svc.Logger = slog.New(capture)

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	phase, found := findRecord(capture.snapshot(), "startup phase complete")
	if !found {
		t.Fatal("missing 'startup phase complete' line")
	}

	if phase.attrs["phase"] != "start_hooks" {
		t.Errorf("phase = %q, want start_hooks", phase.attrs["phase"])
	}

	if phase.attrs["duration"] == "" {
		t.Error("startup phase line must carry the duration")
	}

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	assertServerStopped(t, errCh)
}

func TestStartHookFailureLogsNoPhaseLine(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		StartHooks: []Hook{func(context.Context) error { return errStartHookFailed }},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	capture := &capturedLog{}
	svc.Logger = slog.New(capture)

	if _, err := svc.Start(); err == nil {
		t.Fatal("expected start to fail")
	}

	if _, found := findRecord(capture.snapshot(), "startup phase complete"); found {
		t.Error("a failed startup must not log a phase completion line")
	}
}
