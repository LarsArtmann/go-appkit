# systemd — sd_notify for go-appkit services

Opt-in [systemd](https://www.freedesktop.org/software/systemd/man/latest/sd_notify.html)
service-manager integration: `READY=1` once the listener binds, `STOPPING=1`
at the start of the graceful drain, and the `WATCHDOG=1` keepalive while the
service runs. One import, one `Install` call — the hooks ride appkit's
lifecycle seams (`StartHooks`, `DrainHooks`, `ShutdownHooks`) so the
notifications fire at exactly the right points.

**Outside systemd every hook is a no-op** (`NOTIFY_SOCKET` unset): the same
binary runs unchanged in development and under the service manager.

## Install

```bash
go get github.com/larsartmann/go-appkit/systemd
```

Requires a go-appkit core tag carrying `ServiceConfig.StartHooks`
(v0.8.0+; until that tag exists the module's go.mod carries a dev-only
filesystem `replace` onto the repo root — see _Release state_ below).

## Quick start

```go
cfg := appkit.DefaultServiceConfig()
systemd.Install(&cfg)

svc, err := appkit.NewService(cfg)
if err != nil {
    panic(err)
}
// register routes, then:
err = svc.Run(context.Background())
```

with a unit file like:

```ini
[Service]
Type=notify
WatchdogSec=30
# standard hardening omitted
```

`Type=notify` makes the manager wait for `READY=1` before considering the
service up — which appkit sends only after the socket is bound, exactly the
protocol's intent. `WatchdogSec` arms the liveness watchdog, which the
module feeds at `WatchdogSec/2`.

## Ordering contract

| Hook       | Appkit seam     | Sends                        | Failure behavior                                                    |
| ---------- | --------------- | ---------------------------- | ------------------------------------------------------------------- |
| `Start`    | `StartHooks`    | `READY=1`, starts watchdog   | Send failure returns an error → the whole start fails (fail-closed) |
| `Drain`    | `DrainHooks`    | `STOPPING=1`                 | Send failure returns an error, joined into the shutdown result      |
| `Shutdown` | `ShutdownHooks` | stops the watchdog goroutine | Cannot fail; idempotent                                             |

The watchdog deliberately pings **to the end of the final phase**, through
the whole drain and shutdown: whether the manager still enforces the
timeout after `STOPPING=1` is version-dependent — pinging is safe either
way, stopping early is not.

`Install` appends to the config's hook slices (never replaces), so consumer
hooks compose before and after. `New(opts...)` returns the three hooks as
plain values for manual wiring; each is independently usable.

## Options

| Option       | Default          | Effect                                                       |
| ------------ | ---------------- | ------------------------------------------------------------ |
| `WithLogger` | `slog.Default()` | Logger for notify diagnostics (Debug=success, Warn=degraded) |

## Observability

`Counters()` snapshots the process-global sd_notify activity counters —
`ReadySent`, `StoppingSent`, `WatchdogPings`, `NotifyFailures`. Only
successful sends count: outside systemd everything stays zero, so a
nonzero counter against a non-notify unit is itself a misconfiguration
signal. Read them at scrape time to export sd_notify health as metrics —
a frozen `WatchdogPings` counter under an armed `WatchdogSec` is the
pre-expiry hang signal.

## Error taxonomy

Errors use [go-error-family](https://github.com/LarsArtmann/go-error-family)
classification:

| Code                            | Classification | Meaning                                 |
| ------------------------------- | -------------- | --------------------------------------- |
| `systemd.notify_failed`         | Infrastructure | sd_notify datagram send failed          |
| `systemd.watchdog_check_failed` | Infrastructure | `WATCHDOG_USEC` present but unparseable |

## Testing

Tests exercise the real transport: a bound unixgram socket stands in for
the manager's `$NOTIFY_SOCKET`, and datagrams are asserted byte-for-byte
(`READY=1` first, `WATCHDOG=1` while serving, `STOPPING=1` during
shutdown, silence after the final phase). Env-touching tests are
serialized via a package mutex — the environment is process-global.

## Release state

**Unreleased.** The module requires the core `StartHooks` seam, which has
no published tag yet, so `go.mod` carries a dev-only
`replace github.com/larsartmann/go-appkit => ../`. Release train order:
tag core first, then remove the replace, bump the require to that tag,
and tag `systemd/v0.1.0` (`scripts/pre-tag-checks.sh` rejects any tag
whose go.mod still carries a filesystem replace). After the first tag:
add the module to `integration/go.mod` + the `documentedPins` fixture and
the CI proxy-smoke matrix.

## Build & verify

Hermetic per-module commands (how consumers resolve via the proxy):

```bash
cd systemd && GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
```

(vet/build with the same prefix). Lint from inside the module:
`cd systemd && golangci-lint run ./...` — never from the workspace root
(the root config's depguard allowlist covers only core + family deps),
never concurrently with other modules' lint runs.

## Dependencies

| Module                             | Version | Role                                     |
| ---------------------------------- | ------- | ---------------------------------------- |
| `github.com/coreos/go-systemd/v22` | v22.7.0 | sd_notify protocol (unixgram transport)  |
| `github.com/larsartmann/go-appkit` | v0.7.0* | `ServiceConfig` hook slices, `Hook` type |

\* dev: replaced onto `../` until the first core tag with `StartHooks`
ships; the reference implementation for the notify/watchdog sequence is
`go-daemon`'s `socket.go` (verified 2026-10-01, no go-daemon dependency).
