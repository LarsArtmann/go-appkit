# Integration Research — go-daemon and go-aichat → go-appkit

> **Date:** 2026-10-01. **Type:** Point-in-time research (re-verify before acting).
> **Question:** Do either [`/home/lars/projects/go-daemon`](../../../go-daemon) (unix-socket daemon building blocks) or [`/home/lars/projects/go-aichat`](../../../go-aichat) (AI-chat building blocks) offer go-appkit a real value-add to adopt?
>
> **Verdicts:**
>
> | Candidate / option                                                      | Verdict                                   | One-line reason                                                                                                                                                                    |
> | ----------------------------------------------------------------------- | ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
> | go-daemon as a go-appkit dependency or module                           | 🔴 **NO**                                 | It is a competing host layer (owns its own `http.Server` lifecycle) that is UDS-only and deliberately mechanism-only; importing it inverts layering for no framework value         |
> | **systemd integration (sd_notify READY/STOPPING + watchdog) into core** | 🟢 **YES — the one real gap**             | appkit has zero systemd/`sd_notify` code; go-daemon proves the exact pattern. Needs a new post-listen core hook (additions-only, v1-shaped) before an opt-in module can use it     |
> | Unix-domain-socket serving in core                                      | 🟡 **WORTH CONSIDERING — demand-gated**   | `Start()` hardcodes `"tcp"`; UDS is a small config addition, but it is outside appkit's TCP-HTTP charter                                                                           |
> | Content negotiation (JSON/CBOR) from go-daemon                          | 🟡 **ADOPT THE PATTERN, NOT THE DEP**     | Maps to battery W3-B5 (`httpx`) and should be built there on go-codec — go-daemon's own roadmap pre-agrees the same extraction; it is a reference implementation, not a dependency |
> | `ParseSSEData`, go-daemon `FlightRecorder`, socket clients              | 🔴 **NO**                                 | Duplicates appkit/realtime (go-sse) and appkit/flightrecorder + appkit/otel                                                                                                        |
> | go-aichat code as an appkit module                                      | 🔴 **NO**                                 | Domain-specific (OpenAI-compatible chat): `client`, `contextwindow`, `ssehub` are application-shaped                                                                               |
> | go-aichat/ssehub as a replacement for appkit/realtime                   | 🔴 **NO**                                 | Partial overlap (both wrap go-sse); realtime adds replay-with-dedup + handler/Mount. Surface the fleet duplication, do not chase it                                                |
> | Reverse adoption: go-aichat consumers host on go-appkit                 | 🟢 **RECOMMEND** (their repo, user-gated) | Exact cqrs-htmx ADR-001 / PapDashboard replay: appkit is the host HTTP layer; go-aichat stays the AI layer                                                                         |
> | `integration/` E2E tests against either repo                            | 🟡 **NOT NOW — trigger-gated**            | No shared contract exists to pin until reverse adoption lands                                                                                                                      |
>
> **Bottom line:** no go-appkit dependency on either repo. Exactly **one** durable, demand-ready appkit-side gap surfaced: **systemd `sd_notify` + watchdog**, gated on a new post-listen startup hook in core. Everything else in go-daemon is duplicate or reference-only; everything in go-aichat is either domain-shaped or a fleet-coherence observation (two parallel SSE hubs exist and should converge deliberately, not by accident).

---

## 1. Method and evidence base

Deep-dive of both candidate repos (README, AGENTS, FEATURES, ROADMAP, TODO_LIST, `go.mod`, and the relevant source) cross-checked against appkit's `service.go`, module inventory, TODO_LIST, ROADMAP, and the battery spec.

| Verification                                                                              | Result                                                                                                                                                  |
| ----------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `grep -rn "sd_notify\|SdNotify\|systemd\|Watchdog\|unix socket" go-appkit --include=*.go` | **Zero matches** — core has no systemd or UDS support (the single confirmed gap)                                                                        |
| `grep -rn "go-daemon\|go-aichat" go-appkit` and the reverse                               | **Zero cross-imports** — no coupling in either direction today                                                                                          |
| `go-appkit/service.go:106`                                                                | Listener is **hardcoded `"tcp"`**: `Listen(ctx, "tcp", s.cfg.Addr)` — UDS would need a config change here                                               |
| `go-appkit/service.go:134-154` (`Run`)                                                    | **No post-listen/post-start hook** — `Start()` binds, then selects on ctx/errCh; only `DrainHooks` + `ShutdownHooks` exist (both shutdown-time)         |
| `go-daemon/socket.go:255-270, 327-351`                                                    | Reference impl of the pattern: `SdNotifyReady` after listen, watchdog goroutine at `WatchdogSec/2`, `SdNotifyStopping` before drain                     |
| `go-daemon/go.mod`                                                                        | Single module, `github.com/LarsArtmann/go-daemon`, depends on `coreos/go-systemd/v22 v22.7.0`, go-codec, go-flightrecorder, go-sse/sseparse, prometheus |
| `go-aichat/go.mod`                                                                        | Single module, `github.com/larsartmann/go-aichat`, depends on failsafe-go, go-datastar, go-sse, `sashabaranov/go-openai`, Ginkgo/Gomega; **no appkit**  |
| appkit `FEATURES.md:108`                                                                  | appkit/realtime's own documented limitation is the 64-buffer SSE drop — the same problem class ssehub instruments with drop counters                    |

## 2. Candidate profile — go-daemon

**What it is.** Production-extracted _mechanism_ library for long-lived daemons served over **unix domain sockets** (from `project-discovery-daemon`, its primary consumer). Deliberately policy-free.

| Piece                                                                      | What it does                                                                                                          | appkit relationship                                                                   |
| -------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| `SocketServer`                                                             | Full UDS lifecycle: stale-socket cleanup, `0600` perms, **sd_notify READY/STOPPING + watchdog**, drain, hook contract | **Competing host layer** — appkit owns TCP lifecycle; the _systemd part_ is the value |
| `NewHTTPClient` / `NewHTTPClientNoTimeout`                                 | Unix-socket HTTP clients (pooled keep-alive)                                                                          | No appkit equivalent; niche, no demand                                                |
| `Resolve{Request,Response}Codec`, `WriteEncoded`/`WriteJSON`/`WriteErrorf` | JSON/CBOR content negotiation (RFC 6839 suffixes, q-ranking, `Vary: Accept`); pooled writers                          | **Reference impl** for battery W3-B5; routing is via go-codec, not go-daemon          |
| `ParseSSEData`                                                             | WHATWG SSE framing, pinned to the WPT/Chromium corpus                                                                 | Duplicates go-sse (appkit/realtime's foundation)                                      |
| `FlightRecorder` + `TraceHandler`                                          | Execution trace + Prometheus metrics + rate-limited `go tool trace` streaming                                         | Overlaps appkit/flightrecorder + appkit/otel `frmetrics`                              |
| `IsSocketAlive`                                                            | Socket liveness probe                                                                                                 | UDS-specific; irrelevant to TCP appkit                                                |

**Key structural fact:** go-daemon and appkit solve the _same_ problem (own an `http.Server` lifecycle with graceful drain) for _different transports_ (UDS vs TCP). They are alternatives, not layers. go-daemon's lifecycle is also _simpler_ than appkit's (no readiness gate, no drain window, no phase logging), so appkit has nothing the daemon lib needs back — and go-daemon's mechanism-only charter argues against it adopting appkit-shaped policy.

## 3. This repo's constraints (unchanged by these candidates)

Consumer owns the composition root; growth pattern is thin opt-in modules; "batteries included" never means a fat core; core additions are v1-shaped and demand-driven (the sentinel/hook registry pattern: `NoTimeout`, `NoDrainDelay`, `DrainHooks`, `ShutdownHooks`); dependency weight is a recorded negative; the battery spec (`doc/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md`) already routes both content negotiation (W3-B5) and an outbound-resilience client (W4 `polite`). Nothing here changes those constraints; both candidates are evaluated _against_ them.

## 4. go-daemon — option analysis

### Option A — go-daemon as an appkit dependency / module

**CONTRA (dispositive).** Wrong layer (a daemon-host library, not a framework primitive) and wrong direction (appkit's value flows _out_ to consumers, never _in_ from sibling host layers). the six pieces are each duplicate or reference-only (table in §2). No adoption.

### Option B — systemd integration in appkit core (the one YES)

**PRO**

- **A real, verified gap.** appkit services run under systemd in production; `Type=notify` + `WatchdogSec` (ready signaling + liveness watchdog + `STOPPING=1` for slow shutdowns) is the standard systemd contract, and appkit has **none** of it (zero grep matches).
- **Small and well-precedented.** go-daemon already implements the exact sequence (`socket.go:255-270, 327-351`) with tests; the appkit version is a straight port of the pattern.
- **Fits appkit's lifecycle cleanly once one seam exists.** `STOPPING=1` + watchdog-stop belong in `DrainHooks` (already run at drain start, service.go:188). Only `READY=1` + watchdog-start are homeless: they must fire _after the listener binds, before traffic_. `Service.Start()` binds then spawns `Serve`, and `Run` selects — there is **no post-listen hook today**.

**The enabling change (additions-only, v1-shaped):** core gains a startup hook — e.g. `ServiceConfig.OnStart func(context.Context)` (or `AfterListen`) — invoked in `Start()` after the listener binds and before serving, mirroring the existing `DrainHooks`/`ShutdownHooks`/`OuterMiddlewares` addition pattern. With that seam, systemd support can ship either as an opt-in `systemd` satellite module (preferred — matches "capability arrives as opt-in modules") or as a tiny core option. Either way the _hook_ is the prerequisite, and it is independently useful (post-bind announcements, dynamic port discovery).

**Effort:** core hook ~1 short session (hook + contract test + phase/log wiring); `systemd` module ~1 session (port of go-daemon's notify/watchdog helpers + tests, no go-daemon dependency — depend directly on `coreos/go-systemd/v22`, or vendor the ~40 LOC of `sd.SdNotify` calls).

### Option C — unix-domain-socket serving

**PRO** — nearly free once you touch the listener: `service.go:106` hardcodes `"tcp"`; a `Network`/`SocketPath` config plus the `os.Chmod`/stale-cleanup from go-daemon is small, and the appkit drain/shutdown phases would then apply to UDS daemons.

**CONTRA** — outside appkit's stated TCP-HTTP charter; go-daemon already owns this niche; no consumer demand. **Trigger:** a consumer needs a local-socket appkit service (e.g. a sidecar CLI). Then add `Network`/`SocketPath` and reuse `DrainHooks` for stale-socket cleanup.

### Option D — content negotiation (JSON/CBOR)

**PRO** — go-daemon already built a correct, fuzzed, corpus-pinned implementation; battery W3-B5 (`httpx`) is already routed and wants exactly this.

**CONTRA / routing** — do **not** depend on go-daemon for it. go-daemon's own ROADMAP pre-agrees the extraction ("extracting the pure media-type ↔ codec mapping into go-codec when a second transport consumer appears"). The appkit home is `httpx` (W3) on top of `go-codec`, with go-daemon's `negotiation.go` + golden matrix as the reference and conformance source. Record the pointer; adopt at W3 time.

### Option E — SSE parser / FlightRecorder / socket clients

No. Duplicate (go-sse, appkit/flightrecorder + otel) or niche.

## 5. go-aichat — option analysis

**What it is.** A domain library of AI-chat building blocks for OpenAI-compatible providers: `client` (streaming completions, error taxonomy, `Resilient` circuit breaker via failsafe-go, `RepetitionGuard`), `contextwindow` (token budgeting), `ssehub` (per-channel SSE fan-out). It is an _AI layer_, not an HTTP host layer.

### Option A — go-aichat code into appkit, or appkit depending on it

**NO.** Domain-shaped: OpenAI error kinds, token estimation, chat turn orchestration have no place in a general HTTP framework; appkit depending on an AI library inverts layering. Same "wrong layer" verdict as PapDashboard/cordis.

### Option B — ssehub vs appkit/realtime

Partial overlap, both built on go-sse:

| Concern         | appkit/realtime                                    | go-aichat/ssehub                                     |
| --------------- | -------------------------------------------------- | ---------------------------------------------------- |
| Fan-out         | single-stream `Broadcaster`                        | per-channel routing keys (`<channel>#<seq>`)         |
| Replay          | `Last-Event-ID` replay + live dedup in the handler | `Last-Event-ID` replay from a per-channel ring store |
| Drop visibility | known 64-buffer drop (FEATURES:108)                | `Drops()` counter + `WithOnDrop` hook + `Diag()`     |
| HTTP surface    | `Handler` + `Mount` (heartbeat, CORS, filter)      | transport-agnostic (`Subscribe` returns a channel)   |
| Backpressure    | broadcaster default                                | explicit by-loss, counted                            |

**Recommendation: do not adopt or merge.** realtime's design is deliberately simple and integrated with the appkit stack; ssehub's channel model is a different product shape. Record the _fleet-coherence_ observation — two SSE hubs exist; if both keep growing, one should converge on the other (or both on go-sse primitives) by decision, not drift. appkit/realtime's W5 C1 item (drop/backpressure counters) can borrow the `Drops()`/`WithOnDrop`/`Diag` shape from ssehub as a reference.

### Option C — outbound resilience (`client.Resilient`)

appkit has no outbound-client resilience; battery W4 `polite` is the future home, and go-aichat's failsafe-go-based breaker + "never fast-retry warming-up" policy is a useful _reference_ for its tuning. But it is OpenAI-error-coupled; adopt the idea, not the code, at W4 time.

### Option D — reverse adoption: go-aichat consumers host on go-appkit

**The real prize, and the recommended pattern (their repo, user-gated).** go-aichat is a library; whoever serves it over HTTP (KeyHolderAI today) is a candidate appkit host — the exact cqrs-htmx ADR-001 / PapDashboard replay: swap the hand-rolled `http.Server` + signal handling + shutdown for `appkit.Service`, gaining the drain phase (readiness flip + drain window) that protects open SSE streams, and reuse appkit/realtime or ssehub as the transport. `chatservice`'s transport-neutral `Publisher` interface is a natural seam for an appkit/realtime-backed implementation. Zero appkit code changes expected.

### Option E — `integration/` E2E tests

**Not now.** Neither repo depends on appkit today, so there is no shared contract to pin — a test would just exercise the other repo inside appkit's tree. **Trigger:** go-aichat (or a consumer) ships an appkit-hosted release, or go-aichat adds an appkit-hosted example; then add a smoke test (boot service → `chatservice` turn through an appkit/realtime hub → SSE event received), mirroring `TestJournalBackedReplayThroughAppkitService`.

## 6. Decision

**No go-appkit dependency on either repo.** One durable appkit-side gap and one ecosystem observation:

1. **adopt the systemd pattern** (not the dependency): the actionable appkit change is the **post-listen startup hook** in core, which unlocks an opt-in `systemd` (`sd_notify` + watchdog) module. This is the single clean value-add from go-daemon.
2. **watch, do not merge, the two SSE hubs** — a fleet-coherence call belongs to the owners of appkit/realtime and go-aichat/ssehub together.

**Bounded follow-ups:**

1. ~~Create this research doc~~ (done).
2. TODO_LIST P3: systemd integration item, gated on the core post-listen hook (prerequisite) — with go-daemon `socket.go` as the reference and the core hook flagged as the additions-only enabling seam. **EXECUTED 2026-10-01 (same day, owner-directed):** core `ServiceConfig.StartHooks []Hook` + the opt-in `/systemd` module landed race-green; the TODO_LIST item now tracks only the release train (core v0.8.0 → systemd v0.1.0 → integration pin). One deliberate deviation from §Option B: the watchdog pings to the FINAL phase (ShutdownHooks), not stopped in DrainHooks — whether the manager enforces the watchdog after `STOPPING=1` is version-dependent, so pinging is safe either way (go-daemon's own proven behavior).
3. TODO_LIST P3: SSE-hub coherence + go-aichat reverse-adoption watch item (triggers recorded).
4. Optional, user-gated: propose reverse adoption to the go-aichat/KeyHolderAI side (their repo, their gate — not executed here).

## 7. Appendix — re-entry paths

| If this happens                                               | The integration becomes                                                                                                                      |
| ------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| A consumer runs an appkit service under systemd `Type=notify` | Core gains `ServiceConfig.OnStart` (post-listen hook); a `systemd` satellite module ships `sd_notify` READY/STOPPING + watchdog on top of it |
| A consumer needs a local-socket appkit service                | Core gains `Network`/`SocketPath`; UDS + stale-socket cleanup reused from go-daemon's `socket.go` as reference                               |
| Battery W3 (`httpx`) is picked up                             | Content negotiation built on `go-codec` using go-daemon `negotiation.go` + golden matrix as the conformance source (no go-daemon dependency) |
| go-aichat or KeyHolderAI ships an appkit-hosted release       | `integration/` gains an appkit/realtime + `chatservice` smoke test pinning their published tag                                               |
| appkit/realtime and go-aichat/ssehub both keep growing        | A convergence decision (one wins, or both sit on shared go-sse primitives) — owner call, recorded in the watch item                          |

**Update 2026-10-04 — the systemd trigger fired INVERTED.** bank-sync (an overlap consumer: indirect appkit via cqrs-htmx v4.13.0 AND direct go-daemon) hit the `Type=notify` need and adopted **go-daemon v0.4.0** for it (ADR-017: listener-agnostic `Server`, sd_notify + watchdog + scrape-time `ReadNotifyCounters` metrics) — because this repo's train (core v0.8.0 + `systemd` v0.1.0) was still unreleased. Their ADR confirms the §4 Option B analysis line by line: they hand-rolled appkit's exact drain-window contract around go-daemon's missing drain phase (signal → drain → detached-context cancel; their `-race` caught the first-draft race), and keep a curl-polling readiness gate appkit's drain probe would replace. The verdicts above are unchanged (no go-daemon dependency; pattern, not dep), and the demand signal migrated into this repo: `systemd.Counters()` (notify-activity metrics) shipped into the unreleased module, and the release train is now fleet-critical rather than speculative.
