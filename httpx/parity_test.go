package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-appkit/errorpages"
	"github.com/larsartmann/go-appkit/httpx"
	errorfamily "github.com/larsartmann/go-error-family"
)

// parityErrors is the classification set both renderings must agree on:
// every family the taxonomy defines, constructed the way the fleet does.
func parityErrors(t *testing.T) map[string]error {
	t.Helper()

	return map[string]error{
		"Rejection":      errorfamily.Newf(errorfamily.Rejection, "p.rejected", "the request was rejected"),
		"Conflict":       errorfamily.Newf(errorfamily.Conflict, "p.conflict", "the resource is in conflict"),
		"Transient":      errorfamily.Newf(errorfamily.Transient, "p.transient", "the dependency is unavailable"),
		"Corruption":     errorfamily.Newf(errorfamily.Corruption, "p.corruption", "the data is corrupted"),
		"Infrastructure": errorfamily.WrapInfrastructure(errors.New("disk gone"), "p.infra", "infrastructure failure"),
		"Orchestration":  errorfamily.Newf(errorfamily.Orchestration, "p.orchestration", "coordination failed"),
		"WrappedRejection": errorfamily.WrapRejection(
			errorfamily.Newf(errorfamily.Rejection, "p.inner", "inner cause"),
			"p.outer", "outer cause",
		),
	}
}

// TestClassificationParity_WithEPrrorpages pins the one-taxonomy contract:
// for every family, the httpx API surface and the errorpages rendering
// surface produce the SAME HTTP status, and both speak the same JSON
// field names. A drift here means one surface silently disagrees with
// the other — the failure mode the battery exists to prevent.
func TestClassificationParityWithErrorpages(t *testing.T) {
	t.Parallel()

	for name, err := range parityErrors(t) {
		wantStatus := errorfamily.HTTPStatus(err)

		httpxRec := httptest.NewRecorder()
		httpxHandler := httpx.ResultHandler(func(_ *httpx.Context) (*thing, error) {
			return nil, err
		})
		httpxHandler.ServeHTTP(httpxRec, httptest.NewRequest(http.MethodGet, "/api/x", nil))

		pagesRec := httptest.NewRecorder()
		pagesCfg := errorpages.Config{
			JSONWhen: func(_ *http.Request) bool { return true },
		}
		errorpages.Handler(err, pagesCfg).ServeHTTP(pagesRec, httptest.NewRequest(http.MethodGet, "/x", nil))

		if httpxRec.Code != wantStatus {
			t.Errorf("%s: httpx status = %d, want %d (errorfamily.HTTPStatus)", name, httpxRec.Code, wantStatus)
		}

		if pagesRec.Code != wantStatus {
			t.Errorf("%s: errorpages status = %d, want %d — the two surfaces disagree", name, pagesRec.Code, wantStatus)
		}

		var httpxBody struct {
			Family  string `json:"family"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(httpxRec.Body.Bytes(), &httpxBody); err != nil {
			t.Errorf("%s: httpx body %q is not the contract shape: %v", name, httpxRec.Body.String(), err)
		}

		var pagesBody struct {
			Family  string `json:"family"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(pagesRec.Body.Bytes(), &pagesBody); err != nil {
			t.Errorf("%s: errorpages body %q is not the contract shape: %v", name, pagesRec.Body.String(), err)
		}

		if httpxBody.Family != pagesBody.Family {
			t.Errorf("%s: family %q vs errorpages %q — surfaces disagree", name, httpxBody.Family, pagesBody.Family)
		}
	}
}
