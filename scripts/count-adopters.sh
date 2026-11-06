#!/usr/bin/env bash
# Adoption counter (SUPERB v12 micro-backlog / v13 G-CNT): makes the
# six-month success gate a command, not a claim. Re-derives the
# integration-burden audit's population (meta app-tagged projects that are
# Go apps) and counts, per family module, how many apps require it —
# ANY (including indirect) and DIRECT — in dual columns.
#
#   ./scripts/count-adopters.sh              # table + summary + delta vs baseline
#   ./scripts/count-adopters.sh --baseline   # (re)write the baseline file
#
# LOCAL-ONLY instrument (same class as quiet-run.sh): it reads the
# ~/projects fleet and the meta inventory — CI and pre-tag venues can see
# neither, so the wiring IS the committed baseline plus this re-run at the
# gate. Baseline: docs/review/apps/adoption-baseline.txt
# Audit reference: docs/review/apps/2026-11-06_integration-burden-audit.md
set -euo pipefail
cd "$(dirname "$0")/.."

baseline_file="docs/review/apps/adoption-baseline.txt"
projects_root="${PROJECTS_ROOT:-/home/lars/projects}"

# The composed-blocks cluster from the audit's framing: go-appkit must beat
# the blocks it exists to compose.
modules=(
	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/httputil"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-cqrs-lite"
	"github.com/larsartmann/templ-components"
)

if ! command -v meta >/dev/null 2>&1; then
	echo "FAIL: meta inventory tool not on PATH (local-only instrument)" >&2
	exit 1
fi

# Population: app-tagged projects with at least one go.mod (vendor and
# toolchain dirs excluded) — the audit's Appendix A method.

# resolve_dir: meta names and directory names drift in CASE (papdashboard →
# PapDashboard) and a case-sensitive `[ -d ]` silently drops the project —
# that undercount shipped a wrong baseline once (2026-11-06: population 31
# vs the true 41, two adopters invisible). Exact match first, then a single
# case-insensitive match with a loud warning; zero or >1 matches skips.
resolve_dir() {
	local name="$1" hit
	if [ -d "$projects_root/$name" ]; then
		printf '%s' "$projects_root/$name"
		return 0
	fi
	hit=$(find "$projects_root" -maxdepth 1 -mindepth 1 -type d -iname "$name" 2>/dev/null | head -n 1)
	if [ -n "$hit" ]; then
		echo "WARN: meta name '$name' resolved case-insensitively to ${hit##*/}" >&2
		printf '%s' "$hit"
		return 0
	fi
	return 1
}

declare -a go_apps=()
while IFS= read -r name; do
	[ -z "$name" ] && continue
	dir=$(resolve_dir "$name") || continue
	if find "$dir" -name go.mod -not -path '*/vendor/*' -not -path '*/node_modules/*' \
		-not -path '*/.direnv/*' -not -path '*/result*' -print -quit 2>/dev/null | grep -q .; then
		go_apps+=("$name")
	fi
done < <(meta list --tags app --format json 2>/dev/null | jq -r 'keys[]' 2>/dev/null | sort)

total=${#go_apps[@]}
if ((total == 0)); then
	echo "FAIL: no Go apps found via meta --tags app" >&2
	exit 1
fi

# Counters per module: any_ (any require incl. indirect), direct_ (no
# // indirect on the require line).
summary_lines=()
for module in "${modules[@]}"; do
	any_count=0
	direct_count=0
	for name in "${go_apps[@]}"; do
		dir=$(resolve_dir "$name") || continue
		any_hit=0
		direct_hit=0
		while IFS= read -r gomod; do
			# Per-LINE indirect check: a module can appear BOTH ways in one
			# go.mod (papdashboard: direct core + indirect security sibling) —
			# the old per-FILE check let any `// indirect` sibling line poison
			# the direct verdict (2026-11-06, undercounted direct 7→6).
			while IFS= read -r line; do
				if printf '%s\n' "$line" | grep -Eq "^[[:space:]]*${module}([[:space:]]|/)"; then
					any_hit=1
					if ! printf '%s\n' "$line" | grep -Eq "//[[:space:]]*indirect([[:space:]]|$)"; then
						direct_hit=1
					fi
				fi
			done < "$gomod"
		done < <(find "$dir" -name go.mod -not -path '*/vendor/*' -not -path '*/node_modules/*' \
			-not -path '*/.direnv/*' -not -path '*/result*' 2>/dev/null)
		any_count=$((any_count + any_hit))
		direct_count=$((direct_count + direct_hit))
	done
	summary_lines+=("$(printf '%-45s any %2d/%d  direct %2d/%d' "$module" "$any_count" "$total" "$direct_count" "$total")")
done

echo "Go apps (meta --tags app): $total"
printf '%s\n' "${summary_lines[@]}"

if [[ "${1:-}" == "--baseline" ]]; then
	{
		echo "# Adoption baseline — written $(date '+%Y-%m-%d %H:%M %z') by scripts/count-adopters.sh"
		echo "# Six-month gate (SUPERB v12/v13): go-appkit direct > 10 (beats every composed block)"
		echo "Go apps (meta --tags app): $total"
		printf '%s\n' "${summary_lines[@]}"
	} >"$baseline_file"
	echo "BASELINE WRITTEN: $baseline_file"
	exit 0
fi

if [ -f "$baseline_file" ]; then
	echo
	echo "--- delta vs baseline ($baseline_file) ---"
	if diff <(grep -v '^#' "$baseline_file") <(printf 'Go apps (meta --tags app): %d\n%s\n' "$total" "$(printf '%s\n' "${summary_lines[@]}")") &&
		echo "UNCHANGED since baseline"; then
		:
	fi
else
	echo
	echo "NOTE: no baseline yet — run with --baseline to write one"
fi
