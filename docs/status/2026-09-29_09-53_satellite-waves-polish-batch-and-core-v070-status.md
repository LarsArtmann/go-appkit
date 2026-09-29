# Status — cqrs v0.6.0 train + pre-tag gate + polish batch + four satellite waves + core v0.7.0 (session 2026-09-29, ~06:45–09:55 CEST)

**Predecessor:** `docs/status/2026-09-29_06-40_unification-release-trains-and-buildflow-hardening-status.md` (its §f list is the queue this session worked; T18 was the resume point at ~60%).

**End state at report time:** tree clean (daemon-swept), through the **four satellite waves** (realtime v0.1.2 / flightrecorder v0.1.1 / docs v0.3.1 / health v0.1.4) + **cqrs v0.6.0** — all five tagged, gate-green, pushed, proxy-PASS, integration re-pinned, same-train docs landed. **Core v0.7.0 is mid-train**: changelog written, API diffs (both packages) additions-only, hermetic suite + lint GREEN — tag/push/proxy NOT yet done. The pre-tag gate from the frh v0.1.4 post-mortem now exists, is negative-tested, and guarded every tag this session.

---

## a) FULLY DONE

| Task                                                                     | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`scripts/pre-tag-checks.sh` (report #2, the post-mortem fix)**         | Re-runs all three guards + asserts the TAGGED artifact's go.mod (floor-exact directive, no toolchain line, no filesystem replace). Positive test (v0.6.0 green) AND negative test: **rejects `flightrecorderhealth/v0.1.4` by design** — the exact historical failure. Wired into AGENTS Release Ritual step 4 (cap managed by merging old steps 2-4; structure lint green).                                                                                                                                                                    |
| **T18 — cqrs v0.6.0 (finished the 60%)**                                 | README:49 removal note; CHANGELOG `### Breaking` + `[0.6.0]` dating (folded the stale never-shipped v0.10.1 entry into the sweep's true net delta: error-family v0.10.0→v0.11.0); API diff vs `cqrs/v0.5.0` = removal-only (`SQLitePath`) + 1 doc comment — v0.6.0 correct per 0.x convention; hermetic `test -race` + golangci-lint 0 + cqrs-lint exit-0 (same single pre-existing C023 warning as the old tag, verified side-by-side); tagged → gate green → pushed → **proxy PASS**; AGENTS bullet + storage-posture line + TODO row/header. |
| **T17 — cqrs lifecycle E2E** (`integration/cqrs_lifecycle_test.go`)      | Against PUBLISHED cqrs v0.6.0 + core v0.6.0: `NewEventService(memory)` → Register → StartProjections → `ServiceConfig{ReadyCheck, ShutdownHooks: es.Shutdown}` → HTTP dispatch/query through `testkit.Serve`. Proves: ReadyCheck→/health/ready 200 after StartProjections; command-over-HTTP → query-over-HTTP roundtrip; **in-flight drain** — the ONLY gate-releaser is the DrainHook, so the command provably straddles the shutdown boundary and its HTTP response still lands 204. 5x race-clean.                                          |
| **T19 — otel benchmark re-run (n=5, 1.27.1 floor)**                      | B/op + allocs/op IDENTICAL to the 2026-09-16 table (90 allocs everywhere) — **no allocation regression**. ns/op re-baseline DEFERRED: box under external nix-build storm (load avg 152, ~2.3x wall-clock poisoning); verdict + do-not-bench-under-load note written into otel README instead of a garbage table.                                                                                                                                                                                                                                |
| **T19b — live example E2Es (Go prober, curl banned)**                    | Root example + health example: serve → SIGTERM → **/health/ready 503 while / still serves 200** → connection refused → clean `graceful shutdown complete result=ok` phase logs. Health example: /readyz 503 `shutting_down:true`, /healthz stays 200 (liveness — correct), /health/live 404 (RegisterHealth opt-out — correct). Root example made **PORT-aware** (8080 squatted on this box; default unchanged).                                                                                                                                |
| **T24a — typed `seriesKey`**                                             | metrics.go: struct key (method/route/status) kills the `                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| **T24b — `appkit.Hook`**                                                 | Type ALIAS (not definition) with documented rationale: a definition would make `[]Hook` non-assignable from every consumer's `[]func(ctx) error` literal; alias = readability with zero migration. Verified: existing literals in tests/integration compile unchanged.                                                                                                                                                                                                                                                                          |
| **T25a — errors.Join shape test**                                        | Multi-hook failures: newline-separated (one line per hook), each sentinel reachable via `errors.Is` through the Infrastructure wrapper.                                                                                                                                                                                                                                                                                                                                                                                                         |
| **T25b — hardened-dashboard CSP asserts**                                | **Caught a real doc bug**: `security.BuildCSP` ships directives ALPHABETICALLY sorted, but csp.go's comment claimed a different fixed order — comment corrected (shipped-order documented), E2E now pins `frame-ancestors 'none'` + the 9-directive alphabetical order. security + integration suites green.                                                                                                                                                                                                                                    |
| **T25c — SSE×health drain E2E** (`integration/sse_health_drain_test.go`) | Pins the documented choreography: drain begins → hub.Shutdown closes the stream (frees http.Server.Shutdown from the non-hijacked SSE conn) → mounted.Drain flips /readyz 503 → clean EOF on the stream → Shutdown returns nil. 8x green. **Found + fixed a real race**: the header flush precedes hub subscription, so an early Broadcast intermittently loses events — test now syncs on `hub.SubscriberCount()`. Also: `MustReadNEvents` must read N events in ONE call (per-call scanner over-read drops buffered events).                  |
| **T26a — `testkit.DrainWindowProbe`**                                    | Helper (base as `func() string` — late capture, since `Addr()` is nil inside drain hooks) + its own testkit test (ready 503 + ping 200 during drain).                                                                                                                                                                                                                                                                                                                                                                                           |
| **T26b — health example `-hardened`**                                    | DashboardHardenedPreset + strict `security.BuildCSP` middleware (per-request nonce); security is an example-only health dep; README gotcha updated; **live CSP header verified** (alphabetical, frame-ancestors 'none', per-request nonce).                                                                                                                                                                                                                                                                                                     |
| **Four satellite trains**                                                | realtime v0.1.2 (go-sse v0.6.1) / flightrecorder v0.1.1 (httputil v1.2.0; API diff = godoc-example comments only) / docs v0.3.1 (templ-components v1.19.4) / health v0.1.4 (go-health v0.4.1 + dashboard v0.10.1 sweep + `-hardened`; API diff = doc-comment only). Each: changelog dated, hermetic suite+lint green, API diff, annotated tag, **pre-tag gate green**, pushed, **all four proxy-PASS**.                                                                                                                                         |
| **Integration re-pin**                                                   | core stays v0.6.0, cqrs v0.6.0, health v0.1.4, realtime v0.1.2; **T14 go-health leg realigned v0.2.0→v0.4.1** (the "v0.3.0 evaluation pending" lock premise is obsolete — the health module carries v0.4.1 since its sweep; suite green on it, comment rewritten). Suite + lint 0 + pin-drift 14 OK.                                                                                                                                                                                                                                            |
| **Same-train docs**                                                      | AGENTS: 4 module bullets, release state, 3 stale dep lines (go-sse v0.6.1; frh error-family v0.11.0 + go-health-lock note; health v0.4.1/v0.10.1 + security-example-only). TODO header: unreleased-deltas list corrected (all sweep deltas shipped; core's new deltas named). Pin-drift + structure lint green.                                                                                                                                                                                                                                 |

## b) PARTIALLY DONE

1. **Core v0.7.0 (~70%)**: `[Unreleased]` written (Hook alias, DrainWindowProbe, seriesKey, example PORT, Join-shape test); API diff vs v0.6.0 additions-only in BOTH packages (root: `type Hook`; testkit: `DrainWindowProbe`); hermetic `test -race` green, golangci-lint **0 issues** (fixed the `http.Client` exhaustruct finding in DrainWindowProbe). NOT done: tag `v0.7.0` → gate → push → proxy check → integration pin bump + DrainWindowProbe/Hook migration → AGENTS core bullet + release state ("through core v0.7.0") + TODO header line.
2. **T26a migrations (the "migrate 2 hand-rolled tests" half)**: helper exists but root `drainhooks_test.go` CANNOT import testkit (import cycle: `package appkit` test → testkit → appkit); my migration attempt was reverted. Routes: external-test-package surgery, or keep hand-rolled + pointer comment. Integration's `TestDrainWindowContract` migration becomes possible the moment core v0.7.0 ships.
3. **Realtime README note** for the discovered flush-precedes-subscribe broadcast window (producers: sync on `SubscriberCount()` in tests or accept at-most-once) — identified, not written.
4. **cqrs-lint scorecard re-run** (adoption ritual after the SQLitePath/DLQ changes) — lint ran clean but the scorecard delta was not re-recorded.

## c) NOT STARTED (from the predecessor's queue, in its impact order)

BuildFlow full-mode validation (#16), T29 watchlist (#17), T30 consumer ritual (#18), T32 rituals re-arm (#19), T20 AGENTS slim batches (#20-21), T21 art-dupl audit (#22), T31 lessons.md in crush-config (#23), upstream drafts: BuildFlow go-licenses ≥1.27 runner + go-structure-linter per-tool excludes (#24-25), tail items #35-38 (exhaustruct_v5 exclude-key warning triage — note the ANALOGOUS finding appeared in-repo today and was nolint-fixed, the BuildFlow config-key one is separate —, 5 go-auto-upgrade findings, jscpd policy, 5 lychee doc links), #40 security example appkit bump, #41 CI directives-job watch, #42 DOMAIN_LANGUAGE DLQ terms, #43 /tmp cleanup (core-old/cqrs-old/per-module -old/probe binaries/cspcheck), #44 AGENTS integration-bullet nuance, #45 dependabot re-verify, #46 doc floor-claim sync, #49 pkg.go.dev render check for the 9 new tags, #50 SUPERB v5 doc retrospective marks. **Still gated:** T13 (G3), T15 (upstream filings), T16 (errorpages swap), T22 (pdg), T23 (prompt-crusher), T27 (govulncheck), T28 (browser CSP), T10.1 (owner SSH secret), battery waves.

## d) TOTALLY FUCKED UP (honest)

1. **`flag.Parse()` missing in the `-hardened` example** — the mode silently never activated; my probe suite stayed green because it asserted STATUS CODES, and 200s are 200s with or without a CSP middleware. Caught only because I went one step further and printed the actual header. Lesson (now a personal rule): verify the DISTINGUISHING output, not a superset-compatible side effect.
2. **Invented a `cqrs-lint run` subcommand** — `run` is not a subcommand; it was parsed as a package path, producing a phantom `cqrs/run: no such directory` load error. Cost a debug cycle; the correct invocation is bare `cqrs-lint` (verified against `--help` only after guessing wrong).
3. **SSE E2E flake, two stacked causes**: (a) two sequential `MustReadNEvents(body, 1)` calls — the per-call scanner over-reads and drops the buffered second event; (b) the REAL race — `Broadcast` before the handler subscribed (~20% flake, 15s-timeout failures each time). Fixed with one batched read + the `SubscriberCount()` sync point. Should have designed the sync point FIRST for an async composition test.
4. **testkit helper churn + wasted migration**: wrote `DrainWindowProbe(base string, …)`, then realized late capture needs `func() string`; then attempted the root-test migration WITHOUT checking package topology — import cycle (appkit ↔ testkit in tests) killed it outright. Check import graphs BEFORE building in-repo consumers.
5. **Trusted the predecessor's unreleased-deltas inventory**: health's go.mod carried go-health v0.4.1 + dashboard v0.10.1 with NO changelog entry — the previous report's delta list (realtime/flightrecorder/docs/cqrs) was incomplete. Discovered mid-train (health was about to be left out of the wave); fixed by grepping go.mods vs tags. Future handoffs: enumerate unreleased deltas mechanically, not from memory.
6. **Edit-tool friction**: three view-first retries after the auto-commit daemon reformatted files mid-edit; one multiedit left an over-indented map brace (caught by gci lint); one nolint comment kept a line over golines' limit (moved above the statement). All caught by the gates — the gates earned their keep this session.
7. **T19's letter vs spirit**: the task said "re-run vs README table" — I ran it, proved no allocation regression, and honestly deferred the ns/op column (load 152). Right call, but the table is unchanged; the re-baseline is still open work.

## e) WHAT WE SHOULD IMPROVE

1. **Verify distinguishing behavior, not status codes** — the flag.Parse class dies only when assertions check the thing the feature changes (headers, bodies, metrics), not just "it served".
2. **Topology before tooling**: check import cycles / internal-vs-external test packages before placing helpers intended for in-repo consumers.
3. **Sync async composition tests on observability points** (SubscriberCount, hook-fired channels) — sleeps and "it probably subscribed by now" are the flake factory.
4. **Mechanical unreleased-delta enumeration for handoffs**: `for m in */; do diff <(git show <tag>:m/go.mod) m/go.mod; done` beats memory.
5. **Bench hygiene ritual**: check load average before any benchmark; never re-baseline ns/op under load (now noted in otel README; generalize to AGENTS if it recurs).
6. **Multi-module trains + the pin-drift gate**: per-tag gating requires bumping integration pins BETWEEN tags (done this session, works, but it's ritual friction) — consider documenting the tag→pin dance as an explicit Release Ritual sub-step.
7. **The pre-tag gate paid off immediately** — it guarded 5 tags this session with zero incidents; keep it mandatory (it is now, in the ritual).

## f) NEXT — up to 50 (impact-ordered)

1. **Finish core v0.7.0**: tag → gate (bump integration's core pin to v0.7.0 first for pin-drift) → push → proxy check.
2. **Integration migration**: `TestDrainWindowContract` → `DrainWindowProbe`; switch new E2Es' hook slices to `[]appkit.Hook`; suite + lint + guards + push.
3. Same-train docs for core v0.7.0 (AGENTS core bullet + release state through v0.7.0 + TODO header unreleased line emptied).
4. Realtime README gotcha: flush-precedes-subscribe window + `SubscriberCount()` as the producer sync point.
5. Root `drainhooks_test.go` route decision: external test package (import cycle) vs pointer-comment to the helper — pick one, execute.
6. cqrs-lint scorecard re-run + CHANGELOG delta (adoption ritual).
7. **BuildFlow full-mode validation** (`env -u GOTOOLCHAIN buildflow --build-mode full`).
8. T29 watchlist sweep (cordis, PapDashboard, nixpkgs toolchain, dprint exit-14, cqrs-htmx v5 window) with as-of dates.
9. T30 external-consumer proxy ritual decision + recipe paragraph.
10. T32 rituals re-arm (cqrs cookbook re-verify trigger, core v1 exit-criteria graduation, battery demand re-check).
11. T20 AGENTS slim batch 1 (per-module build commands → module READMEs; cap headroom).
12. T20b/c gotcha-group slim batches 2-3.
13. T21 art-dupl suppression audit (~431 groups) + standing policy + Decision 13 cross-link.
14. T31 lessons.md entry (crush-config repo, committed): verify-distinguishing-output / topology-before-tooling / sync-on-observability / mechanical delta enumeration.
15. Upstream draft: BuildFlow go-licenses ≥1.27 runner.
16. Upstream draft: go-structure-linter per-tool excludes / yaml config honor.
17. Triage BuildFlow's `golangci-lint-config-verify` warning (`exhaustruct_v5.exclude` key not allowed in one module's config — find which, fix key or version).
18. Triage the 5 go-auto-upgrade findings (lo.FromPtr frh; lo.Map/lo.Filter otel/realtime tests) — fix or policy-skip with rationale.
19. jscpd policy decision (13 findings; cross-module `.golangci.yml` 277-line clones are by-design).
20. Fix the 5 broken doc links (lychee: go-plugin-mvp, forks/cordis moved).
21. Check whether the go.work etag replace can drop NOW (realtime/flightrecorder/docs/security/otel graphs re-resolved via the new tags) — if yes, drop + fresh-worktree guard test.
22. SECURITY go.mod: example-only appkit dep v0.5.1 → current (next security-touching train).
23. Verify CI directives job parses the committed go.work on the next push (guards + proxy-smoke visible even with the matrix owner-blocked).
24. DOMAIN_LANGUAGE.md: DLQ poison contract terms (Rejection/Corruption = quarantined; retryable = restarts worker).
25. Clean /tmp scratch (core-old, cqrs-old, per-module -old, appkit-probe, example binaries, cspcheck).
26. AGENTS integration bullet: one clause noting the T14 version-locked legs are hand-realigned to family pins (go-health now v0.4.1).
27. Re-verify dependabot grouped updates reconcile after the 4 go.mod sweeps.
28. Grep-and-sync living docs for the old "through core v0.5.1" floor claim (README/FEATURES/ROADMAP).
29. pkg.go.dev render check for the 9 new tags (crawler lag — not a failure if 404 today).
30. Mark SUPERB v5 plan doc executed with deltas (T01-T12, T14, T17-T19, T24-T26 + this session's trains).
31. otel ns/op re-baseline when load < 2 (n=10, mean-of-10 convention, update README table).
32. Core v0.7.0 fresh-consumer BEHAVIORAL check (full recipe path, not just proxycheck) — the Hook alias/DrainWindowProbe consumer view.
33. Consider a tiny `testkit` follow-up: a two-path drain observation helper only if a second consumer appears (current two-hook pattern stands).
34. art-dupl pass over the session's new test files (two new E2Es share probe-helper shapes — audit before it grows).
35. Document the tag→pin dance in the Release Ritual (step 4 addendum) if multi-module trains stay the norm.
36. Health module: confirm no NEW deltas accumulated post-v0.1.4 (daemon sweeps) before closing the day.
37. Go back through the realtime gotcha list for the same race class documented anywhere else (handler.go comments).
38. Verify the integration module's `documentedGoDirective` fixture + `check-go-directives.sh` still agree after all trains (they should — ran green).
39. Check whether cqrs-htmx picked up core v0.6.0 yet (their etag stub-replace drop — D2 ritual freshness).
40. Consider making `pre-tag-checks.sh` accept multiple tags for wave pushes (gate-all-then-push-all).
41. T13 (G3-gated, USER).
42. T15 (USER-gated): 6 upstream filings.
43. T16 (USER-gated): errorpages statusRecorder → httputil.ResponseRecorder swap.
44. T22 (their repo): pdg backlog.
45. T23 (user intent): prompt-crusher merge resolution.
46. T27: govulncheck on health + security (networked machine).
47. T28: browser CSP pass over DashboardHardenedPreset (chromedp) + THREAT_MODEL link.
48. T10.1: owner sets `SSH_PRIVATE_KEY` → then watch a fully green master run.
49. Battery waves (demand-gated).
50. Write the next session's handoff with the MECHANICAL unreleased-delta enumeration (see e4).

## g) Questions only you can answer

1. **Machine toolchain (carry-over, now 2 sessions old)**: will the nix go default move to ≥1.27.1? Yes → I retire the whole GOTOOLCHAIN-prefix discipline (AGENTS blocks, .buildflow.yml env, hook notes) in one train. No → the prefix stays; I stop asking and keep documenting it.
2. **Core release cadence**: are addition-only core trains (v0.7.0 today; Hook/DrainWindowProbe/seriesKey) welcome per-session, or should additions batch into feature waves so consumers see fewer versions? Today's v0.7.0 unblocks the integration migration immediately — but I can hold additions for waves if you prefer.
3. **G3 (carry-over)**: consumer-pin claims in AGENTS/FEATURES — keep them and I build `scripts/check-consumer-claims.sh` (re-verifies claims after each cqrs-htmx release), or drop exact consumer pins from AGENTS entirely?
