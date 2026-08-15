# Status Report: SUPERB Ecosystem Execution — Session 5 (FINAL, plan complete)

**Date:** 2026-08-16 01:50
**Session scope:** Finishing M18 (the appkit integration spike in cqrs-htmx) and the plan-closing verification sweep. Sessions 1–4 covered M01–M17. With this session, **all 18 milestone groups of the SUPERB ecosystem plan are executed locally**. Everything below was verified in this session unless marked otherwise.

---

## Executive Summary

The plan (`docs/planning/2026-08-15_19-27_SUPERB-ECOSYSTEM-EXECUTION.md`) is **complete except for two items that require user approval**: pushing the go-appkit releases (commits + tags) and filing the templ-components PR. The M18 spike — the last open milestone — is written, tested, benchmarked, committed on cqrs-htmx `spike/appkit-server` (`8028bf2f`), and its verdict is **ADOPT**: ADR-001 Option A is confirmed on evidence, with adoption blocked only on publishing an appkit tag that carries `NoTimeout` and `ReadyCheck`.

One real finding this session: **all six go-appkit modules now require `GOEXPERIMENT=jsonv2`**, including core and flightrecorder, whose AGENTS.md entries claimed "no special flags". The stale commands are fixed in AGENTS.md (core via the `httputil/httpspec` test dependency; flightrecorder imports `encoding/json/v2` directly).

---

## a) FULLY DONE (this session)

1. **M18 spike completed end-to-end** on cqrs-htmx branch `spike/appkit-server` (commit `8028bf2f`):
   - Fixed the last compile error (benchmark closures captured `bundle` out of scope → restructured to bundle-taking method expressions `(*Bundle).RunHandler` / `(*Bundle).RunWithAppkit`, fresh bundle per sub-benchmark since both run funcs Close the bundle on every exit).
   - Fixed a second, subtler benchmark bug: `defer stop()` inside the measured function counted appkit's 2s drain toward `ns/op`, pinning `b.N` at 1 (2.0s/op for a 25µs request). Now `b.StopTimer()` runs before draining.
   - **M18.3 SSE flush:** PASS — headers flushed ≪400ms ahead of the first event, event body intact through appkit's outer + bundle's inner middleware stacks.
   - **M18.2 readiness:** PASS — `/health/ready` projection-aware (polls to 200 via `ReadyCheck`), clean drain + shutdown on context cancel, green under `-race`.
   - Response parity: PASS.
   - **M18.4 benchmark:** baseline-httputil 16178 ns/op vs appkit-service 45049 ns/op (~2.8x). Delta is dominated by appkit's per-request INFO logging; still ~22k req/s single-connection. Directional smoke number, honestly framed.
   - Full cqrs-htmx `setup` module: test `-race -count=1` + vet + build all green.
2. **M18.5 adopt/reject report:** cqrs-htmx `docs/status/2026-08-16_01-38_appkit-spike-adopt.md` — verdict ADOPT, evidence table, honest benchmark reading, four follow-ups. Harvested into cqrs-htmx `TODO_LIST.md` (P3, blocked on upstream tag).
3. **ADR-001 updated** (go-appkit `c728d1d`): status line records spike validation 2026-08-16; Consequences now carry the evidence summary and the single remaining blocker.
4. **Plan DoD ticked** (8 of 10): ADRs, cqrs-lite v4.3.0 line, projectionhost wiring, `/health/ready`, errorpages, journal→SSE bridge, docs currency, zero public API breaks. The two unticked items (tag push + templ-components PR) carry inline annotations stating exactly what exists locally and what blocks them.
5. **Six-module verification sweep:** core, cqrs, docs-mod, errorpages, flightrecorder, realtime — test (with `-race -count=1` where applicable) + vet + build all green, uniformly under `GOEXPERIMENT=jsonv2` (+ `GOWORK=off` for sub-modules).
6. **AGENTS.md corrected:** core and flightrecorder build commands now carry the jsonv2 flag with accurate per-module reasons.

## b) PARTIALLY DONE / BLOCKED ON USER

1. **Push go-appkit** (20+ commits on master, 3 local tags `cqrs/v0.2.0`, `docs/v0.2.0`, `errorpages/v0.1.0` at `e4a4e9d`). After push: `go mod tidy` sweep on sub-modules, true fresh-consumer `go get` against the proxy.
2. **templ-components PR**: fix + tests sit on local branch `fix/errorpage-orchestration-status` @ `c6df43c` (FamilyOrchestration → 500 in `errorpage/styles.go`, map-entry + end-to-end tests).
3. **cqrs-htmx branches**: `feat/transport-package` @ `ac743f30` and the new `spike/appkit-server` @ `8028bf2f` are local-only; merge/push/disposition is the user's call.

## c) NOT STARTED (deliberately)

1. dashboardui migration to `transport.NewJournalSSEStore` — marked at `dashboardui/dashboard.go:63` area; only sensible after the tags are pushed and resolvable.
2. appkit core v0.3.0 release carrying `NoTimeout` + `ReadyCheck` — prerequisite for folding the spike into cqrs-htmx `setup.RunHandler`.

---

## Honest mistakes ledger (this session)

1. The benchmark's first fixed draft still contained a second bug (drain counted in timing) — found only because the appkit number came back as exactly one 2s "operation". Lesson restated: read the number, not just PASS/FAIL.
2. Initially wrote a wrong AGENTS.md reason for flightrecorder's jsonv2 need ("dep chain"); `go list` showed the module imports `encoding/json/v2` directly. Corrected before commit.
3. Sessions 3–5 have now asked the three blocking questions (push, PR, cqrs-htmx disposition) three times with no answer; everything remains local-only by default.

## Verification commands (current truth)

```bash
# All modules: GOEXPERIMENT=jsonv2 required. Sub-modules additionally GOWORK=off.
GOEXPERIMENT=jsonv2 go test ./... -race -count=1 && GOEXPERIMENT=jsonv2 go vet ./... && GOEXPERIMENT=jsonv2 go build ./...
cd cqrs       && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1   # + vet/build
cd docs-mod   && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1   # + vet/build
cd errorpages && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1   # + vet/build
cd flightrecorder && GOEXPERIMENT=jsonv2 go test ./... -race -count=1          # + vet/build
cd realtime   && GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1   # + vet
# cqrs-htmx setup (spike):
cd /home/lars/projects/cqrs-htmx/setup && GOEXPERIMENT=jsonv2 GOWORK=off go test ./... -race -count=1
```
