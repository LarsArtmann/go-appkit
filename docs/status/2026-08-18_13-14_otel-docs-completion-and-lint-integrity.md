# Status Report: OTEL Docs Completion & Lint-Integrity Correction

**Date:** 2026-08-18 13:14 | **Session scope:** continuation of the 2026-08-18 OTEL work — finish the interrupted documentation, run the final verification sweep, fix what it revealed | **Predecessor:** `2026-08-18_12-45_otel-module-and-telemetry-hooks.md` (authoritative record of the OTEL implementation itself)

---

## Executive Summary

This session closed every loose end the predecessor left: AGENTS.md re-applied (plus stale-fact fixes the predecessor never planned), FEATURES.md and TODO_LIST.md updated, and a full 8-module verification sweep run. The sweep **disproved the predecessor's "core lint 0 issues" claim** — 7 real findings surfaced (masked by a corrupted shared lint cache from parallel runs), all fixed across three rounds, ending with every module genuinely green. Nothing committed; the three user-gate questions still stand.

---

## a) FULLY DONE

### 1. Session-state reconstruction
- Read the predecessor status report, git status (33 changed/added files), go.work (discovered it is **gitignored** — on disk, not in HEAD, absent from `git status`; informational only), otel API surface, and the stale doc files before touching anything.

### 2. AGENTS.md — interrupted edit re-applied and extended (19 edits + 1 follow-up)
- Module list "Six" → "Eight"; otel bullet inserted (alias `appkitotel`, unreleased, tag-with-next-wave note).
- GOEXPERIMENT notes: "Six of seven" → "Six of eight"; otel joins flightrecorderhealth in the NOT-required list.
- Build commands: otel block added (no GOEXPERIMENT, GOWORK=off).
- Lint standard: date → 2026-08-18, "All 7" → "All 8 modules", otel ireturn allow documented.
- Release State: pointer to the uncommitted OTEL work + the §g gate.
- Core code-org table: `service.go` (ShutdownHooks sequence), `config.go` (new fields + sentinel registry), `middleware.go` (outer→stack→extra, fresh backing array).
- New "otel Module — Code Organization" table (7 rows) + dependencies table + 7-entry gotchas section (one-Setup-per-process, ForceFlush race, span-name vs route-attr asymmetry, InMemoryExporter reset, health filter, Logging-correlation limitation, GOWORK=off).
- **Beyond the plan — stale-fact fixes:** two dependency tables still said httputil `v0.11.0` (actual: `v0.12.0`); the Testing section still advised `DrainDelay: 0` for fast tests, which is **wrong** (0 applies the 5s default) — corrected to `NoDrainDelay` with the ~30s→6s evidence. cqrs `EventConfig.Metrics` row now mentions `NewOTelProjectionMetrics`.
- One of 19 multiedit edits failed (overlapping targets); caught by outcome-grep, patched with a targeted edit.

### 3. FEATURES.md
- otel section: 12 rows (setup, flush-safe shutdown, middleware bridge, propagation, filters, public-endpoint mode, cardinality-safe metrics, views, log correlation, no-op mode, example) + the known Logging-correlation limitation as prose.
- Core: `OuterMiddlewares`, `ShutdownHooks`, `NoDrainDelay` rows. cqrs: OTel projection metrics adapter row.

### 4. TODO_LIST.md
- Header: 8 modules, otel-unreleased note, "Six of eight" GOEXPERIMENT phrasing.
- P1: two new gated items — commit-the-OTEL-work (§g Q1) and tag-`otel/v0.1.0` (incl. replace-directive drop + hermetic verify + post-tag checks; wave membership = §g Q2).
- Removed the now-obsolete P2 "Document the `DrainDelay: 0` test-ergonomics pattern" (superseded by NoDrainDelay, now documented).
- Harvested near-term backlog: upstream cqrs-lite ForceFlush item (Q3), satellite `DrainDelay: 0` sweep, otel benchmark + benchstat, httputil ctx-Logging proposal, v1.0.0-criteria fold-in.

### 5. Verification sweep — and the lint-integrity correction
- **All 8 modules test+vet(+build) green with `-race -count=1`** (core needs GOEXPERIMENT=jsonv2; satellites hermetic GOWORK=off).
- First lint pass (run as 3 parallel batches) reported 4 core findings — contradicting the predecessor's "0 issues". Fixed those (appendAssign, noctx, noinlineerr, sloglint). The re-run then surfaced **3 more pre-existing findings** (canonicalheader ×2, gosec G602) — proving the parallel run had under-reported: concurrent golangci-lint processes raced on the shared cache at `/mnt/buildcache` (visible "Failed to persist facts" warnings).
- Fixed the remainder; one intermediate fix (make+assign for the aliasing probe) itself tripped makezero — final shape is a composite literal, clean by construction under all three slice-policing linters.
- **Re-linted all 7 satellites strictly sequentially: 0 issues each. Core: 0 issues, `-race` green.** The predecessor report got a correction addendum.
- Persisted the lesson into AGENTS.md linting section: one module's lint at a time; re-run sequentially when results look wrong.

---

## b) PARTIALLY DONE

- **Nothing half-landed.** Every edit this session is applied and verified. The OTEL work as a whole remains **uncommitted** (user gate), which is "prepared", not "partially done".
- Root README otel quick-start snippet (predecessor §f P3-27): still absent — module table row exists, Configuration section does not mention otel. Out of this session's scoped steps; carried in f).

## c) NOT STARTED

- Commit / tag / push of anything (gated on §g Q1/Q2 + the standing push gate).
- cqrs-lite upstream ForceFlush issue/PR (§g Q3).
- otel release prep: drop `replace ../`, require published core, fresh-consumer proxy test, pkg.go.dev check.
- Satellite `DrainDelay: 0` test-suite sweep (realtime/errorpages/docs/flightrecorder).
- otel middleware benchmark; httputil ctx-aware Logging proposal.

## d) TOTALLY FUCKED UP (this session's mistakes, all fixed)

1. **I trusted a corrupted lint run.** Running three lint batches in parallel for speed raced the shared `/mnt/buildcache` — the "4 findings" first report was incomplete (3 missing), and the satellites' simultaneous "0 issues" greens were untrustworthy until I redid everything sequentially. Cost: one full extra lint sweep. Root cause: my parallelization choice; the cache location is not concurrency-safe.
2. **Linter whack-a-mole on one test helper.** appendAssign fix (`outer[:2]`) → gosec G602; second fix (`make` + index assigns) → makezero; third (composite literal) finally clean. Three rounds because I fixed the *reported* finding each time instead of reasoning once about the shape satisfying all slice-policing linters.
3. **Two multiedit partial failures.** middleware_test.go (6/7): I wrote `old_string` from memory (`resp.Header…` context) instead of copying the file's exact `rec.Header().Get(...)` — the exact rule the tool documents. AGENTS.md (18/19): overlapping targets within one batch, and I never root-caused *which* edit collided — I patched by outcome instead. Both caught by verification greps, but each burned a round trip.
4. **First fix batch was scoped to exactly the reported findings** — I didn't anticipate that a cache-corrupted run under-reports; the canonicalheader/G602 wave afterwards was foreseeable the moment I suspected the cache.
5. (Minor) gopls showed stale diagnostics all session (fixed issues kept re-appearing in tool output); I correctly overrode them with real tool runs but never restarted the LSP.

## e) WHAT WE SHOULD IMPROVE (observed this session)

1. **Lint SOP:** sequential-only is now in AGENTS.md; consider also a per-invocation temp `GOLANGCI_CACHE` for parallel-safety, or a `cache clean` step after any anomalous result.
2. **`/mnt/buildcache` persistence warnings** ("no such file or directory" on fact-save) predate this session — environmental, and exactly the flakiness class that produced finding d.1. Worth investigating the mount.
3. **gopls runs without `GOEXPERIMENT=jsonv2`** → stdversion false positives on `json.UnmarshalRead` (an encoding/json/v2 API available under the experiment in 1.26); wiring the env into the LSP config would silence recurring noise. Its unusedfunc notes on `testhelpers_test.go` (`expectError`, `freePort` unused) may be dead helpers worth pruning — pre-existing, files I didn't touch.
4. **Status-report "0 issues" claims should cite command + run mode** (sequential, which cache) — the predecessor's claim rotted within one session; mine now says exactly how it was verified.
5. **Alias-probe test pattern**: the shared-backing-array construction should be the documented idiom (composite literal + reslice) for future "does not mutate config" tests.
6. **cqrs-lint was not re-run this session** (no cqrs code changed; predecessor ran it clean) — but the commit gate should include it uniformly.
7. **Header-casing convention**: canonicalheader forced `X-Outer`/`X-Victim` Pascal-Case; the codebase already had `X-Request-Id`. Codify the convention so test markers don't drift again.

## f) NEXT 50 (prioritized; ≈6 items already harvested into TODO_LIST this session are excluded)

**P0 — ship the OTEL work**
1. Answer §g Q1 → commit (one coherent commit vs per-module core/otel/cqrs).
2. Push the 4 prepared tags (standing user gate from 2026-08-16).
3. otel release prep: drop `replace ../` in `otel/go.mod`, require published core, hermetic re-verify.
4. Tag `otel/v0.1.0` (wave membership = §g Q2); post-tag fresh-consumer proxy test + pkg.go.dev.
5. Pre-commit gate uniformity: run cqrs-lint alongside golangci for the cqrs module.

**P1 — follow-through**
6. Post-push verification for the 4 standing tags (proxy test + pkg.go.dev).
7. Investigate `/mnt/buildcache` warnings; consider isolated per-run lint caches.
8. File cqrs-lite `Provider.Shutdown` ForceFlush issue/PR (§g Q3; verify-before-filing skill applies).
9. Propose httputil ctx-aware `Logging` completion lines (frees TraceHandler correlation).
10. cqrs-htmx `setup`: adopt appkit otel (replaces hand-rolled wiring; ADR-worthy).
11. Harvest the remainder of this + predecessor §f into TODO_LIST when acting on it.

**P2 — polish & hardening**
12. otel middleware benchmark (no-op vs configured) + benchstat; record numbers in README.
13. Sweep satellite tests for `DrainDelay: 0` misuse (hidden 5s tax per shutdown test).
14. Document span-name vs `http.route` asymmetry in otel README (today only in tests).
15. Root README: otel quick-start snippet in Configuration section.
16. cqrs README: link the otel module from the cookbook.
17. otel tests under `-count=2` (global-state bleed guard).
18. Test `NoDrainDelay` + `NoTimeout` combined (SSE + fast shutdown).
19. `WithStdoutMetricReader` for dev parity with `WithStdoutExporter`.
20. Consider `WithMessageEvents` opt-in (byte-count spans).
21. `WithFilteredPaths` method-scoped patterns if ever needed.
22. Document `Provider.Shutdown` idempotency/double-call semantics.
23. Consider `NewTracer`/`NewMeter` component helpers (cqrs parity).
24. Example: SIGTERM prints flushed-span count (demo polish).

**P3 — bigger bets**
25. Baggage correlation-ID helpers (`WithCorrelationID`, cqrs parity).
26. `appkitotel.Transport()` for outbound client spans.
27. Prometheus reader recipe in otel README (stdout + OTLP covered today).
28. errorpages: render `trace_id` on error pages when a span is active.
29. flightrecorder: link snapshot file to active span attr.
30. docs module: emit otel module docs into the generated catalog.
31. `telemetry` umbrella doc page (otel + flightrecorder + health).
32. Route-cardinality guard test: 10k distinct paths → bounded metric series.
33. Evaluate OTel SDK v1.46+ when released (v1.45 pinned).
34. `InitLogger` trace-correlation flag (`LogTraceCorrelation bool`) wiring TraceHandler.
35. Local Jaeger/docker viewing note in otel README.
36. Document SSE long-span histogram implications (10s+ boundary bucket).

**P4 — standing housekeeping**
37. dprint exit-14 on CHANGELOG-only commits (standing P3).
38. go-structure-linter root-package findings acceptance (standing).
39. v1.0.0 exit criteria for core; fold in `OuterMiddlewares`/`ShutdownHooks` as v1-shaped.
40. Go toolchain 1.26.6 bump when nixpkgs carries it (GO-2026-6090/5972).
41. Mechanical API-break check (goapidiff / `go doc` snapshot) at tag time.
42. `go mod tidy` per module post-merge (transitive drift).
43. Wire `GOEXPERIMENT=jsonv2` into the LSP/gopls env (kill stdversion noise).
44. Prune or restore dead test helpers (`expectError`, `freePort` in testhelpers_test.go).
45. Codify Pascal-Case test-marker header convention (canonicalheader).
46. Logging-posture decision (per-request INFO cost, comparison finding 7 — standing P2).
47. realtime SSE-flush E2E test through the default stack (standing).
48. README: document `GOEXPERIMENT=jsonv2` for source builds (standing).
49. FEATURES.md "Consumers" section citing cqrs-htmx ADR-001 (standing).
50. Consider `NoDrainDelay`/sentinel registry mention in core README config table.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Commit strategy (the immediate gate):** one coherent commit for the whole OTEL work (core hooks + otel module + cqrs adapter + docs), or per-module commits matching the repo's `chore(core):`/`chore(satellites):` history style? I can see the history style but not your preference for a change this size.
2. **Release wave:** does `otel/v0.1.0` join the pending push (5 tags total, requires core v0.3.0 to land first since the example needs published core), or wait for a second wave?
3. **Upstream cqrs-lite fix:** prepare the ForceFlush-Shutdown PR (probe-test evidence in hand), or just file the issue and leave it?

---

**Verification state at time of writing:** all 8 modules test+vet+build green (`-race -count=1`, hermetic) · golangci-lint 0 issues per module, verified sequentially with a trusted cache · cqrs-lint not re-run (no cqrs changes this session; clean at predecessor) · nothing committed, nothing tagged, nothing pushed.
