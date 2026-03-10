# Problem Solving Session: Remote Collaboration with multiclaude Supervisor

**Date:** 2026-03-10
**Problem Solver:** arcaven
**Problem Category:** Distributed Systems / Developer Tooling Architecture

---

## PROBLEM DEFINITION

### Initial Problem Statement

A person running Claude Code on a remote computer (different machine, different network) wants to actively collaborate with and send tasks to a multiclaude supervisor session running on another machine. There is no built-in mechanism for this today.

### Refined Problem Statement

multiclaude's architecture assumes all agents (supervisor, workers, persistent agents) share a single filesystem and tmux session on one machine. There is no built-in mechanism for a remote Claude Code instance to: (1) discover and connect to a running supervisor, (2) send structured task requests, (3) receive status updates and results, or (4) participate in the agent communication protocol (message send/ack). The gap is between "single-machine orchestration" and "distributed multi-machine collaboration."

### Problem Context

- multiclaude uses local filesystem state (`state.json`, worktrees) and local tmux
- The `multiclaude message send` protocol is IPC on a single machine
- Claude Code instances have no discovery mechanism for remote supervisors
- There is no authentication/authorization layer for remote agents
- NAT/firewalls typically block direct connections between machines on different networks
- The use case is increasingly common: distributed teams, cloud dev environments, remote pair programming with AI agents

### Success Criteria

1. Remote user can send task requests that the supervisor processes
2. Supervisor can dispatch results/status back to the remote user
3. Latency is acceptable for near-real-time collaboration (< 30 seconds)
4. Security: authenticated, encrypted communication
5. Works across NAT/firewalls without requiring VPN
6. Minimal changes to multiclaude's core architecture
7. Graceful degradation if connection drops

---

## DIAGNOSIS AND ROOT CAUSE ANALYSIS

### Problem Boundaries (Is/Is Not)

| Dimension | IS | IS NOT |
|---|---|---|
| **Where** | Cross-network, cross-machine | Same-machine, same-LAN only |
| **What** | Bidirectional task coordination | One-way monitoring or read-only access |
| **Who** | Remote Claude Code user to multiclaude supervisor | Two independent multiclaude instances federating |
| **When** | Real-time/near-real-time during active sessions | Batch/offline-only sync |
| **Scope** | Task dispatch, status updates, result delivery | Full filesystem or tmux session sharing (though those are candidate approaches) |

**Pattern:** The problem cleanly decomposes into (A) **transport** (how bytes flow between machines) and (B) **protocol** (what those bytes mean in multiclaude's world).

### Root Cause Analysis

**Five Whys:**

1. **Why can't a remote user collaborate?** — multiclaude uses local filesystem state and local tmux; there is no network interface
2. **Why is there no network interface?** — The message protocol (`multiclaude message send/list/read/ack`) is file-based IPC designed for single-machine use
3. **Why was it designed for single-machine?** — The initial architecture prioritized simplicity and the primary use case (one developer, one machine)
4. **Why hasn't it been extended?** — No network abstraction layer exists to bridge the message protocol across machines
5. **Root cause:** multiclaude was designed as a single-machine orchestrator with no transport abstraction separating "message semantics" from "message delivery"

### Contributing Factors

- NAT/firewall traversal makes direct connections between arbitrary machines nontrivial
- Claude Code sessions are ephemeral (no persistent identity for remote agents)
- tmux is inherently a local terminal multiplexer
- No standard "Claude Code federation" protocol exists in the ecosystem
- Security requirements (auth, encryption) add complexity to any network exposure

### System Dynamics

- **Amplifying loop:** As more users adopt multiclaude, demand for remote collaboration grows, but complexity of distributed state management also grows proportionally
- **Dampening loop:** More agents = more messages = higher coordination overhead; network latency amplifies this overhead
- **Leverage point:** The message protocol (`multiclaude message send/list/read/ack`) is the narrowest interface. If we can bridge THAT across networks, most collaboration patterns follow naturally
- **Second leverage point:** MCP (Model Context Protocol) already provides a tool extensibility mechanism in Claude Code. An MCP server that wraps multiclaude commands would let remote Claude Code instances use multiclaude as a "tool" without modifying multiclaude core

---

## ANALYSIS

### Force Field Analysis

**Driving Forces (Supporting Solution):**

1. **Strong user demand** — Remote work and distributed teams are the norm
2. **Git as existing distributed substrate** — multiclaude already uses Git extensively; it's a natural coordination layer
3. **GitHub API** — Provides cloud-based state (issues, PRs, comments) accessible from anywhere
4. **Mature tunneling infrastructure** — SSH, WireGuard, Tailscale, Cloudflare Tunnels are battle-tested
5. **MCP extensibility** — Claude Code already supports MCP servers, providing a natural extension point
6. **Message broker ecosystem** — Redis pub/sub, NATS, MQTT are lightweight and well-understood
7. **multiclaude's modular CLI** — Commands are composable and scriptable, easy to wrap

**Restraining Forces (Blocking Solution):**

1. **Security surface area** — Exposing a supervisor to the network creates attack surface
2. **State synchronization** — Filesystem state (state.json, worktrees) doesn't replicate easily
3. **NAT traversal complexity** — Both machines may be behind firewalls with no inbound access
4. **Session ephemerality** — Claude Code sessions don't persist identity across restarts
5. **Latency sensitivity** — Some operations (message ack, status polling) need near-real-time response
6. **Architectural inertia** — multiclaude core has no network abstractions; adding them is a significant change

### Constraint Identification

| Constraint | Type | Challengeable? |
|---|---|---|
| multiclaude state is filesystem-bound | Technical | Yes — could abstract state access behind an interface |
| Must work across NAT/firewalls | Environmental | No — this is a hard requirement |
| Security (auth + encryption) required | Non-functional | No — non-negotiable for remote access |
| Claude Code sessions are ephemeral | Platform | Partially — could add session persistence layer |
| Must not require VPN setup | Usability | Soft — some users may accept VPN, but shouldn't require it |

### Key Insights

1. **The message protocol is the narrowest bridge point** — abstracting `multiclaude message send/list/read/ack` across a network enables most collaboration patterns
2. **Git + GitHub already provide a distributed coordination substrate** — currently underutilized for inter-machine coordination
3. **MCP servers could act as the network bridge** without modifying multiclaude core at all
4. **The problem decomposes cleanly:** transport (SSH/tunnel/relay) is orthogonal to protocol (message format/semantics) — they can be designed and evolved independently
5. **GitHub Issues already function as a primitive message bus** — multiclaude's envoy agent already triages issues, making this a partially-working solution today

---

## SOLUTION GENERATION

### Methods Used

1. **Morphological Analysis** — Systematically combined transport, protocol, state sync, discovery, auth, and topology options
2. **TRIZ Contradiction Resolution** — Resolved the "security vs. accessibility" and "simplicity vs. distribution" contradictions
3. **Lateral Thinking** — Applied provocative operations to challenge assumptions about what "collaboration" requires

### Generated Solutions

#### Tier 1: Zero-Infrastructure Solutions (Work Today)

**1. SSH + tmux attach (Direct Session Sharing)**
- Remote user SSHs into supervisor machine: `ssh -t user@host tmux attach -t mc-ThreeDoors`
- Full visibility into all agent windows
- Can run `multiclaude` commands directly
- Can launch Claude Code inside the session
- **Pros:** Zero changes needed, full access, lowest latency, encrypted by default
- **Cons:** Requires SSH access (port forwarding or public IP), shares full machine access

**2. tmux over SSH with Claude Code Remote Execution**
- Remote user runs Claude Code on the supervisor's machine via SSH
- Claude Code process runs locally on the supervisor machine but user interacts remotely
- All filesystem access, tmux, and multiclaude commands work natively
- **Pros:** Everything "just works," no protocol bridging needed
- **Cons:** Requires SSH access, compute runs on supervisor machine, network interruptions kill session (mitigate with `mosh` or nested tmux)

**3. GitHub Issues as Message Bus**
- Remote user creates GitHub Issues with structured labels (e.g., `remote-task`, `priority:high`)
- Supervisor's envoy agent picks up issues via existing triage flow
- Results posted as issue comments; issue closed on completion
- **Pros:** Zero infrastructure, works through any firewall, auditable, already partially implemented
- **Cons:** High latency (GitHub API polling), rate limits, not truly real-time

#### Tier 2: Minimal-Infrastructure Solutions (Small New Components)

**4. MCP Bridge Server over SSH**
- Build an MCP server that wraps `multiclaude` CLI commands as MCP tools
- Tools: `send_task`, `check_status`, `list_workers`, `read_messages`, `ack_message`
- Remote Claude Code connects via SSH stdio transport: `ssh user@host /path/to/multiclaude-mcp-server`
- Claude Code already supports remote MCP servers via SSH — this is a native pattern
- **Pros:** Leverages existing MCP infrastructure, programmatic access, SSH provides auth+encryption
- **Cons:** Requires SSH access, requires building the MCP server (small Go binary)

**5. MCP Bridge Server over Cloudflare Tunnel**
- Same MCP server as #4, but exposed via Cloudflare Tunnel instead of SSH
- No inbound ports needed on either machine
- Authentication via Cloudflare Access (SSO, email verification, etc.)
- **Pros:** Works through any NAT/firewall, managed auth, no port forwarding
- **Cons:** Requires Cloudflare account, adds dependency on external service

**6. Tailscale Mesh + Message Bridge Daemon**
- Both machines join a Tailscale network (zero-config, works through NAT)
- A bridge daemon on each machine forwards multiclaude messages over the mesh
- Uses Tailscale's built-in WireGuard encryption and identity
- **Pros:** Zero-config NAT traversal, encrypted, persistent identity, works like LAN
- **Cons:** Requires Tailscale installation on both machines, requires bridge daemon development

**7. Git-Based Async Task Coordination**
- Tasks encoded as files in a dedicated branch (e.g., `remote-tasks/`)
- Remote user commits task files; supervisor watches for new commits via polling or webhook
- Results committed back to result branches; remote user pulls results
- Protocol: `tasks/<uuid>.yaml` with structured fields (action, args, status, result)
- **Pros:** Fully async, works offline, auditable, uses existing Git infrastructure
- **Cons:** High latency (push/pull cycle), complex state management, merge conflicts possible

**8. Lightweight WebSocket Relay**
- A tiny relay server (deployable to Cloudflare Workers, Fly.io, Deno Deploy, etc.)
- Both supervisor and remote user connect outbound to the relay (no NAT issues)
- JSON-RPC messages over WebSocket
- Relay is stateless — just forwards messages between authenticated parties
- **Pros:** Real-time bidirectional, works through any firewall, can be serverless
- **Cons:** Requires deploying and maintaining a relay, adds latency hop

#### Tier 3: Strategic Architecture Solutions (Significant Development)

**9. Network-Native multiclaude Message Protocol**
- Add a transport abstraction layer to multiclaude's message system
- Messages currently go through local filesystem; add network transport as an alternative
- Support multiple transports: local (file), TCP, WebSocket, NATS
- Remote agents register with the supervisor via the network transport
- **Pros:** First-class distributed support, clean architecture, enables federation
- **Cons:** Significant refactoring of multiclaude core, complex to implement correctly

**10. NATS-Based Agent Federation**
- Deploy NATS server (lightweight, ~15MB binary) as central message broker
- All multiclaude instances (local and remote) connect to NATS
- Topics map to multiclaude concepts: `supervisor.tasks`, `worker.{name}.status`, `messages.{agent}`
- Supports pub/sub, request/reply, and queue groups
- **Pros:** Industrial-grade messaging, extremely fast, supports complex topologies
- **Cons:** Requires NATS infrastructure, significant integration work

**11. Redis Pub/Sub + State Store**
- Cloud-hosted Redis serves as both message bus and shared state store
- multiclaude publishes state changes to Redis channels
- Remote agents subscribe and send commands via Redis
- State (state.json equivalent) stored in Redis hashes for shared access
- **Pros:** Mature ecosystem, supports both messaging and state, many hosting options
- **Cons:** Adds external dependency, requires state migration, operational overhead

**12. Hybrid: Git for Tasks + WebSocket for Status**
- Tasks submitted via Git commits (durable, auditable, works offline)
- Real-time status updates streamed via lightweight WebSocket connection
- Combines durability of Git with immediacy of WebSocket
- **Pros:** Best of both worlds — durable tasks, real-time feedback
- **Cons:** Two systems to maintain, complexity in keeping them consistent

### Creative Alternatives

**13. Email-Based Coordination**
- Send tasks as structured emails (YAML in body, labels in subject)
- Supervisor machine runs a local mail server or polls IMAP
- Results sent back as reply emails
- Works through literally any firewall. Hilariously low-tech but functional.

**14. Shared Cloud Document as State**
- Both agents read/write a shared Google Doc or Notion page via API
- Task queue as a table; status as cell values
- Surprisingly viable for low-frequency coordination

**15. DNS TXT Record Coordination**
- Encode small messages in DNS TXT records on a controlled domain
- Both sides can read DNS from anywhere; updates via DNS API
- Works through almost any network restriction
- Impractical for large payloads but fascinating for signaling

---

## SOLUTION EVALUATION

### Evaluation Criteria

| Criterion | Weight | Description |
|---|---|---|
| Effectiveness | 25% | Does it fully solve the bidirectional collaboration problem? |
| Feasibility | 20% | How easy is it to implement with existing tools/skills? |
| Security | 20% | Authentication, encryption, minimal attack surface |
| NAT Traversal | 15% | Works across firewalls without special network configuration |
| Latency | 10% | Near-real-time (< 30s) for interactive use |
| Complexity | 10% | Operational overhead, maintenance burden, failure modes |

### Solution Analysis

| Solution | Effect. | Feasib. | Security | NAT | Latency | Complex. | **Weighted** |
|---|---|---|---|---|---|---|---|
| 1. SSH + tmux | 9 | 10 | 9 | 3 | 10 | 10 | **8.05** |
| 2. tmux over SSH | 8 | 10 | 9 | 3 | 10 | 10 | **7.80** |
| 3. GitHub Issues | 7 | 10 | 9 | 10 | 4 | 9 | **8.00** |
| 4. MCP over SSH | 10 | 8 | 9 | 5 | 9 | 8 | **8.30** |
| 5. MCP + CF Tunnel | 10 | 7 | 9 | 10 | 8 | 6 | **8.45** |
| 6. Tailscale mesh | 9 | 7 | 9 | 10 | 9 | 6 | **8.30** |
| 7. Git-based async | 6 | 9 | 8 | 10 | 3 | 8 | **7.15** |
| 8. WebSocket relay | 9 | 6 | 7 | 10 | 9 | 5 | **7.70** |
| 9. Network-native | 10 | 4 | 8 | 8 | 9 | 3 | **7.15** |
| 10. NATS federation | 9 | 5 | 7 | 8 | 10 | 4 | **7.15** |

### Recommended Solution

**A tiered approach that builds progressively:**

#### Tier 1 — Available TODAY (Zero Development)

**Primary: SSH + tmux attach (#1)**
```bash
# Remote user connects to supervisor machine
ssh -t user@supervisor-host "tmux attach -t mc-ThreeDoors"
# Now has full access to all multiclaude windows and commands
```

**Fallback: GitHub Issues as async message bus (#3)**
- Create issues with label `remote-task` and structured body
- Envoy agent already picks up and triages issues
- Works through any firewall, zero setup

**When to use which:**
- SSH when you have network access to the supervisor machine
- GitHub Issues when you don't (or for async/non-urgent tasks)

#### Tier 2 — Near-Term Build (MCP Bridge, ~200 lines of Go)

**Primary: MCP Bridge Server (#4/#5)**

Build a small MCP server (`multiclaude-mcp-bridge`) that exposes these tools:

| MCP Tool | Maps To | Description |
|---|---|---|
| `send_task` | `multiclaude work "..."` | Dispatch a new worker task |
| `send_message` | `multiclaude message send` | Send a message to any agent |
| `list_messages` | `multiclaude message list` | Check pending messages |
| `read_message` | `multiclaude message read` | Read a specific message |
| `ack_message` | `multiclaude message ack` | Acknowledge a message |
| `worker_list` | `multiclaude worker list` | List active workers |
| `worker_status` | `multiclaude status` | Get system status |
| `repo_history` | `multiclaude repo history` | View task history |

**Transport options (choose based on network situation):**

| Transport | Config | NAT Traversal | Setup |
|---|---|---|---|
| SSH stdio | `ssh user@host multiclaude-mcp-bridge` | Requires SSH access | Minimal |
| Cloudflare Tunnel | `cloudflared tunnel --url localhost:8080` | Full NAT traversal | Moderate |
| Tailscale | Direct connection via Tailscale IP | Full NAT traversal | Moderate |

**Claude Code configuration (remote side):**
```json
{
  "mcpServers": {
    "multiclaude-remote": {
      "command": "ssh",
      "args": ["user@supervisor-host", "/usr/local/bin/multiclaude-mcp-bridge"]
    }
  }
}
```

#### Tier 3 — Strategic (Network-Native multiclaude)

**Add transport abstraction to multiclaude's message protocol:**

```
Current:  Agent → filesystem IPC → Agent
Future:   Agent → Transport Interface → {File | TCP | WebSocket | NATS} → Agent
```

This requires:
1. Define a `MessageTransport` interface in multiclaude
2. Current file-based IPC becomes the "local" transport
3. Add "remote" transports (TCP, WebSocket, NATS)
4. Remote agent registration and discovery protocol
5. Authentication and authorization layer

This is the "right" long-term solution but requires significant multiclaude core changes.

### Rationale

1. **Tiered approach avoids over-engineering** — each tier delivers value independently
2. **Tier 1 validates the use case** with zero investment before building anything
3. **MCP is the natural bridge** — Claude Code already speaks MCP, and multiclaude CLI is already composable. Wrapping one in the other is a clean architectural fit.
4. **SSH provides security for free** — authentication, encryption, and key management are solved problems
5. **Cloudflare Tunnel solves NAT** — for cases where SSH access isn't available, CF Tunnels provide zero-config NAT traversal with managed auth
6. **GitHub Issues provides universal fallback** — works through any network, any firewall, and is already partially implemented via envoy

---

## IMPLEMENTATION PLAN

### Implementation Approach

**Phased rollout: Tier 1 immediately, Tier 2 as a focused epic, Tier 3 as a strategic initiative.**

### Action Steps

#### Tier 1 (Immediate — documentation only)

1. Document SSH + tmux approach in multiclaude docs
2. Document GitHub Issues task submission format
3. Create a `remote-task` label in the repo
4. Add envoy rules to recognize and prioritize remote-task issues

#### Tier 2 (Near-term — MCP Bridge)

1. **Design MCP tool schema** — Define tool names, parameters, return types
2. **Build `multiclaude-mcp-bridge`** — Small Go binary, ~200 lines
   - Use `github.com/mark3labs/mcp-go` or similar MCP SDK
   - Each tool shells out to `multiclaude` CLI commands
   - stdio transport (reads JSON-RPC from stdin, writes to stdout)
3. **Add HTTP transport option** — For Cloudflare Tunnel / Tailscale scenarios
4. **Test with Claude Code** — Verify MCP tools work from remote Claude Code
5. **Document setup** — SSH config, MCP server config, Cloudflare Tunnel setup
6. **Security review** — Ensure no command injection via MCP tool parameters

#### Tier 3 (Strategic — deferred until demand proven)

1. Define `MessageTransport` interface in multiclaude core
2. Implement file transport (current behavior)
3. Implement TCP/WebSocket transport
4. Add agent registration protocol
5. Add authentication layer
6. Integration testing across transports

### Resource Requirements

| Tier | Effort | Dependencies |
|---|---|---|
| Tier 1 | Documentation only | SSH access to supervisor machine |
| Tier 2 | Small Go project (~200 LOC) | MCP SDK, Go toolchain |
| Tier 3 | Significant multiclaude refactoring | Architecture decision, multiclaude maintainer involvement |

### Responsible Parties

- **Tier 1:** Any developer with SSH access can set this up today
- **Tier 2:** multiclaude contributor or ThreeDoors maintainer
- **Tier 3:** multiclaude core maintainer(s)

---

## MONITORING AND VALIDATION

### Success Metrics

1. **Task round-trip time** — Time from remote task submission to result delivery (target: < 60s for Tier 1-2)
2. **Connection reliability** — Percentage of time remote connection is usable (target: > 95%)
3. **Security incidents** — Zero unauthorized access events
4. **User adoption** — Number of remote collaboration sessions per week

### Validation Plan

1. **Tier 1 validation:** Have a second person SSH into the supervisor machine and successfully dispatch a task via `multiclaude work`. Measure latency.
2. **Tier 2 validation:** Remote Claude Code instance uses MCP bridge to dispatch a task, monitors progress, and receives results. Full round-trip without SSH terminal access.
3. **Tier 3 validation:** Two multiclaude instances on different networks coordinate workers on a shared repository without any manual SSH or tunnel setup.

### Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| SSH access not available | Medium | High | Use Cloudflare Tunnel or GitHub Issues fallback |
| MCP bridge command injection | Low | Critical | Strict input validation, allowlist of commands, no shell interpolation |
| Network interruption during task | Medium | Medium | Tasks are durable (Git-based); status recoverable on reconnect |
| Supervisor overload from remote tasks | Low | Medium | Rate limiting in MCP bridge, queue depth limits |
| Stale state after reconnection | Medium | Low | Force-refresh status on reconnect |

### Adjustment Triggers

- **Pivot to Tier 3** if more than 3 users regularly need remote access
- **Add WebSocket relay** if Cloudflare Tunnel proves unreliable
- **Add NATS** if message volume exceeds what polling-based approaches handle
- **Abandon SSH approach** if security review identifies unacceptable risks

---

## LESSONS LEARNED

### Key Learnings

1. **The problem decomposes cleanly into transport and protocol** — solving them independently enables a tiered approach
2. **MCP is an underappreciated bridge mechanism** — Claude Code's MCP support makes it a natural extension point for remote operations
3. **GitHub Issues is already a working (if slow) message bus** — multiclaude's envoy agent already triages issues, making this partially functional today
4. **SSH is still the king of secure remote access** — tunneling MCP over SSH provides auth, encryption, and NAT traversal in one proven package
5. **Cloudflare Tunnel solves the hardest sub-problem (NAT traversal) for free** — zero-config, no inbound ports, managed security

### What Worked

- Morphological analysis revealed non-obvious combinations (MCP + SSH, Git + WebSocket hybrid)
- Is/Is Not analysis clarified that we're solving "bidirectional task coordination" not "full session sharing"
- Force field analysis identified MCP as the key leverage point
- Tiered approach lets us validate cheaply before building

### What to Avoid

- Don't build Tier 3 (network-native multiclaude) before validating Tier 1-2 — premature architecture astronautics
- Don't use NATS/Redis/etc. when SSH stdio works — unnecessary infrastructure
- Don't expose multiclaude commands via HTTP without authentication — command injection risk
- Don't try to replicate filesystem state across machines — use the message protocol as the abstraction layer instead

---

_Generated using BMAD Creative Intelligence Suite - Problem Solving Workflow_
