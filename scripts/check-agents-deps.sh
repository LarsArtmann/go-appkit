#!/usr/bin/env bash
# AGENTS.md Core Dependencies table vs go.mod version-drift guard.
# The table drifted once (httputil v1.4.0 / error-family v0.10.1 vs go.mod
# v1.7.0 / v0.12.0, found 2026-11-03 and hand-fixed twice the same day).
# Mechanical truth: every table row's version must equal the root go.mod
# require for that module.
set -euo pipefail
cd "$(dirname "$0")/.."

status=0
table_lines=$(sed -n '/^## Core Dependencies/,/^## /p' AGENTS.md | grep '^| `github' || true)
if [[ -z "$table_lines" ]]; then
	echo "FAIL: no dependency rows found in the AGENTS.md Core Dependencies table"
	exit 1
fi

while IFS=$'\t' read -r module table_version; do
	go_mod_version=$(grep -m1 "^[[:space:]]*${module} " go.mod | awk '{print $2}' || true)
	if [[ -z "$go_mod_version" ]]; then
		echo "FAIL: $module is in the AGENTS table but NOT in go.mod"
		status=1
	elif [[ "$table_version" != "$go_mod_version" ]]; then
		echo "FAIL: $module AGENTS says $table_version, go.mod says $go_mod_version"
		status=1
	else
		echo "OK: $module $go_mod_version"
	fi
done < <(echo "$table_lines" | sed -E 's/^\| `([^`]+)`[[:space:]]*\|[[:space:]]*([^|[:space:]]+)[[:space:]]*\|.*/\1\t\2/')

if ((status == 0)); then
	echo "AGENTS DEPS GATE GREEN"
fi
exit "$status"
