# Roadmap — go-appkit

> Long-term direction and raw ideas not yet refined into actionable tasks.
> Bounded, short-to-mid-term work lives in [TODO_LIST.md](TODO_LIST.md).
> Point-in-time plans and research live in `doc/planning/`.

**Updated:** 2026-09-17

## North star

A batteries-included SDK for the larsartmann Go ecosystem: when a family
service needs a security posture, an SSE contract, a worker loop, or a SQLite
lease, the answer is `go get github.com/larsartmann/go-appkit/<module>` — not a
fresh 400-LOC hand-roll. "Batteries included" never means a fat core: core
stays a thin, stable HTTP service lifecycle; capability arrives as opt-in,
independently versioned satellite modules.

The module bay is built (cqrs, realtime, otel, health, flightrecorder,
flightrecorderhealth, errorpages, docs, security). Filling it is
demand-driven — the canonical demand analysis and per-battery specs live in
`doc/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md`
(waves W1–W2 shipped; W3–W5 are routed into TODO_LIST; this file holds the
direction, not the tasks).

## v1.0.0 (core)

Core's stability contract. The working exit-criteria draft lives at
`doc/planning/core-v1-exit-criteria.md` (hard criteria: mechanical API-break
check wired into the release ritual, consumer count, docs and telemetry
posture). v1-shaped additions already shipped: `OuterMiddlewares`,
`ShutdownHooks`, `DrainHooks`, the sentinel registry (`NoTimeout`, `NoDrainDelay`),
the metrics/version/testkit surface (v0.5.0).

Open questions on the way there (deliberate USER GATES, not tasks — full
context: TODO_LIST footer and `doc/status/archived/2026-09-17_14-22_setup-usage-verification-and-agentsmd-drift-fix.md` §g):

- The httputil listener-injection API (`NewServerListener`) go/no-go — the
  single unlock for BOTH the composition refactor and the Core TLS option.
- Whether to file the two ready upstream asks (go-sse `ReplayFiltered`,
  httputil Logging request-context).
- Pin philosophy for `integration/` (LATEST-only vs mirroring setup),
  cross-repo tracking posture, and where release-state pins live long-term.

## Ecosystem adoption (reverse direction)

The growth pattern that works: OTHER repos adopt appkit as their host layer —
appkit never grows application-shaped code.

- **cqrs-htmx `setup`** — reference consumer; ADR-001 decided, spike-validated,
  `RunWithAppkit` fold-in pending on their side.
- **PapDashboard** — recommended reverse adoption (their repo); researched
  2026-09-04, door held open in TODO_LIST P3. First concrete core-TLS feature
  request if they keep app-level TLS.
- **go-plugin-mvp (Kernovia)** — same pattern recommended; pre-1.0, gated on
  their license/rename decisions (`doc/planning/2026-09-04_cordis-and-go-plugin-mvp-integration.md`).
- **cordis bridge module** — trigger-gated NOT NOW (`go/v0.1.0` tagged
  2026-09-16 = trigger 1 of 3 met; zero consumer demand, core v1 criteria
  pending).

## Raw ideas (unrefined, no commitment)

- Per-route sampling/verbosity override for logs+traces (Stalwart
  `EventTracingLevel` analog). (The observability umbrella itself SHIPPED
  2026-09-16 as `doc/TELEMETRY.md`.)
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
- Baggage correlation-ID helpers and an outbound-client span transport in the
  otel module (`appkitotel.Transport()`).
- errorpages rendering `trace_id` when a span is active; flightrecorder
  linking the snapshot file to the active span (support-handoff UX).
- Multi-instance realtime: external-bus driver story (NATS/Redis) behind the
  Hub, if a consumer actually scales out.
- Multi-recorder coordination ADR: one `fr.Recorder` serving the HTTP
  middleware + projection host + health triggers in one process (the READMEs
  claim it works; an ADR should own the contract).
- cqrs ops/recipes backlog ported from the 2026-09-04 deep-dive brainstorm:
  CBOR→JSON transcode helper decision for SSE raw-payload consumers; one
  shared `fr.Recorder` demo across HTTP middleware + projections; DLQ admin
  and dead-letter-age alerting recipes; SnapshotStore/ReadModels accessor
  examples (Bundle-reachable, undocumented).
- go-health v0.3.0 aggregate/federation adoption (evaluated 2026-09-20,
  plan T25): `aggregate` merges N in-process probes into one
  go-health-compatible surface (passive, zero goroutines, "name/check"
  namespacing) — natural appkit composition is a `MountAggregate`-style
  handle next to `appkithealth.New` so a service running several surfaces
  (module probe + cqrs projection readiness + subsystem probes) can serve
  ONE /readyz; `federation` rolls up REMOTE go-health HTTP documents (5s
  fetch timeout, 1MiB cap) — that is fleet/ops-plane material, not a
  framework concern. Neither ships now: no consumer demand signal yet.
  Trigger: a real consumer needs single-endpoint multi-surface readiness →
  wrap aggregate in a Mounted-compatible handle (health module stays
  injector-free; go.mod floor would move to go-health v0.3.0, which
  REQUIRES go >= 1.27.1 — gate on the toolchain floor decision too).
