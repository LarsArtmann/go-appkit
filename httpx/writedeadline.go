package httpx

import (
	"net/http"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// WriteDeadline returns a middleware that extends the connection's write
// deadline by d, but only for requests whose path matches one of the
// prefixes — the scalpel next to core's NoTimeout sledgehammer:
//
//	mux := http.NewServeMux()
//	mux.Handle("POST /api/agent/", slowLLMHandler)
//
//	svc.Handler = httpx.WriteDeadline(3*time.Minute, "/api/agent/")(svc.Mux)
//
// Long missions (LLM streams, slow exports) get room to write; every
// other route keeps the service's WriteTimeout. The extension applies at
// request start, before the handler runs, so a handler that outlives the
// original deadline still writes successfully.
//
// ORDERING CONTRACT: this middleware must sit OUTSIDE any wrapper that
// neither implements the deadline controls itself nor Unwrap — a gzip
// wrapper that hides the ResponseWriter chain breaks
// http.ResponseController's reach, and the deadline silently fails to
// extend. The failure is not silent here: when the deadline cannot be
// set on a matching request, the request fails loudly with a classified
// Infrastructure error (the same "wrapper hides the connection" class
// the B3 battery polices) instead of hanging until the original
// deadline kills the connection mid-response.
//
// Prefixes match literally, including any trailing slash — pass
// "/api/agent/" to cover the subtree, "/api/agent/report" to cover a
// single route and everything under it.
func WriteDeadline(extension time.Duration, prefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !matchesAnyPrefix(r.URL.Path, prefixes) {
				next.ServeHTTP(w, r)

				return
			}

			rc := http.NewResponseController(w)
			if err := rc.SetWriteDeadline(time.Now().Add(extension)); err != nil {
				WriteError(w, r, errorfamily.NewInfrastructure(
					"httpx.write_deadline_unreachable",
					"extending the write deadline failed: the ResponseWriter chain hides the connection "+
						"(a wrapper above this middleware neither forwards the deadline controls nor Unwrap)",
				).WithHTTPStatus(http.StatusInternalServerError)) // wiring bug, not an outage: 500, not the family's 503

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func matchesAnyPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if prefix != "" && strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
