# endpoints — Utilization Deep Dive

> **Point-in-time review · 2026-11-06 · go-appkit @ master (mid-SUPERB v12)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`) — but written
> LATER than its siblings: `endpoints/v0.1.0` was tagged 2026-11-06 05:27, three days
> AFTER the 2026-11-03 audit, which therefore could not cover it. This file closes that
> gap (the `httputil.md` report carries a matching errata line).

> **Note on naming:** there is no standalone `/home/lars/projects/endpoints` checkout —
> the library lives as a subdirectory module of the httputil repo
> (`~/projects/httputil/endpoints/`, module path
> `github.com/larsartmann/httputil/endpoints`, tagged `endpoints/v0.1.0`).

## Verdict

**NOT USED — and not declined.** Zero imports across all 14 modules; stronger than
server_timing's go.sum-only presence, it has NO edge in the module graph at all (no
require, no test-chain transit). Neither known consumer (cqrs-htmx setup, PapDashboard)
uses it either. The package is hours old at review time. It is, however, the
**strongest pending-adoption candidate in the family**: it IS the route-registry seam
that owner decision H8 (`svc.Routes()`, SUPERB v12 TL W8) has been waiting on —
shipped stdlib-only, type-compatible with `svc.Mux` as-is.

| Fact              | Value                                                          | Evidence                                                       |
| ----------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| Direct imports    | 0 of 14 modules (consumers too: 0)                             | `rg 'httputil/endpoints' --glob '*.go'` → empty                |
| In module graph?  | No — no require edge, not even go.sum                          | `rg 'httputil/endpoints' -g 'go*.sum'` → empty                 |
| Latest tag        | endpoints/v0.1.0 (2026-11-06 05:27 +0100) = first release      | `git -C ~/projects/httputil log -1 endpoints/v0.1.0`           |
| Dependencies      | stdlib-only (own go.mod, `go 1.27`)                            | `endpoints/go.mod`                                             |
| Adoption path     | H8 `svc.Routes()` seam → B6 golden test → B8 docs feed         | TODO_LIST W3; batteries spec B6/B8                             |

## 1. What the package offers (capability inventory)

A self-describing route listing for `http.ServeMux` — the registry IS the source of
truth, because `*http.ServeMux` cannot enumerate its own patterns
(`~/projects/httputil/endpoints/endpoints.go`):

- `Registry.Handle/HandleFunc(mux, pattern, handler, summary)` — mounts the route on
  the mux AND records it, in one declaration. Never-panic API: nil mux/handler, empty
  pattern, invalid-UTF-8 pattern/summary, and mux-rejected patterns (malformed,
  duplicate — translated from ServeMux's panic) all come back as errors.
- `Registry.Routes() []Route` — sorted (pattern, then method) defensive copy; the
  golden-test input B6 wants.
- `Route{Method, Pattern, Summary}` — Summary is exactly the metadata carrier B8's
  docs feed needs.
- `Register(mux, registry)` — mounts `GET /endpoints` (DefaultPattern); the listing
  includes itself.
- RFC 9110 proactive content negotiation (`negotiate.go`/`qvalue.go`): real q-value
  parsing, `text/html`/`text/*` → HTML table, `application/json`/`application/*`/`*/*`/
  absent → JSON, ties break to JSON, never 406 (lenient posture, mirrors the root
  module's `Language` middleware). `Vary: Accept` + `Cache-Control: no-store`.
- HTML rendering with no inline styles/scripts (strict-CSP safe), all fields
  HTML-escaped, fuzz-tested (`FuzzRenderHTML`); UTF-8 validated at registration so the
  listing is renderable by construction.
- Concurrency-safe: reads off immutable atomic snapshots; routes can be added while a
  listing serves.
- Quality bar: 742 LOC non-test + unit/negotiation/fuzz/bench/sentinel/example tests
  (~1,300 test LOC).

## 2. Why this matters to appkit specifically (the H8 fit)

1. **`svc.Mux` is `*http.ServeMux`** (`service.go:25`) — `Registry.Handle(svc.Mux, …)`
   is type-compatible TODAY. A consumer can adopt endpoints with zero appkit changes;
   the appkit-level question is only whether core/preset EXPOSES the registry.
2. **H8 is literally this seam.** The SUPERB v12 unified execution plan rows
   "`svc.Routes()` seam decision (owner) + B6 golden + B8 docs feed" as TL W8
   [OWNER], and the 13-decision questionnaire owes an httpx-vs-httputil explanation.
   The batteries spec (2026-09-04, pre-dating this package) scoped B6's seam as NEW
   httpx work: "introspectable if registration flows through one seam (which B8 needs
   anyway)". That seam now EXISTS in-family — building a second one in httpx would
   recreate the duplicate-wheel pattern this fleet audits against.
3. **B8's anti-drift promise is the package's core claim**: docs generated from the
   LIVE route table cannot drift from what is served (`docs/` module consumes
   `Routes()`, catalog/v4 renders).
4. **Security posture matches the house charter**: the listing documents the attack
   surface, so the README says mount it deliberately/behind auth — an opt-in preset or
   security-gated leg, exactly how every appkit battery ships (nothing joins the
   default stack unasked).

## 3. Why it is not wired today, and what flips it

- Released hours ago; no appkit or consumer demand signal has reached it yet.
- The flip is a DECISION, not a build: H8's owner answer. Options the decision
  weighs: (a) core adopts `Registry` as `svc.Routes` (routes registered through the
  service, listing opt-in via config), (b) a preset/httpx leg wires it, (c) decline
  and keep direct `svc.Mux` registration (consumers can still adopt it themselves —
  note the summary metadata then lives outside appkit's contract).
- Minor observation for B5, not a defect: errorpages' `wantsJSON`
  (`errorpages/errorpages.go:44`) sniffs Accept with `strings.Contains` where
  endpoints ships a real q-value negotiator — different scope (error pages vs
  listing), but B5's negotiation battery has TWO in-family precedents to reconcile
  (this and httputil `Language`).

## 4. Recommendations

1. Feed this report into the H8 owner decision — the httpx-vs-httputil explanation
   owed there now has a concrete answer: "the seam shipped, in httputil".
2. If H8 adopts: B6's golden test reads `Registry.Routes()` in testkit; B8's docs feed
   consumes the same table; no httpx-owned registry should be written.
3. Do NOT port or re-implement the listing in appkit (same rule as the CV
   X-Response-Time port refusal for server_timing).
4. Re-check version currency at the next httputil release train.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'httputil/endpoints' --glob '*.go' --glob 'go.mod' --glob 'go.sum'   # empty → zero usage, not in graph
rg -n 'httputil/endpoints' /home/lars/projects/cqrs-htmx /home/lars/projects/papdashboard 2>/dev/null | head -1  # empty → consumers neither
git -C /home/lars/projects/httputil log -1 --format='%ci %h %s' endpoints/v0.1.0  # 2026-11-06 05:27 release
ls /home/lars/projects/httputil/endpoints/       # the library's real home
sed -n '25p' service.go                          # Mux *http.ServeMux — drop-in compatible
```
