# Changelog

## [Unreleased]

## [0.1.2] - 2026-10-13

### Changed

- Floor raised `go 1.27.1` → `go 1.27.2` (repo-wide toolchain train).
- Dependency sweep: templ-components v1.20.1 → v1.21.0
  (errorpage/icons/utils), go-error-family v0.10.x → v0.11.0, httputil
  v1.4.1 → v1.5.0 (indirect), a-h/templ v0.3.1070 + misc indirect bumps;
  the example now consumes go-appkit v0.9.0. No API changes.
- Example honors the `PORT` environment variable (8080 stays occupied on
  dev machines — the module-wide convention).

### Decided against

- The long-open TODO "swap hand-rolled `statusRecorder` for
  `httputil.ResponseRecorder`" was evaluated and declined at the source
  level: httputil's recorder (v1.5.0) is a PASS-THROUGH capture (every
  Write/WriteHeader delegates to an underlying writer), while `Wrap` needs
  a body-DISCARDING throwaway to discern 404 vs 405 without leaking the
  mux's bare text to the client. Wrapping a discard writer inside
  httputil's recorder would rename the hand-roll, not remove it. The
  12-line `statusRecorder` stays, with this record closing the ask.

## [0.1.1] - 2026-10-06

### Changed

- Bumped templ-components family v1.19.4 → v1.20.1 (errorpage/icons/utils). Lint cosmetics since v0.1.0 ride along (exhaustruct_v5 nolint on the zero-status recorder, override shadow rename). No API changes.

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
