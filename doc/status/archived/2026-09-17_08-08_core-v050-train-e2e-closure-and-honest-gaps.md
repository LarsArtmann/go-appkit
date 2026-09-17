# Status: Core v0.5.0 Train, E2E Closure, and Honest Gaps (2026-09-17 06:30 → 08:08)

**Session scope:** execute the previous report's §f backlog (the "instructed next
steps" of `2026-09-17_05-52_superb-plan-v2-execution-and-honest-gaps.md`):
core v0.5.0 train, the three silently-dropped E2Es (F104, composition
contracts, fr MetricsHook), health godoc examples, security hygiene, lint
sweep, and the truth-chore batch. **One tag shipped: `v0.5.0` (core).**
16 commits (`c147653..2ba2625`), all pushed; tree clean.

Headline: **all 11 module suites `-race` green, 0 golangci issues in every
module (sequential), structure linter 0, AGENTS within cap, fresh-consumer
proxy test PASS for core v0.5.0.**

---

## a) FULLY DONE

1. **Core v0.5.0 release train (the §f #2/#3 item).** API-break check per
   the Release Ritual (`go doc -all` diff v0.4.0 vs working tree: 78 diff
   lines, ADDITIONS ONLY — `MetricsConfig` type, `Metrics`/`Version` fields,
   doc-comment updates; the one `c` block is doc text). CHANGELOG
   `[Unreleased]` → `[0.5.0] - 2026-09-17`. Hermetic verify green
   (`GOWORK=off GOEXPERIMENT=jsonv2`, `-race`). Annotated tag `v0.5.0`
   pushed (commit `929f363`). **Fresh-consumer proxy test PASS** (clean
   `/tmp` module, `go get go-appkit@v0.5.0`): `/version` → 200
   `{"version":"v0.5.0-consumer-check"}`, `/metrics` with Basic Auth → 200
   with `appkit_build_info`, unauthenticated → 401, shutdown phase logs
   emitted with `result=ok`.
2. **Integration re-pinned** to core v0.5.0 + realtime v0.1.1; full suite
   green 3× on the new pins (pin bump rode daemon commit `ca7205b` — noted
   in §d-7).
3. **F104 — `cqrs.OTelProjectionMetrics` real-cycle E2E** (`c03360e`), the
   silently-dropped item from the SUPERB plan. `cqrs/otelmetrics_e2e_test.go`:
   boots the facade domain with `EventConfig.Metrics` wired to the OTel
   bridge + ManualReader, dispatches a real command, waits for the
   projection worker, asserts `cqrs.projection.event.count[processed] ≥ 1`
   AND that the datapoints carry the REAL `cqrs.projection.name` /
   `cqrs.event.type` attribute values (exported constants — the unit tests
   only ever fed synthetic events). SCOPED INTO THE CQRS MODULE, not
   `integration/`: adding go-cqrs-lite's system engine to the integration
   module would flip it to GOEXPERIMENT-required and break its lean-charter
   (documented in the TODO close). Fixture refactor: `newFacadeServiceCfg`
   (config-injectable boot), `newFacadeService` delegates. 3× `-race`
   stable.
4. **Core composition-contract suite** (`adee737`,
   `integration/composition_contract_test.go`), the §f #20 item, tested
   against PUBLISHED v0.5.0:
   - `TestVersionAndMetricsComposeWithDefaultHealth`: new surfaces coexist
     with default health endpoints — no route conflicts; exact
     `/version` JSON; `text/plain; version=0.0.4` exposition type; 401
     without credentials; stable metric names present; version label on
     build info; served requests land in `responses_total` with
     `status="200"`. Uses `testkit.Serve` — so the testkit sub-package
     itself is proxy-pinned too.
   - `TestDrainWindowContract`: readiness = 503 BEFORE DrainHooks run;
     `/ping` = 200 DURING the drain window (traffic served);
     ShutdownHooks ran; `Running()` false after; post-shutdown request
     fails at connection level. **Discovery (see §d-5):** `svc.Addr()` is
     already nil inside DrainHooks — `Shutdown` reaps `s.ln` FIRST. The
     base URL must be captured before shutdown. Pinned + documented in the
     test header.
5. **fr MetricsHook real-capture E2E** (`8802fdd`,
   `otel/frmetrics_e2e_test.go`), §f #21/#23: a REAL `fr.Recorder`
   (runtime/trace → `bytes.Buffer` sink, `MinAge 50ms / MaxBytes 1MiB`,
   100ms warmup, `recorderMu` singleton guard) drives the hook through the
   recorder's own plumbing — one manual `Snapshot` + latch `Reset` + one
   `SnapshotIf` trigger capture — asserting
   `appkit_flightrecorder_snapshots_total{source=manual}=1`,
   `{source=trigger}=1`, and a positive duration-histogram sum. Sink
   non-emptiness asserted (real trace bytes). The old tests only called the
   hook directly with fabricated events.
6. **Health godoc Examples — F107 + F113** (`a55c21c`,
   `health/example_test.go`), three runnable examples with verified
   `// Output:` blocks: criticality grading (critical pass + non-critical
   fail → `warn` + ready), panicking-check isolation (`warn` + ready),
   and the F113 two-probe aggregate (critical + optional probes on one mux
   with independent route sets; draining the optional probe flips ONLY its
   readiness). Full health suite `-race` green 3×.
7. **Compile-verified README snippets** (`38518ad`): new Metrics-config +
   testkit usage blocks in the core README, compiled AND behavior-checked
   in a scratch consumer module against PROXY-resolved v0.5.0 before
   landing (the `Doc snippets are code` rule, applied proactively this
   time).
8. **Lint sweep — every module 0 issues, sequential** (§f #35): root,
   cqrs, otel, health, integration, security, docs, errorpages. Fixed the
   accumulated root debt: `metrics.go` struct literals now name every
   field (removed two misplaced `exhaustruct_v5` nolint directives — they
   sat AFTER the literal, which is why nolintlint flagged them unused);
   `example/main.go` demo prints moved to `fmt.Fprintln(os.Stdout, …)` +
   named `demoDrainDelay` const (no forbidigo/mnd hits, no nolint
   anchors); cqrs `staleness_test.go` two inline-error findings cleared.
9. **Truth-chore batch** (`98cdf97` + daemon commits):
   - TODO_LIST: integration-expansion CLOSED (5/5 sub-items, F104 scoping
     note), fr-polish CLOSED (MetricsHook E2E evidence), fr-ops-preset
     CLOSED, health-parity `[~]` reduced to govulncheck-only, pkg.go.dev
     item refreshed with today's re-check, header re-truthed (11 modules,
     release state through core v0.5.0). **Count gate held: 38 → 38.**
   - FEATURES.md: three core rows (metrics surface, `/version`, testkit) +
     v0.5.0 ship note.
   - AGENTS.md: core line → v0.5.0, realtime → v0.1.1, frh → v0.1.2,
     health → v0.1.1, integration pin list + composition-suite note,
     Release State line rewritten, combined code-organization row for
     `metrics.go`/`version.go`/`testkit/`.
   - `doc/TELEMETRY.md`: the `shutdown phase skipped` pointer now cites
     the root README "Log volume" section (was the wrong §5).
   - SUPERB plan file: EXECUTION VERDICT banner at top.
   - Pin-test doc retired to `doc/planning/archived/` (its LANDED header
     was already annotated); its one live path citation updated.
   - `doc/status/README.md`: archived count (28) + the archive gate
     command.
   - Root README: security module row + "fully opt-in, nothing in the
     default stack" note.
   - infertypeargs sites (§f #34): VERIFIED OBSOLETE — zero markers in the
     file, zero linter findings; the concern died with a config/toolchain
     change.
   - Dependabot (§f #41): `/security` block diffed against `/realtime` —
     byte-identical format, nothing to change.
   - **§f #45 was a STALE CLAIM:** BOTH archived READMEs already carry the
     Annotation standard (`doc/planning/archived/README.md` since
     2026-09-17, `doc/status/archived/README.md` since 2026-09-16).
     Re-verified before acting; nothing to do.
10. **security module hygiene:** `.golangci.yml` header rewritten
    security-specific (was verbatim realtime inheritance citing family
    deps this module doesn't use) + **fresh-worktree go.work smoke PASS**
    (detached worktree at HEAD, `GOWORK=off` build + `-race` test green
    outside the workspace — go.work is gitignored, so this is what a fresh
    clone sees).
11. **Retract docs/v0.2.0 (§f #47) — evaluated, DEFERRED with rationale**
    recorded in `docs/CHANGELOG.md [Unreleased]/Planned`: the directive
    only affects consumers once a version CARRYING it is fetched, so it
    needs a docs release train; `@latest` already resolves v0.3.0 (highest
    semver) and an explicit `@v0.2.0` fetch fails loudly. Not worth a
    release today.
12. **pkg.go.dev re-check (§f #1):** core v0.4.0, realtime v0.1.1, otel
    v0.1.1 pages RENDER. docs@v0.3.0, health@v0.1.1, security@v0.1.0 still
    404 at 08:08 (~2.5h after push) — each fetch re-enqueues the crawl.
    TODO stays `[~]` with the three named pages.

## b) PARTIALLY DONE

1. ~~**AGENTS.md line budget:** my content edits were line-neutral, but the~~ done (noted — the stale cap note is folded into the TODO_LIST P3 AGENTS slim-down item (2026-09-17))
   ~~structure linter flagged 378/377 — and the finding PRE-EXISTS this~~
   ~~session (the pre-edit commit `80140d7` fails identically; see §d-8).~~
   ~~Fixed by deleting the near-zero-information `doc.go` table row; AGENTS~~
   ~~now at 376 wc-lines, linter 0. The "≤377 counted lines" note in AGENTS~~
   ~~is now known-stale (the binary counts one more than `wc -l`) — the~~
   ~~working buffer is 1 line, not 0.~~
2. ~~**The §f list itself:** 20 of its 50 items were actionable this session;~~ done (accounted — the remainders were harvested into TODO_LIST by the 2026-09-17 docs-health pass)
   ~~18 done, 2 partially (this item and pkg.go.dev). The rest were~~
   ~~user-gated, demand-gated, or env-blocked (see §c).~~
3. ~~**Proxy re-verification of older tags:** only core v0.5.0 got the~~ done (noted — same-day proofs stand; re-verification rides the next release ritual)
   ~~fresh-consumer treatment today; the six 2026-09-16 tags keep their~~
   ~~same-day proofs.~~

## c) NOT STARTED (deliberately — gates, not neglect)

1. **Upstream asks (§f #11–15): ALL USER-GATED** — go-sse
   dedup-aware `ReplayFiltered` filing, httputil Logging request-context
   filing, the httputil `NewServerListener` listener-injection API
   (unblocks the composition refactor AND Core TLS), the composition
   refactor re-run, Core TLS.
2. **govulncheck (§f #18): env-blocked** — `go install` is
   network-restricted here; needs a networked machine (health + security).
3. **Browser CSP pass (§f #22): not attempted** — needs a real browser
   (chromedp or manual); server-side CSP/nonce is proven only.
4. **Batteries W3/W4/W5 (§f #24–32): demand-gated** per the canonical
   battery spec — no consumer demand signal since triage.
5. **cordis bridge (§f in §c): 2 of 3 triggers still unmet** (go/v0.1.0
   tagged = trigger 1 met).
6. **PapDashboard v0.3.0 re-check:** lives in THEIR repo; user-gated.
7. **statusRecorder swap (§f in C28): USER GATE unanswered since
   2026-09-16.**
8. **Core v1.0.0 graduation:** consumer count still too low.
9. **Pareto plan v3:** per §f #49, only AFTER the user answers the gates.

## d) TOTALLY FUCKED UP (all recovered, all instructive)

1. **Port-8080 panic in the proxy consumer script — a REPEAT of last
   session's fuckup #7.** AGENTS says 8080 is SigNoz's port on this box;
   the previous report listed exactly this mistake; I hit it again anyway
   (`panic: service did not start`). Fixed with `Addr: "127.0.0.1:0"`.
   Lesson NOT learned fast enough: env-gotcha checks must happen before
   the first run, not after the first failure.
2. **Edit-tool rejection dance on CHANGELOG.md:** two rejected `edit`
   attempts because the tool tracks VIEW reads, not bash reads; only the
   `view`+`edit` pair worked. Two wasted round trips on a two-line edit.
3. **F104 first draft had three defects** I only caught by post-write
   verification: an `errMetricNotFound` sentinel that doesn't exist,
   wrong attribute keys (`cqrs.projection`/`cqrs.event_type` instead of
   the exported `AttrProjectionName`/`AttrEventType`), and an unused
   `sdkmetric` import. The right order is read-the-bridge-source FIRST,
   write the test SECOND — the exports were sitting in `otelmetrics.go`
   the whole time.
4. **`getWithAuth(t, svc)` SIGSEGV in the drain hook** — nil deref on
   `svc.Addr().String()`. Half-legitimate (it exposed the real
   Addr-nil-during-drain contract, now pinned as a feature of the suite),
   but I found it by PANIC, not by reading `service.go` `Shutdown` first —
   the `s.ln = nil` line was 20 lines into the function I had already
   opened for other reasons. Read the lifecycle code before writing
   lifecycle tests.
5. **Two Go-scoping build-fix cycles in the composition test** (`svc`
   referenced inside its own initializer; `baseURL` declared after the
   closure that captures it). Both are 10-second pre-thought errors.
6. **The `time.MicrostringPlaceholder` typo** — a nonsense identifier
   slipped into an edit while fixing golines. Caught by build; never
   ship a token you couldn't explain.
7. **getWithAuth ctx-refactor missed one call site** (+1 build cycle) and
   the health example's `startErr :=` redeclare (+1). Pattern: mechanical
   refactors across call sites deserve a grep, not memory.
8. **The AGENTS 378-line finding pre-existed and nobody noticed.** The
   previous session's "structure linter 0 / AGENTS at its 377-line cap"
   claim no longer holds: the INSTALLED binary now counts a 377-`wc`-line
   file as 378 (trailing-newline off-by-one — the same binary-vs-source
   drift AGENTS already documents for `exclude_patterns`). I verified the
   pre-edit commit fails identically before fixing (dropped the trivial
   `doc.go` row). Meaning: **the linter was not re-run after the last
   AGENTS edit of the previous session.** The AGENTS note about the cap is
   itself now stale.
9. **Daemon-heuristic commits absorbed three of my changes** (integration
   pins, security header, doc truth batch minus AGENTS). History is
   complete but unreadable in those spots. My per-unit commits were fast
   enough to win only ~60% of the races.

## e) WHAT WE SHOULD IMPROVE (process)

1. **Env-gotcha gate as an executable reflex:** before ANY first run of a
   server-binding script, check AGENTS' port list + banned commands.
   Two sessions, same 8080 mistake. A `freeAddr()`-style default should be
   the FIRST thing written, not a fix.
2. **Read-the-code-before-writing-the-test for lifecycle contracts.** The
   drain-hook panic produced the right test in the end, but via SIGSEGV
   archaeology instead of 20 lines of reading. Cheap rule: for any test
   that observes shutdown, read the shutdown function first.
3. **Write tests against EXPORTED constants, never re-typed string
   literals** — the F104 attribute-key bug was invisible until the
   post-write check. When the SUT exports constants, the test MUST import
   them.
4. **Never trust "0 findings" claims across sessions for gate tools** —
   re-run the structure linter (and the full lint set) after ANY change to
   a capped/linted file, even line-neutral ones. The 378 discovery shows
   gates rot when binaries move under identical-looking files.
5. **The structure-linter binary needs the same buildflow-style upstream
   fix as `exclude_patterns`**: its line count differs from `wc -l` by the
   trailing newline. Until then AGENTS carries a 1-line buffer, not 0.
6. **Beat the daemon by committing within the same tool call as the last
   edit** — worked for code units, lost on doc batches. Batch doc edits
   into ONE commit immediately, not after the next verification step.
7. **The "verify external claims before acting" rule paid for itself
   twice today** (§f #45's stale "gap" claim; the pre-existing 378
   finding) — keep re-verifying the previous report's TODO-adjacent
   claims instead of executing them as truth.

## f) NEXT — ordered by leverage (carry-forward + new discoveries)

**Ship loop / correctness**
1. ~~**Re-check pkg.go.dev for the three lagging pages** (docs@v0.3.0,~~ done (CLOSED 2026-09-17 — all three pages render (verified live by the docs-health pass))
   ~~health@v0.1.1, security@v0.1.0) after the crawl window; close the~~
   ~~`[~]` when all render.~~
2. ~~**Core doc-fix candidate (NEW, from §d-4):** `Service.Addr`'s doc says~~ done (done 2026-09-17 — v0.5.1 tagged + pushed (doc-only))
   ~~"returns nil before Start is called" — it should ALSO say "and from~~
   ~~DrainHooks onward once Shutdown begins". Godoc-only; fold into the~~
   ~~next core tag (v0.5.1 or ride v0.6.0) + a one-line CHANGELOG entry.~~
3. ~~**Same for `Running`:** it reports "has a bound listener", so it flips~~ done (done 2026-09-17 — v0.5.1 carries both godoc fixes (Addr + Running))
   ~~false the moment Shutdown starts — document, or reconsider the~~
   ~~semantics (v1.0.0-relevant: consumers can't distinguish~~
   ~~"never started" from "draining").~~
4. ~~**testkit double-shutdown nuance (NEW):** calling `svc.Shutdown` in the~~ done (routed — TODO_LIST P2 testkit explicit-shutdown helper)
   ~~test body AND letting testkit cleanup shut down again works~~
   ~~(idempotent) but cleanup's `errCh` wait adds latency to every such~~
   ~~test; consider `TestServer.Shutdown` helper that marks the cleanup as~~
   ~~no-op. Low priority.~~
5. ~~**Finish the lint-pass discipline:** the six modules re-linted today~~ done (adopted — the re-lint rule is restated in the 08-35 plan guards)
   ~~are clean; keep the "re-lint after ANY gate-relevant edit" rule so the~~
   ~~next session doesn't inherit another pre-existing failure.~~
6. ~~Root CHANGELOG for the NEXT wave: keep dating `[Unreleased]` sections~~ done (done — the v0.5.0/v0.5.1 dated sections repeat the pattern)
   ~~per-train (the v0.5.0 pattern worked; repeat it).~~
7. ~~**Fresh-consumer proxy test as a repo script:** today's check was a~~ done (done 2026-09-17 — doc/recipes/fresh-consumer-proxy-check.md exists and is indexed)
   ~~hand-rolled /tmp module (third time hand-rolled); commit it as~~
   ~~`doc/recipes/fresh-consumer-proxy-check.md` or a tiny script so the~~
   ~~ritual is copy-paste.~~

**Upstream (user-gated — say the word)**
8. ~~File Draft 1: go-sse dedup-aware `ReplayFiltered` (draft ready at~~ **Won't implement — USER-GATED — TODO_LIST P2 upstream-asks item.**
   ~~`doc/feedback/outgoing/2026-09-16_upstream-asks-gosse-httputil.md`).~~
9. ~~File Draft 2: httputil Logging request-context emit (same file).~~ **Won't implement — USER-GATED — TODO_LIST P2 upstream-asks item.**
10. ~~GREEN-LIGHT DECISION: implement `httputil.NewServerListener(ln, cfg,~~ **Won't implement — USER-GATED — TODO_LIST P2 upstream-asks item (NewServerListener go/no-go).**
    ~~handler)` upstream — unblocks the composition refactor (C20) AND Core~~
    ~~TLS (G1) in one move.~~
11. ~~Then: re-run the composition spike against that API; execute the C20~~ **Won't implement — gated on 10 — TODO_LIST P2 composition item.**
    ~~refactor for real.~~
12. ~~Then: Core TLS option (`ServiceConfig.TLS{CertFile, KeyFile}`) —~~ **Won't implement — gated on 10 — Core TLS in the AGENTS Deferred Register.**
    ~~PapDashboard's first demand.~~
13. ~~statusRecorder USER GATE (open since 2026-09-16): swap or close it.~~ done (STILL OPEN — TODO_LIST P2 USER GATE (statusRecorder))

**Finish the partials**
14. ~~**govulncheck** on health + security from a networked machine.~~ **Won't implement — env-blocked — TODO_LIST P2 govulncheck.**
15. ~~**Browser CSP pass** (chromedp or manual) over the health dashboard~~ done (routed — TODO_LIST P2 browser CSP pass)
    ~~under strict-CSP + nonce — server-side proof exists, browser-side~~
    ~~doesn't.~~
16. ~~**Security module threat-model page** (§f #39 from the last report —~~ done (routed — TODO_LIST P2 security trio)
    ~~NOT done, carried): per-battery threat → test mapping table.~~
17. ~~**Security example service** (§f #42, carried): like errorpages/example,~~ done (routed — TODO_LIST P2 security trio)
    ~~demonstrating the full hardened chain.~~
18. ~~**security + realtime composition integration test** (§f #43,~~ done (routed — TODO_LIST P2 security trio)
    ~~carried): rate-limit in front of SSE.~~
19. ~~**CI dependabot-parity assert** (§f #48, carried): workflow step that~~ done (routed — TODO_LIST P2 CI dependabot-parity assert)
    ~~fails when a module dir lacks a dependabot entry or CI matrix slot.~~
20. ~~**HTML reports sweep** (§f #33, carried): annotate or archive the four~~ done (done 2026-09-17 — docs-health pass banner-annotated the research HTMLs; the 08-16 plan archived)
    ~~never-opened reports (2026-08-15/16, 2026-09-04 research).~~
21. ~~cqrs README cookbook re-verification against scenario/v4 v4.2.0 after~~ done (standing — TODO_LIST P3 cqrs README cookbook item)
    ~~each go-cqrs-lite release (standing ritual, next release triggers it).~~

**Batteries (demand-gated — re-confirm demand first)**
22. ~~W3 B1 ResultHandler family (classification-parity pin with errorpages).~~ **Won't implement — demand-gated — TODO_LIST P3 W3 (B1).**
23. ~~W3 B2 bind+validation.~~ **Won't implement — demand-gated — TODO_LIST P3 W3 (B2).**
24. ~~W3 B9 no-leak error responses.~~ **Won't implement — demand-gated — TODO_LIST P3 W3 (B9).**
25. ~~W3 B4 conditional GET (promotes the idle go-etag dep).~~ **Won't implement — demand-gated — TODO_LIST P3 W3 (B4).**
26. ~~W5 C2 projection→broadcast folded contract (the cqrs+realtime~~ **Won't implement — demand-gated — TODO_LIST P3 W5 (C2).**
    ~~must-have).~~
27. ~~W5 C1 SSE drop/backpressure counters (absorbs TELEMETRY §6 candidate).~~ **Won't implement — demand-gated — TODO_LIST P3 W5 (C1).**
28. ~~W4 D4 atomic file write (floor: go-atomic-write ≥ v0.5.1).~~ **Won't implement — demand-gated — TODO_LIST P3 W4 (D4).**
29. ~~W4 D7 idempotency store.~~ **Won't implement — demand-gated — TODO_LIST P3 W4 (D7).**
30. ~~W1 F2 timing battery (after upstream Draft 2 lands).~~ done (sequenced after upstream Draft 2 (TODO_LIST P2))

**Watchlist / hygiene**
31. ~~Watchlist refresh (cordis consumers count, PapDashboard v0.3.1+,~~ done (rolled into the standing watchlist item (TODO_LIST P3))
    ~~nixpkgs toolchain > 1.26.7, dprint exit-14).~~
32. ~~cqrs `WithRecorderOptions`-style passthrough evaluation if a consumer~~ done (demand-gated — revisit when a consumer asks (cqrs opt-in class))
    ~~asks for recorder tuning beyond `OpsRecorderPreset`.~~
33. ~~Fold the AGENTS "377-line cap" note into "376 target / binary counts~~ done (folded — the TODO_LIST P3 AGENTS slim-down item carries the +1 note)
    ~~+1" once confirmed stable (or fix the binary upstream).~~
34. ~~Consider moving the fresh-consumer proxy check into CI as a~~ done (routed — TODO_LIST P2 fresh-consumer proxy smoke in CI)
    ~~manual-dispatch workflow (needs network in the runner — verify first).~~
35. ~~After the next real consumer adopts core v0.5.0: revisit core v1.0.0~~ done (standing — TODO_LIST P2 v1.0.0 exit-criteria item)
    ~~exit criteria (consumer-count trigger).~~

## g) Questions I cannot answer myself

1. ~~**Listener-injection go/no-go:** may I implement~~ **Won't implement — still USER-GATED — TODO_LIST P2 upstream-asks item.**
   ~~`httputil.NewServerListener(ln, cfg, handler)` in the httputil repo and~~
   ~~re-run the composition spike? It is the single unlock for BOTH the~~
   ~~composition refactor and Core TLS — but it is upstream work in a repo~~
   ~~you own, and the last standing gate on it (Service API posture) is~~
   ~~yours to call.~~
2. ~~**Browser CSP pass:** do you want me to attempt it here with a headless~~ done (routed — TODO_LIST P2 browser CSP pass)
   ~~browser (chromedp is installable only with network access I don't~~
   ~~have), or will you run the manual pass on your machine against the~~
   ~~health example? If manual: the strict-CSP profile to load is~~
   ~~`health` example + `DashboardHardenedPreset`.~~
3. ~~**Structure-linter line counting:** the installed binary counts one~~ done (folded — the TODO_LIST P3 AGENTS slim-down item carries the binary-counts-+1 note)
   ~~more line than `wc -l` (trailing-newline off-by-one), so AGENTS.md can~~
   ~~never sit AT the 377 cap — only below it. Fix the binary upstream in~~
   ~~go-structure-linter (same bucket as the inert `exclude_patterns` bug),~~
   ~~or keep a permanent 1-line buffer in AGENTS and update its stale~~
   ~~"≤377 counted lines" note?~~

---

*Point-in-time snapshot as of 2026-09-17 08:08 CEST. Release-state truth
lives in AGENTS.md → Release State; open work in TODO_LIST.md (38 items).*
