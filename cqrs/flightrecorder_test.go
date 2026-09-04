package cqrs

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	fr "github.com/larsartmann/go-flightrecorder"
)

// recorderMu serializes tests that hold the process-global flight recorder
// slot (runtime/trace allows exactly one active recorder per process).
var recorderMu = newChanMutex()

// syncBuffer guards a bytes.Buffer shared between the projection worker
// goroutine (flight recorder snapshot writes) and the test goroutine
// (length polls in waitFor).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Write(p) //nolint:wrapcheck // pass-through writer, caller is go-flightrecorder
}

func (s *syncBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Len()
}

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

	buf := &syncBuffer{}

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

	buf := &syncBuffer{}

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

// NOT parallel: holds the single process-global flight recorder slot.
func TestEventConfig_FlightRecorderTrigger_FalseGateSkipsCapture(t *testing.T) {
	recorderMu.lock()
	defer recorderMu.unlock()

	buf := &syncBuffer{}

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
		// Gate refuses every capture.
		FlightRecorderTrigger: func(fr.TriggerContext) bool { return false },
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
		"gate-projection",
		func(_ context.Context, _ event.Event) error {
			return errors.New("terminal failure")
		},
		[]event.Type{"test.gate"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.gate")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "worker to reach WorkerFailed", func() bool {
		for _, state := range eventSvc.Host().Status() {
			if state.Name == "gate-projection" && state.Status == projectionhost.WorkerFailed {
				return true
			}
		}

		return false
	})

	// The trigger runs synchronously before the terminal transition, so once
	// WorkerFailed is observed the capture decision is final: no delayed
	// write can appear after this point (restarts are zero, worker is dead).
	time.Sleep(250 * time.Millisecond)

	if buf.Len() != 0 {
		t.Fatalf("gated trigger captured anyway: %d bytes", buf.Len())
	}
}

// NOT parallel: holds the single process-global flight recorder slot.
func TestEventConfig_FlightRecorder_DerivedWiringWinsOverHostOptions(t *testing.T) {
	recorderMu.lock()
	defer recorderMu.unlock()

	derivedBuf := &syncBuffer{}
	otherBuf := &syncBuffer{}

	rec, err := fr.New(fr.WithWriter(derivedBuf), fr.WithMinAge(time.Nanosecond))
	if err != nil {
		t.Fatalf("create derived recorder: %v", err)
	}

	err = rec.Start()
	if err != nil {
		t.Fatalf("start derived recorder: %v", err)
	}

	defer rec.Stop()

	// Deliberately NOT started: if the consumer HostOption wrongly won, the
	// snapshot would go here (a no-op on an unstarted recorder) instead of
	// the derived recorder, and the derived buffer would stay empty.
	other, err := fr.New(fr.WithWriter(otherBuf), fr.WithMinAge(time.Nanosecond))
	if err != nil {
		t.Fatalf("create other recorder: %v", err)
	}

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath:     t.TempDir() + "/test.db",
		FlightRecorder: rec,
		HostOptions: []projectionhost.HostOption{
			projectionhost.WithFlightRecorder(other, nil),
			projectionhost.WithMaxRestarts(0),
			projectionhost.WithBackoff(time.Nanosecond, time.Nanosecond),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"precedence-projection",
		func(_ context.Context, _ event.Event) error {
			return errors.New("terminal failure")
		},
		[]event.Type{"test.precedence"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.precedence")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "trace snapshot in derived recorder buffer", func() bool {
		return derivedBuf.Len() > 0
	})

	if otherBuf.Len() != 0 {
		t.Fatalf("consumer HostOption captured instead of derived wiring: %d bytes", otherBuf.Len())
	}
}
