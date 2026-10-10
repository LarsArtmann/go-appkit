# i18n fleet decision memo — rule once, build twice

Date: 2026-10-10 · Status: AWAITING OWNER RULING · Supersedes: nothing; consolidates two parallel research sessions

Two sessions answered the same owner question independently on 2026-10-09
(22:32 httputil, 22:39 go-appkit). Their findings agree on the ecosystem and
diverge on exactly one architecture question. This memo merges both question
sets so the owner rules ONCE instead of across two repos.

- httputil session report:
  `~/projects/httputil/docs/status/2026-10-09_22-32_i18n-research-language-negotiation-session-status.md`
  (seeded httputil ROADMAP §Post-v1.0: `Language()` middleware)
- go-appkit research: `doc/planning/2026-10-09_i18n-module-research.md`
  (appendix: split-brain annotation; TODO_LIST P3 row)

## THE fork (their g1/f1 = go-appkit Q1)

**Where does HTTP language negotiation live?**

| | Proposal A (RECOMMENDED): httputil `Language()` | Proposal B: go-appkit `/i18n` module |
| --- | --- | --- |
| Scope | HTTP edge only: extractor chain, matching, ctx, `Content-Language`/`Vary` | Everything: negotiation + Translator + guards + hreflang |
| Dep posture | zero-dep primary-subtag default; x/text via `TagMatcher` plugin | x/text Matcher in-dep from day one |
| Precedents | KeyExtractor/Nonce/q-parser/Vary — same-repo patterns | realtime/otel-style satellite module |
| Consumers | ALL 45 httputil consumers | appkit family only |
| Weakness | primary-subtag default mishandles pt/zh script classes (plugin mitigates) | would duplicate httputil's middleware for appkit consumers |

**Recommendation: A, with B shrunk to the content lane.** Negotiation is a
middleware concern; httputil owns middleware; appkit composes httputil
(otel/flightrecorder already depend on it). `appkit/i18n` v0.1.0 keeps ONLY:
`Translator` (go-i18n, visible fallback chain), `MissingIn`/`AllKeys` guard
helpers, `Alternates()`/`Prefix` hreflang helpers, `errorpages.Config.Locale`
seam — it consumes a `language.Tag` from wherever (httputil ctx once shipped,
T49-style middleware until then). Nothing is lost from either sketch; the
negotiation half of the go-appkit sketch is retired.

## Sub-rulings (tick in one pass)

1. **Vary default** (their g2): `Vary: Accept-Language` emitted whenever
   `Language()` runs (RFC-strict, RECOMMENDED) vs only when the header leg
   influenced output (minimalist). Same lifecycle class as their open
   `AbsentEncodingFirstConfigured` question — rule both together.
2. **Matcher posture**: accept the primary-subtag default + make
   `docs/integrations/x-text.md` adapter MANDATORY reading in the middleware
   docs (load-bearing, not optional — the pt/zh bug class is real), or
   promote x/text into httputil core (violates their dep policy; rejected).
3. **Build timing**: httputil design note NOW (their f6-f15, cheap), implementation
   after/with their v1.0; appkit content lane can ship independently ANY time
   (one session, N=1 proven by sperrmuell). Sequenced or parallel — owner's
   appetite call.
4. **sperrmuell cache-bug fix authority** (their g3/f32): add
   `Vary: Accept-Language` + `Content-Language` to sperrmuell's
   header-negotiated `/api/*` responses NOW — live correctness bug,
   independent of all rulings above. 20-minute fix in their repo with
   `verify-gates.sh`.

## Downstream (already routed, no new decisions)

- sperrmuell Matcher upgrade via `TagMatcher` plugin rides httputil adoption
  (their f33); guard-test pattern (visible missing translation) extracted as
  integrations-doc advice (their f34 / go-appkit `MissingIn`).
- errorpages `Config.Locale` seam: additive train, independent (go-appkit
  TODO row).
- Fleet sweep for a second bilingual consumer (go-appkit f-item 1) still
  settles the N=1 demand count — cheap, do before or with the appkit lane.

## Rule-once checklist

- [ ] Fork: A (httputil edge + appkit content lane) or B (appkit module owns all)
- [ ] Vary default: strict vs minimalist
- [ ] x/text adapter doc mandatory: yes/no
- [ ] Build order/timing
- [ ] sperrmuell fix: go / no-go
