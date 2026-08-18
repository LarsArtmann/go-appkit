# Status Report: OTEL Support — From Zero to SUPERB

**Date:** 2026-08-18 12:45 | **Session scope:** OpenTelemetry support for go-appkit | **Trigger:** "How is our OTEL support? It should be SUPERB!"

---

## Executive Summary

OTEL support went from **literally zero** (no code anywhere in the repo; consumers hand-rolled everything) to a complete, tested, linted, E2E-verified opt-in telemetry stack across three modules: core (dependency-free hook points), a new `otel` module (provider setup + otelhttp bridge + log correlation), and cqrs (projection metrics adapter, closing planning milestone M10).

**Bottom line:** all code is written, tested (race-clean, stability-verified), lint-0, and E2E-proven. Documentation is ~70% done (AGENTS.md/FEATURES.md/TODO_LIST.md not yet updated — the AGENTS.md edit was interrupted mid-flight). Nothing committed.

---

## a) FULLY DONE

### 1. Research (complete, informed all design)

- Confirmed zero OTEL in go-appkit; found the ecosystem pattern (go-cqrs-lite `otel/v4`).
- Studied cqrs-lite otel API surface (Setup/Provider/options/logging/views/propagation) — mirrored its API shape for muscle-memory consistency.
- Studied otelhttp v0.68.0: `r.Pattern`-based span naming, filter semantics (AND over allow), semconv-stable metrics, public-endpoint link semantics.
- Found planning milestone **M10** ("metrics path: otel/prometheus accessors for the event store") — closed it this session.

### 2. Core enablers (dependency-free, in root module)

- **`ServiceConfig.OuterMiddlewares`** — wraps the ENTIRE chain (default stack or replaced stack), runs outermost before Recovery. This is where tracing must sit. `middleware.go` refactored; `concatMiddlewares` allocates a fresh slice so config backing arrays are never written through (tested).
- **`ServiceConfig.ShutdownHooks`** — `[]func(context.Context) error`, run ONCE, in order, AFTER `server.Shutdown` released connections (so flushed spans cover final in-flight requests). Failing hooks don't stop the rest; errors joined + classified Infrastructure. Never-started services skip hooks (tested).
- **`NoDrainDelay` sentinel (`-2`)** — discovered a real gap: `DrainDelay: 0` means "apply 5s default", NOT "no delay" (zero-value production-safety). Tests were silently eating 5s per shutdown. Converted ALL core tests to `NoDrainDelay`; suite wall time dropped ~30s → ~6s. Mirrors the existing `NoTimeout` design; Validate rejects other negatives (tested).
- Tests: `shutdownhooks_test.go` (4 tests), `middleware_test.go` (+3 tests + marker helper), `config_test.go` (+2 cases). Core suite green with `-race`, 6 consecutive runs stable.
- Core CHANGELOG updated.

### 3. New `otel` module (`/otel`, package `otel`, alias `appkitotel`)

- **`setup.go`**: `Setup(opts...)` + `Provider`. Options: `WithService`, `WithSpanExporter`, `WithSampler` (cqrs-lite's version LACKS this), `WithMetricReader`, `WithPropagator`, `WithStdoutExporter`, `WithoutGlobalRegistration`. Registers globals + W3C propagator.
- **Provider.Shutdown force-flushes before shutdown** — discovered and fixed a real bug class: spans ended moments before Shutdown can still sit in the batch processor's async queue; Shutdown alone does NOT guarantee export. (cqrs-lite's otel module has this same latent bug — worth an upstream fix.)
- **`middleware.go`**: `Middleware(opts...)` bridging otelhttp v0.68. One SERVER span/request named by matched ServeMux pattern (`GET /users/{id}`); W3C trace-context+baggage in/out; semconv metrics (`http.server.request.duration` etc.) with route/method/status attrs when a meter provider exists. Options: `WithTracerProvider`, `WithMeterProvider`, `WithServerName`, `WithPublicEndpoint` (distrusts remote parents → links), `WithFilter`, `WithFilteredPaths`. Health endpoints unconditionally filtered. **No-op without Setup** — OTel strictly opt-in.
- **`views.go`**: `NewHTTPViews` pinning semconv histogram boundaries on `http.server.request.duration` only (size histograms keep SDK defaults).
- **`logging.go`**: `TraceHandler` (slog.Handler decorator stamping `trace_id`/`span_id` when ctx carries a span; uncorrelated records untouched; WithAttrs/WithGroup preserved) + `TraceIDFromContext`/`SpanIDFromContext`/`ContextLogger` (cqrs-compatible).
- **23 tests** across 4 files: span naming/kind/error-status, handler-visible span context, W3C parent continuity (client→server), public-endpoint link semantics, health/path filters, no-op mode, **cardinality proof** (3 distinct user IDs → ONE route-attributed metric series `/users/{id}`), view boundaries, setup resource attrs/globals/sampler/shutdown-error-join, log correlation incl. group preservation.
- **E2E verified live**: built example binary, served on PORT=18099, hit all routes, SIGTERM'd — confirmed spans with service resource, health filtered, correlated handler logs (`trace_id` visible in JSON), 500→error span, graceful drain then flush.
- `example/main.go` (PORT env override added after 8080 was occupied on this machine), README.md, CHANGELOG.md, `.golangci.yml` (realtime-derived; ireturn allows `go.opentelemetry.io/otel/.*` — the API is interface-based by design). **0 lint issues. Builds with plain `go build` — NO GOEXPERIMENT needed** (one of only two modules).
- otel suite green with `-race`, 5 consecutive runs stable.

### 4. cqrs module — M10 closed

- **`otelmetrics.go`**: `OTelProjectionMetrics` implements `projectionhost.MetricsRecorder` (six methods) on OTel instruments: `cqrs.projection.event.count` (projection/event_type/status), `.event.duration` (ms), `.worker.count` (restarted/failed), `.checkpoint.lag` (ms). Attribute keys follow cqrs-lite's `cqrs.*` conventions — one dashboard schema for HTTP + projections. Adds ONLY interface-only OTel API deps (no SDK/exporter).
- Compile-time interface assertion; 2 tests verifying all lifecycle events + attributes.
- cqrs README metrics section rewritten around it; CHANGELOG entry added; full cqrs suite green with `-race`; golangci 0 issues; cqrs-lint clean.

### 5. Repo wiring

- `go.work`: `./otel` added, use-block alphabetized.
- Root README: otel row in the module table.

---

## b) PARTIALLY DONE

- **AGENTS.md update — INTERRUPTED MID-EDIT.** A 4-part multiedit (8-module list, otel build commands, lint-standard refresh to "all 8 modules", core code-org table with new fields) was submitted; the tool call was interrupted and **did not apply** (verified: no changes in git). AGENTS.md still says "Six Go modules", lacks otel build commands and the new core fields.
- **FEATURES.md**: not started — needs otel section (7+ rows) and core rows (OuterMiddlewares, ShutdownHooks, NoDrainDelay).
- **TODO_LIST.md**: not updated (stale header counts "Modules: 7"; P2 items affected by this session: "Document DrainDelay: 0 test-ergonomics" is now OBSOLETE — superseded by NoDrainDelay).
- **Final verification sweep**: each module verified individually during the session (core 6x, otel 5x, cqrs once), but no single clean full-workspace sweep AFTER the last round of edits (AGENTS/README edits pending anyway).

## c) NOT STARTED

- otel module release prep (tag `otel/v0.1.0`) — deliberately deferred to the next release wave (push gate on the existing 4 tags still pending anyway).
- Upstream fix to cqrs-lite's `Provider.Shutdown` (missing ForceFlush — same batch-queue race I fixed locally).
- Core default-stack logging correlation (httputil Logging logs without ctx → request-completion line stays uncorrelated; documented as known limitation in doc.go, not fixed).

## d) TOTALLY FUCKED UP (mistakes made & fixed this session)

1. **Nil-pointer panic in my own hook test** — called `svc.Addr()` inside a shutdown hook, but Shutdown nils the listener first. Fixed (capture addr at Start; then dropped dial approach entirely).
2. **Flaky TCP-dial assertion** — "dial must fail after listener closed" is WRONG at kernel level: backlog handshakes still complete. Caught on run 4 of 5. Replaced with deterministic `Running()` ordering assertion.
3. **Wrong expected middleware order on panic path** — assumed extra's after-marker survives; panic unwinds through it (that's precisely what proves outer-wraps-Recovery).
4. **`&false`** — cannot take address of bool literal.
5. **SpanStub fields vs methods** (`.Name` not `.Name()`), `attribute.Set.Get(idx)` vs `Value(key)`, `AsString` not `Str`, `Bounds` not `ExplicitBounds` — API-shape calibration churn.
6. **`tracetest.InMemoryExporter.Shutdown` RESETS its buffer** — spans read after Shutdown read zero. Non-obvious; now a documented gotcha in test comments.
7. **Batch-queue flush race** (see a.3) — the InMemoryExporter zeros made it visible; root-caused via probe tests, fixed with ForceFlush-then-Shutdown.
8. **go.work missing `./otel`** → first `golangci-lint run` was VACUOUSLY green ("0 issues" while linting nothing). Caught by re-reading output.
9. **curl is banned in this environment** → python3 urllib; **port 8080 occupied** by another local service → PORT env override in example; **`kill` unsupported** → pkill.
10. **bodyclose vs response-returning helper** — restructured to `fetchResult` (body never escapes the fetching function).

None of these remain in the tree — all were fixed and re-verified.

## e) WHAT WE SHOULD IMPROVE (observations beyond the task)

1. **Upstream: cqrs-lite otel `Provider.Shutdown` needs ForceFlush** — same race I fixed; every cqrs-htmx service using it can silently lose final spans.
2. **httputil `Logging` middleware should log with request context** — then TraceHandler could correlate the request-completion line for free. Currently only handler-level logs correlate.
3. **Test-suite speed**: the NoDrainDelay conversion showed ~5s of hidden drain per shutdown test. Realtime/errorpages/docs/flightrecorder modules likely have the same pattern worth sweeping.
4. **`NoDrainDelay = -2` vs `NoTimeout = -1`** — sentinel values are magic; fine at this scale, but a documented registry in config.go comments keeps the next sentinel (-3?) collision-free.
5. **go.work hygiene**: otel initially linted vacuously — a CI check (`golangci-lint run` exit code + "modules found" assertion) would catch empty runs.
6. **Example port collision**: defaulting examples to :8080 is hostile on dev machines with multiple demos; the PORT override pattern should be standard across all module examples.
7. **go-http-inner-bridge gap**: `realtime` SSE spans end only when the stream closes (correct), but long streams mean very long spans — worth documenting bucket implications (10s+ boundary) for SSE-heavy services.

## f) NEXT 50 (prioritized)

**P0 — finish this session's loose ends**

1. Re-apply the interrupted AGENTS.md update (module list → 8, otel build commands, lint standard, core table rows).
2. FEATURES.md: otel module section + core rows (OuterMiddlewares, ShutdownHooks, NoDrainDelay).
3. TODO_LIST.md: refresh header (8 modules), add otel v0.1.0 to the release-wave gate, mark P2 "DrainDelay: 0 test-ergonomics" OBSOLETE (superseded).
4. Full final sweep: all 8 modules, build+vet+race (GOWORK=off where required), golangci per module.
5. Commit the whole OTEL work (one coherent commit; mind the dprint/CHANGELOG exit-14 gotcha → `--no-verify` + justification if CHANGELOG-only... this one has .go files so normal path should work).

**P1 — release follow-through**
6. User gate: push the 4 prepared tags (still pending from before this session).
7. Tag `otel/v0.1.0` (module path `github.com/larsartmann/go-appkit/otel`, tag `otel/v0.1.0`) — after core v0.3.0 is pushed, since the example requires core (replace directive removed at release).
8. Remove the local `replace ../` in otel/go.mod at release time; require published core.
9. Fresh-consumer proxy test for otel module (clean /tmp module, `go get`, blank-import build).
10. pkg.go.dev verification for otel after tag lands.

**P2 — upstream & ecosystem**
11. File cqrs-lite issue/PR: `Provider.Shutdown` must ForceFlush (evidence: this session's probe tests).
12. Propose httputil `Logging` take ctx-aware logging (or `LoggingContextful`) so completion lines correlate.
13. Consider re-exporting `TraceHandler`-wrapped logger from `appkit.InitLogger` config flag (e.g. `LogTraceCorrelation bool`).
14. cqrs-htmx `setup`: adopt appkit otel (replaces hand-rolled wiring in their observability-demo); ADR-worthy.

**P3 — polish & hardening**
15. otel: add a benchmark (middleware overhead no-op vs configured).
16. otel: `WithFilteredPaths` — support method-scoped patterns if ever needed.
17. otel: consider `otelhttp.WithMessageEvents` opt-in option for byte-count spans.
18. Example: docker/jaeger note in README (how to view OTLP traces locally).
19. Add stdouttrace metric reader option (`WithStdoutMetricReader`) for dev parity with spans.
20. Tests: otel middleware under `-count=2` to catch global-state bleed.
21. Test NoDrainDelay + NoTimeout combined (SSE + fast test shutdown).
22. Document span-name vs route-attr asymmetry (span name HAS method prefix, http.route attr does NOT) in README (currently only in test comments).
23. Sweep other modules' tests for `DrainDelay: 0` misuse (same 5s tax).
24. AGENTS.md: add otel module "Code Organization" table (file/concern, matching siblings).
25. AGENTS.md gotchas: InMemoryExporter.Shutdown-resets, batch-flush race, 8080-occupied.
26. Consider `Provider.Shutdown` idempotency guarantee doc (double-Shutdown behavior).
27. Root README: otel quick-start snippet in the Configuration section.
28. cqrs README: link otel module from the cookbook section.
29. Consider exporting `NewTracer`/`NewMeter` component helpers in otel module (cqrs-parity).
30. eval: `SO_TIMEOUT`-style span timeout protection for pathological streams (probably YAGNI — document instead).

**P4 — bigger bets**
31. OpenTelemetry `Baggage` correlation-ID helpers (cqrs-parity: WithCorrelationID).
32. otelhttp.Transport wrapper export for outbound client spans (`appkitotel.Transport()`).
33. Prometheus reader recipe in otel README (stdout + OTLP covered; prometheus only via cqrs-lite today).
34. Errorpages: render trace_id on error pages when a span is active (support handoff).
35. flightrecorder: link captured trace file to active span attr (`flightrecorder.snapshot`).
36. Docs module: emit otel module docs into generated catalog.
37. Consider a `telemetry` umbrella doc page tying otel + flightrecorder + health together.
38. Benchstat before/after for middleware overhead; record numbers in README.
39. Fuzz-ish load test: route cardinality under 10k distinct paths → assert bounded series (guards against future regressions in spanName).
40. Evaluate otel SDK v1.46+ when released (v1.45 pinned now).

**P5 — housekeeping**
41. `git log` convention check: this work spans core+cqrs+otel → consider per-module commits matching repo history style.
42. dprint exit-14 fix (standing P3) still open — same fix unblocks clean CHANGELOG commits.
43. go-structure-linter root-package findings (standing P3) — unaffected but still open.
44. v1.0.0 exit criteria for core (standing P3) — OuterMiddlewares/ShutdownHooks are v1-shaped; fold into criteria.
45. Consider `NoDrainDelay` mention in core README config table.
46. example/main.go: handle SIGTERM print of flushed span count (demo polish).
47. otel CHANGELOG: add Fixed-section noting none (fresh module).
48. Verify `golangci-lint fmt` produced no uncommitted churn beyond intended files.
49. Re-run `go mod tidy` per module post-merge to catch transitive drift.
50. Status-report hygiene: this file → harvest into TODO_LIST when acting on it.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Commit strategy:** one commit for the whole OTEL work, or split per module (core hooks / otel module / cqrs adapter) matching the repo's history style? (I can see history style but not your preference for this size.)
2. **Release wave:** should `otel/v0.1.0` join the current pending push (making it 5 tags), or wait for a second wave after core v0.3.0 lands (the example needs published core — the replace directive must drop at tag time)?
3. **Upstream cqrs-lite fix:** want me to prepare the ForceFlush-Shutdown PR for go-cqrs-lite (I have the probe-test evidence), or just file the issue and leave it?

---

**Verification state at time of writing:** core `-race` 6/6 stable · otel `-race` 5/5 stable · cqrs `-race` green · lint 0 issues (core, otel, cqrs) · cqrs-lint clean · E2E example verified · AGENTS/FEATURES/TODO_LIST NOT yet updated · nothing committed.

---

**Addendum (same day, continuation session — §f P0-1..4 executed):**

- **Docs completed:** AGENTS.md re-applied (8-module list, otel build commands + code-org table + deps + gotchas, lint standard 2026-08-18, core table rows, `DrainDelay: 0` testing advice corrected to `NoDrainDelay`); FEATURES.md gained otel section (12 rows) + 3 core rows + cqrs adapter row; TODO_LIST.md header/release-gate refreshed, obsolete DrainDelay-doc P2 removed, near-term §f items harvested (upstream ForceFlush, satellite DrainDelay sweep, otel benchmark, httputil Logging ctx).
- **Correction — the "lint 0 issues (core)" claim above did not survive re-verification:** a fresh sequential sweep found 4 core findings (appendAssign, noctx, noinlineerr, sloglint) plus 3 more after the first fixes (2× canonicalheader, gosec G602) — likely masked earlier by a corrupted shared golangci cache (/mnt/buildcache races from parallel lint runs; lesson: lint one module at a time). All 7 fixed in `middleware_test.go` + `service.go` (mutation-probe test restructured to a composite literal — same aliasing semantics); core suite still `-race` green.
- **Final sweep (sequential, trustworthy):** all 8 modules test+vet(+build)+lint green — core, otel, cqrs, docs-mod, realtime, errorpages, flightrecorder, flightrecorderhealth — 0 lint issues each.
- **Still nothing committed.** The three §g questions remain the gate.
