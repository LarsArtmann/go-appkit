# Status Report: Deduplication Session + Toolchain Drift Discovery

- **Date:** 2026-09-24 16:52 CEST
- **Scope of this report:** the art-dupl deduplication session (`-t 3 --type-aware`) on go-appkit, including everything noticed along the way. No other research was done.
- **Trigger:** `art-dupl --sort total-tokens -t 3 --type-aware --rich-text --explain --html` → 3 actionable clone groups (8 clones, 24 tokens; 82 suppressed).

---

## Stat Cards

| Metric | Value |
| --- | --- |
| Actionable clone groups | 3 → **1** (only the deliberately accepted group remains) |
| Detected groups | 85 → 83 |
| Modules verified green (build+vet+test -race+lint) | **2 of 11** (core, health) |
| Repo buildability at session start | **0 of 11** (nothing compiled at any toolchain) |
| go-directives guard | **FAIL on all 11 modules** |
| AGENTS.md length | **377/377 cap** (zero headroom) |

---

## a) FULLY DONE

1. **Clone group triage with judgment** — all 3 actionable groups read in source, classified extract-vs-accept per the deduplicate-code skill.
2. **Group 1 extracted (health):** three identical mutex-guarded `started = false` flips (`health/mount.go`, failed probe Start / failed dashboard Start / Shutdown) → one `setStarted` helper. The "started is only mutated under the mutex" invariant now lives in exactly one place.
3. **Group 3 extracted (core):** `runDrainHooks`/`runShutdownHooks` (identical except hooks slice + error code) → shared `runHooks` in `service.go`, with two thin wrappers. Used the non-formatting `errorfamily.WrapInfrastructure` (vet's printf check rejects passing a variable through `WrapInfrastructuref`'s format parameter). Error codes passed through byte-identical.
4. **Group 2 accepted with rationale:** sorted-map-keys pattern in `cqrs/eventservice.go`, `flightrecorderhealth/adapter.go`, `metrics.go` — three independent modules with no shared dependency, different domain purposes, each already named and documented. Extraction would cost a new module for 4 idiomatic lines.
5. **Verification loop closed:** core + health green on `go build`, `go vet`, `go test -race -count=1` (incl. `shutdownlog_test` phase sequence and drain-window ordering tests), `golangci-lint` 0 issues in both; art-dupl re-run with identical flags: actionable 3 → 1.
6. **Root cause found for "nothing builds":** env pins `GOTOOLCHAIN=local` (go1.26.7) while committed go.mods demand up to go 1.27.1 (go-etag v0.6.0's `server` package) and the untracked `go.work` says `go 1.26`. Cached go1.27.0/1.27.1 toolchains found in `/mnt/buildcache`; `GOTOOLCHAIN=go1.27.1 GOWORK=off` validated as the working override.
7. **Dependency normalization (root + health):** `go mod tidy` — go directive 1.27 → 1.27.1 (forced by the committed dependency graph), dropped the unused bare `go-etag` require + 4 stale go.sum lines per module.
8. **Memory updated:** toolchain-drift gotcha added to AGENTS.md (exactly at the 377-line cap). No manual commits, no pushes (auto-daemon committed the work as `928633f`, `a1aea5b` — expected behavior).

## b) PARTIALLY DONE

1. **Toolchain drift remediation** — root + health normalized; the other **9 modules untouched** (directives still span 1.26.5 / 1.26.7 / 1.27), `go.work` still stale at 1.26, `check-go-directives.sh` still FAIL on all 11.
2. **Deduplication** — 2 of 3 actionable groups extracted; the accepted group remains visible by design; the **82 auto-suppressed groups were never inspected**.
3. **Session verification** — 2 of 11 modules proven green. The other 9 (incl. `integration`, which pins published tags) are untested under go1.27.1.
4. **AGENTS.md** — gotcha note added, but the per-module build-command blocks were NOT rewritten (copy-pasting them still fails until the directive unification lands). File is now at its 377-line cap with zero headroom.
5. **Module CHANGELOGs** — no `[Unreleased]` entries were added for the two refactors (keep-a-changelog scaffolds exist in core and health).

## c) NOT STARTED

- The go-directive unification decision and rollout (needs Lars's policy call, see Questions).
- CI state verification (the `go-directives` job reads `go.work`, which is **untracked** — on a fresh checkout the script has no file to parse, so the job is **suspected red** for a different reason than the mismatch; `ci.yml` has no go.work generation step, only comments).
- Test pinning the `server.drain_hook_failed` / `server.shutdown_hook_failed` error codes (see d-3).
- Audit of what the 82 suppressed clone groups contain.
- HARVEST of this report's section (f) into TODO_LIST.md (deliberately deferred — instructed to wait).
- Cross-project lesson write-up ("check GOTOOLCHAIN/go.work parity before trusting AGENTS build commands") to crush-config `references/lessons.md`.

## d) TOTALLY FUCKED UP

1. **The committed repo was unbuildable at ANY toolchain** (pre-existing, not from this session): go1.26.7+`GOTOOLCHAIN=local` rejects the go directive; go1.27.0 fails on go-etag v0.6.0 (`requires go >= 1.27.1`); the committed go.mod was internally inconsistent until tidied. Every AGENTS.md build command was copy-paste-broken.
2. **`scripts/check-go-directives.sh` — the guard built to catch exactly this class — is FAIL on all 11 modules**, and because `go.work` is untracked, the CI job reading it likely cannot even parse a file on the runner. The guard exists, is committed, and appears to be firing red (or vacuously) in CI. Unverified remotely, flagged.
3. **My claim "pinned contracts preserved" was overbroad.** TRUE: `shutdownlog_test` pins the phase logging, drain-window tests pin ordering. FALSE (discovered post-hoc): **no test anywhere pins the `server.drain_hook_failed` / `server.shutdown_hook_failed` codes** — grep matches only in `service.go`. The refactor preserved them by construction (constants passed through), but there is no regression net for exactly the values consumers match on.
4. **Self-inflicted, caught by gates (2 wasted cycles):** first `runHooks` version didn't compile (`undefined: ctx` — forgot the parameter in my own signature), and the second tripped vet's non-constant-format-string check. Gates worked; craftsmanship didn't. There is also **no pre-edit green baseline** — the baseline attempt failed on the toolchain and I proceeded to edit anyway, so a red test could not have been attributed to pre-existing vs regression.

## e) WHAT WE SHOULD IMPROVE

1. **Baseline-first discipline:** establish a green `go test` baseline BEFORE refactoring, even when the environment fights back. "Environment broken" is a reason to fix the environment first, not to skip the baseline.
2. **Verify the pin, don't assume it:** before claiming a contract is "pinned by tests", grep for the assertion. A godoc promise without a test is a wish.
3. **Inspect what the tool suppresses:** "actionable 3, suppressed 82" — the suppressed 82 were trusted blindly. Suppression heuristics deserve at least one audit pass before declaring zero harmful duplication.
4. **Leave headroom in capped files:** AGENTS.md was at 375/377; adding a 2-line note slammed it to the cap. Trim or extract before adding, not after the next person trips the structure linter.
5. **Changelog-with-refactor:** internal refactors on released modules get a `[Unreleased]` CHANGELOG line in the same train, or the next release ritual reconstructs history from git.
6. **Untracked-but-CI-read files are a drift class of their own:** scripts that read `go.work` in CI require `go.work` to be committed or generated. One `git ls-files` check would have caught it.
7. **First-compile quality:** writing a helper whose body references a parameter that isn't in the signature is a typing-speed error, not a knowledge gap — slow down on signatures.

## f) NEXT — up to 50 things to get done (impact-ordered; brainstorm, not commitment)

**P0 — unblock the repo**
1. Decide the go-directive floor policy (Question 1 below).
2. Unify all 11 go.mods to the chosen floor; regenerate `go work use`.
3. Make `check-go-directives.sh` green locally.
4. Re-run the full per-module matrix (build/vet/test -race) on all 11 modules.
5. Re-run `golangci-lint` on all 11 modules.
6. Verify gopls/LSP diagnostics recover after the go.work fix (restart LSP).
7. Commit or CI-generate `go.work` (decide tracked status; CI job reads it).
8. Verify the CI `go-directives` job's actual remote state (suspected red).
9. Run `./scripts/check-pin-drift.sh` (release-ritual step 5; cheap).
10. Run the `integration` module suite (pins published tags) under the unified state.
11. Update AGENTS.md build-command blocks to match the post-unification reality (drop GOTOOLCHAIN prefix and/or stale GOEXPERIMENT notes).
12. Trim AGENTS.md back under the cap with headroom (extract the toolchain note detail if needed).
13. Add the hook-error-code pinning test (`server.drain_hook_failed` / `server.shutdown_hook_failed`).
14. Add `[Unreleased]` CHANGELOG entries for the `setStarted` and `runHooks` refactors (core + health).
15. Audit auto-commit `ea17377` (9 files) — what drifted the module directives in the first place.

**P1 — close this session's gaps**
16. Inspect the 82 suppressed art-dupl groups; confirm suppression is sound.
17. Re-run art-dupl at `-t 5` and `-t 10` for a comparison baseline.
18. Decide standing policy for the accepted sorted-keys clone (exclude config vs keep visible).
19. Fresh-consumer proxy smoke for core (go.mod/go.sum changed; next tag ships them) per `doc/recipes/fresh-consumer-proxy-check.md`.
20. Audit auto-commits `928633f` / `a1aea5b` (this session's daemon commits) for sane messages/content.
21. Run `check-dependabot-parity.sh` locally once (cheap guard verification).
22. Check whether dependabot's grouped gomod updates are what bumps go directives; add ignore/allow rules if so.
23. Verify go-etag v0.6.0 is actually required transitively by which modules (the "unimported" in its commit message is misleading — `entitytag`/`server` are live in the graph).
24. Wire `check-go-directives.sh` into BuildFlow pre-commit so parity breaks before push, not in CI.
25. Record the cross-project lesson ("check GOTOOLCHAIN/go.work parity before trusting AGENTS build commands") in crush-config `references/lessons.md`.

**P2 — quality / hygiene**
26. HARVEST this report into TODO_LIST.md (P0 items as actionable tasks, the rest as roadmap fuel).
27. `sort.Strings` → `slices.Sort` modernization sweep (Go 1.21+ idiom; noticed at all three accepted sites).
28. Consider a named `type Hook func(context.Context) error` to shorten the `runHooks` signature (config.go already declares the raw func type twice).
29. Live E2E of `example/` and `health/example` after the refactors (tests cover compile paths; demos cover the wired path).
30. Design-decision entry for "accepted duplication" in `doc/planning/design-decisions.md` (one paragraph, so the accept survives personnel changes).
31. Decide whether art-dupl becomes a recurring BuildFlow/ritual step or stays ad-hoc.
32. Dry-run the `go work use` fix in a scratch work file before touching the real one.
33. Decide machine-level `GOTOOLCHAIN` policy (Question 2) — upgrade default toolchain to go1.27.1 or keep the pin.
34. Confirm health module's `go-health v0.3.0` evaluation note in AGENTS is still accurate post-drift (untouched this session, listed for the next docs pass).

**P3 — roadmap fuel (from this session only)**
35. Testkit (`testkit.Serve`) could grow a drain-window assertion helper — the drain contract tests hand-roll it today.
36. `Mounted.setStarted` godoc could cross-link the drain-latch note in `Shutdown` (docs polish).
37. Errorpages/realtime/otel modules: untouched; schedule their go1.27.1 verification with the P0 matrix run (subsumed by item 4 — keep as checklist reminder).
38. Consider vendoring-toolchain pinning (`go.mod` `toolchain` directive) so the floor survives `GOTOOLCHAIN=local` machines.
39. `metrics.go` sortedRoutes: keys embed `|` separators with `SplitN` — a small typed seriesKey would kill the stringly-typed join/split pair (data-model polish noticed in passing).
40. Add CI job ordering: directives-guard first so red parity fails fast before the expensive matrix.
41. AGENTS.md Release Ritual: add "run check-go-directives.sh" next to check-pin-drift (step 5 extension).
42. Consider `errors.Join` output formatting test (multi-hook failure message shape is unobserved).
43. Document the `/mnt/buildcache` toolchain cache trick in the recipes dir (session-specific discovery, generalizes).
44. Add a session-startup doctor step: `go env GOTOOLCHAIN` + `check-go-directives.sh` before trusting any build command.
45. Write the `server.*_hook_failed` codes into `doc/DOMAIN_LANGUAGE.md` if it exists — they are consumer-matchable contracts.
46. Evaluate whether `runHooks`'s "let every hook run" semantics deserve an explicit test (failure in hook 1 doesn't stop hook 2).
47. Add the accepted-duplication rationale as a one-line `// intentional:` marker decision — pending policy from item 18.
48. Sweep for other `var errs []error` + `errors.Join` patterns that could reuse the new `runHooks` shape (none found this session; confirm).
49. Go error-family: consider asking upstream for doc clarification on `WrapInfrastructuref` printf semantics (vet interaction surprised us).
50. Post-unification: re-run the otel benchmark baseline once (module untouched, but toolchain jump 1.26.7 → 1.27.1 invalidates the 2026-09-16 numbers on principle).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Go-directive floor policy:** unify all 11 modules + `go.work` at **go 1.27.1** (accept what go-etag v0.6.0 forces, and expect dependabot to keep pushing), or downgrade/pin go-etag to keep a 1.26.7 floor? This decides items 1–5 and whether today's tidy direction was correct.
2. **`GOTOOLCHAIN=local`:** is the machine pin deliberate (reproducibility) — in which case repo commands should carry explicit overrides — or may it flip to `auto` so go.mod-driven toolchain selection just works? It changes what item 11 and 33 should look like.
3. **`doc/` vs `docs/`:** repo history keeps status/planning/feedback under `doc/` (singular; `docs/` is the catalog Go module), while you and the status-report skill both say `docs/status/`. Which path is canonical going forward? (I followed your explicit instruction this time — this file lives in `docs/status/` — but the split is now real.)

---

*Generated by Crush session 2026-09-24 (deduplication run). Point-in-time snapshot — annotate, don't rewrite, when bringing current.*
