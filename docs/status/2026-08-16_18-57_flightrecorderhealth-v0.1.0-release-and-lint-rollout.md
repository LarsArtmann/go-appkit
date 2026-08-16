# Status Report — flightrecorderhealth v0.1.0 release + repo-wide lint standard rollout

**Date:** 2026-08-16 18:57 (Sunday)
**Session:** Resume of the polish self-review ([2026-08-16_17-43](2026-08-16_17-43_flightrecorderhealth-polish-self-review.md) → fixes reported in [2026-08-16_18-06](2026-08-16_18-06_flightrecorderhealth-defect-fixes.md)). Second half: user approved (1) cutting the v0.1.0 tag and (2) adopting per-module `.golangci.yml` as the repo-wide standard.
**Scope:** This session's run: defect fixes → tag → lint rollout to 5 satellites with real code fixes in 2 of them.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                                                         | Evidence                                                                    |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| 1  | **All 8 self-review d)-defects fixed** (README compile bug, doc.go alias + value-type-registration bug, FEATURES.md fabricated counts, CHANGELOG `[Unreleased]` scaffold, varnamelen cruft, lint-config divergence)                                                                                                                          | prior report + `git log 2d55afc..25335c2`                                   |
| 2  | **README quick start compiles verbatim** — extracted by awk from the README, built in a scratch module against the _published_ go-flightrecorder v0.2.0 + go-health v0.0.2                                                                                                                                                                   | `/tmp/frhealth-readme-check` build clean                                    |
| 3  | **Contract machine-checked** — `go-health v0.0.2` dep (verified identical to local HEAD) + `var _ health.HealthRecorder = (*Trigger)(nil)` and `var _ do.HealthcheckerWithContext = (*Checkable)(nil)`                                                                                                                                       | `contract_test.go`                                                          |
| 4  | **3 runnable godoc examples** with verified output; **benchmark** (~4.7µs/batch); **real `health.New` Probe e2e test** (Probe → Trigger → snapshot on disk)                                                                                                                                                                                  | `example_test.go`, `benchmark_test.go`, `TestIntegration_RealProbeEndToEnd` |
| 5  | **Coverage 100.0% of statements** (bar ~80%) — closed the last nil-guard branch                                                                                                                                                                                                                                                              | `go test -cover`                                                            |
| 6  | **`GOWORK=off` hermetic battery green** (test -race, vet, build, lint) + `-race -count=10` stability + all 7 workspace modules `-race` green                                                                                                                                                                                                 | session log                                                                 |
| 7  | **Tag `flightrecorderhealth/v0.1.0` cut** (annotated, at `d3e3e51`, user-approved) — verified via `git tag -l` BEFORE marking done                                                                                                                                                                                                           | `git tag --points-at HEAD`                                                  |
| 8  | **Module-local `.golangci.yml` created for all 5 satellites** (cqrs, realtime, flightrecorder, errorpages, docs-mod) — reconciled enable-list, documented divergences in each header, standardized test-exclusion union (mnd, exhaustruct, err113, paralleltest, gochecknoglobals, goconst, varnamelen, wsl_v5, funlen, cyclop, testpackage) | 5 new files, `golangci-lint config verify` all valid                        |
| 9  | **flightrecorder module made lint-clean**: sentinel error + `errors.Is`-matchable wrap (`errHTTPStatus`), `rr` → `responseRecorder`, 16× `httptest.NewRequestWithContext`, context-aware GET/POST helpers in tests, noinlineerr fix. Tests pass `-race`, lint 0                                                                              | `flightrecorder/middleware.go`, `flightrecorder_test.go`                    |
| 10 | **errorpages module made lint-clean**: justified `//nolint:exhaustruct` on deliberate zero-value structs, `e` → `sourceErr`, 12× context-aware httptest requests, noinlineerr fixes in test + example. Tests pass `-race`, lint 0                                                                                                            | `errorpages/*.go`, `example/main.go`                                        |
| 11 | **docs-mod lint-clean with zero code changes** (only testpackage finding, excluded by standard)                                                                                                                                                                                                                                              | `golangci-lint run` → 0                                                     |
| 12 | Root docs updated for the tag: TODO_LIST release-state line, AGENTS.md module line + full file table                                                                                                                                                                                                                                         | both files                                                                  |
| 13 | Session status report written with real `date` timestamp                                                                                                                                                                                                                                                                                     | `2026-08-16_18-06_flightrecorderhealth-defect-fixes.md`                     |

## b) PARTIALLY DONE

| # | Item                                      | What's done                                                                                                                 | What's missing / why it stalled                                                                              |
| - | ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| 1 | **Repo-wide lint standard**               | 3 of 5 satellites fully clean (flightrecorder, errorpages, docs-mod); configs written for all 5; exclusion standard settled | **cqrs (~15 findings) and realtime (~22 findings) not yet fixed** — user requested status report mid-rollout |
| 2 | **Root-config alignment**                 | Root still lints only the core module in practice                                                                           | Root `.golangci.yml` still carries depguard with the satellite-hostile allowlist; decision pending (see g)3) |
| 3 | **AGENTS.md lint-standard documentation** | Module configs carry their own rationale headers                                                                            | The repo-wide "lint each module from its own dir" workflow is not yet written into AGENTS.md                 |

## c) NOT STARTED

| # | Item                                                                                                                                                                                                                                | Why it matters                                                   |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| 1 | **cqrs lint fixes** — noinlineerr ×10 (staleness_test), errcheck ×1 (fmt.Fprint), gochecknoinits ×1 (init in flightrecorder_test), nilnil ×1, ireturn ×3, wrapcheck ×5, exhaustruct ×1 (EventService), godoclint ×1 (package godoc) | Touches tagged-release code (v0.3.0) and test idioms; needs care |
| 2 | **realtime lint fixes** — noctx ×12, makezero ×2, wrapcheck ×3, exhaustruct ×1 (hubConfig), varnamelen ×1 (`ch`), nonamedreturns ×2, gocognit ×1 (`Handler` 34 > 30)                                                                | Same class as done modules; gocognit may need a refactor         |
| 3 | Fresh-consumer proxy test (`go get flightrecorderhealth@v0.1.0` from clean module)                                                                                                                                                  | Blocked on push (tag is local-only)                              |
| 4 | pkg.go.dev rendering check                                                                                                                                                                                                          | Blocked on push                                                  |
| 5 | Root AGENTS.md / TODO_LIST closure of the depguard P-item                                                                                                                                                                           | Blocked on g)3                                                   |

## d) TOTALLY FUCKED UP

| # | Item                                                                                                                                                                                                                                                                                                                                                                                          | Impact                                                                             |
| - | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| 1 | **Ran `golangci-lint fmt` on a file with line-attached nolint directives.** The auto-formatter (golines) wrapped `errorpages.Write(...)` calls to multiple lines, stranding my `//nolint:exhaustruct` comment on the closing-paren line: 2 exhaustruct findings resurfaced + 2 nolintlint "unused directive" findings. Hand-repaired by rewriting the block single-line with short directives | 3 wasted lint cycles; churn in `example/main.go`                                   |
| 2 | **Introduced a compile error with a blind variable reuse.** noinlineerr fix in errorpages_test.go changed `if err := json.UnmarshalRead(...)` to `err = ...` without checking `err`'s type — it was `*errorfamily.Error`, not `error`. Compiler caught it (typecheck)                                                                                                                         | 1 wasted cycle                                                                     |
| 3 | **Churned `example/main.go` three times** (long nolint → shortened → hand-unwrap formatter damage) because I attached a >120-char comment to a line under a 120-col formatter                                                                                                                                                                                                                 | Predictable: comment length forces the wrap; should have been short from the start |
| 4 | **Briefly violated my own verification rule**: after the last hand-fix multiedit I moved on without re-running lint/test; the user's status interrupt arrived while errorpages was UNVERIFIED. (Ran the battery immediately after: tests pass, 0 issues — so no harm, but the gap was real)                                                                                                   | The exact failure mode e)1 of the previous report was about to repeat              |
| 5 | **Used `sed -i` on test files again** (16 + 12 httptest.NewRequest replacements) despite last session's sed self-mangle lesson. Scoped correctly and it worked, but the risk pattern stands                                                                                                                                                                                                   | None this time; process risk remains                                               |
| 6 | **Sloppy python heredoc** for config updates (SyntaxWarnings for `\.` escapes, assert-guarded so safe) — noisy tooling where a templated file write would have been cleaner                                                                                                                                                                                                                   | Minor                                                                              |
| 7 | **Inconsistent go-version pins**: my satellite configs say `go: 1.26.5`, root config says `go: 1.26.4`                                                                                                                                                                                                                                                                                        | Cosmetic now, confusion later                                                      |

## e) WHAT WE SHOULD IMPROVE

| # | Item                                                                                                                                               | Rationale      |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| 1 | **Never run `golangci-lint fmt` (or any auto-formatter) on files carrying line-attached nolint directives** — hand-edit instead                    | Prevents d)1   |
| 2 | **Type-check the target variable before converting inline `if err :=` to `err =`** — inline errors often shadow a differently-typed outer variable | Prevents d)2   |
| 3 | **Keep trailing comments short on lines a 120-col formatter owns**; long justifications go on the line above                                       | Prevents d)3   |
| 4 | **One edit → one verify, no exceptions**, even mid-rollout when "the next module" is calling                                                       | Prevents d)4   |
| 5 | **Pin `run.go` consistently** across all configs (root should move to 1.26.5)                                                                      | Config hygiene |
| 6 | **Disk-capacity check belongs at session start** (`df -h` on cache mounts). The /mnt/buildcache failure cost a diagnosis cycle mid-build           | Early warning  |

## f) UP TO 50 THINGS TO GET DONE NEXT

1. Verify errorpages stays green under the daemon's eventual reformat (it is green now: tests + 0 issues)
2. Fix cqrs noinlineerr ×10 in `staleness_test.go` (+1 eventservice_test)
3. Fix cqrs errcheck (unchecked `fmt.Fprint` in metrics_test)
4. Refactor cqrs `flightrecorder_test.go` `init()` away (gochecknoinits; same class as root's httpspec_test item in TODO_LIST)
5. cqrs `nilnil` in eventservice.go — inspect whether nil,nil is a design smell or needs a sentinel
6. cqrs `ireturn` ×3 (DeadLetterStore interface returns) — add to ireturn allow-list or wrap in concrete type
7. cqrs `wrapcheck` ×5 (projectionhost.Host error returns) — wrap at call sites with `%w` or extend ignore-sigs with justification
8. cqrs `exhaustruct` on `EventService` (mu, closed zero by design) — justified nolint or exhaustruct exclude
9. cqrs `godoclint` package-godoc placement in eventservice.go — move/fix doc comment
10. Re-run cqrs lint to recount after the test-exclusion standard landed
11. Fix realtime noctx ×12 (http.Get/NewRequest in realtime_test) — context-aware helpers like flightrecorder's
12. Fix realtime makezero ×2 (`tmp` slice)
13. realtime `wrapcheck` ×3 (go-sse returns) — same decision as cqrs wrapcheck
14. realtime `exhaustruct` hubConfig — justified nolint
15. realtime varnamelen `ch` → rename (e.g. `clientChan`/`events`)
16. realtime nonamedreturns ×2 — check for defer/recover dependency first, then strip names
17. realtime gocognit `Handler` 34>30 — extract helpers or raise threshold with justification (prefer extraction)
18. Root `.golangci.yml`: decide depguard fate (g)3), bump `go: 1.26.4` → `1.26.5`
19. Document the repo-wide lint workflow in AGENTS.md ("lint each module from its own dir; root config covers only core")
20. Close/annotate TODO_LIST depguard P-item once the standard is documented
21. Standardize flightrecorderhealth's test-exclusion list to the union (it lacks funlen/cyclop/testpackage — cosmetic, currently no findings either way)
22. Commit/push decision for the whole wave (5 tags pending push: core v0.3.0, cqrs v0.3.0, realtime v0.1.0, flightrecorder v0.1.0, flightrecorderhealth v0.1.0)
23. Fresh-consumer proxy test after push: `go get github.com/larsartmann/go-appkit/flightrecorderhealth@v0.1.0` in a clean /tmp module
24. pkg.go.dev rendering check for flightrecorderhealth after push
25. `go.sum` reproducibility spot-check on a second checkout
26. Upstream: `go-flightrecorder` `BufferFull()` accessor idea (Checkable could warn pre-overrun)
27. flightrecorderhealth `WithOnCapture(func(fr.SnapshotEvent))` hook
28. flightrecorderhealth `Trigger.Recorder()` accessor
29. flightrecorderhealth probe-style combined wrapper (one handle → Checkable() + Trigger())
30. Repair/free `/mnt/buildcache` or permanently repoint cache env vars (d-blocker for unoverridden builds)
31. Consider `errorfamily` assertion tests in other adapters (pattern from contract_test.go)
32. Add `date` + `df -h` to session-start checklist (memory: AGENTS.md)

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **`/mnt/buildcache` is failing** (GOCACHE + GOMODCACHE + golangci cache live there; disk `/dev/sda1` at 99%, directory reads return I/O errors). Every command this session needed manual cache overrides. Should I (a) wait for you to free/repair the disk, or (b) permanently repoint the cache env vars to `~/.cache/...` in your shell config (needs your approval to edit dotfiles)?
2. **Push the tag wave?** Five tags are local-only pending your gate (core v0.3.0, cqrs v0.3.0, realtime v0.1.0, flightrecorder v0.1.0, flightrecorderhealth v0.1.0). Pushing unblocks the fresh-consumer test and pkg.go.dev checks.
3. **Root `.golangci.yml` depguard**: now that every satellite has its own config, should the root config drop depguard entirely (per-module configs are THE lint path), or keep the root allowlist for the core module only? This decides f)18-20.

---

**Bottom line:** flightrecorderhealth v0.1.0 is tagged, verified (100% coverage, compiler-enforced contract, hermetic battery green), and the repo-wide per-module lint standard is real: 4 of 6 modules lint-clean, ~37 code-level findings fixed or excluded-with-justification across flightrecorder + errorpages + docs-mod. cqrs (~15) and realtime (~22) findings remain, all catalogued in f). The one machine-level blocker outside the repo is the 99%-full buildcache disk.
