# Session Status — CI Split + verify-all.sh + the go.sum Poisoning Incident

**2026-10-14 14:55 CEST · Executor: Crush · Scope: THIS session only (13:17→14:55, resuming the 2026-10-13 17-49 report's three owner decisions; tree clean at `b741aa6`, pushed).**

Owner decisions executed from the 17-49 report: (f2) push — found ALREADY satisfied on resume (master == origin; another session had landed the go-cqrs-lite catalog v4.8.0 train overnight); (f3) split the CI workflow — DONE; (f9) commit the verify loop as a script — DONE. The session's centerpiece was unplanned: **the split's very first unauthenticated CI run caught a real poisoned `integration/go.sum`** (a force-moved release tag + go's stale VCS cache), which consumed most of the session to diagnose and repair. Final CI run `38560059602` on the pushed tip: **30 jobs green — all 12 matrix slots unauthenticated for the first time since ~2026-10-01 — only the designed `ci-preflight` secret gate red.**

---

## a) FULLY DONE (verified)

1. **Resume-state verification** — confirmed master == origin at `cab31f3` on arrival; the two "failed" overnight CI runs (`06a85f9`, `56bc1da`) are designed-red only (ci-preflight secret gate; every real job green); `cab31f3` correctly skips CI (md-only, `paths-ignore`). The "push" decision was already satisfied before I touched anything.
2. **Dependency visibility audit (the split's fact base)** — `gh repo view` over all ten family repos: **every dependency is PUBLIC except `go-structure-linter` (PRIVATE)**. Cross-checked against proxy-smoke's anonymous resolution (the released modules' full transitive trees resolve unauthenticated on every push) and integration's require list (all within the public set).
3. **CI workflow SPLIT** (`.github/workflows/ci.yml`): `test` matrix decoupled from `ci-preflight` — runs unauthenticated on every push; the two SSH steps remain as a **secret-conditional fallback** (job-level `SSH_KEY_PRESENT: ${{ secrets.SSH_PRIVATE_KEY != '' }}`, steps gated on `env.SSH_KEY_PRESENT == 'true'`) so a future private dependency keeps working; `structure-linter` ALONE still rides preflight (its binary repo is the one private family repo — a secret-less run would be a vacuous green, the exact blind spot the guard kills); `ci-preflight` stays as the single loud secret-watchdog with its error message truth-synced (the old text claimed the matrix was skipped — now false).
4. **actionlint validation** — caught a real illegality in my first draft (`secrets` context is NOT allowed in step-level `if`; fixed via the job-env pattern). YAML parse + actionlint clean.
5. **Guards re-run green after the workflow change**: check-dependabot-parity (parses the CI matrix), check-workspace-charter, check-go-directives, check-structure-linter, check-golangci-configs, check-consumer-claims (after its own refresh, see a7), check-pin-drift.
6. **`scripts/verify-all.sh` shipped + verified** — the hand-rolled hermetic loop as one command: per-module build+vet+test `-race -count=1` (GOWORK=off, GOTOOLCHAIN derived from the root go.mod directive) then SEQUENTIAL per-module golangci-lint (integration with `GOWORK=off` per its README), per-module log files + explicit PASS/FAIL EXIT verdict lines, nonzero exit on any failure, `--skip-tests`/`--skip-lint` + module args. First full run: **12/12 tests + 12/12 lint ALL GREEN (179s)**; final run on the repaired tree: **ALL GREEN (49s)**.
7. **Consumer-claims refresh** — the claims guard caught upstream releasing `setup/v4.14.4` after the AGENTS line was written. Verified the new tag at source (raw `setup/go.mod`): pins `go-appkit v0.9.0`, carries the split go-etag/entitytag+server v0.6.1 — pure tag bump; AGENTS + FEATURES refreshed; guard green (9 OK lines incl. the PapDashboard leg).
8. **HARVEST of the 17-49 report** (f5): f1/f2/f3/f5/f9 annotated inline (`→ RESOLVED/DECIDED + DONE 2026-10-14`); f4/f7/f10/f12 routed into TODO_LIST P2/P3; **f11 fixed on sight** — the two missing index rows (15-54, 17-02) added to `docs/status/README.md`, descriptions verified against the actual report headers; f8 folded into AGENTS Gotchas as a standing rule.
9. **TODO_LIST truth-syncs**: P1 rewritten (was "restore secret to unskip the matrix" — now narrowed: the matrix runs unconditionally; the secret gates exactly one leg, `structure-linter`; acceptance = that job RUNS green); "Updated:" header prepended with the session summary; docs module version synced v0.3.2→v0.3.3 (the overnight catalog train).
10. **AGENTS.md truth-syncs**: the CI bullet rewritten (unauthenticated matrix, public-deps proof, structure-linter-private-binary rationale); verify-all.sh added to Build & Verify; NEW Gotcha bullet: integration commands verbatim from its README (f8); NEW Gotcha bullet: **"Stale go VCS cache after a re-tag"** with the full repair recipe (see a11).
11. **THE go.sum POISONING INCIDENT — diagnosed, repaired, prevented-by-documentation.** The split's first live run (`38557606704` on `9d86784`) ran the matrix unauthenticated and `test (./integration)` FAILED at vet: `SECURITY ERROR checksum mismatch` on `go-appkit/health@v0.2.0` (downloaded `zuEq…` vs go.sum `VDI6…`). Root-cause chain, proven at every step: (i) sum.golang.org canonical hash = `zuEq…` — CI's fresh download was RIGHT, go.sum was WRONG; (ii) fresh-GOMODCACHE replication over all 12 modules: only integration poisoned; (iii) byte-comparison of the two zips (sha `00bb1d…` vs `d5fbfc…`) + content diff: the CHANGELOG text differs ("go-health v0.5.0" vs "v0.5.1" + a sweep note) → **the health v0.2.0 tag was force-moved after release**; (iv) go's module cache keeps a bare clone per repo under `cache/vcs/` and updates it with plain `git fetch` — a force-moved tag never updates there (non-fast-forward), so local downloads kept building the PRE-move zip while proxy/CI served post-move bytes; (v) the shared `/mnt/buildcache/go-mod` cache hid the mismatch locally (cached zip verifies against the cached-poisoned go.sum). Repair: strip + re-tidy against the canonical proxy with a fresh GOMODCACHE (**1-line go.sum diff**), purge the stale `cache/vcs` clone (its tag sat at `72c0fc96` vs origin `3b2c58d5`) and the `cache/download/@v` entries (chmod u+w — the shared cache was never nuked), suite green in 7.3s, shared-cache zip re-verified byte-identical to canonical.
12. **Push + close-out ritual, twice**: push 1 (`9d86784`: split + script + docs) → CI exposed the go.sum landmine; push 2 (`b741aa6`: repair + gotcha docs) → CI run `38560059602`: **30 jobs green, ALL 12 matrix slots green (integration 2m18s), only ci-preflight red by design, structure-linter skipped by design.** The matrix is no longer dark — the 2026-10-01..07 blind-spot class is structurally dead.

## b) PARTIALLY DONE

1. **go.sum canonical-checksum guard** — fully diagnosed with a documented repair recipe and a P2 row, but the guard itself (a `--fresh-cache` mode in verify-all.sh, a pre-tag-checks step, or a dedicated CI job) is NOT built. This is the one mechanical fix that would have caught the incident BEFORE push.
2. **P1: the SSH secret** — narrowed to the `structure-linter` leg only (owner-only fix; repo secrets are invisible to agents). Until restored, that one CI leg stays skipped (honestly dark, loudly gated — not vacuously green).
3. **Repo-wide VCS-cache staleness** — integration's go.sum is repaired, but any OTHER force-moved tag in the family (cqrs-htmx, go-sse, …) would poison OTHER boxes' caches the same way; this box's remaining `cache/vcs` entries were not swept against origin (only go-appkit's was).
4. **TODO_LIST "Updated:" header** — accurate but now a ~700-word paragraph; the rolling-summary convention is straining (see e8).
5. **Session-close ritual, partially inherited** — CI verified on the final pushed commit (done, a12), but the ritual's other half ("confirm every NEW workflow job had a live run") is N/A this session — no new jobs were added, only rewired; the matrix's first live unauthenticated run IS the new-job live-run proof.

## c) NOT STARTED (carried; untouched this session by scope discipline)

1. Owner: restore `SSH_PRIVATE_KEY` (P1 — the structure-linter leg).
2. `docs/agents/` directory charter (one file today; write before a second appears).
3. Quiet-gate helper + quiet-window re-run of the full sweep and cqrs `-count=10` (routed from 17-49 f4).
4. SSE-protocol guard: check that test-file `hub.Broadcast(` calls have a `SubscriberCount()` sync nearby (17-49 f7).
5. Fresh `art-dupl -t 1` sweep over the lo-adoption + SSE-test churn (17-49 f10).
6. Realtime subscribe-before-flush (17-49 f12; demand-gated on a no-drop-on-connect consumer).
7. All standing P2/P3 rows: health `WithHealthRecorder` cliff upstream watch (go-health#7), `Service` composition on `httputil.Server` (BLOCKED on httputil#41), go-cqrs-lite#67 GracefulClose watch, go-flightrecorder#2 cooldown watch, go.work leak #1-7 root-cause residual, pma §6 owner decisions, battery waves W3-W5, i18n re-key (next httputil feature tag), otel bench re-baseline (needs the quiet gate), core v1 exit criteria (draft, needs consumer growth), consumer-claim drift ritual posture (USER-GATED), cordis bridge (2 of 3 triggers pending), PapDashboard reverse adoption (their repo), go-aichat/ssehub convergence watch, composition-gap E2E watchlist, go-idempotency adopt-wire decision, BuildFlow#46 watch, dprint exit-14 upstream filing.
8. Govulncheck — NOT run this session (rides dependency trains; this session tagged no releases — correct by policy, noted for completeness).

## d) TOTALLY FUCKED UP (honest)

1. **I violated the repo's own pipe-gotcha TWICE in one session.** `echo "TIDY_EXIT=$?"` and `echo "TEST_EXIT=$?"` after pipes printed the pipe's exit code (`0`) while the underlying commands had FAILED — the exact "failed but exit=0" ghost that AGENTS.md bans with three documented incidents. Both times output inspection caught the lie, but I doc­umented that rule for others and then repeated it myself within the hour. The rule needs to be mechanical for me too: capture `$?` into a variable BEFORE any pipe, or write to a file and read it.
2. **verify-all.sh was declared "done" with a semantic bug**: I syntax-checked (`bash -n`) but never smoke-tested; the first real run failed 12/12 (`GOTOOLCHAIN="1.27.2"` — invalid, needs the `go` prefix). The verdict-line design made the failure loud and instant (that part worked), but "syntax OK" is not "works". A one-module smoke run before the fleet run was the missing step.
3. **My first CI-replication fleet-check was BLIND — and I nearly trusted it.** Verifying all 12 go.sums against the proxy WITH the shared module cache produced a green that proved nothing: cached zips verify against the cached-poisoned go.sums. I caught the blindness (the CI failure was standing evidence against the green), but the first version of the check was a false all-clear that a less suspicious session would have shipped.
4. **Claimed platform behavior from memory, wrong**: `secrets.*` in step-level `if` is illegal in GitHub Actions (context availability). actionlint caught it pre-push — but the claim came from the same place unverified guesses come from. Platform APIs deserve the verify-before-claiming gate too.
5. **Fetched the wrong file for the setup/v4.14.4 verification**: `raw.githubusercontent/<ref>/go.mod` is the repo-ROOT module (which has NO go-appkit require at all); the claims verification needs `setup/go.mod`. Self-caught via the missing require line and re-fetched — but the first fetch, if unexamined, would have "verified" nothing.
6. **Three wasted edit round-trips on stale-read rejections** (AGENTS.md, FEATURES.md — the daemon had touched them after my context loaded them). The mod-time error is the signal to `view` first, not to retry blind.
7. **No pre-push fresh-cache check existed, and I didn't build one before pushing.** The go.sum poison was locally invisible by construction; a cheap fresh-cache verification before push 1 would have caught integration's bad hash without burning a CI run. The guard is now P2 — it should have been built the moment the first checksum error appeared, not routed.

## e) WHAT WE SHOULD IMPROVE

1. **Build the fresh-cache go.sum guard** (f2/f3 below) — `verify-all.sh --fresh-cache` and/or a CI job; the matrix catching it this time was luck plus a workflow change, not a guarantee.
2. **Make exit-code capture mechanical**: never `echo $?` after a pipe — assign first, or redirect to a file. Consider a pre-commit shellcheck-style nudge on scripts/, and hold agent output to the same standard.
3. **Smoke-test scripts on ONE module before declaring fleet-ready** (`./verify-all.sh cqrs` first, then the fleet).
4. **Reconsider the fleet-wide GOPRIVATE convention** (owner decision): every dependency is public; GOPRIVATE buys fresh-tag VCS resolution for dev at the cost of sumdb bypass and proxy CDN — and the sumdb bypass is exactly what let the poisoned go.sum go unnoticed locally. Narrowing it (or dropping it in CI) makes canonical go.sums true by construction.
5. **Codify the tag-move question**: force-moved release tags silently poison every existing local `cache/vcs` clone fleet-wide. Either ban re-tags post-push (release v0.2.1 instead) or write the re-tag ritual (purge caches everywhere + re-record go.sums + CHANGELOG note). The AGENTS gotcha documents the repair; the policy is still open.
6. **View-before-edit after ANY daemon commit** — the mod-time rejection is avoidable friction; the daemon commits continuously in this repo.
7. **Report-index completeness belongs in the report checklist**: the f11 omission survived TWO sessions before being fixed today; "every report has an index row" should be checked at write time, not discovered later.
8. **TODO_LIST header is becoming a ledger** — consider moving the rolling per-session summaries to a linked history file and keeping the header to ~5 lines + a pointer.
9. **Keep verify-all.sh's lint sequential** (concurrent golangci-lint races the shared build cache — a documented under-reporting mode); resist the temptation to parallelize.

## f) Up to 50 things to get done next (ranked; `[NEW]` = born this session, `[CARRY]` = already routed in TODO_LIST)

**Tier 1 — signal integrity (small, high impact)**

| #   | Task                                                                                                                                                                                  | Impact   | Effort    | Source    |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- | --------- |
| f1  | Owner: restore `SSH_PRIVATE_KEY` secret — acceptance: `structure-linter` job RUNS (not skipped) and green                                                                             | Critical | S (owner) | CARRY P1  |
| f2  | `[NEW]` Build `verify-all.sh --fresh-cache`: isolated GOMODCACHE + `GOPRIVATE=` + proxy replication of CI's go.sum verification (would have caught the incident pre-push)             | High     | S         | NEW       |
| f3  | `[NEW]` CI: go.sum-canonical guard job (or fold into an existing guard job) so the class never depends on matrix luck again                                                           | High     | S         | NEW       |
| f4  | `[NEW]` Owner ruling: force-moved tags — ban post-push re-tags (release `.1` instead) or codify the re-tag ritual (purge `cache/vcs` fleet-wide + re-record go.sums + CHANGELOG note) | High     | S (owner) | NEW       |
| f5  | `[NEW]` Owner ruling: narrow/drop fleet GOPRIVATE now that every dep is public — sumdb verification everywhere makes canonical go.sums structural                                     | High     | S (owner) | NEW       |
| f6  | `[NEW]` Sweep this box's remaining `cache/vcs` entries against origin tags (cqrs-htmx, go-sse, …) — other silently-stale clones from the same force-move class                        | Medium   | S         | NEW       |
| f7  | `[CARRY]` Quiet-gate helper (CPU-pinned/taskset protocol) + quiet-window re-run of the sweep and cqrs `-count=10`                                                                     | Medium   | S         | 17-49 f4  |
| f8  | `[CARRY]` SSE-protocol guard for test-file broadcasts (SubscriberCount sync check)                                                                                                    | Medium   | S–M       | 17-49 f7  |
| f9  | `[CARRY]` `docs/agents/` directory charter before the directory grows a second file                                                                                                   | Low      | S         | carried   |
| f10 | `[CARRY]` Fresh `art-dupl -t 1` over lo-adoption + SSE-test churn                                                                                                                     | Low      | S         | 17-49 f10 |

**Tier 2 — releases, consumers, upstream (watch + act on triggers)**

| #   | Task                                                                                                                                                         | Impact | Effort | Source |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ------ |
| f11 | After the next cqrs feature/dep train: release cqrs with the pending `[Unreleased]` (lo runtime site + deps) via the release ritual                          | Medium | M      | CARRY  |
| f12 | Same-train decision for otel v0.3.0 (LoggingFromContext wiring when httputil ships it) — the natural carrier                                                 | Medium | S      | CARRY  |
| f13 | pkg.go.dev re-check (core@v0.9.0, errorpages@v0.1.2 render-checks were <36h old; fr page 48h window closes ~2026-10-15) — ghost re-tag ritual is last resort | Low    | S      | CARRY  |
| f14 | After each cqrs-htmx release: the consumer-claim drift ritual (guard now catches tag drift automatically — posture ruling still USER-GATED)                  | Low    | S      | CARRY  |
| f15 | go-cqrs-lite#67 lands → simplify `EventService.Shutdown` to bare `GracefulClose`                                                                             | Medium | S      | CARRY  |
| f16 | go-flightrecorder#2 lands → delete frh's hand-rolled cooldown                                                                                                | Low    | S      | CARRY  |
| f17 | httputil#41 (`NewServerListener`) lands → re-open the `Service` composition spike + Core TLS (G1)                                                            | High   | L      | CARRY  |
| f18 | go-health#7 lands → close the health recorder-cliff row                                                                                                      | Low    | S      | CARRY  |
| f19 | BuildFlow#46 lands → drop the `.buildflow.yml` `go-work-sync` skip_steps entry                                                                               | Low    | S      | CARRY  |
| f20 | go-structure-linter binary refresh lands upstream → re-verify the structure-linter guard + config-verify against the new binary                              | Low    | S      | CARRY  |

**Tier 3 — hardening born from the incident class**

| #   | Task                                                                                                                                                                   | Impact | Effort    | Source |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | --------- | ------ |
| f21 | `[NEW]` AGENTS: add verify-all.sh to the Release Ritual §2 hermetic-verify step (the script now exists; the ritual still says "run the loop by hand")                  | Low    | S         | NEW    |
| f22 | `[NEW]` Consider a `.github/workflows` md-only-commit fast-path note: cab31f3-class tips legitimately skip CI — document beside paths-ignore so nobody re-investigates | Low    | S         | NEW    |
| f23 | `[NEW]` verify-all.sh: add `--fresh-cache` reuse of one shared fresh cache dir (toolchain downloads once)                                                              | Low    | S         | NEW    |
| f24 | `[NEW]` scripts/: shellcheck pass over all guard scripts (BuildFlow's shellcheck step deadlocks locally — hand-run once)                                               | Low    | S         | NEW    |
| f25 | `[CARRY]` go.work leak #1-7 root-cause residual: inotify hunt or accept the cadence with `check-workspace-charter.sh --fix` (owner options)                            | Medium | M         | CARRY  |
| f26 | `[CARRY]` pma dossier §6 four owner decisions (recommendations exist in the memo)                                                                                      | Medium | S (owner) | CARRY  |
| f27 | `[CARRY]` Consumer-claim posture ruling: keep citing exact consumer pins in AGENTS vs stop (guard now catches tag drift; claims-about-consumers remain manual)         | Low    | S (owner) | CARRY  |
| f28 | `[CARRY]` dprint exit-14: file upstream AFTER the buildflow binary refresh (verify-before-filing against the current binary)                                           | Low    | S         | CARRY  |
| f29 | `[CARRY]` samber/do lazy-healthy note + exhaustruct_v5 schema ask at golangci-lint (external, each needs its own verify pass)                                          | Low    | S each    | CARRY  |

**Tier 4 — product/backlog (demand-gated, ROADMAP fuel)**

| #   | Task                                                                                                                                                    | Impact | Effort | Source   |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| f30 | Battery wave W3 `httpx` module (B1 ResultHandler family; error mapping MUST share errorpages' taxonomy, classification-parity test)                     | High   | L      | CARRY    |
| f31 | Battery W5-C2 projection→broadcast folded contract (must-have for every cqrs+realtime consumer)                                                         | High   | L      | CARRY    |
| f32 | Battery W5-C1 realtime drop/backpressure counters (borrow ssehub `Drops()` shape; dedupe with the convergence watch)                                    | Medium | M      | CARRY    |
| f33 | Battery W3 remainder: B2 bind+validation, B9 no-leak errors, B3 ResponseWriter contract, B4 conditional GET, B5 content negotiation, B7 write deadlines | Medium | L      | CARRY    |
| f34 | Battery W4: `worker` supervisor+pool, `sqlite` ops kit, `polite` outbound client, `do` bridge, D6 webhook, F1/F4 config                                 | Medium | L      | CARRY    |
| f35 | D7 idempotency: adopt-wire go-idempotency vs CV port (audit says in-family coverage exists — decide)                                                    | Medium | S      | CARRY    |
| f36 | B6/B7 route introspection + metadata feed — gated on a core `svc.Routes()` seam core has not agreed to                                                  | Medium | M      | CARRY    |
| f37 | Core TLS option — gated on PapDashboard `PAP_TLS_CERT/KEY` becoming an appkit-hosting requirement (and on f17's seam)                                   | Medium | M      | CARRY    |
| f38 | PapDashboard reverse adoption (their repo hosts on appkit) — watchlist row, their move                                                                  | Low    | –      | CARRY    |
| f39 | cordis bridge — remaining triggers: a consumer states the reactive-composition requirement AND core v1.0.0 criteria ship                                | Low    | M      | CARRY    |
| f40 | go-aichat/ssehub convergence watch (owner-level decision if both SSE wrappers keep growing)                                                             | Low    | S      | CARRY    |
| f41 | Composition-gap E2E watchlist (otel×realtime, security×errorpages, errorpages×otel, fr-through-Service, SSE+health drain) — pick up on module touch     | Low    | M      | CARRY    |
| f42 | otel ns/op re-baseline in a genuinely quiet window (blocked on f7's quiet gate)                                                                         | Low    | S      | CARRY    |
| f43 | Core v1.0.0 exit criteria — graduate from draft when the consumer count grows                                                                           | Medium | M      | CARRY    |
| f44 | health recorder-cliff: revisit only if consumers still hit it post-v0.1.2 (parked decision stands)                                                      | Low    | S      | CARRY    |
| f45 | i18n content lane stays parked (appkit consumers N=1); re-key fires on the next httputil feature tag                                                    | Low    | S      | CARRY    |
| f46 | integration: subscribe-before-flush handler change — ONLY on a no-drop-on-connect consumer demand                                                       | Medium | M      | CARRY    |
| f47 | `[NEW]` Realtime README limitation note: cross-reference the new integration test sync pattern as the interim protocol for test authors                 | Low    | S      | NEW      |
| f48 | `[NEW]` TODO_LIST header slim-down (move rolling summaries to a history file; header keeps ~5 lines + pointer)                                          | Low    | S      | NEW (e8) |
| f49 | `[NEW]` Report-writing checklist addition: "every new report gets an index row at write time" (the f11 class died today — keep it dead)                 | Low    | S      | NEW (e7) |
| f50 | `[NEW]` After f2 lands: negative-test the fresh-cache check (plant a poisoned hash in a scratch copy; guard must go red)                                | Medium | S      | NEW      |

Routing note: f1–f6 are the actionable core (two owner rulings + three small builds); most of Tier 4 is ROADMAP fuel, already correctly entombed in TODO_LIST P3 — do not double-track.

## g) Questions I can NOT figure out myself

1. **The last red leg — secret or visibility?** The only remaining CI red is the missing `SSH_PRIVATE_KEY`, and it gates exactly one job (`structure-linter`, whose binary repo is the one private family repo). Do you want to (a) restore the repo secret, or (b) make `go-structure-linter` public like the rest of the family — which kills the secret need entirely but exposes a tool you may deliberately keep private? (a) and (b) are mutually exclusive paths to the same "zero designed reds" end state.
2. **Was the health v0.2.0 force-moved tag intentional?** The zip bytes prove the tag was re-pointed after release (a CHANGELOG "v0.5.0"→"v0.5.1" correction rode into the tag). If that was you: should post-push re-tags be banned outright (ship corrections as v0.2.1) or get a codified ritual? If it was NOT you, that is a supply-chain conversation, not a hygiene one — and every existing local clone on every machine that cached the pre-move tag is still serving stale bytes for that version.
3. **GOPRIVATE: keep, narrow, or drop?** Every `larsartmann/*` dependency repo is public. GOPRIVATE currently buys fresh-tag VCS resolution during dev trains, at the cost of bypassing the checksum database — which is precisely why a poisoned go.sum could sit unnoticed in this repo, and why the repair needed a fresh-cache tidy. Narrow it to the one private repo, drop it in CI, or keep it fleet-wide as-is?

---

**Evidence anchors**: CI run `38560059602` (30 green / 1 designed red / 1 designed skip) on `b741aa6`; CI run `38557606704` (the catch) on `9d86784`; sumdb lookup `github.com/larsartmann/go-appkit/health@v0.2.0` = `h1:zuEqOadX559Dxzmd3So/D04QXPh+kpxiskvCVmRpLg4=`; poisoned entry was `h1:VDI6+hJsjkAZZculWoNyylMKvH6sv4gBYcYNZ3KDwp8=`; stale vcs-cache tag commit `72c0fc96` vs origin `3b2c58d5`; verify-all first full run 179s ALL GREEN, final 49s ALL GREEN; logs preserved under `/tmp/go-appkit-verify-*` and `/tmp/integration-test*.log`.
