package integration_test

import (
	"log/slog"
	"testing"

	appkit "github.com/larsartmann/go-appkit"
	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/httputil"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newTracedTestService builds an appkit Service wired the way a traced
// consumer composes the otel module: otelhttp Middleware on the outer
// chain (server spans) and Recovery inside the span (exception events),
// with an in-memory span exporter for span assertions. Health
// registration is off and the drain delay zero so composition tests
// start fast and shut down immediately.
func newTracedTestService(t *testing.T) (*appkit.Service, *tracetest.InMemoryExporter) {
	t.Helper()

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

	return svc, spanExporter
}
