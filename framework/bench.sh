#!/bin/bash
# Shootout: wrk all four twins, same endpoint shape.
# Usage: start the four servers first (mux :8112, gin :8120, echo :8130, fiber :8140),
# then: bash bench.sh
set -u

declare -A TARGETS=(
  [mux]="http://localhost:8112/todos"
  [gin]="http://localhost:8120/todos"
  [echo]="http://localhost:8130/todos"
  [fiber]="http://localhost:8140/todos"
)

for name in mux gin echo fiber; do
  echo "=== $name ==="
  wrk -t4 -c100 -d15s "${TARGETS[$name]}" 2>&1 | grep -E 'Requests/sec|Latency|Transfer/sec'
done
