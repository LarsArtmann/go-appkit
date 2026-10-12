## TL;DR

Editing a tool's config file (e.g. `.jscpd.json`) does not invalidate its cached detector result — the next run replays the old verdict. Ask: fold each tool's config-file identity into the result-cache key.

## Problem

The result cache keys on `{ToolName, ContentHash, ToolHash}` (`execution/result_cache.go:31-41`, @ 8a349283a). `ToolHash` is the resolved binary's identity (name + path + size/mtime, `execution/result_cache_3.go:161-172`); `ContentHash` covers the tool's matched input files. A config file that changes tool behavior is in neither.

The jscpd provider is the concrete case: it reads `.jscpd.json` and passes `--config` when present (`tools/providers/jscpd_provider.go:103-105`), but `.jscpd.json` is not in the tool's `FileContract.Inputs` (only the source-code trigger patterns are), so editing it changes no key material.

## Live reproduction (go-appkit, 2026-10-12)

1. `-s jscpd` reported 2 error-severity findings (duplicated JSON dashboard blocks).
2. Added `**/dashboards/**` to the repo's `.jscpd.json` `ignore` list.
3. `-s jscpd` again → same 2 error findings, gate still red.
4. `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s jscpd` → clean pass. The fresh scan also diverged wildly from the cached one (107 findings vs the cached run's 16) — the replay was masking the real result set, not just the 2 errors.

## Existing alternatives, checked

- `--no-result-cache-for jscpd` works but drops caching for the tool entirely — paying a full rescan on every run to fix one stale-config class.
- The env kill-switch has the same cost profile.
- Config edits are precisely when a fresh verdict matters, so TTL waiting is not a mitigation either.

## Ask

Hash each tool's config file(s) into `ToolHash` (or prepend them to `ContentHash`). A generic mechanism reading the tool's declared config paths would fix this class for every provider that shells out with a `--config`, not just jscpd.

💘 Generated with Crush
