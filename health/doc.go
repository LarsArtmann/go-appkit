// Package health bridges go-appkit services with [go-health], the
// three-probe health-check SDK (liveness, readiness, startup with
// critical/non-critical classification, background caching, and shutdown
// awareness), and — optionally — [go-health-dashboard], the real-time HTML
// dashboard with SSE updates, trend history, Prometheus exposition, and
// webhooks.
//
// The module has NO dependency on go-appkit core: Mount works on any
// *http.ServeMux, and the lifecycle handles wire into ServiceConfig.DrainHooks
// and ServiceConfig.ShutdownHooks as plain functions. Consumers alias the
// import (appkithealth is the convention) when they also import the SDK.
//
// Unqualified health.* references in this documentation mean the go-health
// SDK.
//
// # Quick start
//
//	no := false
//
//	checks := map[string]appkithealth.CheckFunc{
//		"database": func(ctx context.Context) error { return db.PingContext(ctx) },
//		"cache":    func(ctx context.Context) error { return cache.PingContext(ctx) },
//	}
//
//	probe := appkithealth.NewProbe(checks, health.WithCriticalServices("database"))
//
//	cfg := appkit.DefaultServiceConfig()
//	cfg.RegisterHealth = &no // this module serves the richer health surface
//	cfg.ReadyCheck = nil     // probe readiness is served by /readyz itself
//
//	svc, err := appkit.NewService(cfg)
//	// ...
//
//	mounted, err := appkithealth.Mount(svc.Mux, probe, appkithealth.WithDashboard(
//		dashboard.WithTrend(300),
//		dashboard.WithMetrics(true),
//	))
//	// ...
//
//	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
//		mounted.Drain()
//		return nil
//	})
//	cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)
//
//	ctx := context.Background()
//	if err := mounted.Start(ctx); err != nil { // probe cache + dashboard pusher
//		// ...
//	}
//
//	_ = svc.Run(ctx) // blocks until SIGINT/SIGTERM
//
// # Routes
//
// Mount registers the Kubernetes probe endpoints (/healthz, /readyz,
// /startupz by default) on the mux. With WithDashboard it additionally
// serves the HTML dashboard at /health (JSON on Accept: application/json),
// the SSE stream at /health/sse, the favicon, and any metrics/trend/export
// endpoints configured through the dashboard options; the probe endpoints
// are then registered by the dashboard itself, so all dashboard routing
// (WithBasePath, WithRoutes) applies uniformly. Set
// ServiceConfig.RegisterHealth to &false so the default httputil endpoints
// do not collide with the dashboard's /health route; the mux panics on the
// duplicate registration if you forget.
//
// # Lifecycle ordering
//
//  1. mounted.Start(ctx) before serving: starts the probe's background
//     evaluation (an initial synchronous batch runs first, so the dashboard's
//     first patch has data) and the dashboard's SSE pusher.
//  2. mounted.Drain() from a DrainHook: flips every readiness surface the
//     probe serves to 503 at the start of appkit's drain window, in lockstep
//     with the framework's own ready probe, while liveness stays 200.
//  3. mounted.Shutdown(ctx) from a ShutdownHook: stops the pusher, drains
//     connected SSE clients, and stops the probe's refresh loop.
//
// Shutdown is safe to call multiple times; Start again after Shutdown
// restarts the surface (useful in tests).
//
// # Classification
//
// Checks are classified through the SDK's WithCriticalServices option: a
// failing critical check makes readiness report fail (503); failing
// non-critical checks degrade the roll-up to warn (readiness stays 200).
// NewProbe checks run concurrently per batch and are panic-isolated per
// check: a panicking check fails as that check's error instead of poisoning
// the batch.
//
// # Build requirement
//
// GOEXPERIMENT=jsonv2 is required (go-health serializes responses with
// encoding/json/v2 and the dashboard depends on go-sse).
//
// [go-health]: https://github.com/larsartmann/go-health
// [go-health-dashboard]: https://github.com/larsartmann/go-health-dashboard
package health
