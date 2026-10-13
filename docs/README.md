# docs

Opt-in auto-documentation module for appkit services: catalogs routes,
config, and service metadata via catalog/v4 and serves a documentation
surface. Package docs live in `doc.go`; the module CHANGELOG records the
release history (including the v0.2.0 ghost-tag incident and the v0.3.0
directory repath).

End-to-end composition coverage for this module is mapped in the
integration module (`../integration/README.md`); the docs seam itself is
pinned end-to-end by `../integration/docs_composition_test.go`.

## Gotcha: spec titles come from `Config.ServiceName`

The OpenAPI/AsyncAPI `info.title` that `docserver` emits reads
`docserver.Config.ServiceName` — NOT the CatalogBuilder title. Only
`catalog.json` carries the builder title. Also: a catalog with ZERO
registered commands/events emits empty `info` sections — register at least
one command/event for real generated-schema coverage. Pinned end-to-end by
`integration/docs_composition_test.go` through the full appkit default
chain (caught 2026-10-12).

## Build & verify

```bash
cd docs
GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.2 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.2 go build ./...
golangci-lint run ./...        # from this directory
```

`GOTOOLCHAIN=go1.27.2` is required (catalog/v4 uses
`encoding/json/v2`); `GOWORK=off` makes the run hermetic.
