package otel

// The real-capture proof for the flight-recorder bridge: an actual
// fr.Recorder (runtime/trace feeding a real snapshot write) drives the
// hook through the recorder's own plumbing — manual and trigger captures
// both — and the emitted metrics carry the recorder's real source and
// kind values, not the synthetic events the unit test feeds.

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	fr "github.com/larsartmann/go-flightrecorder"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// recorderMu serializes recorder tests: Go's runtime/trace allows only ONE
// active flight recorder per process (the frh module's tests use the same
// guard, but test binaries are per-package so each needs its own).
var recorderMu sync.Mutex

func TestFlightRecorderMetricsHook_RealCaptureCycle(t *testing.T) {
	recorderMu.Lock()
	defer recorderMu.Unlock()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = mp.Shutdown(t.Context()) }()

	hook := NewFlightRecorderMetricsHook(mp.Meter("fr-bridge-e2e"))

	sink := &bytes.Buffer{}

	recorder, err := fr.New(
		fr.WithWriter(sink),
		fr.WithMinAge(50*time.Millisecond),
		fr.WithMaxBytes(1<<20),
		fr.WithMetrics(hook),
	)
	if err != nil {
		t.Fatalf("fr.New: %v", err)
	}

	startErr := recorder.Start()
	if startErr != nil {
		t.Fatalf("recorder start: %v", startErr)
	}

	defer func() {
		recorder.Stop()
		_ = recorder.Close()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()

	snapshotErr := recorder.Snapshot(ctx)
	if snapshotErr != nil {
		t.Fatalf("manual snapshot: %v", snapshotErr)
	}

	if sink.Len() == 0 {
		t.Fatal("snapshot sink is empty, want real trace bytes")
	}

	recorder.Reset()

	fired := recorder.SnapshotIf(ctx, fr.TriggerContext{
		Kind:     "http.request",
		Type:     "slow.route",
		Duration: 250 * time.Millisecond,
	}, func(fr.TriggerContext) bool { return true })
	if !fired {
		t.Fatal("SnapshotIf did not fire the always-true trigger")
	}

	var data metricdata.ResourceMetrics

	collectErr := reader.Collect(ctx, &data)
	if collectErr != nil {
		t.Fatalf("collect: %v", collectErr)
	}

	counts, durationCount, durationSum := summarizeFrMetrics(t, data)

	if counts["manual"] != 1 {
		t.Errorf("snapshots_total[source=manual] = %d, want 1 (all: %v)", counts["manual"], counts)
	}

	if counts["trigger"] != 1 {
		t.Errorf("snapshots_total[source=trigger] = %d, want 1 (all: %v)", counts["trigger"], counts)
	}

	if durationCount != 2 {
		t.Errorf("duration observations = %g, want 2", durationCount)
	}

	if durationSum <= 0 {
		t.Errorf("duration sum = %g, want positive from real snapshot writes", durationSum)
	}
}

// summarizeFrMetrics folds the collected resource metrics into snapshot
// counts by source plus duration count and sum.
func summarizeFrMetrics(
	t *testing.T,
	data metricdata.ResourceMetrics,
) (map[string]int64, float64, float64) {
	t.Helper()

	counts := map[string]int64{}
	durationCount := float64(0)
	durationSum := float64(0)

	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch m.Name {
			case frSnapshotsTotal:
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok {
					t.Fatalf("snapshots metric is %T, want a sum", m.Data)
				}

				for _, point := range sum.DataPoints {
					source, _ := point.Attributes.Value("source")
					counts[source.AsString()] += point.Value
				}
			case frSnapshotDuration:
				hist, ok := m.Data.(metricdata.Histogram[float64])
				if !ok {
					t.Fatalf("duration metric is %T, want a histogram", m.Data)
				}

				for _, point := range hist.DataPoints {
					durationCount += float64(point.Count)
					durationSum += point.Sum
				}
			}
		}
	}

	return counts, durationCount, durationSum
}
