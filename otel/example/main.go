// Command otel-demo shows the appkit otel module in action: a service with
// a span on every request, semantic-convention metrics, W3C trace-context
// propagation, trace-correlated handler logs, and a telemetry flush wired
// into graceful shutdown.
//
// Run and try:
//
//	go run ./example
//	curl -i http://localhost:8080/users/alice   # span + correlated log on stdout
//	curl -i http://localhost:8080/health        # no span — health is filtered
//	curl -i http://localhost:8080/boom          # span with error status
//	ctrl-C                                      # graceful drain, then spans flush
//
// Spans are pretty-printed to stdout for the demo; in production, construct
// an OTLP exporter and pass it via appkitotel.WithSpanExporter.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/larsartmann/go-appkit"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"
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
	provider, err := appkitotel.Setup(
		appkitotel.WithService("otel-demo", "1.0.0", "local"),
		appkitotel.WithStdoutExporter(os.Stdout), // development; OTLP in production
	)
	if err != nil {
		return fmt.Errorf("otel setup: %w", err)
	}

	// Tracing wraps the whole request (including the default middleware
	// stack); the provider flushes after the server released its
	// connections during graceful shutdown.
	cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
	cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	defer func() { _ = svc.Close() }()

	// Handler-level logs carry trace_id/span_id; pass the request context.
	logger := slog.New(appkitotel.TraceHandler(slog.NewJSONHandler(os.Stdout, nil)))

	svc.Mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "user fetched", "user_id", r.PathValue("id"))

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("ok"))
	})

	svc.Mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		logger.WarnContext(r.Context(), "handler failing")

		w.WriteHeader(http.StatusInternalServerError)
	})

	return svc.Run(context.Background()) //nolint:wrapcheck // top-level main returns the error as-is
}
