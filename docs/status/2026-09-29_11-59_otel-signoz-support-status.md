# Status Report — OTEL + SigNoz Support Train (otel v0.2.0 work)

**Date:** 2026-09-29 11:59 CEST
**Scope of this report:** the single session that built "superb OTEL + SigNoz support incl. Exceptions and Dashboards" in `github.com/larsartmann/go-appkit/otel`, plus what that session noticed about the surrounding repo. Point-in-time snapshot — it goes stale.
**Verification baseline (fresh at write time):** `go test ./... -race` green (otel module, 44 test funcs), `golangci-lint` **0 issues**, vet+build green, `check-go-directives.sh` green, `check-pin-drift.sh` green, `go-structure-linter` green (AGENTS.md back under the 377-line cap).

---

## a) FULLY DONE

Each item: what / evidence / scope.

1. **OTLP/HTTP export for both signals, one option** — `WithOTLP(...)` + `WithOTLPEndpoint` / `WithOTLPHeaders` / `WithOTLPTimeout` in `otel/otlp.go`. Traces via batch processor, metrics via periodic reader. Base-URL semantics: standard `/v1/traces` + `/v1/metrics` appended (a path on the endpoint becomes a prefix) — this fixed a real trap: the exporters' `WithEndpointURL` normalizes a pathless URL to `/`, so `WithOTLPEndpoint("http://localhost:4318")` would have silently POSTed to `/`.
   _Evidence:_ `otlp_test.go` — 10 tests incl. real OTLP/HTTP round-trips against an in-process protobuf collector stub (decoded `ExportTraceServiceRequest` / `ExportMetricsServiceRequest`, resource assertions). Suite green.
2. **Zero-code setup path** — `Setup()` auto-enables OTLP per signal when `OTEL_EXPORTER_OTLP_ENDPOINT` (or the signal-specific endpoint vars) is set; explicit code wins per signal (safe partial migrations). Plus env-native: `OTEL_SERVICE_NAME`, `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_TRACES_SAMPLER`(+`_ARG`) (achieved by _not_ forcing a sampler anymore — the SDK reads env itself), and new `WithEnvironment` → `deployment.environment`.
   _Evidence:_ dedicated tests for env auto-enable, explicit-beats-env (dead collector receives nothing), mixed explicit-span-exporter + env-metrics, `WithOTLP` + `WithMetricReader` composition, env service name, env resource attrs, env sampler `always_off`.
3. **Exception capture** — `otel/exceptions.go`: `Recovery(logger)` middleware (panic → semconv `exception` event with type/message/stacktrace + error span status + httputil-compatible `panic recovered` log line + 500; `http.ErrAbortHandler` re-panics; nil logger → `slog.Default()`) and `RecordError(ctx, err)` for handled errors. Both produce exactly what SigNoz's Exceptions view indexes (researched: event name `exception`, attrs `exception.type/message/stacktrace`; OTLP :4318/:4317).
   _Evidence:_ 7 tests in `exceptions_test.go` (error panic, non-error panic values, sentinel re-panic, degrade-without-span, nil logger, helper happy path + no-ops).
4. **SigNoz dashboard** — `otel/dashboards/appkit-http-dashboard.json` (schemaVersion v6): request rate, 5xx error rate (LIKE '5%'), p50/p90/p99 latency per route, status-code distribution, top-endpoints table; `service.name` variable (multi+all). Query grammar copied from SigNoz's official `apm/http-api-monitoring.json` template (`rate()`, `p90(durationNano)`, `spanKind = 'Server'`, groupBy shapes). Plus `dashboards/README.md` (import steps, wiring, SigNoz quickstart).
   _Evidence:_ JSON parses; structure field-by-field adapted from the official SigNoz/dashboards repo template. **But see (b)/(d): never imported into a live SigNoz.**
5. **Example app reworked** — `otel/example/main.go`: zero-code OTLP (stdout fallback only when no endpoint env), `Recovery` wired in `ExtraMiddlewares`, `/boom` (handled error → RecordError), `/panic` (panic → Recovery).
   _Evidence:_ live smoke run: `/users/alice` 200, `/boom` 500, `/panic` 500 (recovered), `/health` 200 (filtered); exported stream contained exactly **2 exception events** with `exception.type/message/stacktrace`; `panic recovered` log line present; 3 trace-correlated log lines; graceful SIGINT flushed cleanly.
6. **Docs train** — otel `README.md` (SigNoz-in-3-steps, Exceptions section, zero-code quick start, options table), `doc.go` (SigNoz + Exceptions sections), module `CHANGELOG.md` (`[Unreleased]` with full delta incl. the design-posture change), `AGENTS.md` (module bullet, file table, deps table, 3 new gotchas — line cap satisfied), `TODO_LIST.md` (new P2 release-train item).
   _Evidence:_ structure linter 0 issues (AGENTS ≤377); all guards green.
7. **Known gotchas captured where they hurt** — otelhttp overwrites span status _description_ on 5xx (exception EVENT survives — that's what SigNoz reads); `WithOTLPEndpoint` is a base URL; `Recovery` placement rules. Written into AGENTS.md otel Gotchas + code docs.
   _Evidence:_ status-description behavior pinned indirectly by tests asserting events (the middleware test dropped the description assertion after observing the overwrite).

---

## b) PARTIALLY DONE

1. ~~**otel v0.2.0 release train** — code complete, CHANGELOG drafted under `[Unreleased]`.~~ done — otel v0.2.0 shipped 2026-10-06 (API diff, dating, tag, push, proxy check, integration re-pin all done)
   ~~_Remaining:_ ritual step 1 (mechanical API-break diff vs `otel/v0.1.1` via `git archive` — additions-only expected, **not run**), CHANGELOG dating, annotated tag, push, fresh-consumer proxy check, integration-module re-pin. _Blocker:_ push is user-gated. _Effort:_ M.~~
2. **SigNoz end-to-end validation** — dashboard JSON and the exceptions contract are doc-researched and template-derived, not proven against a live SigNoz ingest (no SigNoz running here).
   _Remaining:_ import dashboard in a real SigNoz, confirm Exceptions view lights up from the example's two exception flavors. _Blocker:_ needs a running SigNoz instance (user infra). _Effort:_ M.
3. **Integration-module coverage of the new surface** — `/integration` pins **PUBLISHED** tags by charter (correctly refuses unreleased APIs), so `Recovery`, `WithOTLP`, env auto-enable are composition-tested only module-locally, not through a full appkit `Service` there. **→ re-pin done 2026-10-06; the full-stack panic-to-exception E2E is watchlisted (composition-gap row)**
   _Remaining:_ after release, re-pin to otel v0.2.0 + add a full-stack panic→exception E2E. _Effort:_ M.
4. **Performance story for the new code** — no benchmark for `Recovery` / `RecordError`; README perf table not extended (panic path is cold-path, but "superb" deserves numbers). _Effort:_ S. **→ routed — TODO_LIST otel polish batch (2026-10-12)**
5. ~~**LSP vs linter mismatch** — LSP keeps reporting a stale `contextcheck` warning at `exceptions.go:44`; real `golangci-lint` says 0 issues. Cosmetic, but it will nag every future session until an LSP restart proves it gone. _Effort:_ S.~~ **Won't implement — moot — LSP lint is documented NOT ground truth (AGENTS 2026-10-07).**

---

## c) NOT STARTED

1. ~~**OTLP logs signal** (slog → SigNoz logs tab) — deliberately out of scope (otel/log API is experimental), but this limitation is **documented nowhere** in the module README.~~ done — 2026-10-12 — the README Design notes state the no-OTLP-logs limitation
2. **Automatic 5xx→exception bridge** — a handler that writes 500 without panicking produces an error span (`hasError`) but NO exception event; SigNoz's Exceptions view won't list it. An opt-in (`WithExceptionsOn5xx`-style) is unplanned-but-plausible. **→ documented as a deliberate non-goal (README 2026-10-12); opt-in only on demand**
3. ~~**SigNoz version-compat note** — schemaVersion v6 imports need SigNoz ≥ ~v0.135; older versions can't import the dashboard. Unstated in `dashboards/README.md`.~~ done — 2026-10-12 — dashboards README version note
4. **Composition-gap watchlist E2Es** (pre-existing TODO P3): otel×realtime SSE spans on the NoTimeout path, otel×health-dashboard filtering — untouched. **→ watchlist — composition-gap row (otel-x-realtime + otel-x-dashboard legs)**
5. **CI repair** — master CI dead 6+ days (`SSH_PRIVATE_KEY` empty), pre-existing P1. **Every green light in this report is local-only; CI never ran these changes.** User-gated (repo secret).
6. **`OTEL_METRIC_EXPORT_INTERVAL` pin test** — the "metrics interval is env-native" claim rests on a source-comment read, not a test. **→ routed — TODO_LIST otel polish batch**
7. **Deferred Register entry for OTLP logs** (trigger: consumer asks for log ingestion) — not added to AGENTS. **→ not added; the README non-goal note owns the revisit trigger**
8. **gzip/compression + retry env notes** — natively supported by the exporters, undocumented in our README. **→ routed — TODO_LIST otel polish batch**

---

## d) TOTALLY FUCKED UP

Radical honesty section.

1. **FEATURES.md otel section is now a split brain I created.** CHANGELOG/README/AGENTS say the new surface exists; `FEATURES.md` §otel still lists only the old one. I know the file-purpose table (FEATURES.md = feature inventory) and skipped it anyway. Impact: next docs-health pass flags it; anyone reading FEATURES sees a lying inventory. _Fix:_ S — one table update.
2. **Dashboard JSON shipped unvalidated against its only consumer.** The single highest-risk artifact of the train (v6 schema quirks, `CompositeQuery` in the table panel, layout `$ref`s) has exactly zero proof it imports cleanly into SigNoz. "JSON parses" ≠ "SigNoz accepts". Mitigation exists (structure copied from official templates) — but copied-from-official is exactly how subtle version skew bites.
3. **The lint-fix phase was sloppy — bulk edits broke the build 3 times.** My regex/python bulk passes created duplicate `err :=` declarations, `undefined: counter`, a mangled import block that summoned `golang.org/x/telemetry/counter` from nowhere, and earlier I nearly shipped `if err := context.Canceled; err != nil`. All caught by the compiler, but that's ~5 wasted cycles from impatience. Correct tool was `golangci-lint run --fix` + targeted edits.
4. **An AGENTS.md merge hit the wrong table row.** My row-merge script folded the otel `exceptions.go` content into the **flightrecorder** module's `middleware.go` row (nearest-match search). Linter stayed green (line count fine) so I nearly shipped corrupt docs; caught on visual inspection and fixed both rows.
5. **Dependency-weight posture change under-communicated:** `google.golang.org/grpc v1.83.1` + `grpc-gateway v2.30.0` landed as **indirect** deps of the otel module. The wire protocol is OTLP/HTTP (no gRPC traffic), but "the exporters are in your tree now" is heavier than the CHANGELOG's one-line posture note conveys. No explicit tradeoff assessment was recorded (e.g. vs keeping exporters consumer-side).
6. **Zero deliberate commits.** All session work landed via ~12 `chore: auto-commit` heuristic commits. No narrative, no clean revert boundaries. (Auto-commit daemon is house policy — but this train touched 15+ files across code/docs/tests and deserved real commits.)
7. **Pre-existing and still fucked:** master CI is a dead signal (P1, 6+ days, every matrix job red → Dependabot PRs false-red by design); go.work untracked but CI reads it; `dprint` exit-14 on CHANGELOG-only commits. Untouched this session (out of scope), but they surround everything above.

---

## e) WHAT WE SHOULD IMPROVE

1. **Verify after every edit batch, not after five.** The build broke 3× because I stacked bulk edits before compiling. Rule: compile + lint after each logical change, even when it feels tedious.
2. **Linter cleanups: `golangci-lint run --fix` first, targeted edits second, regex never.** The regex pass also misfired on non-proto `.Metrics` (metricdata struct) and proto `.GetMetrics()` naming.
3. **Extend the release-ritual doc checklist** — "same-train release-state updates" names AGENTS + TODO_LIST + CHANGELOG but not FEATURES.md. That's exactly the hole this train fell into. Add FEATURES.md to the ritual.
4. **Dashboard artifacts need a validation story** — a `jq`-based structural smoke or (better) a SigNoz import check wherever a SigNoz exists; otherwise every dashboard edit ships blind.
5. **Live smoke tests should be scripted** — today's curl/python one-off proved the wiring but dies with the session. An `example_test.go` (or scripts/) version makes it repeatable evidence.
6. **State non-goals in docs** — "no OTLP logs signal (yet)" and "5xx-without-panic yields no exception event" belong in the README's design-notes/known-limitations, not just in session memory.
7. **Import-block hygiene** — the phantom `golang.org/x/telemetry/counter` import appeared during formatting churn; avoid renaming/adding identifiers that shadow common package names in files being auto-formatted.
8. **Small-n ugliness worth deleting later:** `otlp_test.go` still has a slightly redundant dual receipt-wait in the traces test (request-count + span-count closures); harmless, but it's noise.

---

## f) TOP 50 THINGS WE SHOULD GET DONE NEXT

_Brainstorm ranked roughly by impact; effort: S <30min, M 30min–2h, L >2h. This section is the primary input for `docs-health` HARVEST — per your "wait for instructions", I have NOT harvested; say the word._

| #      | Task                                                                                                                                                                                                                                   | Impact       | Effort | Category    |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ----------- |
| ~~1~~  | ~~Fix FEATURES.md otel section (add v0.2.0 surface: WithOTLP/env-setup/Recovery/RecordError/dashboards)~~ done — FEATURES otel v0.2.0 rows added 2026-09-29; refreshed 2026-10-07/12                                                   | ~~Critical~~ | ~~S~~  | ~~Docs~~    |
| ~~2~~  | ~~Release otel v0.2.0: API-break diff vs v0.1.1, date CHANGELOG, tag, push, proxy check~~ done — otel v0.2.0 shipped 2026-10-06 (tagged, pushed, proxy-verified)                                                                       | ~~Critical~~ | ~~M~~  | ~~Release~~ |
| 3      | Restore CI `SSH_PRIVATE_KEY` secret + one green master run (user-gated)                                                                                                                                                                | Critical     | S      | Infra       |
| 4      | Validate dashboard JSON against a live SigNoz import (needs instance)                                                                                                                                                                  | High         | M      | Quality     |
| 5      | E2E exceptions proof in real SigNoz (example's /panic + /boom visible in Exceptions view)                                                                                                                                              | High         | M      | Quality     |
| 6      | After release: re-pin integration to otel v0.2.0 + full-stack panic→exception E2E through appkit Service **→ re-pin done 2026-10-06; the full-stack panic-to-exception E2E folds into the composition-gap watchlist row**              | High         | M      | Quality     |
| 7      | Add `BenchmarkRecovery`/`BenchmarkRecordError` + extend README perf table                                                                                                                                                              | Medium       | S      | Quality     |
| ~~8~~  | ~~Document non-goals in README: no OTLP logs signal; 5xx-without-panic ⇒ no exception event~~ done — 2026-10-12 — otel README Design notes now state both non-goals (no OTLP logs signal; 5xx-without-panic yields no exception event) | ~~Medium~~   | ~~S~~  | ~~Docs~~    |
| ~~9~~  | ~~SigNoz version-compat note (v6 schema ≥ ~v0.135) in dashboards/README~~ done — 2026-10-12 — dashboards README carries the schemaVersion v6 + unvalidated-against-live-SigNoz note                                                    | ~~Medium~~   | ~~S~~  | ~~Docs~~    |
| 10     | Design + test an opt-in auto-exception-on-5xx (e.g. `Middleware(WithExceptionsOn5xx())`) **→ documented as a deliberate non-goal (README 2026-10-12); build only on consumer demand**                                                  | Medium       | M      | Feature     |
| 11     | Pin `OTEL_METRIC_EXPORT_INTERVAL` env behavior with a test **→ routed — TODO_LIST otel polish batch (2026-10-12)**                                                                                                                     | Low          | S      | Quality     |
| ~~12~~ | ~~Record the grpc-indirect dep tradeoff explicitly (CHANGELOG/AGENTS design note)~~ done — AGENTS deps table + design-posture note                                                                                                     | ~~Medium~~   | ~~S~~  | ~~Docs~~    |
| 13     | Make the example smoke test repeatable (`example_test.go` or scripts/) **→ folds into the TODO_LIST example-smoke harness row (2026-10-12)**                                                                                           | Medium       | M      | Quality     |
| 14     | Add gzip/compression + retry env notes to README production section                                                                                                                                                                    | Low          | S      | Docs        |
| 15     | Deferred Register entry: OTLP logs exporter (trigger: consumer demand / otel/log stabilizes) **→ not added — the README non-goal note (2026-10-12) owns the revisit trigger (otel/log stabilizes or a consumer asks)**                 | Low          | S      | Docs        |
| 16     | Tail-sampling recipe doc (WithSampler vs OTEL_TRACES_SAMPLER combos)                                                                                                                                                                   | Low          | S      | Docs        |
| 17     | SigNoz alert example (5xx error-rate alert JSON) next to the dashboard                                                                                                                                                                 | Low          | S      | Docs        |
| 18     | otel×realtime SSE E2E (spans/metrics on NoTimeout SSE path) — watchlist item                                                                                                                                                           | Medium       | M      | Quality     |
| 19     | otel×health-dashboard E2E (which dashboard routes traced vs filtered)                                                                                                                                                                  | Medium       | M      | Quality     |
| ~~20~~ | ~~Reconcile TODO_LIST P2 "cqrs lacks integration E2E" with the cqrs lifecycle E2E that AGENTS says landed 2026-09-29 (split brain)~~ done — the TODO row now says the cqrs leg is CLOSED                                               | ~~Medium~~   | ~~S~~  | ~~Docs~~    |
| ~~21~~ | ~~`golangci-lint run --fix` pass across all modules to bank the auto-fixable backlog~~ **Won't implement — moot — 0-issues-x12 posture maintained (2026-10-12 fleet scan).**                                                           | ~~Low~~      | ~~M~~  | ~~Cleanup~~ |
| ~~22~~ | ~~Restart/Clear LSP so the stale contextcheck warning on exceptions.go:44 provably dies~~ **Won't implement — moot — LSP lint is documented NOT ground truth (AGENTS 2026-10-07).**                                                    | ~~Low~~      | ~~S~~  | ~~Cleanup~~ |
| 23     | AGENTS slim-down decision (structural; 377-cap will bite the next train again) **→ routed — TODO_LIST P3 AGENTS slim-down**                                                                                                            | Medium       | L      | Cleanup     |
| ~~24~~ | ~~Commit go.work or CI-generate it (CI `go-directives` job vacuous today)~~ done — go.work tracked 2026-09-29 + charter-guarded                                                                                                        | ~~High~~     | ~~S~~  | ~~Infra~~   |
| 25     | BuildFlow dprint exit-14 on CHANGELOG-only commits — upstream fix or document escape hatch **→ routed — TODO_LIST P3 dprint row (upstream Deferred Register)**                                                                         | Low          | M      | Tooling     |
| ~~26~~ | ~~go-structure-linter `exclude_patterns` inert binary — file upstream issue or bump binary~~ **NOT-DO — NOT-DO — CLOSED 2026-09-29: the installed binary honors the yaml config.**                                                     | ~~Low~~      | ~~S~~  | ~~Tooling~~ |
| 27     | Push httputil `docs/integrations/huma.md` (404 fix, Deferred Register) **→ Deferred Register row (httputil huma.md push)**                                                                                                             | Low          | S      | Docs        |
| 28     | errorpages: swap hand-rolled statusRecorder for httputil.ResponseRecorder (USER GATE, open since 09-16) **→ USER gate — TODO_LIST P2 errorpages row**                                                                                  | Medium       | S      | Cleanup     |
| 29     | health: file the upstream go-health recorder-sentinel ask (drafted, USER-gated) **→ USER gate — TODO_LIST P2 upstream-asks row**                                                                                                       | Medium       | S      | Upstream    |
| 30     | File the drafted upstream asks batch (go-sse ReplayFiltered, httputil Logging ctx, NewServerListener go/no-go) **→ USER gate — TODO_LIST P2 upstream-asks row**                                                                        | Medium       | M      | Upstream    |
| 31     | govulncheck on health + security (needs networked machine) **→ TODO_LIST P2 govulncheck (network-blocked env)**                                                                                                                        | Medium       | S      | Quality     |
| 32     | benchstat re-baseline of otel middleware numbers post-1.27.1 (open candidate) **→ TODO_LIST benchstat re-baseline candidate**                                                                                                          | Low          | S      | Quality     |
| ~~33~~ | ~~otel README: cross-link the new flightrecorder metrics hook (frmetrics.go is undocumented in AGENTS file table)~~ done — AGENTS otel file table carries frmetrics.go (2026-10-12); the module bullet + otel README cover the bridge  | ~~Low~~      | ~~S~~  | ~~Docs~~    |
| 34     | Codify "span status description clobbered by otelhttp" as an integration assertion (post re-pin) **→ routed — TODO_LIST otel polish batch**                                                                                            | Low          | S      | Quality     |
| 35     | Example: add metrics-visible-in-SigNoz demo note (counter via Provider meter) **→ routed — TODO_LIST otel polish batch**                                                                                                               | Low          | S      | Docs        |
| 36     | errorpages×otel composition E2E (pretty 500s + exception events together) **→ routed — composition-gap watchlist row (errorpages-x-otel leg noted 2026-10-12)**                                                                        | Low          | M      | Quality     |
| ~~37~~ | ~~flightrecorder×otel doc cross-link (one shared Recorder instance guidance)~~ done — covered — AGENTS flightrecorder gotchas + otel deps table carry the one-active-recorder constraint                                               | ~~Low~~      | ~~S~~  | ~~Docs~~    |
| 38     | Dashboard: exceptions-oriented panel once SigNoz builder supports error-index queries cleanly                                                                                                                                          | Low          | S      | Feature     |
| 39     | Multi-env deploy doc: k8s/Compose snippet with OTEL_* env for appkit services                                                                                                                                                          | Low          | S      | Docs        |
| ~~40~~ | ~~Consider `WithOTLPMetricsInterval` explicit option (only if env-native proves insufficient)~~ **Won't implement — env-native interval has proven sufficient; no option added.**                                                      | ~~Low~~      | ~~S~~  | ~~Feature~~ |
| ~~41~~ | ~~Split-brain sweep: AGENTS vs TODO_LIST vs FEATURES release-state lines after v0.2.0 ships~~ done — release-state lines re-verified 2026-10-12 (zero stale claims)                                                                    | ~~Medium~~   | ~~S~~  | ~~Docs~~    |
| 42     | Battery wave W3 `httpx` module (B1 ResultHandler — error taxonomy parity with errorpages) **→ demand-gated — battery W3 (TODO_LIST P3)**                                                                                               | Low          | L      | Feature     |
| 43     | Battery W4 `worker` supervisor+pool (demand-gated) **→ demand-gated — battery W4 (TODO_LIST P3)**                                                                                                                                      | Low          | L      | Feature     |
| 44     | Battery W5 C2 projection→broadcast folded contract (must-have if cqrs+realtime consumers appear) **→ demand-gated — battery W5 (TODO_LIST P3)**                                                                                        | Low          | L      | Feature     |
| 45     | cqrs README cookbook re-verification vs scenario/v4 (standing ritual, next go-cqrs-lite release) **→ standing ritual — DUE 2026-10-12 (system/v4 v4.13.0 published; TODO_LIST P2)**                                                    | Medium       | S      | Quality     |
| 46     | Consumer-claim drift ritual decision (USER-gated posture question) **→ USER-gated — TODO_LIST Open questions item 1**                                                                                                                  | Medium       | S      | Docs        |
| 47     | `check-pin-drift.sh`: add the otel dashboard JSON to checked artifacts (name/path guard)                                                                                                                                               | Low          | S      | Tooling     |
| ~~48~~ | ~~Template a `.golangci.yml` shared test-exclusion union generator (8 modules carry identical blocks)~~ **Won't implement — generator rejected — per-module .golangci.yml duplication accepted BY DESIGN (Decision 13 amendment).**    | ~~Low~~      | ~~M~~  | ~~Cleanup~~ |
| ~~49~~ | ~~Add `dashboards/` to any packaging decision (should example assets ship with releases?)~~ **Won't implement — moot — dashboards ship in the module tree; no packaging change made.**                                                 | ~~Low~~      | ~~S~~  | ~~Release~~ |
| 50     | PapDashboard/cordis reverse-adoption re-entry triggers review (P3, surfaces the no-TLS gap again) **→ watchlist — TODO_LIST P3 cordis/PapDashboard rows**                                                                              | Low          | S      | Planning    |

---

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Do you have a live SigNoz I can point at (and which version)?** I can't validate the dashboard import or the Exceptions view from here — that needs your instance (or permission to run one). Everything user-facing in this train is doc-researched + locally round-tripped, but SigNoz itself never saw a single span from this work. Which SigNoz version should the dashboard JSON target?
2. ~~**Release now or batch?** Should I run the full v0.2.0 release train (API-break diff → CHANGELOG date → tag → push → proxy check → integration re-pin) immediately, or hold the tag until more trains accumulate? Pushing is yours to authorize either way.~~ done — answered — v0.2.0 shipped 2026-10-06
3. ~~**OTLP logs: build now or wait?** Bringing slog records into SigNoz's logs tab (trace-correlated) requires the experimental `go.opentelemetry.io/otel/log` API — unstable surface in a stable module. Do you want log ingestion badly enough to accept that churn, or do we wait for stabilization?~~ done — answered — WAIT: documented as a non-goal until otel/log stabilizes AND a consumer asks (README 2026-10-12)

---

_Point-in-time snapshot. Section (f) is the HARVEST feed for TODO_LIST/ROADMAP — awaiting your go._
