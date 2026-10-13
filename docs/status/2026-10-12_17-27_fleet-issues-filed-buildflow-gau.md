# Status Report — Fleet Issues Filed: BuildFlow ×4 + go-auto-upgrade ×1

**When:** 2026-10-12 17:27 CEST
**Scope:** the follow-up order "Filed all???? If not DO IT" — verify every fleet-wide BuildFlow candidate at source level, then file. Supersedes the filing-status section of `2026-10-12_15-53_buildflow-gate-triage-agents-slimdown.md`; that report holds the full triage narrative (gate 69→0, AGENTS.md slim-down, leak #8).

---

## Headline

**5 issues filed and verified OPEN** (all source-verified in the local repos at BuildFlow `8a349283a`, drafts voice-checked 0 FAIL / 0 WARN, filed via `gh issue create --body-file`, landing confirmed):

| Issue                                                                        | Repo            | One-line                                                                                                                                           |
| ---------------------------------------------------------------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| [#46](https://github.com/LarsArtmann/BuildFlow/issues/46)                    | BuildFlow       | go-work-sync re-adds deliberately excluded modules — use-directive drift ignores `go_mod_ignore_dirs` (#18's wiring stopped at the go.mod checker) |
| [#47](https://github.com/LarsArtmann/BuildFlow/issues/47)                    | BuildFlow       | result cache replays stale verdicts after tool-config edits (`.jscpd.json` is not key material)                                                    |
| [#48](https://github.com/LarsArtmann/BuildFlow/issues/48)                    | BuildFlow       | jscpd: ≥50-line config-file clones gate-block by default; user `.jscpd.json` replaces curated defaults instead of merging                          |
| [#49](https://github.com/LarsArtmann/BuildFlow/issues/49)                    | BuildFlow       | lychee preflight: authenticate-vs-exclude policy (survey g1) still open — every private-repo project warns every run                               |
| [go-auto-upgrade#3](https://github.com/LarsArtmann/go-auto-upgrade/issues/3) | go-auto-upgrade | no per-migrator opt-out: stdlib2lo suggestions unacknowledgeable for repos deferring lo adoption                                                   |

Drafts preserved at `docs/drafts/buildflow-*.md` + `docs/drafts/go-auto-upgrade-migrator-optout.md`.

## What verification changed about the plan (the part that mattered)

1. **One candidate DIED at Gate 2** (would have been a wrong issue): "auto-configure emits schema-invalid exhaustruct settings" — grep of `golangci_auto_configure_sdk.go` found zero exhaustruct references; git blame pins the dead blocks to the hand-run 2026-09-04 migration (`723a33f`), not the tool. Not filed.
2. **One candidate RE-HOMED**: the lo opt-out's root cause lives in go-auto-upgrade (single toolsdk Option `blame`; `sdk.MigrateFiles` takes no migrator selection) — filed there, not at BuildFlow.
3. **One candidate SHARPENED twice**: (a) the go.work `ignore`-directive machinery already exists (`Checker.IgnoreDirs`, `NewCheckerWithIgnoreDirs`) — the issue is precisely that `NewWorkChecker()` and `DetectGoWorkDrift` never consume it; (b) BuildFlow#28/#40 shipped a findings-gate baseline, so "gomod-check has no ack path" was wrong and folded away — the surviving ask is the ignore plumbing. #29's replace-guard (same silent-destruction class, replaces protected) became the precedent citation.
4. **One earlier claim of mine was corrected mid-verification**: "NewCheckerWithIgnoreDirs has no production callers" was a truncated-grep artifact — `sdk_checkers.go:168` calls it. The conclusion flipped from "dead code" to "wired for the wrong half".

## a) FULLY DONE (this follow-up)

- All 5 filings: verified → drafted (in-repo, never /tmp) → check-draft clean → filed → landing verified via `gh issue view` (all OPEN).
- Duplicate sweep: `gh issue list` full scan of both repos + targeted searches — no existing duplicates (closest: #18/#23/#28/#29, all cited as precedent, none cover the asks).
- TODO_LIST updated: leak row now carries BuildFlow#46 + the "drop skip_steps when #46 lands" trigger; samber/lo row carries go-auto-upgrade#3.

## b) PARTIALLY DONE

- The exhaustruct_v5 empty-settings-schema filing is NOT done — it belongs at **golangci-lint upstream**, a different repo, and its verify pass (schema provenance: bug vs intentional no-settings) hasn't run.
- The 15-53 report's section-f harvest into TODO_LIST/ROADMAP is still pending (docs-health HARVEST).

## c) NOT STARTED

- Everything in the 15-53 report's section (f) that isn't already a TODO_LIST row (1.27.2 floor train, release trains #25/#26, root `.golangci.yml` staleness, SSH secret P1, …).
- Watching the four BuildFlow issues for fixes; each fix retires a consumer workaround (#46 → drop skip_steps; #48 → maybe drop the `.golangci.yml` ignore if defaults ship it; #47 → config edits invalidate honestly).

## d) TOTALLY FUCKED UP

Nothing. One process catch worth naming: my "dead code" claim nearly became a wrong issue — the truncating `head -5` on a grep. The Gate-2 re-check (what wires ignore dirs TODAY?) caught it before filing. That is exactly the failure mode verify-before-filing exists for.

## e) WHAT WE SHOULD IMPROVE

1. **Never trust a truncated grep** — the checker run nearly inherited a false premise from my own earlier output. Re-derive conclusions from full output at filing time (the line-number-freshness rule, applied to claims).
2. **File-first bias is dangerous even in own repos** — 1 of 7 candidates was a wrong premise, 1 belonged in a different repo. The 5 that survived are stronger for the 2 that didn't.
3. Drafts-in-repo worked (skill rule honored); keep it.
4. go-auto-upgrade has an empty issue tracker — findings there will keep landing "as new"; consider a `known-limitations` doc so opt-out-adjacent questions have a landing page.

## f) NEXT (top items, merged with 15-53 §f; ★ = new from this follow-up)

1. ★ Verify + file the exhaustruct_v5 schema issue at golangci-lint (the one remaining candidate).
2. ★ When BuildFlow#46 lands: drop the `go-work-sync` skip_steps, declare `./integration` via the new config mechanism, re-run charter guard.
3. ★ When BuildFlow#48 lands: reconsider `.jscpd.json` (defaults may cover `.golangci.yml`).
4. ~~Root `.golangci.yml` hygiene: `run.go: 1.26.7` + retired `goexperiment.*` build-tags (still untouched).~~ done — swept on the 1.27.2 floor train
5. ~~1.27.2 floor train (unblocks cqrs-htmx v4.14.0 pin; clears govulncheck GO-2026-6617). USER-gated.~~ done — wave 1 (12 go.mods + go.work + fixture)
6. ~~Release trains #25/#26 (code-complete; tag/push USER-gated).~~ done — core v0.9.0 + health v0.2.0 tagged/pushed; issues closed
7. P1: restore `SSH_PRIVATE_KEY` secret + one green master run. USER-gated.
8. samber/lo adoption train (go-auto-upgrade#3 may silence the noise until then). **→ decided test-only 2026-10-12; train pending go-auto-upgrade #3**
9. ~~go-cqrs-lite v4.13.0 cookbook ritual + pin sweep (DUE).~~ done — cqrs v0.8.0 (cookbook re-verified + sweep)
10. ~~Harvest 15-53 §f into TODO_LIST/ROADMAP (docs-health).~~ done — this pass (2026-10-13)
11. ~~Index both of today's reports in `docs/status/README.md`.~~ done — this pass (five rows 15-12 through 20-36)
12. One `buildflow --verbose` run: name the 9 unavailable tools, audit the 67 not-applicable.
13. ~~Enumerate the truncated `+2` cqrs-lint findings from a file, not a pipe.~~ done — cqrs-lint exits 0 at cqrs v0.8.0
14. Charter/directives guard coverage for daemon (pma) commits.
15. `journalctl -t pma` around leak windows; name the leaks #1–7 mutator.
16. Fleet sweep: did go-work-sync re-adds hit sibling repos before #46?
17. Decide `.jscpd.json` ignore strategy (path- vs format-scoped) independent of #48's outcome.
18. Consider `golangci-lint config verify` as a CI job.
19. Browser CSP pass over hardened health dashboard (pre-existing P2).
20. errorpages `statusRecorder` → `httputil.ResponseRecorder` (USER-gated).
21. go-health sentinel ask filing (USER-gated).
22. `NewServerListener` ask (USER-gated).
23. fr cooldown-combinator ask filing (USER-gated).
24. GracefulClose + timers upstream flip watch. **→ TODO row — timers half closed at cqrs v0.8.0; GracefulClose watch open**
25. ~~fr v0.2.1 sweep on next trains.~~ done — fr v0.2.1 in cqrs/frh/otel go.mods (2026-10-12 trains)
26. ~~Example StartHooks migration (trigger met).~~ done — decc462
27. ~~health follow-through after the #25 train.~~ done — health v0.2.0 + evidence in its CHANGELOG
28. pkg.go.dev flightrecorder 404 escalation.
29. ~~otel polish batch.~~ done — 104f28c (TODO row struck)
30. Composition-gap E2E watchlist.
31. Battery waves W3–W5 (demand-gated).
32. Full cqrs-htmx setup hermetic run.
33. ~~Consumer-claim drift posture decision (USER-gated).~~ done — resolved — pins kept + guarded (check-consumer-claims.sh, T05)
34. i18n residue post-httputil v1.6.0.
35. BuildFlow dprint exit-14 on CHANGELOG-only commits (upstream).
36. module-details.md decay watch: re-verify a module's section on its next train. **→ cqrs section re-verified at the v0.8.0 train; watch continues**
37. Confirm the 220-line cap's source of truth and document it where repos can see.
38. `docs/agents/` directory charter (agent-only, never user docs) before it grows.
39. buildflow skill references: document `.jscpd.json` convention + the go-work-sync/charter caveat (PR to the skill repo).
40. ~~Re-run `check-pin-drift.sh` post-trains (held cqrs-htmx leg flips with the 1.27.2 train).~~ done — 27 pin-drift legs green post-trains

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **lychee g1 decision**: the issues are filed, but #49 asks BuildFlow to pick authenticate-vs-exclude — since you own both sides, do you already have a preference (token injection vs namespace default-exclude) that should go into the issue as the decision, rather than leaving it open?
2. **exhaustruct upstream filing**: file at golangci-lint now (needs its own verification pass against their source), or park it until the next fleet lint-config change touches exhaustruct?
3. **pma hook bypass**: #46 fixes the BuildFlow mutator, but leaks #1–7 (gopls-class) still bypass every hook via pma's `core.hooksPath=/var/empty` — do you want a daemon-committed guard (CI-only today) or is the CI net + charter `--fix` cadence the accepted posture?

---

**Files touched this follow-up:** `TODO_LIST.md` (2 rows), `docs/drafts/` (5 new draft files), this report. **External:** BuildFlow #46–#49, go-auto-upgrade #3 (all OPEN, verified).
