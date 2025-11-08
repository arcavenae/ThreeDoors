#!/bin/bash
# ThreeDoors Validation Decision Helper
# Evaluates metrics against validation criteria to inform Epic 1 decision gate

SESSIONS_FILE="$HOME/.threedoors/sessions.jsonl"

if [[ ! -f "$SESSIONS_FILE" ]]; then
    echo "Error: No session data found at $SESSIONS_FILE"
    echo "Run ThreeDoors for at least a week before running validation."
    exit 1
fi

echo "╔════════════════════════════════════════════════════════════╗"
echo "║   ThreeDoors Validation Decision Helper                   ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Calculate metrics
avg_time_to_door=$(cat "$SESSIONS_FILE" | jq -r '.time_to_first_door_seconds' | awk '$1 >= 0 {sum+=$1; count++} END {if(count>0) printf "%.2f", sum/count; else print "-1"}')
refresh_rate=$(cat "$SESSIONS_FILE" | jq '.refreshes_used' | awk '$1 > 0 {refresh_sessions++} {total_sessions++} END {if(total_sessions>0) printf "%.1f", (refresh_sessions/total_sessions)*100; else print "0"}')
total_completions=$(cat "$SESSIONS_FILE" | jq '.tasks_completed' | awk '{sum+=$1} END {print sum}')
total_sessions=$(wc -l < "$SESSIONS_FILE")
avg_completions=$(echo "scale=2; $total_completions / $total_sessions" | bc)

# Get date range
first_session=$(cat "$SESSIONS_FILE" | head -1 | jq -r '.start_time' | cut -d'T' -f1)
last_session=$(cat "$SESSIONS_FILE" | tail -1 | jq -r '.start_time' | cut -d'T' -f1)

echo "📊 VALIDATION PERIOD:"
echo "  $first_session to $last_session ($total_sessions sessions)"
echo ""

echo "📈 METRIC SUMMARY:"
echo "  Average Time to First Door: ${avg_time_to_door}s (target: <30s)"
echo "  Refresh Usage Rate: ${refresh_rate}% (target: ≥30%)"
echo "  Average Completions/Session: ${avg_completions}"
echo "  Total Completions: ${total_completions}"
echo ""

echo "✓ VALIDATION CRITERIA:"
echo ""

# Initialize pass counter
passes=0
total_checks=4

# Check 1: Time to first door
if (( $(echo "$avg_time_to_door >= 0 && $avg_time_to_door < 30" | bc -l) )); then
    echo "  ✅ Time to First Door: PASS"
    echo "     → ${avg_time_to_door}s demonstrates low friction, faster than scrolling"
    passes=$((passes + 1))
elif (( $(echo "$avg_time_to_door < 0" | bc -l) )); then
    echo "  ⚠️  Time to First Door: NO DATA"
    echo "     → No doors selected in any session - unable to validate"
else
    echo "  ❌ Time to First Door: FAIL"
    echo "     → ${avg_time_to_door}s exceeds 30s target - not faster than list"
fi

echo ""

# Check 2: Refresh usage
if (( $(echo "$refresh_rate >= 30" | bc -l) )); then
    echo "  ✅ Refresh Usage: PASS"
    echo "     → ${refresh_rate}% demonstrates option is valuable to workflow"
    passes=$((passes + 1))
elif (( $(echo "$refresh_rate >= 20" | bc -l) )); then
    echo "  ⚠️  Refresh Usage: MARGINAL"
    echo "     → ${refresh_rate}% below target but shows some usage"
else
    echo "  ❌ Refresh Usage: FAIL"
    echo "     → ${refresh_rate}% suggests refresh option not valuable"
fi

echo ""

# Check 3: Completions
if (( $(echo "$avg_completions >= 1" | bc -l) )); then
    echo "  ✅ Completions: PASS"
    echo "     → ${avg_completions} per session demonstrates productivity"
    passes=$((passes + 1))
elif (( $(echo "$avg_completions >= 0.5" | bc -l) )); then
    echo "  ⚠️  Completions: MARGINAL"
    echo "     → ${avg_completions} per session below target - check task sizing"
else
    echo "  ❌ Completions: FAIL"
    echo "     → ${avg_completions} per session suggests tool not enabling progress"
fi

echo ""

# Check 4: Consistency (sessions over multiple days)
unique_days=$(cat "$SESSIONS_FILE" | jq -r '.start_time' | cut -d'T' -f1 | sort -u | wc -l)
if [[ $unique_days -ge 5 ]]; then
    echo "  ✅ Usage Consistency: PASS"
    echo "     → Used on $unique_days different days demonstrates habit formation"
    passes=$((passes + 1))
elif [[ $unique_days -ge 3 ]]; then
    echo "  ⚠️  Usage Consistency: MARGINAL"
    echo "     → Used on $unique_days days - suggest extending validation period"
else
    echo "  ❌ Usage Consistency: FAIL"
    echo "     → Used on $unique_days days - insufficient validation period"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Final recommendation
echo "🎯 FINAL RECOMMENDATION:"
echo ""

if [[ $passes -ge 3 ]]; then
    echo "  ✅ PROCEED TO EPIC 2"
    echo ""
    echo "  Rationale: $passes of $total_checks validation criteria passed."
    echo "  Metrics support the hypothesis that Three Doors UX reduces"
    echo "  friction compared to traditional task lists."
    echo ""
    echo "  Next Steps:"
    echo "  1. Review marginal/failed criteria and note for Epic 2 improvements"
    echo "  2. Begin Apple Notes integration planning (Epic 2)"
    echo "  3. Consider refactoring opportunities while implementing adapter pattern"
elif [[ $passes -ge 2 ]]; then
    echo "  ⚠️  CONDITIONAL PROCEED"
    echo ""
    echo "  Rationale: $passes of $total_checks validation criteria passed."
    echo "  Results are mixed. Consider:"
    echo ""
    echo "  Option A: Extend validation period to gather more data"
    echo "  Option B: Proceed to Epic 2 with awareness of weak areas"
    echo "  Option C: Iterate on Epic 1 to improve failed metrics before Epic 2"
    echo ""
    echo "  Recommendation: Review failed criteria closely and make informed decision."
else
    echo "  ❌ PIVOT OR ABANDON"
    echo ""
    echo "  Rationale: Only $passes of $total_checks validation criteria passed."
    echo "  Three Doors hypothesis not sufficiently validated."
    echo ""
    echo "  Options:"
    echo "  1. Pivot: Redesign UX based on learnings (e.g., different door selection)"
    echo "  2. Iterate: Extend validation period and improve failed areas"
    echo "  3. Abandon: Three Doors concept doesn't reduce friction for this user"
    echo ""
    echo "  Recommendation: Analyze session data for insights before deciding."
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
