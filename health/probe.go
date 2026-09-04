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
			mu sync.Mutex
			wg sync.WaitGroup
		)

		for name, check := range checks {
			wg.Go(func() {
				defer recordCheckPanic(name, results, &mu)

				err := check(ctx)

				mu.Lock()
				results[name] = err
				mu.Unlock()
			})
		}

		wg.Wait()

		return results
	}

	return health.NewWithHealthCheck(batch, opts...)
}

// recordCheckPanic converts a panic inside a check goroutine into that
// check's error. Registered before the check runs so it also catches panics
// thrown mid-execution; recover is nil-valued on the normal path.
func recordCheckPanic(name string, results map[string]error, mu *sync.Mutex) {
	recovered := recover()
	if recovered == nil {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	results[name] = errorfamily.Newf(
		errorfamily.Infrastructure,
		"health.check_panicked",
		"check %q panicked: %v",
		name,
		recovered,
	)
}
