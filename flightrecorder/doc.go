// Package flightrecorder integrates [github.com/larsartmann/go-flightrecorder]
// with HTTP services, providing middleware that automatically captures execution
// traces when requests trigger configurable conditions (errors, latency spikes).
//
// The underlying flight recorder continuously buffers the last few seconds of
// Go runtime trace data in memory. When a problem occurs, a snapshot of exactly
// the problematic time window is written for offline analysis with
// `go tool trace`.
//
// # Quick start
//
// Create a recorder, start it, and wire the middleware into your service:
//
//	rec, err := flightrecorder.New(
//	    flightrecorder.WithSnapshotFile("/tmp/trace.out"),
//	)
//	if err != nil { /* handle */ }
//
//	if err := rec.Start(); err != nil { /* handle */ }
//	defer rec.Close()
//
//	cfg := appkit.DefaultServiceConfig()
//	cfg.ExtraMiddlewares = []httputil.Middleware{
//	    flightrecorder.Middleware(rec, fr.OnErrorOrLatency(100*time.Millisecond)),
//	}
//
// Mount a manual snapshot endpoint for on-demand debugging:
//
//	svc, _ := appkit.NewService(cfg)
//	flightrecorder.Mount(svc.Mux, "POST /debug/flightrecorder/snapshot", rec)
//
// # How the middleware works
//
// For each request, the middleware:
//  1. Wraps the ResponseWriter to capture the HTTP status code.
//  2. Measures request duration.
//  3. After the handler completes, constructs a [fr.TriggerContext] with
//     Kind="http", Type="METHOD /path", Duration, and Err (non-nil if status
//     exceeds the error threshold, default 500).
//  4. Evaluates the trigger function. If it returns true, captures a snapshot.
//  5. Resets the recorder's once-latch so subsequent problematic requests
//     can also capture traces.
//
// The once-latch from go-flightrecorder prevents snapshot races when multiple
// goroutines detect problems simultaneously. Only the first caller in a burst
// captures a trace; the latch is then re-armed via Reset for the next event.
//
// # Process-global singleton
//
// Go's runtime/trace allows only one active flight recorder per process.
// Create a single recorder at startup and share it across all middleware
// instances and handlers.
//
// # Import aliasing
//
// This package is named flightrecorder, same as the underlying library. When
// importing both, alias the underlying library:
//
//	import (
//	    fr "github.com/larsartmann/go-flightrecorder"
//	    "github.com/larsartmann/go-appkit/flightrecorder"
//	)
package flightrecorder
