package health_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-health"
)

// ExampleNewProbe builds a probe from named checks: a critical database
// check that passes and a flapping cache check that fails. A failing
// non-critical check degrades the grade to warn but keeps readiness up —
// the service still receives traffic.
func ExampleNewProbe() {
	probe := appkithealth.NewProbe(
		map[string]appkithealth.CheckFunc{
			"db": func(_ context.Context) error { return nil },
			"cache": func(_ context.Context) error {
				return errors.New("evictions spiking")
			},
		},
		health.WithCriticalServices("db"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startErr := probe.Start(ctx)
	if startErr != nil {
		panic(startErr)
	}

	defer probe.Shutdown()

	readyErr := probe.AwaitReady(ctx)
	if readyErr != nil {
		panic(readyErr)
	}

	fmt.Println("status:", probe.Status())
	fmt.Println("ready:", probe.Ready())
	// Output:
	// status: warn
	// ready: true
}

// ExampleNewProbe_panickingCheck shows the panic isolation: a check that
// panics is converted into that check's classified error instead of
// poisoning the batch — the other checks still report and the probe stays
// ready.
func ExampleNewProbe_panickingCheck() {
	probe := appkithealth.NewProbe(
		map[string]appkithealth.CheckFunc{
			"stable": func(_ context.Context) error { return nil },
			"boom": func(_ context.Context) error {
				panic("index out of range")
			},
		},
		health.WithCriticalServices("stable"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startErr := probe.Start(ctx)
	if startErr != nil {
		panic(startErr)
	}

	defer probe.Shutdown()

	readyErr := probe.AwaitReady(ctx)
	if readyErr != nil {
		panic(readyErr)
	}

	fmt.Println("status:", probe.Status())
	fmt.Println("ready:", probe.Ready())
	// Output:
	// status: warn
	// ready: true
}

// ExampleMount_multiProbeAggregate mounts TWO probes on one mux: a
// critical-surface probe on the conventional paths and an
// optional-dependencies probe on its own route set. Draining the optional
// probe flips ITS readiness to 503 while the critical probe keeps serving
// ready — the lockstep drain contract, applied per surface.
func ExampleMount_multiProbeAggregate() {
	criticalProbe := appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
		"db": func(_ context.Context) error { return nil },
	})

	optionalProbe := appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
		"search": func(_ context.Context) error { return nil },
	})

	mux := http.NewServeMux()

	critical, err := appkithealth.Mount(mux, criticalProbe)
	if err != nil {
		panic(err)
	}

	optional, err := appkithealth.Mount(mux, optionalProbe,
		appkithealth.WithProbeRoutes(health.Routes{
			Liveness:  "/optional/live",
			Readiness: "/optional/ready",
			Startup:   "/optional/startup",
		}),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startErr := critical.Start(ctx)
	if startErr != nil {
		panic(startErr)
	}

	defer func() { _ = critical.Shutdown(ctx) }()

	startErr = optional.Start(ctx)
	if startErr != nil {
		panic(startErr)
	}

	defer func() { _ = optional.Shutdown(ctx) }()

	fmt.Println("critical ready:", critical.Ready())
	fmt.Println("optional ready:", optional.Ready())

	optional.Drain() // load balancers now see 503 on /optional/ready

	fmt.Println("optional ready after drain:", optional.Ready())
	fmt.Println("critical ready after drain:", critical.Ready())
	// Output:
	// critical ready: true
	// optional ready: true
	// optional ready after drain: false
	// critical ready after drain: true
}
