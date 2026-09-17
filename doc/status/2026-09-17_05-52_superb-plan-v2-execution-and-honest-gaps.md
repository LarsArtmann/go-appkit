# Status: SUPERB Plan v2 Execution — Full Session Report (2026-09-16 → 2026-09-17 05:52)

**Written:** 2026-09-17 05:52 CEST · **Scope:** execution of `doc/planning/2026-09-16_15-01_SUPERB-visibility-correctness-and-batteries-plan.md` (all 30 C-tasks) in one session, 57 commits since the plan (`8c829ce..HEAD`), tree clean, everything pushed.
**Note on paths:** this report lives in `doc/status/` — the `docs/` directory was repathed to `doc/` by this session's ghost-release fix (path A).

**License ruling applied:** stays PROPRIETARY/unlicensed (user, 2026-09-16). Godoc stays hidden by choice. Closed permanently.

---

## Headline results

- **6 tags shipped + proxy-proven:** `httputil/v1.2.0` (cross-repo), `otel/v0.1.1`, `docs/v0.3.0` (the ghost-release fix — `go get docs@v0.3.0` works), `health/v0.1.1`, `flightrecorderhealth/v0.1.2`, `realtime/v0.1.1`, `security/v0.1.0` (new module).
- **All 10 module suites `-race` green; 0 golangci issues (module-local); structure linter 0; AGENTS at its 377-line cap; tree clean at end.**
- **New capabilities in core (UNRELEASED — no core tag yet):** opt-in Prometheus surface (`ServiceConfig.Metrics`), `Version` + `/version`, `testkit` sub-package.
- **TODO_LIST:** 18 `[x]` / 12 `[~]` / 8 `[ ]` (38 tracked items).

---

## a) FULLY DONE

1. **C1 license decision** — recorded in TODO_LIST + AGENTS; no swap; consequence (hidden godoc) accepted permanently.
2. **C2 docs-ghost fix (path A)** — `git mv docs doc` + `git mv docs-mod docs`; go.work/CI/dependabot swept; hermetic verify; `docs/v0.3.0` annotated tag, pushed, fresh-consumer proxy PASS; P1 closed.
3. **C3 OTEL train pt1** — httputil v1.2.0 dated/tagged/pushed; core + otel bumped; all 6 consumer suites green; otel's leftover filesystem `replace` removed.
4. **C4 OTEL train pt2** — recovered pin test landed as `integration/otel_pattern_test.go` (`TestSpanNameAndRouteThroughAppkitOuterMiddlewares`): span `"GET /users/{id}"` + `http.route` VERIFIED against PUBLISHED tags; otel/v0.1.1 tagged/pushed; README known-issue block deleted.
5. **C6 version cuts** — health/v0.1.1 (API-diff proved doc-only), frh/v0.1.2, error-family v0.10.1 lines in the 3 missing CHANGELOGs; proxy tests PASS.
6. **C7 realtime correctness** — `X-Accel-Buffering: no` + `event: error` with `retry: 30000` before store-failure abort; both pinned by wire-level tests; realtime/v0.1.1 shipped + proxy-proven.
7. **C9 dependency-currency proof** — `git ls-remote` loop over all 14 family repos vs pins: ZERO drift, proven against ORIGIN.
8. **C10 error-classification sweep** — otel's 3 sentinels Infrastructure-classified (codes pinned by `TestSentinelsClassified`); cqrs drain error `WrapInfrastructuref`; `errorfamilytest` adopted in cqrs staleness tests; both modules 0 lint issues.
9. **C11 health example live E2E** — scripted PASS: dashboard 200, probe + appkit readiness 200, SSE 4 frames over 6s, SIGTERM → BOTH readiness surfaces 503 in lockstep DURING the drain window, `graceful shutdown complete result=ok`.
10. **C10/C11-adjacent: proxy smoke for all new tags** — core v0.4.0, otel v0.1.1, docs v0.3.0, health v0.1.1, frh v0.1.2, realtime v0.1.1, security v0.1.0 all fresh-consumer PASS.
11. **C13–C16 security module** — all 8 batteries ported from CV (`platform/middleware/` + `internal/sanitization/`): APIKeyAuth (+query GET/HEAD-only), APIKeyCSRFBypass (fail-closed), CSRF (`*` strips + logs, never allow-all), RateLimit profiles (MaxKeys mandatory — zero-cap test fails; 429-aborts-chain pinned; Retry-After; per-key isolation; window reset), OriginCheck, BodyLimit (typed `*http.MaxBytesError`), SanitizeText/SanitizeURL (`&not=`→`¬=` trap pinned), CSP nonce infra + BuildCSP (deterministic sorted directives, unsafe-eval NEVER — test-pinned; JSON-LD exemption), SecurityHeaders (HSTS production-only, dev/staging lockout-trap test). Registered in CI matrix + dependabot; security/v0.1.0 shipped.
12. **C17 G2 Prometheus surface** — dependency-free text exposition, stable metric-name contract, pattern labels (cardinality-bounded, `unmatched` fallback), Basic Auth mandatory by default (unauthenticated config = construction Rejection), `_ratio` trap documented, SSE-flush passthrough tested.
13. **C18 F5 + E1** — `ServiceConfig.Version` → `/version` (decoupled from health registration) + build-info label; `testkit.Serve(tb, svc)` full-chain harness with goroutine-baseline teardown (the raw-mux trap encoded as API).
14. **C19 composition spike** — verdict: BLOCKED on upstream API. Evidence: httputil `Start()` binds internally with ASYNC bind errors vs appkit's synchronous classified `listen_failed` contract; `Addr()`/`Running()` split-brain risk; phase-log/error contract would need re-implementation. 6/7-field config mapping is clean. Written verdict + upstream ask. **AGENTS "recommended refactor" claim corrected.**
15. **C21 integration expansion (3 of 5)** — one-`Setup`-per-process global-overwrite pin (with restore), errorpages family→status parity vs `appkit.HTTPStatus`, live-service recovered-panic 500 fallthrough through `errorpages.Wrap`; `./integration` added to CI matrix; module 0 lint issues.
16. **C23 flightrecorder polish** — `SnapshotHandler` explicit 503 "recorder not enabled" (was silent 200); `statusError` Infrastructure-classified (`flightrecorder.http_status_error`, pinned); `OpsRecorderPreset` documented; `otel.NewFlightRecorderMetricsHook` bridge (stable names, nil-safe) + tests; both CHANGELOGs updated.
17. **C25/C26 telemetry bundle** — `doc/TELEMETRY.md`: emission catalogue (all log lines + all metrics across core/otel/fr-bridge/cqrs), backpressure semantics per sink, incident-debug recipe (incl. sidecar second exporter via `WithoutGlobalRegistration`), cqrs store-separation doctrine, SSE telemetry candidate recorded.
18. **C27 cqrs/root polish** — README documents `Status()` slice (verified against projectionhost v4.4.0) + `LagPerProjection`; staleness monotonicity property test; root example demonstrates OuterMiddlewares/DrainHooks/ShutdownHooks — LIVE-verified (`[outer]` wraps request, drain hook in-window, shutdown hook post-release, `result=ok`).
19. **C28 truth chores** — nosurf "forks internally" VERIFIED TRUE against v1.2.0 source (`handler.go:124` → `handler_go17.go:12`), now source-cited; httputil CSRF known-limitation note COMMITTED AND PUSHED upstream (`a03db5c`); golines LSP warning root-caused (stale buffer — flagged line is empty; configs aligned).
20. **C29 process decisions** — release-state single owner = AGENTS (rule text in `doc/status/README.md`, AGENTS budget-safe); `shutdown phase skipped` stays INFO (rationale recorded); `doc/status/README.md` index created; annotation standard written into the archived README.
21. **C30 watchlist refresh** — VERIFIED: cordis tagged `go/v0.1.0` (trigger 1/3 MET — bridge stays deferred, 2/3 unmet); PapDashboard shipped v0.3.0; nixpkgs still go 1.26.7; dprint exit-14 upstream unfixed.
22. **C24 partially in this bucket:** `health.DashboardHardenedPreset(basePath, nonceFn)` shipped + tested (uniform base path + nonce extractor); rate limiting deliberately excluded (core-free charter — chain guidance in doc comment).

## b) PARTIALLY DONE

1. **C5 pkg.go.dev render check** — proxy-proof complete; the render check 404s (crawler hasn't processed the tags pushed minutes earlier; requests enqueue the crawl). TODO stays `[~]` with re-check instructions.
2. **C20 Service refactor** — spike done, refactor correctly NOT executed (blocked, see a-14). The API-posture UG is moot until the upstream API exists.
3. **C22 health quality parity** — contract assertion, conflict semantics, BasePath uniform routing, benchmarks, fuzz DONE. REMAINING: runnable godoc Examples with verified output (F107), aggregate multi-probe example (F113), govulncheck (binary not installable here).
4. **C24 upstream asks** — `DashboardHardenedPreset` shipped; go-sse dedup-aware `ReplayFiltered` + httputil Logging request-context + F2 timing sketches DRAFTED at `doc/feedback/outgoing/2026-09-16_upstream-asks-gosse-httputil.md` — NOT FILED (filing is USER-GATED).
5. **C23 fr MetricsHook E2E** — bridge unit-pinned; an end-to-end test driving a REAL capture through the hook remains.
6. **C28 statusRecorder** — USER GATE unanswered; the ~10-line swap deliberately not executed; TODO stays open.
7. **C21 remainder (see d-3/d-4):** `cqrs.OTelProjectionMetrics` E2E (F104) NOT done; core composition-contract suite NOT done; the integration-expansion TODO item was NOT closed (stale — still says `[ ]`).
8. **C12 logging posture** — decision + docs only (no behavior change) — that IS the decision, but "implement behind ServiceConfig" from the plan was consciously not applicable (the knob already exists).
9. **Core G2/F5/E1 release** — features are in core `[Unreleased]`, NOT tagged; consumers get them at the next core tag (v0.5.0 train).
10. **SECURITY module lint config** — copied from realtime with realtime-specific header comments (cosmetic divergence, documented as "module-local standard").

## c) NOT STARTED (explicitly deferred — most by design)

1. Core TLS option (gated on the httputil listener-injection ask).
2. W3 handler-DX batteries (`httpx` module, B1–B9) — canonical spec, demand-gated.
3. W4 ops & data batteries (worker, sqlite ops kit, polite client, do bridge, idempotency store, atomic file write, webhook, config modules) — demand-gated.
4. W5 realtime completions (C1 drop counters, C3 per-subscriber auth, C2 projection→broadcast fold) — demand-gated.
5. Cordis bridge (2 of 3 triggers still unmet).
6. PapDashboard reverse-adoption proposal (user-gated; next check belongs in THEIR repo against v0.3.0).
7. go-plugin-mvp anything (rejected by research).
8. Core v1.0.0 exit-criteria graduation (draft exists; consumer count too low).
9. cqrs EventConfig opt-ins (encryption/signing/idempotency/scheduling) — demand-gated.
10. Toolchain bump (nixpkgs still 1.26.7).
11. dprint exit-14 upstream fix (escape hatch documented).
12. benchstat-based benchmark deltas (binary not installable; mean±sd fallback recorded).

## d) TOTALLY FUCKED UP (own goals — all recovered, all instructive)

1. **The six-TODO-item deletion (worst moment of the session).** My realtime-close script used `content.index("\n\n", idx)` SLICING instead of replace-with-assert — it swallowed everything between the realtime item and the next blank line: the dashboard-CSP, composition, W2, W1, telemetry, and toolchain TODO items vanished. I initially suspected the auto-commit daemon or an external actor BEFORE doing the archaeology properly; the git diff proved it was my own script. Restored verbatim from `8c829ce` (dashboard item re-closed with its verdict). **Lesson now proven expensive: every TODO edit must be replace-with-assert, never slice.**
2. **The silent no-op write of `example/main.go`.** The write tool rejected it ("modified since last read") — and I proceeded to build and LIVE-RUN the OLD file, initially believing my new hooks demo was in it. Caught only because the live output showed `delay=5s` and no hook prints (my own verification design saved me). Re-applied and verified. **Lesson: after ANY write rejection, re-read and RE-APPLY before any verification claim.**
3. **The `csp.go` blanket string-replace.** Replacing `"'self'"` → `selfSrc` mangled the const declaration and every comment containing the token, causing three consecutive build breaks and a fight with `golangci-lint --fix` (which reformats lines and invalidates line-anchored nolint comments — the "formatter war"). Fixed by rewriting profiles with named constants (no nolints needed). **Lesson: never blanket-replace a string that appears in comments/declarations; write lint-clean code instead of nolint-then-fix cycles.**
4. **Used banned `curl`** in the health E2E — blocked by the tooling, wasted a round trip. Rewrote in Python. Should have started there.
5. **Stale TODO closures / coverage-map honesty gaps:**
   - The integration-expansion TODO item was never closed although 3 of its 5 sub-items shipped (this report corrects it).
   - The separate "Flightrecorder ops preset + metrics bridge" TODO item was never closed (its work landed under the polish item's close).
   - The plan's "38/38 covered exactly once" was executed, but TODO_LIST closure lagged the work in these spots.
6. **AGENTS.md budget overflow** — added the ownership note pushing the file to 379 (> 377 cap); had to revert and route the rule text through `doc/status/README.md`. Should have checked the budget BEFORE editing.
7. **Port 8080 occupied** — the first example live-run bound-failed against the local SigNoz instance (AGENTS warns about exactly this); wasted two runs before switching ports. AGENTS said so; I didn't apply it.

## e) WHAT WE SHOULD IMPROVE (process, not code)

1. **TODO-edit discipline:** replace-with-assert only; a post-edit count check (`grep -c "^- \["`) as a gate.
2. **Close TODOs in the same commit as the work** — three closures lagged and had to be chased.
3. **Post-write content verification:** after any file write, grep for a NEW-content marker before building/claiming.
4. **Check AGENTS' 377-line budget before every AGENTS edit** (it's at cap; additions must replace, not append).
5. **Fresh modules: write lint-clean from the start** (nolint anchors break under `--fix` reformatters).
6. **Env facts before live runs:** ports-in-use, banned commands (curl/wget), GOEXPERIMENT needs — check the AGENTS gotchas FIRST.
7. **Heavy E2Es (F104-type) go into the todo list explicitly** — silently scoping one down mid-flight is how F104 got dropped.
8. **Ship loop discipline:** core now accumulates features in `[Unreleased]` — tag trains should follow feature waves within the same session when the API surface is stable.
9. **LSP distrust worked** — every stale diagnostic was settled by the CLI; keep the CLI-as-truth rule and stop re-litigating.
10. **The auto-commit daemon interleaves with explicit commits** — unavoidable, but staging explicit commits immediately after each unit keeps the history readable.

## f) NEXT — up to 50 things to get done (ordered by leverage)

**Ship loop / release correctness**
1. Re-check pkg.go.dev render for `docs@v0.3.0` + all module pages (crawler should have caught up) — close the P1 `[~]`.
2. Tag **core v0.5.0**: metrics surface + Version + testkit (API-break check first; it's a feature release).
3. Fresh-consumer proxy test for core v0.5.0 + re-pin `integration/` to it.
4. Fold `Doc snippets are code` verification for the NEW README sections (Metrics block, testkit) into a scratch-module compile check.
5. Add FEATURES.md rows for the core features (G2/F5/E1) — the security section got rows, core did not.
6. Add AGENTS "Core Module — Code Organization" rows for `metrics.go`, `version.go`, `testkit/` (swap lines to stay ≤377).
7. Close the stale integration-expansion TODO (3/5 done) and re-scope it to the two remainders.
8. Close the stale "Flightrecorder ops preset + metrics bridge" TODO item (work landed under polish).
9. `security/.golangci.yml`: replace realtime-inherited header comment with a security-specific one (cosmetic).
10. Run the fresh-repo `go.work` smoke: clone to /tmp, verify the security module builds outside the workspace (go.work is gitignored here).

**Upstream (user-gated — say the word)**
11. File Draft 1: go-sse dedup-aware `ReplayFiltered` (draft ready).
12. File Draft 2: httputil Logging request-context emit (draft ready).
13. Implement the httputil listener-injection API (`NewServerListener(ln, cfg, handler)`) upstream — unblocks the composition refactor AND the Core TLS option (G1) in one move.
14. Then re-run the composition spike against the new API and execute C20 for real.
15. Then implement Core TLS (`ServiceConfig.TLS{CertFile, KeyFile}`) — PapDashboard's first demand.

**Finish the partials**
16. Health godoc Examples with verified output (F107 — the frh-module pattern).
17. Health aggregate multi-probe example (F113).
18. govulncheck run on health + security (needs a networked machine for install).
19. `cqrs.OTelProjectionMetrics` E2E in integration (F104 — the silently dropped one).
20. Core composition-contract suite in integration (readiness composition, drain transitions, response parity).
21. fr MetricsHook end-to-end: drive a REAL capture through the bridge and assert the meter.
22. Real browser (chromedp or manual) pass over the health dashboard under a strict-CSP + nonce configuration — the CSP verdict is server-side proven only.
23. Measure `NewFlightRecorderMetricsHook` counters from an actual `Middleware` capture (not just a synthetic event).

**Batteries (demand-gated — re-confirm demand first)**
24. W3 B1 ResultHandler family (classification-parity with errorpages is the pin).
25. W3 B2 bind+validation.
26. W3 B9 no-leak error responses.
27. W3 B4 conditional GET (promotes the idle go-etag dep).
28. W5 C2 projection→broadcast folded contract (the must-have for cqrs+realtime consumers).
29. W5 C1 SSE drop/backpressure counters (absorbs the TELEMETRY.md §6 candidate).
30. W4 D4 atomic file write (floor: go-atomic-write ≥ v0.5.1).
31. W4 D7 idempotency store.
32. W1 F2 timing battery (after upstream Draft 2 lands, shared duration source).

**Hygiene / truth**
33. Sweep the four HTML reports never opened (2026-08-15/16, 2026-09-04 research) — annotate or archive.
34. Verify the 4 remaining unverified `infertypeargs` sites (`cqrs/commands_test.go:73,90,187,245`).
35. Re-run the 6 modules not linted in the last docs-health pass (root, cqrs, docs, otel, errorpages + security now exists) — sequential.
36. Root CHANGELOG: cut a dated section header for the 2026-09-16/17 wave when core tags.
37. Update the SUPERB plan file with execution verdicts inline (it is the predecessor contract; it still shows C-tasks as planned, not done).
38. `doc/TELEMETRY.md`: fix the §-numbering slip (the log-level decision pointer says §5, which is store-separation).
39. SECURITY.md-style threat-model page for the security module (per-battery threat → test mapping table).
40. Add `security` to the root README module table's "status" column note (opt-in, nothing in the default stack).
41. Dependabot: verify the generated /security entry actually matches the sibling format (visual diff against /realtime block).
42. Add a `security` example service (like errorpages/example) demonstrating the full hardened chain composition.
43. Write the integration test for `security` + `realtime` composition (rate-limit in front of SSE).
44. Extend `doc/status/README.md` index with the archived-report count and the gate command.
45. Move the annotation-depth standard ALSO into `doc/planning/archived/` README (F141 covered status only — the plan's F141 says "both archived READMEs"; only one got it — this is a real gap I introduced).
46. Retire the recovered pin-test doc (`doc/planning/archived/2026-09-16_otel-pattern-pin-test.md`) to the archived planning dir (landed + linked).
47. Consider `Retract docs/v0.2.0` in the docs module go.mod so `go get .../docs@latest` can never resolve the ghost tag on a stale proxy.
48. Add dependabot groups coverage check to CI (dependabot covers integration; CI now does too — assert parity in a workflow step).
49. Pareto plan v3 only AFTER the user answers the gates — do not self-start.
50. Sleep the watchlist: next refresh re-checks cordis consumers, PapDashboard v0.3.1+, nixpkgs toolchain.

## g) Questions I cannot answer myself

1. **Core v0.5.0 now or later?** The metrics surface + Version + testkit sit in core `[Unreleased]`. Ship the feature tag now (my recommendation: yes — features are tested and proxy-checkable), or accumulate more W1 first?
2. **May I file the two upstream asks?** Both target your own repos (go-sse, httputil), drafts are ready and source-cited — but filing was explicitly USER-GATED in the plan and I did not file. Approve?
3. **Should I implement the httputil listener-injection API upstream myself** (`NewServerListener`)? It's your repo, it unblocks both the composition refactor and the Core TLS option — but it's a public-API addition to httputil and I want your go/no-go before extending a published module's surface.

---

**Verification state at writing:** tree clean · 57 commits since plan · 10/10 module suites `-race` green · 0 golangci issues · structure linter 0 · AGENTS 377/377 · 7 tags on origin from this train.
