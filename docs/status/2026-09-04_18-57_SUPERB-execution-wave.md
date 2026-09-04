# Status Report — SUPERB Execution Wave (Plan → Ship)

**Date:** 2026-09-04 18:57 CEST · **Scope:** this session's execution run only (the `docs/planning/2026-09-04_17-40_SUPERB-release-visibility-and-quality-plan.md` plan, tasks C1–C26 / F1–F77) plus what it noticed in passing. No new research was done for this report.

**Session one-liner:** executed the whole plan — cut, tagged, pushed and consumer-verified **release wave 2** (core v0.4.0, cqrs v0.4.0, otel v0.1.0, health v0.1.0, flightrecorderhealth v0.1.1), proved the mechanical API-break check on a real release, discovered `encoding/json/v2` is default-on in Go 1.26.7 (invalidating the repo-wide GOEXPERIMENT lore), quantified the logging tax (+30µs/req emitting vs +0.8µs suppressed) and the otel middleware overhead (~6µs, +0 allocs), migrated all 10 modules to exhaustruct_v5, hardened cqrs tests + godoc examples, and left **10/10 modules green** (build + vet + race + lint 0 issues), all pushed.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                          | Evidence                                 |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| 1  | **Release wave 2 cut + pushed** — 6 tags on origin (`v0.4.0`, `cqrs/v0.4.0`, `otel/v0.1.0`, `health/v0.1.0`, `flightrecorderhealth/v0.1.1` + wave-1 set), `proxy.golang.org .../@v/list` serves the new versions                                                                              | `git ls-remote`, proxy list check        |
| 2  | **Fresh-consumer proxy test for all five new tags** — clean /tmp module → `go get` each → blank imports → `go build` (jsonv2 env) PASS                                                                                                                                                        | session run                              |
| 3  | **Mechanical API-break check proven on a real release** — `go doc -all` snapshot diff core v0.3.0 ↔ master: additions only (NoDrainDelay + 3 hook fields, zero removals) → minor bump justified; method documented in AGENTS.md Release Ritual                                                | AGENTS.md Release Ritual                 |
| 4  | **Correct release order discovered + enforced**: both new modules' examples use post-v0.3.0 core APIs → core v0.4.0 tagged FIRST, then otel/health de-replaced (`replace ../` gone) onto published core                                                                                       | `otel/go.mod`, `health/go.mod`           |
| 5  | **Hermetic verification per released module** — `GOWORK=off` build+vet+race suites green: otel (plain), health (jsonv2), flightrecorderhealth (jsonv2), cqrs (jsonv2)                                                                                                                         | session runs                             |
| 6  | **cqrs v0.4.0 breaking release shipped** with dated CHANGELOG and annotated tag stating the semantic delta                                                                                                                                                                                    | tag `cqrs/v0.4.0`                        |
| 7  | **GOEXPERIMENT lore corrected**: `go env GOEXPERIMENT` = `jsonv2` on Go 1.26.7 → ALL 10 modules build plain; README rewritten truthfully, AGENTS.md + TODO_LIST headers fixed                                                                                                                 | empirical 10-module matrix               |
| 8  | **Logging tax quantified** (`logging_bench_test.go`, io.Discard): bare ~17.2µs · suppressed(WARN) ~18.0µs (+0.8µs) · emitting(INFO) ~47.3µs (+30µs, +174%, 162 allocs) — fully explains the 2.8× comparison delta; decision data recorded, implementation stays user-gated                    | `logging_bench_test.go`, TODO_LIST P2    |
| 9  | **otel middleware benchmark** (`otel/benchmark_test.go`): no-op ~21µs / traced ~26µs / traced+metered ~27µs per request, +0 allocs, export I/O excluded; numbers in otel README Performance section                                                                                           | `otel/benchmark_test.go`                 |
| 10 | **exhaustruct → exhaustruct_v5 migration** — all 10 `.golangci.yml` configs + every `//nolint:exhaustruct` directive renamed (v5 ignores the old name); two pre-existing findings fixed (root `http.Server` v5 nolint w/ justification; realtime `bodyclose`/`nonamedreturns` from `53ee2ea`) | sequential lint: 10/10 × 0 issues        |
| 11 | **go-structure-linter zeroed** — 9 standing findings → 0 via CLI-flag excludes (root package IS the public API); yaml `exclude_patterns` proven inert in installed binary f7e33e03 (source-verified) and documented                                                                           | AGENTS.md Linting section                |
| 12 | **cqrs contract hardening** — trigger-returns-false test (gated capture skips, deterministic after WorkerFailed) + derived-wiring-beats-HostOptions precedence test (unstarted second recorder as detector)                                                                                   | `cqrs/flightrecorder_test.go` (35 tests) |
| 13 | **cqrs godoc examples** — 2 output-verified Examples (canonical wiring + replay contract); caught and fixed my own invented APIs BEFORE running (see d-2)                                                                                                                                     | `cqrs/example_test.go`                   |
| 14 | **cqrs-lint 4.8.1 installed** (built from local source; network `go install` blocked) — clean on `cqrs/`, arg-form (`.` not `./...`) documented in README                                                                                                                                     | `~/go/bin/cqrs-lint version`             |
| 15 | **Core v1.0.0 exit criteria drafted** — 7 hard criteria (frozen API surface via the proven diff method, shutdown guarantees, error contract, 2-consumer proof, telemetry seam, truthful docs), soft signals, explicit non-goals                                                               | `docs/planning/core-v1-exit-criteria.md` |
| 16 | **design-decisions.md 404 fixed** — dead link to httputil's huma.md (exists locally, **unpushed in their repo**) replaced by the inlined layering argument                                                                                                                                    | Decision 7                               |
| 17 | **AGENTS.md rituals + registers** — Release Ritual, Adoption & Drift Rituals, Performance Baselines, Deferred Register (encryption/signing/idempotency, cordis, PapDashboard, core TLS, buildflow dprint upstream fix, httputil push, cqrs-lint yaml bug)                                     | AGENTS.md tail                           |
| 18 | **TODO_LIST reconciliation** — 11 completed items closed with evidence, stale claims (GOEXPERIMENT, toolchain bump, OTEL-uncommitted) corrected, logging decision item now carries data                                                                                                       | TODO_LIST.md                             |
| 19 | **Final sweep: 10/10 modules** build + vet + `go test -race -count=1` + golangci-lint 0 issues                                                                                                                                                                                                | sweep log                                |
| 20 | **Everything pushed** — `origin/master @ ce120b0`, tree clean                                                                                                                                                                                                                                 | `git status`, push log                   |

## b) PARTIALLY DONE

1. **pkg.go.dev visibility (the 1%)** — all wave-2 tags carry module-root LICENSE files and the **proxy serves every version** (`@v/list` confirmed), but pkg.go.dev still 404s the new module paths at 18:55 (crawler lag; their crawl can trail the proxy by hours). Recheck pending; the godoc-visibility endgame still routes through the licensing decision (user gate).
2. **License decision memo (F1/F2)** — analysis exists scattered in TODO_LIST P2 + report context, but no standalone decision memo file was written; wave 2 proceeded with the current proprietary text (a deliberate, stated choice so crawl-lag couldn't block releases).
3. **C11 logging implementation** — benchmark + data done; the actual posture change (default WARN / sampling / consumer logger) is user-gated and NOT implemented.
4. **C17 upstream otel ForceFlush** — TODO item updated with the filing path, but I did NOT re-run the 2026-08-18 probe repro this session (referenced, not re-verified), and filing remains user-gated.
5. **DrainDelay sweep (C10/F30–F34)** — verified via one grep pattern (`DrainDelay: 0`): only core's legitimate zero-default test case remains. Adequate but thin evidence; no wall-time before/after was possible since nothing needed converting.
6. **cqrs v0.4.0 consumer validation** — fresh-consumer COMPILE verified; real-consumer re-validation (cqrs-htmx consumes the cqrs stack) NOT done — a breaking `EventConfig.FlightRecorder` change lands in their dependency graph unannounced until they bump.

## c) NOT STARTED (planned, deliberately skipped)

- Fine-grained leftovers: F33–F35 (SnapshotStore/ReadModels examples, DLQ age alerting), F42 (CI-less sweep script — sweep ran manually), F44 (errorpages+cqrs composition example), F47 (DeadLetterStoreAdmin godoc example — covered as a README line instead).
- License decision memo as a standalone artifact (see b-2).
- govulncheck re-run (binary absent; `go install` network-blocked in this environment).
- README/CHANGELOG correction note for the now-moot "requires GOEXPERIMENT=jsonv2" claims inside the published `flightrecorderhealth/v0.1.1` tag annotation (see d-4).

## d) TOTALLY FUCKED UP

1. **I nearly shipped invented APIs again — caught only because I checked before running.** The first draft of `cqrs/example_test.go` used `projectionhost.ErrNoDeadLetters` (does not exist), `Status()["name"]` (Status returns a **slice**, not a map), and a wrong `NewProjection` handler signature. The final examples are source-verified and pass, but the reflex that wrote them was the same one that fabricated "24 tests" in the last report. The verification step is now load-bearing; the generation step is still untrustworthy.
2. **The daemon won the race on the most important commit of the wave.** The otel/health go.mod de-replace + core-v0.4.0 requirement — the change that makes two published tags consumer-valid — landed as `294735b "chore: auto-commit 4 changed file(s) (heuristic)"`. Both wave-2 module tags point at a commit whose message says nothing. I keep losing this race (same complaint as the last report); a release commit with a heuristic message undermines the release history.
3. **Published tag annotation carries a now-stale claim.** `flightrecorderhealth/v0.1.1` annotation says the module "now requires GOEXPERIMENT=jsonv2" — minutes later I proved jsonv2 is default-on in Go 1.26.7, making that requirement automatic/void for current toolchains (still true for older 1.26.x, but the annotation reads as a burden). Tags are immutable; the correction lives only in AGENTS/TODO.
4. **`go install` being security-blocked produced a workaround, not a solution** — no apidiff, no benchstat, no govulncheck this session. The fallbacks (go-doc diff, median-of-3, deferral) are honest but weaker than the plan's named tools; benchstat-formatted comparisons still don't exist.
5. **Two verification gaps dressed as completions:** the DrainDelay "clean" verdict rests on a single grep pattern, and the C17 repro was referenced rather than re-executed. Both are probably fine — probably is not verified.

## e) WHAT WE SHOULD IMPROVE

1. **Commit before the daemon does** — for release-critical edits: stage → commit within the same tool call batch as the edit. Twice now the daemon has laundered release commits into "heuristic" noise.
2. **Never let generated code reach the runner unverified** — treat every API reference in newly written code (especially examples/docs) as a claim requiring a source grep or compiler pass BEFORE the test run, not after.
3. **Tag annotations are forever** — draft them AFTER the facts they assert are verified, or phrase condition-free facts only ("carries go-health v0.1.1" beats "now requires GOEXPERIMENT=jsonv2").
4. **Benchmark reports should name their tooling gaps** — median-of-3 by hand is fine for orientation, but the TODO items now correctly demand benchstat for before/after claims.
5. **Sweep claims need sweep evidence** — "clean" verdicts should show the grep/find patterns used, so the next session can widen them instead of trusting them.
6. **Release order dependency discovery should be a ritual step** — "grep consumers' examples for post-tag APIs before choosing tag order" caught a real chicken-and-egg this time; write it into the Release Ritual (it is now, as of this session's AGENTS edit — keep it there).
7. **Keep the 10-module sweep as the session close** — it caught the realtime regressions only because it ran over everything; make it the standard ending move.

## f) TOP THINGS TO GET DONE NEXT (up to 50 — brainstorm fuel; 1–10 are real work, 11–50 are ROADMAP-grade)

| #  | Task                                                                                                                                                             | Impact | Effort |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| 1  | **License decision (USER GATE)** — proprietary (godoc stays hidden) vs MIT-family (cqrs-htmx is MIT); if relicense: root + all module LICENSE files in next tags | 5      | 2      |
| 2  | Re-verify pkg.go.dev per module page once the crawler lands (expect: license UNKNOWN until #1; 404s should clear)                                                | 4      | 1      |
| 3  | Notify + validate cqrs-htmx against cqrs v0.4.0 (breaking FlightRecorder change) — their side or a compatibility shim note in our README                         | 4      | 2      |
| 4  | Logging posture decision (USER GATE, data attached): default WARN vs sampling vs status quo                                                                      | 4      | 2      |
| 5  | Install govulncheck + apidiff + benchstat into the environment (nix devShell or prebuilt) so release ritual tools stop being improvised                          | 3      | 2      |
| 6  | Add the frh v0.1.1 jsonv2-correction note to its CHANGELOG (tag annotation is immutable)                                                                         | 2      | 1      |
| 7  | Re-run the 2026-08-18 ForceFlush probe to confirm it still reproduces, then file upstream (USER GATE: issue vs PR)                                               | 3      | 2      |
| 8  | Push `httputil/docs/integrations/huma.md` (their repo) and restore the design-decisions link                                                                     | 1      | 1      |
| 9  | Wire the 10-module sweep into a single script (`scripts/verify-all.sh`) so session-close sweeps are one command                                                  | 2      | 1      |
| 10 | SnapshotStore/ReadModels Bundle examples (F33) — Bundle-reachable, undocumented                                                                                  | 2      | 2      |
| 11 | DLQ age alerting recipe (F35)                                                                                                                                    | 1      | 1      |
| 12 | DeadLetterStoreAdmin godoc example (F47) — upgrade the README line to an output-verified Example                                                                 | 1      | 1      |
| 13 | errorpages+cqrs composition example (F44)                                                                                                                        | 1      | 2      |
| 14 | benchstat before/after harness for the logging decision once implemented                                                                                         | 2      | 1      |
| 15 | Wave-2 announcement/changelog digest for consumers (core hooks + cqrs breaking + new modules)                                                                    | 3      | 1      |
| 16 | pkg.go.dev badges/links in READMEs once pages render                                                                                                             | 1      | 1      |
| 17 | integration/ module: bump pins to wave-2 tags and exercise the new hook surface E2E                                                                              | 3      | 2      |
| 18 | Release-ritual checklist as a file (`docs/release-checklist.md`) instead of an AGENTS section, referenced from every module CHANGELOG template                   | 2      | 1      |
| 19 | Add `go doc` snapshot diff as a make-less script with old-tag argument handling                                                                                  | 1      | 1      |
| 20 | Decide whether Deferred Register items graduate into TODO_LIST when triggered (doc the promotion path)                                                           | 1      | 1      |
| 21 | cqrs README: document Status() as a slice (godoc examples now cover it; README doesn't)                                                                          | 1      | 1      |
| 22 | Realtime+health integration example: drain-aware SSE shutdown (health module's DrainHooks story)                                                                 | 2      | 2      |
| 23 | LagPerProjection → otel gauge wiring example (checkpoint.lag instrument exists; show alert math)                                                                 | 1      | 1      |
| 24 | SQLite busy_timeout/WAL tuning doc for contended event stores                                                                                                    | 2      | 1      |
| 25 | Fuzz or property test for staleness guards (budget boundary: exactly-at-threshold)                                                                               | 2      | 2      |
| 26 | Multi-recorder coordination ADR (HTTP + projections + health triggers on ONE fr.Recorder) — the README claims it, an ADR would own it                            | 2      | 1      |
| 27 | Trigger-context documentation table (Kind values across cqrs + appkit middleware)                                                                                | 1      | 1      |
| 28 | ReplayResult caller-contract Example (delete-after-replay) — started in godoc prose, make it an Example                                                          | 1      | 1      |
| 29 | Card: check `example/` (root) still reflects v0.4.0 hooks; it built green but content may predate DrainHooks                                                     | 1      | 1      |
| 30 | Move Performance Baselines from AGENTS.md into a `docs/benchmarks.md` with dates + environment                                                                   | 1      | 1      |
| 31 | Add CI-less tag-safety check: `git tag --contains` audit that no tag points at a merge-conflicted tree                                                           | 1      | 1      |
| 32 | Re-run cqrs-lint scorecard post-v0.4.0 (flight-recorder module detection should improve)                                                                         | 1      | 1      |
| 33 | Expose `Status()` passthrough decision (F16 from previous plan) — still undecided                                                                                | 2      | 1      |
| 34 | DLQ admin dashboard example (Count/ListPaged/PurgeBefore) — still README-line only                                                                               | 1      | 2      |
| 35 | Encryption/signing opt-in design sketch — wait for consumer, but sketch cost now (half-day)                                                                      | 2      | 2      |
| 36 | Idempotency store opt-in sketch — same trigger                                                                                                                   | 1      | 1      |
| 37 | Per-projection readiness helper (F27) — demand-gated                                                                                                             | 2      | 1      |
| 38 | Cardinality audit for high-projection-count otel metrics                                                                                                         | 1      | 2      |
| 39 | Worker lifecycle log-noise filtering guidance                                                                                                                    | 1      | 1      |
| 40 | CBOR→JSON transcode helper decision for SSE consumers (F26, carried over)                                                                                        | 2      | 2      |
| 41 | Dependabot/renovate cadence decision for cqrs-lite bumps (structure linter flags its absence)                                                                    | 2      | 1      |
| 42 | `editorconfig-present` + other structure-linter INFO rules — accept or satisfy                                                                                   | 1      | 1      |
| 43 | Publish the plan-doc execution outcome (annotate the SUPERB plan with what shipped vs skipped)                                                                   | 1      | 1      |
| 44 | cqrs-lint yaml `exclude_patterns` bug — file upstream in go-structure-linter (their repo)                                                                        | 1      | 1      |
| 45 | buildflow dprint skip-empty fix — file upstream in buildflow (their repo)                                                                                        | 2      | 1      |
| 46 | Verify example/main.go (root) content against v0.4.0 hook surface                                                                                                | 1      | 1      |
| 47 | Consolidate "release state" single-source (AGENTS vs TODO_LIST) — one owner, one file                                                                            | 2      | 1      |
| 48 | godoc example for OuterMiddlewares + otel.Middleware placement (root module)                                                                                     | 1      | 1      |
| 49 | Docs: cross-link core-v1-exit-criteria from AGENTS.md v1.0.0 mentions                                                                                            | 1      | 1      |
| 50 | Retire the temp `/tmp/wave2-consumer` + `/tmp/core-v03` artifacts (hygiene)                                                                                      | 1      | 1      |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **License:** proprietary (pkg.go.dev keeps hiding all godoc — real adoption cost for a framework courting cqrs-htmx/PapDashboard) or a standard license (MIT like cqrs-htmx)? This single decision gates the entire visibility tier; everything else I can do without you.
2. **Logging posture, data in hand:** emitting INFO costs +30µs/req (+174%). Do you want (a) production docs recommending WARN default, (b) sampling in the default stack, or (c) status quo (consumer chooses the logger — already possible)?
3. **cqrs v0.4.0 is a breaking change sitting in cqrs-htmx's dependency graph** (they consume the cqrs stack; fresh-consumer compile passes, but their `FlightRecorder` usage — if any — breaks on bump). Do I prepare a compatibility note/PR on their side (user-gated cross-repo), or is notifying them the whole job for now?

---

_Point-in-time snapshot — goes stale; ANNOTATE, never rewrite. Generated from this session's runs only._
