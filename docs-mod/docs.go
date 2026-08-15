// Package docs provides auto-documentation integration for go-appkit services.
// It wraps go-cqrs-lite/catalog/docserver to serve AsyncAPI, OpenAPI, and D2
// diagrams from Go types.
package docs

import (
	"net/http"

	"github.com/larsartmann/go-cqrs-lite/catalog/v4"
	"github.com/larsartmann/go-cqrs-lite/catalog/v4/docserver"
)

// CatalogBuilder wraps catalog.Builder with appkit-friendly defaults.
type CatalogBuilder struct {
	builder *catalog.Builder
}

// NewCatalogBuilder creates a new CatalogBuilder with the given title and version.
func NewCatalogBuilder(title, version string) *CatalogBuilder {
	return &CatalogBuilder{
		builder: catalog.NewBuilder(title, version),
	}
}

// Builder returns the underlying catalog.Builder for direct access
// (AddCommand, AddEvent, AddQuery, etc.).
func (cb *CatalogBuilder) Builder() *catalog.Builder {
	return cb.builder
}

// RegisterDocs mounts the catalog docserver routes on the given mux.
// This adds /docs/openapi, /docs/asyncapi, /docs/diagram, and /docs/catalog.json
// (paths configurable via docserver.Config).
func RegisterDocs(mux *http.ServeMux, cb *CatalogBuilder, cfg docserver.Config) {
	ds := docserver.NewDocsServer(func() *catalog.Catalog {
		return cb.builder.Build()
	}, cfg)

	ds.RegisterRoutes(mux)
}
