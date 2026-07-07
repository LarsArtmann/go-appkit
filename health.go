package appkit

import (
	"net/http"

	"github.com/larsartmann/httputil"
)

// RegisterHealth registers /health, /health/live, and /health/ready on the given mux
// using httputil's default handlers. This is a convenience wrapper around
// httputil.RegisterHealth.
func RegisterHealth(mux *http.ServeMux) {
	httputil.RegisterHealth(mux)
}

// ReadyHandlerWithProbe returns an http.HandlerFunc that calls the provided readiness
// function on each request. When ready returns true, responds 200 {"status":"up"}.
// When false, responds 503 {"status":"down"}. Used for graceful drain: flip the probe
// to false to make Kubernetes stop sending traffic before shutting down.
func ReadyHandlerWithProbe(ready func() bool) http.HandlerFunc {
	return httputil.ReadyHandlerWithProbe(ready)
}
