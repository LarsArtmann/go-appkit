// Command example is the short tour of go-appkit's core: the default
// middleware stack, the v0.4.0 lifecycle hooks (OuterMiddlewares, DrainHooks,
// ShutdownHooks), and graceful drain — all in one runnable file.
//
// Run it, then:
//
//	curl localhost:8080/
//	curl -i localhost:8080/health/ready   # during shutdown: 503 while draining
//	Ctrl+C                                 # watch the shutdown phase logs
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = ":8080"
	cfg.DrainDelay = 3 * time.Second // shorten the default 5s drain for the demo
	cfg = withLifecycle(cfg)

	err := run(cfg)
	if err != nil {
		os.Exit(errorfamily.HandleError(err))
	}
}

// withLifecycle demonstrates the v0.4.0 lifecycle surface. All three hooks
// are optional; production services typically wire OuterMiddlewares for
// tracing (see the otel module) and ShutdownHooks for telemetry flushes.
func withLifecycle(cfg appkit.ServiceConfig) appkit.ServiceConfig {
	// OuterMiddlewares wrap the ENTIRE chain (default stack included) and run
	// outermost — where tracing belongs. A request-timing logger is the
	// dependency-free stand-in here; production wires the otel module's
	// middleware instead.
	cfg.OuterMiddlewares = append(cfg.OuterMiddlewares, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			fmt.Printf("[outer] %s %s took %s\n", r.Method, r.URL.Path, time.Since(start).Round(time.Microsecond))
		})
	})

	// DrainHooks run at drain start: the ready probe is already down but the
	// server still serves, so external readiness signals (health probes, LB
	// marks) flip in lockstep for the whole drain window.
	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
		fmt.Println("[drain] telling the load balancer to stop routing")

		return nil
	})

	// ShutdownHooks run AFTER the server released its connections — the right
	// moment to flush telemetry covering the final in-flight requests.
	cfg.ShutdownHooks = append(cfg.ShutdownHooks, func(context.Context) error {
		fmt.Println("[shutdown] flushing telemetry provider")

		return nil
	})

	return cfg
}

func run(cfg appkit.ServiceConfig) error {
	svc, err := appkit.NewService(cfg)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	defer func() { _ = svc.Close() }()

	svc.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("hello"))
	})

	return svc.Run(context.Background()) //nolint:wrapcheck // top-level main returns the error as-is
}
