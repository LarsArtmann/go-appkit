package systemd

import (
	"log/slog"
	"time"

	sd "github.com/coreos/go-systemd/v22/daemon"
	errorfamily "github.com/larsartmann/go-error-family"
)

// notify sends one sd_notify(3) state line to the socket named by
// $NOTIFY_SOCKET. It reports sent=false with a nil error when the variable
// is unset — i.e. when not running under systemd — so callers can treat
// the whole module as a no-op there. A real send failure is classified as
// an Infrastructure error under the systemd.notify_failed code.
func notify(state string) (sent bool, err error) {
	sent, err = sd.SdNotify(false, state)
	if err != nil {
		return false, errorfamily.WrapInfrastructuref(err, "systemd.notify_failed", "sd_notify %s", state)
	}

	return sent, nil
}

// watchdogInterval reports the watchdog interval systemd configured via
// $WATCHDOG_USEC, or 0 when no watchdog was requested (or the watched PID
// is not this process — the manager sets $WATCHDOG_PID for that case).
func watchdogInterval() (time.Duration, error) {
	interval, err := sd.SdWatchdogEnabled(false)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "systemd.watchdog_check_failed", "read watchdog config")
	}

	return interval, nil
}

// runWatchdog pings WATCHDOG=1 at the given interval — callers pass
// WatchdogSec/2, half the deadline, per sd_notify(3) — until stop is
// closed. A failing ping is logged and retried on the next tick: one lost
// datagram must not kill a healthy service, and the manager is the
// arbiter of when the service is truly stuck.
func runWatchdog(logger *slog.Logger, stop <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			sent, err := notify(sd.SdNotifyWatchdog)
			if err != nil {
				logger.Warn("sd_notify WATCHDOG failed", "error", err)

				continue
			}

			if !sent {
				logger.Debug("sd_notify WATCHDOG skipped: no NOTIFY_SOCKET")
			}
		}
	}
}
