# Status — folded-contract release wave, mid-train (2026-11-08 06:37)

Session mandate: execute the WHOLE actionable TODO_LIST core — the 20.2
folded-contract implementation (done in-tree by the prior session), the httpx
W3 remainder batteries B3/B4/B5/B7 (done in-tree), plus ride-alongs: test-helper
migration closure, httpx LICENSE, systemd backlog, v1.0.0 exit-criteria refresh,
release trains, truth-sync. This session picked up at "CHANGELOG entries were
the task in progress" and ran the release train up to (but not past) the local
tags. **Nothing is pushed. The last combined integration re-run went red
(isolation class) — root-causing it is the gate to the push.**

The owner's three open questions from the 03-28 report were interpreted as
answered by the "GET SHIT DONE — the WHOLE LIST" directive: (1) wave
authorized, (2) realtime test-helper deviation stands (no overrule), (3)
systemd ships in this session (implemented, tagged v0.2.0).

## a) FULLY DONE (this session, all verified by CLI — LSP diagnostics were
stale/false again, CLI disproved all 23 warnings)

- **Four-module verification, green:** cqrs / realtime / httpx / systemd
  suites `GOWORK=off GOTOOLCHAIN=go1.27.3 go test -race -count=1` exit 0;
  CLI `golangci-lint run` 0 issues on all four (run sequentially per the
  cache-race rule). systemd re-verified after my edit; `systemd/example`
  builds.
- **systemd v0.2.0 implemented:** `Install` now returns the wired `Hooks`
  (failed-start watchdog-stop escape hatch), godoc + README + doc.go
  truth-synced, `TestInstall_ReturnsTheWiredHooks` pins idempotency and that
  the returned Shutdown is the wired one. Signature change is
  source-compatible in practice (statement-context call sites compile
  unchanged — Go discards added returns); migration note in the CHANGELOG.
- **CHANGELOGs written and DATED** (datings-first, tags-last discipline):
  cqrs `[0.9.0] - 2026-11-08`, realtime `[0.2.0] - 2026-11-08`,
  httpx `[0.2.0] - 2026-11-08`, systemd `[0.2.0] - 2026-11-08`. Entries
  cover folded notifications, FoldedSource, B3/B4/B5/B7 + the two new direct
  deps + lint-config additions, and the Install return.
- **API-break proofs (release ritual step 1):** `go doc -all` old-tag vs
  tree. cqrs: additions-only (4 removed lines are doc prose); realtime:
  additions-only (3 doc lines); httpx: pure additions (183); systemd: the
  expected `Install` signature change (12 lines) — minor bump + migration
  note is the correct 0.x handling. First attempt diffed the WRONG side
  (git archive extracts the whole repo; old-side doc must run from
  `<extract>/<module>`) — caught by the "package appkit" smell in the diff.
- **Standing guards green:** cqrs-lint v4.21.0 (exit 0), dependabot parity
  (httpx directory entry confirmed), go-directives, workspace-charter,
  module-licenses, all exit 0.
- **httpx LICENSE** verified present in-tree (the v0.1.0 tagged tree lacked
  it; v0.2.0 carries it).
- **Exit-criteria refreshed** (`docs/planning/core-v1-exit-criteria.md`):
  criterion 1 **MET** — v0.10.0→v0.11.0 and v0.11.0→v0.12.0 are two
  consecutive additions-EMPTY diffs (one full frozen minor cycle); criterion
  5 (PapDashboard overlap + observed joint release cycle) is the sole
  remaining v1.0.0 blocker, external-clock. Restraint note recorded: core
  tags between now and v1.0.0 must stay additions-empty.
- **Four annotated tags created LOCALLY on a clean tree:**
  `cqrs/v0.9.0`, `realtime/v0.2.0`, `httpx/v0.2.0`, `systemd/v0.2.0`
  (messages state the semantic deltas).
- **Integration re-pinned via the insteadOf ritual:** `go get` + `go mod
  tidy` against the local tags (canonical zips → go.sum matches what
  consumers get): cqrs v0.9.0, realtime v0.2.0, systemd v0.2.0, **httpx
  v0.2.0 newly pinned**; `documentedPins` updated same-change; insteadOf
  rewrite unset after.
- **`integration/folded_e2e_test.go`** — the W5-C2 discriminator E2E per
  design note §4: positive (broadcast payload must contain the mid-fold
  marker for exactly the trigger event id) + deterministic negative control
  (naive signal-before-fold wiring provably broadcasts a marker-less
  payload; the fold is gated until the stale broadcast is observed). Both
  PASS under `-race` in isolation.
- **`integration/httpx_e2e_test.go`** — B4 Conditional through the default
  stack: 200+strong ETag → If-None-Match 304, and the 304 still carries
  Server-Timing (the default-stack composition pin). PASSES under `-race`
  in isolation.
- **Two real harness bugs root-caused and fixed** (found via three
  instrumented probe tests, all deleted after use):
  1. SSE request context cancelled by a `defer cancel` in the harness
     helper — the stream died when the helper returned.
  2. **The subtle one:** `StartProjections(ctx)` with a helper-scoped
     `defer cancel` KILLS the projection host — `projectionhost.Start`
     derives its run context from the Start ctx, so the workers died the
     moment the harness returned, silently (dispatch still 204s; only the
     projection path goes dead). Fixed: the Start context must outlive the
     helper (es.Shutdown owns host teardown).
- **Full integration suite `-race` GREEN once** (exit=0) after the fixes.
- **FEATURES.md truth-synced:** cqrs/realtime sections gained their first
  "Shipped through" lines (v0.9.0 / v0.2.0), systemd → v0.2.0, httpx →
  v0.2.0 with B3/B4/B5/B7 rows, integration section +2 contract rows
  (folded discriminator, httpx conditional E2E).

## b) PARTIALLY DONE

- **Integration suite stability:** green in isolation and in the first full
  run; the LAST combined `-run "TestFoldedContract|TestHTTPx"` re-run went
  RED (see §d). Root-cause is the gate before any push.
- **Truth-sync:** FEATURES done; **AGENTS.md Release State, TODO_LIST rows,
  docs/agents/module-details.md NOT yet updated** (planned for the same
  master push).
- **Release wave:** tags local-only; pre-tag-checks not yet run; no push.

## c) NOT STARTED

- `./scripts/pre-tag-checks.sh cqrs/v0.9.0 realtime/v0.2.0 httpx/v0.2.0
  systemd/v0.2.0` (12 guards + tagged-artifact asserts; gosum-canonical
  integration leg expected DEFERRED on the four local-only tags).
- Tag push → proxy `.zip` 200 poll → `check-gosum-canonical.sh integration`
  re-run → master push.
- Final fleet verification: `scripts/verify-all.sh`, `scripts/example-smoke.sh`,
  structure-linter over httpxtest, response-drain + sse-protocol re-run
  (my new files carry `respdrain:ignore` SSE carve-outs), CI green on the
  final pushed commit (+ tag-visibility race watch, `gh run rerun` if ghost).
- TODO_LIST row closures (test-helper row incl. the realtime deviation
  note, systemd backlog row, W3-W5 battery row, exit-criteria row) and
  header rewrite; this report's index entry.
- Fresh-consumer proxy checks (`docs/recipes/fresh-consumer-proxy-check.md`)
  for the four tags after push.

## d) TOTALLY FUCKED UP / went wrong

1. **Last combined run (06:2x): `TestHTTPxConditionalThroughDefaultStack`
   FAILED** — testkit.Serve's goroutine-leak detector: "19 goroutines after
   shutdown, baseline 4"; the dump shows watermill GoChannel goroutines.
   The httpx test builds NO engine — these are leftovers from the folded
   engine tests running in PARALLEL in the same process. Process-global
   leak detector + parallel engine tests = cross-test false positive class.
   In isolation the test passes.
2. **Same run: `TestFoldedContract_PreFoldWiringFailsTheSameAssertion` hung
   10m** → suite panic-timeout. It passed in isolation earlier. Deadlock
   not yet root-caused; suspects: gate-release vs es.Shutdown cleanup
   ordering under failure paths, or the SSE read blocking past the client
   timeout under load (the box's load-flake protocol applies — not yet
   checked: `uptime`, `-count=10` hermetic re-runs).
3. **My process errors (all caught in-session):** sed-edited a file outside
   the edit tool (mod-time conflict on re-edit); wrote probe 5 with garbage
   placeholder identifiers (compile noise, immediately discarded); linted
   integration WITHOUT `GOWORK=off` (exit 7 — the exact documented
   gotcha); edited files before Viewing them (tool correctly refused);
   first API-diff pass diffed the wrong side (whole-repo archive). Plus one
   real lint finding in my new code (modernize mapsloop → `maps.Copy`,
   fixed, lint now 0).

## e) WHAT TO IMPROVE

- **Integration commands discipline:** the GOWORK=off omission was MY
  repeat of a documented AGENTS gotcha. Copy from integration/README.md
  verbatim, every time.
- **Test-isolation strategy:** engine-backed tests + `t.Parallel()` +
  testkit's process-global goroutine-leak baseline do not compose. Either
  the folded tests must not leak bus goroutines (hard: watermill gochannel
  leaves senders until closed — maybe `es.Shutdown` ordering), or the leak
  check needs a documented exemption class for engine tests, or engine
  tests drop `t.Parallel()`. Decision needed (question 2).
- **The StartProjections ctx-lifetime trap is a real API footgun:**
 cqrs/`projectionhost` derives the host run context from the Start ctx;
  a short-lived Start ctx silently kills all workers later. Worth a cqrs
  godoc warning + README note (consumer-facing, next cqrs train or an
  upstream ask to go-cqrs-lite).
- **Probe discipline:** instrumented probes found both bugs in minutes —
  keep that tool; but write them carefully (probe 5's placeholder junk was
  pure sloppiness).
- The negative control's determinism gate (blocking Handle) trades a hang
  class for determinism; the cleanup release covers the fatal path, so the
  10m hang means an uncovered path exists — find it, don't paper over it.

## f) NEXT (ranked)

1. Root-cause the combined-run failures: `uptime` (load-flake protocol),
   re-run each new test `-count=10` hermetic and in combination; get a
   deterministic read on the 10m hang.
2. Fix accordingly: cleanup-ordering or SSE-read deadline in the folded
   control; isolation for the httpx leak-detector trip (order, serial
   engine tests, or detector exemption).
3. Re-run FULL integration suite `-race` green ×2 (loaded + quiet).
4. `gofumpt` + `golines -m 120` pass over the new integration files.
5. Re-run response-drain + sse-protocol guards over the new files.
6. AGENTS.md Release State: append the 2026-11-08 four-tag wave entry;
   update "Latest per module" (cqrs v0.9.0, realtime v0.2.0, httpx v0.2.0,
   systemd v0.2.0); module-details systemd/httpx notes if needed.
7. TODO_LIST: close the test-helper row (with realtime deviation note),
   systemd backlog row (WithStatus stays demand-gated), exit-criteria row
   (criterion 1 MET), W3-W5 row (B3/B4/B5/B7 shipped; B6/B8 owner-gated);
   header rewrite.
8. `./scripts/pre-tag-checks.sh cqrs/v0.9.0 realtime/v0.2.0 httpx/v0.2.0
   systemd/v0.2.0` — expect gosum-canonical DEFERRED (local-only tags);
   everything else must be green.
9. Push the four tags ONLY; poll proxy `.zip` per module (download tool);
   expect the first post-tag CI run to be the visibility-race casualty
   (`gh run rerun` per the gotcha).
10. Re-run `./scripts/check-gosum-canonical.sh integration` until green.
11. Push master (release prep + integration re-pin + all truth-sync docs).
12. Fresh-consumer proxy checks for all four tags (scratch module, fresh
    GOMODCACHE, compile-level usage).
13. `scripts/verify-all.sh` full fleet loop; `scripts/example-smoke.sh`;
    structure-linter (httpx/httpxtest must be clean or excluded with
    reason); agents-deps / status-index guards.
14. Verify CI green on the FINAL pushed commit; confirm no new jobs
    silently skipped (docs-only pushes are skipped BY DESIGN).
15. Close out: annotate this report, update the 03-28 report if the owner
    answers differ from my interpretations.
16. (Queue, not this train) cqrs godoc/README: the StartProjections
    ctx-lifetime warning.
17. (Queue) Watch: setup/PapDashboard consumer-claims guards fire on their
    next tags; consumer heads-ups for httpx v0.2.0/systemd v0.2.0 NOT
    needed (no consumer pins them) — verify via check-consumer-claims in
    pre-tag-checks.

## g) Questions for the owner (cannot be derived)

1. **Push authorization, explicit:** tags (4) + master push once everything
   is green — confirm. The mandate reads as yes; pushing is the one
   irreversible step, so I am asking once before doing it.
2. **Goroutine-leak detector policy:** testkit.Serve's leak check is
   process-global; parallel engine tests (watermill bus) trip it
   cross-test. Policy call: (a) engine tests drop t.Parallel() /
   serialize, (b) a documented exemption/annotation for engine tests, or
   (c) the folded tests must genuinely not leak bus goroutines (stricter)?
3. **Negative-control design constraint:** if the 10m hang root-causes to
   the blocking-gate design interacting with shutdown, may I restructure
   the control (e.g., observe-stale-then-assert without blocking Handle),
   or is the deterministic mid-fold gate a hard requirement from the CV
   discriminator spec?

## h) CLOSE-OUT (2026-11-08 ~10:0x, annotated post-completion)

All of §f 6-15 executed; the wave SHIPPED (cqrs v0.9.0, realtime v0.2.0,
httpx v0.2.0, systemd v0.2.0 — pushed, proxy-verified, fresh-consumer
×4, verify-all + example-smoke green, CI green on the final code commit
`6c6c9e6` modulo the designed ci-preflight secret gate). Full narrative:
the 08-38 report §h close-out (hang fix → re-tags → pre-tag-checks →
push → gosum repair → CI fixes). Question answers: (1) push authorized
(owner paste instruction 8) and done; (2) goroutine-leak policy — MOOT
(see the 08-38 root-cause: the trip was the hung test's engine leaking;
tests that complete coexist cleanly with the detector, no policy change
needed); (3) negative-control redesign — YES, restructured (observed
signal + 2s fallback), the design note §4 carries the ERRATUM. Item 16
(cqrs StartProjections ctx-lifetime godoc) and the upstream
sync-publish/blocking-handler ask are PARKED as TODO_LIST rows awaiting
owner input.

— Session closed out.
