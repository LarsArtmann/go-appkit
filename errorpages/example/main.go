// Command errorpages-demo shows the appkit errorpages module in action:
// a service with one real route, a pretty 404 catch-all, and classified
// error rendering with content negotiation.
//
// Run and try:
//
//	curl -i http://localhost:8080/health
//	curl -i http://localhost:8080/no/such/page            # pretty HTML 404
//	curl -i -H 'Accept: application/json' http://localhost:8080/nope
//	curl -i http://localhost:8080/teapot                  # Rejection -> 400
//	curl -i http://localhost:8080/explode                 # Infrastructure -> 503
//	curl -i -X POST http://localhost:8080/health          # plain mux 405
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = ":8080"

	svc, err := appkit.NewService(cfg)
	if err != nil {
		slog.Error("create service", "error", err)

		return
	}

	errorpages.Mount(svc.Mux, errorpages.Config{}) //nolint:exhaustruct_v5 // zero Config

	svc.Mux.HandleFunc("GET /teapot", func(w http.ResponseWriter, r *http.Request) {
		err := errorfamily.NewRejection("demo.teapot", "short and stout")
		errorpages.Write(w, r, err, errorpages.Config{}) //nolint:exhaustruct_v5 // zero Config
	})

	svc.Mux.HandleFunc("GET /explode", func(w http.ResponseWriter, r *http.Request) {
		err := errorfamily.NewInfrastructure("demo.explode", "the database left the building")
		errorpages.Write(w, r, err, errorpages.Config{}) //nolint:exhaustruct_v5 // zero Config
	})

	err = svc.Run(context.Background())
	if err != nil {
		slog.Error("service stopped", "error", err)
	}
}
