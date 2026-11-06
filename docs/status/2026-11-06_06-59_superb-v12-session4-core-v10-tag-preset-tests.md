# Status: SUPERB v12 execution — session 4 (core v0.10.0 tagged, preset P2–P7 tests written, one shutdown-hang bug open)

**Created:** 2026-11-06 06:59 · **Scope:** resumed under the owner's "GET
SHIT DONE — THE WHOLE TODO LIST" directive (read the session-3 report first
as instructed; pushes still held — the directive is a work instruction, not
a push authorization). This session: created the local **v0.10.0** tag,
rode the integration re-pin onto it, closed both pre-tag-checks failures
(one structurally red-by-hold), re-pinned preset to v0.10.0, found and
fixed a REAL route-collision bug in the preset skeleton, added lifecycle
hooks passthrough, and wrote the ENTIRE P2–P7 test suite (8 files, 29
tests) + the ≤40-LOC example. One bug is OPEN: an intermittent 5s shutdown
hang in full-suite runs (isolated to a 3-test interaction, ~1/3
reproducible — details in b).

---

## a) FULLY DONE

1. **Core v0.10.0 tag CREATED (local, not pushed):** annotated tag with the
   semantic-delta message (RegisterClassifications, Server-Timing outermost,
   testkit FreeAddr/WaitForRunning/HTTPGetURL, art-dupl-gate CI job).
   Executed under the blanket execute-everything directive; the staged train
   (API diff additions-only, hermetic build/vet/test -race green, CHANGELOG
   dated, AGENTS/FEATURES synced in session 3) was fully verified before
   tagging.
2. **pre-tag-checks run 1 → 2 real failures, both fixed:**
   - `preset` missing from the pin-guard module map → added to
     `check-pin-drift.sh` tag_pattern (tagless module — the reverse check
     correctly tolerates no-tags-yet).
   - `integration pins core v0.9.0 but latest local tag is v0.10.0` →
     rode the train: bumped integration's core pin to v0.10.0 via the
     insteadOf VCS tidy trick (canonical-zip recipe), updated the
     documentedPins fixture. **Integration suite GREEN (7.767s, -race).**
3. **pre-tag-checks run 2 → green EXCEPT one structurally-held leg:** the
   only remaining FAIL is gosum-canonical's integration leg — it resolves
   deps through an isolated cache against proxy+sumdb, which cannot see an
   unpushed tag. Red-by-hold, exactly as predicted; self-heals the moment
   v0.10.0 reaches origin (the insteadOf tidy produced the canonical-zip
   hash, the proven cqrs-v0.7.0 recipe). All other legs (10 guards) green.
4. **preset re-pinned to core v0.10.0:** go.mod now requires v0.10.0;
   `RegisterClassifications` resolves; build+vet GREEN. The global
   git-config insteadOf rewrite is STILL SET (needed for further local-tag
   tidies; must be unset before any origin-truth operation).
5. **REAL BUG found and fixed in preset.go — RegisterHealth inversion:**
   the skeleton never set `RegisterHealth`, which would have collided
   core's `GET /health` with the dashboard's `/health` (both registrable —
   ServeMux treats method-pattern and bare pattern as distinct — but core's
   more-specific pattern SHADOWS the dashboard). My first fix had the
   logic INVERTED (`!cfg.NoDashboard`); the test suite caught it on the
   next run. Correct: `registerHealth := cfg.NoDashboard` — dashboard owns
   /health when mounted; core's family registers only when NoDashboard.
   Caught by TestNew_ZeroAppCodeBoots + TestNew_NoDashboard, never by
   compilation — the exact class of bug the P2 suite exists to catch.
6. **preset.Config gained lifecycle hooks passthrough** (design-note
   amendment): StartHooks/DrainHooks/ShutdownHooks []appkit.Hook. Without
   them a preset consumer could not register ANY lifecycle hook (systemd
   - spine wiring is internal). Consumer hooks run FIRST in every phase;
     the spine closures run after, so the otel provider flush lands LAST and
     spans cover app shutdown work. Pinned by TestNew_HookWiring
     ([start drain shutdown] order + health-check-run proof of
     Mounted.Start).
7. **P2–P7 test suite WRITTEN (8 files, 29 tests; current pass state in
   verification block):**
   - `spine_test.go`: zero-app-code boot (healthz/readyz/startupz 200 +
     dashboard /health 200 HTML + graceful stop), hook wiring + order,
     version passthrough (/version), NoDashboard coexistence (core family
     - probe family, no collision).
   - `errorpages_test.go`: pretty 404 default (title "Page not found"),
     JSON by Accept ("not_found" code), code-off → stdlib 404, env-off →
     stdlib 404, consumer route wins over the catch-all.
   - `metrics_test.go`: nil stays nil (404), auth-enforced passthrough
     (401 without / 200 + exposition with basic auth; preset never sets
     creds).
   - `otel_test.go`: code-on records spans (injected recording
     SpanExporter + provider Shutdown wired — exporter sees spans AND is
     shut down), code-off + env-on → ZERO spans (code wins, both
     directions, no global state touched), env-on boots (default
     no-exporter Setup), matrix invalid covered in conflict suite.
   - `security_test.go`: none adds nothing (no HSTS, no 413), standard =
     HSTS + nosniff + 413 oversize + 64 KiB passes, strict OriginCheck
     403 cross-origin / 200 allowlisted, strict-without-origins →
     Rejection naming Origins, env standard.
   - `realtime_test.go`: /events streams (sync on SubscriberCount — the
     SSE house rule — then broadcast + ssetest.MustReadNEvents), Hub
     out-field assigned, custom Pattern replaces default, off-by-default
     404, env invalid rejects.
   - `conflict_test.go`: the FULL design-note §5 matrix — 4 switches × 6
     rows (absent-default, env-on, env-off, code-on+env-off ACTIVE,
     code-off+env-on INACTIVE, env-invalid Rejection), table-driven with
     per-leg observables (pretty/plain 404, HSTS present/absent, /events
     mounted/absent + Hub out-field, otel boot/rejection).
   - `env_test.go`: resolveLegs unit tests — defaults, code-wins both
     directions, env-decides-when-nil, empty profile → standard, invalid
     values reject as Rejection naming the var, case-sensitive no-trim.
8. **Test-suite design decision: NO t.Parallel anywhere in preset tests.**
   The four APPKIT_* vars are ambient input to every New/resolveLegs call;
   parallel boot tests race the t.Setenv tests (Setenv forbids parallel on
   the SAME test but does not fence siblings). paralleltest is
   lint-excluded for _test.go in this module, so no lint fallout.
9. **Example app written:** `preset/example/main.go`, 35 LOC — boot,
   one health check, one route, Run.
10. **Test-dep requires added via tidy (not hand-edited):**
    `go-sse/ssetest v0.4.0` (matches realtime's own pin; v0.5.0 would drag
    the held go-sse v0.7.0/templ break) and `otel/sdk v1.47.0` (recording
    exporter).

## b) PARTIALLY DONE

1. **OPEN BUG — intermittent 5s shutdown hang (listener_close deadline
   exceeded), not yet root-caused.** Facts: full-suite runs hit it ~once
   per run on varying tests (ZeroAppCodeBoots, NoDashboard); passes in
   isolation; otel+spine / realtime+spine / conflict+spine all pass;
   **errorpages+security+spine reproduces ~1/3 iterations (-count=3 → 1
   failure)**. All requests complete before shutdown (request logs show
   200s, durations in µs-ms); server.Shutdown then waits the full 5s
   testkit stopTimeout for an "active" connection that never drains.
   Suspects (untested): default-client keep-alive cross-test state,
   errorpages/security client-per-request conns lingering, load-scheduling
   interaction. NEXT: goroutine dump during the 5s window (SIGQUIT to the
   test binary mid-hang, or a pprof dump hook) — the dump names the stuck
   conn/reader immediately.
2. **preset lint not yet run** (golangci-lint from the module dir; test
   files were written against the module's noctx/nlreturn/gci/errcheck
   rules — testkit.HTTPGetContext used instead of http.Get, deferred
   Close via `func() { _ = ... }()` — but the verdict belongs to the real
   binary).
3. **README snippet compile check + README truth pass** (grammar table now
   includes hooks passthrough; quick-start must match the final Config
   shape) — pending.

## c) NOT STARTED (unchanged from session 3, minus P2–P7 test-writing)

P7 lint+snippet+example-smoke wiring; P9 preset release train +
release-state sync; consumer heads-up drafts; R1b–e; S-train; G-CNT;
A1–A5; D1; D3; C2R; R2–R5a; H1–H7; X1–X4; Y/Z; OG memos; final
verify-all; final release-state sync; push-blocked legs (P8, proxy
checks); /tmp cleanup (plus NEW scratch: /tmp/ptc-v0.10.0*.log,
/tmp/preset-test-_.log, /tmp/z_.log, /tmp/zb*.log, /tmp/p404.log).

## d) TOTALLY FUCKED UP (own mistakes, all caught forward)

1. **I read only `tail -30` of the first full-suite run and classified ONE
   failure when there were FIVE** — the EXACT documented no-pipe-truncation
   trap (AGENTS: "Never pipe verify/guard output"; three prior incidents).
   The subsequent "one fix → rerun → new failure" loop cost three
   extra test cycles. Rule reinforced: rg the log, never tail it.
2. **First-draft garbage that diagnostics then caught:** a nonsensical
   `retries()` helper in spine_test.go; dead `echoHandler/serviceMux`
   scaffolding in security_test.go; an `httpWriter` placeholder in the
   example. All removed — but none should have been written.
3. **Wrote `resp, err :=` against single-return helpers** (testkit
   .HTTPGetContext fatals internally and returns only the response) —
   checked battery signatures before writing but skipped the testkit
   helpers. Fix cost a round trip the typecheck predicted.
4. **The RegisterHealth fix itself shipped inverted first** (`!cfg.No
   Dashboard`) — logic written from memory of the doc comment instead of
   re-deriving the truth table. The NoDashboard test caught it in one
   cycle. Same lesson as ever: derive, don't recall.
5. **Session-3's skeleton had a latent boot PANIC risk** (dashboard /health
   vs core /health — the shadowing analysis in a5): the skeleton was
   "compile-verified" but never booted. "Compiles" is not "works" — P2's
   whole reason to exist.

## e) WHAT WE SHOULD IMPROVE

1. **Signature checks belong in a pre-writing pass for ALL APIs a test
   touches** — batteries AND testkit helpers. One `go doc` sweep per file
   would have saved two round trips.
2. **The 5s-hang class needs a standard reproduction recipe** (goroutine
   dump mid-window) documented in AGENTS testing notes once root-caused —
   it will recur for any future SSE/keep-alive-heavy suite.
3. **`assert` on library-generated HTML should target stable TEXT markers,
   not document structure** (doctype/case differs between the dashboard
   full-document and the errorpage fragment). "Page not found" beats
   "<!DOCTYPE html>".

## f) NEXT — in execution order

1. Root-cause the 5s shutdown hang (goroutine dump during the window;
   fix; pin with a regression test if preset-side)
2. Full preset suite green ×3 consecutive (`-race -count=3`)
3. preset lint → 0 issues (module dir, real binary); fix what it reports
4. README truth pass (hooks passthrough + grammar) + scratch-module
   snippet compile check
5. Check `scripts/example-smoke.sh` — add the preset example if the script
   enumerates examples explicitly
6. P9: preset CHANGELOG [0.1.0] + API diff record + fast guards → tag
   `preset/v0.1.0` → pre-tag-checks (expect the same unpushed-tag
   gosum-canonical red-by-hold on core v0.10.0 + preset legs) →
   release-state sync (AGENTS latest-per-module +preset v0.1.0, 13
   released; FEATURES preset section; TODO_LIST; ci.yml "eleven released"
   comment → twelve)
7. Consumer heads-up drafts (cqrs-htmx + PapDashboard) announcing preset
8. R1b–R1e (verify-blocks batch 2, ServerTiming bench baseline, hub.go:80
   note + libs index screenshot + dprint decision, go-retry MVS proof +
   example-smoke vs new stack)
9. S1–S7 scaffold (verdict doc → template pieces → dogfood)
10. G-CNT adoption-counter script + gate table + guard wiring
11. A1–A5 swap recipes (template + dnsblockd + ChastityAPI + PapDashboard
    - pbx-lens)
12. D1 worker supervisor; D3 sqlitekit; C2R realtime FoldedSource
13. R2 deps-guard CI+pre-commit; R3 error-family batch; R4 Server-Timing
    recipe; R5a httpx shortlist verdicts
14. H1–H7 httpx program (B1 + classification-parity test first)
15. X1 compose-on-httputil (re-run spike vs v1.7.0+), X2 TLS, X3 testkit
    migrations, X4 v1-exit watch
16. Y1–Y6 backlog; Z1–Z7 process debts; AGENTS pin-hold one-liner (e1
    from session 3)
17. OG1/OG3–OG6 owner memos (incl. catalog unblock ask)
18. Final `scripts/verify-all.sh` (13 buildable modules + integration);
    final release-state sync; session-close ritual
19. IF push-policy YES: push master + config/v0.1.0 + v0.10.0 +
    preset/v0.1.0; re-run pre-tag-checks (gosum legs self-heal);
    fresh-consumer proxy checks; P8 E2E. IF NO: hold and keep documenting.
20. Housekeeping: /tmp scratch cleanup; UNSET the global insteadOf git
    rewrite when local-tag tidies end.

## g) QUESTIONS FOR THE OWNER (cannot be answered from the repo)

1. **Push policy — 5th ask, now gating THREE local tags:** config/v0.1.0
   (all guards green), v0.10.0 (green except the unpushed-tag gosum leg —
   self-heals on push), preset/v0.1.0 (next). May I push master + tags as
   trains complete, or hold for one owner-reviewed batch at program end?
2. **Upstream unblock — go-cqrs-lite catalog templ v0.3.1100 migration**
   (blocks the integration cqrs-htmx/go-sse re-pin AND cqrs-htmx adopting
   templ-components v1.27.0): schedule inside this program, or leave to
   that repo's backlog? I am ASSUMING the latter (it is your repo, the
   HOLDS + held_pins guard make the wait safe) — correct me if you want it
   in-scope.
3. **Confirm or veto the two decisions I made under the blanket
   directive:** (a) the LOCAL v0.10.0 tag (the accumulate-posture
   deviation — preset needed RegisterClassifications under published-tag
   requires; reversible with `git tag -d v0.10.0` + revert of the
   integration re-pin), and (b) the preset.Config hooks passthrough
   addition (a design-note surface amendment — without it, preset
   consumers cannot register any lifecycle hook).

---

**Verification state at write time:** preset build/vet GREEN; full preset
suite 4/5 failure-iterations green (the 5s hang: errorpages+security+spine
reproduces ~1/3); integration suite GREEN on core v0.10.0; pre-tag-checks
v0.10.0 green except the structurally-held gosum-canonical integration
leg; local tags: config/v0.1.0 + v0.10.0 (unpushed); working tree
auto-committed by pma throughout; insteadOf rewrite still set globally.
