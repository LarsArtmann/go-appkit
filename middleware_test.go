package appkit

import (
	"net/http"
	"testing"

	"github.com/larsartmann/httputil"
)

func TestMiddleware_PanicRecovery_Returns500(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.Mux.HandleFunc("GET /panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/panic")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("panic should return 500, got %d", resp.StatusCode)
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestMiddleware_RequestID_HeaderPresent(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.Mux.HandleFunc("GET /test-reqid", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/test-reqid")
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("X-Request-Id") == "" {
		t.Error("expected X-Request-Id header to be set")
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestMiddleware_SecurityHeaders_Present(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.Mux.HandleFunc("GET /test-sec", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/test-sec")
	defer func() { _ = resp.Body.Close() }()

	tests := []struct {
		header string
		want   string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		got := resp.Header.Get(tt.header)
		if got != tt.want {
			t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
		}
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestMiddleware_ReplacedStack(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 0,
		Middlewares: []httputil.Middleware{
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.Mux.HandleFunc("GET /test-custom", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/test-custom")
	defer func() { _ = resp.Body.Close() }()

	// With replaced stack, default RequestID middleware is gone — no X-Request-Id header.
	if resp.Header.Get("X-Request-Id") != "" {
		t.Error("replaced stack should not include default RequestID middleware")
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}
