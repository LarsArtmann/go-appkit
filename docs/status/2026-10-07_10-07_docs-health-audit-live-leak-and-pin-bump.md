# Session — docs-health AUDIT, Live Leak #4 Catch, cqrs-htmx Pin Bump

> 2026-10-07 10:07 CEST · session report (docs-health AUDIT pass on TODO_LIST +
> FEATURES + live drift fixes caught by the guards mid-audit). Scope: THIS
> session only. Prior context: the 2026-10-07 01-22 session-final report (whose
> (f) list is the harvest source); TODO_LIST was last harvested BEFORE that
> report landed (`f561048`).

---

## a) FULLY DONE

1. **docs-health AUDIT executed** (skill loaded first): TODO_LIST + FEATURES
   verified against code, git history, and LIVE origin CI state (`gh`).
2. **Harvested the 01-22 session-final report** (it postdated the last harvest):
   2 completed TODO items deleted on origin-verified evidence —
   - _CI resilience_: SHIPPED as `ci-preflight` (`67c2a61`) and **observed
     live** on origin run `37546484597` (`5b00c7c`): preflight = the ONE
     designed-red job, `test` matrix = SKIPPED (not 12 reds), all guards +
     11 proxy-smoke jobs green.
   - _workspace-charter observation_: job green on `ae3b2d4` AND `5b00c7c`
     (and it caught real leak #3 before going green).
3. **P1 narrowed + stale companion dropped**: the ssh-agent/git-config steps
   were already gated `github.event_name != 'pull_request'` since `1c4d9b9`
   (verified in `.github/workflows/ci.yml:77`) — the "gate off PR events"
   companion was done work listed as open.
4. **Daemon item updated** with the 01-22 facts (BuildFlow gomod-check
   exonerated; `go work use -r .` sorted-insert signature; reflog-silent
   rewind; push stall) and the leak count corrected to FOUR with verified
   hashes: #4 added in `620a3b1` (09:53:48, DURING this session), removed in
   `19100c5`.
5. **5 new P2 items routed** from the 01-22 (f) list: per-module pin-drift
   guard, push-lag watch, post-commit HEAD-advance hook, cqrs flake hunt,
   AGENTS pipe rule. Open question (6) daemon-location added.
6. **FEATURES.md version-truth fixed**: 7 stale "UNRELEASED at v0.1.x"
   annotations de-staled (otel 5, flightrecorder 2 — both v0.2.0 shipped
   2026-10-06); cqrs v0.7.0 rows ADDED (`EventConfig.Domain`, health/SCREAM
   accessors, engine-backed checkpoints, always-close Shutdown, README sync
   guard); flightrecorder v0.2.0 rows added (async capture, `?download=1`,
   `OpsRecorderLoggerPreset`, example); systemd `Counters()` row added. Every
   cited test name grep-verified to exist.
7. **Guards caught TWO live issues mid-audit** (this is the guards-as-code
   pattern paying off inside the session that documented it):
   - **go.work leak #4** (`./integration` re-added by daemon commit `620a3b1`)
     — removed; charter guard green.
   - **integration cqrs-htmx pin drift** v4.13.1 → v4.13.2 — bumped in
     `integration/go.mod` + `documentedPins` fixture; hermetic
     `GOWORK=off GOTOOLCHAIN=go1.27.1` build + vet + **full suite
     `-race -count=1` GREEN (2.3s)**.
8. **AGENTS same-train fix**: the cross-repo pin parenthetical (still saying
   cqrs-htmx v4.13.0 / go-sse v0.6.1 / httputil v1.4.0) updated to
   v4.13.2 / v0.6.2 / v1.4.1 — the same-train release-state rule applied to my
   own bump (initially missed; see d2).
9. **Claim-wording fix**: "root v4.13.2 published 2026-10-07" → "first seen by
   the pin-drift guard 2026-10-07" (inference ≠ verification).
10. **Formatting via BuildFlow** (`buildflow -s dprint-format --fix`, skill
    loaded first; initial `-s dprint` corrected to the real provider name) +
    structure linter **0 findings** bare (the daemon-committed
    `.go-structure-linter.yaml` is honored).
11. **Final state: all four guards green** (pin-drift, go-directives,
    workspace-charter, structure linter); working tree picked up by the daemon
    (`19100c5`, `31526ce`).
12. **Documentation Health Report delivered inline**: Accuracy 10.00 /
    Fitness 10.00 — scoped to the audited docs with explicit caveats (though
    see d2 for the overclaim inside that report).

## b) PARTIALLY DONE

1. **FEATURES.md VERIFY**: version-truth + newly cited test names verified;
   every historical evidence cell was NOT re-verified.
2. **AGENTS verify**: release claims guard-checked + one stale parenthetical
   fixed; no full AGENTS pass (still 361/377 counted lines — lean).
3. **README.md / ROADMAP.md / per-module CHANGELOG bodies**: untouched,
   not re-verified this session.
4. **`.buildflow.yml` license-check skip REMOVAL** (daemon commits `620a3b1` +
   `31526ce`, NOT authored by this session): flagged to the owner, provenance
   unknown, not investigated further (AGENTS documents that step as unable to
   load this repo — pre-1.27 go in the nix-run sandbox).

## c) NOT STARTED

1. The 5 newly routed P2 items themselves (routing ≠ execution).
2. P1 secret restore (owner-only).
3. core v0.8.0 + systemd v0.1.0 train; Example StartHooks migration; fr v0.2.1
   sweep; all other standing TODO_LIST items (untouched, remain valid).
4. otel×realtime / docs-module / flightrecorder-middleware E2E legs
   (watchlist, not picked up).

## d) TOTALLY FUCKED UP

1. **Wrote an unverified hash claim into TODO_LIST during an accuracy pass**:
   "#4 via heuristic commit `19100c5`" — `19100c5` is the commit that REMOVED
   the leak (my fix); the ADD was `620a3b1`. Caught only in this report's
   self-review, then fixed. Exactly the claim-drift class the session existed
   to eliminate.
2. **Overclaimed in my own health report**: the table row "AGENTS.md — fresh
   (guards green)" was wrong when written — the guards do NOT check AGENTS
   prose, and the cross-repo pin parenthetical was stale (my own pin bump
   made it staler). Found during self-review; fixed (a8). The 10/10 Accuracy
   score was true only after this fix.
3. **Stated an inference as a fact**: "root v4.13.2 published 2026-10-07" —
   I never checked the tag's publish date, only that the guard sees it as
   latest TODAY. Softened to "first seen by the guard".
4. **multiedit anchor bug**: replaced the structural-fix item's HEADER line
   without re-including it in `new_string` — had the match been clean it would
   have silently DELETED a live TODO item. It instead failed + glued its body
   onto my pipe-rule item (intermediate corruption at old line 35), repaired
   in a follow-up edit. Anchor-must-survive rule violated.
5. **Piped verification output through `head` 3+ times** (`gh run list … |
   head -12`, `gh run view … | head -50`) in the SAME session that routed the
   "never pipe verify commands — redirect to a file" TODO item. The 01-22
   report's d1/d3/d4 incidents were the exact lesson. Nothing was lost — by
   luck, not discipline.
6. **Read-first violation**: first FEATURES.md multiedit rejected ("must read
   the file before editing") — bash `sed`/`grep` reads don't count. Known
   rule, violated, corrected via View.
7. **Edited against a moving file twice**: both claim-wording edits bounced on
   stale-read (daemon + dprint reformatting mid-edit). Recovered via sed, but
   the second bounce was avoidable — I re-edited without re-reading after the
   first bounce told me the tree was live.
8. **Never diagnosed WHY multiedit edit 2 failed** — no old_string-vs-file
   diff, just re-applied through different anchors. Understanding > retrying.

## e) WHAT WE SHOULD IMPROVE

1. **Guards-as-code over prose** (echoing 01-22 e4, now with a second proof):
   the per-module pin-drift extension (AGENTS-says-vX vs tag-says-vY) would
   have caught d2 mechanically. It is this harvest's top new item.
2. **Edit discipline**: when replacing an anchor line and inserting before it,
   the anchor MUST appear in `new_string`; with the daemon live in the tree,
   re-View the exact region immediately before every edit.
3. **Pipe discipline applies to ME**: file-redirect every `gh`/guard/grep
   whose truncation could mislead — the rule I routed is the rule I broke.
4. **Claim discipline**: hashes and dates are claims. "Verified via X on DATE"
   is always writable; "X happened on DATE" needs a source.
5. **Run the guard battery at SESSION START**: leak #4 sat in the tree for the
   first half of this session; I first ran the guards after the doc edits. A
   baseline run is cheaper and catches daemon interference before it mixes
   with my diffs.
6. **Daemon heuristic commits erase attribution**: this session's work is in
   history as "chore: auto-commit N changed file(s)" with no what/why. This
   report doubles as the session manifest (the 01-22 report did the same —
   keep the pattern until the daemon learns commit messages).

## f) Top things to get done next (ranked)

1. **P1 (owner): restore `SSH_PRIVATE_KEY`** — with ci-preflight live, it is
   the ONLY step left to a green master (matrix currently skipped).
2. **Per-module pin-drift guard** — AGENTS-says vs tag-says for every module +
   the cross-repo parenthetical; kills the d2 class mechanically.
3. **Push-lag / origin-age watch** (guard script; the 6-day CI blindness
   class).
4. **Post-commit HEAD-advance sanity hook** (design first; the `bb47fbc`
   rewind class).
5. **cqrs flake hunt**: `-count=10` under load once; capture the name if the
   transient recurs (01-22 d3).
6. **AGENTS pipe-rule one-liner** (+ a removal trade; 361/377).
7. **Daemon triage (owner)**: unit/config/logs — explains the rewind, the
   workspace heuristic, and leak cadence (#4 landed mid-session).
8. **`.buildflow.yml` license-check decision (owner)**: restore the skip or
   confirm BuildFlow's runner got a ≥1.27 toolchain (g3).
9. **Full FEATURES.md evidence-cell verify pass** (version-truth done;
   per-row evidence not).
10. **README.md verify pass** (untouched today).
11. **ROADMAP.md verify pass** (untouched today).
12. **Observe the matrix UNSKIPPED and green** once the secret is restored
    (closes P1's acceptance line).
13. **core v0.8.0 + systemd v0.1.0 train** (steps enumerated in TODO P3;
    demand-proven by bank-sync ADR-017).
14. **Example StartHooks migration** (flightrecorder + otel examples) after
    core v0.8.0.
15. **fr v0.2.1 sweep** on the frh + cqrs next trains.
16. **File the go-cqrs-lite upstream ask** (user-gated; draft complete).
17. **File the go-health upstream-ask pack** (user-gated).
18. **File the go-sse `ReplayFiltered` ask** (user-gated).
19. **Compose `Service` on `httputil.Server`** — blocked on `NewServerListener`
    upstream API (standing; do not refactor today).
20. **E2E gaps**: docs-module + flightrecorder-middleware composition legs
    through a live Service.
21. **otel×realtime SSE-through-OuterMiddlewares E2E** (watchlist head).
22. **Browser CSP pass** on the hardened health dashboard (strict-CSP + nonce).
23. **errorpages `statusRecorder` → `httputil.ResponseRecorder`** (USER GATE,
    ~10 lines).
24. **Health/polish backlog batch** (hardened-dashboard assertions, testkit
    drain-window helper, `seriesKey`, etc.).
25. **BuildFlow dprint exit-14 upstream fix** (CHANGELOG-only commits).
26. **AGENTS deep slim-down** (structural; every addition trades a removal
    until then).
27. **Consumer-claim drift ritual posture** (USER GATE: keep exact pins w/
    ritual vs drop them).
28. **`/mnt/buildcache` topology answer** (owner; recurrence risk for the
    2026-09-29 corruption class).

(Items 13–28 are standing TODO_LIST entries restated for completeness — the
authoritative open list lives there, not here.)

## g) Questions I can NOT figure out myself

1. **Where does the pma auto-commit daemon live — unit name, config path, or a
   readable log dir?** Leak #4 (`620a3b1`, 09:53:48) landed DURING this
   session, minutes before I ran the guards; the rewind (`bb47fbc`) is still
   unexplained. From agent shells: no process, no `~/.config/pma`,
   `systemctl` blocked. Its config/logs are the only path to the root-cause.
2. **The `.buildflow.yml` license-check skip removal (not authored by me) —
   deliberate or artifact?** AGENTS documents license-check as skipped because
   go-licenses embeds a pre-1.27 go and cannot load this repo. Did BuildFlow
   upstream fix the toolchain (→ keep the removal), or should the skip be
   restored? It shipped inside daemon heuristic commits `620a3b1`/`31526ce`.
3. **Ratify ci-preflight's posture**: master runs now fail ONE loud job by
   design until the secret exists (matrix skipped, guards green). Acceptable,
   or should preflight warn instead of fail (with a separate alert)? And is
   there an ETA for the `SSH_PRIVATE_KEY` restore?

---

_Format note: written as `.md` at the user's explicit path/format demand —
status-report skill default is HTML; explicit instruction wins. Session
commits (via daemon heuristic batches): `19100c5` (go.work leak-4 fix +
integration pin bump + TODO edits), `31526ce` (FEATURES + dprint + buildflow
yaml + structure-linter yaml); post-report fixes (AGENTS parenthetical, claim
wording) ride the next daemon batch. All four guards green at report time._
