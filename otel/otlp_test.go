package otel

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

// otlpReceipts collects what the stub collector received, one entry per
// OTLP/HTTP export request.
type otlpReceipts struct {
	mu      sync.Mutex
	traces  []*collectortracepb.ExportTraceServiceRequest
	metrics []*collectormetricpb.ExportMetricsServiceRequest
}

func (r *otlpReceipts) spanCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := 0
	for _, req := range r.traces {
		for _, rs := range req.ResourceSpans {
			for _, ss := range rs.ScopeSpans {
				total += len(ss.Spans)
			}
		}
	}

	return total
}

func (r *otlpReceipts) metricCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := 0
	for _, req := range r.metrics {
		for _, rm := range req.ResourceMetrics {
			for _, sm := range rm.ScopeMetrics {
				total += len(sm.Metrics)
			}
		}
	}

	return total
}

// newOTLPCollector runs an in-process OTLP/HTTP receiver (the /v1/traces and
// /v1/metrics endpoints SigNoz exposes) and records every decoded export
// request.
func newOTLPCollector(t *testing.T) (*httptest.Server, *otlpReceipts) {
	t.Helper()

	receipts := &otlpReceipts{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		switch r.URL.Path {
		case "/v1/traces":
			req := &collectortracepb.ExportTraceServiceRequest{}
			if err := proto.Unmarshal(body, req); err != nil {
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			receipts.mu.Lock()
			receipts.traces = append(receipts.traces, req)
			receipts.mu.Unlock()

			writeOTLPResponse(t, w, &collectortracepb.ExportTraceServiceResponse{})
		case "/v1/metrics":
			req := &collectormetricpb.ExportMetricsServiceRequest{}
			if err := proto.Unmarshal(body, req); err != nil {
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			receipts.mu.Lock()
			receipts.metrics = append(receipts.metrics, req)
			receipts.mu.Unlock()

			writeOTLPResponse(t, w, &collectormetricpb.ExportMetricsServiceResponse{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	t.Cleanup(server.Close)

	return server, receipts
}

func writeOTLPResponse(t *testing.T, w http.ResponseWriter, msg proto.Message) {
	t.Helper()

	body, err := proto.Marshal(msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")

	_, _ = w.Write(body)
}

// waitForOTLPReceipts polls until want requests of the given kind arrived,
// because batch/periodic flush timing is asynchronous by design.
func waitForOTLPReceipts(t *testing.T, count func() int, want int) {
	t.Helper()

	const waitStep = 10 * time.Millisecond

	for range 300 {
		if count() >= want {
			return
		}

		time.Sleep(waitStep)
	}

	t.Fatalf("collector received %d requests, want >= %d", count(), want)
}

// TestWithOTLP_ExportsSpansToEndpoint pins the production path: spans flow
// to the configured OTLP/HTTP endpoint carrying the service resource —
// byte-for-byte what SigNoz ingests.
func TestWithOTLP_ExportsSpansToEndpoint(t *testing.T) {
	t.Parallel()

	server, receipts := newOTLPCollector(t)

	provider, err := Setup(
		WithService("otlp-svc", "2.0.0", "pod-9"),
		WithEnvironment("staging"),
		WithOTLP(WithOTLPEndpoint(server.URL)),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "exported")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	waitForOTLPReceipts(t, receipts.spanCount, 1)
	waitForOTLPReceipts(t, func() int {
		receipts.mu.Lock()
		defer receipts.mu.Unlock()

		return len(receipts.traces)
	}, 1)

	receipts.mu.Lock()
	first := receipts.traces[0]
	receipts.mu.Unlock()

	resAttrs := attribute.NewSet(first.ResourceSpans[0].Resource.GetAttributes()...)
	if got, ok := resAttrs.Value(attribute.Key("service.name")); !ok || got.AsString() != "otlp-svc" {
		t.Errorf("exported service.name = %q (found=%v), want otlp-svc", got.AsString(), ok)
	}

	if got, ok := resAttrs.Value(attribute.Key("deployment.environment")); !ok || got.AsString() != "staging" {
		t.Errorf("exported deployment.environment = %q (found=%v), want staging", got.AsString(), ok)
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestWithOTLP_ExportsMetricsToEndpoint proves the metric signal rides the
// same OTLP wiring: a counter recorded against the provider's meter reaches
// the collector.
func TestWithOTLP_ExportsMetricsToEndpoint(t *testing.T) {
	t.Parallel()

	server, receipts := newOTLPCollector(t)

	provider, err := Setup(
		WithService("otlp-metrics", "", ""),
		WithOTLP(WithOTLPEndpoint(server.URL)),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	counter, err := provider.AsMeterProvider().Meter("test").Int64Counter("otlp_test_requests_total")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}

	counter.Add(t.Context(), 1)

	if err := provider.AsMeterProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	waitForOTLPReceipts(t, receipts.metricCount, 1)

	receipts.mu.Lock()
	first := receipts.metrics[0]
	receipts.mu.Unlock()

	if len(first.ResourceMetrics) == 0 ||
		len(first.ResourceMetrics[0].ScopeMetrics) == 0 ||
		len(first.ResourceMetrics[0].ScopeMetrics[0].Metrics) == 0 {
		t.Fatalf("metric export empty: %v", first)
	}

	if got := first.ResourceMetrics[0].ScopeMetrics[0].Metrics[0].GetName(); got != "otlp_test_requests_total" {
		t.Errorf("exported metric = %q, want otlp_test_requests_total", got)
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_EnvEndpointAutoEnablesOTLP is the zero-code contract: nothing
// but OTEL_EXPORTER_OTLP_ENDPOINT in the environment, and Setup wires both
// signals itself.
func TestSetup_EnvEndpointAutoEnablesOTLP(t *testing.T) {
	t.Parallel()

	server, receipts := newOTLPCollector(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)

	provider, err := Setup(
		WithService("env-svc", "", ""),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "env-driven")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	waitForOTLPReceipts(t, receipts.spanCount, 1)

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_ExplicitExporterBeatsEnvEndpoint: code wins per signal — a
// service with its own exporter must not double-export to the environment's
// collector, so the OTLP path stays off for that signal.
func TestSetup_ExplicitExporterBeatsEnvEndpoint(t *testing.T) {
	t.Parallel()

	cold, coldReceipts := newOTLPCollector(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", cold.URL)

	exporter := &tracetest.InMemoryExporter{}

	provider, err := Setup(
		WithService("manual-svc", "", ""),
		WithSpanExporter(exporter),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "in-memory")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	if spans := exporter.GetSpans(); len(spans) != 1 {
		t.Fatalf("explicit exporter recorded %d spans, want 1", len(spans))
	}

	if got := coldReceipts.spanCount(); got != 0 {
		t.Errorf("env collector received %d spans, want 0 — explicit exporter must win", got)
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_OTLPMetricsShareExplicitSpanExporter: WithSpanExporter alone
// switches off only the trace signal; the metric signal still follows the
// environment. Signal independence is what makes partial migrations safe.
func TestSetup_OTLPMetricsShareExplicitSpanExporter(t *testing.T) {
	t.Parallel()

	server, receipts := newOTLPCollector(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)

	exporter, err := stdouttrace.New(stdouttrace.WithWriter(io.Discard))
	if err != nil {
		t.Fatalf("stdout exporter: %v", err)
	}

	provider, err := Setup(
		WithService("mixed-svc", "", ""),
		WithSpanExporter(exporter),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	counter, err := provider.AsMeterProvider().Meter("test").Int64Counter("mixed_test_total")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}

	counter.Add(t.Context(), 1)

	if err := provider.AsMeterProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	waitForOTLPReceipts(t, receipts.metricCount, 1)

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_ExplicitMetricReaderBeatsOTLP: WithOTLP + WithMetricReader
// composes — the custom reader owns the metric signal, OTLP owns traces.
func TestSetup_ExplicitMetricReaderBeatsOTLP(t *testing.T) {
	t.Parallel()

	server, receipts := newOTLPCollector(t)

	reader := sdkmetric.NewManualReader()

	provider, err := Setup(
		WithService("compose-svc", "", ""),
		WithOTLP(WithOTLPEndpoint(server.URL)),
		WithMetricReader(reader),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	counter, err := provider.AsMeterProvider().Meter("test").Int64Counter("compose_test_total")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}

	counter.Add(t.Context(), 1)

	var data sdkmetric.ResourceMetrics
	if err := reader.Collect(t.Context(), &data); err != nil {
		t.Fatalf("collect: %v", err)
	}

	found := false
	for _, sm := range data.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "compose_test_total" {
				found = true
			}
		}
	}

	if !found {
		t.Error("manual reader missing compose_test_total — explicit reader must own the metric signal")
	}

	if got := receipts.metricCount(); got != 0 {
		t.Errorf("collector received %d metric requests, want 0", got)
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_ResourceFromEnvFillsGaps: OTEL_RESOURCE_ATTRIBUTES and
// OTEL_SERVICE_NAME flow into the resource without code, and explicit code
// configuration still wins on conflict.
func TestSetup_ResourceFromEnvFillsGaps(t *testing.T) {
	t.Parallel()

	exporter := &tracetest.InMemoryExporter{}
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=qa-team,region=eu-1")

	provider, err := Setup(
		WithService("env-resource-svc", "", ""),
		WithSpanExporter(exporter),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "resourced")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	resAttrs := attribute.NewSet(spans[0].Resource.Attributes()...)
	if got, ok := resAttrs.Value(attribute.Key("deployment.environment")); !ok || got.AsString() != "qa-team" {
		t.Errorf("deployment.environment = %q (found=%v), want qa-team from OTEL_RESOURCE_ATTRIBUTES", got.AsString(), ok)
	}

	if got, ok := resAttrs.Value(attribute.Key("service.name")); !ok || got.AsString() != "env-resource-svc" {
		t.Errorf("service.name = %q (found=%v), want explicit WithService to beat env", got.AsString(), ok)
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_EnvServiceNameWhenNoWithService: OTEL_SERVICE_NAME alone names
// the service — completes the zero-code story for consumers that discover
// their identity from the platform.
func TestSetup_EnvServiceNameWhenNoWithService(t *testing.T) {
	t.Parallel()

	exporter := &tracetest.InMemoryExporter{}
	t.Setenv("OTEL_SERVICE_NAME", "platform-named")

	provider, err := Setup(WithSpanExporter(exporter), WithoutGlobalRegistration())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "named-by-env")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	resAttrs := attribute.NewSet(spans[0].Resource.Attributes()...)
	if got, ok := resAttrs.Value(attribute.Key("service.name")); !ok || got.AsString() != "platform-named" {
		t.Errorf("service.name = %q (found=%v), want platform-named from OTEL_SERVICE_NAME", got.AsString(), ok)
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestSetup_EnvSamplerHonored: OTEL_TRACES_SAMPLER works without a
// WithSampler option — production can head-sample by deployment env alone.
func TestSetup_EnvSamplerHonored(t *testing.T) {
	t.Parallel()

	exporter := &tracetest.InMemoryExporter{}
	t.Setenv("OTEL_TRACES_SAMPLER", "always_off")

	provider, err := Setup(WithService("sampled-env", "", ""), WithSpanExporter(exporter), WithoutGlobalRegistration())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "dropped-by-env")
	span.End()

	if err := provider.AsTracerProvider().ForceFlush(t.Context()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	if spans := exporter.GetSpans(); len(spans) != 0 {
		t.Errorf("OTEL_TRACES_SAMPLER=always_off exported %d spans, want 0", len(spans))
	}

	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestOTLPSentinelsClassified pins the classification contract of the OTLP
// setup sentinel (Infrastructure family: reaching the collector is an
// environmental concern).
func TestOTLPSentinelsClassified(t *testing.T) {
	t.Parallel()

	errorfamilytest.AssertFamily(t, errOTLPSetup, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, errOTLPSetup, "otel.otlp_exporter_failed")
	errorfamilytest.AssertHTTPStatus(t, errOTLPSetup, 503)
}
