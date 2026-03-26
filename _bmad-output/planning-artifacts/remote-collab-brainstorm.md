# Remote Claude Code Collaboration with Multiclaude — Brainstorm

> **Session:** 2026-03-10 | **Facilitator:** Carson (Brainstorming Coach)
> **Scenario:** User A runs multiclaude supervisor on Machine A. User B (or User A remotely) runs Claude Code on Machine B. How can Machine B interact with Machine A's multiclaude system?

## Current Architecture Constraints

Before brainstorming, key facts about multiclaude's current design:

- **State:** Single `state.json` file with atomic writes, `sync.RWMutex`
- **Messages:** Filesystem-based JSON files in `~/.multiclaude/messages/<repo>/<agent>/`
- **IPC:** Unix domain socket between CLI and daemon (local only)
- **Visibility:** tmux sessions (local only)
- **Workers:** Git worktrees in `~/.multiclaude/wts/` with auto-refresh
- **Daemon loops:** 2-minute intervals for health, messages, nudges
- **DESIGN.md non-goals:** No remote daemon, no web dashboard, no multi-user support

ThreeDoors also has a full **MCP server** (`cmd/threedoors-mcp/`) with stdio and SSE transports, middleware chain, and tool/resource/prompt capabilities.

---

## Category 1: Network Tunnel Approaches

### 1.1 SSH Tunnel to Unix Socket

**Concept:** Forward multiclaude's daemon socket over SSH.

```
Machine B:  claude-code → multiclaude CLI → local socket
                                              ↓ (SSH tunnel)
Machine A:                                  daemon.sock → daemon
```

- `ssh -L /tmp/mc-remote.sock:/Users/skippy/.multiclaude/daemon.sock machineA`
- Set `MULTICLAUDE_SOCKET=/tmp/mc-remote.sock` on Machine B
- All CLI commands work transparently

**Pros:** Zero code changes if socket path is configurable. Full CLI compatibility.
**Cons:** Requires SSH access. Socket forwarding is finicky. tmux visibility lost. Latency on every command.

### 1.2 Tailscale/WireGuard Mesh + Wrapped Socket

**Concept:** Put both machines on a private mesh network. Wrap the Unix socket in a TCP listener.

- Daemon listens on `100.x.y.z:9876` (Tailscale IP) in addition to Unix socket
- Machine B's CLI connects to TCP endpoint
- Tailscale handles auth, encryption, NAT traversal automatically

**Pros:** Feels like local. Tailscale ACLs provide auth. Works from phones/tablets.
**Cons:** Requires daemon code changes (TCP listener). Still no tmux visibility remotely.

### 1.3 SSH to tmux (Direct Attach)

**Concept:** Machine B SSHes into Machine A and attaches to the multiclaude tmux session.

```bash
ssh machineA -t "tmux attach -t mc-ThreeDoors"
```

- Full visibility into all agent windows
- Can type into supervisor window directly
- Read-only mode: `tmux attach -t mc-ThreeDoors -r`

**Pros:** Zero code changes. Full visibility. Already works today.
**Cons:** Not "Claude Code on Machine B" — it's terminal sharing. No local Claude context. Can't run local tools.

---

## Category 2: Git as the Communication Layer

### 2.1 Git-Backed Message Queue

**Concept:** Use a dedicated git branch (e.g., `multiclaude/messages`) as a distributed message queue.

- Machine B pushes message files to the branch
- Machine A's daemon polls the branch (or uses webhooks)
- Messages follow the existing JSON format
- Responses pushed back to the branch

**Pros:** Works anywhere with git access. Audit trail built in. No infrastructure.
**Cons:** Git isn't a message queue — polling latency, merge conflicts on concurrent writes. Noisy git history.

### 2.2 GitHub Issues/Discussions as Command Channel

**Concept:** Machine B creates specially-formatted GitHub issues that Machine A's envoy agent interprets as commands.

```markdown
Title: [multiclaude:dispatch] Implement story 42.3
Labels: multiclaude-command
Body: /implement-story 42.3
```

- Envoy already patrols issues — add command parsing
- Responses posted as issue comments
- Status tracked via labels (`dispatched`, `in-progress`, `complete`)

**Pros:** GitHub UI is the remote interface. Works from any device. Full audit trail. Notification system built in.
**Cons:** High latency (webhook + processing). Limited to text commands. Public visibility concerns for private repos.

### 2.3 Git Notes as Sideband Channel

**Concept:** Use `git notes` to attach metadata to commits without modifying history.

- Machine B adds notes to a sentinel commit: `git notes add -m '{"cmd":"status"}' HEAD`
- Machine A's daemon reads notes on fetch
- Lightweight, doesn't pollute branches or history

**Pros:** Elegant. No extra branches. Part of git protocol.
**Cons:** Notes don't auto-sync (need explicit fetch/push). Fragile. Poor tooling support.

---

## Category 3: MCP Server Bridge

### 3.1 Multiclaude MCP Server (New)

**Concept:** Build an MCP server that exposes multiclaude operations as tools. Machine B's Claude Code connects to it.

```
Machine B: Claude Code → MCP client
                           ↓ (SSE/stdio over network)
Machine A: MCP Server → multiclaude daemon
```

**Exposed tools:**
- `dispatch_worker(task, story_id)` — spawn a worker
- `send_message(recipient, body)` — inter-agent messaging
- `get_status()` — system status overview
- `list_agents()` — active agents and their state
- `get_pr_status(number)` — PR review status
- `read_agent_log(name, lines)` — recent agent output

**Pros:** Native Claude Code integration. Machine B's Claude "just knows" about multiclaude. SSE transport already proven in ThreeDoors MCP server. Type-safe tool definitions.
**Cons:** Requires building the MCP server. Network security for SSE endpoint. Auth model needed.

### 3.2 Bidirectional MCP Bridge

**Concept:** Both machines run MCP servers. They discover each other and relay capabilities.

- Machine A's MCP server exposes multiclaude operations
- Machine B's MCP server exposes local file access, test running, etc.
- Claude Code on either machine sees tools from both

**Pros:** True collaboration — both sides contribute capabilities. Symmetrical design.
**Cons:** Complex. Discovery/auth problems. Tool namespace collisions.

### 3.3 MCP over Tailscale with mTLS

**Concept:** Combine 1.2 and 3.1 — MCP server on Tailscale mesh with mutual TLS.

- Tailscale handles networking and identity
- mTLS ensures only authorized Claude Code instances connect
- MCP SSE transport works over HTTPS

**Pros:** Production-grade security. Claude Code native. Low latency on mesh.
**Cons:** Infrastructure overhead. Certificate management.

---

## Category 4: Cloud Message Queues

### 4.1 Redis Pub/Sub Bridge

**Concept:** Both machines connect to a shared Redis instance. Commands and responses flow through pub/sub channels.

- Channel per agent: `mc:ThreeDoors:supervisor`, `mc:ThreeDoors:merge-queue`
- Machine B publishes commands, subscribes to responses
- Daemon bridges Redis channels to filesystem messages

**Pros:** Real-time. Battle-tested. Rich data structures for state sync.
**Cons:** Requires Redis infrastructure. Another moving part. Connection management.

### 4.2 NATS as Lightweight Backbone

**Concept:** NATS provides publish/subscribe + request/reply + JetStream persistence.

- `mc.ThreeDoors.dispatch` — worker dispatch requests
- `mc.ThreeDoors.status` — system status queries
- `mc.ThreeDoors.messages.>` — agent message routing
- JetStream for durable message delivery

**Pros:** Designed exactly for this. Tiny binary. Built-in auth. Subject-based routing maps perfectly to multiclaude's agent model.
**Cons:** Another service to run. Learning curve. Overkill for 2 machines?

### 4.3 MQTT for Minimal Overhead

**Concept:** MQTT broker (Mosquitto) with topic hierarchy matching agent structure.

- `mc/ThreeDoors/agents/+/status` — agent status updates
- `mc/ThreeDoors/commands/dispatch` — worker dispatch
- QoS levels map to message importance
- Retained messages for status queries

**Pros:** Extremely lightweight. Works on constrained networks. Great for mobile access.
**Cons:** Limited message semantics. No built-in request/reply. Broker required.

### 4.4 Cloudflare Pub/Sub (Serverless)

**Concept:** Use Cloudflare's MQTT-compatible pub/sub — no server to manage.

**Pros:** Zero infrastructure. Global edge network. Free tier generous.
**Cons:** Vendor lock-in. Latency uncertainty. Limited control.

---

## Category 5: GitHub Actions as Relay

### 5.1 Workflow Dispatch API

**Concept:** Machine B triggers GitHub Actions workflows that execute multiclaude commands on Machine A via self-hosted runner.

```
Machine B: gh workflow run multiclaude-dispatch.yml -f command="work 'Fix bug #123'"
                           ↓
GitHub Actions: self-hosted runner on Machine A
                           ↓
Machine A: multiclaude work "Fix bug #123"
```

- Self-hosted runner already has access to Machine A
- Workflow outputs streamed back via Actions logs
- `gh run watch` for real-time status

**Pros:** No custom infrastructure. GitHub auth handles security. Audit trail. Works from any `gh` client.
**Cons:** Actions latency (10-30s startup). Self-hosted runner security. Workflow file maintenance. Rate limits.

### 5.2 Repository Dispatch Events

**Concept:** Machine B sends repository dispatch events; Machine A's runner processes them.

```bash
gh api repos/arcavenae/ThreeDoors/dispatches \
  -f event_type=multiclaude-command \
  -f client_payload='{"action":"dispatch","task":"Fix bug #123"}'
```

**Pros:** Simpler than workflow dispatch. Event-driven.
**Cons:** Fire-and-forget — no response channel without polling. Limited payload size.

---

## Category 6: Shared Filesystem

### 6.1 SSHFS Mount

**Concept:** Mount Machine A's `~/.multiclaude/` on Machine B via SSHFS.

- Machine B's CLI reads/writes directly to Machine A's filesystem
- Message files, state.json, worktrees all accessible
- Daemon on Machine A processes normally

**Pros:** Transparent. No code changes if paths match.
**Cons:** Filesystem latency. Lock contention on state.json. SSHFS reliability issues. Git operations over SSHFS are slow/broken.

### 6.2 Syncthing Continuous Sync

**Concept:** Syncthing keeps `~/.multiclaude/messages/` in sync between machines.

- Machine B writes messages locally; Syncthing replicates to Machine A
- Responses replicate back
- Selective sync — only messages directory, not worktrees

**Pros:** Peer-to-peer. No server. Encrypted. Handles conflicts.
**Cons:** Sync latency (seconds to minutes). Message ordering not guaranteed. Partial sync during writes.

### 6.3 iCloud/Dropbox Shared State

**Concept:** Use cloud storage to sync the messages directory.

**Pros:** Zero setup if already using the service.
**Cons:** Unreliable latency. File locking issues. Not designed for IPC.

---

## Category 7: Webhook/API Endpoints

### 7.1 Lightweight HTTP API on Daemon

**Concept:** Add an HTTP API to the multiclaude daemon alongside the Unix socket.

```
POST /api/v1/dispatch    {"task": "Fix bug #123"}
GET  /api/v1/status
GET  /api/v1/agents
POST /api/v1/message     {"to": "supervisor", "body": "..."}
GET  /api/v1/agent/:name/logs?lines=50
```

- Token-based auth (bearer token in `~/.multiclaude/remote-token`)
- Optional TLS with self-signed cert
- Expose via Tailscale, ngrok, or Cloudflare Tunnel

**Pros:** Standard REST. Any HTTP client works. Easy to secure.
**Cons:** Requires daemon code changes. Auth/TLS complexity. Port management.

### 7.2 ngrok/Cloudflare Tunnel Instant Expose

**Concept:** Wrap the HTTP API (7.1) with ngrok or Cloudflare Tunnel for instant public access.

```bash
# Machine A
ngrok http 9876 --auth-token=xxx

# Machine B
multiclaude --remote https://abc123.ngrok.io
```

**Pros:** Instant. No firewall changes. HTTPS automatic.
**Cons:** Dependency on third-party service. Latency. URL changes on restart (unless paid).

### 7.3 Cloudflare Workers as API Gateway

**Concept:** Cloudflare Worker acts as API gateway; Machine A connects via WebSocket.

- Worker provides stable URL and auth
- Machine A maintains persistent WebSocket to Worker
- Machine B sends HTTP requests to Worker
- Worker relays between HTTP and WebSocket

**Pros:** Stable URL. Edge auth. No port forwarding. Free tier.
**Cons:** WebSocket complexity. Cloudflare dependency. Debugging is harder.

---

## Category 8: Claude Code Hooks & Extensions

### 8.1 Pre/Post Hook Bridge

**Concept:** Configure Claude Code hooks on Machine B that relay actions to Machine A.

```json
// .claude/hooks.json on Machine B
{
  "post_tool_call": {
    "bash": "curl -X POST https://machine-a/api/v1/notify -d '{\"event\": \"$TOOL\", \"result\": \"$RESULT\"}'"
  }
}
```

- Every tool call on Machine B gets relayed to Machine A
- Supervisor sees what remote Claude is doing
- Can trigger responses (e.g., "don't touch that file, it's being worked on")

**Pros:** Lightweight integration. No Claude Code modifications needed.
**Cons:** Hook API limited. One-way without polling. Fragile string interpolation.

### 8.2 Custom Claude Code Slash Commands

**Concept:** Define slash commands on Machine B that interact with Machine A.

```
/remote-status → curl Machine A's API, display formatted result
/remote-dispatch "task" → POST to Machine A's dispatch endpoint
/remote-messages → GET pending messages from Machine A
```

- Commands defined in `.claude/commands/` on Machine B
- Use Bash tool under the hood to call Machine A's API

**Pros:** Native Claude Code UX. Discoverable. Self-documenting.
**Cons:** Requires the API endpoint (7.1) to exist. Commands are repo-specific.

### 8.3 Shared MCP Server Config via Git

**Concept:** Store MCP server configuration in the repo. All collaborators auto-connect to the same MCP servers.

```yaml
# .claude/mcp-servers.json (checked into repo)
{
  "multiclaude-remote": {
    "type": "sse",
    "url": "https://mc.tailnet:9876/mcp",
    "auth": "tailscale-identity"
  }
}
```

- Clone repo → automatically get multiclaude MCP tools
- Zero manual setup for new collaborators

**Pros:** Git-native distribution. Zero onboarding friction.
**Cons:** URL must be stable. Auth tokens can't be in git.

---

## Category 9: Anthropic-Native Multi-Agent

### 9.1 Anthropic Agent SDK Mesh

**Concept:** Use Anthropic's Agent SDK to create a mesh of Claude instances that communicate natively.

- Each machine runs an Agent SDK node
- Agents discover each other via a coordination service
- Anthropic handles the transport and message format
- multiclaude becomes one node in a larger agent mesh

**Pros:** If Anthropic builds this, it would be the canonical solution. Native tool sharing. Conversation context bridging.
**Cons:** Doesn't exist yet (as of 2026-03). Would require Anthropic to add multi-node agent coordination.

### 9.2 Shared Conversation via API

**Concept:** Machine B's Claude Code session includes context from Machine A via API calls.

- Machine B periodically fetches Machine A's supervisor context
- Injected as system prompt or tool results
- Machine B's Claude "knows" what Machine A is doing

**Pros:** Lightweight. No infrastructure changes.
**Cons:** Context window pollution. Stale data. No bidirectional action.

### 9.3 Tool Use Proxy (Claude-to-Claude)

**Concept:** Machine B's Claude defines a "remote_multiclaude" tool that calls Machine A's Claude via API.

```
Machine B Claude → tool_use: remote_multiclaude("check status")
                     ↓ (API call)
Machine A Claude → executes locally → returns result
                     ↓
Machine B Claude ← tool_result: "3 agents active, 2 PRs pending"
```

**Pros:** Claude-native. Rich responses. Context-aware.
**Cons:** Double API cost. Latency. Auth complexity. Recursive agent risk.

---

## Category 10: Wild Ideas

### 10.1 Git Push Hooks as RPC

**Concept:** Machine B pushes to a special branch. Machine A's `post-receive` hook parses the commit message as a command.

```bash
git commit --allow-empty -m "MC:DISPATCH:implement story 42.3"
git push origin mc-commands
```

**Pros:** Works anywhere git works. No additional infrastructure.
**Cons:** Absurd abuse of git. Noisy. Slow. But it would work.

### 10.2 DNS TXT Records as Status Channel

**Concept:** Machine A publishes status to DNS TXT records on a domain it controls.

```
status.mc.example.com TXT "agents=5 prs=3 workers=2"
```

**Pros:** Globally accessible. Cached. Incredibly resilient.
**Cons:** Read-only. Update latency (TTL). Size limits. Truly unhinged.

### 10.3 QR Code Screen Sharing

**Concept:** Machine A renders system status as QR codes in tmux. Machine B's camera reads them.

**Pros:** Air-gapped compatible!
**Cons:** Peak absurdity. But technically works for status monitoring.

### 10.4 Shared Claude Memory

**Concept:** Both machines share a Claude Code memory directory (via git or sync).

- Machine A's supervisor writes status to `MEMORY.md`
- Machine B reads it at conversation start
- Machine B writes requests to a `REQUESTS.md`
- Machine A polls and processes

**Pros:** Already exists as a mechanism. Zero new code.
**Cons:** Not real-time. Manual polling. Memory has size limits.

### 10.5 Matrix/IRC Bot Bridge

**Concept:** Both machines connect to a Matrix room via bot. Commands sent as messages, responses as replies.

**Pros:** Real-time. Multi-user. Rich clients. End-to-end encryption.
**Cons:** Another service. Bot development. Message format complexity.

### 10.6 Bluetooth/Local Network Discovery

**Concept:** For same-LAN scenarios, use mDNS/Bonjour to discover multiclaude daemons.

```
_multiclaude._tcp.local. → 192.168.1.100:9876
```

**Pros:** Zero config on local network. Apple ecosystem friendly.
**Cons:** LAN only. No auth built in. Firewall issues.

---

## Evaluation Matrix

| Approach | Effort | Latency | Security | UX | Works Today? |
|----------|--------|---------|----------|----|-------------|
| 1.1 SSH Tunnel | Low | Low | High | Medium | Almost |
| 1.2 Tailscale + TCP | Medium | Low | High | High | No |
| 1.3 SSH tmux attach | Zero | Zero | High | Low | **YES** |
| 2.2 GitHub Issues | Low | High | Medium | High | Almost |
| 3.1 MCP Server | High | Low | Medium | **Best** | No |
| 4.2 NATS | Medium | Low | High | Medium | No |
| 5.1 GH Actions Relay | Low | High | High | Medium | Almost |
| 7.1 HTTP API | Medium | Low | Medium | High | No |
| 7.2 ngrok expose | Low | Medium | Low | High | No (needs 7.1) |
| 8.2 Slash commands | Low | Varies | Varies | High | Almost |
| 10.4 Shared Memory | Zero | High | High | Low | **YES** |

---

## Top Recommendations (if we were to build)

### Minimum Viable Remote (Today, Zero Code)

1. **SSH tmux attach** (1.3) for visibility
2. **SSH + multiclaude CLI** for commands: `ssh machineA "multiclaude work 'Fix bug'"`
3. **Shared Claude Memory** (10.4) via git for async coordination

### Best Single Investment

**MCP Server for Multiclaude** (3.1) — This gives the most bang for the buck:
- Native Claude Code integration (Machine B's Claude "sees" multiclaude tools)
- SSE transport already proven in ThreeDoors codebase
- Can layer Tailscale (1.2) for networking
- Extensible to multi-user scenarios
- Aligns with Anthropic's MCP ecosystem direction

### Best for GitHub-First Teams

**GitHub Issues as Command Channel** (2.2) + **GH Actions Relay** (5.1):
- No new infrastructure
- Works from any device with GitHub access
- Audit trail built in
- Envoy agent already patrols issues

### Best for Infrastructure-Comfortable Teams

**Tailscale Mesh** (1.2) + **HTTP API** (7.1) + **MCP Server** (3.1):
- Private network handles auth and encryption
- HTTP API for programmatic access
- MCP Server for Claude Code native integration
- Production-grade from day one

---

## Rejected / Deferred Ideas (with rationale)

| Idea | Why Deferred |
|------|-------------|
| SSHFS (6.1) | Git over SSHFS is unreliable; state.json lock contention |
| iCloud/Dropbox (6.3) | Sync latency and file locking make IPC unreliable |
| Bidirectional MCP (3.2) | Over-engineered for the problem; start with 3.1 |
| DNS TXT (10.2) | Read-only, high latency, limited payload — fun but impractical |
| QR Codes (10.3) | Peak absurdity — entertaining but not actionable |
| Git Notes (2.3) | Poor tooling support, fragile sync semantics |
| Redis (4.1) | Overkill for 2 machines; NATS is lighter if we go message queue |

---

## Next Steps (if pursuing)

1. **Validate demand** — Is this a real need or theoretical? How many users want remote collab?
2. **Prototype SSH CLI relay** — Wrap `ssh machineA "multiclaude ..."` in a local alias; test UX
3. **Design MCP server spec** — Define tool schemas for multiclaude operations
4. **Spike Tailscale + daemon TCP** — Minimal code change to add TCP listener alongside Unix socket
5. **RFC to multiclaude upstream** — Propose remote access as a feature area

---

*Generated by brainstorming session. All ideas are intentionally uncritical — evaluation and winnowing is a separate phase.*
