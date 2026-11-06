# Status Report — Server-Timing Adoption + Composition-Ruling Corrections

> **Live remainder:** the Server-Timing adoption session's 50-item ledger is
> fully adjudicated (this archive): the stack shipped (core v0.10.0 at
> 04f0c20), the R3/R4a/R5a verification batches closed their rows, the
> delegation + registry + verdict-record asks all landed, and the gated
> battery/upstream/owner legs carry routed markers into their wired TODO
> rows. Archived 2026-11-06: every §f item carries a verdict; nothing here
> is unowned.
>

> **2026-11-03 20:35 CET · go-appkit @ master · docs/status/2026-11-03_20-35_server-timing-adoption-session-status.md**
> Covers: the fleet audit (earlier today) → the owner's composition-mission ruling → the
> same-session server_timing SHIP + record corrections + the drift sweep just completed.
> Nothing else researched.

## Scope line — verified vs NOT verified

**Verified (fresh commands, this session):**

- `GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1` — root + testkit **ok** (twice:
  after the middleware change and again after the drift fixes); otel module ok after its README edit.
- `go build` + `go vet` hermetic — OK; `golangci-lint run ./...` from repo root — **0 issues**
  (after deduping my first test draft into the existing test); `gofmt -l` — clean.
- `go mod tidy` hermetic — delta exactly 1 line (the direct require); `server_timing` v1.0.3
  already in every go.sum via httputil's chain.
- Targeted tests: `TestMiddleware_DefaultStackHeaders_Present` + `TestMiddleware_ReplacedStack` —
  PASS (header present on default stack; absent under `Middlewares` replacement).
- Drift sweep for the changed contract string (`rg "Recovery → RequestID"` repo-wide) — found 3
  stale living-doc/code mentions; all 3 fixed and re-verified; 2 integration test comments left
  intentionally (they describe the PUBLISHED core v0.9.0 stack those tag-pinned tests actually
  exercise); battery spec + status docs left as point-in-time records.
- Guards: `check-routed-destinations` exit 0, `check-go-directives` exit 0 (run after drift fixes).
- index.html: tag-balance OK; errata callout + corrected matrix rows screenshot-verified in
  headless Firefox (2 crops).
- AGENTS.md: 136/220 lines after ruling bullet + deps-table row.

**NOT verified (and why):**

- **The bench has no real baseline.** I ran `-benchtime 1x` as a smoke (667µs/op is amortized
  server startup, not per-op truth). A proper `-benchtime` run was never done and no number is
  recorded anywhere (AGENTS' perf section has Logging numbers; ServerTiming has none).
- **Header CONTENT is unasserted.** Tests assert presence of `Server-Timing`, not that it carries
  the auto-recorded `total` metric. The contract rests on the upstream library's behavior.
- **"SSE-safe" rests on upstream docs.** No local pin: no test streams a response and asserts the
  header was injected at first write on an SSE/no-timeout service.
- **`scripts/example-smoke.sh` not run** — examples boot/probe through the new stack only via the
  unit suite; the 7-example smoke is a CI leg.
- **Consumer wire-change readiness**: the header rides the NEXT core release; consumer-claims
  guard (`check-consumer-claims.sh`) is a pre-tag ritual and was not run (nothing tagged).
- **dprint/format check** over the touched md files: never run (unknown whether dprint covers them).
- **go.work.sum gained a hash** (`cel.dev/expr v0.25.3`) from ONE targeted test run where I forgot
  the `GOWORK=off` prefix (workspace-mode aggregate). Benign, but it's an unrequested mutation I
  only noticed afterward.

## Brutal self-review (the three questions)

**What did I forget?**

1. **The drift sweep.** I changed a grep-able contract (the middleware stack order) and updated
   middleware.go + README + CHANGELOG — but NOT `config.go:61`'s doc comment,
   `docs/DOMAIN_LANGUAGE.md`'s default-stack row, or `otel/README.md`'s recipe paragraph. The AGENTS
   gotcha list literally documents that these names are contracts, and I still shipped three
   orphaned mentions. Caught during the sweep for THIS report, fixed and verified within the hour —
   but "declare done" came before "grep the contract string repo-wide." That ordering was wrong.
2. **Carried-over defects from the first status report remain unfixed**: the INVALID rg command in
   `httputil.md`'s verify block (flagged at 19:54, still there), the mobile-width screenshot,
   the `realtime/hub.go:80` unexported-const doc comment, the dprint check. Flagging a defect and
   not fixing it in the next round is the failure mode the first report itself warned about.
3. **A real bench baseline.** Recorded a benchmark whose only run is statistically meaningless.
4. **The GOWORK prefix discipline** on a "quick" run — mutated go.work.sum.

**What could I have done better?**

1. Contract-change checklist: grep the old string repo-wide BEFORE declaring done, not while
   writing the retro. (The 2026-10-11 audit found the stack via `middleware.go:14-31` — the same
   class of reference now lives in config/DOMAIN_LANGUAGE/otel-README, and I touched none of them
   until the sweep.)
2. Assert the header's `total` metric, not just presence — one `strings.Contains` away.
3. Run the bench properly (`-benchtime 1s` × count) and record the number next to the Logging
   baseline in AGENTS, so the wire-change cost is a fact, not a claim.
4. Fix flagged defects in the same session that flags them, or explicitly ticket them with an owner.

**What could I still improve?**

1. The Server-Timing story could get a composition E2E in `integration/` (stream + header) at the
   next integration re-pin train, when core's new stack is actually pinnable.
2. The full a-f harvest from BOTH status reports still needs one docs-health pass (TODO_LIST got
   the ruling but not the 50-item lists).
3. Report verify-blocks should be executed at write time — the invalid rg command is proof this
   keeps not happening without a rule.

## a) FULLY DONE

| # | Item                                                                                                                                                                                                                                                           | Evidence                                                                                                                            |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **server_timing ADOPTED into core**: direct require `httputil/server_timing v1.0.3`; `servertiming.ServerTimingMiddleware()` OUTERMOST in the default stack (`ServerTiming → Recovery → RequestID → Logging → [Timeout] → SecurityHeaders`)                    | `go.mod:9`; `middleware.go:11-31`; hermetic `go mod tidy` delta = 1 line                                                            |
| 2 | Test pins: default-stack header present (renamed `TestMiddleware_DefaultStackHeaders_Present`, asserts X-Request-ID + Server-Timing on one request); replaced-stack absence assertion; dedup lint fixed                                                        | `-run` verbose PASS; lint 0 issues                                                                                                  |
| 3 | Bench added: `BenchmarkServerTimingMiddleware` (servertiming_bench_test.go, same harness as Logging benches)                                                                                                                                                   | compiles, runs (`-benchtime 1x` smoke)                                                                                              |
| 4 | Truth-sync: core `CHANGELOG.md` [Unreleased] Added entry; README features + default-stack comment lines; middleware.go docstring contract; `config.go` Middlewares doc comment; `docs/DOMAIN_LANGUAGE.md` default-stack row; `otel/README.md` recipe paragraph | git diff; grep now shows zero stale `Recovery → RequestID` mentions outside point-in-time records + tag-pinned integration comments |
| 5 | Composition-ruling records: dated errata blocks in `docs/review/libs/server_timing.md` + `go-codec.md`; index.html errata callout + 2 corrected matrix rows (screenshot-verified); `docs/research/README.md` row note; status-report addendum (19:54 file)     | python replace asserts passed; HTML balance OK; Firefox crops                                                                       |
| 6 | AGENTS.md: ruling bullet (mission, consequence #1 shipped, #2 promoted, NOT-flipped list) + deps-table row for server_timing                                                                                                                                   | 136/220 lines; git diff                                                                                                             |
| 7 | TODO_LIST: battery W3-W5 row carries the ruling inline; B5 (go-codec) promoted to in-scope                                                                                                                                                                     | git diff line 50                                                                                                                    |
| 8 | Gates: full `-race` suite green ×2 (root+testkit, plus otel), vet/build green, lint 0 issues, gofmt clean, routed-destinations + go-directives guards green after every doc touch                                                                              | command outputs logged this session                                                                                                 |

## b) PARTIALLY DONE

| # | Item                           | Works                                                              | Open                                                                                                                                                        | Effort              |
| - | ------------------------------ | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| 1 | Bench for the new middleware   | Exists, runs                                                       | No real baseline (`-benchtime 1x` only); number not recorded in AGENTS perf section                                                                         | S                   |
| 2 | Header contract pin            | Presence asserted                                                  | `total` metric format not asserted; SSE-streaming injection not pinned locally (upstream-doc claim)                                                         | S                   |
| 3 | Composition-ruling propagation | Reports + index + AGENTS + TODO battery row + status addendum done | go-etag/go-idempotency/go-retry errata NOT added (deliberate — their framing already matches the mission); harvest of both reports' (f) lists still pending | S-M                 |
| 4 | library-deep-dive done-bar     | Gate=full suite green this time; Memory=AGENTS                     | Browser=desktop only (mobile never shot); Index rows exist but the canonical-home question is unanswered                                                    | S                   |
| 5 | Wire-change release readiness  | Code+tests+CHANGELOG ready on master                               | Uncommitted; consumer-claims guard + fresh-consumer proxy check are pre-tag rituals; header reaches consumers only at the next core tag                     | M (a release train) |

## c) NOT STARTED

| #  | Item                                                                                              | Why                                                                  |
| -- | ------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| 1  | W3-httpx `B5 content negotiation on go-codec` (the promoted go-codec seam — new module)           | Battery-sized work; ruling promotes it, design not started           |
| 2  | otel Recovery → `httputil.RecoveryWithConfig`+`OnPanic` refactor (the audit's other code finding) | Awaiting owner go (asked twice, unanswered)                          |
| 3  | Fix the INVALID rg command in `httputil.md` verify block (+ execute all 12 blocks)                | Flagged at 19:54, still unfixed — embarrassment #1                   |
| 4  | Version-drift guard (assert AGENTS deps-table versions == go.mod)                                 | Idea from round 1; the table was edited AGAIN this round by hand     |
| 5  | Mobile-width screenshot of index.html                                                             | Flagged at 19:54, still not done                                     |
| 6  | `realtime/hub.go:80` doc comment (unexported go-sse const reference)                              | Flagged at 19:54, still not done                                     |
| 7  | dprint/format check over new/changed md+html                                                      | Never attempted                                                      |
| 8  | Canonical home decision for per-lib deep-dives (docs/review/libs vs docs/research)                | Owner question, unanswered                                           |
| 9  | Composition E2E for Server-Timing (stream + header) in integration                                | Needs next core tag to pin against (integration pins PUBLISHED tags) |
| 10 | HARVEST of both status reports into TODO_LIST/ROADMAP                                             | Awaiting owner routing preference                                    |

## d) TOTALLY FUCKED UP

No data loss, no broken gates, no failing tests — the tree is green end to end. Honest near-misses:

1. **Three-file contract drift shipped green** (config.go comment, DOMAIN_LANGUAGE.md, otel/README.md
   described the old stack order after my change). Severity if unfixed: medium (every future reader
   of the contract gets a lying name — the exact split-brain the repo's own gotchas warn about).
   Root cause: declared done before grepping the changed contract string repo-wide. Fixed + verified
   within the hour of discovery; the catch happened during THIS report's prep, not at ship time.
2. **A statistically meaningless bench number** was produced and would have been misleading if
   quoted (667µs/op = startup amortization at `-benchtime 1x`). Mitigation: labeled as smoke here
   and in session output; real baseline is f-item #2.
3. **Recurring class**: defects flagged in a status report and NOT fixed by the next round
   (httputil.md invalid rg command, mobile screenshot, hub.go:80 comment). Nothing enforces
   "flagged → fixed or ticketed" — it currently relies on my memory, which is exactly what the
   memory rules say not to trust.
4. **Unrequested tree mutation**: go.work.sum hash added by one forgotten `GOWORK=off` prefix.
   Benign (workspace aggregate, tracked file), but "quick run" exceptions are how the leak class
   in AGENTS happened eight times historically.

## e) WHAT WE SHOULD IMPROVE

1. **Contract-change rule**: when a documented/grep-able contract string changes (stack order,
   log phase names, sentinel names), `rg` the OLD string repo-wide BEFORE declaring done; living
   docs get fixed, point-in-time records get dated errata, tag-pinned test comments stay (they
   describe the pinned tag, not master).
2. **Executable report appendices**: run every "How to verify" block at write time, or label it
   "untested transcript." The invalid rg command is now two reports old.
3. **Flagged→fixed-or-ticketed rule**: anything named in a status report's (d)/(b) must either be
   fixed next round or land as a TODO_LIST row in the same session. No third option.
4. **Record bench baselines**: one proper `-benchtime` run per new benchmark, number into AGENTS'
   perf section next to the Logging numbers (the wire-change cost should be a published fact).
5. **Assert contracts, not presence**: header tests should check `total` metric format; the SSE
   injection claim deserves a local pin when integration can pin it.
6. **GOWORK prefix discipline on EVERY go command**, no exceptions for "quick" runs (go.work.sum
   mutation is the visible symptom).
7. **Keep the good patterns from this round**: copy-template + exact-marker python replaces for
   HTML; annotate-never-rewrite errata blocks for point-in-time reports; drift sweep as part of
   the retro (it worked — make it part of ship, not part of review).

## f) Up to 50 things to get done next

(I: impact, E: effort S<30m / M≤2h / L>2h, C: category. This is the HARVEST feed — belongs in
TODO_LIST/ROADMAP; routed items from round 1 that are DONE are dropped, still-open ones carried.)

**Immediate repair (this session's own loose ends)**

1. ~~Fix the invalid rg command in `httputil.md`'s verify block; execute all 12 report verify-blocks as written, fix failures. I:High E:S C:Quality~~ done — the verify-block rule is in the AGENTS gotcha and the batches ran as written this week
2. ~~Record a real ServerTiming bench baseline (`-benchtime 1s -count 6`, benchstat) into AGENTS' perf section next to the Logging numbers. I:Medium E:S C:Quality~~ done — the AGENTS perf section carries the quiet-run bench numbers (the 13-10 re-baseline)
3. ~~Assert the `Server-Timing` header carries `total` (strings.Contains) in the default-stack test. I:Medium E:S C:Quality~~ done — the server-level total assertion is pinned (middleware_test.go:80+172)
4. ~~Fix `realtime/hub.go:80` doc comment (unexported `sse.defaultSubscriberBuffer`). I:Low E:S C:Documentation~~ done — the hub.go doc-const note landed (13-10, R1d)
5. ~~Mobile-width screenshot of index.html (complete the done-bar Browser item). I:Low E:S C:Quality~~ done — the mobile screenshot landed (13-10)
6. ~~dprint/format check over every new/changed md+html file; fix or record "dprint doesn't cover them". I:Low E:S C:Quality~~ done — the dprint pass ran (R1d) and scripts/docs-annotate-pass.sh now guards the canon
7. ~~Run `scripts/example-smoke.sh` locally once against the new stack (7 examples boot/probe with the header present). I:Medium E:S C:Quality~~ done — example-smoke 8/8 green under quiet-run (p9a)
8. ~~`go list -m all | grep go-retry` in integration — prove or fix the MVS floor claim from round 1. I:Low E:S C:Documentation~~ done — the go-retry MVS proof green (R1e)

**Release train (the wire change reaches consumers here)**
9. ~~Decide + execute the core release train for the Server-Timing stack (API-break check, CHANGELOG date, pre-tag guards incl. consumer-claims, tag, fresh-consumer proxy check). I:High E:M C:Release~~ done at `04f0c20`
10. ~~At the next integration re-pin: refresh the two tag-pinned test comments that still describe the published v0.9.0 stack; add the Server-Timing composition E2E (stream + header). I:Medium E:M C:Quality~~ done — the re-pins landed (the 10-12 train + the preset pin); tag-pinned comments stay per the AGENTS rule
11. ~~Update AGENTS Release State + FEATURES.md in the same train (release ritual step 3). I:Medium E:S C:Documentation~~ done — the same-train ritual executed (AGENTS Release State + FEATURES synced with core v0.10.0)

**Composition ruling follow-through**
12. Build W3-httpx B5 content negotiation on go-codec (the promoted seam; go-daemon `negotiation.go` as reference). I:High E:L C:Feature **→ TODO P3 — Battery program waves W3-W5 (B5 negotiation on go-codec) <!-- fresh:skip: routed in TODO_LIST -->**
13. ~~Refactor `appkitotel.Recovery` onto `httputil.RecoveryWithConfig`+`OnPanic` (+ extend exceptions tests). I:High E:M C:Quality~~ done — the Recovery delegation landed (RecoveryWithConfig + OnPanic; the hand-rolled body deleted; AGENTS records it)
14. ~~Battery-row annotations: go-etag five-module note on B4/W4 rows; `polite` = go-retry + etagclient evaluation. I:Medium E:S C:Documentation~~ done — the battery row carries the go-etag/polite annotations (the fleet-audit errata + row text)
15. ~~Error-family registry DECISION row (adopt-now vs demand-gated) so F2-class findings stop rotting. I:Medium E:S C:Documentation~~ done — the fleet registry SHIPPED (RegisterClassifications: stdlib defaults + MaxBytesError→Rejection; the questionnaire answer)
16. ~~HARVEST both status reports' (f) lists into TODO_LIST/ROADMAP via docs-health. I:High E:S C:Documentation~~ done — the harvest is this week's archive waves
17. ~~Canonical-home decision for per-lib deep-dives + write the rule into the docs map and both indexes. I:Medium E:S C:Documentation~~ done — docs/review/libs/ is canonical (the questionnaire answer; the rule written into AGENTS + the indexes)

**Framework gaps named by the audit (demand-gated unless the ruling flips them)**
18. ~~CORS battery row (browser-API gap; OriginCheck ≠ CORS). I:Medium E:S C:Documentation~~ done — the CORS verdict recorded (R5a; row closed)
19. ~~Trusted-proxy keying note (`KeyExtractorFromTrustedClientIP`) for proxy deployments. I:Medium E:S C:Documentation~~ done — the trusted-proxy verdict recorded (R5a; row closed)
20. ~~Compression/decompression: explicit adopt-or-decline verdict row (httputil families with no story). I:Low E:S C:Documentation~~ done — the compression verdict recorded (R5a; row closed)
21. ~~`LoggingFromContext` adoption-or-N/A verdict (otel README documents the recipe; core still ships plain Logging). I:Low E:S C:Documentation~~ done — the LoggingFromContext verdict recorded (R5a; row closed)
22. ~~error-family fleet-classification bootstrap prototype (RegisterStdlibDefaults + MaxBytesError + context.Canceled). I:Medium E:M C:Feature~~ done — the registry bootstrap SHIPPED (RegisterClassifications)
23. ~~`WrapOnce` at idempotent Shutdown re-wrap sites. I:Low E:S C:Quality~~ done — WrapOnce landed (R3)
24. ~~`errorfamilytest` asserts in otel/systemd tests. I:Low E:S C:Quality~~ done — errorfamilytest asserts landed (R3)
25. ~~`errorfamily.WithContext` at drain-hook error sites. I:Low E:S C:Quality~~ done — the hook_index context landed (R3)

**Upstream/release watches**
26. go-sse next release: absorb breaking `ssetest.WithLastEventID` via re-pin ritual; update `hub.go:80` reference to the exported `DefaultSubscriberBuffer`. I:Medium E:M C:Quality **→ TODO P3 — Upstream watchlist (go-sse absorption + hub.go reference update) <!-- fresh:skip: routed in TODO_LIST -->**
27. go-etag next release: root-shim v1.0.0 removal plan vs integration exclude block. I:Low E:S C:Quality **→ TODO P3 — Upstream watchlist (go-etag root-shim removal plan vs excludes) <!-- fresh:skip: routed in TODO_LIST -->**
28. go-idempotency next release: CHANGELOG hygiene check before any promotion to direct. I:Low E:S C:Quality **→ TODO P3 — Upstream watchlist (go-idempotency CHANGELOG hygiene) <!-- fresh:skip: routed in TODO_LIST -->**
29. httputil next train: server_timing pairing + apidiff frozen-API gate. I:Low E:S C:Quality **→ TODO P3 — Upstream watchlist (httputil pairing + apidiff gate) <!-- fresh:skip: routed in TODO_LIST -->**

**Guard/tooling ideas born this session**
30. ~~Version-drift guard: assert AGENTS deps-table versions == go.mod versions (the table drifted once and was hand-fixed twice). I:High E:M C:Quality~~ done at `dfeb4a5`
31. ~~Add "grep the old contract string repo-wide" to the ship checklist in AGENTS gotchas (this round's lesson #1). I:Medium E:S C:Documentation~~ done — the grep-old-contract gotcha is in AGENTS
32. scripts helper: comment-stripped symbol-scan pipeline for future library audits. I:Low E:S C:Quality **→ TODO P3 — SUPERB v12 micro-backlog (comment-stripped symbol-scan helper for audits) <!-- fresh:skip: routed in TODO_LIST -->**
33. docs-health ANNOTATE pass: the 19:54 status report's stale claims (e.g. "9 non-uses" scorecard card in index.html predates the ruling — the errata callout covers it, verify no OTHER stale card reads wrong). I:Low E:S C:Documentation **→ TODO P3 — docs-health corpus waves 3-4 (the 19-54 annotate pass handles the stale claims) <!-- fresh:skip: routed in TODO_LIST -->**

**Bigger rocks (unchanged, carried)**
34. Core TLS option via the httputil v1.7.0 `NewServerListener` seam (spike re-run + swap). I:High E:L C:Feature **→ TODO P2 — compose Service on httputil.Server (TLS seam via NewServerListener) <!-- fresh:skip: routed in TODO_LIST -->**
35. W3-httpx B1/B2/B9 handler-DX family (error mapping parity with errorpages taxonomy). I:High E:L C:Feature **→ TODO P3 — Battery program waves W3-W5 (B1/B2/B9 handler-DX family) <!-- fresh:skip: routed in TODO_LIST -->**
36. W5-C1 realtime drop counters via go-sse `WithOnDrop` (ssehub `Drops()` reference). I:Medium E:M C:Feature **→ TODO P3 — Battery program waves W3-W5 (W5-C1 drop counters) <!-- fresh:skip: routed in TODO_LIST -->**
37. C2 projection→broadcast folded contract (the must-have W5 item). I:High E:L C:Feature **→ TODO P3 — Battery program waves W3-W5 (C2 folded contract) <!-- fresh:skip: routed in TODO_LIST -->**
38. D7 idempotency adopt-wire decision when demand lands (README recipe first). I:Low E:S C:Documentation (gated) **→ TODO P3 — Battery program waves W3-W5 (D7 idempotency adopt-wire) <!-- fresh:skip: routed in TODO_LIST -->**
39. D4 atomic file write on go-atomic-write ≥ v0.6.0 when config/state batteries fire. I:Low E:S C:Feature (gated) **→ TODO P3 — Battery program waves W3-W5 (D4 atomic file write) <!-- fresh:skip: routed in TODO_LIST -->**
40. `polite` outbound client on go-retry (W4). I:Medium E:L C:Feature (gated) **→ TODO P3 — Battery program waves W3-W5 (W4 polite on go-retry) <!-- fresh:skip: routed in TODO_LIST -->**
41. W5-C3 per-subscriber auth+filter. I:Medium E:M C:Feature (gated) **→ TODO P3 — Battery program waves W3-W5 (W5-C3 per-subscriber auth+filter) <!-- fresh:skip: routed in TODO_LIST -->**
42. samber/do auditlog hooks + lazy-healthcheck upstream ask (owner-gated, ruled 2026-11-02). I:Low E:S C:Documentation (gated) **→ TODO P3 — samber/do items (auditlog + lazy-healthcheck; owner-gated) <!-- fresh:skip: routed in TODO_LIST -->**
43. cqrs EventConfig encryption/signing opt-ins (consumer-gated). I:Low E:L C:Feature (gated) **→ TODO P3 — Battery program waves W3-W5 (EventConfig encryption/signing opt-ins) <!-- fresh:skip: routed in TODO_LIST -->**
44. cordis bridge (trigger 1/3 met; needs consumer statement + core v1.0.0 criteria). I:Low E:L C:Feature (gated) **→ TODO P3 — cordis bridge watchlist (trigger 1/3 met) <!-- fresh:skip: routed in TODO_LIST -->**
45. go-aichat reverse-adoption E2E (trigger-gated on appkit-hosted release of theirs). I:Low E:M C:Quality (gated) **→ TODO P3 — cordis bridge watchlist (go-aichat reverse-adoption E2E) <!-- fresh:skip: routed in TODO_LIST -->**
46. project-dependency-graph tool polish list (their repo). I:Low E:M C:Cleanup (gated) **→ TODO P3 — Upstream watchlist (pdg polish list; their repo) <!-- fresh:skip: routed in TODO_LIST -->**
47. systemd module backlog (Examples, WithStatus, Install-returns-Hooks; ride next systemd train). I:Low E:M C:Quality **→ TODO P3 — systemd module small backlog <!-- fresh:skip: routed in TODO_LIST -->**
48. pbx-lens testkit migration pointer row (their repo, when they adopt core v0.9.0+). I:Low E:S C:Documentation (gated) **→ TODO P3 — Straggler routing row (pbx-lens testkit migration) <!-- fresh:skip: routed in TODO_LIST -->**
49. ~~Decide commit strategy for today's tree (13+ modified/new files; daemon may auto-commit piecemeal). I:Medium E:S C:Cleanup~~ done — the trains committed in logical groups; the daemon race is documented in the AGENTS gotchas
50. ~~AGENTS perf section: also record the default-stack-with-ServerTiming end-to-end delta vs the old stack (one bench pair), so the wire cost story is complete. I:Low E:S C:Quality~~ done — the AGENTS perf section carries the old + new stack numbers (the end-to-end delta data)

## g) Three questions I cannot figure out myself

1. **Is default-ON the intended long-term posture for the Server-Timing header?** I chose it
   because the composition mission says batteries come included, and removal exists via
   `Middlewares` replacement. But it IS a wire change every consumer sees (an extra response
   header, incl. PapDashboard at their next core bump). If you want zero-surprise wire, I flip it
   to an opt-in default-OFF seam — one-line change + test rename.
2. **Do you want the core release train NOW** (tag the Server-Timing stack so cqrs-htmx/other
   consumers can pin it), **or accumulate** more default-stack composition work (otel Recovery
   refactor, B5) into one train? This decides whether I run the pre-tag ritual + consumer-claims
   - fresh-consumer proxy check today or keep the tree on master.
3. **For B5 (go-codec content negotiation): build it as the new `httpx` module per the battery
   spec (B1..B9 family), or as a smaller standalone negotiation seam first?** The battery spec's
   httpx home needs the `svc.Routes()` seam decision for B6/B8 anyway — I can't tell whether you
   want the full module program started or one battery shipped as a proof.

— Report written 2026-11-03 20:35 CET. Awaiting instructions.
