package otel

import (
	"testing"
	"time"

	fr "github.com/larsartmann/go-flightrecorder"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// TestFlightRecorderMetricsHook_Records pins the bridge: a snapshot event
// fed to the hook lands as the documented metric names with source/kind
// attributes on the meter's reader.
func TestFlightRecorderMetricsHook_Records(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = mp.Shutdown(t.Context()) }()

	meter := mp.Meter("fr-bridge-test")

	hook := NewFlightRecorderMetricsHook(meter)
	hook(fr.SnapshotEvent{
		Duration: 25 * time.Millisecond,
		Bytes:    4096,
		Source:   "trigger",
		Kind:     "http.request",
	}, nil)

	var data metricdata.ResourceMetrics

	collectErr := reader.Collect(t.Context(), &data)
	if collectErr != nil {
		t.Fatalf("collect: %v", collectErr)
	}

	foundCount, foundDuration := false, false

	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch m.Name {
			case frSnapshotsTotal:
				foundCount = true
			case frSnapshotDuration:
				foundDuration = true
			}
		}
	}

	if !foundCount {
		t.Error("appkit_flightrecorder_snapshots_total not recorded")
	}

	if !foundDuration {
		t.Error("appkit_flightrecorder_snapshot_duration_seconds not recorded")
	}
}

// TestFlightRecorderMetricsHook_NilMeterIsNoOp pins the optional-embedder
// pattern: a nil meter yields a hook that can be called safely.
func TestFlightRecorderMetricsHook_NilMeterIsNoOp(t *testing.T) {
	t.Parallel()

	hook := NewFlightRecorderMetricsHook(nil)

	hook(fr.SnapshotEvent{Source: "manual"}, nil) // must not panic
}
