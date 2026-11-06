# Changelog

All notable changes to this module are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

## [Unreleased]
