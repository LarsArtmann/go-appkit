package main

import (
	"context"
	"net/http"
	"os"

	appkit "github.com/larsartmann/go-appkit"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = ":8080"

	err := run(cfg)
	if err != nil {
		os.Exit(errorfamily.HandleError(err))
	}
}

func run(cfg appkit.ServiceConfig) error {
	svc, err := appkit.NewService(cfg)
	if err != nil {
		return err
	}

	defer func() { _ = svc.Close() }()

	svc.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("hello"))
	})

	return svc.Run(context.Background())
}
