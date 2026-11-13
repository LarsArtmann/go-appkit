# httpx — the handler-DX battery

`httpx` is the "return (T, error), get a correct HTTP response" seam for
appkit services: JSON serialization of the happy path, classification-aware
error responses, bind + validation for request bodies, and a no-leak
guarantee on every error body — in one handler wrapper.

Part of [go-appkit](../README.md). Proprietary; all batteries are opt-in.

## Install

```bash
go get github.com/larsartmann/go-appkit/httpx@latest
```

## Quick start

```go
import "github.com/larsartmann/go-appkit/httpx"

mux.Handle("GET /api/things/{id}", httpx.ResultHandler(
	func(c *httpx.Context) (*Thing, error) {
		return svc.Get(c.Request.Context(), c.PathValue("id"))
	},
))
```

A nil error serializes as JSON 200 (204 for a nil pointer — no payload);
a non-nil error renders through the classified error contract below.

## Bind + validation

```go
// CreateThingRequest is a tagged ADAPTER type — never an untagged domain
// struct. json/v2 matches field names case-sensitively: a camelCase wire
// field silently misses a snake_case tag and binds the zero value. Wire
// tags are snake_case by convention (the tagliatelle stance).
type CreateThingRequest struct {
	AppID string `json:"app_id"`
}

func (r *CreateThingRequest) Validate() error {
	if r.AppID == "" {
		return errorfamily.Newf(errorfamily.Rejection, "thing.app_id_required", "app_id is required")
	}

	return nil
}

mux.Handle("POST /api/things", httpx.ResultHandlerWithValidation(
	func(c *httpx.Context, in CreateThingRequest) (*Thing, error) {
		return svc.Create(c.Request.Context(), in)
	},
))
```

Decode failures are Rejection/400 under the `bind.*` codes:
`bind.empty_body`, `bind.body_too_large` (1 MiB bind limit),
`bind.unreadable_body`, `bind.invalid_body`.

## The error contract (one taxonomy, two renderings)

Errors map through go-error-family exactly like the
[errorpages](../errorpages) module: same classification, same status
(`errorfamily.HTTPStatus`), same JSON field names. Mount errorpages for the
pretty HTML surface and httpx for the API surface — the two cannot disagree;
the parity contract is pinned by this module's tests.

| Family         | Status | Message on the wire                          |
| -------------- | ------ | -------------------------------------------- |
| Rejection      | 400    | the error's own text, leak patterns scrubbed |
| Conflict       | 409    | the error's own text, leak patterns scrubbed |
| Transient      | 503    | the family default — raw text never leaves   |
| Corruption     | 500    | the family default — raw text never leaves   |
| Infrastructure | 503    | the family default — raw text never leaves   |
| Orchestration  | 500    | the family default — raw text never leaves   |

```json
{
  "family": "infrastructure",
  "code": "db.creds_unreadable",
  "message": "An infrastructure failure interrupted the request.",
  "why": "...",
  "fix": "..."
}
```

## No-leak guarantee (B9)

Error bodies never carry internals: server-side families render the family
default message (paths, SQL, stack frames, driver text never leave the
process); client-facing messages (Rejection/Conflict) keep their authored
text but pass a leak-pattern scrub (POSIX/Windows paths, `.go:NN` source
locations, SQL statements, goroutine headers). An error type with a
`Public() string` method opts its vouched-for text in verbatim — the same
ladder errorpages honors. The diagnostic extras (cause chains, timestamps,
error context maps) are deliberately NOT serialized.

## The ResponseWriter contract (B3)

A wrapper that hides `http.Flusher`/`http.Hijacker` compiles clean, passes
Recorder tests, and stalls production SSE. `WrapWriter` is the blessed
pass-through wrapper (forwards Flusher, Hijacker, and `Unwrap` for
`http.ResponseController`), and `httpxtest.AssertForwardsInterfaces` pins
your own wrappers: it builds a probe base, applies your wrapper factory,
and fails on hidden interfaces, swallowed writes, no-op stubs, and
controller-opaque chains — the buggy plain-embedding shape that Recorder
tests miss fails here.

```go
httpxtest.AssertForwardsInterfaces(t, func(base http.ResponseWriter) http.ResponseWriter {
	return myWrapper{base}
})
```

## Conditional GET (B4)

`Conditional` serves a handler with RFC 9110 conditional semantics: a
strong sha256 ETag over the response body (via the family's go-etag —
weak `W/` forms, comma lists, and `*` in If-None-Match all answer real
304s) plus Last-Modified/If-Modified-Since when the handler stamps
`SetLastModified`. If-None-Match takes precedence per the RFC.

```go
mux.Handle("GET /feed.ics", httpx.Conditional(feedHandler))
```

Byte-determinism is the prerequisite (the ETag hashes response bytes);
responses over 1 MiB stream through untagged; `SkipConditional` excludes
endpoints where buffering is wrong (SSE, large downloads).

## Content negotiation (B5)

`Negotiate` answers "which representation?" in one line — q-values,
wildcards, refusal (`q=0`), and the offers' own order as server
preference; `""` means the client accepts none of them (answer 406 or
your default):

```go
switch httpx.Negotiate(r, "text/html", "application/json") { ... }
```

The go-codec seam (the composition-mission ruling: build the encode
boundary on the family codec): `RequestCodec` resolves the decode codec
from Content-Type (JSON/CBOR, structured `+json`/`+cbor` suffixes; 415
through `WriteError` on anything else) and `ResponseCodec` resolves the
encode codec from Accept (q ranking, tie prefers CBOR, q=0 refuses,
falls back to the request codec when no preference is expressed).

## Per-route write deadlines (B7)

`WriteDeadline` extends the connection's write deadline for matching
prefixes only — the scalpel next to core's `NoTimeout` sledgehammer:

```go
svc.Handler = httpx.WriteDeadline(3*time.Minute, "/api/agent/")(svc.Mux)
```

It must sit OUTSIDE wrappers that neither forward the deadline controls
nor `Unwrap`; through such a wrapper the deadline cannot be set, and the
request fails LOUDLY with a classified Infrastructure error (500) —
never a silent hang (pinned by the ordering-regression test).

## Build & verify

```bash
cd httpx && GOWORK=off GOTOOLCHAIN=go1.27.4 go test ./... -race -count=1
cd httpx && golangci-lint run ./...
```

## Scope

The handler-DX batteries land in waves; v0.1.0 shipped B1
(ResultHandler), B2 (bind + validation), and B9 (no-leak errors) plus the
errorpages parity pin. v0.2.0 ships B3 (ResponseWriter contract +
`httpxtest`), B4 (conditional GET on go-etag), B5 (content negotiation +
the go-codec seam), and B7 (per-route write deadlines). Remaining: B6/B8
route introspection (owner-gated on the H8 seam decision — if adopted,
httpx consumes `httputil/endpoints`, it never grows a second registry).
