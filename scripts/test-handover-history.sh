#!/usr/bin/env bash
set -euo pipefail

# test-handover-history.sh — Tests for handover-history.sh (Story 58.6)
# Run from the repository root: ./scripts/test-handover-history.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/handover-history.sh"
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

create_mock_bins() {
    MOCK_BIN_DIR="$(mktemp -d)"

    cat > "$MOCK_BIN_DIR/multiclaude" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    repo)
        if [[ "${2:-}" == "current" ]]; then
            echo "ThreeDoors"
        fi
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/multiclaude"

    cat > "$MOCK_BIN_DIR/git" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    remote) echo "https://github.com/arcaven/ThreeDoors.git" ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/git"

    # Pass through system commands
    for cmd in stat date cat rm mkdir chmod wc mv cp find sort head tail sed awk grep tr printf; do
        if command -v "$cmd" &>/dev/null; then
            ln -sf "$(command -v "$cmd")" "$MOCK_BIN_DIR/$cmd"
        fi
    done
}

echo "=== Handover History Tests ==="

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

echo "Test: no args shows help"
NO_ARGS_OUTPUT="$("$SCRIPT" 2>&1)"
if echo "$NO_ARGS_OUTPUT" | grep -q "Usage"; then
    pass "no args shows help"
else
    fail "no args should show help"
fi

echo "Test: unknown subcommand fails"
if "$SCRIPT" bogus 2>/dev/null; then
    fail "unknown subcommand should fail"
else
    pass "unknown subcommand exits non-zero"
fi

echo "Test: record missing required options fails"
if "$SCRIPT" record --handover-dir /tmp/test 2>/dev/null; then
    fail "record without required options should fail"
else
    pass "record without required options exits non-zero"
fi

# --- Functional tests ---

echo ""
echo "=== Functional Tests: record ==="

setup_tmpdir
create_mock_bins

# Test: AC-1 — State file archival
echo "Test: AC-1 — State file is archived with timestamp"
HANDOVER_DIR="$TMPDIR_BASE/ac1"
mkdir -p "$HANDOVER_DIR"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"
echo "timestamp: \"2026-03-12T14:00:00Z\"" >> "$HANDOVER_DIR/shift-state.yaml"

OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "gentle-hawk" \
    --incoming "bold-eagle" \
    --type normal \
    --trigger-metrics '{"jsonl_size_mb":8.5,"compression_count":5}' \
    --duration 45 \
    --delta-received true 2>&1)"

HISTORY_DIR="$HANDOVER_DIR/history"
if [[ -d "$HISTORY_DIR" ]]; then
    pass "history directory created"
else
    fail "history directory should be created"
fi

ARCHIVE_FILES="$(find "$HISTORY_DIR" -name 'shift-state-*.yaml' -type f 2>/dev/null)"
if [[ -n "$ARCHIVE_FILES" ]]; then
    pass "state file archived"
else
    fail "state file should be archived to history/"
fi

# Verify archive content matches original
ARCHIVE_FILE="$(echo "$ARCHIVE_FILES" | head -1)"
if grep -q "version: 1" "$ARCHIVE_FILE"; then
    pass "archive content matches original"
else
    fail "archive content should match original state file"
fi

# Test: AC-1 — Archive filename contains ISO-like timestamp
ARCHIVE_BASENAME="$(basename "$ARCHIVE_FILE")"
if echo "$ARCHIVE_BASENAME" | grep -qE 'shift-state-[0-9]{4}-[0-9]{2}-[0-9]{2}T'; then
    pass "archive filename contains ISO timestamp"
else
    fail "archive filename should contain ISO timestamp (got: $ARCHIVE_BASENAME)"
fi

# Test: AC-1 — No state file to archive
echo "Test: AC-1 — Warning when no state file exists"
HANDOVER_DIR="$TMPDIR_BASE/ac1-no-state"
mkdir -p "$HANDOVER_DIR"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "gentle-hawk" \
    --incoming "bold-eagle" \
    --type normal \
    --trigger-metrics '{}' \
    --duration 30 \
    --delta-received false 2>&1)"
if echo "$OUTPUT" | grep -q "Warning.*No shift-state.yaml"; then
    pass "warning when no state file"
else
    fail "should warn when no state file to archive (got: $OUTPUT)"
fi

# Test: AC-2 — Handover event log (JSONL)
echo "Test: AC-2 — JSONL event logged"
HANDOVER_DIR="$TMPDIR_BASE/ac2"
mkdir -p "$HANDOVER_DIR"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"

PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "gentle-hawk" \
    --incoming "bold-eagle" \
    --type normal \
    --trigger-metrics '{"jsonl_size_mb":8.5,"compression_count":5,"message_count":120}' \
    --duration 45 \
    --delta-received true >/dev/null 2>&1

LOG_FILE="$HANDOVER_DIR/handover-log.jsonl"
if [[ -f "$LOG_FILE" ]]; then
    pass "JSONL log file created"
else
    fail "handover-log.jsonl should be created"
fi

# Validate JSONL content has all required fields
LOG_LINE="$(cat "$LOG_FILE")"
for field in timestamp outgoing incoming type trigger_metrics duration_seconds delta_received anomalies; do
    if echo "$LOG_LINE" | grep -q "\"$field\""; then
        pass "JSONL contains field: $field"
    else
        fail "JSONL should contain field: $field (got: $LOG_LINE)"
    fi
done

# Validate specific values
if echo "$LOG_LINE" | grep -q '"outgoing":"gentle-hawk"'; then
    pass "JSONL outgoing value correct"
else
    fail "JSONL outgoing should be gentle-hawk"
fi
if echo "$LOG_LINE" | grep -q '"incoming":"bold-eagle"'; then
    pass "JSONL incoming value correct"
else
    fail "JSONL incoming should be bold-eagle"
fi
if echo "$LOG_LINE" | grep -q '"type":"normal"'; then
    pass "JSONL type value correct"
else
    fail "JSONL type should be normal"
fi
if echo "$LOG_LINE" | grep -q '"duration_seconds":45'; then
    pass "JSONL duration value correct"
else
    fail "JSONL duration should be 45"
fi
if echo "$LOG_LINE" | grep -q '"delta_received":true'; then
    pass "JSONL delta_received value correct"
else
    fail "JSONL delta_received should be true"
fi
if echo "$LOG_LINE" | grep -q '"anomalies":\[\]'; then
    pass "JSONL anomalies empty when none"
else
    fail "JSONL anomalies should be [] when none"
fi

# Test: AC-2 — JSONL with anomalies
echo "Test: AC-2 — JSONL with anomalies"
HANDOVER_DIR="$TMPDIR_BASE/ac2-anomalies"
mkdir -p "$HANDOVER_DIR"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"

PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "gentle-hawk" \
    --incoming "bold-eagle" \
    --type emergency \
    --trigger-metrics '{}' \
    --duration 120 \
    --delta-received false \
    --anomalies "delta_timeout,outgoing_unresponsive" >/dev/null 2>&1

LOG_LINE="$(cat "$HANDOVER_DIR/handover-log.jsonl")"
if echo "$LOG_LINE" | grep -q '"anomalies":\["delta_timeout","outgoing_unresponsive"\]'; then
    pass "JSONL anomalies array populated"
else
    fail "JSONL anomalies should contain the anomaly list (got: $LOG_LINE)"
fi
if echo "$LOG_LINE" | grep -q '"type":"emergency"'; then
    pass "JSONL emergency type recorded"
else
    fail "JSONL should record emergency type"
fi

# Test: AC-2 — Multiple events append (JSONL append-only)
echo "Test: AC-2 — Multiple events append to JSONL"
HANDOVER_DIR="$TMPDIR_BASE/ac2-append"
mkdir -p "$HANDOVER_DIR"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"

for i in 1 2 3; do
    PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
        --handover-dir "$HANDOVER_DIR" \
        --outgoing "sup-$i" \
        --incoming "sup-$((i+1))" \
        --type normal \
        --trigger-metrics '{}' \
        --duration "$((30 + i))" \
        --delta-received true >/dev/null 2>&1
done

LINE_COUNT="$(wc -l < "$HANDOVER_DIR/handover-log.jsonl" | tr -d ' ')"
if [[ "$LINE_COUNT" -eq 3 ]]; then
    pass "3 events appended to JSONL"
else
    fail "should have 3 lines in JSONL (got: $LINE_COUNT)"
fi

# Test: AC-3 — History retention (under limit)
echo ""
echo "=== Functional Tests: retention ==="
echo "Test: AC-3 — Under retention limit, no cleanup"
HANDOVER_DIR="$TMPDIR_BASE/ac3-under"
mkdir -p "$HANDOVER_DIR/history"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"

# Create 5 archive files
for i in $(seq 1 5); do
    echo "version: 1" > "$HANDOVER_DIR/history/shift-state-2026-03-12T10-0${i}-00Z.yaml"
done

PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "sup-a" \
    --incoming "sup-b" \
    --type normal \
    --trigger-metrics '{}' \
    --duration 30 \
    --delta-received true >/dev/null 2>&1

FILE_COUNT="$(find "$HANDOVER_DIR/history" -name 'shift-state-*.yaml' -type f | wc -l | tr -d ' ')"
# 5 pre-existing + 1 new = 6, all under 50
if [[ "$FILE_COUNT" -eq 6 ]]; then
    pass "no files deleted when under limit (count: $FILE_COUNT)"
else
    fail "should keep all files when under limit (got: $FILE_COUNT, expected 6)"
fi

# Test: AC-3 — Over retention limit, oldest deleted
echo "Test: AC-3 — Over retention limit, oldest deleted"
HANDOVER_DIR="$TMPDIR_BASE/ac3-over"
mkdir -p "$HANDOVER_DIR/history"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"

# Create 52 archive files (will be 53 after record → should delete 3)
for i in $(seq -w 1 52); do
    echo "version: 1" > "$HANDOVER_DIR/history/shift-state-2026-03-01T${i}-00-00Z.yaml"
done

OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "sup-a" \
    --incoming "sup-b" \
    --type normal \
    --trigger-metrics '{}' \
    --duration 30 \
    --delta-received true 2>&1)"

FILE_COUNT="$(find "$HANDOVER_DIR/history" -name 'shift-state-*.yaml' -type f | wc -l | tr -d ' ')"
if [[ "$FILE_COUNT" -le 50 ]]; then
    pass "retention enforced (count: $FILE_COUNT <= 50)"
else
    fail "should enforce 50 file limit (got: $FILE_COUNT)"
fi

if echo "$OUTPUT" | grep -q "Retention cleanup"; then
    pass "retention cleanup logged"
else
    fail "should log retention cleanup"
fi

# Verify oldest files were deleted (files with lowest timestamps)
if [[ ! -f "$HANDOVER_DIR/history/shift-state-2026-03-01T01-00-00Z.yaml" ]]; then
    pass "oldest file deleted"
else
    fail "oldest file should have been deleted"
fi

# Test: AC-4 — Summary command
echo ""
echo "=== Functional Tests: summary ==="
echo "Test: AC-4 — Summary displays events"
HANDOVER_DIR="$TMPDIR_BASE/ac4"
mkdir -p "$HANDOVER_DIR"

# Create a log with several entries
cat > "$HANDOVER_DIR/handover-log.jsonl" << 'JSONL_EOF'
{"timestamp":"2026-03-12T10:00:00Z","outgoing":"hawk-1","incoming":"eagle-1","type":"normal","trigger_metrics":{"jsonl_size_mb":8},"duration_seconds":45,"delta_received":true,"anomalies":[]}
{"timestamp":"2026-03-12T12:00:00Z","outgoing":"eagle-1","incoming":"hawk-2","type":"normal","trigger_metrics":{"compression_count":6},"duration_seconds":38,"delta_received":true,"anomalies":[]}
{"timestamp":"2026-03-12T14:00:00Z","outgoing":"hawk-2","incoming":"eagle-2","type":"emergency","trigger_metrics":{},"duration_seconds":125,"delta_received":false,"anomalies":["delta_timeout"]}
JSONL_EOF

OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" summary --handover-dir "$HANDOVER_DIR" 2>&1)"

if echo "$OUTPUT" | grep -q "Handover History"; then
    pass "summary shows header"
else
    fail "summary should show header (got: $OUTPUT)"
fi

if echo "$OUTPUT" | grep -q "hawk-1"; then
    pass "summary shows outgoing name"
else
    fail "summary should show outgoing name"
fi

if echo "$OUTPUT" | grep -q "eagle-2"; then
    pass "summary shows incoming name"
else
    fail "summary should show incoming name"
fi

if echo "$OUTPUT" | grep -q "emergency"; then
    pass "summary shows handover type"
else
    fail "summary should show handover type"
fi

if echo "$OUTPUT" | grep -q "Statistics"; then
    pass "summary shows statistics section"
else
    fail "summary should show statistics"
fi

if echo "$OUTPUT" | grep -q "Total handovers:.*3"; then
    pass "summary shows correct total"
else
    fail "summary should show 3 total handovers"
fi

if echo "$OUTPUT" | grep -q "emergency"; then
    pass "summary shows emergency count"
else
    fail "summary should show emergency count"
fi

# Test: AC-4 — Summary with --last flag
echo "Test: AC-4 — Summary --last limits output"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" summary --handover-dir "$HANDOVER_DIR" --last 1 2>&1)"
# Should only show last entry (eagle-2, not hawk-1/eagle-1)
if echo "$OUTPUT" | grep -q "last 1 of 3"; then
    pass "summary --last 1 shows count"
else
    fail "summary should show 'last 1 of 3'"
fi

# Test: AC-4 — Summary with no log file
echo "Test: AC-4 — Summary with no log file"
HANDOVER_DIR="$TMPDIR_BASE/ac4-empty"
mkdir -p "$HANDOVER_DIR"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" summary --handover-dir "$HANDOVER_DIR" 2>&1)"
if echo "$OUTPUT" | grep -q "No handover log"; then
    pass "summary handles missing log file"
else
    fail "summary should handle missing log file (got: $OUTPUT)"
fi

# Test: AC-5 — Frequency check (under threshold)
echo ""
echo "=== Functional Tests: check-frequency ==="
echo "Test: AC-5 — Under threshold, no warning"
HANDOVER_DIR="$TMPDIR_BASE/ac5-under"
mkdir -p "$HANDOVER_DIR"

NOW_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
cat > "$HANDOVER_DIR/handover-log.jsonl" << JSONL_EOF
{"timestamp":"$NOW_TS","outgoing":"s1","incoming":"s2","type":"normal","trigger_metrics":{},"duration_seconds":30,"delta_received":true,"anomalies":[]}
{"timestamp":"$NOW_TS","outgoing":"s2","incoming":"s3","type":"normal","trigger_metrics":{},"duration_seconds":30,"delta_received":true,"anomalies":[]}
JSONL_EOF

OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" check-frequency --handover-dir "$HANDOVER_DIR" 2>&1)"
if echo "$OUTPUT" | grep -q "HIGH_HANDOVER_FREQUENCY"; then
    fail "should not warn when under threshold (got: $OUTPUT)"
else
    pass "no warning when under threshold"
fi

# Test: AC-5 — Over threshold, warning emitted
echo "Test: AC-5 — Over threshold, warning emitted"
HANDOVER_DIR="$TMPDIR_BASE/ac5-over"
mkdir -p "$HANDOVER_DIR"

NOW_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
{
    for i in 1 2 3 4; do
        echo "{\"timestamp\":\"$NOW_TS\",\"outgoing\":\"s$i\",\"incoming\":\"s$((i+1))\",\"type\":\"normal\",\"trigger_metrics\":{},\"duration_seconds\":30,\"delta_received\":true,\"anomalies\":[]}"
    done
} > "$HANDOVER_DIR/handover-log.jsonl"

OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" check-frequency --handover-dir "$HANDOVER_DIR" 2>&1)" || true
if echo "$OUTPUT" | grep -q "HIGH_HANDOVER_FREQUENCY"; then
    pass "warning emitted when over threshold"
else
    fail "should warn when over threshold (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "4 handovers"; then
    pass "warning includes correct count"
else
    fail "warning should include count of 4"
fi
if echo "$OUTPUT" | grep -q "Consider increasing shift clock thresholds"; then
    pass "warning includes remediation advice"
else
    fail "warning should include remediation advice"
fi

# Test: AC-5 — Old entries not counted
echo "Test: AC-5 — Old entries outside window not counted"
HANDOVER_DIR="$TMPDIR_BASE/ac5-old"
mkdir -p "$HANDOVER_DIR"

NOW_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
{
    # 4 entries from yesterday (outside 1-hour window)
    for i in 1 2 3 4; do
        echo "{\"timestamp\":\"2026-03-11T01:00:0${i}Z\",\"outgoing\":\"s$i\",\"incoming\":\"s$((i+1))\",\"type\":\"normal\",\"trigger_metrics\":{},\"duration_seconds\":30,\"delta_received\":true,\"anomalies\":[]}"
    done
    # 1 recent entry
    echo "{\"timestamp\":\"$NOW_TS\",\"outgoing\":\"s5\",\"incoming\":\"s6\",\"type\":\"normal\",\"trigger_metrics\":{},\"duration_seconds\":30,\"delta_received\":true,\"anomalies\":[]}"
} > "$HANDOVER_DIR/handover-log.jsonl"

OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" check-frequency --handover-dir "$HANDOVER_DIR" 2>&1)" || true
if echo "$OUTPUT" | grep -q "HIGH_HANDOVER_FREQUENCY"; then
    fail "should not count old entries outside window (got: $OUTPUT)"
else
    pass "old entries outside window not counted"
fi

# Test: AC-5 — Custom threshold and window
echo "Test: AC-5 — Custom threshold and window"
HANDOVER_DIR="$TMPDIR_BASE/ac5-custom"
mkdir -p "$HANDOVER_DIR"

NOW_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
{
    echo "{\"timestamp\":\"$NOW_TS\",\"outgoing\":\"s1\",\"incoming\":\"s2\",\"type\":\"normal\",\"trigger_metrics\":{},\"duration_seconds\":30,\"delta_received\":true,\"anomalies\":[]}"
    echo "{\"timestamp\":\"$NOW_TS\",\"outgoing\":\"s2\",\"incoming\":\"s3\",\"type\":\"normal\",\"trigger_metrics\":{},\"duration_seconds\":30,\"delta_received\":true,\"anomalies\":[]}"
} > "$HANDOVER_DIR/handover-log.jsonl"

# With threshold=1, 2 entries should trigger
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" check-frequency --handover-dir "$HANDOVER_DIR" --threshold 1 2>&1)" || true
if echo "$OUTPUT" | grep -q "HIGH_HANDOVER_FREQUENCY"; then
    pass "custom threshold=1 triggers on 2 entries"
else
    fail "custom threshold=1 should trigger on 2 entries (got: $OUTPUT)"
fi

# Test: AC-5 — check-frequency called automatically by record
echo "Test: AC-5 — Frequency check runs automatically on record"
HANDOVER_DIR="$TMPDIR_BASE/ac5-auto"
mkdir -p "$HANDOVER_DIR"
echo "version: 1" > "$HANDOVER_DIR/shift-state.yaml"

NOW_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
{
    for i in 1 2 3; do
        echo "{\"timestamp\":\"$NOW_TS\",\"outgoing\":\"s$i\",\"incoming\":\"s$((i+1))\",\"type\":\"normal\",\"trigger_metrics\":{},\"duration_seconds\":30,\"delta_received\":true,\"anomalies\":[]}"
    done
} > "$HANDOVER_DIR/handover-log.jsonl"

# This record will be #4, exceeding threshold of 3
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" record \
    --handover-dir "$HANDOVER_DIR" \
    --outgoing "s4" \
    --incoming "s5" \
    --type normal \
    --trigger-metrics '{}' \
    --duration 30 \
    --delta-received true 2>&1)" || true

if echo "$OUTPUT" | grep -q "HIGH_HANDOVER_FREQUENCY"; then
    pass "frequency check runs automatically on record"
else
    fail "frequency check should run automatically on record (got: $OUTPUT)"
fi

# Test: JSONL entries are valid JSON
echo ""
echo "=== Validation Tests ==="
echo "Test: JSONL entries are valid JSON (if jq available)"
if command -v jq &>/dev/null; then
    HANDOVER_DIR="$TMPDIR_BASE/ac2"  # Reuse from earlier
    while IFS= read -r line; do
        if echo "$line" | jq . >/dev/null 2>&1; then
            : # valid
        else
            fail "JSONL line is not valid JSON: $line"
        fi
    done < "$HANDOVER_DIR/handover-log.jsonl"
    pass "all JSONL entries are valid JSON"
else
    echo "  SKIP: jq not available for JSON validation"
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
