#!/usr/bin/env bash
set -euo pipefail

# quota-status.sh — Passive Quota Monitoring (Story 73.5)
# Reads Claude Code JSONL session logs to estimate token consumption
# within the rolling 5-hour quota window.
#
# Usage:
#   ./scripts/quota-status.sh [OPTIONS]
#
# Options:
#   --limit TOKENS       Estimated token limit (default: $QUOTA_LIMIT or 88000)
#   --window HOURS       Rolling window size in hours (default: $QUOTA_WINDOW or 5)
#   --project-dir DIR    Claude project directory (default: auto-detect)
#   --json               Output in JSON format
#   --no-color           Disable color output
#   --help, -h           Show this help message
#
# Environment:
#   QUOTA_LIMIT          Estimated token limit (88000 for Max 5x, 220000 for Max 20x)
#   QUOTA_WINDOW         Rolling window size in hours (default: 5)
#
# References:
#   - Story: docs/stories/73.5.story.md
#   - Research: _bmad-output/planning-artifacts/quota-throttling-research.md

# --- Defaults ---
QUOTA_LIMIT="${QUOTA_LIMIT:-88000}"
QUOTA_WINDOW="${QUOTA_WINDOW:-5}"
PROJECT_DIR=""
OUTPUT_JSON=false
USE_COLOR=true

# --- Colors (used in embedded Python script below) ---
# shellcheck disable=SC2034
RED='\033[0;31m'
# shellcheck disable=SC2034
ORANGE='\033[0;91m'
# shellcheck disable=SC2034
YELLOW='\033[0;33m'
# shellcheck disable=SC2034
GREEN='\033[0;32m'
# shellcheck disable=SC2034
BOLD='\033[1m'
# shellcheck disable=SC2034
DIM='\033[2m'
# shellcheck disable=SC2034
RESET='\033[0m'

# --- Argument parsing ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --limit)
            QUOTA_LIMIT="$2"
            shift 2
            ;;
        --window)
            QUOTA_WINDOW="$2"
            shift 2
            ;;
        --project-dir)
            PROJECT_DIR="$2"
            shift 2
            ;;
        --json)
            OUTPUT_JSON=true
            shift
            ;;
        --no-color)
            USE_COLOR=false
            shift
            ;;
        --help|-h)
            awk '/^# quota-status/,/^[^#]/{if(/^#/) {sub(/^# ?/,""); print} else exit}' "$0"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# shellcheck disable=SC2034
if [[ "$USE_COLOR" == "false" ]]; then
    # shellcheck disable=SC2034
    RED=''
    # shellcheck disable=SC2034
    ORANGE=''
    # shellcheck disable=SC2034
    YELLOW=''
    # shellcheck disable=SC2034
    GREEN=''
    # shellcheck disable=SC2034
    BOLD=''
    # shellcheck disable=SC2034
    DIM=''
    # shellcheck disable=SC2034
    RESET=''
fi

# --- Auto-detect project directory ---
if [[ -z "$PROJECT_DIR" ]]; then
    CLAUDE_DIR="$HOME/.claude/projects"
    if [[ ! -d "$CLAUDE_DIR" ]]; then
        echo "Error: Claude projects directory not found at $CLAUDE_DIR" >&2
        exit 1
    fi
    # Find all ThreeDoors project dirs (main checkout + worktrees)
    PROJECT_DIRS=()
    while IFS= read -r dir; do
        PROJECT_DIRS+=("$dir")
    done < <(find "$CLAUDE_DIR" -maxdepth 1 -type d -name '*ThreeDoors*' 2>/dev/null | sort)

    if [[ ${#PROJECT_DIRS[@]} -eq 0 ]]; then
        echo "Error: No ThreeDoors project directories found" >&2
        exit 1
    fi
else
    # If project-dir is given, scan its subdirectories (each is a "project")
    if [[ -d "$PROJECT_DIR" ]]; then
        PROJECT_DIRS=()
        # Check if the dir itself contains JSONL files
        if compgen -G "$PROJECT_DIR/*.jsonl" > /dev/null 2>&1; then
            PROJECT_DIRS+=("$PROJECT_DIR")
        fi
        # Also scan subdirectories
        while IFS= read -r dir; do
            PROJECT_DIRS+=("$dir")
        done < <(find "$PROJECT_DIR" -maxdepth 1 -mindepth 1 -type d 2>/dev/null | sort)
        if [[ ${#PROJECT_DIRS[@]} -eq 0 ]]; then
            PROJECT_DIRS=("$PROJECT_DIR")
        fi
    else
        echo "Error: Project directory not found: $PROJECT_DIR" >&2
        exit 1
    fi
fi

# --- Calculate window boundaries ---
NOW_EPOCH=$(date +%s)
WINDOW_SECONDS=$((QUOTA_WINDOW * 3600))
WINDOW_START_EPOCH=$((NOW_EPOCH - WINDOW_SECONDS))

# --- Parse JSONL files with python3 (jq alternative that's always available on macOS) ---
# Collect all token data within the window across all project dirs
parse_tokens() {
    python3 << 'PYEOF'
import json
import os
import sys
import glob
from datetime import datetime, timezone

project_dirs = os.environ.get("PROJECT_DIRS_LIST", "").split("\n")
window_start = int(os.environ.get("WINDOW_START_EPOCH", "0"))
now_epoch = int(os.environ.get("NOW_EPOCH", "0"))

# Collect per-agent stats
# agent_name -> {input_tokens, output_tokens, cache_creation, cache_read, first_ts, last_ts, interactions}
agents = {}
total_input = 0
total_output = 0
total_cache_creation = 0
total_cache_read = 0
earliest_ts = None
latest_ts = None
interaction_count = 0

for proj_dir in project_dirs:
    proj_dir = proj_dir.strip()
    if not proj_dir or not os.path.isdir(proj_dir):
        continue

    # Determine agent name from directory
    dirname = os.path.basename(proj_dir)
    if "multiclaude-wts-ThreeDoors-" in dirname:
        agent_name = dirname.split("multiclaude-wts-ThreeDoors-")[-1]
    elif "multiclaude-repos-ThreeDoors" in dirname:
        agent_name = "main-checkout"
    else:
        agent_name = dirname

    jsonl_files = glob.glob(os.path.join(proj_dir, "*.jsonl"))
    for jf in jsonl_files:
        try:
            # Quick check: skip files not modified in the window
            mtime = os.path.getmtime(jf)
            if mtime < window_start:
                continue

            with open(jf, "r") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    try:
                        obj = json.loads(line)
                    except json.JSONDecodeError:
                        continue

                    # Get timestamp
                    ts_str = obj.get("timestamp")
                    if not ts_str:
                        continue

                    # Parse ISO timestamp
                    try:
                        ts = datetime.fromisoformat(ts_str.replace("Z", "+00:00"))
                        ts_epoch = ts.timestamp()
                    except (ValueError, AttributeError):
                        continue

                    if ts_epoch < window_start:
                        continue

                    # Extract usage from message
                    msg = obj.get("message", {})
                    if not isinstance(msg, dict):
                        continue
                    usage = msg.get("usage")
                    if not usage:
                        continue

                    inp = usage.get("input_tokens", 0)
                    out = usage.get("output_tokens", 0)
                    cache_create = usage.get("cache_creation_input_tokens", 0)
                    cache_rd = usage.get("cache_read_input_tokens", 0)

                    total_input += inp
                    total_output += out
                    total_cache_creation += cache_create
                    total_cache_read += cache_rd
                    interaction_count += 1

                    if earliest_ts is None or ts_epoch < earliest_ts:
                        earliest_ts = ts_epoch
                    if latest_ts is None or ts_epoch > latest_ts:
                        latest_ts = ts_epoch

                    if agent_name not in agents:
                        agents[agent_name] = {
                            "input_tokens": 0,
                            "output_tokens": 0,
                            "cache_creation": 0,
                            "cache_read": 0,
                            "first_ts": ts_epoch,
                            "last_ts": ts_epoch,
                            "interactions": 0,
                        }
                    a = agents[agent_name]
                    a["input_tokens"] += inp
                    a["output_tokens"] += out
                    a["cache_creation"] += cache_create
                    a["cache_read"] += cache_rd
                    a["interactions"] += 1
                    if ts_epoch < a["first_ts"]:
                        a["first_ts"] = ts_epoch
                    if ts_epoch > a["last_ts"]:
                        a["last_ts"] = ts_epoch

        except (OSError, IOError):
            continue

# Calculate window reset
window_reset_epoch = None
if earliest_ts:
    window_reset_epoch = earliest_ts + int(os.environ.get("WINDOW_SECONDS", "18000"))

result = {
    "total_input": total_input,
    "total_output": total_output,
    "total_cache_creation": total_cache_creation,
    "total_cache_read": total_cache_read,
    "total_billed": total_input + total_output,
    "total_all": total_input + total_output + total_cache_creation + total_cache_read,
    "interaction_count": interaction_count,
    "earliest_ts": earliest_ts,
    "latest_ts": latest_ts,
    "window_reset_epoch": window_reset_epoch,
    "now_epoch": now_epoch,
    "agents": agents,
}

print(json.dumps(result))
PYEOF
}

# Export vars for python
export PROJECT_DIRS_LIST
PROJECT_DIRS_LIST=$(printf '%s\n' "${PROJECT_DIRS[@]}")
export WINDOW_START_EPOCH="$WINDOW_START_EPOCH"
export NOW_EPOCH="$NOW_EPOCH"
export WINDOW_SECONDS="$WINDOW_SECONDS"

TOKEN_DATA=$(parse_tokens)

if [[ -z "$TOKEN_DATA" || "$TOKEN_DATA" == "null" ]]; then
    echo "Error: Failed to parse token data" >&2
    exit 1
fi

# --- Output ---
if [[ "$OUTPUT_JSON" == "true" ]]; then
    # Add quota metadata to JSON output
    python3 -c "
import json, sys
from datetime import datetime, timezone
try:
    import zoneinfo
    pt = zoneinfo.ZoneInfo('America/Los_Angeles')
except Exception:
    pt = None

data = json.loads(sys.stdin.read())
data['quota_limit'] = int('$QUOTA_LIMIT')
data['quota_window_hours'] = int('$QUOTA_WINDOW')
total = data['total_billed']
limit = data['quota_limit']
data['usage_pct'] = round(total / limit * 100, 1) if limit > 0 else 0
if data['usage_pct'] < 70:
    data['level'] = 'green'
elif data['usage_pct'] < 80:
    data['level'] = 'yellow'
elif data['usage_pct'] < 90:
    data['level'] = 'orange'
else:
    data['level'] = 'red'

# Peak/off-peak
now_utc = datetime.fromtimestamp(data['now_epoch'], tz=timezone.utc)
if pt:
    now_pt = now_utc.astimezone(pt)
    data['is_peak'] = 5 <= now_pt.hour < 11
    data['peak_time_pt'] = now_pt.strftime('%H:%M')
else:
    data['is_peak'] = None
    data['peak_time_pt'] = None

# Priority tiers for agents
persistent = {'supervisor','merge-queue','pr-shepherd','arch-watchdog','envoy','project-watchdog','retrospector','main-checkout'}
for name in data.get('agents', {}):
    lower = name.lower()
    if lower in persistent:
        data['agents'][name]['priority_tier'] = 'P1'
    elif '-' in lower and lower not in persistent:
        data['agents'][name]['priority_tier'] = 'P2'
    else:
        data['agents'][name]['priority_tier'] = 'P3'

print(json.dumps(data, indent=2))
" <<< "$TOKEN_DATA"
    exit 0
fi

# --- Formatted output ---
# Use python for all formatting to avoid shell quoting issues with JSON
python3 << FMTEOF
import json
import sys
from datetime import datetime, timezone

data = json.loads('''$TOKEN_DATA''')
limit = $QUOTA_LIMIT
window_h = $QUOTA_WINDOW
use_color = "$USE_COLOR" == "true"

# Colors
if use_color:
    RED = '\033[0;31m'
    ORANGE = '\033[0;91m'
    YELLOW = '\033[0;33m'
    GREEN = '\033[0;32m'
    BOLD = '\033[1m'
    DIM = '\033[2m'
    RESET = '\033[0m'
else:
    RED = ORANGE = YELLOW = GREEN = BOLD = DIM = RESET = ''

total = data['total_billed']
total_all = data['total_all']
pct = (total / limit * 100) if limit > 0 else 0

if pct < 70:
    level_color = GREEN
    level_label = "OK"
elif pct < 80:
    level_color = YELLOW
    level_label = "ELEVATED"
elif pct < 90:
    level_color = ORANGE
    level_label = "WARNING"
else:
    level_color = RED
    level_label = "CRITICAL"

# Peak/off-peak detection (Anthropic peak: 05:00-11:00 PT)
from datetime import timezone as tz
import zoneinfo
try:
    pt = zoneinfo.ZoneInfo("America/Los_Angeles")
except Exception:
    pt = None
now_utc = datetime.fromtimestamp(data['now_epoch'], tz=tz.utc)
if pt:
    now_pt = now_utc.astimezone(pt)
    pt_hour = now_pt.hour
    is_peak = 5 <= pt_hour < 11
    peak_str = f"{RED}PEAK HOURS{RESET} (05:00-11:00 PT)" if is_peak else f"{GREEN}Off-Peak{RESET}"
    peak_time_str = now_pt.strftime('%H:%M PT')
else:
    is_peak = False
    peak_str = f"{DIM}unknown (timezone unavailable){RESET}"
    peak_time_str = now_utc.strftime('%H:%M UTC')

# Empty state
if data['interaction_count'] == 0:
    print(f"{BOLD}══════════════════════════════════════════════{RESET}")
    print(f"{BOLD}  Claude Quota Status{RESET}")
    print(f"{BOLD}══════════════════════════════════════════════{RESET}")
    print()
    print(f"  {YELLOW}No usage data found.{RESET}")
    print()
    print(f"  No JSONL session logs were found in the current quota window.")
    print(f"  This may mean:")
    print(f"    - No Claude sessions have been active in the last {window_h}h")
    print(f"    - The project directory is not configured correctly")
    print()
    print(f"  Check your Claude config directory:")
    print(f"    ~/.claude/projects/")
    print()
    print(f"  {BOLD}Peak Status:{RESET} {peak_str} ({peak_time_str})")
    print()
    print(f"{BOLD}══════════════════════════════════════════════{RESET}")
    sys.exit(0)

# Header
print(f"{BOLD}══════════════════════════════════════════════{RESET}")
print(f"{BOLD}  Claude Quota Status{RESET}")
print(f"{BOLD}══════════════════════════════════════════════{RESET}")
print()

# Overall usage
bar_width = 30
filled = int(bar_width * min(pct, 100) / 100)
bar = '█' * filled + '░' * (bar_width - filled)
print(f"  {BOLD}Usage:{RESET}  {level_color}{total:,}{RESET} / {limit:,} billed tokens ({level_color}{pct:.1f}%{RESET})")
print(f"  {BOLD}Level:{RESET}  {level_color}{level_label}{RESET}")
print(f"  {BOLD}Bar:{RESET}    [{level_color}{bar}{RESET}]")
print(f"  {BOLD}Peak:{RESET}   {peak_str} ({peak_time_str})")
print()

# Token breakdown
print(f"  {BOLD}Breakdown:{RESET}")
print(f"    Input tokens:          {data['total_input']:>10,}")
print(f"    Output tokens:         {data['total_output']:>10,}")
print(f"    Cache creation tokens: {data['total_cache_creation']:>10,}")
print(f"    Cache read tokens:     {data['total_cache_read']:>10,}")
print(f"    Interactions:          {data['interaction_count']:>10,}")
print()

# Window timing
if data['earliest_ts']:
    earliest = datetime.fromtimestamp(data['earliest_ts'], tz=timezone.utc)
    now = datetime.fromtimestamp(data['now_epoch'], tz=timezone.utc)
    reset_epoch = data.get('window_reset_epoch')
    if reset_epoch:
        reset_dt = datetime.fromtimestamp(reset_epoch, tz=timezone.utc)
        remaining = reset_epoch - data['now_epoch']
        if remaining > 0:
            hrs = int(remaining // 3600)
            mins = int((remaining % 3600) // 60)
            remaining_str = f"{hrs}h {mins}m"
        else:
            remaining_str = f"{RED}EXPIRED{RESET}"
    else:
        remaining_str = "unknown"

    print(f"  {BOLD}Window:{RESET}")
    print(f"    Window size:    {window_h}h")
    print(f"    First activity: {earliest.strftime('%H:%M:%S UTC')}")
    print(f"    Window resets:  {reset_dt.strftime('%H:%M:%S UTC') if reset_epoch else 'unknown'}")
    print(f"    Time remaining: {remaining_str}")
    print()

# Priority tier assignment
# P1: persistent agents (supervisor, merge-queue, pr-shepherd, etc.)
# P2: implementation workers
# P3: ephemeral/one-shot agents
PERSISTENT_AGENTS = {
    'supervisor', 'merge-queue', 'pr-shepherd', 'arch-watchdog',
    'envoy', 'project-watchdog', 'retrospector', 'main-checkout',
}

def get_priority_tier(name):
    lower = name.lower()
    if lower in PERSISTENT_AGENTS:
        return 'P1'
    # Workers follow naming patterns like bold-fox, clever-owl
    if any(c in lower for c in ['-']) and lower not in PERSISTENT_AGENTS:
        return 'P2'
    return 'P3'

# Per-agent breakdown
agents = data.get('agents', {})
if agents:
    print(f"  {BOLD}Per-Agent Consumption:{RESET}")
    print(f"    {'Agent':<22} {'Tier':>4} {'Tokens':>10} {'Pct':>6}  {'Interactions':>6}")
    print(f"    {'─' * 22} {'─' * 4} {'─' * 10} {'─' * 6}  {'─' * 6}")

    sorted_agents = sorted(agents.items(), key=lambda x: (
        x[1]['input_tokens'] + x[1]['output_tokens']
    ), reverse=True)

    for name, stats in sorted_agents:
        agent_total = stats['input_tokens'] + stats['output_tokens']
        agent_pct = (agent_total / total * 100) if total > 0 else 0
        tier = get_priority_tier(name)
        # Truncate long names
        display_name = name[:22] if len(name) <= 22 else name[:19] + '...'
        print(f"    {display_name:<22} {tier:>4} {agent_total:>10,} {agent_pct:>5.1f}%  {stats['interactions']:>6}")

    # Identify highest burn rate
    if sorted_agents:
        top_name, top_stats = sorted_agents[0]
        duration = top_stats['last_ts'] - top_stats['first_ts']
        if duration > 0:
            top_total = top_stats['input_tokens'] + top_stats['output_tokens']
            rate = top_total / (duration / 3600)
            print()
            print(f"    {DIM}Highest burn rate: {top_name} @ {rate:,.0f} tokens/hr{RESET}")

print()
print(f"{BOLD}══════════════════════════════════════════════{RESET}")
FMTEOF
