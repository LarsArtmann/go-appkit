# W5-C2 Design Note — the projection→broadcast folded contract

**Date:** 2026-11-07 (SUPERB v14 C5, subtask 20.1) · **Status:** DESIGN
NOTE — the contract implementation (20.2) rides the next cqrs+realtime
train through published tags. This note is the plan-of-record.

## 1. The race

Every future cqrs+realtime consumer hits the same timeline bug if they
wire "journal appended → broadcast a snapshot" naively:

```
t0  journal appends UserRenamed{...}
t1  consumer notifies the hub (broadcast)          ← too early
t2  projection host calls Handle(evt) — the fold starts
t3  Handle returns — the fold is now consistent
```

Between t1 and t3 a subscriber receives PRE-fold state. Worse: no
follow-up broadcast exists — the event that caused the change is never
re-announced, so the client shows stale data until the NEXT unrelated
event. CV hunted exactly this class for four sessions (2026-08-16 flake
hunt) and fixed it at the ordering level: emit only after `Handle(evt)`
returned (`SubscribeFolded`).

## 2. The contract

**Broadcast-strictly-after-fold:** a notification that announces a
projection's state may only be emitted after the fold that incorporates
its trigger event has completed. The payload must reflect the fold that
the trigger caused (or a later one — folds are monotonic; an EARLIER
payload is the violation).

Two surfaces must agree:

- **cqrs side** — the projection wrapper that emits the folded signal.
  The engine's `projection/v4.Projection` is
  `{Name, Handle(ctx, evt), EventTypes}`; a wrapper records the completed
  event BEFORE returning from Handle, so "fold done for evt E" is
  observable exactly when Handle(E) has returned:

  ```go
  // cqrs (sketch, next cqrs train):
  func NotifyFolded(p projection.Projection, folded chan<- FoldedEvent) projection.Projection
  // Handle wraps p.Handle: on success, folded <- FoldedEvent{...} BEFORE
  // returning (a non-blocking send with a bounded buffer: a slow hub must
  // never stall the fold; backpressure is the hub's job, not the fold's).
  ```

- **realtime side** — the broadcaster that turns folded signals into
  hub broadcasts with the snapshot pulled AFTER the signal (so the payload
  includes the trigger's effects):

  ```go
  // realtime (sketch, next realtime train):
  func FoldedSource(folded <-chan FoldedEvent, snapshot func() any, hub *Hub)
  // one goroutine: for each folded event, snapshot() THEN hub.Broadcast —
  // ordering owned here so consumers cannot get it wrong (the spec's
  // FoldedSource shape, adapted to the engine's Projection interface).
  ```

The battery-spec sketch (`realtime.FoldedSource(projection.SubscribeFolded,
...)`) assumed a `SubscribeFolded` upstream; the engine exposes none —
the cqrs wrapper above IS the SubscribeFolded equivalent, and both halves
live in appkit modules so the composition is appkit's to own.

## 3. Ordering hazards the implementation must pin

1. **Snapshot-after-signal, not before:** `FoldedSource` must call
   `snapshot()` after receiving the folded signal, never concurrently
   with the fold. (The signal IS the happens-after edge.)
2. **Slow-hub backpressure:** the folded channel is bounded (suggested:
   64); the wrapper's send is `select`-with-default (drop + a counted
   drop counter — W5-C1's `WithOnDrop` work dedupes into the same
   counter). A dropped folded notification only delays re-sync: the next
   broadcast carries full-state payloads (the W5 invariant).
3. **Multiple projections, one hub:** each projection gets its own folded
   channel + FoldedSource; the hub serializes broadcast order, and
   per-projection ordering is guaranteed by the single FoldedSource
   goroutine per channel.
4. **Replay/rebuild folds:** a projection reset (host rebuild) folds many
   events; the wrapper emits once per Handle — bursts are expected, the
   bounded buffer absorbs them, drops are the documented behavior.

## 4. Acceptance (the discriminator, ported from CV)

`TestProjectionIntegration_SubscribeFoldedReflectsEvent` (cqrs+realtime
integration, E2E through published tags per the pin charter):

- A wrapper store appends a marker MID-FOLD (inside the consumer's
  Handle, before it returns).
- The FoldedSource broadcast's snapshot payload must INCLUDE the marker —
  proving the snapshot was taken after Handle returned (a pre-fold
  broadcast would miss it; that is the discrimination).
- Negative control: a naive notify-before-fold wiring must FAIL the same
  assertion (the test only discriminates if the wrong wiring fails it).

> **ERRATUM 2026-11-08 (control mechanism superseded):** the originally
> envisioned control held the fold open with a BLOCKING gate inside Handle.
> That mechanism is unimplementable on the live engine: watermill's
> `GoChannel.Publish` is synchronous — it waits for the projection worker
> to take the message — so an indefinitely-blocking `Handle`
> back-pressures the entire command dispatch path (proven by goroutine
> dump during the 2026-11-08 integration bring-up; `Dispatch` itself
> deadlocked). The shipped control
> (`integration/folded_e2e_test.go`, `TestFoldedContract_PreFold…`) keeps
> the same discrimination with a NON-BLOCKING deferred marker: the naive
> wrapper signals pre-fold, then defers its marker write until the test
> has observed the stale broadcast (plus a 2s hard fallback so the fold
> can never strand the engine). Determinism is preserved because
> FoldedSource snapshots at signal receipt and the asserted SSE payload is
> that frozen copy — the ordering chain snapshot → broadcast → read →
> marker-write is total. The positive test is unchanged in spirit: the
> payload of the broadcast for event E must contain the marker written by
> E's fold.

## 5. Train plan (20.2)

1. `cqrs` next train: `NotifyFolded` + `FoldedEvent` + the wrapper's
   non-blocking send + drop counter surface (additions-only).
2. `realtime` next train: `FoldedSource` + the full-state re-sync
   invariant doc + the C1 drop-counter dedupe decision recorded.
3. `integration` re-pin when both tags are on origin; the discriminator
   E2E lands there (published tags only, per the charter).
