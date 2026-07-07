# go-appkit

Production-ready HTTP service framework for Go. Composes [httputil](https://github.com/LarsArtmann/httputil),
[charmbracelet/log](https://github.com/charmbracelet/log), and [go-error-family](https://github.com/LarsArtmann/go-error-family)
into one coherent service lifecycle.

## Quick start

```go
package main

import (
	"context"

	appkit "github.com/larsartmann/go-appkit"
	errorfamily "github.com/larsartmann/go-error-family"
	"net/http"
)

func main() {
	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:     ":8080",
		LogLevel: appkit.LogLevelInfo,
	})
	if err != nil {
		panic(err)
	}
	defer svc.Close()

	svc.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	if err := svc.Run(context.Background()); err != nil {
		os.Exit(errorfamily.HandleError(err))
	}
}
```

That gives you:

- HTTP server with graceful drain + shutdown on SIGINT/SIGTERM
- Middleware: Recovery → RequestID → Logging → Timeout → SecurityHeaders
- Health endpoints: `GET /health`, `GET /health/live`, `GET /health/ready`
- Pretty structured logging via charmbracelet/log
- Error classification via go-error-family

## Install

```bash
go get github.com/larsartmann/go-appkit
```

Requires Go 1.26.3 or later.

## Configuration

All config is via `ServiceConfig`. Zero-value fields get production defaults:

| Field              | Type                    | Default   | Description                                      |
| ------------------ | ----------------------- | --------- | ------------------------------------------------ |
| `Addr`             | `string`                | `":8080"` | Listen address                                   |
| `LogLevel`         | `LogLevel`              | `"info"`  | Log level: debug, info, warn, error              |
| `LogFormat`        | `LogFormat`             | `"auto"`  | Log format: text, json, auto                     |
| `ReadTimeout`      | `time.Duration`         | `10s`     | HTTP read timeout                                |
| `WriteTimeout`     | `time.Duration`         | `30s`     | HTTP write timeout                               |
| `IdleTimeout`      | `time.Duration`         | `60s`     | HTTP idle timeout                                |
| `ShutdownTimeout`  | `time.Duration`         | `15s`     | Max time to wait for shutdown                    |
| `DrainDelay`       | `time.Duration`         | `5s`      | Delay after flipping ready probe before shutdown |
| `Middlewares`      | `[]httputil.Middleware` | `nil`     | Replace the default middleware stack             |
| `ExtraMiddlewares` | `[]httputil.Middleware` | `nil`     | Append to the default middleware stack           |
| `RegisterHealth`   | `*bool`                 | `&true`   | Set to `&false` to opt out of health endpoints   |

## Middleware

The default stack is opinionated but replaceable:

```go
// Use defaults (Recovery → RequestID → Logging → Timeout → SecurityHeaders):
svc, _ := appkit.NewService(appkit.ServiceConfig{Addr: ":8080"})

// Replace the entire stack:
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr:        ":8080",
    Middlewares: []httputil.Middleware{
        httputil.Recovery(logger),
        httputil.RequestID(httputil.DefaultRequestIDConfig()),
    },
})

// Extend the default stack:
svc, _ := appkit.NewService(appkit.ServiceConfig{
    Addr:             ":8080",
    ExtraMiddlewares: []httputil.Middleware{myMiddleware},
})
```

## Health endpoints

Three endpoints registered by default (via [httputil](https://github.com/LarsArtmann/httputil)):

- `GET /health` — liveness (always 200 `{"status":"up"}`)
- `GET /health/live` — Kubernetes liveness probe
- `GET /health/ready` — Kubernetes readiness probe (flips to 503 during graceful drain)

The readiness probe is connected to the graceful drain sequence: when `Shutdown` is called,
the probe immediately starts returning 503 so load balancers stop sending traffic before the
server stops accepting connections.

## Lifecycle

```go
// Simple: Run blocks until signal/error, handles shutdown internally.
err := svc.Run(ctx)

// Advanced: Start returns a channel, you manage the lifecycle.
errCh, err := svc.Start()
// ... start other servers, workers, etc.
err := svc.Shutdown(ctx)
```

### Graceful drain sequence

When `Shutdown` is called:

1. Ready probe flips to 503 (load balancer stops sending new traffic)
2. Wait `DrainDelay` (default 5s) for LB propagation
3. `server.Shutdown(ctx)` stops accepting + finishes in-flight requests

## Error handling

Errors use [go-error-family](https://github.com/LarsArtmann/go-error-family) constructors:

```go
import errorfamily "github.com/larsartmann/go-error-family"

// In your handlers — wrap errors at construction:
err := errorfamily.NewRejection("user.not_found", "user %s not found", id)

// appkit re-exports for convenience:
status := appkit.HTTPStatus(err)  // 400 for Rejection
appkit.LogError(err, logger)      // auto-severity (Transient→Warn, others→Error)
```

For HTTP handlers that return errors, use `errorfamily.HTTPHandler`:

```go
svc.Mux.Handle("POST /users", errorfamily.HTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
    // Return a classified error — HTTPHandler maps family→status, writes safe JSON.
    return errorfamily.NewRejection("user.invalid", "name is required")
}))
```

## When NOT to use appkit

- **You want a specific router** (chi, gin, echo): Use [httputil](https://github.com/LarsArtmann/httputil) directly — it gives you middleware without opinionated server lifecycle.
- **You want type-safe API generation**: Use [Huma](https://github.com/danielgtaylor/huma) and wrap `svc.Mux` with `humago.New`.
- **You want a full-stack framework**: Use [Buffalo](https://gobuffalo.io) or [GoFr](https://gofr.dev).

appkit is for services that use Go stdlib `http.ServeMux` (Go 1.22+ method routing) and want
middleware, health, logging, and shutdown wired in one import.

## Development

Standard Go toolchain:

```bash
go test ./...
go vet ./...
go build ./...
```

## License

MIT
