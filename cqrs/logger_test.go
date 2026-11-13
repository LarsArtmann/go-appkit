package cqrs

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
)

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

func TestEventConfig_Logger_FlowsToProjectionWorkers(t *testing.T) {
	t.Parallel()

	handler := &capturingHandler{}

	eventSvc := newTestEventService(t, EventConfig{
		Logger: slog.New(handler),
	})
	boom := errors.New("handler boom")

	proj := projection.NewProjection(
		"logging-projection",
		func(_ context.Context, _ event.Event) error { return boom },
		[]event.Type{"test.logged"},
	)

	err := eventSvc.Host().Register(proj)
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
