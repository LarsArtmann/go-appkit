#!/usr/bin/env bash
# Example smoke harness (SUPERB v6 T13). The seven example dirs carry
# "[no test files]" — until now their only verification was dated
# "live-verified" claims from hand-run sessions. This script makes the
# claim mechanical: build each example hermetically, run it on an
# ephemeral port, probe its health endpoint plus one distinctive route,
# SIGTERM it, and require the graceful-shutdown line in its log.
#
#   ./scripts/example-smoke.sh            # all examples
#   ./scripts/example-smoke.sh otel core  # just these
#
# Probes are path:status pairs; readiness is the FIRST probe retried for
# up to 10s. A service that fails to start, answers wrong, or does not
# shut down gracefully fails its leg.

set -uo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root" || exit 1

go_env=(env GOWORK=off "GOTOOLCHAIN=go1.27.2")
status=0

fail() {
	echo "FAIL: $*"
	status=1
}

probe() {
	"${go_env[@]}" go run ./scripts/probe "http://127.0.0.1:$1$2" "$3" 2>/dev/null
}

smoke() {
	local name="$1" dir="$2" extra_env="$3"
	shift 3

	echo "== $name ($dir) =="

	local bin="/tmp/example-smoke-$name"
	if ! (cd "$dir" && "${go_env[@]}" go build -o "$bin" ./example) >"/tmp/example-smoke-$name-build.log" 2>&1; then
		fail "$name: build failed (see /tmp/example-smoke-$name-build.log)"
		return
	fi

	local log="/tmp/example-smoke-$name.log"
	local port pid attempt

	for attempt in 1 2; do
		port=$((18100 + RANDOM % 3000))
		rm -f "$log"

		# shellcheck disable=SC2086
		env PORT="$port" $extra_env "$bin" >"$log" 2>&1 &
		pid=$!

		local first_path first_status
		first_path="${1%%:*}"
		first_status="${1#*:}"

		local deadline=$((SECONDS + 10))
		until probe "$port" "$first_path" "$first_status"; do
			if ! kill -0 "$pid" 2>/dev/null; then
				fail "$name: exited during startup (attempt $attempt; log: $log)"
				continue 2
			fi
			if ((SECONDS > deadline)); then
				kill "$pid" 2>/dev/null
				wait "$pid" 2>/dev/null
				fail "$name: $first_path never answered $first_status within 10s (attempt $attempt)"
				continue 2
			fi
			sleep 0.2
		done
		break
	done

	kill -0 "$pid" 2>/dev/null || return

	local pair path want
	for pair in "$@"; do
		path="${pair%%:*}"
		want="${pair#*:}"
		if probe "$port" "$path" "$want"; then
			echo "OK:   $name $path -> $want"
		else
			fail "$name $path did not answer $want (log: $log)"
		fi
	done

	kill -TERM "$pid" 2>/dev/null

	local wait_deadline=$((SECONDS + 10))
	while kill -0 "$pid" 2>/dev/null; do
		if ((SECONDS > wait_deadline)); then
			kill -KILL "$pid" 2>/dev/null
			fail "$name: did not exit within 10s of SIGTERM"
			return
		fi
		sleep 0.1
	done
	wait "$pid" 2>/dev/null

	if grep -q 'graceful shutdown complete' "$log"; then
		echo "OK:   $name graceful shutdown"
	else
		fail "$name: no 'graceful shutdown complete' line after SIGTERM (log: $log)"
	fi
}

targets_all=(
	"core|.|API_KEY=|/health/ready:200 /:200"
	"security|security|API_KEY=smoke-key|/health/live:200 /api/data:401"
	"health|health|API_KEY=|/health/ready:200 /hello:200"
	"otel|otel|API_KEY=|/health:200 /users/42:200"
	"flightrecorder|flightrecorder|API_KEY=|/health/ready:200 /ok:200"
	"systemd|systemd|API_KEY=|/health/ready:200 /:200"
	"errorpages|errorpages|API_KEY=|/health:200 /no/such/page:404"
)

wanted=("$@")
for target in "${targets_all[@]}"; do
	IFS='|' read -r name dir extra_env probes <<<"$target"

	if ((${#wanted[@]})); then
		skip=1
		for w in "${wanted[@]}"; do
			[[ "$w" == "$name" ]] && skip=0
		done
		((skip)) && continue
	fi

	# shellcheck disable=SC2086
	smoke "$name" "$root/$dir" "$extra_env" $probes
done

if ((status == 0)); then
	echo "EXAMPLE SMOKE GREEN"
fi
exit "$status"
