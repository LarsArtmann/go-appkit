# Status Report: SUPERB v14 Continued Mandate — C1 Composition Swap, the httpx Battery, and Three More Tags

**Date:** 2026-11-07 11:00 CET
**Origin tip:** `9899e45` · **Tags on origin this session:** `v0.12.0` (core), `otel/v0.4.0`, `config/v0.1.1`, `httpx/v0.1.0` — ALL consumer-verified via `scripts/verify-consumer.sh` behavioral snippets (4/4)
**Base:** the 08:50 session's green close-out (`0ebbb38`, tip-CI verified clean-except-designed-red)
**CI state at close:** tip run `42361610561` — the ONLY failed job is `ci-preflight (designed red until the SSH secret is restored)`. govulncheck live dispatch `42360609010` SUCCESS (httpx matrix leg live-proven). Guards: pin-drift 33 OK / 0 NOTEs, features-claims 8 green, consumer-claims green, tagged-changelogs 5 green, status-index live, routed-destinations 0 unenforced, sse-protocol 17 syncs green, directives 14/14, workspace charter root+13.

## a) Fully done

1. **Tip-CI verification (the prior session's one open thread):** run `42352273884` confirmed clean-except-designed-red. Closed.
2. **C1 — the composition swap SHIPPED as core v0.12.0.** `Service` now delegates the server loop to a FRESH `httputil.Server` per `Start` via `httputil.NewServerListener` — exactly the spike's plan-of-record: appkit keeps the bind + synchronous `listen_failed`; the delegated server is constructed BEFORE the hooks (pure construction — upstream rejections surface as `server.config_invalid` Rejections before any lifecycle state exists); fresh-Server-per-Start honors upstream single-serve ownership (`server.listener_consumed` unreachable in framework flow); restart-after-full-shutdown NOW SERVES AGAIN (new capability, pinned in `delegatedserver_test.go`); `Addr()`/`Running()` views, the six `shutdown phase complete` lines, and the `server.shutdown_failed` classification are byte-identical (`shutdownlog_test.go` green); `ServerConfig.ShutdownTimeout` = the SAME `ServiceConfig.ShutdownTimeout` field (no double-apply). ZERO public API delta (`go doc -all` diff EMPTY at root + testkit — the v1.0.0 freeze-cycle anchor semantics hold). G1 Core TLS `StartTLS` seam sketched in spike doc §6. Verified: full core suite `-race`, preset (47s incl. drain E2Es) + health satellites green, example-smoke GREEN (all 7 examples boot/probe/graceful-stop), AGENTS + module-details + spike doc truth-synced, TODO row closed.
3. **Otel rename train SHIPPED as otel v0.4.0.** semconv re-pinned v1.26.0 → **v1.43.0** (the latest published convention set — probed on the proxy; rides the pinned `go.opentelemetry.io/otel v1.47.0` module, no separate require since upstream folded semconv versioned packages into the main module) WITH the `deployment.environment` → `deployment.environment.name` rename TOGETHER, per the watch row's rule. `WithEnvironment` emits the new key; doc.go / deploy.md / README examples updated; `otlp_test.go` wire pins assert the new attribute; README semconv bullet rewritten (hold rationale → rename record + pre-merger-SigNoz filter migration note). Doc-only godoc diff, wire-behavior change → minor. The SigNoz watch row CLOSED (standing watch remains only for FUTURE key renames).
4. **config v0.1.1 SHIPPED.** x/sys v0.1.0 → v0.49.0 (fleet-consistent with otel's pin): GO-2026-5024 cleared ENTIRELY — the module-level informational finding is GONE from the govulncheck record (re-run clean) — plus the `collectFields` guard-clause refactor that v0.1.0 shipped lint-red. Zero API delta (doc-diff proof). Suite + lint green.
5. **C2+C3+C4 — the httpx battery module SHIPPED as httpx v0.1.0** (NEW 13th released-module-in-waiting; full satellite onboarding: go.work, dependabot, CI test matrix + fleet-lint list + govulncheck matrix, `.golangci.yml` with tagliatelle json:snake per the battery's wire stance, gosum-canonical default list, features-claims pattern registry, AGENTS module list + counts, README member count 13→14, structure-linter local run total=0).
   - **B1 `ResultHandler`**: the return-(T, error) seam — JSON 200 happy path, 204 for nil-pointer payloads, direct-write handlers keep ownership (the `resultWriter` records the first write and steps aside), `http.ResponseController` stays reachable through `Unwrap` (the SSE-stall wrapper class designed out).
   - **B2 `ResultHandlerWithValidation`**: JSON-body bind into tagged adapter types under the `bind.*` Rejection codes (`bind.empty_body`, `bind.body_too_large` at 1 MiB, `bind.unreadable_body`, `bind.invalid_body`), optional `Validate()` hook, the json/v2 case-sensitivity trap documented in code + README (camelCase wire field silently misses a snake_case tag — the CV 400 bug), wire tags snake_case by convention.
   - **B9 `WriteError`/`SanitizeErrorMessage`**: the classified error contract — status from `errorfamily.HTTPStatus` (same mapping errorpages uses), JSON field names matching errorpages' `errorResponse` (one taxonomy, two renderings), the no-leak ladder (server-side families render family defaults — raw internal text NEVER leaves the process; client-facing Rejection/Conflict text passes the leak-pattern scrub for POSIX/Windows paths, source locations, SQL, goroutine headers; `Public()` is the vouched opt-out; the `Message()` method classified errors carry is deliberately NOT honored — documented + tested; cause chains/context maps never serialize).
   - **The classification-parity ORACLE** (`parity_test.go`): drives EVERY family through BOTH httpx and the errorpages module (test-only dep) asserting the same status + contract shape — one taxonomy, two renderings, ENFORCED by test, not convention.
   - Tests: leaked-detail matrix, bind contract table, direct-write ownership, 204 nil semantics. Lint 0, suite green `-race`.
6. **C5 subtask 20.1 — the W5-C2 folded-contract DESIGN NOTE** landed (`docs/planning/2026-11-07_w5-c2-folded-contract-design-note.md`): the race timeline, the broadcast-strictly-after-fold contract, `cqrs.NotifyFolded` + `realtime.FoldedSource` API sketches adapted to the ENGINE's real `Projection` interface (the battery spec's `SubscribeFolded` upstream does not exist — the cqrs wrapper IS that equivalent), four ordering hazards (snapshot-after-signal, bounded channel + drop counting, per-projection sources, replay bursts), the CV discriminator acceptance spec (mid-fold marker must appear in the broadcast payload + the negative control), and the 20.2 train plan.
7. **The four-tag wave pushed + consumer-verified 4/4** with behavioral snippets (core: Service lifecycle + sentinels + RegisterClassifications; otel: Setup + WithEnvironment + WithoutGlobalRegistration; config: Load[T]; httpx: ResultHandler). Gosum-canonical re-run green post-push (the deferred-leg lesson from the tag-move incident APPLIED — verified BEFORE master left this time).
8. **Dependabot PR #58 triaged + merged.** Its CI red was first diagnosed as tag-visibility timing (true for the first run) but the re-run surfaced REAL onboarding gaps (see d3); after the fixes the PR ran green-except-`structure-linter` (the designed secret-less PR red). Squash-merged; integration post-bump: gosum-canonical green, suite `-race` green hermetically (the first attempt ran workspace-mode from the root and was correctly FAILED by the charter guard — the gotcha works).
9. **Release-state sync in the same trains:** AGENTS `_Release State_` (three train narratives + Latest-per-module: core v0.12.0, otel v0.4.0, httpx v0.1.0, config v0.1.1), FEATURES claims (core v0.12.0 + otel v0.4.0 first-ever claim + config v0.1.1 + httpx v0.1.0 first-ever claim; the stale LoggingFromContext known-limitation replaced with its resolution), TODO_LIST header + row closures (composition row CLOSED, SigNoz watch CLOSED, config govulncheck defer resolved).
10. **AGENTS gotchas landed:** the tag-push vs tip-CI visibility race (hit three times, diagnosis + `gh run rerun` fix + proxy-200 prevention) and the paths-ignore correction (docs-only pushes are SKIPPED by design — "verify the final pushed commit" means the final CODE-bearing tree).
11. **ci.yml carries `workflow_dispatch`** — manual runs now possible on docs-only trees.
12. **Two status reports landed + indexed** (the 10:30 interim + this close-out).

## b) Partially done

1. **httpx battery program** — B1/B2/B9 shipped in v0.1.0; B3/B4/B5/B7 are designed and queued (the README Scope section names the wave plan); B6/B8 owner-gated on H8.
2. **C5 folded contract** — 20.1 design note DONE; 20.2 implementation (cqrs `NotifyFolded` + realtime `FoldedSource` + the integration discriminator E2E through published tags) NOT STARTED — deliberate: a cross-module API train is its own session.
3. **httpx LICENSE** — added IN-TREE (fleet-identical proprietary text) but the v0.1.0 TAGGED tree lacks it; pkg.go.dev renders unlicensed until the next httpx tag carries it (the health v0.3.1→v0.3.2 precedent). Recorded as an OPEN TAIL in TODO_LIST.

## c) Not started

1. W4 batteries (worker supervisor, sqlite ops kit, polite outbound, do bridge, D4 atomic write, D6 webhook, D7 idempotency decision).
2. W5-C1 drop/backpressure counters + W5-C3 per-subscriber auth+filter.
3. A1 go-service-template (owner-gated: repo creation), A2 dogfood, A3–A6 consumer-repo recipes.
4. C6 joint absorb (upstream-gated: catalog's templ-migration tag).
5. L8 docs-gates drain + upstream notes; L9 machine ops (🔒 owner authority); L19 watch batch; L20 six-month adoption re-sweep.
6. L13 realtime subscribe-before-flush; L14 systemd remainder; L15 cqrs recipes — all demand-gated.

## d) Totally fucked up (all caught + fixed in-session)

1. **The `writeResult` generic-erasure bug (httpx).** `isNilResult[T]` saw `reflect.Kind` Interface at the `writeResult(any)` boundary and never detected nil pointers — the 204 test caught it on first run. Fix: the check takes `any`, handles untyped nil + nil pointers, with a comment explaining WHY it must not be generic. Lesson: type parameters erase at `any` boundaries; nil-detection helpers should take `any` from birth.
2. **The `Message()` ladder bypass (httpx B9).** The first sanitizer honored error-family's `Message()` method as a "clean message" — which let AUTHORED text bypass the family gate entirely (my own client-facing test exposed it: the scrub never ran because classified errors implement `Message()`). Fix: `Public()` is the ONLY explicit opt-out; classified server families always render the family default; client families get scrubbed raw text; the non-honoring is now a documented, tested stance. Lesson: error-family's `*Error` carries `Message() string` — ANY interface-ladder sanitizer must probe the family gate FIRST or the ladder is a bypass.
3. **The httpx onboarding was incomplete — CI caught BOTH gaps.** I extended `check-features-claims.sh`'s pattern registry but FORGOT `check-pin-drift.sh`'s (a SECOND copy of the same registry lives in another script) and forgot the module LICENSE (the module-doc-claims guard demands one at every released module root). The PR re-run surfaced both. Lesson: "new module" onboarding touches MULTIPLE guard scripts with duplicated pattern registries — there is no checklist; the guards caught what I missed (they work), but the fix loop cost a full CI round-trip each.
4. **Stale-diagnostic whack-a-mole during the httpx lint loop (~9 rounds).** The LSP lint service kept reporting STALE findings (files I had already fixed) while the CLI lagged behind edits made via `sed`/python (which update mtime but not the LSP cache), and several of my python patches applied to views that had drifted (the tooling refused them — caught) or misfired (the `errors.Is` hack, the undefined `nilResultIsThePoint` identifier, the duplicate top-level `linters-settings:` key that silently did NOT activate tagliatelle's json:snake until merged into the nested `linters.settings`). Lesson: after mechanical multi-edit passes, re-read the file before the next patch; config duplication fails SILENTLY in yaml (last-wins/first-wins is not checked by `config verify` for nested-vs-top-level key typos) — verify the linter BEHAVIOR, not the config text.
5. **Workspace-mode integration run.** Ran `go test ./integration/...` from the repo root — setup failed instantly because the charter guard's `TestMain` correctly rejects workspace-governed builds. The AGENTS gotcha worked exactly as designed; my fingers didn't. Re-ran hermetically from `integration/` per its README: green.
6. **The `git pull --ff-only` divergence after the PR merge.** Local master had my TODO-list commit; origin gained the squash-merge — pull refused, fixed with `git rebase origin/master` then push. Lesson: after merging a PR remotely, REBASE local before pulling; ff-only pull assumes no local commits.
7. **Two misdiagnoses of CI reds (serially, on PR #58).** First pass: "tag-visibility timing" — true for run 1, but I stopped digging; the re-run proved REAL causes existed (d3). Second pass: "dropped push events" for the missing runs on ce6f453/3e850f4 — wrong again; ci.yml's `paths-ignore` SKIPS docs-only pushes by design. Lesson (now an AGENTS gotcha): diagnose missing/unexpected CI runs by reading the trigger's paths-ignore filter FIRST, and never accept a timing diagnosis until a re-run comes back clean.
8. **The 08:50 session's snippet-format constraint cost a round-trip:** the fresh-consumer snippets must reference the injected `consumer.` import identifier, not the package's real name (core's snippet said `appkit.` → undefined). Caught by compile, fixed by sed.

## e) Improvements (for the repo + the process)

1. **`ci.yml` gained `workflow_dispatch`** — docs-only trees can now be CI-verified on demand.
2. **Guard registries extended** (`check-features-claims.sh`, `check-pin-drift.sh`, `check-gosum-canonical.sh`) — and the duplication itself is now a KNOWN trap recorded in AGENTS (a "new module" checklist script is the real fix; see f48).
3. **AGENTS gotchas:** the tag-push visibility race (diagnosis recipe + prevention) and the paths-ignore diagnosis rule.
4. **The parity-oracle test pattern** (`parity_test.go`: drive BOTH surfaces over the whole family set) is now a fleet asset — reusable for any future second rendering of the taxonomy.
5. **Fresh-consumer behavioral snippets** for all four new tags are written and reproducible (`scripts/verify-consumer.sh` + `/tmp/consumer-snippets/` — should be checked into `scripts/consumer-snippets/` as fixtures, see f46).
6. **The folded-contract design note** grounds the battery spec's sketch in the engine's REAL API — the next session implements against facts, not guesses.
7. **Spike doc §6** records the swap + G1 sketch — the composition plan-of-record is now fully executed history.

## f) Next tasks (ranked, with effort estimates)

1. **httpx v0.1.1** — carry the LICENSE (tagged tree lacks it; pkg.go.dev renders unlicensed). 15m + ritual.
2. **C5 20.2a** — `cqrs.NotifyFolded` wrapper + `FoldedEvent` + non-blocking bounded send + drop counter surface (additions-only). 60–90m.
3. **C5 20.2b** — `realtime.FoldedSource` (snapshot-after-signal, per-projection goroutine, the full-state re-sync invariant). 60m.
4. **C5 20.2c** — integration discriminator E2E through published tags (mid-fold marker + negative control). 45m after both tags.
5. **httpx B3** — ResponseWriter contract battery (`AssertForwardsInterfaces` must fail on the buggy notFoundInterceptor shape). 60m.
6. **httpx B4** — conditional GET (promote go-etag indirect; sha256 ETag; RFC 9110 weak/comma/`*` matrix). 60m.
7. **httpx B5** — content negotiation (`Negotiate` on go-codec — the promoted seam). 60m.
8. **httpx B7** — per-route write deadlines (outside compression; the ordering regression test). 45m.
9. **httpx fuzzing** — port CV's `result_handler_fuzz_test.go` oracle for the response path. 45m.
10. **H8 owner decision** → if adopted: core exposes the `httputil/endpoints` Registry (same-train ship pre-verified). Blocks f11/f12.
11. **httpx B6** — route introspection + golden-route test (needs H8). 60m.
12. **httpx B8** — route metadata → docs feed (needs H8). 60m.
13. **Auth-bridge owner decision** → staged path: usermgmt-recipe now, session battery demand-gated. Blocks the preset auth leg.
14. **A1** — go-service-template repo (owner creates it; skeleton per the scaffold verdict doc). 90m once created.
15. **A2** — dogfood app from the template (<40 LOC acceptance). 60m after A1.
16. **A3** — dnsblockd battery-swap recipe (−~1,600 LOC exemplar). 60m, in their repo.
17. **A4** — ChastityAPI recipe (−~4,800 LOC). 90m, in their repo.
18. **A5** — PapDashboard security swap recipe. 60m, in their repo.
19. **A6** — pbx-lens cmdguard migration recipe. 30m, in their repo.
20. **C6** — joint go-sse v0.7.0 / cqrs-htmx v4.16.0 / templ-components v1.27.0 absorb (▲ upstream-gated: catalog ships a migrated templ tag; then bump trio + delete hold entries + hub.go doc swap). 30m when unblocked.
21. **W4 worker** — supervisor + pool battery. 100m.
22. **W4 sqlite ops kit** — lease/backup/ledger. 100m.
23. **W4 polite** — outbound client (+ go-etag client+metrics v0.6.0 fit). 60m.
24. **W4 do bridge** — samber/do integration module. 60m.
25. **W4 D4** — atomic file write (go-atomic-write ≥ v0.5.1 floor, Windows syscall trap). 45m.
26. **W4 D6** — webhook battery. 60m.
27. **W4 D7** — idempotency adopt-wire vs 165-LOC port decision note. 30m.
28. **W5-C1** — drop/backpressure counters (`WithOnDrop`; dedupe with ssehub's `Drops()` shape). 60m.
29. **W5-C3** — per-subscriber auth+filter. 60m.
30. **v1.0.0 freeze-cycle process** — core v0.12.0's zero-diff diff means the anchor semantics hold; draft the freeze-cycle operating doc (what the next zero-API-diff tag must look like, who declares the freeze). 45m.
31. **New-module onboarding checklist script** — a `scripts/new-module.sh <name>` that patches ALL the registries (pin-drift, features-claims, gosum-canonical, dependabot, ci.yml ×3, go.work, LICENSE stub, AGENTS counts) — the d3 trap automated away. 90m.
32. **Check in the consumer snippets** as repo fixtures (`scripts/consumer-snippets/{core,otel,config,httpx}.go`) so verify-consumer runs are one-liners per module. 20m.
33. **preset otel leg** — verify the env passthrough emits `deployment.environment.name` post-v0.4.0 (preset still pins otel v0.3.0; its next train picks the rename up). 20m check.
34. **govulncheck nightly** — httpx leg live-proven ✓ (42360609010); nothing pending, but add httpx to the workflow's module matrix documentation if the matrix list is duplicated anywhere. 10m audit.
35. **L8a** — docs-gates annotate_max drain (per-cell strikes → lower the 255 pin). 30m, in docs-gates repo.
36. **L8b** — upstream docs-gates notes (routed≠resolved, compound-id quirk, DF boundary). 30m.
37. **L13** — realtime subscribe-before-flush (sketch exists; demand-gated). 60m on demand.
38. **L14a** — systemd `Install` returns Hooks. 20m, next systemd train.
39. **L14b** — systemd `WithStatus`. 20m, same train.
40. **L15** — cqrs EventConfig opt-in recipes (idempotency middleware via CommandMiddleware). 60m on demand.
41. **L19** — watch batch: otel v1.48 (removes OTEL_GO_X_CARDINALITY_LIMIT), OTLP-logs gate, SigNoz live validation, cordis trigger, SSE-hub convergence, nixpkgs pull re-check. 30m batch.
42. **L20** — six-month adoption re-sweep (`scripts/count-adopters.sh` + archetype matrix delta). 60m, scheduled.
43. **CI structure-linter on PRs** — verify the designed secret-less red posture is documented where contributors will see it (ci.yml comment). 10m.
44. **example-smoke** — consider an httpx example leg (8th). 30m.
45. **Root README module table** — confirm httpx has a row (the member count was bumped; verify the table itself). 10m.
46. **Fleet verify-all with lint** — run the full 13-module lint sweep once post-httpx (the CI fleet-lint covers it, but the local ritual proof is the recorded bar). 5m (it ran for httpx only so far).
47. **SSH secret restore (owner)** — ci-preflight stays designed-red; structure-linter fails on secret-less PR runs until restored. OWNER ACTION.
48. **TODO_LIST hygiene** — the W3-W5 row still narrates httpx as "new module (B1...)"; update to shipped-state + point at the README Scope. 10m.
49. **`docs/review/libs/` errata** — server_timing/go-codec errata blocks may reference the composition row as "pending"; sweep for stale pointers post-C1. 20m.
50. **Session-close ritual** — re-verify the final CI run after the daemon's next auto-commit cycle (docs-only pushes are skipped by design; the 9899e45 run IS the final code-bearing verdict). 5m.

## g) Up to 3 questions I cannot figure out myself

1. **SSH secret restore** — `ci-preflight` remains designed-red and `structure-linter` fails every secret-less PR run (it correctly refuses a vacuous green). Restoring `SSH_PRIVATE_KEY` (owner UI) turns both green and retires the only permanent red. This is a GitHub settings action only you can do.
2. **H8 route seam** — do I build B6/B8 now (core exposes the `httputil/endpoints` Registry — the memo's recommendation, pre-verified for a same-train ship), or do the seam-free batteries (B3/B4/B5/B7) first and B6/B8 after your ruling? The memo: `docs/planning/2026-11-07_h8-route-seam-owner-ask.md`.
3. **Auth bridge** — do I proceed with the staged path from the memo (usermgmt recipe now; a security session battery demand-gated at ≥3 committed consumers), or do you want the battery built speculatively? The memo: `docs/planning/2026-11-07_auth-bridge-owner-decision-memo.md`.
