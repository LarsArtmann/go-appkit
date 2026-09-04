package appkit_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	charm "github.com/charmbracelet/log"
	"github.com/larsartmann/httputil"
)

// These benchmarks isolate the per-request cost of the httputil Logging
// middleware that appkit's default stack includes — the same delta the
// cqrs-htmx comparison measured end-to-end (finding 7, ~2.8x). Output goes
// to io.Discard in every variant: what is measured is formatting + emission
// work, never terminal I/O.

func benchmarkServer(b *testing.B, mw func(http.Handler) http.Handler) *httptest.Server {
	b.Helper()

	server := httptest.NewServer(mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	b.Cleanup(server.Close)

	return server
}

func requestOnce(b *testing.B, server *httptest.Server) {
	b.Helper()

	req, err := http.NewRequestWithContext(b.Context(), http.MethodGet, server.URL, nil)
	if err != nil {
		b.Fatalf("build request: %v", err)
	}

	res, err := server.Client().Do(req)
	if err != nil {
		b.Fatalf("request failed: %v", err)
	}

	_ = res.Body.Close()
}

func BenchmarkRequest_LoggingInfo(b *testing.B) {
	logger := charm.New(io.Discard)
	logger.SetLevel(charm.InfoLevel)
	server := benchmarkServer(b, httputil.Logging(slog.New(logger)))

	b.ResetTimer()

	for b.Loop() {
		requestOnce(b, server)
	}
}

func BenchmarkRequest_LoggingSuppressed(b *testing.B) {
	logger := charm.New(io.Discard)
	logger.SetLevel(charm.WarnLevel)
	server := benchmarkServer(b, httputil.Logging(slog.New(logger)))

	b.ResetTimer()

	for b.Loop() {
		requestOnce(b, server)
	}
}

func BenchmarkRequest_NoLogging(b *testing.B) {
	server := benchmarkServer(b, func(h http.Handler) http.Handler { return h })

	b.ResetTimer()

	for b.Loop() {
		requestOnce(b, server)
	}
}
