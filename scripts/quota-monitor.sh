#!/usr/bin/env bash
set -euo pipefail

# quota-monitor.sh — Cron-Based Quota Monitoring with Window Reset Detection (Story 76.6)
#
# Designed to run every 5 minutes via CronCreate. Reads JSONL session logs,
# evaluates warning thresholds, deduplicates warnings within the same window/tier,
# detects window resets, and sends supervisor notifications.
#
# Usage:
#   ./scripts/quota-monitor.sh [OPTIONS]
#
# Options:
#   --limit TOKENS       Estimated token limit (default: $QUOTA_LIMIT or 88000)
#   --window HOURS       Rolling window size in hours (default: $QUOTA_WINDOW or 5)
#   --project-dir DIR    Claude project directory (default: auto-detect)
#   --state-dir DIR      State directory (default: ~/.multiclaude/quota)
#   --dry-run            Show what would happen without sending messages or writing state
#   --quiet              Suppress stdout except errors
#   --help, -h           Show this help message
#
# State file: ~/.multiclaude/quota/state.json
# Snapshot file: ~/.multiclaude/quota/snapshots.jsonl
#
# Cron setup (supervisor startup checklist):
#   CronCreate("*/5 * * * *", "bash scripts/quota-monitor.sh")
#
# References:
#   - Story: docs/stories/76.6.story.md
#   - Parser: scripts/quota-status.sh (Story 73.5)
#   - Research: _bmad-output/planning-artifacts/quota-throttling-research.md

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
QUOTA_STATUS_SCRIPT="$SCRIPT_DIR/quota-status.sh"

# --- Defaults ---
QUOTA_LIMIT="${QUOTA_LIMIT:-88000}"
QUOTA_WINDOW="${QUOTA_WINDOW:-5}"
PROJECT_DIR=""
STATE_DIR="${HOME}/.multiclaude/quota"
DRY_RUN=false
QUIET=false

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
        --state-dir)
            STATE_DIR="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --quiet)
            QUIET=true
            shift
            ;;
        --help|-h)
            awk '/^# quota-monitor/,/^[^#]/{if(/^#/) {sub(/^# ?/,""); print} else exit}' "$0"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# --- Helpers ---
log() {
    if [[ "$QUIET" == "false" ]]; then
        echo "[quota-monitor] $*"
    fi
}

send_message() {
    local msg="$1"
    if [[ "$DRY_RUN" == "true" ]]; then
        log "DRY-RUN: would send message: $msg"
    else
        multiclaude message send supervisor "$msg" 2>/dev/null || log "WARN: failed to send message to supervisor"
    fi
}

# --- Ensure state directory exists ---
mkdir -p "$STATE_DIR"

STATE_FILE="$STATE_DIR/state.json"
SNAPSHOT_FILE="$STATE_DIR/snapshots.jsonl"

# --- Get current quota data from quota-status.sh ---
QUOTA_ARGS=(--json --no-color --limit "$QUOTA_LIMIT" --window "$QUOTA_WINDOW")
if [[ -n "$PROJECT_DIR" ]]; then
    QUOTA_ARGS+=(--project-dir "$PROJECT_DIR")
fi

QUOTA_DATA=$("$QUOTA_STATUS_SCRIPT" "${QUOTA_ARGS[@]}" 2>/dev/null) || {
    log "ERROR: quota-status.sh failed"
    exit 1
}

if [[ -z "$QUOTA_DATA" || "$QUOTA_DATA" == "null" ]]; then
    log "ERROR: empty quota data"
    exit 1
fi

# --- Export vars for python3 ---
export QUOTA_DATA
export STATE_FILE
export SNAPSHOT_FILE
export DRY_RUN
export QUIET
export QUOTA_LIMIT
export QUOTA_WINDOW

# --- Process with python3 ---
# Handles: window detection, reset detection, threshold evaluation, dedup, snapshot, state update
python3 << 'PYEOF'
import json
import os
import sys
from datetime import datetime, timezone

# Read inputs
quota_data = json.loads(os.environ["QUOTA_DATA"])
state_file = os.environ["STATE_FILE"]
snapshot_file = os.environ["SNAPSHOT_FILE"]
dry_run = os.environ.get("DRY_RUN", "false") == "true"
quiet = os.environ.get("QUIET", "false") == "true"
quota_limit = int(os.environ.get("QUOTA_LIMIT", "88000"))
quota_window = int(os.environ.get("QUOTA_WINDOW", "5"))

def log(msg):
    if not quiet:
        print(f"[quota-monitor] {msg}", flush=True)

# --- Warning tiers (from R-004 / Story 76.3) ---
# tier_name, threshold_pct, label, recommended_action
TIERS = [
    ("green",  70, "CAUTION",  "Monitor usage — consider deferring non-critical agents"),
    ("yellow", 80, "WARNING",  "Pause low-priority workers — focus on critical tasks"),
    ("orange", 90, "ALERT",    "Reduce to essential agents only — complete current work"),
    ("red",    95, "CRITICAL", "Minimal activity only — save budget for urgent needs"),
]

def get_current_tier(usage_pct):
    """Return the highest tier that usage_pct exceeds, or None if below all."""
    matched = None
    for name, threshold, label, action in TIERS:
        if usage_pct >= threshold:
            matched = (name, threshold, label, action)
    return matched

def tier_rank(tier_name):
    """Return numeric rank for tier comparison. Higher = more severe."""
    ranks = {"green": 1, "yellow": 2, "orange": 3, "red": 4}
    return ranks.get(tier_name, 0)

# --- Read existing state ---
state = {}
if os.path.exists(state_file):
    try:
        with open(state_file, "r") as f:
            state = json.load(f)
    except (json.JSONDecodeError, OSError):
        log("WARN: corrupt state file, starting fresh")
        state = {}

prev_window_start = state.get("window_start_epoch")
prev_tier = state.get("last_tier_warned")
prev_check = state.get("last_check_epoch")

# --- Extract current window info ---
now_epoch = quota_data.get("now_epoch", 0)
earliest_ts = quota_data.get("earliest_ts")
usage_pct = quota_data.get("usage_pct", 0)
total_billed = quota_data.get("total_billed", 0)
window_reset_epoch = quota_data.get("window_reset_epoch")

# Current window start is the earliest activity timestamp in the window
current_window_start = earliest_ts

# --- Window reset detection ---
window_reset_detected = False
messages_to_send = []

if prev_window_start is not None and current_window_start is not None:
    # A reset occurred if:
    # 1. The current window start is different from previous (new window)
    # 2. AND the previous window should have expired by now
    prev_reset_time = prev_window_start + (quota_window * 3600)
    if current_window_start != prev_window_start and now_epoch >= prev_reset_time:
        window_reset_detected = True
        window_start_str = datetime.fromtimestamp(
            current_window_start, tz=timezone.utc
        ).strftime("%H:%M UTC")
        messages_to_send.append(
            f"QUOTA_RESET: Window reset detected. New window started at {window_start_str}. Budget: 100%."
        )
        log(f"Window reset detected. New window at {window_start_str}.")
elif prev_window_start is not None and current_window_start is None:
    # No activity in current window — previous window expired
    if now_epoch >= prev_window_start + (quota_window * 3600):
        window_reset_detected = True
        messages_to_send.append(
            "QUOTA_RESET: Previous quota window expired. No current activity. Budget: 100%."
        )
        log("Previous window expired, no current activity.")

# --- Threshold evaluation ---
current_tier = get_current_tier(usage_pct)
should_warn = False

if current_tier is not None:
    tier_name, threshold, label, action = current_tier

    if window_reset_detected:
        # After a reset, always clear previous tier state
        should_warn = True
    elif prev_window_start is not None and current_window_start == prev_window_start:
        # Same window — only warn if tier escalated
        if prev_tier is None or tier_rank(tier_name) > tier_rank(prev_tier):
            should_warn = True
        else:
            log(f"Dedup: already warned for tier {prev_tier} in this window (current: {tier_name})")
    else:
        # New window or first run
        should_warn = True

    if should_warn:
        # Calculate remaining tokens and time
        remaining_tokens = max(0, quota_limit - total_billed)
        remaining_time_str = "unknown"
        if window_reset_epoch and now_epoch:
            remaining_secs = window_reset_epoch - now_epoch
            if remaining_secs > 0:
                hrs = int(remaining_secs // 3600)
                mins = int((remaining_secs % 3600) // 60)
                remaining_time_str = f"{hrs}h {mins}m"
            else:
                remaining_time_str = "expired"

        messages_to_send.append(
            f"QUOTA_WARNING [{label}]: Usage at {usage_pct:.1f}% "
            f"({total_billed:,}/{quota_limit:,} tokens). "
            f"Remaining: ~{remaining_tokens:,} tokens, {remaining_time_str} until reset. "
            f"Recommendation: {action}"
        )
        log(f"Threshold {label} ({threshold}%) triggered at {usage_pct:.1f}%")
else:
    log(f"Usage at {usage_pct:.1f}% — below all warning thresholds")

# --- Send messages ---
for msg in messages_to_send:
    if dry_run:
        log(f"DRY-RUN: would send: {msg}")
    else:
        # Write to stdout for cron to pick up, and try multiclaude message
        log(f"Sending: {msg}")
        try:
            os.system(f'multiclaude message send supervisor "{msg}"')
        except Exception as e:
            log(f"WARN: failed to send message: {e}")

# --- Record snapshot ---
snapshot = {
    "timestamp": datetime.fromtimestamp(now_epoch, tz=timezone.utc).isoformat(),
    "usage_pct": usage_pct,
    "total_billed": total_billed,
    "quota_limit": quota_limit,
    "window_start_epoch": current_window_start,
    "window_reset_epoch": window_reset_epoch,
    "tier": current_tier[0] if current_tier else "none",
    "window_reset_detected": window_reset_detected,
    "warning_sent": should_warn and current_tier is not None,
    "agents": {
        name: {
            "input_tokens": stats.get("input_tokens", 0),
            "output_tokens": stats.get("output_tokens", 0),
            "interactions": stats.get("interactions", 0),
        }
        for name, stats in quota_data.get("agents", {}).items()
    },
}

# Determine peak/off-peak (peak = 05:00-11:00 PT = 12:00-18:00 UTC)
utc_hour = datetime.fromtimestamp(now_epoch, tz=timezone.utc).hour
snapshot["peak_hours"] = 12 <= utc_hour < 18

try:
    with open(snapshot_file, "a") as f:
        f.write(json.dumps(snapshot) + "\n")
    log("Snapshot recorded")
except OSError as e:
    log(f"WARN: failed to write snapshot: {e}")

# --- Update state ---
new_state = {
    "window_start_epoch": current_window_start,
    "last_tier_warned": current_tier[0] if (should_warn and current_tier) else (
        None if window_reset_detected else prev_tier
    ),
    "last_check_epoch": now_epoch,
    "last_usage_pct": usage_pct,
}

try:
    # Atomic write: write to tmp, then rename
    tmp_file = state_file + ".tmp"
    with open(tmp_file, "w") as f:
        json.dump(new_state, f, indent=2)
        f.write("\n")
        f.flush()
        os.fsync(f.fileno())
    os.rename(tmp_file, state_file)
    log("State updated")
except OSError as e:
    log(f"WARN: failed to write state: {e}")

log("Monitor complete")
PYEOF
