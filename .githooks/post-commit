#!/usr/bin/env bash
# HEAD-advance sanity watch (SUPERB plan T06) — post-commit hook.
#
# Damage class: a reflog-SILENT commit rewind (bb47fbc vanished mid-session
# 2026-10-06; daemon prime suspect) plus the daemon's tree-mutation bundling.
# A post-commit hook sees every commit the daemon or a human makes, so it is
# the one observation point neither can bypass without plumbing-level ref
# surgery.
#
# After each commit it verifies the PREVIOUSLY recorded HEAD is still part
# of the history graph, by five discrimination rules (in order — legitimate
# operations pass silently, everything else gets a loud recovery hint):
#
#   1. prev is an ancestor of the new HEAD  -> normal advance or merge
#   2. the new HEAD's parent IS prev        -> amend
#   3. prev == ORIG_HEAD                    -> deliberate reset/rebase (git
#                                             left its own breadcrumb)
#   4. prev is reachable from any branch    -> branch work / kept-alive ref
#   5. otherwise                            -> LOUD warning + recovery hint
#
# WARNING only — it never blocks a commit (a blocked hook would just teach
# the daemon to bypass hooks). The state file doubles as recovery evidence:
# the vanished sha is recorded in $GIT_DIR/head-advance.last.
#
# Install (canonical source lives here; .git/hooks is not tracked):
#   cp scripts/post-commit-head-watch.sh .git/hooks/post-commit
#   chmod +x .git/hooks/post-commit

set -u

git_dir="$(git rev-parse --git-dir)"
state="$git_dir/head-advance.last"
new="$(git rev-parse HEAD)"

if [[ ! -f "$state" ]]; then
	printf '%s\n' "$new" >"$state"
	exit 0
fi

prev="$(cat "$state" 2>/dev/null || true)"

if [[ -z "$prev" || "$prev" == "$new" ]]; then
	printf '%s\n' "$new" >"$state"
	exit 0
fi

ok=0
reason=""

if git merge-base --is-ancestor "$prev" "$new" 2>/dev/null; then
	ok=1
	reason="ancestor (normal advance/merge)"
elif [[ "$(git rev-parse "$new^" 2>/dev/null || true)" == "$prev" ]]; then
	ok=1
	reason="amend"
elif [[ "$prev" == "$(git rev-parse ORIG_HEAD 2>/dev/null || true)" ]]; then
	ok=1
	reason="ORIG_HEAD (deliberate reset/rebase)"
elif git branch --all --contains "$prev" 2>/dev/null | grep -q .; then
	ok=1
	reason="still on a branch"
fi

if ((ok)); then
	printf '%s\n' "$new" >"$state"
	exit 0
fi

cat >&2 <<WARN
post-commit WARNING: HEAD-advance sanity FAILED.
Previous HEAD $prev is neither an ancestor of the new HEAD, nor an amend
of it, nor ORIG_HEAD, nor on any branch — this matches the reflog-silent
rewind damage class (bb47fbc, 2026-10-06). If this was NOT intentional:

    git branch recover-rewind $prev

Investigate BEFORE continuing (the recorded sha above is the vanished tip).
WARN
printf '%s\n' "$new" >"$state"
exit 0
