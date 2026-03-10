# Remote Collaboration Feasibility Analysis — multiclaude Architecture

**Date:** 2026-03-10
**Analyst:** witty-raccoon (BMAD Analyst)
**Scope:** Technical feasibility of remote collaboration with a multiclaude supervisor session

---

## Executive Summary

multiclaude is a **single-machine orchestrator** with no built-in remote access layer. However, its architecture — file-based messaging, Unix socket daemon, tmux sessions, and CLI-driven operations — creates multiple viable integration surfaces for remote collaboration. The most practical near-term approach combines **SSH tunneling for CLI access** with **file sync for message delivery**. A purpose-built remote relay would require upstream changes to multiclaude.

---

## 1. Message System Architecture

### How It Works

multiclaude uses a **file-based message passing system** with daemon-mediated delivery:

- **Storage:** `~/.multiclaude/messages/<repo>/<agent>/msg-<uuid>.json`
- **Format:** JSON files with fields: `id`, `from`, `to`, `timestamp`, `body`, `status`
- **Delivery:** The daemon polls periodically (~60s tick), routes messages from sender directories to recipient directories, and updates status to `"delivered"`
- **Consumption:** Agents poll via `multiclaude message list` and `multiclaude message read <id>`. Acknowledgment via `multiclaude message ack <id>`

### Observed Message Flow

```
1. Agent calls: multiclaude message send <recipient> "<body>"
2. CLI writes JSON to: ~/.multiclaude/messages/<repo>/<recipient>/msg-<uuid>.json
3. Daemon tick detects new messages → routes & marks "delivered"
4. Daemon wakes recipient agent (nudge via tmux send-keys)
5. Recipient reads message on next poll
```

### Can External Sources Inject Messages?

**Yes, trivially.** The message system is purely file-based with no authentication or signature verification. Any process that can write a correctly-formatted JSON file to the correct directory can inject a message that the daemon will deliver. The daemon simply reads files from `~/.multiclaude/messages/<repo>/<agent>/` — it does not verify the `from` field or check provenance.

**Required format:**
```json
{
  "id": "msg-<uuid>",
  "from": "remote-user",
  "to": "supervisor",
  "timestamp": "2026-03-10T12:00:00-05:00",
  "body": "Message content here",
  "status": "pending"
}
```

**Risk:** No access control. Any process with filesystem write access can impersonate any agent. This is a design feature (simplicity) not a bug — multiclaude assumes a trusted single-user environment.

---

## 2. Tmux Session Architecture

### Structure

multiclaude creates one tmux session per repository: `mc-<RepoName>` (e.g., `mc-ThreeDoors`).

```
mc-ThreeDoors (session)
├── supervisor        (window 0) — Claude Code process
├── workspace         (window 1) — workspace shell
├── merge-queue       (window 2) — persistent agent
├── pr-shepherd       (window 3) — persistent agent
├── arch-watchdog     (window 4) — persistent agent
├── envoy             (window 5) — persistent agent
├── project-watchdog  (window 6) — persistent agent
├── jolly-raccoon     (window 7) — worker
├── gentle-panda      (window 8) — worker
├── zealous-koala     (window 9) — worker
├── witty-raccoon     (window 10) — worker
└── proud-eagle       (window 11) — worker
```

Each window runs a Claude Code process (`claude`) with a specific prompt/system message injected at spawn time.

### SSH Sharing Feasibility

**tmux natively supports remote attachment over SSH.** This is one of tmux's core use cases.

```bash
# From remote machine:
ssh user@host -t "tmux attach -t mc-ThreeDoors"

# Read-only attachment (observe without interference):
ssh user@host -t "tmux attach -t mc-ThreeDoors -r"

# Attach to specific window:
ssh user@host -t "tmux select-window -t mc-ThreeDoors:supervisor && tmux attach -t mc-ThreeDoors"
```

**Caveats:**
- tmux socket is at `/private/tmp/tmux-503/default` (per `$TMUX` env var)
- Socket permissions are user-only (uid 503) — SSH must authenticate as the same user
- Multiple clients can attach simultaneously (tmux supports this natively)
- Remote user sees real-time agent activity and can interact with any window
- **Risk:** Direct interaction with agent windows can confuse Claude Code's conversation state

### multiclaude's `attach` Command

`multiclaude agent attach <name> [--read-only]` wraps tmux attachment. This could theoretically be invoked over SSH:

```bash
ssh user@host "multiclaude agent attach supervisor --read-only"
```

---

## 3. File-Based Communication Channels

### State File: `~/.multiclaude/state.json`

- **Purpose:** Central registry of all repos, agents, worktrees, sessions, PIDs
- **Updated by:** Daemon (via Unix socket requests from CLI commands)
- **Format:** JSON with nested structure: `repos.<name>.agents.<name>.{type, worktree_path, tmux_window, session_id, pid, task, created_at, last_nudge}`
- **Sync potential:** Read-only mirroring is safe. Write conflicts would corrupt state — only the daemon should write this file.

### Message Files: `~/.multiclaude/messages/<repo>/<agent>/*.json`

- **Sync potential:** HIGH. These are the primary inter-agent communication channel and are designed for file-based read/write. A remote process could:
  - Write new message files to inject tasks/commands
  - Read message files to monitor agent communication
  - Use rsync, Syncthing, or Unison for bidirectional sync

### Daemon Log: `~/.multiclaude/daemon.log`

- **Purpose:** Operational log with DEBUG/INFO/WARN entries
- **Sync potential:** Read-only tail for remote monitoring
- **Content:** Agent health checks, message routing, worktree refreshes, branch cleanup

### Daemon Socket: `~/.multiclaude/daemon.sock`

- **Type:** Unix domain socket (not TCP)
- **Protocol:** Custom request/response (observed commands: `list_agents`, `get_repo_config`, `add_agent`, `route_messages`, `trigger_refresh`, `trigger_cleanup`)
- **Sync potential:** NONE directly. Unix sockets are local-only. Would require a TCP proxy or SSH tunnel to expose remotely.

### Worktrees: `~/.multiclaude/wts/<repo>/<agent>/`

- **Purpose:** Isolated git worktrees per worker agent
- **Content:** Full git working copies, one per agent
- **Sync potential:** These are standard git repos — agents push to remote, so code changes naturally propagate via git

### Output Directory: `~/.multiclaude/output/<repo>/`

- **Purpose:** Agent output logs (Claude Code conversation transcripts)
- **Sync potential:** Read-only monitoring of agent activity

---

## 4. Remotely Executable CLI Commands

### Commands Safe to Run via SSH

All multiclaude CLI commands communicate with the daemon via the Unix socket. Over SSH, these work natively:

| Command | Purpose | Remote Safety |
|---------|---------|---------------|
| `multiclaude status` | System overview | Safe (read-only) |
| `multiclaude worker list` | List active workers | Safe (read-only) |
| `multiclaude message list` | List pending messages | Safe (read-only) |
| `multiclaude message send <to> <msg>` | Send inter-agent message | Safe (writes message file) |
| `multiclaude message read <id>` | Read a message | Safe (read-only) |
| `multiclaude message ack <id>` | Acknowledge a message | Safe (updates file) |
| `multiclaude diagnostics --json` | Full system diagnostics | Safe (read-only) |
| `multiclaude repo list` | List repositories | Safe (read-only) |
| `multiclaude repo history` | Task history | Safe (read-only) |
| `multiclaude logs [agent]` | View agent logs | Safe (read-only) |
| `multiclaude logs search <pattern>` | Search logs | Safe (read-only) |
| `multiclaude daemon status` | Daemon health | Safe (read-only) |

### Commands That Modify State (Use with Caution)

| Command | Purpose | Risk |
|---------|---------|------|
| `multiclaude worker create <task>` | Spawn new worker | Creates worktree, tmux window, Claude process |
| `multiclaude agent restart <name>` | Restart crashed agent | Kills/respawns Claude process |
| `multiclaude agent complete` | Signal task done | Updates agent state |
| `multiclaude refresh` | Sync worktrees | Git operations on all worktrees |

### SSH Invocation Pattern

```bash
# Remote monitoring
ssh user@host "multiclaude status"
ssh user@host "multiclaude message list"
ssh user@host "multiclaude diagnostics --json"

# Remote task dispatch
ssh user@host 'multiclaude worker create "Implement feature X"'

# Remote message injection
ssh user@host 'multiclaude message send supervisor "Priority: Review PR #123"'

# Remote log tailing
ssh user@host "multiclaude logs supervisor -f"
```

---

## 5. Feasibility Assessment

### Approach A: SSH Direct Access (Highest Feasibility)

**How:** Remote user SSHs into the host machine, runs multiclaude CLI commands directly.

| Aspect | Rating | Notes |
|--------|--------|-------|
| Setup complexity | LOW | Just SSH access + multiclaude in PATH |
| Latency | LOW | Sub-second for CLI commands |
| Bidirectional | YES | Full CLI + tmux access |
| Authentication | SSH keys | Existing infrastructure |
| Observation | FULL | tmux attach for real-time viewing |
| Task dispatch | FULL | `multiclaude worker create` works over SSH |
| Message passing | FULL | `multiclaude message send` works over SSH |

**Limitations:** Requires SSH access to the host. Single point of failure (host machine). No offline/async capability beyond git.

### Approach B: File Sync for Messages (Medium Feasibility)

**How:** Sync `~/.multiclaude/messages/` between machines using rsync, Syncthing, or Unison.

| Aspect | Rating | Notes |
|--------|--------|-------|
| Setup complexity | MEDIUM | Needs file sync infrastructure |
| Latency | MEDIUM | Depends on sync interval (1-60s typical) |
| Bidirectional | PARTIAL | Messages yes, but no state.json writes |
| Authentication | Sync tool | Varies by tool |
| Observation | LIMITED | Messages only, no tmux view |

**Limitations:** Only covers message passing. No agent spawning, no real-time observation. Sync conflicts possible if both sides write simultaneously.

### Approach C: Git-Based Coordination (Low-Medium Feasibility)

**How:** Use a shared git repo branch as a communication channel. Remote user pushes "task request" files; supervisor polls for them.

| Aspect | Rating | Notes |
|--------|--------|-------|
| Setup complexity | LOW | Just git push/pull |
| Latency | HIGH | Minutes (git poll intervals) |
| Bidirectional | YES | Both sides can push |
| Authentication | Git/GitHub | Existing infrastructure |
| Observation | LIMITED | Only committed artifacts visible |

**Limitations:** High latency. Requires polling. Not suitable for real-time collaboration.

### Approach D: TCP Proxy for Daemon Socket (Medium-High Feasibility)

**How:** Expose the Unix socket over TCP using `socat` or SSH port forwarding.

```bash
# On host machine:
socat TCP-LISTEN:9999,reuseaddr,fork UNIX-CONNECT:~/.multiclaude/daemon.sock

# Or via SSH tunnel:
ssh -L 9999:~/.multiclaude/daemon.sock user@host
```

| Aspect | Rating | Notes |
|--------|--------|-------|
| Setup complexity | MEDIUM | Needs socket proxy |
| Latency | LOW | Real-time socket communication |
| Bidirectional | FULL | Full daemon API access |
| Authentication | SSH tunnel | Existing infrastructure |

**Limitations:** Requires understanding the daemon's internal protocol (not documented). Binary protocol, not REST. Would need a client that speaks the daemon wire format.

---

## 6. Architecture Constraints & Risks

### Hard Constraints

1. **Single daemon per machine.** multiclaude assumes one daemon manages all repos on a single host. No clustering or distributed daemon support.
2. **Unix socket only.** The daemon listens on a Unix domain socket, not TCP. No built-in remote API.
3. **No authentication on messages.** Any process with filesystem access can inject messages. Over SSH this is fine (SSH provides auth), but over a network sync it's a security concern.
4. **Claude Code is the agent runtime.** Each agent is a Claude Code process — you can't run agents on a different machine without also running Claude Code there (which requires its own Anthropic API key and session).
5. **tmux is the UI layer.** All agent interaction is mediated through tmux windows. There's no web UI, REST API, or other remote-friendly interface.

### Soft Constraints

1. **State.json is daemon-exclusive.** Only the daemon should write to state.json. Remote processes should treat it as read-only.
2. **Worktree management is local.** git worktrees are filesystem-local; they can't span machines.
3. **Agent PIDs are local.** The daemon tracks agent processes by PID — meaningless across machines.

### Security Considerations

- SSH access grants full control over the multiclaude session
- Message injection requires no additional auth beyond filesystem access
- No audit trail for who sent messages (the `from` field is self-declared)
- Agent prompt files could be modified to inject instructions (no integrity checking)
- Daemon socket is user-only (`srw-------`) but not otherwise protected

---

## 7. Recommended Architecture for Remote Collaboration

### Tier 1: Quick Win (SSH + CLI)

```
Remote Machine                    Host Machine
┌──────────────┐    SSH          ┌─────────────────────┐
│ Claude Code   │───────────────▶│ multiclaude daemon   │
│ (standalone)  │                │ ├── supervisor       │
│               │  SSH commands: │ ├── merge-queue      │
│ ssh host      │  - status      │ ├── pr-shepherd      │
│   multiclaude │  - message     │ ├── workers...       │
│   message     │  - worker      │ └── tmux session     │
│   send ...    │  - attach      │                      │
└──────────────┘                └─────────────────────┘
```

**Implementation:** Zero multiclaude changes. Just SSH.

### Tier 2: Structured Remote Dispatch (Git + Convention)

```
Remote Machine                    Host Machine
┌──────────────┐                ┌─────────────────────┐
│ Claude Code   │    git push   │ multiclaude daemon   │
│               │──────────────▶│                      │
│ Writes task   │               │ Supervisor polls     │
│ request to    │               │ .multiclaude/remote/ │
│ branch file   │               │ for task requests    │
│               │◀──────────────│                      │
│ Reads results │    git pull   │ Workers push results │
└──────────────┘                └─────────────────────┘
```

**Implementation:** Requires a convention for task request files and supervisor polling. No multiclaude code changes, but needs agent definition updates.

### Tier 3: Full Remote API (Requires Upstream Changes)

```
Remote Machine                    Host Machine
┌──────────────┐                ┌─────────────────────┐
│ multiclaude   │  HTTP/gRPC    │ multiclaude daemon   │
│ remote client │──────────────▶│ + TCP listener       │
│               │               │ + auth middleware    │
│               │◀──────────────│ + API endpoints      │
│ Full CLI      │  WebSocket    │ + event streaming    │
│ experience    │               │                      │
└──────────────┘                └─────────────────────┘
```

**Implementation:** Requires significant multiclaude changes: TCP listener, auth layer, API endpoints, event streaming. This would be a feature request to the upstream `dlorenc/multiclaude` project.

---

## 8. Conclusions

| Question | Answer |
|----------|--------|
| Can external sources inject messages? | Yes — write JSON to the messages directory |
| Can tmux be shared via SSH? | Yes — native tmux feature, works immediately |
| Can state.json be synced? | Read-only yes; writes must go through daemon only |
| Can CLI commands run remotely? | Yes — all commands work over SSH |
| Is real-time remote collab possible today? | Yes, via SSH (Tier 1) |
| Does multiclaude need changes for basic remote? | No — SSH provides everything needed |
| Does multiclaude need changes for rich remote? | Yes — Tier 3 requires upstream work |

**Bottom line:** The simplest and most effective remote collaboration approach is SSH. multiclaude's architecture is well-suited to this because all operations are CLI-driven and tmux-based — both of which work transparently over SSH. The message system's file-based nature is a bonus: it means messages can be injected by any authorized process without needing to speak a proprietary protocol.

For a more sophisticated remote experience (web UI, multi-machine orchestration, event streaming), multiclaude would need upstream changes to expose a network-accessible API. This is feasible given the clean daemon architecture but represents significant new scope.
