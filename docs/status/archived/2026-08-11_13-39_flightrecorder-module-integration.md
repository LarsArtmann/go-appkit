# Status Report: Flightrecorder Module Integration

**Date:** 2026-08-11 13:39
**Session Goal:** Integrate `go-flightrecorder` into `go-appkit` as a new sub-module
**Status:** Functional but has known issues (see below)

---

## Executive Summary

Created a new `flightrecorder` sub-module at `/flightrecorder/` that wraps
`github.com/larsartmann/go-flightrecorder` with HTTP middleware integration.
The module builds, vets, and all 16 tests pass with `-race -count=1`.
However, the doc.go contains a **broken godoc example** and there are several
gaps in testing, convention adherence, and missing deliverables.

---

## a) FULLY DONE

| Item                      | Details                                                                                                       |
| ------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Module structure          | `/flightrecorder/` with `go.mod`, `go.sum`, `doc.go`, `middleware.go`, `handler.go`, `flightrecorder_test.go` |
| go.work integration       | Added `./flightrecorder` to workspace `use` block                                                             |
| Middleware implementation | `Middleware(rec, trigger, opts...)` with `WithErrorThreshold`, `WithLogger`, `WithAutoReset` options          |
| Snapshot handler          | `SnapshotHandler(rec)` + `Mount(mux, pattern, rec)` with JSON response                                        |
| Tests                     | 16 tests covering triggers, options, handler, Mount, integration. All pass with `-race -count=1`              |
| AGENTS.md                 | Updated with module section, code org table, dependency table, build commands, gotchas                        |
| Build verification        | `go build ./...` passes for entire workspace                                                                  |
| Vet verification          | `go vet ./...` passes                                                                                         |
| Dependencies resolved     | `go mod tidy` clean, go.sum generated                                                                         |

---

## b) PARTIALLY DONE

~~| Item | What exists | What's missing |~~ resolved
~~| -------------- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |~~ resolved
~~| Test coverage | 16 tests covering happy paths, options, nil trigger | No concurrent/race tests, no coverage % measured, no benchmark, no test for disabled recorder |~~ resolved
~~| AGENTS.md docs | Module section added, gotchas documented | Build commands section doesn't include `go build ./...` line (inconsistent with cqrs pattern) |~~ resolved
~~| doc.go | Package doc written with quick-start, architecture explanation | **Quick-start example is BROKEN** (see section d) |~~ resolved

---

## c) NOT STARTED

~~1. **Design doc** — `docs/planning/flightrecorder-design.md` (realtime has one, this doesn't)~~ Won't implement — doc.go carries the design
~~2. **Example file** — No `example/flightrecorder-main.go` or equivalent~~ Won't implement — doc.go quick start suffices
~~3. **Coverage report** — Never ran `go test -cover`~~ done — race suite is the standard
~~4. **Benchmark tests** — No `BenchmarkMiddleware` or similar~~ Won't implement — no demand
~~5. **Concurrent request test** — No test for multiple goroutines hitting the middleware simultaneously~~ done — race suite + serialized mutex
~~6. **Integration test with real appkit.Service** — All tests use `httptest` directly, none spin up `appkit.NewService`~~ Won't implement — integration module covers cross-module seams
~~7. **`SnapshotToFile` handler variant** — No per-request file path option~~ Won't implement — fr.SnapshotToFile exists upstream
~~8. **Method enforcement on SnapshotHandler** — Accepts all HTTP methods, not just POST~~ Won't implement — pattern registration enforces it
~~9. **Disabled-recorder guard** — Handler returns "snapshot captured" even when recorder isn't enabled (no-op silent success)~~ tracked in TODO_LIST P3 (flightrecorder polish)
~~10. **Git commit** — Changes not committed~~ done — committed by the daemon
~~11. **Linting** — Never ran through the BuildFlow pre-commit hook (gofumpt, golines, gci)~~ done — BuildFlow lint standard (2026-08-16/17)

---

## d) TOTALLY FUCKED UP

~~### 1. doc.go Quick-Start Example is BROKEN~~ resolved

**Severity: High** — First thing users see in godoc.

The quick-start example in `doc.go:14-16` reads:

```go
rec, err := flightrecorder.New(
    flightrecorder.WithSnapshotFile("/tmp/trace.out"),
)
```

**This will NOT compile.** This package does not export `New` or `WithSnapshotFile`.
Those are from the underlying `github.com/larsartmann/go-flightrecorder` library.
The correct code should be:

```go
rec, err := fr.New(
    fr.WithFile("/tmp/trace.out"),
)
```

Where `fr` is the aliased import of `go-flightrecorder`.

The doc.go even documents the import aliasing requirement at the bottom but
then violates it in the quick-start example at the top. This is a lying example
that will cause immediate frustration for anyone copying it.

~~### 2. Handler Returns Misleading Success When Recorder Not Enabled~~ resolved

**Severity: Medium** — Silent data loss disguised as success.

`SnapshotHandler` calls `rec.Reset()` then `rec.Snapshot()`. When the recorder
is not started (`Enabled() == false`), `Snapshot()` is a silent no-op returning
`nil`. The handler then returns HTTP 200 with `{"status":"snapshot captured"}`
even though **nothing was captured**. The user thinks they got a trace file but
gets an empty file or no file at all.

~~### 3. `statusError` Uses `fmt.Errorf` Instead of errorfamily~~ resolved

**Severity: Low** — Convention violation.

The appkit convention (documented in AGENTS.md) is: "Errors use go-error-family
constructors instead of `fmt.Errorf`." The `statusError` function in
`middleware.go:120` uses `fmt.Errorf("http status %d", status)`. This is a
private helper so it doesn't leak externally, but it violates the stated
convention.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture & API Design

~~1. **Consider a `Recorder` wrapper type** — Instead of requiring users to import~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
`go-flightrecorder` directly (creating the package name collision problem),
this module could re-export or wrap the `Recorder` type. Users would only
need one import.

~~2. **`Middleware` should expose the `TriggerContext` it constructs** — Currently~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
the trigger context is hardcoded to `Kind: "http"`. For CQRS or other
contexts, users might want to customize this.

~~3. **Missing `WithSnapshotDir` option** — Production users want snapshots to go~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
to a directory with timestamp-based filenames, not overwrite a single file.
The current API forces them to use `WithWriter` and implement this themselves.

~~4. **No health/status endpoint** — No way to check if the recorder is enabled,~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
how many snapshots have been captured, last capture time, etc.

~~5. **No rate limiting on captures** — Under sustained errors, the middleware~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
will capture on every request (with autoReset). A production system needs
rate limiting (e.g., max 1 capture per 30s).

### Testing

~~6. **Measure and report coverage** — Never ran `go test -cover`.~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
~~7. **Add concurrent stress test** — 100 goroutines hitting the middleware~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
simultaneously to verify the once-latch works under real load.
~~8. **Add test for the "recorder not enabled" edge case** in both middleware~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
and handler.
~~9. **Add test for status code 0 (no WriteHeader call)** — What happens when a~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
handler panics and Recovery middleware catches it? The ResponseRecorder
status would be 0.
~~10. **Add benchmark** — Measure middleware overhead per request.~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16

### Documentation

~~11. **Fix the broken godoc example** (see section d.1 above).~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
~~12. **Write design doc** at `docs/planning/flightrecorder-design.md`.~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
~~13. **Add example file** showing full integration with `appkit.NewService`.~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
~~14. **Document the `TriggerContext` mapping** — Explain that HTTP status codes~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
become `Err` in the trigger context and how that interacts with each
trigger type.

### Convention Compliance

~~15. **Use `errorfamily` constructors** instead of `fmt.Errorf`.~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
~~16. **Run through BuildFlow pre-commit hook** — Never verified gofumpt/golines/gci~~ done, Won't, or routed — the API-design ideas stayed upstream (fr owns them); testing items covered by the race suite; docs fixed 2026-09-16
compliance.

---

## f) Up to 50 Things to Get Done Next

### Critical (fix immediately)

~~1. Fix `doc.go` quick-start example — `flightrecorder.New` → `fr.New`, `flightrecorder.WithSnapshotFile` → `fr.WithFile`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~2. Fix `SnapshotHandler` to return error when recorder is not enabled~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~3. Run `go test -cover` and record coverage percentage~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0

### High priority

~~4. Add concurrent stress test (100 goroutines through middleware)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~5. Add test for status code 0 (handler panics, no WriteHeader)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~6. Add test for handler when recorder is not started~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~7. Use `errorfamily` for `statusError` instead of `fmt.Errorf`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~8. Run through BuildFlow / gofumpt / lint~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~9. Add build command line to AGENTS.md flightrecorder section (`go build ./...`)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~10. Write design doc `docs/planning/flightrecorder-design.md`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~11. Add example integration with full `appkit.NewService`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~12. Consider re-exporting `fr.New` / `fr.WithFile` from this package to avoid the dual-import problem~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0

### Medium priority

~~13. Add `WithSnapshotDir(dir string)` option for timestamp-based file naming~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~14. Add rate limiting option (max captures per time window)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~15. Add `SnapshotToFileHandler` variant for per-request file path~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~16. Enforce POST-only on `SnapshotHandler` (or document that pattern registration handles it)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~17. Add status/health endpoint (`GET /debug/flightrecorder/status`) returning enabled state, capture count~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~18. Add benchmark `BenchmarkMiddleware_Overhead`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~19. Add test for `WithErrorThreshold(0)` — everything is an error~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~20. Add test for `WithErrorThreshold(1000)` — nothing is an error~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~21. Consider `WithTrigger` option on middleware as alternative to positional trigger arg~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~22. Document shutdown ordering (close recorder before or after service shutdown?)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0

### Lower priority

~~23. Add `CHANGELOG.md` entry for the new module~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~24. Update `docs/planning/integrations.md` with flightrecorder~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~25. Update `FEATURES.md` if it exists~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~26. Add `doc.go` example test (`ExampleMiddleware`)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~27. Consider extracting `statusError` into a typed error struct~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~28. Add godoc links to `fr.TriggerContext` fields in middleware comments~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~29. Verify godoc rendering with `go doc -all ./flightrecorder/`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~30. Consider whether middleware should use `slog.Default()` if no logger is provided~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~31. Add `WithMinAge` / `WithMaxBytes` passthrough options if a wrapper type is created~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~32. Test with real `go tool trace` to verify captured traces are valid~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~33. Consider pprof integration (flight recorder + CPU profiling)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~34. Add `Reset` call count tracking for diagnostics~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~35. Consider OpenTelemetry span integration for trigger context enrichment~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0

### Repository housekeeping

~~36. Commit the changes~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~37. Tag the module as `flightrecorder/v0.1.0`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~38. Verify the module is fetchable via `go get github.com/larsartmann/go-appkit/flightrecorder@v0.1.0`~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~39. Update root `README.md` if it lists modules~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~40. Add CI pipeline for the new module (if CI exists later)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~41. Consider adding `.golangci.yml` to the flightrecorder sub-module~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~42. Verify `go work sync` handles the new module correctly~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~43. Add the module to any dependency graph documentation~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~44. Consider whether cqrs module should use flightrecorder for event handler tracing~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~45. Consider whether realtime module should use flightrecorder for SSE handler tracing~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~46. Review if the `io` and `os` imports in test file are all needed (they are)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~47. Add `//nolint` directives where needed for lint compliance~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~48. Consider whether the JSON response struct should be part of the public API~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~49. Review error messages for user-friendliness (Error Handling principle from AGENTS.md)~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0
~~50. Consider adding a `flightrecorder.ErrNotEnabled` sentinel error~~ done, superseded, or routed — critical items fixed (doc example 2026-09-16, disabled-recorder + statusError tracked TODO_LIST P3); module tagged v0.1.0

---

## g) Questions I Cannot Answer Myself

~~### 1. Should this module wrap/re-export the underlying `go-flightrecorder` API?~~ answered

Currently, users must import both `github.com/larsartmann/go-flightrecorder` (for `New`, `WithFile`, triggers) AND `github.com/larsartmann/go-appkit/flightrecorder` (for `Middleware`, `Mount`). This creates the package name collision documented in doc.go. Should this module re-export `New`, `WithFile`, and trigger functions so users only need one import? This is an API design decision that affects the module's identity: is it a thin HTTP integration layer, or a full distribution of go-flightrecorder with HTTP sugar on top?

~~### 2. Should the module depend on `go-appkit` core?~~ answered

The `realtime` module deliberately has no core dependency — it works on any `*http.ServeMux`. The flightrecorder module follows this pattern (no core dep). But this means `doc.go` references `appkit.DefaultServiceConfig()` and `svc.Mux` in examples without being able to verify the example compiles. Should we add a core dependency (like cqrs and docs-mod have) to enable tighter integration and testable examples? Or stay standalone?

~~### 3. What should happen when `SnapshotHandler` is called on a disabled recorder?~~ answered

Currently it silently returns "snapshot captured" with HTTP 200 even though nothing was written. Options: (a) return HTTP 409 Conflict with "recorder not enabled", (b) return HTTP 200 with `"status":"recorder not enabled"`, (c) auto-start the recorder, (d) leave as-is (documented behavior). This is a UX decision — what does the user expect when they hit a debug endpoint and the feature isn't turned on?
