package flightrecorderhealth_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/samber/do/v2"
)

// Static test-error sentinels — err113 forbids dynamic errors.New("...") in tests.
var (
	errTestConnectionRefused = errors.New("connection refused")
	errTestTimeout           = errors.New("timeout")
	errTestServiceDown       = errors.New("service down")
)

// recorderMu serializes tests that call Start/Stop because Go's
// runtime/trace allows only ONE active flight recorder per process.
var recorderMu sync.Mutex

// healthSvc is a configurable service implementing do.HealthcheckerWithContext.
type healthSvc struct {
	err error
}

func (h *healthSvc) HealthCheck(_ context.Context) error {
	return h.err
}

func newTestRecorder(t *testing.T) (*fr.Recorder, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "trace.out")

	rec, err := fr.New(
		fr.WithFile(path),
		fr.WithMinAge(50*time.Millisecond),
		fr.WithMaxBytes(1<<20),
	)
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

	return rec, path
}

func newBufferRecorder(t *testing.T) (*fr.Recorder, *bytes.Buffer) {
	t.Helper()

	buf := &bytes.Buffer{}

	rec, err := fr.New(
		fr.WithWriter(buf),
		fr.WithMinAge(50*time.Millisecond),
		fr.WithMaxBytes(1<<20),
	)
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

	return rec, buf
}

func startRecorder(t *testing.T, rec *fr.Recorder) func() {
	t.Helper()

	recorderMu.Lock()

	err := rec.Start()
	if err != nil {
		recorderMu.Unlock()
		t.Fatalf("rec.Start() error: %v", err)
	}

	// Let the trace buffer accumulate data before any capture attempt.
	time.Sleep(100 * time.Millisecond)

	return func() {
		rec.Stop()
		_ = rec.Close()
		recorderMu.Unlock()
	}
}

// registerSvc registers a named health-checkable service in the injector and
// eagerly invokes it.
func registerSvc(injector do.Injector, name string, err error) {
	do.ProvideNamed(injector, name, func(_ do.Injector) (*healthSvc, error) {
		return &healthSvc{err: err}, nil
	})
	_, _ = do.InvokeNamed[*healthSvc](injector, name)
}

func assertTraceWritten(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("trace file not created at %s: %v", path, err)
	}

	if info.Size() == 0 {
		t.Fatalf("trace file is empty at %s", path)
	}
}

func assertTraceNotWritten(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if info.Size() > 0 {
		t.Fatalf("expected no trace file but %s has %d bytes", path, info.Size())
	}
}

// drainAndStop stops the recorder, which drains in-flight async captures via
// wg.Wait() internally. Called before assertions to ensure the async goroutine
// has completed. Stop/Close are idempotent so the deferred cleanup is safe.
func drainAndStop(rec *fr.Recorder) {
	rec.Stop()
	_ = rec.Close()
}

// --- Checkable tests ---

func TestCheckable_HealthyWhenEnabled(t *testing.T) {
	rec, _ := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	c := frhealth.NewCheckable(rec)

	err := c.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when recorder is enabled, got: %v", err)
	}
}

func TestCheckable_UnhealthyWhenNotEnabled(t *testing.T) {
	rec, _ := newTestRecorder(t)

	c := frhealth.NewCheckable(rec)

	err := c.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error when recorder is not enabled, got nil")
	}
}

func TestCheckable_NilReceiver(t *testing.T) {
	var c *frhealth.Checkable

	err := c.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error for nil Checkable, got nil")
	}
}

func TestCheckable_Name(t *testing.T) {
	rec, _ := newTestRecorder(t)

	c := frhealth.NewCheckable(rec, frhealth.WithCheckableName("my-recorder"))

	if got := c.Name(); got != "my-recorder" {
		t.Fatalf("expected name 'my-recorder', got %q", got)
	}
}

func TestCheckable_DefaultName(t *testing.T) {
	rec, _ := newTestRecorder(t)

	c := frhealth.NewCheckable(rec)

	if got := c.Name(); got != "flight-recorder" {
		t.Fatalf("expected default name 'flight-recorder', got %q", got)
	}
}

// --- Trigger tests ---

func TestTrigger_CapturesOnHealthCheckFailure(t *testing.T) {
	rec, tracePath := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	trigger := frhealth.NewTrigger(rec)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	rec.Reset()

	results := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	if results["database"] == nil {
		t.Fatal("expected database error in results")
	}

	drainAndStop(rec)
	assertTraceWritten(t, tracePath)
}

func TestTrigger_DoesNotCaptureOnSuccess(t *testing.T) {
	rec, tracePath := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	trigger := frhealth.NewTrigger(rec)

	injector := do.New()
	registerSvc(injector, "database", nil)
	registerSvc(injector, "cache", nil)

	rec.Reset()

	results := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	drainAndStop(rec)
	assertTraceNotWritten(t, tracePath)
}

func TestTrigger_NilRecorder_PassThrough(t *testing.T) {
	trigger := frhealth.NewTrigger(nil)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	results := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	if results["database"] == nil {
		t.Fatal("expected pass-through to return injector results directly")
	}
}

func TestTrigger_NilReceiver_PassThrough(t *testing.T) {
	var trigger *frhealth.Trigger

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	results := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	if results["database"] == nil {
		t.Fatal("expected nil Trigger to pass through to injector")
	}
}

func TestTrigger_WithLogger(t *testing.T) {
	rec, _ := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	trigger := frhealth.NewTrigger(rec,
		frhealth.WithTriggerLogger(logger),
		frhealth.WithServiceName("orders-recorder"),
	)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	rec.Reset()

	_ = trigger.RecordHealthCheckWithContext(context.Background(), injector)

	drainAndStop(rec)

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("expected log output, got empty string")
	}

	if !bytes.Contains(buf.Bytes(), []byte("trace snapshot triggered by health check")) {
		t.Fatalf("expected log to contain trigger message, got: %s", logOutput)
	}

	if !bytes.Contains(buf.Bytes(), []byte("database")) {
		t.Fatalf("expected log to contain failing service name, got: %s", logOutput)
	}

	if !bytes.Contains(buf.Bytes(), []byte("orders-recorder")) {
		t.Fatalf("expected log to contain service name, got: %s", logOutput)
	}
}

func TestTrigger_WithServiceName_NotLoggedWhenEmpty(t *testing.T) {
	rec, _ := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	trigger := frhealth.NewTrigger(rec,
		frhealth.WithTriggerLogger(logger),
		// no WithServiceName
	)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	rec.Reset()

	_ = trigger.RecordHealthCheckWithContext(context.Background(), injector)

	drainAndStop(rec)

	if bytes.Contains(buf.Bytes(), []byte("service=")) {
		t.Fatalf("expected no 'service' attribute when service name empty, got: %s", buf.String())
	}
}

func TestTrigger_ConcurrentCooldownIsRaceFree(t *testing.T) {
	rec, _ := newBufferRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	trigger := frhealth.NewTrigger(rec,
		frhealth.WithCooldown(50*time.Millisecond),
	)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	rec.Reset()

	const goroutines = 8
	const iterations = 25

	var wg sync.WaitGroup

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				_ = trigger.RecordHealthCheckWithContext(context.Background(), injector)
			}
		}()
	}

	wg.Wait()

	drainAndStop(rec)
}

func TestTrigger_WithCooldown(t *testing.T) {
	rec, buf := newBufferRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	trigger := frhealth.NewTrigger(rec,
		frhealth.WithCooldown(1*time.Hour),
	)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)

	// First capture — should fire (cooldown not yet started).
	rec.Reset()

	_ = trigger.RecordHealthCheckWithContext(context.Background(), injector)

	drainAndStop(rec)

	firstSize := buf.Len()
	if firstSize == 0 {
		t.Fatal("expected first capture to write trace data")
	}

	// Restart recorder for the second batch.
	err := rec.Start()
	if err != nil {
		t.Fatalf("rec.Start() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Second capture — should be suppressed by cooldown.
	rec.Reset()

	_ = trigger.RecordHealthCheckWithContext(context.Background(), injector)

	drainAndStop(rec)

	secondSize := buf.Len()
	if secondSize > firstSize {
		t.Fatalf("expected cooldown to suppress second capture: first=%d, second=%d", firstSize, secondSize)
	}
}

func TestTrigger_CustomTriggerFunc(t *testing.T) {
	rec, tracePath := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	// Use OnAlways — should capture even when health checks pass.
	trigger := frhealth.NewTrigger(rec,
		frhealth.WithTriggerFunc(fr.OnAlways()),
	)

	injector := do.New()
	registerSvc(injector, "database", nil)

	rec.Reset()

	_ = trigger.RecordHealthCheckWithContext(context.Background(), injector)

	drainAndStop(rec)
	assertTraceWritten(t, tracePath)
}

func TestTrigger_RecordsResultsFaithfully(t *testing.T) {
	rec, _ := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	trigger := frhealth.NewTrigger(rec)

	injector := do.New()
	registerSvc(injector, "database", errTestConnectionRefused)
	registerSvc(injector, "cache", nil)
	registerSvc(injector, "queue", errTestTimeout)

	rec.Reset()

	got := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}

	for _, name := range []string{"database", "cache", "queue"} {
		if _, ok := got[name]; !ok {
			t.Fatalf("expected result for %q, missing", name)
		}
	}

	if got["database"] == nil {
		t.Fatal("expected database to have an error")
	}

	if got["cache"] != nil {
		t.Fatal("expected cache to have no error")
	}

	if got["queue"] == nil {
		t.Fatal("expected queue to have an error")
	}
}

// --- Register tests ---

func TestRegister_CreatesServiceInInjector(t *testing.T) {
	injector := do.New()

	rec, _ := newTestRecorder(t)

	c := frhealth.Register(injector, rec, "my-recorder")

	if c == nil {
		t.Fatal("expected non-nil Checkable from Register")
	}

	if c.Name() != "my-recorder" {
		t.Fatalf("expected name 'my-recorder', got %q", c.Name())
	}

	retrieved, err := do.InvokeNamed[*frhealth.Checkable](injector, "my-recorder")
	if err != nil {
		t.Fatalf("failed to invoke registered service: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected non-nil service from injector")
	}
}

func TestRegister_DefaultName(t *testing.T) {
	injector := do.New()

	rec, _ := newTestRecorder(t)

	c := frhealth.Register(injector, rec, "")

	if c.Name() != "flight-recorder" {
		t.Fatalf("expected default name 'flight-recorder', got %q", c.Name())
	}
}

// --- Integration tests ---

func TestIntegration_HealthDashboardShowsRecorder(t *testing.T) {
	rec, _ := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	injector := do.New()

	frhealth.Register(injector, rec, "flight-recorder")

	results := injector.HealthCheckWithContext(context.Background())

	if len(results) == 0 {
		t.Fatal("expected at least one health-check result from injector")
	}

	recorderErr, ok := results["flight-recorder"]
	if !ok {
		t.Fatalf("expected 'flight-recorder' in results, got: %v", results)
	}

	if recorderErr != nil {
		t.Fatalf("expected healthy recorder, got error: %v", recorderErr)
	}
}

func TestIntegration_TriggerWithFailingService(t *testing.T) {
	rec, tracePath := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	injector := do.New()
	registerSvc(injector, "failing-svc", errTestServiceDown)

	trigger := frhealth.NewTrigger(rec)

	rec.Reset()

	results := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	if results["failing-svc"] == nil {
		t.Fatal("expected failing-svc to have an error")
	}

	drainAndStop(rec)
	assertTraceWritten(t, tracePath)
}

func TestIntegration_CheckableAppearsAsUnhealthyWhenStopped(t *testing.T) {
	rec, _ := newTestRecorder(t)

	injector := do.New()

	frhealth.Register(injector, rec, "flight-recorder")

	// Don't start the recorder — it should appear as unhealthy.
	results := injector.HealthCheckWithContext(context.Background())

	recorderErr, ok := results["flight-recorder"]
	if !ok {
		t.Fatalf("expected 'flight-recorder' in results, got: %v", results)
	}

	if recorderErr == nil {
		t.Fatal("expected unhealthy recorder (not started), got nil error")
	}

	if fmt.Sprint(recorderErr) == "" {
		t.Fatal("expected non-empty error message")
	}
}
