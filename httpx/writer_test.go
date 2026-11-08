package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/httpx/httpxtest"
)

func TestWrapWriter_ForwardsOptionalInterfaces(t *testing.T) {
	t.Parallel()

	httpxtest.AssertForwardsInterfaces(t, func(base http.ResponseWriter) http.ResponseWriter {
		return WrapWriter(base)
	})
}

func TestWrapWriter_FlushReachesFlushableBase(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	w := WrapWriter(recorder)

	flusher, ok := w.(http.Flusher)
	if !ok {
		t.Fatal("WrapWriter does not satisfy http.Flusher over a flushable base")
	}

	flusher.Flush() // must not panic; the recorder accepts Flush

	if !recorder.Flushed {
		t.Fatal("Flush did not reach the flushable base")
	}
}

func TestWrapWriter_HijackReturnsErrorOnNonHijackableBase(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	w := WrapWriter(recorder)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		t.Fatal("WrapWriter does not satisfy http.Hijacker")
	}

	conn, _, err := hijacker.Hijack()
	if err == nil {
		t.Fatal("Hijack over a non-hijackable base must fail loudly, not pretend the connection was taken")
	}

	if conn != nil {
		t.Fatal("Hijack returned a connection alongside an error")
	}

	if !errors.Is(err, http.ErrHijackerNotSupported) {
		t.Fatalf("Hijack error = %v, want it to carry http.ErrHijackerNotSupported", err)
	}
}

func TestWrapWriter_ResponseControllerReachable(t *testing.T) {
	t.Parallel()

	// The recorder supports no deadlines; what is pinned here is that the
	// controller's unwrap chain WORKS through WrapWriter: the error is the
	// recorder's own, not "the ResponseWriter does not implement Unwrap".
	recorder := httptest.NewRecorder()
	w := WrapWriter(WrapWriter(recorder)) // two layers: every level must unwrap

	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Now())

	if _, ok := w.Unwrap().(*Writer); !ok {
		t.Fatal("Unwrap did not expose the underlying writer")
	}
}
