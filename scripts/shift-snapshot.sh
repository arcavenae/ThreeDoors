#!/usr/bin/env bash
set -euo pipefail

# shift-snapshot.sh — Rolling State Snapshot Generator (Story 58.2)
# Generates a YAML snapshot of observable system state for supervisor shift handover.
# Designed to run every 5 minutes as part of the daemon refresh loop.
#
# Usage:
#   ./scripts/shift-snapshot.sh [--repo NAME] [--handover-dir DIR]
#
# Options:
#   --repo NAME          Repository name (default: auto-detected from multiclaude)
#   --handover-dir DIR   Override handover directory (default: ~/.multiclaude/handover/<repo>)
#   --help, -h           Show this help message
#
# Data sources (all external — no supervisor cooperation required):
#   - multiclaude worker list
#   - tmux list-windows -t mc-<repo>
#   - gh pr list --json number,title,statusCheckRollup
#
# Output: ~/.multiclaude/handover/<repo>/shift-state.yaml
#
# References:
#   - Story: docs/stories/58.2.story.md
#   - Research: _bmad-output/planning-artifacts/supervisor-shift-handover/

REPO_NAME=""
HANDOVER_DIR=""
MAX_SIZE_BYTES=10240  # 10KB

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_NAME="$2"; shift 2 ;;
        --handover-dir) HANDOVER_DIR="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: shift-snapshot.sh [--repo NAME] [--handover-dir DIR]"
            echo ""
            echo "Generates a rolling state snapshot for supervisor shift handover."
            echo ""
            echo "Options:"
            echo "  --repo NAME          Repository name (default: auto-detected)"
            echo "  --handover-dir DIR   Override handover directory"
            echo "  --help, -h           Show this help message"
            exit 0
            ;;
        *) echo "Error: Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Auto-detect repo name if not provided
if [[ -z "$REPO_NAME" ]]; then
    if command -v multiclaude &>/dev/null; then
        REPO_NAME="$(multiclaude repo current 2>/dev/null | grep -oE '[^ ]+$' || true)"
    fi
    if [[ -z "$REPO_NAME" ]]; then
        # Fallback: use basename of git remote
        REPO_NAME="$(basename "$(git remote get-url origin 2>/dev/null || echo 'unknown')" .git)"
    fi
fi

# Set handover directory
if [[ -z "$HANDOVER_DIR" ]]; then
    HANDOVER_DIR="$HOME/.multiclaude/handover/$REPO_NAME"
fi

# Task 1: Create handover directory structure
mkdir -p "$HANDOVER_DIR"
chmod 700 "$HANDOVER_DIR"

SNAPSHOT_FILE="$HANDOVER_DIR/shift-state.yaml"
TMP_FILE="$HANDOVER_DIR/.shift-state.yaml.tmp"
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# --- Collect observable state ---

# Task 2: Worker state collection
collect_workers() {
    local worker_output
    worker_output="$(multiclaude worker list 2>/dev/null || echo "")"

    echo "workers:"
    echo "  active:"

    if [[ -n "$worker_output" ]] && ! echo "$worker_output" | grep -qi "no.*workers\|no.*active\|^$"; then
        # Parse multiclaude worker list output
        # Expected format varies; extract name, task, branch from each line
        echo "$worker_output" | while IFS= read -r line; do
            # Skip header/empty lines
            [[ -z "$line" ]] && continue
            [[ "$line" =~ ^[[:space:]]*$ ]] && continue
            [[ "$line" =~ ^NAME|^---|^=== ]] && continue

            # Extract worker name (first column)
            local name
            name="$(echo "$line" | awk '{print $1}')"
            [[ -z "$name" ]] && continue

            echo "    - name: \"$name\""
            echo "      branch: \"work/$name\""

            # Try to extract task description if present (after name column)
            local task
            task="$(echo "$line" | sed "s/^[^ ]* *//" | sed 's/"/\\"/g')"
            if [[ -n "$task" ]]; then
                echo "      task: \"$task\""
            else
                echo "      task: \"\""
            fi

            # Check for PR on this worker's branch
            local pr_num
            pr_num="$(gh pr list --head "work/$name" --json number --jq '.[0].number // empty' 2>/dev/null || echo "")"
            if [[ -n "$pr_num" ]] && [[ "$pr_num" != "null" ]]; then
                echo "      pr: $pr_num"
            else
                echo "      pr: null"
            fi
        done
    fi

    echo "  recently_completed: []"
}

# Task 3: Persistent agent state collection
collect_agents() {
    local tmux_session="mc-$REPO_NAME"
    local windows_output

    echo "persistent_agents:"

    windows_output="$(tmux list-windows -t "$tmux_session" -F '#{window_name}' 2>/dev/null || echo "")"

    if [[ -n "$windows_output" ]]; then
        echo "$windows_output" | while IFS= read -r window_name; do
            [[ -z "$window_name" ]] && continue
            # Skip the supervisor window itself
            [[ "$window_name" == "supervisor" ]] && continue
            echo "  - name: \"$window_name\""
            echo "    status: \"active\""
        done
    fi
}

# Task 4: Open PR collection
collect_prs() {
    local pr_output

    echo "open_prs:"

    pr_output="$(gh pr list --json number,title,statusCheckRollup --limit 50 2>/dev/null || echo "")"

    if [[ -n "$pr_output" ]] && [[ "$pr_output" != "[]" ]]; then
        echo "$pr_output" | jq -r '.[] | "  - number: \(.number)\n    title: \"\(.title | gsub("\""; "\\\""))\"\n    ci: \"\(if .statusCheckRollup == null or (.statusCheckRollup | length) == 0 then "unknown" elif [.statusCheckRollup[] | select(.conclusion != "SUCCESS")] | length == 0 then "passing" elif [.statusCheckRollup[] | select(.conclusion == "FAILURE")] | length > 0 then "failing" else "pending" end)\""' 2>/dev/null || true
    fi
}

# Task 5: Assemble YAML snapshot
# Read existing supervisor delta sections if present
read_supervisor_delta() {
    if [[ ! -f "$SNAPSHOT_FILE" ]]; then
        return
    fi

    # Extract supervisor delta sections (pending_decisions through warnings)
    # These are written by the outgoing supervisor (Story 58.3), not by us
    local in_delta=false
    local delta_content=""

    while IFS= read -r line; do
        # Start capturing at supervisor delta sections
        if echo "$line" | grep -qE '^(pending_decisions|priorities|issue_triage|blockers|warnings):'; then
            in_delta=true
        fi
        # Stop capturing if we hit a new top-level section that's NOT a delta section
        if $in_delta && echo "$line" | grep -qE '^[a-z]' && ! echo "$line" | grep -qE '^(pending_decisions|priorities|issue_triage|blockers|warnings):'; then
            in_delta=false
        fi
        if $in_delta; then
            delta_content="${delta_content}${line}
"
        fi
    done < "$SNAPSHOT_FILE"

    if [[ -n "$delta_content" ]]; then
        printf '%s' "$delta_content"
    fi
}

# Build the snapshot
{
    echo "version: 1"
    echo "timestamp: \"$TIMESTAMP\""
    echo ""
    collect_workers
    echo ""
    collect_agents
    echo ""
    collect_prs
    echo ""

    # AC-8: Preserve existing supervisor delta sections
    delta="$(read_supervisor_delta)"
    if [[ -n "$delta" ]]; then
        printf '%s\n' "$delta"
    fi
} > "$TMP_FILE"

# Task 5 (cont): Atomic write — write to .tmp then rename
# sync ensures data is flushed to disk before rename
sync
mv "$TMP_FILE" "$SNAPSHOT_FILE"

# Task 6: Size monitoring
file_size="$(wc -c < "$SNAPSHOT_FILE" | tr -d ' ')"
if [[ "$file_size" -gt "$MAX_SIZE_BYTES" ]]; then
    echo "WARNING: shift-state.yaml exceeds 10KB size limit (actual: ${file_size} bytes)" >&2
fi

# Report success (for daemon log)
echo "Snapshot updated: $SNAPSHOT_FILE (${file_size} bytes)"
