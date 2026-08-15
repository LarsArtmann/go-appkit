package flightrecorder

import (
	"encoding/json/v2"
	"net/http"

	fr "github.com/larsartmann/go-flightrecorder"
)

// SnapshotHandler returns an [http.Handler] that triggers a manual flight
// recorder snapshot on demand. The snapshot is written to the recorder's
// configured destination (set via [fr.WithFile] or [fr.WithWriter] when
// creating the recorder).
//
// The handler resets the once-latch before snapshotting so the manual capture
// works even if an automatic middleware capture already consumed it.
//
// Register on any mux or router:
//
//	mux.Handle("POST /debug/flightrecorder/snapshot",
//	    flightrecorder.SnapshotHandler(rec))
//
// Or use [Mount] for stdlib mux convenience.
func SnapshotHandler(rec *fr.Recorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// Mount registers a snapshot endpoint on the mux using [SnapshotHandler].
// This is the convenience entry point for stdlib mux consumers.
//
//	flightrecorder.Mount(svc.Mux, "POST /debug/flightrecorder/snapshot", rec)
func Mount(mux *http.ServeMux, pattern string, rec *fr.Recorder) {
	mux.Handle(pattern, SnapshotHandler(rec))
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
