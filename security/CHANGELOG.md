# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Added

- `example/` demo service: the full hardened chain (RateLimit →
  OriginCheck → APIKeyCSRFBypass(CSRF) → APIKeyAuth, per-route limiter
  instances, per-request CSP nonces, typed body limit, environment-tuned
  headers) on a real appkit service, with a curl walkthrough in the
  command doc comment — every documented status code verified live
  against the running demo.
- `THREAT_MODEL.md`: the per-battery threat → design → regression-test
  mapping, the chain-order rationale, composition proofs, and the known
  limits (shared-key auth is a gate, not an identity system).

### Changed

- The module go.mod now requires `github.com/larsartmann/go-appkit`
  v0.5.1 — EXAMPLE-ONLY (the security package itself keeps zero appkit
  dependencies), mirroring the errorpages module's example pattern.

## [0.1.0] - 2026-09-16

### Added

- First release of the security module (package `security`): opt-in HTTP
  security batteries ported from the CV production stack per the canonical
  battery spec (W2). API-key auth with GET/HEAD-only query fallback
  (`APIKeyAuth`), CSRF with the API-key bypass (`CSRF`, `APIKeyCSRFBypass`),
  keyed rate-limit profiles with a mandatory key cap and a 429-aborts-chain
  contract (`RateLimit`), origin checking (`OriginCheck`), typed body
  limits (`BodyLimit`), text/URL sanitization with the `&not=`→`¬=`
  bluemonday trap handled (`SanitizeText`, `SanitizeURL`), CSP nonce
  infrastructure + deterministic policy builder with eval NEVER grantable
  (`GenerateNonce`, `BuildCSP`, `AddNonceToScriptSrc`), and
  environment-tuned security headers with production-only HSTS
  (`SecurityHeaders`). All opt-in; nothing joins any default stack.

## License

PROPRIETARY — see [LICENSE](LICENSE).
