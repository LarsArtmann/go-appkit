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
