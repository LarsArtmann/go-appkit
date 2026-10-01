// Package systemd wires the systemd service-manager notifications
// (sd_notify) into an appkit service lifecycle: READY=1 once the listener
// is bound, STOPPING=1 at the start of the graceful drain, and the
// WATCHDOG=1 keepalive while the service runs.
//
// It exists for services run under systemd with Type=notify (and,
// optionally, WatchdogSec) — with that unit configuration the service
// manager delays considering the service up until READY=1 arrives, kills
// it when the watchdog timestamp goes stale, and tracks its shutdown via
// STOPPING=1. Outside systemd every hook degrades to a no-op, so the same
// binary runs unchanged in development.
//
// # Quick start
//
//	cfg := appkit.DefaultServiceConfig()
//	systemd.Install(&cfg)
//
//	svc, err := appkit.NewService(cfg)
//	if err != nil {
//		panic(err)
//	}
//	// register routes, then:
//	err = svc.Run(context.Background())
//
// with a unit file like:
//
//	[Service]
//	Type=notify
//	WatchdogSec=30
//	# standard hardening omitted
//
// # Ordering contract
//
// The hooks ride appkit's lifecycle seams and inherit their contracts:
//
//   - Start ([Hooks.Start]) runs post-listen, pre-serve (a
//     [appkit.ServiceConfig.StartHooks] hook): READY=1 is only sent after
//     the socket is bound. A send failure fails the whole start — a
//     service the manager never learned about must not serve.
//   - Drain ([Hooks.Drain]) runs at the start of the drain window (a
//     DrainHooks hook), while the listener still serves in-flight
//     traffic: STOPPING=1 reaches the manager before the process begins
//     releasing connections.
//   - the watchdog keepalive pings at WatchdogSec/2 from READY=1 until
//     [Hooks.Shutdown] runs in the final phase (ShutdownHooks): it keeps
//     pinging through the whole drain and shutdown, because whether the
//     manager still enforces the timeout after STOPPING=1 is
//     version-dependent — pinging is safe either way, stopping early is
//     not.
//
// Install appends to the config's hook slices — consumer hooks compose
// before and after.
//
// This module is transport-only: it sends notifications, it does not parse
// unit files, manage socket activation, or read the journal.
package systemd
