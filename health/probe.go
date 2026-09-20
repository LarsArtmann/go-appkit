package health

import (
	"context"
	"sync"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-health"
)

// CheckFunc is one named health check. Return nil when the dependency is
// healthy, an error otherwise. Checks run concurrently on every batch —
// background refreshes and live handler requests alike — and receive the
// batch's timeout-bounded context (5s by default, tunable through the SDK's
// WithTimeout).
type CheckFunc func(ctx context.Context) error

// NewProbe creates a go-health Probe from named checks, without a samber/do
// injector. The probe evaluates all checks as one batch, classifies them
// through the SDK's options (WithCriticalServices gates readiness on the
// named subset), caches the roll-up in the background, and serves the
// three-probe handlers.
//
// # WithHealthRecorder is silently dropped on this path
//
// NewProbe forwards opts to go-health's NewWithHealthCheck, which nils the
// recorder before assembling the probe (go-health accessors.go:
// "cfg.recorder = nil"; its godoc states WithHealthRecorder "has no effect
// here"). A [health.WithHealthRecorder] option passed to NewProbe is
// therefore silently ignored: no HealthRecorder callbacks fire, and a
// flightrecorderhealth.Trigger wired this way captures zero traces — with no
// error and no warning. This mirrors the upstream contract ("the explicit
// function already owns batch execution"), not a defect in NewProbe, but it
// is a trap.
//
// To drive trace capture from health batches, use the injector path
// instead: register the checks in a samber/do injector, build the probe
// with health.New (that path honors WithHealthRecorder), and pass the same
// Trigger to the recorder side. See the package doc's quick start and the
// runnable ExampleNewProbeViaInjector in this package.
//
// Checks are panic-isolated per check: a panicking check fails as that
// check's error ("check %q panicked") instead of poisoning the batch. Other
// checks still report, and the classifier grades the failure by criticality
// exactly like any other failure.
//
// A nil or empty check map produces a probe that always reports pass —
// useful as a placeholder, pointless as a health surface.
//
// All other probe behavior (caching interval, batch timeout, method guard,
// evaluation hooks) is configured through the SDK's own Option values.
func NewProbe(checks map[string]CheckFunc, opts ...health.Option) *health.Probe {
	batch := func(ctx context.Context) map[string]error {
		results := make(map[string]error, len(checks))

		var (
			resultsMu sync.Mutex
			batchWG   sync.WaitGroup
		)

		for name, check := range checks {
			batchWG.Go(func() {
				defer recordCheckPanic(name, results, &resultsMu)

				err := check(ctx)

				resultsMu.Lock()
				results[name] = err
				resultsMu.Unlock()
			})
		}

		batchWG.Wait()

		return results
	}

	return health.NewWithHealthCheck(batch, opts...)
}

// recordCheckPanic converts a panic inside a check goroutine into that
// check's error. Registered before the check runs so it also catches panics
// thrown mid-execution; recover is nil-valued on the normal path.
func recordCheckPanic(name string, results map[string]error, resultsMu *sync.Mutex) {
	recovered := recover()
	if recovered == nil {
		return
	}

	resultsMu.Lock()
	defer resultsMu.Unlock()

	results[name] = errorfamily.Newf(
		errorfamily.Infrastructure,
		"health.check_panicked",
		"check %q panicked: %v",
		name,
		recovered,
	)
}
