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
if (( $# > 0 )); then
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
trap 'echo "gosum-canonical: scratch dirs left behind: $cache $work" >&2' EXIT

status=0
for mod in "${modules[@]}"; do
	label="${mod#./}"
	if [[ -z "$label" ]]; then label="core"; fi
	if [[ ! -f "$mod/go.mod" || ! -f "$mod/go.sum" ]]; then
		echo "FAIL: $label has no go.mod/go.sum pair"
		status=1
		continue
	fi

	echo "== $label =="
	modcopy="$work/${label//\//_}"
	mkdir -p "$modcopy"
	cp "$mod/go.mod" "$mod/go.sum" "$modcopy/"
	if (cd "$modcopy" && \
		GOWORK=off \
		GOTOOLCHAIN="go$floor" \
		GOPRIVATE= GOFLAGS= \
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

if ((status == 0)); then
	echo "GOSUM CANONICAL: ${#modules[@]} module(s) green against proxy + sumdb"
fi
exit "$status"
