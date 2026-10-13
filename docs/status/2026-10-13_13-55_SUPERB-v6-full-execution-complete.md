# Status: SUPERB v6 Full Execution — COMPLETE (T18–T27 + push tail)

**2026-10-13 13:55 CEST · Executor: Crush (glm-5.3) · Plan:**
[`docs/planning/2026-10-12_15-17_SUPERB-v6-trains-guards-and-gated-unlocks.md`](../planning/2026-10-12_15-17_SUPERB-v6-trains-guards-and-gated-unlocks.md)
**· Prior snapshot:** [`2026-10-12_20-36_SUPERB-v6-mid-train-status-report.md`](2026-10-12_20-36_SUPERB-v6-mid-train-status-report.md)
**· Defaults taken (re-issuance without answers):** file family asks, per-train push, samber/lo test-only.

---

## Stat cards

| Done | Partial             | Not started    | Fucked up                    |
| ---- | ------------------- | -------------- | ---------------------------- |
| 12   | 3 (owner/env-gated) | 0 (all routed) | 1 class (daemon attribution) |

## a) FULLY DONE

1. **Push tail** — ~20 verified commits pushed (a2cf750); both train tags
   (cqrs/v0.8.0, flightrecorderhealth/v0.1.7) were already on origin.
2. **T18 annotations complete** — 2026-10-07 ×6, 2026-10-10 ×2,
   2026-10-12 ×2: every §f/§g item resolved inline with evidence or left
   honestly bare (owner-gated: SSH secret, buildcache topology,
   cqrs-htmx Domain adoption). check-rows green (one pre-existing partial
   table from the morning pass, informational only). Five missing report
   rows indexed in `docs/status/README.md` + the `.md`-at-explicit-demand
   convention line. Harvest: 2 TODO rows + ROADMAP fuel; 4 completed rows
   dropped (evidence in CHANGELOGs).
3. **cqrs flake hunt CLOSED with a real find** — `-race -count=10`
   reproduced a genuine test race: `captureHandler` reads unsynchronized
   while parallel tests' construction-time SCREAM logs append via the
   process-global default logger. Fixed with a mutex-guarded `snapshot()`;
   ×10 re-run green (35s, zero races); lint 0. Three reports' flake items
   closed on this evidence.
4. **T23 — errorpages v0.1.2 RELEASED** — floor 1.27.2 + dependency sweep
   (templ-components v1.21.0, error-family v0.11.0, httputil v1.5.0) +
   example PORT. **The swap was DECLINED at source level**: httputil's
   `ResponseRecorder` is pass-through (every Write/WriteHeader delegates,
   recorder.go:109-132); `Wrap` needs body-DISCARDING capture — wrapping a
   discard writer inside it would rename the hand-roll, not remove it.
   Decision record in the CHANGELOG; zero API delta (patch); pre-tag gate
   green (70 OK); pushed; proxy PASS.
5. **Gate bug fixed + mid-train guard catch** — the pin-drift guard caught
   **cqrs-htmx v4.14.1** published mid-train; integration swept to it
   (suite green). The catch exposed `pre-tag-checks.sh`'s unbound `$tag`
   in the guards' fail path (`set -u` turned a FAIL into a crash) — fixed.
6. **T19 pkg.go.dev** — cqrs@v0.8.0, frh@v0.1.7, health@v0.2.0 RENDER;
   flightrecorder was NEVER indexed (even v0.1.1 404s) — index requests
   submitted (page GETs are the documented trigger); proxy resolves
   v0.2.0 cleanly so consumers are unaffected; cosmetic-only (godoc hidden
   by license choice). core@v0.9.0 + errorpages@v0.1.2 404 at <36h/<1h —
   normal crawler lag, re-check next session.
7. **T25 govulncheck — 12/12 modules CLEAN** via `nix shell
   nixpkgs#govulncheck` (the "needs a networked machine" belief was wrong;
   the vuln-DB fetch works here). Noted in AGENTS Build & Verify.
8. **T24 browser CSP pass — real defect found AND fixed** — headless
   Chromium against the health `-hardened` example: the dashboard 404'd
   because `DashboardHardenedPreset("/health", …)`'s basePath is a PREFIX
   (dashboard actually at /health/health) and trend/metrics options were
   silently replaced. Example fixed (empty basePath + appended options);
   integration test gained the missing dashboard-page leg (`/health/health`
   — the earlier `/health` assertions exercise appkit's own default health
   handler). Browser verdict recorded in `security/THREAT_MODEL.md`:
   renders under the strict CSP (nonce'd scripts execute), Tailwind-CDN
   inline styles + Datastar eval blocked — the policy working as designed;
   upstream already documents unsafe-eval as Datastar-required.
9. **T20 pma** — `journalctl -t pma` ran clean: pma logs NO file names
   (0 go.work mentions in 28.5k lines) — the dossier's §6.4 lookup hope is
   dead, recorded as the closing evidence. Owner decision memo written
   (`doc/planning/2026-10-13_pma-owner-decision-memo.md`): keep hook
   bypass, cooldown (not exclude) for train repos, respect/skip staged
   messages, accept the mutator cadence pending BuildFlow#46 — plus this
   session's live race evidence (3 attribution-erasing interleaves, 1
   commit-failed race, 1 pma restart).
10. **T22 — 4 family asks FILED after premise re-verification at master**:
    go-cqrs-lite **#67** (GracefulClose drain-error early return — the
    timers corollary was FIXED upstream and struck from the ask);
    go-health **#7** (the panic-comment stance UPGRADED to a real
    classification bug — `grades` matches the batch-level sentinel only,
    per-check panics WARN on non-critical; + recorder-sentinel decision,
    `Evaluate` cache godoc, restart latch); go-flightrecorder **#2**
    (cooldown combinator); httputil **#41** (`NewServerListener` —
    unblocks Service composition + Core TLS). **Two premises dissolved
    during verification**: go-sse `ReplayFiltered` and httputil
    `LoggingFromContext` already exist at master (no filing; appkit-side
    follow-ups recorded). The dashboard strict-CSP ask was NOT filed —
    upstream's `RecommendedCSP` documents unsafe-eval as
    Datastar-required; filing would relitigate a documented decision.
    samber/do + exhaustruct stay gated (external repos).
11. **T26/T04 residue + structure guard red→green** — CI on the tail:
    **example-smoke PASSED its first origin run**; all 11 proxy-smokes,
    go-directives, cqrs-lint, pin-drift, config-parity, workspace-charter
    green. structure-linter failed at GOPRIVATE install since the job
    landed (invisible red) — now rides ci-preflight like the test matrix.
    Behind that fix a second real find: the v0.8.0 trailing E014 directive
    tripped `package-naming` — moved to a comment block above the clause
    (cqrs-lint still suppresses; structure guard 0 findings; suite green).
    module-details errorpages section re-verified on its train (was 3
    versions stale). 220-cap source confirmed: the installed linter
    binary's agents rule; AGENTS at 126/220.
12. **T27 watchlist swept (as-of 2026-10-13)** — **PapDashboard v0.3.0:
    the reverse-adoption trigger FIRED** — their go.mod requires
    `go-appkit v0.4.0` (core only; their PAP TLS envs were REMOVED with
    the adoption, so the Core-TLS trigger did not fire). Consumer #2
    exists. nixpkgs go = 1.27.2 (matches our floor); httputil v1.6.0 still
    untagged; cqrs-htmx master daemon-active, no v5 signal; batteries
    demand unchanged (zero composition consumers).

## b) PARTIALLY DONE (gated, honestly)

1. **SSH_PRIVATE_KEY** (owner P1): ci-preflight stays the one designed-red
   job; matrix + structure-linter skip until restored.
2. **samber/lo + exhaustruct_v5 filings** (external repos) — recorded,
   stay gated per the family-only default.
3. **BuildFlow #46–#49 follow-ups** — external watch (jscpd reconsider,
   go-work-sync drop-skip, lychee decision, skill PR): routed TODO/ROADMAP.

## c) NOT STARTED

None — every remaining plan item is done, declined-with-evidence, or
routed to an owner/env/external gate recorded in TODO_LIST.

## d) TOTALLY FUCKED UP (honesty ledger)

1. **The daemon ate attribution again (the known §6.3 class, now
   evidenced)**: the race-fix commit and two annotation batches landed as
   `chore: auto-commit N files (heuristic)` — the commit that closed a
   real `-race` find carries no what/why in history. All content verified
   landed; the reports are the manifest (the memo's recommendation 3
   addresses the class).
2. **My first gate run crashed confusingly** — the unbound-`$tag` bug
   masked WHICH guard failed (I read the full log before concluding; the
   pipe rule held).
3. **Date header off by one day** — the env said 10/12; `date` said
   10/13. Caught ~2h in via the daemon's commit timestamps; every
   same-session citation corrected before commit (the repo's own date-first
   rule, applied one turn late).

## e) WHAT WE SHOULD IMPROVE

1. **Verify the premise before executing the plan item** — T23's swap and
   T22's go-sse/Logging asks all dissolved at the source level. The plan
   carried month-old premises; the first move on any "swap/ask" item is
   re-reading the target's current master.
2. **The structure-linter CI job was invisibly red for a day** — a guard
   job that cannot install should look SKIPPED, not FAILED. The
   ci-preflight pattern is now applied; audit other jobs for the same
   class when adding them.
3. **Run the full local guard battery on trains, not just the pre-tag
   five** — the E014/package-naming conflict shipped in v0.8.0 because
   check-structure-linter.sh wasn't in the train's verification. Cheap
   candidate: add it to pre-tag-checks.

## f) NEXT (routed, not stranded)

All items live in TODO_LIST (harvested this session): SSH secret (owner),
BuildFlow follow-ups, lint/config hygiene smalls, docs/agents charter,
i18n residue (gated on httputil v1.6.0), composition-gap watchlist,
cqrs-htmx hermetic run (their repo), PapDashboard version-refresh (their
repo), flightrecorder pkg.go.dev re-check, exhaustruct/samber-do gates.

---

**Verification:** every module touched this session: suite `-race` green +
lint 0 + relevant guards green; errorpages v0.1.2 proxy PASS; 4 issue
URLs confirmed live; CI on final push green except the designed preflight.

_Session commits: 53888d9 (annotations+flake close) · 8d441fa/eeed757/caf4d09
(errorpages v0.1.2 train + gate fix + re-pin) · ce8738e/d960294 (browser
pass + example fix) · b14e29b (pma memo) · 50e80ef (filings) · 5d7d0fb
(CI preflight gating) · 9f07bad (E014 directive) · this harvest — plus
daemon interleaves carrying the same content where they won the race._
