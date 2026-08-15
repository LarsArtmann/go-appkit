# Changelog

## [Unreleased]

### Added

- Nothing yet.

### Fixed

- Nothing yet.

## [0.1.0] - 2026-08-15

Initial release. Requires `GOEXPERIMENT=jsonv2`.

### Added

- `Mount(mux, cfg)` — catch-all pretty 404 (HTML, or the JSON contract when the
  client accepts `application/json`) on any `*http.ServeMux`.
- `Wrap(mux, cfg)` — same rendering for bare 404s and 405s while preserving the
  stdlib mux's `Allow` header and its path-cleaning redirects (doubled slashes,
  dot segments) for custom top-level handler setups.
- `Handler(err, cfg)` / `Write(w, r, err, cfg)` — render any error classified by
  go-error-family as a templ-components error page (HTML) or the JSON error
  contract (`{family, code, message, title, why, fix, context}`), with the same
  family → status mapping as `appkit.HTTPStatus` (Rejection 400, Conflict 409,
  Transient 503, Corruption 500, Infrastructure 503).
- `Config` — `NotFound` props, CSP `Nonce`, `Lang`, `JSONWhen` negotiation rule,
  `Override` hook; zero value ready to use.
- Render-failure safety inherited from templ-components/errorpage: if the page
  itself fails to render, a plain-text response with the correct status is
  written instead of a truncated document.
- Example app under `example/`.
