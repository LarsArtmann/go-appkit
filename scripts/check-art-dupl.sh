#!/usr/bin/env bash
# check-art-dupl.sh — clone-regression gate over the committed art-dupl
# baseline (SUPERB v9, 2026-10-18).
#
# The 2026-09-28 and 2026-10-18 audits adjudicated every clone group the
# detector reports at gate settings (`--type-aware -t 1`, default
# test-threshold): each remaining group is an ACCEPTED class (see AGENTS.md
# "Clone posture") or an accepted test-class variant. This gate makes that
# acceptance STANDING instead of prose: `art-dupl check` reports only clone
# groups whose hash is NOT in scripts/art-dupl-baseline.json, so
# NEW duplication fails the tag gate while the accepted classes stay green.
# (The baseline lives in scripts/ — NOT docs/reports/ — because .gitignore's
# any-depth `reports/` pattern silently eats that path; caught 2026-10-19
# when the pushed tree's guard referenced an untrackable file. Human-context
# snapshot: scripts/art-dupl-gate-snapshot-2026-10-18.txt.)
#
# Scope note: at default test-threshold (30 tokens) small test-file clones
# are below the gate — the periodic manual audit (`art-dupl . --type-aware
# -t 1 --test-threshold 1 --rich-text`) remains the deeper pass. Gate
# settings deliberately mirror go-etag's pre-release check (scripts/
# pre-release-check.sh §OQ8), the family precedent.
#
# Re-baselining (extract-or-accept): if a new group is INTENTIONAL, rerun
#   art-dupl baseline . --type-aware -t 1 \
#     --baseline-path docs/reports/dupl/art-dupl-baseline.json
# and refresh docs/reports/dupl/*-gate-settings-snapshot.txt — and say why
# in the commit message. A baseline bump without a justification will be
# flagged in review.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

baseline="scripts/art-dupl-baseline.json"

if ! command -v art-dupl >/dev/null 2>&1; then
	echo "FAIL: art-dupl binary not found in PATH."
	echo "  install: go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest (public repo)"
	exit 1
fi

if [ ! -f "$baseline" ]; then
	echo "FAIL: baseline $baseline missing — regenerate it (see this script's header)."
	exit 1
fi

out="$(mktemp)"
trap 'rm -f "$out"' EXIT

if ! art-dupl check . --type-aware -t 1 --baseline-path "$baseline" >"$out" 2>&1; then
	echo "FAIL: new clone groups detected against the committed baseline:"
	sed -n '1,40p' "$out"
	echo ""
	echo "  fix: extract the duplication, or re-baseline DELIBERATELY (header of $0)"
	exit 1
fi

echo "OK: no new clone groups vs scripts/art-dupl-baseline.json (gate: --type-aware -t 1)"
