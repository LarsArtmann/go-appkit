# v13 Continuation — R1c/R1d Close-Out (bench record, screenshot, dprint)

> **Live remainder:** none — this close-out landed its deliverables in the
> same window (the AGENTS perf re-baseline, the realtime hub.go doc-const
> note, the mobile screenshot, the dprint coverage record) and carries NO §f
> section; the program rows it references live on the v13 plan + TODO_LIST.
> Archived 2026-11-06 as a zero-item unit (the cadence wave's free archive).
>

**Written:** 2026-11-06 13:10 +0100 · **Session window:** ~13:00–13:10 (continuation of the 12:46 session, executing its exact next steps)

## Context

Resumed on owner instruction ("execute and verify, step by step") after the 12:46 WAIT. The
12:46 report's remainders in MY lane were: R1c bench write-up (data already measured),
R1d (hub.go note + mobile screenshot + dprint), TODO row 35 close. P9c/P8 were marked
sibling-conditional.

## Done (all verified)

1. **R1c — AGENTS perf section re-baselined:** the Server-Timing cost line now carries the
   quiet-run numbers (~14.7-16.6µs/op, 6165 B/op, 71 allocs, `-benchtime 1s -count 6`,
   pinned 4 cores, load ~14) and records the 2026-11-03 47-56µs reading as
   load-contaminated, with the quiet-run-only bench discipline stated. The `total`
   assertion half of R1c needed NO new code: `middleware_test.go` already pins it three
   ways (default-stack header, `ServerTiming` knob on/off, replaced-stack absence).
2. **R1d — hub.go doc-const note:** `realtime/hub.go` `WithBufferSize` now states the 64
   default is go-sse v0.6.2's per-subscriber default, UNEXPORTED upstream (prose-only
   there), with the swap-to-`DefaultSubscriberBuffer` instruction for the next re-pin.
   Realtime suite (-race), build, vet, lint: all green.
3. **R1d — mobile screenshot:** `docs/review/libs/index-mobile-390x844.png` (chromium
   headless, 390×844, `--virtual-time-budget`) — first image asset in-repo; verified
   non-blank mechanically (stdev 0.35, 1098 colors). No repo image convention existed;
   beside-the-HTML is the established pattern now.
4. **R1d — dprint:** check was REAL-RED (exit 20, 8 files — the sibling session's
   12:48-13:01 doc commits landed unformatted, daemon bypasses hooks). `dprint fmt` +
   `check` now green repo-wide. (Also caught the exit-code-pipe trap once more — `$?`
   after `| tail` lied exit=0; unpiped rerun showed 20. Never pipe guard output.)
5. **TODO row 35 CLOSED** (v11 repair backlog — all R1a-e halves landed, evidence inline).
6. **Guards:** all five fatal pre-commit guards green on the touched tree
   (check-agents-deps verbose-verified — my AGENTS edit was its domain).

## Left open, deliberately

- **P8 (integration E2E via preset/v0.1.0) + P9c (proxy checks v0.10.0/config/preset):**
  the sibling session is ACTIVELY executing this lane — their integration preset pin was
  in-flight (uncommitted) at 13:05 and the cross-app-audit report-archive rename at 13:06;
  their 12:49 report §f lists P8 + proxy checks as their next items. Racing them on the
  same files repeats the 08:45 collision class. TAKE-OVER CONDITION: no sibling tree
  movement for ~15 min AND no preset_e2e_test.go in integration/.
- **Owner questions still outstanding** (from the 12:46 report): H8 httputil/endpoints
  seam decision, OG6 machine-go-default ≥1.27.2 (retire the GOTOOLCHAIN prefix?), session
  lane contract. H-train/X-train/S2-S7 stay gated on them.

## Verify commands (as run)

```bash
cd realtime && GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1   # ok 1.1s
cd realtime && golangci-lint run ./...                                         # 0 issues
bash .githooks/pre-commit                                                      # exit 0 (5 guards)
./scripts/check-agents-deps.sh                                                 # GREEN
dprint check                                                                   # exit 0 after fmt
```
