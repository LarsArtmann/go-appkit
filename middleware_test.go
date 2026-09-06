package appkit

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/larsartmann/httputil"
)

func TestMiddleware_PanicRecovery_Returns500(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
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
		DrainDelay: NoDrainDelay,
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

	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("expected X-Request-Id header to be set")
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestMiddleware_SecurityHeaders_Present(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
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
		DrainDelay: NoDrainDelay,
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
	if resp.Header.Get("X-Request-ID") != "" {
		t.Error("replaced stack should not include default RequestID middleware")
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

const (
	before = "before"
	after  = "after"
)

// markerMiddleware appends name to order before and after the inner handler
// runs and sets a response header, making middleware ordering observable.
func markerMiddleware(order *[]string, name string) httputil.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, name+":"+before)

			w.Header().Set("X-"+name, "1")

			next.ServeHTTP(w, r)

			*order = append(*order, name+":"+after)
		})
	}
}

func TestMiddleware_OuterMiddlewares_RunOutsideDefaultStack(t *testing.T) {
	t.Parallel()

	order := []string{}
	disabled := false

	svc, err := NewService(ServiceConfig{
		Addr:             "localhost:0",
		DrainDelay:       NoDrainDelay,
		OuterMiddlewares: []httputil.Middleware{markerMiddleware(&order, "Outer")},
		ExtraMiddlewares: []httputil.Middleware{markerMiddleware(&order, "Extra")},
		RegisterHealth:   &disabled,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A panicking handler proves Recovery sits INSIDE the outer middleware:
	// Recovery converts the panic to 500 and the outer middleware still
	// finishes — its "after" marker must be recorded.
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

	if resp.Header.Get("X-Outer") != "1" {
		t.Error("outer middleware should have run")
	}

	// Default stack still active underneath: RequestID header present.
	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("default RequestID middleware should still run under the outer middleware")
	}

	// The panic unwinds through the extra middleware (inside Recovery), so
	// its after-marker is lost — while the outer middleware, wrapping
	// Recovery, still finishes.
	wantOrder := []string{"Outer:" + before, "Extra:" + before, "Outer:" + after}
	if !slices.Equal(order, wantOrder) {
		t.Errorf("order = %v, want %v", order, wantOrder)
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestMiddleware_OuterMiddlewares_WrapReplacedStack(t *testing.T) {
	t.Parallel()

	order := []string{}

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: NoDrainDelay,
		Middlewares: []httputil.Middleware{
			markerMiddleware(&order, "Replacement"),
		},
		OuterMiddlewares: []httputil.Middleware{markerMiddleware(&order, "Outer")},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.Mux.HandleFunc("GET /wrapped", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/wrapped")
	defer func() { _ = resp.Body.Close() }()

	wantOrder := []string{
		"Outer:" + before,
		"Replacement:" + before,
		"Replacement:" + after,
		"Outer:" + after,
	}
	if !slices.Equal(order, wantOrder) {
		t.Errorf("order = %v, want %v", order, wantOrder)
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestBuildMiddleware_DoesNotMutateConfigSlices(t *testing.T) {
	t.Parallel()

	victim := markerMiddleware(&[]string{}, "Victim")

	// outer and tail share one backing array: index 1 holds victim and is
	// exactly where a buggy append(cfg.OuterMiddlewares, ...) would write.
	backing := []httputil.Middleware{
		markerMiddleware(&[]string{}, "First"),
		victim,
	}
	outer := backing[:1]
	tail := backing

	logger := slog.New(slog.DiscardHandler)

	chain := buildMiddleware(logger, ServiceConfig{OuterMiddlewares: outer})
	if len(chain) < len(defaultMiddlewareStack(logger, ServiceConfig{})) {
		t.Fatal("chain should contain the default stack")
	}

	rec := httptestNewRecorder()
	httputil.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), tail...).
		ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil))

	if rec.Header().Get("X-Victim") != "1" {
		t.Error("config slice backing array was written through; caller slices must stay untouched")
	}
}
