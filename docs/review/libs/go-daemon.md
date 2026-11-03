# go-daemon — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). The adoption
> verdict was already researched and recorded:
> `docs/planning/2026-10-01_go-daemon-and-go-aichat-integration.md` — this report narrows it,
> re-verifies it, and records what SHIPPED since.

## Verdict

**NOT USED — and the 2026-10-01 integration research already ruled NO as a dependency, for
recorded reasons that still hold. The one real gap that research surfaced — systemd
`sd_notify` + watchdog — SHIPPED as the opt-in `/systemd` module (v0.1.0, READY=1 via core
`StartHooks`, STOPPING=1 via `DrainHooks`, watchdog at `WatchdogSec/2`), proving the
adopt-the-gap-not-the-dep pipeline works. Everything else is duplicate or reference-only:
go-daemon is a competing host layer (owns its own socket+`http.Server` lifecycle), UDS-first,
deliberately mechanism-only.**

| Fact           | Value                                                                                                                                          | Evidence                                                       |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| Direct imports | 0                                                                                                                                              | `rg 'go-daemon' --glob '*.go'` → empty                         |
| In graph?      | No                                                                                                                                             | absent from all go.mod/go.sum                                  |
| Latest tag     | v0.4.0 (2026-10-03)                                                                                                                            | `git -C ~/projects/go-daemon tag --sort=-v:refname \| head -1` |
| Module path    | **`github.com/LarsArtmann/go-daemon` — capital LA** (case-sensitive module path; the user-facing lowercase form does NOT resolve on the proxy) | `~/projects/go-daemon/go.mod:1`                                |
| Upstream state | HEAD 31 commits past v0.4.0; `[Unreleased]` carries `JSONv2Codec` + injectable negotiation codecs (`WithJSONCodec`/`WithCBORCodec`)            | `~/projects/go-daemon/CHANGELOG.md:8-32`                       |
| Overlap note   | appkit/systemd v0.1.0 implements the same sd_notify/Watchdog pattern go-daemon's `notify.go` ships                                             | `systemd/notify.go` vs `~/projects/go-daemon/notify.go`        |

## 1. Capability inventory

UDS-daemon building blocks (`~/projects/go-daemon/`, single package):

- **Server lifecycle**: stale-socket cleanup, health-probe liveness, 0600 socket mode,
  graceful drain with hook ordering (`Lifecycle`: `Startup`/`OnListen`/`OnShutdown`),
  sd_notify READY/STOPPING/WATCHDOG integrated, optional external listener (v0.4.0).
- **Content negotiation**: `ResolveRequestCodec`/`ResolveResponseCodec` (Content-Type +
  Accept q-value ranking) over JSON/CBOR — built on go-codec.
- **Transports**: pooled unix-socket HTTP clients (timeout / no-timeout for SSE).
- **Extras**: WHATWG-conformant `ParseSSEData`, flight-recorder integration
  (`MaybeCaptureOnSlowRequest`, trace handler), `JSONv2Codec` (unreleased).
- Explicit non-goals: no supervision (systemd owns it), no PID files, no TLS termination.

## 2. Capability vs applicability (re-verifying the 2026-10-01 table)

| Capability                              | Status                                          | Reason                                                                                                                                                                                                                                         |
| --------------------------------------- | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Server/socket lifecycle as a dependency | ⚪ N/A — ruled NO (2026-10-01, holds)           | Competing host layer: importing it inverts layering — appkit core owns the HTTP lifecycle; go-daemon would own a second one. UDS-first framing also doesn't match appkit's TCP/default posture                                                 |
| sd_notify / watchdog                    | 🟢 ADOPTED — as the /systemd module             | The research's "one real gap" shipped v0.1.0 (core `StartHooks` seam was the prerequisite; shipped since). Same pattern (`WatchdogSec/2`), appkit-owned implementation over core hooks — go-daemon proved the design, appkit did not import it |
| Content negotiation                     | ⚪ N/A — adopt the PATTERN, not the dep (holds) | Pre-agreed home is battery B5 (W3 httpx) built on go-codec; go-daemon's `[Unreleased]` injectable-codec work makes it an even better reference, still not a dependency                                                                         |
| `ParseSSEData`                          | ⚪ N/A                                          | Duplicates appkit/realtime (go-sse owns SSE framing+replay)                                                                                                                                                                                    |
| Flight-recorder integration             | ⚪ N/A                                          | Duplicates appkit/flightrecorder + appkit/otel bridges                                                                                                                                                                                         |
| Socket clients                          | ⚪ N/A                                          | No UDS consumer story in appkit                                                                                                                                                                                                                |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **The ruling is fresh, evidence-based, and unexpired** (2026-10-01, four months old): the
   layering inversion argument is structural, not circumstantial.
2. **The valuable slice already shipped** through the correct mechanism (opt-in module over
   core hooks). Re-adopting via dependency now would create two sd_notify implementations in
   the fleet.
3. **When it flips:** only if a consumer needs go-daemon's specific socket-server lifecycle
   (stale-socket takeover, 0600 UDS serving) INSIDE an appkit service. Even then the likely
   shape is: consumer uses both libraries side by side (go-daemon fronting a UDS endpoint,
   appkit serving TCP), not an import. Watch its `[Unreleased]` negotiation work as B5
   reference material, nothing more.

## 4. Recommendations

1. Keep the 2026-10-01 verdict — re-verified row by row above; no TODO changes needed
   (the systemd row is DONE, the convergence watch lives in TODO_LIST:48).
2. If B5 (content negotiation) enters planning, read go-daemon's
   `negotiation.go` + `[Unreleased]` codec-injection API as the reference implementation
   and build on go-codec per the recorded decision.
3. Path-spelling caution for any future doc/tooling: the module path has capital
   `LarsArtmann` — case-insensitive tooling that rewrites it to lowercase will break proxy
   resolution. (Their repo's quirk, worth knowing before ever scripting against it.)

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'go-daemon' --glob '*.go' --glob 'go*.sum'    # empty → absent
sed -n '8,16p' docs/planning/2026-10-01_go-daemon-and-go-aichat-integration.md   # the ruling table
ls systemd/                                          # the adopted slice (notify.go)
head -1 /home/lars/projects/go-daemon/go.mod         # capital-LA module path
git -C /home/lars/projects/go-daemon tag --sort=-v:refname | head -1   # v0.4.0
```
