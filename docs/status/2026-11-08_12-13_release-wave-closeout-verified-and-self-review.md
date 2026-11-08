# Status — v15 closeout verified end-to-end + self-review (2026-11-08 12:13)

Scope line — what this report VERIFIES vs what it does NOT:

- VERIFIED this session (by command, artifact named inline): the
  gosum-canonical repair (guard green `/tmp/gosum4.log`), integration
  suite `-race` green ×2 post-repair, fresh-consumer compile+vet green
  (scratch module, fresh `GOMODCACHE=/tmp/fc-v15-cache`), verify-all
  ALL GREEN 129s (`/tmp/verify-all-v15.log`), example-smoke GREEN
  (`/tmp/example-smoke-v15.log`), structure-linter + status-index +
  agents-deps + golangci-configs exit 0, httpx lint green under BOTH
  golangci v2.14.0 (`/tmp/gcl2140`, installed exactly as CI) and
  v2.15.0 (local), verify-all.sh lint-loop fix born-red-proven with a
  stub linter, CI run 42469238305 on final code commit `6c6c9e6` =
  43 green / 1 designed red (ci-preflight) / structure-linter
  skipped-by-design, docs-gates pre-flight SAFE over the landed
  TODO_LIST lines (`--duplicates` mode).
- NOT verified here: a full `pre-tag-checks.sh` re-run AFTER the go.sum
  repair (its relevant legs — pin-drift, gosum — ran green standalone
  and in CI; the assembled claim is sound but it is assembled, not one
  invocation); the otel CI flake was classified by one error line +
  non-recurrence, NOT root-caused; consumer-claims green is inherited
  from the prior session's pre-push run (my later changes touch nothing
  those claims read); pkg.go.dev renders for the four new versions were
  NOT probed (crawler lag; proxy `.zip` 200s are the release contract).

Session context: continuation of the 2026-11-08 four-tag wave
(cqrs v0.9.0, realtime v0.2.0, httpx v0.2.0, systemd v0.2.0 — all
pushed + proxy-cached BEFORE this session started). This session: the
gosum repair, the master push, the full verification battery, two CI
failure fixes, and the close-out. Tip: `f96c0be`, tree clean.

## a) FULLY DONE (this session)

1. **gosum-canonical repaired** — root cause: integration/go.sum carried
   FIRST-CUT zip hashes for httpx AND systemd (the insteadOf tidy ran
   before the re-cuts; the handoff only suspected httpx). Repair per
   the AGENTS stale-VCS-cache recipe: strip the wave-pin lines, purge
   the poisoned download-cache entries + the stale VCS bare clone
   (`~/.cache/go-mod/cache/vcs/671aa33…`), re-tidy via
   `GOPROXY=proxy.golang.org` resolution. Net go.sum delta: exactly two
   hash lines. Guard GREEN; **no tag was moved** (origin+proxy final).
2. **Master pushed** through `f96c0be` (explicit detailed commits:
   `08f487d` go.sum repair, `6c6c9e6` wrapcheck version-proofing; the
   daemon took the intermediate content commits).
3. **Fresh-consumer check** — one scratch module consuming all five
   tags (4 wave + core v0.12.0) at compile level: every new API
   (`NotifyFolded`, `FoldedSource`, `WrapWriter`, `Conditional`,
   `Negotiate`, `WriteDeadline`, `Install`-returns-`Hooks`) plus the
   cross-module generic instantiation
   `realtime.FoldedSource[cqrs.FoldedEvent]` — the W5-C2 structural
   contract proven AT THE CONSUMER SITE. Build + vet green.
4. **verify-all ALL GREEN** (14 modules, tests + lint, 129s);
   **example-smoke GREEN** (all examples boot/probe/drain);
   structure-linter, status-index, agents-deps, golangci-configs all
   exit 0.
5. **CI failure #1 fixed — the silent fleet-lint death**: verify-all.sh
   lint loop lacked the test loop's `if` guard; `set -e` aborted the
   script at the first nonzero lint exit BEFORE the verdict line — CI
   showed "exit 1 at --- lint httpx ---" with ZERO findings displayed.
   Fixed (guarded like the test loop), born-red-proven with a stub
   golangci-lint on PATH. This bug hid every CI lint failure since the
   fleet-lint job was born (2026-10-25) — it only manifests on
   FAILURE, so all prior "green" evidence was sound, but the tool's
   failure mode was broken for two weeks.
6. **CI failure #2 fixed — golangci v2.14.0/v2.15.0 wrapcheck drift**:
   CI (v2.14.0) nolintlint-flagged an inline `//nolint:wrapcheck` on
   `httpx/negotiate.go:220` as UNUSED (v2.14 doesn't flag the site);
   local (v2.15.0) NEEDS the suppression (it flags the chained
   `.WithHTTPStatus(` builder). Fixed version-inertly: inline directive
   removed, `.WithHTTPStatus(` added to wrapcheck ignore-sigs (config
   sigs are inert when unused — nolintlint never inspects them). Green
   under BOTH binaries; config schema verifies under both.
7. **CI verdict on final code commit `6c6c9e6`: 43/44 green** — sole
   red the DESIGNED ci-preflight SSH-secret gate (owner-only); otel's
   one-run `CloseIdleConnections` transport flake did NOT recur
   (single occurrence → flake, not regression; nightly `-count=5` leg
   owns it if it returns).
8. **Close-out**: both 06-37 + 08-38 reports annotated (§h blocks,
   questions resolved/dispositioned); TODO_LIST header carries the wave
   final state; the parked cqrs engine-lesson pair landed as an open
   row; docs-gates freshness pre-flight SAFE after rewording (first
   wording would have flipped a 2026-09-23 item to DRIFT).

## b) PARTIALLY DONE

- **golangci version unification**: httpx is now green under both
  2.14.0 and 2.15.0, and this session incidentally produced dual-
  version evidence for all 14 modules (CI ran 2.14 on `6c6c9e6`,
  verify-all ran 2.15 locally) — but the FLEET still runs split
  versions BY CONFIG (CI pins 2.14.0, the local nix binary is 2.15.0).
  Every future suppression decision re-faces the same fork. Unify (see
  §g1).
- **otel `CloseIdleConnections` flake**: dispositioned as the
  documented shared-DefaultTransport race class by signature +
  non-recurrence; root cause NOT instrumented (goroutine-dump proof
  not done for THIS occurrence). If the nightly flake leg catches it,
  the fix pattern is already known (per-test transport isolation).
- **Upstream ask re sync-publish ↔ blocking-handler coupling**: asked
  twice (06-37, 08-38), still no owner answer; parked as a TODO row
  with both options framed. Third ask in §g2.

## c) NOT STARTED (deliberately parked — no verschlimmbessern)

- cqrs `StartProjections(ctx)` doc warning (ctx-lifetime trap) — TODO
  row exists; needs a cqrs train to ride.
- B6/B8 (route introspection + docs feed) — H8-owner-gated, unchanged.
- H8 seam + auth-bridge owner decisions — untouched, memos standing.
- pkg.go.dev render checks for the four new versions — crawler lag
  window; proxy contract already satisfied.
- The parked SUPERB v12/v13 remainder rows (per-repo battery-swap
  recipes, scaffold repo creation, machine-ops tail) — not this
  session's scope.

## d) TOTALLY FUCKED UP

1. **gosum whack-a-mole, AGAIN**: round 1 fixed only httpx (the one the
   log named), re-ran the guard, systemd popped. The fix-as-a-CLASS
   lesson is written in AGENTS and I still fixed the instance. Round 2
   stripped all four wave pins — that should have been round 1. Cost:
   one extra guard cycle (~4 min) and one extra tidy. The handoff even
   said "re-generate integration go.sum against the PUBLISHED tags"
   (plural).
2. **Wrote fresh-consumer code from memory, not from the API surface**:
   `httpx.WithETag()` does not exist (invented), plus a paren mistake —
   two wasted build cycles before grepping `func With` in the actual
   source. The API-shape grep I eventually ran should have come FIRST.
3. **Remove-first, verify-second on the nolint**: I deleted the inline
   directive before checking whether v2.15.0 still needed suppression
   at that site — local lint went red on the chained call, and only
   then did I find the real sig. Checking both versions BEFORE touching
   the line would have been one cycle instead of two.
4. **"Fresh-consumer ×4" was one module, not four**: a single scratch
   module `go get`s all five tags and exercises every API. The EVIDENCE
   is equivalent-or-stronger (compile-level, cross-module
   instantiation impossible in four isolated modules), but the ritual
   wording says one throwaway module per target; my close-out text said
   "×4" without noting the deviation. Corrected here.

## e) WHAT WE SHOULD IMPROVE

1. **Kill the tidy-before-final-tag class MECHANICALLY** — this is the
   THIRD incident of the class (health v0.2.0, the 2026-11-07 four-tag
   train, this wave). The poison is not tag ORDER, it is tidying
   integration against a tag that later moves. Rule candidate: the
   integration re-pin tidy ALWAYS finishes with a pure
   proxy-resolution pass (`GOPRIVATE= GOPROXY=proxy.golang.org go mod
   tidy`) — which would have FAILED loudly on the first-cut hashes
   before any push. Cheap, version-proof, encode-able in
   `scripts/tag-module.sh --tidy-integration`.
2. **Unify the golangci version** (see §g1): two lint versions in one
   fleet is a standing split brain; this episode was its bill.
3. **Config-only wrapcheck suppressions for error-family sites**: an
   inline `//nolint:wrapcheck` on a family-constructor return is
   version-fragile exactly as demonstrated (needed under one binary,
   nolintlint-"unused" under the other). House rule: family sites get
   ignore-sigs entries, never inline nolints. The AGENTS revert-proof
   rule already says "new family API → sig in every consuming config";
   extend it to "and never as an inline nolint".
4. **Guard failure modes deserve negative tests at BIRTH**: the
   verify-all lint-loop swallow bug shipped green in 2026-10-25 because
   it was only ever tested on success. The stub-linter negative probe I
   wrote is 5 lines — it should be a permanent fixture (or the loop's
   shape review) for every guard that redirects output.
5. **docs-gates pre-flight BEFORE landing** (known rule, I did it
   backwards): I landed the TODO lines first, then pre-flighted, then
   had to reword on disk with `--duplicates` modeling. Order matters;
   the tool even refuses the honest check once landed.
6. **Read the script you lean on**: I used verify-all.sh as green
   evidence all session and only READ its lint loop when CI contradicted
   local. Skimming the failure paths of load-bearing scripts at first
   use is the cheap insurance.

## f) NEXT (ranked; honest 28, not padded to 50 — the remainder of a
forced 50 would be re-listing parked rows)

1. golangci version unification decision + train (CI pin ↔ local nix
   binary; absorb-work for 2.15 already done) — closes the standing
   split brain.
2. Encode the proxy-resolution tidy as the MANDATORY final step of the
   integration re-pin (extend `tag-module.sh --tidy-integration`) —
   kills incident class #3.
3. House-rule + sweep: replace any remaining inline `nolint:wrapcheck`
   on error-family sites with config ignore-sigs (fleet-wide grep;
   httpx `result.go:24`, `writer.go:58`, `conditional.go:191`,
   `negotiate.go:218` are pass-through sites — verify each is
   version-stable first, they may be fine).
4. Verify pkg.go.dev renders for cqrs v0.9.0 / realtime v0.2.0 /
   httpx v0.2.0 / systemd v0.2.0 (window should be clear by now).
5. Root-cause-or-close the otel `CloseIdleConnections` flake if the
   nightly `-count=5` leg ever catches it (suspect: a parallel test's
   `CloseIdleConnections` on shared `http.DefaultTransport`; fix =
   per-test transport).
6. cqrs `StartProjections(ctx)` doc warning (parked row; rides next
   cqrs train).
7. Owner decision: sync-publish ↔ blocking-handler upstream ask vs
   appkit doc caveat (parked row; §g2).
8. ci-preflight SSH secret restore (owner-only P1; the ONE designed
   red left).
9. B6/B8 route batteries — blocked on H8 owner decision, memos
   standing.
10. go-sse v0.7.0 + cqrs-htmx v4.16.0 + templ-components unblock watch
    (templ breakage hold; trigger recorded).
11. Per-repo battery-swap recipe instances (dnsblockd first — the
    −1,600 LOC exemplar; consumer repos execute).
12. `go-service-template` sibling repo creation (owner-gated) + the
    dogfood probe.
13. Auth-bridge owner decision (memo standing; gates the preset auth
    leg).
14. H8 route-seam owner decision (memo standing; gates B6/B8 + the
    endpoints-verdict row).
15. httputil `endpoints` v0.1.0 verdict + H8/B6/B8 fit assessment
    (`docs/review/libs/endpoints.md` row).
16. W4/W5 battery remainders (polite caching on go-etag client+metrics
    v0.6.0; realtime W5-C1 `WithOnDrop`/drop-counter TODO).
17. Realtime subscribe-before-flush fix (demand-gated; sketch ready).
18. cqrs EventConfig opt-ins (encryption/signing/idempotency recipes)
    — demand-gated, unchanged.
19. otel OTLP logs signal (gate closed-not-met; re-run consumer audit
    on next claims refresh).
20. SigNoz live validation (accepted-risk; needs an instance).
21. `scripts/count-adopters.sh` re-run after this wave — four new
    tags, zero new consumers expected; watch drift.
22. Consumer heads-up check: does any consumer now pin httpx/systemd?
    (None did at wave time; verify on next claims refresh.)
23. pma owner checklist execution (daemon working-file class; memo §
    checklist standing).
24. Machine-ops tail: /tmp go-build sweeper policy, BuildFlow binary
    rebuild, cache-DB VACUUM (parked P2 row).
25. Machine watch: nixpkgs go still 1.26.10 < floor (standing).
26. otel upstream v1.48 watch (standing).
27. docs-health HARVEST of this report's §f into TODO_LIST (the rows
    above that are new: #2, #3, #5 as watch wording; most others
    already live as rows — dedupe at harvest time).
28. Nightly flake leg review after a week of `-count=5` runs on the
    post-wave tree (otel + integration SSE classes).

## g) QUESTIONS (cannot derive)

1. **golangci version unification — which direction?** CI pins
   v2.14.0; the local nix binary is v2.15.0. The 2.15 absorb-work
   (ignore-sigs in nine configs + schema-verify guard behavior) is
   DONE and this session proved the fleet green under BOTH — so my
   recommendation is CI → v2.15.0 everywhere. Confirm (or rule local
   pins back to 2.14.0).
2. **The upstream ask (third time):** synchronous-publish ↔
   blocking-handler coupling — file the verified upstream ask to
   go-cqrs-lite/watermill, or record an appkit/cqrs doc caveat only?
   This disposition gates parked row #7 and was left unanswered twice.
3. **Approve the ritual hardening:** "the integration re-pin ALWAYS
   ends with a pure proxy-resolution tidy" (+ encode in
   `tag-module.sh --tidy-integration`). It would have caught all three
   incidents of the stale-hash class pre-push. Yes/no as standing rule?

— Session closed out at `f96c0be`; waiting for instructions.
