package otel

import (
	"context"
	"log/slog"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
)

// installTelemetryDiagnostics routes the OTel SDK's self-diagnostics through
// slog instead of the stderr fallback: exporter failures (batch processor
// drops, reader errors) surface as grep-able `otel export failed` error
// lines, and SDK-internal logs (deprecation warnings, env-parse errors)
// route through the global otel.SetLogger sink, which is a no-op by default.
// The wiring follows the same last-write-wins globals contract as the
// provider registration around it.
func installTelemetryDiagnostics(logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.ErrorContext(context.Background(), "otel export failed", "error", err)
	}))

	otel.SetLogger(logr.FromSlogHandler(logger.Handler()))
}
