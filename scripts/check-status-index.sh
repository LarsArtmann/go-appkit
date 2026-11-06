#!/usr/bin/env bash
# check-status-index.sh — docs/status/README.md liveness (both directions).
#
# The status index is the canonical tree of the report corpus; when it
# drifts, the next session re-treads settled work. Two legs:
#
#   1. INDEX->DISK: every filename listed in the Current table or the
#      archived manifest must exist on disk (Current rows under
#      docs/status/, `archived/` rows under docs/status/archived/,
#      `../reviews/` rows under docs/reviews/, `../planning/` rows under
#      docs/planning/). Catches the renamed-while-indexed class.
#   2. DISK->INDEX: every report under docs/status/*.md (except this
#      README) must appear in the index at least once. Catches the
#      wrote-but-never-indexed class.
#
# Exit codes: 0 green · 1 drift found · 2 usage.

set -uo pipefail
cd "$(dirname "$0")/.." || exit 2

INDEX=docs/status/README.md
fail=0

# Leg 1: index -> disk
while IFS= read -r name; do
	case "$name" in
	archived/*) path="docs/status/$name" ;;
	../reviews/*) path="docs/${name#../}" ;;
	../planning/*) path="docs/${name#../}" ;;
	# plain names are ambiguous: the Current table lists docs/status/ and
	# the bulk-archive manifest lists names relative to archived/
	*) if [ -f "docs/status/$name" ]; then
		continue
	elif [ -f "docs/status/archived/$name" ]; then
		continue
	else
		path="docs/status/$name"
		echo "INDEX-STALE: $name (no file at $path or docs/status/archived/$name)"
		fail=1
		continue
	fi ;;
	esac
	if [ ! -f "$path" ]; then
		echo "INDEX-STALE: $name (no file at $path)"
		fail=1
	fi
done < <(grep -oE '`(archived/|\.\./reviews/|\.\./planning/)?[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}-[0-9]{2}_[A-Za-z0-9._-]+\.(md|html)`' "$INDEX" | tr -d '`' | sort -u)

# Leg 2: disk -> index
for f in docs/status/*.md; do
	base="$(basename "$f")"
	[ "$base" = "README.md" ] && continue
	stem="${base%.md}"
	if ! grep -qF "$stem" "$INDEX"; then
		echo "DISK-UNINDEXED: $base (no mention in $INDEX)"
		fail=1
	fi
done

if [ "$fail" -eq 0 ]; then
	echo "OK: status index live (every listed file exists; every report indexed)"
fi
exit "$fail"
