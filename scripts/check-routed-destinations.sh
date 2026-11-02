#!/usr/bin/env bash
# Routed-destination guard (SUPERB v10 M3, 2026-11-02).
#
# Docs-health evidence: the repo's dominant failure mode is doc-lag, and one
# of its recurring classes is DEAD ROUTED POINTERS — a status-report line says
# "→ routed — TODO row X" / "→ routed — docs/f.md" / "→ routed to the owner
# memo recommendation N" and the destination has since been closed, moved, or
# deleted, leaving a pointer to nothing. The 2026-11-02 pass hand-fixed eight
# such instances; nothing grepped for the class. This script is that grep,
# run permanently (pre-tag-checks guard #9 since M4).
#
# What counts as a routed marker (line-level, in the scanned files; inline
# `code spans` are stripped first so prose that QUOTES the marker grammar is
# not itself a marker):
#   → routed …            → routed to …         → TODO row …
#   → TODO_LIST …         → ROADMAP row …       → AGENTS row …
#
# Destination resolution, first match wins per line:
#   0. Line carries ~~ (annotate.go strikethrough) → RESOLVED (closed item,
#      historical text; same semantics as archived/) → skip.
#   1. VENUE-PAIR ("→ TODO/AGENTS", "→ TODO_LIST/ROADMAP", … a venue keyword
#      immediately followed by / or -): recommendation prose naming WHERE work
#      may go, not a pointer to a specific artifact → skip.
#   2. EXTERNAL (their repo | upstream | Open question | CLOSED COMPLETED |
#      filed as): destination lives outside this repo → skip.
#   3. FILE PATH (docs/…, scripts/…, integration/…, dashboards/… ending in
#      .md/.sh/.go/.json/.yaml/.yml): must exist from the repo root.
#   4. RECOMMENDATION N ("owner memo recommendation N"): the owner memo
#      docs/planning/2026-10-13_pma-owner-decision-memo.md must exist and
#      carry N (literal "recommendation N" or table row "| N |").
#   5. VENUE + FRAGMENT ("→ routed — TODO_LIST \"…\"", "→ TODO row — desc"):
#      target = TODO_LIST.md (TODO row, TODO_LIST, watchlist),
#      ROADMAP.md (ROADMAP), AGENTS.md (AGENTS). Fragment = quoted "…", else
#      text after — / – / : up to ; or ( — compressed, markdown stripped —
#      first 5 words must appear (case-insensitive fixed string) in the
#      target. Dead fragment = FAIL (the class this guard kills).
#   6. Trigger fired but no resolvable destination form → UNENFORCED (counted,
#      printed as a note, not a failure — prose arrows are normalized at each
#      report's docs-health sweep, not here).
#
# Scan set: TODO_LIST.md ROADMAP.md AGENTS.md FEATURES.md + docs/status/*.md
# (archived/ excluded — closed history). Extra files may be passed as args.
#
#   ./scripts/check-routed-destinations.sh            # default scan set
#   ./scripts/check-routed-destinations.sh file.md…   # scan specific files
#
# Exit codes: 0 all enforced destinations resolve · 1 dead destination or
# missing file-path target · 2 usage error.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
	sed -n '3,49p' "$0" | sed 's/^# \{0,1\}//'
	exit 0
fi

MEMO="docs/planning/2026-10-13_pma-owner-decision-memo.md"

files=()
for f in TODO_LIST.md ROADMAP.md AGENTS.md FEATURES.md; do
	[[ -f "$f" ]] && files+=("$f")
done
while IFS= read -r f; do files+=("$f"); done < <(find docs/status -maxdepth 1 -name '*.md' -type f | sort)
for f in "$@"; do
	if [[ ! -f "$f" ]]; then
		echo "USAGE: file not found: $f" >&2
		exit 2
	fi
	files+=("$f")
done

status=0
enforced=0
skipped=0
unenforced=0

fail() {
	echo "FAIL: $*"
	status=1
}
note() {
	echo "NOTE: $*"
}

compress() {
	sed -e 's/[*`_]/ /g' -e 's/[[:space:]]\+/ /g' -e 's/^ //' -e 's/ $//'
}

# window <text> — first 5 words, for tolerant fragment matching.
window() {
	echo "$1" | tr -s ' ' | cut -d' ' -f1-5
}

for f in "${files[@]}"; do
	while IFS= read -r hit; do
		lineno="${hit%%:*}"
		line="${hit#*:}"

		# Strip inline-code spans: quoting the grammar is not using it.
		bare="$(sed -E 's/`[^`]*`/ /g' <<<"$line")"
		if ! grep -qE '→ *(routed|TODO row|TODO_LIST|ROADMAP row|AGENTS row|TODO P[123])' <<<"$bare"; then
			continue
		fi

		if grep -q '~~' <<<"$bare"; then
			skipped=$((skipped + 1))
			continue
		fi

		# Destination part only (from the arrow onward; in a table row the
		# marker's cell ends at the next | — later cells are other columns'
		# content and must not classify the marker). The SUBJECT prose may
		# legitimately contain external words ("watch the upstream repo")
		# while the DESTINATION is a local TODO row — classify the pointer,
		# not the topic.
		dest="→ ${bare#*→}"
		dest="${dest%%|*}"

		# Venue-pair prose (TODO/AGENTS, TODO_LIST/ROADMAP, routed-to-pairs).
		if grep -qE '→ *(routed to )?(TODO_LIST|TODO|ROADMAP|AGENTS)[/-]' <<<"$dest"; then
			skipped=$((skipped + 1))
			continue
		fi

		# External destinations (in the destination part).
		if grep -qE '(their repo|upstream|Open question|CLOSED COMPLETED|filed as)' <<<"$dest"; then
			skipped=$((skipped + 1))
			continue
		fi

		# File-path destination: every repo-relative doc path on the line must exist.
		path_bad=0
		while IFS= read -r p; do
			[[ -z "$p" ]] && continue
			if [[ ! -f "$p" ]]; then
				path_bad=1
				fail "$f:$lineno: routed destination file does not exist: $p"
			fi
		done < <(grep -oE '(docs|scripts|integration|dashboards)/[A-Za-z0-9_./-]+\.(md|sh|go|json|yaml|yml)' <<<"$dest" | sort -u || true)
		if [[ "$path_bad" -eq 1 ]]; then
			continue
		fi

		# Owner-memo recommendation destination.
		rec="$(grep -oE 'recommendation [0-9]+' <<<"$dest" | head -n1 || true)"
		rec_num="$(grep -oE '[0-9]+' <<<"$rec" || true)"
		if [[ -n "${rec_num:-}" ]]; then
			enforced=$((enforced + 1))
			if [[ ! -f "$MEMO" ]]; then
				fail "$f:$lineno: owner memo missing: $MEMO (marker: …$rec…)"
				continue
			fi
			if ! grep -qi "$rec" "$MEMO" && ! grep -qE "^\| $rec_num \|" "$MEMO"; then
				fail "$f:$lineno: owner memo has no $rec (neither literal text nor table row) — $MEMO"
			fi
			continue
		fi

		# Venue + fragment destination.
		target=""
		if grep -qE 'TODO row|TODO_LIST|watchlist|TODO P[123]' <<<"$dest"; then
			target="TODO_LIST.md"
		elif grep -q 'ROADMAP' <<<"$dest"; then
			target="ROADMAP.md"
		elif grep -q 'AGENTS' <<<"$dest"; then
			target="AGENTS.md"
		fi
		if [[ -n "$target" ]]; then
			enforced=$((enforced + 1))
			if grep -qE 'TODO P[123]' <<<"$dest"; then
				sec="$(grep -oE 'TODO P[123]' <<<"$dest" | head -n1 | grep -oE 'P[123]')"
				if ! grep -qE "^## $sec" "$target"; then
					fail "$f:$lineno: dead routed destination — TODO_LIST has no $sec section (marker: $(echo "$line" | cut -c1-160))"
				fi
				continue
			fi
			frag=""
			if grep -qE '"[^"]+"' <<<"$dest"; then
				frag="$(grep -oE '"[^"]+"' <<<"$dest" | head -n1 | tr -d '"')"
			else
				frag="$(sed -E 's/.*→[^—–:]*[—–: ]+//' <<<"$dest" | cut -d';' -f1 | cut -d'(' -f1)"
			fi
			frag="$(compress <<<"$frag")"
			if [[ -z "$frag" ]]; then
				note "$f:$lineno: venue destination without extractable fragment (unenforced)"
				unenforced=$((unenforced + 1))
				continue
			fi
			w="$(window "$frag")"
			if ! grep -qiF "$w" "$target"; then
				fail "$f:$lineno: dead routed destination — \"$w\" not found in $target (marker: $(echo "$line" | cut -c1-160))"
			fi
			continue
		fi

		# Trigger fired, nothing resolvable.
		note "$f:$lineno: routed marker without a resolvable destination form (unenforced)"
		unenforced=$((unenforced + 1))
	done < <(grep -nE '→ *(routed|TODO row|TODO_LIST|ROADMAP row|AGENTS row|TODO P[123])' "$f" || true)
done

echo "OK:   routed-destinations: $enforced enforced, $skipped skipped (resolved/external/venue-pair), $unenforced unenforced, $((${#files[@]})) files scanned"
if [[ "$status" -ne 0 ]]; then
	echo "FAIL: routed-destinations: dead routed pointer(s) above — fix the marker or resolve the item (annotate.go)"
fi
exit "$status"
