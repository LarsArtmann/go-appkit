# Status Report: art-dupl "Deduplicate to Zero" Session

**Date:** 2026-11-13 14:43 CET · **Scope:** this session only (art-dupl sweep + dedup refactor). No other repo areas researched, per instruction.

## Scope Line — what this report VERIFIED vs NOT

**VERIFIED (command/artifact):**
- All 15 reported clone groups read at source and adjudicated (staleness_test.go, httpx trio, root metrics/testhelpers/testkit, examples, otel) — every accept verdict is grounded in the actual code, not the tool's suggestion.
- Extraction verified: `cd cqrs && GOWORK=off GOTOOLCHAIN=go1.27.3 go vet ./...` clean; `go test ./... -race -count=1` ok (4.7s); `go test -race -count=5 -run 'CheckStaleness|CheckProjectionStaleness'` ok (1.8s, flake spot-check); `golangci-lint run ./...` → 0 issues; `gofmt -l` clean.
- Detector re-run in BOTH modes after the refactor: user mode (`art-dupl --sort total-tokens -t 1`) → staleness group **gone**, 14 groups remain; gate mode (`art-dupl . --type-aware -t 1`) → 8 groups remain, all accepted classes. No new group formed from the five helper call sites.
- Gate: `./scripts/check-art-dupl.sh` → `OK`, exit 0, against the updated 8-entry baseline (jq-validated).
- Docs: AGENTS.md 153/220 lines (cap respected); the edited posture line pre-flighted through `docs-gates freshness . --pre-flight` → **SAFE, no historical item flips to DRIFT**.
- Git: all session commits pushed, master synced with origin (`ebf58d0` extraction, `e983a84` changelog/baseline/snapshot, AGENTS.md follow-up — all daemon commits).

**NOT VERIFIED (and why):**
- **CI green on the final pushed commit** — the tip run (43435724401) was still `queued` at 14:40. Session-close ritual incomplete; re-check needed.
- **Tonight's govulncheck nightly FAILURE** (run 43358725891, 2026-11-13 03:24Z) — noticed, not triaged (predates this session's commits by hours; out of session scope, flagged in (f)#2).
- The 1:1 mapping of the 8 baseline entries to the 8 current gate-mode groups — the gate proves *no new* groups; *stale-entry* hygiene beyond the one entry I removed is unproven.
- cqrs-lint scorecard (ritual applies to wrapper *feature* changes; this was test-only — deliberately skipped), fresh-consumer proxy check (no tag released).

---

## Session Summary

Ran the full adjudication of the user's `art-dupl -t 1` report (15 groups). One harmful clone existed; everything else was an accepted-by-design class. Extracted the harmful one, re-baselined truthfully, updated docs, and re-verified everything green.

| Group | Content | Verdict |
|---|---|---|
| #1 cqrs staleness_test.go ×5 (45 tok) | 5× identical projection bootstrap, varying only name/event-type | **EXTRACTED** → `startNoOpProjection` (staleness_test.go) |
| #5/#6 root testhelpers vs testkit twins | `waitForRunning`/`httpGet` clones of published `testkit` helpers | **ACCEPT** — structurally forced: testkit imports appkit, so in-package `appkit` tests cannot import testkit (import cycle) |
| #4 Flusher passthrough ×4 | httpx `Writer`/`imsWriter`, root `statusRecorder`, httpxtest `honestWrapper` | **ACCEPT** — cross-module idiom (root can't import satellites); the test double's similarity IS its purpose |
| #2/#3/#7 example mains | `DefaultServiceConfig()` bootstrap + consts | **ACCEPT** — self-contained demos (existing posture) |
| #8-#15 | opts loops ×7 groups, otel err-checks ×4, span guard ×2, Option aliases ×2 | **ACCEPT** — existing posture idioms |

---

## a) FULLY DONE

1. **The only harmful clone extracted.** cqrs `staleness_test.go`: 5×45-token copy-paste bootstrap → one `startNoOpProjection(t, eventSvc, name, eventType)` helper + five one-line call sites (~60 lines removed from the file). Commit `ebf58d0`.
2. **Zero harmful duplication** (skill definition: zero harmful, not zero report lines). Both detector modes confirm: no unadjudicated group remains.
3. **Verification battery green**: vet, `-race` tests (1× whole module + 5× staleness-only), golangci-lint 0 issues, gofmt clean.
4. **Gate green + baseline truthful**: dead entry `93a9b09aaa101148` removed from `scripts/art-dupl-baseline.json` (9→8 entries, jq-valid); `check-art-dupl.sh` OK. CI-safe (targeted removal, all other hashes byte-identical).
5. **Docs & memory updated**: AGENTS.md Clone posture bullet extended (2026-11-13 sweep: extraction + two new accepted classes with rationale + snapshot pointer); cqrs `CHANGELOG.md` `[Unreleased]` entry added; dated snapshot `scripts/art-dupl-gate-snapshot-2026-11-13.txt` written per the gate script's header convention.
6. **docs-gates pre-flight on the AGENTS.md edit**: SAFE.
7. **Everything committed and pushed** (daemon), master == origin/master.

## b) PARTIALLY DONE

1. **Session-close ritual** — commits pushed and local gates green, but the final CI run was `queued` at report time and the nightly govulncheck failure (today 03:24Z) is untriaged. The ritual's "verify green on the FINAL pushed commit" is not closed.
2. **"Zero"** — achieved under the skill's definition; a strictly empty report was never the goal (14 user-mode / 8 gate-mode groups remain deliberately, all with recorded rationale).
3. **Self-review** — the brutal-self-review skill wants its own `docs/reviews/*.html` series entry; per your instruction it is folded into THIS report instead (sections d/e). Prior `docs/reviews/` entries were not read (out-of-scope research ban).

## c) NOT STARTED (deliberate)

1. TODO_LIST/ROADMAP **harvest** of section (f) — awaiting your instruction ("then wait").
2. cqrs **D014**: `FoldedEvent.occurredAt` lacks a json tag — a wire-format decision, not a drive-by fix.
3. cqrs `sse_test.go:174` **scanner.Err()** check missing (gopls warning, pre-existing).
4. **Split brain**: root `CHANGELOG.md` `[Unreleased]` carries a `cqrs:` bullet while cqrs has its own CHANGELOG — one home needed.
5. `check-art-dupl.sh` **stale-entry detection** (today it only fails on new groups).
6. art-dupl **local binary ↔ CI pin** equality check (only blocks future baseline *regenerations*; my targeted removal was pin-safe by construction).
7. Annotating the 2026-08-16 staleness-merge suggestion (#42) that docs-gates surfaced — superseded by the helper extraction.

## d) TOTALLY FUCKED UP

**Nothing shipped broken.** Honest near-misses, in full:

1. **Initial misdiagnosis**: I read the HTML anchor ids as baseline hashes, concluded 4 groups had "drifted" and the gate might be red — wrong mental model (anchor ids are a different hash flavor than check-hashes). I also asserted "the 9-entry baseline covers all 15 groups" *before* understanding the mechanism: conclusion true, reasoning not. Running the gate (the authoritative signal) corrected it. Cost: ~2 wasted tool calls, no damage.
2. The first multiedit left `staleness_test.go` non-compiling between steps (expected in a two-step extraction; never shipped, immediately fixed).
3. No data loss, no reverts, no force operations. The only destructive edit (baseline entry removal) is verified safe by the gate.

## e) WHAT WE SHOULD IMPROVE

1. **Run the cheap authoritative gate FIRST**, then reason — I inferred hash mechanics before measuring. The gate answers the coverage question in one call.
2. **Pre-flight docs-gates freshness at edit time**, not retroactively — I remembered it only during report writing.
3. **Record noticed-but-not-mine findings durably in-session** (D014, scanner.Err(), govulncheck red) — they lived only in my context and would have been lost without this report.
4. **`-count=5` flake spot-checks should be default** for deadline-bearing tests (`waitFor`, SSE deadlines), not a report-time afterthought.
5. **Document structural constraints at the duplication site**: the root `waitForRunning`/`httpGet` twins survive *because* of the testkit→appkit import cycle; without an in-place comment, the next dedup sweep may "fix" them into a compile error.
6. **Split-brain hygiene**: pick ONE changelog home for unreleased cqrs entries (module CHANGELOG is the documented convention; the root `cqrs:` bullet violates it).
7. What went right and should persist: every accepted verdict was grounded by reading the code (re-verification over document-trust); surgical scope (one module, no drive-by fixes); acceptance recorded as gate+baseline, not prose.

## f) Up to 50 things to get done next

Quality over padding: 35 real items, impact-ordered. `[S]` = this session, `[R]` = known repo queue from loaded context. Most `[R]` items are ROADMAP fuel (docs-health HARVEST applies routing rigor).

**P1 — close this session out**
1. `[S]` Verify CI green on final pushed commit (run 43435724401 was queued) — incl. the `art-dupl-gate` job leg.
2. `[R]` Triage tonight's govulncheck nightly failure (run 43358725891, 2026-11-13 03:24Z; fleet was 14/14 clean 2026-11-07).
3. `[S]` HARVEST this section into TODO_LIST/ROADMAP (docs-health) — pending your go-ahead (question g3).
4. `[S]` Add the structural-constraint comment at root `testhelpers_test.go` (testkit import-cycle rationale) so the twins stop looking extractable.
5. `[S]` Sweep the remaining 8 baseline entries against current groups; prune any other stale entry (incl. verifying the `failingServiceNames` posture claim still maps to a live group).
6. `[S]` Extend `check-art-dupl.sh` with stale-entry detection (exact-count semantics, go-etag precedent) — policy call, see question g2.
7. `[S]` Record art-dupl binary-version equality check (local nix binary ↔ CI pin `v0.9.1-0.20261017225421-93432fd910fb`) as a pre-step in the gate script header's re-baseline procedure.

**P2 — cqrs module follow-ups (next cqrs train)**
8. `[R]` D014: `FoldedEvent.occurredAt` json tag — wire-format ruling needed (question g1).
9. `[S]` Fix `sse_test.go:174` missing `scanner.Err()` final check (test-only, minutes).
10. `[R]` Resolve root-CHANGELOG `cqrs:` bullet placement (split brain, items c4).
11. `[S]` Annotate/close the 2026-08-16 staleness-merge suggestion #42 (superseded by helper extraction).
12. `[S]` Decide: generalize `startNoOpProjection` with a handler param for other projection-test sites (flightrecorder/metrics/readiness) — or record declined (parameter-count rule).
13. `[S]` Consolidate cqrs test helpers (`newTestEventService`/`appendTestEvent`/`waitFor` scattered across testhelpers_test.go + logger_test.go) into one helpers file — cohesion, not duplication.
14. `[R]` gopls `infertypeargs` infos (7 sites in cqrs commands*/sse_test) — fix or deliberately suppress.
15. `[R]` cqrs-lint scorecard re-run rides the next cqrs feature train (ritual).

**P2 — dedup tooling cadence**
16. `[S]` Schedule the periodic deep audit from the gate script's scope note (`art-dupl . --type-aware -t 1 --test-threshold 1 --rich-text`) — put a cadence on it.
17. `[S]` Add generator-command header to dated snapshots (consistency with future ones).
18. `[S]` Document in the gate header that the *user-mode* pass (no `--type-aware`) shows a wider set (14 vs 8) — sweeps should reconcile both, as AGENTS.md now records.

**P3 — known repo queue (from AGENTS.md, untouched this session)**
19. `[R]` Consumer notices: cqrs-htmx + PapDashboard adoption drafts still pending (13-decision questionnaire).
20. `[R]` `svc.Routes()` seam answer owed to owner; httputil/endpoints v0.1.0 verdict + H8/B6/B8 fit (`docs/review/libs/endpoints.md`).
21. `[R]` W5-C1: go-sse `WithOnDrop` TODO; watch go-sse's unreleased breaking `ssetest` at next re-pin.
22. `[R]` httputil server-seam refactor (lib score 72/100; CORS/trusted-proxy/compression unadopted).
23. `[R]` httpx B6/B8 remain H8-owner-gated; W4 `polite` battery (go-etag client+metrics v0.6.0 relevance).
24. `[R]` go-error-family utilization (46/100): registry/Classify/WithContext/errorfamilytest engagement — demand-gated.
25. `[R]` Core TLS option — deferred register (PapDashboard `PAP_TLS_CERT/KEY` trigger).
26. `[R]` Cordis bridge — 1 of 3 triggers met; await consumer statement + core v1.0.0 exit criteria.
27. `[R]` PapDashboard reverse-adoption watch (their repo ships the appkit-hosted release).
28. `[R]` Verify the fresh-consumer proxy-check re-run for preset/config actually rode the v13 close-out.
29. `[R]` go-error-modernization sweep (`errors.AsType`) — fleet-wide, not started.
30. `[R]` integration `documentedPins` bump rides next release train (standard ritual).
31. `[R]` Keep AGENTS.md ≤220 lines (now 153 — headroom fine).
32. `[R]` Standing rule reminder for future docs: routed markers must use the `**→ TODO P1|P2|P3 — fragment**` grammar.
33. `[R]` Example mains' hand-rolled PORT env → `appkit/config.Load` dogfood — only if owner flips the self-contained-demos ruling.
34. `[R]` Adopt "module CHANGELOGs only" for unreleased cross-module notes after item 10 lands.
35. `[R]` Watch the nightly `flake` job (green today) alongside govulncheck until #2 is triaged.

Items 36-50: **declined as padding** — anything further would be invention, not observation; ROADMAP fuel belongs to the docs-health HARVEST pass, not this snapshot.

## g) Questions I can NOT figure out myself

1. **Wire contract (D014)**: `FoldedEvent.occurredAt` shipped in cqrs v0.9.0 without a json tag, so JSON encoding uses `occurredAt` (Go field name). Adding `json:"occurredAt"` is a no-op for that spelling but locks the contract deliberately; suppressing D014 records it as intentional. Consumers may already rely on either. **Which way do you rule — tag it or suppress with reason?**
2. **Gate strictness policy**: should `check-art-dupl.sh` move from new-groups-only to exact-count/stale-entry-failing semantics (go-etag precedent, stricter, catches accepted groups silently disappearing or rotting), or stay permissive? Stricter = more re-baseline friction on every refactor.
3. **Harvest authorization**: may I run docs-health HARVEST on section (f) now — and if so, only the P1 `[S]` block into TODO_LIST, or the full 35-item list with `[R]` items routed to ROADMAP?

---

*Awaiting instructions.*
