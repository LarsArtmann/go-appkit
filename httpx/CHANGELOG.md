# Changelog

All notable changes to this module are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the module follows
semantic versioning.

## [Unreleased]

## [0.1.0] - 2026-11-07

### Added

- Initial module (battery W3, SUPERB v14 C2-C4 trains): the handler-DX
  battery. `ResultHandler` — the return-(T,-error) seam: JSON 200 happy
  path, 204 for nil-pointer payloads, direct-write handlers keep ownership
  (the wrapper records the first write and steps aside), and
  `http.ResponseController` stays reachable through `Unwrap` (the
  SSE-stall wrapper class is designed out). `ResultHandlerWithValidation`
  — JSON-body bind into tagged adapter types with the `bind.*` Rejection
  codes (`bind.empty_body`, `bind.body_too_large` at 1 MiB,
  `bind.unreadable_body`, `bind.invalid_body`), optional `Validate()` hook,
  and the case-sensitivity trap documented in code and README (json/v2
  matches case-sensitively; wire tags snake_case by convention).
  `WriteError`/`SanitizeErrorMessage` — the classified error contract:
  status from `errorfamily.HTTPStatus`, JSON field names matching the
  errorpages rendering (one taxonomy, two renderings), and the B9 no-leak
  guarantee (server-side families render family defaults; client-facing
  text passes the leak-pattern scrub; `Public()` is the vouched opt-out;
  cause chains/context maps never serialize).
- Tests: the classification-parity oracle against the errorpages module
  (every family through BOTH surfaces asserts the same status + contract
  shape — test-only dep), the leaked-detail matrix (paths, Windows paths,
  source locations, SQL, goroutine headers — scrubbed or defaulted on the
  wire), the bind contract table, and the direct-write ownership test.
