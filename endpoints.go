package appkit

import (
	"fmt"
	"net/http"

	errorfamily "github.com/larsartmann/go-error-family"
	httputil "github.com/larsartmann/httputil"
	"github.com/larsartmann/httputil/endpoints"
)

// realmEndpoints is the Basic Auth realm the endpoint listing answers with.
const realmEndpoints = "endpoints"

// endpointListingSummary is the listing's description of its own row: the
// listing is a route like every other, so it lists itself.
const endpointListingSummary = "Route listing: every route this service serves"

// EndpointsConfig enables the opt-in route listing on a [Service]. Every
// route registered through [Service.Handle] / [Service.HandleFunc] — and
// every framework route (health, metrics, version) — is mounted on the mux
// AND recorded in the same declaration, and the listing serves the resulting
// table: JSON for API clients, an HTML table for browsers, negotiated on the
// Accept header. http.ServeMux cannot enumerate its own patterns, so without
// this registry a route table is always a second, drift-prone list.
//
//	cfg.Endpoints = &appkit.EndpointsConfig{
//	    BasicAuthUser: "ops",
//	    BasicAuthPass: os.Getenv("ENDPOINTS_PASS"),
//	}
//
// # Authentication is mandatory by default
//
// The listing IS the service's route topology: every path, every method,
// with a description. Protect it with the built-in Basic Auth, wrap it in the
// service's own auth middleware through Guard, or set AllowUnauthenticated
// explicitly (loopback-only deployments, or a proxy that enforces auth in
// front).
type EndpointsConfig struct {
	// Path is the listing's ServeMux pattern. Default: "GET /endpoints".
	Path string

	// Guard, when set, wraps the listing handler with the service's own
	// auth middleware (an API key, a session, an admin gate). It replaces
	// the built-in Basic Auth, and it owns the policy: the listing is served
	// to whoever the guard lets through.
	Guard httputil.Middleware

	// BasicAuthUser and BasicAuthPass enable HTTP Basic Auth on the listing
	// (constant-time comparison, the same guard as the metrics surface).
	// Ignored when Guard is set.
	BasicAuthUser string
	BasicAuthPass string

	// AllowUnauthenticated serves the listing without any guard. Opt-in
	// only: use it when a proxy in front enforces authentication, or for
	// loopback-only services. Ignored when Guard or Basic Auth is set.
	AllowUnauthenticated bool
}

// applyEndpointsDefaults fills zero values.
func (c *EndpointsConfig) applyEndpointsDefaults() {
	if c.Path == "" {
		c.Path = endpoints.DefaultPattern
	}
}

// validate rejects an unprotected listing at construction time. A custom
// Guard counts as protection: the service that supplied it owns its policy.
func (c *EndpointsConfig) validate() error {
	protected := c.Guard != nil ||
		c.BasicAuthUser != "" ||
		c.BasicAuthPass != "" ||
		c.AllowUnauthenticated

	if !protected {
		return fmt.Errorf(
			"%w: the endpoint listing is unauthenticated — set Guard, set BasicAuthUser/BasicAuthPass, "+
				"or set AllowUnauthenticated explicitly (a proxy enforces auth, or loopback-only deployment)",
			errInvalidEndpointsConfig)
	}

	return nil
}

// errInvalidEndpointsConfig rejects unsafe endpoint-listing configurations.
var errInvalidEndpointsConfig error = errorfamily.NewRejection(
	"appkit.endpoints_config_invalid", "invalid endpoint listing configuration")

// handler returns the listing handler behind the configured protection:
// the service's own Guard, the built-in Basic Auth, or nothing (the
// explicitly allowed unauthenticated mode).
func (c *EndpointsConfig) handler(registry *endpoints.Registry) http.Handler {
	switch {
	case c.Guard != nil:
		return c.Guard(registry.Handler())

	case c.BasicAuthUser != "" || c.BasicAuthPass != "":
		return basicAuth(c.BasicAuthUser, c.BasicAuthPass, realmEndpoints, registry.Handler())

	default:
		return registry.Handler()
	}
}
