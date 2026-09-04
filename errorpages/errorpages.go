package errorpages

import (
	"net/http"
	"path"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/templ-components/errorpage"
)

// Config controls how error pages are rendered. The zero value is ready to
// use: default 404 props, HTML with Accept-based JSON negotiation.
type Config struct {
	// NotFound customizes the catch-all 404 page. Zero value uses
	// errorpage.DefaultNotFound404Props().
	NotFound *errorpage.NotFound404Props

	// Nonce is used for CSP-compliant inline styles/scripts in HTML pages.
	Nonce string

	// Lang sets the <html lang> attribute of standalone HTML pages.
	// Default: "en".
	Lang string

	// JSONWhen decides per request whether the JSON contract is served
	// instead of HTML. Default: JSON when the Accept header names
	// application/json (browsers get HTML). Return true unconditionally for
	// API-only services.
	JSONWhen func(r *http.Request) bool

	// Override allows per-error customization of the ErrorPageProps before
	// rendering (same semantics as errorpage.ErrorHandlerConfig.Override).
	Override func(err error, props errorpage.ErrorPageProps) *errorpage.ErrorPageProps
}

// wantsJSON applies the default Accept-header negotiation: JSON when the
// client names application/json, HTML otherwise.
func (c Config) wantsJSON(r *http.Request) bool {
	if c.JSONWhen != nil {
		return c.JSONWhen(r)
	}

	return r != nil && strings.Contains(r.Header.Get("Accept"), "application/json")
}

// Mount registers a catch-all handler on mux that renders the pretty 404 page
// (or JSON contract) for any path no other route matched. Method mismatches
// on registered patterns still produce the stdlib mux's plain 405 with the
// Allow header; use Wrap for pretty 405s as well.
//
// It panics like http.ServeMux.Handle when "/" is already registered.
func Mount(mux *http.ServeMux, cfg Config) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cfg.writeNotFound(w, r)
	})
}

// Wrap returns a handler that serves requests through mux but replaces the
// mux's built-in bare 404 and 405 responses with errorpages rendering.
// The Allow header of 405 responses is preserved. Path-cleaning redirects
// (e.g. "/a//b" to "/a/b", or "/x/../y" to "/y") keep the mux's redirect
// behavior; only the canonical request that follows renders a pretty 404.
// Use when you control the top-level handler (custom server setups); with
// appkit.Service, Mount is the simpler integration.
func Wrap(mux *http.ServeMux, cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matched, pattern := mux.Handler(r)
		// An empty pattern also occurs for non-canonical request paths
		// (doubled slashes or dot segments): the mux answers those with a
		// redirect to the cleaned path, not with a 404. Delegate so the
		// redirect survives; the pretty 404 applies to the canonical
		// request that follows it.
		if pattern != "" || cleanPath(r.URL.EscapedPath()) != r.URL.EscapedPath() {
			mux.ServeHTTP(w, r)

			return
		}

		// No route pattern matched: the mux would write a bare 404 or 405.
		// Discern which by executing the internal handler into a throwaway
		// recorder — it only writes a status line and short text.
		rec := &statusRecorder{ //nolint:exhaustruct_v5 // status is deliberately zero; the wrapped handler sets it via WriteHeader
			header: http.Header{},
		}
		matched.ServeHTTP(rec, r)

		switch rec.status {
		case http.StatusMethodNotAllowed:
			for _, allow := range rec.header.Values("Allow") {
				w.Header().Add("Allow", allow)
			}

			cfg.writeMethodNotAllowed(w, r)
		default:
			cfg.writeNotFound(w, r)
		}
	})
}

// Handler returns an http.Handler rendering err as a classified error page
// (HTML) or error contract (JSON) per request content negotiation. The HTTP
// status derives from the error's go-error-family family.
func Handler(err error, cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.writeError(w, r, err, 0)
	})
}

// Write renders err to w according to r's negotiated representation.
// It is the one-call form of Handler for use inside handlers:
//
//	if err := doWork(r); err != nil {
//	    errorpages.Write(w, r, err, cfg)
//	    return
//	}
func Write(w http.ResponseWriter, r *http.Request, err error, cfg Config) {
	cfg.writeError(w, r, err, 0)
}

// writeError renders err with an optional forced status code (0 = derive
// from the error's family).
func (c Config) writeError(w http.ResponseWriter, r *http.Request, err error, forceStatus int) {
	c.errorpageHandler(r, err, forceStatus).ServeHTTP(w, r)
}

// errorpageHandler adapts Config onto errorpage.ErrorHandlerConfig for a
// specific request, layering the status override (for 404/405 synthetic
// errors) under the consumer's own Override hook.
func (c Config) errorpageHandler(
	r *http.Request,
	err error,
	forceStatus int,
) http.Handler {
	return errorpage.ErrorHandler(err, errorpage.ErrorHandlerConfig{
		Nonce:     c.Nonce,
		Lang:      c.Lang,
		HTMLShell: true,
		JSON:      c.wantsJSON(r),
		Override: func(sourceErr error, props errorpage.ErrorPageProps) *errorpage.ErrorPageProps {
			if forceStatus != 0 {
				props.StatusCode = forceStatus
			}

			if c.Override == nil {
				return &props
			}

			return c.Override(sourceErr, props)
		},
	})
}

// writeNotFound renders the 404 page (or JSON contract).
func (c Config) writeNotFound(w http.ResponseWriter, r *http.Request) {
	if c.wantsJSON(r) {
		c.errorpageHandler(
			r,
			errorfamily.NewRejection("http.not_found", "the requested page does not exist"),
			http.StatusNotFound,
		).ServeHTTP(w, r)

		return
	}

	props := errorpage.DefaultNotFound404Props()
	if c.NotFound != nil {
		props = *c.NotFound
	}

	errorpage.WriteNotFound404(w, r, props, c.Nonce)
}

// writeMethodNotAllowed renders a 405 error page (or JSON contract).
func (c Config) writeMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	err := errorfamily.Newf(
		errorfamily.Rejection,
		"http.method_not_allowed",
		"method %s is not allowed for this path", r.Method,
	)

	c.errorpageHandler(r, err, http.StatusMethodNotAllowed).ServeHTTP(w, r)
}

// cleanPath returns the canonical form of p, mirroring net/http's
// unexported cleanPath: path.Clean with any trailing slash (except on the
// root path) preserved.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}

	return np
}

// statusRecorder captures only the status code and headers of a response.
type statusRecorder struct {
	status int
	header http.Header
}

func (s *statusRecorder) Header() http.Header { return s.header }

func (s *statusRecorder) Write(p []byte) (int, error) { return len(p), nil }

func (s *statusRecorder) WriteHeader(status int) { s.status = status }
