package systemd

import (
	"context"
	"log/slog"
	"sync"

	sd "github.com/coreos/go-systemd/v22/daemon"
	appkit "github.com/larsartmann/go-appkit"
)

// Hooks is the sd_notify lifecycle expressed as the three appkit hook
// kinds, so a service announces systemd's Type=notify protocol at exactly
// the right lifecycle points. Build it with [New] (or wire all of it at
// once with [Install]) and append the fields to the matching
// [appkit.ServiceConfig] slices. Each hook is independently usable — a
// service that only wants STOPPING=1 can mount Drain alone.
type Hooks struct {
	// Start sends READY=1 and starts the WATCHDOG=1 keepalive when the
	// manager requested one. It belongs in StartHooks: appkit runs those
	// after the listener binds and before serving, which is exactly the
	// Type=notify contract — the manager learns the service is up only
	// once it can actually take traffic. A send failure returns an error
	// and therefore fails the whole start. Outside systemd it is a no-op.
	Start appkit.Hook

	// Drain sends STOPPING=1. It belongs in DrainHooks: appkit runs those
	// at the start of the drain window, while in-flight traffic is still
	// served, so the manager observes the shutdown beginning before
	// connections are released. A send failure returns an error (joined
	// into the shutdown result). Outside systemd it is a no-op.
	Drain appkit.Hook

	// Shutdown stops the watchdog keepalive. It belongs in ShutdownHooks
	// — the final lifecycle phase, after connections are released —
	// because whether the manager enforces the watchdog after STOPPING=1
	// is version-dependent: pinging to the very end is safe either way.
	// Idempotent and side-effect-free when no watchdog is running.
	// ShutdownHooks never run for a service whose start failed; on such
	// paths call it yourself (or exit the process, which ends the
	// keepalive anyway).
	Shutdown appkit.Hook
}

// Install appends the full sd_notify lifecycle ([Hooks.Start], Drain, and
// Shutdown) to cfg's StartHooks, DrainHooks, and ShutdownHooks slices, in
// that order. It appends — never replaces — so consumer hooks compose
// around it. Call it once, before [appkit.NewService]; there is nothing
// to uninstall (the hooks are plain config values).
func Install(cfg *appkit.ServiceConfig, opts ...Option) {
	hooks := New(opts...)

	cfg.StartHooks = append(cfg.StartHooks, hooks.Start)
	cfg.DrainHooks = append(cfg.DrainHooks, hooks.Drain)
	cfg.ShutdownHooks = append(cfg.ShutdownHooks, hooks.Shutdown)
}

// hooksConfig carries the New/Install options.
type hooksConfig struct {
	logger *slog.Logger
}

// Option configures the sd_notify lifecycle hooks.
type Option func(*hooksConfig)

// WithLogger sets the logger for notify diagnostics (default
// slog.Default). appkit consumers usually pass their service logger;
// successful notifications log at Debug, degraded behavior at Warn.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *hooksConfig) { cfg.logger = logger }
}

// New builds the sd_notify lifecycle hooks: READY=1 post-listen,
// STOPPING=1 at drain start, and a watchdog keepalive (at WatchdogSec/2)
// that runs from READY=1 until Shutdown fires.
func New(opts ...Option) Hooks {
	cfg := hooksConfig{logger: slog.Default()}

	for _, opt := range opts {
		opt(&cfg)
	}

	stop := make(chan struct{})

	var stopOnce sync.Once

	return Hooks{
		Start: func(context.Context) error {
			sent, err := notify(sd.SdNotifyReady)
			if err != nil {
				return err
			}

			if !sent {
				cfg.logger.Debug("sd_notify skipped: no NOTIFY_SOCKET (not under systemd)")

				return nil
			}

			cfg.logger.Debug("sd_notify READY=1 sent")

			interval, err := watchdogInterval()
			if err != nil {
				// Degrade, never abort: the service itself is healthy; a
				// failed watchdog-config read must not fail its start.
				cfg.logger.Warn("sd_watchdog check failed", "error", err)

				return nil
			}

			if interval <= 0 {
				return nil
			}

			pingEvery := interval / 2

			go runWatchdog(cfg.logger, stop, pingEvery)

			cfg.logger.Debug("watchdog keepalive started", "interval", pingEvery)

			return nil
		},
		Drain: func(context.Context) error {
			sent, err := notify(sd.SdNotifyStopping)
			if err != nil {
				return err
			}

			if sent {
				cfg.logger.Debug("sd_notify STOPPING=1 sent")
			}

			return nil
		},
		Shutdown: func(context.Context) error {
			stopOnce.Do(func() { close(stop) })

			return nil
		},
	}
}
