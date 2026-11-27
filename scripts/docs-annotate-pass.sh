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
# The annotate helper is the docs-health skill's SINGLE-FILE Go source, run
# with `go run` (the prebuilt assets/annotate binary was deleted 2026-11-15:
# it sat 12 days behind its source and missed four subcommands). Override the
# home with $DOCS_HEALTH_SKILL_DIR or the source path with $ANNOTATE_GO.

set -uo pipefail

SKILL_DIR="${DOCS_HEALTH_SKILL_DIR:-/home/lars/.config/crush/skills/docs-health}"
ANNOTATE_GO="${ANNOTATE_GO:-$SKILL_DIR/assets/annotate.go}"
FORMATTER="${FORMATTER:-dprint}"

if [ -n "${ANNOTATE_BIN:-}" ]; then
	echo "docs-annotate-pass: ANNOTATE_BIN is retired (the prebuilt binary is gone, 2026-11-15) — set ANNOTATE_GO to the source path instead" >&2
	exit 2
fi

# The read-only row gate is the judge's: `docs-gates rows` (parity with
# the writer's check-rows proven 2026-11-27 on clean, PARTIAL and missing
# files). The writer source stays as the fallback where no judge is installed.
annotate() {
	if [ "$1" = check-rows ] && command -v docs-gates >/dev/null 2>&1; then
		docs-gates rows "${@:2}"
		return $?
	fi
	if [ ! -f "$ANNOTATE_GO" ]; then
		echo "docs-annotate-pass: helper missing at $ANNOTATE_GO — the gate did NOT run (install the docs-health skill or set ANNOTATE_GO)" >&2
		exit 2
	fi
	if ! command -v go >/dev/null 2>&1; then
		echo "docs-annotate-pass: go not on PATH — the helper runs as 'go run'" >&2
		exit 2
	fi
	go run "$ANNOTATE_GO" "$@"
}

mode="dry"
files=()
for arg in "$@"; do
	case "$arg" in
	--apply) mode="apply" ;;
	--check) mode="check" ;;
	-*)
		echo "unknown flag: $arg" >&2
		exit 2
		;;
	*) files+=("$arg") ;;
	esac
done

if [ "${#files[@]}" -eq 0 ]; then
	echo "usage: $0 [--apply|--check] <file.md> [<file.md>...]" >&2
	exit 2
fi

for f in "${files[@]}"; do
	[ -f "$f" ] || {
		echo "not a file: $f" >&2
		exit 2
	}
done

fail=0
for f in "${files[@]}"; do
	if ! annotate check-rows "$f" >/tmp/dap-rows-1.log 2>&1; then
		echo "GATE-FAIL (pre-format) $f"
		cat /tmp/dap-rows-1.log
		fail=1
		continue
	fi
	if [ "$mode" = "check" ]; then
		echo "OK (check-only) $f"
		continue
	fi
	if [ "$mode" = "apply" ]; then
		if ! "$FORMATTER" fmt "$f" >/tmp/dap-fmt.log 2>&1; then
			echo "FORMAT-FAIL $f"
			cat /tmp/dap-fmt.log
			fail=1
			continue
		fi
	else
		echo "DRY-RUN (no format) $f — rerun with --apply to format"
	fi
	if ! annotate check-rows "$f" >/tmp/dap-rows-2.log 2>&1; then
		echo "GATE-FAIL (post-format) $f"
		cat /tmp/dap-rows-2.log
		fail=1
		continue
	fi
	echo "OK $f"
done

exit "$fail"
