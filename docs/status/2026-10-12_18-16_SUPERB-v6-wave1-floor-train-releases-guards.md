# Status Report — SUPERB v6 Full Execution, Wave 1: Floor Train, Release Trains #25/#26, Guard Quartet

**When:** 2026-10-12 18:16 CEST
**Scope:** this session (≈16:50–18:16) — the "NOW GET SHIT DONE! The WHOLE TODO LIST!" execution of `docs/planning/2026-10-12_15-17_SUPERB-v6-trains-guards-and-gated-unlocks.md`, wave 1 (tiers 1% + 4%). Gates G1/G2 were answered by the execution order itself; G3 defaulted to the non-destructive option (see §g).
**Concurrent sessions:** two landed mid-session (15:53 BuildFlow triage + AGENTS slim-down; 17:27 fleet issue filings) — coordinated via tree state, zero collisions, one of their regressions caught and fixed by a guard (see §a.7).

---

## Headline

| Thing | Before | After |
| --- | --- | --- |
| go floor | 1.27.1 (12 go.mods + go.work) | **1.27.2** everywhere, matrix green |
| pin-drift guard | 1 leg red by design (cqrs-htmx hold) + guard's own toolchain pin stale | **27/27 legs GREEN**, guard now floor-derived |
| Releases | core v0.8.0 / health v0.1.5, #25 #26 open | **core v0.9.0 + health v0.2.0 pushed, proxy-PASS ×2, both issues closed** (0 open) |
| Guards | charter fail-only; no hermetic self-check; structure-lint unguarded in CI; pre-tag single-tag | **charter `--fix`; integration hermetic TestMain; structure-linter CI job; multi-tag pre-tag gate** |
| Commits | — | 13 on origin (all pushed with the train) |

## a) FULLY DONE

1. **T01 — 1.27.2 floor train, complete.** All 12 go.mods + go.work `1.27.1 → 1.27.2` (`go mod edit` + tidy each; tidy was a directive-only no-op — 13 files, one line each, zero dep churn). `documentedGoDirective` fixture + AGENTS toolchain paragraph + both GOTOOLCHAIN prefixes swept. Verification matrix: build+vet 12/12, `test -race` 12/12, sequential golangci-lint 12/12, directives guard 12/12, parity + charter + structure guards green.
2. **Lint-config hygiene rode the train (15:53 report's f)2, extended):** ALL 12 `.golangci.yml` carried a stale `run.go: 1.26.7` (the report only named root) — now `"1.27.2"` everywhere; the five `goexperiment.*` build-tags entries dropped (zero `.go` file carries those tags — analysis-only dead config since the 2026-09-29 retirement). `golangci-lint config verify` green ×12.
3. **T07 — charter `--fix`.** `check-workspace-charter.sh --fix` reconciles go.work via `go work edit -dropuse=`/`-use=` (NOT `go work use -drop` — that flag does not exist; see §d.1). Worktree-tested: planted leak → guard red → `--fix` drops → green → second `--fix` no-op with byte-identical go.work. Pre-commit stays fail-only (warn-and-offer documented in AGENTS).
4. **T08 — integration hermetic self-guard.** New `integration/main_test.go` TestMain probes the resolved workspace (`go env GOWORK`) and hard-fails any workspace-governed build with fix instructions. The covered danger is real and was **proven in simulation**: with a charter leak present, the pre-guard suite passed SILENTLY against local code. Standalone module mode (no go.work found) is hermetic-equivalent and passes; bare in-dir runs already fail at setup. Pre-commit warn leg added for staged `integration/` files. Noctx/gocritic/nlreturn clean; suite green.
5. **T09 — structure-linter config-loss CI guard.** Discovered the binary **exits 0 even when reporting findings** (config renamed → 11 findings → exit 0), so a bare CI step would be worthless. New `scripts/check-structure-linter.sh` (missing-config + `total>0` assertions, both paths tested) + CI `structure-linter` job (installs the public tagged binary).
6. **T10 — multi-tag pre-tag-checks.** `pre-tag-checks.sh tag [tag…]`: guards run once, per-tag artifact assertions loop. Dry-run on two existing tags proved per-tag reporting. Release Ritual step 5 documents the wave form.
7. **T02 — integration re-pin + TWO guard-bug fixes.** cqrs-htmx `v4.13.2 → v4.14.0` (go.mod/go.sum/fixture), suite + lint green. (i) `check-pin-drift.sh` had hard-coded `GOTOOLCHAIN=go1.27.1` — stale the moment the floor train landed; its `go list` died silently on v4.14.0's ≥1.27.2 requirement and every cross-repo leg said "could not query". Now derives the floor from the root go.mod (train-proof). (ii) The 15:53 AGENTS rewrite dropped the closing paren on the "Latest per module" line — the guard's parser flagged it (format drift + missing claim); restored. **All 27 legs green — first fully green run since the hold opened.**
8. **T03 — release trains shipped.** API diffs: core additions-only (`testkit.NewService`) → **v0.9.0**; health doc-only diff, behavior deltas → **minor health/v0.2.0**. CHANGELOGs dated (+ floor notes), same-train docs (AGENTS Release State + core bullet, FEATURES de-Unreleased ×3 + the security example-dep claim made version-exact, TODO header rewritten + 2 rows closed), annotated tags, **multi-tag gate GREEN, pushed** (master + both tags), **fresh-consumer proxy PASS ×2** (recipe script), **#26 and #25 closed** with evidence-carrying comments — zero open issues on the repo.
9. **Release drift caught pre-push by the fixture (the system working):** health v0.2.0 actually rides **go-health v0.5.1** (the NewChecks tag) + dashboard v0.10.2, while its CHANGELOG entry said v0.5.0 and the fixture said v0.5.0 — the 14-16 session bumped deps without updating docs. Fixed the CHANGELOG + module-details deps row + fixture, **deleted the local-only tag and re-tagged on the corrected commit** — zero proxy damage (this is exactly why the ritual orders gate-before-push).
10. **3.9 — `duration_ns` through the full appkit stack.** New `integration/health_durations_test.go`: published core v0.9.0 + health v0.2.0 through `testkit.NewService` (its first in-repo consumer) — slow check self-times ≥25ms, fast >0, ctx-ignoring wedged check fails closed at the 250ms batch deadline with the naming error, roll-up degrades to warn, readiness stays 200. Suite + hermetic lint green.
11. **T21 — evaluated and closed as no-op:** the "hand-rolled Addr/NoDrainDelay" sites are deliberate (core's own suite tests core; testkit's raw-path tests pin the non-NewService path consumers still use). Migrating them would DELETE coverage.
12. **T26 — largely overtaken by the 15:53 session** (AGENTS 126/220 lines, detail in module-details.md). No further extraction done — headroom is large, and cutting more would be verschlimmbesserung. Residual: cap source-of-truth note (see §f).
13. Housekeeping: straggler 15:53 report committed with a detailed message; concurrent-session work deduped (their BuildFlow filings = T22's BuildFlow half — not re-done); every commit this session except two daemon-swept batches carries a full message (see §e.1).

## b) PARTIALLY DONE

1. **T04 — owner actions:** 4.1 push DONE (rode 3.6); 4.2 (`SSH_PRIVATE_KEY` secret restore) and 4.3 (watch one fully green UNSKIPPED matrix run) are owner-only — the pushed train should already prove the guard jobs + proxy smokes green in CI, but nobody has watched the run.
2. **T22 — upstream filings:** BuildFlow half done by the 17:27 concurrent session (#46–#49 + go-auto-upgrade#3, verified OPEN). The FAMILY drafts (go-sse, httputil ×3, go-health pack, samber/do, fr cooldown, GracefulClose) and the exhaustruct_v5-at-golangci-lint candidate remain (were sequenced later in the plan anyway).
3. **T19.3 — render-check of the new tags** on pkg.go.dev (crawler-lag tolerant) not yet done; the fr 404 investigation (T19 main) not started.

## c) NOT STARTED (plan tasks, this session touched none of them)

T05 (consumer-claims script — G3=A chosen, not built), T06 (go-cqrs-lite v4.13.0 cookbook + pin sweep + AGENTS deps-table truth — note: the table's "system v4.10.2 … ARE the latest published" claim is now doubly stale), T11 (health follow-through batch), T12 (otel polish batch), T13 (example-smoke harness), T14 (StartHooks example migration), T15 (fr v0.2.1 sweep on frh/cqrs trains), T16 (cqrs 14 + frh 13 FEATURES evidence cells + health test-name grep), T17 (hygiene A), T18 (hygiene B + annotate the 15 reports of 2026-10-0x + harvest 15-53/17-27 §f + index both unindexed reports), T19 main (pkg.go.dev fr 404), T20 (pma journalctl hunt + owner notes), T23 (errorpages swap), T24 (browser CSP, env), T25 (govulncheck, env), T27 (watch rotation).

## d) TOTALLY FUCKED UP

1. **`go work use -drop` does not exist.** First `--fix` implementation written from memory; the worktree test caught it (`flag provided but not defined`). Fixed to `go work edit -dropuse=`. Lesson: `go help work edit` before writing tool invocations from memory. No damage (worktree discipline).
2. **Health API diff ran against the wrong extraction** (whole-repo archive root documents CORE, not health) — produced a garbage 300-line "diff". Re-ran from the archive's `health/` subdir. The ritual's path discipline exists; I skipped it once.
3. **`ts.URL` vs `ts.FullChainURL`:** wrote the E2E against working-tree memory instead of the PUBLISHED tag's API — the pin-charter lesson in module-details bit me exactly as documented. Fixed via `go doc` in the pinning module.
4. **`${PIPESTATUS[0]}` is unsupported by this shell harness** — 12 lint runs reported empty exit codes; full sequential re-run with file-redirected verdicts needed.
5. **First floor-train lint pass ran integration BARE** (no `GOWORK=off`) and hit the exact T08 gap class (workspace resolution error). Ironic, instructive, fixed hermetically — and it validated the guard's reason to exist.
6. **The E2E test needed two lint-iteration rounds** (8 findings, then one golines) and the daemon committed a linty intermediate because I committed before linting. Lint-before-commit is cheap; do it.
7. **Checkout-SHA typo in the CI job edit** (`…76410b1` vs `…76410c181273ba90b1`) — edit tool no-match caught it; copied exact.

Nothing unrecoverable; zero reverts; every failure was caught by a test, a guard, or the edit tool's exact-match discipline before it could ship.

## e) WHAT WE SHOULD IMPROVE

1. **Commit immediately after each verified micro-batch.** The daemon ate three of my detailed messages this session ("auto-commit N files" instead). The content always landed, but attribution and reasoning were lost — future archaeology will find bare heuristic commits where there should be explanations.
2. **`go help <cmd>` / read the pinned tag's docs BEFORE writing code against memory** (§d.1, §d.3 — both were documented lessons I re-learned the cheap way).
3. **Lint the file before every commit** — one file cost two extra rounds (§d.6).
4. **Exit codes through this harness:** never rely on bash-isms like PIPESTATUS; redirect to a file and echo `$?` directly.
5. **Concurrent-session protocol worked** (View-before-edit, tree-state dedupe, guard-caught their regression) — keep it; consider a standing "check `git log` since your last commit before editing shared files" cadence.
6. The guards' own config must ride the trains they guard (pin-drift's stale toolchain prefix would have failed the NEXT release's gate — caught only because I ran the guard mid-train). New rule-of-thumb: run every guard BEFORE and AFTER each train step, not just at the gate.

## f) NEXT — up to 50, ranked by impact

1. **T06** go-cqrs-lite v4.13.0 response: README cookbook re-verification (scenario/v4 + testutil vs shipped APIs), pin-sweep decision, AGENTS cqrs deps-table truth (stale twice over).
2. **T05** build `scripts/check-consumer-claims.sh` (G3=A): parse AGENTS/FEATURES consumer pins vs `git -C` cqrs-htmx tags; wire into Release Ritual; negative test.
3. **T11** health follow-through: BenchmarkNewProbe pre/post, one real `FuzzNewProbe_PanicIsolation` session, testkit `stopTimeout` godoc note, frh contract re-check vs NewChecks, otel fr-bridge example E2E re-run.
4. **T12** otel polish: `BenchmarkRecovery`/`BenchmarkRecordError` + README perf table, `OTEL_METRIC_EXPORT_INTERVAL` pin test, gzip/retry env notes, tail-sampling recipe, span-clobber integration assertion, metrics-demo note.
5. **T14** StartHooks example migration (fr + otel examples) — trigger met.
6. **T13** example-smoke harness (`scripts/example-smoke.sh`, six examples, build+run+probe+SIGTERM+graceful-line assert).
7. **T15** fr v0.2.1 sweep on the frh + cqrs trains (frh's go-health pin is ALSO stale at v0.2.0 per module-details — fold into the same train decision).
8. **T16** FEATURES evidence cells: cqrs ×14, frh ×13; grep health #25 test names against `probe_test.go`.
9. **T17** hygiene A: `docs/reviews/README.md`, docserver `Config.ServiceName` mirror into `docs/README.md`, drainhooks pointer decision, straggler routing (14-16 §f5/§f6→closed-by-T21/§f13, 14-17 §f29/§f45).
10. **T18** hygiene B: dprint pass over today's edited .md files; annotation pass over the 15 reports of 2026-10-0x; harvest 15-53 + 17-27 §f into TODO/ROADMAP; index 15-53 + 17-27 reports in `docs/status/README.md`.
11. **T19** pkg.go.dev: fr 404 investigation/indexing request; render-check v0.9.0 + health/v0.2.0.
12. **T22** family filings: go-sse ReplayFiltered, httputil Logging-ctx, F2 timing battery, NewServerListener, go-health pack + panic-comment stance, samber/do lazy-healthy; go-licenses premise re-verify FIRST; exhaustruct_v5 schema at golangci-lint (own verify pass, no local clone — needs source fetch).
13. **T23** errorpages `statusRecorder` → `httputil.ResponseRecorder` + release.
14. **T20** pma remainder: `journalctl -t pma` hunt attempt around leak windows; owner notes for §6.
15. **T25** govulncheck health+security — try `nix shell nixpkgs#govulncheck` before declaring env-blocked (the 15:53 session used nix for shellcheck).
16. **T24** browser CSP pass (chromedp availability probe first).
17. **T27** watch rotation: cordis count, PapDashboard, nixpkgs go, cqrs-htmx v5 distance, httputil v1.6.0 (i18n residue + `check-pin-drift`), battery demand re-check, core v1 exit-criteria vs the v0.9.0 surface.
18. **Watch the pushed train's CI run** (guards + proxy smokes should be green; matrix stays SSH-gated — owner).
19. **T04 owner:** restore `SSH_PRIVATE_KEY`, one fully green UNSKIPPED run.
20. samber/lo adoption train decision + execution (policy question pending, §g.2; ~12 sites in 8 modules).
21. BuildFlow #46–#49 + go-auto-upgrade#3 watch: drop `skip_steps`/`.jscpd.json` workarounds as fixes land.
22. `buildflow --verbose` once: name the 9 unavailable tools; audit the 67 not-applicable list.
23. Enumerate the truncated `+2` cqrs-lint findings from a file (never-pipe rule).
24. Consider `golangci-lint config verify` as a CI job (cheap; would have caught the exhaustruct class at authoring).
25. `.jscpd.json` ignore strategy decision (path- vs format-scoped) independent of BuildFlow#48.
26. `docs/agents/` directory charter (agent-only, never user docs) before it grows.
27. Link `docs/agents/module-details.md` from module READMEs (discoverability) or accept pointer-only.
28. Confirm the 220-line AGENTS cap's source of truth (BuildFlow preflight vs linter) and document it fleet-visibly.
29. module-details decay watch: re-verify a module's section on its next train (health's section just got its train — cqrs/frh next).
30. Daemon-bypass guard (post-buildflow charter/directives run or watcher) — the pma 62-second window.
31. Fleet sweep for go-work-sync re-adds in sibling repos (leak #8's mutator was fleet tooling).
32. Roadmap fuel from the 17-27 report §f45 (workspace-exclusion convention, per-tool config schema docs, charter-guard-as-library).
33. Update the buildflow skill's verification table with the 17-27 session's findings (skill-repo PR).
34. cqrs README cookbook standing ritual on each go-cqrs-lite release (T06 institutionalizes the current one).
35. Health-example dep bump (example rides health's next train; v0.2.0 core dep is example-only).
36. Integration docs mention of the new hermetic TestMain in `integration/README.md` (currently only AGENTS/module-details carry it).
37. Status-index entry for THIS report (docs-health owns; fold into T18).
38. AGENTS line-count re-check after T18's harvest edits (126/220 today; headroom is the point).
39. Consider a follow-up `duration_ns` dashboard-side assertion (trend view renders durations) — demand-gated.
40. cqrs-htmx setup suite hermetic run (their repo; standing).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **G3 confirmation:** Full Execution Mode gave no explicit G3 answer, so I took the non-destructive default — KEEP the exact consumer pins in AGENTS/FEATURES and build `check-consumer-claims.sh` to guard them (option A). Strip (option B) stays available until T05 starts. Confirm A, or rule B?
2. **Push scope for the remaining waves:** I pushed master + the two release tags as plan micro-task 3.6 under "the WHOLE TODO LIST" (verification of a release is impossible unpushed — the proxy check IS the contract). The E2E test commit and every later train (T06/T15/T23…) also want pushes to be provable. Is push-per-completed-train authorized for the rest of this execution, or should each push wait for your word?
3. **samber/lo policy (blocks one 15:53 TODO row):** is `samber/lo` acceptable as a RUNTIME require that appkit library consumers inherit (fleet policy names lo+mo required), or test/dev-only with runtime sites kept hand-rolled? This decides the adoption train's shape (the 17:27 session's go-auto-upgrade#3 may silence the noise meanwhile, but the ~12 sites stay).

---

**Files touched this session:** 12 go.mods, go.work, `integration/{go.mod,go.sum,pin_drift_test.go,main_test.go(new),health_durations_test.go(new)}`, `CHANGELOG.md`, `health/CHANGELOG.md`, `FEATURES.md`, `AGENTS.md`, `TODO_LIST.md`, `docs/agents/module-details.md`, 12× `.golangci.yml`, `.githooks/pre-commit`, `.github/workflows/ci.yml`, `scripts/{check-workspace-charter.sh,check-pin-drift.sh,check-structure-linter.sh(new),pre-tag-checks.sh}`, `docs/status/2026-10-12_15-53_….md` (committed), this report. **External:** origin master + `v0.9.0` + `health/v0.2.0` pushed; issues #25/#26 closed.

**Verification state:** directives 12/12, workspace-charter, dependabot-parity, structure-linter (0 findings), pin-drift 27/27, pre-tag multi-tag gate GREEN, race suites 12/12 at the 1.27.2 floor, lint 12/12 (integration hermetic), fresh-consumer proxy PASS ×2, both tags on origin.

**NOW WAITING FOR INSTRUCTIONS.**
