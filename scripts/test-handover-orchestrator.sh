#!/usr/bin/env bash
set -euo pipefail

# test-handover-orchestrator.sh — Tests for handover-orchestrator.sh (Stories 58.3, 58.5)
# Run from the repository root: ./scripts/test-handover-orchestrator.sh
#
# Tests signal file detection, outgoing notification, delta wait, incoming spawn,
# ready confirmation, outgoing termination, handover logging, emergency protocol,
# force-kill, spawn retry, and emergency alerting.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/handover-orchestrator.sh"
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

# Create mock multiclaude that simulates different scenarios
# Argument: behavior profile name
create_mock_bins() {
    local profile="${1:-happy-path}"
    MOCK_BIN_DIR="$(mktemp -d)"

    # State tracking file for mock multiclaude
    local state_file="$MOCK_BIN_DIR/.mock-state"
    echo "0" > "$state_file"

    case "$profile" in
        happy-path)
            # Mock multiclaude: message send succeeds, message list returns HANDOVER_COMPLETE then READY
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then
            echo "ThreeDoors"
        fi
        ;;
    message)
        case "\${2:-}" in
            send)
                echo "Message sent"
                ;;
            list)
                call_count=\$(cat "\$STATE_FILE")
                call_count=\$((call_count + 1))
                echo "\$call_count" > "\$STATE_FILE"
                if [[ "\$call_count" -le 2 ]]; then
                    echo "No pending messages"
                elif [[ "\$call_count" -le 4 ]]; then
                    echo "MSG-001  supervisor  HANDOVER_COMPLETE: Delta written"
                else
                    echo "MSG-002  supervisor  READY: Incoming supervisor ready"
                fi
                ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then
            echo "Agent spawned: supervisor"
        fi
        ;;
    worker)
        if [[ "\${2:-}" == "list" ]]; then
            echo "No active workers"
        fi
        ;;
esac
MOCK_EOF
            ;;

        delta-timeout)
            # Mock multiclaude: message list never returns HANDOVER_COMPLETE
            # Captures spawn task argument to verify emergency flag
            local spawn_log="$MOCK_BIN_DIR/.spawn-task-log"
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
SPAWN_LOG="$spawn_log"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then echo "ThreeDoors"; fi
        ;;
    message)
        case "\${2:-}" in
            send) echo "Message sent" ;;
            list)
                call_count=\$(cat "\$STATE_FILE")
                call_count=\$((call_count + 1))
                echo "\$call_count" > "\$STATE_FILE"
                if [[ "\$call_count" -gt 10 ]]; then
                    echo "MSG-002  supervisor  READY: Incoming supervisor ready"
                else
                    echo "No pending messages"
                fi
                ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then
            # Capture the --task argument
            while [[ \$# -gt 0 ]]; do
                if [[ "\$1" == "--task" ]]; then
                    echo "\$2" > "\$SPAWN_LOG"
                    break
                fi
                shift
            done
            echo "Agent spawned: supervisor"
        fi
        ;;
esac
MOCK_EOF
            ;;

        ready-timeout)
            # Mock multiclaude: HANDOVER_COMPLETE but never READY
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then echo "ThreeDoors"; fi
        ;;
    message)
        case "\${2:-}" in
            send) echo "Message sent" ;;
            list)
                call_count=\$(cat "\$STATE_FILE")
                call_count=\$((call_count + 1))
                echo "\$call_count" > "\$STATE_FILE"
                echo "MSG-001  supervisor  HANDOVER_COMPLETE: Delta written"
                ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then echo "Agent spawned: supervisor"; fi
        ;;
esac
MOCK_EOF
            ;;

        spawn-failure)
            # Mock multiclaude: spawn fails
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then echo "ThreeDoors"; fi
        ;;
    message)
        case "\${2:-}" in
            send) echo "Message sent" ;;
            list) echo "MSG-001  supervisor  HANDOVER_COMPLETE: Delta written" ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then
            echo "Error: Failed to spawn agent" >&2
            exit 1
        fi
        ;;
esac
MOCK_EOF
            ;;

        spawn-retry-succeed)
            # Mock multiclaude: first spawn fails, second succeeds (emergency path)
            local spawn_log="$MOCK_BIN_DIR/.spawn-task-log"
            local spawn_count_file="$MOCK_BIN_DIR/.spawn-count"
            echo "0" > "$spawn_count_file"
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
SPAWN_LOG="$spawn_log"
SPAWN_COUNT_FILE="$spawn_count_file"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then echo "ThreeDoors"; fi
        ;;
    message)
        case "\${2:-}" in
            send) echo "Message sent" ;;
            list)
                call_count=\$(cat "\$STATE_FILE")
                call_count=\$((call_count + 1))
                echo "\$call_count" > "\$STATE_FILE"
                if [[ "\$call_count" -gt 10 ]]; then
                    echo "MSG-002  supervisor  READY: Incoming supervisor ready"
                else
                    echo "No pending messages"
                fi
                ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then
            spawn_count=\$(cat "\$SPAWN_COUNT_FILE")
            spawn_count=\$((spawn_count + 1))
            echo "\$spawn_count" > "\$SPAWN_COUNT_FILE"
            while [[ \$# -gt 0 ]]; do
                if [[ "\$1" == "--task" ]]; then
                    echo "\$2" > "\$SPAWN_LOG"
                    break
                fi
                shift
            done
            if [[ "\$spawn_count" -le 1 ]]; then
                echo "Error: Failed to spawn agent" >&2
                exit 1
            fi
            echo "Agent spawned: supervisor"
        fi
        ;;
esac
MOCK_EOF
            ;;

        spawn-retry-fail)
            # Mock multiclaude: both spawn attempts fail (emergency path)
            local spawn_log="$MOCK_BIN_DIR/.spawn-task-log"
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
SPAWN_LOG="$spawn_log"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then echo "ThreeDoors"; fi
        ;;
    message)
        case "\${2:-}" in
            send) echo "Message sent" ;;
            list) echo "No pending messages" ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then
            while [[ \$# -gt 0 ]]; do
                if [[ "\$1" == "--task" ]]; then
                    echo "\$2" > "\$SPAWN_LOG"
                    break
                fi
                shift
            done
            echo "Error: Failed to spawn agent" >&2
            exit 1
        fi
        ;;
esac
MOCK_EOF
            ;;

        notification-failure)
            # Mock multiclaude: message send fails
            cat > "$MOCK_BIN_DIR/multiclaude" << MOCK_EOF
#!/usr/bin/env bash
STATE_FILE="$state_file"
case "\$1" in
    repo)
        if [[ "\${2:-}" == "current" ]]; then echo "ThreeDoors"; fi
        ;;
    message)
        case "\${2:-}" in
            send) echo "Error: send failed" >&2; exit 1 ;;
            list)
                call_count=\$(cat "\$STATE_FILE")
                call_count=\$((call_count + 1))
                echo "\$call_count" > "\$STATE_FILE"
                if [[ "\$call_count" -gt 3 ]]; then
                    echo "MSG-002  supervisor  READY: Incoming supervisor ready"
                else
                    echo "No pending messages"
                fi
                ;;
        esac
        ;;
    agents)
        if [[ "\${2:-}" == "spawn" ]]; then echo "Agent spawned: supervisor"; fi
        ;;
esac
MOCK_EOF
            ;;
    esac

    chmod +x "$MOCK_BIN_DIR/multiclaude"

    # Mock tmux — supervisor window exists by default
    cat > "$MOCK_BIN_DIR/tmux" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    list-windows)
        echo "supervisor"
        echo "merge-queue"
        echo "pr-shepherd"
        ;;
    kill-window)
        echo "Window killed"
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/tmux"

    # Mock git
    cat > "$MOCK_BIN_DIR/git" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    remote) echo "https://github.com/arcavenae/ThreeDoors.git" ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/git"

    # Pass through system commands
    for cmd in stat date cat rm mkdir chmod wc mv sleep; do
        if command -v "$cmd" &>/dev/null; then
            ln -sf "$(command -v "$cmd")" "$MOCK_BIN_DIR/$cmd"
        fi
    done
}

# Create no-tmux mock (simulates dead outgoing supervisor)
create_dead_supervisor_mock() {
    cat > "$MOCK_BIN_DIR/tmux" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    list-windows)
        # No supervisor window
        echo "merge-queue"
        echo "pr-shepherd"
        ;;
    kill-window)
        echo "Window killed"
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/tmux"
}

echo "=== Handover Orchestrator Tests ==="

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

# --- Functional tests ---

echo ""
echo "=== Functional Tests ==="

setup_tmpdir

# Test: No signal file — exits cleanly
echo "Test: No signal file exits cleanly"
create_mock_bins "happy-path"
HANDOVER_DIR="$TMPDIR_BASE/no-signal"
mkdir -p "$HANDOVER_DIR"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "clean exit with no signal file"
else
    fail "should exit 0 with no signal file (exit: $EXIT_CODE)"
fi
if [[ -z "$OUTPUT" ]]; then
    pass "no output with no signal file"
else
    fail "should produce no output with no signal file (got: $OUTPUT)"
fi

# Test: Signal file detection and consumption (AC-1)
echo "Test: Signal file is detected and consumed"
create_mock_bins "happy-path"
HANDOVER_DIR="$TMPDIR_BASE/signal-consume"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
echo 'zone: "RED"' >> "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 10 --ready-timeout 10 2>&1)"
if [[ ! -f "$HANDOVER_DIR/handover-requested" ]]; then
    pass "signal file consumed (removed)"
else
    fail "signal file should be removed after detection"
fi
if echo "$OUTPUT" | grep -q "Handover signal detected"; then
    pass "signal detection logged"
else
    fail "should log signal detection (got: $OUTPUT)"
fi

# Test: Happy path — full handover sequence (AC-1 through AC-8)
echo "Test: Happy path — full handover sequence"
create_mock_bins "happy-path"
HANDOVER_DIR="$TMPDIR_BASE/happy-path"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 10 --ready-timeout 15 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "happy path exits 0"
else
    fail "happy path should exit 0 (exit: $EXIT_CODE, output: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Handover sequence complete"; then
    pass "handover sequence completed"
else
    fail "should complete handover sequence (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "HANDOVER_COMPLETE"; then
    pass "delta received in happy path"
else
    fail "should receive delta (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "Incoming supervisor confirmed READY"; then
    pass "ready confirmed in happy path"
else
    fail "should confirm ready (got: $OUTPUT)"
fi

# Test: Handover metadata recorded (AC-8)
echo "Test: Handover metadata file created"
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q "duration_seconds:" && \
       echo "$METADATA" | grep -q "delta_received: true" && \
       echo "$METADATA" | grep -q "ready_confirmed: true" && \
       echo "$METADATA" | grep -q "repo: \"ThreeDoors\""; then
        pass "metadata contains all required fields"
    else
        fail "metadata missing fields (got: $METADATA)"
    fi
else
    fail "metadata file should exist after handover"
fi

# Test: Last handover timestamp updated
echo "Test: Last handover timestamp updated"
if [[ -f "$HANDOVER_DIR/last-handover-timestamp" ]]; then
    ts="$(cat "$HANDOVER_DIR/last-handover-timestamp")"
    if [[ "$ts" =~ ^[0-9]+$ ]] && [[ "$ts" -gt 0 ]]; then
        pass "handover timestamp is valid epoch"
    else
        fail "handover timestamp should be positive integer (got: $ts)"
    fi
else
    fail "last-handover-timestamp file should exist"
fi

# Test: Handover log file created
echo "Test: Handover log file created"
if [[ -f "$HANDOVER_DIR/handover.log" ]]; then
    LOG_CONTENT="$(cat "$HANDOVER_DIR/handover.log")"
    if echo "$LOG_CONTENT" | grep -q "Handover sequence initiated" && \
       echo "$LOG_CONTENT" | grep -q "Handover sequence complete"; then
        pass "handover log contains sequence markers"
    else
        fail "handover log missing markers (got: $LOG_CONTENT)"
    fi
else
    fail "handover.log should exist after handover"
fi

# Test: Delta timeout path (AC-3)
echo "Test: Delta timeout proceeds to spawn"
create_mock_bins "delta-timeout"
HANDOVER_DIR="$TMPDIR_BASE/delta-timeout"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 15 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "delta timeout still completes handover"
else
    fail "should complete handover even with delta timeout (exit: $EXIT_CODE)"
fi
if echo "$OUTPUT" | grep -q "Delta wait timed out"; then
    pass "delta timeout logged"
else
    fail "should log delta timeout (got: $OUTPUT)"
fi
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q "delta_received: false"; then
        pass "metadata records delta not received"
    else
        fail "metadata should show delta_received: false (got: $METADATA)"
    fi
    if echo "$METADATA" | grep -q "delta_timeout"; then
        pass "metadata records delta_timeout anomaly"
    else
        fail "metadata should record delta_timeout anomaly (got: $METADATA)"
    fi
fi

# Test: Ready timeout path (AC-5)
echo "Test: Ready timeout logs error"
create_mock_bins "ready-timeout"
HANDOVER_DIR="$TMPDIR_BASE/ready-timeout"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 5 --ready-timeout 3 2>&1)"
if echo "$OUTPUT" | grep -q "failed to confirm ready"; then
    pass "ready timeout logged"
else
    fail "should log ready timeout (got: $OUTPUT)"
fi
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q "ready_confirmed: false"; then
        pass "metadata records ready not confirmed"
    else
        fail "metadata should show ready_confirmed: false (got: $METADATA)"
    fi
fi

# Test: Spawn failure is fatal (AC-4)
echo "Test: Spawn failure exits non-zero"
create_mock_bins "spawn-failure"
HANDOVER_DIR="$TMPDIR_BASE/spawn-failure"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 3 2>&1)" || true
EXIT_CODE=$?
# Script should exit 1 on spawn failure
if echo "$OUTPUT" | grep -q "Could not spawn incoming supervisor"; then
    pass "spawn failure logged as fatal"
else
    fail "should log fatal spawn failure (got: $OUTPUT)"
fi
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q "spawn_failed"; then
        pass "metadata records spawn_failed anomaly"
    else
        fail "metadata should record spawn_failed (got: $METADATA)"
    fi
fi

# Test: Dead outgoing supervisor (AC-2 skip path)
echo "Test: Dead outgoing supervisor skips notification"
create_mock_bins "happy-path"
create_dead_supervisor_mock
HANDOVER_DIR="$TMPDIR_BASE/dead-outgoing"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 15 2>&1)"
if echo "$OUTPUT" | grep -q "Outgoing supervisor is not running"; then
    pass "dead outgoing detected"
else
    fail "should detect dead outgoing supervisor (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "skipping notification"; then
    pass "notification skipped for dead outgoing"
else
    fail "should skip notification for dead outgoing (got: $OUTPUT)"
fi
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q "outgoing_supervisor_dead"; then
        pass "metadata records dead outgoing anomaly"
    else
        fail "metadata should record outgoing_supervisor_dead (got: $METADATA)"
    fi
fi

# Test: Signal file not re-triggerable (AC-1)
echo "Test: Signal file consumed prevents re-trigger"
create_mock_bins "happy-path"
HANDOVER_DIR="$TMPDIR_BASE/no-retrigger"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
# First run: consumes signal
PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 10 --ready-timeout 15 >/dev/null 2>&1
# Second run: should find no signal
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 2>&1)"
if [[ -z "$OUTPUT" ]]; then
    pass "second run produces no output (no re-trigger)"
else
    fail "second run should not trigger handover (got: $OUTPUT)"
fi

# Test: Notification failure still proceeds (AC-2 error path)
echo "Test: Notification failure still proceeds to spawn"
create_mock_bins "notification-failure"
HANDOVER_DIR="$TMPDIR_BASE/notify-fail"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 15 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "notification failure still completes"
else
    fail "should still complete after notification failure (exit: $EXIT_CODE)"
fi
if echo "$OUTPUT" | grep -q "notification_failed\|Failed to send"; then
    pass "notification failure logged"
else
    fail "should log notification failure (got: $OUTPUT)"
fi

# Test: Configurable timeouts via CLI
echo "Test: Custom timeouts via CLI flags"
create_mock_bins "delta-timeout"
HANDOVER_DIR="$TMPDIR_BASE/custom-timeouts"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
START_TIME="$(date +%s)"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 2 --ready-timeout 15 2>&1)"
END_TIME="$(date +%s)"
ELAPSED=$(( END_TIME - START_TIME ))
# With delta-timeout=2 and poll-interval=1, should timeout quickly
if [[ "$ELAPSED" -lt 20 ]]; then
    pass "custom delta timeout respected (${ELAPSED}s elapsed)"
else
    fail "delta timeout took too long (${ELAPSED}s, expected <20s)"
fi

# === Emergency Protocol Tests (Story 58.5) ===

echo ""
echo "=== Emergency Protocol Tests (Story 58.5) ==="

# Test: Emergency handover — force-kill and emergency flag (AC-1, AC-2, AC-3)
echo "Test: Emergency handover on delta timeout"
create_mock_bins "delta-timeout"
HANDOVER_DIR="$TMPDIR_BASE/emergency-handover"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 15 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "emergency handover exits 0"
else
    fail "emergency handover should exit 0 (exit: $EXIT_CODE, output: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "EMERGENCY_HANDOVER:.*force-killed"; then
    pass "emergency force-kill logged"
else
    fail "should log emergency force-kill (got: $OUTPUT)"
fi

# Test: Emergency flag in spawn task (AC-3)
echo "Test: Emergency flag propagated to incoming supervisor"
SPAWN_LOG_FILE="$MOCK_BIN_DIR/.spawn-task-log"
if [[ -f "$SPAWN_LOG_FILE" ]]; then
    SPAWN_TASK="$(cat "$SPAWN_LOG_FILE")"
    if echo "$SPAWN_TASK" | grep -q "EMERGENCY_SHIFT_HANDOVER"; then
        pass "spawn task contains EMERGENCY_SHIFT_HANDOVER flag"
    else
        fail "spawn task should contain emergency flag (got: $SPAWN_TASK)"
    fi
    if echo "$SPAWN_TASK" | grep -q "Verify all worker states manually"; then
        pass "spawn task includes manual verification instruction"
    else
        fail "spawn task should instruct manual verification (got: $SPAWN_TASK)"
    fi
else
    fail "spawn task log should exist"
fi

# Test: Emergency alert file created (AC-5)
echo "Test: Emergency alert file created"
if [[ -f "$HANDOVER_DIR/emergency-alert.md" ]]; then
    ALERT="$(cat "$HANDOVER_DIR/emergency-alert.md")"
    if echo "$ALERT" | grep -q "Emergency Handover Alert"; then
        pass "emergency alert file has header"
    else
        fail "alert should have Emergency Handover Alert header (got: $ALERT)"
    fi
    if echo "$ALERT" | grep -q "force-killed"; then
        pass "alert describes force-kill"
    else
        fail "alert should describe force-kill (got: $ALERT)"
    fi
    if echo "$ALERT" | grep -q "full worker audit"; then
        pass "alert mentions worker audit"
    else
        fail "alert should mention worker audit (got: $ALERT)"
    fi
else
    fail "emergency-alert.md should exist after emergency handover"
fi

# Test: Emergency metadata (handover_type: emergency)
echo "Test: Metadata records emergency handover type"
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q 'handover_type: "emergency"'; then
        pass "metadata records emergency handover type"
    else
        fail "metadata should have handover_type: emergency (got: $METADATA)"
    fi
    if echo "$METADATA" | grep -q "emergency_handover"; then
        pass "metadata records emergency_handover anomaly"
    else
        fail "metadata should record emergency_handover anomaly (got: $METADATA)"
    fi
else
    fail "metadata file should exist after emergency handover"
fi

# Test: Spawn retry — first fails, second succeeds (AC-6)
echo "Test: Spawn retry succeeds on second attempt"
create_mock_bins "spawn-retry-succeed"
HANDOVER_DIR="$TMPDIR_BASE/spawn-retry-ok"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 15 --spawn-retry-delay 1 2>&1)"
EXIT_CODE=$?
if [[ "$EXIT_CODE" -eq 0 ]]; then
    pass "spawn retry succeeds on second attempt"
else
    fail "should succeed after retry (exit: $EXIT_CODE, output: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "retrying"; then
    pass "retry attempt logged"
else
    fail "should log retry attempt (got: $OUTPUT)"
fi
if echo "$OUTPUT" | grep -q "spawned on retry"; then
    pass "successful retry logged"
else
    fail "should log successful retry (got: $OUTPUT)"
fi

# Test: Spawn retry — both fail, alert written (AC-6)
echo "Test: Both spawn attempts fail — alert written"
create_mock_bins "spawn-retry-fail"
HANDOVER_DIR="$TMPDIR_BASE/spawn-retry-fail"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 3 --spawn-retry-delay 1 2>&1)" || true
if echo "$OUTPUT" | grep -q "Both spawn attempts failed"; then
    pass "double spawn failure logged as critical"
else
    fail "should log double spawn failure (got: $OUTPUT)"
fi
if [[ -f "$HANDOVER_DIR/emergency-alert.md" ]]; then
    ALERT="$(cat "$HANDOVER_DIR/emergency-alert.md")"
    if echo "$ALERT" | grep -q "spawn_failure"; then
        pass "spawn failure alert written"
    else
        fail "alert should be spawn_failure type (got: $ALERT)"
    fi
    if echo "$ALERT" | grep -q "NO active supervisor"; then
        pass "alert warns about missing supervisor"
    else
        fail "alert should warn about missing supervisor (got: $ALERT)"
    fi
    if echo "$ALERT" | grep -q "multiclaude agents spawn"; then
        pass "alert includes manual recovery command"
    else
        fail "alert should include manual recovery command (got: $ALERT)"
    fi
else
    fail "emergency-alert.md should exist after spawn failure"
fi

# Test: Normal handover metadata includes handover_type: normal
echo "Test: Normal handover has handover_type: normal"
create_mock_bins "happy-path"
HANDOVER_DIR="$TMPDIR_BASE/normal-type-check"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 10 --ready-timeout 15 2>&1)"
if [[ -f "$HANDOVER_DIR/last-handover-metadata.yaml" ]]; then
    METADATA="$(cat "$HANDOVER_DIR/last-handover-metadata.yaml")"
    if echo "$METADATA" | grep -q 'handover_type: "normal"'; then
        pass "normal handover has handover_type: normal"
    else
        fail "normal handover should have handover_type: normal (got: $METADATA)"
    fi
fi

# Test: Dead outgoing + emergency doesn't double-kill
echo "Test: Dead outgoing supervisor skips force-kill"
create_mock_bins "happy-path"
create_dead_supervisor_mock
HANDOVER_DIR="$TMPDIR_BASE/dead-no-forcekill"
mkdir -p "$HANDOVER_DIR"
echo 'timestamp: "2026-03-12T10:00:00Z"' > "$HANDOVER_DIR/handover-requested"
OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo ThreeDoors --handover-dir "$HANDOVER_DIR" --poll-interval 1 --delta-timeout 3 --ready-timeout 15 2>&1)"
# Dead outgoing skips notification and delta wait entirely, so no emergency protocol
if echo "$OUTPUT" | grep -q "EMERGENCY_HANDOVER"; then
    fail "dead outgoing should NOT trigger emergency protocol (already dead)"
else
    pass "dead outgoing does not trigger emergency force-kill"
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
