# TODO List — go-appkit

> Short-term, actionable, bounded work. Open items only.
> Completed work lives in [CHANGELOG.md](CHANGELOG.md) and each module's CHANGELOG. Long-term vision lives in [ROADMAP.md](ROADMAP.md).

**Updated:** 2026-09-17 (docs-health pass: pkg.go.dev VERIFIED — every module page renders; TODO_LIST rebuilt to open-items-only per its own legend; latest session evidence: `doc/status/2026-09-17_14-22_setup-usage-verification-and-agentsmd-drift-fix.md`) | **Modules:** 11 (core, cqrs, realtime, otel, flightrecorder, flightrecorderhealth, health, errorpages, docs, security, integration — integration is the unreleased E2E test module) | **Release state:** every module tag is ON ORIGIN through **core v0.5.1** (2026-09-17, doc-only); prior waves: v0.5.0 + security v0.1.0 / realtime v0.1.1 / health v0.1.1 / frh v0.1.2 / docs v0.3.0 / otel v0.1.1 + httputil v1.2.0 (2026-09-16), cqrs v0.5.0 (2026-09-07); **`encoding/json/v2` is default-on in Go 1.26.7 — GOEXPERIMENT prefixes are only needed on older 1.26.x toolchains**

> **Release-state ownership (decided 2026-09-16):** `AGENTS.md` → *Release State* is the SINGLE OWNER of release facts. This header links history only; status reports are point-in-time snapshots with as-of dates and never get updated afterward. Index: `doc/status/README.md`.

## Status Legend

- [ ] **OPEN** — actionable, not yet started.
- [~] **PARTIALLY DONE** — started but incomplete.

> No `[x]` items here. When a task finishes, it moves to the module CHANGELOG and is removed from this list.

---

## P1 — High impact (release correctness)

_No open items. Release state is green: every module tag on origin, fresh-consumer proxy-proven, and pkg.go.dev verified 2026-09-17 (all module pages render — docs v0.3.0, health v0.1.1, security v0.1.0, core v0.5.1; godoc hidden by the proprietary-license choice). New release work starts from the Release Ritual (AGENTS.md)._

## P2 — Medium impact (quality & docs)

- [ ] **Pin-drift guard:** a test or CI step asserting the `integration/go.mod` pins — and the AGENTS Release State "ON ORIGIN through" version against `git tag -l` — match the documented pins. The v0.5.0-vs-v0.5.1 AGENTS divergence (`doc/status/2026-09-17_14-22_setup-usage-verification-and-agentsmd-drift-fix.md` §d-1) was findable by one grep and nothing ran it. USER GATE first: pin philosophy (§g-1 — LATEST-only vs mirroring setup's resolution).
- [ ] **Release Ritual additions:** (1) an explicit "update AGENTS.md release-state + module lines" step — v0.5.1 shipped while AGENTS still said v0.5.0 (14-22 §e-2); (2) codify the `docs:` tag-message convention for doc-only releases (v0.5.1 did this well).
- [ ] **Move the integration pin table out of AGENTS.md into `integration/doc.go`** (or the module README) — single source of truth next to go.mod; relieves the AGENTS 376/377 line cap (14-22 §f-6/§e-3).
- [ ] **Depguard deny rule in the root `.golangci.yml`:** forbid `cqrs-htmx/setup` imports repo-wide — the "setup is NEVER a dependency of appkit" invariant is currently enforced only by prose (14-22 §f-7/§e-5).
- [ ] **security module trio:** (1) threat-model page (per-battery threat → test mapping table); (2) example service demonstrating the full hardened chain (like errorpages/example); (3) integration test for security + realtime composition (rate-limit in front of SSE). Sources: 05-52 §f-39/§f-42/§f-43.
- [ ] **Browser CSP pass over the health dashboard under a strict-CSP + nonce profile** (chromedp or manual) — the default-stack verdict is server-side proven; the hardened configuration has no browser-side proof (08-08 §f-15; profile: health example + `DashboardHardenedPreset`).
- [ ] **CI dependabot-parity assert:** workflow step that fails when a module dir lacks a dependabot entry or CI matrix slot (05-52 §f-48; `/security` was verified byte-identical by hand 2026-09-17 — the assert makes it permanent).
- [ ] **Fresh-consumer proxy smoke in CI** as a manual-dispatch workflow — verify runner network access first (08-08 §f-34); the recipe exists (`doc/recipes/fresh-consumer-proxy-check.md`).
- [ ] **testkit explicit-shutdown helper:** calling `svc.Shutdown` in the test body AND letting testkit cleanup shut down again works (idempotent) but pays the cleanup `errCh` wait in every such test; a `TestServer.Shutdown` helper that marks cleanup no-op removes the latency (08-08 §f-4; low priority).
- [ ] **errorpages: replace the hand-rolled `statusRecorder` with `httputil.ResponseRecorder`** (USER GATE, open since 2026-09-16): ~10-line change; two ResponseWriter wrappers, one job, one repo (source: 08-53 ghost-systems + §g-1).
- [~] **Compose `Service` on `httputil.Server` — BLOCKED on upstream API (do not refactor today).** Spike verdict (`doc/planning/2026-09-16_composition-spike-verdict.md`): config mapping clean (6/7 identity + the TLS seam) but httputil's `Start()` binds internally with ASYNC bind errors vs appkit's synchronous classified `listen_failed`; `Addr()`/`Running()` semantics and the pinned shutdown phase-log contract would split-brain. Unblock: the httputil listener-injection API (`NewServerListener(ln, cfg, handler)`) — USER-GATED upstream ask; Core TLS (G1) is gated on the same seam.
- [~] **Upstream asks — DRAFTED, filing USER-GATED:** (1) go-sse dedup-aware `ReplayFiltered`; (2) httputil Logging request-context emit; (3) F2 timing battery sketch (sequenced after 2 — shared duration source); (4) the `NewServerListener` implementation go/no-go. Drafts: `doc/feedback/outgoing/2026-09-16_upstream-asks-gosse-httputil.md`.
- [~] **Health-module quality parity — only govulncheck remains (env-blocked).** Everything else landed 2026-09-16/17 (contract assertion, conflict semantics, BasePath uniform routing, batch benchmark, fuzz, godoc examples F107+F113). govulncheck needs a networked machine (`go install` blocked here); run on health + security.
- [~] **otel: benchstat re-baseline — fresh n=10 mean±sd baseline recorded 2026-09-16** (no-op 20.0µs ± 0.4 / traced 21.8µs ± 0.9 / traced+metered 23.3µs ± 2.0; run-to-run drift on this box is ±25%). Stays open only as a candidate: re-measure with real benchstat (`nix run nixpkgs#benchstat` was never probed) when an optimization candidate exists.
- [~] **Define v1.0.0 exit criteria for core** — draft: `doc/planning/core-v1-exit-criteria.md` (7 hard criteria incl. the mechanical API-break check, 3 soft signals, explicit non-goals). Stays draft until the consumer count grows; fold in the documented-wiring-test lesson + the deeper OTEL backlog from `doc/status/archived/2026-08-18_12-45_otel-module-and-telemetry-hooks.md` §f when it graduates.

## P3 — Demand-gated / watchlist

- [ ] **Battery program waves W3-W5** (triaged 2026-09-04; no consumer demand signal since — long-term). W3 handler DX: new `httpx` module (B1 ResultHandler family — error mapping MUST share errorpages' taxonomy, pinned by a classification-parity test; B2 bind+validation; B9 no-leak error responses; then B3 ResponseWriter contract, B4 conditional GET — promotes the unused go-etag indirect dep, B5 content negotiation, B7 per-route write deadlines; then B6 route introspection + golden test, B8 route metadata → docs feed — both need a core `svc.Routes()` seam core has not agreed to). W4 ops & data: `worker` supervisor+pool, `sqlite` ops kit (lease/backup/ledger), `polite` outbound client, `do` bridge, D7 idempotency store, D4 atomic file write (floor: go-atomic-write ≥ v0.5.1, Windows syscall trap), D6 webhook, F1/F4 config modules. W5 realtime completions: C1 drop/backpressure counters (covers the 08-53 `WithOnDrop` audit item — dedupe, do not double-track), C3 per-subscriber auth+filter, C2 projection→broadcast folded contract (the must-have — without it every cqrs+realtime consumer re-runs the pre-fold-state race). Full API sketches, effort sizing, and acceptance tests: `doc/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md` — the canonical spec; do not duplicate its content here.
- [ ] **cqrs: EventConfig opt-ins for event encryption/signing when a consumer demands it** (encryption/v4, signing/v4; candidates: idempotency/sqlstore, scheduling). Deliberately NOT added now — wrapper surface stays demand-driven. Routed from `doc/research/2026-09-04_go-cqrs-lite-deep-dive.html` finding 5.
- [ ] **Full cqrs-htmx setup suite hermetic run** (bounded, container-aware) to extend the verified 3/3 appkit-composition green to suite-green (14-22 §b-1).
- [ ] **AGENTS deep slim-down decision:** what graduates to module READMEs (Release Ritual? per-module Gotchas?) — line-level edits sit at the 376/377 cap (binary counts +1); the deep cut is a structure decision (09-38 §b-2, 14-41 §b-4).
- [ ] **cqrs README cookbook re-verification** against scenario/v4 v4.2.0 after each go-cqrs-lite release (standing ritual; the next release triggers it).
- [ ] **BuildFlow dprint step fails on CHANGELOG-only commits** (exit 14 "no files found to format"): root fix upstream in buildflow (skip when the staged set ∩ dprint's non-excluded set is empty — AGENTS Deferred Register); escape hatch until then: `--no-verify` + justification.
- [~] **cordis bridge — WATCHLIST (trigger 1 of 3 MET: `go/v0.1.0` tagged, verified 2026-09-16).** Remaining triggers: a real consumer states the reactive-composition requirement AND core v1.0.0 exit criteria shipped. Bridge sketch (~300-500 LOC, zero third-party deps): `doc/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md` §4 Option B. go-plugin-mvp: rejected as appkit dependency; reverse adoption remains the recommendation (user-gated proposal to their TODO_LIST).
- [~] **PapDashboard — WATCHLIST (they shipped v0.3.0).** Reverse-adoption door stays open; the standing checks apply to v0.3.0: family-dep parity, whether they added TLS (their `PAP_TLS_CERT/KEY` was the first concrete core-TLS demand), whether an appkit-hosted release is on their roadmap. Next look belongs in THEIR repo. `doc/planning/2026-09-04_papdashboard-integration.md` remains the verdict doc (NO go-appkit code).
- [~] **Go toolchain watch — nixpkgs still carries go 1.26.7 (checked 2026-09-16).** The GO-2026-6090 fix lands with the next nixpkgs bump; no action available in this environment; re-check on the next nixpkgs pull.
- [ ] **Watchlist refresh (standing):** cordis consumers count; PapDashboard v0.3.1+; nixpkgs toolchain > 1.26.7; dprint exit-14 upstream; cqrs-htmx — their M3 appkit bump to v0.5.0 LANDED 2026-09-17 (setup/go.mod:95); when they land M4 (`Config.Metrics`/`Config.Version` threading), re-verify integration's "deliberately NOT setup's pin" note, refresh the AGENTS reference-consumer line, and smoke-test the metrics surface through setup once; when their default-flip (b)-(f) lands, update the reference-consumer line again.

---

**Open questions awaiting USER decisions** (they gate P2 items above): (1) pin philosophy for `integration/` — LATEST-only vs also mirroring setup's resolution; (2) cross-repo tracking — actively nudge cqrs-htmx items or stay report-only; (3) AGENTS as release-state source (with the CI guard from the pin-drift item) vs moving pins into code-adjacent files entirely. Full context: `doc/status/2026-09-17_14-22_setup-usage-verification-and-agentsmd-drift-fix.md` §g.

_For completed work, see each module's `CHANGELOG.md` and `git log`._
