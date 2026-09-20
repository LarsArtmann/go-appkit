# samber/do v2 × Health Checks — Deep Architecture Review

**Date:** 2026-09-20 10:58 | **Scope:** samber/do v2 utilization across go-appkit + the health/self-health architecture (core `Service`, `/health`, health, flightrecorderhealth, go-health SDK bridge)
**Method:** source-level audit of every `samber/do` call site, go-health v0.1.3/v0.2.0/v0.3.0 module sources, do-v2 v2.1.0 internals, live test runs. Every claim cites `file:line`.

---

## Verdict (TL;DR)

**Q1 — "Are we using samber/do v2 in combination with Health Checks superbly?"**
The bridge (`flightrecorderhealth`) is exemplary: DO-1…DO-6 fully clean, contract-pinned in both directions, version-current on samber/do (v2.1.0 = latest, proxy-verified), race-clean, and its eager registration is _load-bearing_ against a real do-v2 trap (F5). **But not superb overall: adoption 82/100.** The flagship combination — _trace snapshot on health-check failure_ — silently disappears on the convenience probe path (F1), the repo's own health-family versions have drifted three ways (F3), and the full self-health stack has zero CI proof (F2).

**Q2 — "Do we have a PROPER Service Oriented, composable, resilient, self-health architecture?"**
**Yes — structurally excellent: 4.57/5 ("Excellent — maintain and document").** Independent modules, clean contracts, panic-isolated checks, two-phase drain in lockstep with core, fail-closed readiness, race-green. The deductions are two P0 _capability_ gaps (silent feature loss, untested composition) — not structural flaws. Both are fixable without restructuring.

---

## 1. The Architecture As Built

samber/do appears in exactly **one** production file of this repo — deliberate and correct for a library (no DI lock-in; consumers own the composition root):

```
Layer 3  appkit composition      core.Service (readyProbe, DrainHooks/ShutdownHooks, ReadyCheck)
                                 health.Mounted (Start/Drain/Shutdown/Ready)  ← config-level wiring
Layer 2  go-health SDK           health.New(injector, WithHealthRecorder)   ← injector path
                                 NewWithHealthCheck / appkithealth.NewProbe ← injector-free path
Layer 1  samber/do v2 injector   services expose do.HealthcheckerWithContext
                                 injector.HealthCheckWithContext = concurrent fan-out
Bridge   flightrecorderhealth    Checkable (recorder state → do.HealthcheckerWithContext)
                                 Trigger (health.HealthRecorder → fr snapshot on failure)
```

Production samber/do call sites (all of them):

| Site                                      | API                               | Role                                             |
| ----------------------------------------- | --------------------------------- | ------------------------------------------------ |
| `flightrecorderhealth/adapter.go:280`     | `do.ProvideNamed`                 | register `*Checkable` under a name               |
| `flightrecorderhealth/adapter.go:285`     | `do.InvokeNamed` (eager)          | **load-bearing** — see F5                        |
| `flightrecorderhealth/adapter.go:176-186` | `injector.HealthCheckWithContext` | Trigger wraps the do fan-out, adds trace capture |

Everything else in go-appkit is samber/do-free by design; go-health itself is the injector consumer (`go-health@v0.2.0/go.mod` requires `samber/do/v2 v2.1.0`; it surfaces as _indirect_ in `health/go.mod`).

### The two go-health construction paths — and the cliff between them

|                                             | Injector path                                                        | Injector-free path                                                                                                         |
| ------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Constructor                                 | `health.New(injector, …)` (probe.go:310)                             | `NewWithHealthCheck(fn, …)` (accessors.go:59) ← appkit `NewProbe` wraps this                                               |
| `WithHealthRecorder` (→ `frhealth.Trigger`) | **works** — batches delegate through the recorder (probe.go:357-367) | **silently discarded** — `cfg.recorder = nil` (accessors.go:61; v0.1.3:36, v0.3.0:61 — all published versions)             |
| Appkit-facing wrapper                       | none (consumer builds injector)                                      | `appkithealth.NewProbe` (health/probe.go:34) — forwards arbitrary Options _including_ `WithHealthRecorder` with no warning |

---

## 2. samber/do Utilization Audit (gap analysis)

| Capability                                                          | Status                              | Evidence                                                                                                                                                                                                          |
| ------------------------------------------------------------------- | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `do.HealthcheckerWithContext` lifecycle interface                   | 🟢 Fully Leveraged                  | `Checkable` implements it (adapter.go:57); compile-pinned (contract_test.go:22)                                                                                                                                   |
| `health.HealthRecorder` interception                                | 🟢 Fully Leveraged on injector path | `Trigger` (adapter.go:176-235); compile-pinned (contract_test.go:19)                                                                                                                                              |
| Named services                                                      | 🟢 Fully Leveraged                  | `do.ProvideNamed` + name defaulting (adapter.go:273-287)                                                                                                                                                          |
| Global-injector avoidance / library hygiene                         | 🟢 Fully Leveraged                  | zero `do.New` in library code; doc.go:47 teaches consumers to own it                                                                                                                                              |
| Eager-value registration (`do.ProvideNamedValue`)                   | 🟡 Partially Used                   | `Register` hand-rolls eager via lazy provider + explicit invoke (adapter.go:280-285); `do.ProvideNamedValue` is the canonical one-liner (F7)                                                                      |
| `do.Shutdowner*` lifecycle                                          | 🟡 Partially Used                   | bridge types hold no resources (N/A for them); but `Mounted` offers no injector-addressable shutdown adapter although go-health ships the pattern (`AsShutdowner` → `ProbeShutdowner`, accessors.go:139-149) (F4) |
| Probe-registered-into-injector (nested health)                      | 🟡 Available, unused                | `Probe.HealthCheck` satisfies `do.HealthcheckerWithContext` (go-health accessors.go:117-129); appkit never composes it                                                                                            |
| Injector-level options (`do.NewWithOpts`, `WithHealthCheckTimeout`) | 🟡 Available, unsurfaced            | used in go-health v0.3.0 docs (doc.go:41,138); absent from appkit docs/examples                                                                                                                                   |
| `samber-do-auditlog` hooks                                          | 🔴 Missed (minor)                   | not wired anywhere in go-appkit (F8)                                                                                                                                                                              |
| Scopes / transient services                                         | ⚪ N/A                              | no request/tenant-scoped services in a health bridge                                                                                                                                                              |
| `do.Package` grouping                                               | ⚪ N/A                              | bridge registers one service                                                                                                                                                                                      |

**Adoption score: 82/100** — full marks on everything used; deductions for the silent recorder cliff (weighted heaviest), unused lifecycle adapters, and unwired audit hooks.

---

## 3. DO-1…DO-6 Compliance (skill rubric, checked at source)

| Rule                             | Verdict        | Evidence                                                                                                            |
| -------------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------- |
| DO-1 `Must*` in runtime paths    | ✅ Clean       | zero `MustInvoke` outside tests (`rg 'do\.MustInvoke' -g '!*_test.go'` → doc comment only)                          |
| DO-2 `do.New` without Shutdown   | ✅ Clean (N/A) | library never creates an injector; `doc.go:47` quick start hands ownership to the consumer                          |
| DO-3 `Override*` outside tests   | ✅ Clean       | no occurrences                                                                                                      |
| DO-4 package-level injector      | ✅ Clean       | injector is always a parameter (`Register(injector do.Injector, …)`, `RecordHealthCheckWithContext(ctx, injector)`) |
| DO-5 `Invoke` inside loops       | ✅ Clean       | no occurrences                                                                                                      |
| DO-6 cross-service `Shutdown`    | ✅ Clean       | bridge types implement no Shutdowner; `Mounted.Shutdown` is self-contained and idempotent (mount.go:215-228)        |
| Service-locator smell            | ✅ Clean       | no service stores the injector; `Trigger` _receives_ it per call by interface contract                              |
| Injector stored inside a service | ✅ Clean       | `Checkable`/`Trigger` hold only recorder, options, mutex                                                            |

---

## 4. Resilience & Self-Health — What Is Verified Superb

- **Per-check panic isolation** — a panicking check becomes that check's classified `Infrastructure` error; the batch survives (health/probe.go:44-53, 66-82).
- **Two-phase drain, lockstep with core** — `Mounted.Drain` → readiness 503 immediately (even from stale cache) while the refresh loop keeps running; `Mounted.Shutdown` stops it; idempotent; re-Start legal (mount.go:184-228). Wired via `DrainHooks`/`ShutdownHooks` so external readiness flips at the _start_ of core's drain window — the exact reason core gained `DrainHooks` (AGENTS).
- **Fail-closed readiness** — `Ready()` = not shutting down AND roll-up ≠ fail; `/health/ready` = drain probe AND `cfg.ReadyCheck` (service.go:309-316, config.go:108-114).
- **Trace capture can never stall health checks** — `SnapshotIfAsync`, cooldown guarded by a mutex held only across the non-blocking call (adapter.go:195-218); nil-safe pass-through when the recorder is disabled (adapter.go:180-182).
- **Compile-time contract pins both ways** — `health.HealthRecorder = (*Trigger)` and `do.HealthcheckerWithContext = (*Checkable)` (contract_test.go:19-22): upstream interface drift is a build error, not a runtime surprise.
- **Version currency where it counts** — samber/do v2.1.0 = latest (proxy-verified 2026-09-20: `v2.1.0` is newest tag).
- **Both suites green today**: `flightrecorderhealth` and `health` pass `go test ./... -race -count=1` (re-run 2026-09-20); frh documents ~4.7µs/batch no-capture hot path (benchmark_test.go).

---

## 5. Findings

### F1 — P0: The flagship self-diagnosis feature dies silently on the convenience path

A consumer doing the natural thing —
`appkithealth.NewProbe(checks, health.WithHealthRecorder(frhealth.NewTrigger(rec, …)))` —
gets **no error and no trace captures**: go-health nils the recorder (`cfg.recorder = nil`, accessors.go:61; present in every published version v0.1.3 → v0.3.0), documented only in prose. appkit's `NewProbe` forwards the option without warning (probe.go:33-34) and its doc never mentions `HealthRecorder`. Only AGENTS.md (not shipped to consumers) records the trap.
**Impact:** the "capture a trace when health fails" feature — the entire reason flightrecorderhealth exists — vanishes for the appkit-favored, injector-free composition, with green logs.
**Fix:** (a) upstream: `NewWithHealthCheck` should _reject_ `WithHealthRecorder` with a clear sentinel error; (b) appkit (shippable now): warn in `NewProbe` godoc + a defensive `Probe()`-level check or documented example showing the injector path for Trigger users; (c) add the combination to `health/example` so the working wiring is copy-pasteable.

### F2 — P0: The full self-health stack has zero CI proof

`integration/go.mod` pins core/errorpages/otel/realtime — **no health, no flightrecorderhealth** (verified 2026-09-20). The only cross-module health composition is the `health/example` service, verified by _manual_ live E2E (AGENTS), and frh's examples are injector-only (no appkit `Service`, no `Mounted`). The stack this review is about — injector + `Checkable` + `Trigger` + `health.New` + `Mounted` + dashboard + appkit drain lockstep — is nowhere assembled under test.
**Fix:** add `TestHealthStackThroughAppkitService` to `integration/` against published tags: readiness 503 in lockstep on drain, recorder row visible, trigger captures on a failing check.

### F3 — P1: Health-family version drift (three ways) + AGENTS staleness

`flightrecorderhealth` pins go-health **v0.1.3**; `health` pins **v0.2.0**; upstream latest is **v0.3.0** (proxy-verified; adds `aggregate/` + `federation/` packages). Consequence: frh's compile-time `HealthRecorder` pin verifies against a version consumers of the health module no longer resolve. AGENTS.md additionally documents the health module at "go-health v0.1.3, go-health-dashboard v0.8.1" while `health/go.mod` says **v0.2.0 / v0.9.0** (AGENTS line fixed in this pass; the _unreleased_ dep bump means published health v0.1.1 consumers still get v0.1.3/v0.8.1 — worth a release note).
**Fix:** align frh → go-health v0.2.0+ (re-run suite; contract assertion unchanged per source diff), evaluate v0.3.0 for the health module, and release the dep bumps.

### F4 — P1: Lifecycle dualism is unbridged

Injector-centric consumers must coordinate `injector.Shutdown()` _and_ appkit's hooks by hand. go-health v0.2.0 already ships the bridge pattern — `Probe.AsShutdowner()` → `do.ShutdownerWithError` (accessors.go:146-149) — but `Mounted` exposes no equivalent, and bare `AsShutdowner` would stop the probe _without_ stopping the dashboard pusher that `Mounted.Shutdown` owns (mount.go:223-225).
**Fix:** `func (m *Mounted) AsShutdowner() do.ShutdownerWithError` wrapping `Drain()` + `Shutdown()` semantics (one small method; overlaps the W4 "`do` bridge" battery — ship this slice early). Note: adding samber/do to the health module contradicts its core-free/injector-free stance — a tiny `interface{ Shutdown() error }`-shaped return or a documented adapter in _frh_ (which already imports do) keeps the stance intact.

### F5 — P1: do-v2's lazy-healthy trap is undocumented on our side

Verified in do v2.1.0 source: an **unbuilt lazy service reports healthy** — `serviceLazy.healthcheck` returns `nil` when `!s.built` (service_lazy.go:127-152). That is why `Register`'s eager invoke is load-bearing (adapter.go:284-285) — remove it and the recorder silently stops appearing in dashboards. Neither `doc.go` nor README warns consumers that _their own_ lazy services (db, cache) report green until first resolution inside a go-health injector.
**Fix:** one gotcha paragraph + keep the existing eager invoke (correct as written).

### F6 — P1 (hygiene): Root `go.mod` demands go 1.27.1 — repo inconsistency

Auto-commit `d5c6693` (2026-09-18) bumped the root `go.mod` directive 1.26.7 → 1.27.1. No dependency requires it (do v2.1.0 needs go1.18; go-health v0.2.0 needs go1.26; dashboard v0.9.0 needs 1.26.7; root requires were untouched in that commit); all ten satellite modules remain 1.26.7; AGENTS pins 1.26.7; `go.work` still says 1.26.7. Net effect: every gopls/golangci-lint diagnostic in the workspace fails with `module . listed in go.work file requires go >= 1.27.1`, root-module commands fail locally (`GOTOOLCHAIN=local`, nixpkgs at 1.26.7 — already on the watchlist), and CI's `go-version-file: go.mod` now provisions a _different toolchain for the root job_ than the ten satellite jobs.
**Fix (USER decision — deliberate bump or accident?):** revert root directive to 1.26.7 (restores local + LSP + toolchain uniformity), or commit the floor bump properly (go.work + AGENTS + nixpkgs reality check). Not silently "fixed" here because reverting a directive I didn't author could fight an intentional floor raise.

### F7 — P3 polish: `Register` ergonomics

`do.ProvideNamedValue(injector, name, checkable)` is the canonical eager registration and would delete the closure + the ignored error (`_, _ =` adapter.go:285). Also undocumented: `do.ProvideNamed` panics on duplicate names, so calling `Register` twice with the same name panics. Behavior-preserving; the module's Register tests cover it.

### F8 — P3: `samber-do-auditlog` hooks unwired

Registration/invocation/health/shutdown observability hooks exist in the family but no go-appkit example wires them. Low value until a consumer asks for DI-level audit trails.

---

## 6. Rubric Scoring (1-5, evidence-cited)

| Dimension            | Score    | Evidence                                                                                                                                                                         |
| -------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Coupling             | 5        | interfaces at every seam (`HealthRecorder`, `CheckFunc`, duck-typed `PatchLike` in realtime); DI via config+hooks; zero shared mutable state across modules                      |
| Cohesion             | 5        | each module passes the one-paragraph test; health=health surface, frh=bridge; no utils grab-bags                                                                                 |
| Modularity           | 5        | 11 independently versioned modules; `integration` pins published tags; `GOWORK=off` hermetic builds prove boundary integrity                                                     |
| Composability        | 4        | excellent config-level composition (`DrainHooks`/`ShutdownHooks`/`ReadyCheck`/options); −1 for the `RegisterHealth=&false` + dashboard route-conflict traps and the F1 cliff     |
| Scalability          | 4        | horizontal by design (drain 503 → LB re-routes; documented SSE shutdown ordering); process-global singletons (recorder, otel Setup) are honest, documented limits                |
| Service orientation  | 4        | bounded contexts, independent lifecycles, clean contracts, clear data ownership; −1 for consumer-held lifecycle coordination (F4)                                                |
| Dependency direction | 5        | satellites are core-free (health, realtime, otel, frh, flightrecorder); core depends only on httputil/charmbracelet/go-error-family; errors flow one way through go-error-family |
| **Average**          | **4.57** | **Excellent — maintain and document**                                                                                                                                            |

---

## 7. Action Roadmap (impact × effort, ordered)

1. **F1 appkit-side now:** `NewProbe` godoc warning + injector-path Trigger example (hours, removes the silent cliff for consumers without waiting on upstream).
2. **F1 upstream ask (draft, USER-gated filing):** go-health `NewWithHealthCheck` should reject `WithHealthRecorder` with a sentinel error.
3. **F2:** integration test `TestHealthStackThroughAppkitService` against published tags (needs health/frh releases first → depends on 5).
4. **F6 decision** (USER): revert root go directive vs deliberate floor bump — either way, make `go.work`, AGENTS, and CI agree in one commit.
5. **F3:** frh → go-health v0.2.0; health module dep-bump release (v0.1.1 → v0.1.2) so published consumers stop resolving v0.1.3/v0.8.1.
6. **F5:** lazy-healthy gotcha in frh doc.go/README (minutes).
7. **F4:** `Mounted` shutdown adapter (or frh-side adapter to preserve health's injector-free stance).
8. **F7/F8:** polish + auditlog example (demand-gated).

Items 1-6 harvested into `TODO_LIST.md` this pass.

---

## 8. Verification Log (2026-09-20)

- `cd flightrecorderhealth && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1` → **ok**
- `cd health && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1` → **ok**
- `go list -m -versions` (proxy): samber/do/v2 latest = v2.1.0 (pinned ✓); go-health latest = v0.3.0 (health pins v0.2.0, frh pins v0.1.3)
- Source-verified: do-v2 `service_lazy.go:127-152` (lazy-healthy), go-health `accessors.go:56-66` (recorder drop, v0.1.3/v0.2.0/v0.3.0 identical), `probe.go:310,357-367` (recorder path), `accessors.go:117-149` (Probe-as-Healthchecker/Shutdowner), httputil v1.2.0 `health.go:88-94` (GET-qualified default routes), do-v2 `scope.go:294-345` (fan-out + `HealthCheckGlobalTimeout`)
- `git show d5c6693` — root go.mod 1.26.7→1.27.1, no root dep change; satellites untouched

## 9. Sources

- `flightrecorderhealth/adapter.go`, `doc.go`, `contract_test.go`, `example_test.go`; `health/probe.go`, `mount.go`, `doc.go`, `example/main.go`; `health.go`, `config.go:108-114`, `service.go:309-316` (this repo)
- `go-health@v0.1.3/v0.2.0/v0.3.0` module sources; `samber/do/v2@v2.1.0` module sources; `httputil@v1.2.0/health.go`
- AGENTS.md (Testing, Gotchas, Health/frh sections); TODO_LIST.md (P2/P3 health entries); `.github/workflows/ci.yml:62` (`go-version-file`)

---

## ADDENDUM — Scorecard re-run after the 2026-09-20 execution train (plan T27)

Train shipped: health v0.1.2 + flightrecorderhealth v0.1.3 (F3), F1 cliff
godoc + output-pinned injector-path example + integration E2E (F1/F2), root
go.mod revert + CI directive-parity guard (F6), Mounted→do.Shutdowner
adapter in health/doadapter (F4), lazy-healthy gotcha docs (F5),
do.ProvideNamedValue Register with panic contract (F7), C1 Mounted.Start
rollback fix, upstream drafts ready but unfiled (gated), auditlog chaining
documented but dependency demand-gated (F8).

| Dimension            | Was      | Now      | Delta evidence                                                                                                                                        |
| -------------------- | -------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Coupling             | 5        | 5        | unchanged                                                                                                                                             |
| Cohesion             | 5        | 5        | unchanged                                                                                                                                             |
| Modularity           | 5        | 5        | strengthened: integration now pins + proves the health family (T10/T11)                                                                               |
| Composability        | 4        | 5        | the F1 cliff is warned + example-pinned + E2E-proven; route-conflict traps documented; per the rubric's own convention (documented limits earn marks) |
| Scalability          | 4        | 4        | unchanged (process-global singletons remain honest documented limits)                                                                                 |
| Service orientation  | 4        | 5        | F4 closed: health/doadapter gives injector-owned shutdown without breaking health's injector-free API                                                 |
| Dependency direction | 5        | 5        | strengthened: new family edge avoided (doadapter as subpackage, not frh→health); directive parity CI-guarded                                          |
| **Average**          | **4.57** | **4.86** | (5+5+5+5+4+5+5)/7                                                                                                                                     |

**Adoption score: 82 → 93/100.** Remaining deductions, all explicitly gated
rather than forgotten: upstream go-health behaviors are unfixed (recorder
sentinel + Evaluate cache publication + restart rearm — drafts in
`doc/feedback/outgoing/2026-09-20_*`, filing USER-gated), and auditlog
chaining ships as a documented pattern instead of a wired dependency
(demand gate).
