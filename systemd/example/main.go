// Command example is an appkit service wired for systemd Type=notify:
// sd_notify READY=1 once the listener binds, STOPPING=1 at the start of
// the graceful drain, and the WATCHDOG=1 keepalive in between. Outside
// systemd every hook is a no-op, so the same binary runs in development.
//
// Try it under systemd with a unit like:
//
//	[Service]
//	Type=notify
//	WatchdogSec=30
//	ExecStart=/usr/local/bin/example
//
//	PORT overrides the listen port (dev machines often have 8080 occupied —
// same convention as the core and otel examples).
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	appkit "github.com/larsartmann/go-appkit"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-appkit/systemd"
)

func main() {
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = ":8080"

	if port := os.Getenv("PORT"); port != "" {
		cfg.Addr = ":" + port
	}

	// The one-line systemd wiring: appends READY=1 (post-listen),
	// STOPPING=1 (drain start), and the watchdog keepalive + its stop to
	// the config's hook slices.
	systemd.Install(&cfg)

	svc, err := appkit.NewService(cfg)
	if err != nil {
		panic(err)
	}

	defer func() { _ = svc.Close() }()

	svc.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("hello from a Type=notify service"))
	})

	if err := svc.Run(context.Background()); err != nil { //nolint:wrapcheck // top-level main returns the error as-is
		os.Exit(errorfamily.HandleError(err))
	}

	fmt.Fprintln(os.Stdout, "stopped")
}
