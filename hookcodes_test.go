package appkit

import (
	"context"
	"errors"
	"strings"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

// The hook error codes are a consumer-matchable contract (grep-able in logs
// and assertable via errorfamily.Code): consumers key alerting and tests on
// these exact strings, so renaming one is a breaking change.
func TestHookErrorCodes_AreAStableContract(t *testing.T) {
	t.Parallel()

	t.Run("drain hook failure carries server.drain_hook_failed", func(t *testing.T) {
		t.Parallel()

		svc, err := NewService(ServiceConfig{
			Addr:       "localhost:0",
			DrainDelay: NoDrainDelay,
			DrainHooks: []func(context.Context) error{
				func(context.Context) error { return errors.New("readiness flush failed") },
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		errCh, err := svc.Start()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		waitForRunning(t, svc)

		shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
		defer cancel()

		err = svc.Shutdown(shutdownCtx)
		if err == nil {
			t.Fatal("expected joined drain-hook error, got nil")
		}

		if code := errorfamily.Code(err); code != "server.drain_hook_failed" {
			t.Fatalf("expected code server.drain_hook_failed, got %q", code)
		}

		assertServerStopped(t, errCh)
	})

	t.Run("shutdown hook failure carries server.shutdown_hook_failed", func(t *testing.T) {
		t.Parallel()

		svc, err := NewService(ServiceConfig{
			Addr:       "localhost:0",
			DrainDelay: NoDrainDelay,
			ShutdownHooks: []func(context.Context) error{
				func(context.Context) error { return errors.New("telemetry flush failed") },
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		errCh, err := svc.Start()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		waitForRunning(t, svc)

		shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
		defer cancel()

		err = svc.Shutdown(shutdownCtx)
		if err == nil {
			t.Fatal("expected joined shutdown-hook error, got nil")
		}

		if code := errorfamily.Code(err); code != "server.shutdown_hook_failed" {
			t.Fatalf("expected code server.shutdown_hook_failed, got %q", code)
		}

		assertServerStopped(t, errCh)
	})
}

func TestHooks_EveryHookRunsWhenAnEarlierOneFails(t *testing.T) {
	t.Parallel()

	var calls []string

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		ShutdownHooks: []func(context.Context) error{
			func(context.Context) error {
				calls = append(calls, "first")

				return errors.New("first failed")
			},
			func(context.Context) error {
				calls = append(calls, "second")

				return errors.New("second failed")
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	waitForRunning(t, svc)

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	err = svc.Shutdown(shutdownCtx)
	if err == nil {
		t.Fatal("expected joined hook errors, got nil")
	}

	if len(calls) != 2 {
		t.Fatalf("expected both hooks to run, got %v", calls)
	}

	for _, want := range []string{"first failed", "second failed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("joined error missing %q: %v", want, err)
		}
	}

	assertServerStopped(t, errCh)
}
