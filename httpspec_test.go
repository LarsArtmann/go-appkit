package appkit

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/larsartmann/httputil"
	"github.com/larsartmann/httputil/httpspec"
)

func TestHTTPSpec_Conformance(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	httputil.RegisterHealth(mux)

	logger := slog.New(slog.NewTextHandler(&devNullWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
	mws := defaultMiddlewareStack(logger, DefaultServiceConfig())
	wrapped := httputil.Chain(mux, mws...)

	httpspec.Run(t, wrapped,
		httpspec.WithIndexPath("/health"),
	)
}

type devNullWriter struct{}

func (devNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
