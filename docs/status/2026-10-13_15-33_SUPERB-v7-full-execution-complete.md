# SUPERB v7 — Full Execution Complete

**2026-10-13 15:33 CEST · Executor: Crush · Plan: `docs/planning/2026-10-13_14-05_SUPERB-v7-consumer-signal-lo-adoption-and-hygiene.md` (verdicts in its §6)**

All agent-executable tasks in the v7 plan are done or honestly parked.
**No releases were tagged** — every code change this session is either
test-only polish ([Unreleased] entries ride natural trains) or
guards/docs/example surface.

## The 1% (consumer signal)

- **T01 PapDashboard consumer-#2 review** — usage surface exhaustively
  read at v0.3.0: core only, two files, 8 `ServiceConfig` fields + both
  sentinels + `Mux`/`Run`; NO DrainHooks, NO testkit, a hand-rolled
  250ms `testDrainDelay`. Drift verdict v0.4.0→v0.9.0: additions-only,
  zero breakage; their `go 1.26.7` floor is the only hard refresh
  constraint. Strongest finding: their `internal/middleware` hand-rolls
  BodyLimit/APIKeyAuth/RateLimit/CSP — a near-1:1 overlap with
  go-appkit/security's battery set (first real-consumer demand evidence
  for W2). Addendum: `doc/planning/2026-09-04_papdashboard-integration.md` §7.
- **T07 claims guard** — PapDashboard leg born-red proven (a planted
  wrong claim fails on all three legs: split brain, stale tag, missing
  require). AGENTS + FEATURES anchors, keyless ls-remote newest-tag
  truth, tagged go.mod require via the sibling clone.
- **T11 v1 exit criteria graduated** DRAFT→ACTIONABLE (the draft's own
  trigger fired): 5/7 hard criteria met; both remaining blockers are
  external-clock (one additions-empty core tag pair; consumer version
  overlap + one release cycle).

## The 4% (policy compliance + hygiene)

- **T02-T04 samber/lo**: 13 deduplicated `stdlib2lo` findings triaged;
  5 test sites adopted by hand across cqrs/otel/realtime/integration
  (lo v1.53.0; cqrs promoted from indirect); 7 runtime keeps per the
  test-only boundary; 1 false positive declined with evidence
  (`slices.Backward` fold). Root module: zero sites (T03 no-op). All
  four suites + lints green; `[Unreleased]` CHANGELOG entries written.
  Real defect fixed on sight: realtime's CHANGELOG had its 0.1.3 block
  above the `# Changelog` header.
- **T05 hygiene A**: BuildFlow audit — 9 unavailable + 67 N/A all
  correct-by-design for a Go-only repo. jscpd: 115 clones measured,
  105/107 Go = accepted test class; `docs/drafts/**` ignored (body-by-
  design archives); volume documented, not tuned. NEW guard
  `check-golangci-configs.sh` (12/12 verify, negative-tested) + CI job.
- **T06 hygiene B**: `docs/agents/README.md` charter; otel doc.go now
  documents the completion-line limitation WITH fix status (also fixed
  the module-details-claims-doc.go-documents-it split brain); health
  example `-hardened` serves embedded CSS via `WithCSSPath` —
  headless-Chromium-verified: link present, cdn.tailwindcss.com count
  0, zero console violations. THREAT_MODEL updated.

## The 20% (assurance + truth)

- **T08 cqrs-htmx setup suite: FULL GREEN** hermetic (`-race -count=1`,
  1.27.2 floor, ~165 tests, 15.1s) — consumer #1's verified surface
  extended from 3/3 proxy smoke to suite-green.
- **T09 pkg.go.dev**: go-flightrecorder page **RESOLVED** (renders
  v0.2.1 — the v6-filed index requests worked); core@v0.9.0 renders
  (license-hidden godoc = the standing proprietary ruling); errorpages
  v0.1.2 pending normal <48h index lag.
- **T10 otel benchstat n=10 at 1.27.2**: B/op + allocs/op UNCHANGED
  (±0%, recorded); ns/op deferred a second time — ambient load rose
  9→26 mid-run, middleware rows ±40-45% vs the ±25% honesty bar.
- **T13/T21/T24 gated sweep**: all 4 upstream asks still OPEN.
  **BuildFlow #46-49 ALL CLOSED** (v0.8.0): the
  go-work-sync→`go_mod_ignore_dirs` swap is documented in
  `.buildflow.yml` but blocked on the system binary (764c23c predates
  v0.8.0) — owner action; jscpd #48 defaults noted, local ignore kept
  for bare runs; buildflow skill verified current (no PR needed).
  httputil **v1.5.0** tagged (MIT-license release, no
  LoggingFromContext — i18n row stays gated, re-keyed to "next feature
  tag"). Pin-drift guard: 27 OK / 0 FAIL.
- **T27 guards battery: 8/8 green** (go-directives, pin-drift,
  dependabot-parity, workspace-charter, consumer-claims,
  golangci-configs, structure-linter, example-smoke). Structure-linter
  gained a FOURTH exclusion (assets-directory) — the newer binary flags
  any .css and wants repo-root assets/, which go:embed cannot reach;
  the example CSS moved to `health/example/assets/app.css` and the rule
  is excluded with rationale. AGENTS/TODO/FEATURES truth-synced; CI on
  origin verified green except the designed ci-preflight (SSH secret,
  owner P1).

## Honesty ledger

- Daemon attribution losses: 4 commit batches raced this session
  (lo-train files, T06 docs batch, .jscpd+guard, otel bench README);
  content verified correct under daemon commits in every case.
- One self-inflicted edit round-trip on otel/metrics_test.go imports
  (gci shape — resolved by `golangci-lint fmt` as the canonical fixer)
  and one `io` import removed-then-restored (grep too narrow; the
  compile caught it).
- ns/op re-baseline deferred twice now (2026-09-29, 2026-10-13) — both
  times for genuine ambient load; the honesty bar stays.
- errorpages@v0.1.2 pkg.go.dev still pending; re-check next session.

## Still open (gated, honestly parked)

Owner: SSH secret (T12), BuildFlow binary ≥ v0.8.0 refresh, pma rulings
(T14). Upstream: #67/#7/#2/#41 (T13 land-adoptions on close), dprint
exit-14. On-trigger: i18n residue (next httputil feature tag), W3-W5
batteries (demand-gated, zero composition consumers), T08's
session-close push performed with this report.
