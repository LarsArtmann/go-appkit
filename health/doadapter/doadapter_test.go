package doadapter_test

import (
	"context"
	"testing"

	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/health/doadapter"
	"github.com/samber/do/v2"
)

// The adapter's whole purpose is the do.Shutdowner seam; pin the contract.
var _ do.Shutdowner = doadapter.AsShutdowner(nil)

func newStartedMounted(t *testing.T) *appkithealth.Mounted {
	t.Helper()

	probe := appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
		"database": func(context.Context) error { return nil },
	})
	mounted, err := appkithealth.New(probe)
	if err != nil {
		t.Fatalf("appkithealth.New: %v", err)
	}

	err = mounted.Start(context.Background())
	if err != nil {
		t.Fatalf("mounted start: %v", err)
	}

	return mounted
}

// Calling the adapter's Shutdown — what injector.Shutdown does for every
// registered Shutdowner — must drain the surface (readiness not-ready) and
// stop it, and must be safe to call twice (it is, because Mounted.Shutdown
// is idempotent).
func TestAsShutdowner_ShutsSurfaceDown(t *testing.T) {
	t.Parallel()

	mounted := newStartedMounted(t)
	if !mounted.Ready() {
		t.Fatal("expected ready after start")
	}

	shutdowner := doadapter.AsShutdowner(mounted)
	shutdowner.Shutdown()

	if mounted.Ready() {
		t.Fatal("expected not-ready after adapter shutdown")
	}

	shutdowner.Shutdown()
}

// End to end through the injector: a registered adapter runs when the scope
// shuts down, taking the health surface down with it.
func TestAsShutdowner_RunsOnInjectorShutdown(t *testing.T) {
	t.Parallel()

	mounted := newStartedMounted(t)

	injector := do.New()
	do.ProvideNamedValue[do.Shutdowner](injector, "health-surface", doadapter.AsShutdowner(mounted))

	report := injector.Shutdown()

	if report == nil {
		t.Fatal("expected a non-nil shutdown report")
	}
	if mounted.Ready() {
		t.Fatal("expected not-ready after injector shutdown")
	}
}
