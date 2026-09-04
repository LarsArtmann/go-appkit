# Status Report — go-cqrs-lite Deep Dive Session

**Date:** 2026-09-04 17:17 CEST · **Scope:** this session's work only (library-deep-dive on go-cqrs-lite from go-appkit/cqrs + executed fixes) and what it noticed in passing. No new research was done for this report.

**Session one-liner:** audited how go-appkit/cqrs leverages go-cqrs-lite, found master BROKEN (projectionhost v4.4.0 API break), fixed it, surfaced the new trigger capability, race-hardened the tests, updated all docs, wrote the HTML audit at `docs/research/2026-09-04_go-cqrs-lite-deep-dive.html`.

---

## a) FULLY DONE

| # | Item | Evidence |
|---|------|----------|
| 1 | **Release-blocking build break fixed** — `EventConfig.FlightRecorder` migrated from cqrs-lite's deleted `flightrecorder/v4.Recorder` to `go-flightrecorder v0.2.0`; go.mod reconciled (`go-flightrecorder` direct, `flightrecorder/v4` indirect — genuinely still required via `stack/v4`, verified with `go mod why`) | `cqrs/eventservice.go:16,47-66`, `cqrs/go.mod` |
| 2 | **New capability surfaced**: `FlightRecorderTrigger fr.TriggerFunc` (upstream v4.4.0 trigger was unreachable because derived wiring wins over `HostOptions`) | `cqrs/eventservice.go:58-66,201-220` |
| 3 | **Trigger contract test** — asserts the passthrough delivers `Kind="projection"`, `Type=<name>`, non-nil `Err` | `cqrs/flightrecorder_test.go` (`TestEventConfig_FlightRecorderTrigger_ReceivesProjectionContext`) |
| 4 | **Latent race fixed** — v4.4.0 snapshots synchronously on the worker goroutine; both FR tests now use a mutex-guarded `syncBuffer` | `cqrs/flightrecorder_test.go:20-44` |
| 5 | **Full verification wave** — `go build`, `go vet`, `go test -race -count=1` (31 tests, counted), `golangci-lint` 0 issues, `cqrs-lint 4.8.1` clean (2 suppressions still fire, none stale), `docs-mod` build green | session run log |
| 6 | **Docs updated**: README config table (FR type change, trigger row, `WithCheckpointEvery`/`WithOnFailed`/`WithMaxRestarts` examples), CHANGELOG `[Unreleased]` (breaking-change entry + trigger), AGENTS.md dep table + EventConfig option set | `cqrs/README.md:40-43`, `cqrs/CHANGELOG.md`, `AGENTS.md:225-231,241` |
| 7 | **HTML deep-dive report** published (scorecard, 5 findings, Pareto plan, version-currency table) + committed | `docs/research/2026-09-04_go-cqrs-lite-deep-dive.html` |
| 8 | **2 follow-up tasks routed** to TODO_LIST P3 (cqrs-lint binary upgrade; encryption/signing/idempotency opt-ins) | `TODO_LIST.md` P3 tail |
| 9 | **Version-currency audit** — all 7 pinned cqrs-lite/go-flightrecorder modules at latest tags (0 behind) | report appendix table |
| 10 | **Correction applied**: the report's "24 tests" claim fixed to the verified 31, with a visible correction note in the report footer | report footer, this pass |

## b) PARTIALLY DONE

1. **Deep-dive action plan**: items 1–6 of 9 done (build fix, trigger, race, docs, AGENTS/CHANGELOG, re-lint). Items 7–9 open: installed cqrs-lint upgrade (4.6.0→4.8.1; I built 4.8.1 to `/tmp` but did not touch `~/go/bin`), encryption/signing opt-ins, idempotency opt-in — deliberately routed, not built (wrapper surface stays demand-driven).
2. **Release logistics for the breaking change**: CHANGELOG entry written, but the version decision (must be a **breaking** bump — v0.4.0 in 0.x terms), tag, and wave plan are not made. Complication discovered while writing this report: **the long-pending v0.3.0 wave is now on origin** (tags `cqrs/v0.3.0`, `realtime/v0.1.0`, `flightrecorder/v0.1.0` confirmed via `git ls-remote`; master at `ebe34ea` on origin) — **not pushed by this session**. The breaking change is therefore already public on master, untagged. AGENTS.md "PUSH PENDING" section and TODO_LIST P1 are now stale.
3. **Upstream changelog review**: event v4.8.0/v4.9.0 and projectionhost entries read; other modules' recent changes (system v4.6.0, middleware v4.5.x, transport) NOT reviewed for wrapper relevance.
4. **Scorecard interpretation**: honest result obtained (3/27 direct — by design for a wrapper) and the 24 "missing" modules routed into three groups, but I did not validate the tool's detection logic itself (e.g., it misses Bundle-reachable capability by construction).

## c) NOT STARTED

- The **next cqrs release wave** (version decision, tag, push-gate process for the breaking change).
- **EventConfig opt-ins** for encryption/v4, signing/v4, idempotency/sqlstore, scheduling (routed to TODO_LIST).
- **Installed cqrs-lint 4.8.1 upgrade** (`~/go/bin/cqrs-lint` still 4.6.0).
- **Cross-module verification wave** — I verified cqrs + docs-mod only; a parallel session's health-module work (noticed in TODO_LIST/git log, not mine) makes a full 9-module build+test sweep due.
- **AGENTS.md Release State refresh** (push happened; breaking change untagged) — held off because a parallel session is editing the same doc.
- Negative-path trigger test (trigger returns `false` → no capture).

## d) TOTALLY FUCKED UP

1. **I published a fabricated number.** The HTML report said "24 tests" — I never ran a counting command; the real count is 31. Caught and fixed during THIS status pass with a visible correction note, but it violates the verify-your-own-claims rule: a reader trusted it for ~2.5 hours. Root cause: asserting metrics from memory instead of from tool output.
2. **Master was broken at session start and I didn't know for two phases.** Commit `53ee2ea` ("bump Go toolchain… upgrade deps") upgraded go-cqrs-lite without adopting the `WithFlightRecorder` API break, so `cqrs` did not compile on a repo whose release wave was "prepared, push pending". I only found it because the LSP flagged it mid-audit; my first `go build` came after two read-only phases. The memory lesson ("build first, observe the cascade") applies to session START, not just after deletions.
3. **Auto-commit daemon shredded the history of a breaking API migration.** My migration landed as `bd5e872 "chore: auto-commit 10 changed file(s) (heuristic)"` mixed with 7 unrelated LICENSE files; my doc edits landed in `b7f45d5` mixed with a *different session's* go-health audit doc. Nobody can review the breaking change from history, and the commit that carries it doesn't say so.
4. **Concurrent-session blind spot.** A parallel session was committing to the same repo while I worked (health module, AGENTS.md/TODO_LIST edits — e.g. `b7f45d5`, `21436f4`). I worked from a stale 8-module mental model (AGENTS.md header) when the repo is at 9 modules, and never diffed their shared-doc edits against mine. No clobber detected (builds green), but that was luck, not process.

## e) WHAT WE SHOULD IMPROVE

1. **`go build ./...` is step 0 of every session**, before any reading — generalize the delete-cascade lesson to session start.
2. **No published number without an in-session command that produced it** (counts, timings, percentages).
3. **Verify daemon commits immediately**: after each change, check `git log --stat -1` — if my breaking change is about to be buried under "10 changed files (heuristic)", commit it myself first with a real message.
4. **Detect parallel sessions early** (recent `git log` timestamps, TODO_LIST "Updated" header) before editing shared docs (AGENTS.md, TODO_LIST.md).
5. **Test the negative path** for passthrough options (trigger=false), not just the happy gate.
6. **Prefer surfacing upstream capabilities over wrapper API**: the `FlightRecorderTrigger` field (pass-through, 3 lines) was the right shape; keep that pattern for future upstream options.
7. **Release-state facts need a single owner** — push status now lives (stale) in three places: AGENTS.md, TODO_LIST P1, and status reports.
8. cqrs-lint CLI drift (arg form `./...` → `.`) broke silently between 4.6.0 and 4.8.1 — pin the tool version next to the config profile in docs.

## f) TOP THINGS TO GET DONE NEXT (up to 50 — brainstorm fuel; items 1–12 are real work, 13–50 are ROADMAP-grade)

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 1 | Decide next cqrs version (breaking → v0.4.0), cut + tag | 5 | 2 |
| 2 | Refresh AGENTS.md Release State (wave pushed 2026-09-04; breaking change untagged) | 4 | 1 |
| 3 | Annotate TODO_LIST P1 (push gate resolved; tags on origin) — docs-health ANNOTATE, don't rewrite | 4 | 1 |
| 4 | Post-push verification per old P1 checklist: fresh-consumer `go get` + pkg.go.dev for cqrs v0.3.0 | 4 | 2 |
| 5 | Full 9-module build+test+lint sweep (covers parallel health session's work too) | 4 | 2 |
| 6 | Install cqrs-lint 4.8.1 to `~/go/bin` | 2 | 1 |
| 7 | Negative-path test: trigger returns false → no snapshot | 3 | 1 |
| 8 | Document-by-test: HostOptions cannot override derived wiring (Logger/Metrics/FR precedence) | 3 | 1 |
| 9 | AGENTS.md: refresh "Six of eight modules require jsonv2" → current 9-module truth | 2 | 1 |
| 10 | Add `cqrs-lint .` (not `./...`) to README gotcha list | 1 | 1 |
| 11 | Review `.cqrs-lint.json` pinned feature profile against 4.8.1 schema | 2 | 1 |
| 12 | Update docs-health trail: annotate the 2026-08-16 comparison doc where the flight-recorder story changed | 2 | 2 |
| 13 | cqrs/example: shared `fr.Recorder` across appkit middleware + projections | 3 | 2 |
| 14 | Bench `WithCheckpointEvery` 1 vs 100 into README | 2 | 2 |
| 15 | Document `ErrWorkerFailed` sentinel for dashboards/alerts | 2 | 1 |
| 16 | Decide: expose `Status()` passthrough on EventService or keep `Host()` only (one-line ADR in AGENTS) | 2 | 1 |
| 17 | `ForceStop` policy note (when graceful drain is not enough) | 1 | 1 |
| 18 | Encryption/v4 EventConfig opt-in design sketch (when a consumer appears) | 3 | 3 |
| 19 | Signing/v4 opt-in (same trigger) | 3 | 3 |
| 20 | Idempotency/sqlstore opt-in (same trigger) | 2 | 2 |
| 21 | Scheduling `Timer.Actor` audit-trail doc in cookbook | 1 | 1 |
| 22 | Prometheus bridge pairing snippet (cqrs.NewOTelProjectionMetrics + prometheus.Setup) | 2 | 1 |
| 23 | DLQ admin dashboard example (Count/ListPaged/PurgeBefore assertions) | 2 | 2 |
| 24 | ReplayResult contract example: caller must Delete/Purge replayed entries | 2 | 1 |
| 25 | LagPerProjection → alert-threshold example in README | 1 | 1 |
| 26 | CBOR→JSON transcoding helper decision for SSE raw-payload consumers | 3 | 3 |
| 27 | Per-projection readiness helper (`CheckProjectionReadiness(name)`) — demand-gated | 2 | 2 |
| 28 | Staleness budget defaults guidance (what number to pick) | 1 | 1 |
| 29 | Metrics cardinality note for high-projection-count apps | 1 | 1 |
| 30 | Worker lifecycle log-noise filtering guidance (log levels) | 1 | 1 |
| 31 | SQLite WAL checkpoint tuning + busy_timeout doc under contention | 2 | 2 |
| 32 | Multi-recorder coordination ADR (HTTP + projections, one process) | 2 | 2 |
| 33 | SnapshotStore/ReadModels accessor examples (Bundle-reachable, undocumented) | 2 | 2 |
| 34 | BackwardsSource/journal replay tooling example | 1 | 2 |
| 35 | DLQ dead-letter age alerting recipe | 1 | 1 |
| 36 | `.golangci.yml` refresh: exhaustruct → exhaustruct_v5 migration (deprecation warning seen in-session) | 2 | 2 |
| 37 | go.mod toolchain-line consistency check across all 9 modules | 2 | 1 |
| 38 | Verify flightrecorderhealth against go-flightrecorder v0.2.0 after the wave push | 2 | 1 |
| 39 | Upstream (user-gated, verify-before-filing): document synchronous-Snapshot semantics for test pollers | 1 | 2 |
| 40 | logger_test chanMutex → stdlib consideration (simplification review) | 1 | 1 |
| 41 | Decision: single source of truth for release state (AGENTS vs TODO_LIST) | 2 | 1 |
| 42 | CI-less cross-module verification script (no flake.nix here; plain go.work sweep) | 2 | 2 |
| 43 | scenario/v4 drift guard: re-verify README cookbook after next cqrs-lite release | 1 | 1 |
| 44 | ERROR-pages + cqrs composition example (family-aware failures) | 1 | 2 |
| 45 | Query-journal example (Bundle-reachable, no docs) | 1 | 2 |
| 46 | Dependabot-style dep sweep cadence decision (cqrs-lite moves fast) | 2 | 1 |
| 47 | `DeadLetterStoreAdmin` interface assertion example for wrappers | 1 | 1 |
| 48 | Document FlightRecorderTrigger log-noise tradeoff (sync on worker goroutine) | 1 | 1 |
| 49 | Add module-level example_test.go for EventConfig (godoc-checked quick start) | 2 | 2 |
| 50 | Re-run go-cqrs-lite scorecard after each wrapper feature to track adoption trend | 1 | 1 |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Who pushed master + the v0.3.0 wave to origin mid-session (tags confirmed via `git ls-remote` at ~17:05)?** It wasn't me. If the push was intentional and validated (fresh-consumer `go get`, pkg.go.dev), I'll update AGENTS/TODO accordingly; if it was the daemon overstepping its mandate, that's a config problem worth fixing before the next release wave.
2. **Should the breaking cqrs change ship as its own v0.4.0 now, or wait to ride the next combined wave (health + otel)?** Bundling reduces release overhead; shipping now shrinks the window where master is ahead of the newest tag with a public breaking API.
3. **Is the parallel go-health session still active?** It edited AGENTS.md/TODO_LIST while I did; if it's still running, one of us should own the stale Release-State/P1 sections to avoid dueling edits — tell me who and I'll defer (or proceed).

---

_Why 17:17 in the filename: `date` at report start. Point-in-time snapshot — goes stale; ANNOTATE, never rewrite._
