#!/usr/bin/env bash
# check-response-drain.sh — the undrained-body rule (issue54 session §e5):
# three occurrences of a response body closed without being read (preset
# probes, rate-limit SSE E2E, errorpages E2E) held server-side connections
# ACTIVE through server.Shutdown and stalled it past the stop timeout. The
# root cause class, now mechanical: a *_test.go function that calls
# `X.Body.Close()` without ever READING X.Body is a failure.
#
# Mechanism: a repo-local AST walker (scripts/check_response_drain.go,
# go/parser — syntax facts only, per the linter-building decision tree).
# No assignment-source trust exists: verified 2026-11-06 that every response
# helper in the fleet (root httpGet, flightrecorder httpGetURL, preset
# getWithHeaders, testkit.HTTPGetContext) returns the RAW response — the
# caller owns the body. Suppression: `// respdrain:ignore <reason>` on the
# close line, reason required (SSE streams are the legitimate class: Close
# IS the disconnect).
#
# FAST+OFFLINE → runs in all three venues (pre-commit, pre-tag-checks, CI).
#
# Exit codes: 0 clean · 1 findings · 2 walk failure.

set -uo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

floor="$(sed -n 's/^go \([0-9][0-9.]*\)$/\1/p' go.work | head -n1)"
if [[ -z "$floor" ]]; then
	echo "FAIL: could not read the go floor from go.work"
	exit 2
fi

out="$(GOWORK=off GOTOOLCHAIN="go$floor" go run scripts/check_response_drain.go . 2>&1)"
status=$?
if [[ "$status" -eq 0 ]]; then
	echo "RESPONSE DRAIN: every closed test body is read or explicitly ignored (SSE carve-out via respdrain:ignore)"
	exit 0
elif [[ "$status" -eq 1 ]]; then
	echo "$out"
	echo "FAIL: undrained response body/clos site(s) above — drain with io.Copy(io.Discard, X.Body) before Close, or suppress with // respdrain:ignore <reason> (SSE streams only)"
	exit 1
else
	echo "$out"
	echo "FAIL: the drain checker failed to walk the tree"
	exit 2
fi
