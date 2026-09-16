package integration_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/httputil"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestSpanNameAndRouteThroughAppkitOuterMiddlewares pins the documented
// wiring: otel middleware as OuterMiddlewares on a real appkit Service
// (default stack: Recovery → RequestID → Logging → Timeout → SecurityHeaders).
// Span name must be the matched ServeMux pattern and the http.route metric
// attribute must survive the whole chain. This is the regression pin for the
// 2026-09-15 pattern-propagation bug that shipped in every tag before
// httputil v1.2.0 / otel v0.1.1 (otelhttp reads r.Pattern on its own request
// fork; the forking middlewares between otel and the mux shadowed it, so
// spans flushed as bare "GET" and metrics lost http.route). The module
// charter pins PUBLISHED tags, so this test fails against any pin older than
// the train tags — by design.
func TestSpanNameAndRouteThroughAppkitOuterMiddlewares(t *testing.T) {
	t.Parallel()

	spanExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spanExporter))
	mr := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(mr))

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay
	cfg.OuterMiddlewares = []httputil.Middleware{
		appkitotel.Middleware(
			appkitotel.WithTracerProvider(tp),
			appkitotel.WithMeterProvider(mp),
		),
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	svc.Mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := svc.Shutdown(ctx); err != nil {
			t.Errorf("shutdown: %v", err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("server returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop after shutdown")
		}
	})

	deadline := time.Now().Add(2 * time.Second)
	for !svc.Running() {
		if time.Now().After(deadline) {
			t.Fatal("service did not start within timeout")
		}

		time.Sleep(time.Millisecond)
	}

	resp, err := http.Get("http://" + svc.Addr().String() + "/users/42")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	select {
	case err := <-errCh:
		t.Fatalf("server error: %v", err)
	default:
	}

	spans := spanExporter.GetSpans().Snapshots()
	if len(spans) == 0 {
		t.Fatal("no spans recorded")
	}
	spanName := spans[len(spans)-1].Name()
	t.Logf("span name: %q", spanName)
	if spanName != "GET /users/{id}" {
		t.Errorf("span name = %q, want %q (pattern lost through OuterMiddlewares)", spanName, "GET /users/{id}")
	}

	var rm metricdata.ResourceMetrics
	if err := mr.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}
	routeAttr := ""
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.server.request.duration" {
				continue
			}
			hist, ok := m.Data.(metricdata.Histogram[float64])
			if !ok {
				continue
			}
			for _, dp := range hist.DataPoints {
				if v, ok := dp.Attributes.Value(attribute.Key("http.route")); ok {
					routeAttr = v.AsString()
				}
			}
		}
	}
	t.Logf("http.route attribute: %q", routeAttr)
	if routeAttr != "/users/{id}" {
		t.Errorf("http.route = %q, want %q (route attribute lost through OuterMiddlewares)", routeAttr, "/users/{id}")
	}
}
