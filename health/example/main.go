// Command example demonstrates the health module end to end: a go-health
// probe with a critical and a non-critical check, the real-time dashboard,
// and appkit lifecycle wiring (DrainHooks / ShutdownHooks).
//
// Run from the health module directory:
//
//	GOWORK=off GOTOOLCHAIN=go1.27.1 go run ./example
//
// Then open http://localhost:8081/health (PORT overrides). The "cache"
// check fails for 3 seconds of every 15, degrading the dashboard to warn
// without touching readiness — add "cache" to WithCriticalServices to see
// fail instead. SIGTERM/SIGINT flips /readyz and appkit's /health/ready to
// 503 in lockstep (drain window), then stops the dashboard pusher.
//
// Add -hardened to compose the hardened-posture dashboard instead: the
// DashboardHardenedPreset (base path + per-request nonce extraction) behind
// a strict-CSP middleware built with the security module's BuildCSP — the
// middleware a real operator runs OUTSIDE the health module, in front of
// the mux. Verify the header with:
//
//	curl -si localhost:8081/health | grep -i content-security-policy
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/larsartmann/go-appkit"
	appkithealth "github.com/larsartmann/go-appkit/health"
	appkitsecurity "github.com/larsartmann/go-appkit/security"
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
	hardened := flag.Bool("hardened", false, "serve the dashboard behind DashboardHardenedPreset + a strict BuildCSP middleware")

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

	dashboardOpts := []dashboard.Option{
		dashboard.WithTrend(dashboardTrendCount),
		dashboard.WithMetrics(true),
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:" + port

	if *hardened {
		// Bridge the security module's context-based nonce to the preset's
		// request-based extractor contract (same one-liner as the integration
		// module's hardened-dashboard test).
		nonceFn := func(r *http.Request) string { return appkitsecurity.NonceFromContext(r.Context()) }
		dashboardOpts = appkithealth.DashboardHardenedPreset("/health", nonceFn)
		cfg.OuterMiddlewares = []func(http.Handler) http.Handler{hardenedCSP}
	}

	mounted, err := appkithealth.New(probe, appkithealth.WithDashboard(dashboardOpts...))
	if err != nil {
		log.Fatalf("health surface: %v", err)
	}

	// The dashboard's live view rides /health/sse; the default 30s
	// WriteTimeout would cut the stream every 30s (browser auto-reconnects —
	// degraded, not broken). NoTimeout keeps the stream stable for the demo.
	cfg.WriteTimeout = appkit.NoTimeout
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

	dashURL := fmt.Sprintf("http://localhost:%s/health", port)
	log.Printf("dashboard at %s", dashURL) //nolint:gosec // demo: port comes from the operator's own env

	err = svc.Run(ctx)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}

// hardenedCSP mints a per-request nonce, exposes it through the security
// module's context contract, and sets the matching strict policy header —
// the middleware a real operator runs in front of the mux (never inside the
// health module: it is core-free and security-free by design).
func hardenedCSP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := appkitsecurity.GenerateNonce()
		if err != nil {
			http.Error(w, "nonce generation failed", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Security-Policy", appkitsecurity.BuildCSP(appkitsecurity.CSPConfig{
			Environment: appkitsecurity.Production,
			Nonce:       nonce,
		}))
		next.ServeHTTP(w, r.WithContext(appkitsecurity.WithNonce(r.Context(), nonce)))
	})
}
