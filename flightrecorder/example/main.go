// Command flightrecorder-demo shows the complete ops loop for the
// flightrecorder module: a production-shaped recorder (retained, compressed,
// logged), a middleware that captures on the OnAll(OnError, OnLatency)
// composite, a manual snapshot endpoint with HTTP download, and completion
// telemetry flowing into the service log — the seam the HTTP middleware
// cannot see.
//
// Run and try:
//
//	go run ./example                       # TRACE_DIR overrides the snapshot dir
//	curl -i http://localhost:8080/ok           # fast success — no capture
//	curl -i http://localhost:8080/slow-failure # slow AND 500 — OnAll fires, trace.gz lands
//	curl -i http://localhost:8080/fail         # fast 500 — OnAll narrows it out (no capture)
//	curl -X POST 'http://localhost:8080/debug/flightrecorder/snapshot?download=1' -OJ
//	ctrl-C                                     # graceful shutdown; Close drains async captures
//
// The demo uses published core only: the recorder starts before Run and
// closes in ShutdownHooks. On core v0.8.0+ the Start() call belongs in
// StartHooks (after the listener binds, before serving) — see the systemd
// module for that seam.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/larsartmann/go-appkit"
	appkitfr "github.com/larsartmann/go-appkit/flightrecorder"
	errorfamily "github.com/larsartmann/go-error-family"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/larsartmann/httputil"
)

// Demo tuning knobs: retention caps mirror the documented ops values; the
// capture latency threshold keeps the OnAll narrowing visible at demo speed.
const (
	demoMaxSnapshots   = 5
	demoMaxBytes       = 64 << 20
	demoCaptureLatency = 100 * time.Millisecond
	demoSlowRouteDelay = 150 * time.Millisecond
)

func main() {
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = ":8080"

	if port := os.Getenv("PORT"); port != "" {
		cfg.Addr = ":" + port
	}

	err := run(cfg)
	if err != nil {
		os.Exit(errorfamily.HandleError(err))
	}
}

func run(cfg appkit.ServiceConfig) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	traceDir := os.Getenv("TRACE_DIR")
	if traceDir == "" {
		var err error

		traceDir, err = os.MkdirTemp("", "flightrecorder-demo-traces")
		if err != nil {
			return fmt.Errorf("create trace dir: %w", err)
		}
	}

	// Completion telemetry: the middleware logs capture INITIATION
	// (request-correlated); this hook logs the COMPLETION — bytes written,
	// artifact path, sink errors — which is exactly what the middleware
	// never sees because captures run asynchronously.
	metricsHook := func(event fr.SnapshotEvent, err error) {
		attrs := []any{
			"bytes", event.Bytes,
			"path", event.Path,
			"duration", event.Duration.String(),
			"source", event.Source,
			"kind", event.Kind,
			"type", event.Type,
			"compressed", event.Compressed,
		}

		if err != nil {
			attrs = append(attrs, "error", err.Error())
			logger.Error("flightrecorder: snapshot failed", attrs...)

			return
		}

		logger.Info("flightrecorder: snapshot complete", attrs...)
	}

	opts := append(
		appkitfr.OpsRecorderLoggerPreset(traceDir, demoMaxSnapshots, demoMaxBytes, logger),
		fr.WithMetrics(metricsHook),
	)

	rec, err := fr.New(opts...)
	if err != nil {
		return fmt.Errorf("create flight recorder: %w", err)
	}

	// Fail-closed: a demo about captures without a running recorder is a
	// silent lie. (Start returns *AlreadyEnabledError — classified
	// errors.Is(fr.ErrAlreadyEnabled) — if the process-global trace slot is
	// taken; surface that loudly instead of running recorder-less.)
	err = rec.Start()
	if err != nil {
		return errorfamily.Newf(errorfamily.Infrastructure,
			"flightrecorder.start_failed",
			"start flight recorder: %v", err)
	}

	// OnAll narrowing: captures need BOTH an error response AND 100ms+ —
	// fast failures and slow successes do not burn the retention budget.
	cfg.ExtraMiddlewares = []httputil.Middleware{
		appkitfr.Middleware(rec, fr.OnAll(fr.OnError(), fr.OnLatency(demoCaptureLatency)),
			appkitfr.WithLogger(logger)),
	}

	// Close drains in-flight async captures, stops the runtime recorder,
	// and closes the sink — after the server released its connections.
	cfg.ShutdownHooks = []func(context.Context) error{
		func(context.Context) error { return rec.Close() },
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	defer func() { _ = svc.Close() }()

	// Download mode serves the raw trace (gzip here — the preset compresses;
	// the filename says so, see WithSnapshotFilename's caveat).
	appkitfr.Mount(svc.Mux, "POST /debug/flightrecorder/snapshot", rec,
		appkitfr.WithSnapshotFilename("trace.trace.gz"))

	svc.Mux.HandleFunc("GET /ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("ok"))
	})

	svc.Mux.HandleFunc("GET /slow-failure", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(demoSlowRouteDelay)

		http.Error(w, "dependency timeout", http.StatusInternalServerError)
	})

	svc.Mux.HandleFunc("GET /fail", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "fast failure", http.StatusInternalServerError)
	})

	return svc.Run(context.Background()) //nolint:wrapcheck // top-level main returns the error as-is
}
