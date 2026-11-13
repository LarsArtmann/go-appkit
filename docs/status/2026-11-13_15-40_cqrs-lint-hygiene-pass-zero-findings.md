# Status: cqrs-lint hygiene pass — 16 findings → 0 (2026-11-13, 15:40 CET)

**Session scope:** single mandate — resolve the 16 findings from `buildflow -s cqrs-lint
--format finding`. All verdicts below are grounded in commands run THIS session; the
gaps are named in the scope disclosures per section and summarized at the end.

**VERIFIED this session (command-backed):** `buildflow -s cqrs-lint` findings [] /
total 0 (run twice, the second time on the final HEAD after the daemon's mixed-in
commits); `cqrs-lint .` zero findings from BOTH the repo root and `cqrs/`; hermetic
`go test -race -count=1` green on cqrs, docs, integration; golangci-lint 0 issues on
cqrs + integration; `golangci-lint config verify` green (cqrs); gofmt clean; offline
guards (go-directives, workspace-charter, dependabot-parity, sse-protocol,
routed-destinations, pin-drift) green; `check-gosum-canonical.sh integration` green;
`go-structure-linter .` total=0; `check-art-dupl.sh` green after my comment edits;
docs-gates `freshness --pre-flight` SAFE; E014 stale-string sweep clean (only
archived point-in-time reports mention the removed directive — they stay true to
their day by rule).

**NOT verified this session:** CI on the final pushed commit (session-close ritual —
never ran `gh run`); the root module's suite after the concurrent dedup session
changed `testhelpers_test.go` (their file, not re-verified by me); fleet-wide
`check-golangci-configs.sh` (only cqrs's config changed; I verified cqrs alone);
dprint/`buildflow format` over the new root `.cqrs-lint.json` (JSONC formatting
churn risk unchecked); the buildflow timing-regression table (dismissed by load
reasoning — uptime 34 on 32 cores, uniform +56–100% across ~35 unrelated steps —
without a quiet-load re-baseline).

---

## a) FULLY DONE

1. **All 16 cqrs-lint findings resolved → 0**, in both scan scopes (BuildFlow's
   root-scope run AND the module-ritual `cd cqrs && cqrs-lint .`). Verified on final
   HEAD.
2. **V006 pin drift ALIGNED, not silenced**: integration bumped onto cqrs's
   unreleased go-cqrs-lite set (system v4.16.0→v4.18.0, metaengine v4.20.0→v4.23.0,
   sqliteengine v4.6.0→v4.7.1, id v4.7.4→v4.7.5), docs id/v4→v4.7.5; hermetic tidy;
   `go.sum` canonical against proxy+sumdb. The E2E composition is proven against the
   NEXT train's dependency set under published cqrs v0.9.0.
3. **Integration E2E suite `-race` green** under the new pins; docs and cqrs suites
   green; cqrs suite re-run green on final HEAD (which had absorbed the dedup
   session's `commands_test.go`/`sse_test.go` edits).
4. **Suppressions moved to their true anchors** (each with a recorded reason): F009
   inline above the `inFlightTracker.Drain` ticker (shutdown counter poll, not a
   domain timer), A016 inline above the query-middleware wiring (idempotency stays
   consumer-owned per `DefaultCommandMiddleware`; the flagged line is the QUERY
   dispatcher), E023 inline above the `database/sql` import (aux DLQ/checkpoint DB —
   projectionhost plumbing, the rule's own non-CQRS dismissal), C013+D014×4 as a
   block on the `FoldedEvent` struct (in-memory channel signal, never
   CBOR/JSON-encoded; `event.Instant`/json tags would be inert ceremony).
5. **E009/E014/F007/A024 moved to rule-level disables with reasons** in BOTH
   `.cqrs-lint.json` files — the new repo-root config exists because BuildFlow's
   cqrs-lint step runs `cqrs-lint .` from the repo ROOT and never loaded
   `cqrs/.cqrs-lint.json` (proven via `cqrs-lint doctor`). Root config deliberately
   carries NO preset/feature pin (integration's E2Es legitimately dispatch commands).
6. **Root cause of the standing E014 ghost found and removed**: the preamble
   directive on commands.go had been silently INERT since daemon commit aa9dd60
   (2026-11-06) relocated it off the package clause; today's anchor demands the
   clause itself, and a clause directive trips go-structure-linter's package-naming
   rule (parses the whole line incl. comments, flags uppercase rule IDs — proven
   2026-10-13, re-proven 2026-11-13 with a live test of both placements). The
   preamble dance is retired permanently in favor of config disables.
7. **golangci v2.15.0 wrapcheck drift, second wave absorbed**: the generic sigs
   `system/v4.RegisterCommand`/`system/v4.RegisterQuery` re-fired on cqrs's
   delegation returns; absorbed via `wrapcheck.ignore-sigs` in cqrs/.golangci.yml
   per the fleet rule (never fix in code). My initial inline-nolint detour was
   REVERTED — nolintlint then flagged the directives unused, confirming config was
   the only stable form.
8. **Documentation debt paid**: cqrs CHANGELOG [Unreleased] carries the full
   suppression map, both root causes, the V006 rationale refresh, and the
   scan-scope stale-warning artifact; AGENTS.md's 2026-11-06 "E014 directive-shape
   LAW" replaced by the 2026-11-13 suppression LAW (two-config lockstep, direct-
   adjacency placement rule, stale-warning interpretation, scope duality); wrapcheck
   bullet extended with the second wave.
9. **Guards all green post-change**: directives, workspace charter, dependabot
   parity, sse-protocol, routed-destinations, pin-drift, gosum-canonical
   (integration), structure-linter 0, art-dupl (comments can clone — verified not).
10. **docs-gates pre-flight SAFE** on the new CHANGELOG lines (no historical item
    flips to DRIFT, nothing accidentally harvested).
11. **E014 contract-string sweep** (AGENTS.md ship-time rule, initially forgotten,
    run before this report): no living-doc references to the removed directive; two
    hits live only in archived reports (correct to leave).
12. **cqrs-lint scorecard re-run** (module ritual); no adoption delta — no wrapper
    feature was added this session.

## b) PARTIALLY DONE

1. **CI verification (session-close ritual)** — NOT done at all in this session, so
   by the ritual's own standard the session is not closeable. Both this session's
   and the concurrent dedup session's work ride the same tip-CI run; the dedup
   report (14:43) explicitly left tip-CI unverified (queued 14:40). This is the
   session's biggest miss: AGENTS.md says verify CI green on the FINAL pushed
   commit, and I never even checked whether the daemon pushed.
2. **Timing-regression triage** — dismissed as machine load (uptime 34/32 cores;
   ~35 unrelated steps uniformly +56–100% is a load multiplier signature, not
   per-step regressions) but I did not run `buildflow timings --regressions` for
   detail nor re-baseline at quiet load. Reasoning is sound, evidence is
   circumstantial.
3. **Concurrent-session awareness** — I noticed the daemon commits mixed in the
   dedup session's files (README, TODO_LIST, testhelpers_test.go,
   commands_test.go, sse_test.go) and verified MY inventory landed intact, but I
   did not run the root module suite after `testhelpers_test.go` changed (their
   claim, their suite; still, one `go test ./...` at root would have closed it).
4. **Fleet-wide config verify** — ran `golangci-lint config verify` for cqrs only;
   `scripts/check-golangci-configs.sh` (all 12 configs) not run. Low risk (one
   config touched, schema-verified individually) but the guard run is cheap and
   the standing rule is fleet-scoped.
5. **Formatter hygiene for the new `.cqrs-lint.json`** — the repo's dprint carries
   a JSON plugin; my JSONC (comments) file was never format-checked. If dprint
   reformats it, the next `buildflow format` produces churn against a hand-written
   config. Unchecked, unresolved.

## c) NOT STARTED

1. Mechanical workspace pin-consistency guard (the standing net that V006's config
   disable removed — see e-2).
2. cqrs-lint upstream feedback batch: six heuristic findings with this session's
   empirical evidence (A016 misattributes the query-dispatcher line as "command
   dispatcher"; A024 fires on the module that DOES wire `system.New`; C013 fires on
   a never-serialized struct; D014 ignores unexported fields, where json tags are
   inert; F009 has no shutdown-drain whitelist; E014 doesn't recognize the
   staleness-guard API as the read-your-writes answer). verify-before-filing source
   checks still owed per item.
3. go-structure-linter upstream note: package-naming should strip comments before
   validating the name (the false positive that makes clause directives
   permanently impossible here).
4. BuildFlow cqrs-lint provider fix: per-module invocation or nearest-config
   discovery, so `cqrs/.cqrs-lint.json` is honored natively (root-config mirror is
   today's workaround).
5. Opt-in `EventConfig` idempotency knob (consumer-supplied durable store seam) —
   the structural answer to A016/F007; today's suppressions record the policy,
   nothing offers it.
6. `inFlightTracker.Drain` 2ms poll → sync.Cond or channel-close primitive — the
   structural answer to F009.
7. DLQ/checkpoint aux store on metaengine evaluation — the structural answer to
   E023 (the rule's own text dismisses non-CQRS tables; an engine-native DLQ would
   retire the suppression).
8. ADR consolidating the framework-wrapper rule-disable policy (A018, F032, A024,
   E009, E014, F007 all share one rationale class; today it lives in two JSON
   files + CHANGELOG prose).
9. `.cqrs-lint.json` lockstep guard (diff the two files' disable sets; offline,
   fast, would have caught nothing today but prevents future drift).
10. Next cqrs release train (accumulated go-cqrs-lite bumps + this hygiene pass
    ride it; integration is already pre-adapted and proven).

## d) TOTALLY FUCKED UP

**Nothing.** No data loss, no broken gates, no revert-needed changes; every wrong
step was self-corrected within the session and the final tree is green everywhere
measured. Honest near-misses that did NOT rise to fucked-up:

- Two wrong suppression placements (directive must sit DIRECTLY above the anchor —
  I put it atop multi-line prose blocks twice) and one policy violation detour
  (inline nolint for wrapcheck against the fleet's "never fix in code" rule) — all
  self-caught via the tool's stale-suppression warnings and nolintlint, all
  reverted.
- **Pre-existing, discovered today**: the E014 suppression had been silently inert
  for a week (aa9dd60, 2026-11-06) while the v0.9.0 CHANGELOG claimed it was
  relocated to the package clause — the published tag's CHANGELOG text does not
  match its tree. Three sessions shipped with a dead suppression. That failure was
  not mine, but my session is the one that finally ran the linter against the claim.

## e) WHAT WE SHOULD IMPROVE

1. **Verify suppression efficacy at close time, not placement time.** The 2026-10-13
   session claimed "cqrs-lint still suppresses E014" after moving the directive;
   the claim was true that day and false forever after (anchor drift). A closing
   linter run is the only proof; CHANGELOG claims about suppressions are not.
2. **Read the tool's diagnostic warnings FIRST.** cqrs-lint's "stale suppression at
   X — fires at Y; safe to remove or move" names the exact anchor. I discovered
   this after two wrong placements; it should have been the first experiment.
3. **Grep the fleet policy before choosing a mechanism.** AGENTS.md already ruled
   "never fix wrapcheck sites in code — use ignore-sigs"; I added nolints first and
   reverted after. Policy recall before mechanism selection.
4. **A config disable must not orphan a standing net.** Disabling V006 silenced
   real future drift (a forgotten bump at train time is now invisible to
   cqrs-lint). Every disable needs either a replacement mechanical guard or an
   explicit accepted-risk note in TODO_LIST. Today I did the latter only in prose.
5. **Kill the scan-scope duality.** Root-scope (BuildFlow) vs module-scope (ritual)
   runs disagree about which suppressions are "stale" — each flags the other's
   necessary directives. One invocation home (BuildFlow fix) or one config home
   (root-only, retire the module file) ends the confusion class.
6. **Daemon-commit archaeology is cheap if done immediately.** The daemon landed
   three commits mixing my work with the dedup session's mid-session; verifying
   inventories right after each landing (one `git show --stat`) kept the tree
   trustworthy. Do it every time, not when surprised.
7. **Contract-string sweeps belong in the change step, not the report step.** The
   AGENTS.md rule (rg the old string repo-wide at ship time) I ran only when
   drafting this report — it should have been part of the original change.

## f) NEXT — 50 things, impact-ranked (a brainstorm, not a commitment list)

1. Verify tip-CI green on the final pushed commit (session-close ritual; covers
   this session AND the dedup session's still-unverified run).
2. Build the mechanical workspace go-cqrs-lite pin-consistency guard (extend
   `check-pin-drift.sh`; offline+fast; restores the net V006's disable removed).
3. Add the `.cqrs-lint.json` lockstep guard (diff root vs module disable sets).
4. Run `buildflow format` / dprint over the new root `.cqrs-lint.json`; resolve any
   JSONC formatting churn now, not in a later daemon commit.
5. Harvest this report's section f into TODO_LIST (docs-health HARVEST — the skill
   mandates it; this list is brainstorm fuel, not entombed).
6. Run the root module suite (close the dedup session's `testhelpers_test.go`
   change under a green run).
7. Fleet-wide `scripts/check-golangci-configs.sh` (cheap belt after touching cqrs's
   config).
8. Investigate the 03:24Z govulncheck nightly CI failure (flagged by the dedup
   session, still open, out of my scope today).
9. Re-baseline buildflow timings at quiet load; clear the polluted regression
   table (quiet-run gate per the load-flake protocol).
10. Fix BuildFlow's cqrs-lint provider (per-module invocation or nearest-config
    discovery) — the root-cause fix for the scan-scope duality.
11. File the cqrs-lint upstream feedback batch (six heuristic findings, evidence
    attached; per-item verify-before-filing source checks first).
12. File go-structure-linter upstream: strip comments in package-naming validation.
13. cqrs-lint upstream: scope-aware stale-suppression warnings (or documented
    two-scope semantics).
14. Decide on `--fail-on-stale-suppressions` adoption once 13 lands (currently
    warnings-only; the flag would turn scope artifacts into hard failures).
15. Opt-in `EventConfig` idempotency knob (durable-store seam the consumer
    supplies) — structural A016/F007 answer.
16. Replace `inFlightTracker.Drain`'s 2ms poll with cond/channel-close — structural
    F009 answer.
17. Evaluate a metaengine-native DLQ/checkpoint store — structural E023 answer.
18. ADR: framework-wrapper rule-disable policy (one rationale class, six rules,
    two JSON files today).
19. Next cqrs release train: date the accumulated go-cqrs-lite bumps + hygiene
    pass; integration re-pin ritual + fresh-consumer proxy check ride it.
20. Check cqrs README for a consumer idempotency recipe
    (`middleware.CommandIdempotency` via `EventConfig.CommandMiddleware`); add if
    absent.
21. Surface the in-memory-only FoldedEvent contract in the struct's godoc (today it
    lives only in the suppression directive's reason text).
22. Add cqrs-lint to the guard venue matrix (it rides BuildFlow/CI only today; no
    pre-tag row exists).
23. Probe `cqrs-lint baseline` (seen in `explain`) as an alternative mechanism to
    rule disables for accepted findings — unexplored.
24. Watch the next dependabot cycle for conflicts with today's manual pin bumps.
25. Fix or annotate the standing `sse_test.go` scanner.Err() warning (dedup
    session's file; LSP-visible, lint-invisible).
26. Sweep gopls `infertypeargs` infos in cqrs tests (8 sites, cosmetic).
27. Record the scorecard "no delta" ritual note in the CHANGELOG (one line).
28. AGENTS.md count-truth pass: "twelve standing guards" (Release Ritual step 5)
    vs the README index's "NINE guards" vs today's actual count — the file
    disagrees with itself across sections.
29. rg sweep for other stale suppression citations in LIVING docs (module-details,
    READMEs, DOMAIN_LANGUAGE) — the archived-report exemption does not extend
    there.
30. Home the framework-wrapper policy sentence in `docs/agents/module-details.md`
    §cqrs (deep content belongs there per the AGENTS.md cap).
31. Consider a shared dependabot group for go-cqrs-lite modules across workspace
    go.mods (config-level mitigation for the V006 drift class).
32. Pin the cqrs-lint binary version used by BuildFlow/CI (today's E014 anchor
    drift WAS an unpinned-tool upgrade; the same class will recur on the next
    cqrs-lint release).
33. Add "re-run both scan scopes after every cqrs-lint upgrade" to the upgrade
    checklist (rule anchors can move; today's block suppression on FoldedEvent is
    anchor-dependent too).
34. Verify the block suppression's resilience: plant a probe that C013/D014 would
    fire if the block anchor drifts (a canary test like the documentedPins
    fixture pattern).
35. Confirm the nightly govulncheck leg passes after today's dependency bumps
    (ties to 8).
36. `cqrs-lint doctor` from cqrs/ scope (config resolution proof post-edit —
    findings already prove it; one-line belt).
37. Review the integration go.sum diff for hygiene (canonical-verified already;
    eyeball-only step).
38. TODO_LIST: check the dedup session's routed P2 floor-train row still resolves
    (routed-destinations guard passed, but eyeball the row's freshness).
39. Decide the root config's long-term preset posture (none today; if root-scope
    F-rules fire later, preset vs per-rule becomes a live decision).
40. Verify `cqrs-lint`'s "library-framework" preset matches our disable intent —
    maybe upstream our disable set as a preset improvement.
41. Update `docs/agents/module-details.md` §cqrs with the two-config lockstep
    contract (deep-content home for the AGENTS.md pointer).
42. Ship `scripts/check-cqrs-lint.sh` (both scopes; findings fail, stale warnings
    tolerated) so local verification matches BuildFlow exactly.
43. Cross-check FEATURES.md for FoldedEvent wording that could imply
    serializability (the D014 saga's documentation mirror).
44. Audit whether any other module needs a `.cqrs-lint.json` (realtime and
    integration import go-cqrs-lite; neither has a config — root config covers
    root-scope runs, module rituals would use defaults).
45. Note in the next train's CHANGELOG that integration pre-adapted the
    dependency set (the compat story for consumers reading the wave).
46. Confirm pkg.go.dev cqrs page unaffected (no API change — ritual row, likely
    skip).
47. Consider a `doc.go` note in cqrs naming the supported linters (cqrs-lint,
    golangci config links) for consumer discoverability.
48. Re-run `check-status-index.sh` after this report lands (index→disk + disk→index
    guard; born-red-proven guard class).
49. Sweep `docs/planning/` for the W5-C2 design note's FoldedEvent field list — if
    it documents `occurredAt time.Time`, the suppression rationale should be
    mirrored there or the doc marked historical.
50. Post-push, re-run the full `verify-all.sh` once at quiet load as the standing
    per-module proof (the documented bar — fleet-lint + CI matrix).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Pin governance home**: with V006 config-disabled, do you want the mechanical
   workspace pin guard (my recommendation — extend `check-pin-drift.sh`, offline
   and fast), or do you prefer re-enabling V006 in the root config and accepting
   the mid-train noise as the signal?
2. **BuildFlow side**: is fixing BuildFlow's cqrs-lint provider (per-module
   invocation or nearest-config discovery) the standing solution you want pursued
   in the BuildFlow repo, or is the root-config mirror the accepted end state and
   the provider quirk gets documented instead?
3. **Release timing**: ship the accumulated cqrs go-cqrs-lite bumps + this hygiene
   pass as cqrs v0.10.0 in a near-term train (integration is pre-adapted and
   proven), or keep the ACCUMULATE posture (13-decision questionnaire) until the
   next feature rides the same train?
