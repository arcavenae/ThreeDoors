# Persistent Agent Operations Guide

Operational learnings for running persistent multiclaude agents (merge-queue, pr-shepherd, etc.) in long-running sessions.

---

## Known Issue: Context Window Exhaustion Lockup

**First observed:** 2026-03-09
**Agents affected:** merge-queue, pr-shepherd
**Severity:** High — causes silent agent failure with no automatic recovery

### Symptoms

- Agent stops responding to messages and PR events
- No error output — agent simply ceases generating tool calls
- tmux window shows agent idle at prompt with no activity
- PRs accumulate without being processed

### Root Cause Analysis

After ~12 hours of continuous operation (spawned 2026-03-08 18:41, unresponsive by ~2026-03-09 06:35), both merge-queue and pr-shepherd locked up simultaneously. During that window they processed 20+ PR merges. 19 PRs sat unprocessed for ~30 minutes before the supervisor noticed and restarted them.

**Likely causes (in order of probability):**

1. **Context window exhaustion** — Both agents received 71+ messages and ran continuous merge/rebase operations over 12 hours. Each `gh pr checks`, `gh pr view`, and merge cycle adds tool calls and results to the context. Eventually the context fills and the model can no longer generate responses.

2. **Sequential bottleneck compounding** — Each merge requires: update branch → wait for CI (3-5 min) → merge → repeat. With 19 PRs queued simultaneously, every branch update can invalidate other PRs' "up to date" status, creating a cascade of re-updates.

3. **Unclosed quote deadlock** — A Bash tool call containing an unclosed quote causes the shell to wait on stdin for the closing quote, while Claude waits for the tool result. This creates a permanent deadlock. (Previously observed 2026-03-08.)

### Recovery Procedure

```bash
# 1. Kill the frozen agent's tmux window
# (Ctrl-B then & in tmux, or from another terminal:)
tmux kill-window -t <agent-window>

# 2. Remove stale state entries
# Edit ~/.multiclaude/state.json to remove the agent's entry
# (Required because multiclaude tracks agent existence)

# 3. Respawn with --force to override cached state
multiclaude agents spawn --name merge-queue --class persistent --prompt-file agents/merge-queue.md --force
multiclaude agents spawn --name pr-shepherd --class persistent --prompt-file agents/pr-shepherd.md --force

# 4. Nudge the respawned agent with pending work
multiclaude message send merge-queue "Restarted after lockup. Priority PRs: #301, #302, ..."
```

### Prevention

1. **Proactive restart policy** — Restart persistent agents every 4-6 hours or after processing ~15-20 merges. Context exhaustion is cumulative and inevitable in long sessions.

2. **Supervisor health check loop** — Run a periodic check (e.g., every 30 min) where the supervisor pings merge-queue and pr-shepherd and expects a response within 5 minutes. No response = automatic restart.

3. **Transcript inspection for deadlocks** — When an agent appears frozen, check its JSONL transcript (in `~/.claude/projects/`) for the last tool call. Look for:
   - Unclosed quotes in Bash tool calls
   - Extremely long tool results that may have filled context
   - The absence of any recent tool calls (indicates context exhaustion)

4. **Batch awareness** — When many PRs queue up simultaneously, consider staggering merges rather than processing all at once. This reduces the cascade of branch invalidations.

---

## Heartbeat Mechanism

### Overview

All persistent agents go idle after completing startup work because Claude has no internal timers. The heartbeat mechanism uses CronCreate to periodically send HEARTBEAT messages that trigger each agent's polling loop.

### How It Works

1. **Supervisor creates CronCreate jobs** during startup (session-scoped, re-created on every restart)
2. Each cron fires a prompt that sends `multiclaude message send <agent> HEARTBEAT`
3. The agent receives the HEARTBEAT, runs its full polling loop, and acks the message
4. Agents report findings through normal channels (messages to supervisor, spawning workers, etc.)

### Heartbeat Schedule

| Agent | Interval | Cron Expression | Rationale |
|-------|----------|----------------|-----------|
| merge-queue | 7 min | `*/7 * * * *` | High-frequency — PR merges are time-sensitive |
| pr-shepherd | 7 min (offset +3) | `3-59/7 * * * *` | Same frequency as merge-queue, staggered |
| envoy | 11 min | `*/11 * * * *` | Medium — issue triage is important but not urgent |
| retrospector | 13 min | `*/13 * * * *` | Medium — analysis can tolerate slight delay |
| project-watchdog | 23 min | `*/23 * * * *` | Lower frequency — doc sync is batched |
| arch-watchdog | 23 min (offset +5) | `5-59/23 * * * *` | Same frequency as project-watchdog, staggered |

Prime-number intervals prevent all agents from being poked simultaneously and avoid the :00/:30 minute marks.

### Operational Data Sync Schedule

In addition to heartbeats, a dedicated cron syncs retrospector operational data to git:

| Purpose | Interval | Cron Expression | Target Agent | Message |
|---------|----------|----------------|--------------|---------|
| Data sync | 3 hours | `0 */3 * * *` | project-watchdog | `SYNC_OPERATIONAL_DATA` |

This runs at :00 on hours 0, 3, 6, 9, 12, 15, 18, 21 — deliberately NOT using prime intervals (reserved for heartbeats).

**Supervisor startup must include:**
```
CronCreate("0 */3 * * *", "multiclaude message send project-watchdog SYNC_OPERATIONAL_DATA")
```

### Limitations

- **Session-scoped:** CronCreate jobs are lost when the supervisor exits and auto-expire after 3 days
- **Idle-only firing:** Crons only fire while the supervisor's REPL is idle (not mid-query)
- **No guaranteed delivery:** If the supervisor is busy processing a query when a cron fires, that heartbeat is skipped — the next interval will catch it
- **Context impact:** Each heartbeat adds a small amount of context to the target agent. At the configured intervals, this is manageable within the 4-6 hour restart cadence

### Monitoring

When checking agent health, verify heartbeats are working:
- Agent should show activity every 1-2 heartbeat intervals
- If an agent is idle for 3+ intervals, check: (1) supervisor crons exist, (2) messaging is working, (3) agent isn't context-exhausted

---

## Operational Data Convention

### Canonical Path

`docs/operations/` is the canonical directory for all agent-generated operational data files. Any persistent agent that produces operational data (findings, checkpoints, recommendations, metrics) should write files here.

### Tracked Files

| File | Format | Producer | Purpose |
|------|--------|----------|---------|
| `retrospector-findings.jsonl` | JSONL (append-only) | retrospector | Operational findings and analysis results |
| `retrospector-checkpoint.json` | JSON | retrospector | Current analysis state and rolling metrics |
| `retrospector-inbox.jsonl` | JSONL (append-only) | supervisor/agents | Inbound analysis requests for retrospector |
| `retrospector-recommendations.jsonl` | JSONL (append-only) | retrospector | Recommendations pending project-watchdog consumption |

### Sync Pipeline

The `SYNC_OPERATIONAL_DATA` cron (every 3 hours) triggers project-watchdog to:
1. Check `docs/operations/` for uncommitted or untracked data files (`*.jsonl`, `*.json`)
2. If changes exist: create `data-sync/<timestamp>` branch, commit, push, create PR
3. If no changes: do nothing (idempotent)

Data sync PRs go through normal branch protection (PR + CI). merge-queue handles them like any other PR.

### Staleness SLA

`docs/operations/*.jsonl` files in git should never be more than 6 hours stale relative to the main checkout (2x the 3-hour sync interval). If no `data-sync` commits appear in `git log --oneline docs/operations/ --since="8 hours ago"` and retrospector is active, investigate the sync pipeline.

### Adding New Operational Data

When a new agent needs to persist operational data:
1. Create the file in `docs/operations/`
2. Add it to the tracked files table above
3. The sync cron will automatically pick it up — no additional configuration needed
4. For JSONL files: use append-only format for merge-friendliness

---

## General Persistent Agent Best Practices

### Hot-Reload Limitation

Claude agents **cannot hot-reload system prompts**. Telling an agent to "re-read your definition file" is a no-op and can cause the agent to freeze. After changing an agent definition file:

1. Kill the agent's tmux window
2. `git pull` on main so the agent gets latest definitions
3. Respawn with `multiclaude agents spawn`

### OAuth Scope Limitation

merge-queue cannot merge PRs that modify `.github/workflows/` files — the OAuth token lacks `workflow` scope. These PRs must be merged manually by the project owner.

### Monitoring Checklist

| Check | Frequency | Action on Failure |
|-------|-----------|-------------------|
| Agent responding to pings | Every 30 min | Restart agent |
| PR queue depth growing | Every 15 min | Nudge or restart merge-queue |
| Main branch CI status | Every merge | Enter emergency mode if red |
| Agent message backlog | Every 30 min | Restart if messages unacknowledged >30 min |
| Heartbeat crons active | On supervisor restart | Re-create CronCreate jobs (see Heartbeat Mechanism) |
| Agent activity after heartbeat | Every 2-3 intervals | If idle for 3+ intervals, check crons/messaging/context |
| Operational data freshness | Every 8 hours | `git log --oneline docs/operations/ --since="8 hours ago"` — if no commits and retrospector active, investigate sync pipeline |
| Data sync cron active | On supervisor restart | Re-create `SYNC_OPERATIONAL_DATA` CronCreate job (see Operational Data Sync Schedule) |
