package flightrecorder

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/larsartmann/httputil"
)

// defaultErrorThreshold is the HTTP status code at and above which a response
// is considered an error for trigger evaluation.
const defaultErrorThreshold = 500

// MiddlewareOption configures the flight recorder middleware.
type MiddlewareOption func(*middlewareConfig)

type middlewareConfig struct {
	errorThreshold int
	logger         *slog.Logger
	autoReset      bool
}

// WithErrorThreshold sets the HTTP status code at and above which a response
// is considered an error for trigger evaluation. Responses with status >=
// threshold populate [fr.TriggerContext].Err. Default: 500.
func WithErrorThreshold(code int) MiddlewareOption {
	return func(c *middlewareConfig) { c.errorThreshold = code }
}

// WithLogger sets the slog logger for snapshot capture events. When set,
// the middleware logs each capture with the request method, path, duration,
// and status code. Default: no logging.
func WithLogger(logger *slog.Logger) MiddlewareOption {
	return func(c *middlewareConfig) { c.logger = logger }
}

// WithAutoReset controls whether the middleware re-arms the recorder's
// once-latch after each successful capture, allowing multiple snapshots over
// the recorder's lifetime. Default: true.
//
// Disable for strict once-semantics (only the first matching request captures
// a trace, matching go-flightrecorder's default behavior without Reset).
func WithAutoReset(enabled bool) MiddlewareOption {
	return func(c *middlewareConfig) { c.autoReset = enabled }
}

// Middleware returns an [httputil.Middleware] that triggers flight recorder
// snapshots when HTTP requests match the given trigger conditions.
//
// The middleware measures request duration and captures the HTTP status code
// via [httputil.ResponseRecorder]. After the handler completes, it constructs
// a [fr.TriggerContext] and delegates the capture decision to the trigger
// function. If the trigger fires, a snapshot is written to the recorder's
// configured destination (set via [fr.WithFile] or [fr.WithWriter]).
//
// Example — capture on errors or requests slower than 100ms:
//
//	mw := flightrecorder.Middleware(rec, fr.OnErrorOrLatency(100*time.Millisecond))
//
// Example — capture only on 5xx errors with logging:
//
//	mw := flightrecorder.Middleware(rec, fr.OnError(),
//	    flightrecorder.WithLogger(svc.Logger),
//	)
func Middleware(rec *fr.Recorder, trigger fr.TriggerFunc, opts ...MiddlewareOption) httputil.Middleware {
	cfg := middlewareConfig{
		errorThreshold: defaultErrorThreshold,
		logger:         nil,
		autoReset:      true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			responseRecorder := httputil.NewResponseRecorder(w)

			next.ServeHTTP(responseRecorder, r)

			duration := time.Since(start)

			tc := fr.TriggerContext{
				Kind:     "http",
				Type:     r.Method + " " + r.URL.Path,
				Duration: duration,
				Err:      statusError(responseRecorder.Status(), cfg.errorThreshold),
			}

			captured := rec.SnapshotIf(r.Context(), tc, trigger)
			if !captured {
				return
			}

			if cfg.logger != nil {
				cfg.logger.InfoContext(r.Context(),
					"flightrecorder: trace snapshot captured",
					"method", r.Method,
					"path", r.URL.Path,
					"duration", duration,
					"status", responseRecorder.Status(),
				)
			}

			if cfg.autoReset {
				rec.Reset()
			}
		})
	}
}

// errHTTPStatus is the sentinel underlying status-class errors, so callers
// can match any threshold-exceeding status with errors.Is.
var errHTTPStatus = errors.New("http status")

// statusError returns an error if the status code indicates a server error,
// nil otherwise. This populates [fr.TriggerContext].Err so that [fr.OnError]
// and [fr.OnErrorOrLatency] triggers fire on error responses.
func statusError(status, threshold int) error {
	if status >= threshold {
		return fmt.Errorf("%w %d", errHTTPStatus, status)
	}

	return nil
}
