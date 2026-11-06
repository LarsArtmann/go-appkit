# Status: issue #54 resolution — health v0.3.1 released, four pre-existing reds repaired

**Created:** 2026-11-06 11:47 · **Scope:** resolve
[go-appkit#54](https://github.com/LarsArtmann/go-appkit/issues/54) ("cut
go-appkit v0.9.1 + a matching health minor so cvagent can stop riding a
pseudo-version"). Session followed READ → UNDERSTAND → RESEARCH → REFLECT →
execute-and-verify per step. Outcome: **no core tag was needed**, `health
v0.3.1` was cut and published, the full pre-tag gate ended GREEN, the issue
was answered with measured evidence and closed — and five PRE-EXISTING reds
that master was carrying (four failing gates/jobs plus one doc-truth claim)
were repaired on the way, because the release gate could not go green with
them in place.

---

## a) FULLY DONE

1. **Issue triage against source, not against the report.** Verified the ask
   line by line:
   - Server-Timing **is** in a published tag — `v0.10.0` (2026-11-06), tagged
     `middleware.go` puts `servertiming.ServerTimingMiddleware()` outermost on
     the DEFAULT stack (`ServerTiming → Recovery → RequestID → Logging →
     [Timeout] → SecurityHeaders`), `ServerTiming *bool` default on. There is
     no `v0.9.1` and none is needed; a patch there would be the same content
     under a misleading number.
   - Per-check durations shipped in **health v0.2.0** (2026-10-12, #25):
     `NewProbe` forwards to go-health `NewChecks`, so the payload carries
     `duration_ns` per check. `NewWithDetailedCheck` is go-health's own
     constructor, not an appkit API.
   - The genuine gap: health's dependency alignment between `health v0.3.0`
     and the pinned commit (`09a66bf`) — dashboard v0.10.2 → v0.12.2,
     error-family v0.12.0, templ-components v1.26.0, httputil v1.8.0, direct
     `samber/lo`. That content was unpublished-in-tag, which is exactly why
     `go list -m -versions` topped out at v0.3.0 for the consumer.
2. **`health v0.3.1` cut** (patch, correct SemVer for "no API change, deps
   move"): CHANGELOG `[Unreleased]` → `[0.3.1] - 2026-11-06`, annotated tag,
   pushed; artifact asserts (go directive 1.27.2 = floor, no `toolchain`
   line, no filesystem replace) all green.
3. **Zero-API-delta claim MEASURED, not asserted:** `git archive
   health/v0.3.0` → `go doc -all` diff against the working tree → **0 diff
   lines**. (Release Ritual step 1.)
4. **Full pre-tag gate GREEN.** Final run:
   `PRE-TAG GATE GREEN: health/v0.3.1 safe to push` and
   `GOSUM CANONICAL: 12 module(s) green against proxy + sumdb` — all eleven
   guards plus the tagged-artifact asserts.
5. **Integration pin choreography done in one train:** `integration/go.mod`
   health v0.3.0 → v0.3.1 and `documentedPins` moved together; `go mod tidy`
   run against the local repo via the documented `insteadOf` rewrite (so
   go.sum carries CANONICAL hashes), integration suite `-race` green, then
   the rewrite **unset** (verified no leftover global git config).
6. **Consumer proof from outside the repo, through the two tags only.**
   Throwaway module, `GOPRIVATE=''`, `GOPROXY=https://proxy.golang.org`, fresh
   `GOMODCACHE`, requiring `v0.10.0` + `health v0.3.1`:
   ```
   OK Server-Timing: total;desc="Total request";dur=0.023109
   OK /readyz 200: {…,"checks":{"cache":{…,"duration_ns":12419},
                                 "database":{…,"duration_ns":9249}}}
   ```
7. **Checksum truth confirmed against sum.golang.org:**
   `health v0.3.1 h1:8H4KOz5BxkHcBH74ZtmiDid3MAj2sMX4uYrfSJx2BwE=` and the
   `/go.mod` hash match `integration/go.sum` byte for byte (no re-tag class).
8. **Repair 1 — root test build was broken at HEAD (and pushed).**
   `hookclassification_test.go` used `errors.AsType[interface{ Unwrap()
   []error }]`; AsType constrains its type argument to `error`, so the root
   module's tests did not compile → `test (.)` and `fleet-lint` red on
   master for two commits. Fixed with a named `multiUnwrapper` interface that
   folds `Error()` in; core suite `-race` green, root lint 0 issues.
9. **Repair 2 — AGENTS release-state list unparseable → `pin-drift` red on
   every push.** The preset entry carried its release note (with commas)
   inside the `Latest per module:` list; the guard counted 15 claims for 13
   modules and then declared preset unclaimed. Note moved after the list
   terminator; guard green (31 OK, 0 FAIL).
10. **Repair 3 — `check-consumer-claims` red (pre-tag-only guard, so silent
    on master).** FEATURES claimed security's example pins core v0.9.0 while
    `security/go.mod` says v0.10.0. Claim corrected; security module
    build+test re-verified; guard green.
11. **Repair 4 — CI integration failure `TestErrorpagesBehindSecurityStack`.**
    Two requests closed bodies without reading them, so server-side conns
    stayed ACTIVE and `listener_close` burned the 5s stop timeout
    (`result=error`). Drained both; `-count=20` green, full suite green.
12. **`health/LICENSE` added** — health was the ONLY released module without
    the proprietary licence file every sibling carries (never existed in its
    history), so scanners/pkgsite had nothing to read. File + CHANGELOG
    `[Unreleased]` entry landed; rides health's next tag (deferred per the
    config v0.1.0 precedent — see §g-2).
13. **Module docs de-staled (health):** `WithHealthRecorder` gotcha rewritten
    (it now FAILS AT BOOT since v0.3.0 — the old text still described the
    silent-drop era), dependency line refreshed (go-health v0.8.1, dashboard
    v0.12.2, error-family v0.12.0, samber/lo), "NO core dependency" nuanced
    (library surface is core-free; the example is the require).
14. **AGENTS Release Ritual step 6 added:** the published-tag ordering lesson
    (pin-drift wants the newest LOCAL tag, gosum wants proxy resolution —
    mutually unsatisfiable pre-push) plus the empirically-measured proxy lag
    window (~45 min for a `.info` URL that was probed before publication).
15. **Issue answered and CLOSED** with the exact `go get` lines, the
    MVS note (health v0.3.1's own go.mod requires core v0.10.0), the measured
    proof, and the proxy-lag caveat.
16. **CI verified on the last CODE-BEARING commit** `68e7fe5`: **40/41 jobs
    green**, the single red being the designed `ci-preflight`
    SSH-secret watchdog. The final commit `45cfacf` is AGENTS-only and CI
    correctly skips it (`paths-ignore: **/*.md`) — verified in the workflow
    file rather than assumed.

## b) PARTIALLY DONE

1. **Proxy availability.** `proxy.golang.org` now serves `v0.3.1` (a plain
   `go get` with a fresh cache succeeds, gate leg green), but only after a
   ~45-minute window in which `.info`/`.mod` returned 404 while `.zip` and
   the sumdb entry were already live. I documented the window; I did NOT find
   any way to force the proxy to drop a cached resolution failure.
2. **LICENSE parity for the released artifact.** In-tree only — the published
   `health v0.3.1` zip still has no LICENSE. Deliberate (see §g-2), but it is
   a real, if cosmetic, mismatch between tagged and tree state.
3. **`preset/go.mod` still requires `health v0.3.0`.** Published preset v0.1.0
   is immutable; the in-tree pin was left alone so it rides preset's next
   train. Drift exists by decision, not by omission.
4. **Scope of verification.** I ran the FULL suite only for the modules I
   touched (core, health, security, integration) plus lint on health/root,
   health's example smoke, and the 12-module gosum/guard sweep. The other
   eight modules were covered only by CI's matrix, not by a local
   `verify-all.sh` run.
5. **`go.work.sum` churn (+9 `/go.mod` hash lines) was committed by the pma
   daemon** under a generic "auto-commit 3 changed file(s)" message. Additive,
   harmless, and plausibly tooling-induced (likely gopls/workspace load after
   my `integration/go.mod` edit), but I did not prove the origin.
6. **Session report itself** — this file (the item the owner asked for at the
   end, not at the start).

## c) NOT STARTED

1. **`core v0.10.1`** — core's `[Unreleased]` carries a real behavior change
   (`runHooks` now uses `errorfamily.WrapOnce` + `hook_index` context). Not
   required by #54, left unpublished.
2. **`health v0.3.2`** for the LICENSE-only delta (deliberately deferred).
3. **cvagent-side pin flip** — their repo; the issue now carries the exact
   commands.
4. **`docs/review/libs/*` errata** for the dashboard v0.12.2 / templ-components
   v1.24.0 a11y-markup change (the per-lib reviews still describe the older
   stack).
5. **A `verify-all.sh` end-to-end green record** for the whole fleet after the
   release (per-module only this session).
6. **`govulncheck` on health** after the dependency sweep (AGENTS says run it
   on dependency trains; I did not).
7. **The other 11 modules' unpublished dep-sweep drift** — untouched by design,
   but it is the same class that generated this issue.

## d) TOTALLY FUCKED UP (honest list)

1. **I poisoned the proxy cache for my own release.** The gate's integration
   leg (and my probes) requested
   `proxy.golang.org/…/health/@v/v0.3.1.info` BEFORE the tag existed. That
   404 stuck for ~45 minutes after publication, and I spent a long stretch
   re-checking "is the release broken?" — the release was fine the whole time
   (`.zip` + sumdb already correct). Avoidable: probe a URL that has never
   been requested (`.zip`) or just wait for a real `go get`.
2. **I reproduced the very bug I was fixing.** My first AGENTS release-state
   edit put the new health note INSIDE the `Latest per module` list, adding
   commas to the parenthetical — the exact failure mode that was already
   breaking `pin-drift` on master. Caught only on the next gate run, fixed in
   a follow-up commit. Running the guard immediately after the edit would
   have collapsed this into one change.
3. **Three tags for one release (all unpushed).** I created `health/v0.3.1`,
   moved it after discovering integration's go.sum could not exist yet, then
   moved it again after doc fixes. Each move was verified safe (`ls-remote`
   showed no origin tag; nothing was proxy-cached), but the sequence was
   sloppy — the right order is: finish every doc/pin edit → create the tag
   once → publish → never touch it again.
4. **I re-edited `health/CHANGELOG.md` after the tag was live** (only to open
   `[Unreleased]` for the LICENSE). Harmless, but it means the tag and the
   tree diverge in a doc file, which makes later "what is at the tag?"
   questions noisier than necessary.
5. **First draft of the consumer program assumed `testkit` helpers were usable
   from `main`** (they take `testing.TB`). One wasted iteration; read the
   signature first next time.

## e) WHAT WE SHOULD IMPROVE

1. **The release gate is structurally unsatisfiable for any module integration
   pins** (documented as AGENTS step 6). Better: teach `check-pin-drift` to
   accept "pin == newest local tag AND not yet on origin" during the release
   window, and let `check-gosum-canonical` report "deferred until published"
   instead of a red FAIL. That converts a documented workaround into a rule.
2. **Never probe a proxy URL for a version whose tag is not published yet.**
   This session's 45-minute detour. The gate itself is the first offender —
   `check-gosum-canonical` should skip (not fail) a pin whose tag is local-only.
3. **`check-pin-drift`'s claims parser is fragile by construction**: it cuts
   at the first `).` and splits on commas, so ANY prose inside the list
   silently corrupts the claim set (and the error message — "15 entries but
   12 parse" — does not name the culprit). A strict one-entry-per-line table,
   or a parser that reports the offending token, would have caught this class
   the day it landed.
4. **`check-consumer-claims` is pre-tag-only**, so a stale FEATURES claim sat
   red-free on master until a release tried to ship. Same class as fleet-lint
   before 2026-10-25: either give it a CI twin or derive the claim
   mechanically from `security/go.mod`.
5. **Root-cause the undrained-body class once.** This is now the third
   occurrence (preset probes, rate-limit SSE E2E, this errorpages E2E). Copy-
   pasted `io.Copy(io.Discard, resp.Body)` at ~15 sites is not a fix; a
   `testkit` drain-and-close helper plus a lint rule ("test body closed
   without read") would be.
6. **LICENSE coverage is not guarded** while AGENTS asserted it was ("every
   module root carries a LICENSE"). health had none for its entire life. A
   FAST offline `scripts/check-module-licenses.sh` (one line per go.work
   member) would make the claim mechanical.
7. **`docs/agents/module-details.md` still carries stale pin lines** (frh
   says `go-health v0.2.0`; others probable). Only the health section was
   refreshed this session.
8. **`docs/status/README.md` is behind by six reports** (the 2026-11-06
   03-18 / 05-03 / 06-22 / 06-59 / 07-50 / 09-32 files have no index rows).
   The index is the discovery surface; a stale index is a docs-truth bug.
9. **Release commits of modules that integration pins need a script**
   (tag → tidy → go.sum commit → re-tag is mechanical but easy to order
   wrong; I moved a tag twice).
10. **Daemon auto-commits interleave with release trains** and take my edits
    under generic messages. Fine mechanically, but a "release train in
    progress" cooldown (the pma memo already proposes one) would keep history
    legible.

## f) UP TO 50 THINGS TO GET DONE NEXT

1. Decide `core v0.10.1` for the unpublished WrapOnce/`hook_index` change.
2. Add `scripts/check-module-licenses.sh` (FAST+OFFLINE) → pre-commit,
   pre-tag, CI.
3. Index the six missing 2026-11-06 status reports in `docs/status/README.md`.
4. Rewrite `check-pin-drift`'s claims parser (strict list, name the offender).
5. Teach `check-gosum-canonical` to skip local-only tags (with a loud
   "deferred until publish" line) instead of failing the gate.
6. Add the same rule to `check-pin-drift`: "newest local tag not on origin"
   is a release-window state, not drift.
7. Add a CI twin (or go.mod-derived source) for the FEATURES security claim.
8. `testkit` drain-and-close helper + migrate the ~15 integration/test sites.
9. Add a lint/AST rule: response body closed without being read in tests.
10. Audit integration Cleanup funcs for 5s timeouts that mask the same class.
11. Add `-count=5` to the integration CI leg (or a nightly flake job).
12. Script the release dance: `scripts/tag-module.sh` (tidy against local tag,
    fold go.sum, create the tag exactly once).
13. Sweep `docs/agents/module-details.md` stale dependency lines (frh, others).
14. Add `govulncheck` to the dependency-train checklist execution (this train
    skipped it) — run it for health now.
15. Run `scripts/verify-all.sh` end to end and archive the green output.
16. Run `scripts/example-smoke.sh` fleet-wide (only the health leg this session).
17. Re-run `scripts/count-adopters.sh` (the newest-tag column moved).
18. Errata the per-lib reviews for dashboard v0.12.2 / templ-components a11y.
19. Bump `preset`'s health pin to v0.3.1 in-tree (rides preset's next tag).
20. Decide `health v0.3.2` (LICENSE parity) vs ride the next train.
21. Refresh the stale `docs/feedback/outgoing/…recorder-sentinel.md` (it says
    "filing gated" but the upstream ask SHIPPED).
22. Update TODO row 75 ("adopted + pinned by health v0.3.0") to v0.3.1.
23. Add a mechanical check: FEATURES "Shipped through <module> vX.Y.Z" equals
    the newest tag.
24. Document the deliberate hold on go-health v0.8.1 / dashboard v0.12.2
    (v0.10.0 / v0.13.2 exist upstream) — see §g-3.
25. Scratch-verify health against go-health v0.10.0 + dashboard v0.13.2 so the
    next train is a decision instead of a discovery.
26. Add `health_durations_test.go` as FEATURES evidence for the duration row.
27. Send the two pending Server-Timing consumer heads-up drafts (now that
    core v0.10.0 is out, they are dated).
28. Add a `README` note in health about the dashboard markup `role="region"`
    change in the v0.3.1 dependency alignment.
29. Re-check `pkg.go.dev` pages for health v0.3.1 (and the long-open
    flightrecorder 404 row).
30. Investigate `go.work.sum` churn provenance; decide whether to keep
    tracking it or generate it in CI.
31. Add a guard that the tagged CHANGELOG section's version/date matches the
    tag name.
32. Clean up the gate's `/tmp/go-appkit-gosum-*` scratch dirs left on failure
    (they are intentionally kept — document the cleanup command).
33. Consider a `health` module `LICENSE` requirement in the release ritual.
34. Reconcile AGENTS wording: "ten standing guards" vs eleven; "twelve
    modules" vs 13 released + integration.
35. Decide whether `integration/` should state its deliberate LICENSE absence
    in `integration/doc.go`.
36. Add a small `docs/recipes/module-release-ordering.md` (AGENTS step 6 is
    dense; a recipe with the exact command sequence would be reusable).
37. Batch: convert the session's ad-hoc verification program into a repo
    script (`scripts/verify-health-consumer.sh`) so the claim is re-runnable.
38. Add a Server-Timing assertion to a repo test (not just the consumer probe).
39. Sweep integration tests for other undrained bodies beyond the one fixed.
40. Check `Mounted.setStarted` godoc cross-link rendering (added earlier today).
41. Re-verify `health/example` after the dashboard bump (smoke passed; consider
    a golden of the dashboard HTML).
42. Add a `SetStarted`/drain-latch regression test if one is missing.
43. Consider `samber/lo` usage review in `NewProbe` (lo.MapEntries in a hot
    constructor — measure, or revert to the explicit loop if it allocates).
44. Re-run the art-dupl baseline check after new helpers land.
45. Add release-window awareness to the pma daemon (see §e-10).
46. Note the version-count drift on the module inventory in FEATURES/AGENTS.
47. Verify `preset` still builds against health v0.3.1 in a scratch module
    before its next tag.
48. Add the "probe `.zip`, not `.info`" trick to the release skill notes
    (crush-config `references/lessons.md` is the generalizable home).
49. Close out the `docs/status/README.md` "Current" section: archive the
    pre-November reports still sitting there.
50. Schedule the next health train (LICENSE + optional upstream sweep) so the
    deferred items do not become a second pseudo-version report.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **`core v0.10.1` now?** Core's `[Unreleased]` holds a real behavior change
   (`runHooks` classifies via `WrapOnce` and adds `hook_index` to hook
   failures). It is not needed for #54. Ship it as a patch, or let it ride the
   next feature train?
2. **`health v0.3.2` for LICENSE parity?** `v0.3.1` is published without a
   LICENSE (added in-tree afterwards). I deferred it, following the config
   v0.1.0 precedent (doc/metadata-only content after a tag waits). Is "ride
   the next health tag" the right call, or do you want a licence-only patch
   tag now?
3. **Are go-health v0.8.1 / dashboard v0.12.2 deliberate HOLDS?** Upstream
   already has go-health v0.10.0 and go-health-dashboard v0.13.2. The repo
   records holds for go-sse/cqrs-htmx but says nothing about these two. If
   they are held, what is the unblock trigger; if not, should the next train
   sweep them?

---

## Evidence index (commands that produced the claims above)

```bash
# zero-API-delta proof
git archive health/v0.3.0 health | tar -x -C /tmp/api-old-health
(cd /tmp/api-old-health/health && GOWORK=off GOTOOLCHAIN=go1.27.2 go doc -all .) > /tmp/old.txt
(cd health && GOWORK=off GOTOOLCHAIN=go1.27.2 go doc -all .) > /tmp/new.txt
diff /tmp/old.txt /tmp/new.txt        # 0 lines

# module verification
cd health && GOWORK=off GOTOOLCHAIN=go1.27.2 go build ./... && \
  GOWORK=off GOTOOLCHAIN=go1.27.2 go vet ./... && \
  GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1 && \
  GOWORK=off GOTOOLCHAIN=go1.27.2 golangci-lint run ./...
./scripts/example-smoke.sh health

# gate (final, green)
./scripts/pre-tag-checks.sh health/v0.3.1
#   -> PRE-TAG GATE GREEN: health/v0.3.1 safe to push
#   -> GOSUM CANONICAL: 12 module(s) green against proxy + sumdb

# canonical go.sum hashes vs sumdb
#   health v0.3.1 h1:8H4KOz5BxkHcBH74ZtmiDid3MAj2sMX4uYrfSJx2BwE=
#   health v0.3.1/go.mod h1:ZqE/VxMAbZQiaGBXdIkj/cGhMHxSNLEUl0Rgr8vRpw8=
grep "go-appkit/health v0.3.1" integration/go.sum

# proxy-only consumer proof (fresh cache, GOPRIVATE empty)
cd /tmp/proxy-e2e && GOWORK=off GOTOOLCHAIN=go1.27.2 GOPRIVATE='' \
  GOPROXY=https://proxy.golang.org GOMODCACHE=/tmp/proxy-e2e-cache go run .
```

## Commits landed this session (cdaf7b3 → 45cfacf)

| Commit | Subject |
| --- | --- |
| `2ff2c80` | fix: restore the root test build broken by the errors.AsType migration |
| `94503d4` | release: health v0.3.1 so consumers can pin a tag instead of a pseudo-version |
| `c728374` | chore: auto-commit (daemon: integration pin + go.sum + go.work.sum) |
| `4c14899` | docs: keep the release-state module list parseable by the pin-drift guard |
| `e70eb6d` | docs: correct the security example's core pin claim |
| `6f48fe0` | chore: auto-commit (daemon: AGENTS release-ritual step 6) |
| `33f1152` | test: drain response bodies before shutdown in the security/errorpages E2E |
| `68e7fe5` | docs: give the health module the license file every sibling already ships |
| `45cfacf` | docs: record how long a stale proxy 404 really lasts |

**Tag:** `health/v0.3.1` (annotated, on origin, verified via `ls-remote`;
proxy + sumdb live). **Last code-bearing CI run:** `68e7fe5`, 40/41 jobs
green, only the designed `ci-preflight` red. **Issue #54:** answered, closed.
