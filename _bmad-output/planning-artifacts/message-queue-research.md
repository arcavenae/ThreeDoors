# Message Queue for Dark Factory / REPL Agents — Research Report

> **Date:** 2026-03-29
> **Status:** Complete
> **Scope:** Pre-existing work, lightweight alternatives, REPL constraints, evolving Claude features, novel adaptations

---

## Executive Summary

Inter-agent communication in multi-agent AI orchestration is a rapidly evolving space with no single dominant solution. The key tension is between **simplicity** (file-based, zero-dependency) and **capability** (ordering, persistence, back-pressure, priority). For local REPL-based agent orchestration like multiclaude, the research identifies **SQLite-backed queues** as the most promising upgrade path — they add persistence, ordering, and atomic operations while keeping the system single-binary and zero-dependency.

**Key finding:** Every successful multi-agent orchestration system (Claude Code Agent Teams, Overstory, Claude Colony, multiclaude) converges on file-based IPC as the foundation. The question isn't whether to replace files but whether to add structure around them. SQLite is "files with superpowers" — it preserves the operational simplicity while adding the transactional guarantees that file-append lacks.

---

## 1. Lessons Learned / Successful Systems

### 1.1 Framework Communication Patterns

The major multi-agent AI frameworks use fundamentally different communication models:

| Framework | Communication Model | Strengths | Weaknesses |
|-----------|-------------------|-----------|------------|
| **LangGraph** | Graph-based state machine; edges carry typed state between nodes | Fine-grained flow control, checkpointing, human-in-the-loop | Heavyweight; requires graph definition upfront |
| **AutoGen/AG2** | GroupChat with selector; shared conversation context | Natural turn-taking, easy to prototype | Noisy at scale; context window explosion |
| **CrewAI** | Role-based delegation; agents hand off tasks | Intuitive mental model, autonomous delegation | Limited agent-to-agent messaging; sequential bias |
| **OpenAI Agents SDK** | Explicit handoffs; agents transfer control with context | Clean mental model, production-grade | No concurrent multi-agent; strictly sequential |
| **ChatDev** | Role-play chat chains; agents communicate via structured dialogue | Good for software engineering tasks | Rigid phase structure; hard to customize |

**Pattern that works:** Typed message envelopes with explicit routing. Every system that scales beyond 3-4 agents moves from "agents share a conversation" to "agents exchange discrete messages with metadata."

**Pattern that fails:** Shared context windows. AutoGen's GroupChat and similar "everyone sees everything" approaches break down at scale — token costs grow quadratically, and agents waste turns on irrelevant context.

### 1.2 Emerging Protocols

Four agent communication protocols have emerged (2025-2026):

- **MCP (Model Context Protocol)** — Anthropic. Tool-level access, not agent-to-agent. JSON-RPC 2.0 over stdio. Already integrated in Claude Code.
- **A2A (Agent2Agent Protocol)** — Google. HTTP-based, task-oriented, with Agent Cards for discovery. 50+ enterprise partners. Designed for cross-organization agent interop, not local orchestration.
- **ACP (Agent Communication Protocol)** — Cisco. Lightweight messaging between agents within the same organization.
- **ANP (Agent Network Protocol)** — Emerging. Decentralized discovery for internet-scale agent networks.

**Relevance to multiclaude:** MCP is the only protocol directly usable today (Claude Code supports it natively). A2A is over-engineered for local use but its Agent Card concept (capability discovery) is worth stealing. ACP's simplicity aligns with our needs but lacks Go libraries.

### 1.3 Production Multi-Agent Systems

**StrongDM's Dark Factory** (Level 5 autonomy, 3-person team): Uses a linear pipeline — no inter-agent messaging. Spec → coding agent → validation → merge. Isolation is the design philosophy; agents don't talk to each other.

**Overstory** (34-repo ecosystem analysis): SQLite-backed mail system with typed protocol messages (8 types: `worker_done`, `merge_ready`, `dispatch`, `escalation`, etc.). WAL mode, ~1-5ms per query. Supports both fire-and-forget and synchronous request/response with configurable timeouts. **Most architecturally similar to what multiclaude needs.**

**Claude Colony**: File-based IPC via `.colony/messages/` directory. @mention routing (`@frontend`, `@backend`, `@all`). Simple but no persistence guarantees, no ordering, no acknowledgment protocol.

**Claude Code Agent Teams** (native): File-based inbox per agent at `~/.claude/teams/{team-name}/inboxes/{agent-name}.json`. JSON-in-JSON encoding. Polling-based delivery. Messages injected as synthetic conversation turns. Known limitation: no stall detection, no resume, 7x token overhead.

### 1.4 Academic Foundation

Multi-agent systems (MAS) research identifies three communication paradigms:
1. **Direct messaging** (point-to-point) — simple but doesn't scale
2. **Broadcast/multicast** — good for announcements, wastes attention
3. **Blackboard systems** (shared memory space) — agents post to and read from a shared store

The blackboard pattern maps closely to what multiclaude already does (file-based shared state), but with the weakness that it lacks notification — agents must poll.

---

## 2. Lightweight Alternatives Evaluation

### 2.1 Evaluation Criteria

For our use case (local machine, 6-12 agents, REPL-based, Go codebase):

| Criterion | Weight | Description |
|-----------|--------|-------------|
| **Zero dependencies** | HIGH | No external services to start/manage/crash |
| **Go native** | HIGH | Must integrate cleanly with multiclaude's Go codebase |
| **Message persistence** | HIGH | Agent restarts shouldn't lose messages |
| **Ordering guarantees** | MEDIUM | FIFO within a queue; global ordering not required |
| **Priority support** | MEDIUM | URGENT security fix > routine heartbeat |
| **Latency** | LOW | Sub-second is fine; sub-millisecond is unnecessary |
| **Back-pressure** | MEDIUM | Don't flood a busy agent |
| **Observability** | MEDIUM | Able to inspect queue state for debugging |

### 2.2 Option Analysis

#### A. File-Based Queues (Current multiclaude approach)
**How it works:** `multiclaude message send` writes to a daemon-managed file queue. Agents poll via `multiclaude message list`.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★★ | Filesystem only |
| Go native | ★★★★★ | Already implemented |
| Persistence | ★★★☆☆ | Persists but no ACID guarantees; race conditions possible |
| Ordering | ★★☆☆☆ | Timestamp-based, no guaranteed FIFO under concurrent writes |
| Priority | ★☆☆☆☆ | No priority mechanism |
| Latency | ★★★☆☆ | Depends on poll interval (currently 5-30 min) |
| Back-pressure | ★☆☆☆☆ | No mechanism; messages accumulate |
| Observability | ★★★★☆ | Plain text files, easy to inspect |

**Verdict:** Works for current scale but fragile under load. No atomicity, no priority, no delivery guarantees.

#### B. Unix Domain Sockets
**How it works:** Local-only stream/datagram sockets. No network overhead.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★★ | Kernel-provided |
| Go native | ★★★★★ | `net.Dial("unix", path)` |
| Persistence | ★☆☆☆☆ | Volatile — lost on process restart |
| Ordering | ★★★★★ | Stream sockets guarantee ordering |
| Priority | ★☆☆☆☆ | Would need application-level implementation |
| Latency | ★★★★★ | Sub-microsecond |
| Back-pressure | ★★★☆☆ | Socket buffers provide natural back-pressure |
| Observability | ★★☆☆☆ | Requires custom tooling to inspect |

**Verdict:** Too volatile. Agent restarts are common (definition updates, crashes) and would lose all queued messages. Good for real-time streaming, wrong for reliable messaging.

#### C. Named Pipes (FIFOs)
**How it works:** Unidirectional byte streams via filesystem nodes.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★★ | Kernel-provided |
| Go native | ★★★★★ | `os.OpenFile` on FIFO node |
| Persistence | ★☆☆☆☆ | Volatile — kernel buffer only |
| Ordering | ★★★★★ | FIFO by definition |
| Priority | ★☆☆☆☆ | Single pipe = single priority |
| Latency | ★★★★★ | Sub-microsecond |
| Back-pressure | ★★★★☆ | Blocking writes when buffer full |
| Observability | ★☆☆☆☆ | Can't inspect without consuming |

**Verdict:** Same volatility problem as sockets. Also, blocking semantics are tricky with REPL agents that may not be reading.

#### D. SQLite-Based Queues ⭐ RECOMMENDED
**How it works:** SQLite database as a durable, ACID-compliant message store. Go library: `modernc.org/sqlite` (pure Go, no CGo).

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★★ | Single file, embedded in binary |
| Go native | ★★★★★ | Pure Go driver available (no CGo needed) |
| Persistence | ★★★★★ | ACID transactions, WAL mode, crash-safe |
| Ordering | ★★★★★ | Auto-increment IDs + timestamps = guaranteed FIFO |
| Priority | ★★★★★ | Priority column, `ORDER BY priority DESC, id ASC` |
| Latency | ★★★★☆ | ~1-5ms per operation (Overstory benchmarks) |
| Back-pressure | ★★★★☆ | Queue depth queries enable sender-side throttling |
| Observability | ★★★★★ | `sqlite3 queue.db "SELECT * FROM messages"` |

**Schema sketch:**
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    priority INTEGER DEFAULT 0,  -- 0=normal, 1=high, 2=urgent
    msg_type TEXT NOT NULL,       -- heartbeat, task, escalation, etc.
    body TEXT NOT NULL,
    correlation_id TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    delivered_at TEXT,
    acked_at TEXT,
    status TEXT NOT NULL DEFAULT 'pending'  -- pending, delivered, acked, dead
);

CREATE INDEX idx_recipient_status ON messages(recipient, status);
CREATE INDEX idx_priority ON messages(priority DESC, id ASC);
```

**Why this wins:**
- Overstory already proves this works at our exact scale (multi-agent Claude Code orchestration)
- SQLite WAL mode handles concurrent reads/writes from multiple agents
- multiclaude's dispatch queue already uses atomic write patterns — SQLite replaces temp-file-rename with real transactions
- `sqlite3` CLI provides instant observability without custom tooling
- Priority queues, dead letter handling, and message expiry are trivial SQL queries
- Pure Go driver means no CGo complications for cross-compilation

**Prior art:** litequeue (Python), liteq (R), Overstory's SQLite mail system, multiclaude's own `dispatch-queue.yaml` (already YAML-file-backed, easy migration path to SQLite).

#### E. Redis
**How it works:** In-memory data structure store with optional persistence.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★☆☆☆☆ | External server process required |
| Go native | ★★★★★ | Excellent Go clients (go-redis) |
| Persistence | ★★★★☆ | AOF/RDB persistence, but not crash-safe by default |
| Ordering | ★★★★★ | Streams provide ordered, consumer-group delivery |
| Priority | ★★★☆☆ | Sorted sets can model priority, not native |
| Latency | ★★★★★ | Sub-millisecond |
| Back-pressure | ★★★★☆ | Stream trimming, consumer group lag monitoring |
| Observability | ★★★★☆ | redis-cli, MONITOR command |

**Verdict:** Excellent technically but adds an external dependency. For a local development tool, requiring users to install and run Redis is a non-starter. Would only make sense if multiclaude already required Redis for other reasons.

#### F. NATS
**How it works:** Lightweight messaging system. Single binary, embeddable in Go.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★☆ | Single binary, or embeddable in Go process (~20MB RAM) |
| Go native | ★★★★★ | Written in Go, official Go client |
| Persistence | ★★★★★ | JetStream provides exactly-once delivery |
| Ordering | ★★★★★ | Stream ordering guarantees |
| Priority | ★★☆☆☆ | Not native; requires multiple subjects |
| Latency | ★★★★★ | Sub-millisecond |
| Back-pressure | ★★★★★ | Flow control built into JetStream |
| Observability | ★★★★☆ | nats CLI, built-in monitoring |

**Verdict:** The Lobsters 2025 consensus is "NATS is the most enjoyable messaging system." Embeddable in Go means no external process. JetStream adds persistence. **Strong second choice after SQLite.** The concern: NATS is designed for distributed systems — it's capability-overkill for local agent orchestration, and the embedded server adds ~20MB to the binary.

#### G. ZeroMQ
**How it works:** Brokerless messaging library with socket-like API.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★☆☆ | Library dependency, CGo binding (libzmq) |
| Go native | ★★☆☆☆ | CGo wrappers only (pebbe/zmq4), no pure Go |
| Persistence | ★☆☆☆☆ | Volatile — "silently discards messages when queue is full" |
| Ordering | ★★★★☆ | Per-socket ordering |
| Priority | ★☆☆☆☆ | Not supported |
| Latency | ★★★★★ | Extremely fast |
| Back-pressure | ★★☆☆☆ | High-water mark silently drops messages |
| Observability | ★☆☆☆☆ | Opaque; no built-in inspection |

**Verdict:** No. CGo dependency, silent message loss, no persistence. Designed for high-throughput networking, wrong for reliable agent orchestration.

#### H. tmux send-keys (Current injection mechanism)
**How it works:** Pastes text into an agent's tmux pane, simulating keyboard input.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★★ | tmux is already required |
| Go native | ★★★★☆ | Shell exec |
| Persistence | ★☆☆☆☆ | Fire-and-forget; no delivery guarantee |
| Ordering | ★★☆☆☆ | Race conditions under concurrent injection |
| Priority | ★☆☆☆☆ | No mechanism |
| Latency | ★★★★★ | Immediate (if agent is idle) |
| Back-pressure | ★☆☆☆☆ | Injections queue in paste buffer regardless of agent state |
| Observability | ★★☆☆☆ | Only visible in tmux scrollback |

**Verdict:** The root cause of multiclaude's UX pain (R-007). Injections interrupt active agents, corrupt mid-composition text, and are invisible to debugging. Should be replaced, not enhanced.

#### I. Claude Code Hooks
**How it works:** PreToolUse/PostToolUse shell commands that execute at specific lifecycle points.

| Criterion | Rating | Notes |
|-----------|--------|-------|
| Zero dependencies | ★★★★★ | Built into Claude Code |
| Go native | ★★★☆☆ | Shell commands; Go would wrap them |
| Persistence | ★★☆☆☆ | Hook scripts can write to durable storage |
| Ordering | ★★☆☆☆ | Hook execution is synchronous but trigger order isn't guaranteed |
| Priority | ★☆☆☆☆ | No mechanism |
| Latency | ★★★★☆ | Runs in-band with tool execution |
| Back-pressure | ★★★☆☆ | Blocking hooks naturally throttle |
| Observability | ★★★★☆ | Hook output is logged |

**Verdict:** Not a queue replacement, but an excellent **delivery mechanism**. A PostToolUse hook could check the SQLite queue and inject pending messages into the conversation context. This is the bridge between "message stored" and "message delivered to REPL agent."

### 2.3 Recommendation Matrix

| Approach | Now (Prototype) | Soon (multiclaude) | Later (Marvel) |
|----------|-----------------|---------------------|----------------|
| File-based | ✅ Current | ❌ Replace | ❌ Too fragile |
| SQLite queue | ✅ **Best fit** | ✅ **Best fit** | ✅ Scales to 50+ agents |
| NATS embedded | ❌ Overkill | ⚠️ Consider if needs grow | ✅ Natural fit for distributed |
| Redis | ❌ Extra dependency | ❌ Extra dependency | ⚠️ If already in stack |
| Hooks | ✅ **Delivery layer** | ✅ **Delivery layer** | ✅ **Delivery layer** |

---

## 3. REPL Agent Constraints

### 3.1 The Fundamental Problem

REPL agents (Claude Code instances) have a unique constraint that traditional message consumers don't: **they can only receive messages during specific lifecycle phases.** A Claude Code instance cycles through:

```
IDLE (waiting for input) → THINKING (processing) → TOOL_CALL (executing) → THINKING → ... → IDLE
```

Messages can only be reliably injected during **IDLE** or via **hooks during TOOL_CALL**. Injection during THINKING corrupts the context.

### 3.2 Agent States and Message Handling

| State | Can Receive? | Mechanism | Risk |
|-------|-------------|-----------|------|
| **IDLE** (waiting for input) | ✅ Yes | stdin injection, synthetic user turn | Low — this is the natural intake point |
| **THINKING** (LLM processing) | ❌ No | N/A | HIGH — injection corrupts context |
| **TOOL_CALL** (executing a tool) | ⚠️ Via hooks | Pre/PostToolUse hooks | Medium — hook must not block long-running tools |
| **COMPILING/CI** (waiting on external) | ⚠️ Deferred | Queue for later delivery | Low — agent will return to IDLE |
| **CRASHED/STOPPED** | ✅ Queue only | Persistent queue holds messages | None — messages wait for restart |

### 3.3 Required Capabilities

**Message Buffering:** SQLite queue naturally buffers. Messages accumulate in `pending` status. Delivery occurs when the agent's delivery mechanism fires (hook trigger or poll).

**Priority Levels:**
```
PRIORITY 0: Routine (heartbeats, status checks)
PRIORITY 1: Normal (task assignments, PR notifications)
PRIORITY 2: High (merge conflicts, CI failures)
PRIORITY 3: Urgent (security fixes, incident response)
```

Priority determines delivery order AND whether to interrupt an active agent. Priority 0-1 waits for IDLE. Priority 2-3 can use hook injection.

**Deferred Delivery:** The SQLite queue holds messages indefinitely. A delivery daemon (or hook) checks the queue periodically. When the agent reaches IDLE, all pending messages are delivered in priority order.

**Acknowledgment Protocol:**
```
pending → delivered → acked
                   → dead (after N retries or TTL expiry)
```

The current multiclaude protocol already has `message ack`. SQLite adds reliable state tracking — a message delivered but not acked within T seconds can be re-delivered or escalated.

**Back-Pressure:** Query `SELECT COUNT(*) FROM messages WHERE recipient = ? AND status = 'pending'`. If queue depth exceeds threshold, sender receives a back-pressure signal. For heartbeats (the primary flood source), the daemon can skip sending if a heartbeat is already pending.

### 3.4 Claude Code's Input Model

Based on reverse-engineering (dev.to article) and official docs:

- **stdin:** Claude Code reads from stdin in interactive mode. `tmux send-keys` simulates keyboard input to stdin. This is what multiclaude currently uses.
- **Synthetic conversation turns:** Agent Teams writes to inbox JSON files. Claude Code polls and injects new messages as if the user typed them. This is cleaner than stdin injection.
- **Hooks:** Pre/PostToolUse hooks execute shell commands. Output can influence the next turn. This is the most controlled injection point.
- **`--append-system-prompt`:** Can inject context at session start. Not useful for runtime messaging, but useful for initial configuration.
- **MCP tools:** Custom MCP servers can expose tools that the agent calls actively. This reverses the flow — instead of pushing messages to the agent, the agent pulls them by calling a "check messages" tool.

**Recommended delivery stack:**
1. **Primary:** MCP tool (`check_messages`) — agent actively pulls from SQLite queue
2. **Secondary:** PostToolUse hook — checks queue after each tool execution, injects pending messages
3. **Fallback:** tmux send-keys for IDLE agents (when hooks haven't fired recently)

---

## 4. Evolving Claude Features

### 4.1 Current Integration Points (March 2026)

| Feature | Status | Relevance to Message Queue |
|---------|--------|---------------------------|
| **MCP servers** | Stable | Custom `check_messages` tool — agent-initiated polling |
| **Hooks** | Stable | PostToolUse delivery mechanism |
| **Agent Teams** | Experimental | File-based inbox pattern — validates our approach |
| **SendMessage tool** | Experimental | Native agent-to-agent; may eventually replace custom messaging |
| **Subagents** | Stable | In-process only; no cross-session messaging |
| **MCP Elicitation** | New (v2.1.76) | MCP servers can request user input mid-task |
| **`/btw`** | New | Side-channel for low-priority context injection |

### 4.2 MCP as Message Delivery Channel

The strongest near-term opportunity. An MCP server can:

1. **Expose a `check_messages` tool** — agent calls it to poll for pending messages
2. **Expose a `send_message` tool** — agent sends messages through the MCP server to the SQLite queue
3. **Expose a `get_queue_status` tool** — agent checks queue depth, pending counts
4. **Use MCP Elicitation** — server can push structured prompts to the agent (new in v2.1.76)

This eliminates tmux injection entirely. The agent **actively** checks for messages as part of its tool-calling loop, rather than having messages **injected** into its context unpredictably.

**Architecture:**
```
┌─────────────┐     MCP (stdio/JSON-RPC)     ┌──────────────┐
│ Claude Code  │◄──────────────────────────────►│  MCP Server  │
│   Agent      │   check_messages/send_message  │  (Go binary) │
└─────────────┘                                └──────┬───────┘
                                                      │
                                               ┌──────▼───────┐
                                               │   SQLite DB   │
                                               │  (messages)   │
                                               └──────────────┘
```

### 4.3 Anticipating Future REPL Capabilities

**Native agent-to-agent:** Claude Code Agent Teams already has SendMessage. If this moves from experimental to stable, multiclaude could use it directly for agent communication, with SQLite as the persistence/priority layer behind it.

**Event-driven hooks:** If Claude Code adds "on idle" or "on message received" hooks (beyond Pre/PostToolUse), message delivery becomes trivial — fire on idle, deliver pending messages.

**Structured output channels:** If Claude Code gains a structured side-channel (beyond stdout), message delivery could use it instead of synthetic conversation turns.

**Recommendation:** Design the SQLite queue as the **source of truth** regardless of delivery mechanism. Delivery adapters (MCP, hooks, tmux, native SendMessage) are pluggable on top. This future-proofs against API changes.

---

## 5. Unique Adaptations & Novel Work

### 5.1 Existing Approaches

| System | Approach | Key Innovation |
|--------|----------|----------------|
| **multiclaude** | CLI-based file queue + tmux injection | Worktree isolation per agent; daemon refresh loop |
| **Overstory** | SQLite mail + typed messages | Fire-and-forget + request/response; debounce flag |
| **Claude Colony** | File-based IPC + @mention routing | Natural language addressing |
| **Claude Code Agent Teams** | JSON inbox files + polling | Synthetic conversation turn injection |
| **claude-code-agent-farm** | Parallel tmux sessions | Lock-based coordination |
| **oh-my-claudecode** | Teams-first orchestration | Team topology as first-class concept |

### 5.2 What multiclaude Does Differently

multiclaude's architecture is unique in several ways:
1. **Persistent agents** with defined roles (merge-queue, pr-shepherd, etc.) — not ad-hoc teams
2. **Daemon-managed worktrees** with auto-refresh — agents never do git sync
3. **Hook-enforced safety** (git-safety.sh) — policy enforcement via tooling, not prompts
4. **Correlation IDs** to prevent cascade loops — learned from production incidents

These are strengths to preserve in any messaging upgrade.

### 5.3 Novel Opportunities

**1. Typed Message Protocol for AI Agents**

Unlike human messaging, AI agent messages have predictable structure. A typed protocol could include:

```go
type Message struct {
    ID            string    `json:"id"`
    Type          MsgType   `json:"type"`           // heartbeat, task, escalation, notification, query
    Priority      int       `json:"priority"`        // 0-3
    Sender        string    `json:"sender"`
    Recipient     string    `json:"recipient"`       // agent name or @all
    CorrelationID string    `json:"correlation_id"`  // PR-123, Story-45.2
    Body          string    `json:"body"`
    Metadata      map[string]string `json:"metadata"` // arbitrary k/v
    CreatedAt     time.Time `json:"created_at"`
    ExpiresAt     *time.Time `json:"expires_at"`      // TTL for heartbeats
    DeliveredAt   *time.Time `json:"delivered_at"`
    AckedAt       *time.Time `json:"acked_at"`
}
```

**2. Activity-Aware Delivery**

Instead of blind polling, the delivery daemon could:
- Monitor Claude Code's JSONL transcript for recent tool calls (agent is active → defer non-urgent messages)
- Check tmux pane output for prompt character (agent is idle → deliver now)
- Use PostToolUse hooks to piggyback delivery on natural tool-call boundaries

**3. Message Coalescing**

Multiple heartbeat-type messages can be coalesced into a single delivery:
```
"You have 3 pending messages: 2 heartbeats (merged), 1 PR notification (#847 merged)"
```

This reduces context window consumption — the #1 cost driver in multi-agent systems.

### 5.4 Ecosystem Check

- **dollspace-gay/chainlink:** Investigated in R-010. Philosophy: "enforce via tooling, not prompts." No specific message queue innovation, but the typed comment system and session handoff protocol are relevant.
- **OpenClaudia:** Not found as a distinct project. May refer to open-source Claude Code wrappers.
- **multiclaude community MCP servers:** None found specifically for messaging. The MCP ecosystem is focused on tool access (GitHub, databases, APIs) rather than agent-to-agent communication.

---

## 6. Cross-Cutting: Marvel Integration

### 6.1 What Can Be Prototyped NOW (ThreeDoors/multiclaude-enhancements)

1. **SQLite message queue** — Replace file-based `multiclaude message send/list/ack` with SQLite backend
2. **Priority support** — Add priority column and delivery ordering
3. **MCP message server** — Expose `check_messages`/`send_message` as MCP tools
4. **Heartbeat deduplication** — Skip sending heartbeat if one is already pending
5. **Message coalescing** — Merge multiple pending messages into a single delivery
6. **Activity-aware delivery** — Check JSONL transcript for agent activity before injecting

### 6.2 What Needs Marvel's Architecture

1. **Multi-machine messaging** — SQLite is single-machine. Marvel would need NATS or similar for cross-machine agent communication.
2. **Agent discovery** — A2A-style Agent Cards for capability-based routing
3. **Distributed queue** — NATS JetStream or equivalent for reliable cross-machine delivery
4. **Centralized observability** — OTEL integration for message latency, queue depth, delivery success metrics
5. **Authentication/authorization** — Message-level access control (which agents can message which)

### 6.3 OTEL Integration Points

```
queue.message.sent          — counter, labels: sender, recipient, type, priority
queue.message.delivered     — counter + histogram (delivery latency)
queue.message.acked         — counter + histogram (processing latency)
queue.message.dead          — counter (failed deliveries)
queue.depth                 — gauge per recipient
queue.delivery.mechanism    — counter, labels: mcp, hook, tmux, poll
```

---

## 7. Recommendations

### Immediate (This Sprint)
1. **Prototype SQLite message queue** in multiclaude — replace YAML file queue with SQLite
2. **Add priority field** to message protocol
3. **Implement heartbeat deduplication** — check `pending` count before sending

### Short-Term (Next 2-4 Weeks)
4. **Build MCP message server** — `check_messages`, `send_message`, `queue_status` tools
5. **Add PostToolUse delivery hook** — check queue after each tool execution
6. **Implement message coalescing** — reduce context window consumption

### Medium-Term (1-2 Months)
7. **Add message TTL and dead-letter handling** — heartbeats expire after 2x interval
8. **Implement activity-aware delivery** — monitor JSONL transcripts for agent state
9. **Add OTEL metrics** — queue depth, delivery latency, message throughput

### Long-Term (Marvel)
10. **Evaluate NATS embedded** for multi-machine messaging
11. **Implement A2A-style Agent Cards** for capability discovery
12. **Build centralized message dashboard** — real-time queue visibility across all agents

---

## 8. Key Decisions Needed

| ID | Decision | Options | Recommendation |
|----|----------|---------|----------------|
| MQ-D-001 | SQLite driver: CGo (mattn) vs pure Go (modernc) | CGo is faster; pure Go is more portable | Pure Go (`modernc.org/sqlite`) — eliminates cross-compilation issues |
| MQ-D-002 | Queue location: per-agent vs shared | Per-agent isolates failures; shared enables cross-queries | Shared single DB with per-recipient indexes |
| MQ-D-003 | Delivery mechanism priority | MCP > hooks > tmux > poll | MCP primary, hooks secondary, tmux fallback |
| MQ-D-004 | Message TTL for heartbeats | Fixed (e.g., 10min) vs configurable per-type | Configurable per message type, default 2x send interval |
| MQ-D-005 | Back-pressure signal | Reject at sender vs flag on recipient | Flag on recipient (queue depth metric); sender checks before sending |

---

## 9. Open Questions

| ID | Question | Context |
|----|----------|---------|
| MQ-Q-001 | Can Claude Code's MCP server read from SQLite in WAL mode without locking issues when the daemon also writes? | SQLite WAL mode supports concurrent readers + single writer. Need to confirm MCP server process and daemon don't conflict. |
| MQ-Q-002 | What's the maximum message size before it degrades Claude Code's context window? | Current 2000-char limit may be too small for structured messages with metadata. |
| MQ-Q-003 | Should the MCP message server be a standalone binary or embedded in multiclaude daemon? | Standalone is simpler for Claude Code config; embedded avoids another process. |
| MQ-Q-004 | How do we handle message delivery during agent restart? | SQLite persists messages, but the delivery mechanism (hook/MCP) needs to reconnect. |
| MQ-Q-005 | Should message content be opaque (string) or structured (JSON schema per type)? | Typed messages enable better coalescing and routing but add schema maintenance overhead. |

---

## Sources

### Multi-Agent Frameworks & Orchestration
- [Best Multi-Agent Frameworks in 2026](https://gurusup.com/blog/best-multi-agent-frameworks-2026)
- [The Multi-Agent Pattern That Actually Works in Production](https://www.chanl.ai/blog/multi-agent-orchestration-patterns-production-2026)
- [CrewAI vs LangGraph vs AutoGen](https://www.datacamp.com/tutorial/crewai-vs-langgraph-vs-autogen)
- [A Detailed Comparison of Top 6 AI Agent Frameworks in 2026](https://www.turing.com/resources/ai-agent-frameworks)

### Agent Communication Protocols
- [Survey of Agent Interoperability Protocols (MCP, ACP, A2A, ANP)](https://arxiv.org/html/2505.02279v1)
- [Google A2A Protocol Announcement](https://developers.googleblog.com/en/a2a-a-new-era-of-agent-interoperability/)
- [Multi-Agent Communication Protocols: A Technical Deep Dive](https://geekyants.com/blog/multi-agent-communication-protocols-a-technical-deep-dive)

### Message Queue Technology
- [What's your go-to message queue in 2025? (Lobsters)](https://lobste.rs/s/uwp2hd/what_s_your_go_message_queue_2025)
- [NATS Messaging](https://nats.io/)
- [litequeue — SQLite-based queue (Python)](https://github.com/litements/litequeue)
- [ZeroMQ vs NATS comparison](https://stackshare.io/stackups/nats-vs-zeromq)

### Claude Code Architecture
- [Reverse-Engineering Claude Code Agent Teams](https://dev.to/nwyin/reverse-engineering-claude-code-agent-teams-architecture-and-protocol-o49)
- [Claude Code Extensions Explained](https://muneebsa.medium.com/claude-code-extensions-explained-skills-mcp-hooks-subagents-agent-teams-plugins-9294907e84ff)
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [Claude Code MCP Server Setup Guide](https://www.ksred.com/claude-code-as-an-mcp-server-an-interesting-capability-worth-understanding/)

### Multi-Agent Orchestration Tools
- [Overstory — Multi-agent orchestration with pluggable runtimes](https://github.com/jayminwest/overstory)
- [Claude Colony — tmux-based multi-agent with file-based IPC](https://github.com/MakingJamie/claude-colony)
- [Claude Squad — Multi-agent terminal manager](https://github.com/smtg-ai/claude-squad)
- [claude-code-agent-farm](https://github.com/Dicklesworthstone/claude_code_agent_farm)

### Dark Factory Pattern
- [The Dark Factory Pattern (HackerNoon)](https://hackernoon.com/the-dark-factory-pattern-moving-from-ai-assisted-to-fully-autonomous-coding)
- [What Is a Dark Factory AI Agent? (MindStudio)](https://www.mindstudio.ai/blog/what-is-a-dark-factory-ai-agent)
- [StrongDM's Software Factory](https://simonwillison.net/2026/Feb/7/software-factory/)

### Academic & Theoretical
- [Communication Methods in Multi-Agent Reinforcement Learning](https://arxiv.org/html/2601.12886v1)
- [Survey of LLM Agent Communication with MCP](https://arxiv.org/pdf/2506.05364)
- [Coordination Mechanisms in Multi-Agent Systems](https://apxml.com/courses/agentic-llm-memory-architectures/chapter-5-multi-agent-systems/coordination-mechanisms-mas)
