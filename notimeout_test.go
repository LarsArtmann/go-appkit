package appkit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// NoTimeout must remove the write deadline at BOTH layers that would
// otherwise cut long-lived (SSE) responses: the http.Server field and the
// default stack's Timeout middleware.
func TestNoTimeout_DisablesServerDeadline(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:              ":0",
		ReadTimeout:       NoTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      NoTimeout,
		IdleTimeout:       60 * time.Second,
		DrainDelay:        0,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	defer func() { _ = svc.Close() }()

	if svc.server.ReadTimeout != 0 {
		t.Errorf("server.ReadTimeout = %v, want 0 (no deadline)", svc.server.ReadTimeout)
	}

	if svc.server.WriteTimeout != 0 {
		t.Errorf("server.WriteTimeout = %v, want 0 (no deadline)", svc.server.WriteTimeout)
	}

	if svc.server.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("server.ReadHeaderTimeout = %v, want 5s (reaping stays on)", svc.server.ReadHeaderTimeout)
	}

	if svc.cfg.ReadTimeout != NoTimeout || svc.cfg.WriteTimeout != NoTimeout {
		t.Error("applyDefaults overwrote the NoTimeout sentinel")
	}
}

// The default middleware stack must not attach a request-context deadline
// when WriteTimeout is NoTimeout — an SSE stream handler observes none.
func TestNoTimeout_NoRequestContextDeadline(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{Addr: ":0", WriteTimeout: NoTimeout, DrainDelay: 0})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	defer func() { _ = svc.Close() }()

	svc.Mux.HandleFunc("GET /deadline", func(w http.ResponseWriter, r *http.Request) {
		if _, hasDeadline := r.Context().Deadline(); hasDeadline {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	rec := httptestNewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/deadline", nil)
	svc.server.Handler.ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusOK)
}

// End to end over a real listener: with NoTimeout a response that outlives any
// tight deadline completes; the same handler under a 40ms WriteTimeout is cut
// off, proving the opt-out (not luck) keeps the slow response alive.
func TestNoTimeout_SlowResponseSurvives(t *testing.T) {
	t.Parallel()

	startService := func(t *testing.T, write time.Duration) *Service {
		t.Helper()

		svc, err := NewService(ServiceConfig{Addr: ":0", WriteTimeout: write, DrainDelay: 0})
		if err != nil {
			t.Fatalf("NewService: %v", err)
		}

		svc.Mux.HandleFunc("GET /slow", func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(150 * time.Millisecond)
			w.WriteHeader(http.StatusOK)

			_, _ = io.WriteString(w, "done")
		})

		if _, err := svc.Start(); err != nil { //nolint:noinlineerr // concise in closure
			t.Fatalf("start: %v", err)
		}

		waitForRunning(t, svc)

		return svc
	}

	t.Run("NoTimeout keeps slow response alive", func(t *testing.T) {
		t.Parallel()

		svc := startService(t, NoTimeout)
		defer func() { _ = svc.Close() }()

		resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/slow")

		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		if resp.StatusCode != http.StatusOK || string(body) != "done" {
			t.Fatalf("status = %d body = %q, want 200 done", resp.StatusCode, body)
		}
	})

	t.Run("tight WriteTimeout cuts the same response", func(t *testing.T) {
		t.Parallel()

		svc := startService(t, 40*time.Millisecond)
		defer func() { _ = svc.Close() }()

		client := &http.Client{Timeout: 5 * time.Second}

		req, err := http.NewRequestWithContext(
			t.Context(), http.MethodGet, "http://"+svc.Addr().String()+"/slow", nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		resp, err := client.Do(req)

		var status int
		if err != nil {
			status = 0 // connection cut: exactly what WriteTimeout does mid-response
		} else {
			defer func() { _ = resp.Body.Close() }()

			_, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				status = 0 // truncated body, also a cut
			} else {
				status = resp.StatusCode
			}
		}

		if status == http.StatusOK {
			t.Fatal("expected the 40ms WriteTimeout to cut the 150ms response")
		}
	})
}
