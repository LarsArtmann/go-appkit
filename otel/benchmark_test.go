package otel_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// benchmarkHandler is the minimal handler every variant wraps: response is
// fixed, no allocation churn from the handler itself.
func benchmarkHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func runMiddlewareBenchmark(b *testing.B, mw func(http.Handler) http.Handler) {
	b.Helper()

	server := httptest.NewServer(mw(http.HandlerFunc(benchmarkHandler)))
	b.Cleanup(server.Close)

	client := server.Client()

	b.ResetTimer()

	for b.Loop() {
		req, err := http.NewRequestWithContext(b.Context(), http.MethodGet, server.URL, nil)
		if err != nil {
			b.Fatalf("build request: %v", err)
		}

		res, err := client.Do(req)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}

		_ = res.Body.Close()
	}
}

// BenchmarkMiddleware_NoOp measures the middleware with no providers wired —
// the strictly-opt-in baseline. Without a tracer provider the otelhttp
// bridge skips span creation, so this is the "otel module present, telemetry
// absent" cost.
func BenchmarkMiddleware_NoOp(b *testing.B) {
	runMiddlewareBenchmark(b, otel.Middleware())
}

// BenchmarkMiddleware_Traced measures spans only: a tracer provider with a
// batching processor and no exporter (batch never flushes anywhere, so the
// number isolates span construction/enqueue, not export I/O).
func BenchmarkMiddleware_Traced(b *testing.B) {
	tp := sdktrace.NewTracerProvider()

	runMiddlewareBenchmark(b, otel.Middleware(otel.WithTracerProvider(tp)))
}

// BenchmarkMiddleware_TracedAndMetered measures spans + the request-duration
// histogram: the full production instrumentation minus export I/O.
func BenchmarkMiddleware_TracedAndMetered(b *testing.B) {
	tp := sdktrace.NewTracerProvider()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithView(otel.NewHTTPViews()...),
	)

	runMiddlewareBenchmark(b, otel.Middleware(
		otel.WithTracerProvider(tp),
		otel.WithMeterProvider(mp),
	))
}
