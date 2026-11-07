# go-idempotency — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). Supersedes the
> go-idempotency section of `docs/research/2026-10-11_sibling-library-utilization-audit.html`
> (newer and narrower; that audit's findings are re-verified as still-current).

## Verdict

**NOT USED — deliberately, via the documented opt-in posture. Zero imports across all 12
modules; the library rides as an indirect (v0.3.1 = latest tag) in cqrs and integration
because go-cqrs-lite/middleware/v4 is built on it. cqrs documents the posture at both the
code and README level, and the battery-D7 adopt-vs-port annotation from the 2026-10-11
audit is intact in TODO_LIST. The one cheap deliverable that audit recommended — a README
wiring recipe — is still not written and remains gated on demand.**

| Fact             | Value                                                                   | Evidence                                    |
| ---------------- | ----------------------------------------------------------------------- | ------------------------------------------- |
| Direct imports   | 0 of 12 modules                                                         | `rg 'go-idempotency' --glob '*.go'` → empty |
| In-graph version | v0.3.1 (= latest tag), indirect                                         | `cqrs/go.mod:64`, `integration/go.mod:109`  |
| Chain            | cqrs → go-cqrs-lite/middleware/v4 → go-idempotency                      | `go mod why -m` (verified 2026-11-03)       |
| Posture records  | `cqrs/commands.go:24` + `cqrs/README.md:618` + TODO_LIST:50/52          | consistent across all three                 |
| Upstream state   | HEAD 13 commits past v0.3.1 (release-hardening planning; no API change) | `~/projects/go-idempotency`                 |

## 1. Capability inventory

A focused at-most-once SDK (`~/projects/go-idempotency/`):

- **Store contract** (`store.go:54-94`): three methods, with the atomicity guarantee resting
  on a single atomic `CheckAndRecord` (`ErrDuplicate` on live claim, Conflict/409-class;
  `ErrInvalidTTL` rejection). The README explicitly warns off the racy Seen-then-Record
  pattern (README:29-40).
- **middleware** package: `NewCommand(store, ttl, execute)` claim-then-execute wrapper, and
  an HTTP middleware keyed on the `Idempotency-Key` header (missing → 400, duplicate → 409,
  store failure → 503 fail-closed).
- **contract** package: a reusable conformance suite (`RunTests`/`RunTestsStrict`/
  `RunTestsContextAware` for network backends, incl. a 200-goroutine atomicity test and
  negative tests proving the suite catches broken stores) — the tool a consumer uses to
  certify their own Redis/SQL backend.
- **No production backends by design**: Redis/SQL are consumer-side adapters; go-cqrs-lite
  users get `metaengine.DedupStore` instead (README:53). MemoryStore is deprecated for
  production (removal at next major).

## 2. Capability vs applicability

| Capability                                | Status                                | Reason                                                                                                                                                                                                                                                                          |
| ----------------------------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Store + `CheckAndRecord` in cqrs commands | ⚪ N/A today — adoption path recorded | cqrs's EventConfig deliberately excludes idempotency from defaults (changes command semantics, `cqrs/commands.go:24`); TODO_LIST:52 names the cheap path: `EventConfig.CommandMiddleware` + `CommandIdempotency` from go-cqrs-lite/middleware — no new appkit dependency needed |
| HTTP `Idempotency-Key` middleware         | ⚪ N/A today                          | Belongs to the W3 httpx battery if a consumer asks; not a default-stack concern                                                                                                                                                                                                 |
| contract suite                            | ⚪ N/A today                          | Activates when a consumer implements a Store backend (battery D7's likely final shape: budget the backend, not a parallel store)                                                                                                                                                |
| Version currency                          | 🟢                                    | v0.3.1 everywhere = newest published tag                                                                                                                                                                                                                                        |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **Idempotency changes command semantics** — it is opt-in by charter, not a default
   (`cqrs/commands.go:24`). Zero cqrs+realtime composition consumers exist (demand
   re-checked 2026-09-29), so there is no demand signal to gate on.
2. **The wiring needs zero appkit code.** go-cqrs-lite/middleware already wraps the SDK;
   a consumer appends `CommandIdempotency` via `EventConfig.CommandMiddleware`. Adoption is
   a documentation act, not a dependency act.
3. **When it flips:** a real consumer needs at-most-once command processing → write the
   README cookbook paragraph FIRST (import path, Store backend via battery D7's budget,
   `contract.RunTests` certification, `EventConfig.CommandMiddleware` append), keep
   go-idempotency itself indirect. The 2026-10-11 finding stands: do NOT port CV's 165-line
   store — the SDK plus a backend covers it.

## 4. Recommendations

1. Keep the TODO_LIST annotations (D7 adopt-vs-port, cqrs EventConfig candidates) — both
   re-verified current.
2. When idempotency demand materializes, the recipe paragraph is the whole deliverable
   (compile-checked example, no code change, no new require).
3. Upstream hygiene note (their repo, not ours): v0.3.1 shipped without its CHANGELOG
   section (content sits under `[Unreleased]`) — if go-appkit ever promotes the dep to
   direct, re-check their release state first.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'go-idempotency' --glob '*.go'                     # empty → zero imports
grep 'go-idempotency' cqrs/go.mod integration/go.mod     # v0.3.2 / v0.3.1 indirect
cd cqrs && GOWORK=off GOTOOLCHAIN=go1.27.3 go mod why -m github.com/larsartmann/go-idempotency
sed -n '22,26p' cqrs/commands.go; sed -n '616,620p' cqrs/README.md
git -C /home/lars/projects/go-idempotency tag --sort=-v:refname | head -1   # v0.3.2
```

## Errata (2026-11-06, re-run of the verify block)

- Block verified as written; only version drift: go-idempotency released
  **v0.3.2** and cqrs carries it (integration still pins v0.3.1 — member
  go.mod changes ride release trains; MVS resolves both fine).
