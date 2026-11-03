# SUPERB v11 — Composition-Mission Execution Plan

> **2026-11-03 21:31 CET · go-appkit @ master · pareto-planning run**
> Goal: land today's composition-mission work (Server-Timing adoption, otel delegation, fleet
> registry, fleet audit) and sequence everything the 13 owner decisions promoted.
> Basis: the two status reports of today (19:54 / 21:25), the 13 questionnaire answers, TODO_LIST
> battery rows (read this session), and the verified state of the tree (all green: root+otel
> `-race`, lint 0 issues, guards green).

## Context — what already shipped today (do NOT re-plan)

- Fleet audit: `docs/review/libs/` (12 reports + index.html), supersede chains recorded.
- OWNER RULING: composition mission — family libs are IN-SCOPE; demand-gating them contradicts
  the repo's purpose. `docs/review/libs/` is the canonical home for per-lib deep-dives.
- server_timing ADOPTED: direct require v1.0.3, outermost in the default stack, tests + bench +
  CHANGELOG + README + DOMAIN_LANGUAGE + otel-README recipe drift-fixed.
- 13 decisions answered: knob=default-ON-config-knob ✅ shipped · train=ACCUMULATE · B5=FULL httpx
  program · ruling scope=BLANKET W3-W5 promotion · docs home=review/libs · otel refactor=THIS
  TRAIN ✅ shipped · commits=NOW (logical groups) · harvest=NOW · registry=BUILD NOW ✅ shipped ·
  consumer notices=BOTH · next workstream=REPAIR BACKLOG · svc.Routes()=pending owner (httpx
  explained, answer owed) · ST metrics=phases (partially unservable: stamp-on-commit; handler-side
  pattern shipped, framework-phase constraint documented).

## Pareto breakdown

### The 1% that delivers 51% (2 macro tasks)

1. **Land the tree**: truth-sync (CHANGELOG ×2, README knob row, AGENTS/TODO/ROADMAP/indexes
   decision lines) → five logical commits → push. Without this, nothing else exists or compounds.
2. **Decision records + canonical rule written down** (inside task 1's doc edits): the 13 answers,
   the docs-home rule, the blanket promotion — so no future session re-derives or contradicts them.

### The 4% that delivers 64% (5 more macro tasks)

3. Consumer heads-up drafts (cqrs-htmx + PapDashboard) in Lars's voice.
4. Full HARVEST of both status reports into TODO_LIST/ROADMAP (deduped vs battery rows).
5. Report-integrity repairs: fix the invalid rg verify-command; execute all 12 report
   verify-blocks as written; fix what fails.
6. Real ServerTiming bench baseline (`-benchtime 1s -count 6` + benchstat) into AGENTS perf
   section; assert `total` in the server-level test (not just whitebox).
7. Version-drift guard: AGENTS deps-table versions == go.mod versions, wired into pre-tag + CI.

### The 20% that delivers 80% (8 more macro tasks)

8. Misc repairs: hub.go:80 comment, mobile screenshot, dprint check, integration MVS proof,
   example-smoke run.
9. httpx module kickoff: skeleton + doc.go + B1 ResultHandler first slice.
10. httpx B5: content negotiation on go-codec (ForEncoding + Accept q-ranking + tests).
11. W3-W5 promotion rewrite: TODO battery row becomes a sequenced plan (B1-B9, D4/D7, W4 polite,
    W5 C1/C2/C3), blanket-promoted per ruling.
12. Registry adoption docs: README recipe, security BodyLimit cross-link, example wiring.
13. Server-Timing handler-phases recipe (README section) — the servable half of "phases".
14. AGENTS gotcha additions: contract-grep-at-ship rule; report-appendices-are-run rule.
15. ST composition E2E plan for the next integration re-pin (stream + header + knob-off).

### The remaining 80% → 100% (gated/background, scheduled not started)

- Gated batteries execution when sequenced: D7 idempotency adopt-wire, D4 atomic-write,
  W4 `polite` (go-retry + etagclient), W5 C1/C2/C3.
- Upstream watches: go-sse next release (breaking ssetest), go-etag root-shim v1.0.0, go-idempotency
  CHANGELOG hygiene, httputil pairing/apidiff.
- Their-repo items: pbx-lens testkit migration, project-dependency-graph polish, go-aichat
  reverse-adoption E2E, samber/do ask filing, cordis bridge triggers.
- Appkit backlog: LoggingFromContext verdict, compression verdict, WrapOnce, WithContext,
  errorfamilytest adoption, C2/C3 realtime completions, TLS seam spike (accumulated train).

## Macro plan — 30-100 min tasks (ALL todos, sorted by importance/impact/effort/customer-value)

| #   | Task                                                                                                                                                                                                   | Tier | Impact   | Effort | Customer value                         | Depends         |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---- | -------- | ------ | -------------------------------------- | --------------- |
| M01 | Truth-sync + decision records: core+otel CHANGELOGs, README knob row, AGENTS (ruling addendum, docs-home rule, 2 gotchas), TODO (blanket promotion + httpx), ROADMAP (httpx program), both index rules | 1%   | Critical | 60m    | Fleet coherence; no drift              | —               |
| M02 | Five logical commits (adoption / otel / registry / audit reports / records+plans) + push origin master + CI watch                                                                                      | 1%   | Critical | 40m    | Everything lands                       | M01             |
| M03 | Consumer heads-up drafts (cqrs-htmx, PapDashboard) via github-voice → docs/feedback/outgoing/                                                                                                          | 4%   | High     | 30m    | Consumers not surprised by wire change | M02             |
| M04 | Full HARVEST (docs-health): both status reports → TODO_LIST/ROADMAP, deduped                                                                                                                           | 4%   | High     | 45m    | Living backlog truth                   | M02             |
| M05 | Report integrity: fix invalid rg; execute 12 verify-blocks; fix failures                                                                                                                               | 4%   | High     | 45m    | Reports trustworthy                    | M02             |
| M06 | ST bench baseline (real `-benchtime`) + AGENTS perf row + server-level `total` assertion                                                                                                               | 4%   | Medium   | 40m    | Wire-cost is a published fact          | M02             |
| M07 | Version-drift guard script + pre-tag/CI wiring                                                                                                                                                         | 4%   | High     | 60m    | Drift class dies mechanically          | M02             |
| M08 | Misc repairs: hub.go:80, mobile shot, dprint check, MVS proof, example-smoke                                                                                                                           | 20%  | Medium   | 60m    | Debt zero                              | M02             |
| M09 | httpx module kickoff: skeleton, doc.go, B1 ResultHandler slice                                                                                                                                         | 20%  | High     | 100m   | The handler-DX program starts          | M02             |
| M10 | httpx B5: negotiation on go-codec + tests                                                                                                                                                              | 20%  | High     | 100m   | go-codec adoption becomes real         | M09             |
| M11 | W3-W5 promotion rewrite in TODO (sequenced plan)                                                                                                                                                       | 20%  | High     | 30m    | Ruling made executable                 | M04             |
| M12 | Registry adoption docs: README recipe, security cross-link, example call                                                                                                                               | 20%  | Medium   | 30m    | Consumers classify right at boot       | M02             |
| M13 | ST handler-phases recipe (README section, pre-write contract)                                                                                                                                          | 20%  | Medium   | 30m    | Consumers measure phases correctly     | M02             |
| M14 | ST composition E2E plan for next integration re-pin                                                                                                                                                    | 20%  | Medium   | 30m    | Pinned when pinnable                   | —               |
| M15 | Gated battery execution: D7 adopt-wire, D4 atomic-write, W4 polite, W5 C1/C2/C3 (sequenced, one per future train)                                                                                      | 100% | High     | L each | The mission's payload                  | M11             |
| M16 | Upstream watch execution (go-sse absorb, etag shim, idempotency hygiene, httputil pairing) at each next release                                                                                        | 100% | Medium   | S each | Fleet stays current                    | watches in TODO |
| M17 | Their-repo items (pbx-lens, pdg polish, go-aichat E2E, samber/do ask, cordis triggers)                                                                                                                 | 100% | Low      | S each | Fleet health                           | external        |
| M18 | Appkit backlog: LoggingFromContext verdict, compression verdict, WrapOnce, WithContext, errorfamilytest adoption, TLS spike                                                                            | 100% | Medium   | M each | Maximize httputil/err-family           | M02             |

## Micro plan — ≤12 min tasks (ALL todos, sorted same order)

| #       | Task                                                                                                         | Macro | Min              | Impact   |
| ------- | ------------------------------------------------------------------------------------------------------------ | ----- | ---------------- | -------- |
| m01     | core CHANGELOG: extend ST entry with knob + phases contract                                                  | M01   | 8                | Critical |
| m02     | otel CHANGELOG: Recovery delegation entry ([Unreleased])                                                     | M01   | 6                | Critical |
| m03     | core CHANGELOG: RegisterClassifications entry                                                                | M01   | 6                | Critical |
| m04     | README: ServerTiming knob row in config table                                                                | M01   | 5                | High     |
| m05     | AGENTS: ruling bullet addendum (knob, otel, registry, 13 answers)                                            | M01   | 10               | Critical |
| m06     | AGENTS: docs-home rule (review/libs canonical) + 2 gotchas (contract-grep, appendix-run)                     | M01   | 10               | High     |
| m07     | TODO_LIST: blanket W3-W5 promotion + httpx program decision + registry row                                   | M01   | 10               | Critical |
| m08     | ROADMAP: httpx program row + promoted batteries note                                                         | M01   | 8                | High     |
| m09     | docs/research/README + docs/reviews/README: canonical-home rule lines                                        | M01   | 5                | High     |
| m10     | git status → stage group 1 (adoption files) → commit #1 detailed                                             | M02   | 8                | Critical |
| m11     | stage group 2 (otel) → commit #2                                                                             | M02   | 4                | Critical |
| m12     | stage group 3 (registry) → commit #3                                                                         | M02   | 4                | Critical |
| m13     | stage group 4 (audit reports + index rows) → commit #4                                                       | M02   | 6                | Critical |
| m14     | stage group 5 (records + plans + status) → commit #5                                                         | M02   | 6                | Critical |
| m15     | git push origin master; watch pre-push hook; gh run list CI check                                            | M02   | 8                | Critical |
| m16     | Load github-voice skill; draft cqrs-htmx heads-up file                                                       | M03   | 12               | High     |
| m17     | Draft PapDashboard heads-up file                                                                             | M03   | 10               | High     |
| m18     | Load docs-health; HARVEST sweep of report-1 §f (dedupe vs battery rows)                                      | M04   | 12               | High     |
| m19     | HARVEST sweep of report-2 §f (items 15-50)                                                                   | M04   | 12               | High     |
| m20     | HARVEST: ROADMAP fuel entries (httpx program, gated batteries)                                               | M04   | 8                | High     |
| m21     | Fix invalid rg in httputil.md verify block                                                                   | M05   | 5                | High     |
| m22-m32 | Execute verify-blocks 1-11 as written; fix each failure found (one block per micro-slot)                     | M05   | 8 ea             | High     |
| m33     | ServerTiming bench `-benchtime 1s -count 6`; benchstat vs bare                                               | M06   | 12               | Medium   |
| m34     | Record baseline into AGENTS perf section                                                                     | M06   | 5                | Medium   |
| m35     | Add `total` assertion to server-level DefaultStackHeaders test                                               | M06   | 5                | Medium   |
| m36     | Write scripts/check-agents-deps.sh (table vs go.mod versions)                                                | M07   | 25→ split: 12+12 | High     |
| m37     | Wire guard into pre-tag-checks + CI job                                                                      | M07   | 12               | High     |
| m38     | Fix hub.go:80 comment (unexported const reference)                                                           | M08   | 5                | Medium   |
| m39     | Mobile-width screenshot of index.html                                                                        | M08   | 6                | Low      |
| m40     | dprint check over touched md/html; record coverage verdict                                                   | M08   | 8                | Low      |
| m41     | integration `go list -m all \| grep go-retry` MVS proof; correct reports if needed                           | M08   | 6                | Low      |
| m42     | Run scripts/example-smoke.sh against new stack                                                               | M08   | 12               | Medium   |
| m43     | httpx module skeleton: dir, go.mod (floor 1.27.2), doc.go, .golangci.yml (copy core pattern)                 | M09   | 12               | High     |
| m44     | httpx workspace-charter + dependabot + guards wiring (new module checklist)                                  | M09   | 12               | High     |
| m45     | B1 ResultHandler: type sketch + godoc + acceptance test stub (battery spec §B1)                              | M09   | 12               | High     |
| m46     | B5: codec choice mapping (ForEncoding + Accept parse) — design note in doc.go                                | M10   | 12               | High     |
| m47     | B5: negotiation handler impl + table tests (q-values, suffixes, q=0)                                         | M10   | 12 ea ×3         | High     |
| m48     | TODO battery row: rewrite to sequenced promoted plan                                                         | M11   | 12               | High     |
| m49     | Registry README recipe (boot call + scoped-registry alternative)                                             | M12   | 10               | Medium   |
| m50     | security THREAT_MODEL/README cross-link: MaxBytesError → Rejection classification                            | M12   | 6                | Medium   |
| m51     | example/main.go: call RegisterClassifications + comment                                                      | M12   | 5                | Medium   |
| m52     | README: Server-Timing handler-phases recipe section                                                          | M13   | 12               | Medium   |
| m53     | integration re-pin E2E plan row (stream+header+knob-off) into TODO                                           | M14   | 8                | Medium   |
| m54     | Consolidate upstream-watch rows (go-sse/etag/idempotency/httputil) into one watchlist item                   | M16   | 10               | Medium   |
| m55     | Their-repo pointer rows consolidation (pbx-lens, pdg, aichat, samber, cordis)                                | M17   | 10               | Low      |
| m56     | Appkit backlog rows: LoggingFromContext verdict, compression verdict, WrapOnce, WithContext, errorfamilytest | M18   | 12               | Medium   |
| m57     | TLS spike plan row (httputil v1.7.0 seam) into ROADMAP                                                       | M18   | 8                | Medium   |
| m58     | Session-close ritual: verify CI green on final pushed commit; record results                                 | M02   | 10               | High     |
| m59     | Errata pass: annotate both status reports with "items landed" markers                                        | M04   | 8                | Medium   |
| m60     | Final sweep: `rg "Recovery → RequestID"` contract grep + full guards                                         | M01   | 8                | Critical |

## Execution graph

```mermaid
flowchart TD
    A[Green tree: ST stack + knob + otel delegation + registry] --> B[M01 Truth-sync + decision records]
    B --> C[M02 Five logical commits + push + CI watch]
    C --> D1[M03 Consumer drafts]
    C --> D2[M04 Full harvest]
    C --> D3[M05 Report integrity]
    C --> D4[M06 Bench baseline + total pin]
    C --> D5[M07 Version-drift guard]
    C --> D6[M08 Misc repairs]
    D2 --> E1[M11 W3-W5 promotion rewrite]
    D3 --> F1{All 12 verify-blocks pass?}
    F1 -- no --> F2[Fix failures, re-run] --> F1
    F1 -- yes --> G[Repair backlog DONE]
    D4 --> G
    D5 --> G
    D6 --> G
    C --> H[M09 httpx kickoff]
    H --> I[M10 B5 negotiation on go-codec]
    E1 --> J[M15 Gated batteries per train]
    I --> K{Owner: svc.Routes() seam?}
    K -- agree --> L[B6/B8 unblocked]
    K -- later --> M[B1/B2/B5/B9 proceed without introspection]
    G --> N[M12-M14 adoption docs + recipes + E2E plan]
    C --> O[Session-close ritual m58: CI green on final commit]
```

## Verschlimmern guardrails

- No rewrites of point-in-time docs (status/research/review) — dated errata only.
- No new module without the full new-module checklist (charter, dependabot, lint config, guards).
- Every commit leaves `go build`/`go vet`/`-race`/lint green — verified before `git commit`, hooks
  re-verify (directives, dependabot parity, workspace charter, sse-protocol).
- Push only after all five commits are individually green; CI watched on the final commit
  (session-close ritual).
- Integration module stays NON-workspace-member; any httpx module joins the charter via the
  computed checklist, never by hand-editing go.work.
