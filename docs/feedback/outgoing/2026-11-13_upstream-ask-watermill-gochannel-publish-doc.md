# Upstream draft: watermill gochannel — the `Publish` doc comment contradicts `BlockPublishUntilSubscriberAck`

**Status:** DRAFT — filing is USER-gated (premise verified at source, watermill v1.5.3 `pubsub/gochannel/pubsub.go`; see Evidence). Do not file without explicit approval.
**Target repo:** github.com/ThreeDotsLabs/watermill (EXTERNAL — full verify-before-filing gates applied)
**Where:** documentation gap (`GoChannel.Publish` doc comment + the `BlockPublishUntilSubscriberAck` option docs)

## The ask

`GoChannel.Publish`'s doc comment says:

> Publish in GoChannel is NOT blocking until all consumers consume.

But when `BlockPublishUntilSubscriberAck` is enabled (go-cqrs-lite's bus
default, and a documented pattern for ordered processing), `Publish` DOES
block until every subscriber ACKs — and an ACK fires only after the
subscriber's `Handle` returns. `Publish` therefore blocks until all
consumers **finish handling**, which is exactly what the comment says it
never does. The same comment appears to be the source of recurring
surprise in consumers that enable the option for ordering guarantees.

Ask: correct the comment to state both modes, e.g.

> Publish in GoChannel is NOT blocking until all consumers consume,
> **unless `BlockPublishUntilSubscriberAck` is set — then Publish blocks
> until every subscriber ACKs (i.e. until its `Handle` returns)**.

## Evidence

- v1.5.3 `pubsub/gochannel/pubsub.go:86`: the comment above `Publish`.
- Same file :132-134: `if g.config.BlockPublishUntilSubscriberAck { g.waitForAckFromSubscribers(msg, ackedBySubscribers) }` — the blocking branch inside `Publish`.
- Same file :152+: `sendMessage` fans out via a goroutine and returns the ack channel, so without the option the non-blocking claim holds.
- Consumer impact (verified reproduction path, not filed): a subscriber whose `Handle` waits on a condition that only the publishing goroutine can satisfy (a gate opened post-`Publish`) deadlocks `Publish` — surfaced in the go-cqrs-lite bus (whose default is `BlockPublishUntilSubscriberAck=true`, enforced by its lint rule E014), where a gated-fold projection deadlocked command dispatch; root-caused via goroutine dump 2026-11-08 in go-appkit's cqrs module.

## Why upstream

The comment is the first thing a consumer reads; with the option enabled it
describes the opposite of the actual semantics. The fix is two sentences in
one doc comment.

## Verify-before-filing checklist

- Source read, not guessed (v1.5.3, pinned above; commit paths cited).
- Not already documented: the option's field comment says what it does, but
  the `Publish` comment — the surface consumers read first — still carries
  the unconditional NOT-blocking claim.
- Would the proposed change fix it: yes — one doc-comment edit.
- Existing alternatives checked: filing against go-cqrs-lite covers only the
  consumer side; the misleading sentence lives upstream.
- Annoyance test: two-sentence docs fix, real deadlock-class confusion
  behind it.
