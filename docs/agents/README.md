# docs/agents/ — agent-only context

**Charter (SUPERB v7 T06, 2026-10-13):** everything under `docs/agents/`
is AI-session context extracted from `AGENTS.md` to satisfy the
≤220-line cap enforced by go-structure-linter and the BuildFlow preflight
(`docs/agents-md-size`). It exists so `AGENTS.md` can stay lean and point
instead of remember.

## Rules

1. **Agent-only, never user documentation.** Nothing in this directory
   gets linked from module READMEs, the repo README, or any user-facing
   surface. READMEs sell and document the modules; this directory is the
   machine-room memory. Module READMEs deliberately do NOT link here
   (decided 2026-10-13, SUPERB v7 T05.4 — audience separation).
2. **`AGENTS.md` points, these files remember.** A pointer line in
   `AGENTS.md` must exist for every file here, and every file here must
   be reachable from an `AGENTS.md` pointer — no orphan agent docs.
3. **Verbatim-grade extraction.** Content moved out of `AGENTS.md` keeps
   its wording and tables (the extraction on 2026-10-12 was verbatim).
   Updates land here when the underlying module changes, in the same
   train, per the AGENTS.md memory-maintenance protocol.
4. **The exemplar is `module-details.md`** — per-module code
   organization, dependency tables, and gotchas. New agent-context files
   follow its shape: dated extraction headers, per-module sections, no
   task lists (those live in `TODO_LIST.md`) and no change history
   (that is CHANGELOG work).
