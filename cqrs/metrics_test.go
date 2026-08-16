package cqrs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
)

// countingRecorder is a projectionhost.MetricsRecorder that counts every
// lifecycle event. It doubles as the consumer-side pattern: implement the
// interface, keep it concurrency-safe, never block.
type countingRecorder struct {
	processed    atomic.Int64
	errored      atomic.Int64
	deadLettered atomic.Int64
	restarted    atomic.Int64
	failed       atomic.Int64
	checkpoints  atomic.Int64
}

func (r *countingRecorder) EventProcessed(string, string, time.Duration) {
	r.processed.Add(1)
}

func (r *countingRecorder) EventErrored(string, string) { r.errored.Add(1) }

func (r *countingRecorder) EventDeadLettered(string, string) { r.deadLettered.Add(1) }

func (r *countingRecorder) WorkerRestarted(string) { r.restarted.Add(1) }

func (r *countingRecorder) WorkerFailed(string) { r.failed.Add(1) }

func (r *countingRecorder) CheckpointAdvanced(string, time.Duration) { r.checkpoints.Add(1) }

// exposition renders the counters in Prometheus text exposition format —
// the consumer-facing shape a /metrics endpoint would serve.
func (r *countingRecorder) exposition() string {
	return fmt.Sprintf(
		"# TYPE cqrs_projection_events_processed_total counter\n"+
			"cqrs_projection_events_processed_total %d\n"+
			"# TYPE cqrs_projection_events_errored_total counter\n"+
			"cqrs_projection_events_errored_total %d\n"+
			"# TYPE cqrs_projection_events_dead_lettered_total counter\n"+
			"cqrs_projection_events_dead_lettered_total %d\n",
		r.processed.Load(), r.errored.Load(), r.deadLettered.Load(),
	)
}

func TestEventConfig_Metrics_RecordsProjectionLifecycle(t *testing.T) {
	t.Parallel()

	rec := &countingRecorder{}

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		Metrics:    rec,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"metrics-projection",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.metrics"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.metrics")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "processed metric", func() bool { return rec.processed.Load() > 0 })
	waitFor(t, "checkpoint metric", func() bool { return rec.checkpoints.Load() > 0 })

	if got := rec.errored.Load(); got != 0 {
		t.Errorf("errored = %d, want 0", got)
	}
}

func TestEventConfig_Metrics_RecordsErrors(t *testing.T) {
	t.Parallel()

	rec := &countingRecorder{}

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		Metrics:    rec,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	boom := errors.New("handler boom")

	proj := projection.NewProjection(
		"error-metrics-projection",
		func(_ context.Context, _ event.Event) error { return boom },
		[]event.Type{"test.metrics.fail"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.metrics.fail")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "errored metric", func() bool { return rec.errored.Load() > 0 })
}

// TestEventConfig_Metrics_HandlerEndpoint proves the consumer-facing
// metrics path end to end: lifecycle events flow into the recorder while
// a /metrics endpoint serves them over HTTP.
func TestEventConfig_Metrics_HandlerEndpoint(t *testing.T) {
	t.Parallel()

	rec := &countingRecorder{}

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
		Metrics:    rec,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	proj := projection.NewProjection(
		"endpoint-projection",
		func(_ context.Context, _ event.Event) error { return nil },
		[]event.Type{"test.endpoint"},
	)

	err = eventSvc.Host().Register(proj)
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	appendTestEvent(t, eventSvc, "test.endpoint")

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	waitFor(t, "processed metric", func() bool { return rec.processed.Load() > 0 })

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_ = fmt.Fprint(w, rec.exposition())
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, srv.URL+"/metrics", nil)
	if err != nil {
		t.Fatalf("build /metrics request: %v", err)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("get /metrics: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	got := string(body)
	want := fmt.Sprintf("cqrs_projection_events_processed_total %d", rec.processed.Load())

	if !strings.Contains(got, want) {
		t.Fatalf("body = %q, want it to contain %q", got, want)
	}
}
