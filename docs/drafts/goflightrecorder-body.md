## Problem

Triggers fire per event and nothing in the library rate-limits them. With
a dir sink (the documented append-and-retain pattern — `SnapshotToDir`
is deliberately not once-latched), a flapping dependency that fires
`OnError` hundreds of times in a burst writes hundreds of trace files
and burns the retention budget on near-identical captures.

## Ask

A cooldown combinator so a trigger can self-throttle:

```go
// OnCooldown wraps trigger and enforces a minimum interval between fires.
// The first matching event captures; further matches inside the window are
// dropped (and do not re-arm the window).
fr.OnCooldown(30*time.Second, fr.OnError)
```

The exact shape is upstream's call — the essential property is **one
capture per window per trigger chain, decided at trigger-evaluation
time, with no consumer-side bookkeeping**.

## Evidence

The first fleet consumer hand-rolled it: go-appkit's
`flightrecorderhealth` adapter carries a private mutex-plus-timestamp
implementation (`WithCooldown`, `lastCapture` guarded by `sync.Mutex`
for concurrent batch safety), and the go-appkit flightrecorder
middleware's doc.go cookbook prescribes a DIY min-interval recipe. Two
independent implementations of the same throttle = the feature belongs
in the trigger layer.
