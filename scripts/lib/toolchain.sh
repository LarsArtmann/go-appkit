# shellcheck shell=bash
# Shared GOTOOLCHAIN resolution for the repo scripts.
#
# GOTOOLCHAIN accepts toolchain names, not language versions: `go1.27` is
# rejected with "go1.27 is a language version but not a toolchain version
# (go1.27.x)" (proven 2026-11-22, the same day the go directives went
# major.minor). So a patch-level directive is used verbatim, while a
# major.minor directive (the repo floor, `go 1.27`) falls back to the patched
# toolchain the repo tests with.
#
# Keep APPKIT_TOOLCHAIN_PATCH in step with the GOTOOLCHAIN belts in the
# READMEs / AGENTS.md / CI whenever a patched release is adopted.
#
# Usage:
#   source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/toolchain.sh"
#   export GOTOOLCHAIN="$(go_toolchain_for "$floor")"

APPKIT_TOOLCHAIN_PATCH="${APPKIT_TOOLCHAIN_PATCH:-1.27.4}"

go_toolchain_for() {
	local directive="$1"

	if [[ "$directive" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		echo "go$directive"

		return
	fi

	echo "go$APPKIT_TOOLCHAIN_PATCH"
}
