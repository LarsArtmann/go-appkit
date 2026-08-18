package otel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// collect drains the manual reader into a flat metric list.
func collect(t *testing.T, reader *sdkmetric.ManualReader) []metricdata.Metrics {
	t.Helper()

	var data metricdata.ResourceMetrics

	err := reader.Collect(t.Context(), &data)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	metrics := []metricdata.Metrics{}
	for _, scope := range data.ScopeMetrics {
		metrics = append(metrics, scope.Metrics...)
	}

	return metrics
}

func findMetric(t *testing.T, metrics []metricdata.Metrics, name string) metricdata.Metrics {
	t.Helper()

	for _, m := range metrics {
		if m.Name == name {
			return m
		}
	}

	t.Fatalf("metric %q not found among %v", name, metricNames(metrics))

	return metricdata.Metrics{}
}

func metricNames(metrics []metricdata.Metrics) []string {
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.Name)
	}

	return names
}

// attrsOf flattens a histogram's attribute sets for assertions.
func histogramDataPoints(
	t *testing.T,
	m metricdata.Metrics,
) []metricdata.HistogramDataPoint[float64] {
	t.Helper()

	hist, ok := m.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("metric %s is %T, want a histogram", m.Name, m.Data)
	}

	return hist.DataPoints
}

func TestMiddleware_EmitsRouteAttributedDurationMetric(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(t.Context()) })

	tp, _ := newRecordingProvider(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(
		Middleware(WithTracerProvider(tp), WithMeterProvider(mp))(mux),
	)
	t.Cleanup(server.Close)

	for _, userID := range []string{"1", "2", "3"} {
		resp, err := http.Get(server.URL + "/users/" + userID) //nolint:noctx // single-shot test request
		if err != nil {
			t.Fatalf("GET /users/%s: %v", userID, err)
		}

		_ = resp.Body.Close()
	}

	metrics := collect(t, reader)
	duration := findMetric(t, metrics, InstrumentHTTPRequestDuration)

	points := histogramDataPoints(t, duration)
	if len(points) == 0 {
		t.Fatal("duration histogram has no data points")
	}

	var totalCount uint64

	routeAttrFound := false

	for _, point := range points {
		totalCount += point.Count

		// The metric's route attribute carries the mux pattern without the
		// method prefix — three distinct user IDs merged into one series,
		// which is the cardinality guarantee that matters.
		if route, ok := point.Attributes.Value(attribute.Key("http.route")); ok && route.AsString() == "/users/{id}" {
			routeAttrFound = true
		}
	}

	if totalCount != 3 {
		t.Errorf("recorded %d observations across %d points, want 3", totalCount, len(points))
	}

	if !routeAttrFound {
		t.Errorf(
			"no data point attributed with http.route=GET /users/{id}; "+
				"cardinality would explode on parametrized routes. Points: %+v",
			points,
		)
	}
}

func TestViews_PinDurationHistogramBoundaries(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(NewHTTPViews()...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(t.Context()) })

	histogram, err := mp.Meter("test").Float64Histogram(InstrumentHTTPRequestDuration)
	if err != nil {
		t.Fatalf("histogram: %v", err)
	}

	histogram.Record(t.Context(), 0.4)

	metrics := collect(t, reader)
	points := histogramDataPoints(t, findMetric(t, metrics, InstrumentHTTPRequestDuration))
	if len(points) != 1 {
		t.Fatalf("got %d data points, want 1", len(points))
	}

	gotBounds := points[0].Bounds
	if len(gotBounds) != len(HTTPDurationBoundaries) {
		t.Fatalf("bounds = %v, want %v", gotBounds, HTTPDurationBoundaries)
	}

	for i, bound := range gotBounds {
		if bound != HTTPDurationBoundaries[i] {
			t.Fatalf("bounds = %v, want %v", gotBounds, HTTPDurationBoundaries)
		}
	}
}

func TestViews_DoNotTouchSizeHistograms(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(NewHTTPViews()...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(t.Context()) })

	counter, err := mp.Meter("test").Int64Counter("some.other.instrument")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}

	counter.Add(t.Context(), 1)

	metrics := collect(t, reader)
	if len(metrics) != 1 || metrics[0].Name != "some.other.instrument" {
		t.Fatalf("unrelated instruments must pass views untouched, got %v", metricNames(metrics))
	}
}
