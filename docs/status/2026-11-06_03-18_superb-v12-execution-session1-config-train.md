# Status: SUPERB v12 execution — session 1 (config train C1–C4)

**Created:** 2026-11-06 03:18 · **Scope:** this session only — execution of
`docs/planning/2026-11-06_01-48_SUPERB-v12-unified-execution-plan.md` per the owner's
"execute the WHOLE list" directive; default sequencing (no G1 answer): C-train first.
**Nothing outside the plan's C-train was touched.**

---

## a) FULLY DONE

1. **Context gathering for the C-train (plan slices 1–4).** Loaded `how-to-golang` +
   `buildflow` skills (first path guess for buildflow was wrong; corrected). Read: TODO_LIST
   in full; battery spec F1/B1–B9/C1–C3/D1–D7/G-cluster sections; the v12 plan's T2/T1/T3/T4
   task tables (extracted from the HTML); public APIs of realtime, systemd, errorpages,
   security, health, otel, core (`ServiceConfig`, `RegisterClassifications`, `MetricsConfig`);
   the realtime `.golangci.yml` standard; exemplar configs — **quotaboard** (the canonical
   envOr/envBool/envDurationOr hand-roll) and **dnsblockd** (the fleet's ONE koanf user —
   the evidence that settled the koanf verdict).
2. **C1 — `appkit/config` design note:**
   `docs/planning/2026-11-06_02-30_appkit-config-design-note.md`. Tag grammar frozen
   (`env:"NAME"` / `,required` / `default:"..."` / struct-field `env:"PREFIX"` prefixes;
   embedded structs bind promoted), precedence frozen (flag > env > file > default),
   error taxonomy frozen (typed `FieldError` inside one Rejection), **koanf verdict:
   DECLINED** (fleet reality: 21 flat env-first packages + 25 os.Getenv parsers vs one
   koanf repo that already has it), non-goals documented (no watch/F4, no interpolation).
3. **C2/C3 (code) — the `config/` module exists and its tests are GREEN** hermetic
   (`GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1` → ok, 28 tests):
   - `config.go` — `Load[T]`/`Loader[T]`, options (`WithFile`, `WithLookup`,
     `WithValues` flag layer, `WithValidate`), `FieldError`(+Kind enum), `Keys[T]`
     (push-model discoverability), `FieldErrorsOf` (join-tree walker).
   - `fields.go` — reflection walk: nested `env:"PREFIX"` key prefixes (child tags are
     relative), embedded-struct promotion (incl. unexported embedded types), untagged
     fields never touched (what makes embedding `appkit.ServiceConfig` safe).
   - `parse.go` — one shared parser for every layer (string/bool/all int+uint
     widths/float32-64/time.Duration/[]string; decimal-only ints; named-type `Convert`).
   - `file.go` — YAML (go-faster/yaml v0.4.6) + JSON (stdlib v2) file layer; lowercased
     dotted key space (`nested.addr`); **strict unknown-file-key rejection**.
   - `config_test.go` — 28 tests: full precedence matrix (flag>env>file>default,
     empty-flag fallthrough), missing-required → Rejection with field path, bad-value →
     classified never-panic, bad default caught at Load, YAML/JSON/nested/strict/missing/
     unknown-ext file cases, embedded promotion (env AND file key), unsupported type,
     flag layer + unknown field, validate hook, non-struct type param, `Keys` (flat +
     nested), default-uses-process-env (t.Setenv, non-parallel by design), Loader reuse.
   - `example/main.go` runnable (happy + missing-required paths verified live);
     `README.md` (quick start, precedence, tag grammar, error contract, discoverability,
     non-goals, Build&verify); `CHANGELOG.md` [Unreleased]; `LICENSE`; `.golangci.yml`
     (realtime satellite standard).
4. **Module wiring:** go.work gained `./config` (via `check-workspace-charter.sh --fix`,
   guard green "root + 11 modules"); dependabot.yml gained the `/config` gomod entry;
   `check-dependabot-parity` + `check-go-directives` + `check-workspace-charter` all green.
   Deps: go-error-family v0.12.0 (published tag), go-faster/yaml v0.4.6 (public proxy).

## b) PARTIALLY DONE

1. **C2–C4 lint pass — 25 of 49 findings remain** (was mid-iteration when this report was
   requested). Fixed already: err113 ×7 (errorfamily constructors instead of errors.New),
   exhaustruct_v5 ×4 (documented nolints), goconst `(file)` → `fileToken` const,
   forbidigo (example now `fmt.Fprintf(os.Stdout, …)` per fleet convention),
   forcetypeassert (checked bridge in `WithValidate`), cyclop (parseValue split →
   `parseScalar`), wsl/gci/golines/tagalign/formatting batch via `--fix`.
   **Remaining 25:** varnamelen ×7 (`rv` → rename), tagalign ×12 (test+example struct
   tags want `default` BEFORE `env`, alphabetized+aligned — README/design-note sketches
   show the opposite order; reordering the tags is fine but README snippets must be
   updated to match), golines ×3, errcheck ×2 (example Fprintf), nolintlint ×1 (now-unused
   errorlint directive).
2. **CI wiring for the new module — NOT in ci.yml** (found while writing this report):
   the per-module matrix is a HARD-CODED list and lacks `./config`. The dependabot-parity
   guard did NOT catch it (it validates dependabot entries + matrix slots for the modules
   it knows; the gap class is the same one that let `golangci-config-verify` ship
   green-by-assumption on 2026-10-13). **Must add before push.**
3. **README snippet compile-check:** the example binary carries the same struct/API and
   compiles+runs, but the doc-snippets-are-code rule asks for a scratch-module check of
   the README text itself — not yet run.

## c) NOT STARTED (nothing beyond C-train began)

C5 (release train: charter/CHANGELOG dating, API diff, pre-tag-checks, tag, proxy check);
R1a (first interleave batch per default sequencing); **P1–P9 preset train** (the actual
1%→51% lever); S1–S7 scaffold; G-CNT adoption counter; A1–A5 recipes; D3a–c sqlitekit;
D1a–b worker; C2R realtime folded-broadcast; R1b–e/R2–R5a repairs; H1–H7 httpx; X1–X4
compose/TLS/testkit/v1-exit; Y1–Y7; OG1–6 owner memos; Z1–Z7 process debts; final
verify-all + release-state sync (AGENTS/FEATURES/TODO_LIST headers).

## d) TOTALLY FUCKED UP (own mistakes, all caught & corrected forward)

1. **`strings.Cut` misuse (real bug, ~20 min):** I used its third return as "tag present"
   — it means "separator found". Every env-tagged field was silently skipped; tests
   caught it, a reflection probe localized it, `Tag.Lookup` fixed it.
2. **Autofix introduced a semantic bug (the important lesson):** `golangci-lint --fix`'s
   `modernize` rewrote my `FieldErrorsOf` type-switch into an `errors.As` form that
   DESCENDS into `FieldError.Unwrap` and silently drops sibling errors of a join —
   multi-error loads would have reported only the first field problem. Caught by reading
   the diff, rewritten as an exact-node walk. **Lesson: lint autofixes on hand-rolled
   error-tree walkers are semantic changes; review them like code review, not formatting.**
3. **First `FieldErrorsOf` draft was unusable as written** (`errors.As` into a slice is
   not a thing) and shipped dead test scaffolding (unused struct + copy-paste collect
   hack) — replaced by the library helper + clean tests.
4. **process miss:** ran `golangci-lint run --fix` directly; the buildflow skill's
   default is `buildflow --fix` (project AGENTS documents bare per-module golangci-lint
   as the _verdict_ command; for fixing, BuildFlow owns the loop). Scoped to one new
   module and re-verified by tests — deviation noted, not repeated for the next modules.
5. **`rg -rn` hook block (third strike on this host):** the rg-replace-shaped trap again;
   corrected to plain `rg -n`. Also one wrong skill path guess and a sed that mis-ordered
   an import block (fixed via python) — sloppiness class, zero residue.
6. **Design wobble on nested keys:** my first test encoded absolute child keys
   (`NESTED_ADDR` under `env:"NESTED"` → would produce `NESTED_NESTED_ADDR`); resolved to
   the design-note semantics (child tags are relative suffixes). Cost: one test rewrite.

## e) WHAT WE SHOULD IMPROVE

1. **New-module checklist** (this session invented it the hard way; encode it in
   module-details or AGENTS): go.mod floor + LICENSE + satellite `.golangci.yml` +
   charter `--fix` + dependabot entry + **ci.yml matrix slot** + module README
   "Build & verify" + CHANGELOG. The ci.yml slot is the one the guards do not enforce.
2. **Guard gap:** extend `check-dependabot-parity.sh` (or the matrix job) to require a CI
   matrix slot for every go.work member — same mechanical-truth family as the charter
   guard; would have caught (2b) at commit time.
3. **tagalign ordering decision:** the linter wants `default:"…"` before `env:"…"`
   (alphabetical). Decide once, apply to code+README+design note together (contract-string
   hygiene rule).
4. Keep the **probe-test debugging pattern** (tiny in-package reflection probes, delete
   after) — it localized the Cut bug in minutes vs theorizing.

## f) NEXT — up to 50, in execution order (from the unified plan)

1. Finish config lint 0-issues (varnamelen/tagalign/golines/errcheck/nolintlint batch)
2. Add `./config` to ci.yml matrix; actionlint; confirm parity guard
3. tagalign order decision + README/design-note alignment
4. Scratch-module compile check of README snippets (C4 acceptance)
5. `config/docs` module-details row + AGENTS release-state pointer (module count 12→13)
6. C5 release train: CHANGELOG date, API diff additions-only, pre-tag-checks, tag
   config/v0.1.0, fresh-consumer proxy check
7. R1a interleave batch: fix httputil.md rg command
8. R1a: verify-blocks libs 1–3 executed as written, failures filed
9. R1a: verify-blocks libs 4–6
10. P1: preset design — enumerate spine pieces + config surfaces
11. P1: env-switch grammar table freeze (APPKIT_OTEL/SECURITY/REALTIME/ERRORPAGES…)
12. P1: preset.Config struct + options sketch in README
13. P1: posture table per environment (dev/prod/systemd)
14. P1: conflict-matrix test plan (env present + code override + absent)
15. P1: skeleton module + charter + dependabot + ci.yml slot (checklist from e1)
16. P2: wire core Service construction + Run/Shutdown passthrough
17. P2: systemd.Install + Type=notify harness test
18. P2: health dashboard mount + probe wiring
19. P2: test spine boot → /health/ready 200
20. P3: errorpages Mount + pretty-404 test
21. P3: metrics/version + /version test
22. P3: RegisterClassifications + idempotency test
23. ~~P3: full-spine-boots-with-zero-app-code test~~ done at `e26b7e5`
24. P4: otel leg + APPKIT_OTEL parse; env-absent → no otel deps active
25. P4: env-present → spans + log bridge; explicit WithOTEL overrides env
26. P5: security profile enum none/standard/strict matrix test
27. P5: standard profile through security battery; opt-out removes leg cleanly
28. P6: realtime leg Mount wiring + stream smoke test
29. ~~P7: example app main ≤40 LOC + conflict-matrix suite~~ done at `6c74730`
30. P7: boot probe + graceful-stop probe
31. P8: integration E2E skeleton (pins PUBLISHED tags — needs C5+P9 tags)
32. P8: E2E preset boot → health → drain through tags; env legs on/off
33. P9: preset CHANGELOG dating + release-state AGENTS/FEATURES sync
34. P9: tag preset/v0.1.0 + pre-tag-checks + consumer heads-up drafts
35. S1: scaffold verdict doc (template repo vs generator; G3 default if unanswered)
36. S2–S6: template pieces (cmdguard serve verb, templ skeleton, flake/buildflow,
    CI/systemd/Dockerfile, doc skeletons + boot probe)
37. S7: dogfood one real app end-to-end
38. G-CNT: adoption-counter script (dual any/direct) + gate table mode + guard wiring
39. A1–A5: recipe template + dnsblockd/ChastityAPI/PapDashboard/pbx-lens recipes
40. D3a–c: sqlitekit design/lease matrix/backup/ledger/migration helper **→ TODO P3 — Battery program waves W3-W5 (W4 sqlite ops kit: lease/backup/ledger)** <!-- fresh:skip: routed in TODO_LIST -->
41. D1a–b: worker supervisor + pools + panic-restart + shutdown joins
42. C2R: realtime FoldedSource + SubscribeFolded discriminator test **→ TODO P3 — Battery program waves W3-W5 (W5 C2 projection→broadcast folded contract)** <!-- fresh:skip: routed in TODO_LIST -->
43. R1b–e: verify-blocks 7–12, ST bench baseline, hub doc note, MVS proof
44. R2–R5a: deps-guard CI, error-family batch, ST recipe, httpx shortlist verdicts
45. H1–H7: httpx module program (B1→B7 per plan order)
46. X1–X4: compose-on-httputil refactor, TLS, testkit migration, v1-exit close-out
47. Y1–Y7 + Z1–Z7: backlog + process debts (incl. the lessons entry from d2)
48. OG1/OG3–OG6 memo drafts; OG2 heads-ups post when core tags
49. Final: `scripts/verify-all.sh` all modules + full guard suite + release-state sync
50. Final: CI green on final push; session-close ritual (every new job ran live)

## g) QUESTIONS FOR THE OWNER (cannot be answered from the repo)

1. **G1 (asked 3×, still the top blocker):** v11 repair backlog vs v12 spine — I am
   executing the plan's DEFAULT (C-train → one R1 batch → P-train → S-train → waves).
   Confirm, or reorder?
2. **Tag & push policy for this execution session:** the plan's C5/P9/P8 steps end in
   annotated tags + `git push origin master <tag>` (release ritual). My standing rule is
   never-push-without-ask. May I push master+tags to origin as each release task
   completes, or hold ALL tags/pushes for one owner-reviewed batch at session end?
3. **Scaffold shape (G3) pre-answer:** when I reach S1, the recommendation is template
   repo (go-website-template pattern; go-wizard-sdk's engine is unbuilt and I won't block
   on it). Rule now, or accept the default when reached?

---

**Verification state at write time:** config tests green `-race -count=1` hermetic;
guards green (directives, dependabot-parity, workspace-charter); lint 25 findings open
(C-train item 1 above); `./config` missing from ci.yml matrix (item 2); working tree
auto-committed by pma through `92042eb`.
