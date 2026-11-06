#!/usr/bin/env bash
# Tagged-CHANGELOG match guard (2026-11-06, issue54 session §f31): every
# release tag must carry, IN ITS OWN TREE (git show <tag>:…), a CHANGELOG
# section dated for that release — `## [<version>] - <YYYY-MM-DD>` matching
# the tag name and the tag's creation date. This is the release-ritual rule
# "date the module's CHANGELOG [Unreleased] → [version] - <date>" as code;
# before it existed, a tag could ship while its CHANGELOG still said
# [Unreleased] and nothing caught it.
#
# Usage:
#   scripts/check-tagged-changelogs.sh <tag> [<tag>...]   # release mode
#   scripts/check-tagged-changelogs.sh --all              # full sweep
#
# Release mode is what scripts/pre-tag-checks.sh calls (FAST+OFFLINE, runs
# in the pre-tag venue where the tag args exist; --all is manual). History
# note: this guard is intentionally NOT retro-enforced in CI — old tags
# predate the convention — so the green baseline is the release flow itself.
#
# Exit codes: 0 green · 1 drift found · 2 usage.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
fail() {
	echo "FAIL: $*"
	status=1
}
ok() { echo "OK:   $*"; }

if (($# < 1)); then
	echo "usage: $0 <tag> [<tag>...] | --all"
	exit 2
fi

if [[ "$1" == "--all" ]]; then
	tags=()
	while IFS= read -r t; do tags+=("$t"); done < <(git tag -l)
else
	tags=("$@")
fi

for tag in "${tags[@]}"; do
	case "$tag" in
	v[0-9]*) module_dir="" ;;
	*/)
		fail "tag $tag ends with a slash — malformed module tag"
		continue
		;;
	*) module_dir="${tag%/v*}" ;;
	esac
	version="${tag##*/}"
	version="${version#v}"
	version_regex="$(printf '%s' "$version" | sed 's/\./\\./g')"
	changelog="${module_dir:+$module_dir/}CHANGELOG.md"

	if ! git cat-file -e "${tag}^{commit}" 2>/dev/null; then
		fail "$tag does not exist locally (git tag first)"
		continue
	fi
	if ! git cat-file -e "$tag:$changelog" 2>/dev/null; then
		fail "$tag tree carries no $changelog (every released module must ship its CHANGELOG)"
		continue
	fi

	tag_date="$(git for-each-ref "refs/tags/$tag" --format='%(creatordate:short)')"
	section="## \\[$version_regex\\] - $tag_date"
	if git show "$tag:$changelog" | grep -Eq "^${section}[[:space:]]*$"; then
		ok "$tag: $changelog has '## [$version] - $tag_date'"
	else
		found="$(git show "$tag:$changelog" | grep -E "^## \[$version_regex\]" | head -n1 || true)"
		if [[ -z "$found" ]]; then
			fail "$tag: $changelog has NO '## [$version] …' section (CHANGELOG never dated for the release)"
		else
			fail "$tag: $changelog says '$found' but the tag's date is $tag_date (section date drift)"
		fi
	fi
done

if ((status == 0)); then
	echo "TAGGED CHANGELOGS: ${#tags[@]} tag(s) green"
fi
exit "$status"
