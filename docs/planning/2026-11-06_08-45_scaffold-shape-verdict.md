# Scaffold Shape Verdict — template repo, not a generator (SUPERB v12 S1 / G3)

**Decided:** 2026-11-06 · **Owner gate:** G3 (this doc is the recommendation; execution
starts on the owner's nod, or immediately under the v13 default sequencing if no answer) ·
**Input evidence:** `docs/review/apps/2026-11-06_integration-burden-audit.md` §2 (37/39
flake.nix · 20 buildflow · 16 Dockerfile · 6 systemd units — every new app copy-pastes a
cousin), `docs/planning/2026-11-06_07-10_SUPERB-v13-ship-the-spine-refresh-plan.md` Tier 2.

## Options considered

| Option                                          | Verdict | One-line reason                                                                                                                     |
| ----------------------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **Template repo (go-website-template pattern)** | ✅ PICK | Proven fleet pattern (the website template already killed its copy-paste class); zero generated-code maintenance; dogfoodable today |
| `cmd/appkit-new` generator now                  | ❌      | go-wizard-sdk's template engine is UNBUILT — the stated blocker; a hand-rolled generator is code-to-maintain for the same output    |
| Generator later, on go-wizard-sdk               | ⏭ watch | Revisit when go-wizard-sdk ships its engine; a template repo ports INTO a generator almost verbatim (templates are the asset)       |

## The pick

A new repo, `go-appkit-template` (sibling to go-website-template), carrying exactly the
v13 plan's S2–S7 slices:

- `cmd/` main via **cmdguard v4** `serve` verb → `config.Load` → `preset.New` (typed
  flags, exit-code contract) — the preset is the point of the scaffold.
- **templ + templ-components** skeleton page + static assets route (CSP-nonce posture).
- **flake.nix** (ldflags version-stamp pattern) + **.buildflow.yml** (GOWORK off + the
  skip set mirroring this repo) + CI workflow (matrix + the offline guards minimal set) +
  **systemd unit Type=notify** + Dockerfile.
- README/CHANGELOG/AGENTS/FEATURES skeletons + a boot-probe script
  (build→boot→/readyz→graceful-stop, the example-smoke pattern).
- Dogfood gate (S7): scaffold ONE real toy app end-to-end; record time-to-first-route +
  the pain list in the template README before calling it done.

**Maintenance owner:** this repo's TODO_LIST until the template repo exists, then the
template's own repo (go-website-template lifecycle: it maintains itself via dogfooding).

## Why not both, now

The generator's only edge over a template is parameterization (module name, service
name). `git clone + find-replace` covers that at N=1..50 apps. The fleet's actual failure
mode is not "renaming is hard" — it is "the pieces were never assembled once, correctly".
Assembling them ONCE in a template fixes that; the generator is an optimization on top of
a solved problem, gated on an unbuilt dependency.

## Revisit trigger

go-wizard-sdk ships its template engine → port the template's assets into generator
scaffolds; keep the template as the generator's own source of truth (no fork).
