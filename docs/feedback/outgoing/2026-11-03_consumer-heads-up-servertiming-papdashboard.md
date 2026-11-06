# Consumer heads-up — PapDashboard: core v0.10.0 ships `Server-Timing`; preset is your upgrade path

> **Refreshed 2026-11-06 (drafted 2026-11-03) · target: PapDashboard (your repo) · the core
> train is now TAGGED (`v0.10.0`, pushed 2026-11-06); preset `preset/v0.1.0` tags with it ·
> voice: github-voice, AI-drafted (Crush/GLM) — review before posting.**

## Proposed title

`Heads-up: go-appkit core v0.10.0 emits Server-Timing; new preset module composes the family in one call`

## Body draft

Two go-appkit releases relevant to your core pin (`v0.4.0` today):

**core `v0.10.0`** — plain `Service` usage like yours gets the following on
bump:

- One additive response header on every response: `Server-Timing:
  total;dur=<ms>` (W3C, outermost default-stack entry). Same statuses, same
  bodies; browser DevTools and any APM can read it.
- Your notification-hub SSE endpoints are safe — the header stamps at
  response commit, streams unaffected.
- Off switches: `appkit.ServiceConfig.ServerTiming: &false`, or replace
  `Config.Middlewares`.

**`appkit/preset` `v0.1.0`** — the upgrade path for the hosting boilerplate
you hand-rolled in v0.3.0: `preset.New(preset.Config{...})` boots the whole
spine (core stack + graceful drain, systemd sd_notify, health dashboard with
`/healthz /readyz /startupz`, pretty error pages, version stamp) and
env-driven legs for otel tracing, SSE realtime, and security profiles —
`APPKIT_SECURITY=standard` alone adds HSTS/nosniff/body-limit. Example app:
35 LOC to full production posture. Your repo keeps ownership of the swap
(per the integration memo, `docs/planning/2026-09-04_papdashboard-integration.md` §7).

Pinned evidence: go-appkit `v0.10.0` tag, `middleware.go`, `preset/README.md`
(env grammar table), `preset/example/main.go` (the 35-LOC demo).

💘 Generated with Crush
