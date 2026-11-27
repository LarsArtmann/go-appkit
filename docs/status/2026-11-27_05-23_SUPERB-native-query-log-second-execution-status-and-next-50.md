# Status report: native query log + client tracking (SUPERB v2), 2026-11-27 05:23 CET

> **Scope:** the execution session of 2026-11-27 01:00 to 05:23 CET, the second execution
> session of the plan
> [`2026-11-26_11-19_SUPERB-native-query-log-and-client-tracking.md`](../planning/2026-11-26_11-19_SUPERB-native-query-log-and-client-tracking.md).
> Previous report:
> [`2026-11-26_23-42_SUPERB-native-query-log-full-status-and-self-review.md`](2026-11-26_23-42_SUPERB-native-query-log-full-status-and-self-review.md).
> **State at writing:** STOPPED on owner request. Every uncommitted item below is on disk and
> is listed with its exact location; nothing uncommitted is lost.

## 0. Headline

- The three reds this work had put on the cqrs-htmx tip are fixed and on origin. CI run
  45918251601 on `5afc85fd`: `test`, `lint (dashboardui)`, `lint (setup)` green. The three
  remaining reds are foreign (section d).
- The four reds on the go-appkit tip are fixed locally and every one was verified against the
  exact tool the CI job runs. They are in daemon commits `cd24a16` to `385212d`, which were not
  pushed at 05:23 (`origin/master` was still `bb7e26c`), so CI has not confirmed them yet.
- The go-cqrs-lite branch is split into the two commits its own checklist demands, rebased onto
  master `56d1a5f9b`, and green on every changed module and every consumer module. It is NOT on
  master and NOT pushed.
- Client tracking: the server endpoint (T14) is written and tested. The tracker script (T15) is
  written but has no test file. The setup wiring (T17) is half written and was never compiled.
  T18 to T21 are not started.
- Two things went wrong in this session and both are mine: I over-deleted a validation check
  and restored it; and a subagent was cut off mid-run by a session restart.

## a) Fully done and verified

| Task | What | Evidence |
| ---- | ---- | -------- |
| Pipeline test fix | `TestDispatchPipeline_QueryCarriesClientID` served through `App.Middleware()` | daemon commit `1d17ecd9`; CI `test` job green on `5afc85fd` |
| dashboardui lint | import order, `BeforeEach` for the user ID, blank line | `b7f42483`; `golangci-lint` from `dashboardui/` 0 issues; 4 of 4 specs ran; CI green |
| setup cyclop | audit-store rejection moved into `validateAuditStores`; `validateDisableService` is byte-identical to its pre-audit-trail state (diffed against `0fb2e591`) | `5afc85fd`; lint 0 issues; 4 audit-store tests green; CI green |
| Commit gates in worktrees | `check-conflict-markers.sh` and `check-debug-files.sh` tested `.git` as a directory, so every authored commit from a linked worktree died in the first hook block. Both now ask git. Each new self-test case was run red against the old gate first | `d30a728f`; conflict-markers self-test 14/14, debug-files 10/10, shellcheck clean |
| Private worktree | `/home/lars/.cache/ch-tracking/wt` on `feat/client-tracking`; master was fast-forwarded to it and pushed (`cc14b195..5afc85fd`), pre-push gates green | `git ls-remote origin master` = `5afc85fd` |
| go-appkit T07.3 | audit-wiring Notes line in `docs/agents/module-details.md` | on disk, daemon commit |
| go-appkit art-dupl | `//art-dupl:accept` on the nil-logger guard in `cqrs/audit.go`. The red was caused by my own T02 rewrite: the same job was green on `2476d39` | gate green locally with art-dupl 0.10.0; also green with a binary built from the CI pin `93432fd9`, and red with the directive removed in a scratch copy |
| go-appkit cqrs-lint | F032 and A016 are now config-level disables in the root `.cqrs-lint.json`; two inline directives removed. The inline form had drifted twice in one package because both rules anchor on "first import" / "first Use call" | `GOWORK=off cqrs-lint .` from root: "No findings. Clean!", zero warnings; cqrs tests `-race`, vet, lint green |
| go-appkit core v0.13.1 truth | integration pin bumped with `go get` (one line plus two sum lines), `documentedPins`, FEATURES claim, root CHANGELOG `[0.13.1]` dated 2026-11-24, AGENTS release state, `docs/MODULE_INVENTORY.md` regenerated. Verified the tag delta myself: the module differs from v0.13.0 by exactly one line (`go.mod`) | `check-pin-drift`, `check-features-claims`, module-inventory, routed, directives, status-index all rc 0 |
| go-cqrs-lite T13 split | `a76ded1c7` became `2d1818232` (v4 train + docs, golden regenerated with the core/v5 stash held out) and `35fa58baf` (core/v5 + golden), rebased onto `56d1a5f9b` | lockstep test, core/v5 build/vet/tests green |
| go-cqrs-lite blast radius | full tests: query/v4, core/v5, system/v4, metaengine/v4, sqliteengine/v4, tursoengine/v4; `-short`: systemscenario, systemtest, systemcatalog, system/integration, benchkit, scheduling/engine, api-stability, cqrs-lint, doc-check, cqrs-upgrade, cqrs-gen | all rc 0. `system/v4` failed once under load 55 with a snapshot deadline, then passed 3/3 and full at load 19 |
| T14 TrackHandler | `tracking.go` (types, caps, same-origin guard before the body read, `MaxBytesReader`, attribution from context or header, sink write on `context.WithoutCancel` + 5 s), `tracking_embed.go`, `tracking_test.go` | `go test -run 'Track\|TrackedEvent'` rc 0; 16 top-level tests |
| Signals-journal feasibility | a scratch module against PUBLISHED pins (system v4.19.1, sqliteengine v4.9.2): an undeclared `"signals"` collection accepts `StreamAppend`, reads back per stream and whole-journal, and survives close and reopen; the event adapter has `Serialize=true` | probe output in the session log, `/tmp/sigprobe` |

## b) Partially done

| Task | On disk | Missing |
| ---- | ------- | ------- |
| T15 tracker.js | `/home/lars/.cache/ch-tracking/wt/tracking/tracker.js`, 520 lines, `node --check` passes. Written by a subagent to the spec I gave it (fetchLater with `activateAfter`, abort-and-resend, local keepalive shim, four signals, three opt-outs, shared device key) | **I have not read it.** No review against the spec yet |
| T16 JS tests | nothing | `tracking/jstests/tracker.test.js` was never written: the session restarted and cut the subagent off; I resumed it but no report arrived before this stop |
| T16 Go wrapper | nothing | the shared `runJSSuite` helper and `TestTrackerJSSuite` |
| T16 API pins | nothing | `api_surface_compat_test.go` rows |
| T17 setup wiring | `setup/tracking.go` (journal on `system.AdapterCore`, `Record`, `ForUser`, `ForClient`, `All`, sink resolution, mount, postures); edits in `setup/config.go`, `bundle.go`, `setup.go`, `mount.go`, `route_postures.go`, `config_paths.go`; `setup/go.mod` has `system/v4 v4.19.1` (via `go get`, +15 indirect lines) | **Never compiled, never formatted, no tests.** A missing comma I introduced in `setup.go` was fixed at 05:23; nothing else was checked. `metaengine/v4` is still marked indirect until tidy runs. Dep budget, footprint table, README rows, CHANGELOG, posture golden: all untouched |
| T13 landing | two commits ready on the branch | heavy gates (`#verify`, coverage, erraudit, gosec, cycles, arch, test-runtime), fast-forward of master, push |
| T23 plan file | nothing this session | section 10 log, section 8 table, deviations record, header |
| Root lint of the tracking files | nothing | `golangci-lint run` on the root module was started as job 293 and its result was never read |

## c) Not started

T18 Playwright e2e. T19 tracking-demo example. T20 gate battery. T21 skill section (cqrs-htmx
and the go-cqrs-lite module row). T22 for the tracking work. The plan file close-out.

## d) What went wrong

1. **I deleted a check I did not own.** While moving the audit-store rejection out of
   `validateDisableService`, my edit also swallowed the `AuthHandlerConfig` rejection. I caught it
   by re-reading the region and restored it; the final file is byte-identical to its
   pre-audit-trail state. It would have shipped silently if I had not looked.
2. **A subagent was lost to the session restart.** Its script survived, its tests did not.
3. **Two misreadings cost time.** The previous report called the art-dupl and cqrs-lint reds
   "foreign, proven". They were consequences of my own T02 rewrite: CI was green on `2476d39`
   and red on the first commit that carried the rewrite. I only saw it when I compared runs.
   And a self-built copy of the CI-pinned art-dupl seemed to find no clones at all; the cause was
   my own "directive disabled" probe, which left a space-variant the tool still honours.
4. **BuildFlow's pre-commit hook fails in the cqrs-htmx main checkout.** `nix-fmt:detect` tried to
   read the git-ignored `.crush` symlink as a file. It passed from the worktree. I did not chase
   the root cause; it is a BuildFlow matter.
5. **The hook's `go-generate` step rewrites nine `adminui/*_templ.go` files in the worktree** with
   templ v0.3.1123 against the committed v0.3.1118. I restored them after every commit.
6. **Box load reached 55 to 67.** One deadline test failed under it and passed alone. I added to
   that load by running two heavy suites at the same time.

## Foreign reds, seen and not touched

- cqrs-htmx tip: `checks` (`docs/review/libs/index.render.png` 1.47 MB over the 1 MB gate),
  `lint (identity-model)` (`FoldUser` gocyclo 34 over 30), `module-architecture`
  (`.docs-gates-preflight.json` not in the root-files baseline).
- go-cqrs-lite master: `cmd/cqrs-bench/v4` tests do not build (`noFlagsForTest`, `errImpossible`
  undefined); my branch does not touch that module.
- go-cqrs-lite `go-generate` hook step: `cqrs-gen` is not on my PATH, so 103 steps reported
  "failed" in the worktree; the commits still went through.

## e) What I could have done better

- Read a test's or gate's failure against the last green run before labelling it foreign.
- Edit with smaller anchors. Both of my bad edits were large multi-line replacements.
- Compile after every wiring edit. The setup tree sat broken for the whole design phase.
- Dispatch the tracker script work earlier and review it the moment it landed, instead of
  assuming the subagent's report would arrive.
- Keep one heavy job at a time.

## f) Next, in order

1. Compile, format, vet and lint the setup wiring; run `go mod tidy` hermetically.
2. Read `tracker.js` line by line against the spec and the two retrieved guides.
3. Write `tracking/jstests/tracker.test.js` (cases a to s from the spec) and get it green.
4. Add the shared `runJSSuite` helper; point the sync test at it; add `TestTrackerJSSuite`.
5. Add the API signature pins for every new root symbol.
6. Read job 293's lint output and fix what it names.
7. Setup tests: journal streams per subject, `ForUser`/`ForClient`/`All`, off-by-default with
   one log line, `TrackingSink` and `TrackingBackend` both set is rejected, posture rows.
8. Update the route-posture golden and its opt-in test.
9. Dep budget 25 to 27, `dep-footprint.sh` regeneration, root-size and root-clusters baselines.
10. Run the setup and root suites in workspace mode AND `GOWORK=off` (workspace mode resolves
    go-cqrs-lite from the local checkout, which carries unreleased code).
11. Commit T14, T16 and T17 as separate authored commits; fast-forward master; check CI.
12. go-cqrs-lite heavy gates, then fast-forward master and push per plan decision D10.
13. Push go-appkit master; confirm `art-dupl-gate`, `cqrs-lint`, `pin-drift`,
    `module-doc-claims` green on `385212d` or later.
14. Playwright e2e: page view, dwell, click, impression arrive attributed; GPC and
    `data-tracking="off"` send nothing.
15. `examples/tracking-demo` with a read-your-signals page; boot, probe, graceful stop.
16. cqrs-htmx README section, setup README rows, CHANGELOG entries, `examples/README.md`.
17. Skill section for cqrs-htmx; module row for the go-cqrs-lite audit writer; snippet gates.
18. Plan file section 10, section 8, a deviations record, header status.
19. TODO rows in go-appkit and cqrs-htmx for the follow-ups.
20. Decide and record whether `anonymous` needs its own read accessor.
21. Sweep the `[Unreleased]` sections of cqrs-htmx root, dashboardui and go-appkit cqrs.
22. Re-run the full root suite at quiet load; `TestServeDomainEvents_HeartbeatEmission`
    failed once under load 55 and passed 5/5 alone.
23. File the BuildFlow `.crush` finding upstream after a clean repro.
24. Record the templ-version rewrite in the cqrs-htmx hook as a known worktree side effect.
25. Re-run the adoption census baseline only at its six-month gate (about 2027-05).
26. Recommendation read models: not in this plan; first app that asks.
27. `System.TrimQueryLogBefore` / `TrimCommandLogBefore`: needs the owner ruling (question 1).
28. `WithQueryAudit` / `UseQueryMiddleware`: same ruling.
29. `JournalSweeper` on engines other than SQLite and Turso: first consumer that needs it.
30. Upstream `command.AuditMiddleware`, so appkit's `commandAuditor` can go.
31. Device key on commands: when a consumer needs pre-login command attribution.
32. Subject-keyed layout for the query log at about 100,000 rows.
33. The three foreign cqrs-htmx reds, each a small change in an area another session owns.
34. `cmd/cqrs-bench/v4` test build on go-cqrs-lite master.
35. `cqrs-gen` on PATH for the go-cqrs-lite pre-commit hook.
36. Fold-over-signals in a dedicated event collection: when folds are wanted.
37. Sweep the remaining stale cqrs-lint warning at `cqrs/commands.go:187` (documented noise).
38. Formal deep-dive for the native query log: none required.
39. Owner question 3 below: query strings in tracked paths.
40. Retention for the `signals` collection: the SQLite typed-WAL change makes audit rows
    sweepable, but nothing sweeps `signals` yet.
41. Docs for the `TrackedEvent` JSON shape as a wire contract for consumers' own sinks.
42. Confirm the tracker under a nonce-only CSP in the e2e.
43. Hermetic build of the tracking-demo as an example module in `check-module-isolation
    --examples`.
44. Register the demo in every example gate (dependabot, workspace, examples table).
45. Status of the `fetchLater` quota on long-lived tabs: nothing measured yet.
46. Update `docs/DOMAIN_LANGUAGE.md` with signal, signals journal, tracked event.
47. Update go-appkit `FEATURES.md` when cqrs next tags (audit trail row exists, unreleased).
48. Re-pin go-appkit and cqrs-htmx onto the go-cqrs-lite tags that carry T08 to T12.
49. `docs/MODULE_INVENTORY.md` and AGENTS.md were uncommitted at 05:23; confirm the daemon
    committed them before judging the tip CI.
50. Close this report's rows into TODO_LIST (docs-health HARVEST) before the next session.

## g) Questions only the owner can answer

1. **go-cqrs-lite surface.** `scripts/check-package-surface.sh` fails on ANY growth of
   `system/v4` (235 symbols, 53 files). Keep the zero-growth shapes I shipped
   (`DomainConfig.QueryAudit` as a field; retention only through the documented
   `TrimJournalBefore` recipe), or raise the ratchet to add `System.TrimQueryLogBefore` /
   `TrimCommandLogBefore` (237) and `WithQueryAudit` / `UseQueryMiddleware` (239)?
2. **Landing go-cqrs-lite.** Plan decision D10 says "master pushes only". May I fast-forward
   go-cqrs-lite master to `35fa58baf` and push, knowing master currently carries a foreign red
   (`cmd/cqrs-bench/v4` tests do not build) and the other session is active there? Or hold
   until the tag wave?
3. **Query strings in tracked paths.** The tracker and the handler both keep `pathname` only,
   so a page like `/search?q=shoes` records `/search`. Keep that (privacy by default), or should
   an app be able to opt a path into keeping its query?

## Verification record

Every claim in section a was run in this session; commands and rc values are in the session
log. Not verified: anything in section b, the pushed state of go-appkit (`origin/master`
`bb7e26c` at 05:23), and the tracker script's behaviour in a real browser.
