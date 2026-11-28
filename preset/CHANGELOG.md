# Changelog

All notable changes to this module are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-11-13

### Fixed — realtime hub drains in DrainHooks, not ShutdownHooks (graceful-shutdown deadlock with open SSE streams)

- `preset.New` wired the realtime hub's graceful `Shutdown` into
  `ShutdownHooks`, which core runs AFTER `server.Shutdown` has released the
  connections. With an open SSE stream the bounded server wait (default
  `ShutdownTimeout` 15s) always expired first — `graceful shutdown complete
  result=error`, `server.shutdown_failed: context deadline exceeded` — and the
  hub's own contract ("call this BEFORE shutting down the HTTP server so
  browsers auto-reconnect to another instance") was never met. The
  pre-swap core (≤ v0.11.0) masked the wrong-bucket placement with a
  deadline-less server wait that hung until the client's own connection died.
- The hub drain now joins `DrainHooks` (streams terminate while the server
  still serves; `server.Shutdown` then sees idle connections), and the otel
  provider flush stays in `ShutdownHooks` (spans still cover the final
  in-flight requests). Pinned by `TestRealtime_StreamsEvents` (graceful
  shutdown WITH an open stream): shutdown now completes in ~2ms instead of
  burning the 15s deadline. Error classification unchanged
  (`preset.realtime_shutdown_failed`).
- Surfaced by the 2026-11-07 core v0.12.0 composition swap once the 2026-11-08
  daemon dependency bump (e1a1b55) moved the module onto it; the bump's CI run
  was never verified (the session-close ritual gap), so master's `test
  (./preset)` job stayed red from 2026-11-08 to 2026-11-13.

## [0.1.0] - 2026-11-06

### Added — initial module (SUPERB v12 P-train, the composition preset)

- `New(Config)` one-call composition: always-on spine — core `appkit.Service`
  (Server-Timing default stack, RegisterClassifications), `systemd.Install`
  sd_notify, health probe + real-time dashboard (owns the `/health` surface:
  `/health`, `/healthz`, `/readyz`, `/startupz`; core's own health family
  mounts only under `Config.NoDashboard`), pretty error pages, and the
  version stamp (`GET /version`) when `Config.Version` is set.
- Env-driven legs with explicit-code-wins precedence (nil leg field = the env
  switch decides; non-nil leg decides itself): `APPKIT_OTEL` (provider Setup +
  middleware + shutdown flush), `APPKIT_REALTIME` (SSE hub at `/events`,
  configurable pattern, hub exposed via `RealtimeConfig.Hub` out-field),
  `APPKIT_ERRORPAGES` (spine member, default on), `APPKIT_SECURITY`
  (`none`/`standard`/`strict`; strict requires `SecurityConfig.Origins` and
  rejects at boot without it). Invalid values fail fast at `New` with a
  Rejection naming the variable.
- Lifecycle hook passthrough, consumer-first: `Config.StartHooks` /
  `DrainHooks` / `ShutdownHooks` run before the preset's spine closures in
  every phase (health Mounted start/drain, realtime hub drain, otel provider
  flush) — consumer shutdown work stays covered by traces.
- `Config.Metrics` passes through to core verbatim; the preset never invents
  credentials (nil = no metrics surface).
- Test suite pins the whole contract (spine, hooks, env grammar, leg
  matrix, metrics passthrough, graceful SSE shutdown); example app = 35 LOC.
