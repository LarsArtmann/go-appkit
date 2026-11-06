# Status Report — preset release train (SUPERB v13, sessions 5→6): split-brain resolution, hang root-cause, preset v0.1.0

**Written:** 2026-11-06 12:49 +0100 · **Session window:** 08:00–12:45 (active execution
~08:00–08:50, report gap until the 12:45 owner demand) · **Companion:** the concurrent
sibling session (its 12-46 report: `docs/status/2026-11-06_12-46_v13-execution-session-p9a-guards-verify-blocks.md`)
— division of labor settled in practice: I owned the preset/release/CI-repair lane, it
owned the v13 program lanes (P9a verify, R-train, G-CNT, S1, A1). Attribution below is mine.

**Scope line — VERIFIED this session, with the artifact:**
master + `v0.10.0` + `preset/v0.1.0` ON ORIGIN (ls-remote); CI on the final commits green
except the designed `ci-preflight` SSH gate; preset suite green ×9 consecutive
(3 rounds × `-race -count=3`) + lint 0; config suite green + lint 0 (cyclop fix);
cqrs-lint exit 0 (F032 suppressed with reason, E014 re-anchored); example-smoke preset
leg green; insteadOf git rewrite unset; both consumer heads-ups refreshed.

---

## a) FULLY DONE (this session, verified)

1. **Plan split-brain resolved** (commit `1c98b17`, pushed): the sibling's 07-10 plan
   canonical, my racing 07-46 duplicate deleted, TODO_LIST pointer reverted, dated
   erratum appended. The owner's standing directive was read as authorization; the
   one-writer rule (git log + mtimes before every edit/commit) held for the rest of
   the session — two mid-flight collisions were caught by mtime guards, zero lost work.
2. **`v0.10.0` + master pushed** (08:05, clean fast-forward) — the held push from
   session 5; CI verified on the push (found the standing reds, see a4/d).
3. **insteadOf VCS rewrite unset** + preset verified to resolve hermetically through the
   real proxy against the pushed tag.
4. **THE HANG — root-caused and closed.** The intermittent 5s `listener_close` shutdown
   hang (the preset ship-blocker from sessions 4–5): **test hygiene, not a product bug.**
   The old spine/otel tests closed response bodies WITHOUT draining; abandoned bodies
   tear conns down mid-response, the server-side conn goroutine caught reading the
   "next" request sits in the ACTIVE state `http.Server.Shutdown` deliberately skips,
   and under scheduling load the 5s stop-timeout expires. Method: 10-iteration isolated
   harness (0 repro) → compiled-binary + SIGQUIT watcher with live stall detection
   (0 repro across 18 suite runs, quiet AND load-32 + 24 artificial spinners) → diff of
   the pre/post-refactor tests pinned the mechanism (close-without-drain →
   drain-then-close; the sibling's noctx refactor was the accidental fix). Evidence
   recorded in the 07-10 plan erratum + TODO_LIST row (commits `6fd1de8` and prior).
5. **Standing CI reds triaged and repaired** (all-day reds nobody had owned):
   - **config cyclop 14 > 12** (`collectFields`, shipped red in v0.1.0's fleet-lint):
     refactored to guard-clause style + `nestedWalk` helper, behavior-identical;
     suite green, lint 0, CHANGELOG entry.
   - **cqrs-lint F032** (upstream catalog-mandate rule, retro-red via CI's `@latest`
     install): suppressed inline at the `command/v4` import with the recorded reason
     (framework wrapper owns no domain; consumers derive their own catalogs) — inline
     suppression follows the repo's established pattern, NO invented API. The stale
     E014 suppression re-anchored to the package clause in the same edit.
   - **example-smoke preset entry**: the sibling had fixed the probe path
     (`/health/ready` → `/readyz`); I verified the leg GREEN end-to-end.
6. **P7 finished** (the sibling started it; I completed/verified): preset lint
   **0 issues**, README truth pass (dashboard owns the health surface; hooks
   passthrough consumer-first documented), quick-start snippet compile-checked
   in-module, `bootFail` helper extracted to clear the art-dupl gate's 3× clone
   finding (with two now-unneeded wrapcheck nolints dropped).
7. **P9 preset v0.1.0 release** (tag final at `21c1a78`; pushed to origin — the push
   itself was flushed during the session gap, tag verified on origin by ls-remote):
   CHANGELOG dated with the full shipped surface; API surface recorded
   (`go doc -all`, 175 lines); release gate 9× green BEFORE tagging; pre-tag-checks
   run three times (art-dupl + pin-drift findings fixed in-loop); release-state sync:
   AGENTS (thirteen released + preset in Latest-per-module), FEATURES preset section
   (+ the stale config push-hold note corrected), ci.yml proxy-smoke comment
   (matrix already had preset).
8. **Consumer heads-ups refreshed** (`docs/feedback/outgoing/2026-11-03_consumer-heads-up-servertiming-{cqrs-htmx,papdashboard}.md`):
   both rewritten to the tagged reality (v0.10.0 shipped facts) + the preset
   announcement, PapDashboard framed with preset as its boilerplate upgrade path.

## b) PARTIALLY DONE

1. **P9's pre-tag-checks final pass**: my third run failed on MY OWN AGENTS
   Latest-line format bug (prose inside the parseable list); the sibling fixed the
   format (`4c14899`) and the tag reached origin — but the FULL gate suite has not
   been re-run by me against the pushed tag state (gosum-canonical + consumer-claims
   legs especially). Residual risk: low (the failure was the line format alone, and
   the tag's go.mod checks passed), but the ritual says re-run.

## c) NOT STARTED (my lane)

1. **P8: integration E2E through PUBLISHED preset tags** — designed, not written:
   `integration/preset_e2e_test.go` booting a preset-built service with CODE-config
   legs (no env — integration tests run parallel), spine probes + realtime
   stream + security header + graceful stop, bodies drained everywhere (the fresh
   house rule); `integration/go.mod` += preset require + `documentedPins` fixture
   + pin-drift green; run with the README's `GOWORK=off` prefix.
2. **Fresh-consumer proxy checks** for `v0.10.0`, `config/v0.1.0`, `preset/v0.1.0`
   per `docs/recipes/fresh-consumer-proxy-check.md` (isolated GOMODCACHE).
3. **CI verify on the tag commit's proxy-smoke preset leg** — first live run against
   the published tag (the leg was born-green on the untagged module resolution).

## d) TOTALLY FUCKED UP (own mistakes)

1. **The pipe trap, again, documented-by-my-own-AGENTS**: `cqrs-lint . | rg -v INFO |
   head -15` hid the F032 ERROR below the cutoff and `echo exit=$?` printed `head`'s
   exit 0 — I briefly concluded "local doesn't fire F032" and spent three tool calls
   researching a version skew that did not exist. The AGENTS rule exists precisely
   for this; I re-learned it live.
2. **AGENTS Latest-line format bug**: I appended a parenthetical with prose + "08:5x"
   into the machine-parsed release line → pin-drift "15 entries, only 12 parse". The
   sibling repaired it; the guard worked as designed.
3. **Tagged before the gates were fully green** (tag → gate → fix → move tag → gate →
   sync → move tag again). The ritual order is gates-green THEN tag; three tag
   creations for one release wasted a cycle and briefly risked a daemon-commit-swept
   intermediate as the tagged tree (caught: verified the final tag content).
4. **The status-report demand sat unanswered through the session gap** (~08:50→12:45):
   the owner asked for the report at the pause; the execute directive arrived before
   I wrote it. Delivered now — late.

## e) WHAT WE SHOULD IMPROVE

1. **Release-state lines are CODE**: write them parser-clean (plain `name vX.Y.Z`
   commas only, prose OUTSIDE the parseable span) and run `check-pin-drift.sh`
   immediately after editing — not at gate time.
2. **Tag last**: complete every gate (incl. art-dupl + pin-drift) on the working
   tree, THEN create the tag once. The gate-order failure mode is cheap to avoid.
3. **The drain-bodies house rule should be written into AGENTS' testing section**
   (close-without-drain = RST generator = shutdown-flake factory; the sibling applied
   it to integration's security/errorpages E2E in `33f1152` — the lesson is
   circulating, make it durable).
4. **Guard/lint output: file + rg, never pipes** — I owe this rule compliance more
   than anyone (two incidents this session across sessions 5–6).
5. **Concurrent-session coordination worked**: lane ownership (preset/release vs
   program) + mtime re-reads + never staging foreign files got two full trains
   through one tree with zero clobbered work. Keep that protocol.

## f) NEXT — up to 50, in priority order (my lane first, then observed program state)

1. P8 integration E2E through published preset/v0.1.0 (design in c1).
2. `documentedPins` fixture += preset v0.1.0; pin-drift green.
3. Fresh-consumer proxy checks: v0.10.0, config v0.1.0, preset v0.1.0.
4. Re-run full `pre-tag-checks.sh preset/v0.1.0` against the pushed state (close b1).
5. Verify CI on the final commit: every job incl. proxy-smoke preset leg green.
6. AGENTS testing section: add the drain-bodies rule (if the sibling hasn't).
7. Post the two consumer heads-ups (or hand to owner — see g1).
8. TODO_LIST preset row → released state (partially synced by the sibling — verify).
9. cqrs-lint upstream memo: F032 and framework wrappers (appkit/cqrs consumers never
   import go-cqrs-lite directly, so the mandate never reaches them — either an
   upstream framework-tier exemption or a doc note; my inline suppression records
   the local position).
10. Pin cqrs-lint in CI (the `@latest` install made CI retro-red twice — the
    art-dupl-gate pin precedent applies; needs upstream's next stable tag).
11. pkg.go.dev render check for preset once the proxy catches up (LICENSE present).
12. S-train execution: scaffold template repo per the sibling's S1 verdict
    (`2026-11-06_08-45_scaffold-shape-verdict.md`).
13. A2–A5 swap recipes on the A1 template.
14. D1 worker-supervisor battery; D3 sqlitekit (07-10 plan).
15. C2R: realtime-folded preset documentation row.
16. R5a + endpoints.md H8 feed (input now exists: `docs/review/libs/endpoints.md`).
17. H1–H7 httpx core program (B5 full module program per the questionnaire).
18. X1–X4 compose-on-httputil; Y/Z backlog hygiene (07-10 plan).
19. Final `scripts/verify-all.sh` over 13 modules + integration; session-close report.
20. ST bench landed at ~15µs/op (sibling's 12-46 report) — fold into AGENTS perf
    baselines if not already.
21. The stale-proxy-404 duration lesson (`45cfacf`) — read before ANY future re-tag.
22. G-CNT: re-run the adoption counter after preset's announcement window; the
    six-month gate is now a command.

## g) QUESTIONS FOR THE OWNER (cannot be answered from here)

1. **Post the two consumer heads-ups?** Both drafts are final in
   `docs/feedback/outgoing/` (cqrs-htmx issue + PapDashboard issue, your voice,
   review-before-posting flag) — should I file them as GitHub issues now, or do you
   want to read/edit first?
2. **Integration pin timing:** pin preset v0.1.0 in integration NOW (P8 rides the
   pin charter — published-tags-only), or hold until a potential preset v0.1.1
   accumulates? Pinning now couples integration to the freshest tag; holding leaves
   P8 open.
3. **Lane ownership going forward:** the de-facto split (this session: preset,
   releases, integration, consumers; sibling: v13 program S/A/D/H/X trains) worked —
   keep it explicitly for the remaining v13 execution, or re-merge into one session?

---

**Verification state at write time:** master = `9ffff2f` (sibling's docs close-out;
master == origin as of its 12-46 claim, re-verified ls-remote for tags only here);
tags on origin: v0.10.0, config/v0.1.0, preset/v0.1.0, health/v0.3.1 (sibling's
issue-54 train); CI latest run red ONLY on the designed ci-preflight SSH gate; my
uncommitted tree: nothing (both mid-flight AGENTS edit attempts were rejected by
mtime guards and abandoned in favor of the sibling's fixes).
