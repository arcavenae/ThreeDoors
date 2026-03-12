#!/usr/bin/env bash
set -euo pipefail

# test-remote-collab.sh — Tests for remote-collab.sh
# Tests argument parsing, help output, error handling, and subcommand routing.
# Uses a mock multiclaude to avoid requiring the real CLI.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_UNDER_TEST="${SCRIPT_DIR}/remote-collab.sh"

PASS=0
FAIL=0
TOTAL=0

# Create a temp directory for mock binaries
MOCK_DIR=$(mktemp -d)
trap 'rm -rf "$MOCK_DIR"' EXIT

setup_mock_multiclaude() {
    local behavior="${1:-success}"
    case "$behavior" in
        success)
            cat > "${MOCK_DIR}/multiclaude" <<'MOCK'
#!/usr/bin/env bash
case "$1" in
    message)
        case "$2" in
            send) echo "Message sent (id: mock-msg-123)" ;;
            list) echo "No pending messages" ;;
            *) echo "mock: message $2" ;;
        esac
        ;;
    status) echo "mock: system status OK" ;;
    *) echo "mock: $*" ;;
esac
MOCK
            ;;
        has-messages)
            cat > "${MOCK_DIR}/multiclaude" <<'MOCK'
#!/usr/bin/env bash
case "$1" in
    message)
        case "$2" in
            send) echo "Message sent (id: mock-msg-456)" ;;
            list) echo "From: supervisor — Reply to your request" ;;
            *) echo "mock: message $2" ;;
        esac
        ;;
    status) echo "mock: system status OK" ;;
    *) echo "mock: $*" ;;
esac
MOCK
            ;;
        send-fail)
            cat > "${MOCK_DIR}/multiclaude" <<'MOCK'
#!/usr/bin/env bash
echo "connection refused" >&2
exit 1
MOCK
            ;;
    esac
    chmod +x "${MOCK_DIR}/multiclaude"
}

run_test() {
    local name="$1"
    shift
    TOTAL=$((TOTAL + 1))

    local expected_exit="${EXPECTED_EXIT:-0}"
    local actual_exit=0
    local output

    output=$("$@" 2>&1) || actual_exit=$?

    if [[ "$actual_exit" -ne "$expected_exit" ]]; then
        echo "FAIL: ${name}"
        echo "  Expected exit code: ${expected_exit}, got: ${actual_exit}"
        echo "  Output: ${output}"
        FAIL=$((FAIL + 1))
        return
    fi

    if [[ -n "${EXPECTED_PATTERN:-}" ]]; then
        if ! echo "$output" | grep -qE "$EXPECTED_PATTERN"; then
            echo "FAIL: ${name}"
            echo "  Output did not match pattern: ${EXPECTED_PATTERN}"
            echo "  Output: ${output}"
            FAIL=$((FAIL + 1))
            return
        fi
    fi

    if [[ -n "${NOT_EXPECTED_PATTERN:-}" ]]; then
        if echo "$output" | grep -qE "$NOT_EXPECTED_PATTERN"; then
            echo "FAIL: ${name}"
            echo "  Output unexpectedly matched pattern: ${NOT_EXPECTED_PATTERN}"
            echo "  Output: ${output}"
            FAIL=$((FAIL + 1))
            return
        fi
    fi

    echo "PASS: ${name}"
    PASS=$((PASS + 1))
}

# --- Help and usage tests ---

echo "=== Help and Usage Tests ==="

EXPECTED_EXIT=0 EXPECTED_PATTERN="Usage:" \
    run_test "global --help" "$SCRIPT_UNDER_TEST" --help

EXPECTED_EXIT=0 EXPECTED_PATTERN="Usage:" \
    run_test "global help" "$SCRIPT_UNDER_TEST" help

EXPECTED_EXIT=1 EXPECTED_PATTERN="Usage:" \
    run_test "no arguments shows usage and exits 1" "$SCRIPT_UNDER_TEST"

EXPECTED_EXIT=0 EXPECTED_PATTERN="v1\." \
    run_test "--version" "$SCRIPT_UNDER_TEST" --version

EXPECTED_EXIT=0 EXPECTED_PATTERN="send.*recipient.*message" \
    run_test "send --help" "$SCRIPT_UNDER_TEST" send --help

EXPECTED_EXIT=0 EXPECTED_PATTERN="timeout|interval" \
    run_test "wait-reply --help" "$SCRIPT_UNDER_TEST" wait-reply --help

EXPECTED_EXIT=0 EXPECTED_PATTERN="status" \
    run_test "status --help" "$SCRIPT_UNDER_TEST" status --help

EXPECTED_EXIT=0 EXPECTED_PATTERN="pending" \
    run_test "list --help" "$SCRIPT_UNDER_TEST" list --help

# --- Error handling tests ---

echo ""
echo "=== Error Handling Tests ==="

EXPECTED_EXIT=1 EXPECTED_PATTERN="Unknown command" \
    run_test "unknown command" "$SCRIPT_UNDER_TEST" foobar

EXPECTED_EXIT=1 EXPECTED_PATTERN="" \
    run_test "send with no args" "$SCRIPT_UNDER_TEST" send

EXPECTED_EXIT=1 EXPECTED_PATTERN="" \
    run_test "send with one arg" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" send supervisor

EXPECTED_EXIT=1 EXPECTED_PATTERN="must be a positive integer" \
    run_test "wait-reply --timeout non-numeric" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" wait-reply --timeout abc

EXPECTED_EXIT=1 EXPECTED_PATTERN="must be greater than 0" \
    run_test "wait-reply --interval 0" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" wait-reply --interval 0

EXPECTED_EXIT=1 EXPECTED_PATTERN="requires a value" \
    run_test "wait-reply --timeout without value" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" wait-reply --timeout

EXPECTED_EXIT=1 EXPECTED_PATTERN="Unknown option" \
    run_test "wait-reply unknown option" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" wait-reply --bogus

# --- Mock multiclaude tests ---

echo ""
echo "=== Send Command Tests ==="

setup_mock_multiclaude success

EXPECTED_EXIT=0 EXPECTED_PATTERN="mock-msg-123" \
    run_test "send message success" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" send supervisor "Hello world"

EXPECTED_EXIT=0 EXPECTED_PATTERN="Message sent to supervisor" \
    run_test "send prints confirmation with recipient" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" send supervisor "Hello"

setup_mock_multiclaude send-fail

EXPECTED_EXIT=1 EXPECTED_PATTERN="Failed to send" \
    run_test "send failure" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" send supervisor "Hello"

echo ""
echo "=== Status Command Tests ==="

setup_mock_multiclaude success

EXPECTED_EXIT=0 EXPECTED_PATTERN="status OK" \
    run_test "status command" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" status

echo ""
echo "=== List Command Tests ==="

EXPECTED_EXIT=0 EXPECTED_PATTERN="pending" \
    run_test "list command" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" list

echo ""
echo "=== Wait-Reply Tests ==="

setup_mock_multiclaude has-messages

EXPECTED_EXIT=0 EXPECTED_PATTERN="Reply received" \
    run_test "wait-reply finds message immediately" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" wait-reply --timeout 5 --interval 1

setup_mock_multiclaude success

EXPECTED_EXIT=2 EXPECTED_PATTERN="Timeout" \
    run_test "wait-reply times out with no messages" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" wait-reply --timeout 2 --interval 1

# --- Missing multiclaude tests ---

echo ""
echo "=== Missing CLI Tests ==="

EXPECTED_EXIT=1 EXPECTED_PATTERN="multiclaude CLI not found" \
    run_test "send without multiclaude in PATH" env PATH="/usr/bin:/bin" "$SCRIPT_UNDER_TEST" send supervisor "test"

EXPECTED_EXIT=1 EXPECTED_PATTERN="multiclaude CLI not found" \
    run_test "status without multiclaude in PATH" env PATH="/usr/bin:/bin" "$SCRIPT_UNDER_TEST" status

EXPECTED_EXIT=1 EXPECTED_PATTERN="multiclaude CLI not found" \
    run_test "list without multiclaude in PATH" env PATH="/usr/bin:/bin" "$SCRIPT_UNDER_TEST" list

# --- Non-interactive tests ---

echo ""
echo "=== Non-Interactive (SSH Simulation) Tests ==="

# Simulate no TTY by redirecting stdin from /dev/null
EXPECTED_EXIT=0 EXPECTED_PATTERN="mock-msg" \
    run_test "send works without TTY" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" send supervisor "SSH test" </dev/null

EXPECTED_EXIT=0 EXPECTED_PATTERN="status OK" \
    run_test "status works without TTY" env PATH="${MOCK_DIR}:${PATH}" "$SCRIPT_UNDER_TEST" status </dev/null

# --- Summary ---

echo ""
echo "================================"
echo "Results: ${PASS} passed, ${FAIL} failed, ${TOTAL} total"
echo "================================"

if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
