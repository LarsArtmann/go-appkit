package appkit

import (
	"net/http"
	"testing"
)

func TestDefaultHealthHandler(t *testing.T) {
	t.Parallel()

	rec := httptestNewRecorder()
	DefaultHealthHandler(rec, newHealthRequest(t))

	assertStatus(t, rec, http.StatusOK)
	assertHealthBody(t, rec, `{"status":"ok"}`)
}

func TestNewHealthHandler_Ready(t *testing.T) {
	t.Parallel()

	rec := httptestNewRecorder()
	NewHealthHandler(HealthStatusReady)(rec, newHealthRequest(t))

	assertStatus(t, rec, http.StatusOK)
	assertHealthBody(t, rec, `{"status":"ready"}`)
}

func TestNewHealthHandler_Unhealthy(t *testing.T) {
	t.Parallel()

	rec := httptestNewRecorder()
	NewHealthHandler(HealthStatusUnhealthy)(rec, newHealthRequest(t))

	assertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestNewHealthHandler_Degraded(t *testing.T) {
	t.Parallel()

	rec := httptestNewRecorder()
	NewHealthHandler(HealthStatusDegraded)(rec, newHealthRequest(t))

	assertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHealthStatus_HTTPStatus_Unknown(t *testing.T) {
	t.Parallel()

	status := HealthStatus("custom")
	if got := status.HTTPStatus(); got != http.StatusOK {
		t.Errorf("unknown status should default to 200, got %d", got)
	}
}
