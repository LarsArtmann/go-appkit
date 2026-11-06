#!/usr/bin/env bash
# check-touched-modules.sh — run the per-module hermetic verify ONLY for
# modules touched by the staged set (2026-11-06, 16-13 close-out improvement
# #1 made mechanical). Before this, "verify what you touched" was a manual
# discipline; a staged change to two satellites either triggered a full
# verify-all (minutes, skipped under pressure) or nothing (broken builds
# landed and CI caught them a push later).
#
# Mapping: a staged path under <dir>/ touches module <dir>; a repo-root path
# (not under any go.work member dir) touches core ("."). A module whose
# touched set carries NO Go files (.go/.mod/.sum) is skipped — doc/config-only
# commits do not pay a build. Integration is verified with the same hermetic
# prefix as everywhere (its TestMain rejects workspace-mode `go test`).
#
# Default depth: build + vet (seconds per module — safe as a FATAL pre-commit
# guard). `--tests` adds the full `go test ./... -race -count=1` loop (the
# complete per-module verify loop; manual / pre-tag use — minutes per module).
# `--staged` (default) reads `git diff --cached`; pass paths or `--all` for
# explicit module verification.
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
		sed -n '2,30p' "$0"
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

if [[ "$verify_all" == true ]]; then
	touched=("${!members[@]}")
elif ((${#paths[@]} > 0)); then
	touched_source="named modules"
	for p in "${paths[@]}"; do
		if [[ -n "${members[$p]:-}" ]]; then touched+=("$p"); fi
	done
else
	touched_source="staged set"
	while IFS= read -r f; do
		[[ -n "$f" ]] || continue
		dir="${f%%/*}"
		if [[ "$dir" == "${f}" ]]; then
			touched+=(".") # repo-root path
		elif [[ -n "${members[$dir]:-}" ]]; then
			touched+=("$dir")
		fi
	done < <(git diff --cached --name-only 2>/dev/null)
fi

# Dedupe + drop modules with no Go payload in the touched set.
declare -A go_touched=()
if [[ "$verify_all" == true || ${#paths[@]} -gt 0 ]]; then
	for m in "${touched[@]:-}"; do
		[[ -n "$m" ]] && go_touched["$m"]=1
	done
else
	for m in "${touched[@]:-}"; do
		[[ -n "$m" ]] || continue
		go_touched["$m"]="${go_touched[$m]:-}"
	done
	while IFS= read -r f; do
		[[ -n "$f" ]] || continue
		dir="${f%%/*}"
		[[ "$dir" == "${f}" ]] && dir="."
		case "$f" in
		*.go | *.mod | *.sum) [[ -n "${go_touched[$dir]:-}" ]] && go_touched["$dir"]=1 ;;
		esac
	done < <(git diff --cached --name-only 2>/dev/null)
fi

# Second pass sets 0-sentinels above; flip to 1 only for modules with Go files.
declare -A final=()
if [[ "$verify_all" == true || ${#paths[@]} > 0 ]]; then
	for m in "${!go_touched[@]}"; do final["$m"]=1; done
else
	while IFS= read -r f; do
		[[ -n "$f" ]] || continue
		dir="${f%%/*}"
		[[ "$dir" == "${f}" ]] && dir="."
		case "$f" in
		*.go | *.mod | *.sum) final["$dir"]=1 ;;
		esac
	done < <(git diff --cached --name-only 2>/dev/null)
fi

status=0
verified=0
skipped=0
for m in "${!final[@]}"; do
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
	echo "TOUCHED MODULES: $verified module(s) verified ($touched_source; doc-only modules skipped)"
fi
exit "$status"
