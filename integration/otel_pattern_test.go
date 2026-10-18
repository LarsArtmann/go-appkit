package integration_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/httputil"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

		shutdownErr := svc.Shutdown(ctx)
		if shutdownErr != nil {
			t.Errorf("shutdown: %v", shutdownErr)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"http://"+svc.Addr().String()+"/users/42", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
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

	routeAttr := httpRouteAttributeValue(t, mr)
	t.Logf("http.route attribute: %q", routeAttr)
	if routeAttr != "/users/{id}" {
		t.Errorf("http.route = %q, want %q (route attribute lost through OuterMiddlewares)", routeAttr, "/users/{id}")
	}
}

// TestRecoveryExceptionEventThroughAppkitChain pins the
// events-not-descriptions contract through PUBLISHED tags: when
// appkitotel.Recovery runs inside appkitotel.Middleware (innermost, via
// ExtraMiddlewares), otelhttp overwrites the span status DESCRIPTION at
// span end — but the exception EVENT (type/message/stacktrace) survives,
// and that event is what SigNoz's Exceptions view indexes. Asserting
// descriptions would rot the day otelhttp changes its clobber; asserting
// the event pins the consumer-visible contract.
func TestRecoveryExceptionEventThroughAppkitChain(t *testing.T) {
	t.Parallel()

	svc, spanExporter := newTracedTestService(t)
	svc.Mux.HandleFunc("GET /panic", func(http.ResponseWriter, *http.Request) {
		panic("integration boom")
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr := svc.Shutdown(ctx)
		if shutdownErr != nil {
			t.Errorf("shutdown: %v", shutdownErr)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"http://"+svc.Addr().String()+"/panic", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}

	var serverSpan sdktrace.ReadOnlySpan

	for _, span := range spanExporter.GetSpans().Snapshots() {
		if span.Name() == "GET /panic" {
			serverSpan = span
		}
	}

	if serverSpan == nil {
		t.Fatal("no server span named GET /panic recorded")
	}

	events := serverSpan.Events()
	if len(events) != 1 {
		t.Fatalf("server span has %d events, want exactly 1 exception event (events: %v)", len(events), events)
	}
	if events[0].Name != "exception" {
		t.Errorf("event name = %q, want %q", events[0].Name, "exception")
	}

	attrs := lo.SliceToMap(events[0].Attributes, func(kv attribute.KeyValue) (string, string) {
		return string(kv.Key), kv.Value.String()
	})
	if !strings.Contains(attrs["exception.message"], "integration boom") {
		t.Errorf("exception.message = %q, want it to contain the panic value", attrs["exception.message"])
	}
	if attrs["exception.type"] == "" {
		t.Error("exception.type attribute missing from the exception event")
	}
	if attrs["exception.stacktrace"] == "" {
		t.Error("exception.stacktrace attribute missing from the exception event")
	}
	if serverSpan.Status().Code != codes.Error {
		t.Errorf("span status code = %v, want Error", serverSpan.Status().Code)
	}
}

// httpRouteAttributeValue collects the manual reader once and returns the
// http.route attribute observed on the http.server.request.duration
// histogram (empty string when absent).
func httpRouteAttributeValue(t *testing.T, mr *sdkmetric.ManualReader) string {
	t.Helper()

	var rm metricdata.ResourceMetrics
	err := mr.Collect(context.Background(), &rm)
	if err != nil {
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

	return routeAttr
}
