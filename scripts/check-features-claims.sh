#!/usr/bin/env bash
# FEATURES-claims guard (2026-11-06, issue54 session §f7 + §f23): the
# module-section summaries in FEATURES.md ("Shipped through <module>
# **vX.Y.Z**") are version-truth claims that must equal that module's newest
# git tag. The 2026-11-06 class: health shipped v0.3.0 while a FEATURES
# section still claimed an older version — stale claims sat red-free until a
# release tripped over them. This script is that check, mechanical and
# FAST+OFFLINE (all three guard venues: pre-commit, pre-tag-checks, CI).
#
# Per-row "shipped in vX.Y.Z" evidence pointers (table cells naming the
# release that introduced a feature) are HISTORICAL statements and are
# deliberately NOT checked — only the section-level Shipped-through claims
# are version-truth.
#
# Exit codes: 0 green · 1 drift found · 2 usage.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
claims=0

tag_pattern() {
	case "$1" in
	core) echo "v*" ;;
	cqrs | otel | flightrecorder | health | realtime | docs | errorpages | security | systemd | config | preset | httpx) echo "$1/v*" ;;
	frh) echo "flightrecorderhealth/v*" ;;
	*) return 1 ;;
	esac
}

while IFS= read -r line; do
	# Capture: Shipped through <name> **vX.Y.Z**
	name="$(sed -n 's/^Shipped through \([a-z]*\) \*\*v[0-9][0-9.]*\*\*.*/\1/p' <<<"$line")"
	version="$(sed -n 's/^Shipped through [a-z]* \*\*v\([0-9][0-9.]*\)\*\*.*/\1/p' <<<"$line")"
	[[ -n "$name" && -n "$version" ]] || continue
	claims=$((claims + 1))

	if ! pattern="$(tag_pattern "$name")"; then
		echo "FAIL: FEATURES claims unknown module '$name' (extend tag_pattern in check-features-claims.sh)"
		status=1
		continue
	fi
	newest="$(git tag -l -- "$pattern" | sort -V | tail -n1)"
	newest_version="${newest##*/}" # core: v0.10.0 · satellite: v0.1.0
	if [[ -z "$newest" ]]; then
		echo "FAIL: FEATURES claims $name v$version but no tags match $pattern"
		status=1
	elif [[ "v$version" == "$newest_version" ]]; then
		echo "OK:   FEATURES $name claim v$version == newest tag $newest"
	else
		echo "FAIL: FEATURES claims $name v$version but the newest tag is $newest (module shipped without the FEATURES version-truth update)"
		status=1
	fi
done < <(grep "Shipped through" FEATURES.md)

if ((claims == 0)); then
	echo "FAIL: no 'Shipped through <module> **vX.Y.Z**' claims found in FEATURES.md (claim format drift?)"
	status=1
elif ((status == 0)); then
	echo "FEATURES CLAIMS: $claims claim(s) green against the tag set"
fi
exit "$status"
