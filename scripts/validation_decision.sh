#!/usr/bin/env bash
set -euo pipefail

# validation_decision.sh - Evaluate Three Doors validation criteria
# Usage: ./scripts/validation_decision.sh [path/to/sessions.jsonl]

SESSIONS_FILE="${1:-${HOME}/.threedoors/sessions.jsonl}"

# Check for jq
if ! command -v jq &>/dev/null; then
    echo "Error: jq is required but not installed. Install with: brew install jq" >&2
    exit 1
fi

# Check file exists
if [[ ! -f "$SESSIONS_FILE" ]]; then
    echo "No data found: $SESSIONS_FILE does not exist."
    echo "RESULT: CANNOT EVALUATE (no data)"
    exit 0
fi

if [[ ! -s "$SESSIONS_FILE" ]]; then
    echo "No data found: $SESSIONS_FILE is empty."
    echo "RESULT: CANNOT EVALUATE (no data)"
    exit 0
fi

echo "ThreeDoors Validation Decision"
echo "=============================="
echo ""

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_CRITERIA=4

# Criterion 1: Minimum 5 sessions
TOTAL_SESSIONS=$(wc -l < "$SESSIONS_FILE" | tr -d ' ')
if [[ "$TOTAL_SESSIONS" -ge 5 ]]; then
    echo "[PASS] Minimum sessions: $TOTAL_SESSIONS >= 5"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo "[FAIL] Minimum sessions: $TOTAL_SESSIONS < 5 (need at least 5)"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Criterion 2: Average completion rate > 0
AVG_COMPLETED=$(jq -s '[.[].tasks_completed] | add / length' "$SESSIONS_FILE")
COMPLETED_PASS=$(echo "$AVG_COMPLETED" | awk '{print ($1 > 0) ? "1" : "0"}')
if [[ "$COMPLETED_PASS" == "1" ]]; then
    printf "[PASS] Average completion rate: %.1f > 0 tasks/session\n" "$AVG_COMPLETED"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    printf "[FAIL] Average completion rate: %.1f <= 0 tasks/session\n" "$AVG_COMPLETED"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Criterion 3: Average time-to-first-door < 10 seconds (exclude -1 values)
AVG_TTFD=$(jq -s '[.[].time_to_first_door_seconds | select(. >= 0)] | if length == 0 then -1 else add / length end' "$SESSIONS_FILE")
TTFD_PASS=$(echo "$AVG_TTFD" | awk '{print ($1 >= 0 && $1 < 10) ? "1" : "0"}')
if [[ "$AVG_TTFD" == "-1" ]]; then
    echo "[FAIL] Time to first door: N/A (no door selections recorded)"
    FAIL_COUNT=$((FAIL_COUNT + 1))
elif [[ "$TTFD_PASS" == "1" ]]; then
    printf "[PASS] Time to first door: %.1fs < 10s\n" "$AVG_TTFD"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    printf "[FAIL] Time to first door: %.1fs >= 10s\n" "$AVG_TTFD"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Criterion 4: At least 50% of sessions have detail_views > 0
SESSIONS_WITH_DETAILS=$(jq -s '[.[] | select(.detail_views > 0)] | length' "$SESSIONS_FILE")
if [[ "$TOTAL_SESSIONS" -gt 0 ]]; then
    ENGAGEMENT_PCT=$((SESSIONS_WITH_DETAILS * 100 / TOTAL_SESSIONS))
    if [[ "$ENGAGEMENT_PCT" -ge 50 ]]; then
        echo "[PASS] Engagement: ${ENGAGEMENT_PCT}% of sessions have detail views (>= 50%)"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "[FAIL] Engagement: ${ENGAGEMENT_PCT}% of sessions have detail views (< 50%)"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
else
    echo "[FAIL] Engagement: no sessions to evaluate"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=============================="
echo "Results: $PASS_COUNT/$TOTAL_CRITERIA criteria passed"

if [[ "$PASS_COUNT" -eq "$TOTAL_CRITERIA" ]]; then
    echo "RESULT: PASS - Three Doors concept validated!"
else
    echo "RESULT: FAIL - $FAIL_COUNT criteria not met. Continue iterating."
fi
