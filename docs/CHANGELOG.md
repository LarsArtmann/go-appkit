## [0.3.2] - 2026-10-06

### Changed

- Bumped templ-components family v1.19.4 → v1.20.1 (root/htmx/icons/utils — heals the v1.20.0 zero-pseudo go.mod) and catalog v4.6.1 → v4.7.1 (the templ-components v1.20.1 pin carrier; v4.6.1's generated code does not compile against v1.20.1). Dependency-only.

# Changelog

## [0.3.1] - 2026-09-29

### Changed

- Go directive 1.26.7 → 1.27.1 (repo-wide toolchain unification).
  `templ-components` v1.17.0 → v1.19.4 and `go-error-family` v0.10.1 →
  v0.11.0 (indirect; BuildFlow dependency sweep).

### Planned

- Consider `retract v0.2.0` so a stale proxy can never resolve the ghost
  tag; the directive only takes effect once CONSUMERS fetch a version
  carrying it, so it rides the NEXT docs tag. Deferred 2026-09-17: `@latest`
  already resolves v0.3.0 (highest semver) and an explicit `@v0.2.0` fetch
  fails loudly at the proxy — a release train just for the directive is not
  worth it today.

## [0.3.0] - 2026-09-16

### Fixed

- **Un-ghosted the published module.** The `docs/v0.2.0` tag was UNFETCHABLE
  from the module proxy: the module path declared `.../go-appkit/docs` while
  the directory was `docs-mod/`, so the proxy found no `docs/go.mod` at the
  tag and every consumer `go get` failed. Fixed by repathing the directory to
  `docs/` (project documentation moved to `doc/`) so the module path and
  directory agree; `docs/v0.3.0` is the first fetchable docs release. No API
  changes — identical module path and source, only the in-repo directory
  moved.

### Added

- `.cqrs-lint.json` module preset (`library`; disables A018/A009 as
  docs-by-design false positives — the module uses catalog/v4 for doc
  generation, not event stores). Tooling config only, no API change.

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
