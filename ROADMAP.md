# Roadmap — go-appkit

> Long-term direction and raw ideas not yet refined into actionable tasks.
> Bounded, short-to-mid-term work lives in [TODO_LIST.md](TODO_LIST.md).
> Point-in-time plans and research live in `docs/planning/`.

**Updated:** 2026-11-13 (docs-health pass — module-bay list current at 14 released modules; httpx shipped B1-B9 minus the H8-gated pair; W3 complete, W4-W5 routed) — prior: 2026-11-02 (httputil#41 `NewServerListener` SHIPPED upstream — composition + Core TLS unblocked; see the v1.0.0 section)

## North star

A batteries-included SDK for the larsartmann Go ecosystem: when a family
service needs a security posture, an SSE contract, a worker loop, or a SQLite
lease, the answer is `go get github.com/larsartmann/go-appkit/<module>` — not a
fresh 400-LOC hand-roll. "Batteries included" never means a fat core: core
stays a thin, stable HTTP service lifecycle; capability arrives as opt-in,
independently versioned satellite modules.

The module bay is built (cqrs, realtime, otel, health, flightrecorder,
flightrecorderhealth, errorpages, docs, security, httpx, config, systemd,
preset — 14 released modules plus the unreleased integration suite).
Filling it is demand-driven — the canonical demand analysis and per-battery
specs live in `docs/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md`
(waves W1–W3 shipped — B6/B8 remain H8-owner-gated; W4–W5 are routed into
TODO_LIST; this file holds the direction, not the tasks).

**Composition-mission program (OWNER RULING 2026-11-03):** the mission is
"make it easy for my apps to smartly use all my libs together" — family-lib
batteries are IN-SCOPE, demand-gating them contradicts the repo's purpose.
All W3-W5 batteries are blanket-promoted to planned work (sequencing lives
in TODO_LIST). The `httpx` handler-DX module SHIPPED (B1-B9 except the
H8-gated B6/B8; B5 content negotiation built on go-codec) — the remaining
structural ideas are the W4/W5 batteries and the auth-bridge decision (both
owner-gated, TODO_LIST). Server-Timing
now ships in core's default stack with a `ServerTiming *bool` knob —
framework-measured app phases are unservable (stamp-on-commit contract);
sub-phases are handler-side via the request collector.

## v1.0.0 (core)

Core's stability contract. The working exit-criteria draft lives at
`docs/planning/core-v1-exit-criteria.md` (hard criteria: mechanical API-break
check wired into the release ritual, consumer count, docs and telemetry
posture). v1-shaped additions already shipped: `OuterMiddlewares`,
`ShutdownHooks`, `DrainHooks`, the sentinel registry (`NoTimeout`, `NoDrainDelay`),
the metrics/version/testkit surface (v0.5.0).

Open questions on the way there (deliberate USER GATES, not tasks — full
context: TODO_LIST footer and `docs/status/archived/2026-09-17_14-22_setup-usage-verification-and-agentsmd-drift-fix.md` §g):

- ~~The httputil listener-injection API (`NewServerListener`) go/no-go~~
  RESOLVED — SHIPPED upstream 2026-10-25: httputil v1.7.0 carries
  `NewServerListener(ln, cfg, handler)` (closes
  [httputil#41](https://github.com/LarsArtmann/httputil/issues/41)). Both the
  composition refactor and Core TLS are UNBLOCKED; the spike verdict doc
  carries the 2026-10-25 update banner and the work is the TODO_LIST P2
  composition row.
- ~~Whether to file the two ready upstream asks (go-sse `ReplayFiltered`,
  httputil Logging request-context).~~ RESOLVED 2026-10-13: `ReplayFiltered`
  already exists upstream (no ask); `LoggingFromContext` covers the
  request-context completion line (wired appkit-side in otel v0.3.0).
- ~~Pin philosophy for `integration/` (LATEST-only vs mirroring setup),
  cross-repo tracking posture, and where release-state pins live
  long-term.~~ RESOLVED: LATEST-only, enforced by the `documentedPins`
  fixture + `check-pin-drift.sh` (cross-repo legs since 2026-09-28);
  consumer claims guarded by `check-consumer-claims.sh` (2026-10-12).

## Ecosystem adoption (reverse direction)

The growth pattern that works: OTHER repos adopt appkit as their host layer —
appkit never grows application-shaped code.

- **cqrs-htmx `setup`** — reference consumer; ADR-001 decided, `RunWithAppkit`
  SHIPPED (drives an appkit Service; the Metrics/Version threading (M4)
  shipped in setup/v4.14.2 and rides every newer tag — AGENTS Release State
  owns the version truth); the `Bundle.Run`/`RunHandler` internals flip and
  the v5-window revisit remain on their side.
- **PapDashboard** — recommended reverse adoption (their repo); researched
  2026-09-04, door held open in TODO_LIST P3. First concrete core-TLS feature
  request if they keep app-level TLS.
- **go-plugin-mvp (Kernovia)** — same pattern recommended; pre-1.0, gated on
  their license/rename decisions (`docs/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md`).
- **cordis bridge module** — trigger-gated NOT NOW (`go/v0.1.0` tagged
  2026-09-16 = trigger 1 of 3 met; zero consumer demand, core v1 criteria
  pending).

## Raw ideas (unrefined, no commitment)

- Per-route sampling/verbosity override for logs+traces (Stalwart
  `EventTracingLevel` analog). (The observability umbrella itself SHIPPED
  2026-09-16 as `docs/TELEMETRY.md`.)
- Route-cardinality fuzz guard: 10k distinct request paths must produce a
  bounded metric series (double-relevant after the pattern-propagation
  regression).
- `httpx` handler-DX module (ResultHandler family, bind+validate, no-leak
  errors) — W3 of the battery program; requires a core route-registration seam
  (`svc.Routes()`) for introspection that core has not agreed to yet.
- Route metadata → docs live feed: generated docs that cannot drift from
  routes (B8; same core seam prerequisite).
- Config modules (koanf wrapper, hot-reload via fsnotify with
  validate-gate → atomic swap) — F1/F4; demand-gated on a consumer with
  runtime config needs.
- `worker` supervisor + typed pool, `sqlite` ops kit, `polite` outbound
  client, `do` bridge — W4; each is a standalone-module candidate with a
  port-from source in the battery spec.
- errorpages rendering `trace_id` when a span is active; flightrecorder
  linking the snapshot file to the active span (support-handoff UX). The
  baggage correlation-ID helpers and the outbound-client span transport
  (`appkitotel.Transport()`) SHIPPED in otel v0.3.0 (2026-10-18) — rows
  retired from this roadmap.
- Multi-instance realtime: external-bus driver story (NATS/Redis) behind the
  Hub, if a consumer actually scales out.
- Multi-recorder coordination ADR: one `fr.Recorder` serving the HTTP
  middleware + projection host + health triggers in one process (the READMEs
  claim it works; an ADR should own the contract; pairs with the
  go-health aggregate trigger below — both gate on the same multi-surface
  demand signal).
- cqrs ops/recipes backlog ported from the 2026-09-04 deep-dive brainstorm:
  CBOR→JSON transcode helper decision for SSE raw-payload consumers; one
  shared `fr.Recorder` demo across HTTP middleware + projections; DLQ admin
  and dead-letter-age alerting recipes; SnapshotStore/ReadModels accessor
  examples (Bundle-reachable, undocumented).
- Example-app and ops gaps from the 2026-09-04 SUPERB plan:
  errorpages+cqrs composition example (F44 — one cqrs/ES app serving
  classified error pages) and a per-projection readiness helper (F27,
  demand-gated — surface cqrs projection staleness as a go-health probe).
- go-health aggregate/federation adoption (evaluated 2026-09-20, plan T25;
  re-checked 2026-10-12 and re-baselined 2026-11-02: the health module now
  pins go-health v0.8.1 (health v0.3.0), so the old "v0.3.0 floor gate" is
  long resolved — the idea remains demand-gated, unchanged): `aggregate` merges N in-process probes into one
  go-health-compatible surface (passive, zero goroutines, "name/check"
  namespacing) — natural appkit composition is a `MountAggregate`-style
  handle next to `appkithealth.New` so a service running several surfaces
  (module probe + cqrs projection readiness + subsystem probes) can serve
  ONE /readyz; `federation` rolls up REMOTE go-health HTTP documents (5s
  fetch timeout, 1MiB cap) — that is fleet/ops-plane material, not a
  framework concern. Neither ships now: no consumer demand signal yet.
  Trigger: a real consumer needs single-endpoint multi-surface readiness →
  wrap aggregate in a Mounted-compatible handle (health module stays
  injector-free).

- Fleet tooling patterns (harvested 2026-10-13 from the 15-53 §f45 extras):
  workspace-exclusion convention (go.work member exclusions declared via
  tool config, not folklore — the BuildFlow#46 class, fleet-wide);
  per-tool config schema docs (`.buildflow.yml`, `.jscpd.json` — the
  dead-exhaustruct-block class argues for documented schemas); charter
  guard as a library (`check-workspace-charter.sh` is a per-repo copy —
  extract once it drifts between repos).
- Cross-repo follow-through (external, watch-only): quick fleet sweep for
  go-work-sync re-adds in sibling repos (leak #8's mutator was fleet-wide
  tooling, pre-#46); buildflow-skill PR documenting the `.jscpd.json`
  convention + the go-work-sync/charter caveat + the 2026-10-12
  verification-status rows.
- BuildFlow-repo link-check config: lychee excludes for archived
  dependency-diagram trees (`docs/library-analysis/archive`) so the link
  checker stops re-walking settled diagrams.
