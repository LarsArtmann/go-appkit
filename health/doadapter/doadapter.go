// Package doadapter is opt-in glue between the appkit health surface and a
// samber/do injector: it adapts [appkithealth.Mounted] to [do.Shutdowner] so
// that injector.Shutdown() also drains and stops the health surface.
//
// # Why a separate package
//
// The parent health package stays injector-free BY CONTRACT: its API takes
// no injector and this module's public health surface must not grow a
// samber/do dependency. A previous plan considered housing the adapter in
// the flightrecorderhealth module instead; that was rejected — it would add
// a permanent flightrecorderhealth → health family edge for six lines of
// glue, and flightrecorderhealth sits beside (not above) the health module.
// An opt-in subpackage keeps the dependency direction clean: only consumers
// who WANT the bridge import it.
//
// # Usage
//
// Register the adapter after mounting the surface so injector.Shutdown()
// (for example on scope close, or via do's lifecycle hooks) tears the health
// surface down as part of the injector's own shutdown:
//
//	injector := do.New()
//	// ... register health-checkable services ...
//	mounted, err := appkithealth.New(probe)
//	// ...
//	do.ProvideNamedValue(injector, "health-surface", doadapter.AsShutdowner(mounted))
//
// # Ordering
//
// The adapter is a COMPLEMENT to the appkit ServiceConfig wiring, not a
// replacement: mounted.Drain in a DrainHook flips readiness 503 at the start
// of the service's drain window, while injector shutdown typically runs
// later (or not at all — many services never shut their injector). The
// adapter calls Mounted.Shutdown, which drains first (MarkShuttingDown) and
// then stops the refresh loop and the dashboard pusher, so no ordering
// guarantee is lost when both mechanisms are wired.
package doadapter

import (
	"context"

	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/samber/do/v2"
)

// AsShutdowner adapts a Mounted health surface to [do.Shutdowner]: calling
// Shutdown (as injector.Shutdown does for every registered Shutdowner)
// drains the surface — readiness surfaces flip to 503 — and then stops the
// probe's refresh loop and the dashboard pusher. Shutdown is idempotent, so
// the adapter is safe to register even when the service also shuts the
// surface down through ServiceConfig.ShutdownHooks.
func AsShutdowner(mounted *appkithealth.Mounted) do.Shutdowner {
	return mountedShutdowner{mounted: mounted}
}

type mountedShutdowner struct {
	mounted *appkithealth.Mounted
}

func (s mountedShutdowner) Shutdown() {
	_ = s.mounted.Shutdown(context.Background())
}
