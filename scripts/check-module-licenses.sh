#!/usr/bin/env bash
# LICENSE-coverage guard (2026-11-06, issue54 session §e6): AGENTS asserted
# "every module root carries a LICENSE" while health shipped for its whole
# life without one — the claim was prose, nothing enforced it, and pkg.go.dev
# read the module as unlicensed. This script makes the claim mechanical:
# every released module (root + every go.work member except integration)
# must carry a LICENSE file at its module root.
#
# integration/ is EXEMPT by charter: it is never released (integration/doc.go
# states the deliberate absence), so pkg.go.dev never renders it.
#
# FAST+OFFLINE by design — runs in all three guard venues (pre-commit fatal
# set, pre-tag-checks, CI) per the M14 venue matrix.
#
# Exit codes: 0 green · 1 drift found.

set -uo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

status=0
fail() {
	echo "FAIL: $*"
	status=1
}
ok() { echo "OK:   $*"; }

members=(".")
while IFS= read -r use_entry; do
	case "$use_entry" in
	".") continue ;;
	./integration) continue ;;
	.*) members+=("${use_entry#./}") ;;
	*) continue ;;
	esac
done < <(awk '/^use \(/{f=1; next} /^\)/{f=0} f && NF {print $1}' go.work)

for dir in "${members[@]}"; do
	if [[ -f "$dir/LICENSE" ]]; then
		ok "$dir carries LICENSE"
	else
		fail "$dir has no LICENSE at its module root (pkg.go.dev renders it unlicensed; released modules must carry it)"
	fi
done

# integration is deliberately exempt — assert the exemption stays documented,
# so a future release of the module cannot silently inherit the gap.
if [[ -f integration/LICENSE ]]; then
	ok "integration carries LICENSE (charter exemption moot — keep the doc.go note in sync)"
elif grep -qi "LICENSE" integration/doc.go 2>/dev/null; then
	ok "integration exempt (never released; absence stated in integration/doc.go)"
else
	fail "integration has no LICENSE and integration/doc.go does not state the deliberate absence (never-released charter)"
fi

if [[ "$status" -eq 0 ]]; then
	echo "MODULE LICENSES: ${#members[@]} released module(s) green (+ integration exemption documented)"
fi
exit "$status"
