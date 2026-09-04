# Status Report: cqrs-htmx Leverage Session

**Date:** 2026-09-04 17:38 CEST
**Session scope:** "How would you better leverage /home/lars/projects/cqrs-htmx" — research, execution, verification in go-appkit.
**Repos touched:** `go-appkit` (code + docs). `cqrs-htmx` read-only.
**Format note:** user explicitly requested `.md`; the status-report skill's HTML default was overridden per instruction.

---

## TL;DR

The session turned the cqrs-htmx relationship from *documented* into *enforced*: verified all 5 released modules as fresh proxy consumers, discovered and mechanically fixed a pkg.go.dev packaging failure (submodule pages 404 for missing module-root LICENSE files; core godoc hidden by the unclassifiable proprietary license), built a new `integration/` E2E test module that locks the exact composition seams cqrs-htmx's ADR-001 adoption depends on, and corrected 19 days of stale "push pending" state in the repo's gate docs. The single biggest open item is a **license decision that only the user can make** — until then, the project's public storefront (pkg.go.dev) stays dark.

### Self-critique (asked directly: what did you forget / do badly / still owe?)

- **Forgot to read the handler contract before testing it.** I wrote the journal-replay test assuming cold-start replay; `realtime/handler.go:191` returns early on a zero Last-Event-ID by design. Cost: a 10s test failure and a rewrite. Should have read `replayMissedEvents` first — the rule "read the early returns before writing an API-contract test" is now obvious in hindsight.
- **Sloppy multi-file mechanics.** The `connect()` signature refactor landed in two passes (vet caught orphaned call sites), a malformed printf generated a broken `main.go` on the first proxy-smoke run, one AGENTS.md edit mangled a sentence and needed a repair pass, and one stray bogus tool call (`read_mcp_resource` to a nonexistent server) was fired early. Each cost a round trip; none damaged state.
- **Forgot that "done" needs a CHANGELOG entry in this repo.** The auto-commit daemon committed the integration module and LICENSE files with heuristic messages; no CHANGELOG rows exist for them yet (routed to section f).
- **Still owe:** the license decision is surfaced but unmade; the integration module covers 2 of ~5 known cross-repo seams; nothing in this session verified the `RunWithAppkit` full parity suite (only its two most fragile assertions were ported).

---

## a) FULLY DONE

| # | Work | Evidence |
| - | ---- | -------- |
| 1 | Fresh-consumer proxy smoke test for all five released modules (core v0.3.0, cqrs v0.3.0, realtime v0.1.0, flightrecorder v0.1.0, flightrecorderhealth v0.1.0): clean /tmp module → `go get` → blank import → `go build` | All 5 PASS, plain build (no `GOEXPERIMENT` needed even for the jsonv2 modules — blank-import builds don't hit json/v2 paths) |
| 2 | Root-caused the pkg.go.dev failure: all 7 submodule pages 404 (no LICENSE inside any module root — pkg.go.dev does not index such modules) and core renders `License: UNKNOWN` with godoc hidden (PROPRIETARY text is unclassifiable) | Live fetches of `pkg.go.dev/github.com/larsartmann/go-appkit{@v0.3.0,/cqrs@v0.3.0,/realtime@v0.1.0,/flightrecorderhealth@v0.1.0}` — root renders w/ UNKNOWN; all submodules 404 |
| 3 | Mechanical fix: LICENSE copied into all 7 module roots (`cqrs/`, `docs-mod/`, `errorpages/`, `flightrecorder/`, `flightrecorderhealth/`, `otel/`, `realtime/`), mirroring the cqrs-htmx family pattern | Files present; committed by auto-commit daemon; takes effect on pkg.go.dev with the next tagged versions |
| 4 | New `integration/` module (`github.com/larsartmann/go-appkit/integration`): `go.mod` (pins PUBLISHED tags only), `doc.go`, `integration_test.go` with 2 E2E tests, `.golangci.yml` (cloned from realtime, go 1.26.7), `go.work` entry | `go test -race -count=5` green; plain and jsonv2 builds both green; `golangci-lint` 0 issues; gofmt clean; workspace root `go build ./...` green |
| 5 | `TestSSEHeadersFlushThroughAppkitDefaultStack` — ports the cqrs-htmx ADR-001 spike's M18.3 flush test: headers must arrive well before the first (400ms-delayed) event through appkit's full default middleware stack | PASS 5× under `-race`; the exact scenario setup's `run_appkit_test.go` validates |
| 6 | `TestJournalBackedReplayThroughAppkitService` — locks the cross-repo contract: no cold-start replay (zero Last-Event-ID), exact missed-suffix replay on reconnect via `transport.NewJournalSSEStore`, live broadcast interleave, all through an appkit `Service` | PASS 5× under `-race`; first executable proof of the wiring that previously existed only as an AGENTS.md gotcha |
| 7 | FEATURES.md "Consumers" section citing cqrs-htmx `setup` as reference consumer (ADR-001, consumed version, fold-in status, proxy verification date) | File edited; committed |
| 8 | AGENTS.md de-staled and extended: 10-module list, "push pending" → pushed/proxy-verified, Release State rewritten with post-push results + pkg.go.dev gap, integration build commands, realtime journal-replay gotcha marked end-to-end verified, new "Integration Module — Code Organization" section | File edited; committed |
| 9 | TODO_LIST.md truthed: removed completed P1 push item + P2 FEATURES/flush items, converted P1 post-push to partial-with-evidence, added license-decision item, fixed stale "after the pending push" conditionals, module count 9→10 | File edited |
| 10 | Verified the wave-push claim across repos before editing docs: `git ls-remote` (all 5 tags on origin at `f938d65`), cqrs-htmx `setup/go.mod` resolves `go-appkit v0.3.0` from the proxy with replace stripped, their revalidation doc (2026-08-30) confirms pushed | Command outputs in session log |

## b) PARTIALLY DONE

| # | Work | What works | What remains | Blocker | Effort |
| - | ---- | ---------- | ------------ | ------- | ------ |
| 1 | P1 post-push verification | Proxy smoke 5/5 PASS (today) | pkg.go.dev rendering check for future tags; today's check found the license gap instead of green | Next tag wave must include the new LICENSE files; classification needs the license decision | M |
| 2 | pkg.go.dev repair | LICENSE files in all module roots (landed today) | Submodule pages stay 404 and core godoc stays hidden until (a) license decision, (b) re-tag, (c) re-crawl | User decision (license posture) | S after decision |
| 3 | Integration module seam coverage | 2 seams enforced (SSE flush, journal replay) | 3 known seams untested: otel one-Setup-per-process across go-appkit/otel + go-cqrs-lite/otel; cqrs `OTelProjectionMetrics` end-to-end; errorpages family→status parity against `appkit.HTTPStatus` | None — pure scope | M each |
| 4 | Consumer-contract protection | The two most fragile adoption assertions (flush, replay cursor) are now regression-locked | No test suite mirrors the full `RunWithAppkit` parity set (readiness composition, drain transitions, response parity, bench) — those live only in cqrs-htmx's spike tests | None | M |

## c) NOT STARTED

| # | Item | Why not started | Still wanted? |
| - | ---- | --------------- | ------------- |
| 1 | License decision (PROPRIETARY vs standard OSS license) | User gate — legal/business call, not mine to make | Yes — it gates the entire public-adoption story |
| 2 | cqrs-htmx-side ADR-001 fold-in (`RunHandler` internals swap, chain dedupe, `/health` semantics, logging posture) | Their repo, their task; unblocked since the wave pushed | Yes — it is THE consumer milestone |
| 3 | Logging posture decision (comparison finding 7) | Deliberate user decision in TODO_LIST P2 | Yes |
| 4 | Tagging `otel/v0.1.0` and releasing the health work (`health/v0.1.0`, `flightrecorderhealth/v0.1.1`) | Queued next wave; pre-tag steps documented | Yes |
| 5 | Core composition-contract test suite (mirroring run_appkit parity: ReadyCheck composition, drain transitions, response parity) | Considered, deferred — integration module started with the two highest-value seams | Yes, fold into f-3 |
| 6 | The parallel session's cqrs work (`EventConfig.FlightRecorder` → `go-flightrecorder` migration: `eventservice.go`, `flightrecorder_test.go`, `go.mod` appeared modified mid-session) | Deliberately untouched — not authored by this session; investigated and left alone per safety rules | Their session owns it |
| 7 | README `GOEXPERIMENT=jsonv2` documentation; otel benchmark; dprint fix; go-structure-linter config; httputil Logging upstream proposal; v1.0.0 exit criteria; toolchain 1.26.6 bump | Pre-existing TODO_LIST items outside this session's scope | Yes (tracked) |

## d) TOTALLY FUCKED UP

| # | What is broken | Severity | Root cause | Mitigation |
| - | -------------- | -------- | ---------- | ---------- |
| 1 | **The project's public storefront is dark.** pkg.go.dev: all 7 submodule pages 404; core page hides all godoc (`License: UNKNOWN`). Any potential consumer evaluating go-appkit sees nothing. | High (adoption-blocking, not build-breaking) | No LICENSE inside module roots + unclassifiable proprietary license text at root | Mechanical part fixed today (LICENSE in every module root). Full fix gated on the license decision + next tag wave |
| 2 | **License inconsistency across the ecosystem.** go-appkit is PROPRIETARY ("all rights reserved") while cqrs-htmx — the reference consumer that *depends on it* — is MIT and ships in consumer builds. Every cqrs-htmx downstream consumer transitively includes an all-rights-reserved dependency. | High (legal murkiness for the entire ecosystem) | License choice predates the adoption | Only resolvable by the user's license decision |
| 3 | **Gate-doc rot: "push pending" stood for 19 days (Aug 16 → Sep 4) after the wave was pushed (Aug 30).** Three sessions' worth of docs (AGENTS.md, TODO_LIST.md, P1 items) described a done action as pending, and my session started from that stale premise. | Medium (wasted effort, wrong decisions possible — e.g. re-pushing or re-verifying blindly) | The push was executed without flipping the gate docs; no "update gate docs at push time" step existed | Fixed today across AGENTS.md + TODO_LIST.md; process improvement in (e) |
| 4 | **This session's own stumbles** (all recovered, none shipped broken): broken `main.go` in first smoke run (missing package clause); journal-replay test written against an assumed contract → 10s failure + rewrite; `connect()` refactor needed a second vet pass; one AGENTS.md edit truncated a sentence; one bogus tool call to a nonexistent MCP server | Low (session-local waste, ~5 round trips) | Speed over read-first; multi-file refactor split across tool calls | Process fixes in (e) |
| 5 | **Lint-config drift noticed, partially unfixed:** `realtime/.golangci.yml` still declares `go: 1.26.5` on a 1.26.7 repo (my clone bumped integration's copy only). Other module configs may carry the same stale pin. | Low | The toolchain bump commit didn't sweep linter configs | f-9/f-10 |

## e) WHAT WE SHOULD IMPROVE

1. **Read the contract before testing it.** API-contract tests must start from the handler's early returns (`grep -n "return" handler.go` takes 30 seconds). Would have saved the rewrite entirely.
2. **One-shot multi-file refactors.** Signature changes + all call sites belong in a single multiedit, then vet. Splitting them invites orphaned call sites.
3. **Executable proof over documented recipes.** The JournalSSEStore wiring sat as a prose gotcha for weeks with zero verification. Any "wire X into Y" note in AGENTS.md should either have a test in `integration/` or say why it can't. Generalize: documented integration = untested integration.
4. **Gate docs flip at gate time.** Pushing a release wave must include editing TODO_LIST/AGENTS in the same session. The auto-commit daemon makes this nearly free; 19 days of rot was a choice by omission.
5. **pkg.go.dev health belongs in release verification.** go-release's post-push phases check rendering but not license classification. A "fetch each module page, assert no 404 and license != UNKNOWN" step would have caught issue (d1) at the first push. Candidate skill improvement.
6. **Cross-repo truth sourcing.** cqrs-htmx's revalidation doc was *fresher* than go-appkit's own gate docs. When a gate spans repos, check both repos' docs before trusting either.
7. **Session hygiene:** the malformed printf and the bogus tool call were both "move fast, verify never". Cheap fix: dry-run generated content with `cat` before feeding it to a compiler/test run.

## f) Top 50 things we should get done next

> Brainstorm ranked by impact; most beyond the top ~15 are ROADMAP fuel. Feeds `docs-health` HARVEST — route with rigor, don't blanket-import. Impact: C/M/H/L = Critical/Medium/High/Low. Effort: S/M/L.

| # | Task | Impact | Effort | Category |
| - | ---- | ------ | ------ | -------- |
| 1 | Decide the license posture: standard OSS license (MIT, matching cqrs-htmx) vs stay proprietary | Critical | S | Legal/Docs |
| 2 | Cut the next tag wave including the new per-module LICENSE files; re-verify pkg.go.dev renders every module with visible godoc | Critical | M | Release |
| 3 | Build the core composition-contract suite in `integration/` mirroring `RunWithAppkit` parity (readiness composition, drain transitions, response parity) | High | M | Quality |
| 4 | Set up CI (GitHub Actions): per-module matrix build + vet + `-race`; zero CI config exists today | High | M | Quality |
| 5 | Integration test: otel one-`Setup`-per-process contract across go-appkit/otel + go-cqrs-lite/otel globals | High | M | Quality |
| 6 | Integration test: `cqrs.OTelProjectionMetrics` end-to-end against the otel module's meter provider | High | M | Quality |
| 7 | Integration test: errorpages family→status parity against `appkit.HTTPStatus` through a live service | Medium | S | Quality |
| 8 | ADR: "what appkit promises its consumers" — the API/lifecycle contract cqrs-htmx-style consumers may rely on | High | M | Documentation |
| 9 | Tag `otel/v0.1.0`: drop the example's `replace ../`, hermetic verify, annotated tag, fresh-consumer + pkg.go.dev check | High | M | Release |
| 10 | Release the health work: `health/v0.1.0` + `flightrecorderhealth/v0.1.1` (jsonv2 build-contract change in release notes) | High | M | Release |
| 11 | Decide the logging posture (finding 7): level config / sampling / consumer-wired logger + benchstat | High | M | Feature |
| 12 | Add a mechanical API-break check (goapidiff or `go doc` snapshot diff) to the release process | High | M | Quality |
| 13 | Land (or explicitly hand back) the parallel session's cqrs flightrecorder migration — `cqrs/` currently has uncommitted WIP in the working tree | High | M | Feature |
| 14 | Write CHANGELOG entries for this session: integration module, per-module LICENSE files, docs corrections | Medium | S | Documentation |
| 15 | ADR for the `integration/` module: never released, pins published tags, go.work membership | Medium | S | Documentation |
| 16 | Turn the proxy smoke test into a repeatable script (`scripts/verify-consumers.sh` over a module@version list) | Medium | S | Tooling |
| 17 | Define core v1.0.0 exit criteria (absorb `OuterMiddlewares`/`ShutdownHooks` discussion) | High | M | Planning |
| 18 | README: document `GOEXPERIMENT=jsonv2` per-module build requirements for source builders | Medium | S | Documentation |
| 19 | Upstream cqrs-lite otel `Provider.Shutdown` ForceFlush issue (verify-before-filing applies) | Medium | M | Bug |
| 20 | Fix dprint exit-14 on CHANGELOG-only commits (`--allow-no-files` in the BuildFlow hook) | Medium | S | Tooling |
| 21 | Sweep all module `.golangci.yml` files for stale `go:` pins (realtime still says 1.26.5) | Medium | S | Cleanup |
| 22 | Add a LICENSE-presence check to CI so no future module ships without one | Medium | S | Quality |
| 23 | Integration test: replay burst > subscriber buffer (64) — drop-and-heal via Last-Event-ID reconnect | Medium | M | Quality |
| 24 | Integration test: heartbeat visibility with a short `WithHeartbeat` through the appkit stack | Low | S | Quality |
| 25 | Integration test: flightrecorder middleware + realtime + appkit combined stack capture | Medium | M | Quality |
| 26 | Combined showcase example (`example/` refresh): core + realtime + errorpages + otel opt-in | Medium | M | Documentation |
| 27 | Record the now-verified JournalSSEStore recipe in `realtime/README.md` (not just AGENTS.md) | Medium | S | Documentation |
| 28 | Decide/expose a `realtime.Mount` passthrough for `JournalSSEStore`'s `WithMaxReplay` | Medium | S | Feature |
| 29 | otel middleware benchmark (no-op vs configured) + benchstat into the module README | Medium | M | Quality |
| 30 | Propose httputil `Logging` emit-with-request-context upstream (span correlation) | Medium | M | Feature |
| 31 | Satellite test sweep for `DrainDelay: 0` misuse (hidden 5s default tax) | Medium | S | Cleanup |
| 32 | Contributor onboarding: CONTRIBUTING.md for go-appkit (build env, jsonv2, per-module lint) | Medium | M | Documentation |
| 33 | go-structure-linter: configure acceptance of the root-package layout | Low | S | Tooling |
| 34 | pkg.go.dev badges per module in READMEs | Low | S | Documentation |
| 35 | Dependabot/renovate config for the repo's own modules | Medium | S | Tooling |
| 36 | govulncheck in CI; track the 1.26.6 toolchain bump when nixpkgs carries it | Medium | S | Quality |
| 37 | Extract the M18.3-style flush assertions into a reusable httpspec-style helper | Low | M | Quality |
| 38 | Integration test: `FilteredEventStore` path via `realtime.WithFilter` over `JournalSSEStore` | Medium | S | Quality |
| 39 | Cross-repo courtesy note to cqrs-htmx: the composition seams they rely on are now regression-locked in go-appkit | Low | S | Communication |
| 40 | Boundary doc: `health` module vs `flightrecorderhealth` (both bridge go-health) — merge trigger or scope statement | Medium | S | Documentation |
| 41 | Bench ticket: the ~30 extra allocs/op appkit-vs-httputil delta from the ADR-001 adoption bench | Low | L | Quality |
| 42 | ROADMAP: consumer registry (who consumes which module at which version) | Low | S | Planning |
| 43 | Add `check-modules`-style workspace script (isolation, replace-absence, go.work completeness incl. `./integration`) | Low | S | Tooling |
| 44 | Release checklist: flip gate docs (TODO_LIST/AGENTS) in the same session as any push | Low | S | Process |
| 45 | docs-mod: confirm catalog output covers the integration module's mounted routes in the showcase example | Low | S | Documentation |
| 46 | Add dependabot-style automation or a scheduled bump for the `cqrs-htmx` pin in `integration/go.mod` | Low | S | Tooling |
| 47 | Pre-commit hook self-test: assert dprint never sees an empty file set (regression for f-20) | Low | S | Tooling |
| 48 | Write the session's status-report follow-through: run `docs-health` HARVEST on this section f | High | S | Process |
| 49 | Consider `integration/` as a required pre-release gate once items 3-7 land (policy decision) | Medium | S | Planning |
| 50 | Re-run the full 10-module lint sweep sequentially (last verified set predates integration + health) | Medium | M | Quality |

## g) Questions I cannot answer myself

1. **License:** Will go-appkit adopt a standard open-source license (e.g. MIT, matching cqrs-htmx), or stay proprietary? I tried to infer intent from the LICENSE text, cqrs-htmx's MIT choice, and the pkg.go.dev damage — the decision itself is legal/business and only yours. It gates items f-1/f-2 and the whole public-adoption story. *(If proprietary is intentional, I'll reframe f-2 as "make the hidden-docs state deliberate and documented".)*
2. **Push policy:** master is currently ahead of origin (auto-commit daemon + docs commits accumulated today). Do you gate every push manually (and want the gate docs flipped by whoever pushes), or should routine doc/tooling commits ride regular pushes? The 19-day "push pending" rot suggests the gate ownership was ambiguous.
3. **Integration module ambition:** should `integration/` grow into the full cross-repo contract suite (f-3, f-5, f-6, f-7, f-23-25, f-38) and become a required pre-release gate, or stay a minimal 2-test seam guard? I picked minimal-first unilaterally; the effort multiplier is ~4x between the two answers.

---

*Point-in-time snapshot. Section (f) is the HARVEST input — route to `TODO_LIST.md`/`ROADMAP.md`, don't let it die here. Fast-moving parts: the license decision, the cqrs-htmx fold-in, the parallel session's cqrs WIP.*
