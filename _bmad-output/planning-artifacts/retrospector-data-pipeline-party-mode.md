# Party Mode Artifact: Retrospector Data Pipeline

**Date:** 2026-03-13
**Participants:** Winston (Architect), John (PM), Mary (Analyst), Murat (TEA)
**Topic:** Evaluate pipeline approach for periodically committing retrospector operational data to git

## Problem Statement

Retrospector writes findings to `docs/operations/` files (retrospector-findings.jsonl, retrospector-checkpoint.json, retrospector-recommendations.jsonl) but these files are untracked or have uncommitted changes. They exist only in the main checkout where retrospector runs as a persistent agent. Worker agents spawned in worktrees cannot see this data, blocking analytical work.

Three states of brokenness identified:
1. `retrospector-findings.jsonl` — does not exist in repo at all
2. `retrospector-checkpoint.json` — tracked, but has uncommitted local changes
3. `retrospector-inbox.jsonl` / `retrospector-recommendations.jsonl` — tracked but empty in git; real data only in main checkout

## Options Evaluated

### Option A: Give retrospector PR creation authority for data-only commits
- **Pros:** Most direct — producer commits its own data
- **Cons:** Violates observer/participant separation. Retrospector is a monitoring agent; adding PR creation means it participates in the merge queue it observes. Consumes CI cycles for data-only commits. Could conflict with its own PR analysis patterns.
- **Decision:** REJECTED — architectural smell; observer becomes participant

### Option B: project-watchdog periodically commits retrospector data
- **Pros:** project-watchdog already handles doc syncs; natural extension
- **Cons:** Scope creep — mixes planning doc sync with operational data sync. project-watchdog operates on 23-min heartbeat and its core mission is planning doc consistency. Without a dedicated trigger, it would need to poll for changes as part of its existing loop.
- **Decision:** REJECTED as standalone — but project-watchdog adopted as the executor in the hybrid approach

### Option C: Supervisor periodically commits as part of duty cycle
- **Pros:** Supervisor already has broad authority
- **Cons:** Violates established guardrail: "supervisor coordinates, agents execute." Fragile to supervisor restarts — if supervisor goes down, data goes stale indefinitely with no recovery mechanism.
- **Decision:** REJECTED — violates coordination/execution separation

### Option D: Dedicated data-sync cron job
- **Pros:** Lowest blast radius, simplest mechanism, no agent authority changes
- **Cons:** Raw cron can't handle branch protection, PR creation, merge conflicts, or error recovery without an agent executor
- **Decision:** REJECTED as standalone — but cron trigger adopted in the hybrid approach

## Adopted Approach: Hybrid D+B — Cron-Triggered project-watchdog Data Sync

**Rationale:** Combines the reliability of a dedicated cron trigger (Option D) with the intelligence of an agent executor (Option B), avoiding the scope creep of embedding it in project-watchdog's existing polling loop.

### Architecture

1. **Trigger:** New CronCreate entry fires every 2-4 hours, sends `SYNC_OPERATIONAL_DATA` message to project-watchdog
2. **Executor:** project-watchdog receives the message, checks for changed files in `docs/operations/`, creates branch `data-sync/<timestamp>`, commits, pushes, creates minimal PR
3. **Verification:** Supervisor periodically verifies data freshness (check `git log --oneline docs/operations/ --since="8 hours ago"`)
4. **Convention:** `docs/operations/` is the canonical directory for operational data files — any agent writing operational data should place files there, and the sync cron will pick them up

### Failure Mode Analysis

| Scenario | Impact | Recovery |
|----------|--------|----------|
| Cron fires but project-watchdog is down | Data waits for next cycle | No data loss (append-only files) |
| project-watchdog busy with doc sync when data sync arrives | Message queues until next idle | Natural backpressure |
| Main checkout has dirty working tree | project-watchdog can stash/handle | Agent is smart enough to manage git state |
| PR merge conflict with concurrent PR | Standard merge-queue handling | Rebase and retry |
| Supervisor restarts (crons lost) | Crons re-created on startup | Part of startup checklist |

### Acceptance Criteria for Testing (from TEA)

1. After cron fires, operational data files in git match what's on disk in main checkout
2. Staleness SLA: `docs/operations/*.jsonl` files in git are never more than 6 hours behind main checkout (2x commit interval as buffer)
3. Idempotency: if no files changed, no empty commit is created
4. Branch protection compliance: changes go through a PR, not direct to main
5. No interference: data sync PRs don't block or delay story PRs in the merge queue

### Extensibility

The `SYNC_OPERATIONAL_DATA` pattern is generic. Any operational data files placed under `docs/operations/` by any agent will be automatically swept up by the sync cron. This eliminates the need to grant PR authority to individual monitoring agents and centralizes the data-to-git pipeline.

### Supervisor Duty

Add to supervisor standing orders: periodically verify retrospector data is being committed. Check `git log --oneline docs/operations/ --since="8 hours ago"` — if no commits and retrospector is active, investigate the sync pipeline.
