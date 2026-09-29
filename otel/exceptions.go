package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/larsartmann/httputil"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Recovery returns middleware that catches panics in downstream handlers,
// turns them into OpenTelemetry exceptions on the active span, and returns
// 500 Internal Server Error. It is httputil.Recovery upgraded for
// observability: same log line, same response, plus an `exception` span
// event (type, message, stacktrace) and error span status — exactly what
// backends like SigNoz index into their dedicated exceptions view.
//
// Place it inside the span-producing middleware, in the innermost recovery
// position. In an appkit service the default stack's own Recovery sits
// outer, so this one catches first:
//
//	cfg := appkit.DefaultServiceConfig()
//	cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
//	cfg.ExtraMiddlewares = []httputil.Middleware{appkitotel.Recovery(logger)}
//
// A nil logger falls back to slog.Default(). Panics carrying the net/http
// sentinel [http.ErrAbortHandler] are re-panicked unchanged, matching
// httputil.Recovery's contract so the server's silent connection-abort
// handling applies. Without an active recording span the middleware
// degrades gracefully to plain httputil.Recovery behavior.
func Recovery(logger *slog.Logger) httputil.Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				if isAbortPanic(rec) {
					panic(rec)
				}

				recordPanic(req.Context(), rec)

				logger.Error(
					"panic recovered",
					slog.Any("error", rec),
					slog.String("method", req.Method),
					slog.String("path", req.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)

				resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
				resp.WriteHeader(http.StatusInternalServerError)

				_, _ = resp.Write([]byte("Internal Server Error"))
			}()

			next.ServeHTTP(resp, req)
		})
	}
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

// isAbortPanic reports whether v is the net/http ErrAbortHandler sentinel.
// The type assertion is guarded so a panic value of non-comparable type
// cannot crash the comparison itself (mirrors httputil.Recovery).
func isAbortPanic(v any) bool {
	recErr, ok := v.(error)

	return ok && errors.Is(recErr, http.ErrAbortHandler)
}
