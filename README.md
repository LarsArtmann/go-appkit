# go-appkit

Shared application skeleton for Go services: HTTP server with health endpoint,
graceful shutdown, structured logging, and SQLite connection setup.

## Install

```bash
go get github.com/larsartmann/go-appkit
```

## Usage

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
