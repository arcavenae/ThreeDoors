#!/usr/bin/env bash
set -euo pipefail

# handover-orchestrator.sh — Daemon Handover Coordination Logic (Stories 58.3, 58.5)
# Orchestrates the full handover sequence: detect signal, notify outgoing,
# wait for delta, spawn incoming, wait for ready, kill outgoing.
# Includes emergency handover protocol: force-kill unresponsive supervisors,
# spawn with emergency flag, retry on spawn failure, alert user.
# Designed to run from the daemon refresh loop when a handover signal is detected.
#
# Usage:
#   ./scripts/handover-orchestrator.sh [--repo NAME] [--handover-dir DIR]
#                                       [--delta-timeout SECS] [--ready-timeout SECS]
#                                       [--poll-interval SECS] [--spawn-retry-delay SECS]
#
# Options:
#   --repo NAME              Repository name (default: auto-detected)
#   --handover-dir DIR       Override handover directory
#   --delta-timeout SECS     Timeout waiting for outgoing delta (default: 120)
#   --ready-timeout SECS     Timeout waiting for incoming ready (default: 180)
#   --poll-interval SECS     Polling interval for message checks (default: 5)
#   --spawn-retry-delay SECS Delay before retrying failed spawn (default: 30)
#   --help, -h               Show this help message
#
# References:
#   - Story: docs/stories/58.3.story.md, docs/stories/58.5.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

# --- Defaults (configurable via CLI flags) ---
REPO_NAME=""
HANDOVER_DIR=""
DELTA_TIMEOUT=120
READY_TIMEOUT=180
POLL_INTERVAL=5
SPAWN_RETRY_DELAY=30

# --- Parse arguments ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --delta-timeout) DELTA_TIMEOUT="$2"; shift 2 ;;
        --ready-timeout) READY_TIMEOUT="$2"; shift 2 ;;
        --poll-interval) POLL_INTERVAL="$2"; shift 2 ;;
        --spawn-retry-delay) SPAWN_RETRY_DELAY="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: handover-orchestrator.sh [--repo NAME] [--handover-dir DIR]"
            echo ""
            echo "Orchestrates supervisor shift handover sequence."
            echo ""
            echo "Options:"
            echo "  --repo NAME             Repository name (default: auto-detected)"
            echo "  --handover-dir DIR      Override handover directory"
            echo "  --delta-timeout SECS    Timeout for outgoing delta (default: 120)"
            echo "  --ready-timeout SECS    Timeout for incoming ready (default: 180)"
            echo "  --poll-interval SECS    Poll interval for messages (default: 5)"
            echo "  --spawn-retry-delay SECS  Delay before spawn retry (default: 30)"
            echo "  --help, -h              Show this help message"
            exit 0
            ;;
        *) echo "Error: Unknown option: $1" >&2; exit 1 ;;
    esac
done

# --- Auto-detect repo name ---
if [[ -z "$REPO_NAME" ]]; then
    if command -v multiclaude &>/dev/null; then
        REPO_NAME="$(multiclaude repo current 2>/dev/null | grep -oE '[^ ]+$' || true)"
    fi
    if [[ -z "$REPO_NAME" ]]; then
        REPO_NAME="$(basename "$(git remote get-url origin 2>/dev/null || echo 'unknown')" .git)"
    fi
fi

# --- Set handover directory ---
if [[ -z "$HANDOVER_DIR" ]]; then
    HANDOVER_DIR="$HOME/.multiclaude/handover/$REPO_NAME"
fi
mkdir -p "$HANDOVER_DIR"

# --- Logging ---
LOG_FILE="$HANDOVER_DIR/handover.log"

log() {
    local level="$1"
    shift
    local timestamp
    timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    local msg="[$timestamp] [$level] $*"
    echo "$msg"
    echo "$msg" >> "$LOG_FILE"
}

# --- Task 1: Signal file detection and consumption (AC-1) ---
SIGNAL_FILE="$HANDOVER_DIR/handover-requested"

detect_signal() {
    if [[ ! -f "$SIGNAL_FILE" ]]; then
        return 1
    fi

    # Read signal content before removing
    local signal_content=""
    if [[ -s "$SIGNAL_FILE" ]]; then
        signal_content="$(cat "$SIGNAL_FILE")"
    fi

    # Remove signal file atomically to prevent re-triggering
    rm -f "$SIGNAL_FILE"

    log "INFO" "Handover signal detected and consumed"
    if [[ -n "$signal_content" ]]; then
        log "INFO" "Signal content: $signal_content"
    fi

    return 0
}

# --- Task 2: Outgoing supervisor notification (AC-2) ---
notify_outgoing() {
    log "INFO" "Notifying outgoing supervisor of handover request"

    if ! multiclaude message send supervisor "HANDOVER_REQUESTED: Write your delta to shift-state.yaml and signal HANDOVER_COMPLETE" 2>/dev/null; then
        log "WARN" "Failed to send handover notification to supervisor"
        return 1
    fi

    log "INFO" "Handover notification sent to outgoing supervisor"
    return 0
}

# --- Task 3: Delta wait loop (AC-3) ---
# Returns 0 if delta received, 1 if timeout (emergency path)
wait_for_delta() {
    local elapsed=0

    log "INFO" "Waiting for outgoing supervisor delta (timeout: ${DELTA_TIMEOUT}s)"

    while [[ "$elapsed" -lt "$DELTA_TIMEOUT" ]]; do
        # Check for HANDOVER_COMPLETE message
        local messages
        messages="$(multiclaude message list 2>/dev/null || echo "")"

        if echo "$messages" | grep -q "HANDOVER_COMPLETE"; then
            log "INFO" "Received HANDOVER_COMPLETE from outgoing supervisor (${elapsed}s elapsed)"
            return 0
        fi

        # Alternative: check if shift-state.yaml was modified recently
        if [[ -f "$HANDOVER_DIR/shift-state.yaml" ]]; then
            local file_mtime now_epoch
            now_epoch="$(date -u +%s)"
            if stat -f%m "$HANDOVER_DIR/shift-state.yaml" &>/dev/null; then
                file_mtime="$(stat -f%m "$HANDOVER_DIR/shift-state.yaml")"
            else
                file_mtime="$(stat -c%Y "$HANDOVER_DIR/shift-state.yaml")"
            fi

            # If modified within the last poll interval, consider it fresh
            local age=$(( now_epoch - file_mtime ))
            if [[ "$age" -le "$POLL_INTERVAL" ]] && [[ "$elapsed" -gt 0 ]]; then
                log "INFO" "Detected fresh shift-state.yaml update (age: ${age}s, elapsed: ${elapsed}s)"
                return 0
            fi
        fi

        sleep "$POLL_INTERVAL"
        elapsed=$(( elapsed + POLL_INTERVAL ))
    done

    log "WARN" "Delta wait timed out after ${DELTA_TIMEOUT}s — proceeding with emergency protocol"
    return 1
}

# --- Task 4: Incoming supervisor spawn (AC-4) ---
spawn_incoming() {
    local task_msg="$1"

    log "INFO" "Spawning incoming supervisor"

    if ! multiclaude agents spawn --name supervisor --class persistent --prompt-file agents/supervisor.md --task "$task_msg" 2>/dev/null; then
        log "ERROR" "Failed to spawn incoming supervisor"
        return 1
    fi

    log "INFO" "Incoming supervisor spawned successfully"
    return 0
}

# --- Story 58.5: Force-kill outgoing supervisor (AC-2) ---
force_kill_outgoing() {
    local tmux_session="mc-$REPO_NAME"

    log "WARN" "EMERGENCY_HANDOVER: outgoing supervisor unresponsive, force-killed after ${DELTA_TIMEOUT}s timeout"

    # Kill supervisor tmux window if it exists
    if tmux list-windows -t "$tmux_session" -F '#{window_name}' 2>/dev/null | grep -q "^supervisor$"; then
        tmux kill-window -t "$tmux_session:supervisor" 2>/dev/null || true
        log "INFO" "Force-killed supervisor tmux window"
    fi

    # Also kill any supervisor-outgoing window
    if tmux list-windows -t "$tmux_session" -F '#{window_name}' 2>/dev/null | grep -q "supervisor-outgoing"; then
        tmux kill-window -t "$tmux_session:supervisor-outgoing" 2>/dev/null || true
        log "INFO" "Force-killed supervisor-outgoing tmux window"
    fi

    # Verify no supervisor process remains
    if tmux list-windows -t "$tmux_session" -F '#{window_name}' 2>/dev/null | grep -q "^supervisor"; then
        log "ERROR" "Supervisor process may still be running after force-kill"
        return 1
    fi

    log "INFO" "Outgoing supervisor force-killed and verified dead"
    return 0
}

# --- Story 58.5: Spawn incoming with retry (AC-3, AC-6) ---
spawn_incoming_with_retry() {
    local task_msg="$1"

    if spawn_incoming "$task_msg"; then
        return 0
    fi

    log "WARN" "First spawn attempt failed, retrying in ${SPAWN_RETRY_DELAY}s"
    sleep "$SPAWN_RETRY_DELAY"

    if spawn_incoming "$task_msg"; then
        log "INFO" "Incoming supervisor spawned on retry"
        return 0
    fi

    log "ERROR" "CRITICAL: Both spawn attempts failed — system has NO supervisor"
    return 1
}

# --- Story 58.5: User alerting (AC-5) ---
write_emergency_alert() {
    local alert_type="$1"
    local details="$2"
    local alert_file="$HANDOVER_DIR/emergency-alert.md"
    local tmp_file="$HANDOVER_DIR/.emergency-alert.md.tmp"
    local timestamp
    timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    {
        echo "# Emergency Handover Alert"
        echo ""
        echo "**Timestamp:** $timestamp"
        echo "**Repository:** $REPO_NAME"
        echo "**Alert Type:** $alert_type"
        echo ""
        echo "## What Happened"
        echo ""
        echo "$details"
        echo ""
        echo "## Action Required"
        echo ""
        case "$alert_type" in
            emergency_handover)
                echo "- The incoming supervisor is performing a full worker audit"
                echo "- Check the handover log at: $LOG_FILE"
                echo "- Review any discrepancies reported by the incoming supervisor"
                echo "- If issues persist, consider manual intervention"
                ;;
            spawn_failure)
                echo "- **CRITICAL: The system has NO active supervisor**"
                echo "- Manually spawn a supervisor: \`multiclaude agents spawn --name supervisor --class persistent --prompt-file agents/supervisor.md\`"
                echo "- Check daemon logs: \`multiclaude daemon logs\`"
                echo "- Check tmux session: \`tmux list-windows -t mc-$REPO_NAME\`"
                ;;
        esac
    } > "$tmp_file"

    mv "$tmp_file" "$alert_file"
    log "INFO" "Emergency alert written to $alert_file"

    # Also try to send a multiclaude message for visibility
    multiclaude message send supervisor "EMERGENCY_ALERT: $alert_type — check $alert_file for details" 2>/dev/null || true
}

# --- Task 5: Ready confirmation wait (AC-5) ---
# Returns 0 if ready confirmed, 1 if timeout
wait_for_ready() {
    local elapsed=0

    log "INFO" "Waiting for incoming supervisor ready confirmation (timeout: ${READY_TIMEOUT}s)"

    while [[ "$elapsed" -lt "$READY_TIMEOUT" ]]; do
        local messages
        messages="$(multiclaude message list 2>/dev/null || echo "")"

        if echo "$messages" | grep -q "READY"; then
            log "INFO" "Incoming supervisor confirmed READY (${elapsed}s elapsed)"
            return 0
        fi

        sleep "$POLL_INTERVAL"
        elapsed=$(( elapsed + POLL_INTERVAL ))
    done

    log "ERROR" "Incoming supervisor failed to confirm ready after ${READY_TIMEOUT}s"
    return 1
}

# --- Task 6: Outgoing termination (AC-6) ---
terminate_outgoing() {
    local tmux_session="mc-$REPO_NAME"

    log "INFO" "Terminating outgoing supervisor"

    # Find and kill the old supervisor window
    # The spawn in Task 4 replaced the supervisor window, but if the old
    # process is still running in a renamed window, kill it
    if tmux list-windows -t "$tmux_session" -F '#{window_name}' 2>/dev/null | grep -q "supervisor-outgoing"; then
        tmux kill-window -t "$tmux_session:supervisor-outgoing" 2>/dev/null || true
        log "INFO" "Killed outgoing supervisor window (supervisor-outgoing)"
    fi

    # Verify the outgoing process is gone
    # The multiclaude agents spawn --name supervisor replaces the window,
    # so the old process should already be dead. This is a safety check.
    log "INFO" "Outgoing supervisor terminated"
    return 0
}

# --- Task 7: Handover logging (AC-8) ---
log_handover_completion() {
    local start_epoch="$1"
    local delta_received="$2"
    local ready_confirmed="$3"
    local anomalies="${4:-none}"

    local end_epoch
    end_epoch="$(date -u +%s)"
    local duration=$(( end_epoch - start_epoch ))

    local metadata_file="$HANDOVER_DIR/last-handover-metadata.yaml"
    local tmp_file="$HANDOVER_DIR/.last-handover-metadata.yaml.tmp"
    local timestamp
    timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    local handover_type="${5:-normal}"

    {
        echo "timestamp: \"$timestamp\""
        echo "duration_seconds: $duration"
        echo "delta_received: $delta_received"
        echo "ready_confirmed: $ready_confirmed"
        echo "handover_type: \"$handover_type\""
        echo "anomalies: \"$anomalies\""
        echo "repo: \"$REPO_NAME\""
    } > "$tmp_file"

    mv "$tmp_file" "$metadata_file"

    # Update last-handover-timestamp for anti-oscillation guard
    date -u +%s > "$HANDOVER_DIR/last-handover-timestamp"

    log "INFO" "Handover complete: duration=${duration}s delta=$delta_received ready=$ready_confirmed anomalies=$anomalies"
}

# --- Handle dead outgoing supervisor ---
check_outgoing_alive() {
    local tmux_session="mc-$REPO_NAME"

    # Check if supervisor window exists in tmux
    if ! tmux list-windows -t "$tmux_session" -F '#{window_name}' 2>/dev/null | grep -q "^supervisor$"; then
        log "WARN" "Outgoing supervisor is not running (no tmux window found)"
        return 1
    fi

    return 0
}

# === Main execution ===

# AC-1: Detect and consume signal file
if ! detect_signal; then
    # No signal file — nothing to do
    exit 0
fi

HANDOVER_START="$(date -u +%s)"
DELTA_RECEIVED="false"
READY_CONFIRMED="false"
ANOMALIES=""

log "INFO" "=== Handover sequence initiated ==="

# Check if outgoing supervisor is alive
OUTGOING_ALIVE=true
if ! check_outgoing_alive; then
    OUTGOING_ALIVE=false
    ANOMALIES="outgoing_supervisor_dead"
    log "WARN" "Outgoing supervisor already dead — skipping notification and delta wait"
fi

# Track whether this is an emergency handover
EMERGENCY_HANDOVER=false

# AC-2 & AC-3: Notify outgoing and wait for delta
if $OUTGOING_ALIVE; then
    if notify_outgoing; then
        if wait_for_delta; then
            DELTA_RECEIVED="true"
        else
            # Delta timeout — activate emergency protocol (Story 58.5)
            EMERGENCY_HANDOVER=true
            ANOMALIES="${ANOMALIES:+${ANOMALIES},}delta_timeout,emergency_handover"

            # Force-kill the unresponsive outgoing supervisor
            if force_kill_outgoing; then
                OUTGOING_ALIVE=false
            else
                ANOMALIES="${ANOMALIES:+${ANOMALIES},}force_kill_incomplete"
            fi
        fi
    else
        ANOMALIES="${ANOMALIES:+${ANOMALIES},}notification_failed"
    fi
fi

# Build spawn task message based on handover type
STATE_FILE="$HANDOVER_DIR/shift-state.yaml"
if $EMERGENCY_HANDOVER; then
    SPAWN_TASK="EMERGENCY_SHIFT_HANDOVER: Previous supervisor was unresponsive. Using daemon snapshot only (no supervisor delta). Read $STATE_FILE and assume control. Verify all worker states manually."
else
    SPAWN_TASK="SHIFT_HANDOVER: Read $STATE_FILE and assume control"
fi

# Spawn incoming supervisor (with retry for emergency, fatal-exit for normal)
if $EMERGENCY_HANDOVER; then
    if ! spawn_incoming_with_retry "$SPAWN_TASK"; then
        ANOMALIES="${ANOMALIES:+${ANOMALIES},}spawn_failed"
        write_emergency_alert "spawn_failure" "Both spawn attempts failed after emergency handover. The outgoing supervisor was force-killed but no replacement could be started. The system currently has NO active supervisor."
        log_handover_completion "$HANDOVER_START" "$DELTA_RECEIVED" "$READY_CONFIRMED" "${ANOMALIES:-none}"
        exit 1
    fi
    write_emergency_alert "emergency_handover" "The outgoing supervisor was unresponsive and was force-killed after ${DELTA_TIMEOUT}s timeout. The incoming supervisor was spawned in emergency mode and is performing a full worker audit. Some context from the previous supervisor may have been lost."
else
    if ! spawn_incoming "$SPAWN_TASK"; then
        log "ERROR" "Fatal: Could not spawn incoming supervisor"
        ANOMALIES="${ANOMALIES:+${ANOMALIES},}spawn_failed"
        log_handover_completion "$HANDOVER_START" "$DELTA_RECEIVED" "$READY_CONFIRMED" "${ANOMALIES:-none}"
        exit 1
    fi
fi

# Wait for ready confirmation
if wait_for_ready; then
    READY_CONFIRMED="true"
else
    ANOMALIES="${ANOMALIES:+${ANOMALIES},}ready_timeout"
fi

# Terminate outgoing (graceful path only — emergency path already force-killed)
if $OUTGOING_ALIVE; then
    terminate_outgoing
fi

# AC-7 is maintained throughout — at most one supervisor dispatches (daemon is mutex)

# AC-8: Log handover metadata
HANDOVER_TYPE="normal"
if $EMERGENCY_HANDOVER; then
    HANDOVER_TYPE="emergency"
fi
log_handover_completion "$HANDOVER_START" "$DELTA_RECEIVED" "$READY_CONFIRMED" "${ANOMALIES:-none}" "$HANDOVER_TYPE"

log "INFO" "=== Handover sequence complete ==="
