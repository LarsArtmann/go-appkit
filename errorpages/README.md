# go-appkit/errorpages

Pretty, classified error pages (HTML) and error contracts (JSON) for
[go-appkit](../README.md) services, bridging
[templ-components/errorpage](https://github.com/LarsArtmann/templ-components)
to the same [go-error-family](https://github.com/LarsArtmann/go-error-family)
taxonomy appkit already uses for `HTTPStatus()`.

> **Build note:** requires `GOEXPERIMENT=jsonv2` (errorpage uses
> `encoding/json/v2` for the JSON contract).

## What you get

- **HTML**: the templ-components error page — family-styled, with title,
  message, why, fix, and cause chain — wrapped in a minimal HTML shell.
- **JSON**: the machine-readable contract
  `{family, code, message, title, why, fix, context}` that htmx
  `GlobalErrorHandling` turns into toasts client-side.
- **One taxonomy**: Rejection → 400, Conflict → 409, Transient → 503,
  Corruption → 500, Infrastructure → 503. Identical to `appkit.HTTPStatus`.
- **Pretty 404/405**: mount once; unmatched paths and method mismatches stop
  being bare stdlib responses.

## Usage

```go
svc, _ := appkit.NewService(appkit.DefaultServiceConfig())

// Pretty 404 for anything unmatched (JSON when the client accepts it):
errorpages.Mount(svc.Mux, errorpages.Config{})

// Classified errors from your own handlers:
svc.Mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    user, err := findUser(r.PathValue("id"))
    if err != nil {
        errorpages.Write(w, r, err, errorpages.Config{}) // 400/409/503/500 by family
        return
    }
    // ...
})
```

API-only service? Force JSON everywhere:

```go
errorpages.Mount(svc.Mux, errorpages.Config{
    JSONWhen: func(*http.Request) bool { return true },
})
```

### Pretty 405s too

`Mount` cannot intercept the stdlib mux's internal 405 (it is generated inside
`ServeMux.ServeHTTP`). If you control the top-level handler, use `Wrap`
instead — it renders pretty 404 **and** 405 and preserves the `Allow` header:

```go
handler := errorpages.Wrap(mux, errorpages.Config{})
http.ListenAndServe(":8080", handler)
```

## Content negotiation

Per request: JSON when the `Accept` header names `application/json`, HTML
otherwise (browsers). Replace the rule with `Config.JSONWhen`.

## Configuration

| Field      | Type                          | Default         | Effect                                                     |
| ---------- | ----------------------------- | --------------- | ---------------------------------------------------------- |
| `NotFound` | `*errorpage.NotFound404Props` | library default | Customizes the 404 page (links, search, copy).             |
| `Nonce`    | `string`                      | ""              | CSP nonce for inline styles/scripts.                       |
| `Lang`     | `string`                      | "en"            | `<html lang>` for standalone pages.                        |
| `JSONWhen` | `func(*http.Request) bool`    | Accept-based    | Custom JSON/HTML decision rule.                            |
| `Override` | `func(error, props) *props`   | nil             | Per-error prop customization (same as errorpage upstream). |

## API surface

| Symbol                   | Purpose                                                              |
| ------------------------ | -------------------------------------------------------------------- |
| `Mount(mux, cfg)`        | Catch-all pretty 404 on a `*http.ServeMux` (the appkit integration). |
| `Wrap(mux, cfg) Handler` | Top-level handler adding pretty 404 + 405 (Allow preserved).         |
| `Handler(err, cfg)`      | `http.Handler` rendering one classified error.                       |
| `Write(w, r, err, cfg)`  | One-call rendering inside handlers.                                  |

## Example

See [example/main.go](example/main.go) — run it and probe the endpoints
listed in its doc comment.
## Build & verify

```bash
cd errorpages
GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.1 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.1 go build ./...
golangci-lint run ./...        # from this directory
```

`GOTOOLCHAIN=go1.27.1` is required for every command below on
machines whose default toolchain is older (this repo pins
`go 1.27.1` in every go.mod). `GOWORK=off` makes the run
hermetic: the module resolves exactly what its own go.mod pins,
via the module proxy — how consumers resolve it.

