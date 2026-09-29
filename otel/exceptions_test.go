package otel

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// quietLogger keeps panic-recovery log lines out of test output.
func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newPanicServer composes the documented wiring: the span middleware
// outermost, Recovery innermost (appkit's ExtraMiddlewares position).
func newPanicServer(t *testing.T, tp *sdktrace.TracerProvider, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle("GET /boom", handler)

	var chain http.Handler = Recovery(quietLogger())(mux)
	if tp != nil {
		chain = Middleware(WithTracerProvider(tp))(chain)
	}

	server := httptest.NewServer(chain)
	t.Cleanup(server.Close)

	return server
}

// exceptionEvent finds the first `exception` event on a recorded span.
func exceptionEvent(span tracetest.SpanStub) (sdktrace.Event, bool) {
	for _, event := range span.Events {
		if event.Name == string(semconv.ExceptionEventName) {
			return event, true
		}
	}

	return sdktrace.Event{}, false
}

func TestRecovery_RecordsExceptionEventForErrorPanic(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newPanicServer(t, tp, func(_ http.ResponseWriter, _ *http.Request) {
		panic(errors.New("database connection refused"))
	})

	resp, err := http.Get(server.URL + "/boom") //nolint:noctx // single-shot panic scenario
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	event, found := exceptionEvent(spans[0])
	if !found {
		t.Fatalf("span %q has no exception event: %+v", spans[0].Name, spans[0].Events)
	}

	attrs := attribute.NewSet(event.Attributes...)
	want := map[string]string{
		"exception.type":      "*errors.errorString",
		"exception.message":   "database connection refused",
		"exception.stacktrace": "",
	}
	for key := range want {
		if _, ok := attrs.Value(attribute.Key(key)); !ok {
			t.Errorf("exception event misses %s: %v", key, event.Attributes)
		}
	}

	if stack, _ := attrs.Value(attribute.Key("exception.stacktrace")); !strings.Contains(stack.AsString(), "goroutine") {
		t.Errorf("exception.stacktrace = %q, want a Go stack trace", stack.AsString())
	}

	if spans[0].Status.Code != codes.Error {
		t.Errorf("span status = %v, want Error after a panic", spans[0].Status.Code)
	}
}

func TestRecovery_NonErrorPanicRecordsTypedEvent(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newPanicServer(t, tp, func(_ http.ResponseWriter, _ *http.Request) {
		panic("kaboom") //nolint:goerr113 // the point is a non-error panic value
	})

	resp, err := http.Get(server.URL + "/boom") //nolint:noctx // single-shot panic scenario
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	_ = resp.Body.Close()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	event, found := exceptionEvent(spans[0])
	if !found {
		t.Fatalf("span has no exception event: %+v", spans[0].Events)
	}

	attrs := attribute.NewSet(event.Attributes...)
	if got, ok := attrs.Value(attribute.Key("exception.type")); !ok || got.AsString() != "string" {
		t.Errorf("exception.type = %q (found=%v), want the value's type name", got.AsString(), ok)
	}

	if got, ok := attrs.Value(attribute.Key("exception.message")); !ok || got.AsString() != "kaboom" {
		t.Errorf("exception.message = %q (found=%v), want kaboom", got.AsString(), ok)
	}

	if spans[0].Status.Description != "kaboom" {
		t.Errorf("span status description = %q, want kaboom", spans[0].Status.Description)
	}
}

// TestRecovery_AbortHandlerRepanics: the net/http sentinel keeps its
// contract — no exception recording, no 500 write, the panic continues to
// the server's silent connection-abort handling.
func TestRecovery_AbortHandlerRepanics(t *testing.T) {
	t.Parallel()

	repanicked := func() (repanicked bool) {
		defer func() {
			if rec := recover(); rec != nil {
				repanicked = errors.Is(rec.(error), http.ErrAbortHandler) //nolint:errorlint // test-local sentinel probe
			}
		}()

		handler := Recovery(quietLogger())(http.NotFoundHandler())
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))

		return false
	}()

	if !repanicked {
		t.Error("http.ErrAbortHandler was swallowed, want it re-panicked unchanged")
	}
}

// TestRecovery_DegradesWithoutRecordingSpan: no provider, no crash — the
// middleware falls back to plain Recovery behavior (log + 500).
func TestRecovery_DegradesWithoutRecordingSpan(t *testing.T) {
	t.Parallel()

	server := newPanicServer(t, nil, func(_ http.ResponseWriter, _ *http.Request) {
		panic(errors.New("no telemetry here"))
	})

	resp, err := http.Get(server.URL + "/boom") //nolint:noctx // single-shot panic scenario
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 even without a span", resp.StatusCode)
	}
}

// TestRecovery_NilLoggerFallsBackToDefault guards the low-code path: nil
// must not panic the panic handler.
func TestRecovery_NilLoggerFallsBackToDefault(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /boom", func(_ http.ResponseWriter, _ *http.Request) {
		panic(errors.New("nil logger"))
	})

	server := httptest.NewServer(Recovery(nil)(mux))
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/boom") //nolint:noctx // single-shot panic scenario
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecordError_Helper(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)

	ctx, span := tp.Tracer("test").Start(t.Context(), "handler")
	RecordError(ctx, errors.New("checkout declined"))
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	event, found := exceptionEvent(spans[0])
	if !found {
		t.Fatalf("RecordError produced no exception event: %+v", spans[0].Events)
	}

	attrs := attribute.NewSet(event.Attributes...)
	if got, ok := attrs.Value(attribute.Key("exception.message")); !ok || got.AsString() != "checkout declined" {
		t.Errorf("exception.message = %q (found=%v), want checkout declined", got.AsString(), ok)
	}

	if spans[0].Status.Code != codes.Error {
		t.Errorf("span status = %v, want Error", spans[0].Status.Code)
	}
}

func TestRecordError_NilErrorAndNoSpanAreNoOps(t *testing.T) {
	t.Parallel()

	RecordError(t.Context(), nil) // must not panic

	tp, exporter := newRecordingProvider(t)
	_, span := tp.Tracer("test").Start(t.Context(), "clean")

	RecordError(t.Context(), errors.New("wrong context — no span")) // noop span: not recording
	span.End()

	if spans := exporter.GetSpans(); len(spans) != 1 {
		t.Fatalf("noop-span RecordError exported %d spans, want only the clean one", len(spans))
	}

	if spans := exporter.GetSpans(); len(spans) == 1 && len(spans[0].Events) != 0 {
		t.Errorf("noop-span RecordError added events: %+v", spans[0].Events)
	}
}
