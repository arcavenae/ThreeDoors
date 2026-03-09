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
