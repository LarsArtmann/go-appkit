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
# Multi-tag waves (2026-10-12): pass every tag of the wave — the four
# standing guards run ONCE, then each TAGGED ARTIFACT is asserted. An
# 8-tag wave is one gate run, not eight:
#
#   ./scripts/pre-tag-checks.sh v0.9.0 health/v0.2.0 ...
#   git push origin master v0.9.0 health/v0.2.0 ...
#
# Checks:
#   1. The TEN standing guards (directives, pin-drift, dependabot parity,
#      workspace charter, routed-destinations, status-index, golangci-configs,
#      consumer-claims, SSE-protocol, gosum-canonical,
#      art-dupl clone-regression vs the committed baseline).
#      Venue matrix (SUPERB v10 M14): FAST+OFFLINE guards run in all three
#      venues (pre-commit where wired, this gate, CI); SLOW/NETWORK guards
#      (consumer-claims needs the sibling clone, gosum-canonical an isolated
#      proxy cache) run HERE ONLY; TOOL-BINARY guards run in CI only
#      (golangci-config-verify + structure-linter [private binary, rides
#      ci-preflight]); art-dupl runs here + CI (fast once installed).
#      2026-10-18 additions: sse-protocol is an offline grep (test
#      broadcasts must sync on SubscriberCount); gosum-canonical resolves
#      every module's deps through an ISOLATED cache against proxy+sumdb (the
#      2026-10-14 pre-re-tag-hash incident class — slow by design, needs
#      network like pin-drift); art-dupl (SUPERB v9) fails on clone groups
#      not in scripts/art-dupl-baseline.json. 2026-11-02 additions
#      (SUPERB v10): routed-destinations — every "→ routed …" marker in the
#      living docs must resolve (the dead-pointer class the docs-health pass
#      hand-fixed eight instances of); golangci-configs joined so schema
#      drift cannot ship in a tag either (it was CI-only since SUPERB v7).
#   2. Each TAGGED ARTIFACT's go.mod (git show <tag>:…, not the working tree):
#      go directive == the go.work floor, no `toolchain` directive, and no
#      filesystem `replace` (tag hygiene: working-tree replaces used for
#      cross-repo debugging must never ship).
#   3. Each TAGGED ARTIFACT's CHANGELOG (its own tree) carries a
#      `## [<version>] - <tag date>` section — the release-ritual dating
#      rule as code (check-tagged-changelogs.sh, issue54 §f31).
#      2026-11-06 additions: module-licenses + features-claims join the
#      FAST+OFFLINE standing-guard set; the release-window rule in
#      check-pin-drift makes the tag→push window green (loud NOTEs, not
#      reds).

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ $# -lt 1 ]]; then
	echo "usage: $0 <tag> [<tag>...]   (run AFTER git tag, BEFORE git push)"
	exit 2
fi
tags=("$@")

status=0
fail() {
	echo "FAIL: $*"
	status=1
}

# --- 1) Standing guards -----------------------------------------------------
for guard in check-go-directives.sh check-agents-deps.sh check-pin-drift.sh check-dependabot-parity.sh check-workspace-charter.sh check-routed-destinations.sh check-status-index.sh check-golangci-configs.sh check-consumer-claims.sh check-sse-protocol.sh check-gosum-canonical.sh check-art-dupl.sh check-module-licenses.sh check-features-claims.sh; do
	echo "== $guard =="
	if ! "./scripts/$guard"; then
		fail "$guard failed — fix before pushing ${tags[*]}"
	fi
done

# --- 2) Tagged artifacts -----------------------------------------------------
for tag in "${tags[@]}"; do
	if ! git cat-file -e "${tag}^{commit}" 2>/dev/null; then
		fail "tag $tag does not exist locally (git tag first)"
		continue
	fi

	case "$tag" in
	v[0-9]*) module_dir="" ;;
	*/)
		fail "tag $tag ends with a slash — malformed module tag"
		continue
		;;
	*) module_dir="${tag%/v*}" ;;
	esac
	gomod_path="${module_dir:+$module_dir/}go.mod"

	tagged_go="$(git show "$tag:$gomod_path" 2>/dev/null | sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' | head -n1 || true)"
	floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
	if [[ -z "$tagged_go" ]]; then
		fail "$tag:$gomod_path has no parseable 'go <version>' directive"
	elif [[ "$tagged_go" != "$floor" ]]; then
		fail "$tag:$gomod_path declares go $tagged_go but the floor is go $floor (exact match required — the v0.1.4 class)"
	else
		echo "OK:   $tag:$gomod_path declares go $tagged_go (floor)"
	fi

	tagged_toolchain="$(git show "$tag:$gomod_path" 2>/dev/null | sed -n 's/^toolchain go\([0-9][0-9.]*\)$/\1/p' | head -n1 || true)"
	if [[ -n "$tagged_toolchain" ]]; then
		fail "$tag:$gomod_path carries a toolchain directive — this repo's mechanism is the GOTOOLCHAIN prefix"
	fi

	tagged_replace_fs="$(git show "$tag:$gomod_path" | sed -n 's/^replace .*=> \(....*\)/\1/p' | grep -Ev '^github|^golang.org|^google|^gopkg.in' || true)"
	if [[ -n "$tagged_replace_fs" ]]; then
		fail "$tag:$gomod_path carries a filesystem replace: $tagged_replace_fs — remove before tagging (tag hygiene)"
	fi
done

# --- 3) Tagged CHANGELOGs (issue54 §f31): each tagged artifact's own tree
#        must carry a CHANGELOG section dated for the release.
if ! ./scripts/check-tagged-changelogs.sh "${tags[@]}"; then
	fail "tagged-changelogs failed — date the module CHANGELOG before pushing ${tags[*]}"
fi

if ((status == 0)); then
	echo "PRE-TAG GATE GREEN: ${tags[*]} safe to push"
fi
exit "$status"
