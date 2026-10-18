package otel

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel"
)

// captureHandler collects slog records for assertions. Sequential tests
// only — no locking.
type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)

	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func (h *captureHandler) messages() []string {
	msgs := make([]string, 0, len(h.records))
	for _, r := range h.records {
		msgs = append(msgs, r.Message)
	}

	return msgs
}

func TestSetup_DiagnosticsRoutesThroughSlog(t *testing.T) {
	withRestoredGlobals(t)

	capture := &captureHandler{}
	logger := slog.New(capture)

	provider, err := Setup(
		WithService("diagnostics", "", ""),
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

	if msgs := capture.messages(); len(msgs) != 1 || msgs[0] != "otel export failed" {
		t.Fatalf("error handler routed %v, want exactly one `otel export failed`", msgs)
	}

	last := capture.records[len(capture.records)-1]

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

	otel.GetLogger().Info("sdk internal note")

	if msgs := capture.messages(); len(msgs) != 2 || msgs[1] != "sdk internal note" {
		t.Fatalf("SetLogger routed %v, want the SDK internal line second", msgs)
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

	otel.GetLogger().Info("sdk internal note")

	if msgs := capture.messages(); len(msgs) != 0 {
		t.Errorf("SetLogger captured %v, want none — the global sink must stay untouched", msgs)
	}
}
