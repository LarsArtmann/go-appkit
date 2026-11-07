# Auth-Bridge Decision Memo — usermgmt Recipe vs Security Session Battery

**Date:** 2026-11-07 · **Status:** DECISION REQUESTED (owner-gated ○)
**Feeds:** TODO_LIST "Auth-bridge DECISION memo" row (routed 2026-11-06) ·
SUPERB v12 T5.4 · the preset's future auth leg (ships WITHOUT auth either
way — this memo decides whether/when that changes)
**Demand source:** `docs/review/apps/2026-11-06_integration-burden-audit.md`
(10 of 39 audited apps hand-roll JWT/session auth; population corrected to
41 — the audit's per-app table stands)

## The decision in one line

Ten audited apps hand-roll auth despite the family owning
`cqrs-htmx/usermgmt` (event-sourced WebAuthn/OAuth2/TOTP). Decide: route
those apps through **a recipe that mounts usermgmt on appkit hosts**, or
build **a lighter session battery inside appkit/security** — or both, in
staged order.

## Options

| # | Option | New code | Fit | Risk |
| --- | ------ | -------- | --- | ---- |
| a | **usermgmt-on-appkit recipe** (document mounting cqrs-htmx's usermgmt handlers on `svc.Mux`; app stays the appkit host, cqrs-htmx is the auth block) | zero (docs only) | heavyweight auth needs (WebAuthn/OAuth2/TOTP, admin UI); pulls the full cqrs-htmx/system stack into plain non-CQRS apps | dependency weight for simple apps; direction is safe (cqrs-htmx already depends on appkit) |
| b | **Security session battery** (session store + password hash + login/logout handlers inside `appkit/security`, which already ships API-key auth, CSRF, rate limits, origin check) | new, security-critical | the ChastityAPI archetype (its 1,068-LOC `internal/auth` is exactly cookie/csrf/device/password/ratelimit/token session plumbing) | appkit starts OWNING token/session correctness — the highest-consequence surface a framework can own; audit's hand-rolls are heterogeneous (some want WebAuthn, some plain sessions) |
| c | **Staged: (a) now, demand-gate (b)** — recipe first; build the session battery only when ≥3 concrete consumers commit to it | docs now, code later | covers both archetypes without speculative security code | the battery arrives later than a from-day-one design would |

**Recommendation: (c).** Reasoning: the 10-app demand is real but
unconcentrated — no two hand-rolls look alike, which is the classic signal
that a battery would guess wrong about the 80%. usermgmt EXISTS, is
tested, and mounts as HTTP handlers; a recipe unlocks the heavyweight
archetype today with zero new attack surface. The session battery is the
right home for the ChastityAPI archetype but should be shaped by ≥3
committed consumers (the same demand-gating that kept WebSocket, DI, and
cron out of core). CSRF + rate limits + origin check are ALREADY in
security v0.2.0 — the battery, when demand lands, is session-store +
password-hash + handlers on top of existing pieces, not a from-scratch
framework.

## Verified facts the decision rests on

- appkit/security v0.2.0 ships: API-key auth, CSRF, keyed rate limits,
  origin check, body limit, sanitization, CSP nonce/policy — ALL opt-in
  (`security/THREAT_MODEL.md`).
- cqrs-htmx usermgmt: event-sourced user management (WebAuthn/OAuth2/TOTP)
  exposed as HTTP handlers + admin UI; cqrs-htmx depends ON appkit
  (setup → core direction), so both blocks compose in one process.
- Audit exemplars: ChastityAPI `internal/auth` 1,068 LOC
  (cookie 53, csrf 92, device_auth 53, middleware 365, password 100,
  ratelimit 201, request_id 52, token 150) — near-1:1 overlap with
  security + a missing session store.
- The preset's spine deliberately ships without an auth leg; this memo
  gates whether one is added and in which shape.

## If (c) is chosen, the immediate next step

Write the recipe: `docs/recipes/auth-bridge-usermgmt-recipe.md` (copy the
battery-swap template shape: inventory → mount → verify), exercise it
against one real consumer (ChastityAPI recipe A4 is the natural pilot),
and record the session-battery demand gate (3-consumer trigger) in
TODO_LIST P3. First implementation step regardless of choice: verify
usermgmt's handler bundle mounts on `svc.Mux` with zero appkit changes
(one scratch service, boot + register + probe).

**Decision requested:** pick (a)/(b)/(c) on the TODO_LIST P2 auth-bridge
row or reply in-session.
