#!/usr/bin/env bash
# scripts/verify-all.sh — the full hermetic verify loop: per-module build +
# vet + test -race, then sequential per-module golangci-lint.
#
# Encodes the documented AGENTS.md loop as one command. Every module verifies
# with GOWORK=off (consumer-style resolution, how published tags are built)
# and the toolchain derived from the root go.mod directive (the machine
# default is GOTOOLCHAIN=local, which cannot switch). Lint runs SEQUENTIALLY,
# one module at a time: concurrent golangci-lint processes race on the shared
# build cache (/mnt/buildcache) and under-report findings (verified
# 2026-08-18). Output per module lands in a log file under LOG_DIR and the
# verdict is printed to stdout — never swallow this script's output through
# head/tail (the exit-code-swallowing incident class, AGENTS.md Gotchas).
#
# Usage: scripts/verify-all.sh [--skip-tests] [--skip-lint] [module ...]
#        modules are given as directory names (use "." for the root module)
set -euo pipefail

cd "$(git rev-parse --show-toplevel 2>/dev/null || echo .)"

SKIP_TESTS=false
SKIP_LINT=false
MODULES=()
for arg in "$@"; do
  case "$arg" in
    --skip-tests) SKIP_TESTS=true ;;
    --skip-lint) SKIP_LINT=true ;;
    *) MODULES+=("$arg") ;;
  esac
done

if [[ ${#MODULES[@]} -eq 0 ]]; then
  MODULES=(".")
  for d in */go.mod; do MODULES+=("${d%/go.mod}"); done
fi

GO_DIRECTIVE=$(awk '/^go /{print $2; exit}' go.mod)
if [[ -z "$GO_DIRECTIVE" ]]; then
  echo "FATAL: could not read the go directive from go.mod" >&2
  exit 2
fi
export GOTOOLCHAIN="go$GO_DIRECTIVE"
export GOWORK=off
export CGO_ENABLED=1

LOG_DIR=$(mktemp -d /tmp/go-appkit-verify-XXXXXX)
FAILED=()
STARTED=$SECONDS

echo "verify-all: modules=${MODULES[*]}"
echo "verify-all: GOTOOLCHAIN=$GOTOOLCHAIN logs=$LOG_DIR"
verdict() {
  local name="$1" code="$2" log="$3"
  if [[ $code -eq 0 ]]; then
    echo "PASS  $name  EXIT=0  ($(wc -l <"$log") lines)"
  else
    echo "FAIL  $name  EXIT=$code  log=$log"
    tail -30 "$log"
    FAILED+=("$name")
  fi
}

if [[ $SKIP_TESTS == false ]]; then
  for mod in "${MODULES[@]}"; do
    log="$LOG_DIR/$(echo "$mod" | tr '/' '_')_test.log"
    echo "--- test $mod ---"
    if (cd "$mod" && go build ./... && go vet ./... && go test ./... -race -count=1) >"$log" 2>&1; then
      verdict "$mod test" 0 "$log"
    else
      verdict "$mod test" $? "$log"
    fi
  done
fi

if [[ $SKIP_LINT == false ]]; then
  if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "FAIL  golangci-lint not on PATH - install it or pass --skip-lint" >&2
    exit 2
  fi
  for mod in "${MODULES[@]}"; do
    log="$LOG_DIR/$(echo "$mod" | tr '/' '_')_lint.log"
    echo "--- lint $mod ---"
    if [[ "$mod" == "integration" ]]; then
      (cd "$mod" && GOWORK=off golangci-lint run ./...) >"$log" 2>&1
    else
      (cd "$mod" && golangci-lint run ./...) >"$log" 2>&1
    fi
    verdict "$mod lint" $? "$log"
  done
fi

ELAPSED=$((SECONDS - STARTED))
echo "verify-all: ${ELAPSED}s elapsed, logs in $LOG_DIR"
if [[ ${#FAILED[@]} -gt 0 ]]; then
  echo "verify-all: FAILED modules: ${FAILED[*]}"
  exit 1
fi
echo "verify-all: ALL GREEN"
