# Quota Monitoring

## Overview

Passive monitoring of Claude token consumption within the rolling 5-hour quota window.
Reads JSONL session logs from `~/.claude/projects/` to estimate usage.

## Usage

```bash
# From workspace window
just quota

# With options
just quota --json              # Machine-readable JSON output
just quota --no-color          # No ANSI colors (for piping)
just quota --limit 220000     # Max 20x plan limit
just quota --window 3          # Custom window size (hours)
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `QUOTA_LIMIT` | `88000` | Estimated token limit (88K for Max 5x, 220K for Max 20x) |
| `QUOTA_WINDOW` | `5` | Rolling window size in hours |

## Color Levels

| Level | Threshold | Meaning |
|---|---|---|
| Green | < 70% | Normal operation |
| Yellow | 70-85% | Consider reducing agent activity |
| Red | > 85% | Risk of quota exhaustion — pause non-critical work |

## Token Counting

The script tracks **billed tokens** (input + output) for quota comparison. Cache read tokens
are excluded from the quota metric since they are free/heavily discounted. All token categories
are shown in the breakdown for visibility.

## Per-Agent Breakdown

Sessions are mapped to agents by their working directory:
- `~/.multiclaude/wts/ThreeDoors/<agent-name>/` → agent name
- `~/.multiclaude/repos/ThreeDoors` → `main-checkout`

## Automated Monitoring (Story 76.6)

### Cron-Based Monitoring

`scripts/quota-monitor.sh` wraps `quota-status.sh` with automated threshold evaluation,
warning deduplication, window reset detection, and snapshot recording.

**Recommended cron setup** (add to supervisor startup checklist):

```
CronCreate("*/5 * * * *", "bash scripts/quota-monitor.sh")
```

This runs every 5 minutes and:
1. Reads JSONL session logs via `quota-status.sh --json`
2. Evaluates usage against warning tiers (70%, 80%, 90%, 95%)
3. Sends `QUOTA_WARNING` to supervisor if a new tier is reached (deduplicated per window)
4. Sends `QUOTA_RESET` "all clear" when a window reset is detected
5. Records a usage snapshot to `~/.multiclaude/quota/snapshots.jsonl`
6. Updates dedup state at `~/.multiclaude/quota/state.json`

### Warning Tiers

| Tier | Threshold | Label | Recommended Action |
|---|---|---|---|
| Green | 70% | CAUTION | Monitor usage — consider deferring non-critical agents |
| Yellow | 80% | WARNING | Pause low-priority workers — focus on critical tasks |
| Orange | 90% | ALERT | Reduce to essential agents only — complete current work |
| Red | 95% | CRITICAL | Minimal activity only — save budget for urgent needs |

Warnings are advisory-only — they never block, throttle, or kill agents.

### Warning Deduplication

Within the same 5-hour window, a warning is only sent once per tier. If usage escalates
to a higher tier, a new warning is sent. On window reset, all tiers are cleared.

### State Files

| File | Purpose |
|---|---|
| `~/.multiclaude/quota/state.json` | Dedup state: last window start, last tier warned, last check time |
| `~/.multiclaude/quota/snapshots.jsonl` | Historical usage snapshots for retrospector analysis |

### CLI Options

```bash
scripts/quota-monitor.sh --dry-run          # Preview messages without sending
scripts/quota-monitor.sh --quiet            # Suppress stdout (cron-friendly)
scripts/quota-monitor.sh --limit 220000     # Max 20x plan
scripts/quota-monitor.sh --state-dir /tmp   # Custom state directory
```

## Limitations

- Token limits are community-estimated, not official Anthropic numbers
- The 5-hour window reset is approximate (based on earliest activity in window)
- Cache tokens may or may not count toward quota — the exact billing model is opaque
- No real-time streaming — run manually or via cron for periodic checks
