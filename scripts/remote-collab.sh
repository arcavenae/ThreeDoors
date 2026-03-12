#!/usr/bin/env bash
set -euo pipefail

# remote-collab.sh — Bidirectional messaging wrapper for multiclaude
# Sends messages, polls for replies, and displays system status.
# Designed for SSH/non-interactive use.
#
# Usage: ./scripts/remote-collab.sh <command> [options]
# Commands: send, wait-reply, status, list
# Run with --help for details.

readonly VERSION="1.0.0"
readonly DEFAULT_INTERVAL=10
readonly DEFAULT_TIMEOUT=300

# Exit codes
readonly EXIT_SUCCESS=0
readonly EXIT_ERROR=1
readonly EXIT_TIMEOUT=2

usage() {
    cat <<'EOF'
Usage: remote-collab.sh <command> [options]

Commands:
  send <recipient> <message>   Send a message via multiclaude
  wait-reply [options]         Poll for new messages until reply or timeout
  status                       Show multiclaude system status
  list                         List pending messages

Options for wait-reply:
  --timeout <seconds>    Maximum wait time (default: 300)
  --interval <seconds>   Poll interval (default: 10)

Global options:
  --help                 Show this help message
  --version              Show version

Exit codes:
  0  Success / reply received
  1  Error
  2  Timeout (wait-reply only)

Examples:
  remote-collab.sh send supervisor "Review my branch feature-x"
  remote-collab.sh wait-reply --timeout 120 --interval 5
  remote-collab.sh status
  remote-collab.sh list
  ssh host './scripts/remote-collab.sh send supervisor "task request"'
EOF
}

usage_send() {
    cat <<'EOF'
Usage: remote-collab.sh send <recipient> <message>

Send a message to a multiclaude agent.

Arguments:
  recipient   Target agent name (e.g., supervisor, merge-queue)
  message     Message text (quote if it contains spaces)

Example:
  remote-collab.sh send supervisor "Please review PR #42"
EOF
}

usage_wait_reply() {
    cat <<'EOF'
Usage: remote-collab.sh wait-reply [options]

Poll for new messages until a reply arrives or timeout elapses.

Options:
  --timeout <seconds>    Maximum wait time (default: 300)
  --interval <seconds>   Poll interval in seconds (default: 10)

Exit codes:
  0  Reply received
  2  Timeout elapsed without reply

Example:
  remote-collab.sh wait-reply --timeout 120 --interval 5
EOF
}

usage_status() {
    cat <<'EOF'
Usage: remote-collab.sh status

Display current multiclaude system status including agents, workers,
and pending messages.
EOF
}

usage_list() {
    cat <<'EOF'
Usage: remote-collab.sh list

List all pending multiclaude messages.
EOF
}

die() {
    echo "Error: $*" >&2
    exit "$EXIT_ERROR"
}

require_multiclaude() {
    if ! command -v multiclaude &>/dev/null; then
        die "multiclaude CLI not found in PATH"
    fi
}

cmd_send() {
    if [[ $# -lt 2 ]]; then
        usage_send >&2
        exit "$EXIT_ERROR"
    fi

    local recipient="$1"
    local message="$2"

    require_multiclaude

    local output
    output=$(multiclaude message send "$recipient" "$message" 2>&1) || {
        die "Failed to send message: $output"
    }

    echo "$output"
    echo "Message sent to $recipient at $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
}

cmd_wait_reply() {
    local timeout="$DEFAULT_TIMEOUT"
    local interval="$DEFAULT_INTERVAL"

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --timeout)
                [[ $# -lt 2 ]] && die "--timeout requires a value"
                timeout="$2"
                shift 2
                ;;
            --interval)
                [[ $# -lt 2 ]] && die "--interval requires a value"
                interval="$2"
                shift 2
                ;;
            --help)
                usage_wait_reply
                exit "$EXIT_SUCCESS"
                ;;
            *)
                die "Unknown option: $1"
                ;;
        esac
    done

    # Validate numeric arguments
    [[ "$timeout" =~ ^[0-9]+$ ]] || die "--timeout must be a positive integer"
    [[ "$interval" =~ ^[0-9]+$ ]] || die "--interval must be a positive integer"
    [[ "$interval" -gt 0 ]] || die "--interval must be greater than 0"

    require_multiclaude

    local start_time
    start_time=$(date +%s)
    local elapsed=0

    echo "Waiting for reply (timeout: ${timeout}s, interval: ${interval}s)..."

    while [[ "$elapsed" -lt "$timeout" ]]; do
        local output
        output=$(multiclaude message list 2>&1) || true

        # Check if there are messages (non-empty, not just "no messages" type output)
        if [[ -n "$output" ]] && ! echo "$output" | grep -qi "no.*messages\|no.*pending\|empty"; then
            echo "--- Reply received ---"
            echo "$output"
            exit "$EXIT_SUCCESS"
        fi

        sleep "$interval"
        elapsed=$(( $(date +%s) - start_time ))
    done

    echo "Timeout: no reply received after ${timeout} seconds" >&2
    exit "$EXIT_TIMEOUT"
}

cmd_status() {
    require_multiclaude
    multiclaude status
}

cmd_list() {
    require_multiclaude
    multiclaude message list
}

# Main argument parsing
if [[ $# -eq 0 ]]; then
    usage >&2
    exit "$EXIT_ERROR"
fi

case "$1" in
    send)
        shift
        if [[ "${1:-}" == "--help" ]]; then
            usage_send
            exit "$EXIT_SUCCESS"
        fi
        cmd_send "$@"
        ;;
    wait-reply)
        shift
        cmd_wait_reply "$@"
        ;;
    status)
        shift
        if [[ "${1:-}" == "--help" ]]; then
            usage_status
            exit "$EXIT_SUCCESS"
        fi
        cmd_status
        ;;
    list)
        shift
        if [[ "${1:-}" == "--help" ]]; then
            usage_list
            exit "$EXIT_SUCCESS"
        fi
        cmd_list
        ;;
    --help|-h|help)
        usage
        exit "$EXIT_SUCCESS"
        ;;
    --version)
        echo "remote-collab.sh v${VERSION}"
        exit "$EXIT_SUCCESS"
        ;;
    *)
        die "Unknown command: $1. Run with --help for usage."
        ;;
esac
