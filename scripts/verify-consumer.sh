#!/usr/bin/env bash
# verify-consumer.sh — the fresh-consumer behavioral check as one command
# (2026-11-06, issue54 session §f37: the ad-hoc scratch-module program —
# "verify preset still builds against health v0.3.1", the health consumer
# probe — as a repo script so the claim is re-runnable).
#
# Stronger than docs/recipes/_proxycheck (build-only): a REQUIRED usage
# snippet is compiled against the exact published module@version in a
# throwaway module outside the repo (no workspace, no replace, no shared
# cache) — the consumer's API surface is proven, not just the import.
#
# Usage:
#   scripts/verify-consumer.sh <module>@<version> <snippet.go>
#   scripts/verify-consumer.sh <module>@<version>   # blank-import build only
#
# Snippet rules: a valid Go main-package file body WITHOUT the package or
# import clauses — the wrapper injects package main, an import block with
# the target module aliased as `consumer`, and compiles. Anything the
# snippet references beyond the target module must be stdlib.
#
# Exit codes: 0 the snippet compiles against the published tag · 1 it does
# not (proxy, resolution, or API drift) · 2 usage.

set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
	echo "usage: $0 <module>@<version> [<snippet.go>]"
	exit 2
fi

target="$1"
snippet="${2:-}"
module="${target%@*}"
version="${target#*@}"

if [[ -z "$module" || -z "$version" || "$module" == "$target" ]]; then
	echo "usage: $0 <module>@<version> [<snippet.go>]   (target must be module@version)"
	exit 2
fi
if [[ -n "$snippet" && ! -f "$snippet" ]]; then
	echo "snippet not found: $snippet"
	exit 2
fi

floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' "$(dirname "$0")/../go.work" | head -n1)"

scratch="$(mktemp -d /tmp/verify-consumer-XXXXXX)"
cleanup() {
	chmod -R u+w "$scratch" 2>/dev/null || true
	rm -rf "$scratch"
}
trap cleanup EXIT

printf 'module verifyconsumer\n\ngo %s\n\nrequire %s %s\n' "$floor" "$module" "$version" >"$scratch/go.mod"

if [[ -n "$snippet" ]]; then
	{
		echo "package main"
		echo
		echo "import consumer \"$module\""
		echo
		cat "$snippet"
	} >"$scratch/main.go"
else
	{
		echo "package main"
		echo
		echo "import _ \"$module\""
		echo
		echo "func main() {}"
	} >"$scratch/main.go"
fi

echo "== fresh consumer: $module@$version (snippet: ${snippet:-blank import}) =="
# tidy and build MUST share one env: one GOPROXY, one FRESH GOMODCACHE.
# A split (build inheriting machine defaults) resolves the tag twice
# through different paths — and when a tag ever moves between resolution
# and build, or the shared cache holds a stale zip, you get a phantom
# SECURITY ERROR: checksum mismatch (the 2026-11-07 four-tag wave hit).
# GOPRIVATE stays machine-default on purpose: the fleet resolves via VCS.
consumer_cache="$(mktemp -d /tmp/verify-consumer-cache-XXXXXX)"
export GOWORK=off
export GOTOOLCHAIN="go$floor"
export GOPROXY=https://proxy.golang.org
export GOMODCACHE="$consumer_cache"
(cd "$scratch" && go mod tidy >/dev/null && go build ./...)
echo "OK: $module@$version serves the consumer surface as published"
