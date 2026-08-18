package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler returns an slog.Handler that stamps trace_id and span_id
// onto every record whose context carries a valid span — turning log lines
// into trace-correlated events in backends that parse them (Grafana,
// Datadog, Loki, Tempo). Records logged without a span in their context
// pass through unchanged; no placeholder attributes are added.
//
// For the correlation to appear, log with the request context:
//
//	logger := slog.New(appkitotel.TraceHandler(slog.NewJSONHandler(os.Stdout, nil)))
//	logger.InfoContext(r.Context(), "order accepted")
//
// Note: httputil's built-in Logging middleware logs without a context, so
// its request-completion line stays uncorrelated; handler-level logs —
// where the business context lives — are the ones that carry the IDs.
func TraceHandler(next slog.Handler) slog.Handler {
	return traceHandler{next: next}
}

type traceHandler struct {
	next slog.Handler
}

func (h traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h traceHandler) Handle(ctx context.Context, record slog.Record) error {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		record.AddAttrs(
			slog.String(AttrTraceID, spanCtx.TraceID().String()),
			slog.String(AttrSpanID, spanCtx.SpanID().String()),
		)
	}

	//nolint:wrapcheck // decorator delegates to the wrapped handler; wrapping would obscure the delegate's error
	return h.next.Handle(ctx, record)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{next: h.next.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{next: h.next.WithGroup(name)}
}

// Semantic attribute keys emitted on correlated log records.
const (
	// AttrTraceID is the W3C trace identifier of the active span.
	AttrTraceID = "trace_id"
	// AttrSpanID is the W3C span identifier of the active span.
	AttrSpanID = "span_id"
)

// noSpanID is the placeholder logged when no valid span is in the context.
const noSpanID = "none"

// TraceIDFromContext extracts the trace ID from the context. Returns "none"
// if no span is active.
func TraceIDFromContext(ctx context.Context) string {
	return spanIDOrNone(ctx, func(spanCtx trace.SpanContext) string {
		return spanCtx.TraceID().String()
	})
}

// SpanIDFromContext extracts the span ID from the context. Returns "none"
// if no span is active.
func SpanIDFromContext(ctx context.Context) string {
	return spanIDOrNone(ctx, func(spanCtx trace.SpanContext) string {
		return spanCtx.SpanID().String()
	})
}

// ContextLogger returns an *slog.Logger that includes trace_id and span_id
// from the given context. If no span is active, fields are set to "none".
// Prefer [TraceHandler] for automatic correlation; ContextLogger suits
// call sites that construct loggers explicitly.
func ContextLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	return logger.With(
		slog.String(AttrTraceID, TraceIDFromContext(ctx)),
		slog.String(AttrSpanID, SpanIDFromContext(ctx)),
	)
}

// spanIDOrNone extracts the active span context and returns either the
// caller-selected ID string (when the span is valid) or "none". Shared by
// TraceIDFromContext and SpanIDFromContext to keep the validity check and
// fallback string in one place.
func spanIDOrNone(ctx context.Context, pick func(trace.SpanContext) string) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if !spanCtx.IsValid() {
		return noSpanID
	}

	return pick(spanCtx)
}
