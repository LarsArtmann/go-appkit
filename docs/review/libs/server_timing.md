# server_timing — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). Supersedes and
> reaffirms the `httputil/server_timing` section of
> `docs/research/2026-10-11_sibling-library-utilization-audit.html` (newer and narrower).

> **⚠ OWNER RULING 2026-11-03 (post-audit errata — supersedes the "should NOT use now" framing below).**
> The owner clarified the repo's mission: go-appkit exists to make his apps smartly compose the
> family's libraries. Under that mission, server_timing is a PLANNED ADOPTION whose sequencing is
> owner-gated — not a decline. The factual content of this report (zero imports, v1.0.3 in-graph,
> capability surface, F2 routing) is unchanged and verified; the verdict line "should we SHOULD
> NOT use it now" is corrected to: **adopt — design the wiring (default-stack
> `ServerTimingMiddleware()` vs opt-in slot) as the next HTTP-layer change; keep the CV-port
> refusal.** The shared-duration-source sequencing with the Logging middleware (battery F2's
> upstream ask) remains the one design question to settle at implementation time.

## Verdict

**NOT USED — deliberately. Zero imports in any module; the module exists in go.sum only
(v1.0.3) because httputil requires it for its own tests, and Go cannot mark a dependency
test-only. Nothing has changed in appkit's posture since the 2026-10-11 audit; the upstream
version moved v1.0.2 → v1.0.3 (coordinated with httputil v1.6.0). The adoption path is
already recorded: battery F2's timing-header half was re-scoped to wire this library
instead of porting CV's 116-line `X-Response-Time` handler.**

**Note on naming:** there is no standalone `/home/lars/projects/server_timing` checkout —
the library lives as a submodule of the httputil repo
(`~/projects/httputil/server_timing/`, module path
`github.com/larsartmann/httputil/server_timing`, tagged `server_timing/v1.0.3`).

| Fact | Value | Evidence |
|---|---|---|
| Direct imports | 0 of 12 modules | `rg 'server_timing' --glob '*.go'` → empty |
| In graph? | go.sum only (httputil test dep chain) | `go.sum:33-34` + 4 satellite go.sums |
| Latest tag | server_timing/v1.0.3 = in-graph version | `git -C ~/projects/httputil tag \| grep server_timing` |
| Adoption path | battery F2 re-scope (2026-10-11 audit); W3 httpx context | TODO_LIST upstream-ask annotation |

## 1. Capability inventory (why it is worth wiring when the time comes)

A complete W3C Server-Timing implementation in a zero-dependency module
(`~/projects/httputil/server_timing/`):

- `ServerTimingMiddleware()` / `ServerTimingMiddlewareWhen(pred)` — header emission.
- **`WrapServerTiming(w, r)`** — late header injection on a writer wrapper, preserving
  Flusher/Hijacker/Pusher and propagating `r.Pattern`. This is the shape built for
  framework-owned chains — exactly appkit's situation, where the default middleware stack
  is assembled in `middleware.go` and consumers can replace it wholesale.
- Nil-safe context helpers (`MeasureServerTiming`, `RecordServerTiming`,
  `ServerTimingFromContext`) — no-ops without the middleware, so handlers need no branching.
- Auto-recorded `total` at first write; metric-name sanitization.
- Gotcha: measure before the first write or deferred metrics miss the header
  (`server_timing/README.md:68`).

## 2. Capability vs applicability

| Capability | Status | Reason |
|---|---|---|
| Middleware + `WrapServerTiming` | ⚪ N/A today / adoption path recorded | No timing-exposure story exists in the default stack or batteries yet; wiring is routed through battery F2/W3, demand-gated per house rules |
| Context helpers | ⚪ N/A today | Same gate |
| Version currency | 🟢 | v1.0.3 in every go.sum = newest tag |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **No consumer or battery asks for timing exposure yet.** The default stack is
   minimal by charter (5 middlewares); adding Server-Timing is a wire-format change every
   consumer would see.
2. **The plan already routes adoption** — the 2026-10-11 audit killed the redundant CV-port
   sketch (battery F2's header half) and re-scoped it to "wire the sibling". Re-affirmed
   here: the port must not be written.
3. **When it flips:** any of — a consumer asks for DevTools-visible timing, the W3 httpx
   module lands (it would host the wiring), or an APM story wants server-side duration
   buckets alongside otel traces. The integration is additive: one middleware slot (or
   `WrapServerTiming` inside the core response path) with zero new dependencies.

## 4. Recommendations

1. Keep the TODO_LIST F2 annotation (adopt-wire, not port) — re-verified as current.
2. When W3 httpx is designed, make `ServerTimingMiddleware()` the default answer for the
   timing-header half; keep the upstream-ask for shared-duration-source sequencing with the
   Logging middleware emit.
3. Re-check version currency at the next httputil release train (v1.0.x moves with it).

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'server_timing' --glob '*.go'                        # empty → zero imports
rg -n 'server_timing' --glob 'go*.sum' | head -4           # v1.0.3 via httputil test chain
git -C /home/lars/projects/httputil tag --sort=-v:refname | grep server_timing | head -1
ls /home/lars/projects/httputil/server_timing/             # the library's real home
```
