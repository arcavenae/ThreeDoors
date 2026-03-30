#!/usr/bin/env bash
set -euo pipefail

# sm-sprint-health.sh — SM Sprint Health Check
# Queries open PRs for staleness, checks for blocked stories,
# summarizes worker status, and reports risks to supervisor.
#
# Usage:
#   ./scripts/sm-sprint-health.sh [--repo OWNER/REPO] [--stale-hours 24] [--stories-dir docs/stories]
#
# Designed to run every 4 hours via:
#   /loop 4h /bmad-bmm-sprint-status
# Or directly via cron:
#   0 */4 * * * cd /path/to/ThreeDoors && ./scripts/sm-sprint-health.sh

REPO="${REPO:-arcavenae/ThreeDoors}"
STALE_HOURS="${STALE_HOURS:-24}"
STORIES_DIR="${STORIES_DIR:-docs/stories}"
REPORT_TO_SUPERVISOR="${REPORT_TO_SUPERVISOR:-false}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO="$2"; shift 2 ;;
        --stale-hours) STALE_HOURS="$2"; shift 2 ;;
        --stories-dir) STORIES_DIR="$2"; shift 2 ;;
        --report) REPORT_TO_SUPERVISOR="true"; shift ;;
        --help|-h)
            echo "Usage: sm-sprint-health.sh [--repo OWNER/REPO] [--stale-hours N] [--stories-dir DIR] [--report]"
            echo ""
            echo "Options:"
            echo "  --repo          GitHub repo (default: arcavenae/ThreeDoors)"
            echo "  --stale-hours   Hours without activity before PR is stale (default: 24)"
            echo "  --stories-dir   Path to story files (default: docs/stories)"
            echo "  --report        Send report to supervisor via multiclaude message send"
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Check for required tools
for cmd in gh jq; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: $cmd is required but not installed." >&2
        exit 1
    fi
done

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
RISKS_FOUND=0
REPORT=""

add_to_report() {
    REPORT="${REPORT}${1}\n"
}

add_to_report "# SM Sprint Health Report"
add_to_report "Timestamp: ${TIMESTAMP}"
add_to_report ""

# --- 1. Check for stale PRs (>STALE_HOURS without activity) ---
add_to_report "## Open PRs"

STALE_CUTOFF="$(date -u -v-"${STALE_HOURS}"H +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d "${STALE_HOURS} hours ago" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "")"

OPEN_PRS="$(gh pr list --repo "$REPO" --state open --json number,title,updatedAt,author --limit 50 2>/dev/null || echo "[]")"
PR_COUNT="$(echo "$OPEN_PRS" | jq 'length')"
add_to_report "Total open: ${PR_COUNT}"

if [[ -n "$STALE_CUTOFF" ]]; then
    STALE_PRS="$(echo "$OPEN_PRS" | jq --arg cutoff "$STALE_CUTOFF" '[.[] | select(.updatedAt < $cutoff)]')"
    STALE_COUNT="$(echo "$STALE_PRS" | jq 'length')"

    if [[ "$STALE_COUNT" -gt 0 ]]; then
        RISKS_FOUND=1
        add_to_report "RISK: ${STALE_COUNT} stale PR(s) (>${STALE_HOURS}h without activity):"
        STALE_LINES="$(echo "$STALE_PRS" | jq -r '.[] | "  - #\(.number): \(.title) (last updated: \(.updatedAt))"')"
        add_to_report "$STALE_LINES"
    else
        add_to_report "No stale PRs found."
    fi
else
    add_to_report "Warning: Could not compute stale cutoff date. Skipping staleness check."
fi

add_to_report ""

# --- 2. Check for blocked stories ---
add_to_report "## Blocked Stories"

if [[ -d "$STORIES_DIR" ]]; then
    BLOCKED_COUNT=0
    while IFS= read -r story_file; do
        if grep -qi "status:.*blocked" "$story_file" 2>/dev/null; then
            BLOCKED_COUNT=$((BLOCKED_COUNT + 1))
            STORY_NAME="$(basename "$story_file" .story.md)"
            RISKS_FOUND=1
            add_to_report "RISK: Story ${STORY_NAME} is blocked"
        fi
    done < <(find "$STORIES_DIR" -name "*.story.md" -type f 2>/dev/null)

    if [[ "$BLOCKED_COUNT" -eq 0 ]]; then
        add_to_report "No blocked stories found."
    fi
else
    add_to_report "Warning: Stories directory '${STORIES_DIR}' not found."
fi

add_to_report ""

# --- 3. Worker status ---
add_to_report "## Worker Status"

if command -v multiclaude &>/dev/null; then
    WORKER_STATUS="$(multiclaude worker list 2>/dev/null || echo "Could not query worker status.")"
    add_to_report "$WORKER_STATUS"
else
    add_to_report "multiclaude not available — skipping worker status check."
fi

add_to_report ""

# --- 4. Summary ---
add_to_report "## Summary"

if [[ "$RISKS_FOUND" -eq 1 ]]; then
    add_to_report "Risks detected — review items above."
else
    add_to_report "All clear — no risks detected."
fi

# Output report
echo -e "$REPORT"

# Optionally send to supervisor
if [[ "$REPORT_TO_SUPERVISOR" == "true" ]] && command -v multiclaude &>/dev/null; then
    echo -e "$REPORT" | multiclaude message send --to supervisor --stdin 2>/dev/null || \
        echo "Warning: Could not send report to supervisor." >&2
fi

exit 0
