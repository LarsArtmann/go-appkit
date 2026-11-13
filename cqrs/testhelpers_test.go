package cqrs

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// memoryDeployment returns a minimal operator config with a memory engine
// and the two canonical instances.
func memoryDeployment() *system.DeploymentConfig {
	return &system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			defaultEngineName: {Driver: memoryDriver},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: defaultEngineName},
			{Role: system.RoleProjections, Engine: defaultEngineName},
		},
	}
}

// writeFile writes content to path with 0o600 permissions.
func writeFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		return err //nolint:wrapcheck // test helper
	}

	return nil
}

// newTestEventService starts an EventService backed by a throwaway SQLite
// database in a per-test temp dir (unless cfg.DSN is already set) and
// registers its shutdown as test cleanup, replacing the hand-rolled
// New/err-check/defer-shutdown preamble that used to open most tests.
func newTestEventService(t *testing.T, cfg EventConfig) *EventService {
	t.Helper()

	if cfg.DSN == "" {
		cfg.DSN = t.TempDir() + "/test.db"
	}

	svc, err := NewEventService(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Cleanup(func() { _ = svc.Shutdown(context.Background()) })

	return svc
}

// testEventSeq makes appendTestEvent streams unique across calls.
var testEventSeq atomic.Uint64

// appendTestEvent writes a single event of eventType to a fresh stream.
// Each call uses a new stream so saves never conflict on version.
func appendTestEvent(t *testing.T, eventSvc *EventService, eventType event.Type) {
	t.Helper()

	static := testEventSeq.Add(1)

	streamID, err := id.ParseStreamID(fmt.Sprintf("stream-%d", static))
	if err != nil {
		t.Fatalf("parse stream id: %v", err)
	}

	evt, err := event.New(
		eventType,
		streamID,
		"test-stream",
		1,
		map[string]string{"hello": "world"},
	)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	ref := id.NewStreamRef("test-stream", streamID)

	err = eventSvc.System().EventStore().Save(context.Background(), ref, []event.Event{evt}, 0)
	if err != nil {
		t.Fatalf("save event: %v", err)
	}
}

// waitFor polls fn until it returns true or the deadline passes.
func waitFor(t *testing.T, what string, fn func() bool) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		if fn() {
			return
		}

		time.Sleep(10 * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}
