# Changelog

## [0.2.0] - 2026-08-15

First tagged release of the docs module. Requires `GOEXPERIMENT=jsonv2`.

### Changed — catalog v3.7.1 → v4.2.1

- Migrated to `github.com/larsartmann/go-cqrs-lite/catalog/v4` and
  `catalog/v4/docserver` (v4.2.1): v4 builder API, v4 docserver config.
- `CatalogBuilder` keeps the appkit-friendly surface: `NewCatalogBuilder(title,
  version)`, `Builder()` for direct catalog access (`AddCommand`, `AddEvent`,
  `AddQuery`, ...), and `RegisterDocs` mounting `/docs/openapi`, `/docs/asyncapi`,
  `/docs/diagram`, `/docs/catalog.json`.

## [0.1.0] - 2026-07-26

Untagged baseline that shipped inside root `v0.2.0`: catalog v3.7.1-based
`CatalogBuilder` + `RegisterDocs`.
