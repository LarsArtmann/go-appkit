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
#   4. The SigNoz dashboard artifact exists at its documented path and
#      parses as JSON (catches: rename/delete while README/FEATURES still
#      point at it).
#
# Release-window rule (2026-11-06, issue54 session §e1): "newest tag exists
# locally but not on origin" is the NORMAL state between `git tag -a` and
# `git push` (the pre-tag ritual runs this gate in exactly that window), so
# it is a loud NOTE, not drift. Only a claim/pin that matches NEITHER a
# local nor an origin tag is a failure; when origin is unreachable the
# origin leg degrades to local-only verification with its own loud line.
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
note() { echo "NOTE: $*"; }

# --- origin tag state (one roundtrip) ----------------------------------------
# origin_tags: newline-separated tag names known to origin. Empty means the
# ls-remote failed (offline); the origin leg then DEGRADES to local-only
# verification with a loud line instead of failing every claim.
origin_tags="$(git ls-remote --tags origin 2>/dev/null | sed -n 's|.*refs/tags/\([^^]*\)$|\1|p' || true)"
origin_reachable=1
if [[ -z "$origin_tags" ]]; then
	note "origin unreachable (ls-remote failed) — release-window check degraded to local-only verification"
	origin_reachable=0
fi

tag_on_origin() {
	local tag="$1"
	[[ "$origin_reachable" == 1 ]] || return 1
	grep -qxF "$tag" <<<"$origin_tags"
}

# tag_state <full-tag-name>: prints "origin" | "local-only" | "absent".
tag_state() {
	local tag="$1"
	if git cat-file -e "${tag}^{commit}" 2>/dev/null; then
		if tag_on_origin "$tag"; then
			echo origin
		else
			echo local-only
		fi
	else
		echo absent
	fi
}

# 1) AGENTS release line vs git tags.
core_version="$(sed -n 's/.*ON ORIGIN through \*\*core \(v[0-9][0-9.]*\)\*\*.*/\1/p' AGENTS.md | head -n1)"
if [[ -z "$core_version" ]]; then
	core_version="$(sed -n 's/.*Latest per module: core \(v[0-9][0-9.]*\),.*/\1/p' AGENTS.md | head -n1)"
fi
if [[ -z "$core_version" ]]; then
	fail "AGENTS.md has no 'ON ORIGIN through **core vX.Y.Z**' or 'Latest per module: core vX.Y.Z' release line"
elif git tag -l -- "$core_version" | grep -qxF "$core_version"; then
	state="$(tag_state "$core_version")"
	case "$state" in
	origin) ok "AGENTS release line $core_version exists as a tag on origin" ;;
	local-only)
		ok "AGENTS release line $core_version exists locally"
		note "$core_version is LOCAL-ONLY (not yet on origin) — release window: push before consumers can resolve it"
		;;
	absent) fail "AGENTS release line names $core_version but the tag is in an unreachable state" ;;
	esac
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
	cqrs | otel | flightrecorder | health | realtime | docs | errorpages | security | systemd | config | preset) echo "$1/v*" ;;
	frh) echo "flightrecorderhealth/v*" ;;
	*) return 1 ;;
	esac
}

latest_line="$(sed -n 's/.*Latest per module: //p' AGENTS.md | head -n1)"
if [[ -z "$latest_line" ]]; then
	fail "AGENTS.md has no 'Latest per module: ...' line (per-module claims unparseable)"
else
	# The list ends the parenthetical release sentence; cut at the last ')'.
	claims="${latest_line%%).*}"
	# STRICT parser (2026-11-06 rewrite, issue54 §e3): every entry must be
	# exactly "name vX.Y.Z" — the old aggregate-count check ("15 entries but
	# 12 parse") never named the corrupting token, so prose inside the list
	# silently corrupted the claim set. Name every offender.
	pairs=""
	offenders=""
	while IFS= read -r entry; do
		entry="$(printf '%s' "$entry" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
		[[ -n "$entry" ]] || continue
		if [[ "$entry" =~ ^[a-z]+[[:space:]]+v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
			pairs+="$entry"$'\n'
		else
			offenders+="  '$entry' (expected 'name vX.Y.Z')"$'\n'
		fi
	done < <(printf '%s' "$claims" | tr ',' '\n')
	if [[ -n "$offenders" ]]; then
		fail "Latest-per-module line carries unparseable entries (strict list):"
		printf '%s' "$offenders"
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
			state="$(tag_state "$tag_latest")"
			case "$state" in
			origin) ok "$name claim $version == newest tag $tag_latest (on origin)" ;;
			local-only)
				ok "$name claim $version == newest tag $tag_latest (local)"
				note "$tag_latest is LOCAL-ONLY (not yet on origin) — release window: push before consumers can resolve it"
				;;
			absent) fail "$name: claim $version verified against $tag_latest which is in an unreachable state" ;;
			esac
		fi
	done <<<"$pairs"

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
			state="$(tag_state "$latest")"
			case "$state" in
			origin) ok "$mod pinned at latest published tag $latest (on origin)" ;;
			local-only)
				ok "$mod pinned at latest tag $latest (local)"
				note "$latest is LOCAL-ONLY (not yet on origin) — release window: the pin resolves from the proxy only after push"
				;;
			absent) fail "$mod: pin verified against $latest which is in an unreachable state" ;;
			esac
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
#
#     held_pins: DELIBERATE holds, not drift — a recorded decision with an
#     unblock trigger. Green only while the pin is EXACTLY the held version:
#     bumping the pin without removing the hold (or vice versa) goes red, so
#     the two can never drift apart silently. 2026-11-06 hold, empirically
#     proven in scratch runs: go-sse v0.7.0 (and cqrs-htmx v4.16.0, which
#     requires it) pull a-h/templ v0.3.1100, whose removed
#     templ.ResolveAttributeValue breaks generated code in templ-components
#     <= v1.26.0 AND go-cqrs-lite catalog <= v4.12.0 (latest). templ-components
#     v1.27.0 shipped the migration; catalog has not. Unblock trigger: catalog
#     ships a migrated tag AND cqrs-htmx requires templ-components >= v1.27.0 —
#     then bump cqrs-htmx + go-sse + templ-components together and delete the
#     hold entries here (keep the documentedPins comment in sync).
declare -A held_pins=(
	[github.com/larsartmann/cqrs-htmx/v4]=v4.15.0
	[github.com/larsartmann/go-sse]=v0.6.2
)
for mod in github.com/larsartmann/cqrs-htmx/v4 github.com/larsartmann/go-sse github.com/larsartmann/go-sse/ssetest github.com/larsartmann/httputil; do
	want="$(awk '/^require \(/,/^\)/' "$go_mod" | awk -v m="$mod" '$1 == m && $0 !~ /\/\/ indirect/ {print $2}')"
	if [[ -z "$want" ]]; then
		fail "$go_mod has no direct require for $mod"
		continue
	fi
	# Self-pin the toolchain: derive the floor from this tree's root go.mod
	# (train-proof — the 2026-10-12 1.27.2 train made the hard-coded go1.27.1
	# prefix fail silently: cqrs-htmx v4.14.0 requires >= 1.27.2, `go list`
	# died with "requires go >= 1.27.2", swallowed by 2>/dev/null, and every
	# cross-repo leg reported "could not query"). An ambient GOTOOLCHAIN=local
	# on the 1.26.7 machine default kills `go list` the same way.
	repo_floor="$(sed -n 's/^go //p' go.mod | head -n1)"
	latest="$(GOWORK=off GOTOOLCHAIN="go$repo_floor" go list -m -versions "$mod" 2>/dev/null | tr ' ' '\n' | grep -E '^v[0-9]+\.[0-9]+' | sort -V | tail -n1 || true)"
	if [[ -z "$latest" ]]; then
		fail "$mod: could not query the module proxy for published versions"
	elif [[ "$want" == "$latest" ]]; then
		ok "$mod pinned at latest published version $latest"
	elif [[ "${held_pins[$mod]:-}" == "$want" ]]; then
		ok "$mod HELD at $want (recorded decision; latest $latest — see held_pins comment)"
	elif [[ -n "${held_pins[$mod]:-}" ]]; then
		fail "$mod pinned at $want but the recorded hold is ${held_pins[$mod]} (pin and hold drifted apart — update both in one change)"
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

# 4) The SigNoz dashboard artifact exists at its documented path and parses
#    as JSON. The README, dashboards/README, FEATURES, and the deep-dive
#    report all name otel/dashboards/appkit-http-dashboard.json — a rename
#    or deletion must fail loudly here, not break a documented import
#    instructions trail.
dashboard="otel/dashboards/appkit-http-dashboard.json"
if [[ ! -f "$dashboard" ]]; then
	fail "SigNoz dashboard artifact missing at $dashboard (rename recorded in README/dashboards/README/FEATURES?)"
elif ! jq -e '. != null' "$dashboard" >/dev/null 2>&1; then
	fail "$dashboard does not parse as JSON"
else
	schema="$(jq -r '.schemaVersion // "missing"' "$dashboard" 2>/dev/null)"
	ok "SigNoz dashboard artifact $dashboard present, JSON-valid (schemaVersion: $schema)"
fi

exit "$status"
