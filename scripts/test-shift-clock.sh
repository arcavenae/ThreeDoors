#!/usr/bin/env bash
set -euo pipefail

# test-shift-clock.sh — Tests for shift-clock.sh (Story 58.1)
# Run from the repository root: ./scripts/test-shift-clock.sh
#
# Tests transcript discovery, metrics collection, threshold evaluation,
# time floor enforcement, anti-oscillation guard, natural seam detection,
# and signal file creation.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/shift-clock.sh"
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

# Create a synthetic JSONL transcript with configurable compression events
create_transcript() {
    local dir="$1"
    local compressions="${2:-0}"
    local assistant_messages="${3:-10}"
    local start_time="${4:-2026-03-12T10:00:00}"

    local project_dir="$dir/.claude/projects/-test-ThreeDoors"
    mkdir -p "$project_dir"
    local transcript="$project_dir/test-session.jsonl"

    # System init line with timestamp
    echo "{\"type\":\"system\",\"subtype\":\"init\",\"timestamp\":\"${start_time}.000Z\",\"sessionId\":\"test-session\"}" > "$transcript"

    # Assistant messages
    for i in $(seq 1 "$assistant_messages"); do
        echo "{\"type\":\"assistant\",\"timestamp\":\"${start_time}.${i}00Z\"}" >> "$transcript"
    done

    # Compression events (compact_boundary markers)
    for i in $(seq 1 "$compressions"); do
        echo "{\"type\":\"system\",\"subtype\":\"compact_boundary\",\"content\":\"Conversation compacted\",\"timestamp\":\"${start_time}.${i}00Z\",\"compactMetadata\":{\"trigger\":\"auto\",\"preTokens\":167000}}" >> "$transcript"
    done

    echo "$transcript"
}

# Create a large transcript (by padding) to exceed size thresholds
create_large_transcript() {
    local dir="$1"
    local target_mb="$2"
    local compressions="${3:-0}"

    local transcript
    transcript="$(create_transcript "$dir" "$compressions" 10)"

    # Pad to target size
    local target_bytes=$((target_mb * 1024 * 1024))
    local current_size
    current_size="$(wc -c < "$transcript" | tr -d ' ')"

    while [[ "$current_size" -lt "$target_bytes" ]]; do
        # Add large padding lines (1KB each)
        printf '{"type":"progress","data":"%s","timestamp":"2026-03-12T10:00:00.000Z"}\n' "$(head -c 900 /dev/urandom | base64 | tr -d '\n')" >> "$transcript"
        current_size="$(wc -c < "$transcript" | tr -d ' ')"
    done

    echo "$transcript"
}

# Create mock commands
create_mock_bins() {
    MOCK_BIN_DIR="$(mktemp -d)"

    # Mock multiclaude — default: no pending messages
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
    for cmd in stat grep date cat wc mv mkdir chmod head base64 tr printf; do
        if command -v "$cmd" &>/dev/null; then
            ln -sf "$(command -v "$cmd")" "$MOCK_BIN_DIR/$cmd"
        fi
    done
}

echo "=== Shift Clock Tests ==="

# --- Structural tests ---

echo "Test: --help exits successfully"
if "$SCRIPT" --help >/dev/null 2>&1; then
    pass "--help exits 0"
else
    fail "--help should exit 0"
fi

echo "Test: --help shows usage"
HELP_OUTPUT="$("$SCRIPT" --help 2>&1)"
if echo "$HELP_OUTPUT" | grep -q "Usage"; then
    pass "--help shows usage"
else
    fail "--help should show usage"
fi

echo "Test: unknown option fails"
if "$SCRIPT" --bogus 2>/dev/null; then
    fail "unknown option should fail"
else
    pass "unknown option exits non-zero"
fi

# --- Functional tests with synthetic transcripts ---

echo ""
echo "=== Functional Tests ==="

setup_tmpdir
create_mock_bins

# Test: Green zone — small transcript, no compressions
echo "Test: Green zone with small transcript"
FAKE_HOME="$TMPDIR_BASE/green"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 0 5)"
HANDOVER_DIR="$TMPDIR_BASE/handover-green"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "zone=GREEN"; then
    pass "green zone detected"
else
    fail "should detect green zone (got: $OUTPUT)"
fi
if [[ ! -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "no handover signal in green zone"
else
    fail "should not write handover signal in green zone"
fi

# Test: Yellow zone — 3 compressions
echo "Test: Yellow zone with 3 compressions"
FAKE_HOME="$TMPDIR_BASE/yellow"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 3 20)"
HANDOVER_DIR="$TMPDIR_BASE/handover-yellow"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --yellow-compressions 3 2>&1)"
if echo "$OUTPUT" | grep -q "zone=YELLOW"; then
    pass "yellow zone detected"
else
    fail "should detect yellow zone (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "ADVISORY"; then
    pass "advisory logged in yellow zone"
else
    fail "should log advisory in yellow zone"
fi
if [[ ! -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "no handover signal in yellow zone"
else
    fail "should not write handover signal in yellow zone"
fi

# Test: Red zone — 6 compressions (with time floor bypass via --min-session-minutes 0)
echo "Test: Red zone with 6 compressions triggers handover"
FAKE_HOME="$TMPDIR_BASE/red"
mkdir -p "$FAKE_HOME"
# Use a timestamp 60 minutes ago to satisfy time floor
old_time="$(date -u -v-60M +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '-60 minutes' +%Y-%m-%dT%H:%M:%S 2>/dev/null)"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 6 30 "$old_time")"
HANDOVER_DIR="$TMPDIR_BASE/handover-red"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --min-session-minutes 0 --min-handover-gap-minutes 0 --seam-idle-seconds 0 2>&1)"
if echo "$OUTPUT" | grep -q "zone=RED"; then
    pass "red zone detected"
else
    fail "should detect red zone (got: $OUTPUT)"
fi
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "handover signal written in red zone"
else
    fail "should write handover signal in red zone"
fi

# Test: Signal file content
echo "Test: Signal file contains expected fields"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    SIGNAL_CONTENT="$(cat "$HANDOVER_DIR/handover-requested")"
    if echo "$SIGNAL_CONTENT" | grep -q "timestamp:" && \
       echo "$SIGNAL_CONTENT" | grep -q "zone: \"RED\"" && \
       echo "$SIGNAL_CONTENT" | grep -q "metrics:" && \
       echo "$SIGNAL_CONTENT" | grep -q "file_size_bytes:" && \
       echo "$SIGNAL_CONTENT" | grep -q "compression_count: 6" && \
       echo "$SIGNAL_CONTENT" | grep -q "trigger: \"shift-clock\""; then
        pass "signal file contains all required fields"
    else
        fail "signal file missing required fields (content: $SIGNAL_CONTENT)"
    fi
else
    fail "signal file not found for content check"
fi

# Test: Last handover timestamp recorded
echo "Test: Last handover timestamp recorded"
if [[ -f "$HANDOVER_DIR/last-handover-timestamp" ]]; then
    ts="$(cat "$HANDOVER_DIR/last-handover-timestamp")"
    if [[ "$ts" =~ ^[0-9]+$ ]] && [[ "$ts" -gt 0 ]]; then
        pass "last handover timestamp is valid epoch"
    else
        fail "last handover timestamp should be a positive integer (got: $ts)"
    fi
else
    fail "last handover timestamp file should exist"
fi

# Test: Red zone by file size (>10MB)
echo "Test: Red zone by file size threshold"
FAKE_HOME="$TMPDIR_BASE/red-size"
mkdir -p "$FAKE_HOME"
old_time="$(date -u -v-60M +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '-60 minutes' +%Y-%m-%dT%H:%M:%S 2>/dev/null)"
transcript="$(HOME="$FAKE_HOME" create_large_transcript "$FAKE_HOME" 11 0)"
# Fix the start timestamp in the large transcript
sed -i.bak "1s/2026-03-12T10:00:00/$old_time/" "$transcript" 2>/dev/null || true
HANDOVER_DIR="$TMPDIR_BASE/handover-red-size"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --min-session-minutes 0 --min-handover-gap-minutes 0 --seam-idle-seconds 0 2>&1)"
if echo "$OUTPUT" | grep -q "zone=RED"; then
    pass "red zone detected by file size"
else
    fail "should detect red zone by file size (got: $OUTPUT)"
fi

# Test: Yellow zone by file size (>5MB, <10MB)
echo "Test: Yellow zone by file size threshold"
FAKE_HOME="$TMPDIR_BASE/yellow-size"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_large_transcript "$FAKE_HOME" 6 0)"
HANDOVER_DIR="$TMPDIR_BASE/handover-yellow-size"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "zone=YELLOW"; then
    pass "yellow zone detected by file size"
else
    fail "should detect yellow zone by file size (got: $OUTPUT)"
fi

# Test: Time floor enforcement — session too young
echo "Test: Time floor blocks handover for young session"
FAKE_HOME="$TMPDIR_BASE/time-floor"
mkdir -p "$FAKE_HOME"
# Use current time — session just started
now_time="$(date -u +%Y-%m-%dT%H:%M:%S)"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 8 30 "$now_time")"
HANDOVER_DIR="$TMPDIR_BASE/handover-time-floor"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --min-session-minutes 30 --min-handover-gap-minutes 0 --seam-idle-seconds 0 2>&1)"
if echo "$OUTPUT" | grep -q "zone=RED"; then
    pass "red zone detected despite time floor"
else
    fail "should still detect red zone (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "DEFERRED.*time floor"; then
    pass "handover deferred by time floor"
else
    fail "should defer handover due to time floor (got: $OUTPUT)"
fi
if [[ ! -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "no handover signal when time floor blocks"
else
    fail "should not write handover signal when time floor blocks"
fi

# Test: Anti-oscillation guard — recent handover
echo "Test: Anti-oscillation guard blocks rapid handover"
FAKE_HOME="$TMPDIR_BASE/anti-osc"
mkdir -p "$FAKE_HOME"
old_time="$(date -u -v-60M +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '-60 minutes' +%Y-%m-%dT%H:%M:%S 2>/dev/null)"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 8 30 "$old_time")"
HANDOVER_DIR="$TMPDIR_BASE/handover-anti-osc"
mkdir -p "$HANDOVER_DIR"
# Write a recent handover timestamp (5 minutes ago)
echo "$(( $(date +%s) - 300 ))" > "$HANDOVER_DIR/last-handover-timestamp"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --min-session-minutes 0 --min-handover-gap-minutes 30 --seam-idle-seconds 0 2>&1)"
if echo "$OUTPUT" | grep -q "DEFERRED.*anti-oscillation"; then
    pass "handover deferred by anti-oscillation"
else
    fail "should defer handover due to anti-oscillation (got: $OUTPUT)"
fi
if [[ ! -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "no handover signal during anti-oscillation"
else
    fail "should not write handover signal during anti-oscillation"
fi

# Test: Anti-oscillation allows handover after gap
echo "Test: Anti-oscillation allows handover after sufficient gap"
FAKE_HOME="$TMPDIR_BASE/anti-osc-ok"
mkdir -p "$FAKE_HOME"
old_time="$(date -u -v-60M +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '-60 minutes' +%Y-%m-%dT%H:%M:%S 2>/dev/null)"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 8 30 "$old_time")"
HANDOVER_DIR="$TMPDIR_BASE/handover-anti-osc-ok"
mkdir -p "$HANDOVER_DIR"
# Write an old handover timestamp (60 minutes ago)
echo "$(( $(date +%s) - 3600 ))" > "$HANDOVER_DIR/last-handover-timestamp"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --min-session-minutes 0 --min-handover-gap-minutes 30 --seam-idle-seconds 0 2>&1)"
if [[ -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "handover allowed after anti-oscillation gap"
else
    fail "should allow handover after sufficient gap (got: $OUTPUT)"
fi

# Test: Natural seam detection — pending messages block handover
echo "Test: Pending messages block natural seam"
# Create mock with pending messages
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
            echo "MSG-001  supervisor  Need clarification on AC scope"
        fi
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/multiclaude"

FAKE_HOME="$TMPDIR_BASE/seam-msg"
mkdir -p "$FAKE_HOME"
old_time="$(date -u -v-60M +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '-60 minutes' +%Y-%m-%dT%H:%M:%S 2>/dev/null)"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 8 30 "$old_time")"
HANDOVER_DIR="$TMPDIR_BASE/handover-seam-msg"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --min-session-minutes 0 --min-handover-gap-minutes 0 2>&1)"
if echo "$OUTPUT" | grep -q "DEFERRED.*natural seam"; then
    pass "handover deferred by pending messages"
else
    fail "should defer handover due to pending messages (got: $OUTPUT)"
fi

# Restore clean mock
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
        fi
        ;;
esac
MOCK_EOF
chmod +x "$MOCK_BIN_DIR/multiclaude"

# Test: Transcript not found exits gracefully
echo "Test: Missing transcript exits gracefully"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$TMPDIR_BASE/no-transcript" --transcript "/nonexistent/path.jsonl" 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "graceful exit on missing transcript"
else
    fail "should exit gracefully on missing transcript (exit: $EXIT_CODE)"
fi

# Test: Configurable thresholds
echo "Test: Custom thresholds via CLI flags"
FAKE_HOME="$TMPDIR_BASE/custom"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 2 10)"
HANDOVER_DIR="$TMPDIR_BASE/handover-custom"
# With default thresholds, 2 compressions = GREEN. With --yellow-compressions 2, should be YELLOW.
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" --yellow-compressions 2 2>&1)"
if echo "$OUTPUT" | grep -q "zone=YELLOW"; then
    pass "custom yellow threshold respected"
else
    fail "should respect custom yellow threshold (got: $OUTPUT)"
fi

# Test: Metrics reporting includes all three metrics
echo "Test: Output includes all three metrics"
FAKE_HOME="$TMPDIR_BASE/metrics"
mkdir -p "$FAKE_HOME"
transcript="$(HOME="$FAKE_HOME" create_transcript "$FAKE_HOME" 2 15)"
HANDOVER_DIR="$TMPDIR_BASE/handover-metrics"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --transcript "$transcript" 2>&1)"
if echo "$OUTPUT" | grep -q "compressions=2" && \
   echo "$OUTPUT" | grep -q "messages=15" && \
   echo "$OUTPUT" | grep -q "file_size="; then
    pass "all three metrics reported"
else
    fail "should report all three metrics (got: $OUTPUT)"
fi

# --- Results ---
echo ""
echo "=== Results ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total:  $((PASS + FAIL))"
if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
echo "All tests passed."
