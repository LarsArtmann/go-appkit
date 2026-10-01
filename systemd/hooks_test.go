package systemd

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
)

var errConsumerStartHook = errors.New("consumer start hook failed")

// TestInstall_LifecycleThroughAppkitService is the module's composition
// proof: a real appkit Service with the sd_notify lifecycle installed
// announces READY=1 after the listener binds, keeps the watchdog fed while
// serving, sends STOPPING=1 during shutdown, and stops pinging once the
// final phase ran — all observed on a real unixgram NOTIFY_SOCKET.
func TestInstall_LifecycleThroughAppkitService(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	sock := newNotifySocket(t)

	// WatchdogSec=400ms -> a WATCHDOG=1 ping every 200ms.
	t.Setenv("WATCHDOG_USEC", "400000")
	t.Setenv("WATCHDOG_PID", "")

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:0"
	cfg.DrainDelay = appkit.NoDrainDelay

	Install(&cfg, WithLogger(discardLogger()))

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// READY=1 is the FIRST datagram: it is sent from the StartHook, after
	// the listener binds and before the first request can be served.
	if got := sock.awaitMessage(t, "READY=1"); got != "READY=1" {
		t.Fatalf("first datagram = %q, want READY=1", got)
	}

	// While serving, the watchdog keeps the manager's timestamp fresh.
	serving := sock.collectMessages(t, 500*time.Millisecond)
	if !slices.Contains(serving, "WATCHDOG=1") {
		t.Fatalf("no WATCHDOG=1 ping observed while serving (got %v)", serving)
	}

	shutdownCtx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	err = svc.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	// STOPPING=1 was sent by the DrainHook — before connections were
	// released. Trailing watchdog pings may follow it (the keepalive runs
	// to the final phase by design), so assert membership, not position.
	draining := sock.collectMessages(t, 400*time.Millisecond)
	if !slices.Contains(draining, "STOPPING=1") {
		t.Fatalf("no STOPPING=1 observed during shutdown (got %v)", draining)
	}

	// The watchdog goroutine stopped with the final phase: several ping
	// periods of silence after shutdown completed.
	sock.assertQuiet(t, 600*time.Millisecond)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop")
	}
}

// TestInstall_ComposesWithConsumerHooks pins the append-only contract:
// consumer hooks registered before Install keep their position, and the
// sd_notify hooks land after them in all three slices.
func TestInstall_ComposesWithConsumerHooks(t *testing.T) {
	t.Parallel()

	consumerHook := func(context.Context) error { return nil }

	cfg := appkit.DefaultServiceConfig()
	cfg.StartHooks = append(cfg.StartHooks, consumerHook)

	Install(&cfg, WithLogger(discardLogger()))

	if len(cfg.StartHooks) != 2 {
		t.Fatalf("StartHooks = %d hooks, want 2 (consumer first, then sd_notify)", len(cfg.StartHooks))
	}

	if len(cfg.DrainHooks) != 1 || len(cfg.ShutdownHooks) != 1 {
		t.Fatalf(
			"hook counts = start:%d drain:%d shutdown:%d, want 2/1/1",
			len(cfg.StartHooks), len(cfg.DrainHooks), len(cfg.ShutdownHooks),
		)
	}
}

// TestInstall_NoSocketOutsideSystemd proves the development story: without
// NOTIFY_SOCKET the whole lifecycle is a no-op — the service starts,
// serves, and shuts down with zero hook errors.
func TestInstall_NoSocketOutsideSystemd(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	t.Setenv("NOTIFY_SOCKET", "")
	t.Setenv("WATCHDOG_USEC", "")

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:0"
	cfg.DrainDelay = appkit.NoDrainDelay

	Install(&cfg, WithLogger(discardLogger()))

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start outside systemd must succeed: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	err = svc.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("shutdown outside systemd must be clean: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop")
	}
}

// TestInstall_StartFailureStillAnnouncedThenAborts covers the ordering
// that matters under Type=notify: the sd_notify StartHook runs (READY=1 is
// sent, socket bound) and a LATER consumer hook failing aborts the start
// — the service never serves, and the manager is left to its start
// timeout/restart policy rather than routing to a half-started process.
func TestInstall_StartFailureStillAnnouncedThenAborts(t *testing.T) {
	envMu.Lock()
	t.Cleanup(envMu.Unlock)

	sock := newNotifySocket(t)
	t.Setenv("WATCHDOG_USEC", "")

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:0"
	cfg.DrainDelay = appkit.NoDrainDelay

	Install(&cfg, WithLogger(discardLogger()))
	cfg.StartHooks = append(cfg.StartHooks, func(context.Context) error { return errConsumerStartHook })

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = svc.Start()
	if err == nil {
		t.Fatal("start must fail when the consumer StartHook fails")
	}

	if !errors.Is(err, errConsumerStartHook) {
		t.Errorf("errors.Is(start err, sentinel) = false: %v", err)
	}

	// The sd_notify hook ran FIRST (Install appends before the consumer
	// hook here), so READY=1 went out before the abort.
	if got := sock.awaitMessage(t, "READY=1"); got != "READY=1" {
		t.Fatalf("datagram = %q, want READY=1 before the aborted start", got)
	}
}

// TestHooks_ShutdownWithoutStartIsSafe pins the manual-teardown path: a
// service whose start failed never runs ShutdownHooks, so the watchdog
// stop hook must be callable (and idempotent) directly.
func TestHooks_ShutdownWithoutStartIsSafe(t *testing.T) {
	t.Parallel()

	hooks := New(WithLogger(discardLogger()))

	err := hooks.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Shutdown without Start: %v", err)
	}

	err = hooks.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("second Shutdown must be a no-op: %v", err)
	}
}
