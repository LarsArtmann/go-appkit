# H8 Route-Seam Owner Ask — `svc.Routes()` vs `httputil/endpoints`

**Date:** 2026-11-07 · **Status:** DECISION REQUESTED (owner-gated ○)
**Feeds:** SUPERB v12 TL W8 (H8 owner row) · SUPERB v14 A8 · batteries B6/B8
**Source verdict:** `docs/review/libs/endpoints.md` (2026-11-06 deep dive)

## The decision in one line

The route-registry seam H8 was waiting for **already shipped in-family**
(`httputil/endpoints` v0.1.0, 2026-11-06, stdlib-only, type-compatible with
`svc.Mux *http.ServeMux` today). The owner question is no longer "build the
seam?" but "who exposes it?"

## Options

| # | Option | Work | Consequence |
| --- | ------ | ---- | ----------- |
| a | **Core exposes**: `Service` grows an optional `Registry` — routes registered through `svc.Routes.Handle(...)` land on `svc.Mux` AND the registry; listing opt-in via config | ~1 field + 1 config knob + docs; zero default-stack change | appkit owns the adoption contract; B6/B8 read `Registry.Routes()` |
| b | **Preset/httpx leg**: preset's spine (or the future httpx module) wires `endpoints.Register` + `Registry.Handle` for consumers who opt in | a new opt-in leg, no core surface | consumers using plain `svc.Mux` (the majority pattern) get no registry |
| c | **Decline, keep direct `svc.Mux`**: consumers adopt endpoints themselves | zero | summary metadata lives outside appkit's contract; B6/B8 lose their mechanical input |

**Recommendation: (a).** It is the smallest surface that makes B6 (route
introspection golden test) and B8 (docs feed — the anti-drift promise:
docs generated from the LIVE route table cannot drift) mechanical, and it
follows the fleet rule the audit keeps repeating: adopt the family wheel,
do not port it (`endpoints.md` §4: "no httpx-owned registry"). Option (b)
splits the seam across composition layers for no gain; option (c) is the
status quo the audit already flagged as the duplicate-wheel risk if httpx
grew its own registry.

## Facts the decision rests on (verified 2026-11-06)

- Zero current usage: no appkit module, cqrs-htmx setup, or PapDashboard
  imports `httputil/endpoints` (rg empty across all three repos).
- The package is release-quality: never-panic registration API, RFC 9110
  q-value negotiation (HTML/JSON, never 406), strict-CSP-safe rendering,
  fuzz-tested HTML escaping, concurrency-safe snapshots (742 LOC + ~1,300
  test LOC).
- Security posture matches the house charter: the listing documents the
  attack surface, so it mounts behind auth/deliberately — nothing joins
  the default stack unasked.
- B5 note (not a defect): errorpages' `wantsJSON` Accept-sniff vs endpoints'
  q-value negotiator are different scopes; if H8 adopts, B5's negotiation
  battery reconciles against BOTH in-family precedents.

## What this unblocks

- **B6** route introspection + golden test (testkit reads `Registry.Routes()`).
- **B8** route metadata → docs feed via the `docs/` module + catalog/v4.
- **C2** httpx skeleton can proceed WITHOUT owning a registry (B1/B2/B9 do
  not need the seam either way — C2 is not blocked on this answer).

**Decision requested:** pick (a)/(b)/(c) in TODO_LIST P2 (H8 row) or reply
in-session; everything here is pre-verified so the chosen option can ship
in the next train without further research.
