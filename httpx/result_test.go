package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/httpx"
	"github.com/larsartmann/go-error-family"
)

type thing struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestResultHandler_HappyPath(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandler(func(_ *httpx.Context) (*thing, error) {
		return &thing{ID: "42", Name: "answer"}, nil
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/things/42", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var got thing
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}

	if got.ID != "42" || got.Name != "answer" {
		t.Errorf("body = %+v, want the thing", got)
	}
}

// A nil-pointer result is 204: there is no payload, and a bare "null"
// body invites client bugs.
func TestResultHandler_NilResultIs204(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandler(func(_ *httpx.Context) (*thing, error) {
		return nil, nil
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/things/none", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}
}

// A handler that writes its own response (SSE, downloads) owns it: the
// battery must not append a JSON body, and Flusher-style reachability
// through http.ResponseController must survive the wrapper.
func TestResultHandler_DirectWriteTakesOwnership(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandler(func(c *httpx.Context) (string, error) {
		flusher := http.NewResponseController(c.ResponseWriter)

		if err := flusher.Flush(); err != nil {
			return "", err
		}

		_, _ = c.ResponseWriter.Write([]byte("streamed"))

		return "", nil
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stream", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if got := rec.Body.String(); got != "streamed" {
		t.Errorf("body = %q, want the streamed bytes only (no JSON append)", got)
	}
}

func TestResultHandler_ClassifiedErrorMapsStatus(t *testing.T) {
	t.Parallel()

	rejected := errorfamily.Newf(errorfamily.Rejection, "thing.not_found", "thing not found")

	handler := httpx.ResultHandler(func(_ *httpx.Context) (*thing, error) {
		return nil, rejected
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/things/42", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for the classified Rejection", rec.Code)
	}

	var body struct {
		Family  string `json:"family"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body.String(), err)
	}

	if body.Family != "rejection" {
		t.Errorf("family = %q, want rejection", body.Family)
	}

	if body.Code == "" {
		t.Error("code empty: classified errors must carry their code on the wire")
	}
}
