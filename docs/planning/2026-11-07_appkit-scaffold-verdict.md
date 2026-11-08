# Verdict: appkit scaffold — template repo, not a generator

**Date:** 2026-11-07
**Trigger:** SUPERB v12 tier 2 (TODO_LIST P2 scaffold row) — the cross-app
integration audit found the new-app copy-paste class (37/41 flake.nix,
20 buildflow, 16 Dockerfile, 6 systemd units hand-assembled per app) and
asked for a decision: template repo (go-website-template pattern) vs a
`cmd/appkit-new` generator. go-wizard-sdk's template engine is UNBUILT and
must not block this decision.

## Decision

**Build a template repo** (sibling, go-website-template pattern, proposed
name `go-service-template`). The generator is DECLINED until two triggers
fire (below).

## Why the template wins

1. **The pattern is proven in-fleet.** go-website-template already runs the
   exact shape this row asks for: `cmd/<name>` main, `internal/` layers,
   templ + HTMX + Tailwind views, flake.nix + distroless Dockerfile
   coexistence (with a written why-table), AGENTS.md/CHANGELOG skeleton,
   verified build paths. Cloning a sibling repo is how the fleet actually
   starts apps today — the tool should match the habit, not fight it.
2. **Zero new runtime tool.** A generator is a new cmd + its own tests +
   its own release train + a version-skew surface against appkit itself.
   The template carries NONE of that: its drift is bounded by a stamp
   refresh (below), not by software that can be wrong.
3. **go-wizard-sdk is unbuilt.** Writing `cmd/appkit-new` today means
   hand-rolling string templating — exactly the class go-wizard-sdk exists
   to own. Building a worse version first is the premature-generalization
   anti-pattern.
4. **The appkit half of a new app is already one call.** preset v0.1.0
   ships the spine+legs composition (example main.go = 35 LOC, acceptance
   met). What the template adds is everything AROUND the service: cmdguard
   serve verb, views skeleton, flake/buildflow/CI/systemd/Dockerfile, docs
   skeletons. That content is static per archetype — a template's shape.

## Template spec (what `go-service-template` carries)

- `cmd/<name>/main.go`: cmdguard v4 serve verb → `preset.Run(...)` as the
  composition root (explicit code wins over env, per the preset grammar).
- Views skeleton: one templ page + errorpages mounting (pretty 404/405).
- `flake.nix`: build + devShell (GOTOOLCHAIN-agnostic) + `.#test` app.
- `.buildflow.yml` (env.GOWORK off; skip_steps per fleet convention).
- GitHub Actions CI: per-module matrix + the fast offline guards.
- `systemd/` unit: `Type=notify` + `WatchdogSec` (pairs with
  appkit/systemd's READY/STOPPING/watchdog contract).
- `Dockerfile`: distroless, non-root (go-website-template pattern).
- `README.md` / `CHANGELOG.md` / `AGENTS.md` skeletons with the go-appkit
  module table + Build & verify blocks pre-filled from the floor defaults.

## Drift control (the template's one weakness, made mechanical)

- The template carries a version stamp (appkit core/preset pins + go
  floor + generated date) in its AGENTS.md.
- Refresh ritual: any appkit train that moves preset or core bumps the
  stamp and re-runs the dogfood probe (below). Same cadence class as the
  integration re-pin ritual — one repo, one probe, no silent aging.

## Dogfood acceptance (copy from the TODO row)

Scaffold one real app end-to-end from the template: builds, boots, serves
a probe, graceful-stops clean (example-smoke pattern), main.go < 40 LOC.
The template repo is NOT done until one real app passed.

## Declined alternative + re-entry triggers

`cmd/appkit-new` generator — revisit when BOTH hold:

1. Template drift bites: two real scaffoldings between refreshes hit a
   stale pin/API (the refresh ritual failed to keep up).
2. go-wizard-sdk ships its template engine (then the generator is a thin
   wrapper over it, not a hand-rolled templating tool).

Until then: template repo, refresh ritual, dogfood probe. Owner-gated
(next repo to create; the appkit side keeps the trigger row).
