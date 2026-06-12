package appkit

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultShutdownTimeout = 15 * time.Second

// ShutdownConfig controls graceful shutdown behavior.
type ShutdownConfig struct {
	Timeout time.Duration
	Signals []os.Signal
}

// DefaultShutdownConfig returns a config that waits for SIGINT/SIGTERM with a 15s timeout.
func DefaultShutdownConfig() ShutdownConfig {
	return ShutdownConfig{
		Timeout: defaultShutdownTimeout,
		Signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
}

// WaitForSignal blocks until ctx is cancelled or one of the configured signals is received.
// It then calls onShutdown with a context limited by Timeout. Any error from onShutdown is returned.
func WaitForSignal(ctx context.Context, cfg ShutdownConfig, onShutdown func(context.Context) error) error {
	if len(cfg.Signals) == 0 {
		cfg.Signals = DefaultShutdownConfig().Signals
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultShutdownConfig().Timeout
	}

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, cfg.Signals...)
	defer signal.Stop(sigCh)

	select {
	case <-ctx.Done():
		return runShutdown(ctx, cfg.Timeout, onShutdown)
	case sig := <-sigCh:
		slog.Info("received signal, shutting down", "signal", sig)

		return runShutdown(ctx, cfg.Timeout, onShutdown)
	}
}

func runShutdown(parent context.Context, timeout time.Duration, onShutdown func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	return onShutdown(ctx)
}
