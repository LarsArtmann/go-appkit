// Package flightrecorderhealth bridges [github.com/larsartmann/go-flightrecorder]
// with [github.com/larsartmann/go-health], exposing flight-recorder state and
// health-check failures through the health dashboard.
//
// # What this does
//
// The flight recorder continuously buffers Go runtime trace data in memory. When
// a service becomes unhealthy, a trace snapshot captures the problematic time
// window for offline analysis with `go tool trace`.
//
// This package provides two integration points:
//
//  1. [Checkable] — a health-checkable wrapper that reports the recorder's own
//     operational state (enabled, buffer active) in the health dashboard.
//  2. [Trigger] — a [health.HealthRecorder] that intercepts every health-check
//     batch and triggers a trace snapshot when any service fails.
//
// # Quick start
//
// Register the recorder as a health-checkable service so its status appears in
// the dashboard, then wire the trigger to capture a trace snapshot when any
// health check fails:
//
//	rec, err := fr.New(fr.WithSnapshotDir("/var/traces"))
//	if err != nil { /* handle */ }
//	if err := rec.Start(); err != nil { /* handle */ }
//	defer rec.Close()
//
//	injector := do.New()
//	frhealth.Register(injector, rec, "flight-recorder")
//
//	probe := health.New(injector,
//	    health.WithCriticalServices("database", "bot"),
//	    health.WithHealthRecorder(frhealth.NewTrigger(rec,
//	        frhealth.WithTriggerFunc(fr.OnError()),
//	        frhealth.WithServiceName("flight-recorder"),
//	    )),
//	)
//
// # Process-global singleton
//
// Go's runtime/trace allows only one active flight recorder per process. Create
// a single recorder at startup and share it across all integrations.
//
// # Import aliasing
//
// This package name is long. Alias it on import for readability:
//
//	import (
//	    fr "github.com/larsartmann/go-flightrecorder"
//	    "github.com/larsartmann/go-health"
//	    frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
//	)
package flightrecorderhealth
