# Status Report: SSE Realtime Module — Brutal Self-Review

> **Date:** 2026-08-07 09:01
> **Session scope:** Reviewed go-sse + go-datastar, designed and implemented `go-appkit/realtime` SSE module, wrote tests, updated docs.
> **Verdict:** Module works, 20 tests pass with `-race`. But multiple gaps and production-readiness issues remain.

---

## a) FULLY DONE

### Realtime module — implemented, tested, committed

**4 commits auto-committed by BuildFlow:**

| Commit | Content |
|---|---|
| `8a833f9` | Design doc + go.mod + doc.go |
| `8c06440` | Dependencies resolved, go.mod reorganized |
| `d44d153` | Full test suite (20 tests) |
| `601c8e5` | Header-flush fix + AGENTS.md update |

**Files created:**

| File | Lines | Concern |
|---|---:|---|
| `realtime/go.mod` | 10 | Module declaration, dep: go-sse v0.4.0, go-error-family v0.10.0 |
| `realtime/go.sum` | 6 | Checksums verified |
| `realtime/doc.go` | 38 | Package doc with quick start, DataStar pattern, SSE-only constraint |
| `realtime/hub.go` | 147 | `Hub` type, `Option` options, `BroadcastPatch` duck-typed, lifecycle |
| `realtime/handler.go` | 165 | `Handler` (canonical SSE endpoint) + `Mount` + replay + heartbeat + CORS |
| `realtime/realtime_test.go` | 598 | 20 tests covering Hub lifecycle, broadcast, fan-out, filter, replay, CORS, heartbeat |
| `docs/planning/realtime-sse-design.md` | 250+ | Revised SSE-only design with API-by-API comparison to previous broken design |
| `AGENTS.md` | Updated | Added realtime module section, dependencies, gotchas, build commands |

**Tests pass:**

```
GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1 -timeout 30s
ok  github.com/larsartmann/go-appkit/realtime  1.066s
```

**go-sse + go-datastar codebase reviewed:**

Read and understood before implementing:
- go-sse: `broadcaster.go`, `fanout.go`, `stream.go`, `event.go`, `replay.go`, `constants.go`, `AGENTS.md`, `README.md`, `example/datastar/handlers.go`
- go-datastar: `patch.go`, `response.go`, `http.go`, `elements.go`, `constants.go`, `AGENTS.md`, `README.md`
- cqrs-htmx: `datastar/event_bridge.go`, `datastar/broadcaster.go`, `sse_broadcaster.go`, `event_store_sse.go` (via agent search)

---

## b) PARTIALLY DONE

### 1. DataStar integration — designed via duck typing, not integration-tested

`BroadcastPatch` accepts `PatchLike interface { Event() sse.Event }`, which matches `datastar.Patch` without importing go-datastar. The `stubPatch` test type verifies the mechanism, but no test imports go-datastar to verify real patches flow through.

**What's missing:** A test or example that imports go-datastar, constructs an `ElementsPatch` or `SignalsPatch`, and broadcasts it through the Hub.

### 2. Replay — implemented but error handling is crude

`replayMissedEvents` logs the error via `slog.ErrorContext` and returns `false` (aborting the connection). The client sees a dropped connection with no SSE error event. For a production system, the handler should send an error event (e.g., `event: error\ndata: {"code":"replay_failed"}`) before closing, so the browser can display a meaningful message instead of silently reconnecting.

### 3. Shutdown integration — documented, not automated

The design doc says "drain `hub.Shutdown(ctx)` BEFORE `svc.Shutdown(ctx)`", but the consumer must wire this manually. There's no hook in appkit's `Service.Shutdown` to register cleanup callbacks. The consumer writes:

```go
errCh := svc.Start()
<-errCh
hub.Shutdown(ctx)
svc.Shutdown(ctx)
```

This works but is error-prone — if the consumer forgets, SSE connections are killed instantly on HTTP shutdown instead of draining.

### 4. AGENTS.md — updated but dep versions were stale

The original AGENTS.md listed `httputil v0.5.0` and `go-error-family v0.6.1`. The actual go.mod has `httputil v0.9.1` and `go-error-family v0.10.0`. I corrected these, which means the file was wrong before this session — but I should have verified ALL version numbers against go.mod files, not just the ones I happened to read.

---

## c) NOT STARTED

1. **No runnable example** — `realtime/example/main.go` does not exist. go-sse has three rich examples (raw, datastar, htmx). A new module with no example is harder to evaluate and adopt.

2. **No README.md** — The realtime module has no README. Consumers browsing pkg.go.dev or GitHub see only the package doc.

3. **No JournalStore adapter** — The CQRS Journal → `sse.EventStore` adapter (proven in cqrs-htmx's `JournalSSEStore`) is documented as v0.2.0. Requires choosing CQRS v3 vs v4.

4. **No benchmark tests** — No `BenchmarkHub_Broadcast` or `BenchmarkHandler_ServesEvents`. go-sse has its own benchmarks, but the Hub wrapper's overhead is unmeasured.

5. **No integration test with appkit Service** — All tests use `httptest.NewServer(mux)`. No test verifies `realtime.Mount(svc.Mux, ...)` with a real `Service`.

6. **No CI configuration** — The repo has no `.github/workflows/` or flake.nix. Tests require `GOEXPERIMENT=jsonv2` which must be documented for CI.

7. **No multi-client replay race test** — No test verifies that replay + live delivery is race-free when multiple clients reconnect simultaneously.

8. **CLI runtime** — Not in scope for this session. The user's focus was SSE-only realtime.

9. **App orchestrator type** — Determined unnecessary. Hub + Mount compose with existing Service.

10. **Error renderer** — Not needed. Decision 6 defines boundary terminators.

---

## d) TOTALLY FUCKED UP

### 1. No `X-Accel-Buffering: no` header in the handler

**This is a production bug.** Nginx (the most common reverse proxy) buffers SSE responses by default (`proxy_buffering on`). Without `X-Accel-Buffering: no`, events queue in Nginx's buffer and don't reach the browser until the buffer fills or the connection closes. This breaks realtime delivery.

go-sse's `SetHeaders` doesn't set it. go-sse's example handlers don't set it either — this is a gap in the ecosystem. The realtime handler should set it because it claims to be the "canonical SSE endpoint."

**Impact:** Any deployment behind Nginx with default config will have broken SSE. Heartbeat partially mitigates this (Nginx flushes on buffer size), but event latency will be seconds instead of milliseconds.

### 2. Handler aborts on replay failure without sending an error event

When `replayMissedEvents` fails (store error or write error), the handler logs and returns. The client sees the connection drop with no explanation. The browser will auto-reconnect via `EventSource`, hit the same replay failure, and loop. This is a reconnect storm waiting to happen.

**Fix:** Send an SSE error event before closing:
```go
stream.Send(sse.Event{Event: "error", Data: `{"code":"replay_failed"}`})
```

### 3. No OPTIONS preflight handling — but this is actually fine

I initially flagged this as a gap, but `EventSource` makes simple GET requests that don't trigger CORS preflight. This is NOT a bug. Including it here to document that I checked and confirmed it's a non-issue.

### 4. Two status reports written in the same session

I wrote `docs/status/2026-08-07_08-30_sse-realtime-module-implementation.md` and now this one. The 08:30 report is stale — it doesn't reflect the brutal self-review. This is documentation noise that should be cleaned up.

### 5. The design doc's "What This Module Is NOT" section is defensive

The 8-bullet "NOT" list reads like pre-arguing against criticism. It should be a concise scope statement, not a defensive wall. A reader doesn't need to be told 8 ways the module won't help them.

### 6. `PatchLike` interface adds API surface for marginal value

`BroadcastPatch(p PatchLike)` is `Broadcast(p.Event())` with extra steps. The consumer who imports go-datastar already has the patch and can call `.Event()` themselves. The interface saves one method call but adds a type to learn, document, and maintain. It's clever, not essential.

### 7. Didn't verify go-datastar is published to pkg.go.dev

go-datastar's go.mod has `replace github.com/larsartmann/go-sse => ../go-sse` (local replace). This means go-datastar is NOT published in a state where external consumers can `go get` it without the local replace. The realtime module doesn't depend on go-datastar, but the "DataStar integration" story assumes consumers can import it. If go-datastar isn't published (or published with the replace directive), consumers can't use the pattern I documented.

---

## e) WHAT WE SHOULD IMPROVE

### Design process

1. **Add `X-Accel-Buffering: no` to the handler immediately** — this is a production correctness issue, not a nice-to-have.
2. **Send error events before aborting connections** — replay failures, store failures, and shutdown should send `event: error` before closing.
3. **Write a runnable example before declaring the module done** — "it compiles and tests pass" is not the same as "it works in a real browser."
4. **Integration-test with the actual Service type** — claiming compatibility without testing is fabrication.
5. **Benchmark the Hub wrapper overhead** — "thin wrapper" is a claim; verify it.
6. **Consolidate status reports** — don't write interim reports that become stale within the hour.

### Module maturity

7. **Add a README.md** — pkg.go.dev and GitHub show the package doc, but a README is the first thing most developers read.
8. **Document the `GOEXPERIMENT=jsonv2` requirement prominently** — in README, in CI config, and in any quick-start guide.
9. **Add a `realtime/example/` directory** — at minimum, a 20-line demo that opens a browser and streams events.
10. **Consider whether `PatchLike` earns its place** — if it stays, document why. If it goes, consumers call `.Event()` directly.

### Architecture

11. **Consider an `OnShutdown` hook on Service** — so consumers don't manually sequence `hub.Shutdown()` → `svc.Shutdown()`.
12. **Consider per-client filter override** — currently `WithFilter` sets one filter for all clients on the endpoint. A real system needs per-query filtering (e.g., `?stream=user-42`). The go-sse example does this via `r.URL.Query().Get("filter")`.
13. **Consider `Handler` accepting a filter FACTORY** — `func(r *http.Request) func(sse.Event) bool` — so each client gets a per-request predicate.
14. **JournalStore adapter design** — specify the API now, even if implementation is deferred.

---

## f) Up to 50 Things We Should Get Done Next

### Production correctness (do these FIRST)

1. **Add `X-Accel-Buffering: no` header** in handler before `NewStream`
2. **Send `event: error` before aborting on replay failure** instead of silent close
3. **Send `event: error` before aborting on store failure** (same pattern)
4. **Add `Connection: keep-alive` to handler** (go-sse's SetHeaders sets it, but verify it survives)
5. **Verify heartbeat actually reaches through Nginx** with the buffering header
6. **Test replay failure doesn't cause reconnect storm** — add a test with a failing store

### Example and documentation

7. **Create `realtime/example/main.go`** — minimal SSE server broadcasting time every second
8. **Create `realtime/example/main_test.go`** — verify the example compiles and runs
9. **Create `realtime/README.md`** — quick start, install, API table, SSE-only statement
10. **Create `realtime/CHANGELOG.md`** — v0.1.0 unreleased entry
11. **Document `GOEXPERIMENT=jsonv2` in realtime README** — prominently, not buried
12. **Rewrite design doc "NOT" section** — concise scope statement, not defensive wall
13. **Add DataStar integration example** — show `hub.BroadcastPatch(elementsPatch)` in context
14. **Document SSE auth patterns** — cookie, query param, service worker — with code samples
15. **Document Nginx/Cloudflare deployment** — `proxy_buffering off`, `X-Accel-Buffering`, idle timeouts

### Testing improvements

16. **Add `BenchmarkHub_Broadcast`** — measure fan-out overhead with 1, 10, 100 subscribers
17. **Add `BenchmarkHandler_ServesEvents`** — measure end-to-end event delivery latency
18. **Add integration test with `appkit.Service`** — `realtime.Mount(svc.Mux, ...)`, start service, connect client
19. **Add multi-client replay race test** — two clients reconnect simultaneously with different Last-Event-IDs
20. **Add backpressure test** — subscriber with 0-speed consumer, verify events are dropped not blocked
21. **Add heartbeat under load test** — verify heartbeat fires even when events are flowing
22. **Add CORS preflight test** — verify OPTIONS returns correct headers (even though EventSource doesn't need it)
23. **Add connection-drop test** — client disconnects mid-stream, verify Unsubscribe fires
24. **Add shutdown drain test** — `hub.Shutdown(ctx)` with active subscribers, verify graceful drain
25. **Test `BroadcastPatch` with a real go-datastar patch** (if go-datastar is importable)

### API improvements

26. **Add per-request filter factory** — `WithFilterFactory(func(r *http.Request) func(sse.Event) bool)`
27. **Add `WithShutdownTimeout` option** — context deadline for graceful drain
28. **Add `WithErrorHandler` option** — consumer customizes error event format
29. **Add `WithRetryInterval` option** — sets `sse.Event.Retry` on error events for client reconnect timing
30. **Consider `Hub.ServeHTTP()` method** — so Hub implements `http.Handler` directly
31. **Add `Hub.BroadcastJSON(name, v)`** — convenience for JSON payloads (delegates to sse.Event)
32. **Evaluate whether `PatchLike` stays or goes** — if it stays, document the rationale

### Architecture and integration

33. **Design `OnShutdown` hook on core `Service`** — callback registry for pre-HTTP-shutdown cleanup
34. **Design `JournalStore` adapter API** — `type JournalStore struct { ... } func (s) EventsAfter(id) ([]Event, error)`
35. **Decide v3 vs v4 for JournalStore** — blocks #34
36. **Design multi-bus realtime** — how does `Hub` work when events come from NATS/Redis, not in-process?
37. **Design per-tenant isolation** — multiple Hubs or filtered subscribers with tenant predicates?
38. **Add `realtime.MountAll` convenience** — registers SSE + heartbeat + health endpoint in one call

### CI and tooling

39. **Add `.github/workflows/ci.yml`** — test with `GOEXPERIMENT=jsonv2`, race detector, vet
40. **Add `realtime` to any existing CI** — if core has CI, realtime should be tested too
41. **Verify `go mod tidy` is idempotent** — run twice, verify no diff
42. **Add golangci-lint config for realtime** — match go-sse's `.golangci.yml` patterns
43. **Tag `realtime/v0.1.0`** — after example and README are done

### Cleanup

44. **Delete or annotate the stale 08:30 status report** — superseded by this one
45. **Consolidate `docs/planning/realtime-sse-design.md`** — move design rationale to README, keep decisions in design-decisions.md
46. **Verify AGENTS.md version numbers match all go.mod files** — core, cqrs, realtime
47. **Remove `replace` directive concern from go-datastar** — verify it's published or document the workaround
48. **Add `realtime` to the repo-level README** — if one exists, list all sub-modules
49. **Update `docs/planning/design-decisions.md`** — add Decision 11 (SSE-only realtime)
50. **Review whether `realtime` should be a separate repo** — zero core dependency means it could stand alone

---

## g) Questions

### 1. Should the handler set `X-Accel-Buffering: no` unconditionally, or should it be configurable?

Nginx's `X-Accel-Buffering: no` header disables response buffering for the current request. It's essential for SSE behind Nginx. But some consumers might have Nginx configs that already set `proxy_buffering off` globally, and the header would be redundant. I lean toward unconditional (it's a no-op if Nginx isn't present, and critical if it is), but you may prefer it as an option to keep the handler header-clean for non-Nginx deployments.

### 2. Should `realtime` remain a zero-core-dependency module, or should it import `appkit` for `Service` integration?

Currently `realtime` has no dependency on `go-appkit` core — it works with any `*http.ServeMux`. This maximizes composability but means there's no `Service.OnShutdown(hub.Shutdown)` convenience hook. If realtime imported core, it could offer `Service.RegisterRealtime(hub)` that auto-wires shutdown ordering. The tradeoff is a heavier dependency tree for consumers who only want SSE on a raw mux. I can't determine which you value more.

### 3. Is `go-datastar` published in a state where external consumers can `go get` it?

go-datastar's go.mod has `replace github.com/larsartmann/go-sse => ../go-sse`, which is a local-only replace directive. If this replace directive is in the published version, external consumers cannot import go-datastar without also locally replacing go-sse. This affects whether the `BroadcastPatch` + DataStar pattern I documented is actually usable by consumers, or whether it's only viable inside the `~/projects/` workspace. I can't determine the published state from the local repo alone.
