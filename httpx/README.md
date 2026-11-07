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

| Family         | Status | Message on the wire                              |
| -------------- | ------ | ------------------------------------------------ |
| Rejection      | 400    | the error's own text, leak patterns scrubbed     |
| Conflict       | 409    | the error's own text, leak patterns scrubbed     |
| Transient      | 503    | the family default — raw text never leaves       |
| Corruption     | 500    | the family default — raw text never leaves       |
| Infrastructure | 503    | the family default — raw text never leaves       |
| Orchestration  | 500    | the family default — raw text never leaves       |

```json
{"family":"infrastructure","code":"db.creds_unreadable","message":"An infrastructure failure interrupted the request.","why":"...","fix":"..."}
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

## Build & verify

```bash
cd httpx && GOWORK=off GOTOOLCHAIN=go1.27.3 go test ./... -race -count=1
cd httpx && golangci-lint run ./...
```

## Scope

The handler-DX batteries land in waves; v0.1.0 ships B1 (ResultHandler),
B2 (bind + validation), and B9 (no-leak errors) plus the errorpages parity
pin. Later waves: B3 ResponseWriter contract, B4 conditional GET,
B5 content negotiation, B7 per-route write deadlines, B6/B8 route
introspection (owner-gated on the H8 seam decision).
