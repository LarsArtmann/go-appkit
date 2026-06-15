package appkit

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

var errShutdownFailed = errors.New("shutdown failed")

func TestWaitForSignal_CallsShutdownOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	called := make(chan struct{})

	go func() {
		_ = WaitForSignal(ctx, ShutdownConfig{Timeout: time.Second}, newMockShutdown(called))
	}()

	cancel()

	assertShutdownCalled(t, called, "shutdown callback was not called")
}

func TestWaitForSignal_ShutdownErrorReturned(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wantErr := errShutdownFailed

	errCh := make(chan error, 1)

	go func() {
		errCh <- WaitForSignal(ctx, ShutdownConfig{Timeout: time.Second}, func(_ context.Context) error {
			return wantErr
		})
	}()

	cancel()

	assertErrorWithin(t, errCh, wantErr)
}

func TestWaitForSignal_SignalDelivery(t *testing.T) {
	t.Parallel()

	called := make(chan struct{})

	go func() {
		_ = WaitForSignal(context.Background(), ShutdownConfig{
			Timeout: time.Second,
			Signals: []os.Signal{syscall.SIGUSR1},
		}, newMockShutdown(called))
	}()

	time.Sleep(50 * time.Millisecond)

	killErr := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	if killErr != nil {
		t.Fatalf("syscall.Kill failed: %v", killErr)
	}

	assertShutdownCalled(t, called, "shutdown callback was not called after signal")
}

func TestWaitForSignal_DefaultConfig(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	called := make(chan struct{})

	go func() {
		_ = WaitForSignal(ctx, ShutdownConfig{}, newMockShutdown(called))
	}()

	cancel()

	assertShutdownCalled(t, called, "shutdown callback was not called with default config")
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
