package cqrs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
)

// testEventSeq makes appendTestEvent streams unique across calls.
var testEventSeq atomic.Uint64

// capturingHandler is a slog.Handler that records every record it receives.
type capturingHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.records = append(h.records, r.Clone())

	return nil
}

func (h *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *capturingHandler) WithGroup(string) slog.Handler { return h }

// contains reports whether any captured record mentions substr in its
// message or in any top-level attribute value.
func (h *capturingHandler) contains(substr string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, r := range h.records {
		if strings.Contains(r.Message, substr) {
			return true
		}

		mentioned := false

		r.Attrs(func(a slog.Attr) bool {
			if strings.Contains(a.Value.String(), substr) {
				mentioned = true

				return false
			}

			return true
		})

		if mentioned {
			return true
		}
	}

	return false
}

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

	err = eventSvc.Bundle().EventSink.Save(context.Background(), ref, []event.Event{evt}, 0)
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
	}

	t.Fatalf("timed out waiting for %s", what)
}

func TestEventConfig_Logger_FlowsToProjectionWorkers(t *testing.T) {
	t.Parallel()

	handler := &capturingHandler{}

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		Logger:     slog.New(handler),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	boom := errors.New("handler boom")

	proj := projection.NewProjection(
		"logging-projection",
		func(_ context.Context, _ event.Event) error { return boom },
		[]event.Type{"test.logged"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.logged")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "worker error to reach configured logger", func() bool {
		return handler.contains("logging-projection")
	})
}
