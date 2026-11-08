package httpxtest_test

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/httpx/httpxtest"
)

// buggyEmbeddingWrapper is the notFoundInterceptor class: plain embedding
// of http.ResponseWriter, no Unwrap, no explicit Flusher/Hijacker. It
// compiles clean and passes Recorder-based tests — and hides every
// optional interface, which is the production SSE-stall bug.
type buggyEmbeddingWrapper struct {
	http.ResponseWriter
}

// noopLiarWrapper satisfies the type assertions but forwards nothing —
// the findings must catch a stub that only passes capability checks.
type noopLiarWrapper struct {
	http.ResponseWriter
}

func (noopLiarWrapper) Flush() {}

func (noopLiarWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) { return nil, nil, nil }

// opaqueControllerWrapper forwards nothing and answers the controller
// with its own capability — the controller talks to a stranger, never the
// base.
type opaqueControllerWrapper struct {
	http.ResponseWriter
}

func (opaqueControllerWrapper) Flush() {}

func (opaqueControllerWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (opaqueControllerWrapper) EnableFullDuplex() error { return nil }

// The discrimination acceptance (battery B3): the findings must be EMPTY
// for an honest wrapper and NON-EMPTY, with the right reasons, for the
// buggy shapes.
func TestForwardingFindings_Discriminates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		wrap func(http.ResponseWriter) http.ResponseWriter
		want []string
	}{
		{
			name: "plain embedding hides optional interfaces",
			wrap: func(base http.ResponseWriter) http.ResponseWriter {
				return buggyEmbeddingWrapper{base}
			},
			want: []string{
				"hides http.Flusher",
				"hides http.Hijacker",
			},
		},
		{
			name: "no-op stub satisfies assertions but forwards nothing",
			wrap: func(base http.ResponseWriter) http.ResponseWriter {
				return noopLiarWrapper{base}
			},
			want: []string{
				"does not forward Flush",
				"does not forward Hijack",
			},
		},
		{
			name: "controller-opaque wrapper answers with its own capability",
			wrap: func(base http.ResponseWriter) http.ResponseWriter {
				return opaqueControllerWrapper{base}
			},
			want: []string{
				"opaque to http.ResponseController",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			findings := httpxtest.ForwardingFindings(tc.wrap)
			joined := strings.Join(findings, "\n")

			for _, want := range tc.want {
				if !strings.Contains(joined, want) {
					t.Errorf("findings did not flag %q; recorded:\n%s", want, joined)
				}
			}

			if len(findings) == 0 {
				t.Error("findings were EMPTY on a buggy wrapper — the discriminator is broken")
			}
		})
	}
}

func TestForwardingFindings_HonestWrapperIsClean(t *testing.T) {
	t.Parallel()

	// An honest explicit-forwarding wrapper (the shape WrapWriter ships):
	// forwards everything, unwraps for the controller.
	findings := httpxtest.ForwardingFindings(func(base http.ResponseWriter) http.ResponseWriter {
		return honestWrapper{base}
	})

	for _, f := range findings {
		t.Errorf("honest wrapper flagged: %s", f)
	}
}

type honestWrapper struct {
	http.ResponseWriter
}

func (w honestWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w honestWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}

	return h.Hijack()
}

func (w honestWrapper) Unwrap() http.ResponseWriter { return w.ResponseWriter }
