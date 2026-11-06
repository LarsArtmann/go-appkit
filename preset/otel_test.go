package preset_test

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"

	appkitotel "github.com/larsartmann/go-appkit/otel"
	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/testkit"
)

// recordingExporter captures exported spans so tests can observe whether the
// otel middleware recorded anything — without touching global state.
type recordingExporter struct {
	mu     sync.Mutex
	spans  int
	shutDo bool
}

func (e *recordingExporter) ExportSpans(_ context.Context, spans []trace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans += len(spans)
	return nil
}

func (e *recordingExporter) Shutdown(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.shutDo = true
	return nil
}

func (e *recordingExporter) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.spans
}

func (e *recordingExporter) wasShutdown() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.shutDo
}

// TestOTEL_CodeEnabled_RecordsSpans: Enabled=true wires the provider, mounts
// the middleware (spans appear at the exporter), and the provider Shutdown
// joins ShutdownHooks (the exporter is shut down and spans flushed).
func TestOTEL_CodeEnabled_RecordsSpans(t *testing.T) {

	exporter := &recordingExporter{}

	cfg := testConfig()
	cfg.OTEL = &preset.OTELConfig{
		Enabled:     true,
		ServiceName: "preset-test",
		Options:     []appkitotel.SetupOption{appkitotel.WithSpanExporter(exporter)},
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	resp, err := http.Get(ts.FullChainURL + "/healthz")
	if err != nil {
		t.Fatalf("get /healthz: %v", err)
	}
	_ = resp.Body.Close()

	if err := ts.Shutdown(t.Context()); err != nil {
		t.Fatalf("graceful shutdown: %v", err)
	}

	if exporter.count() == 0 {
		t.Error("no spans exported — the otel middleware is not in the chain")
	}

	if !exporter.wasShutdown() {
		t.Error("exporter never shut down — the provider Shutdown hook is not wired")
	}
}

// TestOTEL_CodeDisabled_EnvOn_ZeroSpans: explicit code wins both directions —
// Enabled=false keeps the leg off even with APPKIT_OTEL=on.
func TestOTEL_CodeDisabled_EnvOn_ZeroSpans(t *testing.T) {
	t.Setenv("APPKIT_OTEL", "on")

	exporter := &recordingExporter{}

	cfg := testConfig()
	cfg.OTEL = &preset.OTELConfig{
		Enabled: false,
		Options: []appkitotel.SetupOption{appkitotel.WithSpanExporter(exporter)},
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	resp, err := http.Get(ts.FullChainURL + "/healthz")
	if err != nil {
		t.Fatalf("get /healthz: %v", err)
	}
	_ = resp.Body.Close()

	if err := ts.Shutdown(t.Context()); err != nil {
		t.Fatalf("graceful shutdown: %v", err)
	}

	if exporter.count() != 0 {
		t.Errorf("%d spans exported with the leg code-disabled — env overrode code", exporter.count())
	}
}

// TestOTEL_EnvOn_Boots: the env switch alone enables the leg — Setup runs
// with zero options (export off until something asks) and the service serves.
func TestOTEL_EnvOn_Boots(t *testing.T) {
	t.Setenv("APPKIT_OTEL", "on")

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp, err := http.Get(ts + "/healthz")
	if err != nil {
		t.Fatalf("get /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("/healthz = %d, want 200", resp.StatusCode)
	}
}
