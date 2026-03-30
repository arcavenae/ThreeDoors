#!/usr/bin/env bash
set -euo pipefail

# test-verify-mcp-bridge.sh — Tests for verify-mcp-bridge.sh
# Tests argument parsing, help output, binary detection, and MCP protocol checks.
# Uses a mock MCP bridge binary to avoid requiring the real server.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_UNDER_TEST="${SCRIPT_DIR}/verify-mcp-bridge.sh"

PASS=0
FAIL=0
TOTAL=0

# Create a temp directory for mock binaries
MOCK_DIR=$(mktemp -d)
trap 'rm -rf "$MOCK_DIR"' EXIT

# Create a mock MCP bridge that returns valid responses
create_mock_bridge() {
    local behavior="${1:-success}"
    local binary="${MOCK_DIR}/multiclaude-mcp-bridge"

    case "$behavior" in
        success)
            cat > "$binary" <<'MOCK'
#!/usr/bin/env bash
# Mock MCP bridge — responds to initialize and tools/list
while IFS= read -r line; do
    case "$line" in
        *'"initialize"'*)
            echo '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"multiclaude-mcp-bridge","version":"test"}}}'
            ;;
        *'"tools/list"'*)
            echo '{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"multiclaude_status","description":"Status","inputSchema":{"type":"object","properties":{}}},{"name":"multiclaude_worker_list","description":"Workers","inputSchema":{"type":"object","properties":{}}},{"name":"multiclaude_message_list","description":"Messages","inputSchema":{"type":"object","properties":{}}},{"name":"multiclaude_message_read","description":"Read msg","inputSchema":{"type":"object","properties":{"message_id":{"type":"string"}}}},{"name":"multiclaude_repo_history","description":"History","inputSchema":{"type":"object","properties":{}}}]}}'
            ;;
    esac
done
MOCK
            ;;
        missing-tool)
            cat > "$binary" <<'MOCK'
#!/usr/bin/env bash
while IFS= read -r line; do
    case "$line" in
        *'"initialize"'*)
            echo '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"multiclaude-mcp-bridge","version":"test"}}}'
            ;;
        *'"tools/list"'*)
            echo '{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"multiclaude_status","description":"Status","inputSchema":{"type":"object","properties":{}}},{"name":"multiclaude_worker_list","description":"Workers","inputSchema":{"type":"object","properties":{}}}]}}'
            ;;
    esac
done
MOCK
            ;;
        bad-json)
            cat > "$binary" <<'MOCK'
#!/usr/bin/env bash
while IFS= read -r line; do
    echo "not json at all"
done
MOCK
            ;;
        wrong-name)
            cat > "$binary" <<'MOCK'
#!/usr/bin/env bash
while IFS= read -r line; do
    case "$line" in
        *'"initialize"'*)
            echo '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"wrong-server","version":"test"}}}'
            ;;
        *'"tools/list"'*)
            echo '{"jsonrpc":"2.0","id":2,"result":{"tools":[]}}'
            ;;
    esac
done
MOCK
            ;;
        no-output)
            cat > "$binary" <<'MOCK'
#!/usr/bin/env bash
exit 0
MOCK
            ;;
    esac
    chmod +x "$binary"
    echo "$binary"
}

# shellcheck disable=SC2317,SC2329 # Functions are called below; ShellCheck is confused by exit in heredocs
run_test() {
    local name="$1"
    local expected_exit="${2:-0}"
    shift 2

    TOTAL=$((TOTAL + 1))
    local output
    local actual_exit=0

    output=$("$@" 2>&1) || actual_exit=$?

    if [[ "$actual_exit" -eq "$expected_exit" ]]; then
        PASS=$((PASS + 1))
        printf "  ✓ %s\n" "$name"
    else
        FAIL=$((FAIL + 1))
        printf "  ✗ %s (exit %d, expected %d)\n" "$name" "$actual_exit" "$expected_exit"
        printf "    output: %s\n" "$output" | head -5
    fi
    echo "$output"
}

# shellcheck disable=SC2317,SC2329
run_test_grep() {
    local name="$1"
    local pattern="$2"
    local expected_exit="${3:-0}"
    shift 3

    TOTAL=$((TOTAL + 1))
    local output
    local actual_exit=0

    output=$("$@" 2>&1) || actual_exit=$?

    if [[ "$actual_exit" -ne "$expected_exit" ]]; then
        FAIL=$((FAIL + 1))
        printf "  ✗ %s (exit %d, expected %d)\n" "$name" "$actual_exit" "$expected_exit"
        printf "    output: %s\n" "$output" | head -5
        return
    fi

    if echo "$output" | grep -q "$pattern"; then
        PASS=$((PASS + 1))
        printf "  ✓ %s\n" "$name"
    else
        FAIL=$((FAIL + 1))
        printf "  ✗ %s (pattern '%s' not found)\n" "$name" "$pattern"
        printf "    output: %s\n" "$output" | head -5
    fi
}

echo "== Argument Parsing =="

run_test_grep "--help shows usage" "Usage:" 0 \
    "$SCRIPT_UNDER_TEST" --help

run_test_grep "--version shows version" "verify-mcp-bridge.sh" 0 \
    "$SCRIPT_UNDER_TEST" --version

run_test_grep "unknown option fails" "Unknown option" 1 \
    "$SCRIPT_UNDER_TEST" --invalid-flag

echo ""
echo "== Binary Detection =="

run_test_grep "nonexistent binary fails" "not found" 1 \
    "$SCRIPT_UNDER_TEST" --binary /nonexistent/path/binary

# Create a non-executable file
touch "${MOCK_DIR}/not-executable"
run_test_grep "non-executable binary fails" "not executable" 1 \
    "$SCRIPT_UNDER_TEST" --binary "${MOCK_DIR}/not-executable"

echo ""
echo "== MCP Protocol — Success =="

mock_binary=$(create_mock_bridge success)
run_test_grep "valid bridge passes all checks" "0 failed" 0 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

run_test_grep "finds all 5 tools" "5 tools" 0 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

run_test_grep "identifies server name" "multiclaude-mcp-bridge" 0 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

echo ""
echo "== MCP Protocol — Failure Modes =="

mock_binary=$(create_mock_bridge missing-tool)
run_test_grep "missing tools detected" "Tool missing" 1 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

mock_binary=$(create_mock_bridge bad-json)
run_test_grep "bad JSON detected" "not valid JSON" 1 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

mock_binary=$(create_mock_bridge wrong-name)
run_test_grep "wrong server name detected" "Unexpected server name" 1 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

mock_binary=$(create_mock_bridge no-output)
run_test_grep "no output detected" "No response" 1 \
    "$SCRIPT_UNDER_TEST" --binary "$mock_binary"

echo ""
echo "== Summary =="
printf "  %d passed, %d failed (of %d tests)\n" "$PASS" "$FAIL" "$TOTAL"

if [[ $FAIL -gt 0 ]]; then
    exit 1
fi
exit 0
