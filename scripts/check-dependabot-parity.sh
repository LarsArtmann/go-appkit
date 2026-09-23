#!/usr/bin/env bash
# Config-parity guard. Every Go module directory must be covered by BOTH
# dependabot (weekly gomod updates) and the CI test matrix, and no dependabot
# entry may name a directory that has no go.mod. This makes the hand-run
# 2026-09-17 "/security is byte-identical" check permanent — the class of
# drift it guards (a module added without its dependabot entry or matrix
# slot) previously surfaced only by manual comparison:
#
#   ./scripts/check-dependabot-parity.sh
#
# CI runs it in the config-parity job; locally it needs nothing but the
# checkout and finishes in well under a second.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
fail() {
	echo "FAIL: $*"
	status=1
}
ok() { echo "OK:   $*"; }

# Module directories on disk; root normalized to ".".
moddirs="$(find . -name go.mod -not -path '*/vendor/*' -exec dirname {} \; | sed 's|^\.$|.|; s|^\./||' | sort)"

# Dependabot gomod directories (entry-scoped: ecosystem then directory).
depdirs="$(awk '
	/^[[:space:]]*- package-ecosystem:/ { eco=$(NF) }
	/^[[:space:]]*directory:/ { dir=$2 }
	eco == "gomod" && dir != "" { print dir; dir="" }
' .github/dependabot.yml | sort -u)"

# CI test-matrix module entries (the "." / ./module style list in ci.yml).
cimods="$(sed -n '/matrix:/,/]/p' .github/workflows/ci.yml | tr -d ' ",' | grep -E '^\.$|^\./' | sort -u)"

for dir in $moddirs; do
	entry="."
	depdir="/"
	if [[ "$dir" != "." ]]; then
		entry="./$dir"
		depdir="/$dir"
	fi

	if grep -qxF "$entry" <<<"$cimods"; then
		ok "$dir has a CI matrix slot ($entry)"
	else
		fail "$dir has no CI test-matrix slot (want $entry in .github/workflows/ci.yml)"
	fi

	if grep -qxF "$depdir" <<<"$depdirs"; then
		ok "$dir has a dependabot gomod entry ($depdir)"
	else
		fail "$dir has no dependabot gomod entry (want directory: $depdir in .github/dependabot.yml)"
	fi
done

while read -r depdir; do
	[[ -n "$depdir" ]] || continue
	if [[ "$depdir" == "/" ]]; then
		dir="."
	else
		dir="${depdir#/}"
	fi
	if [[ ! -f "$dir/go.mod" ]]; then
		fail "dependabot entry $depdir names a directory with no go.mod"
	fi
done <<<"$depdirs"

exit "$status"
