package appkit

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer_RegistersHealthEndpoint(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := NewServer(DefaultServerConfig(), mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for listener.
	time.Sleep(50 * time.Millisecond)

	addr := srv.Addr()
	if addr == nil {
		t.Fatal("server address is nil")
	}

	resp, err := http.Get("http://" + addr.String() + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"status":"ok"}`+"\n" {
		t.Errorf("body = %q, want ok JSON", body)
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestNewServer_CustomHealthHandler(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	cfg := DefaultServerConfig()
	cfg.HealthHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	srv := NewServer(cfg, mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
