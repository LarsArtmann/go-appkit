package cqrs

import (
	"context"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Attribute keys on projection telemetry. Values follow go-cqrs-lite's
// cqrs.* semantic-attribute conventions so HTTP (go-appkit/otel) and
// projection dashboards query one schema.
const (
	AttrProjectionName = "cqrs.projection.name"
	AttrEventType      = "cqrs.event.type"
	AttrStatus         = "cqrs.status"
)

// Projection lifecycle statuses recorded on cqrs.projection.event.count and
// cqrs.projection.worker.count.
const (
	StatusProcessed    = "processed"
	StatusErrored      = "errored"
	StatusDeadLettered = "dead_lettered"
	StatusRestarted    = "restarted"
	StatusFailed       = "failed"
)

// OTelProjectionMetrics implements projectionhost.MetricsRecorder on OTel
// instruments, closing the metrics path for EventConfig.Metrics without
// forcing any exporter on this module: the OTel API is interface-only here.
//
// Instruments:
//
//	cqrs.projection.event.count     counter    (projection, event_type, status)
//	cqrs.projection.event.duration  histogram  (projection, event_type) — ms
//	cqrs.projection.worker.count    counter    (projection, status)
//	cqrs.projection.checkpoint.lag  histogram  (projection) — ms
//
// Create via [NewOTelProjectionMetrics] and pass it as EventConfig.Metrics.
// Methods are safe for concurrent use and never block (OTel record calls are
// synchronous in-process enqueue operations).
type OTelProjectionMetrics struct {
	events      metric.Int64Counter
	eventTime   metric.Float64Histogram
	workers     metric.Int64Counter
	checkpointL metric.Float64Histogram
}

var _ projectionhost.MetricsRecorder = (*OTelProjectionMetrics)(nil)

// NewOTelProjectionMetrics builds the recorder from any OTel meter — e.g.
// otel.GetMeterProvider().Meter("myapp") after appkitotel.Setup, or an
// explicitly constructed provider's meter in tests.
//
// If the meter is invalid (duplicate instrument names with conflicting
// types elsewhere in the process), an error surfaces here rather than
// silently dropping metrics.
func NewOTelProjectionMetrics(meter metric.Meter) (*OTelProjectionMetrics, error) {
	events, err := meter.Int64Counter(
		"cqrs.projection.event.count",
		metric.WithDescription("Projection lifecycle events by outcome"),
	)
	if err != nil {
		return nil, err //nolint:wrapcheck // OTel API error passes through with instrument context
	}

	eventTime, err := meter.Float64Histogram(
		"cqrs.projection.event.duration",
		metric.WithDescription("Successful projection handling duration"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err //nolint:wrapcheck // OTel API error passes through with instrument context
	}

	workers, err := meter.Int64Counter(
		"cqrs.projection.worker.count",
		metric.WithDescription("Projection worker restarts and terminal failures"),
	)
	if err != nil {
		return nil, err //nolint:wrapcheck // OTel API error passes through with instrument context
	}

	checkpointL, err := meter.Float64Histogram(
		"cqrs.projection.checkpoint.lag",
		metric.WithDescription("Event age at checkpoint persistence"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err //nolint:wrapcheck // OTel API error passes through with instrument context
	}

	return &OTelProjectionMetrics{
		events:      events,
		eventTime:   eventTime,
		workers:     workers,
		checkpointL: checkpointL,
	}, nil
}

// EventProcessed records a successful handling with its duration.
func (m *OTelProjectionMetrics) EventProcessed(projectionName, eventType string, duration time.Duration) {
	attrs := projectionAttrs(projectionName, eventType)

	m.events.Add(context.Background(), 1, metric.WithAttributes(append(attrs, statusAttr(StatusProcessed))...))

	m.eventTime.Record(context.Background(), float64(duration.Milliseconds()),
		metric.WithAttributes(attrs...),
	)
}

// EventErrored records a failed handling attempt (pre-DLQ).
func (m *OTelProjectionMetrics) EventErrored(projectionName, eventType string) {
	m.events.Add(context.Background(), 1, metric.WithAttributes(
		append(projectionAttrs(projectionName, eventType), statusAttr(StatusErrored))...,
	))
}

// EventDeadLettered records a move to the dead-letter queue.
func (m *OTelProjectionMetrics) EventDeadLettered(projectionName, eventType string) {
	m.events.Add(context.Background(), 1, metric.WithAttributes(
		append(projectionAttrs(projectionName, eventType), statusAttr(StatusDeadLettered))...,
	))
}

// WorkerRestarted records a crash-recovery restart.
func (m *OTelProjectionMetrics) WorkerRestarted(projectionName string) {
	m.workers.Add(context.Background(), 1, metric.WithAttributes(
		projectionAttr(projectionName), statusAttr(StatusRestarted),
	))
}

// WorkerFailed records a terminal worker failure (no further restarts).
func (m *OTelProjectionMetrics) WorkerFailed(projectionName string) {
	m.workers.Add(context.Background(), 1, metric.WithAttributes(
		projectionAttr(projectionName), statusAttr(StatusFailed),
	))
}

// CheckpointAdvanced records event age at checkpoint persistence.
func (m *OTelProjectionMetrics) CheckpointAdvanced(projectionName string, lag time.Duration) {
	m.checkpointL.Record(context.Background(), float64(lag.Milliseconds()),
		metric.WithAttributes(projectionAttr(projectionName)),
	)
}

func projectionAttr(name string) attribute.KeyValue {
	return attribute.String(AttrProjectionName, name)
}

func projectionAttrs(name, eventType string) []attribute.KeyValue {
	return []attribute.KeyValue{
		projectionAttr(name),
		attribute.String(AttrEventType, eventType),
	}
}

func statusAttr(status string) attribute.KeyValue {
	return attribute.String(AttrStatus, status)
}
