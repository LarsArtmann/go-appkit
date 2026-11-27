package preset_test

import (
	"encoding/json/v2"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/preset"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil/endpoints"
)

const (
	endpointsBasicAuthUser = "ops"
	endpointsBasicAuthPass = "listing-secret"
)

// TestEndpoints_NotMountedByDefault: no Endpoints config, no listing — the
// route table is opt-in, like every other surface that documents the service.
func TestEndpoints_NotMountedByDefault(t *testing.T) {
	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	status, _ := getURL(t, servePreset(t, svc)+"/endpoints")

	if status != http.StatusNotFound {
		t.Errorf("GET /endpoints without Endpoints config = %d, want 404", status)
	}
}

// TestEndpoints_ListingCarriesCoreRealtimeAndConsumerRoutes: the listing
// passes through verbatim and carries the realtime stream plus every route
// the consumer registers through the Service.
//
// Known gap, pinned here on purpose: the health surface and the error-pages
// catch-all register on the mux through their own modules, so they are
// served but absent from the listing (see Config.Endpoints).
func TestEndpoints_ListingCarriesCoreRealtimeAndConsumerRoutes(t *testing.T) {
	cfg := testConfig()
	cfg.Endpoints = &appkit.EndpointsConfig{AllowUnauthenticated: true}
	cfg.Realtime = &preset.RealtimeConfig{Enabled: true}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	err = svc.HandleFunc("GET /things", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "List all things")
	if err != nil {
		t.Fatalf("register GET /things: %v", err)
	}

	err = svc.Handle("POST /things/{id}", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), "Update one thing")
	if err != nil {
		t.Fatalf("register POST /things/{id}: %v", err)
	}

	status, body := getURL(t, servePreset(t, svc)+"/endpoints")

	if status != http.StatusOK {
		t.Fatalf("GET /endpoints = %d, want 200; body: %s", status, body)
	}

	var listing endpoints.Response

	err = json.Unmarshal([]byte(strings.TrimSpace(body)), &listing)
	if err != nil {
		t.Fatalf("decode listing: %v; body: %s", err, body)
	}

	listed := make([]string, 0, len(listing.Endpoints))
	for _, route := range listing.Endpoints {
		listed = append(listed, route.Pattern)
	}

	slices.Sort(listed)

	want := []string{"GET /endpoints", "GET /events", "GET /things", "POST /things/{id}"}
	if !slices.Equal(listed, want) {
		t.Errorf("listing = %v, want %v", listed, want)
	}
}

// TestEndpoints_BasicAuthPassesThrough: the preset never invents
// credentials, and the ones the consumer sets guard the listing.
func TestEndpoints_BasicAuthPassesThrough(t *testing.T) {
	cfg := testConfig()
	cfg.Endpoints = &appkit.EndpointsConfig{
		BasicAuthUser: endpointsBasicAuthUser,
		BasicAuthPass: endpointsBasicAuthPass,
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	baseURL := servePreset(t, svc)

	status, _ := getURL(t, baseURL+"/endpoints")

	if status != http.StatusUnauthorized {
		t.Errorf("GET /endpoints without credentials = %d, want 401", status)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/endpoints", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.SetBasicAuth(endpointsBasicAuthUser, endpointsBasicAuthPass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /endpoints with credentials: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /endpoints with credentials = %d, want 200", resp.StatusCode)
	}
}

// TestEndpoints_RejectsUnauthenticatedConfig: an Endpoints config without
// protection fails the boot before any listener binds, with core's code.
func TestEndpoints_RejectsUnauthenticatedConfig(t *testing.T) {
	cfg := testConfig()
	cfg.Endpoints = &appkit.EndpointsConfig{}

	_, err := preset.New(cfg)
	if err == nil {
		t.Fatal("an unauthenticated listing must fail the boot")
	}

	if code := errorfamily.Code(err); code != "appkit.endpoints_config_invalid" {
		t.Errorf("error code = %q, want appkit.endpoints_config_invalid; err: %v", code, err)
	}
}
