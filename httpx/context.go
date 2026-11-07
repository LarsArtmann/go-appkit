package httpx

import (
	"net/http"
)

// Context carries what a result handler needs from the HTTP layer: the
// raw ResponseWriter and Request plus the stdlib path-value accessors.
// It is deliberately thin — the battery owns serialization and error
// rendering, not request plumbing. Handlers that need more reach the
// underlying types directly.
type Context struct {
	// ResponseWriter is the live writer for the request. Writing to it
	// directly (file downloads, SSE, custom encodings) is supported: a
	// handler that writes before returning takes ownership of the
	// response, and the battery skips its own serialization.
	ResponseWriter http.ResponseWriter

	// Request is the incoming request, never nil.
	Request *http.Request
}

// PathValue returns the value of the named wildcard in the route pattern
// that matched the request (stdlib ServeMux semantics; empty string when
// the pattern has no such wildcard).
func (c *Context) PathValue(name string) string {
	return c.Request.PathValue(name)
}

// Header returns the request headers.
func (c *Context) Header(name string) string {
	return c.Request.Header.Get(name)
}

// Query returns the first query parameter value for the given key.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}
