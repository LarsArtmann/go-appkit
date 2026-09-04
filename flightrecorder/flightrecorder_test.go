package flightrecorder_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	appkitfr "github.com/larsartmann/go-appkit/flightrecorder"
	fr "github.com/larsartmann/go-flightrecorder"
)

// httpGetURL issues a context-aware GET against a test server (noctx forbids
// http.Get in tests).
func httpGetURL(t *testing.T, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("building GET %s: %v", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}

	return resp
}

// httpPostJSON issues a context-aware JSON POST against a test server (noctx
// forbids http.Post in tests).
func httpPostJSON(t *testing.T, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("building POST %s: %v", url, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}

	return resp
}

// recorderMu serializes tests that call Start/Stop because Go's
// runtime/trace allows only ONE active flight recorder per process.
var recorderMu sync.Mutex

// newTestRecorder creates a flight recorder that writes snapshots to a temp
// file. The caller must hold recorderMu and call Start/Stop/Close within the
// serialized section.
func newTestRecorder(t *testing.T) (*fr.Recorder, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "trace.out")

	rec, err := fr.New(fr.WithFile(path))
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

	return rec, path
}

// newBufferRecorder creates a flight recorder that writes snapshots to a
// shared *bytes.Buffer. Use this for multi-capture tests where you need to
// verify that content grows between captures (fr.WithFile opens a lazyFile
// that caches its handle, making file-deletion-based verification unreliable).
func newBufferRecorder(t *testing.T) (*fr.Recorder, *bytes.Buffer) {
	t.Helper()

	buf := &bytes.Buffer{}

	rec, err := fr.New(fr.WithWriter(buf))
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

	return rec, buf
}

// startRecorder locks the process-global recorder, starts it, and returns
// a cleanup function that stops and closes it.
func startRecorder(t *testing.T, rec *fr.Recorder) func() {
	t.Helper()

	recorderMu.Lock()

	err := rec.Start()
	if err != nil {
		recorderMu.Unlock()
		t.Fatalf("rec.Start() error: %v", err)
	}

	return func() {
		rec.Stop()
		_ = rec.Close()
		recorderMu.Unlock()
	}
}

// newStartedRecorder returns a fresh recorder that is already started and
// stopped/closed automatically when the test ends, plus its trace output
// path.
func newStartedRecorder(t *testing.T) (*fr.Recorder, string) {
	t.Helper()

	rec, tracePath := newTestRecorder(t)
	t.Cleanup(startRecorder(t, rec))

	return rec, tracePath
}

// assertTraceWritten verifies that a non-empty trace file was created.
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

// assertTraceNotWritten verifies that no trace file was created (or is empty).
func assertTraceNotWritten(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		return // File doesn't exist — pass
	}

	if info.Size() > 0 {
		t.Fatalf("expected no trace file but %s has %d bytes", path, info.Size())
	}
}

// --- Middleware trigger tests ---

func TestMiddleware_CapturesOnError(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, fr.OnError())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	assertTraceWritten(t, tracePath)
}

func TestMiddleware_CapturesOnLatency(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, fr.OnLatency(50*time.Millisecond))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(80 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/slow", nil)
	handler.ServeHTTP(rr, req)

	assertTraceWritten(t, tracePath)
}

func TestMiddleware_DoesNotCaptureOnSuccess(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, fr.OnErrorOrLatency(1*time.Second))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ok", nil)
	handler.ServeHTTP(rr, req)

	assertTraceNotWritten(t, tracePath)
}

func TestMiddleware_CapturesOnErrorOrLatency_ErrorCase(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, fr.OnErrorOrLatency(1*time.Second))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr, req)

	assertTraceWritten(t, tracePath)
}

func TestMiddleware_CapturesOnErrorOrLatency_LatencyCase(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, fr.OnErrorOrLatency(50*time.Millisecond))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(80 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/slow", nil)
	handler.ServeHTTP(rr, req)

	assertTraceWritten(t, tracePath)
}

// --- Middleware option tests ---

func TestMiddleware_WithErrorThreshold(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	// With threshold 400, a 404 should count as error
	mw := appkitfr.Middleware(rec, fr.OnError(),
		appkitfr.WithErrorThreshold(400),
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing", nil)
	handler.ServeHTTP(rr, req)

	assertTraceWritten(t, tracePath)
}

func TestMiddleware_WithErrorThreshold_NotTriggeredBelow(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	// Default threshold is 500, so 404 alone should NOT trigger OnError
	mw := appkitfr.Middleware(rec, fr.OnError())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing", nil)
	handler.ServeHTTP(rr, req)

	assertTraceNotWritten(t, tracePath)
}

func TestMiddleware_WithLogger(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mw := appkitfr.Middleware(rec, fr.OnError(),
		appkitfr.WithLogger(logger),
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/data", nil)
	handler.ServeHTTP(rr, req)

	assertTraceWritten(t, tracePath)

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("expected log output, got empty string")
	}

	if !bytes.Contains(buf.Bytes(), []byte("flightrecorder: trace snapshot captured")) {
		t.Fatalf("expected log to contain capture message, got: %s", logOutput)
	}
}

func TestMiddleware_WithAutoResetDisabled(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, fr.OnError(),
		appkitfr.WithAutoReset(false),
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	// First request should capture
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr1, req1)

	assertTraceWritten(t, tracePath)

	// Delete the trace file so we can verify second request does NOT write
	err := os.Remove(tracePath)
	if err != nil {
		t.Fatalf("failed to remove trace file: %v", err)
	}

	// Second request should NOT capture (once-latch not reset)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr2, req2)

	assertTraceNotWritten(t, tracePath)
}

func TestMiddleware_AutoResetDefault_AllowsMultipleCaptures(t *testing.T) {
	rec, buf := newBufferRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	mw := appkitfr.Middleware(rec, fr.OnError())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	// First request
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr1, req1)

	firstSize := buf.Len()
	if firstSize == 0 {
		t.Fatal("expected first capture to write trace data, got 0 bytes")
	}

	// Second request should also capture (autoReset is default true)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr2, req2)

	secondSize := buf.Len()
	if secondSize <= firstSize {
		t.Fatalf("expected second capture to grow buffer: first=%d, second=%d", firstSize, secondSize)
	}
}

func TestMiddleware_NilTriggerNeverCaptures(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mw := appkitfr.Middleware(rec, nil)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr, req)

	assertTraceNotWritten(t, tracePath)
}

// --- Handler tests ---

func TestSnapshotHandler_Success(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	handler := appkitfr.SnapshotHandler(rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/debug/snapshot", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	assertTraceWritten(t, tracePath)

	var resp struct {
		Status string `json:"status"`
	}

	err := json.UnmarshalRead(rr.Body, &resp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "snapshot captured" {
		t.Fatalf("expected status 'snapshot captured', got %q", resp.Status)
	}
}

func TestSnapshotHandler_WorksWithJsonContentType(t *testing.T) {
	rec, _ := newTestRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	handler := appkitfr.SnapshotHandler(rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/debug/snapshot", nil)
	handler.ServeHTTP(rr, req)

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
}

func TestSnapshotHandler_ResetBeforeSnapshot(t *testing.T) {
	rec, buf := newBufferRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	// First, consume the once-latch via a direct Snapshot
	// (simulating middleware already captured)
	err := rec.Snapshot(t.Context())
	if err != nil {
		t.Fatalf("first snapshot error: %v", err)
	}

	firstSize := buf.Len()
	if firstSize == 0 {
		t.Fatal("expected first snapshot to write trace data, got 0 bytes")
	}

	// Now use the handler — it should Reset first, then capture again
	handler := appkitfr.SnapshotHandler(rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/debug/snapshot", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// The handler's Reset should allow a second capture
	secondSize := buf.Len()
	if secondSize <= firstSize {
		t.Fatalf("expected second capture to grow buffer: first=%d, second=%d", firstSize, secondSize)
	}
}

// --- Mount tests ---

func TestMount_RegistersHandler(t *testing.T) {
	rec, tracePath := newStartedRecorder(t)

	mux := http.NewServeMux()
	appkitfr.Mount(mux, "POST /debug/snapshot", rec)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp := httpPostJSON(t, ts.URL+"/debug/snapshot")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var result struct {
		Status string `json:"status"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Status != "snapshot captured" {
		t.Fatalf("expected 'snapshot captured', got %q", result.Status)
	}

	assertTraceWritten(t, tracePath)
}

// --- Integration: middleware + handler together ---

func TestMiddleware_ThenHandler_ManualSnapshotAfterAutoCapture(t *testing.T) {
	rec, buf := newBufferRecorder(t)

	cleanup := startRecorder(t, rec)
	defer cleanup()

	mw := appkitfr.Middleware(rec, fr.OnError())

	mux := http.NewServeMux()

	// Register an endpoint that returns 500, wrapped with the middleware
	mux.Handle("GET /api/fail", mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})))

	// Register manual snapshot endpoint
	appkitfr.Mount(mux, "POST /debug/snapshot", rec)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Trigger automatic capture via middleware
	resp1 := httpGetURL(t, ts.URL+"/api/fail")

	_ = resp1.Body.Close()

	firstSize := buf.Len()
	if firstSize == 0 {
		t.Fatal("expected middleware capture to write trace data, got 0 bytes")
	}

	// Now manually snapshot — handler resets the latch, should capture again
	resp2 := httpPostJSON(t, ts.URL+"/debug/snapshot")
	defer func() { _ = resp2.Body.Close() }()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	secondSize := buf.Len()
	if secondSize <= firstSize {
		t.Fatalf("expected manual capture to grow buffer: first=%d, second=%d", firstSize, secondSize)
	}
}
