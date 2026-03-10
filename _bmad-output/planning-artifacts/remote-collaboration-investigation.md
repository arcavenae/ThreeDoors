# Remote Claude Code Collaboration with multiclaude — Investigation

**Date:** 2026-03-10
**Type:** Research / Feasibility Assessment
**Status:** Complete

## Problem Statement

Can a Claude Code session running on a different machine (or in a separate terminal outside the multiclaude tmux session) collaborate with a multiclaude supervisor and its agents? What mechanisms exist or could be built?

## Current Architecture Summary

multiclaude uses a **hub-and-spoke** model:
- **Daemon** runs as a background process, manages state via `/Users/skippy/.multiclaude/state.json`
- **Agents** run as Claude Code processes in tmux windows within session `mc-<RepoName>`
- **Messages** are JSON files stored at `/Users/skippy/.multiclaude/messages/<repo>/<agent>/*.json`
- **Workers** get isolated git worktrees at `/Users/skippy/.multiclaude/wts/<repo>/<worker-name>/`
- **IPC** uses a Unix domain socket at `/Users/skippy/.multiclaude/daemon.sock`
- **No MCP servers** are currently configured; all agent communication uses the `multiclaude` CLI

## Approach Analysis

### 1. multiclaude Message Passing (CLI)

**Mechanism:** `multiclaude message send <recipient> "<body>"`

**How it works:**
- Creates a JSON file in the recipient's inbox directory
- Recipient agents poll `multiclaude message list` at intervals (5-30 min depending on agent type)
- Messages persist until acknowledged via `multiclaude message ack <id>`

**Feasibility for remote use: HIGH (local), LOW (truly remote)**

| Scenario | Works? | Notes |
|----------|--------|-------|
| Same machine, different terminal | Yes | Just run `multiclaude message send` from any shell |
| Same machine, standalone Claude Code | Yes | Claude Code can call `multiclaude` via Bash tool |
| Remote machine via SSH | Partial | Can SSH and run `multiclaude message send`, but requires multiclaude installed and state accessible |
| Remote machine without SSH | No | No network-accessible API exists |

**Strengths:**
- Already works today for same-machine scenarios
- No code changes needed — any process with shell access to `multiclaude` can send messages
- Messages are durable (persisted to disk)
- Agents already have polling loops to check for messages

**Weaknesses:**
- High latency — agents poll every 5-30 minutes
- No real-time notification mechanism
- Requires `multiclaude` binary on the calling machine (or SSH access to the host)
- Message protocol is fire-and-forget; no request/response correlation IDs
- Remote Claude Code session has no way to call `multiclaude message list` to receive replies without polling

**Key finding:** A standalone Claude Code session on the same machine can participate TODAY by running `multiclaude message send supervisor "..."` via Bash. The supervisor will see it on its next message poll. Reply comes back the same way. This is the lowest-friction approach.

### 2. SSH + tmux Attach

**Mechanism:** `ssh host "tmux attach -t mc-ThreeDoors -r"` (read-only) or direct attach

**Feasibility: MEDIUM (monitoring), LOW (collaboration)**

**What works:**
- Remote user can watch agent activity in real-time via `tmux attach -t mc-ThreeDoors`
- Can switch between agent windows to monitor different agents
- Read-only mode (`-r` flag or `multiclaude agent attach <name> --read-only`) prevents accidental input

**What doesn't work:**
- tmux attach gives raw terminal access, not structured communication
- Typing into an agent's tmux window sends keystrokes to the running Claude Code process — this is dangerous and unpredictable (it could corrupt the agent's input stream)
- No way to send a "message" through tmux that agents would understand as a task or query
- Requires SSH access with the user's account

**Best use:** Monitoring and debugging, not collaboration. Pair with message passing for actual communication.

### 3. Shared Git Repo as Communication Channel

**Mechanism:** External Claude Code pushes branches/PRs; multiclaude agents process them through normal workflows

**Feasibility: HIGH for code contributions, LOW for real-time coordination**

**How it works today:**
- External contributor creates a branch, pushes code, opens a PR
- merge-queue detects the PR (polls GitHub) and processes it
- pr-shepherd handles rebasing if needed
- arch-watchdog reviews for architectural compliance

**Strengths:**
- Already fully functional — this is the standard open-source contribution model
- External Claude Code needs only git access, no multiclaude installation
- Full audit trail via git history and PR comments
- Works across machines, networks, and even organizations

**Weaknesses:**
- Very high latency (minutes to detect new PR, more for review cycles)
- Only works for code-level changes, not for coordination messages
- No way to ask supervisor a question or request task assignment
- PR-based workflow doesn't support real-time back-and-forth

**Enhancement opportunity:** A convention file (e.g., `.multiclaude/remote-requests.yaml`) in the repo could serve as a structured communication channel. External sessions push requests to a branch, agents poll for changes. This is clunky but entirely possible.

### 4. MCP Server Bridge

**Mechanism:** Run an MCP server that exposes multiclaude messaging as tools callable by any Claude Code session

**Feasibility: HIGH (with development effort)**

**Concept:**
```
Remote Claude Code ←→ MCP Server ←→ multiclaude CLI / state files
```

An MCP server could expose tools like:
- `send_message(recipient, body)` — wraps `multiclaude message send`
- `list_messages()` — wraps `multiclaude message list`
- `ack_message(id)` — wraps `multiclaude message ack`
- `get_agent_status()` — reads state.json
- `list_open_prs()` — wraps `gh pr list`
- `get_roadmap()` — reads ROADMAP.md

**Implementation approaches:**

| Approach | Effort | Network? | Security |
|----------|--------|----------|----------|
| Local MCP (stdio) on same machine | Low | No | Low risk — same user |
| HTTP MCP server on LAN | Medium | Yes | Needs auth |
| HTTP MCP server over internet | Medium | Yes | Needs auth + TLS |
| SSH tunnel to local MCP | Low-Medium | Yes | SSH handles auth |

**Strengths:**
- Native Claude Code integration — tools appear in the agent's tool list
- Structured request/response (not fire-and-forget like CLI messages)
- Could expose rich operations beyond just messaging
- MCP is the standard Claude Code extensibility mechanism

**Weaknesses:**
- Requires building and maintaining the MCP server
- Security considerations for network-exposed servers
- MCP servers are configured per-project in `.claude/settings.json` or globally — remote Claude Code needs the right config
- Adds infrastructure complexity

**This is the most promising approach for robust remote collaboration**, but requires development investment.

### 5. Existing multiclaude Features for Remote Participation

**What exists today:**

| Feature | Remote-friendly? | Notes |
|---------|-------------------|-------|
| `multiclaude message send` | Same-machine only | Requires CLI access |
| `multiclaude agent attach` | Via SSH | Read-only monitoring |
| `multiclaude worker create` | Same-machine only | Spawns local tmux window |
| `multiclaude status` | Same-machine only | Shows agent overview |
| `multiclaude logs` | Via SSH/scp | Can tail agent logs |
| GitHub PR workflow | Fully remote | Standard git workflow |

**What doesn't exist:**
- No network API or HTTP endpoint for multiclaude
- No remote agent registration mechanism
- No way for an external agent to "join" the multiclaude session
- No webhook or push-notification system for messages
- No MCP bridge to multiclaude operations

## Recommended Approaches (Ranked)

### Tier 1: Works Today, Zero Development

**A. Same-machine standalone Claude Code + message CLI**
- Start a Claude Code session in any terminal
- Use `multiclaude message send supervisor "your request"` via Bash
- Poll replies with `multiclaude message list`
- **Latency:** 5-30 minutes per round-trip (agent polling intervals)
- **Best for:** Async task requests, status queries, one-off contributions

**B. SSH + message CLI for remote machines**
- SSH into the multiclaude host
- Run `multiclaude message send` commands
- Or: `ssh host "multiclaude message send supervisor 'your request'"`
- Same latency as above, but works from anywhere with SSH access

**C. Git-based collaboration**
- Push branches and open PRs from any machine
- multiclaude agents process PRs normally
- Use PR comments for coordination (agents monitor PR state)

### Tier 2: Low Development Effort

**D. Wrapper script for bidirectional messaging**
- A shell script that sends a message and polls for replies
- Could reduce the manual polling burden
- Example: `remote-collab.sh send supervisor "Review my branch?" --wait-reply 60`

**E. tmux + message combo for semi-live collaboration**
- SSH + tmux attach for real-time monitoring
- Message CLI for structured communication
- Best of both worlds: see what agents are doing AND send them requests

### Tier 3: Significant Development (Best Long-term)

**F. MCP server bridge**
- Build an MCP server wrapping multiclaude operations
- Remote Claude Code sessions configure it as a tool source
- Enables native tool-based interaction with the multiclaude system
- Could be HTTP-based for true remote access, or SSH-tunneled for security
- **Estimated scope:** ~200-400 lines of Go/Python, plus auth layer for network access

**G. multiclaude daemon HTTP API**
- Add an HTTP API to the daemon alongside the Unix socket
- Expose endpoints: `/messages`, `/agents`, `/status`, `/workers`
- Any HTTP client (including remote Claude Code via Bash/curl) can interact
- **Estimated scope:** Significant — modifying multiclaude core

## Security Considerations

| Approach | Risk Level | Mitigation |
|----------|-----------|------------|
| Same-machine CLI | Low | OS-level user permissions |
| SSH tunneling | Low | SSH auth + key management |
| HTTP MCP on LAN | Medium | Auth tokens, IP allowlisting |
| HTTP MCP on internet | High | TLS, auth, rate limiting, input validation |
| Direct daemon socket exposure | High | Not recommended — redesign needed |

**Critical:** Any network-exposed interface must authenticate requests. The multiclaude message system has no auth — it trusts any process that can write to the message directory. Network exposure without auth would allow arbitrary message injection.

## Rejected / Not Recommended

1. **Typing into tmux windows** — Sends raw keystrokes to Claude Code, unpredictable and dangerous
2. **Direct state.json manipulation** — Race conditions with daemon, could corrupt agent state
3. **Exposing daemon.sock over network** — Unix socket protocol not designed for network use
4. **Polling GitHub issues as a backchannel** — Too slow, wrong abstraction, clutters issue tracker

## Conclusion

**For immediate use:** Same-machine Claude Code sessions can collaborate today via `multiclaude message send/list/ack`. Combined with SSH for remote access, this covers most use cases with zero development.

**For robust remote collaboration:** An MCP server bridge is the natural evolution. It would give remote Claude Code sessions native tool access to multiclaude operations, with proper request/response semantics. This is the recommended investment if remote collaboration becomes a recurring need.

**Key gap:** The multiclaude messaging system is polling-based with 5-30 minute intervals. Any real-time collaboration would benefit from a push notification mechanism or shorter polling intervals for agents expecting replies.
