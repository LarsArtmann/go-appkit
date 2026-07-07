package docs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/catalog/v3/docserver"
)

func TestRegisterDocs_ServesOpenAPIJSON(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	cb := NewCatalogBuilder("Test Service", "1.0.0")

	RegisterDocs(mux, cb, docserver.Config{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("openapi.json: status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Content-Type") == "" {
		t.Error("openapi.json: expected Content-Type header")
	}
}

func TestRegisterDocs_ServesAsyncAPIJSON(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	cb := NewCatalogBuilder("Test Service", "1.0.0")

	RegisterDocs(mux, cb, docserver.Config{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs/asyncapi.json", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("asyncapi.json: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNewCatalogBuilder_ReturnsBuilder(t *testing.T) {
	t.Parallel()

	cb := NewCatalogBuilder("My Service", "2.0.0")
	if cb.Builder() == nil {
		t.Fatal("expected non-nil underlying Builder")
	}
}
