#!/bin/bash
# test_p99.sh - P99 latency test

HTTP_PORT=8083
API_ENDPOINT="http://localhost:$HTTP_PORT/api/v1/speed"
TOTAL_REQUESTS=1000

echo "P99 Latency Test"
echo "Target: P99 < 10ms"
echo ""

TEMP_FILE=$(mktemp)

echo "Running $TOTAL_REQUESTS requests..."

for i in $(seq 1 $TOTAL_REQUESTS); do
    START=$(date +%s%N)
    curl -s -o /dev/null "$API_ENDPOINT?user_id=user_$i"
    END=$(date +%s%N)
    LATENCY=$(( (END - START) / 1000000 ))
    echo $LATENCY >> "$TEMP_FILE"
done

echo "Calculating results..."

SORTED=$(sort -n "$TEMP_FILE")
TOTAL=$(wc -l < "$TEMP_FILE")

P50=$(echo "$SORTED" | sed -n "$((TOTAL * 50 / 100))p")
P95=$(echo "$SORTED" | sed -n "$((TOTAL * 95 / 100))p")
P99=$(echo "$SORTED" | sed -n "$((TOTAL * 99 / 100))p")

rm -f "$TEMP_FILE"

echo ""
echo "Results:"
echo "  P50: ${P50}ms"
echo "  P95: ${P95}ms"
echo "  P99: ${P99}ms"
echo ""

if [ "$P99" -lt 10 ]; then
    echo "PASS: P99 ${P99}ms < 10ms"
    exit 0
else
    echo "FAIL: P99 ${P99}ms >= 10ms"
    exit 1
fi
