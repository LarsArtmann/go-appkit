# Status Report — Decision Batch Executed (Knob, OTel Delegation, Fleet Registry)

> **2026-11-03 21:25 CET · go-appkit @ master · docs/status/2026-11-03_21-25_decision-batch-execution-status.md**
> Covers: the 13 answered Questions-tool decisions and their first execution wave
> (ServerTiming knob + phase-pattern docs, otel Recovery delegation, fleet error
> classifications), plus what of the 13 answers is still open. Nothing else researched.

## Scope line — verified vs NOT verified

**Verified (fresh commands, minutes ago):**

- Root `go test ./... -race -count=1` — **ok** (appkit + testkit), with the three NEW tests
  passing: `TestServiceConfig_ServerTiming_Knob` (nil/&true/&false knob states),
  `TestRegisterClassifications_MaxBytesErrorIsRejection`, `TestRegisterClassifications_StdlibDefaults`.
- otel module `-race` — **ok** after the Recovery delegation refactor (existing exception/panic
  pins hold — the delegation preserved the contract).
- `golangci-lint run ./...` root — **0 issues** (after fixing one `mnd` finding my preallocated
  slice introduced); otel lint clean.
- `go doc`-verified against the real APIs before writing: `httputil.RecoveryConfig{Logger, OnPanic}`
  (hook runs after log, before 500; abort panics re-panic BEFORE the hook — so otel's
  `isAbortPanic` became dead code and was deleted), `errorfamily.RegisterClassificationType[T](family)`,
  `RegisterStdlibDefaults(reg)`, `net/http.MaxBytesError{Limit}` (NO `Actual` field on this Go —
  vet caught my guessed struct literal).
- Actual stdlib classification read from source: `context.Canceled → Rejection`
  ("caller abandoned; retrying without a new request is wrong"), `DeadlineExceeded → Transient`.

**NOT verified / NOT done:**

- **CHANGELOG not yet updated** for the knob, the otel delegation, or the registry — the three
  new code changes are documented only in code/doc-comments so far.
- **README config table** has no `ServerTiming` knob row yet.
- **AGENTS/TODO_LIST/ROADMAP decision records** (13 answers) not yet written down.
- **No commits** — despite `commit_now` being answered; the landing step hasn't run.
- **Consumer heads-up drafts** (`both`) not drafted yet — needs the github-voice skill loaded.
- **HARVEST** (`harvest_now`) not executed yet — needs the docs-health skill loaded.
- **docs/review/libs canonical rule** not yet written into the indexes.
- **example-smoke / fresh-consumer proxy** untouched — train parked per `accumulate`.

## The 13 decisions — record of answers

| #  | Question               | Answer                                                                                       |
| -- | ---------------------- | -------------------------------------------------------------------------------------------- |
| 1  | Server-Timing posture  | **Config knob, default-ON** (`ServerTiming *bool`, RegisterHealth semantics)                 |
| 2  | Release train          | **Accumulate** — hold master, one bigger train later                                         |
| 3  | B5 shape               | **Full httpx module program** (B1-B9 family)                                                 |
| 4  | Ruling scope           | **Blanket promote all W3-W5** batteries to planned work                                      |
| 5  | Docs home              | **docs/review/libs/ canonical**                                                              |
| 6  | otel Recovery refactor | **Into this train** (done now — see a/2)                                                     |
| 7  | Commit strategy        | **I commit now**, logical groups                                                             |
| 8  | Harvest                | **Now, full**                                                                                |
| 9  | Error-family registry  | **Build now** (done now — see a/3)                                                           |
| 10 | Consumer heads-up      | **Draft for both** (cqrs-htmx + PapDashboard)                                                |
| 11 | Next workstream        | **Repair backlog first**                                                                     |
| 12 | svc.Routes() seam      | **Deferred — owner asked "what is httpx vs httputil?"** (answered in chat; decision pending) |
| 13 | ST metric coverage     | **Phases now** — partially unservable, see (d)/#1                                            |

## a) FULLY DONE

| # | Item                                                                                                                                                                                                                                                                                                                            | Evidence                                                                                                 |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| 1 | **ServerTiming config knob**: `ServiceConfig.ServerTiming *bool` — nil/&true ON, &false omits the middleware; DefaultServiceConfig sets &t; knob + stamp-on-commit phase contract documented in the field comment and the stack docstring                                                                                       | `config.go`, `middleware.go`; knob test asserts all three states through `buildMiddleware`+`Chain`       |
| 2 | **otel Recovery delegation**: `appkitotel.Recovery` is now `httputil.RecoveryWithConfig{Logger, OnPanic: recordPanic}` — the hand-replicated log/500/re-panic body is gone; `isAbortPanic` deleted (library re-panics aborts before the hook); exception capture + error status unchanged and pinned by the existing otel suite | otel `-race` ok, lint clean; `otel/exceptions.go` shrank to hook + recordPanic + RecordError             |
| 3 | **Fleet error classifications**: `appkit.RegisterClassifications()` — stdlib defaults (context deadline→Transient, canceled→Rejection, sql/os per the family table) + `*http.MaxBytesError`→Rejection into the default registry; idempotent; deliberately NOT auto-called (global-state hygiene documented)                     | `errors.go` + new `errors_test.go` (Classify→Rejection, HTTPStatus→400, idempotent re-call, stdlib pins) |
| 4 | Tree green end-to-end: root `-race` ok ×2 files, otel `-race` ok, root lint 0 issues, otel lint clean, gofmt clean                                                                                                                                                                                                              | command outputs this session                                                                             |

## b) PARTIALLY DONE

| # | Item                      | Works                                                                                                                                                 | Open                                                                                                                                                                                                                                                                                                                                                                                                 | Effort                    |
| - | ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- |
| 1 | "Phases now" (answer #13) | Knob + collector-in-every-context + the handler-phase pattern (`MeasureServerTiming` before first write) documented in the stack docstring and README | **The literal ask — a framework-measured app-handler phase — is unservable on net/http**: the header is stamped at response commit, and the mux duration always completes after the commit, so it can never appear in the header. The honest deliverable is knob + documented handler-side phases; a framework phase would require delaying commits (breaks SSE). Needs your ack of this constraint. | — (design fact, not work) |
| 2 | Decision records          | This report + code comments                                                                                                                           | AGENTS/TODO_LIST/ROADMAP write-down of all 13 answers + docs-home rule NOT yet done                                                                                                                                                                                                                                                                                                                  | S                         |
| 3 | CHANGELOG                 | Original Server-Timing entry exists                                                                                                                   | Knob, otel delegation, registry entries missing                                                                                                                                                                                                                                                                                                                                                      | S                         |
| 4 | README                    | Features + stack lines updated earlier                                                                                                                | Config-table row for the ServerTiming knob missing                                                                                                                                                                                                                                                                                                                                                   | S                         |
| 5 | Landing                   | All work verified on master                                                                                                                           | `commit_now` not executed (waiting for CHANGELOG/records to complete first so commits are truthful)                                                                                                                                                                                                                                                                                                  | S-M                       |

## c) NOT STARTED (queued by your answers)

| # | Item                                                                                                                                                                                                               | Gate                                                                               |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------- |
| 1 | Consumer heads-up drafts (cqrs-htmx + PapDashboard, your voice)                                                                                                                                                    | github-voice skill; next step                                                      |
| 2 | Full HARVEST of both status reports into TODO_LIST/ROADMAP                                                                                                                                                         | docs-health skill; next step                                                       |
| 3 | Repair backlog (your chosen next workstream): invalid rg verify-command in httputil.md, mobile screenshot, real ServerTiming bench baseline, hub.go:80 comment, dprint check, example-smoke, integration MVS proof | After landing                                                                      |
| 4 | httpx module program (B1-B9, full family per answer #3)                                                                                                                                                            | Big; after repairs; svc.Routes() answer pending your read of the httpx explanation |
| 5 | Blanket-promotion TODO/ROADMAP rewrite of W3-W5 (answer #4)                                                                                                                                                        | Part of the records write-down                                                     |
| 6 | Version-drift guard (AGENTS table vs go.mod)                                                                                                                                                                       | After repairs                                                                      |

## d) TOTALLY FUCKED UP

Nothing at severity — tree is green. But two honest same-class stumbles this round:

1. **I wrote tests against remembered APIs twice, and both guesses were wrong**: (a) asserted
   `context.Canceled → Infrastructure` without reading the family's rationale table — the suite
   went RED (actual: **Rejection**, "caller abandoned; retrying without a new request is wrong");
   (b) constructed `http.MaxBytesError{Limit, Actual}` — `Actual` does not exist on this Go
   (vet caught it; field set is `{Limit int64}` alone). Root cause identical: `go doc` first,
   write second — I did the opposite. Both fixed and now green, but the round's own lesson
   ("verify before writing") was violated twice in ten minutes.
2. **Answer #13 ("phases now") cannot be delivered as literally offered** — my option text
   promised "app-handler duration measured by the framework", and the library's stamp-on-commit
   contract makes that unservable for any response that writes (which is all of them, eventually).
   I shipped the servable part (knob + handler-side phase pattern, documented) instead of a fake
   metric. You believed you were buying a framework phase; you got a config knob + documentation.
   That gap is on me for offering an unservable option.

Carried from earlier rounds (still true): go.work.sum hash mutation from one unprefixed run;
the invalid rg verify-command in httputil.md; mobile screenshot; dprint check — all open in (c).

## e) WHAT WE SHOULD IMPROVE

1. **API-first writing**: before ANY new API usage in code or tests, `go doc` the symbol in THIS
   repo's pinned version. Two wrong-guess compile/test failures in one round is the pattern.
2. **Don't offer unservable options in questionnaires**: verify feasibility (here: the
   stamp-on-commit rule was in the library README I had already inventoried) BEFORE putting an
   option on a decision list. If it turns out unservable, lead the report with that fact.
3. **Answer-records-first landing**: write the 13 answers into AGENTS/TODO/ROADMAP BEFORE coding
   spills further, so every later commit carries the decision context (currently only this report
   - code hold them).
4. **Commit in logical groups as units complete** rather than accumulating 17+ files across
   multiple features (the tree now spans 4 separable changes).
5. Keep what worked: `go doc` verification of hook ordering before the refactor (the
   abort-repanic-before-hook discovery made `isAbortPanic` deletion safe and obvious).

## f) Up to 50 things to get done next

(I: impact, E: effort S<30m / M≤2h / L>2h, C: category. Items already executed this round are
dropped; the surviving list, re-ranked by your answers.)

**Landing the current batch (before anything else)**

1. CHANGELOG: knob, otel delegation, registry entries ([Unreleased]). I:High E:S C:Documentation
2. README: ServerTiming knob row in the config table. I:Medium E:S C:Documentation
3. Write the 13-answer decision record into AGENTS.md + TODO_LIST + ROADMAP (incl. docs/review/libs canonical rule into both indexes). I:High E:S C:Documentation
4. Commit the tree in logical groups: (1) ST adoption+knob, (2) otel delegation, (3) fleet registry, (4) fleet-audit reports+errata+indexes, (5) status reports + decision records. I:High E:M C:Cleanup
5. Consumer heads-up drafts (cqrs-htmx + PapDashboard) in your voice → docs/feedback/outgoing/. I:Medium E:S C:Documentation
6. Full HARVEST: both status reports' (f) lists → TODO_LIST/ROADMAP, deduped against existing battery rows + the blanket W3-W5 promotion. I:High E:S-M C:Documentation

**Repair backlog (your chosen next workstream)**
7. Fix the invalid rg command in httputil.md's verify block; execute all 12 report verify-blocks. I:High E:S C:Quality
8. Real ServerTiming bench baseline (`-benchtime 1s -count 6` + benchstat) into AGENTS perf section. I:Medium E:S C:Quality
9. Mobile-width screenshot of index.html. I:Low E:S C:Quality
10. Fix `realtime/hub.go:80` unexported-const doc comment. I:Low E:S C:Documentation
11. dprint/format check over new/changed md+html. I:Low E:S C:Quality
12. `go list -m all | grep go-retry` MVS proof in integration. I:Low E:S C:Documentation
13. `scripts/example-smoke.sh` run against the new stack. I:Medium E:S C:Quality
14. Assert the `Server-Timing` header carries `total` (strings.Contains) in the server-level test, not just the whitebox one. I:Low E:S C:Quality

**The promoted program (post-repairs)**
15. httpx module skeleton (B-family home) + first battery — awaiting your httpx-vs-httputil read + the svc.Routes() call. I:High E:L C:Feature
16. W3-W5 blanket promotion: rewrite the TODO battery row from "demand-gated" to sequenced plan (B1-B9, D4/D7, W4 polite, W5 C1/C2/C3). I:High E:S C:Documentation
17. otel Recovery refactor is DONE — remove it from any open list and note the delegation in the otel module CHANGELOG entry. I:Low E:S C:Documentation
18. Version-drift guard: AGENTS deps-table == go.mod versions. I:High E:M C:Quality
19. Error-registry adoption: call `appkit.RegisterClassifications()` in the docs example + README error section recipe; errorpages/classified boundaries note. I:Medium E:S C:Documentation
20. Server-Timing composition E2E in integration (stream + header + knob-off) at the next re-pin train. I:Medium E:M C:Quality
21. Named Server-Timing phases: handler-side recipe section in README (db/cache examples) — the servable half of answer #13. I:Medium E:S C:Documentation
22. CORS battery row. I:Medium E:S C:Documentation
23. Trusted-proxy keying note (W3/W4). I:Medium E:S C:Documentation
24. Compression adopt-or-decline verdict row. I:Low E:S C:Documentation
25. `LoggingFromContext` adoption-or-N/A verdict. I:Low E:S C:Documentation
26. `WrapOnce` at idempotent Shutdown re-wrap sites. I:Low E:S C:Quality
27. `errorfamilytest` asserts in otel/systemd tests. I:Low E:S C:Quality
28. `errorfamily.WithContext` at drain-hook error sites. I:Low E:S C:Quality
29. go-sse next-release watch (breaking ssetest change; DefaultSubscriberBuffer reference). I:Medium E:S C:Documentation
30. go-etag release watch (root-shim v1.0.0 vs integration excludes). I:Low E:S C:Quality
31. go-idempotency release watch (CHANGELOG hygiene). I:Low E:S C:Quality
32. httputil next train: server_timing pairing + apidiff gate. I:Low E:S C:Quality
33. MaxBytesError→413 recipe: security module doc cross-link to the new classification. I:Low E:S C:Documentation
34. Consider `context.Canceled → Rejection` interplay with drain-window requests (document that in-flight canceled requests classify non-retryable at boot registries). I:Low E:S C:Documentation
35. End-to-end default-stack bench pair (old-stack vs new-stack) for the wire-cost story. I:Low E:S C:Quality
36. fleet-coherence note: ssehub `Drops()` reference for W5-C1 (carried). I:Low E:S C:Documentation
37. D4 floor note bump to ≥ v0.6.0 when activated. I:Low E:S C:Documentation (gated)
38. go-daemon B5-reference read at httpx design time (negotiation reference, not dep). I:Low E:S C:Documentation (gated)
39. systemd backlog (Examples, WithStatus, Install-returns-Hooks). I:Low E:M C:Quality
40. cordis bridge watch (triggers outstanding). I:Low E:L C:Feature (gated)
41. go-aichat reverse-adoption E2E (gated on their release). I:Low E:M C:Quality (gated)
42. cqrs EventConfig encryption/signing opt-ins (gated). I:Low E:L C:Feature (gated)
43. pbx-lens testkit migration pointer (their repo). I:Low E:S C:Documentation (gated)
44. samber/do auditlog + lazy-healthcheck ask (owner-gated). I:Low E:S C:Documentation (gated)
45. project-dependency-graph polish list (their repo). I:Low E:M C:Cleanup (gated)
46. ServerTiming knob: dependabot coverage check for the new direct require (root module config already sweeps go.mod — verify). I:Low E:S C:Quality
47. Docs: cross-link the knob from the SSE recipe (NoTimeout + knob interplay note). I:Low E:S C:Documentation
48. docs/DOMAIN_LANGUAGE.md: add Server-Timing/knob entry to the vocabulary table. I:Low E:S C:Documentation
49. AGENTS gotchas: add "grep the old contract string repo-wide at ship time" (this round's drift lesson, still unwritten). I:Medium E:S C:Documentation
50. Report-verify-block execution rule: add to the fleet-audit series README ("appendices are run, not transcribed"). I:Medium E:S C:Documentation

## g) Three questions I cannot figure out myself

1. **httpx vs httputil — the explanation you asked for is in my chat reply** (short version:
   httpx = the PLANNED handler-DX module for route/handler batteries — ResultHandler, bind+validate,
   content negotiation, error-response contract; httputil = the SHARED middleware/primitive library
   every project composes, which httpx will sit ON TOP of, not replace). Given that framing:
   **does the svc.Routes() seam get agreed now** (B6/B8 unblocked) **or when httpx reaches B6/B8?**
2. **Given the stamp-on-commit constraint** (a framework-measured app phase can never appear in
   the header for responses that write): **accept the knob + handler-side phases as the final
   answer** for Server-Timing coverage, **or do you want me to research a commit-deferring writer
   wrapper** (non-streaming requests only, SSE passthrough) as a possible later add-on?
3. **Commit grouping preference**: my plan is 5 logical commits (ST adoption+knob / otel delegation /
   fleet registry / audit reports / records+status) — **or do you want one squashed feature commit
   for today's code** with docs separate? (First commit will also carry the pre-commit guard run.)

— Report written 2026-11-03 21:25 CET. Awaiting instructions.
