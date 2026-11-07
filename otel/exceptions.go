package otel

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/larsartmann/httputil"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// Recovery returns middleware that catches panics in downstream handlers,
// turns them into OpenTelemetry exceptions on the active span, and returns
// 500 Internal Server Error. It is httputil.RecoveryWithConfig plus one
// hook: OnPanic records the recovered panic as an `exception` span event
// (type, message, stacktrace) and error span status — exactly what backends
// like SigNoz index into their dedicated exceptions view. Log line, 500
// response body, and the http.ErrAbortHandler re-panic contract are
// httputil's own (the library re-panics abort panics BEFORE the hook, so
// aborts produce no span event — same as before the delegation).
//
// Place it inside the span-producing middleware, in the innermost recovery
// position. In an appkit service the default stack's own Recovery sits
// outer, so this one catches first:
//
//	cfg := appkit.DefaultServiceConfig()
//	cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
//	cfg.ExtraMiddlewares = []httputil.Middleware{appkitotel.Recovery(logger)}
//
// A nil logger falls back to slog.Default() inside RecoveryWithConfig.
// Without an active recording span the middleware degrades gracefully to
// plain httputil.Recovery behavior.
func Recovery(logger *slog.Logger) httputil.Middleware {
	return httputil.RecoveryWithConfig(httputil.RecoveryConfig{
		Logger: logger,
		OnPanic: func(req *http.Request, recovered any) {
			recordPanic(req.Context(), recovered)
		},
	})
}

// RecordError records err on the span active in ctx as a semantic-convention
// exception event (with stacktrace) and marks the span status error. Call it
// wherever a handler handles an error instead of panicking — SigNoz's
// exceptions view and error-rate queries read exactly this span data:
//
//	if err != nil {
//	    appkitotel.RecordError(r.Context(), err)
//	    http.Error(w, "checkout failed", http.StatusInternalServerError)
//	    return
//	}
//
// A nil error or a non-recording span is a no-op.
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())
}

// recordPanic converts a recovered panic value into span exception data:
// error values go through RecordError (exception event + stacktrace + error
// status), anything else becomes an exception event carrying the value's
// type and string form. No-op without a recording span.
func recordPanic(ctx context.Context, rec any) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	if err, ok := rec.(error); ok {
		span.RecordError(err, trace.WithStackTrace(true))
	} else {
		span.AddEvent(string(semconv.ExceptionEventName), trace.WithAttributes(
			semconv.ExceptionTypeKey.String(fmt.Sprintf("%T", rec)),
			semconv.ExceptionMessageKey.String(fmt.Sprint(rec)),
			semconv.ExceptionStacktraceKey.String(string(debug.Stack())),
		))
	}

	span.SetStatus(codes.Error, fmt.Sprint(rec))
}
