// The flightrecorder-middleware-through-a-real-Service composition. Pins the
// capture pipeline end-to-end against PUBLISHED tags: a trigger-scoped
// flightrecorder.Middleware in front of a route on an appkit Service, with
// the SnapshotHandler mounted as the operator endpoint.
//
//   - GET /work (trigger matches) → 200, and the capture lands ASYNCHRONOUSLY
//     as a new timestamped snapshot-*.trace file in the snapshot dir — the
//     request never pays the trace-write latency.
//   - A second GET /work writes a SECOND file (snapshot-dir sink = every
//     initiated capture retains its own file; no once-latch dedupe).
//   - GET /other (trigger returns false) writes NOTHING — trigger filtering.
//   - POST /debug/flightrecorder/snapshot?download=1 → the endpoint streams
//     the current trace buffer as application/octet-stream regardless of the
//     configured sink (SnapshotToWriter bypass), so operators can fetch the
//     artifact over HTTP.
//
// The trigger matches on Type ("GET /work") rather than latency on purpose:
// latency thresholds are flaky under shared-box load, Type matching is
// deterministic.

package integration_test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/flightrecorder"
	"github.com/larsartmann/go-appkit/testkit"
	fr "github.com/larsartmann/go-flightrecorder"
)

func snapshotCount(t *testing.T, dir string) int {
	t.Helper()

	return len(nonEmptySnapshots(t, dir))
}

// nonEmptySnapshots lists snapshot files that carry trace bytes. The library
// creates the file before flushing the trace, so existence alone races the
// write — size is the readiness signal.
func nonEmptySnapshots(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := filepath.Glob(filepath.Join(dir, "snapshot-*.trace*"))
	if err != nil {
		t.Fatalf("glob snapshots: %v", err)
	}

	ready := entries[:0]
	for _, path := range entries {
		info, statErr := os.Stat(path)
		if statErr == nil && info.Size() > 0 {
			ready = append(ready, path)
		}
	}

	return ready
}

func waitForSnapshots(t *testing.T, dir string, want int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if snapshotCount(t, dir) >= want {
			return
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("snapshot dir never reached %d files (has %d): async capture did not land", want, snapshotCount(t, dir))
}

func TestFlightrecorderMiddlewareThroughService(t *testing.T) {
	// Deliberately NOT t.Parallel: fr.Recorder wraps the process-global
	// runtime trace (one active recorder per process). Sequential execution
	// finishes this test's Start-to-Close window before any parallel test
	// body (e.g. health_stack's recorder) resumes, so the two recorders can
	// never contend for the runtime trace.

	snapDir := t.TempDir()

	rec, err := fr.New(fr.WithSnapshotDir(snapDir), fr.WithMinAge(time.Millisecond), fr.WithMaxBytes(1<<20))
	if err != nil {
		t.Fatalf("new recorder: %v", err)
	}

	if startErr := rec.Start(); startErr != nil {
		t.Fatalf("recorder start: %v", startErr)
	}

	// Trace-buffer warm-up: a capture fired immediately after Start flushes
	// a near-empty buffer (the first subtest measured 0 bytes at MinAge=1ms).
	// health_stack sleeps 100ms for the same reason; 200ms guarantees the
	// first capture carries real trace data even under load.
	time.Sleep(200 * time.Millisecond)

	t.Cleanup(func() {
		if closeErr := rec.Close(); closeErr != nil {
			t.Errorf("recorder close: %v", closeErr)
		}
	})

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:       freeAddr(t),
		DrainDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	trigger := func(tc fr.TriggerContext) bool { return tc.Type == "GET /work" }
	svc.Mux.Handle(
		"GET /work",
		flightrecorder.Middleware(rec, trigger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("work done"))
		})),
	)

	// A trigger that never fires: proves the middleware consults the trigger
	// rather than capturing every request.
	never := func(fr.TriggerContext) bool { return false }
	svc.Mux.Handle(
		"GET /other",
		flightrecorder.Middleware(rec, never)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("other"))
		})),
	)

	svc.Mux.Handle("POST /debug/flightrecorder/snapshot", flightrecorder.SnapshotHandler(rec))

	ts := testkit.Serve(t, svc)
	base := ts.FullChainURL

	t.Run("trigger-matched request lands an async snapshot", func(t *testing.T) {
		frE2EFirstCapture(t, base, snapDir)
	})

	t.Run("second matched request writes a second retained file", func(t *testing.T) {
		frE2ESecondCapture(t, base, snapDir)
	})

	t.Run("non-matching trigger captures nothing", func(t *testing.T) {
		frE2ETriggerFiltering(t, base, snapDir)
	})

	t.Run("snapshot endpoint streams the buffer over HTTP", func(t *testing.T) {
		frE2ESnapshotDownload(t, base)
	})
}

func frE2EFirstCapture(t *testing.T, base, snapDir string) {
	t.Helper()

	resp := e2eGet(t, base+"/work")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /work: status = %d, want 200", resp.StatusCode)
	}

	waitNonEmpty := func() []string {
		deadline := time.Now().Add(5 * time.Second)

		for time.Now().Before(deadline) {
			if got := nonEmptySnapshots(t, snapDir); len(got) >= 1 {
				return got
			}

			time.Sleep(25 * time.Millisecond)
		}

		t.Fatalf(
			"no non-empty snapshot within 5s (has %d files): async capture did not flush",
			snapshotCount(t, snapDir),
		)

		return nil
	}

	matches := waitNonEmpty()

	info, statErr := os.Stat(matches[0])
	if statErr != nil {
		t.Fatalf("stat snapshot: %v", statErr)
	}

	if info.Size() == 0 {
		t.Error("snapshot file is empty — the capture wrote no trace bytes")
	}
}

func frE2ESecondCapture(t *testing.T, base, snapDir string) {
	t.Helper()

	resp := e2eGet(t, base+"/work")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /work: status = %d, want 200", resp.StatusCode)
	}

	waitForSnapshots(t, snapDir, 2)
}

func frE2ETriggerFiltering(t *testing.T, base, snapDir string) {
	t.Helper()

	resp := e2eGet(t, base+"/other")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /other: status = %d, want 200", resp.StatusCode)
	}

	time.Sleep(300 * time.Millisecond)

	if got := snapshotCount(t, snapDir); got != 2 {
		t.Errorf("snapshot count after non-matching request = %d, want 2 (unchanged)", got)
	}
}

func frE2ESnapshotDownload(t *testing.T, base string) {
	t.Helper()

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		base+"/debug/flightrecorder/snapshot?download=1",
		nil,
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("POST snapshot download: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("snapshot download: status = %d, want 200", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/octet-stream") {
		t.Errorf("Content-Type = %q, want application/octet-stream", ct)
	}

	body, _ := io.ReadAll(resp.Body)

	if len(body) == 0 {
		t.Error("download body is empty — SnapshotToWriter streamed no trace bytes")
	}
}
