# go-filewatcher — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). First dedicated
> report.

## Verdict

**NOT USED — and no adoption path is currently recorded, by design. Zero imports, zero
graph presence across all 12 modules. No appkit feature watches the filesystem: config is
in-process (`ServiceConfig`), the framework ships no hot-reload story, and no battery or
TODO row names file watching. The only plausible future seam is config hot-reload inside
the demand-gated F1/F4 config modules — and even there it would be an opt-in module, never
core, because the library's dependency weight (fsnotify, go-gitignore, gogenfilter/v3)
would break core's dependency-lean charter.**

| Fact           | Value                                                                                                                  | Evidence                                                            |
| -------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| Direct imports | 0                                                                                                                      | `rg 'go-filewatcher' --glob '*.go'` → empty                         |
| In graph?      | No — absent from every go.mod/go.sum                                                                                   | `rg 'go-filewatcher' --glob 'go*.sum'` → empty                      |
| Latest tag     | v2.4.2                                                                                                                 | `git -C ~/projects/go-filewatcher tag --sort=-v:refname \| head -1` |
| Module path    | **`github.com/larsartmann/go-filewatcher/v2` — the `/v2` suffix is REQUIRED**                                          | `~/projects/go-filewatcher/go.mod:1`, README:57                     |
| Upstream state | HEAD 42 commits past v2.4.2 ([Unreleased] empty; 2.4.x line: NFC normalization, case-sensitivity modes, watch budgets) | `~/projects/go-filewatcher/CHANGELOG.md`                            |

## 1. Capability inventory

A heavyweight, composable watcher over fsnotify (`~/projects/go-filewatcher/`):

- **Watcher**: recursive watching with depth limits, dynamic `Add`/`Remove`/`Reset`,
  watch-budget management (`WithMaxWatches` + safety fraction, Linux-only), ENOSPC polling
  fallback, self-healing of failed watches, case-sensitivity modes (NFC-normalized).
- **17+ filters** (extension/glob/regex/size/age/gitignore/generated-code) with AND/OR/NOT
  combinators; **18 middleware** (logging, recovery, rate limit, dedupe, batching,
  throttling, circuit breaker, exponential backoff, error correlation/sanitization).
- **Debouncing**: per-path or global; **observability**: Prometheus collector + OTel span
  middleware; phantom-typed path IDs; structured `WatcherError` taxonomy.
- Honest limits: inotify budget exhaustion degrades silently (counted in `Stats`);
  budgets are Linux-only; "other platforms retry nothing today" (README:319-321).

## 2. Capability vs applicability

| Capability                            | Status       | Reason                                                                                                                                  |
| ------------------------------------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Config hot-reload                     | ⚪ N/A today | No config-file loading exists to watch — `ServiceConfig` is code-constructed; F1/F4 config modules are unscheduled P3 demand-gated work |
| Dev workflows (docs/codegen watching) | ⚪ N/A       | Repo tooling lives in BuildFlow/scripts, not in shipped Go modules                                                                      |
| Any module                            | ⚪ N/A       | Zero demand signals in TODO_LIST, ROADMAP, or the battery spec (40 proposed batteries, none watch files)                                |
| Version currency                      | 🟢           | v2.4.2 is the newest tag (moot until adoption)                                                                                          |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **No feature needs it.** Unlike go-atomic-write (pre-named for D4) or go-etag (pre-named
   for B4), no battery or TODO row names file watching — adopting now would invent a
   requirement.
2. **Dependency weight vs charter.** Core's allowlist is deliberately lean; fsnotify +
   gitignore + gogenfilter in any default path would violate it. If a future config module
   wants hot-reload, the shape is a separate opt-in module (`appkit/configwatch` or inside
   the F4 module's own go.mod), never core.
3. **When it flips:** F1/F4 config modules get a real hot-reload requirement → evaluate
   then, importing `/v2` (the suffix is mandatory), starting from `WithDebounce` +
   `WithFilter(FilterExtensions…)` + `WithPolling(true)` for the portability story, and
   reviewing the Linux-only budget semantics against the target deployment matrix.

## 4. Recommendations

1. Nothing to track — deliberately no TODO row; this report is the record of the decision.
2. If config hot-reload ever enters planning, cite this file as the prior art and start
   from the opt-in-module shape, not a core dependency.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'filewatcher' --glob '*.go' --glob 'go*.sum' --glob 'go*.mod'   # empty → absent
rg -n 'hot.reload|watcher|inotify' TODO_LIST.md ROADMAP.md docs/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md   # no battery names it
head -1 /home/lars/projects/go-filewatcher/go.mod                     # /v2 module path
```
