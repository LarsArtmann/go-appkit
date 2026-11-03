# go-sse — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). The otel-deep-dive
> (2026-10-18) and sibling audits do not cover go-sse; this is the first dedicated utilization
> report. The SSE-hub fleet-coherence watch (realtime vs go-aichat/ssehub) lives in TODO_LIST.

## Verdict

**USED — realtime is a purpose-built wrapper over go-sse at the latest published pins
(v0.6.2 / ssetest v0.5.0 / sseparse v0.2.1), and the wrapper leverages the hard parts
exactly as designed: graceful `Shutdown` drain, `ReplayFiltered` push-down, heartbeats,
health snapshot, subscriber-count sync (CI-guarded), and the ssetest harness. Score
81/100 (high).** The two honest gaps are already on the books: `WithOnDrop`
(drop-visibility counters, TODO W5-C1) and `OnDisconnect` (disconnect metrics). One
upstream coupling to watch: go-sse's next release carries a breaking
`ssetest.WithLastEventID` signature change and an exported `DefaultSubscriberBuffer`
that realtime's docs currently reference as an unexported const.

| Fact | Value | Evidence |
|---|---|---|
| In-graph versions | go-sse v0.6.2, ssetest v0.5.0, sseparse v0.2.1 — all = latest published tags | `realtime/go.mod:6-7,14`, `integration/go.mod:31-32` |
| Direct importers | realtime (runtime), integration (E2E), cqrs (indirect) | go.mods |
| Latest-tag vs HEAD | v0.6.2 published; HEAD 74 commits ahead — `[Unreleased]` has the breaking ssetest change + `DefaultSubscriberBuffer` export | `~/projects/go-sse/CHANGELOG.md` head |
| Protocol guard | `scripts/check-sse-protocol.sh` (CI + pre-commit) enforces SubscriberCount-sync + timeout-client house rules | AGENTS "Load-flake protocol" |

## 1. Usage evidence

**Hub (realtime/hub.go)** — a thin lifecycle wrapper, per its own doc: pairs
`sse.Broadcaster[sse.Event]` with an optional `sse.EventStore` (`hub.go:19-24`); constructs
via `sse.NewBroadcaster` with `sse.WithBufferSize` (`hub.go:54`); wires
`OnSubscribe`/`OnUnsubscribe` callbacks (`hub.go:60-66`); delegates `Broadcast`/`BroadcastMany`
(`hub.go:99,105`), adds `BroadcastPatch` for DataStar-style patches without importing
go-datastar (`hub.go:119`), delegates graceful `Shutdown` (drain) vs hard `Close`
(`hub.go:127-129,134`), exposes `Health() sse.BroadcasterHealth` and `SubscriberCount()`
(`hub.go:140,145`).

**Canonical endpoint (realtime/handler.go)** — `sse.NewStream` for headers+flusher setup
(`handler.go:101`), starts `stream.Heartbeat(ctx, cfg.heartbeat)` with a `WithHeartbeat`
mount option (default 15s, `handler.go:13-17,38-40,132`), resumes via
`stream.LastEventID()` (`handler.go:203`) and replays through `sse.ReplayFiltered` with
predicate push-down when the store supports `FilteredEventStore` (`handler.go:52,253`).

**Tests** — `ssetest` harness (`MustReadNEvents`, `RequireEventID`, `RequireEventType`,
`Collect`-style helpers) across realtime and integration tests; sseparse re-exports ride
ssetest's compat layer. `sse.ContentType`/`sse.ParseEventID` test-only.

**What is deliberately NOT used** — `WithOnDrop` (drop counters are TODO W5-C1, with the
go-aichat `Drops()` shape named as reference), `OnDisconnect` (no disconnect metrics),
per-request JSON senders (`SendJSON`/`SendData`/`SendKeyed` — realtime's lane is event
streaming, not RPC-over-SSE), and raw `WriteKeyedLines`/`JoinLines` (payload shaping is
consumer-side via `PatchLike`).

## 2. Capability inventory vs adoption

| # | go-sse capability | Status | Evidence |
|---|---|---|---|
| 1 | `Broadcaster` lifecycle (New/Broadcast/BroadcastMany/Subscribe/Unsubscribe/SubscriberCount) | 🟢 Full | `hub.go:54,99,105,145` |
| 2 | `BroadcasterHealth` + `Health()` | 🟢 Full | `hub.go:140`; pinned by integration SSE E2Es |
| 3 | `Shutdown` graceful drain vs `Close` | 🟢 Full | `hub.go:127-134`; drain-before-server-stop contract documented |
| 4 | Option: `WithBufferSize` | 🟢 Full | `hub.go:53-55` |
| 5 | Options: `OnSubscribe`/`OnUnsubscribe` setters | 🟢 Full | `hub.go:60-66` (exposed as `WithOnSubscribe`/`WithOnUnsubscribe`) |
| 6 | Option: `WithOnDrop` | 🔴 Missed | TODO W5-C1 (drop/backpressure counters); known, demand-gated |
| 7 | `Stream` (NewStream/Send) | 🟢 Full | `handler.go:101` |
| 8 | `Stream.Heartbeat` | 🟢 Full | `handler.go:132`; `WithHeartbeat` default 15s |
| 9 | `LastEventID` / resume path | 🟢 Full | `handler.go:203` — uses the library's validated accessor, no hand-parsing |
| 10 | `Replay` / `ReplayFiltered` (push-down) | 🟢 Full | `handler.go:52,253` incl. in-memory post-filter fallback |
| 11 | `EventStore`/`FilteredEventStore` contract | 🟢 Full | interface consumed; no production store shipped (consumer-provided, demand-gated — correct SDK posture) |
| 12 | `OnDisconnect` | 🔴 Missed (small) | no disconnect hook wired; candidate alongside W5-C1 metrics |
| 13 | `Event`/`EventID`/`ParseEventID` | 🟢 Full | throughout hub/handler/tests |
| 14 | `WriteEvent`/`WriteHeartbeat`/`WriteRetry` raw writers | ⚪ N/A | Stream wraps them internally; direct calls would bypass the mutex discipline |
| 15 | `SendJSON`/`SendData`/`SendLines`/`SendKeyed` | ⚪ N/A | realtime's canonical endpoint streams events only; JSON-over-SSE is out of lane |
| 16 | `KeyedLines`/`WriteKeyedLines` (DataStar shaping) | ⚪ N/A | consumers shape payloads via `PatchLike.Event()` |
| 17 | `ContentType`/`EventConnected` constants | 🟡 Partial | test-only (`realtime_test.go:583-584`); runtime headers come from `NewStream` |
| 18 | `ssetest` harness | 🟢 Full | realtime + integration suites (SubscriberCount-sync rule is CI-enforced) |
| 19 | `sseparse` | 🟡 Partial | indirect via ssetest compat re-exports; no direct parse use |

**Score derivation**: A = 16 applicable (14-16 N/A with reasons; 17 and 19 count as
applicable-with-partial). Green = 12 (rows 1,2,3,4,5,7,8,9,10,11,13,18). Partial = 2 × 0.5
= 1.0 (rows 17, 19). Missed = 2 (rows 6, 12). **Score = (12 + 1.0) / 16 = 81/100 (high).**

## 3. Findings (Pareto order)

### F1 — `WithOnDrop` is the last high-value unwired option (already planned)
go-sse fires `OnDrop` per full subscriber buffer; the fleet-coherence note (TODO_LIST) even
names go-aichat ssehub's `Drops()` shape as the reference. When W5-C1 lands, the Hub should
expose drop counters through `WithOnDrop` (respecting the re-entry deadlock rule: the
callback runs under the read lock and must be non-blocking — a counter increment only).

### F2 — Upstream `[Unreleased]` coupling: two changes to absorb at the next go-sse release
1. **Breaking:** `ssetest.WithLastEventID` takes `sse.EventID` (not string) — touches
   realtime/integration tests and the integration PIN (integration pins PUBLISHED tags, so
   the pin re-pin ritual covers it; do not absorb early).
2. **Additive:** `DefaultSubscriberBuffer` becomes exported — realtime's
   `WithBufferSize` doc comment cites "go-sse's `sse.defaultSubscriberBuffer`"
   (`hub.go:80`); update the reference to the exported name on the next train.

### F3 — Version currency is perfect on published tags; 74 unreleased upstream commits
All three in-graph versions equal the newest tags (v0.6.2 / ssetest v0.5.0 / sseparse
v0.2.1). HEAD-ahead content is upstream release-engineering + the F2 items — nothing to
chase now; re-check at the next go-sse release.

### F4 — No production EventStore shipped — correct, keep it that way
Realtime accepts any `sse.EventStore` but ships no implementation; cqrs consumers get
projection-backed stores from go-cqrs-lite. Shipping a default in-memory store would invite
unbounded memory growth in a framework whose stores are consumer-domain (same reasoning as
go-idempotency's no-production-backends rule).

## 4. Should appkit use it MORE? (the explicit answer)

Yes — marginally: `WithOnDrop` (planned), `OnDisconnect` (small), and the two upstream
absorptions at the next release. No — for the send-family and keyed-line helpers: they
serve an interactive SSE-RPC style that realtime deliberately does not own; pulling them in
would blur the wrapper's single responsibility (event fan-out + replay for HTMX/DataStar UIs).

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'sse\.[A-Z]' realtime --glob '!*_test.go' -o | sed 's/.*://' | sort | uniq -c
rg -n 'WithOnDrop|OnDisconnect' realtime/            # empty → F1/F4 gaps stand
rg -n 'defaultSubscriberBuffer' realtime/hub.go      # F2.2 doc-comment reference
git -C /home/lars/projects/go-sse tag --sort=-v:refname | head -1   # v0.6.2
sed -n '1,40p' /home/lars/projects/go-sse/CHANGELOG.md              # [Unreleased] items
```
