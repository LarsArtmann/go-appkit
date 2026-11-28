# Status report: native query log + client tracking (SUPERB v2), 2026-11-27 11:53 CET

> **Scope:** the third execution session of the plan
> [`2026-11-26_11-19_SUPERB-native-query-log-and-client-tracking.md`](../planning/2026-11-26_11-19_SUPERB-native-query-log-and-client-tracking.md),
> 06:07 to 11:53 CET. Previous report:
> [`2026-11-27_05-23_SUPERB-native-query-log-second-execution-status-and-next-50.md`](2026-11-27_05-23_SUPERB-native-query-log-second-execution-status-and-next-50.md).
> **State at writing:** STOPPED on owner request. One authored commit exists on the private
> branch; the rest of the work is uncommitted in the private worktree and listed with its exact
> location below.

## 0. Headline

- Client tracking is built end to end in the private worktree: the endpoint, the embedded
  tracker, the setup signals journal, a Chromium e2e spec (3/3 green) and a runnable example
  (test green with `-race`, smoke boot/probe/stop green). One commit of about five is made.
- The root feature is committed as `ab7c7b01` on `feat/client-tracking` (11 files, +2831/−39,
  hook green). Root lint is 0 issues and the root `-race` suite is green on that tree.
- The setup wiring (T17), the e2e surfaces (T18) and the example (T19) are finished on disk and
  verified, but NOT yet committed, NOT rebased onto master, NOT on master, NOT pushed.
- go-appkit master `f0e0437` is on origin (the daemon pushed it); I pushed nothing. The CI
  result for that tip was not read this session.
- go-cqrs-lite is untouched this session: my branch `feat/query-audit-trail` is still at
  `35fa58baf` on `56d1a5f9b`; master has moved to `4f5545fab` and its main checkout is clean
  again (0 dirty entries).
- Three things are mine and wrong or open: the example's default port 8097 is occupied on this
  machine; the root commit message says 20 tests where the file has 19; and I did not re-run the
  setup lint and suite after the last three setup edits.
- One design gap I noticed while writing this report and have not verified in code: the signal
  envelope is dated `at`, not `occurred_at`, so the time-based journal sweep of T11 probably
  never sweeps signals (section e, item 1).

## a) Fully done and verified (this session)

| Task                   | What                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Evidence                                                                                                                                                      |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T17                    | Setup wiring compiles and is formatted (`gofmt`, `gofumpt`, `golines -m 120`): `setup/tracking_journal.go` (`SignalsJournal` on `system.AdapterCore[cqrshtmx.TrackedEvent]`, one append per subject, whole batch in one transaction where the engine has `Transactional`, `ForUser`/`ForClient`/`All`, nil-safe reads), `setup/tracking.go` (`attachTracking`, `resolveSignalsHome`, session plus enrichment, mount, posture rows, `Bundle.TrackingJournal()`), `setup/config_validate_tracking.go`                                      | `go vet ./setup/` clean                                                                                                                                       |
| T17                    | 16 tests in `setup/tracking_test.go`: journal (nil backend, nil journal, one append per subject in a transaction, empty batch, failing append keeps its cause, round trip through an engine that returns decoded maps), bundle (off with one log line, `TrackingPath "-"` silent, backend mounts endpoint and script, system event adapter is the default home, signed-in visitor attributed, sink wins and builds no journal, the endpoint keeps its own refusals, path options table, invalid configs table, the shell mounts nothing) | green, workspace mode                                                                                                                                         |
| T17                    | `setup/route_postures_test.go` opt-in test now expects `POST /-/track public` and `GET /-/tracker.js public`; the default golden is unchanged (a default bundle mounts no tracking)                                                                                                                                                                                                                                                                                                                                                      | green                                                                                                                                                         |
| T17                    | Hermetic `go mod tidy`: `system/v4 v4.19.1` and `metaengine/v4 v4.26.1` are direct requires; dep budget 25 → 27 in `scripts/checks/check-dep-budgets.sh`                                                                                                                                                                                                                                                                                                                                                                                 | `check-dep-budgets.sh` green (27/27)                                                                                                                          |
| T17                    | Full setup suite and setup lint                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | suite `ok` (12.6 s), lint `0 issues`, both at about 06:50 on the tree BEFORE the last three edits (the typed-nil fix, message wording, the posture test rows) |
| T15                    | Line-by-line review of `tracking/tracker.js` against the plan's T15.1 to T15.9 and the five cached guides; all six named deviations ruled (see below)                                                                                                                                                                                                                                                                                                                                                                                    | reviewed; 23/23 Node tests                                                                                                                                    |
| T16                    | `runJSSuite` helper in `js_suite_test.go`, `TestTrackerJSSuite` in `tracking_js_test.go`, `TestSyncWorkerJSSuite` now uses the helper; API signature pins for every new symbol appended to `api_surface_compat_test.go` (incl. the exact `TrackingSink` method set)                                                                                                                                                                                                                                                                      | green; Node suite 23/23 from Go                                                                                                                               |
| T14                    | `tracking_test.go` rewritten around three helpers (`recorded`, `recordedOne`, `requireReceipt`): 19 tests; `tracking.go` gained `noSignal()`; the unused `nolint:gosec` directive is gone                                                                                                                                                                                                                                                                                                                                                | root lint `0 issues`                                                                                                                                          |
| T20.1/20.2 (root only) | Root `go test ./... -count=1 -race`: every package `ok` (root 24.6 s); root `golangci-lint run`: `0 issues`                                                                                                                                                                                                                                                                                                                                                                                                                              | green                                                                                                                                                         |
| T20.3/20.4 (partial)   | `root-clusters.tsv` gains the `client tracking` cluster (2 files); root-size baseline 55 files / 8,924 LOC → 57 / 9,485                                                                                                                                                                                                                                                                                                                                                                                                                  | `check-root-size.sh`, `check-root-clusters.sh` green                                                                                                          |
| T18                    | `e2e/server/tracking.go` (tracker script, `POST /-/track` behind a real session, `/tracking`, `/tracking/off`, `POST /api/debug/tracking-login` idempotent, `GET /api/debug/signals`, `POST /api/debug/signals/reset`), `e2e/tests/tracker.spec.ts` (3 tests), `e2e/README.md` matrix row and test-server paragraph                                                                                                                                                                                                                      | Chromium 158 (Nix build): `3 passed (9.4s)`                                                                                                                   |
| T19                    | `examples/tracking-demo` (`main.go`, `handlers.go`, `main_test.go`, `README.md`, `go.mod`, `go.sum`); registered in `go.work`, `examples/README.md`, `README.md`, `AGENTS.md` (14 demos)                                                                                                                                                                                                                                                                                                                                                 | `go test . -count=1 -race` `ok`; tidy-stable in workspace and hermetic mode with `go 1.27` (appkit pinned at v0.13.1)                                         |
| T19.4                  | Smoke on a side port (`/tmp/tracking-demo-smoke`, `:18971`): `GET /` 200 with the tracker tag, `GET /-/tracker.js` 200 immutable, `GET /health` 200, `POST /-/track` 202 `{"accepted":2,"dropped":0}`, `GET /signals` lists path, device and click, `GET /signals?device=<key>` 200 names the subject, cross-site beacon 403, SIGTERM stops the process                                                                                                                                                                                  | all green                                                                                                                                                     |
| T22.1 (1 of 5)         | Commit `ab7c7b01 feat(tracking): collect client signals behind a first-party endpoint and an embedded tracker` on `feat/client-tracking` in `/home/lars/.cache/ch-tracking/wt`                                                                                                                                                                                                                                                                                                                                                           | BuildFlow hook green (110 s); no templ file left dirty                                                                                                        |

Rulings on the tracker's deviations from my spec (all accepted, none changed):

| Deviation                                                                                                                         | Ruling | Reason                                                            |
| --------------------------------------------------------------------------------------------------------------------------------- | ------ | ----------------------------------------------------------------- |
| 520 lines against a 300 to 450 target                                                                                             | accept | every line is behavior or its comment; no size gate applies to JS |
| `enable()` emits a page view                                                                                                      | accept | the view starts at consent, so the dwell counts only from there   |
| `disable()` ends the page view without an engagement signal                                                                       | accept | a withdrawn visitor's dwell is not sent                           |
| `enable()` lifts `<html data-tracking="off">`; only GPC is permanent                                                              | accept | the host page owns both the attribute and the call                |
| `pageshow` with `persisted` starts a new page view                                                                                | accept | back-forward cache restore is a new view                          |
| `pagehide` counts as hidden for the keepalive shim                                                                                | accept | a 0 ms timer never runs during unload, a microtask does           |
| impressions need an intersection ratio of at least 0.5, re-scan on every page view, `data-tracking="off"` ancestors suppress them | accept | the observer also reports the initial state of every element      |
| client-side byte clamps and the one-day `visibleMs` cap                                                                           | accept | mirrors the server, saves a refused field                         |
| engagement events carry the page view's referrer                                                                                  | accept | a self-contained row for per-referrer dwell queries               |

## b) Partially done

| Item                    | State                                                                         | What is missing                                                                                                                                                                                                                                                                                                 |
| ----------------------- | ----------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T22 commits and landing | 1 of about 5 commits made (`ab7c7b01`)                                        | commits for setup, e2e, example, docs; rebase onto master `abb70d54` (two docs-only daemon commits ahead); fast-forward; push; read the CI run                                                                                                                                                                  |
| T20 gate battery        | root `-race` and lint, setup suite and lint, root-size, clusters, dep budgets | setup `-race` and a setup lint re-run after the last edits; `check-root-files --strict` (red on a foreign file); module layers; re-export policy; `check-route-postures`; coverage thresholds; readme split-brain; skill claims and snippets; `nix run .#check-modules`; module isolation both legs             |
| T21 docs                | `e2e/README.md`, both example tables, `AGENTS.md` example count               | root README section; `setup/README.md` rows for `TrackingPath`, `TrackerScriptPath`, `TrackingSink`, `TrackingBackend`, routes, troubleshooting; CHANGELOG entries (root, setup); skill section (append, never retitle; load `how-to-write-skills` first); go-cqrs-lite skill module row; snippet compile check |
| go-appkit               | master `f0e0437` is on origin                                                 | the CI result for that tip was not read; my 05:25 addendum to the previous report and its index row are still uncommitted in the main checkout                                                                                                                                                                  |
| T13 go-cqrs-lite        | unchanged: branch `35fa58baf` on `56d1a5f9b`, clean tree                      | rebase onto master `4f5545fab`, the heavy gates, fast-forward, push (owner question 2)                                                                                                                                                                                                                          |
| Plan file               | unchanged since the previous report                                           | section 10 log, section 8 table, section 9 deviations, header status                                                                                                                                                                                                                                            |

## c) Not started

- T23: plan close-out (execution log, verification table, deviations record, header) and the
  TODO rows in go-appkit and cqrs-htmx.
- The final delivery report with its `### Modern Web Guidance` section.
- The Firefox advisory run of `tracker.spec.ts` (the shim path of the tracker is unproven in a
  real browser; only the Node stubs cover it).
- A full Playwright battery run (only `tracker.spec.ts` was run; the other specs were not
  re-run against the e2e server with its new routes).

## d) What went wrong

| # | What                                                                                                                                                                                                                                                | Cost                         | State                                                                                                     |
| - | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- | --------------------------------------------------------------------------------------------------------- |
| 1 | The turn was cut between about 06:50 and 09:28 and resumed by your repeated directive                                                                                                                                                               | about 2.5 hours of wall time | outside my control; nothing was lost, the worktree kept everything                                        |
| 2 | I returned `errorfamily.WrapOnce(...)` directly from `SignalsJournal.Record`; it returns a `*Error`, and a nil `*Error` inside an `error` return is not nil, so every successful write looked like a failure                                        | 1 minute                     | caught by my own test, fixed with an explicit nil check; the commit message of the fix carries the reason |
| 3 | Two long single-shot file writes came out corrupted: the first rewrite of `tracking_test.go` had a garbage tail, and the first `examples/tracking-demo/main.go` had a bogus `//go:embed` line and undefined symbols                                 | about 10 minutes             | both caught by `go vet` in the same step, rewritten in smaller pieces, never committed                    |
| 4 | Three wrong test assumptions: the login page is a site-root catch-all (an unmounted GET answers 200, not 404); the Playwright `request` fixture does not share cookies with the page context; a signed-in visitor's signal leaves the device stream | about 15 minutes             | all three fixed in place; the third is the designed one-signal-one-subject rule                           |
| 5 | The e2e sign-in endpoint registered the same user on every call and failed on a reused server                                                                                                                                                       | 5 minutes                    | fixed: register once, reuse the session                                                                   |
| 6 | The root commit message says "20 handler, wire and script tests"; the file has 19                                                                                                                                                                   | a 2 minute amend             | open; the commit is unpushed and is the branch tip                                                        |
| 7 | The example's default port 8097 is occupied on this machine (`127.0.0.1:8097` is listening at 11:53; not my process)                                                                                                                                | a one-line change            | open; the README and package doc promise that port                                                        |
| 8 | I did not re-run the setup suite and lint after the last three setup edits                                                                                                                                                                          | about 4 minutes              | open                                                                                                      |
| 9 | I used em dashes in authored code (15 in `tracker.js`, 4 in setup messages, 2 in `mount.go`)                                                                                                                                                        | 5 minutes                    | replaced with plain punctuation before the commit                                                         |

Nothing foreign was modified, nothing was force-moved, nothing was pushed by me, and no
`rm`, `git checkout` or `git reset --hard` was used.

## Foreign reds, seen and not touched

- `bash scripts/checks/check-root-files.sh --strict` fails on `.docs-gates-preflight.json`, a
  tracked root file that is not in the baseline (the earlier report's `module-architecture`
  red). The fix is one `--update-baseline` by whoever owns that file.
- `examples/setup-demo/main.go` is not `gofmt`-clean (committed state, not mine).

## e) What I could have done better, and what we should improve

**Process, mine**

1. Commit per module as soon as that module is green. The setup, e2e and example work sat
   uncommitted in a cache directory for five hours.
2. Write at most about 300 lines per tool call and run `go vet` after each. Both corrupted
   writes were long single-shot outputs.
3. Re-run the lint and the suite of a module after the last edit, before moving on.

**Product, found this session (ordered by value)**

1. **Signals are probably not time-sweepable as stored.** The journal envelope dates a signal
   as `at`; the T11 record says the typed-WAL extractor fills the retention column for envelopes
   that carry `occurred_at`. If that is right, the signals collection never shrinks by age, and
   the anonymous stream is one shared stream that only grows. Verify against
   `metaengine/sqliteengine/v4/typed_wal.go`, then either emit `occurred_at` in `encodeSignal`
   or document a manual trim.
2. **`Config.TrackingBackend` costs a type assertion.** `metaengine.Engine` does not name its
   stream-log methods, so a consumer holding an engine writes
   `engine.(metaengine.StreamLogBackend)` (the example needs a helper for it). Either add
   `setup.EngineBackend(engine) (metaengine.StreamLogBackend, error)` or document the one line.
3. **The go-error-family constructors return `*Error`.** Returning one from a function whose
   result type is `error` is a typed-nil trap (item d2). This deserves a lint rule or an AGENTS
   gotcha in every family repo.
4. **Hermetic legs will be red until a cqrs-htmx tag train.** `setup`, `e2e/server` and
   `examples/tracking-demo` consume root and setup API that is not published, so
   `check-module-isolation.sh` and its `--examples` leg fail by construction (the documented
   root-tag-gap class). I did not run them; this is a prediction. The repo's practice is to
   accept the red until the train; the alternative is to cut the train with this push.
5. **The boot line "client tracking is off" is Info for every default bundle**, so every app
   without a signals home prints it at start and every test that builds a default bundle on the
   default logger prints it too. Consider Debug for the default-memory case.
6. **The root package grew by 561 lines** while ADR-0062 burns it down. Tracking is a clean
   candidate for its own subpackage later: it depends on `httperr`, `internal/hx` and the
   context helpers only.
7. **`TrackerScriptTag` returns an HTML string**, which `html/template` consumers must wrap in
   `template.HTML`; a templ consumer needs the safe-class instead. Document the two forms or
   ship a second helper.
8. **The device key is duplicated** between `sync/sync-client.js` and `tracker.js` on purpose
   (two no-build scripts). A Node test that reads both files and asserts the storage key and the
   ULID pattern are identical would turn the "must stay in sync" comment into a gate.
9. **Playwright cookie sharing is tribal knowledge.** `context.request` shares the context's
   cookie jar, the `request` fixture does not; `loginpage-posture.spec.ts` does it right but no
   README says so. One paragraph in `e2e/README.md` saves the next session 15 minutes.
10. **The example ports collide with local services.** 8080 and 8099 are documented as occupied
    here; 8097 is too. The examples README could say "pass a different addr" for every port
    instead of listing three.

## f) Next, in order

1. Change the example's default port from 8097 to a free one (`main.go` const, package doc,
   README, both example tables).
2. Amend `ab7c7b01`'s message (19 tests, not 20). It is the unpushed branch tip.
3. Re-run the setup suite and setup lint after the last edits; fix anything they find.
4. Commit T17 in the worktree: `setup/*`, `scripts/checks/check-dep-budgets.sh`; restore the
   templ files (`/tmp/ch-restore-templ.sh`).
5. Commit T18: `e2e/server/tracking.go`, `e2e/server/main.go`, `e2e/tests/tracker.spec.ts`,
   `e2e/README.md`.
6. Commit T19: `examples/tracking-demo/*`, `go.work`, `examples/README.md`, `README.md`,
   `AGENTS.md`.
7. Verify e1 (the `occurred_at` sweep) against `typed_wal.go`; fix or document.
8. Write the docs of T21: root README section on tracking; `setup/README.md` config rows,
   routes, troubleshooting; CHANGELOG entries for root and setup; `AGENTS.md` bullets for root
   and setup.
9. Skill section: load `how-to-write-skills`, append a client-tracking section to
   `.agents/skills/cqrs-htmx/SKILL.md` (never retitle), add the go-cqrs-lite module row for the
   audit writer.
10. Compile every new README and skill snippet in a scratch module (T21.3).
11. Run the T20 gate battery one script at a time and record each result: `check-route-postures`,
    `check-module-layers`, `check-reexport-policy -- --strict`, `check-readme-splitbrain`,
    `check-skill-claims`, `check-skill-snippets`, `check-coverage-threshold`,
    `check-shellcheck`, `nix run .#check-modules`.
12. Run `check-module-isolation.sh` and `--examples` once to measure the predicted reds.
13. Run the full Playwright battery once (chromium) to prove the new e2e routes break no other
    spec.
14. Run the Firefox advisory: `nix run .#e2e-firefox -- --project=firefox tracker.spec.ts`;
    record the result in the plan.
15. Decide e2 (`EngineBackend` helper or doc line) and e5 (boot line level); implement.
16. Add the Node parity test for the device key (e8) and the e2e README paragraph (e9).
17. Rebase `feat/client-tracking` onto master `abb70d54`; re-run the root and setup suites;
    fast-forward master; push; read the CI run of the pushed tip.
18. Commit the docs (T21) in the same branch before step 17, or as a follow-up commit after it.
19. go-appkit: read the CI run for tip `f0e0437` (`art-dupl-gate`, `cqrs-lint`, `pin-drift`,
    `module-doc-claims`).
20. go-appkit: commit this report, the 05:25 addendum and the index rows.
21. T13: in `/home/lars/.cache/gcl-audit/wt`, `git rebase master` onto `4f5545fab`.
22. T13: run the heavy gates: `nix run .#verify-fast`, `#verify` (exclusive), `#check-coverage`,
    `#erraudit`, `#check-gosec`, `#check-cycles`, `#check-arch`, `#check-test-runtime`.
23. T13: fast-forward the main checkout (clean now) and push (owner question 2); delete the
    backup branch `backup/query-audit-trail-a76ded1c7`.
24. After owner question 1: cut or wait for the tag train (root, then setup), bump the example
    pins, and re-run the isolation legs.
25. T23: plan section 10 rows for T01 to T23, section 8 verification table, section 9 deviations
    (T10 field, T11 no wire change, T12 deferred, T17 value-typed adapter core and
    `NewSignalsJournal` returning an error, the shell mounts no tracking, the T15 rulings above,
    the commit split of T14 to T16), header status line.
26. TODO rows: go-appkit (preset pointer to tracking; the `occurred_at` sweep if deferred) and
    cqrs-htmx (tag train for the isolation legs; the root-package extraction candidate).
27. Final report ending with `### Modern Web Guidance`.
28. After owner question 3: keep the shell as is, or add tracking to it behind an identity seam.
29. Open an issue or add a gotcha for the typed-nil trap of `*errorfamily.Error` (e3).
30. Consider a `docs/guides/client-tracking.md` once the feature is tagged (ROADMAP row now).

## g) Questions only the owner can answer

1. **Tags.** `setup`, `e2e/server` and `examples/tracking-demo` use root and setup API that no
   tag carries, so both module-isolation legs of the CI `checks` job will be red until a
   cqrs-htmx tag train (root, then setup). Do you cut that train with this push, or do I (and
   at which versions), or do you accept the red until the next train as the repo has before?
2. **go-cqrs-lite push.** The main checkout is clean now and master is at `4f5545fab`. May I
   rebase `feat/query-audit-trail` onto it, fast-forward master and push, or does the other
   session still own that window?
3. **The identity-external shell.** `DisableService` mounts no tracking and rejects
   `TrackingSink` and `TrackingBackend` with the do-it-yourself recipe, because ADR-0054 keeps
   the shell free of the session stack and I did not want to change its contract unasked. Keep
   that, or should the shell mount tracking behind an identity seam of its own?

Decisions I made on the three questions of the previous report, all reversible: the zero-growth
`system/v4` shapes stay (the package-surface ratchet is not raised); tracked paths stay
`pathname` only (no per-path opt-in for the query string); the go-cqrs-lite push is still open
(question 2).

## Verification record

| Time (CET) | Command                                                                                                | Result                                                                                              |
| ---------- | ------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| 06:21      | `gofmt -l setup/`, `GOEXPERIMENT=jsonv2 go vet ./setup/`                                               | 5 type errors in the never-compiled wiring (value vs pointer adapter core, `json.Unmarshal` reader) |
| 06:23      | `GOEXPERIMENT=jsonv2 go test ./setup/ -run 'Tracking\|SignalsJournal\|NewShell'`                       | 5 tests red (login catch-all, logger noise, shell validation order)                                 |
| 06:24      | the same                                                                                               | green                                                                                               |
| 06:24      | `GOEXPERIMENT=jsonv2 go test ./setup/ -count=1` (full, workspace mode)                                 | `ok` 15.6 s                                                                                         |
| 06:4x      | `golangci-lint run` from `setup/`                                                                      | 6 issues (forcetypeassert, funlen, gochecknoglobals, golines ×2, wrapcheck)                         |
| 06:5x      | the same, after the fixes                                                                              | `0 issues`                                                                                          |
| 06:5x      | `go test ./setup/ -count=1` (full)                                                                     | `ok` 12.6 s                                                                                         |
| 09:3x      | `go test . -run 'Track\|TrackedEvent\|JSSuite'`                                                        | red once (a table row lost its closing brace in my rewrite), then `ok` 20.8 s                       |
| 09:4x      | `golangci-lint run` from the worktree root                                                             | 14 issues at first, 4 after the rewrite, then `0 issues`                                            |
| 09:4x      | `go test ./... -count=1 -race` from the worktree root                                                  | every package `ok` (root 24.6 s)                                                                    |
| 09:4x      | `check-dep-budgets.sh`, `check-root-size.sh`, `check-root-clusters.sh`, `check-root-files.sh --strict` | green, green (57 / 9,485), green, red on the foreign `.docs-gates-preflight.json`                   |
| 09:39      | `playwright test tracker.spec.ts --project=chromium`                                                   | 2 passed, 1 failed (no user key on the beacon)                                                      |
| 09:43      | a temporary debug spec against a manual server on `:18924`                                             | `cookies after login: []`: the `request` fixture does not share the jar                             |
| 09:45      | `tracker.spec.ts` after switching to `context.request`                                                 | 2 passed, 1 failed (my lookup compared objects from two JSON parses)                                |
| 09:46      | `tracker.spec.ts`                                                                                      | `3 passed (9.4s)`                                                                                   |
| 11:3x      | `go test . -count=1 -race` in `examples/tracking-demo`                                                 | `ok` 1.1 s                                                                                          |
| 11:4x      | `/tmp/t19-smoke.mjs` against `:18971`                                                                  | every probe green; SIGTERM stops the process                                                        |
| 11:4x      | `git commit` in the worktree                                                                           | `ab7c7b01`, BuildFlow hook green (110 s)                                                            |
| 11:53      | `node --test ./tracking/jstests/tracker.test.js`                                                       | 23 tests, 23 pass                                                                                   |

Not run this session: setup `-race`; setup lint after the last edits; the T20 battery beyond
the rows in section a; module isolation; the Firefox and full Playwright batteries; every
go-cqrs-lite gate; the CI runs of both pushed tips.

### Modern Web Guidance

I checked the five cached guides (`full-session-analytics`, `batch-analytics-events`,
`calculate-total-foreground-time`, `detect-initial-visibility-state`, `privacy`) against
`tracker.js` line by line and changed no code because of them this session; the delivery
report will carry the per-guide credits for the changes they did shape.
