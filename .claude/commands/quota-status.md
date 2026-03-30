# /quota-status - Show Claude quota usage status

Display current Claude API quota consumption within the rolling usage window.

## Instructions

Run the quota status script and display the results:

```bash
./scripts/quota-status.sh --no-color
```

Present the output to the user as-is. The script handles all formatting, threshold detection, and per-agent breakdown.

### Options

If the user requests JSON output:
```bash
./scripts/quota-status.sh --json
```

If the user specifies a custom token limit (e.g., for Max 20x plan):
```bash
./scripts/quota-status.sh --limit 220000 --no-color
```

If the user specifies a custom window size:
```bash
./scripts/quota-status.sh --window 5 --no-color
```

### Environment Variables

- `QUOTA_LIMIT` — Estimated token limit (default: 88000 for Max 5x, use 220000 for Max 20x)
- `QUOTA_WINDOW` — Rolling window size in hours (default: 5)

### What it shows

- **Usage summary**: Current window token usage (absolute + percentage) with color-coded status
- **Threshold levels**: GREEN (<70%), YELLOW (70-80%), ORANGE (80-90%), RED (>90%)
- **Peak hours**: Whether we're in Anthropic peak hours (05:00-11:00 PT)
- **Token breakdown**: Input, output, cache creation, cache read tokens
- **Window timing**: Time since window start, estimated reset time
- **Per-agent consumption**: Sorted table showing each agent's token usage, percentage, priority tier, and interaction count

### Troubleshooting

If no data is shown:
1. Check that Claude sessions exist in `~/.claude/projects/`
2. Verify JSONL files are present and recent
3. Try increasing the window: `./scripts/quota-status.sh --window 24 --no-color`
