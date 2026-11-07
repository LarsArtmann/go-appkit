package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/httpx"
	errorfamily "github.com/larsartmann/go-error-family"
)

// bindThing is the canonical tagged adapter type: json tags on every
// field, Validate as the post-bind hook. NEVER bind untagged domain
// structs — json/v2 matches field names case-sensitively, so a
// camelCase wire field silently misses an untagged (or mismatched-tag)
// struct field and binds the zero value (the CV {"appId":...} 400 bug).
type bindThing struct {
	AppID string `json:"app_id"`
	Count int    `json:"count"`
}

func (b *bindThing) Validate() error {
	if b.AppID == "" {
		return errorfamily.Newf(errorfamily.Rejection, "bind.app_id_required", "app_id is required")
	}

	return nil
}

func TestResultHandlerWithValidation_HappyPath(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandlerWithValidation(
		func(_ *httpx.Context, in bindThing) (*bindThing, error) {
			return &in, nil
		})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/things",
		strings.NewReader(`{"app_id":"a1","count":3}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s, want 200", rec.Code, rec.Body.String())
	}

	if got := rec.Body.String(); !strings.Contains(got, `"app_id":"a1"`) {
		t.Errorf("body = %s, want the bound value echoed", got)
	}
}

// json/v2 matches case-sensitively: a camelCase wire field must NOT land
// in a snake_case-tagged struct (the zero value documents the trap this
// battery's docs warn about).
func TestResultHandlerWithValidation_CaseSensitiveBind(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandlerWithValidation(
		func(_ *httpx.Context, in bindThing) (*bindThing, error) {
			return &in, nil
		})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/things",
		strings.NewReader(`{"appId":"a1","COUNT":3}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (app_id required: camelCase wire field must miss the tagged field)", rec.Code)
	}
}

func TestResultHandlerWithValidation_BindErrorContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		status  int
		message string
	}{
		{"empty body", "", http.StatusBadRequest, httpx.CodeBindEmptyBody},
		{"invalid json", "{not json", http.StatusBadRequest, httpx.CodeBindInvalidBody},
	}

	for _, tt := range tests {
		handler := httpx.ResultHandlerWithValidation(func(_ *httpx.Context, in bindThing) (*bindThing, error) {
			return &in, nil
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/things", strings.NewReader(tt.body))
		handler.ServeHTTP(rec, req)

		if rec.Code != tt.status {
			t.Errorf("%s: status = %d, want %d", tt.name, rec.Code, tt.status)
		}

		if !strings.Contains(rec.Body.String(), tt.message) {
			t.Errorf("%s: body = %s, want the bind code %q", tt.name, rec.Body.String(), tt.message)
		}
	}
}

// An oversized body is rejected BEFORE decoding; the leaked-detail matrix
// is in sanitize_test.go.
func TestResultHandlerWithValidation_BodyTooLarge(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandlerWithValidation(func(_ *httpx.Context, in bindThing) (*bindThing, error) {
		return &in, nil
	})

	big := `{"app_id":"` + strings.Repeat("x", 2<<20) + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/things", strings.NewReader(big))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), httpx.CodeBindTooLarge) {
		t.Errorf("body = %s, want the %q code", rec.Body.String(), httpx.CodeBindTooLarge)
	}
}

// The validator hook renders through the classified contract: the
// adapter's Rejection surfaces as a 400 with its own code.
func TestResultHandlerWithValidation_ValidatorRejection(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandlerWithValidation(func(_ *httpx.Context, in bindThing) (*bindThing, error) {
		return &in, nil
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/things", strings.NewReader(`{"count":3}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (validator Rejection)", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "app_id is required") {
		t.Errorf("body = %s, want the validator message", rec.Body.String())
	}
}
