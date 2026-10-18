# SUPERB v9 Post-Train Hardening — Status

**Date:** 2026-10-18 22:33-23:20 CEST · **Plan:** `docs/planning/2026-10-18_22-33_SUPERB-v9-post-train-hardening.md` · **Trigger:** user-directed continuation; attacks the four gaps the 09:47 self-review exposed (trust, standing gates, measurement honesty, small rot).

## Delivered

1. **Trust closure** — `example-smoke.sh` run LOCALLY: EXIT=0 (the only CI job without a local equivalent — gap closed). CI on the final tip `56e41c0` (run 39163825323): fast guards green (go-directives, workspace-charter, root proxy-smoke), ci-preflight red = designed, structure-linter skipped = secret; the bulk was STILL QUEUED at session end — first fully-lit verdict must be checked next session (noted in TODO header).
2. **art-dupl standing gate (the 4%→64% item)** — acceptance converted from prose to a gate, go-etag precedent generalized:
   - `docs/reports/dupl/art-dupl-baseline.json` — committed hash baseline (native `art-dupl baseline`), + human snapshot `2026-10-18_gate-settings-snapshot.txt` (640 groups detected, **9 actionable at gate sensitivity** vs 632 at audit sensitivity).
   - `scripts/check-art-dupl.sh` — native `art-dupl check . --type-aware -t 1` vs the baseline; loud install/regenerate hints; output captured, never piped.
   - **Born-red proven**: planted semantic (renamed-identifier) duplicate → exit 1 listing both sites; removed → exit 0. Also proven: the subcommands accept NO `--test-threshold` (test files below 30 tokens are gate-out-of-scope by design — the periodic `--test-threshold 1` manual audit remains the deeper pass; first plant attempt with a verbatim `_test.go` copy was invisible for exactly this reason, documented in the guard header).
   - Wired as the **eighth** `pre-tag-checks.sh` guard (header updated).
3. **`verify-all.sh --quiet[=LOAD]`** — self-gating re-exec through `scripts/quiet-run.sh` (default load-max 64 = the 17-49 §e2 threshold, waits up to 1h, evidence to `$QUIET_RUN_LOG`); tested end-to-end on the docs module (gate opened at load1 30.87 → PASS test+lint → DONE line). Non-quiet behavior untouched.
4. **Recovery-bench instability ROOT-CAUSED** (three probes, n=10 each, evidence in /tmp/recovery-p{1,2,3}.txt):
   - P1 GOMAXPROCS=1 + 1 primary: happy ±14%, panic ±21%, RecordError ±39% — WORSE (single-P scheduling noise).
   - P2 GOMAXPROCS=4 + 4 primaries: happy 18.70µs ±4%, panic 40.43µs ±3%, RecordError 5.32µs ±7% — **within-run stable**.
   - P3 = identical repeat of P2: 16.39µs ±13% / 33.99µs ±5% / 4.55µs ±3% — **cross-run deltas 12-16%, inside the ±25% bar** but not tight enough to re-baseline absolutes.
   - Verdict: the 8-core pins were the morning's noise; 2026-10-12 table values STAY; re-baseline bar = the 4-core protocol repeating at cross-run <±10%. README note upgraded from "suspected frequency-scaling" to the quantified protocol; otel CHANGELOG `[Unreleased]` extended.
5. **Docs/rot** — module-details gained the `otel_composition_e2e_test.go`/`otel_testharness_test.go` row (the v8 E2E file was missing entirely); watches re-checked (httputil v1.5.0, otel upstream v1.47.0, go-health#7 / fr#2 / httputil#41 all OPEN — unchanged); buildcache read-only profile: **go-build 64G** (gopath du times out under load; fs 85% full, 33G free) — cleanup is an OWNER call on a shared box, report only.
6. **Formatter-damage recurrence recorded** — the 13:20 dprint pass (daemon pre-commit BuildFlow) re-wrapped the morning's otel README paragraphs AND re-doubled a tilde (`~5.7µs` → `~~5.7µs`, markdown strikethrough corruption; same class as the pre-existing `~~3.3us` this morning). Fixed on sight; if the next dprint pass re-doubles it, that is a REPRODUCIBLE BuildFlow/dprint markdown bug worth an upstream issue (premise: observe one more recurrence).

## Verification

Offline guard set green post-changes (directives, dependabot parity, charter, sse-protocol) + `bash -n` on all touched scripts + `check-art-dupl.sh` green on the final tree. Integration/docs suites unaffected by this wave (no Go source changed). Committed and pushed per user instruction (see git log).

## Carried (gated — unchanged from 09:47 report §c/§f)

Owner: SSH secret (P1), pma §6 + ignore edit. User: Q1/Q2/Q3 (asked 09:47, still unanswered). Demand: W3-W5, E2E gaps, opt-ins, OTLP logs, auditlog. Upstream/event: httputil#41, go-health#7, fr#2, i18n tag, otel v1.48, SigNoz watches, cordis, core v1.0. Next-session-first: fully-lit CI verdict on `56e41c0` (or the pushed v9 tip).
