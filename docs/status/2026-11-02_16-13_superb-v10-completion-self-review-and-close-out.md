# SUPERB v10 Completion, Self-Review, and Close-Out

**2026-11-02 16:13 CET · Scope:** the continuation window that executed M15-closeout through M26 + F95 after the owner's "GET SHIT DONE" re-dispatch, plus a brutal self-review of what this window forgot, did badly, and should improve. Part 1 (M3–M15 execution + corpus sweep) lives in this same file above; the 12:25 narrative covers that half.

**State at close:** tip `2d8db0d` == origin/master. CI on the FINAL pushed commit: every job green except `ci-preflight` (designed chronic red — Actions `SSH_PRIVATE_KEY` secret empty) and `structure-linter` (skips by design, rides ci-preflight). Root + testkit + integration `-race` suites green hermetically; root and integration `golangci-lint` **0 issues**; routed-destinations 194/0; directives, charter, dependabot-parity guards green.

---

## a) FULLY DONE

1. **CI triage (both §d "untriaged" rows from part 1 closed):** the tip reds are the DESIGNED chronic class only. `ci-preflight` fails on the empty Actions secret; `structure-linter` skips by design. Part 1's two misreads corrected on the record: run `41426686086` is the _Dependabot Updates_ workflow (not CI — "cbcfeba full SUCCESS" was wrong), and cbcfeba's real CI run `41426476435` is identical in shape to tip (ci-preflight red + structure-linter skip). The ONE real red found anywhere: `structure-linter` fails at INSTALL on dependabot-PR runs (`git ls-remote` exit 128 — Dependabot-secrets context lacks working auth for the private-binary `go install`). Environmental, not code/config; the "daemon config convergence broke it" theory is dead (the PR branch predates that commit).
2. **M15 fully closed:** PR #53 MERGED (errorpage v1.25.1 across 9 modules, matrix green, only the environmental structure-linter red); #46/#28 closed with supersession comments citing `a467328`; #47/#39/#48/#50/#44 were already dependabot-auto-closed (10:11–10:12Z, post-daemon-wave). Zero dependency stragglers remain. Integration `go mod tidy -diff` green hermetic (the lost verification). Lost local verify folded to runner truth per the owner question-3 default.
3. **M16 (F71–F73):** spike doc §5 addendum maps ALL THREE pins against httputil v1.7.0's `NewServerListener` godoc: bind contract SOLVED (appkit keeps the bind + sync `listen_failed`), `Addr()`/`Running()` SOLVED (one listener, two views), phase logs appkit-owned with the single-timeout discipline (`ServerConfig.ShutdownTimeout` from the SAME config field as the ctx deadline). NEW semantics pinned: single-serve ownership → construct a FRESH httputil.Server per `Start`. Swap plan-of-record written into the TODO composition row. The swap itself deliberately NOT done (separate core train).
4. **M17 (F74–F76):** `APIKeyMatches` census 0 runtime callers in-repo + 0 in PapDashboard (they hand-roll their own `internal/middleware.APIKeyAuth`); §7.3 demand signal cross-checked; **KEEP** verdict written to TODO row with the re-review trigger (next auth-touching security train with still-zero adoption AND zero callers). Removal declined: it is the anti-drift helper and the module-level demand signal went UP.
5. **M18 (F77–F79):** `integration/security_errorpages_test.go` green under `-race`. Truth-pins: Mount's catch-all SWALLOWS method mismatches into pretty 404s (the Mount godoc's "plain 405 with Allow" only holds WITHOUT a catch-all registered — Wrap owns pretty 405s); pretty pages are styled `<main>` FRAGMENTS, not full `<html>` documents; they carry the default-stack security headers (X-Content-Type-Options/X-Frame-Options/Referrer-Policy); auth 401 stays plain JSON; Accept: application/json gets the JSON contract; classified `Write` → error-family status with headers.
6. **M19 (F80–F81):** `integration/flightrecorder_e2e_test.go` green under `-race`. Truth-pins: `fr.New` requires explicit `rec.Start()` (without it: zero captures AND 503 downloads); ONE runtime trace per process, so the Start→Close window must stay sequential vs health_stack's recorder (top-level `t.Parallel()` deliberately dropped, documented in-test); snapshot-dir sink retains one file per capture; non-matching triggers capture nothing; `?download=1` streams octet-stream; snapshot file EXISTENCE races the trace flush (poll for non-empty; 200 ms buffer warm-up mirrors health_stack's posture).
7. **M20 (F82):** frh pin-vs-MVS shape note in AGENTS (frh = health = integration all pin go-health v0.8.1 directly — the 2026-09-29 `go work sync` bite class is dormant unless a member pins newer); samber/do lazy-healthcheck draft ruled **FILE-WORTHY** (premise re-verified at samber/do master: `healthcheck()` still returns nil for unbuilt lazy, README still silent; NEW evidence strengthens it — `isHealthchecker()` is false-until-built, so the green-page mechanism may be exclusion rather than nil-pass; filing stays owner-gated); huma.md Deferred Register row RESOLVED (pushed at httputil `108b3bf`, raw render verified).
8. **M21 (F83):** Owner execution checklist added to the pma decision memo — 5 concrete steps: hook bypass (no action), train-repo cooldown (implement), staged-message respect (implement), go.work mutator (closed), and the FINAL `ignore_patterns` ask text with both glob dialects (`.*-body.html` basename vs `docs/**/.*-body.html` doublestar) + restart/verify steps.
9. **M22 (F84):** accept-cadence note written — ZERO gopls-class go.work leaks in the three weeks post-BuildFlow#46 (evidence: only two post-10-12 commits touch go.work, both legitimate train content — `24a0736` charter cleanup, `503a290` the daemon riding the 1.27.1→1.27.2 floor line; charter guard green throughout; journalctl path dead per memo evidence 1). The inotify hunt stays unconditionally justified, not omitted.
10. **M23 (F85):** core-v1 graduation table added for 2026-11-02 — criterion 1 **ARMED**: release-ritual API diff v0.9.0→working tree is ZERO lines (23 days additions-free), so the next core tag cut from this tree IS the additions-empty anchor that starts the freeze cycle; criterion 5 unchanged (claims guard green same day: setup/v4.15.0 → core v0.9.0 = latest; PapDashboard still v0.4.0, their refresh their call). TODO row updated.
11. **M24 (F86–F87 + F95):** watchlist as-of refresh with honest limits — nixpkgs-unstable go **1.26.10** at `0bb2eedc` (BELOW our 1.27.2 floor: nixpkgs-side drift with NO upgrade path through nixpkgs; the row's earlier "1.27.2 matches our floor" was the 10-13 snapshot); otel v1.48 still unreleased (v1.47.0 latest stable, proxy-checked); semconv still v1.26.0; OTLP-logs gate audit stands (claims guard re-ran green same day); go-aichat convergence verified (go-sse v0.6.2 + ssetest v0.5.0 match the fleet exactly, no go-appkit/go-health direct deps); cordis as-of honestly limited (local fork clone carries NO tags — the 09-29 verification stands unrefreshed, recorded as such instead of fabricated).
12. **M25 (F88–F89):** `integration/i18n_language_test.go` green — primary-subtag fallback (da-DK→da), q-weight ordering (da;q=0.9 beats en;q=0.2), unsupported-fr and missing-header fall to default en, Content-Language + Vary: Accept-Language ride every negotiated response through the full default stack, `LanguageFromContext` populated. TODO i18n row: the demand-gated remainder is now DONE.
13. **M26 (F90–F94):** testkit gained `HTTPGetURL` / `HTTPGetContext` / `WaitForRunning` / `WaitForRunningWithin` / `FreeAddr` (the rebind-verified issued-registry reserve migrated from integration) + unit tests + core CHANGELOG `[Unreleased]` entry + README section + FEATURES rows (marked UNRELEASED). Root `appkit_test` migrations done (`metrics_test.go` getURL → `testkit.HTTPGetContext`). **Discovered constraints, now documented:** (1) an in-package (`package appkit`) test file CANNOT import testkit — import cycle — so the whitebox helpers in `testhelpers_test.go` stay local permanently (my first attempt was reverted); (2) satellite + integration migrations ride the next core train because hermetic proxy resolution cannot see unreleased API and integration's charter forbids unreleased pins.
14. **F95:** the exact-token guard lesson (marker tokens are APIs: single definition, derived matchers, near-miss tests, positive exemption fixtures, atomic kind rejection) committed in **crush-config** `227a130` — LOCAL, unpushed (push decision left to owner).
15. **Fleet hygiene riding the wave:** the new direct require (go-appkit/flightrecorder v0.2.0 in integration/go.mod) was caught by the `documentedPins` fixture on the first full-suite run — the guard working as designed; fixture extended. TODO header re-stamped; FEATURES truthed; status-report part 2 appended.

## b) PARTIALLY DONE

1. **Local verify coverage is NEW-CODE-scoped, not whole-fleet:** the three suites I touched ran hermetically under `-race`; the other nine modules' suites were NOT re-run locally this window — they ride runner truth (the 2d8db0d CI matrix is fully green, which covers them). Per the owner question-3 default this is accepted posture, but it is a scoping choice, not an omission to hide.
2. **crush-config lesson:** committed, NOT pushed — deployment (home-manager activation reads the repo) waits on the owner's push call.
3. **cordis consumer re-verification:** honestly inconclusive (the local fork clone carries no tags), so the as-of stays 2026-09-29 rather than a fresh claim.
4. **The 12:25 report's open questions:** answered by execution this window (defaults taken per the re-dispatch), but the report text itself still says "awaiting instructions" — superseded by this close-out, formal annotation left for the next docs pass.
5. **`docs/status/README.md` index:** the part-1 report was indexed in the 12:25 window; THIS close-out section lives in the same file, so the index needs no new entry — but the index's one-line summary for the 12-25 report does not yet mention part 2 (minor staleness, next docs pass).

## c) NOT STARTED (all deliberately gated, none forgotten)

1. **The composition swap itself** (Service on httputil.Server) — mapping + plan-of-record done; the refactor is its own core train (owner-gated timing).
2. **Core TLS option** — rides the swap train (StartTLS comes free).
3. **The additions-empty anchor core tag** (v0.9.x doc-only) that starts the v1.0.0 freeze clock — criterion 1 is ARMED, the tag fires on the next natural core train; owner timing.
4. **Satellite + integration migrations to the new testkit helpers** — blocked by design on the next core tag (hermeticity + pin charter).
5. **Restore the Actions `SSH_PRIVATE_KEY` secret** — owner-only (Settings → Secrets).
6. **The structure-linter-on-dependabot-PRs CI fix** — needs an owner decision (skip the job on dependabot events vs fix SSH setup in that context); recorded as a follow-up, not built.
7. **samber/do ask filing** — draft ruled FILE-WORTHY; filing is explicitly owner-gated per the draft's own header.
8. **Battery waves W3–W5, cqrs opt-ins, PapDashboard reverse adoption, cordis bridge** — all demand/trigger-gated rows, untouched by design.
9. **fr#2 cooldown combinator adoption** — upstream still OPEN; frh hand-rolled cooldown stands.
10. **pma repo changes** (cooldown, staged-respect, ignore_patterns) — owner-only file/service; the checklist is ready, execution is Lars's.

## d) TOTALLY FUCKED UP (the honest list)

1. **I pushed a fleet-lint-red commit.** `7f78999` passed root lint locally (0 issues) but I never linted the INTEGRATION module I had just added three test files to — despite the AGENTS rule "lint each module from its own directory" sitting in the file I'd read an hour earlier. CI caught it (fleet-lint red on integration), I fixed it in `2d8db0d`, but the damage class is exactly the "unverified gate" lesson: a green LOCAL gate on a DIFFERENT module than the one that changed is not verification. One wasted CI cycle and a red commit permanently on master history.
2. **Shell-splice surgery on `security_errorpages_test.go`:** I attempted awk/python heredoc restructuring of the gocognit problem and produced TWO syntax-broken intermediate states (misjudged closure bounds twice, then compounded it with a second splice). Had to `git checkout --` the file and rewrite it whole with `write`. Every intermediate was recoverable, but the whole detour was avoidable: for whole-function restructures the house answer is one clean full-file rewrite or `lsp_replace_symbol`, not stream editing.
3. **The `documentedPins` miss:** I ran `go get github.com/larsartmann/go-appkit/flightrecorder@v0.2.0` in integration without predicting that the pin fixture (whose entire job is catching undocumented requires, and which AGENTS documents) would fail. The guard caught it on the first full-suite run — correct behavior, but I should have updated the fixture in the same breath as the `go get`.
4. **The tparallel flip-flop:** I wrote the flightrecorder subtests with `t.Parallel()`, was flagged, ADDED t.Parallel to the parent-adjacent places, got flagged the opposite way, then finally removed subtest parallelism entirely — when the correct posture (fully sequential, one runtime trace per process) was decidable BEFORE writing the test. Same flip-flop pattern on wsl_v5 blank lines in the testkit test (three iterations where one look at the rule would have settled it).
5. **Background-job results lost twice** (`job_output` interrupted in the prior window, shell IDs evaporating across the interruption boundaries): two test runs' verdicts evaporated and had to be re-run. The durable-capture discipline from part 1's gotchas (capture immediately into notes) was not applied to my own background jobs.
6. **Trusting a stale summary:** part 1's handoff summary claimed "run 41426686086 on cbcfeba = full SUCCESS" and "structure-linter NOT designed — UNTRIAGED". Both wrong. I corrected them early, which is right — but the correction only happened because I re-derived from `gh run view` instead of trusting the handoff, and the handoff had been written as authoritative. Verification of inherited claims (lessons.md 2026-11-01 class) should have been the FIRST action of the window, not the first triage action.

## e) WHAT WE SHOULD IMPROVE

1. **Pre-push module-scoped gate:** never push after touching module X without running X's own lint + vet hermetically. A 30-second local fleet-lint of the touched modules would have saved the red commit. Candidate mechanical fix: a tiny `scripts/check-touched-modules.sh` (git diff --name-only → map to modules → lint each) wired into pre-commit.
2. **Decide test parallelism posture at DESIGN time:** the composition tests share process-global resources (runtime trace, capture dirs). The rule "does this test hold a global? → sequential parent, sequential subtests" should be written down (AGENTS Testing section) so the next E2E doesn't rediscover it via lint ping-pong.
3. **Whole-file rewrites over stream surgery:** any restructuring that touches >30% of a file goes through one `write` (or `lsp_replace_symbol` per function), never awk/python heredocs. The splice detour cost four tool calls and two broken states.
4. **Fixture updates ride dependency changes:** `go get` in integration ⇒ update `documentedPins` in the same edit. Same shape as "CHANGELOG dating rides the release train": group the mutation with its guard.
5. **Background jobs: capture or kill.** Every backgrounded command gets its result read into durable notes (or the session log) before anything else; interrupted windows lose shell IDs, and /tmp evaporates. The part-1 gotcha, re-learned at my own cost.
6. **Inherited-state verification first:** the first action after any handoff summary is re-deriving its load-bearing claims (run IDs, job states, PR states) from the source. Two of part 1's claims were wrong and one of them had me hunting a nonexistent CI failure for a while.
7. **Lint the new-file's module, not the module you're "in":** my mental model anchored on the root module because that's where I started the window. Scope gates to the DIFF, not the session's origin.
8. **godoclint package-doc convention:** new integration test files should NOT carry attached file-header comments (the fleet allows one package doc per package; multiple files with headers trip `godoclint` for ALL of them). Write headers detached from the package clause from line one.

## f) NEXT THINGS (up to 50, ordered by impact)

1. **Restore the repo Actions `SSH_PRIVATE_KEY` secret** (owner) — retires the chronic ci-preflight red and re-arms structure-linter on push runs.
2. **Fix structure-linter on dependabot-PR runs** — either `skip` the job for dependabot actors or wire working SSH auth into that context (every future dependabot PR reds on it until then).
3. **Cut the additions-empty anchor core tag** (v0.9.1, doc+testkit-additions) — starts the v1.0.0 freeze clock; criterion 1 is armed at a zero-line diff.
4. **Execute the composition swap train** (Service → `NewServerListener`) per the TODO plan-of-record; Core TLS rides it.
5. **Push crush-config `227a130`** so the exact-token lesson deploys through home-manager.
6. **File the samber/do lazy-healthcheck ask** (owner-gated; fold in the `isHealthchecker()` exclusion nuance at filing time).
7. **Satellite testkit migrations** (realtime/fr/cqrs `httpGetURL`, `mustDo`, poll spellings) on the next core train.
8. **Integration migrations** (`e2eGet`→`testkit.HTTPGetURL`, `freeAddr`→`testkit.FreeAddr`, `waitRunning`→`WaitForRunning`) after the same tag.
9. **Add `HTTPPostJSON` to testkit** — the inventory found five GET families but MISSED the POST family (`flightrecorder_test.go` hand-rolls `httpPostJSON`); fold it into the same testkit train.
10. **errorpages module follow-up:** clarify the `Mount` godoc — the "plain 405" claim does not hold when a catch-all is registered (my M18 test proved the actual composition); a two-sentence godoc fix on errorpages' next train.
11. **PapDashboard next-tag watch** — claims guard fires on their refresh; criterion 5 overlap unblocks then.
12. **cqrs-htmx v5 window (ADR-0052)** — their repo; unblocked.
13. **go-flightrecorder#2 cooldown combinator** — upstream OPEN; when shipped, delete frh's hand-rolled cooldown on frh's next train.
14. **Machine-default go upgrade decision** (standing since 09-29): nixpkgs can't provide it (1.26.10 drift) — needs an explicit owner path decision (flake-pinned toolchain vs autoload).
15. **pkg.go.dev flightrecorder 404 watch** — owner-gated re-tag last resort stands.
16. **Flip the FEATURES "UNRELEASED" testkit rows** when the anchor tag ships.
17. **Retire the composition-gap watchlist row** (NOTHING REMAINS) at the next docs-health pass, or keep it as the landing pad for new gaps — owner taste.
18. **Annotate the 12:25 report's "awaiting instructions" line** as superseded-by-close-out.
19. **Update `docs/status/README.md`'s 12-25 index line** to mention part 2 (minor staleness).
20. **Write the "test global posture" rule into AGENTS Testing section** (improvement #2 made durable).
21. **Build `check-touched-modules.sh`** (improvement #1 made mechanical) and wire into pre-commit.
22. **Retire TODO row 29's remaining half** after satellite/integration migrations land (the dedup row completes).
23. **Battery W3 `httpx` module** — demand-gated; re-check demand at next consumer sweep.
24. **Battery W4/W5 rows** — same gate. **→ TODO P3 — Battery program waves W3-W5 (W4 ops & data, W5 realtime completions)** <!-- fresh:skip: routed in TODO_LIST -->
25. **cqrs EventConfig opt-ins** — demand-gated (idempotency recipe path documented).
26. **watchlist row 56 consolidation** — it has grown huge; next docs pass should split it into per-target rows or a table.
27. **Re-run `verify-all.sh` end-to-end on a quiet-load window** as the pre-tag gate for #3 (runner truth carried the wave; the pre-tag gate runs it canonically anyway).
28. **Announce/decide the SSE E2E sync rules** — sse-protocol guard green; nothing pending; skip unless drift.
29. **Re-check `golangci-lint-auto-configure` convergence after the next upstream tool bump** — the daemon rewrite of 11 configs (10-13 precedent) will recur; the configs-verify guard covers schema, but the POLICY deltas (noinlineerr stays disabled) need eyeballing.
30. **Consider making ci-preflight's expected-red state visible** — e.g. a workflow-level annotation or a badge note, so "failure" on the tip reads as designed without tribal knowledge (ties to #1: fixing the secret retires this entirely).
31. **Add the structure-linter install-auth diagnosis to its CI job comment** so the next person doesn't re-derive it.
32. **frh's `adapter.go` cooldown** — delete when fr#2 ships (dup of 13, keep one).
33. **`security_realtime_test.go` header** — it carries the package's only attached doc comment now; decide whether it stays the canonical package doc or gets detached like the others (cosmetic consistency).
34. **Battery: `DropCounter` W5-C1** — still parked on cqrs+realtime composition demand (zero consumers).
35. **`testkit` POST/query-param helper ergonomics** — if #9 lands, inventory once more before the train.
36. **core `svc.Routes()` seam** — battery B8 blocked on it; core has not agreed; keep parked.
37. **`httpx` B1 classification-parity test** — rides W3.
38. **Re-verify `example-smoke` covers the systemd example** — it boots all seven examples; green this run; nothing pending.
39. **`govulncheck` sweep on the next dependency train** (last clean 2026-10-13).
40. **dependabot grouped-config audit** — the 10:10 update run recreated PRs correctly post-wave; verify next week's run doesn't re-open closed supersessions (the auto-close behavior worked; watch for regression).
41. **AGENTS.md ≤220 cap** — currently 133; keep module detail in `docs/agents/module-details.md` as content grows.
42. **`docs/agents/module-details.md` §integration** — add the `documentedPins` requires-flightrecorder note (fixture now covers 11 appkit modules).
43. **M19's `waitForSnapshots` timeout** — 5s is fine under current load; if the shared box saturates harder, raise or quiet-run-gate.
44. **Consider `GOMODCACHE` isolation for integration `go get` operations** under load (the escalation path used in part 1's hermetic passes).
45. **TODO "Open questions" section** — machine-default go (#14) is the only true open question left; prune resolved ones at next pass.
46. **Cross-repo tracking posture** — the standing §g question (nudge cqrs-htmx vs report-only) still awaits owner intent.
47. **The four pma decisions** — cooldown + staged-respect implementations are the two actionable ones (checklist ready).
48. **Keep `art-dupl-gate` pin lockstep** — when re-baselining next, move the CI pin with it (documented trap).
49. **Annotate THIS report into the status index** during the next docs pass with its final verdicts.
50. **The v1.0.0 announcement surface** — when the anchor tag + overlap land, README/website get the v1 posture; parked until criteria close.

## g) QUESTIONS FOR LARS (cannot self-answer)

1. **The Actions `SSH_PRIVATE_KEY` secret is empty** (ci-preflight red since ~10:5xZ today). Will you restore it — and if its absence is actually intentional posture, should I demote ci-preflight to a permanently-expected red (badge note / job rename) so tip "failure" stops reading as an incident?
2. **crush-config `227a130` (the exact-token lesson) is committed locally but unpushed.** Push it now so home-manager deploys the lessons update, or is that repo on a different cadence?
3. **The additions-empty anchor tag (v0.9.1)** — criterion 1's freeze clock starts at the next core tag, and the tree is verified zero-diff vs v0.9.0. Do you want that tag cut on the next natural train (it would be the first additions-EMPTY release, pure doc/testkit staging), or should it wait for content so the "anchor" tag carries something users can point at?

_Arte in Aeternum — report ends; awaiting instructions._
