# Workspace → Director Rename & Remote Access Vision

**Date:** 2026-03-29
**Type:** Research Spike
**Status:** Complete

---

## 1. Rename Impact Analysis

### 1.1 multiclaude Internals — DEEPLY HARDCODED

"Workspace" is a **first-class agent type** in the multiclaude Go binary. Renaming requires modifying the multiclaude source code.

**Core references (18 Go files):**

| File | What's hardcoded | Severity |
|------|-----------------|----------|
| `internal/state/state.go:20` | `AgentTypeWorkspace AgentType = "workspace"` — **THE** type constant | Critical |
| `internal/state/state.go:31` | `IsPersistent()` switch includes `AgentTypeWorkspace` | Critical |
| `internal/daemon/daemon.go` | 15+ refs: `routeMessages()` and `wakeAgents()` skip workspace, worktree creation, tmux window named "workspace" | Critical |
| `internal/cli/cli.go` | `multiclaude workspace` command (~200 lines), `addWorkspace`, `removeWorkspace`, `listWorkspaces`, `connectWorkspace`, `validateWorkspaceName` | Critical |
| `pkg/claude/prompt/builder.go:87` | `TypeWorkspace AgentType = "workspace"` — prompt builder type | Critical |
| `internal/diagnostics/collector.go` | Workspace counter in diagnostics | Low |
| `internal/bugreport/collector.go` | Bug report workspace detection | Low |
| `cmd/generate-docs/main.go` | CLI docs generation references "workspace" | Low |

**Branch naming convention:** `workspace/<name>` (e.g., `workspace/default`) — baked into worktree creation and cleanup logic.

**Tmux window name:** Hardcoded as `"workspace"` in `daemon.go:1898`.

**State persistence:** `state.json` stores `"type": "workspace"` for workspace agents.

**Verdict: Renaming in multiclaude requires a source-level fork/patch.** This is not a configuration change.

### 1.2 ThreeDoors Project References

| Location | Count | Context |
|----------|-------|---------|
| `docs/stories/73.1.story.md` | 12 | "workspace-as-primary" UX pattern story |
| `docs/prd/epics-and-stories.md` | 5 | Epic 73 workspace references |
| `docs/prd/epic-list.md` | 3 | Epic 73 summary |
| `docs/decisions/BOARD.md` | ~8 | R-007, P-010, consolidation roadmap refs |
| Memory files (`MEMORY.md`, `multiclaude-startup.md`) | 4 | Startup notes, operator guidance |
| Planning artifacts | 40+ | Multiple research docs reference workspace |
| `docs/stories/66.*.story.md` | 4 | ClickUp `workspace_id` (different concept — ClickUp API) |
| Go source (`internal/cli/connect.go`, etc.) | 7 | ClickUp workspace ID (NOT related — this is ClickUp's term) |

**Important distinction:** The Go codebase uses "workspace" in two contexts:
1. **multiclaude workspace** (the tmux window/agent type) — ~target of rename~
2. **ClickUp workspace** (external API concept) — NOT a rename target

### 1.3 Agent Definitions

No references to "workspace" in `agents/*.md` files. Agent definitions don't mention the workspace concept directly.

---

## 2. Name Evaluation

### Candidates

| Name | Pros | Cons | Verdict |
|------|------|------|---------|
| **director** | Clear authority role; aligns with human oversight metaphor; "the director directs the factory" | Slightly long (8 chars); could conflict with future agent roles | **Recommended** |
| **ds** (director's seat) | Short (2 chars); unique; memorable; quick to type | Cryptic to newcomers; initialism requires explanation | **Good alias** |
| **operator** | Descriptive; common in SRE/DevOps | Generic; conflicts with Kubernetes "operator" concept | Rejected |
| **console** | Familiar to developers | Overloaded (browser console, game console, OS console) | Rejected |
| **bridge** | Evocative (Star Trek); implies command & control | "Too cute"; doesn't translate well to CLI context | Rejected |
| **helm** | Short; nautical metaphor; implies steering | Conflicts with Kubernetes Helm package manager | Rejected |
| **cockpit** | Clear control metaphor | Long; informal | Rejected |
| **seat** | Short; physical metaphor | Too generic; "take a seat" is passive | Rejected |

### Recommendation: `director` (full name) / `ds` (alias)

- `multiclaude director connect` — clear intent
- `multiclaude ds` — power-user shorthand
- Tmux window name: `director` (or `ds`)
- Agent type: `AgentTypeDirector = "director"`
- Branch prefix: `director/<name>` (replaces `workspace/<name>`)

---

## 3. Strategic Vision: Remote Director

### 3.1 Concept

The **Director** is the human's remote control plane into the dark factory. Today it's a local tmux window. Tomorrow it connects to an **upstream director** — becoming a remote access channel for the multiclaude message protocol.

```
┌──────────────────────────────────────────────────────┐
│  REMOTE (laptop, phone, coffee shop)                 │
│                                                       │
│  ┌─────────────────────┐                              │
│  │  Director Client     │                              │
│  │  (Remote REPL)       │                              │
│  │                     │                              │
│  │  > status            │  ←→  WebSocket / SSH tunnel  │
│  │  > msg supervisor    │      to upstream director     │
│  │  > show agents       │                              │
│  └─────────────────────┘                              │
└──────────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────┐
│  UPSTREAM (dark factory machine)                      │
│                                                       │
│  ┌─────────────────────┐  ┌──────────────────┐       │
│  │  Director Daemon     │  │  multiclaude     │       │
│  │  (message bridge)    │  │  daemon          │       │
│  │                     │  │                  │       │
│  │  Reads agent msgs   ├──┤  File-based msgs │       │
│  │  Writes to agents   │  │  tmux sessions   │       │
│  │  Streams status     │  │  worktree mgmt   │       │
│  └─────────────────────┘  └──────────────────┘       │
│                                                       │
│  ┌─────────────────────────────────────────────┐     │
│  │  mc-ThreeDoors tmux session                  │     │
│  │  supervisor | director | merge-queue | ...   │     │
│  └─────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────┘
```

### 3.2 Remote Capabilities

| Capability | Description | Implementation |
|-----------|-------------|----------------|
| **Message feed** | Stream of all agent↔supervisor messages, filtered by relevance | Director daemon tails message JSON files, streams to client |
| **Send messages** | Send messages to any agent (primarily supervisor) | Write JSON to `~/.multiclaude/messages/<repo>/<agent>/` |
| **Agent status** | Live view of agent states, worktrees, PRs | Poll `multiclaude status` output or daemon API |
| **PR overview** | Open PRs, CI status, merge-queue state | `gh pr list` via daemon proxy |
| **Task dispatch** | Spawn workers, trigger stories | `multiclaude work "..."` via daemon proxy |
| **Log streaming** | Tail agent output logs | `multiclaude logs <agent> -f` via daemon proxy |

### 3.3 Remote REPL vs Local REPL

**Key insight:** The remote director does NOT run a local Claude Code agent. Instead:

- **Local director** (current): A Claude Code instance in a tmux window. Full AI capabilities, filesystem access, tool use.
- **Remote director** (future): A lightweight REPL that sends commands to the upstream director daemon. No local Claude Code. Thin client.

The remote REPL is a **command router**, not an AI agent. Commands like `msg supervisor "check PR #850"` translate to `multiclaude message send supervisor "check PR #850"` on the remote machine.

### 3.4 Relationship to Slack Bot (R-009)

| Dimension | Director (Remote) | Slack Bot (R-009) |
|-----------|-------------------|-------------------|
| **Primary user** | The operator (human owner) | Team members, stakeholders |
| **Access model** | Direct, privileged, full control | RBAC-gated, channel-scoped |
| **Interface** | Terminal REPL | Slack messages, slash commands |
| **Latency** | Real-time (WebSocket/SSH) | Near-real-time (Socket Mode) |
| **Use case** | Operating the factory | Monitoring, delegated tasks |
| **Authentication** | SSH keys / mTLS | Slack workspace + RBAC YAML |

**Relationship: Complementary, not overlapping.**

- Director = operator's direct control line (like SSH into a server)
- Slack bot = team interface (like a monitoring dashboard with actions)
- Both connect to the same multiclaude message protocol
- Director has higher privilege than Slack bot (no RBAC restrictions)
- Slack bot has richer notifications (threads, reactions, channels)

The director daemon could be the **shared backend** that both the remote REPL and the Slack bot connect to:

```
Remote REPL  ──┐
                ├──→  Director Daemon  ──→  multiclaude message protocol
Slack Bot    ──┘
```

### 3.5 Marvel Integration

In the Marvel platform vision:
- Director becomes a **first-class remote access channel** in Marvel's architecture
- Marvel's message queue replaces file-based message passing
- Director daemon evolves into Marvel's **operator gateway**
- Content packs (multiclaude, BMAD, etc.) are accessible through the director
- Multiple directors can connect to one Marvel instance (multi-operator)
- Job management, traffic routing, and secrets management are exposed through director REPL

---

## 4. What Changes Now vs Later

### NOW (ThreeDoors scope — zero multiclaude changes)

| Change | Effort | Impact |
|--------|--------|--------|
| Update memory files to use "director" terminology | 30 min | Aligns mental model |
| Update story 73.1 docs to reference "director seat" alongside "workspace" | 30 min | Forward-looking language |
| Add "director" as an alias concept in CLAUDE.md | 10 min | Agents understand the term |
| Update consolidation roadmap to reference "director" vision | 30 min | Strategic alignment |

**Important:** Don't rename multiclaude references yet — the binary still uses "workspace". Document the rename intent but keep operational references accurate to the current binary.

### LATER (multiclaude fork — source changes required)

| Change | Effort | Impact |
|--------|--------|--------|
| Rename `AgentTypeWorkspace` → `AgentTypeDirector` in state.go | 1 hour | Core type change |
| Rename CLI command `workspace` → `director` (keep `workspace` as deprecated alias) | 2 hours | User-facing change |
| Update branch prefix `workspace/` → `director/` | 1 hour | Git convention change |
| Migration: rename existing `workspace/default` branches | 30 min | One-time migration |
| Update daemon workspace creation → director creation | 1 hour | Daemon logic |
| Update all 18 Go files referencing "workspace" | 3 hours | Comprehensive rename |
| Add `ds` as CLI alias for `director` | 30 min | Power-user shorthand |
| **Total estimated effort** | **~1 day** | Full rename in multiclaude |

**Migration strategy:** Add `AgentTypeDirector` as a new constant, keep `AgentTypeWorkspace` as a deprecated alias that maps to Director. State files with `"type": "workspace"` auto-migrate to `"type": "director"` on load. CLI accepts both `multiclaude workspace` and `multiclaude director` during transition period.

### MARVEL (native implementation)

| Change | Description |
|--------|-------------|
| Director as native concept | No more tmux window — proper remote-accessible gateway |
| Director daemon | Standalone service bridging remote clients ↔ Marvel message queue |
| Remote REPL | Lightweight CLI client (Go binary) with WebSocket transport |
| Multi-operator | Multiple directors connected to one Marvel instance |
| Director protocol | JSON-RPC or gRPC, replacing file-based message passing |
| Slack integration | Slack bot connects through director daemon API |

---

## 5. Open Questions

| ID | Question | Context |
|----|----------|---------|
| Q-DIR-1 | Should the multiclaude rename happen before or after the platform extraction to Marvel? | If before: rename in multiclaude fork. If after: skip rename, build "director" natively in Marvel. Trade-off: rename now improves mental model during prototype phase; skip saves wasted effort on a soon-to-be-replaced system. |
| Q-DIR-2 | Should the remote director use WebSocket, SSH tunnel, or gRPC for transport? | WebSocket: web-friendly, works from browser. SSH: zero infrastructure, works today. gRPC: typed, fast, needs proto defs. |
| Q-DIR-3 | Should the director daemon be a separate binary or integrated into multiclaude? | Separate: cleaner separation, independent deployment. Integrated: simpler, one binary to manage. |
| Q-DIR-4 | Should "ds" be the CLI alias from day one, or added later as a power-user shortcut? | Early: builds muscle memory. Later: reduces confusion during transition. |

---

## 6. Decisions

| ID | Decision | Rationale |
|----|----------|-----------|
| D-DIR-1 | "Director" is the chosen name for what multiclaude calls "workspace" | Clear authority role, aligns with human oversight of dark factory, works as both concept and CLI command. "ds" (director's seat) as alias. |
| D-DIR-2 | Remote director vision is complementary to Slack bot, not a replacement | Director = privileged operator control. Slack = team interface. Both use the same message protocol. Director daemon can be shared backend. |
| D-DIR-3 | NOW: Adopt "director" terminology in docs/memory. LATER: Rename in multiclaude fork. MARVEL: Build natively. | Phased approach avoids wasted effort while aligning mental model immediately. |
