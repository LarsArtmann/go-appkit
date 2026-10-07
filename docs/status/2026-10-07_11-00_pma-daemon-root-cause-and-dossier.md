# pma Daemon — Root Cause Located, Evidence Dossier, Owner Triage

> 2026-10-07 11:00 CEST · SUPERB plan T09 (docs/status snapshot; release
> facts live in AGENTS, task state in TODO_LIST).
> **Open question 6 is ANSWERED.** The daemon is located, its config read,
> its source repo identified, and its behavior boundaries established.

## 1) Where the daemon lives (the answer)

| What            | Where / value                                                                                       |
| --------------- | --------------------------------------------------------------------------------------------------- |
| Unit            | `/etc/systemd/system/projects-management-automation.service` (SYSTEM unit, `User=lars`, `SyslogIdentifier=pma`) |
| Source          | https://github.com/LarsArtmann/projects-management-automation (local clone: `~/projects/projects-management-automation`, binary rev 8f430ae) |
| Config          | `/etc/projects-management-automation/service.yaml`                                                   |
| Logs            | journal, query with `journalctl -t pma` (owner; `systemctl` is blocked for agents)                    |
| Health endpoint | `127.0.0.1:9190` (`PMA_HEALTH_LISTEN_ADDR`)                                                           |
| State           | `/var/lib/pma/cooldown-state.json`                                                                   |

Why agents never saw it: it is a system-level unit (not `~/.config/systemd/user/`),
`systemctl` is blocked in agent shells, and its process runs with a nix-store
PATH — `ps` greps for "commit"/"pma" from agent shells found nothing (name is
`projects-management-automation`).

## 2) Config facts (from service.yaml)

- Watches **all of `/home/lars/projects`** (excludes: `forks/`, `archived/`).
- `commit_strategy: on-change`, `debounce_duration: 60s`,
  `min_commit_interval: 120s` — matches the observed 2–5 min cadence and the
  ≤60 s race window that stole five staged file sets from the agent session
  below.
- `auto_stage: true` — ANY working-tree change gets staged, including changes
  another actor is mid-edit on.
- `auto_push: false` — pushes are human/agent; the 30-min push stall class is
  not pma pushing.
- `conflict_strategy: abort` — the daemon aborts on conflicts, it does not
  pick sides.
- Commit messages: AI-first via a LOCAL model (`qwen3.6-moe:35b-a3b` at
  `127.0.0.1:52625`), falling back to the "chore: auto-commit N changed
  file(s) (heuristic)" template when the AI is unavailable.

## 3) What pma IS and IS NOT (source-verified)

- **IS:** watcher + auto-stager + committer. Its only file writes are its own
  state (cooldown, pidfile, config) — `rg 'WriteFile' pma-daemon/` confirms.
  It never edits project content.
- **IS NOT:** the mutator. No `go work use` / go.work handling exists anywhere
  in its source. The `./integration` go.work re-adds (5 leaks now) are made by
  a DIFFERENT tool — the one-line sorted insert matches the signature of an
  editor/gopls-class "sync workspace" action. pma then bundles that mutation
  into a heuristic commit within 60 s, which is why attribution windows die.
- **The rewind (`bb47fbc`)** remains unattributed. pma's conflict strategy is
  abort-only and its committer carries no reset/rebase/update-ref calls; the
  remaining candidates are a `git add` racing an agent `reset` (pma staged the
  pre-reset index state) or another editor action. The post-commit HEAD-advance
  watch (shipped today, `.githooks/post-commit`) catches a recurrence with a
  ready-made recovery command.

## 4) Critical finding: pma DISABLES ALL GIT HOOKS BY DESIGN

The unit sets:

```
GIT_CONFIG_COUNT=1
GIT_CONFIG_KEY_0=core.hooksPath
GIT_CONFIG_VALUE_0=/var/empty
```

So the daemon commits with hooks pointed at an empty directory. Implications:

- The newly-live `.githooks/` guards (SUPERB T05–T07) protect HUMAN and AGENT
  commits only. Daemon commits bypass them by construction.
- Daemon-introduced drift therefore remains covered ONLY by the CI-side
  guards (pin-drift, workspace-charter, go-directives, dependabot-parity) —
  which is exactly where every leak was caught in practice.
- This also retroactively explains why the old `.git/hooks/pre-commit` never
  fired for daemon commits (double-inert: wrong path for everyone, and
  hooksPath=/var/empty for pma).

## 5) Damage-mode ledger (hashes + timestamps)

| # | Event | Evidence |
| - | ----- | -------- |
| 1 | go.work `./integration` re-add, leak #3 | caught by charter guard; fixed `ae3b2d4` 2026-10-06 |
| 2 | go.work re-add, leak #4 | `620a3b1` 2026-10-07 09:53:48; fixed `19100c5` |
| 3 | go.work re-add, leak #5 — bundled with the session report file | `0cfc4ab` 2026-10-07 10:12:20; fixed `103a7ad`; charter guard red on origin run 37592301037 |
| 4 | Staged-file theft from an active agent session (5× in ~30 min) | `1da7bfb` 10:22 (check-pin-drift.sh), `535636e` 10:26 (core CHANGELOG), `6347ba5` 10:33 (systemd module + AGENTS), `80fb3da` 10:42 (integration leg), `b80c8bc` 10:50 (.buildflow.yml) — in every case the content survived (the daemon committed it), but the intended commit message and attribution were lost |
| 5 | Reflog-silent commit rewind | `bb47fbc` vanished 2026-10-06; content recovered from index; post-commit watch now guards the class |
| 6 | ~30-min push stall unnoticed | 22:47Z→23:16Z 2026-10-06; `scripts/check-push-lag.sh` now guards (warning in pre-commit) |
| 7 | `.buildflow.yml` mutations (license-check skip + GOTOOLCHAIN pin removal) | `620a3b1`/`31526ce` + staged 10:50; probed 2026-10-07: BOTH were correct modernizations (license-check passes; toolchain auto-derives) — accepted in `86b612d` |

Pattern: pma never destroyed content. Every damage mode is a RACE or an
ATTRIBUTION loss: it stages mid-flight edits, bundles unrelated mutations,
and overwrites intended commit messages.

## 6) Owner triage checklist (decisions only pma's owner can make)

1. **Hook bypass:** keep `core.hooksPath=/var/empty` (hermetic committer) or
   allow an opt-in per-repo hooks mode? Today the repo's commit-time guards
   cannot see daemon commits at all.
2. **Repo scoping:** should go-appkit (a repo with release trains + tags +
   guards) get a per-repo exclude/cooldown (`exclude_paths` covers only
   `forks/`+`archived/` today), or is CI-side-only coverage accepted?
3. **Attribution:** the AI/fallback messages overwrite intent. An opt-in
   "respect staged message" or "skip repos with staged changes" mode would
   have prevented damage mode 4 entirely.
4. **go.work mutator:** the leak AUTHOR is still unknown (editor/gopls-class
   tool). `journalctl -t pma` around 09:53 / 10:12 will show pma's file-change
   event stream and may name the process that touched go.work first.

## 7) Follow-ups wired today (context for the decisions above)

- `.githooks/pre-commit` + `post-commit` live and verified (guards blocked a
  planted directive drift live; HEAD-advance watch passed a reflog-silent
  rewind simulation).
- `scripts/check-push-lag.sh` (T05) and the per-module pin-drift extension
  (T02, first live run caught the missed docs v0.3.2 claim).
- Charter guard caught leaks #3/#4/#5 — the go.work leak class is now
  mechanically detected within one push.
