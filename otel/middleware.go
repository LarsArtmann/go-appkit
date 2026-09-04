package otel

import (
	"net/http"
	"strings"

	"github.com/larsartmann/httputil"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// ScopeName is the instrumentation scope under which Middleware reports.
const ScopeName = "github.com/larsartmann/go-appkit/otel"

// MiddlewareOption configures [Middleware].
type MiddlewareOption func(*middlewareConfig)

type middlewareConfig struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	serverName     string
	publicEndpoint bool
	filters        []func(*http.Request) bool
}

// WithTracerProvider overrides the tracer provider used by the middleware.
// Default: the process-global provider registered by [Setup] (no-op when
// none is registered). Useful for per-handler isolation in tests.
func WithTracerProvider(p trace.TracerProvider) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.tracerProvider = p
	}
}

// WithMeterProvider overrides the meter provider used for the
// semantic-convention metrics. Default: the process-global provider.
func WithMeterProvider(p metric.MeterProvider) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.meterProvider = p
	}
}

// WithServerName records a server.name attribute on every span and metric —
// meaningful when several virtual servers share one process.
func WithServerName(name string) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.serverName = name
	}
}

// WithPublicEndpoint marks the instrumented endpoints as internet-facing:
// incoming traceparent headers are no longer trusted as parents but recorded
// as links, so untrusted callers cannot conjure trace continuity or skew
// sampling. Use for APIs exposed beyond your own infrastructure.
func WithPublicEndpoint() MiddlewareOption {
	return func(c *middlewareConfig) {
		c.publicEndpoint = true
	}
}

// WithFilter skips instrumentation for requests the function rejects
// (returns false). Health endpoints are always filtered; use this to
// additionally exclude e.g. /metrics, static assets, or readiness probes
// of other layers.
func WithFilter(f func(*http.Request) bool) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.filters = append(c.filters, f)
	}
}

// WithFilteredPaths is [WithFilter] for exact paths and /-prefixes:
//
//	WithFilteredPaths("/metrics", "/static/")
func WithFilteredPaths(paths ...string) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.filters = append(c.filters, func(r *http.Request) bool {
			for _, path := range paths {
				if r.URL.Path == path || strings.HasPrefix(r.URL.Path, path+"/") {
					return false
				}
			}

			return true
		})
	}
}

// Middleware returns an [httputil.Middleware] that instruments every
// request with OpenTelemetry: a SERVER span per request, W3C trace-context
// and baggage propagation, and (when a meter provider is registered)
// semantic-convention HTTP metrics — http.server.request.duration with
// method, route, and status attributes.
//
// Span names follow the matched ServeMux pattern ("GET /users/{id}"),
// falling back to the HTTP method when no pattern matched — semantic
// conventions without route-cardinality blowups. Health endpoints
// (/health, /health/live, /health/ready) are never instrumented.
//
// In an appkit service, assign it to ServiceConfig.OuterMiddlewares so the
// span covers the full request lifetime, including the default middleware
// stack:
//
//	cfg := appkit.DefaultServiceConfig()
//	cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
//
// Without [Setup] (or an explicit provider option) the middleware is a
// near-zero-overhead no-op: it propagates nothing and records nothing,
// which keeps OTel strictly opt-in.
func Middleware(opts ...MiddlewareOption) httputil.Middleware {
	cfg := &middlewareConfig{ //nolint:exhaustruct_v5 // options applied below
		filters: nil,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	otelOpts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(spanName),
		otelhttp.WithPropagators(propagator()),
	}

	if cfg.tracerProvider != nil {
		otelOpts = append(otelOpts, otelhttp.WithTracerProvider(cfg.tracerProvider))
	}

	if cfg.meterProvider != nil {
		otelOpts = append(otelOpts, otelhttp.WithMeterProvider(cfg.meterProvider))
	}

	if cfg.serverName != "" {
		otelOpts = append(otelOpts, otelhttp.WithServerName(cfg.serverName))
	}

	if cfg.publicEndpoint {
		otelOpts = append(otelOpts, otelhttp.WithPublicEndpointFn(func(*http.Request) bool {
			return true
		}))
	}

	// The health filter is unconditional; user filters layer on top. A
	// fresh slice keeps the config's backing array untouched.
	filters := make([]func(*http.Request) bool, 0, len(cfg.filters)+1)
	filters = append(filters, isHealthPass)
	filters = append(filters, cfg.filters...)
	otelOpts = append(otelOpts, otelhttp.WithFilter(allPass(filters)))

	return otelhttp.NewMiddleware(ScopeName, otelOpts...)
}

// spanName names spans per HTTP semantic conventions: the matched routing
// pattern when Go 1.22+ ServeMux routing set it (patterns include the
// method, e.g. "GET /users/{id}"), else the bare method. otelhttp calls
// the formatter twice — once at span start (before routing) and once after
// the inner handler ran — so pattern names apply retroactively.
func spanName(_ string, r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}

	return r.Method
}

// isHealthPass reports whether a request should be instrumented. Health
// endpoints are excluded: load-balancer and kubelet probes fire constantly,
// carry no distributed context, and would dominate span volume.
func isHealthPass(r *http.Request) bool {
	path := r.URL.Path

	return path != "/health" && !strings.HasPrefix(path, "/health/")
}

// allPass combines filters with AND over allow-semantics: a request is
// traced only when every filter allows it.
func allPass(filters []func(*http.Request) bool) func(*http.Request) bool {
	return func(r *http.Request) bool {
		for _, f := range filters {
			if !f(r) {
				return false
			}
		}

		return true
	}
}

// propagator resolves the propagation set: the process-global propagator
// (set by Setup) when one is registered, else the W3C default — so the
// middleware keeps correct propagation even in tests that never touched
// the globals.
func propagator() propagation.TextMapPropagator {
	global := otel.GetTextMapPropagator()
	if global != nil && global.Fields() != nil && len(global.Fields()) > 0 {
		return global
	}

	return NewTextMapPropagator()
}
