# Status Report — Docs-Gates Green: 12-Report Corpus Archive + Durability Guards

> **SUPERSEDED 2026-11-06 ~21:30** by the full-session report
> `2026-11-06_21-30_corpus-archive-execution-durability-guards-and-treadmill.md`
> (this interim covered only the morning half; the afternoon added the 21-53
> archive, both durability guards, the pre-tag green, and the CI
> verification). Every §f item here was already routed with fresh:skip at
> write time — nothing was re-adjudicated for the move.

**2026-11-06 ~14:05 CET · Scope:** executing `docs/planning/2026-11-06_09-37_SUPERB-docs-gates-green-and-corpus-archive.md`
wave 1 + the 20% systemic tier, plus the concurrent-session drift repairs that
fell out of it.

## a) FULLY DONE (verified, evidence cited)

1. **`docs-gates run .` is 5/5 PASS** — freshness, audit-all (pin 0), annotate
   (243 findings, pin 255), routing (STRICT, 0 unrouted), cadence (≤15
   unarchived, oldest 4d). The battery was red on cadence (22 unarchived /
   oldest 30d) plus two mid-session freshness regressions when this wave
   started.
2. **Twelve reports archived** with the full ritual (`status-items`/`prose`/`rows`
   specs read in full before apply → `check-rows` → banner → `git mv` → index):
   pma dossier, 01-25, 01-33, 06-59, 07-50, 12-49, 12-25_superb, 07-14,
   12-25_release, 21-35, 16-13, 21-53. Every archived file carries the
   `> **Live remainder:**` banner and verdict-carrying items (~340
   adjudications total).
3. **The mid-session drift repaired truthfully:** the concurrent p9a session
   left TODO rows 37/38/39 (R3/R4a/R5a) open behind completed work — closed
   with verified evidence (`WrapOnce` in service.go, `RegisterClassifications`
   recipe + example call, `hookclassification_test.go`, the servertiming
   README recipe) after `--explain-item` proved the tokens generic.
4. **Real work executed, not just routed:** AGENTS gained the drain-bodies
   testing rule, the spot-check bar, the parallelism-design-time rule, and the
   rows-vs-live-state session-close line; the P8 preset E2E was verified green
   through the PUBLISHED tag (`2de5875`, `integration/preset_e2e_test.go`); the
   dependabot PR-supersession recipe landed as a `dependabot.yml` comment; the
   docs-module corpus ruling is recorded in `docs/CHANGELOG.md [Unreleased]`;
   the preset-spine TODO row now says released-state with the errata pointer.
5. **Fleet govulncheck sweep RUN** (the 16-13 item that had been open since
   2026-10-13): 13 modules, ONE stdlib finding (GO-2026-6809, crypto/x509,
   fixed in go1.27.3), ZERO dependency findings → minted floor-train TODO row
   instead of hijacking a release train mid-docs-wave.
6. **Two new mechanical guards, born-red-proven:** `scripts/docs-annotate-pass.sh`
   (the four-step formatter canon as one command, dry-run default) and
   `scripts/check-status-index.sh` (index→disk + disk→index liveness; wired
   into pre-tag-checks.sh + the CI `status-index` job; it caught two truncated
   manifest rows I wrote minutes earlier — the exact class it exists for).
7. **Index truth restored:** 8 missing Current rows added (the 2026-11-03
   sessions + the 2026-11-06 v12 sessions), 5 stale Current rows removed
   (pointing at files this wave moved), 11 manifest rows written, count
   89 → 101; `check-status-index.sh` green both directions.

## b) PARTIALLY DONE

- **Wave 2/3 headroom:** 21-35, 16-13, 21-53 archived (the plan's wave-2
  list) — cadence sits at 14 unarchived (floor 15). The wave-3 set (the five
  2026-11-03 reports, ~245 items) is untouched; the two worst offenders
  (03-18, 05-03, 06-22 at 48-49 items each) are this wave's leftovers.

## c) NOT STARTED

- Wave 4 (the 11-06 session reports + this report) — gated on the v13 program
  stopping (per the plan).
- `annotate_max` drain (243 tolerated findings, the G/uniform archived class).
- The upstream docs-gates note for the routed≠resolved semantic gap (found at
  source level; the AGENTS gotcha documents it locally).

## d) TOTALLY FUCKED UP (nothing irreversible)

1. **Three freshness regressions of my own making:** my TODO-row verdicts and
   minted rows pushed `backlog`, `sweep`, `mount`, and `preset` across the DF=8
   genericity boundary and broke live harvests — caught each time by
   `--pre-flight` (twice) or the gate (once), fixed by rewording, never by
   threshold edits. The pre-flight-first habit from the plan §2 works; I
   skipped it twice and paid for it.
2. **Truncated manifest filenames** (`archived/2026-11-02_12-25_superb.md`)
   written by my own index script — the new liveness guard caught them before
   commit. The guard paying for itself immediately is the consolation.
3. **One dead routed marker** (`scripts/check-touched-modules.sh` in a marker —
   the guard stats every path-looking string): reworded without the path
   prefix. The guard grammar table in AGENTS now documents this trap.
4. **A daemon race mid-commit:** the auto-committer took 7 of my staged files
   between `git add` and `git commit` — verified content-complete across the
   two commits before continuing (no rebalancing, no force).

## e) VERIFICATION (commands as run)

- `docs-gates run .` → 5/5 PASS (matrix above).
- `./scripts/check-routed-destinations.sh` → exit 0, 32-33 enforced, 0 dead,
  0 unenforced (the Z2-Z7 marker closed to the `TODO P3` form).
- `./scripts/check-status-index.sh` → exit 0, born-red proven (planted stale
  name → exit 1 → restored).
- `dprint fmt` over the 13 touched files → `check-rows` complete on all.
- `docs-gates annotate` → 243 findings ≤ the 244 baseline.

## f) NEXT (impact-ordered; remaining = wave 3 + long tail)

1. Wave-3 archives: the 03-18/05-03/06-22 v12 session reports (~146 items;
   most route to the v13 plan rows this wave already wired). **→ TODO P3 — docs-health corpus waves 3-4 (wave-3 tier) <!-- fresh:skip: routed in TODO_LIST -->**
2. Floor-train go 1.27.3 (GO-2026-6809). **→ TODO P2 — Floor-train go 1.27.2
   → 1.27.3** <!-- fresh:skip: routed in TODO_LIST -->
3. `annotate_max` drain (G/uniform per-cell strikes in the worst archived
   offenders). **→ TODO P3 — docs-health corpus waves 3-4 (the annotate_max drain rides the same passes) <!-- fresh:skip: routed in TODO_LIST -->**
4. Upstream docs-gates note: routed≠resolved between freshness and routing.
   **→ TODO P3 — SUPERB v12 micro-backlog (process lesson to crush-config)** <!--
   fresh:skip: routed in TODO_LIST -->
5. The nine remaining unarchived reports (19-54, 20-35, 21-25, 22-36, 03-18,
   05-03, 06-22, 09-32, 11-47, 12-46, 13-13) ride waves 3-4. **→ TODO P3 — docs-health corpus waves 3-4 (F-62/F-66 tiers) <!-- fresh:skip: routed in TODO_LIST -->**

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Floor-train timing:** bundle the 1.27.3 bump into the next release train
   (the ritual's member-go.mod rule) or run a standalone floor train like
   2026-10-12? The TODO row is written for either.
2. **`audit_max` stays 0** — confirm the pin is permanent (any future
   concurrent-session report instantly reds the battery until that report's
   items are adjudicated; that is the ratchet working, but it means every
   session must close its own §f before push).
3. **Wave-3 scope:** run the five 2026-11-03 reports as one batch (they share
   the v11/v12 program owners already wired) or interleave with the annotate
   drain? Batch is cheaper per item; the plan leaves the order open.
