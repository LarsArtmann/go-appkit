# go-retry — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). First dedicated
> report; the 2026-10-11 sibling audit did not cover go-retry.

## Verdict

**NOT USED — deliberately. Zero imports across all 12 modules; the library rides as an
indirect (cqrs pins v0.9.0 = latest tag; integration carries a v0.8.0 floor through the
published cqrs tag chain) because go-cqrs-lite/middleware/v4 builds its CQRS retry
middleware on it. The framework itself has no retryable outbound path — its lifecycle
primitives (graceful drain, deadlines) are deadline-based, not retry-based — and retry for
commands is consumer domain by charter. Natural adoption surface if/when the W4 `polite`
outbound-client battery is built.**

| Fact              | Value                                                                                                                                                                                                                                                                                                                                               | Evidence                                                                    |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| Direct imports    | 0 of 12 modules                                                                                                                                                                                                                                                                                                                                     | `rg 'go-retry' --glob '*.go'` → empty                                       |
| In-graph versions | cqrs v0.9.0 (= latest, working tree); integration resolves **v0.8.0** — CORRECTED 2026-11-03: `go list -m all` in integration proves its published-tag chain tops out at v0.8.0 (the earlier "repo build union resolves v0.9.0" claim was wrong — there is no cross-module union for a standalone module build; each module resolves its own graph) | `cqrs/go.mod:65`, `integration/go.mod:110`, `go list -m all` run 2026-11-03 |
| Chain             | cqrs → go-cqrs-lite/middleware/v4 → go-retry                                                                                                                                                                                                                                                                                                        | `go mod why -m` (verified 2026-11-03, both cqrs and integration)            |
| Upstream state    | v0.9.0 (2026-10-27); HEAD 3 commits ahead (lint-config chores, [Unreleased] empty)                                                                                                                                                                                                                                                                  | `~/projects/go-retry`                                                       |

## 1. Capability inventory

A dependency-light retry loop over go-error-family (`~/projects/go-retry/`, single package):

- `Do(ctx, cfg, fn, opts...)` / `DoWithValue[T]` — capped exponential backoff, additive
  jitter by default (`JitterNone` for determinism), `MaxAttempts` counts TOTAL attempts.
- **error-family integration is the differentiator:** `IsRetryable` defaults to
  `errorfamily.IsRetryable` (behavioral classification decides retryability), and
  `FromPolicy(errorfamily.RetryPolicy)` converts the advisory policy — the two libraries
  are designed as a pair.
- Options: `WithIsRetryable`, `WithDelayFunc` (e.g. HTTP `Retry-After`), `WithOnRetry`,
  `WithExhausted`, `WithJitter`, `WithRandomSource` (reproducible tests).
- Sentinels: `ErrExhausted` / `ErrCanceled` / `ErrDeadlineExceeded` (classified; detect via
  `errors.Is` — never a bare `context.Canceled`).
- Nesting is fail-closed (inner exhaustion is not retried by outer loops); invalid config
  fails as a Rejection before attempt 1; fuzzed delay computation.

## 2. Capability vs applicability

| Capability                                       | Status                                           | Reason                                                                                                                                 |
| ------------------------------------------------ | ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| Retry loops in core lifecycle                    | ⚪ N/A by design                                 | Shutdown/drain are deadline-driven (`DrainDelay`, `NoDrainDelay`); a framework retry loop would fight the operator's deadline contract |
| Retry in cqrs commands                           | ⚪ N/A — consumer domain, already wired upstream | go-cqrs-lite/middleware exposes retry over go-retry; cqrs documents it as a non-default (`cqrs/commands.go:24`, README:618)            |
| Outbound-call retry (future `polite` client, W4) | ⚪ N/A today — adoption path recorded            | The battery has no in-framework consumer yet; go-retry (+ `DelayFunc` for Retry-After) is the natural floor when it is designed        |
| Version currency                                 | 🟢                                               | v0.9.0 in cqrs = newest tag; integration floor self-heals at the next integration re-pin                                               |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **No retryable call sites exist in the framework.** Everything appkit owns is inbound
   HTTP lifecycle or delegation to go-cqrs-lite, which already carries the retry middleware.
2. **Sentinel-refusal discipline is already correct upstream:** retry is deliberately absent
   from the CQRS wrapper defaults, matching `cqrs/commands.go:24` — pulling go-retry into
   any appkit module today would be an unused require, the exact thing the battery program's
   anti-recommendations warn against.
3. **When it flips:** the W4 `polite` outbound client (timeouts, backoff, Retry-After) is
   the designed home. At design time, prefer `FromPolicy(errorfamily.RetryPolicy)` so
   consumer-supplied policies flow through, and keep go-retry OUT of core — it belongs in
   the opt-in module.

## 4. Recommendations

1. Keep the posture; nothing to annotate — no TODO row currently names go-retry, and none
   is needed until W4 `polite` enters planning.
2. When W4 `polite` is scoped, record "built on go-retry" in the battery row so the
   adopt-vs-port question is pre-answered (same pattern as the D7/F2 annotations).
3. Version hygiene: integration's v0.8.0 floor resolves to v0.9.0 via MVS in repo builds;
   the next integration re-pin train will tidy it — no action needed.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'larsartmann/go-retry' --glob '*.go'                # empty → zero imports
grep 'go-retry' cqrs/go.mod integration/go.mod            # v0.9.0 / v0.8.0 indirect
cd cqrs && GOWORK=off GOTOOLCHAIN=go1.27.2 go mod why -m github.com/larsartmann/go-retry
git -C /home/lars/projects/go-retry tag --sort=-v:refname | head -1   # v0.9.0
```
