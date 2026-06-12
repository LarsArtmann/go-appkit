# go-appkit

Shared application skeleton for Go services. It wires together the boilerplate every small-to-medium Go service needs: a structured logger, an HTTP server with a health endpoint, graceful shutdown, and SQLite setup.

## Install

```bash
go get github.com/larsartmann/go-appkit
```

Requires Go 1.26.3 or later.

## Features

- **HTTP server wrapper** with configurable timeouts, automatic `GET /health` registration, and graceful shutdown.
- **Health handlers** with default and custom status responses.
- **Graceful shutdown** on `SIGINT`/`SIGTERM` with a configurable timeout.
- **Structured logging** via `log/slog` with JSON, text, or auto-detected output.
- **SQLite setup** using the CGO-free `modernc.org/sqlite` driver with WAL-mode pragmas.

## Quick start

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "time"

    appkit "github.com/larsartmann/go-appkit"
)

func main() {
    logger := appkit.InitLogger(appkit.LoggerConfig{Level: "info"})

    mux := http.NewServeMux()
    mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("hello"))
    })

    srv := appkit.NewServer(appkit.DefaultServerConfig(), mux)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        if err := srv.Start(ctx); err != nil {
            logger.Error("server error", "error", err)
        }
    }()

    if err := appkit.WaitForSignal(ctx, appkit.DefaultShutdownConfig(), func(shutdownCtx context.Context) error {
        return srv.Shutdown(shutdownCtx)
    }); err != nil {
        logger.Error("shutdown error", "error", err)
    }
}
```

## Components

### Logger

`InitLogger` returns a `*slog.Logger`. Supported levels are `debug`, `info`, `warn`, and `error`. The format can be `json`, `text`, or `auto` (text when stderr is a terminal, JSON otherwise). An invalid level panics.

```go
logger := appkit.InitLogger(appkit.LoggerConfig{
    Level:  "debug",
    Format: "json",
})
```

### HTTP Server

`NewServer` wraps `http.Server` and automatically registers `GET /health`. You can override the health handler or any timeout via `ServerConfig`.

```go
cfg := appkit.ServerConfig{
    Port:         8080,
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  120 * time.Second,
}

srv := appkit.NewServer(cfg, mux)
```

Zero-value fields are replaced with defaults from `DefaultServerConfig`.

### Health handler

```go
// Default JSON response: {"status":"ok"}
mux.HandleFunc("GET /health", appkit.DefaultHealthHandler)

// Custom status value
mux.HandleFunc("GET /health", appkit.NewHealthHandler("ready"))
```

### Graceful shutdown

`WaitForSignal` blocks until `SIGINT`/`SIGTERM` or the context is cancelled, then calls your shutdown function with a timeout-bound context.

```go
err := appkit.WaitForSignal(ctx, appkit.DefaultShutdownConfig(), func(shutdownCtx context.Context) error {
    return srv.Shutdown(shutdownCtx)
})
```

### SQLite

`OpenSQLite` opens a SQLite database with sensible defaults: WAL mode, `busy_timeout = 5000`, and `foreign_keys = ON`. You can override pragmas or connection limits via `SQLiteConfig`.

```go
db, err := appkit.OpenSQLite(ctx, appkit.SQLiteConfig{
    Path:         "app.db",
    MaxOpenConns: 10,
})
if err != nil {
    // handle error
}
defer db.Close()
```

## Development

This project uses only the standard Go toolchain:

```bash
go test ./...
go vet ./...
go build ./...
```

## License

MIT
