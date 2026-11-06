#!/usr/bin/env bash
# check-touched-modules.sh — run the per-module hermetic verify ONLY for
# modules touched by the staged set (2026-11-06, 16-13 close-out improvement
# #1 made mechanical). Before this, "verify what you touched" was a manual
# discipline; a staged change to two satellites either triggered a full
# verify-all (minutes, skipped under pressure) or nothing (broken builds
# landed and CI caught them a push later).
#
# Mapping: a staged path under <dir>/ touches module <dir>; a repo-root path
# (not under any go.work member dir) touches core ("."). In staged mode a
# module with NO Go payload among its touched files (.go/.mod/.sum) is
# skipped — doc/config-only commits do not pay a build. Named paths and
# --all verify the named modules regardless of payload. Integration is
# verified with the same hermetic prefix as everywhere (its TestMain rejects
# workspace-mode `go test`).
#
# Default depth: build + vet (seconds per module — safe as a FATAL pre-commit
# guard). `--tests` adds the full `go test ./... -race -count=1` loop (the
# complete per-module verify loop; manual / pre-tag use — minutes per module).
#
# Usage:
#   scripts/check-touched-modules.sh              # staged set, build+vet
#   scripts/check-touched-modules.sh --tests      # staged set, full loop
#   scripts/check-touched-modules.sh --all        # every module, build+vet
#   scripts/check-touched-modules.sh cqrs otel    # named modules
#
# Exit codes: 0 green · 1 verification failure · 2 usage.

set -uo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

run_tests=false
verify_all=false
paths=()
for arg in "$@"; do
	case "$arg" in
	--tests) run_tests=true ;;
	--all) verify_all=true ;;
	--staged) ;; # default
	-h* | --help)
		sed -n '2,32p' "$0"
		exit 0
		;;
	*) paths+=("$arg") ;;
	esac
done

floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$floor" ]]; then
	echo "FAIL: could not read the go floor from go.work"
	exit 1
fi

# Module map from go.work (charter guard keeps this honest).
declare -A members=()
members["."]=1
while IFS= read -r use_entry; do
	case "$use_entry" in
	".") ;;
	.*) members["${use_entry#./}"]=1 ;;
	esac
done < <(awk '/^use \(/{f=1; next} /^\)/{f=0} f && NF {print $1}' go.work)

# Collect touched modules: explicit selection always verifies; staged mode
# requires a Go payload per module.
declare -A pending=()
declare -A payload=()
record() {
	local f="$1" dir
	dir="${f%%/*}"
	[[ "$dir" == "${f}" ]] && dir="."
	if [[ -n "${members[$dir]:-}" ]]; then
		pending["$dir"]=1
		case "$f" in
		*.go | *.mod | *.sum) payload["$dir"]=1 ;;
		esac
	fi
}

if [[ "$verify_all" == true ]]; then
	for m in "${!members[@]}"; do
		pending["$m"]=1
		payload["$m"]=1
	done
elif ((${#paths[@]} > 0)); then
	for p in "${paths[@]}"; do
		if [[ -n "${members[$p]:-}" ]]; then
			pending["$p"]=1
			payload["$p"]=1
		else
			echo "SKIP: '$p' is not a module directory (go.work members: . ${!members[*]})"
		fi
	done
else
	while IFS= read -r f; do
		[[ -n "$f" ]] || continue
		record "$f"
	done < <(git diff --cached --name-only 2>/dev/null)
fi

status=0
verified=0
skipped=0
for m in "${!pending[@]}"; do
	if [[ "$verify_all" == false && ${#paths[@]} -eq 0 && -z "${payload[$m]:-}" ]]; then
		label="$m"
		[[ "$label" == "." ]] && label="core"
		echo "SKIP: $label touched without Go payload (doc/config-only)"
		skipped=$((skipped + 1))
		continue
	fi
	label="$m"
	[[ "$label" == "." ]] && label="core"
	echo "== $label (build+vet$( [[ "$run_tests" == true ]] && echo '+test -race' )) =="
	if (cd "$m" &&
		GOWORK=off GOTOOLCHAIN="go$floor" go build ./... &&
		GOWORK=off GOTOOLCHAIN="go$floor" go vet ./... &&
		{ [[ "$run_tests" == false ]] || GOWORK=off GOTOOLCHAIN="go$floor" go test ./... -race -count=1; }); then
		echo "OK:   $label"
		verified=$((verified + 1))
	else
		echo "FAIL: $label (hermetic verify failed — fix before committing)"
		status=1
	fi
done

if ((status == 0)); then
	echo "TOUCHED MODULES: $verified module(s) verified, $skipped doc-only skip(s)"
fi
exit "$status"
