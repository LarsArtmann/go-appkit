package appkit

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer_RegistersHealthEndpoint(t *testing.T) {
	t.Parallel()

	port := freePort(t)

	mux := http.NewServeMux()
	cfg := ServerConfig{Port: port, RegisterHealth: true}
	srv := NewServer(cfg, mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start(ctx)
	}()

	addr := waitForAddr(t, srv)

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
	cfg.HealthHandler = NewHealthHandler(HealthStatusUnhealthy)

	srv := NewServer(cfg, mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestNewServer_HealthOptOut(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	cfg := DefaultServerConfig()
	cfg.RegisterHealth = false

	NewServer(cfg, mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected /health to not be registered, got status %d", rec.Code)
	}
}

func TestServer_AddrNilBeforeStart(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := NewServer(DefaultServerConfig(), mux)

	if srv.Addr() != nil {
		t.Error("expected nil addr before Start")
	}

	if srv.Running() {
		t.Error("expected Running() to be false before Start")
	}
}

func TestServer_RunningAfterStart(t *testing.T) {
	t.Parallel()

	port := freePort(t)

	mux := http.NewServeMux()
	srv := NewServer(ServerConfig{Port: port, RegisterHealth: false}, mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start(ctx)
	}()

	waitForRunning(t, srv)

	cancel()

	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestServer_Shutdown(t *testing.T) {
	t.Parallel()

	port := freePort(t)

	mux := http.NewServeMux()
	srv := NewServer(ServerConfig{Port: port, RegisterHealth: false}, mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start(ctx)
	}()

	waitForRunning(t, srv)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after shutdown")
	}
}

func TestServer_ShutdownBeforeStart(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := NewServer(DefaultServerConfig(), mux)

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown on unstarted server should be nil, got: %v", err)
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("get free port: %v", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	return port
}

func waitForAddr(t *testing.T, srv *Server) net.Addr {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if addr := srv.Addr(); addr != nil {
			return addr
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("server address is nil after polling")

	return nil
}

func waitForRunning(t *testing.T, srv *Server) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if srv.Running() {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("server did not start within deadline")
}
