# Quiet-Gate Quality Train — Session Status

**Date:** 2026-10-18 08:00-08:45 · **Trigger:** TODO_LIST post-v8 open rows (user-requested full execution) · **Plan:** `docs/planning/2026-10-18_08-01_quiet-gate-quality-train.html`

## What ran (all evidence in-file)

1. **Quiet-gate helper SHIPPED** — `scripts/quiet-run.sh`: float-safe `--load-max` 1-min-load gate (default nproc/2), `--pin-cores K` taskset pinning to least-busy cores, `--wait` loop, evidence lines to stderr + `$QUIET_RUN_LOG`, exit 2 = gate never opened.
   - Born-red twice: (a) the negative test caught an awk dynamic-operator bug (`a op b` is string concat in awk — the gate always opened; fixed to a fixed `<=` comparison); (b) the first live bench pick included the SMT sibling pairs 12/28 + 9/25 → middleware bench sd blew out to ±26-36%; the picker is now SMT-aware (primary threads only, via `topology/thread_siblings_list`) and the noise collapsed to ±5-16%.
2. **Quiet-window re-runs ALL GREEN** (window: load1 11.7-18.4 on 32 cores): `verify-all.sh` EXIT=0 over all 12 modules (build+vet+test+lint, incl. the deduped integration tree); cqrs `-race -count=10` ok 31.5s. **The b1 load-starvation caveat (2026-10-13 17-49) is CLOSED** — gold-standard quiet confirmation delivered.
3. **otel ns/op re-baselined (4th attempt, first through the gate)** — benchstat n=10, go 1.27.2, 8 primary threads pinned: NoOp 19.67µs ±5% / traced 20.45µs ±13% / traced+metered 25.33µs ±16% — within ±9% of 2026-09-16, so the 1.26.7→1.27.2 jump costs nothing measurable; B/op + allocs/op ±0% again. Recovery happy/panic honestly kept at 2026-10-12 values (cross-run unstable 16.35↔27.29µs / 33.45↔40.65µs under identical protocol). README table + derived-costs paragraph + CHANGELOG `[Unreleased]` updated.
4. **art-dupl `-t 1` max-sensitivity sweep** (632 groups, 204 shown): ONE harmful clone — the 21-line verbatim traced-service setup ×3 (`otel_composition_e2e_test.go` ×2, `otel_pattern_test.go`) — extracted to `integration/otel_testharness_test.go` `newTracedTestService(t)`; suite green hermetic (7.7s), lint 0 issues, scoped re-run confirms the group gone. All other groups adjudicated accepted per the standing clone posture (lo sites produced no new clones; SSE-test churn is assertion-conditionals + with/without-Recovery harness variants — intentional).
5. **Upstream watch batch:** go-cqrs-lite#67 CLOSED COMPLETED → the P3 watch row was STALE (cqrs v0.8.0 already ships bare `GracefulClose`, system v4.14.0) — closed. BuildFlow#46 CLOSED → binary ef32fa0 verified EMPIRICALLY (`go-work-sync` green, go.work untouched, charter green). **dprint exit-14: NON-REPRO on the refreshed binary** (throwaway-worktree repro, CHANGELOG-only staged set exits 0) — row resolved, NO issue filed per verify-before-filing. httputil still v1.5.0 (i18n row stays gated). otel upstream v1.47.0 latest (v1.48 unreleased).
6. **pkg.go.dev:** otel@v0.3.0 now RENDERS (08:14 check; the v0.2.0 showing was crawler lag). flightrecorder still 404 — page GET re-issued as crawl trigger; re-tag escalation stays owner-gated.
7. **Open question #5 ANSWERED:** `/mnt/buildcache` is local ext4 on `/dev/sdb1` (autofs) — NOT network-shared; no cross-machine writers, so the 2026-09-29 corruption class cannot recur cross-machine and the no-concurrent-golangci-lint caveat stays a local multi-process soft constraint. (Disk 85% full, 33G free — cleanup candidate someday.)
8. **CI close-out on `dfa0303`:** ALL real jobs green (12-module matrix, 11 proxy-smokes, guards incl. the NEW `sse-protocol` job's first-origin run, example-smoke, cqrs-lint, config-verify); red only on the designed ci-preflight SSH gate; structure-linter skipped (same secret — the live P1).

## TODO rows closed this session

f4 quiet-gate (helper + sweeps + b1), f10 art-dupl, otel ns/op rows, otel benchstat re-baseline, go-cqrs-lite#67 watch (stale-done), BuildFlow#46 residual (binary verified), dprint exit-14 (non-repro), Q5 (buildcache). AGENTS: load-flake protocol gained the quiet-run pointer; Deferred Register dprint row struck resolved.

## Still gated (owner/user/demand) — unchanged

P1 SSH secret (owner); pma §6 + working-file ignore (owner); consumer-claim posture Q1 / cross-repo Q2 / release-state Q3 (user); W3-W5 batteries, composition-gap E2Es, OTLP logs gate (demand); httputil#41 composition+TLS, i18n residue (next httputil feature tag), otel v1.48 (upstream).

## Raw artifacts

`/tmp/otel-bench.txt` (sibling-pair run), `/tmp/otel-bench2.txt` (SMT-aware baseline run), `/tmp/verify-all-run.txt`, `/tmp/cqrs-count10.txt`, `/tmp/quiet-train-evidence.log` (gate decisions), `/tmp/artdupl.txt` (632-group report), `/tmp/dprint-probe*.txt`, `/tmp/watch-upstream.txt`.
