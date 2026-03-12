#!/usr/bin/env bash
set -euo pipefail

# test-supervisor-handover.sh — Tests for Story 58.7 scripts
# Run from the repository root: ./scripts/test-supervisor-handover.sh
#
# Tests manual handover trigger (supervisor-handover.sh),
# handover notifications (handover-notify.sh),
# and shift status display (shift-status.sh).

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
HANDOVER_SCRIPT="$SCRIPT_DIR/supervisor-handover.sh"
NOTIFY_SCRIPT="$SCRIPT_DIR/handover-notify.sh"
STATUS_SCRIPT="$SCRIPT_DIR/shift-status.sh"
PASS=0
FAIL=0
TMPDIR_BASE=""
MOCK_BIN_DIR=""

pass() {
    PASS=$((PASS + 1))
    echo "  PASS: $1"
}

fail() {
    FAIL=$((FAIL + 1))
    echo "  FAIL: $1" >&2
}

setup_tmpdir() {
    TMPDIR_BASE="$(mktemp -d)"
}

cleanup() {
    if [[ -n "$TMPDIR_BASE" ]] && [[ -d "$TMPDIR_BASE" ]]; then
        rm -rf "$TMPDIR_BASE"
    fi
    if [[ -n "${MOCK_BIN_DIR:-}" ]] && [[ -d "$MOCK_BIN_DIR" ]]; then
        rm -rf "$MOCK_BIN_DIR"
    fi
}

trap cleanup EXIT

# Create mock commands
create_mock_bins() {
    MOCK_BIN_DIR="$(mktemp -d)"

    # Mock multiclaude
    cat > "$MOCK_BIN_DIR/multiclaude" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    repo)
        if [[ "${2:-}" == "current" ]]; then
            echo "ThreeDoors"
        fi
        ;;
    message)
        if [[ "${2:-}" == "list" ]]; then
            echo "No pending messages"
        elif [[ "${2:-}" == "send" ]]; then
            echo "Message sent to ${3:-unknown}: ${4:-}"
        fi
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/multiclaude"

    # Mock git
    cat > "$MOCK_BIN_DIR/git" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    remote) echo "https://github.com/arcaven/ThreeDoors.git" ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/git"

    # Mock bc (for file size display)
    if command -v bc &>/dev/null; then
        ln -sf "$(command -v bc)" "$MOCK_BIN_DIR/bc"
    fi

    # Pass through system commands
    for cmd in stat grep date cat wc mv mkdir chmod head base64 tr printf rm read sync; do
        if command -v "$cmd" &>/dev/null; then
            ln -sf "$(command -v "$cmd")" "$MOCK_BIN_DIR/$cmd"
        fi
    done
}

# Create a synthetic JSONL transcript for status tests
create_transcript() {
    local dir="$1"
    local compressions="${2:-0}"
    local assistant_messages="${3:-10}"
    local start_time="${4:-2026-03-12T10:00:00}"

    local project_dir="$dir/.claude/projects/-test-ThreeDoors"
    mkdir -p "$project_dir"
    local transcript="$project_dir/test-session.jsonl"

    echo "{\"type\":\"system\",\"subtype\":\"init\",\"timestamp\":\"${start_time}.000Z\",\"sessionId\":\"test-session\"}" > "$transcript"

    for i in $(seq 1 "$assistant_messages"); do
        echo "{\"type\":\"assistant\",\"timestamp\":\"${start_time}.${i}00Z\"}" >> "$transcript"
    done

    for i in $(seq 1 "$compressions"); do
        echo "{\"type\":\"system\",\"subtype\":\"compact_boundary\",\"content\":\"Conversation compacted\",\"timestamp\":\"${start_time}.${i}00Z\",\"compactMetadata\":{\"trigger\":\"auto\",\"preTokens\":167000}}" >> "$transcript"
    done

    echo "$transcript"
}

echo "=========================================="
echo "=== Story 58.7: Manual Handover Tests ==="
echo "=========================================="

setup_tmpdir
create_mock_bins

# =============================================
# Section 1: supervisor-handover.sh tests
# =============================================

echo ""
echo "=== supervisor-handover.sh ==="

# Test: --help exits successfully
echo "Test: --help exits successfully"
if "$HANDOVER_SCRIPT" --help >/dev/null 2>&1; then
    pass "--help exits 0"
else
    fail "--help should exit 0"
fi

# Test: --help shows usage
echo "Test: --help shows usage"
HELP_OUTPUT="$("$HANDOVER_SCRIPT" --help 2>&1)"
if echo "$HELP_OUTPUT" | grep -q "Usage"; then
    pass "--help shows usage"
else
    fail "--help should show usage"
fi

# Test: unknown option fails
echo "Test: unknown option fails"
if "$HANDOVER_SCRIPT" --bogus 2>/dev/null; then
    fail "unknown option should fail"
else
    pass "unknown option exits non-zero"
fi

# Test AC-1: Manual handover creates signal file
echo "Test AC-1: Manual handover creates signal file"
HANDOVER_DIR="$TMPDIR_BASE/ac1-signal"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$HANDOVER_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" 2>&1)"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "signal file created"
else
    fail "should create signal file (output: $OUTPUT)"
fi

# Test AC-1: Signal file contains manual trigger
echo "Test AC-1: Signal file contains manual trigger"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    SIGNAL="$(cat "$HANDOVER_DIR/handover-requested")"
    if echo "$SIGNAL" | grep -q 'trigger: "manual"'; then
        pass "signal file has trigger: manual"
    else
        fail "signal file should have trigger: manual (got: $SIGNAL)"
    fi
    if echo "$SIGNAL" | grep -q 'zone: "MANUAL"'; then
        pass "signal file has zone: MANUAL"
    else
        fail "signal file should have zone: MANUAL (got: $SIGNAL)"
    fi
    if echo "$SIGNAL" | grep -q "timestamp:"; then
        pass "signal file has timestamp"
    else
        fail "signal file should have timestamp (got: $SIGNAL)"
    fi
else
    fail "signal file not found for content check"
fi

# Test AC-1: Output confirms handover requested
echo "Test AC-1: Output confirms handover requested"
if echo "$OUTPUT" | grep -q "HANDOVER REQUESTED"; then
    pass "output confirms handover requested"
else
    fail "output should confirm handover (got: $OUTPUT)"
fi

# Test AC-2: Anti-oscillation warning with recent handover (non-interactive)
echo "Test AC-2: Anti-oscillation warning blocks non-interactive"
HANDOVER_DIR="$TMPDIR_BASE/ac2-anti-osc"
mkdir -p "$HANDOVER_DIR"
# Write a recent handover timestamp (5 minutes ago)
echo "$(( $(date +%s) - 300 ))" > "$HANDOVER_DIR/last-handover-timestamp"
EXIT_CODE=0
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$HANDOVER_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" 2>&1)" || EXIT_CODE=$?
if [[ "$EXIT_CODE" -ne 0 ]]; then
    pass "non-interactive anti-oscillation exits non-zero"
else
    fail "should exit non-zero when anti-oscillation blocks (exit: $EXIT_CODE)"
fi
if echo "$OUTPUT" | grep -q "WARNING.*minutes ago"; then
    pass "anti-oscillation warning displayed"
else
    fail "should display anti-oscillation warning (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Use --force"; then
    pass "suggests --force flag"
else
    fail "should suggest --force flag (got: $OUTPUT)"
fi
# Signal file should NOT be created
if [[ ! -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "no signal file when anti-oscillation blocks"
else
    fail "should not create signal file when anti-oscillation blocks"
fi

# Test AC-2: Anti-oscillation allows handover after gap
echo "Test AC-2: Anti-oscillation allows after sufficient gap"
HANDOVER_DIR="$TMPDIR_BASE/ac2-gap-ok"
mkdir -p "$HANDOVER_DIR"
# Write an old handover timestamp (60 minutes ago)
echo "$(( $(date +%s) - 3600 ))" > "$HANDOVER_DIR/last-handover-timestamp"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$HANDOVER_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" 2>&1)"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "handover allowed after gap"
else
    fail "should allow handover after sufficient gap (got: $OUTPUT)"
fi

# Test AC-2: Anti-oscillation allows when no previous handover
echo "Test AC-2: Anti-oscillation allows first-ever handover"
HANDOVER_DIR="$TMPDIR_BASE/ac2-first"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$HANDOVER_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" 2>&1)"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "first handover allowed"
else
    fail "first handover should always be allowed (got: $OUTPUT)"
fi

# Test AC-3: --force bypasses anti-oscillation
echo "Test AC-3: --force bypasses anti-oscillation"
HANDOVER_DIR="$TMPDIR_BASE/ac3-force"
mkdir -p "$HANDOVER_DIR"
# Write a very recent handover timestamp (1 minute ago)
echo "$(( $(date +%s) - 60 ))" > "$HANDOVER_DIR/last-handover-timestamp"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$HANDOVER_SCRIPT" --force --repo ThreeDoors --handover-dir "$HANDOVER_DIR" 2>&1)"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "--force creates signal despite recent handover"
else
    fail "--force should create signal despite recent handover (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Force mode"; then
    pass "--force mode announced"
else
    fail "should announce force mode (got: $OUTPUT)"
fi

# Test AC-3: --force signal file contains force flag
echo "Test AC-3: --force signal file contains force flag"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    SIGNAL="$(cat "$HANDOVER_DIR/handover-requested")"
    if echo "$SIGNAL" | grep -q "force: true"; then
        pass "signal file contains force: true"
    else
        fail "signal file should contain force: true (got: $SIGNAL)"
    fi
else
    fail "signal file not found for force flag check"
fi

# =============================================
# Section 2: handover-notify.sh tests
# =============================================

echo ""
echo "=== handover-notify.sh ==="

# Test: --help exits successfully
echo "Test: --help exits successfully"
if "$NOTIFY_SCRIPT" --help >/dev/null 2>&1; then
    pass "notify --help exits 0"
else
    fail "notify --help should exit 0"
fi

# Test: start --help exits successfully
echo "Test: start --help exits successfully"
if "$NOTIFY_SCRIPT" start --help >/dev/null 2>&1; then
    pass "notify start --help exits 0"
else
    fail "notify start --help should exit 0"
fi

# Test: unknown subcommand fails
echo "Test: unknown subcommand fails"
if PATH="$MOCK_BIN_DIR:$PATH" "$NOTIFY_SCRIPT" bogus --repo ThreeDoors 2>/dev/null; then
    fail "unknown subcommand should fail"
else
    pass "unknown subcommand exits non-zero"
fi

# Test AC-4: Start notification for automatic handover
echo "Test AC-4: Automatic handover start notification"
HANDOVER_DIR="$TMPDIR_BASE/ac4-auto"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$NOTIFY_SCRIPT" start --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --trigger shift-clock --zone RED --jsonl-size 10485760 --compressions 6 2>&1)"
if echo "$OUTPUT" | grep -q "SUPERVISOR_HANDOVER:"; then
    pass "start notification contains SUPERVISOR_HANDOVER"
else
    fail "start notification should contain SUPERVISOR_HANDOVER (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Automatic"; then
    pass "start notification says Automatic"
else
    fail "start notification should say Automatic for shift-clock trigger (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "RED"; then
    pass "start notification includes zone"
else
    fail "start notification should include zone (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "compressions: 6"; then
    pass "start notification includes compression metrics"
else
    fail "start notification should include compression metrics (got: $OUTPUT)"
fi

# Test AC-4: Start notification for manual handover
echo "Test AC-4: Manual handover start notification"
HANDOVER_DIR="$TMPDIR_BASE/ac4-manual"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$NOTIFY_SCRIPT" start --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --trigger manual --zone MANUAL 2>&1)"
if echo "$OUTPUT" | grep -q "Manual"; then
    pass "start notification says Manual for manual trigger"
else
    fail "start notification should say Manual (got: $OUTPUT)"
fi

# Test AC-4: Start notification creates log entry
echo "Test AC-4: Start notification creates log entry"
if [[ -f "$HANDOVER_DIR/handover.log" ]]; then
    LOG="$(cat "$HANDOVER_DIR/handover.log")"
    if echo "$LOG" | grep -q "Sending start notification"; then
        pass "start notification logged"
    else
        fail "start notification should be logged (got: $LOG)"
    fi
else
    fail "handover.log should exist after notification"
fi

# Test AC-5: Completion notification
echo "Test AC-5: Handover completion notification"
HANDOVER_DIR="$TMPDIR_BASE/ac5-complete"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$NOTIFY_SCRIPT" complete --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --supervisor supervisor --duration 45 --mode normal 2>&1)"
if echo "$OUTPUT" | grep -q "SUPERVISOR_HANDOVER_COMPLETE:"; then
    pass "completion notification contains SUPERVISOR_HANDOVER_COMPLETE"
else
    fail "completion should contain SUPERVISOR_HANDOVER_COMPLETE (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "supervisor.*is online"; then
    pass "completion names the incoming supervisor"
else
    fail "completion should name the incoming supervisor (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "45 seconds"; then
    pass "completion includes duration"
else
    fail "completion should include duration (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Normal mode"; then
    pass "completion includes mode"
else
    fail "completion should include mode (got: $OUTPUT)"
fi

# Test AC-5: Emergency mode completion
echo "Test AC-5: Emergency mode completion notification"
HANDOVER_DIR="$TMPDIR_BASE/ac5-emergency"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$NOTIFY_SCRIPT" complete --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --supervisor supervisor --duration 90 --mode emergency 2>&1)"
if echo "$OUTPUT" | grep -q "Emergency mode"; then
    pass "emergency mode identified in completion"
else
    fail "should identify emergency mode (got: $OUTPUT)"
fi

# Test AC-5: Completion notification creates log entry
echo "Test AC-5: Completion notification creates log entry"
if [[ -f "$HANDOVER_DIR/handover.log" ]]; then
    LOG="$(cat "$HANDOVER_DIR/handover.log")"
    if echo "$LOG" | grep -q "Sending completion notification"; then
        pass "completion notification logged"
    else
        fail "completion notification should be logged (got: $LOG)"
    fi
else
    fail "handover.log should exist after completion notification"
fi

# =============================================
# Section 3: shift-status.sh tests
# =============================================

echo ""
echo "=== shift-status.sh ==="

# Test: --help exits successfully
echo "Test: --help exits successfully"
if "$STATUS_SCRIPT" --help >/dev/null 2>&1; then
    pass "status --help exits 0"
else
    fail "status --help should exit 0"
fi

# Test: --help shows usage
echo "Test: --help shows usage"
HELP_OUTPUT="$("$STATUS_SCRIPT" --help 2>&1)"
if echo "$HELP_OUTPUT" | grep -q "Usage"; then
    pass "status --help shows usage"
else
    fail "status --help should show usage"
fi

# Test: unknown option fails
echo "Test: unknown option fails"
if "$STATUS_SCRIPT" --bogus 2>/dev/null; then
    fail "unknown option should fail"
else
    pass "status unknown option exits non-zero"
fi

# Test AC-6: Status with transcript shows zone
echo "Test AC-6: Status shows zone classification"
FAKE_HOME="$TMPDIR_BASE/status-zone"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 2 15)"
HANDOVER_DIR="$TMPDIR_BASE/status-handover"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" HOME="$FAKE_HOME" "$STATUS_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "Zone:.*GREEN"; then
    pass "status shows GREEN zone"
else
    fail "status should show GREEN zone for 2 compressions (got: $OUTPUT)"
fi

# Test AC-6: Status shows JSONL size
echo "Test AC-6: Status shows JSONL size"
if echo "$OUTPUT" | grep -q "JSONL Size:"; then
    pass "status shows JSONL size"
else
    fail "status should show JSONL size (got: $OUTPUT)"
fi

# Test AC-6: Status shows compression count
echo "Test AC-6: Status shows compression count"
if echo "$OUTPUT" | grep -q "Compressions:.*2"; then
    pass "status shows compression count"
else
    fail "status should show compression count (got: $OUTPUT)"
fi

# Test AC-6: Status shows session duration
echo "Test AC-6: Status shows session duration"
if echo "$OUTPUT" | grep -q "Session Duration:"; then
    pass "status shows session duration"
else
    fail "status should show session duration (got: $OUTPUT)"
fi

# Test AC-6: Status shows assistant message count
echo "Test AC-6: Status shows assistant messages"
if echo "$OUTPUT" | grep -q "Assistant Messages:.*15"; then
    pass "status shows assistant message count"
else
    fail "status should show assistant message count (got: $OUTPUT)"
fi

# Test AC-6: Status shows last handover time
echo "Test AC-6: Status shows last handover"
if echo "$OUTPUT" | grep -q "Last Handover:.*none"; then
    pass "status shows no previous handover"
else
    fail "status should show last handover (got: $OUTPUT)"
fi

# Test AC-6: Status with last handover metadata
echo "Test AC-6: Status shows last handover metadata"
HANDOVER_DIR="$TMPDIR_BASE/status-with-metadata"
mkdir -p "$HANDOVER_DIR"
echo "$(( $(date +%s) - 1800 ))" > "$HANDOVER_DIR/last-handover-timestamp"
cat > "$HANDOVER_DIR/last-handover-metadata.yaml" << 'EOF'
timestamp: "2026-03-12T10:00:00Z"
duration_seconds: 45
delta_received: true
ready_confirmed: true
anomalies: "none"
repo: "ThreeDoors"
EOF
FAKE_HOME="$TMPDIR_BASE/status-meta"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 0 5)"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" HOME="$FAKE_HOME" "$STATUS_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "Last Handover:.*ago"; then
    pass "status shows relative handover time"
else
    fail "status should show relative handover time (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Last Handover Details"; then
    pass "status shows handover details section"
else
    fail "status should show handover details (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "duration_seconds: 45"; then
    pass "status shows handover duration"
else
    fail "status should show handover duration (got: $OUTPUT)"
fi

# Test AC-6: Status with yellow zone transcript
echo "Test AC-6: Status shows YELLOW zone"
FAKE_HOME="$TMPDIR_BASE/status-yellow"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 4 20)"
HANDOVER_DIR="$TMPDIR_BASE/status-yellow-handover"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" HOME="$FAKE_HOME" "$STATUS_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "Zone:.*YELLOW"; then
    pass "status shows YELLOW zone"
else
    fail "status should show YELLOW zone for 4 compressions (got: $OUTPUT)"
fi

# Test AC-6: Status with red zone transcript
echo "Test AC-6: Status shows RED zone"
FAKE_HOME="$TMPDIR_BASE/status-red"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 7 30)"
HANDOVER_DIR="$TMPDIR_BASE/status-red-handover"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" HOME="$FAKE_HOME" "$STATUS_SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "Zone:.*RED"; then
    pass "status shows RED zone"
else
    fail "status should show RED zone for 7 compressions (got: $OUTPUT)"
fi

# Test AC-6: Status with no transcript
echo "Test AC-6: Status handles missing transcript gracefully"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$STATUS_SCRIPT" --repo ThreeDoors --handover-dir "$TMPDIR_BASE/no-transcript" --transcript "/nonexistent/path.jsonl" 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "status exits 0 even without transcript"
else
    fail "status should exit 0 without transcript (exit: $EXIT_CODE)"
fi
if echo "$OUTPUT" | grep -q "not found"; then
    pass "status indicates transcript not found"
else
    fail "status should indicate transcript not found (got: $OUTPUT)"
fi

# Test AC-6: JSON output mode
echo "Test AC-6: JSON output mode"
FAKE_HOME="$TMPDIR_BASE/status-json"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 2 10)"
HANDOVER_DIR="$TMPDIR_BASE/status-json-handover"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" HOME="$FAKE_HOME" "$STATUS_SCRIPT" --json --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q '"zone": "GREEN"'; then
    pass "JSON output contains zone"
else
    fail "JSON should contain zone (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q '"compression_count": 2'; then
    pass "JSON output contains compression_count"
else
    fail "JSON should contain compression_count (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q '"assistant_messages": 10'; then
    pass "JSON output contains assistant_messages"
else
    fail "JSON should contain assistant_messages (got: $OUTPUT)"
fi

# Test AC-6: Compact output mode
echo "Test AC-6: Compact output mode"
FAKE_HOME="$TMPDIR_BASE/status-compact"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 1 5)"
HANDOVER_DIR="$TMPDIR_BASE/status-compact-handover"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" HOME="$FAKE_HOME" "$STATUS_SCRIPT" --compact --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "^Shift:.*zone=GREEN.*compressions=1.*messages=5"; then
    pass "compact output format correct"
else
    fail "compact format should be single line with metrics (got: $OUTPUT)"
fi

# --- Results ---
echo ""
echo "=========================================="
echo "=== Results ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total:  $((PASS + FAIL))"
if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
echo "All tests passed."
