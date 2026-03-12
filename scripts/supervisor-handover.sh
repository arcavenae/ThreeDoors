#!/usr/bin/env bash
set -euo pipefail

# supervisor-handover.sh — Manual Supervisor Handover Trigger (Story 58.7)
# Allows users to manually trigger a supervisor handover when they notice
# degradation the shift clock missed (false negative recovery — Failure A1).
#
# Usage:
#   ./scripts/supervisor-handover.sh [--force] [--repo NAME] [--handover-dir DIR]
#
# Options:
#   --force              Skip anti-oscillation check entirely (AC-3)
#   --repo NAME          Repository name (default: auto-detected)
#   --handover-dir DIR   Override handover directory
#   --help, -h           Show this help message
#
# Behavior:
#   - Writes the same handover signal file as shift-clock.sh (trigger: "manual")
#   - Checks anti-oscillation guard: warns if recent handover, asks for confirmation (AC-2)
#   - --force bypasses anti-oscillation entirely (AC-3)
#   - The handover proceeds via handover-orchestrator.sh (Story 58.3)
#
# References:
#   - Story: docs/stories/58.7.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

# --- Defaults ---
REPO_NAME=""
HANDOVER_DIR=""
FORCE=false
MIN_HANDOVER_GAP_MINUTES=30

# --- Parse arguments ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --force) FORCE=true; shift ;;
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --min-handover-gap-minutes) MIN_HANDOVER_GAP_MINUTES="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: supervisor-handover.sh [--force] [--repo NAME] [--handover-dir DIR]"
            echo ""
            echo "Manually trigger a supervisor shift handover."
            echo ""
            echo "Options:"
            echo "  --force              Skip anti-oscillation check (AC-3)"
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

# --- AC-2: Anti-oscillation check ---
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
        echo "WARNING: Last handover was ${elapsed_minutes} minutes ago (minimum gap: ${MIN_HANDOVER_GAP_MINUTES} minutes)." >&2
        return 1
    fi

    return 0
}

# --- AC-1 & AC-3: Write handover signal file ---
write_handover_signal() {
    local signal_file="$HANDOVER_DIR/handover-requested"
    local tmp_file="$HANDOVER_DIR/.handover-requested.tmp"
    local timestamp
    timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    {
        echo "timestamp: \"$timestamp\""
        echo "zone: \"MANUAL\""
        echo "metrics:"
        echo "  file_size_bytes: 0"
        echo "  compression_count: 0"
        echo "  assistant_message_count: 0"
        echo "trigger: \"manual\""
        if $FORCE; then
            echo "force: true"
        fi
    } > "$tmp_file"

    # Atomic write
    mv "$tmp_file" "$signal_file"

    echo "HANDOVER REQUESTED: Manual handover signal written to $signal_file"
    echo "The handover orchestrator will pick this up on the next daemon cycle."
}

# === Main execution ===

# AC-3: --force bypasses anti-oscillation entirely
if $FORCE; then
    echo "Force mode: skipping anti-oscillation check."
    write_handover_signal
    exit 0
fi

# AC-2: Check anti-oscillation guard
if ! check_anti_oscillation; then
    # Recent handover — prompt user for confirmation
    if [[ -t 0 ]]; then
        # Interactive terminal — ask for confirmation
        printf "Proceed with handover anyway? [y/N] " >&2
        read -r response
        case "$response" in
            [yY]|[yY][eE][sS])
                echo "User confirmed manual override of anti-oscillation guard."
                write_handover_signal
                exit 0
                ;;
            *)
                echo "Handover cancelled."
                exit 1
                ;;
        esac
    else
        # Non-interactive — suggest --force
        echo "Use --force to override anti-oscillation guard." >&2
        exit 1
    fi
else
    write_handover_signal
fi
