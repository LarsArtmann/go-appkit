# Status Report: httputil v1.7.0 Sweep + 2026-10-25 Fix-Wave Close-Out

**Generated:** 2026-10-25 20:10 CET
**Scope:** This session's arc only — the brutal self-review + fix wave, the httputil#43 filing, and the v1.7.0 same-day sweep. No new research was done for this report; everything below was observed or executed in-session.
**Session position:** Continuation of the 2026-10-25 session (started ~15:00 CET). Prior snapshot: `docs/status/2026-10-25_16-13_brutal-review-fix-wave-status.md` (immutable, superseded by this one).

---

## Headline

Upstream shipped **httputil v1.7.0** (tagged by Lars, same day as the filing): the filesystem-`replace` defect I filed as [#43](https://github.com/LarsArtmann/httputil/issues/43) is FIXED (issue CLOSED COMPLETED), and the **`NewServerListener` listener-injection seam** I filed as [#41](https://github.com/LarsArtmann/httputil/issues/41) on 2026-10-13 shipped in the same tag (also CLOSED COMPLETED). Both go-appkit consumers (integration, otel) were swept to v1.7.0 within the hour — all suites, lint, and 5 guards green. **The Service-composition refactor is UNBLOCKED.**

The single most important unresolved fact: **everything — the CI fix wave AND the v1.7.0 sweep — is still local.** Origin's master is (as of this writing) still red on `art-dupl-gate` (born-broken job) and still carries the 6-day-old failures the wave fixes. Owner-gated push remains the standing P1.

---

## Self-Critique (asked directly: what did I forget? what could be better?)

- **Forgot to check upstream issue state early.** My plan said "comment + close #43" — the owner had already closed BOTH #43 and #41 when tagging v1.7.0. A 5-second `gh issue view` at plan time would have trimmed the plan. Found out only at execution.
- **Forgot (structurally, last wave) the first-run proof.** The `art-dupl-gate` CI job I shipped in SUPERB v9 was born broken — the install path was wrong and the check never executed once on origin. The session-close ritual ("verify every NEW job had a live run") exists precisely because of this failure class, and the job predates the ritual's enforcement. The fix is in tree, but the process lesson is: **a new CI job is not DONE until a runner has run it green.**
- **Could be better: batched the two-file multiedit to the wrong file.** I sent the CHANGELOG edit and the README edit in one call targeted at README.md — the CHANGELOG edit correctly failed and had to be re-issued. Cheap, but sloppy; the tool takes one `file_path` per call.
- **Could be better: the final `verify-all` ran WITHOUT the quiet gate.** Load was moderate (~50, not the 200-400 danger zone), and the result is almost certainly sound — but the house protocol exists and I didn't route through `scripts/quiet-run.sh`. Honest gap, same as the 16:13 snapshot.
- **Noticed, did not act (correctly, per your scope instruction):** the integration `documentedPins` fixture carries `cqrs-htmx v4.14.1` while AGENTS.md's 2026-10-12 train text still says "re-pinned to cqrs-htmx v4.14.0" — a later bump happened that AGENTS text doesn't mention. One-line doc-truth fix, parked (see §f #14).

---

## a) FULLY DONE (evidence-backed)

1. **httputil v1.7.0 verified clean from the consumer side** — tag go.mod has NO `replace` line (raw.githubusercontent + local `git show v1.7.0:go.mod` agree); `go doc github.com/larsartmann/httputil@v1.7.0` resolves through proxy.golang.org (the v1.6.0 "contains one or more replace directives" refusal is gone); `NewServerListener`, `Language`, `LoggingFromContext` all present. CHANGELOG diff v1.6.0..v1.7.0 read: #43 explicitly fixed, `NewServerListener` added (#41), `Shutdown`-before-`Start` no-op fix, nothing removed.
2. **integration module swept to httputil v1.7.0** — `go.mod` + `documentedPins` fixture (`integration/pin_drift_test.go:42`) both at v1.7.0; suite `-race -count=1` green in 7.8s (load ~50); `golangci-lint` 0 issues; only stale version ref in the module is a historical comment (otel_pattern_test.go:29, correctly left).
3. **otel module swept to httputil v1.7.0** — `go.mod` at v1.7.0; suite `-race` green (2.1s); vet 0; lint 0 issues.
4. **otel README recipe re-compile-verified at v1.7.0** — scratch module pinned to published core v0.9.0 + otel v0.3.0 + httputil v1.7.0; the "Correlating the request-completion line" snippet compiles (`RECIPE-COMPILE-OK`, probe go.mod confirmed v1.7.0).
5. **otel CHANGELOG `[Unreleased]` truth-synced** — dependency line now v1.5.0 → v1.7.0 (via v1.6.0) with the same-day #43 resolution story; README floor claim updated to "Requires httputil v1.7.0+ (this module's floor; the API itself shipped in v1.6.0)". The stale `doc/feedback` path the handoff flagged was already fixed — verified, nothing to do.
6. **TODO_LIST truth-synced (7 edits)** — #43 row → `[x]` RESOLVED SAME-DAY with consumer evidence; composition row → `[~]` **UNBLOCKED** (seam shipped, remaining work described precisely incl. the single-serve `server.listener_consumed` caveat); P1 push row now enumerates the v1.7.0 re-pin; header close-out sentence; watchlist httputil entries (#41 CLOSED, v1.7.0 tagged); i18n-row replace CAVEAT struck (moot).
7. **AGENTS.md composition bullet updated** — "BLOCKED on an upstream API" → "blocker is GONE (2026-10-25)", remaining work = the refactor itself (spike re-run vs single-serve ownership, re-pin `Addr()`/`Running()` + phase-log contract). File still 133 lines (cap 220).
8. **Outgoing memo → RESOLVED** — `docs/feedback/outgoing/2026-10-25_upstream-ask-httputil-v160-replace.md` status line now carries the full resolution chain (v1.7.0, #43 CLOSED COMPLETED, consumer verification, same-day sweep).
9. **httputil#43 consumer-verification comment posted and verified landed** — [issuecomment-6401869000](https://github.com/LarsArtmann/httputil/issues/43#issuecomment-6401869000); terse comment register per github-voice (checker 0 FAIL / 0 WARN); draft preserved at `docs/drafts/2026-10-25_httputil-43-close-comment.md`.
10. **Guard re-run after all changes: all green** — `check-art-dupl`, `check-go-directives`, `check-workspace-charter`, `check-sse-protocol`, `check-dependabot-parity` all EXIT 0; `verify-all.sh --skip-tests` ALL GREEN (36s).
11. **(Earlier in this same session-day, from the fix wave — re-listed because it is part of this session's done-work):** art-dupl-gate install re-pointed to `cmd/art-dupl` (fresh-cache proxy build proof); `fleet-lint` CI job added (actionlint-validated); `freeAddr` port-race hardening (issued-registry + rebind probe); integration re-pinned (v1.6.0 then, v1.7.0 now); 7 latent cqrs lint findings fixed; `doc/` → `docs/` consolidation completed; brutal-self-review HTML + wave status report written; **httputil#43 filed** after the 5 verify-before-filing gates (including "check master, not just tags" — which is what caught that the premise was still live at filing time).

## b) PARTIALLY DONE

1. **The 2026-10-25 fix wave itself — 100% in-tree, 0% on origin.** What works: every fix verified locally (12-module build/vet/test `-race` green earlier in the day; today's re-verification green again). What remains: the push + tip-CI verify ritual — `fleet-lint` and the fixed `art-dupl-gate` have NEVER had a runner run. Blocker: standing owner-gated-push practice (correct, but it means "done" is not done). Effort to finish: S (one push + ~10 min of CI watching).
2. **The Service-on-httputil.Server composition — unblocked today, not started.** What works: the upstream seam exists (verified API surface), the 2026-09-16 spike verdict already rated the config mapping 6/7 clean. What remains: the actual refactor — spike re-run against v1.7.0 ownership semantics, `Addr()`/`Running()` + shutdown phase-log re-pin, then the swap. Blocker: none technical; it's a scheduling/ambition decision. Effort: L.
3. **Process fix for the designed-red normalization — identified, not implemented.** What works: the problem is precisely diagnosed (one designed red — ci-preflight's SSH-secret gate — trained everyone to read past reds, which is how 6 days of real red went unnoticed). What remains: relocating ci-preflight out of the required workflow into a scheduled watchdog. Blocker: owner decision on workflow protection. Effort: S once approved.
4. **HARVEST of this report's §f — effectively done inline, not formally.** Nearly every §f item already lives in TODO_LIST.md (harvested during the sweep and the wave). The genuinely new items (#4, #14 below) are the only additions a formal HARVEST pass would make. Effort: S.

## c) NOT STARTED (planned, deliberately not begun)

1. **Push + tip-CI verify** — waiting on your explicit push order (standing practice; I will not push unasked). Priority: CRITICAL — it's the only thing between "fixed locally" and "actually fixed."
2. **ci-preflight relocation** — waiting on your decision (§b.3). Priority: HIGH.
3. **Composition spike re-run vs v1.7.0** — now actionable; parked until you rank it against other work. Priority: MEDIUM-HIGH (it also unblocks Core TLS).
4. **Core TLS option on `ServiceConfig`** — gated on the same seam, now unblocked; still trigger-gated in the deferred register on a real consumer requirement (PapDashboard `PAP_TLS_CERT/KEY`). Priority: MEDIUM.
5. **Dependabot wave merge (10 open PRs)** — deliberately sequenced AFTER the wave push (their CI reds were the repo-level guards now fixed in tree). Priority: HIGH, but ordered.
6. **i18n `Language()` composition E2E leg** — demand-gated (nothing pins it); the cheap half (recipe + pin) is done. Priority: LOW.
7. **Testkit helper dedup (4 copy-pasted families)** — blocked on the next natural core release train (integration pins published tags). Priority: MEDIUM, train-sequenced.

## d) TOTALLY FUCKED UP (radical honesty, session scope)

1. **Origin master is STILL red right now.** The fix wave exists only on this machine. Every hour unpushed: (a) `art-dupl-gate` runs red-by-install-failure on every origin push, (b) the 6-day-old real failures stay unfixed on origin, (c) the 10 open dependabot PRs keep reporting reds against a master whose fixes would make them green. Severity: blocks trust in every CI signal; root cause: push is owner-gated (by design) and the order hasn't been given. Mitigation: none available to me — this is the §g Q1.
2. **The designed-red normalization failure is unaddressed on origin.** ci-preflight still sits in the required workflow as a permanently-red check, which is the exact mechanism that let 6 days of real red normalize. Root cause: known (§b.3). Mitigation: known, awaiting your call (§g Q2). Until then, every fresh CI consumer of this repo re-learns "red is normal" — that is actively harmful.
3. **Nothing else.** No data loss, no broken tests, no half-migrated states, no ghost systems introduced this session. The LSP diagnostic noise on `cqrs/scenario_store_pin_test.go` (7 warnings) is the known-untrustworthy `golangci_lint_ls` false-positive class — CLI verdict is 0 issues; not a real finding.

## e) WHAT WE SHOULD IMPROVE

1. **"New CI job = not done until first green runner run" needs to be a hard personal rule, not a ritual footnote.** The art-dupl-gate class has now bitten twice (golangci-config-verify green-by-assumption 2026-10-13, art-dupl-gate born-broken 2026-10-25). Concrete fix: any session that adds a workflow job ends with either a live runner proof or a TODO_LIST row saying "UNPROVEN ON RUNNER" in the title.
2. **Check upstream issue state BEFORE planning upstream follow-ups.** 5 seconds of `gh issue view` would have saved the phantom "comment + close" plan step. Cheap habit, worth automating into the verify-before-filing flow (its step 1 already re-verifies premises — add "and current issue state if an issue exists").
3. **Route final verify sweeps through `scripts/quiet-run.sh` by default.** Two snapshots in a row now carry the same honest gap. The gate exists, works, and costs one prefix.
4. **The AGENTS.md / documentedPins version-truth drift (cqrs-htmx v4.14.0 vs v4.14.1)** — a version reference in AGENTS.md can silently lag the fixture. The pin-drift guard catches go.mod-vs-latest drift but nothing catches prose-vs-fixture drift. Candidate: a tiny guard that greps AGENTS.md for versioned pin claims and diffs them against `documentedPins` (or: stop stating pin versions in AGENTS prose and always point at the fixture — single source of truth).
5. **Same-day upstream fix → same-day consumer sweep worked extremely well** (filed 17:30-ish, fixed by ~20:00, swept by ~20:05). Worth keeping as the explicit playbook for family-repo asks: file → watch → sweep immediately on the fix tag rather than waiting for trains. This is the first time the loop closed in one day; TODO_LIST's old acceptance ("sweep on their trains") was needlessly slow by comparison.

## f) UP TO 50 THINGS TO GET DONE NEXT (ranked; Impact / Effort / Category)

| #  | Task                                                                                                                                                                                                                     | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ------------- |
| 1  | Push master (wave + v1.7.0 sweep) and run the tip-CI verify ritual on the FINAL pushed commit — every job green except the designed red; `fleet-lint` + fixed `art-dupl-gate` first real runs                                    | Critical | S | Process       |
| 2  | Relocate ci-preflight out of the required workflow into a scheduled secret-watchdog job (owner decision)                                                                                                                     | High | S | Process       |
| 3  | Rebase + merge the 10 open dependabot PRs after the push (error-family v0.12.0 ×4, catalog v4.9.0, go-health v0.7.0, ssetest v0.5.0, grouped bumps)                                                                           | High | S | Cleanup       |
| 4  | Re-run the composition spike against httputil v1.7.0 `NewServerListener` (single-serve ownership, `ListenerAddr()`, `server.listener_consumed`) and write the go/no-go verdict                                                | High | M | Feature       |
| 5  | Restore the `SSH_PRIVATE_KEY` Actions secret (owner-only; un-skips the structure-linter CI leg)                                                                                                                              | High | S | Process       |
| 6  | If spike says go: execute the Service-on-httputil.Server composition refactor (re-pin `Addr()`/`Running()` + shutdown phase-log contract first)                                                                              | High | L | Feature       |
| 7  | If composition lands: Core TLS option on `ServiceConfig` (same seam; PapDashboard trigger check first)                                                                                                                       | Medium | L | Feature       |
| 8  | Run one full `verify-all.sh` (WITH tests) through `scripts/quiet-run.sh` at quiet load — close the two-snapshot honest gap                                                                                                   | Medium | S | Quality       |
| 9  | Fix the AGENTS.md cqrs-htmx pin-version lag (v4.14.0 prose vs v4.14.1 fixture) — one line, do it with the next AGENTS touch                                                                                                  | Low | S | Documentation |
| 10 | Re-run `check-consumer-claims.sh` (network-heavy, skipped today) once before the next tag                                                                                                                                    | Medium | S | Quality       |
| 11 | Date otel's `[Unreleased]` on the next natural otel train (it now carries the v1.7.0 bump + bench protocol notes)                                                                                                            | Medium | S | Release       |
| 12 | Testkit helper dedup (httpGetURL ×2 diverged, poll-until ×3, context-GET ×2, freeAddr sharing) — ride the next core train                                                                                                    | Medium | M | Cleanup       |
| 13 | Guard-symmetry review: consumer-claims (pre-tag-only) vs golangci-configs/structure-linter (CI-only); reconcile fleet-lint with matrix build/vet overlap                                                                     | Medium | M | Quality       |
| 14 | Build the prose-vs-fixture pin-truth guard from §e.4 (or de-version AGENTS prose)                                                                                                                                            | Low | S | Quality       |
| 15 | i18n `Language()` composition E2E leg through the default stack (demand-gated; decide when touching realtime/integration next)                                                                                               | Low | M | Feature       |
| 16 | Hardened health example: consider wiring `LoggingFromContext` there as the second live recipe consumer (after composition settles)                                                                                            | Low | S | Documentation |
| 17 | Update `docs/status/README.md` index with this report (doing in this pass — see closing note)                                                                                                                                | Low | S | Documentation |
| 18 | Optional: mirror the consumer-confirmation comment on httputil#41 (only #43 got one; #41 is equally closed)                                                                                                                  | Low | S | Documentation |
| 19 | After any composition tag: fresh-consumer proxy check per `docs/recipes/fresh-consumer-proxy-check.md` (ritual, not new work)                                                                                                 | Medium | S | Release       |
| 20 | Watchlist pass when convenient: PapDashboard next tag fires the claims guard; cordis consumers count; cqrs+realtime composition-consumer re-check                                                                             | Low | S | Quality       |
| 21 | Keep the art-dupl CI install pin in lockstep with the local binary at every re-baseline (standing rule from the 2026-10-19 false-flag incident)                                                                              | Medium | S | Process       |

That is the honest list — 21 real items, not 50 padded ones. Items 22-50 would be fabrication or duplicates of the above; everything else I noticed this session is either already done (§a) or already tracked. (ROADMAP-class raw ideas — e.g. "composition consumers eventually want the `realtime` SSE stack composed on the same v1.7.0 seam" — belong in ROADMAP, not this table.)

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Do you want me to push master now** (the fix wave + v1.7.0 sweep, all verified) **and then run the tip-CI verify ritual?** Everything is staged behind this. I will not push without your explicit order — that is the one blocker on §d.1.
2. **Do you approve relocating ci-preflight out of the required CI workflow into a separate scheduled watchdog?** It's the root-cause fix for the designed-red normalization that hid 6 days of real failures. I cannot decide workflow-protection policy for the repo.
3. **What is the ambition for the composition refactor now that `NewServerListener` shipped — pursue it as the next major work item (with Core TLS riding the same seam), or keep it parked until a consumer demands it?** Both are defensible; only you know the priority stack.

---

*Point-in-time snapshot. Section (f) items are already harvested into `TODO_LIST.md` (done inline during the sweep) except #8/#9/#10/#14/#18, which are new observations from this closing pass.*
