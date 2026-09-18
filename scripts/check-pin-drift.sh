#!/usr/bin/env bash
# Pin-drift guard (SUPERB plan v4 C1). Both 2026-09-17 drift incidents were
# findable by one grep and nothing ran it; this script is that grep, run
# permanently:
#
#   1. The AGENTS.md release line ("ON ORIGIN through **core vX.Y.Z**") names a
#      tag that actually exists (catches: release shipped while AGENTS still
#      named the previous version).
#   2. integration/go.mod pins the LATEST published tag of every go-appkit
#      family module it requires (catches: release train without the
#      integration pin bump — integration must always test exactly what a
#      fresh consumer resolves from the proxy).
#   3. Root go.mod and go.work declare the same go directive (catches: the
#      accidental go 1.27.1 directive bump that broke every local
#      workspace command on 2026-09-17).
#
# CI runs this from a fetch-depth: 0 checkout (check 1 and 2 need full tag
# history). Locally it runs against the local clone in under a second:
#
#   ./scripts/check-pin-drift.sh

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
fail() { echo "FAIL: $*"; status=1; }
ok() { echo "OK:   $*"; }

# 1) AGENTS release line vs git tags.
core_version="$(sed -n 's/.*ON ORIGIN through \*\*core \(v[0-9][0-9.]*\)\*\*.*/\1/p' AGENTS.md | head -n1)"
if [[ -z "$core_version" ]]; then
	fail "AGENTS.md has no 'ON ORIGIN through **core vX.Y.Z**' release line"
elif git tag -l -- "$core_version" | grep -qxF "$core_version"; then
	ok "AGENTS release line $core_version exists as a tag"
else
	fail "AGENTS release line says $core_version but git tag -l has no such tag"
fi

# 2) integration pins == latest published family tag.
go_mod="integration/go.mod"
if [[ ! -f "$go_mod" ]]; then
	fail "$go_mod missing"
	exit 1
fi
pins="$(awk '/^require \(/,/^\)/' "$go_mod" | awk '$1 ~ /^github\.com\/larsartmann\/go-appkit/ && $0 !~ /\/\/ indirect/ {print $1 " " $2}')"
if [[ -z "$pins" ]]; then
	fail "no direct go-appkit family requires found in $go_mod"
fi
while read -r mod want; do
	[[ -n "$mod" ]] || continue
	sub="${mod#github.com/larsartmann/go-appkit}"
	sub="${sub#/}"
	if [[ -z "$sub" ]]; then
		pattern="v*"
	else
		pattern="$sub/v*"
	fi
	latest="$(git tag -l -- "$pattern" | sort -V | tail -n1)"
	if [[ -z "$latest" ]]; then
		fail "$mod: no tags matching $pattern"
	elif [[ "$want" == "$latest" ]]; then
		ok "$mod pinned at latest published tag $want"
	else
		fail "$mod pinned at $want but latest published tag is $latest (release train shipped without the integration pin bump?)"
	fi
done <<< "$pins"

# 3) root go.mod vs go.work go directive.
gomod_go="$(awk '$1 == "go" { print $2; exit }' go.mod)"
gowork_go="$(awk '$1 == "go" { print $2; exit }' go.work)"
if [[ -z "$gomod_go" || -z "$gowork_go" ]]; then
	fail "could not read the go directive from go.mod/go.work"
elif [[ "$gomod_go" == "$gowork_go" ]]; then
	ok "go.mod and go.work agree on go $gomod_go"
else
	fail "go.mod says go $gomod_go but go.work says go $gowork_go (workspace mismatch — every local workspace command fails)"
fi

exit "$status"
