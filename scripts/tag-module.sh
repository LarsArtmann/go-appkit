#!/usr/bin/env bash
# tag-module.sh — the release dance (AGENTS Release Ritual steps 2-5) as one
# command (2026-11-06, issue54 session §e9: "tag → tidy → go.sum commit →
# re-tag is mechanical but easy to order wrong; I moved a tag twice").
#
# What it does, in order:
#   1. Refuses on a dirty tree (except when --tidy-integration staged the
#      integration go.mod/go.sum update itself) — a dirty tree under the tag
#      is the mid-train class the daemon makes common.
#   2. (--tidy-integration) resolves integration against the NEW tag through
#      the LOCAL repo: sets the url.insteadOf rewrite (AGENTS note 4 — the
#      fleet is GOPRIVATE, tidy resolves tags via VCS, and a not-yet-pushed
#      tag is invisible to the remote), runs `go mod tidy` hermetically,
#      stages integration/go.mod + go.sum, removes the rewrite. YOU commit
#      the staged integration bump before pushing (one commit, message
#      names the pin bump).
#   3. Creates each annotated tag with the --message (refuses if the tag
#      exists — "exactly once").
#   4. Runs scripts/pre-tag-checks.sh over every tag (all standing guards +
#      tagged artifacts + tagged-CHANGELOG dating).
#   5. Pushes master + tags, unless --no-push.
#
# What it deliberately does NOT do: date the CHANGELOG, update AGENTS/
# FEATURES release state, or run the API-break diff — those are judgment
# calls the ritual keeps manual (steps 1-3); this script is the mechanical
# tail. Run it only after they are done.
#
# Usage:
#   scripts/tag-module.sh --message "<semantic delta>" <tag> [<tag>...]
#   scripts/tag-module.sh --message "..." --no-push <tag>       # dry tail
#   scripts/tag-module.sh --message "..." --tidy-integration health/vX.Y.Z
#
# Exit codes: 0 pushed (or --no-push green) · 1 any step failed · 2 usage.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

message=""
push=true
tidy_integration=false
tags=()
while (($# > 0)); do
	case "$1" in
	--message)
		message="${2:?--message needs a value}"
		shift 2
		;;
	--no-push)
		push=false
		shift
		;;
	--tidy-integration)
		tidy_integration=true
		shift
		;;
	-h* | --help)
		sed -n '2,32p' "$0"
		exit 0
		;;
	-*)
		echo "unknown flag: $1"
		exit 2
		;;
	*)
		tags+=("$1")
		shift
		;;
	esac
done

if ((${#tags[@]} == 0)); then
	echo "usage: $0 --message \"<semantic delta>\" <tag> [<tag>...] [--no-push] [--tidy-integration]"
	exit 2
fi
if [[ -z "$message" ]]; then
	echo "FAIL: --message is required (annotated tags state the semantic delta — docs: prefix for doc-only releases)"
	exit 2
fi

for tag in "${tags[@]}"; do
	if git cat-file -e "${tag}^{commit}" 2>/dev/null; then
		echo "FAIL: tag $tag already exists — annotated tags are created exactly once"
		exit 1
	fi
done

# 1) Tree hygiene. A staged integration go.mod/go.sum pair is the expected
#    mid-dance state after step 2; everything else must be clean.
dirty="$(git status --porcelain)"
if [[ -n "$dirty" && "$tidy_integration" != true ]]; then
	echo "FAIL: tree is dirty — commit (or stage via --tidy-integration) first:"
	echo "$dirty"
	exit 1
fi

floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"

if [[ "$tidy_integration" == true ]]; then
	echo "== tidy integration against the local tag (insteadOf rewrite) =="
	git config --local url."$root".insteadOf "https://github.com/larsartmann/go-appkit"
	restore_rewrite() {
		git config --local --unset url."$root".insteadOf || true
	}
	trap restore_rewrite EXIT
	(
		cd integration
		GOWORK=off GOTOOLCHAIN="go$floor" GOFLAGS='' GOPRIVATE='' go mod tidy
	)
	trap - EXIT
	restore_rewrite
	git add integration/go.mod integration/go.sum
	echo "STAGED: integration/go.mod + go.sum — commit this pin bump now, then re-run without --tidy-integration"
	if [[ -n "$(git status --porcelain -- ':!integration/go.mod' ':!integration/go.sum')" ]]; then
		echo "FAIL: other files are still dirty after the tidy — resolve before tagging:"
		git status --porcelain -- ':!integration/go.mod' ':!integration/go.sum'
		exit 1
	fi
	if [[ "$(git status --porcelain -- integration/go.mod integration/go.sum | grep -c '^[AM]')" == "0" ]]; then
		echo "NOTE: tidy produced no integration changes (pin already current) — continuing"
	fi
fi

# 3) Annotated tags, exactly once each.
for tag in "${tags[@]}"; do
	echo "== tag $tag =="
	git tag -a "$tag" -m "$message"
done

# 4) The full pre-tag gate (guards + tagged artifacts + tagged CHANGELOGs).
./scripts/pre-tag-checks.sh "${tags[@]}"

# 5) Push.
if [[ "$push" == true ]]; then
	git push origin master "${tags[@]}"
	echo "PUSHED: master ${tags[*]}"
else
	echo "GREEN (no push): ${tags[*]} — push with: git push origin master ${tags[*]}"
fi
