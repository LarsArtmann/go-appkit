package httpx

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// resultWriter wraps the ResponseWriter to record whether the wrapped
// handler function wrote its own response. It forwards Unwrap so
// http.ResponseController (and everything reachable through it: Flusher,
// Hijacker, per-request deadlines) stays reachable from inside handlers —
// a wrapper that hides those interfaces is the SSE-stall class of bug the
// W3 battery exists to prevent.
type resultWriter struct {
	http.ResponseWriter

	didWrite bool
}

func (w *resultWriter) Write(b []byte) (int, error) {
	w.didWrite = true

	return w.ResponseWriter.Write(b)
}

func (w *resultWriter) WriteHeader(status int) {
	w.didWrite = true

	w.ResponseWriter.WriteHeader(status)
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (w *resultWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// ResultHandler converts a function returning (T, error) into an
// http.Handler. A nil error serializes T as JSON with status 200 (204 when
// T is a nil pointer — there is no payload to speak of); a non-nil error
// renders through the classified error contract (see WriteError) with the
// status from go-error-family's taxonomy.
//
// A handler function that writes to c.ResponseWriter directly (file
// download, SSE stream, custom encoding) takes ownership: the battery
// skips its own serialization once any write has gone through.
func ResultHandler[T any](fn func(c *Context) (T, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &resultWriter{ResponseWriter: w}
		c := &Context{ResponseWriter: rw, Request: r}

		result, err := fn(c)
		if err != nil {
			WriteError(w, r, err)

			return
		}

		if rw.didWrite {
			return
		}

		writeResult(c, result)
	})
}

// writeResult serializes the happy path: 204 for nil-pointer payloads,
// JSON 200 otherwise. Called only when the handler function has not
// written its own response (the caller checks the recording writer).
func writeResult(c *Context, result any) {
	if isNilResult(result) {
		c.ResponseWriter.WriteHeader(http.StatusNoContent)

		return
	}

	c.ResponseWriter.Header().Set("Content-Type", "application/json")

	// Serialization of an in-memory value cannot fail on valid input; a
	// failure means the client is gone or the connection broke mid-write.
	// The status is already committed, so the only honest move is to stop.
	_ = json.NewEncoder(c.ResponseWriter).Encode(result)
}

// isNilResult reports whether the happy path carries no payload: a nil
// pointer of any type (including typed nils like (*Thing)(nil)). Values of
// non-pointer kinds always serialize — an empty struct is information.
func isNilResult[T any](result T) bool {
	rv := reflect.ValueOf(&result).Elem()

	return rv.Kind() == reflect.Pointer && rv.IsNil()
}
