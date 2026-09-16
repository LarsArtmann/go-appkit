package otel

import (
	"context"

	fr "github.com/larsartmann/go-flightrecorder"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// frSnapshotsTotal is the stable metric name of the flight-recorder bridge.
// Part of the module's published metric-name contract.
const frSnapshotsTotal = "appkit_flightrecorder_snapshots_total"

// frSnapshotDuration is the stable histogram name of the bridge.
const frSnapshotDuration = "appkit_flightrecorder_snapshot_duration_seconds"

// NewFlightRecorderMetricsHook bridges the flight recorder's capture events
// into an OTel meter: every snapshot (manual, trigger, or async) increments
// `appkit_flightrecorder_snapshots_total{source,kind}` and records
// `appkit_flightrecorder_snapshot_duration_seconds`. Wire it at recorder
// construction:
//
//	rec, err := fr.New(fr.WithSnapshotDir(dir),
//	    fr.WithMetrics(appkitotel.NewFlightRecorderMetricsHook(meter)))
//
// A nil meter returns a nil-safe no-op hook, matching the recorder's
// optional-embedder pattern.
func NewFlightRecorderMetricsHook(meter metric.Meter) flightrecorderMetricsHook {
	if meter == nil {
		return func(fr.SnapshotEvent, error) {}
	}

	counter, err := meter.Int64Counter(
		frSnapshotsTotal,
		metric.WithDescription("Flight recorder snapshots by source and kind"),
	)
	if err != nil {
		return func(fr.SnapshotEvent, error) {}
	}

	duration, err := meter.Float64Histogram(
		frSnapshotDuration,
		metric.WithDescription("Snapshot write duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return func(fr.SnapshotEvent, error) {}
	}

	return func(event fr.SnapshotEvent, err error) {
		attrs := metric.WithAttributes(
			attribute.String("source", event.Source),
			attribute.String("kind", event.Kind),
		)

		counter.Add(context.Background(), 1, attrs)
		duration.Record(context.Background(), event.Duration.Seconds(), attrs)
	}
}

// flightrecorderMetricsHook aliases the recorder's hook type so the bridge
// stays assignable to fr.WithMetrics without importing fr at consumer call
// sites (the returned value IS a flightrecorder.MetricsHook).
type flightrecorderMetricsHook = func(fr.SnapshotEvent, error)
