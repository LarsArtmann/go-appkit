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

~~- **Nothing half-landed.** Every edit this session is applied and verified. The OTEL work as a whole remains **uncommitted** (user gate), which is "prepared", not "partially done".~~ resolved — commit aaa2427 + wave-2 push 2026-09-04; README otel snippet done in module README
~~- Root README otel quick-start snippet (predecessor §f P3-27): still absent — module table row exists, Configuration section does not mention otel. Out of this session's scoped steps; carried in f).~~ resolved — commit aaa2427 + wave-2 push 2026-09-04; README otel snippet done in module README

## c) NOT STARTED

~~- Commit / tag / push of anything (gated on §g Q1/Q2 + the standing push gate).~~ resolved (tags/push 2026-09-04; upstream verified FIXED 2026-09-15; sweeps + benchmark done; DrainDelay sweep clean)
~~- cqrs-lite upstream ForceFlush issue/PR (§g Q3).~~ resolved (tags/push 2026-09-04; upstream verified FIXED 2026-09-15; sweeps + benchmark done; DrainDelay sweep clean)
~~- otel release prep: drop `replace ../`, require published core, fresh-consumer proxy test, pkg.go.dev check.~~ resolved (tags/push 2026-09-04; upstream verified FIXED 2026-09-15; sweeps + benchmark done; DrainDelay sweep clean)
~~- Satellite `DrainDelay: 0` test-suite sweep (realtime/errorpages/docs/flightrecorder).~~ resolved (tags/push 2026-09-04; upstream verified FIXED 2026-09-15; sweeps + benchmark done; DrainDelay sweep clean)
~~- otel middleware benchmark; httputil ctx-aware Logging proposal.~~ resolved (tags/push 2026-09-04; upstream verified FIXED 2026-09-15; sweeps + benchmark done; DrainDelay sweep clean)

## d) TOTALLY FUCKED UP (this session's mistakes, all fixed)

~~1. **I trusted a corrupted lint run.** Running three lint batches in parallel for speed raced the shared `/mnt/buildcache` — the "4 findings" first report was incomplete (3 missing), and the satellites' simultaneous "0 issues" greens were untrustworthy until I redid everything sequentially. Cost: one full extra lint sweep. Root cause: my parallelization choice; the cache location is not concurrency-safe.~~ owned — sequential-lint SOP now in AGENTS; whack-a-mole lesson recorded
~~2. **Linter whack-a-mole on one test helper.** appendAssign fix (`outer[:2]`) → gosec G602; second fix (`make` + index assigns) → makezero; third (composite literal) finally clean. Three rounds because I fixed the _reported_ finding each time instead of reasoning once about the shape satisfying all slice-policing linters.~~ owned — sequential-lint SOP now in AGENTS; whack-a-mole lesson recorded
~~3. **Two multiedit partial failures.** middleware_test.go (6/7): I wrote `old_string` from memory (`resp.Header…` context) instead of copying the file's exact `rec.Header().Get(...)` — the exact rule the tool documents. AGENTS.md (18/19): overlapping targets within one batch, and I never root-caused _which_ edit collided — I patched by outcome instead. Both caught by verification greps, but each burned a round trip.~~ owned — sequential-lint SOP now in AGENTS; whack-a-mole lesson recorded
~~4. **First fix batch was scoped to exactly the reported findings** — I didn't anticipate that a cache-corrupted run under-reports; the canonicalheader/G602 wave afterwards was foreseeable the moment I suspected the cache.~~ owned — sequential-lint SOP now in AGENTS; whack-a-mole lesson recorded
~~5. (Minor) gopls showed stale diagnostics all session (fixed issues kept re-appearing in tool output); I correctly overrode them with real tool runs but never restarted the LSP.~~ owned — sequential-lint SOP now in AGENTS; whack-a-mole lesson recorded

## e) WHAT WE SHOULD IMPROVE (observed this session)

~~1. **Lint SOP:** sequential-only is now in AGENTS.md; consider also a per-invocation temp `GOLANGCI_CACHE` for parallel-safety, or a `cache clean` step after any anomalous result.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)
~~2. **`/mnt/buildcache` persistence warnings** ("no such file or directory" on fact-save) predate this session — environmental, and exactly the flakiness class that produced finding d.1. Worth investigating the mount.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)
~~3. **gopls runs without `GOEXPERIMENT=jsonv2`** → stdversion false positives on `json.UnmarshalRead` (an encoding/json/v2 API available under the experiment in 1.26); wiring the env into the LSP config would silence recurring noise. Its unusedfunc notes on `testhelpers_test.go` (`expectError`, `freePort` unused) may be dead helpers worth pruning — pre-existing, files I didn't touch.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)
~~4. **Status-report "0 issues" claims should cite command + run mode** (sequential, which cache) — the predecessor's claim rotted within one session; mine now says exactly how it was verified.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)
~~5. **Alias-probe test pattern**: the shared-backing-array construction should be the documented idiom (composite literal + reslice) for future "does not mutate config" tests.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)
~~6. **cqrs-lint was not re-run this session** (no cqrs code changed; predecessor ran it clean) — but the commit gate should include it uniformly.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)
~~7. **Header-casing convention**: canonicalheader forced `X-Outer`/`X-Victim` Pascal-Case; the codebase already had `X-Request-Id`. Codify the convention so test markers don't drift again.~~ absorbed (lint SOP in AGENTS; per-invocation cache; buildcache superseded; count-breakdown rule)

## f) NEXT 50 (prioritized; ≈6 items already harvested into TODO_LIST this session are excluded)

**P0 — ship the OTEL work**

~~1. Answer §g Q1 → commit (one coherent commit vs per-module core/otel/cqrs).~~ done — aaa2427
~~2. Push the 4 prepared tags (standing user gate from 2026-08-16).~~ done — push 2026-09-04
~~3. otel release prep: drop `replace ../` in `otel/go.mod`, require published core, hermetic re-verify.~~ done — replace dropped
~~4. Tag `otel/v0.1.0` (wave membership = §g Q2); post-tag fresh-consumer proxy test + pkg.go.dev.~~ done at otel/v0.1.0
~~5. Pre-commit gate uniformity: run cqrs-lint alongside golangci for the cqrs module.~~ done — CI 9b163ce

**P1 — follow-through**
~~6. Post-push verification for the 4 standing tags (proxy test + pkg.go.dev).~~ done — post-push verification 2026-09-04
~~7. Investigate `/mnt/buildcache` warnings; consider isolated per-run lint caches.~~ NOT-DO — buildcache superseded
~~8. File cqrs-lite `Provider.Shutdown` ForceFlush issue/PR (§g Q3; verify-before-filing skill applies).~~ done — verified FIXED upstream 2026-09-15
~~9. Propose httputil ctx-aware `Logging` completion lines (frees TraceHandler correlation).~~ tracked in TODO_LIST P3 (httputil Logging ctx)
~~10. cqrs-htmx `setup`: adopt appkit otel (replaces hand-rolled wiring; ADR-worthy).~~ NOT-DO — cqrs-htmx side
~~11. Harvest the remainder of this + predecessor §f into TODO_LIST when acting on it.~~ done — this audit harvested it

**P2 — polish & hardening**
~~12. otel middleware benchmark (no-op vs configured) + benchstat; record numbers in README.~~ tracked in TODO_LIST P3 (benchstat)
~~13. Sweep satellite tests for `DrainDelay: 0` misuse (hidden 5s tax per shutdown test).~~ done — sweep clean
~~14. Document span-name vs `http.route` asymmetry in otel README (today only in tests).~~ done — README known issue
~~15. Root README: otel quick-start snippet in Configuration section.~~ done — module README covers it
~~16. cqrs README: link the otel module from the cookbook.~~ done — cqrs README links otel
~~17. otel tests under `-count=2` (global-state bleed guard).~~ NOT-DO — no demand
~~18. Test `NoDrainDelay` + `NoTimeout` combined (SSE + fast shutdown).~~ NOT-DO — covered by NoTimeout+NoDrainDelay tests
~~19. `WithStdoutMetricReader` for dev parity with `WithStdoutExporter`.~~ NOT-DO — no demand
~~20. Consider `WithMessageEvents` opt-in (byte-count spans).~~ NOT-DO — no demand
~~21. `WithFilteredPaths` method-scoped patterns if ever needed.~~ NOT-DO — no demand
~~22. Document `Provider.Shutdown` idempotency/double-call semantics.~~ done — README documents Shutdown
~~23. Consider `NewTracer`/`NewMeter` component helpers (cqrs parity).~~ NOT-DO — no demand
~~24. Example: SIGTERM prints flushed-span count (demo polish).~~ NOT-DO — example polish

**P3 — bigger bets**
~~25. Baggage correlation-ID helpers (`WithCorrelationID`, cqrs parity).~~ tracked in ROADMAP.md
~~26. `appkitotel.Transport()` for outbound client spans.~~ tracked in ROADMAP.md
~~27. Prometheus reader recipe in otel README (stdout + OTLP covered today).~~ NOT-DO — G2 battery instead
~~28. errorpages: render `trace_id` on error pages when a span is active.~~ tracked in ROADMAP.md
~~29. flightrecorder: link snapshot file to active span attr.~~ tracked in ROADMAP.md
~~30. docs module: emit otel module docs into the generated catalog.~~ NOT-DO — docs module untouched
~~31. `telemetry` umbrella doc page (otel + flightrecorder + health).~~ tracked in TODO_LIST P2 (telemetry docs bundle)
~~32. Route-cardinality guard test: 10k distinct paths → bounded metric series.~~ NOT-DO — test idea
~~33. Evaluate OTel SDK v1.46+ when released (v1.45 pinned).~~ done — pinned v1.46.0
~~34. `InitLogger` trace-correlation flag (`LogTraceCorrelation bool`) wiring TraceHandler.~~ NOT-DO — InitLogger stays focused
~~35. Local Jaeger/docker viewing note in otel README.~~ NOT-DO — no local jaeger recipe
~~36. Document SSE long-span histogram implications (10s+ boundary bucket).~~ NOT-DO — SSE histograms documented enough

**P4 — standing housekeeping**
~~37. dprint exit-14 on CHANGELOG-only commits (standing P3).~~ tracked in TODO_LIST P3 (dprint)
~~38. go-structure-linter root-package findings acceptance (standing).~~ done — structure-linter zeroed 2026-09-04
~~39. v1.0.0 exit criteria for core; fold in `OuterMiddlewares`/`ShutdownHooks` as v1-shaped.~~ done — core-v1-exit-criteria.md
~~40. Go toolchain 1.26.6 bump when nixpkgs carries it (GO-2026-6090/5972).~~ done — toolchain at 1.26.7; next bump tracked
~~41. Mechanical API-break check (goapidiff / `go doc` snapshot) at tag time.~~ done — Release Ritual step 1
~~42. `go mod tidy` per module post-merge (transitive drift).~~ done — dep sweeps via dependabot (9b163ce)
~~43. Wire `GOEXPERIMENT=jsonv2` into the LSP/gopls env (kill stdversion noise).~~ NOT-DO — gopls env documented instead
~~44. Prune or restore dead test helpers (`expectError`, `freePort` in testhelpers_test.go).~~ NOT-DO — helpers kept
~~45. Codify Pascal-Case test-marker header convention (canonicalheader).~~ done — Pascal-Case convention in tests
~~46. Logging-posture decision (per-request INFO cost, comparison finding 7 — standing P2).~~ tracked in TODO_LIST P2 (logging posture)
~~47. realtime SSE-flush E2E test through the default stack (standing).~~ done — integration SSE flush test
~~48. README: document `GOEXPERIMENT=jsonv2` for source builds (standing).~~ done — README jsonv2 note
~~49. FEATURES.md "Consumers" section citing cqrs-htmx ADR-001 (standing).~~ done — FEATURES Consumers section
~~50. Consider `NoDrainDelay`/sentinel registry mention in core README config table.~~ done — README config table documents sentinels

## g) QUESTIONS I CANNOT ANSWER MYSELF

~~1. **Commit strategy (the immediate gate):** one coherent commit for the whole OTEL work (core hooks + otel module + cqrs adapter + docs), or per-module commits matching the repo's `chore(core):`/`chore(satellites):` history style? I can see the history style but not your preference for a change this size.~~ Answered: one coherent commit (aaa2427); otel joined wave 2 (pushed 2026-09-04); upstream fix landed without filing (verified 2026-09-15)
~~2. **Release wave:** does `otel/v0.1.0` join the pending push (5 tags total, requires core v0.3.0 to land first since the example needs published core), or wait for a second wave?~~ Answered: one coherent commit (aaa2427); otel joined wave 2 (pushed 2026-09-04); upstream fix landed without filing (verified 2026-09-15)
~~3. **Upstream cqrs-lite fix:** prepare the ForceFlush-Shutdown PR (probe-test evidence in hand), or just file the issue and leave it?~~ Answered: one coherent commit (aaa2427); otel joined wave 2 (pushed 2026-09-04); upstream fix landed without filing (verified 2026-09-15)

---

**Verification state at time of writing:** all 8 modules test+vet+build green (`-race -count=1`, hermetic) · golangci-lint 0 issues per module, verified sequentially with a trusted cache · cqrs-lint not re-run (no cqrs changes this session; clean at predecessor) · nothing committed, nothing tagged, nothing pushed.
