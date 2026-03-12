# Remote Collaboration Guide

Connect to a running multiclaude session from a different machine via SSH to monitor agents, dispatch tasks, and exchange messages — with zero code changes.

## Prerequisites

- SSH access to the machine running the multiclaude session (key-based auth recommended)
- The `multiclaude` binary in PATH on the host machine
- A running multiclaude session (verify with `multiclaude daemon status`)
- For unstable connections: [mosh](https://mosh.org/) as a drop-in SSH replacement

## SSH Setup

### Basic Connection

```bash
# Verify SSH access and multiclaude availability
ssh user@host "multiclaude daemon status"
```

If `multiclaude` is not in PATH over SSH, add its location to your SSH environment or use the full path:

```bash
ssh user@host "/usr/local/bin/multiclaude daemon status"
```

### Persistent Connection (Recommended)

Add an SSH config entry for convenience:

```
# ~/.ssh/config
Host mc-host
    HostName 192.168.1.100
    User skippy
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

Then connect with `ssh mc-host "multiclaude status"`.

For unstable or high-latency connections, use `mosh` instead of SSH. It handles roaming, intermittent connectivity, and provides local echo:

```bash
mosh mc-host -- multiclaude status
```

## Monitoring (tmux)

### Read-Only Session Attach

Attach to the multiclaude tmux session in read-only mode to observe agent activity in real time:

```bash
# Attach read-only — you can see everything but cannot type
ssh -t mc-host "tmux attach -t mc-ThreeDoors -r"
```

Once attached, navigate between agent windows:

| Key | Action |
|-----|--------|
| `Ctrl-b n` | Next window (next agent) |
| `Ctrl-b p` | Previous window |
| `Ctrl-b 0-9` | Jump to window by number |
| `Ctrl-b w` | List all windows (interactive picker) |
| `Ctrl-b d` | Detach (disconnect without stopping anything) |

### Attach to a Specific Agent

```bash
# View a specific agent using multiclaude's attach command
ssh -t mc-host "multiclaude agent attach supervisor --read-only"
```

### Monitor Agent Logs

Tail agent logs without attaching to tmux:

```bash
# Follow supervisor logs in real time
ssh mc-host "multiclaude logs supervisor -f"

# Search across all logs
ssh mc-host 'multiclaude logs search "error"'

# List available log files
ssh mc-host "multiclaude logs list"
```

## Task Dispatch

Send work to the supervisor from your remote machine:

```bash
# Spawn a new worker task
ssh mc-host 'multiclaude worker create "Fix the flaky test in internal/tasks/pool_test.go"'

# Verify the worker started
ssh mc-host "multiclaude worker list"
```

Check system status to see all active agents and workers:

```bash
ssh mc-host "multiclaude status"
```

View task history:

```bash
ssh mc-host "multiclaude repo history"
ssh mc-host "multiclaude repo history --status completed -n 10"
```

## Message Passing

### Send a Message

Send a message to any agent (most commonly the supervisor):

```bash
# Send a message to the supervisor
ssh mc-host 'multiclaude message send supervisor "Priority: Review PR #456 before EOD"'

# Send to a specific agent
ssh mc-host 'multiclaude message send merge-queue "Hold merges until CI is green"'
```

### Check for Replies

Poll for messages addressed to you or pending in the system:

```bash
# List all pending messages
ssh mc-host "multiclaude message list"

# Read a specific message
ssh mc-host "multiclaude message read msg-abc123"

# Acknowledge a message (marks it as handled)
ssh mc-host "multiclaude message ack msg-abc123"
```

**Note:** Agents poll for messages at intervals (5-30 minutes depending on agent type). Message delivery is not instant — expect a delay between sending and the agent acting on your message.

### Bidirectional Conversation

A typical remote interaction flow:

1. Send a message: `ssh mc-host 'multiclaude message send supervisor "What is the status of Epic 40?"'`
2. Wait for the supervisor to process it (check with `multiclaude message list`)
3. Read the reply: `ssh mc-host "multiclaude message read <reply-id>"`
4. Acknowledge: `ssh mc-host "multiclaude message ack <reply-id>"`

## Safety & Caveats

### Command Safety Reference

**Read-only commands** (safe to run freely):

| Command | Purpose |
|---------|---------|
| `multiclaude status` | System overview |
| `multiclaude worker list` | List active workers |
| `multiclaude message list` | List pending messages |
| `multiclaude message read <id>` | Read a message |
| `multiclaude diagnostics --json` | Full system diagnostics |
| `multiclaude repo list` | List repositories |
| `multiclaude repo history` | Task history |
| `multiclaude logs [agent]` | View agent logs |
| `multiclaude logs search <pattern>` | Search logs |
| `multiclaude daemon status` | Daemon health |

**State-modifying commands** (use with caution):

| Command | Purpose | What It Does |
|---------|---------|-------------|
| `multiclaude worker create <task>` | Spawn new worker | Creates worktree, tmux window, starts Claude process |
| `multiclaude message send <to> <msg>` | Send a message | Writes message file for delivery |
| `multiclaude message ack <id>` | Acknowledge message | Marks message as handled |
| `multiclaude agent restart <name>` | Restart agent | Kills and respawns a Claude process |
| `multiclaude agent complete` | Signal task done | Updates agent state |
| `multiclaude refresh` | Sync worktrees | Runs git operations on all worktrees |

### Critical Warning: Do Not Type into Agent tmux Windows

**Never attach to tmux in read-write mode and type into an agent window.** Raw keystrokes are injected directly into the running Claude Code process's input stream. This corrupts the agent's conversation state and can cause unpredictable behavior — garbled tool calls, broken context, or agent crashes.

Always use read-only mode (`-r` flag) when attaching to tmux:

```bash
# SAFE — read-only, you cannot type
ssh -t mc-host "tmux attach -t mc-ThreeDoors -r"

# DANGEROUS — read-write, keystrokes go to agents
ssh -t mc-host "tmux attach -t mc-ThreeDoors"
```

If you need to communicate with an agent, use `multiclaude message send`, not tmux keystrokes.

### Security Notes

- SSH provides authentication and encryption — all commands run over an encrypted channel
- The multiclaude message system has no internal authentication. Any process with filesystem access can send messages as any agent. SSH access is your auth boundary.
- Do not expose the multiclaude daemon socket (`~/.multiclaude/daemon.sock`) over the network. It is a Unix domain socket with no authentication — all access control comes from filesystem permissions.
- The `state.json` file is managed exclusively by the daemon. Never edit it directly — use CLI commands instead.

## GitHub Issues Fallback

When SSH access is unavailable (firewall restrictions, VPN down, mobile), use GitHub Issues as an async task channel.

### Submitting a Task via GitHub Issue

Create an issue with the `remote-task` label and a structured body:

```bash
gh issue create \
  --title "Remote task: Fix pagination in stats view" \
  --label "remote-task" \
  --body "$(cat <<'EOF'
## Task

Fix the pagination bug in the stats view where page 2 shows duplicate entries.

## Priority

Medium

## Context

Reported by users in #general. Likely related to the offset calculation in
internal/tui/stats_view.go.
EOF
)"
```

The `remote-task` label signals to the envoy agent that this issue is an incoming task request. Envoy will triage it through the standard BMAD pipeline and report back on the issue with status updates.

### Monitoring Issue Progress

```bash
# List remote-task issues
gh issue list --label "remote-task" --state open

# Check a specific issue for updates
gh issue view 42
```

Envoy posts acknowledgments and status updates as comments on the issue. When work is complete, the issue is closed with a summary of what was done and links to relevant PRs.

### When to Use GitHub Issues vs SSH

| Scenario | Use |
|----------|-----|
| You have SSH access and need quick action | SSH + `multiclaude` CLI |
| No SSH access (firewall, mobile, different network) | GitHub Issues with `remote-task` label |
| Async request that can wait hours | GitHub Issues |
| Real-time monitoring of agent activity | SSH + tmux attach |
| Urgent task that needs immediate attention | SSH + `multiclaude worker create` |

## Troubleshooting

### Connection drops mid-session

If your SSH connection drops while attached to tmux, the multiclaude session continues running unaffected. Simply reconnect:

```bash
ssh -t mc-host "tmux attach -t mc-ThreeDoors -r"
```

Use `mosh` instead of SSH for automatic reconnection on unstable networks.

### tmux socket permissions

If tmux reports "no sessions" or permission errors:

```bash
# Check the tmux socket
ssh mc-host "ls -la /private/tmp/tmux-*/default"

# You must be the same user that started the multiclaude session
# tmux sockets are user-only (mode 0700)
```

### multiclaude not in PATH

If `multiclaude` is not found over SSH, the SSH session may not load the full shell profile. Options:

```bash
# Use the full path
ssh mc-host "/usr/local/bin/multiclaude status"

# Or source the profile explicitly
ssh mc-host "source ~/.zshrc && multiclaude status"

# Or add to ~/.ssh/environment (requires PermitUserEnvironment in sshd_config)
```

### Daemon not running

```bash
# Check if the daemon is running
ssh mc-host "multiclaude daemon status"

# Start it if needed
ssh mc-host "multiclaude daemon start"

# Check daemon logs for errors
ssh mc-host "multiclaude daemon logs -n 50"
```

### Agent appears frozen

If an agent window shows no activity:

1. Check if the Claude Code process is still running: `ssh mc-host "multiclaude status"`
2. Check agent logs: `ssh mc-host "multiclaude logs <agent-name>"`
3. Restart the agent if needed: `ssh mc-host "multiclaude agent restart <agent-name>"`
