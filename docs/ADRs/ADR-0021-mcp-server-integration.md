# ADR-0021: MCP Server Integration

- **Status:** Accepted
- **Date:** 2026-03-01
- **Decision Makers:** Architecture review, research (PR #156)
- **Related PRs:** #156, #169, #177-#180, #184-#185, #191, #196-#197

## Context

AI assistants (Claude, ChatGPT) increasingly support tool use via the Model Context Protocol (MCP). Exposing ThreeDoors as an MCP server enables LLM agents to query, create, and manage tasks programmatically.

## Decision

Implement ThreeDoors as an **MCP server** with dual transport (stdio and SSE):

1. **Resources** — Read-only access to task lists, analytics, session data
2. **Tools** — Task CRUD operations, status changes, search
3. **Prompts** — Pre-built prompt templates for common task management workflows
4. **Security middleware** — Rate limiting, input validation, audit logging

## Architecture

- MCP server runs as a separate mode of the same binary (`threedoors mcp`)
- Shares the same `TaskProvider` infrastructure as TUI and CLI
- Stdio transport for local Claude Code integration
- SSE transport for network-accessible LLM agents
- Security middleware validates all tool inputs

## Implementation (Epic 24 — 8 stories)

| Story | Component | PR |
|-------|-----------|-----|
| 24.1 | Server scaffold with transports | #177 |
| 24.2 | Read-only resources and query tools | #180 |
| 24.3 | Security middleware | #179 |
| 24.4 | Proposal store and enrichment API | #185 |
| 24.5 | TUI proposal review view | #197 |
| 24.6 | Analytics resources and tools | #184 |
| 24.7 | Task relationship graph | #191 |
| 24.8 | Prompt templates | #196 |

## Consequences

### Positive
- AI assistants can manage tasks on behalf of users
- Proposal workflow lets AI suggest task changes for human review
- Analytics accessible to LLM for insights generation
- Task relationship graph enables complex multi-task operations

### Negative
- MCP protocol is still evolving — may need updates
- Security middleware must prevent prompt injection via task content
- SSE transport exposes tasks over network — requires careful access control
- Adds significant surface area to test and maintain
