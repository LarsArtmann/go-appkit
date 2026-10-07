# SUPERB Execution — Mid-Flight Status (T01–T10 done, T11 in flight, WAITING)

> 2026-10-07 11:00 CEST · point-in-time snapshot; release facts live in
> AGENTS _Release State_, task state in TODO_LIST, plan in
> `docs/planning/2026-10-07_10-11_SUPERB-pareto-execution-plan.md`.
> Mode: full execution was invoked ("GET SHIT DONE"); this report is the
> pause the owner requested. **The tree is GREEN except one deliberately
> flagged mid-flight test (T11) — see (c)/(d).**

## a) FULLY DONE (verified, committed, pushed)

| Task                           | What landed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Evidence                                                                                      |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| T01 (observable half)          | ci-preflight observed designed-red (secret still missing — owner-only); the unexpected workspace-charter red on run 37592301037 root-caused to daemon leak #5 and fixed                                                                                                                                                                                                                                                                                                                                 | commit `103a7ad`; charter green since                                                         |
| T02 per-module pin-drift guard | `check-pin-drift.sh` check 1b: every "Latest per module" claim == newest tag, both directions + reverse coverage from go.work members; **first live run caught the missed docs v0.3.2 same-train claim** (the exact drift class it was built for); AGENTS truth fixed; 3 negative tests red                                                                                                                                                                                                             | commits `1da7bfb`+`0c49125`; battery green                                                    |
| T03 core v0.8.0 train          | API-break check additions-only (StartHooks); suite+lint green; CHANGELOG dated; same-train AGENTS; annotated tag; pre-tag-checks green; pushed; integration re-pinned (canonical-zip url-rewrite trick)                                                                                                                                                                                                                                                                                                 | tag `v0.8.0` on origin; **proxy-check PASS**                                                  |
| T04 systemd v0.1.0 train       | replace lifted onto published core v0.8.0; suite+lint green; CHANGELOG [0.1.0]; CI proxy-smoke slot; integration pin + documentedPins + NEW composition E2E (`TestSystemdLifecycleThroughAppkitService`: READY=1/STOPPING=1 on a real unixgram socket through a live Service on PUBLISHED pins; `startup phase complete` observed live); TODO systemd item closed                                                                                                                                       | tag `systemd/v0.1.0` on origin; **proxy-check PASS**                                          |
| T05 push-lag watch             | `scripts/check-push-lag.sh` (threshold 6h default, env override); tested both directions; pre-commit wiring (warning-only)                                                                                                                                                                                                                                                                                                                                                                              | commit `a7c2796`                                                                              |
| T06 HEAD-advance hook          | `scripts/post-commit-head-watch.sh` — 5 discrimination rules (ancestor/amend/ORIG_HEAD/on-branch/else loud warning + `git branch recover-rewind <sha>`); scratch-repo simulation passed all assertions incl. a reflog-silent rewind; installed via `.githooks/`                                                                                                                                                                                                                                         | commit `1c591aa`                                                                              |
| T07 directive-drift CLASS fix  | **Discovery: `core.hooksPath=.githooks` (machine-wide) made `.git/hooks/pre-commit` INERT — no guard ever ran at commit time.** Live hooks now in tracked `.githooks/`: directives + dependabot-parity + charter FATAL at commit, push-lag warning, post-commit head-watch. Semver-aware compare + toolchain-directive checks already existed (2026-09-29). **Regression proven: a planted `go 1.26` drift was BLOCKED live**; empty restore commit soft-reset away                                     | commit `1c591aa`                                                                              |
| T08 AGENTS pipe rule           | "Never pipe verify/guard output" bullet landed in the batched AGENTS pass; removal trade honored (obsolete hook-reappend instruction out; 363/377 counted; structure linter 0)                                                                                                                                                                                                                                                                                                                          | commit `86b612d`                                                                              |
| T09 daemon dossier             | **ROOT CAUSE FOUND: pma = `projects-management-automation.service`** (system unit; `journalctl -t pma`; config read; Lars's own repo). Source-verified: committer-only (never mutates files — the go.work leaks come from a different tool, author still unknown), **disables ALL git hooks by design (`core.hooksPath=/var/empty`)**, AI messages from local qwen, 60s debounce/120s min-interval, auto_push=false. Full damage ledger + 4-decision owner triage checklist. Open question (6) RESOLVED | dossier `docs/status/2026-10-07_11-00_pma-daemon-root-cause-and-dossier.md`, commit `bd9abba` |
| T10 license-check decision     | Probes: license-check **PASSES** (13/13) with `env -u GOTOOLCHAIN`; lightning 19/19 without any GOTOOLCHAIN (auto-derive proven); ambient `GOTOOLCHAIN=local` still dies (exit 69, now with a loud preflight warn). Verdict: **the daemon's skip removal AND GOTOOLCHAIN-pin removal were correct** — accepted; stale AGENTS rationale struck                                                                                                                                                           | commits `b80c8bc` (daemon) + `86b612d` (AGENTS verdicts)                                      |

Also fixed en route: daemon leak #5 (go.work `./integration` re-add at 10:12,
bundled into the plan push window — `0cfc4ab`), charter green again.

## b) NOT STARTED (16 tasks)

T12 (flightrecorder E2E), T13 (cqrs flake hunt), T14 (browser CSP pass), T15
(upstream batch A — USER-GATE), T16 (errorpages swap), T17 (upstream batch B —
USER-GATE), T18 (docs verify sweep), T19 (examples + fr v0.2.1), T20
(govulncheck), T21 (consumer-claim ritual), T22 (benchstat), T23 (v1.0.0
criteria), T24 (cqrs-htmx suite), T25 (AGENTS slim-down), T26 (health/polish),
T27 (watchlist refresh).

## c) PARTIALLY DONE / IN FLIGHT (mid-flight breakage, flagged)

**T11 — docs-module composition E2E: ~70% done, one failing assertion.**

- DONE: docs@v0.3.2 pinned in integration (+ tidy, rode daemon `8266fee`);
  test written (`integration/docs_composition_test.go`); all three routes
  serve 200 + application/json **through the full appkit chain**, and
  `/docs/catalog.json` contains the builder title.
- FAILING: openapi.json/asyncapi.json bodies have `"title": ""` — a catalog
  built with a title but ZERO registered commands/events emits empty
  `info`. Fix (next step, ≤10 min): register one command/event on
  `cb.Builder()` (real coverage) or assert the stable schema keys
  (`"openapi": "3.0.3"` / `"asyncapi": "3.0.0"`); then the documentedPins
  fixture entry, suite green, commit.
- The failing test is committed WITH THIS REPORT as a flagged WIP (the
  daemon's auto-stage would otherwise land it attribute-less) — the commit
  message names the fix. Integration suite is red ONLY on this file.

## d) WHAT I TOTALLY FUCKED UP (self-critique)

1. **AGENTS multiedit mangling:** my insert-new-bullets edit anchored
   old_string on a bullet PREFIX — the pipe-rule bullet got GLUED onto the
   BuildFlow bullet's body. Caught by immediate re-view, repaired in one
   edit. Rule: never anchor an insert on a partial line.
2. **T11 left the tree red mid-flight** (failing title assertion) — I
   discovered the empty-title behavior by test failure instead of reading
   the docserver source first (read-the-API-before-asserting).
3. **Repeated daemon staging races (5× today, 10:22–10:50):** my
   `add && commit` batches kept losing intended messages to pma
   (`1da7bfb`, `535636e`, `6347ba5`, `80fb3da`, `b80c8bc`). Content always
   survived; attribution lost. I kept repeating the pattern instead of
   committing immediately after each file write.
4. `//` comment syntax in a bash hook (`.githooks/pre-commit` line 11) —
   caught by `bash -n`; sloppy language bleed.
5. First rewind simulation was a tangled one-liner that failed messily;
   rewrote as a script (already a known rule — don't chain complex
   sequences).
6. Misdiagnosed T10's subject at first (thought the staged .buildflow.yml
   diff WAS the license-check skip; it was the GOTOOLCHAIN removal — the
   skip had already been committed). Corrected by reading the actual diff.

## e) KEY DISCOVERIES (context the owner needs)

1. **The old pre-commit hook never ran** — machine-wide
   `core.hooksPath=.githooks` + no `.githooks/` dir = zero commit-time
   guards since the override landed. Now live and daemon-visible (though
   pma itself bypasses hooks — see next).
2. **pma disables ALL git hooks** (`core.hooksPath=/var/empty` in its
   unit): commit-time guards can never cover daemon commits; CI-side
   guards remain the only net for pma-bundled drift (and they DID catch
   every leak).
3. **pma is the bundler, not the mutator** — the go.work re-adds are
   authored by another tool (editor/gopls-class); journal lookup
   documented in the dossier to name it.
4. **BuildFlow modernized twice under our feet, both correct** —
   license-check works again; toolchain auto-derives. AGENTS now reflects
   it.

## f) UP TO 50 NEXT THINGS (execution queue, priority order)

1. T11 finish: register a command on the builder OR assert schema keys;
   documentedPins entry; suite green; proper commit.
2. T12: flightrecorder-middleware E2E via testkit.Serve (trigger → capture →
   snapshot endpoint), suite green, commit.
3. Observe origin CI for both trains (systemd proxy-smoke slot's first
   run; matrix still SKIPPED pending the secret).
4. T13: cqrs `-race -count=10` to FILE; analyze or close as unreproduced.
5. T14: launch health example `-hardened`; chromedp probe or manual CSP
   pass; link evidence in security/THREAT_MODEL.md.
6. T16: errorpages statusRecorder → httputil.ResponseRecorder + suite +
   lint + CHANGELOG [Unreleased].
7. T19: flightrecorder + otel examples StartHooks migration; fr v0.2.1
   bump in frh + cqrs (their next trains — NOT casual bumps).
8. T18: FEATURES evidence cells (core/realtime/security/errorpages then
   cqrs/otel/health/frh/docs/systemd), README claims, ROADMAP
   graduations; fix drift on sight; re-run guards.
9. T26: hardened-dashboard frame-ancestors+sorted-directives assertion,
   testkit drain-window helper, Hook alias/seriesKey (already in core
   v0.7.0? — verify against the polish list), errors.Join message-shape
   test, health example -hardened mode note.
10. T23: v1.0.0 exit-criteria review (keep draft vs graduate).
11. T22: benchstat probe (`nix run nixpkgs#benchstat`); otel n=10
    re-baseline if available.
12. T20: govulncheck health + security (network permitting).
13. T25: AGENTS deep slim-down (Release Ritual → doc/recipe; per-module
    Gotchas → module READMEs), ≤377 loop.
14. T27: watchlist refresh (cordis, PapDashboard, nixpkgs, art-dupl
    audit-pass scoping, dep-graph/dprint/samber owner-notes).
15. T24: cqrs-htmx setup suite hermetic run (their repo, bounded).
16. T15/T17 (USER-GATE): file the drafted upstream asks (go-cqrs-lite
    GracefulClose+timers, fr cooldown, go-sse ReplayFiltered, httputil
    Logging+NewServerListener, go-health pack, samber/do note) — all
    Lars's own repos.
17. T21 (USER-GATE): consumer-claim posture decision + ritual wiring.
18. Upstream note candidate: docs module empty-title behavior (a titled
    builder with zero entries serving `"title": ""` is arguably a module
    bug — check catalog upstream semantics first).
19. Update docs/status/README.md index for the two new reports (dossier +
    this one).
20. Delete the scratch /tmp files (headwatch tests, dumpdocs) — hygiene.
21. Re-run the FULL guards battery once after T11 lands (pre-tag-checks
    without a tag? no — run each guard script directly).
22. FEATURES rows for core v0.8.0 StartHooks + systemd v0.1.0 (rides T18).
23. Consider `journalctl -t pma` corroboration once the owner runs it (§6
    item 4 of the dossier).
24. Re-check AGENTS line budget after T25 slim-down (target ≤377
    permanently, ideally ≤350).
25. TODO_LIST pass at session end: strike completed T-items, re-date
    header, route leftovers.

## g) 3 QUESTIONS (unanswerable without the owner)

1. **pma posture (dossier §6):** of the four decisions — hooks-respecting
   mode, per-repo exclude/cooldown for train repos like go-appkit,
   attribution preservation (respect staged messages), and the
   `journalctl -t pma` lookup for the go.work mutator — which do you want
   pursued (all four are asks to your own pma repo; I can draft them)?
2. **The `SSH_PRIVATE_KEY` secret (T01):** still missing — restore ETA?
   The 12-module test matrix stays skipped (one designed-red preflight)
   until then; everything else on origin is green.
3. **Upstream filing go-ahead (T15/T17):** your "GET SHIT DONE" read to me
   as blanket approval, but every ask files an ISSUE in one of YOUR other
   repos (go-cqrs-lite, go-sse, httputil, go-health, go-flightrecorder) in
   your name — confirm filing now, or hold the drafts in
   `doc/feedback/outgoing/` for review first?

---

**Stopping here as instructed. The execution queue resumes on your word.**
