package appkit

import (
	"encoding/json"
	"net/http"
)

type HealthStatus string

const (
	HealthStatusOK        HealthStatus = "ok"
	HealthStatusReady     HealthStatus = "ready"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

func (s HealthStatus) HTTPStatus() int {
	switch s {
	case HealthStatusOK, HealthStatusReady:
		return http.StatusOK
	case HealthStatusDegraded:
		return http.StatusServiceUnavailable
	case HealthStatusUnhealthy:
		return http.StatusServiceUnavailable
	default:
		return http.StatusOK
	}
}

func DefaultHealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeHealthResponse(w, HealthStatusOK)
}

func NewHealthHandler(status HealthStatus) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeHealthResponse(w, status)
	}
}

func writeHealthResponse(w http.ResponseWriter, status HealthStatus) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.HTTPStatus())

	encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": string(status)})
	if encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
	}
}
