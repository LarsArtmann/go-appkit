# docs

Opt-in auto-documentation module for appkit services: catalogs routes,
config, and service metadata via catalog/v4 and serves a documentation
surface. Package docs live in `doc.go`; the module CHANGELOG records the
release history (including the v0.2.0 ghost-tag incident and the v0.3.0
directory repath).

## Build & verify

```bash
cd docs
GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.1 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.1 go build ./...
golangci-lint run ./...        # from this directory
```

`GOTOOLCHAIN=go1.27.1` is required (catalog/v4 uses
`encoding/json/v2`); `GOWORK=off` makes the run hermetic.
