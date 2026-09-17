package cqrs

// The F104 end-to-end proof: a REAL projection-host cycle (command
// dispatched, event appended, projection worker processes it) drives the
// OTel metrics bridge — not the synthetic recorder calls the unit tests
// use. Pins that EventConfig.Metrics wiring actually reaches the host and
// that the attribute labels carry the real projection name and event type.

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projection/v4"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestEventService_OTelProjectionMetrics_RealProjectionCycle(t *testing.T) {
	t.Parallel()

	provider, reader := newRecordingMeter(t)

	recorder, err := NewOTelProjectionMetrics(provider.Meter("cqrs-e2e"))
	if err != nil {
		t.Fatalf("recorder: %v", err)
	}

	eventSvc := newFacadeServiceCfg(t, EventConfig{
		Driver:  memoryDriver,
		Metrics: recorder,
	})

	t.Cleanup(func() { _ = eventSvc.Shutdown(context.Background()) })

	const projectionName = "metrics-e2e-projection"

	processed := make(chan struct{}, 16)
	err = eventSvc.Host().Register(projection.NewProjection(
		projectionName,
		func(_ context.Context, _ event.Event) error {
			processed <- struct{}{}

			return nil
		},
		[]event.Type{"facade.bumped"},
	))
	if err != nil {
		t.Fatalf("register projection: %v", err)
	}

	err = eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("start projections: %v", err)
	}

	err = eventSvc.Dispatch(context.Background(), newFacadeCommand(t, id.NewStreamID()))
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	select {
	case <-processed:
	case <-time.After(10 * time.Second):
		t.Fatal("projection never processed the dispatched event")
	}

	var metrics []metricdata.Metrics

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		metrics = collectMetrics(t, reader)

		found := false

		for _, m := range metrics {
			found = found || m.Name == "cqrs.projection.event.count"
		}

		if found {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	events := findProjectionMetric(t, metrics, "cqrs.projection.event.count")
	counts, ok := events.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("event count metric is %T, want a sum", events.Data)
	}

	processedTotal := int64(0)
	seenProjection := false
	seenEventType := false

	for _, point := range counts.DataPoints {
		status, _ := point.Attributes.Value("cqrs.status")
		if status.AsString() != "processed" {
			continue
		}

		processedTotal += point.Value

		projectionAttr, _ := point.Attributes.Value(AttrProjectionName)
		seenProjection = seenProjection || projectionAttr.AsString() == projectionName

		typeAttr, _ := point.Attributes.Value(AttrEventType)
		seenEventType = seenEventType || typeAttr.AsString() == "facade.bumped"
	}

	if processedTotal < 1 {
		t.Fatalf("cqrs.projection.event.count[processed] = %d, want at least 1 from the real cycle", processedTotal)
	}

	if !seenProjection {
		t.Errorf("no processed datapoint carries cqrs.projection=%q (all: %v)", projectionName, counts.DataPoints)
	}

	if !seenEventType {
		t.Errorf("no processed datapoint carries %s=%q", AttrEventType, "facade.bumped")
	}
}
