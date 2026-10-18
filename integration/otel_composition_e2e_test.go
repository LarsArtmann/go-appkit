package integration_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	appkithealth "github.com/larsartmann/go-appkit/health"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-health"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
	"github.com/larsartmann/httputil"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// e2eGet issues a GET against the service under test with a request
// context — the noctx-safe shape for these E2Es. Fails the test on error.
func e2eGet(t *testing.T, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build request %s: %v", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do %s: %v", url, err)
	}

	return resp
}

// exceptionEvent returns the span's exception event, or nil.'s exception event, or nil.
func exceptionEvent(span sdktrace.ReadOnlySpan) *sdktrace.Event {
	for i, event := range span.Events() {
		if event.Name == "exception" {
			return &span.Events()[i]
		}
	}

	return nil
}

// spanNamed returns the last span with the given name, or nil.
//
//nolint:ireturn // ReadOnlySpan is the only shape a traced span has
func spanNamed(spans []sdktrace.ReadOnlySpan, name string) sdktrace.ReadOnlySpan {
	var found sdktrace.ReadOnlySpan

	for _, span := range spans {
		if span.Name() == name {
			found = span
		}
	}

	return found
}

// waitRunning blocks until the service answers Running() or fails the test.
func waitRunning(t *testing.T, svc *appkit.Service) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for !svc.Running() {
		if time.Now().After(deadline) {
			t.Fatal("service did not start within timeout")
		}

		time.Sleep(time.Millisecond)
	}
}

// TestBothExceptionFlavorsThroughAppkitChain pins BOTH exception flavors —
// panic (Recovery middleware) and handled error (RecordError) — through one
// real appkit Service wired exactly as the otel README prescribes: otel
// middleware in OuterMiddlewares (the span wraps everything), Recovery in
// ExtraMiddlewares (inside the span). SigNoz's Exceptions view indexes both
// flavors the same way; this is the composition-level pin that the module's
// own unit tests cannot make (they lack the default middleware stack).
func TestBothExceptionFlavorsThroughAppkitChain(t *testing.T) {
	t.Parallel()

	spanExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spanExporter))

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay
	cfg.OuterMiddlewares = []httputil.Middleware{
		appkitotel.Middleware(appkitotel.WithTracerProvider(tp)),
	}
	cfg.ExtraMiddlewares = []httputil.Middleware{
		appkitotel.Recovery(slog.New(slog.DiscardHandler)),
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handledErr := errorfamily.NewRejection("itest.payment", "payment gateway timed out")

	svc.Mux.HandleFunc("GET /panic", func(http.ResponseWriter, *http.Request) {
		panic("integration panic boom")
	})
	svc.Mux.HandleFunc("GET /handled", func(w http.ResponseWriter, r *http.Request) {
		appkitotel.RecordError(r.Context(), handledErr)
		http.Error(w, "checkout failed", http.StatusInternalServerError)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := svc.Shutdown(ctx); shutdownErr != nil {
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

	waitRunning(t, svc)

	base := "http://" + svc.Addr().String()

	resp := e2eGet(t, base+"/panic")
	if err != nil {
		t.Fatalf("panic request: %v", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("panic status = %d, want 500", resp.StatusCode)
	}

	resp = e2eGet(t, base+"/handled")
	if err != nil {
		t.Fatalf("handled request: %v", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("handled status = %d, want 500", resp.StatusCode)
	}

	spans := spanExporter.GetSpans().Snapshots()

	panicSpan := spanNamed(spans, "GET /panic")
	if panicSpan == nil {
		t.Fatalf("no GET /panic span (have %d spans)", len(spans))
	}

	handledSpan := spanNamed(spans, "GET /handled")
	if handledSpan == nil {
		t.Fatalf("no GET /handled span (have %d spans)", len(spans))
	}

	panicEvent := exceptionEvent(panicSpan)
	if panicEvent == nil {
		t.Fatal("panic span carries no exception event — SigNoz's Exceptions view would miss this failure")
	}

	handledEvent := exceptionEvent(handledSpan)
	if handledEvent == nil {
		t.Fatal("handled-error span carries no exception event — RecordError did not survive the chain")
	}

	if panicSpan.Status().Code != codes.Error {
		t.Errorf("panic span status code = %v, want Error", panicSpan.Status().Code)
	}

	if handledSpan.Status().Code != codes.Error {
		t.Errorf("handled span status code = %v, want Error", handledSpan.Status().Code)
	}
}

// sseRouteDatapoint reports whether the collected metrics carry an
// http.server.request.duration datapoint attributed to the given route.
func sseRouteDatapoint(rm *metricdata.ResourceMetrics, route string) bool {
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
				if v, ok := dp.Attributes.Value(attribute.Key("http.route")); ok && v.AsString() == route {
					return true
				}
			}
		}
	}

	return false
}

// TestSSEStreamSpanThroughAppkitOuterMiddlewares pins the NoTimeout span-end
// contract at the composition level: an SSE stream served through the otel
// middleware (OuterMiddlewares) produces ONE server span whose lifetime is
// the STREAM, not the response write — the span must not exist while the
// stream is open and must exist (named "/sse", route-attributed) after the
// client disconnects. The name carries no method prefix because
// realtime.Mount registers a method-less Handle("/sse") — unlike mux
// patterns such as "GET /users/{id}". House protocol applies:
// subscriber-count sync before the broadcast and a timeout client (the
// 2026-10-13 hang class).
func TestSSEStreamSpanThroughAppkitOuterMiddlewares(t *testing.T) {
	t.Parallel()

	spanExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spanExporter))
	mr := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(mr))

	hub := realtime.NewHub()

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:         freeAddr(t),
		ReadTimeout:  appkit.NoTimeout,
		WriteTimeout: appkit.NoTimeout,
		DrainDelay:   time.Millisecond,
		OuterMiddlewares: []httputil.Middleware{
			appkitotel.Middleware(
				appkitotel.WithTracerProvider(tp),
				appkitotel.WithMeterProvider(mp),
			),
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	realtime.Mount(svc.Mux, "/sse", hub)
	testkit.Serve(t, svc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"http://"+svc.Addr().String()+"/sse", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	deadline := time.Now().Add(2 * time.Second)
	for hub.SubscriberCount() != 1 {
		if time.Now().After(deadline) {
			t.Fatal("subscriber never registered with the hub")
		}

		time.Sleep(time.Millisecond)
	}

	hub.Broadcast(sse.Event{Event: "ping", Data: "{}", ID: sse.NewEventID("evt-1")})

	events := ssetest.MustReadNEvents(t, resp.Body, 1)
	if events[0].Type != "ping" {
		t.Fatalf("event type = %q, want ping", events[0].Type)
	}

	// The stream is still open — otelhttp must NOT have ended the span.
	if spans := spanExporter.GetSpans().Snapshots(); spanNamed(spans, "/sse") != nil {
		t.Fatal("/sse span ended while the stream was still open — the NoTimeout span-end contract broke")
	}

	_ = resp.Body.Close()

	deadline = time.Now().Add(2 * time.Second)
	for {
		if span := spanNamed(spanExporter.GetSpans().Snapshots(), "/sse"); span != nil {
			break
		}

		if time.Now().After(deadline) {
			t.Fatal("/sse span never ended after the stream closed")
		}

		time.Sleep(time.Millisecond)
	}

	var rm metricdata.ResourceMetrics
	if err := mr.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	if !sseRouteDatapoint(&rm, "/sse") {
		t.Error(
			"no http.server.request.duration datapoint with http.route=/sse — route attribution lost on the SSE path",
		)
	}
}

// TestDashboardRoutesTracedVsFiltered pins the otel×health-dashboard tracing
// matrix through PUBLISHED tags. The otel middleware filters /health and
// /health/ UNCONDITIONALLY (probe volume, no distributed context) — and the
// go-health-dashboard's DEFAULT routes live under /health (page /health, SSE
// /health/sse, metrics/trend/export/introspect below it — the latter group
// option-gated), so the default dashboard is NOT traced, while the
// kubelet-convention routes (/healthz, /readyz, /startupz) sit OUTSIDE the
// filter and ARE traced. The matrix this test pins (defaults):
//
//	/health (dashboard page)  → filtered, no span
//	/healthz (kubelet live)   → traced
//	/readyz (kubelet ready)   → traced
//
// Consumers who WANT the dashboard traced must move it off /health via
// dashboard.Routes — the filter is deliberately name-based, not
// registration-aware.
func TestDashboardRoutesTracedVsFiltered(t *testing.T) {
	t.Parallel()

	spanExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spanExporter))

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay
	cfg.OuterMiddlewares = []httputil.Middleware{
		appkitotel.Middleware(appkitotel.WithTracerProvider(tp)),
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	mounted, err := appkithealth.New(
		health.New(do.New()),
		appkithealth.WithDashboard(),
	)
	if err != nil {
		t.Fatalf("appkithealth.New: %v", err)
	}

	mounted.RegisterRoutes(svc.Mux)

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := svc.Shutdown(ctx); shutdownErr != nil {
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

	waitRunning(t, svc)

	base := "http://" + svc.Addr().String()

	type expectation struct {
		path     string
		spanName string
		wantSpan bool
		wantHTML bool
	}

	cases := []expectation{
		{"/health", "/health", false, true},
		{"/healthz", "/healthz", true, false},
		{"/readyz", "/readyz", true, false},
	}

	for _, tc := range cases {
		resp := e2eGet(t, base+tc.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s: status = %d, want 200", tc.path, resp.StatusCode)
		}

		if tc.wantHTML && !strings.Contains(string(body), "<") {
			t.Errorf("GET %s: expected an HTML document, got %q", tc.path, truncate(string(body), 80))
		}
	}

	spans := spanExporter.GetSpans().Snapshots()

	for _, tc := range cases {
		span := spanNamed(spans, tc.spanName)

		if tc.wantSpan && span == nil {
			t.Errorf("GET %s: expected a server span (outside the /health filter), none recorded", tc.path)
		}

		if !tc.wantSpan && span != nil {
			t.Errorf("GET %s: expected NO span (health filter), found one — the unconditional filter broke", tc.path)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n] + "..."
}

// TestErrorpagesBehindOtelMiddleware pins the errorpages×otel composition
// contract as it ACTUALLY composes (source-verified against errorpages.Wrap:
// it replaces only the mux's bare 404/405 — it never renders panics or
// handler errors). The honest matrix through PUBLISHED tags:
//
//   - PANIC → recovered by the stack, plain 500 text/plain, exception event
//     on the error span (Recovery's recording is what feeds SigNoz).
//   - HANDLED error → the handler records the exception (RecordError) AND
//     renders the pretty classified page via errorpages.Handler — pretty
//     page for the user, indexed exception for SigNoz, status from the
//     error-family classification (Infrastructure → 503 here).
//   - 404 → pretty HTML from Wrap (no span unless unfiltered — /health-style
//     filtering does not apply here).
func TestErrorpagesBehindOtelMiddleware(t *testing.T) {
	t.Parallel()

	spanExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spanExporter))

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay
	cfg.OuterMiddlewares = []httputil.Middleware{
		appkitotel.Middleware(appkitotel.WithTracerProvider(tp)),
	}
	cfg.ExtraMiddlewares = []httputil.Middleware{
		appkitotel.Recovery(slog.New(slog.DiscardHandler)),
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	errorpages.Wrap(svc.Mux, errorpages.Config{})

	unclassified := errorfamily.NewInfrastructure("itest.unhandled", "unhandled infrastructure failure")
	wantHandledStatus := appkit.HTTPStatus(unclassified)

	svc.Mux.HandleFunc("GET /panic", func(http.ResponseWriter, *http.Request) {
		panic("integration errorpages boom")
	})
	svc.Mux.HandleFunc("GET /handled", func(w http.ResponseWriter, r *http.Request) {
		appkitotel.RecordError(r.Context(), unclassified)
		errorpages.Handler(unclassified, errorpages.Config{}).ServeHTTP(w, r)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := svc.Shutdown(ctx); shutdownErr != nil {
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

	waitRunning(t, svc)

	base := "http://" + svc.Addr().String()

	resp := e2eGet(t, base+"/panic")
	if err != nil {
		t.Fatalf("GET /panic: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("GET /panic: status = %d, want 500", resp.StatusCode)
	}

	if strings.Contains(string(body), "<html") {
		t.Error(
			"GET /panic: body is an HTML page — Wrap must NOT render panics (it owns 404/405 only); contract drifted",
		)
	}

	panicSpan := spanNamed(spanExporter.GetSpans().Snapshots(), "GET /panic")
	if panicSpan == nil {
		t.Fatal("GET /panic: no server span recorded")
	}

	if exceptionEvent(panicSpan) == nil {
		t.Error("GET /panic: span carries no exception event")
	}

	resp = e2eGet(t, base+"/handled")
	if err != nil {
		t.Fatalf("GET /handled: %v", err)
	}

	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != wantHandledStatus {
		t.Fatalf("GET /handled: status = %d, want the family classification %d", resp.StatusCode, wantHandledStatus)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("GET /handled: content-type = %q, want a pretty HTML page", ct)
	}

	if !strings.Contains(string(body), "<") {
		t.Errorf("GET /handled: body is not HTML: %q", truncate(string(body), 80))
	}

	handledSpan := spanNamed(spanExporter.GetSpans().Snapshots(), "GET /handled")
	if handledSpan == nil {
		t.Fatal("GET /handled: no server span recorded")
	}

	if exceptionEvent(handledSpan) == nil {
		t.Error(
			"GET /handled: span carries no exception event — the Exceptions view would be empty while users saw the pretty page",
		)
	}

	resp = e2eGet(t, base+"/definitely-not-a-route")
	if err != nil {
		t.Fatalf("GET /definitely-not-a-route: %v", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /definitely-not-a-route: status = %d, want 404", resp.StatusCode)
	}
}
