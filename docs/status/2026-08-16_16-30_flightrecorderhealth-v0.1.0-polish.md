# Status Report — flightrecorderhealth v0.1.0 polish + tagging readiness

**Date:** 2026-08-16 (resumption)
**Session:** `go-appkit/flightrecorderhealth` quality + lint + docs cleanup
**Predecessor:** [2026-08-16_15-32_flightrecorderhealth-adapter.md](2026-08-16_15-32_flightrecorderhealth-adapter.md)

---

## a) FULLY DONE

| #  | Item                                                                                                              | Evidence                                                                                     |
| -- | ----------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| 1  | `Trigger.lastCapture` is goroutine-safe via `sync.Mutex`                                                          | `flightrecorderhealth/adapter.go:91` (field) + `178-194` (lock/unlock)                       |
| 2  | `WithServiceName(name)` option added                                                                              | `adapter.go:131-137` (option) + tests                                                        |
| 3  | `TestTrigger_ConcurrentCooldownIsRaceFree` — verifies mutex with 8 goroutines × 25 iterations under `-race`       | `adapter_test.go:333-365`                                                                    |
| 4  | `TestTrigger_WithServiceName_NotLoggedWhenEmpty` — verifies "service" attribute omitted when empty                | `adapter_test.go:304-331`                                                                    |
| 5  | All `fmt.Errorf` in adapter.go replaced with `errorfamily.NewRejection` / `NewInfrastructure`                     | `adapter.go:67-78`                                                                           |
| 6  | `go-error-family v0.10.0` added as a module dependency                                                            | `flightrecorderhealth/go.mod:6`                                                              |
| 7  | `flightrecorderhealth/README.md` created                                                                          | module overview, quick start, configuration, error taxonomy                                  |
| 8  | Spurious `GOEXPERIMENT=jsonv2` claim in AGENTS.md fixed                                                           | `AGENTS.md:19-20` (split into "six of seven" + "flightrecorderhealth does NOT")              |
| 9  | Build commands for flightrecorderhealth updated (no `GOEXPERIMENT=jsonv2`)                                        | `AGENTS.md:57-59`                                                                            |
| 10 | `flightrecorderhealth/.golangci.yml` created (module-specific lint config)                                        | mirrors go-flightrecorder exclusions; 0 lint issues                                          |
| 11 | `err113` test findings fixed — static error sentinels                                                             | `adapter_test.go:22-26` (`errTestConnectionRefused`, `errTestTimeout`, `errTestServiceDown`) |
| 12 | `noinlineerr` finding fixed                                                                                       | `adapter_test.go:401-405`                                                                    |
| 13 | `exhaustruct` on `fr.TriggerContext` resolved by populating `Duration`                                            | `adapter.go:188-193`                                                                         |
| 14 | `exhaustruct` on `Trigger{}` constructor with intentional `//nolint:exhaustruct` comment                          | `adapter.go:152`                                                                             |
| 15 | `nlreturn` (blank line before `return`) fixed                                                                     | `adapter.go:200`                                                                             |
| 16 | Duplicate package doc removed from `adapter.go` (kept in `doc.go` per project pattern)                            | `adapter.go:1`                                                                               |
| 17 | `golines` long-line violations fixed (max-len 120)                                                                | all lines now ≤120 chars                                                                     |
| 18 | `CHANGELOG.md` rewritten with full v0.1.0 content (concurrency safety, error taxonomy, GOEXPERIMENT note)         | `flightrecorderhealth/CHANGELOG.md`                                                          |
| 19 | `FEATURES.md` extended with flightrecorderhealth entry (9 rows)                                                   | `FEATURES.md:75-86`                                                                          |
| 20 | `TODO_LIST.md` updated for 7 modules + flightrecorderhealth work-in-progress state                                | `TODO_LIST.md:6`                                                                             |
| 21 | `AGENTS.md` Flightrecorderhealth section expanded (file table, lint config, WithServiceName, errors, concurrency) | `AGENTS.md:105-125`                                                                          |
| 22 | All 20 tests pass with `-race -count=1` (2.24s)                                                                   | `go test ./... -race -count=1`                                                               |
| 23 | All 20 tests pass with `-race -count=10` (13.26s, 200 invocations) — stability confirmed                          | `go test ./... -race -count=10`                                                              |
| 24 | `golangci-lint run ./...` → 0 issues                                                                              | `golangci-lint run ./...` (using module-local `.golangci.yml`)                               |
| 25 | `go vet ./...` clean                                                                                              | no output                                                                                    |
| 26 | `go build ./...` clean (no `GOEXPERIMENT`)                                                                        | no output                                                                                    |
| 27 | Full workspace `go build ./...` clean                                                                             | no output                                                                                    |
| 28 | All 6 sibling modules still pass `-race`                                                                          | cqrs, flightrecorder, realtime, docs, errorpages all `ok`                                    |

## b) PARTIALLY DONE

| # | Item                                                                                       | What's done                                                                      | What's missing                               |
| - | ------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------- | -------------------------------------------- |
| 1 | `flightrecorderhealth/v0.1.0` tag — code/docs/CHANGELOG are release-ready; tag not yet cut | All release prerequisites met; CHANGELOG cut; module-local `.golangci.yml` clean | Annotated tag at HEAD not yet created        |
| 2 | Status-report-only question items from predecessor — three questions remain unanswered     | —                                                                                | Decisions pending user input (see section g) |

## c) NOT STARTED

| # | Item                                                                                                                                                                     | Why it matters                                                                    |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------- |
| 1 | Consolidate `TestIntegration_TriggerWithFailingService` with `TestTrigger_CapturesOnHealthCheckFailure` (both assert trace written after a single failing-service batch) | Test hygiene; current duplication                                                 |
| 2 | `Trigger.Recorder()` accessor for symmetry with DiscordSync-style helpers                                                                                                | API ergonomics                                                                    |
| 3 | Add `ExampleRegister`, `ExampleNewCheckable`, `ExampleNewTrigger` Go example tests                                                                                       | pkg.go.dev rendering                                                              |
| 4 | `flightrecorder.BufferFull()` accessor on `go-flightrecorder` so `Checkable` can warn before overruns                                                                    | Requires upstream change                                                          |
| 5 | Probe-style wrapper that returns both `Checkable()` and `Trigger()` from one handle                                                                                      | Less duplication at the call site                                                 |
| 6 | Multi-module CI pipeline that runs each module's `.golangci.yml` separately                                                                                              | Currently `golangci-lint` only runs with root config (wrong for non-root modules) |
| 7 | Add a benchmark: `BenchmarkTrigger_RecordHealthCheckWithContext_AllPass` to verify no-capture path cost                                                                  | Performance baseline                                                              |
| 8 | Integration test with a real `go-health.Probe` (not just `do.Injector`)                                                                                                  | End-to-end wiring verification                                                    |

## d) TOTALLY FUCKED UP

| # | Item                                                                                                                                                                     | Impact                                                 |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------ |
| 1 | First sed-replacement attempt for error sentinels produced self-referential declarations (`errTestConnectionRefused = errTestConnectionRefused`), caught by the compiler | Wasted 1 build cycle; fixed by re-edit                 |
| 2 | Tried `// nolint:` (with leading space) before reading the `nolintlint` rule — emitted "directive should be written without leading space" errors                        | Wasted 1 lint cycle; fixed by switching to `//nolint:` |

## e) WHAT WE SHOULD IMPROVE

| # | Item                                                                                                                                                                                | Rationale                                                                                                          |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| 1 | Project should add a `golangci-lint` per-module runner script (the root config has depguard allowlists that are wrong for non-root modules; only module-local configs are correct). | Right now the project root lint is misleading — it reports depguard violations on every non-root module by design. |
| 2 | The `Register` convenience function is eager (calls `InvokeNamed` in the constructor). Some users find this surprising; document the contract in the doc.go example or remove it.   | API transparency                                                                                                   |
| 3 | Document the interaction between `WithSnapshotDir` (recorder) and `WithCooldown` (trigger): recommended cooldown values depend on sink type                                         | Operational guidance                                                                                               |
| 4 | Add a `WithOnCapture(func(SnapshotEvent))` option to `Trigger` for per-capture hooks independent of the recorder's own metrics hook                                                 | Extensibility                                                                                                      |
| 5 | Add a "When NOT to use this module" section to the README (e.g., single-check services)                                                                                             | Already done — see `flightrecorderhealth/README.md`                                                                |

## f) UP TO 50 THINGS TO GET DONE NEXT

1. Cut `flightrecorderhealth/v0.1.0` annotated tag at HEAD
2. Push the new tag (user gate — project policy says don't push without approval)
3. Run `pkg.go.dev` rendering check after push (10-min proxy propagation)
4. Add a "release notes" entry to the top-level `go-appkit/CHANGELOG.md` mentioning the new satellite module
5. Consolidate `TestIntegration_TriggerWithFailingService` with `TestTrigger_CapturesOnHealthCheckFailure`
6. Add `Trigger.Recorder()` accessor for ergonomics
7. Add `ExampleRegister`, `ExampleNewCheckable`, `ExampleNewTrigger`
8. Add `BenchmarkTrigger_RecordHealthCheckWithContext_AllPass`
9. Add `WithOnCapture(func(SnapshotEvent))` option
10. End-to-end integration test with real `go-health.Probe` (not just `do.Injector`)
11. Investigate `flightrecorder.BufferFull()` upstream addition
12. Update root `.golangci.yml` to add `flightrecorderhealth` to the depguard allowlist OR document the per-module config pattern in AGENTS.md
13. Add a CI script that runs each module's `.golangci.yml` separately
14. Add a status-report index entry for this report in `docs/status/INDEX.md`
15. Verify the v0.1.0 tag with `GOWORK=off go build ./...` (the hermetic consumer check)
16. Verify the v0.1.0 tag with a fresh `go get github.com/larsartmann/go-appkit/flightrecorderhealth@v0.1.0` from a clean module
17. Update the project root `README.md` module table to note v0.1.0 status
18. Consider adding a `flightrecorderhealth.WithOnHealthCheckPass` shortcut for users who want capture-on-success-only
19. Document the `WithServiceName` multi-trigger use case more prominently (it currently lives only in the README)
20. Add `t.Parallel()` exclusion note to `flightrecorderhealth/AGENTS.md` (the module doesn't have its own AGENTS.md; the note is in go-flightrecorder's)

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Tag now or wait for the release wave at `f938d65` to be pushed?** The other 4 tags from the wave are local-only and the project has a "push pending — user gate" stance. Should `flightrecorderhealth/v0.1.0` be tagged now and pushed together with the wave, or should it wait until the wave is pushed? Current code is release-ready.

2. **Should `flightrecorderhealth` ship its own example in `example/main.go`?** The errorpages and core modules have examples; flightrecorderhealth doesn't. The README quick start is sufficient for now, but a runnable example would surface the wiring more clearly.

3. **Should the `Register` convenience function be marked deprecated in favor of explicit `do.ProvideNamed`?** The eager invocation inside `Register` is a footgun (the constructor calls `InvokeNamed`). DiscordSync doesn't use a convenience; they wire manually. Keeping `Register` is more ergonomic but less transparent.
