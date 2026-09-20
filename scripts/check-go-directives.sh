#!/usr/bin/env bash
# Go-directive parity guard (execution plan 2026-09-20 T05). Every module's
# go.mod must declare the same go directive, and it must match go.work. The
# root go.mod briefly drifted to 1.27.1 (auto-commit d5c6693, 2026-09-18)
# while every satellite and go.work stayed at 1.26.7; the mismatch broke all
# workspace commands, LSP, check-pin-drift.sh, and root builds on the local
# 1.26.7 toolchain. This script is that class of mistake failing fast and
# cheaply, with no git history needed:
#
#   ./scripts/check-go-directives.sh

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

expected="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$expected" ]]; then
	echo "FAIL: go.work has no parseable 'go <version>' directive"
	exit 1
fi

status=0
while IFS= read -r gomod; do
	dir="${gomod#./}"
	version="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' "$gomod" | head -n1)"
	if [[ "$version" == "$expected" ]]; then
		echo "OK:   $dir declares go $version"
	else
		echo "FAIL: $dir declares go ${version:-<none>}, want go $expected (go.work)"
		status=1
	fi
done < <(find . -name go.mod -not -path '*/vendor/*' | sort)

exit $status
