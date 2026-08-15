// Package errorpages gives go-appkit services pretty, go-error-family-aware
// error pages (HTML) and error contracts (JSON) backed by
// github.com/larsartmann/templ-components/errorpage.
//
// Every error rendered through this package is classified with the same
// 5-family taxonomy appkit already uses for HTTPStatus(): Rejection → 400,
// Conflict → 409, Transient → 503, Corruption → 500, Infrastructure → 503.
// HTML responses render the templ-components error page (title, message,
// why, fix, cause chain); JSON responses emit the machine-readable contract
// {family, code, message, title, why, fix, context} that htmx
// GlobalErrorHandling turns into toasts client-side.
//
// # Quick start
//
// Mount the pretty 404 catch-all on your appkit service's mux:
//
//	svc, _ := appkit.NewService(appkit.DefaultServiceConfig())
//	errorpages.Mount(svc.Mux, errorpages.Config{})
//
// Render classified errors from your own handlers:
//
//	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
//	    if err := findUser(r); err != nil {
//	        errorpages.Write(w, r, err, errorpages.Config{})
//	        return
//	    }
//	    // ...
//	})
//
// For pretty 405 responses too, serve the wrapped mux instead of the mux
// (works with any stdlib mux, not just appkit's):
//
//	http.ListenAndServe(":8080", errorpages.Wrap(mux, errorpages.Config{}))
//
// # Content negotiation
//
// Each request picks HTML or JSON: by default JSON is served when the Accept
// header names application/json, HTML otherwise (browsers). Override with
// Config.JSONWhen for custom rules (API-key based, path prefix, anything).
//
// # Build note
//
// Requires GOEXPERIMENT=jsonv2: templ-components/errorpage uses
// encoding/json/v2 for the JSON contract.
package errorpages
