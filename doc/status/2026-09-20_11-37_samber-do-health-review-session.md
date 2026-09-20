# Status Report — samber/do × Health Deep-Review Session

**Timestamp:** 2026-09-20 11:37 CEST | **Session window:** ~10:58–11:37
**Scope:** single-task session — deep review "Are we using samber/do v2 + Health Checks superbly? Is the architecture service-oriented, composable, resilient, self-health?" (skills: samber-do-best-practices, architecture-review, library-deep-dive, status-report)
**Deliverables:** `docs/architecture-understanding/2026-09-20_10-58_samber-do-health-architecture.md` (179 lines, committed in `3140190`) · AGENTS.md:165 dep-truth fix (same commit) · TODO_LIST.md harvest +8/−2 (uncommitted; daemon pending)

---

## Direct answers first (asked in the prompt)

**What did I forget?**

1. **The session's central subject was never executed end-to-end.** I proved every piece by source (do-v2 fan-out, go-health recorder path, Mounted lifecycle) and by suite, but the combined flow — injector + `Checkable` + `Trigger` + `health.New` + `Mounted` + appkit `Service` — has never RUN anywhere, including by me. A utilization review of a composition should have built the composition.
2. **Lint was skipped.** I ran both suites with `-race -count=1` but never `golangci-lint run` on `health/` + `flightrecorderhealth/` — the "0 issues" standing claim was not re-verified this session.
3. **`scripts/check-pin-drift.sh` was never run** after I edited AGENTS.md:165 (a version-bearing line). Low risk (release-state pins untouched), but the ritual exists precisely for this and I didn't close the loop.
4. **No index cross-link:** the new review was not registered in `doc/status/README.md`'s world, and AGENTS (376/377 line cap) got no pointer to the report's F-findings — traceability depends on TODO_LIST alone.
5. **Two near-wrong findings** caught only because I kept reading source: (a) I first read `Register`'s eager `do.InvokeNamed` as removable sloppiness — do-v2 `service_lazy.go:127-152` proves it is load-bearing (unbuilt lazy services report healthy); (b) I first flagged the example's manual `GET /health/ready` as redundant to `cfg.ReadyCheck` — httputil's GET-qualified patterns + the dashboard's method-agnostic `/health` conflict make it forced. Both would have been plausible, wrong recommendations.

**What could I have done better?**

- **Toolchain sanity first:** the workspace LSP was broken the whole session (root `go.mod` go 1.27.1 vs go.work 1.26.7). I worked around it per-module with `GOWORK=off` instead of escalating in minute one — one `go list ./...` at session start would have surfaced F6 immediately.
- **Compose, then judge:** writing the composed-stack example during the session would have upgraded F2 from "documented gap" to "fixed gap".
- **Verify live claims:** the drain-lockstep 503 E2E rests on a prior session's manual verification; re-running `health/example` cost ~2 minutes and I skipped it.

**What could I still improve (process)?**

- Session-start ritual: `go env` + one-module `go list` sanity gate before any analysis.
- For "are we using X superbly" reviews: run the flagship composition before scoring it.
- Re-run, don't inherit: benchmarks (frh ~4.7µs) and live E2Es cited from AGENTS should be re-executed when cheap.

---

## a) FULLY DONE (verifiable, committed or green)

| #   | Item                                                                                                                                                                                | Evidence                                                                                              |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| A1  | Deep-dive report written: 3-layer health architecture map, utilization gap table, DO-1…6 audit, rubric 4.57/5, findings F1–F8, roadmap                                              | `docs/architecture-understanding/2026-09-20_10-58_samber-do-health-architecture.md`, commit `3140190` |
| A2  | samber/do usage survey: exactly ONE production file in the repo (`flightrecorderhealth/adapter.go`); all call sites enumerated                                                      | `rg 'samber/do' --type go` across repo; report §1                                                     |
| A3  | DO-1…DO-6 compliance: all clean (no `MustInvoke` in runtime paths, no global injector, no `Override`, no loops-Invoke, no cross-service shutdown)                                   | report §3, each rule source-cited                                                                     |
| A4  | do-v2 trap source-verified: unbuilt lazy services report HEALTHY (`return nil` when `!built`) → frh's eager invoke is load-bearing                                                  | `samber/do/v2@v2.1.0/service_lazy.go:127-152`                                                         |
| A5  | go-health cliff source-verified across ALL published versions (v0.1.3/v0.2.0/v0.3.0): `NewWithHealthCheck` silently nils `WithHealthRecorder`                                       | `accessors.go:61/36/61` resp.; report §1 table                                                        |
| A6  | Version currency via proxy: samber/do v2.1.0 = latest (pinned ✓); go-health latest = v0.3.0 (health pins v0.2.0, frh pins v0.1.3)                                                   | `go list -m -versions` from satellite dir                                                             |
| A7  | Both health modules re-verified green: `go test ./... -race -count=1` ok                                                                                                            | session run output, frh 2.296s / health 1.045s                                                        |
| A8  | AGENTS.md health-deps line corrected to go.mod truth (v0.2.0 / v0.9.0, marked UNRELEASED); line cap intact at 376/377                                                               | AGENTS.md:165, `wc -l` = 376, commit `3140190`                                                        |
| A9  | Findings F1–F6 harvested into TODO_LIST.md (4 edits: header, 4×P2, 2×P3, open-question #4)                                                                                          | TODO_LIST.md +8/−2                                                                                    |
| A10 | Integration gap mechanically proven: `integration/go.mod` has NO health/frh deps (pins core v0.5.1, errorpages, otel, realtime only)                                                | `rg go-appkit integration/go.mod`                                                                     |
| A11 | httputil route-pattern verification: default health endpoints are GET-qualified (`GET /health`, `/health/live`, `/health/ready`) — explains the dashboard conflict gotcha precisely | `httputil@v1.2.0/health.go:88-94`                                                                     |

## b) PARTIALLY DONE (works now / what's missing / blocker / effort)

| #  | Item                       | Works                                                                                                                       | Missing                                                                                                             | Blocker                                                | Effort |
| -- | -------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ | ------ |
| B1 | F1 silent-recorder cliff   | Root-caused to source level; documented in report + TODO_LIST                                                               | appkit-side fix (NewProbe godoc warning + injector-path example) NOT coded; upstream sentinel-error ask not drafted | none for appkit side; upstream filing is USER-gated    | S/M    |
| B2 | F2 composed-stack proof    | Gap identified and justified                                                                                                | integration test not written; combined flow never executed (also see "forgot" #1)                                   | needs health/frh releases first for published-tag pins | L      |
| B3 | F3 version alignment       | Drift mapped three ways (frh v0.1.3 / health v0.2.0 / upstream v0.3.0)                                                      | no bump executed, no release cut, AGENTS release state untouched                                                    | release sequencing is a USER call (Q2)                 | M      |
| B4 | F6 root go.mod 1.27.1      | Fully diagnosed (auto-commit `d5c6693`, no dep requires it, satellites on 1.26.7, CI `go-version-file` now mixed-toolchain) | NOT fixed — deliberate: reverting a directive I didn't author could fight an intentional floor raise                | USER decision (Q1)                                     | S      |
| B5 | F4 lifecycle adapter       | Design sketched (Mounted→`do.ShutdownerWithError`, pusher-stop semantics preserved)                                         | not implemented                                                                                                     | overlaps W4 "do" battery — ship-slice-or-fold decision | M      |
| B6 | F5 lazy-healthy gotcha doc | Trap verified; doc targets named (frh doc.go, README)                                                                       | paragraphs not written                                                                                              | none                                                   | S      |

## c) NOT STARTED (noticed this session, untouched)

| #  | Item                                                                                                                                                                                                                      | Why not started                                | Priority              |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | --------------------- |
| C1 | `Mounted.Start` asymmetric rollback: `dashboard.Start` failure leaves `started=true` (mount.go:174-179) → retry rejected `health.already_started`; probe path DOES reset (mount.go:164-171). Verify with a test, then fix | noticed during report writing, after test runs | High — small real bug |
| C2 | `failingServiceNames` returns map-iteration order (nondeterministic trigger logs); `firstError` picks an arbitrary failing service's error                                                                                | cosmetic, needs a decision (sort vs doc)       | Low                   |
| C3 | go-health v0.3.0 evaluation (new `aggregate/`, `federation/` packages — unexamined)                                                                                                                                       | out of session scope per user instruction      | Medium                |
| C4 | `Register` polish: `do.ProvideNamedValue` one-liner + document duplicate-name panic                                                                                                                                       | P3 polish, listed only                         | Low                   |
| C5 | samber-do-auditlog hooks example                                                                                                                                                                                          | demand-gated per repo policy                   | Low                   |
| C6 | golangci-lint on health + frh (see "forgot" #2)                                                                                                                                                                           | forgotten, not deliberately skipped            | Medium                |
| C7 | frh benchmark re-run (~4.7µs claim) and health-example live E2E re-run (see "could've done better")                                                                                                                       | time-boxed out                                 | Medium                |

## d) TOTALLY FUCKED UP

Nothing destructive this session — no code deleted, no reverts, no broken builds introduced. Radical-honesty items that ARE fucked up (found, not caused, by this session):

1. **Root `go.mod` demands go 1.27.1 (F6) — the repo cannot build at root in this environment.** Severity: blocks all local root-module work + every workspace LSP/golangci diagnostic (17 errors, all this one cause). Root cause: auto-commit `d5c6693` (2026-09-18) raised the directive with NO dependency requiring it; all 10 satellites remain 1.26.7; `go.work` still says 1.26.7; CI's `go-version-file: go.mod` now provisions a different toolchain for the root job than the satellite jobs. Mitigation: per-module `GOWORK=off` runs (used all session). Fix: one-line revert OR deliberate floor-bump propagation — USER-gated.
2. **The flagship feature can vanish silently (F1).** `appkithealth.NewProbe(checks, health.WithHealthRecorder(frhealth.NewTrigger(...)))` compiles, runs, logs green — and captures zero traces, because go-health nils the recorder on the injector-free path. Severity: silent capability loss for exactly the consumers the convenience API targets. Mitigation today: AGENTS gotcha (not shipped to consumers).
3. **The self-health stack has zero CI proof (F2).** The composition this whole architecture exists for is assembled nowhere under test; the only cross-module health proof is a manually-run example. Severity: regressions in the lockstep/trigger wiring would ship unnoticed.

Session-caused fucked-up count: **0**. Near-misses (caught pre-publication): 2 — see "What did I forget?" #5.

## e) WHAT WE SHOULD IMPROVE

1. **Toolchain gate at session start** — 30s check (`go env`, one `go list`) would have caught F6 immediately instead of 17 noisy diagnostics all session.
2. **"Run the composition" rule for library-utilization reviews** — source-reading proves pieces; only execution proves the composition (F2 exists because I followed the first, not the second).
3. **Shipped-vs-internal knowledge split** — the worst traps (F1) live only in AGENTS.md; consumer-facing surfaces (godoc, README, examples) must carry them, since AGENTS never reaches a consumer.
4. **Verify-before-recommending held, but barely** — two wrong conclusions were one edit away from publication. Keep the "read the called-into source, not just our source" discipline as a hard rule.
5. **Version-alignment sweeps within a repo family** — frh v0.1.3 vs health v0.2.0 vs upstream v0.3.0 happened because dep bumps are per-module and nothing asserts family consistency; a tiny grep-able convention (or the pin-drift script extended to family deps) would catch it.

## f) TOP 50 NEXT TASKS (impact-ranked clusters; ☑ = already harvested into TODO_LIST.md this session — the rest are brainstorm fuel for docs-health HARVEST, not auto-commitments)

**Cluster A — F1 silent cliff (P0)**

| # | Task                                                                    | Impact | Effort | Category              |
| - | ----------------------------------------------------------------------- | ------ | ------ | --------------------- |
| 1 | ☑ NewProbe godoc: warn `WithHealthRecorder` is silently dropped         | High   | S      | Docs                  |
| 2 | ☑ Injector-path Trigger example (health example or godoc example)       | High   | M      | Docs                  |
| 3 | ☑ Draft upstream go-health ask: reject the option with sentinel error   | High   | S      | Upstream (USER-gated) |
| 4 | Defensive `NewProbe` variant erroring on recorder options               | Medium | S      | Feature               |
| 5 | AGENTS gotcha line: in-place pointer to report §F1 (no new lines — cap) | Low    | S      | Docs                  |

**Cluster B — F2 composed proof (P0)**
| 6 | ☑ Add health+frh published pins to integration/go.mod | High | S | Test |
| 7 | ☑ `TestHealthStackThroughAppkitService`: drain-lockstep 503 assertions | High | L | Test |
| 8 | ☑ Trigger-capture assertion in the same test | High | M | Test |
| 9 | ☑ Recorder-row visibility via dashboard cached response | Medium | M | Test |
| 10 | Run the combined stack manually once (pre-CI proof) | High | S | Quality |
| 11 | Composed quick-start in health README/doc.go | Medium | M | Docs |

**Cluster C — F3 releases (P1)**
| 12 | ☑ frh: bump go-health v0.1.3 → v0.2.0 | High | S | Release |
| 13 | Re-run frh suite + contract pins post-bump | High | S | Quality |
| 14 | ☑ Release health module (dep bumps) per Release Ritual | High | M | Release |
| 15 | ☑ Release frh after bump | Medium | M | Release |
| 16 | Fresh-consumer proxy checks for both tags | High | S | Release |
| 17 | Evaluate go-health v0.3.0 (aggregate/federation) | Medium | M | Quality |
| 18 | ☑ AGENTS Release State update in the same train (Ritual step 5) | High | S | Docs |
| 19 | Run `scripts/check-pin-drift.sh` in the same train (and once now) | Medium | S | Quality |

**Cluster D — F6 toolchain (P0, USER-gated)**
| 20 | ☑ USER decision: revert vs deliberate floor bump | Critical | S | Decision |
| 21 | If revert: `go mod edit -go=1.26.7` + tidy + root build green | Critical | S | Bug |
| 22 | If bump: propagate go.work/AGENTS/CI + nixpkgs reality check | High | M | Bug |
| 23 | Verify LSP/gopls diagnostics green post-fix | Medium | S | Quality |
| 24 | CI assert: identical go-directive across all 11 go.mod files | High | S | Quality |

**Cluster E — F4/F5/F7/F8 (P2/P3)**
| 25 | ☑ frh doc.go + README: lazy-healthy gotcha | Medium | S | Docs |
| 26 | Compile-check edited README snippets in scratch module (ritual) | Medium | S | Docs |
| 27 | ☑ Design Mounted shutdown adapter (keep health injector-free) | Medium | S | Design |
| 28 | ☑ Implement + test the adapter | Medium | M | Feature |
| 29 | ☑ `Register`: `do.ProvideNamedValue` refactor | Low | S | Cleanup |
| 30 | ☑ Document `Register` duplicate-name panic | Low | S | Docs |
| 31 | samber-do-auditlog wiring example | Low | M | Feature |

**Cluster F — session-noticed, unfixed**
| 32 | `Mounted.Start`: reset `started` when `dashboard.Start` fails (C1) — test first | High | S/M | Bug |
| 33 | Sort `failingServiceNames` for deterministic logs | Low | S | Quality |
| 34 | Document `firstError` nondeterminism contract | Low | S | Docs |
| 35 | golangci-lint on health + frh (session gap) | Medium | S | Quality |
| 36 | Re-run frh benchmark (~4.7µs baseline) | Low | S | Quality |
| 37 | Re-run health example live E2E (drain-lockstep 503) | Medium | M | Quality |
| 38 | No-dashboard example path teaching `cfg.ReadyCheck = mounted.Ready` | Medium | S | Docs |
| 39 | security+health composed example (`DashboardHardenedPreset` + nonce) | Medium | M | Docs |
| 40 | Verify `Drain()` on never-started probe is harmless; doc or guard | Low | S | Quality |
| 41 | Index the new review in `doc/status/README.md`'s world | Low | S | Docs |
| 42 | Fold the Logging-uncorrelated-line limitation into the drafted upstream-asks batch | Low | M | Upstream (USER-gated) |

**Cluster G — process/meta**
| 43 | Add session-start toolchain sanity ritual to AGENTS Gotchas (in-place) | Medium | S | Process |
| 44 | Run `check-pin-drift.sh` now to validate today's AGENTS edit | Medium | S | Quality |
| 45 | Decide indexing convention for architecture-understanding reports | Low | S | Process |
| 46 | Upstream ask (USER-gated): do-v2 doc note on lazy-healthy semantics | Low | S | Upstream |
| 47 | AGENTS Integration table: add health/frh rows once pins land (F2) | Medium | S | Docs |
| 48 | Verify CI matrix + dependabot parity slots for health/frh | Medium | S | Quality |
| 49 | Reconcile AGENTS "NewProbe bypasses HealthRecorder" wording with report F1 (no split-brain) | Low | S | Docs |
| 50 | After A-C land: re-run this review's adoption score (82/100) to measure the delta | Medium | S | Process |

**HARVEST status:** items 1-3, 6-9, 12, 14-15, 18, 20, 25, 27-28, 29-30 already routed to TODO_LIST.md this session. Items 32 (C1) and 24/44 are the strongest uncaptured candidates if you want a second HARVEST pass.

## g) QUESTIONS I CANNOT ANSWER MYSELF (3)

**Q1 — Root `go.mod` go 1.27.1 (auto-commit `d5c6693`, 2026-09-18): deliberate floor raise or accident?**
What I tried: diffed the commit (only the go line changed in root go.mod; cqrs got dep bumps; the drift script got formatting); checked every family go.mod for a requiring dependency (none — do v2.1.0 needs go1.18, go-health v0.2.0 needs go1.26, dashboard v0.9.0 needs 1.26.7); confirmed satellites + go.work + AGENTS all say 1.26.7. Intent is not inferable from the repo. The answer decides a one-line revert vs a floor-bump propagation (go.work + AGENTS + CI + nixpkgs reality).

**Q2 — Release sequencing for the health family: one train now, or batch later?**
Working tree carries unreleased bumps (health: go-health v0.2.0 + dashboard v0.9.0) and the fixes need them published (F2 pins published tags; F3's whole point is what consumers resolve). Options: (a) Release Ritual now for health + frh + the F1 godoc fix in one train; (b) batch with go-health v0.3.0 evaluation; (c) hold. Policy call — the Ritual constrains mechanics, not timing.

**Q3 — File the upstream go-health ask now ("`NewWithHealthCheck` should reject `WithHealthRecorder` with a sentinel error")?**
What I tried: verified the behavior against v0.1.3/v0.2.0/v0.3.0 sources (all identical, prose-documented only) — the verify-before-filing gate's source-level evidence is in hand. Filing to an external repo is USER-gated in this workspace, so: file now (github-voice flow), or keep it report/TODO-only?

---

**HARVEST note (skill contract):** section (f) feeds docs-health HARVEST. Core findings (F1-F6) were already harvested into TODO_LIST.md during the session; the remaining ~20 uncaptured items above are brainstorm-grade and should be routed with HARVEST rigor (most belong in ROADMAP or stay report-only), not bulk-copied.

_Report ends. Waiting for instructions._

---

## ADDENDUM — 2026-09-20 (execution-train evidence)

**Composed-stack manual proof (plan T03 / M11-M13)** — `TestComposedHealthStack` in a scratch module, against PUBLISHED tags only (`go-appkit/health v0.1.1`, `go-appkit/flightrecorderhealth v0.1.2`, `go-flightrecorder v0.2.0`, `go-health v0.1.3`, `samber/do v2.1.0`; jsonv2, GOPROXY=off):

- Wiring: `fr.New(WithWriter(&buf), WithMinAge(1ms), WithMaxBytes(1MiB))` → `do.New()` + database service + `frhealth.Register(injector, rec, "flight-recorder")` → `health.New(injector, WithHealthRecorder(frhealth.NewTrigger(rec)), WithCriticalServices("database"), WithRefreshInterval(50ms))` → `appkithealth.New(probe)` → `RegisterRoutes` + `httptest`.
- Asserted green: `/healthz` 200 + `/readyz` 200 pre-drain; `flight-recorder` row present in the `/readyz` JSON body; failing `database` surfaces as `/readyz` 503 through the refresh loop (trigger fires on that failing batch); after `mounted.Drain()` — and even after healing the dependency — `/readyz` stays 503 while `/healthz` stays 200 (drain lockstep); recorder buffer non-empty (34,534-byte runtime trace captured). Stable at `-count=3`.
- Gotcha discovered while writing the proof: `Probe.Evaluate` runs a batch but does NOT publish to the background cache — only the refresh loop (`refreshCache`) does; `/readyz` serves the cache. Proof originally failed on exactly this. Worth a godoc line in go-health's `Evaluate` (folded into the T17 upstream draft pack).
- This proof is the manual run that plan T11 turns into a permanent pinned test (`TestHealthStackThroughAppkitService`) after the release train publishes the new tags.

## ADDENDUM 2 — 2026-09-20 (re-verification evidence, plan T19)

- **frh Trigger benchmark** (`BenchmarkTrigger_RecordHealthCheckWithContext_AllPass`, 2 services, no-capture hot path): **2580 ns/op, 1860 B/op, 36 allocs/op** — BETTER than the documented ~4.7µs baseline, favorable drift, no regression. README number left as-is (µs-scale claim remains true; per the otel gotcha, single-run deltas under ±25-45% are machine noise).
- **health/example live E2E** (free port 44519, Go behavioral probe — not curl): pre-drain `/readyz` 200 + `/health/ready` 200 + `/healthz` 200 + dashboard 200; after SIGTERM, during the 5s drain window: `/readyz` 503 + `/health/ready` 503 + `/healthz` 200 (lockstep confirmed live); shutdown phase logs complete with `graceful shutdown complete result=ok`.
- Note: example's doc comment still says port 8080/8081 are typical dev ports — both were occupied here; the probe used a dynamically allocated free port (lesson: examples should default to `127.0.0.1:0`-style guidance or document PORT loudly).
