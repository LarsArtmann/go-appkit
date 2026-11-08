package httpx

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	etagserver "github.com/larsartmann/go-etag/server"
)

// ConditionalOption configures a [Conditional] handler.
type ConditionalOption func(*conditionalConfig)

type conditionalConfig struct {
	etag etagserver.ETagConfig
	skip func(*http.Request) bool
}

// ETagFunc overrides how the response body's entity tag is computed. The
// function receives the response bytes and returns the opaque tag value
// (the wire form gains the quotes and any weakness prefix). The default
// is a strong sha256 of the body — byte-deterministic and collision-proof,
// the right default for feeds and calendars that pollers re-fetch.
func ETagFunc(f func(body []byte) string) ConditionalOption {
	return func(c *conditionalConfig) { c.etag.HashFunc = f }
}

// SkipConditional passes matching requests through untouched — no body
// buffering, no validators. Use it for endpoints where buffering is wrong:
// SSE streams, large file downloads, anything streamed.
func SkipConditional(pred func(*http.Request) bool) ConditionalOption {
	return func(c *conditionalConfig) { c.skip = pred }
}

// Conditional serves h with conditional-request semantics (RFC 9110
// §13): a strong ETag computed from the response body (sha256 by default,
// via the family's go-etag server middleware — weak W/ forms, comma
// lists, and * in If-None-Match all answer real 304s) plus
// Last-Modified/If-Modified-Since when the handler stamps the response
// with [SetLastModified].
//
//	mux.Handle("GET /feed.ics", httpx.Conditional(feedHandler))
//
// Precedence follows the RFC: If-None-Match decides when present and
// If-Modified-Since is ignored; only when the client sends no
// If-None-Match does the time validator run, answering 304 when the
// handler's Last-Modified is not newer than the request's
// If-Modified-Since. A 304 keeps ETag and Last-Modified and drops the
// body-bearing headers.
//
// The ETag hashes the response BYTES: byte-determinism is the
// prerequisite. If the same logical content can serialize differently
// (map iteration order, floating point, timestamps), stabilize the
// serialization or clients will see churning tags. Responses larger than
// 1 MiB stream through untagged (go-etag's buffer bound); streaming
// responses that Flush bypass validators entirely.
//
// GET and HEAD only — conditional semantics for unsafe methods
// (If-Match/412) are a different battery and deliberately not smuggled
// in here.
func Conditional(handler http.Handler, opts ...ConditionalOption) http.Handler {
	cfg := conditionalConfig{etag: etagServerConfig(), skip: nil}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	etagLayer := etagserver.New(cfg.etag)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.skip != nil && cfg.skip(r) {
			handler.ServeHTTP(w, r)

			return
		}

		ims := &imsWriter{ResponseWriter: w, req: r, decided: false, suppress: false}
		etagLayer(handler).ServeHTTP(ims, r)
		ims.finish()
	})
}

// etagServerConfig is the httpx default ETag posture: strong sha256 tags,
// go-etag's 1 MiB buffer bound, every other behavior upstream's default.
func etagServerConfig() etagserver.ETagConfig {
	cfg := etagserver.DefaultETagConfig()
	cfg.HashFunc = sha256Tag
	cfg.Strength = etagserver.Strong

	return cfg
}

// sha256Tag is the battery's default validator function.
func sha256Tag(body []byte) string {
	sum := sha256.Sum256(body)

	return hex.EncodeToString(sum[:])
}

// SetLastModified stamps w with the Last-Modified validator (UTC, IMF-fixdate
// — the format If-Modified-Since round-trips). A handler that knows its
// resource's modification time should set it before writing the body;
// [Conditional] answers If-Modified-Since with 304 when the client's copy
// is not older than this.
func SetLastModified(w http.ResponseWriter, t time.Time) {
	w.Header().Set("Last-Modified", t.UTC().Format(http.TimeFormat))
}

// imsWriter is the If-Modified-Since half of Conditional. It sits OUTSIDE
// the ETag layer and acts only when the request carries no If-None-Match
// (RFC precedence): a cacheable response whose Last-Modified is not newer
// than the request's If-Modified-Since is converted to a bodyless 304.
type imsWriter struct {
	http.ResponseWriter

	req      *http.Request
	decided  bool
	suppress bool
}

// decide runs the time-validator check exactly once, at the first commit
// point (WriteHeader, Write, or Flush — whichever the inner chain hits
// first). Returns true when the response is being suppressed.
func (w *imsWriter) decide() {
	if w.decided {
		return
	}

	w.decided = true
	w.suppress = w.notModifiedSince()
}

// notModifiedSince reports whether this response should become a 304:
// GET/HEAD, a parseable If-Modified-Since, no If-None-Match (precedence),
// and a Last-Modified the client's copy already has.
func (w *imsWriter) notModifiedSince() bool {
	if w.req.Method != http.MethodGet && w.req.Method != http.MethodHead {
		return false
	}

	if w.req.Header.Get("If-None-Match") != "" {
		return false
	}

	ifModifiedSince := w.req.Header.Get("If-Modified-Since")
	if ifModifiedSince == "" {
		return false
	}

	clientTime, err := http.ParseTime(ifModifiedSince)
	if err != nil {
		return false
	}

	lastModified := w.Header().Get("Last-Modified")
	if lastModified == "" {
		return false
	}

	modifiedAt, err := http.ParseTime(lastModified)
	if err != nil {
		return false
	}

	return !modifiedAt.After(clientTime)
}

func (w *imsWriter) WriteHeader(status int) {
	w.decide()

	if w.suppress {
		return
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *imsWriter) Write(b []byte) (int, error) {
	w.decide()

	if w.suppress {
		return len(b), nil // pretend success: the inner layer must not error
	}

	return w.ResponseWriter.Write(b) //nolint:wrapcheck // pass-through by design
}

// Flush forwards the flush unless the response is being suppressed — a
// suppressed response has nothing to stream.
func (w *imsWriter) Flush() {
	w.decide()

	if w.suppress {
		return
	}

	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap keeps http.ResponseController reachability through this layer.
func (w *imsWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// finish commits the suppressed 304 after the inner chain fully returned:
// body-bearing headers drop, validators stay, the status is 304.
func (w *imsWriter) finish() {
	if !w.suppress {
		return
	}

	w.Header().Del("Content-Length")
	w.Header().Del("Content-Type")

	w.ResponseWriter.WriteHeader(http.StatusNotModified)
}
