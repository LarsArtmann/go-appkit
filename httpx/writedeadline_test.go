package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func slowHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(delay)

		_, _ = w.Write([]byte("late but alive"))
	})
}

// The extension works end to end: a write that outlives the server's
// WriteTimeout succeeds when the middleware covers the route.
func TestWriteDeadline_ExtendsDeadlineForMatchingPrefixes(t *testing.T) {
	t.Parallel()

	inner := http.NewServeMux()
	inner.Handle("POST /api/agent/", slowHandler(600*time.Millisecond))

	srv := httptest.NewUnstartedServer(WriteDeadline(5*time.Second, "/api/agent/")(inner))
	defer srv.Close()

	srv.Config.WriteTimeout = 150 * time.Millisecond
	srv.Config.ReadHeaderTimeout = 5 * time.Second

	srv.Start()

	resp, err := postProbe(t, srv.URL+"/api/agent/run")
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK || string(body) != "late but alive" {
		t.Fatalf("status = %d body = %q — the write deadline was not extended", resp.StatusCode, body)
	}
}

// The control: the same slow handler on an UNCOVERED route dies at the
// server's WriteTimeout (proves the extension is the middleware's
// doing, not a slow timeout).
func TestWriteDeadline_UncoveredRouteKeepsServerDeadline(t *testing.T) {
	t.Parallel()

	inner := http.NewServeMux()
	inner.Handle("POST /api/fast/", slowHandler(600*time.Millisecond))

	srv := httptest.NewUnstartedServer(WriteDeadline(5*time.Second, "/api/agent/")(inner))
	defer srv.Close()

	srv.Config.WriteTimeout = 150 * time.Millisecond
	srv.Config.ReadHeaderTimeout = 5 * time.Second

	srv.Start()

	resp, err := postProbe(t, srv.URL+"/api/fast/run")
	if err == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		t.Fatal("the slow uncovered write SUCCEEDED — the server WriteTimeout never applied (control broken)")
	}
}

// The ordering regression (the battery's acceptance): when the writer
// the middleware receives is opaque to http.ResponseController (the
// gzip-wrapper class — no deadline controls, no Unwrap), setting the
// deadline fails and the request must fail LOUDLY with the classified
// Infrastructure error, never silently hang.
func TestWriteDeadline_OpaqueWriterFailsLoudly(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder() // no deadline support, no Unwrap: the opaque-chain shape

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/api/agent/run", strings.NewReader("go"))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("never reached"))
	})

	WriteDeadline(time.Minute, "/api/agent/")(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 — a deadline the chain cannot set must fail loudly", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "write_deadline_unreachable") {
		t.Fatalf("body = %s, want the classified infrastructure code", rec.Body.String())
	}
}

func TestWriteDeadline_NonMatchingPrefixPassesThrough(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/other/route", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	// The recorder cannot set deadlines either — but this route does not
	// match, so the middleware must not touch the deadline at all.
	WriteDeadline(time.Minute, "/api/agent/")(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want the handler's untouched 418", rec.Code)
	}
}

// postProbe is the noctx-safe POST helper for these deadline tests.
func postProbe(t *testing.T, url string) (*http.Response, error) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, strings.NewReader("go"))
	if err != nil {
		t.Fatalf("build POST %s: %v", url, err)
	}

	return http.DefaultClient.Do(req) //nolint:wrapcheck // test helper
}
