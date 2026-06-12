package appkit

import (
	"context"
	"errors"
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
