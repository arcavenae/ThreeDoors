#!/usr/bin/env bash
set -euo pipefail

# verify-mcp-bridge.sh — End-to-end verification of the MCP bridge server
# Checks that the binary exists, the daemon is running, and the MCP
# protocol handshake succeeds with all expected tools listed.
#
# Usage: ./scripts/verify-mcp-bridge.sh [--binary <path>]
# All operations are read-only — nothing is modified.

readonly VERSION="1.0.0"

# Default binary locations to check (in order)
readonly -a DEFAULT_PATHS=(
    "./bin/multiclaude-mcp-bridge"
    "$(command -v multiclaude-mcp-bridge 2>/dev/null || true)"
    "/usr/local/bin/multiclaude-mcp-bridge"
)

# Expected MCP tools that must appear in tools/list response
readonly -a EXPECTED_TOOLS=(
    "multiclaude_status"
    "multiclaude_worker_list"
    "multiclaude_message_list"
    "multiclaude_message_read"
    "multiclaude_repo_history"
)

BINARY=""
PASS=0
FAIL=0
WARN=0

usage() {
    cat <<'EOF'
Usage: verify-mcp-bridge.sh [options]

Verifies the multiclaude MCP bridge server is correctly installed and
functioning. All checks are read-only.

Options:
  --binary <path>    Path to the multiclaude-mcp-bridge binary
  --help             Show this help message
  --version          Show version

Checks performed:
  1. Binary exists and is executable
  2. multiclaude daemon is running
  3. MCP initialize handshake succeeds
  4. All expected tools are listed in tools/list response

Exit codes:
  0  All checks passed
  1  One or more checks failed
EOF
}

# Output helpers
pass() { PASS=$((PASS + 1)); printf "  ✓ %s\n" "$1"; }
fail() { FAIL=$((FAIL + 1)); printf "  ✗ %s\n" "$1"; }
warn() { WARN=$((WARN + 1)); printf "  ! %s\n" "$1"; }
section() { printf "\n== %s ==\n" "$1"; }

# run_with_timeout — portable timeout for piping input to a command.
# Usage: run_with_timeout <seconds> <input_string> <command> [args...]
# Returns the command's stdout. Kills the command if it exceeds the timeout.
run_with_timeout() {
    local secs="$1"
    local input="$2"
    shift 2

    local tmpout
    tmpout=$(mktemp)

    # Run the command in the background, capture output
    echo "$input" | "$@" > "$tmpout" 2>/dev/null &
    local pid=$!

    # Background watchdog to kill if timeout exceeded
    (
        sleep "$secs"
        kill "$pid" 2>/dev/null || true
    ) &
    local watchdog=$!

    # Wait for the command to finish
    wait "$pid" 2>/dev/null || true

    # Kill the watchdog if still running
    kill "$watchdog" 2>/dev/null || true
    wait "$watchdog" 2>/dev/null || true

    cat "$tmpout"
    rm -f "$tmpout"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --binary)
            BINARY="$2"
            shift 2
            ;;
        --help)
            usage
            exit 0
            ;;
        --version)
            echo "verify-mcp-bridge.sh v${VERSION}"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

# --- Check 1: Binary exists and is executable ---
section "Binary Check"

if [[ -n "$BINARY" ]]; then
    if [[ -x "$BINARY" ]]; then
        pass "Binary found at $BINARY"
    elif [[ -f "$BINARY" ]]; then
        fail "Binary exists at $BINARY but is not executable"
        echo "       Fix: chmod +x $BINARY"
    else
        fail "Binary not found at $BINARY"
    fi
else
    found=false
    for path in "${DEFAULT_PATHS[@]}"; do
        if [[ -n "$path" && -x "$path" ]]; then
            BINARY="$path"
            pass "Binary found at $BINARY"
            found=true
            break
        fi
    done
    if ! $found; then
        fail "Binary not found in default locations"
        echo "       Searched: ./bin/multiclaude-mcp-bridge, PATH, /usr/local/bin/"
        echo "       Fix: just build-mcp-bridge"
    fi
fi

# --- Check 2: multiclaude daemon ---
section "Daemon Check"

if command -v multiclaude >/dev/null 2>&1; then
    pass "multiclaude CLI found in PATH"
    if multiclaude daemon status >/dev/null 2>&1; then
        pass "multiclaude daemon is running"
    else
        warn "multiclaude daemon is not running (bridge will fail at runtime)"
        echo "       Fix: multiclaude daemon start"
    fi
else
    warn "multiclaude CLI not found in PATH"
    echo "       The MCP bridge shells out to multiclaude — it must be in PATH"
fi

# --- Check 3: MCP initialize handshake ---
section "MCP Protocol Check"

if [[ -z "$BINARY" || ! -x "$BINARY" ]]; then
    fail "Skipping protocol checks — binary not available"
else
    # Send initialize request and read response
    init_request='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"verify-mcp-bridge","version":"1.0.0"}}}'

    init_response=$(run_with_timeout 5 "$init_request" "$BINARY")

    if [[ -z "$init_response" ]]; then
        fail "No response to initialize request"
    else
        # Check for valid JSON-RPC response
        if echo "$init_response" | head -1 | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null; then
            pass "Initialize response is valid JSON"
        else
            fail "Initialize response is not valid JSON"
            echo "       Response: $init_response"
        fi

        # Check protocol version in response
        if echo "$init_response" | head -1 | python3 -c "
import sys, json
resp = json.load(sys.stdin)
v = resp.get('result', {}).get('protocolVersion', '')
if v: sys.exit(0)
sys.exit(1)
" 2>/dev/null; then
            pass "Server returned protocol version"
        else
            warn "Could not verify protocol version in response"
        fi

        # Check server info
        server_name=$(echo "$init_response" | head -1 | python3 -c "
import sys, json
resp = json.load(sys.stdin)
print(resp.get('result', {}).get('serverInfo', {}).get('name', ''))
" 2>/dev/null || true)

        if [[ "$server_name" == "multiclaude-mcp-bridge" ]]; then
            pass "Server identifies as multiclaude-mcp-bridge"
        else
            fail "Unexpected server name: '${server_name}'"
        fi
    fi

    # --- Check 4: tools/list ---
    section "Tools Check"

    # Send initialize + tools/list (server needs initialize first)
    tools_request='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
    combined_input=$(printf '%s\n%s' "$init_request" "$tools_request")

    tools_output=$(run_with_timeout 5 "$combined_input" "$BINARY")

    if [[ -z "$tools_output" ]]; then
        fail "No response to tools/list request"
    else
        # The output has two lines (one per response). Get the second line.
        tools_response=$(echo "$tools_output" | tail -1)

        # Extract tool names
        tool_names=$(echo "$tools_response" | python3 -c "
import sys, json
resp = json.load(sys.stdin)
tools = resp.get('result', {}).get('tools', [])
for t in tools:
    print(t.get('name', ''))
" 2>/dev/null || true)

        if [[ -z "$tool_names" ]]; then
            fail "Could not parse tool names from tools/list response"
        else
            tool_count=$(echo "$tool_names" | wc -l | tr -d ' ')
            pass "Server reports $tool_count tools"

            for expected in "${EXPECTED_TOOLS[@]}"; do
                if echo "$tool_names" | grep -q "^${expected}$"; then
                    pass "Tool found: $expected"
                else
                    fail "Tool missing: $expected"
                fi
            done
        fi
    fi
fi

# --- Summary ---
section "Summary"
total=$((PASS + FAIL + WARN))
printf "  %d passed, %d failed, %d warnings (of %d checks)\n" "$PASS" "$FAIL" "$WARN" "$total"

if [[ $FAIL -gt 0 ]]; then
    echo ""
    echo "Some checks failed. See above for details and fix suggestions."
    exit 1
fi

if [[ $WARN -gt 0 ]]; then
    echo ""
    echo "All critical checks passed, but there are warnings."
fi

exit 0
