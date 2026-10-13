## Problem

`System.GracefulClose`'s godoc (`system/lifecycle.go:119-121` at `87fc03f7`) promises:

> If the context expires during draining, Close is still attempted; if it
> expires during Close, resources may still be closing in the background.

The implementation does not keep that promise. Phase 1 returns early on the
first drainer error and never reaches the Close phase:

```go
// system/lifecycle.go:130-133
// Phase 1: drain in-flight work.
if err := s.drainAll(ctx); err != nil {
    return fmt.Errorf("system: graceful drain: %w", err) // Close never runs
}
```

A context expiry surfaces as `ctx.Err()` from the drainer, so the common
production case — a SIGTERM with a stuck in-flight handler outliving the
grace period — leaks every engine, the projection host, and all
`RegisterCloser` resources. For an API whose stated purpose is "use this
instead of Close when you need a shutdown deadline", the deadline expiring
must degrade to a raced Close, not to no Close at all.

## Fix candidates

1. Match the godoc: on a drain error, still race `Close()` against the
   remaining context (or run it unconditionally with a fresh background
   context) and `errors.Join(drainErr, closeErr)`.
2. Or weaken the godoc and export the seam honestly, so wrappers don't
   have to re-derive the correct sequencing.

One nuance for whichever fix lands: Phase 2's select has a deliberate
"pre-expired context wins deterministically" guard
(`system/lifecycle.go:140-147`, flake note inline) — a drain-error path
re-entering Phase 2 with an already-expired context must not regress that
property. The simplest correct shape runs Close unconditionally once a
drain error was already collected.

## Consumer evidence

go-appkit's `cqrs.EventService.Shutdown` cannot call `GracefulClose`
because of exactly this gap — a drain timeout would skip the close and
leak engines. It hand-rolls the documented semantics from the public
pieces instead:

```go
drainErr := es.sys.Drain(ctx)
closeErr := es.sys.Close() // ALWAYS — deadline cannot skip it
return errors.Join(drainErr, closeErr)
```

plus `TestEventService_Shutdown_DrainTimeoutStillCloses`, which parks an
in-flight command, expires the drain context, and asserts the registered
closer ran. That wrapper test would be redundant if `GracefulClose`
matched its own contract.

## Related, already fixed

The original finding bundled a corollary — "`Close()` never stops managed
timers" — which is FIXED at master (`system/lifecycle.go:88-91`: `Close`
now calls `stopTimersLocked()`, idempotent; go-appkit's tripwire flipped
on the 2026-10-12 sweep). This issue tracks only the drain-error early
return.
