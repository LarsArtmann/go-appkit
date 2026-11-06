# Status Report — SUPERB v12 Plan Rev 2 + Self-Review Corrections (session segment 01:25–01:33)

**Scope line.** This report covers the SECOND segment of the 2026-11-06 session: the
user-demanded expansion of the v12 plan HTML ("more detailed and just more superb"), and the
self-review pass that immediately followed it. It SUPERSEDES the population figure in the
01-25 report (its "55 tracked" was wrong — see d). **Verified this segment:** re-run of
`meta list --tags app --format json` (56 tracked — length-checked this time, diff-verified
identical key-set vs the first dump); strict direct-count recount for all 22 libs (clean
line-scan, per-app dedup); plan rev-2 HTML re-validated (python HTMLParser: zero parse errors,
zero external refs, all anchors resolve) after three content fixes; dprint behavior probed
single-file vs paired (discovery: repo dprint config has NO html plugin). **NOT verified:**
visual browser render of the plan (no chromium/google-chrome binary on this box; playwright
cache holds only ffmpeg — verified by attempt, so this is environmentally blocked, owner item
stands); everything listed c) below (not started, by design).

---

## a) FULLY DONE (this segment)

1. **Plan rev 2 assembled and validated** — `docs/planning/2026-11-06_01-00-SUPERB-v12-app-composition-preset-plan.html`,
   66.9KB → 102.8KB: dual-definition adoption chart (22 bars, direct solid / indirect
   translucent, appkit highlighted, all 44 hand-computed widths re-checked correct), fleet
   archetypes table, exemplar + root-cause issue cards with `file:line` evidence, acceptance
   criteria on ALL task tables, before/after Go API sketches (config + preset, token-highlighted),
   sequencing trains + risk register, extended sprint breakdown (T2.1 ×10 + T1.1 ×6),
   success-gate baseline→target table, richer D2 graph (archetypes + migration edges + gate
   subgraph), `callout-amber`→`callout-warning` fix (v1 class was unstyled in the template).
2. **modern-web-guidance applied** — `responsive-table` guide: dense tables wrapped in
   scrollable `role="region"` + `tabindex="0"` containers, sticky `<thead>`,
   `container-type: inline-size` (disclosed in the prior reply per skill contract).
3. **Errata discipline maintained** — audit errata extended twice, both same-day, both
   fix-forward: (1) full strict dual-column recount table replacing the partial spot-list;
   (2) denominator correction (58 = 56 tracked + 2 untracked, not 57).
4. **Three content defects in rev-2 caught by this self-review and fixed pre-report:**
   sloppy "— wait," mid-sentence prose in the ChastityAPI card (rewrote as clean eight-file
   breakdown); missing overlap disclosure on the archetypes table (counts sum > 39 because
   members overlap — now stated); wrong "57 app-tagged" denominator in the hero (→ 58 with
   tracked/untracked split).
5. **Post-fix re-validation** — HTML parse-clean, anchors resolve, no external refs, dprint
   clean on the markdown set; daemon picked up each batch (expected).

## b) PARTIALLY DONE

1. **Visual verification of the plan** — structurally validated three ways; pixel-render
   unverified and UNVERIFIABLE on this box (browser absent, proven by attempt).
2. **Archetype table** — overlap now disclosed, but rows still carry soft counts (~8/~10) and
   dual memberships; a single-membership primary-archetype matrix is the honest end-state.
3. **The 01-25 report** — correct as a snapshot except its population line (55→56 tracked);
   superseded by THIS report's scope line rather than rewritten (annotation discipline).
4. **Rev-2 prose fixes** — applied in place; v1 of the plan exists only in git history (see d).

## c) NOT STARTED (unchanged from 01-25; routing already correct)

The v12 build itself: `appkit/config` (T2), `appkit/preset` (T1), scaffold (T3), swap recipes
(T4), auth-bridge memo (T5.4), batteries D1/D3/C2 (T5) — all designed, routed in TODO_LIST,
zero code. Adoption-counter script (dual-column, mechanical success gate) — routed, not built.
v11 repair backlog — untouched, still holding the workstream slot pending the owner's
sequencing call.

## d) TOTALLY FUCKED UP

1. **The population denominator was wrong in THREE artifacts** — audit scope line, plan hero,
   and the 01-25 status report all said 55/57 where the truth is 56 tracked / 58 total. Root
   cause: I eyeballed the first JSON dump's key list and asserted a count I never ran
   (`len` executed for the first time this segment — 56). The per-app data set itself was
   complete and stable (diff-verified); the error was purely in the asserted denominator.
   Aggravating: 39+19=58 made the "57" internally inconsistent ON ITS FACE in two published
   docs, and I shipped it twice without doing the arithmetic. Fixed: audit errata addendum +
   plan hero + supersede line here.
2. **"dprint clean" was vacuously true for the HTML** — the paired `dprint check` that
   "passed" on the plan was carried by the .md file in the same invocation; the repo's dprint
   config has NO html plugin. Discovered when a single-file check exited 14 ("no files
   found"). My earlier clean claim was false assurance for exactly the file it named.
3. **Sloppy thinking-out-loud prose shipped** in rev-2 ("— wait, csrf is its own class:") —
   a draft artifact of my own reasoning loop pasted into a published card. Caught in
   self-review, rewritten.
4. **Rev-2 rewrote a point-in-time artifact in place.** Defensible (same session, same author,
   explicit user demand, footer self-documents "rev 2"), but v1 now lives only in git history
   — the annotate-never-rewrite rule was bent, not followed. Convention question to the owner
   (see g3).

## e) WHAT WE SHOULD IMPROVE

1. **Length-check every denominator at write time.** `len()` is one keystroke; this is the
   second claim-class defect this session (indirect/direct label, then 55/56/57/58). Process
   lesson candidate for crush-config `references/lessons.md` (cross-project class: asserted
   counts vs computed counts).
2. **Probe formatter coverage before trusting a green check** — a pass on a file the tool
   doesn't even scan is worse than no check (it launders false assurance).
3. **Single-membership archetype matrix** for the fleet (primary archetype per app) instead of
   overlapping soft counts.
4. **A headless-render path for HTML reports** — this box cannot browser-check them; either a
   nix-shell chromium devDependency or accept structural-only validation as the documented
   ceiling (owner call).

## f) Up to 50 next (◆ = new this segment; ★ = routed in TODO_LIST already; ⇄ = carried open from 01-25)

1. ★ T2.1–T2.4 `appkit/config` → v0.1.0.
2. ★ T1.1–T1.6 `appkit/preset` → v0.1.0.
3. ★ T3.1–T3.5 scaffold verdict → template → infra → docs skeleton → dogfood.
4. ★ T4.1–T4.5 swap-recipe template + dnsblockd/ChastityAPI/PapDashboard/pbx-lens recipes.
5. ★ T5.4 auth-bridge decision memo (owner gate).
6. ★ T5.1–T5.3 D1 worker supervisor, D3 sqlite+migrations, C2 realtime Mount.
7. ⇄ Adoption-counter script (dual-column) wired as mechanical success gate.
8. ⇄ Sequential-retry deep-reads: CV (13 modules) first, then KeyHolderAI, crm, webphone.
9. ⇄ Depth-audit the 22 config packages (read the 5 largest; firm up the class claim).
10. ⇄ Browser-render check for HTML reports (environmentally blocked here — needs the owner's
    devDependency decision, see e4).
11. ⇄ Preserve the sweep script in-repo (kills the /tmp dependency).
12. ◆ Process lesson → crush-config `references/lessons.md`: "asserted vs computed counts"
    (two same-day instances: require-semantics label, population denominator).
13. ◆ dprint html coverage decision: add an html/markup plugin to the repo config OR document
    html as out of dprint scope (kills the vacuous-pass class).
14. ◆ Single-membership archetype matrix as an audit appendix (primary archetype per app).
15. ◆ Add "rev 2 (expanded)" convention ruling to the planning-docs house rules (see g3).
16. ◆ Extend the plan's success-gate table with per-archetype targets (preset-hosts → 6/6 on
    preset; hand-rollers → ≤4 remaining) when the counter script lands.
17. ⇄ Fine-breakdown (≤15 min) for T1.2 spine wiring before the preset sprint.
18. ⇄ Consider a `meta`-side appkit-adoption tag/flag (one-liner success gate at the inventory
    layer).
19. ⇄ Owner sequencing: v11 repair backlog vs v12 spine (see g1 — still THE blocker).
20. ⇄ TODO_LIST preset-spine row: append the any/direct dual-number citation (it currently
    quotes any-require counts unqualified).

## g) Questions I cannot answer myself

1. **Sequencing (unchanged, now blocking two reports):** SUPERB v11 repair backlog vs SUPERB
   v12 preset spine — which runs first, or interleaved? (Config train T2 is small enough to
   interleave if you want both moving.)
2. **Auth bridge (unchanged):** cqrs-htmx usermgmt recipe for non-cqrs apps vs a security
   session battery — T5.4 decides the shape; the preset v0.x ships without auth either way.
3. **Plan-revision convention:** rev-2 rewrote the v1 plan in place (git history preserves v1).
   Going forward: is in-place revision under explicit user demand the accepted convention for
   same-session plan artifacts, or should revisions be new rev-dated files with supersede
   banners (the status-report pattern)?

---

_Point-in-time snapshot; supersedes the 01-25 report's population line only. Verdicts live
here; execution state lives in TODO_LIST. Index: `docs/status/README.md`._
