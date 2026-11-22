#!/usr/bin/env bash
# Go-directive parity guard (execution plan 2026-09-20 T05, hardened 2026-09-29
# T12, repurposed 2026-11-22).
# go.work holds the repo FLOOR (major.minor `go 1.27` since the 2026-11-22
# owner reset) and no module may declare a go directive BELOW it. A module MAY
# declare ABOVE the floor: `go mod tidy` raises a module's directive to the
# highest `go` version declared anywhere in its dependency graph (published
# appkit tags, go-cqrs-lite), so patch-level directives are dependency-forced,
# not policy — tidy is the authority for each module's exact value. This guard
# catches the damage class that matters: a module sitting below the floor. The
# root go.mod briefly drifted to 1.27.1 (auto-commit d5c6693, 2026-09-18)
# while every satellite and go.work stayed at 1.26.7; the mismatch broke all
# workspace commands, LSP, check-pin-drift.sh, and root builds on the local
# 1.26.7 toolchain. Additionally, NO module may carry a `toolchain` directive:
# the repo's toolchain mechanism is the GOTOOLCHAIN env prefix (AGENTS
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

expected="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$expected" ]]; then
	echo "FAIL: go.work has no parseable 'go <version>' directive"
	exit 1
fi
expected_n="$(normalize "$expected")"

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
	elif [[ "$(normalize "$version")" == "$expected_n" ]]; then
		echo "OK:   $dir declares go $version"
	elif version_lt "$version" "$expected"; then
		echo "FAIL: $dir declares go $version, BELOW the floor go $expected (go.work) — raise it to the floor"
		status=1
	else
		echo "OK:   $dir declares go $version (above the go.work floor go $expected; dependency-forced, tidy is the authority)"
	fi
done < <(find . -name go.mod -not -path '*/vendor/*' | sort)

worktoolchain="$(sed -n 's/^toolchain go\([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -n "$worktoolchain" ]]; then
	echo "FAIL: go.work carries a 'toolchain $worktoolchain' directive — same drift class as go.mod toolchain lines"
	status=1
fi

exit $status
