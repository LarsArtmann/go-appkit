package httpx

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// Writer is the pass-through ResponseWriter wrapper that keeps every
// optional interface reachable: http.Flusher, http.Hijacker, and Unwrap
// for http.ResponseController (per-request deadlines, streaming controls).
//
// A wrapper that hides these interfaces is the SSE-stall class of bug this
// battery exists to prevent: it compiles clean, passes httptest.Recorder
// tests, and then stalls production streams — frames sit in net/http's
// internal buffer until the client times out, because nothing down the
// chain can reach the real connection's Flush. Wrap your wrappers in
// [WrapWriter] (or forward the interfaces yourself) and pin the contract
// with httpxtest.AssertForwardsInterfaces.
type Writer struct {
	http.ResponseWriter
}

// WrapWriter wraps w so optional interfaces stay reachable through it.
// The returned Writer forwards Flusher and Hijacker whenever the
// underlying writer supports them, and Unwrap always, so
// http.ResponseController and direct type assertions find the real
// connection no matter how many wrappers sit above it.
func WrapWriter(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w}
}

// Flush forwards to the underlying writer when it implements
// http.Flusher. When it does not, this is a no-op — flushing was never
// possible through this chain, and claiming the interface keeps callers'
// single code path honest instead of letting the assertion decide
// semantics.
func (w *Writer) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack forwards to the underlying writer when it implements
// http.Hijacker; otherwise it returns an explicit error rather than
// pretending the connection was taken over.
func (w *Writer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("httpx: the underlying ResponseWriter does not support hijacking")
	}

	return hijacker.Hijack() //nolint:wrapcheck // pure delegation to the real hijacker
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (w *Writer) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
