# TODO-Whole-List Execution — Guards Wave, the Drain Rule, Corpus Zero + Pin 0

> **ARCHIVED 2026-11-07 00:25 — fully executed at write time.** Every §f
> item 1-16 carries its inline done-verdict; items 17-24 are routed to the
> standing owner-gated rows (8 enforced routes, routed guard 0 dead).
> Live remainder: the owner gates only — SSH secret, floor-train standalone
> call, machine ops, SKILLS push, the cqrs-lint memo upstream filing.

**Date:** 2026-11-06 23:05 → 2026-11-07 00:25 CET
**Trigger:** the owner mandate to execute the ENTIRE routed TODO list from the
wave-3 close-out (the plan `docs/planning/archived/2026-11-06_22-05_SUPERB-corpus-wave3-and-gate-ratchet.md`
was executed; its remaining tiers were the routed TODO rows).
**Scope:** every actionable non-owner-gated row: the eleven guard/tooling asks,
check-touched-modules, ci-preflight visibility, cqrs-lint debts, integration
hygiene, the watchlist split, the HTML trio, errorpages godoc, the annotate
drain, the long-tail F-78/79/80, the go-health forward-compat scratch, the
consumer heads-ups, the upstream docs-gates notes. Owner-gated rows (SSH
secret, floor-train standalone, machine ops, pma decisions, plan-revision
ruling) explicitly untouched.

## a) GOALS

Execute the whole actionable TODO list with guards green between every move,
commit in logical groups, push, verify CI on the tip.

## b) DONE — code and guards (the 1%→51%)

1. **Six standing guards shipped + wired (all three venues per the M14
   matrix where applicable):** `check-module-licenses.sh` (13 released
   modules; the integration exemption is ASSERTED into existence — the
   guard went born-red until `integration/doc.go` stated the deliberate
   absence, closing issue54 §f35 in passing); `check-pin-drift.sh` claims
   parser STRICT with named offenders (the aggregate-count class died; the
   rewrite ALSO caught a live bug — the printf-missing-newline dropped the
   LAST claim, preset, from every previous run silently) + the
   release-window rule (local-tag-not-on-origin = loud NOTE, not drift;
   origin unreachable degrades gracefully); `check-gosum-canonical.sh`
   defers family pins on local-only tags (probe-proven with a throwaway
   `preset/v9.9.9` tag — the DEFERRED line names the tag); `check-tagged-changelogs.sh`
   (tag's own tree must carry `## [version] - <tagdate>`; born-red on the
   KNOWN config/v0.1.0 RE-CUT REVERTED artifact, green on all thirteen
   other recent tags; wired into pre-tag-checks as step 3);
   `check-features-claims.sh` (Shipped-through == newest tag; drift probe
   proven); `check-touched-modules.sh` (staged-set hermetic build+vet,
   wired FATAL into pre-commit — ran live on its own commit).
2. **The release dance scripted:** `scripts/tag-module.sh` — dirty-tree
   refusal, exactly-once annotated tags, the full pre-tag gate, push,
   `--tidy-integration` (the insteadOf trick from AGENTS note 4).
   Probe-proven end-to-end: the gate correctly REFUSED a throwaway tag
   whose CHANGELOG was not dated.
3. **The drain rule (issue54 §e5 root-caused once):** `scripts/check_response_drain.go`
   — a go/parser AST walker (linter-building decision tree: syntax rule →
   AST walker, no type facts, no plugin) flagging `X.Body.Close()` with no
   read of `X.Body` in the enclosing function. Trust engineering on the
   way in: deferred closures attribute to the enclosing function; NO
   helper-name trust (verified at source: EVERY fleet helper — root
   httpGet, flightrecorder httpGetURL, preset getWithHeaders, testkit
   HTTPGetContext — returns the RAW response; the caller owns the body);
   `// respdrain:ignore <reason>` suppression. Fleet fix: **44 sites
   drained**, **10 SSE carve-outs suppressed with reasons** — and the
   first codemod pass DRAINED OPEN SSE STREAMS and hung the realtime suite
   (io.Copy blocks on an endless stream); that failure is why the
   SSE carve-out is explicit per-site, not a silent rule. Discrimination
   fixture green; wrapper `check-response-drain.sh` wired into pre-commit
   fatal set + pre-tag-checks + CI.
4. **CI:** `flake.yml` nightly `-count=5` integration leg (cron + manual
   dispatch, actionlint green); ci-preflight carries a designed-red
   annotation + job name (16-13 item 30); cqrs-lint CI PINNED at upstream
   stable `v4.21.0` (verified green on the cqrs tree before pinning; the
   `@latest` retro-red class dies) with the bump ritual documented at the
   job.

## c) DONE — docs and corpus (the 4%→64%)

5. **TODO_LIST hygiene:** nine rows CLOSED with evidence (guard batch,
   JSON-v1 STALE verdict — both flagged files already import
   `encoding/json/v2`, verified against the v2 godoc surface, vet clean;
   integration hygiene — the doc-comment is already detached; touched-modules;
   preflight visibility; watchlist split; errorpages godoc; HTML trio;
   cqrs-lint debts), the composition-gap watchlist row rewritten as the
   compact pinned-compositions table, the upstream watchlist records the
   **go-health v0.10.0 + dashboard v0.13.2 forward-compat scratch GREEN**
   (`NewProbe` returns a `*gohealth.Probe` under MVS-forced newer deps —
   isolated-cache scratch via the new `scripts/verify-consumer.sh`
   mechanism; the next health train is a decision, not a discovery).
6. **Annotate corpus drained 243 → 0, and the ratchet CLOSED:**
   `annotate_max` 255 → **0** in `.docs-gates.json` (the designed
   drain-then-lower; the drain: 236 G/uniform per-cell strike
   normalizations across the four 2026-06/07 reports, the nine
   D/verdict-leads-open items rewritten to complete marker-bearing
   verdicts — learned live: the marker vocabulary is strict (`done`,
   `Won't implement`), the verdict rides AFTER the strike, struck cells
   need markers, and table-cell inline code must not contain raw pipes).
   Battery 5/5 PASS at pin 0.
7. **Micro-fixes:** `integration/doc.go` LICENSE-absence statement; health
   README gained the role=region note and LOST the retired
   `GOEXPERIMENT=jsonv2` claim; module-details frh line corrected to
   go-health v0.8.1 (the pin and MVS agree fleet-wide); the
   recorder-sentinel outgoing draft's Status line now says RESOLVED; the
   HTML trio titles fixed (the 09-04 placeholder `<title>` and the 08-15
   MISSING `<title>`); keep-as-reference ruling recorded; errorpages
   `Mount` godoc carries the catch-all 405 clarification (publishes with
   the next errorpages tag); `docs/recipes/docs-gates-cadence-and-anchors.md`
   (the cadence override rationale + the TODO-row-N anchor sweep verdict:
   living docs clean, point-in-time docs stay).

## d) DONE — filings (verified before posting)

8. **Consumer heads-ups POSTED — with a concurrency lesson:** cqrs-htmx
   heads-up + PapDashboard heads-up, voice check 0 FAIL/0 WARN, claims
   verified at source (stack order `middleware.go:11`, `MeasureServerTiming`
   in httputil/server_timing v1.0.3, preset grammar). THEN discovered a
   concurrent session had posted the SAME notices minutes earlier
   (cqrs-htmx#43, PapDashboard#6) — **my issues #44 / #7 closed as
   duplicates of #43/#6 immediately** (terse comment naming the dup).
   LESSON for the next session: `gh issue list` the target repos BEFORE
   posting anything a concurrent session might also be posting.
9. **The three upstream docs-gates notes** written into the SKILLS repo
   (`docs-health/references/upstream-notes.md`, committed locally,
   unpushed — the crush-config precedent): routed≠resolved gate-surface
   gap, the compound-block positional-renumbering quirk (REPRODUCED live
   with a fixture before writing), the DF-boundary `--explain-df`
   observability ask (verified the flag does not exist).

## e) WHAT WE SHOULD IMPROVE

1. **Concurrent-session GitHub races are now REAL** (the #43/#44 double-post):
   the docs corpus had guards; GitHub filings did not. A pre-post
   `gh issue list --search` check should join the filing procedure.
2. `verify-consumer.sh`'s scratch go.mod floor is read from go.work — a
   consumer simulating an OLDER floor cannot be expressed; fine until one
   is needed.
3. The pre-commit fatal set now runs nine guards + touched-modules build+vet;
   commit-time cost is still seconds, but the set should be re-audited if
   anything network-bearing ever wants in (venue matrix says no).
4. `check_response_drain.go` is a build-tagged main under `scripts/` — the
   fleet could promote it into testkit as a vet-analyzable checker if a
   second repo wants it; premature today.

## f) UP TO 50 THINGS TO GET DONE NEXT

1. ~~Execute the eleven guard/tooling asks.~~ done — all eleven shipped + wired (row closed with evidence).
2. ~~The AST rule for body-closed-without-read.~~ done — the walker + fleet fix + three-venue wiring (§b-3).
3. ~~The annotate drain + pin lower.~~ done — 243 → 0, pin 0 (§c-6).
4. ~~cqrs-lint CI pin + F032 memo.~~ done — pinned v4.21.0 + the memo filed in-repo (filing upstream is the owner's maintainer act).
5. ~~Integration JSON-v1 boundary.~~ done — STALE verdict documented (both files on v2 APIs).
6. ~~integration security_realtime_test.go header decision.~~ done — already detached; keep.
7. ~~Watchlist split.~~ done — the compact table replaced the prose monster.
8. ~~HTML trio ruling + title fixes.~~ done — keep-as-reference; both titles fixed.
9. ~~errorpages godoc 405 clarification.~~ done — in-tree; ships with the next errorpages tag.
10. ~~F-78 dprint round-trip on fresh:skip files.~~ done — 5/5 markers survive, check stable.
11. ~~F-79 cadence rationale doc.~~ done — `docs/recipes/docs-gates-cadence-and-anchors.md`.
12. ~~F-80 TODO-row-N anchor sweep.~~ done — living docs clean; point-in-time docs stay (verdict in the recipe doc).
13. ~~go-health forward-compat scratch.~~ done — GREEN under MVS-forced v0.10.0/v0.13.2 (evidence in the upstream watchlist row).
14. ~~Consumer heads-ups.~~ done — posted; dup incident handled (§d-8); the row was closed by the concurrent session for #43/#6.
15. ~~Upstream docs-gates notes.~~ done — committed in the SKILLS repo, unpushed (owner pushes).
16. ~~TODO hygiene pass (nine row closures + table).~~ done — §c-5.
17. Floor-train go 1.27.2 → 1.27.3 — **→ TODO P2 — Floor-train (rides the next release train per the member-go.mod ritual; owner-gated standalone) `<!-- fresh:skip: routed in TODO_LIST -->`**
18. Machine ops batch — **→ TODO P3 — Machine ops (owner present required) `<!-- fresh:skip: routed in TODO_LIST -->`**
19. SSH secret restore — **→ TODO P1 — the standing owner row `<!-- fresh:skip: routed in TODO_LIST -->`**
20. Plan-revision convention ruling — **→ TODO P2 — micro-backlog (owner ruling) `<!-- fresh:skip: routed in TODO_LIST -->`**
21. pma daemon §6 decisions — **→ TODO P2 — pma row (owner checklist) `<!-- fresh:skip: routed in TODO_LIST -->`**
22. preset health-pin v0.3.1 ride — **→ TODO P2 — preset spine row (rides preset's next tag) `<!-- fresh:skip: routed in TODO_LIST -->`**
23. SKILLS repo push (the upstream notes commit is local) — **→ TODO P3 — Machine ops (push verification, same class as the crush-config 20587f0 row) `<!-- fresh:skip: routed in TODO_LIST -->`**
24. The micro-backlog stragglers ( Mounted.setStarted verified this session; per-lib errata; outgoing drafts sweep) — **→ TODO P2 — micro-backlog `<!-- fresh:skip: routed in TODO_LIST -->`**

## g) OPEN QUESTIONS AWAITING USER DECISIONS

1. Floor-train standalone vs ride (GO-2026-6809 is stdlib-only; the row
   documents both paths).
2. Push the SKILLS repo (the notes commit is local-only).
3. The cqrs-lint F032 memo upstream filing (maintainer act on go-cqrs-lite).

## Verify (run as written)

```console
cd /home/lars/projects/go-appkit
docs-gates run .                                   # 5/5 PASS at annotate pin 0
./scripts/check-module-licenses.sh                 # exit 0
./scripts/check-features-claims.sh                 # exit 0
./scripts/check-tagged-changelogs.sh v0.10.0 health/v0.3.1 preset/v0.1.0   # exit 0
./scripts/check-touched-modules.sh --all           # exit 0, 13 modules
./scripts/check-response-drain.sh                  # exit 0
nix shell nixpkgs#actionlint -c actionlint .github/workflows/ci.yml .github/workflows/flake.yml   # exit 0
./scripts/check-routed-destinations.sh             # exit 0
./scripts/check-status-index.sh                    # exit 0
```
