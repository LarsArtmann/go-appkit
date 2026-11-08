package httpxtest_test

import (
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

func (buggyEmbeddingWrapper) Write(b []byte) (int, error) { return len(b), nil }

// noopLiarWrapper satisfies the type assertions but forwards nothing —
// the assertion must catch a stub that only passes capability checks.
type noopLiarWrapper struct {
	http.ResponseWriter
}

func (noopLiarWrapper) Flush()               {}
func (noopLiarWrapper) Hijack() (any, any, error) { return nil, nil, nil }

// The discrimination acceptance (battery B3): the assertion must FAIL on
// the buggy shapes. Run with a stub testing.T that records failures.
func TestAssertForwardsInterfaces_FailsOnBuggyShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		wrap func(http.ResponseWriter) http.ResponseWriter
		want []string{
			"hides http.Flusher",
		}
	}{
		{
			name: "plain embedding hides optional interfaces",
			wrap: func(base http.ResponseWriter) http.ResponseWriter {
				return buggyEmbeddingWrapper{base}
			},
			want: []string{"hides http.Flusher", "hides http.Hijacker"},
		},
		{
			name: "no-op stub satisfies assertions but forwards nothing",
			wrap: func(base http.ResponseWriter) http.ResponseWriter {
				return noopLiarWrapper{base}
			},
			want: []string{"does not forward Flush", "does not forward Hijack", "swallowed"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := &recordingTB{TB: t}
			AssertOn(rec, tc.wrap)

			joined := strings.Join(rec.failures, "\n")
			for _, want := range tc.want {
				if !strings.Contains(joined, want) {
					t.Errorf("assertion did not flag %q on the buggy shape; recorded failures:\n%s", want, joined)
				}
			}

			if len(rec.failures) == 0 {
				t.Error("assertion PASSED on a buggy wrapper — the discriminator is broken")
			}
		})
	}
}
