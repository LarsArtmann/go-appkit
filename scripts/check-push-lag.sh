#!/usr/bin/env bash
# Push-lag watch (SUPERB plan T05). Damage modes that motivated it:
#   - a ~30-min push stall (22:47Z->23:16Z, 2026-10-06) went completely
#     unnoticed while the session continued;
#   - the 6-day CI-blindness window (2026-10-01..07) proved that "local
#     green" without "origin green" is a silent failure mode.
#
# Fails when commits exist on the local branch that are NOT on origin and
# the OLDEST of them is older than the threshold (default 6h). Uses the
# last-known origin state (refresh with git fetch first if you need
# freshness); with a stale remote ref it can only over-report lag, never
# hide it.
#
#   ./scripts/check-push-lag.sh                     # 6h default
#   PUSH_LAG_THRESHOLD_HOURS=1 ./scripts/check-push-lag.sh
#
# Wired into the pre-commit project-guards block as a WARNING (never a
# commit blocker) and runnable standalone / from any guards battery.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

branch="${PUSH_LAG_BRANCH:-master}"
remote_ref="refs/remotes/origin/$branch"
threshold_hours="${PUSH_LAG_THRESHOLD_HOURS:-6}"

if ! git rev-parse --verify --quiet "$remote_ref" >/dev/null; then
	echo "OK: no origin/$branch ref — push-lag watch skipped (local-only clone?)"
	exit 0
fi

count="$(git rev-list --count "$remote_ref..HEAD" 2>/dev/null || echo 0)"
if [[ "$count" -eq 0 ]]; then
	echo "OK: $branch fully pushed (local == origin)"
	exit 0
fi

# git log lists newest first; the last line is the oldest unpushed commit.
oldest_ct="$(git log "$remote_ref..HEAD" --format=%ct | tail -n1)"
now="$(date +%s)"
age_hours=$(((now - oldest_ct) / 3600))

if ((age_hours >= threshold_hours)); then
	echo "FAIL: $count commit(s) on $branch unpushed for ${age_hours}h (threshold ${threshold_hours}h) — push stalled? Run: git push origin $branch"
	exit 1
fi

echo "OK: $count unpushed commit(s) on $branch, oldest ${age_hours}h (< ${threshold_hours}h threshold)"
