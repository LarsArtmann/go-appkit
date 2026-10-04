package systemd

import (
	"log/slog"
	"sync/atomic"
	"time"

	sd "github.com/coreos/go-systemd/v22/daemon"
	errorfamily "github.com/larsartmann/go-error-family"
)

// NotifyCounters is a snapshot of the process-global sd_notify activity
// counters. Only successful sends count: outside systemd (no
// NOTIFY_SOCKET) everything stays zero, so a nonzero counter against a
// non-notify unit is itself a misconfiguration signal — and a counter
// frozen while its counterpart keeps advancing is the operational alarm
// (watchdog pings stalling under an armed WatchdogSec means the process
// stopped making progress).
type NotifyCounters struct {
	// ReadySent counts successful READY=1 notifications.
	ReadySent int64

	// StoppingSent counts successful STOPPING=1 notifications.
	StoppingSent int64

	// WatchdogPings counts successful WATCHDOG=1 keepalive pings.
	WatchdogPings int64

	// NotifyFailures counts failed send attempts across all states.
	NotifyFailures int64
}

// The counters are process-global because sd_notify is: $NOTIFY_SOCKET is
// per-process, so scrape-time readers (Prometheus collectors) must not
// need a handle to the Hooks that sent the notifications. Monotonic for
// the life of the process; safe for concurrent use.
//
//nolint:gochecknoglobals // sd_notify state is process-global by protocol
var (
	notifyReadySent     atomic.Int64
	notifyStoppingSent  atomic.Int64
	notifyWatchdogPings atomic.Int64
	notifyFailures      atomic.Int64
)

// Counters returns a snapshot of the process-global sd_notify activity
// counters, safe for concurrent use from any goroutine.
func Counters() NotifyCounters {
	return NotifyCounters{
		ReadySent:      notifyReadySent.Load(),
		StoppingSent:   notifyStoppingSent.Load(),
		WatchdogPings:  notifyWatchdogPings.Load(),
		NotifyFailures: notifyFailures.Load(),
	}
}

// notifyKind selects which activity counter a successful send feeds.
type notifyKind string

const (
	notifyReady    notifyKind = "READY"
	notifyStopping notifyKind = "STOPPING"
	notifyWatchdog notifyKind = "WATCHDOG"
)

// notify sends one sd_notify(3) state line to the socket named by
// $NOTIFY_SOCKET. It reports sent=false with a nil error when the variable
// is unset — i.e. when not running under systemd — so callers can treat
// the whole module as a no-op there. A real send failure is classified as
// an Infrastructure error under the systemd.notify_failed code. Successful
// sends advance the process-global counters per kind; failures advance
// NotifyFailures.
func notify(kind notifyKind, state string) (bool, error) {
	sent, err := sd.SdNotify(false, state)
	if err != nil {
		notifyFailures.Add(1)

		return false, errorfamily.WrapInfrastructuref(err, "systemd.notify_failed", "sd_notify %s", state)
	}

	if !sent {
		return sent, nil
	}

	switch kind {
	case notifyReady:
		notifyReadySent.Add(1)
	case notifyStopping:
		notifyStoppingSent.Add(1)
	case notifyWatchdog:
		notifyWatchdogPings.Add(1)
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
			sent, err := notify(notifyWatchdog, sd.SdNotifyWatchdog)
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
