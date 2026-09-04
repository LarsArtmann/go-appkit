# TODO List — go-appkit

> Short-term, actionable, bounded work. Open items only.
> Completed work lives in [CHANGELOG.md](CHANGELOG.md). Long-term vision lives in `docs/planning/`.

**Updated:** 2026-09-04 | **Modules:** 9 (core, cqrs, realtime, otel, flightrecorder, flightrecorderhealth, health, errorpages, docs) | **Release state:** core **v0.3.0**, cqrs **v0.3.0**, realtime **v0.1.0**, flightrecorder **v0.1.0**, flightrecorderhealth **v0.1.0** cut (flightrecorderhealth at `d3e3e51`, 2026-08-16, hermetically verified incl. 100% coverage + compile-time contract assertions), **push pending — user gate**; errorpages v0.1.0 and docs v0.2.0 current (their since-tag deltas are test-only / config-only — deliberately not re-tagged); **otel v0.1.0 and health v0.1.0 unreleased** — the 2026-08-18 OTEL work and the 2026-09-04 health work (core `DrainHooks` + new health module + flightrecorderhealth go-health v0.1.1 bump) sit uncommitted/untagged pending the push gate; **eight of nine modules require `GOEXPERIMENT=jsonv2`** (only otel does not)

## Status Legend

- [ ] **OPEN** — actionable, not yet started.
- [~] **PARTIALLY DONE** — started but incomplete.

> No `[x]` items here. When a task finishes, it moves to the module CHANGELOG and is removed from this list.

---

## P1 — High impact (release follow-through)

- [ ] **Push the prepared release wave (USER GATE).** `git push origin master && git push origin v0.3.0 cqrs/v0.3.0 realtime/v0.1.0 flightrecorder/v0.1.0` — master is 1 commit ahead (`f938d65`, release prep); the 4 tags are local-only (verified via `git ls-remote` delta 2026-08-16). This unblocks the cqrs-htmx `setup` adoption (their ADR-001; see their TODO_LIST P3).
- [ ] **Post-push verification for all 4 tags.** (a) Fresh-consumer proxy test per module: clean module in /tmp → `go get <module>@<version>` → blank-import `main.go` → `go build` (the honest smoke test — do not reference symbols that might not exist). (b) pkg.go.dev renders each new version (expect 2-10 min proxy propagation). Source: go-release skill Phases 6-7; carried from session-5 #30/#31.
- [ ] **Commit the OTEL work (USER GATE on strategy).** Working tree holds the full 2026-08-18 OTEL change set: core telemetry hooks (`OuterMiddlewares`, `ShutdownHooks`, `NoDrainDelay` + test-suite speedup ~30s→6s), the new `otel` module (23 tests, race-clean, E2E-verified live), cqrs `OTelProjectionMetrics` (closes M10), and all doc updates. One commit vs per-module commits is status report §g Q1. Commits include `.go` files, so the dprint exit-14 gotcha should not trip.
- [ ] **Tag `otel/v0.1.0` (after core v0.3.0 is pushed).** Pre-tag: drop the example's `replace github.com/larsartmann/go-appkit => ../` in `otel/go.mod`, require published core, hermetic verify (`cd otel && GOWORK=off go test ./... -race -count=1`), then annotated tag. Post-tag: fresh-consumer proxy test + pkg.go.dev. Wave membership (join pending 4-tag push vs second wave) is status report §g Q2.
- [ ] **Release the health work (2026-09-04) after the pending push.** (a) Core `DrainHooks` ships in core's next tag (Unreleased). (b) New `health` module → `health/v0.1.0`: drop the example's `replace ../`, hermetic verify, annotated tag, fresh-consumer proxy test. (c) flightrecorderhealth go-health v0.1.1 bump → `flightrecorderhealth/v0.1.1` (build contract changed: now requires `GOEXPERIMENT=jsonv2` — call it out in the release notes).

---

## P2 — Medium impact (quality & docs)

- [ ] **Decide the logging posture (comparison finding 7).** The cqrs-htmx spike benchmark delta (16,178 vs 45,049 ns/op, ~2.8x) is entirely per-request INFO logging. Options: log-level config beyond `LogLevel`, sampling, or letting consumers wire their own logger. Decide + implement + benchstat before/after. Source: `docs/review/2026-08-16_setup-vs-go-appkit-comparison.md` finding 7 (cqrs-htmx repo view: spike `run_appkit_test.go:236`).
- [ ] **realtime: SSE-flush E2E test** mirroring the cqrs-htmx spike's flush-through-middleware test (headers arrive before first event, through the default stack). Source: session-5 #21; module's own `handler.go` flush claim deserves a lock-step test.
- [ ] **README: document `GOEXPERIMENT=jsonv2` for building from source.** AGENTS.md carries per-module reasons; README (user-facing) does not mention it. Source: session-5 #19.
- [ ] **FEATURES.md: add a "Consumers" section** citing the cqrs-htmx `setup` spike/adoption (ADR-001) as the reference consumer. Source: session-5 #25.
- [ ] **Fix `docs/planning/design-decisions.md:118` lychee 404 + MD013 long lines.** Source: session-5 #13/#14.
- [ ] **Upstream: cqrs-lite otel `Provider.Shutdown` must ForceFlush before Shutdown.** Same batch-queue race fixed locally in go-appkit/otel — spans ended moments before Shutdown sit in the async queue and are silently dropped. Probe-test evidence in the 2026-08-18 OTEL session. Issue vs PR is status report §g Q3 (verify-before-filing applies).
- [ ] **Sweep satellite module tests for `DrainDelay: 0` misuse.** Zero applies the 5s default (hidden test tax); core's conversion to `NoDrainDelay` dropped wall time ~30s→6s. realtime/errorpages/docs/flightrecorder likely carry the same pattern.
- [ ] **Add a mechanical API-break check to the release process** (goapidiff or a `go doc` snapshot diff at tag time) — v0.2.0 was reconstructed after the fact; never again. Source: session-5 #7.
- [ ] **Bump Go toolchain 1.26.5 → 1.26.6 when nixpkgs carries it** (GO-2026-6090 crypto/tls post-handshake flood, GO-2026-5972 encoding/asn1 recursion; both symbol-level findings, patch-class DoS; builds run GOTOOLCHAIN=local so the nixpkgs lock gates this). Source: govulncheck via BuildFlow 2026-08-16; same item tracked in cqrs-htmx TODO_LIST P2.

---

## P3 — Technical debt & tooling

- [ ] **BuildFlow dprint step fails on CHANGELOG-only commits** (exit 14 "no files found to format"): `dprint.json` excludes `**/CHANGELOG.md`, so a commit staging only CHANGELOGs (+ non-dprint files) trips the formatter's empty-set error. Fix: `--allow-no-files` (or skip when no files match). Evidence: the v0.3.0-wave commit needed `--no-verify` on 2026-08-16 (deterministic, retried once).
- [ ] **go-structure-linter root-package findings:** 8 standing "package file at project root" errors — the root package IS the intended public layout for core; configure the linter to accept it instead of living with red.
- [ ] **otel: middleware benchmark (no-op vs configured) + benchstat** — quantify overhead, record numbers in module README. Status report §f P3-15.
- [ ] **Propose httputil `Logging` emit with request context** so completion lines correlate with spans via `TraceHandler` (today only handler-level logs correlate). Status report §f P2-12.
- [ ] **Define v1.0.0 exit criteria for core** (AGENTS.md names v1.0.0 as the core target; no written criteria exist). Ideas live in `docs/planning/framework-architecture.md`; graduate to actionable when the consumer count grows. `OuterMiddlewares`/`ShutdownHooks` are v1-shaped — fold into criteria. Deeper OTEL backlog lives in `docs/status/2026-08-18_12-45_otel-module-and-telemetry-hooks.md` §f.
- [ ] **Revisit cordis bridge module when triggers fire** (researched 2026-09-04, verdict: NOT NOW). `docs/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md` — all three required: (1) cordis tags `go/v0.1.0` (today proxy-pinned to a pseudo-version), (2) a real consumer states the reactive-composition requirement, (3) core v1.0.0 exit criteria shipped. Bridge sketch (~300-500 LOC, zero third-party deps) is in the doc §4 Option B. go-plugin-mvp marketplace integration: rejected as appkit dependency (wrong layer, pre-1.0 churn); recommended reverse adoption (their repo embeds appkit, cqrs-htmx ADR-001 pattern) — proposal to their TODO_LIST is user-gated.

- [ ] **Upgrade installed cqrs-lint 4.6.0 → 4.8.1** (`go install github.com/larsartmann/go-cqrs-lite/cmd/cqrs-lint@latest` from the cqrs-lite repo). Verified 2026-09-04: 4.8.1 reports clean on `cqrs/` (both inline suppressions still fire), but it rejects the `./...` argument form — use `cqrs-lint .` inside the module. Source: `docs/research/2026-09-04_go-cqrs-lite-deep-dive.html`.
- [ ] **cqrs: EventConfig opt-ins for event encryption/signing when a consumer demands it** (encryption/v4, signing/v4; candidates: idempotency/sqlstore, scheduling). Deliberately NOT added now — wrapper surface stays demand-driven. Routed from `docs/research/2026-09-04_go-cqrs-lite-deep-dive.html` finding 5.

---

_For completed work, see each module's `CHANGELOG.md` and `git log`._
