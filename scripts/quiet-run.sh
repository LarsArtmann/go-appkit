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
#   4. RESOURCE GATES — /tmp free space (--tmp-min-mb, default 2048 MB)
#      and GOCACHE/GOMODCACHE writability probes, checked once before the
#      load loop: tmp exhaustion and a root-owned/read-only cache are the
#      2026-11-02 incident class (builds die with stdlib-internals errors
#      that masquerade as dependency breakage).
#      Default provenance (2026-11-07): 2048 MB ≈ one single-module cold
#      GOMODCACHE materialization in /tmp — the go-scripts (verify-consumer,
#      check-gosum-canonical) mktemp caches there and the shared fleet
#      cache measures 9.6 GB across 14 modules, so a cold single-module
#      download lands 0.5–2 GB — plus headroom over /tmp's tmpfs total
#      (48 GB on this machine). Raise --tmp-min-mb for full-fleet cold
#      runs.
#
# Usage:
#   scripts/quiet-run.sh [--load-max N] [--wait SECONDS] [--pin-cores K]
#       [--tmp-min-mb N] -- <cmd...>
#
# Examples:
#   scripts/quiet-run.sh --load-max 16 --pin-cores 8 -- \
#       env GOWORK=off GOTOOLCHAIN=go1.27.4 go test -bench . -benchmem -count=10
#   scripts/quiet-run.sh --wait 3600 -- ./scripts/verify-all.sh
#
# Exit codes: 0 = the command ran and exited 0 (other command exit codes
# are propagated); 2 = the gate never opened (command not started).
set -euo pipefail

load_max=""
pin_cores=0
wait_secs=0
tmp_min_mb=2048
cmd=()

while [ $# -gt 0 ]; do
	case "$1" in
	--load-max)
		load_max="${2:?--load-max needs a value}"
		shift 2
		;;
	--pin-cores)
		pin_cores="${2:?--pin-cores needs a value}"
		shift 2
		;;
	--tmp-min-mb)
		tmp_min_mb="${2:?--tmp-min-mb needs a value}"
		shift 2
		;;
	--wait)
		wait_secs="${2:?--wait needs a value}"
		shift 2
		;;
	--)
		shift
		cmd=("$@")
		break
		;;
	*)
		echo "quiet-run: unknown arg: $1" >&2
		exit 2
		;;
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
	load_max=$(($(nproc) / 2))
fi

# --- Resource gates (static, once, before the load loop) ---------------------
tmp_avail_mb="$(df -Pm /tmp | awk 'NR==2 {print $4}')"
if [ "${tmp_avail_mb:-0}" -lt "$tmp_min_mb" ]; then
	log "GATE CLOSED: /tmp has ${tmp_avail_mb}MB free (< --tmp-min-mb=$tmp_min_mb) — go builds fail on exhausted tmp: not starting: ${cmd[*]}"
	exit 2
fi
log "RESOURCE GATE OK: /tmp ${tmp_avail_mb}MB free (min ${tmp_min_mb}MB)"

cache_dirs="$(go env GOCACHE GOMODCACHE 2>/dev/null || echo "$HOME/.cache/go-build $HOME/go/pkg/mod")"
for d in $cache_dirs; do
	probe="$(mktemp "$d/.quiet-run-probe.XXXXXX" 2>/dev/null)" || probe=""
	if [ -z "$probe" ]; then
		mkdir -p "$d" 2>/dev/null
		probe="$(mktemp "$d/.quiet-run-probe.XXXXXX" 2>/dev/null)" || probe=""
	fi
	if [ -z "$probe" ]; then
		log "GATE CLOSED: cache NOT writable: $d (root-owned/read-only cache is the 2026-11-02 a15 incident class) — fix ownership or point GOCACHE/GOMODCACHE elsewhere: not starting: ${cmd[*]}"
		exit 2
	fi
	# Exact-name cleanup: a window sweep (find -newermt) can delete a
	# concurrent invocation's fresh probe file.
	trash "$probe" 2>/dev/null || rm -f "$probe"
	log "RESOURCE GATE OK: cache writable: $d"
done

# --- Gate loop ---------------------------------------------------------------
deadline=$(($(date +%s) + wait_secs))
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
	# SMT-aware: prefer PRIMARY threads only (first-listed sibling == the cpu
	# itself). Pinning an SMT sibling pair halves throughput for both and makes
	# benchmark timings bimodal — found live 2026-10-18: a busy-based pick of
	# {12,28,9,25,...} included the sibling pairs 12/28 and 9/25 and blew the
	# middleware-bench sd to ±26-36% while sibling-free paths sat at ±2%.
	primary_cores() {
		awk '/^cpu[0-9]+/ { sub(/^cpu/, "", $1); print $1 }' /proc/stat | while read -r c; do
			f="/sys/devices/system/cpu/cpu$c/topology/thread_siblings_list"
			if [ -r "$f" ]; then
				first="$(cut -d, -f1 "$f" | cut -d- -f1)"
				[ "$first" = "$c" ] && echo "$c"
			else
				echo "$c"
			fi
		done
	}
	s1="$(sample_stat)"
	sleep 0.3
	s2="$(sample_stat)"
	cores="$(printf '%s\n%s\n' "$s1" "$s2" | awk '
    { key=$1; busy[key]+=$2; idle[key]+=$3 }
    END { for (c in busy) printf "%s %.6f\n", c, busy[c]/(busy[c]+idle[c]+1) }' |
		sort -k2,2n | awk '{gsub(/^cpu/, "", $1); print $1}' |
		grep -xF -f <(primary_cores) |
		head -n "$pin_cores" | paste -sd, -)"
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
