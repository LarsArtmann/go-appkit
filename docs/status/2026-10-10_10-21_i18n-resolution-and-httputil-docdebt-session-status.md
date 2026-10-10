# Status: i18n Resolution + httputil Doc-Debt Pass (session 2)

**2026-10-10 10:21 CEST · Session scope:** since the 01:43 report — assessed the
httputil 22-32 research report, wrote the (immediately obsoleted) decision
memo, resolved the split-brain across three go-appkit artifacts, then executed
the pre-tag doc-debt batch in httputil (their f2–f6). **This report covers
only this session's run.**

---

## Stat cards

| Done | Partial | Not started | Fucked up |
| ---- | ------- | ----------- | --------- |
| 7    | 2       | 4           | 1 (self-caught, cost: one stale recommendation) |

---

## a) FULLY DONE

1. **22-32 report assessed with verification** — ROADMAP §42 seed confirmed
   real and dated before opining; assessment delivered (their Vary/
   Content-Language cache-bug find, claim-verification discipline, my
   complementary content lane, the split-brain named).
2. **Split-brain annotated non-destructively** (go-appkit): planning doc
   appendix + TODO row pointer, both landed and daemon-committed.
3. **Decision memo written** (`doc/planning/2026-10-10_i18n-fleet-decision-memo.md`)
   merging both sessions' open questions into one rule-once checklist —
   content correct, timing wrong (see §d).
4. **08-11 train report assessed** — ROADMAP §42 GRADUATED/RULED verified;
   the `F07/F08` cross-reference chased to its source (their fleet sweep
   appended the verdict section to MY doc at 03:00 — provenance resolved,
   not assumed); all three go-appkit artifacts updated to RESOLVED
   (planning doc appendix, memo block, TODO row rewritten to residue-only).
5. **httputil doc-debt executed (their f2–f5):** FEATURES recount with
   methodology stated inline (58 examples 43/10/5 · 52 benchmarks / 62
   result rows · 30 fuzz 28/2 — independently counted, matched b1's
   quick-greps exactly); stale fuzz bullet "27 (25+2)" + CSRF enumeration
   fixed; README `LanguageSpecs()` sentence; CHANGELOG Documented bullet
   (RFC §5.2/§12.5.5 correction of record, CDN narrowing, 2 Non-goals,
   D-gates RULED, annotations, recount, nightly); x-text.md concrete
   pt/zh motivation pair.
6. **f6 (their "confirm or add"):** nightly-fuzz census — 25 steps, 30
   targets, `FuzzParseAcceptLanguage` **absent** → added (26/30); the 4
   still-unrotated targets enumerated; their TODO leftover (3) ticked;
   resolution addendum appended to their 08-11 report so no session re-does
   the batch.
7. **Verification:** `nix fmt` 0 changed · `doc-snippet-refs` all resolve ·
   workflow YAML parsed · tree snapshot before every edit (one parallel
   session landed mid-pass — zero collisions) · daemon committed all 7
   files (`0d90d5a`).

## b) PARTIALLY DONE

1. **zh/pt claims in x-text.md** — written from the 08-11 session's verified
   `PreferSameScript` pkg.go.dev fetch + the matchlang record, not
   re-verified first-hand this session. Internal doc, defensible, but their
   e5 lesson (second source before anything citable) applies if this guide
   is ever quoted externally.
2. **Fuzz budget question surfaced only implicitly** — I added the target at
   the uniform 300s; their f23's "consider a budget bump" (3 real bugs in
   45s argues for it) is in the addendum but was not put to the owner as an
   explicit decision item.

## c) NOT STARTED

1. **v1.6.0 tag/push** — owner-gated, not mine; everything downstream queues.
2. **Post-tag cascade** — T24 sperrmuell migration, go-appkit pin-drift +
   `Language()`-through-appkit composition E2E decision, dnsblockd F96,
   docs-site lane.
3. **Their g2/g3 rulings** (Vary end state; evidence bar) — owner-only.
4. **Annotation of my own 01:43 status report** — its f-list items were
   resolved by the fleet's rulings, but the report itself was never
   annotated (docs-health harvest discipline applied to everyone else's
   reports this session, not mine).

## d) TOTALLY FUCKED UP

**One real miss, self-caught one turn late:** I wrote the decision memo at
~08:45 — **38 minutes after** the 08-11 train report landed in the sibling's
`docs/status/` — without listing that directory first. I told the owner the
sperrmuell cache bug was "live" and needed a go/no-go, and recommended a
"rule-once" memo for questions that had already been ruled (all six D-gates,
D5 fix included). The owner had to hand me the 08-11 report for me to
discover my own artifact was obsolete. Mitigations: every memo recommendation
was independently confirmed by the rulings (zero harm to the end state), and
I corrected the record next turn. Root cause: I knew the fleet's status-dir
convention (I write these reports), knew the sibling was mid-execution, and
still synthesized cross-repo state from a 12-hour-old report + my own
inference instead of a 2-second `ls` of their status dir. **Rule adopted:
before any cross-repo synthesis or recommendation, `ls` the sibling's
`docs/status/` for anything newer than the artifact I'm reasoning from.**

Also found during mtime archaeology while writing this report: `date` ran
first this time (10:21) — last session's misdating lesson held.

## e) WHAT WE SHOULD IMPROVE

1. **Cross-session staleness check is now a hard pre-step** (see §d rule).
2. **Annotate my own reports when their items resolve elsewhere** — I
   applied docs-health annotation to two httputil reports and zero of mine.
3. **Parallel-writer snapshot discipline worked and should stay** — the
   go-appkit side had three sessions touching my artifacts (03:00 sweep
   append, daemon, me); every edit verified `count == 1` before landing.
4. **Recount-vs-copy:** the FEATURES numbers were independently recounted,
   not copied from b1's greps — matched exactly; that's the standard.
5. **Decision-surface timing:** a memo that consolidates OPEN questions
   should open with `ls`-verified "as of <artifact>, <time>" — my memo now
   carries a RESOLVED block, but a staleness header would have made the
   obsolescence visible to the next reader instantly.

## f) UP TO 50 THINGS TO GET DONE NEXT (brainstorm, ordered by gate)

**Gate: owner tag/push (1 decision, unblocks 2–9)**
1. Cut + push httputil v1.6.0 (coordinated per RELEASE.md; their g1 still
   unanswered — I will not touch tags without explicit words).
2. T24: sperrmuell migration onto `httputil.Language()` + extractor chain,
   keep `domain.Locale` mapped at one edge; run `verify-gates.sh`.
3. T70 trigger check (x/text Matcher swap — fires only on a third
   script-class locale).
4. go-appkit: `./scripts/check-pin-drift.sh` after the tag (integration's
   httputil leg sweeps to v1.6.0).
5. go-appkit: decide + optionally build the `Language()`-through-appkit
   composition E2E leg (composition-gap watchlist).
6. dnsblockd F96 i18n needs-check (15 min).
7. docs-site lane: sell `Language()` + `LanguageSpecs()`.
8. server_timing coordinated tag if versioned together.
9. Post-tag: collapse ROADMAP §42 to point at CHANGELOG only; pkg.go.dev
   render check of language.go's package comment.

**Gate: owner rulings (2 answers, unblocks 10–12)**
10. Vary end state: two-line documented vs Compression dedupe migration
    (their g2; touches v1.0-shipped behavior).
11. Evidence bar for live-web survey claims (their g3) → then the gin/echo/
    kaptinlin HEAD re-scan (their TODO leftover 1).
12. Fuzz budget for `FuzzParseAcceptLanguage` (uniform 300s now; 3 bugs in
    45s argues for more) — b)2's explicit decision item.

**httputil, no gate (bounded, next session)**
13. TODO leftover 2: sperrmuell-shape cookbook page (D5 worked example,
    hreflang caveat pointer).
14. TODO leftover 4: guard-test pattern advice section in
    docs/integrations/go-i18n.md.
15. TODO leftover 5: Language() misuse classes into the consumer-audit
    checklist post-v1.6.0.
16. c1: `buildflow --build-mode dev` expected-findings composition check.
17. c3: BenchmarkLanguage static-request harness + benchmarks.md rows.
18. c4: formatter×nolint verify-then-retire (one more fmt pass).
19. f18-style micro-Non-goals sweep (`Upgrade`, `Pragma`) — one-line
    records, prevents relitigation.
20. `ExampleLanguageSpecs` in httpspec if the examples audit says thin.
21. Coverage re-verify for language.go rows after any change.
22. `go doc`/pkg.go.dev rendering check post-tag.
23. art-dupl `-t 2` re-baseline on the release tree (fresh date in AGENTS).
24. DOMAIN_LANGUAGE.md glossary: negotiation/primary-subtag/extractor
    entries.
25. httpspec README: link the two new spec names beside the other families.
26. Verified-date stamp inside go-i18n.md if the re-scan confirms.
27. Nightly: corpus seeds review for the new target (6 curated entries).

**go-appkit, no gate**
28. Annotate the 01:43 status report (its f-items resolved fleet-side).
29. AGENTS.md cross-repo i18n pointer — only with a removal (361/377 cap).
30. Quarterly demand-gate re-check for the parked content lane
    (standing-ritual candidate).
31. errorpages `Config.Locale` seam — fires with real demand (parked, keep
    parked).
32. Confirm dprint/BuildFlow formatted the new planning docs without churn.

**Process (cheap, this week)**
33. Encode the §d staleness rule where future sessions trip over it
    (project AGENTS gotcha, cap tradeoff applies).
34. Same for the annotator lesson: `--emit-keys` verbatim (theirs, but the
    class generalizes: never re-derive tool output from memory).
35. Consider a fleet `docs/status/` cross-index or a session-start
    "sibling freshness" glance ritual.
36. When the 04-07 report's three watchers fire or die, annotate them.
37. Keep the two LSP-stale warnings on the known-noise list (CLI is
    authoritative).
38. `docs/status/README.md`: record that `.md`-at-explicit-demand reports
    are the standing dispatch shape (two more data points today).
39. `doc/planning/` index one-liner or explicit leave-as-is decision.
40. Re-run the fleet sweep after T24 lands (does sperrmuell's adoption
    change the appkit-lane N?).

**Watchers (deliberately bare, no action)**
41. `LanguageWhen` route-conditional variant.
42. Matched-vs-default context distinction.
43. `AbsentEncoding`-style empty-header policy doc.
44. Appkit content lane (Translator/guards/Alternates) — fires on real
    demand only; sketches preserved in the research doc.
45. Accept-Charset-style Non-goals for other dead headers (dup of 18 —
    pick one home).
46. Nightly rolling-issue policy: confirm the new target files correctly
    on failure (first red run is the test).
47. corpus: FuzzParseAcceptLanguage seeds from the three fixed bugs'
    repros.
48. Consider `Vary` dedupe as an httpspec spec assertion upgrade (strict
    comma-joined form) only after the g2 ruling.
49. Post-v1.6.0: fresh-consumer proxy check note for httputil consumers in
    go-appkit's integration legs.
50. Owner review + inline annotation of this report (docs-health
    convention).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Tag:** shall I cut + push v1.6.0 now (coordinated per RELEASE.md), or
   are you running the tag yourself? Everything in §f 2–9 waits on this.
2. **Evidence bar (their g3):** for the middleware-survey re-verification —
   is fetch-and-grep of upstream HEADs sufficient, or do you want pinned
   scratch-module builds before the claim is cited anywhere?
3. **Vary end state (their g2):** two `Vary` lines as the documented end
   state, or migrate Compression to the dedupe-aware single-line write in a
   follow-up (touches v1.0-shipped behavior)?

---

*Point-in-time snapshot — goes stale. Section f is docs-health HARVEST input.
`.md` at explicit path demand (standing dispatch shape).*
