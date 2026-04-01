# ADR-0026: Self-Driving Development Pipeline

- **Status:** Accepted (revised 2026-03-31)
- **Date:** 2026-03-01
- **Decision Makers:** Design decision H7, architecture review
- **Related PRs:** #135, #141, #149, #152, #159-#164
- **Related ADRs:** ADR-0025 (Story-Driven Development)

## Context

ThreeDoors uses AI agents (multiclaude workers) for story implementation. Manual story dispatch is a bottleneck. Epic 22 explored automating the dispatch-implement-review cycle.

## Considered Options

1. **Shell script MVP** — Bash script parsing story files and dispatching `multiclaude worker create`
2. **GitHub Actions** — Trigger workers from CI on story file changes
3. **Supervisor enhancement** — Extend multiclaude supervisor with story awareness
4. **`multiclaude pipeline` command** — First-class pipeline support in multiclaude

## Decision

**Shell script MVP** (Option A) as the immediate approach, with the TUI providing dispatch and monitoring capabilities.

## Implementation (Epic 22 — 8 stories)

| Story | Component | PR |
|-------|-----------|-----|
| 22.1 | Dispatch data model and queue persistence | #149 |
| 22.2 | Dispatch engine with multiclaude CLI wrapper | #152 |
| 22.3 | TUI dispatch key binding and confirmation | #163 |
| 22.4 | Dev queue view (list, approve, kill) | #162 |
| 22.5 | Worker status polling and task update loop | #161 |
| 22.6 | Auto-generated review and follow-up tasks | #164 |
| 22.7 | Optional story file generation | #159 |
| 22.8 | Safety guardrails (rate limiting, cost caps, audit) | #160 |

## Rationale

- Shell script approach works immediately with existing multiclaude infrastructure
- TUI integration provides human-in-the-loop oversight
- Safety guardrails (rate limiting, cost caps, audit logging) prevent runaway costs
- Dev queue view gives visibility into active workers

## Consequences

### Positive
- Reduced manual overhead for story dispatch
- Human approval required before dispatch (safety)
- Cost and rate limiting prevent accidental overuse
- Audit log provides full history of automated actions

### Negative
- Shell script approach is fragile for complex workflows
- Requires multiclaude to be running and configured
- Worker failures need manual investigation
- Cost caps are advisory — actual API costs depend on worker behavior

## Post-Incident Revisions

This ADR was written before multi-agent failures were experienced in production. Three of four documented incidents (INC-001, INC-002, INC-004) directly contradict assumptions in the original pipeline design. This section records the corrections.

### 1. Shared checkout assumption was wrong — worktree isolation is required

**Original assumption:** Agents operate in filesystem isolation.

**Reality ([INC-001](../operations/INC-001-pr-shepherd-contamination.md)):** Persistent agents shared a single git checkout. pr-shepherd's `git checkout` and `git rebase` in the shared checkout destroyed uncommitted supervisor work and contaminated the working tree for all agents. The pipeline's safety guardrails (Story 22.8) addressed cost and rate limiting but not filesystem-level isolation between agents.

**Correction:** Every agent — persistent or ephemeral — must operate in a dedicated git worktree. multiclaude now creates isolated worktrees for workers at spawn time and refreshes them via a daemon loop. Persistent agents that need branch operations must use their own worktrees, never the shared checkout.

### 2. Agent definitions and memory can carry harmful cargo-culted instructions

**Original assumption:** Agent definitions and supervisor memory entries are reliable instructions that agents follow faithfully.

**Reality ([INC-002](../operations/INC-002-destructive-git-sync-override.md)):** A mandatory `git fetch origin main && git rebase origin/main` instruction was cargo-culted into supervisor memory from a pre-worktree era. It was applied to 100+ worker dispatches over 3 days, overriding multiclaude's built-in worktree management. At least one worker was left in a detached HEAD state mid-rebase. The instruction was never re-evaluated because it was encoded as a "MUST" rule.

**Correction:** Platform abstractions (multiclaude's worktree management) must be trusted over agent-level instructions. "MUST" rules in agent memory calcify into unquestioned dogma — they require periodic audit and explicit rationale that can be re-validated. Don't override the platform; extend it.

### 3. Inter-agent messaging requires explicit tool routing

**Original assumption:** Inter-agent messaging works — agents can communicate to coordinate the pipeline.

**Reality ([INC-004](../operations/INC-004-sendmessage-tool-silent-drop.md)):** Claude Code's built-in `SendMessage` tool silently drops messages intended for multiclaude agents. It is designed for subagent communication within a single Claude process, not cross-process messaging. All agents used `SendMessage` by default because the model prefers purpose-built tools over Bash commands, even when definitions document the correct CLI syntax. 49+ messages were silently dropped, including critical epic number allocations.

**Correction:** Inter-agent communication must use `multiclaude message send` via Bash, never Claude Code's `SendMessage` tool. Agent definitions now include an INC-004 guardrail in their Communication section. A PreToolUse hook blocking `SendMessage` for known agent names would provide infrastructure-level enforcement.

### General principle

**Soft constraints fail under concurrent agent execution.** Prose instructions in definitions, memory entries, and standing orders are necessary but insufficient. All three incidents share a common pattern: a prose-level constraint (isolation convention, sync instruction, messaging syntax) was either wrong, stale, or ignored by the model's tool preference. Enforcement must be infrastructure — git hooks, CI checks, PreToolUse hooks, and platform-managed worktrees — not prose alone.
