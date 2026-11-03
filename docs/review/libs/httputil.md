# httputil — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). This report
> narrows and updates the httputil rows of `docs/research/2026-09-16_larsartmann-ecosystem-deep-dive.html`
> and the 2026-10-11 sibling audit (newer and narrower wins).

## Verdict

**USED — the deepest sibling adoption in the repo, at the latest tag (v1.7.0), but roughly
two-thirds of the library's capability surface is untouched.** Adoption score **69/100**
(moderate-high): 11 of 18 applicable capability rows fully leveraged, 3 partial, 4 missed.
The misses split into three classes: a planned-but-pending server seam refactor, four
middleware families with no adoption story yet (CORS, ClientIP, compression, trusted-proxy
keying), and one refactor opportunity created *after* the fact (otel's custom Recovery
predates httputil v1.6.0's `OnPanic` hook).

| Fact | Value | Evidence |
|---|---|---|
| In-graph version | v1.7.0 (= latest tag) | `go.mod:8` + 7 satellite go.mods |
| Direct importers | core, flightrecorder, otel, security, integration; indirect in errorpages/health/systemd | `go.mod` files |
| Runtime symbols touched | 21 of ~120+ exported | `rg 'httputil\.[A-Z]' --glob '!*_test.go'` |
| Latest tag vs HEAD | v1.7.0; HEAD 30 commits ahead (deps/apidiff/unreleased work) | `git -C ~/projects/httputil rev-list --count v1.7.0..HEAD` |

## 1. Usage evidence (what appkit actually does with it)

**Core default stack** — 5 of httputil's ~20 middlewares (`middleware.go:14-31`):
`Recovery → RequestID(DefaultRequestIDConfig) → Logging → [Timeout unless WriteTimeout ==
NoTimeout] → SecurityHeaders(DefaultSecurityHeadersConfig)`, chained via `httputil.Chain`
(`service.go:89`) into the `httputil.Middleware` type the whole framework speaks
(`middleware.go:45`).

**Health endpoints** — core re-exports `httputil.RegisterHealth` and
`ReadyHandlerWithProbe`/`HealthHandler`/`LiveHandler` (`health.go:12-13`, `service.go:59-61`).

**Recorder** — `httputil.NewResponseRecorder`/`ResponseRecorder` used at the root
(testkit/service paths). errorpages deliberately keeps its own `statusRecorder`
(`errorpages/errorpages.go:83,206`) because `Wrap` needs body-discarding capture — decision
recorded in errorpages v0.1.2 CHANGELOG. Not a defect; a recorded divergence.

**Security module** — composes instead of reimplementing where httputil already has the wheel:
`CSRFMiddleware`/`CSRFConfig` (`security/csrf.go:13,25,38`),
`KeyedRateLimiterMiddleware`/`KeyedRateLimiterConfig`/`KeyExtractor`
(`security/ratelimit.go:38,105-106`), `SecurityHeaders` for the config type
(`security/headers.go:33,40,55`).

**OTel module** — uses the `httputil.Middleware` type for adapter signatures
(`otel/middleware.go:101,117,122`, `otel/exceptions.go:19-37`) — but ships its OWN Recovery
(see finding F4).

**Tests** — root `httpspec_test.go:10-11` imports `httputil/httpspec` (behavioral HTTP spec
framework) and uses `Language`/`LanguageFromContext` test-only.

## 2. Capability inventory vs adoption

| # | httputil capability | Status | Evidence / gap meaning |
|---|---|---|---|
| 1 | `Chain`/`Compose`/`Middleware` type | 🟢 Full | `service.go:89`, `middleware.go:45`; 10+ sites |
| 2 | `Recovery` | 🟢 Full | default stack `middleware.go:18` |
| 3 | `RequestID` + `DefaultRequestIDConfig` | 🟢 Full | `middleware.go:19` |
| 4 | `Logging` | 🟢 Full | `middleware.go:20` (perf baseline: +30µs emitted, AGENTS) |
| 5 | `LoggingFromContext` | 🟡 Partial | comment-only (`otel/doc.go:109`); zero runtime wiring |
| 6 | `Timeout` + `NoTimeout` opt-out | 🟢 Full | `middleware.go:24-26`; SSE contract pinned |
| 7 | `SecurityHeaders` | 🟢 Full | `middleware.go:28-30`, `security/headers.go` |
| 8 | Health endpoints (`RegisterHealth` + 3 handlers) | 🟢 Full | `health.go:12-13`, `service.go:59-61` |
| 9 | `ResponseRecorder` | 🟢 Full | root usage; errorpages divergence is recorded |
| 10 | CSRF (`CSRFMiddleware`/`CSRFConfig`) | 🟢 Full | via security, `security/csrf.go` |
| 11 | Keyed rate limiting (`KeyedRateLimiter*`) | 🟢 Full | via security, `security/ratelimit.go:105-106` |
| 12 | `httpspec` test framework | 🟢 Full | `httpspec_test.go:10` |
| 13 | `Server`/`NewServerListener` (v1.7.0 seam) | 🟡 Partial | seam SHIPPED, refactor NOT — core still owns a raw `http.Server`; spike verdict + re-run are queued (AGENTS "Architecture and Control Flow", `docs/planning/2026-09-16_composition-spike-verdict.md`) |
| 14 | `MaxBodySize`/`MaxBodySizeMiddleware` | 🟡 Partial | security hand-rolls `BodyLimit` (`security/bodylimit.go:18`) — deliberate typed-413 contract (`security/THREAT_MODEL.md:49`), but two MaxBytesReader wrappers now exist in-family |
| 15 | `Nonce`/CSP builders | 🟡 Partial | security owns `GenerateNonce`+`BuildCSP` policy tiers (`security/csp_nonce.go`); httputil's middleware+`RecommendedCSPWithNonce` unused — two nonce generators in-family |
| 16 | `CORS` | 🔴 Missed | zero usage; `OriginCheck` (`security/origin_check.go`) is CSRF origin validation, NOT resource-sharing CORS — different concern, real gap for browser-API consumers |
| 17 | `ClientIP`/`ClientIPMiddleware`/`TrustedProxySet` | 🔴 Missed | rate limiter defaults to `KeyExtractorFromRemoteAddr`; v1.6.0's trusted-proxy keying and `ClientIP` context are entirely unused |
| 18 | `Compression`/`Decompression` | 🔴 Missed | zero usage; no appkit story for gzip/deflate on either direction |
| 19 | `Metrics` middleware | ⚪ N/A | fleet observability standard is the otel module (OTel semconv); httputil's zero-dep counter hook is for non-otel services |
| 20 | `Language` negotiation | ⚪ N/A | i18n module researched and declined (`docs/planning/2026-10-09_i18n-module-research.md`, decision memo 2026-10-10); test-only usage only |
| 21 | `MiddlewareStack`/`queryparam`/`capabilities` | ⚪ N/A | no route-table or query-binding story yet — battery W3-B2/B6 territory |

**Score derivation** (library-deep-dive rubric): A = 18 applicable rows (19-21 N/A with
reasons). Fully leveraged = 11 rows (1, 2, 3, 4, 6, 7, 8, 9, 10, 11, 12). Partially used =
4 rows × 0.5 = 2.0 (rows 5, 13, 14, 15). Missed = 3 rows (16, 17, 18).
**Score = (11 + 2.0) / 18 = 72/100 (high band, lower edge).**

## 3. Findings (Pareto order)

### F1 — Server seam refactor is the single highest-value adoption step (planned, pending)
httputil v1.7.0 shipped exactly the seam the composition spike asked for:
`NewServerListener(ln, cfg, handler)` with synchronous bind-failure classification and
`ListenerAddr()` from construction (`~/projects/httputil/server.go:216-256`). Core still
binds its own `http.Server` with hand-rolled lifecycle semantics. The spike
(`docs/planning/2026-09-16_composition-spike-verdict.md`) rated the config mapping clean;
the remaining work is re-running the spike against v1.7.0's single-serve
`server.listener_consumed` semantics, re-pinning the `Addr()`/`Running()` shutdown
phase-log contract, then swapping. The Core TLS option rides the same seam. This is
already on the books — this report adds nothing but urgency math: it is the difference
between 72 and ~83 on this scorecard.

### F2 — CORS is a real missed capability, not covered by security's OriginCheck
`security/origin_check.go` validates the Origin header as CSRF defense-in-depth (same-origin
always allowed, allowlist otherwise). httputil `CORS(cfg)` answers a different question:
cross-origin *resource sharing* (preflight, credentialed subdomain wildcards,
private-network/LNA). A browser-API consumer serving cross-origin clients on appkit today
must hand-roll CORS or leave the framework. Demand-gated per house rules, but it belongs on
the W3-httpx battery shortlist next to B4/B5.

### F3 — Trusted-proxy ClientIP keying: shipped upstream v1.6.0, unused here
`KeyExtractorFromTrustedClientIP(cidrs)` + `TrustedProxySet` (`~/projects/httputil/ratelimit_keyed_trusted.go`)
solve the "rate-limit key behind a load balancer" problem appkit's security module will hit
the first time a consumer deploys behind a proxy. One-line wiring when demand appears;
recorded here so the demand signal is recognizable.

### F4 — otel's custom Recovery can now delegate to `RecoveryWithConfig(OnPanic: …)`
`appkitotel.Recovery` hand-replicates httputil.Recovery's behavior plus exception-span
capture (`otel/exceptions.go:19-37`) — written BEFORE httputil v1.6.0 (2026-10-24) shipped
`RecoveryWithConfig`/`OnPanic` for exactly this integration. Refactoring to
`httputil.RecoveryWithConfig` + an `OnPanic` hook that records the exception would delete the
behavioral-duplication surface (log-line contract, `ErrAbortHandler` re-panic) while keeping
the OTel payload. Cheap, testable, and pinned by the existing exceptions tests. **This is the
one actionable code finding of this audit.**

### F5 — Two deliberate in-family duplications to keep an eye on (not defects)
- `security.BodyLimit` vs `httputil.MaxBodySize`: both wrap `http.MaxBytesReader`. The
  security version exists for the typed `*http.MaxBytesError` → 413 contract
  (`security/THREAT_MODEL.md:49`); httputil's middleware classifies config errors and
  integrates with the code registry. Consolidation candidate when W3-httpx lands, not before.
- `security.GenerateNonce`/`BuildCSP` vs `httputil.Nonce`/`RecommendedCSPWithNonce`: appkit
  needs policy-tiered CSP builders (strict production, no eval in any tier); httputil's is a
  middleware + fixed builders. Overlap is partial and deliberate.

## 4. Should appkit use it MORE? (the explicit answer)

Yes — in exactly three places, all already named in living docs: (1) the server-seam
refactor (planned), (2) CORS + trusted-proxy keying when browser/proxy consumers demand
(W3/W4 battery candidates), (3) the otel Recovery delegation (new, cheap, do next train).
No — for Language (i18n declined), Metrics middleware (otel owns observability), and
compression (no demand signal yet; would join the default stack only with a consumer).

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'httputil\.[A-Z]' --glob '*.go' --glob '!*_test.go' -o | sed 's/.*://' | sort -u   # 21 runtime symbols
rg -n 'httputil' --include=go.mod -r '' -g 'go.mod'                                      # v1.7.0 × 8 go.mods
git -C /home/lars/projects/httputil tag --sort=-v:refname | head -1                       # v1.7.0
rg -n 'RecoveryWithConfig|OnPanic' otel/        # empty → F4 stands
sed -n '14,31p' middleware.go                   # the 5-middleware default stack
```
