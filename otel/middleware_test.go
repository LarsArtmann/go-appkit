package otel

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/httputil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// newRecordingProvider returns a tracer provider wired to an in-memory
// exporter, plus a cleanup that flushes before the caller inspects spans.
func newRecordingProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()

	exporter := &tracetest.InMemoryExporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
	)

	t.Cleanup(func() {
		_ = tp.Shutdown(t.Context())
	})

	return tp, exporter
}

func newInstrumentedServer(t *testing.T, tp *sdktrace.TracerProvider, opts ...MiddlewareOption) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /ctx", func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		w.Header().Set("X-Trace-Id", span.SpanContext().TraceID().String())
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(Middleware(append([]MiddlewareOption{WithTracerProvider(tp)}, opts...)...)(mux))
	t.Cleanup(server.Close)

	return server
}

// fetchResult carries what the tests assert on — deliberately NOT a
// *http.Response, so the request body never escapes the fetching function
// (the body is closed where it is opened).
type fetchResult struct {
	status int
	header http.Header
}

func fetch(t *testing.T, url string) fetchResult {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}

	// No test reads the body; status and headers stay valid after Close.
	_ = resp.Body.Close()

	return fetchResult{status: resp.StatusCode, header: resp.Header}
}

func TestMiddleware_SatisfiesHttputilMiddleware(t *testing.T) {
	t.Parallel()

	// The explicit type makes the contract assertion readable: Middleware
	// must remain assignable to the middleware type core's config expects.
	var middleware httputil.Middleware = Middleware() //nolint:staticcheck // QF1011: the explicit type IS the assertion
	_ = middleware
}

func TestMiddleware_RecordsServerSpanNamedByRoutePattern(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp)

	result := fetch(t, server.URL+"/users/42")
	if result.status != http.StatusOK {
		t.Fatalf("status = %d", result.status)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	span := spans[0]
	if span.Name != "GET /users/{id}" {
		t.Errorf("span name = %q, want the ServeMux pattern GET /users/{id}", span.Name)
	}

	if span.SpanKind != trace.SpanKindServer {
		t.Errorf("span kind = %v, want SERVER", span.SpanKind)
	}

	if !hasAttribute(span.Attributes, "http.response.status_code") {
		t.Errorf("span attributes miss http.response.status_code: %v", span.Attributes)
	}
}

func TestMiddleware_HandlerSeesSpanInContext(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp)

	result := fetch(t, server.URL+"/ctx")
	traceID := result.header.Get("X-Trace-Id")
	if traceID == "" || strings.Contains(traceID, "00000000") {
		t.Fatalf("handler saw no valid trace ID in its context: %q", traceID)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	if got := spans[0].SpanContext.TraceID().String(); got != traceID {
		t.Errorf("handler trace ID %q != recorded span trace ID %q", traceID, got)
	}
}

func TestMiddleware_HealthEndpointsNeverInstrumented(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp)

	fetch(t, server.URL+"/health")

	if spans := exporter.GetSpans(); len(spans) != 0 {
		t.Errorf("health endpoint recorded %d spans, want 0", len(spans))
	}
}

func TestMiddleware_FilteredPathsOption(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp, WithFilteredPaths("/metrics"))

	fetch(t, server.URL+"/metrics")

	if spans := exporter.GetSpans(); len(spans) != 0 {
		t.Errorf("filtered path /metrics recorded %d spans, want 0", len(spans))
	}

	fetch(t, server.URL+"/users/7")

	if spans := exporter.GetSpans(); len(spans) != 1 {
		t.Errorf("unfiltered path recorded %d spans, want 1", len(spans))
	}
}

func TestMiddleware_ErrorStatusMarksSpanError(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp)

	result := fetch(t, server.URL+"/boom")
	if result.status != http.StatusInternalServerError {
		t.Fatalf("status = %d", result.status)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	if spans[0].Status.Code != codes.Error {
		t.Errorf("span status = %v, want error for a 500 response", spans[0].Status.Code)
	}
}

func TestMiddleware_ContinuesIncomingTraceContext(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp)

	clientTP, clientExporter := newRecordingProvider(t)
	tracer := clientTP.Tracer("client")

	ctx, parent := tracer.Start(t.Context(), "client.call")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/users/1", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	NewTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	_ = resp.Body.Close()
	parent.End()

	serverSpan := singleSpan(t, exporter)
	parentSpan := singleSpan(t, clientExporter)

	if serverSpan.Parent.SpanID() != parentSpan.SpanContext.SpanID() {
		t.Errorf(
			"server span parent %v != client span %v — trace context was not continued",
			serverSpan.Parent.SpanID(), parentSpan.SpanContext.SpanID(),
		)
	}

	if serverSpan.SpanContext.TraceID() != parentSpan.SpanContext.TraceID() {
		t.Error("server span trace ID differs from client trace ID")
	}
}

func TestMiddleware_PublicEndpointDistrustsRemoteParent(t *testing.T) {
	t.Parallel()

	tp, exporter := newRecordingProvider(t)
	server := newInstrumentedServer(t, tp, WithPublicEndpoint())

	remoteSpanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x1},
		SpanID:  trace.SpanID{0x2},
		Remote:  true,
	})
	ctx := trace.ContextWithSpanContext(t.Context(), remoteSpanCtx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/users/1", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	NewTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	_ = resp.Body.Close()

	span := singleSpan(t, exporter)

	if span.Parent.IsValid() {
		t.Errorf("public endpoint adopted remote parent %v — must start a new root", span.Parent)
	}

	if len(span.Links) != 1 || span.Links[0].SpanContext.SpanID() != remoteSpanCtx.SpanID() {
		t.Errorf("remote span context not recorded as a link: %v", span.Links)
	}
}

func TestMiddleware_NoProvidersConfiguredIsNoOp(t *testing.T) {
	t.Parallel()

	// No WithTracerProvider option and no global registration: the
	// middleware must serve requests without recording anything and without
	// crashing — OTel stays opt-in.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(Middleware()(mux))
	t.Cleanup(server.Close)

	if result := fetch(t, server.URL+"/users/3"); result.status != http.StatusOK {
		t.Fatalf("status = %d", result.status)
	}
}

func singleSpan(t *testing.T, exporter *tracetest.InMemoryExporter) tracetest.SpanStub {
	t.Helper()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want exactly 1", len(spans))
	}

	return spans[0]
}

// hasAttribute reports whether attrs carries key.
func hasAttribute(attrs []attribute.KeyValue, key string) bool {
	for _, kv := range attrs {
		if string(kv.Key) == key {
			return true
		}
	}

	return false
}
