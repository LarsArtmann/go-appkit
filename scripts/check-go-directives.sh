#!/usr/bin/env bash
# Go-directive parity guard (execution plan 2026-09-20 T05, hardened 2026-09-29
# T12, repurposed 2026-11-22, floor-source fixed 2026-11-23).
# TWO contracts, two sources of truth:
#
# 1. PUBLISHED floor = the root go.mod `go` directive (`go 1.27` since the
#    2026-11-22 owner reset). Every module must declare >= it. A module MAY
#    sit ABOVE the floor: `go mod tidy` raises a module's directive to the
#    highest `go` version declared anywhere in its dependency graph (published
#    appkit tags, go-cqrs-lite), so patch-level directives are
#    dependency-forced, not policy — `go mod tidy -diff` is the authority for
#    each module's exact value. MEASURED 2026-11-23: go.work at `go 1.27`
#    with any member above it FAILS every workspace command ("module X listed
#    in go.work file requires go >= 1.27.Y, but go.work lists go 1.27"), so
#    the 2026-11-23 02:57 daemon re-raise of go.work to 1.27.4 was CORRECT.
#
# 2. LOCAL floor = go.work, which Go itself forces to >= every workspace
#    member's directive. This guard checks that direction too: a go.work line
#    BELOW any member is the workspace-build-break class, caught here at
#    commit time instead of at build time. go.work is tooling-managed
#    (BuildFlow go-work-sync normalizes it) — do not hand-raise it to "keep
#    the guard happy"; hand-LOWERING it is the damage this check catches.
#
# Historical damage class: the root go.mod briefly drifted to 1.27.1
# (auto-commit d5c6693, 2026-09-18) while every satellite and go.work stayed
# at 1.26.7; the mismatch broke all workspace commands, LSP,
# check-pin-drift.sh, and root builds on the local 1.26.7 toolchain.
# Additionally, NO module or go.work may carry a `toolchain` directive: the
# repo's toolchain mechanism is the GOTOOLCHAIN env prefix (AGENTS
# "Toolchain" paragraph), and a stray `toolchain` line pins one exact patch
# and reintroduces the same drift class. See .buildflow.yml.
#
#   ./scripts/check-go-directives.sh

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

# Normalize a go directive to major.minor.patch so "go 1.27" and "go 1.27.0"
# compare equal while "go 1.27" vs "go 1.27.1" still fails with a clear
# message. Version components are compared numerically (semver-aware), not
# as strings — a future 1.27.10 must not sort before 1.27.2.
normalize() {
	local v="$1"
	until [[ "$v" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; do
		v="$v.0"
	done
	echo "$v"
}

version_lt() {
	# True if $1 < $2 (numeric component compare, semver-aware).
	local a b
	IFS='.' read -r -a a <<<"$(normalize "$1")"
	IFS='.' read -r -a b <<<"$(normalize "$2")"
	for i in 0 1 2; do
		if ((10#${a[$i]} < 10#${b[$i]})); then
			return 0
		fi
		if ((10#${a[$i]} > 10#${b[$i]})); then
			return 1
		fi
	done
	return 1
}

# Contract 1: the PUBLISHED floor lives in the root go.mod (owner reset
# 2026-11-22). The root module IS the floor-setting module by charter.
floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.mod | head -n1)"
if [[ -z "$floor" ]]; then
	echo "FAIL: root go.mod has no parseable 'go <version>' directive (the published floor)"
	exit 1
fi
floor_n="$(normalize "$floor")"

# Contract 2: go.work must be >= every member directive (Go rejects a lower
# go.work line at build time — measured 2026-11-23).
workgo="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$workgo" ]]; then
	echo "FAIL: go.work has no parseable 'go <version>' directive"
	exit 1
fi
workgo_n="$(normalize "$workgo")"
max_member="$floor"
max_member_dir="root go.mod"

status=0
while IFS= read -r gomod; do
	dir="${gomod#./}"
	version="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' "$gomod" | head -n1)"
	toolchain="$(sed -n 's/^toolchain go\([0-9][0-9.]*\)$/\1/p' "$gomod" | head -n1)"
	if [[ -n "$toolchain" ]]; then
		echo "FAIL: $dir carries a toolchain directive (toolchain go$toolchain) — this repo's mechanism is the GOTOOLCHAIN env prefix (see AGENTS.md); a toolchain line pins one exact patch and reintroduces the 2026-09-18 drift class"
		status=1
	fi
	if [[ -z "$version" ]]; then
		echo "FAIL: $dir has no parseable 'go <version>' directive"
		status=1
		continue
	fi
	if [[ "$(normalize "$version")" == "$floor_n" ]]; then
		echo "OK:   $dir declares go $version"
	elif version_lt "$version" "$floor"; then
		echo "FAIL: $dir declares go $version, BELOW the published floor go $floor (root go.mod) — raise it to the floor"
		status=1
	else
		echo "OK:   $dir declares go $version (above the published floor go $floor; dependency-forced, go mod tidy -diff is the authority)"
	fi
	if version_lt "$max_member" "$version"; then
		max_member="$version"
		max_member_dir="$dir"
	fi
done < <(find . -name go.mod -not -path '*/vendor/*' | sort)

if version_lt "$workgo" "$max_member"; then
	echo "FAIL: go.work lists go $workgo but $max_member_dir requires go >= $max_member — every workspace command dies ('module ... requires go >= $max_member, but go.work lists go $workgo'); raise go.work to go $max_member"
	status=1
else
	echo "OK:   go.work lists go $workgo (>= max member directive go $max_member from $max_member_dir)"
fi

worktoolchain="$(sed -n 's/^toolchain go\([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -n "$worktoolchain" ]]; then
	echo "FAIL: go.work carries a 'toolchain $worktoolchain' directive — same drift class as go.mod toolchain lines"
	status=1
fi

exit $status
