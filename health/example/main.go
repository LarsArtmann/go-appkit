// Command example demonstrates the health module end to end: a go-health
// probe with a critical and a non-critical check, the real-time dashboard,
// and appkit lifecycle wiring (DrainHooks / ShutdownHooks).
//
// Run from the health module directory:
//
//	GOWORK=off GOEXPERIMENT=jsonv2 go run ./example
//
// Then open http://localhost:8081/health (PORT overrides). The "cache"
// check fails for 3 seconds of every 15, degrading the dashboard to warn
// without touching readiness — add "cache" to WithCriticalServices to see
// fail instead. SIGTERM/SIGINT flips /readyz and appkit's /health/ready to
// 503 in lockstep (drain window), then stops the dashboard pusher.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/larsartmann/go-appkit"
	appkithealth "github.com/larsartmann/go-appkit/health"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
)

const (
	defaultPort         = "8081"
	flapWindowSeconds   = 3
	flapPeriodSeconds   = 15
	dashboardTrendCount = 300
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	healthDisabled := false

	checks := map[string]appkithealth.CheckFunc{
		"database": func(context.Context) error { return nil },
		"cache": func(context.Context) error {
			if time.Now().Second()%flapPeriodSeconds < flapWindowSeconds {
				return errorfamily.NewInfrastructure("demo.cache_evicting", "cache evicting")
			}

			return nil
		},
	}

	probe := appkithealth.NewProbe(checks, health.WithCriticalServices("database"))

	mounted, err := appkithealth.New(probe, appkithealth.WithDashboard(
		dashboard.WithTrend(dashboardTrendCount),
		dashboard.WithMetrics(true),
	))
	if err != nil {
		log.Fatalf("health surface: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:" + port
	cfg.RegisterHealth = &healthDisabled
	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
		mounted.Drain()

		return nil
	})
	cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)

	svc, err := appkit.NewService(cfg)
	if err != nil {
		log.Fatalf("service: %v", err)
	}

	// Method-less "/hello" on purpose: the dashboard registers
	// method-agnostic routes (/health), and a "GET /" catch-all trips Go's
	// ServeMux precedence rules ("matches more methods than GET /").
	svc.Mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "hello from appkit") //nolint:errcheck // demo handler
	})

	// Keep a load-balancer-facing readiness endpoint alongside the probe's
	// /readyz — both report 503 during the drain window.
	svc.Mux.HandleFunc("GET /health/ready", appkit.ReadyHandlerWithProbe(mounted.Ready))

	mounted.RegisterRoutes(svc.Mux)

	ctx := context.Background()

	if err := mounted.Start(ctx); err != nil { //nolint:noinlineerr // example brevity
		log.Fatalf("start health surface: %v", err)
	}

	log.Printf("dashboard at http://localhost:%s/health", port) //nolint:gosec // demo: port comes from the operator's own env

	err = svc.Run(ctx)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}
