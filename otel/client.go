package otel

import (
	"net/http"

	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Transport wraps http.DefaultTransport with otelhttp's client
// instrumentation: one CLIENT span per outbound request (named
// `HTTP <method>` per the client semconv default), W3C traceparent/baggage
// injection so the downstream service continues THIS service's trace, and
// client-side semconv metrics. Providers default to the process globals
// registered by [Setup]; without one, spans and metrics are no-ops — the
// same strictly-opt-in contract as [Middleware].
//
//	client := &http.Client{Transport: appkitotel.Transport()}
//
// Set a base transport for connection tuning (timeouts live on the Client,
// not the RoundTripper):
//
//	base := http.DefaultTransport.(*http.Transport).Clone()
//	client := &http.Client{Transport: appkitotel.Transport(WithBaseTransport(base))}
//
// Options are the [MiddlewareOption]s that make sense for a client:
// [WithTracerProvider] and [WithMeterProvider] override the providers
// (isolation in tests); server-side options — filters, server name,
// public-endpoint mode — do not apply to outbound calls and are ignored.
func Transport(opts ...MiddlewareOption) http.RoundTripper {
	cfg := &middlewareConfig{ //nolint:exhaustruct_v5 // options applied below
		filters: nil,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	otelOpts := []otelhttp.Option{
		otelhttp.WithPropagators(propagator()),
	}

	if cfg.tracerProvider != nil {
		otelOpts = append(otelOpts, otelhttp.WithTracerProvider(cfg.tracerProvider))
	}

	if cfg.meterProvider != nil {
		otelOpts = append(otelOpts, otelhttp.WithMeterProvider(cfg.meterProvider))
	}

	return otelhttp.NewTransport(cfg.baseTransport, otelOpts...)
}
