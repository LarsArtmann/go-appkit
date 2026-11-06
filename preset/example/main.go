// Command preset-example boots a complete service with preset.New and one
// application route — the entire SUPERB v12 composition story in one main.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/preset"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	svc, err := preset.New(preset.Config{ //nolint:exhaustruct_v5 // the minimal demo config — zero values are the point
		Addr:    addr,
		Version: "example-0.1.0",
		HealthChecks: map[string]health.CheckFunc{
			"always": func(context.Context) error { return nil },
		},
	})
	if err != nil {
		slog.Error("boot failed", "error", err)
		os.Exit(1)
	}

	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello from a preset app"))
	})

	if err := svc.Run(context.Background()); err != nil {
		slog.Error("service ended", "error", err)
		os.Exit(1)
	}
}
