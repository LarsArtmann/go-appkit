// Package httpx is the handler-DX battery for appkit services: the
// "return (T, error), get a correct HTTP response" seam.
//
// The core of the package is [ResultHandler]: wrap a function that takes a
// request context and returns a value or an error, and the battery owns the
// rest — JSON serialization of the happy path, classification-aware error
// responses (status from [errorfamily.HTTPStatus], the same taxonomy the
// errorpages module renders), and a no-leak guarantee on every error body
// (server-side families never echo raw error text; client-facing text is
// pattern-scrubbed).
//
// # Sketch
//
//	mux.Handle("GET /api/things/{id}", httpx.ResultHandler(
//	    func(c *httpx.Context) (*Thing, error) {
//	        return svc.Get(c.Request.Context(), c.PathValue("id"))
//	    },
//	))
//
// With bind + validation (the JSON body decodes into B, a tagged adapter
// type — never an untagged domain struct; json/v2 matches case-sensitively):
//
//	mux.Handle("POST /api/things", httpx.ResultHandlerWithValidation(
//	    func(c *httpx.Context, in CreateThingRequest) (*Thing, error) {
//	        return svc.Create(c.Request.Context(), in)
//	    },
//	))
//
// # One taxonomy, two renderings
//
// Errors flow through go-error-family exactly like errorpages: same
// classification, same status mapping, same JSON field names. A service
// can mount errorpages for the HTML surface and httpx for the API surface
// without the two disagreeing on what a Rejection looks like. The parity
// contract is pinned by this module's tests, not by convention.
//
// # The rest of the battery
//
//   - [WrapWriter] + httpxtest.AssertForwardsInterfaces — the
//     ResponseWriter wrapper contract as code (B3): hidden Flusher or
//     Hijacker is the production SSE-stall class.
//   - [Conditional] — RFC 9110 conditional GET (B4): strong sha256 ETag
//     via go-etag plus Last-Modified/If-Modified-Since, real 304s.
//   - [Negotiate] / [RequestCodec] / [ResponseCodec] — content
//     negotiation (B5), the encode seam built on the family go-codec.
//   - [WriteDeadline] — per-route write-deadline extension (B7), the
//     scalpel next to core's NoTimeout sledgehammer.
//
// All batteries are opt-in: registering a httpx handler changes nothing
// about the default middleware stack or any other route.
package httpx
