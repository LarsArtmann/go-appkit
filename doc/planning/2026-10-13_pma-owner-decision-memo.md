# pma Owner Decision Memo — dossier §6, four decisions

**2026-10-13 · Executor: Crush (full-execution session) · Input:**
[`docs/status/2026-10-07_11-00_pma-daemon-root-cause-and-dossier.md`](../status/2026-10-07_11-00_pma-daemon-root-cause-and-dossier.md) §6
+ one new day of live evidence (2026-10-12/13: three attribution-erasing
interleaves, one commit-failed race, one pma restart).

---

## New evidence since the dossier (2026-10-13 pass)

1. **`journalctl -t pma` does NOT name files.** The dossier's §6.4 hope —
   that pma's journal "will show pma's file-change event stream and may name
   the process that touched go.work first" — is refuted: the journal logs
   batch counts (`events=1 files=1`), eviction ticks, and commit outcomes,
   but ZERO per-file lines (0 `go.work` mentions across 28.5k lines since
   2026-10-12 09:00). The go.work mutator (leaks #1–7, gopls-class) stays
   unnamed; the remaining paths are a machine-level inotify audit or
   accepting the cadence with the charter guard's `--fix`.
2. **Attribution races are routine, not rare.** Three this session alone:
   the cqrs race-fix (landed as `chore: auto-commit 1 changed file(s)` with
   no what/why — the commit that closed a real `-race` find), two
   annotation batches. Plus one `ERROR commit failed` (12:51:02, racing a
   staged batch). pma restarted mid-day (pid 528082 → 2187, ~13:45) without
   visible behavior change.
3. **The heuristic commits are the ONLY history for most daemon-landed
   work.** Every "what/why" for those batches lives in session reports —
   the dossier's "report doubles as manifest" pattern is now load-bearing.

## The four decisions (dossier §6), with recommendations

| # | Decision | Recommendation | Why |
| - | -------- | -------------- | --- |
| 1 | Hook bypass: keep `core.hooksPath=/var/empty` (hermetic committer) vs opt-in per-repo hooks mode | **Keep bypass; do NOT enable hooks for pma.** | pma commits intermediates that the offline guards (directive drift, pin parity) would reject — a hooked daemon fails to commit at all (worse than lost attribution: lost work-in-progress batches, the 12:51 ERROR class). CI-side guards already catch every leak class within one push; keep the net there. |
| 2 | Per-repo exclude/cooldown for train repos like go-appkit | **YES — cooldown, not exclude.** | Trains (tag, gate, push, same-train docs) are exactly when interleaving costs most: staged batches get bundled, tag-vs-tree ordering can race. A per-repo cooldown (e.g. 10 min quiet-window, or skip-while-clean-tree-locked marker file) preserves the safety net for ordinary sessions. Full exclude loses the net for the 95% case. |
| 3 | Attribution: respect staged messages / skip repos with staged changes | **Respect staged messages (minimum), prefer skip-when-staged.** | Today's race-fix is the worst case: a verified, evidence-carrying commit message existed and was replaced by a heuristic one. Minimum: if a batch is STAGED (index non-empty), commit it AS-IS (no message rewrite). Skip-when-staged also kills the batch-eating class (4 staged batches raced on 2026-10-12 alone). |
| 4 | go.work mutator hunt | **Accept the cadence; inotify hunt only if leaks recur post-#46.** | The journal path is dead (evidence 1). BuildFlow#46 closes the go-work-sync mutator (leak #8 class); if gopls-class leaks still recur after it lands, a time-boxed inotify audit is justified — not before. |

Also carried from the fleet session (17-27 §g3): daemon-committed guard
coverage is CI-only today — consistent with recommendation 1; the
`push-lag` guard stays WARN at commit time. Escalating it to FATAL belongs
to CI policy (a red CI job), not the hook.

**Nothing here is implemented by the agent — all four are pma-repo changes
owned by you.** This memo is the decision input; the dossier remains the
evidence base.
