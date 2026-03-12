#!/usr/bin/env bash
set -euo pipefail

# shift-status.sh — Shift Clock Status Display (Story 58.7)
# Displays current supervisor session metrics for integration with `multiclaude status`.
# Shows zone classification, JSONL size, compression count, session duration, and
# last handover time.
#
# Usage:
#   ./scripts/shift-status.sh [--repo NAME] [--handover-dir DIR] [--transcript PATH]
#                              [--json] [--compact]
#
# Options:
#   --repo NAME          Repository name (default: auto-detected)
#   --handover-dir DIR   Override handover directory
#   --transcript PATH    Override transcript path (default: auto-discovered)
#   --json               Output in JSON format
#   --compact            Single-line compact output
#   --help, -h           Show this help message
#
# Output:
#   Displays supervisor session metrics including zone, JSONL size,
#   compression count, session duration, and last handover time.
#
# References:
#   - Story: docs/stories/58.7.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

# --- Defaults ---
REPO_NAME=""
HANDOVER_DIR=""
TRANSCRIPT_PATH=""
OUTPUT_JSON=false
OUTPUT_COMPACT=false

# Thresholds (same defaults as shift-clock.sh)
YELLOW_COMPRESSIONS=3
YELLOW_SIZE_MB=5
RED_COMPRESSIONS=6
RED_SIZE_MB=10

# --- Parse arguments ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --transcript) TRANSCRIPT_PATH="$2"; shift 2 ;;
        --json) OUTPUT_JSON=true; shift ;;
        --compact) OUTPUT_COMPACT=true; shift ;;
        --help|-h)
            echo "Usage: shift-status.sh [--repo NAME] [--handover-dir DIR] [--transcript PATH]"
            echo "                       [--json] [--compact]"
            echo ""
            echo "Display current supervisor shift clock metrics."
            echo ""
            echo "Options:"
            echo "  --repo NAME          Repository name (default: auto-detected)"
            echo "  --handover-dir DIR   Override handover directory"
            echo "  --transcript PATH    Override transcript path (auto-discovered)"
            echo "  --json               Output in JSON format"
            echo "  --compact            Single-line compact output"
            echo "  --help, -h           Show this help message"
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

# --- Transcript discovery (reuse logic from shift-clock.sh) ---
discover_transcript() {
    if [[ -n "$TRANSCRIPT_PATH" ]]; then
        if [[ -f "$TRANSCRIPT_PATH" ]]; then
            echo "$TRANSCRIPT_PATH"
            return 0
        fi
        return 1
    fi

    local claude_projects_dir="$HOME/.claude/projects"
    if [[ ! -d "$claude_projects_dir" ]]; then
        return 1
    fi

    local latest_jsonl=""
    local latest_mtime=0

    for project_dir in "$claude_projects_dir"/*"$REPO_NAME"*/; do
        [[ -d "$project_dir" ]] || continue
        for jsonl_file in "$project_dir"*.jsonl; do
            [[ -f "$jsonl_file" ]] || continue
            local file_mtime
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
        return 1
    fi

    echo "$latest_jsonl"
}

# --- Metrics collection (reuse logic from shift-clock.sh) ---
collect_metrics() {
    local transcript="$1"

    local file_size
    if stat -f%z "$transcript" &>/dev/null; then
        file_size="$(stat -f%z "$transcript")"
    else
        file_size="$(stat -c%s "$transcript")"
    fi

    local compression_count
    compression_count="$(grep -c '"subtype":"compact_boundary"' "$transcript" 2>/dev/null || echo 0)"

    local assistant_count
    assistant_count="$(grep -c '"type":"assistant"' "$transcript" 2>/dev/null || echo 0)"

    echo "$file_size $compression_count $assistant_count"
}

# --- Zone evaluation (reuse logic from shift-clock.sh) ---
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

# --- Session duration ---
get_session_duration() {
    local transcript="$1"

    local first_timestamp
    first_timestamp="$(grep -m 1 '"timestamp"' "$transcript" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}' || echo "")"

    if [[ -z "$first_timestamp" ]]; then
        echo "unknown"
        return
    fi

    local session_start_epoch
    if date -j -u -f "%Y-%m-%dT%H:%M:%S" "$first_timestamp" "+%s" &>/dev/null; then
        session_start_epoch="$(date -j -u -f "%Y-%m-%dT%H:%M:%S" "$first_timestamp" "+%s")"
    else
        session_start_epoch="$(date -u -d "$first_timestamp" "+%s" 2>/dev/null || echo 0)"
    fi

    local now_epoch
    now_epoch="$(date -u +%s)"
    local elapsed_seconds=$(( now_epoch - session_start_epoch ))

    # Format as HH:MM:SS
    local hours=$((elapsed_seconds / 3600))
    local minutes=$(( (elapsed_seconds % 3600) / 60 ))
    local seconds=$((elapsed_seconds % 60))
    printf "%d:%02d:%02d" "$hours" "$minutes" "$seconds"
}

# --- Last handover time ---
get_last_handover() {
    local timestamp_file="$HANDOVER_DIR/last-handover-timestamp"

    if [[ ! -f "$timestamp_file" ]]; then
        echo "none"
        return
    fi

    local last_epoch
    last_epoch="$(cat "$timestamp_file" 2>/dev/null || echo 0)"

    if [[ "$last_epoch" -eq 0 ]]; then
        echo "none"
        return
    fi

    local now_epoch
    now_epoch="$(date -u +%s)"
    local elapsed_minutes=$(( (now_epoch - last_epoch) / 60 ))

    if [[ "$elapsed_minutes" -lt 60 ]]; then
        echo "${elapsed_minutes}m ago"
    else
        local hours=$((elapsed_minutes / 60))
        local mins=$((elapsed_minutes % 60))
        echo "${hours}h ${mins}m ago"
    fi
}

# --- Last handover metadata ---
get_last_handover_metadata() {
    local metadata_file="$HANDOVER_DIR/last-handover-metadata.yaml"

    if [[ ! -f "$metadata_file" ]]; then
        return
    fi

    cat "$metadata_file"
}

# === Main execution ===

# Discover transcript
transcript=""
transcript_found=true
if ! transcript="$(discover_transcript)"; then
    transcript_found=false
fi

# Collect metrics if transcript found
file_size=0
file_size_mb="0.0"
compression_count=0
assistant_count=0
zone="UNKNOWN"
session_duration="unknown"

if $transcript_found && [[ -n "$transcript" ]]; then
    read -r file_size compression_count assistant_count <<< "$(collect_metrics "$transcript")"
    file_size_mb="$(echo "scale=1; $file_size / 1048576" | bc 2>/dev/null || echo "?")"
    zone="$(evaluate_zone "$file_size" "$compression_count")"
    session_duration="$(get_session_duration "$transcript")"
fi

last_handover="$(get_last_handover)"

# --- Output ---

if $OUTPUT_JSON; then
    echo "{"
    echo "  \"zone\": \"$zone\","
    echo "  \"jsonl_size_bytes\": $file_size,"
    echo "  \"jsonl_size_mb\": \"$file_size_mb\","
    echo "  \"compression_count\": $compression_count,"
    echo "  \"assistant_messages\": $assistant_count,"
    echo "  \"session_duration\": \"$session_duration\","
    echo "  \"last_handover\": \"$last_handover\","
    echo "  \"transcript_found\": $transcript_found"
    echo "}"
elif $OUTPUT_COMPACT; then
    echo "Shift: zone=$zone size=${file_size_mb}MB compressions=$compression_count messages=$assistant_count duration=$session_duration last_handover=$last_handover"
else
    echo "Supervisor Shift Clock Status"
    echo "============================="
    if ! $transcript_found; then
        echo "  Transcript: not found (supervisor may not be running)"
    else
        echo "  Zone:              $zone"
        echo "  JSONL Size:        ${file_size_mb} MB"
        echo "  Compressions:      $compression_count"
        echo "  Assistant Messages: $assistant_count"
        echo "  Session Duration:  $session_duration"
    fi
    echo "  Last Handover:     $last_handover"

    # Show last handover metadata if available
    metadata="$(get_last_handover_metadata)"
    if [[ -n "$metadata" ]]; then
        echo ""
        echo "Last Handover Details"
        echo "---------------------"
        echo "$metadata" | while IFS= read -r line; do
            echo "  $line"
        done
    fi
fi
