# Consumer heads-up — cqrs-htmx: core default stack now emits `Server-Timing`

> **Drafted 2026-11-03 · target: cqrs-htmx (your repo) · post as issue/comment when the core
> train tags · voice: github-voice, AI-drafted (Crush/GLM) — review before posting.**

## Proposed title

`Heads-up: go-appkit core will emit a Server-Timing header on every response (next core minor)`

## Body draft

`go-appkit` core master now ships a W3C `Server-Timing` middleware as the OUTERMOST entry of the
default stack (`ServerTiming → Recovery → RequestID → Logging → [Timeout] → SecurityHeaders`,
commit `a5d9e55`), built on the family's `httputil/server_timing` v1.0.3. It rides the next core
minor — no tag yet (accumulate-train posture).

## What changes for you

- Every response from an appkit-hosted service gains one response header:
  `Server-Timing: total;dur=<ms>`. DevTools/APMs read it; nothing else on the wire moves.
- `setup`'s `RunWithAppkit` path is unaffected functionally. If any of your tests assert exact
  header sets, they will see the new header after the bump.
- SSE responses are safe: the header stamps at response commit, streams keep flowing.

## Opt-outs (both documented in core)

- `appkit.ServiceConfig.ServerTiming: &false` (pointer knob, `RegisterHealth` semantics).
- Replacing `Config.Middlewares` removes it like every other default.

Pinned evidence: go-appkit `a5d9e55`, `middleware.go` (stack contract),
`docs/DOMAIN_LANGUAGE.md` (vocabulary row). Auto-recorded `total` only — sub-phases are
handler-side (`servertiming.MeasureServerTiming` before first write).

💘 Generated with Crush
