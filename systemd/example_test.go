package systemd_test

import (
	"fmt"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/systemd"
)

// ExampleInstall wires the full sd_notify lifecycle into a service config
// in one call: READY=1 post-listen, STOPPING=1 at drain start, watchdog
// keepalive + stop. The hooks are APPENDED, never replaced, so consumer
// hooks compose around them. Outside systemd (no NOTIFY_SOCKET) every
// hook is a no-op, so the same binary runs unchanged in development.
func ExampleInstall() {
	cfg := appkit.DefaultServiceConfig()
	systemd.Install(&cfg)

	fmt.Println(len(cfg.StartHooks), len(cfg.DrainHooks), len(cfg.ShutdownHooks))
	// Output: 1 1 1
}

// ExampleNew builds the three hooks as plain values for manual wiring —
// each is independently usable, so a service that only wants STOPPING=1
// can mount Hooks.Drain alone.
func ExampleNew() {
	hooks := systemd.New()

	fmt.Println(hooks.Start != nil, hooks.Drain != nil, hooks.Shutdown != nil)
	// Output: true true true
}
