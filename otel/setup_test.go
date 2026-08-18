package otel

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// These tests exercise Setup, which registers the process-global providers.
// They therefore run SEQUENTIALLY and restore the previous globals when
// done; do not add t.Parallel to anything calling Setup without
// WithoutGlobalRegistration.

// withRestoredGlobals saves the OTel globals and restores them on cleanup.
func withRestoredGlobals(t *testing.T) {
	t.Helper()

	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()

	prevMP := otel.GetMeterProvider()

	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetMeterProvider(prevMP)
		otel.SetTextMapPropagator(prevProp)
	})
}

func TestSetup_SpansCarryServiceResource(t *testing.T) {
	withRestoredGlobals(t)

	exporter := &tracetest.InMemoryExporter{}

	provider, err := Setup(
		WithService("orders-api", "1.2.3", "pod-7"),
		WithSpanExporter(exporter),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "work")
	span.End()

	// Read spans after an explicit flush but BEFORE Shutdown:
	// tracetest.InMemoryExporter.Shutdown resets its buffer.
	err = provider.AsTracerProvider().ForceFlush(t.Context())
	if err != nil {
		t.Fatalf("flush: %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans, want 1", len(spans))
	}

	err = provider.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	resAttrs := attribute.NewSet(spans[0].Resource.Attributes()...)
	for key, want := range map[string]string{
		"service.name":        "orders-api",
		"service.version":     "1.2.3",
		"service.instance.id": "pod-7",
	} {
		if got, ok := resAttrs.Value(attribute.Key(key)); !ok || got.AsString() != want {
			t.Errorf("resource %s = %q (found=%v), want %q", key, got.AsString(), ok, want)
		}
	}
}

func TestSetup_RegistersGlobalsForMiddleware(t *testing.T) {
	withRestoredGlobals(t)

	exporter := &tracetest.InMemoryExporter{}

	provider, err := Setup(
		WithService("integration", "", ""),
		WithSpanExporter(exporter),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Middleware with NO explicit providers must pick up the globals Setup
	// registered — the primary consumer wiring.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(Middleware()(mux))
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/ping") //nolint:noctx // single-shot integration request
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	_ = resp.Body.Close()

	// The server span ends when the response finishes streaming, which can
	// land microseconds after http.Get returns. ForceFlush in a poll loop
	// covers both the enqueue race and the 5s batch timeout; Shutdown only
	// after the span is visible.
	spans := waitForSpans(t, provider, exporter)
	if len(spans) != 1 || spans[0].Name != "GET /ping" {
		t.Fatalf("global-wired middleware recorded %v, want one GET /ping span", spans)
	}

	err = provider.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// waitForSpans force-flushes the provider in a poll loop until at least one
// span is exported (batch export is asynchronous; the default batch timeout
// is 5s) or the poll window elapses.
const pollInterval = 10 * time.Millisecond

func waitForSpans(
	t *testing.T,
	provider *Provider,
	exporter *tracetest.InMemoryExporter,
) []tracetest.SpanStub {
	t.Helper()

	for range 100 {
		_ = provider.AsTracerProvider().ForceFlush(t.Context())

		if spans := exporter.GetSpans(); len(spans) > 0 {
			return spans
		}

		time.Sleep(pollInterval)
	}

	t.Fatal("no spans were exported within the poll window")

	return nil
}

func TestSetup_WithoutGlobalRegistrationLeavesGlobalsAlone(t *testing.T) {
	withRestoredGlobals(t)

	before := otel.GetTracerProvider()

	provider, err := Setup(
		WithService("isolated", "", ""),
		WithoutGlobalRegistration(),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if after := otel.GetTracerProvider(); after != before {
		t.Error("WithoutGlobalRegistration must leave the global TracerProvider untouched")
	}

	if provider.AsTracerProvider() == nil || provider.AsMeterProvider() == nil {
		t.Error("isolated Provider must still be fully constructed")
	}

	err = provider.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestSetup_StdoutExporterPrettyPrints(t *testing.T) {
	withRestoredGlobals(t)

	buf := &bytes.Buffer{}

	provider, err := Setup(
		WithService("stdout-svc", "", ""),
		WithStdoutExporter(buf),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "visible")
	span.End()

	err = provider.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if !strings.Contains(buf.String(), "visible") {
		t.Errorf("stdout output misses the span name; got:\n%s", buf.String())
	}
}

func TestSetup_SamplerRespected(t *testing.T) {
	withRestoredGlobals(t)

	exporter := &tracetest.InMemoryExporter{}

	provider, err := Setup(
		WithService("sampled", "", ""),
		WithSpanExporter(exporter),
		WithSampler(sdktrace.NeverSample()),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, span := provider.AsTracerProvider().Tracer("test").Start(t.Context(), "dropped")
	span.End()

	err = provider.Shutdown(t.Context())
	if err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if spans := exporter.GetSpans(); len(spans) != 0 {
		t.Errorf("NeverSample exported %d spans, want 0", len(spans))
	}
}

// failingExporter fails Shutdown, proving Provider.Shutdown joins errors
// and keeps the sentinel findable via errors.Is.
type failingExporter struct {
	tracetest.InMemoryExporter
}

var errExporterShutdown = errors.New("exporter refused shutdown")

func (f *failingExporter) Shutdown(context.Context) error {
	return errExporterShutdown
}

func TestSetup_ShutdownJoinsExporterErrors(t *testing.T) {
	withRestoredGlobals(t)

	provider, err := Setup(
		WithService("failing", "", ""),
		WithSpanExporter(&failingExporter{}),
	)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	err = provider.Shutdown(t.Context())

	if !errors.Is(err, errShutdown) {
		t.Errorf("shutdown error = %v, want joined errShutdown", err)
	}

	if !errors.Is(err, errExporterShutdown) {
		t.Errorf("shutdown error = %v, want the exporter failure wrapped", err)
	}
}
