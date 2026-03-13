# Sprint Change Proposal: Retrospector Agent Reliability Fixes

**Date:** 2026-03-12
**Author:** bold-eagle (worker, /plan-work pipeline)
**Epic:** TBD (awaiting project-watchdog allocation)
**Priority:** P1

## Problem Statement

The retrospector agent (SLAES — Epic 51) has three reliability failures that prevent it from operating as designed:

### 1. Messaging Identity Bug (Critical)

Persistent agents spawned via `multiclaude agents spawn` cannot reliably receive messages via `multiclaude message list`. The retrospector gets "No messages" even when the supervisor sends it messages. The agent identity is not properly registered with the daemon's messaging system after spawn or respawn.

**Impact:** Retrospector cannot receive task assignments, saga detection requests, or status checks from supervisor. Two-way communication is broken — retrospector can SEND but cannot RECEIVE.

**Scope:** This affects ALL persistent agents spawned via `multiclaude agents spawn`, not just retrospector. However, retrospector is the most affected because its autonomous operation depends on receiving merge event notifications and supervisor directives.

**Root cause hypothesis:** The `multiclaude agents spawn` command registers the agent with a session identity that differs from the identity used by `multiclaude message send <agent-name>`. The daemon routes messages to an identity the agent doesn't poll. This is a multiclaude infrastructure bug, but we can work around it in agent definitions.

### 2. BOARD.md Write Access (High)

Retrospector cannot persist recommendations to `docs/decisions/BOARD.md` because:
- The retrospector definition says "Cannot: Create PRs (deferred to Phase 2)" — so it has no mechanism to commit BOARD.md changes
- Even if it could write to BOARD.md in the shared checkout, concurrent edits from other agents or humans cause conflicts
- The BOARD.md file is ~400 lines and growing — append-only edits at the bottom still conflict when multiple agents touch it

**Impact:** The core value proposition of SLAES — filing actionable recommendations — is blocked. Retrospector can detect patterns but has no durable output mechanism.

### 3. Context Exhaustion (Medium)

Retrospector runs out of context after processing ~45 PRs and needs restart, losing in-memory analytical state:
- The agent definition specifies restart triggers (20 PRs or 8 hours) but the actual limit is hit at ~45 PRs because the definition itself consumes significant context
- On restart, state must be rebuilt from JSONL findings log, but the findings log may not exist yet (bootstrap problem)
- Analytical patterns (e.g., "the last 10 PRs all had lint failures") are lost on restart — only raw data points survive via JSONL
- No checkpoint mechanism exists between "everything in memory" and "everything lost"

**Impact:** Retrospector has a maximum operational window before degradation. Pattern detection that spans more than one session is unreliable.

## Proposed Approach

### Fix 1: Messaging — Identity Verification + Fallback Polling

Since the messaging identity bug is in multiclaude's daemon (which we don't control from this repo), the fix is a workaround in the agent definition:

1. **Startup identity probe:** On startup, retrospector sends itself a test message and verifies receipt within 30 seconds. If the probe fails, log the failure and fall back to alternative communication.
2. **Fallback: file-based message inbox.** Create `docs/operations/retrospector-inbox.jsonl` as a secondary message channel. Supervisor can append messages there. Retrospector polls this file alongside `multiclaude message list`.
3. **Agent definition update:** Add identity verification to the startup sequence and document the fallback for supervisor.

### Fix 2: BOARD.md — Recommendation Queue File

Instead of writing directly to BOARD.md (which requires PR authority and causes conflicts):

1. **Recommendation queue file:** Retrospector writes recommendations to `docs/operations/retrospector-recommendations.jsonl` (append-only, no conflicts possible).
2. **Periodic batch processing:** Project-watchdog or supervisor periodically reads the queue and applies recommendations to BOARD.md via a governed PR.
3. **Agent definition update:** Change retrospector's output target from BOARD.md to the queue file. Add queue schema documentation.

This separates the "detection" concern (retrospector) from the "persistence" concern (project-watchdog/supervisor), which aligns with the existing authority model.

### Fix 3: Context — Structured Checkpointing

1. **Checkpoint file:** `docs/operations/retrospector-checkpoint.json` stores analytical state between restarts: current mode rotation position, pattern detection state (rolling windows), last processed PR, hours of operation.
2. **Periodic checkpoint writes:** Every 5 PRs or 2 hours, flush analytical state to checkpoint file.
3. **Startup from checkpoint:** On restart, read checkpoint file first, then JSONL log for raw data, then catch up on new merges.
4. **Reduce definition size:** Trim verbose sections of `agents/retrospector.md` that consume context without adding value (e.g., move detailed JSONL schema to a referenced doc instead of inline).

## Rejected Alternatives

| Alternative | Why Rejected |
|---|---|
| Fix multiclaude daemon messaging directly | multiclaude is external tooling; we control agent definitions and scripts in THIS repo, not daemon internals |
| Give retrospector PR creation authority | Violates Phase 1 scope (read + recommend only); PRs from an autonomous agent without human review create risk |
| Write directly to BOARD.md in shared checkout | Causes merge conflicts with concurrent agents; no atomicity guarantee |
| Increase Claude context window for retrospector | Not configurable; model limitation |
| Split retrospector into multiple agents | Over-engineering; the three fixes address the root causes without architectural change |
| Use tmux output as communication fallback | Tmux output is not durable and other agents can't read it programmatically |

## Stories

| Story | Title | Priority | Depends On |
|---|---|---|---|
| 62.1 | Messaging Identity Verification + File-Based Fallback | P1 | None |
| 62.2 | Recommendation Queue File + BOARD.md Batch Pipeline | P1 | None |
| 62.3 | Structured Checkpointing + Context Budget Optimization | P1 | None |

All three stories are independent and can be parallelized.

## Risk Assessment

- **Low risk:** All changes are to agent definition files (`.md`) and operational files (`.jsonl`, `.json`). No application code changes.
- **Testing:** Manual validation by spawning retrospector and verifying each fix. No automated tests needed (infrastructure, not code).
- **Rollback:** Revert the agent definition file to the previous version.
