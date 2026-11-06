# Status: SUPERB v12 execution — session 3 (pin-hold discovery, C5 closed, core v0.10.0 staged, preset skeleton + code)

**Created:** 2026-11-06 06:22 · **Scope:** resumed under the owner's standing
"execute the WHOLE list" directive (pushes still held). This session: closed
C5 (pre-tag-checks GREEN), DISPROVED session 2's integration re-pin plan with
scratch evidence and replaced it with recorded pin HOLDS + a guard allowance,
wrote + froze the P1 preset design note, scaffolded the preset module with
full first-draft code, and staged the core v0.10.0 release train end-to-end
(tag NOT yet created — see question 3).

---

## a) FULLY DONE

1. **f1 — pin-drift module map extended:** `config` added to
   `scripts/check-pin-drift.sh` tag_pattern (`config/v*`); AGENTS claim now
   verifies against the tag (was the first of the 3 pre-tag FAILs).
2. **The "(local)" parser mystery solved (backwards from session 2's
   claim):** the latest-per-module parser cuts the claims list at the first
   `)` after it — the dropped "(local)" suffix's paren was accidentally
   LOAD-BEARING (the list terminator). Removing it glued trailing prose onto
   the last entry (12 entries / 11 parse). Fixed by restoring the `)`
   after `config v0.1.0` in AGENTS.md (the 4d53277 format). Errata appended
   to the session-2 report.
3. **f2 — the integration "re-pin" DISPROVEN and replaced with recorded
   HOLDS** (the session's biggest reversal; all evidence from fresh-cache
   scratch runs at /tmp/integ-repin, /tmp/integ-chtmx):
   - go-sse v0.7.0 → requires a-h/templ v0.3.1100 → `templ.ResolveAttributeValue`
     removed → templ-components ≤ v1.26.0 generated code breaks
     (`utils/svg`) — exactly what the f7b00bb hold comment (2026-11-04)
     predicted; session 2's "de-risked" scratch was REALTIME-ONLY (no templ
     in its graph).
   - cqrs-htmx v4.16.0 drags go-sse v0.7.0 via MVS → same break (the
     cqrs-htmx leg is NOT independent drift).
   - templ-components v1.27.0 HAS shipped the migration, but go-cqrs-lite
     catalog ≤ v4.12.0 (latest published) still breaks
     (`docserver/specview_templ.go`) → the whole re-pin is blocked on an
     UPSTREAM catalog release. No integration pin changes made.
4. **held_pins guard allowance added** (the structural fix): check-pin-drift
   2b now has a `held_pins` map (cqrs-htmx v4.15.0, go-sse v0.6.2) — green
   ONLY while the pin is exactly the held version; pin-bumped-without-hold
   or hold-without-pin goes red. **Born-red proven** (a copy with one entry
   removed correctly FAILs that leg). documentedPins fixture comment
   rewritten with the full empirical picture + unblock trigger. Without
   this, the next push would have shipped a permanently-red pin-drift CI
   job (f7b00bb held the pin but taught the guard nothing).
5. **f3 — C5 CLOSED: `./scripts/pre-tag-checks.sh config/v0.1.0` → ALL
   GREEN** (112 OK lines, zero FAILs; verdict line: "PRE-TAG GATE GREEN:
   config/v0.1.0 safe to push"). Push still held per policy.
6. **P1 — preset design note written and FROZEN:**
   `docs/planning/2026-11-06_appkit-preset-design-note.md` — grammar table
   (4 switches), Config/leg sketches, posture table, conflict-matrix test
   plan, wiring order (nil-guard closures; hook append order), non-goals,
   module facts. One design error caught at review: preset must NOT require
   `config` (unused requires don't survive tidy; env parsing is direct
   os.Getenv with fail-fast Rejection).
7. **P1 skeleton — preset/ module created:** go.mod (floor 1.27.2), LICENSE,
   `.golangci.yml` (derived from systemd's satellite config, header adapted),
   README (grammar table + quick start + Build & verify), CHANGELOG
   ([Unreleased]), doc.go (package doc carrying the frozen grammar).
8. **P2–P6 first-draft CODE WRITTEN** (4 files): `config.go` (Config +
   TimeoutsConfig + OTEL/Security/Realtime/ErrorPages leg structs),
   `env.go` (triBool + security enum, Rejection fail-fast with
   `preset.env_invalid` codes), `preset.go` (New: spine + legs, nil-guard
   closures, otel provider flush on startup-error paths, ExtraMiddlewares
   for the security profile with Set-semantics overwrite rationale),
   `security.go` (standard = SecurityHeaders(Production)+BodyLimit(1 MiB);
   strict = +OriginCheck; strict-without-origins Rejection). Battery
   signatures re-verified via `go doc` BEFORE freezing (health Mount/Mounted
   lifecycle, otel Setup/WithService/Middleware/Provider.Shutdown, realtime
   NewHub/Mount/Hub.Shutdown, security HeadersConfig/OriginCheck(*slog.Logger),
   errorpages Config/Mount, systemd.Install, core ServiceConfig full field
   walk). `go mod tidy` resolved exactly the design-note require set (all
   published tags).
9. **Core v0.10.0 release train STAGED** (trigger: preset's spine requires
   `RegisterClassifications()`, which is working-tree-only — core latest
   published is v0.9.0; preset requires published tags):
   - API diff v0.9.0 → tree: **additions-only** (root: RegisterClassifications;
     testkit: FreeAddr, WaitForRunning/Within, HTTPGetURL/Context) → minor
     bump. Hermetic build+vet+test -race: GREEN.
   - CHANGELOG dated `[0.10.0] - 2026-11-06`.
   - Same-train docs: AGENTS.md (module count 13→14 incl. preset; core
     bullet + v0.10.0 items; preset bullet; Release State entry;
     latest-per-module → core v0.10.0), FEATURES.md (3 UNRELEASED rows →
     v0.10.0 + 2 new rows + version-truth paragraph — evidence files
     verified to exist).
10. **New-module wiring for preset, all guards green:** go.work charter
    `--fix` (root + 12 members incl. ./preset), dependabot `/preset` entry,
    ci.yml test-matrix + proxy-smoke slots (actionlint green), and the fast
    offline guard sweep GREEN: directives, dependabot-parity,
    workspace-charter, golangci-configs (preset's config verifies),
    agents-deps, art-dupl.

## b) PARTIALLY DONE

1. **Core v0.10.0:** staged (verified, dated, docs-synced, fast guards
   green) — the TAG is deliberately NOT created pending owner question 3
   (it releases earlier than the recorded "ACCUMULATE" posture; the sole
   driver is preset's published-tag require). Slow network guards
   (pre-tag-checks) not yet run for it.
2. **preset/ module:** full first-draft code compiles against core v0.9.0
   EXCEPT `appkit.RegisterClassifications` (undefined until core v0.10.0
   exists + preset re-pins to it via the VCS insteadOf tidy trick). NO
   tests written yet (P2–P7 test plan exists in the design note); lint not
   yet run.
3. **Integration re-pin train:** resolved to HOLDS (see a3/a4) — nothing
   left to do in THIS repo until upstream catalog migrates.

## c) NOT STARTED (unchanged from session 2, minus a/b above)

P2–P7 tests + example; P8 E2E (push-blocked); P9 preset release; consumer
heads-up drafts; R1b–e; S-train scaffold; G-CNT; A1–A5; D1 worker; D3
sqlitekit; C2R realtime folded broadcast; R2–R5a; H1–H7 httpx; X1–X4; Y1–Y7;
OG1–6; Z1–Z7; final verify-all + release-state sync; the config fresh-consumer
proxy check (push-blocked).

## d) TOTALLY FUCKED UP (own mistakes, all caught forward)

1. **I started bumping the integration pins before re-reading the fixture
   comment that documented the deliberate hold** (f7b00bb, one day old).
   Session 2's report framed both pins as "external drift... batched SOON";
   the fixture said go-sse was HELD with a trigger. The scratch failed
   exactly as the comment predicted. Same lesson class as session 2's d4
   (predecessor claims are hypotheses): triage "drift" from the RECORDED
   DECISION first, guards assert, comments explain.
2. **Session 2's "(local) suffix BROKE the parser" was backwards** — the
   suffix was accidentally load-bearing. I re-derived the parser mechanics
   instead of re-trusting the note (correct move), errata appended to the
   session-2 report.
3. **Two wasted cycles, zero damage:** the born-red guard probe first ran
   the script copy from /tmp and died on `AGENTS.md: No such file` (the
   script cds relative to its own path — re-ran with root pinned); my
   design-note draft required `config` in preset's go.mod for "docs
   symmetry" (tidy would prune it — caught at design review).

## e) WHAT WE SHOULD IMPROVE

1. **A deliberate pin hold MUST land with a held_pins entry in the same
   change** — f7b00bb held go-sse but taught the guard nothing, leaving a
   permanently-red pin-drift CI job armed for the next push. held_pins now
   exists; encode the rule (the guard's comment does; AGENTS should carry
   one line too — routed to f12 below).
2. **Errata timing:** disprove a point-in-time claim → append the errata
   immediately (this session did; make it reflex).
3. **Cross-repo blocker surfaced:** the integration re-pin (and
   cqrs-htmx's templ-components v1.27.0 adoption) now waits on go-cqrs-lite
   catalog's templ v0.3.1100 migration — an upstream task in the owner's
   other repo (question 2).
4. **gopls diagnostics for preset/ are noise** (BrokenImport/go-mod-tidy
   spam from workspace resolution) — the CLI tidy/build is the verdict;
   same known-stale class as config/ in session 2. Do not chase them.

## f) NEXT — in execution order

1. Owner answers g below → create (or redirect) the core v0.10.0 tag
2. `./scripts/pre-tag-checks.sh v0.10.0` → expect green
3. preset go.mod: core → v0.10.0 (VCS insteadOf tidy trick); build+vet green
4. P2 spine test: zero-app-code boot (empty Config, no env) — /healthz
   /readyz 200, dashboard 200, graceful stop
5. P2 hook-wiring test: Mounted.Start/Drain/Shutdown actually run at the
   right phases
6. P3 errorpages tests: pretty 404 HTML + JSON-by-Accept; `off` removes
   catch-all; consumer routes win over the catch-all
7. P3 metrics passthrough test: nil stays nil; set Metrics → /metrics;
   preset never sets creds
8. P4 otel tests: env absent → zero activity; on → middleware present;
   explicit config wins both ways
9. P5 security posture matrix: none/standard/strict headers, 413 oversize,
   OriginCheck reject, strict-without-origins Rejection
10. P6 realtime test: /events streams (sync on SubscriberCount), Hub
    out-field assigned, hub drained at shutdown
11. P7 conflict-matrix suite: 4 switches × (nil/env/code/invalid) per the
    design-note §5 table
12. P7 example/main.go ≤ 40 LOC
13. P7 preset lint → 0 issues (one module at a time)
14. P7 README snippet compile check (scratch module)
15. P9 preset CHANGELOG [0.1.0] + API diff record + fast guards → tag
    `preset/v0.1.0` → pre-tag-checks (core v0.10.0 unpushed will need the
    insteadOf trick there too — verify)
16. P9 release-state sync: AGENTS (latest-per-module +preset, 13 released),
    FEATURES (preset section), TODO_LIST, ci.yml "eleven released" comment
17. Consumer heads-up drafts (cqrs-htmx + PapDashboard) announcing preset
18. R1b: verify-blocks libs 7–12 (grep `-o '…\.[A-Z]'` first)
19. R1c: ServerTiming bench baseline (quiet-run, pinned cores, -count 6)
20. R1d: hub.go:80 doc-const note + libs index screenshot + dprint coverage
21. R1e: go-retry MVS proof + example-smoke vs the ServerTiming stack
22. S1: scaffold verdict doc (template-repo default per G3)
23. S2–S6: template pieces (cmdguard serve verb, templ skeleton,
    flake+buildflow, CI+systemd+Dockerfile, doc skeletons + boot probe)
24. S7: dogfood one real app end-to-end
25. G-CNT: adoption-counter script (dual any/direct) + gate table + guard
26. A1: swap-recipe template in docs/recipes/
27. A2–A5: dnsblockd / ChastityAPI / PapDashboard / pbx-lens recipes
28. D1: worker supervisor (×100 goroutine-baseline stable, panic-restart
    counts, start-during-shutdown joins)
29. D3: sqlitekit (Open+WithLease, Backup VACUUM INTO, OpenLedger,
    migration helper, SetMaxOpenConns(1) doc)
30. C2R: realtime FoldedSource + SubscribeFolded discriminator test **→ TODO P3 — Battery program waves W3-W5 (W5 C2 projection→broadcast folded contract)** <!-- fresh:skip: routed in TODO_LIST -->
31. ~~R2: deps-guard CI job~~ done at `dfeb4a5`
32. R3: error-family adoption batch (README recipe, BodyLimit cross-link,
    errorfamilytest asserts, WrapOnce, WithContext)
33. R4: Server-Timing handler-phase recipe
34. R5a: httpx shortlist verdicts
35. H1–H7: httpx program (B1 error mapping + classification-parity test
    first; B6/B8 stay owner-blocked on the svc.Routes() seam)
36. X1: compose-on-httputil refactor (re-run the spike vs v1.7.0+ ownership
    semantics, then the swap)
37. X2: core TLS option (rides the X1 seam)
38. X3: testkit migrations as trains come up
39. X4: core v1 exit-criteria watch
40. Y1–Y7: backlog (incl. Y-push-window items)
41. Z1: add the `-o` truncation class to crush-config lessons (repo commit)
42. Z2–Z7: process debts **→ TODO P3 — v13 plan Z2-Z7 process-debt rows (docs/planning/2026-11-06_07-10_SUPERB-v13-ship-the-spine-refresh-plan.md)** <!-- fresh:skip: routed in the v13 plan -->
43. AGENTS one-liner: "pin holds ride held_pins in the same change" (from e1)
44. OG1/OG3–OG6: owner memos (incl. the catalog unblock ask)
45. Final `scripts/verify-all.sh` (now 13 buildable modules + integration)
46. Final release-state sync (14-module truth across AGENTS/FEATURES/TODO)
47. CI green on the final pushed commit + session-close ritual (every new
    job ran)
48. IF push-policy YES: push master + config/v0.1.0 + v0.10.0 +
    preset/v0.1.0; fresh-consumer proxy checks; P8 unblocks
49. IF NO: hold; keep documenting the blocked-on-push state
50. Housekeeping: /tmp scratch cleanup (integ-repin, integ-chtmx, integ-gomod,
    old-core, probe-scripts, pd-probe)

## g) QUESTIONS FOR THE OWNER (cannot be answered from the repo)

1. **Push policy — 4th ask, now gating THREE local tags:** config/v0.1.0
   (pre-tag-checks GREEN), core v0.10.0 (staged green), preset/v0.1.0
   (next). May I push master + tags as each train completes, or hold
   everything for one owner-reviewed batch at program end? Note: origin's
   CI pin-drift job can only stay green on pushes that carry this session's
   held_pins guard; and preset's published-tag E2E + all fresh-consumer
   proxy checks need origin tags.
2. **Upstream unblock — go-cqrs-lite catalog:** the integration go-sse
   v0.7.0 / cqrs-htmx v4.16.0 re-pin is blocked on catalog shipping the
   templ v0.3.1100 migration (regenerate + golden re-baseline; it's your
   repo, latest tag v4.12.0 predates it). Schedule that migration inside
   this program (it also unlocks cqrs-htmx adopting templ-components
   v1.27.0), or leave it to that repo's own backlog?
3. **Core v0.10.0 timing — confirm or veto BEFORE I tag:** your recorded
   posture (2026-11-03 questionnaire) was ACCUMULATE, no immediate tag. The
   train is fully staged, but the only driver for releasing NOW is preset's
   spine requiring RegisterClassifications under published-tag requires.
   Tag v0.10.0 now (my recommendation — the alternative is preset shipping
   without the registry call and re-adding it later), or veto and preset
   v0.1.0 drops the RegisterClassifications spine member until your
   intended core train?

---

**Verification state at write time:** core build+vet+test -race GREEN;
preset tidy GREEN (requires resolve against published tags; RegisterClassifications
compile-blocked on core v0.10.0 by design); config pre-tag-checks GREEN
(112 OK); guards green: directives, dependabot-parity, workspace-charter,
golangci-configs (incl. preset), agents-deps, art-dupl, pin-drift (with the
two recorded HELDs, born-red proven); tags local: config/v0.1.0 (safe to
push), v0.10.0 STAGED-NOT-CREATED (question 3); working tree auto-committed
by pma throughout.
