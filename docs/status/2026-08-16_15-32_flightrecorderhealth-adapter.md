# Status Report — flightrecorderhealth adapter module

**Date:** 2026-08-16 15:32
**Session:** `go-appkit/flightrecorderhealth` creation + integration

---

## a) FULLY DONE

| #  | Item                                                                                                         | Evidence                                |
| -- | ------------------------------------------------------------------------------------------------------------ | --------------------------------------- |
| 1  | New module `github.com/larsartmann/go-appkit/flightrecorderhealth` created                                   | `go-appkit/flightrecorderhealth/go.mod` |
| 2  | `go.work` updated to include `./flightrecorderhealth`                                                        | `go-appkit/go.work:9`                   |
| 3  | `Checkable` type implementing `do.HealthcheckerWithContext`                                                  | `adapter.go:21-84`                      |
| 4  | `Trigger` type implementing `health.HealthRecorder`                                                          | `adapter.go:93-200`                     |
| 5  | `Register` convenience function wiring `Checkable` into samber/do                                            | `adapter.go:241-260`                    |
| 6  | All 7 functional options (`WithCheckableName`, `WithTriggerFunc`, `WithTriggerLogger`, `WithCooldown`, etc.) | `adapter.go`                            |
| 7  | Package documentation (`doc.go`) with quick start, import aliasing guidance                                  | `doc.go`                                |
| 8  | `CHANGELOG.md` with v0.1.0 entry                                                                             | `CHANGELOG.md`                          |
| 9  | 18 tests covering Checkable, Trigger, Register, nil-safety, integration                                      | `adapter_test.go`                       |
| 10 | All tests pass with `-race -count=1`                                                                         | `go test ./... -race` → 18 PASS, 2.03s  |
| 11 | `go vet ./...` clean                                                                                         | no output                               |
| 12 | Full workspace `go build ./...` clean                                                                        | no output                               |
| 13 | `README.md` module table updated with new module row                                                         | `README.md:69-70`                       |
| 14 | `AGENTS.md` updated: module list, GOEXPERIMENT note (6→7), build commands, code organization section         | `AGENTS.md`                             |
| 15 | Proper async capture drain in tests (warmup sleep + MinAge/MaxBytes)                                         | `adapter_test.go:54-65`                 |
| 16 | Nil-safe `Trigger` and `Checkable` (both pass-through when nil)                                              | `adapter.go:64-66, 169-171`             |

## b) PARTIALLY DONE

| # | Item                                                                                                                                                                                                                                    | What's done                             | What's missing                                  |
| - | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------- | ----------------------------------------------- |
| 1 | `Register` function passes the recorder pointer but the `Checkable` is captured in the closure — if the recorder pointer is mutated, `Register` won't see it. This is fine for the singleton pattern but the contract isn't documented. | Function works correctly for normal use | No doc note about the closure capture semantics |
| 2 | The `cooldown` timer field is not goroutine-safe — concurrent health-check batches could race on `t.lastCapture`. Health probes are single-goroutine in `go-health`, but this is an undocumented assumption.                            | Field exists and works in serial tests  | No `sync.Mutex` or atomic; not documented       |

## c) NOT STARTED

| # | Item                                                                                                                                                                                                                                        | Why it matters                                                                  |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| 1 | `go-appkit/flightrecorderhealth/README.md`                                                                                                                                                                                                  | Every other module has one; this one doesn't                                    |
| 2 | Linting with `golangci-lint` (project root has `.golangci.yml`; never ran it for this module)                                                                                                                                               | Linter config covers `paralleltest`, `varnamelen`, etc. — might surface issues  |
| 3 | Example wiring in `example/main.go` or a new example showing flightrecorder + health + dashboard together                                                                                                                                   | DiscordSync has this; this module's value would be more visible with an example |
| 4 | `GOEXPERIMENT=jsonv2` dep note: the module does NOT actually import `encoding/json/v2`. The AGENTS.md note says it requires jsonv2 "via go-health → samber/do" — this is incorrect. go-health and samber/do are both plain `encoding/json`. | Spurious build requirement added to AGENTS.md                                   |
| 5 | Release tag (`flightrecorderhealth/v0.1.0`) — CHANGELOG entry exists but no annotated tag                                                                                                                                                   | Not part of the wave at `f938d65`                                               |
| 6 | `TODO_LIST.md` / `FEATURES.md` updates for the new module                                                                                                                                                                                   | Project hygiene                                                                 |

## d) TOTALLY FUCKED UP

| # | Item                                                                                                                                                                                                                                                                                                                                                                                                                        | Impact                                                                                      |
| - | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| 1 | **Double-brace syntax errors introduced in `adapter_test.go` by sloppy edit replacement** (5 occurrences: `^}}` instead of `}`). Caught by compiler, fixed via `sed -i`. Wasted 2 build cycles.                                                                                                                                                                                                                             | Wasted time; would have been invisible without `-race -v`                                   |
| 2 | **Initial `Trigger` design only called `SnapshotIfAsync` when `hasFailures(results)` was true.** This silently broke `fr.OnAlways()` — a user wiring `OnAlways` with all-healthy services would see zero captures. Caught by `TestTrigger_CustomTriggerFunc`. Fix: removed the `hasFailures` early-return; now always evaluates the trigger. The `tc.Err` field carries the first error so `OnError` still fires correctly. | API would have been a footgun for `OnAlways` users                                          |
| 3 | **`tc.Err` was never populated in the first version of `Trigger`.** Without it, `fr.OnError()` evaluates `tc.Err != nil` → false → never fires. Silent: every "capture on failure" user would see zero captures. Caught by `TestTrigger_CapturesOnHealthCheckFailure`. Fixed by adding `firstError(results)`.                                                                                                               | The primary use case of the module would not have worked                                    |
| 4 | **Spurious `GOEXPERIMENT=jsonv2` claim in AGENTS.md.** The module imports nothing from `encoding/json/v2`. The transitive chain doesn't reach jsonv2 either. The build works without `GOEXPERIMENT=jsonv2` set. I added it because the other modules need it.                                                                                                                                                               | Misleading documentation; future agents will think they need the experiment for this module |
| 5 | **Test file used a custom `mockInjector` struct that did not implement the full `do.Injector` interface.** Compilation failed on first test run. Fixed by switching to real `do.New()` + registered services.                                                                                                                                                                                                               | Wasted a full test cycle                                                                    |
| 6 | **The `timer` field and `WithNowFunc` option were never used.** `Checkable.HealthCheck` doesn't reference time. Caught on final review and removed.                                                                                                                                                                                                                                                                         | Dead code shipped to the first build                                                        |

## e) WHAT WE SHOULD IMPROVE

| #  | Item                                                                                                                                                                                                                                                                                                                                        | Rationale                                                                                         |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| 1  | Add a `sync.Mutex` around `Trigger.lastCapture` writes/reads. Health probes are single-goroutine today, but the `HealthRecorder` interface is documented as batch-callable, and `WithCooldown` + concurrent callers would race.                                                                                                             | Robustness; the cooldown contract is "minimum duration between captures" — races could violate it |
| 2  | Add a `BufferFull()` / `MaxBytes` accessor to go-flightrecorder so `Checkable` can warn before the buffer overruns. Currently `Checkable` only checks `Enabled()`.                                                                                                                                                                          | `Checkable` can't distinguish "healthy running" from "healthy but silently dropping data"         |
| 3  | Make the `Register` signature consistent with `go-health-dashboard`'s `Register(injector, probe, opts...)` — currently it takes positional `(injector, rec, name, opts...)`. The `name` should fold into an option.                                                                                                                         | API consistency with sibling modules                                                              |
| 4  | Add `WithServiceName(name)` to `Trigger` for symmetry with `Checkable.WithCheckableName` — currently the trigger logs `failed_services` but not which trigger fired.                                                                                                                                                                        | Observability                                                                                     |
| 5  | Consider a `Probe.Recorder()` accessor pattern (like DiscordSync does) so the integration wraps the recorder once and exposes both `Checkable` and `Trigger` from the same handle.                                                                                                                                                          | Less duplication at the call site                                                                 |
| 6  | Add `golangci-lint` to the verification loop for new modules. The project's `.golangci.yml` enables 90+ linters — they should be run before declaring complete.                                                                                                                                                                             | Code quality gate consistency                                                                     |
| 7  | The `Trigger` doc comment claims it "captures a trace snapshot when health checks fail" but the implementation actually always evaluates the trigger (which may or may not fire). The doc should match the behavior.                                                                                                                        | Accuracy                                                                                          |
| 8  | Replace `fmt.Errorf` with `go-error-family` constructors (the rest of `go-appkit` uses error-family throughout — `adapter.go` is inconsistent).                                                                                                                                                                                             | Project consistency                                                                               |
| 9  | The integration test `TestIntegration_TriggerWithFailingService` duplicates `TestTrigger_CapturesOnHealthCheckFailure` almost exactly. Consolidate or differentiate (e.g., one using `Register`, the other using raw `NewTrigger`).                                                                                                         | Test hygiene                                                                                      |
| 10 | Document the `Recorder` singleton constraint interaction with `Checkable.Enabled()`: if another package already started a recorder, `Start()` returns `ErrAlreadyEnabled` and `Enabled()` will report false on _this_ recorder (since it never started). The dashboard would show this as unhealthy even though the process HAS a recorder. | Correctness of the dashboard signal                                                               |

## f) UP TO 50 THINGS TO GET DONE NEXT

1. Create `flightrecorderhealth/README.md` (module overview, install, quick start)
2. Fix the spurious `GOEXPERIMENT=jsonv2` line in AGENTS.md (line 18 area)
3. Run `golangci-lint run ./flightrecorderhealth/...` and fix any findings
4. Add `WithServiceName` option to `Trigger`
5. Make `Trigger.lastCapture` goroutine-safe with `sync.Mutex`
6. Replace `fmt.Errorf` with `go-error-family` constructors in `adapter.go`
7. Add `WithCheckableNowFunc` (or remove the dropped `WithNowFunc` properly — already done, just verify it's gone everywhere)
8. Consolidate `TestIntegration_TriggerWithFailingService` with `TestTrigger_CapturesOnHealthCheckFailure`
9. Fix the misleading "capture when health checks fail" doc on `Trigger` to match the always-evaluate behavior
10. Add a `BufferFull()` accessor to go-flightrecorder so `Checkable` can warn before overruns
11. Add `Probe`-style helper that wraps recorder once and exposes `Checkable()` + `Trigger()` accessors
12. Create an example: minimal HTTP server with flightrecorder + health + dashboard + flightrecorderhealth adapter
13. Tag `flightrecorderhealth/v0.1.0` annotated at HEAD
14. Add `flightrecorderhealth` to `go-appkit`'s top-level `TODO_LIST.md`
15. Add `flightrecorderhealth` entry to `go-appkit`'s `FEATURES.md`
16. Verify the module works when the recorder is `nil` after construction (re-nil-ing should pass-through, currently `Checkable` errors and `Trigger` passes through — this asymmetry may be surprising)
17. Add a benchmark: `BenchmarkTrigger_RecordHealthCheckWithContext_AllPass` to ensure the no-capture path is cheap
18. Add an integration test using `go-health.Probe` + `flightrecorderhealth.Trigger` end-to-end (not just `do.Injector`)
19. Verify `WithCooldown` doesn't fire on the very first capture (it shouldn't, but the code reads `!t.lastCapture.IsZero()` which is correct — add a test for it)
20. Add a `SnapshotEvent` callback option so consumers can observe captures from the Trigger path (currently the recorder's own `WithMetrics` is the only observer)
21. Document the interaction between `WithSnapshotDir` (recorder) and the cooldown option (recommended 30s–60s for dir sinks)
22. Add a `WithOnCapture(func(SnapshotEvent))` option to `Trigger` for per-capture hooks independent of the recorder's own metrics hook
23. Add `go mod tidy` verification to the release checklist for this module
24. Verify `go-health v0.0.2` is the actual latest version (we pin to it; if there's a v0.0.3 with bugfixes, the consumer needs to bump)
25. Verify the module path `github.com/larsartmann/go-appkit/flightrecorderhealth` doesn't conflict with any existing module on pkg.go.dev
26. Run `go test ./flightrecorderhealth/... -race -count=10` to catch flakiness (currently only `-count=1`)
27. Add `go test -cover` target and verify coverage meets project standard (~80%+ per other modules)
28. Consider an `interface{}` export so consumers can type-assert the `Trigger` to `health.HealthRecorder` without importing both — currently they must import `health` to use it
29. Add a "When NOT to use this module" section to the README (e.g., when only one health check exists, the cooldown makes the adapter a no-op)
30. Document that `Register` is eager: it invokes the service in the constructor. If the caller wants lazy registration, use `ProvideNamed` manually with `NewCheckable`.
31. Add `TestCheckable_Name_NilReceiver` — currently `Name()` handles nil but `HealthCheck` also handles nil; the symmetry should be tested
32. Investigate whether `SnapshotToWriter` (recorder's escape hatch) should be exposed via the `Trigger` for HTTP-download captures
33. Add a `Trigger.Recorder()` accessor for symmetry with the `Recorder()` accessor pattern used in DiscordSync
34. Ensure `go vet -all ./flightrecorderhealth/...` is clean (currently only `go vet` ran)
35. Check `gofmt -l ./flightrecorderhealth/` and `goimports -l ./flightrecorderhealth/`
36. Add a test that verifies `Checkable` works when registered via `do.ProvideNamed` directly (not through `Register`) — to prove both paths work
37. Add a test that verifies `Trigger` works when the injector contains zero services — should pass-through cleanly
38. Add a test that verifies the recorder's `Stop()` is idempotent when `Trigger` has captured during shutdown (the existing `SnapshotIfAsync` shutdown handling should cover this, but no test proves it for our adapter)
39. Consider exposing `WithContext` variants of all options for consistency with the recorder's context-aware API
40. Add CI-specific verification: `GOWORK=off go test ./flightrecorderhealth/...` works (other modules are tested this way in CI)
41. Verify the `go.sum` is reproducible across machines
42. Add a note in `doc.go` about the `Recorder` singleton constraint specifically for `health-recorder` users (who might also wire the recorder for HTTP middleware — they'd hit `ErrAlreadyEnabled`)
43. Test with Go race detector under `-count=100` to catch timing-sensitive bugs in the async capture path
44. Profile `RecordHealthCheckWithContext` to ensure the no-capture path doesn't allocate (currently constructs a `fr.TriggerContext` struct every call)
45. Add `ExampleRegister`, `ExampleNewCheckable`, `ExampleNewTrigger` to the package (Go example tests for pkg.go.dev rendering)
46. Consider a `Shutdown()` method on `Trigger` that drains any in-flight async captures explicitly (currently relies on `recorder.Stop()` being called by the caller)
47. Verify the adapter works correctly when wired into a `go-health-dashboard.Dashboard` (currently only verified against the raw `health.Probe`)
48. Document the dependency direction: `flightrecorderhealth` depends on `go-health` (not the other way around) — this is important for modularity
49. Run `go test ./flightrecorderhealth/... -race -timeout 60s` to ensure no infinite-loop bugs in the async capture path
50. Plan a v0.2.0 that adds the `BufferFull()` accessor to the recorder (requires upstream change) and `WithOnCapture` to the trigger

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Should `flightrecorderhealth` require `GOEXPERIMENT=jsonv2`?** I assumed yes (because other go-appkit modules do), but the module imports nothing from `encoding/json/v2` and `go-health` / `samber/do` are plain `encoding/json`. Do you want me to (a) leave the AGENTS.md note as-is for build-environment uniformity, (b) correct the doc and accept that this module builds without jsonv2, or (c) add an `encoding/json/v2` import somewhere to justify the flag?

2. **Should the `Trigger` log message include the trigger function name (`fr.OnError`, `fr.OnLatency`, etc.) or is the current "triggered by health check" sufficient?** This is a logging-style preference that affects the dashboard/SIEM integration downstream — I'd default to including it but the project convention is unclear.

3. **Is `Register` worth keeping as a convenience, or should it be removed in favor of explicit `do.ProvideNamed` + `NewCheckable`?** The convenience hides the eager invocation in the constructor (it calls `InvokeNamed` internally), which some users find surprising. DiscordSync doesn't use it; they wire manually. Keeping it is more ergonomic but less transparent.
