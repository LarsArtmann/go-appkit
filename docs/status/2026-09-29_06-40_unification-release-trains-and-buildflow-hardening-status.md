# Status — Unification + Release Trains + BuildFlow Hardening (session 2026-09-29, ~01:30–06:40 CEST)

**Predecessor:** `docs/status/2026-09-29_00-38_docs-health-full-pass.md` · **Plan executed:** `docs/planning/2026-09-29_00-45_SUPERB-v5-unification-trains-and-gated-unlocks.md` (user order: "execute the WHOLE todo list" — treated as the G1/G2 gate-opening word; genuinely external gates stayed closed).

**End state: tree clean @ `3b71f66`, all guards green (directives 11/11, pin-drift, parity, structure-linter 0, AGENTS cap OK), BuildFlow fast mode PASSED (from 42 failed steps at session start), five new tags on origin, every one proxy-verified.**

---

## a) FULLY DONE

| Task | Evidence |
|---|---|
| **T01 — unification train A**: 7 satellites → go 1.27.1 | `go mod edit -go=1.27.1` + tidy per module (cqrs, docs, realtime, errorpages, flightrecorder, otel, flightrecorderhealth); re-probed 11/11 |
| **T02 — train B**: go.work (already 1.27.1, 11 dirs) + AGENTS toolchain paragraph rewritten + `documentedGoDirective` verified (already 1.27.1) + all 3 guards green | `check-go-directives.sh` OK 11/11 |
| **Pin-drift guard bug fix**: `check-pin-drift.sh:88` inherited ambient `GOTOOLCHAIN=local` → every cross-repo proxy query died silently | self-pinned `GOTOOLCHAIN=go1.27.1` in the script |
| **T03 — 11-module matrix**: build+vet ✓, `test -race -count=1` ✓ ×11, sequential golangci-lint ✓ ×11, cqrs-lint ✓ | incl. root `hookcodes_test.go` dupl→table-driven refactor + err113 static sentinels; cqrs goimports named-import fix |
| **cqrs DLQ test fix (upstream semantic change)**: projectionhost v4.5.1 quarantines ONLY Rejection/Corruption to the DLQ; old fixture (plain error + fails-ALL-events) relied on a backoff timing window v4.5 closed | fixture now `errorfamily.NewRejection` + fails only the poison event; deterministic; suite green |
| **go-etag ambiguous-import resolution** (workspace-mode union: cqrs-htmx→`go-etag v0.4.0` still shipping `server/` vs httputil v1.4.0 importing split `go-etag/server`): three-part fix | ① integration `exclude` etag v0.4.0+v0.5.0 ② `go.work` replace `go-etag → v0.6.0` (committed) ③ core httputil v1.2.0→v1.4.0 (etag v0.6.0 at the source) |
| **T05 — core v0.6.0**: API diff vs v0.5.1 additions-only (testkit.Shutdown + godoc), CHANGELOG dated (+httputil bump line), annotated tag, push, **proxy PASS** | tag `v0.6.0` |
| **T06 — security v0.2.0**: godoc-only diff, same ritual, **proxy PASS** | tag `security/v0.2.0` |
| **T07 — health v0.1.3**: godoc-only diff, same ritual, **proxy PASS** | tag `health/v0.1.3` |
| **T08 — frh v0.1.4**: tagged, pushed, **proxy PASS** | tag `flightrecorderhealth/v0.1.4` |
| **frh v0.1.5 — directive-floor correction release**: v0.1.4's artifact shipped `go 1.27` (BuildFlow normalize downgrade AFTER my unification, BEFORE my tag); restored 1.27.1, honest CHANGELOG (both entries corrected), tagged+pushed, **proxy PASS** | commit `fa15f30`, tag `flightrecorderhealth/v0.1.5` |
| **T09 — integration re-pin**: core v0.6.0 / security v0.2.0 / health v0.1.3 / frh v0.1.4→v0.1.5 / ssetest v0.4.0 (guard caught the newer ssetest) + `ts.Shutdown` cleanup simplification; suite + lint green | `pin_drift_test.go` + `security_realtime_test.go` |
| **T10.2 — CI ssh steps gated off `pull_request`** (ssh-agent + private-URL git config) | `ci.yml` commit `1c4d9b9` |
| **T11 — go.work + go.work.sum tracked** (forced add past the nix-managed global ignore); fresh-checkout worktree: guard green | commit `1c4d9b9` |
| **T12 — guard hardening**: `check-go-directives.sh` rewritten (semver-normalized compare with BELOW/ABOVE-floor messages + `toolchain` directive coverage; negative-tested in a worktree — caught+fixed my own regex bug: `toolchain go1.27.1` carries the `go` prefix) | script + tests |
| **T12 — `.buildflow.yml` (new)**: `env.GOTOOLCHAIN: go1.27.1` + skips (`license-check` — nix-run go-licenses embeds pre-1.27 go, upstream fix; `go-structure-linter` — step lacks the repo's 3 exclude flags; `go-version-auto-configure` — fleet major.minor policy, the v0.1.4 downgrade culprit) | BuildFlow fast: **PASSED** (was 42 failures) |
| **T12 — pre-commit hook**: project-guards block (directives + dependabot-parity) appended after the BuildFlow block, with re-append warning | `.git/hooks/pre-commit` |
| **T14 — composition-contract pins**: go-health v0.2.0 / do v2.1.0 / go-flightrecorder v0.2.0 legs in `documentedPins` (version-locked to family pins, deliberately NOT proxy-LATEST — go-health v0.3.0 evaluation pending) | suite + lint green |
| **T04 — AGENTS GOEXPERIMENT retirement**: prefixes swapped to `GOTOOLCHAIN=go1.27.1` (the plan's "strip prefixes" inverted by reality: the machine default is still go1.26.7/GOTOOLCHAIN=local, so the prefix is REQUIRED — documented why) | AGENTS build-command blocks + gotchas (248/302/324/336) + health/doc.go |
| **Same-train docs**: AGENTS release state (through core v0.6.0), 4 module bullets, deps tables (httputil v1.4.0, go-sse v0.6.1, branded-id v0.6.0, full 12-row cqrs-lite sweep, error-family v0.11.0), TODO_LIST header + 2 completed rows removed; **4 new `[Unreleased]` CHANGELOG entries** for the BuildFlow sweep deltas (realtime, flightrecorder, docs, cqrs) | cap 377/377 (fixed a 378 overflow by merging the old BuildFlow gotcha) |
| 5 status-report-relevant pushes verified: origin/master == HEAD at every tag point | `git rev-parse` checks |

## b) PARTIALLY DONE

1. **T18 — cqrs v0.6.0 (`SQLitePath` removal)**: code + test + README-pending. `eventservice.go` field/branches removed, deprecated-alias test deleted, build+vet+`test -race` GREEN, lint pending re-run, **README line 49 edit not applied** (interrupted), **CHANGELOG Breaking section + v0.6.0 dating NOT written**, **API diff vs cqrs/v0.5.0 NOT run**, **NOT tagged**. All changes are committed (daemon swept) but the release train is ~60% done.
2. **BuildFlow full-mode validation**: only `--build-mode fast` verified; full mode (~5-10 min, includes test-race) never run this session.
3. **Consumer-side unblock verification**: core v0.6.0 unblocks cqrs-htmx's go-etag stub-replace drop — their move, not verified by us (their repo).

## c) NOT STARTED (un-gated plan tail)

T17 (cqrs lifecycle E2E through a live Service — highest-value uncovered module), T19 (otel bench re-baseline + `example/`/`health/example` live E2Es), T24 (typed `seriesKey` + named `Hook` type), T25 (errors.Join shape test + frame-ancestors asserts + SSE×health E2E), T26 (testkit drain-window helper + health-example `-hardened` mode), T29 (watchlist sweep), T30 (external-consumer ritual decision), T32 (rituals re-arm), T20 (AGENTS deep slim-down), T21 (art-dupl suppression audit), T31 (lessons.md entry in crush-config — needs a commit in that repo).

**STILL GATED (untouched, by design):** T13 (G3 consumer-pin policy), T15 (6 upstream filings), T16 (errorpages statusRecorder swap), T22 (pdg, their repo), T23 (prompt-crusher intent), T27 (govulncheck, networked machine), T28 (browser CSP, Chrome), T10.1 (owner SSH secret), battery waves (demand).

## d) TOTALLY FUCKED UP (honest list)

1. **frh v0.1.4 shipped with `go 1.27` instead of 1.27.1.** Root cause: I tagged from a tree that BuildFlow's normalize step had silently downgraded AFTER my green 11/11 probe, and I did NOT re-run `check-go-directives.sh` between the BuildFlow run and tagging — violating the plan's own Verschlimmbesserung guard #3 (frozen snapshot + guards before facts/tags). Caught ~2 hours later by the same guard, fixed with the v0.1.5 correction release. **Lesson that must become ritual: guards re-run immediately before every `git tag`, plus a `git show <tag>:<module>/go.mod` check before push.**
2. **Three-time repetition of the same failure class**: ambient `GOTOOLCHAIN=local` broke (a) check-pin-drift's proxy queries, (b) the golangci-lint sweep, (c) my own re-run where I forgot the export. I fixed each ad hoc instead of globalizing the fix after hit #1. Session-end state: fixed at all three layers (script self-pin, .buildflow.yml env, AGENTS docs), but it cost ~4 debug cycles.
3. **go-etag misdiagnosis #1**: after the load-graph test passed on the settled tree I declared the ambiguity dead; it reappeared under workspace mode in lint. I then tried `go get` (tidy stripped the require), then excludes alone (don't bind the workspace union) before landing the correct 3-part fix. Should have run `go mod graph` in WORKSPACE mode first — it immediately revealed published-tag go.mods feeding the union.
4. **Stale-evidence trust**: the first full build sweep (job 07B) straddled BuildFlow's mid-flight tree mutation; I briefly treated its all-OK as current-tree truth before redoing it.
5. **Small edit-tool hygiene lapses**: two multiedit batches failed partially on exact-match/whitespace (hookcodes_test var block near-duplicated for a moment; AGENTS/TODO/CHANGELOG edits each needed a view-first retry); the guard negative-tests first ran against HEAD's OLD script in the worktree (uncommitted new script) — a false-confidence moment. Also used python for a TODO_LIST row deletion instead of the edit tool.

## e) WHAT WE SHOULD IMPROVE

1. **Make guard-before-tag mechanical**: add a `scripts/pre-tag-checks.sh` (directives + pin-drift + parity + `git show <tag>:…` directive assert) and reference it in AGENTS Release Ritual step 4.5 — the v0.1.4 class dies only when the check is in the command sequence, not in memory.
2. **One env invariant everywhere**: every command chain in this repo needs `GOTOOLCHAIN=go1.27.1` until the machine default moves; consider a repo-local direnv/.envrc or documenting `go() { command GOTOOLCHAIN=go1.27.1 go "$@"; }` in AGENTS.
3. **Workspace-mode vs standalone-mode is now a REAL repo contract** (go.work replace + integration excludes + published-tag graph edges): deserves its own AGENTS gotcha with the `go mod graph` diagnostic recipe.
4. **BuildFlow interaction knowledge is now hard-won and written down** (caller-wins env, nix-run sandbox isolation, result-cache replay of failures — `BUILDFLOW_NO_RESULT_CACHE=1`) — keep the AGENTS gotcha current; the upstream asks (go-licenses ≥1.27 runner, structure-linter per-tool excludes) belong in `doc/feedback/outgoing/` drafts.
5. **Tag hygiene for 0.x patches carrying floor bumps**: the CHANGELOG must state the directive bump (done this train) AND the tagged artifact must be verified (new ritual step, see #1).

## f) NEXT — up to 50 (impact-ordered)

1. **Finish T18**: cqrs README line 49 edit → CHANGELOG `### Breaking` + `[0.6.0] - <date>` → API diff vs `cqrs/v0.5.0` (expect removals → breaking, migration note = rename `SQLitePath`→`DSN`) → lint+cqrs-lint → tag `cqrs/v0.6.0` → push → proxy check → AGENTS/TODO same-train.
2. Add **pre-tag-checks.sh** + Release Ritual step (post-mortem action, e1).
3. **T17** — cqrs lifecycle E2E through a live appkit Service (register/dispatch/query + in-flight drain) → 8/10 composition-proven.
4. **T19** — otel benchmark re-run (n≥5) vs README table on the 1.27.1 floor.
5. **T19b** — `example/` + `health/example` live E2E via Go prober.
6. **T24a** — typed `seriesKey` in metrics.go (kill the `|`-join/SplitN pair).
7. **T24b** — `type Hook func(context.Context) error` in config/service signatures.
8. **T25a** — errors.Join message-shape test (multi-hook failure text).
9. **T25b** — hardened-dashboard `frame-ancestors 'none'` + sorted-order asserts.
10. **T25c** — SSE×health lifecycle E2E (drain during an open stream).
11. **T26a** — testkit drain-window helper + migrate 2 hand-rolled tests.
12. **T26b** — health example `-hardened` mode + README note.
13. **realtime v0.1.2 train** (go-sse v0.6.1 + directive bump — [Unreleased] ready).
14. **flightrecorder v0.1.1 train** (httputil v0.11.0→v1.2.0 + directive — [Unreleased] ready).
15. **docs v0.3.1 train** (templ-components v1.19.4 + directive — [Unreleased] ready).
16. **BuildFlow full-mode run** (one clean validation incl. test-race).
17. **T29** — watchlist refresh sweep (cordis consumers, PapDashboard, nixpkgs toolchain, dprint exit-14, cqrs-htmx v5 window) with as-of dates.
18. **T30** — external-consumer proxy ritual: write the decision + recipe paragraph.
19. **T32** — rituals re-arm (cqrs cookbook re-verify trigger, core v1 exit-criteria graduation check, battery demand re-check).
20. **T20** — AGENTS deep slim-down batch 1 (extract per-module build commands → module READMEs; restores cap headroom).
21. **T20b/c** — slim batches 2–3 (gotcha groups).
22. **T21** — art-dupl suppression audit (~431 groups) + standing policy + Decision 13 cross-link.
23. **T31** — lessons.md entry in crush-config (commit there): "check GOTOOLCHAIN/go.work parity AND re-run guards between dependency-automation runs and tagging."
24. **Upstream draft: BuildFlow go-licenses ≥1.27 runner** (from the new license-check skip rationale).
25. **Upstream draft: go-structure-linter** — per-tool rule excludes / honor yaml config (from the new skip rationale).
26. Watch **cqrs-htmx**: they can drop their go-etag stub-replace now (core v0.6.0) — their repo, their move; keep the claim fresh (D2 ritual).
27. **T13** (G3-gated) — consumer-claim ritual script or strip pins from AGENTS.
28. **T15** (USER-gated) — 6 upstream filings (go-health pack could be direct fixes in our repo).
29. **T16** (USER-gated) — errorpages statusRecorder → httputil.ResponseRecorder swap.
30. **T22** (their repo) — pdg backlog sweep.
31. **T23** (user intent) — prompt-crusher merge resolution.
32. **T27** — govulncheck on health + security (networked machine).
33. **T28** — browser CSP pass over DashboardHardenedPreset (chromedp) + THREAT_MODEL link.
34. **T10.1** — owner sets `SSH_PRIVATE_KEY` secret → then T10.3 watch one fully green master run (incl. the now-committed go.work feeding the directives job).
35. Triage the **golangci-lint-config-verify warning** (`exhaustruct_v5.exclude` not allowed in one module's config — BuildFlow flagged; find which module, fix key or linter version).
36. Triage the **5 go-auto-upgrade findings** (lo.FromPtr in frh adapter; lo.Map/lo.Filter in otel/realtime tests) — fix or policy-skip with rationale.
37. **jscpd policy decision**: 13 findings — the cross-module `.golangci.yml` duplication (277-line clones) is by-design per-module configs; suppress in dprint/jscpd config with rationale or accept the noise.
38. Fix the **5 broken doc links** lychee found (doc/planning references to sibling paths that moved: `go-plugin-mvp`, `forks/cordis`).
39. Consider **dropping the go.work etag replace** once realtime/flightrecorder/docs/otel/security trains ship (their graphs then carry etag ≥ the split) — keep until then.
40. **SECURITY go.mod**: example-only appkit dep still at v0.5.1 — bump to v0.6.0 on the next security-touching train (BuildFlow may auto-bump; verify CHANGELOG then).
41. Verify **CI directives job** actually parses the committed go.work on the next push (guards+proxy-smoke should be visible even with the matrix owner-blocked).
42. **DOMAIN_LANGUAGE.md**: add the DLQ poison contract terms (Rejection=quarantined, retryable=restarts) — the v4.5 semantic change deserves vocabulary.
43. Clean **/tmp/core-old** + any scratch dirs from the release ritual.
44. **AGENTS integration bullet**: the pins prose still says "LATEST published of every family module" — the T14 version-locked legs add a nuance worth one clause.
45. Re-verify **dependabot config** picks up nothing weird from the 4 go.mod sweeps (grouped updates should reconcile cleanly).
46. **README.md (root)**: still says nothing wrong, but the release-state claim "through core v0.5.1" appears in FEATURES/ROADMAP? — grep-and-sync all living docs for the old floor claim.
47. **health module**: go-health v0.3.0 evaluation (tracked TODO) — the T14 lock note references it; schedule the eval or re-affirm the wait.
48. Consider **retiring `GOTOOLCHAIN=go1.27.1` from .buildflow.yml** once the machine default moves (paired with #2's direnv decision).
49. **post-release pkg.go.dev render check** for the 5 new tags (crawler lag ritual — NOT a failure if 404 today).
50. **Retrospective fix in SUPERB v5 doc**: mark T01–T12/T14 executed with deltas (frh v0.1.5 correction, etag fixes) so the plan mirrors reality for the next session.

## g) Questions only you can answer

1. **G3 (still open)**: consumer-pin claims in AGENTS/FEATURES — keep them and I build `scripts/check-consumer-claims.sh` (re-verifies claims after each cqrs-htmx release), or drop exact consumer pins from AGENTS entirely? T13 is blocked on this word.
2. **Train cadence for the sweep-carrying modules**: realtime/flightrecorder/docs each carry small unreleased deltas (dep bumps + directive) — ship them as individual trains NOW (3 more tag+push cycles) or batch them into the next natural train (e.g., together with cqrs v0.6.0 / the next feature wave)? Either is executable; it's a release-noise preference.
3. **Machine toolchain**: the local default go is 1.26.7 with `GOTOOLCHAIN=local` (nix-managed, read-only to me) — is upgrading the machine's default go to ≥1.27.1 on your agenda? Yes → I retire the entire GOTOOLCHAIN-prefix discipline (AGENTS blocks, .buildflow.yml env, hook notes) in one follow-up train. No → the prefix stays and I keep documenting it.
