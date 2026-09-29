# Status Report — Queue Sweep: BuildFlow Verdict, Extraction-Safe go-etag, T20/T21/T29-T32 Closure

**Date:** 2026-09-29 ~12:45 CEST
**Scope of this report:** the single session that executed the handoff
queue left by the 11:47 report — BuildFlow verdict loop, #39/#40, T29
write-back, T30–T32, T20 batch 1, T21 audit, upstream drafts, small
items (#37/#43/#45/f20/f21/f31), and a reproduced-and-fixed testkit
flake. Point-in-time snapshot — it goes stale.
**Verification baseline (fresh at write time):** full workspace
`go test ./... -race` GREEN across all 11 modules; all four guards green
(pin-drift, go-directives, dependabot-parity, go-structure-linter);
golangci-lint 0 issues on every module touched (otel, cqrs, security,
root/testkit); tree clean at `7ff0b36`.

---

## a) FULLY DONE

1. **BuildFlow full-mode verdict loop (the open hazard).** The run had
   deadlocked 1h37m on networked children (`lychee` needs egress;
   `shellcheck` cannot take 1.5h on 4 scripts) — D-state, SIGTERM-proof,
   killed with `pkill -9 -x`. Its otel mutations (proto getters,
   `t.Context()`, `slog.DiscardHandler`, nolint cleanup, err/if splits)
   were reviewed hunk-by-hunk, verified race-green + lint-0, and
   committed (73353f8 + daemon sweeps). Full-tree audit vs the report
   commit confirmed otel was the only code target; AGENTS/TODO deltas
   were the concurrent SigNoz session's same-train docs.
2. **#39 — go.work etag replace dropped, PROPERLY.** First removal broke
   the workspace root build with an ambiguous import: integration's
   pinned errorpages v0.1.0 / otel v0.1.1 pull pre-extraction go-etag
   root (v0.1.1/v0.3.1) whose `server/` dir collides with the extracted
   `go-etag/server` module. The replace was LOAD-BEARING. Fix: extend
   integration's exclude block past the extraction boundary (v0.1.1,
   v0.3.1) so plain MVS resolves v0.6.0 in hermetic AND workspace mode
   (e0db033). The `go work sync` side effects (frh go-health v0.2.0→v0.4.1,
   realtime ssetest v0.2.0→v0.4.0 silent bumps of deliberate pins) were
   REVERTED — member go.mod changes ride release trains only.
3. **#40 — security follow-through.** Example-only appkit dep → v0.7.0
   (race-green, lint-0); CHANGELOG `[Unreleased]` entry added; the v0.2.0
   entry corrected with an erratum (it claimed v0.5.1 while the tagged
   artifact silently carried v0.6.0) (4cee4f4).
4. **T29 write-back.** cordis `go/v0.1.0` re-verified on the fork;
   AGENTS Deferred Register row annotated (trigger 1/3 met);
   TODO watchlist as-of dates refreshed (PapDashboard v0.3.0, cqrs-htmx
   v4.12.0, battery demand re-check). Probe results: installed buildflow
   = e881e96; buildflow is NOT in nixpkgs; `nix eval nixpkgs#go.version`
   is UNANSWERABLE in this environment (times out >30s repeatedly —
   matches the hung probes from the previous session).
5. **T30 — Decision 14:** external consumers verified on-touch +
   quarterly sweep, never per-release (9066b67).
6. **T32 — graduation table** in `core-v1-exit-criteria.md`: 4/7 hard
   criteria met at v0.7.0; consumer proof (criterion 5) is the
   structural blocker; freeze window opens at the NEXT core tag. cqrs
   cookbook trigger wording verified consistent in TODO + AGENTS.
7. **FEATURES.md otel split-brain closed** — v0.2.0 surface rows added
   (dashboard row honestly marked "unvalidated against live SigNoz").
8. **T20 batch 1 — AGENTS slimmed 377 → 342 lines.** Every module README
   now carries a Build & verify section (10 appends + 3 new compact
   READMEs for docs/flightrecorder/integration, which had none); AGENTS
   keeps the universal hermetic pattern + the lint-from-module-dir /
   never-concurrent rules; root README's Development section states the
   1.27.1 floor (2893665).
9. **f31 — CI proxy-smoke skip ROOT-CAUSED and fixed.** Not mysterious:
   `needs: test` + the matrix dying at the owner-gated ssh-agent step ⇒
   dependent skipped 0s. The job verifies the PROXY, not the tree —
   decoupled (d486532). Stale `GOEXPERIMENT=jsonv2` workflow env dropped.
10. **T21 — the promised art-dupl audit pass.** Full-visibility scan
    (corrected invocation, no `--type-aware`): 477 groups, ALL in
    _test.go files; 349 single-statement noise; ONE harmful clone found:
    the 20× `NewEventService(EventConfig{DSN: t.TempDir()...})` +
    err-check + defer-Shutdown preamble across six cqrs test files —
    collapsed onto `newTestEventService(t, EventConfig)` (t.Cleanup
    preserves defer ordering; DSN-defaulting keeps explicit-DSN sites
    expressible), net −132 lines. Memory-driver/error-path/explicit-DSN
    variants accepted with rationale (config IS the test's subject).
    Outcome recorded in Decision 13 (3cf0d8a, b50c72c).
11. **#37 + Decision 13 amendment:** per-module `.golangci.yml`
    duplication accepted BY DESIGN (standalone-correct configs; generator
    rejected); art-dupl runs event-driven from here.
12. **#43 /tmp sweep:** 16 session-residue files trashed (never rm);
    systemd-private dirs and other projects' scratch untouched.
13. **#45 dependabot eyeball:** 11 gomod dirs + github-actions, grouped
    minor+patch weekly, limit 5 — parity guard green.
14. **Upstream drafts (gated):** go-licenses toolchain-awareness ask
    drafted (f2d94c7). The go-structure-linter draft was NOT needed —
    see (d1).

## b) PARTIALLY DONE

1. **f20/f21 AGENTS note** — done as part of the BuildFlow gotcha bullet
   (full-mode mutates, verdict loop, deadlock mechanics, go-work-sync
   ban, etag-exclusion pattern). NOT separately double-checked against
   the lessons.md wording; both exist and agree in substance.
2. **#49 pkg.go.dev re-check:** flightrecorder v0.1.1 and docs v0.3.1
   STILL 404 at ~12:20 (crawler lag, ~12h post-tag; realtime/health/core
   render). Re-check next session; escalate only if still 404 at 48h.

## c) NOT STARTED (deliberate)

1. **otel v0.2.0 release train** — code-complete by the concurrent
   SigNoz session; its ritual step 1 (API-break diff) unrun; push is
   user-gated and the owning session is waiting. NOT mine to start.
2. **T20 batches 2–3** (AGENTS gotcha-group extraction) — batch 1
   already brought the file to 342/377; the remaining groups are
   higher-risk to move (module gotchas reference each other).
3. **Gated set unchanged:** T13 (G3), T15 (upstream filings — the two
   drafts wait), T16, T22, T23, T27, T28, T10.1, battery waves W3–W5,
   SigNoz validation (needs the user's instance).

## d) TOTALLY FUCKED UP (radical honesty)

1. **I nearly filed an obsolete upstream draft.** AGENTS' Deferred
   Register said the structure-linter yaml `exclude_patterns` was inert
   in binary f7e33e03 — but the INSTALLED binary is aaca8e20 and HONORS
   the config now (proven: bare run 11 findings → 0 with the config).
   One verification command stood between me and a wrong-by-default
   upstream issue. Row CLOSED instead (b696a08); the "wait for a newer
   binary" escape path had simply succeeded.
2. **`go work sync` ran before I understood what it back-propagates.**
   It silently bumped deliberate per-module pins inside released
   artifacts (frh, realtime). Caught by reviewing the daemon-committed
   diff, reverted — but the correct order was: read what sync does,
   predict the member deltas, THEN run it (or not).
3. **The first #39 removal was based on the handoff's "redundant"
   theory and broke the workspace root build.** The replace was
   load-bearing across a package-extraction boundary. Recovered with
   the proper fix (excludes in integration), but the handoff's premise
   deserved a workspace-build test BEFORE the edit.
4. **Three blind helper-refactor compile cycles in cqrs** (missing
   testing import, `err` redeclarations, one-level-shallow literal
   indent) — the count-asserted mechanical replace was fine, the
   aftercare wasn't batched with a compile between steps. Cost: ~3
   wasted cycles, zero shipped breakage.
5. **Pre-existing and still standing:** master CI dead on the ssh gate
   (T10.1, owner-gated); the golangci-lint LSP diagnostic on
   testkit.go:152 is stale (real linter: 0 issues).

## e) TOP NEXT ACTIONS (for the next session)

1. **Run the otel v0.2.0 release train** (or tell the SigNoz session to)
   — ritual step 1 diff → CHANGELOG date → tag → gate → push → proxy →
   integration re-pin. Highest-value unblocked release.
2. **Re-check pkg.go.dev** for flightrecorder v0.1.1 + docs v0.3.1; if
   still 404 at 48h post-tag, investigate (re-tag ritual exists for the
   docs-ghost class).
3. **Fix FEATURES/docs-health ritual gap** (SigNoz session's f3): add
   FEATURES.md to the same-train release-state update list in the
   release ritual (AGENTS step 3).
4. **Answer the 3+3 user questions** (mine: machine toolchain ≥1.27.1?,
   G3 posture?, cordis promotion?; SigNoz session's: live SigNoz?,
   release now or batch?, OTLP logs now or wait?).
5. **T20 batches 2–3** when the next train needs AGENTS headroom.

## f) THE QUESTIONS I CANNOT ANSWER MYSELF (blockers)

1. **Machine toolchain:** upgrade the default `go` to ≥1.27.1? Every
   command currently needs the `GOTOOLCHAIN=go1.27.1` prefix; the
   unification is done, the prefix tax is per-command forever otherwise.
2. **G3 (consumer-claims guard):** keep consumer pins + build
   `scripts/check-consumer-claims.sh`, or strip them? Blocks T13.
3. **cordis bridge:** promote to a tracked TODO now that trigger 1/3 is
   lit, or keep cold? (Watchlist wording updated today either way.)
4. **otel v0.2.0 push authorization** (the SigNoz session's g2, still
   open): release train now or batch?

---

_Point-in-time snapshot. All claims above verified against the tree at
7ff0b36 unless explicitly marked as lagging (pkg.go.dev crawler)._
