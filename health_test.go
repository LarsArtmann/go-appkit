package appkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_ReturnsUp(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	RegisterHealth(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, newHealthRequest(t))

	assertStatus(t, rec, http.StatusOK)

	var resp map[string]string

	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if resp["status"] != "up" {
		t.Errorf("status = %q, want %q", resp["status"], "up")
	}
}

func TestReadyHandlerWithProbe_Ready(t *testing.T) {
	t.Parallel()

	handler := ReadyHandlerWithProbe(func() bool { return true })

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	assertStatus(t, rec, http.StatusOK)
}

func TestReadyHandlerWithProbe_NotReady(t *testing.T) {
	t.Parallel()

	handler := ReadyHandlerWithProbe(func() bool { return false })

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	assertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestRegisterHealth_RegistersThreeRoutes(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	RegisterHealth(mux)

	for _, path := range []string{"/health", "/health/live", "/health/ready"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}
}

func TestReadyEndpoint_ComposesDrainProbeWithReadyCheck(t *testing.T) {
	t.Parallel()

	externalReady := false

	svc, err := NewService(ServiceConfig{
		DrainDelay: 0,
		ReadyCheck: func() bool { return externalReady },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	get := func() int {
		rec := httptest.NewRecorder()
		svc.Mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

		return rec.Code
	}

	if code := get(); code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while external check false, got %d", code)
	}

	externalReady = true

	if code := get(); code != http.StatusOK {
		t.Fatalf("expected 200 after external check flips true, got %d", code)
	}

	// Drain must still force 503 even when the external check is happy:
	// Shutdown flips the probe before unbinding.
	svc.readyProbe.Store(false)

	if code := get(); code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 after drain probe flipped, got %d", code)
	}
}

func TestReadyEndpoint_ProbeOnlyWhenReadyCheckNil(t *testing.T) {
	t.Parallel()

	svc, err := NewService(ServiceConfig{DrainDelay: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	svc.Mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	assertStatus(t, rec, http.StatusOK)
}
