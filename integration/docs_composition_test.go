package integration_test

// T11 — docs-module composition E2E, against PUBLISHED tags (docs v0.3.2 +
// core v0.8.0): the auto-documentation surface (catalog docserver: OpenAPI,
// AsyncAPI, D2, catalog.json) mounted on a real appkit Service's mux and
// fetched through the FULL default middleware chain (recovery, logging,
// health) — proving the stack neither breaks docserver responses nor
// buffers them into timeouts, and that generated content (the catalog title)
// survives the roundtrip. Module-local tests cover the docserver itself;
// this leg pins the consumer composition.

import (
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	appkit "github.com/larsartmann/go-appkit"
	appkitdocs "github.com/larsartmann/go-appkit/docs"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/catalog/v4/docserver"
)

func TestDocsCompositionThroughAppkitService(t *testing.T) {
	t.Parallel()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = "localhost:0"
	cfg.DrainDelay = appkit.NoDrainDelay

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	cb := appkitdocs.NewCatalogBuilder("SUPERB Docs E2E", "v0.3.2-test")
	appkitdocs.RegisterDocs(svc.Mux, cb, docserver.Config{})

	ts := testkit.Serve(t, svc)
	baseURL := ts.FullChainURL

	for _, tt := range []struct {
		path       string
		wantInBody string
	}{
		{"/docs/openapi.json", "SUPERB Docs E2E"},
		{"/docs/asyncapi.json", "SUPERB Docs E2E"},
		{"/docs/catalog.json", "SUPERB Docs E2E"},
	} {
		resp, err := http.Get(baseURL + tt.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tt.path, err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("GET %s: read body: %v", tt.path, err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", tt.path, resp.StatusCode)
		}

		if ct := resp.Header.Get("Content-Type"); !slices.Contains([]string{"application/json", "application/json; charset=utf-8"}, ct) {
			t.Errorf("GET %s: Content-Type = %q, want application/json", tt.path, ct)
		}

		if !strings.Contains(string(body), tt.wantInBody) {
			t.Errorf("GET %s: body does not contain %q (len %d)", tt.path, tt.wantInBody, len(body))
		}
	}

	if err := ts.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
