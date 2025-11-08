#!/bin/bash
# ThreeDoors Session Analysis Script
# Analyzes sessions.jsonl to provide validation metrics

SESSIONS_FILE="$HOME/.threedoors/sessions.jsonl"

if [[ ! -f "$SESSIONS_FILE" ]]; then
    echo "No session data found at $SESSIONS_FILE"
    echo "Run ThreeDoors a few times to collect session data."
    exit 1
fi

echo "=== ThreeDoors Session Analysis ==="
echo ""

echo "📊 Total Sessions:"
wc -l < "$SESSIONS_FILE"
echo ""

echo "⏱️  Average Time to First Door (seconds):"
cat "$SESSIONS_FILE" | jq -r '.time_to_first_door_seconds' | awk '$1 >= 0 {sum+=$1; count++} END {if(count>0) printf "%.2f\n", sum/count; else print "N/A (no doors selected)"}'
echo ""

echo "✅ Total Tasks Completed:"
cat "$SESSIONS_FILE" | jq '.tasks_completed' | awk '{sum+=$1} END {print sum}'
echo ""

echo "📈 Average Tasks per Session:"
cat "$SESSIONS_FILE" | jq '.tasks_completed' | awk '{sum+=$1; count++} END {if(count>0) printf "%.2f\n", sum/count; else print 0}'
echo ""

echo "🔄 Refresh Usage Rate (% of sessions using refresh):"
cat "$SESSIONS_FILE" | jq '.refreshes_used' | awk '$1 > 0 {refresh_sessions++} {total_sessions++} END {if(total_sessions>0) printf "%.1f%%\n", (refresh_sessions/total_sessions)*100; else print "0%"}'
echo ""

echo "⏰ Average Session Duration (minutes):"
cat "$SESSIONS_FILE" | jq '.duration_seconds' | awk '{sum+=$1; count++} END {if(count>0) printf "%.2f\n", sum/count/60; else print 0}'
echo ""

echo "📅 Sessions by Date:"
cat "$SESSIONS_FILE" | jq -r '.start_time' | cut -d'T' -f1 | sort | uniq -c | awk '{print "  " $2 ": " $1 " session(s)"}'
echo ""

echo "🔍 Detail View Engagement:"
cat "$SESSIONS_FILE" | jq '[.detail_views, .doors_viewed] | @csv' | awk -F',' '{detail+=$1; doors+=$2} END {if(doors>0) printf "%.1f%% of door views led to detail view\n", (detail/doors)*100; else print "No door views yet"}'
echo ""

echo "📝 Notes Activity:"
cat "$SESSIONS_FILE" | jq '.notes_added' | awk '{sum+=$1; count++} END {if(count>0) printf "%.2f notes per session (total: %d)\n", sum/count, sum; else print "0 notes"}'
