# Status: i18n Research Session (go-appkit)

**2026-10-10 01:43 CEST · Session scope:** "How do we help with i18n and co?" —
research from `~/projects/sperrmuell-direct` + web, design for composability,
record findings. **This report covers ONLY this session's run.**

Sources of truth for this snapshot: session tool calls + file mtimes
(planning doc written **2026-10-09 22:39**, TODO row 22:40 — NOT 10-08 as
first labeled; corrected mid-report, see §d).

---

## Stat cards

| Done | Partial | Not started | Fucked up              |
| ---- | ------- | ----------- | ---------------------- |
| 6    | 2       | 3           | 2 (1 fixed in-session) |

---

## a) FULLY DONE

1. **READ — sperrmuell-direct T49 pattern inventory** (complete): `internal/i18n/i18n.go`
   (go-i18n v2 bundle, embedded JSON, init load, T/S/TData/SData/TPlural,
   visible fallback chain locale → default → key + slog), `internal/domain/locale.go`
   (closed `Locale uint8` enum, parse/Prefix/Other/JSON wire), `cmd/server/locale.go`
   (`withLocale` middleware: prefix > Accept-Language > default, slug→mux rewrite,
   308 wrong-locale, publicPath ctx for switcher), hreflang/x-default/sitemap
   (`routes.go`, `ui.go`), guard suite + `scripts/gen-i18n-keys.py`
   (idempotent keys.go generator, fails loudly on drift), pins (go-i18n
   v2.6.1, x/text v0.42.0).
2. **UNDERSTOOD — go-appkit family surfaces**: `errorpages.Config.Lang` is a
   static-string seam (no per-request locale); templ-components already has
   `Base.Locale` prop + `SEOAlternate` hreflang type + locale-aware
   relative-time (`documentElement.lang` → `Intl.RelativeTimeFormat`);
   batteries spec + TODO_LIST had ZERO i18n entries (greenfield confirmed).
3. **RESEARCHED — ecosystem** (web, cited): go-i18n v2.6.1 active/MIT,
   CLDR plurals; `x/text/language.Matcher` canonical for Accept-Language
   (CLDR mutual-intelligibility: pt-BR/pt-PT, zh-Hans/zh-Hant, sr scripts —
   exactly what T49's hand-rolled parser gets wrong); x/text
   message/currency for numbers; **no `x/text/date` exists**; alternatives
   (gotext/spreak = gettext, kaptinlin = ICU — watcher only); SEO: path
   prefix + reciprocal hreflang + x-default + 308s + never
   Accept-Language-redirect (Googlebot is en-US).
4. **REFLECTED/DESIGNED — module sketch**: 13th opt-in module `/i18n`
   (`Server` matcher middleware, `Negotiate` one-shot, `Translator` over
   go-i18n, `MissingIn` guard helpers, `Prefix`/`Alternates`), API speaks
   `language.Tag` (no closed enum), no core dep, no slug router, no
   auto-redirect; anti-recommendations + trigger conditions written down.
5. **WRITTEN — `doc/planning/2026-10-09_i18n-module-research.md`** (138
   lines): pattern inventory table, ecosystem verdict, API sketch, routed
   adjacent surfaces ("and co"), recommendation.
6. **WRITTEN — TODO_LIST P3 demand-gated row** (line 58) with TRIGGER,
   sizing (S/M, one session), and the sibling `errorpages.Config.Locale`
   train.

## b) PARTIALLY DONE

1. **Demand analysis ("N=1 consumer")** — claimed but NOT verified: I never
   grepped cv, cqrs-htmx, PapDashboard, go-aichat, or templ-components'
   website for existing i18n/locale usage. The demand-gate framing in the
   planning doc + TODO row rests on an unchecked count.
2. **sperrmuell `routes.go` reading** — `RouteID`/`RoutePath`/route registry
   read only via grep hits; the "slug translation stays app-owned, port as
   cookbook" claim leans on `locale.go` (read fully) + partial `routes.go`.

## c) NOT STARTED

1. **The `i18n` module itself** — deliberately: P3 demand gate (N=1 today).
2. **`errorpages.Config.Locale` seam** — routed as additive sibling train,
   not started.
3. **sperrmuell-direct Matcher upgrade** — the research's most directly
   actionable payoff (replace their hand-rolled `ParseAcceptLanguage` with
   `x/text` matcher for pt/zh/script classes) is recorded nowhere but this
   report — needs routing to THEIR TODO_LIST.

## d) TOTALLY FUCKED UP

1. **Misdated artifacts (fixed in-session, root cause named):** planning doc
   - TODO row were labeled `2026-10-08` because I trusted the conversation
     env header instead of running `date`. mtimes prove **2026-10-09 22:39**.
     Corrected this turn: file renamed to `2026-10-09_i18n-module-research.md`,
     internal `Date:` fixed, TODO row date + doc path fixed (remaining
     `2026-10-08` in the doc refers to sperrmuell's T49 ship date — correct).
     Root cause: an unverified external claim (env date) encoded into artifact
     names — exactly the `verify-external-claims` failure mode. Rule going
     forward: `date` before any timestamped filename, always.
2. **Unverified fact published as fact:** "N=1 bilingual consumer" in a
   shipped planning doc + TODO row without a fleet sweep (see b1). Not
   corrected by the misdate fix — needs the grep sweep before anyone acts
   on the demand gate.

## e) WHAT WE SHOULD IMPROVE

1. ~~**Timestamped artifacts: run `date` first, never the env header** (proven~~ **Won't implement — lesson held — the 10-21 session ran date first (its §d); not AGENTS-encoded (cap discipline).**
   ~~2 days off this session).~~
2. ~~**Demand claims need a fleet sweep** before gating work on them — one~~ done — executed — the 03:00 fleet sweep settled N=4
   ~~`rg -il "i18n|locale"` across sibling projects, ~2 minutes, was skipped.~~
3. **Read-coverage honesty**: sketches that cite consumer code as
   "cookbook source" should list which files were read fully vs skimmed
   (b2 would have been visible immediately).
4. ~~**Ecosystem version claims** (go-i18n "v2.6.1 latest", "active") came~~ **NOT-DO — moot — module never fired; no pinning happened.**
   ~~from one agentic fetch — fine for research, but re-verify against the~~
   ~~module proxy before pinning at module-build time.~~
5. ~~**Cross-repo findings need a routing decision at discovery time** — the~~ done — routed — the Matcher upgrade became the T24 sperrmuell migration (10-21 f2)
   ~~sperrmuell Matcher upgrade surfaced mid-research and ended nowhere~~
   ~~actionable (c3).~~

## f) UP TO 50 THINGS TO GET DONE NEXT (brainstorm, sorted by impact)

**Verify the research's own claims**

1. ~~Fleet sweep for i18n/locale demand: grep cv, cqrs-htmx, PapDashboard,~~ done — N=4 (03:00 fleet sweep; TODO row 55 trail)
   ~~go-aichat, templ-components website, docs module — settle N=1 or N=k.~~
2. ~~Verify go-i18n + x/text latest tags against the module proxy (pre-pin).~~ done — httputil language.go ships on x/text (v1.6.0 Unreleased)
3. ~~Read sperrmuell `routes.go` fully (RouteID/RoutePath registry) so the~~ done — via the 08-11/22-32 sibling sessions (D5 worked example)
   ~~cookbook claim is backed by full-coverage reading.~~
4. ~~Verify templ-components `SEOAlternate` contract in full (Lang/URL,~~ **Won't implement — moot — content lane parked pending demand.**
   ~~absolute-URL requirement) before the module maps to it.~~
5. ~~Confirm `x/text` v0.42.0 floor works at the family's go 1.27.1 (fresh~~ done — httputil v1.6.0 prerelease-check 9/9 at the family floor
   ~~scratch-module compile).~~

**Route the cross-repo payoff**
6. ~~Write the sperrmuell-direct Matcher-upgrade item into THEIR TODO_LIST~~ done — superseded by the T24 sperrmuell migration (10-21 f2)
~~(pending Lars's answer to Q2 below).~~
7. ~~Backport decision: does sperrmuell adopt appkit's module later (reverse~~ done — ruled — decision memo 2026-10-10 (negotiation lives once in httputil)
~~direction), or keep its internal package and only share the recipe?~~

**When the i18n trigger fires — module build order (each = bounded)**
8. ~~`/i18n` scaffold: go.mod (1.27.1), doc.go, LICENSE, README, .golangci.yml~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~(test-exclusion union), go.work entry + workspace-charter guard.~~
9. ~~`NewServer(supported, default, opts...)` — cached `language.Matcher`.~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
10. ~~`Middleware`: prefix-strip (opt) > cookie (opt) > Accept-Language >~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~default; ctx injection; public-path preservation for switchers.~~
11. ~~`WithContext`/`FromContext` context helpers (nil-safe like frh).~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
12. ~~`Negotiate(header, supported, fallback)` one-shot for JSON APIs.~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
13. ~~Cookie semantics decision: `?lang=`→cookie set, in v0.1 or deferred.~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
14. ~~`NewTranslator(fs.FS, default, opts...)` — JSON unmarshal default,~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~embed-friendly.~~
15. ~~`T`/`S`/`Plural` with visible fallback chain + slog warnings~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~(T49 semantics, preserved).~~
16. ~~`MissingIn(tag)`/`AllKeys()` guard helpers (parity tests as library).~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
17. ~~`Prefix(tag, canonical)` + `Alternate`/`Alternates(base, current,~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
    ~~supported)` incl. x-default.~~
18. ~~Negotiation table tests: de-AT, pt-BR vs pt-PT, zh-TW, sr-Latn,~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~q-values, ties, garbage headers, empty header.~~
19. ~~Fallback-chain tests: missing-in-locale warn path, missing-everywhere~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~error path, key-as-last-resort visibility.~~
20. ~~`example/` bilingual demo (embed FS, /en prefix, switcher, plural keys).~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
21. ~~README cookbook: slug translation + 308s + hreflang/sitemap recipe,~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~ported from sperrmuell routes.go (documented, not built).~~
22. ~~`integration/` pin + composition E2E leg (testkit.Serve, 1ms drain).~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
23. ~~Fresh-consumer proxy check + tag v0.1.0 (release ritual, additions-only~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~n/a — new module).~~
24. ~~AGENTS.md entry (observes the 361/377 line cap: one removal per~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~addition) + Release State wave line.~~
25. ~~`cqrs`-style scorecard equivalent: none exists — skip deliberately,~~ **NOT-DO — module ruled NOT-FIRE 2026-10-10 — sketches preserved in the research doc RESOLVED appendix.**
~~note in CHANGELOG.~~

**Adjacent seams ("and co")**
26. `errorpages.Config.Locale func(*http.Request) language.Tag` (default: **→ parked pending real demand (TODO row 55)**
i18n.FromContext when present) — additive, own train.
27. templ-components train: translated default 404/405 copy (de/en props)
or document English-only stance — owner decision.
28. templ-components: doc the relative-time ↔ `documentElement.lang` contract
next to `Base.Locale`.
29. ~~Date localization decision note: no x/text/date exists — per-locale~~ **NOT-DO — moot — module parked.**
~~layouts server-side or client Intl; write it into the module README~~
~~when built.~~
30. ~~v0.2 candidate: `message.Printer`/currency helpers in the module.~~ **NOT-DO — moot — module parked.**
31. Watcher: kaptinlin/go-i18n (full ICU MessageFormat) — re-check on any
consumer demand for gender/select.

**Session-hygiene / process debt surfaced this session**
32. ~~Encode the `date`-first rule where future sessions trip over it~~ **Won't implement — lesson held (10-21 §d ran date first); not AGENTS-encoded (cap discipline).**
~~(project AGENTS gotcha candidate — line-cap tradeoff).~~
33. ~~AGENTS.md cross-repo context pointer for the i18n research (only with a~~ **Won't implement — TODO row 55 is the tracker (per the 2026-10-12 slim-down).**
~~removal — cap discipline).~~
34. ~~Batteries-spec routing decision: i18n is NOT CV-derived — does it get a~~ **NOT-DO — TODO row stays the sole tracker — module never fired.**
~~cluster ID there, or does the TODO row remain the sole tracker?~~
35. ~~docs-health HARVEST: fold §f items 1-7 + 26-31 into TODO_LIST/ROADMAP~~ done — TODO row rewritten to residue-only (10-21 a4)
~~(this report is the input; do not entomb).~~
36. ~~Annotate the planning doc if the fleet sweep changes the demand gate~~ done — planning doc carries the RESOLVED appendix (10-21 a2/a4)
~~(docs-health VERIFY/ANNOTATE, non-destructive).~~

**Sperrmuell-side improvements noticed (their repo, route via Q2)**
37. ~~Replace `i18n.ParseAcceptLanguage` primary-subtag cut with~~ done — superseded by the T24 sperrmuell migration (10-21 f2)
~~`language.NewMatcher` (correctness: pt/zh/script classes).~~
38. ~~Their `ParseLocale` silently ignores region subtags — Matcher would~~ done — superseded by the T24 sperrmuell migration (10-21 f2)
~~also fix `zh-Hans`/`zh-Hant` distinction they currently collapse.~~
39. ~~Guard suite extension: q-value tie-breaking test for their parser~~ done — superseded by the T24 migration verify-gates (10-21 f2)
~~(order-stable) survives the Matcher swap.~~

**Cheap hygiene in THIS repo (noticed while working)**
40. ~~`docs/status/README.md` — confirm the `.md`-at-explicit-demand~~ done — convention line added to docs/status/README.md (2026-10-12)
~~convention is recorded there (this report is another data point).~~
41. ~~`doc/planning/` has no index/README — 20+ dated docs, discoverability~~ done — doc/planning/README.md exists
~~by grep only (candidate one-liner index or leave as-is, decide once).~~
42. ~~TODO_LIST row 58 wording: "13th opt-in module" — verify the count when~~ **NOT-DO — moot — 12 modules; i18n never fired.**
~~it lands (12 today; systemd pushed it to 12? recount at build time).~~
43. ~~When the module lands: `scripts/check-workspace-charter.sh` forces the~~ **NOT-DO — moot — module never built.**
~~go.work entry — expected, budget for it.~~
44. ~~Verify dprint/BuildFlow formats the two new md files without churn~~ done — dprint 2026-10-12 recheck green
~~(dprint.json excludes CHANGELOG only — planning docs ARE formatted).~~

**Later / speculative (ROADMAP fuel, not commitments)**
45. ~~`testkit` negotiation probe helper (assert ctx tag after middleware).~~ **NOT-DO — moot — module parked.**
46. ~~otel: locale as a span attribute? (privacy-sensitive — default NO,~~ **NOT-DO — not-do — privacy-sensitive, default NO (the item s own stance).**
~~document why if ever asked).~~
47. ~~security: `?key=` API-key auth and locale negotiation interplay on~~ **NOT-DO — moot — no cookbook; module parked.**
~~technical paths (middleware-order note for the cookbook).~~
48. ~~realtime SSE + locale: per-connection locale for server-pushed events~~ **NOT-DO — moot — no consumer.**
~~(only if a consumer pushes localized events — none today).~~
49. ~~docs module (catalog): locale-aware doc rendering — YAGNI until asked.~~ **NOT-DO — YAGNI — the item s own stance.**
50. Re-run the demand gate quarterly while P3 (standing-ritual candidate). **→ folded into the TODO i18n residue row (standing while P3)**

---

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

**Q1 — Build the `i18n` module NOW or honor the demand gate?** ~~The sketch is
one session of work; the gate (N=1, unverified — item 1 would settle it) says
wait. Your call: extract now while T49 is fresh, or hold P3?~~ **RULED
2026-10-10: the module does NOT fire — negotiation lives ONCE in httputil
`Language()` (v1.6.0); appkit content lane parked pending real demand
(decision memo + TODO row).**

**Q2 — Where does the sperrmuell Matcher-upgrade item live?** ~~It's their repo,
their priority. Write it into sperrmuell-direct's TODO_LIST now, or leave it
parked in this report until you next work there?~~ **SUPERSEDED — the T24
sperrmuell migration onto `httputil.Language()` (10-21 f2) carries the Matcher
swap; gated on the v1.6.0 tag.**

**Q3 — Which canonical shape does the module cookbook optimize for?** ~~German-canonical-unprefixed + `/en` (sperrmuell's SEO shape) as the worked
example, or also an EN-canonical global-SaaS shape (all locales prefixed,
`x-default` → `/`)? Affects defaults, example, and docs depth.~~ **MOOT —
module parked; the sperrmuell-shape worked example lives in httputil's
cookbook (10-21 f13).**

---

_Point-in-time snapshot — goes stale. Section f is docs-health HARVEST input.
Written as `.md` per explicit user instruction (skill default is HTML;
one-off override, not propagated into the skill)._
