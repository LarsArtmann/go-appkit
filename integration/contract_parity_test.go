package integration_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/errorpages"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	errorfamily "github.com/larsartmann/go-error-family"
	"go.opentelemetry.io/otel"
)

// TestOneOtelSetupPerProcess pins the documented contract: appkitotel.Setup
// registers the PROCESS-GLOBAL tracer/meter providers and propagator, and a
// second Setup call silently overwrites them. Two stacks both calling Setup
// (go-appkit/otel and go-cqrs-lite/otel) is therefore a hazard — exactly one
// owner per process; the last Setup wins and the first owner's
// instrumentation silently reads the wrong provider. This test pins the
// overwrite so the one-owner rule can never rot quietly.
func TestOneOtelSetupPerProcess(t *testing.T) {
	// NOT parallel: mutates process globals (restored on cleanup).
	prevTP := otel.GetTracerProvider()
	prevMP := otel.GetMeterProvider()
	prevProp := otel.GetTextMapPropagator()

	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetMeterProvider(prevMP)
		otel.SetTextMapPropagator(prevProp)
	})

	before := otel.GetTracerProvider()

	provider, err := appkitotel.Setup(appkitotel.WithService("integration-owner-test", "", ""))
	if err != nil {
		t.Fatalf("first Setup: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := provider.Shutdown(ctx)
		if err != nil {
			t.Errorf("provider shutdown: %v", err)
		}
	}()

	first := otel.GetTracerProvider()
	if first == before {
		t.Fatal("Setup must register the global tracer provider (the documented wiring)")
	}

	// The hazard half: a second Setup REPLACES the globals — the first owner
	// keeps a provider nothing reads anymore. This is why exactly ONE Setup
	// owner is allowed per process.
	provider2, err := appkitotel.Setup(appkitotel.WithService("integration-owner-test-2", "", ""))
	if err != nil {
		t.Fatalf("second Setup: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := provider2.Shutdown(ctx)
		if err != nil {
			t.Errorf("provider2 shutdown: %v", err)
		}
	}()

	if otel.GetTracerProvider() == first {
		t.Error("second Setup must overwrite the global (that is why exactly ONE owner is allowed)")
	}
}

// TestErrorpagesFamilyStatusParity pins the classification contract: for
// every go-error-family family, the HTTP status the errorpages handler
// writes equals what appkit.HTTPStatus classifies. A drift between the two
// mappings breaks every consumer's error contract.
func TestErrorpagesFamilyStatusParity(t *testing.T) {
	t.Parallel()

	families := []struct {
		name string
		err  error
	}{
		{"rejection", errorfamily.NewRejection("itest.code", "rejection")},
		{"conflict", errorfamily.NewConflict("itest.code", "conflict")},
		{"transient", errorfamily.NewTransient("itest.code", "transient")},
		{"corruption", errorfamily.NewCorruption("itest.code", "corruption")},
		{"infrastructure", errorfamily.NewInfrastructure("itest.code", "infrastructure")},
	}

	for _, family := range families {
		want := appkit.HTTPStatus(family.err)
		if want == 0 {
			t.Fatalf("%s: appkit.HTTPStatus returned 0 — classification contract broken", family.name)
		}

		handler := errorpages.Handler(family.err, errorpages.Config{})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/boom", nil))

		if rec.Code != want {
			t.Errorf("%s: errorpages wrote %d, appkit.HTTPStatus classifies %d — mappings drifted",
				family.name, rec.Code, want)
		}
	}
}

// TestErrorpagesUnclassifiedThroughLiveService pins the 500 fallthrough
// through a live appkit Service with the errorpages Wrap mounted: an error
// no family classification matches must surface as 500, not leak a status
// from an arbitrary upstream path.
func TestErrorpagesUnclassifiedThroughLiveService(t *testing.T) {
	t.Parallel()

	registerHealthOff := false
	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.RegisterHealth = &registerHealthOff
	cfg.DrainDelay = appkit.NoDrainDelay

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	errorpages.Wrap(svc.Mux, errorpages.Config{})
	svc.Mux.HandleFunc("GET /boom", func(w http.ResponseWriter, _ *http.Request) {
		panic("boom") // recovered by the default stack, rendered by Wrap
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := svc.Shutdown(ctx)
		if err != nil {
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

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+svc.Addr().String()+"/boom", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	defer resp.Body.Close() //nolint:errcheck // read-side drain: the status is already captured above

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("recovered panic through Wrap: status = %d, want 500", resp.StatusCode)
	}
}
