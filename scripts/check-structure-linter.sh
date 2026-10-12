#!/usr/bin/env bash
# Structure-linter config-loss guard (2026-10-12): go-structure-linter exits 0
# even when it REPORTS findings, so a silently lost config (the yaml renamed,
# deleted, or its exclude_patterns emptied) would stay green in CI while the
# 11 deliberate-layout findings come back. This guard fails on any non-zero
# findings total or a missing config file. The committed
# .go-structure-linter.yaml carries the three layout-rule exclusions
# (root-package-files, internal-directory, examples-directory) — the root
# package IS the public `appkit` package; see AGENTS.md "Structure linting".
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ ! -f .go-structure-linter.yaml ]]; then
	echo "FAIL: .go-structure-linter.yaml is missing — the layout-rule exclusions are config-carried; a bare run reports 11 false findings on this deliberately-flat repo"
	exit 1
fi

out="$(go-structure-linter . 2>&1)" || {
	echo "FAIL: go-structure-linter did not run:"
	echo "$out" | head -5
	exit 1
}

if grep -qE 'total=[1-9][0-9]*' <<<"$out"; then
	echo "FAIL: go-structure-linter reported findings (config loss or real structural drift):"
	grep -E 'total=|WARN|ERROR' <<<"$out" | head -20
	exit 1
fi

echo "OK: go-structure-linter reports 0 findings with the committed config"
