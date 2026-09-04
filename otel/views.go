package otel

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Instrument names emitted by otelhttp under stable HTTP semantic
// conventions. Exposed for dashboard, alert, and test authoring.
const (
	// InstrumentHTTPRequestDuration is the per-request server latency
	// histogram, attributed with http.request.method, http.route (when a
	// ServeMux pattern matched), and http.response.status_code.
	InstrumentHTTPRequestDuration = "http.server.request.duration"
)

// HTTPDurationBoundaries pins the semantic-convention-recommended explicit
// histogram boundaries for HTTP request duration (seconds): 0, 5ms, 10ms,
// 25ms, 50ms, 75ms, 100ms, 250ms, 500ms, 750ms, 1s, 2.5s, 5s, 7.5s, 10s.
var HTTPDurationBoundaries = []float64{ //nolint:gochecknoglobals // a constant set expressed as a slice
	0, 0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
}

// NewHTTPViews returns SDK metric views pinning the HTTP duration histogram
// boundaries. Setup applies them automatically; use them directly when
// building your own MeterProvider.
//
// Only InstrumentHTTPRequestDuration is matched: request/response size
// histograms keep the SDK default boundaries, which are correct for bytes.
func NewHTTPViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{ //nolint:exhaustruct_v5 // only Name is a filter criterion
				Name: InstrumentHTTPRequestDuration,
			},
			sdkmetric.Stream{ //nolint:exhaustruct_v5 // only Aggregation is configured
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{ //nolint:exhaustruct_v5 // NoMinMax stays false
					Boundaries: HTTPDurationBoundaries,
				},
			},
		),
	}
}
