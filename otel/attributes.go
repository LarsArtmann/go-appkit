package otel

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// ServiceResourceAttributes returns the resource attributes identifying the
// service in all telemetry: service.name (always), service.version and
// service.instance.id (when non-empty).
func ServiceResourceAttributes(serviceName, serviceVersion, instanceID string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
	}

	if serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(serviceVersion))
	}

	if instanceID != "" {
		attrs = append(attrs, semconv.ServiceInstanceID(instanceID))
	}

	return attrs
}

// NewTextMapPropagator returns a composite propagator combining W3C trace
// context and W3C baggage propagation — the recommended default for HTTP
// services: trace context propagates spans, baggage propagates correlation
// attributes across service boundaries. Setup registers it globally; this
// accessor exists for consumers wiring their own provider stack.
func NewTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
