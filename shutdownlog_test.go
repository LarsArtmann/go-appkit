package appkit

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"
	"testing"
	"time"
)

var errShutdownHookFailed = errors.New("shutdown hook failed")

type capturedRecord struct {
	message string
	attrs   map[string]string
}

type capturedLog struct {
	mu      sync.Mutex
	records []capturedRecord
}

func (c *capturedLog) Enabled(context.Context, slog.Level) bool { return true }

func (c *capturedLog) Handle(_ context.Context, r slog.Record) error {
	rec := capturedRecord{message: r.Message, attrs: map[string]string{}}

	r.Attrs(func(a slog.Attr) bool {
		rec.attrs[a.Key] = a.Value.String()

		return true
	})

	c.mu.Lock()
	defer c.mu.Unlock()

	c.records = append(c.records, rec)

	return nil
}

func (c *capturedLog) WithAttrs([]slog.Attr) slog.Handler { return c }

func (c *capturedLog) WithGroup(string) slog.Handler { return c }

func (c *capturedLog) snapshot() []capturedRecord {
	c.mu.Lock()
	defer c.mu.Unlock()

	return slices.Clone(c.records)
}

func shutdownPhases(records []capturedRecord) []string {
	phases := make([]string, 0, 4)

	for _, rec := range records {
		if rec.message == "shutdown phase complete" {
			phases = append(phases, rec.attrs["phase"])
		}
	}

	return phases
}

func findRecord(records []capturedRecord, message string) (capturedRecord, bool) {
	for _, rec := range records {
		if rec.message == message {
			return rec, true
		}
	}

	return capturedRecord{}, false
}

func TestShutdownLogsPhaseSequence(t *testing.T) {
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

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	records := capture.snapshot()

	want := []string{"ready_flip", "drain_hooks", "listener_close", "shutdown_hooks"}
	if got := shutdownPhases(records); !slices.Equal(got, want) {
		t.Errorf("phases = %v, want %v", got, want)
	}

	skipped, found := findRecord(records, "shutdown phase skipped")
	if !found || skipped.attrs["phase"] != "drain_wait" {
		t.Errorf("expected drain_wait skipped line, got %+v (found=%t)", skipped, found)
	}

	final, found := findRecord(records, "graceful shutdown complete")
	if !found {
		t.Fatal("missing final shutdown line")
	}

	if final.attrs["result"] != "ok" {
		t.Errorf("result = %q, want %q", final.attrs["result"], "ok")
	}

	if final.attrs["total"] == "" {
		t.Error("final shutdown line must carry the total duration")
	}

	if last := records[len(records)-1]; last.message != "graceful shutdown complete" {
		t.Errorf("final line must come last, got %q", last.message)
	}

	assertServerStopped(t, errCh)
}

func TestShutdownLogsDrainWaitPhaseWhenEnabled(t *testing.T) {
	t.Parallel()

	const drainDelay = 25 * time.Millisecond

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: drainDelay,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	capture := &capturedLog{}
	svc.Logger = slog.New(capture)

	if _, err := svc.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	records := capture.snapshot()

	want := []string{"ready_flip", "drain_hooks", "drain_wait", "listener_close", "shutdown_hooks"}
	if got := shutdownPhases(records); !slices.Equal(got, want) {
		t.Fatalf("phases = %v, want %v", got, want)
	}

	var drainWait capturedRecord

	for _, rec := range records {
		if rec.message == "shutdown phase complete" && rec.attrs["phase"] == "drain_wait" {
			drainWait = rec
		}
	}

	if drainWait.attrs["phase"] != "drain_wait" {
		t.Fatal("missing drain_wait phase record")
	}

	duration, err := time.ParseDuration(drainWait.attrs["duration"])
	if err != nil {
		t.Fatalf("parse drain_wait duration %q: %v", drainWait.attrs["duration"], err)
	}

	if duration < drainDelay {
		t.Errorf("drain_wait duration = %v, want >= %v", duration, drainDelay)
	}

	if _, found := findRecord(records, "draining traffic"); !found {
		t.Error("expected the pre-wait 'draining traffic' line to stay")
	}
}

func TestShutdownLogsErrorResultWhenHookFails(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		ShutdownHooks: []func(context.Context) error{
			func(context.Context) error { return errShutdownHookFailed },
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	capture := &capturedLog{}
	svc.Logger = slog.New(capture)

	if _, err := svc.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	shutdownErr := svc.Shutdown(shutdownCtx)
	if !errors.Is(shutdownErr, errShutdownHookFailed) {
		t.Fatalf("shutdown err = %v, want %v", shutdownErr, errShutdownHookFailed)
	}

	records := capture.snapshot()

	final, found := findRecord(records, "graceful shutdown complete")
	if !found {
		t.Fatal("missing final shutdown line")
	}

	if final.attrs["result"] != "error" {
		t.Errorf("result = %q, want %q", final.attrs["result"], "error")
	}

	if got := shutdownPhases(records); !slices.Equal(
		got,
		[]string{"ready_flip", "drain_hooks", "listener_close", "shutdown_hooks"},
	) {
		t.Errorf("phases = %v, want the full sequence despite the hook failure", got)
	}
}
