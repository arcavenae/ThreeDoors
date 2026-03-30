# INC-004: SendMessage Tool Silently Drops Inter-Agent Messages

**Discovered:** 2026-03-30
**Severity:** Critical — complete messaging blackout for project-watchdog
**Status:** Fixed (agent definitions updated)

## Summary

All multiclaude agents were using Claude Code's built-in `SendMessage` tool instead of `multiclaude message send` (Bash CLI) for inter-agent communication. The `SendMessage` tool is designed for subagent communication within a single Claude Code process (teammates spawned via the `Agent` tool). It has zero connection to multiclaude's inter-agent messaging system. Messages sent via `SendMessage` are silently accepted and silently dropped — no error, no delivery.

## Impact

- **project-watchdog:** 16+ messages silently dropped, including Epic 77 allocation reply, governance alerts, HEARTBEAT responses, and data sync notifications
- **merge-queue:** 33+ messages silently dropped (escalations, scope alerts, emergency mode reports)
- **All other agents:** Affected but less visibly — their primary outputs are actions (merging PRs, rebasing, reviewing), not messages

The most critical failure: project-watchdog is the MUTEX for epic/story number allocation. When it allocated Epic 77 and sent the reply via `SendMessage`, the supervisor never received it, leading to coordination failure and the user having to manually relay the information.

## Root Cause

Claude Code exposes a `SendMessage` tool as a deferred tool available to all agents. When an agent is instructed to "send a message to supervisor," Claude's model prefers the purpose-built `SendMessage` tool over crafting a `Bash("multiclaude message send supervisor '...'")` call, even when the agent definition explicitly documents the Bash command syntax.

The `SendMessage` tool's description says "Send a message to another agent" — which is technically accurate for Claude Code's native multi-agent system but misleading in the multiclaude context where "agents" are separate tmux-based Claude processes, not subagents within the same process.

## Evidence

From project-watchdog's session transcript (`0c0c5787-4df6-4c19-a248-ba6f7ff9fc89.jsonl`):
- First action after receiving EPIC NUMBER REQUEST: `ToolSearch("select:SendMessage")` to fetch the tool schema
- Then used `SendMessage(to="supervisor", message="Allocated Epic 77...")` — silently dropped
- No `Bash("multiclaude message send ...")` commands were ever executed

Contrast: daemon-injected messages (`📨 Message from supervisor: HEARTBEAT`) arrive via tmux pane injection and work correctly. The failure is exclusively in the outbound direction (agent → supervisor/other agents).

## Fix

Added **INC-004 guardrail** to the Communication section of all 11 agent definition files:

```markdown
**CRITICAL — INC-004: Use `multiclaude message send` via Bash, NEVER the `SendMessage` tool.**

Claude Code's built-in `SendMessage` tool is for subagent communication within a single Claude process —
it does NOT route through multiclaude's inter-agent messaging. Messages sent via `SendMessage` are
silently dropped. Always use Bash.
```

**Files modified:**
- `agents/project-watchdog.md`
- `agents/merge-queue.md`
- `agents/arch-watchdog.md`
- `agents/envoy.md`
- `agents/pr-shepherd.md`
- `agents/retrospector.md`
- `agents/worker.md`
- `agents/supervisor.md`
- `agents/release-manager.md`
- `agents/research-supervisor.md`
- `agents/reviewer.md`

## Required Follow-Up

1. **Restart all persistent agents** after this PR merges — agent definitions are baked in at spawn time (see Known Bug in MEMORY.md)
2. **Re-send Epic 77 allocation** manually to supervisor after restart
3. **Audit project-watchdog's dropped messages** — the 16 lost messages include governance alerts about Story 73.7 self-assignment that were never actioned

## Prevention

The guardrail text is incident-hardened and follows the INC-001/002/003 pattern. Future agents will see the warning prominently in their Communication section. However, a stronger fix would be a PreToolUse hook that blocks `SendMessage` calls with `to` values matching known multiclaude agent names — similar to how `git-safety.sh` blocks dangerous git commands.
