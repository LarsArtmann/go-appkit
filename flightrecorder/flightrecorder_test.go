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

// gatedWriter is an io.Writer whose Write blocks until released. It proves
// the middleware does not pay trace-write latency on the request path: the
// response must complete while a capture write is stuck on the gate.
type gatedWriter struct {
	writeStarted chan struct{}
	writeDone    chan struct{}
	releaseOnce  sync.Once
	release      chan struct{}
	mu           sync.Mutex
	data         []byte
}

func newGatedWriter() *gatedWriter {
	return &gatedWriter{
		writeStarted: make(chan struct{}),
		writeDone:    make(chan struct{}),
		release:      make(chan struct{}),
	}
}

func (w *gatedWriter) Write(p []byte) (int, error) {
	select {
	case <-w.writeStarted:
	default:
		close(w.writeStarted)
	}

	<-w.release

	w.mu.Lock()
	w.data = append(w.data, p...)
	w.mu.Unlock()

	close(w.writeDone)

	return len(p), nil
}

func (w *gatedWriter) unblock() {
	w.releaseOnce.Do(func() { close(w.release) })
}

func (w *gatedWriter) bytesWritten() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return len(w.data)
}

// waitForTraceFile polls until a non-empty trace file exists at path (async
// captures complete after the request returns; Stop/Close drain, but they
// permanently stop the recorder, so in-flight tests poll instead).
func waitForTraceFile(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		info, err := os.Stat(path)
		if err == nil && info.Size() > 0 {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("trace file not written within 2s at %s", path)
}

// waitForTraceCount polls until dir holds at least want non-empty files.
func waitForTraceCount(t *testing.T, dir string, want int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir(dir)
		if err == nil {
			nonEmpty := 0

			for _, entry := range entries {
				if info, statErr := entry.Info(); statErr == nil && info.Size() > 0 {
					nonEmpty++
				}
			}

			if nonEmpty >= want {
				return
			}
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("dir %s never reached %d non-empty trace files within 2s", dir, want)
}

// countingWriter is a goroutine-safe io.Writer for assertions on async
// captures (bytes.Buffer races with a concurrent capture goroutine under
// -race).
type countingWriter struct {
	mu     sync.Mutex
	total  int
	writes int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.total += len(p)
	w.writes++

	return len(p), nil
}

func (w *countingWriter) totalBytes() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.total
}

// waitForBytes polls until the writer has received at least want bytes.
func waitForBytes(t *testing.T, w *countingWriter, want int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if w.totalBytes() >= want {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("writer never reached %d bytes within 2s (have %d)", want, w.totalBytes())
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

	waitForTraceFile(t, tracePath)
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

	waitForTraceFile(t, tracePath)
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

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rr.Code)
	}

	waitForTraceFile(t, tracePath)
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

	waitForTraceFile(t, tracePath)
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

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	waitForTraceFile(t, tracePath)
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

	waitForTraceFile(t, tracePath)

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("expected log output, got empty string")
	}

	if !bytes.Contains(buf.Bytes(), []byte("flightrecorder: trace capture initiated")) {
		t.Fatalf("expected log to contain capture-initiation message, got: %s", logOutput)
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

	waitForTraceFile(t, tracePath)

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
	dir := t.TempDir()

	rec, err := fr.New(
		fr.WithSnapshotDir(dir),
		fr.WithSnapshotPrefix("trace"),
	)
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

	cleanup := startRecorder(t, rec)
	defer cleanup()

	mw := appkitfr.Middleware(rec, fr.OnError())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	// First request captures to the first timestamped file (dir sinks are
	// not once-latched — every initiated capture writes a new file).
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr1, req1)

	waitForTraceCount(t, dir, 1)

	// Second request should also capture (autoReset is default true)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)
	handler.ServeHTTP(rr2, req2)

	waitForTraceCount(t, dir, 2)
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

// TestMiddleware_CaptureIsNonBlocking pins the async-capture contract: the
// request must never pay trace-write latency. A capture whose Write is stuck
// on a gate must not block the response; releasing the gate must land the
// bytes (Stop/Close drain in-flight async captures, but here the goroutine
// finishes on its own once the gate opens).
func TestMiddleware_CaptureIsNonBlocking(t *testing.T) {
	gw := newGatedWriter()

	rec, err := fr.New(fr.WithWriter(gw))
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

	recorderMu.Lock()

	if err := rec.Start(); err != nil {
		recorderMu.Unlock()
		t.Fatalf("rec.Start() error: %v", err)
	}

	t.Cleanup(func() {
		rec.Stop()
		_ = rec.Close()
		recorderMu.Unlock()
	})
	t.Cleanup(gw.unblock) // runs FIRST (LIFO): the drain must not wait on the gate

	mw := appkitfr.Middleware(rec, fr.OnError())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fail", nil)

	served := make(chan struct{})
	go func() {
		handler.ServeHTTP(rr, req)
		close(served)
	}()

	select {
	case <-served:
	case <-time.After(5 * time.Second):
		gw.unblock()
		t.Fatal("middleware blocked on the capture write; requests must not pay trace-write latency")
	}

	select {
	case <-gw.writeStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("capture never reached the writer")
	}

	if written := gw.bytesWritten(); written != 0 {
		t.Fatalf("blocked capture already wrote %d bytes before release", written)
	}

	gw.unblock()

	select {
	case <-gw.writeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("capture did not finish after release")
	}

	if gw.bytesWritten() == 0 {
		t.Fatal("expected trace bytes after release, got 0")
	}
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
	cw := &countingWriter{}

	rec, err := fr.New(fr.WithWriter(cw))
	if err != nil {
		t.Fatalf("fr.New() error: %v", err)
	}

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

	// The async capture must land BEFORE the manual snapshot: the handler
	// resets the once-latch, and if the latch still carries the middleware's
	// arm the manual write is silently skipped.
	waitForBytes(t, cw, 1)

	firstSize := cw.totalBytes()

	// Now manually snapshot — handler resets the latch, should capture again
	resp2 := httpPostJSON(t, ts.URL+"/debug/snapshot")
	defer func() { _ = resp2.Body.Close() }()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	if secondSize := cw.totalBytes(); secondSize <= firstSize {
		t.Fatalf("expected manual capture to grow writer: first=%d, second=%d", firstSize, secondSize)
	}
}
