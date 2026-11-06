# Consumer heads-up — cqrs-htmx: core v0.10.0 ships `Server-Timing`; `appkit/preset` v0.1.0 announced

> **Refreshed 2026-11-06 (drafted 2026-11-03) · target: cqrs-htmx (your repo) · the core train
> is now TAGGED (`v0.10.0`, pushed 2026-11-06); preset `preset/v0.1.0` tags with it · voice:
> github-voice, AI-drafted (Crush/GLM) — review before posting.**

## Proposed title

`Heads-up: go-appkit core v0.10.0 emits Server-Timing on every response; new one-call preset module`

## Body draft

Two go-appkit releases landed that touch your integration surface:

**core `v0.10.0`** (tagged, on the proxy):

- Every response from an appkit-hosted service gains one header:
  `Server-Timing: total;dur=<ms>` — W3C Server-Timing, outermost default-stack
  entry (`ServerTiming → Recovery → RequestID → Logging → [Timeout] →
  SecurityHeaders`), built on `httputil/server_timing` v1.0.3. DevTools/APMs
  read it; nothing else on the wire moves.
- `setup`'s `RunWithAppkit` path is functionally unaffected. If any of your
  tests assert exact header sets, they will see the new header after the bump.
- SSE responses are safe: the header stamps at response commit, streams keep
  flowing.
- Opt-outs: `appkit.ServiceConfig.ServerTiming: &false` (pointer knob,
  `RegisterHealth` semantics), or replace `Config.Middlewares`.
- Also in the train: `appkit.RegisterClassifications()` (opt-in fleet error
  classifications; deliberately not auto-called) and testkit helpers
  (`FreeAddr`, `WaitForRunning`, `HTTPGetURL`/`HTTPGetContext`).

**`appkit/preset` `v0.1.0`** (new module): one-call composition of the family
for apps — `preset.New(Config)` boots the always-on spine (core stack,
sd_notify, health dashboard, error pages, version stamp) plus env-driven
otel/realtime/security legs (`APPKIT_OTEL|REALTIME|ERRORPAGES|SECURITY`,
explicit code wins). Example app: 35 LOC to full production posture. Not a
cqrs-htmx dependency question — your setup stays the composition root for
CQRS apps; preset is the story for the plain-Service apps around them.

Pinned evidence: go-appkit `v0.10.0` tag, `middleware.go` (stack contract),
`preset/README.md` (grammar table). Auto-recorded `total` only — sub-phases
stay handler-side (`servertiming.MeasureServerTiming` before first write).

💘 Generated with Crush
