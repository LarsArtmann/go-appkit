package httpx

import (
	"encoding/json"
	"io"
	"net/http"

	errorfamily "github.com/larsartmann/go-error-family"
)

// maxBindBody bounds the JSON body a validation handler will read: 1 MiB
// covers every legitimate request payload for a typed bind; anything
// larger is a Rejection before the decoder runs. The security module's
// body-limit battery is the tunable, per-route alternative.
const maxBindBody = 1 << 20

// Bind codes — the grep-able contract of the validation battery.
const (
	CodeBindEmptyBody   = "bind.empty_body"
	CodeBindTooLarge    = "bind.body_too_large"
	CodeBindUnreadable  = "bind.unreadable_body"
	CodeBindInvalidBody = "bind.invalid_body"
)

// Validator is the optional post-bind hook for adapter types: when B
// implements it, ResultHandlerWithValidation runs Validate after the
// decode and renders a returned error through the classified contract.
// A plain error reads as a server-side family there; use a classified
// Rejection (errorfamily.Newf(errorfamily.Rejection, ...)) for
// field-level failures so clients get a 400.
type Validator interface {
	Validate() error
}

// bindError is a classified Rejection carrying a bind.* code; its text is
// client-facing (it describes the request, not the internals).
type bindError struct {
	code    string
	message string
}

func (e *bindError) Error() string { return e.message }

func (e *bindError) ErrorFamily() errorfamily.Family { return errorfamily.Rejection }

func (e *bindError) ErrorCode() string { return e.code }

// ResultHandlerWithValidation is ResultHandler with a JSON-body bind step:
// the request body decodes into B (a tagged adapter type — never an
// untagged domain struct; json/v2 matches field names case-sensitively, so
// a camelCase wire field silently misses a snake_case struct field and
// binds the zero value), then fn runs with the bound value. Decode
// failures render as Rejection/400 under the bind.* codes above; the
// happy path and error path are ResultHandler's.
//
// The body is expected to be JSON: a missing or empty body binds under
// bind.empty_body. Handlers that take no input use ResultHandler instead.
func ResultHandlerWithValidation[B any, T any](fn func(c *Context, in B) (T, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bound B

		bindErr := bindJSON(r, &bound)
		if bindErr != nil {
			WriteError(w, r, bindErr)

			return
		}

		if validator, ok := any(&bound).(Validator); ok {
			if err := validator.Validate(); err != nil {
				WriteError(w, r, err)

				return
			}
		}

		ctx := &Context{ResponseWriter: w, Request: r}

		result, err := fn(ctx, bound)
		if err != nil {
			WriteError(w, r, err)

			return
		}

		writeResult(ctx, result)
	})
}

// bindJSON decodes r's body into v under the bind.* Rejection contract.
func bindJSON(r *http.Request, v any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBindBody+1))
	if err != nil {
		return &bindError{code: CodeBindUnreadable, message: "request body could not be read"}
	}

	switch {
	case len(body) > maxBindBody:
		return &bindError{code: CodeBindTooLarge, message: "request body exceeds the 1 MiB bind limit"}
	case len(body) == 0:
		return &bindError{code: CodeBindEmptyBody, message: "request body is empty: expected JSON"}
	}

	if err := json.Unmarshal(body, v); err != nil {
		return &bindError{code: CodeBindInvalidBody, message: "request body is not valid JSON for this type"}
	}

	return nil
}
