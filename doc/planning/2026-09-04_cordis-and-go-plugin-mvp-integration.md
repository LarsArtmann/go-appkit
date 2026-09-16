# Integration Research — cordis and go-plugin-mvp → go-appkit

> **Date:** 2026-09-04. **Type:** Point-in-time research (re-verify before acting).
> **Question:** Can and should go-appkit integrate [`/home/lars/forks/cordis`](../../../forks/cordis) (Go port of the Cordis meta-framework) and/or [`/home/lars/projects/go-plugin-mvp`](../../go-plugin-mvp) (Kernovia plugin marketplace)?
>
> **Verdicts:**
>
> | Option                                                              | Verdict                                       | One-line reason                                                                                      |
> | ------------------------------------------------------------------- | --------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
> | cordis as core's composition backbone                               | 🔴 **NO**                                     | Inverts the locked "consumer owns the composition root" decision (ADR-001, design-decisions #11)     |
> | cordis as an opt-in bridge module (`appkitcordis`)                  | 🟡 **NOT NOW — trigger-gated**                | Technically clean, but zero consumer demand + cordis is untagged and one day past its first green CI |
> | go-plugin-mvp marketplace as an appkit dependency                   | 🔴 **NO**                                     | Wrong layer (application subsystem, not framework primitive) + catastrophic weight + pre-1.0 churn   |
> | Reverse adoption: go-plugin-mvp embeds appkit as its host framework | 🟢 **RECOMMEND** (in that repo, not this one) | Exact replay of the cqrs-htmx ADR-001 pattern; net/http-native; family deps already aligned          |
>
> **Bottom line:** integrate **neither as a go-appkit dependency today**. The valuable integration is the _reverse direction_ — go-plugin-mvp adopting go-appkit as its host HTTP layer — which needs no code in go-appkit at all. Revisit a cordis bridge module only when cordis tags `go/v0.1.0` AND a real consumer asks for reactive composition.

---

## 1. Method and evidence base

Three parallel deep-dives (both candidate repos + this repo's stated direction), followed by live verification in this session:

| Verification                                                               | Result                                                                                                                                                  |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cd cordis/go && go build ./... && go vet ./... && go test ./... -count=1` | **GREEN** (single package, 0.012s)                                                                                                                      |
| `go list -m github.com/LarsArtmann/cordis/go@latest`                       | Resolves at pseudo-version `v0.0.0-20260904134728-73854d7a6229` (importable today, but **no `go/vX.Y.Z` tag exists**)                                   |
| `go list -m github.com/larsartmann/go-plugin-mvp/marketplace@latest`       | **v0.2.1** (published family live on the proxy)                                                                                                         |
| `go-plugin-mvp/LICENSE`                                                    | **PROPRIETARY** ("unauthorized copying, distribution, modification, or use … strictly prohibited")                                                      |
| `go-appkit/LICENSE`                                                        | Also **PROPRIETARY** — so license is _not_ a blocker between the two LarsArtmann repos (it only matters if/when either goes MIT for external consumers) |
| `git status` go-appkit                                                     | Clean tree; `master` is **14 commits ahead of origin** (push gate from TODO_LIST P1 still pending)                                                      |
| `git status` cordis                                                        | `main` synced with origin; two uncommitted files (`.github/workflows/ports.yml`, `go/golden_test.go`)                                                   |
| Repo docs sweep                                                            | Zero mentions of "cordis" anywhere in go-appkit; `go-plugin-mvp` appears exactly once (`integrations.md:71`, as a charmbracelet/log consumer)           |

---

## 2. Candidate profiles

### 2.1 cordis (Go port) — reactive composition meta-framework

**What it is.** The Go flagship port of [cordiverse/cordis](https://github.com/cordiverse/cordis) (TypeScript original). An application is a tree of **contexts** carrying services, event listeners, and effects. Every plugin instance runs in a **fiber**: an effect scope whose registrations (services, listeners, timers, nested plugins, cleanups) roll back LIFO on unload. Plugins declare dependencies via **inject** and stay `Pending` until their services appear; when a service is withdrawn, dependents unload and **reload in place** when it returns — "the set of running code is a pure function of the set of available services."

| Attribute     | Value                                                                                                                                                                                                                                                                                                                                                |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Module path   | `github.com/LarsArtmann/cordis/go`                                                                                                                                                                                                                                                                                                                   |
| Go version    | 1.26                                                                                                                                                                                                                                                                                                                                                 |
| Dependencies  | **ZERO** — pure stdlib (context, slog, sync, reflect, …)                                                                                                                                                                                                                                                                                             |
| License       | MIT (© Shigma, upstream)                                                                                                                                                                                                                                                                                                                             |
| Size          | 25 files / 5,132 LOC (2,673 lib + 2,440 test, ~48%)                                                                                                                                                                                                                                                                                                  |
| Public API    | Contexts (`New/Extend/Isolate/Intercept/Batch`), typed services (`Provide[T]/Get[T]/MustGet[T]`), typed + named events (5 dispatch modes: Emit/Parallel/Serial/Bail/Waterfall), fibers (6 states, `StdContext()` per activation), plugins with generics-typed config (`NewPlugin[C]`, `Start`, `Inject`), logger with slog bridge (`NewSlogHandler`) |
| Quality       | Race-tested, ~84.8% coverage (self-reported), cross-language golden tests shared with Rust/Zig ports, **first fully green CI on 2026-09-04**                                                                                                                                                                                                         |
| Release state | **Untagged** (no `go/v0.1.0`; proxy serves a branch pseudo-version). No CHANGELOG, no `.golangci.yml`                                                                                                                                                                                                                                                |
| HTTP surface  | **None whatsoever** — no net/http, no middleware, no server concept; roadmap targets loader/hmr/timer/group parity, not HTTP                                                                                                                                                                                                                         |

**Verified locally:** build + vet + test green in 0.012s.

### 2.2 go-plugin-mvp — Kernovia plugin marketplace

**What it is.** An embeddable, multi-layer **plugin marketplace for Go applications**: versioned catalog, per-app/per-tenant installations, function ACLs, signing with Fulcio keyless + Rekor transparency-log verification (offline, stdlib-only at runtime), scheduler, live SSE dashboards, Super Admin/Tenant Admin UIs (Templ/HTMX). Execution tiers (ADR-0006): WASM via Extism/Wazero (default, sandboxed) and operator-registered native Go functions (`nativert`); dynamic `.so` planned. Event-sourced on go-cqrs-lite (**v3.7.4**). Per its ADR-0004 it is the marketplace subsystem of the **Kernovia micro-kernel vision** — the marketplace _is the product_, the host plugin system is its execution engine.

| Attribute                     | Value                                                                                                                                                                                                                                                                                                                 |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Modules                       | Root `github.com/larsartmann/go-plugin-mvp` v0.3.0 + 5 published sub-modules (`marketplace`, `server`, `ui`, `sqlite`, `container`) at **v0.2.1** (v0.2.0 retracted as broken)                                                                                                                                        |
| Go version                    | 1.26.7; **`GOEXPERIMENT=jsonv2` mandatory** until Go 1.27 (guard test enforces it)                                                                                                                                                                                                                                    |
| License                       | **PROPRIETARY**, repo **private** (GitHub Actions broken on it — TODO #63; local `nix run .#verify` is the standing gate)                                                                                                                                                                                             |
| Size                          | marketplace/ alone: **218 .go files** (87 non-test); total 153 test files; 9 TinyGo WASM reference plugins                                                                                                                                                                                                            |
| Direct deps (root)            | Extism go-sdk v1.7.1, Wazero v1.12.0, Watermill v1.5.3, charmbracelet/log v1.0.0                                                                                                                                                                                                                                      |
| Direct deps (marketplace)     | templ v0.3.1020, go-cqrs-lite **/v3** v3.7.4, otel **v1.46.0**, go-error-family v0.10.0, go-branded-id v0.5.1, jsonschema/v6, ulid/v2                                                                                                                                                                                 |
| Embedding                     | `container.Build(ctx, Options) (do.Injector, error)` (samber/do **optional** — SDK core is DI-agnostic) → `srv.Routes() http.Handler` (own ServeMux) → `mux.Handle("/marketplace/", …)`; auth = host-supplied `Authorizer` (nil ⇒ AllowAll, dev-only); headless mode (`WASMDir: ""`) runs UI without plugin execution |
| Known gaps                    | Planned **module-path rename to "Kernovia"** (#60, gated on license decision #32), `internal/*` → public `pkg/` move (#59), native-tier upgrade defect (#80), GOEXPERIMENT debt (#62), no v1.0 milestone                                                                                                              |
| Family overlap with go-appkit | go-error-family v0.10.0 ✔, go-branded-id v0.5.1 ✔, charmbracelet/log v1.0.0 ✔, GOEXPERIMENT=jsonv2 ✔, Go 1.26.x ✔ — but go-cqrs-lite **v3 vs appkit's v4**, otel 1.46 vs appkit's 1.45                                                                                                                                |

---

## 3. This repo's constraints (what any integration must respect)

Locked decisions and verified facts that bound the design space:

1. **The consumer owns the composition root.** README "When NOT to use" (:194-198: want a full-stack framework → Buffalo/GoFr); design-decisions #7 ("Huma is a consumer choice, not a framework dependency"), #11/ADR-001 (appkit = generic server layer, cqrs-htmx `setup/v4` owns the one-call composition root). Anything that moves composition _into_ appkit inverts this.
2. **Thin bridge modules, opt-in, independently versioned.** The growth pattern is: new satellite module bridging a larsartmann-family library (otel, realtime, flightrecorder, flightrecorderhealth, health). Nothing heavyweight enters core; core has no DB dep (decision #4).
3. **Dependency weight is a recorded negative.** The cmdguard analysis (library-analysis archive) rejected a library partly for dragging "samber/do … far heavier than appkit's entire current footprint". samber/do is confined to flightrecorderhealth's compile-time interface assertion ("exists solely for the compile-time interface assertion — no runtime usage", AGENTS.md:134).
4. **No dynamic registration surface in core.** The extension points are config-time function slices: `OuterMiddlewares` (config.go:59), `DrainHooks` (config.go:67), `ShutdownHooks` (config.go:79), `ReadyCheck` (config.go:93), `Middlewares`/`ExtraMiddlewares` (config.go:52). No event bus, no subscribe API, no runtime plugin registry. TODO_LIST has **zero** plugin/DI/marketplace items; the only composition-root discussion defers even go-cqrs-lite's own `system/v4` behind a trigger (design-decisions #2, T3).
5. **Release state discipline.** v1.0.0 exit criteria for core are still unwritten (TODO_LIST P3); the 4-tag wave push is a standing user gate (master is 14 commits ahead). Adding dependencies now front-runs the v1-shaping work.

---

## 4. Option analysis — PRO / CONTRA

### Option A — cordis as core's composition backbone

Replace `ServiceConfig` wiring with a reactive cordis context: services/plugins auto-activate as dependencies appear; fibers roll back on shutdown.

**PRO**

- cordis is the highest-quality zero-dep candidate imaginable for this role: stdlib-only, MIT, race-tested, golden-tested, generics-typed config, single-mutex concurrency design, slog bridge that maps 1:1 onto appkit's charmbracelet/slog logger.
- Would give appkit a story no competitor has: _temporal_ service composition (unplug the DB → dependents unload cleanly → replug → dependents reload).
- No GOEXPERIMENT/toolchain burden (plain `encoding/json`), no version skew with any existing dep.

**CONTRA**

- **Philosophy inversion.** cordis _is_ a composition-root framework; embedding it as core's backbone would make appkit own what ADR-001/design-decisions #11 deliberately assign to the consumer. cqrs-htmx's `setup/v4` — the reference consumer — would have to be rewritten around fibers for zero requested benefit.
- **Semantics mismatch.** appkit's lifecycle is static and one-shot (config → start → drain → shutdown); cordis's value is _reactive reload_. Mapping fibers onto `DrainHooks`/`ShutdownHooks` captures the least valuable 20% of cordis while adding a second lifecycle model that consumers must reconcile with the first.
- **Maturity asymmetry.** appkit advertises "production-ready" and targets v1.0.0; cordis' Go port had its **first green CI today**, is untagged (pseudo-version pins carry no stability promise), has no CHANGELOG/lint config, self-reports a ~13.5% coverage gap, and its own roadmap (loader/hmr/timer/group) doesn't include integration work.
- **No demand.** No go-appkit consumer (cqrs-htmx, examples) uses cordis; no TODO/FEATURES/ROADMAP item wants it.

### Option B — cordis as an opt-in satellite bridge module (`appkitcordis`)

What a bridge would actually do (sketch, ~300-500 LOC): create the cordis root context before `svc.Start`; derive fiber goroutines from `fiber.StdContext()`; dispose the root fiber in a `DrainHook` (so rollback happens during drain, before connections close); map cordis fiber states into the `health` module's checks; route cordis logs through `NewSlogHandler` into appkit's logger; register an `OuterMiddleware` exposing plugin state on request context if needed.

**PRO**

- Follows the established bridge-module pattern exactly (otel/health/flightrecorderhealth precedents); zero new third-party deps for core; independently versioned; failure-isolated (consumers never see it).
- cordis is genuinely cheap to host: no GOEXPERIMENT, no transitive deps, 0.012s test suite.
- Real (if future) consumer story: an app that wants in-process reactive composition — e.g., feature modules that activate when their backing service appears — inside a normal appkit HTTP service with unified health/shutdown.

**CONTRA**

- **Zero demand today.** Modules in this repo get born from a consumer need (otel ← telemetry wave; health ← dashboard wave). A bridge without a driver is speculative surface to maintain, test, lint, release, and document — against an explicit repo value of bounded, actionable scope.
- **Upstream release dependency.** appkit would require cordis to tag `go/v0.1.0` first (proxy-published, semver'd, CHANGELOG'd); until then only pseudo-version pins exist, which violate the repo's release hygiene (goapidiff TODO item exists precisely because of untracked API drift).
- **Crowding.** The repo's open queue is already: push gate → otel/health release wave → v1 exit criteria. Adding a speculative module now delays the items with actual consumers waiting.
- The reactive dimension doesn't compose with core's config-time model without _new_ core API (e.g., dynamic `ReadyCheck` re-evaluation), which is v1-shape churn for a hypothetical consumer.

**Trigger to revisit (all three):** (1) cordis tags `go/v0.1.0` with a CHANGELOG; (2) a real consumer (cqrs-htmx, Kernovia, or a new app) states the reactive-composition requirement; (3) core's v1.0.0 exit criteria are written and shipped.

### Option C — go-plugin-mvp marketplace as an appkit dependency

**PRO**

- Family alignment is unusually good: shared go-error-family/go-branded-id/charmbracelet/log versions, jsonv2 toolchain, Go 1.26.7 — no dependency-hell risk; a mount-glue would be thin (`mux.Handle` + health fold + error mapping, ~100-300 LOC).
- appkit's SSE hardening (`NoTimeout`, realtime's flush-through-middleware discipline) fits the marketplace's SSE dashboards.

**CONTRA — dispositive**

- **Wrong layer (archived-precedent class).** cmdguard was rejected as "CLI, not server library — wrong layer". The marketplace is an _application subsystem/product_ (event store, tenancy, billing-adjacent ACLs, admin UIs); a framework importing it drags Wazero/Extism/Watermill/templ/cqrs-lite/sqlite into **every appkit importer**, including services that want none of it. That is the exact cost-shape the cmdguard analysis called "severe".
- **Version skew it would import:** go-cqrs-lite **v3.7.4** (marketplace) vs appkit cqrs **v4**; otel 1.46 vs appkit's 1.45. Not fatal, but a bridge would pin appkit's ecosystem backwards until the marketplace migrates.
- **Churn surface:** pre-1.0 with a planned **module-path rename** (#60) and `internal→pkg` move (#59) — an appkit dependency would break twice more by design; the v0.2.0→v0.2.1 retraction cycle shows the rename risk is real.
- **Core has no runtime to host it.** Plugins are installed/configured at _deployment time_ through the marketplace's own HTTP surface; appkit's config-time slices neither need nor model that. Wiring it would demand new core API — scope creep against an empty TODO backlog.
- License/repo posture: both repos are proprietary so no _legal_ blocker, but the private repo's broken CI (TODO #63) means appkit would depend on a project whose standing gate is a local nix run, not CI.

### Option D — reverse adoption: go-plugin-mvp embeds go-appkit as its host framework

**PRO (this is the real prize)**

- **The embed path is already net/http-native.** `srv.Routes()` returns a plain `http.Handler` (its own ServeMux, prefix-configurable); the embedding guide's canonical line is `mux.Handle("/marketplace/", srv.Routes())` — a drop-in for `svc.Mux`. No adapters exist or are needed (zero chi/echo/gin anywhere in that repo).
- **Exact precedent exists and succeeded:** cqrs-htmx ADR-001 ("appkit-as-foundation", verdict ADOPT, spike-validated 2026-08-16, re-verified 2026-09-04). The marketplace would get, in one import: graceful drain (its SSE audit feed and in-flight WASM calls benefit from `DrainHooks` + ready-flip 503), health endpoints, otel middleware + trace-correlated logging (its otel_metrics currently self-wired), pretty classified error pages (both sides already speak go-error-family v0.10.0 — family→status maps identically), and the `NoTimeout` SSE-safe server posture.
- **Work is small and on their side:** a `RunWithAppkit`-style composition root in `marketplace/examples` or a short `integrations/appkit` guide; `samber/do` is already present via `container`, and `container.Options.Server.Middleware` is the designed injection point for appkit's outer middlewares; fold `/healthz` into appkit's health module.
- **Zero code in go-appkit.** If gaps surface (e.g., auth-bridge middleware ergonomics), they arrive as concrete consumer feedback — the same way cqrs-htmx drove `ReadyCheck`.

**CONTRA / caveats**

- It's a decision for the go-plugin-mvp repo (private, pre-1.0, rename pending) — not something go-appkit can execute. appkit's role is to stay compatible and maybe publish an adoption note.
- Marketplace's own `/healthz` and appkit's `/health*` must be reconciled (dedupe or prefix) — trivial but real design work on their side.

### Option E — cross-pollination insight: cordis × go-plugin-mvp (outside appkit)

Worth recording because the user owns both: Kernovia's micro-kernel vision (ADR-0004: "let users and operators compose your application at deployment time") and cordis's reactive fiber model are **natural long-term companions**. A marketplace installation is semantically a `cordis.Start` of a plugin fiber whose injected services are provided by the host; uninstalling = `Dispose` → LIFO rollback; dependency disappearance = reactive unload. If any two of these projects ever merge philosophies, it is cordis (runtime model) + go-plugin-mvp (trust/distribution layer) — with go-appkit remaining the HTTP host both sit inside. This is a Kernovia-roadmap note, not an appkit work item.

---

## 5. Decision

**Do not integrate either as a go-appkit dependency now.**

- **cordis:** park with explicit triggers (§4 Option B). Its zero-dep nature means the option never expires — the bridge is ~a weekend of work whenever the triggers fire.
- **go-plugin-mvp:** the useful direction is Option D, executed in _that_ repo; record the recommendation and stay net/http-compatible (which requires nothing — it already is).
- **Immediate go-appkit priorities are unaffected and upstream-blocking anyway:** push the 14-commit gate, ship otel/v0.1.0 + health/v0.1.0, write core's v1.0.0 exit criteria.

**Actionable follow-ups (bounded):**

1. ~~Create this research doc~~ (done).
2. TODO_LIST P3 item referencing this doc with the cordis trigger conditions (added below).
3. Optional, user-gated: propose Option D to go-plugin-mvp's TODO_LIST (their repo, their gate — not executed here).

---

## 6. Appendix — why "not now" is not "never" (cheap re-entry paths)

| If this happens                                                    | The integration becomes                                                                                                                                                                                                |
| ------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| cordis tags `go/v0.1.0` + a consumer requests reactive composition | `appkitcordis` satellite: `DrainHook` disposes root fiber, `StdContext` derives plugin goroutines, fiber states → health checks, slog bridge → charmbracelet. ~300-500 LOC, zero third-party deps                      |
| go-plugin-mvp wants a framework host                               | ADR-001 replay in their repo: `RunWithAppkit` spike in `marketplace/examples`, `container.Options.Server.Middleware` ← appkit outer middlewares, `/healthz` folded into `health` module. No go-appkit changes expected |
| Kernovia adopts a runtime model for installed plugins              | cordis fibers as the activation substrate behind `marketplace.PluginRuntime` (their seam, their decision)                                                                                                              |
| A consumer wants _both_                                            | Compose D + cordis bridge independently — neither blocks the other                                                                                                                                                     |
