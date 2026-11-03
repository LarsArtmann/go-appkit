# Consumer heads-up — PapDashboard: core default stack now emits `Server-Timing`

> **Drafted 2026-11-03 · target: PapDashboard (your repo) · post as issue/comment when the core
> train tags · voice: github-voice, AI-drafted (Crush/GLM) — review before posting.**

## Proposed title

`Heads-up: go-appkit core will emit a Server-Timing header on every response (next core minor)`

## Body draft

`go-appkit` core master adds a W3C `Server-Timing` middleware as the outermost default-stack
entry (commit `a5d9e55`, family module `httputil/server_timing` v1.0.3). Plain `Service` usage
like yours gets it automatically once you bump core — every response gains exactly one header:
`Server-Timing: total;dur=<ms>`.

## What changes for you

- Nothing functional: same statuses, same bodies, one additive response header your browser
  DevTools and any APM can read.
- Your notification-hub SSE endpoints are safe — the header stamps at response commit, streams
  unaffected.
- If you ever want it off: `appkit.ServiceConfig.ServerTiming: &false` (or replace
  `Config.Middlewares`).

No action needed now; core has not tagged this yet (accumulate-train). Pinned evidence:
go-appkit `a5d9e55`, `middleware.go`, `README.md` config table (`ServerTiming` row).

💘 Generated with Crush
