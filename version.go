package appkit

import (
	"encoding/json"
	"net/http"
)

// versionHandler serves the configured build version as JSON at GET
// /version. Registered by NewService when ServiceConfig.Version is set.
// The /health payload shape is owned by httputil upstream, so the version
// lives on its own endpoint rather than being smuggled into /health.
func versionHandler(version string) http.HandlerFunc {
	type versionPayload struct {
		Version string `json:"version"`
	}

	payload, err := json.Marshal(versionPayload{Version: version})
	if err != nil { //nolint:noinlineerr // unreachable for a plain string field
		payload = []byte(`{"version":"unknown"}`)
	}

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}
}
