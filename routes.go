package appkit

import (
	"net/http"

	errorfamily "github.com/larsartmann/go-error-family"
	httputil "github.com/larsartmann/httputil"
)

// Handle registers handler on the service mux under pattern and records the
// route for the endpoint listing, in one declaration: a route registered
// here is served AND listed, so the listing can never omit it. Routes
// registered directly on [Service.Mux] bypass the listing.
//
// pattern is a ServeMux pattern ("GET /users/{id}", "/health", ...) and
// summary is the one-line description the listing shows next to it. A
// pattern the mux rejects (malformed, or already registered) comes back as a
// Rejection instead of the panic http.ServeMux raises.
func (s *Service) Handle(pattern string, handler http.Handler, summary string) error {
	err := s.Routes.Handle(s.Mux, pattern, handler, summary)
	if err != nil {
		return errorfamily.WrapRejectionf(err, "appkit.route_rejected", "register route %q", pattern)
	}

	return nil
}

// HandleFunc is [Service.Handle] for a bare handler function.
func (s *Service) HandleFunc(
	pattern string,
	handler func(http.ResponseWriter, *http.Request),
	summary string,
) error {
	err := s.Routes.HandleFunc(s.Mux, pattern, handler, summary)
	if err != nil {
		return errorfamily.WrapRejectionf(err, "appkit.route_rejected", "register route %q", pattern)
	}

	return nil
}

// registerCoreRoutes mounts the framework's own routes through the route
// registry, so the endpoint listing covers them exactly like consumer
// routes. Every framework pattern is a literal; an error here means a
// consumer-supplied pattern (Metrics.Path, Endpoints.Path) is not a valid
// ServeMux pattern and surfaces as a Rejection before any listener binds.
// The returned collector is nil when metrics are off.
func (s *Service) registerCoreRoutes() (*metricsCollector, error) {
	err := s.registerHealthRoutes()
	if err != nil {
		return nil, err
	}

	collector, err := s.registerMetricsRoute()
	if err != nil {
		return nil, err
	}

	err = s.registerVersionRoute()
	if err != nil {
		return nil, err
	}

	err = s.registerEndpointListing()

	return collector, err
}

// registerHealthRoutes mounts the default health family.
func (s *Service) registerHealthRoutes() error {
	if s.cfg.RegisterHealth != nil && !*s.cfg.RegisterHealth {
		return nil
	}

	err := s.HandleFunc("GET /health", httputil.HealthHandler(), `Liveness probe: 200 {"status":"up"}`)
	if err != nil {
		return err
	}

	err = s.HandleFunc("GET /health/live", httputil.LiveHandler(), "Kubernetes liveness probe")
	if err != nil {
		return err
	}

	return s.HandleFunc(
		"GET /health/ready",
		httputil.ReadyHandlerWithProbe(s.ready),
		"Readiness probe: 200 while ready, 503 while draining or when ReadyCheck fails",
	)
}

// registerMetricsRoute validates the metrics configuration and mounts the
// exposition behind its Basic Auth. It returns the collector the default
// stack needs, or nil when metrics are off.
func (s *Service) registerMetricsRoute() (*metricsCollector, error) {
	if s.cfg.Metrics == nil {
		return nil, nil //nolint:nilnil // a nil collector IS the metrics-off signal
	}

	s.cfg.Metrics.applyMetricsDefaults()

	err := s.cfg.Metrics.validate()
	if err != nil {
		return nil, err
	}

	collector := newMetricsCollector(s.cfg.Version)

	err = s.Handle(
		"GET "+s.cfg.Metrics.Path,
		metricsAuth(s.cfg.Metrics, collector.handler()),
		"Prometheus metrics exposition (request durations, response totals, build info)",
	)
	if err != nil {
		return nil, err
	}

	return collector, nil
}

// registerVersionRoute mounts the version stamp when one is configured.
func (s *Service) registerVersionRoute() error {
	if s.cfg.Version == "" {
		return nil
	}

	return s.HandleFunc("GET /version", versionHandler(s.cfg.Version), "Build version as JSON")
}

// registerEndpointListing mounts the listing behind its configured guard. A
// route added after this call still appears: every request to the listing
// reads a fresh snapshot of the registry.
func (s *Service) registerEndpointListing() error {
	if s.cfg.Endpoints == nil {
		return nil
	}

	s.cfg.Endpoints.applyEndpointsDefaults()

	err := s.cfg.Endpoints.validate()
	if err != nil {
		return err
	}

	return s.Handle(s.cfg.Endpoints.Path, s.cfg.Endpoints.handler(s.Routes), endpointListingSummary)
}
