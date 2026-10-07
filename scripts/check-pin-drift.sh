#!/usr/bin/env bash
# Pin-drift guard (SUPERB plan v4 C1). Both 2026-09-17 drift incidents were
# findable by one grep and nothing ran it; this script is that grep, run
# permanently:
#
#   1. The AGENTS.md release line ("ON ORIGIN through **core vX.Y.Z**" or the
#      "Latest per module: core vX.Y.Z, ..." form) names a tag that actually
#      exists (catches: release shipped while AGENTS still named the previous
#      version).
#   1b. EVERY module's claim in the "Latest per module" list equals that
#      module's newest tag, and every tagged module is claimed (catches the
#      2026-10-06 class — family train shipped while AGENTS still said
#      "pending" — for all modules, not just core).
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
fail() {
	echo "FAIL: $*"
	status=1
}
ok() { echo "OK:   $*"; }

# 1) AGENTS release line vs git tags.
core_version="$(sed -n 's/.*ON ORIGIN through \*\*core \(v[0-9][0-9.]*\)\*\*.*/\1/p' AGENTS.md | head -n1)"
if [[ -z "$core_version" ]]; then
	core_version="$(sed -n 's/.*Latest per module: core \(v[0-9][0-9.]*\),.*/\1/p' AGENTS.md | head -n1)"
fi
if [[ -z "$core_version" ]]; then
	fail "AGENTS.md has no 'ON ORIGIN through **core vX.Y.Z**' or 'Latest per module: core vX.Y.Z' release line"
elif git tag -l -- "$core_version" | grep -qxF "$core_version"; then
	ok "AGENTS release line $core_version exists as a tag"
else
	fail "AGENTS release line says $core_version but git tag -l has no such tag"
fi

# 1b) Per-module claims: every "name vX.Y.Z" in the "Latest per module" list
#     must equal that module's newest tag, and every tagged module must be
#     claimed. Catches the 2026-10-06 class (family train shipped while
#     AGENTS still said "pending") for EVERY module, mechanically — the
#     same-train release-state rule (Release Ritual step 3) as code.
tag_pattern() {
	case "$1" in
	core) echo "v*" ;;
	cqrs | otel | flightrecorder | health | realtime | docs | errorpages | security | systemd) echo "$1/v*" ;;
	frh) echo "flightrecorderhealth/v*" ;;
	*) return 1 ;;
	esac
}

latest_line="$(sed -n 's/.*Latest per module: //p' AGENTS.md | head -n1)"
if [[ -z "$latest_line" ]]; then
	fail "AGENTS.md has no 'Latest per module: ...' line (per-module claims unparseable)"
else
	# The list ends the parenthetical release sentence; cut at the first ').'
	claims="${latest_line%%).*}"
	pairs="$(printf '%s\n' "$claims" | tr ',' '\n' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
	entry_count="$(printf '%s\n' "$pairs" | grep -c . || true)"
	pair_count="$(printf '%s\n' "$pairs" | awk 'NF == 2 && $2 ~ /^v[0-9]+\.[0-9]+\.[0-9]+$/ {n++} END {print n + 0}')"
	if [[ "$entry_count" != "$pair_count" ]]; then
		fail "Latest-per-module line has $entry_count entries but only $pair_count parse as 'name vX.Y.Z' (format drift?)"
	fi
	declare -A claimed=()
	while read -r name version; do
		[[ -n "$name" ]] || continue
		if ! pattern="$(tag_pattern "$name")"; then
			fail "unknown module short name '$name' in the Latest-per-module line (extend tag_pattern in check-pin-drift.sh)"
			continue
		fi
		claimed["$name"]=1
		tag_latest="$(git tag -l -- "$pattern" | sort -V | tail -n1)"
		tag_version="${tag_latest#*/}"
		if [[ -z "$tag_latest" ]]; then
			fail "$name: AGENTS claims $version but no tags match $pattern (claim ahead of a tag that does not exist)"
		elif [[ "$version" != "$tag_version" ]]; then
			fail "$name: AGENTS claims $version but the newest tag is $tag_latest (same-train release-state update missed?)"
		else
			ok "$name claim $version == newest tag $tag_latest"
		fi
	done <<<"$(printf '%s\n' "$pairs" | awk 'NF == 2 && $2 ~ /^v[0-9]+\.[0-9]+\.[0-9]+$/')"

	# Reverse: every TAGGED module must appear in the claims (module set from
	# go.work use entries — the charter guard keeps that list honest).
	while read -r use_entry; do
		case "$use_entry" in
		".") dir_name="core" ;;
		./integration) continue ;;
		.*) dir_name="${use_entry#./}" ;;
		*) continue ;;
		esac
		short="$dir_name"
		[[ "$dir_name" == "flightrecorderhealth" ]] && short="frh"
		if ! pattern="$(tag_pattern "$short" 2>/dev/null)"; then
			fail "go.work member $use_entry is missing from the pin-guard module map (extend tag_pattern)"
			continue
		fi
		tag_latest="$(git tag -l -- "$pattern" | sort -V | tail -n1)"
		if [[ -n "$tag_latest" && -z "${claimed[$short]:-}" ]]; then
			fail "$dir_name has tags (newest $tag_latest) but is not claimed in the Latest-per-module line"
		fi
	done < <(awk '/^use \(/{f=1; next} /^\)/{f=0} f && NF {print $1}' go.work)
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
	else
		latest_version="${latest#"$sub/"}"
		if [[ "$want" == "$latest_version" ]]; then
			ok "$mod pinned at latest published tag $latest"
		else
			fail "$mod pinned at $want but latest published tag is $latest (release train shipped without the integration pin bump?)"
		fi
	fi
done <<<"$pins"

# 2b) Cross-repo contract pins (cqrs-htmx, go-sse, ssetest, httputil) ==
#     newest PUBLISHED version. The 2026-09-28 audit found cqrs-htmx pinned
#     at v4.9.0 while v4.12.0 was latest — nothing guarded cross-repo pins,
#     so integration was NOT testing what a fresh consumer resolves. These
#     live in other repos, so the newest version comes from the module proxy
#     (needs network; CI and dev machines have it).
for mod in github.com/larsartmann/cqrs-htmx/v4 github.com/larsartmann/go-sse github.com/larsartmann/go-sse/ssetest github.com/larsartmann/httputil; do
	want="$(awk '/^require \(/,/^\)/' "$go_mod" | awk -v m="$mod" '$1 == m && $0 !~ /\/\/ indirect/ {print $2}')"
	if [[ -z "$want" ]]; then
		fail "$go_mod has no direct require for $mod"
		continue
	fi
	# Self-pin the toolchain: this tree's go.mods declare go 1.27.1, so an
	# ambient GOTOOLCHAIN=local on the 1.26.7 machine default kills every
	# `go list` here with "requires go >= 1.27.1" (swallowed by 2>/dev/null).
	latest="$(GOWORK=off GOTOOLCHAIN=go1.27.1 go list -m -versions "$mod" 2>/dev/null | tr ' ' '\n' | grep -E '^v[0-9]+\.[0-9]+' | sort -V | tail -n1 || true)"
	if [[ -z "$latest" ]]; then
		fail "$mod: could not query the module proxy for published versions"
	elif [[ "$want" == "$latest" ]]; then
		ok "$mod pinned at latest published version $latest"
	else
		fail "$mod pinned at $want but latest published version is $latest (bump integration/go.mod + the documentedPins fixture together)"
	fi
done

# 3) root go.mod vs go.work go directive.
gomod_go="$(awk '$1 == "go" { print $2; exit }' go.mod 2>/dev/null || true)"
gowork_go="$(awk '$1 == "go" { print $2; exit }' go.work 2>/dev/null || true)"
if [[ -z "$gomod_go" || -z "$gowork_go" ]]; then
	fail "could not read the go directive from go.mod/go.work"
elif [[ "$gomod_go" == "$gowork_go" ]]; then
	ok "go.mod and go.work agree on go $gomod_go"
else
	fail "go.mod says go $gomod_go but go.work says go $gowork_go (workspace mismatch — every local workspace command fails)"
fi

exit "$status"
