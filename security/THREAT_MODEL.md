# Threat Model — security module

Per-battery threat → design → test mapping for the opt-in HTTP security
batteries. Each battery is a port of the CV production stack's middleware
(`platform/middleware/`, `internal/sanitization/`); the full battery spec
lives in `doc/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md`
(W2). Every file's doc comment carries its threat in place; this page is the
cross-battery view.

## Scope

**In scope:** unauthenticated/cross-site HTTP attacks against a single
service's endpoints — credential guessing, cross-site request forgery,
request flooding, memory exhaustion via unbounded input, script injection
through stored or reflected content, protocol downgrade of browser
connections.

**Out of scope (deliberate):** user identity/session management (shared-key
auth only — no user database), TLS termination (core has no TLS; put a
reverse proxy in front), WAF-grade request inspection, multi-service
topology (mTLS, service mesh).

## The chain and its order

```
RateLimit → OriginCheck → APIKeyCSRFBypass(CSRF(...)) → APIKeyAuth(key) → handler
```

The order is load-bearing:

- **RateLimit first** — a flooding client must be rejected before it spends
  CSRF tokens, key comparisons, or body bytes.
- **Bypass OUTSIDE the CSRF middleware** — the wrapper decides whether the
  token dance runs; the guard INSIDE it means a header-bearing request skips
  the token dance but never the authentication (fail-closed, pinned by
  `TestAPIKeyCSRFBypass_StillValidates`).
- **APIKeyAuth last before the handler** — every cheaper rejection has
  already happened when the constant-time comparison runs.

## Per-battery threat map

| Battery                                              | Threat it stops                                                                                                                                                                                                                                                                                                          | Design pin                                                                                                                               | Regression tests                                                                                                                                                                                                  |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `APIKeyAuth(key)`                                    | Credential guessing + keys leaking via logs/history on writes: SHA-256 + constant-time compare (no timing oracle); `?key=` accepted on GET/HEAD only, so keys never ride URLs on state-changing requests; wrong key fails closed; empty key = inactive (dev/test ergonomics without a config flag).                      | Non-ambient credential: a cross-site page cannot set a custom header (CORS preflight blocks it).                                         | `TestAPIKeyAuth_*` (7): accepts correct key, query key rejected on POST / accepted on GET, header wins over query, empty key inactive, wrong key fails closed, `TestAPIKeyMatches` shares the hashing discipline. |
| `APIKeyCSRFBypass(csrf)`                             | Double protection friction: machine clients cannot do the browser token dance — but CSRF only matters for AMBIENT credentials (cookies); a non-ambient key header is itself cross-site-proof.                                                                                                                            | Bypass requires the key AND the key is still validated downstream by `APIKeyAuth` (the bypass never authenticates).                      | `TestAPIKeyCSRFBypass_HeaderSkipsTokenDance`, `_StillValidates`, `_WithoutHeaderKeepsCSRF`.                                                                                                                       |
| `CSRF(cfg, logger)`                                  | Cross-site state-changing requests riding ambient browser credentials; plus the misconfiguration trap: `TrustedOrigins: ["*"]` silently meaning allow-all.                                                                                                                                                               | Double-submit cookie via nosurf; `*` is stripped and logged loudly — degrade-with-visibility, never allow-all.                           | `TestCSRF_StarNeverMeansAllowAll`, `_CrossSitePostRejected`, `_CleanConfigUntouched`.                                                                                                                             |
| `RateLimit(profile)`                                 | Request flooding per client; AND the subtle one — a KEYED limiter without a key cap is a memory-exhaustion DoS (every attacker IP mints a bucket), so `MaxKeys` is mandatory. Plus the 2026-08-16 CV bug: a later handler overwriting the 429 with a 200 (every limit bypassable).                                       | 429 + `Retry-After` ABORTS the chain — `next` is never invoked, so the 429 cannot be overwritten. Per-profile key caps 5k–10k in CV.     | `TestRateLimit_429NotOverwritten`, `_RetryAfterHeader`, `_ProfilesCarryMandatoryMaxKeys`, `_KeysAreIsolated`, `_WindowRespected`.                                                                                 |
| `OriginCheck(origins, logger)`                       | Defense-in-depth CSRF for browser form posts (state-changing endpoints reachable without cookies, where the token dance is the only other layer). Same-origin always allowed; no-Origin/no-Referer machine clients pass.                                                                                                 | Origin header validated against the allowlist with Referer fallback; `*` allowed only with a loud log (`_AllowAllLogs`), never silently. | `TestOriginCheck_SameOriginAllowed`, `_CrossOriginRejected`, `_NoOriginOrRefererPasses`, `_RefererFallback`, `_CaseInsensitiveMatch`, `_AllowAllLogs`.                                                            |
| `BodyLimit(n)`                                       | Memory/disk exhaustion via unbounded request bodies — including the chunked-encoding bypass: unlike a Content-Length check, the limit is enforced on the actual byte stream.                                                                                                                                             | `http.MaxBytesReader` → typed `*http.MaxBytesError`; handlers map it (errors.AsType) to 413 — never a silent truncation.                 | `TestBodyLimit_TypedErrorOnOversize`, `_UnderLimitPasses`.                                                                                                                                                        |
| `SanitizeText` / `SanitizeTextSlice` / `SanitizeURL` | Stored/reflected XSS through user text; control-character smuggling. And the sanitizer's own trap: URLs must never run through the text sanitizer — bluemonday entity-rewrites query strings (`&not=` → `¬=`), silently corrupting links.                                                                                | Strict bluemonday policy (all elements/attributes stripped) + control-char removal; URLs use a separate validation path.                 | `TestSanitizeText_StripsTagsAndControlChars`, `_PreservesNewlinesAndTabs`, `TestSanitizeTextSlice`, `TestSanitizeURL_EntityQueryUnmangled` (the `&not=` pin), `_Validation`.                                      |
| `GenerateNonce` + `BuildCSP` + `AddNonceToScriptSrc` | Script injection via inserted content; AND the configuration traps: `unsafe-eval` (script-src * does NOT imply eval, but DataStar/Alpine-style expression compilation dies without it — the answer is NO eval in ANY environment, not a dev carve-out that leaks to prod); JSON-LD over-blocking (it is data, not code). | Deterministic policy builder with per-request 128-bit nonces (OWASP minimum); eval not emit-able in any tier.                            | `TestCSPProductionBlocksInlineScripts`, `TestCSPHeaderParsesForEveryEnvironment`, `TestCSP_JSONLDExemption`, `TestCSP_WildcardDoesNotImplyEval`, `TestAddNonceToScriptSrc`, `TestNonceContextRoundTrip`.          |
| `SecurityHeaders(cfg)`                               | Browser-side protections missing (MIME sniffing, framing/clickjacking, referrer leakage, over-powerful permissions); AND the HSTS lockout trap: a browser that once saw HSTS on a LAN-over-http host refuses plaintext afterwards — so HSTS (two-year, includeSubDomains, preload) is PRODUCTION-ONLY.                   | Standard header set delegated to httputil; environment switch with explicit Development/Staging overrides.                               | `TestSecurityHeaders_HSTSOffOutsideProduction`, `_HSTSOnInProduction`, `_StandardHeadersPresent`, `_CSPPassedThrough`.                                                                                            |

## Composition proofs

The per-battery unit tests above are the floor. Composition is proven at two
more levels:

- **Cross-module E2E:** `integration/hardened_dashboard_test.go` composes
  the health module's `DashboardHardenedPreset` with this module's nonce
  machinery behind a strict CSP, against published tags.
- **Runnable example:** `example/` wires the full chain (canonical order,
  per-request nonces, per-route limiter instances) on an appkit service with
  a curl walkthrough in its doc comment; every documented status code was
  verified live against the running demo (2026-09-23).

## Known limits

- Shared-key auth has no rotation, revocation, or per-client identity —
  it is a machine-client gate, not an identity system.
- `OriginCheck` passes requests with no Origin and no Referer (machine
  clients); it is defense-in-depth, never the only CSRF layer.
- The rate limiter is in-process: instances behind a load balancer each keep
  their own buckets (per-instance, not global, limits).

## License

PROPRIETARY — see [LICENSE](LICENSE).
