package appkit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHealthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	DefaultHealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if body := rec.Body.String(); body != `{"status":"ok"}`+"\n" {
		t.Errorf("body = %q, want %q", body, `{"status":"ok"}`+"\n")
	}
}

func TestNewHealthHandler_Ready(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewHealthHandler(HealthStatusReady)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if body := rec.Body.String(); body != `{"status":"ready"}`+"\n" {
		t.Errorf("body = %q, want %q", body, `{"status":"ready"}`+"\n")
	}
}

func TestNewHealthHandler_Unhealthy(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewHealthHandler(HealthStatusUnhealthy)(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestNewHealthHandler_Degraded(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewHealthHandler(HealthStatusDegraded)(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestHealthStatus_HTTPStatus_Unknown(t *testing.T) {
	t.Parallel()

	status := HealthStatus("custom")
	if got := status.HTTPStatus(); got != http.StatusOK {
		t.Errorf("unknown status should default to 200, got %d", got)
	}
}
