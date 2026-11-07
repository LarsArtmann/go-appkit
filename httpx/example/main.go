// Example httpx battery: a JSON API surface where every handler returns
// (T, error) and the battery owns serialization, classification-aware
// errors, and the no-leak contract.
package main

import (
	"net/http"

	"github.com/larsartmann/go-appkit/httpx"
	errorfamily "github.com/larsartmann/go-error-family"
)

type greeting struct {
	Message string `json:"message"`
}

type greetRequest struct {
	Name string `json:"name"`
}

func (r *greetRequest) Validate() error {
	if r.Name == "" {
		return errorfamily.Newf(errorfamily.Rejection, "greet.name_required", "name is required")
	}

	return nil
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("GET /api/greeting", httpx.ResultHandler(func(_ *httpx.Context) (*greeting, error) {
		return &greeting{Message: "hello"}, nil
	}))

	mux.Handle("POST /api/greeting", httpx.ResultHandlerWithValidation(
		func(_ *httpx.Context, in greetRequest) (*greeting, error) {
			return &greeting{Message: "hello, " + in.Name}, nil
		},
	))

	server := &http.Server{ //nolint:exhaustruct_v5,gosec // example: only the fields that matter to a demo
		Addr:    "localhost:8080",
		Handler: mux,
	}

	_ = server.ListenAndServe()
}
