// Command otel-demo shows the appkit otel module in action: a service with
// a span on every request, semantic-convention metrics, W3C trace-context
// propagation, trace-correlated handler logs, exceptions in both flavors
// (panics and handled errors), flight-recorder captures reported as OTel
// metrics, and a telemetry flush wired into graceful shutdown.
//
// Run and try:
//
//	go run ./example
//	curl -i http://localhost:8080/users/alice   # span + correlated log
//	curl -i http://localhost:8080/boom          # handled error -> exception event
//	curl -i http://localhost:8080/panic         # panic -> exception event + stack trace
//	curl -i http://localhost:8080/slow          # 150ms route -> fr capture -> appkit_flightrecorder_snapshots_total{type="GET /slow"}
//	curl -i http://localhost:8080/health        # no span — health is filtered
//	ctrl-C                                      # graceful drain, then telemetry flushes
//
// Telemetry follows the environment, zero code: set
// OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 (a local SigNoz) and the
// spans and metrics stream straight into SigNoz — traces, metrics, and the
// exception events behind its Exceptions view. Without the variable, spans
// pretty-print to stdout for local inspection. Import
// dashboards/appkit-http-dashboard.json for a ready-made SigNoz dashboard.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/larsartmann/go-appkit"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	errorfamily "github.com/larsartmann/go-error-family"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/larsartmann/httputil"
)

// Demo tuning knobs: named constants keep the demo honest about what the
// retention caps and the slow-route threshold are.
const (
	demoMaxSnapshots   = 5
	demoMaxBytes       = 64 << 20
	demoSlowRouteDelay = 150 * time.Millisecond
	demoCaptureLatency = 100 * time.Millisecond
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

// Demo failures: static sentinels keep the errors classifiable and grep-able.
var (
	errPaymentGatewayTimeout = errors.New("payment gateway timed out")
	errCacheStampede         = errors.New("cache stampede")
)

func run(cfg appkit.ServiceConfig) error {
	// Handler-level logs carry trace_id/span_id; create the logger before
	// the Recovery middleware needs it.
	logger := slog.New(appkitotel.TraceHandler(slog.NewJSONHandler(os.Stdout, nil)))

	telemetry := []appkitotel.SetupOption{
		appkitotel.WithService("otel-demo", "1.0.0", "local"),
		appkitotel.WithEnvironment("development"),
	}

	// Zero-code OTLP: with OTEL_EXPORTER_OTLP_ENDPOINT set (e.g. a local
	// SigNoz), Setup builds the OTLP exporters itself. Without it, the demo
	// falls back to pretty-printing spans on stdout.
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		telemetry = append(telemetry, appkitotel.WithStdoutExporter(os.Stdout))
	}

	provider, err := appkitotel.Setup(telemetry...)
	if err != nil {
		return fmt.Errorf("otel setup: %w", err)
	}

	// Flight recorder: ONE per process (Go runtime constraint). Captures on
	// error-status or >100ms requests land as timestamped files, and every
	// capture completion flows into OTel metrics via the bridge — in SigNoz,
	// chart appkit_flightrecorder_snapshots_total by type to see which
	// operations produced traces.
	traceDir, err := os.MkdirTemp("", "otel-demo-traces")
	if err != nil {
		return fmt.Errorf("create trace dir: %w", err)
	}

	rec, err := fr.New(
		fr.WithSnapshotDir(traceDir),
		fr.WithSnapshotPrefix("trace"),
		fr.WithMaxSnapshots(demoMaxSnapshots),
		fr.WithMaxBytes(demoMaxBytes),
		fr.WithMetrics(appkitotel.NewFlightRecorderMetricsHook(provider.AsMeterProvider().Meter("otel-demo"))),
	)
	if err != nil {
		return fmt.Errorf("create flight recorder: %w", err)
	}

	err = rec.Start()
	if err != nil {
		return fmt.Errorf("start flight recorder: %w", err)
	}

	// Tracing wraps the whole request (including the default middleware
	// stack); Recovery turns panics into exception events inside those
	// spans; the recorder closes during graceful shutdown (Close drains
	// in-flight async captures); the provider flushes after the server
	// released its connections.
	cfg.ShutdownHooks = []func(context.Context) error{
		func(context.Context) error { return rec.Close() }, // Close drains in-flight async captures
		provider.Shutdown,
	}

	cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
	cfg.ExtraMiddlewares = []httputil.Middleware{appkitotel.Recovery(logger)}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	defer func() { _ = svc.Close() }()

	svc.Mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "user fetched", "user_id", r.PathValue("id"))

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("ok"))
	})

	svc.Mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		appkitotel.RecordError(r.Context(), errPaymentGatewayTimeout) // SigNoz Exceptions view entry

		logger.ErrorContext(r.Context(), "checkout failed", "error", errPaymentGatewayTimeout.Error())

		http.Error(w, "checkout failed", http.StatusInternalServerError)
	})

	svc.Mux.HandleFunc("GET /panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic(errCacheStampede) // Recovery records the exception, answers 500
	})

	svc.Mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		time.Sleep(demoSlowRouteDelay)

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("finally done"))

		rec.SnapshotIf(r.Context(), fr.TriggerContext{ //nolint:exhaustruct_v5 // success route: no Err
			Kind:     "http",
			Type:     "GET /slow",
			Duration: time.Since(start),
		}, fr.OnLatency(demoCaptureLatency))
	})

	return svc.Run(context.Background()) //nolint:wrapcheck // top-level main returns the error as-is
}
