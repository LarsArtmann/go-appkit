package otel

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// captureHandler collects slog records for assertions. It is installed as
// the process-global OTel log sink, so it must tolerate concurrent writes
// from SDK internals (export goroutines, provider constructors).
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.records = append(h.records, r)

	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func (h *captureHandler) messages() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	msgs := make([]string, 0, len(h.records))
	for _, r := range h.records {
		msgs = append(msgs, r.Message)
	}

	return msgs
}

func (h *captureHandler) contains(msg string) bool {
	for _, m := range h.messages() {
		if m == msg {
			return true
		}
	}

	return false
}

func TestSetup_DiagnosticsRoutesThroughSlog(t *testing.T) {
	withRestoredGlobals(t)

	capture := &captureHandler{}
	logger := slog.New(capture)

	provider, err := Setup(
		WithService("diagnostics", "", ""),
		WithSpanExporter(&tracetest.InMemoryExporter{}),
		WithDiagnosticsLogger(logger),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	defer func() {
		if err := provider.Shutdown(t.Context()); err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	}()

	otel.GetErrorHandler().Handle(errors.New("boom"))

	found := -1

	for i, m := range capture.messages() {
		if m == "otel export failed" {
			found = i

			break
		}
	}

	if found < 0 {
		t.Fatalf("error handler routed %v, want an `otel export failed` line", capture.messages())
	}

	capture.mu.Lock()
	last := capture.records[found]
	capture.mu.Unlock()

	if last.Level != slog.LevelError {
		t.Errorf("export failure logged at %v, want error level", last.Level)
	}

	var errAttr string

	last.Attrs(func(a slog.Attr) bool {
		if a.Key == "error" {
			errAttr = a.Value.String()
		}

		return true
	})

	if errAttr != "boom" {
		t.Errorf("export failure carried error=%q, want boom", errAttr)
	}
}

func TestSetup_SetLoggerRoutesSDKInternalsThroughSlog(t *testing.T) {
	withRestoredGlobals(t)

	capture := &captureHandler{}

	provider, err := Setup(
		WithService("sdk-logs", "", ""),
		WithSpanExporter(&tracetest.InMemoryExporter{}),
		WithDiagnosticsLogger(slog.New(capture)),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	defer func() {
		if err := provider.Shutdown(t.Context()); err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	}()

	_, span := otel.GetTracerProvider().Tracer("diagnostics-test").Start(t.Context(), "op")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	if !capture.contains("exporting spans") {
		t.Fatalf("SDK internal logs captured %v, want the batch processor's `exporting spans` line", capture.messages())
	}
}

func TestSetup_WithoutGlobalRegistrationSkipsDiagnostics(t *testing.T) {
	withRestoredGlobals(t)

	sentinelFired := false
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(error) { sentinelFired = true }))

	capture := &captureHandler{}

	provider, err := Setup(
		WithService("isolated", "", ""),
		WithDiagnosticsLogger(slog.New(capture)),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	defer func() {
		if err := provider.Shutdown(t.Context()); err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	}()

	otel.GetErrorHandler().Handle(errors.New("boom"))

	if !sentinelFired {
		t.Error("WithoutGlobalRegistration must leave the global error handler untouched")
	}

	if msgs := capture.messages(); len(msgs) != 0 {
		t.Errorf("capture got %v, want none — diagnostics must not route through the passed logger when global registration is skipped", msgs)
	}
}
