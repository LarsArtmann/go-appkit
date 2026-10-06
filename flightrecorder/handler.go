package flightrecorder

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strconv"

	fr "github.com/larsartmann/go-flightrecorder"
)

// defaultSnapshotFilename is the download-mode attachment name. The
// downloaded bytes are the raw (writer-path) trace; append ".gz" via
// [WithSnapshotFilename] when the recorder was built with compression.
const defaultSnapshotFilename = "trace.trace"

// HandlerOption configures [SnapshotHandler].
type HandlerOption func(*handlerConfig)

type handlerConfig struct {
	filename string
}

// WithSnapshotFilename sets the attachment filename served by download mode
// (`?download=1`). Default: "trace.trace".
//
// CAVEAT: the downloaded bytes respect the recorder's compression setting —
// a recorder built with [fr.WithCompression] (e.g. via [OpsRecorderPreset])
// serves gzip. Name the artifact accordingly:
//
//	flightrecorder.WithSnapshotFilename("trace.trace.gz")
func WithSnapshotFilename(name string) HandlerOption {
	return func(c *handlerConfig) { c.filename = name }
}

// SnapshotHandler returns an [http.Handler] that triggers a manual flight
// recorder snapshot on demand. The snapshot is written to the recorder's
// configured destination (set via [fr.WithFile], [fr.WithWriter], or
// [fr.WithSnapshotDir] when creating the recorder).
//
// The handler resets the once-latch before snapshotting so the manual capture
// works even if an automatic middleware capture already consumed it.
//
// Download mode — append `?download=1` to stream the trace itself as an
// `application/octet-stream` attachment (Content-Length exact, filename from
// [WithSnapshotFilename]; see its compression caveat). This works for ANY
// sink: it bypasses the configured destination and writes the current trace
// buffer straight into the response via [fr.Recorder.SnapshotToWriter], so
// operators can fetch the artifact over HTTP without shell access. Errors
// keep the JSON contract (500 "snapshot failed").
//
// Register on any mux or router:
//
//	mux.Handle("POST /debug/flightrecorder/snapshot",
//	    flightrecorder.SnapshotHandler(rec))
//
// Or use [Mount] for stdlib mux convenience.
func SnapshotHandler(rec *fr.Recorder, opts ...HandlerOption) http.Handler {
	cfg := handlerConfig{filename: defaultSnapshotFilename}

	for _, opt := range opts {
		opt(&cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Explicit not-enabled status: a silent 200 "snapshot captured" on a
		// disabled recorder is data loss on a debug endpoint — the operator
		// believes they have a trace they do not have.
		if !rec.Enabled() {
			writeSnapshotResponse(w, http.StatusServiceUnavailable,
				"recorder not enabled",
				"start the recorder at boot (or via its own lifecycle) before requesting snapshots")

			return
		}

		if r.URL.Query().Get("download") != "" {
			snapshotDownload(rec, cfg.filename, w, r)

			return
		}

		// Re-arm the latch so a manual snapshot works even if the
		// middleware already captured an automatic one.
		rec.Reset()

		err := rec.Snapshot(r.Context())
		if err != nil {
			writeSnapshotResponse(w, http.StatusInternalServerError, "snapshot failed", err.Error())

			return
		}

		writeSnapshotResponse(w, http.StatusOK, "snapshot captured", "")
	})
}

// snapshotDownload buffers the trace ([fr.Recorder.SnapshotToWriter]) before
// touching the ResponseWriter: a mid-write failure still yields the JSON
// error contract instead of a truncated 200, and Content-Length is exact.
func snapshotDownload(rec *fr.Recorder, filename string, w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := rec.SnapshotToWriter(r.Context(), &buf)
	if err != nil {
		writeSnapshotResponse(w, http.StatusInternalServerError, "snapshot failed", err.Error())

		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(buf.Bytes())
}

// Mount registers a snapshot endpoint on the mux using [SnapshotHandler].
// This is the convenience entry point for stdlib mux consumers; options pass
// through (e.g. [WithSnapshotFilename]).
//
//	flightrecorder.Mount(svc.Mux, "POST /debug/flightrecorder/snapshot", rec)
//
// Fetch the artifact with:
//
//	curl -X POST 'http://localhost:8080/debug/flightrecorder/snapshot?download=1' -OJ
func Mount(mux *http.ServeMux, pattern string, rec *fr.Recorder, opts ...HandlerOption) {
	mux.Handle(pattern, SnapshotHandler(rec, opts...))
}

type snapshotResponse struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func writeSnapshotResponse(w http.ResponseWriter, code int, status, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_ = json.MarshalWrite(w, snapshotResponse{Status: status, Detail: detail})
}
