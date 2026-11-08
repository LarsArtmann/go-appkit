# Whole-List Execution Report — Guards, the Drain Rule, Corpus Zero, and the Three Incidents

**Date:** 2026-11-07 01:34 CET (session ran 2026-11-06 23:05 → 2026-11-07 01:20)
**Trigger:** the owner mandate: execute the ENTIRE actionable TODO list from
the wave-3 close-out, keep the guards green between every move, push, verify
CI. This report covers THAT run only — no other research folded in.
**Session base:** entered at `4bab701` (wave-3 close-out, gates 5/5, corpus
zero), exited at `cd1b2e9` (pushed; CI green on the final commit with only
the annotated designed red).

---

## a) FULLY DONE (verified, on origin)

1. **The eleven guard/tooling asks — all shipped + wired** (TODO row closed
   with evidence): strict claims parser naming offenders + the
   release-window rule (`check-pin-drift.sh` — origin-aware via one
   ls-remote roundtrip, graceful offline degradation, loud NOTE for
   local-only tags instead of drift); `check-gosum-canonical.sh` local-only
   pin DEFERRAL (probe-proven with a throwaway tag before it ever mattered);
   `check-tagged-changelogs.sh` (wired as pre-tag step 3; born-red on the
   KNOWN config/v0.1.0 re-cut artifact, green on thirteen other tags);
   `check-features-claims.sh` (drift-probe proven);
   `check-module-licenses.sh` (born-red until the integration exemption was
   stated in `integration/doc.go` — the guard wrote a doc); `tag-module.sh`
   (dirty-tree refusal, exactly-once tags, full gate, push,
   `--tidy-integration`); the nightly `-count=5` integration workflow
   (`flake.yml`, FIRST LIVE RUN SUCCESS via workflow_dispatch); the
   ci-preflight designed-red annotation + job name (16-13 item 30);
   `check-touched-modules.sh` (fatal in pre-commit, ran live on its own
   commit); `verify-consumer.sh` (scratch module + exact version + usage
   snippet; born-green on health@v0.3.1 with a NewProbe probe); cqrs-lint
   CI PINNED at upstream stable v4.21.0 (verified green on the cqrs tree
   first) + the F032 framework-wrapper memo drafted in-repo.
2. **The drain rule — issue54 §e5 root-caused once:**
   `scripts/check_response_drain.go` (go/parser AST walker per the
   linter-building decision tree), 44 fleet sites drained with
   `io.Copy(io.Discard, X.Body)`, 10 SSE carve-outs suppressed with
   explicit reasons, wrapper `check-response-drain.sh` wired into
   pre-commit + pre-tag + CI. Hermetic `-race` suites green across root,
   otel, preset, realtime, flightrecorder, integration (published pins).
3. **Annotate corpus drained 243 → 0 and the ratchet closed:** `annotate_max`
   255 → 0 (drain first, then lower — the gate's own instruction). The
   drain: 236 G/uniform per-cell strike normalizations across the four
   2026-06/07 reports + nine D/verdict-leads-open items rewritten to
   marker-bearing verdicts (learned live: the marker vocabulary is strict —
   `done`/`Won't implement` — the verdict rides AFTER the strike, struck
   cells need markers, table-cell inline code must not contain raw pipes).
4. **TODO_LIST hygiene:** ten+ rows CLOSED with evidence (guard batch,
   JSON-v1 STALE verdict — both flagged files already on `encoding/json/v2`
   APIs, verified against the v2 godoc; integration hygiene — the header is
   already detached; touched-modules; preflight visibility; watchlist
   split; errorpages godoc; HTML trio; cqrs-lint debts; the heads-ups row
   closed by the concurrent session for #43/#6). The composition-gap
   watchlist monster rewritten as the compact pinned-compositions table.
   The go-health#7 claim now carries the v0.3.1 re-pin (issue54 item 22).
5. **The go-health forward-compat scratch — GREEN:** health v0.3.1 compiles
   AND `NewProbe` returns a `*gohealth.Probe` under MVS-forced go-health
   v0.10.0 + dashboard v0.13.2 (isolated-cache scratch). The next health
   train is a decision, not a discovery. Recorded in the upstream watchlist
   row.
6. **Docs micro-fixes:** the HTML trio ruled keep-as-reference with both
   broken `<title>`s fixed (09-04 placeholder, 08-15 missing); errorpages
   `Mount` godoc carries the catch-all-405 clarification (ships with the
   next errorpages tag); health README gained the role=region note and
   lost the retired `GOEXPERIMENT=jsonv2` claim; module-details frh line
   corrected to go-health v0.8.1; the recorder-sentinel draft Status line
   now says RESOLVED; `docs/recipes/docs-gates-cadence-and-anchors.md`
   (cadence override rationale + the F-80 anchor-sweep verdict: living
   docs clean, point-in-time docs stay).
7. **Long-tail F-78/79/80 done:** dprint round-trip on a fresh:skip file
   verified (5/5 markers survive, check stable); the cadence rationale doc
   (§6 above); the anchor sweep (no `TODO row N` anchors in living docs).
8. **The three upstream docs-gates notes** written into the SKILLS repo
   (`docs-health/references/upstream-notes.md`): routed≠resolved
   gate-surface gap; the compound-block positional-renumbering quirk
   (REPRODUCED live with a fixture before writing); the DF-boundary
   `--explain-df` ask (verified the flag does not exist). Committed
   LOCALLY in SKILLS — not pushed (owner's push).
9. **AGENTS.md lessons recorded:** the probe-tag incident, the
   concurrent-filing race, and the corrected v2.15.0 drift truth (the
   findings were REAL; the config sig-extension was the fix).
10. **Verification spine:** battery 5/5 PASS at annotate pin 0 · routed
    guard 0 dead / 0 unenforced · index liveness green both directions ·
    pre-tag gate GREEN with all fourteen guards · CI green on the final
    pushed commit with only the annotated designed red · actionlint green
    on both workflows · all thirteen modules lint 0 under the new binary.

## b) PARTIALLY DONE

1. **The upstream notes are local-only** — the SKILLS commit is unpushed
   (push is the owner's). The notes exist and are verified; delivery is
   pending one `git push` in the SKILLS checkout.
2. **The testkit drain-and-close helper (issue54 §e5's full ask)** — the
   LINT RULE shipped and the ~44 sites drain inline, but the copy-paste
   consolidation into a `testkit` helper was NOT built (the ask said
   "copy-pasted io.Copy at ~15 sites is not a fix; a helper plus a lint
   rule would be" — I shipped half: the rule). The rule prevents the
   REGRESSION class; the helper remains the dedup half.
3. **`check_response_drain.go` distribution** — a repo-local
   build-tagged walker, not a golangci plugin or a promoted testkit
   analyzer. Right-sized for today, but it is one repo's rule, not a
   fleet instrument.
4. **The marker-vocabulary learnings from the annotate drain** (done-after-
   strike, struck-cell markers, pipe-free code spans) are recorded in this
   report and the archived wave reports but NOT yet promoted into the
   docs-health skill's annotation-placement reference — they are exactly
   the class the `repair` verbs should mechanize.

## c) NOT STARTED (deliberately — owner-gated or demand-gated)

1. Floor-train go 1.27.2 → 1.27.3 (standalone vs ride — owner call; the
   row documents both paths; GO-2026-6809 is stdlib-only).
2. Machine ops batch (owner present required): /tmp sweep policy,
   quiet-run gates, BuildFlow binary rebuild, cache VACUUM, interrogate
   install, root-owned ~/.cache/go-mod.
3. The `SSH_PRIVATE_KEY` secret restore (owner-only; ci-preflight remains
   the single designed red).
4. pma daemon §6 decisions + the working-file ignore ask (owner's config).
5. Plan-revision convention ruling (owner ask from the 01-33 report).
6. The cqrs-lint F032 memo's UPSTREAM filing (maintainer act on
   go-cqrs-lite; the memo is drafted and committed in-repo).
7. Per-lib errata for dashboard v0.12.2 / templ-components a11y
   (issue54 item 18, routed to micro-backlog) — untouched this session.
8. Everything demand-gated and untouched by design: the scaffold decision
   - template, adoption-wave swap recipes, the W3-W5 battery program
     (httpx/worker/polite/realtime completions), the `httputil.Server`
     composition refactor (unblocked, unplanned), core v1.0.0 exit-criteria
     graduation, the pkg.go.dev flightrecorder render watch, the satellite
     test-helper migrations (next core train), the preset six-month
     adoption re-sweep, AuthBridge/SigNoz/cross-repo posture gates.

## d) TOTALLY FUCKED UP (own goals, in order of severity)

1. **The SSE drain codemod hung the realtime suite for 10 minutes.** The
   first codemod pass inserted `io.Copy(io.Discard, resp.Body)` before
   closes in SSE tests — io.Copy BLOCKS on an open-ended stream; four
   tests wedged to the 10m panic. Root cause of the mistake: my
   discrimination fixture had no streaming-body case, and my helper-trust
   regex was factually wrong for this fleet (I trusted `httpGet`-style
   names to drain; source-verified AFTER the failure, every fleet helper
   returns the RAW response). The shipped design (explicit per-site
   `respdrain:ignore` carve-outs) is correct BECAUSE the failure forced
   the distinction — but the correct sequence was: verify helper behavior
   FIRST, then codemod. Cost: one suite hang, one re-run, one unused
   import fix, one revert of nine insertions.
2. **My probe tag reached origin and broke CI fleet-wide.** I created
   `preset/v9.9.9` for the gosum-deferral probe, deleted it locally…
   and the pma daemon PUSHED it in its next cycle. pin-drift went red on
   origin (newest tag = v9.9.9), requiring an origin tag deletion. No
   proxy-cache damage (nothing ever pinned it), but a fleet-wide red
   from my own junk ref. The AGENTS rule now says: probe tags are
   create-probe-delete in ONE command chain + `git ls-remote` closure
   check. The rule existed as "delete after probing" BEFORE this
   session's summary — I followed it and it was still insufficient.
3. **The heads-up double-post.** I posted cqrs-htmx#44 and PapDashboard#7
   with claims verified and voice-checked — without checking whether a
   concurrent session was posting the SAME notices. It was, minutes
   earlier (#43/#6). Mine closed as duplicates. verify-before-filing
   gates the CONTENT; nothing gated the CONCURRENCY. Now in AGENTS.md:
   `gh issue list --search` before any filing.
4. **My own dedup gate caught my own clone.** `check_response_drain.go`
   shipped with two byte-identical ident helpers (`identText`/`pkgName`);
   the pre-tag probe flagged a NEW clone group against the baseline.
   Fixed at the cause (one delegates to the other), no re-baseline —
   but the author of a dedup wave authored a clone.
5. **The pin-drift rewrite shipped a bug on its first run** (the claims
   builder loop used `printf '%s'` without a trailing newline, dropping
   the LAST entry — preset — from every run). Caught immediately by the
   guard going red on the current tree; the strict parser's value was
   proven by it catching its own author. Still: the rewrite was not
   run against a probe BEFORE its first real invocation.
6. **Small but real:** (a) I invoked a nonexistent
   `docs-gates freshness --check-rows` flag (check-rows belongs to the
   annotate binary); (b) several "logical" commits landed as daemon
   fragments (8+ auto-commits interleave the session; the drain-rule
   commit carried 2 of ~20 files) — history is more fragmented than the
   house style wants, and the daemon-proof rule forbids rebalancing; (c)
   a piped `EXIT=$?` read tail's exit code twice before I re-grounded in
   the no-pipe rule.

## e) WHAT WE SHOULD IMPROVE

1. **Codemods need negative fixtures BEFORE the tree, not after the
   suite.** A streaming-body fixture in the drain-rule probe set would
   have caught the SSE class in seconds. Generalized: every codemod that
   touches semantic behavior runs against a fixture corpus first (the
   annotate tooling already works this way; my codemods did not).
2. **Verify the domain FACTS before trusting naming conventions.** The
   helper-trust regex encoded "httpGet drains" — one `grep -A6 'func
   httpGet'` would have falsified it in 30 seconds. The failure cost
   more than the check would have.
3. **One-command commit+push discipline** (`git add … && git commit … &&
   git push` in a single chain) shrinks the daemon race window — I used
   it sometimes and lost fragments other times. Make it the default
   posture for every commit this repo.
4. **CI fleet-lint should pin the golangci-lint version** (the
   golangci-config-verify job pins v2.14.0 for schema; fleet-lint floats).
   The v2.15.0 drift went unnoticed until a push; a pinned install
   converts machine drift into a deliberate bump ritual.
5. **The annotate repair verbs should grow a `reword-verdict` /
   `rebalance-strikes` mode** — the 243→0 drain was ~30 hand-built
   replacements across ten files; the tool could mechanize the whole
   class (it already has repair/unwrap/repair-partial).
6. **Filing procedure gains a concurrency gate:** `gh issue list
   --search <topic>` before ANY issue/PR post (recorded in AGENTS; the
   verify-before-filing skill could own it permanently).
7. **The docs-gates routed≠resolved note (already drafted upstream)**
   deserves the gate-surface fix itself: a freshness summary line
   counting routed-open items would have made several of this week's
   judgment calls legible without tribal knowledge.

## f) UP TO 50 THINGS TO GET DONE NEXT

1. Push the SKILLS repo (the docs-gates upstream notes commit is local).
2. File the cqrs-lint F032 memo upstream (go-cqrs-lite maintainer act).
3. Floor-train go 1.27.2 → 1.27.3 — decide standalone vs ride, execute.
4. Restore the `SSH_PRIVATE_KEY` secret (retires the designed red + the
   structure-linter skip + the preflight job itself).
5. Build the `testkit` drain-and-close helper; migrate the 44 inline
   `io.Copy(io.Discard, …)` sites to it (dedup half of §e5).
6. Promote the marker-vocabulary learnings into the docs-health skill's
   annotation-placement reference (done-after-strike, struck-cell
   markers, pipe-free table code spans).
7. Add `reword-verdict` / `rebalance-strikes` verbs to the annotate
   tooling (mechanize what the 243→0 drain did by hand).
8. Add a streaming-body case to the drain checker's fixture set (the
   fixture that would have caught §d-1).
9. Pin golangci-lint in the fleet-lint CI job (kill the machine-drift
   class; bump = deliberate ritual).
10. Ratify the `annotate_max` ratchet-down semantics (see g-3) — pin 0
    vs a small buffer.
11. Per-lib errata for dashboard v0.12.2 + templ-components a11y
    (issue54 item 18, the one micro-backlog straggler untouched).
12. Machine ops batch (owner present): /tmp sweep policy, quiet-run
    tmp-space + cache-writability gates, BuildFlow binary rebuild
    (74a9212 → 2d6bbee), cache-DB VACUUM + legacy state-DB cleanup,
    interrogate install, root-owned ~/.cache/go-mod fix.
13. pma daemon §6 decisions (owner checklist: hook bypass keep, cooldown,
    staged-respect, ignore_patterns for `.*-body.html`).
14. Plan-revision convention ruling ("rev N (expanded)" — owner ask from
    the 01-33 report).
15. Preset health-pin v0.3.1 ride — preset's next tag.
16. Decide `health v0.3.2` (LICENSE parity tag) vs ride the next train
    (owner call, routed 2026-11-06).
17. Scaffold decision + template (v12 tier 2 — verdict doc first).
18. Adoption-wave battery-swap recipes (dnsblockd, ChastityAPI,
    PapDashboard security swap, pbx-lens cmdguard migration).
19. The W3 httpx module program (B1-B9) — the composition-mission's
    biggest planned surface; H8 `svc.Routes()` seam decision feeds it.
20. The `httputil/endpoints` v0.1.0 verdict (route-registry wheel; H8/B6
    fit) — `docs/review/libs/endpoints.md`.
21. The `httputil.Server` composition refactor (unblocked since v1.7.0;
    plan-of-record exists; separately planned core train).
22. Core v1.0.0 exit criteria — the additions-empty tag pair + consumer
    version overlap blockers.
23. The satellite test-helper migrations (realtime/fr httpGetURL →
    testkit; rides the next core train).
24. pkg.go.dev flightrecorder render watch (owner-gated re-tag last
    resort).
25. The consumer heads-up follow-through: watch #43/#6 for consumer
    responses; the claims guard fires on their next tags.
26. The concurrent-filing gate: fold `gh issue list --search` into the
    verify-before-filing skill permanently (it currently lives only in
    AGENTS.md).
27. `check-consumer-claims.sh` CI twin (issue54 §e4 — it is
    pre-tag-only; a stale FEATURES claim class sat red-free before).
28. The `docs/research HTML trio` responsive-baseline ask for FUTURE
    reports (html-report-kit lane; 12/15 tracked artifacts overflow a
    500px viewport — class observation from the audit loose ends).
29. The audit loose-ends remainder: planning-docs "rev N (expanded)"
    convention (same as 14; owner).
30. `docs-gates freshness` routed-count surface (upstream note 1's
    proposal — implement if the docs-gates window opens).
31. `--explain-df` observability (upstream note 3's proposal).
32. The compound-id printed-id mapping (upstream note 2's proposal —
    `emit-keys --printed`).
33. A `flake.yml` run history review after a week of nightly data (is
    `-count=5` finding anything? calibrate the cadence).
34. Nightly `govulncheck` in CI (the floor-train row's mechanical twin —
    catches the NEXT GO-2026-xxxx before a train, not during).
35. `check-pin-drift.sh`: extract the claims parser into a testable
    fixture mode (`--claims-file`) so parser rewrites are probe-able
    without touching AGENTS.md (the §d-5 class).
36. Probe-tag hygiene as a script: `scripts/probe-tag.sh` (create-probe-
    delete-closure in one chain; the AGENTS rule made mechanical).
37. The `security_realtime_test.go` floating-header pattern: sweep the
    other integration test files for accidental attached package docs.
38. `errorpages` next-tag prep: the godoc clarification + anything else
    pending rides one docs-only train.
39. The annotate drain post-mortem as a skill note: "drain to 0, then
    lower the pin" is now PROVEN — write it into the docs-health SKILL
    as the standing annotate-max ritual.
40. Sweep the archived corpus for other `routed — TODO_LIST` bare
    verdicts (the class the 14-22 file carried; the checker flags
    them only as missing markers — a grep sweep finds the rest).
41. `verify-consumer.sh`: support multi-module targets in one scratch
    (composition verification: preset+health+realtime together).
42. Forward-compat scratches for the OTHER fleet deps (cqrs-htmx v4.16,
    go-sse v0.7.0/templ-components v1.27 — the held-pins unblock
    triggers deserve the same decision-not-discovery treatment).
43. The held-pins removal runbook: catalog v4.13+ migrated tag check →
    cqrs-htmx v4.16 + go-sse v0.7 + templ-components v1.27 bump wave.
44. `docs/recipes/fresh-consumer-proxy-check.md` → point at
    `verify-consumer.sh` (the recipe predates the script; one canonical
    entry point).
45. AGENTS guard-count wording: the fatal pre-commit set is now NINE
    guards + touched-modules — re-count the prose (the "five OFFLINE
    standing guards" comment was already updated; audit the rest).
46. The `flake.yml` cron offset: 04:17 UTC lands mid-workday CET;
    consider 02:17 if the nightly competes with anything.
47. `middleware_test.go` and the drained root tests: the six inserted
    blank lines are cosmetic silencers — a tiny `httpGetDrained` helper
    in testhelpers_test.go would remove the wsl_v5 pressure at the root
    (pairs with 5).
48. Sweep for other linters' version drift proactively: run ONE bare
    latest-binary lint pass per quarter and absorb REAL findings before
    CI does (the v2.15.0 class, generalized).
49. The SKILLS repo's docs-gates notes → also file the annotate
    marker-vocabulary findings there (this report §b-4 once promoted).
50. Update `docs/recipes/_proxycheck/README` cross-links: two fresh-
    consumer entry points exist now (recipe script + verify-consumer.sh);
    one should defer to the other.

## g) QUESTIONS FOR YOU (cannot answer from the repo)

1. **Floor-train:** cut go 1.27.2 → 1.27.3 as a standalone floor train
   NOW (clears GO-2026-6809 immediately; integration re-pin ritual
   applies), or hold to the next release train as the TODO row's default
   says?
2. **SKILLS push + upstream filings:** push the SKILLS repo (the
   docs-gates notes commit is local-only) and file the cqrs-lint F032
   memo as a go-cqrs-lite issue — both are maintainer acts I will not do
   unbidden. Proceed with both, either, neither?
3. **The annotate pin:** I lowered `annotate_max` 255 → 0 after draining
   to zero, reading the ratchet as drain-then-lower (the gate's own
   instruction says so; the wave guardrail's "no threshold edits" was
   aimed at defeating reds). Ratify pin 0 as the standing state, or do
   you want a small buffer (e.g. 5) to absorb incidental findings
   without a same-day ritual?
