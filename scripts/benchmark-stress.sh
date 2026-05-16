#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

RATE="${RATE:-10000}"
DURATION="${DURATION:-30s}"
STOP_AFTER="${STOP_AFTER:-$DURATION}"
SEED="${SEED:-1}"
BENCHMARK_NAME="${BENCHMARK_NAME:-$(date +%Y%m%d-%H%M%S)}"

OUT_DIR="${OUT_DIR:-$ROOT_DIR/results/stress-benchmark-$BENCHMARK_NAME}"
PEACOCK_BIN="$ROOT_DIR/target/peacock"
CPU_PROFILE="$OUT_DIR/peacock.cpu.prof"
PPROF_TEXT="$OUT_DIR/peacock.pprof.txt"
METRICS="$OUT_DIR/metrics.txt"
TTY_LOG="$OUT_DIR/peacock.tty.log"

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

write_metrics() {
  local status="$1"
  local status_text="$2"

  {
    printf 'started_at_epoch=%s\n' "$start_epoch"
    printf 'rate=%s\n' "$RATE"
    printf 'duration=%s\n' "$DURATION"
    printf 'stop_after=%s\n' "$STOP_AFTER"
    printf 'seed=%s\n' "$SEED"
    printf 'benchmark_name=%s\n' "$BENCHMARK_NAME"
    printf 'elapsed_ms=%s\n' "$elapsed_ms"
    printf 'benchmark_exit_status=%s\n' "$status"
    printf 'benchmark_status=%s\n' "$status_text"
    printf 'tty_log=%s\n' "$TTY_LOG"
    printf 'cpu_profile=%s\n' "$CPU_PROFILE"
    printf 'pprof_text=%s\n' "$PPROF_TEXT"
  } > "$METRICS"
}

require go
require script

mkdir -p "$OUT_DIR" "$ROOT_DIR/target"

printf 'building peacock...\n'
go build -o "$PEACOCK_BIN" "$ROOT_DIR/cmd/peacock"

printf 'running stress benchmark: name=%s rate=%s duration=%s stop_after=%s seed=%s\n' "$BENCHMARK_NAME" "$RATE" "$DURATION" "$STOP_AFTER" "$SEED"

start_epoch=$(date +%s)
start_ns=$(date +%s%N)

PIPELINE='go run "$ROOT_DIR/testdata/stress-test/main.go" \
  -rate "$RATE" \
  -duration "$DURATION" \
  -seed "$SEED" \
  | TERM=xterm-256color "$PEACOCK_BIN" --cpuprofile "$CPU_PROFILE"'

export PIPELINE ROOT_DIR RATE DURATION SEED PEACOCK_BIN CPU_PROFILE

set +e
(sleep "$STOP_AFTER"; printf '\003') \
  | script -q -e -f -c 'bash -o pipefail -c "$PIPELINE"' "$TTY_LOG" >/dev/null 2>&1
benchmark_status=$?
set -e

end_ns=$(date +%s%N)
elapsed_ns=$((end_ns - start_ns))
elapsed_ms=$((elapsed_ns / 1000000))

benchmark_status_text="exited"

write_metrics "$benchmark_status" "$benchmark_status_text"

if [[ -s "$CPU_PROFILE" ]]; then
  go tool pprof -top "$PEACOCK_BIN" "$CPU_PROFILE" > "$PPROF_TEXT"
else
  printf 'warning: CPU profile was not written or is empty: %s\n' "$CPU_PROFILE" >&2
fi

printf 'results written to %s\n' "$OUT_DIR"
printf 'metrics: %s\n' "$METRICS"
printf 'cpu profile: %s\n' "$CPU_PROFILE"
if [[ -s "$PPROF_TEXT" ]]; then
  printf 'pprof top: %s\n' "$PPROF_TEXT"
fi

exit "$benchmark_status"
