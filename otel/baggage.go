package otel

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
)

// CorrelationIDKey is the baggage member key the correlation helpers read
// and write. It is part of the grep-able contract: whatever the caller put
// in `correlation_id` baggage travels across every service boundary this
// fleet's propagators touch.
const CorrelationIDKey = "correlation_id"

// CorrelationID returns the correlation ID carried in the context's baggage,
// or "" when absent. Baggage arrives through the W3C baggage propagation the
// [Middleware] performs on inbound requests; the value is whatever the
// CALLER set — this helper reads, never invents.
func CorrelationID(ctx context.Context) string {
	return baggage.FromContext(ctx).Member(CorrelationIDKey).Value()
}

// WithCorrelationID returns a copy of ctx carrying id as the correlation
// baggage member — downstream services receive it through [Transport]'s W3C
// baggage injection. An empty id returns ctx unchanged, and a value the W3C
// baggage grammar rejects (non-US-ASCII, control characters) leaves ctx
// unchanged rather than corrupting the header: this helper sets, never
// invents. An existing member under the key is replaced.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}

	member, err := baggage.NewMember(CorrelationIDKey, id)
	if err != nil {
		return ctx
	}

	updated, err := baggage.FromContext(ctx).SetMember(member)
	if err != nil {
		return ctx
	}

	return baggage.ContextWithBaggage(ctx, updated)
}
