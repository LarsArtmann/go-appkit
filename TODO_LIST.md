# TODO List — go-appkit

> Short-term, actionable, bounded work. Open items only.
> Completed work lives in [CHANGELOG.md](CHANGELOG.md). Long-term vision lives in `docs/planning/`.

**Updated:** 2026-08-16 | **Modules:** 7 (core, cqrs, realtime, flightrecorder, flightrecorderhealth, errorpages, docs) | **Release state:** core **v0.3.0**, cqrs **v0.3.0**, realtime **v0.1.0**, flightrecorder **v0.1.0** cut at `f938d65` (2026-08-16), hermetically verified, **push pending — user gate**; errorpages v0.1.0 and docs v0.2.0 current (their since-tag deltas are test-only / config-only — deliberately not re-tagged); flightrecorderhealth **v0.1.0 uncut — work-in-progress** | **Six modules require `GOEXPERIMENT=jsonv2`** (flightrecorderhealth does NOT — its deps are plain `encoding/json`)

## Status Legend

- [ ] **OPEN** — actionable, not yet started.
- [~] **PARTIALLY DONE** — started but incomplete.

> No `[x]` items here. When a task finishes, it moves to the module CHANGELOG and is removed from this list.

---

## P1 — High impact (release follow-through)

- [ ] **Push the prepared release wave (USER GATE).** `git push origin master && git push origin v0.3.0 cqrs/v0.3.0 realtime/v0.1.0 flightrecorder/v0.1.0` — master is 1 commit ahead (`f938d65`, release prep); the 4 tags are local-only (verified via `git ls-remote` delta 2026-08-16). This unblocks the cqrs-htmx `setup` adoption (their ADR-001; see their TODO_LIST P3).
- [ ] **Post-push verification for all 4 tags.** (a) Fresh-consumer proxy test per module: clean module in /tmp → `go get <module>@<version>` → blank-import `main.go` → `go build` (the honest smoke test — do not reference symbols that might not exist). (b) pkg.go.dev renders each new version (expect 2-10 min proxy propagation). Source: go-release skill Phases 6-7; carried from session-5 #30/#31.

---

## P2 — Medium impact (quality & docs)

- [ ] **Decide the logging posture (comparison finding 7).** The cqrs-htmx spike benchmark delta (16,178 vs 45,049 ns/op, ~2.8x) is entirely per-request INFO logging. Options: log-level config beyond `LogLevel`, sampling, or letting consumers wire their own logger. Decide + implement + benchstat before/after. Source: `docs/review/2026-08-16_setup-vs-go-appkit-comparison.md` finding 7 (cqrs-htmx repo view: spike `run_appkit_test.go:236`).
- [ ] **realtime: SSE-flush E2E test** mirroring the cqrs-htmx spike's flush-through-middleware test (headers arrive before first event, through the default stack). Source: session-5 #21; module's own `handler.go` flush claim deserves a lock-step test.
- [ ] **README: document `GOEXPERIMENT=jsonv2` for building from source.** AGENTS.md carries per-module reasons; README (user-facing) does not mention it. Source: session-5 #19.
- [ ] **FEATURES.md: add a "Consumers" section** citing the cqrs-htmx `setup` spike/adoption (ADR-001) as the reference consumer. Source: session-5 #25.
- [ ] **Fix `docs/planning/design-decisions.md:118` lychee 404 + MD013 long lines.** Source: session-5 #13/#14.
- [ ] **`httpspec_test.go` init-function refactor** (`gochecknoinits`/`godoclint` flag it; BuildFlow root run 2026-08-16). Source: session-5 #34.
- [ ] **Document the `DrainDelay: 0` test-ergonomics pattern in AGENTS.md** (how to make shutdown/drain tests fast). Source: session-5 #22/#23.
- [ ] **Verify `example/main.go` quick start compiles with plain `go build`** (README quick-start honesty). Source: session-5 #23.
- [ ] **Add a mechanical API-break check to the release process** (goapidiff or a `go doc` snapshot diff at tag time) — v0.2.0 was reconstructed after the fact; never again. Source: session-5 #7.
- [ ] **Bump Go toolchain 1.26.5 → 1.26.6 when nixpkgs carries it** (GO-2026-6090 crypto/tls post-handshake flood, GO-2026-5972 encoding/asn1 recursion; both symbol-level findings, patch-class DoS; builds run GOTOOLCHAIN=local so the nixpkgs lock gates this). Source: govulncheck via BuildFlow 2026-08-16; same item tracked in cqrs-htmx TODO_LIST P2.

---

## P3 — Technical debt & tooling

- [ ] **BuildFlow dprint step fails on CHANGELOG-only commits** (exit 14 "no files found to format"): `dprint.json` excludes `**/CHANGELOG.md`, so a commit staging only CHANGELOGs (+ non-dprint files) trips the formatter's empty-set error. Fix: `--allow-no-files` (or skip when no files match). Evidence: the v0.3.0-wave commit needed `--no-verify` on 2026-08-16 (deterministic, retried once).
- [ ] **golangci depguard allowlists per satellite module.** Family imports (go-sse, go-flightrecorder, templ-components/errorpage, go-cqrs-lite/…) are flagged "not allowed from list 'main'" across modules (counts at `f938d65`: cqrs 54, realtime 31, flightrecorder 39, errorpages 26, docs 3, root 14 — pre-existing, warnings). Either configure per-module lists or record the accepted-warning rationale.
- [ ] **go-structure-linter root-package findings:** 8 standing "package file at project root" errors — the root package IS the intended public layout for core; configure the linter to accept it instead of living with red.
- [ ] **Define v1.0.0 exit criteria for core** (AGENTS.md names v1.0.0 as the core target; no written criteria exist). Ideas live in `docs/planning/framework-architecture.md`; graduate to actionable when the consumer count grows.

---

_For completed work, see each module's `CHANGELOG.md` and `git log`._
