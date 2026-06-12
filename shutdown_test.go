package appkit

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWaitForSignal_CallsShutdownOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	called := make(chan struct{})

	go func() {
		_ = WaitForSignal(ctx, ShutdownConfig{Timeout: time.Second}, func(_ context.Context) error {
			close(called)

			return nil
		})
	}()

	cancel()

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown callback was not called")
	}
}

func TestWaitForSignal_ShutdownErrorReturned(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wantErr := errors.New("shutdown failed")

	errCh := make(chan error, 1)

	go func() {
		errCh <- WaitForSignal(ctx, ShutdownConfig{Timeout: time.Second}, func(_ context.Context) error {
			return wantErr
		})
	}()

	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown did not return")
	}
}

func TestWaitForSignal_SignalDelivery(t *testing.T) {
	t.Parallel()

	called := make(chan struct{})

	go func() {
		_ = WaitForSignal(context.Background(), ShutdownConfig{
			Timeout: time.Second,
			Signals: []os.Signal{syscall.SIGUSR1},
		}, func(_ context.Context) error {
			close(called)

			return nil
		})
	}()

	time.Sleep(50 * time.Millisecond)

	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown callback was not called after signal")
	}
}

func TestWaitForSignal_DefaultConfig(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	called := make(chan struct{})

	go func() {
		_ = WaitForSignal(ctx, ShutdownConfig{}, func(_ context.Context) error {
			close(called)

			return nil
		})
	}()

	cancel()

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown callback was not called with default config")
	}
}

func TestWaitForSignal_DefaultTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultShutdownConfig()

	if cfg.Timeout != 15*time.Second {
		t.Errorf("default timeout = %v, want %v", cfg.Timeout, 15*time.Second)
	}

	if len(cfg.Signals) != 2 {
		t.Fatalf("expected 2 default signals, got %d", len(cfg.Signals))
	}
}
