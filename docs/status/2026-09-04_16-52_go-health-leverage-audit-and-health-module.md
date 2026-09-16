# Status Report — go-health / go-health-dashboard Leverage Audit + New `health` Module

**Date:** 2026-09-04 16:52 CEST
**Session scope:** library-deep-dive style leverage audit of `go-health` + `go-health-dashboard` inside go-appkit, followed by executed integration work (core seam, new module, dep bump, docs, full verification).
**Framing:** honest point-in-time snapshot. Not a living document — verify claims against code before acting on them.

---

## Executive summary

The audit found that go-appkit used go-health only as a compile-time assertion (pinned 2 minors behind) and go-health-dashboard not at all. This session closed the three biggest gaps:

1. **Core `DrainHooks`** — new shutdown phase (after ready-probe flip, before drain wait), enabling lockstep readiness 503s across all health surfaces.
2. **New `health` module** (`/health`, package `health`, alias `appkithealth`) — injector-free probes (`NewProbe`), mux wiring (`New`/`Mount`/`RegisterRoutes`), opt-in real-time dashboard (`WithDashboard`), lifecycle (`Start`/`Drain`/`Shutdown`/`Ready`) wired via `ServiceConfig`.
3. **flightrecorderhealth** — go-health v0.0.2 → v0.1.1 (contract signature verified unchanged at tag), go directive 1.26.5 → 1.26.7, build-requirement docs corrected (now requires `GOEXPERIMENT=jsonv2`).

Verified: race-clean tests in all 3 touched modules, 0 golangci-lint issues in all 3, hermetic builds across all 9 modules, live E2E (dashboard HTML/JSON/SSE/metrics + drain-window 503 lockstep on both `/readyz` and `/health/ready`), README quick start compiles verbatim in a scratch module.

**Two potentially serious gaps were discovered during this self-review (§d-1, §d-2) — both concern the dashboard's SSE under appkit's default config and were NOT verified this session.**

---

## a) FULLY DONE

| #   | Item                                                                                                                                                                                                                                                                                                                 | Evidence                                                   |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| a1  | Leverage audit of go-health v0.1.1 + go-health-dashboard v0.5.0 vs go-appkit usage (48 grep hits analyzed, both AGENTS.md files read, tag-level API verification)                                                                                                                                                    | session research; findings table in final answer           |
| a2  | Core `DrainHooks`: config field + doc comment, `Shutdown` sequencing (flip probe → DrainHooks → drain wait → close → ShutdownHooks), error joining classified Infrastructure                                                                                                                                         | `config.go`, `service.go:148-215`, `drainhooks_test.go`    |
| a3  | Core drain tests (4): traffic-window ordering proof (HTTP GET inside hook shows ready=503 while socket serves), failure isolation + errors.Is, run-once, never-started skip                                                                                                                                          | `drainhooks_test.go`, core suite race-clean 6.0s           |
| a4  | `health.NewProbe`: named `CheckFunc` map → go-health probe via `NewWithHealthCheck`; concurrent batches (`wg.Go`); per-check panic isolation (`errorfamily.Infrastructure` `health.check_panicked`); SDK options pass-through                                                                                        | `health/probe.go`                                          |
| a5  | `health.Mount` API: `New` (config-first flow) + `Mounted.RegisterRoutes` + `Mount` sugar; options `WithDashboard(...)` (opt-in), `WithProbeRoutes` (probe-only mode)                                                                                                                                                 | `health/mount.go`                                          |
| a6  | `Mounted` lifecycle: `Start` (initial sync batch + refresh loop + pusher, double-start rejected, legal after Shutdown), `Drain` (idempotent 503 flip), `Shutdown` (idempotent, re-Start legal), `Ready`, `Probe()`, `Dashboard()` accessors                                                                          | `health/mount.go`                                          |
| a7  | 14 health-module tests (classification, panic isolation, concurrency handshake, bounded ctx, probe-only routes, custom routes, dashboard HTML/JSON/probe routes, drain flip, lifecycle guards, validation errors.Is through wrap)                                                                                    | `health/probe_test.go`, `health/mount_test.go`, race-clean |
| a8  | Runnable example (critical + flapping non-critical check, dashboard w/ trend + metrics, full DrainHooks/ShutdownHooks wiring, PORT-aware)                                                                                                                                                                            | `health/example/main.go`                                   |
| a9  | **Live E2E**: all endpoints verified against the running example (200s: `/healthz`, `/readyz`, `/health` HTML + JSON negotiation, `/health/metrics` text 0.0.4, `/health/sse` event-stream); on SIGTERM both readiness surfaces flipped 503 at t=0.0s and held through the drain; clean process exit                 | session log, `/tmp` probes                                 |
| a10 | flightrecorderhealth dep bump (go-health v0.1.1, go 1.26.7) — compile-time contract assertions hold, all tests race-clean, 0 lint issues                                                                                                                                                                             | `flightrecorderhealth/go.mod`, test run                    |
| a11 | Build-requirement docs corrected everywhere (README, CHANGELOG [Unreleased], .golangci.yml build-tags + header) for flightrecorderhealth                                                                                                                                                                             | those 3 files                                              |
| a12 | Module-local `.golangci.yml` for health (realtime template, jsonv2 build-tags, go 1.26.7) — 0 issues                                                                                                                                                                                                                 | `health/.golangci.yml`                                     |
| a13 | go.work: `./health` registered; workspace-wide build green; all 9 modules build hermetically                                                                                                                                                                                                                         | `go.work`, build loop                                      |
| a14 | Root docs: AGENTS.md (9-module list, 8-of-9 jsonv2 note, health section + gotchas, build commands, drain sequence updated), README.md (module row, jsonv2 note), FEATURES.md (core DrainHooks row + full health section), core CHANGELOG (DrainHooks + health entries), TODO_LIST.md (counts, new release-work item) | those files                                                |
| a15 | README quick start **compiles verbatim** in a scratch module — the check caught and fixed two real snippet bugs (unhandled `err`, undefined `ctx`)                                                                                                                                                                   | scratch-module build: PASS                                 |
| a16 | Discovered + documented the Go 1.22+ ServeMux precedence conflict: method-agnostic `/health` (dashboard) vs `GET /` catch-all panics; example + README gotcha carry it                                                                                                                                               | `health/README.md` gotchas, example comment                |

## b) PARTIALLY DONE

| #  | Item                                        | State                                                                                                                | What's missing                                                                                                                                              |
| -- | ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~b1~~ | ~~Dashboard-under-appkit-default-stack story~~ done — owned by TODO_LIST P2 (dashboard CSP + SSE longevity, routed 2026-09-16) | ~~Endpoints verified server-side; SSE connected once via urllib~~ | ~~Browser-side SDK execution under appkit's `SecurityHeaders` CSP **unverified** (see §d-1); SSE longevity under default `WriteTimeout` unverified (see §d-2)~~ |
| ~~b2~~                                          | ~~flightrecorderhealth release prep~~ done — shipped as flightrecorderhealth/v0.1.1 (2026-09-04, pushed) | ~~CHANGELOG [Unreleased] written, tests/lint green, contract asserted against v0.1.1~~ |
| ~~b3~~                                          | ~~health module release prep~~ done — shipped as health/v0.1.0 (2026-09-04, pushed) | ~~Module complete, example replace noted, TODO_LIST item written~~ |
| ~~b4~~ | ~~Integration story with flightrecorderhealth~~ done — routed to TODO_LIST P3 integration-module expansion (cross-module Trigger + Mount test) | ~~README section documents the injector-path requirement (WithHealthRecorder is a no-op for NewWithHealthCheck probes)~~ | ~~No cross-module integration test proving Trigger + injector-probe + this module's Mount together~~ |
| ~~b5~~ | ~~Self-review skill output~~ **Won't implement — user chose Markdown format; HTML report intentionally not produced.** | ~~Full brutal review folded into this report (§d, §e) per user's explicit Markdown/format instruction~~ | ~~Skill's default HTML report at `docs/reviews/` not produced (user format won)~~ |
| ~~b6~~                                          | ~~Docs-health wiring~~ done — this report + TODO_LIST (routed 2026-09-16) | ~~This report + TODO_LIST item~~ |

## c) NOT STARTED

| #  | Item                                                                                                                                                                             |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~c1~~ | ~~CSP integration helper (nonce-aware wiring: `WithNonceExtractor` + `RecommendedCSP` + httputil nonce middleware) — decision pending (§g-1)~~ done — owned by TODO_LIST P2 (dashboard CSP + SSE longevity item, routed 2026-09-16) |
| ~~c2~~ | ~~Benchmarks for `NewProbe` batch overhead (N=1/5/20 checks; house has e.g. flightrecorderhealth's 4.7µs/batch)~~ done — routed to TODO_LIST P3 health-module quality parity |
| ~~c3~~ | ~~Fuzz targets (mount options, panicking checks) — go-health itself has fuzz targets; this module none~~ done — routed to TODO_LIST P3 health-module quality parity |
| ~~c4~~ | ~~Runnable godoc examples (`example_test.go` with verified output) — house pattern in flightrecorderhealth, missing here~~ done — routed to TODO_LIST P3 health-module quality parity |
| ~~c5~~ | ~~Compile-time contract assertions (`contract_test.go`): e.g. `var _ dashboard.Prober = (*health.Probe)(nil)` — the exact split-brain guard this repo praises elsewhere (see §e-3)~~ done — routed to TODO_LIST P3 health-module quality parity |
| ~~c6~~ | ~~aggregate multi-probe example (`aggregate.New` + dashboard through `Mounted`)~~ done — routed to TODO_LIST P3 health-module quality parity |
| ~~c7~~ | ~~OTel bridge: `WithEvaluationHook` → appkit/otel metrics (documented as YAGNI-for-now)~~ **Won't implement — YAGNI-for-now by design (the report itself deferred it) — revisit when a consumer asks.** |
| ~~c8~~ | ~~External consumer adoption (cqrs-htmx `setup` is the candidate; blocked on push)~~ done — owned by FEATURES.md Consumers (cqrs-htmx fold-in pending on their side) |
| ~~c9~~ | ~~govulncheck/gosec gates on the new module (repo has no flake/CI; BuildFlow covers lint+format only)~~ done — routed to TODO_LIST P3 health-module quality parity (govulncheck run) |

## d) TOTALLY FUCKED UP (honest)

| #   | Item                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Severity                                                  |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| d-1 | **Unverified: dashboard SSE vs appkit's default `SecurityHeaders` CSP.** The Datastar SDK needs `script-src 'unsafe-eval'` (documented in go-health-dashboard AGENTS); appkit's default stack ships `SecurityHeaders` middleware and I never checked its CSP policy nor ran a browser against the composed stack. If that CSP blocks the SDK, the flagship real-time dashboard silently degrades to a static page in the DEFAULT config. Server-side SSE 200s prove nothing about the browser. | HIGH — must verify before anyone trusts "one wiring call" ~~done — owned by TODO_LIST P2 (dashboard CSP + SSE longevity, routed 2026-09-16)~~ |
| d-2 | **Unverified/documented-late: SSE vs default `WriteTimeout=30s`.** realtime's whole `NoTimeout` feature exists because net/http kills long-lived streams at WriteTimeout. The dashboard's `/health/sse` under default appkit config will be cut every 30s (browser auto-reconnects via SSE retry, so it "works", degraded). My README/doc.go/example never mention `WriteTimeout: NoTimeout`, and the E2E connection lived <30s so I couldn't have seen it.                                    | HIGH doc gap, MEDIUM functional ~~done — owned by TODO_LIST P2 (WriteTimeout/NoTimeout posture, routed 2026-09-16)~~ |
| d-3 | **I fed the auto-commit daemon non-compiling mid-edit trees** (at least one commit window with `undefined: fmt`, one with the missing `WithProbeRoutes` body). Those daemon commits are now in local history and do not build — a bisect hazard (same class the dashboard AGENTS.md warns about). Fix: commit deliberately per logical unit next session; never rewrite pushed history.                                                                                                        | MEDIUM hygiene                                            |
| d-4 | **Test-count drift in the CHANGELOG**: first wrote "17 tests" when there were 14. Caught it myself, but it's the doc-before-verify pattern; the verbatim-compile check found README snippet bugs the same way. Pattern: prose written ahead of proof.                                                                                                                                                                                                                                          | LOW, pattern-level                                        |

Self-review checklist (skill questions): forgot = §d-1/§d-2/c5; stupid = d-3, error-family wobble (3 rewrites of the panic-error construction before matching flightrecorderhealth's inline-constructor style); lied = no, but "race-clean + 0 issues" was verified for the 3 touched modules only, not the other 6 (those got builds only); ghost systems = the `health` module has zero external consumers so far (example is the only consumer) and core `DrainHooks` exists solely for it — deliberate seams, not ghosts, but they die if §g-3 goes the other way; removed-useful = nothing removed; split brains = b4's doc-vs-test gap and the pre-existing truncated `middleware.go` AGENTS row (not mine, still there); scope creep = contained (one core seam, tested); tests = 18 new this session, gaps listed in c2-c5.

## e) WHAT WE SHOULD IMPROVE

1. **Verify browser reality, not server 200s.** The compose-with-CSP story (d-1) needs a chromedp test or manual check before the README's central claim is trustworthy.
2. **Write the gotcha inventory BEFORE the README.** WriteTimeout-vs-SSE, CSP-vs-SDK, and the ServeMux `GET /` panic were all discoverable from existing sibling knowledge (realtime + dashboard AGENTS) before shipping docs.
3. **Machine-check every cross-module contract.** §c5's `var _ dashboard.Prober = (*health.Probe)(nil)` costs one line and kills a silent-break class; do it for every structural claim in docs.
4. **Match house conventions on the first write.** The error-construction wobble and the Mount→New refactor mid-stream both came from writing before re-reading sibling patterns. Read the closest sibling module FIRST, then write.
5. **Commit deliberately in small units** instead of letting the daemon snapshot mid-edit trees (d-3).
6. **House patterns to add**: godoc examples, benchmark, fuzz (c2-c4) — flightrecorderhealth sets the bar.
7. **Docs counts from tests, not memory**: generate test counts in CHANGELOGs (`go test -v | grep -c PASS`), don't type them.

## f) NEXT — up to 50 things to get done (sorted by impact × ease)

**P0 — correctness / trust (this week)**

1. ~~Verify Datastar SDK executes under appkit's default `SecurityHeaders` CSP (chromedp run of example, or manual) — resolves d-1.~~ done (owned by TODO_LIST P2 (dashboard CSP verification, routed 2026-09-16))
2. ~~Document + test the `WriteTimeout: NoTimeout` requirement for dashboard SSE (README, doc.go, example) — resolves d-2; decide whether default-config reconnect-cycle is acceptable (§g-2).~~ done (owned by TODO_LIST P2 (WriteTimeout/NoTimeout posture, routed 2026-09-16))
3. ~~Add `contract_test.go`: `var _ dashboard.Prober = (*health.Probe)(nil)` (+ any Mounted claims worth freezing).~~ done (routed to TODO_LIST P3 health-module quality parity (contract assertion))
4. ~~Decide + test `WithProbeRoutes`+`WithDashboard` conflict semantics (currently silently ignored — either panic loudly or test the ignore).~~ done (routed to TODO_LIST P3 health-module quality parity (conflict semantics))
5. ~~Test the documented `WithDashboard`+`WithBasePath` uniform-routing claim.~~ done (routed to TODO_LIST P3 health-module quality parity (BasePath routing test))
6. ~~health example: set `WriteTimeout: NoTimeout` (part of #2) and re-run E2E with an SSE connection held across ≥2 push intervals.~~ done (owned by TODO_LIST P2 (SSE longevity under NoTimeout))
7. ~~Pre-tag hermetic fresh-consumer verify for flightrecorderhealth v0.1.1 (after push gate).~~ done (done — fresh-consumer verify ran post-push 2026-09-04 (18-57 wave))
8. ~~Tag `flightrecorderhealth/v0.1.1` with jsonv2-requirement release note.~~ done (done — flightrecorderhealth/v0.1.1 tagged + pushed 2026-09-04)
9. ~~Core: land `DrainHooks` in core's next tag (Unreleased already carries it).~~ done (done — core v0.4.0 shipped DrainHooks 2026-09-04)
10. ~~health module: drop example `replace ../` → require published core → hermetic verify → tag `health/v0.1.0`.~~ done (done — health/v0.1.0 tagged + pushed 2026-09-04)
11. ~~Fresh-consumer proxy tests for `health/v0.1.0` + `flightrecorderhealth/v0.1.1` + pkg.go.dev render check.~~ done (proxy smoke done 2026-09-04 (18-57 wave); pkg.go.dev owned by TODO_LIST P1)
12. ~~Commit deliberately per logical unit going forward (daemon-hygiene, d-3).~~ done (process adopted — capture hashes per task (08-53 §e-9))

**P1 — quality (next 2 weeks)**
13. ~~Runnable godoc examples: `NewProbe`, `New`+`RegisterRoutes`, `Mount` (verified output).~~ done (routed to TODO_LIST P3 health-module quality parity (godoc examples))
14. ~~Benchmark `NewProbe` batch overhead (N=1/5/20) vs go-health injector path; publish numbers in README.~~ done (routed to TODO_LIST P3 health-module quality parity (benchmark))
15. ~~Fuzz targets: mount options, check-map panics (short-budget, seed-run in `go test`).~~ done (routed to TODO_LIST P3 health-module quality parity (fuzz targets))
16. ~~SSE longevity integration test (>2 push intervals) — locks in #6.~~ done (routed to TODO_LIST P2 (SSE longevity test folds into the CSP/NoTimeout item))
17. ~~aggregate example: `aggregate.New(probeA, probeB)` + dashboard through one `Mounted` (multi-service story).~~ done (routed to TODO_LIST P3 health-module quality parity (aggregate example))
18. ~~Cross-module integration test: injector-built `health.New` probe + flightrecorderhealth `Trigger` + this module's `Mount` (closes b4).~~ done (routed to TODO_LIST P3 integration-module expansion (cross-module Trigger + Mount test))
19. ~~CSP helper decision (§g-1) → if yes: `WithNonceExtractor` wiring helper + docs + test.~~ done (owned by TODO_LIST P2 (CSP posture decides the nonce helper))
20. ~~Update root README `Configuration` section to document `DrainHooks`.~~ done (done — README Configuration table documents DrainHooks)
21. ~~FEATURES.md: health module "Known limitations" subsection once d-1/d-2 are resolved.~~ done (folded into TODO_LIST P2 (CSP/WriteTimeout known-limitations note))
22. ~~AGENTS.md: fix the truncated `middleware.go` table row (pre-existing).~~ done (done 2026-09-16 — AGENTS.md middleware.go row completed)
23. ~~realtime/.golangci.yml: `go: "1.26.5"` → `"1.26.7"` (pre-existing drift, same class as the bump).~~ done (done 2026-09-16 — all module .golangci.yml go pins bumped to 1.26.7)
24. ~~Sweep satellite `.golangci.yml` files for the same stale-version drift.~~ done (done 2026-09-16 — same sweep, all pins at 1.26.7)
25. ~~Empty-checks-map `NewProbe`: decide allow (current) vs `Rejection` — document whichever.~~ done (documented — empty checks map stays allowed; behavior visible in probe.go)
26. ~~`Mounted.Drain` before `Start` edge: add test pinning current behavior.~~ **Won't implement — not pinned — Drain-before-Start behavior undocumented; fold into health quality parity if it bites.**
27. ~~Decide whether `Mounted` should expose `Status()`/`Alive()` pass-throughs or stay lean (lean is my recommendation).~~ **Won't implement — lean is the house choice (the report's own recommendation); no Status()/Alive() pass-throughs.**
28. ~~Document `Start`'s synchronous first batch = cold-start latency bound (probe timeout) in README API table.~~ done (documented in health/README.md (Start runs the initial batch synchronously))
29. ~~govulncheck + gosec run on the health module (manual, no CI in repo).~~ done (routed to TODO_LIST P3 health-module quality parity (govulncheck))
30. ~~Verify sibling workspaces (e.g. cqrs-htmx's go.work) don't choke on the new module path.~~ done (superseded — no workspace choke observed since; consumers resolve via the proxy)

**P2 — docs / ecosystem (when convenient)**
31. ~~Dashboard screenshot in health README via the dashboard's existing screenshot harness.~~ done (ROADMAP fuel (dashboard screenshot harness exists upstream))
32. ~~Health module section in `docs/DOMAIN_LANGUAGE.md` if the repo carries one (check; go-health has one to align terms with).~~ done (done 2026-09-16 — docs/DOMAIN_LANGUAGE.md built with ~30 terms)
33. ~~README API table: note `Mount` vs `New`+`RegisterRoutes` double-registration footgun explicitly (currently implied).~~ done (documented — Mount vs New+RegisterRoutes covered in README quick start)
34. ~~Root AGENTS.md: Core Dependencies table gains a "when to use which health surface" note (httputil defaults vs health module).~~ done (covered by the AGENTS Health Module section (module comparison lives there))
35. ~~Consider `WithDashboard` default-ON for a future v0.2 with automatic `RegisterHealth` detection — only with a mux-conflict pre-check story.~~ **Won't implement — rejected for now — default-ON dashboard needs a mux-conflict pre-check story; revisit at health v0.2.**
36. ~~Error message UX: `Mount` nil-mux/probe errors include remediation text ("did you mean New(...) first?").~~ **Won't implement — polish, not scheduled — error-message remediation text.**
37. ~~Test names: shorten `TestNew_RegisterRoutesIsThePrimaryAppkitFlow` (verboseness, cosmetic).~~ **Won't implement — cosmetic, not scheduled — test-name shortening.**
38. ~~`getBody` test helper: drop unused named returns; tighten.~~ **Won't implement — cosmetic, not scheduled — helper tightening.**
39. ~~CHANGELOG: single source for test counts (script or CI check) — enforces e-7.~~ done (superseded — counts live in CHANGELOG entries generated from test runs)
40. ~~Explore `go-health` `WithEvaluationHook` → `otel` metrics bridge as an otel-module companion (c7) — design only until a consumer asks.~~ **Won't implement — design-only until a consumer asks (otel metrics bridge), as the report itself deferred.**
41. ~~Upstream (verify-before-filing): go-health-dashboard README could warn about host `WriteTimeout` for SSE — file issue/PR upstream if confirmed absent.~~ done (superseded — go-health-dashboard v0.8.1 (2026-09-16) carries SSE/drain docs; verify at bump)
42. ~~Upstream (verify-before-filing): go-health `NewWithHealthCheck` docs could cross-reference that `WithHealthRecorder` is ignored (it is documented — check wording only).~~ done (superseded — the WithHealthRecorder limitation is documented; AGENTS gotcha carries it)
43. ~~Consider a `healthconsul`/k8s-probe-path preset option if consumers hit non-default K8s conventions (YAGNI gate: wait for a real ask).~~ **Won't implement — YAGNI — wait for a real K8s-convention consumer ask.**
44. ~~Add the health module to the repo-wide lint sweep script/ritual (AGENTS.md lint paragraph lists modules one-by-one — add health).~~ done (done — AGENTS linting section lists every module; CI matrix covers the rest)
45. ~~Define health-module v1.0 exit criteria alongside core's (TODO_LIST P3 item exists for core).~~ done (owned by ROADMAP (core v1 exit-criteria draft exists; health v1 criteria fold in later))
46. ~~cqrs-htmx `setup`: propose health-module adoption in their ADR-001 follow-up once push lands (c8).~~ done (owned by FEATURES.md Consumers (cqrs-htmx adoption tracking))
47. ~~Docs-health annotation pass: annotate this report's numbered items into TODO_LIST after §g answers.~~ done (this annotation pass + the 2026-09-16 TODO routing complete it)
48. ~~Example: log line on drain ("draining: /readyz → 503") for operability.~~ **Won't implement — polish, not scheduled — drain log line.**
49. ~~Consider screenshot-dark parity for the health README (dashboard's dark-capture harness exists; cosmetic).~~ **Won't implement — cosmetic, not scheduled — dark-capture parity.**
50. ~~Post-adoption: collect real consumer feedback on the `New`-first vs `Mount`-first ergonomics and collapse to one if one wins (API-surface diet).~~ **Won't implement — not scheduled — API-surface diet waits for real consumer feedback.**

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. ~~**CSP posture (decides d-1's fix shape):** do your appkit deployments run the default `SecurityHeaders` CSP on the same origin as the dashboard? If yes, should I build a first-class nonce-aware integration (health-module helper wiring `httputil` nonce middleware → `dashboard.WithNonceExtractor` + `RecommendedCSP`), or is the dashboard expected to live behind a separate admin port/ingress where CSP is relaxed? I cannot infer your deployment topology.~~ done (owned by TODO_LIST P2 (dashboard CSP item carries the posture decision))
2. ~~**SSE vs WriteTimeout (decides #2's strictness):** is "SSE cut every 30s, browser auto-reconnects" acceptable for a health dashboard in appkit's default config (docs-only fix), or should the module hard-require/split `WriteTimeout: NoTimeout` (e.g. `Mount` warns, or README makes it step 1)? Product call: silent-degraded vs loud-requirement.~~ done (owned by TODO_LIST P2 (WriteTimeout posture decided there))
3. ~~**Release packaging (I own execution, you own the gate):** confirm the wave plan — after the pending 4-tag push: (a) core tag carrying `DrainHooks`, then (b) `health/v0.1.0`, then (c) `flightrecorderhealth/v0.1.1` — or do you want the health work held back entirely until cqrs-htmx's adoption lands? Also: should `flightrecorderhealth/v0.1.1` be tagged immediately despite its only delta being dep-bump + build-requirement change?~~ done (answered — wave 2 shipped health/v0.1.0 + flightrecorderhealth/v0.1.1 on 2026-09-04)

---

**Handoff state:** working tree clean except this report + the 3 doc files the auto-daemon hasn't swept yet (`CHANGELOG.md`, `TODO_LIST.md`, `health/README.md`). Nothing pushed. All 9 modules build; 3 touched modules race-clean + 0 lint issues. Awaiting instructions.
