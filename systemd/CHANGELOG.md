# Changelog

## [0.1.1] - 2026-11-07

### Added

- Docs: runnable godoc Examples (`ExampleInstall`, `ExampleNew` in
  `example_test.go`, the flightrecorderhealth convention) — Install's
  append-only contract and New's three independently usable hooks are
  now compile-checked and output-pinned; the examples double as the
  scratch-compile proof for the README quick-start snippet.

### Fixed

- Docs: the README's _Release state_, _Install_, and dependencies
  sections still described the pre-release posture (dev-only filesystem
  `replace`, "no published tag yet") — rewritten to the shipped reality
  (systemd v0.1.0, core v0.10.0 published require, no replace). The
  watchdog section gains the first-ping rationale (first tick one full
  `WatchdogSec/2` after `READY=1`, never immediately — the manager arms
  the deadline at readiness and the half-deadline cadence carries a full
  missed-datagram margin).

### Changed

- Toolchain: floor train go 1.27.2 → 1.27.3 (go.mod directive; repo-wide
  stdlib-vuln clearance train).
- Tests: `TestNotify_SendFailureIsClassified` now asserts through
  `errorfamilytest` (AssertCode + AssertFamily) instead of a manual code
  comparison — the module's classified-error contract is pinned by the
  family's own test kit (error-family adoption batch).

## [0.1.0] - 2026-10-07

### Added

- Initial implementation of the opt-in systemd (sd_notify) module — the
  one durable value-add from the 2026-10-01 go-daemon/go-aichat research
  (`doc/planning/2026-10-01_go-daemon-and-go-aichat-integration.md`),
  gated in that research on the new core post-listen seam, which ships in
  the same train as `ServiceConfig.StartHooks`.
- `Install(&cfg, opts...)` — the one-liner: appends `READY=1`
  (post-listen, pre-serve), `STOPPING=1` (drain start), and the watchdog
  keepalive + its stop to the config's `StartHooks`/`DrainHooks`/
  `ShutdownHooks`. Append-only, so consumer hooks compose around it.
- `New(opts...)` returning the three hooks as plain `appkit.Hook` values
  for manual wiring; each independently usable. `WithLogger` option.
- Watchdog keepalive at `WatchdogSec/2` (via `coreos/go-systemd/v22`
  `SdWatchdogEnabled`), pinging through the whole drain and shutdown to
  the final phase — whether the manager enforces the timeout after
  `STOPPING=1` is version-dependent, so pinging is safe either way and
  stopping early is not (deliberate deviation from the research doc's
  parenthetical, following go-daemon's proven `socket.go` behavior).
- Error taxonomy: `systemd.notify_failed` / `systemd.watchdog_check_failed`
  (both Infrastructure, go-error-family).
- Outside systemd (`NOTIFY_SOCKET` unset) every hook is a no-op: the same
  binary runs unchanged in development.
- Tests exercise the real unixgram transport against a bound
  `$NOTIFY_SOCKET` stand-in: `READY=1` as the first datagram, `WATCHDOG=1`
  while serving, `STOPPING=1` during shutdown, silence after the final
  phase — plus the full lifecycle through a live `appkit.Service`.
- `example/main.go` — Type=notify demo service with the unit-file snippet.
- `Counters()` — snapshot of the process-global sd_notify activity
  counters (`ReadySent`/`StoppingSent`/`WatchdogPings`/`NotifyFailures`;
  only successful sends count) for scrape-time metrics export — a frozen
  watchdog-ping counter under an armed `WatchdogSec` is the pre-expiry
  hang signal. Surface mirrors go-daemon's `ReadNotifyCounters`, where
  bank-sync ADR-017 (2026-10-04) proved the demand.

### Release gating

- Requires the first core tag carrying `StartHooks` — v0.8.0, tagged the
  same day. The dev-only `replace github.com/larsartmann/go-appkit => ../`
  was lifted and the require bumped to the published tag before this tag
  (`scripts/pre-tag-checks.sh` rejects filesystem replaces); integration
  pin + `documentedPins` entry + CI proxy-smoke slot ride the same train.
