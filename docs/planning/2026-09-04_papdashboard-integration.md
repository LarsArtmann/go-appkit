# Integration Research — PapDashboard → go-appkit

> **Date:** 2026-09-04. **Type:** Point-in-time research (re-verify before acting).
> **Question:** Can and should go-appkit integrate [`/home/lars/projects/PapDashboard`](../../../../PapDashboard) (event-sourced notification hub, published as `github.com/larsartmann/papdashboard` v0.2.0)?
>
> **Verdicts:**
>
> | Option | Verdict | One-line reason |
> | ------ | ------- | --------------- |
> | PapDashboard code as a go-appkit module | 🔴 **NO** | It is an application (cmd/server binary, SQLite, migrations, templ UI), not a framework primitive — same "wrong layer" class as the go-plugin-mvp rejection |
> | go-appkit depends on `papdashboard/sdk` | 🔴 **NO** | A framework importing an application's client SDK inverts layering for zero framework value |
> | Extract reusable pieces (notify, middleware, enum) into appkit modules | 🔴 **NO** | Nothing matches appkit's charter (thin bridges around larsartmann HTTP infra); `notify/` would be a standalone library someday, not an appkit module |
> | Reverse adoption: PapDashboard hosts on go-appkit | 🟢 **RECOMMEND** (in that repo, user-gated) | Exact cqrs-htmx ADR-001 replay; family deps already version-identical; appkit's drain phase is a capability they genuinely lack for their SSE + outbound-notify surfaces |
> | `integration/` E2E test consuming PapDashboard | 🟡 **NOT NOW — trigger-gated** | No shared contract exists to pin until reverse adoption lands; after it does, an sdk-driven smoke test mirrors the journal-replay pattern |
> | go-appkit services publish into PapDashboard via `sdk` | 🟢 **Fine, consumer-side** | Zero framework code; optional example enrichment, not a repo change |
>
> **Bottom line:** no go-appkit code changes today. The valuable integration is the *reverse direction* — PapDashboard adopting go-appkit as its host HTTP layer — executed in that repo when its current open queue (E14 server-rendered fragments, M1 Mnemosyne spike) allows. One genuine appkit-side gap surfaced: **core has no TLS support**, and PapDashboard is the first candidate consumer that would miss it.

---

## 1. Method and evidence base

Deep-dive of the candidate repo (README, AGENTS, TODO_LIST, `cmd/server/main.go`, `go.mod`, middleware wiring) plus live verification, following the cordis/go-plugin-mvp method:

| Verification | Result |
| ------------ | ------ |
| `cd PapDashboard && GOEXPERIMENT=jsonv2 go build ./... && go vet ./internal/middleware/ ./cmd/server/` | **GREEN** (exit 0 both) |
| `go list -m github.com/larsartmann/papdashboard@latest` | **v0.2.0** (published on the proxy; local tags v0.1.0, v0.2.0) |
| `PapDashboard/LICENSE` | **PROPRIETARY** — same posture as go-appkit; not a blocker between the two LarsArtmann repos |
| `grep -rn "appkit" PapDashboard/` | **Zero mentions** — no adoption plans recorded on their side |
| cqrs-htmx root `v4@v4.9.0` go.mod | **No go-appkit requirement** — PapDashboard's `cqrshtmx.Chain` usage is appkit-free even at the latest train; no coupling in either direction |
| `grep TLS` across go-appkit | **No matches** — core has no TLS support (the one adoption gap) |
| Family deps in PapDashboard go.mod | go-error-family **v0.10.0** (identical), samber/do/v2 **v2.1.0** (identical), Go **1.26.7** (identical), `GOEXPERIMENT=jsonv2` already **mandatory** there (AGENTS.md:37) |
| Importable surface | Only `pkg/enum` + `sdk/` — everything else lives under `internal/` |

## 2. Candidate profile — PapDashboard

**What it is.** A central **notification hub**: external applications publish questions, alerts, and notifications (typed SDK, REST, generic `/api/ingest`); state is event-sourced (go-cqrs-lite v4 decider pattern) over SQLite (WAL); a reactive templ/HTMX dashboard updates over SSE; outbound channels (webhook, Slack, Discord, Pushover, SMTP) fan events back out; optional LLM alert-insight enrichment and vision AI screenshot review.

| Attribute | Value |
| --------- | ----- |
| Module path | `github.com/larsartmann/papdashboard` (single module, **published v0.2.0**) |
| Go version | 1.26.7; **`GOEXPERIMENT=jsonv2` mandatory** (codec/v4) |
| Size | ~30 packages; `internal/` encapsulated; importable: `pkg/enum`, `sdk/` (typed client, sentinel errors, FTS5 search params) |
| HTTP layer | Huma v2 via `humago` adapter on `net/http.ServeMux` + plain-mux routes; `cqrshtmx.Chain(Recovery, SecurityHeaders)` + own CORS, `APIKeyAuth`, `RateLimit`, `RequestLog`, metrics middleware |
| Lifecycle (`cmd/server/main.go:369-430`) | Own `http.Server` (timeout constants), optional `ListenAndServeTLS` (`PAP_TLS_CERT/KEY`), own SIGINT/SIGTERM handling, `srv.Shutdown(ctx)` with timeout. **No drain phase** — no readiness flip, no drain window; SSE clients and outbound channels are hard-cut on shutdown |
| Health | Single custom `GET /api/health` (string check; docker-compose probes it) |
| Observability | Prometheus `/metrics` (own middleware), OTel SDK + otelhttp v0.70.0 (indirect), errorfamily exit codes at CLI boundary, RFC 9457 problem details with error-family URNs at the API boundary |
| Quality posture | Ginkgo/Gomega BDD suites, integration tests over real DI + SQLite, CI on GitHub Actions, golangci-lint 60+ linters, Nix flake |
| Own roadmap | Open items: E14 server-rendered list fragments (templ-components `Table`/`EmptyState`), M1 Mnemosyne memory sidecar spike. **Zero framework-host items** |

## 3. This repo's constraints (unchanged by this candidate)

Same bounds as the cordis/go-plugin-mvp research (`2026-09-04_cordis-and-go-plugin-mvp-integration.md` §3): consumer owns the composition root; growth pattern is thin opt-in bridge modules; dependency weight is a recorded negative; no dynamic registration surface in core; release-state discipline (otel/health wave + v1 exit criteria outrank speculative work). PapDashboard adds one new angle: it is the first candidate that is a **published, mature application** (v0.2.0 on the proxy, production-blocker backlog complete) rather than an untagged or pre-1.0 library — so the "wrong layer" rejection is even starker, and the reverse-adoption recommendation is correspondingly more actionable.

## 4. Option analysis

### Option A/B — PapDashboard code into appkit, or appkit depending on `papdashboard/sdk`

**PRO** — none meaningful. **CONTRA (dispositive):** PapDashboard is an *application* (binary + storage + UI + workers + migrations); appkit modules are framework primitives. Its importable surface (`pkg/enum`, `sdk/`) has no HTTP-framework value, and the valuable parts (`notify/` channels, `internal/middleware`, `insight/`, `a2ui/`) are both `internal/`-locked and application-shaped. This is the go-plugin-mvp "wrong layer" verdict with a cleaner example. The only durable extraction is `notify/` as a standalone outbound-notification library someday — its home would be its own repo, and it is currently coupled to PapDashboard's `internal/events` bus types.

### Option C — Reverse adoption: PapDashboard hosts on go-appkit (the real prize)

**PRO**
- **Every load-bearing version already matches** (verified in go.mod): go-error-family v0.10.0, samber/do v2.1.0, Go 1.26.7, jsonv2 toolchain. No skew to reconcile; their error-family URN mapping and exit-code handling keep working unchanged.
- **The adoption seam is exactly the cqrs-htmx ADR-001 shape.** Their `buildHandler` returns a plain `http.Handler` built on a `ServeMux`; humago mounts on any mux; handing `svc.Mux` to the same registration code is a mechanical move. Their `cmd/server/integration_test.go` tests `buildHandler` directly, so the test suite survives the swap.
- **Real capability gain, not cosmetics:** appkit's drain phase (ready-probe flip + `DrainHooks` + drain delay before connections close) is something PapDashboard lacks today — shutdown currently hard-cuts the SSE stream and in-flight outbound webhooks. appkit's `NoTimeout` posture is also the designed answer for their SSE endpoint under WriteTimeout.
- **Deletions on their side:** ~60 lines of server/signal/shutdown lifecycle (main.go:369-430), `cqrshtmx.RecoveryMiddleware` + `SecurityHeadersMiddleware` (appkit default stack: Recovery → RequestID → Logging → SecurityHeaders), `internal/middleware/logging.go` (appkit Logging), their custom `/api/health` (appkit `/health` + optionally the health module with live/ready and the dashboard). `cqrshtmx.Chain`/`HTMXScriptHandler` can stay — cqrs-htmx root carries no appkit dep, so both libraries compose without conflict.
- **Optional follow-ons, all opt-in:** appkit otel module (they currently hand-wire prometheus + otelhttp), errorpages for the non-API surface, `health.Mount` dashboard replacing nothing but adding live trend UI.

**CONTRA / gaps (concrete, resolvable)**
1. **TLS.** appkit core has no TLS support (verified: zero matches repo-wide); PapDashboard ships `PAP_TLS_CERT/KEY` → `ListenAndServeTLS`. Their DEPLOYMENT.md targets internet-facing *behind a reverse proxy*, so app-level TLS is a dev convenience — they can drop it at the proxy, **or** this becomes the first concrete consumer demand for a core TLS option (the healthy demand-driven way for appkit to grow). Decide on their side; if they ask for core TLS, that is an appkit feature request with a real consumer behind it.
2. **Timeout profile.** Their SSE endpoint under appkit's default Timeout middleware needs the SSE-safe posture (`WriteTimeout: NoTimeout` drops the Timeout middleware entirely) — supported, but it is a config decision to make consciously, not a default.
3. **It is their queue, not ours.** Their open items are E14 + M1; adoption slots behind those or jumps them by user priority. go-appkit cannot execute this — only stay compatible (it already is) and hold the door open.

**Effort estimate (their side):** ~1 session — composition-root swap, middleware re-order, health fold, docs (DEPLOYMENT.md, AGENTS tech-stack table), integration-test green. The huma/OpenAPI surface is untouched.

### Option D — `integration/` E2E test consuming PapDashboard

**PRO** — the pattern exists (journal-replay test pins published tags across repos) and PapDashboard *is* proxy-published at v0.2.0, so a pin is technically possible today.

**CONTRA** — the integration module's charter is to exercise a **shared contract** between appkit and a consumer as the proxy resolves it. Today PapDashboard resolves *without* appkit; a test would just be testing PapDashboard inside go-appkit's repo — no appkit code path involved, pure maintenance weight. **Trigger to revisit:** PapDashboard releases a version whose host layer is appkit (post Option C); then add an sdk-driven smoke test (boot service → sdk client → create/list/notification round-trip → SSE event received), mirroring `TestJournalBackedReplayThroughAppkitService`.

### Option E — go-appkit services publish into PapDashboard via `sdk`

The opposite arrow: an appkit-based service (e.g. the example service or cqrs-htmx apps) pushing lifecycle notifications into a running PapDashboard through their typed SDK (v0.2.0, `CreateNotification`, sentinel errors). Zero framework code, genuinely useful demo of the family composing end-to-end, but it is a *consumer-side* sketch — record it, do it inside `example/` only if the user wants the demo.

## 5. Decision

**No go-appkit changes.** Integrate neither the application nor its SDK into this repo. The actionable outcome is the **reverse-adoption recommendation for PapDashboard** (their repo, user-gated against their E14/M1 queue), plus one recorded appkit-side signal: if PapDashboard wants to keep app-level TLS, core gains its first TLS feature request with a real consumer attached.

**Actionable follow-ups (bounded):**

1. ~~Create this research doc~~ (done).
2. TODO_LIST P3 item with the Option C/D triggers (added below).
3. Optional, user-gated: propose Option C to PapDashboard's TODO_LIST (their repo, their gate — not executed here).

## 6. Appendix — re-entry paths

| If this happens | The integration becomes |
| --------------- | ----------------------- |
| PapDashboard's owner green-lights Option C | Their repo: swap `startHTTPServer` → `appkit.Service`, fold health, re-order middleware, decide TLS-at-proxy vs core TLS request, decide SSE timeout posture. go-appkit changes: none expected |
| PapDashboard asks for app-level TLS | First consumer-backed core TLS feature (ServiceConfig TLS option or `Listener` injection) — demand-driven, sized then |
| PapDashboard ships an appkit-hosted release | `integration/` gains an sdk-driven E2E smoke test pinning their published tag |
| A larsartmann service wants outbound notifications | Use `papdashboard/sdk` on the consumer side; if channel fan-out (Slack/Discord/email/webhook) is needed *without* PapDashboard, that is a future standalone `go-notify` library — not an appkit module |
