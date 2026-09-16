package security_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

// TestBodyLimit_TypedErrorOnOversize pins the contract: reads beyond the
// limit fail with a *http.MaxBytesError (typed, mappable to 413) — never
// silent truncation.
func TestBodyLimit_TypedErrorOnOversize(t *testing.T) {
	t.Parallel()

	const limit = 16

	handler := security.BodyLimit(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("oversize read: no error, expected MaxBytesError")

			return
		}

		if _, ok := errors.AsType[*http.MaxBytesError](err); !ok {
			t.Errorf("oversize read: err = %T, want *http.MaxBytesError (typed, not silent truncation)", err)
		}

		w.WriteHeader(http.StatusOK)
	}))

	body := strings.Repeat("x", limit*4)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", strings.NewReader(body))
	req.ContentLength = int64(len(body))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestBodyLimit_UnderLimitPasses(t *testing.T) {
	t.Parallel()

	handler := security.BodyLimit(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("under-limit read: %v", err)
		}

		if len(n) != 10 {
			t.Errorf("read %d bytes, want 10 (no truncation under the limit)", len(n))
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", strings.NewReader("0123456789"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
