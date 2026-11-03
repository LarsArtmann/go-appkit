# go-atomic-write — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). First dedicated
> report; the adoption path is battery D4 in the canonical battery spec.

## Verdict

**NOT USED — and correctly so: the framework is stateless HTTP. Zero imports, zero graph
presence (not even indirect) across all 12 modules. The library is already NAMED as the
floor for battery D4 (atomic file write, "floor: go-atomic-write ≥ v0.5.1, Windows syscall
trap") in the canonical battery spec and TODO_LIST — the only plausible adoption surfaces
are D4, the F1/F4 config modules, and the W4 sqlite ops kit, all demand-gated P3 work.**

| Fact           | Value                                                                                                                                                     | Evidence                                                             |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| Direct imports | 0                                                                                                                                                         | `rg 'go-atomic-write' --glob '*.go'` → empty                         |
| In graph?      | No — absent from every go.mod and go.sum                                                                                                                  | `rg 'go-atomic-write' --glob 'go*.sum'` → empty                      |
| Latest tag     | v0.6.0                                                                                                                                                    | `git -C ~/projects/go-atomic-write tag --sort=-v:refname \| head -1` |
| Battery floor  | "≥ v0.5.1 (Windows trap)" — v0.6.0 satisfies it                                                                                                           | battery spec line 65, TODO_LIST:50                                   |
| Upstream state | HEAD 17 commits past v0.6.0; the go-1.27 directive bump sits in their `[Unreleased]` (their CHANGELOG has no 0.6.0 section yet — their hygiene, not ours) | `~/projects/go-atomic-write/CHANGELOG.md` head                       |

## 1. Capability inventory

Crash-safe, race-free file writes (`~/projects/go-atomic-write/`, single package):

- **Three failure modes of `os.WriteFile` addressed**: lost updates (cross-process
  `flock(2)`/`LockFileEx` locking), partial writes (unique temp file + fsync + atomic
  rename, dir-fsync on POSIX), interleaved writers (locking + TOCTOU detection).
- **TOCTOU detection**: xxhash64 `Fingerprint` type + `WriteVerified(path, data, fp)` →
  `ErrConcurrentModification` sentinel; `FingerprintFile`/`FingerprintFromBytes`.
- **Variants**: `Write`, `WriteWithPerm`, `WriteIfChanged` (no mtime bump when identical),
  `WriteFunc`/`WriteFuncVerified` (streaming writer callback — no Reader API by design).
- **Windows handling**: `MoveFileEx` retried on sharing violations (antivirus contention) —
  exactly the trap that broke `GOOS=windows` builds pre-v0.5.1, hence the documented floor.
- Honest limits: fingerprint detects changes, not attackers (non-cryptographic hash);
  directory fsync is POSIX-only.

## 2. Capability vs applicability

| Capability                     | Status                         | Reason                                                                                                                             |
| ------------------------------ | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| Any file write in appkit today | ⚪ N/A                         | No module persists files: no config-file writes, no state files, no ledger — the framework is an in-memory HTTP service by charter |
| Battery D4 (atomic file write) | ⚪ N/A today — floor pre-named | Battery spec + TODO_LIST already route it: wrap/promote go-atomic-write with the ≥ v0.5.1 floor in code comments; demand-gated     |
| F1/F4 config modules           | ⚪ N/A today                   | Config hot-reload/persistence is P3 demand-gated; if it lands, `WriteFuncVerified` is the natural persistence primitive            |
| Version currency               | 🟢                             | v0.6.0 is the newest tag; floor satisfied with headroom                                                                            |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **Nothing writes files.** An unused require would be dead weight against the
   dependency-lean posture (go-atomic-write itself carries xxhash + flock deps).
2. **The adoption decision is already made correctly upstream-of-code**: battery D4 names
   the library and the version floor, so when the battery is built, the only question left
   is the wrapper shape — the trap that motivates the floor (Windows sharing violation) is
   fixed in v0.6.0.
3. **When it flips:** D4 or a config/state-writing battery lands → import directly
   (`atomicwrite.WriteFuncVerified` for streaming, `WriteIfChanged` for config rewrites),
   keep the floor comment, and add a Windows CI leg for whichever module carries it.

## 4. Recommendations

1. Keep the D4 routing as-is — re-verified current (TODO_LIST:50, battery spec line 65/382).
2. At adoption time, bump the documented floor from "≥ v0.5.1" to "≥ v0.6.0" (go-1.27 fleet
   alignment) unless a consumer constraint forbids it.
3. Upstream hygiene note (their repo): the v0.6.0 CHANGELOG section is missing — if we
   adopt, verify their release notes before pinning (same discipline as the
   go-idempotency note).

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'go-atomic-write|atomicwrite' --glob '*.go' --glob 'go*.sum'   # empty → absent
rg -n 'go-atomic-write' TODO_LIST.md docs/feedback/processed/2026-09-04_batteries-included-sdk-gap-analysis.md
git -C /home/lars/projects/go-atomic-write tag --sort=-v:refname | head -1   # v0.6.0
```
