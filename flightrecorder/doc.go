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
//	rec, err := fr.New(
//	    fr.WithFile("/tmp/trace.out"),
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
// Fetch the artifact over HTTP — append ?download=1 to stream the trace as
// an attachment (no shell access needed; honors WithSnapshotFilename):
//
//	curl -X POST 'http://localhost:8080/debug/flightrecorder/snapshot?download=1' -OJ
//
// # How the middleware works
//
// For each request, the middleware:
//  1. Wraps the ResponseWriter to capture the HTTP status code.
//  2. Measures request duration.
//  3. After the handler completes, constructs a [fr.TriggerContext] with
//     Kind="http", Type="METHOD /path", Duration, and Err (non-nil if status
//     exceeds the error threshold, default 500).
//  4. Evaluates the trigger function. If it returns true, INITIATES a
//     snapshot asynchronously ([fr.Recorder.SnapshotIfAsync] with
//     [context.WithoutCancel]) — the request never pays trace-write
//     latency. Completion telemetry (bytes, path, sink errors) belongs to
//     the recorder's [fr.WithMetrics] hook; WithLogger logs the initiation
//     with method/path/duration/status for request correlation.
//  5. With auto-reset (default), re-arms the recorder's once-latch so
//     subsequent problematic requests can also capture.
//
// The once-latch from go-flightrecorder prevents snapshot races when multiple
// goroutines detect problems simultaneously. Only the first caller in a burst
// captures a trace; the latch is then re-armed via Reset for the next event.
// Sink choice matters for repeated captures: a snapshot-dir sink writes a new
// timestamped, retained file per initiated capture (deterministic), while a
// writer/file sink is once-latched — with auto-reset re-armed per initiation,
// tight bursts may still deduplicate (prefer the dir sink; see
// [OpsRecorderPreset]).
//
// # Cookbook
//
// Capture slow failures only — [fr.OnAll] requires EVERY sub-trigger to
// fire, so fast 500s and slow 200s burn no retention budget:
//
//	mw := flightrecorder.Middleware(rec,
//	    fr.OnAll(fr.OnError(), fr.OnLatency(100*time.Millisecond)))
//
// Share ONE recorder with another subsystem (e.g. go-appkit/cqrs's
// EventConfig.FlightRecorder): Go allows a single active recorder per
// process, and a second [fr.Recorder.Start] in the same process fails with
// [fr.ErrAlreadyEnabled] (typed: [fr.AlreadyEnabledError]). Match it
// explicitly instead of string-comparing the runtime error:
//
//	if err := rec.Start(); err != nil {
//	    if errors.Is(err, fr.ErrAlreadyEnabled) { /* already running — fine */ }
//	    else { /* fail closed */ }
//	}
//
// Dir-sink captures are not rate-limited: every trigger match writes a new
// timestamped file. For flapping dependencies (a retry storm fires OnError
// repeatedly), narrow the trigger with a minimum-interval latch — a tiny
// atomic wrapper composes with any [fr.TriggerFunc] via [fr.OnAll]:
//
//	func minInterval(min time.Duration) fr.TriggerFunc {
//	    var last atomic.Int64 // unix nano of the last fire; 0 = never
//	    return func(fr.TriggerContext) bool {
//	        now := time.Now().UnixNano()
//	        if prev := last.Load(); prev != 0 && now-prev < int64(min) {
//	            return false
//	        }
//	        last.Store(now)
//	        return true
//	    }
//	}
//
//	// fr.OnAll(fr.OnError(), fr.OnLatency(100*time.Millisecond), minInterval(30*time.Second))
//
// (30–60s is a sensible range. The go-appkit/flightrecorderhealth module
// ships the same idea for health-check triggers as a
// flightrecorderhealth.WithCooldown option on NewTrigger.)
//
// # Process-global singleton
//
// Go's runtime/trace allows only one active flight recorder per process.
// Create a single recorder at startup and share it across all middleware
// instances and handlers.
//
// # Relation to go-appkit/cqrs's projection flight recorder
//
// Since cqrs v0.4.0, EventConfig.FlightRecorder takes
// [github.com/larsartmann/go-flightrecorder.Recorder] — the SAME type this
// module's middleware takes. Start ONE recorder at startup and hand it to
// both layers (Go allows only a single active recorder per process); each
// layer captures with its own trigger.
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
