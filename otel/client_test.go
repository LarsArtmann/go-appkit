package otel

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func newClientTestProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()

	return sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter)), exporter
}

func flushedSpans(
	t *testing.T,
	provider *sdktrace.TracerProvider,
	exporter *tracetest.InMemoryExporter,
) []tracetest.SpanStub {
	t.Helper()

	if err := provider.ForceFlush(t.Context()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	return exporter.GetSpans()
}

func TestTransport_ClientSpanAndTraceparentInjection(t *testing.T) {
	t.Parallel()

	provider, exporter := newClientTestProvider(t)

	var traceparent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent = r.Header.Get("Traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Transport: Transport(WithTracerProvider(provider))}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/downstream", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	spans := flushedSpans(t, provider, exporter)

	if len(spans) != 1 {
		t.Fatalf("exported %d spans, want exactly one client span", len(spans))
	}

	span := spans[0]

	if span.Name != "HTTP GET" {
		t.Errorf("client span named %q, want `HTTP GET`", span.Name)
	}

	if span.SpanKind != trace.SpanKindClient {
		t.Errorf("span kind %v, want client", span.SpanKind)
	}

	parts := strings.Split(traceparent, "-")

	if len(parts) != 4 || parts[0] != "00" {
		t.Fatalf("server received traceparent %q, want a valid W3C value", traceparent)
	}

	if span.SpanContext.TraceID().String() != parts[1] {
		t.Errorf("client span trace %s does not continue into the injected traceparent %s",
			span.SpanContext.TraceID(), parts[1])
	}
}

func TestTransport_SpanEndsOnBodyClose(t *testing.T) {
	t.Parallel()

	provider, exporter := newClientTestProvider(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "payload")
	}))
	defer server.Close()

	client := &http.Client{Transport: Transport(WithTracerProvider(provider))}

	resp, err := mustDo(t, client, server.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if spans := flushedSpans(t, provider, exporter); len(spans) != 0 {
		t.Fatalf("client span exported before body close, want it held open (got %d spans)", len(spans))
	}

	_, _ = io.Copy(io.Discard, resp.Body)

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	if spans := flushedSpans(t, provider, exporter); len(spans) != 1 {
		t.Fatalf("exported %d spans after body close, want exactly one", len(spans))
	}
}

func TestTransport_BaseTransportReceivesRequest(t *testing.T) {
	t.Parallel()

	baseCalled := false
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		baseCalled = true

		return &http.Response{StatusCode: http.StatusTeapot, Body: http.NoBody, Header: http.Header{}}, nil
	})

	client := &http.Client{Transport: Transport(WithBaseTransport(base))}

	resp, err := mustDo(t, client, "http://example.invalid/downstream")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if !baseCalled {
		t.Fatal("WithBaseTransport round-tripper never saw the request")
	}

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("status %d, want the base transport's response", resp.StatusCode)
	}
}

func TestTransport_NoOpWithoutProvider(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Transport: Transport()}

	resp, err := mustDo(t, client, server.URL)
	if err != nil {
		t.Fatalf("get through the unconfigured transport: %v", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status %d, want the plain pass-through response", resp.StatusCode)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// mustDo issues GET via client.Do with the test context — the noctx-safe
// shape this suite's clients use.
func mustDo(t *testing.T, client *http.Client, url string) (*http.Response, error) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	return client.Do(req)
}
