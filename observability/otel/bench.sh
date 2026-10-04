#!/usr/bin/env bash
# bench.sh — fire N authenticated requests at svc-a, then verify the
# metrics moved: counter delta, histogram buckets, and one exemplar
# carrying a trace_id (paste it into Jaeger to see the waterfall).
#
# Usage: ./bench.sh [N] [CONCURRENCY]   (defaults: 200 20)
# Env:   BASE (default http://localhost:8110), METRICS (default http://localhost:2112/metrics)
set -euo pipefail

BASE="${BASE:-http://localhost:8110}"
METRICS_URL="${METRICS:-http://localhost:2112/metrics}"
N="${1:-200}"
C="${2:-20}"

counter() {
	curl -s -H "Accept: application/openmetrics-text" "$METRICS_URL" |
		awk '/^http_requests_total\{route="\/start"/ {print $2}'
}

echo "==> mint token"
TOKEN=$(curl -s -X POST "$BASE/token" -H 'Content-Type: application/json' \
	-d '{"tenant_id":"bench","user_id":"bench"}' |
	grep -o '"token":"[^"]*"' | cut -d'"' -f4)
[ -n "$TOKEN" ] || { echo "mint failed"; exit 1; }

echo "==> counter before: $(counter)"
BEFORE=$(counter); BEFORE=${BEFORE%.*}

echo "==> firing $N requests (concurrency $C)"
CODES=$(seq "$N" | xargs -P "$C" -I{} curl -s -o /dev/null -w "%{http_code}\n" \
	"$BASE/start" -H "Authorization: Bearer $TOKEN")
OK=$(echo "$CODES" | grep -c '^200$' || true)
echo "==> 200s: $OK/$N"

AFTER=$(counter); AFTER=${AFTER%.*}
echo "==> counter after: $AFTER (delta: $((AFTER - BEFORE)))"
[ "$((AFTER - BEFORE))" -eq "$N" ] || { echo "WARN: delta != N, some requests missed metrics"; }

echo "==> histogram buckets (/start 200):"
curl -s -H "Accept: application/openmetrics-text" "$METRICS_URL" |
	grep '^http_request_duration_seconds_bucket{route="/start",status="200"' |
	awk '{print $1, $2}'

echo "==> exemplar sample (trace_id -> Jaeger):"
curl -s -H "Accept: application/openmetrics-text" "$METRICS_URL" |
	grep -m1 -o '# {trace_id="[^"]*"}' || echo "no exemplar found"

echo "==> in-flight now (want 0 — nonzero means a forgotten Dec):"
curl -s -H "Accept: application/openmetrics-text" "$METRICS_URL" |
	grep '^http_inflight_requests' || echo "no inflight series"

echo "==> response size quantiles (/start):"
curl -s -H "Accept: application/openmetrics-text" "$METRICS_URL" |
	grep '^http_response_size_bytes{route="/start"' || echo "no summary series"

[ "$OK" -eq "$N" ]
