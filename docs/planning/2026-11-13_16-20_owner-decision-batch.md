# Owner Decision Batch — eight rulings, one ask (SUPERB v16 M4)

**Date:** 2026-11-13 16:20 CET · **Purpose:** every open owner gate in one document. Each ruling unblocks a bounded task; answer inline (a/b/c) or by reference. Nothing here blocks a release — all are quality/roadmap decisions.

## D1 — D014 wire format: `FoldedEvent.occurredAt` json tag (TODO P2)

`FoldedEvent.occurredAt` shipped in cqrs v0.9.0 without a json tag, so JSON encoding uses the Go field name (`occurredAt`). Consumers may already rely on either spelling.

- **a) Tag it** — add `json:"occurredAt"`: a no-op for the current spelling, locks the contract deliberately, clears D014. Wire format unchanged.
- **b) Suppress with reason** — records intent, keeps the field name Go-idiomatic if it ever renames.

**Recommendation: a** — the folded contract is a WIRE contract (cqrs ↔ realtime cross-module); locking the spelling beats leaving it accidental. 30-minute task (M8).

## D2 — art-dupl gate policy: new-groups-only vs exact-count (TODO P2)

`check-art-dupl.sh` fails on NEW clone groups; accepted groups silently disappearing (or rotting) go unnoticed. go-etag's precedent is exact-count strictness.

- **a) Stay permissive** — zero re-baseline friction on refactors.
- **b) Exact-count/stale-entry-failing** — catches accepted-group drift; costs a baseline touch on every refactor that legitimately changes a group.

**Recommendation: b, but warn-first for one cycle** (a soft-fail run to size the friction before making it fatal). Cadence: schedule the deep audit (`art-dupl . --type-aware -t 1 --test-threshold 1 --rich-text`) quarterly alongside the fleet-audit review.

## D3 — H8 route seam (`docs/planning/2026-11-07_h8-route-seam-owner-ask.md`)

Unblocks httpx B6 (rate-limit battery) + B8 (auth battery) and the `svc.Routes()` answer. Full memo exists — ruling needed.

## D4 — auth-bridge (`docs/planning/2026-11-07_auth-bridge-owner-decision-memo.md`)

usermgmt-on-appkit recipe vs a security-module session battery; recommendation there = staged. Blocks preset ever growing an auth leg.

## D5 — plan-revision convention (SUPERB v12 micro-backlog remainder)

When an executed plan gains a revision, is the revision an EDIT of the plan file, an addendum section, or a dated successor file? TODO rows currently reference both styles.

**Recommendation: dated successor file for scope changes, inline addendum for errata** (matches the errata pattern already in the corpus).

## D6 — planning-docs "rev N (expanded)" convention (audit loose ends)

Naming convention for revised planning docs (suffix vs new file). Same class as D5 — could be folded into one naming ruling.

**Recommendation: fold into D5** — one convention answer covers both.

## D7 — watermill synchronous-publish upstream ask (engine-lesson pair, TODO P2)

The gated-fold back-pressure discovery (watermill `GoChannel.Publish` is synchronous; an indefinitely-blocking projection `Handle` deadlocks dispatch) is drafted as an upstream ask vs recorded appkit-side only.

- **a) File upstream** (draft per verify-before-filing; drafted in M7).
- **b) Appkit/cqrs doc caveat only** (the design-note §4 erratum already carries the in-repo record).

**Recommendation: a** — the failure class hits every go-cqrs-lite consumer, not just appkit; the draft is written and gated.

## D8 — pma daemon dossier §6 (four decisions, `docs/planning/2026-10-13_pma-owner-decision-memo.md`)

Hook-bypass KEEP (recommended), train-repo cooldown, staged-message respect, go.work-mutator cadence acceptance. All four have recommendations in the memo; rulings retire the TODO row.

---

**Answer format:** `D1: a, D2: b (warn-first), D3: see memo, ...` — any unanswered ruling stays parked on its TODO row with its current recommendation.
