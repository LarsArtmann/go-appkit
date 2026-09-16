# security

Opt-in HTTP security batteries for appkit services. Nothing here joins any
default middleware stack — every battery is an explicit middleware or
function you wire per route or per service.

`go get github.com/larsartmann/go-appkit/security`

## Batteries

| Battery                    | One-liner                                                                    |
| -------------------------- | ---------------------------------------------------------------------------- |
| `APIKeyAuth(key)`          | Shared-secret guard: `X-Api-Key` header (any method) or `?key=` (GET/HEAD only, so keys never ride URLs on writes). SHA-256 + constant-time. Empty key = inactive (dev). |
| `APIKeyCSRFBypass(csrf)`   | Wraps CSRF so header-bearing requests skip the token dance — the key is still validated downstream (fail-closed). |
| `CSRF(cfg, logger)`        | Double-submit CSRF (httputil's nosurf) with the `*` misconfiguration degraded loudly, never allow-all. |
| `RateLimit(profile)`       | Keyed rate limiting with a MANDATORY `MaxKeys` cap (a keyed limiter without a cap is a memory DoS). 429 aborts the chain — it can never be overwritten by a later handler. Named profiles: `GeneralProfile`, `AnalysisProfile`, `ExportProfile`, `ContactProfile`. |
| `OriginCheck(origins, lg)` | Origin→Referer validation with same-origin always allowed; no-header requests pass (machine clients). |
| `BodyLimit(n)`             | `http.MaxBytesReader` — oversize reads fail as a typed `*http.MaxBytesError`, never silent truncation. |
| `SanitizeText` / `SanitizeURL` | HTML strip (bluemonday strict + control-char removal) and URL validation. URLs must NEVER run through the text sanitizer: bluemonday entity-rewrites query strings (`&not=` → `¬=`). |
| `GenerateNonce` + `BuildCSP` | CSP nonce infrastructure and a deterministic policy builder. `'unsafe-eval'` is never emitted, in ANY environment — `script-src *` does not imply eval, and DataStar/Alpine-style expression compilation dies without it. JSON-LD exemption available (data, not code). |
| `SecurityHeaders(cfg)`     | Environment-tuned headers. HSTS (two-year, includeSubDomains, preload) is production-only — dev/staging stay clean (the LAN-over-http lockout trap). |

## Middleware order

Security properties live in the order:

```
RateLimit → OriginCheck → APIKeyCSRFBypass(CSRF(...)) → APIKeyAuth(key) → handler
```

The bypass OUTSIDE the CSRF middleware; the guard INSIDE it — a
header-bearing request skips the token dance but never the authentication.

## Provenance

Ports of the CV family stack's production middleware
(`platform/middleware/`, `internal/sanitization/`), specified in
`doc/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md`
(W2). Every doc comment carries the threat model that motivated it; the
regression tests port CV's acceptance pins (429-not-overwritten,
query-key-rejected-on-POST, the `&not=` trap, eval-never, HSTS-off-outside-production).

## License

PROPRIETARY — see [LICENSE](LICENSE).
