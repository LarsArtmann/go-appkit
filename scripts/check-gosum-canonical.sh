#!/usr/bin/env bash
# go.sum canonical-checksum guard (2026-10-14 incident): CI's first
# unauthenticated matrix run caught integration/go.sum carrying a
# PRE-re-tag zip hash for health v0.2.0 (SECURITY ERROR: checksum mismatch
# on the fresh runner) — the shared local GOMODCACHE hid it, because the
# cached zip verifies against the cached-poisoned go.sum. A force-moved tag
# never updates go's VCS cache, so local tidy keeps recording the pre-move
# hash forever (full diagnosis: AGENTS.md Gotchas, "Stale go VCS cache
# after a re-tag").
#
# This guard replicates the fresh-runner check locally: every module's
# dependency tree is resolved from an ISOLATED, throwaway GOMODCACHE
# against the public proxy + checksum database. A go.sum entry that
# disagrees with sum.golang.org dies HERE — before the tag push — instead
# of on the next CI runner. Each module's go.mod/go.sum are verified from
# a COPY, so the guard never writes to the working tree (a missing-sum
# entry would otherwise be auto-added and silently dirty the repo).
#
# Needs network (proxy.golang.org + sum.golang.org). Slow by design: full
# per-module downloads into an empty cache (shared across the modules of
# one run, so the toolchain downloads once). CI's fresh-runner matrix
# covers every push; this script exists for the pre-tag gate
# (scripts/pre-tag-checks.sh) and on-demand runs.
#
# Usage: scripts/check-gosum-canonical.sh [module ...]
#   Args are module dirs (repo-relative names like "cqrs" or paths);
#   default: all twelve modules.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

default_modules=(
	"." cqrs realtime security systemd otel health
	flightrecorder flightrecorderhealth errorpages docs integration
)
if (($# > 0)); then
	modules=("$@")
else
	modules=("${default_modules[@]}")
fi

floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$floor" ]]; then
	echo "FAIL: could not read the go floor from go.work" >&2
	exit 1
fi

cache="$(mktemp -d /tmp/go-appkit-gosum-XXXXXX)"
work="$(mktemp -d /tmp/go-appkit-gosum-mod-XXXXXX)"
# Green runs remove their scratch dirs; failed runs keep them (and say so)
# so the disagreeing hash can be inspected in the throwaway cache. An
# unexpected set -e abort keeps them too (run_status is still 1).
run_status=1
cleanup() {
	if ((run_status == 0)); then
		# go extracts module-cache entries 0444/0555; rm needs the dirs
		# writable again first (AGENTS Gotchas, "Stale go VCS cache").
		chmod -R u+w "$cache" "$work" 2>/dev/null || true
		rm -rf "$cache" "$work"
	else
		echo "gosum-canonical: failed run kept its scratch dirs for inspection: $cache $work" >&2
	fi
}
trap cleanup EXIT

status=0
deferred=0

# Release-window deferral (2026-11-06, issue54 session §e2): a family pin on
# a tag that exists LOCALLY but not on origin cannot resolve from the proxy —
# probing it turns the release window into a 45-minute red. Deferral is the
# loud, honest answer ("deferred until publish"), not a silent skip. Only
# go-appkit's OWN tags are checked: sibling-repo requires (cqrs-htmx, go-sse,
# httputil) are published by their own repos and resolve normally.
origin_tags="$(git ls-remote --tags origin 2>/dev/null | sed -n 's|.*refs/tags/\([^^]*\)$|\1|p' || true)"
if [[ -z "$origin_tags" ]]; then
	echo "NOTE: origin unreachable — release-window deferral disabled (downloads will fail loudly on local-only pins)" >&2
fi

local_only_family_pin() {
	# Prints the local-only tag name if any go-appkit family require in
	# $1/go.mod resolves to a local-only tag; empty otherwise.
	local gomod="$1" mod_path tag
	while IFS=' ' read -r mod_path version; do
		[[ -n "$mod_path" ]] || continue
		case "$mod_path" in
		github.com/larsartmann/go-appkit) tag="$version" ;;
		github.com/larsartmann/go-appkit/*) tag="${mod_path#github.com/larsartmann/go-appkit/}/$version" ;;
		*) continue ;;
		esac
		if git cat-file -e "${tag}^{commit}" 2>/dev/null && ! grep -qxF "$tag" <<<"$origin_tags"; then
			echo "$tag"
			return 0
		fi
	done < <(awk '/^require \(/,/^\)/' "$gomod" | awk '$1 ~ /^github\.com\/larsartmann\/go-appkit/ {print $1 " " $2}')
	return 1
}

for mod in "${modules[@]}"; do
	label="${mod#./}"
	if [[ -z "$label" ]]; then label="core"; fi
	if [[ ! -f "$mod/go.mod" || ! -f "$mod/go.sum" ]]; then
		echo "FAIL: $label has no go.mod/go.sum pair"
		status=1
		continue
	fi

	if [[ -n "$origin_tags" ]] && stuck_tag="$(local_only_family_pin "$mod/go.mod")"; then
		echo "== $label =="
		echo "DEFERRED: $label — family pin resolves to LOCAL-ONLY tag $stuck_tag (not on origin yet)"
		echo "          deferred until publish: run this guard AFTER `git push origin master $stuck_tag`"
		deferred=$((deferred + 1))
		continue
	fi

	echo "== $label =="
	modcopy="$work/${label//\//_}"
	mkdir -p "$modcopy"
	cp "$mod/go.mod" "$mod/go.sum" "$modcopy/"
	if (cd "$modcopy" &&
		GOWORK=off \
			GOTOOLCHAIN="go$floor" \
			GOPRIVATE='' GOFLAGS='' \
			GOPROXY=https://proxy.golang.org \
			GOMODCACHE="$cache" \
			go mod download all); then
		echo "OK: $label go.sum agrees with proxy + sumdb"
	else
		echo "FAIL: $label go.sum disagrees with the checksum database (pre-re-tag hash?)"
		echo "      repair recipe: AGENTS.md Gotchas, 'Stale go VCS cache after a re-tag'"
		status=1
	fi
done

if ((deferred > 0)); then
	echo "GOSUM CANONICAL: $deferred module(s) DEFERRED (local-only family pins — release window); verified module(s) green against proxy + sumdb"
elif ((status == 0)); then
	echo "GOSUM CANONICAL: ${#modules[@]} module(s) green against proxy + sumdb"
fi
run_status=$status
exit "$status"
