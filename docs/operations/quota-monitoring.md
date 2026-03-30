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

## Limitations

- Token limits are community-estimated, not official Anthropic numbers
- The 5-hour window reset is approximate (based on earliest activity in window)
- Cache tokens may or may not count toward quota — the exact billing model is opaque
- No real-time streaming — run manually or via cron for periodic checks
