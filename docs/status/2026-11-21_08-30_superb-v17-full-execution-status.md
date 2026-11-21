# SUPERB v17 full execution — the default-onramp wave

**As of:** 2026-11-21 08:2x · **Plan:** `docs/planning/2026-11-21_01_56-SUPERB-v17-the-default-onramp-plan.html` (24 tasks) · **Corpus:** `docs/review/apps/2026-11-17_go-appkit-adoption-review.html` · **Verdict:** FULL EXECUTION MODE ran end-to-end; the release wave SHIPPED (core v0.13.0 + security v0.3.0, gate green, proxy-verified); every task has a verdict below, three are honest partials with recorded reasons.

## 1% tier (T1–T4)

- **T1 the-default ruling** — ENCODED. The ruling is the owner's 2026-11-21 directive ("make go-appkit THE default every other project builds on"). Landed: `stack-selection` SKILL.md Phase 4 FLEET DEFAULT bullet + Phase 6 wiring row (`~/projects/SKILLS/stack-selection`), the STACK-template emitter's Service-base row, the global crush-config AGENTS.md policy line (fleet default under Project Discovery), and the appkit README claim ("The fleet default", with the decision date).
- **T2 go-service-template** — SHIPPED. `github.com/LarsArtmann/go-service-template` CREATED (public) + filled per the scaffold verdict: cmdguard v5.1.0 `serve` CLI (+`version`/`doctor` batteries), preset v0.1.1 spine, httpx v0.2.0 route, templ page (committed codegen; fleet-migrated pairing: templ v0.3.1100 + templ-components v1.27.2 + go-health-dashboard v0.13.3 — cmdguard v5.1.0's templ ≥1100 floor forced the aligned stack), build-tag-gated cqrs stub, flake via go-nix-helpers go-standard with the go 1.27.4 tarball pin (nixpkgs go_1_27 = 1.27.3 < floor), BuildFlow/CI/systemd/distroless-Docker, AGENTS stamp block + refresh ritual, README, metadata.yaml. Suite green; lint 0 (fleet config copied); nix build produces the real binary.
- **T3 dogfood probe** — PASSED same day: build, boot (`/healthz` JSON, `/api/status`, templ `/`, pretty 404), SIGTERM → the full five-phase drain lines + `graceful shutdown complete result=ok`. Friction recorded in the template CHANGELOG (templ pairing; `APP_ADDR` earned its keep — :8080 was occupied by an unrelated dev server).
- **T4 policy wiring** — VERIFIED: the skill edit is live through the fan-out (symlink → SKILLS repo), `check-skill-fanout.sh` green (54 symlinks clean, no store pins).

## 4% tier (T5–T12)

- **T5 listener-seam design note** — `docs/planning/2026-11-21_listener-injection-seam-design-note.md` (8 contract pins, test matrix, rejected alternatives).
- **T6 listener seam** — SHIPPED as core **v0.13.0** `ServiceConfig.Listener`: ownership transfers at `Start` (every failure path + graceful shutdown close the supplied listener), single-serve carried over, classifications + six phase lines unchanged; `listener_test.go` (5 pins incl. unix socket + both failure-path ownership pins) + the integration unix-socket E2E **through the published tag**. Additions-only API.
- **T7 logging seam** — SHIPPED as core v0.13.0 `ServiceConfig.LogHandler slog.Handler` (pinned: lifecycle lines flow through any handler) + the owner memo `docs/planning/2026-11-21_fleet-log-lane-owner-memo.md` (option 3 flex recommended; the default-lane flip is the census trigger). **Immediate payoff: dnsblockd's rejection blocker fired its own refresh rule the same day** (see T13).
- **T8 leaf-decoupling spike** — `docs/planning/2026-11-21_leaf-decoupling-spike.md`: option A verdict (local `Hook` alias in systemd; aliases unify so it is ZERO API break; go-appkit require → test-only), routed to TODO_LIST P2 to ride the next systemd train.
- **T9–T11 adoption surfaces** — README Start-here (preset-first + three recipes + template link), One-line-wins table, the runnable verify-recipes block (live-probed: Server-Timing header + drain lines captured), `docs/CAPABILITIES.md` (per-capability map with since-versions), `docs/MODULE_INVENTORY.md` + `scripts/module-inventory.sh` (origin-truth design so it cannot drift at tag time) wired into pre-tag-checks as `check-module-inventory.sh` (born-proven: stale detection + the dprint-alignment/timestamp normalization + the no-op-regen guard).
- **T12 divergence stances** — ROADMAP section with a stance per nsfw-classifier blocker (① SHIPPED today, ② declined-with-reason, ③ accepted-as-divergent, ④ seam shipped + census trigger); security/README links to it.

## 20% tier (T13–T24)

- **T13 dnsblockd** — FIRST LEG EXECUTED: CSP nonce → `security.GenerateNonce` (v0.3.0; zero-appkit-dep module; byte-identical 16-byte std-base64 contract); their suite green. Their review RE-FRESHED per its own protocol: **the slog blocker FIRED today** (v0.13.0 LogHandler is the pluggable backend their rejection demanded) — recorded demand signal: `BuildCSP` lacks a ReportURI field.
- **T14 ChastityAPI** — RE-JUDGED + REFRESHED: their ADR-0004 selective-adoption doctrine re-affirmed with the seam-wave facts; the three swap legs re-judged honestly (app-shaped halves dominate: JWT dev-fallback gate, secret-never-logged semantics, the httputil-based CSP variant with style-src-attr that `BuildCSP` cannot express — recorded demand signal). Realtime v0.2.0 folded contract now PUBLISHED for their trigger.
- **T15 PapDashboard** — EXECUTED SUBSET: security v0.3.0 PINNED + constant-time compare → `security.APIKeyMatches` (their middleware + cmd/server suites green; pre-existing failures in their api/auditrun/federation packages are NOT this change and stay theirs). The honest split recorded in their `docs/planning/2026-11-21_security-swap-verdict.md`: RateLimit/BodyLimit are ONE DEPLOYMENT RULING away (XFF-trust posture + the 413 eager-vs-typed cutover); publicPaths/OriginGuard stay app-shaped forever.
- **T16 pbx-lens** — DECLINED BY THEIR FRESHER AUTHORITY (the honest no-op): pbx-lens's own cmdguard review (2026-11-11/12) REJECTS with a no-cobra stack policy and three unfired flip triggers, re-verified unfired 2026-11-21; their review carries the reconciliation annotation and the leg routes to trigger-watch. The go-appkit plan row is corrected.
- **T17 cmdguard agent page** — the walked loop EXISTED (landed by the 2026-11-20 session) but was not reproducible: the flagship taskctl did not wire `WithOutputFormat`, so the doc's `--output json` bytes were unreachable. FIXED AT SOURCE: taskctl now registers the `--output` flag (Table stays the human default); the loop walked end-to-end against the real binary and the doc now carries the REAL envelope bytes (`required_flag` / `title`) + the walked-date note. taskctl tests green.
- **T18–T21** — done (see 4% tier / adoption surfaces; T19: counter re-baselined — **42 Go apps, go-appkit any 9/42, direct 7/42** (`docs/review/apps/adoption-baseline.txt`, 2026-11-21 03:04; population grew 41→42 with no adopter change); T20: Release Ritual step 7 (lockstep-tag notice, `[Unreleased]` loudness, post-push inventory regen); T21: CAPABILITIES page; T23: the stamp-bump bullet lives in AGENTS Adoption & Drift Rituals.
- **T22 security train** — SHIPPED as security **v0.3.0** (the samber/lo runtime adoption + example dep consolidated at v0.12.0; zero package API delta) — the fresh pin T15 needed; it existed the SAME DAY T15 ran.
- **T24 reciprocity** — appkit README Related-projects (cmdguard named for the ops-CLI leg) + the fleet-default paragraph added to THREE sibling skills (cqrs-htmx Step-1 path tree, go-cqrs-lite Works-with, stack-selection).

## Release wave facts

- Tags: `v0.13.0` + `security/v0.3.0` (annotated; CHANGELOG-dated FIRST, tagged LAST — the incident rule held). Integration re-pinned (core v0.13.0, security v0.3.0) + `documentedPins` fixture + the new unix-socket E2E, tidied via the insteadOf trick, suite `-race` green.
- `pre-tag-checks.sh v0.13.0 security/v0.3.0`: **GREEN (155 OK)** — including the new inventory guard. The proxy's stuck-404 class delayed gosum-canonical ~30 min (`.info` 200 / `.zip` 404 flapping across frontends); the deferred leg re-ran green: "integration go.sum agrees with proxy + sumdb".
- Fresh-consumer checks: core v0.13.0 (Listener + LogHandler compile-level usage, fresh GOMODCACHE) GREEN; security v0.3.0 (BuildCSP/GenerateNonce usage) GREEN.
- The daemon pushed the wave mid-prep (its normal behavior); master pushed clean at `ae96480`; CI on the tip verified at close-out.

## Honest partials (nothing silent)

1. **T13/T14/T15 deeper legs** are one deployment/ADR ruling away each — recorded in the consumer repos' own docs, not silently dropped.
2. **T16** did not execute by design: the fresher local verdict wins (challenge-instructions discipline); corrected on both sides.
3. **`go.mod` requires note:** security's example dep was already v0.12.0 (daemon sweeps landed it ahead of the CHANGELOG text) — the dated entry records the shipped require.

## Session totals

26/26 todo items; 2 tags shipped; 1 new repo (public, CI green after the action-version fix); 6 repos touched (go-appkit, go-service-template, SKILLS/stack-selection, crush-config, cmdguard, PapDashboard, dnsblockd, ChastityAPI, pbx-lens); every fleet guard green at close-out.
