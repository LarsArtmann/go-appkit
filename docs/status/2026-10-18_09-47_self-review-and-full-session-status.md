# Self-Review + Full Session Status — Quiet-Gate Quality Train

**Date:** 2026-10-18 09:47 CEST · **Session window:** 08:00-09:47 · **Trigger:** user-requested full execution of the post-v8 TODO_LIST + this self-review
**Companions:** plan `docs/planning/2026-10-18_08-01_quiet-gate-quality-train.html` · execution report `docs/status/2026-10-18_08-45_quiet-gate-quality-train-status.md` (this file adds the honest a-g pass; the /tmp artifacts referenced there are inlined below so they survive reboot)

## a) FULLY DONE (verified)

| Work | Evidence |
|---|---|
| `scripts/quiet-run.sh` — mechanical quiet gate (load gate + SMT-aware taskset pinning + `--wait` + evidence log) | 3 self-tests green; live-proven twice (bench runs); 4 offline guards + bash -n green |
| Quiet-window full sweep: `verify-all.sh` EXIT=0, 12/12 modules build+vet+test+lint | gate log: load1 13.78 start / 13.58 after |
| Quiet-window cqrs `-race -count=10` — ok 31.5s (**b1 load-starvation caveat CLOSED**) | gate log: load1 13.58 / 12.06 |
| otel ns/op re-baseline (4th attempt, first gated): NoOp 19.67µs ±5%, traced 20.45µs ±13%, metered 25.33µs ±16% — within ±9% of 2026-09-16 | benchstat n=10, load1 11.7, 8 primary threads; README table + derived-costs + CHANGELOG `[Unreleased]` |
| art-dupl `-t 1` max-sensitivity sweep (632 groups): 1 harmful extraction — `newTracedTestService(t)` ×3 sites → `integration/otel_testharness_test.go`; rest accepted per standing posture | integration suite green hermetic 7.7s; lint 0 issues; scoped re-run shows group gone; no non-test groups beyond accepted classes |
| Upstream watch outcomes: go-cqrs-lite#67 CLOSED (row was stale — cqrs v0.8.0 already simplified); BuildFlow#46 CLOSED + binary ef32fa0 verified empirically; dprint exit-14 NON-REPRO → resolved, no filing | gh issue views + worktree repro + `go-work-sync` green with charter guard |
| pkg.go.dev: otel@v0.3.0 now RENDERS; flightrecorder GET re-issued | page fetch 08:14 |
| Open Q5 ANSWERED: buildcache = local ext4 `/dev/sdb1` (autofs), NOT machine-shared | findmnt evidence |
| CI close-out on `dfa0303`: all real jobs green incl. sse-protocol first-origin run | `gh run view 39102243431` |
| Docs truth-sync: TODO_LIST 8 rows closed + header; AGENTS 2 pointers (quiet-run, Deferred-Register strike); status index de-rotted (+5 rows) | file diffs, guards green |

Inlined evidence (survives /tmp):

```
quiet-run gate log (UTC): 06:34:16 GATE OPEN load1=18.44 pin=12,28,9,25,13,24,14,10 (run 1, SMT-buggy pick)
                           06:35:43 DONE exit=0 load1=13.78 | verify-all EXIT=0 | cqrs-10 EXIT=0
run 2 (SMT-aware):         pin=12,9,13,14,10,11,15,8  load1=11.72→15.28
Middleware_NoOp            19.67µ ± 5%   7.659Ki ± 0%   90 ± 0%
Middleware_Traced          20.45µ ± 13% 10.67Ki ± 0%   90 ± 0%
Middleware_TracedAndMetered 25.33µ ± 16% 10.67Ki ± 0%   90 ± 0%
Recovery_HappyPath         27.29µ ± 35%  4.565Ki ± 0%   54 ± 0%   (run 1: 16.35µ ± 36% — cross-run UNSTABLE)
Recovery_PanicPath         40.65µ ± 15% 11.30Ki ± 0%   81 ± 0%   (run 1: 33.45µ ±  2% — cross-run UNSTABLE)
RecordError_Recording       4.712µ ± 23% 3.969Ki ± 0%   12 ± 0%
RecordError_NoSpan          5.597n ± 14%  0 B / 0 allocs
```

## b) PARTIALLY DONE

1. **Session-close ritual** — CI verified green on `dfa0303`, but the session's own commits (ahead 5 at 09:20) await the daemon push; the first-origin CI run on the NEW tip is unconfirmed. Next session must check it (all local CI-equivalents ran green on the exact tree).
2. **art-dupl acceptance is prose, not a gate** — the skill's "gate, not prose" standard (committed `art-dupl baseline` file + `check` CI mode) was NOT created; 632-group acceptance lives in TODO/AGENTS prose only.
3. **quiet-run.sh adoption** — helper shipped + proven, but nothing wires it into `verify-all.sh` (opt-in `--quiet` flag) or CI; the protocol depends on someone remembering it exists.
4. **pkg.go.dev flightrecorder** — trigger GETs re-issued, page still 404; re-tag escalation owner-gated (untouched by design).

## c) NOT STARTED (deliberately — gated rows, unchanged)

P1 SSH secret (owner); pma §6 decisions + `.*-body.html` ignore edit (owner); Q1/Q2/Q3 user decisions; W3-W5 batteries, 3 composition-gap E2Es, subscribe-before-flush, EventConfig opt-ins, OTLP logs, auditlog (demand); httputil#41 compose+TLS, go-health#7, fr#2, i18n residue (next httputil feature tag), otel v1.48 watch, SigNoz watches, cordis, core v1.0.0 criteria (5/7) — all carried in TODO_LIST with unblock conditions.

## d) TOTALLY FUCKED UP (honest ledger)

1. **quiet-run.sh v1 shipped a gate that ALWAYS OPENED** — awk cannot evaluate a variable as an operator (`a op b` is string concat → truthy); `--load-max 0.01` happily started `true`. Caught ONLY because the negative test was run. Lesson: a gate without a born-red negative test is theater. (Fixed; negative test retained as the record.)
2. **First bench run pinned SMT sibling pairs (12/28, 9/25)** → two threads per physical core → bimodal timings, middleware sd ±26-36% — the exact noise class the gate exists to prevent. Caught by reading benchstat's sd column, not by the gate itself. Additionally my FIRST diagnosis said "hybrid perf/E cores"; topology check showed homogeneous Zen + SMT — diagnosis was wrong before it was right.
3. **Wrote "B/op UNCHANGED" while traced B/op had moved 10.84Ki→10.67Ki (~1.6%) cross-session** — imprecise honesty language in the honesty-defining doc. Fixed on sight 09:47 (now: allocs exact; B/op ±0% within run, ~2% across environments). House convention was already loose here; this session tightens it.
4. **The 08:45 status report cites /tmp artifacts as its raw evidence** — /tmp evaporates on reboot; a point-in-time report should inline what matters (done here, §a; the 08:45 file stays as-is per snapshot rules, this file supersedes it for evidence).
5. **Two failed edit rounds on TODO_LIST row 71** (assumed-match on a long segment with mixed unicode arrows/em-dashes; succeeded only with a byte-extracted tail). Minor waste, no damage — lesson: extract-then-edit on unicode-dense rows.

## e) WHAT TO IMPROVE (process, from this session's scars)

1. **Born-red is not just for repo guards** — every gate/helper ships with a negative test that PROVES it can fail (the awk bug would have silently invalidated every "quiet-window" claim after it).
2. **Topology-aware pinning is mandatory for wall-clock numbers on this box** — SMT siblings, not just busy%, decide honesty; quiet-run.sh now encodes it, but any future bench recipe must go THROUGH the helper, never hand-rolled taskset.
3. **Make art-dupl acceptance a committed baseline gate** (`art-dupl baseline` + `check` in CI) instead of re-adjudicating 632 prose groups every sweep.
4. **Inline evidence into status reports** — no /tmp references without inlining the decisive numbers.
5. **verify-all.sh could gain a `--quiet` mode** wrapping itself in quiet-run — one command = gold-standard sweep.
6. **Recovery-bench instability is a real open measurement problem** (16↔27µs across identical-protocol runs) — needs a dedicated investigation (GOMAXPROCS=1 single-core pin, cpufreq governor check, perf counters) before those rows can ever re-baseline.
7. **Unicode-dense rows: extract exact bytes first, then edit** (two wasted rounds today).

## f) NEXT (ranked; ~top 30 of the open queue)

**P1 (correctness/trust):**
1. Verify first-origin CI run on the daemon-pushed session tip (session-close ritual) — next session start.
2. Owner: restore `SSH_PRIVATE_KEY` secret → structure-linter leg runs (the live P1).
3. Re-run `scripts/example-smoke.sh` locally on the session tree (the one CI job with no local equivalent run today).

**P2 (quality):**
4. `art-dupl baseline` committed + `check` wired into CI (acceptance becomes a gate).
5. Recovery-bench stability investigation (single-core pin / governor / perf) — then re-baseline or document the instability's cause.
6. `verify-all.sh --quiet` (self-gating sweep) + mention in AGENTS Build&Verify.
7. Owner: pma §6 four decisions (memo carries recommendations) + `docs/**/.*-body.html` ignore_patterns edit.
8. Watch: httputil next feature tag (Language()/LoggingFromContext) → pin-drift sweep + Language() E2E leg + otel doc-limitation wiring.
9. Watch: otel upstream v1.48.0 release → README cardinality-limit refresh on the next natural train.
10. Watch: go-health#7 / go-flightrecorder#2 / httputil#41.
11. Consumer-claim drift ritual — blocked on Q1 (below).
12. `docs/agents/module-details.md`: add `otel_testharness_test.go` + `scripts/quiet-run.sh` rows (one-line each) if the tables enumerate them.
13. `/mnt/buildcache` at 85% (33G free) — schedule a cleanup pass (oldest buildcache/ + vcs purge per the AGENTS repair recipe).
14. flightrecorder pkg.go.dev: keep weekly page GETs; escalate to the owner-gated docs-ghost re-tag if still 404 at ~2 weeks post-tag.

**P3 (demand/watch — unchanged):**
15-30. W3 httpx module (B1/B2/B9 first); W4 ops kit decisions (D7 adopt-wire go-idempotency vs port); W5 C2 projection→broadcast folded contract (the must-have when a cqrs+realtime consumer appears); composition-gap E2Es security×errorpages, flightrecorder-via-Service, SSE+health drain-during-stream; realtime subscribe-before-flush (demand); cqrs EventConfig opt-in README recipe (cheap path); OTLP logs design dust-gathering check on next claims refresh; SigNoz instance watch; semconv flag watch; PeriodicReader timeout YAGNI watch; cordis trigger 2/3 watch; nixpkgs go floor watch; core v1.0.0 exit criteria (additions-empty tag pair + consumer overlap); pbx-lens testkit migration (their repo); project-dependency-graph polish (their repo); push-lag FATAL escalation (owner, dossier §6).

## g) QUESTIONS (cannot be figured out — owner rulings)

1. **Consumer-claim posture (Q1):** keep exact consumer pins in AGENTS with the re-verify ritual, or drop exact pins for claims-about-consumers? (Gates the drift-ritual row.)
2. **Cross-repo tracking (Q2):** actively nudge cqrs-htmx / pbx-lens items in THEIR repos, or stay report-only?
3. **flightrecorder re-tag escalation:** the pkg.go.dev page is the last never-indexed module (6+ days, trigger GETs exhausted, /fetch 404s fleet-wide). Execute the owner-gated docs-ghost re-tag ritual, or keep waiting?

**Posture:** waiting for instructions.
