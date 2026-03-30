#!/usr/bin/env bash
set -euo pipefail

# handover-history.sh — Handover History & Audit Trail (Story 58.6)
# Archives handover state files, logs events to JSONL, manages retention,
# provides a summary command, and detects anomalous handover frequency.
#
# Subcommands:
#   record        Archive state file, log event to JSONL, enforce retention
#   summary       Display recent handover events
#   check-frequency  Alert if handover frequency exceeds threshold
#
# Usage:
#   ./scripts/handover-history.sh record [--repo NAME] [--handover-dir DIR] \
#       --outgoing NAME --incoming NAME --type normal|emergency \
#       --trigger-metrics JSON --duration SECS --delta-received BOOL \
#       [--anomalies "comma,separated,list"]
#
#   ./scripts/handover-history.sh summary [--repo NAME] [--handover-dir DIR] \
#       [--last N]
#
#   ./scripts/handover-history.sh check-frequency [--repo NAME] [--handover-dir DIR] \
#       [--threshold N] [--window SECS]
#
# Options (global):
#   --repo NAME          Repository name (default: auto-detected)
#   --handover-dir DIR   Override handover directory
#   --help, -h           Show this help message
#
# Files:
#   history/shift-state-<ISO-timestamp>.yaml   Archived state snapshots
#   handover-log.jsonl                          Append-only handover event log
#
# References:
#   - Story: docs/stories/58.6.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

RETENTION_LIMIT=50
FREQUENCY_THRESHOLD=3
FREQUENCY_WINDOW=3600  # 1 hour in seconds

# --- Parse subcommand and global options ---
SUBCOMMAND=""
REPO_NAME=""
HANDOVER_DIR=""

# record options
OUTGOING=""
INCOMING=""
HANDOVER_TYPE=""
TRIGGER_METRICS=""
DURATION=""
DELTA_RECEIVED=""
ANOMALIES=""

# summary options
LAST_N=10

show_help() {
    cat << 'HELPEOF'
Usage: handover-history.sh <subcommand> [options]

Subcommands:
  record            Archive state file, log event, enforce retention
  summary           Display recent handover events
  check-frequency   Alert if handover frequency exceeds threshold

Global Options:
  --repo NAME          Repository name (default: auto-detected)
  --handover-dir DIR   Override handover directory
  --help, -h           Show this help message

Record Options:
  --outgoing NAME            Outgoing supervisor name
  --incoming NAME            Incoming supervisor name
  --type normal|emergency    Handover type
  --trigger-metrics JSON     Trigger metrics as JSON string
  --duration SECS            Handover duration in seconds
  --delta-received BOOL      Whether delta was received (true/false)
  --anomalies LIST           Comma-separated anomaly list (optional)

Summary Options:
  --last N                   Show last N events (default: 10)

Check-Frequency Options:
  --threshold N              Max handovers per window (default: 3)
  --window SECS              Window size in seconds (default: 3600)
HELPEOF
}

if [[ $# -eq 0 ]]; then
    show_help
    exit 0
fi

SUBCOMMAND="$1"
shift

if [[ "$SUBCOMMAND" == "--help" ]] || [[ "$SUBCOMMAND" == "-h" ]]; then
    show_help
    exit 0
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --outgoing) OUTGOING="$2"; shift 2 ;;
        --incoming) INCOMING="$2"; shift 2 ;;
        --type) HANDOVER_TYPE="$2"; shift 2 ;;
        --trigger-metrics) TRIGGER_METRICS="$2"; shift 2 ;;
        --duration) DURATION="$2"; shift 2 ;;
        --delta-received) DELTA_RECEIVED="$2"; shift 2 ;;
        --anomalies) ANOMALIES="$2"; shift 2 ;;
        --last) LAST_N="$2"; shift 2 ;;
        --threshold) FREQUENCY_THRESHOLD="$2"; shift 2 ;;
        --window) FREQUENCY_WINDOW="$2"; shift 2 ;;
        --help|-h) show_help; exit 0 ;;
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

HISTORY_DIR="$HANDOVER_DIR/history"
LOG_FILE="$HANDOVER_DIR/handover-log.jsonl"
STATE_FILE="$HANDOVER_DIR/shift-state.yaml"

# --- Subcommand: record ---
# AC-1: Archive state file
# AC-2: Append JSONL event
# AC-3: Enforce retention limit
cmd_record() {
    # Validate required options
    if [[ -z "$OUTGOING" ]]; then
        echo "Error: --outgoing is required for record" >&2
        exit 1
    fi
    if [[ -z "$INCOMING" ]]; then
        echo "Error: --incoming is required for record" >&2
        exit 1
    fi
    if [[ -z "$HANDOVER_TYPE" ]]; then
        echo "Error: --type is required for record" >&2
        exit 1
    fi
    if [[ -z "$DURATION" ]]; then
        echo "Error: --duration is required for record" >&2
        exit 1
    fi
    if [[ -z "$DELTA_RECEIVED" ]]; then
        echo "Error: --delta-received is required for record" >&2
        exit 1
    fi

    local timestamp
    timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    # AC-1: Archive state file to history/
    mkdir -p "$HISTORY_DIR"

    if [[ -f "$STATE_FILE" ]]; then
        local archive_name="shift-state-${timestamp}.yaml"
        # Sanitize colons in filename for filesystem compatibility
        archive_name="$(echo "$archive_name" | tr ':' '-')"
        cp "$STATE_FILE" "$HISTORY_DIR/$archive_name"
        echo "Archived: $HISTORY_DIR/$archive_name"
    else
        echo "Warning: No shift-state.yaml to archive" >&2
    fi

    # AC-2: Append JSONL event
    local anomalies_json="[]"
    if [[ -n "$ANOMALIES" ]]; then
        # Convert comma-separated list to JSON array
        # shellcheck disable=SC2001 # Pattern uses backreference (&) which parameter expansion can't do
        anomalies_json="[$(echo "$ANOMALIES" | sed 's/[^,]*/"&"/g')]"
    fi

    local trigger_json
    if [[ -n "$TRIGGER_METRICS" ]]; then
        trigger_json="$TRIGGER_METRICS"
    else
        trigger_json="{}"
    fi

    # Build JSON entry by writing directly to avoid bash brace expansion issues
    {
        echo -n '{"timestamp":"'"$timestamp"'"'
        echo -n ',"outgoing":"'"$OUTGOING"'"'
        echo -n ',"incoming":"'"$INCOMING"'"'
        echo -n ',"type":"'"$HANDOVER_TYPE"'"'
        echo -n ',"trigger_metrics":'"$trigger_json"
        echo -n ',"duration_seconds":'"$DURATION"
        echo -n ',"delta_received":'"$DELTA_RECEIVED"
        echo -n ',"anomalies":'"$anomalies_json"
        echo '}'
    } >> "$LOG_FILE"
    echo "Logged: handover event to $LOG_FILE"

    # AC-3: Enforce retention limit
    enforce_retention

    # AC-5: Check frequency after recording
    check_frequency_internal
}

# AC-3: History directory management
enforce_retention() {
    if [[ ! -d "$HISTORY_DIR" ]]; then
        return
    fi

    local file_count
    file_count="$(find "$HISTORY_DIR" -name 'shift-state-*.yaml' -type f 2>/dev/null | wc -l | tr -d ' ')"

    if [[ "$file_count" -gt "$RETENTION_LIMIT" ]]; then
        local excess=$(( file_count - RETENTION_LIMIT ))
        echo "Retention cleanup: removing $excess oldest archive(s) (limit: $RETENTION_LIMIT)"

        # Delete oldest files (by name, which sorts chronologically due to ISO timestamps)
        find "$HISTORY_DIR" -name 'shift-state-*.yaml' -type f 2>/dev/null | \
            sort | \
            head -n "$excess" | \
            while IFS= read -r old_file; do
                rm -f "$old_file"
                echo "  Removed: $(basename "$old_file")"
            done
    fi
}

# --- Subcommand: summary ---
# AC-4: Display recent handover events
cmd_summary() {
    if [[ ! -f "$LOG_FILE" ]]; then
        echo "No handover log found at $LOG_FILE"
        exit 0
    fi

    local total_entries
    total_entries="$(wc -l < "$LOG_FILE" | tr -d ' ')"

    if [[ "$total_entries" -eq 0 ]]; then
        echo "No handover events recorded."
        exit 0
    fi

    echo "=== Handover History (last $LAST_N of $total_entries events) ==="
    echo ""

    # Header
    printf "%-22s %-14s %-14s %-10s %8s %-6s %s\n" \
        "TIMESTAMP" "OUTGOING" "INCOMING" "TYPE" "DURATION" "DELTA" "ANOMALIES"
    printf '%.0s-' {1..95}
    echo ""

    # Display last N entries
    tail -n "$LAST_N" "$LOG_FILE" | while IFS= read -r line; do
        local ts out inc typ dur delta anoms
        ts="$(echo "$line" | sed -n 's/.*"timestamp":"\([^"]*\)".*/\1/p')"
        out="$(echo "$line" | sed -n 's/.*"outgoing":"\([^"]*\)".*/\1/p')"
        inc="$(echo "$line" | sed -n 's/.*"incoming":"\([^"]*\)".*/\1/p')"
        typ="$(echo "$line" | sed -n 's/.*"type":"\([^"]*\)".*/\1/p')"
        dur="$(echo "$line" | sed -n 's/.*"duration_seconds":\([0-9]*\).*/\1/p')"
        delta="$(echo "$line" | sed -n 's/.*"delta_received":\([a-z]*\).*/\1/p')"
        anoms="$(echo "$line" | sed -n 's/.*"anomalies":\[\([^]]*\)\].*/\1/p' | tr -d '"')"

        if [[ -z "$anoms" ]]; then
            anoms="none"
        fi

        printf "%-22s %-14s %-14s %-10s %6ss %-6s %s\n" \
            "$ts" "$out" "$inc" "$typ" "$dur" "$delta" "$anoms"
    done

    echo ""

    # Summary statistics
    local total_duration avg_duration emergency_count
    total_duration="$(awk -F'"duration_seconds":' '{print $2}' "$LOG_FILE" | awk -F'[^0-9]' '{s+=$1} END {print s+0}')"
    avg_duration=$(( total_duration / total_entries ))
    emergency_count="$(grep -c '"type":"emergency"' "$LOG_FILE" || true)"
    local normal_count=$(( total_entries - emergency_count ))
    local delta_true
    delta_true="$(grep -c '"delta_received":true' "$LOG_FILE" || true)"
    local delta_pct=$(( delta_true * 100 / total_entries ))

    echo "--- Statistics ---"
    echo "Total handovers:     $total_entries ($normal_count normal, $emergency_count emergency)"
    echo "Avg duration:        ${avg_duration}s"
    echo "Delta success rate:  ${delta_pct}% ($delta_true/$total_entries)"
}

# --- Subcommand: check-frequency ---
# AC-5: Anomaly alerting
cmd_check_frequency() {
    check_frequency_internal
}

check_frequency_internal() {
    if [[ ! -f "$LOG_FILE" ]]; then
        return
    fi

    local now_epoch
    now_epoch="$(date -u +%s)"
    local window_start=$(( now_epoch - FREQUENCY_WINDOW ))
    local count=0

    # Count handovers within the time window
    while IFS= read -r line; do
        local ts
        ts="$(echo "$line" | sed -n 's/.*"timestamp":"\([^"]*\)".*/\1/p')"
        if [[ -z "$ts" ]]; then
            continue
        fi

        # Convert ISO timestamp to epoch
        local entry_epoch
        if date -j -f "%Y-%m-%dT%H:%M:%SZ" "$ts" +%s &>/dev/null 2>&1; then
            # macOS
            entry_epoch="$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$ts" +%s 2>/dev/null || echo 0)"
        else
            # GNU/Linux
            entry_epoch="$(date -d "$ts" +%s 2>/dev/null || echo 0)"
        fi

        if [[ "$entry_epoch" -ge "$window_start" ]]; then
            count=$(( count + 1 ))
        fi
    done < "$LOG_FILE"

    if [[ "$count" -gt "$FREQUENCY_THRESHOLD" ]]; then
        local window_hours=$(( FREQUENCY_WINDOW / 3600 ))
        echo "WARNING: HIGH_HANDOVER_FREQUENCY: $count handovers in the last ${window_hours} hour(s). Consider increasing shift clock thresholds." >&2
        return 1
    fi

    return 0
}

# --- Dispatch subcommand ---
case "$SUBCOMMAND" in
    record) cmd_record ;;
    summary) cmd_summary ;;
    check-frequency) cmd_check_frequency ;;
    *)
        echo "Error: Unknown subcommand: $SUBCOMMAND" >&2
        echo "Use --help for usage information." >&2
        exit 1
        ;;
esac
