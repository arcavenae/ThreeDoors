#!/bin/bash
# poc_applescript_read.sh - Read tasks from Apple Notes via osascript plaintext
# Requires: macOS, Automation permission for Notes.app

set -euo pipefail

NOTE_TITLE="ThreeDoors Tasks"
# SECURITY NOTE: NOTE_TITLE is hardcoded. If ever sourced from user input,
# it MUST be sanitized to prevent AppleScript injection (escape " and \).

echo "=== AppleScript Read PoC ==="
echo "Reading note: $NOTE_TITLE"
echo "---"

# Approach 1: plaintext text (preferred)
# Using heredoc to avoid shell interpolation issues with osascript
PLAINTEXT=$(osascript <<APPLESCRIPT 2>&1
tell application "Notes" to get plaintext text of note "$NOTE_TITLE"
APPLESCRIPT
) || {
    echo "[FAIL] osascript read failed: $PLAINTEXT" >&2
    exit 1
}

# Parse tasks
LINE_NUM=0
TASK_COUNT=0
while IFS= read -r line; do
    LINE_NUM=$((LINE_NUM + 1))
    trimmed=$(echo "$line" | sed 's/^[[:space:]]*//')
    [ -z "$trimmed" ] && continue

    TASK_COUNT=$((TASK_COUNT + 1))
    if echo "$trimmed" | grep -qE '^\- \[x\]|^\* \[x\]|^\- \[X\]'; then
        STATUS="DONE"
    else
        STATUS="TODO"
    fi
    # Strip checkbox prefix
    TEXT=$(echo "$trimmed" | sed 's/^- \[.\] //;s/^\* \[.\] //')
    echo "$TASK_COUNT. [$STATUS] $TEXT"
done <<< "$PLAINTEXT"

echo "---"
echo "[OK] Read $TASK_COUNT tasks from \"$NOTE_TITLE\""

# Also test body (HTML) approach for comparison
echo ""
echo "=== HTML Body Approach (fallback) ==="
HTML_BODY=$(osascript -e "tell application \"Notes\" to get body of note \"$NOTE_TITLE\"" 2>&1) || {
    echo "[WARN] HTML body read failed: $HTML_BODY" >&2
    echo "[INFO] plaintext text approach is sufficient"
}
if [ -n "${HTML_BODY:-}" ]; then
    echo "HTML body length: ${#HTML_BODY} chars"
    echo "First 200 chars: ${HTML_BODY:0:200}"
fi
