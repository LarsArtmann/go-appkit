# TODO List — go-appkit

> Short-term, actionable, bounded work. Open items only.
> Completed work lives in [CHANGELOG.md](CHANGELOG.md). Long-term vision lives in `docs/planning/`.

**Updated:** 2026-09-04 (execution wave) | **Modules:** 10 (core, cqrs, realtime, otel, flightrecorder, flightrecorderhealth, health, errorpages, docs, integration — integration is the unreleased E2E test module) | **Release state:** core **v0.3.0**, cqrs **v0.3.0**, realtime **v0.1.0**, flightrecorder **v0.1.0**, flightrecorderhealth **v0.1.0** cut and **PUSHED 2026-08-30** (tags on origin; fresh-consumer proxy-verified 2026-09-04); **wave 2 cut and pushed 2026-09-04: core v0.4.0 (lifecycle hooks), cqrs v0.4.0 (BREAKING FlightRecorder migration), otel v0.1.0, health v0.1.0, flightrecorderhealth v0.1.1 — all fresh-consumer proxy-tested**; errorpages v0.1.0 and docs v0.2.0 current (their since-tag deltas are test-only / config-only — deliberately not re-tagged); **`encoding/json/v2` is default-on in Go 1.26.7 — GOEXPERIMENT prefixes are only needed on older 1.26.x toolchains**

## Status Legend

- [ ] **OPEN** — actionable, not yet started.
- [~] **PARTIALLY DONE** — started but incomplete.

> No `[x]` items here. When a task finishes, it moves to the module CHANGELOG and is removed from this list.

---

## P1 — High impact (release follow-through)

- [~] **pkg.go.dev verification for wave-1 + wave-2 tags — BLOCKED on the licensing decision (below).** Proxy smoke DONE 2026-09-04 for both waves (clean /tmp module → `go get` → blank import → `go build`): core v0.3.0/v0.4.0, cqrs v0.3.0/v0.4.0, realtime v0.1.0, flightrecorder v0.1.0, flightrecorderhealth v0.1.0/v0.1.1, otel v0.1.0, health v0.1.0. All wave-2 tags carry the module-root LICENSE files (mechanical fix landed 2026-09-04). Remaining: (a) pkg.go.dev crawl/re-render check per module page, (b) the licensing decision itself — until the root LICENSE is a classifiable text, core renders `License: UNKNOWN` with godoc hidden. See P2 licensing decision.

> Closed 2026-09-04 (execution wave): OTEL commit (`aaa2427`), `otel/v0.1.0` + `health/v0.1.0` + `flightrecorderhealth/v0.1.1` tags (replaces dropped, hermetically verified, fresh-consumer proxy-tested), core v0.4.0 with DrainHooks, mechanical API-break check (go-doc snapshot diff, proven on core v0.3.0→v0.4.0), README GOEXPERIMENT section, DrainDelay sweep (satellites already clean), cqrs-lint 4.8.1 installed, go-structure-linter zeroed (CLI-flag excludes), otel middleware benchmark, core v1.0.0 exit criteria draft (`docs/planning/core-v1-exit-criteria.md`), design-decisions.md 404 (link inlined), cqrs trigger/precedence tests + godoc examples.

---

## P2 — Medium impact (quality & docs)

- [ ] **Decide the license posture (USER GATE) — pkg.go.dev is hiding the project.** The root LICENSE is an unclassifiable PROPRIETARY text: pkg.go.dev marks the core module `License: UNKNOWN` and hides ALL godoc, and submodule pages 404 outright without a module-root LICENSE (mechanical fix landed 2026-09-04: LICENSE copied into all 7 module roots, mirroring the cqrs-htmx family pattern). Decisions needed: (1) keep proprietary (docs stay hidden — real adoption cost for a framework courting consumers like cqrs-htmx) or adopt a standard license (cqrs-htmx itself is MIT); (2) ship the LICENSE files in the next tag wave, then re-verify pkg.go.dev renders every module with visible godoc.
- [ ] **Decide the logging posture (comparison finding 7) — data now exists.** Per-request benchmark (2026-09-04, `logging_bench_test.go`, output to io.Discard): bare ~17.2µs; suppressed (WARN) ~18.0µs (+0.8µs); emitting (INFO) ~47.3µs (+30µs, +174%, 162 allocs) — the charmbracelet formatting of the emitted line IS the 2.8x delta vs the cqrs-htmx spike. Options: default WARN in production docs, sampling, or consumer-provided logger (already possible). Decide + implement + benchstat before/after. Source: `docs/review/2026-08-16_setup-vs-go-appkit-comparison.md` finding 7.
- [ ] **README: document `GOEXPERIMENT=jsonv2` for building from source.** AGENTS.md carries per-module reasons; README (user-facing) does not mention it. Source: session-5 #19.
- [ ] **Upstream: cqrs-lite otel `Provider.Shutdown` must ForceFlush before Shutdown.** Same batch-queue race fixed locally in go-appkit/otel — spans ended moments before Shutdown sit in the async queue and are silently dropped. Repro probe exists from the 2026-08-18 OTEL session; check for an existing upstream fix first, then file. Issue vs PR is status report §g Q3 (verify-before-filing applies).
- [ ] **Sweep satellite module tests for `DrainDelay: 0` misuse.** Zero applies the 5s default (hidden test tax); core's conversion to `NoDrainDelay` dropped wall time ~30s→6s. realtime/errorpages/docs/flightrecorder likely carry the same pattern.
- [ ] **Add a mechanical API-break check to the release process** (goapidiff or a `go doc` snapshot diff at tag time) — v0.2.0 was reconstructed after the fact; never again. Source: session-5 #7.
- [ ] **Bump Go toolchain past 1.26.7 when nixpkgs carries it** (GO-2026-6090 crypto/tls post-handshake flood, GO-2026-5972 encoding/asn1 recursion were the 1.26.5-era findings; all modules already sit at go 1.26.7 so the superseded 1.26.6 item is closed — this tracks the NEXT bump; builds run GOTOOLCHAIN=local so the nixpkgs lock gates this). Source: govulncheck via BuildFlow 2026-08-16; same item tracked in cqrs-htmx TODO_LIST P2.

---

## P3 — Technical debt & tooling

- [ ] **BuildFlow dprint step fails on CHANGELOG-only commits** (exit 14 "no files found to format"): `dprint.json` excludes `**/CHANGELOG.md`, so a commit staging only CHANGELOGs (+ non-dprint files) trips the formatter's empty-set error. Root fix is upstream in buildflow (skip when the staged set ∩ dprint's non-excluded set is empty — see AGENTS.md Deferred Register); escape hatch until then: `--no-verify` + justification. Evidence: the v0.3.0-wave commit needed `--no-verify` on 2026-08-16 (deterministic, retried once).
- [ ] **otel: middleware benchmark (no-op vs configured) + benchstat** — DONE 2026-09-04: `otel/benchmark_test.go` (no-op ~21µs / traced ~26µs / traced+metered ~27µs per request; export I/O excluded), numbers recorded in otel README. Remaining: benchstat-formatted before/after once an optimization candidate exists.
- [ ] **Propose httputil `Logging` emit with request context** so completion lines correlate with spans via `TraceHandler` (today only handler-level logs correlate). Status report §f P2-12.
- [~] **Define v1.0.0 exit criteria for core** — DRAFT WRITTEN 2026-09-04: `docs/planning/core-v1-exit-criteria.md` (7 hard criteria incl. the mechanical API-break check, 3 soft signals, explicit non-goals). Stays draft until the consumer count grows; fold in deeper OTEL backlog from `docs/status/2026-08-18_12-45_otel-module-and-telemetry-hooks.md` §f when it graduates.
- [ ] **Revisit cordis bridge module when triggers fire** (researched 2026-09-04, verdict: NOT NOW). `docs/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md` — all three required: (1) cordis tags `go/v0.1.0` (today proxy-pinned to a pseudo-version), (2) a real consumer states the reactive-composition requirement, (3) core v1.0.0 exit criteria shipped. Bridge sketch (~300-500 LOC, zero third-party deps) is in the doc §4 Option B. go-plugin-mvp marketplace integration: rejected as appkit dependency (wrong layer, pre-1.0 churn); recommended reverse adoption (their repo embeds appkit, cqrs-htmx ADR-001 pattern) — proposal to their TODO_LIST is user-gated.

- [ ] **PapDashboard integration: hold the reverse-adoption door open** (researched 2026-09-04, verdict: NO go-appkit code). `docs/planning/2026-09-04_papdashboard-integration.md` — PapDashboard (event-sourced notification hub, published as `github.com/larsartmann/papdashboard` v0.2.0) rejected as appkit module/dependency: it is an application, not a framework primitive, and only `pkg/enum` + `sdk/` are importable (`internal/` encapsulation). Recommended: PapDashboard hosts on appkit in THEIR repo — every load-bearing version already matches (go-error-family v0.10.0, samber/do v2.1.0, Go 1.26.7, jsonv2 already mandatory; their `cqrshtmx.Chain` is appkit-free even at cqrs-htmx v4.9.0) and appkit's drain phase fixes a real gap (their SSE stream + outbound webhooks are hard-cut on shutdown today). Proposal to their TODO_LIST is user-gated behind their open E14/M1 queue. appkit-side signal worth tracking: core has NO TLS support; their `PAP_TLS_CERT/KEY` is the first concrete consumer demand for a core TLS option if they want to keep app-level TLS. `integration/` sdk-driven E2E test becomes actionable only after they ship an appkit-hosted release.

- [ ] **Upgrade installed cqrs-lint 4.6.0 → 4.8.1** — DONE 2026-09-04: built 4.8.1 from the local cqrs-lite repo into `~/go/bin`; clean on `cqrs/` (2 suppressions still fire). Arg-form note (`.` not `./...`) documented in cqrs README.
- [ ] **cqrs: EventConfig opt-ins for event encryption/signing when a consumer demands it** (encryption/v4, signing/v4; candidates: idempotency/sqlstore, scheduling). Deliberately NOT added now — wrapper surface stays demand-driven. Routed from `docs/research/2026-09-04_go-cqrs-lite-deep-dive.html` finding 5.

---

_For completed work, see each module's `CHANGELOG.md` and `git log`._
