## Finding 1 (bug): a recovered per-check panic is classified WARN when the check is non-critical, contradicting `ErrCheckPanicked`'s contract

`ErrCheckPanicked`'s godoc (`checks.go:87-89`) is explicit:

> like the batch-level [ErrPanicDuringHealthCheck], a recovered panic must
> read as a failure, never a warning.

But `classifier.grades` (`classifier.go:99-113`) only special-cases the
BATCH-level sentinel:

```go
func (c classifier) grades(name string, err error) Status {
	if errors.Is(err, ErrPanicDuringHealthCheck) {
		return StatusFail
	}
	// ...
	if _, critical := c.critical[name]; critical {
		return StatusFail
	}

	return StatusWarn
}
```

A per-check panic arrives as `ErrCheckPanicked` — which does not match —
so a non-critical service that panics grades `StatusWarn`, exactly what
the godoc forbids. Minimal fix: `errors.Is(err, ErrCheckPanicked)` joins
the first branch. go-appkit pins the current (warn) behavior by test,
so whichever way this goes, say it and I'll flip the pin.

## Finding 2 (decision): `WithHealthRecorder` on the function-path constructors is silently dead

`NewWithHealthCheck`/`NewWithDetailedCheck` accept the option and nil it
(`accessors.go:55-57` documents the drop in prose). The flagship wiring
(`flightrecorderhealth.NewTrigger` + `WithHealthRecorder`) compiles and
runs against the function path and silently captures nothing — go-appkit
hit exactly this and now warns in its own godoc. Options: reject with a
sentinel (`ErrRecorderNotApplicable`, breaking at 0.x), or expose
`Probe.Recorder() HealthRecorder` so callers can assert non-nil in
tests. I lean sentinel — the accessor still permits the silent
mis-wiring in production.

## Finding 3 (docs): `Probe.Evaluate` does not publish to the served cache

`Evaluate` (`probe.go:657-685`) runs a full batch but never publishes to
the background cache — only the refresh loop does. A test that flips a
dependency and asserts through `/readyz` or `CachedResponse` keeps
seeing the stale pass. Worth a godoc line ("does not update the cache")
or an `EvaluateAndCache` variant.

## Finding 4 (decision): `Shutdown` is a one-way valve but `Start` reads re-armable

`shuttingDown` is `Store(true)`-only (`probe.go:606/:646` — never
cleared). A consumer who stops and re-Starts a probe gets a surface that
serves routes but reports readiness 503 forever (runtime-verified
2026-09-20). Either clear the latch in `Start` or document the one-way
valve on both methods.
