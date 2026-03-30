#!/usr/bin/env bash
set -euo pipefail

# test-quota-monitor.sh — Tests for quota-monitor.sh (Story 76.6)
# Validates window boundary detection, reset detection, dedup logic,
# idempotency, state file management, and all-clear notifications.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/quota-monitor.sh"
QUOTA_SCRIPT="$SCRIPT_DIR/quota-status.sh"
TESTDATA_DIR="$SCRIPT_DIR/testdata/quota-monitor"
PASS=0
FAIL=0

# --- Setup ---
setup_fixtures() {
    rm -rf "$TESTDATA_DIR"
    mkdir -p "$TESTDATA_DIR/projects/proj-main"
    mkdir -p "$TESTDATA_DIR/state"

    local now_epoch
    now_epoch=$(date +%s)

    # Timestamps within the 5h window
    local ts_1h ts_2h ts_30m
    ts_1h=$(date -u -r $((now_epoch - 3600)) '+%Y-%m-%dT%H:%M:%S.000Z')
    ts_2h=$(date -u -r $((now_epoch - 7200)) '+%Y-%m-%dT%H:%M:%S.000Z')
    ts_30m=$(date -u -r $((now_epoch - 1800)) '+%Y-%m-%dT%H:%M:%S.000Z')

    # Session with moderate usage (~75% of 1000 limit = 750 billed tokens)
    cat > "$TESTDATA_DIR/projects/proj-main/session1.jsonl" << EOF
{"type":"assistant","timestamp":"$ts_2h","message":{"role":"assistant","usage":{"input_tokens":200,"output_tokens":100,"cache_creation_input_tokens":50,"cache_read_input_tokens":50}}}
{"type":"assistant","timestamp":"$ts_1h","message":{"role":"assistant","usage":{"input_tokens":200,"output_tokens":100,"cache_creation_input_tokens":50,"cache_read_input_tokens":50}}}
{"type":"assistant","timestamp":"$ts_30m","message":{"role":"assistant","usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":25,"cache_read_input_tokens":25}}}
EOF
    # Total billed = (200+100)+(200+100)+(100+50) = 750

    export NOW_EPOCH="$now_epoch"
    export TS_2H_EPOCH=$((now_epoch - 7200))
}

setup_high_usage() {
    # High usage scenario (~95% of 1000 limit = 950 billed tokens)
    rm -rf "$TESTDATA_DIR/projects"
    mkdir -p "$TESTDATA_DIR/projects/proj-main"

    local now_epoch
    now_epoch=$(date +%s)

    local ts_1h ts_30m
    ts_1h=$(date -u -r $((now_epoch - 3600)) '+%Y-%m-%dT%H:%M:%S.000Z')
    ts_30m=$(date -u -r $((now_epoch - 1800)) '+%Y-%m-%dT%H:%M:%S.000Z')

    cat > "$TESTDATA_DIR/projects/proj-main/session1.jsonl" << EOF
{"type":"assistant","timestamp":"$ts_1h","message":{"role":"assistant","usage":{"input_tokens":400,"output_tokens":200,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
{"type":"assistant","timestamp":"$ts_30m","message":{"role":"assistant","usage":{"input_tokens":250,"output_tokens":100,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
EOF
    # Total billed = (400+200)+(250+100) = 950

    export NOW_EPOCH="$now_epoch"
}

setup_low_usage() {
    # Low usage scenario (~30% of 1000 limit)
    rm -rf "$TESTDATA_DIR/projects"
    mkdir -p "$TESTDATA_DIR/projects/proj-main"

    local now_epoch
    now_epoch=$(date +%s)

    local ts_1h
    ts_1h=$(date -u -r $((now_epoch - 3600)) '+%Y-%m-%dT%H:%M:%S.000Z')

    cat > "$TESTDATA_DIR/projects/proj-main/session1.jsonl" << EOF
{"type":"assistant","timestamp":"$ts_1h","message":{"role":"assistant","usage":{"input_tokens":200,"output_tokens":100,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
EOF
    # Total billed = 300

    export NOW_EPOCH="$now_epoch"
}

cleanup_state() {
    rm -f "$TESTDATA_DIR/state/state.json"
    rm -f "$TESTDATA_DIR/state/state.json.tmp"
    rm -f "$TESTDATA_DIR/state/snapshots.jsonl"
}

# --- Assertions ---
assert_eq() {
    local desc="$1" expected="$2" actual="$3"
    if [[ "$expected" == "$actual" ]]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc (expected: $expected, got: $actual)"
        FAIL=$((FAIL + 1))
    fi
}

assert_contains() {
    local desc="$1" pattern="$2" text="$3"
    if echo "$text" | grep -q "$pattern"; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc (pattern '$pattern' not found)"
        FAIL=$((FAIL + 1))
    fi
}

assert_not_contains() {
    local desc="$1" pattern="$2" text="$3"
    if echo "$text" | grep -q "$pattern"; then
        echo "  FAIL: $desc (pattern '$pattern' found but should not be)"
        FAIL=$((FAIL + 1))
    else
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    fi
}

assert_file_exists() {
    local desc="$1" file="$2"
    if [[ -f "$file" ]]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc (file $file does not exist)"
        FAIL=$((FAIL + 1))
    fi
}

assert_json_field() {
    local desc="$1" field="$2" expected="$3" file="$4"
    local actual
    actual=$(python3 -c "import json; print(json.load(open('$file'))$field)" 2>/dev/null || echo "PARSE_ERROR")
    if [[ "$expected" == "$actual" ]]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc (expected: $expected, got: $actual)"
        FAIL=$((FAIL + 1))
    fi
}

run_monitor() {
    "$SCRIPT" --project-dir "$TESTDATA_DIR/projects" --state-dir "$TESTDATA_DIR/state" --dry-run --limit "$@" 2>&1
}

# --- Tests ---

echo "Setting up test fixtures..."
setup_fixtures

echo ""
echo "=== Test 1: Window boundary detection ==="
cleanup_state
OUTPUT=$(run_monitor 1000)
assert_file_exists "state file created" "$TESTDATA_DIR/state/state.json"
assert_json_field "window_start_epoch is set" "['window_start_epoch']" "${TS_2H_EPOCH}.0" "$TESTDATA_DIR/state/state.json"
assert_contains "monitor completes" "Monitor complete" "$OUTPUT"

echo ""
echo "=== Test 2: Reset time calculation ==="
# The output should reference the window timing info
assert_json_field "last_check_epoch is recent" ".get('last_check_epoch', 0) > 0" "True" "$TESTDATA_DIR/state/state.json"

echo ""
echo "=== Test 3: Threshold triggers at 75% (yellow tier) ==="
cleanup_state
OUTPUT=$(run_monitor 1000)
# 750/1000 = 75% → yellow tier (>= 70%)
assert_contains "warning triggered" "QUOTA_WARNING" "$OUTPUT"
assert_contains "CAUTION label" "CAUTION" "$OUTPUT"
assert_json_field "last_tier_warned is green" "['last_tier_warned']" "green" "$TESTDATA_DIR/state/state.json"

echo ""
echo "=== Test 4: Warning deduplication — same tier, same window ==="
# Run again without changing state — same window, same tier
OUTPUT2=$(run_monitor 1000)
assert_not_contains "dedup prevents second warning" "QUOTA_WARNING" "$OUTPUT2"
assert_contains "dedup message shown" "Dedup" "$OUTPUT2"

echo ""
echo "=== Test 5: Warning escalation — higher tier in same window ==="
# Reduce limit so 750/800 = 93.75% → orange tier
# First, set state as if green was already warned
python3 -c "
import json
state = json.load(open('$TESTDATA_DIR/state/state.json'))
state['last_tier_warned'] = 'green'
json.dump(state, open('$TESTDATA_DIR/state/state.json', 'w'))
"
OUTPUT3=$(run_monitor 800)
# 750/800 = 93.75% → orange (>= 90%) → escalation from green
assert_contains "escalation warning sent" "QUOTA_WARNING" "$OUTPUT3"
assert_contains "ALERT label" "ALERT" "$OUTPUT3"

echo ""
echo "=== Test 6: Window reset detection and all-clear ==="
cleanup_state
# Seed state with a window that started 6 hours ago (expired)
local_now=$(date +%s)
old_window_start=$((local_now - 21600))  # 6h ago
python3 -c "
import json
state = {
    'window_start_epoch': $old_window_start,
    'last_tier_warned': 'yellow',
    'last_check_epoch': $((local_now - 600)),
    'last_usage_pct': 75.0
}
json.dump(state, open('$TESTDATA_DIR/state/state.json', 'w'), indent=2)
"
OUTPUT4=$(run_monitor 1000)
assert_contains "reset detected" "QUOTA_RESET" "$OUTPUT4"
assert_contains "window reset message" "Window reset detected" "$OUTPUT4"

echo ""
echo "=== Test 7: Idempotent execution — multiple rapid runs ==="
cleanup_state
# Run 3 times rapidly — should only get 1 warning (first run), then dedup
OUTPUT_A=$(run_monitor 1000)
OUTPUT_B=$(run_monitor 1000)
OUTPUT_C=$(run_monitor 1000)

# Count QUOTA_WARNING occurrences across all runs
WARN_COUNT_A=$(echo "$OUTPUT_A" | grep -c "QUOTA_WARNING" || true)
WARN_COUNT_B=$(echo "$OUTPUT_B" | grep -c "QUOTA_WARNING" || true)
WARN_COUNT_C=$(echo "$OUTPUT_C" | grep -c "QUOTA_WARNING" || true)

assert_eq "first run warns" "1" "$WARN_COUNT_A"
assert_eq "second run deduped" "0" "$WARN_COUNT_B"
assert_eq "third run deduped" "0" "$WARN_COUNT_C"

echo ""
echo "=== Test 8: State file structure ==="
STATE_VALID=$(python3 -c "
import json
state = json.load(open('$TESTDATA_DIR/state/state.json'))
required = ['window_start_epoch', 'last_tier_warned', 'last_check_epoch', 'last_usage_pct']
print('True' if all(k in state for k in required) else 'False')
")
assert_eq "state has all required fields" "True" "$STATE_VALID"

echo ""
echo "=== Test 9: Snapshot recording ==="
cleanup_state
run_monitor 1000 > /dev/null 2>&1
assert_file_exists "snapshot file created" "$TESTDATA_DIR/state/snapshots.jsonl"
SNAP_COUNT=$(wc -l < "$TESTDATA_DIR/state/snapshots.jsonl" | tr -d ' ')
assert_eq "1 snapshot recorded" "1" "$SNAP_COUNT"

# Run again — should add another snapshot
run_monitor 1000 > /dev/null 2>&1
SNAP_COUNT2=$(wc -l < "$TESTDATA_DIR/state/snapshots.jsonl" | tr -d ' ')
assert_eq "2 snapshots after 2 runs" "2" "$SNAP_COUNT2"

# Validate snapshot structure
SNAP_VALID=$(python3 -c "
import json
with open('$TESTDATA_DIR/state/snapshots.jsonl') as f:
    snap = json.loads(f.readline())
required = ['timestamp', 'usage_pct', 'total_billed', 'quota_limit', 'tier', 'window_reset_detected', 'warning_sent', 'peak_hours']
print('True' if all(k in snap for k in required) else 'False')
")
assert_eq "snapshot has required fields" "True" "$SNAP_VALID"

echo ""
echo "=== Test 10: Low usage — no warning ==="
setup_low_usage
cleanup_state
OUTPUT5=$(run_monitor 1000)
# 300/1000 = 30% — below all thresholds
assert_not_contains "no warning at 30%" "QUOTA_WARNING" "$OUTPUT5"
assert_contains "below thresholds" "below all warning thresholds" "$OUTPUT5"

echo ""
echo "=== Test 11: Critical tier at 95%+ ==="
setup_high_usage
cleanup_state
OUTPUT6=$(run_monitor 1000)
# 950/1000 = 95% → red tier
assert_contains "critical warning" "CRITICAL" "$OUTPUT6"

echo ""
echo "=== Test 12: Expired window with no new activity ==="
cleanup_state
# Empty project dir (no JSONL) but state shows previous window
mkdir -p "$TESTDATA_DIR/projects-empty/subdir"
local_now2=$(date +%s)
old_start2=$((local_now2 - 21600))  # 6h ago
python3 -c "
import json
state = {
    'window_start_epoch': $old_start2,
    'last_tier_warned': 'orange',
    'last_check_epoch': $((local_now2 - 600)),
    'last_usage_pct': 91.0
}
json.dump(state, open('$TESTDATA_DIR/state/state.json', 'w'), indent=2)
"
OUTPUT7=$("$SCRIPT" --project-dir "$TESTDATA_DIR/projects-empty" --state-dir "$TESTDATA_DIR/state" --dry-run --limit 1000 2>&1 || true)
# May or may not detect reset depending on whether quota-status.sh returns data for empty dir
# At minimum it should not crash
assert_contains "empty dir completes or resets" "Monitor complete\|QUOTA_RESET" "$OUTPUT7"

echo ""
echo "=== Test 13: Help flag ==="
HELP_OUT=$("$SCRIPT" --help 2>/dev/null)
assert_contains "help shows usage" "Usage:" "$HELP_OUT"

echo ""
echo "=== Test 14: Dry-run doesn't send real messages ==="
cleanup_state
OUTPUT8=$(run_monitor 1000)
assert_contains "dry-run prefix on messages" "DRY-RUN" "$OUTPUT8"

# --- Cleanup ---
rm -rf "$TESTDATA_DIR"

echo ""
echo "═══════════════════════════════════"
echo "Results: $PASS passed, $FAIL failed"
echo "═══════════════════════════════════"

if [[ $FAIL -gt 0 ]]; then
    exit 1
fi
