#!/usr/bin/env bash
# golangci-lint config-verify guard (SUPERB v7 T05, 2026-10-13).
# `golangci-lint config verify` validates a .golangci.yml against the
# binary's schema — it is the check that would have caught the dead
# exhaustruct_v5 settings blocks at authoring time (2026-10-12: v2.14.0's
# schema accepts NO exhaustruct_v5 settings; the excludes in all 12 module
# configs never suppressed anything and sat unread for weeks). This repo
# deliberately carries 12 standalone configs (no inheritance exists), so
# one schema drift breaks twelve lint runs at once. The command prints
# nothing on success and exits non-zero with the offending key on drift.
#
#   ./scripts/check-golangci-configs.sh
#
# CI: the golangci-config-verify job installs the pinned golangci-lint
# (public release, no SSH key needed) and runs this script.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
while IFS= read -r gomod; do
	dir="$(dirname "${gomod#./}")"
	if out="$(cd "$dir" && golangci-lint config verify 2>&1)"; then
		echo "OK:   ${dir} .golangci.yml verifies"
	else
		echo "FAIL: ${dir} .golangci.yml rejected:"
		echo "$out" | sed 's/^/      /'
		status=1
	fi
done < <(find . -name go.mod -not -path '*/vendor/*' | sort)

exit $status
