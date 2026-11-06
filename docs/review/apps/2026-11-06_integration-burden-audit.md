# Cross-App Integration Burden Audit — the 39 Go apps

**As of:** 2026-11-06 · **Scope:** every `meta`-tracked `app`-tagged project that is a Go app (39 of the 57 app-tagged; 19 non-Go excluded) · **Method:** scripted sweep of all `go.mod` requires + app-level `rg` presence scans + entrypoint deep-reads of six exemplar apps · **Companion plan:** `docs/planning/2026-11-06_01-00-SUPERB-v12-app-composition-preset-plan.html`

This is the **consumer-side mirror** of `docs/review/libs/` (2026-11-03, supply-side). That audit asked "are we using our libs to the max?" — this one asks the owner's actual question: **"why do I keep rebuilding the same integrations from my own building blocks?"** The answer is measurable.

House rules: point-in-time snapshot, annotate never rewrite; execution state lives in TODO_LIST/ROADMAP.

---

## 1. The headline paradox

First-party library adoption across the 39 Go apps (direct `go.mod` requires):

| Lib               | Apps | Lib                 | Apps | Lib                    | Apps  |
| ----------------- | ---- | ------------------- | ---- | ---------------------- | ----- |
| go-branded-id     | 34   | go-retry            | 16   | cmdguard               | 9     |
| go-error-family   | 34   | go-datastar         | 13   | go-ndjson              | 9     |
| go-etag           | 25   | go-idempotency      | 13   | go-sqlitestore         | 9     |
| go-sse            | 24   | go-atomic-write     | 11   | samber-do-auditlog     | 9     |
| httputil          | 24   | go-health           | 11   | **go-appkit (direct)** | **6** |
| go-codec          | 23   | go-output           | 11   | go-finding             | 5     |
| go-cqrs-lite      | 22   | go-health-dashboard | 10   | go-filewatcher         | 5     |
| go-flightrecorder | 22   | cqrs-htmx           | 10   | go-graph-rag           | 4     |
| templ-components  | 21   |                     |      | go-aichat              | 3     |

**The atomic building blocks are everywhere — the composition layer is nowhere.**
`go-appkit`, the repo whose stated mission is _"make it easy for my apps to smartly use all my libs together"_, is the **least-adopted core-family module (6/39 direct; 8/39 counting the two cqrs-htmx-setup transitive rides)** — below every single block it exists to compose. The same ten-lib cluster (branded-id, error-family, etag, go-sse, httputil, codec, cqrs-lite, flightrecorder, templ-components, retry) is hand-assembled, per app, ~20–30 times.

Six direct adopters: quotaboard, media-atlas, pbx-lens, PapDashboard, Rolls-Royce-mtuGoHelpCenter-golang, legal-graph-ai-system. Two transitive (via cqrs-htmx `setup`): crm, crm-stage-t13b (+webphone/dnsblockd-family riding setup elsewhere).

## 2. The duplication ledger (what gets rebuilt per app)

Presence scans, app-level (39 Go apps; test files excluded):

| Integration shape                                      | Apps | What already exists on the shelf                 |
| ------------------------------------------------------ | ---- | ------------------------------------------------ |
| Hand-rolled `internal/config` package                  | 22   | — (F1 config battery, demand-gated)              |
| Hand-parsed `os.Getenv` config structs                 | 25   | — (same)                                         |
| Embedded `CREATE TABLE` migrations (ad-hoc SQL)        | 16   | — (D3 SQLite ops kit + migration helper)         |
| Hand-opened sqlite (`sql.Open`/modernc)                | 15   | go-sqlitestore, metaengine/sqliteengine          |
| Hand-rolled JWT/session auth                           | 10   | cqrs-htmx `usermgmt` (WebAuthn/OAuth2/TOTP)      |
| Hand-rolled SSE (`text/event-stream` writers)          | 9    | go-sse 24-app fleet pin, appkit/realtime         |
| Hand-rolled graceful shutdown (`signal.NotifyContext`) | 9    | core `Service.Run/Shutdown` + DrainHooks         |
| Hand-wired charmbracelet/slog setup                    | 7    | core `InitLogger`                                |
| flake.nix                                              | 37   | — (scaffold)                                     |
| `.buildflow.yml`                                       | 20   | — (scaffold)                                     |
| Dockerfile                                             | 16   | — (scaffold)                                     |
| systemd `.service` unit                                | 6    | — (scaffold; appkit/systemd for the notify side) |

Every row is either a shipped-but-unadopted battery, a spec'd-but-unbuilt battery (battery-spec cluster letter in parentheses), or raw bootstrap copy-paste (scaffold).

## 3. Shipped-but-not-adopted — the damning exemplars

### 3.1 dnsblockd — ~1,600 LOC re-implementing three shipped appkit modules

| dnsblockd hand-roll                                                     | LOC   | Shipped equivalent                                         |
| ----------------------------------------------------------------------- | ----- | ---------------------------------------------------------- |
| `internal/otel/otel.go` + `sdnsbridge.go`                               | 777   | appkit/otel v0.3.0 (OTLP, self-diagnostics, log bridge)    |
| `internal/sdnotify/sdnotify.go`                                         | 59    | appkit/systemd v0.1.0 (`Install(&cfg)` one-call)           |
| `internal/middleware/` csrf+ratelimit+requestid+security+statusrecorder | ~590  | appkit/security v0.2.0 + httputil middleware               |
| `internal/logging/` multi-handler + sampling                            | 206   | core logging; sampling is the documented gated perf option |
| `internal/ratelimit/`                                                   | 248   | httputil keyed rate limits (security A3)                   |
| `internal/config/` config + validation                                  | 2,172 | — (F1)                                                     |

Plus a cobra+fang CLI (cmdguard exists, 9 adopters) and a hand-rolled startup-phase timer (core logs `startup phase` analog: `shutdown phase complete`). Nothing here is exotic — it is the standard battery set, rewritten.

### 3.2 ChastityAPI — ~4,800 LOC of infra packages

`internal/auth` 1,068 (cookie 53, csrf 92, device_auth 53, middleware 365, password 100, ratelimit 201, request_id 52, token 150), `internal/logger` 541, `internal/config` 409, `internal/captcha` 151, `internal/csp` 110, `internal/sse`, `internal/db` + hand-wired sqliteengine/timerengine/scheduling in `cmd/server/main.go` (JWT dev-secret gating, rate-limit consts, scheduler join choreography at lines 41–150). Near-1:1 battery overlap: security (csrf/ratelimit/request-id/CSP), usermgmt (password/token/session), realtime (sse), D1 (scheduler joins), F1 (config). Also hand-rolls `x/time/rate` buckets.

### 3.3 PapDashboard (appkit host, still hand-rolls security)

`internal/middleware/auth.go` APIKeyAuth + BodyLimit + RateLimit + CSP overlap security v0.2.0's battery set near-1:1 — already documented in the M17 ghost-API review (TODO_LIST P3). Being an appkit host did not automatically compose the batteries; each is still a separate opt-in wiring decision.

### 3.4 pbx-lens — the CLI contract, hand-rolled

`cmd/pbx-lens/main.go:19–109`: map-of-subcommands dispatch + `flag.FlagSet` + hand-rolled `fatalCode` error-family exit-code mapping (Rejection/Conflict 1, Transient 75, Infrastructure 69, Corruption 65...). That IS the cmdguard typed-flag + exit-code contract, reimplemented. Meanwhile its `serve` path uses appkit + go-sqlitestore correctly.

### 3.5 The healthy poles (what "good" looks like today)

- **quotaboard** (`main.go:80–135`): cqrs-htmx `MustNew` + `appkit.NewService` + `systemd.Install` + StartHooks/ShutdownHooks refresher. Still hand-writes: slog text handler, `config.FromEnv`, dispatcher + recovery middleware, route mounting.
- **media-atlas**: cmdguard CLI (15 verbs) + appkit + system + testkit + errorpages — but each subsystem is a hand-written `register*(service, sys, cfg)` glue function.
- **cqrs-htmx `setup`** (10 adopters): the fleet's ONLY preset — and the only composition product with real adoption. `setup.New(Config)` → dispatcher + auth routes + SSE + errorpages + appkit service in one call. **Presets get adopted; parts catalogs don't.**

## 4. Root causes (why 98% of app time is re-assembly)

1. **Timing + no swap recipes.** Most apps predate the batteries (core v0.9.0, security v0.2.0, systemd v0.1.0, otel v0.3.0 — all Aug–Oct 2026). Nothing made adopting cheaper than keeping the hand-roll: no per-app battery-swap checklist, no preset to migrate onto.
2. **No appkit-side preset.** The one composition product that works (cqrs-htmx `setup`) covers only event-sourced HTMX apps. Non-cqrs apps — and even setup apps for config/logger glue — assemble by hand.
3. **Opt-in-everything posture.** Each battery is a separate module + separate wiring decision (§3.3 proves hosting appkit ≠ composed batteries). Per-app willpower is the limiting resource; env-driven defaults are absent.
4. **No scaffold.** Every new app starts by copy-pasting a cousin's config/logger/CLI/flake/buildflow (§2's bottom four rows: 37/39/20/16/6). The fleet has website templates and an (unbuilt) wizard SDK — no service scaffold.

## 5. What "make it easy" means — the program

Full Pareto ranking, task tables, and the execution graph live in the companion plan (`docs/planning/2026-11-06_01-00-SUPERB-v12-app-composition-preset-plan.html`). Summary:

| Tier          | Lever                                                                                                                                                                                                                                                    | Evidence base                        | Destination                                 |
| ------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | ------------------------------------------- |
| **1% → 51%**  | **`appkit/preset`** — one-call, env-driven standard service composing ONLY shipped batteries (core + systemd + health dashboard + errorpages + metrics/version + `RegisterClassifications` + optional otel/realtime/security via env)                    | §1 paradox; §3.5 setup proof         | new module, `preset/`                       |
| rides T1      | **Config battery (F1, demand-proven)** — env-first typed config                                                                                                                                                                                          | 22 config packages, 25 env parsers   | elevates F1 from demand-gated               |
| **4% → 64%**  | **Scaffold `appkit new`** — generates cmdguard serve verb + preset + templ + flake + buildflow + CI + systemd + Dockerfile                                                                                                                               | §2 bottom rows; no existing scaffold | template repo or `cmd/` generator           |
| **20% → 80%** | **Adoption wave** — per-repo battery-swap recipes (dnsblockd −1,600 LOC, ChastityAPI −4,800, PapDashboard security, pbx-lens cmdguard)                                                                                                                   | §3                                   | recipes in appkit docs; work in their repos |
| **20% → 80%** | **Battery completions with fleet-proven demand** — D1 worker supervisor, D3 sqlite ops + migrations, C2 realtime `Mount`/`SubscribeFolded`, auth-bridge decision (usermgmt recipe vs security session battery — owner gate), W3 httpx (already promoted) | §2 rows 3–10                         | TODO_LIST P2/P3                             |

**Success metric for the whole program:** six months out, `meta`-inventory re-sweep shows go-appkit direct adoption > cqrs-htmx setup's 10, and the §2 ledger rows trending toward the shipped-battery column.

## Appendix A — method + raw artifacts

- Population: `meta list --tags app --format json` (55 tracked) + 2 untracked; Go classification by recursive `go.mod` presence (vendor/node_modules excluded) → 39 Go apps.
- Adoption: every `go.mod`'s `github.com/larsartmann/*` requires, aggregated per lib (raw dump: session artifact `/tmp/goapp-fp.txt`).
- Signals: app-level `rg -l` presence over non-test `.go` files + `find` for infra files.
- LOC: `wc -l` per package, test files excluded.
- Entrypoint deep-reads: dnsblockd, quotaboard, pbx-lens, ChastityAPI, media-atlas (+ CV/KeyHolderAI structural scan).
- Known limitations: presence ≠ usage depth; transitive requires (cqrs-htmx setup pulling appkit) distinguished only for the headline number; CV's 13-module tree counted once per app.

## Appendix B — the 39 Go apps

CV, ai-router-implementation, Cyberdom, dnsblockd, wise-transaction-visualization, GmbH, standard-bug-tracking-schema, desire-secrets, Code-Quality-Agent, dynamic-markdown-site, knowledge-graph, quotaboard, go-taskqueue, media-atlas, pbx-lens, crm, mr-sync, emeet-pixyd, crm-stage-t13b, webphone, nsfw-classifier, cv-stage-crmfix, file-and-image-renamer, testing, overview, PapDashboard, Rolls-Royce-mtuGoHelpCenter-golang, project-discovery-daemon, storbi, accountability-system, RedditParse, ksef-sandbox, legal-graph-ai-system, KeyHolderAI, ChastityAPI, nobletary, AutoCart, Kernovia, io-surface.
