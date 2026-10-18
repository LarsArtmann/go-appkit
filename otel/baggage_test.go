package otel

import (
	"context"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel/propagation"
)

func TestCorrelationID_RoundTripsThroughW3CBaggage(t *testing.T) {
	t.Parallel()

	ctx := WithCorrelationID(context.Background(), "req-abc-123")

	carrier := http.Header{}
	NewTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(carrier))

	ctx2 := NewTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(carrier))

	if got := CorrelationID(ctx2); got != "req-abc-123" {
		t.Fatalf("correlation ID across the W3C carrier = %q, want req-abc-123", got)
	}
}

func TestCorrelationID_AbsentAndInvalidAreNoOps(t *testing.T) {
	t.Parallel()

	if got := CorrelationID(context.Background()); got != "" {
		t.Fatalf("absent baggage returned %q, want empty", got)
	}

	base := context.Background()

	if got := WithCorrelationID(base, ""); got != base {
		t.Fatal("empty ID must return the context unchanged — the helper sets, never invents")
	}

	if got := WithCorrelationID(base, "bad\x00value"); got != base {
		t.Fatal("W3C-invalid value must return the context unchanged, not corrupt the header")
	}
}

func TestCorrelationID_ReplacesExistingMember(t *testing.T) {
	t.Parallel()

	ctx := WithCorrelationID(context.Background(), "first")
	ctx = WithCorrelationID(ctx, "second")

	if got := CorrelationID(ctx); got != "second" {
		t.Fatalf("re-set correlation ID = %q, want second", got)
	}
}
