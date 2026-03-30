#!/usr/bin/env bash
set -euo pipefail

# test-quota-status.sh — Tests for quota-status.sh (Story 73.5)
# Validates token parsing, window calculation, color coding, and per-agent breakdown.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/quota-status.sh"
TESTDATA_DIR="$SCRIPT_DIR/testdata/quota"
PASS=0
FAIL=0

# --- Setup test fixtures ---
setup_fixtures() {
    rm -rf "$TESTDATA_DIR"
    mkdir -p "$TESTDATA_DIR/proj-main"
    mkdir -p "$TESTDATA_DIR/proj-wts-ThreeDoors-bold-fox"
    mkdir -p "$TESTDATA_DIR/proj-wts-ThreeDoors-clever-owl"

    # Current time for timestamps within window
    local now_epoch
    now_epoch=$(date +%s)
    local ts1 ts2 ts3 ts4
    ts1=$(date -u -r $((now_epoch - 3600)) '+%Y-%m-%dT%H:%M:%S.000Z')   # 1h ago
    ts2=$(date -u -r $((now_epoch - 7200)) '+%Y-%m-%dT%H:%M:%S.000Z')   # 2h ago
    ts3=$(date -u -r $((now_epoch - 1800)) '+%Y-%m-%dT%H:%M:%S.000Z')   # 30m ago
    ts4=$(date -u -r $((now_epoch - 86400)) '+%Y-%m-%dT%H:%M:%S.000Z')  # 24h ago (outside window)

    # Main checkout session — 2 interactions within window, 1 outside
    cat > "$TESTDATA_DIR/proj-main/session1.jsonl" << EOF
{"type":"human","timestamp":"$ts2","cwd":"/test","sessionId":"s1","message":"hello"}
{"type":"assistant","timestamp":"$ts2","message":{"role":"assistant","usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":200,"cache_read_input_tokens":300}}}
{"type":"assistant","timestamp":"$ts1","message":{"role":"assistant","usage":{"input_tokens":150,"output_tokens":75,"cache_creation_input_tokens":100,"cache_read_input_tokens":50}}}
{"type":"assistant","timestamp":"$ts4","message":{"role":"assistant","usage":{"input_tokens":9999,"output_tokens":9999,"cache_creation_input_tokens":9999,"cache_read_input_tokens":9999}}}
EOF

    # bold-fox session — 1 interaction
    cat > "$TESTDATA_DIR/proj-wts-ThreeDoors-bold-fox/session2.jsonl" << EOF
{"type":"assistant","timestamp":"$ts3","message":{"role":"assistant","usage":{"input_tokens":500,"output_tokens":200,"cache_creation_input_tokens":0,"cache_read_input_tokens":1000}}}
EOF

    # clever-owl session — 1 interaction
    cat > "$TESTDATA_DIR/proj-wts-ThreeDoors-clever-owl/session3.jsonl" << EOF
{"type":"assistant","timestamp":"$ts1","message":{"role":"assistant","usage":{"input_tokens":50,"output_tokens":25,"cache_creation_input_tokens":10,"cache_read_input_tokens":15}}}
EOF

    # File with corrupt JSON (should be handled gracefully)
    cat > "$TESTDATA_DIR/proj-main/corrupt.jsonl" << EOF
not valid json
{"type":"assistant","timestamp":"$ts1","message":{"role":"assistant","usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
EOF
}

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

assert_json_field() {
    local desc="$1" field="$2" expected="$3" json="$4"
    local actual
    actual=$(python3 -c "import json,sys; print(json.loads(sys.stdin.read())$field)" <<< "$json" 2>/dev/null || echo "PARSE_ERROR")
    if [[ "$expected" == "$actual" ]]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc (expected: $expected, got: $actual)"
        FAIL=$((FAIL + 1))
    fi
}

# --- Tests ---

echo "Setting up test fixtures..."
setup_fixtures

echo ""
echo "=== Test 1: JSON output with correct token sums ==="
# Expected tokens within window (excluding the 24h-ago entry and including corrupt file's valid line):
#   main session1: (100+50) + (150+75) = 375
#   main corrupt:  (10+5) = 15
#   bold-fox:      (500+200) = 700
#   clever-owl:    (50+25) = 75
#   Total billed: 375 + 15 + 700 + 75 = 1165
JSON_OUT=$("$SCRIPT" --project-dir "$TESTDATA_DIR" --limit 88000 --json --no-color 2>/dev/null)

assert_json_field "total_billed is 1165" "['total_billed']" "1165" "$JSON_OUT"
assert_json_field "total_input is 810" "['total_input']" "810" "$JSON_OUT"
assert_json_field "total_output is 355" "['total_output']" "355" "$JSON_OUT"
assert_json_field "interaction_count is 5" "['interaction_count']" "5" "$JSON_OUT"

echo ""
echo "=== Test 2: Per-agent breakdown ==="
# Check agents are identified
AGENT_COUNT=$(python3 -c "import json,sys; print(len(json.loads(sys.stdin.read())['agents']))" <<< "$JSON_OUT")
assert_eq "3 agents detected" "3" "$AGENT_COUNT"

echo ""
echo "=== Test 3: Usage percentage and level ==="
assert_json_field "usage_pct is 1.3" "['usage_pct']" "1.3" "$JSON_OUT"
assert_json_field "level is green (< 70%)" "['level']" "green" "$JSON_OUT"

# Test yellow threshold (1165/1553 ≈ 75%)
YELLOW_OUT=$("$SCRIPT" --project-dir "$TESTDATA_DIR" --limit 1553 --json --no-color 2>/dev/null)
assert_json_field "level is yellow at ~75%" "['level']" "yellow" "$YELLOW_OUT"

# Test red threshold (1165/1354 ≈ 86%)
RED_OUT=$("$SCRIPT" --project-dir "$TESTDATA_DIR" --limit 1354 --json --no-color 2>/dev/null)
assert_json_field "level is red at ~86%" "['level']" "red" "$RED_OUT"

echo ""
echo "=== Test 4: Formatted output contains expected sections ==="
FMT_OUT=$("$SCRIPT" --project-dir "$TESTDATA_DIR" --limit 88000 --no-color 2>/dev/null)
assert_contains "header present" "Claude Quota Status" "$FMT_OUT"
assert_contains "usage line present" "88,000 billed tokens" "$FMT_OUT"
assert_contains "breakdown section" "Input tokens:" "$FMT_OUT"
assert_contains "window section" "Window size:" "$FMT_OUT"
assert_contains "per-agent section" "Per-Agent Consumption" "$FMT_OUT"

echo ""
echo "=== Test 5: Window boundary — old entries excluded ==="
# The 24h-ago entry (9999+9999 = 19998 billed) should NOT be in the total
# Total should be 1165, not 1165 + 19998
assert_json_field "old entries excluded" "['total_billed']" "1165" "$JSON_OUT"

echo ""
echo "=== Test 6: Corrupt JSONL handled gracefully ==="
# Script should not crash; the valid line in corrupt.jsonl should be counted
assert_contains "script completes with corrupt file" "Claude Quota Status" "$FMT_OUT"

echo ""
echo "=== Test 7: Help flag ==="
HELP_OUT=$("$SCRIPT" --help 2>/dev/null)
assert_contains "help shows usage" "Usage:" "$HELP_OUT"

echo ""
echo "=== Test 8: Empty project dir ==="
mkdir -p "$TESTDATA_DIR/empty-proj"
EMPTY_OUT=$("$SCRIPT" --project-dir "$TESTDATA_DIR/empty-proj" --json --no-color 2>/dev/null)
assert_json_field "empty dir yields 0 tokens" "['total_billed']" "0" "$EMPTY_OUT"
assert_json_field "empty dir is green" "['level']" "green" "$EMPTY_OUT"

# --- Cleanup ---
rm -rf "$TESTDATA_DIR"

echo ""
echo "═══════════════════════════════════"
echo "Results: $PASS passed, $FAIL failed"
echo "═══════════════════════════════════"

if [[ $FAIL -gt 0 ]]; then
    exit 1
fi
