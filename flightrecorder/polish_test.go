package flightrecorder_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/flightrecorder"
	fr "github.com/larsartmann/go-flightrecorder"
)

// TestSnapshotHandler_NotEnabledIsExplicit pins the debug-surface contract:
// a DISABLED recorder must answer with an explicit 503 "recorder not
// enabled", never a silent 200 "snapshot captured" (data loss disguised as
// success).
func TestSnapshotHandler_NotEnabledIsExplicit(t *testing.T) {
	t.Parallel()

	rec, err := fr.New(fr.WithWriter(&strings.Builder{}))
	if err != nil {
		t.Fatalf("fr.New: %v", err)
	}

	handler := flightrecorder.SnapshotHandler(rec)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/debug/flightrecorder/snapshot", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("disabled recorder: status = %d, want 503", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "recorder not enabled") {
		t.Errorf("disabled recorder body must say so, got: %s", recorder.Body.String())
	}
}

// TestOpsRecorderPreset_OptionCount sanity: the preset must carry the full
// documented option set (dir, prefix, cap, bytes, compression, min-age).
func TestOpsRecorderPreset_OptionCount(t *testing.T) {
	t.Parallel()

	const (
		wantOptions = 6
		presetSnaps = 5
		presetBytes = 64 << 20
	)

	opts := flightrecorder.OpsRecorderPreset(t.TempDir(), presetSnaps, presetBytes)
	if len(opts) != wantOptions {
		t.Errorf(
			"preset carries %d options, want 6 (dir, prefix, max snapshots, max bytes, compression, min age)",
			len(opts),
		)
	}
}

// mustRecorder builds a disabled recorder for construction-time assertions.
func mustRecorder(t *testing.T) *fr.Recorder {
	t.Helper()

	rec, err := fr.New(fr.WithWriter(&strings.Builder{}))
	if err != nil {
		t.Fatalf("fr.New: %v", err)
	}

	return rec
}
