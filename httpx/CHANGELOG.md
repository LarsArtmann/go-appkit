# Changelog

All notable changes to this module are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the module follows
semantic versioning.

## [0.2.0] - 2026-11-08

### Added

- B3 ResponseWriter contract: `WrapWriter(w) *Writer` keeps optional
  interfaces reachable through wrappers — Flush and Hijack forward
  whenever the base supports them (Hijack fails with a static error
  when it cannot), `Unwrap` always, so `http.ResponseController` and
  direct type assertions find the real connection through any chain.
  Companion test-only package `httpxtest` (`AssertForwardsInterfaces`,
  `ForwardingFindings`) pins the forwarding guarantee from consumer
  test code — the probe discriminates hidden interfaces, swallowed
  writes/status, no-op stubs, and controller-opaque chains (three
  buggy-wrapper archetypes pinned as failing).
- B4 conditional requests: `Conditional(handler, opts...)` composes
  go-etag/server's ETag middleware (default strong sha256) with the
  httpx-side If-Modified-Since layer — IMS is honored ONLY when
  If-None-Match is absent (RFC 9110 precedence), on GET/HEAD, decided
  at the first write; a 304 strips content headers but keeps
  validators. Options: `ETagFunc` (custom hash), `SkipConditional`
  (per-request opt-out). Helper `SetLastModified(w, t)` stamps
  Last-Modified from server code.
- B5 content negotiation on go-codec (the owner-ruled family encode
  boundary): `Negotiate(r, offers...)` picks the best offer by
  q-ranking, then specificity, then offer order (q=0 refuses; absent
  Accept = `*/*`; the most specific matching range governs — a q=0
  exact refusal is never resurrected by `*/*`); `RequestCodec(r)` maps
  Content-Type to the go-codec JSON/CBOR codecs (+json/+cbor suffix
  forms; empty → JSON; unknown → `ErrUnsupportedMediaType`, a
  Rejection carrying HTTPStatus 415); `ResponseCodec(r, reqCodec)`
  picks the response codec from the Accept header (tie → CBOR;
  wildcards express no preference → the request codec's family).
- B7 per-route write deadlines: `WriteDeadline(extension, prefixes...)`
  returns a middleware that extends the stream write deadline on
  matching routes (`ResponseController.SetWriteDeadline` before the
  next handler); non-matching routes pass through untouched; an opaque
  ResponseController chain fails loudly (Infrastructure
  `httpx.write_deadline_unreachable`, HTTP 500 — a wiring bug, not an
  outage).
- Dependencies: go-etag/server v0.6.1 and go-codec v0.3.3 promoted to
  direct requires (entitytag v0.6.1 rides indirect). Lint config: the
  fleet `_test.go` exclusion union (a v0.1.0 onboarding gap) and an
  ireturn allow for go-codec's Codec interface (the otel OTel-API
  precedent).

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
