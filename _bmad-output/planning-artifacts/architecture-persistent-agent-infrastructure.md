# Architecture Decision: Persistent BMAD Agent Infrastructure

**Date:** 2026-03-08
**Status:** Proposed
**Context:** PR #249 research + course correction party mode validation
**Scope:** Agent communication, lifecycle, resource management, and coordination for persistent BMAD governance agents

---

## 1. Overview

This document extends the existing ThreeDoors architecture to cover the persistent agent infrastructure that enables autonomous project governance. It does NOT modify the ThreeDoors application architecture — this is purely development infrastructure.

### Problem

With 210+ merged PRs across 30+ epics, manual governance doesn't scale. Planning documents (story status, ROADMAP.md, PRD alignment) and architecture documents drift from reality after every merge. The current 3 persistent agents (merge-queue, pr-shepherd, envoy) handle operational concerns but not governance.

### Solution

Add 2 persistent agents (project-watchdog, arch-watchdog) and 2 cron jobs (SM sprint health, QA coverage audit) to create a self-governing project lifecycle.

---

## 2. Agent Communication Architecture

### Decision: Message-Driven Communication via multiclaude

**Pattern:** Point-to-point messaging using `multiclaude message send`

**Rationale:**
- Established multiclaude primitive — no new infrastructure needed
- Each message is explicit and traceable
- No shared state files (prevents race conditions across worktrees)
- No webhook infrastructure required (polling + messages is sufficient at current scale)

### Rejected Alternatives

| Alternative | Why Rejected |
|------------|-------------|
| Shared file protocol | Race conditions in separate worktrees; file-based coordination adds complexity |
| Webhook-based events | multiclaude doesn't support webhook triggers; over-engineering for scale |
| Dense agent mesh (N*N channels) | Combinatorial explosion; hub-and-spoke is simpler and sufficient |

### Communication Topology

```
                    ┌─────────────────┐
                    │   supervisor     │
                    │  (human/agent)   │
                    └────────┬────────┘
                             │ escalations only
                    ┌────────┴────────┐
                    │                 │
         ┌─────────▼──────┐  ┌──────▼─────────┐
         │ project-watchdog│  │  arch-watchdog  │
         │  (PM - persist) │  │ (Arch - persist)│
         │                 │  │                 │
         │ Monitors:       │  │ Monitors:       │
         │ - Merged PRs    │◄─┤ - Code changes  │
         │ - Story status  │──►- Arch docs      │
         │ - ROADMAP.md    │  │ - New patterns   │
         │ - PRD alignment │  │ - Design records │
         └────────┬────────┘  └──────┬─────────┘
                  │                   │
         ┌────────┴───────────────────┴────────┐
         │           Message Bus               │
         │    (multiclaude message send)        │
         └────────┬───────────────────┬────────┘
                  │                   │
    ┌─────────────▼──┐          ┌─────▼──────────┐
    │  merge-queue    │          │  pr-shepherd   │
    │  (persist)      │          │  (persist)     │
    │  Merges PRs     │          │  Rebases PRs   │
    └─────────────────┘          └────────────────┘
                  │
    ┌─────────────▼──┐
    │    envoy        │
    │  (persist)      │
    │  Issue triage   │
    └─────────────────┘

    ┌─────────────────┐          ┌────────────────┐
    │  SM cron (4h)   │          │ QA cron (weekly)│
    │  Sprint health  │          │ Coverage audit  │
    └─────────────────┘          └────────────────┘
```

### Message Flow: PR Merge Cascade

The primary governance flow triggered on every PR merge:

```
1. merge-queue merges PR #NNN
   │
2. project-watchdog detects merge (polling every 10-15 min)
   ├── Updates story X.Y status → "Done (PR #NNN)"
   ├── Updates ROADMAP.md epic progress
   ├── Checks PRD alignment
   │   └── If drift: messages arch-watchdog
   │         └── arch-watchdog reviews architecture docs
   │             └── If update needed: updates docs, messages project-watchdog
   │                 └── project-watchdog flags affected stories
   │
3. envoy cross-checks open issues (existing behavior)
   └── If PR fixes open issue: comments and closes
```

### Cascade Prevention

Each message includes a **correlation ID** (the PR number). Agents track processed PR numbers and skip already-processed ones. This prevents:
- Circular notifications (A messages B, B messages A about the same PR)
- Duplicate processing on agent restart
- Amplification loops

**Implementation:** Each agent maintains a local list of recently processed PR numbers (last 50). On restart, re-scan last 10 merged PRs to catch any missed during downtime, but skip those already in the processed list.

---

## 3. Agent Lifecycle Management

### Decision: multiclaude-Native Lifecycle

Agents are spawned, monitored, and restarted using existing multiclaude infrastructure.

### Startup

```bash
# Spawn persistent agents
multiclaude agents spawn --name project-watchdog --class persistent --prompt-file agents/project-watchdog.md
multiclaude agents spawn --name arch-watchdog --class persistent --prompt-file agents/arch-watchdog.md

# Start cron jobs
# SM: via /loop skill (runs in supervisor or dedicated session)
/loop 4h /bmad-bmm-sprint-status

# QA: via system cron (weekly is too long for /loop)
# crontab entry: 0 9 * * 1 multiclaude work "Run weekly QA coverage audit" --repo ThreeDoors
```

### Startup Ordering

**No ordering required.** project-watchdog and arch-watchdog operate independent polling loops with message bridges. Either can start first. If one sends a message before the other is ready, multiclaude queues the message for delivery when the recipient starts.

### Monitoring

- **Health check:** `multiclaude worker list` shows agent status
- **Logs:** `multiclaude logs <agent-name>` for troubleshooting
- **Metrics:** Each agent logs its polling cycle count and actions taken

### Restart and Recovery

- **Crash recovery:** `multiclaude agent restart <name>` restarts a crashed agent
- **Idempotent operations:** All agent actions are idempotent — re-processing a PR produces the same result
- **Catch-up on restart:** On startup, agents re-scan the last 10 merged PRs to catch anything missed during downtime
- **No data loss:** Agents don't maintain critical state that would be lost on crash — they derive state from git history and GitHub API

### Shutdown

```bash
# Stop individual agent
multiclaude worker rm <agent-name>

# Stop all (including persistent agents)
multiclaude stop-all
```

---

## 4. Resource Management

### Decision: Conservative Polling with Adaptive Intervals

### Resource Budget (5 Persistent Agents)

| Agent | Poll Interval | Est. API Calls/Hour | Notes |
|-------|---------------|---------------------|-------|
| merge-queue | 5-10 min | 6-12 | Existing |
| pr-shepherd | 10-15 min | 4-6 | Existing |
| envoy | 15-20 min | 3-4 | Existing |
| project-watchdog | 10-15 min | 4-6 | **New** |
| arch-watchdog | 20-30 min | 2-3 | **New** |
| **Total** | | **19-31/hour** | |

Cron additions:
- SM (every 4h): ~6 calls/day
- QA (weekly): ~1 call/week

### tmux Session Management

Each persistent agent runs in its own tmux window managed by multiclaude. 5 persistent agents = 5 tmux windows. Each agent has its own git worktree, preventing file conflicts.

### Scaling Limits

- **Recommended max:** 6-7 persistent agents before coordination overhead dominates
- **Current proposal:** 5 persistent agents — well within limits
- **Scaling strategy:** If adding more agents, evaluate consolidation (merging roles) before adding new agents

### Adaptive Polling (Future Enhancement)

Not implemented in MVP, but noted for Phase 2 tuning:
- Increase polling frequency during high-merge periods (multiple PRs merged in 1 hour)
- Decrease polling frequency during quiet periods (no merges in 2+ hours)
- Backoff on API errors or rate limits

---

## 5. Conflict Resolution

### Decision: Authority Boundaries + Separate Worktrees

### File Domain Separation

Each agent edits only files within its authority domain. There is NO overlap:

| Agent | Editable Files | Read-Only Access |
|-------|---------------|-----------------|
| project-watchdog | `docs/stories/*.md`, `ROADMAP.md` | All code, all docs |
| arch-watchdog | `docs/architecture/*.md` | All code, all docs |
| merge-queue | None (merges PRs, doesn't edit files) | All |
| pr-shepherd | None (rebases branches, doesn't edit files) | All |
| envoy | None (comments on issues, doesn't edit files) | All |

### Why This Prevents Conflicts

1. **No overlapping write domains:** project-watchdog writes to story files; arch-watchdog writes to architecture docs. They never edit the same file.
2. **Separate worktrees:** Each agent operates in its own git worktree via multiclaude. File system isolation prevents race conditions.
3. **No shared state files:** Communication is via messages, not shared files.

### Edge Case: Both Agents Process Same PR

If both agents detect the same merged PR simultaneously:
- project-watchdog updates story status and ROADMAP.md
- arch-watchdog checks code changes against architecture docs
- These are independent operations on different files — no conflict
- If arch-watchdog finds drift and messages project-watchdog, the message includes the correlation PR number, so project-watchdog knows the context

### Edge Case: Agent Creates a Commit While Another Agent's PR is Being Merged

- Each agent operates in its own worktree on its own branch
- Agents that create commits (project-watchdog updating story files, arch-watchdog updating arch docs) should create PRs, not push directly to main
- merge-queue handles merging these PRs through the normal process

---

## 6. Integration with multiclaude Infrastructure

### Agent Definition Files

Agent behavior is defined in markdown files at `agents/<agent-name>.md`. These files specify:
- **Monitoring surface:** What to poll and how often
- **Trigger model:** What events trigger action
- **Authority boundaries:** What files can be edited
- **Escalation rules:** When to message supervisor
- **Restart behavior:** How to recover from downtime

### Existing Agent Compatibility

The new agents complement existing agents without modification:

| Existing Agent | Interaction with New Agents |
|---------------|---------------------------|
| merge-queue | Triggers project-watchdog and arch-watchdog indirectly (they detect merges via polling) |
| pr-shepherd | No interaction — rebasing is orthogonal to governance |
| envoy | Parallel operation — envoy checks issues, project-watchdog checks story/PRD alignment |

### multiclaude Message Protocol

Messages follow the existing `multiclaude message send <recipient> <message>` format:

```bash
# project-watchdog notifies arch-watchdog of PRD drift
multiclaude message send arch-watchdog "PRD drift detected: PR #NNN changed scope of Epic 28. Verify architecture docs for alignment. Correlation: PR-NNN"

# arch-watchdog notifies project-watchdog of architecture update
multiclaude message send project-watchdog "Architecture docs updated for pattern change in internal/tasks/. Stories referencing old pattern may need tech notes. Correlation: PR-NNN"

# Either agent escalates to supervisor
multiclaude message send supervisor "Scope change detected: PR #NNN introduces feature not in ROADMAP.md. Needs human decision. Correlation: PR-NNN"
```

---

## 7. Key Design Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Communication model | Message-driven (multiclaude messages) | Established primitive; no new infrastructure |
| Topology | Hub-and-spoke (PM hub + Architect independent loop) | Simpler than mesh; covers both governance gaps |
| Polling intervals | 10-15 min (PM), 20-30 min (Architect) | Balances freshness with API cost |
| Conflict prevention | Authority boundaries + separate worktrees | Zero overlap in write domains |
| Cascade prevention | Correlation ID per PR | Prevents circular notifications |
| Recovery model | Idempotent operations + catch-up scan | No critical state to lose on crash |
| Lifecycle management | multiclaude-native (spawn, restart, rm) | Uses existing infrastructure |
| Scaling limit | 6-7 persistent agents max | Beyond this, coordination overhead dominates |
| SM/QA model | Cron jobs, not persistent agents | Periodic summarization, not continuous monitoring |

---

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Agent creates incorrect doc updates | Medium | Low | All updates go through PRs; merge-queue validates |
| API rate limiting from too many agents | Low | Medium | Conservative polling intervals; backoff on errors |
| Agent stuck in infinite loop | Very Low | Low | multiclaude monitors agent health; manual restart available |
| Stale correlation ID list on long downtime | Low | Low | Re-scan last 10 PRs on restart covers typical gaps |
| Agent worktree diverges from main | Low | Medium | Agents should rebase before creating commits |
