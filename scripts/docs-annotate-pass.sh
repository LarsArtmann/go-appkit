#!/usr/bin/env bash
# docs-annotate-pass.sh — the formatter canon as one command.
#
# The docs-health annotation ritual is four steps that MUST run as one
# atomic pass (a formatter is a cooperating rewriter that does not know
# about verdict markers; skipping the second check-rows is exactly how a
# formatter pass damaged seven archived files before the gate caught it):
#
#     annotate  ->  check-rows  ->  format  ->  check-rows
#
# Usage:
#   ./scripts/docs-annotate-pass.sh <file.md> [<file.md>...]   # default: DRY-RUN
#   ./scripts/docs-annotate-pass.sh --apply <file.md> ...      # run the formatter
#   ./scripts/docs-annotate-pass.sh --check <file.md> ...      # check-rows only
#
# Exit codes: 0 every gate green · 1 a check-rows gate failed (the file(s)
# named in the output are the offenders — per-cell strikes only, never a
# strike across pipes; repair with the annotate tooling, not by hand).
#
# The annotate binary lives in the docs-health skill's assets dir (path
# pinned like AGENTS pins its tool paths; override with $ANNOTATE_BIN).

set -uo pipefail

ANNOTATE_BIN="${ANNOTATE_BIN:-/home/lars/.config/crush/skills/docs-health/assets/annotate}"
FORMATTER="${FORMATTER:-dprint}"

mode="dry"
files=()
for arg in "$@"; do
  case "$arg" in
    --apply) mode="apply" ;;
    --check) mode="check" ;;
    -*) echo "unknown flag: $arg" >&2; exit 2 ;;
    *) files+=("$arg") ;;
  esac
done

if [ "${#files[@]}" -eq 0 ]; then
  echo "usage: $0 [--apply|--check] <file.md> [<file.md>...]" >&2
  exit 2
fi

for f in "${files[@]}"; do
  [ -f "$f" ] || { echo "not a file: $f" >&2; exit 2; }
done

fail=0
for f in "${files[@]}"; do
  if ! "$ANNOTATE_BIN" check-rows "$f" > /tmp/dap-rows-1.log 2>&1; then
    echo "GATE-FAIL (pre-format) $f"; cat /tmp/dap-rows-1.log; fail=1; continue
  fi
  if [ "$mode" = "check" ]; then
    echo "OK (check-only) $f"; continue
  fi
  if [ "$mode" = "apply" ]; then
    if ! "$FORMATTER" fmt "$f" > /tmp/dap-fmt.log 2>&1; then
      echo "FORMAT-FAIL $f"; cat /tmp/dap-fmt.log; fail=1; continue
    fi
  else
    echo "DRY-RUN (no format) $f — rerun with --apply to format"
  fi
  if ! "$ANNOTATE_BIN" check-rows "$f" > /tmp/dap-rows-2.log 2>&1; then
    echo "GATE-FAIL (post-format) $f"; cat /tmp/dap-rows-2.log; fail=1; continue
  fi
  echo "OK $f"
done

exit "$fail"
