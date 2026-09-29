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
var (
	errFlushFailed  = errors.New("flush failed")
	errFirstFailed  = errors.New("first failed")
	errSecondFailed = errors.New("second failed")
)

func TestHookErrorCodes_AreAStableContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hooks    func(func(context.Context) error) ServiceConfig
		wantCode string
	}{
		{
			name: "drain hook failure carries server.drain_hook_failed",
			hooks: func(fail func(context.Context) error) ServiceConfig {
				return ServiceConfig{
					Addr:       "localhost:0",
					DrainDelay: NoDrainDelay,
					DrainHooks: []func(context.Context) error{fail},
				}
			},
			wantCode: "server.drain_hook_failed",
		},
		{
			name: "shutdown hook failure carries server.shutdown_hook_failed",
			hooks: func(fail func(context.Context) error) ServiceConfig {
				return ServiceConfig{
					Addr:          "localhost:0",
					DrainDelay:    NoDrainDelay,
					ShutdownHooks: []func(context.Context) error{fail},
				}
			},
			wantCode: "server.shutdown_hook_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewService(tt.hooks(func(context.Context) error {
				return errFlushFailed
			}))
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
				t.Fatalf("expected joined %s error, got nil", tt.wantCode)
			}

			if code := errorfamily.Code(err); code != tt.wantCode {
				t.Fatalf("expected code %s, got %q", tt.wantCode, code)
			}

			assertServerStopped(t, errCh)
		})
	}
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

				return errFirstFailed
			},
			func(context.Context) error {
				calls = append(calls, "second")

				return errSecondFailed
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
