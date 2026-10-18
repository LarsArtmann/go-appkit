#!/usr/bin/env bash
# quiet-run.sh — mechanical quiet-window gate for load-sensitive work.
#
# This box is shared and regularly runs at load 200-400 on 32 cores, which
# starves deadline-bearing tests (AGENTS "Load-flake protocol") and poisons
# wall-clock benchmarks (the ±25% honesty bar, otel README). This helper
# makes "quiet" mechanical instead of a judgment call (born 2026-10-13
# 17-49 §e2; TODO_LIST P2 "Quiet-gate helper"):
#
#   1. LOAD GATE — refuse (exit 2) unless the 1-minute load average is
#      <= --load-max (default: half the online cores).
#   2. CPU PIN   — optionally pin the command to the K least-busy cores
#      via taskset (--pin-cores K, default 0 = off), sampled from
#      /proc/stat over a short window.
#   3. EVIDENCE  — load before/after and chosen cores go to stderr and to
#      $QUIET_RUN_LOG (one line each, if set), so a deferral or a poisoned
#      run can be proven after the fact.
#
# Usage:
#   scripts/quiet-run.sh [--load-max N] [--wait SECONDS] [--pin-cores K] -- <cmd...>
#
# Examples:
#   scripts/quiet-run.sh --load-max 16 --pin-cores 8 -- \
#       env GOWORK=off GOTOOLCHAIN=go1.27.2 go test -bench . -benchmem -count=10
#   scripts/quiet-run.sh --wait 3600 -- ./scripts/verify-all.sh
#
# Exit codes: 0 = the command ran and exited 0 (other command exit codes
# are propagated); 2 = the gate never opened (command not started).
set -euo pipefail

load_max=""
pin_cores=0
wait_secs=0
cmd=()

while [ $# -gt 0 ]; do
  case "$1" in
    --load-max) load_max="${2:?--load-max needs a value}"; shift 2 ;;
    --pin-cores) pin_cores="${2:?--pin-cores needs a value}"; shift 2 ;;
    --wait) wait_secs="${2:?--wait needs a value}"; shift 2 ;;
    --) shift; cmd=("$@"); break ;;
    *) echo "quiet-run: unknown arg: $1" >&2; exit 2 ;;
  esac
done

if [ ${#cmd[@]} -eq 0 ]; then
  sed -n '2,30p' "$0" >&2
  echo "quiet-run: no command given (put it after --)" >&2
  exit 2
fi

log() {
  echo "quiet-run: $*" >&2
  if [ -n "${QUIET_RUN_LOG:-}" ]; then
    printf '%s quiet-run: %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$*" >>"$QUIET_RUN_LOG"
  fi
}

# Float-safe comparison: load1 <= ref?
load1() { awk '{print $1}' /proc/loadavg; }
gate_open() { awk -v a="$1" -v b="$2" 'BEGIN { exit !(a <= b) }'; }

if [ -z "$load_max" ]; then
  load_max=$(( $(nproc) / 2 ))
fi

# --- Gate loop ---------------------------------------------------------------
deadline=$(( $(date +%s) + wait_secs ))
while :; do
  l1="$(load1)"
  if gate_open "$l1" "$load_max"; then
    break
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    log "GATE CLOSED: load1=$l1 > load-max=$load_max (waited ${wait_secs}s) — not starting: ${cmd[*]}"
    exit 2
  fi
  sleep 15
done

# --- Core picker: K least-busy cores over a 0.3s /proc/stat sample -----------
pin_prefix=()
if [ "$pin_cores" -gt 0 ]; then
  sample_stat() { awk '/^cpu[0-9]+/ {idle=$5; busy=$2+$3+$4+$6+$7+$8; print $1, busy, idle}' /proc/stat; }
  s1="$(sample_stat)"
  sleep 0.3
  s2="$(sample_stat)"
  cores="$(printf '%s\n%s\n' "$s1" "$s2" | awk '
    { key=$1; busy[key]+=$2; idle[key]+=$3 }
    END { for (c in busy) printf "%s %.6f\n", c, busy[c]/(busy[c]+idle[c]+1) }' \
    | sort -k2,2n | head -n "$pin_cores" | awk '{gsub(/^cpu/, "", $1); printf "%s,",$1 }' | sed 's/,$//')"
  if [ -z "$cores" ]; then
    log "GATE ERROR: core sampling failed — not starting"
    exit 2
  fi
  pin_prefix=(taskset -c "$cores")
  log "GATE OPEN: load1=$(load1) load-max=$load_max pin=$cores cmd=${cmd[*]}"
else
  log "GATE OPEN: load1=$(load1) load-max=$load_max (no pinning) cmd=${cmd[*]}"
fi

# --- Run ---------------------------------------------------------------------
set +e
"${pin_prefix[@]}" "${cmd[@]}"
rc=$?
set -e
log "DONE: exit=$rc load1-after=$(load1)"
exit "$rc"
