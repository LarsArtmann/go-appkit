# Upstream draft: go-cqrs-lite system — `GracefulClose` skips `Close` when a drainer errors, contradicting its godoc

**Status:** DRAFT — filing is USER-gated (source-verified against the pinned tag AND master; see Evidence). Do not file without explicit approval.
**Target repo:** github.com/larsartmann/go-cqrs-lite (own repo — decision note, not an external ask)
**Evidence versions:** system/v4 v4.10.2 (the published latest); behavior verified identical on master at the time of the 2026-10-06 go-appkit adoption-closure pass.

## The ask

`System.GracefulClose`'s godoc (system.go, v4.10.2) promises:

> If the context expires during draining, Close is still attempted; if it
> expires during Close, resources may still be closing in the background.

The implementation does not keep that promise. Phase 1 returns early on the
FIRST drainer error and never reaches the Close phase:

```go
// Phase 1: drain in-flight work.
if err := s.drainAll(ctx); err != nil {
    return fmt.Errorf("system: graceful drain: %w", err) // <- Close never runs
}

// Phase 2: close with context race.  (unreached)
```

`drainAll` returns the first error (a context expiry surfaces as
`ctx.Err()` from the drainer), so the common production case — a SIGTERM
with a stuck in-flight handler outliving the grace period — leaks every
engine, the projection host, and all `RegisterCloser` resources. For a
library whose selling point is "use this instead of Close when you need a
shutdown deadline", the deadline expiring must degrade to a raced Close,
not to no Close at all.

Two candidate fixes (upstream's call):

1. Match the godoc: on a drain error, still race `Close()` against the
   remaining context (or run it unconditionally with a fresh
   background context) and `errors.Join(drainErr, closeErr)`.
2. Or weaken the godoc and export the seam honestly, so wrappers don't
   have to re-derive the correct sequencing.

## Evidence (consumer-side workaround that would not need to exist)

go-appkit's `cqrs.EventService.Shutdown` (v0.7.0 train, 2026-10-06) cannot
call `GracefulClose` because of exactly this gap — a drain-timeout would
skip the close and leak engines. It hand-rolls the documented semantics on
top of the public pieces instead:

```go
drainErr := es.sys.Drain(ctx)      // the same Phase 1
closeErr := es.sys.Close()         // ALWAYS — deadline cannot skip it
return errors.Join(drainErr, closeErr)
```

plus a regression test that parks an in-flight command, expires the drain
context, and asserts the registered closer ran
(`TestEventService_Shutdown_DrainTimeoutStillCloses`). The wrapper's test
would be redundant if `GracefulClose` matched its own contract.

One nuance for whichever fix lands: `GracefulClose`'s Phase-2 select has a
deliberate "pre-expired context wins deterministically" guard
(system.go:324-328, observed flake noted inline) — a drain-error path that
re-enters Phase 2 with an already-expired context must not regress that
property; the simplest correct shape runs Close unconditionally when a
drain error was already collected.

## Corollary found by the wrapper's behavioral tests (2026-10-06, same session)

`System.Close()` never stops managed timers: `stopTimers()` is called only
in `GracefulClose` Phase 0 (timers.go / system.go). Any consumer that
assembles its own shutdown from the public pieces — Drain + Close, exactly
what the workaround above must do — leaks every scheduler registered via
`ManageTimers` past shutdown; they keep dispatching against closed engines
until process exit. The system's own Phase-0 comment ("scheduler dispatch
must not race the drain/close phases") states the invariant; only the
GracefulClose path enforces it.

Candidate fix: make `stopTimers` idempotent (it already is — nils
`s.timerCancel`) and call it at the top of `Close()` as well, or export a
`StopTimers()` so wrappers can sequence it before Drain. Either way the
wrappers that hand scheduler lifecycle to `ManageTimers` get the documented
"system-owned concern" behavior on every shutdown path, not just
GracefulClose.

Pinned consumer-side by `cqrs/domain_test.go`
`TestEventService_DomainTimers_LifecycleOwnedBySystem` as a characterization
tripwire (asserts the scheduler is NOT stopped within 100ms of Shutdown;
fires the moment upstream lands the fix).

## Self-review (verify-before-filing gate)

- Read `system.go:262-330` and `shutdown.go` at the v4.10.2 tag directly
  (module cache), not from memory: confirmed the early return at the drain
  phase and the contradicting godoc two paragraphs above it.
- Checked master (local go-cqrs-lite working tree during the 2026-10-06
  deep-dive): same shape.
- The wrapper's drain-timeout test is green against the workaround, so the
  claim "Close still runs on drain timeout" is pinned in THIS repo's suite,
  independent of upstream's behavior.
