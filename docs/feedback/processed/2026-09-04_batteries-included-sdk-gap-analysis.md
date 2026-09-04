# Feedback: go-appkit's path to a REAL batteries-included SDK

> **Triage (2026-09-04, same day):** claims verified against HEAD and the CV sources — all 25 port-from paths exist; the appkit facts check out (go-etag indirect-unused, `service.go` TLS seam, 5-middleware default stack, `Hub.Health`/`SubscriberCount`). One factual error found and corrected inline below (G3 phase order). **Shipped same day:** F3 (README gopls note), G3 (shutdown phase logging — core `CHANGELOG.md` [Unreleased]). **Routed:** W1 leftovers + W2 security → `TODO_LIST.md` P2; W3-W5 → `TODO_LIST.md` P3 (the doc stays the canonical spec — do not duplicate its sketches). Endorsed items already tracked: F2 (enriched), G1, G4.

- **From:** the CV consumer perspective (`github.com/LarsArtmann/CV` — the family's deepest hand-rolled HTTP stack, 13,876 LOC runtime layer, 51 test files, every invariant scar-dated)
- **Date:** 2026-09-04 (replaces the first draft — this version adds per-item API sketches, port-from source inventory, effort sizing, acceptance test names, and 8 previously-missed batteries)
- **Method:** every gap verified against go-appkit HEAD (core `middleware.go`/`service.go`/`config.go`/go.mod, `realtime/`, `health/`, `errorpages/`, `cqrs/`, TODO_LIST) and against CV's production code. Nothing proposes a feature that exists; where TODO_LIST tracks an item, this doc marks it **ENDORSED** and adds consumer evidence.

---

## How to read this

- **Port-from** = the CV implementation to lift. These are running, production-scarred code paths with existing tests — the design work is done; the work is extraction + generalization.
- **Effort** = S (<150 LOC), M (150–500), L (500+), measured against the port-from LOC plus module scaffolding.
- **Acceptance** = the tests that must exist when the battery ships. Names in backticks are CV's actual tests — port them; they encode the failure that motivated the battery.
- **Status** = NEW (nothing tracked appkit-side) or ENDORSED (appkit TODO_LIST already tracks it; this adds demand signal).

### The thesis (one paragraph)

"Batteries included" does not mean a fat core. It means: **when a family service needs a security posture, an SSE contract, a worker loop, or a SQLite lease, the answer is `go get github.com/larsartmann/go-appkit/<module>` — not a fresh 400-LOC hand-roll.** CV is the existence proof of the demand: it hand-rolled every battery below while appkit existed, because each battery shipped zero minutes too early. The module bay is already built (`health`, `otel`, `realtime`, `errorpages`, `cqrs`, `docs`, `flightrecorder`) — this doc fills the bay.

---

## Top 7 quick wins (port ≈ 1:1, tests already exist)

| # | Battery | Port-from | LOC | Why first |
| --- | --- | --- | --- | --- |
| 1 | **A2 API-key auth + CSRF bypass** | `internal/middleware/apikey.go` + `csrf_apikey_bypass.go` | 68+40 | Every scripted/timer client in the family needs it today |
| 2 | **A3 Rate-limit profiles** | `internal/middleware/ratelimit.go` | 57 | Protects appkit's own `/health/sse` from pollers |
| 3 | **A4 Origin check** | `internal/middleware/origin_check.go` | 109 | Defense-in-depth one-file middleware |
| 4 | **A8 Body limit** | `internal/middleware/body_limit.go` | 26 | Smallest battery in the doc |
| 5 | **A6 Sanitization** | `internal/sanitization/sanitization.go` | ~60 | Text + URL (with the `&not=`→`¬=` bluemonday trap) |
| 6 | **B4 Conditional GET** | `internal/features/pipeline/handlers/interviews_ics_handler.go` | ~80 | Promotes `go-etag` (already an unused indirect dep) |
| 7 | **D7 Idempotency store** | `career-pipeline/eventstore/idempotency.go` | 165 | Self-contained; closes a whole race class |

---

## Master matrix

| ID | Battery | Module | Effort | Impact | Port-from (CV) | Status |
| --- | --- | --- | --- | --- | --- | --- |
| A1 | CSRF middleware | `security` | S | H | httputil `CSRFMiddleware` (already used by CV) | NEW |
| A2 | API-key auth + CSRF bypass | `security` | S | H | `internal/middleware/apikey.go`, `csrf_apikey_bypass.go` | NEW |
| A3 | Rate-limit profiles (keyed) | `security` | S | H | `internal/middleware/ratelimit.go` | NEW |
| A4 | Origin check | `security` | S | M | `internal/middleware/origin_check.go` | NEW |
| A5 | CSP nonce + policy build | `security` | M | H | `csp_nonce.go` + `consolidated_security.go` (690) | NEW |
| A6 | Text/URL sanitization | `security` | S | M | `internal/sanitization/sanitization.go` | NEW |
| A7 | Env-tuned security headers (HSTS) | core option | S | M | `middleware_chain.go` `securityHeadersConfig` | NEW |
| A8 | Body limit | `security` | S | M | `internal/middleware/body_limit.go` (26) | NEW |
| B1 | ResultHandler family | `httpx` | M | H | `internal/middleware/result_handler.go` (452) | NEW |
| B2 | Bind + validation (integrated) | `httpx` | S | H | `ResultHandlerWithValidation[T,V]`, `httpx/validation.go` (154) | NEW |
| B3 | ResponseWriter contract helper | `httpx`/`testkit` | S | H | `internal/httpx/response_writer.go` (149) | NEW |
| B4 | Conditional GET (ETag/304) | `httpx` | S | M | ICS handler | NEW |
| B5 | Content negotiation | `httpx` | S | M | `/admin/health` pattern | NEW |
| B6 | Route introspection + golden test | `httpx`+`testkit` | M | H | `internal/server/routes_test.go` | NEW |
| B7 | Per-route write deadlines | `httpx` | S | M | `internal/middleware/write_deadline.go` | NEW |
| B8 | Route metadata → docs feed | `docs` | M | M | — | NEW |
| B9 | No-leak error responses | `httpx` | S | H | `internal/security/error_handler.go`, `secure_error_handler.go` (64) | NEW |
| C1 | SSE drop/backpressure counters | `realtime` | S | H | `/api/pipeline/sse-stats` | NEW |
| C2 | Projection→broadcast contract | `realtime`+`cqrs` | M | H | `DashboardProjection.SubscribeFolded` | NEW |
| C3 | Per-subscriber auth + filter | `realtime` | S | M | — | NEW |
| D1 | Worker supervisor + pool | `worker` | M | H | `internal/pkg/worker/` (473) | NEW |
| D2 | samber/do bridge | `do` | M | M | `internal/di/container.go` (`resolve[T]` pattern) | NEW |
| D3 | SQLite ops kit (lease/backup/ledger) | `sqlite` | L | H | `eventstore` lease + `deadportals` ledger | NEW |
| D4 | Atomic file write | `fs` | S | M | go-atomic-write ≥ v0.5.1 (Windows trap) | NEW |
| D5 | Polite outbound client | `polite` | M | H | `career-pipeline/scanner/http_politeness.go` (205) | NEW |
| D6 | Best-effort webhook sender | `webhook` | S | M | go-health-dashboard `WithWebhook` | NEW |
| D7 | Idempotency store | `cqrs` or `worker` | S | H | `career-pipeline/eventstore/idempotency.go` (165) | NEW |
| D8 | Scheduler (tick, jitter, single-flight) | `worker` | S | M | SystemNix timers are external; no CV port | NEW |
| E1 | Full-chain test harness | `testkit` | M | H | `tests/integration` harness | NEW |
| E2 | SSE test client + flake-triage helpers | `testkit` | S | H | `eventually`, `dumpFailureState` | NEW |
| E3 | Rate-limit bucket isolation | `testkit` | S | M | `pipelineJourneyClient` XFF | NEW |
| E4 | Middleware chain-order snapshot | `testkit`+core | S | M | implicit in CV integration tests | NEW |
| F1 | koanf config module | `config` | M | M | CV `internal/config` | NEW |
| F2 | Request-ID log correlation + timing | core | S | M | `internal/middleware/timing.go` (116) | ENDORSED (TODO P3) |
| F3 | ~~README build notes (json/v2)~~ | docs | S | M | — | **DONE 2026-09-04** (README JSON v2 note existed since the execution wave; gopls false-`string literal not terminated` editor note added — core CHANGELOG [Unreleased]) |
| F4 | Config hot-reload | `config` | M | M | `career-pipeline/profileagent/platform_spec.go` | NEW |
| F5 | BuildInfo battery | core | S | M | `internal/version/version.go` | NEW |
| G1 | TLS support | core | M | H | — | ENDORSED (TODO P3, PapDashboard demand) |
| G2 | Prometheus surface in core | core | M | M | CV `/metrics` | NEW |
| G3 | ~~Shutdown phase logging~~ | core | S | M | CV drain logs | **DONE 2026-09-04** (per-phase INFO lines + total line; core CHANGELOG [Unreleased]) |
| G4 | Licensing + API-break check + v1 criteria | process | S | H | — | ENDORSED (TODO P1/P2) |

---

# Cluster A — Security (`appkit/security`, new module)

The single largest body of code family services re-implement, and the only kind where a naive re-implementation is actively dangerous. None of it belongs in the default stack (see anti-recommendations) — all of it belongs behind one `go get`.

## A1 — CSRF middleware · S · NEW

**Problem:** every family service with browser forms needs double-submit CSRF; the subtle half (a `Trusted-Origin` header does **not** bypass nosurf's token check — empirically verified CV 2026-08-29) is a trap each consumer re-discovers.

**API sketch:**

```go
csrf := security.CSRF(security.CSRFConfig{
    TrustedOrigins: []string{"http://cv.home.lan:8088"}, // "*" must log-and-degrade, NEVER mean allow-all
    Secure:         true,
})
```

**Accept:** same-site form POST passes; cross-origin 403; `*`-in-allowlist logs the misconfiguration and degrades (CV pins this — `httputil` CSRF rejects `*` in `TrustedOrigins` and only logs); skip-composition test with A2.

## A2 — API-key auth + CSRF bypass · S (68+40 LOC port) · NEW

**Problem:** machine clients (timers, scripts, calendar clients) cannot do the browser cookie dance; query-string keys leak into access logs.

**API sketch:**

```go
// guard: header X-API-Key (SHA-256 + constant-time); ?key= fallback for GET/HEAD ONLY
mux.Handle("/api/pipeline/scan", security.APIKeyAuth(key)(scanHandler))

// wrap the CSRF middleware: header-bearing requests skip the token dance,
// key still validated downstream — fail-closed, never auth-free
chain := security.APIKeyCSRFBypass(csrf)(next)
```

**Port-from:** `internal/middleware/apikey.go`, `internal/middleware/csrf_apikey_bypass.go` (the doc comment on `APIKeyCSRFBypass` is the threat model in 15 lines — lift it verbatim).

**Accept:** port CV's four pins: `TestAPIKeyAuth_QueryKeyRejectedOnPost`, header-wins-over-query, `TestAPIKeyAuth_AcceptsCorrectKey` (guards the missing-`Next()` empty-200 trap), and the bypass-still-validates test.

## A3 — Rate-limit profiles · S (57 LOC port) · NEW

**Problem:** unprotected endpoints die under polling dashboards; the short-circuit contract (a 429 must abort the chain) is where naive implementations fail — CV's 2026-08-16 bug made every limit bypassable (later handlers overwrote 429 with 200).

**API sketch:**

```go
// profiles as presets; Key pluggable (ClientIP honoring X-Forwarded-For)
analysis := security.RateLimit(security.AnalysisProfile) // 10/min, burst 15
export   := security.RateLimit(security.ExportProfile)   // 5/min, burst 8
custom   := security.RateLimit(security.RateLimitConfig{Limit: 30, Burst: 60, MaxKeys: 10_000})
```

`MaxKeys` is not optional — CV caps key-space per profile (5–10k) as an anti-memory-exhaustion bound; a keyed limiter without a key cap is a DoS vector.

**Accept:** 429-not-overwritten regression (port `internal/httpx/adapters_regression_test.go`); `Retry-After` header; shared-bucket-under-XFF semantics documented (see E3).

## A4 — Origin check · S (109 LOC port) · NEW

**API sketch:**

```go
chain := security.OriginCheck([]string{"http://cv.home.lan"})(next) // Origin→Referer fallback, 403 on miss
```

**Accept:** port `origin_check_test.go` + the integration variant.

## A5 — CSP nonce infrastructure + policy builder · M (690 LOC port) · NEW

**Problem:** the family's most expensive CSP lesson — DataStar/Alpine-style clients compile `data-*` expressions via `new Function()`: dead under `script-src 'self'`, and `script-src *` does **not** imply eval. A whole dashboard class died on this (CV 2026-08-16).

**API sketch:**

```go
// mint + context + policy assembly, mirroring CV's proven surface:
nonce, ctx := security.NewNonce(r.Context())
policy := security.BuildCSP(security.CSPConfig{
    Environment: security.Production, // script-src 'self' + 'nonce-…'; unsafe-eval NEVER, any env
    Nonce:       nonce,
    StyleInline: true, // style-src 'unsafe-inline' stays grantable BY DECISION (dynamic inline styles can't be hashed)
})
// component libraries consume the extractor contract (go-health-dashboard already takes WithNonceExtractor)
```

**Port-from:** `internal/middleware/csp_nonce.go` (`GenerateNonce`, `WithNonce`, `NonceFromContext`, `AddNonceToScriptSrc`) + `consolidated_security.go` (`CreateConsolidatedSecurityConfig`, `buildConsolidatedCSP`, `ConsolidatedSecurityMiddleware`).

**Accept:** port `TestCSPProductionBlocksInlineScripts`, `TestCSPHeaderParsesForEveryEnvironment` (validity-checked tokens, sorted/deterministic order), JSON-LD exemption test. Document the `*`-does-not-imply-eval semantics in the module README.

## A6 — Text/URL sanitization · S (~60 LOC port) · NEW

**API sketch:**

```go
clean := security.SanitizeText(userInput)          // bluemonday strict + control-char strip
u, ok := security.SanitizeURL(rawURL)              // parse → http/https + host; NO bluemonday
```

**Why the URL variant exists:** `Text()` entity-rewrites query strings (`&not=` → `¬=` — verified live in CV, 2026-09-04). One battery, two functions, trap included.

**Accept:** the `&not=` regression test.

## A7 — Environment-tuned security headers · S · NEW

```go
appkit.SecurityHeadersConfig{Environment: appkit.Production} // → HSTS max-age=63072000; includeSubDomains; preload
```

Dev/staging must leave HSTS off (LAN-over-http lockout otherwise). Today `DefaultSecurityHeadersConfig()` is the only shape.

## A8 — Body limit · S (26 LOC port) · NEW

```go
chain := security.BodyLimit(2 << 20)(next) // oversize fails as a typed error, never silent truncation
```

---

# Cluster B — Handler DX (`appkit/httpx`, new module)

appkit core speaks `http.HandlerFunc` over stdlib mux; every richer behavior is hand-rolled per service. CV's `internal/httpx` (2,752 LOC) is the port-ready proof. Depends on httputil + go-error-family only — core untouched.

## B1 — ResultHandler family · M (452 LOC port) · NEW

**Problem:** the "return `(T, error)`, get a correct HTTP response" seam is the highest-DX battery in existence for this family, and its error mapping must share errorpages' taxonomy or HTML/JSON error surfaces drift.

**API sketch:**

```go
// typed handler → auto-serialized response; errors map through go-error-family
// (Rejection→400, Conflict→409, Transient→503, Corruption→500 — same taxonomy errorpages speaks)
mux.Handle("GET /api/things/{id}", httpx.ResultHandler(func(c *httpx.Context) (*Thing, error) {
    return svc.Get(c.Request.Context(), c.PathValue("id"))
}))

// with bind + validation (B2 integrated, CV ships exactly this):
mux.Handle("POST /api/things", httpx.ResultHandlerWithValidation(
    func(c *httpx.Context, in CreateThingRequest) (*Thing, error) { return svc.Create(c.Request.Context(), in) },
))
```

**Accept:** classification-parity test vs `errorpages` (CV's `error_classification_parity_test.go` is the oracle — one taxonomy, two renderings); fuzz the response path (CV's `result_handler_fuzz_test.go`).

## B2 — Bind + validation · S (154 LOC port) · NEW

Folded into B1's family (above). Rules to encode in docs + tests: bind **tagged adapter types**, never untagged domain structs (json/v2 matches case-sensitively — CV's `{"appId":...}` 400 bug); wire tags snake_case (`tagliatelle` stance).

## B3 — ResponseWriter wrapper contract as CODE · S (149 LOC port) · NEW

**Problem:** wrappers that hide `http.Flusher` compile clean and pass `httptest.Recorder` tests, then stall SSE in production (CV's notFoundInterceptor: frames sat in net/http's 4KB buffer until client timeout).

**API sketch:**

```go
w := httpx.WrapWriter(next)          // forwards Flusher, Hijacker, Unwrap, ResponseController
httpxtest.AssertForwardsInterfaces(t, w) // fails any wrapper that hides optional interfaces
```

**Accept:** `AssertForwardsInterfaces` must fail on the buggy notFoundInterceptor shape (that's the discrimination proof).

## B4 — Conditional GET battery · S (~80 LOC port) · NEW

**Problem:** calendar clients/pollers fetch whole feeds; `go-etag v0.2.0` sits in core's go.mod as an unused indirect dep.

**API sketch:**

```go
mux.Handle("GET /feed", httpx.Conditional(feedHandler, httpx.ETagFromSha256()))
// strong sha256 ETag + Last-Modified; RFC 9110: weak W/ forms, comma lists, * → real 304s
```

**Accept:** port CV's `TestInterviewsICSHandler_ConditionalGET` matrix (weak/comma/`*`/If-Modified-Since); byte-determinism prerequisite documented.

## B5 — Content negotiation · S · NEW

```go
switch httpx.Negotiate(r, "text/html", "application/json") { ... }
```

Two in-family implementations already (CV `/admin/health`, go-health-dashboard). One battery.

## B6 — Route introspection + golden-route test · M · NEW

**Problem:** accidental route additions/removals are invisible until production; CV's golden (115 routes, set AND count) has caught multiple.

**API sketch:**

```go
for _, rt := range svc.Routes() { /* Method, Pattern, Name, Meta */ }
httpxtest.AssertRouteSet(t, svc, "testdata/routes.golden")
```

Stdlib-mux 1.22+ patterns make this introspectable if registration flows through one seam (which B8 needs anyway).

## B7 — Per-route write deadlines · S · NEW

```go
// extend the write deadline for matching prefixes ONLY (LLM missions, slow exports);
// must sit OUTSIDE compression — gzip wrapper does not Unwrap, ResponseController can't pass through
chain := httpx.WriteDeadline(3*time.Minute, "/api/agent/")(next)
```

**Accept:** the ordering regression (deadline through a gzip wrapper = fail). `NoTimeout` (today's only tool) is too blunt — the rest of the surface should keep deadlines.

## B8 — Route metadata → docs feed · M · NEW

If B6's seam carries optional metadata (summary, types), `docs-mod` generates from the LIVE route table — docs cannot drift from routes. Removes the parallel-catalog pass.

## B9 — No-leak error responses · S (port ~64+error_handler) · NEW

**Problem:** error bodies leaking internals (paths, SQL, stack) are a standing LAN-surface risk; CV has BOTH a middleware (`secure_error_handler.go`) and `SanitizeErrorMessage(err, cfg)`.

**Accept:** leaked-detail matrix test (path fragments, hostnames, SQL keywords scrubbed); JSON contract matches errorpages' shape — one error rendering story across the module family.

---

# Cluster C — Realtime completions (existing module, three additions)

## C1 — Drop/backpressure observability · S · NEW

**Problem:** SSE staleness is invisible until a user complains; CV built `/api/pipeline/sse-stats` (`droppedStoreNotifications`, `droppedBroadcasts`, …) and Gatus alerts on it.

```go
health := hub.Health() // extend sse.BroadcasterHealth with drop counts + buffer saturation
```

Document the invariant: a dropped notification only delays — the next delivered broadcast re-syncs (full-state payloads).

## C2 — The projection→broadcast contract · M · NEW — **the must-have of this cluster**

**Problem:** broadcasting on raw-event notification races the projection's own fold → clients receive pre-fold state and never see the update (no follow-up broadcast exists). CV's 4-session flake hunt (2026-08-16) ended at `SubscribeFolded`: emit only after `Handle(evt)` returned.

```go
// helper that owns the ordering guarantee so consumers can't get it wrong:
realtime.FoldedSource(projection.SubscribeFolded, func() any { return projection.Snapshot() }, hub.Broadcast)
```

**Accept:** port CV's discriminator — a wrapper store appends mid-fold; the broadcast must provably include its trigger (`TestProjectionIntegration_SubscribeFoldedReflectsEvent`). Without this wiring shipped as code, every cqrs+realtime consumer re-runs the race.

## C3 — Per-subscriber auth + filtering · S · NEW

```go
h := realtime.NewHandler(hub,
    realtime.WithAuthorize(func(r *http.Request) error { return keyGuard(r) }),
    realtime.WithFilter(func(evt sse.Event, r *http.Request) bool { ... }),
)
```

CV's admin/pipeline SSE rides the API-key guard; today a consumer must wrap the whole handler and reimplement replay/dedup to filter.

---

# Cluster D — Operations & data

## D1 — Worker supervisor + typed pool · M (473 LOC port) · NEW

**Problem:** every family service hand-rolls run/stop for background loops; CV's set: projection fold loop, reply-loop poller, boot-warm goroutine (had to be explicitly joined in `Shutdown` — leak class), PDF pool.

**API sketch:**

```go
sup := worker.NewSupervisor(svc.Logger)
sup.Go("reply-poller", worker.Config{Restart: worker.BackoffJitter()}, pollLoop) // panic → restart+backoff
sup.Tick("funnel-scan", worker.Every(6*time.Hour, worker.WithJitter(30*time.Second)), scanOnce) // single-flight: overlap skipped, counted
err := sup.Stop(ctx) // cancel → join with per-worker timeout; started-during-stop must NOT leak
```

**Port-from:** `internal/pkg/worker/` (`Manager`, `RegisterPool`, `StopAll(timeout)`, `WorkerPool[T]`, `Task/TaskResult`) — CV's manager is global-singleton shaped; generalize to instance-shaped.

**Accept:** start/stop ×100 goroutine-baseline stable; panic-restart counts; start-during-shutdown joins (CV's boot-warm lesson as a test).

## D2 — samber/do bridge · M · NEW

Core is do-free (verified) and stays so — but every family composition root is do-based, and without a bridge, appkit's lifecycle and the DI lifecycle are two shutdown owners over one process (CV's split-brain class). ~300 LOC module:

```go
doappkit.Register(injector, svc)   // *Service as do service: Shutdown cascades, Healthcheck reflects drain state
// plus the error-propagating resolve helper CV standardized after banning do.MustInvoke repo-wide:
svc, err := doappkit.Resolve[*MyService](injector)
```

## D3 — SQLite ops kit · L · NEW

Four batteries in one module (the cqrs module covers event-sourced storage; this covers everything else):

```go
db, err := sqlite.Open("app.db", sqlite.WithLease("poller"), sqlite.WithPRAGMAs(sqlite.Defaults))
// lease: hostname+pid file, Signal(0) stale-detect, reentrant refcount → second process fails FAST
// with a typed *LeaseHeldError naming owner pid/host + fix hint (CV ended the stale-reader class with this)
err = sqlite.Backup(ctx, db, "backup.db")      // VACUUM INTO: consistent snapshot, safe beside a writer, refuse-overwrite
led, err := sqlite.OpenLedger("dead.json", sqlite.WithCap(50), sqlite.WithPerms(0o600)) // atomic JSON, cap+dedupe
```

Plus the discipline docs CV paid for: `SetMaxOpenConns(1)` REQUIRED for PRAGMA persistence; modernc sqlite ignores ctx cancellation under fsync contention → `:memory:` for tests; WAL trade-offs.

**Accept:** lease matrix (takeover/stale/foreign-host/refcount); backup round-trip; ledger atomicity under concurrent writers.

## D4 — Atomic file write · S · NEW

Wrap/promote go-atomic-write with the floor documented in code comments: **≥ v0.5.1** — earlier versions reference `syscall.ERROR_SHARING_VIOLATION` (undefined on Windows) and break every `GOOS=windows` build transitively. CV hit this as a release-blocking trap.

## D5 — Polite outbound client · M (205 LOC port) · NEW

**Problem:** any service doing outbound fetching needs pacing, caching, SSRF posture, and typed failure classification — CV's scanner layer is the battle-tested implementation (ETag behavior live-verified against Greenhouse/Ashby/Lever).

```go
client := polite.New(polite.Config{
    MinIntervalPerHost: time.Second,     // cross-host never waits for each other
    ResponseCache:      polite.Cache64MiB, // GET+ETag ≤4MiB entries; If-None-Match → 304 replay; POST never cached
    AllowHosts:         []string{"api.lever.co"}, // SSRF: allowlist + redirect rejection
    BodyLimit:          32 << 20,        // streaming cap; oversize = typed ErrBodyLimit, never truncation
})
cls := polite.Classify(err) // http/body-limit/decode/network + Retryable/IsDead predicates
```

**Accept:** port `http_politeness_test.go` pins wholesale (revalidation, refetch-on-ETag-change, POST exclusion, cross-client cache survival, same-host spacing, byte budget, replacement semantics).

## D6 — Best-effort webhook sender · S · NEW

Extract from go-health-dashboard v0.5.0's `WithWebhook` (right semantics already shipped there): bounded in-flight (slow receiver NEVER blocks the push loop), per-attempt timeout, no-retry default, optional HMAC header.

## D7 — Idempotency store · S (165 LOC port) · NEW

**Problem:** at-most-once work under notification races (projections, webhooks, SSE bridges) — CV folds every event through `CheckAndRecord`.

```go
store := worker.NewIdempotencyStore(worker.WithTTL(time.Hour))
if store.CheckAndRecord(worker.ProcessorKey("crm-sync", evt.ID())) { /* first delivery */ }
store.Close() // owns its evict goroutine — Close() must JOIN it (CV's evictLoop leak, found by the teardown suite)
```

**Accept:** double-delivery collapse test; evict-loop shutdown joins (goroutine baseline); TTL expiry frees keys.

## D8 — Scheduler · S · NEW

See `sup.Tick` in D1 (6h tick + jitter + single-flight + missed-tick policy documented). CV runs these as SystemNix timers today — in-process scheduling is the battery for non-Nix consumers; keep it tick-shaped (NOT cron-expression parsing — see anti-recommendations).

---

# Cluster E — Testing toolkit (`appkit/testkit`)

## E1 — Full-chain harness + the mux≠handler trap as API · M · NEW

```go
srv := testkit.Serve(t, svc)   // returns FullChainURL (stdlib chain + engine) AND MuxURL
                              // teardown: close conns → shutdown chain LIFO → goroutine-baseline assert
```

Encode the lessons as API docs: raw-mux registrations are INVISIBLE to route discovery and skip the stdlib chain (three CV production bugs were only visible through the full chain); a test server that leaks ~25 goroutines makes every SSE test load-flaky (CV's teardown suite caught a real evictLoop leak on first run).

## E2 — SSE test client + flake-triage helpers · S · NEW

```go
c := testkit.SSE(t, srv.FullChainURL+"/health/sse")
c.ExpectEvent(t, "datastar-patch-elements", 5*time.Second)
testkit.Eventually(t, 10*time.Second, cond) // poll, never fixed sleeps
testkit.DumpGoroutines(t.Failed())          // pprof goroutine dump on failure — the triage protocol as code
```

## E3 — Rate-limit bucket isolation · S · NEW

```go
client := testkit.UniqueXFFClient(t) // unique X-Forwarded-For per test — httptest clients all appear as 127.0.0.1
```

Mandatory once A3 ships; CV learned this when sibling tests poisoned each other's buckets.

## E4 — Chain-order snapshot · S · NEW

```go
testkit.AssertChainOrder(t, svc, "recovery", "requestid", "logging", "securityheaders")
```

Middleware ORDER is where the security properties live (A2 bypass AFTER rate limiting; B7 deadline OUTSIDE compression). Core must expose the built chain names; the test pins them.

---

# Cluster F — Config & DX

## F1 — koanf module (`appkit/config`) · M · NEW

Defaults-YAML const → file → env overrides, with the family's traps encoded in the type system or docs: `WeaklyTypedInput` silent coercion (validate with in-range invalid values, never out-of-range overflow), unsigned numerics at boundaries, empty-string→nil `*string` hook, and the **env-override discoverability helper**:

```go
configtest.AssertEnvOverridesReachable(t, map[string]string{"CV_AGENT_API_KEY": "…"}) // koanf pull-model only sees registered keys
```

CV pins this per feature because silent env overrides each cost a debugging session.

## F2 — Request-ID log correlation + timing header · S · ENDORSED (TODO P3 tracks the httputil emit)

Seed request ID into the handler-scoped logger (cheap core version of otel's `TraceHandler` trace correlation) + `X-Response-Time` (CV `timing.go`, 116 LOC).

## F3 — README build notes · S · ENDORSED (TODO P2)

Add consumer-side evidence: CV's gopls produces FALSE "string literal not terminated" on json/v2 imports when the LSP env lacks `GOEXPERIMENT=jsonv2` — cost a misdiagnosed file-corruption session. The README note prevents confusion, not just build failures.

## F4 — Config hot-reload · M · NEW

```go
cfg, err := config.Watch(file, config.WithValidateGate(myValidate)) // fsnotify + validate → atomic swap → hook
```

Port-from: `career-pipeline/profileagent/platform_spec.go` (hot-reloadable automation YAML). Never serve a half-swapped config: validate-gate then swap atomically.

## F5 — BuildInfo · S · NEW

```go
// internal/version: linker-mutated var (GoReleaser -X), "dev" default
appkit.WithVersion(version.Version) // → /health version field + /version endpoint + build info metric (G2)
```

CV's `cv --version` stamps `main.version/commit/date`; go-health already takes `WithVersion` — one story.

---

# Cluster G — Core hygiene (the short list where CORE must move)

## G1 — TLS · M · ENDORSED (TODO P3; PapDashboard's `PAP_TLS_CERT/KEY` is first demand)

`service.go:67` deliberately zeroes `http.Server` TLS fields — the seam exists. Add `ServiceConfig.TLS{CertFile, KeyFile}` (+ ACME later); drain sequence identical to the HTTP path; cert-rotation docs.

## G2 — Prometheus surface in core (opt-in) · M · NEW

`ServiceConfig.Metrics{Path: "/metrics"}` → request-duration histogram by route pattern, in-flight gauge, response totals, build info — WITHOUT the otel module (CV's `/metrics` is pure Prometheus; Gatus feeds on it; the otel metered path (~27µs/req) stays the tool when tracing is on). Encode the unit-1 trap in the helper docs: OTEL's Prometheus exporter appends `_ratio` to unit-1 metrics — CV hit it live.

## G3 — Shutdown phase logging · S · DONE 2026-09-04

~~One INFO line per phase with duration: ready-flip → drain wait → drain hooks → listener close → shutdown hooks → joined errors.~~ Shipped 2026-09-04 with the TRUE execution order (the draft above had it wrong — drain hooks run inside the traffic window, BEFORE the wait): `ready_flip` → `drain_hooks` → `drain_wait` (or a `shutdown phase skipped` line under `NoDrainDelay`) → `listener_close` → `shutdown_hooks`, then `graceful shutdown complete` with total + `result` (`ok`/`error`). Phase names are a grep-able contract; `shutdownlog_test.go` pins the sequence. CV diagnoses deploys from exactly these logs.

## G4 — Process endorsements (no new work; priority signal)

1. **Licensing decision** — pkg.go.dev hides ALL core godoc today (`License: UNKNOWN`); submodule pages 404. Highest-leverage non-code fix on the board; blocks every external adopter AND every AI agent doing library research.
2. **Mechanical API-break check at tag time** — family dep drift is the #1 integration cost (CV's dep-wave lockstep trap); a silent API break is the worst variant.
3. **v1.0.0 exit criteria** — the draft exists; this doc's batteries ARE the practical freeze-scope. v1.0.0 before security/httpx/testkit stabilize would freeze a half-battery.
4. **DrainDelay sweep + logging posture** — both already tracked; endorsed.

---

# Anti-recommendations (keep the batteries from rotting the core)

1. **No security cluster in the DEFAULT stack.** Defaults stay Recovery→RequestID→Logging→(Timeout)→SecurityHeaders. CSRF/auth/rate-limit are opt-in with docs — a batteries framework whose defaults break the 12-line localhost demo fails its own design center.
2. **No DI container in core.** D2's bridge is the ceiling.
3. **No templ/UI in core.** errorpages already bridges templ-components correctly — as a module.
4. **No cron-expression parser.** D8 stays tick+jitter+single-flight; cron DSLs are a dependency and complexity magnet with no family consumer.
5. **No WebSocket in realtime.** SSE-only is stated, correct, load-bearing.
6. **No storage engines.** D3 owns operational discipline (lease/backup/ledger), never schema/ORM choices.
7. **Don't promote realtime/httpx into core** — CV's SSE and result-handler complexity budgets prove they belong behind module boundaries.

---

# Sequencing

| Wave | Items | Gate to next wave |
| --- | --- | --- |
| **W0 (process, parallel)** | G4 licensing, F3 README | pkg.go.dev renders all modules |
| **W1 foundations** | G1 TLS, G2 metrics, G3 shutdown logs, F2 timing, E1 testkit core, F5 buildinfo | testkit verifies every later battery |
| **W2 security** | A2, A3, A4, A8, A6 (quick wins) then A1, A5, A7 | 429-bypass + CSP-pin tests green upstream |
| **W3 handler DX** | B1, B2, B9 then B3, B4, B5, B7, then B6, B8 | classification-parity vs errorpages pinned |
| **W4 ops & data** | D7, D4, D6 (small) then D1, D5, then D3, D2, D8, F1, F4 | lease + politeness pin-tests ported |
| **W5 realtime** | C1, C3, then C2 | folded-broadcast discriminator green |

---

# Traceability

- **CV port-from paths** (all verified to exist 2026-09-04): `internal/middleware/{apikey,ratelimit,origin_check,csp_nonce,consolidated_security,body_limit,timing,secure_error_handler,result_handler,write_deadline,csrf_apikey_bypass}.go`; `internal/sanitization/sanitization.go`; `internal/security/error_handler.go`; `internal/httpx/{result,validation,response_writer,adapters}.go`; `internal/server/routes_test.go`; `internal/pkg/worker/{manager,pool}.go`; `internal/version/version.go`; `career-pipeline/scanner/http_politeness.go`; `career-pipeline/eventstore/idempotency.go`; `career-pipeline/profileagent/platform_spec.go`; `internal/features/pipeline/handlers/interviews_ics_handler.go`.
- **appkit facts verified at HEAD 2026-09-04:** core go.mod (httputil v0.12.0, go-error-family v0.10.0, go-etag v0.2.0 indirect, NO samber/do); `middleware.go` 5-middleware default stack; `service.go` drain sequence + TLS zero-fields at `:67`; `realtime/hub.go` (store replay, Shutdown/Close, Health, SubscriberCount); `errorpages/README.md` taxonomy parity; `cqrs/README.md` EventService+DLQ; TODO_LIST (TLS gap, licensing, logging posture, API-break check, v1 criteria draft, DrainDelay sweep, request-context logging).
- **Cross-repo research on file:** `docs/review/2026-08-16_setup-vs-go-appkit-comparison.md` (finding 7 = F2's logging posture); `docs/planning/core-v1-exit-criteria.md`; `docs/planning/2026-09-04_papdashboard-integration.md` (TLS demand).
