#!/usr/bin/env bash
# rollcall.sh — Quick overview of all running multiclaude agents
# Usage: ./scripts/rollcall.sh [--detail] [--repo REPO]
set -euo pipefail

# Parse arguments
DETAIL=false
REPO=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --detail) DETAIL=true; shift ;;
        --repo)   REPO="$2"; shift 2 ;;
        *)        echo "Usage: rollcall.sh [--detail] [--repo REPO]"; exit 1 ;;
    esac
done

# Auto-detect repo from state.json if not specified
STATE_FILE="$HOME/.multiclaude/state.json"
if [[ ! -f "$STATE_FILE" ]]; then
    echo "Error: state.json not found at $STATE_FILE"
    exit 1
fi

if ! command -v jq &>/dev/null; then
    echo "Error: jq required. Install with: brew install jq"
    exit 1
fi

if [[ -z "$REPO" ]]; then
    REPO=$(jq -r '.repos | keys[0] // empty' "$STATE_FILE")
    if [[ -z "$REPO" ]]; then
        echo "Error: no repos found in state.json. Specify with --repo NAME"
        exit 1
    fi
fi

SESSION="mc-${REPO}"
USER_NAME=$(whoami)
JSONL_DIR="$HOME/.claude/projects/-Users-${USER_NAME}--multiclaude-repos-${REPO}"

# Colors (disabled if stdout is not a terminal)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BOLD='\033[1m'
    DIM='\033[2m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BOLD='' DIM='' NC=''
fi

# Header
echo -e "${BOLD}${REPO} Rollcall — $(date '+%Y-%m-%d %H:%M %Z')${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Parse agents from state.json
agents=$(jq -r --arg repo "$REPO" '
    .repos[$repo].agents // {} | to_entries[] |
    [.key, (.value.type // "persistent"), (.value.pid // "0"), (.value.session_id // ""), (.value.created_at // ""), (.value.task // "")] |
    @tsv
' "$STATE_FILE" 2>/dev/null || true)

if [[ -z "$agents" ]]; then
    echo "No agents found for repo: $REPO"
    exit 0
fi

# Counters
persistent_count=0
worker_count=0
dead_count=0

# Table header
if $DETAIL; then
    printf " ${DIM}%-18s %-12s %-9s %7s %8s %7s %-8s %s${NC}\n" \
        "AGENT" "TYPE" "STATUS" "MEM(MB)" "JSONL" "CTX(K)" "UP" "TASK"
    echo "──────────────────────────────────────────────────────────────────────────────────────"
else
    printf " ${DIM}%-18s %-12s %-9s %8s %s${NC}\n" \
        "AGENT" "TYPE" "STATUS" "JSONL" "TASK"
    echo "──────────────────────────────────────────────────────────────────────────────────"
fi

while IFS=$'\t' read -r name atype pid session_id created_at task; do
    [[ -z "$name" ]] && continue

    # Status check via kill -0
    if [[ "$pid" != "0" ]] && kill -0 "$pid" 2>/dev/null; then
        status="${GREEN}● alive${NC}"
    else
        status="${RED}● dead${NC}"
        ((dead_count++))
    fi

    # Type label and counting
    case "$atype" in
        worker)     type_label="worker";     ((worker_count++)) ;;
        supervisor) type_label="supervisor"; ((persistent_count++)) ;;
        workspace)  type_label="workspace";  ((persistent_count++)) ;;
        *)          type_label="persistent"; ((persistent_count++)) ;;
    esac

    # JSONL file size
    jsonl_str="—"
    jsonl_path=""
    if [[ -n "$session_id" ]]; then
        jsonl_path="${JSONL_DIR}/${session_id}.jsonl"
        if [[ -f "$jsonl_path" ]]; then
            jsonl_bytes=$(stat -f%z "$jsonl_path" 2>/dev/null || stat -c%s "$jsonl_path" 2>/dev/null || echo 0)
            if [[ "$jsonl_bytes" -gt 0 ]]; then
                jsonl_mb=$(printf "%.1f" "$(echo "scale=2; $jsonl_bytes / 1048576" | bc 2>/dev/null)" 2>/dev/null || echo "?")
                jsonl_str="${jsonl_mb}MB"
            fi
        fi
    fi

    # Task summary (first line only, truncate to 35 chars)
    task_short=""
    if [[ -n "$task" ]]; then
        task_short=$(printf '%s' "$task" | tr '\n' ' ' | sed 's/\\n/ /g' | sed 's/^Research task: //' | sed 's/^\/implement-story /▶ /' | cut -c1-35)
    fi

    if $DETAIL; then
        # Memory (RSS in MB) — find the claude node process spawned under this agent
        mem_mb="—"
        if [[ "$pid" != "0" ]] && kill -0 "$pid" 2>/dev/null; then
            # The PID in state.json is the shell; claude is a child process
            # Look for a 'claude' child process via pgrep
            claude_pid=$(pgrep -P "$pid" 2>/dev/null | head -1 || true)
            target_pid="${claude_pid:-$pid}"
            # Sum RSS of the target and its children (node spawns workers)
            rss_kb=$(ps -o rss= -p "$target_pid" 2>/dev/null | tr -d ' ' || true)
            if [[ -n "$rss_kb" && "$rss_kb" -gt 100 ]]; then
                mem_mb=$((rss_kb / 1024))
            fi
        fi

        # Context estimate from last JSONL assistant message
        ctx_str="—"
        if [[ -f "${jsonl_path:-}" ]]; then
            ctx_str=$(python3 -c "
import json, sys
last_usage = None
try:
    with open(sys.argv[1]) as f:
        for line in f:
            try:
                d = json.loads(line)
            except json.JSONDecodeError:
                continue
            if d.get('type') == 'assistant' and 'message' in d:
                u = d['message'].get('usage', {})
                if u:
                    last_usage = u
    if last_usage:
        total = last_usage.get('input_tokens', 0) + last_usage.get('cache_read_input_tokens', 0)
        print(f'~{total // 1000}K')
    else:
        print('—')
except Exception:
    print('—')
" "$jsonl_path" 2>/dev/null || echo "—")
        fi

        # Uptime (extract time from created_at ISO timestamp)
        up_time="—"
        if [[ -n "$created_at" && "$created_at" != "null" ]]; then
            up_time=$(echo "$created_at" | sed 's/.*T//' | cut -c1-5)
        fi

        printf " %-18s %-12s %-19s %7s %8s %7s %-8s %s\n" \
            "$name" "$type_label" "$(echo -e "$status")" "$mem_mb" "$jsonl_str" "$ctx_str" "$up_time" "$task_short"
    else
        printf " %-18s %-12s %-19s %8s %s\n" \
            "$name" "$type_label" "$(echo -e "$status")" "$jsonl_str" "$task_short"
    fi
done <<< "$agents"

echo "──────────────────────────────────────────────────────────────────────────────────"

# Daemon status
daemon_status="${RED}stopped${NC}"
daemon_pid=$(pgrep -f "multiclaude.*daemon" 2>/dev/null | head -1 || true)
if [[ -n "$daemon_pid" ]]; then
    daemon_status="${GREEN}running${NC} (PID ${daemon_pid})"
fi

# Dead agent warning
dead_warning=""
if [[ "$dead_count" -gt 0 ]]; then
    dead_warning=" | ${RED}${dead_count} dead${NC}"
fi

echo -e " Daemon: ${daemon_status} | Agents: ${persistent_count} persistent + ${worker_count} workers${dead_warning} | Repo: ${REPO}"

if $DETAIL; then
    echo -e " ${DIM}CTX(K) = estimated context tokens from last JSONL message (input + cache_read) in thousands${NC}"
fi
