package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	errorfamily "github.com/larsartmann/go-error-family"
)

// errorResponse is the JSON error contract. Field names match the
// errorpages module's JSON rendering (templ-components/errorpage's
// errorResponse) so one taxonomy serves two surfaces: errorpages for the
// pretty HTML, httpx for the API. Fields outside this set (diagnostic
// context values, cause chains, timestamps) are deliberately omitted —
// they are developer diagnostics and the no-leak contract keeps them off
// the wire.
type errorResponse struct {
	Family  string `json:"family"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Why     string `json:"why,omitempty"`
	Fix     string `json:"fix,omitempty"`
}

// WriteError renders err as the classified JSON error response: the HTTP
// status comes from [errorfamily.HTTPStatus] (the same mapping errorpages
// uses), the message passes the no-leak sanitizer (see SanitizeErrorMessage),
// and code/why/fix derive from the classification when the error is one.
// A nil error writes nothing — callers should check errors before reaching
// for the response.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	family := errorfamily.Classify(err)

	resp := errorResponse{ //nolint:exhaustruct_v5 // Why and Fix are set conditionally below, from the classification
		Family:  family.String(),
		Code:    errorfamily.Code(err),
		Message: SanitizeErrorMessage(err),
	}

	if classified, ok := errors.AsType[errorfamily.Classified](err); ok {
		resp.Why = classified.ErrorFamily().DefaultWhy()
		resp.Fix = classified.ErrorFamily().DefaultFix()
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(errorfamily.HTTPStatus(err))

	// The status line is already committed; an encode failure here means
	// the client is gone, and there is no honest recovery.
	_ = json.NewEncoder(w).Encode(resp) //nolint:errchkjson // the contract types are plain strings; failure is connection death
}
