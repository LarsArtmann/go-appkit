# Status: SUPERB v12 execution — session 2 (config release closure + R1a + P1 research)

> **Live remainder:** the config-release session's 48-item ledger is fully
> adjudicated (this archive): the P1-P9 preset build shipped (21c1a78), the
> P8 E2E proven through the published tag (2de5875), G-CNT shipped
> (06e9a77), the push-policy hold resolved in effect (all tags pushed), and
> the scaffold/battery/machine-ops legs carry routed markers into their
> wired rows. Archived 2026-11-06: every §f item carries a verdict; nothing
> here is unowned.

**Created:** 2026-11-06 05:03 · **Scope:** session 2 of the "execute the WHOLE list"
directive (session 1 report: `2026-11-06_03-18_...`). This session: finished the C-train
lint/CI/README items, ran the C5 release train to a LOCAL tag (push held), executed the
entire R1a verify-block batch (libs 1–6), did the P1 preset design research, and hit —
then fixed — the first art-dupl gate catch.

---

## a) FULLY DONE

1. **C2–C3 lint closure — config at 0 issues** (was 25 at session end 1):
   - varnamelen ×7: `rv` → `target` rename through config.go (Load + apply\* +
     checkRequired + setParsed).
   - golines ×3: applied the pinned formatter (`golines --max-len=120`) — the diffs
     were pure line-splitting, reviewed before applying.
   - errcheck ×2: example now `_, _ = fmt.Fprintf(os.Stdout, …)` (the cqrs
     `metrics_test.go` discard precedent).
   - nolintlint ×1: removed the unused `//nolint:errorlint` (errorlint doesn't fire on
     `any`-typed assertions).
   - tagalign ×12: **this uncovered a real formatter/linter conflict** — see d/e.
   - Verification: `golangci-lint run ./...` → **0 issues**; tests green
     `-race -count=1` hermetic; `go vet` clean; gofmt clean.
2. **C4 — README snippet compile check:** scratch module at /tmp/readmecheck with a
   filesystem replace → all three README snippets (quick start, errors/errors.As,
   Keys) COMPILE_OK against the real API.
3. **CI wiring:** `./config` added to the ci.yml test matrix AND the proxy-smoke
   matrix (harmless until pushed; the v0.1.0 tag will ride the same push);
   "ten released modules" comment → eleven; actionlint green.
4. **Guard-gap claim from session 1 DISPROVEN (good news):** session 1 said the
   dependabot-parity guard "did NOT catch" the missing ci.yml slot. Wrong — the guard
   ALREADY checks matrix slots (line 47 of check-dependabot-parity.sh). Proven
   **born-red**: temporarily removing `./config` from ci.yml → guard FAILs with the
   exact right message; restored → green (26 OK lines). No guard extension needed.
5. **R1a — the full verify-block batch for libs 1–6, executed AS WRITTEN** (the
   "run as written, not transcribed" rule):
   - **httputil.md**: found the shipped rg bug — `rg -o 'httputil\.[A-Z]'` truncates
     every match to ONE character (`-o` prints only the matched span), so it can
     never produce "21 runtime symbols". Fixed to `[A-Z][A-Za-z]*` → verified count
     24. Also: httputil is now **v1.8.0** (apidiff gate, policy-vet, go-health
     guide, nonce CSP-conflict detection — read from their CHANGELOG), and F4
     (otel Recovery delegation) is RESOLVED (7 hits in otel/). Verify block fixed +
     dated errata block added.
   - **go-codec.md**: verified clean end-to-end (zero imports, 3 indirect pins,
     mod-why chain, v0.3.3 latest). No edits needed.
   - **go-error-family.md**: same `-o` truncation bug (fixed, 17 full symbols);
     the "empty → gap rows stand" claim moved — `appkit.RegisterClassifications()`
     (2026-11-03) engages `RegisterClassificationType`. Verify block fixed + errata.
   - **go-sse.md**: same `-o` bug (fixed); **go-sse v0.7.0 released** — exports
     `DefaultSubscriberBuffer` (the watched const); the BREAKING
     `ssetest.WithLastEventID(string→sse.EventID)` lives in the ssetest WORKING
     TREE only (ssetest's latest module tag is still v0.5.0, so appkit's pins are
     unaffected until they tag ssetest/v0.6.0). The F2.2 verify target was
     retargeted (hub.go documents the concept textually). Verify block fixed +
     errata with the re-pin adaptation instructions.
   - **go-etag.md**: verified as written; only drift: latest tag **v0.7.0** (was
     v0.6.1). Decoded the "duplicate v0.x lines" red herring (exclude-block entries
     carry no "exclude" text on entry lines; the sed range still lands at 173–178).
     Errata added.
   - **go-idempotency.md**: verified as written; drift: **v0.3.2** (cqrs carries it;
     integration still v0.3.1 — trains rule). Errata added.
6. **Realtime → go-sse v0.7.0 re-pin feasibility (scratch, not committed):** copied
   realtime to /tmp, bumped the require to v0.7.0, fresh-GOMODCACHE proxy tidy,
   `-race -count=1` → **GREEN**. The re-pin itself rides a realtime release train
   (member go.mod changes ride trains), but the evidence is banked in the errata.
7. **C5 release train (docs half):** module-details gained the full §config section
   (code org, deps, gotchas — incl. the tagalign/golines conflict as the fleet
   answer); AGENTS.md: module count 12→13 (also fixed the stale "systemd unreleased"
   text — systemd IS released), config bullet added, release-state narrative +
   latest-per-module updated, core-deps table httputil v1.7.0→v1.8.0; FEATURES.md
   gained the config section (8 FULLY_FUNCTIONAL rows); TODO_LIST v12-spine row →
   (a) DONE; config CHANGELOG dated [0.1.0] - 2026-11-06; `go doc -all` API surface
   captured (133 lines, for the release-ritual record).
8. **Annotated tag `config/v0.1.0` created locally** (twice — see d4) with the
   semantic-delta message. **PUSH HELD** per the standing no-push-without-ask rule.
9. **art-dupl gate catch — fixed at source:** the guard flagged 2 new 3-line clone
   groups in `config/fields.go` (the duplicated struct-recursion append-pair and the
   duplicated `if !hasTag`). Restructured the field walk from an if/continue chain
   into a `switch` + `walkChild` closure — both clones eliminated, no baseline
   change needed; tests stay green (28), vet clean; guard re-run GREEN.

## b) PARTIALLY DONE

1. **C5 pre-tag-checks: 3 remaining FAILs (all external drift, not config code):**
   - `check-agents-deps.sh` FAILED first on my own session drift (httputil v1.7.0 in
     the AGENTS deps table vs v1.8.0 in go.mods) — **FIXED and guard GREEN now**.
   - `check-pin-drift.sh`: (i) "./config is missing from the pin-guard module map
     (extend tag_pattern)" — the guard script needs a config entry (mechanical,
     next session's first item); (ii) integration pins drifted vs published tags:
     cqrs-htmx v4.15.0 → v4.16.0 and go-sse v0.6.2 → v0.7.0 (integration is a
     LATEST-pins charter module; re-pin + documentedPins fixture together; the
     go-sse leg is de-risked by the scratch verification above).
   - Also caught + fixed in the same pass: my "(local)" suffix on the
     latest-per-module line broke the guard's parser (12 entries, 11 parse).
2. **P1 preset design — research done, code not started:** extracted the preset plan
   (T1.1–T1.6 acceptance criteria + the main() sketch), gathered every battery's
   exact signatures (core ServiceConfig/MetricsConfig, systemd.Install, health
   Mount/Mounted/WithDashboard + route paths /healthz /readyz /startupz — NO
   collision with core's /health family, errorpages Config/Mount/Wrap, otel
   Setup/Provider/WithService/Middleware, realtime NewHub/Mount/Hub.Shutdown,
   security's full battery list + HeadersConfig/RateLimitConfig/Environment).
   Grammar drafted (see f). No module skeleton yet.

## c) NOT STARTED (unchanged from session 1, minus what's above)

P2–P9 (preset spine code + example + E2E + release); S-train scaffold; G-CNT
adoption counter; A1–A5 recipes; D1 worker; D3 sqlitekit; C2R realtime folded
broadcast; R1b–e (verify-blocks 7–12, ST bench baseline, hub doc note, MVS proof —
note R1a's method is now proven, R1b is mechanical); R2–R5a; H1–H7 httpx; X1–X4;
Y1–Y7; OG1–6; Z1–Z7; final verify-all + release-state sync; the pending consumer
heads-up drafts (OG2 class). Also not done: the fresh-consumer proxy check for
config/v0.1.0 (needs the push).

## d) TOTALLY FUCKED UP (own mistakes, all caught & corrected forward)

1. **multiedit REPLACED instead of appended (AGENTS.md):** adding the config module
   bullet, my old_string targeted the systemd bullet and my new_string contained
   only the config bullet — the systemd bullet was DELETED. Caught by reading the
   result immediately; restored in the next edit. Lesson: when "adding" a sibling
   entry, target the INSERTION POINT (e.g. the following header), never a
   neighboring bullet I don't intend to rewrite.
2. **`rg -rn` again — the FOURTH hook strike on this host.** I typed a filler
   `rg -rn ""` into a compound command without thinking. Blocked by the hook, zero
   damage, but the recurrence itself is the failure: the reflex is not installed.
   Mitigation going forward: write rg commands character-by-character when they
   carry flags, and prefer plain `rg -n` even when a glob would be marginally
   tidier.
3. **Tag created BEFORE the art-dupl fix was done:** I tagged config/v0.1.0 while
   pre-tag-checks was still running, then the art-dupl gate failed on the working
   tree — the tag pointed at a pre-fix commit. Fixed by deleting and re-tagging
   locally (safe ONLY because unpushed; a pushed tag must never move). Lesson: tag
   AFTER the full guard verdict, not in parallel with it.
4. **Session-1 misdiagnosis carried in:** session 1's claim that the parity guard
   does not check CI matrix slots was wrong (it does, born-red proven). I repeated
   it uncritically in this session's early notes before testing it. Lesson: a
   predecessor session's "guard gap" claim is a hypothesis, not a fact — one
   born-red probe settles it.
5. Minor shell sloppiness: `tail -1 f1 f2 …` (invalid context in this shell), a
   first multiedit rejected by the mtime check (daemon re-committed AGENTS.md
   mid-edit — correct tool behavior, I just had to re-read). Zero residue.

## e) WHAT WE SHOULD IMPROVE

1. **tagalign vs golines is a FLEET-latent conflict, now with a fleet answer:** any
   module whose structs mix present/absent tag keys (required fields with no
   default!) hits mutually-exclusive canonical forms — tagalign's align mode
   collapses absent keys, golines' realign-tags pads them; golangci v2.14 exposes
   no golines realign setting (schema-probed). Answer: `tagalign: align: false`
   (sort order still enforced; probe-verified it still flags env-before-default).
   Encoded in config/.golangci.yml + module-details §config. Copy it into the next
   satellite's config when tags are optional.
2. **The `-o` truncation bug class shipped THREE times in the libs docs** (httputil,
   go-error-family, go-sse — all `[A-Z]` one-char matches with `-o`). R1b (libs
   7–12) should grep every verify block for `-o '…\.[A-Z]'` patterns BEFORE
   running, and the crush-config lesson entry (asserted-vs-computed counts) should
   name this class explicitly — Z1/Z4 still open.
3. **Pre-tag-checks order:** run the FAST offline guards to zero BEFORE creating the
   tag (directives, parity, art-dupl, agents-deps), then tag, then the slow network
   guards (gosum-canonical, pin-drift). This session ran them in the wrong order
   and paid a re-tag.
4. **External drift arrives on its own schedule:** three upstream releases
   (httputil v1.8.0, go-sse v0.7.0, cqrs-htmx v4.16.0) landed between sessions and
   tripped the release gates. That's the guards working — but the integration
   re-pin train should be batched SOON rather than accumulating (it also un-blocks
   the pin-guard map extension for ./config in the same commit).
5. **The `walkChild` closure pattern for tree-walkers** (config/fields.go) kept the
   dedupe surgical without an extracted helper — worth remembering for the httpx
   tree-walk code later.

## f) NEXT — up to 50, in execution order

1. ~~Extend check-pin-drift.sh module map with `./config` (tag_pattern `config/v`)~~ done — the pin-drift guard covers config (the release ritual ran it green at the train)
2. ~~Integration re-pin train: cqrs-htmx v4.16.0 + go-sse v0.7.0 (+ documentedPins~~ done — integration re-pinned across the later trains (documentedPins grew; the preset pin added at 2de5875)
   ~~fixture, GOWORK=off suite, ssetest call-site check — expect NO changes, ssetest~~
   ~~stays v0.5.0)~~
3. ~~Re-run pre-tag-checks config/v0.1.0 → expect all-green (then C5 is fully closed,~~ done — pre-tag-checks green at the config train (the tag pushed 614cc2f; this week's full run green against health/v0.3.1)
   ~~push still held)~~
4. ~~P1: write the preset design note (grammar table freeze + posture table +~~ done at `21c1a78`
   ~~conflict matrix) — research is done~~
5. ~~P1: skeleton module `preset/` (go.mod floor 1.27.2, LICENSE, satellite~~ done at `21c1a78`
   ~~.golangci.yml — WITHOUT tagalign-align tweak unless tags appear, charter --fix,~~
   ~~dependabot, ci.yml matrix slots BOTH)~~
6. ~~P2: preset.New spine: ServiceConfig passthrough + systemd.Install +~~ done at `21c1a78`
   ~~RegisterClassifications + Version stamp~~
7. ~~P2: health leg: NewProbe(empty checks default) + Mount(WithDashboard) +~~ done at `21c1a78`
   ~~StartHook-wired Mounted.Start + Drain/Shutdown hooks via nil-guard closures~~
8. ~~P2: spine boot test: /healthz+/readyz 200, dashboard route 200, graceful stop~~ done at `21c1a78`
9. ~~P3: errorpages leg (spine-ON by default, APPKIT_ERRORPAGES=off opts out)~~ done at `21c1a78`
10. ~~P3: metrics passthrough documented (preset never invents creds; Version stamps)~~ done at `21c1a78`
11. ~~P3: full-spine-boots-with-zero-app-code test~~ done at `e26b7e5`
12. ~~P4: otel leg (APPKIT_OTEL trool; Setup+WithService(Name,Version)+Middleware→~~ done at `21c1a78`
    ~~OuterMiddlewares+ShutdownHook; env-absent → zero otel activity)~~
13. ~~P4: conflict test: explicit OTEL config wins over env both ways~~ done at `21c1a78`
14. ~~P5: security profiles: none/standard(SecurityHeaders+BodyLimit)/strict(+~~ done at `21c1a78`
    ~~OriginCheck with required origins); APPKIT_SECURITY enum parse~~
15. ~~P5: posture-matrix test (headers present per profile; opt-out removes cleanly)~~ done at `21c1a78`
16. ~~P6: realtime leg: NewHub + Mount(mux, path or "/events") + Hub.Shutdown→~~ done at `21c1a78`
    ~~ShutdownHook; APPKIT_REALTIME trool~~
17. ~~P7: example app main ≤40 LOC through preset~~ done at `21c1a78`
18. ~~P7: conflict-matrix suite (code-vs-env for all four switches + invalid values →~~ done at `21c1a78`
    ~~Rejection fail-fast)~~
19. ~~P8: integration E2E through PUBLISHED tags (needs preset v0.1.0 tagged AND~~ done at `2de5875`
    ~~config pushed — blocked on push policy)~~
20. ~~P9: preset CHANGELOG + release ritual + tag (HOLD push per policy)~~ done at `21c1a78`
21. P9: consumer heads-up drafts (cqrs-htmx + PapDashboard) **→ TODO P3 — post consumer heads-ups (OG2 drafts) <!-- fresh:skip: routed in TODO_LIST -->**
22. ~~R1b: verify-blocks libs 7–12 (go-retry, go-atomic-write, go-filewatcher,~~ done — R1b executed (verify-blocks 7-12; the p9a session)
    ~~go-daemon, go-output, server_timing) — grep for the `-o` bug class first~~
23. ~~R1c: ServerTiming bench baseline (quiet-run, pinned cores, -benchtime 1s~~ done — R1c executed (the AGENTS perf re-baseline; the 13-10 closeout)
    ~~-count 6) + AGENTS perf update~~
24. ~~R1d: hub.go:80 doc-const note + libs index mobile screenshot + dprint coverage~~ done — R1d executed (hub note + screenshot + dprint record; the 13-10 closeout)
25. ~~R1e: integration go-retry MVS proof + example-smoke vs the ServerTiming stack~~ done — R1e executed (the MVS proof + example-smoke 8/8; p9a)
26. ~~S1: scaffold verdict doc (template repo recommended; G3 default if unanswered)~~ done — the scaffold verdict doc landed (docs/planning/2026-11-06_08-45_scaffold-shape-verdict.md)
27. S2–S6: template pieces (cmdguard serve verb, templ skeleton, flake+buildflow, **→ TODO P3 — v13 plan S2-S7 scaffold rows (template pieces) <!-- fresh:skip: routed in TODO_LIST -->**
    CI+systemd+Dockerfile, doc skeletons + boot probe)
28. S7: dogfood one real app end-to-end **→ TODO P3 — v13 plan S2-S7 scaffold rows (S7 dogfood) <!-- fresh:skip: routed in TODO_LIST -->**
29. ~~G-CNT: adoption-counter script (dual any/direct) + gate table + guard wiring~~ done at `06e9a77`
30. ~~A1: swap-recipe template in docs/recipes/~~ done — the A1 template landed (docs/recipes/battery-swap-recipe.template.md + index row)
31. A2: dnsblockd recipe (−~1,600 LOC) **→ TODO P3 — v13 plan A2-A5 swap-recipe rows (dnsblockd) <!-- fresh:skip: routed in TODO_LIST -->**
32. A3: ChastityAPI recipe (−~4,800 LOC class; JWT→usermgmt flagged) **→ TODO P3 — v13 plan A2-A5 swap-recipe rows (ChastityAPI) <!-- fresh:skip: routed in TODO_LIST -->**
33. A4: PapDashboard security swap recipe **→ TODO P3 — v13 plan A2-A5 swap-recipe rows (PapDashboard) <!-- fresh:skip: routed in TODO_LIST -->**
34. A5: pbx-lens cmdguard migration recipe **→ TODO P3 — v13 plan A2-A5 swap-recipe rows (pbx-lens) <!-- fresh:skip: routed in TODO_LIST -->**
35. D1: worker supervisor + pools (acceptance per battery spec) **→ TODO P3 — Battery program waves W3-W5 (D1 worker supervisor) <!-- fresh:skip: routed in TODO_LIST -->**
36. D3: sqlitekit design + lease + backup + ledger + migration helper **→ TODO P3 — Battery program waves W3-W5 (D3 sqlitekit) <!-- fresh:skip: routed in TODO_LIST -->**
37. C2R: realtime FoldedSource + SubscribeFolded discriminator test **→ TODO P3 — Battery program waves W3-W5 (W5 C2 projection→broadcast folded contract)** <!-- fresh:skip: routed in TODO_LIST -->
38. ~~R2–R5a: deps-guard CI, error-family batch, ST recipe, httpx shortlist verdicts~~ done — R2-R5a executed (agents-deps dfeb4a5; the R3/R4a/R5a rows closed 2026-11-06)
39. H1–H7: httpx module program (B1 first; classification-parity test) **→ TODO P3 — Battery program waves W3-W5 (H1-H7 httpx program) <!-- fresh:skip: routed in TODO_LIST -->**
40. X1: compose-on-httputil refactor (the v1.7.0-seam swap plan) **→ TODO P2 — compose Service on httputil.Server (X1 the v1.7.0-seam swap) <!-- fresh:skip: routed in TODO_LIST -->**
41. X2/X3: TLS option + testkit migration as their trains come up **→ TODO P2 — compose Service on httputil.Server (X2 TLS + X3 testkit migrations ride the trains) <!-- fresh:skip: routed in TODO_LIST -->**
42. X4: core v1 exit-criteria watch (additions-empty anchor) **→ TODO P2 — core v1.0.0 exit criteria (X4 the additions-empty anchor watch) <!-- fresh:skip: routed in TODO_LIST -->**
43. Y1–Y7 + Z1–Z7 backlog/process debts (incl. Z1: add the `-o` truncation class to **→ TODO P3 — v13 plan Y/Z rows (backlog + process debts; Z1 the -o truncation class) <!-- fresh:skip: routed in TODO_LIST -->**
    the crush-config lessons entry)
44. OG1/OG3–OG6 owner memos; OG2 heads-ups when core tags **→ TODO P3 — v13 plan OG rows (memos; OG2 drafts ready) <!-- fresh:skip: routed in TODO_LIST -->**
45. ~~Final `scripts/verify-all.sh` (13 modules) + full guard suite~~ done — verify-all ran on the tagged trees (this week's waves + the p9a guard runs)
46. ~~Final release-state sync (AGENTS/FEATURES/TODO_LIST headers; 13-module truth)~~ done — the release-state sync shipped (AGENTS thirteen-released truth; FEATURES synced)
47. ~~CI green on the final pushed commit + session-close ritual (every new job ran)~~ done — CI verified on the final pushed commits (designed reds only; the status-index job live-green)
48. ~~IF owner answers push-policy YES: push master + config/v0.1.0 together, then~~ done — the push policy resolved YES in effect: master + config/v0.1.0 + preset/v0.1.0 + v0.10.0 all pushed
    ~~fresh-consumer proxy check, then P8 unblocks~~
49. ~~IF NO: hold; the proxy check + preset's published-tag requires stay documented~~ done — moot — the hold lifted (the proxy checks ride the P9c close-out route)
50. Housekeeping: /tmp scratch dirs (readmecheck, rt-repin, tagprobe, glcfg) are **→ TODO P3 — Machine ops (/tmp scratch dirs: readmecheck, rt-repin, tagprobe, glcfg) <!-- fresh:skip: routed in TODO_LIST -->**
    throwaway — leave or clean, no repo impact

## g) QUESTIONS FOR THE OWNER (cannot be answered from the repo)

1. **Push policy (asked in every report since session 1 — THE gating item):**
   config/v0.1.0 is tagged locally; C5's last step (fresh-consumer proxy check) and
   P8 (preset E2E through published tags) both require the tag on origin. May I
   `git push origin master config/v0.1.0` (and future session tags) as each train
   completes — or hold ALL pushes for one owner-reviewed batch at program end?
2. **G3 scaffold shape:** S1 recommends a template repo (go-website-template
   pattern; go-wizard-sdk's engine is unbuilt). Rule now, or shall I proceed with
   the template-repo default when the S-train starts?
3. **Security profile contents (P5):** I will freeze `standard` =
   SecurityHeaders(production) + BodyLimit(1 MiB) and `strict` = standard +
   OriginCheck(consumer-supplied origins) — composed only of shipped security v0.2.0
   batteries, no new capability. Veto or amend any ingredient now, or accept at
   P5 review?

---

**Verification state at write time:** config tests green `-race -count=1` (28),
lint 0 issues, vet/gofmt clean, verify-all(config) PASS; guards green: directives,
dependabot-parity (born-red matrix-slot proof), workspace-charter, golangci-configs,
agents-deps, art-dupl; pre-tag-checks(config/v0.1.0): 108 OK, 3 FAIL (pin-drift:
module-map extension + integration re-pin v4.16.0/v0.7.0 — external drift, fix
queued as f1/f2); tag config/v0.1.0 exists locally on the current commit, PUSH
HELD; working tree auto-committed by pma throughout.

---

## ERRATA 2026-11-06 (session 3)

1. **§a6 + §b1's "go-sse leg de-risked (scratch green)" was WRONG for the
   integration graph.** The scratch verified the REALTIME module only (no
   templ in its graph). Fresh scratch runs against integration's actual graph
   (2026-11-06): go-sse v0.7.0 AND cqrs-htmx v4.16.0 both fail to build —
   templ v0.3.1100's removed `templ.ResolveAttributeValue` breaks
   templ-components ≤ v1.26.0 (`utils/svg`) and go-cqrs-lite catalog ≤ v4.12.0
   (`docserver/specview_templ.go`); templ-components v1.27.0 shipped the
   migration but catalog (latest v4.12.0) has not, and cqrs-htmx v4.16.0 still
   requires templ-components v1.22.0. Both pins are now RECORDED HOLDS
   (`held_pins` in scripts/check-pin-drift.sh, exact-match + born-red proven),
   not drift. Unblock trigger recorded in the guard + the documentedPins
   comment.
2. **§d/§b's "the '(local)' suffix BROKE the guard parser" was backwards:**
   the suffix's `)` was accidentally the parser's list terminator. Removing
   the suffix left the trailing prose glued to the last entry (12 entries /
   11 parse). Fixed by restoring the `)` after `config v0.1.0` in AGENTS.md.
