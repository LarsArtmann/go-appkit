package health

import (
	"context"
	"fmt"

	health "github.com/larsartmann/go-health"
	"github.com/samber/do/v2"
)

// ExampleNewProbe_recorderViaInjector demonstrates the injector path, the
// only construction route that honors health.WithHealthRecorder: go-health's
// NewWithHealthCheck — which NewProbe forwards to — nils the recorder before
// assembling the probe, so any recorder option passed to NewProbe is
// silently ignored. When trace capture on health batches matters, resolve
// the checks from a samber/do injector instead, exactly as the
// flightrecorderhealth module's Register + NewTrigger wiring does.
//
// The countingRecorder stands in for flightrecorderhealth.NewTrigger: both
// implement health.HealthRecorder by delegating to the injector's
// HealthCheckWithContext, which scans every registered service that
// satisfies do's health-checker interfaces.
func ExampleNewProbe_recorderViaInjector() {
	injector := do.New()

	do.ProvideNamed(injector, "database", func(do.Injector) (*healthyDB, error) {
		return &healthyDB{}, nil
	})
	_, _ = do.InvokeNamed[*healthyDB](injector, "database")

	recorder := &countingRecorder{}

	probe := health.New(injector, health.WithHealthRecorder(recorder))

	response := probe.Evaluate(context.Background())

	fmt.Println("batches seen by recorder:", recorder.batches)
	fmt.Println("database check:", response.Checks["database"].Status)
	fmt.Println("probe status:", response.Status)
	// Output:
	// batches seen by recorder: 1
	// database check: pass
	// probe status: pass
}

type healthyDB struct{}

func (*healthyDB) HealthCheck(context.Context) error { return nil }

type countingRecorder struct{ batches int }

func (c *countingRecorder) RecordHealthCheckWithContext(
	ctx context.Context,
	injector do.Injector,
) map[string]error {
	c.batches++
	return injector.HealthCheckWithContext(ctx)
}
