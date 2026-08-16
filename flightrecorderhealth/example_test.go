package flightrecorderhealth_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/samber/do/v2"
)

// failingDB is a stand-in service whose health check always fails, so the
// trigger examples capture without needing a real dependency.
type failingDB struct{}

func (failingDB) HealthCheck(_ context.Context) error {
	return errors.New("connection refused")
}

// ExampleRegister wires a flight recorder into a samber/do injector so its
// operational state appears as a row in the health dashboard.
func ExampleRegister() {
	rec, err := fr.New(fr.WithWriter(io.Discard))
	if err != nil {
		panic(err) // production code: handle the error
	}

	injector := do.New()
	frhealth.Register(injector, rec, "flight-recorder")

	// The recorder is not started in this example, so it reports unhealthy —
	// proof that the recorder's own state is visible in the dashboard.
	results := injector.HealthCheckWithContext(context.Background())

	fmt.Println("flight-recorder healthy:", results["flight-recorder"] == nil)
	// Output: flight-recorder healthy: false
}

// ExampleNewCheckable constructs a health-checkable wrapper with a custom
// display name for the dashboard.
func ExampleNewCheckable() {
	rec, err := fr.New(fr.WithWriter(io.Discard))
	if err != nil {
		panic(err) // production code: handle the error
	}

	checkable := frhealth.NewCheckable(rec, frhealth.WithCheckableName("trace-recorder"))

	fmt.Println(checkable.Name())
	// Output: trace-recorder
}

// ExampleNewTrigger captures a trace snapshot when any health check fails.
// The default trigger function is fr.OnError, and the capture itself is
// asynchronous, so the health-check loop is never delayed by trace I/O.
func ExampleNewTrigger() {
	rec, err := fr.New(fr.WithWriter(io.Discard))
	if err != nil {
		panic(err) // production code: handle the error
	}

	trigger := frhealth.NewTrigger(rec,
		frhealth.WithTriggerFunc(fr.OnError()),
		frhealth.WithCooldown(30*time.Second),
	)

	injector := do.New()
	do.ProvideNamed(injector, "database", func(_ do.Injector) (failingDB, error) {
		return failingDB{}, nil
	})
	_, _ = do.InvokeNamed[failingDB](injector, "database")

	results := trigger.RecordHealthCheckWithContext(context.Background(), injector)

	fmt.Println("database failing:", results["database"] != nil)
	// Output: database failing: true
}
