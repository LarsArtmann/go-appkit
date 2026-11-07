package appkit

import (
	"context"
	"testing"
	"time"
)

// A config appkit's own Validate accepts but the delegated server rejects
// (ReadHeaderTimeout after ReadTimeout expires is a misordering httputil
// surfaces at construction) must fail Start synchronously with a classified
// Rejection before any observable lifecycle state exists.
func TestStart_DelegatedConfigRejection(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:              "localhost:0",
		ReadTimeout:       10 * time.Millisecond,
		ReadHeaderTimeout: 20 * time.Millisecond,
		DrainDelay:        NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Start()
	if err == nil {
		_ = svc.Close()

		t.Fatal("expected Start to reject the misordered timeout config")
	}

	if svc.Addr() != nil {
		t.Error("expected nil Addr after a rejected start")
	}

	if svc.Running() {
		t.Error("expected Running() false after a rejected start")
	}
}

// The delegated server is single-serve by upstream contract, so Service must
// construct a fresh one for every Start: a full shutdown must not poison a
// restart (the served-out listener is consumed, a new listener is bound).
func TestStart_RestartAfterFullShutdown(t *testing.T) {
	t.Parallel()

	disabled := false

	svc, err := NewService(ServiceConfig{
		Addr:           "localhost:0",
		RegisterHealth: &disabled,
		DrainDelay:     NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	defer func() { _ = svc.Close() }()

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("first start: %v", err)
	}

	waitForRunning(t, svc)
	_ = svc.Shutdown(context.Background())

	assertServerStopped(t, errCh)

	errCh, err = svc.Start()
	if err != nil {
		t.Fatalf("restart after shutdown: %v", err)
	}

	waitForRunning(t, svc)

	if svc.Addr() == nil {
		t.Fatal("expected non-nil Addr after restart")
	}

	_ = svc.Shutdown(context.Background())

	assertServerStopped(t, errCh)
}

// The delegated server's config carries the SAME ShutdownTimeout field that
// Shutdown uses for its context deadline (one source of truth) and passes
// every other timeout through with the NoTimeout sentinel mapped to zero.
func TestServerConfig_MapsServiceConfig(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:            "localhost:0",
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    NoTimeout,
		ShutdownTimeout: 7 * time.Second,
		DrainDelay:      NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	cfg := svc.serverConfig(":9")

	if cfg.Addr != ":9" {
		t.Errorf("Addr = %q, want the bound listener address", cfg.Addr)
	}

	if cfg.ReadTimeout != 3*time.Second {
		t.Errorf("ReadTimeout = %v, want 3s passthrough", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 0 {
		t.Errorf("WriteTimeout = %v, want 0 (NoTimeout sentinel mapped)", cfg.WriteTimeout)
	}

	if cfg.ShutdownTimeout != 7*time.Second {
		t.Errorf("ShutdownTimeout = %v, want the same config field Shutdown deadlines on", cfg.ShutdownTimeout)
	}
}
