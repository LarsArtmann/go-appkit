# Status Report — Integration Shutdown-Flake Root Cause, Adoption-Counter Corrections, v13 Backlog Wave

**Written:** 2026-11-06 ~14:25 +0100 · **Session window:** ~13:28–14:25 ·
**Trigger:** owner instruction to execute the 13-13 report's §f NEXT list (42 items, impact-ordered).
**Method:** executed the actionable items top-down; two subagent fan-outs (config depth-audit,
archetype matrix) ran read-only against consumer repos; the probe-site audit subagent's static
verdicts were checked against a live goroutine dump before being believed.

## a) FULLY DONE (this session, verified)

1. **§f1 CLOSED — the standing structure-linter red.** `cqrs/commands.go:1` carried the E014
   directive back ON the package clause (6fd1de8's "re-anchor" reintroduced the exact pattern
   9f07bad had proven broken). Moved to the blank-separated comment block above the clause
   (9f07bad shape). Verified green on ALL SIX gates: structure-linter 0 findings, cqrs
   build+vet+`-race` suite green, cqrs-lint exit 0 (2 suppressions effective — the E014
   "stale suppression" warning is cosmetic; above-line anchoring was tested and does NOT
   work for package-clause findings, so this shape is final), golangci-lint **pinned v2.14.0
   → 0 issues**. Version-drift trap found on the way: the machine PATH now carries
   golangci-lint **v2.15.0**, which reports 8 wrapcheck findings on untouched
   `eventservice.go` — v2.14.0 (the fleet/CI pin) reports 0. The local binary drifted past
   the fleet pin; the CI job is the verdict that counts.
2. **§f2 CLOSED — P8 verified by me.** Integration hermetic (`GOWORK=off`, verbatim README
   commands): tidy no-op, suite green, vet 0, lint 0 (pinned v2.14.0). documentedPins fixture
   and pins confirmed stable (core v0.10.0 + preset v0.1.0 + the full 14-module pin set).
3. **§f3-5 CLOSED — P9c, all three fresh-consumer proxy checks PASS** (preset/v0.1.0,
   config/v0.1.0, core v0.10.0) via the recipe script with `GOPRIVATE=` cleared, GOPROXY
   forced, isolated GOMODCACHE — genuine proxy-side resolution.
4. **§f11 CLOSED — pkg.go.dev preset@v0.1.0 RENDERS** (Valid go.mod, Imports 16, Published
   Nov 6). License "UNKNOWN" is the fleet-decided proprietary posture (LICENSE is IN the
   tag; ls-verified), not a holdout.
5. **§f9-10 CLOSED — OG2 consumer heads-ups POSTED with claims verified at source level
   first** (verify-before-filing gate): `Server-Timing: total;dur=<ms>` ms-units confirmed
   in servertiming v1.0.3 (`float64(d)/float64(time.Millisecond)`), stack order + `&false`
   knob + preset env grammar + dashboard endpoints confirmed against tree. Filed:
   cqrs-htmx [#43](https://github.com/LarsArtmann/cqrs-htmx/issues/43), PapDashboard
   [#6](https://github.com/LarsArtmann/PapDashboard/issues/6). TODO row 68 marked [✓].
6. **§f8 CLOSED — R4b Server-Timing composition E2E LANDED**
   (`integration/server_timing_e2e_test.go`, 3 tests: default-stack header contract,
   knob-off omission, shutdown-loop tie): 3/3 green under `-race`, lint 0, pinned tags.
7. **THE MOVING INTEGRATION SHUTDOWN-FLAKE ROOT-CAUSED AND FIXED** (the session's deepest
   dig — §f2's verification kept failing on DIFFERENT tests each run: i18n, then
   security_errorpages, never in isolation, no-race reproducible). Static audit said
   everything drains (true — every probe drains). A temporary stack-dump instrumentation
   caught the holder live: `net/http.(*conn).serve` parked in `readRequest` on a conn whose
   request never came — a **StateNew** spare conn dialed by the shared `http.DefaultTransport`
   under parallel-burst contention — and stdlib `closeIdleConns` (issue 22682 grace) refuses
   to close a StateNew conn until it is 5s old, exactly the test cleanup's 5s ctx. Fix:
   `http.DefaultClient.CloseIdleConnections()` at the top of the two explicit shutdown
   cleanups (every per-call client in the suite shares DefaultTransport's pool, so one call
   covers all). Post-fix: **9 consecutive full-suite `-race` greens** (3×3 loop) + 20/20
   isolated + the class-2 RST generators from the audit ALSO drained (waitReady ×100,
   postStatus, contract_parity's false "drain" comment, hardened_dashboard, flightrecorder ×3).
   AGENTS.md Testing gotcha now carries the second mechanism.
8. **§f15 CLOSED — crush-config lessons committed** (`20587f0`, local): asserted-vs-computed
   counts (new) + redirect-reflex addendum (extension of the existing pipe lesson — no
   parallel entry, per the no-duplication rule). Hooks passed, dprint clean.
9. **§f16 CLOSED — dprint html coverage DECIDED: documented exclusion.** `**/*.html` in
   `dprint.json` excludes + rationale in the AGENTS.md linting ¶ (the 15 tracked html files
   are point-in-time rendered reports; a future html plugin must lift the exclusion
   deliberately). Also formatted the daemon-drifted docs/status/README.md tables —
   `dprint check` green repo-wide.
10. **§f18/§f20 CLOSED — adoption counter CORRECTED (two real bugs), re-baselined, per-archetype
    targets in hand.** The archetype-matrix subagent's 41-app population contradicted the
    counter's 31 — reconciliation found: (1) **case-sensitive dir resolution silently
    dropped 15 meta-name↔dirname mismatches** (population 31→41; two adopters invisible:
    PapDashboard, RR-golang); (2) **per-go.mod (not per-line) indirect check** let
    PapDashboard's indirect `security` sibling poison its DIRECT core verdict (direct 6→7).
    TRUE numbers: **any 9/41, direct 7/41**, baseline rewritten (14:13). `meta` `appkit` tag
    applied to all 9 adopters → the population leg is now `meta ls --tags appkit`. Archetype
    matrix: daemons 25 / cli-daemons 11 / clis 5; appkit adoption 6/25 + 3/11 + 0/5 —
    the frontier is daemons.
11. **§f22 CLOSED — config depth-audit (subagent) → audit doc Errata 2.** "22 hand-rolled
    config packages" is reproducible but undercounts (28 apps at any config path, 27
    distinct; CV/cv-stage byte-identical worktrees) AND overstates (all top-5 are
    koanf/viper-BACKED — F1 deletes 15-30% of each package, not the packages). Appended as
    a dated errata block, dprint-clean.
12. **§f21 CLOSED — browser-render probe with the honest overflow check** (13-13 §e's
    scrollWidth probe, chromium headless on /tmp copies): **12/15 tracked html artifacts
    overflow a 500px viewport**, 3 clean. Class observation (desktop-designed report
    corpus), not 12 defects; recorded here, artifacts untouched.
13. **§f23 CLOSED + §f39 CLOSED.** meta `appkit` tag (9 projects; `meta tags --list` is the
    clean one-liner) — and a NEW find: **meta's tag filter matches null-tagged projects**
    (`ls --tags zzznotarealtag` returns monitor365-ui-sdk, an untracked-tags record);
    upstream-ask candidate, meta repo. Screenshot referenced from `docs/review/libs/index.html`
    Appendix ("Render evidence" ¶ — wording kept honest: non-blank, not mobile-OK).
14. **TODO_LIST dispositions** — header numbers corrected (41/9/7 with the counter-bug
    note), preset row dual-number citation fixed, row 46 + row 50 legs closed inline,
    row 68 marked [✓] with issue links.
15. Machine env incident diagnosed (workaround held all session): `~/.cache/go-mod` became
    **root-owned at 12:04 today** — every bare `go` command died `mkdir .../cache: permission
    denied` until commands carried `GOMODCACHE=/mnt/buildcache/go/pkg/mod`. Something ran as
    root on this box mid-day; owner should check what and restore ownership.

## b) PARTIALLY DONE

- The archetype-matrix subagent's direct count (7) initially disagreed with the first fixed
  counter (6) — reconciled as counter bug #2 (per-file indirect); the subagent's static
  flake analysis (SSE teardown + scheduling) did NOT name the StateNew mechanism and
  disputed the abandoned-client theory — resolved empirically by the live dump, not by
  either static story. Reports are inputs; the stack dump was the verdict.
- The counter is now SLOW (~2 min: per-line `printf|grep` forks over ~2k go.mods). A single
  awk pass would fix it; noted, not done.

## c) NOT STARTED (still open in the §f ledger)

- A2-A5 swap recipes, D1/D3 batteries, C2R realtime, X-train compose refactor, S2-S7
  scaffold program (owner-gated), H-train httpx (owner-gated on H8), pkg.go.dev
  flightrecorder page (owner-gated re-tag last resort), six-month adoption re-sweep.

## d) TOTALLY FUCKED UP (caught, all recovered)

1. **The counter itself was the asserted-number class it exists to prevent** — shipped with
   two verdict bugs hours ago and cited as truth ("any 7/31 direct 5/31" went into the
   baseline + my earlier tag set). Caught because a second instrument (the archetype
   subagent) disagreed; fixed + re-baselined + re-tagged. The lesson I committed to
   crush-config today bit ME today.
2. Two wrong mechanism theories for the flake before the decisive evidence (abandoned
   clients → i18n shared-client hardening; then drain-again suspicion). The 20/20-isolated
   green + moving target + no-race repro forced the stack-dump approach that actually
   settled it.
3. `runtime.Stack([]byte(""), true)` — misread the signature (buf is IN, not OUT); compile
   error wasted one loop iteration.
4. Environment friction (not mine, but cost time): root-owned GOMODCACHE (see a15),
   golangci-lint v2.15.0 drift.

## e) WHAT WE SHOULD IMPROVE

- **Two instruments disagreeing is a feature**: the counter/archetype cross-check is the
  only reason the counter bugs surfaced. Cheap corollary: any committed local instrument
  should be cross-checked by an independent method ONCE before its numbers get cited.
- Live goroutine dumps beat static analysis for "what is blocking" questions — the
  temp-instrumentation pattern (WriteFile the Stack dump in the failing cleanup) is
  reusable; consider a testkit `OnShutdownTimeout` hook upstream-side.
- Suite loops ran at load 16-25 (not quiet-run-gated); the fix makes the suite robust at
  that load, which is the practical point — but benches still get quiet-run only.

## f) NEXT — impact-ordered

1. **Push master + CI verify on the FINAL pushed commit** (§f7's gating action; the 13:27
   daemon commit + this session's ~10 files). Verify: structure-linter job RUNS (not
   skipped) post-push, proxy-smoke preset leg, agents-deps live run.
2. meta null-tags bug → FILED post-report: project-meta [#4](https://github.com/LarsArtmann/project-meta/issues/4)
   (verified repro + version stamped in the issue).
3. golangci-lint v2.15.0 machine drift → check-golangci-configs.sh + CI pin stay v2.14.0;
   consider pinning the local nix profile.
4. Root-owned ~/.cache/go-mod → owner machine fix (see a15). **→ TODO P3 — Machine ops (root-owned go-mod cache machine fix) <!-- fresh:skip: routed in TODO_LIST -->**
5. Then the standing §f ledger: A2 dnsblockd swap recipe next (template ready, −~1,600 LOC).

## g) Questions I cannot answer myself

1. What process ran as root at ~12:04 today and chowned `~/.cache/go-mod`? (BuildFlow?
   a nix activation? another session under sudo?) — decides whether the fix is `chown`
   back or a deeper guard.
2. File the meta null-tags-filter bug upstream now, or batch it with the next meta touch?
