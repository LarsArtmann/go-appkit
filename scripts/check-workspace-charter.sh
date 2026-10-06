#!/usr/bin/env bash
# Workspace-charter guard (2026-10-06): go.work membership must be exactly
# "the root module + every depth-1 directory carrying a go.mod, EXCEPT
# ./integration". The pin charter (integration/doc.go) pins PUBLISHED tags
# only; workspace membership would resolve those pins to local code and
# merges integration's requires into the repo-level aggregate node in
# project-dependency-graph, creating a false cqrs-htmx cycle.
#
# Why a guard: the auto-commit daemon re-added ./integration twice after
# manual reverts (leaked mid-session, restored in d2b5da0, re-leaked in
# e7fbc07 within minutes). Manual fixes lose that race; this script makes
# the next leak fail CI and the pre-tag gate loudly instead of silently
# shipping. When a NEW module joins the repo, go.work must gain it in the
# same change (this guard enforces the unified-toolchain membership policy).

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ ! -f go.work ]]; then
	echo "FAIL: go.work not found at repo root"
	exit 1
fi

# Declared: entries of the `use (...)` block.
declared="$(awk '/^use \(/{f=1;next} /^\)[[:space:]]*$/{f=0} f{gsub(/^[[:space:]]+|[[:space:]]+$/,""); if($0!="") print}' go.work | sort -u)"

# Expected: root + every depth-1 go.mod dir, minus the ./integration exception.
expected="$(
	{
		echo "."
		ls -d */go.mod 2>/dev/null | sed 's|/go.mod$||; s|^|./|'
	} | grep -vx './integration' | sort -u
)"

status=0

if [[ "$(grep -c '^use ' go.work)" != "1" ]]; then
	echo "FAIL: go.work must carry exactly one 'use' block opener (found $(grep -c '^use ' go.work))"
	status=1
fi

if [[ "$declared" == "$expected" ]]; then
	echo "OK: go.work membership matches the charter (root + $(grep -vc '^\.$' <<<"$expected") modules, ./integration excluded)"
else
	echo "FAIL: go.work membership deviates from the charter:"
	diff --label declared --label expected <(printf '%s\n' "$declared") <(printf '%s\n' "$expected") || true
	echo "fix: add/remove members in go.work (or, if a module legitimately leaves the"
	echo "charter, update the exception list in scripts/check-workspace-charter.sh)"
	status=1
fi

exit "$status"
