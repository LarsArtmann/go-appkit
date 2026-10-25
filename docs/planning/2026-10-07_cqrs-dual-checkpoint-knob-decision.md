# Decision: cqrs keeps two checkpoint knobs until v1

> **Status:** Decided 2026-10-07 (pre-v1 review item from the v0.7.0 adoption-closure follow-ups).
> **Scope:** `cqrs` module API surface.

## The question

`EventConfig.CheckpointStore` and `EventConfig.Domain.CheckpointStore` (upstream
`system.DomainConfig.CheckpointStore`) are two knobs for the same concern:
where projection checkpoints persist. Both are the same `event.CheckpointStore`
interface. The merge contract pinned at v0.7.0 says the wrapper-level knob wins
when both are set. Two knobs is a smell — decide now whether v1.0.0 keeps both.

## Decision

**Keep both knobs, wrapper wins, revisit only if a real consumer collides.**

1. The knobs serve disjoint consumer populations:
   - `EventConfig.CheckpointStore` is the ONLY knob for raw-projection
     consumers (no `Domain` config) who need a custom store — e.g. the legacy
     SQL checkpoint table via `eventstore.NewSQLiteCheckpointStore`. Removing
     it breaks them with no replacement path.
   - `Domain.CheckpointStore` is upstream's knob; consumers who think purely in
     `system.DomainConfig` terms (config assembled elsewhere, shared across
     systems) expect it to work. We cannot remove it and should not fight it.
2. The precedence rule is uniform with the rest of the merge contract
   (wrapper/derived wiring wins: `HostOptions`, bootstrap projection, drain
   tracker). One rule, already pinned by tests and documented in the
   `EventConfig.Domain` godoc and README.
3. Nil defaults converge on the same behavior: system's engine-backed
   `system_checkpoints` collection (ADR-0142) — sqlite/postgres persist across
   restart, memory does not. The knobs only diverge when a consumer
   deliberately sets one.

## Risk accepted

A consumer who sets BOTH knobs gets the wrapper's silently. That is the
documented contract, but silent-wins can surprise. Mitigation candidate for the
next minor train (NOT v1-blocking): emit a construction-time WARN when both are
non-nil ("EventConfig.CheckpointStore overrides Domain.CheckpointStore").

## Revisit triggers

- A consumer reports the dual-knob confusion in practice → add the WARN, or
  deprecate `Domain.CheckpointStore` in OUR docs (upstream stays).
- Upstream removes or retypes `DomainConfig.CheckpointStore` → this decision
  collapses to the wrapper knob only; update the merge-contract tests.

## References

- `cqrs/eventservice.go` — `EventConfig.CheckpointStore` and `EventConfig.Domain`
  godoc (merge contract, Timers caveat).
- go-cqrs-lite `system/v4@v4.10.2` `config_types.go:96` — upstream knob semantics.
- ADR-0142 (upstream) — engine-backed checkpoint default.
