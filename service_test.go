package appkit

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewService_Defaults(t *testing.T) {
	t.Parallel()

	svc, err := NewService(DefaultServiceConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = svc.Close() }()

	if svc.Logger == nil {
		t.Fatal("expected non-nil Logger")
	}

	if svc.Mux == nil {
		t.Fatal("expected non-nil Mux")
	}

	if svc.cfg.Addr != defaultAddr {
		t.Errorf("Addr = %q, want %q", svc.cfg.Addr, defaultAddr)
	}

	if svc.cfg.ShutdownTimeout != defaultShutdownTimeout {
		t.Errorf("ShutdownTimeout = %v, want %v", svc.cfg.ShutdownTimeout, defaultShutdownTimeout)
	}
}

func TestNewService_CustomConfig(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:     "localhost:0",
		LogLevel: LogLevelDebug,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = svc.Close() }()

	if svc.cfg.Addr != "localhost:0" {
		t.Errorf("Addr = %q, want %q", svc.cfg.Addr, "localhost:0")
	}

	if svc.cfg.LogLevel != LogLevelDebug {
		t.Errorf("LogLevel = %q, want %q", svc.cfg.LogLevel, LogLevelDebug)
	}
}

func TestNewService_RegisterHealthFalse(t *testing.T) {
	t.Parallel()

	disabled := false

	svc, err := NewService(ServiceConfig{
		Addr:           "localhost:0",
		RegisterHealth: &disabled,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/health")

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("health should not be registered, got status %d", resp.StatusCode)
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestNewService_Validation_NegativeReadTimeout(t *testing.T) {
	t.Parallel()

	_, err := NewService(ServiceConfig{
		Addr:        "localhost:0",
		ReadTimeout: -2 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error for negative ReadTimeout")
	}
}

func TestNewService_Validation_NegativeWriteTimeout(t *testing.T) {
	t.Parallel()

	_, err := NewService(ServiceConfig{
		Addr:         "localhost:0",
		WriteTimeout: -2 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error for negative WriteTimeout")
	}
}

func TestNewService_Validation_NegativeShutdownTimeout(t *testing.T) {
	t.Parallel()

	_, err := NewService(ServiceConfig{
		Addr:            "localhost:0",
		ShutdownTimeout: -1,
	})
	if err == nil {
		t.Fatal("expected error for negative ShutdownTimeout")
	}
}

func TestService_Start_Addr_Running(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{Addr: "localhost:0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svc.Addr() != nil {
		t.Fatal("expected nil Addr before Start")
	}

	if svc.Running() {
		t.Fatal("expected Running() to be false before Start")
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	if svc.Addr() == nil {
		t.Fatal("expected non-nil Addr after Start")
	}

	if !svc.Running() {
		t.Fatal("expected Running() to be true after Start")
	}

	_ = svc.Shutdown(context.Background())

	assertServerStopped(t, errCh)

	if svc.Running() {
		t.Fatal("expected Running() to be false after Shutdown")
	}
}

func TestService_HealthEndpoints(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{Addr: "localhost:0", DrainDelay: NoDrainDelay})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	base := "http://" + svc.Addr().String()

	for _, path := range []string{"/health", "/health/live", "/health/ready"} {
		resp := httpGet(t, t.Context(), base+path)

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status = %d, want %d", path, resp.StatusCode, http.StatusOK)
		}

		if !strings.Contains(string(body), `"up"`) {
			t.Errorf("%s: body should contain status up, got %s", path, body)
		}
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}

func TestService_Shutdown_DrainSequence(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		DrainDelay: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	if !svc.readyProbe.Load() {
		t.Fatal("ready probe should be true before shutdown")
	}

	start := time.Now()

	_ = svc.Shutdown(context.Background())

	assertServerStopped(t, errCh)

	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Errorf("shutdown should wait at least DrainDelay, took %v", elapsed)
	}

	if svc.readyProbe.Load() {
		t.Fatal("ready probe should be false after shutdown")
	}
}

func TestService_Close_Idempotent(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{Addr: "localhost:0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	err = svc.Close()
	if err != nil {
		t.Fatalf("first Close: %v", err)
	}

	err = svc.Close()
	if err != nil {
		t.Fatalf("second Close: %v", err)
	}

	assertServerStopped(t, errCh)
}

func TestService_CustomRoute(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{Addr: "localhost:0", DrainDelay: NoDrainDelay})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("world"))
	})

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	resp := httpGet(t, t.Context(), "http://"+svc.Addr().String()+"/hello")

	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "world" {
		t.Errorf("body = %q, want %q", body, "world")
	}

	_ = svc.Close()

	assertServerStopped(t, errCh)
}
