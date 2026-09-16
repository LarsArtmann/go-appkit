# OTEL Pattern-Propagation Pin Test — land with the release train

**Created:** 2026-09-16 · **Type:** release-train artifact (recoverable recipe +
verbatim test). **Owner:** TODO_LIST P2 "OTEL REGRESSION" release train.

The 2026-09-16 pattern-propagation fix lives in httputil master (`ff44c5f`)
but ships in NO tag. The integration module pins PUBLISHED tags by charter, so
the pin test below **cannot compile there until the train runs**. It lived only
in `/tmp/appkit-otel-verify/` (ephemeral) — this document is its durable home.

## When to land it

Executing the TODO_LIST P2 release train (tag httputil → bump core + otel →
re-tag otel):

1. Copy `TestSpanNameAndRouteThroughAppkitOuterMiddlewares` from this document
   into `integration/integration_test.go` (package `integration`).
2. `integration/go.mod` gains: `go.opentelemetry.io/otel/sdk`,
   `go.opentelemetry.io/otel/sdk/metric`, `go.opentelemetry.io/otel/sdk/trace`,
   and (transitively) the otel module requirement at the new tag.
   The otel SDK deps make the module heavier — acceptable: the charter pins
   what consumers resolve, and the test pins the documented composition.
3. Bump the pins (core, otel, httputil via core) to the train tags, then
   `go test -race -count=1 ./...` inside `integration/`.
4. Expected FAILURES before the train, PASS after: this is the regression pin
   that closes the blind spot which let the bug ship (all 23 otel-module tests
   wrap the mux directly; none tested the documented `OuterMiddlewares` wiring).
5. Delete the otel README known-issue block in the same change.

## Provenance

- Authored 2026-09-16 in `/tmp/appkit-otel-verify` (A/B harness): run A with
  published httputil v1.1.1 → span `"GET"`, no route attribute (production bug
  reproduced); run B identical except the local fixed httputil → span
  `"GET /users/{id}"` + `http.route /users/{id}`. Zero added allocations.
- Session report: `docs/status/2026-09-16_08-40_otel-pattern-propagation-fix-and-self-review.md` §a-5.

## The test (verbatim, module `scratch`)

```go
package scratch

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
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
// attribute must survive the whole chain.
func TestSpanNameAndRouteThroughAppkitOuterMiddlewares(t *testing.T) {
	spanExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spanExporter))
	mr := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(mr))

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "127.0.0.1:0"
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
		t.Fatalf("NewService: %v", err)
	}
	svc.Mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = svc.Shutdown(shutdownCtx)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for !svc.Running() {
		if time.Now().After(deadline) {
			t.Fatal("service never started")
		}
		time.Sleep(5 * time.Millisecond)
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
```

Adaptation notes for `integration/`: rename the package, drop the `Running()`
poll in favor of the module's existing wait helpers, keep
`RegisterHealth: &false` + `NoDrainDelay` (integration's other tests use a 1ms
explicit `DrainDelay` — either works), and remember the span-name assertion
carries the method prefix while `http.route` does not.
