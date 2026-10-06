# Upstream draft: go-flightrecorder — a rate-limiting trigger combinator (`OnCooldown` / `WithCooldown` trigger option)

**Status:** DRAFT — filing is USER-gated (verify-before-filing gate satisfied 2026-10-06; see self-review below). Do not file without explicit approval.
**Target repo:** github.com/larsartmann/go-flightrecorder (own repo — decision note, not an external ask)
**Evidence versions:** v0.2.1 (verified at HEAD `48dfd20` post-race-fix; the relevant surface is unchanged since v0.2.0)

## The ask

Triggers fire per event, and nothing in the library rate-limits them: with a
dir sink (the documented append-and-retain pattern — `SnapshotToDir` is
deliberately not once-latched), a flapping dependency that fires `OnError`
hundreds of times in a burst writes hundreds of trace files and burns the
retention budget on near-identical captures. Add a cooldown combinator so a
trigger can self-throttle, e.g.:

```go
// OnCooldown wraps trigger and enforces a minimum interval between fires.
// The first matching event captures; further matches inside the window are
// dropped (and do not re-arm the window).
fr.OnCooldown(30*time.Second, fr.OnError())

// or, symmetric with the existing option style:
fr.New(recorderOptions..., fr.WithCooldownTrigger(30*time.Second)) // applies to trigger evaluation
```

The exact shape is upstream's call — the essential property is: **one
capture per window per trigger chain, decided at trigger-evaluation time,
with no consumer-side bookkeeping.**

## Evidence

Every consumer that captures on repeatable events needs this, and the first
one in the fleet hand-rolled it. go-appkit's flightrecorderhealth adapter
(`adapter.go`, `Trigger.RecordHealthCheckWithContext`) carries a private
mutex-plus-timestamp implementation:

```go
t.mu.Lock()

if t.cooldown > 0 && !t.lastCapture.IsZero() && time.Since(t.lastCapture) < t.cooldown {
    t.mu.Unlock()

    return results
}

captured := t.rec.SnapshotIfAsync(context.WithoutCancel(ctx), tc, t.triggerFunc)

if captured {
    t.lastCapture = time.Now()
}

t.mu.Unlock()
```

Every adapter with a different shape (the HTTP middleware in
go-appkit/flightrecorder, the cqrs projection triggers) re-invents the same
bookkeeping or goes without. The HTTP middleware case is the sharpest: it
cannot wrap the recorder (consumer-built) and today documents a "write your
own `minInterval` atomic wrapper" recipe in its package docs — three lines
of atomics that every consumer must get right (the once-latch dedupes a
burst only for writer sinks, and only until `Reset` re-arms it).

## Why upstream is the right layer

1. **The trigger package already owns composition** (`OnAll`/`OnAny`/
   `OnErrorOrLatency`) — cooldown is a fourth combinator, not a new concept.
2. **Retention is already upstream's problem** (`WithMaxSnapshots` prunes
   AFTER the write); cooldown prevents the write, which retention cannot do
   (the damage — write churn, goroutine spawn per event, gzip CPU — precedes
   the prune).
3. **A single vetted implementation** replaces per-consumer atomic-latch
   code, including the subtlety of whether a dropped event inside the window
   should extend the window (it should not) and whether the first event
   always captures (it should).

## Proposed semantics (minimal, prescriptive only where it matters)

- `OnCooldown(window, wrapped TriggerFunc) TriggerFunc`: passes through when
  the wrapped trigger is false. On a wrapped true: if a prior fire is inside
  the window, drop (return false); else record `now` and return true.
- Dropped events do NOT extend the window (fixed-window, not sliding) —
  matches the frh hand-roll and is the least surprising for retention math.
- Zero/negative window = pass-through (degenerate to `wrapped`), mirroring
  how `WithCompression(0)` means off.
- Concurrency: one atomic `int64` (unix-nano of last fire) suffices — no
  mutex, matching the library's zero-dependency posture. (CAS or
  load-then-store both fine; the window math tolerates a rare lost update
  under contention.)

## Self-review (verify-before-filing gate, 2026-10-06)

- [x] Feature ABSENT upstream: `grep -in "cooldown|rate.limit|throttle"`
      over go-flightrecorder's TODO_LIST.md, ROADMAP.md, FEATURES.md at
      HEAD `48dfd20` — zero hits. Not a duplicate; not already planned.
- [x] Evidence VERIFIED at source: the frh code block above is quoted from
      `flightrecorderhealth/adapter.go` (2026-10-06 working tree; the
      cooldown path predates this draft and ships in frh v0.1.6).
- [x] Consumer pain is REAL and RECURRING: two call sites (frh adapter;
      go-appkit/flightrecorder doc.go cookbook recipe) independently
      implement or prescribe cooldown logic.
- [x] API proposal fits the existing surface (trigger combinators + option
      style); no new dependency; no breaking change.
- [ ] Filed — BLOCKED on owner approval (this repo's rule for upstream asks).
