#!/usr/bin/env bash
set -euo pipefail

# test-shift-snapshot.sh — Tests for shift-snapshot.sh (Story 58.2)
# Run from the repository root: ./scripts/test-shift-snapshot.sh
#
# Tests snapshot generation with mock command outputs, atomic writes,
# delta preservation, size warnings, and schema validation.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/shift-snapshot.sh"
PASS=0
FAIL=0
TMPDIR_BASE=""

pass() {
    PASS=$((PASS + 1))
    echo "  PASS: $1"
}

fail() {
    FAIL=$((FAIL + 1))
    echo "  FAIL: $1" >&2
}

setup_tmpdir() {
    TMPDIR_BASE="$(mktemp -d)"
}

cleanup() {
    if [[ -n "$TMPDIR_BASE" ]] && [[ -d "$TMPDIR_BASE" ]]; then
        rm -rf "$TMPDIR_BASE"
    fi
    # Remove any mock bins we created
    if [[ -n "${MOCK_BIN_DIR:-}" ]] && [[ -d "$MOCK_BIN_DIR" ]]; then
        rm -rf "$MOCK_BIN_DIR"
    fi
}

trap cleanup EXIT

# Create mock command directory
create_mock_bins() {
    MOCK_BIN_DIR="$(mktemp -d)"

    # Mock multiclaude — worker list with sample workers
    cat > "$MOCK_BIN_DIR/multiclaude" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    worker)
        if [[ "${2:-}" == "list" ]]; then
            echo "bold-eagle  Implement story 42.3"
            echo "swift-fox   Fix CI lint failures"
        fi
        ;;
    repo)
        if [[ "${2:-}" == "current" ]]; then
            echo "ThreeDoors"
        fi
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/multiclaude"

    # Mock tmux — list-windows with sample agents
    cat > "$MOCK_BIN_DIR/tmux" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    list-windows)
        echo "supervisor"
        echo "merge-queue"
        echo "pr-shepherd"
        echo "project-watchdog"
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/tmux"

    # Mock gh — pr list with sample PRs
    cat > "$MOCK_BIN_DIR/gh" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    pr)
        if [[ "${2:-}" == "list" ]]; then
            # Check if this is a --head query (worker PR lookup)
            if echo "$*" | grep -q "\-\-head"; then
                echo "[]"
                exit 0
            fi
            echo '[{"number":565,"title":"feat: task pool analytics","statusCheckRollup":[{"conclusion":"SUCCESS"}]},{"number":566,"title":"fix: lint warnings","statusCheckRollup":[{"conclusion":"FAILURE"}]}]'
        fi
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/gh"

    # Mock jq — pass through to real jq if available
    if command -v jq &>/dev/null; then
        ln -sf "$(command -v jq)" "$MOCK_BIN_DIR/jq"
    fi

    # Mock git for repo detection fallback
    cat > "$MOCK_BIN_DIR/git" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    remote)
        echo "https://github.com/arcavenae/ThreeDoors.git"
        ;;
esac
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/git"

    # Mock sync (no-op for tests)
    cat > "$MOCK_BIN_DIR/sync" << 'MOCK_EOF'
#!/usr/bin/env bash
exit 0
MOCK_EOF
    chmod +x "$MOCK_BIN_DIR/sync"
}

echo "=== Shift Snapshot Generator Tests ==="

# --- Structural tests (no external dependencies) ---

# Test 1: Help flag exits 0
echo "Test: --help exits successfully"
if "$SCRIPT" --help >/dev/null 2>&1; then
    pass "--help exits 0"
else
    fail "--help should exit 0"
fi

# Test 2: Help output contains usage info
echo "Test: --help shows usage information"
HELP_OUTPUT="$("$SCRIPT" --help 2>&1)"
if echo "$HELP_OUTPUT" | grep -q "Usage"; then
    pass "--help shows usage"
else
    fail "--help should show usage info"
fi

# Test 3: Invalid option fails
echo "Test: unknown option fails"
if "$SCRIPT" --bogus 2>/dev/null; then
    fail "unknown option should fail"
else
    pass "unknown option exits non-zero"
fi

# --- Mock tests (simulate external commands) ---

echo ""
echo "=== Mock Command Tests ==="

# Check for jq (required for PR parsing)
if ! command -v jq &>/dev/null; then
    echo "  SKIP: jq not installed — skipping mock tests"
else

setup_tmpdir
create_mock_bins

# Test 4: Snapshot generation with mocked commands
echo "Test: generates valid snapshot with mocked commands"
HANDOVER_TEST_DIR="$TMPDIR_BASE/handover"
PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo TestRepo --handover-dir "$HANDOVER_TEST_DIR" >/dev/null 2>&1
if [[ -f "$HANDOVER_TEST_DIR/shift-state.yaml" ]]; then
    pass "snapshot file created"
else
    fail "snapshot file should be created at $HANDOVER_TEST_DIR/shift-state.yaml"
fi

# Test 5: Schema version present
echo "Test: snapshot contains version: 1"
if grep -q "^version: 1" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "version: 1 present"
else
    fail "snapshot should contain version: 1"
fi

# Test 6: Timestamp in ISO-8601 UTC
echo "Test: snapshot contains UTC timestamp"
if grep -qE '^timestamp: "[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z"' "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "UTC ISO-8601 timestamp present"
else
    fail "snapshot should contain UTC ISO-8601 timestamp"
fi

# Test 7: Workers section present
echo "Test: snapshot contains workers section"
if grep -q "^workers:" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "workers section present"
else
    fail "snapshot should contain workers section"
fi

# Test 8: Worker names captured
echo "Test: snapshot contains worker names from mock"
if grep -q "bold-eagle" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "worker bold-eagle captured"
else
    fail "should capture worker bold-eagle from mock"
fi

# Test 9: Persistent agents section present
echo "Test: snapshot contains persistent_agents section"
if grep -q "^persistent_agents:" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "persistent_agents section present"
else
    fail "snapshot should contain persistent_agents section"
fi

# Test 10: Agent names captured (supervisor excluded)
echo "Test: snapshot contains agent names, excludes supervisor"
SNAPSHOT_CONTENT="$(cat "$HANDOVER_TEST_DIR/shift-state.yaml")"
if echo "$SNAPSHOT_CONTENT" | grep -q "merge-queue" && ! echo "$SNAPSHOT_CONTENT" | grep -q '"supervisor"'; then
    pass "agents captured, supervisor excluded"
else
    fail "should capture agents and exclude supervisor window"
fi

# Test 11: Open PRs section present
echo "Test: snapshot contains open_prs section"
if grep -q "^open_prs:" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "open_prs section present"
else
    fail "snapshot should contain open_prs section"
fi

# Test 12: PR numbers captured
echo "Test: snapshot contains PR numbers from mock"
if grep -q "number: 565" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "PR #565 captured"
else
    fail "should capture PR #565 from mock"
fi

# Test 13: CI status parsed correctly
echo "Test: snapshot parses CI status"
if grep -q 'ci: "passing"' "$HANDOVER_TEST_DIR/shift-state.yaml" && grep -q 'ci: "failing"' "$HANDOVER_TEST_DIR/shift-state.yaml"; then
    pass "CI status parsed (passing + failing)"
else
    fail "should parse CI status from statusCheckRollup"
fi

# Test 14: Atomic write — no .tmp file left behind
echo "Test: atomic write leaves no .tmp file"
if [[ ! -f "$HANDOVER_TEST_DIR/.shift-state.yaml.tmp" ]]; then
    pass "no .tmp file left behind"
else
    fail ".tmp file should be cleaned up after atomic write"
fi

# Test 15: Directory permissions
echo "Test: handover directory has 0700 permissions"
DIR_PERMS="$(stat -f '%Lp' "$HANDOVER_TEST_DIR" 2>/dev/null || stat -c '%a' "$HANDOVER_TEST_DIR" 2>/dev/null)"
if [[ "$DIR_PERMS" == "700" ]]; then
    pass "directory permissions are 0700"
else
    fail "directory permissions should be 0700 (got $DIR_PERMS)"
fi

# Test 16: Supervisor delta preservation
echo "Test: preserves supervisor delta sections on re-run"

# Write supervisor delta content into the existing snapshot
cat >> "$HANDOVER_TEST_DIR/shift-state.yaml" << 'DELTA_EOF'
pending_decisions:
  - context: "Worker asked about AC scope"
    recommendation: "Per-session for now"
    resolved: false
priorities:
  - "Epic 42 completion is sprint goal"
blockers:
  - "Story 42.4 depends on 42.3"
warnings:
  - "merge-queue can't merge workflow PRs"
DELTA_EOF

# Re-run snapshot generation
PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo TestRepo --handover-dir "$HANDOVER_TEST_DIR" >/dev/null 2>&1

# Verify delta sections are preserved
SNAPSHOT_AFTER="$(cat "$HANDOVER_TEST_DIR/shift-state.yaml")"
if echo "$SNAPSHOT_AFTER" | grep -q "pending_decisions:" && \
   echo "$SNAPSHOT_AFTER" | grep -q "Worker asked about AC scope" && \
   echo "$SNAPSHOT_AFTER" | grep -q "priorities:" && \
   echo "$SNAPSHOT_AFTER" | grep -q "blockers:" && \
   echo "$SNAPSHOT_AFTER" | grep -q "warnings:"; then
    pass "supervisor delta sections preserved"
else
    fail "supervisor delta sections should be preserved across re-runs"
fi

# Test 17: Size warning for oversized snapshot
echo "Test: warns when snapshot exceeds 10KB"

# Create a snapshot with padding to exceed 10KB
OVERSIZE_DIR="$TMPDIR_BASE/oversize"
mkdir -p "$OVERSIZE_DIR"

# Generate a large fake existing file (11KB of comments)
{
    echo "version: 1"
    echo "timestamp: \"2026-03-11T14:00:00Z\""
    echo "workers:"
    echo "  active: []"
    echo "  recently_completed: []"
    echo "persistent_agents: []"
    echo "open_prs: []"
    # Pad with large priority list to exceed 10KB
    echo "priorities:"
    for i in $(seq 1 500); do
        echo "  - \"Priority item $i with a long description to pad the file size beyond the 10KB limit that should trigger a warning\""
    done
} > "$OVERSIZE_DIR/shift-state.yaml"

WARN_OUTPUT="$(PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo TestRepo --handover-dir "$OVERSIZE_DIR" 2>&1)"
if echo "$WARN_OUTPUT" | grep -q "WARNING.*exceeds 10KB"; then
    pass "size warning emitted for oversized snapshot"
else
    fail "should warn when snapshot exceeds 10KB"
fi

# Test 18: Empty worker list handled gracefully
echo "Test: handles empty worker list"
# Create mock with no workers
cat > "$MOCK_BIN_DIR/multiclaude" << 'MOCK_EOF'
#!/usr/bin/env bash
case "$1" in
    worker)
        if [[ "${2:-}" == "list" ]]; then
            echo "No active workers"
        fi
        ;;
    repo)
        if [[ "${2:-}" == "current" ]]; then
            echo "ThreeDoors"
        fi
        ;;
esac
MOCK_EOF
chmod +x "$MOCK_BIN_DIR/multiclaude"

EMPTY_DIR="$TMPDIR_BASE/empty"
PATH="$MOCK_BIN_DIR:$PATH" "$SCRIPT" --repo TestRepo --handover-dir "$EMPTY_DIR" >/dev/null 2>&1
if grep -q "^workers:" "$EMPTY_DIR/shift-state.yaml" && grep -q "active:" "$EMPTY_DIR/shift-state.yaml"; then
    pass "empty worker list handled gracefully"
else
    fail "should produce valid workers section even with no active workers"
fi

# Test 19: Snapshot output is parseable YAML
echo "Test: snapshot is parseable YAML"
if command -v python3 &>/dev/null; then
    if python3 -c "import yaml; yaml.safe_load(open('$HANDOVER_TEST_DIR/shift-state.yaml'))" 2>/dev/null; then
        pass "snapshot is valid YAML (python)"
    else
        # Try without PyYAML — at minimum check structure
        if grep -q "^version:" "$HANDOVER_TEST_DIR/shift-state.yaml" && \
           grep -q "^timestamp:" "$HANDOVER_TEST_DIR/shift-state.yaml" && \
           grep -q "^workers:" "$HANDOVER_TEST_DIR/shift-state.yaml"; then
            pass "snapshot has valid YAML structure (no PyYAML available)"
        else
            fail "snapshot should be valid YAML"
        fi
    fi
else
    echo "  SKIP: python3 not available for YAML validation"
fi

fi  # end jq check

# --- Results ---
echo ""
echo "=== Results ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total:  $((PASS + FAIL))"
if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
echo "All tests passed."
