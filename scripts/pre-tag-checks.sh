#!/usr/bin/env bash
# Pre-tag gate (post-mortem of flightrecorderhealth v0.1.4, 2026-09-29).
# v0.1.4 shipped `go 1.27` because BuildFlow's normalize step downgraded the
# directive AFTER the last green guard probe and BEFORE `git tag`; nothing
# re-ran the guards between the automation run and the tag. This script makes
# the guard-before-push sequence mechanical. Run it AFTER `git tag` and
# BEFORE `git push`:
#
#   git tag -a <tag> -m "..."
#   ./scripts/pre-tag-checks.sh <tag>   # <- this must be green
#   git push origin master <tag>
#
# Checks:
#   1. The three standing guards (directives, pin-drift, dependabot parity).
#   2. The TAGGED ARTIFACT's go.mod (git show <tag>:…, not the working tree):
#      go directive == the go.work floor, no `toolchain` directive, and no
#      filesystem `replace` (tag hygiene: working-tree replaces used for
#      cross-repo debugging must never ship).

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ $# -ne 1 ]]; then
	echo "usage: $0 <tag>   (run AFTER git tag, BEFORE git push)"
	exit 2
fi
tag="$1"

status=0
fail() {
	echo "FAIL: $*"
	status=1
}

# --- 1) Standing guards -----------------------------------------------------
for guard in check-go-directives.sh check-pin-drift.sh check-dependabot-parity.sh; do
	echo "== $guard =="
	if ! "./scripts/$guard"; then
		fail "$guard failed — fix before pushing $tag"
	fi
done

# --- 2) Tagged artifact -----------------------------------------------------
if ! git cat-file -e "${tag}^{commit}" 2>/dev/null; then
	fail "tag $tag does not exist locally (git tag first)"
	exit 1
fi

case "$tag" in
v[0-9]*) module_dir="" ;;
*/)
	fail "tag $tag ends with a slash — malformed module tag"
	exit 1
	;;
*) module_dir="${tag%/v*}" ;;
esac
gomod_path="${module_dir:+$module_dir/}go.mod"

tagged_go="$(git show "$tag:$gomod_path" 2>/dev/null | sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' | head -n1)"
floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$tagged_go" ]]; then
	fail "$tag:$gomod_path has no parseable 'go <version>' directive"
elif [[ "$tagged_go" != "$floor" ]]; then
	fail "$tag:$gomod_path declares go $tagged_go but the floor is go $floor (exact match required — the v0.1.4 class)"
else
	echo "OK:   $tag:$gomod_path declares go $tagged_go (floor)"
fi

tagged_toolchain="$(git show "$tag:$gomod_path" | sed -n 's/^toolchain go\([0-9][0-9.]*\)$/\1/p' | head -n1)"
if [[ -n "$tagged_toolchain" ]]; then
	fail "$tag:$gomod_path carries a toolchain directive — this repo's mechanism is the GOTOOLCHAIN prefix"
fi

tagged_replace_fs="$(git show "$tag:$gomod_path" | sed -n 's/^replace .*=> \(....*\)/\1/p' | grep -Ev '^github|^golang.org|^google|^gopkg.in' || true)"
if [[ -n "$tagged_replace_fs" ]]; then
	fail "$tag:$gomod_path carries a filesystem replace: $tagged_replace_fs — remove before tagging (tag hygiene)"
fi

if ((status == 0)); then
	echo "PRE-TAG GATE GREEN: $tag is safe to push"
fi
exit "$status"
