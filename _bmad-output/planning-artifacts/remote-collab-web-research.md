# Remote Claude Code Collaboration — Web Research Report

> **Date:** 2026-03-10
> **Scope:** GitHub repos, Anthropic docs, MCP bridges, multi-agent orchestrators, community discussions, and tools for remote/distributed Claude Code collaboration.

---

## Table of Contents

1. [Anthropic Official Features](#1-anthropic-official-features)
2. [Multi-Agent Orchestrators](#2-multi-agent-orchestrators)
3. [MCP Bridges & Inter-Instance Communication](#3-mcp-bridges--inter-instance-communication)
4. [Remote Execution Approaches](#4-remote-execution-approaches)
5. [Hooks & Extensions for Multi-Agent Workflows](#5-hooks--extensions-for-multi-agent-workflows)
6. [Community Discussion & Resources](#6-community-discussion--resources)
7. [Comparison Matrix](#7-comparison-matrix)
8. [Key Findings & Recommendations](#8-key-findings--recommendations)

---

## 1. Anthropic Official Features

### 1.1 Agent Teams (TeammateTool)

Anthropic's native multi-agent system, currently experimental.

- **Status:** Experimental, behind feature flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
- **Architecture:** Team lead + teammates, each with own context window. Shared task list with dependency tracking. Inbox-based messaging for inter-agent communication.
- **Display Modes:** In-process (all in one terminal, `Shift+Down` to cycle) or split panes (tmux/iTerm2)
- **Communication:** Direct teammate messaging, broadcast, shutdown requests. File locking for task claiming.
- **Storage:** `~/.claude/teams/{team-name}/config.json` and `~/.claude/tasks/{team-name}/`
- **Remote/Distributed:** **NOT supported.** All teammates run on the same machine. No cross-machine dispatch.
- **Limitations:** No session resumption, no nested teams, one team per session, lead is fixed, split panes require tmux/iTerm2.
- **Docs:** https://code.claude.com/docs/en/agent-teams
- **Blog (Addy Osmani):** https://addyosmani.com/blog/claude-code-agent-teams/

### 1.2 Remote Control

Mobile/browser access to a local Claude Code session.

- **How it works:** `claude remote-control` or `/rc` generates QR code/URL. Local session polls Anthropic API; cloud routes messages between devices.
- **Key property:** Everything runs locally — code, MCP servers, tools stay on your machine. Cloud handles message routing only.
- **Security:** Outbound HTTPS only, no inbound ports opened. TLS transport.
- **Availability:** Max subscribers ($100-$200/mo), Pro coming soon. Not on Team/Enterprise.
- **Multi-agent angle:** Boris Cherny (head of Claude Code) runs 5+ parallel agents simultaneously, managing them from phone "like a dispatch controller" — 300+ PRs/month.
- **Docs:** https://code.claude.com/docs/en/remote-control

### 1.3 Claude Code Remote (Cloud Execution)

Async cloud execution using `&` prefix.

- **How it works:** Commands prefixed with `&` dispatch to Anthropic's cloud infrastructure. Works asynchronously — you can close terminal, switch devices.
- **Requires:** Claude Max subscription.
- **Runs:** Same Claude Code capabilities (file ops, bash, MCP) but on Anthropic's managed cloud environment.
- **Docs:** Referenced in hooks documentation at https://code.claude.com/docs/en/hooks-guide

### 1.4 SSH Tunnel Support

Native remote server access.

- **How it works:** Claude Code connects to remote servers via SSH tunnels. Compatible with VS Code Remote.
- **Use case:** Working on code in remote environments without custom tooling.
- **Blog:** https://medium.com/@joe.njenga/how-i-use-claude-code-ssh-to-connect-to-any-remote-server-like-a-pro-f08ed35e5569

### 1.5 Claude Agent SDK

Programmatic Claude Code access.

- **TypeScript:** `@anthropic-ai/claude-agent-sdk`
- **Python:** `claude-code-sdk`
- **Features:** Session management, custom tools as callbacks, partial message streaming, structured outputs, auto-exit.
- **Use case:** Building custom orchestrators that programmatically control Claude Code instances.

---

## 2. Multi-Agent Orchestrators

### 2.1 Multiclaude

The orchestrator this project uses.

- **Architecture:** Supervisor + persistent agents + ephemeral workers, all in tmux sessions. Git worktrees for isolation. Daemon manages refresh/sync.
- **Remote support:** **Single-machine only.** All agents run on the same machine in tmux.
- **Strengths:** Multiplayer review mode, long-prompt-then-walk-away workflow, persistent agent definitions, merge queue integration.
- **Comparison note:** "Gas Town is more complex and better for solo devs working on hobby projects, while Multiclaude offers support for team usage/review."

### 2.2 Gas Town

By Steve Yegge. Multi-agent workspace manager.

- **Architecture:** "The Mayor" (coordinator) + "Polecats" (workers) + "Rigs" (project containers) + "Beads" (work items) + "Convoys" (work bundles).
- **Communication:** Mailbox-based agent coordination, git-backed persistent state.
- **Remote support:** **No remote/distributed capability.** Local workspace manager only (`~/gt/`).
- **Kubernetes operator exists:** https://github.com/boshu2/gastown-operator (OpenShift-native, FIPS-compliant) — may enable distributed deployment.
- **Repo:** https://github.com/steveyegge/gastown

### 2.3 Claude Squad

By smtg-ai. Terminal UI for managing multiple AI agents.

- **Architecture:** tmux sessions + git worktrees per task. Supports Claude Code, Aider, Codex, Gemini, OpenCode.
- **Features:** Background task completion, auto-accept mode, change review before push.
- **Remote support:** **No remote dispatch.** Local-only orchestration.
- **Repo:** https://github.com/smtg-ai/claude-squad
- **Docs:** https://smtg-ai.github.io/claude-squad/

### 2.4 Claude Swarm (affaan-m)

Hackathon project using Claude Agent SDK.

- **Architecture:** Opus 4.6 analyzes codebase → breaks into dependency graph → parallel subtasks via Agent SDK.
- **Remote support:** Not documented.
- **Repo:** https://github.com/affaan-m/claude-swarm

### 2.5 Ruflo (formerly Claude Flow)

By ruvnet. Comprehensive orchestration platform.

- **Architecture:** Multi-layered: CLI/MCP entry → Q-Learning Router + Mixture of Experts → Swarm Coordination → 60+ agents.
- **Scale:** 5,800+ commits, 215 MCP tools, 60+ agents, 8 AgentDB controllers.
- **Swarm topologies:** Mesh, hierarchical, ring, star.
- **Consensus:** Raft, Byzantine Fault Tolerant, Gossip, CRDT, weighted voting.
- **Inter-agent communication:** Shared namespace with LRU caching and SQLite persistence.
- **Remote support:** Architecture supports distributed topologies conceptually, but documentation focuses on single-machine deployment.
- **Repo:** https://github.com/ruvnet/ruflo

### 2.6 Claude Code Agent Farm

By Dicklesworthstone. Scale-focused orchestrator.

- **Architecture:** Python orchestration script manages 20-50 Claude Code agents via tmux. Lock-based coordination with `active_work_registry.json`.
- **Workflow modes:** Bug fixing (random chunks), best practices sweeps, cooperating agents.
- **Monitoring:** Real-time dashboard with context warnings, heartbeat tracking.
- **Remote support:** **No remote agent support.** All agents run locally via tmux.
- **Repo:** https://github.com/Dicklesworthstone/claude_code_agent_farm

### 2.7 Agent Orchestrator (ComposioHQ)

Agent-agnostic orchestrator.

- **Architecture:** Plugin-based with 8 swappable abstraction layers. TypeScript.
- **Runtimes:** tmux (default), Docker, Kubernetes, direct process spawning.
- **Agent support:** Claude Code, Codex, Aider, OpenCode.
- **Isolation:** Git worktrees per agent.
- **Autonomous features:** CI failure recovery, code review handling, merge operations.
- **Remote support:** Docker/K8s runtimes could enable distributed deployment, but docs focus on single-machine.
- **Repo:** https://github.com/ComposioHQ/agent-orchestrator

### 2.8 Metaswarm

Self-improving multi-agent framework.

- **Architecture:** 18 specialized agents, 13 orchestration skills, 15 commands. TDD enforcement, quality gates.
- **Repo:** https://github.com/dsifry/metaswarm

### 2.9 ccswarm

Git worktree-based multi-agent orchestration.

- **Architecture:** Template-based scaffolding, task delegation, git worktree isolation.
- **Repo:** https://github.com/nwiizo/ccswarm

### 2.10 Oh-My-ClaudeCode (OMC)

Teams-first orchestration.

- **Architecture:** Team as canonical orchestration surface (v4.1.7+). Legacy swarm keyword/skill removed.
- **Repo:** https://github.com/yeachan-heo/oh-my-claudecode

### 2.11 Claude Octopus

Multi-tentacled orchestrator.

- **Features:** 32 specialized personas, 38 commands, 50 skills. Coordinates Claude Code, Codex CLI, Gemini CLI.
- **Repo:** https://github.com/nyldn/claude-octopus

### 2.12 Claude Code Agentrooms

Desktop app + API for multi-agent orchestration.

- **Architecture:** Hub-and-spoke: Frontend → Orchestrator → Local/Remote agents via HTTP.
- **Remote support:** **YES — explicitly supports remote agents** via HTTP endpoints (e.g., `mac-mini.local:8081`, `cloud-instance:8081`).
- **Agent routing:** `@agent-name` mention-based routing. General requests auto-decomposed.
- **Auth:** Uses Claude CLI authentication — no API keys needed.
- **Stack:** React + TypeScript frontend, Deno backend, Electron desktop.
- **Repo:** https://github.com/baryhuang/claude-code-by-agents
- **Website:** https://claudecode.run/

---

## 3. MCP Bridges & Inter-Instance Communication

### 3.1 Claude Relay (WebSocket + MCP)

**The most directly relevant project for cross-machine Claude Code communication.**

- **Architecture:** Central WebSocket relay server (`server.js` on port 9999) + MCP server per Claude Code instance.
- **Flow:** Claude Code → MCP Server (stdio) → WebSocket → Relay Server → other instances.
- **Cross-machine:** SSH port forwarding via `autossh` for persistent tunnels. macOS LaunchAgent for auto-start.
- **Session tracking:** `~/claude-relay/sessions/registry.json` — all instances can discover peers.
- **MCP Tools:** `relay_send`, `relay_receive`, `relay_peers`, `relay_status`, `relay_sessions`.
- **Human-readable IDs:** Shell aliases assign names (CC-1, CODEX, etc.) before launching.
- **Repo:** https://github.com/gvorwaller/claude-relay

### 3.2 Claude Code MCP Server (steipete)

Agent-in-agent pattern via MCP.

- **How it works:** Exposes Claude Code as an MCP server. Other LLMs (Cursor, Windsurf) can invoke Claude Code tools programmatically.
- **Pattern:** Nested agent architecture — parent agent delegates file ops, git, terminal to Claude Code.
- **Tools exposed:** Bash, Read, Write, Edit, LS, GrepTool, GlobTool, Replace.
- **Remote angle:** MCP's HTTP transport could theoretically enable remote invocation, but the project focuses on local stdio.
- **Repo:** https://github.com/steipete/claude-code-mcp

### 3.3 Claude Telegram Bridge

Async communication bridge.

- **How it works:** MCP server bridges Claude Code to Telegram. Claude sends questions/progress to phone; user replies route back.
- **Multi-session:** Each session gets unique IDs; Telegram swipe-reply routes to correct session.
- **Use case:** Async monitoring/interaction when away from terminal.
- **Source:** https://lobehub.com/mcp/ricardoagl-claude-telegram-bridge

### 3.4 HTTP-to-Stdio MCP Bridge (NOVA Guide)

Bridge for containerized Claude Code.

- **How it works:** Lightweight HTTP-to-stdio bridge enables containerized Claude Code to access stdio-based MCP servers on host.
- **Use case:** Docker/container deployments where MCP servers use stdio but Claude Code needs HTTP.
- **Repo:** https://github.com/majkonautic/NOVA_claude-code-mcp-guide

### 3.5 Remote MCP Servers (Anthropic Official)

Anthropic's official remote MCP support.

- **Transport:** Streamable HTTP and SSE for remote MCP server connections.
- **MCP Connector:** Messages API can connect to remote MCP servers directly without a separate client.
- **Custom Connectors:** Bridge between Claude and remote MCP servers.
- **Blog:** https://claude.com/blog/claude-code-remote-mcp
- **Docs:** https://platform.claude.com/docs/en/agents-and-tools/remote-mcp-servers

---

## 4. Remote Execution Approaches

### 4.1 Coder (Used by Anthropic Engineers)

Cloud development environments for Claude Code.

- **How Anthropic uses it:** Engineers run Claude Code in isolated cloud environments. Multiple agents in parallel, each with own workspace.
- **Benefits:** No local resource contention, multi-repo context pre-configured, security isolation ("lethal traffic" principle).
- **Quote:** "Teams building remote infrastructure now will scale instantly as models improve."
- **Blog:** https://coder.com/blog/building-for-2026-why-anthropic-engineers-are-running-claude-code-remotely-with-c

### 4.2 Tailscale SSH + Claude Code

VPN mesh for remote development.

- **How it works:** Tailscale creates encrypted mesh network; Claude Code connects via SSH to any machine on the mesh.
- **Blog:** https://tsoporan.com/blog/remote-ai-development-claude-code-tailscale/

### 4.3 AI Coding Agent Dashboard

Cross-device orchestration.

- **How it works:** Web dashboard for orchestrating Claude Code across devices.
- **Blog:** https://blog.marcnuri.com/ai-coding-agent-dashboard

---

## 5. Hooks & Extensions for Multi-Agent Workflows

### 5.1 Multi-Agent Observability (disler)

Real-time monitoring across Claude Code agents.

- **Architecture:** Hook scripts → HTTP POST → Bun server → SQLite → WebSocket → Vue dashboard.
- **Events tracked:** All 12 Claude Code hook types (PreToolUse, PostToolUse, SessionStart, SubagentStart, etc.).
- **Session tracking:** Dual-color coding for concurrent agent visualization.
- **Remote capable:** HTTP-based — dashboard can be viewed from any network location.
- **Repo:** https://github.com/disler/claude-code-hooks-multi-agent-observability

### 5.2 Hooks Mastery (disler)

Comprehensive hooks guide.

- **Repo:** https://github.com/disler/claude-code-hooks-mastery

### 5.3 Continuous Claude (parcadei)

Context management via hooks.

- **How it works:** Hooks maintain state via ledgers and handoffs. MCP execution without context pollution. Agent orchestration with isolated context windows.
- **Repo:** https://github.com/parcadei/Continuous-Claude-v3

### 5.4 TeammateIdle / TaskCompleted Hooks

Anthropic's official hooks for agent team quality gates.

- **TeammateIdle:** Runs when a teammate is about to go idle. Exit code 2 sends feedback to keep working.
- **TaskCompleted:** Runs when task marked complete. Exit code 2 prevents completion with feedback.
- **Docs:** https://code.claude.com/docs/en/hooks-guide

---

## 6. Community Discussion & Resources

### 6.1 Unofficial Claude Code Forum

- Community-built forum organized by topic: Skills, MCP, Hooks, Subagents, Prompting, Memory, Troubleshooting.
- **Blog about it:** https://skooloflife.medium.com/why-we-built-the-unofficial-claude-code-forum-d7096cc1fd87

### 6.2 Anthropic Discord

- Official Discord community for Anthropic & Claude.ai.
- Fast-moving channels — historical answers hard to find (motivating the unofficial forum).
- **Link:** https://discord.com/invite/6PPFFzqPDZ

### 6.3 awesome-claude-code

Curated list of skills, hooks, slash-commands, agent orchestrators, applications, and plugins.

- **Repo:** https://github.com/hesreallyhim/awesome-claude-code

### 6.4 Claude Code Swarm Orchestration Guides (Gists)

Comprehensive guides to TeammateTool and Task system patterns.

- **Swarm Orchestration Skill:** https://gist.github.com/kieranklaassen/4f2aba89594a4aea4ad64d753984b2ea
- **Multi-Agent Orchestration System:** https://gist.github.com/kieranklaassen/d2b35569be2c7f1412c64861a219d51f

### 6.5 Key Blog Posts

- **Addy Osmani — Claude Code Swarms:** https://addyosmani.com/blog/claude-code-agent-teams/
- **Agent Teams Setup Guide:** https://darasoba.medium.com/how-to-set-up-and-use-claude-code-agent-teams-and-actually-get-great-results-9a34f8648f6d
- **Configure Claude Code for Agent Teams:** https://medium.com/@haberlah/configure-claude-code-to-power-your-agent-team-90c8d3bca392
- **SitePoint — Claude Code Agent Teams:** https://www.sitepoint.com/anthropic-claude-code-agent-teams/
- **Complete Guide 2026:** https://claudefa.st/blog/guide/agents/agent-teams
- **Remote Control Guide:** https://claudefa.st/blog/guide/development/remote-control-guide
- **Claude Code's Hidden Multi-Agent System:** https://paddo.dev/blog/claude-code-hidden-swarm/

### 6.6 Security Notes

- Claude Code CVEs reported (Feb 2026) for RCE via untrusted repos. Fixed in recent releases.
- **The Register:** https://www.theregister.com/2026/02/26/clade_code_cves/
- **The Hacker News:** https://thehackernews.com/2026/02/claude-code-flaws-allow-remote-code.html

---

## 7. Comparison Matrix

| Tool | Remote Agents | Cross-Machine | Inter-Agent Messaging | MCP Integration | Isolation | License |
|------|:---:|:---:|:---:|:---:|:---:|:---:|
| **Agent Teams (Anthropic)** | No | No | Yes (mailbox) | N/A (native) | Context windows | Proprietary |
| **Multiclaude** | No | No | Yes (file-based) | No | Git worktrees + tmux | — |
| **Gas Town** | No | No | Yes (mailbox) | No | Git hooks | — |
| **Claude Squad** | No | No | No | No | Git worktrees + tmux | OSS |
| **Claude Relay** | **Yes** | **Yes (SSH)** | **Yes (WebSocket)** | **Yes** | Per-instance | OSS |
| **Claude Code Agentrooms** | **Yes** | **Yes (HTTP)** | **Yes (@mentions)** | No | Per-agent workdir | OSS |
| **Claude Code MCP Server** | Partial | Partial (MCP HTTP) | Yes (MCP tools) | **Yes** | Per-invocation | OSS |
| **Ruflo** | Conceptual | Conceptual | Yes (shared namespace) | **Yes** | Scoped memory | OSS |
| **Agent Farm** | No | No | Yes (lock files) | No | tmux panes | OSS |
| **Agent Orchestrator (Composio)** | Via Docker/K8s | Possible | Yes (centralized) | No | Git worktrees | OSS |
| **Coder (Anthropic's approach)** | **Yes** | **Yes** | Via git/PRs | Yes | Cloud VMs | Commercial |

---

## 8. Key Findings & Recommendations

### Finding 1: No Mature Cross-Machine Orchestrator Exists

Despite the explosion of multi-agent Claude Code tools (15+ found), **only two explicitly support remote/cross-machine agents**:
- **Claude Relay** (WebSocket + MCP + SSH tunnels) — most promising for inter-instance communication
- **Claude Code Agentrooms** (HTTP-based hub-and-spoke) — most mature for remote agent deployment

Everything else (multiclaude, Gas Town, Claude Squad, Agent Farm, etc.) is **single-machine only**.

### Finding 2: Anthropic's Official Direction

Anthropic is building toward remote execution through:
1. **Remote Control** — session mobility (not multi-agent)
2. **Claude Code Remote (`&` prefix)** — cloud execution
3. **Remote MCP Servers** — Streamable HTTP transport for remote tools
4. **Agent Teams** — native multi-agent, but local-only currently
5. **Coder partnership** — cloud dev environments for agents

The gap is clear: **Agent Teams + Remote MCP + Remote Control are separate features that haven't been unified into a distributed multi-agent system.**

### Finding 3: Practical Remote Patterns Today

For teams wanting cross-machine Claude Code collaboration today:

1. **Claude Relay** for real-time inter-instance messaging via WebSocket/MCP
2. **Coder** for cloud-hosted agent environments
3. **Tailscale/SSH** for secure tunnels between machines
4. **Git/PRs** as the coordination layer (what multiclaude effectively uses)
5. **Remote MCP servers** for shared tool access across instances

### Finding 4: Multiclaude's Position

Multiclaude is architecturally well-positioned for a remote extension because:
- It already has a daemon managing agent lifecycles
- Inter-agent messaging exists (file-based)
- Git worktrees provide isolation
- tmux sessions are the unit of agent execution

To add remote support, multiclaude would need:
- A relay/bridge for cross-machine agent messaging (Claude Relay pattern)
- Remote worktree provisioning (Coder pattern or SSH-based)
- Agent spawning on remote hosts (SSH + tmux or Docker)
- Centralized task/message coordination (could extend existing file-based system to use a shared server)

### Finding 5: MCP Is the Convergence Point

MCP (Model Context Protocol) is emerging as the standard for inter-agent communication:
- Anthropic official support for remote MCP servers
- Claude Relay uses MCP for agent discovery and messaging
- Claude Code MCP Server enables agent-in-agent patterns
- Ruflo deeply integrates MCP for tool dispatch

Any remote collaboration solution for multiclaude should be **MCP-native** to stay aligned with the ecosystem direction.

---

## Appendix: All Repositories Found

| Repository | URL |
|---|---|
| claude-relay | https://github.com/gvorwaller/claude-relay |
| claude-code-by-agents (Agentrooms) | https://github.com/baryhuang/claude-code-by-agents |
| claude-squad | https://github.com/smtg-ai/claude-squad |
| gastown | https://github.com/steveyegge/gastown |
| ruflo (claude-flow) | https://github.com/ruvnet/ruflo |
| claude-code-mcp | https://github.com/steipete/claude-code-mcp |
| claude_code_agent_farm | https://github.com/Dicklesworthstone/claude_code_agent_farm |
| claude-code-hooks-multi-agent-observability | https://github.com/disler/claude-code-hooks-multi-agent-observability |
| claude-swarm | https://github.com/affaan-m/claude-swarm |
| ccswarm | https://github.com/nwiizo/ccswarm |
| metaswarm | https://github.com/dsifry/metaswarm |
| oh-my-claudecode | https://github.com/yeachan-heo/oh-my-claudecode |
| claude-octopus | https://github.com/nyldn/claude-octopus |
| agent-orchestrator | https://github.com/ComposioHQ/agent-orchestrator |
| multi-agent-squad | https://github.com/bijutharakan/multi-agent-squad |
| agents (wshobson) | https://github.com/wshobson/agents |
| awesome-claude-code | https://github.com/hesreallyhim/awesome-claude-code |
| gastown-operator (K8s) | https://github.com/boshu2/gastown-operator |
| gastown-gui | https://github.com/web3dev1337/gastown-gui |
| Continuous-Claude-v3 | https://github.com/parcadei/Continuous-Claude-v3 |
| claude-code-hooks-mastery | https://github.com/disler/claude-code-hooks-mastery |
| NOVA claude-code-mcp-guide | https://github.com/majkonautic/NOVA_claude-code-mcp-guide |
| claude-agentic-framework | https://github.com/dralgorhythm/claude-agentic-framework |
| agent-swarm (desplega-ai) | https://github.com/desplega-ai/agent-swarm |
| claude-swarm-orchestration | https://github.com/MaTriXy/claude-swarm-orchestration |
| claude-code-system-prompts | https://github.com/Piebald-AI/claude-code-system-prompts |
| superpowers (Agent Teams issue) | https://github.com/obra/superpowers/issues/429 |
| myclaude | https://github.com/cexll/myclaude |
