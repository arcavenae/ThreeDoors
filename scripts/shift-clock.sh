#!/usr/bin/env bash
set -euo pipefail

# shift-clock.sh — Supervisor Transcript Monitor (Story 58.1)
# Monitors the supervisor's JSONL transcript for context window utilization signals.
# Designed to run every 5 minutes as part of the daemon refresh loop.
#
# Usage:
#   ./scripts/shift-clock.sh [--repo NAME] [--handover-dir DIR] [--transcript PATH]
#                            [--min-session-minutes N] [--min-handover-gap-minutes N]
#                            [--yellow-compressions N] [--yellow-size-mb N]
#                            [--red-compressions N] [--red-size-mb N]
#                            [--seam-idle-seconds N]
#
# Options:
#   --repo NAME                    Repository name (default: auto-detected)
#   --handover-dir DIR             Override handover directory
#   --transcript PATH              Override transcript path (default: auto-discovered)
#   --min-session-minutes N        Minimum session age before handover (default: 30)
#   --min-handover-gap-minutes N   Minimum gap between handovers (default: 30)
#   --yellow-compressions N        Compression count for yellow zone (default: 3)
#   --yellow-size-mb N             JSONL size in MB for yellow zone (default: 5)
#   --red-compressions N           Compression count for red zone (default: 6)
#   --red-size-mb N                JSONL size in MB for red zone (default: 10)
#   --seam-idle-seconds N          Seconds of inactivity for natural seam (default: 60)
#   --help, -h                     Show this help message
#
# Output:
#   Logs zone classification and metrics to stdout.
#   Writes handover signal file when conditions are met.
#
# References:
#   - Story: docs/stories/58.1.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

# --- Defaults (configurable via CLI flags) ---
REPO_NAME=""
HANDOVER_DIR=""
TRANSCRIPT_PATH=""
MIN_SESSION_MINUTES=30
MIN_HANDOVER_GAP_MINUTES=30
YELLOW_COMPRESSIONS=3
YELLOW_SIZE_MB=5
RED_COMPRESSIONS=6
RED_SIZE_MB=10
SEAM_IDLE_SECONDS=60

# --- Parse arguments ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --transcript) TRANSCRIPT_PATH="$2"; shift 2 ;;
        --min-session-minutes) MIN_SESSION_MINUTES="$2"; shift 2 ;;
        --min-handover-gap-minutes) MIN_HANDOVER_GAP_MINUTES="$2"; shift 2 ;;
        --yellow-compressions) YELLOW_COMPRESSIONS="$2"; shift 2 ;;
        --yellow-size-mb) YELLOW_SIZE_MB="$2"; shift 2 ;;
        --red-compressions) RED_COMPRESSIONS="$2"; shift 2 ;;
        --red-size-mb) RED_SIZE_MB="$2"; shift 2 ;;
        --seam-idle-seconds) SEAM_IDLE_SECONDS="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: shift-clock.sh [--repo NAME] [--handover-dir DIR] [--transcript PATH]"
            echo ""
            echo "Monitors the supervisor's JSONL transcript for context window utilization."
            echo ""
            echo "Options:"
            echo "  --repo NAME                    Repository name (default: auto-detected)"
            echo "  --handover-dir DIR             Override handover directory"
            echo "  --transcript PATH              Override transcript path (auto-discovered)"
            echo "  --min-session-minutes N        Min session age before handover (default: 30)"
            echo "  --min-handover-gap-minutes N   Min gap between handovers (default: 30)"
            echo "  --yellow-compressions N        Compressions for yellow zone (default: 3)"
            echo "  --yellow-size-mb N             JSONL size MB for yellow zone (default: 5)"
            echo "  --red-compressions N           Compressions for red zone (default: 6)"
            echo "  --red-size-mb N                JSONL size MB for red zone (default: 10)"
            echo "  --seam-idle-seconds N          Idle seconds for natural seam (default: 60)"
            echo "  --help, -h                     Show this help message"
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

# --- AC-1: Transcript discovery ---
# Find the supervisor's active JSONL transcript at ~/.claude/projects/<project-hash>/<session-id>.jsonl
discover_transcript() {
    if [[ -n "$TRANSCRIPT_PATH" ]]; then
        if [[ -f "$TRANSCRIPT_PATH" ]]; then
            echo "$TRANSCRIPT_PATH"
            return 0
        fi
        echo "Error: Specified transcript not found: $TRANSCRIPT_PATH" >&2
        return 1
    fi

    local claude_projects_dir="$HOME/.claude/projects"
    if [[ ! -d "$claude_projects_dir" ]]; then
        echo "Error: Claude projects directory not found: $claude_projects_dir" >&2
        return 1
    fi

    # Find the most recently modified JSONL file across all project directories
    # that match the repo name pattern (the supervisor's project hash dir)
    local latest_jsonl=""
    local latest_mtime=0

    # Look in project dirs that reference this repo
    for project_dir in "$claude_projects_dir"/*"$REPO_NAME"*/; do
        [[ -d "$project_dir" ]] || continue
        for jsonl_file in "$project_dir"*.jsonl; do
            [[ -f "$jsonl_file" ]] || continue
            local file_mtime
            # macOS stat vs Linux stat
            if stat -f%m "$jsonl_file" &>/dev/null; then
                file_mtime="$(stat -f%m "$jsonl_file")"
            else
                file_mtime="$(stat -c%Y "$jsonl_file")"
            fi
            if [[ "$file_mtime" -gt "$latest_mtime" ]]; then
                latest_mtime="$file_mtime"
                latest_jsonl="$jsonl_file"
            fi
        done
    done

    if [[ -z "$latest_jsonl" ]]; then
        echo "Error: No JSONL transcript found for repo $REPO_NAME" >&2
        return 1
    fi

    echo "$latest_jsonl"
}

# --- AC-2: Metrics collection ---
collect_metrics() {
    local transcript="$1"

    # File size in bytes
    local file_size
    if stat -f%z "$transcript" &>/dev/null; then
        file_size="$(stat -f%z "$transcript")"
    else
        file_size="$(stat -c%s "$transcript")"
    fi

    # Compression event count (compact_boundary markers)
    local compression_count
    compression_count="$(grep -c '"subtype":"compact_boundary"' "$transcript" 2>/dev/null || echo 0)"

    # Assistant message count
    local assistant_count
    assistant_count="$(grep -c '"type":"assistant"' "$transcript" 2>/dev/null || echo 0)"

    echo "$file_size $compression_count $assistant_count"
}

# --- AC-3: Three-tier threshold evaluation ---
evaluate_zone() {
    local file_size="$1"
    local compression_count="$2"

    local yellow_size_bytes=$((YELLOW_SIZE_MB * 1024 * 1024))
    local red_size_bytes=$((RED_SIZE_MB * 1024 * 1024))

    if [[ "$compression_count" -ge "$RED_COMPRESSIONS" ]] || [[ "$file_size" -ge "$red_size_bytes" ]]; then
        echo "RED"
    elif [[ "$compression_count" -ge "$YELLOW_COMPRESSIONS" ]] || [[ "$file_size" -ge "$yellow_size_bytes" ]]; then
        echo "YELLOW"
    else
        echo "GREEN"
    fi
}

# --- AC-4: Time floor enforcement ---
# Returns 0 (true) if time floor is satisfied, 1 (false) if too early
check_time_floor() {
    local transcript="$1"

    # Get the earliest timestamp from the transcript (session start)
    local first_timestamp
    first_timestamp="$(grep -m 1 '"timestamp"' "$transcript" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}' || echo "")"

    if [[ -z "$first_timestamp" ]]; then
        # Can't determine session start — allow handover
        return 0
    fi

    local session_start_epoch
    # macOS date vs GNU date — timestamps are UTC (end with Z)
    if date -j -u -f "%Y-%m-%dT%H:%M:%S" "$first_timestamp" "+%s" &>/dev/null; then
        session_start_epoch="$(date -j -u -f "%Y-%m-%dT%H:%M:%S" "$first_timestamp" "+%s")"
    else
        session_start_epoch="$(date -u -d "$first_timestamp" "+%s" 2>/dev/null || echo 0)"
    fi

    local now_epoch
    now_epoch="$(date -u +%s)"
    local elapsed_minutes=$(( (now_epoch - session_start_epoch) / 60 ))

    if [[ "$elapsed_minutes" -lt "$MIN_SESSION_MINUTES" ]]; then
        echo "Session age: ${elapsed_minutes}m (floor: ${MIN_SESSION_MINUTES}m)" >&2
        return 1
    fi

    return 0
}

# --- AC-5: Anti-oscillation guard ---
# Returns 0 (true) if enough time has passed since last handover, 1 (false) if too soon
check_anti_oscillation() {
    local last_handover_file="$HANDOVER_DIR/last-handover-timestamp"

    if [[ ! -f "$last_handover_file" ]]; then
        return 0
    fi

    local last_handover_epoch
    last_handover_epoch="$(cat "$last_handover_file" 2>/dev/null || echo 0)"

    local now_epoch
    now_epoch="$(date -u +%s)"
    local elapsed_minutes=$(( (now_epoch - last_handover_epoch) / 60 ))

    if [[ "$elapsed_minutes" -lt "$MIN_HANDOVER_GAP_MINUTES" ]]; then
        echo "Last handover: ${elapsed_minutes}m ago (minimum gap: ${MIN_HANDOVER_GAP_MINUTES}m)" >&2
        return 1
    fi

    return 0
}

# --- AC-6: Natural seam detection ---
# Returns 0 (true) if at a natural seam, 1 (false) if mid-operation
check_natural_seam() {
    local transcript="$1"
    local now_epoch
    now_epoch="$(date -u +%s)"
    local idle_threshold=$((now_epoch - SEAM_IDLE_SECONDS))

    # Check for recent multiclaude work commands (tool_use with "multiclaude work" in the content)
    local last_tool_timestamp
    last_tool_timestamp="$(grep '"tool_use"' "$transcript" | tail -1 | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}' || echo "")"

    if [[ -n "$last_tool_timestamp" ]]; then
        local tool_epoch
        # Timestamps are UTC
        if date -j -u -f "%Y-%m-%dT%H:%M:%S" "$last_tool_timestamp" "+%s" &>/dev/null; then
            tool_epoch="$(date -j -u -f "%Y-%m-%dT%H:%M:%S" "$last_tool_timestamp" "+%s")"
        else
            tool_epoch="$(date -u -d "$last_tool_timestamp" "+%s" 2>/dev/null || echo 0)"
        fi

        if [[ "$tool_epoch" -gt "$idle_threshold" ]]; then
            echo "Active tool call detected (${last_tool_timestamp})" >&2
            return 1
        fi
    fi

    # Check for pending messages
    if command -v multiclaude &>/dev/null; then
        local pending_messages
        pending_messages="$(multiclaude message list 2>/dev/null || echo "")"
        if [[ -n "$pending_messages" ]] && ! echo "$pending_messages" | grep -qi "no.*messages\|no.*pending\|^$"; then
            echo "Pending messages detected" >&2
            return 1
        fi
    fi

    return 0
}

# --- AC-7: Signal file output ---
write_handover_signal() {
    local zone="$1"
    local file_size="$2"
    local compression_count="$3"
    local assistant_count="$4"

    local signal_dir="$HANDOVER_DIR"
    local signal_file="$signal_dir/handover-requested"
    local tmp_file="$signal_dir/.handover-requested.tmp"
    local timestamp
    timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    {
        echo "timestamp: \"$timestamp\""
        echo "zone: \"$zone\""
        echo "metrics:"
        echo "  file_size_bytes: $file_size"
        echo "  compression_count: $compression_count"
        echo "  assistant_message_count: $assistant_count"
        echo "trigger: \"shift-clock\""
    } > "$tmp_file"

    # Atomic write
    mv "$tmp_file" "$signal_file"

    # Record handover timestamp for anti-oscillation
    date -u +%s > "$HANDOVER_DIR/last-handover-timestamp"

    echo "HANDOVER REQUESTED: Signal file written to $signal_file"
}

# === Main execution ===

# AC-1: Discover transcript
transcript="$(discover_transcript)" || exit 0

# AC-2: Collect metrics
read -r file_size compression_count assistant_count <<< "$(collect_metrics "$transcript")"

# Convert file size for display
file_size_mb="$(echo "scale=2; $file_size / 1048576" | bc 2>/dev/null || echo "?")"

# AC-3: Evaluate zone
zone="$(evaluate_zone "$file_size" "$compression_count")"

echo "Shift clock: zone=$zone file_size=${file_size_mb}MB compressions=$compression_count messages=$assistant_count transcript=$transcript"

# Green zone — nothing to do
if [[ "$zone" == "GREEN" ]]; then
    exit 0
fi

# Yellow zone — advisory logging only
if [[ "$zone" == "YELLOW" ]]; then
    echo "ADVISORY: Approaching context limits. Consider wrapping up current work."
    exit 0
fi

# Red zone — check handover conditions
# AC-5: Anti-oscillation guard
if ! check_anti_oscillation; then
    echo "DEFERRED: Handover pending anti-oscillation guard"
    exit 0
fi

# AC-4: Time floor enforcement
if ! check_time_floor "$transcript"; then
    echo "DEFERRED: Handover pending time floor"
    exit 0
fi

# AC-6: Natural seam detection
if ! check_natural_seam "$transcript"; then
    echo "DEFERRED: Handover pending natural seam"
    exit 0
fi

# AC-7: All conditions met — write signal file
write_handover_signal "$zone" "$file_size" "$compression_count" "$assistant_count"
