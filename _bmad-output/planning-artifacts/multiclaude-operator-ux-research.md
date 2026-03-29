# multiclaude Operator UX & Architecture Pain Points

**Date:** 2026-03-29
**Type:** Research Spike
**Worker:** brave-dolphin
**Status:** Complete

---

## Executive Summary

The multiclaude supervisor experience has significant UX friction caused by a fundamental architectural choice: **all agent communication (wake nudges, message delivery, heartbeat responses) is implemented via `tmux send-keys` / `paste-buffer` injection directly into the target agent's tmux pane**. When the human operator works in the supervisor's tmux window, these injections corrupt their input, interrupt their flow, and create confusion about what's happening.

The workspace window exists specifically to solve this problem but is underutilized. The CronCreate-based heartbeat system adds a second layer of injection into the supervisor specifically. Multiple architectural improvements are possible, ranging from immediate workflow fixes to medium-term protocol changes.

---

## Finding 1: The Prompt Injection Mechanism (Root Cause)

### How It Works

All inter-agent communication in multiclaude flows through the same mechanism:

1. **Message files** are written to `~/.multiclaude/messages/<repo>/<agent>/msg-*.json` (JSON with id, from, to, timestamp, body, status)
2. **The daemon's `routeMessages()` loop** (runs every 2 minutes) reads pending messages and delivers them
3. **Delivery mechanism:** `tmux set-buffer -- "$text" && tmux paste-buffer -t <target> && tmux send-keys -t <target> Enter`

This is defined in `pkg/tmux/client.go:319` (`SendKeysLiteralWithEnter`) and used by:
- **Message delivery** (`daemon.go:409`): Formats as `📨 Message from <sender>: <body>` and pastes into the target pane
- **Wake nudges** (`daemon.go:474`): Every 2 minutes, sends role-specific prompts like "Status check: Review worker progress and check merge queue." to each agent's pane
- **Initial messages** (`cli.go:5906`): When spawning agents, sends the initial task prompt

### Why This Causes Problems for the Supervisor

The supervisor window receives **three categories** of injected text:

1. **Daemon wake nudges** (every 2 minutes): `"Status check: Review worker progress and check merge queue."`
2. **Message deliveries** (whenever an agent sends to supervisor): `"📨 Message from merge-queue: PR #828 merged successfully"`
3. **CronCreate prompt injections** (from heartbeat crons): These fire prompts into the Claude REPL as if typed by the user

When the human is typing in the supervisor window, a paste-buffer injection mid-keystroke will **insert the message text at the cursor position**, mangling both the user's input and the message. The Enter key at the end then submits the corrupted combined text.

### Source Code References

| File | Line | Function | Purpose |
|------|------|----------|---------|
| `pkg/tmux/client.go` | 319 | `SendKeysLiteralWithEnter()` | Atomic paste+enter via `sh -c` |
| `internal/daemon/daemon.go` | 437 | `wakeAgents()` | 2-min nudge loop |
| `internal/daemon/daemon.go` | 372 | `routeMessages()` | Message delivery loop |
| `internal/daemon/daemon.go` | 474 | `SendKeysLiteralWithEnter(...)` | Wake delivery call site |
| `internal/daemon/daemon.go` | 409 | `SendKeysLiteralWithEnter(...)` | Message delivery call site |

---

## Finding 2: The Workspace Window — The Intended Human Environment

### What the Workspace Is

The workspace window is a **dedicated tmux window with its own worktree** on the `workspace/default` branch. It is explicitly designed to be free from automated interruptions:

- `wakeAgents()` (daemon.go:447): `if agent.Type == state.AgentTypeWorkspace { continue }` — **skips workspace**
- `routeMessages()` (daemon.go:387): `if agent.Type == state.AgentTypeWorkspace { continue }` — **skips workspace**

The workspace window is **the only window that never receives automated injections**. It's a clean shell in the repo's worktree, intended for the human to use directly.

### Current Setup

```
tmux windows in mc-ThreeDoors:
0: supervisor     ← Claude agent (receives wake + messages + cron prompts)
1: workspace      ← Clean shell (NO automated injections)
2: merge-queue    ← Claude agent
3: pr-shepherd    ← Claude agent
4: arch-watchdog  ← Claude agent
5: envoy          ← Claude agent
6: project-watchdog ← Claude agent
7: retrospector   ← Claude agent
8: brave-dolphin  ← Worker (this session)
```

### Recommendation

**The human should NOT work in the supervisor window.** The supervisor window belongs to the supervisor Claude agent. The human should:
1. Use the **workspace window** for direct CLI/git operations
2. Use **`multiclaude message send supervisor "..."` from the workspace** to communicate with the supervisor agent
3. Use **`multiclaude agent attach <name> --read-only`** to observe agent activity without interfering

---

## Finding 3: CronCreate Heartbeat Architecture

### How CronCreate Interacts with the Supervisor

CronCreate schedules prompts that fire **into the Claude REPL session that created them**. When the supervisor creates heartbeat crons like:
```
CronCreate("*/7 * * * *", "multiclaude message send merge-queue HEARTBEAT")
```

This fires the text `multiclaude message send merge-queue HEARTBEAT` as a user prompt into the **supervisor's Claude session** every 7 minutes. The supervisor's Claude then processes this by running the Bash command, which writes a message file, which the daemon picks up and delivers to merge-queue.

### The Double-Injection Problem

The supervisor receives injections from **two independent systems**:

1. **Daemon wake loop** (every 2 minutes): Pastes "Status check: ..." via tmux
2. **CronCreate jobs** (every 7-23 minutes per agent): Fires prompts into Claude REPL

These compete for the supervisor's attention. When both fire simultaneously, the supervisor gets confused or one prompt overwrites the other.

### Why Daemon-Level Heartbeats Would Be Better

The daemon already has a wake loop (`wakeLoop`, daemon.go:432) that runs every 2 minutes. Heartbeats could be implemented as a daemon feature rather than CronCreate jobs:

- **No injection into supervisor**: The daemon would write heartbeat messages directly to the target agent's message directory
- **No CronCreate dependency**: Heartbeats would survive supervisor restarts without needing to be recreated
- **Configurable per-agent**: The daemon already has per-agent-type logic in `wakeAgents()`
- **Deduplication**: The daemon already tracks `LastNudge` per agent and skips recently nudged agents

---

## Finding 4: Message Queue Architecture Assessment

### Current Architecture

```
[Agent A] → multiclaude message send B "text"
  → writes JSON file to ~/.multiclaude/messages/<repo>/B/msg-*.json
  → calls daemon socket: route_messages
  → daemon reads pending messages for B
  → daemon pastes message into B's tmux pane via paste-buffer
  → marks message as "delivered" in JSON file
```

### Strengths of Current Approach

1. **Simplicity**: No external dependencies, just files and tmux
2. **Debuggability**: Messages are JSON files you can read/inspect
3. **Persistence**: Messages survive daemon restarts
4. **Audit trail**: Message history preserved on disk

### Weaknesses

1. **Delivery = tmux paste**: If the agent's Claude is mid-tool-call, the paste goes into the terminal buffer and may be lost or misinterpreted
2. **No acknowledgment protocol**: "Delivered" means "pasted into tmux" not "agent processed it"
3. **No backpressure**: Can flood an agent with messages faster than it can process
4. **Race conditions**: Two concurrent pastes to the same pane can interleave
5. **Human interference**: If human is in the pane, message gets mangled with human input

### Would a Proper Message Queue Help?

**Assessment: Not worth the complexity at current scale.**

| Option | Pros | Cons |
|--------|------|------|
| Redis pub/sub | Real acknowledgment, backpressure | External dependency, overkill for ~10 agents |
| NATS | Lightweight, perfect for this | Still an external process to manage |
| Unix domain sockets | No external deps, fast | Claude Code can't listen on sockets (stdin-only) |
| Named pipes (FIFO) | Simple IPC | Same stdin problem — Claude reads from its own stdin |

**The fundamental constraint:** Claude Code agents read from stdin (their Claude REPL). There is no way to push messages to a Claude agent except by injecting text into its stdin or tmux pane. A message queue would still need the tmux injection step for the last mile.

The real improvement path is **reducing injections** (daemon-level heartbeats, batched messages) rather than changing the transport.

---

## Finding 5: End-to-End Heartbeat Flow

### Current Flow (CronCreate-based)

```
1. CronCreate fires in supervisor Claude session
2. Supervisor runs: multiclaude message send merge-queue HEARTBEAT
3. CLI writes msg-*.json to merge-queue's message dir
4. CLI calls daemon socket: route_messages
5. Daemon reads pending message
6. Daemon does: tmux set-buffer "📨 Message from supervisor: HEARTBEAT" &&
                 tmux paste-buffer -t mc-ThreeDoors:merge-queue &&
                 tmux send-keys -t mc-ThreeDoors:merge-queue Enter
7. merge-queue's Claude processes the heartbeat prompt
8. merge-queue may respond (writing another message back)
9. Daemon delivers response back to supervisor via tmux paste
```

Steps 1-4 inject into the supervisor (CronCreate prompt). Steps 8-9 inject again when the response comes back. **Two injections per heartbeat round-trip.**

### Proposed Flow (Daemon-native heartbeats)

```
1. Daemon heartbeat timer fires for merge-queue
2. Daemon does: tmux set-buffer "HEARTBEAT" &&
                 tmux paste-buffer -t mc-ThreeDoors:merge-queue &&
                 tmux send-keys -t mc-ThreeDoors:merge-queue Enter
3. merge-queue processes, may file a message for supervisor
4. Daemon delivers only substantive messages to supervisor
```

**Zero supervisor injections for routine heartbeats.** Only substantive messages (PR merged, error found, etc.) reach the supervisor.

---

## Recommendations

### Short-Term (Workflow Changes — No Code Required)

| # | Action | Impact | Effort |
|---|--------|--------|--------|
| S-1 | **Human works in workspace window, not supervisor** | Eliminates all injection interference for human | Behavioral change only |
| S-2 | **Remove CronCreate heartbeats entirely** — the daemon wake loop already nudges all agents every 2 minutes | Reduces supervisor injections by ~60% | Delete cron setup from MEMORY.md startup checklist |
| S-3 | **Use `multiclaude message send` from workspace** to communicate with supervisor | Clean separation of human input from automated messages | Behavioral change |

### Medium-Term (multiclaude Patches)

| # | Action | Impact | Effort |
|---|--------|--------|--------|
| M-1 | **Daemon-native heartbeats** with configurable per-agent intervals | Eliminates CronCreate dependency, survives restarts | ~2-3 days |
| M-2 | **Message batching** — daemon accumulates messages for a target and delivers them as a single paste every N seconds | Reduces injection frequency, prevents interleaving | ~1-2 days |
| M-3 | **Supervisor message filtering** — daemon only delivers messages above a priority threshold, queues low-priority for polling | Reduces noise in supervisor | ~1 day |
| M-4 | **`--quiet` mode for supervisor** — disables wake nudges for the supervisor specifically (it gets prompted by CronCreate or human instead) | Human-triggered supervision | ~0.5 day |

### Long-Term (Architecture Evolution)

| # | Action | Impact | Effort |
|---|--------|--------|--------|
| L-1 | **Claude Code MCP server integration** — agents expose an MCP tool for receiving messages, avoiding tmux injection entirely | Clean message protocol, no tmux dependency | Depends on Claude Code MCP evolution |
| L-2 | **Operator dashboard** — web UI showing agent status, message queues, and allowing human to send commands | Rich UX, no tmux wrestling | ~1-2 weeks |
| L-3 | **Agent-to-agent direct protocol** — agents write to each other's `--append-system-prompt-file` instead of tmux injection | Cleaner delivery, queued per-turn | Requires Claude Code API changes |

---

## Rejected Approaches

| Approach | Why Rejected |
|----------|--------------|
| Redis/NATS message queue | Adds external dependency; Claude Code's stdin-only architecture means tmux injection is still needed for last-mile delivery. Complexity without solving the root problem. |
| Disable daemon wake loop entirely | Agents would go dormant without periodic nudges. The wake loop is necessary for agent liveness. |
| Human runs Claude in supervisor window alongside the agent | Two Claude processes would fight for stdin. Not architecturally supported. |
| Separate human tmux session outside mc-ThreeDoors | Loses tmux integration (can't `tmux select-window` to observe agents). Workspace window is the right solution. |

---

## Open Questions

| # | Question | Impact |
|---|----------|--------|
| OQ-1 | Should the daemon wake loop interval be configurable per-agent-type? Currently hardcoded to 2 minutes for all. | Persistent agents (merge-queue) may need faster wakes than watchdogs |
| OQ-2 | Could Claude Code's `--append-system-prompt-file` be used as a message channel? (Agent reads new messages from file each turn) | Would eliminate tmux injection entirely |
| OQ-3 | Should multiclaude support a "human operator" agent type that is never injected but can read all messages? | Proper operator UX without workspace workaround |
| OQ-4 | Is the 2-minute wake interval too aggressive? Many agents have nothing to report. | Reducing to 5 minutes would cut injections by 60% |

---

## Appendix: Key Source Files

| File | Purpose |
|------|---------|
| `/Users/skippy/work/multiclaude/pkg/tmux/client.go` | tmux operations including `SendKeysLiteralWithEnter` |
| `/Users/skippy/work/multiclaude/internal/daemon/daemon.go` | Daemon loops: wake, message routing, worktree refresh |
| `/Users/skippy/work/multiclaude/internal/messages/` | Message file management |
| `/Users/skippy/work/multiclaude/internal/state/state.go` | Agent types (workspace, supervisor, etc.) |
| `/Users/skippy/work/multiclaude/internal/cli/cli.go` | CLI commands including workspace management |
| `~/.multiclaude/messages/ThreeDoors/<agent>/` | Per-agent message directories |
| `~/.multiclaude/daemon.log` | Daemon activity log |
