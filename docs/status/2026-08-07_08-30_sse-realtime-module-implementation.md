# Status Report: SSE Realtime Module — Design + Implementation

> **Date:** 2026-08-07 08:30
> **Session scope:** Revised the realtime design from SSE+WebSocket to SSE-only, grounded in go-sse and go-datastar. Implemented the realtime module.
> **Verdict:** Design corrected, module implemented, all tests pass.

---

## a) FULLY DONE

### Realtime module implemented and tested

**`go-appkit/realtime/`** — new opt-in module (4th alongside core, cqrs, docs):

| File                | Lines | Concern                                                              |
| ------------------- | ----- | -------------------------------------------------------------------- |
| `go.mod`            | 9     | Module: `github.com/larsartmann/go-appkit/realtime`, dep: go-sse v0.4.0 |
| `doc.go`            | 35    | Package doc with quick start, DataStar pattern, SSE-only constraint  |
| `hub.go`            | 133   | `Hub` type, `Option` functional options, `BroadcastPatch` duck-typed |
| `handler.go`        | 135   | `Handler` (canonical SSE endpoint) + `Mount` + replay + heartbeat + CORS |
| `realtime_test.go`  | 400+  | 20 tests: Hub lifecycle, broadcast, fan-out, filter, replay, CORS, heartbeat |
| `docs/planning/realtime-sse-design.md` | 250+ | Revised design doc with API-by-API comparison to previous broken design |

### Design corrections from previous session

Every failure from the 2026-08-07 06:55 report is addressed:

| Previous Failure | Correction |
|---|---|
| Built custom Hub from scratch | Uses go-sse's `Broadcaster[T]` directly — no reimplementation |
| Fabricated `system.DefaultSQLiteDeployment` | No system dependency at all — realtime is CQRS-agnostic |
| Invented `errorfamily.Classify(err)` | No error classification needed — handler uses slog for logging |
| No CLI framework decision | Not needed — realtime module is HTTP-only, no CLI scope |
| Ignored locked design decisions | Design doc references Decisions 4, 9, 10 explicitly |
| No WebSocket plan | **WebSocket eliminated entirely. SSE only.** |
| No SSE library decision | **go-sse v0.4.0 chosen and verified** — already used by go-datastar and cqrs-htmx |
| No auth model | Documented: consumer's middleware, not module's concern |
| No CORS discussion | Implemented: `WithCORSOrigin` option, default `*` |
| No reconnect/replay design | Implemented: `WithStore` + go-sse's `Replay`/`ReplayFiltered` |

### All tests pass

```
GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1 -timeout 30s
ok  github.com/larsartmann/go-appkit/realtime  1.076s
```

20 tests covering:
- Hub: creation, store pairing, buffer size, broadcast, broadcastMany, broadcastPatch, subscriber count, health, shutdown, close, onSubscribe/onUnsubscribe
- Handler: serves events, CORS (default + custom), fan-out to multiple clients, filter predicate, replay on reconnect, no-replay without store, heartbeat sent, returns http.Handler

---

## b) PARTIALLY DONE

### DataStar integration — documented, not tested with real go-datastar

The `BroadcastPatch` method and `PatchLike` interface are designed to work with go-datastar patches, but no integration test imports go-datastar. The duck-typed interface means any `Event() sse.Event` type works, which is verified by `stubPatch` in tests.

### JournalStore adapter — designed, deferred

The CQRS Journal → `sse.EventStore` adapter (extracted from cqrs-htmx's `JournalSSEStore`) is documented in the design doc as a v0.2.0 deliverable. Deferred because it requires choosing a CQRS version (v3 vs v4), which is an unresolved question.

---

## c) NOT STARTED

1. **cqrs v3 → v4 migration** — still unresolved. The realtime module works regardless.
2. **CLI runtime** — not part of this session's scope (user focused on SSE-only realtime).
3. **App orchestrator type** — determined to be unnecessary. Hub + Mount compose with existing Service.
4. **Error renderer** — Decision 6 already defines boundary terminators. No new error rendering needed.
5. **Multi-instance realtime** — in-process broadcaster only. Documented as a limitation.

---

## d) WHAT WENT WELL

1. **Read go-sse and go-datastar source before designing** — every API used was verified against actual source code, not READMEs.
2. **Studied cqrs-htmx's EventBridge pattern** — proved the integration works before implementing.
3. **No core dependency** — realtime module mounts on any `*http.ServeMux`, not just appkit's Service.
4. **Caught the header-flush bug** — tests revealed that `sse.NewStream` doesn't flush headers; added explicit flush in handler.
5. **Kept it thin** — realtime module is ~270 lines of source code, leveraging go-sse's ~800 lines of transport.

---

## e) KEY DECISIONS MADE

1. **SSE only.** No WebSocket. This is a hard constraint from the user.
2. **go-sse v0.4.0 as the transport.** Already proven in go-datastar and cqrs-htmx.
3. **No go-datastar dependency.** `PatchLike` interface provides duck-typed compatibility.
4. **No go-cqrs-lite dependency.** Consumer wires domain events via closures.
5. **No go-appkit core dependency.** `Mount` works on any `*http.ServeMux`.
6. **Handler flushes headers immediately** after `NewStream` — critical for client connection confirmation.
7. **Replay aborts on error** — safer than continuing with partial replay + live events.

---

## f) REMAINING QUESTIONS

1. **Should cqrs migrate from v3 to v4?** — The realtime module works regardless. This affects only the cqrs module's future JournalStore adapter.
2. **Should we add a `realtime/example/` directory?** — A runnable demo like go-sse's example/datastar/ would showcase the integration.
3. **Should the core `Service` add an `OnShutdown` callback?** — Would simplify the consumer's shutdown wiring (currently manual: `hub.Shutdown(ctx)` then `svc.Shutdown(ctx)`).
