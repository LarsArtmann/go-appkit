#!/usr/bin/env bash
# SSE-test protocol guard (17-49 f7, 2026-10-18): realtime.Handler flushes
# response headers BEFORE it subscribes to the hub (realtime/handler.go),
# so a test that broadcasts right after connect() races the subscribe step.
# A dropped broadcast then hangs ssetest.MustReadNEvents until the 10m
# panic — the 2026-10-13 rate-limit SSE E2E incident class (fixed in
# dd85808). House protocol (realtime/README.md, Limitations): sync on
# hub.SubscriberCount() before broadcasting.
#
# This guard fails on any .Broadcast( call in a _test.go file with no
# subscription sync within the preceding WINDOW lines. Recognized syncs:
#   - SubscriberCount(    — the poll itself
#   - waitForSubscriber(  — realtime's test helper (wraps SubscriberCount)
#   - .Subscribe(         — direct synchronous channel subscribe (no race)
# Comment text does not count: lines are comment-stripped before matching,
# so a "// SubscriberCount is the sync point" comment cannot satisfy the
# guard for an unsynced call.
#
# Usage: scripts/check-sse-protocol.sh [dir]   (default: repo root)
# The optional dir exists so the guard can be exercised against fixtures
# without touching the tree.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
scan_dir="${1:-$root}"

if [[ ! -d "$scan_dir" ]]; then
	echo "FAIL: scan dir '$scan_dir' does not exist" >&2
	exit 1
fi

window=40
status=0
files=0
calls=0

while IFS= read -r -d '' file; do
	files=$((files + 1))
	while IFS= read -r call_line; do
		calls=$((calls + 1))
		from=$((call_line - window))
		(( from < 1 )) && from=1
		ctx="$(sed -n "${from},$((call_line - 1))p" "$file" | sed 's|//.*||')"
		if ! grep -qE 'SubscriberCount\(|waitForSubscriber\(|\.Subscribe\(' <<<"$ctx"; then
			echo "FAIL: $file:$call_line broadcasts without a subscription sync in the preceding $window lines"
			echo "      sync first (poll hub.SubscriberCount(), or subscribe the channel directly) — see realtime/README.md Limitations"
			status=1
		fi
	done < <(grep -n '\.Broadcast(' "$file" | cut -d: -f1 || true)
done < <(find "$scan_dir" -name '*_test.go' -type f -print0 | sort -z)

if ((status == 0)); then
	echo "OK: $calls broadcast call(s) across $files test file(s) all carry a subscription sync"
fi
exit "$status"
