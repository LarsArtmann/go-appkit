#!/usr/bin/env bash
# Tidy-stability guard (2026-11-23, the go-directive floor-reset train).
# Every module's go.mod must be exactly what `go mod tidy` produces — tidy is
# the authority for each module's `go` directive (it raises a module to the
# highest `go` version declared anywhere in its dependency graph; published
# appkit tags and go-cqrs-lite declare patch versions, so most satellites sit
# at 1.27.2–1.27.4 above the root's `go 1.27` published floor) and for
# go.sum contents. A dirty `tidy -diff` means the checked-in files churn on
# the next agent- or BuildFlow-run tidy — the 2026-11-22/23 directive reset
# shipped exactly that class until it was measured and re-set per module.
#
# OFFLINE on a warm module cache; only a cold graph reaches the network.
# Wire-in: pre-tag-checks.sh guard loop (+ CI go-tidy-stable job).
#
#   ./scripts/check-go-tidy-stable.sh              # all 15 modules
#   ./scripts/check-go-tidy-stable.sh cqrs httpx   # subset (pre-commit scope)

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

# shellcheck source=lib/toolchain.sh
source "scripts/lib/toolchain.sh"

status=0
# Optional module args (dir names as printed by find, e.g. "cqrs", "." for
# the root) scope the sweep — pre-commit passes the staged set so the hook
# stays seconds; pre-tag-checks runs the full fleet.
declare -A wanted
if [[ "$#" -gt 0 ]]; then
	for m in "$@"; do wanted["$m"]=1; done
fi
while IFS= read -r gomod; do
	dir="${gomod#./}"
	if [[ "$#" -gt 0 && ! -v "wanted[$dir]" ]]; then
		continue
	fi
	moddir="$(dirname "$gomod")"
	# Resolve the toolchain from THIS module's directive: a patch-level
	# directive is used verbatim, a major.minor floor falls back to the
	# patched belt (scripts/lib/toolchain.sh).
	directive="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' "$gomod" | head -n1)"
	toolchain="$(go_toolchain_for "$directive")"
	if diff_output="$(cd "$moddir" && GOWORK=off GOTOOLCHAIN="$toolchain" GOFLAGS= go mod tidy -diff 2>&1)"; then
		echo "OK:   $dir tidy-stable (GOTOOLCHAIN=$toolchain)"
	else
		echo "FAIL: $dir NOT tidy-stable (go mod tidy -diff, GOTOOLCHAIN=$toolchain):"
		echo "$diff_output" | head -20
		status=1
	fi
done < <(find . -name go.mod -not -path '*/vendor/*' | sort)

if [[ "$status" -ne 0 ]]; then
	echo ""
	echo "Fix: run 'GOWORK=off go mod tidy' in each failing module (or re-set its go directive to the tidy-stable value), then re-run this guard."
fi

exit $status
