package cqrs

import (
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func newRecordingMeter(t *testing.T) (*sdkmetric.MeterProvider, *sdkmetric.ManualReader) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

	return provider, reader
}

func collectMetrics(t *testing.T, reader *sdkmetric.ManualReader) []metricdata.Metrics {
	t.Helper()

	var data metricdata.ResourceMetrics

	err := reader.Collect(t.Context(), &data)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	metrics := []metricdata.Metrics{}
	for _, scope := range data.ScopeMetrics {
		metrics = append(metrics, scope.Metrics...)
	}

	return metrics
}

func findProjectionMetric(t *testing.T, metrics []metricdata.Metrics, name string) metricdata.Metrics {
	t.Helper()

	for _, m := range metrics {
		if m.Name == name {
			return m
		}
	}

	t.Fatalf("metric %q not found", name)

	return metricdata.Metrics{}
}

func TestOTelProjectionMetrics_RecordsAllLifecycleEvents(t *testing.T) {
	t.Parallel()

	provider, reader := newRecordingMeter(t)

	recorder, err := NewOTelProjectionMetrics(provider.Meter("cqrs-test"))
	if err != nil {
		t.Fatalf("recorder: %v", err)
	}

	recorder.EventProcessed("tasks", "task.created", 25*time.Millisecond)
	recorder.EventErrored("tasks", "task.renamed")
	recorder.EventDeadLettered("tasks", "task.renamed")
	recorder.WorkerRestarted("tasks")
	recorder.WorkerFailed("tasks")
	recorder.CheckpointAdvanced("tasks", 300*time.Millisecond)

	metrics := collectMetrics(t, reader)

	events := findProjectionMetric(t, metrics, "cqrs.projection.event.count")
	counts, ok := events.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("event count metric is %T, want a sum", events.Data)
	}

	statuses := map[string]int64{}
	for _, point := range counts.DataPoints {
		status, _ := point.Attributes.Value("cqrs.status")
		statuses[status.AsString()] = point.Value
	}

	for status, want := range map[string]int64{
		"processed":     1,
		"errored":       1,
		"dead_lettered": 1,
	} {
		if statuses[status] != want {
			t.Errorf(
				"cqrs.projection.event.count[%s] = %d, want %d (all: %v)",
				status,
				statuses[status],
				want,
				statuses,
			)
		}
	}

	duration := findProjectionMetric(t, metrics, "cqrs.projection.event.duration")
	hist, ok := duration.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("duration metric is %T, want a histogram", duration.Data)
	}

	if len(hist.DataPoints) != 1 || hist.DataPoints[0].Count != 1 || hist.DataPoints[0].Sum != 25 {
		t.Errorf("duration histogram = %+v, want one observation summing to 25ms", hist.DataPoints)
	}

	workers := findProjectionMetric(t, metrics, "cqrs.projection.worker.count")
	workerCounts, ok := workers.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("worker metric is %T, want a sum", workers.Data)
	}

	workerStatuses := map[string]int64{}
	for _, point := range workerCounts.DataPoints {
		status, _ := point.Attributes.Value("cqrs.status")
		workerStatuses[status.AsString()] = point.Value
	}

	if workerStatuses["restarted"] != 1 || workerStatuses["failed"] != 1 {
		t.Errorf("worker counts = %v, want one restarted and one failed", workerStatuses)
	}

	lag := findProjectionMetric(t, metrics, "cqrs.projection.checkpoint.lag")
	lagHist, ok := lag.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("lag metric is %T, want a histogram", lag.Data)
	}

	if len(lagHist.DataPoints) != 1 || lagHist.DataPoints[0].Sum != 300 {
		t.Errorf("lag histogram = %+v, want one observation summing to 300ms", lagHist.DataPoints)
	}
}

func TestOTelProjectionMetrics_AttributesCarryProjectionAndType(t *testing.T) {
	t.Parallel()

	provider, reader := newRecordingMeter(t)

	recorder, err := NewOTelProjectionMetrics(provider.Meter("cqrs-test"))
	if err != nil {
		t.Fatalf("recorder: %v", err)
	}

	recorder.EventProcessed("orders", "order.placed", time.Millisecond)

	metrics := collectMetrics(t, reader)
	events := findProjectionMetric(t, metrics, "cqrs.projection.event.count")
	counts := events.Data.(metricdata.Sum[int64]) //nolint:forcetypeassert // shape asserted by the sibling test

	point := counts.DataPoints[0]

	for key, want := range map[string]string{
		"cqrs.projection.name": "orders",
		"cqrs.event.type":      "order.placed",
		"cqrs.status":          "processed",
	} {
		got, _ := point.Attributes.Value(attribute.Key(key))
		if got.AsString() != want {
			t.Errorf("attribute %s = %q, want %q", key, got.AsString(), want)
		}
	}
}
