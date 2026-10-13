#!/usr/bin/env bash
# Consumer-claims guard (SUPERB v6 T05, G3 posture A: keep the exact pins,
# check them mechanically). AGENTS.md and FEATURES.md make exact claims
# about what the reference consumer (cqrs-htmx `setup`) pins and what this
# repo's example deps require. Those claims went stale repeatedly (the
# setup/v4.13.2 era claim survived three consumer releases), so they are
# machine-checked now:
#
#   1. AGENTS.md consumer line "setup/vX.Y.Z (date, published-tag verified)
#      pins core vX.Y.Z": the named consumer tag exists in the sibling
#      clone, is PUBLISHED on cqrs-htmx origin, is the NEWEST published
#      setup/* tag, and its setup/go.mod really requires the claimed
#      go-appkit version.
#   2. FEATURES.md Consumers block names the SAME consumer tag and pin.
#   3. FEATURES.md security section "(pinned vX.Y.Z" == security/go.mod's
#      example-only go-appkit require.
#   4. (§6) Consumer #2 PapDashboard: AGENTS/FEATURES anchors, newest
#      published v* tag, and the tagged go.mod's go-appkit require.
#
# Scope: dev machines and the release ritual (pre-tag-checks runs this).
# CI does not run it — the content legs need the sibling clones:
#   ../cqrs-htmx   (override: CQRS_HTMX_DIR=/path/to/clone)
#   ../PapDashboard (override: PAPDASHBOARD_DIR=/path/to/clone)
# The published-tag legs use keyless https ls-remote (works everywhere).
#
# Claim-anchor format drift is a FAILURE, never a silent skip: if the
# prose moves, this guard breaks until the regex here is taught the new
# shape (same policy as check-pin-drift.sh's Latest-per-module parser).

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
fail() {
	echo "FAIL: $*"
	status=1
}
ok() { echo "OK:   $*"; }

consumer_dir="${CQRS_HTMX_DIR:-$root/../cqrs-htmx}"
consumer_url="https://github.com/LarsArtmann/cqrs-htmx.git"

# --- 1) Parse the claim anchors ----------------------------------------------
agents_pair="$(sed -n 's/.*setup\/\(v[0-9][0-9.]*\) ([0-9][0-9-]*, published-tag verified) pins core \(v[0-9][0-9.]*\).*/\1 \2/p' AGENTS.md | head -n1)"
if [[ -z "$agents_pair" ]]; then
	fail "AGENTS.md consumer line has no parseable 'setup/vX.Y.Z (date, published-tag verified) pins core vX.Y.Z' anchor (format drift — teach the regex)"
else
	ok "AGENTS.md consumer line parsed: setup/$agents_pair"
fi

features_pin="$(sed -n 's/.*Consumed version: \*\*`go-appkit \(v[0-9][0-9.]*\)`\*\*.*/\1/p' FEATURES.md | head -n1)"
if [[ -z "$features_pin" ]]; then
	fail "FEATURES.md Consumers block has no parseable 'Consumed version: **\`go-appkit vX.Y.Z\`**' anchor"
fi
features_tag="$(sed -n 's/.*PUBLISHED `setup\/\(v[0-9][0-9.]*\)` tag.*/\1/p' FEATURES.md | head -n1)"
if [[ -z "$features_tag" ]]; then
	fail "FEATURES.md Consumers block has no parseable 'PUBLISHED \`setup/vX.Y.Z\` tag' anchor"
fi

security_pin="$(sed -n 's/.*tracks the latest core it needs (pinned \(v[0-9][0-9.]*\).*/\1/p' FEATURES.md | head -n1)"
if [[ -z "$security_pin" ]]; then
	fail "FEATURES.md security section has no parseable '(pinned vX.Y.Z' anchor"
fi

if ((status == 1)); then
	exit "$status"
fi

agents_tag="${agents_pair%% *}"
agents_pin="${agents_pair##* }"

# --- 2) Cross-file consistency ------------------------------------------------
if [[ "$agents_tag" == "$features_tag" && "$agents_pin" == "$features_pin" ]]; then
	ok "AGENTS.md and FEATURES.md agree: setup/$agents_tag pins core $agents_pin"
else
	fail "claim split brain: AGENTS.md says setup/$agents_tag pins core $agents_pin but FEATURES.md says setup/$features_tag / $features_pin"
fi

# --- 3) Published-tag truth (keyless ls-remote) -------------------------------
published="$(git ls-remote "$consumer_url" 'refs/tags/setup/v*' 2>/dev/null | awk -F/ '{print $NF}' | grep -v '\^{}' | sort -Vu || true)"
if [[ -z "$published" ]]; then
	fail "could not list published setup/* tags from $consumer_url (network?)"
	exit "$status"
fi
newest="$(printf '%s\n' "$published" | tail -n1)"
if [[ "$agents_tag" == "$newest" ]]; then
	ok "claimed consumer tag setup/$agents_tag is the newest published setup/* tag"
else
	fail "claimed consumer tag setup/$agents_tag but the newest published is setup/$newest (consumer released; refresh the AGENTS/FEATURES claims)"
fi

# --- 4) Tag content truth (sibling clone) --------------------------------------
if [[ ! -d "$consumer_dir/.git" ]]; then
	fail "cqrs-htmx clone not found at $consumer_dir (set CQRS_HTMX_DIR or clone the sibling)"
	exit "$status"
fi
if ! git -C "$consumer_dir" tag -l -- "setup/$agents_tag" | grep -qxF "setup/$agents_tag"; then
	git -C "$consumer_dir" fetch origin "refs/tags/setup/$agents_tag:refs/tags/setup/$agents_tag" --no-tags 2>/dev/null || true
fi
if ! git -C "$consumer_dir" tag -l -- "setup/$agents_tag" | grep -qxF "setup/$agents_tag"; then
	fail "setup/$agents_tag not present in $consumer_dir even after a fetch (clone stale?)"
	exit "$status"
fi
tagged_pin="$(git -C "$consumer_dir" show "setup/$agents_tag:setup/go.mod" 2>/dev/null | sed -n 's/^[[:space:]]*github\.com\/larsartmann\/go-appkit \(v[0-9][0-9.]*\)[[:space:]]*$/\1/p' | head -n1)"
if [[ -z "$tagged_pin" ]]; then
	fail "setup/$agents_tag:setup/go.mod has no direct github.com/larsartmann/go-appkit require (module layout changed?)"
elif [[ "$tagged_pin" == "$agents_pin" ]]; then
	ok "setup/$agents_tag really pins go-appkit $tagged_pin"
else
	fail "claim says setup/$agents_tag pins core $agents_pin but the tag requires $tagged_pin"
fi

# --- 5) security example-dep claim vs go.mod -----------------------------------
if grep -qE "^[[:space:]]*github\.com/larsartmann/go-appkit ${security_pin//./\\.}[[:space:]]*$" security/go.mod; then
	ok "security example dep really pinned at $security_pin (security/go.mod)"
else
	actual="$(sed -n 's/^[[:space:]]*github\.com\/larsartmann\/go-appkit \(v[0-9][0-9.]*\)[[:space:]]*$/\1/p' security/go.mod | head -n1)"
	fail "FEATURES security claim says pinned $security_pin but security/go.mod requires $actual"
fi

# --- 6) PapDashboard consumer-#2 claims (SUPERB v7 T07, 2026-10-13) -------------
# Same three legs as the cqrs-htmx consumer: AGENTS anchor, FEATURES anchor,
# newest-published-tag truth (keyless ls-remote), and tagged go.mod truth
# (sibling clone; their tags are UN-namespaced v0.x.y on the repo root).
pap_dir="${PAPDASHBOARD_DIR:-$root/../PapDashboard}"
pap_url="https://github.com/LarsArtmann/PapDashboard.git"

pap_pair="$(sed -n 's/.*Consumer #2 ([0-9-]*): PapDashboard \(v[0-9][0-9.]*\) requires go-appkit \(v[0-9][0-9.]*\).*/\1 \2/p' AGENTS.md | head -n1)"
if [[ -z "$pap_pair" ]]; then
	fail "AGENTS.md has no parseable 'Consumer #2 (date): PapDashboard vX.Y.Z requires go-appkit vX.Y.Z' anchor (format drift — teach the regex)"
else
	ok "AGENTS.md consumer-#2 line parsed: PapDashboard $pap_pair"
fi

pap_features="$(sed -n 's/.*`\(v[0-9][0-9.]*\)`\*\* — published.*`go-appkit \(v[0-9][0-9.]*\)`.*/\1 \2/p' FEATURES.md | head -n1)"
if [[ -z "$pap_features" ]]; then
	fail "FEATURES.md Consumers block has no parseable consumer-#2 'vX.Y.Z' + 'go-appkit vX.Y.Z' anchor"
fi

if ((status == 1)); then
	exit "$status"
fi

pap_tag="${pap_pair%% *}"
pap_pin="${pap_pair##* }"

if [[ "$pap_pair" == "$pap_features" ]]; then
	ok "AGENTS.md and FEATURES.md agree: PapDashboard $pap_tag pins go-appkit $pap_pin"
else
	fail "consumer-#2 claim split brain: AGENTS.md says PapDashboard $pap_pair but FEATURES.md says $pap_features"
fi

pap_published="$(git ls-remote "$pap_url" 'refs/tags/v*' 2>/dev/null | awk -F/ '{print $NF}' | grep -v '\^{}' | sort -Vu || true)"
if [[ -z "$pap_published" ]]; then
	fail "could not list published v* tags from $pap_url (network?)"
else
	pap_newest="$(printf '%s\n' "$pap_published" | tail -n1)"
	if [[ "$pap_tag" == "$pap_newest" ]]; then
		ok "claimed consumer-#2 tag $pap_tag is the newest published PapDashboard tag"
	else
		fail "claimed consumer-#2 tag $pap_tag but the newest published is $pap_newest (PapDashboard released; refresh the AGENTS/FEATURES claims)"
	fi
fi

if [[ ! -d "$pap_dir/.git" ]]; then
	fail "PapDashboard clone not found at $pap_dir (set PAPDASHBOARD_DIR or clone the sibling)"
else
	pap_tagged_pin="$(git -C "$pap_dir" show "$pap_tag:go.mod" 2>/dev/null | sed -n 's/^[[:space:]]*github\.com\/larsartmann\/go-appkit \(v[0-9][0-9.]*\)[[:space:]]*$/\1/p' | head -n1)"
	if [[ -z "$pap_tagged_pin" ]]; then
		fail "$pap_tag:go.mod has no direct github.com/larsartmann/go-appkit require (did PapDashboard drop or rename the dep?)"
	elif [[ "$pap_tagged_pin" == "$pap_pin" ]]; then
		ok "$pap_tag really pins go-appkit $pap_tagged_pin"
	else
		fail "claim says PapDashboard $pap_tag pins go-appkit $pap_pin but the tag requires $pap_tagged_pin"
	fi
fi

exit "$status"
