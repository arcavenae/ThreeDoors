#!/usr/bin/env bash
set -euo pipefail

# handover-notify.sh — Handover Notification Helper (Story 58.7)
# Sends user-facing notifications at handover start and completion.
# Called by the handover orchestrator at key points in the sequence.
#
# Usage:
#   ./scripts/handover-notify.sh start [--repo NAME] [--handover-dir DIR]
#                                      [--trigger TYPE] [--zone ZONE]
#                                      [--jsonl-size BYTES] [--compressions N]
#   ./scripts/handover-notify.sh complete [--repo NAME] [--handover-dir DIR]
#                                          [--supervisor NAME] [--duration SECS]
#                                          [--mode MODE]
#
# Subcommands:
#   start      Send notification when handover begins (AC-4)
#   complete   Send notification when handover completes (AC-5)
#
# Options:
#   --repo NAME          Repository name (default: auto-detected)
#   --handover-dir DIR   Override handover directory
#   --trigger TYPE       Trigger type: "shift-clock" or "manual" (start only)
#   --zone ZONE          Zone at trigger time: RED/YELLOW/MANUAL (start only)
#   --jsonl-size BYTES   JSONL transcript size in bytes (start only)
#   --compressions N     Compression count (start only)
#   --supervisor NAME    Incoming supervisor name (complete only)
#   --duration SECS      Handover duration in seconds (complete only)
#   --mode MODE          Handover mode: "normal" or "emergency" (complete only)
#   --help, -h           Show this help message
#
# References:
#   - Story: docs/stories/58.7.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

# --- Parse subcommand ---
SUBCOMMAND="${1:-}"
if [[ -z "$SUBCOMMAND" ]] || [[ "$SUBCOMMAND" == "--help" ]] || [[ "$SUBCOMMAND" == "-h" ]]; then
    echo "Usage: handover-notify.sh {start|complete} [OPTIONS]"
    echo ""
    echo "Send user-facing handover notifications."
    echo ""
    echo "Subcommands:"
    echo "  start      Notify when handover begins (AC-4)"
    echo "  complete   Notify when handover completes (AC-5)"
    echo ""
    echo "Run 'handover-notify.sh start --help' or 'handover-notify.sh complete --help' for details."
    exit 0
fi
shift

# --- Defaults ---
REPO_NAME=""
HANDOVER_DIR=""
# Start-specific
TRIGGER=""
ZONE=""
JSONL_SIZE=""
COMPRESSIONS=""
# Complete-specific
SUPERVISOR_NAME=""
DURATION=""
MODE="normal"

# --- Parse arguments ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --trigger) TRIGGER="$2"; shift 2 ;;
        --zone) ZONE="$2"; shift 2 ;;
        --jsonl-size) JSONL_SIZE="$2"; shift 2 ;;
        --compressions) COMPRESSIONS="$2"; shift 2 ;;
        --supervisor) SUPERVISOR_NAME="$2"; shift 2 ;;
        --duration) DURATION="$2"; shift 2 ;;
        --mode) MODE="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: handover-notify.sh $SUBCOMMAND [OPTIONS]"
            echo ""
            if [[ "$SUBCOMMAND" == "start" ]]; then
                echo "Send notification when handover begins."
                echo ""
                echo "Options:"
                echo "  --trigger TYPE       Trigger type: shift-clock or manual"
                echo "  --zone ZONE          Zone: RED/YELLOW/MANUAL"
                echo "  --jsonl-size BYTES   JSONL transcript size"
                echo "  --compressions N     Compression count"
            elif [[ "$SUBCOMMAND" == "complete" ]]; then
                echo "Send notification when handover completes."
                echo ""
                echo "Options:"
                echo "  --supervisor NAME    Incoming supervisor name"
                echo "  --duration SECS      Handover duration in seconds"
                echo "  --mode MODE          Mode: normal or emergency"
            fi
            echo "  --repo NAME          Repository name (default: auto-detected)"
            echo "  --handover-dir DIR   Override handover directory"
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

# --- AC-4: Start notification ---
notify_start() {
    local trigger_desc="${TRIGGER:-unknown}"
    local zone_desc="${ZONE:-unknown}"
    local metrics=""

    if [[ -n "$JSONL_SIZE" ]]; then
        local size_mb
        size_mb="$(echo "scale=1; $JSONL_SIZE / 1048576" | bc 2>/dev/null || echo "?")"
        metrics="JSONL size: ${size_mb}MB"
    fi
    if [[ -n "$COMPRESSIONS" ]]; then
        if [[ -n "$metrics" ]]; then
            metrics="$metrics, compressions: $COMPRESSIONS"
        else
            metrics="compressions: $COMPRESSIONS"
        fi
    fi

    local trigger_label
    case "$trigger_desc" in
        shift-clock) trigger_label="Automatic" ;;
        manual)      trigger_label="Manual" ;;
        *)           trigger_label="Unknown ($trigger_desc)" ;;
    esac

    local message="SUPERVISOR_HANDOVER: $trigger_label shift change initiated. Zone: $zone_desc."
    if [[ -n "$metrics" ]]; then
        message="$message Metrics: [$metrics]."
    fi
    message="$message New supervisor starting."

    log "INFO" "Sending start notification: $message"

    if command -v multiclaude &>/dev/null; then
        multiclaude message send supervisor "$message" 2>/dev/null || true
    fi

    echo "$message"
}

# --- AC-5: Completion notification ---
notify_complete() {
    local supervisor="${SUPERVISOR_NAME:-supervisor}"
    local duration="${DURATION:-?}"
    local handover_mode="${MODE:-normal}"

    local mode_label
    case "$handover_mode" in
        normal)    mode_label="Normal" ;;
        emergency) mode_label="Emergency" ;;
        *)         mode_label="$handover_mode" ;;
    esac

    local message="SUPERVISOR_HANDOVER_COMPLETE: New supervisor [$supervisor] is online. Handover took ${duration} seconds. ${mode_label} mode."

    log "INFO" "Sending completion notification: $message"

    if command -v multiclaude &>/dev/null; then
        multiclaude message send supervisor "$message" 2>/dev/null || true
    fi

    echo "$message"
}

# === Main execution ===

case "$SUBCOMMAND" in
    start)
        notify_start
        ;;
    complete)
        notify_complete
        ;;
    *)
        echo "Error: Unknown subcommand: $SUBCOMMAND (expected 'start' or 'complete')" >&2
        exit 1
        ;;
esac
