#!/usr/bin/env bash
set -euo pipefail

# analyze_sessions.sh - Analyze ThreeDoors session metrics
# Usage: ./scripts/analyze_sessions.sh [path/to/sessions.jsonl]

SESSIONS_FILE="${1:-${HOME}/.threedoors/sessions.jsonl}"

# Check for jq
if ! command -v jq &>/dev/null; then
    echo "Error: jq is required but not installed. Install with: brew install jq" >&2
    exit 1
fi

# Check file exists
if [[ ! -f "$SESSIONS_FILE" ]]; then
    echo "No data found: $SESSIONS_FILE does not exist."
    exit 0
fi

# Check file not empty
if [[ ! -s "$SESSIONS_FILE" ]]; then
    echo "No data found: $SESSIONS_FILE is empty."
    exit 0
fi

TOTAL_SESSIONS=$(wc -l < "$SESSIONS_FILE" | tr -d ' ')

if [[ "$TOTAL_SESSIONS" -eq 0 ]]; then
    echo "No data found: no sessions recorded."
    exit 0
fi

echo "ThreeDoors Session Analysis"
echo "==========================="
echo "Total sessions: $TOTAL_SESSIONS"

# Average duration (in minutes)
AVG_DURATION=$(jq -s '[.[].duration_seconds] | add / length / 60' "$SESSIONS_FILE")
printf "Average duration: %.1f minutes\n" "$AVG_DURATION"

# Average tasks completed
AVG_COMPLETED=$(jq -s '[.[].tasks_completed] | add / length' "$SESSIONS_FILE")
printf "Average tasks completed: %.1f per session\n" "$AVG_COMPLETED"

# Average refreshes
AVG_REFRESHES=$(jq -s '[.[].refreshes_used] | add / length' "$SESSIONS_FILE")
printf "Average refreshes: %.1f per session\n" "$AVG_REFRESHES"

# Average detail views
AVG_DETAILS=$(jq -s '[.[].detail_views] | add / length' "$SESSIONS_FILE")
printf "Average detail views: %.1f per session\n" "$AVG_DETAILS"

# Average time to first door (exclude -1 values = no door selected)
AVG_TTFD=$(jq -s '[.[].time_to_first_door_seconds | select(. >= 0)] | if length == 0 then "N/A" else (add / length | tostring + "s") end' "$SESSIONS_FILE" | tr -d '"')
echo "Average time to first door: $AVG_TTFD"

# Door position preference
TOTAL_SELECTIONS=$(jq -s '[.[].door_selections | length] | add' "$SESSIONS_FILE")
if [[ "$TOTAL_SELECTIONS" -gt 0 ]]; then
    LEFT=$(jq -s '[.[].door_selections[] | select(.door_position == 0)] | length' "$SESSIONS_FILE")
    CENTER=$(jq -s '[.[].door_selections[] | select(.door_position == 1)] | length' "$SESSIONS_FILE")
    RIGHT=$(jq -s '[.[].door_selections[] | select(.door_position == 2)] | length' "$SESSIONS_FILE")
    LEFT_PCT=$((LEFT * 100 / TOTAL_SELECTIONS))
    CENTER_PCT=$((CENTER * 100 / TOTAL_SELECTIONS))
    RIGHT_PCT=$((RIGHT * 100 / TOTAL_SELECTIONS))
    echo "Door position preference: Left=${LEFT_PCT}% Center=${CENTER_PCT}% Right=${RIGHT_PCT}%"
else
    echo "Door position preference: N/A (no selections)"
fi

# Mood distribution
MOOD_DATA=$(jq -s '[.[].mood_entries_detail[] | .mood] | group_by(.) | map({mood: .[0], count: length}) | sort_by(-.count)' "$SESSIONS_FILE" 2>/dev/null || echo "[]")
if [[ "$MOOD_DATA" != "[]" ]]; then
    MOOD_LINE=$(echo "$MOOD_DATA" | jq -r 'map(.mood + "=" + (.count | tostring)) | join(" ")')
    echo "Mood distribution: $MOOD_LINE"
else
    echo "Mood distribution: N/A (no mood entries)"
fi
