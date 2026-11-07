# Status Report: C1 Composition Swap, the httpx Battery, and Three More Tags — the Continued-Mandate Session

**Date:** 2026-11-07 ~10:30 CET · **Session:** continued-mandate execution (the owner's
"WHOLE TODO LIST" instruction carried the session past the 08:50 wave report)
· **Origin tip:** `85906e1` (+ the tag set below) · **Base:** the 08:50
session's green close-out (tip `0ebbb38`, tip-CI verified clean-except-designed-red).

## a) DONE (shipped, verified)

- **Tip-CI verification (the 08:50 session's one open thread):** run
  `42352273884` confirmed — the ONLY failed job was `ci-preflight`
  (designed red, SSH secret). Closed.
- **C1 — the composition swap SHIPPED as core v0.12.0** (the spike's
  plan-of-record, executed verbatim): `Service` now delegates the server
  loop to a FRESH `httputil.Server` per `Start` via `NewServerListener` —
  appkit keeps the bind + sync `listen_failed`, constructs the delegated
  server BEFORE the hooks (upstream rejections surface as
  `server.config_invalid` Rejections before any lifecycle state exists),
  fresh-server-per-Start honors upstream single-serve ownership, and
  restart-after-full-shutdown NOW SERVES AGAIN (pinned,
  `delegatedserver_test.go`). `Addr()`/`Running()` views, the six
  `shutdown phase complete` lines, and the `server.shutdown_failed`
  classification are byte-identical (`shutdownlog_test.go` green). ZERO
  public API delta (`go doc -all` diff EMPTY — the v1.0.0 freeze-cycle
  anchor semantics hold). G1 Core TLS `StartTLS` seam sketched in spike §6.
  Verified: full core suite `-race`, preset (47s incl. drain E2Es) + health
  satellites, example-smoke GREEN (7 examples boot/probe/graceful-stop),
  spike doc §5-§6 + AGENTS + module-details truth-synced.
- **Otel rename train SHIPPED as otel v0.4.0:** semconv re-pinned
  **v1.26.0 → v1.43.0** (latest published convention set; rides the pinned
  `go.opentelemetry.io/otel v1.47.0` module — upstream folded semconv
  versioned packages into the main module, no separate require) WITH the
  `deployment.environment` → `deployment.environment.name` rename TOGETHER,
  per the watch row's rule. `WithEnvironment` emits the new key; doc.go /
  deploy.md / README examples updated; `otlp_test.go` wire pins assert the
  new attribute; README semconv bullet rewritten (hold rationale → rename
  record + pre-merger-SigNoz filter migration note). Doc-only godoc diff,
  wire-behavior change → minor. The TODO watch row CLOSED.
- **config v0.1.1:** x/sys v0.1.0 → v0.49.0 (fleet-consistent with otel's
  pin) — GO-2026-5024 cleared ENTIRELY (the module-level informational
  finding is gone from the scan record) + the `collectFields` complexity
  refactor that v0.1.0 shipped lint-red. Zero API delta.
- **C2+C3+C4 — the httpx battery module SHIPPED as httpx v0.1.0** (NEW
  module, full satellite onboarding): `ResultHandler` (JSON 200 / 204
  nil-pointer / direct-write handlers keep ownership /
  `http.ResponseController` reachable through `Unwrap` — the SSE-stall
  wrapper class designed out), `ResultHandlerWithValidation` (bind into
  tagged adapter types; `bind.*` Rejection codes; 1 MiB cap; `Validate()`
  hook; the json/v2 case-sensitivity trap documented in code + README),
  `WriteError`/`SanitizeErrorMessage` (classified contract with errorpages'
  JSON field names; server-side families render family defaults — raw
  internal text never leaves the process; client-facing text passes the
  leak-pattern scrub; `Public()` is the vouched opt-out; cause chains and
  context maps never serialize), and the **classification-parity ORACLE**:
  `parity_test.go` drives every family through BOTH httpx and the errorpages
  module (test-only dep) asserting the same status + contract shape — one
  taxonomy, two renderings, ENFORCED. Onboarding: go.work, dependabot,
  CI test matrix + fleet-lint list + govulncheck matrix, `.golangci.yml`
  (tagliatelle json:snake per the battery's wire stance), gosum-canonical
  default list, features-claims pattern, AGENTS module list + counts,
  README member count 13→14. Lint 0; suite green `-race`; structure-linter
  local run total=0.
- **C5 subtask 20.1 — the W5-C2 folded-contract DESIGN NOTE** landed
  (`docs/planning/2026-11-07_w5-c2-folded-contract-design-note.md`): the
  race timeline, the broadcast-strictly-after-fold contract,
  `cqrs.NotifyFolded` + `realtime.FoldedSource` API sketches adapted to the
  engine's real `Projection` interface (the spec's `SubscribeFolded`
  upstream does not exist — the cqrs wrapper IS that equivalent), four
  ordering hazards, the CV discriminator acceptance spec (mid-fold marker
  must appear in the broadcast payload + the negative control), and the
  20.2 train plan. NOT STARTED: the 20.2 implementation (deliberate — a
  cross-module API train through published tags is its own session).
- **Four tags pushed + consumer-verified** (fresh-consumer behavioral
  snippets, `scripts/verify-consumer.sh`): core v0.12.0 (Service lifecycle
  + sentinels + RegisterClassifications), otel v0.4.0 (Setup +
  WithEnvironment + WithoutGlobalRegistration), config v0.1.1 (Load[T]),
  httpx v0.1.0 (ResultHandler). Gosum-canonical green post-push (the
  deferred-leg lesson from the tag-move incident APPLIED — verified before
  master left). Release-state sync (AGENTS `_Release State_` + FEATURES
  claims + TODO header + claims-guard httpx pattern + gosum-canonical
  module list) landed in the same trains.
- **Incident (small, caught + fixed):** the mid-session tip run `42355646085`
  failed pin-drift + module-doc-claims — TAG-VISIBILITY TIMING (the run's
  checkout preceded the otel/config tag pushes by ~3 minutes; v0.12.0
  passed because the daemon had pushed it at creation). Re-run of the same
  commit → clean-except-designed-red. Lesson candidate (see e).

## b) PARTIAL

- None. Everything started in this session reached its verified end state.

## c) NOT STARTED (deliberate, per the plan's own guardrails)

- **C5 20.2** — the folded-contract implementation (cqrs `NotifyFolded` +
  realtime `FoldedSource` + the integration discriminator E2E through
  published tags): the design note is the plan-of-record; the
  implementation is a cross-module API train (next session opener alongside
  the W3 remainder).
- **C6** — the joint go-sse/cqrs-htmx/templ-components absorb: still
  upstream-gated (▲ catalog's migrated templ tag).
- **W3 remainder** — B3 ResponseWriter contract, B4 conditional GET, B5
  negotiation, B7 write deadlines (B6/B8 owner-gated on H8): httpx v0.1.0
  carries B1/B2/B9 only; the module README's Scope section names the wave
  plan.
- **W4/W5 remainder, L8-L20 rows** — unchanged from the 08:50 report's
  queue (machine-ops 🔒, owner rulings ○, watches, demand-gated batteries).
- **Dependabot PR #58** (integration minor-and-patch, non-appkit deps):
  its first run failed on the same tag-visibility timing; branch updated +
  re-run launched. Verdict: check before close-out (row below).

## d) FUCKED UP (and caught)

- **The writeResult generic-erasure bug** (httpx): `isNilResult[T]` saw
  `Kind Interface` at the `writeResult(any)` boundary and never detected
  nil pointers — the 204 test caught it immediately. Fix: the check takes
  `any` and handles both untyped nil and nil pointers, with a comment
  explaining WHY it must not be generic.
- **The `Message()` ladder bypass** (httpx B9): the first sanitizer honored
  error-family's `Message()` method as a "clean message" — which let
  authored text bypass the family gate entirely (my own client-facing test
  exposed it: the scrub never ran). Fix: `Public()` is the ONLY explicit
  opt-out; classified server families always render the family default;
  client families get scrubbed raw text. The `Message()` non-honoring is
  now a documented, tested stance.
- **Two yaml/python edits applied to a stale view** (the file-changed-since-
  read dance with the daemon racing): both caught by the tooling, re-read,
  re-applied cleanly. No damage.
- **The tag-object-vs-commit sha misread** during the tag wave (briefly
  concluded the daemon had committed BETWEEN tag commands; actually all
  three tags peeled to the same commit). Caught by the tree-verification
  step before any push — the verify-each-tag's-tree discipline prevented a
  phantom incident.

## e) IMPROVEMENTS SHIPPED

- `scripts/check-features-claims.sh`: httpx added to the tag_pattern
  registry (a new module's FIRST claim otherwise dead-ends in an opaque
  FAIL).
- `scripts/check-gosum-canonical.sh`: httpx added to the default module
  list (auto-discovery is NOT how this guard works — a new module must be
  added by hand; now recorded here).
- **Timing-class lesson (candidate for AGENTS gotchas, not yet landed):**
  CI runs triggered within minutes of a tag push race the tag's visibility
  in the runner's fetch — pin-drift/module-doc-claims/proxy-smoke fail
  with "claims newer than tags" ghosts. Mitigation used twice this session:
  verify the failure is the timing class (`gh run view --log-failed` shows
  claim-vs-tag mismatches ONLY for tags pushed after the run started), then
  `gh run rerun` the same commit. A standing rule "after pushing tags, wait
  for proxy-200 BEFORE expecting the tip run green" is the prevention.

## f) NEXT (in queue, per TODO_LIST)

1. **C5 20.2** — folded-contract implementation per the design note
   (cqrs train → realtime train → integration discriminator E2E).
2. **W3 remainder** — B3/B4/B5/B7 into httpx (one train each or batched).
3. **H8 + auth-bridge owner decisions** — both memos packaged and open;
   B6/B8 and the preset auth leg gate on them.
4. The rest of the 08:50 queue (A1 template repo when the owner creates
   it, consumer-repo recipes, watches, machine-ops 🔒).

## g) QUESTIONS for the owner (carried — all recorded in TODO_LIST)

1. **H8 route seam** — core exposes `httputil/endpoints` Registry?
   (`docs/planning/2026-11-07_h8-route-seam-owner-ask.md`)
2. **Auth bridge** — staged path (recipe now, session battery demand-gated)?
   (`docs/planning/2026-11-07_auth-bridge-owner-decision-memo.md`)
3. **Push-discipline standing rule** (from the 08:50 report) + the
   machine-ops authority (L9 🔒).
