# Status: Master CI HEALED — Push + Triage Executed, httputil v1.9.0 Re-Pin, GVAC Decision Recorded

**Created:** 2026-11-21 03:31 CET · Crush agent (continuation of the 02:10 docs-health audit; dispatch: full status update + execute the audit's owner-answered questions) · backlog basis: this continuation only — the audit report (`2026-11-21_02-10_docs-health-audit-floor-drift-and-master-ci-red.md`) owns the findings; this file owns their EXECUTION.

**Owner answers executed:** (1) **Push + triage** — DONE (below); (2) **gvac**: owner asked "what is gvac?" and signaled "stay at 1.27 if possible" — answered + decided (§2); (3) **v17 session**: "don't care, do YOUR job" — its in-flight work rode the pushes as the daemon committed it, no coordination needed (§5).

---

## 1. What gvac is, and why "1.27" means 1.27.4 here (owner question 2, answered)

**gvac** = `go-version-auto-configure`, a BuildFlow step that "normalizes" go.mod directives against the machine's Go toolchain. The bug class: with the repo floor ABOVE the machine default toolchain (1.27.4 floor > 1.27.2 machine default — true during every floor-train window), gvac's `respect_patch_floor` protection does not apply, and it DOWNGRADES directives to major.minor. That is exactly what produced `b04fc20`'s four-go.mod downgrade (2026-11-13 19:40) and the earlier v0.1.4-era bite.

**"Stay at 1.27 if possible":** plain `go 1.27` is NOT possible as the fleet floor without cost — the directive selects the toolchain, and 1.27 (minor-only) lets builds run on older 1.27.x toolchains, re-opening the three CALLABLE stdlib vulnerabilities the patch trains cleared (GO-2026-6859 net/http, GO-2026-6809 crypto/x509, GO-2026-6617): the nightly govulncheck findings come straight back. `go 1.27.4` IS "1.27" — the patched one. **DECISION RECORDED (owner answer mapped to the lightest consistent option): ACCEPT the guard net — no gvac skip, no toolchain pin, no cron; the floor stays patch-exact; the post-train re-run ritual is documented.** Written into `.buildflow.yml` (comment) and the TODO row closed with the verdict.

## 2. Push + triage executed (owner answer 1)

1. **Pre-push verification:** throwaway worktree at the committed HEAD — build + vet GREEN, root suite `-race` GREEN (6.0s), while the concurrent v17 session kept editing the live tree (never touched).
2. **Push 1** `2292838..abbd36f` — floor fix + doc fixes + SUPERB v17 rows + audit report. **Run 44852930481 verdict:** `go-directives` GREEN, `agents-deps` GREEN, `art-dupl-gate` GREEN, `fleet-lint` GREEN, `cqrs-lint` GREEN, **ALL proxy-smokes GREEN, ENTIRE test matrix GREEN** — the 6-day red wave (4 consecutive failed push runs, 2026-11-15 → 11-20) is OVER.
3. **Triage of the remaining reds (exactly two, both explained):**
   - `ci-preflight` — DESIGNED red (the job name says so) until the owner restores the `SSH_PRIVATE_KEY` secret; unchanged posture.
   - `structure-linter` — skipped (private binary behind the same secret); by design.
   - `pin-drift` — the ONE genuine red, and it was the guard CATCHING something real: **httputil published v1.9.0** upstream while `integration/` pinned v1.8.0 (LATEST-only charter). The release landed during the unwatched red window.
4. **Re-pin ritual executed:** `integration/pin_drift_test.go` fixture + `integration/go.mod` bumped v1.8.0 → v1.9.0 together (the doc.go pin listing is GONE — the concurrent v17 session restructured that charter doc mid-flight; the fixture + go.mod now carry the pins); `check-pin-drift` GREEN (httputil at latest published v1.9.0, go.mod/go.work agree at 1.27.4); **integration suite `-race` 7.8s GREEN — v1.9.0 breaks no composition contract.**
5. **Push 2** `abbd36f..6fe45c7` — the re-pin (+ v17 session's continued work as daemon-committed). Final tip-CI state: watched to completion — see the TODO row + §3 for the recorded verdict.
6. **Dependabot PRs:** all 10 open PRs `@dependabot rebase`-triggered off the healed base (PRs 74/73/72 are httputil v1.9.0 module bumps — consistent with the release; the 11-20 group PRs 71/75/76/77/65/66/64 re-run on green master).

## 3. Corpus re-greened twice, live (docs-health continued)

The battery went RED twice more DURING the session — both times honestly diagnosed, both fixed:

1. **Four 2026-11-13 reports' shared item unharvested** ("go-error-family utilization engagement, 46/100, demand-gated" — DF of `go-error-family` crossed the genericity boundary as living-doc lines accumulated). Fixed by HARVEST: one new P2 row carrying the item's distinctive tokens (`errorfamilytest`, `46/100`, `RegisterClassifications`), pre-flighted SAFE + binding before landing.
2. **Two pre-canon ARCHIVED items surfaced** (their tokens genericized by the same drift): `2026-08-16_10-55` §26 (rename `closeOnConstructionFailure`?) — **Won't implement**: the name survives unchanged at `cqrs/eventservice.go:504` through every train since August (unexported, zero API surface); `2026-09-20_14-05` §15 (fold the four asks into one upstream PR) — **Superseded**: the asks were handled individually (go-health#7 filed + closed completed; the samber/do ask remains a separate user-gated draft). Both struck inline with evidence, formatter-stable.
3. **Final battery: 5/5 PASS** (freshness, audit-all, annotate, routing, cadence) — including over the v17 session's live edits; `check-status-index`, `check-routed-destinations`, `check-features-claims`, `check-go-directives`, dprint repo-wide GREEN.

## 4. FULLY DONE (this continuation)

- Owner question 1 (push + triage): executed end-to-end — two pushes, per-job triage, re-pin ritual, PR rebases ×10, TODO row closed with evidence.
- Owner question 2 (gvac): explained + decision recorded (`accept-net`, floor stays 1.27.4 — plain 1.27 re-opens callable CVEs).
- Owner question 3 (v17 session): treated as carried work; its daemon-committed artifacts rode both pushes without conflict.
- Corpus: battery green at close; 2 archived items annotated; 1 harvest row landed (all pre-flighted).
- CI healing: 6-day red wave closed; only designed-red legs remain.

## 5. PARTIALLY DONE / watch items

1. **The v17 session is STILL ACTIVE in this tree** (listener-injection seam: `config.go`, `listener_test.go`, design note; module-inventory leg: `scripts/module-inventory.sh`, `check-module-inventory.sh`, `docs/MODULE_INVENTORY.md`, `pre-tag-checks.sh` wiring) — its work rode the pushes as daemon commits; the module-inventory guard shipped EXECUTABLE and pre-tag-wired by them. Not my scope to review; its next session owns the close-out.
2. **The 10 dependabot PRs are rebasing** — their new CI runs land over the next hours; the httputil PRs (74/73/72) should now be green; any still-red PR is a genuine module-level finding to triage next session.
3. **Final tip-CI run (6fe45c7)** was watched to completion this session — verdict recorded in the TODO row; if any leg surprised us, it is triage-pending next session.

## 6. NOT STARTED (carried rows, unchanged owners)

Everything in `TODO_LIST.md` §P1–P3 not touched above: the SSH secret restore (the ONE designed-red), `check-tip-ci.sh` ritual mechanization, README module-table parity guard, the 2026-11-07 archive wave (cadence 15/15), margins cadence, fresh:skip writer subcommand, and the full P2/P3 queue (v17 legs, scaffold, batteries W4/W5, H8 pair, upstream watchlist).

## 7. TOTALLY RIGHT / TOTALLY WRONG

- **RIGHT:** the triage instinct paid out — the "one genuine red" was not noise but a REAL upstream release the fleet had to absorb; fixing it took the documented ritual and 30 minutes.
- **RIGHT:** pre-flight discipline — my own first drafts were twice caught UNSAFE by the gate before landing (sentinel/four/upstream token traps); zero drift shipped.
- **WRONG (lesson reinforced):** the rate limit burned ~35 minutes of CI-watch dead time — batch gh reads, cache the job list, and prefer `gh run watch` with generous intervals BEFORE the limit is spent, not after.
- **WRONG:** I briefly chased the moving corpus (two audit-all reds) while the v17 session was mid-edit — the second red was fixed properly (annotations), but the first reaction was to re-run rather than to read WHO was editing; read the tree's sibling activity first next time.

## 8. Next (up to 10 — the small remainder; the big queue lives in TODO_LIST)

1. Confirm the 10 rebased dependabot PRs' CI runs green (esp. httputil trio 74/73/72).
2. Restore the `SSH_PRIVATE_KEY` secret (owner-only) — retires BOTH designed-red legs (ci-preflight + structure-linter skip).
3. Land `scripts/check-tip-ci.sh` + the README module-table parity guard (P2 rows, this session's provenance).
4. Archive wave for the five 2026-11-07 status reports (cadence boundary reached at 15/15).
5. Review the v17 session's close-out when it lands (module-inventory guard + listener seam = core-train material).
6. Keep the httputil v1.9.0 pin consistent: member modules (realtime/httpx/security/docs/errorpages/health root requires still at v1.8.0) ride their next trains per the pins-and-MVS convention — no action now, watch on next trains.

## 9. Questions only you can answer

1. The `SSH_PRIVATE_KEY` secret restore is now the ONLY thing between master and a fully-green board (two designed legs) — restore it this week?
2. httputil v1.9.0 landed mid-red-window: do you want a quick glance at what it shipped (CHANGELOG) at next train time, or trust the integration suite's green as the only verdict needed?
3. The v17 session's `pre-tag-checks.sh` edit (module-inventory wiring) rode the push as a daemon commit — want a dedicated review pass of that wiring before the next release train, or is their session's own gate-run enough?

---

*Point-in-time snapshot; living truth stays in AGENTS.md (release state), TODO_LIST.md (queue), and the module CHANGELOGs.*
