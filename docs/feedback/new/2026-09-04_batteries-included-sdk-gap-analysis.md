# Feedback: What go-appkit needs to become a REAL batteries-included SDK

- **From:** the CV consumer perspective (`github.com/LarsArtmann/CV` — the family's deepest hand-rolled HTTP stack)
- **Date:** 2026-09-04
- **Method:** every gap below was verified against go-appkit source at HEAD (core `middleware.go`/`service.go`/`config.go`, `realtime/`, `health/`, `errorpages/`, `cqrs/`, TODO_LIST) and against CV's production code. Nothing here proposes a feature that already exists; where TODO_LIST already tracks an item, this doc endorses it and adds consumer evidence instead of duplicating it.
- **Audience:** go-appkit maintainer (you). References marked **[CV-AGENTS]** point at `CV/AGENTS.md` sections; **[appkit-TODO]** at `go-appkit/TODO_LIST.md`.

---

## 0. Executive summary

go-appkit's core is deliberately lean, and that is correct. But "batteries included" and "lean core" are **not** in tension — the module system is already the battery bay (`health`, `otel`, `realtime`, `errorpages`, `cqrs`, `docs`, `flightrecorder`). What is missing is the rest of the battery pack that real family services demonstrably need: **the security cluster, the handler-DX layer, long-running-worker supervision, storage/politeness kits, and a testing toolkit.**

The evidence base is not hypothetical. CV runs a **13,876-LOC** hand-rolled HTTP/runtime layer (server 4,117 + httpx 2,752 + middleware 7,007, non-test) guarded by **51 test files**, of which at least 9 invariants exist **only because production broke without them** (SSE Flush hiding, 429 overwrite bypass, recycled-writer corruption, goroutine leaks, nosurf origin-trust mismatch, CSP eval kill). That stack consumes **20 distinct `httputil` symbols** — the same library appkit composes. Every item below is one of two things:

1. **A module CV (or PapDashboard/cqrs-htmx) already hand-rolled** and would have consumed from appkit instead, or
2. **A known TODO_LIST item** where this doc adds the missing consumer-demand signal.

**Thesis: batteries as opt-in modules, core unchanged.** Each proposal names its target module. Core only grows where the battery is impossible to build outside it (TLS, metrics surface, one or two lifecycle hooks).

---

## 1. The demand signal, in one table

What the reference consumers hand-rolled *while appkit existed*:

| Consumer need | CV built it as | appkit today | Verdict |
| --- | --- | --- | --- |
| CSRF + API-key bypass | `httputil.CSRFMiddleware` + `APIKeyCSRFBypass` | nothing | **Battery gap A1/A2** |
| API-key auth | `middleware/apikey.go` (const-time, header/query rules) | nothing | **A2** |
| Rate limiting (profiles, keyed) | `ratelimit.go` — 4 profiles, ClientIP+XFF | nothing | **A3** |
| Origin check | `origin_check.go` | nothing | **A4** |
| CSP nonce plumbing | `csp_nonce.go` + ConsolidatedSecurityMiddleware + templ wiring | SecurityHeaders only | **A5** |
| Input sanitization | bluemonday Text/URL | nothing | **A6** |
| Typed handler result + error classification | `httpx` ResultHandler[T] + classification parity tests | go-error-family exports only; no handler seam | **B1** |
| Bind/validate | httpx bind/validation | nothing | **B2** |
| Conditional GET (ETag/Last-Modified) | ICS feed: sha256 ETag, RFC 9110 weak/comma, 304s | `go-etag` sits in core as an *indirect* dep, unused | **B4** |
| Content negotiation | `/admin/health` HTML vs JSON | nothing generic | **B5** |
| Route introspection + golden tests | `goldenRoutes` + `TestRouteDiscovery` | nothing | **B6** |
| Long-request budgets | `ExtendWriteDeadline` per prefix (LLM missions) | `NoTimeout` global only | **B7** |
| SSE observability + projection wiring | sse-stats drop counters, SubscribeFolded discipline | `Hub.Health()`/`SubscriberCount()` exist; no counters surface, no projection contract | **C1/C2** |
| Background worker supervision | projection loops, reply-loop poller, warmups — all hand-rolled | nothing | **D1** |
| DI bridge | CV is samber/do end-to-end | core is do-free (verified: no samber/do in core go.mod) | **D2** |
| SQLite operational kit | PRAGMA discipline, leases, VACUUM backup | cqrs module has storage; no general kit | **D3** |
| Atomic file writes | go-atomic-write ≥ v0.5.1 (Windows trap documented) | nothing | **D4** |
| Polite outbound HTTP | HostPacer, ETag cache w/ byte budget, SSRF guard, Retry-After | nothing | **D5** |
| Webhook sender | (go-health-dashboard has one internally) | nothing | **D6** |
| Test harness kit | full-chain vs mux trap, SSE test client, leak assertions, XFF bucket helper | `integration/` module (cross-repo, not a consumer kit) | **E1–E4** |
| Config + env-overrides | koanf + discoverability tests | nothing | **F1** |
| TLS | — | **known gap** [appkit-TODO, PapDashboard demand] | **G1** |
| Core metrics surface | `/metrics` (Prometheus) | otel module only | **G2** |

---

## 2. Battery cluster A — Security (the biggest gap)

CV's security middleware is the single largest body of code appkit consumers must otherwise re-invent — and the only kind where a naive re-implementation is actively dangerous. All of it lives in `httputil`-compatible shapes today, so most items are *composition*, not invention.

### A1 — CSRF middleware (module: `appkit/security` or promoted httputil default)

- **What:** double-submit cookie CSRF with trusted-origin allowlist; config for Secure/SameSite; the trusted-origins propagation rule (`"*"` must NOT silently mean allow-all for CSRF — CV logs-and-degrades that misconfiguration **[CV-AGENTS: LAN development]**).
- **Why:** every family service with browser forms needs it; CV got burned by the subtle half: a `Trusted-Origin` header does **not** bypass nosurf's token check (empirically verified 2026-08-29). That trap deserves to be codified once, upstream, with tests.
- **Shape:** `security.CSRF(cfg)` + `security.CSRFConfig{TrustedOrigins, Secure, ...}` mirroring `httputil.CSRFConfig`.
- **Accept:** same-site form POST flow test; cross-origin 403; `*`-origins behavior pinned; skip-rule integration with A2.

### A2 — Non-ambient-credential auth: API-key middleware + CSRF bypass

- **What:** `security.APIKey(key)` (SHA-256 + constant-time compare, `X-API-Key` header) plus the two rules CV learned the hard way: (1) `?key=` query fallback for **safe methods only** — a browser link or calendar client cannot set headers, but query strings land in access logs so writes must stay header-only (pinned: `TestAPIKeyAuth_QueryKeyRejectedOnPost`, header-wins precedence test); (2) requests bearing the header skip CSRF entirely (non-ambient credential: cross-site pages cannot set custom headers), while auth still validates downstream — fail-closed (`TestAPIKeyAuth_AcceptsCorrectKey` guards the `c.Next()` pass-through trap: returning without it serves an empty 200).
- **Why:** CV's SystemNix `cv-scan.timer`, the interviews ICS feed, and every scripted client ride exactly this. PapDashboard's webhooks will want it the day they add an inbound control surface.
- **Shape:** `security.APIKeyAuth(key)` + `security.APIKeyCSRFBypass(csrf Middleware)` (a wrapper, not a flag — the composition IS the design).
- **Accept:** the four CV pin-tests ported upstream; docs state the ambient/non-ambient threat model in five sentences.

### A3 — Keyed rate limiting with profiles

- **What:** `security.RateLimit(security.RateLimitConfig{Limit, Burst, Key: KeyExtractorFromClientIP})` + named profiles (CV: general 60/min·100, analysis 10/min·15, export 5/min·8, contact 8/min·10) + `Retry-After` on 429 + **the short-circuit contract**: a rejected request must abort the chain so later handlers cannot overwrite the 429 with a 200. That bug made every CV limit bypassable on 2026-08-16.
- **Why:** k6 load tests against CV are shaped around these profiles; the framework's own health endpoints will need them the first time a dashboard polls `/health/sse` in a loop (CV's admin-hub e2e blew the analysis bucket with 14 SSE connections — a documented flake).
- **Shape:** profiles as `security.Profile` presets; `Key` pluggable (`KeyExtractorFromClientIP` already exists in httputil and honors `X-Forwarded-For` — document the shared-bucket consequence for test harnesses).
- **Accept:** 429-not-overwritten regression test upstream; XFF uniqueness helper for tests (see E3).

### A4 — Origin check

- **What:** `security.OriginCheck(allowed []string)` — Origin→Referer fallback, 403 on miss, `"*"`/empty = allow-all (deliberately, but loudly).
- **Why:** defense-in-depth behind CSRF for state-changing endpoints; CV applies it to pipeline/A.Team/chat/contact.
- **Accept:** same contract as CV's `origin_check_test.go`.

### A5 — CSP nonce infrastructure (the module that prevents the next DataStar incident)

- **What:** per-request nonce minting + context helper (`NonceFromContext`), an optional nonce extractor seam for component libraries (CV feeds it to go-health-dashboard via `WithNonceExtractor`), and a documented production policy: `script-src 'self'` + nonce, `'unsafe-eval'` and `unsafe-inline` for scripts **never granted** — with the explanation of the subtle killer: `script-src *` does NOT imply eval, and DataStar/Alpine compile expressions via `new Function()` and die silently.
- **Why:** the DataStar removal is the single most expensive CSP lesson in the family (a whole dashboard class died on it, 2026-08-16). appkit's consumers will hit the exact same wall the first time they mount an SSE-patch UI. `style-src 'unsafe-inline'` stays grantable by decision (dynamic inline styles can't be hashed) — encode that nuance in docs.
- **Shape:** `security.CSPNonce()` middleware + `security.Nonce(r)` + option hooks matching what go-health-dashboard's `WithNonceExtractor` already consumes (upstream symmetry: dashboard and framework share one extractor contract).
- **Accept:** production-CSP test (inline script blocked, nonced script executes, JSON-LD exempt); policy parse/determinism tests (CV's `csp_parse_test.go` is the port-ready oracle).

### A6 — Input sanitization

- **What:** `security.SanitizeText` (bluemonday strict + control-char strip) and `security.SanitizeURL` (parse → http/https scheme + host required — NO bluemonday, which entity-rewrites query strings: `&not=` → `¬=`, verified live in CV 2026-09-04).
- **Why:** all CV form endpoints (contact, chat, A.Team) sanitize; the URL variant exists because the text variant corrupts URLs. One battery, two functions, hard-won edge case included.
- **Accept:** the `&not=` regression test.

### A7 — Environment-tuned security headers

- **What:** extend the default `SecurityHeaders` composition with env awareness: HSTS `max-age=63072000; includeSubDomains; preload` **only in production** (CV's `securityHeadersConfig`), dev/staging leave it off so LAN http works.
- **Why:** today every consumer re-derives this five-liner; getting HSTS wrong on a LAN host is a lockout.

---

## 3. Battery cluster B — Handler DX (the `httpx` layer, as a module)

This is the layer appkit consumers currently cannot have at all: appkit core exposes stdlib mux + `http.HandlerFunc`, and every richer behavior (typed results, binding, negotiation) gets hand-rolled per service. CV's `internal/httpx` is the proof-of-concept; propose it as **`appkit/httpx`** (new module, core untouched, depends on httputil + go-error-family).

### B1 — Typed result handlers with error classification

- **What:** `httpx.ResultHandler[T](func(*httpx.Context) (T, error)) http.HandlerFunc` — auto-serialization (json/v2), success shaping, and error mapping through go-error-family's taxonomy (Rejection→400, Conflict→409, Transient→503, Corruption/Infrastructure→500/503 — the SAME taxonomy `errorpages` already speaks, so HTML error pages and JSON contracts stay consistent for free).
- **Why:** CV pins classification parity between middleware and error pages (`error_classification_parity_test.go`) and fuzzes the result handler (`result_handler_fuzz_test.go`). Sharing the taxonomy with errorpages is the design win — one battery feeds the other.
- **Accept:** classification table test + a fuzz seed corpus upstream.

### B2 — Bind + validate

- **What:** typed request binding (json/v2; case-sensitive matching documented loudly — CV's untagged-struct 400 bug on `{"appId":...}` is the cautionary tale), validation hooks, and the wire-DTO rule: bind tagged adapter types, never untagged domain structs.
- **Shape:** `httpx.Bind[T](c) (T, error)` + tag-snake_case policy note (the family's `tagliatelle` stance).

### B3 — Response-writer wrapper contract (as code, not docs)

- **What:** a `httpx.WrapWriter` that forwards `Flush`/`Hijack`/`Unwrap`/`ResponseController` — plus a test helper that FAILS any wrapper which hides optional interfaces. The notFoundInterceptor bug (SSE frames stuck in net/http's 4KB buffer because an innermost wrapper hid `Flusher`) is exactly the failure this battery prevents; it compiles fine and passes unit tests with `httptest.Recorder`.
- **Why:** CV's postmortem states the rule in prose; appkit can make the rule executable.

### B4 — Conditional GET battery (promote `go-etag` from indirect dep to feature)

- **What:** `httpx.Conditional(handler, ETagor/LastModifiedSource)` — strong/weak ETag handling (RFC 9110: `W/` prefix, comma lists, `*`), `If-Modified-Since`, 304 with correct Vary; byte-deterministic body ⇒ stable hash.
- **Why:** CV's interviews ICS feed serves calendar clients that poll the entire feed; weak forms from Google Calendar had to be handled for real (`TestInterviewsICSHandler_ConditionalGET`). `go-etag v0.2.0` already sits in core's go.mod as an indirect dependency — this battery gives it a reason to exist.
- **Accept:** the weak/`*`/comma matrix test, ported.

### B5 — Content negotiation helper

- **What:** `httpx.Negotiate(r, "text/html", "application/json")` with q-value-aware defaulting.
- **Why:** CV content-negotiates `/admin/health` (browser→HTML, kubelet-ish→JSON); go-health-dashboard implements its own internally. Two implementations in the family already; batteries should be one.

### B6 — Route introspection + golden-route test helper

- **What:** `Service.Routes() []RouteInfo{Method, Pattern, Name}` (populated from a route-registration wrapper) + `httpxtest.AssertRouteSet(t, mux, golden)` — the golden-routes pattern (exact set AND count) that catches accidental route additions/removals in CI.
- **Why:** CV's route golden is 115 routes and has caught multiple accidents; stdlib mux 1.22+ patterns make this trivially introspectable if registration goes through one seam. Bonus: the same metadata feeds `docs-mod` (see B8).

### B7 — Per-route write-deadline budgets

- **What:** `httpx.WriteDeadline(d, prefix)` middleware — extends the server write deadline for matching routes only, with the documented constraint that it must sit OUTSIDE compression (the gzip wrapper does not `Unwrap`, so `ResponseController` cannot reach the connection writer through it).
- **Why:** CV's agent missions answer after minutes against a 30s server-wide write timeout; `NoTimeout` (appkit's current answer) is too blunt — the rest of the surface SHOULD keep deadlines.
- **Accept:** the compression-ordering regression test (deadline reached through gzip wrapper = must fail the ordering).

### B8 — docs-mod feed: register metadata at route registration

- **What:** if B6's route wrapper carries optional metadata (summary, tags, request/response types), `docs-mod` generates OpenAPI from the LIVE route table instead of a separate catalog pass — docs cannot drift from routes.
- **Why:** docs-mod v0.2.0 exists (catalog/v4) but is a parallel surface today; CV's route golden + docs drift is the failure mode this closes.

---

## 4. Battery cluster C — Realtime/eventing additions

The `realtime` module's bones are right (Last-Event-ID replay via `WithStore`, graceful `Shutdown` vs hard `Close`, `Health()`, `SubscriberCount()`, buffer sizing, subscribe hooks). Three additions make it production-complete for CV-shaped consumers:

### C1 — Drop/backpressure observability surface

- **What:** expose go-sse's per-subscriber drop counts + buffer saturation through `Hub.Health()` into a `GET /metrics`-shaped snapshot, and document the "dropped notifications only delay, next delivery re-syncs" invariant.
- **Why:** CV's `/api/pipeline/sse-stats` (`droppedStoreNotifications`, `droppedFoldedNotifications`, `droppedBroadcasts`, `payloadDecodeFailures`) was built because SSE staleness was invisible; Gatus now alerts on it. Batteries should ship with the gauges the operator will eventually demand.

### C2 — The projection→broadcast contract (kill the stale-broadcast race class)

- **What:** a documented + helper-ized wiring for "broadcast AFTER the projection folded the event": `realtime.FoldedSource(subscribeFn, snapshotFn)` — the helper holds the ordering guarantee (emit only after fold returns), so every consumer gets CV's `SubscribeFolded` discipline without rediscovering the race (broadcast-on-raw-event can publish pre-fold state; SSE clients then never see the update because no follow-up broadcast exists).
- **Why:** this exact race cost CV a 4-session flake hunt (2026-08-16, the SSE stale-broadcast root cause). It is a *framework-shaped* pitfall: anyone wiring cqrs projections (module D of this doc) to realtime hits it.
- **Accept:** a test where a wrapper store appends mid-fold and the broadcast provably includes the trigger (CV's `TestProjectionIntegration_SubscribeFoldedReflectsEvent` is the oracle).

### C3 — Per-subscriber auth + namespaced filtering

- **What:** `realtime.Handler` option: `WithAuthorize(func(r *http.Request) error)` evaluated at subscribe; event namespace filtering per subscriber (`WithFilter(func(evt) bool)` or topic prefixes).
- **Why:** CV's admin/pipeline SSE surfaces ride the API-key guard + rate budgets; today a consumer must wrap the whole handler and reimplement replay/dedup to filter. (AGENTS says filtering exists via go-sse duck-typing — surface it as a first-class option with the auth hook.)

---

## 5. Battery cluster D — Operations & data

### D1 — Background worker supervisor (`appkit/worker`)

- **What:** `worker.Supervisor` — run/stop lifecycle for pollers/persisters/warmups: ctx cancellation, stop-timeout join, panic recovery with restart+backoff+jitter, run counter metric, "started before/after shutdown" edge semantics (a worker started during shutdown must not leak — CV had to JOIN its boot-warm goroutine in `Shutdown` explicitly).
- **Why:** EVERY family service hand-rolls this loop: CV's DashboardProjection fold loop, reply-loop poller, DI shutdown chain; PapDashboard's notification dispatch; cqrs projection hosts already have their own. The third hand-rolled copy is the signal.
- **Accept:** goroutine-leak test (start/stop 100×, baseline stable); panic-restart test; start-during-shutdown test.

### D2 — `appkit/do` bridge module (samber/do integration, opt-in)

- **What:** register the `*Service` in a do injector (`doappkit.Register(injector, svc)`), cascade `svc.Shutdown` via `do.Shutdowner`, expose `do.Healthchecker` for the service, and provide the error-propagating `resolve[T]` helper CV built after banning `do.MustInvoke` repo-wide (forbidigo-enforced! — lazy provider closures + boot-time `mustResolve` accessors only).
- **Why:** core is do-free (verified), which is right for core — but the family's composition roots are ALL do-based (CV 37+ providers; PapDashboard; cqrs-htmx). Without the bridge, appkit's lifecycle and the DI lifecycle are two shutdown owners over one process — precisely the split-brain CV hunted down. The bridge makes ONE owner: do.
- **Shape:** ~300 LOC module; the cordis-bridge research doc already sketches the pattern for a sibling problem — same layer discipline.

### D3 — SQLite operational kit (`appkit/sqlite`)

- **What:** the PRAGMA/open-conn discipline (WAL-ish setup, `busy_timeout`, `SetMaxOpenConns(1)` semantics documented as REQUIRED for PRAGMA persistence), an **ownership lease** (file-based, hostname+pid+liveness, stale-detect via Signal 0, reentrant refcount — CV's `StoreLeaseHeldError` pattern that ended the stale-reader class), `Backup(ctx, path)` via `VACUUM INTO` (refuse-overwrite, safe beside a running writer), and a ledger-file helper (atomic 0600 JSON with cap+dedupe — the deadportals/lessons-accounts pattern CV now has four instances of).
- **Why:** the cqrs module covers the EVENT-SOURCED store, but every service also keeps non-CQRS state (CV: funnels, lessons, accounts, dead-portal markers; PapDashboard: notifications). The lease alone would have prevented the "CLI + server both open the DSN" incident class CV documented at length.
- **Accept:** lease takeover/stale/refcount matrix; backup round-trip; the modernc-sqlite fsync/context-cancellation caveat in docs (`:memory:` for tests).

### D4 — Atomic file write (`appkit/fs` or endorse go-atomic-write)

- **What:** promote/wrap `go-atomic-write` (already family-standard) with the version floor documented: ≥ v0.5.1 — earlier versions reference `syscall.ERROR_SHARING_VIOLATION` and break every Windows build transitively (CV's documented release trap).
- **Why:** CV writes ledgers, patches, session exports, eval summaries — all through it. Batteries: one call, correct perms (0600 option), refuse-overwrite option.

### D5 — Polite outbound HTTP client (`appkit/polite`)

- **What:** a client kit for crawling/API-polling services: per-host pacing (`HostPacer`: min interval per host, cross-host never waits, Retry-After honored WITHOUT stacking pacing), an ETag-revalidation response cache (GET+ETag ≤ N bytes cached, `If-None-Match` → 304 replay, POST never cached, total byte budget, replacement semantics), SSRF guard (host allowlist, redirect rejection, timeouts, streaming body-size caps that fail as typed oversize errors — never silent truncation), and typed error classification (`ScanErrorKind`: http/body-limit/decode/network + `Retryable`/`IsDead` predicates).
- **Why:** CV's scanner layer is exactly this, battle-tested against real portals (ETag support live-verified against Greenhouse/Ashby/Lever; body-limit sized to Ashby's dual-description payloads). ANY family service doing outbound fetching (PapDashboard notification sources, job portals, webhooks verification) inherits the politeness/SSRF posture for free.
- **Accept:** the CV pin list ports directly: revalidation, refetch-on-ETag-change, POST exclusion, cross-client cache survival, same-host spacing, byte budget, replacement.

### D6 — Best-effort webhook sender

- **What:** `appkit/webhook`: transition-triggered JSON pushes — bounded in-flight goroutines (a slow receiver can NEVER block the caller), per-attempt timeout, no-retry-by-default, optional HMAC signature header, `WithPublicMode`-style payload masking hook.
- **Why:** go-health-dashboard v0.5.0 shipped one internally (`WithWebhook`, best-effort 10s, bounded in-flight — the exact right semantics). Extract to the framework so otel alert-fanout, cqrs DLQ notifications, and PapDashboard deliveries share one implementation.

---

## 6. Battery cluster E — Testing toolkit (`appkit/testkit`)

appkit's `integration/` module tests the framework cross-repo; consumers need the same caliber of harness for THEIR services. CV built all of these by hand:

### E1 — Full-chain harness + the "mux ≠ handler" trap as API

- **What:** `testkit.Serve(t, svc)` returning BOTH the raw-mux URL and the full-chain URL, with teardown that (a) closes connections, (b) runs the shutdown chain LIFO, (c) asserts goroutine baseline — and docs stating loudly: raw-mux registrations (`HandleFunc`) are INVISIBLE to route discovery and skip stdlib middleware; tests that claim "full server" must use the full handler.
- **Why:** that distinction hid three production bugs in CV (see §7) and the goroutine leak discipline (found ~25 leaked goroutines per test server) is what made the SSE journey suite stable.

### E2 — SSE test client

- **What:** `testkit.SSEClient(url)` — read frames with deadline, assert event names, resume with Last-Event-ID, count drops; plus the flake-triage trio as helpers: instrumented byte counters, goroutine dump at failure (`pprof.Lookup("goroutine").WriteTo(f, 2)`), and an `eventually` poll helper (CV's `eventually` pattern vs fixed sleeps).
- **Why:** CV's Flake Triage Protocol exists because SSE tests flake differently; the helpers are the protocol in code.

### E3 — Rate-limit isolation helper

- **What:** `testkit.UniqueXFFClient()` — per-test client identity via unique `X-Forwarded-For` so tests never share a limiter bucket (httptest clients all appear as 127.0.0.1 otherwise).
- **Why:** CV learned this when two tests poisoned each other's buckets; with A3 batteries this becomes mandatory hygiene.

### E4 — Chain-order snapshot

- **What:** `testkit.AssertMiddlewareOrder(t, svc, []string{"recovery", "requestid", ...})` — snapshot the built chain (core knows the order; expose it) and diff.
- **Why:** middleware ORDER is where the security properties live (A2's bypass AFTER rate limiting; B7's deadline OUTSIDE compression); today only CV's integration tests pin ordering implicitly.

---

## 7. Battery cluster F — Config & DX

### F1 — `appkit/config` (koanf module)

- **What:** koanf load (defaults-YAML const → file → env overrides) with the family's hard-won knobs documented IN the type: `WeaklyTypedInput` coercion traps (use in-range invalid values, not out-of-range), unsigned types for numerics at the boundary, empty-string→nil `*string` hook — plus a **discoverability test helper**: assert every env override a feature documents is actually reachable by the pull-model loader (koanf only sees registered keys; CV pins this per feature because silent env overrides cost a debugging session each).
- **Why:** every family service uses koanf the same way and re-trips the same traps.

### F2 — Request-ID log correlation in core

- **What:** seed the request ID into the handler-scoped logger so completion lines correlate without the otel module (the otel module's `TraceHandler` already does trace-ID correlation — this is the cheap core version).
- **Why:** [appkit-TODO] already tracks "httputil Logging emit with request context" for span correlation; the request-ID-only variant is the same item at lower cost. Endorse + widen.

### F3 — README: build-from-source notes

- **What:** document `GOEXPERIMENT=jsonv2` (older toolchains) and the per-module reasons in the user-facing README — [appkit-TODO already tracks this]; endorsing with consumer evidence: CV's gopls produces FALSE "string literal not terminated" errors on json/v2 imports when the LSP env lacks the experiment, which cost a misdiagnosed "file corruption" session. The README note prevents the consumer-side confusion, not just the build failure.

---

## 8. Battery cluster G — Core hygiene (the short list where CORE must move)

### G1 — TLS support in core — **endorse; demand is now real**

[appkit-TODO] already records "core has NO TLS support" with PapDashboard's `PAP_TLS_CERT/KEY` as first concrete demand. Add CV's: the SystemNix deployment terminates TLS upstream today, but the day a CV-family service deploys WITHOUT a reverse proxy, `ServiceConfig.TLS{CertFile, KeyFile}` (or ACME) is the blocker. `http.Server` TLS fields are deliberate zero-values today (`service.go:67`) — the seam exists. **Acceptance:** `RunTLS`, graceful drain identical to HTTP path, docs for cert rotation.

### G2 — Prometheus metrics surface in core (opt-in, tiny)

- **What:** `ServiceConfig.Metrics = &MetricsConfig{Path: "/metrics"}` → request duration histogram by route pattern, in-flight gauge, response total by code, build info — WITHOUT the otel module (many services want Prometheus but not tracing; CV's `/metrics` is pure Prometheus and Gatus/Alerter feed on it).
- **Why:** the otel module's metered path (~27µs/req) is the right tool when tracing is on; a no-otel ~µs-scale counter path covers the rest. Health-dashboard already exposes its own `/health/metrics` when enabled — core-level metrics make the pattern uniform.
- **Accept:** unit-1 label trap documented (CV/AGENTS: OTEL Prometheus exporter appends `_ratio` to unit-1 metrics — a real footgun worth encoding in the helper).

### G3 — Shutdown observability

- **What:** log each shutdown phase with duration (ready-flip, drain wait, drain hooks, listener close, shutdown hooks, errors joined) at INFO.
- **Why:** CV diagnoses drain behavior from logs during deploys; phases today are only visible in code. ~20 LOC, pure win.

### G4 — Endorse existing tracked items (no new work proposed, just priority signal)

1. **Licensing decision** — pkg.go.dev hides ALL godoc of the core module today (`License: UNKNOWN`) and submodule pages 404. For a framework courting consumers this is the highest-leverage non-code fix on the board. Every external adopter (and AI agent doing library research!) hits the hidden-docs wall first. **[appkit-TODO P2, USER GATE]**
2. **Mechanical API-break check at tag time** — endorsed; CV's dep-wave experience says version drift between family repos is the #1 integration cost, and a silent API break is the worst variant.
3. **v1.0.0 exit criteria** — the draft exists; this doc's §2–6 batteries are the practical definition of "API we can commit to": security cluster, httpx, testkit, worker, sqlite, polite. Until those stabilize, v1.0.0 would freeze a half-battery.
4. **DrainDelay sweep in satellites** — endorsed (`NoDrainDelay` misuse is a hidden test tax).
5. **Logging posture decision** — the +30µs/emit benchmark is real; CV's stance: keep pretty logging in dev, sample or demote in production docs, and never log per-request at INFO by default (CV logs at DEBUG for request lines; access logs live in the reverse proxy).

---

## 9. Anti-recommendations (what NOT to build — scoped so the batteries don't rot)

1. **No DI container in core.** The `do` bridge (D2) is the ceiling; core stays stdlib-composable.
2. **No templ/UI components in core.** errorpages already bridges templ-components the right way — as a module.
3. **No ORM/DB layer beyond D3's operational kit.** Storage ENGINES stay consumer choice; appkit owns operational discipline only.
4. **No security cluster in the DEFAULT stack.** Defaults stay Recovery→RequestID→Logging→(Timeout)→SecurityHeaders. CSRF/auth/rate-limit are opt-in middlewares with docs — the DEFAULTS must stay harmless for the 12-line quick start (a batteries framework whose defaults break localhost demos fails its own design center).
5. **No WebSocket in realtime.** The SSE-only constraint is stated, correct, and load-bearing — keep it.
6. **Don't promote `realtime` into core.** CV's SSE discipline (folded-broadcast ordering, drop counters) shows the complexity budget; it belongs behind the module boundary even after C1–C3.

---

## 10. Suggested sequencing (impact × effort, consumer-evidence-weighted)

| Wave | Items | Why now |
| --- | --- | --- |
| **W1 (foundations other batteries need)** | G1 TLS, G2 metrics, G3 shutdown logs, F2 request-ID logs, E1 testkit core | TLS unblocks PapDashboard hosting; metrics+logs are referenced by A3/C1; testkit is how every later battery gets verified |
| **W2 (security battery)** | A1 CSRF, A2 API-key, A3 rate limit, A4 origin, A5 CSP nonce, A6 sanitize, A7 HSTS | Largest hand-rolled surface in the family; highest re-implementation risk; pure composition of existing httputil pieces + CV's pinned tests port nearly 1:1 |
| **W3 (handler DX)** | B1–B8 as `appkit/httpx` | The CV seam consumers cannot get otherwise; B6 feeds docs-mod (B8) |
| **W4 (ops & data)** | D1 worker, D2 do-bridge, D3 sqlite kit, D4 atomic write, D5 polite client, D6 webhook | Each has a running in-family implementation to port from (CV + go-health-dashboard) — low design risk, high payoff |
| **W5 (realtime completions)** | C1–C3 | Smallest delta on an existing module; C2's race contract is the must-have |
| **Continuous** | G4 endorsements (licensing!, API-break check, v1 criteria), F3 README | Non-code, adoption-critical |

---

## 11. Traceability

- CV invariants cited: **[CV-AGENTS]** sections *Traps → Shell & Git*, *Browser & Load Testing*, *Health Probes & DI Audit*, *PipelineStore*, *Portal Scanners*, *API-Key Guard*, *Continuous Funnel*, *One-Click Apply Funnel*, *Reply Loop + Calendar Feed*, *Operator Admin Hub*.
- appkit facts verified at HEAD 2026-09-04: core go.mod (httputil v0.12.0, go-error-family v0.10.0, go-etag v0.2.0 indirect, NO samber/do), `middleware.go` default stack (5 middlewares), `service.go` drain sequence + TLS zero-values (`:67`), `realtime/hub.go` (store replay, Shutdown/Close, Health, SubscriberCount), `errorpages/README.md` (taxonomy parity with HTTPStatus), `cqrs/README.md` (EventService, DLQ), TODO_LIST (TLS gap, licensing, logging posture, API-break check, v1 criteria, DrainDelay sweep, F2 request-ID correlation).
- Cross-repo research already on file: `docs/review/2026-08-16_setup-vs-go-appkit-comparison.md` (10 findings — this doc endorses findings it overlaps: #7 logging), `docs/planning/core-v1-exit-criteria.md` (draft), `docs/planning/2026-09-04_papdashboard-integration.md` (TLS demand).
