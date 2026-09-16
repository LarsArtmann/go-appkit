package security

import "net/http"

// BodyLimit wraps the request body with http.MaxBytesReader so reads beyond
// maxBytes fail with a *http.MaxBytesError — a TYPED error, never silent
// truncation: handlers that decode the body surface a decode error carrying
// the limit, and can map it (errors.AsType) to a 413.
//
// Unlike a Content-Length header check, this enforces the limit on the
// actual byte stream, so a client cannot stream an unbounded body past a
// forged or absent Content-Length.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
