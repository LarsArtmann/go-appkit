package otel

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newTestTracer(t *testing.T) (context.Context, func()) {
	t.Helper()

	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = tp.Shutdown(t.Context()) })

	ctx, span := tp.Tracer("test").Start(t.Context(), "correlated")

	return ctx, func() { span.End() }
}

func TestTraceHandler_StampsTraceAndSpanIDs(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(TraceHandler(slog.NewJSONHandler(buf, nil)))

	ctx, end := newTestTracer(t)

	logger.InfoContext(ctx, "business event")

	wantTraceID := TraceIDFromContext(ctx)
	wantSpanID := SpanIDFromContext(ctx)
	end()

	out := buf.String()
	if !strings.Contains(out, `"trace_id":"`+wantTraceID+`"`) {
		t.Errorf("log line misses correlated trace_id %q: %s", wantTraceID, out)
	}

	if !strings.Contains(out, `"span_id":"`+wantSpanID+`"`) {
		t.Errorf("log line misses correlated span_id %q: %s", wantSpanID, out)
	}
}

func TestTraceHandler_LeavesUncorrelatedRecordsUntouched(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(TraceHandler(slog.NewJSONHandler(buf, nil)))

	logger.Info("no context here")

	if out := buf.String(); strings.Contains(out, "trace_id") || strings.Contains(out, "span_id") {
		t.Errorf("uncorrelated record gained trace attributes: %s", out)
	}
}

func TestTraceHandler_PreservesWithAttrsAndWithGroup(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(TraceHandler(slog.NewJSONHandler(buf, nil))).
		With(slog.String("service", "test-svc")).
		WithGroup("request")

	ctx, end := newTestTracer(t)

	logger.InfoContext(ctx, "grouped", slog.String("path", "/orders"))

	wantTraceID := TraceIDFromContext(ctx)
	end()

	out := buf.String()
	for _, want := range []string{
		`"service":"test-svc"`,
		wantTraceID,
		`"request":{"path":"/orders"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("decorated logger output misses %q: %s", want, out)
		}
	}
}

func TestTraceAndSpanIDFromContext_NoneWithoutSpan(t *testing.T) {
	t.Parallel()

	if got := TraceIDFromContext(context.Background()); got != "none" {
		t.Errorf("TraceIDFromContext = %q, want none", got)
	}

	if got := SpanIDFromContext(context.Background()); got != "none" {
		t.Errorf("SpanIDFromContext = %q, want none", got)
	}
}

func TestContextLogger_CarriesIDs(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))

	ctx, end := newTestTracer(t)

	ContextLogger(logger, ctx).Info("via context logger")

	wantTraceID := TraceIDFromContext(ctx)
	end()

	if !strings.Contains(buf.String(), `"trace_id":"`+wantTraceID+`"`) {
		t.Errorf("ContextLogger output misses trace_id %q: %s", wantTraceID, buf.String())
	}
}
