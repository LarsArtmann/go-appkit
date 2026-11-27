package appkit_test

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/testkit"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil/endpoints"
)

const (
	endpointsBasicAuthUser = "ops"
	endpointsBasicAuthPass = "listing-secret"
	metricsBasicAuthUser   = "metrics"
	metricsBasicAuthPass   = "secret"
	testVersion            = "1.2.3-test"
)

// newEndpointsService builds a service with every framework route enabled
// (health family, metrics, version) and an open endpoint listing, then lets
// mutate adjust the configuration.
func newEndpointsService(t *testing.T, mutate func(*appkit.ServiceConfig)) *appkit.Service {
	t.Helper()

	cfg := appkit.DefaultServiceConfig()
	cfg.Version = testVersion
	cfg.Metrics = &appkit.MetricsConfig{
		BasicAuthUser: metricsBasicAuthUser,
		BasicAuthPass: metricsBasicAuthPass,
	}
	cfg.Endpoints = &appkit.EndpointsConfig{AllowUnauthenticated: true}

	if mutate != nil {
		mutate(&cfg)
	}

	return testkit.NewService(t, cfg)
}

// registerConsumerRoutes mounts two consumer routes through the service
// seam, the way a real service does after NewService returns.
func registerConsumerRoutes(t *testing.T, svc *appkit.Service) {
	t.Helper()

	err := svc.HandleFunc("GET /gadgets", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "List all gadgets")
	if err != nil {
		t.Fatalf("register GET /gadgets: %v", err)
	}

	err = svc.Handle("POST /gadgets/{id}", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), "Update one gadget")
	if err != nil {
		t.Fatalf("register POST /gadgets/{id}: %v", err)
	}
}

// getListing requests path on the running server and returns the drained
// response. configure may set headers on the request.
func getListing(
	t *testing.T,
	server *testkit.TestServer,
	path string,
	configure func(*http.Request),
) (int, http.Header, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.FullChainURL+path, nil)
	if err != nil {
		t.Fatalf("build GET %s request: %v", path, err)
	}

	if configure != nil {
		configure(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read GET %s body: %v", path, err)
	}

	return resp.StatusCode, resp.Header, string(body)
}

// decodeListing returns the patterns the listing JSON carries, sorted.
func decodeListing(t *testing.T, body string) []string {
	t.Helper()

	var listing endpoints.Response

	err := json.Unmarshal([]byte(strings.TrimSpace(body)), &listing)
	if err != nil {
		t.Fatalf("decode listing: %v; body: %s", err, body)
	}

	patterns := make([]string, 0, len(listing.Endpoints))
	for _, route := range listing.Endpoints {
		patterns = append(patterns, route.Pattern)
	}

	slices.Sort(patterns)

	return patterns
}

func TestEndpoints_NotMountedByDefault(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.DefaultServiceConfig())
	server := testkit.Serve(t, svc)

	status, _, body := getListing(t, server, "/endpoints", nil)

	if status != http.StatusNotFound {
		t.Errorf("GET /endpoints without Endpoints config: status = %d, want 404; body: %s", status, body)
	}
}

func TestEndpoints_RegistryRecordsRoutesWithoutListing(t *testing.T) {
	t.Parallel()

	// The registry is always there: a service that enables the listing
	// later (or inspects Routes in tests) still sees every route.
	svc := testkit.NewService(t, appkit.DefaultServiceConfig())
	registerConsumerRoutes(t, svc)

	patterns := routePatterns(svc.Routes.Routes())

	for _, want := range []string{"GET /health", "GET /gadgets", "POST /gadgets/{id}"} {
		if !slices.Contains(patterns, want) {
			t.Errorf("route %q is missing from the registry; recorded: %v", want, patterns)
		}
	}
}

// routePatterns projects registry rows onto their ServeMux patterns.
func routePatterns(routes []endpoints.Route) []string {
	patterns := make([]string, 0, len(routes))
	for _, route := range routes {
		patterns = append(patterns, route.Pattern)
	}

	return patterns
}

func TestEndpoints_ListingCoversFrameworkAndConsumerRoutes(t *testing.T) {
	t.Parallel()

	svc := newEndpointsService(t, nil)
	registerConsumerRoutes(t, svc)
	server := testkit.Serve(t, svc)

	status, header, body := getListing(t, server, "/endpoints", nil)

	if status != http.StatusOK {
		t.Fatalf("GET /endpoints: status = %d, want 200; body: %s", status, body)
	}

	if contentType := header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	if vary := header.Get("Vary"); vary != "Accept" {
		t.Errorf("Vary = %q, want Accept", vary)
	}

	if cacheControl := header.Get("Cache-Control"); cacheControl != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cacheControl)
	}

	want := []string{
		"GET /endpoints",
		"GET /gadgets",
		"GET /health",
		"GET /health/live",
		"GET /health/ready",
		"GET /metrics",
		"GET /version",
		"POST /gadgets/{id}",
	}

	got := decodeListing(t, body)

	if !slices.Equal(got, want) {
		t.Errorf("listing = %v, want %v", got, want)
	}

	var listing endpoints.Response

	err := json.Unmarshal([]byte(strings.TrimSpace(body)), &listing)
	if err != nil {
		t.Fatalf("decode listing: %v", err)
	}

	for _, route := range listing.Endpoints {
		if route.Summary == "" {
			t.Errorf("route %q has no summary", route.Pattern)
		}
	}
}

func TestEndpoints_RoutesOnTheRawMuxAreNotListed(t *testing.T) {
	t.Parallel()

	// The documented caveat: Service.Mux stays public for libraries that
	// mount their own route table, and such routes are served but absent
	// from the listing. Register through the service to be listed.
	svc := newEndpointsService(t, nil)
	svc.Mux.HandleFunc("GET /raw", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := testkit.Serve(t, svc)

	_, _, body := getListing(t, server, "/endpoints", nil)

	if slices.Contains(decodeListing(t, body), "GET /raw") {
		t.Error("a route registered directly on Service.Mux must not appear in the listing")
	}
}

func TestEndpoints_NegotiatesHTML(t *testing.T) {
	t.Parallel()

	svc := newEndpointsService(t, nil)
	server := testkit.Serve(t, svc)

	status, header, body := getListing(t, server, "/endpoints", func(req *http.Request) {
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
	})

	if status != http.StatusOK {
		t.Fatalf("GET /endpoints (html): status = %d, want 200", status)
	}

	if contentType := header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}

	for _, want := range []string{"<table", `<th scope="col">Pattern</th>`, "GET /health"} {
		if !strings.Contains(body, want) {
			t.Errorf("HTML listing lacks %q; body: %.400s", want, body)
		}
	}
}

func TestEndpoints_BasicAuthEnforced(t *testing.T) {
	t.Parallel()

	svc := newEndpointsService(t, func(cfg *appkit.ServiceConfig) {
		cfg.Endpoints = &appkit.EndpointsConfig{
			BasicAuthUser: endpointsBasicAuthUser,
			BasicAuthPass: endpointsBasicAuthPass,
		}
	})
	server := testkit.Serve(t, svc)

	status, header, _ := getListing(t, server, "/endpoints", nil)

	if status != http.StatusUnauthorized {
		t.Errorf("without credentials: status = %d, want 401", status)
	}

	if want := `Basic realm="endpoints"`; header.Get("WWW-Authenticate") != want {
		t.Errorf("WWW-Authenticate = %q, want %q", header.Get("WWW-Authenticate"), want)
	}

	status, _, body := getListing(t, server, "/endpoints", func(req *http.Request) {
		req.SetBasicAuth(endpointsBasicAuthUser, endpointsBasicAuthPass)
	})

	if status != http.StatusOK {
		t.Fatalf("with credentials: status = %d, want 200; body: %s", status, body)
	}
}

func TestEndpoints_GuardProtectsListing(t *testing.T) {
	t.Parallel()

	// A service with its own auth middleware hands it to the listing; the
	// built-in Basic Auth is not required then.
	svc := newEndpointsService(t, func(cfg *appkit.ServiceConfig) {
		cfg.Endpoints = &appkit.EndpointsConfig{
			Guard: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("X-Admin") != "yes" {
						w.WriteHeader(http.StatusForbidden)

						return
					}

					next.ServeHTTP(w, r)
				})
			},
		}
	})
	server := testkit.Serve(t, svc)

	status, _, _ := getListing(t, server, "/endpoints", nil)

	if status != http.StatusForbidden {
		t.Errorf("without the guard's credential: status = %d, want 403", status)
	}

	status, _, _ = getListing(t, server, "/endpoints", func(req *http.Request) {
		req.Header.Set("X-Admin", "yes")
	})

	if status != http.StatusOK {
		t.Errorf("with the guard's credential: status = %d, want 200", status)
	}
}

func TestEndpoints_CustomPath(t *testing.T) {
	t.Parallel()

	svc := newEndpointsService(t, func(cfg *appkit.ServiceConfig) {
		cfg.Endpoints.Path = "GET /routes"
	})
	server := testkit.Serve(t, svc)

	status, _, body := getListing(t, server, "/routes", nil)

	if status != http.StatusOK {
		t.Fatalf("GET /routes: status = %d, want 200; body: %s", status, body)
	}

	if !slices.Contains(decodeListing(t, body), "GET /routes") {
		t.Error("the listing must list itself at its custom path")
	}

	status, _, _ = getListing(t, server, "/endpoints", nil)

	if status != http.StatusNotFound {
		t.Errorf("GET /endpoints with a custom listing path: status = %d, want 404", status)
	}
}

func TestEndpoints_RejectsUnauthenticatedConstruction(t *testing.T) {
	t.Parallel()

	cfg := appkit.DefaultServiceConfig()
	cfg.Endpoints = &appkit.EndpointsConfig{}

	_, err := appkit.NewService(cfg)
	if err == nil {
		t.Fatal("an unauthenticated listing must be rejected at construction")
	}

	if code := errorfamily.Code(err); code != "appkit.endpoints_config_invalid" {
		t.Errorf("error code = %q, want appkit.endpoints_config_invalid; err: %v", code, err)
	}
}

func TestEndpoints_RejectsInvalidPathAtConstruction(t *testing.T) {
	t.Parallel()

	// Before the registry seam this pattern panicked inside http.ServeMux
	// during NewService; it is a construction error now.
	cfg := appkit.DefaultServiceConfig()
	cfg.Endpoints = &appkit.EndpointsConfig{
		Path:                 "not-a-serve-mux-pattern",
		AllowUnauthenticated: true,
	}

	_, err := appkit.NewService(cfg)
	if err == nil {
		t.Fatal("an invalid listing pattern must be rejected at construction")
	}

	if !strings.Contains(err.Error(), "not-a-serve-mux-pattern") {
		t.Errorf("error must name the rejected pattern; err: %v", err)
	}
}

func TestService_HandleRejectsDuplicateRoute(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.DefaultServiceConfig())

	err := svc.HandleFunc("GET /health", func(http.ResponseWriter, *http.Request) {},
		"a duplicate of the framework route")
	if err == nil {
		t.Fatal("registering a pattern twice must be rejected")
	}

	if code := errorfamily.Code(err); code != "appkit.route_rejected" {
		t.Errorf("error code = %q, want appkit.route_rejected; err: %v", code, err)
	}

	if !strings.Contains(err.Error(), `"GET /health"`) {
		t.Errorf("error must name the rejected route; err: %v", err)
	}

	// The rejected route left nothing behind: the framework's own row is
	// the only one for that pattern.
	count := 0

	for _, route := range svc.Routes.Routes() {
		if route.Pattern == "GET /health" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("GET /health is recorded %d times, want 1", count)
	}
}

func TestService_HandleRejectsEmptyPattern(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.DefaultServiceConfig())

	err := svc.HandleFunc("", func(http.ResponseWriter, *http.Request) {}, "")
	if err == nil {
		t.Fatal("an empty pattern must be rejected")
	}

	if code := errorfamily.Code(err); code != "appkit.route_rejected" {
		t.Errorf("error code = %q, want appkit.route_rejected; err: %v", code, err)
	}
}

// rawMuxRegistration matches a direct registration on a ServeMux value
// (mux.Handle, svc.Mux.HandleFunc, ...). The core module owns exactly one
// place that may talk to the service mux: the registry seam in routes.go.
var rawMuxRegistration = regexp.MustCompile(`\b(?:[A-Za-z_]\w*\.)*(?:m|mux|\w*[mM]ux)\.(?:Handle|HandleFunc)\(`)

// TestCoreModule_RegistersOnlyThroughTheRegistry pins the property the
// listing relies on: no framework route bypasses [appkit.Service.Handle].
func TestCoreModule_RegistersOnlyThroughTheRegistry(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		for lineNo, line := range strings.Split(string(source), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}

			if rawMuxRegistration.MatchString(trimmed) {
				t.Errorf(
					"%s:%d registers directly on the mux (%s) — framework routes go through Service.Handle/HandleFunc",
					name,
					lineNo+1,
					trimmed,
				)
			}
		}
	}
}
