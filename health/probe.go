package health

import (
	"context"

	"github.com/larsartmann/go-health"
)

// CheckFunc is one named health check. Return nil when the dependency is
// healthy, an error otherwise. Checks run concurrently on every batch —
// background refreshes and live handler requests alike — and receive the
// batch's timeout-bounded context (5s by default, tunable through the SDK's
// WithTimeout). A check that ignores its context is abandoned at the batch
// deadline with a fail-closed error naming it.
type CheckFunc func(ctx context.Context) error

// NewProbe creates a go-health Probe from named checks, without a samber/do
// injector. It forwards to go-health's NewChecks: checks run concurrently
// on every batch and each is self-timed, so every served response carries
// per-check duration_ns alongside the probe-observed since (dashboards
// rendering durations light up without any wiring change). A nil or empty
// check map produces a probe that always reports pass — useful as a
// placeholder, pointless as a health surface.
//
// # WithHealthRecorder is rejected on this path
//
// NewProbe forwards opts to go-health's NewChecks, where the given checks own
// batch execution and a recorder has nothing to contribute. Through go-health
// v0.7.0 that option was silently dropped: no HealthRecorder callbacks fired,
// a flightrecorderhealth.Trigger wired this way captured zero traces, and
// nothing said so.
//
// Since go-health v0.8.0 the mis-wiring fails at boot: `Start` returns
// [health.ErrRecorderNotApplicable] naming the constructor, so a silent
// zero-trace deployment is no longer possible. Prefer the construct that says
// what you mean:
//
//   - recorder callbacks (traces, audit logs): register the checks in a
//     samber/do injector and build the probe with health.New, which honors the
//     recorder. See the package doc's quick start and the runnable
//     ExampleNewProbe_recorderViaInjector.
//   - no injector but a recorder that owns its checks: go-health's
//     NewWithRecorder.
//   - plain functions or a named-check map: NewProbe, as here.
//
// Failure modes, all owned by go-health's per-check runner: a panicking
// check is recovered into that check's error (wrapping go-health's
// ErrCheckPanicked — "health: check panicked: ...") instead of poisoning
// the batch, and other checks still report; a nil CheckFunc in the map
// fails closed as that check's error (wrapping ErrNilCheck); a check that
// ignores its context is abandoned at the batch deadline with a fail-closed
// error naming it, so one wedged dependency cannot hang the batch. Status
// grading is unchanged: critical failures fail readiness, non-critical
// failures degrade to warn.
//
// All other probe behavior (caching interval, batch timeout, method guard,
// evaluation hooks) is configured through the SDK's own Option values.
// Callers that must self-time composed checks (proxying an upstream probe's
// own timings, batching externally) can use go-health's NewWithDetailedCheck
// directly — the resulting *health.Probe composes with everything in this
// package unchanged.
func NewProbe(checks map[string]CheckFunc, opts ...health.Option) *health.Probe {
	goChecks := make(map[string]health.CheckFunc, len(checks))

	for name, check := range checks {
		goChecks[name] = health.CheckFunc(check)
	}

	return health.NewChecks(goChecks, opts...)
}
