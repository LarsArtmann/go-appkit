package cqrs

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
)

// recorderMu serializes tests that hold the process-global flight recorder
// slot (runtime/trace allows exactly one active recorder per process).
var recorderMu = newChanMutex()

type chanMutex struct{ ch chan struct{} }

func (m chanMutex) lock()   { <-m.ch }
func (m chanMutex) unlock() { m.ch <- struct{}{} }

func newChanMutex() chanMutex {
	m := chanMutex{ch: make(chan struct{}, 1)}
	m.unlock()

	return m
}

// NOT parallel: holds the single process-global flight recorder slot.
func TestEventConfig_FlightRecorder_CapturesOnWorkerFailure(t *testing.T) {
	recorderMu.lock()
	defer recorderMu.unlock()

	buf := &bytes.Buffer{}

	rec, err := fr.New(fr.WithWriter(buf), fr.WithMinAge(time.Nanosecond))
	if err != nil {
		t.Fatalf("create recorder: %v", err)
	}

	err = rec.Start()
	if err != nil {
		t.Fatalf("start recorder: %v", err)
	}

	defer rec.Stop()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath:     t.TempDir() + "/test.db",
		FlightRecorder: rec,
		// Fail fast: first handler error exhausts the restart budget, no
		// backoff sleeps.
		HostOptions: []projectionhost.HostOption{
			projectionhost.WithMaxRestarts(0),
			projectionhost.WithBackoff(time.Nanosecond, time.Nanosecond),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"fr-projection",
		func(_ context.Context, _ event.Event) error {
			return errors.New("terminal failure")
		},
		[]event.Type{"test.fr"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.fr")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "worker to reach WorkerFailed", func() bool {
		for _, state := range eventSvc.Host().Status() {
			if state.Name == "fr-projection" && state.Status == projectionhost.WorkerFailed {
				return true
			}
		}

		return false
	})

	waitFor(t, "trace snapshot in recorder buffer", func() bool {
		return buf.Len() > 0
	})
}

// NOT parallel: holds the single process-global flight recorder slot.
func TestEventConfig_FlightRecorderTrigger_ReceivesProjectionContext(t *testing.T) {
	recorderMu.lock()
	defer recorderMu.unlock()

	buf := &bytes.Buffer{}

	rec, err := fr.New(fr.WithWriter(buf), fr.WithMinAge(time.Nanosecond))
	if err != nil {
		t.Fatalf("create recorder: %v", err)
	}

	err = rec.Start()
	if err != nil {
		t.Fatalf("start recorder: %v", err)
	}

	defer rec.Stop()

	var mu sync.Mutex

	var seen []fr.TriggerContext

	trigger := func(tc fr.TriggerContext) bool {
		mu.Lock()
		seen = append(seen, tc)
		mu.Unlock()

		return true
	}

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath:            t.TempDir() + "/test.db",
		FlightRecorder:        rec,
		FlightRecorderTrigger: trigger,
		HostOptions: []projectionhost.HostOption{
			projectionhost.WithMaxRestarts(0),
			projectionhost.WithBackoff(time.Nanosecond, time.Nanosecond),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"trigger-projection",
		func(_ context.Context, _ event.Event) error {
			return errors.New("terminal failure")
		},
		[]event.Type{"test.trigger"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.trigger")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "custom trigger invoked", func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(seen) > 0
	})

	mu.Lock()
	tc := seen[0]
	mu.Unlock()

	if tc.Kind != "projection" || tc.Type != "trigger-projection" || tc.Err == nil {
		t.Fatalf("unexpected trigger context: kind=%q type=%q err=%v", tc.Kind, tc.Type, tc.Err)
	}

	waitFor(t, "trace snapshot in recorder buffer", func() bool {
		return buf.Len() > 0
	})
}
