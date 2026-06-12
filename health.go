package appkit

import (
	"encoding/json"
	"net/http"
)

// DefaultHealthHandler writes {"status":"ok"}.
func DefaultHealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// NewHealthHandler returns a handler that reports the given status.
func NewHealthHandler(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}
